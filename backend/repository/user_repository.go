package repository

import (
	"errors"
	"livechat/backend/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserRepository 定義用戶儲存庫接口
type UserRepository interface {
	CreateUser(user *model.User) error
	GetUserByID(id string) (*model.User, error)
	GetUserByUsername(username string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
	CheckUserCredentials(username, password string) (*model.User, error)
}

// UserDB 接口定義了 UserRepository 所需的 GORM 方法
type UserDB interface {
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Model(value interface{}) *gorm.DB
	Count(count *int64) *gorm.DB
}

// UserRepositoryImpl 實現 UserRepository 接口
type UserRepositoryImpl struct {
	db UserDB
}

// NewUserRepository 創建一個新的用戶儲存庫
func NewUserRepository(db UserDB) UserRepository {
	return &UserRepositoryImpl{
		db: db,
	}
}

// CreateUser 創建一個新用戶
func (r *UserRepositoryImpl) CreateUser(user *model.User) error {
	// 檢查用戶名是否已存在
	var count int64
	r.db.Model(&model.User{}).Where("username = ?", user.Username).Count(&count)
	if count > 0 {
		return ErrUserAlreadyExists
	}

	// 檢查電子郵件是否已存在
	r.db.Model(&model.User{}).Where("email = ?", user.Email).Count(&count)
	if count > 0 {
		return ErrEmailAlreadyExists
	}

	// 哈希密碼
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// 創建用戶
	result := r.db.Create(user)
	return result.Error
}

// GetUserByID 根據 ID 獲取用戶
func (r *UserRepositoryImpl) GetUserByID(id string) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByUsername 根據用戶名獲取用戶
func (r *UserRepositoryImpl) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, "username = ?", username)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail 根據電子郵件獲取用戶
func (r *UserRepositoryImpl) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	result := r.db.First(&user, "email = ?", email)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// UpdateUser 更新用戶
func (r *UserRepositoryImpl) UpdateUser(user *model.User) error {
	result := r.db.Save(user)
	return result.Error
}

// DeleteUser 刪除用戶
func (r *UserRepositoryImpl) DeleteUser(id string) error {
	result := r.db.Delete(&model.User{}, "id = ?", id)
	return result.Error
}

// CheckUserCredentials 檢查用戶憑證
func (r *UserRepositoryImpl) CheckUserCredentials(username, password string) (*model.User, error) {
	user, err := r.GetUserByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 檢查密碼
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
