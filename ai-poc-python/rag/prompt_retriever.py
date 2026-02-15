"""
Prompt Retriever for Dynamic Prompt Composition.

Composes prompts based on detected intent and message keywords.
"""

from agent.prompt_fragments.base import BASE_PROMPT
from agent.prompt_fragments.intents.create_event import CREATE_EVENT_PROMPT
from agent.prompt_fragments.intents.update_event import UPDATE_EVENT_PROMPT
from agent.prompt_fragments.intents.search_event import SEARCH_EVENT_PROMPT
from agent.prompt_fragments.intents.delete_event import DELETE_EVENT_PROMPT
from agent.prompt_fragments.intents.calculate_price import CALCULATE_PRICE_PROMPT
from agent.prompt_fragments.intents.preview_event import PREVIEW_EVENT_PROMPT
from agent.prompt_fragments.dsl_specs.overview import DSL_OVERVIEW_PROMPT
from agent.prompt_fragments.dsl_specs.pricing_rules import PRICING_RULES_PROMPT
from agent.prompt_fragments.dsl_specs.validation_rules import VALIDATION_RULES_PROMPT
from agent.prompt_fragments.dsl_specs.form_schema import FORM_SCHEMA_PROMPT
from agent.prompt_fragments.dsl_specs.conditions import CONDITIONS_PROMPT

from .intent_documents import UPDATE_KEYWORDS


class PromptRetriever:
    """
    Composes dynamic prompts based on intent and keywords.
    """

    # Intent to prompt fragments mapping
    INTENT_PROMPTS = {
        "create_event": [CREATE_EVENT_PROMPT],
        "update_event": [UPDATE_EVENT_PROMPT],
        "search_event": [SEARCH_EVENT_PROMPT],
        "get_event": [SEARCH_EVENT_PROMPT],  # Reuse search prompt
        "delete_event": [DELETE_EVENT_PROMPT],
        "calculate_price": [CALCULATE_PRICE_PROMPT],
        "preview_event": [PREVIEW_EVENT_PROMPT],
        "general": [],
    }

    # All DSL specs for create_event
    ALL_DSL_SPECS = [
        DSL_OVERVIEW_PROMPT,
        PRICING_RULES_PROMPT,
        VALIDATION_RULES_PROMPT,
        FORM_SCHEMA_PROMPT,
        CONDITIONS_PROMPT,
    ]

    def compose(self, intent_id: str, message: str = "") -> str:
        """
        Compose a prompt based on intent and message content.

        Args:
            intent_id: The detected intent ID
            message: The original user message (for keyword detection)

        Returns:
            Combined prompt string
        """
        fragments = [BASE_PROMPT]

        # Add intent-specific prompt
        intent_prompts = self.INTENT_PROMPTS.get(intent_id, [])
        fragments.extend(intent_prompts)

        # Add DSL specs based on intent
        dsl_specs = self._get_dsl_specs(intent_id, message)
        fragments.extend(dsl_specs)

        return "\n\n---\n\n".join(fragments)

    def _get_dsl_specs(self, intent_id: str, message: str) -> list:
        """
        Determine which DSL specs to include based on intent and keywords.
        """
        # create_event needs all specs
        if intent_id == "create_event":
            return self.ALL_DSL_SPECS

        # update_event: detect keywords to load relevant specs
        if intent_id == "update_event":
            return self._detect_update_specs(message)

        # Other intents don't need DSL specs
        return []

    def _detect_update_specs(self, message: str) -> list:
        """
        Detect which DSL specs are needed for update_event based on keywords.
        """
        message_lower = message.lower()
        specs = set()
        matched_categories = []

        # Check each keyword category
        for category, keywords in UPDATE_KEYWORDS.items():
            for keyword in keywords:
                if keyword in message_lower:
                    matched_categories.append(category)
                    break

        # Map categories to specs
        if "pricing" in matched_categories or "discount" in matched_categories:
            specs.add(PRICING_RULES_PROMPT)
            specs.add(CONDITIONS_PROMPT)

        if "form" in matched_categories:
            specs.add(FORM_SCHEMA_PROMPT)

        if "validation" in matched_categories:
            specs.add(VALIDATION_RULES_PROMPT)
            specs.add(CONDITIONS_PROMPT)

        # If no specific keywords detected, include overview only
        if not specs:
            specs.add(DSL_OVERVIEW_PROMPT)

        return list(specs)

    def get_stats(self, intent_id: str, message: str = "") -> dict:
        """
        Get statistics about the composed prompt.

        Returns:
            dict with fragment_count and estimated_tokens
        """
        prompt = self.compose(intent_id, message)

        # Rough token estimate (1 token ≈ 4 characters for Chinese)
        estimated_tokens = len(prompt) // 2

        fragments = [BASE_PROMPT]
        fragments.extend(self.INTENT_PROMPTS.get(intent_id, []))
        fragments.extend(self._get_dsl_specs(intent_id, message))

        return {
            "intent_id": intent_id,
            "fragment_count": len(fragments),
            "char_count": len(prompt),
            "estimated_tokens": estimated_tokens,
        }
