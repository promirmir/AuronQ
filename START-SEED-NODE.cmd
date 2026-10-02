@echo off
setlocal EnableExtensions
cd /d "%~dp0"

REM Bootstrap/public-node launcher.
REM Normal users should run START-AURONQ.cmd instead.
REM If Tailscale Funnel is available, this exposes the local AuronQ node
REM through the device's stable public *.ts.net HTTPS hostname and advertises
REM that endpoint to the AuronQ P2P network.

set "TS="
for /f "delims=" %%I in ('where tailscale.exe 2^>nul') do if not defined TS set "TS=%%I"
if not defined TS if exist "%ProgramFiles%\Tailscale\tailscale.exe" set "TS=%ProgramFiles%\Tailscale\tailscale.exe"
if not defined TS if exist "%ProgramFiles(x86)%\Tailscale\tailscale.exe" set "TS=%ProgramFiles(x86)%\Tailscale\tailscale.exe"

set "AURONQ_ADVERTISE="
set "TSHOST="

if defined TS (
  for /f "usebackq delims=" %%A in (`powershell -NoProfile -Command "$j = & '%TS%' status --json | ConvertFrom-Json; if ($j.Self.DNSName) { $j.Self.DNSName.TrimEnd('.') }"`) do set "TSHOST=%%A"
)

if defined TSHOST set "AURONQ_ADVERTISE=https://%TSHOST%"

if defined TS (
  echo Configuring Tailscale Funnel for AuronQ...
  "%TS%" funnel --bg --yes http://127.0.0.1:18444
  if errorlevel 1 (
    echo WARNING: Tailscale Funnel could not be enabled.
  )
) else (
  echo WARNING: tailscale.exe was not found. This node will run, but it will not be publicly advertised through Funnel.
)

if defined AURONQ_ADVERTISE (
  echo AuronQ public endpoint: %AURONQ_ADVERTISE%
) else (
  echo WARNING: no public HTTPS advertise endpoint detected.
)

call "%~dp0START-AURONQ.cmd"
endlocal
