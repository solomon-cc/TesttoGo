package entity

import (
	"time"

	"gorm.io/gorm"
)

type Paper struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	CreatorID    uint           `json:"creator_id"`
	Grade        string         `json:"grade"`        // 年级
	Subject      string         `json:"subject"`      // 科目
	Type         string         `json:"type"`         // practice, exam, training
	Difficulty   string         `json:"difficulty"`   // easy, medium, hard
	Status       string         `json:"status"`       // draft, published, closed
	Questions    string         `json:"questions"`    // 题目ID的JSON字符串
	Duration     int            `json:"duration"`     // 单位：分钟
	StartTime    *time.Time     `json:"start_time"`   // 开始时间
	EndTime      *time.Time     `json:"end_time"`     // 结束时间
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Creator User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}
