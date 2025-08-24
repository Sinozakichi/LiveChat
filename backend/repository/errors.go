package repository

import "errors"

// 共同的錯誤定義
var (
	ErrUserNotFound       = errors.New("用戶不存在")
	ErrUserAlreadyExists  = errors.New("用戶已存在")
	ErrEmailAlreadyExists = errors.New("電子郵件已被使用")
	ErrInvalidCredentials = errors.New("用戶名或密碼錯誤")
	ErrRoomNotFound       = errors.New("聊天室不存在")
)
