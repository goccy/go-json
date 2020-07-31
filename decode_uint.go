package json

type uintDecoder struct {
	op func(uintptr, uint64)
}

func newUintDecoder(op func(uintptr, uint64)) *uintDecoder {
	return &uintDecoder{op: op}
}

var pow10u64 = [...]uint64{
	1e00, 1e01, 1e02, 1e03, 1e04, 1e05, 1e06, 1e07, 1e08, 1e09,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
}

func (d *uintDecoder) parseUint(b []byte) uint64 {
	maxDigit := len(b)
	sum := uint64(0)
	for i := 0; i < maxDigit; i++ {
		c := uint64(b[i]) - 48
		digitValue := pow10u64[maxDigit-i-1]
		sum += c * digitValue
	}
	return sum
}

func (d *uintDecoder) decodeStreamByte(s *stream) ([]byte, error) {
	for {
		switch s.char() {
		case ' ', '\n', '\t', '\r':
			s.cursor++
			continue
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := s.cursor
			for {
				s.cursor++
				if numTable[s.char()] {
					continue
				} else if s.char() == nul {
					if s.read() {
						s.cursor-- // for retry current character
						continue
					}
				}
				break
			}
			num := s.buf[start:s.cursor]
			return num, nil
		case nul:
			if s.read() {
				continue
			}
		}
		break
	}
	return nil, errUnexpectedEndOfJSON("number(unsigned integer)", s.totalOffset())
}

func (d *uintDecoder) decodeByte(buf []byte, cursor int64) ([]byte, int64, error) {
	buflen := int64(len(buf))
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			continue
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
			return num, cursor, nil
		default:
			return nil, 0, errInvalidCharacter(buf[cursor], "number(unsigned integer)", cursor)
		}
	}
	return nil, 0, errUnexpectedEndOfJSON("number(unsigned integer)", cursor)
}

func (d *uintDecoder) decodeStream(s *stream, p uintptr) error {
	bytes, err := d.decodeStreamByte(s)
	if err != nil {
		return err
	}
	d.op(p, d.parseUint(bytes))
	return nil
}

func (d *uintDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	bytes, c, err := d.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	cursor = c
	d.op(p, d.parseUint(bytes))
	return cursor, nil
}
