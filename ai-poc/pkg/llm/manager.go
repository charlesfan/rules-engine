package llm

import (
	"context"
	"fmt"
	"sync"
)

// Manager manages multiple LLM providers and handles switching between them
type Manager struct {
	providers map[ProviderType]Provider
	current   ProviderType
	mu        sync.RWMutex
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers: make(map[ProviderType]Provider),
		current:   ProviderOllama, // Default to Ollama
	}
}

// Register adds a provider to the manager
func (m *Manager) Register(providerType ProviderType, provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[providerType] = provider
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

	// Register Ollama with default config
	m.Register(ProviderOllama, NewOllamaProvider(DefaultOllamaConfig()))

	// Register Claude with default config
	m.Register(ProviderClaude, NewClaudeProvider(DefaultClaudeConfig()))

	return m
}
