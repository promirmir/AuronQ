@echo off
setlocal
cd /d "%~dp0"
call "%~dp0START-AURONQ.cmd"
timeout /t 8 /nobreak >nul
set "TS=%ProgramFiles%\Tailscale\tailscale.exe"
if not exist "%TS%" set "TS=%ProgramFiles(x86)%\Tailscale\tailscale.exe"
if exist "%TS%" (
  "%TS%" funnel --bg http://127.0.0.1:18444
)
endlocal
