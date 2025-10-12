#!/usr/bin/env python3
"""
PDF RAG Analyzer for Medical Reports
Processes patient reports, lab results, and medical documents using RAG.
"""

import os
import json
import asyncio
import logging
from typing import Dict, List, Any, Optional
from datetime import datetime
import hashlib

import PyPDF2
import fitz  # PyMuPDF for better PDF text extraction
from PIL import Image
import io
import base64

from langchain_community.embeddings import OllamaEmbeddings
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain_community.vectorstores import Qdrant
from langchain.schema import Document
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class MedicalReportSection(dict):
    """Structured representation of a medical report section."""
    def __init__(self, **data):
        super().__init__(data)
        self.section_type = data.get('section_type', 'unknown')
        self.content = data.get('content', '')
        self.confidence = data.get('confidence', 0.0)
        self.page_number = data.get('page_number', 1)
        self.extraction_method = data.get('extraction_method', 'text')

class PDFAnalysisResult(dict):
    """Complete PDF analysis result."""
    def __init__(self, **data):
        super().__init__(data)
        self.file_path = data.get('file_path', '')
        self.file_hash = data.get('file_hash', '')
        self.total_pages = data.get('total_pages', 0)
        self.sections = data.get('sections', [])
        self.extracted_data = data.get('extracted_data', {})
        self.rag_chunks = data.get('rag_chunks', [])
        self.analysis_summary = data.get('analysis_summary', '')
        self.processed_at = data.get('processed_at', datetime.now().isoformat())

