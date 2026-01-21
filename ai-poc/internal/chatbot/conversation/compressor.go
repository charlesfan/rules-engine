package conversation

import (
	"context"
	"fmt"
	"strings"

	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

/**
 * LLMCompressor uses an LLM to compress conversation history into a summary.
 * It preserves key information like:
 *   - Conversation topic/purpose
 *   - Confirmed decisions and rules
 *   - Important context for ongoing discussion
 */
type LLMCompressor struct {
	provider llm.Provider
	prompt   string
}

// NewLLMCompressor creates a new LLM-based compressor
func NewLLMCompressor(provider llm.Provider) *LLMCompressor {
	return &LLMCompressor{
		provider: provider,
		prompt:   defaultCompressPrompt,
	}
}

// NewLLMCompressorWithPrompt creates a compressor with custom prompt
func NewLLMCompressorWithPrompt(provider llm.Provider, prompt string) *LLMCompressor {
	return &LLMCompressor{
		provider: provider,
		prompt:   prompt,
	}
}

const defaultCompressPrompt = `你是一個對話摘要助手。請將以下對話歷史壓縮成簡潔的摘要。

要求：
1. 保留對話的主題和目的
2. 保留所有已確認的決定、規則或重要資訊
3. 保留用戶的關鍵偏好或要求
4. 移除寒暄、重複的內容
5. 使用條列式整理重點
6. 摘要長度控制在 200 字以內

{{EXISTING_SUMMARY}}

需要壓縮的對話：
{{MESSAGES}}

請輸出摘要（純文字，不要加入任何標記）：`

/**
 * Compress compresses existing summary and new messages into a new summary.
 * If existingSummary is empty, only the new messages are summarized.
 */
func (c *LLMCompressor) Compress(ctx context.Context, existingSummary string, messages []llm.Message) (string, error) {
	if len(messages) == 0 {
		return existingSummary, nil
	}

	// Build the prompt
	prompt := c.buildPrompt(existingSummary, messages)

	// Call LLM
	req := llm.ChatRequest{
		Messages: []llm.Message{
			llm.NewUserMessage(prompt),
		},
	}

	resp, err := c.provider.Chat(ctx, req)
	if err != nil {
		return "", fmt.Errorf("LLM compression failed: %w", err)
	}

	return strings.TrimSpace(resp.Content), nil
}

// buildPrompt constructs the compression prompt
func (c *LLMCompressor) buildPrompt(existingSummary string, messages []llm.Message) string {
	prompt := c.prompt

	// Add existing summary section
	if existingSummary != "" {
		summarySection := fmt.Sprintf("現有摘要：\n%s\n", existingSummary)
		prompt = strings.Replace(prompt, "{{EXISTING_SUMMARY}}", summarySection, 1)
	} else {
		prompt = strings.Replace(prompt, "{{EXISTING_SUMMARY}}", "", 1)
	}

	// Format messages
	var msgBuilder strings.Builder
	for _, msg := range messages {
		role := "用戶"
		if msg.Role == llm.RoleAssistant {
			role = "AI"
		} else if msg.Role == llm.RoleSystem {
			role = "系統"
		}
		msgBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	prompt = strings.Replace(prompt, "{{MESSAGES}}", msgBuilder.String(), 1)

	return prompt
}

/**
 * NoOpCompressor is a compressor that does nothing.
 * Useful for testing or when compression is not needed.
 */
type NoOpCompressor struct{}

// NewNoOpCompressor creates a no-op compressor
func NewNoOpCompressor() *NoOpCompressor {
	return &NoOpCompressor{}
}

// Compress returns the existing summary unchanged
func (c *NoOpCompressor) Compress(ctx context.Context, existingSummary string, messages []llm.Message) (string, error) {
	return existingSummary, nil
}

/**
 * SimpleCompressor concatenates messages into a simple text summary.
 * This is a fallback when LLM is not available.
 */
type SimpleCompressor struct {
	maxLength int
}

// NewSimpleCompressor creates a simple text-based compressor
func NewSimpleCompressor(maxLength int) *SimpleCompressor {
	if maxLength <= 0 {
		maxLength = 500
	}
	return &SimpleCompressor{maxLength: maxLength}
}

// Compress creates a simple text summary
func (c *SimpleCompressor) Compress(ctx context.Context, existingSummary string, messages []llm.Message) (string, error) {
	var builder strings.Builder

	// Add existing summary
	if existingSummary != "" {
		builder.WriteString(existingSummary)
		builder.WriteString("\n---\n")
	}

	// Add message summaries (just user messages for brevity)
	for _, msg := range messages {
		if msg.Role == llm.RoleUser {
			content := msg.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			builder.WriteString("• ")
			builder.WriteString(content)
			builder.WriteString("\n")
		}
	}

	result := builder.String()

	// Truncate if too long
	if len(result) > c.maxLength {
		result = result[:c.maxLength] + "..."
	}

	return result, nil
}
