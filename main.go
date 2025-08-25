package main

import (
	"fmt"
	"livechat/backend/handler"
	"livechat/backend/middleware"
	"livechat/backend/migrations"
	"livechat/backend/repository"
	"livechat/backend/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 載入環境變數
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// 初始化數據庫
	db, err := initDB()
	if err != nil {
		fmt.Printf("Database initialization error: %v\n", err)
		return
	}

	// 執行資料庫遷移
	migrator := migrations.NewMigrator(db)
	if err := migrator.MigrateUp(); err != nil {
		fmt.Printf("Migration error: %v\n", err)
		return
	}

	// 創建儲存庫
	clientRepo := repository.NewClientRepository()
	roomRepo := repository.NewRoomRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 創建服務
	broadcastService := service.NewBroadcastService(clientRepo)
	roomService := service.NewRoomService(roomRepo)
	userService := service.NewUserService(userRepo)

	// 創建處理器
	wsHandler := handler.NewWebSocketHandler(broadcastService, handler.WithLogger(&handler.DefaultLogger{}))
	roomHandler := handler.NewRoomHandler(roomService)
	userHandler := handler.NewUserHandler(userService)

	// 創建 Gin 路由
	router := gin.Default()

	// 加載 HTML 模板
	router.LoadHTMLGlob("frontend/*.html")

	// 設置會話中間件
	router.Use(middleware.SessionMiddleware(userService))

	// 註冊用戶相關路由
	userHandler.RegisterRoutes(router)

	// 註冊聊天室相關路由
	roomHandler.RegisterRoutes(router)

	// WebSocket 路由
	router.GET("/ws", func(c *gin.Context) {
		wsHandler.HandleConnection(c.Writer, c.Request)
	})

	// 靜態文件服務 - 使用更具體的路徑，避免與 API 路由衝突
	router.Static("/static", "./frontend/css")
	router.Static("/css", "./frontend/css") // 添加CSS路由映射
	router.Static("/js", "./frontend/js")
	router.StaticFile("/", "./frontend/index.html")           // 登入頁面設為首頁
	router.StaticFile("/rooms.html", "./frontend/rooms.html") // 聊天室列表頁面
	router.StaticFile("/chat.html", "./frontend/chat.html")   // 聊天頁面
	router.StaticFile("/login.html", "./frontend/login.html") // 保留原登入頁面路由
	router.StaticFile("/register.html", "./frontend/register.html")

	// 啟動服務器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server started at :" + port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// 優雅關閉
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	fmt.Println("Shutting down server...")
	fmt.Println("Server gracefully stopped")
}

// 初始化數據庫
func initDB() (*gorm.DB, error) {
	// 從環境變數獲取資料庫連線字串
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// 如果環境變數未設置，使用預設連線字串
		dbURL = "postgres://postgres:eoae368619220@db.ywhbgozuehgeadaqorsx.supabase.co:5432/postgres?sslmode=require"
		fmt.Println("Warning: DATABASE_URL not set, using default connection string")
	}

	// 連接 PostgreSQL 資料庫
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to PostgreSQL database")
	return db, nil
}
