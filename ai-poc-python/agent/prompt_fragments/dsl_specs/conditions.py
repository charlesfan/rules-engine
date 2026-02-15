"""
Condition Types DSL Specification.
"""

CONDITIONS_PROMPT = """## Condition 類型

條件用於 pricing_rules 和 validation_rules 的觸發判斷。

### equals - 等於
```json
{"type": "equals", "field": "user.race_type", "value": "full_marathon"}
```

### compare - 比較運算
支援運算子：`>`, `<`, `>=`, `<=`, `==`, `!=`
```json
{"type": "compare", "field": "team_size", "operator": ">=", "value": 5}
```

### datetime_before / datetime_after - 時間比較
```json
{"type": "datetime_before", "field": "register_date", "value": "2025-03-31T23:59:59+08:00"}
{"type": "datetime_after", "field": "register_date", "value": "2025-01-01T00:00:00+08:00"}
```

### field_empty - 欄位為空
```json
{"type": "field_empty", "field": "user.race_type"}
```

### field_exists - 欄位存在
```json
{"type": "field_exists", "field": "team"}
```

### and - 且（所有條件都要滿足）
```json
{
  "type": "and",
  "conditions": [
    {"type": "compare", "field": "team_size", "operator": ">=", "value": 3},
    {"type": "compare", "field": "team_size", "operator": "<=", "value": 5}
  ]
}
```

### or - 或（任一條件滿足）
```json
{
  "type": "or",
  "conditions": [
    {"type": "equals", "field": "user.race_type", "value": "full_marathon"},
    {"type": "equals", "field": "user.race_type", "value": "half_marathon"}
  ]
}
```

### always_true - 永遠為真
```json
{"type": "always_true"}
```

### 常見組合範例

#### 早鳥 + 團報條件
```json
{
  "type": "and",
  "conditions": [
    {"type": "datetime_before", "field": "register_date", "value": "2025-03-31T23:59:59+08:00"},
    {"type": "compare", "field": "team_size", "operator": ">=", "value": 3}
  ]
}
```"""
