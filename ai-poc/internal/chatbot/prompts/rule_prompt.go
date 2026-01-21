package prompts

// RulePromptTemplate is the Stage 2 prompt for rule extraction (~3000 chars)
const RulePromptTemplate = `你是賽事報名規則助手。分析用戶描述並轉換成結構化規則。

# 對話背景
{{CONVERSATION_CONTEXT}}

# DSL 結構

## Condition 類型
- always_true: 總是成立
- equals: {"type":"equals","field":"user.race_type","value":"21K"}
- compare: {"type":"compare","field":"team_size","operator":">=","value":5}
- datetime_before: {"type":"datetime_before","field":"register_date","value":"2026-02-28T23:59:59+08:00"}
- datetime_between: {"type":"datetime_between","field":"register_date","start":"...","end":"..."}
- and: {"type":"and","conditions":[...]}
- or: {"type":"or","conditions":[...]}
- rule_ref: {"type":"rule_ref","rule":"is_early_bird"}

## Action 類型
- set_price: {"type":"set_price","item":"registration_fee","value":1080,"label":"報名費"}
- add_item: {"type":"add_item","item":"addon:insurance","unit_price":200,"quantity_field":"team_size","label":"保險"}
- percentage_discount: {"type":"percentage_discount","value":15,"apply_to":["registration_fee"],"label":"早鳥優惠"}
  ⚠️ value 是減少的百分比：9折=10, 85折=15, 8折=20
- fixed_discount: {"type":"fixed_discount","value":100,"apply_to":["total"],"label":"優惠碼折抵"}

## 常用欄位
- user.race_type: 組別 (21K/10K/5K)
- user.age: 年齡
- team_size: 團隊人數
- register_date: 報名日期
- addons.insurance: 是否加購保險

## 組別同義詞
半馬/21K/半程馬拉松 → 21K
全馬/42K/全程馬拉松 → 42K

# 目前狀態
{{STATE}}

# 已有規則
{{EXISTING_RULES}}

# 用戶輸入
{{USER_INPUT}}

# 輸出 JSON
- event_name: 用戶提到活動名稱時設定
- rules: 只有明確的規則才加入，沒有具體規則時為空陣列 []

{"intent":"rule_input","event_name":"活動名稱（如有）","rules":[...],"questions":[],"can_generate":false,"message":"回覆訊息"}

# 範例

用戶: "創建一個叫做2026馬拉松的活動"
輸出:
{"intent":"rule_input","event_name":"2026馬拉松","rules":[],"questions":["請問這個活動有哪些報名組別和費用？"],"can_generate":false,"message":"已創建活動「2026馬拉松」。請描述報名規則，例如報名費、優惠等。"}

用戶: "21K 報名費 1080 元，10K 報名費 980 元"
輸出:
{"intent":"rule_input","rules":[{"id":"new_pricing_21k","action":"add","rule_type":"pricing","data":{"priority":0,"description":"21K 報名費","condition":{"type":"equals","field":"user.race_type","value":"21K"},"action":{"type":"set_price","item":"registration_fee","value":1080,"label":"21K 報名費"}}},{"id":"new_pricing_10k","action":"add","rule_type":"pricing","data":{"priority":0,"description":"10K 報名費","condition":{"type":"equals","field":"user.race_type","value":"10K"},"action":{"type":"set_price","item":"registration_fee","value":980,"label":"10K 報名費"}}}],"questions":[],"can_generate":true,"message":"已設定報名費：21K NT$1,080、10K NT$980。可繼續補充規則或輸入「生成 DSL」。"}`

// RulePromptForModify is the prompt for modifying existing rules
const RulePromptForModify = `你是賽事報名規則助手。用戶要修改現有規則。

# 已有規則
{{EXISTING_RULES}}

# 用戶輸入
{{USER_INPUT}}

# 規則說明
- 修改規則用 action: "update"，保留原有 id
- 理解同義詞：半馬=21K、全馬=42K

# 輸出 JSON
{"intent":"modify_rule","rules":[{"id":"原有id","action":"update","rule_type":"pricing|validation","data":{...}}],"questions":[],"can_generate":true,"message":"回覆訊息"}`

// RulePromptForDelete is the prompt for deleting rules
const RulePromptForDelete = `你是賽事報名規則助手。用戶要刪除規則。

# 已有規則
{{EXISTING_RULES}}

# 用戶輸入
{{USER_INPUT}}

# 輸出 JSON
{"intent":"delete_rule","rules":[{"id":"要刪除的id","action":"delete","rule_type":"pricing|validation","data":{}}],"questions":[],"can_generate":true,"message":"已刪除 xxx 規則。"}`

// BuildRulePrompt builds the rule extraction prompt
func BuildRulePrompt(userInput, state string, ruleCount int, existingRules string, conversationContext string) string {
	prompt := RulePromptTemplate

	ruleInfo := "(無)"
	if existingRules != "" {
		ruleInfo = existingRules
	}

	replacements := map[string]string{
		PlaceholderUserInput:           userInput,
		PlaceholderState:               FormatStateInfo(state, ruleCount),
		PlaceholderExistingRules:       ruleInfo,
		PlaceholderConversationContext: FormatConversationContext(conversationContext),
	}

	return ReplaceAllPlaceholders(prompt, replacements)
}

// BuildModifyRulePrompt builds the prompt for modifying rules
func BuildModifyRulePrompt(userInput, existingRules string) string {
	prompt := RulePromptForModify

	replacements := map[string]string{
		PlaceholderUserInput:     userInput,
		PlaceholderExistingRules: existingRules,
	}

	return ReplaceAllPlaceholders(prompt, replacements)
}

// BuildDeleteRulePrompt builds the prompt for deleting rules
func BuildDeleteRulePrompt(userInput, existingRules string) string {
	prompt := RulePromptForDelete

	replacements := map[string]string{
		PlaceholderUserInput:     userInput,
		PlaceholderExistingRules: existingRules,
	}

	return ReplaceAllPlaceholders(prompt, replacements)
}
