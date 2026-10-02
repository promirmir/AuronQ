@echo off
setlocal
cd /d "%~dp0"

REM Normal AuronQ user launcher. The node opens outbound connections automatically.
REM The firewall rule only permits inbound TCP/18444 from the local subnet; public
REM reachability is optional and not required for a normal node.
netsh advfirewall firewall show rule name="AuronQ Mainnet TCP 18444 LAN" >nul 2>&1
if errorlevel 1 (
  powershell -NoProfile -ExecutionPolicy Bypass -Command "Start-Process cmd -Verb RunAs -ArgumentList '/c netsh advfirewall firewall add rule name=\"AuronQ Mainnet TCP 18444 LAN\" dir=in action=allow protocol=TCP localport=18444 remoteip=LocalSubnet profile=any' -Wait" >nul 2>&1
)
start "" "%~dp0AuronQ-Desktop.exe"
endlocal