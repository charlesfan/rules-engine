"""
fixed_discount Action Specification.

Used for fixed amount discounts like group discounts.
"""

FIXED_DISCOUNT_SPEC = """## fixed_discount - 固定金額折扣

用於固定金額減免，如團報每人減100元、滿額折扣。

### 結構

```json
{
  "id": "規則唯一ID",
  "priority": 100,
  "description": "規則描述",
  "condition": { ... },
  "action": {
    "type": "fixed_discount",
    "target": "registration_fee",
    "value": 折扣金額,
    "label": "顯示名稱"
  }
}
```

### Action 欄位

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `type` | ✅ 必要 | 固定為 "fixed_discount" |
| `target` | ✅ 必要 | 要折扣的項目，通常為 "registration_fee" |
| `value` | ✅ 必要 | 折扣金額（正整數）|
| `label` | ✅ 必要 | 顯示給用戶的名稱 |

### 範例：團報優惠（5人以上每人減100）

```json
{
  "id": "group_discount_5",
  "priority": 100,
  "description": "團報5人以上每人減100",
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

### 範例：階梯式團報優惠

```json
[
  {
    "id": "group_discount_3",
    "priority": 100,
    "description": "團報3-4人每人減50",
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
  },
  {
    "id": "group_discount_5",
    "priority": 99,
    "description": "團報5人以上每人減100",
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
      "label": "大團優惠"
    }
  }
]
```

### Priority 建議

- discount 規則的 priority 建議設為 100+
- 階梯式優惠中，較大折扣的 priority 數字較小（先判斷）
"""
