package decoder

import (
	"fmt"
	"reflect"
	"unicode/utf8"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type stringDecoder struct {
	structName string
	fieldName  string
}

func newStringDecoder(structName, fieldName string) *stringDecoder {
	return &stringDecoder{
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *stringDecoder) errUnmarshalType(typeName string, offset int64) *errors.UnmarshalTypeError {
	return &errors.UnmarshalTypeError{
		Value:  typeName,
		Type:   reflect.TypeOf(""),
		Offset: offset,
		Struct: d.structName,
		Field:  d.fieldName,
	}
}

func (d *stringDecoder) DecodeStream(s *Stream, depth int64, p unsafe.Pointer) error {
	bytes, err := d.decodeStreamByte(s)
	if err != nil {
		return err
	}
	if bytes == nil {
		return nil
	}
	**(**string)(unsafe.Pointer(&p)) = *(*string)(unsafe.Pointer(&bytes))
	s.reset()
	return nil
}

func (d *stringDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	bytes, c, err := d.decodeByte(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}
	if bytes == nil {
		return c, nil
	}
	cursor = c
	**(**string)(unsafe.Pointer(&p)) = *(*string)(unsafe.Pointer(&bytes))
	return cursor, nil
}

func (d *stringDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	bytes, c, err := d.decodeByte(ctx.Buf, cursor)
	if err != nil {
		return nil, 0, err
	}
	if bytes == nil {
		return [][]byte{nullbytes}, c, nil
	}
	return [][]byte{bytes}, c, nil
}

var (
	hexToInt = [256]int{
		'0': 0,
		'1': 1,
		'2': 2,
		'3': 3,
		'4': 4,
		'5': 5,
		'6': 6,
		'7': 7,
		'8': 8,
		'9': 9,
		'A': 10,
		'B': 11,
		'C': 12,
		'D': 13,
		'E': 14,
		'F': 15,
		'a': 10,
		'b': 11,
		'c': 12,
		'd': 13,
		'e': 14,
		'f': 15,
	}
)

func unicodeToRune(code []byte) rune {
	var r rune
	for i := 0; i < len(code); i++ {
		r = r*16 + rune(hexToInt[code[i]])
	}
	return r
}

var isHex = [256]int8{
	'0': 1,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'A': 10,
	'B': 11,
	'C': 12,
	'D': 13,
	'E': 14,
	'F': 15,
	'a': 10,
	'b': 11,
	'c': 12,
	'd': 13,
	'e': 14,
	'f': 15,
}

var utf8First = [256]uint8{
	0xC2: 0x02, 0xC3: 0x02, 0xC4: 0x02, 0xC5: 0x02, 0xC6: 0x02, 0xC7: 0x02, 0xC8: 0x02, 0xC9: 0x02, 0xCA: 0x02, 0xCB: 0x02, 0xCC: 0x02, 0xCD: 0x02, 0xCE: 0x02, 0xCF: 0x02, 0xD0: 0x02, 0xD1: 0x02, 0xD2: 0x02, 0xD3: 0x02, 0xD4: 0x02, 0xD5: 0x02, 0xD6: 0x02, 0xD7: 0x02, 0xD8: 0x02, 0xD9: 0x02, 0xDA: 0x02, 0xDB: 0x02, 0xDC: 0x02, 0xDD: 0x02, 0xDE: 0x02, 0xDF: 0x02,
	0xE0: 0x13,
	0xE1: 0x03, 0xE2: 0x03, 0xE3: 0x03, 0xE4: 0x03, 0xE5: 0x03, 0xE6: 0x03, 0xE7: 0x03, 0xE8: 0x03, 0xE9: 0x03, 0xEA: 0x03, 0xEB: 0x03, 0xEC: 0x03, 0xEE: 0x03, 0xEF: 0x3,
	0xED: 0x23,
	0xF0: 0x34,
	0xF1: 0x04, 0xF2: 0x04, 0xF3: 0x04,
	0xF4: 0x44,
}

var utf8AcceptRanges = [16]struct{ lo, hi uint8 }{
	0: {0x80, 0xBF},
	1: {0xA0, 0xBF},
	2: {0x80, 0x9F},
	3: {0x90, 0xBF},
	4: {0x80, 0x8F},
}

var unescapeMap = [256]byte{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'u':  'u',
}

const (
	inStringInvalidUTF8 = 0
	inStringASCII       = 1
	inStringSentinel    = 2
	inStringStartEscape = 3
	inStringEnd         = 4
	inStringStartMB     = 5
)

var inStringTypes [256]uint8

func init() {
	for i := range inStringTypes {
		inStringTypes[i] = inStringInvalidUTF8
	}
	for i := 0; i < 0x80; i++ {
		inStringTypes[i] = inStringASCII
	}
	inStringTypes[nul] = inStringSentinel
	inStringTypes['\\'] = inStringStartEscape
	inStringTypes['"'] = inStringEnd
	for i := 0xC2; i <= 0xF4; i++ {
		inStringTypes[i] = inStringStartMB
	}
}

func stringBytes(s *Stream) ([]byte, int64, error) {
	_, cursor, p := s.stat()
	cursor++ // skip double quote char

	start := cursor
	dst := cursor
	inplace := true
	first := int64(-1)
	for {
		c := char(p, cursor)
		if t := inStringTypes[c]; t == inStringASCII {
			cursor++
			dst++
			continue
		} else if t == inStringStartMB {
			x := utf8First[c]
			sz := int64(x & 7)
			if s.syncBufptr(s.requires(cursor, sz), &p) < 0 {
				goto RuneError
			}
			accept := utf8AcceptRanges[x>>4]
			c1 := char(p, cursor+1)
			if c1 < accept.lo || accept.hi < c1 {
				goto RuneError
			}
			if sz > 2 {
				c2 := char(p, cursor+2)
				if c2 < 0x80 || 0xBF < c2 {
					goto RuneError
				}
			}
			if sz > 3 {
				c3 := char(p, cursor+3)
				if c3 < 0x80 || 0xBF < c3 {
					goto RuneError
				}
			}
			cursor += sz
			dst += sz
			continue
		} else if t == inStringStartEscape {
			if first < 0 {
				first = cursor
			}
			cursor++
			if s.syncBufptr(s.requires(cursor, 1), &p) < 0 {
				goto ERROR
			}
			ec := char(p, cursor)
			if unescapeMap[ec] == 0 {
				return nil, cursor, errors.ErrInvalidCharacter(char(p, cursor), "in string escape code", cursor)
			}
			if ec != 'u' {
				cursor++
				dst++
				continue
			}
			if s.syncBufptr(s.requires(cursor, 5), &p) < 0 {
				goto ERROR
			}
			c1, c2, c3, c4 := char4(p, cursor+1)
			if o := checkHex(c1, c2, c3, c4); o > 0 {
				return nil, cursor + o, errors.ErrSyntax(fmt.Sprintf("json: invalid character %c in \\u hexadecimal character escape", char(p, cursor+o)), cursor+o)
			}
			r := decodeHexRune(c1, c2, c3, c4)
			*ptrUint16(p, cursor+1) = uint16(r)
		NextUnicode:
			if 0xD800 <= r && r < 0xE000 {
				const runeError = 65533
				if s.syncBufptr(s.requires(cursor, 5+6), &p) >= 0 && char(p, cursor+5) == '\\' && char(p, cursor+6) == 'u' {
					cursor2 := cursor + 6
					c1, c2, c3, c4 := char4(p, cursor2+1)
					if o := checkHex(c1, c2, c3, c4); o > 0 {
						return nil, cursor2 + o, errors.ErrSyntax(fmt.Sprintf("json: invalid character %c in \\u hexadecimal character escape", char(p, cursor2+o)), cursor2+o)
					}
					r2 := decodeHexRune(c1, c2, c3, c4)
					*ptrUint16(p, cursor2+1) = uint16(r2)
					if r2 < 0xDC00 || 0xE000 <= r2 {
						*ptrUint16(p, cursor+1) = runeError
						dst += 3
						cursor = cursor2
						r = r2
						goto NextUnicode
					}
					dst += 4
					cursor = cursor2 + 5
				} else {
					*ptrUint16(p, cursor+1) = runeError
					dst += 3
					cursor += 5
				}
			} else {
				cursor += 5
				dst += runeLen(r)
			}
			continue
		} else if t == inStringEnd {
			if first < 0 {
				return s.buf[start:cursor], cursor + 1, nil
			}
			if inplace {
				src := unsafeAdd(p, int(first))
				unescapeString(src, src)
				return s.buf[start:dst], cursor + 1, nil
			}
			src := unsafeAdd(p, int(start))
			b := make([]byte, dst-start+1) // MEMO: 最後に1バイト無いと unescapeString の中で最後に unsafe.Pointer が invalid な領域を指す
			data := (*sliceHeader)(unsafe.Pointer(&b)).data
			unescapeString(src, data)
			return b[:len(b)-1], cursor + 1, nil
		} else if t == inStringSentinel {
			if s.read() {
				p = s.bufptr()
				continue
			}
			goto ERROR
		}
	RuneError:
		if first < 0 {
			first = cursor
		}
		*(*byte)(unsafeAdd(p, int(cursor))) = nul
		cursor++
		dst += 3
		if cursor < dst {
			inplace = false
		}
	}
ERROR:
	return nil, s.length, errors.ErrUnexpectedEndOfJSON("string", s.offset+s.length)
}

func (d *stringDecoder) decodeStreamByte(s *Stream) ([]byte, error) {
	for {
		switch s.char() {
		case ' ', '\n', '\t', '\r':
			s.cursor++
			continue
		case '[':
			return nil, d.errUnmarshalType("array", s.totalOffset())
		case '{':
			return nil, d.errUnmarshalType("object", s.totalOffset())
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return nil, d.errUnmarshalType("number", s.totalOffset())
		case '"':
			b, cursor, err := stringBytes(s)
			s.cursor = cursor
			if err != nil {
				return nil, err
			}
			return b, nil
		case 'n':
			if err := nullBytes(s); err != nil {
				return nil, err
			}
			return nil, nil
		case nul:
			if s.read() {
				continue
			}
		}
		break
	}
	return nil, errors.ErrInvalidBeginningOfValue(s.char(), s.totalOffset())
}

func (d *stringDecoder) decodeByte(buf []byte, cursor int64) ([]byte, int64, error) {
	for {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			cursor++
		case '[':
			return nil, cursor, d.errUnmarshalType("array", cursor)
		case '{':
			return nil, cursor, d.errUnmarshalType("object", cursor)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return nil, cursor, d.errUnmarshalType("number", cursor)
		case '"':
			s := bytesStream{buf: buf, length: int64(len(buf))}
			cursor++
			p := (*sliceHeader)(unsafe.Pointer(&buf)).data

			start := cursor
			dst := cursor
			inplace := true
			first := int64(-1)
			for {
				c := char(p, cursor)
				if t := inStringTypes[c]; t == inStringASCII {
					cursor++
					dst++
					continue
				} else if t == inStringStartMB {
					x := utf8First[c]
					sz := int64(x & 7)
					if s.syncBufptr(s.requires(cursor, sz), &p) < 0 {
						goto RuneError
					}
					accept := utf8AcceptRanges[x>>4]
					c1 := char(p, cursor+1)
					if c1 < accept.lo || accept.hi < c1 {
						goto RuneError
					}
					if sz > 2 {
						c2 := char(p, cursor+2)
						if c2 < 0x80 || 0xBF < c2 {
							goto RuneError
						}
					}
					if sz > 3 {
						c3 := char(p, cursor+3)
						if c3 < 0x80 || 0xBF < c3 {
							goto RuneError
						}
					}
					cursor += sz
					dst += sz
					continue
				} else if t == inStringStartEscape {
					if first < 0 {
						first = cursor
					}
					cursor++
					if s.syncBufptr(s.requires(cursor, 1), &p) < 0 {
						goto ERROR
					}
					ec := char(p, cursor)
					if unescapeMap[ec] == 0 {
						return nil, cursor, errors.ErrInvalidCharacter(char(p, cursor), "in string escape code", cursor)
					}
					if ec != 'u' {
						cursor++
						dst++
						continue
					}
					if s.syncBufptr(s.requires(cursor, 5), &p) < 0 {
						goto ERROR
					}
					c1, c2, c3, c4 := char4(p, cursor+1)
					if o := checkHex(c1, c2, c3, c4); o > 0 {
						return nil, cursor + o, errors.ErrSyntax(fmt.Sprintf("json: invalid character %c in \\u hexadecimal character escape", char(p, cursor+o)), cursor+o)
					}
					r := decodeHexRune(c1, c2, c3, c4)
					*ptrUint16(p, cursor+1) = uint16(r)
				NextUnicode:
					if 0xD800 <= r && r < 0xE000 {
						const runeError = 65533
						if s.syncBufptr(s.requires(cursor, 5+6), &p) >= 0 && char(p, cursor+5) == '\\' && char(p, cursor+6) == 'u' {
							cursor2 := cursor + 6
							c1, c2, c3, c4 := char4(p, cursor2+1)
							if o := checkHex(c1, c2, c3, c4); o > 0 {
								return nil, cursor2 + o, errors.ErrSyntax(fmt.Sprintf("json: invalid character %c in \\u hexadecimal character escape", char(p, cursor2+o)), cursor2+o)
							}
							r2 := decodeHexRune(c1, c2, c3, c4)
							*ptrUint16(p, cursor2+1) = uint16(r2)
							if r2 < 0xDC00 || 0xE000 <= r2 {
								*ptrUint16(p, cursor+1) = runeError
								dst += 3
								cursor = cursor2
								r = r2
								goto NextUnicode
							}
							dst += 4
							cursor = cursor2 + 5
						} else {
							*ptrUint16(p, cursor+1) = runeError
							dst += 3
							cursor += 5
						}
					} else {
						cursor += 5
						dst += runeLen(r)
					}
					continue
				} else if t == inStringEnd {
					if first < 0 {
						return s.buf[start:cursor], cursor + 1, nil
					}
					if inplace {
						src := unsafeAdd(p, int(first))
						unescapeString(src, src)
						return s.buf[start:dst], cursor + 1, nil
					}
					src := unsafeAdd(p, int(start))
					b := make([]byte, dst-start+1) // MEMO: 最後に1バイト無いと unescapeString の中で最後に unsafe.Pointer が invalid な領域を指す
					data := (*sliceHeader)(unsafe.Pointer(&b)).data
					unescapeString(src, data)
					return b[:len(b)-1], cursor + 1, nil
				} else if t == inStringSentinel {
					if s.read() {
						p = s.bufptr()
						continue
					}
					goto ERROR
				}
			RuneError:
				if first < 0 {
					first = cursor
				}
				*(*byte)(unsafeAdd(p, int(cursor))) = nul
				cursor++
				dst += 3
				if cursor < dst {
					inplace = false
				}
			}
		ERROR:
			return nil, s.length, errors.ErrUnexpectedEndOfJSON("string", s.offset+s.length)
		case nul:
			return nil, cursor, errors.ErrUnexpectedEndOfJSON("string", cursor)
		case 'n':
			if err := validateNull(buf, cursor); err != nil {
				return nil, cursor, err
			}
			return nil, cursor + 4, nil
		default:
			return nil, cursor, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
		}
	}
}

func unsafeAdd(ptr unsafe.Pointer, offset int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + uintptr(offset))
}

