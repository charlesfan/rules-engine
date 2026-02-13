# AI 賽事上架助手 POC 開發計劃

## 專案目標

透過自然語言對話，協助賽事主辦方產生完整的 DSL JSON（包含 FormSchema），讓前端可以動態渲染報名表單。

---

## 系統架構

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端介面                                 │
│                                                                  │
│  ┌─────────────────────┐         ┌─────────────────────────┐   │
│  │     Streamlit       │         │     Go 內建頁面          │   │
│  │     :8501           │         │     :8080               │   │
│  │                     │         │                         │   │
│  │  • AI 對話建立賽事  │         │  • /        規則建立器   │   │
│  │  • DSL CRUD 操作    │         │  • /editor  JSON 編輯   │   │
│  │  • RAG 意圖判斷     │         │  • /preview 表單預覽    │   │
│  └──────────┬──────────┘         └────────────┬────────────┘   │
│             │                                  │                │
└─────────────┼──────────────────────────────────┼────────────────┘
              │              HTTP                │
              └──────────────┬───────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Go API Server :8080                         │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  Events CRUD                                                 ││
│  │  POST   /api/events              建立賽事                    ││
│  │  GET    /api/events              列出/搜尋賽事               ││
│  │  GET    /api/events/:id          取得單一賽事                ││
│  │  PUT    /api/events/:id          更新賽事                    ││
│  │  DELETE /api/events/:id          刪除賽事                    ││
│  │  POST   /api/events/:id/validate 驗證 DSL                    ││
│  │  POST   /api/events/:id/calculate 計算價格                   ││
│  ├─────────────────────────────────────────────────────────────┤│
│  │  Rules (現有)                                                ││
│  │  POST   /api/rules/validate      驗證 DSL 語法               ││
│  │  POST   /api/rules/evaluate      評估規則                    ││
│  │  GET    /api/rules/examples      取得範例                    ││
│  └─────────────────────────────────────────────────────────────┘│
│                             │                                    │
│  ┌──────────────────────────┴──────────────────────────────────┐│
│  │                      Rule Engine                             ││
│  │                (parser/evaluator/calculator)                 ││
│  └──────────────────────────────────────────────────────────────┘│
│                             │                                    │
│  ┌──────────────────────────┴──────────────────────────────────┐│
│  │                      PostgreSQL                              ││
│  └──────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Python Agent (RAG 架構)                       │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  ChromaDB - 意圖判斷 + Prompt 片段檢索                       ││
│  │  • 用戶訊息 → 意圖分類                                       ││
│  │  • 意圖 → 對應的 Prompt 片段                                 ││
│  └─────────────────────────────────────────────────────────────┘│
│                             │                                    │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  動態 Prompt 組合                                            ││
│  │  BASE_PROMPT + INTENT_PROMPT + DSL_SPEC_PROMPT              ││
│  └─────────────────────────────────────────────────────────────┘│
│                             │                                    │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  LangChain ReAct Agent + Tools                              ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

---

## 技術棧

| 層級 | 技術 |
|------|------|
| UI | Streamlit |
| Agent | LangChain + Claude API |
| RAG | LangChain + ChromaDB |
| 主資料庫 | PostgreSQL (JSONB) |
| API Server | Go + Gin |
| DSL 驗證/計算 | Go Rule Engine |
| 容器化 | Docker Compose |

---

## 專案結構

