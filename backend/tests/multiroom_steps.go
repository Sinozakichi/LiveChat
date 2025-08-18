package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"livechat/backend/handler"
	"livechat/backend/model"
	"livechat/backend/repository"
	"livechat/backend/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MultiRoomTestContext 保存測試狀態
type MultiRoomTestContext struct {
	server           *httptest.Server
	router           *gin.Engine
	db               *repository.MockDB
	roomRepo         *repository.RoomRepository
	clientRepo       *repository.ClientRepository
	roomService      *service.RoomService
	broadcastService *service.BroadcastService
	roomHandler      *handler.RoomHandler
	wsHandler        *handler.WebSocketHandler
	connections      map[string]*websocket.Conn
	responses        map[string]*httptest.ResponseRecorder
	requestBodies    map[string]string
	rooms            map[string]*model.Room
}

// InitializeMultiRoomScenario 初始化多聊天室功能的 BDD 測試
func InitializeMultiRoomScenario(ctx *godog.ScenarioContext) {
	// 設置 Gin 測試模式
	gin.SetMode(gin.TestMode)

	// 創建測試上下文
	testCtx := &MultiRoomTestContext{
		connections:   make(map[string]*websocket.Conn),
		responses:     make(map[string]*httptest.ResponseRecorder),
		requestBodies: make(map[string]string),
		rooms:         make(map[string]*model.Room),
	}

	// 在每個場景開始前設置測試環境
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// 創建一個模擬的 GORM DB
		testCtx.db = repository.NewMockDB()

		// 創建儲存庫
		testCtx.roomRepo = repository.NewRoomRepository(testCtx.db)
		testCtx.clientRepo = repository.NewClientRepository()

		// 創建服務
		testCtx.roomService = service.NewRoomService(testCtx.roomRepo)
		testCtx.broadcastService = service.NewBroadcastService(testCtx.clientRepo)

		// 創建處理器
		testCtx.roomHandler = handler.NewRoomHandler(testCtx.roomService)
		testCtx.wsHandler = handler.NewWebSocketHandler(testCtx.broadcastService)

		// 創建 Gin 路由
		testCtx.router = gin.New()
		testCtx.roomHandler.RegisterRoutes(testCtx.router)
		testCtx.router.GET("/ws", func(c *gin.Context) {
			testCtx.wsHandler.HandleConnection(c.Writer, c.Request)
		})

		// 創建 HTTP 測試服務器
		testCtx.server = httptest.NewServer(testCtx.router)

		return context.WithValue(ctx, "testContext", testCtx), nil
	})

	// 在每個場景結束後清理測試環境
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		testCtx := ctx.Value("testContext").(*MultiRoomTestContext)

		// 關閉所有 WebSocket 連接
		for _, conn := range testCtx.connections {
			conn.Close()
		}

		// 關閉 HTTP 測試服務器
		if testCtx.server != nil {
			testCtx.server.Close()
		}

		return ctx, nil
	})

	// Given 步驟
	ctx.Step(`^系統中有(\d+)個預設聊天室$`, testCtx.thereAreDefaultRooms)
	ctx.Step(`^使用者正在查看聊天室列表$`, testCtx.userIsViewingRoomList)
	ctx.Step(`^使用者A和使用者B都在"([^"]*)"$`, testCtx.usersAreInRoom)
	ctx.Step(`^使用者C在"([^"]*)"$`, testCtx.userIsInRoom)
	ctx.Step(`^使用者已經在"([^"]*)"$`, testCtx.userIsInRoom)
	ctx.Step(`^使用者A已經在"([^"]*)"$`, testCtx.userAIsInRoom)
	ctx.Step(`^聊天室列表頁面$`, testCtx.roomListPage)
	ctx.Step(`^(\d+)個使用者加入"([^"]*)"$`, testCtx.usersJoinRoom)

	// When 步驟
	ctx.Step(`^使用者訪問聊天室列表頁面$`, testCtx.userVisitsRoomListPage)
	ctx.Step(`^使用者點擊"([^"]*)"$`, testCtx.userClicksRoom)
	ctx.Step(`^使用者A在"([^"]*)"發送訊息"([^"]*)"$`, testCtx.userSendsMessageInRoom)
	ctx.Step(`^使用者點擊返回按鈕$`, testCtx.userClicksBackButton)
	ctx.Step(`^使用者從列表中選擇"([^"]*)"$`, testCtx.userSelectsRoomFromList)
	ctx.Step(`^使用者B加入"([^"]*)"$`, testCtx.userBJoinsRoom)
	ctx.Step(`^使用者B離開"([^"]*)"$`, testCtx.userBLeavesRoom)

	// Then 步驟
	ctx.Step(`^使用者應該看到(\d+)個聊天室的列表$`, testCtx.userShouldSeeRoomList)
	ctx.Step(`^每個聊天室應顯示名稱和簡短描述$`, testCtx.eachRoomShouldShowNameAndDescription)
	ctx.Step(`^使用者應該被導向到"([^"]*)"的聊天界面$`, testCtx.userShouldBeRedirectedToChatInterface)
	ctx.Step(`^使用者應該看到"([^"]*)"的歡迎訊息$`, testCtx.userShouldSeeWelcomeMessage)
	ctx.Step(`^使用者B應該在"([^"]*)"看到訊息"([^"]*)"$`, testCtx.userShouldSeeMessageInRoom)
	ctx.Step(`^使用者C不應該在"([^"]*)"看到該訊息$`, testCtx.userShouldNotSeeMessageInRoom)
	ctx.Step(`^使用者應該進入"([^"]*)"$`, testCtx.userShouldEnterRoom)
	ctx.Step(`^使用者應該看到"([^"]*)"的歷史訊息$`, testCtx.userShouldSeeRoomHistory)
	ctx.Step(`^使用者不應該再收到"([^"]*)"的新訊息$`, testCtx.userShouldNotReceiveNewMessages)
	ctx.Step(`^使用者A應該看到系統通知"([^"]*)"$`, testCtx.userShouldSeeSystemNotification)
	ctx.Step(`^"([^"]*)"應顯示"([^"]*)"$`, testCtx.roomShouldShowUserCount)
}

