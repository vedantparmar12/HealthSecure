@echo off
echo Stopping HealthSecure Application...
echo ====================================

echo.
echo Stopping Qdrant Vector Database...
taskkill /f /im qdrant.exe >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :6333') do taskkill /f /pid %%a >nul 2>&1

echo Stopping Backend Go Server...
taskkill /f /im go.exe >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :8081') do taskkill /f /pid %%a >nul 2>&1

echo Stopping AI Service Python...
taskkill /f /im python.exe >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :5000') do taskkill /f /pid %%a >nul 2>&1

echo Stopping Frontend Node.js...
taskkill /f /im node.exe >nul 2>&1
for /f "tokens=5" %%a in ('netstat -aon ^| findstr :3002') do taskkill /f /pid %%a >nul 2>&1

timeout /t 2 /nobreak >nul

echo.
echo ====================================
echo HealthSecure Application Stopped
echo ====================================
echo.
pause