package parser

import (
	"testing"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParse_ValidMinimalRuleSet tests parsing a minimal valid RuleSet
func TestParse_ValidMinimalRuleSet(t *testing.T) {
	jsonData := []byte(`{
		"event_id": "test-event-001",
		"version": "1.0",
		"name": "Test Event",
		"pricing_rules": [],
		"validation_rules": []
	}`)

	parser := NewParser()
	ruleSet, err := parser.Parse(jsonData)

	require.NoError(t, err)
	assert.NotNil(t, ruleSet)
	assert.Equal(t, "test-event-001", ruleSet.EventID)
	assert.Equal(t, "1.0", ruleSet.Version)
	assert.Equal(t, "Test Event", ruleSet.Name)
	assert.Empty(t, ruleSet.PricingRules)
	assert.Empty(t, ruleSet.ValidationRules)
}

// TestParse_WithVariables tests parsing a RuleSet with variables
func TestParse_WithVariables(t *testing.T) {
	jsonData := []byte(`{
		"event_id": "test-event-002",
		"version": "1.0",
		"name": "Test Event with Variables",
		"variables": {
			"base_price": 1000,
			"discount_rate": 10,
			"max_participants": 100
		},
		"pricing_rules": [],
		"validation_rules": []
	}`)

	parser := NewParser()
	ruleSet, err := parser.Parse(jsonData)

	require.NoError(t, err)
	assert.NotNil(t, ruleSet)
	assert.NotNil(t, ruleSet.Variables)
	assert.Equal(t, float64(1000), ruleSet.Variables["base_price"])
	assert.Equal(t, float64(10), ruleSet.Variables["discount_rate"])
	assert.Equal(t, float64(100), ruleSet.Variables["max_participants"])
}

// TestParse_WithComputedFields tests parsing a RuleSet with computed fields
func TestParse_WithComputedFields(t *testing.T) {
	jsonData := []byte(`{
		"event_id": "test-event-003",
		"version": "1.0",
		"name": "Test Event with Computed Fields",
		"computed_fields": {
			"subtotal": {
				"type": "sum_prices",
				"description": "Subtotal of all items",
				"items": ["registration_fee", "addon:*"]
			},
			"addon_count": {
				"type": "count_items",
				"description": "Number of addons",
				"items": ["addon:*"]
			}
		},
		"pricing_rules": [],
		"validation_rules": []
	}`)

	parser := NewParser()
	ruleSet, err := parser.Parse(jsonData)

	require.NoError(t, err)
	assert.NotNil(t, ruleSet)
	assert.NotNil(t, ruleSet.ComputedFields)
	assert.Contains(t, ruleSet.ComputedFields, "subtotal")
	assert.Contains(t, ruleSet.ComputedFields, "addon_count")

	subtotal := ruleSet.ComputedFields["subtotal"]
	assert.Equal(t, "sum_prices", subtotal.Type)
	assert.Equal(t, []string{"registration_fee", "addon:*"}, subtotal.Items)
}

// TestParse_CompleteRuleSet tests parsing a complete RuleSet with all components
func TestParse_CompleteRuleSet(t *testing.T) {
	jsonData := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"variables": {
			"full_marathon_price": 1050,
			"insurance_price": 91
		},
		"computed_fields": {
			"subtotal": {
				"type": "sum_prices",
				"items": ["registration_fee", "addon:*"]
			}
		},
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Set base registration price",
				"condition": {
					"type": "always_true"
				},
				"action": {
					"type": "set_price",
					"item": "registration_fee",
					"value": 1050,
					"label": "Marathon Registration Fee"
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
	ruleSet, err := parser.Parse(jsonData)

	require.NoError(t, err)
	assert.NotNil(t, ruleSet)
	assert.Equal(t, "marathon-2025", ruleSet.EventID)
	assert.Len(t, ruleSet.PricingRules, 1)
	assert.Len(t, ruleSet.ValidationRules, 1)
	assert.NotEmpty(t, ruleSet.Variables)
	assert.NotEmpty(t, ruleSet.ComputedFields)
}

// TestParse_InvalidJSON tests parsing invalid JSON
func TestParse_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{invalid json}`)

	parser := NewParser()
	ruleSet, err := parser.Parse(jsonData)

	assert.Error(t, err)
	assert.Nil(t, ruleSet)
	assert.Contains(t, err.Error(), "failed to parse DSL")
}

// TestValidate_MissingEventID tests validation with missing event_id
func TestValidate_MissingEventID(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		Version:         "1.0",
		Name:            "Test Event",
		PricingRules:    []*dsl.PricingRule{},
		ValidationRules: []*dsl.ValidationRule{},
	}

	parser := NewParser()
	err := parser.Validate(ruleSet)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event_id is required")
}

// TestValidate_RuleRefExists tests validation of rule_ref references
func TestValidate_RuleRefExists(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		Version: "1.0",
		Name:    "Test Event",
		RuleDefs: map[string]*dsl.RuleDef{
			"is_early_bird": {
				Type:        "condition",
				Description: "Check if registration is during early bird period",
				Expression: &dsl.Expression{
					Type: "always_true",
				},
			},
		},
		PricingRules: []*dsl.PricingRule{
			{
				ID:          "early_bird_discount",
				Priority:    100,
				Description: "Early bird discount",
				Condition: &dsl.Expression{
					Type: "rule_ref",
					Rule: "is_early_bird",
				},
				Action: &dsl.Action{
					Type:  "percentage_discount",
					Value: 10,
					Label: "Early Bird 10% Off",
				},
			},
		},
		ValidationRules: []*dsl.ValidationRule{},
	}

	parser := NewParser()
	err := parser.Validate(ruleSet)

	assert.NoError(t, err)
}

// TestValidate_RuleRefNotFound tests validation error when rule_ref references non-existent rule
func TestValidate_RuleRefNotFound(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID:  "test-event",
		Version:  "1.0",
		Name:     "Test Event",
		RuleDefs: map[string]*dsl.RuleDef{},
		PricingRules: []*dsl.PricingRule{
			{
				ID:          "discount",
				Priority:    100,
				Description: "Discount rule",
				Condition: &dsl.Expression{
					Type: "rule_ref",
					Rule: "non_existent_rule",
				},
				Action: &dsl.Action{
					Type:  "percentage_discount",
					Value: 10,
					Label: "Discount",
				},
			},
		},
		ValidationRules: []*dsl.ValidationRule{},
	}

	parser := NewParser()
	err := parser.Validate(ruleSet)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non_existent_rule")
	assert.Contains(t, err.Error(), "not found")
}

// TestValidate_ExpressionTypes tests validation of various expression types
func TestValidate_ExpressionTypes(t *testing.T) {
	tests := []struct {
		name        string
		expression  *dsl.Expression
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid always_true",
			expression: &dsl.Expression{
				Type: "always_true",
			},
			shouldError: false,
		},
		{
			name: "valid equals",
			expression: &dsl.Expression{
				Type:  "equals",
				Field: "user.race_type",
				Value: "full_marathon",
			},
			shouldError: false,
		},
		{
			name: "equals missing field",
			expression: &dsl.Expression{
				Type:  "equals",
				Value: "full_marathon",
			},
			shouldError: true,
			errorMsg:    "requires field",
		},
		{
			name: "valid compare",
			expression: &dsl.Expression{
				Type:     "compare",
				Field:    "user.age",
				Operator: ">=",
				Value:    18,
			},
			shouldError: false,
		},
		{
			name: "compare missing field",
			expression: &dsl.Expression{
				Type:     "compare",
				Operator: ">=",
				Value:    18,
			},
			shouldError: true,
			errorMsg:    "requires field",
		},
		{
			name: "unknown expression type",
			expression: &dsl.Expression{
				Type: "unknown_type",
			},
			shouldError: true,
			errorMsg:    "unknown expression type",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateExpression(tt.expression, nil)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_AndOrConditions tests validation of and/or conditions
func TestValidate_AndOrConditions(t *testing.T) {
	tests := []struct {
		name        string
		expression  *dsl.Expression
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid and with multiple conditions",
			expression: &dsl.Expression{
				Type: "and",
				Conditions: []*dsl.Expression{
					{Type: "always_true"},
					{Type: "always_true"},
				},
			},
			shouldError: false,
		},
		{
			name: "and with empty conditions",
			expression: &dsl.Expression{
				Type:       "and",
				Conditions: []*dsl.Expression{},
			},
			shouldError: true,
			errorMsg:    "requires at least one condition",
		},
		{
			name: "valid or with multiple conditions",
			expression: &dsl.Expression{
				Type: "or",
				Conditions: []*dsl.Expression{
					{Type: "always_true"},
					{Type: "always_true"},
				},
			},
			shouldError: false,
		},
		{
			name: "or with empty conditions",
			expression: &dsl.Expression{
				Type:       "or",
				Conditions: []*dsl.Expression{},
			},
			shouldError: true,
			errorMsg:    "requires at least one condition",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateExpression(tt.expression, nil)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_NotCondition tests validation of not condition
func TestValidate_NotCondition(t *testing.T) {
	tests := []struct {
		name        string
		expression  *dsl.Expression
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid not with condition",
			expression: &dsl.Expression{
				Type: "not",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			shouldError: false,
		},
		{
			name: "not without condition",
			expression: &dsl.Expression{
				Type:      "not",
				Condition: nil,
			},
			shouldError: true,
			errorMsg:    "requires a condition",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateExpression(tt.expression, nil)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_ArrayConditions tests validation of array conditions
func TestValidate_ArrayConditions(t *testing.T) {
	tests := []struct {
		name        string
		expression  *dsl.Expression
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid array_any",
			expression: &dsl.Expression{
				Type:  "array_any",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			shouldError: false,
		},
		{
			name: "array_any without array field",
			expression: &dsl.Expression{
				Type: "array_any",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			shouldError: true,
			errorMsg:    "requires array field",
		},
		{
			name: "valid array_all",
			expression: &dsl.Expression{
				Type:  "array_all",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			shouldError: false,
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateExpression(tt.expression, nil)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_DatetimeBetween tests validation of datetime_between
func TestValidate_DatetimeBetween(t *testing.T) {
	tests := []struct {
		name        string
		expression  *dsl.Expression
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid datetime_between",
			expression: &dsl.Expression{
				Type:  "datetime_between",
				Field: "register_date",
				Start: "2025-01-01T00:00:00+08:00",
				End:   "2025-12-31T23:59:59+08:00",
			},
			shouldError: false,
		},
		{
			name: "datetime_between missing field",
			expression: &dsl.Expression{
				Type:  "datetime_between",
				Start: "2025-01-01T00:00:00+08:00",
				End:   "2025-12-31T23:59:59+08:00",
			},
			shouldError: true,
			errorMsg:    "requires field",
		},
		{
			name: "datetime_between missing start",
			expression: &dsl.Expression{
				Type:  "datetime_between",
				Field: "register_date",
				End:   "2025-12-31T23:59:59+08:00",
			},
			shouldError: true,
			errorMsg:    "requires start and end",
		},
		{
			name: "datetime_between missing end",
			expression: &dsl.Expression{
				Type:  "datetime_between",
				Field: "register_date",
				Start: "2025-01-01T00:00:00+08:00",
			},
			shouldError: true,
			errorMsg:    "requires start and end",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateExpression(tt.expression, nil)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_InList tests validation of in_list
func TestValidate_InList(t *testing.T) {
	tests := []struct {
		name        string
		expression  *dsl.Expression
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid in_list",
			expression: &dsl.Expression{
				Type:  "in_list",
				Field: "user.license_plate",
				List:  "$data_sources.car_owners",
			},
			shouldError: false,
		},
		{
			name: "in_list missing field",
			expression: &dsl.Expression{
				Type: "in_list",
				List: "$data_sources.car_owners",
			},
			shouldError: true,
			errorMsg:    "requires field and list",
		},
		{
			name: "in_list missing list",
			expression: &dsl.Expression{
				Type:  "in_list",
				Field: "user.license_plate",
			},
			shouldError: true,
			errorMsg:    "requires field and list",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.validateExpression(tt.expression, nil)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidate_NilExpression tests validation with nil expression
func TestValidate_NilExpression(t *testing.T) {
	parser := NewParser()
	err := parser.validateExpression(nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}
