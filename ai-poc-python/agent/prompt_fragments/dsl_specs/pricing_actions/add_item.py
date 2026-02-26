"""
add_item Action Specification.

Used to add additional items like shipping fee, insurance, merchandise.
"""

ADD_ITEM_SPEC = """## add_item - 新增項目

用於新增額外項目，如運費、保險、加購商品等。

### 結構

```json
{
  "id": "規則唯一ID",
  "priority": 10,
  "description": "規則描述",
  "condition": { ... },
  "action": {
    "type": "add_item",
    "item": "addon:項目ID",
    "fixed_price": 價格數值,
    "label": "顯示名稱"
  }
}
```

### Action 欄位

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `type` | ✅ 必要 | 固定為 "add_item" |
| `item` | ✅ 必要 | 項目識別碼，建議以 "addon:" 開頭 |
| `fixed_price` | ✅ 必要 | 固定價格數值 |
| `label` | ✅ 必要 | 顯示給用戶的名稱 |

### 範例：宅配運費

```json
{
  "id": "shipping_fee",
  "priority": 10,
  "description": "宅配運費",
  "condition": {
    "type": "equals",
    "field": "user.delivery_method",
    "value": "shipping"
  },
  "action": {
    "type": "add_item",
    "item": "addon:shipping",
    "fixed_price": 150,
    "label": "宅配運費"
  }
}
```

### 範例：保險費

```json
{
  "id": "insurance_fee",
  "priority": 10,
  "description": "意外險",
  "condition": {
    "type": "equals",
    "field": "user.want_insurance",
    "value": true
  },
  "action": {
    "type": "add_item",
    "item": "addon:insurance",
    "fixed_price": 100,
    "label": "意外險"
  }
}
```

### 範例：紀念 T-shirt 加購

```json
{
  "id": "tshirt_addon",
  "priority": 10,
  "description": "紀念 T-shirt",
  "condition": {
    "type": "equals",
    "field": "user.want_tshirt",
    "value": true
  },
  "action": {
    "type": "add_item",
    "item": "addon:tshirt",
    "fixed_price": 350,
    "label": "紀念 T-shirt"
  }
}
```

### Priority 建議

add_item 規則的 priority 建議設為 10-99，在基本價格之後計算。
"""
