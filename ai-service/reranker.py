#!/usr/bin/env python3
"""
Reranker module for RAG with multiple reranking strategies.
Significantly improves retrieval accuracy for medical documents.
"""

import logging
from typing import List, Dict, Any, Optional
from dataclasses import dataclass

logger = logging.getLogger(__name__)

@dataclass
class RerankResult:
    """Reranked document with updated score."""
    chunk_id: str
    content: str
    original_score: float
    rerank_score: float
    metadata: Dict[str, Any]


class BaseReranker:
    """Base class for rerankers."""
    
    def rerank(self, query: str, documents: List[Dict[str, Any]], top_k: int = 5) -> List[RerankResult]:
        """Rerank documents based on query relevance."""
        raise NotImplementedError


class CrossEncoderReranker(BaseReranker):
    """
    Cross-Encoder reranker using sentence-transformers.
    Best for medical documents with high accuracy.
    """
    
    def __init__(self, model_name: str = "cross-encoder/ms-marco-MiniLM-L-6-v2"):
        """
        Initialize cross-encoder reranker.
        
        Models:
        - cross-encoder/ms-marco-MiniLM-L-6-v2 (fast, general purpose)
        - cross-encoder/ms-marco-MiniLM-L-12-v2 (more accurate)
        - BAAI/bge-reranker-base (best for medical/technical)
        - BAAI/bge-reranker-large (highest accuracy)
        """
        try:
            from sentence_transformers import CrossEncoder
            self.model = CrossEncoder(model_name)
            self.model_name = model_name
            logger.info(f"Cross-encoder reranker loaded: {model_name}")
        except ImportError:
            logger.error("sentence-transformers not installed. Run: pip install sentence-transformers")
            self.model = None
        except Exception as e:
            logger.error(f"Failed to load cross-encoder: {e}")
            self.model = None
    
    def rerank(self, query: str, documents: List[Dict[str, Any]], top_k: int = 5) -> List[RerankResult]:
        """Rerank documents using cross-encoder."""
        if not self.model:
            logger.warning("Cross-encoder not available, returning original order")
            return self._fallback_rerank(query, documents, top_k)
        
        try:
            # Prepare query-document pairs
            pairs = [[query, doc['content']] for doc in documents]
            
            # Get cross-encoder scores
            scores = self.model.predict(pairs)
            
            # Create reranked results
            results = []
            for doc, score in zip(documents, scores):
                results.append(RerankResult(
                    chunk_id=doc['chunk_id'],
                    content=doc['content'],
                    original_score=doc.get('similarity', 0.0),
                    rerank_score=float(score),
                    metadata=doc.get('metadata', {})
                ))
            
            # Sort by rerank score (descending)
            results.sort(key=lambda x: x.rerank_score, reverse=True)
            
            return results[:top_k]
            
        except Exception as e:
            logger.error(f"Cross-encoder reranking failed: {e}")
            return self._fallback_rerank(query, documents, top_k)
    
    def _fallback_rerank(self, query: str, documents: List[Dict[str, Any]], top_k: int) -> List[RerankResult]:
        """Fallback to original order if reranking fails."""
        results = []
        for doc in documents[:top_k]:
            results.append(RerankResult(
                chunk_id=doc['chunk_id'],
                content=doc['content'],
                original_score=doc.get('similarity', 0.0),
                rerank_score=doc.get('similarity', 0.0),
                metadata=doc.get('metadata', {})
            ))
        return results


