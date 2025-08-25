// 聊天室頁面的 JavaScript

document.addEventListener('DOMContentLoaded', function() {
    // 獲取 URL 參數
    const urlParams = new URLSearchParams(window.location.search);
    const roomId = urlParams.get('roomId');
    const username = urlParams.get('username');
    
    // 如果沒有聊天室 ID 或用戶名，則返回列表頁面
    if (!roomId || !username) {
        window.location.href = 'rooms.html';
        return;
    }
    
    // 獲取 DOM 元素
    const roomTitle = document.getElementById('roomTitle');
    const roomDescription = document.getElementById('roomDescription');
    const onlineCount = document.getElementById('onlineCount');
    const chatMessages = document.getElementById('chatMessages');
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');
    const leaveRoomButton = document.getElementById('leaveRoom');
    const usersList = document.getElementById('usersList');
    
    // WebSocket 連接
    let socket = null;
    
    // 載入聊天室信息
    loadRoomInfo();
    
    // 顯示當前用戶信息
    displayCurrentUser();
    
    // 連接 WebSocket
    connectWebSocket();
    
    // 設置發送訊息事件
    sendButton.addEventListener('click', sendMessage);
    messageInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            sendMessage();
        }
    });
    
    // 設置離開聊天室事件
    leaveRoomButton.addEventListener('click', function() {
        leaveRoom();
        window.location.href = 'rooms.html';
    });
    
    // 顯示當前用戶信息
    function displayCurrentUser() {
        const currentUserElement = document.getElementById('currentUser');
        if (!currentUserElement) return;
        
        // 直接使用URL參數中的用戶名
        currentUserElement.textContent = `用戶：${username}`;
    }
    
    // 載入聊天室信息
    function loadRoomInfo() {
        fetch(`/api/rooms/${roomId}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('無法獲取聊天室信息');
                }
                return response.json();
            })
            .then(room => {
                roomTitle.textContent = room.name;
                roomDescription.textContent = room.description || '無描述';
                // updateOnlineCount(room.activeUsers); // 在線人數統計功能暫時註解掉
            })
            .catch(error => {
                console.error('Error:', error);
                addSystemMessage(`載入聊天室信息失敗: ${error.message}`);
            });
        
        // 載入聊天室用戶功能暫時註解掉
        // loadRoomUsers();
        
        // 載入聊天室訊息
        loadRoomMessages();
    }
    
    // 載入聊天室用戶功能暫時註解掉
    /*
    function loadRoomUsers() {
        fetch(`/api/rooms/${roomId}/users`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('無法獲取聊天室用戶');
                }
                return response.json();
            })
            .then(users => {
                updateUsersList(users);
            })
            .catch(error => {
                console.error('Error:', error);
            });
    }
    */
    
    // 載入聊天室訊息
    function loadRoomMessages() {
        fetch(`/api/rooms/${roomId}/messages?limit=50`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('無法獲取聊天室訊息');
                }
                return response.json();
            })
            .then(messages => {
                chatMessages.innerHTML = ''; // 清空訊息區域
                
                if (messages.length === 0) {
                    addSystemMessage('聊天室中還沒有訊息。');
                } else {
                    messages.forEach(message => {
                        if (message.IsSystemMessage) {
                            addSystemMessage(message.Content);
                        } else {
                            addMessage(message.UserID, message.Content, new Date(message.CreatedAt));
                        }
                    });
                }
                
                scrollToBottom();
            })
            .catch(error => {
                console.error('Error:', error);
                addSystemMessage(`載入訊息失敗: ${error.message}`);
            });
    }
    
    // 連接 WebSocket
    function connectWebSocket() {
        // 確定 WebSocket 協議
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        
        // 創建 WebSocket 連接
        socket = new WebSocket(`${protocol}//${host}/ws?username=${encodeURIComponent(username)}&roomId=${roomId}`);
        
        // 連接打開時
        socket.onopen = function() {
            addSystemMessage('已連接到聊天室。');
            sendJoinRoomMessage();
        };
        
        // 接收訊息時
        socket.onmessage = function(event) {
            try {
                const message = JSON.parse(event.data);
                
                if (message.type === 'system') {
                    addSystemMessage(message.content);
                } else {
                    addMessage(message.sender || '匿名', message.content, new Date(message.time * 1000));
                }
                
                scrollToBottom();
            } catch (error) {
                console.error('Error parsing message:', error);
                addSystemMessage(event.data);
                scrollToBottom();
            }
        };
        
        // 連接關閉時
        socket.onclose = function() {
            addSystemMessage('與伺服器的連接已關閉。');
            scrollToBottom();
        };
        
        // 連接錯誤時
        socket.onerror = function(error) {
            console.error('WebSocket Error:', error);
            addSystemMessage('連接錯誤，請刷新頁面重試。');
            scrollToBottom();
        };
    }
    
    // 發送加入聊天室訊息
    function sendJoinRoomMessage() {
        if (socket && socket.readyState === WebSocket.OPEN) {
            const message = {
                type: 'join_room',
                target: roomId
            };
            socket.send(JSON.stringify(message));
        }
    }
    
    // 發送訊息
    function sendMessage() {
        const content = messageInput.value.trim();
        
        if (!content) return;
        
        if (socket && socket.readyState === WebSocket.OPEN) {
            const message = {
                content: content,
                sender: username,
                time: Math.floor(Date.now() / 1000)
            };
            
            socket.send(JSON.stringify(message));
            messageInput.value = '';
        } else {
            addSystemMessage('連接已關閉，無法發送訊息。');
        }
    }
    
    // 離開聊天室
    function leaveRoom() {
        if (socket && socket.readyState === WebSocket.OPEN) {
            const message = {
                type: 'leave_room'
            };
            socket.send(JSON.stringify(message));
            socket.close();
        }
    }
    
    // 添加訊息到聊天區域
    function addMessage(sender, content, time) {
        const messageElement = document.createElement('div');
        messageElement.className = `message ${sender === username ? 'sent' : 'received'} mb-3`;
        
        let messageHTML = '';
        
        if (sender === username) {
            // 自己發送的訊息
            messageHTML = `
                <div class="d-flex flex-column align-items-end">
                    <div class="sender text-end text-primary">${escapeHtml(sender)}</div>
                    <div class="content bg-primary text-white rounded-3">${escapeHtml(content)}</div>
                    <div class="time">${formatTime(time)}</div>
                </div>
            `;
        } else {
            // 接收到的訊息
            messageHTML = `
                <div class="d-flex flex-column align-items-start">
                    <div class="sender text-secondary">${escapeHtml(sender)}</div>
                    <div class="content bg-light rounded-3">${escapeHtml(content)}</div>
                    <div class="time">${formatTime(time)}</div>
                </div>
            `;
        }
        
        messageElement.innerHTML = messageHTML;
        chatMessages.appendChild(messageElement);
    }
    
    // 添加系統訊息
    function addSystemMessage(content) {
        const messageElement = document.createElement('div');
        messageElement.className = 'system-message';
        messageElement.innerHTML = `<i class="bi bi-info-circle me-1"></i> ${escapeHtml(content)}`;
        
        chatMessages.appendChild(messageElement);
    }
    
    // 更新在線用戶數功能暫時註解掉
    /*
    function updateOnlineCount(count) {
        onlineCount.innerHTML = `<i class="bi bi-people-fill"></i> ${count} 人在線`;
    }
    */
    
    // 更新用戶列表功能暫時註解掉
    /*
    function updateUsersList(users) {
        usersList.innerHTML = '';
        
        if (users.length === 0) {
            usersList.innerHTML = `
                <li class="list-group-item text-center text-muted">
                    <small>沒有活躍用戶</small>
                </li>
            `;
            return;
        }
        
        users.forEach(user => {
            if (user.IsActive) {
                const userElement = document.createElement('li');
                userElement.className = 'list-group-item d-flex align-items-center';
                
                // 如果是當前用戶，加上標記
                if (user.UserID === username) {
                    userElement.innerHTML = `
                        <i class="bi bi-person-fill me-2 text-primary"></i>
                        <span class="fw-bold">${escapeHtml(user.UserID || '匿名')}</span>
                        <span class="badge bg-primary ms-auto">您</span>
                    `;
                } else {
                    userElement.innerHTML = `
                        <i class="bi bi-person me-2"></i>
                        ${escapeHtml(user.UserID || '匿名')}
                    `;
                }
                
                usersList.appendChild(userElement);
            }
        });
    }
    */
    
    // 滾動到底部
    function scrollToBottom() {
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
    
    // 格式化時間
    function formatTime(date) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    
    // HTML 轉義函數，防止 XSS 攻擊
    function escapeHtml(unsafe) {
        if (!unsafe) return '';
        return unsafe
            .replace(/&/g, "&amp;")
            .replace(/</g, "&lt;")
            .replace(/>/g, "&gt;")
            .replace(/"/g, "&quot;")
            .replace(/'/g, "&#039;");
    }
    
    // 頁面關閉前離開聊天室
    window.addEventListener('beforeunload', function() {
        leaveRoom();
    });
});