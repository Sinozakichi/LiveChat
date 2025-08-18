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

// 整合測試：測試完整的聊天流程
func TestChatIntegration(t *testing.T) {
	// 安排 (Arrange)
	clientRepo := repository.NewClientRepository()
	broadcastService := service.NewBroadcastService(clientRepo)
	wsHandler := handler.NewWebSocketHandler(broadcastService)

	// 創建測試伺服器
	server := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	defer server.Close()

	// 將 HTTP URL 轉換為 WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// 測試：用戶連接到聊天室
	t.Run("User connects to chat room", func(t *testing.T) {
		// 建立 WebSocket 連接
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		assert.NoError(t, err, "應該能夠連接到 WebSocket 伺服器")
		defer conn.Close()

		// 發送訊息
		message := "Hello, World!"
		err = conn.WriteMessage(websocket.TextMessage, []byte(message))
		assert.NoError(t, err, "應該能夠發送訊息")

		// 讀取回應（自己的訊息）
		_, receivedMsg, err := conn.ReadMessage()
		assert.NoError(t, err, "應該能夠接收訊息")
		assert.Equal(t, message, string(receivedMsg), "接收到的訊息應該與發送的訊息相同")
	})

	// 測試：多個用戶交流
	t.Run("Multiple users exchange messages", func(t *testing.T) {
		// 創建新的測試環境，避免與其他測試衝突
		localClientRepo := repository.NewClientRepository()
		localBroadcastService := service.NewBroadcastService(localClientRepo)
		localWsHandler := handler.NewWebSocketHandler(localBroadcastService)
		localServer := httptest.NewServer(http.HandlerFunc(localWsHandler.HandleConnection))
		defer localServer.Close()
		localWsURL := "ws" + strings.TrimPrefix(localServer.URL, "http")

		// 建立多個 WebSocket 連接
		dialer := websocket.Dialer{}
		conn1, _, err1 := dialer.Dial(localWsURL, nil)
		assert.NoError(t, err1, "用戶1應該能夠連接")
		defer conn1.Close()

		// 用戶1發送訊息
		message1 := "Message from user 1"
		err := conn1.WriteMessage(websocket.TextMessage, []byte(message1))
		assert.NoError(t, err, "用戶1應該能夠發送訊息")

		// 用戶1接收自己的訊息
		_, receivedMsg1, err := conn1.ReadMessage()
		assert.NoError(t, err, "用戶1應該能夠接收自己的訊息")
		assert.Equal(t, message1, string(receivedMsg1), "用戶1應該能夠收到自己的訊息")

		// 連接用戶2
		conn2, _, err2 := dialer.Dial(localWsURL, nil)
		assert.NoError(t, err2, "用戶2應該能夠連接")
		defer conn2.Close()

		// 用戶2發送訊息
		message2 := "Message from user 2"
		err = conn2.WriteMessage(websocket.TextMessage, []byte(message2))
		assert.NoError(t, err, "用戶2應該能夠發送訊息")

		// 用戶1應該收到用戶2的訊息
		_, receivedMsg2, err := conn1.ReadMessage()
		assert.NoError(t, err, "用戶1應該能夠接收訊息")
		assert.Equal(t, message2, string(receivedMsg2), "用戶1收到的訊息應該與用戶2發送的訊息相同")

		// 用戶2接收自己的訊息
		_, receivedMsg3, err := conn2.ReadMessage()
		assert.NoError(t, err, "用戶2應該能夠接收自己的訊息")
		assert.Equal(t, message2, string(receivedMsg3), "用戶2應該能夠收到自己的訊息")
	})

	// 測試：發送 JSON 格式的訊息
	t.Run("Send JSON formatted message", func(t *testing.T) {
		// 創建新的測試環境，避免與其他測試衝突
		localClientRepo := repository.NewClientRepository()
		localBroadcastService := service.NewBroadcastService(localClientRepo)
		localWsHandler := handler.NewWebSocketHandler(localBroadcastService)
		localServer := httptest.NewServer(http.HandlerFunc(localWsHandler.HandleConnection))
		defer localServer.Close()
		localWsURL := "ws" + strings.TrimPrefix(localServer.URL, "http")

		// 建立 WebSocket 連接
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(localWsURL, nil)
		assert.NoError(t, err, "應該能夠連接到 WebSocket 伺服器")
		defer conn.Close()

		// 創建 JSON 訊息
		jsonMsg := map[string]interface{}{
			"type":    "message",
			"content": "JSON message",
			"time":    time.Now().Unix(),
		}
		jsonBytes, _ := json.Marshal(jsonMsg)

		// 發送 JSON 訊息
		err = conn.WriteMessage(websocket.TextMessage, jsonBytes)
		assert.NoError(t, err, "應該能夠發送 JSON 訊息")

		// 讀取回應
		_, receivedMsg, err := conn.ReadMessage()
		assert.NoError(t, err, "應該能夠接收訊息")

		// 解析接收到的 JSON
		var receivedJSON map[string]interface{}
		err = json.Unmarshal(receivedMsg, &receivedJSON)
		assert.NoError(t, err, "應該能夠解析接收到的 JSON")
		assert.Equal(t, "message", receivedJSON["type"], "訊息類型應該匹配")
		assert.Equal(t, "JSON message", receivedJSON["content"], "訊息內容應該匹配")
	})
}

