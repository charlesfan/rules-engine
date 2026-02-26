"""
Select Field Specification.

Used for dropdown select fields with predefined options.
"""

SELECT_FIELD_SPEC = """## select - 下拉選單欄位

用於預定義選項的選擇，如賽事組別、性別、T-shirt 尺寸等。

### 結構

```json
{
  "id": "race_type",
  "label": "賽事組別",
  "type": "select",
  "field": "user.race_type",
  "required": true,
  "options": [
    {"label": "顯示文字", "value": "實際值"},
    ...
  ]
}
```

### 特有屬性

| 屬性 | 說明 |
|------|------|
| `options` | 選項陣列（必填）|

### options 結構

```json
{
  "label": "顯示給用戶的文字",
  "value": "實際儲存的值"
}
```

### 常見範例

**賽事組別**（重要：options 的 value 要對應 pricing_rules 中的值）
```json
{
  "id": "race_type",
  "label": "賽事組別",
  "type": "select",
  "field": "user.race_type",
  "required": true,
  "options": [
    {"label": "請選擇組別", "value": ""},
    {"label": "全程馬拉松 (42K) - NT$1,500", "value": "full_marathon"},
    {"label": "半程馬拉松 (21K) - NT$1,200", "value": "half_marathon"},
    {"label": "迷你馬拉松 (10K) - NT$800", "value": "mini_marathon"}
  ]
}
```

**性別**
```json
{
  "id": "gender",
  "label": "性別",
  "type": "select",
  "field": "user.gender",
  "required": true,
  "options": [
    {"label": "請選擇", "value": ""},
    {"label": "男", "value": "male"},
    {"label": "女", "value": "female"}
  ]
}
```

**T-shirt 尺寸**
```json
{
  "id": "tshirt_size",
  "label": "T-shirt 尺寸",
  "type": "select",
  "field": "user.tshirt_size",
  "required": false,
  "options": [
    {"label": "請選擇", "value": ""},
    {"label": "S", "value": "S"},
    {"label": "M", "value": "M"},
    {"label": "L", "value": "L"},
    {"label": "XL", "value": "XL"},
    {"label": "2XL", "value": "2XL"}
  ]
}
```

**物流方式**
```json
{
  "id": "delivery_method",
  "label": "物資領取方式",
  "type": "select",
  "field": "user.delivery_method",
  "required": true,
  "options": [
    {"label": "現場領取（免運費）", "value": "pickup"},
    {"label": "宅配到府（運費 $150）", "value": "shipping"}
  ]
}
```

### 重要提醒

1. **第一個選項**：建議設為空值的提示，如 `{"label": "請選擇", "value": ""}`
2. **與 pricing_rules 對應**：race_type 的 value 必須與 pricing_rules 中 condition 的 value 一致
3. **價格顯示**：建議在 label 中顯示價格，幫助用戶選擇
"""
