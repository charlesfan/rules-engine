"""
Condition Types DSL Specifications.

Contains specifications for all condition types:
- equals: Equality check
- compare: Numeric comparison
- datetime: Date/time comparison
- logic: Logical operators (and, or, always_true)
- field_check: Field existence/empty checks
"""

from .equals import EQUALS_SPEC
from .compare import COMPARE_SPEC
from .datetime import DATETIME_SPEC
from .logic import LOGIC_SPEC
from .field_check import FIELD_CHECK_SPEC

__all__ = [
    "EQUALS_SPEC",
    "COMPARE_SPEC",
    "DATETIME_SPEC",
    "LOGIC_SPEC",
    "FIELD_CHECK_SPEC",
]
