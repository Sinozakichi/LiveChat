// 聊天室列表頁面的 JavaScript

document.addEventListener('DOMContentLoaded', function() {
    const roomListElement = document.getElementById('roomList');
    const createRoomForm = document.getElementById('createRoomForm');
    const createRoomSection = document.getElementById('createRoomSection');
    
    // 載入聊天室列表
    loadRoomList();
    
    // 檢查用戶角色並決定是否顯示創建聊天室區塊
    checkUserRoleAndShowCreateRoom();
    
    // 顯示當前用戶信息
    displayCurrentUser();
    
    // 設置創建聊天室表單提交事件
    createRoomForm.addEventListener('submit', function(e) {
        e.preventDefault();
        createRoom();
    });
    
    // 檢查用戶角色並決定是否顯示創建聊天室區塊
    function checkUserRoleAndShowCreateRoom() {
        fetch('/api/user')
            .then(response => {
                if (!response.ok) {
                    throw new Error('無法獲取用戶資訊');
                }
                return response.json();
            })
            .then(user => {
                // 只有管理員才能看到創建聊天室區塊
                if (user.role === 'admin') {
                    createRoomSection.style.display = 'block';
                } else {
                    createRoomSection.style.display = 'none';
                }
            })
            .catch(error => {
                console.error('Error checking user role:', error);
                // 如果無法獲取用戶資訊，預設隱藏創建聊天室區塊
                createRoomSection.style.display = 'none';
            });
    }
    
    // 顯示當前用戶信息
    function displayCurrentUser() {
        const currentUserElement = document.getElementById('currentUser');
        if (!currentUserElement) return;
        
        fetch('/api/user')
            .then(response => {
                if (!response.ok) {
                    throw new Error('無法獲取用戶資訊');
                }
                return response.json();
            })
            .then(user => {
                currentUserElement.textContent = `歡迎，${user.username}`;
            })
            .catch(error => {
                console.error('Error getting current user:', error);
                currentUserElement.textContent = '未登入';
            });
    }
    
    // 載入聊天室列表
    function loadRoomList() {
        roomListElement.innerHTML = `
            <div class="d-flex justify-content-center">
                <div class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">正在載入...</span>
                </div>
            </div>
            <p class="text-center mt-2">正在載入聊天室列表...</p>
        `;
        
        console.log("Fetching rooms from API...");
        fetch('/api/rooms')
            .then(response => {
                console.log("API response status:", response.status);
                if (!response.ok) {
                    throw new Error(`無法獲取聊天室列表 (Status: ${response.status})`);
                }
                return response.json();
            })
            .then(rooms => {
                console.log("Received rooms:", rooms);
                displayRooms(rooms);
            })
            .catch(error => {
                console.error('Error:', error);
                roomListElement.innerHTML = `
                    <div class="alert alert-danger" role="alert">
                        <i class="bi bi-exclamation-triangle-fill me-2"></i>
                        載入聊天室列表失敗: ${error.message}
                    </div>
                    <button class="btn btn-primary mt-3" onclick="location.reload()">重新整理</button>
                `;
            });
    }
    
    // 顯示聊天室列表
    function displayRooms(rooms) {
        if (rooms.length === 0) {
            roomListElement.innerHTML = `
                <div class="alert alert-info" role="alert">
                    <i class="bi bi-info-circle-fill me-2"></i>
                    目前沒有可用的聊天室。請使用下方表單創建一個新的聊天室！
                </div>
            `;
            // 突出顯示創建聊天室表單
            document.querySelector('.create-room-section').classList.add('border', 'border-primary', 'p-3', 'rounded');
            return;
        }
        
        let html = '<div class="row row-cols-1 row-cols-md-2 g-3">';
        rooms.forEach(room => {
            html += `
                <div class="col">
                    <div class="card room-card h-100">
                        <div class="card-body">
                            <h5 class="card-title">${escapeHtml(room.name)}</h5>
                            <p class="card-text text-muted">${escapeHtml(room.description || '無描述')}</p>
                        </div>
                        <div class="card-footer bg-transparent d-flex justify-content-between align-items-center">
                            <!-- 在線人數統計功能暫時註解掉
                            <span class="badge bg-primary rounded-pill">
                                <i class="bi bi-people-fill"></i> ${room.activeUsers} 人在線
                            </span>
                            -->
                            <div></div> <!-- 左側空白 -->
                            <button class="btn btn-sm btn-outline-primary join-btn" data-room-id="${room.id}">
                                <i class="bi bi-box-arrow-in-right"></i> 加入聊天室
                            </button>
                        </div>
                    </div>
                </div>
            `;
        });
        html += '</div>';
        
        roomListElement.innerHTML = html;
        
        // 為加入聊天室按鈕添加事件監聽器
        document.querySelectorAll('.join-btn').forEach(button => {
            button.addEventListener('click', function() {
                const roomId = this.getAttribute('data-room-id');
                joinRoom(roomId);
            });
        });
    }
    
    // 創建聊天室
    function createRoom() {
        const roomName = document.getElementById('roomName').value;
        const roomDescription = document.getElementById('roomDescription').value;
        const isPublic = document.getElementById('isPublic').checked;
        const maxUsers = parseInt(document.getElementById('maxUsers').value);
        
        const roomData = {
            name: roomName,
            description: roomDescription,
            isPublic: isPublic,
            maxUsers: maxUsers
        };
        
        // 顯示載入中狀態
        const submitBtn = createRoomForm.querySelector('button[type="submit"]');
        const originalText = submitBtn.innerHTML;
        submitBtn.innerHTML = '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> 創建中...';
        submitBtn.disabled = true;
        
        fetch('/api/rooms', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(roomData)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('創建聊天室失敗');
            }
            return response.json();
        })
        .then(room => {
            // 顯示成功訊息
            showAlert('聊天室創建成功！', 'success');
            loadRoomList();
            createRoomForm.reset();
        })
        .catch(error => {
            console.error('Error:', error);
            showAlert(`創建聊天室失敗: ${error.message}`, 'danger');
        })
        .finally(() => {
            // 恢復按鈕狀態
            submitBtn.innerHTML = originalText;
            submitBtn.disabled = false;
        });
    }
    
    // 加入聊天室
    function joinRoom(roomId) {
        // 從API獲取當前登入用戶資訊
        fetch('/api/user')
            .then(response => {
                if (!response.ok) {
                    throw new Error('請先登入系統');
                }
                return response.json();
            })
            .then(user => {
                // 使用登入用戶的用戶名導航到聊天室頁面
                window.location.href = `chat.html?roomId=${roomId}&username=${encodeURIComponent(user.username)}`;
            })
            .catch(error => {
                console.error('Error getting user info:', error);
                // 如果無法獲取用戶資訊，重定向到登入頁面
                alert('請先登入系統');
                window.location.href = '/';
            });
    }
    
    // 顯示警告訊息
    function showAlert(message, type = 'info') {
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type} alert-dismissible fade show`;
        alertDiv.setAttribute('role', 'alert');
        
        alertDiv.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
        `;
        
        // 插入到表單前面
        createRoomForm.parentNode.insertBefore(alertDiv, createRoomForm);
        
        // 5秒後自動消失
        setTimeout(() => {
            alertDiv.classList.remove('show');
            setTimeout(() => alertDiv.remove(), 150);
        }, 5000);
    }
    
    // HTML 轉義函數，防止 XSS 攻擊
    function escapeHtml(unsafe) {
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
});