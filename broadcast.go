package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// 用戶管理邏輯
type Client struct {
	ID   string
	Conn *websocket.Conn
}

var clients = make(map[string]*Client)

// 消息廣播邏輯
// 為所有連接的用戶實現消息廣播功能
func broadcastMessage(message []byte) {
	for _, client := range clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Println("Write error:", err)
			client.Conn.Close()
			delete(clients, client.ID)
		}
	}
}
