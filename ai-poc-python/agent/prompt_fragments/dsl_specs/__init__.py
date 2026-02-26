"""
DSL Specification Prompt Fragments.

Organized structure for DSL knowledge base:
- overview: DSL structure and required fields
- pricing_actions/: set_price, add_item, percentage_discount, fixed_discount
- conditions/: equals, compare, datetime, logic, field_check
- validation_rules: validation rule specification
- form_schema/: form field types
- patterns/: common business patterns (early_bird, group_discount, etc.)
"""

# Overview
from .overview import DSL_OVERVIEW_PROMPT

# Validation Rules (not split further)
from .validation_rules import VALIDATION_RULES_PROMPT

# Pricing Actions
from .pricing_actions import (
    SET_PRICE_SPEC,
    ADD_ITEM_SPEC,
    PERCENTAGE_DISCOUNT_SPEC,
    FIXED_DISCOUNT_SPEC,
)

# Conditions
from .conditions import (
    EQUALS_SPEC,
    COMPARE_SPEC,
    DATETIME_SPEC,
    LOGIC_SPEC,
    FIELD_CHECK_SPEC,
)

# Form Schema
from .form_schema import (
    FORM_OVERVIEW_SPEC,
    TEXT_FIELDS_SPEC,
    NUMBER_FIELD_SPEC,
    SELECT_FIELD_SPEC,
    CHECKBOX_FIELD_SPEC,
)

# Patterns
from .patterns import (
    EARLY_BIRD_PATTERN,
    GROUP_DISCOUNT_PATTERN,
    TIERED_PRICING_PATTERN,
    AGE_RESTRICTION_PATTERN,
)

__all__ = [
    # Overview
    "DSL_OVERVIEW_PROMPT",
    "VALIDATION_RULES_PROMPT",
    # Pricing Actions
    "SET_PRICE_SPEC",
    "ADD_ITEM_SPEC",
    "PERCENTAGE_DISCOUNT_SPEC",
    "FIXED_DISCOUNT_SPEC",
    # Conditions
    "EQUALS_SPEC",
    "COMPARE_SPEC",
    "DATETIME_SPEC",
    "LOGIC_SPEC",
    "FIELD_CHECK_SPEC",
    # Form Schema
    "FORM_OVERVIEW_SPEC",
    "TEXT_FIELDS_SPEC",
    "NUMBER_FIELD_SPEC",
    "SELECT_FIELD_SPEC",
    "CHECKBOX_FIELD_SPEC",
    # Patterns
    "EARLY_BIRD_PATTERN",
    "GROUP_DISCOUNT_PATTERN",
    "TIERED_PRICING_PATTERN",
    "AGE_RESTRICTION_PATTERN",
]
