@echo off
echo ========================================
echo Testing RAG Setup with Docling + Qdrant
echo ========================================
echo.

echo [1/4] Checking Ollama...
curl -s http://localhost:11434 >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Ollama is not running!
    echo Please start Ollama first
    pause
    exit /b 1
)
echo ✅ Ollama is running

echo.
echo [2/4] Checking Qdrant...
curl -s http://localhost:6333 >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Qdrant is not running!
    echo Run setup_qdrant.bat to install and start Qdrant
    pause
    exit /b 1
)
echo ✅ Qdrant is running

echo.
echo [3/4] Installing Python dependencies...
cd ai-service
pip install -q docling docling-core qdrant-client pymupdf
if %errorlevel% neq 0 (
    echo ❌ Failed to install dependencies
    echo Try manually: pip install docling docling-core qdrant-client
    pause
    exit /b 1
)
echo ✅ Dependencies installed

echo.
echo [4/4] Testing Docling RAG Analyzer...
python -c "import asyncio; from docling_rag_analyzer import DoclingRAGAnalyzer; analyzer = DoclingRAGAnalyzer(); print('✅ DoclingRAGAnalyzer initialized successfully'); print(f'Qdrant URL: {analyzer.qdrant_url}'); print(f'Collection: {analyzer.collection_name}'); print(f'Vector Size: {analyzer.vector_size}')"

if %errorlevel% neq 0 (
    echo.
    echo ❌ RAG setup test failed
    echo Check the error message above
    pause
    exit /b 1
)

echo.
echo ========================================
echo ✅ ALL CHECKS PASSED!
echo ========================================
echo.
echo Your RAG system is ready to use with:
echo   • Docling: Advanced PDF processing with OCR
echo   • Qdrant: Vector database for similarity search
echo   • Ollama: Local embeddings (mxbai-embed-large)
echo.
echo Next steps:
echo   1. Place PDF documents in a folder
echo   2. Run: python ingest_with_docling.py --test-single path\to\document.pdf
echo   3. Or batch process: python ingest_with_docling.py --documents path\to\folder
echo.
cd ..
pause