// 整合測試：測試用戶斷開連接
func TestUserDisconnection(t *testing.T) {
	// 安排 (Arrange)
	clientRepo := repository.NewClientRepository()
	broadcastService := service.NewBroadcastService(clientRepo)
	wsHandler := handler.NewWebSocketHandler(broadcastService)

	// 創建測試伺服器
	server := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	defer server.Close()

	// 將 HTTP URL 轉換為 WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// 建立 WebSocket 連接
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err, "應該能夠連接到 WebSocket 伺服器")

	// 發送訊息
	message := "Hello before disconnect"
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	assert.NoError(t, err, "應該能夠發送訊息")

	// 讀取回應
	_, _, err = conn.ReadMessage()
	assert.NoError(t, err, "應該能夠接收訊息")

	// 關閉連接
	conn.Close()

	// 等待一段時間讓伺服器處理斷開連接
	time.Sleep(100 * time.Millisecond)

	// 驗證客戶端已從儲存庫中移除
	assert.Equal(t, 0, clientRepo.Count(), "斷開連接後儲存庫應該是空的")
}

// 整合測試：測試伺服器重啟
func TestServerRestart(t *testing.T) {
	// 創建一個新的儲存庫，專門用於這個測試
	clientRepo := repository.NewClientRepository()
	broadcastService := service.NewBroadcastService(clientRepo)
	wsHandler := handler.NewWebSocketHandler(broadcastService)

	// 創建第一個測試伺服器
	server1 := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	wsURL1 := "ws" + strings.TrimPrefix(server1.URL, "http")

	// 建立 WebSocket 連接
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL1, nil)
	assert.NoError(t, err, "應該能夠連接到第一個伺服器")

	// 發送訊息到第一個伺服器
	message1 := "Hello before restart"
	err = conn.WriteMessage(websocket.TextMessage, []byte(message1))
	assert.NoError(t, err, "應該能夠發送訊息到第一個伺服器")

	// 讀取回應
	_, receivedMsg1, err := conn.ReadMessage()
	assert.NoError(t, err, "應該能夠從第一個伺服器接收訊息")
	assert.Equal(t, message1, string(receivedMsg1), "接收到的訊息應該與發送的訊息相同")

	// 關閉第一個伺服器（模擬伺服器重啟）
	server1.Close()

	// 關閉連接
	conn.Close()

	// 等待一段時間，確保連接已經關閉
	time.Sleep(100 * time.Millisecond)

	// 清空儲存庫，模擬伺服器重啟後的狀態
	clientRepo.Clear()

	// 創建第二個測試伺服器（模擬重啟後的伺服器）
	server2 := httptest.NewServer(http.HandlerFunc(wsHandler.HandleConnection))
	defer server2.Close()
	wsURL2 := "ws" + strings.TrimPrefix(server2.URL, "http")

	// 嘗試連接到新的伺服器
	conn2, _, err := dialer.Dial(wsURL2, nil)
	assert.NoError(t, err, "應該能夠連接到重啟後的伺服器")
	defer conn2.Close()

	// 發送訊息到新伺服器
	message2 := "Hello after restart"
	err = conn2.WriteMessage(websocket.TextMessage, []byte(message2))
	assert.NoError(t, err, "應該能夠在重啟後發送訊息")

	// 讀取回應
	_, receivedMsg2, err := conn2.ReadMessage()
	assert.NoError(t, err, "應該能夠在重啟後接收訊息")
	assert.Equal(t, message2, string(receivedMsg2), "接收到的訊息應該與發送的訊息相同")

	// 驗證客戶端數量
	assert.Equal(t, 1, clientRepo.Count(), "重啟後應該只有一個客戶端連接")
}
