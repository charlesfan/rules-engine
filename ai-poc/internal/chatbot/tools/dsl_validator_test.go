package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDSLValidator_Validate_ValidDSL(t *testing.T) {
	validator := NewDSLValidator()

	validDSL := []byte(`{
		"event_id": "marathon-2025",
		"version": "1.0",
		"name": "Marathon 2025",
		"variables": {
			"base_price": 1000
		},
		"pricing_rules": [
			{
				"id": "base_price",
				"priority": 0,
				"description": "Basic registration fee",
				"condition": {"type": "always_true"},
				"action": {
					"type": "set_price",
					"item": "registration_fee",
					"value": "$variables.base_price",
					"label": "Registration Fee"
				}
			}
		],
		"validation_rules": []
	}`)

	result := validator.Validate(validDSL)

	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
	require.NotNil(t, result.RuleSet)
	assert.Equal(t, "marathon-2025", result.RuleSet.EventID)
}

func TestDSLValidator_Validate_InvalidJSON(t *testing.T) {
	validator := NewDSLValidator()

	invalidJSON := []byte(`{invalid json}`)

	result := validator.Validate(invalidJSON)

	assert.False(t, result.Valid)
	require.NotEmpty(t, result.Errors)
	assert.Equal(t, "syntax", result.Errors[0].Type)
}

func TestDSLValidator_Validate_MissingEventID(t *testing.T) {
	validator := NewDSLValidator()

	dsl := []byte(`{
		"version": "1.0",
		"name": "Test Event",
		"pricing_rules": [],
		"validation_rules": []
	}`)

	result := validator.Validate(dsl)

	assert.False(t, result.Valid)
	// Should have error about missing event_id
	hasEventIDError := false
	for _, err := range result.Errors {
		if err.Field == "event_id" || err.Message == "event_id is required" {
			hasEventIDError = true
			break
		}
	}
	assert.True(t, hasEventIDError)
}

func TestDSLValidator_Validate_DuplicateRuleID(t *testing.T) {
	validator := NewDSLValidator()

	dsl := []byte(`{
		"event_id": "test",
		"pricing_rules": [
			{
				"id": "rule1",
				"priority": 0,
				"description": "First rule",
				"condition": {"type": "always_true"},
				"action": {"type": "set_price", "value": 100}
			},
			{
				"id": "rule1",
				"priority": 10,
				"description": "Duplicate ID",
				"condition": {"type": "always_true"},
				"action": {"type": "set_price", "value": 200}
			}
		],
		"validation_rules": []
	}`)

	result := validator.Validate(dsl)

	assert.False(t, result.Valid)
	// Should have duplicate ID error
	hasDuplicateError := false
	for _, err := range result.Errors {
		if err.Type == "semantic" && err.Message == "Duplicate rule ID: rule1" {
			hasDuplicateError = true
			break
		}
	}
	assert.True(t, hasDuplicateError)
}

func TestDSLValidator_Validate_UndefinedVariable(t *testing.T) {
	validator := NewDSLValidator()

	dsl := []byte(`{
		"event_id": "test",
		"variables": {},
		"pricing_rules": [
			{
				"id": "rule1",
				"priority": 0,
				"description": "Uses undefined variable",
				"condition": {"type": "always_true"},
				"action": {
					"type": "set_price",
					"value": "$variables.undefined_var"
				}
			}
		],
		"validation_rules": []
	}`)

	result := validator.Validate(dsl)

	assert.False(t, result.Valid)
	// Should have reference error
	hasReferenceError := false
	for _, err := range result.Errors {
		if err.Type == "reference" {
			hasReferenceError = true
			break
		}
	}
	assert.True(t, hasReferenceError)
}

func TestDSLValidator_Validate_MultipleDiscountsWarning(t *testing.T) {
	validator := NewDSLValidator()

	dsl := []byte(`{
		"event_id": "test",
		"pricing_rules": [
			{
				"id": "discount1",
				"priority": 100,
				"description": "Early bird",
				"condition": {"type": "always_true"},
				"action": {"type": "percentage_discount", "value": 10}
			},
			{
				"id": "discount2",
				"priority": 110,
				"description": "Team discount",
				"condition": {"type": "always_true"},
				"action": {"type": "percentage_discount", "value": 5}
			}
		],
		"validation_rules": []
	}`)

	result := validator.Validate(dsl)

	// Should have warning about missing discount_stacking
	hasStackingWarning := false
	for _, warn := range result.Warnings {
		if warn == "Multiple discount rules found but no discount_stacking mode specified. Defaulting to 'multiplicative'." {
			hasStackingWarning = true
			break
		}
	}
	assert.True(t, hasStackingWarning)
}

