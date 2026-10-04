$ErrorActionPreference = "Stop"

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
    throw "nvcc.exe not found. Install NVIDIA CUDA Toolkit 12.x and reopen PowerShell."
}

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$out = Join-Path $here "auronq-aqm64-cuda.dll"
$src = Join-Path $here "aqm64_cuda.cu"

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

Write-Host "Building AuronQ AQM64 CUDA backend..."
& $nvcc @args
if ($LASTEXITCODE -ne 0) {
    throw "nvcc failed with exit code $LASTEXITCODE"
}

Write-Host "Built: $out"
Write-Host "Next: copy the DLL next to auronq-gpu-miner.exe and run:"
Write-Host "  .\auronq-gpu-miner.exe --self-test"
