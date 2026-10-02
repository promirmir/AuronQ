# Final AuronQ Mainnet Genesis Ceremony — 1.7.0-rc2

This procedure creates the private ML-DSA-87 seed controlling the transparent 210,000 AURQ genesis founder allocation and mines the immutable mainnet genesis. AuronQ 1.7.0-rc2 enables this operation only with `--ack-unaudited-mainnet`.

The implementation has passed the project's internal test/vet/race checks and cross-backend ML-DSA interoperability checks, but it has **not** received an independent professional consensus/cryptographic audit. Read `AUDIT-2026-10-02.md` before proceeding.

## 1. Verify the exact source/build

```bash
go test ./...
go vet ./...
go test -race ./...
go build -trimpath -ldflags='-s -w' -o auronq ./cmd/auronq
./auronq version
sha256sum auronq
```

Expected version: `AuronQ 1.7.0-rc2`.

OpenSSL is not required for normal operation; AuronQ ships a portable ML-DSA-87 implementation. OpenSSL 3.5+ may be used independently for interoperability testing.

## 2. Create the password without storing it in shell history

```bash
umask 077
read -r -s -p 'Founder wallet password: ' PW; echo
printf '%s' "$PW" > founder-password.txt
unset PW
chmod 600 founder-password.txt
```

Use a unique high-entropy password.

## 3. Generate the wallet and mine genesis

```bash
./auronq init-mainnet \
  --network network.json \
  --wallet founder.wallet \
  --password-file founder-password.txt \
  --name 'AuronQ Mainnet' \
  --threads 8 \
  --ack-unaudited-mainnet
```

The command stages both outputs and commits them only after a valid production-AQM64 genesis is found and validated. Do not regenerate genesis merely to obtain a preferred address/hash.

## 4. Record public facts

```bash
./auronq verify-network --network network.json | tee GENESIS-RECORD.txt
sha256sum network.json auronq >> GENESIS-RECORD.txt
./auronq wallet-info --wallet founder.wallet
```

Confirm that the founder address is identical in the wallet and network definition.

## 5. Back up the private wallet

Create at least two encrypted offline copies of `founder.wallet`, store the password separately, and verify the copies. After verification, remove `founder-password.txt` from any machine that will remain online.

`network.json` is public. `founder.wallet`, password and seed are private.

## 6. Add public bootstrap infrastructure

Bring up several reachable nodes using the exact same `network.json`. Then add seed/DNS bootstrap metadata with `network-set-seeds`. Bootstrap fields are intentionally excluded from Network ID, so seed rotation does not create a consensus fork.

## 7. Freeze and publish

Publish the exact source tag, `network.json`, Network ID, genesis hash, founder address/allocation, release binaries and SHA-256 checksums together. Any later consensus change is a protocol upgrade and must never be disguised as replacement of the original genesis/network definition.