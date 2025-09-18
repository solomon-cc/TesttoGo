package response

import "time"

// UserSettingsResponse represents user settings in API responses
type UserSettingsResponse struct {
	UserID      uint                    `json:"user_id"`
	Learning    LearningSettingsResp    `json:"learning"`
	Interface   InterfaceSettingsResp   `json:"interface"`
	Notifications NotificationSettingsResp `json:"notifications"`
	Privacy     PrivacySettingsResp     `json:"privacy"`
	LastUpdated time.Time               `json:"last_updated"`
	Version     string                  `json:"version"`
}

// LearningSettingsResp represents learning preferences in response
type LearningSettingsResp struct {
	DefaultMode     string    `json:"default_mode"`
	DailyTarget     int       `json:"daily_target"`
	ReminderEnabled bool      `json:"reminder_enabled"`
	ReminderTime    time.Time `json:"reminder_time"`
	StudyDays       []int     `json:"study_days"`
	AutoSave        bool      `json:"auto_save"`
	ShowHints       bool      `json:"show_hints"`
}

// InterfaceSettingsResp represents UI preferences in response
type InterfaceSettingsResp struct {
	Theme           string `json:"theme"`
	FontSize        string `json:"font_size"`
	SidebarCollapse bool   `json:"sidebar_collapse"`
	Animations      bool   `json:"animations"`
	Density         string `json:"density"`
}

// NotificationSettingsResp represents notification preferences in response
type NotificationSettingsResp struct {
	Desktop          bool `json:"desktop"`
	StudyReminder    bool `json:"study_reminder"`
	Achievements     bool `json:"achievements"`
	PracticeComplete bool `json:"practice_complete"`
	Email            bool `json:"email"`
}

// PrivacySettingsResp represents privacy preferences in response
type PrivacySettingsResp struct {
	Analytics     bool   `json:"analytics"`
	DataSync      bool   `json:"data_sync"`
	DataRetention string `json:"data_retention"`
}