"""
Configuration management using Pydantic.

Pydantic 會自動從環境變數讀取設定，支援型別轉換和驗證。
類似 Go 的 struct + envconfig。

使用方式：
    from config.settings import settings
    print(settings.go_api_url)
"""

from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    """
    應用程式設定。

    Pydantic 會自動：
    1. 從環境變數讀取（忽略大小寫）
    2. 從 .env 檔案讀取
    3. 使用預設值（如果沒有設定）
    4. 進行型別轉換（例如 str → int）
    """

    # Go API Server URL
    go_api_url: str = "http://localhost:8080"

    # Anthropic API Key
    anthropic_api_key: str = ""

    # Model name
    model_name: str = "claude-sonnet-4-20250514"

    # ChromaDB settings (Phase 3)
    chroma_host: str = "localhost"
    chroma_port: int = 8000

    class Config:
        """Pydantic 設定"""

        # 從 .env 檔案讀取環境變數
        env_file = ".env"

        # 環境變數名稱忽略大小寫
        # GO_API_URL 和 go_api_url 都可以
        case_sensitive = False


@lru_cache()
def get_settings() -> Settings:
    """
    取得設定（有快取）。

    @lru_cache() 確保只會建立一次 Settings 實例，
    後續呼叫會回傳同一個實例（Singleton pattern）。

    Returns:
        Settings 實例
    """
    return Settings()


# 方便直接 import 使用
# from config.settings import settings
settings = get_settings()
