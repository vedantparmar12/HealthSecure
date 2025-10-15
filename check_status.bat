@echo off
REM ============================================
REM HealthSecure - Check Service Status
REM ============================================

title HealthSecure - Service Status Check

echo.
echo ============================================
echo     HealthSecure - Service Status
echo ============================================
echo.
echo Checking all services...
echo.

REM ============================================
REM Check Qdrant (Port 6333)
REM ============================================

echo [1] Qdrant Vector Database (Port 6333):
netstat -ano | findstr ":6333" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Status: RUNNING
    echo    URL: http://localhost:6333/dashboard
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":6333"') do (
        echo    PID: %%a
        goto :qdrant_done
    )
    :qdrant_done
) else (
    echo    Status: STOPPED
)
echo.

REM ============================================
REM Check Go Backend (Port 8081)
REM ============================================

echo [2] Go Backend API (Port 8081):
netstat -ano | findstr ":8081" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Status: RUNNING
    echo    URL: http://localhost:8081
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8081"') do (
        echo    PID: %%a
        goto :backend_done
    )
    :backend_done
) else (
    echo    Status: STOPPED
)
echo.

REM ============================================
REM Check AI Service (Port 5000)
REM ============================================

echo [3] Python AI Service (Port 5000):
netstat -ano | findstr ":5000" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Status: RUNNING
    echo    URL: http://localhost:5000
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5000"') do (
        echo    PID: %%a
        goto :ai_done
    )
    :ai_done
) else (
    echo    Status: STOPPED
)
echo.

REM ============================================
REM Check Frontend (Port 3000)
REM ============================================

echo [4] Next.js Frontend (Port 3000):
netstat -ano | findstr ":3000" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Status: RUNNING
    echo    URL: http://localhost:3000
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000"') do (
        echo    PID: %%a
        goto :frontend_done
    )
    :frontend_done
) else (
    echo    Status: STOPPED
)
echo.

REM ============================================
REM Test Connectivity
REM ============================================

echo ============================================
echo Testing Service Connectivity...
echo ============================================
echo.

echo [1] Testing Qdrant...
curl -s http://localhost:6333 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Response: OK
) else (
    echo    Response: NOT REACHABLE
)

echo.
echo [2] Testing Go Backend...
curl -s http://localhost:8081 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Response: OK
) else (
    echo    Response: NOT REACHABLE
)

echo.
echo [3] Testing AI Service...
curl -s http://localhost:5000 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Response: OK
) else (
    echo    Response: NOT REACHABLE
)

echo.
echo [4] Testing Frontend...
curl -s http://localhost:3000 >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo    Response: OK
) else (
    echo    Response: NOT REACHABLE
)

echo.
echo ============================================
echo     Status Check Complete
echo ============================================
echo.

REM Check if all services are running
set "ALL_RUNNING=1"
netstat -ano | findstr ":6333" >nul 2>&1
if %ERRORLEVEL% NEQ 0 set "ALL_RUNNING=0"

netstat -ano | findstr ":8081" >nul 2>&1
if %ERRORLEVEL% NEQ 0 set "ALL_RUNNING=0"

netstat -ano | findstr ":5000" >nul 2>&1
if %ERRORLEVEL% NEQ 0 set "ALL_RUNNING=0"

netstat -ano | findstr ":3000" >nul 2>&1
if %ERRORLEVEL% NEQ 0 set "ALL_RUNNING=0"

if "%ALL_RUNNING%"=="1" (
    echo [OK] All services are running!
    echo.
    echo Quick Links:
    echo   - Frontend:  http://localhost:3000
    echo   - Dashboard: http://localhost:3000/dashboard
    echo   - Login:     http://localhost:3000/login
) else (
    echo [WARNING] Some services are not running
    echo.
    echo To start all services, run: start_all.bat
)

echo.
echo Press any key to exit...
pause >nul
