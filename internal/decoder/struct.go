package decoder

import (
	"fmt"
	"math/bits"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type structFieldSet struct {
	dec         Decoder
	offset      uintptr
	isTaggedKey bool
	fieldIdx    int
	key         string
	keyLen      int64
	err         error
}

type structDecoder struct {
	// fields are the fields of the struct in their order, by their exact keys.
	fields []*structFieldSet
	keys   *structKeys
	// fieldUniqueNameNum is the number of the fields which differ by their folded keys: under FirstWinOption,
	// the rest of an object is skipped when every one of them was decoded.
	fieldUniqueNameNum int
	stringDecoder      *stringDecoder
	structName         string
	fieldName          string
}

func newStructDecoder(structName, fieldName string) *structDecoder {
	return &structDecoder{
		keys:          newStructKeys(nil),
		stringDecoder: newStringDecoder(structName, fieldName),
		structName:    structName,
		fieldName:     fieldName,
	}
}

// setFields sets the fields of the struct, which are in the order of the struct, and makes their tables.
func (d *structDecoder) setFields(fields []*structFieldSet) {
	d.fields = fields
	d.keys = newStructKeys(fields)
	indexByFolded := map[string]int{}
	for _, field := range fields {
		folded := string(appendFoldedKey(nil, []byte(field.key)))
		idx, exists := indexByFolded[folded]
		if !exists {
			idx = len(indexByFolded)
			indexByFolded[folded] = idx
		}
		field.fieldIdx = idx
	}
	d.fieldUniqueNameNum = len(indexByFolded)
}

func (d *structDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}
	buflen := int64(len(buf))
	cursor = skipWhiteSpace(buf, cursor)
	b := (*sliceHeader)(unsafe.Pointer(&buf)).data
	switch char(b, cursor) {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 4
		return cursor, nil
	case '{':
	default:
		return 0, errors.ErrInvalidBeginningOfValue(char(b, cursor), cursor)
	}
	cursor++
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == '}' {
		cursor++
		return cursor, nil
	}
	var (
		seenFields   map[int]struct{}
		seenFieldNum int
	)
	firstWin := (ctx.Option.Flags & FirstWinOption) != 0
	if firstWin {
		seenFields = make(map[int]struct{}, d.fieldUniqueNameNum)
	}
	disallowUnknownFields := (ctx.Option.Flags & DisallowUnknownFieldsOption) != 0
	for {
		cursor = skipWhiteSpace(buf, cursor)
		if char(b, cursor) != '"' {
			return 0, errors.ErrInvalidBeginningOfValue(char(b, cursor), cursor)
		}
		var (
			field *structFieldSet
			key   []byte
			c     int64
		)
		// A key of ASCII without an escape which ends within 16 bytes is found by the words of the buffer,
		// which may be read up to its capacity: the nul byte at its end stops a key before it. Any other key,
		// and a key near the end of a buffer which has no room after it, is scanned as a string.
		if start := cursor + 1; start+16 <= int64(cap(buf)) {
			w0 := load64(buf, start)
			n := keyLengthInWord(w0)
			var w1 uint64
			if n == 8 {
				w1 = load64(buf, start+8)
				n = 8 + keyLengthInWord(w1)
			}
			if n < 16 && buf[start+int64(n)] == '"' {
				key = buf[start : start+int64(n)]
				// the words of the key, as keyWords makes them
				if n < 8 {
					w0 &= 1<<(8*uint(n)) - 1
					w1 = 0
				} else {
					w1 = load64(buf, start+int64(n)-8)
				}
				if keys := d.keys; keys.hasLength(n) {
					if n < 8 {
						// the first entry of the folded key, looked at here: most keys are found in it.
						fw := foldASCIIWord(w0)
						e := &keys.entries[keys.index(fw, 0, n)]
						if e.n == n && e.w0 == fw && e.w1 == 0 && len(e.fields) == 1 {
							field = e.fields[0]
						} else if e.fields != nil {
							field = keys.findASCII(key, w0, w1)
						}
					} else {
						field = keys.findASCII(key, w0, w1)
					}
				}
				c = start + int64(n) + 1
			}
		}
		if c == 0 {
			rawKey, next, info, err := d.stringDecoder.scanString(buf, cursor)
			if err != nil {
				return 0, err
			}
			field, key = d.keys.lookup(rawKey, info)
			c = next
		}
		cursor = skipWhiteSpace(buf, c)
		if char(b, cursor) != ':' {
			return 0, errors.ErrExpected("colon after object key", cursor)
		}
		cursor++
		if cursor >= buflen {
			return 0, errors.ErrExpected("object value after colon", cursor)
		}
		if field != nil {
			if field.err != nil {
				return 0, field.err
			}
			if firstWin {
				if _, exists := seenFields[field.fieldIdx]; exists {
					c, err := skipValue(buf, cursor, depth)
					if err != nil {
						return 0, err
					}
					cursor = c
				} else {
					c, err := field.dec.Decode(ctx, cursor, depth, unsafe.Add(p, field.offset))
					if err != nil {
						return 0, err
					}
					cursor = c
					seenFieldNum++
					if d.fieldUniqueNameNum <= seenFieldNum {
						return skipCompound(buf, cursor, 1, depth)
					}
					seenFields[field.fieldIdx] = struct{}{}
				}
			} else {
				c, err := field.dec.Decode(ctx, cursor, depth, unsafe.Add(p, field.offset))
				if err != nil {
					return 0, err
				}
				cursor = c
			}
		} else if disallowUnknownFields {
			return 0, fmt.Errorf("json: unknown field %q", key)
		} else {
			c, err := skipValue(buf, cursor, depth)
			if err != nil {
				return 0, err
			}
			cursor = c
		}
		cursor = skipWhiteSpace(buf, cursor)
		if char(b, cursor) == '}' {
			cursor++
			return cursor, nil
		}
		if char(b, cursor) != ',' {
			return 0, errors.ErrExpected("comma after object element", cursor)
		}
		cursor++
	}
}

func (d *structDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: struct decoder does not support decode path")
}

// keyLengthInWord returns the position in the word of the first byte which ends a simple key: a quote,
// a backslash, a control character or a byte which is not ASCII, or 8 if the word has none.
func keyLengthInWord(w uint64) int {
	special := byteMask(w, '"') | byteMask(w, '\\') | hasLess(w, 0x20) | w&msb
	return bits.TrailingZeros64(special) / 8
}
