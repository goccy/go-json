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
