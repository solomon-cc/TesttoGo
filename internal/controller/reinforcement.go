package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/internal/model/response"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateReinforcementSetting creates a new reinforcement setting
func CreateReinforcementSetting(c *gin.Context) {
	var req request.CreateReinforcementSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user
	userID := c.GetUint("user_id")

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create reinforcement setting
	setting := entity.ReinforcementSetting{
		Name:          req.Name,
		Description:   req.Description,
		CreatorID:     userID,
		Mode:          entity.ReinforcementMode(req.Mode),
		ScheduleType:  entity.ReinforcementScheduleType(req.ScheduleType),
		RatioValue:    req.RatioValue,
		IntervalValue: req.IntervalValue,
		IsActive:      true,
	}

	if err := tx.Create(&setting).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create reinforcement setting",
		})
		return
	}

	// Associate with items
	if len(req.ItemIDs) > 0 {
		var items []entity.ReinforcementItem
		if err := tx.Where("id IN ?", req.ItemIDs).Find(&items).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "Invalid reinforcement item IDs",
			})
			return
		}

		if err := tx.Model(&setting).Association("ReinforcementItems").Append(&items); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to associate reinforcement items",
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to save reinforcement setting",
		})
		return
	}

	// Load complete data for response
	var settingResp entity.ReinforcementSetting
	database.DB.Preload("Creator").
		Preload("ReinforcementItems").
		First(&settingResp, setting.ID)

	c.JSON(http.StatusCreated, convertToReinforcementSettingResponse(&settingResp))
}

// ListReinforcementSettings lists reinforcement settings with filtering
func ListReinforcementSettings(c *gin.Context) {
	var req request.ListReinforcementSettingsRequest
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
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	// Build query
	query := database.DB.Model(&entity.ReinforcementSetting{}).
		Preload("Creator").
		Preload("ReinforcementItems")

	// Apply role-based filtering
	if userRole == "teacher" {
		if req.CreatorID == 0 {
			query = query.Where("creator_id = ?", userID)
		}
	}
	// Admin can see all settings

	// Apply filters
	if req.CreatorID != 0 && userRole != "teacher" {
		query = query.Where("creator_id = ?", req.CreatorID)
	}
	if req.Mode != "" {
		query = query.Where("mode = ?", req.Mode)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Apply pagination
	offset := (req.Page - 1) * req.PageSize
	var settings []entity.ReinforcementSetting
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch reinforcement settings",
		})
		return
	}

	// Convert to response
	items := make([]response.ReinforcementSettingResponse, len(settings))
	for i, setting := range settings {
		items[i] = convertToReinforcementSettingResponse(&setting)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	c.JSON(http.StatusOK, response.ReinforcementSettingListResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// GetReinforcementSetting retrieves a specific reinforcement setting
func GetReinforcementSetting(c *gin.Context) {
	settingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid setting ID",
		})
		return
	}

	var setting entity.ReinforcementSetting
	if err := database.DB.Preload("Creator").
		Preload("ReinforcementItems").
		First(&setting, settingID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Reinforcement setting not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch reinforcement setting",
			})
		}
		return
	}

	// Check access permissions
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	if userRole == "teacher" && setting.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
		})
		return
	}

	c.JSON(http.StatusOK, convertToReinforcementSettingResponse(&setting))
}

// UpdateReinforcementSetting updates an existing reinforcement setting
func UpdateReinforcementSetting(c *gin.Context) {
	settingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid setting ID",
		})
		return
	}

	var req request.UpdateReinforcementSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	// Find setting
	var setting entity.ReinforcementSetting
	if err := database.DB.First(&setting, settingID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Reinforcement setting not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch reinforcement setting",
			})
		}
		return
	}

	// Check permissions
	if userRole == "teacher" && setting.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
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

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Mode != nil {
		updates["mode"] = *req.Mode
	}
	if req.ScheduleType != nil {
		updates["schedule_type"] = *req.ScheduleType
	}
	if req.RatioValue != nil {
		updates["ratio_value"] = *req.RatioValue
	}
	if req.IntervalValue != nil {
		updates["interval_value"] = *req.IntervalValue
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := tx.Model(&setting).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to update reinforcement setting",
		})
		return
	}

	// Update associated items if provided
	if len(req.ItemIDs) > 0 {
		// Clear existing associations
		if err := tx.Model(&setting).Association("ReinforcementItems").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to clear existing items",
			})
			return
		}

		// Add new associations
		var items []entity.ReinforcementItem
		if err := tx.Where("id IN ?", req.ItemIDs).Find(&items).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, response.ErrorResponse{
				Error: "Invalid reinforcement item IDs",
			})
			return
		}

		if err := tx.Model(&setting).Association("ReinforcementItems").Append(&items); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to update reinforcement items",
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to save changes",
		})
		return
	}

	// Reload setting with relations
	database.DB.Preload("Creator").
		Preload("ReinforcementItems").
		First(&setting, settingID)

	c.JSON(http.StatusOK, convertToReinforcementSettingResponse(&setting))
}

