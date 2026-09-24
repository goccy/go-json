package decoder

import (
	"fmt"
	"math"
	"math/bits"
	"sort"
	"strings"
	"unicode"
	"unicode/utf16"
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
	fieldMap           map[string]*structFieldSet
	fieldUniqueNameNum int
	stringDecoder      *stringDecoder
	structName         string
	fieldName          string
	isTriedOptimize    bool
	keyBitmapUint8     [][256]uint8
	keyBitmapUint16    [][256]uint16
	sortedFieldSets    []*structFieldSet
	keyDecoder         func(*structDecoder, []byte, int64) (int64, *structFieldSet, error)
}

var (
	largeToSmallTable [256]byte
)

func init() {
	for i := 0; i < 256; i++ {
		c := i
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		largeToSmallTable[i] = byte(c)
	}
}

func toASCIILower(s string) string {
	b := []byte(s)
	for i := range b {
		b[i] = largeToSmallTable[b[i]]
	}
	return string(b)
}

func newStructDecoder(structName, fieldName string, fieldMap map[string]*structFieldSet) *structDecoder {
	return &structDecoder{
		fieldMap:      fieldMap,
		stringDecoder: newStringDecoder(structName, fieldName),
		structName:    structName,
		fieldName:     fieldName,
		keyDecoder:    decodeKey,
	}
}

const (
	allowOptimizeMaxKeyLen   = 64
	allowOptimizeMaxFieldLen = 16
)

