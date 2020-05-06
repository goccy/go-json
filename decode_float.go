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

func (d *floatDecoder) decodeByte(buf []byte, cursor int) ([]byte, int, error) {
	buflen := len(buf)
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
			num := buf[start:cursor]
			return num, cursor, nil
		}
	}
	return nil, 0, errors.New("unexpected error number")
}

func (d *floatDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	bytes, c, err := d.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	cursor = c
	s := *(*string)(unsafe.Pointer(&bytes))
	f64, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	d.op(p, f64)
	return cursor, nil
}
