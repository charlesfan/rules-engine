package evaluator

import (
	"testing"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a basic context
func createBasicContext() *dsl.Context {
	return &dsl.Context{
		RegisterDate: time.Date(2025, 10, 15, 10, 0, 0, 0, time.UTC),
		User: map[string]interface{}{
			"name":      "John Doe",
			"age":       25,
			"race_type": "full_marathon",
			"email":     "john@example.com",
		},
		TeamSize: 1,
		Addons:   map[string]interface{}{},
		Variables: map[string]interface{}{
			"base_price": 1000,
		},
	}
}

// TestEvaluateExpression_AlwaysTrue tests always_true expression
func TestEvaluateExpression_AlwaysTrue(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	expr := &dsl.Expression{
		Type: "always_true",
	}

	result, err := evaluator.EvaluateExpression(expr, ctx)

	require.NoError(t, err)
	assert.True(t, result)
}

// TestEvaluateExpression_Equals tests equals expression with different value types
func TestEvaluateExpression_Equals(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    interface{}
		expected bool
	}{
		{
			name:     "string equals - match",
			field:    "user.race_type",
			value:    "full_marathon",
			expected: true,
		},
		{
			name:     "string equals - no match",
			field:    "user.race_type",
			value:    "half_marathon",
			expected: false,
		},
		{
			name:     "number equals - match (int)",
			field:    "user.age",
			value:    25,
			expected: true,
		},
		{
			name:     "number equals - match (float64)",
			field:    "user.age",
			value:    25.0,
			expected: true,
		},
		{
			name:     "number equals - no match",
			field:    "user.age",
			value:    30,
			expected: false,
		},
		{
			name:     "field not exists",
			field:    "user.non_existent",
			value:    "something",
			expected: false,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type:  "equals",
				Field: tt.field,
				Value: tt.value,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_Compare tests compare expression with different operators
func TestEvaluateExpression_Compare(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		operator string
		value    interface{}
		expected bool
	}{
		{
			name:     "greater than - true",
			field:    "user.age",
			operator: ">",
			value:    20,
			expected: true,
		},
		{
			name:     "greater than - false",
			field:    "user.age",
			operator: ">",
			value:    30,
			expected: false,
		},
		{
			name:     "less than - true",
			field:    "user.age",
			operator: "<",
			value:    30,
			expected: true,
		},
		{
			name:     "less than - false",
			field:    "user.age",
			operator: "<",
			value:    20,
			expected: false,
		},
		{
			name:     "greater than or equal - true (greater)",
			field:    "user.age",
			operator: ">=",
			value:    20,
			expected: true,
		},
		{
			name:     "greater than or equal - true (equal)",
			field:    "user.age",
			operator: ">=",
			value:    25,
			expected: true,
		},
		{
			name:     "greater than or equal - false",
			field:    "user.age",
			operator: ">=",
			value:    30,
			expected: false,
		},
		{
			name:     "less than or equal - true (less)",
			field:    "user.age",
			operator: "<=",
			value:    30,
			expected: true,
		},
		{
			name:     "less than or equal - true (equal)",
			field:    "user.age",
			operator: "<=",
			value:    25,
			expected: true,
		},
		{
			name:     "less than or equal - false",
			field:    "user.age",
			operator: "<=",
			value:    20,
			expected: false,
		},
		{
			name:     "equal - true",
			field:    "user.age",
			operator: "==",
			value:    25,
			expected: true,
		},
		{
			name:     "equal - false",
			field:    "user.age",
			operator: "==",
			value:    30,
			expected: false,
		},
		{
			name:     "not equal - true",
			field:    "user.age",
			operator: "!=",
			value:    30,
			expected: true,
		},
		{
			name:     "not equal - false",
			field:    "user.age",
			operator: "!=",
			value:    25,
			expected: false,
		},
		{
			name:     "int vs float64 comparison",
			field:    "user.age",
			operator: "==",
			value:    25.0,
			expected: true,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type:     "compare",
				Field:    tt.field,
				Operator: tt.operator,
				Value:    tt.value,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_And tests and expression
func TestEvaluateExpression_And(t *testing.T) {
	tests := []struct {
		name       string
		conditions []*dsl.Expression
		expected   bool
	}{
		{
			name: "all true",
			conditions: []*dsl.Expression{
				{Type: "always_true"},
				{Type: "always_true"},
				{Type: "always_true"},
			},
			expected: true,
		},
		{
			name: "one false",
			conditions: []*dsl.Expression{
				{Type: "always_true"},
				{Type: "equals", Field: "user.age", Value: 30}, // false
				{Type: "always_true"},
			},
			expected: false,
		},
		{
			name: "all false",
			conditions: []*dsl.Expression{
				{Type: "equals", Field: "user.age", Value: 30},
				{Type: "equals", Field: "user.age", Value: 35},
			},
			expected: false,
		},
		{
			name: "complex conditions - all true",
			conditions: []*dsl.Expression{
				{Type: "equals", Field: "user.race_type", Value: "full_marathon"},
				{Type: "compare", Field: "user.age", Operator: ">=", Value: 18},
				{Type: "compare", Field: "user.age", Operator: "<=", Value: 65},
			},
			expected: true,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type:       "and",
				Conditions: tt.conditions,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_Or tests or expression
func TestEvaluateExpression_Or(t *testing.T) {
	tests := []struct {
		name       string
		conditions []*dsl.Expression
		expected   bool
	}{
		{
			name: "at least one true",
			conditions: []*dsl.Expression{
				{Type: "equals", Field: "user.age", Value: 30}, // false
				{Type: "always_true"},                          // true
				{Type: "equals", Field: "user.age", Value: 35}, // false
			},
			expected: true,
		},
		{
			name: "all true",
			conditions: []*dsl.Expression{
				{Type: "always_true"},
				{Type: "always_true"},
			},
			expected: true,
		},
		{
			name: "all false",
			conditions: []*dsl.Expression{
				{Type: "equals", Field: "user.age", Value: 30},
				{Type: "equals", Field: "user.age", Value: 35},
			},
			expected: false,
		},
		{
			name: "complex conditions - one true",
			conditions: []*dsl.Expression{
				{Type: "equals", Field: "user.race_type", Value: "full_marathon"}, // true
				{Type: "equals", Field: "user.race_type", Value: "half_marathon"}, // false
			},
			expected: true,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type:       "or",
				Conditions: tt.conditions,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_Not tests not expression
func TestEvaluateExpression_Not(t *testing.T) {
	tests := []struct {
		name      string
		condition *dsl.Expression
		expected  bool
	}{
		{
			name: "not true = false",
			condition: &dsl.Expression{
				Type: "always_true",
			},
			expected: false,
		},
		{
			name: "not false = true",
			condition: &dsl.Expression{
				Type:  "equals",
				Field: "user.age",
				Value: 30,
			},
			expected: true,
		},
		{
			name: "double not - not (not true) = true",
			condition: &dsl.Expression{
				Type: "not",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			expected: true,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type:      "not",
				Condition: tt.condition,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_DatetimeBefore tests datetime_before expression
func TestEvaluateExpression_DatetimeBefore(t *testing.T) {
	tests := []struct {
		name         string
		registerDate time.Time
		compareDate  string
		expected     bool
	}{
		{
			name:         "before - true",
			registerDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			compareDate:  "2025-10-01T00:00:00Z",
			expected:     true,
		},
		{
			name:         "before - false (after)",
			registerDate: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			compareDate:  "2025-10-01T00:00:00Z",
			expected:     false,
		},
		{
			name:         "before - false (equal)",
			registerDate: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			compareDate:  "2025-10-01T00:00:00Z",
			expected:     false,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createBasicContext()
			ctx.RegisterDate = tt.registerDate

			expr := &dsl.Expression{
				Type:  "datetime_before",
				Field: "register_date",
				Value: tt.compareDate,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_DatetimeAfter tests datetime_after expression
func TestEvaluateExpression_DatetimeAfter(t *testing.T) {
	tests := []struct {
		name         string
		registerDate time.Time
		compareDate  string
		expected     bool
	}{
		{
			name:         "after - true",
			registerDate: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			compareDate:  "2025-10-01T00:00:00Z",
			expected:     true,
		},
		{
			name:         "after - false (before)",
			registerDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			compareDate:  "2025-10-01T00:00:00Z",
			expected:     false,
		},
		{
			name:         "after - false (equal)",
			registerDate: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			compareDate:  "2025-10-01T00:00:00Z",
			expected:     false,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createBasicContext()
			ctx.RegisterDate = tt.registerDate

			expr := &dsl.Expression{
				Type:  "datetime_after",
				Field: "register_date",
				Value: tt.compareDate,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_DatetimeBetween tests datetime_between expression
func TestEvaluateExpression_DatetimeBetween(t *testing.T) {
	tests := []struct {
		name         string
		registerDate time.Time
		startDate    string
		endDate      string
		expected     bool
	}{
		{
			name:         "within range",
			registerDate: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
			startDate:    "2025-10-01T00:00:00Z",
			endDate:      "2025-10-31T23:59:59Z",
			expected:     true,
		},
		{
			name:         "before range",
			registerDate: time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
			startDate:    "2025-10-01T00:00:00Z",
			endDate:      "2025-10-31T23:59:59Z",
			expected:     false,
		},
		{
			name:         "after range",
			registerDate: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			startDate:    "2025-10-01T00:00:00Z",
			endDate:      "2025-10-31T23:59:59Z",
			expected:     false,
		},
		{
			name:         "equal to start (inclusive)",
			registerDate: time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
			startDate:    "2025-10-01T00:00:00Z",
			endDate:      "2025-10-31T23:59:59Z",
			expected:     true,
		},
		{
			name:         "equal to end (inclusive)",
			registerDate: time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC),
			startDate:    "2025-10-01T00:00:00Z",
			endDate:      "2025-10-31T23:59:59Z",
			expected:     true,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createBasicContext()
			ctx.RegisterDate = tt.registerDate

			expr := &dsl.Expression{
				Type:  "datetime_between",
				Field: "register_date",
				Start: tt.startDate,
				End:   tt.endDate,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_FieldExists tests field_exists expression
func TestEvaluateExpression_FieldExists(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{
			name:     "field exists",
			field:    "user.name",
			expected: true,
		},
		{
			name:     "field not exists",
			field:    "user.non_existent_field",
			expected: false,
		},
		{
			name:     "nested field exists",
			field:    "user.email",
			expected: true,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type:  "field_exists",
				Field: tt.field,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_FieldEmpty tests field_empty expression
func TestEvaluateExpression_FieldEmpty(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name":         "John",
			"empty_string": "",
			"nil_value":    nil,
			"zero_number":  0,
		},
		Addons: map[string]interface{}{
			"empty_array":  []interface{}{},
			"filled_array": []interface{}{1, 2, 3},
		},
	}

	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{
			name:     "non-empty string",
			field:    "user.name",
			expected: false,
		},
		{
			name:     "empty string",
			field:    "user.empty_string",
			expected: true,
		},
		{
			name:     "nil value",
			field:    "user.nil_value",
			expected: true,
		},
		{
			name:     "field not exists",
			field:    "user.non_existent",
			expected: true,
		},
		{
			name:     "zero number is not empty",
			field:    "user.zero_number",
			expected: false,
		},
		{
			name:     "empty array",
			field:    "addons.empty_array",
			expected: true,
		},
		{
			name:     "filled array",
			field:    "addons.filled_array",
			expected: false,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type:  "field_empty",
				Field: tt.field,
			}

			result, err := evaluator.EvaluateExpression(expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetFieldValue tests field value retrieval with nested paths
func TestGetFieldValue(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
		User: map[string]interface{}{
			"name": "John",
			"profile": map[string]interface{}{
				"age":  25,
				"city": "Taipei",
			},
		},
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{"name": "Alice", "age": 30},
			},
		},
		TeamSize: 5,
		Addons: map[string]interface{}{
			"insurance": "plan_a",
		},
		ComputedValues: map[string]interface{}{
			"subtotal": 1500.0,
		},
	}

	tests := []struct {
		name          string
		field         string
		expectedValue interface{}
		shouldError   bool
	}{
		{
			name:          "get user field",
			field:         "user.name",
			expectedValue: "John",
			shouldError:   false,
		},
		{
			name:          "get nested user field",
			field:         "user.profile.age",
			expectedValue: 25,
			shouldError:   false,
		},
		{
			name:          "get team_size",
			field:         "team_size",
			expectedValue: 5,
			shouldError:   false,
		},
		{
			name:          "get register_date",
			field:         "register_date",
			expectedValue: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
			shouldError:   false,
		},
		{
			name:          "get addon field",
			field:         "addons.insurance",
			expectedValue: "plan_a",
			shouldError:   false,
		},
		{
			name:          "get computed field",
			field:         "$computed.subtotal",
			expectedValue: 1500.0,
			shouldError:   false,
		},
		{
			name:        "non-existent field",
			field:       "user.non_existent",
			shouldError: true,
		},
		{
			name:        "unknown root field",
			field:       "unknown.field",
			shouldError: true,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.GetFieldValue(tt.field, ctx)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

// TestEvaluate_ValidationRules tests full evaluation with validation rules
func TestEvaluate_ValidationRules(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		Version: "1.0",
		Name:    "Test Event",
		ValidationRules: []*dsl.ValidationRule{
			{
				ID:          "age_limit",
				Description: "Must be at least 18 years old",
				Condition: &dsl.Expression{
					Type:     "compare",
					Field:    "user.age",
					Operator: "<",
					Value:    18,
				},
				ErrorType:    "blocking",
				ErrorMessage: "Participants must be at least 18 years old",
			},
			{
				ID:          "health_warning",
				Description: "Health warning",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
				ErrorType:    "warning",
				ErrorMessage: "Please consult your doctor before participating",
			},
		},
		PricingRules: []*dsl.PricingRule{},
	}

	tests := []struct {
		name             string
		age              int
		expectedValid    bool
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name:             "age >= 18 (valid)",
			age:              25,
			expectedValid:    true,
			expectedErrors:   0,
			expectedWarnings: 1, // health warning always shows
		},
		{
			name:             "age < 18 (invalid)",
			age:              15,
			expectedValid:    false,
			expectedErrors:   1,
			expectedWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createBasicContext()
			ctx.User["age"] = tt.age

			evaluator := NewEvaluator(ruleSet)
			result, err := evaluator.Evaluate(ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedValid, result.Valid)
			assert.Len(t, result.Errors, tt.expectedErrors)
			assert.Len(t, result.Warnings, tt.expectedWarnings)

			if tt.expectedErrors > 0 {
				assert.Equal(t, "age_limit", result.Errors[0].RuleID)
				assert.Equal(t, "blocking", result.Errors[0].Type)
			}

			if tt.expectedWarnings > 0 {
				assert.Equal(t, "health_warning", result.Warnings[0].RuleID)
				assert.Equal(t, "warning", result.Warnings[0].Type)
			}
		})
	}
}

// TestEvaluateExpression_NilExpression tests error handling for nil expression
func TestEvaluateExpression_NilExpression(t *testing.T) {
	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)
	ctx := createBasicContext()

	result, err := evaluator.EvaluateExpression(nil, ctx)

	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "expression is nil")
}

// TestCompareValues_TypeHandling tests type handling in value comparison
func TestCompareValues_TypeHandling(t *testing.T) {
	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		operator string
		expected bool
	}{
		{
			name:     "int vs int",
			a:        25,
			b:        25,
			operator: "==",
			expected: true,
		},
		{
			name:     "int vs float64",
			a:        25,
			b:        25.0,
			operator: "==",
			expected: true,
		},
		{
			name:     "float64 vs int",
			a:        25.0,
			b:        25,
			operator: "==",
			expected: true,
		},
		{
			name:     "float64 vs float64",
			a:        25.5,
			b:        25.5,
			operator: "==",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.compareValues(tt.a, tt.b, tt.operator)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_ArrayAny tests array_any expression in detail
func TestEvaluateExpression_ArrayAny(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{
					"name":   "Alice",
					"gender": "female",
					"age":    25,
				},
				map[string]interface{}{
					"name":   "Bob",
					"gender": "male",
					"age":    30,
				},
				map[string]interface{}{
					"name":   "Charlie",
					"gender": "male",
					"age":    28,
				},
			},
		},
	}

	tests := []struct {
		name      string
		expr      *dsl.Expression
		expected  bool
		shouldErr bool
	}{
		{
			name: "any member is female - true",
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:  "equals",
					Field: "user.gender",
					Value: "female",
				},
			},
			expected: true,
		},
		{
			name: "any member age > 35 - false",
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:     "compare",
					Field:    "user.age",
					Operator: ">",
					Value:    35,
				},
			},
			expected: false,
		},
		{
			name: "any member age >= 30 - true",
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:     "compare",
					Field:    "user.age",
					Operator: ">=",
					Value:    30,
				},
			},
			expected: true,
		},
		{
			name: "complex condition - any male over 25",
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type: "and",
					Conditions: []*dsl.Expression{
						{
							Type:  "equals",
							Field: "user.gender",
							Value: "male",
						},
						{
							Type:     "compare",
							Field:    "user.age",
							Operator: ">",
							Value:    25,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "empty array - false",
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "team.empty_array",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			expected: false,
		},
		{
			name: "non-existent array - false",
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "team.non_existent",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			expected: false,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	// Add empty array to context
	ctx.Team["empty_array"] = []interface{}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateExpression(tt.expr, ctx)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_ArrayAll tests array_all expression in detail
func TestEvaluateExpression_ArrayAll(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{
					"name":   "Alice",
					"gender": "female",
					"age":    25,
					"paid":   true,
				},
				map[string]interface{}{
					"name":   "Bob",
					"gender": "male",
					"age":    30,
					"paid":   true,
				},
				map[string]interface{}{
					"name":   "Charlie",
					"gender": "male",
					"age":    28,
					"paid":   true,
				},
			},
		},
	}

	tests := []struct {
		name      string
		expr      *dsl.Expression
		expected  bool
		shouldErr bool
	}{
		{
			name: "all members paid - true",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:  "equals",
					Field: "user.paid",
					Value: true,
				},
			},
			expected: true,
		},
		{
			name: "all members are male - false",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:  "equals",
					Field: "user.gender",
					Value: "male",
				},
			},
			expected: false,
		},
		{
			name: "all members age >= 25 - true",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:     "compare",
					Field:    "user.age",
					Operator: ">=",
					Value:    25,
				},
			},
			expected: true,
		},
		{
			name: "all members age >= 30 - false",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:     "compare",
					Field:    "user.age",
					Operator: ">=",
					Value:    30,
				},
			},
			expected: false,
		},
		{
			name: "complex condition - all members paid AND age > 20",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type: "and",
					Conditions: []*dsl.Expression{
						{
							Type:  "equals",
							Field: "user.paid",
							Value: true,
						},
						{
							Type:     "compare",
							Field:    "user.age",
							Operator: ">",
							Value:    20,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "empty array - true (vacuous truth)",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "team.empty_array",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			expected: true,
		},
		{
			name: "non-existent array - false",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "team.non_existent",
				Condition: &dsl.Expression{
					Type: "always_true",
				},
			},
			expected: false,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	// Add empty array to context
	ctx.Team["empty_array"] = []interface{}{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateExpression(tt.expr, ctx)

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_InList tests in_list expression
func TestEvaluateExpression_InList(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		DataSources: map[string]*dsl.DataSource{
			"approved_users": {
				Type:        "external_list",
				Description: "Approved users list",
				Fields:      []string{"user_id", "name"},
			},
			"car_owners": {
				Type:        "external_list",
				Description: "Car owners list",
				Fields:      []string{"name", "plate_number"},
			},
		},
	}

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"user_id":       "user123",
			"license_plate": "ABC-1234",
		},
		DataSources: map[string]interface{}{
			"approved_users": []interface{}{
				map[string]interface{}{
					"user_id": "user123",
					"name":    "John",
				},
				map[string]interface{}{
					"user_id": "user456",
					"name":    "Jane",
				},
			},
			"car_owners": []interface{}{
				map[string]interface{}{
					"name":         "John",
					"plate_number": "ABC-1234",
				},
				map[string]interface{}{
					"name":         "Jane",
					"plate_number": "XYZ-5678",
				},
			},
		},
	}

	tests := []struct {
		name     string
		expr     *dsl.Expression
		expected bool
	}{
		{
			name: "user in approved list - true",
			expr: &dsl.Expression{
				Type:       "in_list",
				Field:      "user.user_id",
				List:       "$data_sources.approved_users",
				MatchField: "user_id",
			},
			expected: true,
		},
		{
			name: "license plate in car owners - true",
			expr: &dsl.Expression{
				Type:       "in_list",
				Field:      "user.license_plate",
				List:       "$data_sources.car_owners",
				MatchField: "plate_number",
			},
			expected: true,
		},
		{
			name: "non-existent user_id - false",
			expr: &dsl.Expression{
				Type:       "in_list",
				Field:      "user.user_id",
				List:       "$data_sources.approved_users",
				MatchField: "user_id",
			},
			expected: false,
		},
	}

	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Modify user_id for the third test case
			if tt.name == "non-existent user_id - false" {
				ctx.User["user_id"] = "user999"
			} else {
				ctx.User["user_id"] = "user123"
			}

			result, err := evaluator.EvaluateExpression(tt.expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_RuleRef tests rule_ref expression
func TestEvaluateExpression_RuleRef(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		RuleDefs: map[string]*dsl.RuleDef{
			"is_adult": {
				Type:        "condition",
				Description: "Check if user is an adult",
				Expression: &dsl.Expression{
					Type:     "compare",
					Field:    "user.age",
					Operator: ">=",
					Value:    18,
				},
			},
			"is_early_bird": {
				Type:        "condition",
				Description: "Check if registered during early bird period",
				Expression: &dsl.Expression{
					Type:  "datetime_before",
					Field: "register_date",
					Value: "2025-10-01T00:00:00Z",
				},
			},
			"is_vip": {
				Type:        "condition",
				Description: "Check if user is VIP",
				Expression: &dsl.Expression{
					Type:  "equals",
					Field: "user.vip_status",
					Value: true,
				},
			},
		},
	}

	tests := []struct {
		name        string
		ctx         *dsl.Context
		ruleRef     string
		expected    bool
		shouldError bool
	}{
		{
			name: "is_adult - true",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User: map[string]interface{}{
					"age": 25,
				},
			},
			ruleRef:  "is_adult",
			expected: true,
		},
		{
			name: "is_adult - false",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User: map[string]interface{}{
					"age": 15,
				},
			},
			ruleRef:  "is_adult",
			expected: false,
		},
		{
			name: "is_early_bird - true",
			ctx: &dsl.Context{
				RegisterDate: time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
				User:         map[string]interface{}{},
			},
			ruleRef:  "is_early_bird",
			expected: true,
		},
		{
			name: "is_early_bird - false",
			ctx: &dsl.Context{
				RegisterDate: time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC),
				User:         map[string]interface{}{},
			},
			ruleRef:  "is_early_bird",
			expected: false,
		},
		{
			name: "is_vip - true",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User: map[string]interface{}{
					"vip_status": true,
				},
			},
			ruleRef:  "is_vip",
			expected: true,
		},
		{
			name: "non-existent rule",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User:         map[string]interface{}{},
			},
			ruleRef:     "non_existent_rule",
			shouldError: true,
		},
	}

	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &dsl.Expression{
				Type: "rule_ref",
				Rule: tt.ruleRef,
			}

			result, err := evaluator.EvaluateExpression(expr, tt.ctx)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestEvaluateExpression_NestedArrayConditions tests nested array operations
func TestEvaluateExpression_NestedArrayConditions(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{
					"name":   "Alice",
					"gender": "female",
					"age":    25,
				},
				map[string]interface{}{
					"name":   "Bob",
					"gender": "male",
					"age":    30,
				},
				map[string]interface{}{
					"name":   "Charlie",
					"gender": "male",
					"age":    17,
				},
			},
		},
	}

	tests := []struct {
		name     string
		expr     *dsl.Expression
		expected bool
	}{
		{
			name: "has female OR has underage member",
			expr: &dsl.Expression{
				Type: "or",
				Conditions: []*dsl.Expression{
					{
						Type:  "array_any",
						Array: "team.members",
						Condition: &dsl.Expression{
							Type:  "equals",
							Field: "user.gender",
							Value: "female",
						},
					},
					{
						Type:  "array_any",
						Array: "team.members",
						Condition: &dsl.Expression{
							Type:     "compare",
							Field:    "user.age",
							Operator: "<",
							Value:    18,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "all adults AND has at least one female",
			expr: &dsl.Expression{
				Type: "and",
				Conditions: []*dsl.Expression{
					{
						Type:  "array_all",
						Array: "team.members",
						Condition: &dsl.Expression{
							Type:     "compare",
							Field:    "user.age",
							Operator: ">=",
							Value:    18,
						},
					},
					{
						Type:  "array_any",
						Array: "team.members",
						Condition: &dsl.Expression{
							Type:  "equals",
							Field: "user.gender",
							Value: "female",
						},
					},
				},
			},
			expected: false, // Charlie is 17
		},
		{
			name: "NOT all members are male",
			expr: &dsl.Expression{
				Type: "not",
				Condition: &dsl.Expression{
					Type:  "array_all",
					Array: "team.members",
					Condition: &dsl.Expression{
						Type:  "equals",
						Field: "user.gender",
						Value: "male",
					},
				},
			},
			expected: true, // Alice is female
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateExpression(tt.expr, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluateExpression_ErrorHandling tests error handling for various edge cases
func TestEvaluateExpression_ErrorHandling(t *testing.T) {
	now := time.Now()
	ctx := &dsl.Context{
		RegisterDate: now,
		User: map[string]interface{}{
			"user_id": "user1",
			"age":     25,
			"email":   "test@example.com",
		},
		DataSources: map[string]interface{}{
			"approved_users": []map[string]interface{}{
				{"user_id": "user1"},
			},
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	tests := []struct {
		name        string
		expr        *dsl.Expression
		shouldError bool
		errorMsg    string
	}{
		{
			name: "unknown expression type",
			expr: &dsl.Expression{
				Type: "unknown_type",
			},
			shouldError: true,
			errorMsg:    "unknown expression type",
		},
		{
			name: "datetime_before with invalid format",
			expr: &dsl.Expression{
				Type:  "datetime_before",
				Field: "register_date",
				Value: "invalid-date-format",
			},
			shouldError: true,
		},
		{
			name: "datetime_after with invalid format",
			expr: &dsl.Expression{
				Type:  "datetime_after",
				Field: "register_date",
				Value: "not-a-date",
			},
			shouldError: true,
		},
		{
			name: "datetime_between with invalid start",
			expr: &dsl.Expression{
				Type:  "datetime_between",
				Field: "register_date",
				Start: "bad-start",
				End:   now.Add(24 * time.Hour).Format(time.RFC3339),
			},
			shouldError: true,
		},
		{
			name: "datetime_between with invalid end",
			expr: &dsl.Expression{
				Type:  "datetime_between",
				Field: "register_date",
				Start: now.Add(-24 * time.Hour).Format(time.RFC3339),
				End:   "bad-end",
			},
			shouldError: true,
		},
		{
			name: "datetime_before with non-time field",
			expr: &dsl.Expression{
				Type:  "datetime_before",
				Field: "user.email",
				Value: now.Format(time.RFC3339),
			},
			shouldError: true,
		},
		{
			name: "array_any with non-array field",
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "user.email",
				Condition: &dsl.Expression{
					Type:  "always_true",
					Value: true,
				},
			},
			shouldError: true,
		},
		{
			name: "array_all with non-array field",
			expr: &dsl.Expression{
				Type:  "array_all",
				Array: "user.age",
				Condition: &dsl.Expression{
					Type:  "always_true",
					Value: true,
				},
			},
			shouldError: true,
		},
		{
			name: "in_list with invalid data source",
			expr: &dsl.Expression{
				Type:       "in_list",
				Field:      "user.user_id",
				List:       "$data_sources.non_existent_list",
				MatchField: "user_id",
			},
			shouldError: true,
		},
		{
			name: "compare with invalid operator",
			expr: &dsl.Expression{
				Type:     "compare",
				Field:    "user.age",
				Operator: "invalid_op",
				Value:    18,
			},
			shouldError: true,
			errorMsg:    "unknown operator",
		},
		{
			name: "compare non-numeric values",
			expr: &dsl.Expression{
				Type:     "compare",
				Field:    "user.email",
				Operator: ">",
				Value:    18,
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateExpression(tt.expr, ctx)
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.False(t, result)
			}
		})
	}
}

// TestEvaluateExpression_NilHandling tests nil value handling
func TestEvaluateExpression_NilHandling(t *testing.T) {
	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	tests := []struct {
		name     string
		ctx      *dsl.Context
		expr     *dsl.Expression
		expected bool
	}{
		{
			name: "nil user map - field_exists returns false",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User:         nil,
			},
			expr: &dsl.Expression{
				Type:  "field_exists",
				Field: "user.name",
			},
			expected: false,
		},
		{
			name: "nil field value - equals returns false",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User: map[string]interface{}{
					"name": nil,
				},
			},
			expr: &dsl.Expression{
				Type:  "equals",
				Field: "user.name",
				Value: "John",
			},
			expected: false,
		},
		{
			name: "nil team - array operations return false",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User: map[string]interface{}{
					"user_id": "user1",
				},
				Team: nil,
			},
			expr: &dsl.Expression{
				Type:  "array_any",
				Array: "team.members",
				Condition: &dsl.Expression{
					Type:  "always_true",
					Value: true,
				},
			},
			expected: false,
		},
		{
			name: "empty user map - field_exists returns false",
			ctx: &dsl.Context{
				RegisterDate: time.Now(),
				User:         map[string]interface{}{},
			},
			expr: &dsl.Expression{
				Type:  "field_exists",
				Field: "user.name",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateExpression(tt.expr, tt.ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEvaluate_EmptyRules tests evaluation with empty rules
func TestEvaluate_EmptyRules(t *testing.T) {
	tests := []struct {
		name       string
		ruleSet    *dsl.RuleSet
		shouldPass bool
	}{
		{
			name: "empty validation rules",
			ruleSet: &dsl.RuleSet{
				EventID:         "test-event",
				ValidationRules: []*dsl.ValidationRule{},
				PricingRules:    []*dsl.PricingRule{},
			},
			shouldPass: true,
		},
		{
			name: "nil validation rules",
			ruleSet: &dsl.RuleSet{
				EventID:         "test-event",
				ValidationRules: nil,
				PricingRules:    nil,
			},
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := NewEvaluator(tt.ruleSet)
			ctx := &dsl.Context{
				RegisterDate: time.Now(),
				User: map[string]interface{}{
					"user_id": "user1",
				},
			}

			result, err := evaluator.Evaluate(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.shouldPass, result.Valid)
		})
	}
}

// TestCompareValues_EdgeCases tests edge cases in value comparison
func TestCompareValues_EdgeCases(t *testing.T) {
	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	tests := []struct {
		name        string
		a           interface{}
		b           interface{}
		operator    string
		expected    bool
		shouldError bool
	}{
		{
			name:     "zero values equal",
			a:        0,
			b:        0,
			operator: "==",
			expected: true,
		},
		{
			name:     "negative numbers",
			a:        -10,
			b:        -5,
			operator: "<",
			expected: true,
		},
		{
			name:     "float comparison",
			a:        10.5,
			b:        10.5,
			operator: "==",
			expected: true,
		},
		{
			name:     "mixed int and float",
			a:        10,
			b:        10.0,
			operator: "==",
			expected: true,
		},
		{
			name:        "string comparison fails",
			a:           "10",
			b:           10,
			operator:    ">",
			shouldError: true,
		},
		{
			name:        "bool comparison fails",
			a:           true,
			b:           1,
			operator:    ">",
			shouldError: true,
		},
		{
			name:     "large numbers",
			a:        1000000,
			b:        999999,
			operator: ">",
			expected: true,
		},
		{
			name:        "invalid operator",
			a:           10,
			b:           5,
			operator:    "invalid_op",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.compareValues(tt.a, tt.b, tt.operator)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestToFloat64_InvalidTypes tests type conversion
func TestToFloat64_InvalidTypes(t *testing.T) {
	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	tests := []struct {
		name        string
		value       interface{}
		expectOk    bool
		expectedVal float64
	}{
		{
			name:     "string value",
			value:    "not a number",
			expectOk: false,
		},
		{
			name:     "bool value",
			value:    true,
			expectOk: false,
		},
		{
			name:     "nil value",
			value:    nil,
			expectOk: false,
		},
		{
			name:     "map value",
			value:    map[string]interface{}{"key": "value"},
			expectOk: false,
		},
		{
			name:     "slice value",
			value:    []int{1, 2, 3},
			expectOk: false,
		},
		{
			name:        "valid int",
			value:       42,
			expectOk:    true,
			expectedVal: 42.0,
		},
		{
			name:        "valid float",
			value:       3.14,
			expectOk:    true,
			expectedVal: 3.14,
		},
		{
			name:        "valid int64",
			value:       int64(100),
			expectOk:    true,
			expectedVal: 100.0,
		},
		{
			name:        "valid float32",
			value:       float32(2.5),
			expectOk:    true,
			expectedVal: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := evaluator.toFloat64(tt.value)
			assert.Equal(t, tt.expectOk, ok)
			if tt.expectOk {
				assert.Equal(t, tt.expectedVal, result)
			}
		})
	}
}

// TestEvaluateExpression_ComplexNesting tests deeply nested conditions
func TestEvaluateExpression_ComplexNesting(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"age":    25,
			"gender": "female",
			"paid":   true,
		},
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{
					"user": map[string]interface{}{
						"age":    20,
						"gender": "female",
						"paid":   true,
					},
				},
				map[string]interface{}{
					"user": map[string]interface{}{
						"age":    30,
						"gender": "male",
						"paid":   true,
					},
				},
			},
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	// ((age >= 18) AND (gender == "female")) OR (paid == true)
	expr := &dsl.Expression{
		Type: "or",
		Conditions: []*dsl.Expression{
			{
				Type: "and",
				Conditions: []*dsl.Expression{
					{
						Type:     "compare",
						Field:    "user.age",
						Operator: ">=",
						Value:    18,
					},
					{
						Type:  "equals",
						Field: "user.gender",
						Value: "female",
					},
				},
			},
			{
				Type:  "equals",
				Field: "user.paid",
				Value: true,
			},
		},
	}

	result, err := evaluator.EvaluateExpression(expr, ctx)
	require.NoError(t, err)
	assert.True(t, result)
}

// TestEvaluateExpression_FieldPathEdgeCases tests edge cases in field path resolution
func TestEvaluateExpression_FieldPathEdgeCases(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"profile": map[string]interface{}{
				"contact": map[string]interface{}{
					"email": "test@example.com",
				},
			},
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	tests := []struct {
		name     string
		expr     *dsl.Expression
		expected bool
	}{
		{
			name: "deeply nested field exists",
			expr: &dsl.Expression{
				Type:  "field_exists",
				Field: "user.profile.contact.email",
			},
			expected: true,
		},
		{
			name: "partial path exists but final field missing",
			expr: &dsl.Expression{
				Type:  "field_exists",
				Field: "user.profile.contact.phone",
			},
			expected: false,
		},
		{
			name: "middle path missing",
			expr: &dsl.Expression{
				Type:  "field_exists",
				Field: "user.profile.address.city",
			},
			expected: false,
		},
		{
			name: "deeply nested equals",
			expr: &dsl.Expression{
				Type:  "equals",
				Field: "user.profile.contact.email",
				Value: "test@example.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateExpression(tt.expr, ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetFieldValue_EdgeCases tests edge cases in field value retrieval
func TestGetFieldValue_EdgeCases(t *testing.T) {
	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name":  "John",
			"empty": "",
			"zero":  0,
			"false": false,
			"null":  nil,
		},
	}

	ruleSet := &dsl.RuleSet{EventID: "test-event"}
	evaluator := NewEvaluator(ruleSet)

	tests := []struct {
		name        string
		field       string
		expected    interface{}
		shouldError bool
	}{
		{
			name:     "empty string value",
			field:    "user.empty",
			expected: "",
		},
		{
			name:     "zero numeric value",
			field:    "user.zero",
			expected: 0,
		},
		{
			name:     "false boolean value",
			field:    "user.false",
			expected: false,
		},
		{
			name:     "explicit nil value",
			field:    "user.null",
			expected: nil,
		},
		{
			name:        "non-existent field",
			field:       "user.nonexistent",
			shouldError: true,
		},
		{
			name:     "register_date field",
			field:    "register_date",
			expected: ctx.RegisterDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.getFieldValue(tt.field, ctx)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
