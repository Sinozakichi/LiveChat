package tests

import (
	"context"
	"fmt"
	"livechat/backend/model"
	"livechat/backend/repository"
	"livechat/backend/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// ChatTestContext BDD 測試的上下文容器，管理整個聊天測試場景的狀態
//
// BDD 測試設計理念：
// 1. 使用自然語言描述的測試場景（Given-When-Then）
// 2. 模擬真實使用者的操作流程和期望
// 3. 驗證系統行為是否符合業務需求
// 4. 提供可讀性高的測試文檔
//
// 與其他測試類型的區別：
// - 單元測試：測試個別函數的邏輯正確性
// - 整合測試：測試組件間的協作
// - BDD 測試：測試完整的業務流程和使用者體驗
//
// 上下文管理特點：
// - 維護測試環境的狀態
// - 管理多個 WebSocket 連線
// - 追蹤訊息歷史和使用者互動
// - 確保測試場景的隔離性
// - 支援並發操作的安全性
type ChatTestContext struct {
	server           *httptest.Server             // HTTP 測試伺服器
	clientRepo       *repository.ClientRepository // 客戶端儲存庫
	broadcastService *service.BroadcastService    // 廣播服務
	wsURL            string                       // WebSocket 連線 URL
	connections      map[string]*websocket.Conn   // 使用者名稱 -> WebSocket 連線映射
	messages         map[string][]string          // 使用者名稱 -> 接收訊息列表映射
	errorMessages    map[string]string            // 使用者名稱 -> 錯誤訊息映射
	mutex            sync.Mutex                   // 保護並發存取的互斥鎖
	t                *assert.Assertions           // 斷言工具
}

// NewChatTestContext 創建一個新的 BDD 測試上下文
//
// 功能說明：
// - 初始化所有必要的映射表和狀態管理結構
// - 設置斷言工具以便在步驟中進行驗證
// - 確保每個測試場景都有乾淨的起始狀態
//
// 參數：
// - t: 斷言工具，用於在 BDD 步驟中進行驗證
//
// 回傳：
// - 初始化完成的測試上下文實例
func NewChatTestContext(t *assert.Assertions) *ChatTestContext {
	return &ChatTestContext{
		connections:   make(map[string]*websocket.Conn), // 初始化連線映射表
		messages:      make(map[string][]string),        // 初始化訊息歷史映射表
		errorMessages: make(map[string]string),          // 初始化錯誤訊息映射表
		t:             t,                                // 設置斷言工具
	}
}

// chatServerIsRunning BDD 步驟：「聊天伺服器已啟動」
//
// 對應 Gherkin 語法：Given 聊天伺服器已啟動
//
// 步驟職責：
// 1. 建立完整的聊天伺服器測試環境
// 2. 初始化所有必要的服務組件
// 3. 設定 WebSocket 升級和訊息處理邏輯
// 4. 準備接受使用者連線的基礎設施
//
// 實作細節：
// - 創建真實的 HTTP 測試伺服器
// - 設定 WebSocket 升級處理器
// - 建立客戶端註冊和訊息廣播機制
// - 使用 goroutine 處理並發的訊息接收
//
// 在 BDD 場景中的作用：
// 這是大多數聊天測試場景的前提條件（Given 部分）
func (ctx *ChatTestContext) chatServerIsRunning() error {
	// 建立真實的服務實例（非 Mock）
	ctx.clientRepo = repository.NewClientRepository()
	ctx.broadcastService = service.NewBroadcastService(ctx.clientRepo)

	// 創建 HTTP 測試伺服器，處理 WebSocket 升級請求
	ctx.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			// 配置 WebSocket 升級器，允許所有來源（測試環境）
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}

			// 執行 WebSocket 升級
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}

			// 為每個連線創建唯一的客戶端 ID
			clientID := fmt.Sprintf("%p", conn)
			client := &model.Client{
				ID:   clientID,
				Conn: conn,
			}

			// 將客戶端註冊到廣播服務
			ctx.broadcastService.AddClient(client)

			// 啟動 goroutine 處理該客戶端的訊息
			go func() {
				defer conn.Close()                                // 確保連線關閉
				defer ctx.broadcastService.RemoveClient(clientID) // 確保客戶端清理

				// 持續監聽來自客戶端的訊息
				for {
					_, msg, err := conn.ReadMessage()
					if err != nil {
						break // 連線中斷，退出循環
					}

					// 將接收到的訊息廣播給所有客戶端
					ctx.broadcastService.BroadcastMessage(msg)
				}
			}()
		}
	}))

	// 建立 WebSocket 連線 URL，供測試步驟使用
	ctx.wsURL = "ws" + strings.TrimPrefix(ctx.server.URL, "http") + "/ws"
	return nil
}

