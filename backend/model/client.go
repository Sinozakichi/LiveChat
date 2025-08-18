package model

import (
	"github.com/gorilla/websocket"
)

// Client 代表一個連接到 WebSocket 的用戶
type Client struct {
	ID         string          // 客戶端唯一識別碼
	Conn       *websocket.Conn // WebSocket 連接
	UserName   string          // 使用者名稱，可選
	RoomID     string          // 當前所在聊天室 ID
	IsActive   bool            // 客戶端是否活躍
	JoinedAt   int64           // 加入時間戳
	LastActive int64           // 最後活躍時間戳
}

// NewClient 創建一個新的客戶端
func NewClient(id string, conn *websocket.Conn) *Client {
	now := getCurrentTimestamp()
	return &Client{
		ID:         id,
		Conn:       conn,
		RoomID:     "",
		IsActive:   true,
		JoinedAt:   now,
		LastActive: now,
	}
}

// SetUserName 設置客戶端的使用者名稱
func (c *Client) SetUserName(name string) {
	c.UserName = name
}

// SetRoomID 設置客戶端的聊天室 ID
func (c *Client) SetRoomID(roomID string) {
	c.RoomID = roomID
}

// UpdateActivity 更新客戶端的活躍狀態
func (c *Client) UpdateActivity() {
	c.LastActive = getCurrentTimestamp()
}

// Deactivate 將客戶端標記為非活躍
func (c *Client) Deactivate() {
	c.IsActive = false
}
