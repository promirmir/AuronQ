$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

function Need-Command([string]$name, [string]$hint) {
    $cmd = Get-Command $name -ErrorAction SilentlyContinue
    if (-not $cmd) {
        throw "$name not found. $hint"
    }
    return $cmd.Source
}

Write-Host ""
Write-Host "AuronQ GPU Miner - Windows prototype build"
Write-Host "==========================================="
Write-Host ""

$go = Need-Command "go.exe" "Install Go and reopen PowerShell."
$nvcc = Need-Command "nvcc.exe" "Install NVIDIA CUDA Toolkit 13.x and reopen PowerShell."

Write-Host "Go:   $go"
Write-Host "NVCC: $nvcc"
& $go version
& $nvcc --version | Select-Object -Last 4

$dist = Join-Path $root "dist\AuronQ-GPU-Miner-v0.1-prototype"
if (Test-Path $dist) {
    Remove-Item -Recurse -Force $dist
}
New-Item -ItemType Directory -Force -Path $dist | Out-Null

Write-Host ""
Write-Host "[1/3] Building CUDA backend..."
powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $root "gpu\cuda\build-windows.ps1")
if ($LASTEXITCODE -ne 0) {
    throw "CUDA backend build failed with exit code $LASTEXITCODE"
}

Write-Host ""
Write-Host "[2/3] Building Go miner..."
& $go build -trimpath -o (Join-Path $dist "auronq-gpu-miner.exe") .\cmd\auronq-gpu-miner
if ($LASTEXITCODE -ne 0) {
    throw "Go build failed with exit code $LASTEXITCODE"
}

Copy-Item (Join-Path $root "gpu\cuda\auronq-aqm64-cuda.dll") (Join-Path $dist "auronq-aqm64-cuda.dll") -Force
Copy-Item (Join-Path $root "gpu\cuda\README.md") (Join-Path $dist "README-GPU-MINER.md") -Force

Write-Host ""
Write-Host "[3/3] Running mandatory GPU/CPU AQM64 equivalence self-test..."
Push-Location $dist
try {
    .\auronq-gpu-miner.exe --self-test
    if ($LASTEXITCODE -ne 0) {
        throw "GPU/CPU AQM64 self-test failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "SUCCESS"
Write-Host "Built and self-tested:"
Write-Host "  $dist"
Write-Host ""
Write-Host "Do not start Mainnet mining unless the self-test printed SELF-TEST OK."
