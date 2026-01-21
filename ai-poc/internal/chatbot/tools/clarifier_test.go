package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClarifier_Analyze_MissingBasePrice(t *testing.T) {
	clarifier := NewClarifier(nil)

	// Empty rules - should ask for base price
	rules := []ExtractedRuleInfo{}

	result, err := clarifier.Analyze(context.Background(), "", rules)
	require.NoError(t, err)

	assert.False(t, result.CanProceed)
	assert.NotEmpty(t, result.Questions)

	// Should have high priority question about base price
	hasBasePriceQuestion := false
	for _, q := range result.Questions {
		if q.ID == "missing_base_price" && q.Priority == PriorityHigh {
			hasBasePriceQuestion = true
			break
		}
	}
	assert.True(t, hasBasePriceQuestion)
}

func TestClarifier_Analyze_CompletePricingRule(t *testing.T) {
	clarifier := NewClarifier(nil)

	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Complete:    true,
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "報名費 1000 元", rules)
	require.NoError(t, err)

	// Should be able to proceed (though may have low priority suggestions)
	highPriorityCount := 0
	for _, q := range result.Questions {
		if q.Priority == PriorityHigh {
			highPriorityCount++
		}
	}
	assert.Equal(t, 0, highPriorityCount)
}

func TestClarifier_Analyze_EarlyBirdMissingDeadline(t *testing.T) {
	clarifier := NewClarifier(nil)

	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
		{
			Type:        RuleTypeDiscount,
			Description: "早鳥優惠",
			Details: map[string]interface{}{
				"discount": float64(10),
			},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "報名費 1000 元，早鳥優惠 9 折", rules)
	require.NoError(t, err)

	// Should ask for early bird deadline
	hasDeadlineQuestion := false
	for _, q := range result.Questions {
		if q.ID == "early_bird_deadline" {
			hasDeadlineQuestion = true
			assert.Equal(t, PriorityHigh, q.Priority)
			break
		}
	}
	assert.True(t, hasDeadlineQuestion)
}

func TestClarifier_Analyze_TeamDiscountMissingInfo(t *testing.T) {
	clarifier := NewClarifier(nil)

	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
		{
			Type:        RuleTypeDiscount,
			Description: "團報優惠",
			Details:     map[string]interface{}{},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "報名費 1000 元，有團報優惠", rules)
	require.NoError(t, err)

	// Should ask for team min size and discount
	hasMinSizeQuestion := false
	hasDiscountQuestion := false
	for _, q := range result.Questions {
		if q.ID == "team_min_size" {
			hasMinSizeQuestion = true
		}
		if q.ID == "team_discount" {
			hasDiscountQuestion = true
		}
	}
	assert.True(t, hasMinSizeQuestion)
	assert.True(t, hasDiscountQuestion)
}

func TestClarifier_Analyze_MultipleDiscounts(t *testing.T) {
	clarifier := NewClarifier(nil)

	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
		{
			Type:        RuleTypeDiscount,
			Description: "早鳥優惠",
			Details: map[string]interface{}{
				"discount": float64(10),
				"deadline": "2025-10-01",
			},
		},
		{
			Type:        RuleTypeDiscount,
			Description: "團報優惠",
			Details: map[string]interface{}{
				"discount": float64(5),
				"min_size": float64(5),
			},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "", rules)
	require.NoError(t, err)

	// Should ask about discount stacking
	hasStackingQuestion := false
	for _, q := range result.Questions {
		if q.ID == "discount_stacking" {
			hasStackingQuestion = true
			assert.Equal(t, PriorityMedium, q.Priority)
			assert.NotEmpty(t, q.Options)
			break
		}
	}
	assert.True(t, hasStackingQuestion)
}

func TestClarifier_Analyze_SuggestDiscount(t *testing.T) {
	clarifier := NewClarifier(nil)

	// Only pricing, no discount
	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "", rules)
	require.NoError(t, err)

	// Should suggest discount (low priority)
	hasSuggestDiscount := false
	for _, q := range result.Questions {
		if q.ID == "suggest_discount" {
			hasSuggestDiscount = true
			assert.Equal(t, PriorityLow, q.Priority)
			break
		}
	}
	assert.True(t, hasSuggestDiscount)
}

