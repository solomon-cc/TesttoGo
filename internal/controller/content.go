package controller

import (
	"net/http"
	"strconv"
	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/internal/model/response"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateGrade creates a new grade
func CreateGrade(c *gin.Context) {
	var req request.CreateGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate age range
	if req.AgeMax <= req.AgeMin {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Maximum age must be greater than minimum age",
		})
		return
	}

	// Check if code already exists
	var existing entity.Grade
	if err := database.DB.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Grade code already exists",
		})
		return
	}

	// Create grade
	grade := entity.Grade{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		AgeMin:      req.AgeMin,
		AgeMax:      req.AgeMax,
		Order:       req.Order,
		IsActive:    true,
	}

	if err := database.DB.Create(&grade).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create grade",
		})
		return
	}

	c.JSON(http.StatusCreated, convertToGradeResponse(&grade))
}

// ListGrades lists all grades
func ListGrades(c *gin.Context) {
	var grades []entity.Grade
	if err := database.DB.Where("is_active = ?", true).
		Order("order ASC, created_at ASC").
		Find(&grades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch grades",
		})
		return
	}

	// Convert to response
	items := make([]response.GradeResponse, len(grades))
	for i, grade := range grades {
		items[i] = convertToGradeResponse(&grade)
	}

	c.JSON(http.StatusOK, items)
}

// GetGrade retrieves a specific grade
func GetGrade(c *gin.Context) {
	gradeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid grade ID",
		})
		return
	}

	var grade entity.Grade
	if err := database.DB.First(&grade, gradeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Grade not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch grade",
			})
		}
		return
	}

	c.JSON(http.StatusOK, convertToGradeResponse(&grade))
}

// CreateSubject creates a new subject
func CreateSubject(c *gin.Context) {
	var req request.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Check if code already exists
	var existing entity.Subject
	if err := database.DB.Where("code = ?", req.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Subject code already exists",
		})
		return
	}

	// Create subject
	subject := entity.Subject{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		Order:       req.Order,
		IsActive:    true,
	}

	if err := database.DB.Create(&subject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create subject",
		})
		return
	}

	c.JSON(http.StatusCreated, convertToSubjectResponse(&subject))
}

// ListSubjects lists all subjects
func ListSubjects(c *gin.Context) {
	var subjects []entity.Subject
	if err := database.DB.Where("is_active = ?", true).
		Preload("Topics", "is_active = ?", true).
		Order("order ASC, created_at ASC").
		Find(&subjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch subjects",
		})
		return
	}

	// Convert to response
	items := make([]response.SubjectResponse, len(subjects))
	for i, subject := range subjects {
		items[i] = convertToSubjectResponse(&subject)
	}

	c.JSON(http.StatusOK, items)
}

// GetSubject retrieves a specific subject with its topics
func GetSubject(c *gin.Context) {
	subjectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid subject ID",
		})
		return
	}

	var subject entity.Subject
	if err := database.DB.Preload("Topics", "is_active = ?", true).
		First(&subject, subjectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Subject not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch subject",
			})
		}
		return
	}

	c.JSON(http.StatusOK, convertToSubjectResponse(&subject))
}

// GetSubjectByCode retrieves a subject by its code
func GetSubjectByCode(c *gin.Context) {
	code := c.Param("code")

	var subject entity.Subject
	if err := database.DB.Where("code = ? AND is_active = ?", code, true).
		Preload("Topics", "is_active = ?", true).
		First(&subject).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Subject not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch subject",
			})
		}
		return
	}

	c.JSON(http.StatusOK, convertToSubjectResponse(&subject))
}

// CreateTopic creates a new topic
func CreateTopic(c *gin.Context) {
	var req request.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Verify subject exists
	var subject entity.Subject
	if err := database.DB.First(&subject, req.SubjectID).Error; err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Subject not found",
		})
		return
	}

	// Check if code already exists within subject
	var existing entity.Topic
	if err := database.DB.Where("subject_id = ? AND code = ?", req.SubjectID, req.Code).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Topic code already exists in this subject",
		})
		return
	}

	// Create topic
	topic := entity.Topic{
		SubjectID:       req.SubjectID,
		Name:            req.Name,
		Code:            req.Code,
		Description:     req.Description,
		FullDescription: req.FullDescription,
		Icon:            req.Icon,
		Color:           req.Color,
		Order:           req.Order,
		IsActive:        true,
	}

	if err := database.DB.Create(&topic).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create topic",
		})
		return
	}

	// Load with subject info for response
	database.DB.Preload("Subject").First(&topic, topic.ID)

	c.JSON(http.StatusCreated, convertToTopicResponse(&topic))
}

