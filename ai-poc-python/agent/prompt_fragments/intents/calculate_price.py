"""
Calculate Price Intent Prompt.

Loaded when user wants to calculate registration price.
"""

CALCULATE_PRICE_PROMPT = """## 計算價格

使用 calculate_price tool 計算報名費用。

### Context 格式

```json
{
  "register_date": "2025-03-15T10:00:00+08:00",
  "user": {
    "name": "王小明",
    "email": "test@example.com",
    "race_type": "full_marathon"
  },
  "team_size": 1
}
```

### 常用欄位

| 欄位 | 說明 | 範例 |
|------|------|------|
| register_date | 報名時間（ISO 8601） | "2025-03-15T10:00:00+08:00" |
| user.race_type | 組別 | "full_marathon" |
| user.name | 姓名 | "王小明" |
| user.email | Email | "test@example.com" |
| team_size | 團體人數 | 5 |

### 回應格式

計算結果會包含：
- 各項目費用明細
- 已套用的折扣
- 最終價格
- 驗證結果（是否可以報名）"""
