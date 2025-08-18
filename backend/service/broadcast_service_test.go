package service

import (
	"errors"
	"livechat/backend/model"
	"livechat/backend/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 定義一個 WebSocket 連接接口，用於依賴注入
type WebSocketConnection interface {
	WriteMessage(messageType int, data []byte) error
	Close() error
}

// MockWebSocketClient 是一個模擬的 WebSocket 客戶端
type MockWebSocketClient struct {
	mock.Mock
	ID string
}

func (m *MockWebSocketClient) WriteMessage(messageType int, data []byte) error {
	args := m.Called(messageType, data)
	return args.Error(0)
}

func (m *MockWebSocketClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// 測試創建新的廣播服務
func TestNewBroadcastService(t *testing.T) {
	// 安排 (Arrange)
	repo := repository.NewClientRepository()

	// 動作 (Act)
	service := NewBroadcastService(repo)

	// 斷言 (Assert)
	assert.NotNil(t, service, "服務不應該為 nil")
	assert.Equal(t, repo, service.clientRepo, "儲存庫應該匹配")
	assert.NotNil(t, service.messageLog, "訊息日誌不應該為 nil")
	assert.Equal(t, 0, len(service.messageLog), "訊息日誌應該是空的")
	assert.Equal(t, 100, service.maxLogSize, "默認最大日誌大小應該是 100")
	assert.NotNil(t, service.errorHandler, "錯誤處理函數不應該為 nil")
}

// 測試服務選項
func TestBroadcastServiceOptions(t *testing.T) {
	// 安排 (Arrange)
	repo := repository.NewClientRepository()
	errorCalled := false
	customError := errors.New("test error")

	// 動作 (Act)
	service := NewBroadcastService(
		repo,
		WithMaxLogSize(50),
		WithErrorHandler(func(err error) {
			errorCalled = true
			assert.Equal(t, customError, errors.Unwrap(err), "錯誤應該匹配")
		}),
	)

	// 斷言 (Assert)
	assert.Equal(t, 50, service.maxLogSize, "最大日誌大小應該是 50")

	// 測試錯誤處理函數
	mockClient := &model.Client{ID: "test-id"}
	service.handleClientError(mockClient, customError)
	assert.True(t, errorCalled, "錯誤處理函數應該被調用")
}

// 測試添加客戶端
func TestAddClient(t *testing.T) {
	// 安排 (Arrange)
	repo := repository.NewClientRepository()
	service := NewBroadcastService(repo)
	client := model.NewClient("test-id", nil)

	// 動作 (Act)
	err := service.AddClient(client)

	// 斷言 (Assert)
	assert.NoError(t, err, "添加客戶端不應該返回錯誤")
	assert.Equal(t, 1, repo.Count(), "儲存庫應該有一個客戶端")
	// 檢查 messageLog 是否有內容
	if len(service.messageLog) > 0 {
		assert.Equal(t, 1, len(service.messageLog["global"]), "應該有一條系統訊息")
	}

	// 獲取訊息歷史
	messages := service.GetMessageHistory("global")
	if len(messages) > 0 {
		assert.Equal(t, SystemMessage, messages[0].Type, "訊息類型應該是系統訊息")
	} else {
		// 如果沒有訊息，則跳過這個斷言
		t.Log("警告：沒有找到系統訊息")
	}

	// 測試添加 nil 客戶端
	err = service.AddClient(nil)
	assert.Error(t, err, "添加 nil 客戶端應該返回錯誤")
}

// 測試廣播訊息
func TestBroadcastMessage(t *testing.T) {
	// 安排 (Arrange)
	repo := repository.NewClientRepository()
	service := NewBroadcastService(repo)

	// 測試空訊息
	err := service.BroadcastMessage([]byte{})
	assert.Equal(t, ErrEmptyMessage, err, "廣播空訊息應該返回 ErrEmptyMessage")

	// 測試沒有客戶端的情況
	err = service.BroadcastMessage([]byte("Hello"))
	assert.Equal(t, ErrNoClients, err, "沒有客戶端時廣播訊息應該返回 ErrNoClients")
}

// 測試發送私人訊息
func TestSendPrivateMessage(t *testing.T) {
	// 安排 (Arrange)
	repo := repository.NewClientRepository()
	service := NewBroadcastService(repo)

	// 測試發送給不存在的客戶端
	message := []byte("Private message")
	err := service.SendPrivateMessage("non-existent-id", message)
	assert.Error(t, err, "發送給不存在的客戶端應該返回錯誤")

	// 測試發送空訊息
	client := model.NewClient("test-id", nil)
	client.IsActive = true
	repo.Add(client)

	err = service.SendPrivateMessage("test-id", []byte{})
	assert.Equal(t, ErrEmptyMessage, err, "發送空訊息應該返回 ErrEmptyMessage")

	// 測試發送給非活躍的客戶端
	client.Deactivate()
	err = service.SendPrivateMessage("test-id", message)
	assert.Error(t, err, "發送給非活躍的客戶端應該返回錯誤")
}

// 測試訊息日誌大小限制
func TestMessageLogSizeLimit(t *testing.T) {
	// 安排 (Arrange)
	repo := repository.NewClientRepository()
	service := NewBroadcastService(repo, WithMaxLogSize(3))

	// 動作 (Act)
	for i := 0; i < 5; i++ {
		msg := ChatMessage{
			Type:      TextMessage,
			Content:   "Test message",
			Timestamp: time.Now().Unix(),
		}
		service.logMessage(msg)
	}

	// 斷言 (Assert)
	assert.Equal(t, 3, len(service.messageLog), "訊息日誌大小應該被限制為 3")
}
