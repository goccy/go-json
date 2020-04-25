package json

import (
	"errors"
)

type arrayDecoder struct {
	elemType     *rtype
	size         uintptr
	valueDecoder decoder
	alen         int
}

func newArrayDecoder(dec decoder, elemType *rtype, alen int) *arrayDecoder {
	return &arrayDecoder{
		valueDecoder: dec,
		elemType:     elemType,
		size:         elemType.Size(),
		alen:         alen,
	}
}

func (d *arrayDecoder) decode(ctx *context, p uintptr) error {
	buf := ctx.buf
	buflen := ctx.buflen
	cursor := ctx.cursor
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '[':
			idx := 0
			for {
				ctx.cursor = cursor + 1
				if err := d.valueDecoder.decode(ctx, p+uintptr(idx)*d.size); err != nil {
					return err
				}
				cursor = ctx.skipWhiteSpace()
				switch buf[cursor] {
				case ']':
					return nil
				case ',':
					idx++
					continue
				default:
					return errors.New("syntax error array")
				}
			}
		}
	}
	return errors.New("unexpected error array")
}
