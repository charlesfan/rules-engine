package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charlesfan/rules-engine/ai-poc/internal/chatbot/tools"
	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/charlesfan/rules-engine/internal/rules/parser"
)

/**
 * Agent orchestrates the chatbot workflow using a unified prompt approach:
 *   1. User input → Agent forwards to LLM with context
 *   2. LLM analyzes and returns incremental operations (add/update/delete)
 *   3. Agent applies operations to RuleSet
 *   4. parser.Validate() validates after each operation
 *   5. Rollback on validation failure
 */

// AgentState represents the current state of the conversation
type AgentState int

const (
	StateIdle       AgentState = iota // Initial state, no rules collected
	StateCollecting                   // Collecting rules from user
	StateClarifying                   // Waiting for user to clarify
	StateReady                        // Ready to generate DSL
)

// String returns the string representation of AgentState
func (s AgentState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateCollecting:
		return "collecting"
	case StateClarifying:
		return "clarifying"
	case StateReady:
		return "ready"
	default:
		return "unknown"
	}
}

// EventInfo contains basic event information
type EventInfo struct {
	EventID   string `json:"event_id"`
	EventName string `json:"event_name"`
	Version   string `json:"version"`
}

// Agent is the main orchestrator for the chatbot
type Agent struct {
	provider  llm.Provider
	parser    *parser.Parser
	validator *tools.DSLValidator

	// State management
	state     AgentState
	eventInfo EventInfo

	// Single source of truth - the RuleSet from rules package
	ruleSet *dsl.RuleSet

	// Generated DSL cache
	lastDSL json.RawMessage

	// Pending questions from LLM
	pendingQuestions []string

	// Conversation context (summary from memory)
	conversationContext string
}

// NewAgent creates a new Agent with all components
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{
		provider:  provider,
		parser:    parser.NewParser(),
		validator: tools.NewDSLValidator(),
		state:     StateIdle,
		ruleSet:   newEmptyRuleSet("event-001", "1.0", ""),
		eventInfo: EventInfo{
			EventID: "event-001",
			Version: "1.0",
		},
	}
}

// newEmptyRuleSet creates an empty RuleSet with basic info
func newEmptyRuleSet(eventID, version, name string) *dsl.RuleSet {
	return &dsl.RuleSet{
		EventID:         eventID,
		Version:         version,
		Name:            name,
		Variables:       make(map[string]interface{}),
		RuleDefs:        make(map[string]*dsl.RuleDef),
		PricingRules:    []*dsl.PricingRule{},
		ValidationRules: []*dsl.ValidationRule{},
	}
}

// inferFormSchema automatically generates FormSchema from pricing rules
// It extracts race types from conditions like {"type":"equals","field":"user.race_type","value":"21K"}
func (a *Agent) inferFormSchema() {
	// Collect unique race types from pricing rules
	raceTypes := make(map[string]string) // value -> label

	for _, rule := range a.ruleSet.PricingRules {
		if rule.Condition == nil {
			continue
		}
		a.extractRaceTypesFromCondition(rule.Condition, rule.Description, raceTypes)
	}

	// If no race types found, don't create FormSchema
	if len(raceTypes) == 0 {
		return
	}

	// Build FormSchema
	var options []dsl.FormFieldOption
	for value, label := range raceTypes {
		options = append(options, dsl.FormFieldOption{
			Label: label,
			Value: value,
		})
	}

	a.ruleSet.FormSchema = &dsl.FormSchema{
		Fields: []dsl.FormField{
			{
				ID:       "race_type",
				Label:    "報名組別",
				Type:     "select",
				Field:    "user.race_type",
				Required: true,
				Options:  options,
			},
		},
	}
}

