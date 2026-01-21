# Technical Specification

## Project Structure

```
ai-poc/
├── README.md                 # Project overview
├── SPEC.md                   # This file - technical specification
├── DATABASE.md               # Database schema for RAG
│
├── cmd/
│   └── chatbot/
│       └── main.go           # CLI entry point
│
├── internal/
│   └── chatbot/
│       ├── agent/            # LLM Agent logic
│       │   ├── agent.go      # Main agent orchestration
│       │   ├── prompts.go    # Prompt templates
│       │   └── intent.go     # Intent detection
│       │
│       ├── conversation/     # Conversation management
│       │   ├── memory.go     # Conversation history
│       │   └── session.go    # Session management
│       │
│       ├── commands/         # CLI commands
│       │   ├── handler.go    # Command router
│       │   ├── model.go      # /model command
│       │   ├── status.go     # /status command
│       │   └── export.go     # /export command
│       │
│       └── tools/            # Agent tools
│           ├── dsl_generator.go   # DSL generation
│           └── dsl_validator.go   # DSL validation (uses Rule Engine)
│
├── pkg/
│   ├── llm/                  # LLM client abstraction
│   │   ├── provider.go       # Provider interface
│   │   ├── ollama.go         # Ollama implementation
│   │   ├── claude.go         # Claude implementation
│   │   └── config.go         # Configuration
│   │
│   └── embedding/            # Embedding client (Phase 1)
│       ├── provider.go       # Provider interface
│       └── ollama.go         # Ollama embedding
│
└── config/
    └── config.yaml           # Default configuration
```

## LLM Provider Interface

### Interface Definition

```go
// pkg/llm/provider.go

package llm

import (
    "context"
    "encoding/json"
)

// Message represents a chat message
type Message struct {
    Role    Role   `json:"role"`
    Content string `json:"content"`
}

type Role string

const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
)

// Provider defines the LLM provider interface
type Provider interface {
    // Name returns the provider name
    Name() string

    // Chat sends messages and returns a response
    Chat(ctx context.Context, messages []Message) (string, error)

    // ChatWithJSON sends messages and returns a JSON response
    ChatWithJSON(ctx context.Context, messages []Message, schema any) (json.RawMessage, error)

    // Available checks if the provider is available
    Available(ctx context.Context) bool
}

// ProviderType represents supported LLM providers
type ProviderType string

const (
    ProviderOllama ProviderType = "ollama"
    ProviderClaude ProviderType = "claude"
)
```

### Ollama Implementation

```go
// pkg/llm/ollama.go

package llm

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type OllamaConfig struct {
    BaseURL string `yaml:"base_url"`
    Model   string `yaml:"model"`
}

type OllamaProvider struct {
    config OllamaConfig
    client *http.Client
}

func NewOllamaProvider(config OllamaConfig) *OllamaProvider {
    if config.BaseURL == "" {
        config.BaseURL = "http://localhost:11434"
    }
    if config.Model == "" {
        config.Model = "llama3.1:8b"
    }
    return &OllamaProvider{
        config: config,
        client: &http.Client{},
    }
}

func (p *OllamaProvider) Name() string {
    return fmt.Sprintf("ollama/%s", p.config.Model)
}

func (p *OllamaProvider) Chat(ctx context.Context, messages []Message) (string, error) {
    // Implementation details...
}

func (p *OllamaProvider) Available(ctx context.Context) bool {
    // Check if Ollama is running
    resp, err := p.client.Get(p.config.BaseURL + "/api/tags")
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}
```

### Claude Implementation

```go
// pkg/llm/claude.go

package llm

import (
    "context"
    "encoding/json"
    "os"

    "github.com/anthropics/anthropic-sdk-go"
)

type ClaudeConfig struct {
    APIKey string `yaml:"api_key"`
    Model  string `yaml:"model"`
}

type ClaudeProvider struct {
    config ClaudeConfig
    client *anthropic.Client
}

func NewClaudeProvider(config ClaudeConfig) *ClaudeProvider {
    if config.APIKey == "" {
        config.APIKey = os.Getenv("ANTHROPIC_API_KEY")
    }
    if config.Model == "" {
        config.Model = "claude-sonnet-4-20250514"
    }
    return &ClaudeProvider{
        config: config,
        client: anthropic.NewClient(),
    }
}

func (p *ClaudeProvider) Name() string {
    return fmt.Sprintf("claude/%s", p.config.Model)
}

func (p *ClaudeProvider) Chat(ctx context.Context, messages []Message) (string, error) {
    // Implementation using Anthropic SDK...
}

func (p *ClaudeProvider) Available(ctx context.Context) bool {
    return p.config.APIKey != ""
}
```

