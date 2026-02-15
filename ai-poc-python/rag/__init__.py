"""
RAG Module for Intent Classification and Dynamic Prompt Composition.

This module provides:
- Intent classification using ChromaDB embeddings
- Dynamic prompt composition based on detected intent
"""

from .intent_classifier import IntentClassifier
from .prompt_retriever import PromptRetriever

__all__ = ["IntentClassifier", "PromptRetriever"]
