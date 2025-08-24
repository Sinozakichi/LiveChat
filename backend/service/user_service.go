package service

import (
	"errors"
	"livechat/backend/model"
	"livechat/backend/repository"
	"regexp"
)

// 定義錯誤
var (
	ErrInvalidUsername = errors.New("無效的用戶名")
	ErrInvalidEmail    = errors.New("無效的電子郵件格式")
	ErrWeakPassword    = errors.New("密碼必須包含大小寫字母、數字至少其2者，且長度至少為8位")
	ErrUnauthorized    = errors.New("未授權的操作")
)

// UserService 定義用戶服務接口
type UserService interface {
	RegisterUser(username, email, password string) (*model.User, error)
	LoginUser(username, password string) (*model.User, error)
	GetUserByID(id string) (*model.User, error)
	IsAdmin(user *model.User) bool
}

// UserServiceImpl 實現 UserService 接口
type UserServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService 創建一個新的用戶服務
func NewUserService(userRepo repository.UserRepository) UserService {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

// RegisterUser 註冊新用戶
func (s *UserServiceImpl) RegisterUser(username, email, password string) (*model.User, error) {
	// 驗證用戶名
	if len(username) < 3 || len(username) > 20 {
		return nil, ErrInvalidUsername
	}

	// 驗證電子郵件
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	// 驗證密碼強度
	if !isStrongPassword(password) {
		return nil, ErrWeakPassword
	}

	// 創建用戶
	user := &model.User{
		Username: username,
		Email:    email,
		Password: password,
		Role:     "user", // 默認角色為普通用戶
	}

	// 保存用戶
	err := s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// LoginUser 用戶登入
func (s *UserServiceImpl) LoginUser(username, password string) (*model.User, error) {
	return s.userRepo.CheckUserCredentials(username, password)
}

// GetUserByID 根據 ID 獲取用戶
func (s *UserServiceImpl) GetUserByID(id string) (*model.User, error) {
	return s.userRepo.GetUserByID(id)
}

// IsAdmin 檢查用戶是否為管理員
func (s *UserServiceImpl) IsAdmin(user *model.User) bool {
	return user != nil && user.Role == "admin"
}

// isStrongPassword 檢查密碼是否足夠強
func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		if 'A' <= char && char <= 'Z' {
			hasUpper = true
		} else if 'a' <= char && char <= 'z' {
			hasLower = true
		} else if '0' <= char && char <= '9' {
			hasDigit = true
		}
	}

	// 至少滿足兩個條件
	return (hasUpper && hasLower) || (hasUpper && hasDigit) || (hasLower && hasDigit)
}
