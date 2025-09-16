package controller

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/internal/model/response"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
)

// @Summary 创建题目
// @Description 创建新的题目
// @Tags 题目
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param request body request.CreateQuestionRequest true "创建题目请求参数"
// @Success 200 {object} map[string]interface{} "返回创建的题目ID"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/questions [post]
func CreateQuestion(c *gin.Context) {
	var req request.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")

	// 验证科目和主题关系
	var subjectID, topicID *uint
	var subjectCode, topicCode string

	// 优先使用ID字段，如果没有则使用字符串字段进行查找
	if req.SubjectID != nil && req.TopicID != nil && *req.SubjectID > 0 && *req.TopicID > 0 {
		// 验证科目存在
		var subject entity.Subject
		if err := database.DB.Where("id = ? AND is_active = ?", *req.SubjectID, true).First(&subject).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "科目不存在或已禁用"})
			return
		}

		// 验证主题存在并且属于该科目
		var topic entity.Topic
		if err := database.DB.Where("id = ? AND subject_id = ? AND is_active = ?", *req.TopicID, *req.SubjectID, true).First(&topic).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "主题不存在或不属于该科目"})
			return
		}

		subjectID = req.SubjectID
		topicID = req.TopicID
		subjectCode = subject.Code
		topicCode = topic.Code
	} else if req.Subject != "" && req.Topic != "" {
		// 使用字符串查找对应的ID
		var subject entity.Subject
		if err := database.DB.Where("code = ? AND is_active = ?", req.Subject, true).First(&subject).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "科目代码不存在: " + req.Subject})
			return
		}

		var topic entity.Topic
		if err := database.DB.Where("code = ? AND subject_id = ? AND is_active = ?", req.Topic, subject.ID, true).First(&topic).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "主题代码不存在或不属于该科目: " + req.Topic})
			return
		}

		subjectID = &subject.ID
		topicID = &topic.ID
		subjectCode = req.Subject
		topicCode = req.Topic
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "必须提供科目和主题信息（ID或代码）"})
		return
	}

	question := entity.Question{
		Title:       req.Title,
		Type:        entity.QuestionType(req.Type),
		Difficulty:  req.Difficulty,
		Grade:       req.Grade,
		SubjectID:   subjectID,
		TopicID:     topicID,
		Subject:     subjectCode, // 保持向后兼容
		Topic:       topicCode,   // 保持向后兼容
		Options:     req.Options,
		Answer:      req.Answer,
		Explanation: req.Explanation,
		CreatorID:   userID,
		MediaURLs:   req.MediaURLs,
		LayoutType:  req.LayoutType,
		ElementData: req.ElementData,
		Tags:        req.Tags,
	}

	if err := database.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建题目失败"})
		return
	}

	// 重新查询题目以获取关联的科目和主题信息
	var createdQuestion entity.Question
	if err := database.DB.Preload("SubjectRef").Preload("TopicRef").First(&createdQuestion, question.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取创建的题目信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       createdQuestion.ID,
		"question": createdQuestion,
	})
}

