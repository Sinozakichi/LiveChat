package middleware

import (
	"livechat/backend/model"
	"livechat/backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserResponse 是用戶的 API 響應格式
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// 全局session存儲 - 在實際應用中應該使用Redis或數據庫
var sessionStore = make(map[string]*model.User)

// SessionMiddleware 創建一個會話中間件
func SessionMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 cookie 中獲取會話 ID
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Next()
			return
		}

		// 從會話存儲中獲取用戶
		user, exists := sessionStore[sessionID]
		if !exists {
			c.Next()
			return
		}

		// 將用戶信息設置到上下文中
		c.Set("user", &UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		})

		c.Next()
	}
}

// SetSession 設置用戶session
func SetSession(sessionID string, user *model.User) {
	sessionStore[sessionID] = user
}

// RemoveSession 移除用戶session
func RemoveSession(sessionID string) {
	delete(sessionStore, sessionID)
}

// AuthRequired 創建一個需要認證的中間件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登入"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminRequired 創建一個需要管理員權限的中間件
func AdminRequired(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登入"})
			c.Abort()
			return
		}

		userResponse, ok := userValue.(*UserResponse)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用戶數據格式錯誤"})
			c.Abort()
			return
		}

		// 獲取完整的用戶信息
		user, err := userService.GetUserByID(userResponse.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取用戶信息失敗"})
			c.Abort()
			return
		}

		// 檢查用戶是否為管理員
		if !userService.IsAdmin(user) {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理員權限"})
			c.Abort()
			return
		}

		c.Next()
	}
}
