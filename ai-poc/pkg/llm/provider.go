// Package llm provides a unified interface for interacting with different LLM providers.
package llm

import (
	"context"
	"encoding/json"
)

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a chat message
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to the LLM
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	JSONMode    bool      `json:"json_mode,omitempty"` // Request JSON output
}

// ChatResponse represents a response from the LLM
type ChatResponse struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason,omitempty"`
	TokensUsed   int    `json:"tokens_used,omitempty"`
}

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider name (e.g., "ollama/llama3.1:8b")
	Name() string

	// Chat sends a chat request and returns the response
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// ChatStream sends a chat request and streams the response (optional)
	// ChatStream(ctx context.Context, req ChatRequest) (<-chan string, error)

	// Available checks if the provider is available and ready
	Available(ctx context.Context) bool
}

// ProviderType represents supported LLM providers
type ProviderType string

const (
	ProviderOllama ProviderType = "ollama"
	ProviderClaude ProviderType = "claude"
)

// String returns the string representation of the provider type
func (p ProviderType) String() string {
	return string(p)
}

// ParseProviderType parses a string to ProviderType
func ParseProviderType(s string) (ProviderType, bool) {
	switch s {
	case "ollama":
		return ProviderOllama, true
	case "claude":
		return ProviderClaude, true
	default:
		return "", false
	}
}

// Helper functions for creating messages

// NewSystemMessage creates a system message
func NewSystemMessage(content string) Message {
	return Message{Role: RoleSystem, Content: content}
}

// NewUserMessage creates a user message
func NewUserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

// NewAssistantMessage creates an assistant message
func NewAssistantMessage(content string) Message {
	return Message{Role: RoleAssistant, Content: content}
}

// ChatWithJSON is a helper that sends a request expecting JSON response
func ChatWithJSON(ctx context.Context, p Provider, req ChatRequest, result any) error {
	req.JSONMode = true
	resp, err := p.Chat(ctx, req)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(resp.Content), result)
}
