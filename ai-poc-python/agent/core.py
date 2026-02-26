"""
LangChain Agent Core.

這是 Agent 的核心模組，負責：
1. 建立 LLM（Claude）連線
2. 將 Tools 綁定到 Agent
3. 管理對話歷史
4. 執行 Agent Loop

新架構（v2）：
- LLM 直接理解用戶意圖（不需要 RAG 意圖分類）
- LLM 透過 search_dsl_rules tool 查詢 DSL 規則範例
- 使用精簡的 system prompt
"""

from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage, AIMessage
from langgraph.prebuilt import create_react_agent

from config.settings import settings
from tools.events import all_tools
from agent.prompts import SYSTEM_PROMPT


class EventAgent:
    """
    賽事上架助手 Agent。

    使用 LangChain 的 ReAct Agent 架構：
    - ReAct = Reasoning + Acting
    - Agent 會「思考」要做什麼，然後「行動」（呼叫 Tool）

    v2 架構：
    - LLM 直接理解意圖，不需要 RAG 意圖分類
    - 需要 DSL 範例時，LLM 會呼叫 search_dsl_rules tool
    """

    def __init__(self):
        """初始化 Agent。"""
        self.llm = ChatAnthropic(
            model=settings.model_name,
            api_key=settings.anthropic_api_key,
            max_tokens=4096,
        )

        # Create agent with static prompt
        # DSL knowledge is retrieved via search_dsl_rules tool when needed
        self.agent = create_react_agent(
            model=self.llm,
            tools=all_tools,
            prompt=SYSTEM_PROMPT,
        )

        # Conversation history
        self.messages: list = []

        print("[Agent] Initialized with static prompt + DSL search tool")

    def chat(self, user_message: str) -> str:
        """
        與 Agent 對話。

        Args:
            user_message: 用戶輸入的訊息

        Returns:
            Agent 的回應文字
        """
        # 1. 加入用戶訊息到歷史
        self.messages.append(HumanMessage(content=user_message))

        # 2. 執行 Agent (ReAct loop)
        result = self.agent.invoke({
            "messages": self.messages,
        })

        # 3. 取得 Agent 回應
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
