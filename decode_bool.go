package json

import (
	"errors"
	"unsafe"
)

type boolDecoder struct{}

func newBoolDecoder() *boolDecoder {
	return &boolDecoder{}
}

func (d *boolDecoder) decode(ctx *context, p uintptr) error {
	ctx.skipWhiteSpace()
	buf := ctx.buf
	cursor := ctx.cursor
	switch buf[cursor] {
	case 't':
		if cursor+3 >= ctx.buflen {
			return errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+1] != 'r' {
			return errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+2] != 'u' {
			return errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+3] != 'e' {
			return errors.New("unexpected error. invalid bool character")
		}
		ctx.cursor += 4
		*(*bool)(unsafe.Pointer(p)) = true
	case 'f':
		if cursor+4 >= ctx.buflen {
			return errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+1] != 'a' {
			return errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+2] != 'l' {
			return errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+3] != 's' {
			return errors.New("unexpected error. invalid bool character")
		}
		if buf[cursor+4] != 'e' {
			return errors.New("unexpected error. invalid bool character")
		}
		ctx.cursor += 5
		*(*bool)(unsafe.Pointer(p)) = false
	}
	return nil
}