func TestClarifier_Analyze_ValidationRuleAgeLimit(t *testing.T) {
	clarifier := NewClarifier(nil)

	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
		{
			Type:        RuleTypeValidation,
			Description: "年齡限制",
			Details:     map[string]interface{}{},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "需要年齡限制", rules)
	require.NoError(t, err)

	// Should ask for min age
	hasMinAgeQuestion := false
	for _, q := range result.Questions {
		if q.ID == "validation_min_age" {
			hasMinAgeQuestion = true
			break
		}
	}
	assert.True(t, hasMinAgeQuestion)
}

func TestClarifier_Analyze_AddonMissingPrice(t *testing.T) {
	clarifier := NewClarifier(nil)

	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
		{
			Type:        RuleTypeAddon,
			Description: "紀念衫",
			Details:     map[string]interface{}{},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "有賣紀念衫", rules)
	require.NoError(t, err)

	// Should ask for addon price
	hasAddonPriceQuestion := false
	for _, q := range result.Questions {
		if q.Affects == "addon" {
			hasAddonPriceQuestion = true
			break
		}
	}
	assert.True(t, hasAddonPriceQuestion)
}

func TestClarificationResult_FormatQuestions(t *testing.T) {
	result := &ClarificationResult{
		Questions: []ClarificationQuestion{
			{
				ID:                "q1",
				Priority:          PriorityHigh,
				Question:          "報名費是多少？",
				DefaultSuggestion: "1000 元",
				Affects:           "pricing",
			},
			{
				ID:       "q2",
				Priority: PriorityMedium,
				Question: "是否有早鳥優惠？",
				Affects:  "discount",
			},
			{
				ID:       "q3",
				Priority: PriorityLow,
				Question: "是否有加購項目？",
				Affects:  "addon",
			},
		},
	}

	formatted := result.FormatQuestions()

	assert.Contains(t, formatted, "需要確認的問題")
	assert.Contains(t, formatted, "必須確認")
	assert.Contains(t, formatted, "建議確認")
	assert.Contains(t, formatted, "可選確認")
	assert.Contains(t, formatted, "報名費是多少")
	assert.Contains(t, formatted, "1000 元")
}

func TestClarificationResult_FormatQuestions_Empty(t *testing.T) {
	result := &ClarificationResult{
		Questions: []ClarificationQuestion{},
	}

	formatted := result.FormatQuestions()
	assert.Empty(t, formatted)
}

func TestClarificationResult_GetHighPriorityQuestions(t *testing.T) {
	result := &ClarificationResult{
		Questions: []ClarificationQuestion{
			{ID: "q1", Priority: PriorityHigh},
			{ID: "q2", Priority: PriorityMedium},
			{ID: "q3", Priority: PriorityHigh},
			{ID: "q4", Priority: PriorityLow},
		},
	}

	highPriority := result.GetHighPriorityQuestions()
	assert.Len(t, highPriority, 2)
	assert.Equal(t, "q1", highPriority[0].ID)
	assert.Equal(t, "q3", highPriority[1].ID)
}

func TestClarifier_Analyze_WithLLM(t *testing.T) {
	mockResponse := `{
		"questions": [
			{
				"priority": "medium",
				"question": "是否有其他折扣優惠？",
				"default_suggestion": "無",
				"affects": "discount"
			}
		],
		"can_proceed": true,
		"reason": "基本資訊完整"
	}`

	clarifier := NewClarifier(&mockLLMProvider{response: mockResponse})

	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "報名費 1000 元", rules)
	require.NoError(t, err)

	// Should have merged questions from pattern + LLM
	assert.NotEmpty(t, result.Questions)
}

func TestClarifier_CanProceedLogic(t *testing.T) {
	clarifier := NewClarifier(nil)

	// Only low/medium priority questions - can proceed
	rules := []ExtractedRuleInfo{
		{
			Type:        RuleTypePricing,
			Description: "報名費",
			Details: map[string]interface{}{
				"price": float64(1000),
			},
		},
	}

	result, err := clarifier.Analyze(context.Background(), "", rules)
	require.NoError(t, err)

	// With base price, should be able to proceed
	assert.True(t, result.CanProceed)

	// Without base price - cannot proceed
	rules2 := []ExtractedRuleInfo{}

	result2, err := clarifier.Analyze(context.Background(), "", rules2)
	require.NoError(t, err)

	assert.False(t, result2.CanProceed)
}
