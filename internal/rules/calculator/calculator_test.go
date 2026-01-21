package calculator

import (
	"testing"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExpressionEvaluator is a mock implementation of ExpressionEvaluator for testing
type MockExpressionEvaluator struct {
	ctx *dsl.Context
}

func (m *MockExpressionEvaluator) EvaluateExpression(expr *dsl.Expression, ctx *dsl.Context) (bool, error) {
	// Simple mock that always returns true for testing
	return true, nil
}

func (m *MockExpressionEvaluator) GetFieldValue(field string, ctx *dsl.Context) (interface{}, error) {
	// Mock field value retrieval
	switch field {
	case "team_size":
		return ctx.TeamSize, nil
	case "$computed.subtotal":
		if ctx.ComputedValues != nil {
			return ctx.ComputedValues["subtotal"], nil
		}
		return 0.0, nil
	case "$computed.total_water_bottles":
		if ctx.ComputedValues != nil {
			return ctx.ComputedValues["total_water_bottles"], nil
		}
		return 0, nil
	default:
		return nil, nil
	}
}

func createTestContext() *dsl.Context {
	return &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"race_type": "full_marathon",
		},
		TeamSize:       1,
		Variables:      map[string]interface{}{},
		ComputedValues: map[string]interface{}{},
	}
}

// TestSetPrice tests set_price action
func TestSetPrice(t *testing.T) {
	tests := []struct {
		name           string
		action         *dsl.Action
		expectedPrice  float64
		expectedLabel  string
		expectedItemID string
	}{
		{
			name: "set price with explicit item ID",
			action: &dsl.Action{
				Type:  "set_price",
				Item:  "registration_fee",
				Value: 1050,
				Label: "Marathon Registration Fee",
			},
			expectedPrice:  1050,
			expectedLabel:  "Marathon Registration Fee",
			expectedItemID: "registration_fee",
		},
		{
			name: "set price without item ID (default)",
			action: &dsl.Action{
				Type:  "set_price",
				Value: 950,
				Label: "Half Marathon Fee",
			},
			expectedPrice:  950,
			expectedLabel:  "Half Marathon Fee",
			expectedItemID: "registration_fee",
		},
		{
			name: "set price with float value",
			action: &dsl.Action{
				Type:  "set_price",
				Item:  "registration_fee",
				Value: 1050.50,
				Label: "Registration Fee",
			},
			expectedPrice:  1050.50,
			expectedLabel:  "Registration Fee",
			expectedItemID: "registration_fee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleSet := &dsl.RuleSet{
				EventID: "test-event",
			}
			calc := NewCalculator(ruleSet)
			breakdown := &dsl.PriceBreakdown{
				Items: make(map[string]*dsl.PriceItem),
			}

			err := calc.setPrice(breakdown, tt.action)

			require.NoError(t, err)
			assert.Contains(t, breakdown.Items, tt.expectedItemID)

			item := breakdown.Items[tt.expectedItemID]
			assert.Equal(t, tt.expectedPrice, item.UnitPrice)
			assert.Equal(t, tt.expectedPrice, item.OriginalPrice)
			assert.Equal(t, tt.expectedPrice, item.FinalPrice)
			assert.Equal(t, tt.expectedLabel, item.Name)
			assert.Equal(t, 1, item.Quantity)
		})
	}
}

// TestSetPrice_WithVariables tests set_price with variable references
func TestSetPrice_WithVariables(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		Variables: map[string]interface{}{
			"full_marathon_price": 1050,
			"half_marathon_price": 950,
		},
	}

	tests := []struct {
		name          string
		value         interface{}
		expectedPrice float64
	}{
		{
			name:          "variable reference",
			value:         "$variables.full_marathon_price",
			expectedPrice: 1050,
		},
		{
			name:          "direct number",
			value:         850,
			expectedPrice: 850,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewCalculator(ruleSet)
			breakdown := &dsl.PriceBreakdown{
				Items: make(map[string]*dsl.PriceItem),
			}

			action := &dsl.Action{
				Type:  "set_price",
				Value: tt.value,
				Label: "Registration Fee",
			}

			err := calc.setPrice(breakdown, action)

			require.NoError(t, err)
			item := breakdown.Items["registration_fee"]
			assert.Equal(t, tt.expectedPrice, item.UnitPrice)
		})
	}
}

