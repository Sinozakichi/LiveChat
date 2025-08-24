package repository

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockDB 是一個模擬的 GORM DB
type MockDB struct {
	mock.Mock
	*gorm.DB
}

// 設置一個真實的 gorm.DB 實例，用於測試中返回
func (m *MockDB) SetGormDB(db *gorm.DB) {
	m.DB = db
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	if len(args) > 0 {
		if mockResult, ok := args.Get(0).(*MockGormDB); ok {
			return &gorm.DB{Error: mockResult.Err}
		}
	}
	if m.DB != nil {
		return m.DB
	}
	return &gorm.DB{}
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	if len(args) > 0 {
		if mockResult, ok := args.Get(0).(*MockGormDB); ok {
			return &gorm.DB{Error: mockResult.Err}
		}
	}
	if m.DB != nil {
		return m.DB
	}
	return &gorm.DB{}
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	if len(args) > 0 {
		if mockResult, ok := args.Get(0).(*MockGormDB); ok {
			return &gorm.DB{Error: mockResult.Err}
		}
	}
	if m.DB != nil {
		return m.DB
	}
	return &gorm.DB{}
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	m.Called(value)
	if m.DB != nil {
		return m.DB
	}
	return &gorm.DB{}
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	m.Called(value, conds)
	if m.DB != nil {
		return m.DB
	}
	return &gorm.DB{}
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	m.Called(query, args)
	// 返回嵌入的DB來支持鏈式調用
	return m.DB
}

func (m *MockDB) Order(value interface{}) *gorm.DB {
	m.Called(value)
	return m.DB
}

func (m *MockDB) Limit(limit int) *gorm.DB {
	m.Called(limit)
	return m.DB
}

func (m *MockDB) Count(count *int64) *gorm.DB {
	m.Called(count)
	return m.DB
}

func (m *MockDB) Model(value interface{}) *gorm.DB {
	m.Called(value)
	return m.DB
}

// MockGormDB 是一個模擬的 GORM DB 結果
type MockGormDB struct {
	mock.Mock
	Err error
}

func (m *MockGormDB) Error() error {
	return m.Err
}

func (m *MockGormDB) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

// NewMockDB 創建一個新的模擬 DB
func NewMockDB() *MockDB {
	// 創建一個內存SQLite資料庫用於測試
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	mockDB := &MockDB{
		Mock: mock.Mock{},
		DB:   db,
	}

	return mockDB
}
