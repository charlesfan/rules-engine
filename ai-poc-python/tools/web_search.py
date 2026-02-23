"""
Web Search Tool for LangChain Agent.

Uses DuckDuckGo Search (free, no API key required).
"""

from langchain_core.tools import tool
from langchain_community.tools import DuckDuckGoSearchResults


# Initialize DuckDuckGo search
ddg_search = DuckDuckGoSearchResults(
    num_results=5,
    output_format="list",
)


@tool
def web_search(query: str) -> str:
    """
    Search the web for public information.
    Use this tool to find reference information about other events or explain unfamiliar terms.

    Args:
        query: Search keywords in Chinese or English.
               Example: "馬拉松報名費", "路跑早鳥優惠"

    Returns:
        Search results with titles, snippets, and links.
    """
    try:
        results = ddg_search.invoke(query)

        if not results:
            return "No results found."

        # Format results for better readability
        formatted = []
        for r in results:
            formatted.append(
                f"- {r.get('title', 'N/A')}\n"
                f"  {r.get('snippet', 'N/A')}\n"
                f"  Link: {r.get('link', 'N/A')}"
            )

        return "\n\n".join(formatted)

    except Exception as e:
        return f"Error searching web: {str(e)}"
