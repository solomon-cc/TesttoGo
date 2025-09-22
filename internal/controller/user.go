package controller

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"testogo/internal/model/entity"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
	userID := c.GetUint("userID")

	// Calculate total statistics from user answers
	var totalQuestions int64
	var totalCorrect int64
	database.DB.Model(&entity.UserAnswer{}).Where("user_id = ?", userID).Count(&totalQuestions)
	database.DB.Model(&entity.UserAnswer{}).Where("user_id = ? AND is_correct = ?", userID, true).Count(&totalCorrect)

	// Calculate average accuracy
	var averageAccuracy float64
	if totalQuestions > 0 {
		averageAccuracy = float64(totalCorrect) / float64(totalQuestions) * 100
	}

	// Calculate total study time from all user performance records
	var totalStudyTime int64
	database.DB.Model(&entity.UserPerformance{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(time_spent), 0)").
		Scan(&totalStudyTime)

	// Calculate study days (days with questions answered)
	var studyDays int64
	database.DB.Model(&entity.UserPerformance{}).
		Where("user_id = ? AND questions_answered > 0", userID).
		Count(&studyDays)

	// Get type analysis - statistics by question type
	typeAnalysis := []map[string]interface{}{}
	rows, err := database.DB.Raw(`
		SELECT
			q.type,
			COUNT(*) as total,
			SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct,
			ROUND(SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 0) as accuracy
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		WHERE ua.user_id = ? AND ua.deleted_at IS NULL AND q.deleted_at IS NULL
		GROUP BY q.type
		ORDER BY total DESC
	`, userID).Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var qType string
			var total, correct, accuracy int
			rows.Scan(&qType, &total, &correct, &accuracy)
			typeAnalysis = append(typeAnalysis, map[string]interface{}{
				"type":     qType,
				"total":    total,
				"correct":  correct,
				"accuracy": accuracy,
			})
		}
	}

	// Get difficulty analysis - statistics by difficulty level
	difficultyAnalysis := []map[string]interface{}{}
	rows2, err := database.DB.Raw(`
		SELECT
			CASE
				WHEN q.difficulty <= 2 THEN 'easy'
				WHEN q.difficulty <= 4 THEN 'medium'
				ELSE 'hard'
			END as difficulty_level,
			COUNT(*) as total,
			SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct,
			ROUND(SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 0) as accuracy
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		WHERE ua.user_id = ? AND ua.deleted_at IS NULL AND q.deleted_at IS NULL
		GROUP BY difficulty_level
		ORDER BY
			CASE difficulty_level
				WHEN 'easy' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'hard' THEN 3
			END
	`, userID).Rows()

	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var difficulty string
			var total, correct, accuracy int
			rows2.Scan(&difficulty, &total, &correct, &accuracy)
			difficultyAnalysis = append(difficultyAnalysis, map[string]interface{}{
				"difficulty": difficulty,
				"total":      total,
				"correct":    correct,
				"accuracy":   accuracy,
			})
		}
	}

	// Get weakness areas - topics/subjects with low accuracy (< 70%)
	weaknessAreas := []map[string]interface{}{}
	rows3, err := database.DB.Raw(`
		SELECT
			COALESCE(q.topic, '未分类') as name,
			COUNT(*) as total,
			SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct,
			ROUND(SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 0) as accuracy
		FROM user_answers ua
		JOIN questions q ON ua.question_id = q.id
		WHERE ua.user_id = ? AND ua.deleted_at IS NULL AND q.deleted_at IS NULL
		GROUP BY q.topic
		HAVING COUNT(*) >= 5 AND accuracy < 70
		ORDER BY accuracy ASC, total DESC
		LIMIT 6
	`, userID).Rows()

	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var name string
			var total, correct, accuracy int
			rows3.Scan(&name, &total, &correct, &accuracy)

			// Generate description based on topic name
			description := generateWeaknessDescription(name, accuracy)

			weaknessAreas = append(weaknessAreas, map[string]interface{}{
				"id":          len(weaknessAreas) + 1,
				"name":        name,
				"accuracy":    accuracy,
				"total":       total,
				"correct":     correct,
				"description": description,
			})
		}
	}

	// Generate learning suggestions based on performance
	suggestions := generateLearningSuggestions(averageAccuracy, int(totalQuestions), weaknessAreas)

	// Prepare response
	resp := map[string]interface{}{
		"total_questions":     totalQuestions,
		"average_accuracy":    averageAccuracy,
		"total_study_time":    totalStudyTime,
		"study_days":          studyDays,
		"type_analysis":       typeAnalysis,
		"difficulty_analysis": difficultyAnalysis,
		"weakness_areas":      weaknessAreas,
		"suggestions":         suggestions,
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserAnswerHistory retrieves user's answer history grouped by sessions
func GetUserAnswerHistory(c *gin.Context) {
	// Get current user info
	currentUserID := c.GetUint("userID")
	userRole, _ := c.Get("userRole")

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	answerType := c.Query("answer_type") // single or paper
	targetUserID := c.Query("user_id")   // 可选：指定查看特定用户的历史（仅老师/管理员）

	// 确定要查询的用户ID
	var queryUserID uint
	if userRole == "teacher" || userRole == "admin" {
		// 老师和管理员可以查看所有用户的历史
		if targetUserID != "" {
			// 如果指定了user_id参数，查看指定用户的历史
			if uid, err := strconv.ParseUint(targetUserID, 10, 32); err == nil {
				queryUserID = uint(uid)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
				return
			}
		} else {
			// 如果没有指定user_id，查看所有用户的历史（设置为0表示查看所有）
			queryUserID = 0
		}
	} else {
		// 学生只能查看自己的历史
		queryUserID = currentUserID
	}

	// For paper-based answers, group by paper_id
	// For single answers, group by time windows (same day or within 1 hour)
	var sessions []map[string]interface{}

	if answerType == "paper" || answerType == "" {
		// Get paper-based sessions
		paperSessions := getPaperSessions(queryUserID)
		sessions = append(sessions, paperSessions...)
	}

	if answerType == "single" || answerType == "" {
		// Get single answer sessions (grouped by time)
		singleSessions := getSingleAnswerSessions(queryUserID)
		sessions = append(sessions, singleSessions...)
	}

	// Sort sessions by created_at DESC
	sort.Slice(sessions, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, sessions[i]["created_at"].(string))
		timeJ, _ := time.Parse(time.RFC3339, sessions[j]["created_at"].(string))
		return timeI.After(timeJ)
	})

	// Apply pagination
	total := len(sessions)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		sessions = []map[string]interface{}{}
	} else {
		if end > total {
			end = total
		}
		sessions = sessions[start:end]
	}

	totalPages := (total + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"items":       sessions,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// getPaperSessions gets sessions based on paper submissions
func getPaperSessions(userID uint) []map[string]interface{} {
	var results []map[string]interface{}

	// Build query based on userID (0 means all users)
	var query string
	var args []interface{}

	if userID == 0 {
		// Query all users
		query = `
			SELECT
				ua.user_id,
				ua.paper_id,
				COUNT(*) as question_count,
				SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct_count,
				MAX(ua.created_at) as created_at,
				'paper' as answer_type,
				u.username as student_name
			FROM user_answers ua
			JOIN users u ON ua.user_id = u.id
			WHERE ua.paper_id > 0 AND ua.deleted_at IS NULL
			GROUP BY ua.user_id, ua.paper_id
			ORDER BY MAX(ua.created_at) DESC
		`
		args = []interface{}{}
	} else {
		// Query specific user
		query = `
			SELECT
				ua.user_id,
				ua.paper_id,
				COUNT(*) as question_count,
				SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct_count,
				MAX(ua.created_at) as created_at,
				'paper' as answer_type,
				u.username as student_name
			FROM user_answers ua
			JOIN users u ON ua.user_id = u.id
			WHERE ua.user_id = ? AND ua.paper_id > 0 AND ua.deleted_at IS NULL
			GROUP BY ua.user_id, ua.paper_id
			ORDER BY MAX(ua.created_at) DESC
		`
		args = []interface{}{userID}
	}

	rows, err := database.DB.Raw(query, args...).Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var userIDResult, paperID, questionCount, correctCount int
		var createdAt time.Time
		var answerType, studentName string

		rows.Scan(&userIDResult, &paperID, &questionCount, &correctCount, &createdAt, &answerType, &studentName)

		accuracyRate := 0
		if questionCount > 0 {
			accuracyRate = (correctCount * 100) / questionCount
		}

		var sessionID string
		var title string
		if userID == 0 {
			// For all users view, include user info in ID and title
			sessionID = fmt.Sprintf("paper_%d_user_%d", paperID, userIDResult)
			title = fmt.Sprintf("%s - 试卷练习 #%d", studentName, paperID)
		} else {
			// For single user view, simpler format
			sessionID = fmt.Sprintf("paper_%d", paperID)
			title = fmt.Sprintf("试卷练习 #%d", paperID)
		}

		results = append(results, map[string]interface{}{
			"id":            sessionID,
			"user_id":       userIDResult,
			"paper_id":      paperID,
			"title":         title,
			"question_count": questionCount,
			"correct_count":  correctCount,
			"accuracy_rate":  accuracyRate,
			"answer_type":    answerType,
			"created_at":     createdAt.Format(time.RFC3339),
			"student_name":   studentName,
		})
	}

	return results
}

