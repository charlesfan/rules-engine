package parser

import (
	"encoding/json"
	"fmt"

	"github.com/charlesfan/rules-engine/internal/rules/dsl"
)

// Parser DSL 解析器
type Parser struct{}

// NewParser 建立新的 Parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse 解析 JSON 格式的 DSL
func (p *Parser) Parse(data []byte) (*dsl.RuleSet, error) {
	var ruleSet dsl.RuleSet
	if err := json.Unmarshal(data, &ruleSet); err != nil {
		return nil, fmt.Errorf("failed to parse DSL: %w", err)
	}

	// 驗證規則集
	if err := p.Validate(&ruleSet); err != nil {
		return nil, fmt.Errorf("invalid rule set: %w", err)
	}

	return &ruleSet, nil
}

// Validate 驗證規則集的有效性
func (p *Parser) Validate(ruleSet *dsl.RuleSet) error {
	if ruleSet.EventID == "" {
		return fmt.Errorf("event_id is required")
	}

	// 驗證 rule definitions 引用
	for _, pricingRule := range ruleSet.PricingRules {
		if err := p.validateExpression(pricingRule.Condition, ruleSet.RuleDefs); err != nil {
			return fmt.Errorf("invalid pricing rule %s: %w", pricingRule.ID, err)
		}
	}

	for _, validationRule := range ruleSet.ValidationRules {
		if err := p.validateExpression(validationRule.Condition, ruleSet.RuleDefs); err != nil {
			return fmt.Errorf("invalid validation rule %s: %w", validationRule.ID, err)
		}
	}

	return nil
}

// validateExpression 驗證表達式
func (p *Parser) validateExpression(expr *dsl.Expression, ruleDefs map[string]*dsl.RuleDef) error {
	if expr == nil {
		return fmt.Errorf("expression cannot be nil")
	}

	switch expr.Type {
	case "rule_ref":
		// 檢查引用的規則是否存在
		if expr.Rule == "" {
			return fmt.Errorf("rule_ref requires rule field")
		}
		if ruleDefs != nil {
			if _, exists := ruleDefs[expr.Rule]; !exists {
				return fmt.Errorf("referenced rule '%s' not found", expr.Rule)
			}
		}

	case "and", "or":
		if len(expr.Conditions) == 0 {
			return fmt.Errorf("%s requires at least one condition", expr.Type)
		}
		for i, cond := range expr.Conditions {
			if err := p.validateExpression(cond, ruleDefs); err != nil {
				return fmt.Errorf("%s condition %d: %w", expr.Type, i, err)
			}
		}

	case "not":
		if expr.Condition == nil {
			return fmt.Errorf("not requires a condition")
		}
		return p.validateExpression(expr.Condition, ruleDefs)

	case "array_any", "array_all", "array_count", "array_sum":
		if expr.Array == "" {
			return fmt.Errorf("%s requires array field", expr.Type)
		}
		if expr.Condition != nil {
			return p.validateExpression(expr.Condition, ruleDefs)
		}

	case "equals", "compare", "datetime_before", "datetime_after":
		if expr.Field == "" {
			return fmt.Errorf("%s requires field", expr.Type)
		}

	case "datetime_between":
		if expr.Field == "" {
			return fmt.Errorf("datetime_between requires field")
		}
		if expr.Start == nil || expr.End == nil {
			return fmt.Errorf("datetime_between requires start and end")
		}

	case "in_list":
		if expr.Field == "" || expr.List == "" {
			return fmt.Errorf("in_list requires field and list")
		}

	case "always_true", "field_exists", "field_empty":
		// 這些類型不需要額外驗證

	default:
		return fmt.Errorf("unknown expression type: %s", expr.Type)
	}

	return nil
}
