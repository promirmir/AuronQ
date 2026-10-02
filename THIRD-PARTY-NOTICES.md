# Third-party notices

AuronQ includes portable cryptographic code adapted from:

- **KarpelesLab/mldsa** — ML-DSA implementation following NIST FIPS 204, MIT License.
  Source: https://github.com/KarpelesLab/mldsa
- **golang.org/x/crypto/sha3** — SHAKE/SHA-3 sponge implementation, BSD-style Go license.

The portable AuronQ ML-DSA-87 path is used to make consensus verification and
wallet signing independent of Windows CNG / OpenSSL availability.
