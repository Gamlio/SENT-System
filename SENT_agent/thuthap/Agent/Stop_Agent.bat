@echo off
color 0c
echo ===================================================
echo   DUNG HOAT DONG VA GO BO SENTINEL AGENT
echo ===================================================
echo.

echo [1] Dang huy tien trinh ngam...
taskkill /F /IM offline_agent.exe >nul 2>&1

echo [2] Dang xoa khoa tu dong khoi chay (Registry)...
reg delete "HKCU\Software\Microsoft\Windows\CurrentVersion\Run" /v "SENT_Agent_Service" /f >nul 2>&1

echo [3] Dang xoa cac file JSON con sot lai...
del /F /Q sent_data_*.json >nul 2>&1

echo.
echo HOAN TAT! [cite_start]Agent da bi go bo hoan toan khoi he thong. [cite: 8]
timeout /t 3 >nul