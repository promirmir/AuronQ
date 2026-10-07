#!/usr/bin/env bash
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CXX="${CXX:-g++}"
SRC="$HERE/aqm64_opencl.cpp"
OUT="$HERE/libauronq-aqm64-opencl.so"

echo "Building AuronQ vendor-neutral OpenCL backend..."
"$CXX" -O2 -std=c++17 -shared -fPIC "$SRC" -ldl -o "$OUT"

test -f "$OUT"
echo "Built: $OUT"
