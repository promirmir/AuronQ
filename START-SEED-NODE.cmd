@echo off
setlocal
cd /d "%~dp0"

REM Bootstrap-operator launcher.
REM Normal users should run START-AURONQ.cmd instead.
REM When Tailscale is installed and Funnel is authorized, this exposes the local
REM AuronQ full node at the machine's stable public *.ts.net HTTPS hostname.
REM Funnel runs through the Tailscale service, so no console must stay open.
call "%~dp0START-AURONQ.cmd"
timeout /t 8 /nobreak >nul

set "TS=%ProgramFiles%\Tailscale\tailscale.exe"
if not exist "%TS%" set "TS=%ProgramFiles(x86)%\Tailscale\tailscale.exe"
if exist "%TS%" (
  "%TS%" funnel --bg http://127.0.0.1:18444 >nul 2>&1
)
endlocal
