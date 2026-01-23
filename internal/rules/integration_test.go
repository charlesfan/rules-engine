package rules

import (
	"testing"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/charlesfan/rules-engine/internal/rules/evaluator"
	"github.com/charlesfan/rules-engine/internal/rules/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_SimpleMarathonRegistration tests basic marathon registration scenario
func TestIntegration_SimpleMarathonRegistration(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"variables": {
			"full_marathon_price": 1050,
			"half_marathon_price": 950
		},
		"pricing_rules": [
			{
				"id": "full_marathon_price",
				"priority": 0,
				"description": "Full marathon registration fee",
				"condition": {
					"type": "equals",
					"field": "user.race_type",
					"value": "full_marathon"
				},
				"action": {
					"type": "set_price",
					"item": "registration_fee",
					"value": "$variables.full_marathon_price",
					"label": "Full Marathon Registration Fee"
				}
			}
		],
		"validation_rules": []
	}`)

	// Parse DSL
	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	// Create context
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name":      "John Doe",
			"race_type": "full_marathon",
		},
		TeamSize: 1,
	}

	// Evaluate
	eval := evaluator.NewEvaluator(ruleSet)
	result, err := eval.Evaluate(ctx)

	require.NoError(t, err)
	assert.True(t, result.Valid)
	assert.NotNil(t, result.Price)
	assert.Equal(t, 1050.0, result.Price.FinalPrice)
	assert.Equal(t, 1050.0, result.Price.SubTotal)
	assert.Equal(t, 0.0, result.Price.TotalDiscount)
}

// TestIntegration_TeamRegistration tests team registration with quantity-based pricing
func TestIntegration_TeamRegistration(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "relay-race-2025",
		"version": "1.0",
		"name": "Relay Race 2025",
		"variables": {
			"per_person_fee": 500,
			"insurance_per_person": 91
		},
		"pricing_rules": [
			{
				"id": "team_registration_fee",
				"priority": 0,
				"description": "Team registration fee",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "add_item",
					"item": "registration_fee",
					"unit_price": "$variables.per_person_fee",
					"quantity_field": "team_size",
					"label": "Team Registration Fee"
				}
			},
			{
				"id": "team_insurance",
				"priority": 20,
				"description": "Team insurance",
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
					"label": "Team Insurance"
				}
			}
		],
		"validation_rules": []
	}`)

	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User:         map[string]interface{}{},
		TeamSize:     4,
		Addons: map[string]interface{}{
			"insurance": true,
		},
	}

	eval := evaluator.NewEvaluator(ruleSet)
	result, err := eval.Evaluate(ctx)

	require.NoError(t, err)
	assert.True(t, result.Valid)

	// 4 people * 500 = 2000
	// 4 people * 91 = 364
	// Total = 2364
	assert.Equal(t, 2364.0, result.Price.FinalPrice)
	assert.Equal(t, 2364.0, result.Price.SubTotal)
}

