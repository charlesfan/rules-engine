"""
Early Bird Discount Pattern.

Complete example for implementing early bird discounts.
"""

EARLY_BIRD_PATTERN = """## 早鳥優惠模式

完整的早鳥優惠設定範例，包含時間條件和百分比折扣。

### 基本早鳥優惠（單一時段）

3月底前報名享9折優惠：

```json
{
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "基本報名費",
      "condition": {"type": "always_true"},
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1500,
        "label": "報名費"
      }
    },
    {
      "id": "early_bird",
      "priority": 100,
      "description": "早鳥優惠9折",
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

### 階梯式早鳥優惠（多時段）

超早鳥8折 → 早鳥9折 → 正常價：

```json
{
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "基本報名費",
      "condition": {"type": "always_true"},
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1500,
        "label": "報名費"
      }
    },
    {
      "id": "super_early_bird",
      "priority": 99,
      "description": "超早鳥優惠8折（2月底前）",
      "condition": {
        "type": "datetime_before",
        "field": "register_date",
        "value": "2026-02-28T23:59:59+08:00"
      },
      "action": {
        "type": "percentage_discount",
        "target": "registration_fee",
        "value": 20,
        "label": "超早鳥優惠"
      }
    },
    {
      "id": "early_bird",
      "priority": 100,
      "description": "早鳥優惠9折（3月底前）",
      "condition": {
        "type": "and",
        "conditions": [
          {"type": "datetime_after", "field": "register_date", "value": "2026-02-28T23:59:59+08:00"},
          {"type": "datetime_before", "field": "register_date", "value": "2026-03-31T23:59:59+08:00"}
        ]
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

### 早鳥 + 會員雙重優惠

早鳥9折，會員再享95折（可疊加）：

```json
{
  "pricing_rules": [
    {
      "id": "base_price",
      "priority": 0,
      "description": "基本報名費",
      "condition": {"type": "always_true"},
      "action": {
        "type": "set_price",
        "item": "registration_fee",
        "value": 1500,
        "label": "報名費"
      }
    },
    {
      "id": "early_bird",
      "priority": 100,
      "description": "早鳥優惠9折",
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
    },
    {
      "id": "member_discount",
      "priority": 101,
      "description": "會員優惠95折",
      "condition": {
        "type": "equals",
        "field": "user.is_member",
        "value": true
      },
      "action": {
        "type": "percentage_discount",
        "target": "registration_fee",
        "value": 5,
        "label": "會員優惠"
      }
    }
  ],
  "discount_stacking": {
    "mode": "additive",
    "description": "折扣累加計算"
  }
}
```

### 重要提醒

1. **時間格式**：使用 ISO 8601 格式，包含時區 `+08:00`
2. **Priority 順序**：較大折扣的 priority 數字較小（先判斷）
3. **折扣疊加**：設定 `discount_stacking.mode` 控制疊加方式
"""
