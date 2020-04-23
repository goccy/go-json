package json

import (
	"errors"
	"unsafe"
)

type stringDecoder struct {
}

func newStringDecoder() *stringDecoder {
	return &stringDecoder{}
}

func (d *stringDecoder) decode(ctx *context, p uintptr) error {
	bytes, err := d.decodeByte(ctx)
	if err != nil {
		return err
	}
	*(*string)(unsafe.Pointer(p)) = *(*string)(unsafe.Pointer(&bytes))
	return nil
}

func (d *stringDecoder) decodeByte(ctx *context) ([]byte, error) {
	ctx.skipWhiteSpace()
	buf := ctx.buf
	cursor := ctx.cursor
	buflen := ctx.buflen
	if buf[cursor] != '"' {
		return nil, errors.New("unexpected error key delimiter")
	}
	start := cursor + 1
	cursor++
	for ; cursor < buflen; cursor++ {
		tk := buf[cursor]
		if tk == '\\' {
			continue
		}
		if tk == '"' {
			break
		}
	}
	if buf[cursor] != '"' {
		return nil, errors.New("unexpected error string")
	}
	literal := buf[start:cursor]
	//fmt.Printf("string = [%s]\n", string(literal))
	cursor++
	ctx.cursor = cursor
	return literal, nil
}