// @Summary 获取题目列表
// @Description 获取题目列表，支持分页和过滤
// @Tags 题目
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param type query string false "题目类型"
// @Param difficulty query string false "题目难度"
// @Param grade query string false "年级"
// @Param subject query string false "科目"
// @Param topic query string false "主题"
// @Success 200 {array} entity.Question "题目列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/questions [get]
func ListQuestions(c *gin.Context) {
	var questions []entity.Question
	query := database.DB.Preload("SubjectRef").Preload("TopicRef").Order("id desc")

	// 支持按类型、难度、年级过滤
	if qType := c.Query("type"); qType != "" {
		query = query.Where("type = ?", qType)
	}
	if difficulty := c.Query("difficulty"); difficulty != "" {
		// 将字符串难度转换为数字
		var difficultyInt int
		switch difficulty {
		case "easy":
			difficultyInt = 1
		case "medium":
			difficultyInt = 2
		case "hard":
			difficultyInt = 3
		default:
			// 如果是数字字符串，直接转换
			if d, err := strconv.Atoi(difficulty); err == nil {
				difficultyInt = d
			} else {
				difficultyInt = 1 // 默认简单
			}
		}
		query = query.Where("difficulty = ?", difficultyInt)
	}
	if grade := c.Query("grade"); grade != "" {
		query = query.Where("grade = ?", grade)
	}

	// 支持按科目过滤 - 优先使用ID，兼容字符串
	if subjectID := c.Query("subject_id"); subjectID != "" {
		// 将字符串ID转换为数字
		if subjectIDNum, err := strconv.Atoi(subjectID); err == nil && subjectIDNum > 0 {
			query = query.Where("subject_id = ?", subjectIDNum)
		}
	} else if subject := c.Query("subject"); subject != "" {
		// 兼容字符串查询 - 查找对应的科目ID
		var subjectEntity entity.Subject
		if err := database.DB.Where("code = ? AND is_active = ?", subject, true).First(&subjectEntity).Error; err == nil {
			query = query.Where("subject_id = ?", subjectEntity.ID)
		} else {
			// 如果找不到科目ID，则使用旧的字符串匹配
			query = query.Where("subject = ?", subject)
		}
	}

	// 支持按主题过滤 - 优先使用ID，兼容字符串
	if topicID := c.Query("topic_id"); topicID != "" {
		// 将字符串ID转换为数字
		if topicIDNum, err := strconv.Atoi(topicID); err == nil && topicIDNum > 0 {
			query = query.Where("topic_id = ?", topicIDNum)
		}
	} else if topic := c.Query("topic"); topic != "" {
		// 兼容字符串查询 - 查找对应的主题ID
		var topicEntity entity.Topic
		if err := database.DB.Where("code = ? AND is_active = ?", topic, true).First(&topicEntity).Error; err == nil {
			query = query.Where("topic_id = ?", topicEntity.ID)
		} else {
			// 如果找不到主题ID，则使用旧的字符串匹配
			query = query.Where("topic = ?", topic)
		}
	}

	// 分页处理 - 支持limit参数和count参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 如果有limit参数，使用limit作为page_size
	if limitParam := c.Query("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			pageSize = limit
			page = 1 // limit模式下默认获取第一页
		}
	}

	// 如果有count参数，使用count作为page_size（专项练习模式）
	if countParam := c.Query("count"); countParam != "" {
		if count, err := strconv.Atoi(countParam); err == nil && count > 0 {
			if count > 100 {
				count = 100 // 限制最多100道题
			}
			pageSize = count
			page = 1 // count模式下默认获取第一页
		}
	}

	// 使用相同的查询条件计算总数
	var total int64
	query.Model(&entity.Question{}).Count(&total)

	// 检查是否为专项练习模式（有count参数）
	isRandomMode := c.Query("count") != ""

	var err error

	// 根据模式选择排序方式
	if isRandomMode {
		// 专项练习模式：使用随机排序
		err = query.Order("id DESC").Limit(pageSize * 3).Find(&questions).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目列表失败"})
			return
		}

		// 如果获取到的题目数量超过需要的数量，随机选择
		if len(questions) > pageSize {
			rand.Seed(time.Now().UnixNano())
			for i := len(questions) - 1; i > 0; i-- {
				j := rand.Intn(i + 1)
				questions[i], questions[j] = questions[j], questions[i]
			}
			questions = questions[:pageSize]
		}
	} else {
		// 普通模式：使用分页
		err = query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&questions).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目列表失败"})
			return
		}
	}

	// 根据模式选择响应格式
	if isRandomMode {
		// 专项练习模式：返回简化的题目数组（与原GetRandomQuestions保持一致）
		c.JSON(http.StatusOK, questions)
	} else {
		// 普通模式：返回带统计数据的响应
		var questionsWithStats []response.QuestionWithStatsResponse
		for _, question := range questions {
			// 计算统计数据
			var totalAttempts, correctCount int64
			database.DB.Model(&entity.UserAnswer{}).Where("question_id = ?", question.ID).Count(&totalAttempts)
			database.DB.Model(&entity.UserAnswer{}).Where("question_id = ? AND is_correct = ?", question.ID, true).Count(&correctCount)

			// 计算正确率
			var correctRate float64
			if totalAttempts > 0 {
				correctRate = float64(correctCount) / float64(totalAttempts) * 100
			}

			questionWithStats := response.QuestionWithStatsResponse{
				ID:          question.ID,
				Title:       question.Title,
				Type:        string(question.Type),
				Difficulty:  question.Difficulty,
				Grade:       question.Grade,
				Subject:     question.Subject,
				Topic:       question.Topic,
				Options:     question.Options,
				Answer:      question.Answer,
				Explanation: question.Explanation,
				CreatorID:   question.CreatorID,
				MediaURL:    question.MediaURL,
				MediaURLs:   question.MediaURLs,
				LayoutType:  question.LayoutType,
				ElementData: question.ElementData,
				Tags:        question.Tags,
				CreatedAt:   question.CreatedAt,
				UpdatedAt:   question.UpdatedAt,
				UsageCount:  totalAttempts,
				CorrectRate: correctRate,
			}
			questionsWithStats = append(questionsWithStats, questionWithStats)
		}

		// 确保items始终是数组而不是null
		if questionsWithStats == nil {
			questionsWithStats = []response.QuestionWithStatsResponse{}
		}

		c.JSON(http.StatusOK, gin.H{
			"total": total,
			"items": questionsWithStats,
		})
	}
}