func unescapeString(src, dst unsafe.Pointer) {
	for {
		c := char(src, 0)
		switch c {
		case '"':
			return
		case '\\':
			escapeChar := char(src, 1)
			if escapeChar != 'u' {
				*(*byte)(dst) = unescapeMap[escapeChar]
				src = unsafeAdd(src, 2)
				dst = unsafeAdd(dst, 1)
			} else {
				code := rune(*ptrUint16(src, 2))
				if code >= 0xD800 && code < 0xDC00 {
					lo := rune(*ptrUint16(src, 8))
					code = (code-0xD800)<<10 | (lo - 0xDC00) + 0x10000
					src = unsafeAdd(src, 6)
				}
				var b [utf8.UTFMax]byte
				n := utf8.EncodeRune(b[:], code)
				switch n {
				case 4:
					*(*byte)(unsafeAdd(dst, 3)) = b[3]
					fallthrough
				case 3:
					*(*byte)(unsafeAdd(dst, 2)) = b[2]
					fallthrough
				case 2:
					*(*byte)(unsafeAdd(dst, 1)) = b[1]
					fallthrough
				case 1:
					*(*byte)(unsafeAdd(dst, 0)) = b[0]
				}
				src = unsafeAdd(src, 6)
				dst = unsafeAdd(dst, n)
			}
		case nul:
			*(*byte)(unsafeAdd(dst, 0)) = 0xEF
			*(*byte)(unsafeAdd(dst, 1)) = 0xBF
			*(*byte)(unsafeAdd(dst, 2)) = 0xBD
			src = unsafeAdd(src, 1)
			dst = unsafeAdd(dst, 3)
		default:
			*(*byte)(dst) = c
			src = unsafeAdd(src, 1)
			dst = unsafeAdd(dst, 1)
		}
	}
}

