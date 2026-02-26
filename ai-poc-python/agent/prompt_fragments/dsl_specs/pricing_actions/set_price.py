"""
set_price Action Specification.

Used to set base registration fee for different race categories.
"""

SET_PRICE_SPEC = """## set_price - 設定基本價格

用於設定報名費、組別價格。每個定價組別需要一個 set_price 規則。

### 結構

```json
{
  "id": "規則唯一ID",
  "priority": 0,
  "description": "規則描述",
  "condition": { ... },
  "action": {
    "type": "set_price",
    "item": "registration_fee",
    "value": 價格數值,
    "label": "顯示名稱"
  }
}
```

### Action 欄位

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `type` | ✅ 必要 | 固定為 "set_price" |
| `item` | ✅ 必要 | 項目識別碼，通常為 "registration_fee" |
| `value` | ✅ 必要 | 價格數值（整數）|
| `label` | ✅ 必要 | 顯示給用戶的名稱 |

### 範例：全馬報名費

```json
{
  "id": "full_marathon_fee",
  "priority": 0,
  "description": "全馬報名費",
  "condition": {
    "type": "equals",
    "field": "user.race_type",
    "value": "full_marathon"
  },
  "action": {
    "type": "set_price",
    "item": "registration_fee",
    "value": 1500,
    "label": "全馬報名費"
  }
}
```

### 範例：半馬報名費

```json
{
  "id": "half_marathon_fee",
  "priority": 0,
  "description": "半馬報名費",
  "condition": {
    "type": "equals",
    "field": "user.race_type",
    "value": "half_marathon"
  },
  "action": {
    "type": "set_price",
    "item": "registration_fee",
    "value": 1200,
    "label": "半馬報名費"
  }
}
```

### Priority 建議

set_price 規則的 priority 建議設為 0-9，確保基本價格最先計算。
"""
