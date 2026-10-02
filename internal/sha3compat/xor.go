// Portable little-endian XOR helpers derived from golang.org/x/crypto/sha3.
package sha3

import "encoding/binary"

func xorIn(d *state, buf []byte) {
	for i := 0; len(buf) >= 8; i++ {
		d.a[i] ^= binary.LittleEndian.Uint64(buf[:8])
		buf = buf[8:]
	}
}
func copyOut(d *state, b []byte) {
	for i := 0; len(b) >= 8; i++ {
		binary.LittleEndian.PutUint64(b[:8], d.a[i])
		b = b[8:]
	}
}