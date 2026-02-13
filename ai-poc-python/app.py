"""
Streamlit Chat UI for Event Agent.

Streamlit 是 Python 的 Web UI 框架，特點：
- 無需前端知識（HTML/CSS/JS）
- 用 Python 寫 UI
- 自動處理狀態管理
- 適合快速做 prototype

執行方式：
    cd ai-poc-python
    streamlit run app.py
"""

import streamlit as st
from agent.core import EventAgent


# ============================================================
# Streamlit 基本概念
# ============================================================
#
# 1. st.session_state - 跨 rerun 的狀態保存
#    Streamlit 每次互動都會重新執行整個 script
#    用 session_state 保存需要持久化的資料
#
# 2. st.chat_message - 聊天氣泡 UI
#    自動處理用戶/AI 的不同樣式
#
# 3. st.chat_input - 聊天輸入框
#    自動處理 Enter 送出
#
# ============================================================


def init_session_state():
    """
    初始化 session state。

    session_state 是 Streamlit 的全域狀態管理，
    用來保存跨 rerun 需要保留的資料。
    """
    # 對話歷史：[{"role": "user/assistant", "content": "..."}]
    if "messages" not in st.session_state:
        st.session_state.messages = []

    # Agent 實例
    if "agent" not in st.session_state:
        st.session_state.agent = None

    # 是否已初始化
    if "initialized" not in st.session_state:
        st.session_state.initialized = False


def init_agent():
    """
    初始化 Agent。

    分開初始化是因為 Agent 建立可能需要一些時間，
    且需要環境變數設定正確。
    """
    if not st.session_state.initialized:
        try:
            st.session_state.agent = EventAgent()
            st.session_state.initialized = True
        except Exception as e:
            st.error(f"Agent 初始化失敗：{str(e)}")
            st.error("請確認 .env 檔案中的 ANTHROPIC_API_KEY 設定正確")
            st.stop()  # 停止執行


def display_chat_history():
    """
    顯示聊天歷史。

    st.chat_message 會根據 role 自動選擇樣式：
    - "user": 用戶頭像，靠右
    - "assistant": AI 頭像，靠左
    """
    for message in st.session_state.messages:
        with st.chat_message(message["role"]):
            st.markdown(message["content"])


def handle_user_input(user_input: str):
    """
    處理用戶輸入。

    Args:
        user_input: 用戶輸入的文字
    """
    # 1. 顯示用戶訊息
    with st.chat_message("user"):
        st.markdown(user_input)

    # 2. 加入對話歷史
    st.session_state.messages.append({
        "role": "user",
        "content": user_input,
    })

    # 3. 取得 Agent 回應
    with st.chat_message("assistant"):
        # st.spinner 顯示載入動畫
        with st.spinner("思考中..."):
            response = st.session_state.agent.chat(user_input)

        # 顯示回應
        st.markdown(response)

    # 4. 加入對話歷史
    st.session_state.messages.append({
        "role": "assistant",
        "content": response,
    })


def main():
    """主程式。"""
    # ========== 頁面設定 ==========
    st.set_page_config(
        page_title="賽事上架助手",
        page_icon="🏃",
        layout="wide",
    )

    # ========== 標題 ==========
    st.title("🏃 賽事上架助手")
    st.caption("透過對話建立和管理賽事報名規則")

    # ========== 側邊欄 ==========
    with st.sidebar:
        st.header("設定")

        # 清除對話按鈕
        if st.button("🗑️ 清除對話", use_container_width=True):
            st.session_state.messages = []
            if st.session_state.agent:
                st.session_state.agent.clear_history()
            st.rerun()  # 重新執行整個 script

        st.divider()  # 分隔線

        # 使用說明
        st.header("使用說明")
        st.markdown("""
        **你可以這樣問：**

        - 列出所有賽事
        - 搜尋「馬拉松」相關賽事
        - 我想建立一個路跑賽事
        - 修改「2026大湖路跑」的報名費
        - 刪除某個賽事

        **建立賽事時，我會詢問：**
        1. 賽事名稱
        2. 組別與價格
        3. 優惠規則
        4. 報名欄位
        """)

    # ========== 初始化 ==========
    init_session_state()
    init_agent()

    # ========== 聊天介面 ==========
    # 顯示歷史訊息
    display_chat_history()

    # 輸入框
    # st.chat_input 會在頁面底部顯示固定的輸入框
    if user_input := st.chat_input("輸入訊息..."):
        handle_user_input(user_input)


# ============================================================
# Python 的 main pattern
# ============================================================
# if __name__ == "__main__":
#     這段程式碼只會在直接執行這個檔案時執行
#     如果是被 import，則不會執行
#
# 但 Streamlit 有自己的執行方式，所以直接呼叫 main()
# ============================================================

if __name__ == "__main__":
    main()
