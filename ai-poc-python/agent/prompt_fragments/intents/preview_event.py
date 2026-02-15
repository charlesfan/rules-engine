"""
Preview Event Intent Prompt.

Loaded when user wants to preview the registration page.
"""

PREVIEW_EVENT_PROMPT = """## 報名頁面預覽

當用戶想要測試或預覽報名頁面時，提供以下格式的連結：

```
http://localhost:8080/preview?event_id={event_id}
```

例如，如果賽事的 event_id 是 `marathon-2026`，就回覆：
「您可以在這裡測試報名頁面：http://localhost:8080/preview?event_id=marathon-2026」

預覽頁面會：
- 根據 DSL 的 form_schema 動態生成表單
- 即時計算價格和折扣
- 顯示驗證結果"""
