# Rules Engine 折扣選擇策略設計

> 日期：2026-01-26
> 狀態：設計討論中

## 背景

Rules Engine 的目標是**用組裝規則取代重寫程式碼**，讓活動主辦方可以透過配置來設定報名規則，而非每次都重寫報名頁面和邏輯。

根據歷史賽事規則分析，折扣邏輯非常複雜且無法完全列舉：
- 折扣類型：百分比折扣、固定金額折扣
- 選擇邏輯：取最佳 N 個、互斥、條件組合
- 套用順序：先百分比再固定、或相反
- 上限控制：總折扣金額上限

## 問題分析

### 當前能力

| 功能 | 狀態 | 說明 |
|------|------|------|
| 折扣全部疊加 | ✅ | `additive` 模式 |
| 折扣相乘 | ✅ | `multiplicative` 模式 |
| 只取最佳折扣 | ✅ | `best_only` 模式 |
| 條件互斥 | ✅ | 用 AND/OR/NOT 組合條件 |

### 缺失能力

| 功能 | 狀態 | 需求場景 |
|------|------|---------|
| 取最佳 N 個折扣 | ❌ | 多折扣但限制疊加數量 |
| 混合折扣比較 | ❌ | 百分比 vs 固定金額比較 |
| 套用順序控制 | ❌ | 先百分比再固定 |
| 總折扣上限 | ❌ | 最多折 500 元 |
| 自訂選擇邏輯 | ❌ | 無法列舉的複雜規則 |

### 核心挑戰

```
問題：折扣選擇邏輯無法完全列舉
矛盾：如何設計「足夠靈活」的 engine？
```

**不可行的方式：**
```go
type DiscountStacking struct {
    Mode string // best_n, best_by_amount, best_by_percent,
                // first_n, last_n, weighted_n...
                // 永遠列舉不完
}
```

---

## 設計方案

### 架構：分層 + 可擴展策略

```
┌─────────────────────────────────────────────┐
│  Layer 3: 折扣選擇策略 (Pluggable)          │
│  - 內建策略: all, best_only, best_n         │
│  - 自訂策略: 表達式 or 腳本 or Hook         │
├─────────────────────────────────────────────┤
│  Layer 2: 折扣規則 (Declarative)            │
│  - 條件 + 動作                              │
│  - 計算折扣金額                             │
├─────────────────────────────────────────────┤
│  Layer 1: 基礎價格 (Declarative)            │
│  - set_price, add_item                      │
└─────────────────────────────────────────────┘
```

### 方案 1：擴展 DiscountStacking 配置

```go
// dsl/types.go
type DiscountStacking struct {
    Mode             string   `json:"mode"`              // additive, multiplicative, best_only, best_n
    MaxDiscounts     int      `json:"max_discounts"`     // best_n 模式: 取幾個
    CompareBy        string   `json:"compare_by"`        // amount | percentage
    ApplyOrder       string   `json:"apply_order"`       // percentage_first | fixed_first | by_priority
    MaxTotalDiscount *float64 `json:"max_total_discount"` // 總折扣上限
    Description      string   `json:"description"`
}
```

**DSL 範例：**
```json
{
  "discount_stacking": {
    "mode": "best_n",
    "max_discounts": 2,
    "compare_by": "amount",
    "apply_order": "percentage_first",
    "max_total_discount": 500,
    "description": "取折扣金額最大的2個，先算百分比再算固定，最多折500元"
  }
}
```

### 方案 2：表達式引擎

分離「折扣定義」和「折扣選擇」：

```json
{
  "pricing_rules": [
    {"id": "early_bird", "action": {"type": "percentage_discount", "value": 20}},
    {"id": "group", "action": {"type": "percentage_discount", "value": 10}},
    {"id": "coupon", "action": {"type": "fixed_discount", "value": 300}}
  ],

  "discount_selection": {
    "strategy": "expression",
    "expression": "top_n_by_amount(qualified_discounts, 2)",
    "apply_order": "percentage_first",
    "max_total": 500
  }
}
```

**內建選擇函數：**
```go
var BuiltinSelectors = map[string]SelectorFunc{
    "all":              selectAll,           // 全部套用
    "best_by_amount":   selectBestByAmount,  // 取金額最大的 1 個
    "top_n_by_amount":  selectTopNByAmount,  // 取金額最大的 N 個
    "top_n_by_percent": selectTopNByPercent, // 取百分比最大的 N 個
    "first_n":          selectFirstN,        // 取前 N 個 (by priority)
}
```

**複雜表達式支援：**
```json
{
  "expression": "
    let percentage_discounts = filter(qualified, d => d.type == 'percentage');
    let fixed_discounts = filter(qualified, d => d.type == 'fixed');
    let best_percent = top_n_by_amount(percentage_discounts, 1);
    let best_fixed = top_n_by_amount(fixed_discounts, 1);
    concat(best_percent, best_fixed)
  "
}
```

### 方案 3：80/20 法則 - DSL + 自訂 Hook

對於無法用 DSL 表達的極端情況，提供程式碼擴展點：

```json
{
  "discount_selection": {
    "strategy": "custom",
    "handler": "nb10k_taipei_2025_discount_selector"
  }
}
```

```go
// 註冊自訂處理器
engine.RegisterDiscountSelector("nb10k_taipei_2025_discount_selector",
    func(qualified []Discount, ctx *Context) []Discount {
        // 這場賽事的特殊邏輯
        // 完全自由的 Go 程式碼
        // 但只寫選擇邏輯，不是整個報名頁
    })
```

