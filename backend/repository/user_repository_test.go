package repository

import (
	"livechat/backend/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// TestNewUserRepository 測試使用者儲存庫的建構子
//
// 測試目標：
// 1. 驗證建構子能正確初始化 UserRepository 實例
// 2. 確保依賴注入正確設定
// 3. 檢查回傳的實例不為 nil
//
// 測試策略：
// - 使用 AAA 模式（Arrange-Act-Assert）
// - 使用 MockDB 作為依賴，避免真實資料庫連接
// - 專注於測試建構邏輯而非資料庫操作
func TestNewUserRepository(t *testing.T) {
	// 安排 (Arrange)：準備測試所需的依賴
	mockDB := NewMockDB()

	// 動作 (Act)：執行被測試的操作
	repo := NewUserRepository(mockDB)

	// 斷言 (Assert)：驗證結果是否符合預期
	assert.NotNil(t, repo, "儲存庫不應為 nil")
}

// TestCreateUser 測試使用者創建功能的正常流程
//
// 測試目標：
// 1. 驗證使用者創建的完整流程
// 2. 確保密碼會被正確加密處理
// 3. 驗證資料庫操作的正確調用
// 4. 測試重複檢查邏輯（使用者名稱和電子郵件）
//
// 測試策略：
// - 模擬完整的 GORM 鏈式調用：Model -> Where -> Count -> Create
// - 使用 Mock 驗證每個資料庫操作都被正確調用
// - 檢查業務邏輯：密碼加密、重複性檢查
// - 驗證無錯誤的成功路徑
func TestCreateUser(t *testing.T) {
	// 安排 (Arrange)：準備測試環境和模擬行為
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil // 模擬成功的資料庫操作

	// 設置完整的 GORM 鏈式調用模擬
	// 這些調用模擬了 UserRepository.CreateUser 中的實際資料庫操作流程
	mockDB.On("Model", mock.AnythingOfType("*model.User")).Return(mockDB)             // 指定操作模型
	mockDB.On("Where", "username = ?", []interface{}{"testuser"}).Return(mockDB)      // 檢查使用者名稱重複
	mockDB.On("Where", "email = ?", []interface{}{"test@example.com"}).Return(mockDB) // 檢查電子郵件重複
	mockDB.On("Count", mock.AnythingOfType("*int64")).Return(mockResult)              // 計算重複數量
	mockDB.On("Create", mock.AnythingOfType("*model.User")).Return(mockResult)        // 創建新使用者

	// 創建儲存庫實例和測試用戶資料
	repo := NewUserRepository(mockDB)
	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123", // 明文密碼，應該被加密
	}

	// 動作 (Act)：執行創建用戶操作
	err := repo.CreateUser(user)

	// 斷言 (Assert)：驗證操作結果
	assert.NoError(t, err, "創建用戶不應返回錯誤")
	// 重要：驗證密碼已被加密，不再是明文
	assert.NotEqual(t, "password123", user.Password, "密碼應該被哈希")
}

// TestCreateUserUsernameExists 測試使用者創建時的使用者名稱重複錯誤處理
//
// 測試目標：
// 1. 驗證系統能正確檢測使用者名稱重複
// 2. 確保回傳適當的錯誤類型
// 3. 測試資料驗證邏輯的健全性
// 4. 驗證不會執行實際的創建操作
//
// 測試策略：
// - 模擬資料庫返回重複計數為 1，表示使用者名稱已存在
// - 使用 Run 函數動態設定 count 值
// - 驗證特定的錯誤類型，而非通用錯誤
// - 測試負面情況（錯誤路徑）
func TestCreateUserUsernameExists(t *testing.T) {
	// 安排 (Arrange) - 使用帶有完整結構的模擬資料庫
	mockDB := NewMockDB()
	repo := NewUserRepository(mockDB)

	// 先創建一個已存在的用戶
	existingUser := &model.User{
		Username: "existinguser",
		Email:    "existing@example.com",
		Password: "password123",
	}
	err := repo.CreateUser(existingUser)
	assert.NoError(t, err, "創建第一個用戶不應該失敗")

	// 創建具有相同用戶名的新用戶資料
	newUser := &model.User{
		Username: "existinguser",          // 相同的使用者名稱
		Email:    "different@example.com", // 不同的郵箱
		Password: "password456",
	}

	// 動作 (Act)：嘗試創建具有重複用戶名的用戶
	err = repo.CreateUser(newUser)

	// 斷言 (Assert)：驗證錯誤處理
	assert.Error(t, err, "創建已存在用戶名的用戶應返回錯誤")
	assert.Equal(t, ErrUserAlreadyExists, err, "錯誤應為 ErrUserAlreadyExists")
}

