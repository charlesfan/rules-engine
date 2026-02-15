"""
Intent Classifier using ChromaDB + Gemini Embedding.

Uses semantic similarity to classify user messages into intents.
Embeddings are computed on the client side using Gemini API.
"""

import chromadb
import google.generativeai as genai

from config.settings import settings
from .intent_documents import INTENT_DEFINITIONS, INTENT_BY_ID


class GeminiEmbedder:
    """Compute embeddings using Google Gemini API."""

    MODEL_NAME = "models/gemini-embedding-001"

    def __init__(self):
        if not settings.google_api_key:
            raise ValueError("GOOGLE_API_KEY is required")
        genai.configure(api_key=settings.google_api_key)

    def embed(self, texts: list[str]) -> list[list[float]]:
        """Generate embeddings for a list of texts using batch API."""
        # Use batch embedding to reduce API calls
        result = genai.embed_content(
            model=self.MODEL_NAME,
            content=texts,  # Pass list directly for batch processing
            task_type="retrieval_document",
        )
        return result["embedding"]

    def embed_query(self, text: str) -> list[float]:
        """Generate embedding for a single query text."""
        result = genai.embed_content(
            model=self.MODEL_NAME,
            content=text,
            task_type="retrieval_query",
        )
        return result["embedding"]


class IntentClassifier:
    """
    Intent classifier using ChromaDB vector similarity.

    Uses Gemini API for embedding computation on the client side,
    then stores/queries vectors in ChromaDB server.
    """

    COLLECTION_NAME = "intent_examples_gemini"

    def __init__(self, use_persistent: bool = True):
        """
        Initialize the classifier.

        Args:
            use_persistent: If True, connect to ChromaDB server.
                          If False, use in-memory storage (for testing).
        """
        # Initialize Gemini embedder
        self.embedder = GeminiEmbedder()
        print("[IntentClassifier] Using Gemini Embedding (client-side)")

        # Connect to ChromaDB
        if use_persistent:
            self.client = chromadb.HttpClient(
                host=settings.chroma_host,
                port=settings.chroma_port,
            )
        else:
            self.client = chromadb.Client()

        # Create collection (no embedding_function - we compute embeddings ourselves)
        self.collection = self.client.get_or_create_collection(
            name=self.COLLECTION_NAME,
            metadata={"description": "Intent classification examples (Gemini embedding)"},
        )

        if self.collection.count() == 0:
            self._initialize_collection()

    def _initialize_collection(self):
        """Populate the collection with intent examples."""
        documents = []
        metadatas = []
        ids = []

        for intent in INTENT_DEFINITIONS:
            intent_id = intent["id"]
            for i, example in enumerate(intent["examples"]):
                documents.append(example)
                metadatas.append({
                    "intent_id": intent_id,
                    "intent_name": intent["name"],
                })
                ids.append(f"{intent_id}_{i}")

        # Compute embeddings on client side
        print(f"[IntentClassifier] Computing embeddings for {len(documents)} examples...")
        embeddings = self.embedder.embed(documents)

        # Add to ChromaDB with pre-computed embeddings
        self.collection.add(
            ids=ids,
            embeddings=embeddings,
            metadatas=metadatas,
            documents=documents,
        )
        print(f"[IntentClassifier] Initialized with {len(documents)} examples")

    def classify(self, message: str, top_k: int = 3) -> dict:
        """
        Classify a user message into an intent.

        Args:
            message: The user's message
            top_k: Number of similar examples to consider

        Returns:
            dict with intent_id, intent_name, confidence, similar_examples
        """
        # Compute query embedding on client side
        query_embedding = self.embedder.embed_query(message)

        # Query ChromaDB with pre-computed embedding
        results = self.collection.query(
            query_embeddings=[query_embedding],
            n_results=top_k,
        )

        if not results["metadatas"] or not results["metadatas"][0]:
            return self._default_result()

        # Count votes and calculate similarity for each intent
        intent_scores = {}
        distances = results["distances"][0] if results["distances"] else []

        for i, metadata in enumerate(results["metadatas"][0]):
            intent_id = metadata["intent_id"]
            distance = distances[i] if i < len(distances) else 1.0
            similarity = 1.0 / (1.0 + distance)

            if intent_id not in intent_scores:
                intent_scores[intent_id] = {
                    "total_similarity": 0.0,
                    "count": 0,
                    "examples": [],
                }

            intent_scores[intent_id]["total_similarity"] += similarity
            intent_scores[intent_id]["count"] += 1
            intent_scores[intent_id]["examples"].append({
                "text": results["documents"][0][i],
                "similarity": round(similarity, 3),
            })

        # Find best matching intent
        best_intent_id = max(
            intent_scores.keys(),
            key=lambda x: intent_scores[x]["total_similarity"]
        )

        best = intent_scores[best_intent_id]
        confidence = best["total_similarity"] / best["count"]
        intent_info = INTENT_BY_ID.get(best_intent_id, {})

        return {
            "intent_id": best_intent_id,
            "intent_name": intent_info.get("name", best_intent_id),
            "confidence": round(confidence, 3),
            "similar_examples": best["examples"],
        }

    def _default_result(self) -> dict:
        return {
            "intent_id": "general",
            "intent_name": "一般對話",
            "confidence": 0.0,
            "similar_examples": [],
        }

    def reset(self):
        """Reset the collection (for updating examples)."""
        self.client.delete_collection(self.COLLECTION_NAME)
        self.collection = self.client.create_collection(
            name=self.COLLECTION_NAME,
            metadata={"description": "Intent classification examples (Gemini embedding)"},
        )
        self._initialize_collection()