// TestIntegration_EarlyBirdDiscount tests early bird discount with datetime condition
func TestIntegration_EarlyBirdDiscount(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Base registration fee",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "set_price",
					"value": 1000,
					"label": "Registration Fee"
				}
			},
			{
				"id": "early_bird_discount",
				"priority": 100,
				"description": "Early bird 10% discount",
				"condition": {
					"type": "datetime_before",
					"field": "register_date",
					"value": "2025-10-01T00:00:00Z"
				},
				"action": {
					"type": "percentage_discount",
					"value": 10,
					"apply_to": ["registration_fee"],
					"label": "Early Bird 10% Off"
				}
			}
		],
		"validation_rules": []
	}`)

	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	tests := []struct {
		name             string
		registerDate     time.Time
		expectedPrice    float64
		expectedDiscount float64
	}{
		{
			name:             "early bird (before deadline)",
			registerDate:     time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
			expectedPrice:    900, // 1000 * 0.9
			expectedDiscount: 100,
		},
		{
			name:             "regular (after deadline)",
			registerDate:     time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
			expectedPrice:    1000,
			expectedDiscount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &dsl.Context{
				RegisterDate: tt.registerDate,
				User:         map[string]interface{}{},
				TeamSize:     1,
			}

			eval := evaluator.NewEvaluator(ruleSet)
			result, err := eval.Evaluate(ctx)

			require.NoError(t, err)
			assert.True(t, result.Valid)
			assert.Equal(t, tt.expectedPrice, result.Price.FinalPrice)
			assert.Equal(t, tt.expectedDiscount, result.Price.TotalDiscount)
		})
	}
}

// TestIntegration_AddonsWithShipping tests addons with conditional shipping fees
func TestIntegration_AddonsWithShipping(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Base registration fee",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "set_price",
					"value": 1000,
					"label": "Registration Fee"
				}
			},
			{
				"id": "shipping_1_to_2",
				"priority": 10,
				"description": "Shipping fee for 1-2 people",
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
					"label": "Shipping (1-2 people)"
				}
			},
			{
				"id": "shipping_3_to_5",
				"priority": 10,
				"description": "Shipping fee for 3-5 people",
				"condition": {
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
				},
				"action": {
					"type": "add_item",
					"item": "addon:shipping",
					"fixed_price": 200,
					"label": "Shipping (3-5 people)"
				}
			}
		],
		"validation_rules": []
	}`)

	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	tests := []struct {
		name          string
		teamSize      int
		expectedTotal float64
		hasShipping   bool
		shippingFee   float64
	}{
		{
			name:          "1 person - shipping 150",
			teamSize:      1,
			expectedTotal: 1150,
			hasShipping:   true,
			shippingFee:   150,
		},
		{
			name:          "2 people - shipping 150",
			teamSize:      2,
			expectedTotal: 1150,
			hasShipping:   true,
			shippingFee:   150,
		},
		{
			name:          "4 people - shipping 200",
			teamSize:      4,
			expectedTotal: 1200,
			hasShipping:   true,
			shippingFee:   200,
		},
		{
			name:          "6 people - no shipping rule",
			teamSize:      6,
			expectedTotal: 1000,
			hasShipping:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &dsl.Context{
				RegisterDate: time.Now(),
				User:         map[string]interface{}{},
				TeamSize:     tt.teamSize,
			}

			eval := evaluator.NewEvaluator(ruleSet)
			result, err := eval.Evaluate(ctx)

			require.NoError(t, err)
			assert.True(t, result.Valid)
			assert.Equal(t, tt.expectedTotal, result.Price.FinalPrice)

			if tt.hasShipping {
				assert.Contains(t, result.Price.Items, "addon:shipping")
				assert.Equal(t, tt.shippingFee, result.Price.Items["addon:shipping"].FinalPrice)
			} else {
				assert.NotContains(t, result.Price.Items, "addon:shipping")
			}
		})
	}
}

