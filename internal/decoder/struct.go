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
		field, c, err := d.decodeKey(buf, cursor, disallowUnknownFields)
		if err != nil {
			return 0, err
		}
		if char(b, c) == ':' {
			cursor = c + 1
		} else {
			cursor = skipWhiteSpace(buf, c)
			if char(b, cursor) != ':' {
				return 0, errors.ErrExpected("colon after object key", cursor)
			}
			cursor++
		}
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

// decodeKey reads the key of an object at cursor, and returns its field or nil, and the position after it.
// A key which matches no field is an error if disallowUnknownFields is set. It is a function of its own, so that
// the loop of Decode keeps its values in the registers, and the key is not returned, which would take registers
// of the loop too.
func (d *structDecoder) decodeKey(buf []byte, cursor int64, disallowUnknownFields bool) (*structFieldSet, int64, error) {
	if buf[cursor] != '"' {
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] != '"' {
			return nil, 0, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
		}
	}
	// A key of ASCII without an escape is found by the words of the buffer, which may be read up to its capacity:
	// the nul byte at its end stops a key before it. Any other key, and a key near the end of a buffer which has
	// no room after it, is scanned as a string.
	if start := cursor + 1; start+16 <= int64(cap(buf)) {
		keys := d.keys
		w0 := load64(buf, start)
		special := specialKeyBytes(w0)
		if special != 0 {
			// A key of less than 8 bytes, which is most keys. The first entry of the table for it is looked at
			// here: most keys are found in it. Its index is made from the bytes of the key before their fold, and
			// so is computed while they are folded, and the mask of the bytes from the bits of the special byte.
			n := bits.TrailingZeros64(special) / 8
			if buf[start+int64(n)] != '"' {
				return d.decodeKeyNotASCII(buf, cursor, disallowUnknownFields)
			}
			m := (special ^ (special - 1)) >> 8
			w0 &= m
			fw := foldASCIIWord(w0)
			e := &keys.entries[keys.index(w0|bit5&m, 0)]
			var field *structFieldSet
			if e.short == fw {
				field = e.unique
			} else if e.fields != nil && keys.hasLength(n) {
				field = keys.findASCII(buf[start:start+int64(n)], w0, 0)
			}
			if field == nil && disallowUnknownFields {
				return nil, 0, unknownFieldError(buf[start : start+int64(n)])
			}
			return field, start + int64(n) + 1, nil
		}
		// A key of 8 bytes or more, whose end is looked for word by word. It is looked up by its first and last
		// words, as a short key is.
		end := start + 8
		special = specialKeyBytes(load64(buf, end))
		for special == 0 {
			end += 8
			if end+8 > int64(cap(buf)) {
				return d.decodeKeyByScan(buf, cursor, disallowUnknownFields)
			}
			special = specialKeyBytes(load64(buf, end))
		}
		end += int64(bits.TrailingZeros64(special) / 8)
		if buf[end] != '"' {
			return d.decodeKeyNotASCII(buf, cursor, disallowUnknownFields)
		}
		key := buf[start:end]
		n := len(key)
		var field *structFieldSet
		// A key of a length which no field has, which is a key of no field in most objects of such a length,
		// is not looked up.
		if keys.hasLength(n) {
			w1 := load64(buf, end-8)
			fw0, fw1 := foldASCIIWord(w0), foldASCIIWord(w1)
			e := &keys.entries[keys.index(w0|bit5, w1|bit5)]
			if e.n == n && e.w0 == fw0 && e.w1 == fw1 && e.unique != nil && (n <= 16 || equalFoldedASCII(key, e.folded)) {
				field = e.unique
			} else if e.fields != nil {
				field = keys.findASCII(key, w0, w1)
			}
		}
		if field == nil && disallowUnknownFields {
			return nil, 0, unknownFieldError(key)
		}
		return field, end + 1, nil
	}
	return d.decodeKeyByScan(buf, cursor, disallowUnknownFields)
}

