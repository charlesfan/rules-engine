package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

func TestDSLGenerator_GenerateSimple_BasicPricing(t *testing.T) {
	generator := NewDSLGenerator(nil)

	req := GenerationRequest{
		EventID:   "marathon-2025",
		EventName: "Marathon 2025",
		Version:   "1.0",
		ConfirmedRules: []ConfirmedRule{
			{
				Type:        RuleTypePricing,
				Description: "基本報名費",
				Details: map[string]interface{}{
					"price": float64(1000),
				},
			},
		},
	}

	result, err := generator.GenerateSimple(req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Parse the generated DSL
	var dsl map[string]interface{}
	err = json.Unmarshal(result.DSL, &dsl)
	require.NoError(t, err)

	assert.Equal(t, "marathon-2025", dsl["event_id"])
	assert.Equal(t, "Marathon 2025", dsl["name"])
	assert.Equal(t, "1.0", dsl["version"])

	// Check pricing rules
	pricingRules, ok := dsl["pricing_rules"].([]interface{})
	require.True(t, ok)
	assert.Len(t, pricingRules, 1)
}

func TestDSLGenerator_GenerateSimple_WithDiscount(t *testing.T) {
	generator := NewDSLGenerator(nil)

	req := GenerationRequest{
		EventID:   "event-001",
		EventName: "Test Event",
		ConfirmedRules: []ConfirmedRule{
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
					"deadline": "2025-10-01T00:00:00Z",
				},
			},
		},
	}

	result, err := generator.GenerateSimple(req)
	require.NoError(t, err)

	var dsl map[string]interface{}
	err = json.Unmarshal(result.DSL, &dsl)
	require.NoError(t, err)

	pricingRules, ok := dsl["pricing_rules"].([]interface{})
	require.True(t, ok)
	assert.Len(t, pricingRules, 2)

	// Check the discount rule has datetime_before condition
	discountRule := pricingRules[1].(map[string]interface{})
	condition := discountRule["condition"].(map[string]interface{})
	assert.Equal(t, "datetime_before", condition["type"])
	assert.Equal(t, "register_date", condition["field"])
}

func TestDSLGenerator_GenerateSimple_MultipleDiscounts(t *testing.T) {
	generator := NewDSLGenerator(nil)

	req := GenerationRequest{
		EventID:   "event-001",
		EventName: "Test Event",
		ConfirmedRules: []ConfirmedRule{
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
			{
				Type:        RuleTypeDiscount,
				Description: "團報優惠",
				Details: map[string]interface{}{
					"discount": float64(5),
					"min_size": float64(5),
				},
			},
		},
	}

	result, err := generator.GenerateSimple(req)
	require.NoError(t, err)

	var dsl map[string]interface{}
	err = json.Unmarshal(result.DSL, &dsl)
	require.NoError(t, err)

	// Should have discount_stacking
	stacking, ok := dsl["discount_stacking"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "multiplicative", stacking["mode"])
}

func TestDSLGenerator_GenerateSimple_WithAddon(t *testing.T) {
	generator := NewDSLGenerator(nil)

	req := GenerationRequest{
		EventID:   "event-001",
		EventName: "Test Event",
		ConfirmedRules: []ConfirmedRule{
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
				Details: map[string]interface{}{
					"price": float64(350),
					"item":  "tshirt",
				},
			},
		},
	}

	result, err := generator.GenerateSimple(req)
	require.NoError(t, err)

	var dsl map[string]interface{}
	err = json.Unmarshal(result.DSL, &dsl)
	require.NoError(t, err)

	pricingRules, ok := dsl["pricing_rules"].([]interface{})
	require.True(t, ok)
	assert.Len(t, pricingRules, 2)

	// Check addon rule
	addonRule := pricingRules[1].(map[string]interface{})
	action := addonRule["action"].(map[string]interface{})
	assert.Equal(t, "add_item", action["type"])
	assert.Equal(t, "addon:tshirt", action["item"])
}