// extractRaceTypesFromCondition recursively extracts race types from condition expressions
func (a *Agent) extractRaceTypesFromCondition(expr *dsl.Expression, description string, raceTypes map[string]string) {
	if expr == nil {
		return
	}

	// Check if this is a race_type equals condition
	if expr.Type == "equals" && expr.Field == "user.race_type" {
		if value, ok := expr.Value.(string); ok {
			// Use description as label if available, otherwise use value
			label := value
			if description != "" {
				// Try to extract a meaningful label from description
				// e.g., "21K 報名費" -> "21K"
				if strings.Contains(description, value) {
					label = value
				}
			}
			// Map common race type names
			switch value {
			case "21K":
				label = "半程馬拉松 (21K)"
			case "42K":
				label = "全程馬拉松 (42K)"
			case "10K":
				label = "10K"
			case "5K":
				label = "5K"
			}
			raceTypes[value] = label
		}
	}

	// Recursively check nested conditions
	if expr.Conditions != nil {
		for _, cond := range expr.Conditions {
			a.extractRaceTypesFromCondition(cond, description, raceTypes)
		}
	}
	if expr.Condition != nil {
		a.extractRaceTypesFromCondition(expr.Condition, description, raceTypes)
	}
}

// Response represents the agent's response to user input
type Response struct {
	Message     string          `json:"message"`
	DSL         json.RawMessage `json:"dsl,omitempty"`
	Questions   string          `json:"questions,omitempty"`
	Summary     string          `json:"summary,omitempty"`
	State       string          `json:"state"`
	RuleCount   int             `json:"rule_count"`
	CanGenerate bool            `json:"can_generate"`
}

// Process handles user input and returns a response using two-stage LLM calls
func (a *Agent) Process(ctx context.Context, input string, conversationSummary string) (*Response, error) {
	// Store conversation context for use in prompts
	a.conversationContext = conversationSummary

	// If no LLM provider, use fallback
	if a.provider == nil {
		return a.processFallback(ctx, input)
	}

	// Stage 1: Intent Detection
	fmt.Printf("\n[DEBUG] Stage 1: Detecting intent...\n")
	fmt.Printf("[DEBUG] Conversation context: %s\n", truncateForDebug(conversationSummary, 100))
	detected, err := a.detectIntent(ctx, input)
	if err != nil {
		fmt.Printf("[DEBUG] Intent detection failed: %v, using fallback\n", err)
		return a.processFallback(ctx, input)
	}

	fmt.Printf("[DEBUG] Detected intent: %s\n", detected.Intent)
	fmt.Printf("[DEBUG] Raw intent response: %s\n[/DEBUG]\n", detected.Raw)

	// Stage 2: Route to appropriate handler based on intent
	switch detected.Intent {
	case IntentGeneralChat:
		return a.handleGeneralChat(ctx, input)

	case IntentRuleInput:
		return a.handleRuleOperation(ctx, input, IntentRuleInput)

	case IntentModifyRule:
		return a.handleRuleOperation(ctx, input, IntentModifyRule)

	case IntentDeleteRule:
		return a.handleRuleOperation(ctx, input, IntentDeleteRule)

	case IntentClarifyResponse:
		return a.handleClarifyResponse(ctx, input)

	case IntentDSLRequest:
		return a.handleDSLGeneration(ctx)

	default:
		// Default to general chat
		return a.handleGeneralChat(ctx, input)
	}
}