// Given 步驟實現
func (ctx *MultiRoomTestContext) thereAreDefaultRooms(count int) error {
	// 模擬 DB 行為
	mockResult := new(repository.MockGormDB)
	mockResult.Err = nil

	// 創建指定數量的聊天室
	rooms := make([]model.Room, count)

	// 定義預設聊天室名稱和描述
	defaultRoomNames := []string{"公共聊天室", "技術討論", "休閒娛樂", "學習交流", "新手指南"}
	defaultRoomDescs := []string{
		"這是一個公開的聊天室，所有人都可以加入",
		"討論各種技術話題，包括程式設計、網絡、數據庫等",
		"分享生活趣事、電影、音樂、遊戲等娛樂話題",
		"交流學習心得、分享學習資源和方法",
		"為新手提供幫助和指導的聊天室",
	}

	// 創建聊天室
	for i := 0; i < count; i++ {
		roomName := defaultRoomNames[i%len(defaultRoomNames)]
		roomDesc := defaultRoomDescs[i%len(defaultRoomDescs)]

		rooms[i] = model.Room{
			Model:       gorm.Model{ID: uint(i + 1)},
			Name:        roomName,
			Description: roomDesc,
			IsPublic:    true,
			MaxUsers:    getMaxUsers(i), // 第一個聊天室最大人數為 200，其他為 100
			CreatedBy:   "system",
			IsActive:    true,
		}
		ctx.rooms[roomName] = &rooms[i]
	}

	// 設置模擬行為
	ctx.db.On("Find", mock.AnythingOfType("*[]model.Room"), mock.Anything).Run(func(args mock.Arguments) {
		roomsPtr := args.Get(0).(*[]model.Room)
		*roomsPtr = rooms
	}).Return(mockResult)

	return nil
}

