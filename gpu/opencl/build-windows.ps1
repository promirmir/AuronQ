$ErrorActionPreference = "Stop"

function Import-MSVCEnvironment {
    if (Get-Command cl.exe -ErrorAction SilentlyContinue) { return }

    $vswhere = Join-Path ${env:ProgramFiles(x86)} "Microsoft Visual Studio\Installer\vswhere.exe"
    $vsPath = $null
    if (Test-Path $vswhere) {
        $raw = & $vswhere -latest -products * -property installationPath
        if ($raw) { $vsPath = ($raw | Select-Object -First 1).Trim() }
    }
    if (-not $vsPath) {
        foreach ($root in @("C:\Program Files\Microsoft Visual Studio","C:\Program Files (x86)\Microsoft Visual Studio")) {
            if (-not (Test-Path $root)) { continue }
            $candidate = Get-ChildItem $root -Filter VsDevCmd.bat -File -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
            if ($candidate) {
                $vsPath = Split-Path -Parent (Split-Path -Parent $candidate.FullName)
                break
            }
        }
    }
    if (-not $vsPath) { throw "Visual Studio installation was not found. Install Desktop development with C++." }

    $devCmd = Join-Path $vsPath "Common7\Tools\VsDevCmd.bat"
    if (-not (Test-Path $devCmd)) {
        $candidate = Get-ChildItem $vsPath -Filter VsDevCmd.bat -File -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($candidate) { $devCmd = $candidate.FullName }
    }
    if (-not (Test-Path $devCmd)) { throw "VsDevCmd.bat not found." }

    $tmp = Join-Path $env:TEMP ("auronq-opencl-vsenv-" + [guid]::NewGuid().ToString("N") + ".cmd")
    try {
        @("@echo off", ('call "{0}" -arch=amd64 -host_arch=amd64 >nul' -f $devCmd), "if errorlevel 1 exit /b %errorlevel%", "set") |
            Set-Content -Path $tmp -Encoding ASCII
        $envLines = & cmd.exe /d /c $tmp
        if ($LASTEXITCODE -ne 0) { throw "Failed to initialize Visual Studio C++ environment." }
    } finally {
        Remove-Item $tmp -Force -ErrorAction SilentlyContinue
    }
    foreach ($line in $envLines) {
        $idx = $line.IndexOf("=")
        if ($idx -gt 0) { Set-Item -Path ("Env:" + $line.Substring(0,$idx)) -Value $line.Substring($idx+1) }
    }
    if (-not (Get-Command cl.exe -ErrorAction SilentlyContinue)) { throw "cl.exe not found after Visual Studio environment setup." }
}

Import-MSVCEnvironment

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$src = Join-Path $here "aqm64_opencl.cpp"
$out = Join-Path $here "auronq-aqm64-opencl.dll"

Write-Host "Building AuronQ vendor-neutral OpenCL backend..."
& cl.exe /nologo /O2 /std:c++17 /EHsc /LD $src /Fe:$out
if ($LASTEXITCODE -ne 0) { throw "OpenCL backend build failed with exit code $LASTEXITCODE" }
if (-not (Test-Path $out)) { throw "OpenCL DLL was not created: $out" }
Write-Host "Built: $out"
