"""
Intent Definitions for RAG Classification.

Each intent has:
- id: Unique identifier
- name: Display name
- description: What this intent does
- examples: Example user messages that trigger this intent
"""

INTENT_DEFINITIONS = [
    {
        "id": "create_event",
        "name": "建立賽事",
        "description": "用戶想要建立新的賽事活動",
        "examples": [
            "我想建立一個新賽事",
            "幫我建立馬拉松活動",
            "新增一個路跑活動",
            "建立2026大湖馬拉松",
            "我要辦一場比賽",
            "建立新的報名活動",
            "幫我建立一個鐵人三項賽事",
            "新增賽事",
            "create new event",
            "我想辦一個健走活動",
        ],
    },
    {
        "id": "update_event",
        "name": "修改賽事",
        "description": "用戶想要修改現有賽事的任何設定（價格、優惠、表單、驗證等）",
        "examples": [
            # 價格相關
            "修改報名費",
            "改一下價格",
            "全馬改成1500元",
            "調整報名費用",
            # 優惠相關
            "加一個早鳥優惠",
            "新增團報折扣",
            "設定9折優惠",
            "加優惠",
            # 表單相關
            "加一個欄位",
            "新增電話欄位",
            "修改報名表單",
            "加上緊急聯絡人",
            # 驗證相關
            "加入年齡限制",
            "限制18歲以上才能報名",
            "設定報名條件",
            # 一般修改
            "修改這個賽事",
            "更新賽事設定",
            "編輯活動內容",
            "改一下設定",
        ],
    },
    {
        "id": "search_event",
        "name": "搜尋賽事",
        "description": "用戶想要搜尋或列出現有賽事",
        "examples": [
            "列出所有賽事",
            "搜尋馬拉松",
            "有哪些賽事",
            "查詢賽事",
            "找一下路跑活動",
            "顯示所有活動",
            "看看有什麼賽事",
            "list events",
            "搜尋",
            "查一下有哪些活動",
        ],
    },
    {
        "id": "get_event",
        "name": "查看賽事詳情",
        "description": "用戶想要查看特定賽事的詳細資訊",
        "examples": [
            "查看馬拉松詳情",
            "顯示這個賽事的內容",
            "看一下大湖馬拉松",
            "這個賽事的規則是什麼",
            "詳細資訊",
            "查看DSL",
            "顯示完整設定",
            "看一下這個賽事",
        ],
    },
    {
        "id": "delete_event",
        "name": "刪除賽事",
        "description": "用戶想要刪除現有賽事",
        "examples": [
            "刪除這個賽事",
            "移除馬拉松活動",
            "刪掉這個活動",
            "取消這個賽事",
            "刪除",
            "移除",
            "不要這個賽事了",
        ],
    },
    {
        "id": "calculate_price",
        "name": "計算價格",
        "description": "用戶想要計算或預覽報名費用",
        "examples": [
            "算一下報名費",
            "計算價格",
            "全馬報名要多少錢",
            "預覽費用",
            "早鳥價多少",
            "團報5人的價格",
            "幫我算價格",
            "這樣報名要多少",
        ],
    },
    {
        "id": "preview_event",
        "name": "預覽報名頁面",
        "description": "用戶想要預覽或測試報名頁面",
        "examples": [
            "預覽報名頁面",
            "測試報名",
            "看一下報名表",
            "預覽表單",
            "測試一下",
            "打開報名頁面",
            "我想看報名畫面",
            "preview",
        ],
    },
    {
        "id": "general",
        "name": "一般對話",
        "description": "一般性問題或不屬於其他類別",
        "examples": [
            "你好",
            "你是誰",
            "幫助",
            "謝謝",
            "可以做什麼",
            "DSL是什麼",
            "怎麼使用",
            "說明一下",
        ],
    },
]

# For easy lookup
INTENT_BY_ID = {intent["id"]: intent for intent in INTENT_DEFINITIONS}

# Keywords for detecting which DSL specs to load for update_event
UPDATE_KEYWORDS = {
    "pricing": [
        "價格", "費用", "報名費", "金額", "元", "塊", "錢",
        "price", "cost", "fee",
    ],
    "discount": [
        "優惠", "折扣", "早鳥", "團報", "折", "減",
        "discount", "promotion", "coupon",
    ],
    "form": [
        "欄位", "表單", "輸入", "填寫", "選項",
        "field", "form", "input",
    ],
    "validation": [
        "驗證", "限制", "條件", "規則", "必填", "上限", "下限",
        "validation", "rule", "limit", "require",
    ],
}
