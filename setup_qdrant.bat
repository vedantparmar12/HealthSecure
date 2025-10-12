@echo off
echo ========================================
echo Setting up Qdrant Vector Database
echo ========================================
echo.

REM Check if Qdrant is already running
curl -s http://localhost:6333 >nul 2>&1
if %errorlevel% equ 0 (
    echo Qdrant is already running on port 6333
    echo.
    goto :menu
)

echo [1/3] Downloading Qdrant standalone binary...
echo.

REM Create qdrant directory if it doesn't exist
if not exist "qdrant" mkdir qdrant
cd qdrant

REM Download Qdrant for Windows
echo Downloading Qdrant v1.7.4 for Windows...
curl -L https://github.com/qdrant/qdrant/releases/download/v1.7.4/qdrant-x86_64-pc-windows-msvc.zip -o qdrant.zip

if not exist "qdrant.zip" (
    echo ERROR: Failed to download Qdrant
    echo.
    echo Please download manually from:
    echo https://github.com/qdrant/qdrant/releases
    pause
    exit /b 1
)

echo.
echo [2/3] Extracting Qdrant...
tar -xf qdrant.zip
if %errorlevel% neq 0 (
    echo ERROR: Failed to extract Qdrant
    echo Please extract qdrant.zip manually
    pause
    exit /b 1
)

echo.
echo [3/3] Starting Qdrant server...
echo.

REM Start Qdrant in a new window
start "Qdrant Vector Database" cmd /k "qdrant.exe"

echo Waiting for Qdrant to start...
timeout /t 5 /nobreak >nul

REM Verify Qdrant is running
curl -s http://localhost:6333 >nul 2>&1
if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo SUCCESS! Qdrant is now running
    echo ========================================
    echo.
    echo Qdrant Dashboard: http://localhost:6333/dashboard
    echo Qdrant API: http://localhost:6333
    echo.
) else (
    echo.
    echo WARNING: Qdrant may not have started properly
    echo Please check the Qdrant window for errors
    echo.
)

cd ..

:menu
echo ========================================
echo Next Steps:
echo ========================================
echo 1. Install Python dependencies:
echo    cd ai-service
echo    pip install -r requirements.txt
echo.
echo 2. Run the test:
echo    python -c "import asyncio; from docling_rag_analyzer import DoclingRAGAnalyzer; asyncio.run(DoclingRAGAnalyzer().search_similar_chunks('test'))"
echo.
echo 3. Ingest documents:
echo    python ingest_with_docling.py --documents ./sample_docs
echo.
pause
