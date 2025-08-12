package response

import "time"

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