// ListTopics lists topics for a subject
func ListTopics(c *gin.Context) {
	subjectID, err := strconv.ParseUint(c.Param("subject_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid subject ID",
		})
		return
	}

	var topics []entity.Topic
	query := database.DB.Where("subject_id = ? AND is_active = ?", subjectID, true).
		Preload("Subject").
		Order("order ASC, created_at ASC")

	if err := query.Find(&topics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch topics",
		})
		return
	}

	// Convert to response
	items := make([]response.TopicResponse, len(topics))
	for i, topic := range topics {
		items[i] = convertToTopicResponse(&topic)
		
		// Get question count for each topic
		var questionCount int64
		database.DB.Model(&entity.Question{}).
			Where("tags LIKE ?", "%"+topic.Code+"%").
			Count(&questionCount)
		items[i].QuestionCount = int(questionCount)
	}

	c.JSON(http.StatusOK, items)
}

// GetTopic retrieves a specific topic
func GetTopic(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid topic ID",
		})
		return
	}

	var topic entity.Topic
	if err := database.DB.Preload("Subject").
		First(&topic, topicID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Topic not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch topic",
			})
		}
		return
	}

	resp := convertToTopicResponse(&topic)
	
	// Get question count
	var questionCount int64
	database.DB.Model(&entity.Question{}).
		Where("tags LIKE ?", "%"+topic.Code+"%").
		Count(&questionCount)
	resp.QuestionCount = int(questionCount)

	c.JSON(http.StatusOK, resp)
}

// GetTopicByCode retrieves a topic by its code within a subject
func GetTopicByCode(c *gin.Context) {
	subjectCode := c.Param("subject_code")
	topicCode := c.Param("topic_code")

	// Find subject first
	var subject entity.Subject
	if err := database.DB.Where("code = ? AND is_active = ?", subjectCode, true).
		First(&subject).Error; err != nil {
		c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error: "Subject not found",
		})
		return
	}

	// Find topic within subject
	var topic entity.Topic
	if err := database.DB.Where("subject_id = ? AND code = ? AND is_active = ?", 
		subject.ID, topicCode, true).
		Preload("Subject").
		First(&topic).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Topic not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch topic",
			})
		}
		return
	}

	resp := convertToTopicResponse(&topic)
	
	// Get question count
	var questionCount int64
	database.DB.Model(&entity.Question{}).
		Where("tags LIKE ?", "%"+topic.Code+"%").
		Count(&questionCount)
	resp.QuestionCount = int(questionCount)

	c.JSON(http.StatusOK, resp)
}

// Helper functions to convert entities to responses
func convertToGradeResponse(grade *entity.Grade) response.GradeResponse {
	return response.GradeResponse{
		ID:          grade.ID,
		Name:        grade.Name,
		Code:        grade.Code,
		Description: grade.Description,
		AgeMin:      grade.AgeMin,
		AgeMax:      grade.AgeMax,
		Order:       grade.Order,
		IsActive:    grade.IsActive,
		CreatedAt:   grade.CreatedAt,
		UpdatedAt:   grade.UpdatedAt,
	}
}

func convertToSubjectResponse(subject *entity.Subject) response.SubjectResponse {
	resp := response.SubjectResponse{
		ID:          subject.ID,
		Name:        subject.Name,
		Code:        subject.Code,
		Description: subject.Description,
		Icon:        subject.Icon,
		Color:       subject.Color,
		Order:       subject.Order,
		IsActive:    subject.IsActive,
		CreatedAt:   subject.CreatedAt,
		UpdatedAt:   subject.UpdatedAt,
	}

	// Convert topics
	if len(subject.Topics) > 0 {
		resp.Topics = make([]response.TopicResponse, len(subject.Topics))
		for i, topic := range subject.Topics {
			resp.Topics[i] = convertToTopicResponse(&topic)
		}
	}

	return resp
}

func convertToTopicResponse(topic *entity.Topic) response.TopicResponse {
	resp := response.TopicResponse{
		ID:              topic.ID,
		SubjectID:       topic.SubjectID,
		Name:            topic.Name,
		Code:            topic.Code,
		Description:     topic.Description,
		FullDescription: topic.FullDescription,
		Icon:            topic.Icon,
		Color:           topic.Color,
		Order:           topic.Order,
		IsActive:        topic.IsActive,
		CreatedAt:       topic.CreatedAt,
		UpdatedAt:       topic.UpdatedAt,
	}

	if topic.Subject.Name != "" {
		resp.SubjectName = topic.Subject.Name
	}

	return resp
}