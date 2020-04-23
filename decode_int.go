package json

import (
	"errors"
	"math"
)

type intDecoder struct {
	op func(uintptr, int64)
}

func newIntDecoder(op func(uintptr, int64)) *intDecoder {
	return &intDecoder{op: op}
}

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
		digitValue := int64(math.Pow10(maxDigit - i - 1))
		sum += c * digitValue
	}
	if isNegative {
		return -1 * sum
	}
	return sum
}

func (d *intDecoder) decodeByte(ctx *context) ([]byte, error) {
	ctx.skipWhiteSpace()
	buf := ctx.buf
	cursor := ctx.cursor
	buflen := ctx.buflen
	switch buf[cursor] {
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		start := ctx.cursor
		cursor++
		for ; cursor < buflen; cursor++ {
			tk := int(buf[cursor])
			if int('0') <= tk && tk <= int('9') {
				continue
			}
			break
		}
		num := ctx.buf[start:cursor]
		ctx.cursor = cursor
		//fmt.Printf("number = [%s]\n", string(num))
		return num, nil
	}
	return nil, errors.New("unexpected error number")
}

func (d *intDecoder) decode(ctx *context, p uintptr) error {
	bytes, err := d.decodeByte(ctx)
	if err != nil {
		return err
	}
	d.op(p, d.parseInt(bytes))
	return nil
}
