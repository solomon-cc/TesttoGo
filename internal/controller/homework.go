package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/internal/model/response"
	"testogo/pkg/database"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateHomework creates a new homework assignment
func CreateHomework(c *gin.Context) {
	var req request.CreateHomeworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user (teacher/admin)
	userID := c.GetUint("userID")

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Serialize reinforcement settings
	reinforcementJSON, _ := json.Marshal(req.ReinforcementSettings)

	// Create homework
	homework := entity.Homework{
		Title:                 req.Title,
		Description:           req.Description,
		CreatorID:             userID,
		Grade:                 req.Grade,
		Subject:               req.Subject,
		Status:                entity.HomeworkStatusActive,
		ScheduleType:          entity.HomeworkScheduleType(req.ScheduleType),
		QuestionsPerDay:       req.QuestionsPerDay,
		ShowHints:             req.ShowHints,
		ReinforcementSettings: string(reinforcementJSON),
	}

	if err := tx.Create(&homework).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create homework",
		})
		return
	}

	// Create student assignments
	for _, assignment := range req.StudentAssignments {
		hwAssignment := entity.HomeworkAssignment{
			HomeworkID:             homework.ID,
			StudentID:              assignment.StudentID,
			ReinforcementSettingID: assignment.ReinforcementSettingID,
		}
		if err := tx.Create(&hwAssignment).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to create homework assignment",
			})
			return
		}
	}

	// Create homework questions
	for _, question := range req.Questions {
		hwQuestion := entity.HomeworkQuestion{
			HomeworkID: homework.ID,
			QuestionID: question.QuestionID,
			DayOfWeek:  question.DayOfWeek,
			Order:      question.Order,
		}
		if err := tx.Create(&hwQuestion).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to create homework question",
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to save homework",
		})
		return
	}

	// Load complete homework data for response
	var homeworkResp entity.Homework
	database.DB.Preload("Creator").
		Preload("HomeworkAssignments").
		Preload("HomeworkQuestions").
		First(&homeworkResp, homework.ID)

	c.JSON(http.StatusCreated, convertToHomeworkResponse(&homeworkResp))
}

// ListHomework lists homework assignments with filtering
func ListHomework(c *gin.Context) {
	var req request.ListHomeworkRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
		})
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// Get current user info
	userID := c.GetUint("userID")
	userRole := c.GetString("user_role")

	// Build query
	query := database.DB.Model(&entity.Homework{}).Preload("Creator")
	
	// Apply role-based filtering
	if userRole == "user" { // Student
		// Only show homework assigned to this student
		query = query.Joins("JOIN homework_assignments ON homework_assignments.homework_id = homework.id").
			Where("homework_assignments.student_id = ?", userID)
	} else if userRole == "teacher" {
		// Teachers see only their own homework unless AdminView is specified
		if !req.AdminView {
			query = query.Where("creator_id = ?", userID)
		}
		// If AdminView is true but user is teacher, still restrict to own homework
		if req.AdminView {
			query = query.Where("creator_id = ?", userID)
		}
	} else if userRole == "admin" {
		// Admin can see all homework when AdminView is true, otherwise their own
		if !req.AdminView {
			query = query.Where("creator_id = ?", userID)
		}
		// If AdminView is true, show all homework (no additional filter)
	}

	// Apply filters
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Grade != "" {
		query = query.Where("grade = ?", req.Grade)
	}
	if req.Subject != "" {
		query = query.Where("subject = ?", req.Subject)
	}
	if req.CreatorID != 0 && userRole != "user" {
		query = query.Where("creator_id = ?", req.CreatorID)
	}
	if req.DateFrom != "" {
		query = query.Where("start_date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("end_date <= ?", req.DateTo)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Apply pagination
	offset := (req.Page - 1) * req.PageSize
	var homework []entity.Homework
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&homework).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch homework",
		})
		return
	}

	// Convert to response
	items := make([]response.HomeworkResponse, len(homework))
	for i, hw := range homework {
		if userRole == "user" {
			// For students, include completion status
			items[i] = convertToHomeworkResponseWithCompletion(&hw, userID)
		} else {
			// For teachers and admins, use basic conversion
			items[i] = convertToHomeworkResponse(&hw)
		}
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	c.JSON(http.StatusOK, response.HomeworkListResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// GetHomework retrieves a specific homework assignment
func GetHomework(c *gin.Context) {
	homeworkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid homework ID",
		})
		return
	}

	var homework entity.Homework
	if err := database.DB.Preload("Creator").
		Preload("HomeworkAssignments").
		Preload("HomeworkAssignments.Student").
		Preload("HomeworkAssignments.ReinforcementSetting").
		Preload("HomeworkQuestions").
		Preload("HomeworkQuestions.Question").
		Preload("HomeworkSubmissions").
		First(&homework, homeworkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Homework not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch homework",
			})
		}
		return
	}

	// Check access permissions
	userID := c.GetUint("userID")
	userRole := c.GetString("user_role")
	
	if userRole == "user" {
		// Check if student is assigned to this homework
		assigned := false
		for _, assignment := range homework.HomeworkAssignments {
			if assignment.StudentID == userID {
				assigned = true
				break
			}
		}
		if !assigned {
			c.JSON(http.StatusForbidden, response.ErrorResponse{
				Error: "Access denied",
			})
			return
		}
	} else if userRole == "teacher" && homework.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
		})
		return
	}
	// Admin has access to all homework

	if userRole == "user" {
		// For students, include completion status
		c.JSON(http.StatusOK, convertToHomeworkResponseWithCompletion(&homework, userID))
	} else {
		// For teachers and admins, use basic conversion
		c.JSON(http.StatusOK, convertToHomeworkResponse(&homework))
	}
}

