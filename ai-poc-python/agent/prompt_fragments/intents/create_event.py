"""
Create Event Intent Prompt.

Loaded when user wants to create a new event.
"""

CREATE_EVENT_PROMPT = """## 建立賽事的流程

當用戶想建立新賽事時，請依序詢問：

1. **賽事名稱**：例如「2026大湖馬拉松」
2. **組別與價格**：例如「全馬1050、半馬950」
3. **優惠規則**（可選）：例如「早鳥9折、團報5人以上再折100」
4. **報名欄位**：例如「姓名、性別、生日、緊急聯絡人」

收集完資訊後，生成 DSL 並使用 create_event tool 建立賽事。

## 完整 DSL 範例

以下是一個完整的馬拉松賽事 DSL：

```json
{
  "event_id": "marathon-2026",
  "version": "1.0",
  "name": "2026馬拉松賽事",
  "variables": {
    "full_marathon_price": 2000,
    "half_marathon_price": 1500,
    "mini_marathon_price": 1000
  },
  "pricing_rules": [
    {
      "id": "full_marathon",
      "priority": 0,
      "description": "全馬報名費",
      "condition": {
        "type": "equals",
        "field": "user.race_type",
        "value": "full_marathon"
      },
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 2000,
        "label": "全馬報名費"
      }
    },
    {
      "id": "half_marathon",
      "priority": 0,
      "description": "半馬報名費",
      "condition": {
        "type": "equals",
        "field": "user.race_type",
        "value": "half_marathon"
      },
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1500,
        "label": "半馬報名費"
      }
    },
    {
      "id": "mini_marathon",
      "priority": 0,
      "description": "迷你馬報名費",
      "condition": {
        "type": "equals",
        "field": "user.race_type",
        "value": "mini_marathon"
      },
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1000,
        "label": "迷你馬報名費"
      }
    }
  ],
  "validation_rules": [
    {
      "id": "race_type_required",
      "description": "必須選擇組別",
      "condition": {
        "type": "field_empty",
        "field": "user.race_type"
      },
      "error_type": "blocking",
      "error_message": "請選擇賽事組別"
    }
  ],
  "discount_stacking": {
    "mode": "additive",
    "description": "折扣累加計算"
  },
  "form_schema": {
    "fields": [
      {
        "id": "name",
        "label": "姓名",
        "type": "text",
        "field": "user.name",
        "required": true
      },
      {
        "id": "email",
        "label": "Email",
        "type": "email",
        "field": "user.email",
        "required": true
      },
      {
        "id": "race_type",
        "label": "賽事組別",
        "type": "select",
        "field": "user.race_type",
        "required": true,
        "options": [
          {"label": "請選擇組別", "value": ""},
          {"label": "全馬 - NT$2,000", "value": "full_marathon"},
          {"label": "半馬 - NT$1,500", "value": "half_marathon"},
          {"label": "迷你馬 - NT$1,000", "value": "mini_marathon"}
        ]
      }
    ]
  }
}
```"""
