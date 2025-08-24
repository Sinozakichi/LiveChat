package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// Migration002UserSchema 添加用戶表
type Migration002UserSchema struct{}

// ID 返回遷移 ID
func (m Migration002UserSchema) ID() string {
	return "002_user_schema"
}

// Up 執行遷移
func (m Migration002UserSchema) Up(db *gorm.DB) error {
	fmt.Println("Running migration: 002_user_schema")

	// 創建 users 表
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			deleted_at TIMESTAMP,
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'user',
			is_verified BOOLEAN DEFAULT FALSE,
			CONSTRAINT users_username_unique UNIQUE (username),
			CONSTRAINT users_email_unique UNIQUE (email)
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// 創建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)").Error; err != nil {
		return fmt.Errorf("failed to create index on users.username: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
		return fmt.Errorf("failed to create index on users.email: %w", err)
	}

	fmt.Println("Migration 002_user_schema completed successfully")
	return nil
}

// Down 回滾遷移
func (m Migration002UserSchema) Down(db *gorm.DB) error {
	fmt.Println("Rolling back migration: 002_user_schema")

	if err := db.Exec("DROP TABLE IF EXISTS users").Error; err != nil {
		return fmt.Errorf("failed to drop users table: %w", err)
	}

	fmt.Println("Rollback of 002_user_schema completed successfully")
	return nil
}
