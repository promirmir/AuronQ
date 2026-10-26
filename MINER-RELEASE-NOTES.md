# AuronQ Universal Miner v0.4.9 Alpha

## Verified node failover and GitHub-independent operation

v0.4.9 packages the reviewed runtime-independence work merged through PR #189.

### Node resilience

- Solo/standalone mining prefers the configured or local full node.
- If that endpoint becomes unavailable, the miner can switch to another endpoint that independently reports the exact AuronQ Mainnet Network ID.
- Verified peers are cached locally for later starts.
- Repeatedly failing endpoints are quarantined locally.
- Peers learned through AuronQ peer gossip are independently probed before persistence.
- A clear same-height conflicting tip fails closed instead of silently choosing one chain.
- CPU remote mining stops after repeated node transport loss rather than continuing indefinitely against potentially stale work.
- Multi-GPU child workers inherit the same network configuration and use separate peer caches.

### Hardware compatibility retained

The existing AUTO backend order and safety behavior remain: current NVIDIA CUDA, legacy CUDA, Kepler CUDA, vendor-neutral OpenCL GPU, then native CPU. Every native accelerator must still pass the canonical AQM64 self-test before Mainnet mining.

### Role separation

The miner/light client does not pretend to be a P2P full node. Public peer advertisement is performed only by the actual full-node component when it is reachable.

## Consensus unchanged

No changes to genesis, Network ID, AQM64 consensus parameters, difficulty, monetary policy, block/transaction validity, wallet signing format or cumulative-work chain selection.

This remains alpha software. Hardware compatibility and long-duration unattended operation should still be tested on the actual machine before relying on it.