class PDFRAGAnalyzer:
    """Advanced PDF analyzer with RAG capabilities for medical reports."""

    def __init__(self, ollama_base_url: str = "http://localhost:11434"):
        self.ollama_base_url = ollama_base_url
        self.embeddings = OllamaEmbeddings(
            model="mxbai-embed-large",
            base_url=ollama_base_url
        )
        self.text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=1000,
            chunk_overlap=200,
            separators=["\n\n", "\n", ".", " "]
        )

        # Qdrant configuration
        self.qdrant_url = os.getenv("QDRANT_URL", "http://localhost:6333")
        self.qdrant_api_key = os.getenv("QDRANT_API_KEY", None)
        self.collection_name = os.getenv("QDRANT_COLLECTION_NAME", "healthsecure_medical_docs")
        self.vector_size = int(os.getenv("QDRANT_VECTOR_SIZE", "768"))

        # Initialize Qdrant client
        self.qdrant_client = None
        self._init_qdrant_client()

        # Medical report section patterns
        self.section_patterns = {
            'patient_info': [
                r'patient\s+information',
                r'demographics',
                r'patient\s+details',
                r'name.*dob',
                r'patient\s+id'
            ],
            'chief_complaint': [
                r'chief\s+complaint',
                r'presenting\s+complaint',
                r'reason\s+for\s+visit',
                r'history\s+of\s+present\s+illness'
            ],
            'vital_signs': [
                r'vital\s+signs',
                r'vitals',
                r'blood\s+pressure',
                r'temperature',
                r'pulse\s+rate'
            ],
            'lab_results': [
                r'laboratory\s+results',
                r'lab\s+results',
                r'blood\s+work',
                r'chemistry\s+panel',
                r'complete\s+blood\s+count'
            ],
            'diagnosis': [
                r'diagnosis',
                r'impression',
                r'assessment',
                r'diagnostic\s+impression'
            ],
            'treatment_plan': [
                r'treatment\s+plan',
                r'plan\s+of\s+care',
                r'recommendations',
                r'medications',
                r'follow\s+up'
            ],
            'imaging': [
                r'radiology',
                r'x-ray',
                r'ct\s+scan',
                r'mri',
                r'ultrasound',
                r'imaging\s+results'
            ]
        }

        # Temporary storage for processed files
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

    async def store_in_qdrant(self, chunks: List[Dict[str, Any]]):
        """Store RAG chunks in Qdrant vector database."""
        if not self.qdrant_client:
            logger.warning("Qdrant client not available, skipping storage")
            return

        points = []
        for chunk in chunks:
            if chunk.get('embedding'):
                point = PointStruct(
                    id=hash(chunk['chunk_id']) % (2**63 - 1),  # Convert to valid ID
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

    async def search_similar_chunks(self, query: str, limit: int = 5) -> List[Dict[str, Any]]:
        """Search for similar chunks in Qdrant."""
        if not self.qdrant_client:
            logger.warning("Qdrant client not available")
            return []

        try:
            # Generate query embedding
            query_embedding = await asyncio.to_thread(
                self.embeddings.embed_query,
                query
            )

            # Search in Qdrant
            search_results = self.qdrant_client.search(
                collection_name=self.collection_name,
                query_vector=query_embedding,
                limit=limit,
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

            return results

        except Exception as e:
            logger.error(f"Failed to search in Qdrant: {e}")
            return []

    async def analyze_pdf(self, file_path: str, patient_context: Optional[Dict[str, Any]] = None) -> PDFAnalysisResult:
        """Complete analysis of a medical PDF report."""
        logger.info(f"Starting analysis of PDF: {file_path}")

        try:
            # Generate file hash for caching
            file_hash = self._calculate_file_hash(file_path)

            # Check if already processed
            if file_hash in self.processed_files:
                logger.info(f"Using cached analysis for {file_path}")
                return self.processed_files[file_hash]

            # Extract text and metadata from PDF
            extracted_content = await self._extract_pdf_content(file_path)

            # Classify and structure the content
            sections = await self._classify_sections(extracted_content['text'])

            # Extract structured data
            structured_data = await self._extract_structured_data(sections, patient_context)

            # Create RAG chunks for semantic search
            rag_chunks = await self._create_rag_chunks(extracted_content['text'], file_path)

            # Store chunks in Qdrant for future retrieval
            await self.store_in_qdrant(rag_chunks)

            # Generate analysis summary
            analysis_summary = await self._generate_analysis_summary(sections, structured_data)

            # Create result object
            result = PDFAnalysisResult(
                file_path=file_path,
                file_hash=file_hash,
                total_pages=extracted_content['total_pages'],
                sections=sections,
                extracted_data=structured_data,
                rag_chunks=rag_chunks,
                analysis_summary=analysis_summary,
                images=extracted_content.get('images', [])
            )

            # Cache the result
            self.processed_files[file_hash] = result

            logger.info(f"PDF analysis completed: {len(sections)} sections, {len(rag_chunks)} RAG chunks")
            return result

        except Exception as e:
            logger.error(f"PDF analysis failed: {e}")
            raise

    def _calculate_file_hash(self, file_path: str) -> str:
        """Calculate SHA-256 hash of the file."""
        hash_sha256 = hashlib.sha256()
        with open(file_path, "rb") as f:
            for chunk in iter(lambda: f.read(4096), b""):
                hash_sha256.update(chunk)
        return hash_sha256.hexdigest()

    async def _extract_pdf_content(self, file_path: str) -> Dict[str, Any]:
        """Extract text, images, and metadata from PDF."""
        content = {
            'text': '',
            'total_pages': 0,
            'images': [],
            'metadata': {}
        }

        try:
            # Use PyMuPDF for better text extraction
            pdf_document = fitz.open(file_path)
            content['total_pages'] = len(pdf_document)

            all_text = []
            page_images = []

            for page_num, page in enumerate(pdf_document, 1):
                # Extract text
                page_text = page.get_text()
                if page_text.strip():
                    all_text.append(f"--- Page {page_num} ---\n{page_text}")

                # Extract images
                image_list = page.get_images()
                for img_index, img in enumerate(image_list):
                    xref = img[0]
                    pix = fitz.Pixmap(pdf_document, xref)

                    if pix.n - pix.alpha < 4:  # GRAY or RGB
                        img_data = pix.tobytes("png")
                        img_base64 = base64.b64encode(img_data).decode('utf-8')

                        page_images.append({
                            'page': page_num,
                            'index': img_index,
                            'data': img_base64,
                            'format': 'png',
                            'size': (pix.width, pix.height)
                        })

                    pix = None

            content['text'] = '\n\n'.join(all_text)
            content['images'] = page_images

            # Get metadata
            content['metadata'] = pdf_document.metadata

            pdf_document.close()

        except Exception as e:
            logger.warning(f"PyMuPDF extraction failed, falling back to PyPDF2: {e}")

            # Fallback to PyPDF2
            with open(file_path, 'rb') as file:
                pdf_reader = PyPDF2.PdfReader(file)
                content['total_pages'] = len(pdf_reader.pages)

                all_text = []
                for page_num, page in enumerate(pdf_reader.pages, 1):
                    page_text = page.extract_text()
                    if page_text.strip():
                        all_text.append(f"--- Page {page_num} ---\n{page_text}")

                content['text'] = '\n\n'.join(all_text)

        return content

    async def _classify_sections(self, text: str) -> List[MedicalReportSection]:
        """Classify text into medical report sections."""
        import re

        sections = []
        lines = text.split('\n')
        current_section = None
        current_content = []

        for line_num, line in enumerate(lines):
            line = line.strip()
            if not line:
                continue

            # Check if line matches any section pattern
            section_found = None
            for section_type, patterns in self.section_patterns.items():
                for pattern in patterns:
                    if re.search(pattern, line, re.IGNORECASE):
                        section_found = section_type
                        break
                if section_found:
                    break

            if section_found:
                # Save previous section if exists
                if current_section and current_content:
                    sections.append(MedicalReportSection(
                        section_type=current_section,
                        content='\n'.join(current_content),
                        confidence=0.8,
                        page_number=self._estimate_page_number(line_num, len(lines)),
                        extraction_method='pattern_matching'
                    ))

                # Start new section
                current_section = section_found
                current_content = [line]
            else:
                # Add to current section
                if current_section:
                    current_content.append(line)

        # Add final section
        if current_section and current_content:
            sections.append(MedicalReportSection(
                section_type=current_section,
                content='\n'.join(current_content),
                confidence=0.8,
                page_number=self._estimate_page_number(len(lines), len(lines)),
                extraction_method='pattern_matching'
            ))

        return sections

    def _estimate_page_number(self, line_number: int, total_lines: int) -> int:
        """Estimate page number based on line position."""
        # Rough estimation: assume ~50 lines per page
        return max(1, (line_number // 50) + 1)

    async def _extract_structured_data(self, sections: List[MedicalReportSection], patient_context: Optional[Dict[str, Any]]) -> Dict[str, Any]:
        """Extract structured data from classified sections."""
        import re

        structured_data = {
            'patient_info': {},
            'vital_signs': {},
            'lab_values': {},
            'diagnoses': [],
            'medications': [],
            'key_findings': []
        }

        for section in sections:
            content = section.content

            if section.section_type == 'patient_info':
                # Extract patient information
                patterns = {
                    'name': r'name[:\s]+([^\n]+)',
                    'dob': r'(?:dob|date\s+of\s+birth)[:\s]+([^\n]+)',
                    'age': r'age[:\s]+(\d+)',
                    'gender': r'(?:gender|sex)[:\s]+([^\n]+)',
                    'mrn': r'(?:mrn|medical\s+record)[:\s#]+([^\n]+)'
                }

                for field, pattern in patterns.items():
                    match = re.search(pattern, content, re.IGNORECASE)
                    if match:
                        structured_data['patient_info'][field] = match.group(1).strip()

            elif section.section_type == 'vital_signs':
                # Extract vital signs
                vital_patterns = {
                    'blood_pressure': r'(?:bp|blood\s+pressure)[:\s]+(\d+/\d+)',
                    'temperature': r'(?:temp|temperature)[:\s]+(\d+\.?\d*)\s*[°f°c]?',
                    'heart_rate': r'(?:hr|heart\s+rate|pulse)[:\s]+(\d+)',
                    'respiratory_rate': r'(?:rr|resp\s+rate)[:\s]+(\d+)',
                    'oxygen_saturation': r'(?:o2\s+sat|spo2)[:\s]+(\d+)%?'
                }

                for vital, pattern in vital_patterns.items():
                    match = re.search(pattern, content, re.IGNORECASE)
                    if match:
                        structured_data['vital_signs'][vital] = match.group(1).strip()

            elif section.section_type == 'lab_results':
                # Extract lab values (simplified)
                lab_lines = content.split('\n')
                for line in lab_lines:
                    # Look for patterns like "Glucose: 120 mg/dL"
                    lab_match = re.search(r'([a-zA-Z\s]+)[:\s]+(\d+\.?\d*)\s*([a-zA-Z/]+)?', line)
                    if lab_match:
                        test_name = lab_match.group(1).strip()
                        value = lab_match.group(2).strip()
                        unit = lab_match.group(3).strip() if lab_match.group(3) else ''

                        structured_data['lab_values'][test_name.lower()] = {
                            'value': value,
                            'unit': unit
                        }

            elif section.section_type == 'diagnosis':
                # Extract diagnoses
                diag_lines = content.split('\n')
                for line in diag_lines:
                    line = line.strip()
                    if line and len(line) > 5:  # Skip very short lines
                        # Remove numbers and common prefixes
                        cleaned = re.sub(r'^\d+[\.\)]\s*', '', line)
                        if cleaned:
                            structured_data['diagnoses'].append(cleaned)

        return structured_data

    async def _create_rag_chunks(self, text: str, source_file: str) -> List[Dict[str, Any]]:
        """Create RAG-ready chunks from the PDF text."""
        # Split text into chunks
        documents = [Document(page_content=text, metadata={'source': source_file})]
        chunks = self.text_splitter.split_documents(documents)

        rag_chunks = []
        for i, chunk in enumerate(chunks):
            # Generate embedding for the chunk
            try:
                embedding = await asyncio.to_thread(
                    self.embeddings.embed_query,
                    chunk.page_content
                )

                rag_chunks.append({
                    'chunk_id': f"{source_file}_{i}",
                    'content': chunk.page_content,
                    'embedding': embedding,
                    'metadata': {
                        'source_file': source_file,
                        'chunk_index': i,
                        'character_count': len(chunk.page_content)
                    }
                })
            except Exception as e:
                logger.warning(f"Failed to generate embedding for chunk {i}: {e}")
                # Add chunk without embedding
                rag_chunks.append({
                    'chunk_id': f"{source_file}_{i}",
                    'content': chunk.page_content,
                    'embedding': None,
                    'metadata': {
                        'source_file': source_file,
                        'chunk_index': i,
                        'character_count': len(chunk.page_content)
                    }
                })

        return rag_chunks

    async def _generate_analysis_summary(self, sections: List[MedicalReportSection], structured_data: Dict[str, Any]) -> str:
        """Generate a comprehensive analysis summary."""
        summary_parts = []

        summary_parts.append(f"Medical Report Analysis Summary")
        summary_parts.append("=" * 40)

        # Document structure
        summary_parts.append(f"Document contains {len(sections)} identified sections:")
        for section in sections:
            summary_parts.append(f"  - {section.section_type.replace('_', ' ').title()}")

        # Patient information summary
        if structured_data.get('patient_info'):
            summary_parts.append("\nPatient Information:")
            for key, value in structured_data['patient_info'].items():
                summary_parts.append(f"  - {key.replace('_', ' ').title()}: {value}")

        # Key findings
        if structured_data.get('diagnoses'):
            summary_parts.append(f"\nDiagnoses Identified: {len(structured_data['diagnoses'])}")

        if structured_data.get('lab_values'):
            summary_parts.append(f"\nLab Values Extracted: {len(structured_data['lab_values'])}")

        if structured_data.get('vital_signs'):
            summary_parts.append("\nVital Signs:")
            for vital, value in structured_data['vital_signs'].items():
                summary_parts.append(f"  - {vital.replace('_', ' ').title()}: {value}")

        return '\n'.join(summary_parts)

    async def search_similar_content(self, query: str, file_hash: str, top_k: int = 5) -> List[Dict[str, Any]]:
        """Search for similar content within a processed PDF."""
        if file_hash not in self.processed_files:
            return []

        analysis_result = self.processed_files[file_hash]
        rag_chunks = analysis_result.get('rag_chunks', [])

        if not rag_chunks:
            return []

        try:
            # Generate query embedding
            query_embedding = await asyncio.to_thread(
                self.embeddings.embed_query,
                query
            )

            # Calculate similarities
            similarities = []
            for chunk in rag_chunks:
                if chunk.get('embedding'):
                    # Simple cosine similarity (would use proper vector search in production)
                    similarity = self._cosine_similarity(query_embedding, chunk['embedding'])
                    similarities.append({
                        'chunk': chunk,
                        'similarity': similarity
                    })

            # Sort by similarity and return top results
            similarities.sort(key=lambda x: x['similarity'], reverse=True)
            return [item['chunk'] for item in similarities[:top_k]]

        except Exception as e:
            logger.error(f"Similarity search failed: {e}")
            return []

    def _cosine_similarity(self, vec1: List[float], vec2: List[float]) -> float:
        """Calculate cosine similarity between two vectors."""
        import math

        if len(vec1) != len(vec2):
            return 0.0

        dot_product = sum(a * b for a, b in zip(vec1, vec2))
        magnitude1 = math.sqrt(sum(a * a for a in vec1))
        magnitude2 = math.sqrt(sum(a * a for a in vec2))

        if magnitude1 == 0 or magnitude2 == 0:
            return 0.0

        return dot_product / (magnitude1 * magnitude2)

# Test function
async def test_pdf_analyzer():
    """Test the PDF analyzer with a sample medical report."""
    analyzer = PDFRAGAnalyzer()

    # This would test with an actual medical PDF
    # result = await analyzer.analyze_pdf("/path/to/medical/report.pdf")
    # print(f"Analysis complete: {result.analysis_summary}")

    print("PDF RAG Analyzer initialized successfully")

if __name__ == "__main__":
    asyncio.run(test_pdf_analyzer())