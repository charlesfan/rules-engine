package computed

import (
	"testing"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestBreakdown creates a test PriceBreakdown with sample items
func createTestBreakdown() *dsl.PriceBreakdown {
	return &dsl.PriceBreakdown{
		Items: map[string]*dsl.PriceItem{
			"registration_fee": {
				ID:              "registration_fee",
				OriginalPrice:   1000,
				DiscountedPrice: 900,
			},
			"addon:insurance": {
				ID:              "addon:insurance",
				OriginalPrice:   200,
				DiscountedPrice: 180,
			},
			"addon:medal": {
				ID:              "addon:medal",
				OriginalPrice:   300,
				DiscountedPrice: 270,
			},
			"addon:tshirt": {
				ID:              "addon:tshirt",
				OriginalPrice:   500,
				DiscountedPrice: 450,
			},
		},
	}
}

// TestSumPrices tests sum_prices computation
func TestSumPrices(t *testing.T) {
	tests := []struct {
		name        string
		items       []string
		breakdown   *dsl.PriceBreakdown
		expectedSum float64
	}{
		{
			name:        "sum specific items",
			items:       []string{"registration_fee", "addon:insurance"},
			breakdown:   createTestBreakdown(),
			expectedSum: 1080, // 900 + 180
		},
		{
			name:        "sum with wildcard addon:*",
			items:       []string{"addon:*"},
			breakdown:   createTestBreakdown(),
			expectedSum: 900, // 180 + 270 + 450
		},
		{
			name:        "sum all items",
			items:       []string{"registration_fee", "addon:*"},
			breakdown:   createTestBreakdown(),
			expectedSum: 1800, // 900 + 180 + 270 + 450
		},
		{
			name:        "sum non-existent items",
			items:       []string{"non_existent"},
			breakdown:   createTestBreakdown(),
			expectedSum: 0,
		},
		{
			name:        "nil breakdown",
			items:       []string{"registration_fee"},
			breakdown:   nil,
			expectedSum: 0,
		},
		{
			name:  "empty items in breakdown",
			items: []string{"registration_fee"},
			breakdown: &dsl.PriceBreakdown{
				Items: map[string]*dsl.PriceItem{},
			},
			expectedSum: 0,
		},
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum, err := evaluator.sumPrices(tt.items, tt.breakdown)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedSum, sum)
		})
	}
}

// TestCountItems tests count_items computation
func TestCountItems(t *testing.T) {
	tests := []struct {
		name          string
		items         []string
		breakdown     *dsl.PriceBreakdown
		expectedCount int
	}{
		{
			name:          "count specific items",
			items:         []string{"registration_fee", "addon:insurance"},
			breakdown:     createTestBreakdown(),
			expectedCount: 2,
		},
		{
			name:          "count with wildcard addon:*",
			items:         []string{"addon:*"},
			breakdown:     createTestBreakdown(),
			expectedCount: 3, // insurance, medal, tshirt
		},
		{
			name:          "count all items",
			items:         []string{"registration_fee", "addon:*"},
			breakdown:     createTestBreakdown(),
			expectedCount: 4,
		},
		{
			name:          "count non-existent items",
			items:         []string{"non_existent"},
			breakdown:     createTestBreakdown(),
			expectedCount: 0,
		},
		{
			name:          "nil breakdown",
			items:         []string{"registration_fee"},
			breakdown:     nil,
			expectedCount: 0,
		},
		{
			name:  "empty items in breakdown",
			items: []string{"registration_fee"},
			breakdown: &dsl.PriceBreakdown{
				Items: map[string]*dsl.PriceItem{},
			},
			expectedCount: 0,
		},
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := evaluator.countItems(tt.items, tt.breakdown)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

// TestItemPrice tests item_price computation
func TestItemPrice(t *testing.T) {
	tests := []struct {
		name          string
		itemID        string
		breakdown     *dsl.PriceBreakdown
		expectedPrice float64
	}{
		{
			name:          "get existing item price",
			itemID:        "registration_fee",
			breakdown:     createTestBreakdown(),
			expectedPrice: 900,
		},
		{
			name:          "get addon item price",
			itemID:        "addon:insurance",
			breakdown:     createTestBreakdown(),
			expectedPrice: 180,
		},
		{
			name:          "get non-existent item",
			itemID:        "non_existent",
			breakdown:     createTestBreakdown(),
			expectedPrice: 0,
		},
		{
			name:          "nil breakdown",
			itemID:        "registration_fee",
			breakdown:     nil,
			expectedPrice: 0,
		},
		{
			name:   "empty items in breakdown",
			itemID: "registration_fee",
			breakdown: &dsl.PriceBreakdown{
				Items: map[string]*dsl.PriceItem{},
			},
			expectedPrice: 0,
		},
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := evaluator.itemPrice(tt.itemID, tt.breakdown)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPrice, price)
		})
	}
}

