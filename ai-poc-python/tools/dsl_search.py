"""
DSL Search Tool for LangChain Agent.

Allows the agent to search for relevant DSL rule examples
when generating or explaining DSL configurations.
"""

from langchain_core.tools import tool

from rag.dsl_retriever import get_dsl_retriever


@tool
def search_dsl_rules(query: str) -> str:
    """
    Search for DSL rule examples and specifications.
    Use this tool when you need to look up DSL syntax, examples, or patterns.

    When to use:
    - Creating new events (need pricing rules, form schema examples)
    - Setting up discounts (need percentage_discount, fixed_discount examples)
    - Adding validation rules (need condition examples)
    - Explaining DSL features to users

    Args:
        query: Describe what you need in natural language.
               Examples:
               - "早鳥優惠設定"
               - "團報折扣怎麼寫"
               - "表單下拉選單欄位"
               - "年齡限制驗證"

    Returns:
        Relevant DSL specifications with examples.
    """
    try:
        retriever = get_dsl_retriever()
        results = retriever.search(query, top_k=3)

        if not results:
            return "No relevant DSL specifications found."

        # Format results for LLM
        formatted_parts = []
        for r in results:
            formatted_parts.append(
                f"## {r['title']}\n"
                f"(相關度: {r['similarity']})\n\n"
                f"{r['content']}"
            )

        return "\n\n---\n\n".join(formatted_parts)

    except Exception as e:
        return f"Error searching DSL rules: {str(e)}"
