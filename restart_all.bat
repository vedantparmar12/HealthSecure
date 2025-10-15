@echo off
REM ============================================
REM HealthSecure - Restart All Services
REM ============================================

title HealthSecure - Restart All Services

echo.
echo ============================================
echo     HealthSecure - Restarting All Services
echo ============================================
echo.

REM Stop all services first
echo [INFO] Stopping all running services...
call stop_all.bat

echo.
echo [INFO] Waiting 3 seconds before restart...
timeout /t 3 /nobreak >nul
echo.

REM Start all services
echo [INFO] Starting all services...
call start_all.bat
