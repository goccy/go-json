//go:build !amd64

package encoder

import (
	"unsafe"
)

// hasEscapeSIMD is not available: the string is scanned by words.
func (e *stringEscape) hasEscapeSIMD(_ unsafe.Pointer, _ int) (found, ok bool) {
	return false, false
}
