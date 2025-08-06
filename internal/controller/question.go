package controller

import (
	"net/http"
	"strconv"

	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
)

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
		Options:     req.Options,
		Answer:      req.Answer,
		Explanation: req.Explanation,
		CreatorID:   userID,
		Tags:        req.Tags,
	}

	if err := database.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建题目失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": question.ID})
}

func ListQuestions(c *gin.Context) {
	var questions []entity.Question
	query := database.DB.Order("id desc")

	// 支持按类型和难度过滤
	if qType := c.Query("type"); qType != "" {
		query = query.Where("type = ?", qType)
	}
	if difficulty := c.Query("difficulty"); difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
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

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": questions,
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
		"title":       req.Title,
		"difficulty":  req.Difficulty,
		"options":     req.Options,
		"answer":      req.Answer,
		"explanation": req.Explanation,
		"tags":        req.Tags,
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
