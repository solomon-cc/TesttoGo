package response

import "time"

// HomeworkResponse represents homework data in API responses
type HomeworkResponse struct {
	ID                    uint                         `json:"id"`
	Title                 string                       `json:"title"`
	Description           string                       `json:"description"`
	CreatorID             uint                         `json:"creator_id"`
	CreatorName           string                       `json:"creator_name"`
	TeacherName           string                       `json:"teacher_name"`  // 为前端兼容性添加
	Grade                 string                       `json:"grade"`
	Subject               string                       `json:"subject"`
	Status                string                       `json:"status"`
	ScheduleType          string                       `json:"schedule_type"`
	StartDate             *time.Time                   `json:"start_date,omitempty"`
	EndDate               *time.Time                   `json:"end_date,omitempty"`
	QuestionsPerDay       int                          `json:"questions_per_day"`
	ShowHints             bool                         `json:"show_hints"`
	ReinforcementSettings map[string]interface{}       `json:"reinforcement_settings"`
	IsCompleted           bool                         `json:"is_completed"`
	CreatedAt             time.Time                    `json:"created_at"`
	UpdatedAt             time.Time                    `json:"updated_at"`
	Assignments           []HomeworkAssignmentResponse `json:"assignments,omitempty"`
	Questions             []HomeworkQuestionResponse   `json:"questions,omitempty"`
	Submissions           []HomeworkSubmissionResponse `json:"submissions,omitempty"`
	Stats                 *HomeworkStatsResponse       `json:"stats,omitempty"`
}

// HomeworkAssignmentResponse represents homework assignment data
type HomeworkAssignmentResponse struct {
	ID                     uint                          `json:"id"`
	StudentID              uint                          `json:"student_id"`
	StudentName            string                        `json:"student_name"`
	ReinforcementSettingID *uint                         `json:"reinforcement_setting_id"`
	ReinforcementSetting   *ReinforcementSettingResponse `json:"reinforcement_setting,omitempty"`
	CreatedAt              time.Time                     `json:"created_at"`
}

// HomeworkQuestionResponse represents homework question data
type HomeworkQuestionResponse struct {
	ID         uint             `json:"id"`
	QuestionID uint             `json:"question_id"`
	Question   *QuestionResponse `json:"question,omitempty"`
	DayOfWeek  int              `json:"day_of_week"`
	Order      int              `json:"order"`
	CreatedAt  time.Time        `json:"created_at"`
}

// HomeworkSubmissionResponse represents homework submission data
type HomeworkSubmissionResponse struct {
	ID               uint                              `json:"id"`
	StudentID        uint                              `json:"student_id"`
	StudentName      string                            `json:"student_name"`
	SubmissionDate   time.Time                         `json:"submission_date"`
	QuestionsTotal   int                               `json:"questions_total"`
	QuestionsCorrect int                               `json:"questions_correct"`
	TimeSpent        int                               `json:"time_spent"`
	Score            int                               `json:"score"`
	IsCompleted      bool                              `json:"is_completed"`
	QuestionAnswers  []HomeworkQuestionAnswerResponse  `json:"question_answers,omitempty"`
	CreatedAt        time.Time                         `json:"created_at"`
	UpdatedAt        time.Time                         `json:"updated_at"`
}

// HomeworkQuestionAnswerResponse represents individual question answer
type HomeworkQuestionAnswerResponse struct {
	ID         uint             `json:"id"`
	QuestionID uint             `json:"question_id"`
	Question   *QuestionResponse `json:"question,omitempty"`
	Answer     string           `json:"answer"`
	IsCorrect  bool             `json:"is_correct"`
	TimeSpent  int              `json:"time_spent"`
	CreatedAt  time.Time        `json:"created_at"`
}

// HomeworkStatsResponse represents homework statistics
type HomeworkStatsResponse struct {
	TotalStudents         int                            `json:"total_students"`
	CompletedSubmissions  int                            `json:"completed_submissions"`
	PendingSubmissions    int                            `json:"pending_submissions"`
	AverageScore          float64                        `json:"average_score"`
	AverageTimeSpent      float64                        `json:"average_time_spent"`
	CompletionRate        float64                        `json:"completion_rate"`
	DifficultyStats       map[string]int                 `json:"difficulty_stats"`
	StudentProgress       []StudentProgressResponse      `json:"student_progress"`
	ReinforcementStats    *ReinforcementStatsResponse    `json:"reinforcement_stats,omitempty"`
}

// StudentProgressResponse represents individual student's progress
type StudentProgressResponse struct {
	StudentID            uint      `json:"student_id"`
	StudentName          string    `json:"student_name"`
	QuestionsTotal       int       `json:"questions_total"`
	QuestionsCorrect     int       `json:"questions_correct"`
	Score                int       `json:"score"`
	TimeSpent            int       `json:"time_spent"`
	IsCompleted          bool      `json:"is_completed"`
	LastSubmissionDate   *time.Time `json:"last_submission_date"`
	ReinforcementsEarned int       `json:"reinforcements_earned"`
}

// HomeworkListResponse represents paginated homework list
type HomeworkListResponse struct {
	Items       []HomeworkResponse `json:"items"`
	Total       int64              `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalPages  int                `json:"total_pages"`
}

// HomeworkAdjustmentResponse represents homework adjustment data
type HomeworkAdjustmentResponse struct {
	ID          uint                   `json:"id"`
	TeacherID   uint                   `json:"teacher_id"`
	TeacherName string                 `json:"teacher_name"`
	AdjustType  string                 `json:"adjust_type"`
	Description string                 `json:"description"`
	Changes     map[string]interface{} `json:"changes"`
	CreatedAt   time.Time              `json:"created_at"`
}