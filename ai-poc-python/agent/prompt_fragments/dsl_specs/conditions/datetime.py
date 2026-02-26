"""
datetime Condition Specification.

Used for date/time comparisons, commonly for early bird discounts.
"""

DATETIME_SPEC = """## datetime_before / datetime_after - 時間比較

用於時間判斷，常用於早鳥優惠、報名期限。

### datetime_before 結構

當目前時間「早於」指定時間時為 true。

```json
{
  "type": "datetime_before",
  "field": "register_date",
  "value": "ISO 8601 時間字串"
}
```

### datetime_after 結構

當目前時間「晚於」指定時間時為 true。

```json
{
  "type": "datetime_after",
  "field": "register_date",
  "value": "ISO 8601 時間字串"
}
```

### 欄位說明

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `type` | ✅ 必要 | "datetime_before" 或 "datetime_after" |
| `field` | ✅ 必要 | 通常為 "register_date" |
| `value` | ✅ 必要 | ISO 8601 格式的時間字串（含時區）|

### 時間格式

使用 ISO 8601 格式，包含時區：
- `2026-03-31T23:59:59+08:00` - 台灣時區
- `2026-03-31T15:59:59Z` - UTC 時間

### 範例：早鳥優惠（3月底前）

```json
{
  "type": "datetime_before",
  "field": "register_date",
  "value": "2026-03-31T23:59:59+08:00"
}
```

### 範例：晚鳥加價（6月後）

```json
{
  "type": "datetime_after",
  "field": "register_date",
  "value": "2026-06-01T00:00:00+08:00"
}
```

### 範例：特定時段優惠（3月到5月）

使用 `and` 組合兩個時間條件：

```json
{
  "type": "and",
  "conditions": [
    {"type": "datetime_after", "field": "register_date", "value": "2026-03-01T00:00:00+08:00"},
    {"type": "datetime_before", "field": "register_date", "value": "2026-05-31T23:59:59+08:00"}
  ]
}
```
"""