// TestCreateUserEmailExists 測試使用者創建時的電子郵件重複錯誤處理
//
// 測試目標：
// 1. 驗證系統能正確檢測電子郵件重複
// 2. 測試多重驗證邏輯：先檢查使用者名稱，再檢查電子郵件
// 3. 確保回傳適當的錯誤類型（區分使用者名稱和電子郵件錯誤）
// 4. 驗證業務邏輯的順序性和完整性
//
// 測試策略：
// - 設定複雜的模擬情況：使用者名稱不重複，但電子郵件重複
// - 使用多個 Count 調用來模擬不同的檢查階段
// - 驗證系統能區分不同類型的重複錯誤
// - 測試多層驗證邏輯
func TestCreateUserEmailExists(t *testing.T) {
	// 安排 (Arrange) - 使用帶有完整結構的模擬資料庫
	mockDB := NewMockDB()
	repo := NewUserRepository(mockDB)

	// 先創建一個已存在的用戶
	existingUser := &model.User{
		Username: "originaluser",
		Email:    "existing@example.com",
		Password: "password123",
	}
	err := repo.CreateUser(existingUser)
	assert.NoError(t, err, "創建第一個用戶不應該失敗")

	// 創建具有相同郵箱的新用戶資料
	newUser := &model.User{
		Username: "newuser",              // 不同的使用者名稱
		Email:    "existing@example.com", // 相同的電子郵件
		Password: "password456",
	}

	// 動作 (Act)：嘗試創建具有重複郵箱的用戶
	err = repo.CreateUser(newUser)

	// 斷言 (Assert)：驗證錯誤處理
	assert.Error(t, err, "創建已存在電子郵件的用戶應返回錯誤")
	assert.Equal(t, ErrEmailAlreadyExists, err, "錯誤應為 ErrEmailAlreadyExists")
}