func TestDSLValidator_Validate_PriorityConflictWarning(t *testing.T) {
	validator := NewDSLValidator()

	dsl := []byte(`{
		"event_id": "test",
		"pricing_rules": [
			{
				"id": "rule1",
				"priority": 0,
				"description": "First rule",
				"condition": {"type": "always_true"},
				"action": {"type": "set_price", "value": 100}
			},
			{
				"id": "rule2",
				"priority": 0,
				"description": "Same priority",
				"condition": {"type": "always_true"},
				"action": {"type": "set_price", "value": 200}
			}
		],
		"validation_rules": []
	}`)

	result := validator.Validate(dsl)

	// Should have warning about same priority
	hasPriorityWarning := false
	for _, warn := range result.Warnings {
		if len(warn) > 0 && warn[0:8] == "Multiple" {
			hasPriorityWarning = true
			break
		}
	}
	assert.True(t, hasPriorityWarning)
}

func TestDSLValidator_Validate_Summary(t *testing.T) {
	validator := NewDSLValidator()

	dsl := []byte(`{
		"event_id": "test",
		"variables": {
			"price": 1000,
			"discount": 10
		},
		"pricing_rules": [
			{
				"id": "base",
				"priority": 0,
				"description": "Base price",
				"condition": {"type": "always_true"},
				"action": {"type": "set_price", "value": 1000}
			},
			{
				"id": "early_bird",
				"priority": 100,
				"description": "Early bird discount",
				"condition": {"type": "datetime_before", "field": "register_date", "value": "2025-10-01"},
				"action": {"type": "percentage_discount", "value": 10}
			}
		],
		"validation_rules": [
			{
				"id": "age_check",
				"description": "Age verification",
				"condition": {"type": "compare", "field": "user.age", "operator": "lt", "value": 18},
				"error_type": "blocking",
				"error_message": "Must be 18 or older"
			}
		]
	}`)

	result := validator.Validate(dsl)

	require.NotNil(t, result.Summary)
	assert.Equal(t, 2, result.Summary.PricingRuleCount)
	assert.Equal(t, 1, result.Summary.ValidationRuleCount)
	assert.Equal(t, 2, result.Summary.VariableCount)
	assert.Contains(t, result.Summary.UsedActionTypes, "set_price")
	assert.Contains(t, result.Summary.UsedActionTypes, "percentage_discount")
	assert.Contains(t, result.Summary.UsedConditionTypes, "always_true")
	assert.Contains(t, result.Summary.UsedConditionTypes, "datetime_before")
	assert.Contains(t, result.Summary.UsedConditionTypes, "compare")
}

func TestValidationResult_FormatErrors(t *testing.T) {
	result := &ValidationResult{
		Errors: []ValidationError{
			{Field: "event_id", Message: "Required field missing", Type: "semantic"},
			{Field: "pricing_rules[0]", Message: "Invalid action", Type: "syntax"},
		},
	}

	formatted := result.FormatErrors()
	assert.Contains(t, formatted, "驗證錯誤")
	assert.Contains(t, formatted, "event_id")
	assert.Contains(t, formatted, "Required field missing")
}

func TestValidationResult_FormatWarnings(t *testing.T) {
	result := &ValidationResult{
		Warnings: []string{
			"Multiple discounts without stacking",
			"Same priority rules",
		},
	}

	formatted := result.FormatWarnings()
	assert.Contains(t, formatted, "警告")
	assert.Contains(t, formatted, "Multiple discounts")
}

func TestValidationResult_FormatSummary(t *testing.T) {
	result := &ValidationResult{
		Summary: &ValidationSummary{
			PricingRuleCount:    3,
			ValidationRuleCount: 1,
			VariableCount:       2,
			UsedActionTypes:     []string{"set_price", "percentage_discount"},
			UsedConditionTypes:  []string{"always_true", "datetime_before"},
		},
	}

	formatted := result.FormatSummary()
	assert.Contains(t, formatted, "DSL 摘要")
	assert.Contains(t, formatted, "3 條")
	assert.Contains(t, formatted, "定價規則")
}

func TestDSLValidator_ValidateString(t *testing.T) {
	validator := NewDSLValidator()

	dsl := `{
		"event_id": "test",
		"pricing_rules": [
			{
				"id": "rule1",
				"priority": 0,
				"description": "Test",
				"condition": {"type": "always_true"},
				"action": {"type": "set_price", "value": 100}
			}
		],
		"validation_rules": []
	}`

	result := validator.ValidateString(dsl)

	assert.True(t, result.Valid)
}
