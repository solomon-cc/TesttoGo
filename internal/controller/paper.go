package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
)

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
		Title:     req.Title,
		CreatorID: c.GetUint("userID"),
		Questions: string(questionIDs),
		Duration:  req.Duration,
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
		"id":        paper.ID,
		"title":     paper.Title,
		"duration":  paper.Duration,
		"questions": questions,
	})
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
	var papers []entity.Paper
	if err := database.DB.Find(&papers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取试卷列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"papers": papers})
}
