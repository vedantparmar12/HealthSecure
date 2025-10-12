@echo off
echo Testing HealthSecure Backend Login
echo =====================================
echo.

echo [1/3] Testing backend health...
curl -s http://localhost:8081/health
echo.
echo.

echo [2/3] Getting default users...
curl -s http://localhost:8081/api/auth/defaults
echo.
echo.

echo [3/3] Testing doctor login...
curl -X POST http://localhost:8081/api/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"email\":\"dr.smith@hospital.local\",\"password\":\"Doctor123\"}"
echo.
echo.

echo =====================================
echo Test Complete
echo =====================================
pause
