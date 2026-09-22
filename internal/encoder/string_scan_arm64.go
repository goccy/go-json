package encoder

import (
	"unsafe"
)

// The scan of a string for a byte to escape by NEON, 16 bytes at a time.

//go:noescape
func scanStringNEON(p unsafe.Pointer, n int, tables *nibbleTables) int

// minSIMDScanLength is the length from which a string is scanned by SIMD.
const minSIMDScanLength = 16

// hasEscapeSIMD is whether a byte of the string may need an escape, by SIMD. The second result is false if
// the string is short, and the string is not looked at.
func (e *stringEscape) hasEscapeSIMD(src unsafe.Pointer, n int) (bool, bool) {
	if n < minSIMDScanLength {
		return false, false
	}
	return scanStringNEON(src, n, &e.tables) != 0, true
}
