@echo off
echo Stopping HealthSecure Application...
echo ====================================

echo.
echo Stopping Backend Go Server...
taskkill /f /im go.exe >nul 2>&1

echo Stopping AI Service Python...
taskkill /f /im python.exe >nul 2>&1

echo Stopping Frontend Node.js...
taskkill /f /im node.exe >nul 2>&1

echo.
echo ====================================
echo HealthSecure Application Stopped
echo ====================================
echo.
pause