func char4(p unsafe.Pointer, offset int64) (byte, byte, byte, byte) {
	return char(p, offset), char(p, offset+1), char(p, offset+2), char(p, offset+3)
}

func checkHex(v1, v2, v3, v4 byte) int64 {
	if isHex[v1] == 0 {
		return 1
	}
	if isHex[v2] == 0 {
		return 2
	}
	if isHex[v3] == 0 {
		return 3
	}
	if isHex[v4] == 0 {
		return 4
	}
	return 0
}

func decodeHexRune(v1, v2, v3, v4 byte) rune {
	return rune(hexToInt[v1]<<12 | hexToInt[v2]<<8 | hexToInt[v3]<<4 | hexToInt[v4])
}

func runeLen(r rune) int64 {
	if r <= 127 {
		return 1
	} else if r <= 2047 {
		return 2
	} else {
		return 3
	}
}

type bytesStream struct {
	buf    []byte
	length int64
	offset int64
}

func (b *bytesStream) read() bool {
	return false
}

func (b *bytesStream) requires(cursor, n int64) int {
	if cursor+n >= b.length {
		return -1
	}
	return 0
}

func (b *bytesStream) syncBufptr(r int, p *unsafe.Pointer) int {
	return r
}

func (b *bytesStream) bufptr() unsafe.Pointer {
	panic("unreachable")
}
