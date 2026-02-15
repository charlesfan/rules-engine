"""
Prompt Fragments for Dynamic Prompt Composition.

This module contains all prompt fragments organized by category:
- base: Core persona and dialogue style (always loaded)
- intents: Intent-specific guidance
- dsl_specs: DSL specification documents
"""

from .base import BASE_PROMPT

# Intent prompts
from .intents.create_event import CREATE_EVENT_PROMPT
from .intents.update_event import UPDATE_EVENT_PROMPT
from .intents.search_event import SEARCH_EVENT_PROMPT
from .intents.delete_event import DELETE_EVENT_PROMPT
from .intents.calculate_price import CALCULATE_PRICE_PROMPT
from .intents.preview_event import PREVIEW_EVENT_PROMPT

# DSL spec prompts
from .dsl_specs.overview import DSL_OVERVIEW_PROMPT
from .dsl_specs.pricing_rules import PRICING_RULES_PROMPT
from .dsl_specs.validation_rules import VALIDATION_RULES_PROMPT
from .dsl_specs.form_schema import FORM_SCHEMA_PROMPT
from .dsl_specs.conditions import CONDITIONS_PROMPT

# Intent to fragments mapping
INTENT_FRAGMENTS = {
    "create_event": [
        CREATE_EVENT_PROMPT,
        DSL_OVERVIEW_PROMPT,
        PRICING_RULES_PROMPT,
        VALIDATION_RULES_PROMPT,
        FORM_SCHEMA_PROMPT,
        CONDITIONS_PROMPT,
    ],
    "update_pricing": [
        UPDATE_EVENT_PROMPT,
        PRICING_RULES_PROMPT,
        CONDITIONS_PROMPT,
    ],
    "add_discount": [
        UPDATE_EVENT_PROMPT,
        PRICING_RULES_PROMPT,
        CONDITIONS_PROMPT,
    ],
    "update_form": [
        UPDATE_EVENT_PROMPT,
        FORM_SCHEMA_PROMPT,
    ],
    "add_validation": [
        UPDATE_EVENT_PROMPT,
        VALIDATION_RULES_PROMPT,
        CONDITIONS_PROMPT,
    ],
    "search_event": [
        SEARCH_EVENT_PROMPT,
    ],
    "delete_event": [
        DELETE_EVENT_PROMPT,
    ],
    "calculate_price": [
        CALCULATE_PRICE_PROMPT,
    ],
    "preview_event": [
        PREVIEW_EVENT_PROMPT,
    ],
    "general": [],  # Only base prompt
}


def compose_prompt(intent: str) -> str:
    """
    Compose a full prompt based on detected intent.

    Args:
        intent: The detected user intent

    Returns:
        Combined prompt string
    """
    fragments = [BASE_PROMPT]
    fragments.extend(INTENT_FRAGMENTS.get(intent, []))
    return "\n\n---\n\n".join(fragments)