// TestCountArrayField tests count_array_field computation
func TestCountArrayField(t *testing.T) {
	ctx := &dsl.Context{
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
		name          string
		field         *dsl.ComputedField
		expectedCount int
	}{
		{
			name: "count by specific value",
			field: &dsl.ComputedField{
				Type:  "count_array_field",
				Array: "team.members",
				Field: "gender",
				Value: "male",
			},
			expectedCount: 2,
		},
		{
			name: "count by specific value (single match)",
			field: &dsl.ComputedField{
				Type:  "count_array_field",
				Array: "team.members",
				Field: "gender",
				Value: "female",
			},
			expectedCount: 1,
		},
		{
			name: "count by field existence (no value specified)",
			field: &dsl.ComputedField{
				Type:  "count_array_field",
				Array: "team.members",
				Field: "name",
				Value: nil,
			},
			expectedCount: 3, // all members have name
		},
		{
			name: "count non-existent value",
			field: &dsl.ComputedField{
				Type:  "count_array_field",
				Array: "team.members",
				Field: "gender",
				Value: "other",
			},
			expectedCount: 0,
		},
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := evaluator.countArrayField(tt.field, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

// TestCountArrayField_NestedPath tests count_array_field with nested paths
func TestCountArrayField_NestedPath(t *testing.T) {
	ctx := &dsl.Context{
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{
					"name": "Alice",
					"addons": map[string]interface{}{
						"water_bottle": map[string]interface{}{
							"quantity": 2,
						},
					},
				},
				map[string]interface{}{
					"name": "Bob",
					"addons": map[string]interface{}{
						"water_bottle": map[string]interface{}{
							"quantity": 1,
						},
					},
				},
				map[string]interface{}{
					"name":   "Charlie",
					"addons": map[string]interface{}{},
				},
			},
		},
	}

	field := &dsl.ComputedField{
		Type:  "count_array_field",
		Array: "team.members",
		Field: "addons.water_bottle.quantity",
		Value: nil, // count members who have this field
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	count, err := evaluator.countArrayField(field, ctx)

	require.NoError(t, err)
	assert.Equal(t, 2, count) // Alice and Bob have water_bottle quantity
}

// TestSumArrayField tests sum_array_field computation
func TestSumArrayField(t *testing.T) {
	ctx := &dsl.Context{
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{
					"name": "Alice",
					"age":  25,
					"addons": map[string]interface{}{
						"water_bottle": map[string]interface{}{
							"quantity": 2,
						},
					},
				},
				map[string]interface{}{
					"name": "Bob",
					"age":  30,
					"addons": map[string]interface{}{
						"water_bottle": map[string]interface{}{
							"quantity": 3,
						},
					},
				},
				map[string]interface{}{
					"name": "Charlie",
					"age":  28,
				},
			},
		},
	}

	tests := []struct {
		name        string
		field       *dsl.ComputedField
		expectedSum float64
	}{
		{
			name: "sum simple numeric field",
			field: &dsl.ComputedField{
				Type:  "sum_array_field",
				Array: "team.members",
				Field: "age",
			},
			expectedSum: 83, // 25 + 30 + 28
		},
		{
			name: "sum nested field",
			field: &dsl.ComputedField{
				Type:  "sum_array_field",
				Array: "team.members",
				Field: "addons.water_bottle.quantity",
			},
			expectedSum: 5, // 2 + 3 (Charlie doesn't have it)
		},
		{
			name: "sum non-existent field",
			field: &dsl.ComputedField{
				Type:  "sum_array_field",
				Array: "team.members",
				Field: "non_existent",
			},
			expectedSum: 0,
		},
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum, err := evaluator.sumArrayField(tt.field, ctx)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedSum, sum)
		})
	}
}

// TestSumArrayField_DifferentTypes tests sum_array_field with different numeric types
func TestSumArrayField_DifferentTypes(t *testing.T) {
	ctx := &dsl.Context{
		Team: map[string]interface{}{
			"members": []interface{}{
				map[string]interface{}{"value": int(10)},
				map[string]interface{}{"value": int64(20)},
				map[string]interface{}{"value": float32(15.5)},
				map[string]interface{}{"value": float64(30.5)},
			},
		},
	}

	field := &dsl.ComputedField{
		Type:  "sum_array_field",
		Array: "team.members",
		Field: "value",
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	sum, err := evaluator.sumArrayField(field, ctx)

	require.NoError(t, err)
	assert.InDelta(t, 76.0, sum, 0.01) // 10 + 20 + 15.5 + 30.5
}

// TestGetNestedValue tests nested value extraction from maps
func TestGetNestedValue(t *testing.T) {
	m := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "deep_value",
			},
			"number": 42,
		},
		"simple": "value",
	}

	tests := []struct {
		name          string
		path          string
		expectedValue interface{}
	}{
		{
			name:          "simple path",
			path:          "simple",
			expectedValue: "value",
		},
		{
			name:          "nested path",
			path:          "level1.level2.level3",
			expectedValue: "deep_value",
		},
		{
			name:          "nested number",
			path:          "level1.number",
			expectedValue: 42,
		},
		{
			name:          "non-existent path",
			path:          "non.existent",
			expectedValue: nil,
		},
		{
			name:          "partial path",
			path:          "level1.non_existent",
			expectedValue: nil,
		},
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := evaluator.getNestedValue(m, tt.path)
			assert.Equal(t, tt.expectedValue, value)
		})
	}
}