// TestAddItem tests add_item action
func TestAddItem(t *testing.T) {
	tests := []struct {
		name           string
		action         *dsl.Action
		teamSize       int
		expectedPrice  float64
		expectedQty    int
		expectedUnit   float64
	}{
		{
			name: "unit price * quantity",
			action: &dsl.Action{
				Type:          "add_item",
				Item:          "addon:insurance",
				UnitPrice:     91,
				QuantityField: "team_size",
				Label:         "Insurance",
			},
			teamSize:      3,
			expectedPrice: 273, // 91 * 3
			expectedQty:   3,
			expectedUnit:  91,
		},
		{
			name: "fixed price (no multiplication)",
			action: &dsl.Action{
				Type:       "add_item",
				Item:       "addon:shipping",
				FixedPrice: 150,
				Label:      "Shipping Fee",
			},
			teamSize:      3,
			expectedPrice: 150, // fixed, not multiplied
			expectedQty:   1,
			expectedUnit:  150,
		},
		{
			name: "unit price without quantity field (default to 1)",
			action: &dsl.Action{
				Type:      "add_item",
				Item:      "addon:medal",
				UnitPrice: 200,
				Label:     "Medal",
			},
			teamSize:      3,
			expectedPrice: 200, // 200 * 1
			expectedQty:   1,
			expectedUnit:  200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleSet := &dsl.RuleSet{
				EventID: "test-event",
			}
			calc := NewCalculator(ruleSet)
			breakdown := &dsl.PriceBreakdown{
				Items: make(map[string]*dsl.PriceItem),
			}

			ctx := createTestContext()
			ctx.TeamSize = tt.teamSize

			evaluator := &MockExpressionEvaluator{ctx: ctx}

			err := calc.addItem(breakdown, tt.action, ctx, evaluator)

			require.NoError(t, err)
			assert.Contains(t, breakdown.Items, tt.action.Item)

			item := breakdown.Items[tt.action.Item]
			assert.Equal(t, tt.expectedPrice, item.OriginalPrice)
			assert.Equal(t, tt.expectedQty, item.Quantity)
			assert.Equal(t, tt.expectedUnit, item.UnitPrice)
			assert.Equal(t, tt.action.Label, item.Name)
		})
	}
}

// TestAddItem_WithVariables tests add_item with variable references
func TestAddItem_WithVariables(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		Variables: map[string]interface{}{
			"insurance_price": 91,
			"shipping_price":  150,
		},
	}

	tests := []struct {
		name          string
		unitPrice     interface{}
		fixedPrice    interface{}
		teamSize      int
		expectedPrice float64
	}{
		{
			name:          "unit price from variable",
			unitPrice:     "$variables.insurance_price",
			fixedPrice:    nil,
			teamSize:      2,
			expectedPrice: 182, // 91 * 2
		},
		{
			name:          "fixed price from variable",
			unitPrice:     nil,
			fixedPrice:    "$variables.shipping_price",
			teamSize:      5,
			expectedPrice: 150, // fixed, not multiplied
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewCalculator(ruleSet)
			breakdown := &dsl.PriceBreakdown{
				Items: make(map[string]*dsl.PriceItem),
			}

			ctx := createTestContext()
			ctx.TeamSize = tt.teamSize

			evaluator := &MockExpressionEvaluator{ctx: ctx}

			action := &dsl.Action{
				Type:          "add_item",
				Item:          "addon:test",
				UnitPrice:     tt.unitPrice,
				FixedPrice:    tt.fixedPrice,
				QuantityField: "team_size",
				Label:         "Test Item",
			}

			err := calc.addItem(breakdown, action, ctx, evaluator)

			require.NoError(t, err)
			item := breakdown.Items["addon:test"]
			assert.Equal(t, tt.expectedPrice, item.OriginalPrice)
		})
	}
}

// TestAddItem_ZeroQuantity tests that items with zero quantity are not added
func TestAddItem_ZeroQuantity(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	calc := NewCalculator(ruleSet)
	breakdown := &dsl.PriceBreakdown{
		Items: make(map[string]*dsl.PriceItem),
	}

	ctx := createTestContext()
	ctx.TeamSize = 0

	evaluator := &MockExpressionEvaluator{ctx: ctx}

	action := &dsl.Action{
		Type:          "add_item",
		Item:          "addon:insurance",
		UnitPrice:     91,
		QuantityField: "team_size",
		Label:         "Insurance",
	}

	err := calc.addItem(breakdown, action, ctx, evaluator)

	require.NoError(t, err)
	assert.NotContains(t, breakdown.Items, "addon:insurance")
}