func TestDSLGenerator_GenerateSimple_Defaults(t *testing.T) {
	generator := NewDSLGenerator(nil)

	// Empty request should use defaults
	req := GenerationRequest{
		ConfirmedRules: []ConfirmedRule{
			{
				Type:        RuleTypePricing,
				Description: "報名費",
			},
		},
	}

	result, err := generator.GenerateSimple(req)
	require.NoError(t, err)

	var dsl map[string]interface{}
	err = json.Unmarshal(result.DSL, &dsl)
	require.NoError(t, err)

	assert.Equal(t, "event-001", dsl["event_id"])
	assert.Equal(t, "Event", dsl["name"])
	assert.Equal(t, "1.0", dsl["version"])
}

func TestExtractCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		lang     string
		expected string
	}{
		{
			name: "json code block",
			content: "Here is the DSL:\n```json\n{\"event_id\": \"test\"}\n```\nEnd.",
			lang:     "json",
			expected: `{"event_id": "test"}`,
		},
		{
			name: "generic code block",
			content: "Result:\n```\n{\"id\": \"123\"}\n```",
			lang:     "json",
			expected: `{"id": "123"}`,
		},
		{
			name:     "no code block",
			content:  "Just plain text",
			lang:     "json",
			expected: "",
		},
		{
			name: "multiple code blocks",
			content: "First:\n```json\n{\"first\": true}\n```\nSecond:\n```json\n{\"second\": true}\n```",
			lang:     "json",
			expected: `{"first": true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCodeBlock(tt.content, tt.lang)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractExplanation(t *testing.T) {
	content := "Here is the DSL:\n```json\n{\"id\": \"test\"}\n```\n\nThis DSL creates a basic rule."
	dslJSON := `{"id": "test"}`

	explanation := extractExplanation(content, dslJSON)
	assert.Contains(t, explanation, "This DSL creates a basic rule")
}

// mockLLMProvider is a mock LLM provider for testing
type mockLLMProvider struct {
	response string
	err      error
}

func (m *mockLLMProvider) Name() string {
	return "mock"
}

func (m *mockLLMProvider) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &llm.ChatResponse{Content: m.response}, nil
}

func (m *mockLLMProvider) Available(ctx context.Context) bool {
	return true
}

func TestDSLGenerator_Generate_WithLLM(t *testing.T) {
	mockResponse := "```json\n" + `{
		"event_id": "test-event",
		"version": "1.0",
		"name": "Test Event",
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
	}` + "\n```\n\nThis DSL sets a basic registration fee of 1000."

	generator := NewDSLGenerator(&mockLLMProvider{response: mockResponse})

	req := GenerationRequest{
		EventID:   "test-event",
		EventName: "Test Event",
		Version:   "1.0",
		ConfirmedRules: []ConfirmedRule{
			{
				Type:        RuleTypePricing,
				Description: "報名費 1000 元",
				Details: map[string]interface{}{
					"price": float64(1000),
				},
			},
		},
	}

	result, err := generator.Generate(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	var dsl map[string]interface{}
	err = json.Unmarshal(result.DSL, &dsl)
	require.NoError(t, err)

	assert.Equal(t, "test-event", dsl["event_id"])
	assert.NotEmpty(t, result.Explanation)
}

func TestDSLGenerator_Generate_NoProvider(t *testing.T) {
	generator := NewDSLGenerator(nil)

	req := GenerationRequest{
		ConfirmedRules: []ConfirmedRule{
			{
				Type:        RuleTypePricing,
				Description: "報名費",
			},
		},
	}

	_, err := generator.Generate(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no LLM provider")
}

func TestDSLGenerator_Generate_NoRules(t *testing.T) {
	generator := NewDSLGenerator(&mockLLMProvider{response: ""})

	req := GenerationRequest{
		ConfirmedRules: []ConfirmedRule{},
	}

	_, err := generator.Generate(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no confirmed rules")
}
