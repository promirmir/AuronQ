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

$version = "0.4.6-alpha"

Write-Host ""
Write-Host "AuronQ Universal Miner v$version - Windows build"
Write-Host "================================================"
Write-Host ""

$go = Need-Command "go.exe" "Install Go and reopen PowerShell."
$nvccCmd = Get-Command "nvcc.exe" -ErrorAction SilentlyContinue

Write-Host "Go:   $go"
& $go version

$cudaBuilt = $false
$legacyCudaBuilt = $false
$keplerCudaBuilt = $false
$openclBuilt = $false
if ($nvccCmd) {
    Write-Host "NVCC: $($nvccCmd.Source)"
    & $nvccCmd.Source --version | Select-Object -Last 4
    Write-Host ""
    Write-Host "[1/5] Building optional NVIDIA CUDA backend..."
    powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $root "gpu\cuda\build-windows.ps1")
    if ($LASTEXITCODE -ne 0) {
        throw "CUDA backend build failed with exit code $LASTEXITCODE"
    }
    $cudaBuilt = Test-Path (Join-Path $root "gpu\cuda\auronq-aqm64-cuda.dll")

    $supportedLegacy = @(& $nvccCmd.Source --list-gpu-code 2>$null) -join " "
    if ($supportedLegacy -match "(^|\s)sm_61(\s|$)") {
        Write-Host ""
        Write-Host "[1a/5] Building optional legacy Maxwell/Pascal/Volta CUDA backend..."
        powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $root "gpu\cuda\build-windows-legacy.ps1")
        if ($LASTEXITCODE -ne 0) { throw "Legacy CUDA backend build failed with exit code $LASTEXITCODE" }
        $legacyCudaBuilt = Test-Path (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-legacy.dll")
    } else {
        Write-Host "Current NVCC does not emit sm_61; legacy CUDA DLL will not be rebuilt locally."
        $legacyCudaBuilt = Test-Path (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-legacy.dll")
    }

    if ($supportedLegacy -match "(^|\s)sm_35(\s|$)") {
        Write-Host ""
        Write-Host "[1aa/5] Building optional Kepler CUDA backend..."
        powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $root "gpu\cuda\build-windows-kepler.ps1")
        if ($LASTEXITCODE -ne 0) { throw "Kepler CUDA backend build failed with exit code $LASTEXITCODE" }
        $keplerCudaBuilt = Test-Path (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-kepler.dll")
    } else {
        Write-Host "Current NVCC does not emit sm_35; Kepler CUDA DLL will not be rebuilt locally."
        $keplerCudaBuilt = Test-Path (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-kepler.dll")
    }
} else {
    Write-Host "NVCC: not found - CUDA acceleration will be skipped."
    $legacyCudaBuilt = Test-Path (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-legacy.dll")
    $keplerCudaBuilt = Test-Path (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-kepler.dll")
}

Write-Host ""
Write-Host "[1b/5] Building optional vendor-neutral OpenCL backend..."
try {
    powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $root "gpu\opencl\build-windows.ps1")
    if ($LASTEXITCODE -eq 0) {
        $openclBuilt = Test-Path (Join-Path $root "gpu\opencl\auronq-aqm64-opencl.dll")
    }
} catch {
    Write-Warning ("OpenCL backend build skipped: " + $_.Exception.Message)
}

$dist = Join-Path $root "dist\AuronQ-Miner-v$version-Windows-x64"
if (Test-Path $dist) {
    Remove-Item -Recurse -Force $dist
}
New-Item -ItemType Directory -Force -Path $dist | Out-Null

Write-Host ""
Write-Host "[2/5] Building universal miner worker..."
& $go build -trimpath -o (Join-Path $dist "auronq-miner-worker.exe") .\cmd\auronq-gpu-miner
if ($LASTEXITCODE -ne 0) {
    throw "Universal miner worker build failed with exit code $LASTEXITCODE"
}

Write-Host ""
Write-Host "[3/5] Building Windows GUI and full-node CLI..."
& $go build -trimpath -ldflags "-H=windowsgui" -o (Join-Path $dist "AuronQ-Miner.exe") .\cmd\auronq-gpu-miner-gui
if ($LASTEXITCODE -ne 0) {
    throw "GUI build failed with exit code $LASTEXITCODE"
}
& $go build -trimpath -o (Join-Path $dist "auronq.exe") .\cmd\auronq
if ($LASTEXITCODE -ne 0) {
    throw "Full-node CLI build failed with exit code $LASTEXITCODE"
}

if ($cudaBuilt) {
    Copy-Item (Join-Path $root "gpu\cuda\auronq-aqm64-cuda.dll") (Join-Path $dist "auronq-aqm64-cuda.dll") -Force
}
if ($legacyCudaBuilt) {
    Copy-Item (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-legacy.dll") (Join-Path $dist "auronq-aqm64-cuda-legacy.dll") -Force
}
if ($keplerCudaBuilt) {
    Copy-Item (Join-Path $root "gpu\cuda\auronq-aqm64-cuda-kepler.dll") (Join-Path $dist "auronq-aqm64-cuda-kepler.dll") -Force
}
if ($openclBuilt) {
    Copy-Item (Join-Path $root "gpu\opencl\auronq-aqm64-opencl.dll") (Join-Path $dist "auronq-aqm64-opencl.dll") -Force
}
Copy-Item (Join-Path $root "network.json") (Join-Path $dist "network.json") -Force
Copy-Item (Join-Path $root "bootstrap.json") (Join-Path $dist "bootstrap.json") -Force
Copy-Item (Join-Path $root "UNIVERSAL-MINER-GUIDE.md") (Join-Path $dist "UNIVERSAL-MINER-GUIDE.md") -Force
Copy-Item (Join-Path $root "GPU-MINER-GUIDE.md") (Join-Path $dist "GPU-MINER-GUIDE.md") -Force
if (Test-Path (Join-Path $root "LICENSE")) {
    Copy-Item (Join-Path $root "LICENSE") (Join-Path $dist "LICENSE") -Force
}

$guiExe = Join-Path $dist "AuronQ-Miner.exe"
$workerExe = Join-Path $dist "auronq-miner-worker.exe"
$nodeExe = Join-Path $dist "auronq.exe"
if (-not (Test-Path $guiExe)) { throw "Windows GUI executable missing: $guiExe" }
if (-not (Test-Path $workerExe)) { throw "Universal miner worker missing: $workerExe" }
if (-not (Test-Path $nodeExe)) { throw "Full-node CLI missing: $nodeExe" }

$guiHash = (Get-FileHash $guiExe -Algorithm SHA256).Hash
$workerHash = (Get-FileHash $workerExe -Algorithm SHA256).Hash
if ($guiHash -eq $workerHash) {
    throw "Invalid Windows package: GUI and miner worker are identical files."
}

Write-Host ""
Write-Host "[4/5] Running mandatory canonical CPU-fallback AQM64 self-test..."
Push-Location $dist
try {
    .\auronq-miner-worker.exe --backend cpu --cpu-threads 1 --self-test
    if ($LASTEXITCODE -ne 0) {
        throw "CPU fallback AQM64 self-test failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}

Write-Host ""
Write-Host "[5/5] Packaging..."
$zipOut = Join-Path $root "dist\AuronQ-Miner-v$version-Windows-x64.zip"
if (Test-Path $zipOut) { Remove-Item $zipOut -Force }
Compress-Archive -Path (Join-Path $dist "*") -DestinationPath $zipOut -CompressionLevel Optimal

Write-Host ""
Write-Host "SUCCESS"
Write-Host "Package: $zipOut"
Write-Host "GUI SHA256:    $guiHash"
Write-Host "Worker SHA256: $workerHash"
if ($cudaBuilt -and $legacyCudaBuilt -and $keplerCudaBuilt -and $openclBuilt) {
    Write-Host "Acceleration: current CUDA + legacy CUDA + Kepler CUDA + OpenCL GPU + CPU fallback"
} elseif ($cudaBuilt -and $legacyCudaBuilt -and $openclBuilt) {
    Write-Host "Acceleration: NVIDIA CUDA + legacy CUDA + vendor-neutral OpenCL GPU + native CPU fallback"
} elseif ($cudaBuilt -and $openclBuilt) {
    Write-Host "Acceleration: NVIDIA CUDA + vendor-neutral OpenCL GPU + native CPU fallback"
} elseif ($cudaBuilt) {
    Write-Host "Acceleration: NVIDIA CUDA + native CPU fallback"
} elseif ($openclBuilt) {
    Write-Host "Acceleration: OpenCL GPU + native CPU fallback"
} else {
    Write-Host "Acceleration: native CPU fallback"
}
Write-Host ""
Write-Host "Double-click AuronQ-Miner.exe. Leave compute backend on AUTO."
Write-Host "AUTO uses validated current CUDA, legacy CUDA, Kepler CUDA, OpenCL GPU, then CPU."
