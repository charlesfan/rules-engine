"""
Logic Condition Specification.

Used for combining conditions (and, or) and always_true.
"""

LOGIC_SPEC = """## 邏輯條件 - and / or / always_true

用於組合多個條件或建立永遠成立的條件。

---

### and - 且（所有條件都要滿足）

```json
{
  "type": "and",
  "conditions": [
    { 條件1 },
    { 條件2 },
    ...
  ]
}
```

#### 範例：3-5人團報

```json
{
  "type": "and",
  "conditions": [
    {"type": "compare", "field": "team_size", "operator": ">=", "value": 3},
    {"type": "compare", "field": "team_size", "operator": "<=", "value": 5}
  ]
}
```

#### 範例：早鳥 + 會員

```json
{
  "type": "and",
  "conditions": [
    {"type": "datetime_before", "field": "register_date", "value": "2026-03-31T23:59:59+08:00"},
    {"type": "equals", "field": "user.is_member", "value": true}
  ]
}
```

---

### or - 或（任一條件滿足）

```json
{
  "type": "or",
  "conditions": [
    { 條件1 },
    { 條件2 },
    ...
  ]
}
```

#### 範例：全馬或半馬

```json
{
  "type": "or",
  "conditions": [
    {"type": "equals", "field": "user.race_type", "value": "full_marathon"},
    {"type": "equals", "field": "user.race_type", "value": "half_marathon"}
  ]
}
```

#### 範例：VIP 或大團（10人以上）

```json
{
  "type": "or",
  "conditions": [
    {"type": "equals", "field": "user.is_vip", "value": true},
    {"type": "compare", "field": "team_size", "operator": ">=", "value": 10}
  ]
}
```

---

### always_true - 永遠為真

用於無條件套用的規則，常用於單一組別或統一價格。

```json
{"type": "always_true"}
```

#### 範例：統一報名費

```json
{
  "id": "base_price",
  "priority": 0,
  "description": "統一報名費",
  "condition": {"type": "always_true"},
  "action": {
    "type": "set_price",
    "item": "registration_fee",
    "value": 800,
    "label": "報名費"
  }
}
```

---

### 巢狀組合

and 和 or 可以巢狀使用：

```json
{
  "type": "and",
  "conditions": [
    {"type": "datetime_before", "field": "register_date", "value": "2026-03-31T23:59:59+08:00"},
    {
      "type": "or",
      "conditions": [
        {"type": "equals", "field": "user.is_member", "value": true},
        {"type": "compare", "field": "team_size", "operator": ">=", "value": 3}
      ]
    }
  ]
}
```

上述條件表示：「早鳥期間」且「會員或3人以上團報」
"""
