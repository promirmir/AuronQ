# Contributing to AuronQ

Contributions that improve correctness, security, portability, documentation, testing and network resilience are welcome.

## Before opening a pull request

1. Search existing issues and pull requests.
2. Keep the change focused on one problem.
3. For consensus-sensitive changes, describe the exact compatibility impact.
4. Add or update tests where practical.
5. Do not mix consensus changes with unrelated refactors.

## Development checks

The standard local checks are:

```bash
go test ./...
go vet ./...
```

On Linux, also run:

```bash
go test -race ./...
```

The repository Makefile also provides:

```bash
make test
make vet
make build
make smoke
make integration
```

## Consensus and cryptography changes

Changes affecting any of the following require exceptional care:

- block or transaction validity;
- chain selection;
- difficulty or Proof-of-Work;
- serialization;
- Network ID or genesis assumptions;
- transaction signatures;
- wallet key handling;
- peer synchronization rules.

A pull request in these areas should explain:

- the exact rule being changed;
- backward-compatibility implications;
- whether a network upgrade is required;
- test coverage;
- security assumptions.

Do not silently change Mainnet consensus behavior.

## Documentation contributions

Documentation improvements are encouraged, especially for:

- node setup;
- mining;
- wallet usage;
- reproducible builds;
- protocol behavior;
- peer discovery;
- threat modeling;
- independent review.

Avoid promotional claims that exceed what the code, tests or independent evidence demonstrate.

## Security reports

Do not disclose exploitable vulnerabilities in a public issue. Follow [SECURITY.md](SECURITY.md).

## Pull request checklist

- [ ] The change has a clear purpose.
- [ ] Tests pass locally.
- [ ] Relevant documentation was updated.
- [ ] No unrelated files were changed.
- [ ] Consensus impact is explicitly stated.
- [ ] Security implications were considered.
