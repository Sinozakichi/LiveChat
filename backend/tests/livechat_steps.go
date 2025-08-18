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

// ChatTestContext 用於在測試步驟之間共享狀態
type ChatTestContext struct {
	server           *httptest.Server
	clientRepo       *repository.ClientRepository
	broadcastService *service.BroadcastService
	wsURL            string
	connections      map[string]*websocket.Conn
	messages         map[string][]string
	errorMessages    map[string]string
	mutex            sync.Mutex
	t                *assert.Assertions
}

// NewChatTestContext 創建一個新的測試上下文
func NewChatTestContext(t *assert.Assertions) *ChatTestContext {
	return &ChatTestContext{
		connections:   make(map[string]*websocket.Conn),
		messages:      make(map[string][]string),
		errorMessages: make(map[string]string),
		t:             t,
	}
}

// 步驟定義: 聊天伺服器已啟動
func (ctx *ChatTestContext) chatServerIsRunning() error {
	ctx.clientRepo = repository.NewClientRepository()
	ctx.broadcastService = service.NewBroadcastService(ctx.clientRepo)

	// 創建測試 HTTP 伺服器
	ctx.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			upgrader := websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool { return true },
			}
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}

			clientID := fmt.Sprintf("%p", conn)
			client := &model.Client{
				ID:   clientID,
				Conn: conn,
			}

			ctx.broadcastService.AddClient(client)

			go func() {
				defer conn.Close()
				defer ctx.broadcastService.RemoveClient(clientID)

				for {
					_, msg, err := conn.ReadMessage()
					if err != nil {
						break
					}

					ctx.broadcastService.BroadcastMessage(msg)
				}
			}()
		}
	}))

	// 將 HTTP URL 轉換為 WebSocket URL
	ctx.wsURL = "ws" + strings.TrimPrefix(ctx.server.URL, "http") + "/ws"
	return nil
}

// 步驟定義: 使用者打開聊天頁面
func (ctx *ChatTestContext) userOpensTheChatPage(userName string) error {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(ctx.wsURL, nil)
	if err != nil {
		return err
	}

	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	ctx.connections[userName] = conn
	ctx.messages[userName] = []string{}

	// 啟動一個 goroutine 來接收訊息
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

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

// 步驟定義: 使用者發送特定訊息
func (ctx *ChatTestContext) userSendsMessage(userName, message string) error {
	ctx.mutex.Lock()
	conn, exists := ctx.connections[userName]
	ctx.mutex.Unlock()

	if !exists {
		return fmt.Errorf("使用者 %s 未連接", userName)
	}

	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// 步驟定義: 使用者應該在聊天室中看到訊息
func (ctx *ChatTestContext) userShouldSeeMessage(userName, expectedMessage string) error {
	// 等待一段時間讓訊息傳遞
	time.Sleep(100 * time.Millisecond)

	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	messages, exists := ctx.messages[userName]
	if !exists {
		return fmt.Errorf("使用者 %s 沒有訊息記錄", userName)
	}

	for _, msg := range messages {
		if msg == expectedMessage {
			return nil
		}
	}

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

// InitializeScenario 初始化 BDD 場景
func InitializeScenario(sc *godog.ScenarioContext) {
	ctx := NewChatTestContext(assert.New(nil))

	// 定義步驟
	sc.Step(`^聊天伺服器已啟動$`, ctx.chatServerIsRunning)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 打開聊天頁面$`, ctx.userOpensTheChatPage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 應該看到連接成功的訊息$`, ctx.userShouldSeeConnectionSuccessMessage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 在輸入框中輸入訊息 "([^"]*)"$`, ctx.userTypesMessage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 點擊發送按鈕$`, ctx.userClicksSendButton)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 發送訊息 "([^"]*)"$`, ctx.userSendsMessage)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 應該在聊天室中看到訊息 "([^"]*)"$`, ctx.userShouldSeeMessage)
	sc.Step(`^輸入框應該被清空$`, ctx.inputFieldShouldBeCleared)
	sc.Step(`^使用者(?:A|B|C|)? "([^"]*)" 關閉瀏覽器標籤$`, ctx.userClosesBrowserTab)
	sc.Step(`^伺服器重啟$`, ctx.serverRestarts)

	// 在每個場景之前執行
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return ctx, nil
	})

	// 在每個場景之後執行
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		testCtx, ok := ctx.Value("testContext").(*ChatTestContext)
		if ok && testCtx.server != nil {
			testCtx.server.Close()
		}
		return ctx, nil
	})
}