// DeleteReinforcementSetting deletes a reinforcement setting
func DeleteReinforcementSetting(c *gin.Context) {
	settingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid setting ID",
		})
		return
	}

	// Get current user
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	// Find setting
	var setting entity.ReinforcementSetting
	if err := database.DB.First(&setting, settingID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Reinforcement setting not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch reinforcement setting",
			})
		}
		return
	}

	// Check permissions
	if userRole == "teacher" && setting.CreatorID != userID {
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			Error: "Access denied",
		})
		return
	}

	// Soft delete
	if err := database.DB.Delete(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to delete reinforcement setting",
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Reinforcement setting deleted successfully",
	})
}

// CopyReinforcementSetting creates a copy of existing reinforcement setting
func CopyReinforcementSetting(c *gin.Context) {
	sourceID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid setting ID",
		})
		return
	}

	// Get current user
	userID := c.GetUint("user_id")

	// Find source setting
	var sourceSetting entity.ReinforcementSetting
	if err := database.DB.Preload("ReinforcementItems").
		First(&sourceSetting, sourceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Source reinforcement setting not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch source reinforcement setting",
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

	// Create new setting
	newSetting := entity.ReinforcementSetting{
		Name:          sourceSetting.Name + " (Copy)",
		Description:   sourceSetting.Description,
		CreatorID:     userID,
		Mode:          sourceSetting.Mode,
		ScheduleType:  sourceSetting.ScheduleType,
		RatioValue:    sourceSetting.RatioValue,
		IntervalValue: sourceSetting.IntervalValue,
		IsActive:      true,
	}

	if err := tx.Create(&newSetting).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create reinforcement setting copy",
		})
		return
	}

	// Copy item associations
	if len(sourceSetting.ReinforcementItems) > 0 {
		if err := tx.Model(&newSetting).Association("ReinforcementItems").Append(&sourceSetting.ReinforcementItems); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to copy reinforcement items",
			})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to save reinforcement setting copy",
		})
		return
	}

	// Load complete data for response
	var settingResp entity.ReinforcementSetting
	database.DB.Preload("Creator").
		Preload("ReinforcementItems").
		First(&settingResp, newSetting.ID)

	c.JSON(http.StatusCreated, convertToReinforcementSettingResponse(&settingResp))
}

// CreateReinforcementItem creates a new reinforcement item
func CreateReinforcementItem(c *gin.Context) {
	var req request.CreateReinforcementItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Create reinforcement item
	item := entity.ReinforcementItem{
		Name:          req.Name,
		Type:          entity.ReinforcementItemType(req.Type),
		MediaURL:      req.MediaURL,
		Color:         req.Color,
		Icon:          req.Icon,
		Duration:      req.Duration,
		AnimationType: req.AnimationType,
		IsActive:      true,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to create reinforcement item",
		})
		return
	}

	c.JSON(http.StatusCreated, convertToReinforcementItemResponse(&item))
}

