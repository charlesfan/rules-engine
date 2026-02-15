"""
DSL Overview Prompt.

Basic structure of DSL JSON format with required/optional field specifications.
"""

DSL_OVERVIEW_PROMPT = """## DSL 完整規格

DSL 是 JSON 格式的規則定義。以下是完整結構：

```json
{
  "event_id": "unique-event-id",
  "version": "1.0",
  "name": "賽事名稱",
  "variables": {...},
  "data_sources": {...},
  "rule_definitions": {...},
  "computed_fields": {...},
  "pricing_rules": [...],
  "validation_rules": [...],
  "discount_stacking": {...},
  "form_schema": {...}
}
```

### 必要欄位 vs 選擇欄位

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `event_id` | ✅ 必要 | 唯一英文識別碼，如 "marathon-2026" |
| `version` | ✅ 必要 | 版本號，通常為 "1.0" |
| `name` | ✅ 必要 | 賽事顯示名稱 |
| `pricing_rules` | ✅ 必要 | 定價規則陣列（至少要有基本價格規則）|
| `validation_rules` | ✅ 必要 | 驗證規則陣列（可為空陣列 `[]`）|
| `form_schema` | ⚠️ 建議 | 表單欄位定義（前端渲染需要）|
| `variables` | ⭕ 選擇 | 變數定義，可在規則中用 `$variables.xxx` 引用 |
| `discount_stacking` | ⭕ 選擇 | 折扣計算方式，預設為 additive |
| `data_sources` | ⭕ 選擇 | 外部資料源（如車主名單）|
| `rule_definitions` | ⭕ 選擇 | 可重用的條件定義 |
| `computed_fields` | ⭕ 選擇 | 計算欄位定義 |

### 最小可用 DSL 範例

```json
{
  "event_id": "simple-event-2026",
  "version": "1.0",
  "name": "簡單活動",
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "基本報名費",
      "condition": {"type": "always_true"},
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 500,
        "label": "報名費"
      }
    }
  ],
  "validation_rules": [],
  "form_schema": {
    "fields": [
      {"id": "name", "label": "姓名", "type": "text", "field": "user.name", "required": true}
    ]
  }
}
```

### discount_stacking 模式

| 模式 | 說明 |
|------|------|
| `additive` | 折扣累加（預設）：原價 - 折扣A - 折扣B |
| `multiplicative` | 折扣乘法：原價 × (1-折扣A%) × (1-折扣B%) |
| `best_only` | 只取最優惠的折扣 |"""
