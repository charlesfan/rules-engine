package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
	"github.com/charlesfan/rules-engine/internal/rules/dsl"
)

// mockLLMProvider implements llm.Provider for testing
type mockLLMProvider struct {
	response string
	err      error
}

func (m *mockLLMProvider) Chat(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &llm.ChatResponse{
		Content:    m.response,
		TokensUsed: 100,
	}, nil
}

func (m *mockLLMProvider) Name() string {
	return "mock"
}

func (m *mockLLMProvider) Available(ctx context.Context) bool {
	return true
}

func TestNewAgent(t *testing.T) {
	agent := NewAgent(nil)

	assert.NotNil(t, agent)
	assert.Equal(t, StateIdle, agent.state)
	assert.Empty(t, agent.ruleSet.PricingRules)
	assert.Empty(t, agent.ruleSet.ValidationRules)
	assert.Equal(t, "event-001", agent.eventInfo.EventID)
}

func TestAgent_Process_GeneralChat_NoLLM(t *testing.T) {
	agent := NewAgent(nil)

	resp, err := agent.Process(context.Background(), "你好", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "賽事規則")
	assert.Equal(t, "idle", resp.State)
	assert.Equal(t, 0, resp.RuleCount)
}

func TestAgent_Process_RuleCreate_NoLLM(t *testing.T) {
	agent := NewAgent(nil)

	resp, err := agent.Process(context.Background(), "報名費 1000 元", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "已理解")
	assert.GreaterOrEqual(t, resp.RuleCount, 1)
}

func TestAgent_Process_DSLRequest_NoRules(t *testing.T) {
	agent := NewAgent(nil)

	resp, err := agent.Process(context.Background(), "生成 DSL", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "沒有規則")
	assert.Equal(t, 0, resp.RuleCount)
}

func TestAgent_Process_DSLRequest_WithRules(t *testing.T) {
	agent := NewAgent(nil)

	// Add a rule directly using dsl types
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules, &dsl.PricingRule{
		ID:          "pricing_1",
		Priority:    0,
		Description: "報名費",
		Condition:   &dsl.Expression{Type: "always_true"},
		Action: &dsl.Action{
			Type:  "set_price",
			Item:  "registration_fee",
			Value: 1000,
			Label: "報名費",
		},
	})
	agent.state = StateReady

	resp, err := agent.Process(context.Background(), "生成 DSL", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "DSL")
	assert.NotNil(t, resp.DSL)
}

func TestAgent_SetEventInfo(t *testing.T) {
	agent := NewAgent(nil)

	agent.SetEventInfo(EventInfo{
		EventID:   "marathon-2025",
		EventName: "台北馬拉松",
		Version:   "2.0",
	})

	assert.Equal(t, "marathon-2025", agent.eventInfo.EventID)
	assert.Equal(t, "台北馬拉松", agent.eventInfo.EventName)
	assert.Equal(t, "2.0", agent.eventInfo.Version)
	// Also verify RuleSet is updated
	assert.Equal(t, "marathon-2025", agent.ruleSet.EventID)
	assert.Equal(t, "2.0", agent.ruleSet.Version)
}

func TestAgent_Clear(t *testing.T) {
	agent := NewAgent(nil)

	// Add some state
	agent.state = StateReady
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules, &dsl.PricingRule{
		ID:          "test_1",
		Description: "test",
		Condition:   &dsl.Expression{Type: "always_true"},
	})
	agent.lastDSL = []byte(`{"test": true}`)

	// Clear
	agent.Clear()

	assert.Equal(t, StateIdle, agent.state)
	assert.Empty(t, agent.ruleSet.PricingRules)
	assert.Empty(t, agent.ruleSet.ValidationRules)
	assert.Nil(t, agent.lastDSL)
}

func TestAgent_GetRulesSummary_Empty(t *testing.T) {
	agent := NewAgent(nil)

	summary := agent.GetRulesSummary()
	assert.Contains(t, summary, "沒有已建立的規則")
}

func TestAgent_GetRulesSummary_WithRules(t *testing.T) {
	agent := NewAgent(nil)
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules,
		&dsl.PricingRule{ID: "pricing_1", Description: "報名費", Condition: &dsl.Expression{Type: "always_true"}},
		&dsl.PricingRule{ID: "discount_1", Description: "早鳥優惠", Condition: &dsl.Expression{Type: "always_true"}},
	)

	summary := agent.GetRulesSummary()

	assert.Contains(t, summary, "2 條規則")
	assert.Contains(t, summary, "報名費")
}

