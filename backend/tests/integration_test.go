package tests

import (
	"encoding/json"
	"livechat/backend/handler"
	"livechat/backend/repository"
	"livechat/backend/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// TestChatIntegration 完整的聊天系統整合測試
//
// 整合測試目標：
// 1. 驗證完整的 WebSocket 聊天流程，從連線到訊息傳送
// 2. 測試真實的組件協作：Repository -> Service -> Handler
// 3. 確保 WebSocket 升級和通信的正確性
// 4. 驗證訊息廣播機制的端到端功能
//
// 整合測試策略：
// - 使用真實的組件實例而非 Mock（除了資料庫）
// - 測試真實的 WebSocket 連線和通信
// - 驗證多個測試場景：單用戶、多用戶、JSON 訊息
// - 模擬真實的使用者互動流程
//
// 與單元測試的區別：
// - 單元測試：測試個別組件的邏輯正確性
// - 整合測試：測試組件間的協作和完整流程
func TestChatIntegration(t *testing.T) {
	// 安排 (Arrange)：建立真實的服務鏈，模擬生產環境
	clientRepo := repository.NewClientRepository()              // 真實的客戶端儲存庫
	broadcastService := service.NewBroadcastService(clientRepo) // 真實的廣播服務
	wsHandler := handler.NewWebSocketHandler(broadcastService)  // 真實的 WebSocket 處理器

	// 創建 HTTP 測試伺服器，模擬真實的 WebSocket 服務器環境
	server := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	defer server.Close()

	// 將 HTTP URL 轉換為 WebSocket URL，以便客戶端連線
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// 整合測試場景 1：單一使用者的基本連線和訊息傳送
	//
	// 測試目標：
	// - 驗證 WebSocket 連線建立的完整流程
	// - 測試訊息的端到端傳輸
	// - 確保伺服器能正確處理和回傳訊息
	// - 驗證基本的聊天功能
	t.Run("User connects to chat room", func(t *testing.T) {
		// 建立真實的 WebSocket 連線（非 Mock）
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		assert.NoError(t, err, "應該能夠連接到 WebSocket 伺服器")
		defer conn.Close()

		// 測試訊息發送：模擬使用者輸入訊息
		message := "Hello, World!"
		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		assert.NoError(t, err, "應該能夠發送訊息")

		// 測試訊息接收：驗證廣播機制的正確性
		// 在單使用者情況下，使用者會收到自己發送的訊息
		_, receivedMsg, err := conn.ReadMessage()
		assert.NoError(t, err, "應該能夠接收訊息")
		assert.Equal(t, message, string(receivedMsg), "接收到的訊息應該與發送的訊息相同")
	})

	// 整合測試場景 2：多使用者即時訊息交流
	//
	// 測試目標：
	// - 驗證多使用者同時連線的穩定性
	// - 測試訊息廣播的正確性：一個使用者發送，所有使用者接收
	// - 確保使用者間的訊息不會混淆或遺失
	// - 驗證並發連線的處理能力
	//
	// 重要測試點：
	// - 使用獨立的測試環境避免狀態污染
	// - 測試訊息的雙向傳播
	// - 驗證每個使用者都能收到其他使用者的訊息
	t.Run("Multiple users exchange messages", func(t *testing.T) {
		// 創建獨立的測試環境，確保與其他測試案例完全隔離
		localClientRepo := repository.NewClientRepository()
		localBroadcastService := service.NewBroadcastService(localClientRepo)
		localWsHandler := handler.NewWebSocketHandler(localBroadcastService)
		localServer := httptest.NewServer(http.HandlerFunc(localWsHandler.HandleConnection))
		defer localServer.Close()
		localWsURL := "ws" + strings.TrimPrefix(localServer.URL, "http")

		// 階段 1：建立第一個使用者連線
		dialer := websocket.Dialer{}
		conn1, _, err1 := dialer.Dial(localWsURL, nil)
		assert.NoError(t, err1, "用戶1應該能夠連接")
		defer conn1.Close()

		// 測試第一個使用者的訊息發送和接收
		message1 := "Message from user 1"
		err := conn1.WriteMessage(websocket.TextMessage, []byte(message1))
		assert.NoError(t, err, "用戶1應該能夠發送訊息")

		// 驗證使用者能接收到自己的訊息（確認廣播機制運作）
		_, receivedMsg1, err := conn1.ReadMessage()
		assert.NoError(t, err, "用戶1應該能夠接收自己的訊息")
		assert.Equal(t, message1, string(receivedMsg1), "用戶1應該能夠收到自己的訊息")

		// 階段 2：第二個使用者加入聊天
		conn2, _, err2 := dialer.Dial(localWsURL, nil)
		assert.NoError(t, err2, "用戶2應該能夠連接")
		defer conn2.Close()

		// 階段 3：測試跨使用者的訊息廣播
		message2 := "Message from user 2"
		err = conn2.WriteMessage(websocket.TextMessage, []byte(message2))
		assert.NoError(t, err, "用戶2應該能夠發送訊息")

		// 關鍵測試：用戶1應該能接收到用戶2的訊息
		_, receivedMsg2, err := conn1.ReadMessage()
		assert.NoError(t, err, "用戶1應該能夠接收訊息")
		assert.Equal(t, message2, string(receivedMsg2), "用戶1收到的訊息應該與用戶2發送的訊息相同")

		// 驗證用戶2也能接收到自己的訊息
		_, receivedMsg3, err := conn2.ReadMessage()
		assert.NoError(t, err, "用戶2應該能夠接收自己的訊息")
		assert.Equal(t, message2, string(receivedMsg3), "用戶2應該能夠收到自己的訊息")
	})

	// 整合測試場景 3：JSON 格式訊息的完整處理
	//
	// 測試目標：
	// - 驗證系統對結構化 JSON 訊息的處理能力
	// - 測試 JSON 序列化和反序列化的正確性
	// - 確保複雜格式的訊息能正確傳輸和保持結構
	// - 驗證前端和後端之間的資料格式相容性
	//
	// 實際應用場景：
	// - 前端發送包含 metadata 的訊息
	// - 系統訊息和命令的傳輸
	// - 富文本或多媒體訊息的處理
	t.Run("Send JSON formatted message", func(t *testing.T) {
		// 創建獨立的測試環境
		localClientRepo := repository.NewClientRepository()
		localBroadcastService := service.NewBroadcastService(localClientRepo)
		localWsHandler := handler.NewWebSocketHandler(localBroadcastService)
		localServer := httptest.NewServer(http.HandlerFunc(localWsHandler.HandleConnection))
		defer localServer.Close()
		localWsURL := "ws" + strings.TrimPrefix(localServer.URL, "http")

		// 建立 WebSocket 連線
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(localWsURL, nil)
		assert.NoError(t, err, "應該能夠連接到 WebSocket 伺服器")
		defer conn.Close()

		// 創建結構化的 JSON 訊息，模擬真實的前端訊息格式
		jsonMsg := map[string]interface{}{
			"type":    "message",         // 訊息類型
			"content": "JSON message",    // 訊息內容
			"time":    time.Now().Unix(), // 時間戳
		}
		jsonBytes, _ := json.Marshal(jsonMsg)

		// 測試 JSON 訊息的發送
		err = conn.WriteMessage(websocket.TextMessage, jsonBytes)
		assert.NoError(t, err, "應該能夠發送 JSON 訊息")

		// 測試 JSON 訊息的接收
		_, receivedMsg, err := conn.ReadMessage()
		assert.NoError(t, err, "應該能夠接收訊息")

		// 關鍵測試：驗證 JSON 結構的完整性
		var receivedJSON map[string]interface{}
		err = json.Unmarshal(receivedMsg, &receivedJSON)
		assert.NoError(t, err, "應該能夠解析接收到的 JSON")

		// 驗證 JSON 欄位的正確性
		assert.Equal(t, "message", receivedJSON["type"], "訊息類型應該匹配")
		assert.Equal(t, "JSON message", receivedJSON["content"], "訊息內容應該匹配")
	})
}

// TestUserDisconnection 測試使用者斷線的完整處理流程
//
// 整合測試目標：
// 1. 驗證使用者正常斷線時的資源清理
// 2. 測試客戶端儲存庫的狀態一致性
// 3. 確保斷線不會造成記憶體洩漏
// 4. 驗證連線生命週期的完整管理
//
// 重要測試點：
// - 連線建立 -> 訊息傳輸 -> 正常斷線 -> 資源清理
// - 驗證斷線後客戶端是否從活躍列表中移除
// - 測試系統的自動清理機制
// - 確保斷線處理的非同步性質
func TestUserDisconnection(t *testing.T) {
	// 安排 (Arrange)：建立完整的測試環境
	clientRepo := repository.NewClientRepository()
	broadcastService := service.NewBroadcastService(clientRepo)
	wsHandler := handler.NewWebSocketHandler(broadcastService)

	// 創建測試伺服器
	server := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	defer server.Close()

	// 將 HTTP URL 轉換為 WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// 階段 1：建立連線並驗證基本功能
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err, "應該能夠連接到 WebSocket 伺服器")

	// 階段 2：測試連線狀態下的正常操作
	message := "Hello before disconnect"
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	assert.NoError(t, err, "應該能夠發送訊息")

	// 確認訊息能正常接收，驗證連線確實有效
	_, _, err = conn.ReadMessage()
	assert.NoError(t, err, "應該能夠接收訊息")

	// 階段 3：主動關閉連線，模擬使用者斷線
	conn.Close()

	// 階段 4：等待伺服器的非同步清理處理完成
	// WebSocket 斷線處理通常是非同步的，需要給系統時間清理資源
	time.Sleep(100 * time.Millisecond)

	// 階段 5：驗證資源清理的完整性
	// 這是整合測試的關鍵：確保斷線後客戶端從儲存庫中完全移除
	assert.Equal(t, 0, clientRepo.Count(), "斷開連接後儲存庫應該是空的")
}

