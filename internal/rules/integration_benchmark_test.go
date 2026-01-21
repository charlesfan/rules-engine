package rules

import (
	"testing"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/charlesfan/rules-engine/internal/rules/evaluator"
	"github.com/charlesfan/rules-engine/internal/rules/parser"
)

// BenchmarkEndToEnd_SimpleRegistration benchmarks a simple end-to-end registration
func BenchmarkEndToEnd_SimpleRegistration(b *testing.B) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"variables": {
			"full_marathon_price": 1050
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

	p := parser.NewParser()
	ruleSet, _ := p.Parse(dslJSON)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name":      "John Doe",
			"race_type": "full_marathon",
		},
		TeamSize: 1,
	}

	eval := evaluator.NewEvaluator(ruleSet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eval.Evaluate(ctx)
	}
}

// BenchmarkEndToEnd_ComplexScenario benchmarks a complex scenario with discounts and addons
func BenchmarkEndToEnd_ComplexScenario(b *testing.B) {
	dslJSON := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"variables": {
			"full_marathon_price": 1050,
			"insurance_price": 91
		},
		"rule_definitions": {
			"is_early_bird": {
				"type": "condition",
				"description": "Early bird period",
				"expression": {
					"type": "datetime_before",
					"field": "register_date",
					"value": "2025-10-01T00:00:00Z"
				}
			}
		},
		"computed_fields": {
			"subtotal": {
				"type": "sum_prices",
				"items": ["registration_fee", "addon:*"]
			}
		},
		"pricing_rules": [
			{
				"id": "full_marathon_price",
				"priority": 0,
				"description": "Full marathon price",
				"condition": {
					"type": "equals",
					"field": "user.race_type",
					"value": "full_marathon"
				},
				"action": {
					"type": "set_price",
					"item": "registration_fee",
					"value": "$variables.full_marathon_price",
					"label": "Full Marathon Fee"
				}
			},
			{
				"id": "insurance",
				"priority": 10,
				"description": "Insurance addon",
				"condition": {
					"type": "equals",
					"field": "user.add_insurance",
					"value": true
				},
				"action": {
					"type": "add_item",
					"item": "addon:insurance",
					"unit_price": "$variables.insurance_price",
					"label": "Insurance"
				}
			},
			{
				"id": "early_bird_discount",
				"priority": 100,
				"description": "Early bird discount",
				"condition": {
					"type": "rule_ref",
					"rule": "is_early_bird"
				},
				"action": {
					"type": "percentage_discount",
					"value": 20,
					"apply_to": ["registration_fee"],
					"label": "Early Bird 20% Off"
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
					"value": 1500
				},
				"action": {
					"type": "percentage_discount",
					"value": 5,
					"apply_to": ["total"],
					"label": "Volume 5% Off"
				}
			}
		],
		"validation_rules": [
			{
				"id": "age_limit",
				"description": "Age must be at least 18",
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
	ruleSet, _ := p.Parse(dslJSON)

	ctx := &dsl.Context{
		RegisterDate: time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC), // Early bird
		User: map[string]interface{}{
			"name":          "John Doe",
			"age":           25,
			"race_type":     "full_marathon",
			"add_insurance": true,
		},
		TeamSize: 1,
	}

	eval := evaluator.NewEvaluator(ruleSet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eval.Evaluate(ctx)
	}
}

// BenchmarkEndToEnd_TeamRegistration benchmarks team registration with array operations
func BenchmarkEndToEnd_TeamRegistration(b *testing.B) {
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
					"type": "always_true"
				},
				"action": {
					"type": "add_item",
					"item": "addon:insurance",
					"unit_price": "$variables.insurance_per_person",
					"quantity_field": "team_size",
					"label": "Team Insurance"
				}
			},
			{
				"id": "team_discount",
				"priority": 100,
				"description": "Team size discount",
				"condition": {
					"type": "compare",
					"field": "team_size",
					"operator": ">=",
					"value": 3
				},
				"action": {
					"type": "percentage_discount",
					"value": 10,
					"apply_to": ["total"],
					"label": "Team 10% Off"
				}
			}
		],
		"validation_rules": [
			{
				"id": "has_female_member",
				"description": "Team must have at least one female member",
				"condition": {
					"type": "not",
					"condition": {
						"type": "array_any",
						"array": "team.members",
						"condition": {
							"type": "equals",
							"field": "user.gender",
							"value": "female"
						}
					}
				},
				"error_type": "blocking",
				"error_message": "Team must have at least one female member"
			}
		]
	}`)

	p := parser.NewParser()
	ruleSet, _ := p.Parse(dslJSON)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User:         map[string]interface{}{},
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{"name": "Alice", "gender": "female", "age": 25},
				map[string]interface{}{"name": "Bob", "gender": "male", "age": 30},
				map[string]interface{}{"name": "Charlie", "gender": "male", "age": 28},
			},
		},
		TeamSize: 3,
	}

	eval := evaluator.NewEvaluator(ruleSet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eval.Evaluate(ctx)
	}
}

// BenchmarkParseAndEvaluate benchmarks parsing + evaluation together
func BenchmarkParseAndEvaluate(b *testing.B) {
	dslJSON := []byte(`{
		"event_id": "test-event",
		"version": "1.0",
		"name": "Test Event",
		"variables": {
			"price": 1000
		},
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"condition": {"type": "always_true"},
				"action": {
					"type": "set_price",
					"value": "$variables.price",
					"label": "Registration Fee"
				}
			}
		],
		"validation_rules": []
	}`)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User:         map[string]interface{}{"name": "John"},
		TeamSize:     1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := parser.NewParser()
		ruleSet, _ := p.Parse(dslJSON)
		eval := evaluator.NewEvaluator(ruleSet)
		_, _ = eval.Evaluate(ctx)
	}
}
