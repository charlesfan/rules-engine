# AI 賽事上架助手

透過自然語言對話，協助賽事主辦方建立和管理賽事報名規則。

## 快速開始

### 1. 安裝依賴

```bash
cd ai-poc-python

# 建立虛擬環境（建議）
python -m venv venv
source venv/bin/activate  # macOS/Linux
# venv\Scripts\activate   # Windows

# 安裝套件
pip install -r requirements.txt
```

### 2. 設定環境變數

```bash
# 複製範本
cp .env.example .env

# 編輯 .env，填入你的 API key
# ANTHROPIC_API_KEY=sk-ant-xxx
```

### 3. 啟動 Go API Server

```bash
# 在專案根目錄
docker compose up -d postgres
go run ./cmd/rules-engine-demo
```

### 4. 啟動 Streamlit

```bash
cd ai-poc-python
streamlit run app.py
```

打開瀏覽器 http://localhost:8501

## 專案結構

```
ai-poc-python/
├── app.py              # Streamlit 入口
├── requirements.txt    # Python 依賴
├── .env.example        # 環境變數範本
│
├── agent/
│   ├── core.py         # LangChain Agent
│   └── prompts.py      # System Prompt
│
├── tools/
│   ├── http_client.py  # HTTP Client
│   └── events.py       # Events CRUD Tools
│
├── config/
│   └── settings.py     # 配置管理
│
└── rag/                # Phase 3 - RAG
    └── documents/
```

## 使用方式

在聊天介面中，你可以：

- **列出賽事**：「列出所有賽事」
- **搜尋賽事**：「搜尋馬拉松相關賽事」
- **建立賽事**：「我想建立一個路跑賽事」
- **修改賽事**：「修改 2026大湖路跑 的報名費」
- **刪除賽事**：「刪除這個賽事」
- **計算價格**：「幫我算一下全馬3人團報的費用」

## 技術棧

- **UI**: Streamlit
- **Agent**: LangChain + Claude API
- **HTTP Client**: httpx
- **Config**: Pydantic Settings
