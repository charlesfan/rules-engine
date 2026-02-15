"""
Pricing Rules DSL Specification.

Complete specification for pricing_rules including all action types.
"""

PRICING_RULES_PROMPT = """## pricing_rules 規格

每個 pricing_rule 的欄位：

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `id` | ✅ 必要 | 規則唯一識別碼 |
| `priority` | ✅ 必要 | 優先順序（數字越小越先執行）|
| `description` | ✅ 必要 | 規則描述 |
| `condition` | ✅ 必要 | 觸發條件 |
| `action` | ✅ 必要 | 執行動作 |

### Action 類型

#### 1. set_price - 設定基本價格
```json
{
  "id": "full_marathon_registration",
  "priority": 0,
  "description": "全程馬拉松報名費",
  "condition": {
    "type": "equals",
    "field": "user.race_type",
    "value": "full_marathon"
  },
  "action": {
    "type": "set_price",
    "item": "registration_fee",
    "value": 1050,
    "label": "全程馬拉松報名費"
  }
}
```

#### 2. add_item - 新增項目（如運費、保險）
```json
{
  "id": "shipping_fee",
  "priority": 10,
  "description": "宅配運費",
  "condition": {
    "type": "compare",
    "field": "team_size",
    "operator": ">=",
    "value": 1
  },
  "action": {
    "type": "add_item",
    "item": "addon:shipping",
    "fixed_price": 150,
    "label": "宅配運費"
  }
}
```

#### 3. percentage_discount - 百分比折扣
```json
{
  "id": "early_bird_discount",
  "priority": 100,
  "description": "早鳥優惠9折",
  "condition": {
    "type": "datetime_before",
    "field": "register_date",
    "value": "2025-03-31T23:59:59+08:00"
  },
  "action": {
    "type": "percentage_discount",
    "target": "registration_fee",
    "value": 10,
    "label": "早鳥優惠"
  }
}
```

#### 4. fixed_discount - 固定金額折扣
```json
{
  "id": "group_discount",
  "priority": 100,
  "description": "團報優惠",
  "condition": {
    "type": "compare",
    "field": "team_size",
    "operator": ">=",
    "value": 5
  },
  "action": {
    "type": "fixed_discount",
    "target": "registration_fee",
    "value": 100,
    "label": "團報優惠"
  }
}
```

### Priority 建議

| 類型 | Priority 範圍 |
|------|--------------|
| set_price | 0-9 |
| add_item | 10-99 |
| discount | 100+ |"""
