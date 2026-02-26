"""
Tiered Pricing Pattern.

Complete example for implementing multi-tier pricing for different race categories.
"""

TIERED_PRICING_PATTERN = """## 分級定價模式

完整的多組別定價設定範例，不同組別有不同價格。

### 基本分級定價（全馬、半馬、迷你馬）

```json
{
  "event_id": "marathon-2026",
  "version": "1.0",
  "name": "2026馬拉松賽事",
  "pricing_rules": [
    {
      "id": "full_marathon_price",
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
        "value": 1500,
        "label": "全馬報名費"
      }
    },
    {
      "id": "half_marathon_price",
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
        "value": 1200,
        "label": "半馬報名費"
      }
    },
    {
      "id": "mini_marathon_price",
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
        "value": 800,
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
  "form_schema": {
    "fields": [
      {
        "id": "race_type",
        "label": "賽事組別",
        "type": "select",
        "field": "user.race_type",
        "required": true,
        "options": [
          {"label": "請選擇組別", "value": ""},
          {"label": "全程馬拉松 (42K) - NT$1,500", "value": "full_marathon"},
          {"label": "半程馬拉松 (21K) - NT$1,200", "value": "half_marathon"},
          {"label": "迷你馬拉松 (10K) - NT$800", "value": "mini_marathon"}
        ]
      }
    ]
  }
}
```

### 分級定價 + 統一折扣

不同組別價格，但共用同一個早鳥優惠：

```json
{
  "pricing_rules": [
    {
      "id": "full_marathon_price",
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
        "value": 1500,
        "label": "全馬報名費"
      }
    },
    {
      "id": "half_marathon_price",
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
        "value": 1200,
        "label": "半馬報名費"
      }
    },
    {
      "id": "early_bird",
      "priority": 100,
      "description": "早鳥優惠9折（所有組別適用）",
      "condition": {
        "type": "datetime_before",
        "field": "register_date",
        "value": "2026-03-31T23:59:59+08:00"
      },
      "action": {
        "type": "percentage_discount",
        "target": "registration_fee",
        "value": 10,
        "label": "早鳥優惠"
      }
    }
  ]
}
```

### 使用 variables 簡化

用變數管理價格，方便統一修改：

```json
{
  "variables": {
    "full_price": 1500,
    "half_price": 1200,
    "mini_price": 800
  },
  "pricing_rules": [
    {
      "id": "full_marathon_price",
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
        "value": "$variables.full_price",
        "label": "全馬報名費"
      }
    }
  ]
}
```

### 重要提醒

1. **options 對應**：form_schema 中 race_type 的 options value 必須與 pricing_rules 中的 condition value 一致
2. **必填驗證**：建議在 validation_rules 中加入 race_type 必填驗證
3. **價格顯示**：建議在 select options 的 label 中顯示價格
"""
