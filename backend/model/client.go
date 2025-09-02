package model

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

// 錯誤定義
var (
	ErrClientInactive = errors.New("客戶端已停用")
)

// Client 代表一個連接到 WebSocket 的用戶
//
// 並發安全設計：
// 1. writeMu 保護 WebSocket 寫入操作，防止並發寫入錯誤
// 2. 提供 SafeWriteMessage 方法確保線程安全的訊息發送
// 3. 所有 WebSocket 寫入操作都應通過 SafeWriteMessage 進行
type Client struct {
	ID         string          // 客戶端唯一識別碼
	Conn       *websocket.Conn // WebSocket 連接
	UserName   string          // 使用者名稱，可選
	RoomID     string          // 當前所在聊天室 ID
	IsActive   bool            // 客戶端是否活躍
	JoinedAt   int64           // 加入時間戳
	LastActive int64           // 最後活躍時間戳
	writeMu    sync.Mutex      // WebSocket 寫入操作保護鎖
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

// SafeWriteMessage 線程安全的 WebSocket 訊息寫入方法
//
// 功能：
// 1. 使用 Mutex 保護 WebSocket 寫入操作
// 2. 防止多個 goroutine 同時寫入造成的 "concurrent write" 錯誤
// 3. 在寫入失敗時自動標記客戶端為非活躍狀態
//
// 參數：
// - messageType: WebSocket 訊息類型 (通常是 websocket.TextMessage)
// - data: 要發送的資料
//
// 回傳：
// - error: 寫入錯誤，如果成功則為 nil
func (c *Client) SafeWriteMessage(messageType int, data []byte) error {
	// 檢查客戶端是否仍然活躍
	if !c.IsActive {
		return ErrClientInactive
	}

	// 使用鎖保護寫入操作
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 執行實際的寫入操作
	err := c.Conn.WriteMessage(messageType, data)
	if err != nil {
		// 寫入失敗時自動停用客戶端
		c.IsActive = false
		return err
	}

	// 寫入成功時更新活躍時間
	c.UpdateActivity()
	return nil
}
