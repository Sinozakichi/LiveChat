package handler

import (
	"fmt"
	"livechat/backend/model"
	"livechat/backend/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoomService 定義了聊天室服務的接口
type RoomService interface {
	GetRoom(roomID string) (*model.Room, error)
	GetAllRooms() ([]model.Room, error)
	CreateRoom(data service.RoomData, createdBy string) (*model.Room, error)
	JoinRoom(roomID string, userID string, role string) error
	LeaveRoom(roomID string, userID string) error
	GetRoomMessages(roomID string, limit int) ([]model.Message, error)
	SendMessage(roomID string, userID string, content string) error
	SendSystemMessage(roomID string, content string) error
	GetRoomActiveUserCount(roomID string) (int64, error)
	GetRoomUsers(roomID string) ([]model.RoomUser, error)
}

// RoomHandler 處理聊天室相關的 HTTP 請求
type RoomHandler struct {
	roomService RoomService
}

// RoomResponse 是聊天室的 API 響應格式
type RoomResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`
	MaxUsers    int    `json:"maxUsers"`
	CreatedBy   string `json:"createdBy"`
	ActiveUsers int64  `json:"activeUsers"`
}

// CreateRoomRequest 是創建聊天室的請求格式
type CreateRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`
	MaxUsers    int    `json:"maxUsers"`
}

// NewRoomHandler 創建一個新的聊天室處理器
func NewRoomHandler(roomService RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

// RegisterRoutes 註冊聊天室相關的路由
func (h *RoomHandler) RegisterRoutes(router *gin.Engine) {
	rooms := router.Group("/api/rooms")
	{
		rooms.GET("", h.GetAllRooms)
		rooms.GET("/:id", h.GetRoom)
		rooms.POST("", h.CreateRoom)
		rooms.GET("/:id/messages", h.GetRoomMessages)
		rooms.GET("/:id/users", h.GetRoomUsers)
	}
}

// GetAllRooms 獲取所有聊天室
func (h *RoomHandler) GetAllRooms(c *gin.Context) {
	// 獲取所有聊天室
	rooms, err := h.roomService.GetAllRooms()
	if err != nil {
		fmt.Printf("Error getting rooms: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取聊天室失敗"})
		return
	}

	fmt.Printf("Found %d rooms in database\n", len(rooms))
	for i, room := range rooms {
		fmt.Printf("Room %d: ID=%d, Name=%s\n", i+1, room.ID, room.Name)
	}

	// 構建響應
	var response []RoomResponse
	for _, room := range rooms {
		// 獲取活躍用戶數
		activeUsers, err := h.roomService.GetRoomActiveUserCount(strconv.Itoa(int(room.ID)))
		if err != nil {
			activeUsers = 0
		}

		response = append(response, RoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			IsPublic:    room.IsPublic,
			MaxUsers:    room.MaxUsers,
			CreatedBy:   room.CreatedBy,
			ActiveUsers: activeUsers,
		})
	}

	fmt.Printf("Sending %d rooms to frontend\n", len(response))
	c.JSON(http.StatusOK, response)
}

// GetRoom 獲取特定聊天室
func (h *RoomHandler) GetRoom(c *gin.Context) {
	// 獲取聊天室 ID
	roomID := c.Param("id")

	// 獲取聊天室
	room, err := h.roomService.GetRoom(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "聊天室不存在"})
		return
	}

	// 獲取活躍用戶數
	activeUsers, err := h.roomService.GetRoomActiveUserCount(roomID)
	if err != nil {
		activeUsers = 0
	}

	// 構建響應
	response := RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		IsPublic:    room.IsPublic,
		MaxUsers:    room.MaxUsers,
		CreatedBy:   room.CreatedBy,
		ActiveUsers: activeUsers,
	}

	c.JSON(http.StatusOK, response)
}

// CreateRoom 創建一個新的聊天室
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	// 解析請求
	var request CreateRoomRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求"})
		return
	}

	// 獲取用戶 ID
	userID := c.GetString("userID")
	if userID == "" {
		// 如果未設置用戶 ID，使用默認值
		userID = "system"
	}

	// 創建聊天室
	roomData := service.RoomData{
		Name:        request.Name,
		Description: request.Description,
		IsPublic:    request.IsPublic,
		MaxUsers:    request.MaxUsers,
	}

	room, err := h.roomService.CreateRoom(roomData, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建聊天室失敗"})
		return
	}

	// 構建響應
	response := RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		IsPublic:    room.IsPublic,
		MaxUsers:    room.MaxUsers,
		CreatedBy:   room.CreatedBy,
		ActiveUsers: 0,
	}

	c.JSON(http.StatusCreated, response)
}

// GetRoomMessages 獲取聊天室的訊息
func (h *RoomHandler) GetRoomMessages(c *gin.Context) {
	// 獲取聊天室 ID
	roomID := c.Param("id")

	// 獲取訊息數量限制
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// 獲取訊息
	messages, err := h.roomService.GetRoomMessages(roomID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取訊息失敗"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetRoomUsers 獲取聊天室的用戶
func (h *RoomHandler) GetRoomUsers(c *gin.Context) {
	// 獲取聊天室 ID
	roomID := c.Param("id")

	// 獲取用戶
	users, err := h.roomService.GetRoomUsers(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "獲取用戶失敗"})
		return
	}

	c.JSON(http.StatusOK, users)
}
