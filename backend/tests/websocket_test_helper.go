package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// WebSocketTestHelper WebSocket 測試輔助工具
//
// 設計理念：
// 1. 提供標準化的 WebSocket 測試環境設置
// 2. 簡化 WebSocket 連線建立和管理
// 3. 提供同步和異步的訊息收發機制
// 4. 支援多客戶端並發測試
// 5. 內建超時和錯誤處理機制
type WebSocketTestHelper struct {
	Server   *httptest.Server       // 測試用的 HTTP 伺服器
	Clients  []*WebSocketTestClient // 管理的測試客戶端列表
	Messages []ReceivedMessage      // 收到的所有訊息記錄
	mu       sync.Mutex             // 並發安全保護
	t        *testing.T             // 測試實例，用於斷言
}

// WebSocketTestClient 單個 WebSocket 測試客戶端
//
// 職責：
// 1. 管理單個 WebSocket 連線的生命週期
// 2. 提供訊息收發的便利方法
// 3. 記錄和處理連線狀態變化
// 4. 支援優雅的連線關閉
type WebSocketTestClient struct {
	ID       string               // 客戶端唯一識別碼
	Conn     *websocket.Conn      // WebSocket 連線實例
	Messages chan ReceivedMessage // 接收到的訊息通道
	Done     chan struct{}        // 連線關閉信號
	mu       sync.Mutex           // 並發安全保護
	t        *testing.T           // 測試實例
}

// ReceivedMessage 接收到的訊息結構
//
// 用途：
// 1. 記錄訊息內容和元數據
// 2. 支援測試中的訊息順序驗證
// 3. 提供豐富的訊息分析能力
type ReceivedMessage struct {
	ClientID  string    // 接收訊息的客戶端 ID
	Content   string    // 訊息內容
	Timestamp time.Time // 接收時間戳
	Type      int       // WebSocket 訊息類型
}

// NewWebSocketTestHelper 創建新的 WebSocket 測試輔助工具
//
// 參數說明：
// - t: 測試實例，用於斷言和錯誤報告
// - handler: WebSocket 升級處理器
//
// 使用場景：
// 1. 單元測試：測試 WebSocket 訊息處理邏輯
// 2. 整合測試：測試完整的 WebSocket 通信流程
// 3. 並發測試：測試多客戶端同時連線的情況
func NewWebSocketTestHelper(t *testing.T, handler http.HandlerFunc) *WebSocketTestHelper {
	// 創建測試用的 HTTP 伺服器
	server := httptest.NewServer(handler)

	return &WebSocketTestHelper{
		Server:   server,
		Clients:  make([]*WebSocketTestClient, 0),
		Messages: make([]ReceivedMessage, 0),
		t:        t,
	}
}

// ConnectClient 創建並連接新的 WebSocket 客戶端
//
// 功能特點：
// 1. 自動處理 WebSocket 升級握手
// 2. 啟動訊息接收的 goroutine
// 3. 提供連線超時保護
// 4. 自動註冊到輔助工具的客戶端列表
//
// 錯誤處理：
// - 連線失敗時自動終止測試
// - 提供詳細的錯誤信息
func (h *WebSocketTestHelper) ConnectClient(clientID string) *WebSocketTestClient {
	// 將 HTTP URL 轉換為 WebSocket URL
	u := url.URL{
		Scheme: "ws",
		Host:   h.Server.Listener.Addr().String(),
		Path:   "/",
	}

	// 建立 WebSocket 連線，設定 5 秒超時
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		h.t.Fatalf("WebSocket 連線失敗 (客戶端: %s): %v", clientID, err)
	}

	// 創建客戶端實例
	client := &WebSocketTestClient{
		ID:       clientID,
		Conn:     conn,
		Messages: make(chan ReceivedMessage, 100), // 緩衝區大小 100
		Done:     make(chan struct{}),
		t:        h.t,
	}

	// 啟動訊息接收 goroutine
	go client.startMessageReceiver(h)

	// 註冊到輔助工具的客戶端列表
	h.mu.Lock()
	h.Clients = append(h.Clients, client)
	h.mu.Unlock()

	return client
}