func GetQuestion(c *gin.Context) {
	id := c.Param("id")
	var question entity.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}
	c.JSON(http.StatusOK, question)
}

func UpdateQuestion(c *gin.Context) {
	id := c.Param("id")
	var req request.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var question entity.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 更新题目信息
	updates := map[string]interface{}{
		"title":        req.Title,
		"difficulty":   req.Difficulty,
		"grade":        req.Grade,
		"subject":      req.Subject,
		"topic":        req.Topic,
		"options":      req.Options,
		"answer":       req.Answer,
		"explanation":  req.Explanation,
		"media_urls":   req.MediaURLs,
		"layout_type":  req.LayoutType,
		"element_data": req.ElementData,
		"tags":         req.Tags,
	}

	if err := database.DB.Model(&question).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新题目失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func DeleteQuestion(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&entity.Question{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除题目失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// @Summary 单题答题
// @Description 用户对单个题目进行答题
// @Tags 题目
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param id path int true "题目ID"
// @Param request body request.SingleAnswerRequest true "答题请求参数"
// @Success 200 {object} response.QuestionAnswerResponse "答题结果"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "题目不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/questions/{id}/answer [post]
func AnswerQuestion(c *gin.Context) {
	questionID := c.Param("id")
	userID := c.GetUint("userID")

	var req request.SingleAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取题目信息
	var question entity.Question
	if err := database.DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 判断答案是否正确
	isCorrect := checkAnswer(question.Answer, req.Answer)
	score := 0
	if isCorrect {
		score = 10 // 单题答对得10分
	}

	// 保存答题记录
	userAnswer := entity.UserAnswer{
		UserID:     userID,
		QuestionID: uint(mustParseInt(questionID)),
		Answer:     req.Answer,
		Score:      score,
		IsCorrect:  isCorrect,
		AnswerType: "single",
		PaperID:    0, // 单题答题不关联试卷
	}

	if err := database.DB.Create(&userAnswer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存答题记录失败"})
		return
	}

	// 构造响应
	resp := response.QuestionAnswerResponse{
		QuestionID:  userAnswer.QuestionID,
		UserAnswer:  userAnswer.Answer,
		IsCorrect:   userAnswer.IsCorrect,
		Score:       userAnswer.Score,
		Explanation: question.Explanation,
		AnsweredAt:  userAnswer.CreatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

// 辅助函数：检查答案是否正确
func checkAnswer(correctAnswer, userAnswer string) bool {
	// 去除前后空格并转为小写进行比较
	correct := strings.TrimSpace(strings.ToLower(correctAnswer))
	user := strings.TrimSpace(strings.ToLower(userAnswer))
	return correct == user
}

// @Summary 获取题目答题统计
// @Description 获取指定题目的答题统计信息
// @Tags 题目
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param id path int true "题目ID"
// @Success 200 {object} response.QuestionStatisticsResponse "题目统计信息"
// @Failure 404 {object} map[string]interface{} "题目不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/questions/{id}/statistics [get]
func GetQuestionStatistics(c *gin.Context) {
	questionID := c.Param("id")

	// 验证题目是否存在
	var question entity.Question
	if err := database.DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 统计该题目的答题记录
	var totalAttempts, correctCount int64
	database.DB.Model(&entity.UserAnswer{}).Where("question_id = ?", questionID).Count(&totalAttempts)
	database.DB.Model(&entity.UserAnswer{}).Where("question_id = ? AND is_correct = ?", questionID, true).Count(&correctCount)

	// 计算正确率
	var accuracyRate float64
	if totalAttempts > 0 {
		accuracyRate = float64(correctCount) / float64(totalAttempts) * 100
	}

	resp := response.QuestionStatisticsResponse{
		QuestionID:    uint(mustParseInt(questionID)),
		Title:         question.Title,
		TotalAttempts: totalAttempts,
		CorrectCount:  correctCount,
		AccuracyRate:  accuracyRate,
	}

	c.JSON(http.StatusOK, resp)
}

// 辅助函数：字符串转整数
func mustParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
