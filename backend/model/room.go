package model

import (
	"time"

	"gorm.io/gorm"
)

// Room 代表一個聊天室
type Room struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	Name        string         `gorm:"size:255;not null"`
	Description string         `gorm:"type:text"`
	IsPublic    bool           `gorm:"default:true"`
	MaxUsers    int            `gorm:"default:100"`
	CreatedBy   string         `gorm:"size:255"`
	IsActive    bool           `gorm:"default:true"`
}

// RoomUser 代表用戶與聊天室的關聯
type RoomUser struct {
	gorm.Model
	RoomID       string    `gorm:"size:255;index"`
	UserID       string    `gorm:"size:255;index"`
	Role         string    `gorm:"size:20;default:'member'"`
	JoinedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	LastActiveAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	IsActive     bool      `gorm:"default:true"`
}

// Message 代表聊天訊息
type Message struct {
	gorm.Model
	RoomID          string `gorm:"size:255;index"`
	UserID          string `gorm:"size:255;index"`
	Content         string `gorm:"type:text;not null"`
	IsSystemMessage bool   `gorm:"default:false"`
}

// TableName 指定 Room 模型的表名
func (Room) TableName() string {
	return "rooms"
}

// TableName 指定 RoomUser 模型的表名
func (RoomUser) TableName() string {
	return "room_users"
}

// TableName 指定 Message 模型的表名
func (Message) TableName() string {
	return "messages"
}
