package calculator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charlesfan/rules-engine/internal/rules/computed"
	"github.com/charlesfan/rules-engine/internal/rules/dsl"
)

// Calculator 價格計算器
type Calculator struct {
	ruleSet *dsl.RuleSet
}

// NewCalculator 建立新的 Calculator
func NewCalculator(ruleSet *dsl.RuleSet) *Calculator {
	return &Calculator{
		ruleSet: ruleSet,
	}
}

// Calculate 計算價格
func (c *Calculator) Calculate(ctx *dsl.Context, evaluator ExpressionEvaluator) (*dsl.PriceBreakdown, []string, error) {
	breakdown := &dsl.PriceBreakdown{
		Items:         make(map[string]*dsl.PriceItem),
		Discounts:     []dsl.DiscountItem{},
		SubTotal:      0,
		TotalDiscount: 0,
		FinalPrice:    0,
	}

	appliedRules := []string{}

	// 按照 priority 排序規則
	sortedRules := make([]*dsl.PricingRule, len(c.ruleSet.PricingRules))
	copy(sortedRules, c.ruleSet.PricingRules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Priority < sortedRules[j].Priority
	})

	// === Phase 0: 評估計算欄位（需要在使用之前先計算）===
	if c.ruleSet.ComputedFields != nil {
		computedEval := computed.NewFieldEvaluator(c.ruleSet)
		if err := computedEval.Evaluate(ctx, breakdown); err != nil {
			return nil, nil, fmt.Errorf("failed to evaluate computed fields: %w", err)
		}
	}

	// Phase 1: 設定基礎價格和加入項目
	for _, rule := range sortedRules {
		matches, err := evaluator.EvaluateExpression(rule.Condition, ctx)
		if err != nil || !matches {
			continue
		}

		appliedRules = append(appliedRules, rule.ID)

		switch rule.Action.Type {
		case "set_price":
			// 設定價格
			if err := c.setPrice(breakdown, rule.Action); err != nil {
				return nil, nil, err
			}

		case "add_item":
			// 加入項目（加購）
			if err := c.addItem(breakdown, rule.Action, ctx, evaluator); err != nil {
				return nil, nil, err
			}
		}
	}

	// 計算小計
	breakdown.SubTotal = c.calculateSubTotal(breakdown)

	// === Phase 1.5: 重新評估計算欄位（針對依賴 breakdown.Items 的欄位）===
	// 在 Phase 1 之後，breakdown.Items 已經建立完成
	// 現在可以正確計算 sum_prices, count_items, item_price 等欄位
	if c.ruleSet.ComputedFields != nil {
		computedEval := computed.NewFieldEvaluator(c.ruleSet)
		if err := computedEval.Evaluate(ctx, breakdown); err != nil {
			return nil, nil, fmt.Errorf("failed to re-evaluate computed fields after Phase 1: %w", err)
		}
	}

	// Phase 2: 套用折扣
	for _, rule := range sortedRules {
		matches, err := evaluator.EvaluateExpression(rule.Condition, ctx)
		if err != nil || !matches {
			continue
		}

		switch rule.Action.Type {
		case "percentage_discount":
			if err := c.applyPercentageDiscount(breakdown, rule); err != nil {
				return nil, nil, err
			}

		case "fixed_discount":
			if err := c.applyFixedDiscount(breakdown, rule); err != nil {
				return nil, nil, err
			}

		case "price_cap":
			if err := c.applyPriceCap(breakdown, rule.Action); err != nil {
				return nil, nil, err
			}
		}
	}

	// 計算總折扣
	breakdown.TotalDiscount = c.calculateTotalDiscount(breakdown)

	// 計算最終價格
	breakdown.FinalPrice = c.calculateFinalPrice(breakdown)

	// 確保價格不為負數
	if breakdown.FinalPrice < 0 {
		breakdown.FinalPrice = 0
	}

	return breakdown, appliedRules, nil
}

// setPrice 設定價格
func (c *Calculator) setPrice(breakdown *dsl.PriceBreakdown, action *dsl.Action) error {
	// 解析 action.Value（可能是數字或變數引用）
	value, err := c.resolveValue(action.Value)
	if err != nil {
		return err
	}

	itemID := action.Item
	if itemID == "" {
		itemID = "registration_fee"
	}

	breakdown.Items[itemID] = &dsl.PriceItem{
		ID:              itemID,
		Name:            action.Label,
		Quantity:        1,
		UnitPrice:       value,
		OriginalPrice:   value,
		DiscountedPrice: value,
		FinalPrice:      value,
	}

	return nil
}

