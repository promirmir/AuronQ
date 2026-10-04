$ErrorActionPreference = "Stop"

function Import-MSVCEnvironment {
    if (Get-Command cl.exe -ErrorAction SilentlyContinue) {
        return
    }

    $vswhere = Join-Path ${env:ProgramFiles(x86)} "Microsoft Visual Studio\Installer\vswhere.exe"
    $vsPath = $null

    if (Test-Path $vswhere) {
        $raw = & $vswhere -latest -products * -property installationPath
        if ($raw) {
            $vsPath = ($raw | Select-Object -First 1).Trim()
        }
    }

    if (-not $vsPath) {
        $roots = @(
            "C:\Program Files\Microsoft Visual Studio",
            "C:\Program Files (x86)\Microsoft Visual Studio"
        )
        foreach ($root in $roots) {
            if (-not (Test-Path $root)) { continue }
            $candidate = Get-ChildItem $root -Filter VsDevCmd.bat -File -Recurse -ErrorAction SilentlyContinue |
                Select-Object -First 1
            if ($candidate) {
                $vsPath = Split-Path -Parent (Split-Path -Parent $candidate.FullName)
                break
            }
        }
    }

    if (-not $vsPath) {
        throw "Visual Studio installation was not found. Open Visual Studio Installer and install 'Desktop development with C++'."
    }

    $devCmd = Join-Path $vsPath "Common7\Tools\VsDevCmd.bat"
    if (-not (Test-Path $devCmd)) {
        $candidate = Get-ChildItem $vsPath -Filter VsDevCmd.bat -File -Recurse -ErrorAction SilentlyContinue |
            Select-Object -First 1
        if ($candidate) {
            $devCmd = $candidate.FullName
        }
    }

    if (-not (Test-Path $devCmd)) {
        throw "VsDevCmd.bat was not found in Visual Studio at: $vsPath"
    }

    Write-Host "Loading MSVC environment from:"
    Write-Host "  $devCmd"

    $tmpCmd = Join-Path $env:TEMP ("auronq-vsenv-" + [guid]::NewGuid().ToString("N") + ".cmd")
    try {
        @(
            "@echo off",
            ('call "{0}" -arch=amd64 -host_arch=amd64 >nul' -f $devCmd),
            "if errorlevel 1 exit /b %errorlevel%",
            "set"
        ) | Set-Content -Path $tmpCmd -Encoding ASCII

        $envLines = & cmd.exe /d /c $tmpCmd
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to initialize Visual Studio C++ build environment."
        }
    } finally {
        Remove-Item $tmpCmd -Force -ErrorAction SilentlyContinue
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
        $clCandidate = Get-ChildItem $vsPath -Filter cl.exe -File -Recurse -ErrorAction SilentlyContinue |
            Where-Object { $_.FullName -match "Hostx64\\x64\\cl\.exe$" } |
            Select-Object -First 1
        if ($clCandidate) {
            $env:PATH = "$($clCandidate.Directory.FullName);$env:PATH"
        }
    }

    if (-not (Get-Command cl.exe -ErrorAction SilentlyContinue)) {
        throw "cl.exe was not found. In Visual Studio Installer add 'Desktop development with C++' (MSVC x64/x86 build tools + Windows SDK)."
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
    "--cudart", "static",
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
