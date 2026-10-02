# AuronQ 1.7.0-rc4

This is the P2P propagation hotfix release candidate produced after a real two-node mainnet test exposed a missed-broadcast recovery gap.

Normal Windows use remains zero-touch: extract the ZIP and run `AuronQ-Desktop.exe`.

The mainnet genesis and Network ID are unchanged. rc4 changes networking behavior only: a node with the stronger chain can repair a known peer that is behind on the same branch by pushing the missing validated blocks during periodic synchronization.

The published `network.json` points at the stable official bootstrap manifest:

`https://raw.githubusercontent.com/promirmir/AuronQ/main/bootstrap.json`

That URL is discovery metadata only and is excluded from Network ID. The manifest may rotate reachable seed URLs without changing genesis or consensus.
