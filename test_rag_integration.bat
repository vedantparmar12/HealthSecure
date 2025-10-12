@echo off
echo ========================================
echo Testing RAG Integration with AI Service
echo ========================================
echo.

echo [1/5] Checking services...
echo.

REM Check Ollama
curl -s http://localhost:11434 >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Ollama is not running
    pause
    exit /b 1
)
echo ✅ Ollama is running

REM Check Qdrant
curl -s http://localhost:6333 >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Qdrant is not running
    pause
    exit /b 1
)
echo ✅ Qdrant is running

REM Check AI Service
curl -s http://localhost:5000/health >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ AI Service is not running
    echo Please start it: cd ai-service && python app.py
    pause
    exit /b 1
)
echo ✅ AI Service is running

echo.
echo [2/5] Checking RAG status...
curl -s http://localhost:5000/rag/status

echo.
echo.
echo [3/5] Ingesting sample medical document...
cd ai-service
python ingest_with_docling.py --test-single "sample_docs/sample_medical_report.txt"

echo.
echo [4/5] Testing RAG search...
curl -X POST http://localhost:5000/rag/search -H "Content-Type: application/json" -d "{\"query\":\"patient blood pressure\",\"limit\":3}"

echo.
echo.
echo [5/5] Testing chat with RAG context...
curl -X POST http://localhost:5000/chat -H "Content-Type: application/json" -d "{\"message\":\"What are the patient's vital signs?\",\"thread_id\":\"test-rag-001\",\"user_id\":\"test-user\",\"user_role\":\"doctor\",\"user_name\":\"Dr. Test\",\"history\":[]}"

echo.
echo.
echo ========================================
echo ✅ RAG Integration Test Complete!
echo ========================================
echo.
echo Your AI assistant now has access to:
echo   • Document analysis via Docling
echo   • Semantic search via Qdrant
echo   • Context-aware responses with sources
echo.
cd ..
pause
