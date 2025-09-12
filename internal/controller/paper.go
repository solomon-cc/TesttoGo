package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
)

// subjectNameToCode 将科目中文名称转换为英文代码
func subjectNameToCode(name string) string {
	subjectMap := map[string]string{
		"数学":   "math",
		"语言词汇": "vocabulary",
		"阅读":   "reading",
		"识字":   "literacy",
	}
	if code, exists := subjectMap[name]; exists {
		return code
	}
	return name // 如果映射不存在，返回原值
}

// @Summary 创建试卷
// @Description 创建新的试卷
// @Tags 试卷
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param request body request.CreatePaperRequest true "创建试卷请求参数"
// @Success 200 {object} map[string]interface{} "返回创建的试卷ID"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/papers [post]
func CreatePaper(c *gin.Context) {
	var req request.CreatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证题目是否存在
	var count int64
	database.DB.Model(&entity.Question{}).Where("id IN ?", req.QuestionIDs).Count(&count)
	if int(count) != len(req.QuestionIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "包含不存在的题目"})
		return
	}

	// 将题目ID列表转换为JSON字符串
	questionIDs, err := json.Marshal(req.QuestionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "题目列表序列化失败"})
		return
	}

	paper := entity.Paper{
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   c.GetUint("userID"),
		Grade:       req.Grade,
		Subject:     req.Subject,
		Type:        req.Type,
		Difficulty:  req.Difficulty,
		Status:      req.Status,
		Questions:   string(questionIDs),
		TotalScore:  req.TotalScore,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
	}

	if err := database.DB.Create(&paper).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建试卷失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": paper.ID})
}

func GetPaper(c *gin.Context) {
	id := c.Param("id")
	var paper entity.Paper
	if err := database.DB.First(&paper, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "试卷不存在"})
		return
	}

	// 解析题目ID列表
	var questionIDs []uint
	if err := json.Unmarshal([]byte(paper.Questions), &questionIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析题目列表失败"})
		return
	}

	// 获取题目详情
	var questions []entity.Question
	if err := database.DB.Find(&questions, questionIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取题目详情失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          paper.ID,
		"title":       paper.Title,
		"description": paper.Description,
		"grade":       paper.Grade,
		"subject":     subjectNameToCode(paper.Subject),
		"type":        paper.Type,
		"difficulty":  paper.Difficulty,
		"status":      paper.Status,
		"total_score": paper.TotalScore,
		"start_time":  paper.StartTime,
		"end_time":    paper.EndTime,
		"questions":   questions,
		"created_at":  paper.CreatedAt,
		"updated_at":  paper.UpdatedAt,
	})
}

// @Summary 更新试卷
// @Description 更新现有试卷
// @Tags 试卷
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param id path int true "试卷ID"
// @Param request body request.CreatePaperRequest true "更新试卷请求参数"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "试卷不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/papers/{id} [put]
func UpdatePaper(c *gin.Context) {
	id := c.Param("id")
	var paper entity.Paper
	if err := database.DB.First(&paper, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "试卷不存在"})
		return
	}

	var req request.CreatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证题目是否存在（如果提供了题目ID列表）
	if len(req.QuestionIDs) > 0 {
		var count int64
		database.DB.Model(&entity.Question{}).Where("id IN ?", req.QuestionIDs).Count(&count)
		if int(count) != len(req.QuestionIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "包含不存在的题目"})
			return
		}

		// 将题目ID列表转换为JSON字符串
		questionIDs, err := json.Marshal(req.QuestionIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "题目列表序列化失败"})
			return
		}
		paper.Questions = string(questionIDs)
	}

	// 更新试卷信息
	paper.Title = req.Title
	paper.Description = req.Description
	paper.Grade = req.Grade
	paper.Subject = req.Subject
	paper.Type = req.Type
	paper.Difficulty = req.Difficulty
	paper.Status = req.Status
	paper.TotalScore = req.TotalScore
	paper.StartTime = req.StartTime
	paper.EndTime = req.EndTime

	if err := database.DB.Save(&paper).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新试卷失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "试卷更新成功", "id": paper.ID})
}

