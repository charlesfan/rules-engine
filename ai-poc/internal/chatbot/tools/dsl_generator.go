package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charlesfan/rules-engine/ai-poc/internal/chatbot/prompts"
	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

/**
 * DSLGenerator generates DSL JSON from confirmed rules using LLM.
 * It handles:
 *   - Rule extraction from natural language
 *   - DSL structure generation
 *   - Variable extraction and naming
 */

// DSLGenerator generates DSL from rules
type DSLGenerator struct {
	provider llm.Provider
}

// NewDSLGenerator creates a new DSL generator
func NewDSLGenerator(provider llm.Provider) *DSLGenerator {
	return &DSLGenerator{
		provider: provider,
	}
}

// GenerationRequest contains the input for DSL generation
type GenerationRequest struct {
	EventID        string          `json:"event_id"`
	EventName      string          `json:"event_name"`
	Version        string          `json:"version"`
	ConfirmedRules []ConfirmedRule `json:"confirmed_rules"`
}

// ConfirmedRule represents a rule that has been confirmed by the user
type ConfirmedRule struct {
	Type        RuleType               `json:"type"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
}

// RuleType represents the type of rule
type RuleType string

const (
	RuleTypePricing    RuleType = "pricing"
	RuleTypeDiscount   RuleType = "discount"
	RuleTypeValidation RuleType = "validation"
	RuleTypeAddon      RuleType = "addon"
)

// GenerationResult contains the generated DSL and explanation
type GenerationResult struct {
	DSL         json.RawMessage `json:"dsl"`
	Explanation string          `json:"explanation"`
	Warnings    []string        `json:"warnings,omitempty"`
}

/**
 * Generate generates DSL from confirmed rules using LLM.
 * Returns the generated DSL JSON and a human-readable explanation.
 */
func (g *DSLGenerator) Generate(ctx context.Context, req GenerationRequest) (*GenerationResult, error) {
	if g.provider == nil {
		return nil, fmt.Errorf("no LLM provider set")
	}

	if len(req.ConfirmedRules) == 0 {
		return nil, fmt.Errorf("no confirmed rules provided")
	}

	// Set defaults
	if req.EventID == "" {
		req.EventID = "event-001"
	}
	if req.EventName == "" {
		req.EventName = "Event"
	}
	if req.Version == "" {
		req.Version = "1.0"
	}

	// Build prompt
	prompt := g.buildPrompt(req)

	// Call LLM
	llmReq := llm.ChatRequest{
		Messages: []llm.Message{
			llm.NewSystemMessage(prompts.SystemPrompt),
			llm.NewUserMessage(prompt),
		},
	}

	resp, err := g.provider.Chat(ctx, llmReq)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	return g.parseResponse(resp.Content)
}

// buildPrompt constructs the generation prompt
func (g *DSLGenerator) buildPrompt(req GenerationRequest) string {
	prompt := prompts.DSLGenerationPrompt

	// Replace placeholders
	prompt = strings.Replace(prompt, "{{EVENT_ID}}", req.EventID, 1)
	prompt = strings.Replace(prompt, "{{EVENT_NAME}}", req.EventName, 1)
	prompt = strings.Replace(prompt, "{{VERSION}}", req.Version, 1)

	// Format confirmed rules
	var rulesBuilder strings.Builder
	for i, rule := range req.ConfirmedRules {
		rulesBuilder.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, rule.Type, rule.Description))
		if len(rule.Details) > 0 {
			detailsJSON, _ := json.MarshalIndent(rule.Details, "   ", "  ")
			rulesBuilder.WriteString(fmt.Sprintf("   詳細: %s\n", string(detailsJSON)))
		}
	}

	prompt = strings.Replace(prompt, "{{CONFIRMED_RULES}}", rulesBuilder.String(), 1)

	return prompt
}

// parseResponse parses the LLM response to extract DSL and explanation
func (g *DSLGenerator) parseResponse(content string) (*GenerationResult, error) {
	result := &GenerationResult{}

	// Extract JSON from code block
	dslJSON := extractCodeBlock(content, "json")
	if dslJSON == "" {
		// Try to find raw JSON
		dslJSON = extractJSON(content)
	}

	if dslJSON == "" {
		return nil, fmt.Errorf("no DSL JSON found in response")
	}

	// Validate JSON structure
	var dslMap map[string]interface{}
	if err := json.Unmarshal([]byte(dslJSON), &dslMap); err != nil {
		return nil, fmt.Errorf("invalid DSL JSON: %w", err)
	}

	result.DSL = json.RawMessage(dslJSON)

	// Extract explanation (text after the JSON block)
	result.Explanation = extractExplanation(content, dslJSON)

	// Check for potential issues
	result.Warnings = g.checkWarnings(dslMap)

	return result, nil
}

// extractCodeBlock extracts content from a markdown code block
func extractCodeBlock(content, lang string) string {
	// Look for ```json or ```
	startMarkers := []string{"```" + lang, "```"}

	for _, startMarker := range startMarkers {
		start := strings.Index(content, startMarker)
		if start == -1 {
			continue
		}

		// Move past the marker and newline
		start += len(startMarker)
		for start < len(content) && (content[start] == '\n' || content[start] == '\r') {
			start++
		}

		// Find the closing ```
		end := strings.Index(content[start:], "```")
		if end == -1 {
			continue
		}

		return strings.TrimSpace(content[start : start+end])
	}

	return ""
}

