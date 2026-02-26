"""
Events CRUD Tools for LangChain Agent.

LangChain Tool 是讓 Agent 可以執行的動作。
每個 Tool 有：
- name: 工具名稱（Agent 會用這個名稱呼叫）
- description: 描述（Agent 根據這個決定何時使用）
- 執行函數: 實際執行的邏輯

使用 @tool 裝飾器是最簡單的定義方式。
"""

import json
from typing import Any
from langchain_core.tools import tool
from tools.http_client import api_client
from tools.web_search import web_search
from tools.dsl_search import search_dsl_rules


@tool
def search_events(query: str = "") -> str:
    """
    Search for events by name or description.
    Use this tool when user wants to find, list, or search for events.

    Args:
        query: Search keyword (e.g., "馬拉松", "2026大湖").
               Leave empty to list all events.

    Returns:
        JSON string containing list of matching events.
    """
    try:
        # 呼叫 GET /api/events?q=...
        params = {"q": query} if query else None
        result = api_client.get("/api/events", params=params)

        # 格式化輸出，讓 Agent 容易閱讀
        events = result.get("data", [])
        if not events:
            return "No events found."

        # 只回傳摘要資訊，避免 token 太長
        summaries = []
        for event in events:
            summaries.append({
                "id": event["id"],
                "name": event["name"],
                "status": event["status"],
            })

        return json.dumps(summaries, ensure_ascii=False, indent=2)

    except Exception as e:
        return f"Error searching events: {str(e)}"


@tool
def get_event(event_id: str) -> str:
    """
    Get full details of a specific event by ID.
    Use this tool when you need to see the complete DSL rules of an event.

    Args:
        event_id: The UUID of the event.

    Returns:
        JSON string containing the full event with DSL.
    """
    try:
        result = api_client.get(f"/api/events/{event_id}")
        return json.dumps(result, ensure_ascii=False, indent=2)
    except Exception as e:
        return f"Error getting event: {str(e)}"


@tool
def create_event(name: str, description: str, dsl: dict[str, Any]) -> str:
    """
    Create a new event with DSL rules.
    Use this tool when user wants to create a new event/race/competition.

    Args:
        name: Event name (e.g., "2026大湖馬拉松").
        description: Event description.
        dsl: The DSL rules as a dictionary. Must include at least
             event_id, pricing_rules, and validation_rules.

    Returns:
        JSON string containing the created event with ID.
    """
    try:
        data = {
            "name": name,
            "description": description,
            "dsl": dsl,
        }
        result = api_client.post("/api/events", data=data)
        return json.dumps(result, ensure_ascii=False, indent=2)
    except Exception as e:
        return f"Error creating event: {str(e)}"


@tool
def update_event(
    event_id: str,
    name: str | None = None,
    description: str | None = None,
    dsl: dict[str, Any] | None = None,
    status: str | None = None,
) -> str:
    """
    Update an existing event.
    Use this tool when user wants to modify an event's name, description, DSL, or status.

    Args:
        event_id: The UUID of the event to update.
        name: New name (optional).
        description: New description (optional).
        dsl: New DSL rules (optional).
        status: New status like "draft", "published" (optional).

    Returns:
        JSON string containing the updated event.
    """
    try:
        # 只包含要更新的欄位
        data: dict[str, Any] = {}
        if name is not None:
            data["name"] = name
        if description is not None:
            data["description"] = description
        if dsl is not None:
            data["dsl"] = dsl
        if status is not None:
            data["status"] = status

        if not data:
            return "Error: No fields to update"

        result = api_client.put(f"/api/events/{event_id}", data=data)
        return json.dumps(result, ensure_ascii=False, indent=2)
    except Exception as e:
        return f"Error updating event: {str(e)}"


@tool
def delete_event(event_id: str) -> str:
    """
    Delete an event.
    Use this tool when user wants to remove an event.

    Args:
        event_id: The UUID of the event to delete.

    Returns:
        Success or error message.
    """
    try:
        api_client.delete(f"/api/events/{event_id}")
        return f"Event {event_id} deleted successfully."
    except Exception as e:
        return f"Error deleting event: {str(e)}"


@tool
def validate_event(event_id: str) -> str:
    """
    Validate an event's DSL.
    Use this tool to check if the DSL rules are valid.

    Args:
        event_id: The UUID of the event to validate.

    Returns:
        Validation result (valid or error details).
    """
    try:
        result = api_client.post(f"/api/events/{event_id}/validate", data={})
        if result.get("valid"):
            return "DSL is valid."
        else:
            return f"DSL validation failed: {result.get('error', 'Unknown error')}"
    except Exception as e:
        return f"Error validating event: {str(e)}"


@tool
def calculate_price(event_id: str, context: dict[str, Any]) -> str:
    """
    Calculate registration price for an event.
    Use this tool to preview pricing based on user's registration data.

    Args:
        event_id: The UUID of the event.
        context: Registration context including user info, team info, addons, etc.
                 Example: {"user": {"race_type": "half_marathon"}, "team_size": 3}

    Returns:
        Price breakdown and validation results.
    """
    try:
        data = {"context": context}
        result = api_client.post(f"/api/events/{event_id}/calculate", data=data)
        return json.dumps(result, ensure_ascii=False, indent=2)
    except Exception as e:
        return f"Error calculating price: {str(e)}"


# 匯出所有 tools，方便 Agent 使用
all_tools = [
    search_events,
    get_event,
    create_event,
    update_event,
    delete_event,
    validate_event,
    calculate_price,
    web_search,
    search_dsl_rules,  # DSL knowledge retrieval
]
