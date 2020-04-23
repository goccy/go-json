package json

import (
	"errors"
	"fmt"
	"math"
)

type uintDecoder struct {
	op func(uintptr, uint64)
}

func newUintDecoder(op func(uintptr, uint64)) *uintDecoder {
	return &uintDecoder{op: op}
}

func (d *uintDecoder) parseUint(b []byte) uint64 {
	maxDigit := len(b)
	sum := uint64(0)
	for i := 0; i < maxDigit; i++ {
		c := uint64(b[i]) - 48
		digitValue := uint64(math.Pow10(maxDigit - i - 1))
		sum += c * digitValue
	}
	return sum
}

func (d *uintDecoder) decodeByte(ctx *context) ([]byte, error) {
	ctx.skipWhiteSpace()
	buf := ctx.buf
	buflen := ctx.buflen
	cursor := ctx.cursor
	switch buf[cursor] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		start := cursor
		cursor++
		for ; cursor < buflen; cursor++ {
			tk := int(buf[cursor])
			if int('0') <= tk && tk <= int('9') {
				continue
			}
			break
		}
		num := buf[start:cursor]
		fmt.Printf("number = [%s]\n", string(num))
		ctx.cursor = cursor
		return num, nil
	}
	return nil, errors.New("unexpected error number")
}

func (d *uintDecoder) decode(ctx *context, p uintptr) error {
	bytes, err := d.decodeByte(ctx)
	if err != nil {
		return err
	}
	d.op(p, d.parseUint(bytes))
	return nil
}
