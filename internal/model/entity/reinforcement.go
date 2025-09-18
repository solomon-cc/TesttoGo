package entity

import (
	"time"

	"gorm.io/gorm"
)

type ReinforcementMode string
type ReinforcementScheduleType string
type ReinforcementItemType string

const (
	// Learning modes
	ModeStudy ReinforcementMode = "study"
	ModeTest  ReinforcementMode = "test"

	// Schedule types
	ScheduleVR ReinforcementScheduleType = "VR" // Variable Ratio
	ScheduleFR ReinforcementScheduleType = "FR" // Fixed Ratio
	ScheduleVI ReinforcementScheduleType = "VI" // Variable Interval
	ScheduleFI ReinforcementScheduleType = "FI" // Fixed Interval

	// Reinforcement item types
	ItemTypeVideo      ReinforcementItemType = "video"
	ItemTypeGame       ReinforcementItemType = "game"
	ItemTypeVirtual    ReinforcementItemType = "virtual"
	ItemTypeAnimation  ReinforcementItemType = "animation"
	ItemTypeSound      ReinforcementItemType = "sound"

	// Legacy types (for backward compatibility)
	ItemTypeFlower     ReinforcementItemType = "flower"
	ItemTypeStar       ReinforcementItemType = "star"
	ItemTypeBadge      ReinforcementItemType = "badge"
	ItemTypeFireworks  ReinforcementItemType = "fireworks"
)

// ReinforcementSetting represents a reinforcement configuration
type ReinforcementSetting struct {
	ID                uint                      `gorm:"primarykey" json:"id"`
	Name              string                    `gorm:"type:varchar(100)" json:"name"`
	Description       string                    `gorm:"type:text" json:"description"`
	CreatorID         uint                      `json:"creator_id"`
	Mode              ReinforcementMode         `gorm:"type:enum('study','test');default:'study'" json:"mode"`
	ScheduleType      ReinforcementScheduleType `gorm:"type:enum('VR','FR','VI','FI')" json:"schedule_type"`
	RatioValue        int                       `gorm:"comment:'For VR/FR: number of correct answers needed'" json:"ratio_value"`
	IntervalValue     int                       `gorm:"comment:'For VI/FI: time interval in minutes'" json:"interval_value"`
	IsActive          bool                      `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	DeletedAt         gorm.DeletedAt            `gorm:"index" json:"-"`

	// Relations
	Creator                     User                        `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	ReinforcementItems          []ReinforcementItem         `gorm:"many2many:reinforcement_setting_items;" json:"items,omitempty"`
	HomeworkAssignments         []HomeworkAssignment        `gorm:"foreignKey:ReinforcementSettingID" json:"homework_assignments,omitempty"`
	ReinforcementLogs           []ReinforcementLog          `gorm:"foreignKey:ReinforcementSettingID" json:"logs,omitempty"`
}

