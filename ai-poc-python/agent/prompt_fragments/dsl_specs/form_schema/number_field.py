"""
Number Field Specification.

Used for numeric input with optional min/max validation.
"""

NUMBER_FIELD_SPEC = """## number - 數字輸入欄位

用於需要輸入數字的欄位，如年齡、人數等。

### 結構

```json
{
  "id": "age",
  "label": "年齡",
  "type": "number",
  "field": "user.age",
  "required": true,
  "min": 1,
  "max": 120
}
```

### 特有屬性

| 屬性 | 說明 |
|------|------|
| `min` | 最小值 |
| `max` | 最大值 |
| `default_value` | 預設值（數字）|
| `step` | 步進值（選填）|

### 常見範例

**年齡欄位**
```json
{
  "id": "age",
  "label": "年齡",
  "type": "number",
  "field": "user.age",
  "required": true,
  "min": 1,
  "max": 120,
  "placeholder": "請輸入年齡"
}
```

**團報人數**
```json
{
  "id": "team_size",
  "label": "團報人數",
  "type": "number",
  "field": "team_size",
  "required": false,
  "default_value": 1,
  "min": 1,
  "max": 100,
  "description": "3人以上可享團報優惠"
}
```

**身高（公分）**
```json
{
  "id": "height",
  "label": "身高 (cm)",
  "type": "number",
  "field": "user.height",
  "required": false,
  "min": 100,
  "max": 250,
  "description": "用於配置合適尺寸的號碼布"
}
```

**衣服尺寸（用於計算）**
```json
{
  "id": "tshirt_quantity",
  "label": "加購 T-shirt 數量",
  "type": "number",
  "field": "user.tshirt_quantity",
  "required": false,
  "default_value": 0,
  "min": 0,
  "max": 5
}
```

### 注意事項

- `min` 和 `max` 會在前端做基本驗證
- 如需更複雜的驗證邏輯，請在 `validation_rules` 中定義
"""
