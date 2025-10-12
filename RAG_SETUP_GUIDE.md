# HealthSecure RAG Setup Guide

## 🎯 What Was Fixed

Your RAG (Retrieval Augmented Generation) system wasn't working because:

1. **❌ Qdrant vector database was not running** - Required for storing and searching document embeddings
2. **❌ No documents were ingested** - Empty knowledge base meant no context to retrieve
3. **❌ Using old PDF libraries** - PyPDF2/PyMuPDF lack advanced features

## ✅ What's New

### 1. **Docling Integration** (Replacing PyPDF2/PyMuPDF)
- **Built-in OCR**: Automatically reads scanned PDFs
- **Smart Chunking**: HybridChunker with tokenization-aware splitting
- **Table Extraction**: Preserves complex medical tables with structure
- **Batch Processing**: Efficiently process multiple documents
- **Better Accuracy**: Advanced layout analysis and reading order

### 2. **Qdrant Vector Database**
- High-performance vector similarity search
- Stores document embeddings for semantic search
- Fast retrieval of relevant context for AI responses

### 3. **Enhanced Embeddings**
- Using `mxbai-embed-large` model (1024 dimensions)
- Better semantic understanding of medical content

## 🚀 Quick Start

### Step 1: Install Qdrant

```bash
# Run the setup script
setup_qdrant.bat
```

This will:
- Download Qdrant standalone binary for Windows
- Extract and configure it
- Start Qdrant on port 6333
- Verify it's running

**Alternative (if you have Docker):**
```bash
docker run -p 6333:6333 qdrant/qdrant
```

### Step 2: Install Python Dependencies

```bash
cd ai-service
pip install docling docling-core qdrant-client pymupdf
```

Or update all dependencies:
```bash
pip install -r requirements.txt
```

### Step 3: Verify Setup

```bash
test_rag_setup.bat
```

This checks:
- ✅ Ollama is running
- ✅ Qdrant is running
- ✅ Dependencies are installed
- ✅ DoclingRAGAnalyzer initializes correctly

## 📄 Ingesting Documents

### Test with Single Document

```bash
cd ai-service
python ingest_with_docling.py --test-single "path/to/medical_report.pdf"
```

This will:
1. Process the PDF with Docling (OCR if needed)
2. Extract tables, images, and metadata
3. Create smart chunks using HybridChunker
4. Generate embeddings with Ollama
5. Store in Qdrant vector database
6. Test search functionality

### Batch Process Multiple Documents

```bash
python ingest_with_docling.py --documents "path/to/documents_folder"
```

Supported formats:
- PDF (with or without OCR)
- DOCX
- PPTX
- Images (PNG, JPG)

### Custom File Patterns

```bash
python ingest_with_docling.py --documents "docs" --patterns "*.pdf" "*.docx"
```

## 🔍 How RAG Works Now

### Before (Not Working)
```
User Question → AI → Generic Response (No Context)
```

### After (With RAG)
```
User Question 
    ↓
Docling: Extract & chunk medical documents
    ↓
Ollama: Generate embeddings
    ↓
Qdrant: Store vectors
    ↓
User Query → Qdrant Search → Retrieve relevant chunks
    ↓
AI (with context) → Accurate, source-based response
```

## 📊 What Docling Extracts

### Medical Reports
- Patient information sections
- Vital signs tables
- Lab results with values
- Diagnosis sections
- Treatment plans
- Medical images and charts

### Document Structure
- Tables with preserved structure
- Section headings and hierarchy
- Metadata (title, authors, dates)
- Images with captions
- Reading order preservation

## 🧪 Testing RAG

After ingesting documents, test search:

```python
from docling_rag_analyzer import DoclingRAGAnalyzer
import asyncio

async def test():
    analyzer = DoclingRAGAnalyzer()
    
    # Search for relevant chunks
    results = await analyzer.search_similar_chunks(
        "patient blood pressure readings",
        limit=5
    )
    
    for result in results:
        print(f"Similarity: {result['similarity']:.3f}")
        print(f"Content: {result['content'][:200]}...")
        print(f"Source: {result['metadata']['source_file']}")
        print("-" * 80)

asyncio.run(test())
```

## 🔧 Configuration

Edit `ai-service/.env`:

```env
# Qdrant Configuration
QDRANT_URL=http://localhost:6333
QDRANT_COLLECTION_NAME=healthsecure_medical_docs
QDRANT_VECTOR_SIZE=1024

# Embeddings
EMBEDDING_MODEL=mxbai-embed-large
LLM_BASE_URL=http://localhost:11434/v1
```

## 📁 Project Structure

```
HealthSecure/
├── ai-service/
│   ├── docling_rag_analyzer.py    # NEW: Docling-based RAG
│   ├── pdf_rag_analyzer.py        # OLD: PyPDF2-based (deprecated)
│   ├── ingest_with_docling.py     # NEW: Document ingestion
│   ├── requirements.txt           # Updated dependencies
│   ├── .env                       # Updated config
│   └── sample_docs/               # Place test documents here
├── setup_qdrant.bat               # Qdrant installation
└── test_rag_setup.bat             # Verify setup
```

## 🎓 Advanced Features

### Hybrid Chunking
Docling's HybridChunker combines:
- **Hierarchical chunking**: Respects document structure
- **Token-aware splitting**: Optimal size for embeddings
- **Smart merging**: Combines small chunks intelligently

### OCR Support
- Automatically detects scanned PDFs
- Extracts text from images
- Handles mixed digital/scanned documents

### Table Extraction
- Preserves table structure
- Maintains cell relationships
- Exports to pandas DataFrame

## 🐛 Troubleshooting

### Qdrant Not Running
```bash
# Check if Qdrant is running
curl http://localhost:6333

# If not, start it
cd qdrant
qdrant.exe
```

### Ollama Not Running
```bash
# Check Ollama
curl http://localhost:11434

# Start Ollama if needed
```

### Import Errors
```bash
# Reinstall dependencies
pip install --upgrade docling docling-core qdrant-client
```

### Embedding Model Not Found
```bash
# Pull the model with Ollama
ollama pull mxbai-embed-large
```

## 📈 Performance Tips

1. **Batch Processing**: Process multiple documents at once
2. **Vector Size**: 1024 dimensions balances accuracy and speed
3. **Chunk Size**: 512 tokens optimal for medical documents
4. **Qdrant Indexing**: Automatic HNSW indexing for fast search

## 🔒 Security Notes

- Qdrant runs locally (no cloud uploads)
- Embeddings generated locally with Ollama
- All documents stay on your infrastructure
- HIPAA-compliant deployment ready

## ✨ Next Steps

1. **Ingest Your Documents**: Add medical PDFs to `sample_docs/`
2. **Test Search**: Verify RAG retrieval works
3. **Integrate with AI Service**: Use retrieved context in responses
4. **Monitor Performance**: Check chunk quality and search relevance

## 📞 Support

If you encounter issues:
1. Check all services are running (Ollama, Qdrant)
2. Verify dependencies are installed
3. Run `test_rag_setup.bat` for diagnostics
4. Check logs in ai-service for detailed errors

---

**🎉 Your RAG system is now powered by Docling + Qdrant!**
