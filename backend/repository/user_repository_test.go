package repository

import (
	"livechat/backend/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// 測試創建新的用戶儲存庫
func TestNewUserRepository(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()

	// 動作 (Act)
	repo := NewUserRepository(mockDB)

	// 斷言 (Assert)
	assert.NotNil(t, repo, "儲存庫不應為 nil")
}

// 測試創建用戶
func TestCreateUser(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 設置模擬行為
	mockDB.On("Model", mock.AnythingOfType("*model.User")).Return(mockDB)
	mockDB.On("Where", "username = ?", []interface{}{"testuser"}).Return(mockDB)
	mockDB.On("Where", "email = ?", []interface{}{"test@example.com"}).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockResult)
	mockDB.On("Create", mock.AnythingOfType("*model.User")).Return(mockResult)

	repo := NewUserRepository(mockDB)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// 動作 (Act)
	err := repo.CreateUser(user)

	// 斷言 (Assert)
	assert.NoError(t, err, "創建用戶不應返回錯誤")
	assert.NotEqual(t, "password123", user.Password, "密碼應該被哈希")
}

// 測試創建用戶 - 用戶名已存在
func TestCreateUserUsernameExists(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 設置模擬行為
	mockDB.On("Model", mock.AnythingOfType("*model.User")).Return(mockDB)
	mockDB.On("Where", "username = ?", []interface{}{"existinguser"}).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		count := args.Get(0).(*int64)
		*count = 1
	}).Return(mockResult)

	repo := NewUserRepository(mockDB)
	user := &model.User{
		Username: "existinguser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// 動作 (Act)
	err := repo.CreateUser(user)

	// 斷言 (Assert)
	assert.Error(t, err, "創建已存在用戶名的用戶應返回錯誤")
	assert.Equal(t, ErrUserAlreadyExists, err, "錯誤應為 ErrUserAlreadyExists")
}

// 測試創建用戶 - 電子郵件已存在
func TestCreateUserEmailExists(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 設置模擬行為
	mockDB.On("Model", mock.AnythingOfType("*model.User")).Return(mockDB)
	mockDB.On("Where", "username = ?", []interface{}{"newuser"}).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		count := args.Get(0).(*int64)
		*count = 0 // 用戶名不存在
	}).Return(mockResult)
	mockDB.On("Where", "email = ?", []interface{}{"existing@example.com"}).Return(mockDB)
	mockDB.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		count := args.Get(0).(*int64)
		*count = 1 // 電子郵件存在
	}).Return(mockResult)

	repo := NewUserRepository(mockDB)
	user := &model.User{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	// 動作 (Act)
	err := repo.CreateUser(user)

	// 斷言 (Assert)
	assert.Error(t, err, "創建已存在電子郵件的用戶應返回錯誤")
	assert.Equal(t, ErrEmailAlreadyExists, err, "錯誤應為 ErrEmailAlreadyExists")
}

// 測試根據 ID 獲取用戶
func TestGetUserByID(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	expectedUser := &model.User{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Role:      "user",
	}

	// 設置模擬行為
	mockDB.On("First", mock.AnythingOfType("*model.User"), "id = ?", "1").Run(func(args mock.Arguments) {
		user := args.Get(0).(*model.User)
		*user = *expectedUser
	}).Return(mockResult)

	repo := NewUserRepository(mockDB)

	// 動作 (Act)
	user, err := repo.GetUserByID("1")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取用戶不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, "1", user.ID, "用戶 ID 應該匹配")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
}

// 測試根據用戶名獲取用戶
func TestGetUserByUsername(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	expectedUser := &model.User{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Role:      "user",
	}

	// 設置模擬行為
	mockDB.On("First", mock.AnythingOfType("*model.User"), "username = ?", "testuser").Run(func(args mock.Arguments) {
		user := args.Get(0).(*model.User)
		*user = *expectedUser
	}).Return(mockResult)

	repo := NewUserRepository(mockDB)

	// 動作 (Act)
	user, err := repo.GetUserByUsername("testuser")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取用戶不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
}

// 測試檢查用戶憑證
func TestCheckUserCredentials(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 創建一個哈希密碼
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	expectedUser := &model.User{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Role:      "user",
	}

	// 設置模擬行為
	mockDB.On("First", mock.AnythingOfType("*model.User"), "username = ?", "testuser").Run(func(args mock.Arguments) {
		user := args.Get(0).(*model.User)
		*user = *expectedUser
	}).Return(mockResult)

	repo := NewUserRepository(mockDB)

	// 動作 (Act)
	user, err := repo.CheckUserCredentials("testuser", "password123")

	// 斷言 (Assert)
	assert.NoError(t, err, "檢查有效憑證不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
}

// 測試檢查用戶憑證 - 無效密碼
func TestCheckUserCredentialsInvalidPassword(t *testing.T) {
	// 安排 (Arrange)
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 創建一個哈希密碼
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	expectedUser := &model.User{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Role:      "user",
	}

	// 設置模擬行為
	mockDB.On("First", mock.AnythingOfType("*model.User"), "username = ?", "testuser").Run(func(args mock.Arguments) {
		user := args.Get(0).(*model.User)
		*user = *expectedUser
	}).Return(mockResult)

	repo := NewUserRepository(mockDB)

	// 動作 (Act)
	user, err := repo.CheckUserCredentials("testuser", "wrongpassword")

	// 斷言 (Assert)
	assert.Error(t, err, "檢查無效密碼應返回錯誤")
	assert.Equal(t, ErrInvalidCredentials, err, "錯誤應為 ErrInvalidCredentials")
	assert.Nil(t, user, "用戶應為 nil")
}
