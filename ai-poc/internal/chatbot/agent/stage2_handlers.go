package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charlesfan/rules-engine/ai-poc/internal/chatbot/prompts"
	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

// handleGeneralChat handles general chat intent (Stage 2)
func (a *Agent) handleGeneralChat(ctx context.Context, input string) (*Response, error) {
	// Build chat prompt
	prompt := prompts.BuildChatPrompt(
		input,
		a.state.String(),
		a.getRuleCount(),
		a.buildExistingRulesContext(),
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
		// Fallback to default greeting
		return a.defaultChatResponse(input), nil
	}

	// Parse response
	chatResp, err := parseChatResponse(resp.Content)
	if err != nil {
		// Use raw content as message
		return &Response{
			Message:     extractMessageFromContent(resp.Content),
			State:       a.state.String(),
			RuleCount:   a.getRuleCount(),
			CanGenerate: a.getRuleCount() > 0,
		}, nil
	}

	return &Response{
		Message:     chatResp.Message,
		State:       a.state.String(),
		RuleCount:   a.getRuleCount(),
		CanGenerate: chatResp.CanGenerate || a.getRuleCount() > 0,
	}, nil
}

// handleRuleOperation handles rule_input, modify_rule, delete_rule intents (Stage 2)
func (a *Agent) handleRuleOperation(ctx context.Context, input string, intent string) (*Response, error) {
	// Build appropriate prompt based on intent
	var prompt string
	existingRules := a.buildExistingRulesContext()

	switch intent {
	case IntentModifyRule:
		prompt = prompts.BuildModifyRulePrompt(input, existingRules)
	case IntentDeleteRule:
		prompt = prompts.BuildDeleteRulePrompt(input, existingRules)
	default: // IntentRuleInput
		prompt = prompts.BuildRulePrompt(
			input,
			a.state.String(),
			a.getRuleCount(),
			existingRules,
			a.conversationContext,
		)
	}

	// Call LLM
	req := llm.ChatRequest{
		Messages: []llm.Message{
			llm.NewUserMessage(prompt),
		},
		JSONMode: true,
	}

	resp, err := a.provider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Debug: print raw response
	fmt.Printf("\n[DEBUG] Stage 2 rule operation raw response:\n%s\n[/DEBUG]\n", resp.Content)

	// Parse LLM response
	llmResp, err := ParseLLMResponse(resp.Content)
	if err != nil {
		return &Response{
			Message:     fmt.Sprintf("⚠️ 無法解析 LLM 回應: %v\n\n原始回應: %s", err, truncateString(resp.Content, 200)),
			State:       a.state.String(),
			RuleCount:   a.getRuleCount(),
			CanGenerate: a.getRuleCount() > 0,
		}, nil
	}

	// Process the LLM response (apply rules, etc.)
	return a.processLLMResponse(ctx, llmResp)
}

// handleClarifyResponse handles clarify_response intent (Stage 2)
func (a *Agent) handleClarifyResponse(ctx context.Context, input string) (*Response, error) {
	// Build clarify prompt
	prompt := prompts.BuildClarifyPrompt(
		input,
		a.pendingQuestions,
		a.buildExistingRulesContext(),
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
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Debug
	fmt.Printf("\n[DEBUG] Stage 2 clarify response raw:\n%s\n[/DEBUG]\n", resp.Content)

	// Parse LLM response
	llmResp, err := ParseLLMResponse(resp.Content)
	if err != nil {
		return &Response{
			Message:     fmt.Sprintf("⚠️ 無法解析回應: %v", err),
			State:       a.state.String(),
			RuleCount:   a.getRuleCount(),
			CanGenerate: a.getRuleCount() > 0,
		}, nil
	}

	// Clear pending questions if answered
	if len(llmResp.Questions) == 0 {
		a.pendingQuestions = nil
	}

	return a.processLLMResponse(ctx, llmResp)
}

// handleDSLGeneration handles dsl_request intent (Stage 2)
// This doesn't need LLM call - it directly generates DSL from current RuleSet
func (a *Agent) handleDSLGeneration(ctx context.Context) (*Response, error) {
	response := &Response{
		State:       a.state.String(),
		RuleCount:   a.getRuleCount(),
		CanGenerate: a.getRuleCount() > 0,
	}

	return a.handleDSLRequest(ctx, response)
}

// defaultChatResponse returns a default greeting response
func (a *Agent) defaultChatResponse(input string) *Response {
	inputLower := strings.ToLower(input)

	var message string
	if strings.Contains(inputLower, "目前") || strings.Contains(inputLower, "規則") {
		if a.getRuleCount() == 0 {
			message = "目前沒有已建立的規則。\n\n請描述您的賽事報名規則，例如：\n- 「報名費 1000 元」\n- 「早鳥優惠 9 折到 2/28」\n- 「團報 5 人以上 95 折」"
		} else {
			message = a.GetRulesSummary() + "\n\n可以輸入「生成 DSL」產生規則，或繼續補充其他規則。"
		}
	} else {
		message = "你好！我是賽事報名規則助手。\n\n請描述您的賽事報名規則，例如：\n- 「報名費 1000 元」\n- 「早鳥優惠 9 折到 2/28」\n- 「團報 5 人以上 95 折」\n\n我會幫您整理成系統可執行的規則。"
	}

	return &Response{
		Message:     message,
		State:       a.state.String(),
		RuleCount:   a.getRuleCount(),
		CanGenerate: a.getRuleCount() > 0,
	}
}

// parseChatResponse parses the chat prompt response
func parseChatResponse(content string) (*prompts.ChatResponse, error) {
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return nil, &ParseError{Message: "no JSON found", Content: content}
	}

	var resp prompts.ChatResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, &ParseError{Message: err.Error(), Content: jsonStr}
	}

	return &resp, nil
}

// extractMessageFromContent tries to extract a message from LLM content
func extractMessageFromContent(content string) string {
	// Try to find message in JSON
	jsonStr := extractJSON(content)
	if jsonStr != "" {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
			if msg, ok := data["message"].(string); ok && msg != "" {
				return msg
			}
			if resp, ok := data["response"].(string); ok && resp != "" {
				return resp
			}
		}
	}

	// Return cleaned content
	content = strings.TrimSpace(content)
	if content == "" {
		return "收到您的訊息。"
	}
	return content
}

// truncateString truncates a string to maxLen with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
