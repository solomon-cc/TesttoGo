package entity

import (
	"time"

	"gorm.io/gorm"
)

type HomeworkStatus string

const (
	HomeworkStatusDraft     HomeworkStatus = "draft"
	HomeworkStatusActive    HomeworkStatus = "active"
	HomeworkStatusCompleted HomeworkStatus = "completed"
	HomeworkStatusArchived  HomeworkStatus = "archived"
)

type HomeworkScheduleType string

const (
	HomeworkScheduleWeekly HomeworkScheduleType = "weekly"
	HomeworkScheduleDaily  HomeworkScheduleType = "daily"
)

// Homework represents a homework assignment created by teachers/admins
type Homework struct {
	ID                    uint                 `gorm:"primarykey" json:"id"`
	Title                 string               `gorm:"type:varchar(200)" json:"title"`
	Description           string               `gorm:"type:text" json:"description"`
	CreatorID             uint                 `json:"creator_id"`
	Grade                 string               `gorm:"type:varchar(20)" json:"grade"`
	Subject               string               `gorm:"type:varchar(50)" json:"subject"`
	Status                HomeworkStatus       `gorm:"type:enum('draft','active','completed','archived');default:'draft'" json:"status"`
	ScheduleType          HomeworkScheduleType `gorm:"type:enum('weekly','daily');default:'daily'" json:"schedule_type"`
	StartDate             time.Time            `json:"start_date"`
	EndDate               time.Time            `json:"end_date"`
	QuestionsPerDay       int                  `gorm:"default:10" json:"questions_per_day"`
	IsTimeLimited         bool                 `gorm:"default:false" json:"is_time_limited"`
	TimeLimit             int                  `gorm:"default:0" json:"time_limit"` // minutes
	ShowHints             bool                 `gorm:"default:true" json:"show_hints"`
	ReinforcementSettings string               `gorm:"type:json" json:"reinforcement_settings"` // JSON data
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
	DeletedAt             gorm.DeletedAt       `gorm:"index" json:"-"`

	// Relations
	Creator              User                   `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
	HomeworkAssignments  []HomeworkAssignment   `gorm:"foreignKey:HomeworkID" json:"assignments,omitempty"`
	HomeworkQuestions    []HomeworkQuestion     `gorm:"foreignKey:HomeworkID" json:"questions,omitempty"`
	HomeworkSubmissions  []HomeworkSubmission   `gorm:"foreignKey:HomeworkID" json:"submissions,omitempty"`
	HomeworkAdjustments  []HomeworkAdjustment   `gorm:"foreignKey:HomeworkID" json:"adjustments,omitempty"`
}

// HomeworkAssignment represents assignment of homework to specific students
type HomeworkAssignment struct {
	ID                     uint    `gorm:"primarykey" json:"id"`
	HomeworkID             uint    `json:"homework_id"`
	StudentID              uint    `json:"student_id"`
	ReinforcementSettingID *uint   `json:"reinforcement_setting_id"` // Override homework's default setting
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`

	// Relations
	Homework             Homework             `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
	Student              User                 `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	ReinforcementSetting *ReinforcementSetting `gorm:"foreignKey:ReinforcementSettingID" json:"reinforcement_setting,omitempty"`
}

// HomeworkQuestion represents questions included in homework
type HomeworkQuestion struct {
	ID         uint `gorm:"primarykey" json:"id"`
	HomeworkID uint `json:"homework_id"`
	QuestionID uint `json:"question_id"`
	DayOfWeek  int  `gorm:"comment:'1-7, Monday to Sunday, 0 for all days'" json:"day_of_week"`
	Order      int  `gorm:"default:0" json:"order"`
	CreatedAt  time.Time `json:"created_at"`

	// Relations
	Homework Homework `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
	Question Question `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

// HomeworkSubmission represents student's homework submission
type HomeworkSubmission struct {
	ID               uint      `gorm:"primarykey" json:"id"`
	HomeworkID       uint      `json:"homework_id"`
	StudentID        uint      `json:"student_id"`
	SubmissionDate   time.Time `json:"submission_date"`
	QuestionsTotal   int       `json:"questions_total"`
	QuestionsCorrect int       `json:"questions_correct"`
	TimeSpent        int       `json:"time_spent"` // minutes
	Score            int       `json:"score"`
	IsCompleted      bool      `gorm:"default:false" json:"is_completed"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relations
	Homework          Homework                  `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
	Student           User                      `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	QuestionAnswers   []HomeworkQuestionAnswer  `gorm:"foreignKey:SubmissionID" json:"question_answers,omitempty"`
}

// HomeworkQuestionAnswer represents individual question answers in homework submission
type HomeworkQuestionAnswer struct {
	ID           uint   `gorm:"primarykey" json:"id"`
	SubmissionID uint   `json:"submission_id"`
	QuestionID   uint   `json:"question_id"`
	Answer       string `gorm:"type:text" json:"answer"`
	IsCorrect    bool   `json:"is_correct"`
	TimeSpent    int    `json:"time_spent"` // seconds
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	Submission HomeworkSubmission `gorm:"foreignKey:SubmissionID" json:"submission,omitempty"`
	Question   Question           `gorm:"foreignKey:QuestionID" json:"question,omitempty"`
}

// HomeworkAdjustment represents teacher's mid-assignment adjustments
type HomeworkAdjustment struct {
	ID          uint   `gorm:"primarykey" json:"id"`
	HomeworkID  uint   `json:"homework_id"`
	TeacherID   uint   `json:"teacher_id"`
	AdjustType  string `gorm:"type:varchar(50)" json:"adjust_type"` // add_questions, remove_questions, change_difficulty, etc.
	Description string `gorm:"type:text" json:"description"`
	Changes     string `gorm:"type:json" json:"changes"` // JSON data of specific changes
	CreatedAt   time.Time `json:"created_at"`

	// Relations
	Homework Homework `gorm:"foreignKey:HomeworkID" json:"homework,omitempty"`
	Teacher  User     `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
}