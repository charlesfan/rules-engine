package evaluator

import (
	"testing"
	"time"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
)

// BenchmarkEvaluateExpression_Simple benchmarks simple expression evaluation
func BenchmarkEvaluateExpression_Simple(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewEvaluator(ruleSet)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"age":       25,
			"race_type": "full_marathon",
		},
	}

	expr := &dsl.Expression{
		Type:     "compare",
		Field:    "user.age",
		Operator: ">=",
		Value:    18,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluateExpression(expr, ctx)
	}
}

// BenchmarkEvaluateExpression_Complex benchmarks complex nested expression
func BenchmarkEvaluateExpression_Complex(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewEvaluator(ruleSet)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"age":       25,
			"race_type": "full_marathon",
			"paid":      true,
		},
	}

	expr := &dsl.Expression{
		Type: "and",
		Conditions: []*dsl.Expression{
			{
				Type:     "compare",
				Field:    "user.age",
				Operator: ">=",
				Value:    18,
			},
			{
				Type: "or",
				Conditions: []*dsl.Expression{
					{
						Type:  "equals",
						Field: "user.race_type",
						Value: "full_marathon",
					},
					{
						Type:  "equals",
						Field: "user.race_type",
						Value: "half_marathon",
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluateExpression(expr, ctx)
	}
}

// BenchmarkEvaluateExpression_ArrayAny benchmarks array_any operation
func BenchmarkEvaluateExpression_ArrayAny(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewEvaluator(ruleSet)

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{"gender": "female", "age": 25},
				map[string]interface{}{"gender": "male", "age": 30},
				map[string]interface{}{"gender": "male", "age": 28},
			},
		},
	}

	expr := &dsl.Expression{
		Type:  "array_any",
		Array: "team.members",
		Condition: &dsl.Expression{
			Type:  "equals",
			Field: "user.gender",
			Value: "female",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluateExpression(expr, ctx)
	}
}

// BenchmarkEvaluateExpression_DatetimeCompare benchmarks datetime comparison
func BenchmarkEvaluateExpression_DatetimeCompare(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewEvaluator(ruleSet)

	ctx := &dsl.Context{
		RegisterDate: time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		User:         map[string]interface{}{},
	}

	expr := &dsl.Expression{
		Type:  "datetime_before",
		Field: "register_date",
		Value: "2025-10-01T00:00:00Z",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.EvaluateExpression(expr, ctx)
	}
}

// BenchmarkEvaluate_WithValidation benchmarks full evaluation with validation rules
func BenchmarkEvaluate_WithValidation(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		ValidationRules: []*dsl.ValidationRule{
			{
				ID:          "age_limit",
				Description: "Age check",
				Condition: &dsl.Expression{
					Type:     "compare",
					Field:    "user.age",
					Operator: "<",
					Value:    18,
				},
				ErrorType:    "blocking",
				ErrorMessage: "Must be 18+",
			},
			{
				ID:          "name_required",
				Description: "Name check",
				Condition: &dsl.Expression{
					Type:  "field_empty",
					Field: "user.name",
				},
				ErrorType:    "blocking",
				ErrorMessage: "Name required",
			},
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
					Value: 1000,
					Label: "Registration Fee",
				},
			},
		},
	}

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"name": "John Doe",
			"age":  25,
		},
	}

	evaluator := NewEvaluator(ruleSet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(ctx)
	}
}

// BenchmarkEvaluate_ManyValidationRules benchmarks evaluation with many validation rules
func BenchmarkEvaluate_ManyValidationRules(b *testing.B) {
	validationRules := make([]*dsl.ValidationRule, 20)
	for i := 0; i < 20; i++ {
		validationRules[i] = &dsl.ValidationRule{
			ID:          "rule_" + string(rune(i)),
			Description: "Test rule",
			Condition: &dsl.Expression{
				Type:     "compare",
				Field:    "user.age",
				Operator: ">=",
				Value:    i,
			},
			ErrorType:    "blocking",
			ErrorMessage: "Test error",
		}
	}

	ruleSet := &dsl.RuleSet{
		EventID:         "test-event",
		ValidationRules: validationRules,
		PricingRules:    []*dsl.PricingRule{},
	}

	ctx := &dsl.Context{
		RegisterDate: time.Now(),
		User: map[string]interface{}{
			"age": 25,
		},
	}

	evaluator := NewEvaluator(ruleSet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.Evaluate(ctx)
	}
}

// BenchmarkGetFieldValue benchmarks field value retrieval
func BenchmarkGetFieldValue(b *testing.B) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewEvaluator(ruleSet)

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = evaluator.GetFieldValue("user.profile.contact.email", ctx)
	}
}