### Provider Manager

```go
// pkg/llm/manager.go

package llm

import (
    "context"
    "fmt"
    "sync"
)

type Manager struct {
    providers map[ProviderType]Provider
    current   ProviderType
    mu        sync.RWMutex
}

func NewManager() *Manager {
    return &Manager{
        providers: make(map[ProviderType]Provider),
        current:   ProviderOllama, // Default to Ollama
    }
}

func (m *Manager) Register(providerType ProviderType, provider Provider) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.providers[providerType] = provider
}

func (m *Manager) Switch(providerType ProviderType) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, ok := m.providers[providerType]; !ok {
        return fmt.Errorf("provider %s not registered", providerType)
    }
    m.current = providerType
    return nil
}

func (m *Manager) Current() Provider {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.providers[m.current]
}

func (m *Manager) CurrentType() ProviderType {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.current
}

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
```

## Configuration

```yaml
# config/config.yaml

llm:
  default_provider: "ollama"

  ollama:
    base_url: "http://localhost:11434"
    model: "llama3.1:8b"
    # Alternative models:
    # model: "qwen2.5:14b"  # Better Chinese support

  claude:
    api_key: "${ANTHROPIC_API_KEY}"
    model: "claude-sonnet-4-20250514"
    # Alternative models:
    # model: "claude-3-5-haiku-20241022"  # Cheaper

embedding:
  provider: "ollama"
  model: "nomic-embed-text"
  base_url: "http://localhost:11434"

# Phase 1: Database configuration
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "${DB_PASSWORD}"
  database: "dsl_chatbot"
```

## Prompt Templates

### System Prompt

```go
// internal/chatbot/agent/prompts.go

const SystemPrompt = `你是一個專門處理賽事報名規則的 AI 助手。

## 你的職責
1. 幫助用戶將賽事報名規則轉換成系統可執行的 DSL
2. 確認規則的完整性，主動詢問可能遺漏的細節
3. 用清楚易懂的方式解釋生成的 DSL

## 對話原則
1. 如果用戶的輸入不明確，先確認是否為賽事報名規則
2. 生成 DSL 前，確認所有模糊的細節
3. 生成後，用人話解釋每條規則
4. 主動提示常見的遺漏規則

## DSL 語法規範
{{DSL_SYNTAX}}

## 規則範例
{{RULE_EXAMPLES}}
`

const IntentDetectionPrompt = `分析以下用戶輸入，判斷是否為賽事報名規則。

用戶輸入: "{{USER_INPUT}}"

請回答：
1. 類型: [rule/chat/ambiguous]
2. 信心度: [high/medium/low]
3. 如果是 ambiguous，建議的確認問題是什麼？

以 JSON 格式回答。
`

const DSLGenerationPrompt = `根據以下確認的規則資訊，生成 DSL。

規則描述:
{{RULE_DESCRIPTION}}

已確認的細節:
{{CONFIRMED_DETAILS}}

請生成符合規範的 DSL JSON，並用繁體中文解釋每條規則的含義。
`
```

## CLI Main Loop

```go
// cmd/chatbot/main.go

package main

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "strings"

    "backend/ai-poc/internal/chatbot/agent"
    "backend/ai-poc/internal/chatbot/commands"
    "backend/ai-poc/internal/chatbot/conversation"
    "backend/ai-poc/pkg/llm"
)

func main() {
    ctx := context.Background()

    // Initialize LLM Manager
    manager := llm.NewManager()
    manager.Register(llm.ProviderOllama, llm.NewOllamaProvider(llm.OllamaConfig{}))
    manager.Register(llm.ProviderClaude, llm.NewClaudeProvider(llm.ClaudeConfig{}))

    // Initialize Agent
    chatAgent := agent.NewDSLAgent(manager)

    // Initialize Conversation Memory
    memory := conversation.NewMemory()

    // Initialize Command Handler
    cmdHandler := commands.NewHandler(manager)

    // Print welcome message
    printWelcome(manager)

    // Main loop
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("\n你: ")
        if !scanner.Scan() {
            break
        }

        input := strings.TrimSpace(scanner.Text())
        if input == "" {
            continue
        }

        // Check for exit
        if input == "exit" || input == "quit" {
            fmt.Println("再見！")
            break
        }

        // Check for commands (starts with /)
        if strings.HasPrefix(input, "/") {
            result := cmdHandler.Handle(ctx, input)
            fmt.Printf("\n%s\n", result)
            continue
        }

        // Process conversation
        response, err := chatAgent.Chat(ctx, memory, input)
        if err != nil {
            fmt.Printf("\n錯誤: %v\n", err)
            continue
        }

        fmt.Printf("\nAI: %s\n", response)
    }
}

func printWelcome(manager *llm.Manager) {
    fmt.Println("╔════════════════════════════════════════════════════════╗")
    fmt.Println("║          DSL Chatbot - 賽事規則轉換助手                ║")
    fmt.Println("╠════════════════════════════════════════════════════════╣")
    fmt.Printf("║  當前模型: %-44s ║\n", manager.Current().Name())
    fmt.Println("║                                                        ║")
    fmt.Println("║  指令:                                                 ║")
    fmt.Println("║    /model   - 切換 LLM (Ollama ↔ Claude)              ║")
    fmt.Println("║    /status  - 顯示當前狀態                            ║")
    fmt.Println("║    /clear   - 清除對話歷史                            ║")
    fmt.Println("║    /help    - 顯示說明                                ║")
    fmt.Println("║    exit     - 結束程式                                ║")
    fmt.Println("╚════════════════════════════════════════════════════════╝")
}
```

