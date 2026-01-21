package computed

import (
	"fmt"
	"strings"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
)

// FieldEvaluator 計算欄位評估器
type FieldEvaluator struct {
	ruleSet *dsl.RuleSet
}

// NewFieldEvaluator 建立新的計算欄位評估器
func NewFieldEvaluator(ruleSet *dsl.RuleSet) *FieldEvaluator {
	return &FieldEvaluator{
		ruleSet: ruleSet,
	}
}

// Evaluate 評估所有計算欄位
func (e *FieldEvaluator) Evaluate(ctx *dsl.Context, breakdown *dsl.PriceBreakdown) error {
	if e.ruleSet.ComputedFields == nil {
		return nil
	}

	// 初始化 ComputedValues
	if ctx.ComputedValues == nil {
		ctx.ComputedValues = make(map[string]interface{})
	}

	// 評估每個計算欄位
	for name, field := range e.ruleSet.ComputedFields {
		value, err := e.evaluateField(field, ctx, breakdown)
		if err != nil {
			return fmt.Errorf("failed to evaluate computed field %s: %w", name, err)
		}
		ctx.ComputedValues[name] = value
	}

	return nil
}

// evaluateField 評估單一計算欄位
func (e *FieldEvaluator) evaluateField(field *dsl.ComputedField, ctx *dsl.Context, breakdown *dsl.PriceBreakdown) (interface{}, error) {
	switch field.Type {
	case "sum_prices":
		return e.sumPrices(field.Items, breakdown)

	case "count_items":
		return e.countItems(field.Items, breakdown)

	case "item_price":
		return e.itemPrice(field.Item, breakdown)

	case "count_array_field":
		return e.countArrayField(field, ctx)

	case "sum_array_field":
		return e.sumArrayField(field, ctx)

	default:
		return nil, fmt.Errorf("unknown computed field type: %s", field.Type)
	}
}

// sumPrices 計算價格總和
func (e *FieldEvaluator) sumPrices(items []string, breakdown *dsl.PriceBreakdown) (float64, error) {
	if breakdown == nil || breakdown.Items == nil {
		return 0, nil
	}

	total := 0.0
	for _, pattern := range items {
		// 支援萬用字元
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			for itemID, item := range breakdown.Items {
				if strings.HasPrefix(itemID, prefix) {
					total += item.DiscountedPrice
				}
			}
		} else {
			// 精確匹配
			if item, exists := breakdown.Items[pattern]; exists {
				total += item.DiscountedPrice
			}
		}
	}

	return total, nil
}

// countItems 計算項目數量
func (e *FieldEvaluator) countItems(items []string, breakdown *dsl.PriceBreakdown) (int, error) {
	if breakdown == nil || breakdown.Items == nil {
		return 0, nil
	}

	count := 0
	for _, pattern := range items {
		// 支援萬用字元
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			for itemID := range breakdown.Items {
				if strings.HasPrefix(itemID, prefix) {
					count++
				}
			}
		} else {
			// 精確匹配
			if _, exists := breakdown.Items[pattern]; exists {
				count++
			}
		}
	}

	return count, nil
}

// itemPrice 取得單一項目價格
func (e *FieldEvaluator) itemPrice(itemID string, breakdown *dsl.PriceBreakdown) (float64, error) {
	if breakdown == nil || breakdown.Items == nil {
		return 0, nil
	}

	if item, exists := breakdown.Items[itemID]; exists {
		return item.DiscountedPrice, nil
	}

	return 0, nil
}

// countArrayField 計算陣列中符合條件的欄位數量
func (e *FieldEvaluator) countArrayField(field *dsl.ComputedField, ctx *dsl.Context) (int, error) {
	// 取得陣列
	arrayValue, err := e.getFieldValue(field.Array, ctx)
	if err != nil {
		return 0, nil // 陣列不存在時返回 0
	}

	arraySlice, ok := arrayValue.([]interface{})
	if !ok {
		return 0, fmt.Errorf("field %s is not an array", field.Array)
	}

	count := 0
	for _, item := range arraySlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// 支援嵌套路徑，如 "addons.water_bottle.quantity"
		fieldValue := e.getNestedValue(itemMap, field.Field)
		if fieldValue == nil {
			continue
		}

		// 檢查是否符合指定的值
		if field.Value != nil {
			if fieldValue == field.Value {
				count++
			}
		} else {
			// 如果沒有指定值，只要欄位存在就計數
			count++
		}
	}

	return count, nil
}

// sumArrayField 加總陣列中指定欄位的數值
func (e *FieldEvaluator) sumArrayField(field *dsl.ComputedField, ctx *dsl.Context) (float64, error) {
	// 取得陣列
	arrayValue, err := e.getFieldValue(field.Array, ctx)
	if err != nil {
		return 0, nil // 陣列不存在時返回 0
	}

	arraySlice, ok := arrayValue.([]interface{})
	if !ok {
		return 0, fmt.Errorf("field %s is not an array", field.Array)
	}

	sum := 0.0
	for _, item := range arraySlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// 支援嵌套路徑，如 "addons.water_bottle.quantity"
		fieldValue := e.getNestedValue(itemMap, field.Field)
		if fieldValue == nil {
			continue
		}

		// 轉換為數字
		switch v := fieldValue.(type) {
		case int:
			sum += float64(v)
		case int64:
			sum += float64(v)
		case float32:
			sum += float64(v)
		case float64:
			sum += v
		}
	}

	return sum, nil
}

// getNestedValue 從 map 中取得嵌套路徑的值
func (e *FieldEvaluator) getNestedValue(m map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = m

	for _, part := range parts {
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = currentMap[part]
		if current == nil {
			return nil
		}
	}

	return current
}

// getFieldValue 從 Context 取得欄位值
func (e *FieldEvaluator) getFieldValue(fieldPath string, ctx *dsl.Context) (interface{}, error) {
	parts := strings.Split(fieldPath, ".")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty field path")
	}

	var current interface{}

	// 根據第一部分決定起始位置
	switch parts[0] {
	case "user":
		current = ctx.User
	case "team":
		current = ctx.Team
	case "addons":
		current = ctx.Addons
	default:
		return nil, fmt.Errorf("unknown root field: %s", parts[0])
	}

	// 遍歷剩餘部分
	for i := 1; i < len(parts); i++ {
		if current == nil {
			return nil, fmt.Errorf("field not found: %s", fieldPath)
		}

		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot traverse non-map field: %s", parts[i-1])
		}

		current = currentMap[parts[i]]
	}

	return current, nil
}
