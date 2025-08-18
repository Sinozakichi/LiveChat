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

// MockLogger 是一個模擬的日誌記錄器
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.Called(msg, args)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.Called(msg, args)
}

// MockBroadcastService 是一個模擬的廣播服務
type MockBroadcastService struct {
	mock.Mock
}

func (m *MockBroadcastService) AddClient(client *model.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockBroadcastService) RemoveClient(clientID string) error {
	args := m.Called(clientID)
	return args.Error(0)
}

func (m *MockBroadcastService) BroadcastMessage(message []byte) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockBroadcastService) BroadcastToRoom(roomID string, message []byte) error {
	args := m.Called(roomID, message)
	return args.Error(0)
}

func (m *MockBroadcastService) SendPrivateMessage(targetID string, message []byte) error {
	args := m.Called(targetID, message)
	return args.Error(0)
}

func (m *MockBroadcastService) GetClient(clientID string) (*model.Client, error) {
	args := m.Called(clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Client), args.Error(1)
}

func (m *MockBroadcastService) GetMessageHistory(roomID string) []service.ChatMessage {
	args := m.Called(roomID)
	return args.Get(0).([]service.ChatMessage)
}

func (m *MockBroadcastService) GetClientsInRoom(roomID string) []*model.Client {
	args := m.Called(roomID)
	return args.Get(0).([]*model.Client)
}

// 測試創建新的 WebSocket 處理器
func TestNewWebSocketHandler(t *testing.T) {
	// 安排 (Arrange)
	mockBroadcastService := new(MockBroadcastService)

	// 動作 (Act)
	handler := NewWebSocketHandler(mockBroadcastService)

	// 斷言 (Assert)
	assert.NotNil(t, handler, "處理器不應該為 nil")
	assert.NotNil(t, handler.upgrader, "升級器不應該為 nil")
	assert.NotNil(t, handler.broadcastService, "廣播服務不應該為 nil")
	assert.NotNil(t, handler.logger, "日誌記錄器不應該為 nil")
}

// 測試處理器選項
func TestWebSocketHandlerOptions(t *testing.T) {
	// 安排 (Arrange)
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	customCheckOrigin := func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://example.com"
	}

	// 動作 (Act)
	handler := NewWebSocketHandler(
		mockBroadcastService,
		WithLogger(mockLogger),
		WithCheckOrigin(customCheckOrigin),
	)

	// 斷言 (Assert)
	assert.Equal(t, mockLogger, handler.logger, "日誌記錄器應該匹配")

	// 測試 CheckOrigin 函數
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")
	assert.True(t, handler.upgrader.CheckOrigin(req), "應該允許來自 example.com 的請求")

	req.Header.Set("Origin", "https://other.com")
	assert.False(t, handler.upgrader.CheckOrigin(req), "不應該允許來自 other.com 的請求")
}

