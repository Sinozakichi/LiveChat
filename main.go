package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

// websocket.Upgrader 是 websocket 套件中用來處理將 HTTP 連接升級為 WebSocket 連接的過程
var upgrader = websocket.Upgrader{
	//websocket.Upgrader的欄位之一，用來檢查 WebSocket 連接的來源
	CheckOrigin: func(r *http.Request) bool {
		//表示接受所有來源的 WebSocket 連接
		return true
	},
}

// 處理 WebSocket 連接的函數
func handleConnection(w http.ResponseWriter, r *http.Request) {
	//將 HTTP 連接升級為 WebSocket 連接｛
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade:", err)
		return
	}
	defer conn.Close() // 在函數返回前關閉連接

	// 為每個新連接創建一個唯一的 ID
	clientID := fmt.Sprintf("%p", conn)
	client := &Client{
		ID:   clientID,
		Conn: conn,
	}
	// 將新的連接加入到 clients 集合中
	clients[clientID] = client

	for {
		// 讀取 WebSocket 訊息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			delete(clients, clientID)
			break
		}
		fmt.Printf("Received: %s\n", msg)
		// 將接收到的訊息回傳給客戶端
		broadcastMessage(msg)
		//conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func main() {
	// 設定端口，根據環境變數來獲取
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 如果沒設定，則使用 8080
	}
	fmt.Println("Server started at :" + port)

	// 設定路由，當請求 URI 為 "/ws" 時，路由到 handleConnection 函數
	http.HandleFunc("/ws", handleConnection)

	// 啟動 HTTP 伺服器，監聽指定端口
	http.ListenAndServe(":"+port, nil)
}
