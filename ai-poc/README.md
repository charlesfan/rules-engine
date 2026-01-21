# DSL Chatbot POC

## Overview

This POC aims to build a conversational AI assistant that helps convert event registration rules into DSL (Domain Specific Language) format.

### Pain Points to Solve

| Problem | Description |
|---------|-------------|
| Complex Rules | Event registration rules are numerous and complex |
| Inconsistent Formats | Source files vary: images (jpg/png), PDF, DOCX, plain text |
| Poor DSL Readability | Generated DSL is large, complex, and hard to modify |
| Manual Maintenance | Current process requires manual DSL editing |

### Goals

1. **Conversational DSL Generation** - Generate DSL through natural language conversation
2. **Two-Stage Intent Detection** - Distinguish between general chat and rule definition
3. **Validation Integration** - Use existing Rule Engine to validate generated DSL
4. **RAG Enhancement** - Improve accuracy with similar rule examples
5. **Continuous Learning** - Save confirmed DSL for future reference

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DSL Chatbot POC Architecture                         │
└─────────────────────────────────────────────────────────────────────────┘

Phase 0:
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   Terminal   │ ──── │   Go CLI     │ ──── │   LLM API   │
│   (stdin)    │      │              │      │ (Switchable) │
└──────────────┘      └──────┬───────┘      └──────────────┘
                             │
                      ┌──────▼───────┐
                      │ Rule Engine  │
                      │  (Existing)  │
                      └──────────────┘

Phase 1 (Add RAG):
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   Terminal   │ ──── │   Go CLI     │ ──── │   LLM API   │
│   (stdin)    │      │              │      │ (Switchable) │
└──────────────┘      └──────┬───────┘      └──────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
       ┌──────▼─────┐ ┌──────▼─────┐ ┌──────▼──────┐
       │ Rule Engine│ │  pgvector  │ │   Ollama    │
       │ (Existing) │ │ (Vector DB)│ │ (Embedding) │
       └────────────┘ └────────────┘ └─────────────┘
```

## Tech Stack

| Component | Choice | Cost | Notes |
|-----------|--------|------|-------|
| **LLM** | Ollama (default) + Claude API (switchable) | Free / Pay-per-use | Switch via `/model` command |
| **Embedding** | Ollama nomic-embed-text | Free | Local, 768 dimensions |
| **Vector DB** | pgvector | Free | Extends existing PostgreSQL |
| **Backend** | Go | - | Consistent with existing codebase |
| **CLI** | Terminal (stdin/stdout) | - | No UI needed for POC |

### LLM Options

| Provider | Model | Cost | Use Case |
|----------|-------|------|----------|
| Ollama | llama3.1:8b | Free | Development, testing |
| Ollama | qwen2.5:14b | Free | Better Chinese support |
| Claude | claude-sonnet-4-20250514 | ~$3/1M input | Production quality |
| Claude | claude-3-5-haiku | ~$0.80/1M input | Cost-effective |

### Embedding Model

```bash
# Install Ollama embedding model
ollama pull nomic-embed-text  # 768 dimensions, ~274MB
```

## Development Phases

### Phase 0: Basic Chat + DSL Generation

**Duration**: ~5 days

**Components**:
1. Terminal chatbot (stdin/stdout)
2. Prompt engineering for DSL generation
3. LLM API integration (switchable: Ollama/Claude)
4. Rule Engine integration for validation

**Features**:
- `/model` command to switch LLM provider
- Two-stage intent detection (is this a rule?)
- DSL generation with validation
- Clarification questions for ambiguous rules

### Phase 1: RAG Enhancement

**Duration**: ~5 days

**Components**:
1. pgvector database setup
2. Embedding integration (Ollama)
3. Retrieval + Generation pipeline
4. Continuous learning (auto-save confirmed DSL)

**Features**:
- Similar rule retrieval
- Context-enhanced generation
- User feedback collection
- Automatic example accumulation

## Conversation Flow

### Intent Detection (Two-Stage)

```
User Input
    │
    ▼