func SubmitPaper(c *gin.Context) {
	id := c.Param("id")
	paperID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的试卷ID"})
		return
	}

	var answers []request.SubmitAnswerRequest
	if err := c.ShouldBindJSON(&answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	var totalScore int

	// 开启事务
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, answer := range answers {
		// 获取题目信息
		var question entity.Question
		if err := tx.First(&question, answer.QuestionID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "题目不存在"})
			return
		}

		// 计算得分
		score := 0
		if question.Answer == answer.Answer {
			score = 10 // 每题10分
		}
		totalScore += score

		// 保存答题记录
		userAnswer := entity.UserAnswer{
			UserID:     userID,
			PaperID:    uint(paperID),
			QuestionID: answer.QuestionID,
			Answer:     answer.Answer,
			Score:      score,
		}

		if err := tx.Create(&userAnswer).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存答题记录失败"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交试卷失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "提交成功",
		"score":   totalScore,
	})
}

func GetPaperResult(c *gin.Context) {
	paperID := c.Param("id")
	userID := c.GetUint("userID")

	var answers []entity.UserAnswer
	if err := database.DB.Where("user_id = ? AND paper_id = ?", userID, paperID).Find(&answers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取答题记录失败"})
		return
	}

	totalScore := 0
	for _, answer := range answers {
		totalScore += answer.Score
	}

	c.JSON(http.StatusOK, gin.H{
		"answers":    answers,
		"totalScore": totalScore,
	})
}

func ListPapers(c *gin.Context) {
	// 构建响应结构体
	type PaperWithStats struct {
		ID           uint        `json:"id"`
		Title        string      `json:"title"`
		Description  string      `json:"description"`
		CreatorID    uint        `json:"creator_id"`
		Grade        string      `json:"grade"`
		Subject      string      `json:"subject"` // 将被转换为英文代码
		Type         string      `json:"type"`
		Difficulty   string      `json:"difficulty"`
		Status       string      `json:"status"`
		Questions    string      `json:"questions"`
		Duration     int         `json:"duration"`
		TotalScore   int         `json:"total_score"`
		StartTime    *time.Time  `json:"start_time"`
		EndTime      *time.Time  `json:"end_time"`
		CreatedAt    time.Time   `json:"created_at"`
		UpdatedAt    time.Time   `json:"updated_at"`
		Creator      entity.User `json:"creator,omitempty"`
		AttemptCount int         `json:"attempt_count"`
		AverageScore float64     `json:"average_score"`
	}

	var papers []entity.Paper
	if err := database.DB.Preload("Creator").Find(&papers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取试卷列表失败"})
		return
	}

	// 为每个试卷添加统计数据
	var papersWithStats []PaperWithStats
	for _, paper := range papers {
		// 统计参与人数：查询该试卷的不重复用户数
		var attemptCount int64
		database.DB.Model(&entity.UserAnswer{}).
			Where("paper_id = ?", paper.ID).
			Distinct("user_id").
			Count(&attemptCount)

		// 计算平均分：获取该试卷所有用户的总分，然后计算平均值
		var averageScore float64
		if attemptCount > 0 {
			// 获取每个用户的总分
			var userScores []struct {
				UserID     uint
				TotalScore int
			}

			database.DB.Model(&entity.UserAnswer{}).
				Select("user_id, SUM(score) as total_score").
				Where("paper_id = ?", paper.ID).
				Group("user_id").
				Scan(&userScores)

			// 计算平均分
			if len(userScores) > 0 {
				var totalSum int
				for _, score := range userScores {
					totalSum += score.TotalScore
				}
				averageScore = float64(totalSum) / float64(len(userScores))
			}
		}

		// 创建包含统计数据的试卷对象，转换subject字段
		paperWithStats := PaperWithStats{
			ID:           paper.ID,
			Title:        paper.Title,
			Description:  paper.Description,
			CreatorID:    paper.CreatorID,
			Grade:        paper.Grade,
			Subject:      subjectNameToCode(paper.Subject), // 转换为英文代码
			Type:         paper.Type,
			Difficulty:   paper.Difficulty,
			Status:       paper.Status,
			Questions:    paper.Questions,
			Duration:     paper.Duration,
			TotalScore:   paper.TotalScore,
			StartTime:    paper.StartTime,
			EndTime:      paper.EndTime,
			CreatedAt:    paper.CreatedAt,
			UpdatedAt:    paper.UpdatedAt,
			Creator:      paper.Creator,
			AttemptCount: int(attemptCount),
			AverageScore: averageScore,
		}
		papersWithStats = append(papersWithStats, paperWithStats)
	}

	c.JSON(http.StatusOK, gin.H{"data": papersWithStats, "total": len(papersWithStats)})
}