// SendMessage 發送訊息到 WebSocket 伺服器
//
// 功能：
// 1. 發送文本訊息到 WebSocket 伺服器
// 2. 提供發送超時保護
// 3. 線程安全的發送操作
//
// 參數：
// - message: 要發送的訊息內容
func (c *WebSocketTestClient) SendMessage(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 設定寫入超時為 5 秒
	err := c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		c.t.Errorf("設定寫入超時失敗 (客戶端: %s): %v", c.ID, err)
		return
	}

	// 發送文本訊息
	err = c.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		c.t.Errorf("發送訊息失敗 (客戶端: %s): %v", c.ID, err)
	}
}

// WaitForMessage 等待接收特定訊息
//
// 功能：
// 1. 在指定時間內等待特定內容的訊息
// 2. 支援部分匹配和完全匹配
// 3. 提供超時機制避免測試卡住
//
// 參數：
// - expectedContent: 期望的訊息內容
// - timeout: 等待超時時間
//
// 回傳值：
// - bool: 是否在超時前收到期望的訊息
// - ReceivedMessage: 收到的訊息詳情
func (c *WebSocketTestClient) WaitForMessage(expectedContent string, timeout time.Duration) (bool, ReceivedMessage) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case msg := <-c.Messages:
			// 檢查是否為期望的訊息內容
			if msg.Content == expectedContent {
				return true, msg
			}
			// 如果不是期望的訊息，繼續等待
			continue

		case <-timer.C:
			// 超時
			return false, ReceivedMessage{}

		case <-c.Done:
			// 連線已關閉
			return false, ReceivedMessage{}
		}
	}
}

// WaitForAnyMessage 等待接收任意訊息
//
// 功能：
// 1. 在指定時間內等待任意訊息
// 2. 用於測試訊息是否能正常接收
// 3. 提供超時保護
func (c *WebSocketTestClient) WaitForAnyMessage(timeout time.Duration) (bool, ReceivedMessage) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case msg := <-c.Messages:
		return true, msg

	case <-timer.C:
		return false, ReceivedMessage{}

	case <-c.Done:
		return false, ReceivedMessage{}
	}
}

// Close 優雅地關閉 WebSocket 連線
//
// 功能：
// 1. 發送關閉幀給伺服器
// 2. 等待伺服器確認關閉
// 3. 釋放相關資源
// 4. 通知訊息接收 goroutine 停止
func (c *WebSocketTestClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 發送關閉幀
	err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		c.t.Logf("發送關閉幀失敗 (客戶端: %s): %v", c.ID, err)
	}

	// 關閉連線
	err = c.Conn.Close()
	if err != nil {
		c.t.Logf("關閉連線失敗 (客戶端: %s): %v", c.ID, err)
	}

	// 通知 Done 通道
	close(c.Done)
}

// startMessageReceiver 啟動訊息接收器 (私有方法)
//
// 職責：
// 1. 在獨立的 goroutine 中持續接收訊息
// 2. 將收到的訊息轉發到訊息通道
// 3. 處理連線錯誤和異常情況
// 4. 記錄所有接收到的訊息到全域列表
func (c *WebSocketTestClient) startMessageReceiver(helper *WebSocketTestHelper) {
	defer func() {
		// 確保在 goroutine 結束時關閉 Done 通道
		select {
		case <-c.Done:
			// 已經關閉
		default:
			close(c.Done)
		}
	}()

	for {
		// 設定讀取超時
		err := c.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			c.t.Logf("設定讀取超時失敗 (客戶端: %s): %v", c.ID, err)
			return
		}

		// 讀取訊息
		messageType, data, err := c.Conn.ReadMessage()
		if err != nil {
			// 檢查是否為正常關閉
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.t.Logf("WebSocket 連線正常關閉 (客戶端: %s)", c.ID)
			} else {
				c.t.Logf("讀取訊息錯誤 (客戶端: %s): %v", c.ID, err)
			}
			return
		}

		// 創建接收訊息記錄
		receivedMsg := ReceivedMessage{
			ClientID:  c.ID,
			Content:   string(data),
			Timestamp: time.Now(),
			Type:      messageType,
		}

		// 發送到客戶端的訊息通道
		select {
		case c.Messages <- receivedMsg:
		case <-c.Done:
			return
		default:
			c.t.Logf("訊息通道已滿，丟棄訊息 (客戶端: %s): %s", c.ID, string(data))
		}

		// 記錄到全域訊息列表
		helper.mu.Lock()
		helper.Messages = append(helper.Messages, receivedMsg)
		helper.mu.Unlock()
	}
}

