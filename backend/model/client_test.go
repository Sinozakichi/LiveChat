package model

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// 測試創建新客戶端
func TestNewClient(t *testing.T) {
	// 安排 (Arrange)
	mockTime := time.Date(2025, 8, 1, 10, 0, 0, 0, time.UTC)
	mockTimestamp := mockTime.Unix()
	SetTimeNow(func() time.Time {
		return mockTime
	})
	defer ResetTimeNow()

	// 動作 (Act)
	client := NewClient("test-id", nil)

	// 斷言 (Assert)
	assert.Equal(t, "test-id", client.ID, "客戶端 ID 應該匹配")
	assert.Nil(t, client.Conn, "WebSocket 連接應該為 nil")
	assert.True(t, client.IsActive, "新客戶端應該是活躍的")
	assert.Equal(t, mockTimestamp, client.JoinedAt, "加入時間應該正確")
	assert.Equal(t, mockTimestamp, client.LastActive, "最後活躍時間應該正確")
	assert.Empty(t, client.UserName, "使用者名稱應該為空")
}

// 測試設置使用者名稱
func TestSetUserName(t *testing.T) {
	// 安排 (Arrange)
	client := NewClient("test-id", nil)

	// 動作 (Act)
	client.SetUserName("TestUser")

	// 斷言 (Assert)
	assert.Equal(t, "TestUser", client.UserName, "使用者名稱應該已設置")
}

// 測試更新活躍狀態
func TestUpdateActivity(t *testing.T) {
	// 安排 (Arrange)
	mockTime1 := time.Date(2025, 8, 1, 10, 0, 0, 0, time.UTC)
	SetTimeNow(func() time.Time {
		return mockTime1
	})
	client := NewClient("test-id", nil)
	initialLastActive := client.LastActive

	// 設置新的模擬時間
	mockTime2 := time.Date(2025, 8, 1, 10, 5, 0, 0, time.UTC)
	SetTimeNow(func() time.Time {
		return mockTime2
	})

	// 動作 (Act)
	client.UpdateActivity()

	// 斷言 (Assert)
	assert.NotEqual(t, initialLastActive, client.LastActive, "最後活躍時間應該已更新")
	assert.Equal(t, mockTime2.Unix(), client.LastActive, "最後活躍時間應該匹配新的時間戳")
	ResetTimeNow()
}

// 測試停用客戶端
func TestDeactivate(t *testing.T) {
	// 安排 (Arrange)
	client := NewClient("test-id", nil)
	assert.True(t, client.IsActive, "新客戶端應該是活躍的")

	// 動作 (Act)
	client.Deactivate()

	// 斷言 (Assert)
	assert.False(t, client.IsActive, "客戶端應該被標記為非活躍")
}

// 測試帶有實際 WebSocket 連接的客戶端
func TestClientWithConnection(t *testing.T) {
	// 這個測試需要模擬 WebSocket 連接
	// 在實際測試中，我們可能需要使用 mock 或 stub
	// 這裡我們只是驗證可以創建帶有連接的客戶端

	// 安排 (Arrange)
	mockConn := &websocket.Conn{}

	// 動作 (Act)
	client := NewClient("test-id", mockConn)

	// 斷言 (Assert)
	assert.NotNil(t, client.Conn, "WebSocket 連接不應該為 nil")
	assert.Equal(t, mockConn, client.Conn, "WebSocket 連接應該匹配")
}