// UpdateHomework updates an existing homework assignment
func UpdateHomework(c *gin.Context) {
	homeworkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid homework ID",
		})
		return
	}

	var req request.UpdateHomeworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user
	userID := c.GetUint("userID")
	userRole := c.GetString("user_role")

	// Find homework
	var homework entity.Homework
	if err := database.DB.First(&homework, homeworkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Homework not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch homework",
			})
		}
		return
	}

	// Check permissions
	if userRole == "teacher" && homework.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
		})
		return
	}
	// Admin has access to all homework

	// Update fields
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	// End date update removed - no longer needed
	if req.QuestionsPerDay != nil {
		updates["questions_per_day"] = *req.QuestionsPerDay
	}
	if req.ShowHints != nil {
		updates["show_hints"] = *req.ShowHints
	}
	if req.ReinforcementSettings != nil {
		reinforcementJSON, _ := json.Marshal(*req.ReinforcementSettings)
		updates["reinforcement_settings"] = string(reinforcementJSON)
	}

	if err := database.DB.Model(&homework).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to update homework",
		})
		return
	}

	// Reload homework with relations
	database.DB.Preload("Creator").
		Preload("HomeworkAssignments").
		Preload("HomeworkQuestions").
		First(&homework, homeworkID)

	c.JSON(http.StatusOK, convertToHomeworkResponse(&homework))
}

// DeleteHomework deletes a homework assignment
func DeleteHomework(c *gin.Context) {
	homeworkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid homework ID",
		})
		return
	}

	// Get current user
	userID := c.GetUint("userID")
	userRole := c.GetString("user_role")

	// Find homework
	var homework entity.Homework
	if err := database.DB.First(&homework, homeworkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Homework not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch homework",
			})
		}
		return
	}

	// Check permissions
	if userRole == "teacher" && homework.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
		})
		return
	}
	// Admin has access to all homework

	// Soft delete
	if err := database.DB.Delete(&homework).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to delete homework",
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Homework deleted successfully",
	})
}