// TestIntegration_VolumeDiscountWithComputedField tests volume discount using computed field
func TestIntegration_VolumeDiscountWithComputedField(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"computed_fields": {
			"subtotal": {
				"type": "sum_prices",
				"description": "Subtotal of all items",
				"items": ["registration_fee", "addon:*"]
			}
		},
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Base registration fee",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "set_price",
					"value": 1000,
					"label": "Registration Fee"
				}
			},
			{
				"id": "insurance",
				"priority": 20,
				"description": "Insurance",
				"condition": {
					"type": "equals",
					"field": "addons.insurance",
					"value": true
				},
				"action": {
					"type": "add_item",
					"item": "addon:insurance",
					"unit_price": 500,
					"label": "Insurance"
				}
			},
			{
				"id": "volume_discount_1500",
				"priority": 110,
				"description": "Volume discount for orders over 1500",
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
					"label": "Volume Discount -200"
				}
			}
		],
		"validation_rules": []
	}`)

	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	tests := []struct {
		name             string
		hasInsurance     bool
		expectedPrice    float64
		expectedDiscount float64
	}{
		{
			name:             "without insurance (1000) - no discount",
			hasInsurance:     false,
			expectedPrice:    1000,
			expectedDiscount: 0,
		},
		{
			name:             "with insurance (1500) - volume discount applies",
			hasInsurance:     true,
			expectedPrice:    1300, // (1000 + 500) - 200
			expectedDiscount: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &dsl.Context{
				RegisterDate: time.Now(),
				User:         map[string]interface{}{},
				TeamSize:     1,
				Addons: map[string]interface{}{
					"insurance": tt.hasInsurance,
				},
			}

			eval := evaluator.NewEvaluator(ruleSet)
			result, err := eval.Evaluate(ctx)

			require.NoError(t, err)
			assert.True(t, result.Valid)
			assert.Equal(t, tt.expectedPrice, result.Price.FinalPrice)
			assert.Equal(t, tt.expectedDiscount, result.Price.TotalDiscount)
		})
	}
}

// TestIntegration_MultipleDiscounts tests stacking of multiple discounts
func TestIntegration_MultipleDiscounts(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Base registration fee",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "set_price",
					"value": 1000,
					"label": "Registration Fee"
				}
			},
			{
				"id": "discount_1",
				"priority": 100,
				"description": "First discount 10%",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "percentage_discount",
					"value": 10,
					"apply_to": ["registration_fee"],
					"label": "Discount 1 (10%)"
				}
			},
			{
				"id": "discount_2",
				"priority": 110,
				"description": "Second discount 5%",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "percentage_discount",
					"value": 5,
					"apply_to": ["registration_fee"],
					"label": "Discount 2 (5%)"
				}
			}
		],
		"validation_rules": []
	}`)

	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User:         map[string]interface{}{},
		TeamSize:     1,
	}

	eval := evaluator.NewEvaluator(ruleSet)
	result, err := eval.Evaluate(ctx)

	require.NoError(t, err)
	assert.True(t, result.Valid)

	// Multiplicative stacking:
	// After discount 1: 1000 * 0.9 = 900
	// After discount 2: 900 * 0.95 = 855
	assert.InDelta(t, 855.0, result.Price.FinalPrice, 0.01)
	assert.Len(t, result.Price.Discounts, 2)
}

