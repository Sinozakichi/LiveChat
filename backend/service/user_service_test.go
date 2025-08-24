package service

import (
	"livechat/backend/model"
	"livechat/backend/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository 是一個模擬的用戶儲存庫
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) CheckUserCredentials(username, password string) (*model.User, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// 測試創建新的用戶服務
func TestNewUserService(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)

	// 動作 (Act)
	service := NewUserService(mockRepo)

	// 斷言 (Assert)
	assert.NotNil(t, service, "服務不應為 nil")
}

// 測試註冊用戶 - 成功
func TestRegisterUserSuccess(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	mockRepo.On("CreateUser", mock.AnythingOfType("*model.User")).Return(nil)

	service := NewUserService(mockRepo)

	// 動作 (Act)
	user, err := service.RegisterUser("testuser", "test@example.com", "Password123")

	// 斷言 (Assert)
	assert.NoError(t, err, "註冊用戶不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
	assert.Equal(t, "test@example.com", user.Email, "電子郵件應該匹配")
	assert.Equal(t, "user", user.Role, "角色應該為 user")
	mockRepo.AssertExpectations(t)
}

// 測試註冊用戶 - 無效用戶名
func TestRegisterUserInvalidUsername(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// 動作 (Act)
	user, err := service.RegisterUser("ab", "test@example.com", "Password123")

	// 斷言 (Assert)
	assert.Error(t, err, "註冊無效用戶名應返回錯誤")
	assert.Equal(t, ErrInvalidUsername, err, "錯誤應為 ErrInvalidUsername")
	assert.Nil(t, user, "用戶應為 nil")
}

// 測試註冊用戶 - 無效電子郵件
func TestRegisterUserInvalidEmail(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// 動作 (Act)
	user, err := service.RegisterUser("testuser", "invalid-email", "Password123")

	// 斷言 (Assert)
	assert.Error(t, err, "註冊無效電子郵件應返回錯誤")
	assert.Equal(t, ErrInvalidEmail, err, "錯誤應為 ErrInvalidEmail")
	assert.Nil(t, user, "用戶應為 nil")
}

// 測試註冊用戶 - 弱密碼
func TestRegisterUserWeakPassword(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	// 動作 (Act)
	user, err := service.RegisterUser("testuser", "test@example.com", "123456")

	// 斷言 (Assert)
	assert.Error(t, err, "註冊弱密碼應返回錯誤")
	assert.Equal(t, ErrWeakPassword, err, "錯誤應為 ErrWeakPassword")
	assert.Nil(t, user, "用戶應為 nil")
}

// 測試註冊用戶 - 用戶名已存在
func TestRegisterUserUsernameExists(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	mockRepo.On("CreateUser", mock.AnythingOfType("*model.User")).Return(repository.ErrUserAlreadyExists)

	service := NewUserService(mockRepo)

	// 動作 (Act)
	user, err := service.RegisterUser("existinguser", "test@example.com", "Password123")

	// 斷言 (Assert)
	assert.Error(t, err, "註冊已存在用戶名應返回錯誤")
	assert.Equal(t, repository.ErrUserAlreadyExists, err, "錯誤應為 ErrUserAlreadyExists")
	assert.Nil(t, user, "用戶應為 nil")
	mockRepo.AssertExpectations(t)
}

// 測試登入用戶 - 成功
func TestLoginUserSuccess(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	expectedUser := &model.User{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "user",
	}
	mockRepo.On("CheckUserCredentials", "testuser", "Password123").Return(expectedUser, nil)

	service := NewUserService(mockRepo)

	// 動作 (Act)
	user, err := service.LoginUser("testuser", "Password123")

	// 斷言 (Assert)
	assert.NoError(t, err, "登入用戶不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
	mockRepo.AssertExpectations(t)
}

// 測試登入用戶 - 無效憑證
func TestLoginUserInvalidCredentials(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	mockRepo.On("CheckUserCredentials", "testuser", "wrongpassword").Return(nil, repository.ErrInvalidCredentials)

	service := NewUserService(mockRepo)

	// 動作 (Act)
	user, err := service.LoginUser("testuser", "wrongpassword")

	// 斷言 (Assert)
	assert.Error(t, err, "登入無效憑證應返回錯誤")
	assert.Equal(t, repository.ErrInvalidCredentials, err, "錯誤應為 ErrInvalidCredentials")
	assert.Nil(t, user, "用戶應為 nil")
	mockRepo.AssertExpectations(t)
}

// 測試檢查用戶是否為管理員
func TestIsAdmin(t *testing.T) {
	// 安排 (Arrange)
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	adminUser := &model.User{
		ID:       "1",
		Username: "admin",
		Role:     "admin",
	}

	regularUser := &model.User{
		ID:       "2",
		Username: "user",
		Role:     "user",
	}

	// 動作 & 斷言 (Act & Assert)
	assert.True(t, service.IsAdmin(adminUser), "管理員用戶應返回 true")
	assert.False(t, service.IsAdmin(regularUser), "普通用戶應返回 false")
	assert.False(t, service.IsAdmin(nil), "nil 用戶應返回 false")
}