// CopyHomework creates a copy of existing homework
func CopyHomework(c *gin.Context) {
	sourceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid homework ID",
		})
		return
	}

	var req request.CopyHomeworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user
	userID := c.GetUint("userID")

	// Find source homework
	var sourceHomework entity.Homework
	if err := database.DB.Preload("HomeworkQuestions").
		First(&sourceHomework, sourceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Source homework not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch source homework",
			})
		}
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create new homework
	newHomework := entity.Homework{
		Title:                 req.NewTitle,
		Description:           sourceHomework.Description,
		CreatorID:             userID,
		Grade:                 sourceHomework.Grade,
		Subject:               sourceHomework.Subject,
		Status:                entity.HomeworkStatusDraft,
		ScheduleType:          sourceHomework.ScheduleType,
		QuestionsPerDay:       sourceHomework.QuestionsPerDay,
		ShowHints:             sourceHomework.ShowHints,
		ReinforcementSettings: sourceHomework.ReinforcementSettings,
	}

	if err := tx.Create(&newHomework).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create homework copy",
		})
		return
	}

	// Copy questions if requested
	if req.CopyQuestions {
		for _, question := range sourceHomework.HomeworkQuestions {
			newQuestion := entity.HomeworkQuestion{
				HomeworkID: newHomework.ID,
				QuestionID: question.QuestionID,
				DayOfWeek:  question.DayOfWeek,
				Order:      question.Order,
			}
			if err := tx.Create(&newQuestion).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Error: "Failed to copy homework questions",
				})
				return
			}
		}
	}

	// Copy or assign specific students
	if req.CopyStudents || len(req.StudentIDs) > 0 {
		studentIDs := req.StudentIDs
		if req.CopyStudents && len(studentIDs) == 0 {
			// Get original student assignments
			var assignments []entity.HomeworkAssignment
			if err := tx.Where("homework_id = ?", sourceID).Find(&assignments).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Error: "Failed to fetch original assignments",
				})
				return
			}
			for _, assignment := range assignments {
				studentIDs = append(studentIDs, assignment.StudentID)
			}
		}

		// Create new assignments
		for _, studentID := range studentIDs {
			assignment := entity.HomeworkAssignment{
				HomeworkID: newHomework.ID,
				StudentID:  studentID,
			}
			if err := tx.Create(&assignment).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Error: "Failed to create homework assignments",
				})
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to save homework copy",
		})
		return
	}

	// Load complete homework data for response
	var homeworkResp entity.Homework
	database.DB.Preload("Creator").
		Preload("HomeworkAssignments").
		Preload("HomeworkQuestions").
		First(&homeworkResp, newHomework.ID)

	c.JSON(http.StatusCreated, convertToHomeworkResponse(&homeworkResp))
}

// SubmitHomework handles student homework submission
func SubmitHomework(c *gin.Context) {
	var req request.SubmitHomeworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user (student)
	userID := c.GetUint("userID")

	// Verify homework exists and student is assigned
	var homework entity.Homework
	if err := database.DB.Preload("HomeworkAssignments").
		First(&homework, req.HomeworkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Homework not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch homework",
			})
		}
		return
	}

	// Check if student is assigned
	assigned := false
	for _, assignment := range homework.HomeworkAssignments {
		if assignment.StudentID == userID {
			assigned = true
			break
		}
	}
	if !assigned {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "You are not assigned to this homework",
		})
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate score and correctness
	totalQuestions := len(req.QuestionAnswers)
	correctAnswers := 0
	
	// Create submission record
	submission := entity.HomeworkSubmission{
		HomeworkID:       req.HomeworkID,
		StudentID:        userID,
		SubmissionDate:   req.SubmissionDate,
		QuestionsTotal:   totalQuestions,
		TimeSpent:        req.TimeSpent,
		IsCompleted:      true,
	}

	if err := tx.Create(&submission).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create submission",
		})
		return
	}

	// Process each answer
	for _, answer := range req.QuestionAnswers {
		// Get correct answer
		var question entity.Question
		if err := tx.First(&question, answer.QuestionID).Error; err != nil {
			continue
		}

		// Check if answer is correct (simplified - real logic would be more complex)
		isCorrect := answer.Answer == question.Answer
		if isCorrect {
			correctAnswers++
		}

		// Create answer record
		questionAnswer := entity.HomeworkQuestionAnswer{
			SubmissionID: submission.ID,
			QuestionID:   answer.QuestionID,
			Answer:       answer.Answer,
			IsCorrect:    isCorrect,
			TimeSpent:    answer.TimeSpent,
		}

		if err := tx.Create(&questionAnswer).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to save answer",
			})
			return
		}
	}

	// Update submission with final score
	score := int(float64(correctAnswers) / float64(totalQuestions) * 100)
	if err := tx.Model(&submission).Updates(map[string]interface{}{
		"questions_correct": correctAnswers,
		"score":            score,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to update submission score",
		})
		return
	}

	// Update user performance stats
	today := time.Now().Format("2006-01-02")
	var performance entity.UserPerformance
	if err := tx.Where("user_id = ? AND date = ?", userID, today).
		First(&performance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new performance record
			performance = entity.UserPerformance{
				UserID:              userID,
				Date:                today,
				QuestionsAnswered:   totalQuestions,
				QuestionsCorrect:    correctAnswers,
				TimeSpent:           req.TimeSpent,
				HomeworkCompleted:   1,
			}
			tx.Create(&performance)
		}
	} else {
		// Update existing record
		tx.Model(&performance).Updates(map[string]interface{}{
			"questions_answered":   performance.QuestionsAnswered + totalQuestions,
			"questions_correct":    performance.QuestionsCorrect + correctAnswers,
			"time_spent":          performance.TimeSpent + req.TimeSpent,
			"homework_completed":   performance.HomeworkCompleted + 1,
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to save submission",
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Message: "Homework submitted successfully",
		Data: map[string]interface{}{
			"score":             score,
			"correct_answers":   correctAnswers,
			"total_questions":   totalQuestions,
			"submission_id":     submission.ID,
		},
	})
}