// extractJSON extracts a JSON object from content (same as in intent.go)
func extractJSON(content string) string {
	start := strings.Index(content, "{")
	if start == -1 {
		return ""
	}

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

// extractExplanation extracts the explanation text from the response
func extractExplanation(content, dslJSON string) string {
	// Find where the JSON ends
	jsonEnd := strings.Index(content, dslJSON)
	if jsonEnd == -1 {
		return ""
	}
	jsonEnd += len(dslJSON)

	// Skip past closing code block if present
	remaining := content[jsonEnd:]
	if idx := strings.Index(remaining, "```"); idx != -1 {
		remaining = remaining[idx+3:]
	}

	return strings.TrimSpace(remaining)
}

// checkWarnings checks for potential issues in the generated DSL
func (g *DSLGenerator) checkWarnings(dsl map[string]interface{}) []string {
	var warnings []string

	// Check for missing event_id
	if _, ok := dsl["event_id"]; !ok {
		warnings = append(warnings, "Missing event_id")
	}

	// Check for empty pricing_rules
	if rules, ok := dsl["pricing_rules"].([]interface{}); ok && len(rules) == 0 {
		warnings = append(warnings, "No pricing rules defined")
	}

	// Check for discount_stacking when there are multiple discount rules
	if rules, ok := dsl["pricing_rules"].([]interface{}); ok {
		discountCount := 0
		for _, rule := range rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				if action, ok := ruleMap["action"].(map[string]interface{}); ok {
					actionType, _ := action["type"].(string)
					if strings.Contains(actionType, "discount") {
						discountCount++
					}
				}
			}
		}
		if discountCount > 1 {
			if _, hasStacking := dsl["discount_stacking"]; !hasStacking {
				warnings = append(warnings, "Multiple discounts defined but no discount_stacking mode specified")
			}
		}
	}

	return warnings
}

/**
 * GenerateSimple generates a basic DSL structure without LLM.
 * Useful for testing or when LLM is not available.
 */
func (g *DSLGenerator) GenerateSimple(req GenerationRequest) (*GenerationResult, error) {
	// Set defaults
	if req.EventID == "" {
		req.EventID = "event-001"
	}
	if req.EventName == "" {
		req.EventName = "Event"
	}
	if req.Version == "" {
		req.Version = "1.0"
	}

	dsl := map[string]interface{}{
		"event_id": req.EventID,
		"version":  req.Version,
		"name":     req.EventName,
		"variables": map[string]interface{}{
			"base_price": 1000,
		},
		"pricing_rules":    []interface{}{},
		"validation_rules": []interface{}{},
	}

	// Convert confirmed rules to DSL rules
	pricingRules := []interface{}{}
	validationRules := []interface{}{}
	variables := dsl["variables"].(map[string]interface{})

	for i, rule := range req.ConfirmedRules {
		switch rule.Type {
		case RuleTypePricing, RuleTypeDiscount, RuleTypeAddon:
			pricingRule := g.buildPricingRule(rule, i, variables)
			if pricingRule != nil {
				pricingRules = append(pricingRules, pricingRule)
			}
		case RuleTypeValidation:
			validationRule := g.buildValidationRule(rule, i)
			if validationRule != nil {
				validationRules = append(validationRules, validationRule)
			}
		}
	}

	dsl["pricing_rules"] = pricingRules
	dsl["validation_rules"] = validationRules

	// Add discount stacking if multiple discounts
	discountCount := 0
	for _, rule := range req.ConfirmedRules {
		if rule.Type == RuleTypeDiscount {
			discountCount++
		}
	}
	if discountCount > 1 {
		dsl["discount_stacking"] = map[string]interface{}{
			"mode":        "multiplicative",
			"description": "折扣連乘計算",
		}
	}

	jsonBytes, err := json.MarshalIndent(dsl, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DSL: %w", err)
	}

	return &GenerationResult{
		DSL:         json.RawMessage(jsonBytes),
		Explanation: "DSL generated from confirmed rules (simple mode)",
	}, nil
}