func TestAgentState_String(t *testing.T) {
	tests := []struct {
		state    AgentState
		expected string
	}{
		{StateIdle, "idle"},
		{StateCollecting, "collecting"},
		{StateClarifying, "clarifying"},
		{StateReady, "ready"},
		{AgentState(99), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.state.String())
	}
}

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"報名費 1000 元", 1000},
		{"價格是2500", 2500},
		{"沒有數字", 0},
		{"100塊錢", 100},
	}

	for _, tt := range tests {
		result := extractNumber(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestExtractDiscount(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"9折", 10},
		{"95折", 5},
		{"八折", 0}, // Chinese number not supported
		{"10%", 10},
		{"15%優惠", 15},
		{"沒有折扣", 0},
	}

	for _, tt := range tests {
		result := extractDiscount(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestExtractTeamSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"5人", 5},
		{"10 人以上", 10},
		{"團報三人", 0}, // Chinese number not supported
		{"沒有人數", 0},
	}

	for _, tt := range tests {
		result := extractTeamSize(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestAgent_SimpleExtractRules(t *testing.T) {
	agent := NewAgent(nil)

	tests := []struct {
		input         string
		expectedCount int
	}{
		{"報名費 1000 元", 1},
		{"早鳥優惠 9 折", 1},
		{"報名費 1000 元，早鳥 9 折", 2},
		{"團報 5 人 95 折", 1},
		{"年齡限制 18 歲", 1},
		{"可加購紀念衫", 1},
	}

	for _, tt := range tests {
		rules := agent.simpleExtractRules(tt.input)
		assert.Len(t, rules, tt.expectedCount, "input: %s", tt.input)
	}
}

func TestAgent_Process_WithMockLLM(t *testing.T) {
	// Test with mock LLM returning proper JSON response
	mockResp := `{
		"intent": "rule_input",
		"rules": [
			{
				"id": "new_pricing_21k",
				"action": "add",
				"rule_type": "pricing",
				"data": {
					"priority": 0,
					"description": "21K 報名費",
					"condition": {"type": "equals", "field": "user.race_type", "value": "21K"},
					"action": {"type": "set_price", "item": "registration_fee", "value": 1080, "label": "21K 報名費"}
				}
			}
		],
		"variables": {
			"price_21k": {"action": "add", "value": 1080}
		},
		"questions": [],
		"can_generate": true,
		"message": "已設定 21K 報名費 NT$1,080。"
	}`

	mockProvider := &mockLLMProvider{response: mockResp}
	agent := NewAgent(mockProvider)

	resp, err := agent.Process(context.Background(), "21K 報名費 1080 元", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "1,080")
	assert.Equal(t, 1, resp.RuleCount)
	assert.True(t, resp.CanGenerate)
}

func TestAgent_Process_ModifyRule_WithMockLLM(t *testing.T) {
	// Test rule modification with mock LLM
	mockResp := `{
		"intent": "modify_rule",
		"rules": [
			{
				"id": "pricing_1",
				"action": "update",
				"rule_type": "pricing",
				"data": {
					"priority": 0,
					"description": "21K 報名費",
					"condition": {"type": "equals", "field": "user.race_type", "value": "21K"},
					"action": {"type": "set_price", "item": "registration_fee", "value": 1300, "label": "21K 報名費"}
				}
			}
		],
		"questions": [],
		"can_generate": true,
		"message": "已將 21K 報名費改為 NT$1,300。"
	}`

	mockProvider := &mockLLMProvider{response: mockResp}
	agent := NewAgent(mockProvider)

	// Pre-add a rule
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules, &dsl.PricingRule{
		ID:          "pricing_1",
		Priority:    0,
		Description: "21K 報名費",
		Condition:   &dsl.Expression{Type: "equals", Field: "user.race_type", Value: "21K"},
		Action:      &dsl.Action{Type: "set_price", Item: "registration_fee", Value: 1080},
	})

	resp, err := agent.Process(context.Background(), "把半馬改成 1300 元", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "1,300")
	assert.Equal(t, 1, resp.RuleCount)
}

func TestAgent_Process_GeneralChat_WithMockLLM(t *testing.T) {
	mockResp := `{
		"intent": "general_chat",
		"rules": [],
		"questions": [],
		"can_generate": false,
		"message": "你好！我是賽事報名規則助手。請描述您的報名規則。"
	}`

	mockProvider := &mockLLMProvider{response: mockResp}
	agent := NewAgent(mockProvider)

	resp, err := agent.Process(context.Background(), "你好", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "賽事報名規則助手")
	assert.Equal(t, 0, resp.RuleCount)
	assert.False(t, resp.CanGenerate)
}

func TestAgent_Process_DSLRequest_WithMockLLM(t *testing.T) {
	mockResp := `{
		"intent": "dsl_request",
		"rules": [],
		"questions": [],
		"can_generate": true,
		"message": "好的，正在生成 DSL..."
	}`

	mockProvider := &mockLLMProvider{response: mockResp}
	agent := NewAgent(mockProvider)

	// Pre-add a rule
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules, &dsl.PricingRule{
		ID:          "pricing_1",
		Priority:    0,
		Description: "報名費",
		Condition:   &dsl.Expression{Type: "always_true"},
		Action: &dsl.Action{
			Type:  "set_price",
			Item:  "registration_fee",
			Value: 1000,
		},
	})

	resp, err := agent.Process(context.Background(), "生成 DSL", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "DSL")
	assert.NotNil(t, resp.DSL)
}

func TestAgent_Process_WithQuestions(t *testing.T) {
	mockResp := `{
		"intent": "rule_input",
		"rules": [
			{
				"id": "new_discount_early_bird",
				"action": "add",
				"rule_type": "pricing",
				"data": {
					"priority": 100,
					"description": "早鳥優惠 85 折（截止日待確認）",
					"condition": {"type": "always_true"},
					"action": {"type": "percentage_discount", "value": 15, "apply_to": ["registration_fee"]}
				}
			}
		],
		"questions": ["早鳥優惠的截止日期是什麼時候？"],
		"can_generate": false,
		"message": "已記錄早鳥優惠 85 折。請問截止日期是什麼時候？"
	}`

	mockProvider := &mockLLMProvider{response: mockResp}
	agent := NewAgent(mockProvider)

	resp, err := agent.Process(context.Background(), "早鳥優惠 85 折", "")
	require.NoError(t, err)

	assert.Contains(t, resp.Message, "截止日期")
	assert.Contains(t, resp.Questions, "截止日期")
	assert.False(t, resp.CanGenerate)
	assert.Equal(t, "clarifying", resp.State)
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"key": "value"}`, `{"key": "value"}`},
		{`Some text before {"key": "value"} and after`, `{"key": "value"}`},
		{`{"nested": {"inner": "value"}}`, `{"nested": {"inner": "value"}}`},
		{`No JSON here`, ""},
		{`{"incomplete": `, ""},
	}

	for _, tt := range tests {
		result := extractJSON(tt.input)
		assert.Equal(t, tt.expected, result, "input: %s", tt.input)
	}
}

func TestParseLLMResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkIntent string
	}{
		{
			name: "valid response",
			input: `{
				"intent": "rule_input",
				"rules": [],
				"can_generate": true,
				"message": "test"
			}`,
			expectError: false,
			checkIntent: "rule_input",
		},
		{
			name:        "invalid JSON",
			input:       "not json",
			expectError: true,
		},
		{
			name: "response with extra text",
			input: `Here is the response: {
				"intent": "general_chat",
				"rules": [],
				"can_generate": false,
				"message": "hello"
			}`,
			expectError: false,
			checkIntent: "general_chat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := ParseLLMResponse(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.checkIntent, resp.Intent)
			}
		})
	}
}

func TestParseLLMResponse_FlatRuleFormat(t *testing.T) {
	// Test that we can parse the flat format that some LLMs return
	// (action is an object instead of a string)
	input := `{
		"event_name": "2026 馬拉松 1",
		"rules": [
			{
				"id": "register_date_rule",
				"condition": {
					"type": "datetime_between",
					"field": "register_date",
					"start": "2025-08-01T00:00:00+08:00",
					"end": "2025-08-31T23:59:59+08:00"
				},
				"action": {
					"type": "set_price",
					"item": "registration_fee",
					"value": 1000,
					"label": "報名費"
				}
			}
		],
		"can_generate": true,
		"message": "已創建活動「2026 馬拉松 1」"
	}`

	resp, err := ParseLLMResponse(input)
	require.NoError(t, err)

	assert.Equal(t, "2026 馬拉松 1", resp.EventName)
	assert.Len(t, resp.Rules, 1)

	rule := resp.Rules[0]
	assert.Equal(t, "register_date_rule", rule.ID)
	assert.Equal(t, "add", rule.Action)       // Should be normalized to "add"
	assert.Equal(t, "pricing", rule.RuleType) // Should be inferred as pricing

	// Verify data is properly structured
	assert.NotEmpty(t, rule.Data)
}

func TestParseLLMResponse_StandardFormat(t *testing.T) {
	// Test that standard format still works
	input := `{
		"intent": "rule_input",
		"rules": [
			{
				"id": "new_pricing_21k",
				"action": "add",
				"rule_type": "pricing",
				"data": {
					"priority": 0,
					"description": "21K 報名費",
					"condition": {"type": "equals", "field": "user.race_type", "value": "21K"},
					"action": {"type": "set_price", "item": "registration_fee", "value": 1080}
				}
			}
		],
		"can_generate": true,
		"message": "已設定報名費"
	}`

	resp, err := ParseLLMResponse(input)
	require.NoError(t, err)

	assert.Equal(t, "rule_input", resp.Intent)
	assert.Len(t, resp.Rules, 1)

	rule := resp.Rules[0]
	assert.Equal(t, "new_pricing_21k", rule.ID)
	assert.Equal(t, "add", rule.Action)
	assert.Equal(t, "pricing", rule.RuleType)
	assert.NotEmpty(t, rule.Data)
}

func TestParseLLMResponse_MixedFormats(t *testing.T) {
	// Test with both standard and flat format rules
	input := `{
		"intent": "rule_input",
		"rules": [
			{
				"id": "standard_rule",
				"action": "add",
				"rule_type": "pricing",
				"data": {
					"description": "標準格式",
					"condition": {"type": "always_true"},
					"action": {"type": "set_price", "item": "fee", "value": 500}
				}
			},
			{
				"id": "flat_rule",
				"condition": {"type": "equals", "field": "type", "value": "21K"},
				"action": {"type": "add_item", "item": "shirt", "unit_price": 300}
			}
		],
		"message": "test"
	}`

	resp, err := ParseLLMResponse(input)
	require.NoError(t, err)

	assert.Len(t, resp.Rules, 2)

	// Standard format
	assert.Equal(t, "standard_rule", resp.Rules[0].ID)
	assert.Equal(t, "add", resp.Rules[0].Action)
	assert.Equal(t, "pricing", resp.Rules[0].RuleType)

	// Flat format should also work
	assert.Equal(t, "flat_rule", resp.Rules[1].ID)
	assert.Equal(t, "add", resp.Rules[1].Action)
	assert.Equal(t, "pricing", resp.Rules[1].RuleType)
}

func TestAgent_GetRuleSet(t *testing.T) {
	agent := NewAgent(nil)
	agent.SetEventInfo(EventInfo{
		EventID:   "test-event",
		EventName: "Test Event",
		Version:   "1.0",
	})

	// Add rules
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules, &dsl.PricingRule{
		ID:          "pricing_1",
		Priority:    0,
		Description: "報名費",
		Condition:   &dsl.Expression{Type: "always_true"},
		Action: &dsl.Action{
			Type:  "set_price",
			Item:  "registration_fee",
			Value: 1000,
		},
	})

	// Add rule definition
	agent.ruleSet.RuleDefs["is_early_bird"] = &dsl.RuleDef{
		Type:        "condition",
		Description: "早鳥期間",
		Expression: &dsl.Expression{
			Type:  "datetime_before",
			Field: "register_date",
			Value: "2026-02-28T23:59:59+08:00",
		},
	}

	// Add variable
	agent.ruleSet.Variables["base_price"] = 1000

	ruleSet := agent.GetRuleSet()

	assert.Equal(t, "test-event", ruleSet.EventID)
	assert.Equal(t, "1.0", ruleSet.Version)
	assert.Len(t, ruleSet.PricingRules, 1)
	assert.NotNil(t, ruleSet.RuleDefs["is_early_bird"])
	assert.Equal(t, 1000, ruleSet.Variables["base_price"])
}

func TestAgent_Snapshot_And_Rollback(t *testing.T) {
	agent := NewAgent(nil)

	// Add initial rule
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules, &dsl.PricingRule{
		ID:          "pricing_1",
		Description: "報名費",
		Condition:   &dsl.Expression{Type: "always_true"},
	})

	// Take snapshot
	snapshot := agent.snapshotRuleSet()

	// Modify
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules, &dsl.PricingRule{
		ID:          "pricing_2",
		Description: "折扣",
		Condition:   &dsl.Expression{Type: "always_true"},
	})

	assert.Len(t, agent.ruleSet.PricingRules, 2)

	// Rollback
	agent.restoreRuleSet(snapshot)

	assert.Len(t, agent.ruleSet.PricingRules, 1)
	assert.Equal(t, "pricing_1", agent.ruleSet.PricingRules[0].ID)
}

func TestAgent_DeleteRule(t *testing.T) {
	agent := NewAgent(nil)

	// Add rules
	agent.ruleSet.PricingRules = append(agent.ruleSet.PricingRules,
		&dsl.PricingRule{ID: "pricing_1", Description: "規則1", Condition: &dsl.Expression{Type: "always_true"}},
		&dsl.PricingRule{ID: "pricing_2", Description: "規則2", Condition: &dsl.Expression{Type: "always_true"}},
	)

	assert.Len(t, agent.ruleSet.PricingRules, 2)

	// Delete first rule
	err := agent.deleteRule("pricing_1", RuleTypePricingLLM)
	require.NoError(t, err)

	assert.Len(t, agent.ruleSet.PricingRules, 1)
	assert.Equal(t, "pricing_2", agent.ruleSet.PricingRules[0].ID)
}
