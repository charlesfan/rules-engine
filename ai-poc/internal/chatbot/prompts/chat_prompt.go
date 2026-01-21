package prompts

// ChatPromptTemplate is the Stage 2 prompt for general chat (~600 chars)
const ChatPromptTemplate = `你是賽事報名規則助手。用繁體中文友善回覆。

## 你的能力
- 幫助用戶將報名規則轉換成 DSL
- 解釋已有規則
- 回答規則相關問題

## 對話背景
{{CONVERSATION_CONTEXT}}

## 目前狀態
{{STATE}}

## 已有規則摘要
{{EXISTING_RULES}}

## 用戶輸入
{{USER_INPUT}}

## 輸出 JSON
{"message": "回覆訊息", "can_generate": false}`

// ChatResponse represents the parsed chat response
type ChatResponse struct {
	Message     string `json:"message"`
	CanGenerate bool   `json:"can_generate"`
}

// BuildChatPrompt builds the chat prompt
func BuildChatPrompt(userInput, state string, ruleCount int, existingRulesSummary string, conversationContext string) string {
	prompt := ChatPromptTemplate

	ruleInfo := "(無)"
	if existingRulesSummary != "" {
		ruleInfo = existingRulesSummary
	}

	replacements := map[string]string{
		PlaceholderUserInput:           userInput,
		PlaceholderState:               FormatStateInfo(state, ruleCount),
		PlaceholderExistingRules:       ruleInfo,
		PlaceholderConversationContext: FormatConversationContext(conversationContext),
	}

	return ReplaceAllPlaceholders(prompt, replacements)
}
