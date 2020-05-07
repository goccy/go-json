package json

import (
	"errors"
)

type intDecoder struct {
	op func(uintptr, int64)
}

func newIntDecoder(op func(uintptr, int64)) *intDecoder {
	return &intDecoder{op: op}
}

var (
	pow10i64 = [...]int64{
		1e00, 1e01, 1e02, 1e03, 1e04, 1e05, 1e06, 1e07, 1e08, 1e09,
		1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18,
	}
)

func (d *intDecoder) parseInt(b []byte) int64 {
	isNegative := false
	if b[0] == '-' {
		b = b[1:]
		isNegative = true
	}
	maxDigit := len(b)
	sum := int64(0)
	for i := 0; i < maxDigit; i++ {
		c := int64(b[i]) - 48
		digitValue := pow10i64[maxDigit-i-1]
		sum += c * digitValue
	}
	if isNegative {
		return -1 * sum
	}
	return sum
}

var (
	numTable = [256]bool{
		'0': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'9': true,
	}
)

func (d *intDecoder) decodeByte(buf []byte, cursor int) ([]byte, int, error) {
	for {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			cursor++
			continue
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := cursor
			cursor++
		LOOP:
			if numTable[buf[cursor]] {
				cursor++
				goto LOOP
			}
			num := buf[start:cursor]
			return num, cursor, nil
		default:
			goto ERROR
		}
	}
ERROR:
	return nil, 0, errors.New("unexpected error number")
}

func (d *intDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	bytes, c, err := d.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	cursor = c
	d.op(p, d.parseInt(bytes))
	return cursor, nil
}