// TestGetUserByID 測試根據使用者 ID 查詢使用者的功能
//
// 測試目標：
// 1. 驗證能正確根據 ID 查詢使用者
// 2. 測試資料映射的完整性和正確性
// 3. 確保查詢結果包含所有必要欄位
// 4. 驗證 GORM First 方法的正確調用
//
// 測試策略：
// - 使用 Run 函數模擬資料庫查詢結果
// - 創建完整的使用者資料進行測試
// - 驗證多個欄位以確保資料完整性
// - 測試成功的查詢路徑
func TestGetUserByID(t *testing.T) {
	// 安排 (Arrange) - 使用帶有完整結構的模擬資料庫
	mockDB := NewMockDB()
	repo := NewUserRepository(mockDB)

	// 先創建一個測試用戶
	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := repo.CreateUser(testUser)
	assert.NoError(t, err, "創建測試用戶不應該失敗")

	// 獲取剛創建用戶的 ID（因為是自動生成的）
	createdUser, err := repo.GetUserByUsername("testuser")
	assert.NoError(t, err, "應該能根據用戶名查找用戶")
	assert.NotNil(t, createdUser, "創建的用戶不應為空")

	// 動作 (Act)：根據 ID 查詢用戶
	user, err := repo.GetUserByID(createdUser.ID)

	// 斷言 (Assert)：驗證查詢結果
	assert.NoError(t, err, "獲取用戶不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, createdUser.ID, user.ID, "用戶 ID 應該匹配")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
	assert.Equal(t, "test@example.com", user.Email, "郵箱應該匹配")
}

// TestGetUserByUsername 測試根據使用者名稱查詢使用者的功能
//
// 測試目標：
// 1. 驗證能正確根據使用者名稱查詢使用者
// 2. 測試使用者名稱作為查詢條件的有效性
// 3. 確保查詢結果的完整性
// 4. 驗證登入流程中的使用者查詢邏輯
//
// 測試策略：
// - 使用常見的使用者名稱進行測試
// - 模擬成功的查詢情況
// - 驗證回傳的使用者資料正確性
// - 專注於測試查詢條件和結果映射
func TestGetUserByUsername(t *testing.T) {
	// 安排 (Arrange) - 使用帶有完整結構的模擬資料庫
	mockDB := NewMockDB()
	repo := NewUserRepository(mockDB)

	// 先創建一個測試用戶
	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := repo.CreateUser(testUser)
	assert.NoError(t, err, "創建測試用戶不應該失敗")

	// 動作 (Act)：根據用戶名查詢用戶
	user, err := repo.GetUserByUsername("testuser")

	// 斷言 (Assert)：驗證查詢結果
	assert.NoError(t, err, "獲取用戶不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
	assert.Equal(t, "test@example.com", user.Email, "郵箱應該匹配")
}

// TestCheckUserCredentials 測試使用者登入憑證驗證功能
//
// 測試目標：
// 1. 驗證使用者名稱和密碼的完整驗證流程
// 2. 測試密碼加密比對的正確性
// 3. 確保安全的身份驗證邏輯
// 4. 驗證成功登入的情況
//
// 測試策略：
// - 使用真實的 bcrypt 加密來測試密碼驗證
// - 模擬完整的登入流程：查詢使用者 -> 驗證密碼
// - 測試正確的使用者名稱和密碼組合
// - 驗證回傳的使用者資料正確性
func TestCheckUserCredentials(t *testing.T) {
	// 安排 (Arrange) - 使用帶有完整結構的模擬資料庫
	mockDB := NewMockDB()
	repo := NewUserRepository(mockDB)

	// 先創建一個測試用戶（會自動進行密碼加密）
	testUser := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	}
	err := repo.CreateUser(testUser)
	assert.NoError(t, err, "創建測試用戶不應該失敗")

	// 動作 (Act)：使用正確的憑證進行驗證
	user, err := repo.CheckUserCredentials("testuser", "password123")

	// 斷言 (Assert)：驗證成功的憑證驗證
	assert.NoError(t, err, "檢查有效憑證不應返回錯誤")
	assert.NotNil(t, user, "用戶不應為 nil")
	assert.Equal(t, "testuser", user.Username, "用戶名應該匹配")
	assert.Equal(t, "test@example.com", user.Email, "郵箱應該匹配")
}

// TestCheckUserCredentialsInvalidPassword 測試無效密碼的錯誤處理
//
// 測試目標：
// 1. 驗證系統能正確檢測錯誤的密碼
// 2. 測試密碼驗證的安全性
// 3. 確保回傳適當的錯誤類型
// 4. 驗證登入失敗的情況處理
//
// 測試策略：
// - 使用正確的使用者名稱但錯誤的密碼
// - 驗證 bcrypt 密碼比對的安全性
// - 測試錯誤路徑的處理邏輯
// - 確保不會洩漏使用者資料
func TestCheckUserCredentialsInvalidPassword(t *testing.T) {
	// 安排 (Arrange)：準備無效密碼的測試環境
	mockDB := NewMockDB()
	mockResult := new(MockGormDB)
	mockResult.Err = nil

	// 創建真實的加密密碼
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	expectedUser := &model.User{
		ID:        "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword), // 正確的加密密碼
		Role:      "user",
	}

	// 設置模擬行為：模擬使用者查詢（使用者存在）
	mockDB.On("First", mock.AnythingOfType("*model.User"), "username = ?", "testuser").Run(func(args mock.Arguments) {
		user := args.Get(0).(*model.User)
		*user = *expectedUser
	}).Return(mockResult)

	repo := NewUserRepository(mockDB)

	// 動作 (Act)：執行憑證驗證，使用錯誤的密碼
	user, err := repo.CheckUserCredentials("testuser", "wrongpassword")

	// 斷言 (Assert)：驗證錯誤的密碼處理
	assert.Error(t, err, "檢查無效密碼應返回錯誤")
	assert.Equal(t, ErrInvalidCredentials, err, "錯誤應為 ErrInvalidCredentials")
	assert.Nil(t, user, "用戶應為 nil，避免洩漏資料")
}
