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

SUPPORTED_CODE="$("$NVCC" --list-gpu-code 2>/dev/null | tr '\n' ' ')"
SUPPORTED_ARCH="$("$NVCC" --list-gpu-arch 2>/dev/null | tr '\n' ' ')"
WANTED=(61 70 72 75 80 86 87 88 89 90 100 103 110 120 121)
GENCODE=()
LATEST=""
for cc in "${WANTED[@]}"; do
  if grep -Eq "(^|[[:space:]])sm_${cc}([[:space:]]|$)" <<<"$SUPPORTED_CODE"; then
    GENCODE+=( -gencode "arch=compute_${cc},code=sm_${cc}" )
  fi
  if grep -Eq "(^|[[:space:]])compute_${cc}([[:space:]]|$)" <<<"$SUPPORTED_ARCH"; then
    LATEST="$cc"
  fi
done
if [[ -n "$LATEST" ]]; then
  GENCODE+=( -gencode "arch=compute_${LATEST},code=compute_${LATEST}" )
fi
if [[ "${#GENCODE[@]}" -eq 0 ]]; then
  echo "No supported CUDA GPU architecture targets detected from nvcc." >&2
  exit 1
fi
echo "CUDA targets: ${GENCODE[*]}"

"$NVCC" -O3 -std=c++17 -shared --cudart static -Xcompiler "-O2,-fPIC" "${GENCODE[@]}" "$SRC" -o "$OUT"

test -f "$OUT"
echo "Built: $OUT"
