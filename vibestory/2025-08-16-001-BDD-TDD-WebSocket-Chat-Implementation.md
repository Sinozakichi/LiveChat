---
id: "vibestory-2025-08-16-001"
related_feature: "livechat-websocket-communication"
datetime: "2025-08-16T19:45:00+08:00"
stage: "feature-complete"
tags:
  - feature
  - websocket
  - architecture
  - testing
  - BDD
  - TDD
summary: "使用 BDD 與 TDD 方法論實作即時聊天功能，包含多層架構設計與完整測試覆蓋"
---

## What happened

此次完成了「使用 BDD 與 TDD 方法論實作 WebSocket 即時聊天功能」的開發工作。  
主要工作包含：專案結構重構為分層架構、實作 BDD 功能場景與步驟定義、按照 TDD 流程為每一層撰寫單元測試、實作整合測試，以及前後端程式碼分離。

具體完成的內容有：
1. 設計並實作 BDD 功能場景（features/livechat.feature）
2. 實作 BDD 步驟定義（backend/tests/livechat_steps.go）
3. 按照 TDD 流程實作 model、repository、service、handler 四層架構
4. 為每一層撰寫完整的單元測試
5. 實作端對端的整合測試
6. 前端代碼分離為 HTML、CSS、JS 三部分

## Why it's designed like this

1. **分層架構（Layered Architecture）**：採用 handler、service、repository、model 四層架構，實現關注點分離，提高代碼可維護性和可測試性。
   - Handler 層：處理 HTTP/WebSocket 請求與回應
   - Service 層：實現業務邏輯
   - Repository 層：管理數據存取
   - Model 層：定義數據模型

2. **依賴注入（Dependency Injection）**：各層之間通過接口而非具體實現進行依賴，提高代碼的鬆耦合性和可測試性。例如 WebSocketHandler 依賴 BroadcastService 接口而非具體實現。

3. **BDD 場景設計**：基於用戶視角設計功能場景，確保開發的功能符合實際需求。場景包括：
   - 使用者連接到聊天室
   - 發送與接收訊息
   - 多用戶交流
   - 處理錯誤情況（空訊息）
   - 伺服器重啟恢復

4. **TDD 測試驅動開發**：嚴格遵循 Red-Green-Refactor 流程，先撰寫測試，再實現功能，最後重構代碼。確保每個功能都有對應的測試覆蓋。

5. **前後端分離**：將前端代碼分為 HTML 結構、CSS 樣式和 JS 邏輯，提高可維護性和可讀性。

## Reflection

原本計劃直接使用簡單的 WebSocket 實現，但在開發過程中發現需要更完善的錯誤處理和連接管理機制。特別是在實現多用戶交流和伺服器重啟場景時，需要更嚴謹的設計。

在測試方面，最初嘗試直接模擬 WebSocket 連接，但由於 websocket.Conn 是具體類型而非接口，導致測試困難。後來改用依賴注入和接口設計，使測試變得更加容易。

整合測試中遇到的主要挑戰是處理 WebSocket 的異步特性，需要小心處理訊息的發送和接收順序，以及連接的建立和關閉。

BDD 與 TDD 結合的開發方式雖然前期投入較大，但顯著提高了代碼質量和可維護性。特別是在重構過程中，完整的測試套件提供了強大的安全網，使我們能夠自信地進行代碼改進。

## Next

- 實現用戶認證機制，允許用戶使用名稱登入
- 添加聊天室功能，支援多個獨立的聊天空間
- 實現訊息持久化，使用數據庫儲存歷史訊息
- 添加訊息類型支援（文字、圖片、檔案等）
- 優化前端界面，提供更好的使用者體驗
- 實現訊息送達確認機制
- 考慮使用 Redis Pub/Sub 支援多實例部署