// decodeKeyNotASCII is decodeKey for a key which has a byte which is not ASCII, or an escape. A key without an
// escape is read word by word too, and looked up by its bytes as they are among the keys which are not ASCII
// first: a key is usually the same as the key of its field, which it is found by whatever the fold of its runes,
// because a key which is the same wins over the others. Any other key is folded rune by rune if its runes may
// fold to the ones of a key, and else is of no field. A key with an escape is scanned as a string.
func (d *structDecoder) decodeKeyNotASCII(buf []byte, cursor int64, disallowUnknownFields bool) (*structFieldSet, int64, error) {
	start := cursor + 1
	end := start
	for {
		if end+8 > int64(cap(buf)) {
			return d.decodeKeyByScan(buf, cursor, disallowUnknownFields)
		}
		if special := keyEndBytes(load64(buf, end)); special != 0 {
			end += int64(bits.TrailingZeros64(special) / 8)
			break
		}
		end += 8
	}
	if buf[end] != '"' {
		return d.decodeKeyByScan(buf, cursor, disallowUnknownFields)
	}
	key := buf[start:end]
	keys := d.keys
	var field *structFieldSet
	if keys.exact != nil {
		field = keys.findExact(key)
	}
	if field == nil {
		if keys.mayFold(key) {
			field, key = keys.lookup(key, stringInfo{firstEscape: -1, nonASCII: true})
		} else if disallowUnknownFields {
			key = decodeLiteral(key, stringInfo{firstEscape: -1, nonASCII: true})
		}
	}
	if field == nil && disallowUnknownFields {
		return nil, 0, unknownFieldError(key)
	}
	return field, end + 1, nil
}

// decodeKeyByScan is decodeKey for any key, which it scans as a string.
func (d *structDecoder) decodeKeyByScan(buf []byte, cursor int64, disallowUnknownFields bool) (*structFieldSet, int64, error) {
	rawKey, next, info, err := d.stringDecoder.scanString(buf, cursor)
	if err != nil {
		return nil, 0, err
	}
	field, key := d.keys.lookup(rawKey, info)
	if field == nil && disallowUnknownFields {
		return nil, 0, unknownFieldError(key)
	}
	return field, next, nil
}

func unknownFieldError(key []byte) error {
	return fmt.Errorf("json: unknown field %q", key)
}

// keyLengthInWord returns the position in the word of the first byte which ends a simple key: a quote,
// a backslash, a control character or a byte which is not ASCII, or 8 if the word has none.
func keyLengthInWord(w uint64) int {
	return bits.TrailingZeros64(specialKeyBytes(w)) / 8
}

// specialKeyBytes returns the word which has the top bit of the first byte of w which ends a simple key, and
// maybe the top bits of bytes after it: a quote, a backslash, a control character or a byte which is not ASCII.
func specialKeyBytes(w uint64) uint64 {
	// Only the first such byte is looked for, so the masks may be wrong in the bytes after it, which saves
	// instructions: a subtraction borrows from the next byte only at a byte which is found.
	// A byte of a quote or a backslash is 0 in q or s, whose subtraction sets its top bit; the one of a control
	// character sets the top bit of the subtraction of 0x20, and the one of a byte which is not ASCII is set.
	q := w ^ ('"' * lsb)
	s := w ^ ('\\' * lsb)
	return (((q - lsb) &^ q) | ((s - lsb) &^ s) | (w - 0x20*lsb) | w) & msb
}

// keyEndBytes is specialKeyBytes for a key which may have bytes which are not ASCII: it has the top bit of the
// first byte of w which is a quote, a backslash or a control character, and maybe of bytes after it.
func keyEndBytes(w uint64) uint64 {
	// a byte of 0x80 or more is none of them: its top bit is cleared from the subtractions, which don't borrow
	// from the next byte at such a byte.
	q := w ^ ('"' * lsb)
	s := w ^ ('\\' * lsb)
	return (((q - lsb) &^ q) | ((s - lsb) &^ s) | ((w - 0x20*lsb) &^ w)) & msb
}
