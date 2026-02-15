"""
Update Event Intent Prompt.

Loaded when user wants to modify an existing event.
"""

UPDATE_EVENT_PROMPT = """## 修改賽事流程

當用戶想修改賽事時：

1. **確認目標賽事**：使用 get_event 取得目前的 DSL
2. **了解修改內容**：詢問用戶想修改什麼（價格、優惠、欄位等）
3. **確認修改**：在執行前向用戶確認變更內容
4. **執行修改**：使用 update_event tool 更新賽事

### 常見修改類型

- **修改價格**：更新 pricing_rules 中的 set_price action
- **新增優惠**：在 pricing_rules 加入 percentage_discount 或 fixed_discount
- **修改欄位**：更新 form_schema.fields
- **新增驗證**：在 validation_rules 加入新規則

### 注意事項

- 修改 pricing_rules 時，確保 form_schema 的 options 同步更新
- 新增折扣規則時，priority 通常設為 100（在基本價格之後執行）
- 修改前務必確認用戶意圖，避免誤改"""
