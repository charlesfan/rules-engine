package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
	"github.com/charlesfan/rules-engine/internal/rules/parser"
)

/**
 * DSLValidator validates generated DSL against the Rule Engine schema.
 * It uses the existing parser from the rules package to ensure
 * the generated DSL is valid and can be executed.
 */

// DSLValidator validates DSL against the Rule Engine
type DSLValidator struct {
	parser *parser.Parser
}

// NewDSLValidator creates a new DSL validator
func NewDSLValidator() *DSLValidator {
	return &DSLValidator{
		parser: parser.NewParser(),
	}
}

// ValidationResult contains the result of DSL validation
type ValidationResult struct {
	Valid    bool               `json:"valid"`
	RuleSet  *dsl.RuleSet       `json:"rule_set,omitempty"`
	Errors   []ValidationError  `json:"errors,omitempty"`
	Warnings []string           `json:"warnings,omitempty"`
	Summary  *ValidationSummary `json:"summary,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Type    string `json:"type"` // syntax, semantic, reference
}

// ValidationSummary summarizes the validated DSL
type ValidationSummary struct {
	PricingRuleCount    int      `json:"pricing_rule_count"`
	ValidationRuleCount int      `json:"validation_rule_count"`
	VariableCount       int      `json:"variable_count"`
	RuleDefCount        int      `json:"rule_def_count"`
	ComputedFieldCount  int      `json:"computed_field_count"`
	UsedActionTypes     []string `json:"used_action_types"`
	UsedConditionTypes  []string `json:"used_condition_types"`
}

/**
 * Validate validates DSL JSON against the Rule Engine schema.
 * Returns detailed validation result with errors and warnings.
 */
func (v *DSLValidator) Validate(dslJSON []byte) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Step 1: Check JSON syntax
	var rawDSL map[string]interface{}
	if err := json.Unmarshal(dslJSON, &rawDSL); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "root",
			Message: fmt.Sprintf("Invalid JSON syntax: %v", err),
			Type:    "syntax",
		})
		return result
	}

	// Step 2: Check required fields
	requiredFields := []string{"event_id"}
	for _, field := range requiredFields {
		if _, ok := rawDSL[field]; !ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Required field '%s' is missing", field),
				Type:    "semantic",
			})
		}
	}

	// Step 3: Use Rule Engine parser for full validation
	ruleSet, err := v.parser.Parse(dslJSON)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "root",
			Message: err.Error(),
			Type:    "semantic",
		})
		return result
	}

	result.RuleSet = ruleSet

	// Step 4: Additional semantic checks
	v.checkSemantics(ruleSet, result)

	// Step 5: Generate summary
	result.Summary = v.generateSummary(ruleSet)

	return result
}

// ValidateString validates DSL from a string
func (v *DSLValidator) ValidateString(dslStr string) *ValidationResult {
	return v.Validate([]byte(dslStr))
}

// checkSemantics performs additional semantic checks
func (v *DSLValidator) checkSemantics(ruleSet *dsl.RuleSet, result *ValidationResult) {
	// Check for duplicate rule IDs
	seenIDs := make(map[string]bool)

	for _, rule := range ruleSet.PricingRules {
		if seenIDs[rule.ID] {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("pricing_rules.%s", rule.ID),
				Message: fmt.Sprintf("Duplicate rule ID: %s", rule.ID),
				Type:    "semantic",
			})
			result.Valid = false
		}
		seenIDs[rule.ID] = true

		// Check action validity
		v.checkAction(rule.Action, rule.ID, result)
	}

	for _, rule := range ruleSet.ValidationRules {
		if seenIDs[rule.ID] {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("validation_rules.%s", rule.ID),
				Message: fmt.Sprintf("Duplicate rule ID: %s", rule.ID),
				Type:    "semantic",
			})
			result.Valid = false
		}
		seenIDs[rule.ID] = true

		// Check error type
		if rule.ErrorType != "blocking" && rule.ErrorType != "warning" {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Rule %s: Unknown error_type '%s', should be 'blocking' or 'warning'",
					rule.ID, rule.ErrorType))
		}
	}

	// Check for priority conflicts
	v.checkPriorityConflicts(ruleSet.PricingRules, result)

	// Check variable references
	v.checkVariableReferences(ruleSet, result)

	// Check for missing discount stacking
	discountCount := 0
	for _, rule := range ruleSet.PricingRules {
		if rule.Action != nil && strings.Contains(rule.Action.Type, "discount") {
			discountCount++
		}
	}
	if discountCount > 1 && ruleSet.DiscountStacking == nil {
		result.Warnings = append(result.Warnings,
			"Multiple discount rules found but no discount_stacking mode specified. Defaulting to 'multiplicative'.")
	}
}

// checkAction validates an action
func (v *DSLValidator) checkAction(action *dsl.Action, ruleID string, result *ValidationResult) {
	if action == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   fmt.Sprintf("pricing_rules.%s.action", ruleID),
			Message: "Action is required for pricing rules",
			Type:    "semantic",
		})
		result.Valid = false
		return
	}

	validActionTypes := map[string]bool{
		"set_price":           true,
		"add_item":            true,
		"percentage_discount": true,
		"fixed_discount":      true,
		"price_cap":           true,
		"replace_price":       true,
	}

	if !validActionTypes[action.Type] {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Rule %s: Unknown action type '%s'", ruleID, action.Type))
	}

	// Check discount rules have apply_to
	if strings.Contains(action.Type, "discount") && len(action.ApplyTo) == 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Rule %s: Discount action without apply_to will apply to all items", ruleID))
	}
}

// checkPriorityConflicts checks for rules with the same priority
func (v *DSLValidator) checkPriorityConflicts(rules []*dsl.PricingRule, result *ValidationResult) {
	priorityMap := make(map[int][]string)

	for _, rule := range rules {
		priorityMap[rule.Priority] = append(priorityMap[rule.Priority], rule.ID)
	}

	for priority, ruleIDs := range priorityMap {
		if len(ruleIDs) > 1 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Multiple rules with same priority %d: %v. Execution order may be unpredictable.",
					priority, ruleIDs))
		}
	}
}

// checkVariableReferences checks that all variable references exist
func (v *DSLValidator) checkVariableReferences(ruleSet *dsl.RuleSet, result *ValidationResult) {
	// Collect all variable references from actions
	for _, rule := range ruleSet.PricingRules {
		if rule.Action == nil {
			continue
		}

		// Check Value field
		if strVal, ok := rule.Action.Value.(string); ok && strings.HasPrefix(strVal, "$variables.") {
			varName := strings.TrimPrefix(strVal, "$variables.")
			if ruleSet.Variables == nil || ruleSet.Variables[varName] == nil {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("pricing_rules.%s.action.value", rule.ID),
					Message: fmt.Sprintf("Referenced variable '%s' is not defined", varName),
					Type:    "reference",
				})
				result.Valid = false
			}
		}

		// Check UnitPrice field
		if strVal, ok := rule.Action.UnitPrice.(string); ok && strings.HasPrefix(strVal, "$variables.") {
			varName := strings.TrimPrefix(strVal, "$variables.")
			if ruleSet.Variables == nil || ruleSet.Variables[varName] == nil {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("pricing_rules.%s.action.unit_price", rule.ID),
					Message: fmt.Sprintf("Referenced variable '%s' is not defined", varName),
					Type:    "reference",
				})
				result.Valid = false
			}
		}
	}
}

// generateSummary generates a summary of the validated DSL
func (v *DSLValidator) generateSummary(ruleSet *dsl.RuleSet) *ValidationSummary {
	summary := &ValidationSummary{
		PricingRuleCount:    len(ruleSet.PricingRules),
		ValidationRuleCount: len(ruleSet.ValidationRules),
		UsedActionTypes:     []string{},
		UsedConditionTypes:  []string{},
	}

	if ruleSet.Variables != nil {
		summary.VariableCount = len(ruleSet.Variables)
	}
	if ruleSet.RuleDefs != nil {
		summary.RuleDefCount = len(ruleSet.RuleDefs)
	}
	if ruleSet.ComputedFields != nil {
		summary.ComputedFieldCount = len(ruleSet.ComputedFields)
	}

	// Collect unique action types
	actionTypes := make(map[string]bool)
	for _, rule := range ruleSet.PricingRules {
		if rule.Action != nil {
			actionTypes[rule.Action.Type] = true
		}
	}
	for t := range actionTypes {
		summary.UsedActionTypes = append(summary.UsedActionTypes, t)
	}

	// Collect unique condition types
	conditionTypes := make(map[string]bool)
	for _, rule := range ruleSet.PricingRules {
		v.collectConditionTypes(rule.Condition, conditionTypes)
	}
	for _, rule := range ruleSet.ValidationRules {
		v.collectConditionTypes(rule.Condition, conditionTypes)
	}
	for t := range conditionTypes {
		summary.UsedConditionTypes = append(summary.UsedConditionTypes, t)
	}

	return summary
}

// collectConditionTypes recursively collects condition types
func (v *DSLValidator) collectConditionTypes(expr *dsl.Expression, types map[string]bool) {
	if expr == nil {
		return
	}

	types[expr.Type] = true

	// Recurse into nested conditions
	for _, cond := range expr.Conditions {
		v.collectConditionTypes(cond, types)
	}
	if expr.Condition != nil {
		v.collectConditionTypes(expr.Condition, types)
	}
}

// FormatErrors formats validation errors for display
func (r *ValidationResult) FormatErrors() string {
	if len(r.Errors) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("驗證錯誤:\n")
	for i, err := range r.Errors {
		sb.WriteString(fmt.Sprintf("  %d. [%s] %s: %s\n", i+1, err.Type, err.Field, err.Message))
	}
	return sb.String()
}

// FormatWarnings formats warnings for display
func (r *ValidationResult) FormatWarnings() string {
	if len(r.Warnings) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("警告:\n")
	for i, warn := range r.Warnings {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, warn))
	}
	return sb.String()
}

// FormatSummary formats the summary for display
func (r *ValidationResult) FormatSummary() string {
	if r.Summary == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("DSL 摘要:\n")
	sb.WriteString(fmt.Sprintf("  定價規則: %d 條\n", r.Summary.PricingRuleCount))
	sb.WriteString(fmt.Sprintf("  驗證規則: %d 條\n", r.Summary.ValidationRuleCount))
	sb.WriteString(fmt.Sprintf("  變數: %d 個\n", r.Summary.VariableCount))
	if r.Summary.RuleDefCount > 0 {
		sb.WriteString(fmt.Sprintf("  規則定義: %d 個\n", r.Summary.RuleDefCount))
	}
	if r.Summary.ComputedFieldCount > 0 {
		sb.WriteString(fmt.Sprintf("  計算欄位: %d 個\n", r.Summary.ComputedFieldCount))
	}
	if len(r.Summary.UsedActionTypes) > 0 {
		sb.WriteString(fmt.Sprintf("  使用的動作類型: %v\n", r.Summary.UsedActionTypes))
	}
	if len(r.Summary.UsedConditionTypes) > 0 {
		sb.WriteString(fmt.Sprintf("  使用的條件類型: %v\n", r.Summary.UsedConditionTypes))
	}
	return sb.String()
}
