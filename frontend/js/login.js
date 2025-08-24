document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('login-form');
    const errorMessage = document.getElementById('error-message');

    loginForm.addEventListener('submit', function(event) {
        event.preventDefault();
        
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;

        // 隱藏錯誤訊息
        errorMessage.classList.add('d-none');

        // 發送登入請求
        fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => {
                    throw new Error(data.error || '登入失敗');
                });
            }
            return response.json();
        })
        .then(data => {
            // 登入成功，重定向到聊天室選擇頁面
            window.location.href = '/rooms.html';
        })
        .catch(error => {
            // 顯示錯誤訊息
            errorMessage.textContent = error.message;
            errorMessage.classList.remove('d-none');
        });
    });
});
