$ErrorActionPreference = "Stop"

function Import-MSVCEnvironment {
    if (Get-Command cl.exe -ErrorAction SilentlyContinue) { return }

    $vswhere = Join-Path ${env:ProgramFiles(x86)} "Microsoft Visual Studio\Installer\vswhere.exe"
    $vsPath = $null
    if (Test-Path $vswhere) {
        $raw = & $vswhere -latest -products * -property installationPath
        if ($raw) { $vsPath = ($raw | Select-Object -First 1).Trim() }
    }
    if (-not $vsPath) { throw "Visual Studio C++ build tools were not found." }

    $devCmd = Join-Path $vsPath "Common7\Tools\VsDevCmd.bat"
    if (-not (Test-Path $devCmd)) { throw "VsDevCmd.bat was not found." }

    $tmp = Join-Path $env:TEMP ("auronq-kepler-vsenv-" + [guid]::NewGuid().ToString("N") + ".cmd")
    try {
        @(
            "@echo off",
            ('call "{0}" -arch=amd64 -host_arch=amd64 >nul' -f $devCmd),
            "if errorlevel 1 exit /b %errorlevel%",
            "set"
        ) | Set-Content -Path $tmp -Encoding ASCII
        $envLines = & cmd.exe /d /c $tmp
        if ($LASTEXITCODE -ne 0) { throw "Failed to initialize MSVC environment." }
    } finally {
        Remove-Item $tmp -Force -ErrorAction SilentlyContinue
    }
    foreach ($line in $envLines) {
        $idx = $line.IndexOf("=")
        if ($idx -gt 0) {
            Set-Item -Path ("Env:" + $line.Substring(0,$idx)) -Value $line.Substring($idx+1)
        }
    }
    if (-not (Get-Command cl.exe -ErrorAction SilentlyContinue)) {
        throw "cl.exe was not found after Visual Studio environment setup."
    }
}

Import-MSVCEnvironment

$nvcc = $null
if ($env:CUDA_PATH) {
    $candidate = Join-Path $env:CUDA_PATH "bin\nvcc.exe"
    if (Test-Path $candidate) { $nvcc = $candidate }
}
if (-not $nvcc) {
    $cmd = Get-Command nvcc.exe -ErrorAction SilentlyContinue
    if ($cmd) { $nvcc = $cmd.Source }
}
if (-not $nvcc) { throw "nvcc.exe not found. Kepler build requires CUDA 11.8." }

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$src = Join-Path $here "aqm64_cuda.cu"
$out = Join-Path $here "auronq-aqm64-cuda-kepler.dll"

$supportedCode = @(& $nvcc --list-gpu-code 2>$null) -join " "
$supportedArch = @(& $nvcc --list-gpu-arch 2>$null) -join " "
$wanted = @("35","37")
$gencode = @()
$virtual = @()

foreach ($cc in $wanted) {
    if ($supportedCode -match "(^|\s)sm_$cc(\s|$)") {
        $gencode += @("-gencode", "arch=compute_$cc,code=sm_$cc")
    }
    if ($supportedArch -match "(^|\s)compute_$cc(\s|$)") {
        $virtual += [int]$cc
    }
}

if ($virtual.Count -gt 0) {
    $latest = ($virtual | Measure-Object -Maximum).Maximum
    $gencode += @("-gencode", "arch=compute_$latest,code=compute_$latest")
}
if ($gencode.Count -eq 0) {
    throw "This CUDA Toolkit cannot build maintained Kepler targets sm_35/sm_37. Use CUDA 11.8."
}

Write-Host "Building AuronQ Kepler CUDA backend..."
Write-Host "NVCC: $nvcc"
Write-Host "Kepler targets: $($gencode -join ' ')"

$args = @(
    "-O3",
    "-std=c++17",
    "-shared",
    "--cudart", "static",
    "-Xcompiler", "/O2 /MD"
) + $gencode + @(
    $src,
    "-o", $out
)

& $nvcc @args
if ($LASTEXITCODE -ne 0) { throw "Kepler CUDA build failed with exit code $LASTEXITCODE" }
if (-not (Test-Path $out)) { throw "Kepler CUDA DLL was not created: $out" }
Write-Host "Built: $out"