// Helper function to convert entity to response
func convertToHomeworkResponse(hw *entity.Homework) response.HomeworkResponse {
	var reinforcementSettings map[string]interface{}
	if hw.ReinforcementSettings != "" {
		json.Unmarshal([]byte(hw.ReinforcementSettings), &reinforcementSettings)
	}

	resp := response.HomeworkResponse{
		ID:                    hw.ID,
		Title:                 hw.Title,
		Description:           hw.Description,
		CreatorID:             hw.CreatorID,
		Grade:                 hw.Grade,
		Subject:               hw.Subject,
		Status:                string(hw.Status),
		ScheduleType:          string(hw.ScheduleType),
		StartDate:             hw.StartDate,
		EndDate:               hw.EndDate,
		QuestionsPerDay:       hw.QuestionsPerDay,
		ShowHints:             hw.ShowHints,
		ReinforcementSettings: reinforcementSettings,
		IsCompleted:           false, // Default value, will be set by caller if needed
		CreatedAt:             hw.CreatedAt,
		UpdatedAt:             hw.UpdatedAt,
	}

	if hw.Creator.Username != "" {
		resp.CreatorName = hw.Creator.Username
		resp.TeacherName = hw.Creator.Username // 为了前端兼容性，同时设置 teacher_name 字段
	}

	// Convert assignments
	if len(hw.HomeworkAssignments) > 0 {
		resp.Assignments = make([]response.HomeworkAssignmentResponse, len(hw.HomeworkAssignments))
		for i, assignment := range hw.HomeworkAssignments {
			resp.Assignments[i] = response.HomeworkAssignmentResponse{
				ID:                     assignment.ID,
				StudentID:              assignment.StudentID,
				ReinforcementSettingID: assignment.ReinforcementSettingID,
				CreatedAt:              assignment.CreatedAt,
			}
			if assignment.Student.Username != "" {
				resp.Assignments[i].StudentName = assignment.Student.Username
			}
		}
	}

	// Convert questions
	if len(hw.HomeworkQuestions) > 0 {
		resp.Questions = make([]response.HomeworkQuestionResponse, len(hw.HomeworkQuestions))
		for i, question := range hw.HomeworkQuestions {
			resp.Questions[i] = response.HomeworkQuestionResponse{
				ID:         question.ID,
				QuestionID: question.QuestionID,
				DayOfWeek:  question.DayOfWeek,
				Order:      question.Order,
				CreatedAt:  question.CreatedAt,
			}
		}
	}

	return resp
}

// Helper function to convert entity to response with completion status for specific user
func convertToHomeworkResponseWithCompletion(hw *entity.Homework, userID uint) response.HomeworkResponse {
	resp := convertToHomeworkResponse(hw)

	// Check if the homework is completed by the current user
	if userID > 0 {
		var submission entity.HomeworkSubmission
		err := database.DB.Where("homework_id = ? AND student_id = ? AND is_completed = ?",
			hw.ID, userID, true).First(&submission).Error
		resp.IsCompleted = (err == nil)
	}

	return resp
}

