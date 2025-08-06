package entity

import (
	"time"

	"gorm.io/gorm"
)

type QuestionType string

const (
	TypeChoice QuestionType = "choice" // 选择题
	TypeJudge  QuestionType = "judge"  // 判断题
	TypeFillIn QuestionType = "fillin" // 填空题
	TypeMath   QuestionType = "math"   // 加减法题
)

type Question struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"type:text" json:"title"`
	Type        QuestionType   `gorm:"type:varchar(20)" json:"type"`
	Difficulty  int            `gorm:"type:tinyint;default:1" json:"difficulty"` // 1-5
	Options     string         `gorm:"type:text" json:"options"`                 // JSON格式存储选项
	Answer      string         `gorm:"type:text" json:"answer"`
	Explanation string         `gorm:"type:text" json:"explanation"` // 答案解释
	CreatorID   uint           `json:"creator_id"`
	MediaURL    string         `gorm:"type:varchar(255)" json:"media_url"` // 媒体资源URL
	Tags        string         `gorm:"type:varchar(255)" json:"tags"`      // 逗号分隔的标签
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserAnswer struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	UserID     uint           `json:"user_id"`
	PaperID    uint           `json:"paper_id"`
	QuestionID uint           `json:"question_id"`
	Answer     string         `gorm:"type:text" json:"answer"`
	Score      int            `json:"score"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
