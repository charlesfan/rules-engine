package agent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
)

// LLMResponse represents the structured response from LLM
type LLMResponse struct {
	Intent           string                       `json:"intent"`
	EventName        string                       `json:"event_name,omitempty"`
	Rules            []LLMRule                    `json:"rules,omitempty"`
	RuleDefinitions  map[string]LLMRuleDefinition `json:"rule_definitions,omitempty"`
	Variables        map[string]LLMVariable       `json:"variables,omitempty"`
	DiscountStacking *LLMDiscountStacking         `json:"discount_stacking,omitempty"`
	Questions        []string                     `json:"questions,omitempty"`
	CanGenerate      bool                         `json:"can_generate"`
	Message          string                       `json:"message"`
	// Alternative field names that some LLMs might use
	Response string `json:"response,omitempty"`
}

// LLMRule represents a rule extracted by LLM
type LLMRule struct {
	ID       string          `json:"id"`
	Action   string          `json:"action"`    // add, update, delete
	RuleType string          `json:"rule_type"` // pricing, validation
	Data     json.RawMessage `json:"data"`
}

// LLMPricingRuleData represents pricing rule data
type LLMPricingRuleData struct {
	Priority    int            `json:"priority"`
	Description string         `json:"description"`
	Condition   map[string]any `json:"condition"`
	Action      map[string]any `json:"action"`
}

// LLMValidationRuleData represents validation rule data
type LLMValidationRuleData struct {
	Description  string         `json:"description"`
	Condition    map[string]any `json:"condition"`
	ErrorType    string         `json:"error_type"`
	ErrorMessage string         `json:"error_message"`
}

// LLMRuleDefinition represents a rule definition
type LLMRuleDefinition struct {
	Action string                `json:"action"` // add, update, delete
	Data   LLMRuleDefinitionData `json:"data"`
}

// LLMRuleDefinitionData represents rule definition data
type LLMRuleDefinitionData struct {
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Expression  map[string]any `json:"expression"`
}

// LLMVariable represents a variable
type LLMVariable struct {
	Action string `json:"action"` // add, update, delete
	Value  any    `json:"value"`
}

// LLMDiscountStacking represents discount stacking configuration
type LLMDiscountStacking struct {
	Mode        string `json:"mode"` // multiplicative, additive, best_only
	Description string `json:"description"`
}

// Intent constants
const (
	IntentRuleInput       = "rule_input"
	IntentModifyRule      = "modify_rule"
	IntentDeleteRule      = "delete_rule"
	IntentDSLRequest      = "dsl_request"
	IntentGeneralChat     = "general_chat"
	IntentClarifyResponse = "clarify_response"
)

