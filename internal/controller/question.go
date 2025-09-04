package controller

import (
	"net/http"
	"strconv"
	"strings"

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
	question := entity.Question{
		Title:       req.Title,
		Type:        entity.QuestionType(req.Type),
		Difficulty:  req.Difficulty,
		Grade:       req.Grade,
		Subject:     req.Subject,
		Topic:       req.Topic,
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

	c.JSON(http.StatusOK, gin.H{"id": question.ID})
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
	query := database.DB.Order("id desc")

	// 支持按类型、难度、年级、科目、主题过滤
	if qType := c.Query("type"); qType != "" {
		query = query.Where("type = ?", qType)
	}
	if difficulty := c.Query("difficulty"); difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if grade := c.Query("grade"); grade != "" {
		query = query.Where("grade = ?", grade)
	}
	if subject := c.Query("subject"); subject != "" {
		query = query.Where("subject = ?", subject)
	}
	if topic := c.Query("topic"); topic != "" {
		query = query.Where("topic = ?", topic)
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	var total int64
	database.DB.Model(&entity.Question{}).Count(&total)

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&questions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目列表失败"})
		return
	}

	// 构建带统计数据的响应
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

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": questionsWithStats,
	})
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

// @Summary 随机获取题目
// @Description 根据条件随机获取题目进行练习
// @Tags 题目
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param type query string false "题目类型"
// @Param difficulty query int false "难度等级(1-5)"
// @Param grade query string false "年级"
// @Param subject query string false "科目"
// @Param topic query string false "主题"
// @Param tags query string false "题目标签"
// @Param count query int false "题目数量,默认1"
// @Success 200 {array} entity.Question "题目列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/questions/random [get]
func GetRandomQuestions(c *gin.Context) {
	count, _ := strconv.Atoi(c.DefaultQuery("count", "1"))
	if count > 10 {
		count = 10 // 限制最多10道题
	}

	query := database.DB.Model(&entity.Question{})

	// 按类型过滤
	if qType := c.Query("type"); qType != "" {
		query = query.Where("type = ?", qType)
	}

	// 按难度过滤
	if difficulty := c.Query("difficulty"); difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	// 按年级过滤
	if grade := c.Query("grade"); grade != "" {
		query = query.Where("grade = ?", grade)
	}

	// 按科目过滤
	if subject := c.Query("subject"); subject != "" {
		query = query.Where("subject = ?", subject)
	}

	// 按主题过滤
	if topic := c.Query("topic"); topic != "" {
		query = query.Where("topic = ?", topic)
	}

	// 按标签过滤
	if tags := c.Query("tags"); tags != "" {
		query = query.Where("tags LIKE ?", "%"+tags+"%")
	}

	var questions []entity.Question
	if err := query.Order("RAND()").Limit(count).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目失败"})
		return
	}

	c.JSON(http.StatusOK, questions)
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
