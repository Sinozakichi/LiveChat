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
	// 安排 (Arrange) - 使用帶有完整結構的模擬資料庫
	mockDB := NewMockDBWithSchema()

	expectedRoom := &model.Room{
		ID:          "1",
		Name:        "測試聊天室",
		Description: "這是一個測試聊天室",
		IsPublic:    true,
		MaxUsers:    100,
		CreatedBy:   "system",
		IsActive:    true,
	}

	// 先插入測試資料到真實的記憶體資料庫
	err := mockDB.DB.Create(expectedRoom).Error
	assert.NoError(t, err, "插入測試資料不應該失敗")

	repo := NewRoomRepository(mockDB)

	// 動作 (Act)
	room, err := repo.GetRoom("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取聊天室不應該返回錯誤")
	assert.NotNil(t, room, "返回的聊天室不應該為空")
	assert.Equal(t, expectedRoom.ID, room.ID, "聊天室 ID 應該匹配")
	assert.Equal(t, expectedRoom.Name, room.Name, "聊天室名稱應該匹配")
}

// TestGetAllRooms 測試獲取所有聊天室功能
//
// 測試目標：
// 1. 驗證能正確查詢所有活躍的聊天室
// 2. 測試 is_active = true 的查詢條件
// 3. 確保回傳的聊天室資料完整性
// 4. 驗證 rooms 表的查詢邏輯
//
// 測試策略：
// - 使用帶有完整資料庫結構的 MockDB
// - 預先插入測試聊天室資料
// - 測試真實的 SQL 查詢操作（包含活躍狀態過濾）
// - 驗證查詢結果的正確性
func TestGetAllRooms(t *testing.T) {
	// 安排 (Arrange)：使用完整資料庫結構的 MockDB
	mockDB := NewMockDBWithSchema()
	repo := NewRoomRepository(mockDB)

	// 準備測試資料：插入測試聊天室
	testRooms := []model.Room{
		{
			ID:          "room-1",
			Name:        "公共聊天室",
			Description: "歡迎大家",
			IsActive:    true,
		},
		{
			ID:          "room-2",
			Name:        "技術討論",
			Description: "技術交流專區",
			IsActive:    true,
		},
		{
			ID:          "room-3",
			Name:        "已停用聊天室",
			Description: "這個應該被過濾掉",
			IsActive:    false, // 這個聊天室應該被過濾掉
		},
	}

	// 將測試資料插入到真實的記憶體資料庫中
	for _, room := range testRooms {
		if !room.IsActive {
			// 處理 IsActive=false 的 GORM 零值問題
			err := mockDB.DB.Create(&room).Error
			assert.NoError(t, err, "插入測試聊天室不應該失敗")
			err = mockDB.DB.Model(&room).Where("id = ?", room.ID).Update("is_active", false).Error
			assert.NoError(t, err, "更新聊天室 IsActive 不應該失敗")
		} else {
			err := mockDB.DB.Create(&room).Error
			assert.NoError(t, err, "插入測試聊天室不應該失敗")
		}
	}

	// 動作 (Act)：執行獲取所有聊天室查詢
	rooms, err := repo.GetAllRooms()

	// 斷言 (Assert)：驗證查詢結果
	assert.NoError(t, err, "獲取所有聊天室不應該返回錯誤")
	assert.Equal(t, 2, len(rooms), "應該有 2 個活躍聊天室（過濾掉非活躍聊天室）")

	// 驗證回傳的聊天室資料正確性
	if len(rooms) >= 2 {
		// 確保只回傳活躍聊天室
		for _, room := range rooms {
			assert.True(t, room.IsActive, "所有回傳的聊天室都應該是活躍狀態")
		}

		// 驗證特定聊天室存在
		roomIDs := make([]string, len(rooms))
		for i, room := range rooms {
			roomIDs[i] = room.ID
		}
		assert.Contains(t, roomIDs, "room-1", "應該包含 room-1")
		assert.Contains(t, roomIDs, "room-2", "應該包含 room-2")
		assert.NotContains(t, roomIDs, "room-3", "不應該包含非活躍的 room-3")
	}
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

// TestGetRoomUsers 測試獲取聊天室用戶功能
//
// 測試目標：
// 1. 驗證能正確查詢特定聊天室的活躍使用者
// 2. 測試資料庫查詢條件的正確性
// 3. 確保回傳的使用者資料完整性
// 4. 驗證 room_users 表的查詢邏輯
//
// 測試策略：
// - 使用帶有完整資料庫結構的 MockDB
// - 預先插入測試資料到資料庫
// - 測試真實的 SQL 查詢操作
// - 驗證查詢結果的正確性
func TestGetRoomUsers(t *testing.T) {
	// 安排 (Arrange)：使用完整資料庫結構的 MockDB
	mockDB := NewMockDBWithSchema()
	repo := NewRoomRepository(mockDB)

	// 準備測試資料：在真實資料庫中插入測試記錄
	testUsers := []model.RoomUser{
		{
			RoomID:   "room-1",
			UserID:   "user-1",
			Role:     "member",
			IsActive: true,
		},
		{
			RoomID:   "room-1",
			UserID:   "user-2",
			Role:     "admin",
			IsActive: true,
		},
		{
			RoomID:   "room-1",
			UserID:   "user-3",
			Role:     "member",
			IsActive: false, // 這個使用者應該被過濾掉
		},
	}

	// 將測試資料插入到真實的記憶體資料庫中
	for i, user := range testUsers {
		t.Logf("插入使用者 %d: UserID=%s, IsActive=%t", i, user.UserID, user.IsActive)

		// 對於 IsActive=false 的情況，我們需要特殊處理
		// 因為 GORM 會忽略零值並使用資料庫預設值
		if !user.IsActive {
			// 先插入，然後明確更新 IsActive 欄位
			err := mockDB.DB.Create(&user).Error
			assert.NoError(t, err, "插入測試資料不應該失敗")

			// 明確更新 IsActive 為 false
			err = mockDB.DB.Model(&user).Where("user_id = ? AND room_id = ?", user.UserID, user.RoomID).Update("is_active", false).Error
			assert.NoError(t, err, "更新 IsActive 不應該失敗")
		} else {
			// 對於 IsActive=true 的情況，正常插入即可
			err := mockDB.DB.Create(&user).Error
			assert.NoError(t, err, "插入測試資料不應該失敗")
		}

		// 驗證插入後的資料
		var insertedUser model.RoomUser
		err := mockDB.DB.Where("user_id = ? AND room_id = ?", user.UserID, user.RoomID).First(&insertedUser).Error
		assert.NoError(t, err, "應該能找到剛插入的使用者")
		t.Logf("插入後驗證: UserID=%s, IsActive=%t", insertedUser.UserID, insertedUser.IsActive)
	}

	// 動作 (Act)：執行獲取聊天室使用者查詢
	users, err := repo.GetRoomUsers("room-1")

	// 調試資訊：打印實際查詢結果
	t.Logf("查詢到的使用者數量: %d", len(users))
	for i, user := range users {
		t.Logf("使用者 %d: UserID=%s, IsActive=%t", i, user.UserID, user.IsActive)
	}

	// 斷言 (Assert)：驗證查詢結果
	assert.NoError(t, err, "獲取聊天室用戶不應該返回錯誤")
	assert.Equal(t, 2, len(users), "應該有 2 個活躍用戶（過濾掉非活躍用戶）")

	// 驗證回傳的使用者資料正確性
	if len(users) >= 2 {
		// 確保只回傳活躍使用者
		for _, user := range users {
			assert.True(t, user.IsActive, "所有回傳的使用者都應該是活躍狀態")
			assert.Equal(t, "room-1", user.RoomID, "所有使用者都應該屬於指定聊天室")
		}

		// 驗證特定使用者存在
		userIDs := make([]string, len(users))
		for i, user := range users {
			userIDs[i] = user.UserID
		}
		assert.Contains(t, userIDs, "user-1", "應該包含 user-1")
		assert.Contains(t, userIDs, "user-2", "應該包含 user-2")
		assert.NotContains(t, userIDs, "user-3", "不應該包含非活躍的 user-3")
	}
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
