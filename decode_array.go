package json

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

func (d *arrayDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	buflen := int64(len(buf))
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
					return 0, errInvalidCharacter(buf[cursor], "array", cursor)
				}
			}
		}
	}
	return 0, errUnexpectedEndOfJSON("array", cursor)
}
