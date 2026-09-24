package decoder

import (
	"bytes"
	"fmt"
	"math/bits"
	"reflect"
	"strconv"
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

func (d *stringDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	s, c, ok, err := d.decodeString(ctx, cursor)
	if err != nil {
		return 0, err
	}
	if ok {
		**(**string)(unsafe.Pointer(&p)) = s
	}
	return c, nil
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

var runeErrBytes = []byte(string(utf8.RuneError))

// The flags of a string which scanString found.
const (
	// stringEscaped is set if the string has an escape: its bytes are not the ones of the value.
	stringEscaped = 1 << iota
	// stringNonASCII is set if a byte of the string is not ASCII: the string may be invalid UTF-8.
	stringNonASCII
)

// decodeByte returns the bytes of the string at cursor, decoded in place in buf, or nil for null.
func (d *stringDecoder) decodeByte(buf []byte, cursor int64) ([]byte, int64, error) {
	literal, next, flags, err := d.scanString(buf, cursor)
	if err != nil || literal == nil {
		return literal, next, err
	}
	return decodeLiteral(literal, flags), next, nil
}

// decodeLiteral decodes the escapes of the literal in place, and replaces its invalid UTF-8.
func decodeLiteral(literal []byte, flags uint8) []byte {
	if flags&stringEscaped != 0 {
		literal = literal[:unescapeTo(unsafe.Pointer(unsafe.SliceData(literal)), literal)]
	}
	if flags&stringNonASCII != 0 && !utf8.Valid(literal) {
		literal = coerceUTF8(literal)
	}
	return literal
}

// decodeString returns the string at cursor as a value to store, and false for null.
// An escaped string which is copied out of the buffer ( see makeString ) is decoded into its copy directly.
func (d *stringDecoder) decodeString(ctx *RuntimeContext, cursor int64) (string, int64, bool, error) {
	literal, next, flags, err := d.scanString(ctx.Buf, cursor)
	if err != nil || literal == nil {
		return "", next, false, err
	}
	if flags&stringEscaped != 0 && ctx.copiesStrings() && len(literal) <= maxArenaStringSize {
		dst := ctx.reserveArena(len(literal))
		n := unescapeTo(unsafe.Pointer(unsafe.SliceData(dst)), literal)
		decoded := dst[:n]
		if flags&stringNonASCII != 0 && !utf8.Valid(decoded) {
			return string(coerceUTF8(decoded)), next, true, nil
		}
		ctx.arena = ctx.arena[:len(ctx.arena)+n]
		return unsafe.String(unsafe.SliceData(decoded), n), next, true, nil
	}
	return ctx.makeString(decodeLiteral(literal, flags), next-1), next, true, nil
}

// scanString finds the string at cursor, and returns its bytes as they are in buf, the position after it and
// its flags, or nil for null.
func (d *stringDecoder) scanString(buf []byte, cursor int64) ([]byte, int64, uint8, error) {
	for {
		switch buf[cursor] {
		case ' ', '\n', '\t', '\r':
			cursor++
		case '[':
			return nil, 0, 0, d.errUnmarshalType("array", cursor)
		case '{':
			return nil, 0, 0, d.errUnmarshalType("object", cursor)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return nil, 0, 0, d.errUnmarshalType("number", cursor)
		case '"':
			cursor++
			start := cursor
			b := (*sliceHeader)(unsafe.Pointer(&buf)).data
			buflen := int64(len(buf))
			escaped := 0
			// high accumulates the bytes of the string: it is not zero if one of them is not ASCII.
			var high uint64
			for {
				// The words with nothing to look at are skipped eight bytes at a time: a word is read
				// only where the buffer has room for it, so that its end is scanned byte by byte.
				for cursor+8 <= buflen {
					w := load64(buf, cursor)
					special := byteMask(w, '"') | byteMask(w, '\\') | hasLess(w, 0x20)
					if special != 0 {
						i := int64(bits.TrailingZeros64(special) / 8)
						high |= w & msb & (1<<(uint(i)*8) - 1)
						cursor += i
						break
					}
					high |= w & msb
					cursor += 8
				}
				c := char(b, cursor)
				switch c {
				case '\\':
					escaped++
					cursor++
					switch char(b, cursor) {
					case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
						cursor++
					case 'u':
						if cursor+5 >= buflen {
							return nil, 0, 0, errors.ErrUnexpectedEndOfJSON("escaped string", cursor)
						}
						for i := int64(1); i <= 4; i++ {
							c := char(b, cursor+i)
							if !(('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')) {
								return nil, 0, 0, errors.ErrSyntax(fmt.Sprintf("json: invalid character %c in \\u hexadecimal character escape", c), cursor+i)
							}
						}
						cursor += 5
					default:
						return nil, 0, 0, errors.ErrUnexpectedEndOfJSON("escaped string", cursor)
					}
				case '"':
					var flags uint8
					if escaped > 0 {
						flags |= stringEscaped
					}
					if high&msb != 0 {
						flags |= stringNonASCII
					}
					return buf[start:cursor], cursor + 1, flags, nil
				case nul:
					return nil, 0, 0, errors.ErrUnexpectedEndOfJSON("string", cursor)
				default:
					if c < 0x20 {
						return nil, 0, 0, errors.ErrSyntax(fmt.Sprintf("invalid character %s in string literal", quoteChar(c)), cursor+1)
					}
					high |= uint64(c)
					cursor++
				}
			}
		case 'n':
			if err := validateNull(buf, cursor); err != nil {
				return nil, 0, 0, err
			}
			cursor += 4
			return nil, cursor, 0, nil
		default:
			return nil, 0, 0, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
		}
	}
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
}

func unsafeAdd(ptr unsafe.Pointer, offset int) unsafe.Pointer {
	return unsafe.Add(ptr, offset)
}

// unescapeTo decodes the escapes of the bytes of an escaped string into out, which may be the bytes themselves,
// and returns the length of the result, which is at most the length of the bytes.
func unescapeTo(out unsafe.Pointer, buf []byte) int {
	p := (*sliceHeader)(unsafe.Pointer(&buf)).data
	end := unsafeAdd(p, len(buf))
	first := bytes.IndexByte(buf, '\\')
	if out != p {
		copy(unsafe.Slice((*byte)(out), first), buf[:first])
	}
	src := unsafeAdd(p, first)
	dst := unsafeAdd(out, first)
	for src != end {
		c := char(src, 0)
		if c == '\\' {
			escapeChar := char(src, 1)
			if escapeChar != 'u' {
				*(*byte)(dst) = unescapeMap[escapeChar]
				src = unsafeAdd(src, 2)
				dst = unsafeAdd(dst, 1)
			} else {
				v1 := hexToInt[char(src, 2)]
				v2 := hexToInt[char(src, 3)]
				v3 := hexToInt[char(src, 4)]
				v4 := hexToInt[char(src, 5)]
				code := rune((v1 << 12) | (v2 << 8) | (v3 << 4) | v4)
				if code >= 0xd800 && code < 0xdc00 && uintptr(unsafeAdd(src, 11)) < uintptr(end) {
					if char(src, 6) == '\\' && char(src, 7) == 'u' {
						v1 := hexToInt[char(src, 8)]
						v2 := hexToInt[char(src, 9)]
						v3 := hexToInt[char(src, 10)]
						v4 := hexToInt[char(src, 11)]
						lo := rune((v1 << 12) | (v2 << 8) | (v3 << 4) | v4)
						if lo >= 0xdc00 && lo < 0xe000 {
							code = (code-0xd800)<<10 | (lo - 0xdc00) + 0x10000
							src = unsafeAdd(src, 6)
						}
					}
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
		} else {
			*(*byte)(dst) = c
			src = unsafeAdd(src, 1)
			dst = unsafeAdd(dst, 1)
		}
	}
	return int(uintptr(dst) - uintptr(out))
}

// coerceUTF8 returns a copy of the literal in which every byte of an invalid UTF-8 sequence
// is replaced by utf8.RuneError, as encoding/json does.
func coerceUTF8(literal []byte) []byte {
	out := make([]byte, 0, len(literal)+2*utf8.UTFMax)
	for i := 0; i < len(literal); {
		r, size := utf8.DecodeRune(literal[i:])
		if r == utf8.RuneError && size == 1 {
			out = append(out, runeErrBytes...)
		} else {
			out = append(out, literal[i:i+size]...)
		}
		i += size
	}
	return out
}

// hasLess returns the word whose byte is 0x80 where the byte of w is less than n, which is at most 128,
// exactly for the lowest such byte: a byte above it may be marked wrongly by a borrow.
func hasLess(w, n uint64) uint64 {
	return (w - n*lsb) & ^w & msb
}

// quoteChar formats c as a quoted character literal, as encoding/json does in its errors.
func quoteChar(c byte) string {
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}
