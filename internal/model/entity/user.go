package entity

import (
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleUser    Role = "user"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	Password  string         `gorm:"type:varchar(255)" json:"-"`
	Role      Role           `gorm:"type:enum('user','teacher','admin');default:'user'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
