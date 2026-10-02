#!/usr/bin/env bash
set -euo pipefail
BIN="${1:-./auronq}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WORK="$ROOT/.two-node"
rm -rf "$WORK"; mkdir -p "$WORK"

export AURONQ_WALLET_PASSWORD='integration-founder-password-not-for-production'
"$BIN" init-testnet --network "$WORK/network.json" --wallet "$WORK/founder.wallet" --threads 4 >/dev/null
unset AURONQ_WALLET_PASSWORD

"$BIN" node --network "$WORK/network.json" --data "$WORK/node1" --listen 127.0.0.1:19544 >"$WORK/node1.log" 2>&1 &
P1=$!
P2=""
cleanup(){
  kill "$P1" 2>/dev/null || true
  if [[ -n "$P2" ]]; then kill "$P2" 2>/dev/null || true; fi
  wait "$P1" 2>/dev/null || true
  if [[ -n "$P2" ]]; then wait "$P2" 2>/dev/null || true; fi
}
trap cleanup EXIT
sleep 1

"$BIN" mine --node http://127.0.0.1:19544 --wallet "$WORK/founder.wallet" --threads 4 --once >/dev/null

"$BIN" node --network "$WORK/network.json" --data "$WORK/node2" --listen 127.0.0.1:19545 --peers http://127.0.0.1:19544 >"$WORK/node2.log" 2>&1 &
P2=$!
sleep 2
H2="$($BIN status --node http://127.0.0.1:19545 | awk '/^Height:/ {print $2}')"
[[ "$H2" == "1" ]]

# Relay a signed transaction from node 2 to its configured peer (node 1).
export AURONQ_WALLET_PASSWORD='integration-recipient-password-not-for-production'
"$BIN" wallet-new --network "$WORK/network.json" --out "$WORK/recipient.wallet" >/dev/null
unset AURONQ_WALLET_PASSWORD
RECIPIENT="$($BIN wallet-info --wallet "$WORK/recipient.wallet" | awk '/^Address:/ {print $2}')"
export AURONQ_WALLET_PASSWORD='integration-founder-password-not-for-production'
"$BIN" send --node http://127.0.0.1:19545 --network "$WORK/network.json" --wallet "$WORK/founder.wallet" --to "$RECIPIENT" --amount 1.00000000 >/dev/null
unset AURONQ_WALLET_PASSWORD
sleep 1
S1="$($BIN status --node http://127.0.0.1:19544)"
S2="$($BIN status --node http://127.0.0.1:19545)"
H1="$(printf '%s\n' "$S1" | awk '/^Height:/ {print $2}')"
H2="$(printf '%s\n' "$S2" | awk '/^Height:/ {print $2}')"
T1="$(printf '%s\n' "$S1" | awk '/^Tip:/ {print $2}')"
T2="$(printf '%s\n' "$S2" | awk '/^Tip:/ {print $2}')"
M1="$(printf '%s\n' "$S1" | awk '/^Mempool:/ {print $2}')"
M2="$(printf '%s\n' "$S2" | awk '/^Mempool:/ {print $2}')"
[[ "$H1" == "1" && "$H2" == "1" && "$T1" == "$T2" && "$M1" == "1" && "$M2" == "1" ]]

echo "AuronQ two-node sync/relay test: PASS"
cleanup; trap - EXIT