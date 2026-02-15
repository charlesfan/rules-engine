"""
Form Schema DSL Specification.
"""

FORM_SCHEMA_PROMPT = """## form_schema 規格

定義報名表單的欄位。

```json
{
  "fields": [
    {
      "id": "name",
      "label": "姓名",
      "type": "text",
      "field": "user.name",
      "required": true,
      "placeholder": "請輸入姓名"
    },
    {
      "id": "email",
      "label": "Email",
      "type": "email",
      "field": "user.email",
      "required": true
    },
    {
      "id": "age",
      "label": "年齡",
      "type": "number",
      "field": "user.age",
      "required": true,
      "min": 1,
      "max": 120
    },
    {
      "id": "race_type",
      "label": "賽事組別",
      "type": "select",
      "field": "user.race_type",
      "required": true,
      "options": [
        {"label": "請選擇組別", "value": ""},
        {"label": "全程馬拉松 - NT$1,050", "value": "full_marathon"},
        {"label": "半程馬拉松 - NT$950", "value": "half_marathon"}
      ]
    },
    {
      "id": "team_size",
      "label": "團體報名人數",
      "type": "number",
      "field": "team_size",
      "default_value": 1,
      "min": 1,
      "max": 100
    }
  ]
}
```

### 欄位屬性

| 屬性 | 必要性 | 說明 |
|------|--------|------|
| `id` | ✅ 必要 | 欄位唯一識別碼 |
| `label` | ✅ 必要 | 顯示標籤 |
| `type` | ✅ 必要 | 欄位類型 |
| `field` | ✅ 必要 | 對應的 context 路徑 |
| `required` | ⭕ 選擇 | 是否必填，預設 false |
| `placeholder` | ⭕ 選擇 | 提示文字 |
| `default_value` | ⭕ 選擇 | 預設值 |
| `min` / `max` | ⭕ 選擇 | 數字範圍（number 類型）|
| `options` | ⚠️ 條件必要 | 選項（select/radio 類型必填）|
| `validation` | ⭕ 選擇 | 額外驗證規則 |

### 欄位類型

| 類型 | 說明 | 特有屬性 |
|------|------|---------|
| text | 文字輸入 | placeholder |
| email | Email 輸入 | placeholder |
| number | 數字輸入 | min, max, default_value |
| select | 下拉選單 | options |
| checkbox | 核取方塊 | default_value (boolean) |

### field 路徑對應

| field 路徑 | Context 位置 |
|-----------|-------------|
| user.name | context.user.name |
| user.email | context.user.email |
| user.race_type | context.user.race_type |
| team_size | context.team_size |"""