// TestEvaluate tests full evaluation of all computed fields
func TestEvaluate(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
		ComputedFields: map[string]*dsl.ComputedField{
			"subtotal": {
				Type:        "sum_prices",
				Description: "Subtotal of all items",
				Items:       []string{"registration_fee", "addon:*"},
			},
			"addon_count": {
				Type:        "count_items",
				Description: "Number of addons",
				Items:       []string{"addon:*"},
			},
			"registration_price": {
				Type:        "item_price",
				Description: "Registration fee price",
				Item:        "registration_fee",
			},
		},
	}

	ctx := &dsl.Context{
		ComputedValues: nil, // will be initialized by Evaluate
	}

	breakdown := createTestBreakdown()

	evaluator := NewFieldEvaluator(ruleSet)
	err := evaluator.Evaluate(ctx, breakdown)

	require.NoError(t, err)
	assert.NotNil(t, ctx.ComputedValues)

	// Check computed values
	assert.Equal(t, 1800.0, ctx.ComputedValues["subtotal"])
	assert.Equal(t, 3, ctx.ComputedValues["addon_count"])
	assert.Equal(t, 900.0, ctx.ComputedValues["registration_price"])
}

// TestEvaluate_NilComputedFields tests evaluation when no computed fields are defined
func TestEvaluate_NilComputedFields(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID:        "test-event",
		ComputedFields: nil,
	}

	ctx := &dsl.Context{}
	breakdown := createTestBreakdown()

	evaluator := NewFieldEvaluator(ruleSet)
	err := evaluator.Evaluate(ctx, breakdown)

	require.NoError(t, err)
}

// TestEvaluateField_UnknownType tests error handling for unknown field types
func TestEvaluateField_UnknownType(t *testing.T) {
	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}

	field := &dsl.ComputedField{
		Type: "unknown_type",
	}

	ctx := &dsl.Context{}
	breakdown := createTestBreakdown()

	evaluator := NewFieldEvaluator(ruleSet)
	_, err := evaluator.evaluateField(field, ctx, breakdown)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown computed field type")
}

// TestGetFieldValue tests field value retrieval from context
func TestGetFieldValue(t *testing.T) {
	ctx := &dsl.Context{
		User: map[string]interface{}{
			"name": "John",
			"profile": map[string]interface{}{
				"age": 25,
			},
		},
		Team: map[string]interface{}{
			"name": "Team A",
		},
		Addons: map[string]interface{}{
			"insurance": "plan_a",
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
			name:          "get team field",
			field:         "team.name",
			expectedValue: "Team A",
			shouldError:   false,
		},
		{
			name:          "get addon field",
			field:         "addons.insurance",
			expectedValue: "plan_a",
			shouldError:   false,
		},
		{
			name:        "unknown root field",
			field:       "unknown.field",
			shouldError: true,
		},
		{
			name:          "non-existent nested field returns nil",
			field:         "user.non_existent",
			expectedValue: nil,
			shouldError:   false, // getFieldValue returns nil for non-existent fields, not error
		},
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := evaluator.getFieldValue(tt.field, ctx)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

// TestCountArrayField_NonExistentArray tests error handling when array doesn't exist
func TestCountArrayField_NonExistentArray(t *testing.T) {
	ctx := &dsl.Context{
		Team: map[string]interface{}{},
	}

	field := &dsl.ComputedField{
		Type:  "count_array_field",
		Array: "team.non_existent",
		Field: "name",
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	count, err := evaluator.countArrayField(field, ctx)

	// When field doesn't exist, getFieldValue returns (nil, nil)
	// Then type assertion fails and returns error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not an array")
	assert.Equal(t, 0, count)
}

// TestSumArrayField_NonExistentArray tests error handling when array doesn't exist
func TestSumArrayField_NonExistentArray(t *testing.T) {
	ctx := &dsl.Context{
		Team: map[string]interface{}{},
	}

	field := &dsl.ComputedField{
		Type:  "sum_array_field",
		Array: "team.non_existent",
		Field: "value",
	}

	ruleSet := &dsl.RuleSet{
		EventID: "test-event",
	}
	evaluator := NewFieldEvaluator(ruleSet)

	sum, err := evaluator.sumArrayField(field, ctx)

	// When field doesn't exist, getFieldValue returns (nil, nil)
	// Then type assertion fails and returns error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not an array")
	assert.Equal(t, 0.0, sum)
}
