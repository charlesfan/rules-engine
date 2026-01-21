package prompts

// IntentPromptTemplate is the Stage 1 prompt for intent detection (~600 chars)
// This prompt is designed to be lightweight for small LLM models
const IntentPromptTemplate = `你是意圖分析助手。分析用戶輸入，判斷意圖類型。

## 意圖類型
- rule_input: 描述新規則（報名費、折扣、限制、日期）
- modify_rule: 修改現有規則
- delete_rule: 刪除規則
- dsl_request: 要求生成 DSL（「生成」「產生」「完成」「確認」）
- general_chat: 一般對話、問候、詢問
- clarify_response: 回答問題、補充資訊

## 對話背景
{{CONVERSATION_CONTEXT}}

## 用戶輸入
{{USER_INPUT}}

## 目前狀態
{{STATE}}
{{PENDING_QUESTIONS}}

## 輸出 JSON
{"intent": "意圖類型"}`

// IntentResponse represents the parsed intent detection response
type IntentResponse struct {
	Intent string `json:"intent"`
}

// BuildIntentPrompt builds the intent detection prompt
func BuildIntentPrompt(userInput, state string, hasPendingQuestions bool, ruleCount int, conversationContext string) string {
	prompt := IntentPromptTemplate

	replacements := map[string]string{
		PlaceholderUserInput:           userInput,
		PlaceholderState:               FormatStateInfo(state, ruleCount),
		PlaceholderPendingQuestions:    FormatHasPendingQuestions(hasPendingQuestions),
		PlaceholderConversationContext: FormatConversationContext(conversationContext),
	}

	return ReplaceAllPlaceholders(prompt, replacements)
}