// Action constants
const (
	ActionAdd    = "add"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// RuleType constants
const (
	RuleTypePricingLLM    = "pricing"
	RuleTypeValidationLLM = "validation"
)

// ruleIDCounter is used to generate unique rule IDs when LLM doesn't provide one
var ruleIDCounter uint64

// generateRuleID creates a unique rule ID based on rule type
func generateRuleID(ruleType string) string {
	counter := atomic.AddUint64(&ruleIDCounter, 1)
	if ruleType == "" {
		ruleType = RuleTypePricingLLM
	}
	return fmt.Sprintf("llm_%s_%d", ruleType, counter)
}

// rawLLMRule is a flexible structure to handle different LLM output formats
type rawLLMRule struct {
	ID           string          `json:"id"`
	Action       json.RawMessage `json:"action"` // Could be string or object
	RuleType     string          `json:"rule_type"`
	Data         json.RawMessage `json:"data"`
	Condition    json.RawMessage `json:"condition"`     // Flat format
	Priority     *int            `json:"priority"`      // Flat format
	Description  string          `json:"description"`   // Flat format
	ErrorType    string          `json:"error_type"`    // Flat format (validation)
	ErrorMessage string          `json:"error_message"` // Flat format (validation)
}

// rawLLMResponse is a flexible structure to handle different LLM output formats
type rawLLMResponse struct {
	Intent           string                       `json:"intent"`
	EventName        string                       `json:"event_name,omitempty"`
	Rules            []rawLLMRule                 `json:"rules,omitempty"`
	RuleDefinitions  map[string]LLMRuleDefinition `json:"rule_definitions,omitempty"`
	Variables        map[string]LLMVariable       `json:"variables,omitempty"`
	DiscountStacking *LLMDiscountStacking         `json:"discount_stacking,omitempty"`
	Questions        []string                     `json:"questions,omitempty"`
	CanGenerate      bool                         `json:"can_generate"`
	Message          string                       `json:"message"`
	Response         string                       `json:"response,omitempty"`
}

// ParseLLMResponse parses the LLM response JSON
func ParseLLMResponse(content string) (*LLMResponse, error) {
	// Extract JSON from content (LLM might include extra text)
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return nil, &ParseError{Message: "no JSON found in response", Content: content}
	}

	// First parse into flexible structure
	var rawResp rawLLMResponse
	if err := json.Unmarshal([]byte(jsonStr), &rawResp); err != nil {
		return nil, &ParseError{Message: err.Error(), Content: jsonStr}
	}

	// Convert to standard LLMResponse
	resp := &LLMResponse{
		Intent:           rawResp.Intent,
		EventName:        rawResp.EventName,
		RuleDefinitions:  rawResp.RuleDefinitions,
		Variables:        rawResp.Variables,
		DiscountStacking: rawResp.DiscountStacking,
		Questions:        rawResp.Questions,
		CanGenerate:      rawResp.CanGenerate,
		Message:          rawResp.Message,
		Response:         rawResp.Response,
	}

	// Normalize rules from raw format
	for _, rawRule := range rawResp.Rules {
		normalizedRule, err := normalizeRawRule(rawRule)
		if err != nil {
			// Skip invalid rules instead of failing
			continue
		}
		if normalizedRule != nil {
			resp.Rules = append(resp.Rules, *normalizedRule)
		}
	}

	// Normalize response: use Response field if Message is empty
	if resp.Message == "" && resp.Response != "" {
		resp.Message = resp.Response
	}

	// Normalize intent: map alternative names to standard ones
	resp.Intent = normalizeIntent(resp.Intent)

	return resp, nil
}

// normalizeRawRule converts a raw LLM rule to standard format
func normalizeRawRule(raw rawLLMRule) (*LLMRule, error) {
	rule := &LLMRule{
		ID:       raw.ID,
		RuleType: raw.RuleType,
	}

	// Auto-generate ID if missing
	if rule.ID == "" {
		rule.ID = generateRuleID(rule.RuleType)
	}

	// Handle action field - could be string or object
	if len(raw.Action) > 0 {
		// Try to parse as string first
		var actionStr string
		if err := json.Unmarshal(raw.Action, &actionStr); err == nil {
			// It's a string (expected format)
			rule.Action = actionStr
			rule.Data = raw.Data
		} else {
			// It's an object (flat format) - LLM put pricing action here
			// Set action to "add" and build data from flat fields
			rule.Action = ActionAdd

			// Build data from flat format
			data := make(map[string]any)

			// Add condition if present
			if len(raw.Condition) > 0 {
				var condition map[string]any
				if err := json.Unmarshal(raw.Condition, &condition); err == nil {
					data["condition"] = condition
				}
			}

			// Add action (the pricing action object)
			var actionObj map[string]any
			if err := json.Unmarshal(raw.Action, &actionObj); err == nil {
				data["action"] = actionObj
				// Infer rule_type from action
				if rule.RuleType == "" {
					rule.RuleType = inferRuleTypeFromAction(actionObj)
				}
			}

			// Add optional fields
			if raw.Priority != nil {
				data["priority"] = *raw.Priority
			}
			if raw.Description != "" {
				data["description"] = raw.Description
			}
			if raw.ErrorType != "" {
				data["error_type"] = raw.ErrorType
			}
			if raw.ErrorMessage != "" {
				data["error_message"] = raw.ErrorMessage
			}

			// Marshal data back to JSON
			dataBytes, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			rule.Data = dataBytes
		}
	} else if len(raw.Condition) > 0 {
		// No action field but has condition - flat format without action
		rule.Action = ActionAdd

		data := make(map[string]any)

		var condition map[string]any
		if err := json.Unmarshal(raw.Condition, &condition); err == nil {
			data["condition"] = condition
		}

		if raw.Priority != nil {
			data["priority"] = *raw.Priority
		}
		if raw.Description != "" {
			data["description"] = raw.Description
		}
		if raw.ErrorType != "" {
			data["error_type"] = raw.ErrorType
			rule.RuleType = RuleTypeValidationLLM
		}
		if raw.ErrorMessage != "" {
			data["error_message"] = raw.ErrorMessage
		}

		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		rule.Data = dataBytes
	} else if len(raw.Data) > 0 {
		// Standard format with data field
		// Try to parse action as string
		var actionStr string
		if len(raw.Action) > 0 {
			if err := json.Unmarshal(raw.Action, &actionStr); err == nil {
				rule.Action = actionStr
			}
		}
		if rule.Action == "" {
			rule.Action = ActionAdd
		}
		rule.Data = raw.Data
	} else {
		// No usable data
		return nil, nil
	}

	// Ensure rule_type is set
	if rule.RuleType == "" {
		rule.RuleType = inferRuleTypeFromData(rule.Data)
	}

	return rule, nil
}

