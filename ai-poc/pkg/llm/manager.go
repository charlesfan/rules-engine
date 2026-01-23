package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charlesfan/rules-engine/ai-poc/pkg/config"
)

// Manager manages multiple LLM providers and handles switching between them
type Manager struct {
	providers map[ProviderType]Provider
	timeouts  map[ProviderType]time.Duration // timeout per provider
	current   ProviderType
	mu        sync.RWMutex
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[ProviderType]Provider),
		timeouts:  make(map[ProviderType]time.Duration),
		current:   ProviderOllama, // Default to Ollama
	}
}

// Register adds a provider to the manager
func (m *Manager) Register(providerType ProviderType, provider Provider) {
	m.RegisterWithTimeout(providerType, provider, 2*time.Minute) // default 2 min
}

// RegisterWithTimeout adds a provider with specific timeout
func (m *Manager) RegisterWithTimeout(providerType ProviderType, provider Provider, timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[providerType] = provider
	m.timeouts[providerType] = timeout
}

// CurrentTimeout returns the timeout for the current provider
func (m *Manager) CurrentTimeout() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.timeouts[m.current]; ok {
		return t
	}
	return 2 * time.Minute // fallback default
}

// Switch changes the current active provider
func (m *Manager) Switch(providerType ProviderType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.providers[providerType]; !ok {
		return fmt.Errorf("provider %s not registered", providerType)
	}
	m.current = providerType
	return nil
}

// Toggle switches between available providers
func (m *Manager) Toggle(ctx context.Context) (ProviderType, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get available providers
	var available []ProviderType
	for pt, p := range m.providers {
		if p.Available(ctx) {
			available = append(available, pt)
		}
	}

	if len(available) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	if len(available) == 1 {
		return m.current, fmt.Errorf("only one provider available: %s", m.current)
	}

	// Find next provider
	for _, pt := range available {
		if pt != m.current {
			m.current = pt
			return pt, nil
		}
	}

	return m.current, nil
}

// Current returns the current active provider
func (m *Manager) Current() Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.providers[m.current]
}

// CurrentType returns the type of the current active provider
func (m *Manager) CurrentType() ProviderType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// Get returns a specific provider by type
func (m *Manager) Get(providerType ProviderType) (Provider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.providers[providerType]
	return p, ok
}

// ListRegistered returns all registered provider types
func (m *Manager) ListRegistered() []ProviderType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]ProviderType, 0, len(m.providers))
	for pt := range m.providers {
		types = append(types, pt)
	}
	return types
}

// ListAvailable returns provider types that are currently available
func (m *Manager) ListAvailable(ctx context.Context) []ProviderType {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var available []ProviderType
	for pt, p := range m.providers {
		if p.Available(ctx) {
			available = append(available, pt)
		}
	}
	return available
}

// Chat sends a chat request using the current provider
func (m *Manager) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	provider := m.Current()
	if provider == nil {
		return nil, fmt.Errorf("no provider available")
	}
	return provider.Chat(ctx, req)
}

// Status returns the status of all registered providers
type ProviderStatus struct {
	Type      ProviderType `json:"type"`
	Name      string       `json:"name"`
	Available bool         `json:"available"`
	IsCurrent bool         `json:"is_current"`
}

// GetStatus returns the status of all registered providers
func (m *Manager) GetStatus(ctx context.Context) []ProviderStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]ProviderStatus, 0, len(m.providers))
	for pt, p := range m.providers {
		statuses = append(statuses, ProviderStatus{
			Type:      pt,
			Name:      p.Name(),
			Available: p.Available(ctx),
			IsCurrent: pt == m.current,
		})
	}
	return statuses
}

// DefaultManager creates a manager with default providers configured
func DefaultManager() *Manager {
	m := NewManager()

	ollamaConfig := DefaultOllamaConfig()
	claudeConfig := DefaultClaudeConfig()

	// Register Ollama with default config
	m.RegisterWithTimeout(ProviderOllama, NewOllamaProvider(ollamaConfig), time.Duration(ollamaConfig.Timeout)*time.Second)

	// Register Claude with default config
	m.RegisterWithTimeout(ProviderClaude, NewClaudeProvider(claudeConfig), time.Duration(claudeConfig.Timeout)*time.Second)

	return m
}

// NewManagerFromConfig creates a manager from configuration
func NewManagerFromConfig(cfg *config.Config) *Manager {
	m := NewManager()

	// Set default provider
	switch cfg.LLM.DefaultProvider {
	case "claude":
		m.current = ProviderClaude
	default:
		m.current = ProviderOllama
	}

	// Register Ollama
	ollamaConfig := OllamaConfig{
		BaseURL: cfg.LLM.Ollama.BaseURL,
		Model:   cfg.LLM.Ollama.Model,
		Timeout: cfg.LLM.Ollama.Timeout,
	}
	if ollamaConfig.BaseURL == "" {
		ollamaConfig.BaseURL = "http://localhost:11434"
	}
	if ollamaConfig.Model == "" {
		ollamaConfig.Model = "llama3.1:8b"
	}
	if ollamaConfig.Timeout == 0 {
		ollamaConfig.Timeout = 120
	}
	m.RegisterWithTimeout(ProviderOllama, NewOllamaProvider(ollamaConfig), time.Duration(ollamaConfig.Timeout)*time.Second)

	// Register Claude
	claudeConfig := ClaudeConfig{
		APIKey:  cfg.LLM.Claude.APIKey,
		Model:   cfg.LLM.Claude.Model,
		Timeout: cfg.LLM.Claude.Timeout,
	}
	if claudeConfig.Model == "" {
		claudeConfig.Model = "claude-sonnet-4-20250514"
	}
	if claudeConfig.Timeout == 0 {
		claudeConfig.Timeout = 120
	}
	m.RegisterWithTimeout(ProviderClaude, NewClaudeProvider(claudeConfig), time.Duration(claudeConfig.Timeout)*time.Second)

	return m
}
