"""
percentage_discount Action Specification.

Used for percentage-based discounts like early bird, member discounts.
"""

PERCENTAGE_DISCOUNT_SPEC = """## percentage_discount - 百分比折扣

用於打折優惠，如早鳥9折、會員85折。value 是折扣百分比（10 = 打9折）。

### 結構

```json
{
  "id": "規則唯一ID",
  "priority": 100,
  "description": "規則描述",
  "condition": { ... },
  "action": {
    "type": "percentage_discount",
    "target": "registration_fee",
    "value": 折扣百分比,
    "label": "顯示名稱"
  }
}
```

### Action 欄位

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `type` | ✅ 必要 | 固定為 "percentage_discount" |
| `target` | ✅ 必要 | 要打折的項目，通常為 "registration_fee" |
| `value` | ✅ 必要 | 折扣百分比（10 = 九折，15 = 八五折）|
| `label` | ✅ 必要 | 顯示給用戶的名稱 |

### 折扣計算方式

| value 值 | 實際折扣 | 計算方式 |
|---------|---------|---------|
| 10 | 9折 | 原價 × 0.9 |
| 15 | 85折 | 原價 × 0.85 |
| 20 | 8折 | 原價 × 0.8 |

### 範例：早鳥優惠 9 折

```json
{
  "id": "early_bird_discount",
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
}
```

### 範例：會員 85 折

```json
{
  "id": "member_discount",
  "priority": 100,
  "description": "會員優惠85折",
  "condition": {
    "type": "equals",
    "field": "user.is_member",
    "value": true
  },
  "action": {
    "type": "percentage_discount",
    "target": "registration_fee",
    "value": 15,
    "label": "會員優惠"
  }
}
```

### Priority 建議

discount 規則的 priority 建議設為 100+，在基本價格和加購項目之後計算。
"""
