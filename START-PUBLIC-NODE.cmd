@echo off
setlocal
cd /d "%~dp0"

echo AuronQ Public Node
echo.
echo This mode opens TCP/18444 in Windows Firewall so other AuronQ nodes can
echo reach this computer if your Internet connection/router also permits it.
echo If you are behind a router, forward TCP 18444 to this PC.
echo CGNAT users normally need a VPS/public relay instead.
echo.
pause

netsh advfirewall firewall show rule name="AuronQ Mainnet TCP 18444 Public" >nul 2>&1
if errorlevel 1 (
  powershell -NoProfile -ExecutionPolicy Bypass -Command "Start-Process cmd -Verb RunAs -ArgumentList '/c netsh advfirewall firewall add rule name="AuronQ Mainnet TCP 18444 Public" dir=in action=allow protocol=TCP localport=18444 profile=any' -Wait"
)

start "" "AuronQ-Desktop.exe"
exit /b 0