// 測試處理文本訊息
func TestProcessTextMessage(t *testing.T) {
	// 安排 (Arrange)
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	// 測試沒有聊天室的客戶端
	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
		RoomID:   "",
	}

	// 測試廣播訊息
	message := []byte("Hello, World!")
	mockBroadcastService.On("BroadcastMessage", message).Return(nil)

	// 動作 (Act)
	handler.processTextMessage(client, message)

	// 斷言 (Assert)
	mockBroadcastService.AssertCalled(t, "BroadcastMessage", message)

	// 測試有聊天室的客戶端
	clientWithRoom := &model.Client{
		ID:       "test-id-2",
		UserName: "TestUser2",
		RoomID:   "room-1",
	}

	// 測試廣播訊息到聊天室
	roomMessage := []byte("Hello, Room!")
	mockBroadcastService.On("BroadcastToRoom", "room-1", roomMessage).Return(nil)

	// 動作 (Act)
	handler.processTextMessage(clientWithRoom, roomMessage)

	// 斷言 (Assert)
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-1", roomMessage)

	// 測試私人訊息
	privatePayload := MessagePayload{
		Type:    "private",
		Content: "Private message",
		Target:  "target-id",
	}
	privateMessage, _ := json.Marshal(privatePayload)

	// 模擬 SendPrivateMessage 被調用
	mockBroadcastService.On("SendPrivateMessage", "target-id", mock.Anything).Return(nil)

	// 動作 (Act)
	handler.processTextMessage(client, privateMessage)

	// 斷言 (Assert)
	mockBroadcastService.AssertCalled(t, "SendPrivateMessage", "target-id", mock.Anything)

	// 測試加入聊天室
	joinRoomPayload := MessagePayload{
		Type:   "join_room",
		Target: "room-2",
	}
	joinRoomMessage, _ := json.Marshal(joinRoomPayload)

	// 模擬 BroadcastToRoom 被調用
	mockBroadcastService.On("BroadcastToRoom", "room-2", mock.Anything).Return(nil)

	// 動作 (Act)
	handler.processTextMessage(client, joinRoomMessage)

	// 斷言 (Assert)
	assert.Equal(t, "room-2", client.RoomID, "客戶端應該加入聊天室 room-2")
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-2", mock.Anything)
}

// 測試處理私人訊息
func TestHandlePrivateMessage(t *testing.T) {
	// 安排 (Arrange)
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
	}

	payload := MessagePayload{
		Type:    "private",
		Content: "Private message",
		Target:  "target-id",
	}

	// 模擬 SendPrivateMessage 被調用
	mockBroadcastService.On("SendPrivateMessage", "target-id", mock.Anything).Return(nil)

	// 動作 (Act)
	handler.handlePrivateMessage(client, payload)

	// 斷言 (Assert)
	mockBroadcastService.AssertCalled(t, "SendPrivateMessage", "target-id", mock.Anything)

	// 驗證發送的訊息格式
	call := mockBroadcastService.Calls[len(mockBroadcastService.Calls)-1]
	sentMessage := call.Arguments.Get(1).([]byte)

	var parsedMessage map[string]interface{}
	err := json.Unmarshal(sentMessage, &parsedMessage)
	assert.NoError(t, err, "應該能夠解析訊息")
	assert.Equal(t, "private", parsedMessage["type"], "訊息類型應該是 private")
	assert.Equal(t, "Private message", parsedMessage["content"], "訊息內容應該匹配")
	assert.Equal(t, "TestUser", parsedMessage["from"], "發送者應該是 TestUser")
	assert.NotNil(t, parsedMessage["time"], "時間戳應該存在")
}

// 測試處理加入聊天室
func TestHandleJoinRoom(t *testing.T) {
	// 安排 (Arrange)
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
		RoomID:   "",
	}

	// 模擬 BroadcastToRoom 被調用
	mockBroadcastService.On("BroadcastToRoom", "room-1", mock.Anything).Return(nil)

	// 動作 (Act)
	handler.handleJoinRoom(client, "room-1")

	// 斷言 (Assert)
	assert.Equal(t, "room-1", client.RoomID, "客戶端應該加入聊天室 room-1")
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-1", mock.Anything)
}

// 測試處理離開聊天室
func TestHandleLeaveRoom(t *testing.T) {
	// 安排 (Arrange)
	mockBroadcastService := new(MockBroadcastService)
	mockLogger := new(MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	handler := NewWebSocketHandler(mockBroadcastService, WithLogger(mockLogger))

	client := &model.Client{
		ID:       "test-id",
		UserName: "TestUser",
		RoomID:   "room-1",
	}

	// 模擬 BroadcastToRoom 被調用
	mockBroadcastService.On("BroadcastToRoom", "room-1", mock.Anything).Return(nil)

	// 動作 (Act)
	handler.handleLeaveRoom(client)

	// 斷言 (Assert)
	assert.Equal(t, "", client.RoomID, "客戶端應該離開聊天室")
	mockBroadcastService.AssertCalled(t, "BroadcastToRoom", "room-1", mock.Anything)
}