┌─────────────────┐
│  Intent Check   │
│  (Is this a     │
│   rule?)        │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌───────┐  ┌───────────────┐
│ Chat  │  │   Ambiguous   │───────┐
└───┬───┘  └───────────────┘       │
    │                              ▼
    │                    ┌─────────────────┐
    │                    │ Confirm: Is     │
    │                    │ this a rule?    │
    │                    └────────┬────────┘
    │                             │
    │         ┌───────────────────┤
    │         │                   │
    │         ▼                   ▼
    │    ┌─────────┐         ┌─────────┐
    │    │   Yes   │         │   No    │
    │    └────┬────┘         └────┬────┘
    │         │                   │
    │         ▼                   │
    │  ┌──────────────┐           │
    │  │ Rule Process │           │
    │  └──────────────┘           │
    │         │                   │
    ▼         ▼                   ▼
┌─────────────────────────────────────┐
│            Response                 │
└─────────────────────────────────────┘
```

### Rule Processing Flow

```
User describes rule
    │
    ▼
┌─────────────────────────────────────┐
│  Step 1: Rule Understanding (LLM)   │
│  - Extract rule type                │
│  - Identify conditions & actions    │
│  - Mark unclear parts               │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Step 2: Clarification              │
│  - Ask about ambiguous parts        │
│  - Confirm edge cases               │
│  - Confirm rule relationships       │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Step 3: Generate DSL (LLM)         │
│  - Generate based on confirmed info │
│  - Use JSON mode for format         │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Step 4: Validate DSL               │
│  - Call Rule Engine Parser          │
│  - Check syntax correctness         │
│  - If failed, ask LLM to fix        │
└──────────────────┬──────────────────┘
                   │
                   ▼
┌─────────────────────────────────────┐
│  Step 5: Output Result              │
│  - Display DSL                      │
│  - Explain each rule in plain text  │
│  - Suggest possibly missing rules   │
└─────────────────────────────────────┘
```

### Example Conversation

```
┌─────────────────────────────────────────────────────────────┐
│ User: 早鳥報名 3/1 前打 8 折，3 人以上團報再打 9 折         │
├─────────────────────────────────────────────────────────────┤
│ AI: 我理解了以下規則：                                      │
│                                                             │
│ 📋 規則 1: 早鳥優惠                                         │
│    - 條件：報名日期 < 2025/3/1                              │
│    - 優惠：8 折                                             │
│                                                             │
│ 📋 規則 2: 團體優惠                                         │
│    - 條件：團隊人數 >= 3                                    │
│    - 優惠：9 折                                             │
│                                                             │
│ ⚠️ 請確認以下問題：                                         │
│ 1. 早鳥和團報可以同時享用嗎？（疊加 or 擇優）               │
│ 2. 早鳥優惠有名額限制嗎？                                   │
│ 3. 團報的「3人以上」包含 3 人嗎？                           │
├─────────────────────────────────────────────────────────────┤
│ User: 可以疊加，早鳥限前 100 名，3 人包含                   │
├─────────────────────────────────────────────────────────────┤
│ AI: 已更新規則，以下是完整的 DSL：                          │
│                                                             │
│ ```json                                                     │
│ { "rules": [...] }                                          │
│ ```                                                         │
│                                                             │
│ ✅ 已通過語法驗證                                           │
│ 需要我解釋這個 DSL 的內容嗎？                               │
└─────────────────────────────────────────────────────────────┘
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `/model` | Switch LLM provider (Ollama ↔ Claude) |
| `/status` | Show current configuration |
| `/clear` | Clear conversation history |
| `/export` | Export generated DSL to file |
| `/help` | Show available commands |
| `exit` | Exit the chatbot |

## File References

- **Rule Engine**: `internal/rules/` (parser, evaluator, calculator)
- **DSL Types**: `internal/rules/dsl/types.go`
- **Integration Tests**: `internal/rules/integration_test.go`
