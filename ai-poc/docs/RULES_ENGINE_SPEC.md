# Rules Engine DSL Specification

> 本文件詳細說明 rules engine 的 DSL 結構與執行邏輯，供 AI Prompt 開發參考。

## 目錄
1. [整體架構](#整體架構)
2. [DSL 結構 (RuleSet)](#dsl-結構-ruleset)
3. [Context 執行上下文](#context-執行上下文)
4. [Expression 條件表達式](#expression-條件表達式)
5. [Action 動作類型](#action-動作類型)
6. [Pricing Rules 定價規則](#pricing-rules-定價規則)
7. [Validation Rules 驗證規則](#validation-rules-驗證規則)
8. [Computed Fields 計算欄位](#computed-fields-計算欄位)
9. [Discount Stacking 折扣堆疊](#discount-stacking-折扣堆疊)
10. [價格計算流程](#價格計算流程)
11. [完整範例](#完整範例)

---

## 整體架構

```
┌─────────────────────────────────────────────────────────────┐
│                        DSL JSON                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Parser (parser.go)                                          │
│  - 解析 JSON 為 RuleSet 結構                                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Evaluator (evaluator.go)                                    │
│  - 評估 Expression 條件                                       │
│  - 執行 ValidationRules                                       │
│  - 呼叫 Calculator 計算價格                                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Calculator (calculator.go)                                  │
│  Phase 0: 評估 ComputedFields                                 │
│  Phase 1: 執行 set_price, add_item                           │
│  Phase 1.5: 重新評估 ComputedFields                           │
│  Phase 2: 執行折扣 (percentage_discount, fixed_discount)      │
│  Final: 計算最終價格                                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  EvaluationResult                                            │
│  - Valid: bool                                               │
│  - Price: PriceBreakdown                                     │
│  - Errors: []ValidationError                                 │
│  - Warnings: []ValidationError                               │
└─────────────────────────────────────────────────────────────┘
```

---

## DSL 結構 (RuleSet)

```json
{
  "event_id": "string (必填)",
  "version": "string",
  "name": "string",
  "variables": {},
  "rule_definitions": {},
  "computed_fields": {},
  "pricing_rules": [],
  "validation_rules": [],
  "discount_stacking": {},
  "form_schema": {}
}
```

### 各欄位說明

| 欄位 | 類型 | 必填 | 說明 |
|------|------|------|------|
| `event_id` | string | ✅ | 賽事識別碼 |
| `version` | string | | 版本號 |
| `name` | string | | 賽事名稱 |
| `variables` | map[string]any | | 變數定義，供規則引用 |
| `rule_definitions` | map[string]RuleDef | | 可重用的條件定義 |
| `computed_fields` | map[string]ComputedField | | 計算欄位定義 |
| `pricing_rules` | []PricingRule | | 定價規則陣列 |
| `validation_rules` | []ValidationRule | | 驗證規則陣列 |
| `discount_stacking` | DiscountStacking | | 折扣堆疊方式 |
| `form_schema` | FormSchema | | 表單欄位定義 |

---

## Context 執行上下文

Context 是執行規則時傳入的資料結構：

```go
type Context struct {
    RegisterDate   time.Time              // 報名日期時間
    User           map[string]interface{} // 使用者資料
    Team           map[string]interface{} // 團隊資料
    TeamSize       int                    // 團隊人數
    Addons         map[string]interface{} // 加購項目
    Variables      map[string]interface{} // 變數（自動從 RuleSet 載入）
    ComputedValues map[string]interface{} // 計算欄位的值（執行時填入）
}
```

### 常用欄位路徑

| 路徑 | 說明 | 範例值 |
|------|------|--------|
| `register_date` | 報名日期 | (time.Time) |
| `team_size` | 團隊人數 | `4` |
| `user.name` | 使用者姓名 | `"John"` |
| `user.age` | 使用者年齡 | `25` |
| `user.race_type` | 報名組別 | `"21K"`, `"full_marathon"` |
| `user.email` | 電子郵件 | `"john@example.com"` |
| `team.name` | 團隊名稱 | `"Team A"` |
| `team.members` | 團隊成員陣列 | `[{...}, {...}]` |
| `addons.insurance` | 是否加購保險 | `true` |
| `addons.tshirt` | 是否加購紀念衫 | `true` |
| `addons.tshirt_size` | 紀念衫尺寸 | `"L"` |
| `$computed.subtotal` | 計算欄位：小計 | (由 computed_fields 計算) |
| `$variables.price` | 變數引用 | (從 variables 取值) |

---

## Expression 條件表達式

Expression 用於定義條件，支援以下類型：

### 基本類型

#### `always_true`
總是成立。

```json
{ "type": "always_true" }
```

#### `equals`
欄位值等於指定值。

```json
{
  "type": "equals",
  "field": "user.race_type",
  "value": "21K"
}
```

#### `compare`
數值比較。

```json
{
  "type": "compare",
  "field": "user.age",
  "operator": ">=",    // 支援: >, <, >=, <=
  "value": 18
}
```

```json
{
  "type": "compare",
  "field": "team_size",
  "operator": ">=",
  "value": 5
}
```

### 日期時間類型

#### `datetime_before`
日期早於指定時間。

```json
{
  "type": "datetime_before",
  "field": "register_date",
  "value": "2026-02-28T23:59:59+08:00"
}
```

#### `datetime_after`
日期晚於指定時間。

```json
{
  "type": "datetime_after",
  "field": "register_date",
  "value": "2026-01-01T00:00:00+08:00"
}
```

#### `datetime_between`
日期在指定區間內（包含邊界）。

```json
{
  "type": "datetime_between",
  "field": "register_date",
  "start": "2026-01-01T00:00:00+08:00",
  "end": "2026-03-31T23:59:59+08:00"
}
```

### 邏輯組合類型

#### `and`
所有條件都成立。

```json
{
  "type": "and",
  "conditions": [
    { "type": "compare", "field": "team_size", "operator": ">=", "value": 3 },
    { "type": "compare", "field": "team_size", "operator": "<=", "value": 5 }
  ]
}
```

#### `or`
任一條件成立。

```json
{
  "type": "or",
  "conditions": [
    { "type": "equals", "field": "user.race_type", "value": "21K" },
    { "type": "equals", "field": "user.race_type", "value": "half_marathon" }
  ]
}
```

#### `not`
條件反轉。

```json
{
  "type": "not",
  "condition": {
    "type": "datetime_between",
    "field": "register_date",
    "start": "2026-01-01T00:00:00Z",
    "end": "2026-03-31T23:59:59Z"
  }
}
```

### 進階類型

#### `rule_ref`
引用 rule_definitions 中定義的規則。

```json
{
  "type": "rule_ref",
  "rule": "is_early_bird"
}
```

#### `field_exists`
欄位存在。

```json
{
  "type": "field_exists",
  "field": "user.email"
}
```

#### `field_empty`
欄位為空。

```json
{
  "type": "field_empty",
  "field": "user.phone"
}
```

#### `in_list`
值在外部資料源列表中。

```json
{
  "type": "in_list",
  "field": "user.car_plate",
  "list": "$data_sources.car_owners",
  "match_field": "plate_number"
}
```

#### `array_any`
陣列中任一元素符合條件。

```json
{
  "type": "array_any",
  "array": "team.members",
  "condition": {
    "type": "compare",
    "field": "user.age",
    "operator": "<",
    "value": 18
  }
}
```

#### `array_all`
陣列中所有元素都符合條件。

```json
{
  "type": "array_all",
  "array": "team.members",
  "condition": {
    "type": "compare",
    "field": "user.age",
    "operator": ">=",
    "value": 18
  }
}
```

---

## Action 動作類型

Action 定義規則觸發時的動作。

### `set_price`
設定價格項目。

```json
{
  "type": "set_price",
  "item": "registration_fee",
  "value": 1080,
  "label": "21K 報名費"
}
```

支援變數引用：
```json
{
  "type": "set_price",
  "item": "registration_fee",
  "value": "$variables.price_21k",
  "label": "21K 報名費"
}
```

| 欄位 | 說明 |
|------|------|
| `item` | 項目 ID（預設 `registration_fee`） |
| `value` | 價格（數字或 `$variables.xxx`） |
| `label` | 顯示標籤 |

### `add_item`
新增項目（支援數量計算）。

**以單價計算（unit_price × quantity）：**
```json
{
  "type": "add_item",
  "item": "addon:insurance",
  "unit_price": "$variables.insurance_price",
  "quantity_field": "team_size",
  "label": "團體保險"
}
```

**固定價格（不乘數量）：**
```json
{
  "type": "add_item",
  "item": "addon:shipping",
  "fixed_price": 150,
  "label": "運費"
}
```

| 欄位 | 說明 |
|------|------|
| `item` | 項目 ID（建議用 `addon:xxx` 格式） |
| `unit_price` | 單價（會乘以 quantity） |
| `fixed_price` | 固定價格（不乘數量） |
| `quantity_field` | 數量來源欄位（如 `team_size`） |
| `label` | 顯示標籤 |

### `percentage_discount`
百分比折扣。

```json
{
  "type": "percentage_discount",
  "value": 15,
  "apply_to": ["registration_fee"],
  "label": "早鳥 85 折"
}
```

**重要：`value` 是折扣百分比，不是折數！**
- 9 折 = `value: 10`（減 10%）
- 85 折 = `value: 15`（減 15%）
- 8 折 = `value: 20`（減 20%）

| 欄位 | 說明 |
|------|------|
| `value` | 折扣百分比（10 = 打九折） |
| `apply_to` | 套用目標（item ID 陣列，支援 `addon:*`、`total`） |
| `label` | 顯示標籤 |

### `fixed_discount`
固定金額折扣。

```json
{
  "type": "fixed_discount",
  "value": 200,
  "apply_to": ["total"],
  "label": "滿額折 200"
}
```

| 欄位 | 說明 |
|------|------|
| `value` | 折扣金額 |
| `apply_to` | 套用目標（`total` 會按比例分配到各項目） |
| `label` | 顯示標籤 |

### `price_cap`
價格封頂。

```json
{
  "type": "price_cap",
  "value": 5000,
  "apply_to": ["total"],
  "label": "價格上限 5000"
}
```

---

## Pricing Rules 定價規則

```json
{
  "id": "string (唯一識別碼)",
  "priority": 0,
  "description": "規則說明",
  "condition": { /* Expression */ },
  "action": { /* Action */ }
}
```

### Priority 執行順序

規則按 `priority` 由小到大執行：

| Priority 範圍 | 建議用途 |
|---------------|----------|
| 0-49 | 基本價格設定（set_price, add_item） |
| 50-99 | 加購項目 |
| 100-149 | 百分比折扣 |
| 150-199 | 固定金額折扣 |
| 200+ | 價格封頂 |

---

## Validation Rules 驗證規則

```json
{
  "id": "string",
  "description": "規則說明",
  "condition": { /* Expression - 觸發錯誤的條件 */ },
  "error_type": "blocking | warning",
  "error_message": "錯誤訊息"
}
```

| error_type | 說明 |
|------------|------|
| `blocking` | 阻止報名，result.Valid = false |
| `warning` | 警告但可繼續報名 |

**注意：condition 成立時會觸發錯誤，所以邏輯要反過來想！**

範例：年齡限制 18 歲
```json
{
  "id": "age_limit",
  "description": "年齡限制",
  "condition": {
    "type": "compare",
    "field": "user.age",
    "operator": "<",
    "value": 18
  },
  "error_type": "blocking",
  "error_message": "參賽者須年滿 18 歲"
}
```

---

## Computed Fields 計算欄位

計算欄位讓你可以基於價格項目或其他資料計算衍生值。

### `sum_prices`
加總指定項目的價格。

```json
{
  "subtotal": {
    "type": "sum_prices",
    "description": "小計",
    "items": ["registration_fee", "addon:*"]
  }
}
```

支援萬用字元 `*`（如 `addon:*` 匹配所有 addon 開頭的項目）。

### `count_items`
計算項目數量。

```json
{
  "addon_count": {
    "type": "count_items",
    "items": ["addon:*"]
  }
}
```

### `item_price`
取得單一項目價格。

```json
{
  "reg_fee": {
    "type": "item_price",
    "item": "registration_fee"
  }
}
```

### `count_array_field`
計算陣列中符合條件的數量。

```json
{
  "adult_count": {
    "type": "count_array_field",
    "array": "team.members",
    "field": "is_adult",
    "value": true
  }
}
```

### `sum_array_field`
加總陣列中指定欄位的值。

```json
{
  "total_age": {
    "type": "sum_array_field",
    "array": "team.members",
    "field": "age"
  }
}
```

### 在條件中使用計算欄位

```json
{
  "type": "compare",
  "field": "$computed.subtotal",
  "operator": ">=",
  "value": 2000
}
```

---

## Discount Stacking 折扣堆疊

```json
{
  "discount_stacking": {
    "mode": "multiplicative",
    "description": "折扣連乘計算"
  }
}
```

| Mode | 說明 | 計算方式 |
|------|------|----------|
| `multiplicative` | 連乘（預設） | 1000 × 0.9 × 0.95 = 855 |
| `additive` | 相加 | 1000 × (1 - 0.1 - 0.05) = 850 |
| `best_only` | 只取最優惠 | 取折扣金額最大的一個 |

**注意：目前 Calculator 實作為連乘模式，best_only 需額外實作。**

---

## 價格計算流程

```
Phase 0: 評估 ComputedFields（初始）
         ↓
Phase 1: 執行 set_price, add_item（建立 Items）
         計算 SubTotal
         ↓
Phase 1.5: 重新評估 ComputedFields（Items 已建立）
         ↓
Phase 2: 執行折扣（按 priority 順序）
         - percentage_discount: item.DiscountedPrice -= price × (value/100)
         - fixed_discount: 按比例分配到各項目
         - price_cap: 封頂價格
         ↓
Final:   計算 TotalDiscount = SubTotal - sum(DiscountedPrice)
         計算 FinalPrice = sum(DiscountedPrice)
```

### PriceBreakdown 結果結構

```go
type PriceBreakdown struct {
    Items         map[string]*PriceItem  // 所有項目
    Discounts     []DiscountItem         // 折扣明細
    SubTotal      float64                // 小計（原價總和）
    TotalDiscount float64                // 總折扣金額
    FinalPrice    float64                // 最終價格
}

type PriceItem struct {
    ID              string   // 項目 ID
    Name            string   // 顯示名稱
    Quantity        int      // 數量
    UnitPrice       float64  // 單價
    OriginalPrice   float64  // 原價（= 單價 × 數量）
    DiscountedPrice float64  // 折扣後價格
    FinalPrice      float64  // 最終價格
}
```

---

## 完整範例

### 範例 1：基本馬拉松報名

```json
{
  "event_id": "marathon-2025",
  "version": "1.0",
  "name": "2025 馬拉松",
  "variables": {
    "price_full": 1050,
    "price_half": 950
  },
  "pricing_rules": [
    {
      "id": "price_full_marathon",
      "priority": 0,
      "description": "全馬報名費",
      "condition": {
        "type": "equals",
        "field": "user.race_type",
        "value": "full_marathon"
      },
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": "$variables.price_full",
        "label": "全馬報名費"
      }
    },
    {
      "id": "price_half_marathon",
      "priority": 0,
      "description": "半馬報名費",
      "condition": {
        "type": "equals",
        "field": "user.race_type",
        "value": "half_marathon"
      },
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": "$variables.price_half",
        "label": "半馬報名費"
      }
    }
  ],
  "validation_rules": []
}
```

### 範例 2：早鳥折扣

```json
{
  "event_id": "race-2026",
  "version": "1.0",
  "name": "2026 路跑賽",
  "rule_definitions": {
    "is_early_bird": {
      "type": "condition",
      "description": "早鳥期間",
      "expression": {
        "type": "datetime_before",
        "field": "register_date",
        "value": "2026-02-28T23:59:59+08:00"
      }
    }
  },
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "基本報名費",
      "condition": { "type": "always_true" },
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1000,
        "label": "報名費"
      }
    },
    {
      "id": "early_bird_discount",
      "priority": 100,
      "description": "早鳥 85 折",
      "condition": {
        "type": "rule_ref",
        "rule": "is_early_bird"
      },
      "action": {
        "type": "percentage_discount",
        "value": 15,
        "apply_to": ["registration_fee"],
        "label": "早鳥優惠 85 折"
      }
    }
  ],
  "validation_rules": []
}
```

### 範例 3：團報 + 階梯式運費

```json
{
  "event_id": "team-relay-2026",
  "version": "1.0",
  "name": "2026 接力賽",
  "variables": {
    "per_person_fee": 500,
    "insurance_per_person": 91
  },
  "pricing_rules": [
    {
      "id": "team_registration",
      "priority": 0,
      "description": "團隊報名費",
      "condition": { "type": "always_true" },
      "action": {
        "type": "add_item",
        "item": "registration_fee",
        "unit_price": "$variables.per_person_fee",
        "quantity_field": "team_size",
        "label": "團隊報名費"
      }
    },
    {
      "id": "team_insurance",
      "priority": 20,
      "description": "團隊保險",
      "condition": {
        "type": "equals",
        "field": "addons.insurance",
        "value": true
      },
      "action": {
        "type": "add_item",
        "item": "addon:insurance",
        "unit_price": "$variables.insurance_per_person",
        "quantity_field": "team_size",
        "label": "團體保險"
      }
    },
    {
      "id": "shipping_1_to_2",
      "priority": 30,
      "description": "運費 1-2 人",
      "condition": {
        "type": "and",
        "conditions": [
          { "type": "compare", "field": "team_size", "operator": ">=", "value": 1 },
          { "type": "compare", "field": "team_size", "operator": "<=", "value": 2 }
        ]
      },
      "action": {
        "type": "add_item",
        "item": "addon:shipping",
        "fixed_price": 150,
        "label": "運費 (1-2人)"
      }
    },
    {
      "id": "shipping_3_to_5",
      "priority": 30,
      "description": "運費 3-5 人",
      "condition": {
        "type": "and",
        "conditions": [
          { "type": "compare", "field": "team_size", "operator": ">=", "value": 3 },
          { "type": "compare", "field": "team_size", "operator": "<=", "value": 5 }
        ]
      },
      "action": {
        "type": "add_item",
        "item": "addon:shipping",
        "fixed_price": 200,
        "label": "運費 (3-5人)"
      }
    }
  ],
  "validation_rules": []
}
```

### 範例 4：滿額折扣（使用 computed_fields）

```json
{
  "event_id": "marathon-2026",
  "version": "1.0",
  "name": "2026 馬拉松",
  "computed_fields": {
    "subtotal": {
      "type": "sum_prices",
      "description": "小計",
      "items": ["registration_fee", "addon:*"]
    }
  },
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "報名費",
      "condition": { "type": "always_true" },
      "action": {
        "type": "set_price",
        "value": 1000,
        "label": "報名費"
      }
    },
    {
      "id": "insurance",
      "priority": 20,
      "description": "保險",
      "condition": {
        "type": "equals",
        "field": "addons.insurance",
        "value": true
      },
      "action": {
        "type": "add_item",
        "item": "addon:insurance",
        "unit_price": 500,
        "label": "運動傷害險"
      }
    },
    {
      "id": "volume_discount",
      "priority": 110,
      "description": "滿 1500 折 200",
      "condition": {
        "type": "compare",
        "field": "$computed.subtotal",
        "operator": ">=",
        "value": 1500
      },
      "action": {
        "type": "fixed_discount",
        "value": 200,
        "apply_to": ["total"],
        "label": "滿額折扣 -200"
      }
    }
  ],
  "validation_rules": []
}
```

### 範例 5：完整驗證規則

```json
{
  "event_id": "marathon-2026",
  "version": "1.0",
  "name": "2026 馬拉松",
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "報名費",
      "condition": { "type": "always_true" },
      "action": {
        "type": "set_price",
        "value": 1000,
        "label": "報名費"
      }
    }
  ],
  "validation_rules": [
    {
      "id": "age_limit",
      "description": "年齡限制 18 歲以上",
      "condition": {
        "type": "compare",
        "field": "user.age",
        "operator": "<",
        "value": 18
      },
      "error_type": "blocking",
      "error_message": "參賽者須年滿 18 歲"
    },
    {
      "id": "registration_deadline",
      "description": "報名截止日",
      "condition": {
        "type": "datetime_after",
        "field": "register_date",
        "value": "2026-04-09T23:59:59+08:00"
      },
      "error_type": "blocking",
      "error_message": "報名已截止"
    },
    {
      "id": "health_notice",
      "description": "健康聲明",
      "condition": { "type": "always_true" },
      "error_type": "warning",
      "error_message": "請確認身體狀況適合參賽"
    }
  ]
}
```

---

## 附錄：折扣計算範例

### 連乘模式（multiplicative）

原價 1000，先 9 折再 95 折：
```
Step 1: 1000 × 0.9 = 900
Step 2: 900 × 0.95 = 855
Final: 855
```

### 相加模式（additive）

原價 1000，10% + 5% 折扣：
```
Total discount: 10% + 5% = 15%
Final: 1000 × 0.85 = 850
```

### 只取最優惠（best_only）

原價 1000，有 10% 和 5% 兩個折扣：
```
10% discount = 100
5% discount = 50
Best: 100
Final: 1000 - 100 = 900
```
