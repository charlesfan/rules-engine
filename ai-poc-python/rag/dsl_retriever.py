"""
DSL Knowledge Retriever.

Retrieves relevant DSL specifications based on user query.
Uses Gemini Embedding + ChromaDB for semantic search,
then dynamically loads content from source files.
"""

import importlib
from typing import Optional

import chromadb
import google.generativeai as genai

from config.settings import settings
from .dsl_index import DSL_INDEX, get_index_for_embedding


class DSLRetriever:
    """
    DSL knowledge retriever using ChromaDB vector similarity.

    Uses Gemini API for embedding computation on the client side,
    then stores/queries vectors in ChromaDB server.
    """

    COLLECTION_NAME = "dsl_knowledge"

    def __init__(self, use_persistent: bool = True):
        """
        Initialize the retriever.

        Args:
            use_persistent: If True, connect to ChromaDB server.
                          If False, use in-memory storage (for testing).
        """
        # Check for API key
        if not settings.google_api_key:
            raise ValueError("GOOGLE_API_KEY is required for DSL retrieval")

        genai.configure(api_key=settings.google_api_key)
        self.embedding_model = "models/gemini-embedding-001"

        # Connect to ChromaDB
        if use_persistent:
            self.client = chromadb.HttpClient(
                host=settings.chroma_host,
                port=settings.chroma_port,
            )
        else:
            self.client = chromadb.Client()

        # Create collection
        self.collection = self.client.get_or_create_collection(
            name=self.COLLECTION_NAME,
            metadata={"description": "DSL knowledge base for event registration rules"},
        )

        # Initialize if empty
        if self.collection.count() == 0:
            self._initialize_collection()

    def _initialize_collection(self):
        """Populate the collection with DSL index data."""
        index_data = get_index_for_embedding()

        if not index_data:
            print("[DSLRetriever] No index data to initialize")
            return

        ids = [item["id"] for item in index_data]
        texts = [item["text"] for item in index_data]
        metadatas = [item["metadata"] for item in index_data]

        # Compute embeddings using Gemini
        print(f"[DSLRetriever] Computing embeddings for {len(texts)} DSL specs...")
        embeddings = self._embed_batch(texts)

        # Add to ChromaDB
        self.collection.add(
            ids=ids,
            embeddings=embeddings,
            metadatas=metadatas,
            documents=texts,
        )
        print(f"[DSLRetriever] Initialized with {len(texts)} DSL specs")

    def _embed_batch(self, texts: list[str]) -> list[list[float]]:
        """Generate embeddings for a list of texts."""
        result = genai.embed_content(
            model=self.embedding_model,
            content=texts,
            task_type="retrieval_document",
        )
        return result["embedding"]

    def _embed_query(self, text: str) -> list[float]:
        """Generate embedding for a single query."""
        result = genai.embed_content(
            model=self.embedding_model,
            content=text,
            task_type="retrieval_query",
        )
        return result["embedding"]

    def _load_content(self, module_path: str, variable_name: str) -> Optional[str]:
        """
        Dynamically load content from a Python module.

        Args:
            module_path: Full module path, e.g., "agent.prompt_fragments.dsl_specs.pricing_actions.set_price"
            variable_name: Variable name to retrieve, e.g., "SET_PRICE_SPEC"

        Returns:
            The content string, or None if loading fails.
        """
        try:
            module = importlib.import_module(module_path)
            return getattr(module, variable_name, None)
        except Exception as e:
            print(f"[DSLRetriever] Error loading {module_path}.{variable_name}: {e}")
            return None

    def search(self, query: str, top_k: int = 3) -> list[dict]:
        """
        Search for relevant DSL specifications.

        Args:
            query: User's question or description
            top_k: Number of results to return

        Returns:
            List of dicts with title, content, and similarity score
        """
        # Compute query embedding
        query_embedding = self._embed_query(query)

        # Query ChromaDB
        results = self.collection.query(
            query_embeddings=[query_embedding],
            n_results=top_k,
        )

        if not results["metadatas"] or not results["metadatas"][0]:
            return []

        # Process results and load content
        output = []
        distances = results["distances"][0] if results["distances"] else []

        for i, metadata in enumerate(results["metadatas"][0]):
            distance = distances[i] if i < len(distances) else 1.0
            similarity = 1.0 / (1.0 + distance)

            # Load actual content from source file
            content = self._load_content(
                metadata["module"],
                metadata["variable"],
            )

            if content:
                output.append({
                    "title": metadata["title"],
                    "content": content,
                    "similarity": round(similarity, 3),
                })

        return output

    def reset(self):
        """Reset the collection (for updating index)."""
        try:
            self.client.delete_collection(self.COLLECTION_NAME)
        except Exception:
            pass

        self.collection = self.client.create_collection(
            name=self.COLLECTION_NAME,
            metadata={"description": "DSL knowledge base for event registration rules"},
        )
        self._initialize_collection()


# Singleton instance for reuse
_retriever_instance: Optional[DSLRetriever] = None


def get_dsl_retriever() -> DSLRetriever:
    """Get or create the singleton DSLRetriever instance."""
    global _retriever_instance
    if _retriever_instance is None:
        _retriever_instance = DSLRetriever()
    return _retriever_instance
