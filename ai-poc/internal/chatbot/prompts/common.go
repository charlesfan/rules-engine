package prompts

import (
	"strings"
)

// Template placeholder constants
const (
	PlaceholderUserInput           = "{{USER_INPUT}}"
	PlaceholderState               = "{{STATE}}"
	PlaceholderExistingRules       = "{{EXISTING_RULES}}"
	PlaceholderPendingQuestions    = "{{PENDING_QUESTIONS}}"
	PlaceholderRelatedRules        = "{{RELATED_RULES}}"
	PlaceholderConversationContext = "{{CONVERSATION_CONTEXT}}"
)

// ReplacePlaceholder replaces a single placeholder with value
func ReplacePlaceholder(template, placeholder, value string) string {
	return strings.ReplaceAll(template, placeholder, value)
}

// ReplaceAllPlaceholders replaces multiple placeholders from a map
func ReplaceAllPlaceholders(template string, replacements map[string]string) string {
	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// FormatStateInfo formats agent state for prompt
func FormatStateInfo(state string, ruleCount int) string {
	if ruleCount == 0 {
		return "狀態: " + state + " (尚無規則)"
	}
	return "狀態: " + state + " (已有 " + intToString(ruleCount) + " 條規則)"
}

// FormatHasPendingQuestions formats pending questions indicator
func FormatHasPendingQuestions(hasPending bool) string {
	if hasPending {
		return "有待回答的問題: 是"
	}
	return "有待回答的問題: 否"
}

// FormatConversationContext formats conversation context for prompt
func FormatConversationContext(summary string) string {
	if summary == "" {
		return "(無對話歷史)"
	}
	return summary
}

// intToString converts int to string without importing strconv
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	negative := false
	if n < 0 {
		negative = true
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
