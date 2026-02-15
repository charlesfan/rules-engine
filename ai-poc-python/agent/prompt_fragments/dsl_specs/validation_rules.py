"""
Validation Rules DSL Specification.
"""

VALIDATION_RULES_PROMPT = """## validation_rules 規格

驗證規則用於檢查報名資料是否有效。

```json
{
  "id": "race_type_required",
  "description": "必須選擇賽事組別",
  "condition": {
    "type": "field_empty",
    "field": "user.race_type"
  },
  "error_type": "blocking",
  "error_message": "請選擇賽事組別"
}
```

### 欄位說明

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `id` | ✅ 必要 | 規則唯一識別碼 |
| `description` | ✅ 必要 | 規則描述 |
| `condition` | ✅ 必要 | 觸發條件（當條件為 true 時觸發錯誤）|
| `error_type` | ✅ 必要 | "blocking"（阻止報名）或 "warning"（警告但可繼續）|
| `error_message` | ✅ 必要 | 顯示給用戶的錯誤訊息 |

### 常見驗證範例

#### 必填欄位
```json
{
  "id": "name_required",
  "description": "姓名必填",
  "condition": {
    "type": "field_empty",
    "field": "user.name"
  },
  "error_type": "blocking",
  "error_message": "請輸入姓名"
}
```

#### 數值範圍
```json
{
  "id": "age_range",
  "description": "年齡限制",
  "condition": {
    "type": "or",
    "conditions": [
      {"type": "compare", "field": "user.age", "operator": "<", "value": 18},
      {"type": "compare", "field": "user.age", "operator": ">", "value": 70}
    ]
  },
  "error_type": "blocking",
  "error_message": "參賽者年齡須介於 18-70 歲"
}
```"""