**優點：**
- 常見情況 (80%) 用 DSL 配置
- 特殊情況 (20%) 用程式碼（但只寫選擇邏輯）
- 比重寫整個報名頁面省 90% 工作量

---

## 混合折扣比較問題

### 問題

百分比折扣和固定金額折扣如何比較？

```
折扣 A: 8折 (20% off)
折扣 C: 固定折 300 元

哪個比較優惠？→ 取決於基礎價格
```

| 基礎價格 | A (8折) 省 | C (固定300) 省 | 較優惠 |
|----------|-----------|---------------|--------|
| 1000 元 | 200 元 | 300 元 | C |
| 2000 元 | 400 元 | 300 元 | A |
| 1500 元 | 300 元 | 300 元 | 相同 |

### 解決方案

比較時計算「實際折扣金額」：

```go
func (c *Calculator) calculateDiscountAmount(breakdown *dsl.PriceBreakdown,
    rule *dsl.PricingRule) float64 {

    switch rule.Action.Type {
    case "percentage_discount":
        percent := rule.Action.Value.(float64)
        targetTotal := c.getTargetTotal(breakdown, rule.Action.ApplyTo)
        return targetTotal * (percent / 100.0)

    case "fixed_discount":
        return rule.Action.Value.(float64)
    }

    return 0
}
```

---

## 套用順序問題

### 問題

先套用哪個會影響最終結果：

```
基礎價格: 1500 元
折扣 A: 8折
折扣 C: 固定 300 元

方式 1: 先 A 再 C
1500 × 0.8 = 1200 → 1200 - 300 = 900 ✓

方式 2: 先 C 再 A
1500 - 300 = 1200 → 1200 × 0.8 = 960 ✗
```

### 解決方案

統一計算方式，由配置決定順序：

```go
type ApplyOrder string

const (
    ApplyOrderPercentageFirst ApplyOrder = "percentage_first" // 先百分比再固定
    ApplyOrderFixedFirst      ApplyOrder = "fixed_first"      // 先固定再百分比
    ApplyOrderByPriority      ApplyOrder = "by_priority"      // 按規則 priority
)
```

---

## 實作計劃

### Phase 1: 常見模式 (MVP)

**目標：覆蓋 80% 場景**

```
1. 修改 dsl/types.go
   - 擴展 DiscountStacking 結構
   - 新增 MaxDiscounts, CompareBy, ApplyOrder, MaxTotalDiscount

2. 修改 calculator/calculator.go
   - 重構 Phase 2 折扣處理邏輯
   - 實作 selectDiscounts() 策略選擇
   - 實作 calculateDiscountAmount() 金額計算
   - 實作 applyInOrder() 順序套用
   - 實作 enforceMaxTotal() 上限控制

3. 新增測試案例
   - best_n 模式測試
   - 混合折扣比較測試
   - 套用順序測試
   - 折扣上限測試
```

**交付物：**
- 支援 `best_n` 模式
- 支援 `percentage_first` / `fixed_first` 順序
- 支援總折扣上限

### Phase 2: 表達式引擎

**目標：覆蓋更多場景**

```
1. 設計表達式語法
   - 支援 filter, sort, top_n, concat 等函數
   - 支援簡單的 lambda 表達式

2. 實作表達式解析器
   - 可考慮用 govaluate 或類似庫

3. 整合到 discount_selection
```

**交付物：**
- 表達式語法規格
- 表達式引擎實作
- 覆蓋更複雜的選擇邏輯

### Phase 3: 自訂 Hook (按需)

**目標：極端情況的逃生口**

```
1. 定義 DiscountSelector 介面
2. 實作註冊機制
3. 支援在 DSL 中引用自訂 handler
```

**交付物：**
- Hook 註冊 API
- 範例自訂 selector

---

## 對 Chatbot 模板重構的影響

當 Rules Engine 支援更複雜的折扣選擇後，Chatbot 的模板系統需要：

1. **新增 discount_stacking 模板**
   ```go
   type DiscountStackingTemplate struct{}

   func (t *DiscountStackingTemplate) Build(info *ExtractedInfo) (*dsl.DiscountStacking, error) {
       // 根據 LLM 提取的資訊組裝
   }
   ```

2. **LLM 提取折扣疊加設定**
   ```json
   {
     "intent": "set_discount_stacking",
     "values": {
       "mode": "best_n",
       "max_discounts": 2,
       "apply_order": "percentage_first"
     }
   }
   ```

3. **自然語言對應**
   - "取最優惠的2個折扣" → `mode: best_n, max_discounts: 2`
   - "先算百分比折扣" → `apply_order: percentage_first`
   - "最多折500元" → `max_total_discount: 500`

---

## 待確認事項

1. **Phase 1 的優先級？** 是否立即開始實作？

2. **最複雜的歷史賽事範例？** 用來驗證設計是否足夠

3. **表達式語法偏好？**
   - 類 SQL: `SELECT TOP 2 FROM discounts ORDER BY amount DESC`
   - 類函數式: `top_n_by_amount(filter(discounts, d => d.type == 'percentage'), 2)`
   - 類 JSON Path: 其他

4. **是否需要 Phase 3 的 Hook？** 還是 Phase 2 的表達式就足夠？

---

## 參考資料

- [Rules Engine DSL Types](/internal/rules/dsl/types.go)
- [Calculator 實作](/internal/rules/calculator/calculator.go)
- [大湖馬拉松範例](/examples/dahoo-marathon-2026.json)