// TestPercentageDiscount tests percentage_discount action
func TestPercentageDiscount(t *testing.T) {
	tests := []struct {
		name             string
		initialItems     map[string]*dsl.PriceItem
		discountPercent  float64
		applyTo          []string
		expectedDiscount float64
		expectedPrices   map[string]float64
	}{
		{
			name: "discount on registration_fee",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
			},
			discountPercent:  10, // 10% off
			applyTo:          []string{"registration_fee"},
			expectedDiscount: 100,
			expectedPrices: map[string]float64{
				"registration_fee": 900, // 1000 - 100
			},
		},
		{
			name: "discount on specific addon",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   100,
					DiscountedPrice: 100,
				},
			},
			discountPercent:  20, // 20% off
			applyTo:          []string{"addon:insurance"},
			expectedDiscount: 20,
			expectedPrices: map[string]float64{
				"registration_fee": 1000, // unchanged
				"addon:insurance":  80,   // 100 - 20
			},
		},
		{
			name: "discount on all addons (wildcard)",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   100,
					DiscountedPrice: 100,
				},
				"addon:medal": {
					ID:              "addon:medal",
					OriginalPrice:   200,
					DiscountedPrice: 200,
				},
			},
			discountPercent:  15, // 15% off
			applyTo:          []string{"addon:*"},
			expectedDiscount: 45, // 100*0.15 + 200*0.15
			expectedPrices: map[string]float64{
				"registration_fee": 1000, // unchanged
				"addon:insurance":  85,   // 100 - 15
				"addon:medal":      170,  // 200 - 30
			},
		},
		{
			name: "discount on total (all items)",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   100,
					DiscountedPrice: 100,
				},
			},
			discountPercent:  10, // 10% off
			applyTo:          []string{"total"},
			expectedDiscount: 110, // (1000 + 100) * 0.10
			expectedPrices: map[string]float64{
				"registration_fee": 900, // 1000 - 100
				"addon:insurance":  90,  // 100 - 10
			},
		},
		{
			name: "empty apply_to defaults to registration_fee",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   100,
					DiscountedPrice: 100,
				},
			},
			discountPercent:  10, // 10% off
			applyTo:          []string{},
			expectedDiscount: 100, // 1000 * 0.10
			expectedPrices: map[string]float64{
				"registration_fee": 900, // 1000 - 100
				"addon:insurance":  100, // unchanged
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleSet := &dsl.RuleSet{
				EventID: "test-event",
			}
			calc := NewCalculator(ruleSet)
			breakdown := &dsl.PriceBreakdown{
				Items:     tt.initialItems,
				Discounts: []dsl.DiscountItem{},
			}

			rule := &dsl.PricingRule{
				ID:       "test_discount",
				Priority: 100,
				Action: &dsl.Action{
					Type:    "percentage_discount",
					Value:   tt.discountPercent,
					ApplyTo: tt.applyTo,
					Label:   "Test Discount",
				},
			}

			err := calc.applyPercentageDiscount(breakdown, rule)

			require.NoError(t, err)

			// Check discounted prices
			for itemID, expectedPrice := range tt.expectedPrices {
				item := breakdown.Items[itemID]
				assert.InDelta(t, expectedPrice, item.DiscountedPrice, 0.01,
					"Item %s: expected %.2f, got %.2f", itemID, expectedPrice, item.DiscountedPrice)
			}

			// Check discount record
			if tt.expectedDiscount > 0 {
				assert.Len(t, breakdown.Discounts, 1)
				assert.InDelta(t, tt.expectedDiscount, breakdown.Discounts[0].Amount, 0.01)
				assert.Equal(t, "percentage", breakdown.Discounts[0].Type)
			}
		})
	}
}