// truncateForDebug truncates a string for debug output
func truncateForDebug(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// getRuleCount returns total number of rules
func (a *Agent) getRuleCount() int {
	return len(a.ruleSet.PricingRules) + len(a.ruleSet.ValidationRules)
}

// generateDefaultMessage generates a default message when LLM returns empty message
func (a *Agent) generateDefaultMessage(resp *LLMResponse) string {
	switch resp.Intent {
	case IntentGeneralChat:
		return "你好！我是賽事報名規則助手。請描述您的賽事規則，我會幫您轉換成 DSL。"
	case IntentRuleInput:
		if len(resp.Rules) > 0 {
			return fmt.Sprintf("已理解 %d 條規則。", len(resp.Rules))
		}
		return "已收到您的規則描述。"
	case IntentModifyRule:
		return "已更新規則。"
	case IntentDeleteRule:
		return "已刪除規則。"
	case IntentDSLRequest:
		return "正在生成 DSL..."
	case IntentClarifyResponse:
		return "感謝您的說明。"
	default:
		return "已處理您的請求。"
	}
}

// processLLMResponse processes the structured LLM response
func (a *Agent) processLLMResponse(ctx context.Context, resp *LLMResponse) (*Response, error) {
	// Snapshot current state for rollback
	snapshot := a.snapshotRuleSet()

	// Apply operations
	var validationErrors []string

	// Process event name if provided
	if resp.EventName != "" {
		a.ruleSet.Name = resp.EventName
		a.eventInfo.EventName = resp.EventName
	}

	// Process rules (add/update/delete)
	for _, rule := range resp.Rules {
		if err := a.applyRuleOperation(rule); err != nil {
			validationErrors = append(validationErrors, err.Error())
		}
	}

	// Process rule definitions
	for name, def := range resp.RuleDefinitions {
		if err := a.applyRuleDefOperation(name, def); err != nil {
			validationErrors = append(validationErrors, err.Error())
		}
	}

	// Process variables
	for name, v := range resp.Variables {
		a.applyVariableOperation(name, v)
	}

	// Process discount stacking
	if resp.DiscountStacking != nil {
		a.ruleSet.DiscountStacking = &dsl.DiscountStacking{
			Mode:        resp.DiscountStacking.Mode,
			Description: resp.DiscountStacking.Description,
		}
	}

	// Validate the updated RuleSet
	if err := a.parser.Validate(a.ruleSet); err != nil {
		// Rollback on validation failure
		a.restoreRuleSet(snapshot)
		validationErrors = append(validationErrors, err.Error())
	}

	// Update state based on response
	a.updateState(resp)

	// Store pending questions
	a.pendingQuestions = resp.Questions

	// Build response message
	message := resp.Message
	if len(validationErrors) > 0 {
		message += "\n\n⚠️ 驗證警告：\n" + strings.Join(validationErrors, "\n")
	}

	// Build response
	response := &Response{
		Message:     message,
		State:       a.state.String(),
		RuleCount:   a.getRuleCount(),
		CanGenerate: resp.CanGenerate && len(validationErrors) == 0,
	}

	// If DSL request, generate DSL
	if resp.Intent == IntentDSLRequest && response.CanGenerate {
		return a.handleDSLRequest(ctx, response)
	}

	// Add questions if any
	if len(resp.Questions) > 0 {
		response.Questions = strings.Join(resp.Questions, "\n")
	}

	return response, nil
}

// snapshotRuleSet creates a deep copy of current RuleSet for rollback
func (a *Agent) snapshotRuleSet() *dsl.RuleSet {
	data, _ := json.Marshal(a.ruleSet)
	var snapshot dsl.RuleSet
	json.Unmarshal(data, &snapshot)
	return &snapshot
}

// restoreRuleSet restores RuleSet from snapshot
func (a *Agent) restoreRuleSet(snapshot *dsl.RuleSet) {
	a.ruleSet = snapshot
}

// applyRuleOperation applies add/update/delete operation for a rule
func (a *Agent) applyRuleOperation(rule LLMRule) error {
	// Parse rule data
	var data struct {
		Priority     int             `json:"priority"`
		Description  string          `json:"description"`
		Condition    json.RawMessage `json:"condition"`
		Action       json.RawMessage `json:"action"`
		ErrorType    string          `json:"error_type"`
		ErrorMessage string          `json:"error_message"`
	}
	if err := json.Unmarshal(rule.Data, &data); err != nil {
		return fmt.Errorf("invalid rule data: %w", err)
	}

	// Skip incomplete rules (empty or null condition)
	if len(data.Condition) == 0 || string(data.Condition) == "{}" || string(data.Condition) == "null" {
		return nil // Skip silently - rule is incomplete
	}

	// Parse condition
	var condition dsl.Expression
	if err := json.Unmarshal(data.Condition, &condition); err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	// Skip if condition type is empty (incomplete rule from LLM)
	if condition.Type == "" {
		return nil
	}

	// Normalize rule_type (handle LLM mistakes like "pricing|validation")
	ruleType := normalizeRuleType(rule.RuleType, data.Action)

	// Default error_type for validation rules
	if ruleType == RuleTypeValidationLLM && data.ErrorType == "" {
		data.ErrorType = "blocking"
	}

	// Resolve ID for new rules
	id := rule.ID
	if strings.HasPrefix(id, "new_") {
		id = fmt.Sprintf("%s_%d", ruleType, a.getRuleCount()+1)
	}

	switch rule.Action {
	case ActionAdd:
		return a.addRule(id, ruleType, data, &condition)
	case ActionUpdate:
		return a.updateRule(id, ruleType, data, &condition)
	case ActionDelete:
		return a.deleteRule(id, ruleType)
	}

	return nil
}

// normalizeRuleType normalizes rule_type from LLM output
// Handles common mistakes like "pricing|validation" or wrong classification
func normalizeRuleType(ruleType string, actionData json.RawMessage) string {
	ruleTypeLower := strings.ToLower(ruleType)

	// If action contains pricing-related types, it's a pricing rule
	if len(actionData) > 0 {
		actionStr := string(actionData)
		pricingActions := []string{"set_price", "add_item", "percentage_discount", "fixed_discount", "price_cap"}
		for _, pa := range pricingActions {
			if strings.Contains(actionStr, pa) {
				return RuleTypePricingLLM
			}
		}
	}

	// Check if contains "pricing" keyword
	if strings.Contains(ruleTypeLower, "pricing") {
		return RuleTypePricingLLM
	}

	// Check if contains "validation" keyword
	if strings.Contains(ruleTypeLower, "validation") {
		return RuleTypeValidationLLM
	}

	// Default: if has action data, likely pricing; otherwise validation
	if len(actionData) > 2 { // more than just "{}"
		return RuleTypePricingLLM
	}

	return RuleTypeValidationLLM
}

// addRule adds a new rule to RuleSet
func (a *Agent) addRule(id, ruleType string, data struct {
	Priority     int             `json:"priority"`
	Description  string          `json:"description"`
	Condition    json.RawMessage `json:"condition"`
	Action       json.RawMessage `json:"action"`
	ErrorType    string          `json:"error_type"`
	ErrorMessage string          `json:"error_message"`
}, condition *dsl.Expression) error {

	if ruleType == RuleTypePricingLLM {
		var action dsl.Action
		if err := json.Unmarshal(data.Action, &action); err != nil {
			return fmt.Errorf("invalid action: %w", err)
		}

		a.ruleSet.PricingRules = append(a.ruleSet.PricingRules, &dsl.PricingRule{
			ID:          id,
			Priority:    data.Priority,
			Description: data.Description,
			Condition:   condition,
			Action:      &action,
		})
	} else {
		a.ruleSet.ValidationRules = append(a.ruleSet.ValidationRules, &dsl.ValidationRule{
			ID:           id,
			Description:  data.Description,
			Condition:    condition,
			ErrorType:    data.ErrorType,
			ErrorMessage: data.ErrorMessage,
		})
	}

	return nil
}

// updateRule updates an existing rule in RuleSet
func (a *Agent) updateRule(id, ruleType string, data struct {
	Priority     int             `json:"priority"`
	Description  string          `json:"description"`
	Condition    json.RawMessage `json:"condition"`
	Action       json.RawMessage `json:"action"`
	ErrorType    string          `json:"error_type"`
	ErrorMessage string          `json:"error_message"`
}, condition *dsl.Expression) error {

	if ruleType == RuleTypePricingLLM {
		for i, rule := range a.ruleSet.PricingRules {
			if rule.ID == id {
				var action dsl.Action
				if err := json.Unmarshal(data.Action, &action); err != nil {
					return fmt.Errorf("invalid action: %w", err)
				}

				a.ruleSet.PricingRules[i] = &dsl.PricingRule{
					ID:          id,
					Priority:    data.Priority,
					Description: data.Description,
					Condition:   condition,
					Action:      &action,
				}
				return nil
			}
		}
		// Not found, add as new
		return a.addRule(id, ruleType, data, condition)
	} else {
		for i, rule := range a.ruleSet.ValidationRules {
			if rule.ID == id {
				a.ruleSet.ValidationRules[i] = &dsl.ValidationRule{
					ID:           id,
					Description:  data.Description,
					Condition:    condition,
					ErrorType:    data.ErrorType,
					ErrorMessage: data.ErrorMessage,
				}
				return nil
			}
		}
		// Not found, add as new
		return a.addRule(id, ruleType, data, condition)
	}
}

// deleteRule removes a rule from RuleSet
func (a *Agent) deleteRule(id, ruleType string) error {
	if ruleType == RuleTypePricingLLM {
		for i, rule := range a.ruleSet.PricingRules {
			if rule.ID == id {
				a.ruleSet.PricingRules = append(a.ruleSet.PricingRules[:i], a.ruleSet.PricingRules[i+1:]...)
				return nil
			}
		}
	} else {
		for i, rule := range a.ruleSet.ValidationRules {
			if rule.ID == id {
				a.ruleSet.ValidationRules = append(a.ruleSet.ValidationRules[:i], a.ruleSet.ValidationRules[i+1:]...)
				return nil
			}
		}
	}
	return nil // Not found is OK for delete
}

// applyRuleDefOperation applies operation for rule definition
func (a *Agent) applyRuleDefOperation(name string, def LLMRuleDefinition) error {
	switch def.Action {
	case ActionAdd, ActionUpdate:
		exprBytes, _ := json.Marshal(def.Data.Expression)
		var expr dsl.Expression
		if err := json.Unmarshal(exprBytes, &expr); err != nil {
			return fmt.Errorf("invalid expression for %s: %w", name, err)
		}

		if a.ruleSet.RuleDefs == nil {
			a.ruleSet.RuleDefs = make(map[string]*dsl.RuleDef)
		}
		a.ruleSet.RuleDefs[name] = &dsl.RuleDef{
			Type:        def.Data.Type,
			Description: def.Data.Description,
			Expression:  &expr,
		}

	case ActionDelete:
		delete(a.ruleSet.RuleDefs, name)
	}

	return nil
}

// applyVariableOperation applies operation for variable
func (a *Agent) applyVariableOperation(name string, v LLMVariable) {
	if a.ruleSet.Variables == nil {
		a.ruleSet.Variables = make(map[string]interface{})
	}

	switch v.Action {
	case ActionAdd, ActionUpdate:
		a.ruleSet.Variables[name] = v.Value
	case ActionDelete:
		delete(a.ruleSet.Variables, name)
	}
}

// updateState updates agent state based on LLM response
func (a *Agent) updateState(resp *LLMResponse) {
	if a.getRuleCount() == 0 {
		a.state = StateIdle
		return
	}

	if len(resp.Questions) > 0 && !resp.CanGenerate {
		a.state = StateClarifying
		return
	}

	if resp.CanGenerate {
		a.state = StateReady
		return
	}

	a.state = StateCollecting
}

// handleDSLRequest generates DSL from RuleSet
func (a *Agent) handleDSLRequest(ctx context.Context, response *Response) (*Response, error) {
	if a.getRuleCount() == 0 {
		response.Message = "目前沒有規則可以生成 DSL。請先描述您的賽事規則。"
		response.CanGenerate = false
		return response, nil
	}

	// Infer FormSchema from pricing rules (extract race types, etc.)
	a.inferFormSchema()

	// Final validation
	if err := a.parser.Validate(a.ruleSet); err != nil {
		response.Message = fmt.Sprintf("⚠️ DSL 驗證失敗：%v\n請修正規則後重新生成。", err)
		response.CanGenerate = false
		return response, nil
	}

	// Marshal RuleSet to JSON
	dslBytes, err := json.MarshalIndent(a.ruleSet, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DSL: %w", err)
	}

	// Additional validation with tools.DSLValidator
	validationResult := a.validator.Validate(dslBytes)

	// Build response
	var msgBuilder strings.Builder

	if !validationResult.Valid {
		msgBuilder.WriteString("⚠️ DSL 生成完成，但有驗證錯誤：\n")
		msgBuilder.WriteString(validationResult.FormatErrors())
		msgBuilder.WriteString("\n請修正規則後重新生成。")
	} else {
		msgBuilder.WriteString("✅ DSL 生成成功！\n\n")

		// Add warnings if any
		if len(validationResult.Warnings) > 0 {
			msgBuilder.WriteString("⚠️ 注意事項：\n")
			msgBuilder.WriteString(validationResult.FormatWarnings())
			msgBuilder.WriteString("\n")
		}

		// Add summary
		if validationResult.Summary != nil {
			msgBuilder.WriteString("\n📊 DSL 摘要：\n")
			msgBuilder.WriteString(validationResult.FormatSummary())
		}

		// Add DSL content
		msgBuilder.WriteString("\n```json\n")
		msgBuilder.Write(dslBytes)
		msgBuilder.WriteString("\n```\n")

		a.lastDSL = dslBytes
	}

	response.Message = msgBuilder.String()
	response.DSL = dslBytes

	return response, nil
}

// buildExistingRulesContext builds a text representation of existing rules for the prompt
func (a *Agent) buildExistingRulesContext() string {
	if a.getRuleCount() == 0 {
		return ""
	}

	var sb strings.Builder
	i := 1

	// Pricing rules
	for _, rule := range a.ruleSet.PricingRules {
		sb.WriteString(fmt.Sprintf("%d. [pricing] id=%s: %s\n", i, rule.ID, rule.Description))
		i++
	}

	// Validation rules
	for _, rule := range a.ruleSet.ValidationRules {
		sb.WriteString(fmt.Sprintf("%d. [validation] id=%s: %s\n", i, rule.ID, rule.Description))
		i++
	}

	// Rule definitions
	if len(a.ruleSet.RuleDefs) > 0 {
		sb.WriteString("\n規則定義:\n")
		for name, def := range a.ruleSet.RuleDefs {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", name, def.Description))
		}
	}

	// Variables
	if len(a.ruleSet.Variables) > 0 {
		sb.WriteString("\n變數:\n")
		for name, val := range a.ruleSet.Variables {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", name, val))
		}
	}

	// Discount stacking
	if a.ruleSet.DiscountStacking != nil {
		sb.WriteString(fmt.Sprintf("\n折扣疊加模式: %s (%s)\n",
			a.ruleSet.DiscountStacking.Mode,
			a.ruleSet.DiscountStacking.Description))
	}

	return sb.String()
}

// processFallback handles input when LLM is unavailable
func (a *Agent) processFallback(ctx context.Context, input string) (*Response, error) {
	inputLower := strings.ToLower(input)

	// Check for DSL generation request
	if strings.Contains(inputLower, "生成") && strings.Contains(inputLower, "dsl") ||
		strings.Contains(inputLower, "完成") ||
		strings.Contains(inputLower, "確認規則") {
		return a.handleDSLRequest(ctx, &Response{
			State:       a.state.String(),
			RuleCount:   a.getRuleCount(),
			CanGenerate: a.getRuleCount() > 0,
		})
	}

	// Check for rule-related keywords
	ruleKeywords := []string{
		"報名費", "費用", "價格", "元",
		"折扣", "優惠", "打折", "折",
		"早鳥", "團報", "團體",
		"限制", "條件", "驗證",
		"截止", "期限", "日期",
		"保險", "加購", "紀念",
	}

	hasRuleKeyword := false
	for _, keyword := range ruleKeywords {
		if strings.Contains(inputLower, keyword) {
			hasRuleKeyword = true
			break
		}
	}

	if hasRuleKeyword {
		// Extract basic rules using pattern matching
		rules := a.simpleExtractRules(input)
		for _, rule := range rules {
			id := fmt.Sprintf("%s_%d", rule.Type, a.getRuleCount()+1)
			a.ruleSet.PricingRules = append(a.ruleSet.PricingRules, &dsl.PricingRule{
				ID:          id,
				Priority:    0,
				Description: rule.Description,
				Condition:   &dsl.Expression{Type: "always_true"},
			})
		}

		a.state = StateCollecting

		return &Response{
			Message:     fmt.Sprintf("已理解 %d 條規則（無 LLM 模式，僅基本解析）。請輸入「生成 DSL」來產生規則。", len(rules)),
			State:       a.state.String(),
			RuleCount:   a.getRuleCount(),
			CanGenerate: a.getRuleCount() > 0,
		}, nil
	}

	// General greeting
	return &Response{
		Message:     "您好！我是賽事規則 DSL 助手。請描述您的賽事報名規則，我會幫您轉換成系統可執行的 DSL。",
		State:       a.state.String(),
		RuleCount:   a.getRuleCount(),
		CanGenerate: a.getRuleCount() > 0,
	}, nil
}

// simpleExtractRules performs basic pattern-based rule extraction (fallback)
func (a *Agent) simpleExtractRules(input string) []tools.ExtractedRuleInfo {
	rules := []tools.ExtractedRuleInfo{}
	inputLower := strings.ToLower(input)

	// Check for pricing rules
	if strings.Contains(inputLower, "報名費") || strings.Contains(inputLower, "費用") ||
		strings.Contains(inputLower, "價格") || strings.Contains(input, "NT$") {
		price := extractNumber(input)
		details := map[string]interface{}{}
		if price > 0 {
			details["price"] = price
		}
		rules = append(rules, tools.ExtractedRuleInfo{
			Type:        tools.RuleTypePricing,
			Description: fmt.Sprintf("報名費 %.0f 元", price),
			Complete:    price > 0,
			Details:     details,
		})
	}

	// Check for early bird discount
	if strings.Contains(inputLower, "早鳥") {
		discount := extractDiscount(input)
		rules = append(rules, tools.ExtractedRuleInfo{
			Type:        tools.RuleTypeDiscount,
			Description: fmt.Sprintf("早鳥優惠 %.0f%% off", discount),
			Complete:    discount > 0,
			Details:     map[string]interface{}{"discount": discount},
		})
	}

	// Check for group discount
	if strings.Contains(inputLower, "團報") || strings.Contains(inputLower, "團體") {
		discount := extractDiscount(input)
		size := extractTeamSize(input)
		rules = append(rules, tools.ExtractedRuleInfo{
			Type:        tools.RuleTypeDiscount,
			Description: fmt.Sprintf("團報 %d 人以上 %.0f%% off", size, discount),
			Complete:    discount > 0 && size > 0,
			Details:     map[string]interface{}{"discount": discount, "min_size": size},
		})
	}

	// Check for validation rules
	if strings.Contains(inputLower, "截止") || strings.Contains(inputLower, "期限") {
		rules = append(rules, tools.ExtractedRuleInfo{
			Type:        tools.RuleTypeValidation,
			Description: "報名截止日期",
			Complete:    false,
			Details:     map[string]interface{}{},
		})
	}

	if strings.Contains(inputLower, "年齡") || strings.Contains(inputLower, "歲") {
		rules = append(rules, tools.ExtractedRuleInfo{
			Type:        tools.RuleTypeValidation,
			Description: "年齡限制",
			Complete:    false,
			Details:     map[string]interface{}{},
		})
	}

	// Check for addon rules
	if strings.Contains(inputLower, "加購") || strings.Contains(inputLower, "紀念衫") ||
		strings.Contains(inputLower, "保險") {
		rules = append(rules, tools.ExtractedRuleInfo{
			Type:        tools.RuleTypeAddon,
			Description: "加購項目",
			Complete:    false,
			Details:     map[string]interface{}{},
		})
	}

	return rules
}

// SetEventInfo sets the event information
func (a *Agent) SetEventInfo(info EventInfo) {
	a.eventInfo = info
	a.ruleSet.EventID = info.EventID
	a.ruleSet.Version = info.Version
	a.ruleSet.Name = info.EventName
}

// GetState returns the current state
func (a *Agent) GetState() AgentState {
	return a.state
}

// GetRuleCount returns the number of stored rules
func (a *Agent) GetRuleCount() int {
	return a.getRuleCount()
}

// GetLastDSL returns the last generated DSL
func (a *Agent) GetLastDSL() json.RawMessage {
	return a.lastDSL
}

// GetRuleSet returns the current RuleSet (for external use like calculator)
func (a *Agent) GetRuleSet() *dsl.RuleSet {
	return a.ruleSet
}

// Clear resets the agent state
func (a *Agent) Clear() {
	a.state = StateIdle
	a.ruleSet = newEmptyRuleSet(a.eventInfo.EventID, a.eventInfo.Version, a.eventInfo.EventName)
	a.pendingQuestions = nil
	a.lastDSL = nil
}

// GetRulesSummary returns a summary of stored rules
func (a *Agent) GetRulesSummary() string {
	if a.getRuleCount() == 0 {
		return "目前沒有已建立的規則"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("目前有 %d 條規則：\n", a.getRuleCount()))

	i := 1
	for _, rule := range a.ruleSet.PricingRules {
		sb.WriteString(fmt.Sprintf("  %d. [pricing] %s ✓\n", i, rule.Description))
		i++
	}
	for _, rule := range a.ruleSet.ValidationRules {
		sb.WriteString(fmt.Sprintf("  %d. [validation] %s ✓\n", i, rule.Description))
		i++
	}

	return sb.String()
}

// Helper functions for extracting values from text

// extractNumber extracts a number from text
func extractNumber(text string) float64 {
	var num float64
	for i := 0; i < len(text); i++ {
		if text[i] >= '0' && text[i] <= '9' {
			start := i
			for i < len(text) && (text[i] >= '0' && text[i] <= '9') {
				i++
			}
			fmt.Sscanf(text[start:i], "%f", &num)
			if num > 0 {
				return num
			}
		}
	}
	return 0
}

// extractDiscount extracts discount percentage from text
func extractDiscount(text string) float64 {
	textLower := strings.ToLower(text)
	var discount float64

	// Look for patterns like "9折", "95折"
	if strings.Contains(textLower, "折") {
		idx := strings.Index(textLower, "折")
		if idx > 0 {
			numStr := ""
			for i := idx - 1; i >= 0 && (text[i] >= '0' && text[i] <= '9' || text[i] == '.'); i-- {
				numStr = string(text[i]) + numStr
			}
			if numStr != "" {
				fmt.Sscanf(numStr, "%f", &discount)
				// Convert to percentage off (e.g., 9折 = 10% off)
				if discount > 0 && discount < 10 {
					discount = (10 - discount) * 10
				} else if discount >= 10 && discount < 100 {
					discount = 100 - discount
				}
			}
		}
	}

	// Percentage patterns
	if strings.Contains(text, "%") {
		idx := strings.Index(text, "%")
		if idx > 0 {
			numStr := ""
			for i := idx - 1; i >= 0 && (text[i] >= '0' && text[i] <= '9' || text[i] == '.'); i-- {
				numStr = string(text[i]) + numStr
			}
			if numStr != "" {
				fmt.Sscanf(numStr, "%f", &discount)
			}
		}
	}

	return discount
}

// extractTeamSize extracts team size from text
func extractTeamSize(text string) int {
	textLower := strings.ToLower(text)

	if strings.Contains(textLower, "人") {
		idx := strings.Index(textLower, "人")
		if idx > 0 {
			i := idx - 1
			for i >= 0 && text[i] == ' ' {
				i--
			}
			numStr := ""
			for i >= 0 && text[i] >= '0' && text[i] <= '9' {
				numStr = string(text[i]) + numStr
				i--
			}
			if numStr != "" {
				var size int
				fmt.Sscanf(numStr, "%d", &size)
				return size
			}
		}
	}

	return 0
}
