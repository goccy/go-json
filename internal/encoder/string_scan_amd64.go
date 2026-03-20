package encoder

import "unsafe"

// scanEscapeBasic scans n bytes starting at p for characters that need JSON escaping.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), non-ASCII (>= 0x80), '"', '\'.
// Implemented in string_scan_amd64.s using AVX2/SSE2.
//
//go:noescape
func scanEscapeBasic(p unsafe.Pointer, n int) int

// scanEscapeHTML scans n bytes starting at p for characters that need HTML-safe JSON escaping.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), non-ASCII (>= 0x80), '"', '\', '<', '>', '&'.
// Implemented in string_scan_amd64.s using AVX2/SSE2.
//
//go:noescape
func scanEscapeHTML(p unsafe.Pointer, n int) int

// scanEscapeBasicASCIIOnly scans for characters needing JSON escaping, ignoring non-ASCII bytes.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), '"', '\'. Does NOT flag bytes >= 0x80.
// Uses XOR-0x80 unsigned comparison trick. Ideal for non-NormalizeUTF8 paths
// where high bytes are passed through as-is.
// Implemented in string_scan_amd64.s using AVX2/SSE2.
//
//go:noescape
func scanEscapeBasicASCIIOnly(p unsafe.Pointer, n int) int

// scanEscapeHTMLASCIIOnly scans for characters needing HTML-safe JSON escaping, ignoring non-ASCII bytes.
// Returns the index of the first byte needing escape, or n if none found.
// Detects: control chars (< 0x20), '"', '\', '<', '>', '&'. Does NOT flag bytes >= 0x80.
// Uses XOR-0x80 unsigned comparison trick.
// Implemented in string_scan_amd64.s using AVX2/SSE2.
//
//go:noescape
func scanEscapeHTMLASCIIOnly(p unsafe.Pointer, n int) int

// stringptr returns the underlying data pointer of a string.
//
//go:nosplit
func stringptr(s string) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&s))
}
