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

mapfile -t SUPPORTED_CODE < <("$NVCC" --list-gpu-code 2>/dev/null | sed -n -E 's/^[[:space:]]*(sm_[0-9]+[a-z]?)[[:space:]]*$/\1/p')
mapfile -t SUPPORTED_ARCH < <("$NVCC" --list-gpu-arch 2>/dev/null | sed -n -E 's/^[[:space:]]*(compute_[0-9]+[a-z]?)[[:space:]]*$/\1/p')

declare -A ARCH_SET=()
for arch in "${SUPPORTED_ARCH[@]}"; do ARCH_SET["$arch"]=1; done

GENCODE=()
for code in "${SUPPORTED_CODE[@]}"; do
  suffix="${code#sm_}"
  arch="compute_${suffix}"
  if [[ -n "${ARCH_SET[$arch]:-}" ]]; then
    GENCODE+=( -gencode "arch=${arch},code=${code}" )
  fi
done

LATEST=""
if [[ "${#SUPPORTED_ARCH[@]}" -gt 0 ]]; then
  LATEST="$(printf '%s\n' "${SUPPORTED_ARCH[@]}" | sort -V | tail -n1)"
  GENCODE+=( -gencode "arch=${LATEST},code=${LATEST}" )
fi

if [[ "${#GENCODE[@]}" -eq 0 ]]; then
  echo "No CUDA GPU architecture targets detected from nvcc." >&2
  exit 1
fi
echo "CUDA targets from installed toolkit: ${GENCODE[*]}"

"$NVCC" -O3 -std=c++17 -shared --cudart static -Xcompiler "-O2,-fPIC" "${GENCODE[@]}" "$SRC" -o "$OUT"

test -f "$OUT"
echo "Built: $OUT"