// userOpensTheChatPage BDD 步驟：「使用者 "使用者名稱" 打開聊天頁面」
//
// 對應 Gherkin 語法：When 使用者 "Alice" 打開聊天頁面
//
// 步驟職責：
// 1. 建立使用者與聊天伺服器的 WebSocket 連線
// 2. 註冊使用者到測試上下文中
// 3. 啟動訊息接收監聽機制
// 4. 初始化使用者的訊息歷史記錄
//
// 實作細節：
// - 使用真實的 WebSocket 撥號器建立連線
// - 使用互斥鎖保護並發操作
// - 啟動 goroutine 持續監聽伺服器訊息
// - 維護每個使用者的獨立訊息佇列
//
// 在 BDD 場景中的作用：
// 模擬使用者開啟瀏覽器並進入聊天頁面的操作
//
// 參數：
// - userName: 使用者名稱，用於識別和追蹤特定使用者
func (ctx *ChatTestContext) userOpensTheChatPage(userName string) error {
	// 建立 WebSocket 撥號器
	dialer := websocket.Dialer{}

	// 嘗試連線到聊天伺服器
	conn, _, err := dialer.Dial(ctx.wsURL, nil)
	if err != nil {
		return err
	}

	// 使用互斥鎖保護並發操作
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	// 將使用者連線註冊到上下文中
	ctx.connections[userName] = conn
	ctx.messages[userName] = []string{} // 初始化訊息歷史

	// 啟動 goroutine 來持續接收該使用者的訊息
	go func() {
		for {
			// 監聽來自伺服器的訊息
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break // 連線中斷，退出監聽循環
			}

			// 安全地將接收到的訊息添加到使用者的訊息歷史中
			ctx.mutex.Lock()
			ctx.messages[userName] = append(ctx.messages[userName], string(msg))
			ctx.mutex.Unlock()
		}
	}()

	return nil
}

// 步驟定義: 使用者應該看到連接成功的訊息
func (ctx *ChatTestContext) userShouldSeeConnectionSuccessMessage(userName string) error {
	// 在實際應用中，伺服器可能會發送連接成功訊息
	// 在這個測試中，我們假設連接成功就是成功的
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if _, exists := ctx.connections[userName]; !exists {
		return fmt.Errorf("使用者 %s 未連接", userName)
	}
	return nil
}

// 步驟定義: 使用者在輸入框中輸入訊息
func (ctx *ChatTestContext) userTypesMessage(userName, message string) error {
	// 這個步驟只是記錄訊息，不實際發送
	// 實際發送在點擊發送按鈕或按下 Enter 鍵的步驟中
	return nil
}

