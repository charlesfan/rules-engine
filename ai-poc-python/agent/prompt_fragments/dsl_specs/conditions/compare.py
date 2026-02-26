"""
compare Condition Specification.

Used for numeric comparisons.
"""

COMPARE_SPEC = """## compare - 數值比較

用於數值大小比較，常用於團報人數、年齡限制。

### 結構

```json
{
  "type": "compare",
  "field": "欄位路徑",
  "operator": "比較運算子",
  "value": 比較數值
}
```

### 欄位說明

| 欄位 | 必要性 | 說明 |
|------|--------|------|
| `type` | ✅ 必要 | 固定為 "compare" |
| `field` | ✅ 必要 | 要比較的欄位路徑 |
| `operator` | ✅ 必要 | 比較運算子 |
| `value` | ✅ 必要 | 比較的數值 |

### 支援的運算子

| 運算子 | 說明 | 範例 |
|--------|------|------|
| `>` | 大於 | team_size > 5 |
| `<` | 小於 | age < 18 |
| `>=` | 大於等於 | team_size >= 3 |
| `<=` | 小於等於 | age <= 70 |
| `==` | 等於 | team_size == 1 |
| `!=` | 不等於 | team_size != 0 |

### 範例：團報人數判斷

```json
{"type": "compare", "field": "team_size", "operator": ">=", "value": 5}
```

### 範例：年齡下限

```json
{"type": "compare", "field": "user.age", "operator": ">=", "value": 18}
```

### 範例：年齡上限

```json
{"type": "compare", "field": "user.age", "operator": "<=", "value": 70}
```

### 常用 field 路徑

| 路徑 | 說明 |
|------|------|
| `team_size` | 團報人數 |
| `user.age` | 年齡 |
| `quantity` | 數量 |
"""
