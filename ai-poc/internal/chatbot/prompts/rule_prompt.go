package prompts

// RulePromptTemplate is the Stage 2 prompt for rule extraction (~3000 chars)
const RulePromptTemplate = `你是賽事報名規則助手。分析用戶描述並轉換成結構化規則。

# ⚠️ 核心規則（必須遵守）
1. 只處理「用戶輸入」中明確提到的內容
2. 用戶沒有說具體金額/數字 → rules 必須是空陣列 []
3. 絕對不要自己編造規則、組別或金額！

# DSL 結構

## Condition 類型
- always_true: 總是成立
- equals: {"type":"equals","field":"user.race_type","value":"xxx"}
- compare: {"type":"compare","field":"team_size","operator":">=","value":5}
- datetime_before: {"type":"datetime_before","field":"register_date","value":"2026-02-28T23:59:59+08:00"}
- datetime_between: {"type":"datetime_between","field":"register_date","start":"...","end":"..."}
- and/or: {"type":"and","conditions":[...]}

## Action 類型
- set_price: {"type":"set_price","item":"registration_fee","value":金額,"label":"描述"}
- add_item: {"type":"add_item","item":"addon:xxx","unit_price":金額,"label":"描述"}
- percentage_discount: {"type":"percentage_discount","value":折扣百分比,"apply_to":["registration_fee"],"label":"描述"}
  （9折=10, 85折=15, 8折=20）
- fixed_discount: {"type":"fixed_discount","value":金額,"apply_to":["total"],"label":"描述"}

## 常用欄位
user.race_type（組別）、user.age（年齡）、team_size（團隊人數）、register_date（報名日期）

## 組別同義詞
半馬=21K、全馬=42K

# 目前狀態
{{STATE}}

# 已有規則
{{EXISTING_RULES}}

# 用戶輸入
{{USER_INPUT}}

# 輸出格式
{"intent":"rule_input","event_name":"活動名稱或空字串","rules":[只放用戶明確說的規則],"questions":[],"can_generate":布林值,"message":"回覆"}

# 範例

## 範例1：只創建活動（沒有金額 → rules 必須空）
用戶: "創建一個馬拉松活動"
輸出: {"intent":"rule_input","event_name":"馬拉松活動","rules":[],"questions":[],"can_generate":false,"message":"已創建活動。請描述報名組別和費用。"}

## 範例2：新增一個組別（有明確金額）
用戶: "新增5K組，報名費500元"
輸出: {"intent":"rule_input","event_name":"","rules":[{"id":"new_pricing_5k","action":"add","rule_type":"pricing","data":{"priority":0,"description":"5K 報名費","condition":{"type":"equals","field":"user.race_type","value":"5K"},"action":{"type":"set_price","item":"registration_fee","value":500,"label":"5K 報名費"}}}],"questions":[],"can_generate":true,"message":"已新增 5K 組，報名費 NT$500。"}`

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