// addItem 加入項目
func (c *Calculator) addItem(breakdown *dsl.PriceBreakdown, action *dsl.Action, ctx *dsl.Context, evaluator ExpressionEvaluator) error {
	quantity := 1

	// 取得數量
	if action.QuantityField != "" {
		quantityValue, err := evaluator.GetFieldValue(action.QuantityField, ctx)
		if err == nil {
			if q, err := c.toFloat64(quantityValue); err == nil {
				quantity = int(q)
			}
		}
	}

	if quantity <= 0 {
		return nil
	}

	itemID := action.Item
	if itemID == "" {
		return fmt.Errorf("add_item requires item field")
	}

	var originalPrice float64
	var unitPrice float64

	// 判斷使用固定價格還是單價 * 數量
	if action.FixedPrice != nil {
		// 使用固定價格（不乘以數量）
		fixedPrice, err := c.resolveValue(action.FixedPrice)
		if err != nil {
			return fmt.Errorf("failed to resolve fixed_price: %w", err)
		}
		originalPrice = fixedPrice
		unitPrice = fixedPrice // 顯示用
		quantity = 1           // 固定價格時數量設為 1
	} else if action.UnitPrice != nil {
		// 使用單價 * 數量
		resolvedUnitPrice, err := c.resolveValue(action.UnitPrice)
		if err != nil {
			return fmt.Errorf("failed to resolve unit_price: %w", err)
		}
		unitPrice = resolvedUnitPrice
		originalPrice = resolvedUnitPrice * float64(quantity)
	} else {
		return fmt.Errorf("add_item requires either unit_price or fixed_price")
	}

	breakdown.Items[itemID] = &dsl.PriceItem{
		ID:              itemID,
		Name:            action.Label,
		Quantity:        quantity,
		UnitPrice:       unitPrice,
		OriginalPrice:   originalPrice,
		DiscountedPrice: originalPrice,
		FinalPrice:      originalPrice,
	}

	return nil
}

// applyPercentageDiscount 套用百分比折扣
func (c *Calculator) applyPercentageDiscount(breakdown *dsl.PriceBreakdown, rule *dsl.PricingRule) error {
	discountPercent, err := c.resolveValue(rule.Action.Value)
	if err != nil {
		return err
	}

	targets := rule.Action.ApplyTo
	if len(targets) == 0 {
		targets = []string{"registration_fee"}
	}

	totalDiscountAmount := 0.0

	for _, target := range targets {
		// 處理萬用字元 addon:*
		if target == "addon:*" {
			for itemID, item := range breakdown.Items {
				if strings.HasPrefix(itemID, "addon:") {
					discountAmount := item.DiscountedPrice * (discountPercent / 100.0)
					item.DiscountedPrice -= discountAmount
					totalDiscountAmount += discountAmount
				}
			}
		} else if target == "total" || target == "subtotal" {
			// 對所有項目打折
			for _, item := range breakdown.Items {
				discountAmount := item.DiscountedPrice * (discountPercent / 100.0)
				item.DiscountedPrice -= discountAmount
				totalDiscountAmount += discountAmount
			}
		} else {
			// 特定項目
			if item, exists := breakdown.Items[target]; exists {
				discountAmount := item.DiscountedPrice * (discountPercent / 100.0)
				item.DiscountedPrice -= discountAmount
				totalDiscountAmount += discountAmount
			}
		}
	}

	if totalDiscountAmount > 0 {
		breakdown.Discounts = append(breakdown.Discounts, dsl.DiscountItem{
			RuleID:    rule.ID,
			Label:     rule.Action.Label,
			Type:      "percentage",
			Value:     discountPercent,
			Amount:    totalDiscountAmount,
			AppliedTo: strings.Join(targets, ", "),
		})
	}

	return nil
}