// TestFixedDiscount tests fixed_discount action
func TestFixedDiscount(t *testing.T) {
	tests := []struct {
		name             string
		initialItems     map[string]*dsl.PriceItem
		discountAmount   float64
		applyTo          []string
		expectedDiscount float64
		expectedPrices   map[string]float64
	}{
		{
			name: "fixed discount on total (proportional)",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   500,
					DiscountedPrice: 500,
				},
			},
			discountAmount:   300, // total 1500, discount 300
			applyTo:          []string{"total"},
			expectedDiscount: 300,
			expectedPrices: map[string]float64{
				"registration_fee": 800, // 1000 - (1000/1500 * 300) = 1000 - 200
				"addon:insurance":  400, // 500 - (500/1500 * 300) = 500 - 100
			},
		},
		{
			name: "fixed discount on specific item",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   500,
					DiscountedPrice: 500,
				},
			},
			discountAmount:   200,
			applyTo:          []string{"registration_fee"},
			expectedDiscount: 200,
			expectedPrices: map[string]float64{
				"registration_fee": 800, // 1000 - 200
				"addon:insurance":  500, // unchanged
			},
		},
		{
			name: "fixed discount larger than item price",
			initialItems: map[string]*dsl.PriceItem{
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   100,
					DiscountedPrice: 100,
				},
			},
			discountAmount:   200, // larger than item price
			applyTo:          []string{"addon:insurance"},
			expectedDiscount: 100, // capped at item price
			expectedPrices: map[string]float64{
				"addon:insurance": 0, // cannot go below 0
			},
		},
		{
			name: "empty apply_to defaults to total",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
			},
			discountAmount:   100,
			applyTo:          []string{},
			expectedDiscount: 100,
			expectedPrices: map[string]float64{
				"registration_fee": 900,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleSet := &dsl.RuleSet{
				EventID: "test-event",
			}
			calc := NewCalculator(ruleSet)
			breakdown := &dsl.PriceBreakdown{
				Items:     tt.initialItems,
				Discounts: []dsl.DiscountItem{},
			}

			rule := &dsl.PricingRule{
				ID:       "test_discount",
				Priority: 110,
				Action: &dsl.Action{
					Type:    "fixed_discount",
					Value:   tt.discountAmount,
					ApplyTo: tt.applyTo,
					Label:   "Fixed Discount",
				},
			}

			err := calc.applyFixedDiscount(breakdown, rule)

			require.NoError(t, err)

			// Check discounted prices
			for itemID, expectedPrice := range tt.expectedPrices {
				item := breakdown.Items[itemID]
				assert.InDelta(t, expectedPrice, item.DiscountedPrice, 0.01,
					"Item %s: expected %.2f, got %.2f", itemID, expectedPrice, item.DiscountedPrice)
			}

			// Check discount record
			if tt.expectedDiscount > 0 {
				assert.Len(t, breakdown.Discounts, 1)
				assert.InDelta(t, tt.expectedDiscount, breakdown.Discounts[0].Amount, 0.01)
				assert.Equal(t, "fixed", breakdown.Discounts[0].Type)
			}
		})
	}
}

// TestPriceCap tests price_cap action
func TestPriceCap(t *testing.T) {
	tests := []struct {
		name           string
		initialItems   map[string]*dsl.PriceItem
		capPrice       float64
		applyTo        []string
		expectedPrices map[string]float64
	}{
		{
			name: "cap total price",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
				"addon:insurance": {
					ID:              "addon:insurance",
					OriginalPrice:   500,
					DiscountedPrice: 500,
				},
			},
			capPrice: 1200, // total is 1500, cap at 1200
			applyTo:  []string{"total"},
			expectedPrices: map[string]float64{
				"registration_fee": 800, // 1000 * (1200/1500)
				"addon:insurance":  400, // 500 * (1200/1500)
			},
		},
		{
			name: "cap specific item",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   1000,
					DiscountedPrice: 1000,
				},
			},
			capPrice: 800,
			applyTo:  []string{"registration_fee"},
			expectedPrices: map[string]float64{
				"registration_fee": 800,
			},
		},
		{
			name: "price below cap (no change)",
			initialItems: map[string]*dsl.PriceItem{
				"registration_fee": {
					ID:              "registration_fee",
					OriginalPrice:   500,
					DiscountedPrice: 500,
				},
			},
			capPrice: 1000,
			applyTo:  []string{"registration_fee"},
			expectedPrices: map[string]float64{
				"registration_fee": 500, // unchanged
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleSet := &dsl.RuleSet{
				EventID: "test-event",
			}
			calc := NewCalculator(ruleSet)
			breakdown := &dsl.PriceBreakdown{
				Items: tt.initialItems,
			}

			action := &dsl.Action{
				Type:    "price_cap",
				Value:   tt.capPrice,
				ApplyTo: tt.applyTo,
				Label:   "Price Cap",
			}

			err := calc.applyPriceCap(breakdown, action)

			require.NoError(t, err)

			// Check capped prices
			for itemID, expectedPrice := range tt.expectedPrices {
				item := breakdown.Items[itemID]
				assert.InDelta(t, expectedPrice, item.DiscountedPrice, 0.01,
					"Item %s: expected %.2f, got %.2f", itemID, expectedPrice, item.DiscountedPrice)
			}
		})
	}
}