func (d *structDecoder) tryOptimize() {
	fieldUniqueNameMap := map[string]int{}
	fieldIdx := -1
	for k, v := range d.fieldMap {
		lower := strings.ToLower(k)
		idx, exists := fieldUniqueNameMap[lower]
		if exists {
			v.fieldIdx = idx
		} else {
			fieldIdx++
			v.fieldIdx = fieldIdx
		}
		fieldUniqueNameMap[lower] = fieldIdx
	}
	d.fieldUniqueNameNum = len(fieldUniqueNameMap)

	if d.isTriedOptimize {
		return
	}
	fieldMap := map[string]*structFieldSet{}
	conflicted := map[string]struct{}{}
	for k, v := range d.fieldMap {
		key := strings.ToLower(k)
		if key != k {
			if key != toASCIILower(k) {
				d.isTriedOptimize = true
				return
			}
			// already exists same key (e.g. Hello and HELLO has same lower case key
			if _, exists := conflicted[key]; exists {
				d.isTriedOptimize = true
				return
			}
			conflicted[key] = struct{}{}
		}
		if field, exists := fieldMap[key]; exists {
			if field != v {
				d.isTriedOptimize = true
				return
			}
		}
		fieldMap[key] = v
	}

	if len(fieldMap) > allowOptimizeMaxFieldLen {
		d.isTriedOptimize = true
		return
	}

	var maxKeyLen int
	sortedKeys := []string{}
	for key := range fieldMap {
		keyLen := len(key)
		if keyLen > allowOptimizeMaxKeyLen {
			d.isTriedOptimize = true
			return
		}
		if maxKeyLen < keyLen {
			maxKeyLen = keyLen
		}
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	// By allocating one extra capacity than `maxKeyLen`,
	// it is possible to avoid the process of comparing the index of the key with the length of the bitmap each time.
	bitmapLen := maxKeyLen + 1
	if len(sortedKeys) <= 8 {
		keyBitmap := make([][256]uint8, bitmapLen)
		for i, key := range sortedKeys {
			for j := 0; j < len(key); j++ {
				c := key[j]
				keyBitmap[j][c] |= (1 << uint(i))
			}
			d.sortedFieldSets = append(d.sortedFieldSets, fieldMap[key])
		}
		d.keyBitmapUint8 = keyBitmap
		d.keyDecoder = decodeKeyByBitmapUint8
	} else {
		keyBitmap := make([][256]uint16, bitmapLen)
		for i, key := range sortedKeys {
			for j := 0; j < len(key); j++ {
				c := key[j]
				keyBitmap[j][c] |= (1 << uint(i))
			}
			d.sortedFieldSets = append(d.sortedFieldSets, fieldMap[key])
		}
		d.keyBitmapUint16 = keyBitmap
		d.keyDecoder = decodeKeyByBitmapUint16
	}
}

// decode from '\uXXXX'
func decodeKeyCharByUnicodeRune(buf []byte, cursor int64) ([]byte, int64, error) {
	const defaultOffset = 4
	const surrogateOffset = 6

	if cursor+defaultOffset >= int64(len(buf)) {
		return nil, 0, errors.ErrUnexpectedEndOfJSON("escaped string", cursor)
	}

	r := unicodeToRune(buf[cursor : cursor+defaultOffset])
	if utf16.IsSurrogate(r) {
		cursor += defaultOffset
		if cursor+surrogateOffset >= int64(len(buf)) || buf[cursor] != '\\' || buf[cursor+1] != 'u' {
			return []byte(string(unicode.ReplacementChar)), cursor + defaultOffset - 1, nil
		}
		cursor += 2
		r2 := unicodeToRune(buf[cursor : cursor+defaultOffset])
		if r := utf16.DecodeRune(r, r2); r != unicode.ReplacementChar {
			return []byte(string(r)), cursor + defaultOffset - 1, nil
		}
	}
	return []byte(string(r)), cursor + defaultOffset - 1, nil
}

func decodeKeyCharByEscapedChar(buf []byte, cursor int64) ([]byte, int64, error) {
	c := buf[cursor]
	cursor++
	switch c {
	case '"':
		return []byte{'"'}, cursor, nil
	case '\\':
		return []byte{'\\'}, cursor, nil
	case '/':
		return []byte{'/'}, cursor, nil
	case 'b':
		return []byte{'\b'}, cursor, nil
	case 'f':
		return []byte{'\f'}, cursor, nil
	case 'n':
		return []byte{'\n'}, cursor, nil
	case 'r':
		return []byte{'\r'}, cursor, nil
	case 't':
		return []byte{'\t'}, cursor, nil
	case 'u':
		return decodeKeyCharByUnicodeRune(buf, cursor)
	}
	return nil, cursor, nil
}

func decodeKeyByBitmapUint8(d *structDecoder, buf []byte, cursor int64) (int64, *structFieldSet, error) {
	var (
		curBit uint8 = math.MaxUint8
	)
	b := (*sliceHeader)(unsafe.Pointer(&buf)).data
	for {
		switch char(b, cursor) {
		case ' ', '\n', '\t', '\r':
			cursor++
		case '"':
			cursor++
			c := char(b, cursor)
			switch c {
			case '"':
				cursor++
				return cursor, nil, nil
			case nul:
				return 0, nil, errors.ErrUnexpectedEndOfJSON("string", cursor)
			}
			keyIdx := 0
			bitmap := d.keyBitmapUint8
			start := cursor
			for {
				c := char(b, cursor)
				switch c {
				case '"':
					fieldSetIndex := bits.TrailingZeros8(curBit)
					field := d.sortedFieldSets[fieldSetIndex]
					keyLen := cursor - start
					cursor++
					if keyLen < field.keyLen {
						// early match
						return cursor, nil, nil
					}
					return cursor, field, nil
				case nul:
					return 0, nil, errors.ErrUnexpectedEndOfJSON("string", cursor)
				case '\\':
					cursor++
					chars, nextCursor, err := decodeKeyCharByEscapedChar(buf, cursor)
					if err != nil {
						return 0, nil, err
					}
					for _, c := range chars {
						curBit &= bitmap[keyIdx][largeToSmallTable[c]]
						if curBit == 0 {
							return decodeKeyNotFound(b, cursor)
						}
						keyIdx++
					}
					cursor = nextCursor
				default:
					curBit &= bitmap[keyIdx][largeToSmallTable[c]]
					if curBit == 0 {
						return decodeKeyNotFound(b, cursor)
					}
					keyIdx++
				}
				cursor++
			}
		default:
			return cursor, nil, errors.ErrInvalidBeginningOfValue(char(b, cursor), cursor)
		}
	}
}

func decodeKeyByBitmapUint16(d *structDecoder, buf []byte, cursor int64) (int64, *structFieldSet, error) {
	var (
		curBit uint16 = math.MaxUint16
	)
	b := (*sliceHeader)(unsafe.Pointer(&buf)).data
	for {
		switch char(b, cursor) {
		case ' ', '\n', '\t', '\r':
			cursor++
		case '"':
			cursor++
			c := char(b, cursor)
			switch c {
			case '"':
				cursor++
				return cursor, nil, nil
			case nul:
				return 0, nil, errors.ErrUnexpectedEndOfJSON("string", cursor)
			}
			keyIdx := 0
			bitmap := d.keyBitmapUint16
			start := cursor
			for {
				c := char(b, cursor)
				switch c {
				case '"':
					fieldSetIndex := bits.TrailingZeros16(curBit)
					field := d.sortedFieldSets[fieldSetIndex]
					keyLen := cursor - start
					cursor++
					if keyLen < field.keyLen {
						// early match
						return cursor, nil, nil
					}
					return cursor, field, nil
				case nul:
					return 0, nil, errors.ErrUnexpectedEndOfJSON("string", cursor)
				case '\\':
					cursor++
					chars, nextCursor, err := decodeKeyCharByEscapedChar(buf, cursor)
					if err != nil {
						return 0, nil, err
					}
					for _, c := range chars {
						curBit &= bitmap[keyIdx][largeToSmallTable[c]]
						if curBit == 0 {
							return decodeKeyNotFound(b, cursor)
						}
						keyIdx++
					}
					cursor = nextCursor
				default:
					curBit &= bitmap[keyIdx][largeToSmallTable[c]]
					if curBit == 0 {
						return decodeKeyNotFound(b, cursor)
					}
					keyIdx++
				}
				cursor++
			}
		default:
			return cursor, nil, errors.ErrInvalidBeginningOfValue(char(b, cursor), cursor)
		}
	}
}

func decodeKeyNotFound(b unsafe.Pointer, cursor int64) (int64, *structFieldSet, error) {
	for {
		cursor++
		switch char(b, cursor) {
		case '"':
			cursor++
			return cursor, nil, nil
		case '\\':
			cursor++
			if char(b, cursor) == nul {
				return 0, nil, errors.ErrUnexpectedEndOfJSON("string", cursor)
			}
		case nul:
			return 0, nil, errors.ErrUnexpectedEndOfJSON("string", cursor)
		}
	}
}

func decodeKey(d *structDecoder, buf []byte, cursor int64) (int64, *structFieldSet, error) {
	key, c, err := d.stringDecoder.decodeByte(buf, cursor)
	if err != nil {
		return 0, nil, err
	}
	cursor = c
	k := *(*string)(unsafe.Pointer(&key))
	field, exists := d.fieldMap[k]
	if !exists {
		return cursor, nil, nil
	}
	return cursor, field, nil
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
		keyStart := cursor
		c, field, err := d.keyDecoder(d, buf, cursor)
		if err != nil {
			return 0, err
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
			key, _, err := d.stringDecoder.decodeByte(buf, skipWhiteSpace(buf, keyStart))
			if err != nil {
				return 0, err
			}
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
