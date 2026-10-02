# AuronQ Mainnet Launch — 1.7.0-rc2

AuronQ 1.7.0-rc2 is **mainnet-capable**, but it is not independently audited financial infrastructure. `init-mainnet` therefore requires the explicit `--ack-unaudited-mainnet` acknowledgement.

## What becomes immutable

The first final `network.json` fixes the founder address, genesis block, network byte, coinbase maturity, protocol version and the consensus fingerprint/Network ID. Keep the founder wallet private. Publish `network.json`; never publish `founder.wallet`, its password or seed.

Bootstrap metadata (`seed_peers` and `dns_seeds`) can be updated later without changing Network ID.

## 1. Freeze and verify the exact release source

Run from the source tree that will be tagged publicly:

```bash
go test ./...
go vet ./...
go test -race ./...
go build -trimpath -ldflags='-s -w' -o auronq ./cmd/auronq
./auronq version
```

The expected version is `AuronQ 1.7.0-rc2`.

## 2. Generate the founder wallet and genesis locally

Do this on a trusted machine. The founder private key must be generated on your computer, never by a website, seed node or third party.

### Linux/macOS

This avoids putting the actual password into shell history:

```bash
umask 077
read -r -s -p 'Founder wallet password: ' PW; echo
printf '%s' "$PW" > founder-password.txt
unset PW
chmod 600 founder-password.txt

./auronq init-mainnet \
  --network network.json \
  --wallet founder.wallet \
  --password-file founder-password.txt \
  --name "AuronQ Mainnet" \
  --threads 8 \
  --ack-unaudited-mainnet
```

### Windows PowerShell

This also keeps the password itself out of PowerShell history:

```powershell
$sec = Read-Host "Founder wallet password" -AsSecureString
$bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($sec)
try {
  $env:AURONQ_WALLET_PASSWORD = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($bstr)
  .\auronq-cli.exe init-mainnet --network .\network.json --wallet .\founder.wallet --name "AuronQ Mainnet" --threads 8 --ack-unaudited-mainnet
} finally {
  Remove-Item Env:AURONQ_WALLET_PASSWORD -ErrorAction SilentlyContinue
  [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
}
```

Genesis mining uses production AQM64 and is probabilistic. Do not interrupt it merely because a block is not found near the statistical mean.

## 3. Verify and record immutable public facts

Linux/macOS:

```bash
./auronq verify-network --network network.json | tee GENESIS-RECORD.txt
sha256sum network.json auronq >> GENESIS-RECORD.txt
./auronq wallet-info --wallet founder.wallet
```

Windows:

```powershell
.\auronq-cli.exe verify-network --network .\network.json
Get-FileHash .\network.json -Algorithm SHA256
Get-FileHash .\auronq-cli.exe -Algorithm SHA256
.\auronq-cli.exe wallet-info --wallet .\founder.wallet
```

The founder address printed by `wallet-info` must match the founder address recorded in `network.json`/`verify-network`.

## 4. Back up private material before exposing a node

Create at least two offline copies of `founder.wallet` and keep the password separately. Verify that the copies are readable. Losing both the wallet and its recovery copies makes the 210,000 AURQ founder allocation inaccessible; obtaining the encrypted wallet **and** password allows an attacker to spend it.

After verified backups, delete any temporary plaintext password file from a machine that will remain network-connected.

## 5. Start public bootstrap/full nodes

For a real public release, use at least three publicly reachable nodes on different providers/networks where practical. Seeds are rendezvous points only; they have no consensus authority.

Example:

```bash
./auronq node \
  --network network.json \
  --data /var/lib/auronq \
  --listen 0.0.0.0:18444 \
  --advertise http://PUBLIC_IP:18444
```

Then add public bootstrap entries:

```bash
./auronq network-set-seeds \
  --network network.json \
  --seed http://SEED1:18444,http://SEED2:18444,http://SEED3:18444
```

Re-run `verify-network`: Network ID must remain unchanged after bootstrap-only edits.

## 6. Build the public user package

Place the final public `network.json` **in the same directory as `AuronQ-Desktop.exe`**. On first launch, Desktop 1.7.0-rc2 verifies and copies that bundled network definition into the user's AuronQ configuration directory automatically. Existing user network configuration is never silently overwritten.

A fresh user experience can therefore be: extract release -> run Desktop -> connect/sync, without manually entering peer IPs, provided the published `network.json` contains reachable bootstrap infrastructure.

## 7. Publish GitHub release facts

Recommended release-candidate tag: `v1.7.0-rc2`.

Publish:

- exact source tree/tag/commit;
- Windows/Linux binaries;
- final public `network.json`;
- Network ID and genesis hash;
- founder public address and disclosed 210,000 AURQ allocation;
- SHA-256 checksums;
- protocol/security/audit documents.

Never publish `founder.wallet`, wallet passwords, seed material, node private data or temporary ceremony files.

## 8. Claims that must remain accurate

It is accurate to say AuronQ uses ML-DSA-87 post-quantum transaction signatures and the AQM64 proof-of-work construction. It is **not** accurate to call the code independently audited, formally verified, ASIC-proof, unhackable or guaranteed quantum-safe forever.

See `AUDIT-2026-10-02.md` for the internal pre-release audit and the unresolved risks that remain before material real-world value should depend on the network.
