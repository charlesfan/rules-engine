package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ClaudeConfig holds configuration for the Claude provider
type ClaudeConfig struct {
	APIKey  string `yaml:"api_key" json:"api_key"`
	Model   string `yaml:"model" json:"model"`
	BaseURL string `yaml:"base_url" json:"base_url"`
	Timeout int    `yaml:"timeout" json:"timeout"` // Timeout in seconds
}

// DefaultClaudeConfig returns the default Claude configuration
func DefaultClaudeConfig() ClaudeConfig {
	return ClaudeConfig{
		APIKey:  os.Getenv("ANTHROPIC_API_KEY"),
		Model:   "claude-sonnet-4-20250514",
		BaseURL: "https://api.anthropic.com",
		Timeout: 120,
	}
}

// ClaudeProvider implements the Provider interface for Claude API
type ClaudeProvider struct {
	config ClaudeConfig
	client *http.Client
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(config ClaudeConfig) *ClaudeProvider {
	defaults := DefaultClaudeConfig()

	if config.APIKey == "" {
		config.APIKey = defaults.APIKey
	}
	if config.Model == "" {
		config.Model = defaults.Model
	}
	if config.BaseURL == "" {
		config.BaseURL = defaults.BaseURL
	}
	if config.Timeout == 0 {
		config.Timeout = defaults.Timeout
	}

	return &ClaudeProvider{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return fmt.Sprintf("claude/%s", p.config.Model)
}

// claudeRequest represents the Claude API request format
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse represents the Claude API response format
type claudeResponse struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Role         string          `json:"role"`
	Content      []claudeContent `json:"content"`
	Model        string          `json:"model"`
	StopReason   string          `json:"stop_reason"`
	StopSequence string          `json:"stop_sequence,omitempty"`
	Usage        claudeUsage     `json:"usage"`
}

type claudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// claudeError represents an error response from Claude API
type claudeError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Chat sends a chat request to Claude
func (p *ClaudeProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if p.config.APIKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}

	// Separate system message from other messages
	var systemPrompt string
	var claudeMessages []claudeMessage

	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			systemPrompt = msg.Content
		} else {
			claudeMessages = append(claudeMessages, claudeMessage{
				Role:    string(msg.Role),
				Content: msg.Content,
			})
		}
	}

	// Ensure at least one message
	if len(claudeMessages) == 0 {
		return nil, fmt.Errorf("at least one non-system message is required")
	}

	// Set max tokens
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	// Build request
	claudeReq := claudeRequest{
		Model:     p.config.Model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages:  claudeMessages,
	}

	// Marshal request
	body, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		var errResp claudeError
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, fmt.Errorf("claude API error: %s - %s", errResp.Error.Type, errResp.Error.Message)
		}
		return nil, fmt.Errorf("claude returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text content
	var content string
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &ChatResponse{
		Content:      content,
		FinishReason: claudeResp.StopReason,
		TokensUsed:   claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
	}, nil
}

// Available checks if Claude API is available (API key is set)
func (p *ClaudeProvider) Available(ctx context.Context) bool {
	return p.config.APIKey != ""
}

// SetModel changes the model used by this provider
func (p *ClaudeProvider) SetModel(model string) {
	p.config.Model = model
}

// GetModel returns the current model
func (p *ClaudeProvider) GetModel() string {
	return p.config.Model
}

// AvailableModels returns a list of available Claude models
func (p *ClaudeProvider) AvailableModels() []string {
	return []string{
		"claude-sonnet-4-20250514",
		"claude-3-5-sonnet-20241022",
		"claude-3-5-haiku-20241022",
		"claude-3-opus-20240229",
	}
}
