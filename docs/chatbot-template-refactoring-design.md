# Chatbot 模板組裝重構設計

> 日期：2026-01-26
> 狀態：設計討論中

## 背景

目前 Chatbot 使用 LLM 直接生成完整的 DSL JSON，但經常出現格式錯誤。每次遇到問題就增加 prompt 範例，導致：

1. **Prompt 膨脹** - 越來越長，成本增加
2. **維護困難** - 範例互相矛盾
3. **治標不治本** - 修一個 case 可能破壞另一個

## 解決方案：模板組裝

### 核心概念

```
目前流程：
用戶輸入 → LLM 生成完整 JSON → 解析 → 驗證 → 使用
           ↑ 容易出錯，格式複雜

重構後：
用戶輸入 → LLM 提取關鍵資訊 → 程式碼組裝 DSL → 驗證 → 使用
           ↑ 簡單結構            ↑ 模板填充
```

### LLM 輸出簡化結構

```go
// LLM 只需要輸出這個簡化結構
type ExtractedInfo struct {
    Intent    string         `json:"intent"`     // 意圖
    EventName string         `json:"event_name"` // 活動名稱
    RuleType  string         `json:"rule_type"`  // pricing | validation
    Subtype   string         `json:"subtype"`    // 見下表
    Values    map[string]any `json:"values"`     // 關鍵數值
    Exclusive string         `json:"exclusive"`  // 互斥條件 (optional)
    Message   string         `json:"message"`    // 回覆訊息
}
```

---

## RuleSubtype 對照表

### Pricing Rules

| Subtype | 說明 | 需要的 Values |
|---------|------|---------------|
| `race_fee` | 組別報名費 | `race_type`, `price` |
| `early_bird` | 早鳥折扣 | `deadline`, `discount_percent` |
| `group_discount` | 團報折扣 | `min_size`, `discount_percent` |
| `addon_fixed` | 固定價加購 | `item_name`, `price` |
| `addon_unit` | 單價加購 | `item_name`, `unit_price` |
| `shipping_tier` | 運費級距 | `min_size`, `max_size`, `price` |
| `fixed_discount` | 固定金額折扣 | `amount`, `condition?` |
| `coupon` | 折扣碼 | `code`, `discount_type`, `value` |

### Validation Rules

| Subtype | 說明 | 需要的 Values |
|---------|------|---------------|
| `deadline` | 報名截止 | `date` |
| `registration_period` | 報名期間 | `start_date`, `end_date` |
| `age_limit` | 年齡限制 | `min_age?`, `max_age?` |
| `required_field` | 必填欄位 | `field_name` |
| `minor_consent` | 未成年同意 | `age_threshold` |

### Exclusive 互斥條件

| Value | 說明 | 產生的條件 |
|-------|------|-----------|
| `individual` | 僅限個人 | `team_size < 3` |
| `group` | 僅限團體 | `team_size >= 3` |

---

## LLM 輸出對比

### 目前方式（複雜）

```json
{
  "intent": "rule_input",
  "rules": [{
    "id": "new_pricing_10k",
    "action": "add",
    "rule_type": "pricing",
    "data": {
      "priority": 0,
      "description": "10K 報名費",
      "condition": {"type":"equals","field":"user.race_type","value":"10K"},
      "action": {"type":"set_price","item":"registration_fee","value":880,"label":"10K 報名費"}
    }
  }]
}
```

### 重構後（簡單）

```json
{
  "intent": "add_rule",
  "rule_type": "pricing",
  "subtype": "race_fee",
  "values": {
    "race_type": "10K",
    "price": 880
  },
  "message": "已新增 10K 組，報名費 NT$880。"
}
```

---

## 模板系統設計

### 模板介面

```go
// internal/chatbot/templates/template.go

type PricingTemplate interface {
    Build(info *ExtractedInfo) (*dsl.PricingRule, error)
}

type ValidationTemplate interface {
    Build(info *ExtractedInfo) (*dsl.ValidationRule, error)
}
```

### 範例：組別報名費模板

```go
// internal/chatbot/templates/race_fee.go

type RaceFeeTemplate struct{}

func (t *RaceFeeTemplate) Build(info *ExtractedInfo) (*dsl.PricingRule, error) {
    raceType := info.Values["race_type"].(string)
    price := info.Values["price"].(float64)

    return &dsl.PricingRule{
        ID:          fmt.Sprintf("pricing_%s", strings.ToLower(raceType)),
        Priority:    0,
        Description: fmt.Sprintf("%s 報名費", raceType),
        Condition: &dsl.Expression{
            Type:  "equals",
            Field: "user.race_type",
            Value: raceType,
        },
        Action: &dsl.Action{
            Type:  "set_price",
            Item:  "registration_fee",
            Value: price,
            Label: fmt.Sprintf("%s 報名費", raceType),
        },
    }, nil
}
```

