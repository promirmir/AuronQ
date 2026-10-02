# AuronQ-PoW v1 (AQM64)

AQM64 is the proof-of-work construction introduced for the AuronQ 1.6.0 pre-mainnet testnet. It is AuronQ-specific **composition**, not a newly invented cryptographic primitive.

## Consensus parameters

- algorithm ID: `0x0001`
- Argon2 mode/version: Argon2id v1.3
- memory: 65,536 KiB per mining lane
- time cost: 2
- parallelism: 1
- Argon2 output: 64 bytes
- pre/final XOF: SHAKE256
- final proof: 512 bits, interpreted big-endian against target
- target block interval: 600 seconds

See `PROTOCOL.md` for the byte-exact construction and domain separators.

## Design intent

AQM64 makes each nonce attempt memory-expensive so mining is not simply a very high-throughput SHA loop. The design is intended to make ordinary CPU mining useful during network bootstrapping and to raise the memory/bandwidth cost of specialized implementations. No claim of ASIC resistance or ASIC-proofness is made.

## Mainnet status

AQM64 parameters used by 1.7.0-rc2 are frozen into the rc2 consensus fingerprint/Network ID and `init-mainnet` is enabled only with an explicit unaudited-mainnet acknowledgement. This is a compatibility freeze, not an independent security endorsement. Public CPU/GPU benchmarking, independent deterministic-vector reproduction, time-memory tradeoff analysis, malicious resource-exhaustion testing and independent consensus/cryptographic review remain outstanding.