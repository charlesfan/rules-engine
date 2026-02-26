"""
Group Discount Pattern.

Complete example for implementing group registration discounts.
"""

GROUP_DISCOUNT_PATTERN = """## 團報優惠模式

完整的團報優惠設定範例，包含人數條件和固定金額折扣。

### 基本團報優惠（5人以上）

5人以上團報，每人減100元：

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
      "id": "group_discount",
      "priority": 100,
      "description": "團報優惠（5人以上）",
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
  ],
  "form_schema": {
    "fields": [
      {
        "id": "team_size",
        "label": "團報人數",
        "type": "number",
        "field": "team_size",
        "default_value": 1,
        "min": 1,
        "max": 100,
        "description": "5人以上可享團報優惠，每人減100元"
      }
    ]
  }
}
```

### 階梯式團報優惠

3-4人減50、5-9人減100、10人以上減150：

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
      "id": "group_discount_large",
      "priority": 98,
      "description": "大團優惠（10人以上減150）",
      "condition": {
        "type": "compare",
        "field": "team_size",
        "operator": ">=",
        "value": 10
      },
      "action": {
        "type": "fixed_discount",
        "target": "registration_fee",
        "value": 150,
        "label": "大團優惠"
      }
    },
    {
      "id": "group_discount_medium",
      "priority": 99,
      "description": "中團優惠（5-9人減100）",
      "condition": {
        "type": "and",
        "conditions": [
          {"type": "compare", "field": "team_size", "operator": ">=", "value": 5},
          {"type": "compare", "field": "team_size", "operator": "<=", "value": 9}
        ]
      },
      "action": {
        "type": "fixed_discount",
        "target": "registration_fee",
        "value": 100,
        "label": "中團優惠"
      }
    },
    {
      "id": "group_discount_small",
      "priority": 100,
      "description": "小團優惠（3-4人減50）",
      "condition": {
        "type": "and",
        "conditions": [
          {"type": "compare", "field": "team_size", "operator": ">=", "value": 3},
          {"type": "compare", "field": "team_size", "operator": "<=", "value": 4}
        ]
      },
      "action": {
        "type": "fixed_discount",
        "target": "registration_fee",
        "value": 50,
        "label": "小團優惠"
      }
    }
  ]
}
```

### 團報 + 早鳥雙重優惠

早鳥9折，團報5人以上再減100（可疊加）：

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
      "id": "early_bird",
      "priority": 100,
      "description": "早鳥優惠9折",
      "condition": {
        "type": "datetime_before",
        "field": "register_date",
        "value": "2026-03-31T23:59:59+08:00"
      },
      "action": {
        "type": "percentage_discount",
        "target": "registration_fee",
        "value": 10,
        "label": "早鳥優惠"
      }
    },
    {
      "id": "group_discount",
      "priority": 101,
      "description": "團報優惠（5人以上）",
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
  ],
  "discount_stacking": {
    "mode": "additive",
    "description": "早鳥折扣後再減團報金額"
  }
}
```

### 重要提醒

1. **team_size 欄位**：需要在 form_schema 中定義團報人數欄位
2. **Priority 順序**：較大折扣的 priority 數字較小（先判斷）
3. **階梯優惠**：使用 and 組合 >= 和 <= 條件來定義範圍
"""
