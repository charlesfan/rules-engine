package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/charlesfan/rules-engine/ai-poc/internal/chatbot/agent"
	"github.com/charlesfan/rules-engine/ai-poc/internal/chatbot/conversation"
	"github.com/charlesfan/rules-engine/ai-poc/pkg/config"
	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadFromDefaultPaths()
	if err != nil {
		fmt.Printf("Warning: Failed to load config, using defaults: %v\n", err)
		cfg = config.Default()
	}

	// Initialize LLM Manager from config
	manager := llm.NewManagerFromConfig(cfg)

	// Initialize Memory with compressor
	memory := conversation.NewMemory()
	setupCompressor(manager, memory)

	// Initialize Agent with current provider
	chatAgent := agent.NewAgent(manager.Current())

	// Print welcome message
	printWelcome(ctx, manager)

	// Setup readline with auto-completion
	completer := readline.NewPrefixCompleter(
		readline.PcItem("/model"),
		readline.PcItem("/status"),
		readline.PcItem("/rules"),
		readline.PcItem("/clear"),
		readline.PcItem("/help"),
		readline.PcItem("exit"),
		readline.PcItem("quit"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "你: ",
		HistoryFile:       "/tmp/dsl-chatbot-history.txt",
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		fmt.Printf("Failed to initialize readline: %v\n", err)
		return
	}
	defer rl.Close()

	// Main loop
	for {
		// Read input (with multi-line support via backslash)
		input, ok := readMultiLine(rl)
		if !ok {
			break
		}

		input = strings.TrimSpace(input)
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
			handleCommand(ctx, manager, memory, chatAgent, input)
			continue
		}

		// Process with Agent
		handleChat(ctx, manager, memory, chatAgent, input)
	}
}

// readMultiLine reads input with multi-line support using backslash continuation
func readMultiLine(rl *readline.Instance) (string, bool) {
	var lines []string

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if len(lines) > 0 {
					// Cancel multi-line input
					fmt.Println("(已取消)")
					return "", true
				}
				return "", true
			} else if err == io.EOF {
				return "", false
			}
			return "", false
		}

		// Check if line ends with backslash (continuation)
		if strings.HasSuffix(line, "\\") {
			// Remove trailing backslash and add to lines
			lines = append(lines, strings.TrimSuffix(line, "\\"))
			// Change prompt for continuation
			rl.SetPrompt(".. ")
		} else {
			lines = append(lines, line)
			// Reset prompt
			rl.SetPrompt("你: ")
			break
		}
	}

	return strings.Join(lines, "\n"), true
}

// setupCompressor configures the memory compressor using current LLM provider
func setupCompressor(manager *llm.Manager, memory *conversation.Memory) {
	provider := manager.Current()
	if provider != nil {
		compressor := conversation.NewLLMCompressor(provider)
		memory.SetCompressor(compressor)
	}
}

func printWelcome(ctx context.Context, manager *llm.Manager) {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║          DSL Chatbot - 賽事規則轉換助手                ║")
	fmt.Println("╠════════════════════════════════════════════════════════╣")
	fmt.Printf("║  當前模型: %-44s ║\n", manager.Current().Name())
	fmt.Println("║                                                        ║")
	fmt.Println("║  指令:                                                 ║")
	fmt.Println("║    /model   - 切換 LLM (Ollama ↔ Claude)              ║")
	fmt.Println("║    /status  - 顯示當前狀態                            ║")
	fmt.Println("║    /rules   - 顯示已建立的規則                        ║")
	fmt.Println("║    /clear   - 清除對話歷史和規則                      ║")
	fmt.Println("║    /help    - 顯示說明（含快捷鍵）                    ║")
	fmt.Println("║    exit     - 結束程式                                ║")
	fmt.Println("║                                                        ║")
	fmt.Println("║  提示: 行尾加 \\ 可多行輸入，↑↓ 瀏覽歷史            ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	// Check provider availability
	available := manager.ListAvailable(ctx)
	if len(available) == 0 {
		fmt.Println("\n⚠️  警告: 沒有可用的 LLM Provider")
		fmt.Println("   請確認 Ollama 已啟動，或已設定 ANTHROPIC_API_KEY")
	} else {
		fmt.Printf("\n✅ 可用的 Provider: %v\n", available)
	}
	fmt.Println()
}

func handleCommand(ctx context.Context, manager *llm.Manager, memory *conversation.Memory, chatAgent *agent.Agent, input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "/model":
		handleModelSwitch(ctx, manager, memory, chatAgent)
	case "/status":
		handleStatus(ctx, manager, memory, chatAgent)
	case "/rules":
		handleRules(chatAgent)
	case "/clear":
		handleClear(memory, chatAgent)
	case "/help":
		handleHelp()
	default:
		fmt.Printf("未知指令: %s\n輸入 /help 查看可用指令\n", cmd)
	}
}

func handleModelSwitch(ctx context.Context, manager *llm.Manager, memory *conversation.Memory, chatAgent *agent.Agent) {
	currentType := manager.CurrentType()
	fmt.Printf("當前: %s\n", manager.Current().Name())

	// Toggle to next available provider
	newType, err := manager.Toggle(ctx)
	if err != nil {
		fmt.Printf("❌ 切換失敗: %v\n", err)
		return
	}

	if newType == currentType {
		fmt.Println("ℹ️  只有一個可用的 Provider，無法切換")
		return
	}

	// Update compressor to use new provider
	setupCompressor(manager, memory)

	// Note: Agent keeps existing state but will use new provider for new requests
	// We could create a new agent here if needed

	fmt.Printf("✅ 已切換到: %s\n", manager.Current().Name())
}