// GetHomeworkSubmissions retrieves submissions for a specific homework
func GetHomeworkSubmissions(c *gin.Context) {
	homeworkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid homework ID",
		})
		return
	}

	var req request.HomeworkSubmissionQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
		})
		return
	}

	// Set defaults
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	// Verify homework exists and user has access
	userID := c.GetUint("userID")
	userRole := c.GetString("user_role")
	
	var homework entity.Homework
	if err := database.DB.First(&homework, homeworkID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error: "Homework not found",
		})
		return
	}

	// Check permissions
	if userRole == "teacher" && homework.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
		})
		return
	}
	// Admin has access to all homework

	// Build query
	query := database.DB.Model(&entity.HomeworkSubmission{}).
		Where("homework_id = ?", homeworkID).
		Preload("Student")

	// Apply filters
	if req.StudentID != 0 {
		query = query.Where("student_id = ?", req.StudentID)
	}
	if req.IsCompleted != nil {
		query = query.Where("is_completed = ?", *req.IsCompleted)
	}
	if req.DateFrom != "" {
		query = query.Where("submission_date >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("submission_date <= ?", req.DateTo)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get submissions with pagination
	offset := (req.Page - 1) * req.PageSize
	var submissions []entity.HomeworkSubmission
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("submission_date DESC").
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch submissions",
		})
		return
	}

	// Convert to response
	items := make([]response.HomeworkSubmissionResponse, len(submissions))
	for i, submission := range submissions {
		items[i] = response.HomeworkSubmissionResponse{
			ID:               submission.ID,
			StudentID:        submission.StudentID,
			StudentName:      submission.Student.Username,
			SubmissionDate:   submission.SubmissionDate,
			QuestionsTotal:   submission.QuestionsTotal,
			QuestionsCorrect: submission.QuestionsCorrect,
			TimeSpent:        submission.TimeSpent,
			Score:            submission.Score,
			IsCompleted:      submission.IsCompleted,
			CreatedAt:        submission.CreatedAt,
			UpdatedAt:        submission.UpdatedAt,
		}
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"total":       total,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"total_pages": totalPages,
	})
}

// AdjustHomework allows teachers to make mid-assignment adjustments
func AdjustHomework(c *gin.Context) {
	homeworkID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid homework ID",
		})
		return
	}

	var req request.AdjustHomeworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user
	userID := c.GetUint("userID")
	userRole := c.GetString("user_role")

	// Verify homework exists and user has access
	var homework entity.Homework
	if err := database.DB.First(&homework, homeworkID).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error: "Homework not found",
		})
		return
	}

	// Check permissions
	if userRole == "teacher" && homework.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
		})
		return
	}
	// Admin has access to all homework

	// Serialize changes
	changesJSON, _ := json.Marshal(req.Changes)

	// Create adjustment record
	adjustment := entity.HomeworkAdjustment{
		HomeworkID:  uint(homeworkID),
		TeacherID:   userID,
		AdjustType:  req.AdjustType,
		Description: req.Description,
		Changes:     string(changesJSON),
	}

	if err := database.DB.Create(&adjustment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create adjustment record",
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Message: "Homework adjustment recorded successfully",
		Data: map[string]interface{}{
			"adjustment_id": adjustment.ID,
		},
	})
}

// GetHomeworkHistory retrieves homework history for copying
func GetHomeworkHistory(c *gin.Context) {
	// Get current user
	userID := c.GetUint("userID")
	userRole := c.GetString("user_role")

	// Build query
	query := database.DB.Model(&entity.Homework{}).
		Preload("Creator")

	// Apply role-based filtering
	if userRole == "teacher" {
		query = query.Where("creator_id = ?", userID)
	}
	// Admin can see all homework

	// Get recent homework for copying
	var homework []entity.Homework
	if err := query.Order("created_at DESC").
		Limit(50). // Recent 50 homework assignments
		Find(&homework).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch homework history",
		})
		return
	}

	// Convert to response
	items := make([]response.HomeworkResponse, len(homework))
	for i, hw := range homework {
		items[i] = convertToHomeworkResponse(&hw)
	}

	c.JSON(http.StatusOK, items)
}