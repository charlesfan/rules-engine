package prompts

// ClarifyPromptTemplate is the Stage 2 prompt for handling clarification responses (~1000 chars)
const ClarifyPromptTemplate = `你是規則釐清助手。用戶正在回答你之前的問題。

# 對話背景
{{CONVERSATION_CONTEXT}}

# 待釐清的問題
{{PENDING_QUESTIONS}}

# 相關規則（可能需要更新）
{{RELATED_RULES}}

# 用戶回答
{{USER_INPUT}}

# 任務
1. 解析用戶回答
2. 更新相關規則（如需要）
3. 確認是否還有其他問題

# Condition 類型參考
- datetime_before: {"type":"datetime_before","field":"register_date","value":"2026-02-28T23:59:59+08:00"}
- compare: {"type":"compare","field":"team_size","operator":">=","value":5}
- rule_ref: {"type":"rule_ref","rule":"is_early_bird"}

# Action 類型參考
- percentage_discount: value 是減少百分比（9折=10, 85折=15）
- set_price: {"type":"set_price","item":"registration_fee","value":1080}

# 輸出 JSON
{
  "intent": "clarify_response",
  "rules": [
    {
      "id": "要更新的規則id",
      "action": "update",
      "rule_type": "pricing|validation",
      "data": {
        "priority": 0,
        "description": "更新後的描述",
        "condition": {...},
        "action": {...}
      }
    }
  ],
  "rule_definitions": {
    "name": {
      "action": "add|update",
      "data": {"type":"condition","description":"...","expression":{...}}
    }
  },
  "questions": [],
  "can_generate": true,
  "message": "回覆訊息"
}`

// BuildClarifyPrompt builds the clarification response prompt
func BuildClarifyPrompt(userInput string, pendingQuestions []string, relatedRules string, conversationContext string) string {
	prompt := ClarifyPromptTemplate

	questionsText := "(無)"
	if len(pendingQuestions) > 0 {
		questionsText = ""
		for i, q := range pendingQuestions {
			questionsText += intToString(i+1) + ". " + q + "\n"
		}
	}

	rulesText := "(無)"
	if relatedRules != "" {
		rulesText = relatedRules
	}

	replacements := map[string]string{
		PlaceholderUserInput:           userInput,
		PlaceholderPendingQuestions:    questionsText,
		PlaceholderRelatedRules:        rulesText,
		PlaceholderConversationContext: FormatConversationContext(conversationContext),
	}

	return ReplaceAllPlaceholders(prompt, replacements)
}