func handleStatus(ctx context.Context, manager *llm.Manager, memory *conversation.Memory, chatAgent *agent.Agent) {
	fmt.Println("\n📊 系統狀態")
	fmt.Println("─────────────────────────────────────")

	// LLM Status
	fmt.Println("LLM Providers:")
	statuses := manager.GetStatus(ctx)
	for _, s := range statuses {
		current := ""
		if s.IsCurrent {
			current = " ← 當前"
		}
		available := "❌"
		if s.Available {
			available = "✅"
		}
		fmt.Printf("  %s %s (%s)%s\n", available, s.Type, s.Name, current)
	}

	// Agent Status
	fmt.Println("\n規則狀態:")
	fmt.Printf("  Agent 狀態: %s\n", chatAgent.GetState().String())
	fmt.Printf("  已建立規則: %d 條\n", chatAgent.GetRuleCount())

	// Memory Status
	fmt.Println("\n對話記憶:")
	stats := memory.GetStats()
	fmt.Printf("  訊息數量: %d (最近) / ~%d (總計估計)\n", stats.RecentMessageCount, stats.EstimatedTotal)
	fmt.Printf("  已壓縮次數: %d\n", stats.CompressCount)
	fmt.Printf("  有摘要: %v\n", stats.HasSummary)
	fmt.Printf("  對話時長: %s\n", stats.Duration.Round(time.Second))

	if stats.HasSummary {
		fmt.Println("\n當前摘要:")
		fmt.Printf("  %s\n", memory.GetSummary())
	}

	fmt.Println("─────────────────────────────────────")
}

func handleRules(chatAgent *agent.Agent) {
	fmt.Println("\n📋 已建立的規則")
	fmt.Println("─────────────────────────────────────")
	fmt.Println(chatAgent.GetRulesSummary())

	if chatAgent.GetRuleCount() > 0 {
		fmt.Println("\n💡 輸入「生成 DSL」來產生規則")
	}
	fmt.Println("─────────────────────────────────────")
}

func handleClear(memory *conversation.Memory, chatAgent *agent.Agent) {
	memory.Clear()
	chatAgent.Clear()
	fmt.Println("✅ 對話歷史和規則已清除")
}

func handleHelp() {
	fmt.Print(`
📖 可用指令
─────────────────────────────────────
/model   - 切換 LLM 提供者 (Ollama ↔ Claude)
/status  - 顯示當前系統狀態
/rules   - 顯示已建立的規則
/clear   - 清除對話歷史和規則
/help    - 顯示此說明
exit     - 結束程式

⌨️  快捷鍵
─────────────────────────────────────
←/→      - 移動游標
↑/↓      - 瀏覽歷史訊息
Tab      - 自動完成指令
Ctrl+A   - 移到行首
Ctrl+E   - 移到行尾
Ctrl+W   - 刪除前一個詞
Ctrl+U   - 清除整行
Ctrl+C   - 取消輸入/結束程式
\        - 行尾加 \ 可換行繼續輸入

💡 使用流程
─────────────────────────────────────
1. 直接輸入賽事規則描述
   例如: "報名費 1000 元，早鳥優惠 9 折"
2. AI 會自動識別規則並詢問需要釐清的細節
3. 使用 /rules 查看已建立的規則
4. 輸入「生成 DSL」來產生最終規則
5. AI 會生成並驗證 DSL

📝 對話記憶
─────────────────────────────────────
對話會自動壓縮，保留重要上下文
使用 /status 查看當前對話摘要

🔧 環境設定
─────────────────────────────────────
Ollama: 確保 ollama serve 正在運行
Claude: 設定 ANTHROPIC_API_KEY 環境變數
`)
}

func handleChat(ctx context.Context, manager *llm.Manager, memory *conversation.Memory, chatAgent *agent.Agent, input string) {
	provider := manager.Current()
	if provider == nil {
		fmt.Println("❌ 沒有可用的 LLM Provider")
		return
	}

	// Check availability
	if !provider.Available(ctx) {
		fmt.Printf("❌ %s 目前不可用\n", provider.Name())
		fmt.Println("   請確認服務已啟動")
		return
	}

	// Add user message to memory
	memory.AddUserMessage(input)

	// Try to compress if needed
	if err := memory.TryCompress(ctx); err != nil {
		// Log but don't fail - compression is optional
		fmt.Printf("⚠️  對話壓縮失敗: %v\n", err)
	}

	// Show thinking indicator
	fmt.Printf("🤔 思考中 (%s)...\n", provider.Name())

	// Send request with timeout from config
	ctx, cancel := context.WithTimeout(ctx, manager.CurrentTimeout())
	defer cancel()

	start := time.Now()

	// Process with Agent - pass conversation context (recent messages or summary)
	conversationContext := memory.GetConversationContext()
	resp, err := chatAgent.Process(ctx, input, conversationContext)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 錯誤: %v\n", err)
		return
	}

	// Add assistant response to memory
	memory.AddAssistantMessage(resp.Message)

	// Display response
	fmt.Printf("\nAI: %s\n", resp.Message)

	// Show status bar
	stateInfo := ""
	if resp.RuleCount > 0 {
		stateInfo = fmt.Sprintf(" | %d 規則", resp.RuleCount)
	}
	if resp.CanGenerate {
		stateInfo += " | 可生成"
	}
	fmt.Printf("\n─── %.2fs | %s%s ───\n\n", elapsed.Seconds(), resp.State, stateInfo)
}
