"""
Delete Event Intent Prompt.

Loaded when user wants to delete an event.
"""

DELETE_EVENT_PROMPT = """## 刪除賽事

刪除賽事前務必確認：

1. **確認目標**：先使用 get_event 確認是正確的賽事
2. **二次確認**：向用戶確認「確定要刪除 XXX 賽事嗎？此操作無法復原。」
3. **執行刪除**：確認後使用 delete_event tool

### 確認話術範例

「您確定要刪除「2026大湖馬拉松」嗎？刪除後無法復原。請回覆「確定」以繼續。」"""
