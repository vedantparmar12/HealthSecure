#!/usr/bin/env python3
"""
Enhanced PDF RAG Analyzer using Docling
Replaces PyPDF2/PyMuPDF with Docling for better OCR, chunking, and batch processing.
"""

import os
import json
import asyncio
import logging
from typing import Dict, List, Any, Optional
from datetime import datetime
import hashlib
from pathlib import Path

from docling.document_converter import DocumentConverter
from docling.datamodel.base_models import InputFormat
from docling.datamodel.pipeline_options import PdfPipelineOptions, TableFormerMode
from docling.chunking import HybridChunker
from docling_core.types.doc import ImageRefMode, PictureItem, TableItem

from langchain_community.embeddings import OllamaEmbeddings
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class DoclingRAGAnalyzer:
    """Advanced PDF analyzer using Docling with RAG capabilities."""

    def __init__(self, ollama_base_url: str = "http://localhost:11434"):
        self.ollama_base_url = ollama_base_url
        
        # Initialize embeddings
        self.embeddings = OllamaEmbeddings(
            model="mxbai-embed-large",
            base_url=ollama_base_url
        )

        # Qdrant configuration
        self.qdrant_url = os.getenv("QDRANT_URL", "http://localhost:6333")
        self.qdrant_api_key = os.getenv("QDRANT_API_KEY", None)
        self.collection_name = os.getenv("QDRANT_COLLECTION_NAME", "healthsecure_medical_docs")
        self.vector_size = int(os.getenv("QDRANT_VECTOR_SIZE", "1024"))  # mxbai-embed-large uses 1024

        # Initialize Qdrant client
        self.qdrant_client = None
        self._init_qdrant_client()

        # Initialize Docling converter with advanced options
        self._init_docling_converter()

        # Initialize Docling chunker
        self._init_chunker()

        # Cache for processed files
        self.processed_files = {}

    def _init_qdrant_client(self):
        """Initialize Qdrant client and create collection if needed."""
        try:
            if self.qdrant_api_key:
                self.qdrant_client = QdrantClient(
                    url=self.qdrant_url,
                    api_key=self.qdrant_api_key,
                    check_compatibility=False
                )
            else:
                self.qdrant_client = QdrantClient(url=self.qdrant_url, check_compatibility=False)

            # Create collection if it doesn't exist
            collections = self.qdrant_client.get_collections().collections
            collection_exists = any(col.name == self.collection_name for col in collections)

            if not collection_exists:
                self.qdrant_client.create_collection(
                    collection_name=self.collection_name,
                    vectors_config=VectorParams(
                        size=self.vector_size,
                        distance=Distance.COSINE
                    )
                )
                logger.info(f"Created Qdrant collection: {self.collection_name}")
            else:
                logger.info(f"Using existing Qdrant collection: {self.collection_name}")

        except Exception as e:
            logger.warning(f"Failed to initialize Qdrant client: {e}")
            self.qdrant_client = None

    def _init_docling_converter(self):
        """Initialize Docling converter with OCR and table extraction."""
        # Configure PDF pipeline with OCR and advanced table extraction
        pipeline_options = PdfPipelineOptions(
            do_ocr=True,  # Enable OCR for scanned documents
            do_table_structure=True,  # Extract table structure
            table_structure_options={
                "mode": TableFormerMode.ACCURATE,  # Use accurate mode for medical tables
                "do_cell_matching": True
            }
        )

        # Initialize converter without pipeline_options parameter (use default initialization)
        self.converter = DocumentConverter(
            allowed_formats=[
                InputFormat.PDF,
                InputFormat.DOCX,
                InputFormat.PPTX,
                InputFormat.IMAGE,
                InputFormat.HTML
            ]
        )
        logger.info("Docling converter initialized with OCR and table extraction")

    def _init_chunker(self):
        """Initialize Docling HybridChunker for smart chunking."""
        # Use HybridChunker with tokenizer-aware chunking
        # This combines hierarchical structure with optimal token sizes
        self.chunker = HybridChunker(
            tokenizer="sentence-transformers/all-MiniLM-L6-v2",  # Fast tokenizer
            max_tokens=512,  # Optimal for embeddings
            merge_peers=True  # Merge small chunks intelligently
        )
        logger.info("Docling HybridChunker initialized")

    async def analyze_pdf(self, file_path: str, patient_context: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Complete analysis of a medical PDF using Docling."""
        logger.info(f"Starting Docling analysis of: {file_path}")

        try:
            # Generate file hash for caching
            file_hash = self._calculate_file_hash(file_path)

            # Check if already processed
            if file_hash in self.processed_files:
                logger.info(f"Using cached analysis for {file_path}")
                return self.processed_files[file_hash]

            # Convert document using Docling
            result = self.converter.convert(file_path)
            doc = result.document

            # Export to different formats
            markdown_content = doc.export_to_markdown()
            json_content = doc.export_to_dict()

            # Extract structured data
            structured_data = self._extract_structured_data(doc, patient_context)

            # Create smart chunks using HybridChunker
            chunks = list(self.chunker.chunk(doc))

            # Create RAG chunks with embeddings
            rag_chunks = await self._create_rag_chunks(chunks, file_path, doc)

            # Store in Qdrant
            await self.store_in_qdrant(rag_chunks)

            # Generate analysis summary
            analysis_summary = self._generate_analysis_summary(doc, structured_data)

            result_data = {
                "file_path": file_path,
                "file_hash": file_hash,
                "total_pages": len(doc.pages),
                "markdown": markdown_content,
                "json": json_content,
                "structured_data": structured_data,
                "rag_chunks": rag_chunks,
                "analysis_summary": analysis_summary,
                "processed_at": datetime.now().isoformat()
            }

            # Cache the result
            self.processed_files[file_hash] = result_data

            logger.info(f"Docling analysis completed: {len(chunks)} chunks created")
            return result_data

        except Exception as e:
            logger.error(f"Docling analysis failed: {e}")
            raise

    def _calculate_file_hash(self, file_path: str) -> str:
        """Calculate SHA-256 hash of the file."""
        hash_sha256 = hashlib.sha256()
        with open(file_path, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                hash_sha256.update(chunk)
        return hash_sha256.hexdigest()

    def _extract_structured_data(self, doc, patient_context: Optional[Dict[str, Any]]) -> Dict[str, Any]:
        """Extract structured medical data from Docling document."""
        structured_data = {
            "patient_info": {},
            "vital_signs": {},
            "lab_values": {},
            "diagnoses": [],
            "medications": [],
            "tables": [],
            "images": []
        }

        # Extract tables (common in medical reports)
        for item, level in doc.iterate_items():
            if isinstance(item, TableItem):
                structured_data["tables"].append({
                    "data": item.export_to_dataframe().to_dict() if hasattr(item, 'export_to_dataframe') else {},
                    "caption": getattr(item, 'caption', ''),
                    "page": getattr(item, 'page', None)
                })
            elif isinstance(item, PictureItem):
                structured_data["images"].append({
                    "caption": getattr(item, 'caption', ''),
                    "page": getattr(item, 'page', None)
                })

        # Extract metadata
        if hasattr(doc, 'meta'):
            structured_data["metadata"] = {
                "title": getattr(doc.meta, 'title', ''),
                "authors": getattr(doc.meta, 'authors', []),
                "creation_date": getattr(doc.meta, 'creation_date', ''),
            }

        return structured_data

    async def _create_rag_chunks(self, chunks, source_file: str, doc) -> List[Dict[str, Any]]:
        """Create RAG-ready chunks from Docling chunks with embeddings."""
        rag_chunks = []

        for i, chunk in enumerate(chunks):
            chunk_text = chunk.text if hasattr(chunk, 'text') else str(chunk)
            
            # Generate embedding
            try:
                embedding = await asyncio.to_thread(
                    self.embeddings.embed_query,
                    chunk_text
                )

                rag_chunks.append({
                    "chunk_id": f"{Path(source_file).stem}_{i}",
                    "content": chunk_text,
                    "embedding": embedding,
                    "metadata": {
                        "source_file": source_file,
                        "chunk_index": i,
                        "character_count": len(chunk_text),
                        "page": getattr(chunk, 'page', None),
                        "doc_items": getattr(chunk, 'doc_items', [])
                    }
                })
            except Exception as e:
                logger.warning(f"Failed to generate embedding for chunk {i}: {e}")
                rag_chunks.append({
                    "chunk_id": f"{Path(source_file).stem}_{i}",
                    "content": chunk_text,
                    "embedding": None,
                    "metadata": {
                        "source_file": source_file,
                        "chunk_index": i,
                        "character_count": len(chunk_text)
                    }
                })

        return rag_chunks

    async def store_in_qdrant(self, chunks: List[Dict[str, Any]]):
        """Store RAG chunks in Qdrant vector database."""
        if not self.qdrant_client:
            logger.warning("Qdrant client not available, skipping storage")
            return

        points = []
        for chunk in chunks:
            if chunk.get('embedding'):
                point = PointStruct(
                    id=hash(chunk['chunk_id']) % (2**63 - 1),
                    vector=chunk['embedding'],
                    payload={
                        'chunk_id': chunk['chunk_id'],
                        'content': chunk['content'],
                        'metadata': chunk['metadata']
                    }
                )
                points.append(point)

        if points:
            try:
                self.qdrant_client.upsert(
                    collection_name=self.collection_name,
                    points=points
                )
                logger.info(f"Stored {len(points)} chunks in Qdrant")
            except Exception as e:
                logger.error(f"Failed to store chunks in Qdrant: {e}")

    async def search_similar_chunks(self, 
                                    query: str, 
                                    limit: int = 5, 
                                    use_reranker: bool = True,
                                    use_hybrid: bool = False) -> List[Dict[str, Any]]:
        """
        Search for similar chunks with optional hybrid retrieval and reranking.
        
        Args:
            query: Search query
            limit: Number of results
            use_reranker: Apply reranking (default: True)
            use_hybrid: Use hybrid retrieval Dense+BM25+ColBERT (default: False)
        """
        if not self.qdrant_client:
            logger.warning("Qdrant client not available")
            return []

        try:
            # HYBRID RETRIEVAL: Dense + BM25 + ColBERT
            if use_hybrid:
                try:
                    from hybrid_retriever import UltimateHybridRetriever
                    
                    # Initialize hybrid retriever (if not cached)
                    if not hasattr(self, '_hybrid_retriever'):
                        self._hybrid_retriever = UltimateHybridRetriever(
                            qdrant_client=self.qdrant_client,
                            embeddings=self.embeddings,
                            collection_name=self.collection_name
                        )
                        logger.info("🚀 Hybrid retriever initialized")
                    
                    # Use hybrid search
                    results = await self._hybrid_retriever.search(
                        query=query,
                        top_k=limit * 3 if use_reranker else limit
                    )
                    logger.info(f"✅ Hybrid retrieval (Dense+BM25+ColBERT) used")
                    
                except Exception as e:
                    logger.warning(f"Hybrid retrieval failed: {e}, falling back to dense only")
                    use_hybrid = False
            
            # STANDARD DENSE RETRIEVAL: Qdrant only
            if not use_hybrid:
                # Generate query embedding
                query_embedding = await asyncio.to_thread(
                    self.embeddings.embed_query,
                    query
                )

                # Search in Qdrant - get more results for reranking
                search_limit = limit * 3 if use_reranker else limit
                search_results = self.qdrant_client.search(
                    collection_name=self.collection_name,
                    query_vector=query_embedding,
                    limit=search_limit,
                    with_payload=True
                )

                # Format results
                results = []
                for hit in search_results:
                    results.append({
                        'chunk_id': hit.payload['chunk_id'],
                        'content': hit.payload['content'],
                        'similarity': hit.score,
                        'metadata': hit.payload['metadata']
                    })

            # Apply reranking if enabled
            if use_reranker and results:
                try:
                    # Try advanced hybrid reranker first (best performance)
                    try:
                        from advanced_reranker import create_advanced_reranker
                        reranker = create_advanced_reranker("balanced")  # balanced = fast + accurate
                        reranked = reranker.rerank(query, results, top_k=limit)
                        logger.info(f"✅ Advanced hybrid reranking applied")
                    except ImportError:
                        # Fallback to basic reranker
                        from reranker import create_reranker
                        reranker = create_reranker("cross-encoder")
                        reranked = reranker.rerank(query, results, top_k=limit)
                        logger.info(f"✅ Basic cross-encoder reranking applied")
                    
                    # Convert back to dict format
                    results = []
                    for r in reranked:
                        result_dict = {
                            'chunk_id': r.chunk_id,
                            'content': r.content,
                            'similarity': r.original_score,
                            'rerank_score': r.rerank_score,
                            'metadata': r.metadata
                        }
                        
                        # Add detailed scores if available
                        if hasattr(r, 'flashrank_score') and r.flashrank_score:
                            result_dict['flashrank_score'] = r.flashrank_score
                        if hasattr(r, 'cross_encoder_score') and r.cross_encoder_score:
                            result_dict['cross_encoder_score'] = r.cross_encoder_score
                        if hasattr(r, 'mxbai_score') and r.mxbai_score:
                            result_dict['mxbai_score'] = r.mxbai_score
                        if hasattr(r, 'keyword_score') and r.keyword_score:
                            result_dict['keyword_score'] = r.keyword_score
                        
                        results.append(result_dict)
                    
                    logger.info(f"Reranked {len(results)} results")
                except Exception as e:
                    logger.warning(f"Reranking failed, using original order: {e}")
                    results = results[:limit]
            
            return results

        except Exception as e:
            logger.error(f"Failed to search in Qdrant: {e}")
            return []

    async def batch_process_documents(self, directory_path: str, file_patterns: List[str] = None) -> Dict[str, Any]:
        """Batch process multiple documents using Docling."""
        if file_patterns is None:
            file_patterns = ["*.pdf", "*.docx", "*.pptx", "*.png", "*.jpg"]

        path = Path(directory_path)
        if not path.exists():
            logger.error(f"Directory not found: {directory_path}")
            return {"error": "Directory not found"}

        files_to_process = []
        for pattern in file_patterns:
            files_to_process.extend(path.glob(pattern))

        logger.info(f"Starting batch processing of {len(files_to_process)} files")

        results = {
            "total_files": len(files_to_process),
            "processed": [],
            "failed": []
        }

        for file_path in files_to_process:
            try:
                result = await self.analyze_pdf(str(file_path))
                results["processed"].append({
                    "file": file_path.name,
                    "chunks": len(result["rag_chunks"]),
                    "status": "success"
                })
            except Exception as e:
                logger.error(f"Failed to process {file_path}: {e}")
                results["failed"].append({
                    "file": file_path.name,
                    "error": str(e)
                })

        logger.info(f"Batch processing completed: {len(results['processed'])} succeeded, {len(results['failed'])} failed")
        return results

    def _generate_analysis_summary(self, doc, structured_data: Dict[str, Any]) -> str:
        """Generate comprehensive analysis summary."""
        summary_parts = []

        summary_parts.append("Medical Report Analysis (Docling Enhanced)")
        summary_parts.append("=" * 50)

        summary_parts.append(f"Total Pages: {len(doc.pages)}")
        summary_parts.append(f"Tables Extracted: {len(structured_data.get('tables', []))}")
        summary_parts.append(f"Images Found: {len(structured_data.get('images', []))}")

        if structured_data.get('metadata'):
            meta = structured_data['metadata']
            if meta.get('title'):
                summary_parts.append(f"Title: {meta['title']}")
            if meta.get('authors'):
                summary_parts.append(f"Authors: {', '.join(meta['authors'])}")

        return '\n'.join(summary_parts)


# Test function
async def test_docling_analyzer():
    """Test the Docling RAG analyzer."""
    analyzer = DoclingRAGAnalyzer()
    logger.info("Docling RAG Analyzer initialized successfully")
    
    # Test with a sample PDF
    # result = await analyzer.analyze_pdf("/path/to/medical/report.pdf")
    # print(result["analysis_summary"])


if __name__ == "__main__":
    asyncio.run(test_docling_analyzer())