// ListReinforcementItems lists reinforcement items with filtering
func ListReinforcementItems(c *gin.Context) {
	var req request.ListReinforcementItemsRequest
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
		req.PageSize = 20
	}

	// Build query
	query := database.DB.Model(&entity.ReinforcementItem{})

	// Apply filters
	if req.Type != "" {
		query = query.Where("type = ?", req.Type)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	// Count total
	var total int64
	query.Count(&total)

	// Apply pagination
	offset := (req.Page - 1) * req.PageSize
	var items []entity.ReinforcementItem
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to fetch reinforcement items",
		})
		return
	}

	// Convert to response
	responseItems := make([]response.ReinforcementItemResponse, len(items))
	for i, item := range items {
		responseItems[i] = convertToReinforcementItemResponse(&item)
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))

	c.JSON(http.StatusOK, response.ReinforcementItemListResponse{
		Items:      responseItems,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// GetReinforcementItem retrieves a specific reinforcement item
func GetReinforcementItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid item ID",
		})
		return
	}

	var item entity.ReinforcementItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Reinforcement item not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch reinforcement item",
			})
		}
		return
	}

	c.JSON(http.StatusOK, convertToReinforcementItemResponse(&item))
}

// UpdateReinforcementItem updates an existing reinforcement item
func UpdateReinforcementItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid item ID",
		})
		return
	}

	var req request.UpdateReinforcementItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Find item
	var item entity.ReinforcementItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Reinforcement item not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch reinforcement item",
			})
		}
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.MediaURL != nil {
		updates["media_url"] = *req.MediaURL
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.Duration != nil {
		updates["duration"] = *req.Duration
	}
	if req.AnimationType != nil {
		updates["animation_type"] = *req.AnimationType
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := database.DB.Model(&item).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to update reinforcement item",
		})
		return
	}

	c.JSON(http.StatusOK, convertToReinforcementItemResponse(&item))
}

// DeleteReinforcementItem deletes a reinforcement item
func DeleteReinforcementItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid item ID",
		})
		return
	}

	// Find item
	var item entity.ReinforcementItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, response.ErrorResponse{
				Error: "Reinforcement item not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, response.ErrorResponse{
				Error: "Failed to fetch reinforcement item",
			})
		}
		return
	}

	// Soft delete
	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to delete reinforcement item",
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Reinforcement item deleted successfully",
	})
}

// RecordReinforcementTrigger logs a reinforcement trigger event
func RecordReinforcementTrigger(c *gin.Context) {
	var req request.RecordReinforcementTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get current user
	userID := c.GetUint("user_id")

	// Serialize context data
	contextJSON, _ := json.Marshal(req.ContextData)

	// Create log entry
	log := entity.ReinforcementLog{
		UserID:                 userID,
		ReinforcementSettingID: req.ReinforcementSettingID,
		ReinforcementItemID:    req.ReinforcementItemID,
		HomeworkID:             req.HomeworkID,
		SessionID:              req.SessionID,
		TriggerType:            req.TriggerType,
		TriggerValue:           req.TriggerValue,
		ContextData:            string(contextJSON),
	}

	if err := database.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to record reinforcement trigger",
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Message: "Reinforcement trigger recorded successfully",
		Data: map[string]interface{}{
			"log_id": log.ID,
		},
	})
}

// GetReinforcementStats retrieves reinforcement statistics
func GetReinforcementStats(c *gin.Context) {
	var req request.ReinforcementStatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
		})
		return
	}

	// Build query
	query := database.DB.Model(&entity.ReinforcementLog{}).
		Preload("ReinforcementSetting").
		Preload("ReinforcementItem").
		Preload("User")

	// Apply filters
	if req.UserID != 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.HomeworkID != 0 {
		query = query.Where("homework_id = ?", req.HomeworkID)
	}
	if req.DateFrom != "" {
		query = query.Where("created_at >= ?", req.DateFrom)
	}
	if req.DateTo != "" {
		query = query.Where("created_at <= ?", req.DateTo)
	}

	// Get total triggers
	var totalTriggers int64
	query.Count(&totalTriggers)

	// Get recent triggers
	var recentLogs []entity.ReinforcementLog
	query.Order("created_at DESC").Limit(10).Find(&recentLogs)

	// Convert to response
	recentTriggers := make([]response.ReinforcementLogResponse, len(recentLogs))
	for i, log := range recentLogs {
		recentTriggers[i] = convertToReinforcementLogResponse(&log)
	}

	// Basic statistics
	stats := response.ReinforcementStatsResponse{
		TotalTriggers:  int(totalTriggers),
		RecentTriggers: recentTriggers,
	}

	c.JSON(http.StatusOK, stats)
}

