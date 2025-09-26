#!/usr/bin/env python3
"""
HealthSecure Document Ingestion System
Ingests medical documents and creates vector embeddings for RAG.
"""

import os
import asyncio
import logging
from pathlib import Path
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
import uuid
from datetime import datetime

import aiomysql

# For parsing connection string
from urllib.parse import urlparse

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class DocumentChunk:
    chunk_id: str
    document_id: str
    content: str
    metadata: Dict[str, Any]
    embedding: Optional[List[float]] = None

class HealthSecureIngestion:
    """Document ingestion pipeline for HealthSecure knowledge base."""

    def __init__(self,
                 db_connection_string: str = None,
                 ollama_base_url: str = "http://localhost:11434",
                 embedding_model: str = "mxbai-embed-large",
                 chunk_size: int = 1000,
                 chunk_overlap: int = 200):

        self.db_connection_string = db_connection_string or os.getenv(
            "DATABASE_URL",
            "mysql+aiomysql://healthsecure_user:password@localhost:3306/healthsecure"
        )
        self.ollama_base_url = ollama_base_url
        self.embedding_model = embedding_model
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap

        # Initialize embeddings client
        self.embeddings = OllamaEmbeddings(
            model=embedding_model,
            base_url=ollama_base_url
        )

        # Initialize text splitter
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=chunk_size,
            chunk_overlap=chunk_overlap,
            separators=["\n\n", "\n", ". ", " ", ""]
        )

    async def _get_db_connection_params(self) -> Dict[str, Any]:
        """Parses the database connection string into aiomysql parameters."""
        parsed_url = urlparse(self.db_connection_string)
        return {
            "host": parsed_url.hostname or "localhost",
            "port": parsed_url.port or 3306,
            "user": parsed_url.username or "healthsecure_user",
            "password": parsed_url.password or "password",
            "db": parsed_url.path.lstrip('/') or "healthsecure",
            "autocommit": True
        }

    async def setup_database(self):
        """Initialize database tables for document storage."""
        try:
            db_params = await self._get_db_connection_params()
            conn = await aiomysql.connect(**db_params)
            async with conn.cursor() as cur:
                # Create documents table
                await cur.execute("""
                    CREATE TABLE IF NOT EXISTS rag_documents (
                        id VARCHAR(36) PRIMARY KEY,
                        title VARCHAR(255) NOT NULL,
                        source_path VARCHAR(500) NOT NULL,
                        document_type VARCHAR(50) NOT NULL,
                        file_size INT,
                        content_hash VARCHAR(64),
                        metadata JSON,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
                    );
                """)

                # Create chunks table with vector embeddings
                await cur.execute("""
                    CREATE TABLE IF NOT EXISTS rag_chunks (
                        id VARCHAR(36) PRIMARY KEY,
                        document_id VARCHAR(36) NOT NULL,
                        chunk_index INT NOT NULL,
                        content TEXT NOT NULL,
                        content_embedding JSON, -- Storing embeddings as JSON string
                        metadata JSON,
                        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                        FOREIGN KEY (document_id) REFERENCES rag_documents(id) ON DELETE CASCADE
                    );
                """)

                # Create indexes for better search performance
                await cur.execute("""
                    CREATE INDEX IF NOT EXISTS idx_rag_chunks_document_id
                    ON rag_chunks(document_id);
                """)

                # Note: MySQL does not have native vector indexing like PostgreSQL's pgvector.
                # Vector similarity search will typically be done via UDFs or application-side calculation.
                # If you have a custom UDF for vector similarity, you might add an index on content_embedding
                # if your UDF can leverage it (e.g., a spatial index if using specific vector types).

            await conn.commit()
            conn.close()

            logger.info("Database tables created successfully")

        except Exception as e:
            logger.error(f"Failed to setup database: {e}")
            raise

    def load_document(self, file_path: str) -> List[DocumentChunk]:
        """Load and process a single document."""
        try:
            path = Path(file_path)
            document_id = str(uuid.uuid4())

            # Choose appropriate loader based on file extension
            if path.suffix.lower() == '.pdf':
                loader = PyPDFLoader(str(path))
            elif path.suffix.lower() in ['.txt', '.md']:
                loader = TextLoader(str(path))
            elif path.suffix.lower() in ['.doc', '.docx']:
                loader = UnstructuredWordDocumentLoader(str(path))
            elif path.suffix.lower() in ['.html', '.htm']:
                loader = UnstructuredHTMLLoader(str(path))
            else:
                logger.warning(f"Unsupported file type: {path.suffix}")
                return []

            # Load document
            documents = loader.load()

            # Split into chunks
            chunks = []
            for doc_idx, doc in enumerate(documents):
                text_chunks = self.text_splitter.split_text(doc.page_content)

                for chunk_idx, chunk_text in enumerate(text_chunks):
                    chunk = DocumentChunk(
                        chunk_id=str(uuid.uuid4()),
                        document_id=document_id,
                        content=chunk_text.strip(),
                        metadata={
                            "source": str(path),
                            "document_title": path.name,
                            "document_type": path.suffix.lower(),
                            "chunk_index": chunk_idx,
                            "page_number": doc_idx + 1,
                            "file_size": path.stat().st_size,
                            "created_at": datetime.now().isoformat(),
                            **doc.metadata
                        }
                    )
                    chunks.append(chunk)

            logger.info(f"Loaded {len(chunks)} chunks from {path.name}")
            return chunks

        except Exception as e:
            logger.error(f"Failed to load document {file_path}: {e}")
            return []

    def generate_embeddings(self, chunks: List[DocumentChunk]) -> List[DocumentChunk]:
        """Generate embeddings for document chunks."""
        try:
            texts = [chunk.content for chunk in chunks]
            embeddings = self.embeddings.embed_documents(texts)

            for chunk, embedding in zip(chunks, embeddings):
                chunk.embedding = embedding

            logger.info(f"Generated embeddings for {len(chunks)} chunks")
            return chunks

        except Exception as e:
            logger.error(f"Failed to generate embeddings: {e}")
            return chunks

    async def store_document(self, chunks: List[DocumentChunk]):
        """Store document and chunks in database."""
        if not chunks:
            return

        try:
            db_params = await self._get_db_connection_params()
            conn = await aiomysql.connect(**db_params)
            async with conn.cursor() as cur:

                # Get document info from first chunk
                first_chunk = chunks[0]
                metadata = first_chunk.metadata

                # Insert document record
                await cur.execute("""
                    INSERT INTO rag_documents (id, title, source_path, document_type, file_size, metadata)
                    VALUES (%s, %s, %s, %s, %s, %s)
                    ON DUPLICATE KEY UPDATE
                        title = VALUES(title),
                        source_path = VALUES(source_path),
                        document_type = VALUES(document_type),
                        file_size = VALUES(file_size),
                        metadata = VALUES(metadata),
                        updated_at = CURRENT_TIMESTAMP
                """, (
                    first_chunk.document_id,
                    metadata.get("document_title", "Unknown"),
                    metadata.get("source", "Unknown"),
                    metadata.get("document_type", "unknown"),
                    metadata.get("file_size", 0),
                    json.dumps(metadata) # Store metadata as JSON string
                ))

                # Prepare chunk data for batch insert
                chunk_data_values = []
                for chunk in chunks:
                    chunk_data_values.append((
                        chunk.chunk_id,
                        chunk.document_id,
                        chunk.metadata.get("chunk_index", 0),
                        chunk.content,
                        json.dumps(chunk.embedding) if chunk.embedding else None, # Store embedding as JSON string
                        json.dumps(chunk.metadata) # Store metadata as JSON string
                    ))

                # Batch insert chunks using executemany
                # aiomysql does not have execute_values, so we use executemany
                # For very large batches, consider breaking into smaller chunks or using a different approach.
                insert_chunk_sql = """
                    INSERT INTO rag_chunks (id, document_id, chunk_index, content, content_embedding, metadata)
                    VALUES (%s, %s, %s, %s, %s, %s)
                """
                await cur.executemany(insert_chunk_sql, chunk_data_values)

            await conn.commit()
            conn.close()

            logger.info(f"Stored {len(chunks)} chunks for document {first_chunk.document_id}")

        except Exception as e:
            logger.error(f"Failed to store document: {e}")
            raise

    def ingest_document(self, file_path: str):
        """Complete ingestion pipeline for a single document."""
        logger.info(f"Starting ingestion for: {file_path}")

        # Load and process document
        chunks = self.load_document(file_path)
        if not chunks:
            logger.warning(f"No chunks extracted from {file_path}")
            return

        # Generate embeddings
        chunks = self.generate_embeddings(chunks)

        # Store in database
        self.store_document(chunks)

        logger.info(f"Successfully ingested: {file_path}")

    def ingest_directory(self, directory_path: str, file_patterns: List[str] = None):
        """Ingest all documents in a directory."""
        if file_patterns is None:
            file_patterns = ["*.pdf", "*.txt", "*.md", "*.doc", "*.docx", "*.html", "*.htm"]

        path = Path(directory_path)
        if not path.exists():
            logger.error(f"Directory not found: {directory_path}")
            return

        files_to_process = []
        for pattern in file_patterns:
            files_to_process.extend(path.glob(pattern))

        logger.info(f"Found {len(files_to_process)} files to process")

        for file_path in files_to_process:
            try:
                self.ingest_document(str(file_path))
            except Exception as e:
                logger.error(f"Failed to process {file_path}: {e}")
                continue

        logger.info(f"Completed directory ingestion: {directory_path}")

    async def clear_documents(self):
        """Clear all documents from the knowledge base."""
        try:
            db_params = await self._get_db_connection_params()
            conn = await aiomysql.connect(**db_params)
            async with conn.cursor() as cur:
                await cur.execute("DELETE FROM rag_chunks;")
                await cur.execute("DELETE FROM rag_documents;")
            await conn.commit()
            conn.close()

            logger.info("Cleared all documents from knowledge base")

        except Exception as e:
            logger.error(f"Failed to clear documents: {e}")
            raise

async def main():
    """CLI interface for document ingestion."""
    import argparse

    parser = argparse.ArgumentParser(description="HealthSecure Document Ingestion")
    parser.add_argument("--documents", "-d", required=True, help="Path to documents directory")
    parser.add_argument("--clear", action="store_true", help="Clear existing documents first")
    parser.add_argument("--setup-db", action="store_true", help="Setup database tables")
    parser.add_argument("--db-url", help="Database connection URL")
    parser.add_argument("--embedding-model", default="mxbai-embed-large", help="Embedding model name")

    args = parser.parse_args()

    # Initialize ingestion pipeline
    ingestion = HealthSecureIngestion(
        db_connection_string=args.db_url,
        embedding_model=args.embedding_model
    )

    # Setup database if requested
    if args.setup_db:
        logger.info("Setting up database tables...")
        await ingestion.setup_database()

    # Clear existing documents if requested
    if args.clear:
        logger.info("Clearing existing documents...")
        await ingestion.clear_documents()

    # Ingest documents
    logger.info(f"Starting document ingestion from: {args.documents}")
    await ingestion.ingest_directory(args.documents)
    logger.info("Document ingestion completed!")

if __name__ == "__main__":
    asyncio.run(main())