```
rules-engine/
│
├── cmd/rules-engine-demo/           # Go API Server
│
├── internal/
│   ├── rules/                       # Rule Engine (現有)
│   ├── config/                      # 配置管理
│   ├── store/                       # PostgreSQL 存取層
│   └── api/                         # Events CRUD handlers
│
├── ai-poc-python/                   # Python Agent
│   ├── app.py                       # Streamlit 入口
│   ├── requirements.txt
│   ├── .env.example
│   │
│   ├── agent/
│   │   ├── core.py                  # Agent 主邏輯
│   │   ├── prompts.py               # System Prompt（完整版）
│   │   └── prompt_fragments/        # Prompt 片段（Phase 3）
│   │       ├── base.py
│   │       ├── intents/
│   │       │   ├── create_event.py
│   │       │   ├── update_event.py
│   │       │   └── ...
│   │       └── dsl_specs/
│   │           ├── pricing_rules.py
│   │           ├── validation_rules.py
│   │           └── form_schema.py
│   │
│   ├── tools/
│   │   ├── http_client.py           # HTTP Client
│   │   └── events.py                # Events CRUD Tools
│   │
│   ├── rag/
│   │   ├── intent_classifier.py     # 意圖分類器（Phase 3）
│   │   ├── prompt_retriever.py      # Prompt 檢索器（Phase 3）
│   │   ├── embeddings.py            # ChromaDB 操作
│   │   └── documents/               # 意圖定義文件
│   │
│   └── config/
│       └── settings.py
│
├── docker-compose.yml
├── Dockerfile
│
├── examples/                        # DSL 範例
└── docs/                            # 文件
```

---

## 開發階段

### Phase 1: Go API Server 擴充 ✅ 已完成

**目標**：新增 Events CRUD API + PostgreSQL + Docker 環境

| 任務 | 狀態 |
|------|------|
| docker-compose.yml | ✅ |
| Dockerfile (Go) | ✅ |
| internal/config | ✅ |
| internal/store | ✅ |
| internal/api | ✅ |
| 整合到 main.go | ✅ |

---

### Phase 2: Python Agent 基礎 ✅ 已完成

**目標**：建立 Streamlit + LangChain Agent 基礎架構

| 任務 | 狀態 |
|------|------|
| 專案結構 & requirements.txt | ✅ |
| config/settings.py | ✅ |
| tools/http_client.py | ✅ |
| tools/events.py (7 個 Tools) | ✅ |
| agent/core.py | ✅ |
| agent/prompts.py (基礎版) | ✅ |
| app.py (Streamlit UI) | ✅ |
| Dockerfile (Python) | ✅ |

---

### Phase 2.5: 修正問題 + 完善 System Prompt ⏳ 進行中

**目標**：讓基本功能可以正常運作

| 任務 | 說明 |
|------|------|
| 2.5.1 | 修正 `create_react_agent` 參數錯誤 |
| 2.5.2 | 豐富 System Prompt，加入完整 DSL 規格與範例 |
| 2.5.3 | 測試基本對話功能（建立、修改、查詢賽事） |

**驗收標準**：
- Agent 可正確生成 DSL
- 建立賽事功能正常運作

---

### Phase 3: RAG 意圖判斷 + 動態 Prompt

**目標**：使用 RAG 做意圖判斷，根據意圖動態組合 Prompt

#### 架構

```
User Message
    │
    ▼
┌─────────────────────────────────────────┐
│         ChromaDB 意圖判斷                │
│                                         │
│  "我想建立賽事" → intent: create_event  │
│  "修改價格"     → intent: update_pricing│
│  "加優惠"       → intent: add_discount  │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│         動態 Prompt 組合                 │
│                                         │
│  BASE_PROMPT                            │
│  + INTENT_PROMPTS[intent]               │
│  + DSL_SPEC_PROMPTS[related_specs]      │
└─────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────┐
│         LangChain Agent 執行            │
└─────────────────────────────────────────┘
```

#### 任務

| 任務 | 說明 |
|------|------|
| 3.1 | 拆分 Phase 2.5 的 System Prompt 為多個片段 |
| 3.2 | 定義意圖類別與對應的 Prompt 片段 |
| 3.3 | ChromaDB 設置 |
| 3.4 | 實作意圖分類器 (intent_classifier.py) |
| 3.5 | 實作 Prompt 檢索與組合 (prompt_retriever.py) |
| 3.6 | 對話狀態管理（多輪對話） |
| 3.7 | 測試與優化 |