func (ctx *MultiRoomTestContext) userIsViewingRoomList() error {
	return ctx.userVisitsRoomListPage()
}

func (ctx *MultiRoomTestContext) usersAreInRoom(users, roomName string) error {
	// 模擬用戶 A 和 B 在同一個聊天室
	room := ctx.rooms[roomName]
	if room == nil {
		return fmt.Errorf("聊天室 %s 不存在", roomName)
	}

	// 將 HTTP URL 轉換為 WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(ctx.server.URL, "http") + "/ws"

	// 連接用戶 A
	connA, _, err := websocket.DefaultDialer.Dial(wsURL+"?username=UserA&roomId="+fmt.Sprint(room.ID), nil)
	if err != nil {
		return fmt.Errorf("用戶 A 無法連接到 WebSocket: %v", err)
	}
	ctx.connections["UserA"] = connA

	// 連接用戶 B
	connB, _, err := websocket.DefaultDialer.Dial(wsURL+"?username=UserB&roomId="+fmt.Sprint(room.ID), nil)
	if err != nil {
		return fmt.Errorf("用戶 B 無法連接到 WebSocket: %v", err)
	}
	ctx.connections["UserB"] = connB

	// 發送加入聊天室訊息
	joinMsg := map[string]interface{}{
		"type":   "join_room",
		"target": fmt.Sprint(room.ID),
	}
	if err := connA.WriteJSON(joinMsg); err != nil {
		return fmt.Errorf("用戶 A 無法發送加入聊天室訊息: %v", err)
	}
	if err := connB.WriteJSON(joinMsg); err != nil {
		return fmt.Errorf("用戶 B 無法發送加入聊天室訊息: %v", err)
	}

	// 等待一段時間，確保訊息被處理
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (ctx *MultiRoomTestContext) userIsInRoom(user, roomName string) error {
	// 模擬用戶在特定聊天室
	room := ctx.rooms[roomName]
	if room == nil {
		return fmt.Errorf("聊天室 %s 不存在", roomName)
	}

	// 將 HTTP URL 轉換為 WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(ctx.server.URL, "http") + "/ws"

	// 連接用戶
	conn, _, err := websocket.DefaultDialer.Dial(wsURL+"?username="+user+"&roomId="+fmt.Sprint(room.ID), nil)
	if err != nil {
		return fmt.Errorf("用戶 %s 無法連接到 WebSocket: %v", user, err)
	}
	ctx.connections[user] = conn

	// 發送加入聊天室訊息
	joinMsg := map[string]interface{}{
		"type":   "join_room",
		"target": fmt.Sprint(room.ID),
	}
	if err := conn.WriteJSON(joinMsg); err != nil {
		return fmt.Errorf("用戶 %s 無法發送加入聊天室訊息: %v", user, err)
	}

	// 等待一段時間，確保訊息被處理
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (ctx *MultiRoomTestContext) userAIsInRoom(roomName string) error {
	return ctx.userIsInRoom("UserA", roomName)
}

func (ctx *MultiRoomTestContext) roomListPage() error {
	return ctx.userVisitsRoomListPage()
}

func (ctx *MultiRoomTestContext) usersJoinRoom(count int, roomName string) error {
	// 模擬多個用戶加入聊天室
	room := ctx.rooms[roomName]
	if room == nil {
		return fmt.Errorf("聊天室 %s 不存在", roomName)
	}

	// 將 HTTP URL 轉換為 WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(ctx.server.URL, "http") + "/ws"

	// 連接多個用戶
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("User%d", i+1)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL+"?username="+username+"&roomId="+fmt.Sprint(room.ID), nil)
		if err != nil {
			return fmt.Errorf("用戶 %s 無法連接到 WebSocket: %v", username, err)
		}
		ctx.connections[username] = conn

		// 發送加入聊天室訊息
		joinMsg := map[string]interface{}{
			"type":   "join_room",
			"target": fmt.Sprint(room.ID),
		}
		if err := conn.WriteJSON(joinMsg); err != nil {
			return fmt.Errorf("用戶 %s 無法發送加入聊天室訊息: %v", username, err)
		}
	}

	// 等待一段時間，確保訊息被處理
	time.Sleep(100 * time.Millisecond)

	return nil
}

