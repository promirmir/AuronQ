#!/usr/bin/env bash
set -euo pipefail
BIN="${1:-./auronq}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORK="$ROOT/.smoke"
rm -rf "$WORK"; mkdir -p "$WORK"
export AURONQ_WALLET_PASSWORD='smoke-founder-password-not-for-production'
"$BIN" init-testnet --network "$WORK/network.json" --wallet "$WORK/founder.wallet" --threads 4 >/dev/null
unset AURONQ_WALLET_PASSWORD
"$BIN" node --network "$WORK/network.json" --data "$WORK/node" --listen 127.0.0.1:19444 >"$WORK/node.log" 2>&1 &
PID=$!
cleanup(){ kill "$PID" 2>/dev/null || true; wait "$PID" 2>/dev/null || true; }
trap cleanup EXIT
sleep 1
export AURONQ_WALLET_PASSWORD='smoke-recipient-password-not-for-production'
"$BIN" wallet-new --network "$WORK/network.json" --out "$WORK/recipient.wallet" >/dev/null
unset AURONQ_WALLET_PASSWORD
RECIPIENT="$($BIN wallet-info --wallet "$WORK/recipient.wallet" | awk '/^Address:/ {print $2}')"
export AURONQ_WALLET_PASSWORD='smoke-founder-password-not-for-production'
"$BIN" send --node http://127.0.0.1:19444 --network "$WORK/network.json" --wallet "$WORK/founder.wallet" --to "$RECIPIENT" --amount 25.50000000 >/dev/null
unset AURONQ_WALLET_PASSWORD
"$BIN" mine --node http://127.0.0.1:19444 --wallet "$WORK/founder.wallet" --threads 4 --once >/dev/null
BAL="$($BIN balance --node http://127.0.0.1:19444 --address "$RECIPIENT")"
echo "$BAL" | grep -q 'Spendable: 25.50000000 AURQ'
"$BIN" verify-network --network "$WORK/network.json" >/dev/null
echo "AuronQ smoke test: PASS"
cleanup; trap - EXIT