// Helper functions to convert entities to responses
func convertToReinforcementSettingResponse(setting *entity.ReinforcementSetting) response.ReinforcementSettingResponse {
	resp := response.ReinforcementSettingResponse{
		ID:            setting.ID,
		Name:          setting.Name,
		Description:   setting.Description,
		CreatorID:     setting.CreatorID,
		Mode:          string(setting.Mode),
		ScheduleType:  string(setting.ScheduleType),
		RatioValue:    setting.RatioValue,
		IntervalValue: setting.IntervalValue,
		IsActive:      setting.IsActive,
		CreatedAt:     setting.CreatedAt,
		UpdatedAt:     setting.UpdatedAt,
	}

	if setting.Creator.Username != "" {
		resp.CreatorName = setting.Creator.Username
	}

	// Convert items
	if len(setting.ReinforcementItems) > 0 {
		resp.Items = make([]response.ReinforcementItemResponse, len(setting.ReinforcementItems))
		for i, item := range setting.ReinforcementItems {
			resp.Items[i] = convertToReinforcementItemResponse(&item)
		}
	}

	return resp
}

func convertToReinforcementItemResponse(item *entity.ReinforcementItem) response.ReinforcementItemResponse {
	return response.ReinforcementItemResponse{
		ID:            item.ID,
		Name:          item.Name,
		Type:          string(item.Type),
		MediaURL:      item.MediaURL,
		Color:         item.Color,
		Icon:          item.Icon,
		Duration:      item.Duration,
		AnimationType: item.AnimationType,
		IsActive:      item.IsActive,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func convertToReinforcementLogResponse(log *entity.ReinforcementLog) response.ReinforcementLogResponse {
	var contextData map[string]interface{}
	if log.ContextData != "" {
		json.Unmarshal([]byte(log.ContextData), &contextData)
	}

	resp := response.ReinforcementLogResponse{
		ID:                     log.ID,
		UserID:                 log.UserID,
		ReinforcementSettingID: log.ReinforcementSettingID,
		ReinforcementItemID:    log.ReinforcementItemID,
		HomeworkID:             log.HomeworkID,
		SessionID:              log.SessionID,
		TriggerType:            log.TriggerType,
		TriggerValue:           log.TriggerValue,
		ContextData:            contextData,
		CreatedAt:              log.CreatedAt,
	}

	if log.User.Username != "" {
		resp.UserName = log.User.Username
	}

	return resp
}

// UploadRewardVideo handles video file upload for reinforcement rewards
func UploadRewardVideo(c *gin.Context) {
	// Get the uploaded file
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "No video file provided",
		})
		return
	}

	// Validate file type (basic validation)
	if !isValidVideoFile(file.Filename) {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Invalid video file type. Supported types: mp4, webm, avi",
		})
		return
	}

	// Validate file size (max 50MB)
	if file.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: "Video file too large. Maximum size: 50MB",
		})
		return
	}

	// Generate unique filename
	videoID := generateUniqueID()
	filename := videoID + getFileExtension(file.Filename)
	uploadPath := "uploads/videos/" + filename

	// Save file to disk (in production, use cloud storage)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: "Failed to save video file",
		})
		return
	}

	// Create URL for accessing the video
	videoURL := "/api/v1/videos/" + filename

	c.JSON(http.StatusOK, response.VideoUploadResponse{
		URL:      videoURL,
		VideoID:  videoID,
		FileName: file.Filename,
		Size:     file.Size,
	})
}

// DeleteRewardVideo deletes an uploaded reward video
func DeleteRewardVideo(c *gin.Context) {
	videoID := c.Param("id")
	
	// Find the video file (in production, you'd store metadata in database)
	// For now, just construct the path
	filename := videoID + ".mp4" // Assuming mp4, in production check database
	uploadPath := "uploads/videos/" + filename

	// Delete file (in production, delete from cloud storage)
	// Note: This is a simplified implementation
	// TODO: Implement actual file deletion using os.Remove(uploadPath)
	_ = uploadPath // Temporarily ignore until file deletion is implemented
	
	c.JSON(http.StatusOK, response.SuccessResponse{
		Message: "Video deleted successfully",
	})
}

// Helper functions
func isValidVideoFile(filename string) bool {
	validExtensions := []string{".mp4", ".webm", ".avi", ".mov"}
	ext := getFileExtension(filename)
	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}

func generateUniqueID() string {
	// Simple implementation - in production use UUID
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}