package repository

import (
	"livechat/backend/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 測試創建新的聊天室儲存庫
func TestNewRoomRepository(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()

	// 動作 (Act)
	repo := NewRoomRepository(mockDB)

	// 斷言 (Assert)
	assert.NotNil(t, repo, "儲存庫不應該為 nil")
	assert.Equal(t, mockDB, repo.db, "DB 應該匹配")
}

// 測試獲取聊天室
func TestGetRoom(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	expectedRoom := &model.Room{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      "測試聊天室",
	}

	// 設置模擬行為
	mockDB.On("First", mock.AnythingOfType("*model.Room"), mock.Anything).Run(func(args mock.Arguments) {
		room := args.Get(0).(*model.Room)
		*room = *expectedRoom
	}).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	room, err := repo.GetRoom("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室不應該返回錯誤")
	assert.Equal(t, expectedRoom.ID, room.ID, "聊天室 ID 應該匹配")
	assert.Equal(t, expectedRoom.Name, room.Name, "聊天室名稱應該匹配")
	mockDB.AssertExpectations(t)
}

// 測試獲取所有聊天室
func TestGetAllRooms(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	expectedRooms := []model.Room{
		{ID: "1", CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "聊天室1"},
		{ID: "2", CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: "聊天室2"},
	}

	// 設置模擬行為
	mockDB.On("Find", mock.AnythingOfType("*[]model.Room"), mock.Anything).Run(func(args mock.Arguments) {
		rooms := args.Get(0).(*[]model.Room)
		*rooms = expectedRooms
	}).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	rooms, err := repo.GetAllRooms()

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取所有聊天室不應該返回錯誤")
	assert.Equal(t, 2, len(rooms), "應該有 2 個聊天室")
	assert.Equal(t, "1", rooms[0].ID, "第一個聊天室的 ID 應該匹配")
	assert.Equal(t, "聊天室1", rooms[0].Name, "第一個聊天室的名稱應該匹配")
	mockDB.AssertExpectations(t)
}

// 測試創建聊天室
func TestCreateRoom(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	room := &model.Room{
		Name:        "新聊天室",
		Description: "這是一個新的聊天室",
		IsPublic:    true,
		MaxUsers:    100,
		CreatedBy:   "user-123",
		IsActive:    true,
	}

	// 設置模擬行為
	mockDB.On("Create", mock.AnythingOfType("*model.Room")).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	err := repo.CreateRoom(room)

	// 斷言 (Assert)
	assert.NoError(t, err, "創建聊天室不應該返回錯誤")
	mockDB.AssertExpectations(t)
}

// 測試更新聊天室
func TestUpdateRoom(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	room := &model.Room{
		ID:          "1",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Name:        "更新的聊天室",
		Description: "這是一個更新的聊天室",
	}

	// 設置模擬行為
	mockDB.On("Save", mock.AnythingOfType("*model.Room")).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	err := repo.UpdateRoom(room)

	// 斷言 (Assert)
	assert.NoError(t, err, "更新聊天室不應該返回錯誤")
	mockDB.AssertExpectations(t)
}

// 測試獲取聊天室用戶
func TestGetRoomUsers(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	expectedUsers := []model.RoomUser{
		{RoomID: "1", UserID: "user-1", IsActive: true},
		{RoomID: "1", UserID: "user-2", IsActive: true},
	}

	// 設置模擬行為
	mockDB.On("Where", "room_id = ? AND is_active = ?", []interface{}{"1", true}).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]model.RoomUser")).Run(func(args mock.Arguments) {
		users := args.Get(0).(*[]model.RoomUser)
		*users = expectedUsers
	}).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	users, err := repo.GetRoomUsers("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室用戶不應該返回錯誤")
	assert.Equal(t, 2, len(users), "應該有 2 個用戶")
	assert.Equal(t, "user-1", users[0].UserID, "第一個用戶的 ID 應該匹配")
	mockDB.AssertExpectations(t)
}

// 測試用戶加入聊天室
func TestJoinRoom(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 設置模擬行為
	mockDB.On("Create", mock.AnythingOfType("*model.RoomUser")).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	err := repo.JoinRoom("1", "user-123", "member")

	// 斷言 (Assert)
	assert.NoError(t, err, "加入聊天室不應該返回錯誤")
	mockDB.AssertExpectations(t)
}

// 測試用戶離開聊天室
func TestLeaveRoom(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockFirstResult := new(MockGormDB)
	mockFirstResult.Err = nil
	mockSaveResult := new(MockGormDB)
	mockSaveResult.Err = nil

	expectedRoomUser := &model.RoomUser{
		RoomID:   "1",
		UserID:   "user-123",
		IsActive: true,
	}

	// 設置模擬行為
	mockDB.On("Where", "room_id = ? AND user_id = ? AND is_active = ?", []interface{}{"1", "user-123", true}).Return(mockDB)
	mockDB.On("First", mock.AnythingOfType("*model.RoomUser")).Run(func(args mock.Arguments) {
		user := args.Get(0).(*model.RoomUser)
		*user = *expectedRoomUser
	}).Return(mockFirstResult)
	mockDB.On("Save", mock.AnythingOfType("*model.RoomUser")).Return(mockSaveResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	err := repo.LeaveRoom("1", "user-123")

	// 斷言 (Assert)
	assert.NoError(t, err, "離開聊天室不應該返回錯誤")
	mockDB.AssertExpectations(t)
}

// 測試獲取聊天室訊息
func TestGetRoomMessages(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	expectedMessages := []model.Message{
		{RoomID: "1", UserID: "user-1", Content: "訊息1"},
		{RoomID: "1", UserID: "user-2", Content: "訊息2"},
	}

	// 設置模擬行為
	mockDB.On("Where", "room_id = ?", []interface{}{"1"}).Return(mockDB)
	mockDB.On("Order", "created_at desc").Return(mockDB)
	mockDB.On("Limit", 50).Return(mockDB)
	mockDB.On("Find", mock.AnythingOfType("*[]model.Message")).Run(func(args mock.Arguments) {
		messages := args.Get(0).(*[]model.Message)
		*messages = expectedMessages
	}).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	messages, err := repo.GetRoomMessages("1", 50)

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室訊息不應該返回錯誤")
	assert.Equal(t, 2, len(messages), "應該有 2 條訊息")
	assert.Equal(t, "訊息1", messages[0].Content, "第一條訊息的內容應該匹配")
	mockDB.AssertExpectations(t)
}

// 測試保存訊息
func TestSaveMessage(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	message := &model.Message{
		RoomID:  "1",
		UserID:  "user-123",
		Content: "Hello, World!",
	}

	// 設置模擬行為
	mockDB.On("Create", mock.AnythingOfType("*model.Message")).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	err := repo.SaveMessage(message)

	// 斷言 (Assert)
	assert.NoError(t, err, "保存訊息不應該返回錯誤")
	mockDB.AssertExpectations(t)
}

// 測試計算活躍用戶數
func TestCountActiveUsers(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockWhereResult := mockDB
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 設置模擬行為
	mockDB.On("Where", "room_id = ? AND is_active = ?", []interface{}{"1", true}).Return(mockWhereResult)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		count := args.Get(0).(*int64)
		*count = 5
	}).Return(mockResult)

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	count, err := repo.CountActiveUsers("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "計算活躍用戶數不應該返回錯誤")
	assert.Equal(t, int64(5), count, "活躍用戶數應該匹配")
	mockDB.AssertExpectations(t)
}
