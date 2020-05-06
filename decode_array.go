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

func (d *arrayDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	buflen := len(buf)
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '[':
			idx := 0
			for {
				cursor++
				c, err := d.valueDecoder.decode(buf, cursor, p+uintptr(idx)*d.size)
				if err != nil {
					return 0, err
				}
				cursor = c
				cursor = skipWhiteSpace(buf, cursor)
				switch buf[cursor] {
				case ']':
					cursor++
					return cursor, nil
				case ',':
					idx++
					continue
				default:
					return 0, errors.New("syntax error array")
				}
			}
		}
	}
	return 0, errors.New("unexpected error array")
}
