"""
Checkbox Field Specification.

Used for boolean input fields.
"""

CHECKBOX_FIELD_SPEC = """## checkbox - 核取方塊欄位

用於是/否的選擇，如同意條款、加購選項等。

### 結構

```json
{
  "id": "agree_terms",
  "label": "我同意活動條款",
  "type": "checkbox",
  "field": "user.agree_terms",
  "required": true
}
```

### 特有屬性

| 屬性 | 說明 |
|------|------|
| `default_value` | 預設值（true/false）|

### 常見範例

**同意條款**
```json
{
  "id": "agree_terms",
  "label": "我已閱讀並同意活動條款與個資聲明",
  "type": "checkbox",
  "field": "user.agree_terms",
  "required": true
}
```

**加購選項**
```json
{
  "id": "want_tshirt",
  "label": "加購紀念 T-shirt（NT$350）",
  "type": "checkbox",
  "field": "user.want_tshirt",
  "required": false,
  "default_value": false
}
```

**保險選項**
```json
{
  "id": "want_insurance",
  "label": "加購意外險（NT$100）",
  "type": "checkbox",
  "field": "user.want_insurance",
  "required": false,
  "default_value": false,
  "description": "建議購買，保障您的運動安全"
}
```

**會員身份**
```json
{
  "id": "is_member",
  "label": "我是俱樂部會員（享會員優惠）",
  "type": "checkbox",
  "field": "user.is_member",
  "required": false,
  "default_value": false
}
```

**宅配選項**
```json
{
  "id": "want_shipping",
  "label": "選擇宅配寄送（運費 NT$150）",
  "type": "checkbox",
  "field": "user.want_shipping",
  "required": false,
  "default_value": false,
  "description": "不勾選則需於活動當天現場領取"
}
```

### 搭配 pricing_rules

checkbox 欄位常與 add_item 或 discount 規則搭配：

```json
{
  "id": "tshirt_addon",
  "priority": 10,
  "description": "加購 T-shirt",
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
"""