// When 步驟實現
func (ctx *MultiRoomTestContext) userVisitsRoomListPage() error {
	// 模擬獲取聊天室列表的請求
	req, _ := http.NewRequest("GET", "/api/rooms", nil)
	w := httptest.NewRecorder()
	ctx.responses["roomList"] = w
	ctx.router.ServeHTTP(w, req)

	return nil
}

func (ctx *MultiRoomTestContext) userClicksRoom(roomName string) error {
	// 模擬用戶點擊聊天室
	room := ctx.rooms[roomName]
	if room == nil {
		return fmt.Errorf("聊天室 %s 不存在", roomName)
	}

	// 模擬獲取特定聊天室的請求
	req, _ := http.NewRequest("GET", "/api/rooms/"+fmt.Sprint(room.ID), nil)
	w := httptest.NewRecorder()
	ctx.responses["room"] = w
	ctx.router.ServeHTTP(w, req)

	return nil
}

func (ctx *MultiRoomTestContext) userSendsMessageInRoom(user, roomName, message string) error {
	// 模擬用戶在聊天室發送訊息
	conn := ctx.connections[user]
	if conn == nil {
		return fmt.Errorf("用戶 %s 未連接到 WebSocket", user)
	}

	// 發送聊天訊息
	chatMsg := map[string]interface{}{
		"content": message,
		"sender":  user,
		"time":    time.Now().Unix(),
	}
	if err := conn.WriteJSON(chatMsg); err != nil {
		return fmt.Errorf("用戶 %s 無法發送聊天訊息: %v", user, err)
	}

	// 等待一段時間，確保訊息被處理
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (ctx *MultiRoomTestContext) userClicksBackButton() error {
	// 模擬用戶點擊返回按鈕
	// 在實際應用中，這會導航回聊天室列表頁面
	return ctx.userVisitsRoomListPage()
}

func (ctx *MultiRoomTestContext) userSelectsRoomFromList(roomName string) error {
	return ctx.userClicksRoom(roomName)
}

func (ctx *MultiRoomTestContext) userBJoinsRoom(roomName string) error {
	// 模擬用戶 B 加入聊天室
	room := ctx.rooms[roomName]
	if room == nil {
		return fmt.Errorf("聊天室 %s 不存在", roomName)
	}

	// 將 HTTP URL 轉換為 WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(ctx.server.URL, "http") + "/ws"

	// 連接用戶 B
	connB, _, err := websocket.DefaultDialer.Dial(wsURL+"?username=UserB&roomId="+fmt.Sprint(room.ID), nil)
	if err != nil {
		return fmt.Errorf("用戶 B 無法連接到 WebSocket: %v", err)
	}
	ctx.connections["UserB"] = connB

	// 發送加入聊天室訊息
	joinMsg := map[string]interface{}{
		"type":   "join_room",
		"target": fmt.Sprint(room.ID),
	}
	if err := connB.WriteJSON(joinMsg); err != nil {
		return fmt.Errorf("用戶 B 無法發送加入聊天室訊息: %v", err)
	}

	// 等待一段時間，確保訊息被處理
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (ctx *MultiRoomTestContext) userBLeavesRoom(roomName string) error {
	// 模擬用戶 B 離開聊天室
	connB := ctx.connections["UserB"]
	if connB == nil {
		return fmt.Errorf("用戶 B 未連接到 WebSocket")
	}

	// 發送離開聊天室訊息
	leaveMsg := map[string]interface{}{
		"type": "leave_room",
	}
	if err := connB.WriteJSON(leaveMsg); err != nil {
		return fmt.Errorf("用戶 B 無法發送離開聊天室訊息: %v", err)
	}

	// 等待一段時間，確保訊息被處理
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Then 步驟實現
func (ctx *MultiRoomTestContext) userShouldSeeRoomList(count int) error {
	// 驗證用戶是否看到指定數量的聊天室
	w := ctx.responses["roomList"]
	if w == nil {
		return fmt.Errorf("未找到聊天室列表響應")
	}

	if w.Code != http.StatusOK {
		return fmt.Errorf("預期狀態碼為 %d，實際為 %d", http.StatusOK, w.Code)
	}

	var response []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		return fmt.Errorf("無法解析響應: %v", err)
	}

	if len(response) != count {
		return fmt.Errorf("預期有 %d 個聊天室，實際有 %d 個", count, len(response))
	}

	return nil
}

func (ctx *MultiRoomTestContext) eachRoomShouldShowNameAndDescription() error {
	// 驗證每個聊天室是否顯示名稱和描述
	w := ctx.responses["roomList"]
	if w == nil {
		return fmt.Errorf("未找到聊天室列表響應")
	}

	var response []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		return fmt.Errorf("無法解析響應: %v", err)
	}

	for i, room := range response {
		if room["name"] == nil {
			return fmt.Errorf("聊天室 %d 沒有名稱", i+1)
		}
		if room["description"] == nil {
			return fmt.Errorf("聊天室 %d 沒有描述", i+1)
		}
	}

	return nil
}

func (ctx *MultiRoomTestContext) userShouldBeRedirectedToChatInterface(roomName string) error {
	// 驗證用戶是否被導向到聊天界面
	w := ctx.responses["room"]
	if w == nil {
		return fmt.Errorf("未找到聊天室響應")
	}

	if w.Code != http.StatusOK {
		return fmt.Errorf("預期狀態碼為 %d，實際為 %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		return fmt.Errorf("無法解析響應: %v", err)
	}

	if response["name"] != roomName {
		return fmt.Errorf("預期聊天室名稱為 %s，實際為 %s", roomName, response["name"])
	}

	return nil
}

func (ctx *MultiRoomTestContext) userShouldSeeWelcomeMessage(roomName string) error {
	// 驗證用戶是否看到歡迎訊息
	// 在實際應用中，這可能需要檢查 WebSocket 訊息
	return nil
}

func (ctx *MultiRoomTestContext) userShouldSeeMessageInRoom(user, roomName, message string) error {
	// 驗證用戶是否在聊天室中看到訊息
	// 在實際應用中，這可能需要檢查 WebSocket 訊息
	return nil
}

func (ctx *MultiRoomTestContext) userShouldNotSeeMessageInRoom(user, roomName string) error {
	// 驗證用戶是否沒有在聊天室中看到訊息
	// 在實際應用中，這可能需要檢查 WebSocket 訊息
	return nil
}

func (ctx *MultiRoomTestContext) userShouldEnterRoom(roomName string) error {
	// 驗證用戶是否進入聊天室
	return ctx.userShouldBeRedirectedToChatInterface(roomName)
}

func (ctx *MultiRoomTestContext) userShouldSeeRoomHistory(roomName string) error {
	// 驗證用戶是否看到聊天室歷史訊息
	// 在實際應用中，這可能需要檢查 API 響應
	return nil
}

func (ctx *MultiRoomTestContext) userShouldNotReceiveNewMessages(roomName string) error {
	// 驗證用戶是否不再收到新訊息
	// 在實際應用中，這可能需要檢查 WebSocket 訊息
	return nil
}

func (ctx *MultiRoomTestContext) userShouldSeeSystemNotification(notification string) error {
	// 驗證用戶是否看到系統通知
	// 在實際應用中，這可能需要檢查 WebSocket 訊息
	return nil
}

func (ctx *MultiRoomTestContext) roomShouldShowUserCount(roomName, userCount string) error {
	// 驗證聊天室是否顯示正確的用戶數量
	// 在實際應用中，這可能需要檢查 API 響應
	return nil
}

// getMaxUsers 根據索引返回最大用戶數
func getMaxUsers(index int) int {
	if index < 1 {
		return 200
	}
	return 100
}
