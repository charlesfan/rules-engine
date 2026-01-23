package prompts

// IntentPromptTemplate is the Stage 1 prompt for intent detection (~600 chars)
// This prompt is designed to be lightweight for small LLM models
const IntentPromptTemplate = `你是意圖分析助手。分析用戶輸入，判斷意圖類型。

## 意圖類型
- rule_input: 創建活動、新增規則、描述規則（「創建」「增加」「新增」「加入」「設定」）
- modify_rule: 修改「已存在」的規則（「改成」「調整」「更新」現有規則的數值）
- delete_rule: 刪除規則（「刪除」「移除」）
- dsl_request: 要求生成 DSL（「生成」「產生」「完成」「確認」「看一下」）
- general_chat: 一般對話、問候、詢問
- clarify_response: 回答問題、補充資訊

## 判斷重點
- 「創建活動 xxx」「建立活動」→ rule_input
- 「增加 xxx 組別」「新增 xxx」→ rule_input
- 「把 xxx 改成 yyy」「調整 xxx 為」→ modify_rule（只有明確說「改」「調整」現有的才是）

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