### 範例：早鳥折扣模板（支援互斥）

```go
// internal/chatbot/templates/early_bird.go

type EarlyBirdTemplate struct{}

func (t *EarlyBirdTemplate) Build(info *ExtractedInfo) (*dsl.PricingRule, error) {
    deadline := info.Values["deadline"].(string)
    discount := info.Values["discount_percent"].(float64)

    // 基礎條件：日期在截止日之前
    baseCondition := &dsl.Expression{
        Type:  "datetime_before",
        Field: "register_date",
        Value: formatDeadline(deadline),
    }

    // 處理互斥條件
    finalCondition := baseCondition
    if info.Exclusive != "" {
        if exclCond, ok := ExclusiveConditions[info.Exclusive]; ok {
            finalCondition = &dsl.Expression{
                Type:       "and",
                Conditions: []*dsl.Expression{baseCondition, exclCond},
            }
        }
    }

    return &dsl.PricingRule{
        ID:          "pricing_early_bird",
        Priority:    10,
        Description: fmt.Sprintf("早鳥優惠 %v%% off", discount),
        Condition:   finalCondition,
        Action: &dsl.Action{
            Type:    "percentage_discount",
            Value:   discount,
            ApplyTo: []string{"registration_fee"},
            Label:   "早鳥優惠",
        },
    }, nil
}

// 互斥條件對照表
var ExclusiveConditions = map[string]*dsl.Expression{
    "individual": {
        Type:     "compare",
        Field:    "team_size",
        Operator: "<",
        Value:    3,
    },
    "group": {
        Type:     "compare",
        Field:    "team_size",
        Operator: ">=",
        Value:    3,
    },
}
```

### 模板註冊表

```go
// internal/chatbot/templates/registry.go

var PricingTemplates = map[string]PricingTemplate{
    "race_fee":       &RaceFeeTemplate{},
    "early_bird":     &EarlyBirdTemplate{},
    "group_discount": &GroupDiscountTemplate{},
    "addon_fixed":    &AddonFixedTemplate{},
    "addon_unit":     &AddonUnitTemplate{},
    "shipping_tier":  &ShippingTierTemplate{},
    "fixed_discount": &FixedDiscountTemplate{},
    "coupon":         &CouponTemplate{},
}

var ValidationTemplates = map[string]ValidationTemplate{
    "deadline":            &DeadlineTemplate{},
    "registration_period": &RegistrationPeriodTemplate{},
    "age_limit":           &AgeLimitTemplate{},
    "required_field":      &RequiredFieldTemplate{},
    "minor_consent":       &MinorConsentTemplate{},
}

func BuildPricingRule(info *ExtractedInfo) (*dsl.PricingRule, error) {
    tmpl, ok := PricingTemplates[info.Subtype]
    if !ok {
        return nil, fmt.Errorf("unknown pricing subtype: %s", info.Subtype)
    }
    return tmpl.Build(info)
}

func BuildValidationRule(info *ExtractedInfo) (*dsl.ValidationRule, error) {
    tmpl, ok := ValidationTemplates[info.Subtype]
    if !ok {
        return nil, fmt.Errorf("unknown validation subtype: %s", info.Subtype)
    }
    return tmpl.Build(info)
}
```

---

## 簡化後的 Prompt

```go
const SimpleExtractionPrompt = `你是規則提取助手。從用戶輸入中提取關鍵資訊。

## 規則類型

### pricing (定價規則)
| subtype | 說明 | 需要的 values |
|---------|------|---------------|
| race_fee | 組別報名費 | race_type, price |
| early_bird | 早鳥折扣 | deadline, discount_percent |
| group_discount | 團報折扣 | min_size, discount_percent |
| addon_fixed | 固定價加購 | item_name, price |
| shipping_tier | 運費級距 | min_size, max_size, price |

### validation (驗證規則)
| subtype | 說明 | 需要的 values |
|---------|------|---------------|
| deadline | 報名截止 | date |
| registration_period | 報名期間 | start_date, end_date |
| age_limit | 年齡限制 | min_age, max_age |

### exclusive (互斥條件，可選)
- individual: 僅限個人報名
- group: 僅限團體報名

## 輸出格式
{"intent":"add_rule","rule_type":"類型","subtype":"子類型","values":{...},"exclusive":"","message":"回覆"}

