package encoder

import "unsafe"

// scanEscapeBasic scans n bytes starting at p for characters that need JSON escaping.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), non-ASCII (>= 0x80), '"', '\'.
// Implemented in string_scan_amd64.s using SSE2.
//
//go:noescape
func scanEscapeBasic(p unsafe.Pointer, n int) int

// scanEscapeHTML scans n bytes starting at p for characters that need HTML-safe JSON escaping.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), non-ASCII (>= 0x80), '"', '\', '<', '>', '&'.
// Implemented in string_scan_amd64.s using SSE2.
//
//go:noescape
func scanEscapeHTML(p unsafe.Pointer, n int) int

// stringptr returns the underlying data pointer of a string.
//
//go:nosplit
func stringptr(s string) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&s))
}
