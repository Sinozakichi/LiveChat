package service

import (
	"livechat/backend/model"
)

// RoomRepository 定義了聊天室儲存庫的接口
type RoomRepository interface {
	GetRoom(roomID string) (*model.Room, error)
	GetAllRooms() ([]model.Room, error)
	CreateRoom(room *model.Room) error
	UpdateRoom(room *model.Room) error
	GetRoomUsers(roomID string) ([]model.RoomUser, error)
	JoinRoom(roomID string, userID string, role string) error
	LeaveRoom(roomID string, userID string) error
	UpdateUserActivity(roomID string, userID string) error
	GetRoomMessages(roomID string, limit int) ([]model.Message, error)
	SaveMessage(message *model.Message) error
	CountActiveUsers(roomID string) (int64, error)
}

// RoomService 處理聊天室的業務邏輯
type RoomService struct {
	roomRepo RoomRepository
}

// RoomData 包含創建聊天室所需的數據
type RoomData struct {
	Name        string
	Description string
	IsPublic    bool
	MaxUsers    int
}

// NewRoomService 創建一個新的聊天室服務
func NewRoomService(roomRepo RoomRepository) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
	}
}

// GetRoom 獲取指定的聊天室
func (s *RoomService) GetRoom(roomID string) (*model.Room, error) {
	return s.roomRepo.GetRoom(roomID)
}

// GetAllRooms 獲取所有聊天室
func (s *RoomService) GetAllRooms() ([]model.Room, error) {
	return s.roomRepo.GetAllRooms()
}

// CreateRoom 創建一個新的聊天室
func (s *RoomService) CreateRoom(data RoomData, createdBy string) (*model.Room, error) {
	room := &model.Room{
		Name:        data.Name,
		Description: data.Description,
		IsPublic:    data.IsPublic,
		MaxUsers:    data.MaxUsers,
		CreatedBy:   createdBy,
		IsActive:    true,
	}

	err := s.roomRepo.CreateRoom(room)
	if err != nil {
		return nil, err
	}

	return room, nil
}

// JoinRoom 用戶加入聊天室
func (s *RoomService) JoinRoom(roomID string, userID string, role string) error {
	// 檢查聊天室是否存在
	_, err := s.roomRepo.GetRoom(roomID)
	if err != nil {
		return err
	}

	// 加入聊天室
	return s.roomRepo.JoinRoom(roomID, userID, role)
}

// LeaveRoom 用戶離開聊天室
func (s *RoomService) LeaveRoom(roomID string, userID string) error {
	return s.roomRepo.LeaveRoom(roomID, userID)
}

// GetRoomMessages 獲取聊天室的訊息
func (s *RoomService) GetRoomMessages(roomID string, limit int) ([]model.Message, error) {
	return s.roomRepo.GetRoomMessages(roomID, limit)
}

// SendMessage 發送訊息到聊天室
func (s *RoomService) SendMessage(roomID string, userID string, content string) error {
	// 檢查聊天室是否存在
	_, err := s.roomRepo.GetRoom(roomID)
	if err != nil {
		return err
	}

	// 創建訊息
	message := &model.Message{
		RoomID:          roomID,
		UserID:          userID,
		Content:         content,
		IsSystemMessage: false,
	}

	// 保存訊息
	err = s.roomRepo.SaveMessage(message)
	if err != nil {
		return err
	}

	// 更新用戶活躍狀態
	return s.roomRepo.UpdateUserActivity(roomID, userID)
}

// SendSystemMessage 發送系統訊息到聊天室
func (s *RoomService) SendSystemMessage(roomID string, content string) error {
	// 檢查聊天室是否存在
	_, err := s.roomRepo.GetRoom(roomID)
	if err != nil {
		return err
	}

	// 創建系統訊息
	message := &model.Message{
		RoomID:          roomID,
		UserID:          "",
		Content:         content,
		IsSystemMessage: true,
	}

	// 保存訊息
	return s.roomRepo.SaveMessage(message)
}

// GetRoomActiveUserCount 獲取聊天室的活躍用戶數
func (s *RoomService) GetRoomActiveUserCount(roomID string) (int64, error) {
	return s.roomRepo.CountActiveUsers(roomID)
}

// GetRoomUsers 獲取聊天室的用戶
func (s *RoomService) GetRoomUsers(roomID string) ([]model.RoomUser, error) {
	return s.roomRepo.GetRoomUsers(roomID)
}
