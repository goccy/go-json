package json

import (
	"strconv"
	"unsafe"
)

type numberDecoder struct {
	stringDecoder *stringDecoder
	op            func(unsafe.Pointer, Number)
	structName    string
	fieldName     string
}

func newNumberDecoder(structName, fieldName string, op func(unsafe.Pointer, Number)) *numberDecoder {
	return &numberDecoder{
		stringDecoder: newStringDecoder(structName, fieldName),
		op:            op,
		structName:    structName,
		fieldName:     fieldName,
	}
}

func (d *numberDecoder) decodeStream(s *stream, depth int64, p unsafe.Pointer) error {
	bytes, err := d.decodeStreamByte(s)
	if err != nil {
		return err
	}
	if _, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&bytes)), 64); err != nil {
		return errSyntax(err.Error(), s.totalOffset())
	}
	d.op(p, Number(string(bytes)))
	s.reset()
	return nil
}

func (d *numberDecoder) decode(buf []byte, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	bytes, c, err := d.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	if _, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&bytes)), 64); err != nil {
		return 0, errSyntax(err.Error(), c)
	}
	cursor = c
	s := *(*string)(unsafe.Pointer(&bytes))
	d.op(p, Number(s))
	return cursor, nil
}

func (d *numberDecoder) decodeStreamByte(s *stream) ([]byte, error) {
	for {
		switch s.char() {
		case ' ', '\n', '\t', '\r':
			s.cursor++
			continue
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return floatBytes(s), nil
		case 'n':
			if err := nullBytes(s); err != nil {
				return nil, err
			}
			return nil, nil
		case '"':
			return d.stringDecoder.decodeStreamByte(s)
		case nul:
			if s.read() {
				continue
			}
			goto ERROR
		default:
			goto ERROR
		}
	}
ERROR:
	return nil, errUnexpectedEndOfJSON("json.Number", s.totalOffset())
}

func (d *numberDecoder) decodeByte(buf []byte, cursor int64) ([]byte, int64, error) {
	buflen := int64(len(buf))
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := cursor
			cursor++
			for ; cursor < buflen; cursor++ {
				if floatTable[buf[cursor]] {
					continue
				}
				break
			}
			num := buf[start:cursor]
			return num, cursor, nil
		case 'n':
			if cursor+3 >= buflen {
				return nil, 0, errUnexpectedEndOfJSON("null", cursor)
			}
			if buf[cursor+1] != 'u' {
				return nil, 0, errInvalidCharacter(buf[cursor+1], "null", cursor)
			}
			if buf[cursor+2] != 'l' {
				return nil, 0, errInvalidCharacter(buf[cursor+2], "null", cursor)
			}
			if buf[cursor+3] != 'l' {
				return nil, 0, errInvalidCharacter(buf[cursor+3], "null", cursor)
			}
			cursor += 4
			return nil, cursor, nil
		case '"':
			return d.stringDecoder.decodeByte(buf, cursor)
		default:
			return nil, 0, errUnexpectedEndOfJSON("json.Number", cursor)
		}
	}
	return nil, 0, errUnexpectedEndOfJSON("json.Number", cursor)
}
