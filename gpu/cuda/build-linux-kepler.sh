#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NVCC="${NVCC:-nvcc}"
SRC="$HERE/aqm64_cuda.cu"
OUT="$HERE/libauronq-aqm64-cuda-kepler.so"

command -v "$NVCC" >/dev/null 2>&1 || { echo "nvcc not found; Kepler build requires CUDA 11.8" >&2; exit 1; }

SUPPORTED_CODE="$("$NVCC" --list-gpu-code 2>/dev/null | tr '\n' ' ')"
SUPPORTED_ARCH="$("$NVCC" --list-gpu-arch 2>/dev/null | tr '\n' ' ')"
WANTED=(35 37)
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
  echo "This CUDA Toolkit cannot build maintained Kepler targets sm_35/sm_37. Use CUDA 11.8." >&2
  exit 1
fi

echo "Kepler CUDA targets: ${GENCODE[*]}"
"$NVCC" -O3 -std=c++17 -shared --cudart static -Xcompiler "-O2,-fPIC" "${GENCODE[@]}" "$SRC" -o "$OUT"
test -f "$OUT"
echo "Built: $OUT"
