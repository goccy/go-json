//go:build !amd64 && !arm64

package encoder

import (
	"unsafe"
)

// hasEscapeSIMD is not available: the string is scanned by words.
func (e *stringEscape) hasEscapeSIMD(_ unsafe.Pointer, _ int) (bool, bool) {
	return false, false
}

// scanBytesSIMD is not available: the bytes are scanned one by one.
func scanBytesSIMD(_ unsafe.Pointer, _ int, _ *nibbleTables) (bool, bool) {
	return false, false
}

// indexEscapeSIMD scans nothing: the index of the first byte to escape has no SIMD scan on this architecture,
// whose words are scanned instead.
func indexEscapeSIMD(src unsafe.Pointer, n int, tables *nibbleTables) (int, bool) {
	return 0, false
}
