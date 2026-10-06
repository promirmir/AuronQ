#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NVCC="${NVCC:-$(command -v nvcc || true)}"
if [[ -z "$NVCC" ]]; then
  echo "nvcc not found. Install NVIDIA CUDA Toolkit 12.x/13.x or set NVCC=/path/to/nvcc." >&2
  exit 1
fi

SRC="$HERE/aqm64_cuda.cu"
OUT="$HERE/libauronq-aqm64-cuda.so"

echo "NVCC: $NVCC"
echo "Building AuronQ AQM64 CUDA backend for Linux amd64..."

"$NVCC"   -O3   -std=c++17   -shared   --cudart static   -Xcompiler "-O2,-fPIC"   -gencode arch=compute_75,code=sm_75   -gencode arch=compute_86,code=sm_86   -gencode arch=compute_89,code=sm_89   -gencode arch=compute_89,code=compute_89   "$SRC"   -o "$OUT"

test -f "$OUT"
echo "Built: $OUT"
