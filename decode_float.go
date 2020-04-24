package json

import (
	"errors"
	"strconv"
	"unsafe"
)

type floatDecoder struct {
	op func(uintptr, float64)
}

func newFloatDecoder(op func(uintptr, float64)) *floatDecoder {
	return &floatDecoder{op: op}
}

func (d *floatDecoder) decodeByte(ctx *context) ([]byte, error) {
	buf := ctx.buf
	cursor := ctx.cursor
	buflen := ctx.buflen
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := cursor
			cursor++
			for ; cursor < buflen; cursor++ {
				tk := int(buf[cursor])
				if (int('0') <= tk && tk <= int('9')) || tk == '.' || tk == 'e' || tk == 'E' {
					continue
				}
				break
			}
			num := ctx.buf[start:cursor]
			ctx.cursor = cursor
			return num, nil
		}
	}
	return nil, errors.New("unexpected error number")
}

func (d *floatDecoder) decode(ctx *context, p uintptr) error {
	bytes, err := d.decodeByte(ctx)
	if err != nil {
		return err
	}
	s := *(*string)(unsafe.Pointer(&bytes))
	f64, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	d.op(p, f64)
	return nil
}
