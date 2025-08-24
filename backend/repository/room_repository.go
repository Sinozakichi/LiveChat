package repository

import (
	"errors"
	"fmt"
	"livechat/backend/model"
	"time"

	"gorm.io/gorm"
)

// 定義錯誤
var (
	ErrRoomNotFound = errors.New("聊天室不存在")
	ErrUserNotFound = errors.New("用戶不存在")
)

// RoomRepository 管理聊天室數據
type RoomRepository struct {
	db DB
}

// DB 接口定義了 RoomRepository 所需的 GORM 方法
type DB interface {
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Order(value interface{}) *gorm.DB
	Limit(limit int) *gorm.DB
	Count(count *int64) *gorm.DB
}

// NewRoomRepository 創建一個新的聊天室儲存庫
func NewRoomRepository(db DB) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

// GetRoom 獲取指定的聊天室
func (r *RoomRepository) GetRoom(roomID string) (*model.Room, error) {
	var room model.Room

	// 直接使用字串 ID 查詢
	result := r.db.First(&room, "id = ?", roomID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, result.Error
	}

	return &room, nil
}

// GetAllRooms 獲取所有聊天室
func (r *RoomRepository) GetAllRooms() ([]model.Room, error) {
	var rooms []model.Room

	fmt.Println("Repository: Getting all rooms from database...")
	result := r.db.Find(&rooms, "is_active = ?", true)
	if result.Error != nil {
		fmt.Printf("Repository: Error getting rooms: %v\n", result.Error)
		return nil, result.Error
	}

	fmt.Printf("Repository: Found %d rooms in database\n", len(rooms))
	return rooms, nil
}

// CreateRoom 創建一個新的聊天室
func (r *RoomRepository) CreateRoom(room *model.Room) error {
	result := r.db.Create(room)
	return result.Error
}

// UpdateRoom 更新聊天室信息
func (r *RoomRepository) UpdateRoom(room *model.Room) error {
	result := r.db.Save(room)
	return result.Error
}

// GetRoomUsers 獲取聊天室的所有活躍用戶
func (r *RoomRepository) GetRoomUsers(roomID string) ([]model.RoomUser, error) {
	var users []model.RoomUser

	result := r.db.Where("room_id = ? AND is_active = ?", roomID, true).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// JoinRoom 用戶加入聊天室
func (r *RoomRepository) JoinRoom(roomID string, userID string, role string) error {
	roomUser := model.RoomUser{
		RoomID:       roomID,
		UserID:       userID,
		Role:         role,
		JoinedAt:     time.Now(),
		LastActiveAt: time.Now(),
		IsActive:     true,
	}

	result := r.db.Create(&roomUser)
	return result.Error
}

// LeaveRoom 用戶離開聊天室
func (r *RoomRepository) LeaveRoom(roomID string, userID string) error {
	var roomUser model.RoomUser

	result := r.db.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&roomUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return result.Error
	}

	roomUser.IsActive = false
	result = r.db.Save(&roomUser)
	return result.Error
}

// UpdateUserActivity 更新用戶在聊天室的活躍狀態
func (r *RoomRepository) UpdateUserActivity(roomID string, userID string) error {
	var roomUser model.RoomUser

	result := r.db.Where("room_id = ? AND user_id = ? AND is_active = ?", roomID, userID, true).First(&roomUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return result.Error
	}

	roomUser.LastActiveAt = time.Now()
	result = r.db.Save(&roomUser)
	return result.Error
}

// GetRoomMessages 獲取聊天室的訊息
func (r *RoomRepository) GetRoomMessages(roomID string, limit int) ([]model.Message, error) {
	var messages []model.Message

	result := r.db.Where("room_id = ?", roomID).Order("created_at desc").Limit(limit).Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}

	return messages, nil
}

// SaveMessage 保存聊天訊息
func (r *RoomRepository) SaveMessage(message *model.Message) error {
	result := r.db.Create(message)
	return result.Error
}

// CountActiveUsers 計算聊天室的活躍用戶數
func (r *RoomRepository) CountActiveUsers(roomID string) (int64, error) {
	var count int64

	result := r.db.Where("room_id = ? AND is_active = ?", roomID, true).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}
