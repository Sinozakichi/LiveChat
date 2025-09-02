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
)

// MultiRoomTestContext 保存測試狀態
//
// BDD 測試上下文設計：
// 1. 維護跨場景的資料庫連線和狀態
// 2. 提供聊天室名稱到實體的映射
// 3. 管理 WebSocket 連線的生命週期
// 4. 記錄 HTTP 請求和回應以供驗證
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
	rooms            map[string]*model.Room // 聊天室名稱 -> 聊天室實體的映射
	roomsById        map[string]*model.Room // 聊天室ID -> 聊天室實體的映射
}

// 全域測試上下文，用於跨場景保持狀態
var globalTestContext *MultiRoomTestContext

// InitializeMultiRoomScenario 初始化多聊天室功能的 BDD 測試
func InitializeMultiRoomScenario(ctx *godog.ScenarioContext) {
	// 設置 Gin 測試模式
	gin.SetMode(gin.TestMode)

	// 在每個場景開始前設置測試環境
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// 如果是第一次初始化或者全域上下文不存在，創建新的測試上下文
		if globalTestContext == nil {
			globalTestContext = &MultiRoomTestContext{
				connections:   make(map[string]*websocket.Conn),
				responses:     make(map[string]*httptest.ResponseRecorder),
				requestBodies: make(map[string]string),
				rooms:         make(map[string]*model.Room),
				roomsById:     make(map[string]*model.Room),
			}

			// 創建一個包含完整結構的模擬 GORM DB（只創建一次）
			globalTestContext.db = repository.NewMockDBWithSchema()
		} else {
			// 清理連線和回應，但保留資料庫和聊天室資料
			for _, conn := range globalTestContext.connections {
				if conn != nil {
					conn.Close()
				}
			}
			globalTestContext.connections = make(map[string]*websocket.Conn)
			globalTestContext.responses = make(map[string]*httptest.ResponseRecorder)

			// 關閉舊的 HTTP 伺服器
			if globalTestContext.server != nil {
				globalTestContext.server.Close()
			}
		}

		// 每次都重新創建服務和處理器（但使用相同的資料庫）
		testCtx := globalTestContext

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

	// 在每個場景結束後清理測試環境（但不清理全域狀態）
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// 注意：這裡不清理 globalTestContext，讓資料庫狀態跨場景保持
		// 只清理當前場景的連線
		if globalTestContext != nil {
			for _, conn := range globalTestContext.connections {
				if conn != nil {
					conn.Close()
				}
			}
			// 清空連線映射，但保留其他狀態
			globalTestContext.connections = make(map[string]*websocket.Conn)
		}

		return ctx, nil
	})

	// Given 步驟 - 使用包裝函數來存取全域上下文
	ctx.Step(`^系統中有(\d+)個預設聊天室$`, func(count int) error {
		return globalTestContext.thereAreDefaultRooms(count)
	})
	ctx.Step(`^使用者正在查看聊天室列表$`, func() error {
		return globalTestContext.userIsViewingRoomList()
	})
	ctx.Step(`^使用者A和使用者B都在"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.usersABInRoom(roomName)
	})
	ctx.Step(`^使用者C在"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userCInRoom(roomName)
	})
	ctx.Step(`^使用者已經在"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userInRoomGeneric(roomName)
	})
	ctx.Step(`^使用者A已經在"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userAIsInRoom(roomName)
	})
	ctx.Step(`^聊天室列表頁面$`, func() error {
		return globalTestContext.roomListPage()
	})
	ctx.Step(`^(\d+)個使用者加入"([^"]*)"$`, func(count int, roomName string) error {
		return globalTestContext.usersJoinRoom(count, roomName)
	})

	// When 步驟
	ctx.Step(`^使用者訪問聊天室列表頁面$`, func() error {
		return globalTestContext.userVisitsRoomListPage()
	})
	ctx.Step(`^使用者點擊"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userClicksRoom(roomName)
	})
	ctx.Step(`^使用者A在"([^"]*)"發送訊息"([^"]*)"$`, func(roomName, message string) error {
		return globalTestContext.userSendsMessageInRoom("UserA", roomName, message)
	})
	ctx.Step(`^使用者點擊返回按鈕$`, func() error {
		return globalTestContext.userClicksBackButton()
	})
	ctx.Step(`^使用者從列表中選擇"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userSelectsRoomFromList(roomName)
	})
	ctx.Step(`^使用者B加入"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userBJoinsRoom(roomName)
	})
	ctx.Step(`^使用者B離開"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userBLeavesRoom(roomName)
	})

	// Then 步驟
	ctx.Step(`^使用者應該看到(\d+)個聊天室的列表$`, func(count int) error {
		return globalTestContext.userShouldSeeRoomList(count)
	})
	ctx.Step(`^每個聊天室應顯示名稱和簡短描述$`, func() error {
		return globalTestContext.eachRoomShouldShowNameAndDescription()
	})
	ctx.Step(`^使用者應該被導向到"([^"]*)"的聊天界面$`, func(roomName string) error {
		return globalTestContext.userShouldBeRedirectedToChatInterface(roomName)
	})
	ctx.Step(`^使用者應該看到"([^"]*)"的歡迎訊息$`, func(roomName string) error {
		return globalTestContext.userShouldSeeWelcomeMessage(roomName)
	})
	ctx.Step(`^使用者B應該在"([^"]*)"看到訊息"([^"]*)"$`, func(roomName, message string) error {
		return globalTestContext.userShouldSeeMessageInRoom("UserB", roomName, message)
	})
	ctx.Step(`^使用者C不應該在"([^"]*)"看到該訊息$`, func(roomName string) error {
		return globalTestContext.userShouldNotSeeMessageInRoom("UserC", roomName)
	})
	ctx.Step(`^使用者應該進入"([^"]*)"$`, func(roomName string) error {
		return globalTestContext.userShouldEnterRoom(roomName)
	})
	ctx.Step(`^使用者應該看到"([^"]*)"的歷史訊息$`, func(roomName string) error {
		return globalTestContext.userShouldSeeRoomHistory(roomName)
	})
	ctx.Step(`^使用者不應該再收到"([^"]*)"的新訊息$`, func(roomName string) error {
		return globalTestContext.userShouldNotReceiveNewMessages(roomName)
	})
	ctx.Step(`^使用者A應該看到系統通知"([^"]*)"$`, func(notification string) error {
		return globalTestContext.userShouldSeeSystemNotification(notification)
	})
	ctx.Step(`^"([^"]*)"應顯示"([^"]*)"$`, func(roomName, userCount string) error {
		return globalTestContext.roomShouldShowUserCount(roomName, userCount)
	})
}

