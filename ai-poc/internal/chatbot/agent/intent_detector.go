package agent

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/charlesfan/rules-engine/ai-poc/internal/chatbot/prompts"
	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

// DetectedIntent represents the result of intent detection
type DetectedIntent struct {
	Intent string
	Raw    string // Raw LLM response for debugging
}

// detectIntent performs Stage 1 intent detection
func (a *Agent) detectIntent(ctx context.Context, input string) (*DetectedIntent, error) {
	// Build intent detection prompt
	hasPendingQuestions := len(a.pendingQuestions) > 0
	prompt := prompts.BuildIntentPrompt(
		input,
		a.state.String(),
		hasPendingQuestions,
		a.getRuleCount(),
		a.conversationContext,
	)

	// Call LLM
	req := llm.ChatRequest{
		Messages: []llm.Message{
			llm.NewUserMessage(prompt),
		},
		JSONMode: true,
	}

	resp, err := a.provider.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse response
	intent, err := parseIntentResponse(resp.Content)
	if err != nil {
		// Try fallback keyword-based detection
		intent = detectIntentByKeywords(input, hasPendingQuestions)
	}

	return &DetectedIntent{
		Intent: intent,
		Raw:    resp.Content,
	}, nil
}

// parseIntentResponse parses the LLM response for intent
func parseIntentResponse(content string) (string, error) {
	// Extract JSON from content
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return "", &ParseError{Message: "no JSON found", Content: content}
	}

	var resp prompts.IntentResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return "", &ParseError{Message: err.Error(), Content: jsonStr}
	}

	// Normalize intent
	return normalizeIntent(resp.Intent), nil
}

// detectIntentByKeywords performs keyword-based intent detection as fallback
func detectIntentByKeywords(input string, hasPendingQuestions bool) string {
	inputLower := strings.ToLower(input)

	// Check for DSL request keywords first (highest priority)
	dslKeywords := []string{"生成", "產生", "完成", "確認", "dsl", "規則"}
	for _, kw := range dslKeywords {
		if strings.Contains(inputLower, kw) {
			// More specific check for DSL generation
			if strings.Contains(inputLower, "生成") || strings.Contains(inputLower, "產生") {
				return IntentDSLRequest
			}
		}
	}

	// If there are pending questions, user is likely responding
	if hasPendingQuestions {
		// Check if it's a simple acknowledgment or answer
		ackKeywords := []string{"是", "對", "好", "沒有", "可以", "不用", "不是", "ok", "yes", "no"}
		for _, kw := range ackKeywords {
			if strings.Contains(inputLower, kw) {
				return IntentClarifyResponse
			}
		}
		// If user provides a value (date, number, etc.), treat as clarify
		if containsValueLikeContent(inputLower) {
			return IntentClarifyResponse
		}
	}

	// Check for delete keywords
	deleteKeywords := []string{"刪除", "移除", "取消", "不要"}
	for _, kw := range deleteKeywords {
		if strings.Contains(inputLower, kw) {
			return IntentDeleteRule
		}
	}

	// Check for modify keywords
	modifyKeywords := []string{"改", "修改", "調整", "更新", "換成", "改成"}
	for _, kw := range modifyKeywords {
		if strings.Contains(inputLower, kw) {
			return IntentModifyRule
		}
	}

	// Check for rule-related keywords
	ruleKeywords := []string{
		"報名費", "費用", "價格", "元", "nt$", "nt",
		"折扣", "優惠", "打折", "折",
		"早鳥", "團報", "團體",
		"限制", "條件", "驗證",
		"截止", "期限", "日期",
		"保險", "加購", "紀念",
		"21k", "10k", "5k", "3k", "半馬", "全馬",
	}
	for _, kw := range ruleKeywords {
		if strings.Contains(inputLower, kw) {
			return IntentRuleInput
		}
	}

	// Check for greeting/general chat
	greetingKeywords := []string{"你好", "嗨", "哈囉", "hello", "hi", "謝謝", "感謝", "什麼是", "怎麼", "請問"}
	for _, kw := range greetingKeywords {
		if strings.Contains(inputLower, kw) {
			return IntentGeneralChat
		}
	}

	// Default to general chat
	return IntentGeneralChat
}

// containsValueLikeContent checks if input contains value-like content (dates, numbers)
func containsValueLikeContent(input string) bool {
	// Simple check for dates or numbers
	for _, c := range input {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	// Check for date patterns
	dateIndicators := []string{"/", "-", "月", "日", "號"}
	for _, d := range dateIndicators {
		if strings.Contains(input, d) {
			return true
		}
	}
	return false
}
