package migrations

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Migration 定義了遷移接口
type Migration interface {
	Name() string
	Up(db *gorm.DB) error
	Down(db *gorm.DB) error
}

// MigrationRecord 代表資料庫中的遷移記錄
type MigrationRecord struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"size:255;not null;uniqueIndex"`
	AppliedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName 指定 MigrationRecord 模型的表名
func (MigrationRecord) TableName() string {
	return "migrations"
}

// Migrator 處理資料庫遷移
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator 創建一個新的遷移器
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db: db,
		migrations: []Migration{
			Migration001InitialSchema{},
			// 在這裡添加更多遷移
		},
	}
}

// Initialize 初始化遷移表
func (m *Migrator) Initialize() error {
	// 創建遷移表
	if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

// MigrateUp 執行所有未應用的遷移
func (m *Migrator) MigrateUp() error {
	if err := m.Initialize(); err != nil {
		return err
	}

	for _, migration := range m.migrations {
		// 檢查遷移是否已經應用
		var count int64
		if err := m.db.Model(&MigrationRecord{}).Where("name = ?", migration.Name()).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count == 0 {
			// 執行遷移
			fmt.Printf("Applying migration: %s\n", migration.Name())
			if err := migration.Up(m.db); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Name(), err)
			}

			// 記錄遷移
			record := MigrationRecord{
				Name:      migration.Name(),
				AppliedAt: time.Now(),
			}
			if err := m.db.Create(&record).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Name(), err)
			}
			fmt.Printf("Migration %s applied successfully\n", migration.Name())
		} else {
			fmt.Printf("Migration %s already applied, skipping\n", migration.Name())
		}
	}

	return nil
}

// MigrateDown 回滾最後一個遷移
func (m *Migrator) MigrateDown() error {
	if err := m.Initialize(); err != nil {
		return err
	}

	// 獲取最後一個應用的遷移
	var lastMigration MigrationRecord
	if err := m.db.Order("applied_at DESC").First(&lastMigration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Println("No migrations to roll back")
			return nil
		}
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	// 找到對應的遷移
	var targetMigration Migration
	for _, migration := range m.migrations {
		if migration.Name() == lastMigration.Name {
			targetMigration = migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %s not found in registered migrations", lastMigration.Name)
	}

	// 執行回滾
	fmt.Printf("Rolling back migration: %s\n", targetMigration.Name())
	if err := targetMigration.Down(m.db); err != nil {
		return fmt.Errorf("failed to roll back migration %s: %w", targetMigration.Name(), err)
	}

	// 刪除遷移記錄
	if err := m.db.Delete(&lastMigration).Error; err != nil {
		return fmt.Errorf("failed to delete migration record %s: %w", targetMigration.Name(), err)
	}

	fmt.Printf("Migration %s rolled back successfully\n", targetMigration.Name())
	return nil
}
