#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-0.3.6-alpha}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST="$ROOT/dist"
DIR="$DIST/AuronQ-GPU-Miner-v$VERSION-Linux-amd64"
ARCHIVE="$DIST/AuronQ-GPU-Miner-v$VERSION-Linux-amd64.tar.gz"

mkdir -p "$DIST"
rm -rf "$DIR" "$ARCHIVE"
mkdir -p "$DIR"

bash "$ROOT/gpu/cuda/build-linux.sh"

echo "Building Linux miner..."
(
  cd "$ROOT"
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -o "$DIR/auronq-gpu-miner" ./cmd/auronq-gpu-miner
)

cp "$ROOT/gpu/cuda/libauronq-aqm64-cuda.so" "$DIR/"
cp "$ROOT/network.json" "$DIR/"
cp "$ROOT/bootstrap.json" "$DIR/"
cp "$ROOT/gpu/cuda/README.md" "$DIR/README-GPU-MINER.md"
if [[ -f "$ROOT/GPU-MINER-GUIDE.md" ]]; then
  cp "$ROOT/GPU-MINER-GUIDE.md" "$DIR/"
fi
if [[ -f "$ROOT/LICENSE" ]]; then
  cp "$ROOT/LICENSE" "$DIR/"
fi

cat > "$DIR/START-HERE-LINUX.txt" <<'TXT'
AuronQ GPU Miner - Linux amd64

1. Install a current proprietary NVIDIA driver and verify:
   nvidia-smi

2. Run the correctness test:
   ./auronq-gpu-miner --self-test --device 0

3. Solo mining example:
   ./auronq-gpu-miner --node http://127.0.0.1:18444 --address aurq1... --devices all --self-test --auto-tune --thermal-auto --thermal-limit 81

See GPU-MINER-GUIDE.md for the complete Windows/Linux guide.
TXT

chmod +x "$DIR/auronq-gpu-miner"
tar -C "$DIST" -czf "$ARCHIVE" "$(basename "$DIR")"
sha256sum "$ARCHIVE" > "$DIST/SHA256SUMS-GPU-MINER-LINUX.txt"

echo "Built: $ARCHIVE"
