# Development TODO

## Phase 0: Basic Chat + DSL Generation

### Week 1

#### Day 1: Project Setup + LLM Provider ✅ COMPLETED
- [x] Create directory structure
- [x] Implement `pkg/llm/provider.go` - Provider interface
- [x] Implement `pkg/llm/ollama.go` - Ollama implementation
- [x] Implement `pkg/llm/claude.go` - Claude implementation
- [x] Implement `pkg/llm/manager.go` - Provider manager
- [x] Add configuration loading (`config/config.yaml`)
- [x] Test LLM provider switching (10 unit tests passing)

#### Day 2: CLI Framework + Commands ✅ COMPLETED
- [x] Implement `cmd/chatbot/main.go` - Entry point with memory integration
- [x] Implement `/model` command - Switch LLM provider
- [x] Implement `/status` command - Show current status (with memory stats)
- [x] Implement `/clear` command - Clear conversation
- [x] Implement `/help` command - Show help
- [x] Implement `internal/chatbot/conversation/memory.go` - Progressive summarization
- [x] Implement `internal/chatbot/conversation/compressor.go` - LLM-based compression
- [x] Add 14 unit tests for conversation module
- [ ] Test CLI interaction with live LLM

#### Day 3: Intent Detection + Prompts ✅ COMPLETED
- [x] Implement `internal/chatbot/agent/prompts.go` - Prompt templates
- [x] Implement `internal/chatbot/agent/intent.go` - Intent detection (two-stage)
- [x] Add 13 unit tests for intent detection
- [x] Test intent detection with various inputs

#### Day 4: DSL Generation + Validation ✅ COMPLETED
- [x] Implement `internal/chatbot/tools/dsl_generator.go` - DSL generation
- [x] Implement `internal/chatbot/tools/dsl_validator.go` - Integrate Rule Engine
- [x] Implement `internal/chatbot/tools/clarifier.go` - Clarification question logic
- [x] Add 35 unit tests for DSL tools

#### Day 5: Integration + Testing ✅ COMPLETED
- [x] Implement `internal/chatbot/agent/agent.go` - Main agent orchestration
- [x] Implement `internal/chatbot/agent/agent_test.go` - 21 unit tests
- [x] Move prompts to `internal/chatbot/prompts/` package (resolve import cycle)
- [x] Update `cmd/chatbot/main.go` to use Agent
- [x] Add `/rules` command to show pending rules
- [ ] End-to-end testing with live LLM (manual)
- [ ] Bug fixes and optimization (ongoing)

### Deliverables
- [x] Working terminal chatbot
- [x] LLM provider switching (/model command)
- [x] Two-stage intent detection
- [x] DSL generation with validation
- [x] Clarification questions for ambiguous rules

---

## Phase 1: RAG Enhancement

### Week 2

#### Day 6: Database Setup
- [ ] Create migration files for pgvector tables
- [ ] Run migrations
- [ ] Seed initial rule examples
- [ ] Test database connection

#### Day 7: Embedding Integration
- [ ] Implement `pkg/embedding/provider.go` - Embedding interface
- [ ] Implement `pkg/embedding/ollama.go` - Ollama embedding
- [ ] Test embedding generation

#### Day 8: Vector Search
- [ ] Implement `internal/chatbot/repository/rule_examples.go`
- [ ] Implement similar rule retrieval
- [ ] Integrate retrieval into prompt construction
- [ ] Test RAG-enhanced generation

#### Day 9: Continuous Learning
- [ ] Implement conversation logging
- [ ] Implement user feedback collection
- [ ] Implement auto-save for confirmed DSL
- [ ] Add admin verification workflow (basic)

#### Day 10: Integration + Testing
- [ ] End-to-end RAG testing
- [ ] Performance optimization
- [ ] Bug fixes
- [ ] Documentation update

### Deliverables
- [ ] pgvector database with rule examples
- [ ] Embedding-based similar rule retrieval
- [ ] RAG-enhanced DSL generation
- [ ] Continuous learning (auto-save confirmed DSL)

---

## Future Enhancements (Post-POC)

### Document Processing
- [ ] PDF parsing
- [ ] Image OCR
- [ ] DOCX parsing

### UI Development
- [ ] Web-based chat interface
- [ ] DSL visual editor
- [ ] Admin dashboard for verification

### Advanced Features
- [ ] Multi-turn clarification
- [ ] Rule conflict detection
- [ ] Simulation/preview of rules

---

## Progress Tracking

| Phase | Status | Progress | Notes |
|-------|--------|----------|-------|
| Phase 0 | ✅ Completed | 100% | Day 1-5 completed |
| Phase 1 | Not Started | 0% | Depends on Phase 0 |

Last Updated: 2026-01-19

### Completed Files
- `pkg/llm/provider.go` - Provider interface
- `pkg/llm/ollama.go` - Ollama implementation
- `pkg/llm/claude.go` - Claude implementation
- `pkg/llm/manager.go` - Provider manager with switching
- `pkg/llm/manager_test.go` - 10 unit tests
- `cmd/chatbot/main.go` - CLI entry point with Agent integration
- `config/config.yaml` - Configuration file
- `internal/chatbot/conversation/memory.go` - Progressive summarization
- `internal/chatbot/conversation/compressor.go` - LLM-based compression
- `internal/chatbot/conversation/memory_test.go` - 14 unit tests
- `internal/chatbot/prompts/prompts.go` - Prompt templates (moved from agent)
- `internal/chatbot/agent/intent.go` - Two-stage intent detection
- `internal/chatbot/agent/intent_test.go` - 13 unit tests
- `internal/chatbot/agent/agent.go` - Main agent orchestration
- `internal/chatbot/agent/agent_test.go` - 21 unit tests
- `internal/chatbot/tools/dsl_generator.go` - DSL generation with LLM
- `internal/chatbot/tools/dsl_validator.go` - Rule Engine integration
- `internal/chatbot/tools/clarifier.go` - Clarification question generator
- `internal/chatbot/tools/dsl_generator_test.go` - 12 unit tests
- `internal/chatbot/tools/dsl_validator_test.go` - 12 unit tests
- `internal/chatbot/tools/clarifier_test.go` - 13 unit tests

### Total Unit Tests: 95+
