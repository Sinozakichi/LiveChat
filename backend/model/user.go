package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 代表一個用戶
type User struct {
	ID         string `gorm:"primaryKey;type:uuid"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	Username   string         `gorm:"size:255;not null;uniqueIndex"`
	Email      string         `gorm:"size:255;not null;uniqueIndex"`
	Password   string         `gorm:"column:password_hash;size:255;not null"` // 存儲哈希後的密碼
	Role       string         `gorm:"size:50;default:'user'"`
	IsVerified bool           `gorm:"default:false"`
}

// BeforeCreate hook在創建用戶前自動生成UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定 User 模型的表名
func (User) TableName() string {
	return "users"
}