// inferRuleTypeFromAction infers rule type from action object
func inferRuleTypeFromAction(action map[string]any) string {
	if actionType, ok := action["type"].(string); ok {
		pricingActions := []string{"set_price", "add_item", "percentage_discount", "fixed_discount", "price_cap"}
		for _, pa := range pricingActions {
			if actionType == pa {
				return RuleTypePricingLLM
			}
		}
	}
	return RuleTypePricingLLM // Default to pricing
}

// inferRuleTypeFromData infers rule type from data field
func inferRuleTypeFromData(data json.RawMessage) string {
	if len(data) == 0 {
		return RuleTypePricingLLM
	}

	dataStr := string(data)
	// Check for validation rule indicators
	if strings.Contains(dataStr, "error_type") || strings.Contains(dataStr, "error_message") {
		return RuleTypeValidationLLM
	}
	// Check for pricing action indicators
	pricingActions := []string{"set_price", "add_item", "percentage_discount", "fixed_discount", "price_cap"}
	for _, pa := range pricingActions {
		if strings.Contains(dataStr, pa) {
			return RuleTypePricingLLM
		}
	}
	return RuleTypePricingLLM // Default to pricing
}

// normalizeIntent maps alternative intent names to standard ones
func normalizeIntent(intent string) string {
	switch strings.ToLower(intent) {
	case "greeting", "hello", "hi":
		return IntentGeneralChat
	case "rule_create", "create_rule", "add_rule":
		return IntentRuleInput
	case "modify", "update", "edit":
		return IntentModifyRule
	case "delete", "remove":
		return IntentDeleteRule
	case "generate", "dsl", "generate_dsl":
		return IntentDSLRequest
	case "clarify", "answer":
		return IntentClarifyResponse
	default:
		return intent
	}
}

// ParseError represents a parsing error
type ParseError struct {
	Message string
	Content string
}

func (e *ParseError) Error() string {
	return "parse error: " + e.Message
}

// IsDSLRequest returns true if the intent is to generate DSL
func (r *LLMResponse) IsDSLRequest() bool {
	return r.Intent == IntentDSLRequest
}

// IsGeneralChat returns true if the intent is general chat
func (r *LLMResponse) IsGeneralChat() bool {
	return r.Intent == IntentGeneralChat
}

// HasRules returns true if there are rules to process
func (r *LLMResponse) HasRules() bool {
	return len(r.Rules) > 0
}

// HasQuestions returns true if there are questions to ask
func (r *LLMResponse) HasQuestions() bool {
	return len(r.Questions) > 0
}

// GetAddRules returns rules with action "add"
func (r *LLMResponse) GetAddRules() []LLMRule {
	var rules []LLMRule
	for _, rule := range r.Rules {
		if rule.Action == ActionAdd {
			rules = append(rules, rule)
		}
	}
	return rules
}

// GetUpdateRules returns rules with action "update"
func (r *LLMResponse) GetUpdateRules() []LLMRule {
	var rules []LLMRule
	for _, rule := range r.Rules {
		if rule.Action == ActionUpdate {
			rules = append(rules, rule)
		}
	}
	return rules
}

// GetDeleteRules returns rules with action "delete"
func (r *LLMResponse) GetDeleteRules() []LLMRule {
	var rules []LLMRule
	for _, rule := range r.Rules {
		if rule.Action == ActionDelete {
			rules = append(rules, rule)
		}
	}
	return rules
}

// extractJSON extracts JSON object from a string that may contain other text
func extractJSON(content string) string {
	// Find JSON object boundaries
	start := strings.Index(content, "{")
	if start == -1 {
		return ""
	}

	// Find matching closing brace
	depth := 0
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[start : i+1]
			}
		}
	}

	return ""
}
