package llm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProvider is a mock LLM provider for testing
type MockProvider struct {
	name      string
	available bool
	response  string
}

func NewMockProvider(name string, available bool) *MockProvider {
	return &MockProvider{
		name:      name,
		available: available,
		response:  "Mock response from " + name,
	}
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{
		Content:    p.response,
		TokensUsed: 10,
	}, nil
}

func (p *MockProvider) Available(ctx context.Context) bool {
	return p.available
}

func TestManager_Register(t *testing.T) {
	m := NewManager()

	mock := NewMockProvider("test-provider", true)
	m.Register(ProviderOllama, mock)

	registered := m.ListRegistered()
	assert.Contains(t, registered, ProviderOllama)
}

func TestManager_Switch(t *testing.T) {
	m := NewManager()

	m.Register(ProviderOllama, NewMockProvider("ollama", true))
	m.Register(ProviderClaude, NewMockProvider("claude", true))

	// Default should be Ollama
	assert.Equal(t, ProviderOllama, m.CurrentType())

	// Switch to Claude
	err := m.Switch(ProviderClaude)
	require.NoError(t, err)
	assert.Equal(t, ProviderClaude, m.CurrentType())

	// Switch back to Ollama
	err = m.Switch(ProviderOllama)
	require.NoError(t, err)
	assert.Equal(t, ProviderOllama, m.CurrentType())
}

func TestManager_Switch_NotRegistered(t *testing.T) {
	m := NewManager()

	err := m.Switch(ProviderClaude)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestManager_Toggle(t *testing.T) {
	ctx := context.Background()
	m := NewManager()

	m.Register(ProviderOllama, NewMockProvider("ollama", true))
	m.Register(ProviderClaude, NewMockProvider("claude", true))

	// Start with Ollama
	assert.Equal(t, ProviderOllama, m.CurrentType())

	// Toggle should switch to Claude
	newType, err := m.Toggle(ctx)
	require.NoError(t, err)
	assert.Equal(t, ProviderClaude, newType)
	assert.Equal(t, ProviderClaude, m.CurrentType())

	// Toggle again should switch back to Ollama
	newType, err = m.Toggle(ctx)
	require.NoError(t, err)
	assert.Equal(t, ProviderOllama, newType)
}

func TestManager_Toggle_OnlyOneAvailable(t *testing.T) {
	ctx := context.Background()
	m := NewManager()

	m.Register(ProviderOllama, NewMockProvider("ollama", true))
	m.Register(ProviderClaude, NewMockProvider("claude", false)) // Not available

	// Toggle should fail when only one is available
	_, err := m.Toggle(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only one provider available")
}

func TestManager_ListAvailable(t *testing.T) {
	ctx := context.Background()
	m := NewManager()

	m.Register(ProviderOllama, NewMockProvider("ollama", true))
	m.Register(ProviderClaude, NewMockProvider("claude", false))

	available := m.ListAvailable(ctx)
	assert.Len(t, available, 1)
	assert.Contains(t, available, ProviderOllama)
}

func TestManager_Chat(t *testing.T) {
	ctx := context.Background()
	m := NewManager()

	m.Register(ProviderOllama, NewMockProvider("ollama", true))

	req := ChatRequest{
		Messages: []Message{
			NewUserMessage("Hello"),
		},
	}

	resp, err := m.Chat(ctx, req)
	require.NoError(t, err)
	assert.Contains(t, resp.Content, "Mock response")
}

func TestManager_GetStatus(t *testing.T) {
	ctx := context.Background()
	m := NewManager()

	m.Register(ProviderOllama, NewMockProvider("ollama", true))
	m.Register(ProviderClaude, NewMockProvider("claude", false))

	statuses := m.GetStatus(ctx)
	assert.Len(t, statuses, 2)

	// Find Ollama status
	var ollamaStatus, claudeStatus ProviderStatus
	for _, s := range statuses {
		if s.Type == ProviderOllama {
			ollamaStatus = s
		} else if s.Type == ProviderClaude {
			claudeStatus = s
		}
	}

	assert.True(t, ollamaStatus.Available)
	assert.True(t, ollamaStatus.IsCurrent)
	assert.False(t, claudeStatus.Available)
	assert.False(t, claudeStatus.IsCurrent)
}

func TestMessage_Helpers(t *testing.T) {
	sys := NewSystemMessage("You are helpful")
	assert.Equal(t, RoleSystem, sys.Role)
	assert.Equal(t, "You are helpful", sys.Content)

	user := NewUserMessage("Hello")
	assert.Equal(t, RoleUser, user.Role)
	assert.Equal(t, "Hello", user.Content)

	asst := NewAssistantMessage("Hi there")
	assert.Equal(t, RoleAssistant, asst.Role)
	assert.Equal(t, "Hi there", asst.Content)
}

func TestProviderType_Parse(t *testing.T) {
	tests := []struct {
		input    string
		expected ProviderType
		ok       bool
	}{
		{"ollama", ProviderOllama, true},
		{"claude", ProviderClaude, true},
		{"openai", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		pt, ok := ParseProviderType(tt.input)
		assert.Equal(t, tt.ok, ok, "input: %s", tt.input)
		if ok {
			assert.Equal(t, tt.expected, pt)
		}
	}
}
