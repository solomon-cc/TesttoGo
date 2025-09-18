package request

import "time"

// UpdateUserSettingsRequest represents the request to update user settings
type UpdateUserSettingsRequest struct {
	Learning      *LearningSettingsReq      `json:"learning,omitempty"`
	Interface     *InterfaceSettingsReq     `json:"interface,omitempty"`
	Notifications *NotificationSettingsReq  `json:"notifications,omitempty"`
	Privacy       *PrivacySettingsReq       `json:"privacy,omitempty"`
}

// LearningSettingsReq represents learning preferences in request
type LearningSettingsReq struct {
	DefaultMode     *string    `json:"default_mode,omitempty" binding:"omitempty,oneof=learning test"`
	DailyTarget     *int       `json:"daily_target,omitempty" binding:"omitempty,min=1,max=200"`
	ReminderEnabled *bool      `json:"reminder_enabled,omitempty"`
	ReminderTime    *time.Time `json:"reminder_time,omitempty"`
	StudyDays       *[]int     `json:"study_days,omitempty" binding:"omitempty,dive,min=0,max=6"`
	AutoSave        *bool      `json:"auto_save,omitempty"`
	ShowHints       *bool      `json:"show_hints,omitempty"`
}

// InterfaceSettingsReq represents UI preferences in request
type InterfaceSettingsReq struct {
	Theme           *string `json:"theme,omitempty" binding:"omitempty,oneof=light dark auto"`
	FontSize        *string `json:"font_size,omitempty" binding:"omitempty,oneof=small medium large"`
	SidebarCollapse *bool   `json:"sidebar_collapse,omitempty"`
	Animations      *bool   `json:"animations,omitempty"`
	Density         *string `json:"density,omitempty" binding:"omitempty,oneof=comfortable standard compact"`
}

// NotificationSettingsReq represents notification preferences in request
type NotificationSettingsReq struct {
	Desktop          *bool `json:"desktop,omitempty"`
	StudyReminder    *bool `json:"study_reminder,omitempty"`
	Achievements     *bool `json:"achievements,omitempty"`
	PracticeComplete *bool `json:"practice_complete,omitempty"`
	Email            *bool `json:"email,omitempty"`
}

// PrivacySettingsReq represents privacy preferences in request
type PrivacySettingsReq struct {
	Analytics     *bool   `json:"analytics,omitempty"`
	DataSync      *bool   `json:"data_sync,omitempty"`
	DataRetention *string `json:"data_retention,omitempty" binding:"omitempty,oneof=1month 3months 6months 1year forever"`
}