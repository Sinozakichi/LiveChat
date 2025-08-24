package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"livechat/backend/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService 是一個模擬的用戶服務
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) RegisterUser(username, email, password string) (*model.User, error) {
	args := m.Called(username, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) LoginUser(username, password string) (*model.User, error) {
	args := m.Called(username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) IsAdmin(user *model.User) bool {
	args := m.Called(user)
	return args.Bool(0)
}

// 設置 Gin 測試環境
func setupUserRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.LoadHTMLGlob("../../frontend/*.html") // 加載 HTML 模板
	return router
}

// 測試創建新的用戶處理器
func TestNewUserHandler(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockUserService)

	// 動作 (Act)
	handler := NewUserHandler(mockService)

	// 斷言 (Assert)
	assert.NotNil(t, handler, "處理器不應為 nil")
}

// 測試註冊用戶 - 成功
func TestRegisterSuccess(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupUserRouter()
	handler.RegisterRoutes(router)

	user := &model.User{
		ID:       "1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	mockService.On("RegisterUser", "testuser", "test@example.com", "Password123").Return(user, nil)

	// 創建請求
	reqBody := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "Password123",
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusCreated, w.Code, "狀態碼應該是 201")

	var response UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "應該能夠解析響應")
	assert.Equal(t, "testuser", response.Username, "用戶名應該匹配")
	assert.Equal(t, "test@example.com", response.Email, "電子郵件應該匹配")
	assert.Equal(t, "user", response.Role, "角色應該匹配")

	mockService.AssertExpectations(t)
}

// 測試註冊用戶 - 無效請求
func TestRegisterInvalidRequest(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupUserRouter()
	handler.RegisterRoutes(router)

	// 創建無效請求 (缺少必要字段)
	reqBody := map[string]string{
		"username": "testuser",
		// 缺少 email 和 password
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusBadRequest, w.Code, "狀態碼應該是 400")
}

// 測試註冊用戶 - 用戶名已存在
func TestRegisterUsernameExists(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupUserRouter()
	handler.RegisterRoutes(router)

	mockService.On("RegisterUser", "existinguser", "test@example.com", "Password123").Return(nil, errors.New("用戶名已被使用"))

	// 創建請求
	reqBody := RegisterRequest{
		Username: "existinguser",
		Email:    "test@example.com",
		Password: "Password123",
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusBadRequest, w.Code, "狀態碼應該是 400")
	mockService.AssertExpectations(t)
}

// 測試登入用戶 - 成功
func TestLoginSuccess(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupUserRouter()
	handler.RegisterRoutes(router)

	user := &model.User{
		ID:       "1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
	}

	mockService.On("LoginUser", "testuser", "Password123").Return(user, nil)

	// 創建請求
	reqBody := LoginRequest{
		Username: "testuser",
		Password: "Password123",
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusOK, w.Code, "狀態碼應該是 200")

	var response UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "應該能夠解析響應")
	assert.Equal(t, "testuser", response.Username, "用戶名應該匹配")
	assert.Equal(t, "test@example.com", response.Email, "電子郵件應該匹配")
	assert.Equal(t, "user", response.Role, "角色應該匹配")

	mockService.AssertExpectations(t)
}

// 測試登入用戶 - 無效憑證
func TestLoginInvalidCredentials(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupUserRouter()
	handler.RegisterRoutes(router)

	mockService.On("LoginUser", "testuser", "wrongpassword").Return(nil, errors.New("用戶名或密碼錯誤"))

	// 創建請求
	reqBody := LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	reqJSON, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "狀態碼應該是 401")
	mockService.AssertExpectations(t)
}

// 測試登出用戶
func TestLogout(t *testing.T) {
	// 安排 (Arrange)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router := setupUserRouter()
	handler.RegisterRoutes(router)

	// 創建請求
	req, _ := http.NewRequest("GET", "/api/logout", nil)
	w := httptest.NewRecorder()

	// 動作 (Act)
	router.ServeHTTP(w, req)

	// 斷言 (Assert)
	assert.Equal(t, http.StatusOK, w.Code, "狀態碼應該是 200")

	// 檢查 cookie 是否被刪除
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}

	assert.NotNil(t, sessionCookie, "應該有 session_id cookie")
	assert.True(t, sessionCookie.MaxAge < 0, "cookie 應該被設置為過期")
}
