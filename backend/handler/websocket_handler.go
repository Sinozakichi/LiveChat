package handler

import (
	"encoding/json"
	"fmt"
	"livechat/backend/model"
	"livechat/backend/service"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// MessagePayload 定義前端發送的訊息格式
type MessagePayload struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Target  string `json:"target,omitempty"` // 用於私人訊息
}

// BroadcastService 定義了廣播服務的接口
type BroadcastService interface {
	AddClient(client *model.Client) error
	RemoveClient(clientID string) error
	BroadcastMessage(message []byte) error
	BroadcastToRoom(roomID string, message []byte) error
	SendPrivateMessage(targetID string, message []byte) error
	GetClient(clientID string) (*model.Client, error)
	GetMessageHistory(roomID string) []service.ChatMessage
	GetClientsInRoom(roomID string) []*model.Client
}

// WebSocketHandler 處理 WebSocket 連接
type WebSocketHandler struct {
	upgrader         websocket.Upgrader
	broadcastService BroadcastService
	logger           Logger
}

// HandlerOption 定義處理器選項
type HandlerOption func(*WebSocketHandler)

// WithLogger 設置日誌記錄器
func WithLogger(logger Logger) HandlerOption {
	return func(h *WebSocketHandler) {
		h.logger = logger
	}
}

// WithCheckOrigin 設置來源檢查函數
func WithCheckOrigin(checkOrigin func(r *http.Request) bool) HandlerOption {
	return func(h *WebSocketHandler) {
		h.upgrader.CheckOrigin = checkOrigin
	}
}

// Logger 定義日誌接口
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// DefaultLogger 默認日誌實現
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("INFO: "+msg+"\n", args...)
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("ERROR: "+msg+"\n", args...)
}

// NewWebSocketHandler 創建一個新的 WebSocket 處理器
func NewWebSocketHandler(broadcastService BroadcastService, opts ...HandlerOption) *WebSocketHandler {
	h := &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 默認允許所有來源
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		broadcastService: broadcastService,
		logger:           &DefaultLogger{},
	}

	// 應用選項
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// HandleConnection 處理新的 WebSocket 連接
func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // 或限定域名

	// 將 HTTP 連接升級為 WebSocket 連接
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	// 設置連接參數
	conn.SetReadLimit(4096) // 限制讀取大小
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 為每個新連接創建一個唯一的 ID
	clientID := fmt.Sprintf("%p", conn)
	client := model.NewClient(clientID, conn)

	// 從查詢參數獲取用戶名（如果有）
	userName := r.URL.Query().Get("username")
	if userName != "" {
		client.SetUserName(userName)
	}

	// 從查詢參數獲取聊天室 ID（如果有）
	roomID := r.URL.Query().Get("roomId")
	if roomID != "" {
		client.SetRoomID(roomID)
	}

	// 將客戶端添加到服務
	err = h.broadcastService.AddClient(client)
	if err != nil {
		h.logger.Error("Failed to add client: %v", err)
		conn.Close()
		return
	}

	h.logger.Info("New client connected: %s, Room: %s", clientID, roomID)

	// 確保在連接關閉時清理資源
	defer func() {
		h.logger.Info("Client disconnected: %s", clientID)

		// 如果客戶端在聊天室中，發送離開通知
		if client.RoomID != "" {
			systemMsg := fmt.Sprintf("使用者 %s 已離開聊天室", client.UserName)
			h.broadcastService.BroadcastToRoom(client.RoomID, []byte(systemMsg))
		}

		h.broadcastService.RemoveClient(clientID)
		conn.Close()
	}()

	// 啟動 ping 發送器
	go h.startPingSender(conn)

	// 處理接收到的訊息
	h.handleMessages(conn, client)
}

// 啟動 ping 發送器
func (h *WebSocketHandler) startPingSender(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			return
		}
	}
}

// 處理接收到的訊息
func (h *WebSocketHandler) handleMessages(conn *websocket.Conn, client *model.Client) {
	for {
		// 讀取 WebSocket 訊息
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("Read error: %v", err)
			}
			break
		}

		// 更新客戶端活躍狀態
		client.UpdateActivity()

		// 根據訊息類型處理
		switch messageType {
		case websocket.TextMessage:
			h.processTextMessage(client, msg)
		case websocket.BinaryMessage:
			h.logger.Info("Received binary message from %s, ignoring", client.ID)
		}
	}
}

// 處理文本訊息
func (h *WebSocketHandler) processTextMessage(client *model.Client, msg []byte) {
	h.logger.Info("Received from %s: %s", client.ID, string(msg))

	// 嘗試解析為 JSON 格式
	var payload MessagePayload
	if err := json.Unmarshal(msg, &payload); err == nil {
		// 成功解析為 JSON
		switch payload.Type {
		case "private":
			if payload.Target != "" {
				h.handlePrivateMessage(client, payload)
				return
			}
		case "join_room":
			if payload.Target != "" {
				h.handleJoinRoom(client, payload.Target)
				return
			}
		case "leave_room":
			h.handleLeaveRoom(client)
			return
		}
	}

	// 如果客戶端在聊天室中，將訊息廣播到該聊天室
	if client.RoomID != "" {
		err := h.broadcastService.BroadcastToRoom(client.RoomID, msg)
		if err != nil {
			h.logger.Error("Failed to broadcast message to room: %v", err)
		}
	} else {
		// 否則廣播到所有客戶端
		err := h.broadcastService.BroadcastMessage(msg)
		if err != nil {
			h.logger.Error("Failed to broadcast message: %v", err)
		}
	}
}

// 處理私人訊息
func (h *WebSocketHandler) handlePrivateMessage(client *model.Client, payload MessagePayload) {
	// 創建私人訊息
	privateMsg, err := json.Marshal(map[string]interface{}{
		"type":    "private",
		"content": payload.Content,
		"from":    client.UserName,
		"time":    time.Now().Unix(),
	})

	if err != nil {
		h.logger.Error("Failed to marshal private message: %v", err)
		return
	}

	// 發送私人訊息
	err = h.broadcastService.SendPrivateMessage(payload.Target, privateMsg)
	if err != nil {
		h.logger.Error("Failed to send private message: %v", err)
	}
}

// 處理加入聊天室
func (h *WebSocketHandler) handleJoinRoom(client *model.Client, roomID string) {
	// 如果客戶端已經在聊天室中，先離開
	if client.RoomID != "" {
		h.handleLeaveRoom(client)
	}

	// 設置新的聊天室 ID
	client.SetRoomID(roomID)

	// 發送系統訊息通知其他用戶
	systemMsg := fmt.Sprintf("使用者 %s 已加入聊天室", client.UserName)
	h.broadcastService.BroadcastToRoom(roomID, []byte(systemMsg))

	h.logger.Info("Client %s joined room %s", client.ID, roomID)
}

// 處理離開聊天室
func (h *WebSocketHandler) handleLeaveRoom(client *model.Client) {
	if client.RoomID == "" {
		return
	}

	roomID := client.RoomID

	// 發送系統訊息通知其他用戶
	systemMsg := fmt.Sprintf("使用者 %s 已離開聊天室", client.UserName)
	h.broadcastService.BroadcastToRoom(roomID, []byte(systemMsg))

	// 清除聊天室 ID
	client.SetRoomID("")

	h.logger.Info("Client %s left room %s", client.ID, roomID)
}
