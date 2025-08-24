package service

import (
	"livechat/backend/model"
	"livechat/backend/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoomRepository 是一個模擬的聊天室儲存庫
type MockRoomRepository struct {
	mock.Mock
}

func (m *MockRoomRepository) GetRoom(roomID string) (*model.Room, error) {
	args := m.Called(roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Room), args.Error(1)
}

func (m *MockRoomRepository) GetAllRooms() ([]model.Room, error) {
	args := m.Called()
	return args.Get(0).([]model.Room), args.Error(1)
}

func (m *MockRoomRepository) CreateRoom(room *model.Room) error {
	args := m.Called(room)
	return args.Error(0)
}

func (m *MockRoomRepository) UpdateRoom(room *model.Room) error {
	args := m.Called(room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetRoomUsers(roomID string) ([]model.RoomUser, error) {
	args := m.Called(roomID)
	return args.Get(0).([]model.RoomUser), args.Error(1)
}

func (m *MockRoomRepository) JoinRoom(roomID string, userID string, role string) error {
	args := m.Called(roomID, userID, role)
	return args.Error(0)
}

func (m *MockRoomRepository) LeaveRoom(roomID string, userID string) error {
	args := m.Called(roomID, userID)
	return args.Error(0)
}

func (m *MockRoomRepository) UpdateUserActivity(roomID string, userID string) error {
	args := m.Called(roomID, userID)
	return args.Error(0)
}

func (m *MockRoomRepository) GetRoomMessages(roomID string, limit int) ([]model.Message, error) {
	args := m.Called(roomID, limit)
	return args.Get(0).([]model.Message), args.Error(1)
}

func (m *MockRoomRepository) SaveMessage(message *model.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockRoomRepository) CountActiveUsers(roomID string) (int64, error) {
	args := m.Called(roomID)
	return args.Get(0).(int64), args.Error(1)
}

// 測試創建新的聊天室服務
func TestNewRoomService(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)

	// 動作 (Act)
	service := NewRoomService(mockRepo)

	// 斷言 (Assert)
	assert.NotNil(t, service, "服務不應該為 nil")
	assert.Equal(t, mockRepo, service.roomRepo, "儲存庫應該匹配")
}

// 測試獲取聊天室
func TestGetRoom(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)
	expectedRoom := &model.Room{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "測試聊天室",
	}

	mockRepo.On("GetRoom", "1").Return(expectedRoom, nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	room, err := service.GetRoom("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室不應該返回錯誤")
	assert.Equal(t, expectedRoom, room, "聊天室應該匹配")
	mockRepo.AssertExpectations(t)
}

// 測試獲取所有聊天室
func TestGetAllRooms(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)
	expectedRooms := []model.Room{
		{ID: "1", CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "聊天室1"},
		{ID: "2", CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "聊天室2"},
	}

	mockRepo.On("GetAllRooms").Return(expectedRooms, nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	rooms, err := service.GetAllRooms()

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取所有聊天室不應該返回錯誤")
	assert.Equal(t, expectedRooms, rooms, "聊天室列表應該匹配")
	mockRepo.AssertExpectations(t)
}

// 測試創建聊天室
func TestCreateRoom(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)

	roomData := RoomData{
		Name:        "新聊天室",
		Description: "這是一個新的聊天室",
		IsPublic:    true,
		MaxUsers:    100,
	}

	mockRepo.On("CreateRoom", mock.AnythingOfType("*model.Room")).Return(nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	room, err := service.CreateRoom(roomData, "user-123")

	// 斷言 (Assert)
	assert.NoError(t, err, "創建聊天室不應該返回錯誤")
	assert.Equal(t, roomData.Name, room.Name, "聊天室名稱應該匹配")
	assert.Equal(t, roomData.Description, room.Description, "聊天室描述應該匹配")
	assert.Equal(t, roomData.IsPublic, room.IsPublic, "聊天室公開狀態應該匹配")
	assert.Equal(t, roomData.MaxUsers, room.MaxUsers, "聊天室最大用戶數應該匹配")
	assert.Equal(t, "user-123", room.CreatedBy, "聊天室創建者應該匹配")
	mockRepo.AssertExpectations(t)
}

// 測試加入聊天室
func TestJoinRoom(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)
	room := &model.Room{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "測試聊天室",
	}

	mockRepo.On("GetRoom", "1").Return(room, nil)
	mockRepo.On("JoinRoom", "1", "user-123", "member").Return(nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	err := service.JoinRoom("1", "user-123", "member")

	// 斷言 (Assert)
	assert.NoError(t, err, "加入聊天室不應該返回錯誤")
	mockRepo.AssertExpectations(t)

	// 測試聊天室不存在的情況
	mockRepo.On("GetRoom", "999").Return(nil, repository.ErrRoomNotFound)

	err = service.JoinRoom("999", "user-123", "member")
	assert.Error(t, err, "加入不存在的聊天室應該返回錯誤")
	assert.Equal(t, repository.ErrRoomNotFound, err, "錯誤應該是 ErrRoomNotFound")
}

// 測試離開聊天室
func TestLeaveRoom(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)

	mockRepo.On("LeaveRoom", "1", "user-123").Return(nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	err := service.LeaveRoom("1", "user-123")

	// 斷言 (Assert)
	assert.NoError(t, err, "離開聊天室不應該返回錯誤")
	mockRepo.AssertExpectations(t)
}

// 測試獲取聊天室訊息
func TestGetRoomMessages(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)
	expectedMessages := []model.Message{
		{RoomID: "1", UserID: "user-1", Content: "訊息1"},
		{RoomID: "1", UserID: "user-2", Content: "訊息2"},
	}

	mockRepo.On("GetRoomMessages", "1", 50).Return(expectedMessages, nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	messages, err := service.GetRoomMessages("1", 50)

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室訊息不應該返回錯誤")
	assert.Equal(t, expectedMessages, messages, "訊息列表應該匹配")
	mockRepo.AssertExpectations(t)
}

// 測試發送訊息
func TestSendMessage(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)
	room := &model.Room{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "測試聊天室",
	}

	mockRepo.On("GetRoom", "1").Return(room, nil)
	mockRepo.On("SaveMessage", mock.AnythingOfType("*model.Message")).Return(nil)
	mockRepo.On("UpdateUserActivity", "1", "user-123").Return(nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	err := service.SendMessage("1", "user-123", "Hello, World!")

	// 斷言 (Assert)
	assert.NoError(t, err, "發送訊息不應該返回錯誤")
	mockRepo.AssertExpectations(t)

	// 測試聊天室不存在的情況
	mockRepo.On("GetRoom", "999").Return(nil, repository.ErrRoomNotFound)

	err = service.SendMessage("999", "user-123", "Hello, World!")
	assert.Error(t, err, "發送訊息到不存在的聊天室應該返回錯誤")
	assert.Equal(t, repository.ErrRoomNotFound, err, "錯誤應該是 ErrRoomNotFound")
}

// 測試發送系統訊息
func TestSendSystemMessage(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)
	room := &model.Room{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "測試聊天室",
	}

	mockRepo.On("GetRoom", "1").Return(room, nil)
	mockRepo.On("SaveMessage", mock.AnythingOfType("*model.Message")).Return(nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	err := service.SendSystemMessage("1", "使用者已加入聊天室")

	// 斷言 (Assert)
	assert.NoError(t, err, "發送系統訊息不應該返回錯誤")
	mockRepo.AssertExpectations(t)
}

// 測試獲取聊天室活躍用戶數
func TestGetRoomActiveUserCount(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)

	mockRepo.On("CountActiveUsers", "1").Return(int64(5), nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	count, err := service.GetRoomActiveUserCount("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室活躍用戶數不應該返回錯誤")
	assert.Equal(t, int64(5), count, "活躍用戶數應該匹配")
	mockRepo.AssertExpectations(t)
}

// 測試獲取聊天室用戶
func TestGetRoomUsers(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockRoomRepository)
	expectedUsers := []model.RoomUser{
		{RoomID: "1", UserID: "user-1", IsActive: true},
		{RoomID: "1", UserID: "user-2", IsActive: true},
	}

	mockRepo.On("GetRoomUsers", "1").Return(expectedUsers, nil)

	service := NewRoomService(mockRepo)

	// 動作 (Act)
	users, err := service.GetRoomUsers("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室用戶不應該返回錯誤")
	assert.Equal(t, expectedUsers, users, "用戶列表應該匹配")
	mockRepo.AssertExpectations(t)
}
