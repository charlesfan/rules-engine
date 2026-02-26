"""
Field Check Condition Specification.

Used for checking field existence or emptiness.
"""

FIELD_CHECK_SPEC = """## 欄位檢查 - field_empty / field_exists

用於檢查欄位是否存在或為空，常用於驗證規則。

---

### field_empty - 欄位為空

當欄位不存在或值為空時為 true。

```json
{
  "type": "field_empty",
  "field": "欄位路徑"
}
```

#### 範例：組別必填驗證

```json
{
  "id": "race_type_required",
  "description": "必須選擇組別",
  "condition": {
    "type": "field_empty",
    "field": "user.race_type"
  },
  "error_type": "blocking",
  "error_message": "請選擇賽事組別"
}
```

#### 範例：姓名必填驗證

```json
{
  "id": "name_required",
  "description": "姓名必填",
  "condition": {
    "type": "field_empty",
    "field": "user.name"
  },
  "error_type": "blocking",
  "error_message": "請輸入姓名"
}
```

---

### field_exists - 欄位存在

當欄位存在且有值時為 true。

```json
{
  "type": "field_exists",
  "field": "欄位路徑"
}
```

#### 範例：有團隊資料時顯示團隊欄位

```json
{
  "type": "field_exists",
  "field": "team"
}
```

#### 範例：有推薦碼時套用折扣

```json
{
  "id": "referral_discount",
  "priority": 100,
  "description": "推薦碼折扣",
  "condition": {
    "type": "field_exists",
    "field": "user.referral_code"
  },
  "action": {
    "type": "fixed_discount",
    "target": "registration_fee",
    "value": 50,
    "label": "推薦優惠"
  }
}
```

---

### 常用於驗證規則

field_empty 最常用於 validation_rules 的必填欄位檢查：

```json
{
  "validation_rules": [
    {
      "id": "name_required",
      "condition": {"type": "field_empty", "field": "user.name"},
      "error_type": "blocking",
      "error_message": "請輸入姓名"
    },
    {
      "id": "email_required",
      "condition": {"type": "field_empty", "field": "user.email"},
      "error_type": "blocking",
      "error_message": "請輸入 Email"
    },
    {
      "id": "race_type_required",
      "condition": {"type": "field_empty", "field": "user.race_type"},
      "error_type": "blocking",
      "error_message": "請選擇賽事組別"
    }
  ]
}
```
"""
