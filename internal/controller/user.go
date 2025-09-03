package controller

import (
	"net/http"
	"strconv"
	"time"

	"testogo/internal/model/entity"
	"testogo/internal/model/response"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary 获取用户列表
// @Description 获取用户列表，支持分页
// @Tags 用户
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param page query integer false "页码，默认1"
// @Param page_size query integer false "每页数量，默认10"
// @Success 200 {object} map[string]interface{} "用户列表和总数"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users [get]
func ListUsers(c *gin.Context) {
	var users []entity.User
	query := database.DB.Order("id desc")

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	var total int64
	database.DB.Model(&entity.User{}).Count(&total)

	err := query.Select("id, username, role, created_at, updated_at").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": users,
	})
}

// @Summary 更新用户角色
// @Description 更新指定用户的角色
// @Tags 用户
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param id path integer true "用户ID"
// @Param request body map[string]string true "角色信息 {\"role\": \"user|teacher|admin\"}"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users/{id}/role [put]
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Role string `json:"role" binding:"required,oneof=user teacher admin"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户是否存在
	var user entity.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 更新用户角色
	if err := database.DB.Model(&user).Update("role", input.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户角色失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// 检查用户是否存在
	var user entity.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 不允许删除管理员账户
	if user.Role == entity.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除管理员账户"})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetUserPerformance retrieves user's learning performance statistics
func GetUserPerformance(c *gin.Context) {
	// Get current user
	userID := c.GetUint("user_id")
	
	// Get today's date
	today := time.Now().Format("2006-01-02")
	
	// Get or create today's performance record
	var performance entity.UserPerformance
	if err := database.DB.Where("user_id = ? AND date = ?", userID, today).
		First(&performance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default performance record
			performance = entity.UserPerformance{
				UserID:              userID,
				Date:                today,
				QuestionsAnswered:   0,
				QuestionsCorrect:    0,
				TimeSpent:           0,
				StreakDays:          calculateStreakDays(userID),
				WeeklyTarget:        50,
				HomeworkCompleted:   0,
				ReinforcementsEarned: 0,
			}
			database.DB.Create(&performance)
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch performance data",
			})
			return
		}
	}

	// Calculate accuracy rate
	var accuracyRate float64
	if performance.QuestionsAnswered > 0 {
		accuracyRate = float64(performance.QuestionsCorrect) / float64(performance.QuestionsAnswered) * 100
	}

	resp := response.UserPerformanceResponse{
		UserID:              performance.UserID,
		Date:                performance.Date,
		TodayLearned:        performance.QuestionsAnswered,
		QuestionsAnswered:   performance.QuestionsAnswered,
		QuestionsCorrect:    performance.QuestionsCorrect,
		TimeSpent:           performance.TimeSpent,
		Streak:              performance.StreakDays,
		WeeklyTarget:        performance.WeeklyTarget,
		HomeworkCompleted:   performance.HomeworkCompleted,
		ReinforcementsEarned: performance.ReinforcementsEarned,
		AccuracyRate:        accuracyRate,
		CreatedAt:           performance.CreatedAt,
		UpdatedAt:           performance.UpdatedAt,
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserAnswerHistory retrieves user's answer history
func GetUserAnswerHistory(c *gin.Context) {
	// Get current user
	userID := c.GetUint("user_id")
	
	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	answerType := c.Query("answer_type") // single or paper
	
	// Build query
	query := database.DB.Model(&entity.UserAnswer{}).
		Where("user_id = ?", userID).
		Preload("Question")
	
	if answerType != "" {
		query = query.Where("answer_type = ?", answerType)
	}
	
	// Count total
	var total int64
	query.Count(&total)
	
	// Get answers with pagination
	var answers []entity.UserAnswer
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&answers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch answer history",
		})
		return
	}

	// Convert to response format
	items := make([]map[string]interface{}, len(answers))
	for i, answer := range answers {
		items[i] = map[string]interface{}{
			"id":          answer.ID,
			"question_id": answer.QuestionID,
			"question":    answer.Question,
			"paper_id":    answer.PaperID,
			"answer":      answer.Answer,
			"score":       answer.Score,
			"is_correct":  answer.IsCorrect,
			"answer_type": answer.AnswerType,
			"created_at":  answer.CreatedAt,
		}
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// Helper function to calculate streak days
func calculateStreakDays(userID uint) int {
	var performances []entity.UserPerformance
	database.DB.Where("user_id = ? AND questions_answered > 0", userID).
		Order("date DESC").
		Find(&performances)
	
	if len(performances) == 0 {
		return 0
	}
	
	streak := 0
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	
	for i, perf := range performances {
		perfDate, _ := time.Parse("2006-01-02", perf.Date)
		expectedDate := today.AddDate(0, 0, -i)
		
		if perfDate.Equal(expectedDate) {
			streak++
		} else {
			break
		}
	}
	
	return streak
}
