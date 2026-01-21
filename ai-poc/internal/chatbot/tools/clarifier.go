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
 * Clarifier generates clarification questions when rule information is incomplete.
 * It uses both pattern-based analysis and LLM for generating questions.
 */

// Clarifier generates clarification questions
type Clarifier struct {
	provider llm.Provider
}

// NewClarifier creates a new clarifier
func NewClarifier(provider llm.Provider) *Clarifier {
	return &Clarifier{
		provider: provider,
	}
}

// ClarificationQuestion represents a question to ask the user
type ClarificationQuestion struct {
	ID                string   `json:"id"`
	Priority          Priority `json:"priority"`
	Question          string   `json:"question"`
	DefaultSuggestion string   `json:"default_suggestion,omitempty"`
	Affects           string   `json:"affects"`
	Options           []string `json:"options,omitempty"`
}

// Priority represents the priority of a clarification question
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// ClarificationResult contains clarification questions and analysis
type ClarificationResult struct {
	Questions   []ClarificationQuestion `json:"questions"`
	CanProceed  bool                    `json:"can_proceed"`
	Reason      string                  `json:"reason"`
	Assumptions []string                `json:"assumptions,omitempty"`
}

// ExtractedRuleInfo represents information extracted from user input
type ExtractedRuleInfo struct {
	Type        RuleType               `json:"type"`
	Description string                 `json:"description"`
	Complete    bool                   `json:"complete"`
	Details     map[string]interface{} `json:"details"`
	Missing     []string               `json:"missing"`
}

/**
 * Analyze analyzes extracted rules and generates clarification questions.
 * Uses pattern-based analysis for common cases and LLM for complex cases.
 */
func (c *Clarifier) Analyze(ctx context.Context, userInput string, extractedRules []ExtractedRuleInfo) (*ClarificationResult, error) {
	// First, do pattern-based analysis
	result := c.patternBasedAnalysis(extractedRules)

	// If we have LLM and there are complex cases, enhance with LLM
	if c.provider != nil && len(result.Questions) < 3 {
		llmResult, err := c.llmAnalysis(ctx, userInput, extractedRules)
		if err == nil && llmResult != nil {
			// Merge LLM questions with pattern-based questions
			result = c.mergeResults(result, llmResult)
		}
	}

	return result, nil
}

// patternBasedAnalysis generates questions using pattern matching
func (c *Clarifier) patternBasedAnalysis(extractedRules []ExtractedRuleInfo) *ClarificationResult {
	result := &ClarificationResult{
		Questions:  []ClarificationQuestion{},
		CanProceed: true,
	}

	hasPricing := false
	hasDiscount := false
	discountCount := 0

	for _, rule := range extractedRules {
		switch rule.Type {
		case RuleTypePricing:
			hasPricing = true
			c.analyzePricingRule(rule, result)
		case RuleTypeDiscount:
			hasDiscount = true
			discountCount++
			c.analyzeDiscountRule(rule, result)
		case RuleTypeValidation:
			c.analyzeValidationRule(rule, result)
		case RuleTypeAddon:
			c.analyzeAddonRule(rule, result)
		}
	}

	// Common questions based on overall context
	if !hasPricing {
		result.Questions = append(result.Questions, ClarificationQuestion{
			ID:                "missing_base_price",
			Priority:          PriorityHigh,
			Question:          "請問基本報名費是多少？",
			DefaultSuggestion: "1000 元",
			Affects:           "pricing",
		})
		result.CanProceed = false
	}

	// Ask about discount stacking if multiple discounts
	if discountCount > 1 {
		result.Questions = append(result.Questions, ClarificationQuestion{
			ID:                "discount_stacking",
			Priority:          PriorityMedium,
			Question:          "有多個折扣，折扣可以疊加嗎？疊加方式是？",
			DefaultSuggestion: "連乘計算（如 9 折再 95 折 = 85.5 折）",
			Affects:           "discount",
			Options:           []string{"連乘計算", "相加計算", "只取最優惠的一個"},
		})
	}

	// Suggest common features that might be missing
	if hasPricing && !hasDiscount {
		result.Questions = append(result.Questions, ClarificationQuestion{
			ID:                "suggest_discount",
			Priority:          PriorityLow,
			Question:          "是否有任何折扣優惠？（如早鳥優惠、團報優惠等）",
			DefaultSuggestion: "無",
			Affects:           "discount",
		})
	}

	if len(result.Questions) == 0 {
		result.Reason = "規則資訊完整，可以生成 DSL"
	} else {
		result.Reason = fmt.Sprintf("有 %d 個問題需要確認", len(result.Questions))
		// Only block if there are high priority questions
		result.CanProceed = true
		for _, q := range result.Questions {
			if q.Priority == PriorityHigh {
				result.CanProceed = false
				break
			}
		}
	}

	return result
}

