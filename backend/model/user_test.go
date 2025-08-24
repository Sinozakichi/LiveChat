package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 測試 User 模型的基本屬性
func TestUserModel(t *testing.T) {
	// 安排 (Arrange)
	user := User{
		ID:         "1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "hashedpassword",
		Role:       "user",
		IsVerified: true,
	}

	// 斷言 (Assert)
	assert.Equal(t, "1", user.ID, "User ID 應該匹配")
	assert.Equal(t, "testuser", user.Username, "User 用戶名應該匹配")
	assert.Equal(t, "test@example.com", user.Email, "User 電子郵件應該匹配")
	assert.Equal(t, "hashedpassword", user.Password, "User 密碼應該匹配")
	assert.Equal(t, "user", user.Role, "User 角色應該匹配")
	assert.True(t, user.IsVerified, "User 應該已驗證")
}

// 測試 User 模型的表名
func TestUserTableName(t *testing.T) {
	// 安排 (Arrange)
	user := User{}

	// 動作 (Act)
	tableName := user.TableName()

	// 斷言 (Assert)
	assert.Equal(t, "users", tableName, "User 表名應該為 users")
}