// TestCalculate_PriceCalculations tests price calculation methods
func TestCalculate_PriceCalculations(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	calc := NewCalculator(ruleSet)

	breakdown := &dsl.PriceBreakdown{
		Items: map[string]*dsl.PriceItem{
			"registration_fee": {
				ID:              "registration_fee",
				OriginalPrice:   1000,
				DiscountedPrice: 900,
				FinalPrice:      0,
			},
			"addon:insurance": {
				ID:              "addon:insurance",
				OriginalPrice:   200,
				DiscountedPrice: 180,
				FinalPrice:      0,
			},
		},
	}

	// Test calculateSubTotal
	subtotal := calc.calculateSubTotal(breakdown)
	assert.Equal(t, 1200.0, subtotal)

	// Test calculateTotalDiscount
	totalDiscount := calc.calculateTotalDiscount(breakdown)
	assert.Equal(t, 120.0, totalDiscount) // (1000-900) + (200-180)

	// Test calculateFinalPrice
	finalPrice := calc.calculateFinalPrice(breakdown)
	assert.Equal(t, 1080.0, finalPrice) // 900 + 180
}

// TestCalculate_NegativePriceHandling tests that final price cannot be negative
func TestCalculate_NegativePriceHandling(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		PricingRules: []*dsl.PricingRule{
			{
				ID:       "set_base_price",
				Priority: 0,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:  "set_price",
					Value: 100,
					Label: "Base Price",
				},
			},
			{
				ID:       "huge_discount",
				Priority: 100,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:  "fixed_discount",
					Value: 200, // discount > price
					Label: "Huge Discount",
				},
			},
		},
	}

	calc := NewCalculator(ruleSet)
	ctx := createTestContext()
	evaluator := &MockExpressionEvaluator{ctx: ctx}

	breakdown, _, err := calc.Calculate(ctx, evaluator)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, breakdown.FinalPrice, 0.0, "Final price should not be negative")
}

// TestCalculate_PhaseExecution tests that rules execute in correct phase order
func TestCalculate_PhaseExecution(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		PricingRules: []*dsl.PricingRule{
			{
				ID:       "add_insurance",
				Priority: 20, // Phase 1
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:      "add_item",
					Item:      "addon:insurance",
					UnitPrice: 100,
					Label:     "Insurance",
				},
			},
			{
				ID:       "set_base_price",
				Priority: 0, // Phase 1 (should execute first)
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:  "set_price",
					Value: 1000,
					Label: "Registration Fee",
				},
			},
			{
				ID:       "discount",
				Priority: 100, // Phase 2
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:    "percentage_discount",
					Value:   10,
					ApplyTo: []string{"total"},
					Label:   "10% Off",
				},
			},
		},
	}

	calc := NewCalculator(ruleSet)
	ctx := createTestContext()
	evaluator := &MockExpressionEvaluator{ctx: ctx}

	breakdown, appliedRules, err := calc.Calculate(ctx, evaluator)

	require.NoError(t, err)
	assert.NotNil(t, breakdown)

	// Check that rules were applied in priority order
	assert.Equal(t, []string{"set_base_price", "add_insurance", "discount"}, appliedRules)

	// Check final calculations
	// Original: 1000 + 100 = 1100
	// After 10% discount: 1100 * 0.9 = 990
	assert.InDelta(t, 990.0, breakdown.FinalPrice, 0.01)
}
