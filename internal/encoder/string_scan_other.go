//go:build !amd64

package encoder

import (
	"math/bits"
	"unsafe"
)

// stringptr returns the underlying data pointer of a string.
//
//go:nosplit
func stringptr(s string) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&s))
}

// scanEscapeBasic scans n bytes starting at p for characters that need JSON escaping.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), non-ASCII (>= 0x80), '"', '\'.
// Fallback implementation using uint64 chunk processing.
func scanEscapeBasic(p unsafe.Pointer, n int) int {
	i := 0
	// Process 8 bytes at a time
	for i+8 <= n {
		chunk := *(*uint64)(unsafe.Pointer(uintptr(p) + uintptr(i)))
		mask := chunk | (chunk - (lsb * 0x20)) |
			((chunk ^ (lsb * '"')) - lsb) |
			((chunk ^ (lsb * '\\')) - lsb)
		if (mask & msb) != 0 {
			return i + bits.TrailingZeros64(mask&msb)/8
		}
		i += 8
	}
	// Process remaining bytes
	bp := (*[1 << 30]byte)(p)
	for i < n {
		c := bp[i]
		if c < 0x20 || c >= 0x80 || c == '"' || c == '\\' {
			return i
		}
		i++
	}
	return n
}

// scanEscapeHTML scans n bytes starting at p for characters that need HTML-safe JSON escaping.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), non-ASCII (>= 0x80), '"', '\', '<', '>', '&'.
// Fallback implementation using uint64 chunk processing.
func scanEscapeHTML(p unsafe.Pointer, n int) int {
	i := 0
	// Process 8 bytes at a time
	for i+8 <= n {
		chunk := *(*uint64)(unsafe.Pointer(uintptr(p) + uintptr(i)))
		mask := chunk | (chunk - (lsb * 0x20)) |
			((chunk ^ (lsb * '"')) - lsb) |
			((chunk ^ (lsb * '\\')) - lsb) |
			((chunk ^ (lsb * '<')) - lsb) |
			((chunk ^ (lsb * '>')) - lsb) |
			((chunk ^ (lsb * '&')) - lsb)
		if (mask & msb) != 0 {
			return i + bits.TrailingZeros64(mask&msb)/8
		}
		i += 8
	}
	// Process remaining bytes
	bp := (*[1 << 30]byte)(p)
	for i < n {
		c := bp[i]
		if c < 0x20 || c >= 0x80 || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			return i
		}
		i++
	}
	return n
}
