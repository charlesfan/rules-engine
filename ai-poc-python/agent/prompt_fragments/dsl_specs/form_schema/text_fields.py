"""
Text and Email Field Specification.

Used for text input and email input fields.
"""

TEXT_FIELDS_SPEC = """## text / email - 文字輸入欄位

### text - 一般文字輸入

```json
{
  "id": "name",
  "label": "姓名",
  "type": "text",
  "field": "user.name",
  "required": true,
  "placeholder": "請輸入姓名"
}
```

#### 特有屬性

| 屬性 | 說明 |
|------|------|
| `placeholder` | 輸入框提示文字 |
| `maxLength` | 最大字元數（選填）|
| `pattern` | 正則表達式驗證（選填）|

#### 常見範例

**姓名欄位**
```json
{
  "id": "name",
  "label": "姓名",
  "type": "text",
  "field": "user.name",
  "required": true,
  "placeholder": "請輸入真實姓名"
}
```

**緊急聯絡人**
```json
{
  "id": "emergency_contact",
  "label": "緊急聯絡人",
  "type": "text",
  "field": "user.emergency_contact",
  "required": true,
  "placeholder": "請輸入緊急聯絡人姓名"
}
```

**聯絡電話**
```json
{
  "id": "phone",
  "label": "聯絡電話",
  "type": "text",
  "field": "user.phone",
  "required": true,
  "placeholder": "0912-345-678"
}
```

---

### email - Email 輸入

```json
{
  "id": "email",
  "label": "Email",
  "type": "email",
  "field": "user.email",
  "required": true,
  "placeholder": "your@email.com"
}
```

會自動驗證 email 格式。

#### 常見範例

**Email 欄位**
```json
{
  "id": "email",
  "label": "電子郵件",
  "type": "email",
  "field": "user.email",
  "required": true,
  "placeholder": "example@email.com",
  "description": "賽事通知將寄送至此信箱"
}
```
"""
