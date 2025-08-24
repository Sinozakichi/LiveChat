package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// Migration001InitialSchema 初始化資料庫結構
type Migration001InitialSchema struct{}

// ID 返回遷移 ID
func (m Migration001InitialSchema) ID() string {
	return "001_initial_schema"
}

// Up 執行遷移
func (m Migration001InitialSchema) Up(db *gorm.DB) error {
	fmt.Println("Running migration: 001_initial_schema")

	// 創建 rooms 表 (已由模型中的 TableName 方法指定表名)
	// 禁用外鍵約束，避免遷移錯誤
	if err := db.Exec("CREATE TABLE IF NOT EXISTS rooms (id UUID PRIMARY KEY, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP, name VARCHAR(255) NOT NULL, description TEXT, is_public BOOLEAN DEFAULT true, max_users INTEGER DEFAULT 100, created_by VARCHAR(255), is_active BOOLEAN DEFAULT true)").Error; err != nil {
		return fmt.Errorf("failed to create rooms table: %w", err)
	}

	// 創建 room_users 表 (已由模型中的 TableName 方法指定表名)
	if err := db.Exec("CREATE TABLE IF NOT EXISTS room_users (id SERIAL PRIMARY KEY, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP, room_id VARCHAR(255), user_id VARCHAR(255), role VARCHAR(20) DEFAULT 'member', joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, is_active BOOLEAN DEFAULT true)").Error; err != nil {
		return fmt.Errorf("failed to create room_users table: %w", err)
	}

	// 創建 messages 表 (已由模型中的 TableName 方法指定表名)
	if err := db.Exec("CREATE TABLE IF NOT EXISTS messages (id SERIAL PRIMARY KEY, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP, room_id VARCHAR(255), user_id VARCHAR(255), content TEXT NOT NULL, is_system_message BOOLEAN DEFAULT false)").Error; err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	// 創建索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_room_users_room_id ON room_users(room_id)").Error; err != nil {
		return fmt.Errorf("failed to create index on room_users.room_id: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_room_users_user_id ON room_users(user_id)").Error; err != nil {
		return fmt.Errorf("failed to create index on room_users.user_id: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id)").Error; err != nil {
		return fmt.Errorf("failed to create index on messages.room_id: %w", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id)").Error; err != nil {
		return fmt.Errorf("failed to create index on messages.user_id: %w", err)
	}

	fmt.Println("Migration 001_initial_schema completed successfully")
	return nil
}

// Down 回滾遷移
func (m Migration001InitialSchema) Down(db *gorm.DB) error {
	fmt.Println("Rolling back migration: 001_initial_schema")

	// 刪除表格（按照相反的順序）
	if err := db.Exec("DROP TABLE IF EXISTS messages").Error; err != nil {
		return fmt.Errorf("failed to drop messages table: %w", err)
	}

	if err := db.Exec("DROP TABLE IF EXISTS room_users").Error; err != nil {
		return fmt.Errorf("failed to drop room_users table: %w", err)
	}

	if err := db.Exec("DROP TABLE IF EXISTS rooms").Error; err != nil {
		return fmt.Errorf("failed to drop rooms table: %w", err)
	}

	fmt.Println("Rollback of 001_initial_schema completed successfully")
	return nil
}
