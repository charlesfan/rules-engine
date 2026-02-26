"""
Pricing Actions DSL Specifications.

Contains specifications for all pricing action types:
- set_price: Set base price
- add_item: Add additional items (shipping, insurance)
- percentage_discount: Percentage-based discounts
- fixed_discount: Fixed amount discounts
"""

from .set_price import SET_PRICE_SPEC
from .add_item import ADD_ITEM_SPEC
from .percentage_discount import PERCENTAGE_DISCOUNT_SPEC
from .fixed_discount import FIXED_DISCOUNT_SPEC

__all__ = [
    "SET_PRICE_SPEC",
    "ADD_ITEM_SPEC",
    "PERCENTAGE_DISCOUNT_SPEC",
    "FIXED_DISCOUNT_SPEC",
]