// buildPricingRule builds a pricing rule from a confirmed rule
func (g *DSLGenerator) buildPricingRule(rule ConfirmedRule, index int, variables map[string]interface{}) map[string]interface{} {
	priority := index * 10

	switch rule.Type {
	case RuleTypePricing:
		price, _ := rule.Details["price"].(float64)
		if price == 0 {
			price = 1000
		}
		varName := "base_price"
		variables[varName] = price

		return map[string]interface{}{
			"id":          fmt.Sprintf("pricing_%d", index),
			"priority":    priority,
			"description": rule.Description,
			"condition": map[string]interface{}{
				"type": "always_true",
			},
			"action": map[string]interface{}{
				"type":  "set_price",
				"item":  "registration_fee",
				"value": fmt.Sprintf("$variables.%s", varName),
				"label": rule.Description,
			},
		}

	case RuleTypeDiscount:
		discount, _ := rule.Details["discount"].(float64)
		if discount == 0 {
			discount = 10
		}

		condition := map[string]interface{}{
			"type": "always_true",
		}

		// Check for early bird
		if deadline, ok := rule.Details["deadline"].(string); ok {
			condition = map[string]interface{}{
				"type":  "datetime_before",
				"field": "register_date",
				"value": deadline,
			}
		}

		// Check for team discount
		if minSize, ok := rule.Details["min_size"].(float64); ok {
			condition = map[string]interface{}{
				"type":     "compare",
				"field":    "team_size",
				"operator": "gte",
				"value":    minSize,
			}
		}

		return map[string]interface{}{
			"id":          fmt.Sprintf("discount_%d", index),
			"priority":    100 + priority,
			"description": rule.Description,
			"condition":   condition,
			"action": map[string]interface{}{
				"type":     "percentage_discount",
				"value":    discount,
				"apply_to": []string{"registration_fee"},
				"label":    rule.Description,
			},
		}

	case RuleTypeAddon:
		price, _ := rule.Details["price"].(float64)
		item, _ := rule.Details["item"].(string)
		if item == "" {
			item = "addon"
		}

		return map[string]interface{}{
			"id":          fmt.Sprintf("addon_%d", index),
			"priority":    50 + priority,
			"description": rule.Description,
			"condition": map[string]interface{}{
				"type":  "equals",
				"field": fmt.Sprintf("addons.%s", item),
				"value": true,
			},
			"action": map[string]interface{}{
				"type":       "add_item",
				"item":       fmt.Sprintf("addon:%s", item),
				"unit_price": price,
				"label":      rule.Description,
			},
		}
	}

	return nil
}

// buildValidationRule builds a validation rule from a confirmed rule
func (g *DSLGenerator) buildValidationRule(rule ConfirmedRule, index int) map[string]interface{} {
	errorType := "blocking"
	if t, ok := rule.Details["error_type"].(string); ok {
		errorType = t
	}

	condition := map[string]interface{}{
		"type": "always_true",
	}

	// Check for age limit
	if minAge, ok := rule.Details["min_age"].(float64); ok {
		condition = map[string]interface{}{
			"type":     "compare",
			"field":    "user.age",
			"operator": "lt",
			"value":    minAge,
		}
	}

	// Check for deadline
	if deadline, ok := rule.Details["deadline"].(string); ok {
		condition = map[string]interface{}{
			"type":  "datetime_after",
			"field": "register_date",
			"value": deadline,
		}
	}

	errorMessage, _ := rule.Details["error_message"].(string)
	if errorMessage == "" {
		errorMessage = rule.Description
	}

	return map[string]interface{}{
		"id":            fmt.Sprintf("validation_%d", index),
		"description":   rule.Description,
		"condition":     condition,
		"error_type":    errorType,
		"error_message": errorMessage,
	}
}