## Command Handler

```go
// internal/chatbot/commands/handler.go

package commands

import (
    "context"
    "fmt"
    "strings"

    "backend/ai-poc/pkg/llm"
)

type Handler struct {
    manager *llm.Manager
}

func NewHandler(manager *llm.Manager) *Handler {
    return &Handler{manager: manager}
}

func (h *Handler) Handle(ctx context.Context, input string) string {
    parts := strings.Fields(input)
    if len(parts) == 0 {
        return "無效的指令"
    }

    cmd := strings.ToLower(parts[0])

    switch cmd {
    case "/model":
        return h.handleModel(ctx)
    case "/status":
        return h.handleStatus(ctx)
    case "/clear":
        return h.handleClear()
    case "/help":
        return h.handleHelp()
    default:
        return fmt.Sprintf("未知指令: %s\n輸入 /help 查看可用指令", cmd)
    }
}

func (h *Handler) handleModel(ctx context.Context) string {
    current := h.manager.CurrentType()
    available := h.manager.ListAvailable(ctx)

    var newProvider llm.ProviderType
    if current == llm.ProviderOllama {
        newProvider = llm.ProviderClaude
    } else {
        newProvider = llm.ProviderOllama
    }

    // Check if target provider is available
    isAvailable := false
    for _, p := range available {
        if p == newProvider {
            isAvailable = true
            break
        }
    }

    if !isAvailable {
        return fmt.Sprintf("❌ %s 目前不可用\n請確認服務已啟動或 API Key 已設定", newProvider)
    }

    if err := h.manager.Switch(newProvider); err != nil {
        return fmt.Sprintf("❌ 切換失敗: %v", err)
    }

    return fmt.Sprintf("✅ 已切換到 %s", h.manager.Current().Name())
}

func (h *Handler) handleStatus(ctx context.Context) string {
    current := h.manager.Current()
    available := h.manager.ListAvailable(ctx)

    var sb strings.Builder
    sb.WriteString("📊 系統狀態\n")
    sb.WriteString("─────────────────────\n")
    sb.WriteString(fmt.Sprintf("當前模型: %s\n", current.Name()))
    sb.WriteString(fmt.Sprintf("可用模型: %v\n", available))

    return sb.String()
}

func (h *Handler) handleClear() string {
    // Will be implemented with memory
    return "✅ 對話歷史已清除"
}

func (h *Handler) handleHelp() string {
    return `📖 可用指令
─────────────────────
/model   - 切換 LLM 提供者 (Ollama ↔ Claude)
/status  - 顯示當前系統狀態
/clear   - 清除對話歷史
/export  - 匯出已生成的 DSL
/help    - 顯示此說明
exit     - 結束程式

💡 使用提示
─────────────────────
1. 直接輸入賽事規則描述即可開始
2. AI 會自動判斷並詢問需要釐清的細節
3. 確認後會生成 DSL 並驗證
`
}
```

## Dependencies

```go
// go.mod additions

require (
    github.com/anthropics/anthropic-sdk-go v0.2.0
    gopkg.in/yaml.v3 v3.0.1
)
```

## Environment Variables

```bash
# .env (for development)

# Claude API (optional, only needed when using Claude)
ANTHROPIC_API_KEY=sk-ant-xxx

# Ollama (optional, defaults to localhost:11434)
OLLAMA_BASE_URL=http://localhost:11434

# Database (Phase 1)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=dsl_chatbot
```