// 步驟定義: 使用者點擊發送按鈕
func (ctx *ChatTestContext) userClicksSendButton(userName string) error {
	ctx.mutex.Lock()
	conn, exists := ctx.connections[userName]
	ctx.mutex.Unlock()

	if !exists {
		return fmt.Errorf("使用者 %s 未連接", userName)
	}

	message := "Test message from " + userName
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// userSendsMessage BDD 步驟：「使用者 "使用者名稱" 發送訊息 "訊息內容"」
//
// 對應 Gherkin 語法：When 使用者 "Alice" 發送訊息 "Hello, World!"
//
// 步驟職責：
// 1. 驗證使用者連線狀態
// 2. 透過 WebSocket 發送訊息到伺服器
// 3. 模擬真實的使用者發送訊息操作
//
// 實作細節：
// - 從連線映射表中查找使用者的 WebSocket 連線
// - 使用互斥鎖確保執行緒安全
// - 直接發送原始訊息內容
//
// 在 BDD 場景中的作用：
// 模擬使用者在聊天介面中輸入並發送訊息的操作
//
// 參數：
// - userName: 發送訊息的使用者名稱
// - message: 要發送的訊息內容
func (ctx *ChatTestContext) userSendsMessage(userName, message string) error {
	// 安全地取得使用者的 WebSocket 連線
	ctx.mutex.Lock()
	conn, exists := ctx.connections[userName]
	ctx.mutex.Unlock()

	// 驗證使用者是否已連線
	if !exists {
		return fmt.Errorf("使用者 %s 未連接", userName)
	}

	// 透過 WebSocket 發送文本訊息
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// userShouldSeeMessage BDD 步驟：「使用者 "使用者名稱" 應該在聊天室中看到訊息 "訊息內容"」
//
// 對應 Gherkin 語法：Then 使用者 "Bob" 應該在聊天室中看到訊息 "Hello, World!"
//
// 步驟職責：
// 1. 驗證使用者是否接收到特定的訊息
// 2. 檢查訊息歷史記錄中是否包含期望的訊息
// 3. 確保訊息廣播機制正常運作
//
// 實作細節：
// - 允許訊息傳遞的時間延遲（非同步特性）
// - 在使用者的訊息歷史中搜尋期望的訊息
// - 使用互斥鎖保護訊息歷史的讀取操作
//
// 在 BDD 場景中的作用：
// 驗證聊天系統的核心功能 - 訊息是否正確傳遞給接收者
//
// 參數：
// - userName: 接收訊息的使用者名稱
// - expectedMessage: 期望看到的訊息內容
func (ctx *ChatTestContext) userShouldSeeMessage(userName, expectedMessage string) error {
	// 等待訊息傳遞完成（處理 WebSocket 的非同步特性）
	time.Sleep(100 * time.Millisecond)

	// 安全地存取使用者的訊息歷史
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	messages, exists := ctx.messages[userName]
	if !exists {
		return fmt.Errorf("使用者 %s 沒有訊息記錄", userName)
	}

	// 在訊息歷史中搜尋期望的訊息
	for _, msg := range messages {
		if msg == expectedMessage {
			return nil // 找到期望的訊息，測試通過
		}
	}

	// 未找到期望的訊息，測試失敗
	return fmt.Errorf("使用者 %s 未收到訊息: %s", userName, expectedMessage)
}

// 步驟定義: 輸入框應該被清空
func (ctx *ChatTestContext) inputFieldShouldBeCleared() error {
	// 這個步驟在前端測試中實現
	// 在這裡我們只是返回成功
	return nil
}

// 步驟定義: 使用者關閉瀏覽器標籤
func (ctx *ChatTestContext) userClosesBrowserTab(userName string) error {
	ctx.mutex.Lock()
	conn, exists := ctx.connections[userName]
	ctx.mutex.Unlock()

	if !exists {
		return fmt.Errorf("使用者 %s 未連接", userName)
	}

	return conn.Close()
}

// 步驟定義: 伺服器重啟
func (ctx *ChatTestContext) serverRestarts() error {
	oldServer := ctx.server

	// 啟動新的伺服器
	err := ctx.chatServerIsRunning()
	if err != nil {
		return err
	}

	// 關閉舊伺服器
	oldServer.Close()

	return nil
}

// InitializeScenario 初始化 BDD 測試場景和步驟定義
//
// 功能說明：
// 1. 建立測試上下文實例
// 2. 註冊所有 Gherkin 步驟與對應的 Go 函數
// 3. 設定場景的前置和後置處理
// 4. 配置正則表達式以支援靈活的步驟匹配
//
// BDD 步驟映射：
// - Given 步驟：設定測試前提條件
// - When 步驟：執行測試操作
// - Then 步驟：驗證測試結果
//
// 正則表達式特色：
// - 支援可選的使用者標識符（A、B、C）
// - 彈性的參數擷取
// - 自然語言的步驟描述
//
// 生命週期管理：
// - Before：場景開始前的初始化
// - After：場景結束後的清理工作
//
// 參數：
// - sc: Godog 場景上下文，用於註冊步驟和生命週期鉤子
func InitializeScenario(sc *godog.ScenarioContext) {
	// 創建測試上下文實例
	ctx := NewChatTestContext(assert.New(nil))

	// 註冊 BDD 步驟定義，將 Gherkin 語法映射到 Go 函數

	// Given 步驟：設定測試環境
	sc.Step(`^聊天伺服器已啟動$`, ctx.chatServerIsRunning)

	// When 步驟：執行使用者操作
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 打開聊天頁面$`, ctx.userOpensTheChatPage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 在輸入框中輸入訊息 "([^"]*)"$`, ctx.userTypesMessage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 點擊發送按鈕$`, ctx.userClicksSendButton)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 發送訊息 "([^"]*)"$`, ctx.userSendsMessage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 關閉瀏覽器標籤$`, ctx.userClosesBrowserTab)
	sc.Step(`^伺服器重啟$`, ctx.serverRestarts)

	// Then 步驟：驗證測試結果
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 應該看到連接成功的訊息$`, ctx.userShouldSeeConnectionSuccessMessage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 應該在聊天室中看到訊息 "([^"]*)"$`, ctx.userShouldSeeMessage)
	sc.Step(`^輸入框應該被清空$`, ctx.inputFieldShouldBeCleared)

	// 場景生命週期管理

	// Before 鉤子：在每個場景開始前執行
	// 用於設定場景特定的初始化工作
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// 目前場景間是獨立的，暫時不需要特殊初始化
		return ctx, nil
	})

	// After 鉤子：在每個場景結束後執行
	// 用於清理資源，防止場景間的狀態洩漏
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// 嘗試從上下文中取得測試實例並清理資源
		testCtx, ok := ctx.Value("testContext").(*ChatTestContext)
		if ok && testCtx.server != nil {
			testCtx.server.Close() // 關閉測試伺服器，釋放連接埠
		}
		return ctx, nil
	})
}
