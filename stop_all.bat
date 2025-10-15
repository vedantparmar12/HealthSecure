@echo off
REM ============================================
REM HealthSecure - Stop All Services
REM ============================================
REM This script stops all running services:
REM 1. Qdrant Vector Database (Port 6333)
REM 2. Go Backend API (Port 8081)
REM 3. Python AI Service (Port 5000)
REM 4. Next.js Frontend (Port 3000)
REM ============================================

title HealthSecure - Stopping All Services

echo.
echo ============================================
echo     HealthSecure - Stopping All Services
echo ============================================
echo.

REM ============================================
REM Stop Frontend (Port 3000)
REM ============================================

echo [STEP 1/4] Stopping Next.js Frontend (Port 3000)...

for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000"') do (
    echo [INFO] Killing process ID: %%a
    taskkill /PID %%a /F >nul 2>&1
)

echo [OK] Frontend stopped
echo.

REM ============================================
REM Stop AI Service (Port 5000)
REM ============================================

echo [STEP 2/4] Stopping Python AI Service (Port 5000)...

for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5000"') do (
    echo [INFO] Killing process ID: %%a
    taskkill /PID %%a /F >nul 2>&1
)

echo [OK] AI Service stopped
echo.

REM ============================================
REM Stop Go Backend (Port 8081)
REM ============================================

echo [STEP 3/4] Stopping Go Backend (Port 8081)...

for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8081"') do (
    echo [INFO] Killing process ID: %%a
    taskkill /PID %%a /F >nul 2>&1
)

echo [OK] Go Backend stopped
echo.

REM ============================================
REM Stop Qdrant (Port 6333)
REM ============================================

echo [STEP 4/4] Stopping Qdrant (Port 6333)...

for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":6333"') do (
    echo [INFO] Killing process ID: %%a
    taskkill /PID %%a /F >nul 2>&1
)

REM Also try to kill qdrant.exe by name
taskkill /IM qdrant.exe /F >nul 2>&1

echo [OK] Qdrant stopped
echo.

REM ============================================
REM Additional Cleanup
REM ============================================

echo [INFO] Performing additional cleanup...

REM Kill any remaining node.exe processes (Frontend)
taskkill /IM node.exe /F >nul 2>&1

REM Kill any remaining python.exe processes (AI Service)
REM WARNING: This might kill other Python processes too
REM Uncomment the line below if you want to force kill all Python processes
REM taskkill /IM python.exe /F >nul 2>&1

REM Kill any remaining go.exe processes (Backend)
taskkill /IM go.exe /F >nul 2>&1

echo [OK] Cleanup completed
echo.

REM ============================================
REM Verify All Services Stopped
REM ============================================

echo [INFO] Verifying all services are stopped...
echo.

REM Check each port
set "ALL_STOPPED=1"

netstat -ano | findstr ":6333" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Port 6333 ^(Qdrant^) is still in use
    set "ALL_STOPPED=0"
) else (
    echo [OK] Port 6333 ^(Qdrant^) is free
)

netstat -ano | findstr ":8081" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Port 8081 ^(Go Backend^) is still in use
    set "ALL_STOPPED=0"
) else (
    echo [OK] Port 8081 ^(Go Backend^) is free
)

netstat -ano | findstr ":5000" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Port 5000 ^(AI Service^) is still in use
    set "ALL_STOPPED=0"
) else (
    echo [OK] Port 5000 ^(AI Service^) is free
)

netstat -ano | findstr ":3000" >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [WARNING] Port 3000 ^(Frontend^) is still in use
    set "ALL_STOPPED=0"
) else (
    echo [OK] Port 3000 ^(Frontend^) is free
)

echo.

if "%ALL_STOPPED%"=="1" (
    echo ============================================
    echo     All Services Stopped Successfully!
    echo ============================================
) else (
    echo ============================================
    echo     Some Services May Still Be Running
    echo ============================================
    echo.
    echo [INFO] Please close any remaining terminal windows manually
    echo [INFO] or restart your computer if issues persist
)

echo.
echo Press any key to exit...
pause >nul
