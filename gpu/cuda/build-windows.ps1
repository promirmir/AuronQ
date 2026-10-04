$ErrorActionPreference = "Stop"

function Import-MSVCEnvironment {
    if (Get-Command cl.exe -ErrorAction SilentlyContinue) {
        return
    }

    $vswhere = Join-Path ${env:ProgramFiles(x86)} "Microsoft Visual Studio\Installer\vswhere.exe"
    if (-not (Test-Path $vswhere)) {
        throw "Visual Studio Installer/vswhere.exe not found. Install 'Desktop development with C++' in Visual Studio."
    }

    $vsPath = (& $vswhere -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath).Trim()
    if (-not $vsPath) {
        throw "MSVC C++ build tools were not found. Open Visual Studio Installer and add 'Desktop development with C++'."
    }

    $devCmd = Join-Path $vsPath "Common7\Tools\VsDevCmd.bat"
    if (-not (Test-Path $devCmd)) {
        throw "VsDevCmd.bat not found under $vsPath"
    }

    Write-Host "Loading MSVC environment from:"
    Write-Host "  $devCmd"

    $envLines = & cmd.exe /s /c ('""{0}" -arch=amd64 -host_arch=amd64 >nul && set"' -f $devCmd)
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to initialize Visual Studio C++ build environment."
    }
    foreach ($line in $envLines) {
        $idx = $line.IndexOf("=")
        if ($idx -gt 0) {
            $name = $line.Substring(0, $idx)
            $value = $line.Substring($idx + 1)
            Set-Item -Path "Env:$name" -Value $value
        }
    }

    if (-not (Get-Command cl.exe -ErrorAction SilentlyContinue)) {
        throw "cl.exe is still not available after loading the Visual Studio C++ environment."
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
if (-not $nvcc) {
    throw "nvcc.exe not found. Install NVIDIA CUDA Toolkit 13.x and reopen PowerShell."
}

$cl = (Get-Command cl.exe).Source
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$out = Join-Path $here "auronq-aqm64-cuda.dll"
$src = Join-Path $here "aqm64_cuda.cu"

Write-Host "MSVC: $cl"
Write-Host "NVCC: $nvcc"
Write-Host "Building AuronQ AQM64 CUDA backend..."

$args = @(
    "-O3",
    "-std=c++17",
    "-shared",
    "-Xcompiler", "/O2 /MD",
    "-gencode", "arch=compute_75,code=sm_75",
    "-gencode", "arch=compute_86,code=sm_86",
    "-gencode", "arch=compute_89,code=sm_89",
    "-gencode", "arch=compute_89,code=compute_89",
    $src,
    "-o", $out
)

& $nvcc @args
if ($LASTEXITCODE -ne 0) {
    throw "nvcc failed with exit code $LASTEXITCODE"
}

if (-not (Test-Path $out)) {
    throw "CUDA build reported success but DLL was not created: $out"
}

Write-Host "Built: $out"
Write-Host "Next: copy the DLL next to auronq-gpu-miner.exe and run:"
Write-Host "  .\auronq-gpu-miner.exe --self-test"
