package repository

import (
	"livechat/backend/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 測試創建新的客戶端儲存庫
func TestNewClientRepository(t *testing.T) {
	// 動作 (Act)
	repo := NewClientRepository()

	// 斷言 (Assert)
	assert.NotNil(t, repo, "儲存庫不應該為 nil")
	assert.NotNil(t, repo.clients, "clients map 不應該為 nil")
	assert.Equal(t, 0, len(repo.clients), "新儲存庫應該是空的")
}

// 測試添加客戶端
func TestAddClient(t *testing.T) {
	// 安排 (Arrange)
	repo := NewClientRepository()
	client := model.NewClient("test-id", nil)

	// 動作 (Act)
	err := repo.Add(client)

	// 斷言 (Assert)
	assert.NoError(t, err, "添加客戶端不應該返回錯誤")
	assert.Equal(t, 1, repo.Count(), "儲存庫應該有一個客戶端")

	// 測試添加 nil 客戶端
	err = repo.Add(nil)
	assert.Error(t, err, "添加 nil 客戶端應該返回錯誤")

	// 測試添加重複的客戶端
	err = repo.Add(client)
	assert.Equal(t, ErrClientExists, err, "添加重複的客戶端應該返回 ErrClientExists")
}

// 測試獲取客戶端
func TestGetClient(t *testing.T) {
	// 安排 (Arrange)
	repo := NewClientRepository()
	client := model.NewClient("test-id", nil)
	repo.Add(client)

	// 動作 (Act)
	retrievedClient, err := repo.Get("test-id")

	// 斷言 (Assert)
	assert.NoError(t, err, "獲取存在的客戶端不應該返回錯誤")
	assert.Equal(t, client, retrievedClient, "獲取的客戶端應該與添加的客戶端相同")

	// 測試獲取不存在的客戶端
	_, err = repo.Get("non-existent-id")
	assert.Equal(t, ErrClientNotFound, err, "獲取不存在的客戶端應該返回 ErrClientNotFound")
}

// 測試移除客戶端
func TestRemoveClient(t *testing.T) {
	// 安排 (Arrange)
	repo := NewClientRepository()
	client := model.NewClient("test-id", nil)
	repo.Add(client)

	// 動作 (Act)
	err := repo.Remove("test-id")

	// 斷言 (Assert)
	assert.NoError(t, err, "移除存在的客戶端不應該返回錯誤")
	assert.Equal(t, 0, repo.Count(), "移除後儲存庫應該是空的")

	// 測試移除不存在的客戶端
	err = repo.Remove("non-existent-id")
	assert.Equal(t, ErrClientNotFound, err, "移除不存在的客戶端應該返回 ErrClientNotFound")
}

// 測試獲取所有客戶端
func TestGetAllClients(t *testing.T) {
	// 安排 (Arrange)
	repo := NewClientRepository()
	client1 := model.NewClient("test-id-1", nil)
	client2 := model.NewClient("test-id-2", nil)
	repo.Add(client1)
	repo.Add(client2)

	// 動作 (Act)
	allClients := repo.GetAll()

	// 斷言 (Assert)
	assert.Equal(t, 2, len(allClients), "應該有兩個客戶端")
	assert.Equal(t, client1, allClients["test-id-1"], "第一個客戶端應該匹配")
	assert.Equal(t, client2, allClients["test-id-2"], "第二個客戶端應該匹配")
}

// 測試獲取活躍客戶端
func TestGetActiveClients(t *testing.T) {
	// 安排 (Arrange)
	repo := NewClientRepository()
	client1 := model.NewClient("test-id-1", nil)
	client2 := model.NewClient("test-id-2", nil)
	client2.Deactivate() // 將第二個客戶端標記為非活躍
	repo.Add(client1)
	repo.Add(client2)

	// 動作 (Act)
	activeClients := repo.GetActiveClients()

	// 斷言 (Assert)
	assert.Equal(t, 1, len(activeClients), "應該只有一個活躍的客戶端")
	assert.Equal(t, client1, activeClients[0], "活躍的客戶端應該是第一個客戶端")
}

// 測試計數客戶端
func TestCountClients(t *testing.T) {
	// 安排 (Arrange)
	repo := NewClientRepository()
	client1 := model.NewClient("test-id-1", nil)
	client2 := model.NewClient("test-id-2", nil)
	repo.Add(client1)
	repo.Add(client2)

	// 動作 (Act)
	count := repo.Count()

	// 斷言 (Assert)
	assert.Equal(t, 2, count, "應該有兩個客戶端")
}

// 測試清空儲存庫
func TestClearRepository(t *testing.T) {
	// 安排 (Arrange)
	repo := NewClientRepository()
	client1 := model.NewClient("test-id-1", nil)
	client2 := model.NewClient("test-id-2", nil)
	repo.Add(client1)
	repo.Add(client2)
	assert.Equal(t, 2, repo.Count(), "初始應該有兩個客戶端")

	// 動作 (Act)
	repo.Clear()

	// 斷言 (Assert)
	assert.Equal(t, 0, repo.Count(), "清空後儲存庫應該是空的")
}
