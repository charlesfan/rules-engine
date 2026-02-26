"""
Form Schema Overview Specification.

General structure and common attributes for form fields.
"""

FORM_OVERVIEW_SPEC = """## form_schema 概述

定義報名表單的欄位，前端會根據此結構動態生成表單。

### 基本結構

```json
{
  "form_schema": {
    "fields": [
      { 欄位1 },
      { 欄位2 },
      ...
    ]
  }
}
```

### 通用欄位屬性

| 屬性 | 必要性 | 說明 |
|------|--------|------|
| `id` | ✅ 必要 | 欄位唯一識別碼 |
| `label` | ✅ 必要 | 顯示標籤 |
| `type` | ✅ 必要 | 欄位類型 |
| `field` | ✅ 必要 | 對應的 context 路徑 |
| `required` | ⭕ 選擇 | 是否必填，預設 false |
| `placeholder` | ⭕ 選擇 | 提示文字 |
| `default_value` | ⭕ 選擇 | 預設值 |
| `description` | ⭕ 選擇 | 欄位說明文字 |

### 欄位類型

| 類型 | 說明 | 特有屬性 |
|------|------|---------|
| `text` | 文字輸入 | placeholder |
| `email` | Email 輸入 | placeholder |
| `number` | 數字輸入 | min, max |
| `select` | 下拉選單 | options |
| `checkbox` | 核取方塊 | - |

### field 路徑對應

| field 路徑 | Context 位置 | 說明 |
|-----------|-------------|------|
| `user.name` | context.user.name | 姓名 |
| `user.email` | context.user.email | Email |
| `user.age` | context.user.age | 年齡 |
| `user.gender` | context.user.gender | 性別 |
| `user.race_type` | context.user.race_type | 賽事組別 |
| `team_size` | context.team_size | 團報人數 |

### 最小範例

```json
{
  "form_schema": {
    "fields": [
      {"id": "name", "label": "姓名", "type": "text", "field": "user.name", "required": true},
      {"id": "email", "label": "Email", "type": "email", "field": "user.email", "required": true}
    ]
  }
}
```
"""
