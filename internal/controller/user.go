package controller

import (
	"net/http"
	"strconv"
	"time"

	"testogo/internal/model/entity"
	"testogo/internal/model/response"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// @Summary 获取用户列表
// @Description 获取用户列表，支持分页和筛选
// @Tags 用户
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param page query integer false "页码，默认1"
// @Param page_size query integer false "每页数量，默认10"
// @Param role query string false "角色筛选，可选值：user, teacher, admin"
// @Param status query string false "状态筛选，可选值：active"
// @Success 200 {object} map[string]interface{} "用户列表和总数"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users [get]
func ListUsers(c *gin.Context) {
	var users []entity.User
	query := database.DB.Order("id desc")

	// 筛选参数
	role := c.Query("role")
	status := c.Query("status")

	// 应用筛选条件
	if role != "" {
		query = query.Where("role = ?", role)
	}
	
	// status筛选（假设active表示正常用户，可以根据实际业务逻辑调整）
	if status == "active" {
		// 这里可以添加状态筛选逻辑，比如筛选未被删除的用户
		// 由于使用了gorm的软删除，deleted_at为null的就是active状态
		query = query.Where("deleted_at IS NULL")
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	
	// 计算总数（应用相同的筛选条件）
	var total int64
	countQuery := database.DB.Model(&entity.User{})
	if role != "" {
		countQuery = countQuery.Where("role = ?", role)
	}
	if status == "active" {
		countQuery = countQuery.Where("deleted_at IS NULL")
	}
	countQuery.Count(&total)

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

// @Summary 创建用户
// @Description 管理员创建新用户
// @Tags 用户
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param request body map[string]string true "用户信息 {\"username\": \"用户名\", \"password\": \"密码\", \"role\": \"user|teacher|admin\"}"
// @Success 201 {object} map[string]interface{} "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 409 {object} map[string]interface{} "用户名已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users [post]
func CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role" binding:"required,oneof=user teacher admin"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	var existingUser entity.User
	if err := database.DB.Where("username = ?", input.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	// Create user
	user := entity.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Role:     entity.Role(input.Role),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// Return user info without password
	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
	})
}

// @Summary 更新用户信息
// @Description 管理员更新用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param id path integer true "用户ID"
// @Param request body map[string]string true "用户信息 {\"username\": \"用户名\", \"password\": \"密码\", \"role\": \"user|teacher|admin\"}"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "用户不存在"
// @Failure 409 {object} map[string]interface{} "用户名已存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users/{id} [put]
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Username string `json:"username,omitempty" binding:"omitempty,min=3,max=20"`
		Password string `json:"password,omitempty" binding:"omitempty,min=6"`
		Role     string `json:"role,omitempty" binding:"omitempty,oneof=user teacher admin"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var user entity.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// Prepare update data
	updateData := make(map[string]interface{})

	// Update username if provided
	if input.Username != "" && input.Username != user.Username {
		// Check if new username already exists
		var existingUser entity.User
		if err := database.DB.Where("username = ? AND id != ?", input.Username, id).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		updateData["username"] = input.Username
	}

	// Update password if provided
	if input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}
		updateData["password"] = string(hashedPassword)
	}

	// Update role if provided
	if input.Role != "" {
		updateData["role"] = input.Role
	}

	// Perform update if there's data to update
	if len(updateData) > 0 {
		if err := database.DB.Model(&user).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
			return
		}
	}

	// Fetch updated user info
	if err := database.DB.Select("id, username, role, created_at, updated_at").First(&user, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
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
