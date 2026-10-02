# AuronQ (AURQ)

AuronQ is a Bitcoin-inspired UTXO proof-of-work cryptocurrency using ML-DSA-87 transaction signatures and the AQM64 proof-of-work construction.

## Mainnet identity

Network ID:

`44e62c2ace002a6660c14e252173c1aa303529c68e40c998e92da2b453f44f30b1e58c94d533587e2186004593fb856c433fcdb5418ed430ec8617e29529365c`

Genesis hash:

`5750a455c04bfe93c9edfef1a12744b05e29ac6da1a9dd5b790566629dea2080c581beb2f0324efba2067c9efb113ed7a29598fd3ffbed965ced31f265d0cec4`

Founder genesis allocation: 210,000 AURQ. Maximum supply: 21,000,000 AURQ. Initial block subsidy: 49.5 AURQ. Coinbase maturity: 100 blocks.

## Windows rc4

For two Windows PCs on the same private LAN:

1. Download the Windows x64 rc4 ZIP.
2. Extract the complete archive.
3. Run `START-AURONQ.cmd` on each PC and approve UAC on first launch.
4. The launcher adds a Windows Firewall TCP 18444 rule limited to the local subnet and opens AuronQ Desktop.
5. rc4 automatically scans the local RFC1918 /24 segment, verifies the AuronQ Network ID and adds matching LAN peers. No Tailscale, Cloudflare or manual peer entry is needed for two PCs on the same LAN.

The client also reads the stable bootstrap manifest:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

That manifest currently contains no global public seed. Therefore unrelated networks on the public Internet still require at least one globally reachable bootstrap node before zero-touch global discovery is possible.

## Verified rc4 behavior

A real two-computer test reached the same chain at height 5 and the same tip after mining and synchronization. rc4 also fixed the missed-broadcast case by pushing missing same-branch blocks to a peer that is behind.

The frozen local tree passed:

- `go test ./...`
- `go vet ./...`
- `go test -race ./...`
- Windows amd64 Desktop/CLI build
- Linux amd64 CLI build

## Security status

This is a release candidate. AQM64 and the overall consensus/network implementation have not received an independent professional security or cryptographic audit. Do not treat the project as production-audited financial infrastructure yet.
