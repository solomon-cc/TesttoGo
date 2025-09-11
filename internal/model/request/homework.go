package request

import "time"

// CreateHomeworkRequest represents the request to create homework
type CreateHomeworkRequest struct {
	Title                 string                     `json:"title" binding:"required,min=1,max=200"`
	Description           string                     `json:"description"`
	Grade                 string                     `json:"grade" binding:"required"`
	Subject               string                     `json:"subject" binding:"required"`
	ScheduleType          string                     `json:"schedule_type" binding:"required,oneof=weekly daily"`
	StartDate             time.Time                  `json:"start_date" binding:"required"`
	EndDate               time.Time                  `json:"end_date" binding:"required"`
	QuestionsPerDay       int                        `json:"questions_per_day" binding:"min=1,max=100"`
	ShowHints             bool                       `json:"show_hints"`
	ReinforcementSettings map[string]interface{}     `json:"reinforcement_settings"`
	StudentAssignments    []HomeworkAssignmentRequest `json:"student_assignments"`
	Questions             []HomeworkQuestionRequest  `json:"questions"`
}

// HomeworkAssignmentRequest represents student assignment within homework
type HomeworkAssignmentRequest struct {
	StudentID              uint `json:"student_id" binding:"required"`
	ReinforcementSettingID *uint `json:"reinforcement_setting_id"`
}

// HomeworkQuestionRequest represents questions to include in homework
type HomeworkQuestionRequest struct {
	QuestionID uint `json:"question_id" binding:"required"`
	DayOfWeek  int  `json:"day_of_week" binding:"min=0,max=7"`
	Order      int  `json:"order"`
}

// UpdateHomeworkRequest represents the request to update homework
type UpdateHomeworkRequest struct {
	Title                 *string                    `json:"title,omitempty"`
	Description           *string                    `json:"description,omitempty"`
	Status                *string                    `json:"status,omitempty"`
	EndDate               *time.Time                 `json:"end_date,omitempty"`
	QuestionsPerDay       *int                       `json:"questions_per_day,omitempty"`
	ShowHints             *bool                      `json:"show_hints,omitempty"`
	ReinforcementSettings *map[string]interface{}    `json:"reinforcement_settings,omitempty"`
}

// CopyHomeworkRequest represents the request to copy homework
type CopyHomeworkRequest struct {
	NewTitle      string    `json:"new_title" binding:"required"`
	NewStartDate  time.Time `json:"new_start_date" binding:"required"`
	NewEndDate    time.Time `json:"new_end_date" binding:"required"`
	CopyStudents  bool      `json:"copy_students"`
	CopyQuestions bool      `json:"copy_questions"`
	StudentIDs    []uint    `json:"student_ids"`
}

// SubmitHomeworkRequest represents homework submission
type SubmitHomeworkRequest struct {
	HomeworkID      uint                            `json:"homework_id" binding:"required"`
	SubmissionDate  time.Time                       `json:"submission_date" binding:"required"`
	QuestionAnswers []HomeworkQuestionAnswerRequest `json:"question_answers" binding:"required"`
	TimeSpent       int                             `json:"time_spent"`
}

// HomeworkQuestionAnswerRequest represents individual question answer in submission
type HomeworkQuestionAnswerRequest struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
	TimeSpent  int    `json:"time_spent"`
}

// AdjustHomeworkRequest represents homework adjustment by teacher
type AdjustHomeworkRequest struct {
	AdjustType  string                 `json:"adjust_type" binding:"required"`
	Description string                 `json:"description"`
	Changes     map[string]interface{} `json:"changes" binding:"required"`
}

// ListHomeworkRequest represents query parameters for listing homework
type ListHomeworkRequest struct {
	Page         int    `form:"page" binding:"min=1"`
	PageSize     int    `form:"page_size" binding:"min=1,max=100"`
	Status       string `form:"status"`
	Grade        string `form:"grade"`
	Subject      string `form:"subject"`
	StudentID    uint   `form:"student_id"`
	CreatorID    uint   `form:"creator_id"`
	DateFrom     string `form:"date_from"`
	DateTo       string `form:"date_to"`
}

// HomeworkSubmissionQuery represents query parameters for homework submissions
type HomeworkSubmissionQuery struct {
	Page         int    `form:"page" binding:"min=1"`
	PageSize     int    `form:"page_size" binding:"min=1,max=100"`
	StudentID    uint   `form:"student_id"`
	IsCompleted  *bool  `form:"is_completed"`
	DateFrom     string `form:"date_from"`
	DateTo       string `form:"date_to"`
}