// TestIntegration_ValidationRules tests blocking and warning validation rules
func TestIntegration_ValidationRules(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Base registration fee",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "set_price",
					"value": 1000,
					"label": "Registration Fee"
				}
			}
		],
		"validation_rules": [
			{
				"id": "age_limit",
				"description": "Age limit 18+",
				"condition": {
					"type": "compare",
					"field": "user.age",
					"operator": "<",
					"value": 18
				},
				"error_type": "blocking",
				"error_message": "Participants must be at least 18 years old"
			},
			{
				"id": "registration_period",
				"description": "Registration period check",
				"condition": {
					"type": "not",
					"condition": {
						"type": "datetime_between",
						"field": "register_date",
						"start": "2025-09-01T00:00:00Z",
						"end": "2025-12-31T23:59:59Z"
					}
				},
				"error_type": "blocking",
				"error_message": "Registration is only open from 2025-09-01 to 2025-12-31"
			},
			{
				"id": "health_warning",
				"description": "Health warning",
				"condition": {
					"type": "always_true"
				},
				"error_type": "warning",
				"error_message": "Please ensure you are in good health before participating"
			}
		]
	}`)

	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	tests := []struct {
		name             string
		age              int
		registerDate     time.Time
		expectedValid    bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name:             "valid registration",
			age:              25,
			registerDate:     time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // health warning
		},
		{
			name:             "age below limit",
			age:              15,
			registerDate:     time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
			expectedValid:    false,
			expectedErrors:   1, // age limit
			expectedWarnings: 1, // health warning
		},
		{
			name:             "outside registration period",
			age:              25,
			registerDate:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedValid:    false,
			expectedErrors:   1, // registration period
			expectedWarnings: 1, // health warning
		},
		{
			name:             "multiple blocking errors",
			age:              15,
			registerDate:     time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedValid:    false,
			expectedErrors:   2, // age + registration period
			expectedWarnings: 1, // health warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &dsl.Context{
				RegisterDate: tt.registerDate,
				User: map[string]interface{}{
					"age": tt.age,
				},
				TeamSize: 1,
			}

			eval := evaluator.NewEvaluator(ruleSet)
			result, err := eval.Evaluate(ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedValid, result.Valid)
			assert.Len(t, result.Errors, tt.expectedErrors)
			assert.Len(t, result.Warnings, tt.expectedWarnings)

			// Check that price is still calculated even with errors
			assert.NotNil(t, result.Price)
			assert.Equal(t, 1000.0, result.Price.FinalPrice)
		})
	}
}

// TestIntegration_CompleteMarathonScenario tests a complete, realistic marathon registration
func TestIntegration_CompleteMarathonScenario(t *testing.T) {
	dslJSON := []byte(`{
		"event_id": "taipei-marathon-2025",
		"version": "1.0",
		"name": "2025 Taipei Marathon",
		"variables": {
			"full_marathon_price": 1050,
			"half_marathon_price": 950,
			"insurance_price": 91
		},
		"computed_fields": {
			"subtotal": {
				"type": "sum_prices",
				"description": "Subtotal",
				"items": ["registration_fee", "addon:*"]
			}
		},
		"pricing_rules": [
			{
				"id": "full_marathon_price",
				"priority": 0,
				"description": "Full marathon",
				"condition": {
					"type": "equals",
					"field": "user.race_type",
					"value": "full_marathon"
				},
				"action": {
					"type": "set_price",
					"value": "$variables.full_marathon_price",
					"label": "Full Marathon (42km)"
				}
			},
			{
				"id": "half_marathon_price",
				"priority": 0,
				"description": "Half marathon",
				"condition": {
					"type": "equals",
					"field": "user.race_type",
					"value": "half_marathon"
				},
				"action": {
					"type": "set_price",
					"value": "$variables.half_marathon_price",
					"label": "Half Marathon (21km)"
				}
			},
			{
				"id": "insurance",
				"priority": 20,
				"description": "Insurance",
				"condition": {
					"type": "equals",
					"field": "addons.insurance",
					"value": true
				},
				"action": {
					"type": "add_item",
					"item": "addon:insurance",
					"unit_price": "$variables.insurance_price",
					"quantity_field": "team_size",
					"label": "Sports Injury Insurance"
				}
			},
			{
				"id": "early_bird",
				"priority": 100,
				"description": "Early bird discount",
				"condition": {
					"type": "datetime_before",
					"field": "register_date",
					"value": "2025-10-01T00:00:00+08:00"
				},
				"action": {
					"type": "percentage_discount",
					"value": 10,
					"apply_to": ["registration_fee"],
					"label": "Early Bird 10% Off"
				}
			},
			{
				"id": "volume_discount",
				"priority": 110,
				"description": "Volume discount",
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
					"label": "Volume Discount -200"
				}
			}
		],
		"validation_rules": [
			{
				"id": "age_limit",
				"description": "Age limit",
				"condition": {
					"type": "compare",
					"field": "user.age",
					"operator": "<",
					"value": 18
				},
				"error_type": "blocking",
				"error_message": "Participants must be at least 18 years old"
			}
		]
	}`)

	p := parser.NewParser()
	ruleSet, err := p.Parse(dslJSON)
	require.NoError(t, err)

	// Scenario: Early bird full marathon with insurance for 3 people
	ctx := &dsl.Context{
		RegisterDate: time.Date(2025, 9, 15, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		User: map[string]interface{}{
			"name":      "John Doe",
			"age":       30,
			"race_type": "full_marathon",
		},
		TeamSize: 3,
		Addons: map[string]interface{}{
			"insurance": true,
		},
	}

	eval := evaluator.NewEvaluator(ruleSet)
	result, err := eval.Evaluate(ctx)

	require.NoError(t, err)
	assert.True(t, result.Valid)

	// Calculations:
	// Base: 1050
	// Early bird 10% off: 1050 * 0.9 = 945
	// Insurance: 91 * 3 = 273
	// Subtotal: 945 + 273 = 1218
	// No volume discount (< 2000)
	// Final: 1218

	assert.InDelta(t, 1218.0, result.Price.FinalPrice, 0.01)
	assert.Contains(t, result.Price.Items, "registration_fee")
	assert.Contains(t, result.Price.Items, "addon:insurance")
	assert.Len(t, result.Price.Discounts, 1) // early bird only
	assert.Equal(t, "early_bird", result.Price.Discounts[0].RuleID)
}
