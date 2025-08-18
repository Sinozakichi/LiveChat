# LiveChat

一個使用 Go 語言和 WebSocket 技術構建的即時聊天應用。

## 專案概述

LiveChat 是一個簡單的即時聊天應用，使用 WebSocket 技術實現即時通訊功能。用戶可以在瀏覽器中打開應用，輸入訊息並即時與其他在線用戶交流。

## 技術架構

### 後端 (Go)
- 使用 Go 語言和 Gorilla WebSocket 套件實現 WebSocket 服務器
- 採用分層架構設計：
  - Handler：處理 HTTP 請求與回應
  - Service：處理業務邏輯
  - Repository：管理數據
  - Model：定義數據模型

### 前端
- HTML5 + CSS3 + JavaScript
- 使用原生 WebSocket API 實現即時通訊

## 專案結構

```
LiveChat/
├── backend/
│   ├── handler/       # 處理 HTTP 請求與回應
│   ├── model/         # 定義數據模型
│   ├── repository/    # 數據管理
│   └── service/       # 業務邏輯
├── frontend/
│   ├── css/           # 樣式表
│   ├── js/            # JavaScript 腳本
│   └── index.html     # 主頁面
├── main.go            # 應用入口點
└── go.mod             # Go 模組定義
```

## 功能特點

- 即時訊息傳遞
- 多用戶聊天
- 簡潔直觀的用戶界面

## 運行方式

1. 確保已安裝 Go 環境
2. 克隆此倉庫
3. 執行 `go run main.go`
4. 在瀏覽器中訪問 `http://localhost:8080`

## 部署

本應用已配置為可在 Fly.io 平台上部署。