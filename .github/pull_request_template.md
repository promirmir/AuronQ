## Summary

Describe the change and the problem it solves.

## Type of change

- [ ] Bug fix
- [ ] Documentation
- [ ] Tests
- [ ] Build / CI
- [ ] Networking
- [ ] Wallet
- [ ] Mining
- [ ] Consensus-sensitive change
- [ ] Other

## Validation

- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] Relevant platform build tested
- [ ] Documentation updated where needed

## Consensus / compatibility impact

State explicitly whether this changes block validity, transaction validity, serialization, chain selection, Proof-of-Work, Network ID, genesis assumptions, wallet format, peer protocol or mining behavior.

If none, write: **No consensus or protocol impact.**

## Security considerations

Describe relevant security implications, or state why none are expected.

## Mainnet safety gate

If this PR changes any consensus-sensitive behavior (block/transaction validity, serialization, PoW, difficulty, timestamps, chain selection, issuance, signature verification, Network ID or genesis):

- [ ] This is **not** being treated as an ordinary patch release.
- [ ] A dedicated testnet/devnet activation plan exists.
- [ ] Compatibility / fork behavior is documented.
- [ ] Mainnet activation requires an explicit version and activation mechanism.
- [ ] Independent review is requested before activation.

If none of the above apply, write: **No Mainnet consensus change.**

## Additional notes
