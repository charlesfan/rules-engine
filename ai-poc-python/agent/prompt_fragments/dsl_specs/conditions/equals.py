"""
equals Condition Specification.

Used for exact value matching.
"""

EQUALS_SPEC = """## equals - 等於比較

用於精確值比對，常用於組別選擇、布林值判斷。

### 結構

```json
{
  "type": "equals",
  "field": "欄位路徑",
  "value": "比對值"
}
```

### 欄位說明

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `type` | ✅ 必要 | 固定為 "equals" |
| `field` | ✅ 必要 | 要比對的欄位路徑 |
| `value` | ✅ 必要 | 要比對的值（字串、數字或布林值）|

### 範例：組別選擇

```json
{"type": "equals", "field": "user.race_type", "value": "full_marathon"}
```

### 範例：布林值判斷

```json
{"type": "equals", "field": "user.is_member", "value": true}
```

### 範例：性別判斷

```json
{"type": "equals", "field": "user.gender", "value": "female"}
```

### 常用 field 路徑

| 路徑 | 說明 |
|------|------|
| `user.race_type` | 賽事組別 |
| `user.gender` | 性別 |
| `user.is_member` | 是否會員 |
| `user.delivery_method` | 物流方式 |
"""
