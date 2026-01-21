# Chatbot 重構工作記錄

## 問題
原本的 `UnifiedSystemPrompt` 有 25K 字元，小模型（qwen2.5:7b）無法正確處理，會回傳錯誤格式的 JSON。

## 解決方案
**方案 B - 分階段處理**

```
用戶輸入 → Intent Detection (精簡) → 根據 intent 選對應 prompt → 處理結果
```

## 架構設計

```
用戶輸入
    ↓
┌─────────────────────────────────┐
│  Stage 1: Intent Detection      │  ← 精簡 prompt (~500 字)
│  判斷: general_chat / rule_input│
│        / modify_rule / dsl_request
└─────────────────────────────────┘
    ↓ intent
┌─────────────────────────────────┐
│  Stage 2: 根據 intent 選擇 prompt │
├─────────────────────────────────┤
│ general_chat    → ChatPrompt        │  ← ~500 字
│ rule_input      → RuleExtractPrompt │  ← ~3000 字 (含 DSL 結構)
│ modify_rule     → RuleModifyPrompt  │  ← ~2000 字
│ dsl_request     → (直接生成)         │
│ clarify_response→ ClarifyPrompt     │  ← ~1000 字
└─────────────────────────────────┘
    ↓
Agent 處理結果
```

## 待完成工作

- [ ] 建立 `IntentDetectionPrompt`（~500 字）
- [ ] 建立各 intent 專用 prompts：
  - [ ] `GeneralChatPrompt`（~500 字）
  - [ ] `RuleExtractionPrompt`（~3000 字，含 DSL 結構）
  - [ ] `RuleModificationPrompt`（~2000 字）
  - [ ] `ClarificationPrompt`（~1000 字）
- [ ] 重構 `Agent.Process()` 支援多輪 LLM 呼叫
- [ ] 更新測試
- [ ] 移除 debug 輸出

## 相關檔案

- `internal/chatbot/prompts/unified_prompt.go` - 目前的統一 prompt（需重構）
- `internal/chatbot/agent/agent.go` - Agent 主邏輯（需重構）
- `internal/chatbot/agent/llm_response.go` - LLM 回應解析

## 目前使用的模型

- `qwen2.5:7b` - 支援 128K context，中文好，適合這個應用

## 恢復工作

重開後告訴 Claude：「繼續 chatbot 重構，參考 WORK_IN_PROGRESS.md」