// analyzePricingRule analyzes a pricing rule for missing information
func (c *Clarifier) analyzePricingRule(rule ExtractedRuleInfo, result *ClarificationResult) {
	// Check for price
	if _, ok := rule.Details["price"]; !ok {
		result.Questions = append(result.Questions, ClarificationQuestion{
			ID:       fmt.Sprintf("pricing_price_%s", rule.Description),
			Priority: PriorityHigh,
			Question: fmt.Sprintf("「%s」的價格是多少？", rule.Description),
			Affects:  "pricing",
		})
	}

	// Check if there might be different categories
	desc := strings.ToLower(rule.Description)
	if strings.Contains(desc, "馬拉松") || strings.Contains(desc, "路跑") {
		if _, ok := rule.Details["categories"]; !ok {
			result.Questions = append(result.Questions, ClarificationQuestion{
				ID:                "pricing_categories",
				Priority:          PriorityMedium,
				Question:          "是否有不同組別（如全馬、半馬、10K）？各組報名費是否不同？",
				DefaultSuggestion: "單一組別",
				Affects:           "pricing",
			})
		}
	}
}

// analyzeDiscountRule analyzes a discount rule for missing information
func (c *Clarifier) analyzeDiscountRule(rule ExtractedRuleInfo, result *ClarificationResult) {
	desc := strings.ToLower(rule.Description)

	// Early bird discount
	if strings.Contains(desc, "早鳥") {
		if _, ok := rule.Details["deadline"]; !ok {
			result.Questions = append(result.Questions, ClarificationQuestion{
				ID:       "early_bird_deadline",
				Priority: PriorityHigh,
				Question: "早鳥優惠的截止日期是？",
				Affects:  "discount",
			})
		}
		if _, ok := rule.Details["discount"]; !ok {
			result.Questions = append(result.Questions, ClarificationQuestion{
				ID:                "early_bird_discount",
				Priority:          PriorityHigh,
				Question:          "早鳥優惠的折扣是多少？",
				DefaultSuggestion: "9 折",
				Affects:           "discount",
			})
		}
	}

	// Team discount
	if strings.Contains(desc, "團報") || strings.Contains(desc, "團體") {
		if _, ok := rule.Details["min_size"]; !ok {
			result.Questions = append(result.Questions, ClarificationQuestion{
				ID:                "team_min_size",
				Priority:          PriorityHigh,
				Question:          "團報優惠的最低人數是？",
				DefaultSuggestion: "5 人",
				Affects:           "discount",
			})
		}
		if _, ok := rule.Details["discount"]; !ok {
			result.Questions = append(result.Questions, ClarificationQuestion{
				ID:                "team_discount",
				Priority:          PriorityHigh,
				Question:          "團報優惠的折扣是多少？",
				DefaultSuggestion: "95 折",
				Affects:           "discount",
			})
		}
	}
}

// analyzeValidationRule analyzes a validation rule for missing information
func (c *Clarifier) analyzeValidationRule(rule ExtractedRuleInfo, result *ClarificationResult) {
	desc := strings.ToLower(rule.Description)

	// Age limit
	if strings.Contains(desc, "年齡") || strings.Contains(desc, "歲") {
		if _, ok := rule.Details["min_age"]; !ok {
			result.Questions = append(result.Questions, ClarificationQuestion{
				ID:                "validation_min_age",
				Priority:          PriorityMedium,
				Question:          "最低年齡限制是多少歲？",
				DefaultSuggestion: "18 歲",
				Affects:           "validation",
			})
		}
	}

	// Deadline
	if strings.Contains(desc, "截止") || strings.Contains(desc, "期限") {
		if _, ok := rule.Details["deadline"]; !ok {
			result.Questions = append(result.Questions, ClarificationQuestion{
				ID:       "validation_deadline",
				Priority: PriorityMedium,
				Question: "報名截止日期是？",
				Affects:  "validation",
			})
		}
	}
}

// analyzeAddonRule analyzes an addon rule for missing information
func (c *Clarifier) analyzeAddonRule(rule ExtractedRuleInfo, result *ClarificationResult) {
	if _, ok := rule.Details["price"]; !ok {
		result.Questions = append(result.Questions, ClarificationQuestion{
			ID:       fmt.Sprintf("addon_price_%s", rule.Description),
			Priority: PriorityMedium,
			Question: fmt.Sprintf("「%s」的價格是多少？", rule.Description),
			Affects:  "addon",
		})
	}
}

