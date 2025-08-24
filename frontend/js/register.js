document.addEventListener('DOMContentLoaded', function() {
    const registerForm = document.getElementById('register-form');
    const errorMessage = document.getElementById('error-message');

    registerForm.addEventListener('submit', function(event) {
        event.preventDefault();
        
        const username = document.getElementById('username').value;
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;

        // 隱藏錯誤訊息
        errorMessage.classList.add('d-none');

        // 客戶端驗證
        if (username.length < 3 || username.length > 20) {
            errorMessage.textContent = '用戶名長度必須在 3-20 個字符之間';
            errorMessage.classList.remove('d-none');
            return;
        }

        const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        if (!emailRegex.test(email)) {
            errorMessage.textContent = '請輸入有效的電子郵件地址';
            errorMessage.classList.remove('d-none');
            return;
        }

        if (password.length < 8) {
            errorMessage.textContent = '密碼長度至少為 8 位';
            errorMessage.classList.remove('d-none');
            return;
        }

        // 檢查密碼強度
        let hasUpper = false, hasLower = false, hasDigit = false;
        for (let i = 0; i < password.length; i++) {
            const char = password.charAt(i);
            if (char >= 'A' && char <= 'Z') hasUpper = true;
            else if (char >= 'a' && char <= 'z') hasLower = true;
            else if (char >= '0' && char <= '9') hasDigit = true;
        }

        // 至少滿足兩個條件
        if (!((hasUpper && hasLower) || (hasUpper && hasDigit) || (hasLower && hasDigit))) {
            errorMessage.textContent = '密碼必須包含大小寫字母、數字至少其2者';
            errorMessage.classList.remove('d-none');
            return;
        }

        // 發送註冊請求
        fetch('/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                email: email,
                password: password
            })
        })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => {
                    throw new Error(data.error || '註冊失敗');
                });
            }
            return response.json();
        })
        .then(data => {
            // 註冊成功，重定向到聊天室選擇頁面
            window.location.href = '/rooms.html';
        })
        .catch(error => {
            // 顯示錯誤訊息
            errorMessage.textContent = error.message;
            errorMessage.classList.remove('d-none');
        });
    });
});
