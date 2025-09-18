package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"testogo/internal/model/entity"
	"testogo/internal/model/request"
	"testogo/internal/model/response"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary 获取用户设置
// @Description 获取当前用户的所有设置
// @Tags 用户设置
// @Accept json
// @Produce json
// @Security BasicAuth
// @Success 200 {object} response.UserSettingsResponse "用户设置"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users/settings [get]
func GetUserSettings(c *gin.Context) {
	userID := c.GetUint("userID")

	// 查询用户设置
	var userSettings entity.UserSettings
	err := database.DB.Where("user_id = ?", userID).First(&userSettings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有设置记录，返回默认设置
			defaultSettings := entity.DefaultUserSettings()
			resp := response.UserSettingsResponse{
				UserID: userID,
				Learning: response.LearningSettingsResp{
					DefaultMode:     defaultSettings.Learning.DefaultMode,
					DailyTarget:     defaultSettings.Learning.DailyTarget,
					ReminderEnabled: defaultSettings.Learning.ReminderEnabled,
					ReminderTime:    defaultSettings.Learning.ReminderTime,
					StudyDays:       defaultSettings.Learning.StudyDays,
					AutoSave:        defaultSettings.Learning.AutoSave,
					ShowHints:       defaultSettings.Learning.ShowHints,
				},
				Interface: response.InterfaceSettingsResp{
					Theme:           defaultSettings.Interface.Theme,
					FontSize:        defaultSettings.Interface.FontSize,
					SidebarCollapse: defaultSettings.Interface.SidebarCollapse,
					Animations:      defaultSettings.Interface.Animations,
					Density:         defaultSettings.Interface.Density,
				},
				Notifications: response.NotificationSettingsResp{
					Desktop:          defaultSettings.Notifications.Desktop,
					StudyReminder:    defaultSettings.Notifications.StudyReminder,
					Achievements:     defaultSettings.Notifications.Achievements,
					PracticeComplete: defaultSettings.Notifications.PracticeComplete,
					Email:            defaultSettings.Notifications.Email,
				},
				Privacy: response.PrivacySettingsResp{
					Analytics:     defaultSettings.Privacy.Analytics,
					DataSync:      defaultSettings.Privacy.DataSync,
					DataRetention: defaultSettings.Privacy.DataRetention,
				},
				LastUpdated: defaultSettings.LastUpdated,
				Version:     defaultSettings.Version,
			}

			c.JSON(http.StatusOK, resp)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户设置失败"})
		return
	}

	// 解析JSON设置
	var settingsData entity.UserSettingsData
	if err := json.Unmarshal([]byte(userSettings.Settings), &settingsData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析用户设置失败"})
		return
	}

	// 构建响应
	resp := response.UserSettingsResponse{
		UserID: userID,
		Learning: response.LearningSettingsResp{
			DefaultMode:     settingsData.Learning.DefaultMode,
			DailyTarget:     settingsData.Learning.DailyTarget,
			ReminderEnabled: settingsData.Learning.ReminderEnabled,
			ReminderTime:    settingsData.Learning.ReminderTime,
			StudyDays:       settingsData.Learning.StudyDays,
			AutoSave:        settingsData.Learning.AutoSave,
			ShowHints:       settingsData.Learning.ShowHints,
		},
		Interface: response.InterfaceSettingsResp{
			Theme:           settingsData.Interface.Theme,
			FontSize:        settingsData.Interface.FontSize,
			SidebarCollapse: settingsData.Interface.SidebarCollapse,
			Animations:      settingsData.Interface.Animations,
			Density:         settingsData.Interface.Density,
		},
		Notifications: response.NotificationSettingsResp{
			Desktop:          settingsData.Notifications.Desktop,
			StudyReminder:    settingsData.Notifications.StudyReminder,
			Achievements:     settingsData.Notifications.Achievements,
			PracticeComplete: settingsData.Notifications.PracticeComplete,
			Email:            settingsData.Notifications.Email,
		},
		Privacy: response.PrivacySettingsResp{
			Analytics:     settingsData.Privacy.Analytics,
			DataSync:      settingsData.Privacy.DataSync,
			DataRetention: settingsData.Privacy.DataRetention,
		},
		LastUpdated: settingsData.LastUpdated,
		Version:     settingsData.Version,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary 更新用户设置
// @Description 更新当前用户的设置
// @Tags 用户设置
// @Accept json
// @Produce json
// @Security BasicAuth
// @Param request body request.UpdateUserSettingsRequest true "设置更新数据"
// @Success 200 {object} response.UserSettingsResponse "更新后的用户设置"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users/settings [put]
func UpdateUserSettings(c *gin.Context) {
	userID := c.GetUint("userID")

	var req request.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有设置或创建默认设置
	var userSettings entity.UserSettings
	var settingsData entity.UserSettingsData

	err := database.DB.Where("user_id = ?", userID).First(&userSettings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的设置记录
			settingsData = entity.DefaultUserSettings()
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户设置失败"})
			return
		}
	} else {
		// 解析现有设置
		if err := json.Unmarshal([]byte(userSettings.Settings), &settingsData); err != nil {
			// 如果解析失败，使用默认设置
			settingsData = entity.DefaultUserSettings()
		}
	}

	// 更新学习设置
	if req.Learning != nil {
		if req.Learning.DefaultMode != nil {
			settingsData.Learning.DefaultMode = *req.Learning.DefaultMode
		}
		if req.Learning.DailyTarget != nil {
			settingsData.Learning.DailyTarget = *req.Learning.DailyTarget
		}
		if req.Learning.ReminderEnabled != nil {
			settingsData.Learning.ReminderEnabled = *req.Learning.ReminderEnabled
		}
		if req.Learning.ReminderTime != nil {
			settingsData.Learning.ReminderTime = *req.Learning.ReminderTime
		}
		if req.Learning.StudyDays != nil {
			settingsData.Learning.StudyDays = *req.Learning.StudyDays
		}
		if req.Learning.AutoSave != nil {
			settingsData.Learning.AutoSave = *req.Learning.AutoSave
		}
		if req.Learning.ShowHints != nil {
			settingsData.Learning.ShowHints = *req.Learning.ShowHints
		}
	}

	// 更新界面设置
	if req.Interface != nil {
		if req.Interface.Theme != nil {
			settingsData.Interface.Theme = *req.Interface.Theme
		}
		if req.Interface.FontSize != nil {
			settingsData.Interface.FontSize = *req.Interface.FontSize
		}
		if req.Interface.SidebarCollapse != nil {
			settingsData.Interface.SidebarCollapse = *req.Interface.SidebarCollapse
		}
		if req.Interface.Animations != nil {
			settingsData.Interface.Animations = *req.Interface.Animations
		}
		if req.Interface.Density != nil {
			settingsData.Interface.Density = *req.Interface.Density
		}
	}

	// 更新通知设置
	if req.Notifications != nil {
		if req.Notifications.Desktop != nil {
			settingsData.Notifications.Desktop = *req.Notifications.Desktop
		}
		if req.Notifications.StudyReminder != nil {
			settingsData.Notifications.StudyReminder = *req.Notifications.StudyReminder
		}
		if req.Notifications.Achievements != nil {
			settingsData.Notifications.Achievements = *req.Notifications.Achievements
		}
		if req.Notifications.PracticeComplete != nil {
			settingsData.Notifications.PracticeComplete = *req.Notifications.PracticeComplete
		}
		if req.Notifications.Email != nil {
			settingsData.Notifications.Email = *req.Notifications.Email
		}
	}

	// 更新隐私设置
	if req.Privacy != nil {
		if req.Privacy.Analytics != nil {
			settingsData.Privacy.Analytics = *req.Privacy.Analytics
		}
		if req.Privacy.DataSync != nil {
			settingsData.Privacy.DataSync = *req.Privacy.DataSync
		}
		if req.Privacy.DataRetention != nil {
			settingsData.Privacy.DataRetention = *req.Privacy.DataRetention
		}
	}

	// 更新时间戳
	settingsData.LastUpdated = time.Now()

	// 序列化设置数据
	settingsJSON, err := json.Marshal(settingsData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化设置数据失败"})
		return
	}

	// 保存到数据库
	if userSettings.ID == 0 {
		// 创建新记录
		userSettings = entity.UserSettings{
			UserID:   userID,
			Settings: string(settingsJSON),
		}
		if err := database.DB.Create(&userSettings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户设置失败"})
			return
		}
	} else {
		// 更新现有记录
		if err := database.DB.Model(&userSettings).Update("settings", string(settingsJSON)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户设置失败"})
			return
		}
	}

	// 构建响应
	resp := response.UserSettingsResponse{
		UserID: userID,
		Learning: response.LearningSettingsResp{
			DefaultMode:     settingsData.Learning.DefaultMode,
			DailyTarget:     settingsData.Learning.DailyTarget,
			ReminderEnabled: settingsData.Learning.ReminderEnabled,
			ReminderTime:    settingsData.Learning.ReminderTime,
			StudyDays:       settingsData.Learning.StudyDays,
			AutoSave:        settingsData.Learning.AutoSave,
			ShowHints:       settingsData.Learning.ShowHints,
		},
		Interface: response.InterfaceSettingsResp{
			Theme:           settingsData.Interface.Theme,
			FontSize:        settingsData.Interface.FontSize,
			SidebarCollapse: settingsData.Interface.SidebarCollapse,
			Animations:      settingsData.Interface.Animations,
			Density:         settingsData.Interface.Density,
		},
		Notifications: response.NotificationSettingsResp{
			Desktop:          settingsData.Notifications.Desktop,
			StudyReminder:    settingsData.Notifications.StudyReminder,
			Achievements:     settingsData.Notifications.Achievements,
			PracticeComplete: settingsData.Notifications.PracticeComplete,
			Email:            settingsData.Notifications.Email,
		},
		Privacy: response.PrivacySettingsResp{
			Analytics:     settingsData.Privacy.Analytics,
			DataSync:      settingsData.Privacy.DataSync,
			DataRetention: settingsData.Privacy.DataRetention,
		},
		LastUpdated: settingsData.LastUpdated,
		Version:     settingsData.Version,
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary 重置用户设置
// @Description 将用户设置重置为默认值
// @Tags 用户设置
// @Accept json
// @Produce json
// @Security BasicAuth
// @Success 200 {object} response.UserSettingsResponse "重置后的用户设置"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/users/settings/reset [post]
func ResetUserSettings(c *gin.Context) {
	userID := c.GetUint("userID")

	// 获取默认设置
	settingsData := entity.DefaultUserSettings()

	// 序列化设置数据
	settingsJSON, err := json.Marshal(settingsData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "序列化设置数据失败"})
		return
	}

	// 更新或创建设置记录
	var userSettings entity.UserSettings
	err = database.DB.Where("user_id = ?", userID).First(&userSettings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			userSettings = entity.UserSettings{
				UserID:   userID,
				Settings: string(settingsJSON),
			}
			if err := database.DB.Create(&userSettings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户设置失败"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户设置失败"})
			return
		}
	} else {
		// 更新现有记录
		if err := database.DB.Model(&userSettings).Update("settings", string(settingsJSON)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "重置用户设置失败"})
			return
		}
	}

	// 构建响应
	resp := response.UserSettingsResponse{
		UserID: userID,
		Learning: response.LearningSettingsResp{
			DefaultMode:     settingsData.Learning.DefaultMode,
			DailyTarget:     settingsData.Learning.DailyTarget,
			ReminderEnabled: settingsData.Learning.ReminderEnabled,
			ReminderTime:    settingsData.Learning.ReminderTime,
			StudyDays:       settingsData.Learning.StudyDays,
			AutoSave:        settingsData.Learning.AutoSave,
			ShowHints:       settingsData.Learning.ShowHints,
		},
		Interface: response.InterfaceSettingsResp{
			Theme:           settingsData.Interface.Theme,
			FontSize:        settingsData.Interface.FontSize,
			SidebarCollapse: settingsData.Interface.SidebarCollapse,
			Animations:      settingsData.Interface.Animations,
			Density:         settingsData.Interface.Density,
		},
		Notifications: response.NotificationSettingsResp{
			Desktop:          settingsData.Notifications.Desktop,
			StudyReminder:    settingsData.Notifications.StudyReminder,
			Achievements:     settingsData.Notifications.Achievements,
			PracticeComplete: settingsData.Notifications.PracticeComplete,
			Email:            settingsData.Notifications.Email,
		},
		Privacy: response.PrivacySettingsResp{
			Analytics:     settingsData.Privacy.Analytics,
			DataSync:      settingsData.Privacy.DataSync,
			DataRetention: settingsData.Privacy.DataRetention,
		},
		LastUpdated: settingsData.LastUpdated,
		Version:     settingsData.Version,
	}

	c.JSON(http.StatusOK, resp)
}