// Package jsonnum reads JSON numbers by the grammar of RFC 8259, section 6, for the encoder and the decoder of
// encoding/json and for jsontext. The callers write their own errors from where a number is invalid.
package jsonnum

import (
	"encoding/binary"
	"math/bits"
)

// Place is where Scan found a number invalid.
type Place uint8

const (
	Valid    Place = iota // the number is valid
	Start                 // the first digit, after the sign if any
	Fraction              // the first digit after the decimal point
	Exponent              // the first digit of the exponent, after its sign if any
)

// Scan returns the length of the longest JSON number which b starts with, and Valid; or, if the number is
// invalid, the position of the byte where it is, which is len(b) if b ends there, and the place of that byte.
// The byte after a valid number is not looked at: it may be anything, which the caller checks.
func Scan(b []byte) (int, Place) {
	i := 0
	if i < len(b) && b[i] == '-' {
		i++
	}
	switch {
	case i == len(b):
		return i, Start
	case b[i] == '0':
		i++
	case '1' <= b[i] && b[i] <= '9':
		i = ScanDigits(b, i+1)
	default:
		return i, Start
	}
	if i < len(b) && b[i] == '.' {
		if i++; i == len(b) || b[i] < '0' || '9' < b[i] {
			return i, Fraction
		}
		i = ScanDigits(b, i+1)
	}
	if i < len(b) && (b[i] == 'e' || b[i] == 'E') {
		if i++; i < len(b) && (b[i] == '+' || b[i] == '-') {
			i++
		}
		if i == len(b) || b[i] < '0' || '9' < b[i] {
			return i, Exponent
		}
		i = ScanDigits(b, i+1)
	}
	return i, Valid
}

// IsValid reports whether b is one JSON number, without anything before or after it.
func IsValid(b []byte) bool {
	n, p := Scan(b)
	return p == Valid && n == len(b)
}

// End returns the end of the valid JSON number which starts b, if a byte which ends it follows it in b, or -1:
// at an error, and at a number which may continue after b. It is the fast check of a scanner which reads its
// input by parts.
func End(b []byte) int {
	i := 0
	if i < len(b) && b[i] == '-' {
		i++
	}
	switch {
	case i == len(b):
		return -1
	case b[i] == '0':
		i++
	case '1' <= b[i] && b[i] <= '9':
		i = ScanDigits(b, i+1)
	default:
		return -1
	}
	if i < len(b) && b[i] == '.' {
		if i++; i == len(b) || b[i] < '0' || '9' < b[i] {
			return -1
		}
		i = ScanDigits(b, i+1)
	}
	if i < len(b) && (b[i] == 'e' || b[i] == 'E') {
		if i++; i < len(b) && (b[i] == '+' || b[i] == '-') {
			i++
		}
		if i == len(b) || b[i] < '0' || '9' < b[i] {
			return -1
		}
		i = ScanDigits(b, i+1)
	}
	if i == len(b) {
		return -1
	}
	return i
}

// ScanDigits returns the position of the first byte of b from i which is not a digit, or len(b). Long runs of
// digits, as the fractions of floats have, are looked at 8 bytes at a time.
func ScanDigits(b []byte, i int) int {
	for i+8 <= len(b) {
		// a digit XOR '0' is below 10: adding 0x76 to each byte of the low 7 bits carries into the top bit from
		// 10 on, and any byte of 0x80 or more has its top bit already.
		w := binary.LittleEndian.Uint64(b[i:]) ^ 0x3030303030303030
		if m := ((w &^ 0x8080808080808080) + 0x7676767676767676 | w) & 0x8080808080808080; m != 0 {
			return i + bits.TrailingZeros64(m)/8
		}
		i += 8
	}
	for i < len(b) && '0' <= b[i] && b[i] <= '9' {
		i++
	}
	return i
}
