package repository

import (
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDB 是一個模擬的 GORM DB
type MockDB struct {
	mock.Mock
	gormDB *gorm.DB
}

// 設置一個真實的 gorm.DB 實例，用於測試中返回
func (m *MockDB) SetGormDB(db *gorm.DB) {
	m.gormDB = db
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	m.Called(dest, conds)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	m.Called(dest, conds)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	m.Called(value)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	m.Called(value)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	m.Called(value, conds)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	m.Called(query, args)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Order(value interface{}) *gorm.DB {
	m.Called(value)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Limit(limit int) *gorm.DB {
	m.Called(limit)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
}

func (m *MockDB) Count(count *int64) *gorm.DB {
	m.Called(count)
	if m.gormDB != nil {
		return m.gormDB
	}
	return &gorm.DB{}
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
	mockDB := new(MockDB)

	// 設置默認的模擬行為
	mockDB.On("First", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Find", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Create", mock.Anything).Return(nil)
	mockDB.On("Save", mock.Anything).Return(nil)
	mockDB.On("Delete", mock.Anything, mock.Anything).Return(nil)
	mockDB.On("Where", mock.Anything, mock.Anything).Return(mockDB)
	mockDB.On("Order", mock.Anything).Return(mockDB)
	mockDB.On("Limit", mock.Anything).Return(mockDB)
	mockDB.On("Count", mock.Anything).Return(nil)

	return mockDB
}
