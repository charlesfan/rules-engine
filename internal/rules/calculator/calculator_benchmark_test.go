package calculator

import (
	"testing"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
)

// MockEvaluatorForBenchmark is a simple mock for benchmarking
type MockEvaluatorForBenchmark struct{}

func (m *MockEvaluatorForBenchmark) EvaluateExpression(expr *dsl.Expression, ctx *dsl.Context) (bool, error) {
	return true, nil
}

func (m *MockEvaluatorForBenchmark) GetFieldValue(field string, ctx *dsl.Context) (interface{}, error) {
	switch field {
	case "team_size":
		return ctx.TeamSize, nil
	default:
		return nil, nil
	}
}

// BenchmarkCalculate_SimplePrice benchmarks simple price calculation
func BenchmarkCalculate_SimplePrice(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		PricingRules: []*dsl.PricingRule{
			{
				ID:       "base_price",
				Priority: 0,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:  "set_price",
					Value: 1000,
					Label: "Registration Fee",
				},
			},
		},
	}

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"race_type": "full_marathon",
		},
		TeamSize: 1,
	}

	calc := NewCalculator(ruleSet)
	evaluator := &MockEvaluatorForBenchmark{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = calc.Calculate(ctx, evaluator)
	}
}

// BenchmarkCalculate_WithDiscounts benchmarks calculation with multiple discounts
func BenchmarkCalculate_WithDiscounts(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		Variables: map[string]interface{}{
			"base_price": 1000,
		},
		PricingRules: []*dsl.PricingRule{
			{
				ID:       "base_price",
				Priority: 0,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:  "set_price",
					Value: "$variables.base_price",
					Label: "Registration Fee",
				},
			},
			{
				ID:       "early_bird_discount",
				Priority: 100,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:    "percentage_discount",
					Value:   20,
					ApplyTo: []string{"registration_fee"},
					Label:   "Early Bird 20%",
				},
			},
			{
				ID:       "group_discount",
				Priority: 110,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:    "percentage_discount",
					Value:   10,
					ApplyTo: []string{"registration_fee"},
					Label:   "Group 10%",
				},
			},
		},
	}

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"race_type": "full_marathon",
		},
		TeamSize: 1,
	}

	calc := NewCalculator(ruleSet)
	evaluator := &MockEvaluatorForBenchmark{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = calc.Calculate(ctx, evaluator)
	}
}

// BenchmarkCalculate_WithAddons benchmarks calculation with addons
func BenchmarkCalculate_WithAddons(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		PricingRules: []*dsl.PricingRule{
			{
				ID:       "base_price",
				Priority: 0,
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
				ID:       "insurance",
				Priority: 10,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:          "add_item",
					Item:          "addon:insurance",
					UnitPrice:     91,
					QuantityField: "team_size",
					Label:         "Insurance",
				},
			},
			{
				ID:       "tshirt",
				Priority: 11,
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				Action: &dsl.Action{
					Type:          "add_item",
					Item:          "addon:tshirt",
					UnitPrice:     200,
					QuantityField: "team_size",
					Label:         "T-Shirt",
				},
			},
		},
	}

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"race_type": "full_marathon",
		},
		TeamSize: 3,
	}

	calc := NewCalculator(ruleSet)
	evaluator := &MockEvaluatorForBenchmark{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = calc.Calculate(ctx, evaluator)
	}
}

// BenchmarkCalculate_ManyRules benchmarks calculation with many pricing rules
func BenchmarkCalculate_ManyRules(b *testing.B) {
	pricingRules := []*dsl.PricingRule{
		{
			ID:       "base_price",
			Priority: 0,
			Condition: &dsl.Expression{
				Type: "always_true",
			},
			Action: &dsl.Action{
				Type:  "set_price",
				Value: 1000,
				Label: "Registration Fee",
			},
		},
	}

	// Add 10 discount rules
	for i := 1; i <= 10; i++ {
		pricingRules = append(pricingRules, &dsl.PricingRule{
			ID:       "discount_" + string(rune(i)),
			Priority: 100 + i,
			Condition: &dsl.Expression{
				Type: "always_true",
			},
			Action: &dsl.Action{
				Type:    "percentage_discount",
				Value:   float64(i),
				ApplyTo: []string{"registration_fee"},
				Label:   "Discount",
			},
		})
	}

	ruleSet := &dsl.RuleSet{
		EventID:      "test-event",
		PricingRules: pricingRules,
	}

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"race_type": "full_marathon",
		},
		TeamSize: 1,
	}

	calc := NewCalculator(ruleSet)
	evaluator := &MockEvaluatorForBenchmark{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = calc.Calculate(ctx, evaluator)
	}
}

// BenchmarkApplyPercentageDiscount benchmarks percentage discount application
func BenchmarkApplyPercentageDiscount(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}

	calc := NewCalculator(ruleSet)

	breakdown := &dsl.PriceBreakdown{
		Items: map[string]*dsl.PriceItem{
			"registration_fee": {
				ID:              "registration_fee",
				OriginalPrice:   1000,
				DiscountedPrice: 1000,
			},
			"addon:insurance": {
				ID:              "addon:insurance",
				OriginalPrice:   200,
				DiscountedPrice: 200,
			},
			"addon:tshirt": {
				ID:              "addon:tshirt",
				OriginalPrice:   300,
				DiscountedPrice: 300,
			},
		},
		Discounts: []dsl.DiscountItem{},
	}

	rule := &dsl.PricingRule{
		ID:       "discount",
		Priority: 100,
		Action: &dsl.Action{
			Type:    "percentage_discount",
			Value:   10,
			ApplyTo: []string{"total"},
			Label:   "10% Off",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset prices
		breakdown.Items["registration_fee"].DiscountedPrice = 1000
		breakdown.Items["addon:insurance"].DiscountedPrice = 200
		breakdown.Items["addon:tshirt"].DiscountedPrice = 300
		breakdown.Discounts = []dsl.DiscountItem{}

		_ = calc.applyPercentageDiscount(breakdown, rule)
	}
}
