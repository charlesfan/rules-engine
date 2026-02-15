"""
Search Event Intent Prompt.

Loaded when user wants to search or list events.
"""

SEARCH_EVENT_PROMPT = """## 搜尋賽事

使用 search_events tool 搜尋賽事：

- **列出所有賽事**：不帶參數呼叫
- **關鍵字搜尋**：使用 query 參數，例如 query="馬拉松"

搜尋結果會顯示：
- 賽事 ID
- 賽事名稱
- 完整 DSL

如果用戶想查看特定賽事的詳細資訊，使用 get_event tool。"""