## 範例

用戶: "10K報名費880元"
輸出: {"intent":"add_rule","rule_type":"pricing","subtype":"race_fee","values":{"race_type":"10K","price":880},"message":"已新增 10K 組，報名費 NT$880。"}

用戶: "報名截止日期是4月30日"
輸出: {"intent":"add_rule","rule_type":"validation","subtype":"deadline","values":{"date":"2026-04-30"},"message":"已設定報名截止日期為 2026/4/30。"}

用戶: "早鳥優惠85折到2月底，僅限個人報名"
輸出: {"intent":"add_rule","rule_type":"pricing","subtype":"early_bird","values":{"deadline":"2026-02-28","discount_percent":15},"exclusive":"individual","message":"已設定早鳥優惠 85 折（僅限個人報名），截止日期 2026/2/28。"}

用戶: "團體報名3人以上9折"
輸出: {"intent":"add_rule","rule_type":"pricing","subtype":"group_discount","values":{"min_size":3,"discount_percent":10},"message":"已設定團體報名 3 人以上享 9 折優惠。"}

## 用戶輸入
{{USER_INPUT}}`
```

---

## 優點比較

| 面向 | 目前方式 | 模板組裝 |
|------|---------|---------|
| Prompt 長度 | ~3000 字 | ~800 字 |
| LLM 任務 | 生成複雜 JSON | 提取關鍵值 |
| 格式錯誤率 | 高 | 低 |
| 維護性 | 差（改 prompt） | 好（改模板） |
| 可測試性 | 難 | 容易（模板單測） |
| 一致性 | LLM 決定 | 程式碼保證 |
| 擴展性 | 加範例 | 加模板 |

---

## 實作計劃

### Phase 1: 基礎架構

```
1. 定義 ExtractedInfo 結構
   - internal/chatbot/agent/extracted_info.go

2. 實作模板系統
   - internal/chatbot/templates/template.go (介面)
   - internal/chatbot/templates/registry.go (註冊表)
   - internal/chatbot/templates/pricing/*.go (各定價模板)
   - internal/chatbot/templates/validation/*.go (各驗證模板)

3. 簡化 prompt
   - internal/chatbot/prompts/extraction_prompt.go

4. 修改 Agent 流程
   - 解析 ExtractedInfo
   - 調用模板 Build
   - 加入 RuleSet
```

### Phase 2: 完善模板

```
1. 補充所有 subtype 模板
2. 處理邊界情況
3. 加入驗證邏輯
```

### Phase 3: 整合測試

```
1. 模板單元測試
2. 整合測試（LLM → 模板 → DSL）
3. 與 Rules Engine 整合測試
```

---

## 檔案結構

```
internal/chatbot/
├── agent/
│   ├── agent.go
│   ├── extracted_info.go      # 新增
│   └── ...
├── prompts/
│   ├── extraction_prompt.go   # 新增（簡化版）
│   └── ...
└── templates/                  # 新增目錄
    ├── template.go            # 介面定義
    ├── registry.go            # 模板註冊
    ├── helpers.go             # 輔助函數
    ├── pricing/
    │   ├── race_fee.go
    │   ├── early_bird.go
    │   ├── group_discount.go
    │   ├── addon.go
    │   └── shipping.go
    └── validation/
        ├── deadline.go
        ├── period.go
        ├── age_limit.go
        └── required_field.go
```

---

## 與 Rules Engine 的關係

模板系統依賴 Rules Engine 的 DSL 結構：

```
Chatbot 模板 ──產生──> dsl.PricingRule / dsl.ValidationRule
                              │
                              ▼
                    Rules Engine 執行
```

當 Rules Engine 擴展新功能（如 `best_n` 折扣選擇）時，Chatbot 需要：

1. 新增對應的 subtype（如 `discount_stacking`）
2. 實作對應的模板
3. 更新 prompt 說明

---

## 待確認事項

1. **是否立即開始實作？**

2. **先處理哪些 subtype？** 建議優先：
   - race_fee
   - early_bird
   - group_discount
   - deadline

3. **模板放在 chatbot 還是獨立 package？**
   - 選項 A: `internal/chatbot/templates/`
   - 選項 B: `pkg/dsl-templates/`

4. **是否需要模板版本控制？** 以防未來 DSL 結構變更

---

## 參考資料

- [Rules Engine DSL Types](/internal/rules/dsl/types.go)
- [大湖馬拉松範例](/examples/dahoo-marathon-2026.json)
- [折扣選擇策略設計](/docs/rules-engine-discount-selection-design.md)
