package encoder

import (
	"unsafe"
)

// The scan of a string for a byte to escape by AVX2, 32 bytes at a time.

func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
func xgetbv() (eax, edx uint32)

//go:noescape
func scanStringAVX2(p unsafe.Pointer, n int, tables *nibbleTables) int

// hasAVX2 is whether the CPU and the OS support AVX2.
var hasAVX2 = detectAVX2()

func detectAVX2() bool {
	maxID, _, _, _ := cpuid(0, 0)
	if maxID < 7 {
		return false
	}
	_, _, ecx1, _ := cpuid(1, 0)
	const osxsave, avx = 1 << 27, 1 << 28
	if ecx1&osxsave == 0 || ecx1&avx == 0 {
		return false
	}
	// the OS saves the YMM registers.
	if eax, _ := xgetbv(); eax&6 != 6 {
		return false
	}
	_, ebx7, _, _ := cpuid(7, 0)
	const avx2 = 1 << 5
	return ebx7&avx2 != 0
}

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
	if !hasAVX2 || n < minSIMDScanLength {
		return false, false
	}
	return scanStringAVX2(src, n, tables) != 0, true
}