// getSingleAnswerSessions gets sessions for single practice answers grouped by time windows
func getSingleAnswerSessions(userID uint) []map[string]interface{} {
	var results []map[string]interface{}

	// Build query based on userID (0 means all users)
	var query string
	var args []interface{}

	if userID == 0 {
		// Query all users, group by user and date
		query = `
			SELECT
				ua.user_id,
				DATE(ua.created_at) as practice_date,
				COUNT(*) as question_count,
				SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct_count,
				MAX(ua.created_at) as created_at,
				'single' as answer_type,
				u.username as student_name
			FROM user_answers ua
			JOIN users u ON ua.user_id = u.id
			WHERE (ua.paper_id = 0 OR ua.paper_id IS NULL) AND ua.deleted_at IS NULL
			GROUP BY ua.user_id, DATE(ua.created_at)
			ORDER BY practice_date DESC
		`
		args = []interface{}{}
	} else {
		// Query specific user
		query = `
			SELECT
				ua.user_id,
				DATE(ua.created_at) as practice_date,
				COUNT(*) as question_count,
				SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct_count,
				MAX(ua.created_at) as created_at,
				'single' as answer_type,
				u.username as student_name
			FROM user_answers ua
			JOIN users u ON ua.user_id = u.id
			WHERE ua.user_id = ? AND (ua.paper_id = 0 OR ua.paper_id IS NULL) AND ua.deleted_at IS NULL
			GROUP BY ua.user_id, DATE(ua.created_at)
			ORDER BY practice_date DESC
		`
		args = []interface{}{userID}
	}

	rows, err := database.DB.Raw(query, args...).Rows()

	if err != nil {
		return results
	}
	defer rows.Close()

	for rows.Next() {
		var userIDResult int
		var practiceDate string
		var questionCount, correctCount int
		var createdAt time.Time
		var answerType, studentName string

		rows.Scan(&userIDResult, &practiceDate, &questionCount, &correctCount, &createdAt, &answerType, &studentName)

		accuracyRate := 0
		if questionCount > 0 {
			accuracyRate = (correctCount * 100) / questionCount
		}

		var sessionID string
		var title string
		if userID == 0 {
			// For all users view, include user info in ID and title
			sessionID = fmt.Sprintf("single_%s_user_%d", practiceDate, userIDResult)
			title = fmt.Sprintf("%s - 随机练习 %s", studentName, practiceDate)
		} else {
			// For single user view, simpler format
			sessionID = fmt.Sprintf("single_%s", practiceDate)
			title = fmt.Sprintf("随机练习 - %s", practiceDate)
		}

		results = append(results, map[string]interface{}{
			"id":            sessionID,
			"user_id":       userIDResult,
			"practice_date": practiceDate,
			"title":         title,
			"question_count": questionCount,
			"correct_count":  correctCount,
			"accuracy_rate":  accuracyRate,
			"answer_type":    answerType,
			"created_at":     createdAt.Format(time.RFC3339),
			"student_name":   studentName,
		})
	}

	return results
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

// Helper function to generate weakness description based on topic
func generateWeaknessDescription(topicName string, accuracy int) string {
	descriptions := map[string]string{
		"addition":       "加法运算需要多加练习，建议从简单的一位数加法开始复习",
		"subtraction":    "减法运算掌握不够熟练，可以通过实物演示来加深理解",
		"multiplication": "乘法口诀需要加强记忆，建议每天背诵乘法表",
		"division":       "除法概念理解有待提高，可以通过分组游戏来练习",
		"fractions":      "分数概念需要强化，建议通过画图和实物来理解分数的含义",
		"geometry":       "几何图形识别能力需要提升，多观察生活中的图形",
		"measurement":    "测量单位换算需要加强练习，记住常用单位之间的关系",
		"reading":        "阅读理解能力有待提高，建议多读课外书籍",
		"vocabulary":     "词汇量需要扩充，每天学习3-5个新词汇",
		"grammar":        "语法规则掌握不够牢固，需要系统复习语法知识",
		"writing":        "写作表达能力需要提升，多练习造句和作文",
	}

	if description, exists := descriptions[topicName]; exists {
		return description
	}

	// Generate default description based on accuracy
	if accuracy < 40 {
		return "这个知识点掌握情况较差，建议从基础开始重新学习"
	} else if accuracy < 60 {
		return "这个知识点需要多加练习，可以通过专项训练来提高"
	} else {
		return "这个知识点基本掌握，再多练习几道题就能完全掌握了"
	}
}

// Helper function to generate learning suggestions based on performance
func generateLearningSuggestions(accuracy float64, totalQuestions int, weaknessAreas []map[string]interface{}) []map[string]interface{} {
	suggestions := []map[string]interface{}{}

	// Suggestion based on overall accuracy
	if accuracy < 60 {
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "加强基础练习",
			"content": "您的整体正确率偏低，建议从基础题目开始，循序渐进地提高答题准确率。每天坚持练习20-30道基础题目。",
		})
	} else if accuracy < 80 {
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "提高答题技巧",
			"content": "您的基础不错，可以尝试一些解题技巧和方法，提高答题效率和准确率。建议多总结错题经验。",
		})
	} else {
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "挑战更高难度",
			"content": "您的基础很扎实，可以尝试挑战更高难度的题目，或者帮助其他小朋友一起学习进步。",
		})
	}

	// Suggestion based on total questions
	if totalQuestions < 50 {
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "增加练习量",
			"content": "建议每天增加练习时间，多做一些题目来巩固所学知识。熟能生巧，练习越多掌握越牢固。",
		})
	} else if totalQuestions > 500 {
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "注重学习效率",
			"content": "您的练习量很大，现在可以更注重学习质量和效率，专攻薄弱环节，做到精益求精。",
		})
	}

	// Suggestion based on weakness areas
	if len(weaknessAreas) > 3 {
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "专项突破薄弱点",
			"content": "发现您有多个薄弱知识点，建议每天选择1-2个薄弱点进行专项练习，逐个击破。",
		})
	} else if len(weaknessAreas) > 0 {
		topWeakness := weaknessAreas[0]["name"].(string)
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "重点关注" + topWeakness,
			"content": "建议重点练习" + topWeakness + "相关题目，可以寻求老师或家长的帮助来理解相关概念。",
		})
	} else {
		suggestions = append(suggestions, map[string]interface{}{
			"title":   "保持学习节奏",
			"content": "您各方面掌握都很均衡，继续保持现在的学习节奏，定期复习已学内容，预习新知识。",
		})
	}

	// General motivation suggestion
	suggestions = append(suggestions, map[string]interface{}{
		"title":   "坚持每日学习",
		"content": "学习贵在坚持，建议每天固定时间进行学习，养成良好的学习习惯。记住：每一次练习都是进步！",
	})

	return suggestions
}
