@echo off
setlocal
cd /d "%~dp0"

REM Operator launcher for the two official bootstrap nodes.
REM It starts the full AuronQ node and, when Tailscale is installed and Funnel is
REM authorized for this tailnet, exposes localhost:18444 at the machine's stable
REM public *.ts.net HTTPS hostname. Funnel is backgrounded by the Tailscale service;
REM no PowerShell/console window has to remain open.
call "%~dp0START-AURONQ.cmd"
timeout /t 8 /nobreak >nul

set "TS=%ProgramFiles%\Tailscale\tailscale.exe"
if not exist "%TS%" set "TS=%ProgramFiles(x86)%\Tailscale\tailscale.exe"
if exist "%TS%" (
  "%TS%" funnel --bg http://127.0.0.1:18444 >nul 2>&1
)
endlocal