// llmAnalysis uses LLM for more sophisticated analysis
func (c *Clarifier) llmAnalysis(ctx context.Context, userInput string, extractedRules []ExtractedRuleInfo) (*ClarificationResult, error) {
	// Format extracted info
	var infoBuilder strings.Builder
	for i, rule := range extractedRules {
		infoBuilder.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, rule.Type, rule.Description))
		if len(rule.Details) > 0 {
			detailsJSON, _ := json.MarshalIndent(rule.Details, "   ", "  ")
			infoBuilder.WriteString(fmt.Sprintf("   詳細: %s\n", string(detailsJSON)))
		}
		if len(rule.Missing) > 0 {
			infoBuilder.WriteString(fmt.Sprintf("   缺少: %v\n", rule.Missing))
		}
	}

	prompt := prompts.ClarificationPrompt
	prompt = strings.Replace(prompt, "{{USER_INPUT}}", userInput, 1)
	prompt = strings.Replace(prompt, "{{EXTRACTED_INFO}}", infoBuilder.String(), 1)

	req := llm.ChatRequest{
		Messages: []llm.Message{
			llm.NewUserMessage(prompt),
		},
	}

	resp, err := c.provider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	return c.parseLLMResponse(resp.Content)
}

// parseLLMResponse parses the LLM response
func (c *Clarifier) parseLLMResponse(content string) (*ClarificationResult, error) {
	// Extract JSON from response
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var rawResult struct {
		Questions []struct {
			Priority          string `json:"priority"`
			Question          string `json:"question"`
			DefaultSuggestion string `json:"default_suggestion"`
			Affects           string `json:"affects"`
		} `json:"questions"`
		CanProceed bool   `json:"can_proceed"`
		Reason     string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawResult); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result := &ClarificationResult{
		Questions:  make([]ClarificationQuestion, 0, len(rawResult.Questions)),
		CanProceed: rawResult.CanProceed,
		Reason:     rawResult.Reason,
	}

	for i, q := range rawResult.Questions {
		priority := PriorityMedium
		switch q.Priority {
		case "high":
			priority = PriorityHigh
		case "low":
			priority = PriorityLow
		}

		result.Questions = append(result.Questions, ClarificationQuestion{
			ID:                fmt.Sprintf("llm_q_%d", i),
			Priority:          priority,
			Question:          q.Question,
			DefaultSuggestion: q.DefaultSuggestion,
			Affects:           q.Affects,
		})
	}

	return result, nil
}

// mergeResults merges pattern-based and LLM results
func (c *Clarifier) mergeResults(patternResult, llmResult *ClarificationResult) *ClarificationResult {
	// Start with pattern result
	result := patternResult

	// Add LLM questions that aren't duplicates
	seenQuestions := make(map[string]bool)
	for _, q := range result.Questions {
		seenQuestions[q.Question] = true
	}

	for _, q := range llmResult.Questions {
		if !seenQuestions[q.Question] {
			result.Questions = append(result.Questions, q)
		}
	}

	// Add assumptions from LLM
	result.Assumptions = llmResult.Assumptions

	// Update CanProceed based on merged questions
	result.CanProceed = true
	for _, q := range result.Questions {
		if q.Priority == PriorityHigh {
			result.CanProceed = false
			break
		}
	}

	return result
}

// FormatQuestions formats questions for display
func (r *ClarificationResult) FormatQuestions() string {
	if len(r.Questions) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("需要確認的問題:\n")

	// Group by priority
	high := []ClarificationQuestion{}
	medium := []ClarificationQuestion{}
	low := []ClarificationQuestion{}

	for _, q := range r.Questions {
		switch q.Priority {
		case PriorityHigh:
			high = append(high, q)
		case PriorityMedium:
			medium = append(medium, q)
		case PriorityLow:
			low = append(low, q)
		}
	}

	formatQuestion := func(q ClarificationQuestion) string {
		s := fmt.Sprintf("  • %s", q.Question)
		if q.DefaultSuggestion != "" {
			s += fmt.Sprintf("\n    （建議: %s）", q.DefaultSuggestion)
		}
		if len(q.Options) > 0 {
			s += fmt.Sprintf("\n    選項: %v", q.Options)
		}
		return s
	}

	if len(high) > 0 {
		sb.WriteString("\n❗ 必須確認:\n")
		for _, q := range high {
			sb.WriteString(formatQuestion(q) + "\n")
		}
	}

	if len(medium) > 0 {
		sb.WriteString("\n❓ 建議確認:\n")
		for _, q := range medium {
			sb.WriteString(formatQuestion(q) + "\n")
		}
	}

	if len(low) > 0 {
		sb.WriteString("\n💡 可選確認:\n")
		for _, q := range low {
			sb.WriteString(formatQuestion(q) + "\n")
		}
	}

	return sb.String()
}

// GetHighPriorityQuestions returns only high priority questions
func (r *ClarificationResult) GetHighPriorityQuestions() []ClarificationQuestion {
	result := []ClarificationQuestion{}
	for _, q := range r.Questions {
		if q.Priority == PriorityHigh {
			result = append(result, q)
		}
	}
	return result
}
