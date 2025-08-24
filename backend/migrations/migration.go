package migrations

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Migration 定義遷移接口
type Migration interface {
	ID() string
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

// Migrator 是遷移管理器
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator 創建一個新的遷移管理器
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db: db,
		migrations: []Migration{
			Migration001InitialSchema{},
			Migration002UserSchema{},
			Migration003RenamePasswordColumn{},
		},
	}
}

// MigrateUp 執行所有未應用的遷移
func (m *Migrator) MigrateUp() error {
	// 確保遷移記錄表存在並具有正確的結構
	var tableExists int64
	m.db.Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'migrations' AND table_type = 'BASE TABLE'").Scan(&tableExists)

	if tableExists > 0 {
		// 檢查ID欄位類型是否正確
		var idType string
		m.db.Raw("SELECT data_type FROM information_schema.columns WHERE table_schema = CURRENT_SCHEMA() AND table_name = 'migrations' AND column_name = 'id'").Scan(&idType)

		if idType != "uuid" {
			fmt.Println("Detected incorrect ID type in migrations table, recreating...")
			// 刪除現有的migrations表
			if err := m.db.Exec("DROP TABLE IF EXISTS migrations").Error; err != nil {
				return fmt.Errorf("failed to drop existing migrations table: %w", err)
			}
			tableExists = 0
		}
	}

	if tableExists == 0 {
		// 創建migrations表，確保ID是UUID類型
		if err := m.db.Exec(`
			CREATE TABLE migrations (
				id UUID PRIMARY KEY,
				name VARCHAR(255) NOT NULL UNIQUE,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`).Error; err != nil {
			return fmt.Errorf("failed to create migrations table: %w", err)
		}
		fmt.Println("Created migrations table with UUID primary key")
	}

	// 執行遷移
	for _, migration := range m.migrations {
		// 檢查遷移是否已應用
		var count int64
		if err := m.db.Model(&MigrationRecord{}).Where("name = ?", migration.ID()).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			fmt.Printf("Migration %s already applied, skipping.\n", migration.ID())
			continue
		}

		// 應用遷移
		fmt.Printf("Applying migration: %s\n", migration.ID())
		if err := migration.Up(m.db); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.ID(), err)
		}

		// 記錄遷移
		record := MigrationRecord{
			ID:        uuid.New().String(),
			Name:      migration.ID(),
			AppliedAt: time.Now(),
		}
		if err := m.db.Create(&record).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.ID(), err)
		}

		fmt.Printf("Migration %s applied successfully.\n", migration.ID())
	}

	return nil
}

// MigrateDown 回滾所有遷移
func (m *Migrator) MigrateDown() error {
	// 確保遷移記錄表存在
	if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to auto migrate migration records: %w", err)
	}

	// 按照相反的順序回滾遷移
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]

		// 檢查遷移是否已應用
		var record MigrationRecord
		result := m.db.Where("name = ?", migration.ID()).First(&record)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				fmt.Printf("Migration %s not applied, skipping rollback.\n", migration.ID())
				continue
			}
			return fmt.Errorf("failed to check migration status: %w", result.Error)
		}

		// 回滾遷移
		fmt.Printf("Rolling back migration: %s\n", migration.ID())
		if err := migration.Down(m.db); err != nil {
			return fmt.Errorf("failed to roll back migration %s: %w", migration.ID(), err)
		}

		// 刪除遷移記錄
		if err := m.db.Delete(&record).Error; err != nil {
			return fmt.Errorf("failed to delete migration record %s: %w", migration.ID(), err)
		}

		fmt.Printf("Migration %s rolled back successfully.\n", migration.ID())
	}

	return nil
}
