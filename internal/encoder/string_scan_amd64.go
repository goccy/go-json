package encoder

import (
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

// The scan of a string for a byte to escape by AVX2, 32 bytes at a time.

//go:noescape
func scanStringAVX2(p unsafe.Pointer, n int, tables *nibbleTables) int

//go:noescape
func indexEscapeAVX2(p unsafe.Pointer, n int, tables *nibbleTables) int

// minSIMDScanLength is the length from which a string is scanned by SIMD: the setup of the registers costs
// about as much as a few words of the scalar scan.
const minSIMDScanLength = 32

// hasEscapeSIMD is whether a byte of the string may need an escape, by SIMD. The second result is false if
// the CPU doesn't support it or the string is short, and the string is not looked at.
func (e *stringEscape) hasEscapeSIMD(src unsafe.Pointer, n int) (bool, bool) {
	return scanBytesSIMD(src, n, &e.tables)
}

// scanBytesSIMD is whether a byte of the n bytes at src is in the tables, by SIMD. The second result is false
// if the CPU doesn't support it or n is small, and the bytes are not looked at.
func scanBytesSIMD(src unsafe.Pointer, n int, tables *nibbleTables) (bool, bool) {
	if !runtime.HasAVX2 || n < minSIMDScanLength {
		return false, false
	}
	return scanStringAVX2(src, n, tables) != 0, true
}

// indexEscapeSIMD returns the index of the first of the n bytes at src which may need an escape by the tables,
// or n if there is none, by SIMD. The second result is false if the CPU doesn't support it or n is small, and
// the bytes are not looked at.
func indexEscapeSIMD(src unsafe.Pointer, n int, tables *nibbleTables) (int, bool) {
	if !runtime.HasAVX2 || n < minSIMDScanLength {
		return 0, false
	}
	return indexEscapeAVX2(src, n, tables), true
}
