package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// 測試 Room 模型的基本屬性
func TestRoomModel(t *testing.T) {
	// 安排 (Arrange)
	room := Room{
		Model: gorm.Model{
			ID: 1,
		},
		Name:        "測試聊天室",
		Description: "這是一個測試聊天室",
		IsPublic:    true,
		MaxUsers:    100,
		CreatedBy:   "user-123",
		IsActive:    true,
	}

	// 斷言 (Assert)
	assert.Equal(t, uint(1), room.ID, "Room ID 應該匹配")
	assert.Equal(t, "測試聊天室", room.Name, "Room 名稱應該匹配")
	assert.Equal(t, "這是一個測試聊天室", room.Description, "Room 描述應該匹配")
	assert.True(t, room.IsPublic, "Room 應該是公開的")
	assert.Equal(t, 100, room.MaxUsers, "Room 最大用戶數應該匹配")
	assert.Equal(t, "user-123", room.CreatedBy, "Room 創建者應該匹配")
	assert.True(t, room.IsActive, "Room 應該是活躍的")
}

// 測試 RoomUser 模型的基本屬性
func TestRoomUserModel(t *testing.T) {
	// 安排 (Arrange)
	now := time.Now()
	roomUser := RoomUser{
		Model: gorm.Model{
			ID: 1,
		},
		RoomID:       "room-123",
		UserID:       "user-123",
		Role:         "member",
		JoinedAt:     now,
		LastActiveAt: now,
		IsActive:     true,
	}

	// 斷言 (Assert)
	assert.Equal(t, uint(1), roomUser.ID, "RoomUser ID 應該匹配")
	assert.Equal(t, "room-123", roomUser.RoomID, "RoomUser 聊天室 ID 應該匹配")
	assert.Equal(t, "user-123", roomUser.UserID, "RoomUser 用戶 ID 應該匹配")
	assert.Equal(t, "member", roomUser.Role, "RoomUser 角色應該匹配")
	assert.Equal(t, now, roomUser.JoinedAt, "RoomUser 加入時間應該匹配")
	assert.Equal(t, now, roomUser.LastActiveAt, "RoomUser 最後活躍時間應該匹配")
	assert.True(t, roomUser.IsActive, "RoomUser 應該是活躍的")
}

// 測試 Message 模型的基本屬性
func TestMessageModel(t *testing.T) {
	// 安排 (Arrange)
	now := time.Now()
	message := Message{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: now,
		},
		RoomID:          "room-123",
		UserID:          "user-123",
		Content:         "Hello, World!",
		IsSystemMessage: false,
	}

	// 斷言 (Assert)
	assert.Equal(t, uint(1), message.ID, "Message ID 應該匹配")
	assert.Equal(t, "room-123", message.RoomID, "Message 聊天室 ID 應該匹配")
	assert.Equal(t, "user-123", message.UserID, "Message 用戶 ID 應該匹配")
	assert.Equal(t, "Hello, World!", message.Content, "Message 內容應該匹配")
	assert.False(t, message.IsSystemMessage, "Message 不應該是系統訊息")
	assert.Equal(t, now, message.CreatedAt, "Message 創建時間應該匹配")
}
