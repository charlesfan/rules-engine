"""
DSL Knowledge Index.

Contains index data for all DSL specifications.
Each entry points to a source file and variable name for content retrieval.
"""

DSL_INDEX = [
    # ========== Pricing Actions ==========
    {
        "id": "pricing_set_price",
        "title": "set_price - 設定基本價格",
        "description": "設定報名費、組別價格。用於定義不同組別的基本定價，如全馬1500、半馬1200。",
        "keywords": ["報名費", "定價", "價格", "組別", "費用", "全馬", "半馬", "set_price"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.pricing_actions.set_price",
            "variable": "SET_PRICE_SPEC",
        },
    },
    {
        "id": "pricing_add_item",
        "title": "add_item - 新增項目",
        "description": "新增額外項目費用，如運費、保險、加購商品、紀念品。",
        "keywords": ["運費", "保險", "加購", "商品", "宅配", "T-shirt", "紀念品", "add_item"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.pricing_actions.add_item",
            "variable": "ADD_ITEM_SPEC",
        },
    },
    {
        "id": "pricing_percentage_discount",
        "title": "percentage_discount - 百分比折扣",
        "description": "百分比折扣優惠，如早鳥9折、會員85折。value是折扣百分比（10=九折）。",
        "keywords": ["早鳥", "折扣", "打折", "九折", "八折", "優惠", "百分比", "會員", "percentage_discount"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.pricing_actions.percentage_discount",
            "variable": "PERCENTAGE_DISCOUNT_SPEC",
        },
    },
    {
        "id": "pricing_fixed_discount",
        "title": "fixed_discount - 固定金額折扣",
        "description": "固定金額減免，如團報每人減100元、滿額折扣。",
        "keywords": ["團報", "減免", "折價", "每人減", "團體", "固定", "金額", "fixed_discount"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.pricing_actions.fixed_discount",
            "variable": "FIXED_DISCOUNT_SPEC",
        },
    },

    # ========== Conditions ==========
    {
        "id": "condition_equals",
        "title": "equals - 等於比較",
        "description": "精確值比對條件，用於組別選擇、布林值判斷，如race_type等於full_marathon。",
        "keywords": ["等於", "比對", "組別", "選擇", "equals"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.conditions.equals",
            "variable": "EQUALS_SPEC",
        },
    },
    {
        "id": "condition_compare",
        "title": "compare - 數值比較",
        "description": "數值大小比較條件，支援 >, <, >=, <=, ==, != 運算子。用於團報人數、年齡限制。",
        "keywords": ["大於", "小於", "人數", "年齡", "數值", "比較", "compare"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.conditions.compare",
            "variable": "COMPARE_SPEC",
        },
    },
    {
        "id": "condition_datetime",
        "title": "datetime - 時間比較",
        "description": "時間判斷條件，用於早鳥優惠截止時間、報名期限。支援datetime_before和datetime_after。",
        "keywords": ["時間", "日期", "之前", "之後", "早鳥", "截止", "期限", "datetime"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.conditions.datetime",
            "variable": "DATETIME_SPEC",
        },
    },
    {
        "id": "condition_logic",
        "title": "and/or/always_true - 邏輯條件",
        "description": "邏輯組合條件。and表示所有條件都要滿足，or表示任一條件滿足，always_true永遠為真。",
        "keywords": ["且", "或", "同時", "組合", "多條件", "and", "or", "always_true"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.conditions.logic",
            "variable": "LOGIC_SPEC",
        },
    },
    {
        "id": "condition_field_check",
        "title": "field_empty/field_exists - 欄位檢查",
        "description": "檢查欄位是否存在或為空，常用於驗證規則的必填欄位檢查。",
        "keywords": ["必填", "欄位", "空", "存在", "驗證", "field_empty", "field_exists"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.conditions.field_check",
            "variable": "FIELD_CHECK_SPEC",
        },
    },

    # ========== Form Schema ==========
    {
        "id": "form_overview",
        "title": "form_schema - 表單結構概述",
        "description": "表單欄位定義的基本結構和通用屬性說明，包含field路徑對應。",
        "keywords": ["表單", "欄位", "form_schema", "結構", "屬性"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.form_schema.overview",
            "variable": "FORM_OVERVIEW_SPEC",
        },
    },
    {
        "id": "form_text",
        "title": "text/email - 文字輸入欄位",
        "description": "文字輸入和Email輸入欄位，用於姓名、聯絡電話、緊急聯絡人等。",
        "keywords": ["文字", "姓名", "電話", "email", "聯絡", "輸入", "text"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.form_schema.text_fields",
            "variable": "TEXT_FIELDS_SPEC",
        },
    },
    {
        "id": "form_number",
        "title": "number - 數字輸入欄位",
        "description": "數字輸入欄位，支援min/max範圍設定。用於年齡、團報人數、數量等。",
        "keywords": ["數字", "年齡", "人數", "數量", "number", "min", "max"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.form_schema.number_field",
            "variable": "NUMBER_FIELD_SPEC",
        },
    },
    {
        "id": "form_select",
        "title": "select - 下拉選單欄位",
        "description": "下拉選單欄位，用於賽事組別、性別、T-shirt尺寸等預定義選項的選擇。",
        "keywords": ["下拉", "選單", "組別", "選擇", "select", "options", "性別", "尺寸"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.form_schema.select_field",
            "variable": "SELECT_FIELD_SPEC",
        },
    },
    {
        "id": "form_checkbox",
        "title": "checkbox - 核取方塊欄位",
        "description": "核取方塊欄位，用於是/否選擇，如同意條款、加購選項、會員身份。",
        "keywords": ["核取", "勾選", "同意", "加購", "是否", "checkbox", "條款"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.form_schema.checkbox_field",
            "variable": "CHECKBOX_FIELD_SPEC",
        },
    },

    # ========== Validation Rules ==========
    {
        "id": "validation_rules",
        "title": "validation_rules - 驗證規則",
        "description": "驗證規則用於檢查報名資料是否有效，支援blocking（阻止報名）和warning（警告）兩種類型。",
        "keywords": ["驗證", "必填", "錯誤", "blocking", "warning", "error_message"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.validation_rules",
            "variable": "VALIDATION_RULES_PROMPT",
        },
    },

    # ========== Overview ==========
    {
        "id": "dsl_overview",
        "title": "DSL 完整規格概述",
        "description": "DSL JSON格式的完整結構說明，包含必要欄位、選填欄位、discount_stacking模式。",
        "keywords": ["DSL", "結構", "規格", "event_id", "version", "必要", "選填"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.overview",
            "variable": "DSL_OVERVIEW_PROMPT",
        },
    },

    # ========== Patterns ==========
    {
        "id": "pattern_early_bird",
        "title": "早鳥優惠模式",
        "description": "完整的早鳥優惠設定範例，包含單一時段、階梯式早鳥、早鳥+會員雙重優惠。",
        "keywords": ["早鳥", "優惠", "時間", "限時", "折扣", "階梯", "pattern"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.patterns.early_bird",
            "variable": "EARLY_BIRD_PATTERN",
        },
    },
    {
        "id": "pattern_group_discount",
        "title": "團報優惠模式",
        "description": "完整的團報優惠設定範例，包含基本團報、階梯式團報、團報+早鳥雙重優惠。",
        "keywords": ["團報", "團體", "人數", "團隊", "優惠", "階梯", "pattern"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.patterns.group_discount",
            "variable": "GROUP_DISCOUNT_PATTERN",
        },
    },
    {
        "id": "pattern_tiered_pricing",
        "title": "分級定價模式",
        "description": "完整的多組別定價設定範例，不同組別有不同價格，如全馬、半馬、迷你馬。",
        "keywords": ["分級", "定價", "組別", "全馬", "半馬", "不同價格", "pattern"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.patterns.tiered_pricing",
            "variable": "TIERED_PRICING_PATTERN",
        },
    },
    {
        "id": "pattern_age_restriction",
        "title": "年齡限制模式",
        "description": "完整的年齡驗證設定範例，包含年齡下限、年齡範圍、不同組別不同限制、敬老優惠。",
        "keywords": ["年齡", "限制", "歲", "長青", "敬老", "驗證", "pattern"],
        "source": {
            "module": "agent.prompt_fragments.dsl_specs.patterns.age_restriction",
            "variable": "AGE_RESTRICTION_PATTERN",
        },
    },
]


def get_index_for_embedding() -> list[dict]:
    """
    Get index data formatted for embedding.
    Combines title, description, and keywords into searchable text.
    """
    result = []
    for item in DSL_INDEX:
        # Combine title, description, and keywords for better embedding
        searchable_text = f"{item['title']}\n{item['description']}\n{' '.join(item['keywords'])}"
        result.append({
            "id": item["id"],
            "text": searchable_text,
            "metadata": {
                "title": item["title"],
                "module": item["source"]["module"],
                "variable": item["source"]["variable"],
            },
        })
    return result
