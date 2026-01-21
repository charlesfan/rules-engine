# Rules Engine DSL 參考文件

## 目錄

1. [概述](#概述)
2. [RuleSet 結構](#ruleset-結構)
3. [變數定義 (Variables)](#變數定義-variables)
4. [定價規則 (Pricing Rules)](#定價規則-pricing-rules)
5. [驗證規則 (Validation Rules)](#驗證規則-validation-rules)
6. [計算欄位 (Computed Fields)](#計算欄位-computed-fields)
7. [表單定義 (Form Schema)](#表單定義-form-schema)
8. [表達式類型 (Expression Types)](#表達式類型-expression-types)
9. [完整範例](#完整範例)

---

## 概述

Rules Engine DSL 是一個基於 JSON 的領域特定語言，用於定義活動報名的規則、定價邏輯和驗證條件。透過 DSL，您可以在不修改程式碼的情況下配置複雜的業務規則。

### 核心概念

- **兩階段計算**：Phase 1 (priority 0-99) 設定基礎價格和加入項目，Phase 2 (priority 100+) 套用折扣
- **優先級排序**：規則按照 `priority` 值從小到大執行
- **條件表達式**：支援複雜的邏輯運算（and, or, not, 比較等）
- **計算欄位**：可在規則中引用動態計算的值
- **動態表單**：根據配置自動產生報名表單

---

## RuleSet 結構

### 頂層欄位

```json
{
  "event_id": "string",           // 活動唯一識別碼
  "version": "string",            // 規則版本號
  "name": "string",               // 活動名稱
  "variables": {},                // 變數定義
  "pricing_rules": [],            // 定價規則陣列
  "validation_rules": [],         // 驗證規則陣列
  "computed_fields": {},          // 計算欄位定義
  "discount_stacking": {},        // 折扣疊加模式
  "form_schema": {}               // 表單定義
}
```

### 欄位說明

| 欄位 | 類型 | 必填 | 說明 |
|------|------|------|------|
| `event_id` | string | 是 | 活動的唯一識別碼，建議使用 kebab-case |
| `version` | string | 是 | 規則版本號，建議使用語義化版本 (e.g., "1.0") |
| `name` | string | 是 | 活動顯示名稱 |
| `variables` | object | 否 | 全域變數定義（可在 action 中使用 `$variables.*` 引用） |
| `pricing_rules` | array | 是 | 定價規則列表 |
| `validation_rules` | array | 否 | 驗證規則列表 |
| `computed_fields` | object | 否 | 計算欄位定義 |
| `discount_stacking` | object | 否 | 折扣疊加設定 |
| `form_schema` | object | 否 | 動態表單定義 |

---

## 變數定義 (Variables)

定義可在規則中重複使用的常數值。

### 結構

```json
{
  "variables": {
    "變數名稱": 數值或字串,
    "full_marathon_price": 1050,
    "insurance_plan_a_price": 91
  }
}
```

### 範例

```json
{
  "variables": {
    "base_registration_fee": 1200,
    "early_bird_discount": 100,
    "group_discount_threshold": 10,
    "max_team_size": 100,
    "addon_water_bottle_price": 150,
    "addon_compression_pants_price": 800
  }
}
```

### 使用變數引用

在 action 的 `value`、`unit_price`、`fixed_price` 欄位中，可以使用 `$variables.變數名稱` 來引用變數：

```json
{
  "pricing_rules": [
    {
      "id": "team_registration_fee",
      "priority": 1,
      "condition": {"type": "always_true"},
      "action": {
        "type": "set_price",
        "value": "$variables.base_registration_fee",
        "label": "團隊報名費"
      }
    },
    {
      "id": "addon_water_bottles",
      "priority": 2,
      "condition": {"type": "always_true"},
      "action": {
        "type": "add_item",
        "item": "addon:water_bottle",
        "unit_price": "$variables.addon_water_bottle_price",
        "quantity_field": "team_size",
        "label": "運動水壺"
      }
    }
  ]
}
```

> **注意**：變數引用僅在 pricing action 的數值欄位中有效，條件表達式中仍需使用具體數值。

---

## 定價規則 (Pricing Rules)

定價規則定義如何計算報名費用，包括基礎價格、加購項目和折扣。

### 規則結構

```json
{
  "id": "string",              // 規則唯一識別碼
  "priority": number,          // 執行優先級 (0-99: Phase 1, 100+: Phase 2)
  "description": "string",     // 規則說明
  "condition": {},             // 觸發條件（Expression）
  "action": {}                 // 執行動作
}
```

### Priority 優先級

| 範圍 | 階段 | 用途 |
|------|------|------|
| 0-9 | Phase 1 | 設定基礎價格 (set_price) |
| 10-99 | Phase 1 | 加入項目 (add_item) |
| 100-199 | Phase 2 | 套用折扣 (percentage_discount, fixed_discount) |
| 200+ | Phase 2 | 價格封頂 (price_cap) |

### Action 類型

#### 1. set_price - 設定基礎價格

設定報名費的基礎價格。通常每個活動只會觸發一個 set_price 規則。

```json
{
  "action": {
    "type": "set_price",
    "item": "registration_fee",    // 項目 ID（可選，預設為 "registration_fee"）
    "value": 1050,                 // 價格（必填）
    "label": "全程馬拉松報名費"      // 顯示名稱（必填）
  }
}
```

**範例：根據賽事組別設定價格**

```json
{
  "id": "full_marathon_registration",
  "priority": 0,
  "description": "全程馬拉松報名費",
  "condition": {
    "type": "equals",
    "field": "user.race_type",
    "value": "full_marathon"
  },
  "action": {
    "type": "set_price",
    "item": "registration_fee",
    "value": 1050,
    "label": "全程馬拉松報名費 (含晶片)"
  }
}
```

#### 2. add_item - 加入項目

加入額外的收費項目，如加購、運費等。

```json
{
  "action": {
    "type": "add_item",
    "item": "addon:保險",          // 項目 ID（必填，建議使用 "addon:" 前綴）
    "unit_price": 91,             // 單價（與 fixed_price 二擇一）
    "fixed_price": 150,           // 固定價格（與 unit_price 二擇一）
    "quantity_field": "team_size", // 數量來源欄位（可選）
    "label": "特定活動傷害險"       // 顯示名稱（必填）
  }
}
```

**unit_price vs fixed_price**

- `unit_price`: 單價，會乘以數量 (quantity)
  - 計算方式：`總價 = unit_price × quantity`
  - 適用於：按人數計費的項目（保險、餐費等）

- `fixed_price`: 固定價格，不論數量多少都是固定金額
  - 計算方式：`總價 = fixed_price`
  - 適用於：固定費用項目（運費、手續費等）

**範例 1：按人數計費的保險**

```json
{
  "id": "insurance_plan_a",
  "priority": 20,
  "description": "加購傷害險方案A",
  "condition": {
    "type": "equals",
    "field": "addons.insurance_plan",
    "value": "plan_a"
  },
  "action": {
    "type": "add_item",
    "item": "addon:insurance",
    "unit_price": 91,
    "quantity_field": "team_size",
    "label": "特定活動傷害險 方案A"
  }
}
```

**範例 2：固定運費**

```json
{
  "id": "shipping_1_to_2",
  "priority": 10,
  "description": "宅配運費 (1-2人)",
  "condition": {
    "type": "and",
    "conditions": [
      {
        "type": "compare",
        "field": "team_size",
        "operator": ">=",
        "value": 1
      },
      {
        "type": "compare",
        "field": "team_size",
        "operator": "<=",
        "value": 2
      }
    ]
  },
  "action": {
    "type": "add_item",
    "item": "addon:shipping",
    "fixed_price": 150,
    "label": "宅配運費 (1-2人)"
  }
}
```

#### 3. percentage_discount - 百分比折扣

套用百分比折扣到指定項目或總額。

```json
{
  "action": {
    "type": "percentage_discount",
    "value": 10,                      // 折扣百分比（10 = 打9折）
    "apply_to": ["registration_fee"], // 套用目標（可選）
    "label": "早鳥優惠 9折"             // 顯示名稱（必填）
  }
}
```

**apply_to 目標選項**

- `["registration_fee"]` - 只對報名費打折
- `["addon:insurance"]` - 只對保險打折
- `["addon:*"]` - 對所有加購項目打折（萬用字元）
- `["total"]` 或 `["subtotal"]` - 對所有項目打折
- 省略或空陣列 - 預設對 registration_fee 打折

**範例：長距離優惠**

```json
{
  "id": "long_distance_discount",
  "priority": 100,
  "description": "長距離賽事 9折優惠",
  "condition": {
    "type": "or",
    "conditions": [
      {
        "type": "equals",
        "field": "user.race_type",
        "value": "full_marathon"
      },
      {
        "type": "equals",
        "field": "user.race_type",
        "value": "half_marathon"
      }
    ]
  },
  "action": {
    "type": "percentage_discount",
    "value": 10,
    "apply_to": ["registration_fee"],
    "label": "長距離賽事優惠 9折"
  }
}
```

#### 4. fixed_discount - 固定金額折扣

扣減固定金額。

```json
{
  "action": {
    "type": "fixed_discount",
    "value": 200,                   // 折扣金額
    "apply_to": ["total"],          // 套用目標（可選）
    "label": "滿額折扣 -200"         // 顯示名稱（必填）
  }
}
```

**範例：滿額折扣**

```json
{
  "id": "volume_discount_2000",
  "priority": 110,
  "description": "滿 2000 折 200",
  "condition": {
    "type": "compare",
    "field": "$computed.subtotal",
    "operator": ">=",
    "value": 2000
  },
  "action": {
    "type": "fixed_discount",
    "value": 200,
    "apply_to": ["total"],
    "label": "滿 2000 折 200"
  }
}
```

#### 5. price_cap - 價格封頂

限制價格上限。

```json
{
  "action": {
    "type": "price_cap",
    "value": 5000,                  // 最高價格
    "apply_to": ["total"],          // 套用目標（可選）
    "label": "價格封頂"              // 顯示名稱（必填）
  }
}
```

**範例：總價封頂**

```json
{
  "id": "max_price_cap",
  "priority": 200,
  "description": "總價最高 5000 元",
  "condition": {
    "type": "always_true"
  },
  "action": {
    "type": "price_cap",
    "value": 5000,
    "apply_to": ["total"],
    "label": "價格封頂 5000"
  }
}
```

---

## 驗證規則 (Validation Rules)

驗證規則定義報名限制和條件檢查。

### 規則結構

```json
{
  "id": "string",              // 規則唯一識別碼
  "description": "string",     // 規則說明
  "condition": {},             // 觸發條件（當條件為 true 時顯示錯誤）
  "error_type": "string",      // 錯誤類型："blocking" 或 "warning"
  "error_message": "string"    // 錯誤訊息
}
```

### Error Type 錯誤類型

| 類型 | 說明 | 效果 |
|------|------|------|
| `blocking` | 阻擋性錯誤 | 無法完成報名，`valid` 為 false |
| `warning` | 警告訊息 | 可以報名，但顯示提醒 |

### 範例 1：報名時間限制

```json
{
  "id": "registration_period",
  "description": "報名時間限制",
  "condition": {
    "type": "not",
    "condition": {
      "type": "datetime_between",
      "field": "register_date",
      "start": "2025-09-22T13:00:00+08:00",
      "end": "2025-12-15T23:59:59+08:00"
    }
  },
  "error_type": "blocking",
  "error_message": "報名時間為 2025/9/22 13:00 至 2025/12/15 23:59，目前不在報名時間內"
}
```

### 範例 2：年齡限制

```json
{
  "id": "minor_needs_consent",
  "description": "未滿18歲需檢附家長同意書",
  "condition": {
    "type": "and",
    "conditions": [
      {
        "type": "compare",
        "field": "user.age",
        "operator": "<",
        "value": 18
      },
      {
        "type": "not",
        "condition": {
          "type": "equals",
          "field": "user.has_parent_consent",
          "value": true
        }
      }
    ]
  },
  "error_type": "blocking",
  "error_message": "未滿18歲參賽者需檢附家長同意切結書"
}
```

### 範例 3：健康警告

```json
{
  "id": "health_warning",
  "description": "健康狀況警告",
  "condition": {
    "type": "always_true"
  },
  "error_type": "warning",
  "error_message": "請注意：如有心臟、血管、糖尿病等病歷者，請勿參加。"
}
```

---

## 計算欄位 (Computed Fields)

計算欄位允許定義動態計算的值，可在條件表達式中引用。

### 結構

```json
{
  "computed_fields": {
    "欄位名稱": {
      "type": "計算類型",
      "description": "說明",
      // 其他參數視類型而定
    }
  }
}
```

### 計算類型

#### 1. sum_prices - 價格總和

計算指定項目的價格總和。

```json
{
  "subtotal": {
    "type": "sum_prices",
    "description": "所有項目小計",
    "items": ["registration_fee", "addon:*"]
  }
}
```

**items 參數**
- 可以是具體的項目 ID：`["registration_fee", "addon:insurance"]`
- 支援萬用字元：`["addon:*"]` 匹配所有 addon 開頭的項目

#### 2. count_items - 項目數量

計算匹配的項目數量。

```json
{
  "addon_count": {
    "type": "count_items",
    "description": "加購項目數量",
    "items": ["addon:*"]
  }
}
```

#### 3. item_price - 單一項目價格

取得特定項目的價格。

```json
{
  "registration_price": {
    "type": "item_price",
    "description": "報名費金額",
    "item": "registration_fee"
  }
}
```

#### 4. count_array_field - 陣列欄位計數

計算陣列中符合條件的項目數量。

```json
{
  "female_count": {
    "type": "count_array_field",
    "description": "計算女性隊員人數",
    "array": "team.members",
    "field": "gender",
    "value": "female"
  }
}
```

**參數說明**
- `array`: 陣列路徑 (如 `team.members`)
- `field`: 要檢查的欄位路徑，支援巢狀路徑 (如 `addons.water_bottle.quantity`)
- `value`: 要匹配的值（可選，若不指定則只要欄位存在即計數）

#### 5. sum_array_field - 陣列欄位加總 ✨ 新增

加總陣列中指定欄位的數值。

```json
{
  "total_water_bottles": {
    "type": "sum_array_field",
    "description": "計算所有隊員購買的水壺總數量",
    "array": "team.members",
    "field": "addons.water_bottle.quantity"
  }
}
```

**參數說明**
- `array`: 陣列路徑 (如 `team.members`)
- `field`: 要加總的欄位路徑，支援巢狀路徑 (如 `addons.water_bottle.quantity`)
- 自動轉換數值型別 (int, int64, float32, float64)
- 若陣列不存在或項目為 nil，返回 0

**完整範例：團隊加購項目統計**

```json
{
  "computed_fields": {
    "total_water_bottles": {
      "type": "sum_array_field",
      "description": "所有隊員購買的水壺總數量",
      "array": "team.members",
      "field": "addons.water_bottle.quantity"
    },
    "total_compression_pants": {
      "type": "sum_array_field",
      "description": "所有隊員購買的壓縮褲總數量",
      "array": "team.members",
      "field": "addons.compression_pants.quantity"
    }
  },
  "pricing_rules": [
    {
      "id": "addon_water_bottles",
      "priority": 2,
      "description": "加購水壺",
      "condition": {
        "type": "compare",
        "field": "$computed.total_water_bottles",
        "operator": ">",
        "value": 0
      },
      "action": {
        "type": "add_item",
        "item": "addon:water_bottle",
        "label": "運動水壺",
        "unit_price": 150,
        "quantity_field": "$computed.total_water_bottles"
      }
    }
  ]
}
```

### 使用計算欄位

在條件表達式中使用 `$computed.欄位名稱` 引用計算欄位：

```json
{
  "computed_fields": {
    "subtotal": {
      "type": "sum_prices",
      "description": "小計",
      "items": ["registration_fee", "addon:*"]
    }
  },
  "pricing_rules": [
    {
      "id": "volume_discount",
      "priority": 110,
      "condition": {
        "type": "compare",
        "field": "$computed.subtotal",
        "operator": ">=",
        "value": 2000
      },
      "action": {
        "type": "fixed_discount",
        "value": 200,
        "label": "滿 2000 折 200"
      }
    }
  ]
}
```

---

## 表單定義 (Form Schema)

動態表單定義前端報名表單的欄位和配置。

### 結構

```json
{
  "form_schema": {
    "fields": [
      {
        "id": "欄位 ID",
        "label": "顯示標籤",
        "type": "欄位類型",
        "field": "資料欄位路徑",
        "required": true/false,
        // 其他參數視類型而定
      }
    ]
  }
}
```

### 欄位類型

#### 1. text - 文字輸入

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

#### 2. email - Email 輸入

```json
{
  "id": "email",
  "label": "Email",
  "type": "email",
  "field": "user.email",
  "required": true,
  "placeholder": "請輸入 Email"
}
```

#### 3. number - 數字輸入

```json
{
  "id": "age",
  "label": "年齡",
  "type": "number",
  "field": "user.age",
  "required": true,
  "min": 1,
  "max": 120,
  "placeholder": "請輸入年齡"
}
```

**特殊欄位：team_size**

```json
{
  "id": "team_size",
  "label": "團體報名人數",
  "type": "number",
  "field": "team_size",
  "default_value": 1,
  "min": 1,
  "max": 100
}
```

> **注意**：`team_size` 是特殊的頂層欄位，不屬於 user 或 addons。

#### 4. select - 下拉選單

```json
{
  "id": "race_type",
  "label": "賽事組別",
  "type": "select",
  "field": "user.race_type",
  "required": true,
  "options": [
    { "label": "請選擇組別", "value": "" },
    { "label": "全程馬拉松 (42公里)", "value": "full_marathon" },
    { "label": "半程馬拉松 (21公里)", "value": "half_marathon" }
  ]
}
```

#### 5. checkbox - 勾選框

```json
{
  "id": "has_parent_consent",
  "label": "我已取得家長同意",
  "type": "checkbox",
  "field": "user.has_parent_consent",
  "default_value": false
}
```

### Field 欄位路徑

欄位路徑決定資料在 Context 中的位置：

| 路徑格式 | 說明 | 範例 |
|----------|------|------|
| `user.欄位名` | 使用者資料 | `user.name`, `user.age` |
| `addons.欄位名` | 加購選項 | `addons.insurance_plan` |
| `team_size` | 團體人數（特殊頂層欄位） | `team_size` |
| `team.members.索引.欄位名` | 團隊成員資料 ✨ | `team.members.0.name`, `team.members.0.gender` |
| `team.members.索引.addons.項目.屬性` | 團隊成員加購項目 ✨ | `team.members.0.addons.water_bottle.quantity` |

### 巢狀路徑支援 ✨ 新增

前端的 `setNestedValue` 方法現在支援任意深度的巢狀路徑，可以正確處理團隊成員的個人加購項目：

**範例：團隊成員加購項目表單**

```json
{
  "form_schema": {
    "fields": [
      {
        "id": "member_0_name",
        "label": "隊員 1 - 姓名",
        "type": "text",
        "field": "team.members.0.name",
        "required": true
      },
      {
        "id": "member_0_gender",
        "label": "隊員 1 - 性別",
        "type": "select",
        "field": "team.members.0.gender",
        "required": true,
        "options": [
          {"label": "男性", "value": "male"},
          {"label": "女性", "value": "female"}
        ]
      },
      {
        "id": "member_0_water_bottle_qty",
        "label": "隊員 1 - 運動水壺數量 (NT$150/個)",
        "type": "number",
        "field": "team.members.0.addons.water_bottle.quantity",
        "default_value": 0,
        "min": 0,
        "max": 10
      },
      {
        "id": "member_0_compression_pants_size",
        "label": "隊員 1 - 壓縮褲尺寸 (NT$800)",
        "type": "select",
        "field": "team.members.0.addons.compression_pants.size",
        "options": [
          {"label": "不需要", "value": ""},
          {"label": "S", "value": "S"},
          {"label": "M", "value": "M"},
          {"label": "L", "value": "L"},
          {"label": "XL", "value": "XL"}
        ]
      }
    ]
  }
}
```

**產生的 Context 結構**

```json
{
  "team": {
    "members": [
      {
        "name": "張小明",
        "gender": "female",
        "addons": {
          "water_bottle": {
            "quantity": 2
          },
          "compression_pants": {
            "size": "M"
          }
        }
      }
    ]
  }
}
```

---

## 表達式類型 (Expression Types)

表達式用於條件判斷，決定規則是否觸發。

### 1. always_true - 恆真

總是返回 true，用於無條件觸發的規則。

```json
{
  "type": "always_true"
}
```

### 2. equals - 相等比較

檢查欄位值是否等於指定值。

```json
{
  "type": "equals",
  "field": "user.race_type",
  "value": "full_marathon"
}
```

### 3. compare - 數值比較

比較欄位值與指定值的大小關係。

```json
{
  "type": "compare",
  "field": "user.age",
  "operator": ">=",
  "value": 18
}
```

**支援的運算子**
- `>` - 大於
- `<` - 小於
- `>=` - 大於等於
- `<=` - 小於等於
- `==` - 等於
- `!=` - 不等於

### 4. and - 邏輯與

所有子條件都必須為 true。

```json
{
  "type": "and",
  "conditions": [
    {
      "type": "compare",
      "field": "team_size",
      "operator": ">=",
      "value": 3
    },
    {
      "type": "compare",
      "field": "team_size",
      "operator": "<=",
      "value": 5
    }
  ]
}
```

### 5. or - 邏輯或

任一子條件為 true 即可。

```json
{
  "type": "or",
  "conditions": [
    {
      "type": "equals",
      "field": "user.race_type",
      "value": "full_marathon"
    },
    {
      "type": "equals",
      "field": "user.race_type",
      "value": "half_marathon"
    }
  ]
}
```

### 6. not - 邏輯非

反轉子條件的結果。

```json
{
  "type": "not",
  "condition": {
    "type": "equals",
    "field": "user.has_parent_consent",
    "value": true
  }
}
```

### 7. datetime_before - 時間早於

檢查時間是否早於指定時間。

```json
{
  "type": "datetime_before",
  "field": "register_date",
  "value": "2025-10-01T00:00:00+08:00"
}
```

### 8. datetime_after - 時間晚於

檢查時間是否晚於指定時間。

```json
{
  "type": "datetime_after",
  "field": "register_date",
  "value": "2025-09-22T13:00:00+08:00"
}
```

### 9. datetime_between - 時間區間

檢查時間是否在指定區間內。

```json
{
  "type": "datetime_between",
  "field": "register_date",
  "start": "2025-09-22T13:00:00+08:00",
  "end": "2025-12-15T23:59:59+08:00"
}
```

### 10. field_exists - 欄位存在

檢查欄位是否存在。

```json
{
  "type": "field_exists",
  "field": "user.certification_number"
}
```

### 11. field_empty - 欄位為空

檢查欄位是否為空（null、空字串、空陣列等）。

```json
{
  "type": "field_empty",
  "field": "user.race_type"
}
```

### 12. in_list - 在列表中

檢查欄位值是否在資料源列表中。

```json
{
  "type": "in_list",
  "field": "user.license_plate",
  "list": "$data_sources.car_owners",
  "match_field": "plate_number"
}
```

### 13. array_any - 陣列任一匹配

檢查陣列中是否有任一元素滿足條件。

```json
{
  "type": "array_any",
  "array": "team.members",
  "condition": {
    "type": "equals",
    "field": "user.gender",
    "value": "female"
  }
}
```

> **重要**：在 `array_any` 和 `array_all` 的 `condition` 中，陣列項目會被放入臨時 context 的 `user` 欄位中，因此必須使用 `user.欄位名` 而非直接使用 `欄位名`。

**錯誤範例**：
```json
{
  "type": "array_any",
  "array": "team.members",
  "condition": {
    "type": "equals",
    "field": "gender",  // ❌ 錯誤：找不到欄位
    "value": "female"
  }
}
```

**正確範例**：
```json
{
  "type": "array_any",
  "array": "team.members",
  "condition": {
    "type": "equals",
    "field": "user.gender",  // ✅ 正確：使用 user. 前綴
    "value": "female"
  }
}
```

### 14. array_all - 陣列全部匹配

檢查陣列中所有元素是否都滿足條件。

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

> **注意**：與 `array_any` 相同，必須使用 `user.欄位名` 格式。

---

## 完整範例

以下是一個完整的馬拉松活動 DSL 範例：

```json
{
  "event_id": "dahoo-strawberry-marathon-2026",
  "version": "1.0",
  "name": "2026大湖草莓文化嘉年華馬拉松",

  "variables": {
    "full_marathon_price": 1050,
    "half_marathon_price": 950,
    "insurance_plan_a_price": 91
  },

  "computed_fields": {
    "subtotal": {
      "type": "sum_prices",
      "description": "所有項目小計",
      "items": ["registration_fee", "addon:*"]
    }
  },

  "pricing_rules": [
    {
      "id": "full_marathon_registration",
      "priority": 0,
      "description": "全程馬拉松報名費",
      "condition": {
        "type": "equals",
        "field": "user.race_type",
        "value": "full_marathon"
      },
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1050,
        "label": "全程馬拉松報名費 (含晶片)"
      }
    },
    {
      "id": "shipping_1_to_2",
      "priority": 10,
      "description": "宅配運費 (1-2人)",
      "condition": {
        "type": "and",
        "conditions": [
          {
            "type": "compare",
            "field": "team_size",
            "operator": ">=",
            "value": 1
          },
          {
            "type": "compare",
            "field": "team_size",
            "operator": "<=",
            "value": 2
          }
        ]
      },
      "action": {
        "type": "add_item",
        "item": "addon:shipping",
        "fixed_price": 150,
        "label": "宅配運費 (1-2人)"
      }
    },
    {
      "id": "insurance_plan_a",
      "priority": 20,
      "description": "加購傷害險方案A",
      "condition": {
        "type": "equals",
        "field": "addons.insurance_plan",
        "value": "plan_a"
      },
      "action": {
        "type": "add_item",
        "item": "addon:insurance",
        "unit_price": 91,
        "quantity_field": "team_size",
        "label": "特定活動傷害險 方案A"
      }
    },
    {
      "id": "long_distance_discount",
      "priority": 100,
      "description": "長距離賽事 9折優惠",
      "condition": {
        "type": "or",
        "conditions": [
          {
            "type": "equals",
            "field": "user.race_type",
            "value": "full_marathon"
          },
          {
            "type": "equals",
            "field": "user.race_type",
            "value": "half_marathon"
          }
        ]
      },
      "action": {
        "type": "percentage_discount",
        "value": 10,
        "apply_to": ["registration_fee"],
        "label": "長距離賽事優惠 9折"
      }
    },
    {
      "id": "volume_discount_2000",
      "priority": 110,
      "description": "滿 2000 折 200",
      "condition": {
        "type": "compare",
        "field": "$computed.subtotal",
        "operator": ">=",
        "value": 2000
      },
      "action": {
        "type": "fixed_discount",
        "value": 200,
        "apply_to": ["total"],
        "label": "滿 2000 折 200"
      }
    }
  ],

  "validation_rules": [
    {
      "id": "registration_period",
      "description": "報名時間限制",
      "condition": {
        "type": "not",
        "condition": {
          "type": "datetime_between",
          "field": "register_date",
          "start": "2025-09-22T13:00:00+08:00",
          "end": "2025-12-15T23:59:59+08:00"
        }
      },
      "error_type": "blocking",
      "error_message": "報名時間為 2025/9/22 13:00 至 2025/12/15 23:59，目前不在報名時間內"
    },
    {
      "id": "minor_needs_consent",
      "description": "未滿18歲需檢附家長同意書",
      "condition": {
        "type": "and",
        "conditions": [
          {
            "type": "compare",
            "field": "user.age",
            "operator": "<",
            "value": 18
          },
          {
            "type": "not",
            "condition": {
              "type": "equals",
              "field": "user.has_parent_consent",
              "value": true
            }
          }
        ]
      },
      "error_type": "blocking",
      "error_message": "未滿18歲參賽者需檢附家長同意切結書"
    }
  ],

  "form_schema": {
    "fields": [
      {
        "id": "name",
        "label": "姓名",
        "type": "text",
        "field": "user.name",
        "required": true,
        "placeholder": "請輸入姓名"
      },
      {
        "id": "age",
        "label": "年齡",
        "type": "number",
        "field": "user.age",
        "required": true,
        "min": 1,
        "max": 120
      },
      {
        "id": "race_type",
        "label": "賽事組別",
        "type": "select",
        "field": "user.race_type",
        "required": true,
        "options": [
          { "label": "請選擇組別", "value": "" },
          { "label": "全程馬拉松 (42公里) - NT$1,050", "value": "full_marathon" },
          { "label": "半程馬拉松 (21公里) - NT$950", "value": "half_marathon" }
        ]
      },
      {
        "id": "team_size",
        "label": "團體報名人數",
        "type": "number",
        "field": "team_size",
        "default_value": 1,
        "min": 1,
        "max": 100
      },
      {
        "id": "insurance_plan",
        "label": "特定活動傷害險",
        "type": "select",
        "field": "addons.insurance_plan",
        "options": [
          { "label": "不加保", "value": "" },
          { "label": "方案A - NT$91", "value": "plan_a" }
        ]
      },
      {
        "id": "has_parent_consent",
        "label": "我已取得家長同意（未滿18歲必填）",
        "type": "checkbox",
        "field": "user.has_parent_consent",
        "default_value": false
      }
    ]
  }
}
```

---

## 常見問題

### Q1: 為什麼改變團報人數，價格沒有變化？

**A:** 檢查以下幾點：

1. **確認有使用 `unit_price` 而非 `fixed_price`**
   - `unit_price` 會乘以數量
   - `fixed_price` 是固定金額，不受數量影響

2. **確認有指定 `quantity_field`**
   ```json
   {
     "action": {
       "type": "add_item",
       "item": "addon:insurance",
       "unit_price": 91,
       "quantity_field": "team_size",  // 必須指定
       "label": "保險"
     }
   }
   ```

3. **檢查運費規則是否使用 `fixed_price`**
   - 運費通常是固定價格，不應該乘以人數
   - 應該使用條件判斷不同人數區間，設定不同的 `fixed_price`

### Q2: 如何實現滿額折扣？

**A:** 使用 `computed_fields` 計算小計，然後在折扣規則中引用：

```json
{
  "computed_fields": {
    "subtotal": {
      "type": "sum_prices",
      "items": ["registration_fee", "addon:*"]
    }
  },
  "pricing_rules": [
    {
      "id": "volume_discount",
      "priority": 110,
      "condition": {
        "type": "compare",
        "field": "$computed.subtotal",
        "operator": ">=",
        "value": 2000
      },
      "action": {
        "type": "fixed_discount",
        "value": 200,
        "label": "滿 2000 折 200"
      }
    }
  ]
}
```

### Q3: 如何讓折扣只對特定項目生效？

**A:** 使用 `apply_to` 參數指定目標項目：

```json
{
  "action": {
    "type": "percentage_discount",
    "value": 10,
    "apply_to": ["registration_fee"],  // 只對報名費打折
    "label": "報名費 9折"
  }
}
```

### Q4: Priority 應該如何設定？

**A:** 遵循以下原則：

- 0-9: 基礎價格 (set_price)
- 10-99: 加購項目 (add_item)
- 100-199: 折扣 (percentage_discount, fixed_discount)
- 200+: 價格封頂 (price_cap)

同類規則之間，優先執行的設定較小的 priority。

---

## 版本歷史

- **v1.1** (2025-10-16)
  - 新增 `sum_array_field` 計算欄位類型
  - 新增變數引用支援 (`$variables.*`)
  - 新增巢狀路徑支援（團隊成員加購項目）
  - 修正 `array_any`/`array_all` 欄位路徑問題
  - 計算欄位評估移至 Phase 0

- **v1.0** (2025-10-15)
  - 初始版本
  - 支援基本定價、驗證、計算欄位和動態表單
