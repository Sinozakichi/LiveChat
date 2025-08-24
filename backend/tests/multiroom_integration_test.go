package tests

import (
	"encoding/json"
	"livechat/backend/handler"
	"livechat/backend/model"
	"livechat/backend/repository"
	"livechat/backend/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 測試多聊天室功能的整合測試
func TestMultiRoomIntegration(t *testing.T) {
	// 設置 Gin 測試模式
	gin.SetMode(gin.TestMode)

	// 創建一個模擬的 GORM DB
	db := repository.NewMockDB()

	// 創建儲存庫
	roomRepo := repository.NewRoomRepository(db)
	clientRepo := repository.NewClientRepository()

	// 創建服務
	roomService := service.NewRoomService(roomRepo)
	broadcastService := service.NewBroadcastService(clientRepo)

	// 創建處理器
	roomHandler := handler.NewRoomHandler(roomService)
	wsHandler := handler.NewWebSocketHandler(broadcastService)

	// 創建 Gin 路由
	router := gin.New()
	roomHandler.RegisterRoutes(router)
	router.GET("/ws", func(c *gin.Context) {
		wsHandler.HandleConnection(c.Writer, c.Request)
	})

	// 測試創建聊天室
	t.Run("Create_Room", func(t *testing.T) {
		// 模擬創建聊天室的請求
		reqBody := `{"name":"測試聊天室","description":"這是一個測試聊天室","isPublic":true,"maxUsers":50}`
		req, _ := http.NewRequest("POST", "/api/rooms", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		// 模擬 DB 行為
		mockResult := new(repository.MockGormDB)
		mockResult.Err = nil
		db.On("Create", mock.AnythingOfType("*model.Room")).Return(mockResult)

		// 執行請求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 驗證結果
		assert.Equal(t, http.StatusCreated, w.Code, "應該返回 201 Created")
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "應該能夠解析響應")
		assert.Equal(t, "測試聊天室", response["name"], "聊天室名稱應該匹配")
	})

	// 測試獲取聊天室列表
	t.Run("Get_Room_List", func(t *testing.T) {
		// 模擬獲取聊天室列表的請求
		req, _ := http.NewRequest("GET", "/api/rooms", nil)

		// 模擬 DB 行為
		mockResult := new(repository.MockGormDB)
		mockResult.Err = nil
		rooms := []model.Room{
			{ID: "test-room-1", Name: "公共聊天室", Description: "這是一個公開的聊天室，所有人都可以加入", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: "test-room-2", Name: "技術討論", Description: "討論各種技術話題，包括程式設計、網絡、數據庫等", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		db.On("Find", mock.AnythingOfType("*[]model.Room"), mock.Anything).Run(func(args mock.Arguments) {
			roomsPtr := args.Get(0).(*[]model.Room)
			*roomsPtr = rooms
		}).Return(mockResult)
		db.On("Where", mock.Anything, mock.Anything, mock.Anything).Return(db)
		db.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			countPtr := args.Get(0).(*int64)
			*countPtr = 5
		}).Return(mockResult)

		// 執行請求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code, "應該返回 200 OK")
		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "應該能夠解析響應")
		assert.Equal(t, 2, len(response), "應該有 2 個聊天室")
		assert.Equal(t, "公共聊天室", response[0]["name"], "第一個聊天室的名稱應該匹配")
	})

	// 測試獲取特定聊天室
	t.Run("Get_Room", func(t *testing.T) {
		// 模擬獲取特定聊天室的請求
		req, _ := http.NewRequest("GET", "/api/rooms/1", nil)

		// 模擬 DB 行為
		mockResult := new(repository.MockGormDB)
		mockResult.Err = nil
		room := &model.Room{
			ID:          "test-room-1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "測試聊天室",
			Description: "這是一個測試聊天室",
			IsPublic:    true,
			MaxUsers:    100,
			CreatedBy:   "user-123",
		}
		db.On("First", mock.AnythingOfType("*model.Room"), mock.Anything).Run(func(args mock.Arguments) {
			roomPtr := args.Get(0).(*model.Room)
			*roomPtr = *room
		}).Return(mockResult)
		db.On("Where", mock.Anything, mock.Anything, mock.Anything).Return(db)
		db.On("Count", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
			countPtr := args.Get(0).(*int64)
			*countPtr = 5
		}).Return(mockResult)

		// 執行請求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code, "應該返回 200 OK")
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "應該能夠解析響應")
		assert.Equal(t, "測試聊天室", response["name"], "聊天室名稱應該匹配")
		assert.Equal(t, "這是一個測試聊天室", response["description"], "聊天室描述應該匹配")
	})

	// 測試 WebSocket 連接和聊天室訊息
	t.Run("WebSocket_Room_Messages", func(t *testing.T) {
		// 創建 HTTP 測試服務器
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsHandler.HandleConnection(w, r)
		}))
		defer server.Close()

		// 將 HTTP URL 轉換為 WebSocket URL
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?username=TestUser&roomId=1"

		// 連接到 WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("無法連接到 WebSocket: %v", err)
		}
		defer conn.Close()

		// 發送加入聊天室訊息
		joinMsg := map[string]interface{}{
			"type":   "join_room",
			"target": "1",
		}
		err = conn.WriteJSON(joinMsg)
		assert.NoError(t, err, "應該能夠發送加入聊天室訊息")

		// 等待一段時間，確保訊息被處理
		time.Sleep(100 * time.Millisecond)

		// 發送聊天訊息
		chatMsg := map[string]interface{}{
			"content": "Hello, Room!",
			"sender":  "TestUser",
			"time":    time.Now().Unix(),
		}
		err = conn.WriteJSON(chatMsg)
		assert.NoError(t, err, "應該能夠發送聊天訊息")

		// 等待一段時間，確保訊息被處理
		time.Sleep(100 * time.Millisecond)

		// 發送離開聊天室訊息
		leaveMsg := map[string]interface{}{
			"type": "leave_room",
		}
		err = conn.WriteJSON(leaveMsg)
		assert.NoError(t, err, "應該能夠發送離開聊天室訊息")

		// 等待一段時間，確保訊息被處理
		time.Sleep(100 * time.Millisecond)
	})

	// 測試多個用戶在不同聊天室
	t.Run("Multiple_Users_Different_Rooms", func(t *testing.T) {
		// 創建一個新的客戶端儲存庫和廣播服務，以避免與其他測試衝突
		clientRepo := repository.NewClientRepository()
		broadcastService := service.NewBroadcastService(clientRepo)
		wsHandler := handler.NewWebSocketHandler(broadcastService)

		// 創建 HTTP 測試服務器
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			wsHandler.HandleConnection(w, r)
		}))
		defer server.Close()

		// 用戶 1 連接到聊天室 1
		wsURL1 := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?username=User1&roomId=1"
		conn1, _, err := websocket.DefaultDialer.Dial(wsURL1, nil)
		if err != nil {
			t.Fatalf("用戶 1 無法連接到 WebSocket: %v", err)
		}
		defer conn1.Close()

		// 用戶 2 連接到聊天室 2
		wsURL2 := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?username=User2&roomId=2"
		conn2, _, err := websocket.DefaultDialer.Dial(wsURL2, nil)
		if err != nil {
			t.Fatalf("用戶 2 無法連接到 WebSocket: %v", err)
		}
		defer conn2.Close()

		// 用戶 1 發送加入聊天室訊息
		joinMsg1 := map[string]interface{}{
			"type":   "join_room",
			"target": "1",
		}
		err = conn1.WriteJSON(joinMsg1)
		assert.NoError(t, err, "用戶 1 應該能夠發送加入聊天室訊息")

		// 用戶 2 發送加入聊天室訊息
		joinMsg2 := map[string]interface{}{
			"type":   "join_room",
			"target": "2",
		}
		err = conn2.WriteJSON(joinMsg2)
		assert.NoError(t, err, "用戶 2 應該能夠發送加入聊天室訊息")

		// 等待一段時間，確保訊息被處理
		time.Sleep(100 * time.Millisecond)

		// 用戶 1 發送聊天訊息到聊天室 1
		chatMsg1 := map[string]interface{}{
			"content": "Hello from Room 1!",
			"sender":  "User1",
			"time":    time.Now().Unix(),
		}
		err = conn1.WriteJSON(chatMsg1)
		assert.NoError(t, err, "用戶 1 應該能夠發送聊天訊息")

		// 用戶 2 發送聊天訊息到聊天室 2
		chatMsg2 := map[string]interface{}{
			"content": "Hello from Room 2!",
			"sender":  "User2",
			"time":    time.Now().Unix(),
		}
		err = conn2.WriteJSON(chatMsg2)
		assert.NoError(t, err, "用戶 2 應該能夠發送聊天訊息")

		// 等待一段時間，確保訊息被處理
		time.Sleep(100 * time.Millisecond)

		// 驗證客戶端數量
		clients := clientRepo.GetActiveClients()
		assert.Equal(t, 2, len(clients), "應該有 2 個活躍客戶端")

		// 驗證聊天室分配
		roomClients := make(map[string]int)
		for _, client := range clients {
			roomClients[client.RoomID]++
		}
		assert.Equal(t, 1, roomClients["1"], "聊天室 1 應該有 1 個客戶端")
		assert.Equal(t, 1, roomClients["2"], "聊天室 2 應該有 1 個客戶端")
	})
}
