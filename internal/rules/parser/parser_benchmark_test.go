package parser

import (
	"testing"
)

// BenchmarkParse_SimpleRuleSet benchmarks parsing a simple rule set
func BenchmarkParse_SimpleRuleSet(b *testing.B) {
	jsonData := []byte(`{
		"event_id": "test-event-001",
		"version": "1.0",
		"name": "Test Event",
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Base price",
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
		"validation_rules": []
	}`)

	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParse_ComplexRuleSet benchmarks parsing a complex rule set
func BenchmarkParse_ComplexRuleSet(b *testing.B) {
	jsonData := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"variables": {
			"full_marathon_price": 1050,
			"half_marathon_price": 950,
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
				"id": "half_marathon_price",
				"priority": 0,
				"description": "Half marathon price",
				"condition": {
					"type": "equals",
					"field": "user.race_type",
					"value": "half_marathon"
				},
				"action": {
					"type": "set_price",
					"item": "registration_fee",
					"value": "$variables.half_marathon_price",
					"label": "Half Marathon Fee"
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

	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParse_WithManyRules benchmarks parsing with many rules
func BenchmarkParse_WithManyRules(b *testing.B) {
	jsonData := []byte(`{
		"event_id": "test-event",
		"version": "1.0",
		"name": "Test Event",
		"pricing_rules": [
			{"id": "rule_1", "priority": 0, "condition": {"type": "always_true"}, "action": {"type": "set_price", "value": 1000, "label": "Rule 1"}},
			{"id": "rule_2", "priority": 1, "condition": {"type": "always_true"}, "action": {"type": "percentage_discount", "value": 5, "label": "Rule 2"}},
			{"id": "rule_3", "priority": 2, "condition": {"type": "always_true"}, "action": {"type": "percentage_discount", "value": 3, "label": "Rule 3"}},
			{"id": "rule_4", "priority": 3, "condition": {"type": "always_true"}, "action": {"type": "fixed_discount", "value": 50, "label": "Rule 4"}},
			{"id": "rule_5", "priority": 4, "condition": {"type": "always_true"}, "action": {"type": "percentage_discount", "value": 2, "label": "Rule 5"}},
			{"id": "rule_6", "priority": 5, "condition": {"type": "always_true"}, "action": {"type": "percentage_discount", "value": 1, "label": "Rule 6"}},
			{"id": "rule_7", "priority": 6, "condition": {"type": "always_true"}, "action": {"type": "fixed_discount", "value": 20, "label": "Rule 7"}},
			{"id": "rule_8", "priority": 7, "condition": {"type": "always_true"}, "action": {"type": "percentage_discount", "value": 5, "label": "Rule 8"}},
			{"id": "rule_9", "priority": 8, "condition": {"type": "always_true"}, "action": {"type": "percentage_discount", "value": 3, "label": "Rule 9"}},
			{"id": "rule_10", "priority": 9, "condition": {"type": "always_true"}, "action": {"type": "percentage_discount", "value": 2, "label": "Rule 10"}}
		],
		"validation_rules": [
			{"id": "val_1", "condition": {"type": "field_exists", "field": "user.name"}, "error_type": "blocking", "error_message": "Name required"},
			{"id": "val_2", "condition": {"type": "field_exists", "field": "user.email"}, "error_type": "blocking", "error_message": "Email required"},
			{"id": "val_3", "condition": {"type": "compare", "field": "user.age", "operator": ">=", "value": 18}, "error_type": "blocking", "error_message": "Age check"}
		]
	}`)

	parser := NewParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidate benchmarks rule set validation
func BenchmarkValidate(b *testing.B) {
	jsonData := []byte(`{
		"event_id": "test-event",
		"version": "1.0",
		"name": "Test Event",
		"rule_definitions": {
			"rule_1": {"type": "condition", "expression": {"type": "always_true"}},
			"rule_2": {"type": "condition", "expression": {"type": "field_exists", "field": "user.name"}}
		},
		"pricing_rules": [
			{
				"id": "price_1",
				"priority": 0,
				"condition": {"type": "rule_ref", "rule": "rule_1"},
				"action": {"type": "set_price", "value": 1000, "label": "Price"}
			}
		],
		"validation_rules": [
			{
				"id": "val_1",
				"condition": {"type": "rule_ref", "rule": "rule_2"},
				"error_type": "blocking",
				"error_message": "Validation"
			}
		]
	}`)

	parser := NewParser()
	ruleSet, _ := parser.Parse(jsonData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.Validate(ruleSet)
	}
}