class CohereReranker(BaseReranker):
    """
    Cohere Rerank API - cloud-based, highly accurate.
    ⚠️ NOT RECOMMENDED: Requires API key and costs money.
    ✅ RECOMMENDED: Use advanced_reranker.py instead (FREE, local, no API)
    """
    
    def __init__(self, api_key: Optional[str] = None):
        """
        Initialize Cohere reranker.
        
        Get free API key at: https://cohere.com/
        """
        try:
            import cohere
            import os
            
            self.api_key = api_key or os.getenv("COHERE_API_KEY")
            if not self.api_key:
                logger.warning("Cohere API key not found. Set COHERE_API_KEY env variable.")
                self.client = None
                return
            
            self.client = cohere.Client(self.api_key)
            logger.info("Cohere reranker initialized")
            
        except ImportError:
            logger.error("cohere not installed. Run: pip install cohere")
            self.client = None
        except Exception as e:
            logger.error(f"Failed to initialize Cohere: {e}")
            self.client = None
    
    def rerank(self, query: str, documents: List[Dict[str, Any]], top_k: int = 5) -> List[RerankResult]:
        """Rerank documents using Cohere API."""
        if not self.client:
            logger.warning("Cohere not available, returning original order")
            return self._fallback_rerank(query, documents, top_k)
        
        try:
            # Prepare documents for Cohere
            docs = [doc['content'] for doc in documents]
            
            # Call Cohere rerank
            response = self.client.rerank(
                query=query,
                documents=docs,
                top_n=top_k,
                model="rerank-english-v3.0"  # or "rerank-multilingual-v3.0"
            )
            
            # Create reranked results
            results = []
            for result in response.results:
                idx = result.index
                original_doc = documents[idx]
                
                results.append(RerankResult(
                    chunk_id=original_doc['chunk_id'],
                    content=original_doc['content'],
                    original_score=original_doc.get('similarity', 0.0),
                    rerank_score=result.relevance_score,
                    metadata=original_doc.get('metadata', {})
                ))
            
            return results
            
        except Exception as e:
            logger.error(f"Cohere reranking failed: {e}")
            return self._fallback_rerank(query, documents, top_k)
    
    def _fallback_rerank(self, query: str, documents: List[Dict[str, Any]], top_k: int) -> List[RerankResult]:
        """Fallback to original order."""
        results = []
        for doc in documents[:top_k]:
            results.append(RerankResult(
                chunk_id=doc['chunk_id'],
                content=doc['content'],
                original_score=doc.get('similarity', 0.0),
                rerank_score=doc.get('similarity', 0.0),
                metadata=doc.get('metadata', {})
            ))
        return results


class HybridReranker(BaseReranker):
    """
    Hybrid reranker combining multiple strategies.
    Uses cross-encoder + keyword matching for best results.
    """
    
    def __init__(self, cross_encoder_model: str = "BAAI/bge-reranker-base"):
        self.cross_encoder = CrossEncoderReranker(cross_encoder_model)
        logger.info("Hybrid reranker initialized")
    
    def rerank(self, query: str, documents: List[Dict[str, Any]], top_k: int = 5) -> List[RerankResult]:
        """Rerank using hybrid approach."""
        # First, use cross-encoder
        results = self.cross_encoder.rerank(query, documents, top_k * 2)
        
        # Then apply keyword boost
        query_keywords = set(query.lower().split())
        
        for result in results:
            content_keywords = set(result.content.lower().split())
            keyword_overlap = len(query_keywords & content_keywords) / len(query_keywords)
            
            # Boost score if high keyword overlap
            result.rerank_score = result.rerank_score * (1 + keyword_overlap * 0.1)
        
        # Re-sort and return top_k
        results.sort(key=lambda x: x.rerank_score, reverse=True)
        return results[:top_k]


def create_reranker(reranker_type: str = "cross-encoder", **kwargs) -> BaseReranker:
    """
    Factory function to create reranker.
    
    Args:
        reranker_type: "cross-encoder", "cohere", or "hybrid"
        **kwargs: Additional arguments for reranker
    
    Returns:
        Reranker instance
    """
    reranker_map = {
        "cross-encoder": CrossEncoderReranker,
        "cohere": CohereReranker,
        "hybrid": HybridReranker
    }
    
    reranker_class = reranker_map.get(reranker_type)
    if not reranker_class:
        logger.warning(f"Unknown reranker type: {reranker_type}, using cross-encoder")
        reranker_class = CrossEncoderReranker
    
    return reranker_class(**kwargs)


# Test function
def test_reranker():
    """Test reranker with sample documents."""
    documents = [
        {
            'chunk_id': 'doc1',
            'content': 'Patient blood pressure is 145/92 mmHg, elevated hypertension.',
            'similarity': 0.75
        },
        {
            'chunk_id': 'doc2',
            'content': 'Heart rate measured at 88 bpm, within normal range.',
            'similarity': 0.70
        },
        {
            'chunk_id': 'doc3',
            'content': 'The weather is sunny today with clear skies.',
            'similarity': 0.85  # High similarity but irrelevant
        }
    ]
    
    query = "What is the patient's blood pressure?"
    
    # Test cross-encoder
    print("Testing Cross-Encoder Reranker:")
    reranker = CrossEncoderReranker()
    results = reranker.rerank(query, documents, top_k=3)
    
    for i, result in enumerate(results, 1):
        print(f"{i}. Score: {result.rerank_score:.4f} (original: {result.original_score:.2f})")
        print(f"   {result.content[:80]}...")
        print()


if __name__ == "__main__":
    test_reranker()
