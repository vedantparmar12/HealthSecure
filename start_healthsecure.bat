@echo off
echo Starting HealthSecure Full Stack Application
echo =========================================

echo.
echo [0/4] Stopping any existing processes...
taskkill /f /im go.exe >nul 2>&1
taskkill /f /im python.exe >nul 2>&1
taskkill /f /im node.exe >nul 2>&1
taskkill /f /im qdrant.exe >nul 2>&1
REM Kill processes on specific ports
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8081') do taskkill /f /pid %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :5000') do taskkill /f /pid %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :3002') do taskkill /f /pid %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :6333') do taskkill /f /pid %%a >nul 2>&1
timeout /t 2 /nobreak >nul

echo.
echo [1/4] Starting Qdrant Vector Database on port 6333...
cd qdrant
if exist qdrant.exe (
    start "Qdrant Vector DB" cmd /k "qdrant.exe"
    echo Qdrant started from local executable
) else (
    echo WARNING: Qdrant executable not found in qdrant folder
    echo Please run setup_qdrant.bat first or start Qdrant manually
)
cd ..

echo.
echo [2/4] Starting Backend Go Server on port 8081...
cd backend
set SERVER_PORT=8081
start "HealthSecure Backend" cmd /k "set SERVER_PORT=8081 && go run cmd/server/main.go"

echo.
echo [3/4] Starting AI Service on port 5000...
cd ..\ai-service
set AI_SERVICE_PORT=5000
start "HealthSecure AI Service" cmd /k "set AI_SERVICE_PORT=5000 && python app.py"

echo.
echo [4/4] Starting Frontend on port 3002...
cd ..\frontend
start "HealthSecure Frontend" cmd /k "npm run dev -- --port 3002"

echo.
echo =========================================
echo HealthSecure Application Starting...
echo =========================================
echo.
echo Qdrant DB:  http://localhost:6333
echo Backend:    http://localhost:8081
echo AI Service: http://localhost:5000
echo Frontend:   http://localhost:3002
echo.
echo Login Credentials:
echo   Doctor: dr.smith@hospital.local / Doctor123
echo   Nurse:  nurse.jane@hospital.local / Nurse123
echo   Admin:  admin@hospital.local / admin123
echo.
echo Waiting for services to start...
timeout /t 8 /nobreak >nul
echo.
echo All services should now be running!
echo Press any key to close this window...
pause >nul