#### 意圖分類

| 意圖 | 觸發詞範例 | 需要的 Prompt 片段 |
|------|-----------|-------------------|
| create_event | 建立賽事、新增活動 | 建立流程 + pricing_rules + form_schema |
| search_event | 查詢、找、列出 | 搜尋說明 |
| update_pricing | 修改價格、改費用 | pricing_rules 規格 |
| add_discount | 加優惠、折扣 | discount 規格 |
| update_form | 修改欄位、加欄位 | form_schema 規格 |
| delete_event | 刪除、移除 | 刪除確認流程 |
| calculate_price | 算價格、預覽費用 | context 格式說明 |

**驗收標準**：
- RAG 意圖判斷準確率 > 80%
- Prompt 根據意圖動態組合
- Token 使用量比完整 Prompt 減少 30%+

---

### Phase 4: 完善功能

**目標**：優化與額外功能

| 任務 | 說明 |
|------|------|
| 4.1 | Web Search tool（搜尋類似賽事參考） |
| 4.2 | 對話流程優化（更好的引導） |
| 4.3 | 錯誤處理改善 |
| 4.4 | UI 優化（表單預覽整合） |
| 4.5 | 性能測試與優化 |

---

## Agent Tools 總覽

### Events CRUD (呼叫 Go API)

| Tool | API Endpoint | 說明 |
|------|--------------|------|
| `search_events` | GET /api/events?q=... | 搜尋賽事 |
| `get_event` | GET /api/events/:id | 取得完整 DSL |
| `create_event` | POST /api/events | 建立新賽事 |
| `update_event` | PUT /api/events/:id | 更新賽事 |
| `delete_event` | DELETE /api/events/:id | 刪除賽事 |
| `validate_event` | POST /api/events/:id/validate | 驗證 DSL |
| `calculate_price` | POST /api/events/:id/calculate | 計算報名費用 |

---

## Prompt 開發策略

### 階段演進

```
Phase 2.5: 完整 System Prompt（單一大檔案）
    │
    │ 功能驗證完成後
    ▼
Phase 3: 拆分為 Prompt 片段
    │
    │ • BASE_PROMPT（固定）
    │ • INTENT_PROMPTS（按意圖）
    │ • DSL_SPEC_PROMPTS（按 DSL 結構）
    │
    ▼
Phase 3: RAG 意圖判斷 + 動態組合
```

### Prompt 片段設計

```
prompt_fragments/
├── base.py                 # 基礎人設、對話風格
├── intents/
│   ├── create_event.py     # 建立賽事的引導流程
│   ├── update_event.py     # 修改賽事的流程
│   ├── search_event.py     # 搜尋說明
│   └── ...
└── dsl_specs/
    ├── pricing_rules.py    # pricing_rules 完整規格 + 範例
    ├── validation_rules.py # validation_rules 完整規格 + 範例
    ├── form_schema.py      # form_schema 完整規格 + 範例
    └── discount.py         # 折扣規則說明
```

---

## Docker Compose 服務

```yaml
services:
  postgres:      # PostgreSQL 資料庫 (:5432)
  api:           # Go API Server (:8080)
  streamlit:     # Python Agent UI (:8501)
  chromadb:      # RAG 向量資料庫 (:8000) - Phase 3
```

---

## 環境變數

### Go API Server
```
DATABASE_URL=postgres://postgres:postgres@postgres:5432/rules_engine?sslmode=disable
SERVER_PORT=8080
```

### Python Agent
```
GO_API_URL=http://api:8080
ANTHROPIC_API_KEY=sk-ant-...
MODEL_NAME=claude-sonnet-4-20250514
CHROMA_HOST=chromadb
CHROMA_PORT=8000
```

---

## 已確認事項

- [x] LLM 選擇：Claude API
- [x] 資料庫：PostgreSQL
- [x] RAG 用途：意圖判斷 + 動態 Prompt 組合
- [ ] Web Search 服務：待定
