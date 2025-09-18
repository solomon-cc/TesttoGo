package entity

import (
	"time"

	"gorm.io/gorm"
)

// UserSettings represents user's application settings
type UserSettings struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"uniqueIndex" json:"user_id"`
	Settings  string         `gorm:"type:text" json:"settings"` // JSON format settings
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// LearningSettings represents learning preferences
type LearningSettings struct {
	DefaultMode     string    `json:"default_mode"`      // learning, test
	DailyTarget     int       `json:"daily_target"`      // daily question target
	ReminderEnabled bool      `json:"reminder_enabled"`  // reminder notifications
	ReminderTime    time.Time `json:"reminder_time"`     // reminder time
	StudyDays       []int     `json:"study_days"`        // days of week for reminders
	AutoSave        bool      `json:"auto_save"`         // auto save progress
	ShowHints       bool      `json:"show_hints"`        // show hints in learning mode
}

// InterfaceSettings represents UI preferences
type InterfaceSettings struct {
	Theme           string `json:"theme"`            // light, dark, auto
	FontSize        string `json:"font_size"`        // small, medium, large
	SidebarCollapse bool   `json:"sidebar_collapse"` // sidebar collapsed state
	Animations      bool   `json:"animations"`       // enable animations
	Density         string `json:"density"`          // comfortable, standard, compact
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	Desktop           bool `json:"desktop"`            // desktop notifications
	StudyReminder     bool `json:"study_reminder"`     // study reminders
	Achievements      bool `json:"achievements"`       // achievement notifications
	PracticeComplete  bool `json:"practice_complete"`  // practice completion notifications
	Email             bool `json:"email"`              // email notifications
}

// PrivacySettings represents privacy preferences
type PrivacySettings struct {
	Analytics     bool   `json:"analytics"`      // allow analytics
	DataSync      bool   `json:"data_sync"`      // enable data sync
	DataRetention string `json:"data_retention"` // data retention period
}

// UserSettingsData represents the complete user settings structure
type UserSettingsData struct {
	Learning      LearningSettings      `json:"learning"`
	Interface     InterfaceSettings     `json:"interface"`
	Notifications NotificationSettings  `json:"notifications"`
	Privacy       PrivacySettings       `json:"privacy"`
	LastUpdated   time.Time             `json:"last_updated"`
	Version       string                `json:"version"`
}

// DefaultUserSettings returns default settings for new users
func DefaultUserSettings() UserSettingsData {
	return UserSettingsData{
		Learning: LearningSettings{
			DefaultMode:     "learning",
			DailyTarget:     20,
			ReminderEnabled: true,
			ReminderTime:    time.Date(2024, 1, 1, 19, 0, 0, 0, time.UTC),
			StudyDays:       []int{1, 2, 3, 4, 5}, // Monday to Friday
			AutoSave:        true,
			ShowHints:       true,
		},
		Interface: InterfaceSettings{
			Theme:           "light",
			FontSize:        "medium",
			SidebarCollapse: false,
			Animations:      true,
			Density:         "standard",
		},
		Notifications: NotificationSettings{
			Desktop:          false,
			StudyReminder:    true,
			Achievements:     true,
			PracticeComplete: true,
			Email:            false,
		},
		Privacy: PrivacySettings{
			Analytics:     true,
			DataSync:      true,
			DataRetention: "1year",
		},
		LastUpdated: time.Now(),
		Version:     "1.0.0",
	}
}