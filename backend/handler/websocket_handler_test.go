package handler

import (
	"encoding/json"
	"livechat/backend/model"
	"livechat/backend/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger 是模擬的日誌記錄器，用於測試 WebSocket 處理器的日誌記錄功能
//
// 設計目標：
// 1. 驗證重要事件是否被正確記錄
// 2. 測試錯誤情況的日誌記錄
// 3. 確保不會遺漏關鍵的系統資訊
// 4. 支援日誌級別的測試（Info vs Error）
type MockLogger struct {
	mock.Mock
}

// Info 模擬資訊級別的日誌記錄
// 用於測試正常操作的日誌記錄，如使用者連線、訊息傳送等
func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

// Error 模擬錯誤級別的日誌記錄
// 用於測試異常情況的日誌記錄，如連線失敗、訊息處理錯誤等
func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

// MockBroadcastService 是模擬的廣播服務，用於測試 WebSocket 處理器的核心功能
//
// 設計目標：
// 1. 模擬所有廣播相關的操作，避免對真實廣播服務的依賴
// 2. 驗證 WebSocket 處理器是否正確調用廣播服務的方法
// 3. 測試不同的廣播場景：全域廣播、房間廣播、私人訊息
// 4. 支援錯誤情況的模擬，測試錯誤處理邏輯
type MockBroadcastService struct {
	mock.Mock
}

// AddClient 模擬添加客戶端到廣播服務
// 測試場景：使用者連線時的客戶端註冊
func (m *MockBroadcastService) AddClient(client *model.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

// RemoveClient 模擬從廣播服務中移除客戶端
// 測試場景：使用者斷線時的客戶端清理
func (m *MockBroadcastService) RemoveClient(clientID string) error {
	args := m.Called(clientID)
	return args.Error(0)
}

// BroadcastMessage 模擬全域訊息廣播
// 測試場景：向所有連線的使用者發送系統訊息或公告
func (m *MockBroadcastService) BroadcastMessage(message []byte) error {
	args := m.Called(message)
	return args.Error(0)
}

// BroadcastToRoom 模擬房間內訊息廣播
// 測試場景：向特定聊天室內的所有使用者發送訊息
func (m *MockBroadcastService) BroadcastToRoom(roomID string, message []byte) error {
	args := m.Called(roomID, message)
	return args.Error(0)
}

// SendPrivateMessage 模擬私人訊息發送
// 測試場景：點對點的私人訊息傳送
func (m *MockBroadcastService) SendPrivateMessage(targetID string, message []byte) error {
	args := m.Called(targetID, message)
	return args.Error(0)
}

// GetClient 模擬客戶端查詢
// 測試場景：根據客戶端 ID 查找特定的連線客戶端
func (m *MockBroadcastService) GetClient(clientID string) (*model.Client, error) {
	args := m.Called(clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Client), args.Error(1)
}

// GetMessageHistory 模擬訊息歷史記錄查詢
// 測試場景：獲取房間或全域的歷史訊息
func (m *MockBroadcastService) GetMessageHistory(roomID string) []service.ChatMessage {
	args := m.Called(roomID)
	return args.Get(0).([]service.ChatMessage)
}

// GetClientsInRoom 模擬房間內客戶端查詢
// 測試場景：獲取特定房間內的所有連線使用者
func (m *MockBroadcastService) GetClientsInRoom(roomID string) []*model.Client {
	args := m.Called(roomID)
	return args.Get(0).([]*model.Client)
}

// TestNewWebSocketHandler 測試 WebSocket 處理器的建構子
//
// 測試目標：
// 1. 驗證 WebSocket 處理器能正確初始化
// 2. 確保所有必要的組件都被正確設定
// 3. 檢查依賴注入的正確性
// 4. 驗證預設配置的合理性
//
// 測試策略：
// - 使用最小化的依賴（只有 BroadcastService）
// - 檢查所有內部組件的初始化狀態
// - 驗證預設配置是否適合基本使用場景
// - 確保建構子不會回傳 nil 或未初始化的組件
func TestNewWebSocketHandler(t *testing.T) {
	// 安排 (Arrange)：準備最基本的依賴
	mockBroadcastService := new(MockBroadcastService)

	// 動作 (Act)：創建 WebSocket 處理器實例
	handler := NewWebSocketHandler(mockBroadcastService)

	// 斷言 (Assert)：驗證所有組件都正確初始化
	assert.NotNil(t, handler, "處理器不應該為 nil")
	assert.NotNil(t, handler.upgrader, "WebSocket 升級器不應該為 nil")
	assert.NotNil(t, handler.broadcastService, "廣播服務不應該為 nil")
	assert.NotNil(t, handler.logger, "日誌記錄器不應該為 nil")
}

// TestWebSocketHandlerOptions 測試 WebSocket 處理器的選項配置功能
//
// 測試目標：
// 1. 驗證選項模式（Options Pattern）的正確實現
// 2. 測試自訂日誌記錄器的設定
// 3. 驗證 CORS 檢查函數的自訂配置
// 4. 確保選項能正確覆蓋預設配置
//
// 測試策略：
// - 測試多個選項的同時配置
// - 驗證自訂的 CheckOrigin 函數邏輯
// - 檢查依賴替換的正確性
// - 測試安全性相關的配置（CORS）
func TestWebSocketHandlerOptions(t *testing.T) {
	// 安排 (Arrange)：準備自訂配置的測試環境
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	// 設定日誌記錄器的模擬行為，允許任何日誌調用
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	// 定義自訂的 CORS 檢查函數：只允許來自 example.com 的請求
	customCheckOrigin := func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://example.com"
	}

	// 動作 (Act)：使用選項模式創建處理器
	handler := NewWebSocketHandler(
		mockBroadcastService,
		WithLogger(mockLogger),             // 設定自訂日誌記錄器
		WithCheckOrigin(customCheckOrigin), // 設定自訂 CORS 檢查
	)

	// 斷言 (Assert)：驗證選項配置的正確性
	assert.Equal(t, mockLogger, handler.logger, "日誌記錄器應該匹配自訂的實例")

	// 測試自訂 CheckOrigin 函數的安全性邏輯
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	assert.True(t, handler.upgrader.CheckOrigin(req), "應該允許來自 example.com 的請求")

	// 測試 CORS 安全性：拒絕非授權的來源
	req.Header.Set("Origin", "https://other.com")
	assert.False(t, handler.upgrader.CheckOrigin(req), "不應該允許來自 other.com 的請求")
}

// TestProcessTextMessage 測試 WebSocket 文本訊息處理的核心邏輯
//
// 測試目標：
// 1. 驗證不同類型訊息的正確路由和處理
// 2. 測試全域廣播、房間廣播、私人訊息的分發邏輯
// 3. 確保訊息格式解析的正確性
// 4. 驗證客戶端狀態對訊息處理的影響
//
// 測試策略：
// - 測試多種訊息類型：純文本、JSON 格式、特殊命令
// - 驗證客戶端房間狀態對訊息路由的影響
// - 檢查廣播服務方法的正確調用
// - 測試複雜的訊息處理場景
func TestProcessTextMessage(t *testing.T) {
	// 安排 (Arrange)：準備訊息處理的測試環境
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	// 測試場景 1：沒有加入聊天室的客戶端（全域廣播）
	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
		RoomID:   "", // 空的房間 ID 表示未加入任何聊天室
	}

	// 測試簡單文本訊息的全域廣播
	message := []byte("Hello, World!")
	mockBroadcastService.On("BroadcastMessage", message).Return(nil)

	// 動作 (Act)：處理全域廣播訊息
	handler.processTextMessage(client, message)

	// 斷言 (Assert)：驗證全域廣播被正確調用
	mockBroadcastService.AssertCalled(t, "BroadcastMessage", message)

	// 測試場景 2：已加入聊天室的客戶端（房間廣播）
	clientWithRoom := &model.Client{
		ID:       "test-id-2",
		UserName: "TestUser2",
		RoomID:   "room-1", // 已加入房間 room-1
	}

	// 測試房間內訊息廣播
	roomMessage := []byte("Hello, Room!")
	mockBroadcastService.On("BroadcastToRoom", "room-1", roomMessage).Return(nil)

	// 動作 (Act)：處理房間內廣播訊息
	handler.processTextMessage(clientWithRoom, roomMessage)

	// 斷言 (Assert)：驗證房間廣播被正確調用
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-1", roomMessage)

	// 測試場景 3：JSON 格式的私人訊息
	privatePayload := MessagePayload{
		Type:    "private",
		Content: "Private message",
		Target:  "target-id",
	}
	privateMessage, _ := json.Marshal(privatePayload)

	// 設定私人訊息的模擬行為
	mockBroadcastService.On("SendPrivateMessage", "target-id", mock.Anything).Return(nil)

	// 動作 (Act)：處理私人訊息
	handler.processTextMessage(client, privateMessage)

	// 斷言 (Assert)：驗證私人訊息發送被正確調用
	mockBroadcastService.AssertCalled(t, "SendPrivateMessage", "target-id", mock.Anything)

	// 測試場景 4：加入聊天室的命令訊息
	joinRoomPayload := MessagePayload{
		Type:   "join_room",
		Target: "room-2",
	}
	joinRoomMessage, _ := json.Marshal(joinRoomPayload)

	// 設定加入聊天室的模擬行為
	mockBroadcastService.On("BroadcastToRoom", "room-2", mock.Anything).Return(nil)

	// 動作 (Act)：處理加入聊天室命令
	handler.processTextMessage(client, joinRoomMessage)

	// 斷言 (Assert)：驗證客戶端狀態更新和房間通知
	assert.Equal(t, "room-2", client.RoomID, "客戶端應該加入聊天室 room-2")
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-2", mock.Anything)
}

