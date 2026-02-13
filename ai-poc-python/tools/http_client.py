"""
HTTP Client for calling Go API.

使用 httpx 作為 HTTP client，它比 requests 更現代化：
- 支援 async/await
- 更好的型別提示
- 自動處理 JSON
"""

import httpx
from typing import Any
from config.settings import settings


class APIClient:
    """
    Go API 的 HTTP Client。

    封裝所有 HTTP 呼叫，提供統一的錯誤處理。
    """

    def __init__(self, base_url: str | None = None):
        """
        初始化 client。

        Args:
            base_url: API 基礎 URL，預設從設定讀取
        """
        self.base_url = base_url or settings.go_api_url
        # httpx.Client 支援連線池，效能較好
        self.client = httpx.Client(
            base_url=self.base_url,
            timeout=30.0,  # 30 秒 timeout
            headers={"Content-Type": "application/json"},
        )

    def get(self, path: str, params: dict | None = None) -> dict[str, Any]:
        """
        發送 GET 請求。

        Args:
            path: API 路徑，例如 "/api/events"
            params: Query parameters，例如 {"q": "馬拉松"}

        Returns:
            API 回應的 JSON（轉成 dict）

        Raises:
            httpx.HTTPStatusError: 如果回應狀態碼 >= 400
        """
        response = self.client.get(path, params=params)
        response.raise_for_status()  # 4xx/5xx 會拋出例外
        return response.json()

    def post(self, path: str, data: dict[str, Any]) -> dict[str, Any]:
        """
        發送 POST 請求。

        Args:
            path: API 路徑
            data: 要發送的 JSON 資料

        Returns:
            API 回應的 JSON
        """
        response = self.client.post(path, json=data)
        response.raise_for_status()
        return response.json()

    def put(self, path: str, data: dict[str, Any]) -> dict[str, Any]:
        """
        發送 PUT 請求。

        Args:
            path: API 路徑
            data: 要發送的 JSON 資料

        Returns:
            API 回應的 JSON
        """
        response = self.client.put(path, json=data)
        response.raise_for_status()
        return response.json()

    def delete(self, path: str) -> None:
        """
        發送 DELETE 請求。

        Args:
            path: API 路徑
        """
        response = self.client.delete(path)
        response.raise_for_status()

    def close(self):
        """關閉 client，釋放連線池資源。"""
        self.client.close()

    def __enter__(self):
        """支援 with 語法。"""
        return self

    def __exit__(self, *args):
        """離開 with 區塊時自動關閉。"""
        self.close()


# 全域 client 實例（Singleton）
api_client = APIClient()
