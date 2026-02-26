"""
Common Business Patterns for Event Registration.

Contains complete DSL examples for common scenarios:
- early_bird: Early bird discount patterns
- group_discount: Group registration discount patterns
- tiered_pricing: Multi-tier pricing patterns
- age_restriction: Age-based validation patterns
"""

from .early_bird import EARLY_BIRD_PATTERN
from .group_discount import GROUP_DISCOUNT_PATTERN
from .tiered_pricing import TIERED_PRICING_PATTERN
from .age_restriction import AGE_RESTRICTION_PATTERN

__all__ = [
    "EARLY_BIRD_PATTERN",
    "GROUP_DISCOUNT_PATTERN",
    "TIERED_PRICING_PATTERN",
    "AGE_RESTRICTION_PATTERN",
]