// TestServerRestart 測試伺服器重啟情況下的系統恢復能力
//
// 整合測試目標：
// 1. 驗證伺服器重啟後系統能正常恢復服務
// 2. 測試狀態重置和資源清理的正確性
// 3. 確保重啟後新連線能正常建立和運作
// 4. 驗證系統的容錯和恢復機制
//
// 測試場景：
// - 模擬生產環境中的伺服器重啟情況
// - 驗證重啟前後的功能一致性
// - 測試狀態清理和重新初始化
// - 確保重啟不會留下殘留狀態
//
// 重要測試點：
// - 重啟前的正常運作 -> 伺服器關閉 -> 狀態清理 -> 重新啟動 -> 功能恢復
func TestServerRestart(t *testing.T) {
	// 安排 (Arrange)：建立專用的測試環境
	clientRepo := repository.NewClientRepository()
	broadcastService := service.NewBroadcastService(clientRepo)
	wsHandler := handler.NewWebSocketHandler(broadcastService)

	// 階段 1：建立第一個伺服器實例（重啟前）
	server1 := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	wsURL1 := "ws" + strings.TrimPrefix(server1.URL, "http")

	// 階段 2：測試重啟前的正常運作
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL1, nil)
	assert.NoError(t, err, "應該能夠連接到第一個伺服器")

	// 驗證重啟前的基本功能
	message1 := "Hello before restart"
	err = conn.WriteMessage(websocket.TextMessage, []byte(message1))
	assert.NoError(t, err, "應該能夠發送訊息到第一個伺服器")

	_, receivedMsg1, err := conn.ReadMessage()
	assert.NoError(t, err, "應該能夠從第一個伺服器接收訊息")
	assert.Equal(t, message1, string(receivedMsg1), "接收到的訊息應該與發送的訊息相同")

	// 階段 3：模擬伺服器重啟過程
	server1.Close() // 關閉第一個伺服器，模擬伺服器停止
	conn.Close()    // 關閉客戶端連線

	// 等待連線完全關閉和資源釋放
	time.Sleep(100 * time.Millisecond)

	// 模擬伺服器重啟後的狀態重置
	clientRepo.Clear() // 清空客戶端儲存庫，模擬重啟後的乾淨狀態

	// 階段 4：建立重啟後的新伺服器實例
	server2 := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	defer server2.Close()
	wsURL2 := "ws" + strings.TrimPrefix(server2.URL, "http")

	// 階段 5：測試重啟後的功能恢復
	conn2, _, err := dialer.Dial(wsURL2, nil)
	assert.NoError(t, err, "應該能夠連接到重啟後的伺服器")
	defer conn2.Close()

	// 驗證重啟後的基本功能完全恢復
	message2 := "Hello after restart"
	err = conn2.WriteMessage(websocket.TextMessage, []byte(message2))
	assert.NoError(t, err, "應該能夠在重啟後發送訊息")

	_, receivedMsg2, err := conn2.ReadMessage()
	assert.NoError(t, err, "應該能夠在重啟後接收訊息")
	assert.Equal(t, message2, string(receivedMsg2), "接收到的訊息應該與發送的訊息相同")

	// 階段 6：驗證重啟後的狀態正確性
	assert.Equal(t, 1, clientRepo.Count(), "重啟後應該只有一個客戶端連接")
}
