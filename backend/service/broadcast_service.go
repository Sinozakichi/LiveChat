package service

import (
	"errors"
	"fmt"
	"livechat/backend/model"
	"livechat/backend/repository"
	"time"

	"github.com/gorilla/websocket"
)

// 定義錯誤
var (
	ErrEmptyMessage = errors.New("訊息不能為空")
	ErrNoClients    = errors.New("沒有連接的客戶端")
)

// MessageType 定義訊息類型
type MessageType int

const (
	TextMessage MessageType = iota
	SystemMessage
	ErrorMessage
)

// ChatMessage 代表一個聊天訊息
type ChatMessage struct {
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	Sender    string      `json:"sender,omitempty"`
	RoomID    string      `json:"roomId,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// BroadcastService 處理消息廣播邏輯
type BroadcastService struct {
	clientRepo   *repository.ClientRepository
	messageLog   map[string][]ChatMessage // 按聊天室 ID 組織訊息日誌
	maxLogSize   int
	errorHandler func(error)
}

// BroadcastServiceOption 定義服務選項
type BroadcastServiceOption func(*BroadcastService)

// WithMaxLogSize 設置最大日誌大小
func WithMaxLogSize(size int) BroadcastServiceOption {
	return func(s *BroadcastService) {
		s.maxLogSize = size
	}
}

// WithErrorHandler 設置錯誤處理函數
func WithErrorHandler(handler func(error)) BroadcastServiceOption {
	return func(s *BroadcastService) {
		s.errorHandler = handler
	}
}

// NewBroadcastService 創建一個新的廣播服務
func NewBroadcastService(clientRepo *repository.ClientRepository, opts ...BroadcastServiceOption) *BroadcastService {
	service := &BroadcastService{
		clientRepo:   clientRepo,
		messageLog:   make(map[string][]ChatMessage),
		maxLogSize:   100, // 默認最多保存 100 條訊息
		errorHandler: func(err error) { fmt.Println("Error:", err) },
	}

	// 應用選項
	for _, opt := range opts {
		opt(service)
	}

	return service
}

// AddClient 添加一個新的客戶端
func (s *BroadcastService) AddClient(client *model.Client) error {
	if client == nil {
		return errors.New("客戶端不能為空")
	}

	err := s.clientRepo.Add(client)
	if err != nil {
		return err
	}

	// 如果客戶端已加入聊天室，發送系統訊息通知
	if client.RoomID != "" {
		// 發送系統訊息通知新客戶端加入
		systemMsg := ChatMessage{
			Type:      SystemMessage,
			Content:   "新用戶加入聊天室",
			RoomID:    client.RoomID,
			Timestamp: time.Now().Unix(),
		}

		s.logMessage(systemMsg)

		// 廣播系統訊息給聊天室的其他用戶
		s.BroadcastToRoom(client.RoomID, []byte(systemMsg.Content))
	}

	return nil
}

// RemoveClient 移除一個客戶端
func (s *BroadcastService) RemoveClient(clientID string) error {
	return s.clientRepo.Remove(clientID)
}

// GetClient 獲取一個客戶端
func (s *BroadcastService) GetClient(clientID string) (*model.Client, error) {
	return s.clientRepo.Get(clientID)
}

// BroadcastMessage 向所有連接的客戶端廣播消息
func (s *BroadcastService) BroadcastMessage(message []byte) error {
	if len(message) == 0 {
		return ErrEmptyMessage
	}

	clients := s.clientRepo.GetActiveClients()
	if len(clients) == 0 {
		return ErrNoClients
	}

	// 記錄訊息
	chatMsg := ChatMessage{
		Type:      TextMessage,
		Content:   string(message),
		Timestamp: time.Now().Unix(),
	}
	s.logMessage(chatMsg)

	// 廣播訊息
	for _, client := range clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			s.handleClientError(client, err)
		} else {
			client.UpdateActivity()
		}
	}

	return nil
}

// BroadcastToRoom 向特定聊天室的所有客戶端廣播消息
func (s *BroadcastService) BroadcastToRoom(roomID string, message []byte) error {
	if len(message) == 0 {
		return ErrEmptyMessage
	}

	if roomID == "" {
		return errors.New("聊天室 ID 不能為空")
	}

	clients := s.clientRepo.GetActiveClients()
	if len(clients) == 0 {
		return ErrNoClients
	}

	// 記錄訊息
	chatMsg := ChatMessage{
		Type:      TextMessage,
		Content:   string(message),
		RoomID:    roomID,
		Timestamp: time.Now().Unix(),
	}
	s.logMessage(chatMsg)

	// 廣播訊息到特定聊天室
	roomClients := 0
	for _, client := range clients {
		if client.RoomID == roomID {
			roomClients++
			err := client.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				s.handleClientError(client, err)
			} else {
				client.UpdateActivity()
			}
		}
	}

	if roomClients == 0 {
		return errors.New("聊天室中沒有活躍的客戶端")
	}

	return nil
}

// SendPrivateMessage 發送私人訊息給指定客戶端
func (s *BroadcastService) SendPrivateMessage(targetID string, message []byte) error {
	if len(message) == 0 {
		return ErrEmptyMessage
	}

	client, err := s.clientRepo.Get(targetID)
	if err != nil {
		return err
	}

	if !client.IsActive {
		return errors.New("客戶端不活躍")
	}

	err = client.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		s.handleClientError(client, err)
		return err
	}

	client.UpdateActivity()
	return nil
}

// GetAllMessageHistory 獲取所有訊息歷史
func (s *BroadcastService) GetAllMessageHistory() map[string][]ChatMessage {
	return s.messageLog
}

// 處理客戶端錯誤
func (s *BroadcastService) handleClientError(client *model.Client, err error) {
	s.errorHandler(fmt.Errorf("客戶端 %s 錯誤: %w", client.ID, err))
	client.Deactivate()
	if client.Conn != nil {
		client.Conn.Close()
	}
}

// 記錄訊息
func (s *BroadcastService) logMessage(msg ChatMessage) {
	roomID := msg.RoomID
	if roomID == "" {
		roomID = "global" // 全局訊息使用 "global" 作為鍵
	}

	// 確保聊天室的訊息日誌已初始化
	if _, exists := s.messageLog[roomID]; !exists {
		s.messageLog[roomID] = make([]ChatMessage, 0)
	}

	// 添加訊息到聊天室的日誌
	s.messageLog[roomID] = append(s.messageLog[roomID], msg)

	// 如果超過最大日誌大小，刪除最舊的訊息
	if len(s.messageLog[roomID]) > s.maxLogSize {
		s.messageLog[roomID] = s.messageLog[roomID][1:]
	}
}

// GetMessageHistory 獲取特定聊天室的訊息歷史
func (s *BroadcastService) GetMessageHistory(roomID string) []ChatMessage {
	if roomID == "" {
		roomID = "global" // 全局訊息使用 "global" 作為鍵
	}

	if messages, exists := s.messageLog[roomID]; exists {
		return messages
	}

	return []ChatMessage{}
}

// GetClientsInRoom 獲取特定聊天室的所有客戶端
func (s *BroadcastService) GetClientsInRoom(roomID string) []*model.Client {
	clients := s.clientRepo.GetActiveClients()
	var roomClients []*model.Client

	for _, client := range clients {
		if client.RoomID == roomID {
			roomClients = append(roomClients, client)
		}
	}

	return roomClients
}
