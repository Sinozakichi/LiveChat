package handler

import (
	"bytes"
	"encoding/json"
	"livechat/backend/model"
	"livechat/backend/service"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoomService 是一個模擬的聊天室服務
type MockRoomService struct {
	mock.Mock
}

func (m *MockRoomService) GetRoom(roomID string) (*model.Room, error) {
	args := m.Called(roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Room), args.Error(1)
}

func (m *MockRoomService) GetAllRooms() ([]model.Room, error) {
	args := m.Called()
	return args.Get(0).([]model.Room), args.Error(1)
}

func (m *MockRoomService) CreateRoom(data service.RoomData, createdBy string) (*model.Room, error) {
	args := m.Called(data, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Room), args.Error(1)
}

func (m *MockRoomService) JoinRoom(roomID string, userID string, role string) error {
	args := m.Called(roomID, userID, role)
	return args.Error(0)
}

func (m *MockRoomService) LeaveRoom(roomID string, userID string) error {
	args := m.Called(roomID, userID)
	return args.Error(0)
}

func (m *MockRoomService) GetRoomMessages(roomID string, limit int) ([]model.Message, error) {
	args := m.Called(roomID, limit)
	return args.Get(0).([]model.Message), args.Error(1)
}

func (m *MockRoomService) SendMessage(roomID string, userID string, content string) error {
	args := m.Called(roomID, userID, content)
	return args.Error(0)
}

func (m *MockRoomService) SendSystemMessage(roomID string, content string) error {
	args := m.Called(roomID, content)
	return args.Error(0)
}

func (m *MockRoomService) GetRoomActiveUserCount(roomID string) (int64, error) {
	args := m.Called(roomID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoomService) GetRoomUsers(roomID string) ([]model.RoomUser, error) {
	args := m.Called(roomID)
	return args.Get(0).([]model.RoomUser), args.Error(1)
}

// 設置 Gin 測試環境
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// 測試獲取所有聊天室
func TestGetAllRooms(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockRoomService)
	handler := NewRoomHandler(mockService)
	router := setupRouter()
	handler.RegisterRoutes(router)

	// 模擬數據
	rooms := []model.Room{
		{ID: "1", Name: "聊天室1", Description: "描述1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "2", Name: "聊天室2", Description: "描述2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	// 設置模擬行為
	mockService.On("GetAllRooms").Return(rooms, nil)
	mockService.On("GetRoomActiveUserCount", "1").Return(int64(5), nil)
	mockService.On("GetRoomActiveUserCount", "2").Return(int64(3), nil)

	// 創建請求
	req, _ := http.NewRequest("GET", "/api/rooms", nil)
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusOK, w.Code, "狀態碼應該是 200")

	var response []RoomResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "應該能夠解析響應")
	assert.Equal(t, 2, len(response), "應該有 2 個聊天室")
	assert.Equal(t, "聊天室1", response[0].Name, "第一個聊天室的名稱應該匹配")
	assert.Equal(t, int64(5), response[0].ActiveUsers, "第一個聊天室的活躍用戶數應該匹配")

	mockService.AssertExpectations(t)
}

// 測試獲取特定聊天室
func TestGetRoom(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockRoomService)
	handler := NewRoomHandler(mockService)
	router := setupRouter()
	handler.RegisterRoutes(router)

	// 模擬數據
	room := &model.Room{
		ID:          "1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Name:        "測試聊天室",
		Description: "這是一個測試聊天室",
		IsPublic:    true,
		MaxUsers:    100,
		CreatedBy:   "user-123",
	}

	// 設置模擬行為
	mockService.On("GetRoom", "1").Return(room, nil)
	mockService.On("GetRoomActiveUserCount", "1").Return(int64(5), nil)

	// 創建請求
	req, _ := http.NewRequest("GET", "/api/rooms/1", nil)
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusOK, w.Code, "狀態碼應該是 200")

	var response RoomResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "應該能夠解析響應")
	assert.Equal(t, "測試聊天室", response.Name, "聊天室名稱應該匹配")
	assert.Equal(t, int64(5), response.ActiveUsers, "活躍用戶數應該匹配")

	mockService.AssertExpectations(t)
}

// 測試創建聊天室
func TestCreateRoom(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockRoomService)
	handler := NewRoomHandler(mockService)
	router := setupRouter()
	handler.RegisterRoutes(router)

	// 模擬數據
	request := CreateRoomRequest{
		Name:        "新聊天室",
		Description: "這是一個新的聊天室",
		IsPublic:    true,
		MaxUsers:    100,
	}

	room := &model.Room{
		ID:          "1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Name:        "新聊天室",
		Description: "這是一個新的聊天室",
		IsPublic:    true,
		MaxUsers:    100,
		CreatedBy:   "user-123",
	}

	// 設置模擬行為
	mockService.On("CreateRoom", mock.AnythingOfType("service.RoomData"), "system").Return(room, nil)

	// 創建請求
	requestBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/rooms", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusCreated, w.Code, "狀態碼應該是 201")

	var response RoomResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "應該能夠解析響應")
	assert.Equal(t, "新聊天室", response.Name, "聊天室名稱應該匹配")

	mockService.AssertExpectations(t)
}

// 測試獲取聊天室訊息
func TestGetRoomMessages(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockRoomService)
	handler := NewRoomHandler(mockService)
	router := setupRouter()
	handler.RegisterRoutes(router)

	// 模擬數據
	messages := []model.Message{
		{RoomID: "1", UserID: "user-1", Content: "訊息1"},
		{RoomID: "1", UserID: "user-2", Content: "訊息2"},
	}

	// 設置模擬行為
	mockService.On("GetRoomMessages", "1", 50).Return(messages, nil)

	// 創建請求
	req, _ := http.NewRequest("GET", "/api/rooms/1/messages", nil)
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusOK, w.Code, "狀態碼應該是 200")

	var response []model.Message
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "應該能夠解析響應")
	assert.Equal(t, 2, len(response), "應該有 2 條訊息")
	assert.Equal(t, "訊息1", response[0].Content, "第一條訊息的內容應該匹配")

	mockService.AssertExpectations(t)
}

// 測試獲取聊天室用戶
func TestGetRoomUsers(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockRoomService)
	handler := NewRoomHandler(mockService)
	router := setupRouter()
	handler.RegisterRoutes(router)

	// 模擬數據
	users := []model.RoomUser{
		{RoomID: "1", UserID: "user-1", IsActive: true},
		{RoomID: "1", UserID: "user-2", IsActive: true},
	}

	// 設置模擬行為
	mockService.On("GetRoomUsers", "1").Return(users, nil)

	// 創建請求
	req, _ := http.NewRequest("GET", "/api/rooms/1/users", nil)
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusOK, w.Code, "狀態碼應該是 200")

	var response []model.RoomUser
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "應該能夠解析響應")
	assert.Equal(t, 2, len(response), "應該有 2 個用戶")
	assert.Equal(t, "user-1", response[0].UserID, "第一個用戶的 ID 應該匹配")

	mockService.AssertExpectations(t)
}
