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

// scanEscapeBasicASCIIOnly scans for characters needing JSON escaping, ignoring non-ASCII bytes.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), '"', '\'. Does NOT flag bytes >= 0x80.
func scanEscapeBasicASCIIOnly(p unsafe.Pointer, n int) int {
	i := 0
	bp := (*[1 << 30]byte)(p)
	// Process 8 bytes at a time
	for i+8 <= n {
		chunk := *(*uint64)(unsafe.Pointer(uintptr(p) + uintptr(i)))
		// Check for control chars (< 0x20) excluding high bytes:
		// Use the hasless technique but mask out high-bit bytes first
		lowASCII := chunk & ^(chunk >> 1) & (lsb * 0x40) // detect bytes where bit6=1 and bit7=0... no
		// Simpler: check each special char and control range
		// For control chars: byte < 0x20 AND byte < 0x80
		// Control chars have bit7=0 and bits6..5 = 00 (values 0x00-0x1F)
		// Use: NOT(bit7) AND NOT(bit6) AND NOT(bit5) -> too complex for uint64
		// Just use scalar for simplicity in fallback
		hasEscape := false
		for k := 0; k < 8; k++ {
			c := bp[i+k]
			if c < 0x20 || c == '"' || c == '\\' {
				return i + k
			}
			_ = hasEscape
		}
		i += 8
	}
	for i < n {
		c := bp[i]
		if c < 0x20 || c == '"' || c == '\\' {
			return i
		}
		i++
	}
	return n
}

// scanEscapeHTMLASCIIOnly scans for characters needing HTML-safe JSON escaping, ignoring non-ASCII bytes.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), '"', '\', '<', '>', '&'. Does NOT flag bytes >= 0x80.
func scanEscapeHTMLASCIIOnly(p unsafe.Pointer, n int) int {
	i := 0
	bp := (*[1 << 30]byte)(p)
	for i+8 <= n {
		for k := 0; k < 8; k++ {
			c := bp[i+k]
			if c < 0x20 || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
				return i + k
			}
		}
		i += 8
	}
	for i < n {
		c := bp[i]
		if c < 0x20 || c == '"' || c == '\\' || c == '<' || c == '>' || c == '&' {
			return i
		}
		i++
	}
	return n
}
