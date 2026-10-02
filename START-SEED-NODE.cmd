@echo off
setlocal
cd /d "%~dp0"

REM Bootstrap/public-node launcher.
REM Normal users should run START-AURONQ.cmd instead.
REM When Tailscale is installed and Funnel is authorized, this exposes the local
REM AuronQ full node at the machine's stable public *.ts.net HTTPS hostname and
REM advertises that HTTPS endpoint to other AuronQ nodes.

set "TS=%ProgramFiles%Tailscale	ailscale.exe"
if not exist "%TS%" set "TS=%ProgramFiles(x86)%Tailscale	ailscale.exe"

set "AURONQ_ADVERTISE="
if exist "%TS%" (
  for /f "usebackq delims=" %%A in (`powershell -NoProfile -Command "$j = & '%TS%' status --json | ConvertFrom-Json; if ($j.Self.DNSName) { $j.Self.DNSName.TrimEnd('.') }"`) do set "TSHOST=%%A"
  if defined TSHOST set "AURONQ_ADVERTISE=https://%TSHOST%"
)

call "%~dp0START-AURONQ.cmd"
timeout /t 8 /nobreak >nul

if exist "%TS%" (
  "%TS%" funnel --bg http://127.0.0.1:18444 >nul 2>&1
)

if defined AURONQ_ADVERTISE (
  echo AuronQ public endpoint: %AURONQ_ADVERTISE%
) else (
  echo WARNING: no public HTTPS advertise endpoint detected.
)
endlocal