// TestHandlePrivateMessage 測試私人訊息處理的專門邏輯
//
// 測試目標：
// 1. 驗證私人訊息的完整處理流程
// 2. 測試訊息格式的正確構建和序列化
// 3. 確保發送者資訊被正確添加
// 4. 驗證時間戳的自動生成
//
// 測試策略：
// - 測試私人訊息的端到端處理
// - 驗證訊息 JSON 格式的正確性
// - 檢查所有必要欄位的存在和正確性
// - 確保 Mock 服務被正確調用
func TestHandlePrivateMessage(t *testing.T) {
	// 安排 (Arrange)：準備私人訊息處理的測試環境
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	// 創建發送私人訊息的客戶端
	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
	}

	// 構建私人訊息的載荷
	payload := MessagePayload{
		Type:    "private",
		Content: "Private message",
		Target:  "target-id", // 目標客戶端 ID
	}

	// 設定廣播服務的模擬行為
	mockBroadcastService.On("SendPrivateMessage", "target-id", mock.Anything).Return(nil)

	// 動作 (Act)：處理私人訊息
	handler.handlePrivateMessage(client, payload)

	// 斷言 (Assert)：驗證私人訊息發送被正確調用
	mockBroadcastService.AssertCalled(t, "SendPrivateMessage", "target-id", mock.Anything)

	// 驗證發送的訊息格式的完整性和正確性
	call := mockBroadcastService.Calls[len(mockBroadcastService.Calls)-1]
	sentMessage := call.Arguments.Get(1).([]byte)

	// 解析發送的 JSON 訊息
	var parsedMessage map[string]interface{}
	err := json.Unmarshal(sentMessage, &parsedMessage)
	assert.NoError(t, err, "應該能夠解析訊息")

	// 驗證訊息的各個重要欄位
	assert.Equal(t, "private", parsedMessage["type"], "訊息類型應該是 private")
	assert.Equal(t, "Private message", parsedMessage["content"], "訊息內容應該匹配")
	assert.Equal(t, "TestUser", parsedMessage["from"], "發送者應該是 TestUser")
	assert.NotNil(t, parsedMessage["time"], "時間戳應該存在")
}

