@echo off
REM ============================================
REM HealthSecure - Start All Services
REM ============================================
REM This script starts all required services:
REM 1. Qdrant Vector Database (Port 6333)
REM 2. Go Backend API (Port 8081)
REM 3. Python AI Service (Port 5000)
REM 4. Next.js Frontend (Port 3000)
REM ============================================

title HealthSecure - Starting All Services

echo.
echo ============================================
echo     HealthSecure - Complete Startup
echo ============================================
echo.

REM Get the directory where this script is located
set "ROOT_DIR=%~dp0"
cd /d "%ROOT_DIR%"

echo [INFO] Root directory: %ROOT_DIR%
echo.

REM ============================================
REM Check if services are already running
REM ============================================

echo [STEP 1/5] Checking for running services...
echo.

REM Check Qdrant (Port 6333)
netstat -ano | findstr ":6333" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Qdrant appears to be already running on port 6333
) else (
    echo [OK] Port 6333 is available for Qdrant
)

REM Check Go Backend (Port 8081)
netstat -ano | findstr ":8081" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Go Backend appears to be already running on port 8081
) else (
    echo [OK] Port 8081 is available for Go Backend
)

REM Check AI Service (Port 5000)
netstat -ano | findstr ":5000" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] AI Service appears to be already running on port 5000
) else (
    echo [OK] Port 5000 is available for AI Service
)

REM Check Frontend (Port 3000)
netstat -ano | findstr ":3000" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Frontend appears to be already running on port 3000
) else (
    echo [OK] Port 3000 is available for Frontend
)

echo.
timeout /t 2 /nobreak >nul

REM ============================================
REM Start Qdrant Vector Database
REM ============================================

echo [STEP 2/5] Starting Qdrant Vector Database...
echo.

if exist "%ROOT_DIR%qdrant\qdrant.exe" (
    echo [INFO] Starting Qdrant on port 6333...
    start "HealthSecure - Qdrant" /D "%ROOT_DIR%qdrant" cmd /k "qdrant.exe"

    REM Wait for Qdrant to be ready (max 30 seconds)
    echo [INFO] Waiting for Qdrant to start...
    set /a QDRANT_RETRY=0
    :WAIT_QDRANT
    curl -s http://localhost:6333 >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Qdrant started successfully
        goto :QDRANT_READY
    )
    set /a QDRANT_RETRY+=1
    if %QDRANT_RETRY% GEQ 30 (
        echo [WARNING] Qdrant may not be ready yet, continuing anyway...
        goto :QDRANT_READY
    )
    ping 127.0.0.1 -n 2 >nul
    goto :WAIT_QDRANT
    :QDRANT_READY
) else (
    echo [ERROR] Qdrant executable not found at: %ROOT_DIR%qdrant\qdrant.exe
    echo [INFO] Please run setup_qdrant.bat first
    pause
    exit /b 1
)

echo.

REM ============================================
REM Start Go Backend API
REM ============================================

echo [STEP 3/5] Starting Go Backend API...
echo.

REM Check for compiled executable first (faster)
if exist "%ROOT_DIR%backend\main.exe" (
    echo [INFO] Starting Go Backend from compiled executable on port 8081...
    start "HealthSecure - Go Backend" /D "%ROOT_DIR%backend" cmd /k "main.exe"

    REM Wait for Backend to be ready (max 120 seconds for migrations)
    echo [INFO] Waiting for Go Backend to start (this may take up to 2 minutes for database migrations)...
    set /a BACKEND_RETRY=0
    :WAIT_BACKEND
    curl -s http://localhost:8081/health >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Go Backend started successfully and is healthy
        goto :BACKEND_READY
    )
    set /a BACKEND_RETRY+=1
    if %BACKEND_RETRY% GEQ 120 (
        echo [WARNING] Backend may not be ready yet, continuing anyway...
        echo [INFO] Check the "HealthSecure - Go Backend" window for details
        goto :BACKEND_READY
    )
    if %BACKEND_RETRY% EQU 30 (
        echo [INFO] Still waiting... Backend is likely running database migrations
    )
    if %BACKEND_RETRY% EQU 60 (
        echo [INFO] Still waiting... This is normal for first-time startup
    )
    ping 127.0.0.1 -n 2 >nul
    goto :WAIT_BACKEND
    :BACKEND_READY
) else if exist "%ROOT_DIR%backend\cmd\server\main.go" (
    echo [INFO] Starting Go Backend from source on port 8081...
    start "HealthSecure - Go Backend" /D "%ROOT_DIR%backend" cmd /k "go run cmd/server/main.go"

    REM Wait for Backend to be ready (max 120 seconds for migrations)
    echo [INFO] Waiting for Go Backend to start (this may take up to 2 minutes for database migrations)...
    set /a BACKEND_RETRY=0
    :WAIT_BACKEND_SOURCE
    curl -s http://localhost:8081/health >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Go Backend started successfully and is healthy
        goto :BACKEND_SOURCE_READY
    )
    set /a BACKEND_RETRY+=1
    if %BACKEND_RETRY% GEQ 120 (
        echo [WARNING] Backend may not be ready yet, continuing anyway...
        echo [INFO] Check the "HealthSecure - Go Backend" window for details
        goto :BACKEND_SOURCE_READY
    )
    if %BACKEND_RETRY% EQU 30 (
        echo [INFO] Still waiting... Backend is likely running database migrations
    )
    if %BACKEND_RETRY% EQU 60 (
        echo [INFO] Still waiting... This is normal for first-time startup
    )
    ping 127.0.0.1 -n 2 >nul
    goto :WAIT_BACKEND_SOURCE
    :BACKEND_SOURCE_READY
) else (
    echo [ERROR] Go Backend not found
    echo [INFO] Looking for: %ROOT_DIR%backend\main.exe OR %ROOT_DIR%backend\cmd\server\main.go
    pause
    exit /b 1
)

