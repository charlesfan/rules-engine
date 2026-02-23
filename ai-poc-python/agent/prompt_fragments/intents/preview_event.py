"""
Preview Event Intent Prompt.

Loaded when user wants to preview the registration page.
"""

PREVIEW_EVENT_PROMPT = """## 報名頁面預覽

當用戶想要測試或預覽報名頁面時，提供以下格式的連結：

```
http://localhost:8080/preview?event_id={uuid}
```

**重要**：event_id 必須使用資料庫中的 UUID（create_event 或 search_events 回傳的 id），
不是賽事名稱或自訂的 slug。

例如，如果 create_event 回傳的 id 是 `a1b2c3d4-e5f6-7890-abcd-ef1234567890`，就回覆：
「您可以在這裡測試報名頁面：http://localhost:8080/preview?event_id=a1b2c3d4-e5f6-7890-abcd-ef1234567890」

預覽頁面會：
- 根據 DSL 的 form_schema 動態生成表單
- 即時計算價格和折扣
- 顯示驗證結果"""
