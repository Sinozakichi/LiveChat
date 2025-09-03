package repository

import (
	"livechat/backend/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockDB 是一個混合型的資料庫模擬器，結合了 Mock 行為和真實的 SQLite 記憶體資料庫
// 這種設計提供了以下優點：
// 1. 測試隔離性：每個測試實例都有獨立的記憶體資料庫
// 2. 真實性：使用真實的 GORM 和 SQLite，避免純 Mock 可能遺漏的邊界情況
// 3. 效能：記憶體資料庫執行速度極快
// 4. 靈活性：可以通過 Mock 控制特定的行為或錯誤情況
type MockDB struct {
	mock.Mock // 提供 Mock 功能，用於模擬特定行為或錯誤
	*gorm.DB  // 嵌入真實的 GORM DB 實例，提供完整的資料庫功能
}

// SetGormDB 設置真實的 GORM DB 實例，用於需要真實資料庫操作的測試場景
// 通常在需要完整資料庫功能而非純 Mock 行為時使用
func (m *MockDB) SetGormDB(db *gorm.DB) {
	m.DB = db
}

// First 模擬 GORM 的 First 方法，用於查詢單筆記錄
// 此方法結合了 Mock 行為和真實資料庫查詢：
// 1. 首先檢查是否有設定的 Mock 行為（用於模擬錯誤或特定回應）
// 2. 如果沒有 Mock 行為，則使用真實的資料庫實例進行查詢
// 3. 這種設計讓測試既可以測試正常流程，也可以測試錯誤處理
func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	// 如果有真實的 DB 實例，直接使用真實的 First 查詢
	if m.DB != nil {
		return m.DB.First(dest, conds...)
	}

	// 否則使用 Mock 行為
	args := m.Called(dest, conds)
	if len(args) > 0 {
		if mockResult, ok := args.Get(0).(*MockGormDB); ok {
			// 回傳模擬的錯誤結果，用於測試錯誤處理路徑
			return &gorm.DB{Error: mockResult.Err}
		}
	}
	// 防禦性程式設計：回傳空的 DB 實例避免 nil pointer
	return &gorm.DB{}
}

// Find 模擬 GORM 的 Find 方法，用於查詢多筆記錄
// 行為模式與 First 相同：優先使用 Mock 行為，否則使用真實資料庫
func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	// 如果有真實的 DB 實例，直接使用真實的 Find 查詢
	if m.DB != nil {
		return m.DB.Find(dest, conds...)
	}

	// 否則使用 Mock 行為
	args := m.Called(dest, conds)
	if len(args) > 0 {
		if mockResult, ok := args.Get(0).(*MockGormDB); ok {
			return &gorm.DB{Error: mockResult.Err}
		}
	}
	return &gorm.DB{}
}

// Create 模擬 GORM 的 Create 方法，用於創建新記錄
// 在測試中常用於驗證新增操作的正確性和錯誤處理
func (m *MockDB) Create(value interface{}) *gorm.DB {
	// 如果有真實的 DB 實例，直接使用真實的 Create 操作
	if m.DB != nil {
		return m.DB.Create(value)
	}

	// 否則使用 Mock 行為
	args := m.Called(value)
	if len(args) > 0 {
		if mockResult, ok := args.Get(0).(*MockGormDB); ok {
			return &gorm.DB{Error: mockResult.Err}
		}
	}
	return &gorm.DB{}
}

// Save 模擬 GORM 的 Save 方法，用於更新記錄
// 主要用於測試更新操作的流程
func (m *MockDB) Save(value interface{}) *gorm.DB {
	// 如果有真實的 DB 實例，直接使用真實的 Save 操作
	if m.DB != nil {
		return m.DB.Save(value)
	}

	// 否則使用 Mock 行為
	m.Called(value)
	return &gorm.DB{}
}

// Delete 模擬 GORM 的 Delete 方法，用於刪除記錄
// 在測試中驗證刪除操作的邏輯正確性
func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	m.Called(value, conds)
	if m.DB != nil {
		return m.DB
	}
	return &gorm.DB{}
}

// Where 模擬 GORM 的 Where 方法，用於添加查詢條件
// 支援 GORM 的鏈式調用模式，是查詢建構的重要組件
func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	// 如果有真實的 DB 實例，直接使用真實的 Where 查詢
	if m.DB != nil {
		return m.DB.Where(query, args...)
	}

	// 否則使用 Mock 行為
	m.Called(query, args)
	return m.DB
}

