package entity

import (
	"time"

	"gorm.io/gorm"
)

type Paper struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `json:"title"`
	CreatorID uint           `json:"creator_id"`
	Questions string         `json:"questions"` // 题目ID的JSON字符串
	Duration  int            `json:"duration"`  // 单位：分钟
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