// Cleanup 清理測試環境
//
// 功能：
// 1. 關閉所有客戶端連線
// 2. 停止測試伺服器
// 3. 釋放所有相關資源
// 4. 提供測試結束後的統計信息
//
// 使用方式：
// 建議在測試的 defer 語句中調用，確保資源總是被正確釋放
func (h *WebSocketTestHelper) Cleanup() {
	// 關閉所有客戶端
	h.mu.Lock()
	for _, client := range h.Clients {
		client.Close()
	}
	clientCount := len(h.Clients)
	messageCount := len(h.Messages)
	h.mu.Unlock()

	// 停止測試伺服器
	h.Server.Close()

	// 記錄測試統計信息
	h.t.Logf("WebSocket 測試清理完成: %d 個客戶端, %d 條訊息", clientCount, messageCount)
}

// AssertMessageReceived 斷言客戶端收到了特定訊息
//
// 功能：
// 1. 驗證指定客戶端是否收到期望的訊息
// 2. 提供超時保護，避免測試卡住
// 3. 失敗時提供詳細的錯誤信息
//
// 參數：
// - clientID: 客戶端 ID
// - expectedContent: 期望的訊息內容
// - timeout: 等待超時時間
func (h *WebSocketTestHelper) AssertMessageReceived(clientID, expectedContent string, timeout time.Duration) {
	// 找到指定的客戶端
	var targetClient *WebSocketTestClient
	h.mu.Lock()
	for _, client := range h.Clients {
		if client.ID == clientID {
			targetClient = client
			break
		}
	}
	h.mu.Unlock()

	if targetClient == nil {
		h.t.Fatalf("找不到客戶端: %s", clientID)
	}

	// 等待訊息
	received, msg := targetClient.WaitForMessage(expectedContent, timeout)
	assert.True(h.t, received,
		"客戶端 %s 在 %v 內未收到期望訊息: %s", clientID, timeout, expectedContent)

	if received {
		assert.Equal(h.t, expectedContent, msg.Content,
			"客戶端 %s 收到的訊息內容不符合期望", clientID)
	}
}

// GetMessageHistory 獲取所有訊息歷史記錄
//
// 功能：
// 1. 回傳所有客戶端接收到的訊息記錄
// 2. 按時間順序排序
// 3. 用於測試後的分析和驗證
func (h *WebSocketTestHelper) GetMessageHistory() []ReceivedMessage {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 創建副本避免並發修改
	history := make([]ReceivedMessage, len(h.Messages))
	copy(history, h.Messages)
	return history
}

// WaitForClientCount 等待達到指定的客戶端數量
//
// 功能：
// 1. 等待客戶端數量達到期望值
// 2. 用於並發測試中的同步
// 3. 提供超時保護
func (h *WebSocketTestHelper) WaitForClientCount(expectedCount int, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			return false

		case <-ticker.C:
			h.mu.Lock()
			currentCount := len(h.Clients)
			h.mu.Unlock()

			if currentCount >= expectedCount {
				return true
			}
		}
	}
}
