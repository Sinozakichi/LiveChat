package handler

import (
	"livechat/backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler 處理用戶相關的 HTTP 請求
type UserHandler struct {
	userService service.UserService
}

// RegisterRequest 是註冊請求的格式
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest 是登入請求的格式
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 是用戶的 API 響應格式
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// NewUserHandler 創建一個新的用戶處理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterRoutes 註冊路由
func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/login", h.ShowLoginPage)
	router.GET("/register", h.ShowRegisterPage)
	router.POST("/api/register", h.Register)
	router.POST("/api/login", h.Login)
	router.GET("/api/logout", h.Logout)
	router.GET("/api/user", h.GetCurrentUser)
}

// ShowLoginPage 顯示登入頁面
func (h *UserHandler) ShowLoginPage(c *gin.Context) {
	// 檢查用戶是否已登入
	_, exists := c.Get("user")
	if exists {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

// ShowRegisterPage 顯示註冊頁面
func (h *UserHandler) ShowRegisterPage(c *gin.Context) {
	// 檢查用戶是否已登入
	_, exists := c.Get("user")
	if exists {
		c.Redirect(http.StatusFound, "/")
		return
	}
	c.HTML(http.StatusOK, "register.html", gin.H{})
}

// Register 處理用戶註冊請求
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 註冊用戶
	user, err := h.userService.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 創建會話
	sessionID := uuid.New().String()
	c.SetCookie("session_id", sessionID, 0, "/", "", false, true) // 無過期時間
	c.Set("user", user)

	// 返回用戶信息
	c.JSON(http.StatusCreated, UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	})
}

// Login 處理用戶登入請求
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 驗證用戶
	user, err := h.userService.LoginUser(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用戶名或密碼錯誤"})
		return
	}

	// 創建會話
	sessionID := uuid.New().String()
	c.SetCookie("session_id", sessionID, 0, "/", "", false, true) // 無過期時間
	c.Set("user", user)

	// 返回用戶信息
	c.JSON(http.StatusOK, UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	})
}

// Logout 處理用戶登出請求
func (h *UserHandler) Logout(c *gin.Context) {
	// 刪除會話
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// GetCurrentUser 獲取當前登入用戶
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// 從上下文中獲取用戶
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登入"})
		return
	}

	user, ok := userValue.(*UserResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用戶數據格式錯誤"})
		return
	}

	c.JSON(http.StatusOK, user)
}
