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
│  │  • RAG 語法查詢     │         │  • /preview 表單預覽    │   │
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
│  │  Events CRUD (新增)                                          ││
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
│                    Python Agent 專屬                             │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  ChromaDB - RAG (DSL 語法文件、範例查詢)                     ││
│  └─────────────────────────────────────────────────────────────┘│
│  ┌─────────────────────────────────────────────────────────────┐│
│  │  Web Search - 搜尋類似賽事參考                               ││
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
├── cmd/rules-engine-demo/           # Go API Server (現有 + 擴充)
│
├── internal/
│   ├── rules/                       # Rule Engine (現有)
│   ├── config/                      # 配置管理 (新增)
│   ├── store/                       # PostgreSQL 存取層 (新增)
│   └── api/                         # Events CRUD handlers (新增)
│
├── ai-poc-python/                   # Python Agent (新增)
│   ├── app.py                       # Streamlit 入口
│   ├── requirements.txt
│   ├── .env.example
│   │
│   ├── agent/
│   │   ├── core.py                  # Agent 主邏輯
│   │   └── prompts.py               # System prompts
│   │
│   ├── tools/
│   │   ├── events.py                # Events CRUD (HTTP client)
│   │   ├── validation.py            # validate_dsl, calculate_price
│   │   └── knowledge.py             # query_dsl_docs, web_search
│   │
│   ├── rag/
│   │   ├── embeddings.py            # ChromaDB 操作
│   │   └── documents/               # DSL 語法文件
│   │
│   └── config/
│       └── settings.py
│
├── docker-compose.yml               # 新增
├── Dockerfile                       # 新增
│
├── examples/                        # DSL 範例 (現有)
└── docs/                            # 文件 (現有)
```

---

## 開發階段

### Phase 1: Go API Server 擴充

**目標**：新增 Events CRUD API + PostgreSQL + Docker 環境

| 任務 | 說明 |
|------|------|
| 1.1 | 建立 `docker-compose.yml` (PostgreSQL + API) |
| 1.2 | 建立 `Dockerfile` |
| 1.3 | 新增 `internal/config` 配置管理 |
| 1.4 | 新增 `internal/store` PostgreSQL 存取層 |
| 1.5 | 新增 `internal/api` Events CRUD handlers |
| 1.6 | 修改 `main.go` 整合新功能 |
| 1.7 | 測試 API endpoints |

**驗收標準**：
- `docker-compose up` 可啟動完整環境
- Events CRUD API 正常運作
- 現有頁面 (`/`, `/editor`, `/preview`) 正常運作

---

### Phase 2: Python Agent 基礎

**目標**：建立 Streamlit + LangChain Agent 基礎架構

| 任務 | 說明 |
|------|------|
| 2.1 | 建立 `ai-poc-python/` 專案結構 |
| 2.2 | Streamlit chat UI |
| 2.3 | LangChain Agent 骨架 |
| 2.4 | Events CRUD tools (HTTP client) |
| 2.5 | validate_dsl / calculate_price tools |

**驗收標準**：
- 可透過對話建立新賽事
- 可查詢/修改現有賽事
- 可驗證 DSL 並計算價格

---

### Phase 3: RAG 整合

**目標**：加入 DSL 語法知識庫

| 任務 | 說明 |
|------|------|
| 3.1 | 整理 DSL 語法文件 (markdown) |
| 3.2 | ChromaDB 設置與 embedding |
| 3.3 | query_dsl_docs tool |
| 3.4 | 改善 DSL 生成品質 |

**驗收標準**：
- Agent 可正確查詢 DSL 語法
- 生成的 DSL 符合規範

---

### Phase 4: 完善功能

**目標**：加入額外功能與優化

| 任務 | 說明 |
|------|------|
| 4.1 | Web Search tool |
| 4.2 | 對話流程優化 |
| 4.3 | 錯誤處理與提示改善 |
| 4.4 | UI 優化 |

---

## Agent Tools 總覽

### Events CRUD (呼叫 Go API)

| Tool | API Endpoint | 說明 |
|------|--------------|------|
| `create_event` | POST /api/events | 建立新賽事 |
| `search_events` | GET /api/events?q=... | 搜尋賽事 |
| `get_event` | GET /api/events/:id | 取得完整 DSL |
| `update_event` | PUT /api/events/:id | 更新賽事 |
| `delete_event` | DELETE /api/events/:id | 刪除賽事 |

### Validation & Calculation (呼叫 Go API)

| Tool | API Endpoint | 說明 |
|------|--------------|------|
| `validate_dsl` | POST /api/events/:id/validate | 驗證 DSL 語法 |
| `calculate_price` | POST /api/events/:id/calculate | 計算報名費用 |

### Knowledge Query (Python 內部)

| Tool | 說明 |
|------|------|
| `query_dsl_docs` | RAG 查詢 DSL 語法文件 |
| `web_search` | 搜尋類似賽事參考 |

---

## 對話範例

### 建立新賽事

```
主辦方: 我要建立一個鐵人三項賽事

Agent: 好的！請問這個賽事有哪些組別和價格？

主辦方: 標準組3000元、半程組2000元、體驗組1500元

Agent: 了解。請問有什麼優惠方案嗎？例如早鳥、團報、學生優惠等？

主辦方: 早鳥到3月底前報名打9折，5人以上團報再折500

Agent: 報名時需要收集哪些資料？

主辦方: 姓名、性別、生日、聯絡方式、緊急聯絡人、T-shirt尺寸

Agent: [呼叫 create_event]
       [呼叫 validate_dsl]

       ✅ 已建立賽事「鐵人三項賽事」
       - ID: 550e8400-e29b-41d4-a716-446655440000
       - 3 個組別定價規則
       - 2 個優惠規則（早鳥、團報）
       - 6 個表單欄位

       要預覽表單或調整內容嗎？
```

### 修改賽事

```
主辦方: 我想修改「2026大湖路跑」的10K組報名費，從1200改成1000

Agent: [呼叫 search_events("2026大湖路跑")]
       [呼叫 get_event(id)]

       找到賽事，目前 10K組 報名費為 $1,200。
       確認要修改為 $1,000 嗎？

主辦方: 對，確認

Agent: [呼叫 update_event]
       [呼叫 validate_dsl]

       ✅ 已更新！
       - 10K組報名費：$1,200 → $1,000
       - DSL 驗證通過
```

---

## Docker Compose 服務

```yaml
services:
  postgres:      # PostgreSQL 資料庫
  api:           # Go API Server (:8080)
  streamlit:     # Python Agent UI (:8501)
  chromadb:      # RAG 向量資料庫 (:8000)
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
ANTHROPIC_API_KEY=sk-...
CHROMA_HOST=chromadb
CHROMA_PORT=8000
```

---

## 待確認事項

- [ ] LLM 選擇：Claude API / OpenAI / Ollama
- [ ] Web Search 服務：Tavily / DuckDuckGo / SerpAPI
