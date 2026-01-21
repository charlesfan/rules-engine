/**
 * Package conversation provides conversation history management with progressive summarization.
 */
package conversation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

// Compressor compresses conversation history into a summary
type Compressor interface {
	// Compress takes existing summary and recent messages, returns new summary
	Compress(ctx context.Context, existingSummary string, messages []llm.Message) (string, error)
}

// MemoryConfig holds configuration for Memory
type MemoryConfig struct {
	/**
	 * CompressThreshold is the number of recent messages before triggering compression
	 * Default: 10 (5 rounds of user-assistant conversation)
	 */
	CompressThreshold int

	/**
	 * KeepAfterCompress is the number of recent messages to keep after compression
	 * These messages won't be included in the summary to maintain immediate context
	 * Default: 4 (2 rounds of conversation)
	 */
	KeepAfterCompress int
}

// DefaultMemoryConfig returns the default memory configuration
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		CompressThreshold: 10,
		KeepAfterCompress: 4,
	}
}

/**
 * Memory stores conversation history with progressive summarization
 *
 * Structure:
 *   [System Prompt] + [Summary] + [Recent Messages]
 *
 * When recent messages exceed CompressThreshold:
 *   1. Messages to compress = recent[0 : len-KeepAfterCompress]
 *   2. New summary = Compress(old summary + messages to compress)
 *   3. Keep only recent[len-KeepAfterCompress:]
 */
type Memory struct {
	summary        string        // Accumulated conversation summary
	recentMessages []llm.Message // Recent messages (not yet summarized)
	config         MemoryConfig
	compressor     Compressor
	metadata       map[string]any
	createdAt      time.Time
	lastCompressAt time.Time
	compressCount  int // Number of times compression has occurred
	mu             sync.RWMutex
}

// NewMemory creates a new conversation memory with default config
func NewMemory() *Memory {
	return NewMemoryWithConfig(DefaultMemoryConfig())
}

// NewMemoryWithConfig creates a new conversation memory with custom config
func NewMemoryWithConfig(config MemoryConfig) *Memory {
	return &Memory{
		summary:        "",
		recentMessages: make([]llm.Message, 0),
		config:         config,
		metadata:       make(map[string]any),
		createdAt:      time.Now(),
	}
}

// SetCompressor sets the compressor for summarization
func (m *Memory) SetCompressor(c Compressor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compressor = c
}

// Add adds a message to the conversation history
func (m *Memory) Add(msg llm.Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recentMessages = append(m.recentMessages, msg)
}

// AddUserMessage adds a user message to the history
func (m *Memory) AddUserMessage(content string) {
	m.Add(llm.NewUserMessage(content))
}

// AddAssistantMessage adds an assistant message to the history
func (m *Memory) AddAssistantMessage(content string) {
	m.Add(llm.NewAssistantMessage(content))
}

// NeedsCompression checks if compression should be triggered
func (m *Memory) NeedsCompression() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.recentMessages) >= m.config.CompressThreshold
}

/**
 * Compress compresses the conversation history
 * Returns error if no compressor is set or compression fails
 */
func (m *Memory) Compress(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.compressor == nil {
		return fmt.Errorf("no compressor set")
	}

	if len(m.recentMessages) < m.config.CompressThreshold {
		return nil // No need to compress
	}

	// Calculate how many messages to compress
	compressEnd := len(m.recentMessages) - m.config.KeepAfterCompress
	if compressEnd <= 0 {
		return nil // Not enough messages to compress
	}

	messagesToCompress := m.recentMessages[:compressEnd]
	messagesToKeep := m.recentMessages[compressEnd:]

	// Compress
	newSummary, err := m.compressor.Compress(ctx, m.summary, messagesToCompress)
	if err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	// Update state
	m.summary = newSummary
	m.recentMessages = make([]llm.Message, len(messagesToKeep))
	copy(m.recentMessages, messagesToKeep)
	m.lastCompressAt = time.Now()
	m.compressCount++

	return nil
}

/**
 * TryCompress attempts compression if threshold is reached
 * This is a convenience method that combines NeedsCompression and Compress
 */