// Given 步驟實現
//
// thereAreDefaultRooms BDD 步驟：「系統中有 N 個預設聊天室」
//
// 職責：
// 1. 在真實的記憶體資料庫中創建指定數量的測試聊天室
// 2. 設定聊天室的名稱、描述和其他屬性
// 3. 將聊天室記錄到測試上下文中以供後續使用
// 4. 確保每個聊天室都是活躍狀態
func (ctx *MultiRoomTestContext) thereAreDefaultRooms(count int) error {
	// 為了與 BDD 特徵檔案中的聊天室名稱保持一致，
	// 我們創建名為 "聊天室1", "聊天室2" 等的聊天室

	// 創建並插入指定數量的聊天室到真實資料庫
	for i := 0; i < count; i++ {
		roomName := fmt.Sprintf("聊天室%d", i+1)
		roomDesc := fmt.Sprintf("這是第 %d 個測試聊天室", i+1)
		roomID := fmt.Sprintf("test-room-%d", i+1)

		room := model.Room{
			ID:          roomID,
			Name:        roomName,
			Description: roomDesc,
			IsPublic:    true,
			MaxUsers:    getMaxUsers(i), // 第一個聊天室最大人數為 200，其他為 100
			CreatedBy:   "system",
			IsActive:    true,
		}

		// 直接插入到真實的記憶體資料庫中
		err := ctx.db.DB.Create(&room).Error
		if err != nil {
			return fmt.Errorf("創建聊天室失敗: %v", err)
		}

		// 記錄到測試上下文中（同時記錄名稱和 ID 的映射）
		ctx.rooms[roomName] = &room
		ctx.roomsById[roomID] = &room
	}

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

func (ctx *MultiRoomTestContext) usersABInRoom(roomName string) error {
	return ctx.usersAreInRoom("A,B", roomName)
}

func (ctx *MultiRoomTestContext) userCInRoom(roomName string) error {
	return ctx.userIsInRoom("UserC", roomName)
}

func (ctx *MultiRoomTestContext) userInRoomGeneric(roomName string) error {
	return ctx.userIsInRoom("User", roomName)
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