// TestHandleJoinRoom 測試客戶端加入聊天室的處理邏輯
//
// 測試目標：
// 1. 驗證客戶端能正確加入指定的聊天室
// 2. 測試客戶端狀態的正確更新
// 3. 確保加入聊天室的通知被正確廣播
// 4. 驗證房間管理的基本功能
//
// 測試策略：
// - 測試從無房間狀態到加入房間的轉換
// - 驗證客戶端的 RoomID 屬性被正確設定
// - 檢查房間廣播通知的發送
// - 確保 Mock 服務的方法被正確調用
func TestHandleJoinRoom(t *testing.T) {
	// 安排 (Arrange)：準備加入聊天室的測試環境
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	// 創建尚未加入任何聊天室的客戶端
	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
		RoomID:   "", // 初始狀態：未加入任何聊天室
	}

	// 設定房間廣播的模擬行為
	mockBroadcastService.On("BroadcastToRoom", "room-1", mock.Anything).Return(nil)

	// 動作 (Act)：執行加入聊天室操作
	handler.handleJoinRoom(client, "room-1")

	// 斷言 (Assert)：驗證加入聊天室的結果
	assert.Equal(t, "room-1", client.RoomID, "客戶端應該加入聊天室 room-1")
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-1", mock.Anything)
}

// TestHandleLeaveRoom 測試客戶端離開聊天室的處理邏輯
//
// 測試目標：
// 1. 驗證客戶端能正確離開當前聊天室
// 2. 測試客戶端狀態的正確清理
// 3. 確保離開聊天室的通知被正確廣播
// 4. 驗證房間狀態管理的完整性
//
// 測試策略：
// - 測試從已加入房間狀態到離開房間的轉換
// - 驗證客戶端的 RoomID 被正確清空
// - 檢查離開房間通知的發送
// - 確保通知在狀態清理前發送到正確的房間
func TestHandleLeaveRoom(t *testing.T) {
	// 安排 (Arrange)：準備離開聊天室的測試環境
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	// 創建已加入聊天室的客戶端
	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
		RoomID:   "room-1", // 初始狀態：已在聊天室 room-1 中
	}

	// 設定房間廣播的模擬行為
	mockBroadcastService.On("BroadcastToRoom", "room-1", mock.Anything).Return(nil)

	// 動作 (Act)：執行離開聊天室操作
	handler.handleLeaveRoom(client)

	// 斷言 (Assert)：驗證離開聊天室的結果
	assert.Equal(t, "", client.RoomID, "客戶端應該離開聊天室")
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-1", mock.Anything)
}
