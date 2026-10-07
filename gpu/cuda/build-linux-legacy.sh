#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NVCC="${NVCC:-nvcc}"
SRC="$HERE/aqm64_cuda.cu"
OUT="$HERE/libauronq-aqm64-cuda-legacy.so"

command -v "$NVCC" >/dev/null 2>&1 || { echo "nvcc not found; legacy build requires CUDA 12.x" >&2; exit 1; }

SUPPORTED_CODE="$("$NVCC" --list-gpu-code 2>/dev/null | tr '\n' ' ')"
SUPPORTED_ARCH="$("$NVCC" --list-gpu-arch 2>/dev/null | tr '\n' ' ')"
WANTED=(50 52 53 60 61 62 70 72)
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
  echo "This CUDA Toolkit cannot build maintained legacy Maxwell/Pascal/Volta targets. Use CUDA 12.x." >&2
  exit 1
fi

echo "Legacy CUDA targets: ${GENCODE[*]}"
"$NVCC" -O3 -std=c++17 -shared --cudart static -Xcompiler "-O2,-fPIC" "${GENCODE[@]}" "$SRC" -o "$OUT"
test -f "$OUT"
echo "Built: $OUT"
