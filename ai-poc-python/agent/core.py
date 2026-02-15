"""
LangChain Agent Core.

這是 Agent 的核心模組，負責：
1. 建立 LLM（Claude）連線
2. 將 Tools 綁定到 Agent
3. 管理對話歷史
4. 執行 Agent Loop
5. RAG 意圖分類 + 動態 Prompt 組合

LangChain Agent 的運作原理：
1. 收到用戶訊息
2. RAG 分類意圖，組合對應的 Prompt
3. LLM 決定是否需要使用 Tool
4. 如果需要，執行 Tool 並將結果給 LLM
5. LLM 根據 Tool 結果生成回應
6. 重複 3-5 直到 LLM 決定直接回應
"""

from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
from langgraph.prebuilt import create_react_agent

from config.settings import settings
from tools.events import all_tools

# RAG components (optional, graceful fallback if not available)
try:
    from rag.intent_classifier import IntentClassifier
    from rag.prompt_retriever import PromptRetriever
    RAG_AVAILABLE = True
except Exception as e:
    print(f"[Agent] RAG not available: {e}")
    RAG_AVAILABLE = False

# Fallback to full prompt if RAG not available
from agent.prompts import SYSTEM_PROMPT


class EventAgent:
    """
    賽事上架助手 Agent。

    使用 LangChain 的 ReAct Agent 架構：
    - ReAct = Reasoning + Acting
    - Agent 會「思考」要做什麼，然後「行動」（呼叫 Tool）

    RAG 增強功能：
    - 意圖分類：判斷用戶想做什麼
    - 動態 Prompt：根據意圖載入對應的規格文件
    """

    def __init__(self, use_rag: bool = True):
        """
        初始化 Agent。

        Args:
            use_rag: 是否啟用 RAG 意圖分類（預設啟用）
        """
        self.llm = ChatAnthropic(
            model=settings.model_name,
            api_key=settings.anthropic_api_key,
            max_tokens=4096,
        )

        # RAG components
        self.use_rag = use_rag and RAG_AVAILABLE
        self.intent_classifier = None
        self.prompt_retriever = None

        if self.use_rag:
            try:
                print("[Agent] Initializing RAG (this may take a while on first run)...")
                self.intent_classifier = IntentClassifier()
                self.prompt_retriever = PromptRetriever()
                print("[Agent] RAG enabled successfully")
            except Exception as e:
                print(f"[Agent] RAG initialization failed: {e}")
                print("[Agent] Falling back to full prompt mode")
                self.use_rag = False

        # Create default agent (will be recreated with dynamic prompt if RAG enabled)
        self.agent = create_react_agent(
            model=self.llm,
            tools=all_tools,
            prompt=SYSTEM_PROMPT,
        )

        # Conversation history
        self.messages: list = []

        # Last classification result (for debugging)
        self.last_intent: dict = {}

    def chat(self, user_message: str) -> str:
        """
        與 Agent 對話。

        Args:
            user_message: 用戶輸入的訊息

        Returns:
            Agent 的回應文字
        """
        # 1. RAG: 分類意圖並組合動態 Prompt
        if self.use_rag:
            self._update_prompt_for_intent(user_message)

        # 2. 加入用戶訊息到歷史
        self.messages.append(HumanMessage(content=user_message))

        # 3. 執行 Agent (ReAct loop)
        result = self.agent.invoke({
            "messages": self.messages,
        })

        # 3. 取得 Agent 回應
        # result["messages"] 包含所有訊息（包括 Tool 呼叫）
        # 我們只需要最後一個 AI 回應
        response_messages = result["messages"]

        # 找到最後一個 AI 訊息
        ai_response = ""
        for msg in reversed(response_messages):
            if isinstance(msg, AIMessage) and msg.content:
                # AIMessage.content 可能是 str 或 list
                if isinstance(msg.content, str):
                    ai_response = msg.content
                elif isinstance(msg.content, list):
                    # 如果是 list，找到 text 內容
                    for block in msg.content:
                        if isinstance(block, dict) and block.get("type") == "text":
                            ai_response = block.get("text", "")
                            break
                        elif isinstance(block, str):
                            ai_response = block
                            break
                break

        # 4. 更新對話歷史
        # 只保留最終的 AI 回應，避免歷史太長
        self.messages.append(AIMessage(content=ai_response))

        return ai_response

    def clear_history(self):
        """清除對話歷史。"""
        self.messages = []

    def get_history(self) -> list:
        """
        取得對話歷史。

        Returns:
            訊息列表，每個元素是 (role, content) tuple
        """
        history = []
        for msg in self.messages:
            if isinstance(msg, HumanMessage):
                history.append(("user", msg.content))
            elif isinstance(msg, AIMessage):
                history.append(("assistant", msg.content))
        return history

    def _update_prompt_for_intent(self, message: str):
        """
        根據用戶訊息分類意圖，並更新 Agent 的 Prompt。

        Args:
            message: 用戶訊息
        """
        # Classify intent
        self.last_intent = self.intent_classifier.classify(message)
        intent_id = self.last_intent["intent_id"]

        # Compose dynamic prompt
        dynamic_prompt = self.prompt_retriever.compose(intent_id, message)

        # Get prompt stats for logging
        stats = self.prompt_retriever.get_stats(intent_id, message)
        print(f"[RAG] Intent: {intent_id} ({self.last_intent['confidence']:.2f}), "
              f"Fragments: {stats['fragment_count']}, "
              f"Tokens: ~{stats['estimated_tokens']}")

        # Recreate agent with new prompt
        self.agent = create_react_agent(
            model=self.llm,
            tools=all_tools,
            prompt=dynamic_prompt,
        )

    def get_last_intent(self) -> dict:
        """
        取得最後一次的意圖分類結果（用於除錯）。

        Returns:
            dict with intent_id, intent_name, confidence, similar_examples
        """
        return self.last_intent