// Order 模擬 GORM 的 Order 方法，用於設定查詢結果的排序
// 在測試分頁和排序功能時會用到
func (m *MockDB) Order(value interface{}) *gorm.DB {
	m.Called(value)
	return m.DB
}

// Limit 模擬 GORM 的 Limit 方法，用於限制查詢結果數量
// 主要用於測試分頁功能
func (m *MockDB) Limit(limit int) *gorm.DB {
	m.Called(limit)
	return m.DB
}

// Count 模擬 GORM 的 Count 方法，用於計算記錄數量
// 在測試統計功能和分頁總數時使用
func (m *MockDB) Count(count *int64) *gorm.DB {
	m.Called(count)
	return m.DB
}

// Model 模擬 GORM 的 Model 方法，用於指定操作的模型
// 是 GORM 查詢建構的起始點，在所有模型操作測試中都會使用
func (m *MockDB) Model(value interface{}) *gorm.DB {
	// 如果有真實的 DB 實例，直接使用真實的 Model 操作
	if m.DB != nil {
		return m.DB.Model(value)
	}

	// 否則使用 Mock 行為
	m.Called(value)
	return m.DB
}

// Update 模擬 GORM 的 Update 方法，用於更新記錄
func (m *MockDB) Update(column string, value interface{}) *gorm.DB {
	// 如果有真實的 DB 實例，直接使用真實的 Update 操作
	if m.DB != nil {
		return m.DB.Update(column, value)
	}

	// 否則使用 Mock 行為
	m.Called(column, value)
	return m.DB
}

// MockGormDB 是模擬 GORM 操作結果的輔助結構
// 用於控制 GORM 操作的回傳值，特別是錯誤狀態
// 這讓測試可以模擬各種資料庫操作的成功或失敗情況
type MockGormDB struct {
	mock.Mock
	Err error // 用於模擬資料庫操作錯誤
}

// Error 回傳模擬的錯誤狀態
// 實現 GORM 的錯誤介面，讓測試可以驗證錯誤處理邏輯
func (m *MockGormDB) Error() error {
	return m.Err
}

// RowsAffected 回傳模擬的受影響行數
// 用於測試 UPDATE、DELETE 等操作的影響行數檢查
func (m *MockGormDB) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

// NewMockDB 創建一個新的混合型測試資料庫實例
//
// 這個函數的設計理念：
// 1. 每次調用都創建全新的記憶體資料庫，確保測試隔離性
// 2. 使用 SQLite 的 ":memory:" 模式，資料存在記憶體中，測試結束後自動清除
// 3. 結合 Mock 和真實資料庫，既可以測試正常流程，也可以模擬錯誤情況
// 4. 效能優秀：記憶體資料庫比磁碟資料庫快數十倍
//
// 使用場景：
// - 單元測試：測試 Repository 層的資料存取邏輯
// - 整合測試：測試多個組件的協作
// - 錯誤處理測試：透過 Mock 模擬各種異常情況
func NewMockDB() *MockDB {
	// 創建一個內存SQLite資料庫用於測試
	// ":memory:" 表示資料庫完全存在記憶體中，不會寫入磁碟
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		// 如果連接失敗，測試環境有問題，應立即中止
		panic("failed to connect to test database")
	}

	// 自動遷移所有必要的資料表結構
	// 這樣 NewMockDB() 就可以支援完整的資料庫操作
	err = db.AutoMigrate(
		&model.User{},     // 使用者表
		&model.Room{},     // 聊天室表
		&model.RoomUser{}, // 聊天室使用者關聯表
		&model.Message{},  // 訊息表
	)
	if err != nil {
		panic("failed to migrate database schema: " + err.Error())
	}

	// 創建混合型 MockDB 實例
	mockDB := &MockDB{
		Mock: mock.Mock{}, // 初始化 Mock 功能
		DB:   db,          // 嵌入真實的 GORM 實例
	}

	return mockDB
}

// NewMockDBWithSchema 創建包含完整資料庫結構的測試資料庫
//
// 注意：現在 NewMockDB() 已經包含完整的資料庫結構了，
// 這個函數主要是為了向後兼容性而保留
// 建議新的測試直接使用 NewMockDB()
func NewMockDBWithSchema() *MockDB {
	// 現在直接委託給 NewMockDB()
	return NewMockDB()
}
