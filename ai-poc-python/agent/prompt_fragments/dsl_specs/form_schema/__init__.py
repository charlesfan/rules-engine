"""
Form Schema DSL Specifications.

Contains specifications for form field types:
- overview: General form schema structure
- text_fields: text, email fields
- number_field: number input
- select_field: dropdown select
- checkbox_field: checkbox input
"""

from .overview import FORM_OVERVIEW_SPEC
from .text_fields import TEXT_FIELDS_SPEC
from .number_field import NUMBER_FIELD_SPEC
from .select_field import SELECT_FIELD_SPEC
from .checkbox_field import CHECKBOX_FIELD_SPEC

__all__ = [
    "FORM_OVERVIEW_SPEC",
    "TEXT_FIELDS_SPEC",
    "NUMBER_FIELD_SPEC",
    "SELECT_FIELD_SPEC",
    "CHECKBOX_FIELD_SPEC",
]