echo.

REM ============================================
REM Start Python AI Service
REM ============================================

echo [STEP 4/5] Starting Python AI Service...
echo.

if exist "%ROOT_DIR%ai-service\app.py" (
    echo [INFO] Checking for Python virtual environment...

    REM Try to use virtual environment if it exists
    if exist "%ROOT_DIR%ai-service\.venv\Scripts\activate.bat" (
        echo [INFO] Using virtual environment (.venv)
        start "HealthSecure - AI Service" /D "%ROOT_DIR%ai-service" cmd /k ".venv\Scripts\activate.bat && python app.py"
    ) else if exist "%ROOT_DIR%ai-service\venv\Scripts\activate.bat" (
        echo [INFO] Using virtual environment (venv)
        start "HealthSecure - AI Service" /D "%ROOT_DIR%ai-service" cmd /k "venv\Scripts\activate.bat && python app.py"
    ) else (
        echo [WARNING] No virtual environment found, using system Python
        start "HealthSecure - AI Service" /D "%ROOT_DIR%ai-service" cmd /k "python app.py"
    )

    REM Wait for AI Service to be ready (max 60 seconds)
    echo [INFO] Waiting for AI Service to start...
    set /a AI_RETRY=0
    :WAIT_AI
    curl -s http://localhost:5000 >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [OK] AI Service started successfully
        goto :AI_READY
    )
    set /a AI_RETRY+=1
    if %AI_RETRY% GEQ 60 (
        echo [WARNING] AI Service may not be ready yet, continuing anyway...
        goto :AI_READY
    )
    ping 127.0.0.1 -n 2 >nul
    goto :WAIT_AI
    :AI_READY
) else (
    echo [ERROR] AI Service not found at: %ROOT_DIR%ai-service\app.py
    pause
    exit /b 1
)

echo.

REM ============================================
REM Start Next.js Frontend
REM ============================================

echo [STEP 5/5] Starting Next.js Frontend...
echo.

if exist "%ROOT_DIR%frontend\package.json" (
    echo [INFO] Starting Next.js Frontend on port 3000...
    start "HealthSecure - Frontend" /D "%ROOT_DIR%frontend" cmd /k "npm run dev"

    REM Wait for Frontend to be ready (max 60 seconds)
    echo [INFO] Waiting for Frontend to start...
    set /a FRONTEND_RETRY=0
    :WAIT_FRONTEND
    curl -s http://localhost:3000 >nul 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Frontend started successfully
        goto :FRONTEND_READY
    )
    set /a FRONTEND_RETRY+=1
    if %FRONTEND_RETRY% GEQ 60 (
        echo [WARNING] Frontend may not be ready yet, continuing anyway...
        goto :FRONTEND_READY
    )
    ping 127.0.0.1 -n 2 >nul
    goto :WAIT_FRONTEND
    :FRONTEND_READY
) else (
    echo [ERROR] Frontend not found at: %ROOT_DIR%frontend\package.json
    pause
    exit /b 1
)

echo.

REM ============================================
REM Startup Complete
REM ============================================

echo ============================================
echo     All Services Started Successfully!
echo ============================================
echo.
echo Service URLs:
echo.
echo   [1] Qdrant Dashboard:  http://localhost:6333/dashboard
echo   [2] Go Backend API:    http://localhost:8081
echo   [3] AI Service API:    http://localhost:5000
echo   [4] Frontend Website:  http://localhost:3000
echo.
echo ============================================
echo.
echo [INFO] Four terminal windows have been opened:
echo   - HealthSecure - Qdrant
echo   - HealthSecure - Go Backend
echo   - HealthSecure - AI Service
echo   - HealthSecure - Frontend
echo.
echo [INFO] To stop all services, close each terminal window
echo        or run: stop_all.bat
echo.
echo [INFO] Opening frontend in browser in 5 seconds...
echo.

ping 127.0.0.1 -n 6 >nul

REM Open the frontend in default browser
start http://localhost:3000

echo ============================================
echo     HealthSecure is Ready to Use!
echo ============================================
echo.
echo Press any key to close this window...
echo (Services will continue running in separate windows)
pause >nul