func (m *Memory) TryCompress(ctx context.Context) error {
	if !m.NeedsCompression() {
		return nil
	}
	return m.Compress(ctx)
}

// GetMessages returns all recent messages (not including summary)
func (m *Memory) GetMessages() []llm.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]llm.Message, len(m.recentMessages))
	copy(result, m.recentMessages)
	return result
}

// GetSummary returns the current conversation summary
func (m *Memory) GetSummary() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.summary
}

// GetConversationContext returns conversation context for LLM prompts
// - If messages <= threshold: returns all messages formatted
// - If messages > threshold: returns compressed summary
func (m *Memory) GetConversationContext() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// If we have a summary (already compressed), use it + recent messages
	if m.summary != "" {
		var sb strings.Builder
		sb.WriteString("[對話摘要]\n")
		sb.WriteString(m.summary)
		if len(m.recentMessages) > 0 {
			sb.WriteString("\n\n[近期對話]\n")
			sb.WriteString(m.formatMessages(m.recentMessages))
		}
		return sb.String()
	}

	// No compression yet - return all recent messages
	if len(m.recentMessages) == 0 {
		return ""
	}

	return m.formatMessages(m.recentMessages)
}

// formatMessages formats messages for context, extracting key information
func (m *Memory) formatMessages(messages []llm.Message) string {
	var sb strings.Builder
	for i, msg := range messages {
		if i > 0 {
			sb.WriteString("\n")
		}
		role := "用戶"
		if msg.Role == "assistant" {
			role = "助手"
		}
		sb.WriteString(fmt.Sprintf("%s: %s", role, msg.Content))
	}
	return sb.String()
}

// Clear removes all messages and summary from the conversation
func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.summary = ""
	m.recentMessages = make([]llm.Message, 0)
	m.createdAt = time.Now()
	m.lastCompressAt = time.Time{}
	m.compressCount = 0
}

// Len returns the number of recent messages (not including summarized)
func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.recentMessages)
}

/**
 * TotalLen returns an estimate of total messages (including summarized)
 * This is an approximation since summarized messages are compressed
 */
func (m *Memory) TotalLen() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Estimate: each compression combines ~(threshold - keep) messages
	compressedPerRound := m.config.CompressThreshold - m.config.KeepAfterCompress
	estimatedCompressed := m.compressCount * compressedPerRound

	return estimatedCompressed + len(m.recentMessages)
}

// SetMetadata sets a metadata value
func (m *Memory) SetMetadata(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[key] = value
}

// GetMetadata gets a metadata value
func (m *Memory) GetMetadata(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.metadata[key]
	return v, ok
}

// BuildRequest creates a ChatRequest with summary and recent messages
func (m *Memory) BuildRequest(systemPrompt string) llm.ChatRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var messages []llm.Message

	// 1. Add system prompt if provided
	if systemPrompt != "" {
		messages = append(messages, llm.NewSystemMessage(systemPrompt))
	}

	// 2. Add summary as system context if exists
	if m.summary != "" {
		summaryMsg := llm.NewSystemMessage(fmt.Sprintf(
			"[對話摘要]\n%s\n[以下是最近的對話]",
			m.summary,
		))
		messages = append(messages, summaryMsg)
	}

	// 3. Add recent messages
	messages = append(messages, m.recentMessages...)

	return llm.ChatRequest{
		Messages: messages,
	}
}

// Stats holds statistics about the memory state
type Stats struct {
	RecentMessageCount int           `json:"recent_message_count"`
	HasSummary         bool          `json:"has_summary"`
	CompressCount      int           `json:"compress_count"`
	EstimatedTotal     int           `json:"estimated_total"`
	Duration           time.Duration `json:"duration"`
	CreatedAt          time.Time     `json:"created_at"`
	LastCompressAt     time.Time     `json:"last_compress_at,omitempty"`
}

// GetStats returns statistics about the memory
func (m *Memory) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Stats{
		RecentMessageCount: len(m.recentMessages),
		HasSummary:         m.summary != "",
		CompressCount:      m.compressCount,
		EstimatedTotal:     m.TotalLen(),
		Duration:           time.Since(m.createdAt),
		CreatedAt:          m.createdAt,
		LastCompressAt:     m.lastCompressAt,
	}
}
