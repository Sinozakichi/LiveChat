@user_loginregister
Feature: 用戶登入與註冊功能
  作為一個網站訪客
  我希望能夠註冊新帳號並登入系統
  以便能夠使用聊天室功能並保持個人身份

  Scenario: 訪客查看登入頁面
    Given 訪客訪問網站首頁
    Then 訪客應該看到登入表單
    And 登入表單應包含用戶名和密碼欄位
    And 登入頁面應有"註冊新帳號"的連結

  Scenario: 訪客查看註冊頁面
    Given 訪客訪問網站首頁
    When 訪客點擊"註冊新帳號"連結
    Then 訪客應該看到註冊表單
    And 註冊表單應包含用戶名、電子郵件和密碼欄位

  Scenario: 訪客成功註冊新帳號
    Given 訪客在註冊頁面
    When 訪客填寫有效的用戶名"testuser"
    And 訪客填寫有效的電子郵件"test@example.com"
    And 訪客填寫有效的密碼"Password123!"
    And 訪客點擊註冊按鈕
    Then 系統應創建新用戶
    And 系統應對密碼進行加密後再存入資料庫
    And 訪客應自動登入系統
    And 訪客應被導向到聊天室選擇頁面

  Scenario: 密碼應該被加密存儲
    Given 訪客在註冊頁面
    When 訪客填寫有效的用戶名"secureuser"
    And 訪客填寫有效的電子郵件"secure@example.com"
    And 訪客填寫有效的密碼"SecurePass123!"
    And 訪客點擊註冊按鈕
    Then 系統應創建新用戶
    And 系統應對密碼進行加密後再存入資料庫
    And 資料庫中的密碼不應該是明文"SecurePass123!"
    And 系統應能夠驗證正確的密碼"SecurePass123!"

  Scenario: 訪客嘗試使用已存在的用戶名註冊
    Given 系統中已存在用戶名為"existinguser"的帳號
    And 訪客在註冊頁面
    When 訪客填寫用戶名"existinguser"
    And 訪客填寫有效的電子郵件"new@example.com"
    And 訪客填寫有效的密碼"Password123!"
    And 訪客點擊註冊按鈕
    Then 訪客應看到錯誤訊息"用戶名已被使用"
    And 訪客應留在註冊頁面

  Scenario: 訪客嘗試使用已存在的電子郵件註冊
    Given 系統中已存在電子郵件為"existing@example.com"的帳號
    And 訪客在註冊頁面
    When 訪客填寫有效的用戶名"newuser"
    And 訪客填寫電子郵件"existing@example.com"
    And 訪客填寫有效的密碼"Password123!"
    And 訪客點擊註冊按鈕
    Then 訪客應看到錯誤訊息"電子郵件已被使用"
    And 訪客應留在註冊頁面

  Scenario: 訪客嘗試使用弱密碼註冊
    Given 訪客在註冊頁面
    When 訪客填寫有效的用戶名"testuser"
    And 訪客填寫有效的電子郵件"test@example.com"
    And 訪客填寫弱密碼"123456"
    And 訪客點擊註冊按鈕
    Then 訪客應看到錯誤訊息"密碼必須包含大小寫字母、數字至少其2者，且長度至少為8位"
    And 訪客應留在註冊頁面

  Scenario: 用戶成功登入
    Given 系統中已存在用戶名為"existinguser"和密碼為"Password123!"的帳號
    And 訪客在登入頁面
    When 訪客填寫用戶名"existinguser"
    And 訪客填寫密碼"Password123!"
    And 訪客點擊登入按鈕
    Then 系統應驗證加密後的密碼
    And 訪客應成功登入系統
    And 系統應創建包含用戶資訊的Session
    And 訪客應被導向到聊天室選擇頁面

  Scenario: 用戶使用錯誤密碼登入
    Given 系統中已存在用戶名為"existinguser"和密碼為"Password123!"的帳號
    And 訪客在登入頁面
    When 訪客填寫用戶名"existinguser"
    And 訪客填寫密碼"wrongpassword"
    And 訪客點擊登入按鈕
    Then 系統應驗證加密後的密碼
    And 訪客應看到錯誤訊息"用戶名或密碼錯誤"
    And 訪客應留在登入頁面

  Scenario: 用戶使用不存在的用戶名登入
    Given 訪客在登入頁面
    When 訪客填寫用戶名"nonexistentuser"
    And 訪客填寫密碼"Password123!"
    And 訪客點擊登入按鈕
    Then 訪客應看到錯誤訊息"用戶名或密碼錯誤"
    And 訪客應留在登入頁面

  Scenario: 已登入用戶訪問聊天室選擇頁面
    Given 用戶已成功登入系統
    When 用戶訪問聊天室選擇頁面
    Then 用戶應看到可用的聊天室列表
    And 用戶名應顯示在頁面上

  Scenario: 已登入的管理員用戶可以創建新聊天室
    Given 具有管理員角色的用戶已登入系統
    When 用戶訪問聊天室選擇頁面
    Then 用戶應看到"創建新聊天室"按鈕
    When 用戶點擊"創建新聊天室"按鈕
    And 用戶填寫聊天室名稱和描述
    And 用戶點擊提交按鈕
    Then 系統應創建新聊天室
    And 新聊天室應出現在聊天室列表中

  Scenario: 已登入的普通用戶無法創建新聊天室
    Given 具有普通用戶角色的用戶已登入系統
    When 用戶訪問聊天室選擇頁面
    Then 用戶不應看到"創建新聊天室"按鈕與介面

  Scenario: 用戶登入後加入聊天室不需再輸入用戶名
    Given 用戶已成功登入系統
    When 用戶訪問聊天室選擇頁面
    And 用戶點擊某個聊天室
    Then 系統應檢查用戶的Session資料
    And 系統應從Session中獲取用戶名
    And 用戶應直接進入聊天室
    And 用戶不應被要求輸入用戶名
    And 用戶名應顯示在聊天室介面中

  Scenario: 用戶登入後系統創建Session
    Given 系統中已存在用戶名為"testuser"和電子郵件為"test@example.com"的帳號
    When 用戶成功登入系統
    Then 系統應創建新的Session
    And Session應包含用戶名"testuser"
    And Session應包含用戶電子郵件"test@example.com"
    And Session應包含用戶角色資訊
    And Session應設置過期時間無上限

  Scenario: 用戶登出系統
    Given 用戶已成功登入系統
    When 用戶點擊"登出"按鈕
    Then 用戶應成功登出系統
    And 用戶應被重定向到登入頁面
    And 用戶的Session應被銷毀