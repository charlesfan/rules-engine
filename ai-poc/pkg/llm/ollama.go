package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaConfig holds configuration for the Ollama provider
type OllamaConfig struct {
	BaseURL string `yaml:"base_url" json:"base_url"`
	Model   string `yaml:"model" json:"model"`
	Timeout int    `yaml:"timeout" json:"timeout"` // Timeout in seconds
}

// DefaultOllamaConfig returns the default Ollama configuration
func DefaultOllamaConfig() OllamaConfig {
	return OllamaConfig{
		BaseURL: "http://localhost:11434",
		//Model:   "llama3.1:8b",
		//Model:   "qwen2.5:32b",
		Model:   "llama3.1:70b",
		Timeout: 120,
	}
}

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	config OllamaConfig
	client *http.Client
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(config OllamaConfig) *OllamaProvider {
	if config.BaseURL == "" {
		config.BaseURL = DefaultOllamaConfig().BaseURL
	}
	if config.Model == "" {
		config.Model = DefaultOllamaConfig().Model
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultOllamaConfig().Timeout
	}

	return &OllamaProvider{
		config: config,
		client: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return fmt.Sprintf("ollama/%s", p.config.Model)
}

// ollamaChatRequest represents the Ollama API request format
type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Format   string          `json:"format,omitempty"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
}

// ollamaChatResponse represents the Ollama API response format
type ollamaChatResponse struct {
	Model           string        `json:"model"`
	CreatedAt       string        `json:"created_at"`
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	TotalDuration   int64         `json:"total_duration,omitempty"`
	PromptEvalCount int           `json:"prompt_eval_count,omitempty"`
	EvalCount       int           `json:"eval_count,omitempty"`
}

// Chat sends a chat request to Ollama
func (p *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages to Ollama format
	ollamaMessages := make([]ollamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		ollamaMessages[i] = ollamaMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	// Build request
	ollamaReq := ollamaChatRequest{
		Model:    p.config.Model,
		Messages: ollamaMessages,
		Stream:   false,
	}

	// Set JSON mode if requested
	if req.JSONMode {
		ollamaReq.Format = "json"
	}

	// Set options if provided
	if req.Temperature > 0 || req.MaxTokens > 0 {
		ollamaReq.Options = &ollamaOptions{}
		if req.Temperature > 0 {
			ollamaReq.Options.Temperature = req.Temperature
		}
		if req.MaxTokens > 0 {
			ollamaReq.Options.NumPredict = req.MaxTokens
		}
	}

	// Marshal request
	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

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

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var ollamaResp ollamaChatResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ChatResponse{
		Content:    ollamaResp.Message.Content,
		TokensUsed: ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
	}, nil
}

// Available checks if Ollama is available
func (p *OllamaProvider) Available(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ListModels returns available models from Ollama
func (p *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}
	return models, nil
}

// SetModel changes the model used by this provider
func (p *OllamaProvider) SetModel(model string) {
	p.config.Model = model
}

// GetModel returns the current model
func (p *OllamaProvider) GetModel() string {
	return p.config.Model
}
