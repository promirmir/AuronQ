#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-0.4.6-alpha}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST="$ROOT/dist"
DIR="$DIST/AuronQ-Miner-v$VERSION-Linux-amd64"
ARCHIVE="$DIST/AuronQ-Miner-v$VERSION-Linux-amd64.tar.gz"

mkdir -p "$DIST"
rm -rf "$DIR" "$ARCHIVE"
mkdir -p "$DIR"

CUDA_BUILT=0
LEGACY_CUDA_BUILT=0
OPENCL_BUILT=0
if command -v nvcc >/dev/null 2>&1; then
  echo "Building optional NVIDIA CUDA backend..."
  bash "$ROOT/gpu/cuda/build-linux.sh"
  if [[ -f "$ROOT/gpu/cuda/libauronq-aqm64-cuda.so" ]]; then
    CUDA_BUILT=1
  fi
else
  echo "nvcc not found - CUDA acceleration will be skipped."
fi

if command -v g++ >/dev/null 2>&1; then
  echo "Building vendor-neutral OpenCL backend..."
  bash "$ROOT/gpu/opencl/build-linux.sh"
  if [[ -f "$ROOT/gpu/opencl/libauronq-aqm64-opencl.so" ]]; then
    OPENCL_BUILT=1
  fi
else
  echo "g++ not found - OpenCL backend build skipped."
fi

echo "Building Linux Universal Miner..."
(
  cd "$ROOT"
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -o "$DIR/auronq-miner" ./cmd/auronq-gpu-miner
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -o "$DIR/auronq" ./cmd/auronq
)

if [[ "$CUDA_BUILT" == "1" ]]; then
  cp "$ROOT/gpu/cuda/libauronq-aqm64-cuda.so" "$DIR/"
fi
if [[ "$LEGACY_CUDA_BUILT" == "1" ]]; then
  cp "$ROOT/gpu/cuda/libauronq-aqm64-cuda-legacy.so" "$DIR/"
fi
if [[ "$OPENCL_BUILT" == "1" ]]; then
  cp "$ROOT/gpu/opencl/libauronq-aqm64-opencl.so" "$DIR/"
fi
cp "$ROOT/network.json" "$DIR/"
cp "$ROOT/bootstrap.json" "$DIR/"
cp "$ROOT/UNIVERSAL-MINER-GUIDE.md" "$DIR/"
cp "$ROOT/GPU-MINER-GUIDE.md" "$DIR/"
if [[ -f "$ROOT/LICENSE" ]]; then
  cp "$ROOT/LICENSE" "$DIR/"
fi

cat > "$DIR/START-HERE-LINUX.txt" <<'TXT'
AuronQ Universal Miner - Linux amd64

1. Start the bundled AuronQ full node:
   ./auronq node --network ./network.json --data ./node-data --listen 127.0.0.1:18444

2. In another terminal run the mandatory backend correctness test:
   ./auronq-miner --backend auto --self-test

3. Solo mining:
   ./auronq-miner --backend auto --node http://127.0.0.1:18444 --address aurq1... --self-test --thermal-auto --thermal-limit 81

AUTO uses validated NVIDIA CUDA first, then validated legacy CUDA for older
NVIDIA generations, then vendor-neutral OpenCL GPU, then native CPU AQM64. Generic OpenCL GPU mining uses a conservative
no-temperature-sensor duty profile unless a vendor-specific safety path exists.

For a completely portable CPU-only build, use build-universal-miner-packages.sh.
See UNIVERSAL-MINER-GUIDE.md.
TXT

chmod +x "$DIR/auronq-miner" "$DIR/auronq"

echo "Running canonical CPU-fallback self-test..."
"$DIR/auronq-miner" --backend cpu --cpu-threads 1 --self-test

tar -C "$DIST" -czf "$ARCHIVE" "$(basename "$DIR")"
sha256sum "$ARCHIVE" > "$DIST/SHA256SUMS-MINER-LINUX.txt"

echo "Built: $ARCHIVE"
if [[ "$CUDA_BUILT" == "1" && "$OPENCL_BUILT" == "1" ]]; then
  echo "Acceleration: NVIDIA CUDA + OpenCL GPU + native CPU fallback"
elif [[ "$CUDA_BUILT" == "1" ]]; then
  echo "Acceleration: NVIDIA CUDA + native CPU fallback"
elif [[ "$OPENCL_BUILT" == "1" ]]; then
  echo "Acceleration: OpenCL GPU + native CPU fallback"
else
  echo "Acceleration: native CPU fallback"
fi
