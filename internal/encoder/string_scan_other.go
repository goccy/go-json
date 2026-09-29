//go:build !amd64 && !arm64

package encoder

import (
	"unsafe"
)

// maxOnePassLength is the length of the longest string which appendStringOnePass looks at and copies in one pass:
// a longer one is looked at by words, and copied after.
const maxOnePassLength = 64

// hasEscapeSIMD is not available: the string is scanned by words.
func (e *stringEscape) hasEscapeSIMD(_ unsafe.Pointer, _ int) (bool, bool) {
	return false, false
}

// scanBytesSIMD is not available: the bytes are scanned one by one.
func scanBytesSIMD(_ unsafe.Pointer, _ int, _ *nibbleTables) (bool, bool) {
	return false, false
}

// hasEscapeLoop is false: the loop of the escapes by SIMD is of amd64 only ( see string_scan_amd64.go ).
const hasEscapeLoop = false

// appendEscapedSIMD appends nothing: it is never called on this architecture.
func appendEscapedSIMD(buf []byte, _ string, _ *nibbleTables, _ bool) ([]byte, int) {
	return buf, 0
}