// ReinforcementItem represents a reward item
type ReinforcementItem struct {
	ID            uint                  `gorm:"primarykey" json:"id"`
	Name          string                `gorm:"type:varchar(100)" json:"name"`
	Type          ReinforcementItemType `gorm:"type:varchar(20)" json:"type"` // Changed to varchar to support all types
	Description   string                `gorm:"type:text" json:"description"`
	ContentURL    string                `gorm:"type:varchar(500)" json:"content_url"` // For videos, audio files
	PreviewURL    string                `gorm:"type:varchar(500)" json:"preview_url"` // For thumbnails, icons
	MediaURL      string                `gorm:"type:varchar(500)" json:"media_url"`   // Legacy field, kept for compatibility
	Color         string                `gorm:"type:varchar(7)" json:"color"`         // Hex color code
	Icon          string                `gorm:"type:varchar(100)" json:"icon"`        // Icon class or name
	Duration      int                   `gorm:"default:3000" json:"duration"`        // Display duration in ms or seconds
	AnimationType string                `gorm:"type:varchar(50)" json:"animation_type"` // Animation style
	GameType      string                `gorm:"type:varchar(50)" json:"game_type"`      // For games: puzzle, memory, etc.
	Difficulty    string                `gorm:"type:varchar(20)" json:"difficulty"`     // For games: easy, medium, hard
	RewardPoints  int                   `gorm:"default:0" json:"reward_points"`         // For virtual rewards
	Volume        int                   `gorm:"default:80" json:"volume"`               // For sound effects (0-100)
	Tags          string                `gorm:"type:text" json:"tags"`                  // JSON array of tags
	IsActive      bool                  `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	DeletedAt     gorm.DeletedAt        `gorm:"index" json:"-"`

	// Relations
	ReinforcementSettings []ReinforcementSetting `gorm:"many2many:reinforcement_setting_items;" json:"settings,omitempty"`
	ReinforcementLogs     []ReinforcementLog     `gorm:"foreignKey:ReinforcementItemID" json:"logs,omitempty"`
}

// ReinforcementLog represents when a reinforcement was triggered
type ReinforcementLog struct {
	ID                     uint      `gorm:"primarykey" json:"id"`
	UserID                 uint      `json:"user_id"`
	ReinforcementSettingID uint      `json:"reinforcement_setting_id"`
	ReinforcementItemID    uint      `json:"reinforcement_item_id"`
	HomeworkID             *uint     `json:"homework_id"`           // Optional, for homework-specific reinforcements
	SessionID              string    `gorm:"type:varchar(100)" json:"session_id"` // Practice session identifier
	TriggerType            string    `gorm:"type:varchar(50)" json:"trigger_type"` // ratio, interval, manual
	TriggerValue           int       `json:"trigger_value"`         // The count or time that triggered it
	ContextData            string    `gorm:"type:text" json:"context_data"` // Additional context as JSON
	CreatedAt              time.Time `json:"created_at"`

	// Relations
	User                 User                 `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ReinforcementSetting ReinforcementSetting `gorm:"foreignKey:ReinforcementSettingID" json:"setting,omitempty"`
	ReinforcementItem    ReinforcementItem    `gorm:"foreignKey:ReinforcementItemID" json:"item,omitempty"`
	Homework             *Homework            `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
}

// Grade represents different grade levels
type Grade struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"type:varchar(50)" json:"name"`        // "一年级", "二年级", etc.
	Code        string `gorm:"type:varchar(20);uniqueIndex" json:"code"` // "grade1", "grade2", etc.
	Description string `gorm:"type:varchar(100)" json:"description"` // Age range description
	AgeMin      int    `json:"age_min"`
	AgeMax      int    `json:"age_max"`
	Order       int    `gorm:"default:0" json:"order"` // Display order
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Subject represents different subjects
type Subject struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	Name        string `gorm:"type:varchar(50)" json:"name"`             // "数学", "语文", etc.
	Code        string `gorm:"type:varchar(20);uniqueIndex" json:"code"` // "math", "chinese", etc.
	Description string `gorm:"type:varchar(200)" json:"description"`
	Icon        string `gorm:"type:varchar(50)" json:"icon"`   // Icon identifier
	Color       string `gorm:"type:varchar(7)" json:"color"`   // Hex color for gradient
	Order       int    `gorm:"default:0" json:"order"`         // Display order
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	Topics []Topic `gorm:"foreignKey:SubjectID" json:"topics,omitempty"`
}

// Topic represents different topics within subjects
type Topic struct {
	ID              uint   `gorm:"primarykey" json:"id"`
	SubjectID       uint   `json:"subject_id"`
	Name            string `gorm:"type:varchar(100)" json:"name"` // "加法", "减法", etc.
	Code            string `gorm:"type:varchar(50)" json:"code"`  // "addition", "subtraction", etc.
	Description     string `gorm:"type:varchar(200)" json:"description"`
	FullDescription string `gorm:"type:text" json:"full_description"`
	Icon            string `gorm:"type:varchar(50)" json:"icon"`  // Icon identifier
	Color           string `gorm:"type:varchar(7)" json:"color"`  // Hex color
	Order           int    `gorm:"default:0" json:"order"`        // Display order within subject
	IsActive        bool   `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relations
	Subject   Subject    `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	Questions []Question `gorm:"foreignKey:Tags" json:"questions,omitempty"` // Questions tagged with this topic
}

// UserPerformance represents user learning statistics
type UserPerformance struct {
	ID                  uint    `gorm:"primarykey" json:"id"`
	UserID              uint    `json:"user_id"`
	Date                string  `gorm:"type:date" json:"date"` // YYYY-MM-DD format
	QuestionsAnswered   int     `gorm:"default:0" json:"questions_answered"`
	QuestionsCorrect    int     `gorm:"default:0" json:"questions_correct"`
	TimeSpent           int     `gorm:"default:0" json:"time_spent"` // minutes
	StreakDays          int     `gorm:"default:0" json:"streak_days"`
	WeeklyTarget        int     `gorm:"default:50" json:"weekly_target"`
	HomeworkCompleted   int     `gorm:"default:0" json:"homework_completed"`
	ReinforcementsEarned int    `gorm:"default:0" json:"reinforcements_earned"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}