// applyFixedDiscount 套用固定金額折扣
func (c *Calculator) applyFixedDiscount(breakdown *dsl.PriceBreakdown, rule *dsl.PricingRule) error {
	discountAmount, err := c.resolveValue(rule.Action.Value)
	if err != nil {
		return err
	}

	targets := rule.Action.ApplyTo
	if len(targets) == 0 {
		targets = []string{"total"}
	}

	appliedAmount := 0.0

	for _, target := range targets {
		if target == "total" || target == "subtotal" {
			// 對總價扣減
			// 需要按比例分配到各個項目
			totalPrice := 0.0
			for _, item := range breakdown.Items {
				totalPrice += item.DiscountedPrice
			}

			if totalPrice > 0 {
				for _, item := range breakdown.Items {
					ratio := item.DiscountedPrice / totalPrice
					itemDiscount := discountAmount * ratio
					item.DiscountedPrice -= itemDiscount
					appliedAmount += itemDiscount
				}
			}
		} else {
			// 特定項目
			if item, exists := breakdown.Items[target]; exists {
				actualDiscount := discountAmount
				if actualDiscount > item.DiscountedPrice {
					actualDiscount = item.DiscountedPrice
				}
				item.DiscountedPrice -= actualDiscount
				appliedAmount += actualDiscount
			}
		}
	}

	if appliedAmount > 0 {
		breakdown.Discounts = append(breakdown.Discounts, dsl.DiscountItem{
			RuleID:    rule.ID,
			Label:     rule.Action.Label,
			Type:      "fixed",
			Value:     discountAmount,
			Amount:    appliedAmount,
			AppliedTo: strings.Join(targets, ", "),
		})
	}

	return nil
}

// applyPriceCap 套用價格封頂
func (c *Calculator) applyPriceCap(breakdown *dsl.PriceBreakdown, action *dsl.Action) error {
	capPrice, err := c.resolveValue(action.Value)
	if err != nil {
		return err
	}

	targets := action.ApplyTo
	if len(targets) == 0 {
		targets = []string{"total"}
	}

	for _, target := range targets {
		if target == "total" {
			// 封頂總價
			totalPrice := 0.0
			for _, item := range breakdown.Items {
				totalPrice += item.DiscountedPrice
			}

			if totalPrice > capPrice {
				// 按比例調整各項目價格
				ratio := capPrice / totalPrice
				for _, item := range breakdown.Items {
					item.DiscountedPrice = item.DiscountedPrice * ratio
				}
			}
		} else {
			// 封頂特定項目
			if item, exists := breakdown.Items[target]; exists {
				if item.DiscountedPrice > capPrice {
					item.DiscountedPrice = capPrice
				}
			}
		}
	}

	return nil
}

// calculateSubTotal 計算小計
func (c *Calculator) calculateSubTotal(breakdown *dsl.PriceBreakdown) float64 {
	total := 0.0
	for _, item := range breakdown.Items {
		total += item.OriginalPrice
	}
	return total
}

// calculateTotalDiscount 計算總折扣
func (c *Calculator) calculateTotalDiscount(breakdown *dsl.PriceBreakdown) float64 {
	originalTotal := 0.0
	discountedTotal := 0.0

	for _, item := range breakdown.Items {
		originalTotal += item.OriginalPrice
		discountedTotal += item.DiscountedPrice
	}

	return originalTotal - discountedTotal
}

// calculateFinalPrice 計算最終價格
func (c *Calculator) calculateFinalPrice(breakdown *dsl.PriceBreakdown) float64 {
	total := 0.0
	for _, item := range breakdown.Items {
		item.FinalPrice = item.DiscountedPrice
		total += item.FinalPrice
	}
	return total
}

// resolveValue 解析值（支援變數引用，如 $variables.price）
func (c *Calculator) resolveValue(v interface{}) (float64, error) {
	// 如果是字符串，檢查是否為變數引用
	if strVal, ok := v.(string); ok {
		if strings.HasPrefix(strVal, "$variables.") {
			// 解析變數引用
			varName := strings.TrimPrefix(strVal, "$variables.")
			if c.ruleSet.Variables != nil {
				if varValue, exists := c.ruleSet.Variables[varName]; exists {
					return c.toFloat64(varValue)
				}
			}
			return 0, fmt.Errorf("variable not found: %s", varName)
		}
		// 如果是純字符串但不是變數引用，返回錯誤
		return 0, fmt.Errorf("cannot convert string to float64: %s", strVal)
	}

	// 否則直接轉換
	return c.toFloat64(v)
}

// toFloat64 轉換為 float64
func (c *Calculator) toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// ExpressionEvaluator 表達式評估器介面
type ExpressionEvaluator interface {
	EvaluateExpression(expr *dsl.Expression, ctx *dsl.Context) (bool, error)
	GetFieldValue(field string, ctx *dsl.Context) (interface{}, error)
}
