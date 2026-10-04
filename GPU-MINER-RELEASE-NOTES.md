# AuronQ GPU Miner v0.2.1 Alpha

First public Windows alpha of the standalone NVIDIA CUDA miner for AuronQ Mainnet.

## Included

- native Windows application styled like AuronQ Desktop
- Polish and English interface
- NVIDIA CUDA AQM64 mining
- GPU/CPU AQM64 self-test
- offline benchmark
- live hashrate, block and node status
- use of an existing AuronQ Desktop full node when available
- own full validating node when Desktop is not running
- P2P peer discovery and persisted peer store
- optional UPnP TCP/18444 mapping and peer announcement while mining
- protection against starting two conflicting local AuronQ nodes
- distinct GUI and GPU worker executables
- statically linked CUDA runtime in the packaged DLL

## Real hardware validation

Validated on an NVIDIA GeForce RTX 4050 Laptop GPU.

Full GPU/CPU AQM64 equivalence self-test: PASS.

Measured prototype throughput:
- batch 20: 118.733 H/s
- batch 40: 226.263 H/s
- batch 60: 316.227 H/s
- batch 64: 299.377 H/s

Real AuronQ Mainnet blocks were found and accepted through the ordinary full-node validation path.

## Requirements

- Windows x64
- NVIDIA GPU with CUDA Compute Capability 7.5+
- current NVIDIA driver

The packaged release does not require the CUDA Toolkit. The Toolkit is only needed to build from source.

## Important

This is alpha software. The CUDA implementation has not received an independent professional security/cryptography audit. There are no consensus changes in this release. Every submitted block is validated by the normal AuronQ full node.

If UPnP is unavailable or the machine is behind CGNAT, mining still works, but the node remains outbound-only unless TCP/18444 is made publicly reachable by another method.
