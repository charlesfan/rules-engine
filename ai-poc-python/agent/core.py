"""
LangChain Agent Core.

這是 Agent 的核心模組，負責：
1. 建立 LLM（Claude）連線
2. 將 Tools 綁定到 Agent
3. 管理對話歷史
4. 執行 Agent Loop

LangChain Agent 的運作原理：
1. 收到用戶訊息
2. LLM 決定是否需要使用 Tool
3. 如果需要，執行 Tool 並將結果給 LLM
4. LLM 根據 Tool 結果生成回應
5. 重複 2-4 直到 LLM 決定直接回應
"""

from langchain_anthropic import ChatAnthropic
from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
from langgraph.prebuilt import create_react_agent

from config.settings import settings
from agent.prompts import SYSTEM_PROMPT
from tools.events import all_tools


class EventAgent:
    """
    賽事上架助手 Agent。

    使用 LangChain 的 ReAct Agent 架構：
    - ReAct = Reasoning + Acting
    - Agent 會「思考」要做什麼，然後「行動」（呼叫 Tool）
    """

    def __init__(self):
        """
        初始化 Agent。

        建立 LLM 和 Agent 實例。
        """
        # 1. 建立 Claude LLM
        # ChatAnthropic 是 LangChain 對 Claude API 的封裝
        self.llm = ChatAnthropic(
            model=settings.model_name,
            api_key=settings.anthropic_api_key,
            max_tokens=4096,
        )

        # 2. 建立 ReAct Agent
        # create_react_agent 會自動處理：
        # - Tool 呼叫的 loop
        # - 錯誤重試
        # - 對話歷史管理
        self.agent = create_react_agent(
            model=self.llm,
            tools=all_tools,
            prompt=SYSTEM_PROMPT,  # System prompt
        )

        # 3. 對話歷史
        # 保存所有訊息，讓 Agent 有記憶
        self.messages: list = []

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

        # 2. 執行 Agent
        # invoke() 會執行完整的 ReAct loop
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
