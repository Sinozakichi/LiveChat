package repository

import (
	"errors"
	"livechat/backend/model"
	"sync"
)

// 定義錯誤
var (
	ErrClientNotFound = errors.New("客戶端不存在")
	ErrClientExists   = errors.New("客戶端已存在")
)

// ClientRepository 管理所有連接的客戶端
type ClientRepository struct {
	clients map[string]*model.Client
	mutex   sync.RWMutex
}

// NewClientRepository 創建一個新的客戶端儲存庫
func NewClientRepository() *ClientRepository {
	return &ClientRepository{
		clients: make(map[string]*model.Client),
	}
}

// Add 添加一個新的客戶端
func (r *ClientRepository) Add(client *model.Client) error {
	if client == nil {
		return errors.New("客戶端不能為空")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.clients[client.ID]; exists {
		return ErrClientExists
	}

	r.clients[client.ID] = client
	return nil
}

// Remove 移除一個客戶端
func (r *ClientRepository) Remove(clientID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.clients[clientID]; !exists {
		return ErrClientNotFound
	}

	delete(r.clients, clientID)
	return nil
}

// Get 獲取指定的客戶端
func (r *ClientRepository) Get(clientID string) (*model.Client, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	client, exists := r.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// GetAll 獲取所有客戶端
func (r *ClientRepository) GetAll() map[string]*model.Client {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// 創建一個副本以避免並發問題
	clientsCopy := make(map[string]*model.Client)
	for id, client := range r.clients {
		clientsCopy[id] = client
	}

	return clientsCopy
}

// GetActiveClients 獲取所有活躍的客戶端
func (r *ClientRepository) GetActiveClients() []*model.Client {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var activeClients []*model.Client
	for _, client := range r.clients {
		if client.IsActive {
			activeClients = append(activeClients, client)
		}
	}

	return activeClients
}

// Count 獲取客戶端總數
func (r *ClientRepository) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.clients)
}

// Clear 清空所有客戶端
func (r *ClientRepository) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.clients = make(map[string]*model.Client)
}
