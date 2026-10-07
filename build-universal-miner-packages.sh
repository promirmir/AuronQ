#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-0.4.4-alpha}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST="$ROOT/dist/universal-v$VERSION"

rm -rf "$DIST"
mkdir -p "$DIST"

targets=(
  "windows amd64"
  "windows arm64"
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
)

archives=()

for target in "${targets[@]}"; do
  read -r GOOS_VALUE GOARCH_VALUE <<<"$target"
  PLATFORM="$GOOS_VALUE-$GOARCH_VALUE"
  NAME="AuronQ-Universal-Miner-v$VERSION-$PLATFORM"
  DIR="$DIST/$NAME"
  mkdir -p "$DIR"

  EXT=""
  if [[ "$GOOS_VALUE" == "windows" ]]; then EXT=".exe"; fi

  echo "Building $PLATFORM ..."
  (
    cd "$ROOT"
    CGO_ENABLED=0 GOOS="$GOOS_VALUE" GOARCH="$GOARCH_VALUE" \
      go build -trimpath -o "$DIR/auronq-miner$EXT" ./cmd/auronq-gpu-miner
    CGO_ENABLED=0 GOOS="$GOOS_VALUE" GOARCH="$GOARCH_VALUE" \
      go build -trimpath -o "$DIR/auronq$EXT" ./cmd/auronq
  )

  cp "$ROOT/network.json" "$DIR/"
  cp "$ROOT/bootstrap.json" "$DIR/"
  cp "$ROOT/UNIVERSAL-MINER-GUIDE.md" "$DIR/"
  if [[ -f "$ROOT/LICENSE" ]]; then cp "$ROOT/LICENSE" "$DIR/"; fi

  cat > "$DIR/START-HERE.txt" <<TXT
AuronQ Universal Miner v$VERSION
Platform: $PLATFORM

SAFE DEFAULT
------------
Use backend AUTO. CUDA is used only when a compatible local accelerator is
available; otherwise the miner falls back to a conservative CPU profile.

1) Start the full node:
   auronq$EXT node --network network.json --data node-data --listen 127.0.0.1:18444

2) In a second terminal run the backend correctness test:
   auronq-miner$EXT --backend auto --self-test

3) Start mining with your AURQ address:
   auronq-miner$EXT --backend auto --node http://127.0.0.1:18444 --address aurq1... --self-test --thermal-auto --thermal-limit 81

CPU safe mode defaults to about one quarter of logical CPUs, max 2 lanes.
Each AQM64 lane uses about 64 MiB.
See UNIVERSAL-MINER-GUIDE.md for details.
TXT

  if [[ "$GOOS_VALUE" == "windows" ]]; then
    ARCHIVE="$DIST/$NAME.zip"
    (
      cd "$DIST"
      zip -qr "$(basename "$ARCHIVE")" "$NAME"
    )
  else
    chmod +x "$DIR/auronq-miner" "$DIR/auronq"
    ARCHIVE="$DIST/$NAME.tar.gz"
    tar -C "$DIST" -czf "$ARCHIVE" "$NAME"
  fi
  archives+=("$ARCHIVE")
done

(
  cd "$DIST"
  sha256sum "${archives[@]##*/}" > SHA256SUMS-UNIVERSAL-MINER.txt
)

echo "Universal packages:"
printf '  %s\n' "${archives[@]}"
