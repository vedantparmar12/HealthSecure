@echo off
echo Starting HealthSecure Full Stack Application
echo =========================================

echo.
echo [0/3] Stopping any existing processes...
taskkill /f /im go.exe >nul 2>&1
taskkill /f /im python.exe >nul 2>&1
taskkill /f /im node.exe >nul 2>&1
timeout /t 2 /nobreak >nul

echo.
echo [1/3] Starting Backend Go Server on port 8081...
cd backend
set SERVER_PORT=8081
start "HealthSecure Backend" cmd /k "set SERVER_PORT=8081 && go run cmd/server/main.go"

echo.
echo [2/3] Starting AI Service on port 5001...
cd ..\ai-service
set AI_SERVICE_PORT=5001
start "HealthSecure AI Service" cmd /k "set AI_SERVICE_PORT=5001 && python app.py"

echo.
echo [3/3] Starting Frontend on port 3002...
cd ..\frontend
start "HealthSecure Frontend" cmd /k "npm run dev -- --port 3002"

echo.
echo =========================================
echo HealthSecure Application Starting...
echo =========================================
echo.
echo Backend:    http://localhost:8081
echo AI Service: http://localhost:5001
echo Frontend:   http://localhost:3002
echo.
echo Waiting for services to start...
timeout /t 5 /nobreak >nul
echo.
echo Press any key to close this window...
pause >nul