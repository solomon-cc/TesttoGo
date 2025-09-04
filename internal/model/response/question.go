package response

import "time"

// QuestionResponse 题目响应
type QuestionResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Difficulty  int       `json:"difficulty"`
	Options     string    `json:"options,omitempty"`
	Answer      string    `json:"answer,omitempty"` // 仅对教师/管理员显示
	Explanation string    `json:"explanation,omitempty"`
	MediaURL    string    `json:"media_url,omitempty"`
	Tags        string    `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// QuestionAnswerResponse 单题答题响应
type QuestionAnswerResponse struct {
	QuestionID  uint      `json:"question_id"`
	UserAnswer  string    `json:"user_answer"`
	IsCorrect   bool      `json:"is_correct"`
	Score       int       `json:"score"`
	Explanation string    `json:"explanation,omitempty"` // 答案解释
	AnsweredAt  time.Time `json:"answered_at"`
}

// UserAnswerHistoryResponse 用户答题历史响应
type UserAnswerHistoryResponse struct {
	ID          uint      `json:"id"`
	QuestionID  uint      `json:"question_id"`
	PaperID     uint      `json:"paper_id,omitempty"`
	Title       string    `json:"question_title"`
	Type        string    `json:"question_type"`
	UserAnswer  string    `json:"user_answer"`
	IsCorrect   bool      `json:"is_correct"`
	Score       int       `json:"score"`
	AnsweredAt  time.Time `json:"answered_at"`
}

// QuestionStatisticsResponse 题目答题统计响应
type QuestionStatisticsResponse struct {
	QuestionID    uint    `json:"question_id"`
	Title         string  `json:"title"`
	TotalAttempts int64   `json:"total_attempts"`
	CorrectCount  int64   `json:"correct_count"`
	AccuracyRate  float64 `json:"accuracy_rate"`
}

// QuestionWithStatsResponse 带统计数据的题目响应（用于列表）
type QuestionWithStatsResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Difficulty  int       `json:"difficulty"`
	Grade       string    `json:"grade"`
	Subject     string    `json:"subject"`
	Topic       string    `json:"topic"`
	Options     string    `json:"options"`
	Answer      string    `json:"answer"`
	Explanation string    `json:"explanation"`
	CreatorID   uint      `json:"creator_id"`
	MediaURL    string    `json:"media_url"`
	Tags        string    `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// 统计字段
	UsageCount   int64   `json:"usageCount"`   // 使用次数（总答题次数）
	CorrectRate  float64 `json:"correctRate"`  // 答对率
}