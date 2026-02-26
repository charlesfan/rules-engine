"""
Age Restriction Pattern.

Complete example for implementing age-based validation rules.
"""

AGE_RESTRICTION_PATTERN = """## 年齡限制模式

完整的年齡驗證設定範例，包含年齡欄位和驗證規則。

### 基本年齡限制（18歲以上）

```json
{
  "validation_rules": [
    {
      "id": "age_minimum",
      "description": "年齡限制18歲以上",
      "condition": {
        "type": "compare",
        "field": "user.age",
        "operator": "<",
        "value": 18
      },
      "error_type": "blocking",
      "error_message": "參賽者須年滿18歲"
    }
  ],
  "form_schema": {
    "fields": [
      {
        "id": "age",
        "label": "年齡",
        "type": "number",
        "field": "user.age",
        "required": true,
        "min": 1,
        "max": 120,
        "description": "須年滿18歲方可報名"
      }
    ]
  }
}
```

### 年齡範圍限制（18-70歲）

```json
{
  "validation_rules": [
    {
      "id": "age_range",
      "description": "年齡限制18-70歲",
      "condition": {
        "type": "or",
        "conditions": [
          {"type": "compare", "field": "user.age", "operator": "<", "value": 18},
          {"type": "compare", "field": "user.age", "operator": ">", "value": 70}
        ]
      },
      "error_type": "blocking",
      "error_message": "參賽者須介於18-70歲"
    }
  ],
  "form_schema": {
    "fields": [
      {
        "id": "age",
        "label": "年齡",
        "type": "number",
        "field": "user.age",
        "required": true,
        "min": 18,
        "max": 70,
        "description": "參賽者須介於18-70歲"
      }
    ]
  }
}
```

### 不同組別不同年齡限制

全馬需20歲以上，半馬16歲以上，迷你馬不限：

```json
{
  "validation_rules": [
    {
      "id": "full_marathon_age",
      "description": "全馬年齡限制20歲以上",
      "condition": {
        "type": "and",
        "conditions": [
          {"type": "equals", "field": "user.race_type", "value": "full_marathon"},
          {"type": "compare", "field": "user.age", "operator": "<", "value": 20}
        ]
      },
      "error_type": "blocking",
      "error_message": "全程馬拉松參賽者須年滿20歲"
    },
    {
      "id": "half_marathon_age",
      "description": "半馬年齡限制16歲以上",
      "condition": {
        "type": "and",
        "conditions": [
          {"type": "equals", "field": "user.race_type", "value": "half_marathon"},
          {"type": "compare", "field": "user.age", "operator": "<", "value": 16}
        ]
      },
      "error_type": "blocking",
      "error_message": "半程馬拉松參賽者須年滿16歲"
    }
  ]
}
```

### 年齡警告（非阻擋）

70歲以上顯示警告但可繼續報名：

```json
{
  "validation_rules": [
    {
      "id": "age_warning",
      "description": "高齡參賽者警告",
      "condition": {
        "type": "compare",
        "field": "user.age",
        "operator": ">",
        "value": 70
      },
      "error_type": "warning",
      "error_message": "建議70歲以上參賽者先諮詢醫師，確認身體狀況適合參加"
    }
  ]
}
```

### 長青組優惠（年齡折扣）

60歲以上享敬老優惠：

```json
{
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "基本報名費",
      "condition": {"type": "always_true"},
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1500,
        "label": "報名費"
      }
    },
    {
      "id": "senior_discount",
      "priority": 100,
      "description": "敬老優惠8折",
      "condition": {
        "type": "compare",
        "field": "user.age",
        "operator": ">=",
        "value": 60
      },
      "action": {
        "type": "percentage_discount",
        "target": "registration_fee",
        "value": 20,
        "label": "敬老優惠"
      }
    }
  ]
}
```

### 重要提醒

1. **error_type**：`blocking` 阻止報名，`warning` 只顯示警告
2. **validation 邏輯**：condition 為 true 時觸發錯誤，所以用 `<` 檢查年齡不足
3. **前端驗證**：form_schema 的 min/max 只做前端初步驗證，完整驗證靠 validation_rules
"""
