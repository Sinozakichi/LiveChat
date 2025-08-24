# Vibestory 規範

---
title: "Vibestory 規範"
version: "1.1"
created: "2025-08-05"
updated: "2025-08-22"
---

## 📖 Vibestory 是什麼？

**Vibestory** 是一種針對開發流程中「心流記錄（flow state log）」與「設計與反思軌跡」的敘事格式。

### 設計目的

- 作為 AI Agent 的語意入口與行為上下文
- 作為 Feature 開發過程中的行為說明、動機紀錄與心智模型映射
- 替未來的自我或其他開發者提供 feature-level 的知識脈絡
- 對比單純 Git Commit Message 的薄弱語意層，vibestory 能更貼近開發者的思路與上下文

## ✏️ Vibestory 的格式規範

### 記錄時機

- 每當 `/features` 下的一個 featurefile 被完整的實作完成時

### 記錄路徑

- 請創建一個新的 markdown 檔案，並將其放置於 `/vibestory` 路徑下

### 檔名規範

檔名格式：YYYY-MM-DD-No-Title.md
- YYYY: 年份
- MM: 月份
- DD: 日期
- No: 序號
- Title: 標題（由 AI Agent 依此次 vibestory 生成過程命名一合適標題）

### 文件結構

每一篇 Vibestory 均採用 **YAML Frontmatter + Markdown Description** 的格式，結構範例如下：

#### 📦 YAML Frontmatter

```yaml
id: "vibestory-2025-08-05-001"
related_feature: "user-can-send-message"
datetime: "2025-08-05T14:30:00+08:00"
stage: "feature-complete"
tags:
  - feature
  - websocket
  - architecture
summary: "使用者能夠成功透過 WebSocket 發送訊息給所有其他在線使用者"
```

#### 內容區段

##### What happened

此次完成了「使用者能夠透過 WebSocket 傳送訊息到所有聊天室成員」的實作。  
這包括在 server 建立 WebSocket hub、client 設定 socket 傳送/接收邏輯、整合 Redis Pub/Sub 機制。

##### Why it's designed like this

1. **WebSocket Hub Pattern**：為了支援多聊天室與 broadcast message，採用了 centralized Hub pattern
2. **Redis Pub/Sub**：為了將訊息從一個 pod 廣播到所有 pod，整合了 Redis Pub/Sub
3. **Message Structuring**：訊息格式採用 JSON，並附加 senderId, roomId, timestamp 作為基本 metadata

##### Reflection

原本以為可以直接在 Go Server 裡保留所有連線，但考量到之後部署上會是多 pod，因此必須引入 Pub/Sub 模型。這部分踩了一些坑，例如 goroutine 泄漏、channel 沒有 close 的問題。

##### Next

- 加入訊息持久化邏輯（可能選擇 SQLite 或 MySQL）
- 考慮 message queue 的重試機制與錯誤處理

## 📖 Vibestory 脈絡延續原則

Vibestory 不僅是單一功能開發的記錄，更是專案演進的連續性敘事。每篇 Vibestory 應當與前篇建立明確的脈絡關聯，展現專案的演進軌跡。

### 脈絡延續要求

1. **前後關聯**：每篇 Vibestory 應明確參考前篇中的 "Next" 部分，說明哪些計劃已被實現，哪些仍在進行中。

2. **進展追蹤**：清楚記錄專案從上一篇 Vibestory 至今的整體進展，而非僅關注當前開發的單一功能。

3. **全面性**：涵蓋所有重要的架構變更、功能新增、技術升級等，即使這些變更分屬不同的功能領域。

4. **技術債務追蹤**：記錄在開發過程中發現的技術債務，以及如何處理或計劃處理這些債務。

### 撰寫指南

1. **回顧前篇**：在撰寫新的 Vibestory 前，應先閱讀前一篇 Vibestory，特別是其 "Next" 部分。

2. **進度對照**：在 "What happened" 部分，應明確指出哪些前篇提及的計劃已經實現。

3. **整合視角**：不應將不同功能或技術變更割裂處理，而應從整體架構和系統演進的角度進行敘述。

4. **設計連續性**：在 "Why it's designed like this" 部分，應說明當前設計如何延續或改進了前篇中的設計理念。

5. **反思深化**：在 "Reflection" 部分，應包含對前篇中提出的挑戰和解決方案的再評估。

## 📋 Vibestory 內容完整性檢查清單

為確保 Vibestory 的完整性和質量，請在提交前檢查以下項目：

- [ ] 是否明確參考了前篇 Vibestory 中的 "Next" 計劃？
- [ ] 是否涵蓋了從上一篇至今的所有重要變更？
- [ ] 是否從整體架構和系統演進的角度進行敘述？
- [ ] 是否說明了當前設計與前篇設計理念的連續性？
- [ ] 是否包含對前篇挑戰和解決方案的再評估？
- [ ] "Next" 部分是否為未來開發提供了明確的方向？
- [ ] 是否使用了正確的 YAML Frontmatter 格式和必要字段？
- [ ] 是否按照規定的文件結構組織內容？
- [ ] 是否使用了恰當的標籤來分類本次開發工作？