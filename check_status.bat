@echo off
echo Checking HealthSecure Application Status...
echo ==========================================

echo.
echo [Backend] Checking http://localhost:8081/health
curl -s http://localhost:8081/health 2>nul || echo Backend: NOT RUNNING

echo.
echo.
echo [AI Service] Checking http://localhost:5001/health
curl -s http://localhost:5001/health 2>nul || echo AI Service: NOT RUNNING

echo.
echo.
echo [Frontend] Checking http://localhost:3002
curl -s http://localhost:3002 2>nul || echo Frontend: NOT RUNNING

echo.
echo.
echo ==========================================
echo Status Check Complete
echo ==========================================
echo.
pause