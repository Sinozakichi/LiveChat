package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// Migration003RenamePasswordColumn 重新命名密碼欄位
type Migration003RenamePasswordColumn struct{}

// ID 返回遷移 ID
func (m Migration003RenamePasswordColumn) ID() string {
	return "003_rename_password_column"
}

// Up 執行遷移
func (m Migration003RenamePasswordColumn) Up(db *gorm.DB) error {
	fmt.Println("Running migration: 003_rename_password_column")

	// 檢查是否存在password欄位
	var passwordColumnExists int64
	db.Raw("SELECT count(*) FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'users' AND column_name = 'password'").Scan(&passwordColumnExists)

	// 檢查是否存在password_hash欄位
	var passwordHashColumnExists int64
	db.Raw("SELECT count(*) FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'users' AND column_name = 'password_hash'").Scan(&passwordHashColumnExists)

	if passwordColumnExists > 0 && passwordHashColumnExists == 0 {
		// 重新命名password欄位為password_hash
		if err := db.Exec("ALTER TABLE users RENAME COLUMN password TO password_hash").Error; err != nil {
			return fmt.Errorf("failed to rename password column to password_hash: %w", err)
		}
		fmt.Println("Successfully renamed password column to password_hash")
	} else if passwordHashColumnExists > 0 {
		fmt.Println("password_hash column already exists, skipping rename")
	} else {
		// 如果兩個欄位都不存在，創建password_hash欄位
		if err := db.Exec("ALTER TABLE users ADD COLUMN password_hash VARCHAR(255) NOT NULL DEFAULT ''").Error; err != nil {
			return fmt.Errorf("failed to add password_hash column: %w", err)
		}
		fmt.Println("Added password_hash column to users table")
	}

	fmt.Println("Migration 003_rename_password_column completed successfully")
	return nil
}

// Down 回滾遷移
func (m Migration003RenamePasswordColumn) Down(db *gorm.DB) error {
	fmt.Println("Rolling back migration: 003_rename_password_column")

	// 檢查是否存在password_hash欄位
	var passwordHashColumnExists int64
	db.Raw("SELECT count(*) FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'users' AND column_name = 'password_hash'").Scan(&passwordHashColumnExists)

	if passwordHashColumnExists > 0 {
		// 重新命名password_hash欄位為password
		if err := db.Exec("ALTER TABLE users RENAME COLUMN password_hash TO password").Error; err != nil {
			return fmt.Errorf("failed to rename password_hash column to password: %w", err)
		}
		fmt.Println("Successfully renamed password_hash column back to password")
	}

	fmt.Println("Rollback of 003_rename_password_column completed successfully")
	return nil
}
