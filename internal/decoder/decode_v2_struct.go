package decoder

import (
	"bytes"
	"encoding/binary"
	stderrors "errors"
	"fmt"
	"math/bits"
	"reflect"
	"strings"
	"unicode/utf8"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonfields"
	"github.com/goccy/go-json/internal/runtime"
	"github.com/goccy/go-json/internal/textcoder"
)

// The structs and the maps of the semantics of encoding/json/v2. The members of a struct are the fields which
// jsonfields finds, whose names are matched exactly, and by their folded names ( see jsonfields.FoldName ) only if
// their `case` option or V2MatchCaseInsensitiveNames says so. A name of no field is decoded into the embedded
// fallback of the struct, if it has one, or else skipped, or is an error with V2RejectUnknownMembers. An object
// which has a field twice, or a name of no field twice, is an error unless V2AllowDuplicateNames. An object is
// merged into a map, whose entries are decoded into the values they had.

var (
	errNilField         = stderrors.New("cannot set embedded pointer to unexported struct type")
	errInvalidStringTag = stderrors.New("invalid use of `string` tag option")
	errRawNotObject     = stderrors.New("embedded raw value must be a JSON object")
)

// v2StructDecoder is the decoder of a struct of v2.
type v2StructDecoder struct {
	typ  reflect.Type
	zero zeroer
	// initErr is the error of the definition of the struct, which an object of it reports: it has no fields then.
	initErr *jsonfields.Error
	// fields are the fields by their indexes, which are the bits of the set of the fields which an object has.
	// keyDecoder finds the field of a name as v1 finds it, exactly or else by its folded name, which is taken if
	// the name is the name of the field as it is.
	fields     []*structFieldSet
	keyDecoder *structDecoder
	// foldedTo is the field which a name of the folded name of each field matches ( see jsonfields.FoldName ), by the
	// indexes of the fields: [0] without V2MatchCaseInsensitiveNames, where only the fields of the `case:ignore`
	// option match a name so, and [1] with it, where all but the ones of `case:strict` do. It is nil where no field
	// matches the name, and ambiguousField where several do. A name is of the folded name of the field which v1
	// finds for it, or which foldKeys finds for it without its '_' and '-' ( see foldKey ). foldKeys is a table as
	// v1's of the names of the fields without their '_' and '-', v1's own if no name of a field has them. foldAlways
	// is whether a field matches by its `case:ignore` option, without V2MatchCaseInsensitiveNames.
	foldedTo   [2][]*structFieldSet
	foldKeys   *structKeys
	foldAlways bool
	fallback   *v2Fallback
}

// compileStructV2 returns the decoder of the struct type typ. It stays in c.decoders after the compile, as the one
// of compileStruct does, unless a decoder which declines values to it is there ( see compileNextV2 ).
func (c *compiler) compileStructV2(typ reflect.Type) (Decoder, error) {
	d := &v2StructDecoder{typ: typ, zero: newZeroer(typ)}
	if _, exists := c.decoders[uintptr(runtime.TypePtr(typ))]; !exists {
		c.decoders[uintptr(runtime.TypePtr(typ))] = d
	}
	fields := jsonfields.Of(typ)
	switch {
	case fields.Err != nil:
		d.initErr = fields.Err
		return d, nil
	case fields.FormatErr != nil:
		d.initErr = fields.FormatErr
		return d, nil
	}
	// the fields whose names may match a name by their folded names, by the folded names
	folded := map[string][]*structFieldSet{}
	for i := range fields.List {
		f := &fields.List[i]
		var dec Decoder
		var err error
		if f.String {
			dec, err = c.stringTagCompiler().compileV2(f.Type, typ.Name(), f.Name)
		} else {
			dec, err = compile(f.Type, typ.Name(), f.Name, c)
		}
		if err != nil {
			return nil, err
		}
		offset, dec := fieldPathV2(typ, typ, f.Index, dec)
		set := &structFieldSet{dec: dec, offset: offset, key: f.Name, keyLen: int64(len(f.Name)), fieldIdx: i}
		d.fields = append(d.fields, set)
		if f.Casing&jsonfields.CaseStrict == 0 {
			key := string(jsonfields.FoldName([]byte(f.Name)))
			folded[key] = append(folded[key], set)
			d.foldAlways = d.foldAlways || f.Casing&jsonfields.CaseIgnore != 0
		}
	}
	d.foldedTo = [2][]*structFieldSet{make([]*structFieldSet, len(d.fields)), make([]*structFieldSet, len(d.fields))}
	dashes := false
	for i, f := range d.fields {
		for _, other := range folded[string(jsonfields.FoldName([]byte(f.key)))] {
			if fields.List[other.fieldIdx].Casing&jsonfields.CaseIgnore != 0 {
				d.foldedTo[0][i] = onlyField(d.foldedTo[0][i], other)
			}
			d.foldedTo[1][i] = onlyField(d.foldedTo[1][i], other)
		}
		dashes = dashes || strings.ContainsAny(f.key, "_-")
	}
	keys := newStructKeys(d.fields)
	d.foldKeys = keys
	if dashes {
		// the fields under their names without '_' and '-', of their indexes
		without := make([]*structFieldSet, len(d.fields))
		for i, f := range d.fields {
			var arr [64]byte
			key, _ := foldKey(arr[:0], []byte(f.key))
			w := *f
			w.key, w.keyLen = string(key), int64(len(key))
			without[i] = &w
		}
		d.foldKeys = newStructKeys(without)
	}
	// every name of a field by its bytes as they are, which a name as the input has it is found by ( see
	// plainFieldV2 )
	keys.makeExact(d.fields, true)
	d.keyDecoder = &structDecoder{fields: d.fields, keys: keys, stringDecoder: skipStringDecoder}
	if fields.Fallback != nil {
		fb, err := c.compileFallback(typ, fields.Fallback)
		if err != nil {
			return nil, err
		}
		d.fallback = fb
	}
	return d, nil
}

// stringTagCompiler returns the compiler of the values of a field of the `string` option: a number is decoded from
// a JSON string, through the pointers to it, and a value of any other type can't be decoded.
func (c *compiler) stringTagCompiler() *compiler {
	return &compiler{decoders: map[uintptr]Decoder{}, v2: true, stringifyNumbers: true, stringTag: true, funcs: c.funcs}
}

// fieldPathV2 returns the offset of the field of the struct typ at index, by reflect.Type.FieldByIndex, and the
// decoder of its value, which dec decodes: a field in a struct which an embedded pointer points to is decoded
// through the pointer, which is allocated if it is nil and can be set. root is the struct type of the object.
func fieldPathV2(root, typ reflect.Type, index []int, dec Decoder) (uintptr, Decoder) {
	sf := typ.Field(index[0])
	if len(index) == 1 {
		return sf.Offset, dec
	}
	if sf.Type.Kind() == reflect.Pointer {
		offset, inner := fieldPathV2(root, sf.Type.Elem(), index[1:], dec)
		return sf.Offset, &v2EmbeddedPtrDecoder{
			elem: sf.Type.Elem(), offset: offset, dec: inner, settable: sf.IsExported(), root: root,
		}
	}
	offset, inner := fieldPathV2(root, sf.Type, index[1:], dec)
	return sf.Offset + offset, inner
}

// v2EmbeddedPtrDecoder decodes a field of a struct which an embedded pointer points to, through the pointer: a nil
// pointer is set to a new value, if it can be set, and is an error of the struct of the object, root, else.
type v2EmbeddedPtrDecoder struct {
	elem     reflect.Type
	offset   uintptr
	dec      Decoder
	settable bool
	root     reflect.Type
}

func (d *v2EmbeddedPtrDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	if *(*unsafe.Pointer)(p) == nil {
		if !d.settable {
			cursor = skipWhiteSpace(ctx.Buf, cursor)
			return 0, &errors.UnmarshalError{GoType: d.root, Start: cursor, NoKind: true, Err: errNilField}
		}
		*(*unsafe.Pointer)(p) = newValue(d.elem)
	}
	return d.dec.Decode(ctx, cursor, depth, unsafe.Add(*(*unsafe.Pointer)(p), d.offset))
}

func (d *v2EmbeddedPtrDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

func (d *v2StructDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		d.zero.zero(p)
		return cursor + 4, nil
	case '{':
	default:
		return ctx.skipTypeError(cursor, depth, d.typ)
	}
	if d.initErr != nil {
		return 0, &errors.UnmarshalError{GoType: d.initErr.GoType, Start: cursor, Err: d.initErr.Err}
	}
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}
	cursor = skipWhiteSpace(buf, cursor+1)
	if buf[cursor] == '}' {
		return cursor + 1, nil
	}
	rejectDuplicateNames := ctx.Option.V2&V2AllowDuplicateNames == 0
	fold := d.foldAlways || ctx.Option.V2&V2MatchCaseInsensitiveNames != 0
	// seen are the fields which the object has, by their bits, after the first 64 in more, and unknown the names
	// of no field which it has.
	var seen uint64
	var more []uint64
	var unknown nameSet
	for {
		start := cursor
		if buf[start] != '"' {
			// white space after the comma, which is checked here, as most names follow the comma at once
			if start = skipWhiteSpace(buf, start); buf[start] != '"' {
				return 0, syntaxErrorAt(buf, start, whereKey)
			}
		}
		var field *structFieldSet
		var end int64
		var plain bool
		if fold {
			field, end, plain = d.foldedFieldV2(ctx, buf, start)
		} else {
			field, end, plain = plainFieldV2(d.keyDecoder.keys, buf, start)
		}
		// name is the name decoded, which a name of no field is checked by, or nil for a name whose field the lookup
		// decided, which is as the input has it
		var name []byte
		if !plain {
			// a name with an escape, a byte which is not ASCII or a control character, a name near the end of the
			// buffer ( see plainFieldV2 ), or a name whose field by its folded name is not decided ( see foldedFieldV2 )
			candidate, next, err := d.keyDecoder.decodeKey(buf, start, false)
			if err != nil {
				return 0, err
			}
			end = next
			if field, name, err = d.memberOf(ctx, buf[start:end], candidate); err != nil {
				if e, ok := err.(*errors.UnmarshalError); ok {
					// an error of the name, at it
					e.Start = start
				}
				return 0, err
			}
		}
		// the name is checked before the colon after it, as encoding/json/v2 reads the name alone
		if field != nil {
			if rejectDuplicateNames {
				if idx := field.fieldIdx; idx < 64 {
					// the first 64 fields are the bits of seen, in a register
					bit := uint64(1) << uint(idx)
					if seen&bit != 0 {
						return 0, &errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}
					}
					seen |= bit
				} else if fieldSeen(idx, &more, len(d.fields)) {
					return 0, &errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}
				}
			}
		} else if d.fallback == nil {
			// a member of no field, whose value is skipped, and the members of no field after it
			next, done, err := d.skipUnknownMembers(ctx, &unknown, fold, buf, start, end, name, depth)
			if err != nil {
				return 0, err
			}
			if done {
				unknown.end(ctx.v2)
				return next, nil
			}
			cursor = next
			continue
		} else if err := d.unknownMember(ctx, &unknown, rejectDuplicateNames, buf, start, end, name); err != nil {
			return 0, err
		}
		if buf[end] == ':' {
			cursor = end + 1
		} else {
			cursor = skipWhiteSpace(buf, end)
			if buf[cursor] != ':' {
				return 0, syntaxErrorAt(buf, cursor, whereAfterKey)
			}
			cursor++
		}
		var err error
		if field != nil {
			cursor, err = field.dec.Decode(ctx, cursor, depth, unsafe.Add(p, field.offset))
		} else {
			cursor, err = d.fallback.decodeMember(ctx, cursor, depth, p, start, end)
		}
		if err != nil {
			return 0, err
		}
		cursor = skipWhiteSpace(buf, cursor)
		switch buf[cursor] {
		case '}':
			unknown.end(ctx.v2)
			return cursor + 1, nil
		case ',':
			cursor++
		default:
			return 0, syntaxErrorAt(buf, cursor, whereAfterPair)
		}
	}
}

// skipUnknownMembers skips the member of no field of a struct without a fallback whose name is at start in buf, before
// end, decoded as name, or nil if Decode found it of no field without decoding it ( see plainFieldV2 and
// foldedFieldV2 ), and the members of no field after it: up to the end of the object, after which it returns the
// position and done, or up to a name which is of a field or which Decode decodes, as a name which is not plain or is
// near the end of the buffer ( see plainFieldV2 ), or whose field by its folded name is not decided ( see
// foldedFieldV2 ), whose position it returns, which Decode reads again. A member of no field is an error with
// V2RejectUnknownMembers. It is a loop of its own, so that the loop of Decode is entered once for the members of no
// field one after another, as most such members are, and the state of that loop is not kept around their calls.
func (d *v2StructDecoder) skipUnknownMembers(ctx *RuntimeContext, unknown *nameSet, fold bool, buf []byte, start, end int64, name []byte, depth int64) (int64, bool, error) {
	if ctx.Option.V2&V2RejectUnknownMembers != 0 {
		return 0, false, &errors.UnmarshalError{GoType: d.typ, Start: start, Err: errors.ErrUnknownName}
	}
	rejectDuplicateNames := ctx.Option.V2&V2AllowDuplicateNames == 0
	// the values are skipped as skipValueV2 skips them, by the kind of their first byte: the objects and the arrays by
	// skipNestedV2 unless v2 skips them as v1 does, the strings by skipStringV2 if they must be valid UTF-8, and the
	// others by v1
	nested := rejectDuplicateNames || ctx.Option.V2&V2AllowInvalidUTF8 == 0
	validUTF8 := ctx.Option.V2&V2AllowInvalidUTF8 == 0
	for {
		if rejectDuplicateNames {
			var added bool
			if name == nil {
				added = unknown.add(ctx.V2Call(), buf, start+1, end-1)
			} else {
				added = unknown.addCopy(ctx.V2Call(), buf, name)
			}
			if !added {
				return 0, false, &errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}
			}
		}
		cursor := end
		if buf[cursor] == ':' {
			cursor++
		} else {
			if cursor = skipWhiteSpace(buf, cursor); buf[cursor] != ':' {
				return 0, false, syntaxErrorAt(buf, cursor, whereAfterKey)
			}
			cursor++
		}
		var err error
		switch cursor = skipWhiteSpace(buf, cursor); buf[cursor] {
		case '{', '[':
			if nested {
				cursor, err = skipNestedV2(ctx, cursor, depth)
			} else {
				cursor, err = skipValue(buf, cursor, depth)
			}
		case '"':
			if validUTF8 {
				cursor, err = skipStringV2(buf, cursor)
			} else {
				cursor, err = skipValue(buf, cursor, depth)
			}
		default:
			cursor, err = skipValue(buf, cursor, depth)
		}
		if err != nil {
			return 0, false, err
		}
		switch cursor = skipWhiteSpace(buf, cursor); buf[cursor] {
		case '}':
			return cursor + 1, true, nil
		case ',':
			cursor++
		default:
			return 0, false, syntaxErrorAt(buf, cursor, whereAfterPair)
		}
		if start = cursor; buf[start] != '"' {
			if start = skipWhiteSpace(buf, start); buf[start] != '"' {
				return 0, false, syntaxErrorAt(buf, start, whereKey)
			}
		}
		var field *structFieldSet
		var plain bool
		if fold {
			field, end, plain = d.foldedFieldV2(ctx, buf, start)
		} else {
			field, end, plain = plainFieldV2(d.keyDecoder.keys, buf, start)
		}
		if field != nil || !plain {
			return start, false, nil
		}
		name = nil
	}
}

// unknownMember checks the name of no field at start in buf, before end, decoded as name, or nil if Decode found it of
// no field without decoding it ( see skipUnknownMembers ), of a struct with a fallback: a name which the object had
// before is an error if rejectDuplicateNames. unknown are the names of no field which the object had. It is a function
// of its own, so that the loop of Decode keeps nothing of it.
func (d *v2StructDecoder) unknownMember(ctx *RuntimeContext, unknown *nameSet, rejectDuplicateNames bool, buf []byte, start, end int64, name []byte) error {
	if !rejectDuplicateNames {
		return nil
	}
	var added bool
	if name == nil {
		added = unknown.add(ctx.V2Call(), buf, start+1, end-1)
	} else {
		added = unknown.addCopy(ctx.V2Call(), buf, name)
	}
	if !added {
		return &errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}
	}
	return nil
}

func (d *v2StructDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// fieldSeen reports whether the field of the index, 64 or more, was met before in an object of n fields, and marks
// it met: its bit is in more, which is made when such a field is met. The first 64 fields are marked by the loops of
// the objects themselves, in a register.
func fieldSeen(idx int, more *[]uint64, n int) bool {
	if *more == nil {
		*more = make([]uint64, (n-64+63)/64)
	}
	idx -= 64
	w, bit := &(*more)[idx/64], uint64(1)<<uint(idx%64)
	met := *w&bit != 0
	*w |= bit
	return met
}

// plainFieldV2 finds the field of the name at cursor in buf, a quote, if the name has only bytes of ASCII which are
// not an escape or a control character ( a plain name, which most names are ): the field whose name is the name as
// it is. It returns the field or nil, the position after the name, and whether the name is plain. The name is read
// word by word, as decodeKey reads it, up to the capacity of buf, whose nul byte at the end stops a name before it. A
// name which is not plain, or which is near the end of a buffer without room for a word after it, is left to
// decodeKey and memberOf.
func plainFieldV2(keys *structKeys, buf []byte, cursor int64) (*structFieldSet, int64, bool) {
	start := cursor + 1
	if start+16 > int64(cap(buf)) {
		return nil, 0, false
	}
	w0 := load64(buf, start)
	if special := specialKeyBytes(w0); special != 0 {
		// a name of less than 8 bytes, whose word is the bytes before the special byte
		n := int64(bits.TrailingZeros64(special) / 8)
		if buf[start+n] != '"' {
			return nil, 0, false
		}
		next := start + n + 1
		w0 &= (special ^ (special - 1)) >> 8
		if !keys.hasExactLength(int(n)) {
			return nil, next, true
		}
		// the first entry of the name is looked at here, as decodeKey looks at the first entry of a key: most names
		// are found in it
		e := &keys.exact[indexOf((w0|bit5)&(1<<(uint(n)*8&63)-1), 0, keys.exactShift)]
		if e.n == int(n) && e.w0 == w0 && e.unique != nil {
			return e.unique, next, true
		}
		if e.unique == nil {
			return nil, next, true
		}
		return keys.findExact(buf[start:start+n], w0, 0), next, true
	}
	end := start + 8
	special := specialKeyBytes(load64(buf, end))
	for special == 0 {
		end += 8
		if end+8 > int64(cap(buf)) {
			return nil, 0, false
		}
		special = specialKeyBytes(load64(buf, end))
	}
	end += int64(bits.TrailingZeros64(special) / 8)
	if buf[end] != '"' {
		return nil, 0, false
	}
	if !keys.hasExactLength(int(end - start)) {
		return nil, end + 1, true
	}
	return keys.findExact(buf[start:end], w0, load64(buf, end-8)), end + 1, true
}

// foldedPlainFieldV2 is plainFieldV2 for the names which may match a field by their folded names: the field which
// structDecoder.decodeKey finds for a plain name by its folded name, and, for a name of no such field, the number of
// its '_' and '-', which the words of the name are counted for as they are read ( see foldKey ).
func foldedPlainFieldV2(keys *structKeys, buf []byte, cursor int64) (*structFieldSet, int64, int, bool) {
	start := cursor + 1
	if start+16 > int64(cap(buf)) {
		return nil, 0, 0, false
	}
	w0 := load64(buf, start)
	if special := specialKeyBytes(w0); special != 0 {
		n := int64(bits.TrailingZeros64(special) / 8)
		if buf[start+n] != '"' {
			return nil, 0, 0, false
		}
		next := start + n + 1
		w0 &= (special ^ (special - 1)) >> 8
		if keys.hasLength(int(n)) {
			if field := keys.findASCII(buf[start:start+n], w0, 0); field != nil {
				return field, next, 0, true
			}
		}
		return nil, next, dashCount(w0), true
	}
	dashes := dashCount(w0)
	end := start + 8
	w := load64(buf, end)
	special := specialKeyBytes(w)
	for special == 0 {
		dashes += dashCount(w)
		end += 8
		if end+8 > int64(cap(buf)) {
			return nil, 0, 0, false
		}
		w = load64(buf, end)
		special = specialKeyBytes(w)
	}
	end += int64(bits.TrailingZeros64(special) / 8)
	if buf[end] != '"' {
		return nil, 0, 0, false
	}
	if keys.hasLength(int(end - start)) {
		if field := keys.findASCII(buf[start:end], w0, load64(buf, end-8)); field != nil {
			return field, end + 1, 0, true
		}
	}
	// the bytes of the last word before the special byte
	return nil, end + 1, dashes + dashCount(w&(special-1)), true
}

// dashCount returns the number of the bytes of w which are '_' or '-', which FoldName drops ( see foldKey ).
func dashCount(w uint64) int {
	return bits.OnesCount64(byteMask(w, '_') | byteMask(w, '-'))
}

// foldedFieldV2 finds the field of the name at start in buf, a quote, which may match a field by its folded name, if
// the name is plain ( see foldedPlainFieldV2 ) and its field is decided without memberOf: the field which decodeKey
// finds for it, if its folded name matches that field ( see foldedTo ), or no field, if no field may match it by its
// folded name without its '_' and '-' ( see foldKey ). It returns the field or nil, the position after the name, and
// whether the field is decided so.
func (d *v2StructDecoder) foldedFieldV2(ctx *RuntimeContext, buf []byte, start int64) (*structFieldSet, int64, bool) {
	field, end, dashes, plain := foldedPlainFieldV2(d.keyDecoder.keys, buf, start)
	if !plain {
		return nil, 0, false
	}
	folded := d.foldedTo[0]
	if ctx.Option.V2&V2MatchCaseInsensitiveNames != 0 {
		folded = d.foldedTo[1]
	}
	if field != nil {
		return field, end, folded[field.fieldIdx] == field
	}
	// a name of no field of v1, which the table of the names of the fields without '_' and '-' may find without its
	// own, but not if no field has its length without them, which foldedPlainFieldV2 counts
	return nil, end, dashes == 0 && d.foldKeys == d.keyDecoder.keys || !d.foldKeys.hasLength(int(end-start-2)-dashes)
}

// memberOf returns the field of the name, quoted as the input has it, or nil, and the name decoded if it was:
// candidate is the field which v1 finds, exactly or by its folded name, which is the field if the name is its name
// as it is. Any other name matches the field which its folded name matches ( see foldedTo ), but a name which
// several fields match so is an error. A name of a byte which is not ASCII or of an escape must have the text which
// v2 allows.
func (d *v2StructDecoder) memberOf(ctx *RuntimeContext, quoted []byte, candidate *structFieldSet) (*structFieldSet, []byte, error) {
	name := quoted[1 : len(quoted)-1]
	if candidate != nil && string(name) == candidate.key {
		// the name of candidate as it is, which has neither a backslash nor invalid UTF-8 ( see jsonfields.Of ): it
		// has no escape, and nothing to look at
		return candidate, name, nil
	}
	plain := !hasNotASCIIOrEscape(name)
	insensitive := ctx.Option.V2&V2MatchCaseInsensitiveNames != 0
	folded := d.foldedTo[0]
	if insensitive {
		folded = d.foldedTo[1]
	}
	if !plain {
		if ctx.Option.V2&V2AllowInvalidUTF8 == 0 && !validV2Text(quoted) {
			return nil, nil, errV2Text
		}
		if name = unquotedName(quoted); candidate != nil && string(name) == candidate.key {
			return candidate, name, nil
		}
	}
	var match *structFieldSet
	switch {
	case candidate != nil:
		// the field of the folded name of candidate, which v1 found by the name folded as FoldName folds it, but for
		// '_' and '-', which no rune folds to: it is the folded name of the name
		match = folded[candidate.fieldIdx]
	case insensitive || d.foldAlways:
		// the name without its '_' and '-' in the table of the fields without theirs, which folds it as FoldName
		// folds it: v1 found nothing in the same lookup of the name without them in its table. A name of ASCII of
		// a length which no field has without them is of no field, as the lookup finds ( see structKeys.lookup ).
		if n := foldedLength(name); (n < len(name) || d.foldKeys != d.keyDecoder.keys) && (!plain || d.foldKeys.hasLength(n)) {
			var arr [64]byte
			key, nonASCII := foldKey(arr[:0], name)
			if field, _ := d.foldKeys.lookup(key, stringInfo{firstEscape: -1, nonASCII: nonASCII}); field != nil {
				match = folded[field.fieldIdx]
			}
		}
	}
	if match == ambiguousField {
		// a name which several fields match by its folded name is an error, as in encoding/json/v2, whatever their
		// order
		return nil, nil, &errors.UnmarshalError{GoType: d.typ, Err: errors.ErrAmbiguousName}
	}
	return match, name, nil
}

// ambiguousField stands for the fields of a folded name in foldedTo of v2StructDecoder which several fields match.
var ambiguousField = &structFieldSet{}

// onlyField returns the field which a folded name matches when it matched match before, nil if none did, and field
// too ( see foldedTo of v2StructDecoder ).
func onlyField(match, field *structFieldSet) *structFieldSet {
	if match == nil {
		return field
	}
	return ambiguousField
}

// foldKey returns name without its '_' and '-', appended to out if it has one of them, and whether it has a byte
// which is not ASCII. FoldName drops them, and no rune folds to them: v1's lookup of the key, which folds names as
// FoldName folds them but for '_' and '-', finds a field of the folded name of the name in a table of the names of
// the fields without them.
func foldKey(out, name []byte) ([]byte, bool) {
	i, nonASCII := 0, false
	for ; i < len(name) && name[i] != '_' && name[i] != '-'; i++ {
		nonASCII = nonASCII || name[i] >= utf8.RuneSelf
	}
	if i == len(name) {
		return name, nonASCII
	}
	out = append(out, name[:i]...)
	for _, c := range name[i+1:] {
		if c != '_' && c != '-' {
			out = append(out, c)
			nonASCII = nonASCII || c >= utf8.RuneSelf
		}
	}
	return out, nonASCII
}

// foldedLength returns the length of name without its '_' and '-', which foldKey drops.
func foldedLength(name []byte) int {
	n, i := len(name), 0
	for ; i+8 <= len(name); i += 8 {
		w := binary.LittleEndian.Uint64(name[i:])
		n -= bits.OnesCount64(byteMask(w, '_') | byteMask(w, '-'))
	}
	for _, c := range name[i:] {
		if c == '_' || c == '-' {
			n--
		}
	}
	return n
}

// hasNotASCIIOrEscape reports whether b has a byte which is not ASCII or a backslash, which it looks for eight bytes
// at a time.
func hasNotASCIIOrEscape(b []byte) bool {
	i := 0
	for ; i+8 <= len(b); i += 8 {
		w := binary.LittleEndian.Uint64(b[i:])
		// a backslash is a zero byte of s, whose subtraction sets its top bit
		s := w ^ ('\\' * lsb)
		if (w|(s-lsb)&^s)&msb != 0 {
			return true
		}
	}
	for _, c := range b[i:] {
		if c >= 0x80 || c == '\\' {
			return true
		}
	}
	return false
}

// unquotedName returns the string quoted, as the input has it, decoded.
func unquotedName(quoted []byte) []byte {
	raw, next, info, _ := skipStringDecoder.scanString(quoted, 0)
	if next < 0 {
		raw, _, info, _ = skipStringDecoder.scanStringRest(quoted, raw, -next-1, info)
	}
	return decodeLiteral(raw, info)
}

// skipperOf returns the skipper of the call of ctx of its options, which it takes for the first value: the options
// of a call which a method makes on the context of the call around it may differ. A value is read by a skipper only
// for names twice or invalid UTF-8, which the options don't both allow.
func skipperOf(ctx *RuntimeContext) unsafe.Pointer {
	call := ctx.V2Call()
	i := 0
	switch {
	case ctx.Option.V2&V2AllowDuplicateNames != 0:
		i = 2 // for invalid UTF-8 only
	case ctx.Option.V2&V2AllowInvalidUTF8 != 0:
		i = 1
	}
	if call.skippers[i] == nil {
		call.skippers[i] = textcoder.TakeSkipper(i == 2, i == 1)
	}
	return call.skippers[i]
}

// releaseSkippers puts back the skippers which the call took.
func releaseSkippers(call *V2Call) {
	for i, s := range call.skippers {
		if s != nil {
			textcoder.ReleaseSkipper(s)
			call.skippers[i] = nil
		}
	}
}

// v2Fallback is the embedded fallback of a struct, which the members of no field are decoded into: a jsontext.Value,
// which they are appended to, or a map of string keys, whose entries they are.
type v2Fallback struct {
	// steps are the fields on the way to the fallback from the struct, the fallback last: an embedded pointer is
	// allocated if it is nil.
	steps []fallbackStep
	// mapType is the type of a map, whose values value decodes, or nil for a jsontext.Value.
	mapType reflect.Type
	value   Decoder
}

// fallbackStep is a field on the way to an embedded fallback: its offset, and the type which it points to if it is a
// pointer.
type fallbackStep struct {
	offset uintptr
	elem   reflect.Type
}

func (c *compiler) compileFallback(root reflect.Type, f *jsonfields.Field) (*v2Fallback, error) {
	fb := &v2Fallback{}
	typ := root
	for _, i := range f.Index {
		sf := typ.Field(i)
		step := fallbackStep{offset: sf.Offset}
		typ = sf.Type
		if typ.Kind() == reflect.Pointer {
			step.elem, typ = typ.Elem(), typ.Elem()
		}
		fb.steps = append(fb.steps, step)
	}
	if typ != jsonfields.RawValueType {
		fb.mapType = typ
		value, err := compile(typ.Elem(), "", "", c)
		if err != nil {
			return nil, err
		}
		fb.value = value
	}
	return fb, nil
}

// target returns the address of the fallback of the struct at p.
func (fb *v2Fallback) target(p unsafe.Pointer) unsafe.Pointer {
	for _, step := range fb.steps {
		p = unsafe.Add(p, step.offset)
		if step.elem != nil {
			if *(*unsafe.Pointer)(p) == nil {
				*(*unsafe.Pointer)(p) = newValue(step.elem)
			}
			p = *(*unsafe.Pointer)(p)
		}
	}
	return p
}

// openRaw returns the jsontext.Value fallback of the struct at p, made ready for one more member to be appended: an
// object without its end and with a comma after its last member. A value which is not an object is an error.
func (fb *v2Fallback) openRaw(p unsafe.Pointer) (*[]byte, error) {
	b := (*[]byte)(fb.target(p))
	if len(*b) == 0 {
		*b = append(*b, '{')
		return b, nil
	}
	*b = bytes.TrimRight(*b, " \t\r\n")
	if len(*b) == 0 || (*b)[len(*b)-1] != '}' {
		return nil, errRawNotObject
	}
	*b = bytes.TrimRight((*b)[:len(*b)-1], " \t\r\n")
	if n := len(*b); n == 0 || (*b)[n-1] != ',' && (*b)[n-1] != '{' {
		*b = append(*b, ',')
	}
	return b, nil
}

// decodeMember decodes the value at cursor of the member whose name, between start and end of the input, quoted, is
// of no field of the struct at p, into its fallback.
func (fb *v2Fallback) decodeMember(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer, start, end int64) (int64, error) {
	quoted := ctx.Buf[start:end]
	target := fb.target(p)
	if fb.mapType == nil {
		// the value which the member is appended to is checked before the value of the member is read
		b, err := fb.openRaw(p)
		if err != nil {
			// the error of the member, at its name
			return 0, &errors.UnmarshalError{GoType: jsonfields.RawValueType, Start: start, Err: err}
		}
		valueStart := skipWhiteSpace(ctx.Buf, cursor)
		valueEnd, err := ctx.skipValueV2(valueStart, depth)
		if err != nil {
			// a syntax error of the value, which encoding/json/v2 reads whole: the v2 json package reads it whole
			// to find it again ( see readsWhole ).
			return 0, &errors.UnmarshalError{GoType: jsonfields.RawValueType, Start: valueStart, Err: err}
		}
		*b = append(*b, quoted...)
		*b = append(*b, ':')
		*b = append(*b, ctx.Buf[valueStart:valueEnd]...)
		*b = append(*b, '}')
		return valueEnd, nil
	}
	m := valueAt(ptrTypeOf(fb.mapType), target)
	if m.IsNil() {
		m.Set(reflect.MakeMap(fb.mapType))
	}
	key := reflect.ValueOf(string(unquotedName(quoted))).Convert(fb.mapType.Key())
	v := reflect.New(fb.mapType.Elem())
	if old := m.MapIndex(key); old.IsValid() {
		v.Elem().Set(old)
	}
	next, err := fb.value.Decode(ctx, cursor, depth, v.UnsafePointer())
	m.SetMapIndex(key, v.Elem())
	return next, err
}

// compileMapV2 returns the decoder of the map type typ: the names of an object are decoded as the values of the key
// type, but the numbers, which are in the names, as with StringifyNumbers.
func (c *compiler) compileMapV2(typ reflect.Type, structName, fieldName string) (Decoder, error) {
	if typ == interfaceMapType {
		// a function of unmarshaling which takes a string, an interface{} or a type of its values decodes the
		// names or the values: the map is then decoded as any other
		if iface := c.newV2InterfaceDecoder(emptyInterfaceType); !iface.fromAny {
			return &v2AnyMapDecoder{iface: iface}, nil
		}
	}
	d := &v2MapDecoder{typ: typ, ptrType: ptrTypeOf(typ)}
	if _, exists := c.decoders[uintptr(runtime.TypePtr(typ))]; !exists {
		// a decoder which declines values to it stays there ( see compileNextV2 )
		c.decoders[uintptr(runtime.TypePtr(typ))] = d
		defer delete(c.decoders, uintptr(runtime.TypePtr(typ)))
	}
	// the keys are compiled apart: a type of them, compiled for the names, is not the one of a value
	keyCompiler := &compiler{decoders: map[uintptr]Decoder{}, v2: true, stringifyNumbers: true, funcs: c.funcs}
	keyDec, err := keyCompiler.compileV2(typ.Key(), structName, fieldName)
	if err != nil {
		return nil, err
	}
	valueDec, err := compile(typ.Elem(), structName, fieldName, c)
	if err != nil {
		return nil, err
	}
	d.keyDecoder, d.valueDecoder = keyDec, valueDec
	return d, nil
}

// v2MapDecoder is the decoder of a map of v2. An object is merged into the map: the value of a name which the map
// has as a key is decoded into the value of the key. A name which is the key of an earlier member of the object, by
// its value, is an error unless V2AllowDuplicateNames. null sets the map to nil.
type v2MapDecoder struct {
	typ          reflect.Type
	ptrType      unsafe.Pointer
	keyDecoder   Decoder
	valueDecoder Decoder
}

var emptyStructType = reflect.TypeOf(struct{}{})

// uniqueKeyNames reports whether two names of an object which dec decodes into equal keys of the kind are equal: a
// bool, an integer, and a string, unless its invalid UTF-8 is replaced, which are decoded by their kind. A method
// or a function of unmarshaling may decode two names into one key, and one name into two keys.
func uniqueKeyNames(dec Decoder, kind reflect.Kind, allowInvalidUTF8 bool) bool {
	switch dec.(type) {
	case *v2FuncDecoder, *v2UnmarshalFromDecoder, *v2UnmarshalJSONDecoder, *v2UnmarshalTextDecoder:
		return false
	}
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	case reflect.String:
		return !allowInvalidUTF8
	}
	return false
}

func (d *v2MapDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		*(*unsafe.Pointer)(p) = nil
		return cursor + 4, nil
	case '{':
	default:
		return ctx.skipTypeError(cursor, depth, d.typ)
	}
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}
	m := valueAt(d.ptrType, p)
	if m.IsNil() {
		m.Set(reflect.MakeMap(d.typ))
	}
	cursor = skipWhiteSpace(buf, cursor+1)
	if buf[cursor] == '}' {
		return cursor + 1, nil
	}
	rejectDuplicateNames := ctx.Option.V2&V2AllowDuplicateNames == 0
	// the keys of the object, if the map had keys before, which a name of the object can't be told from by the map;
	// and the names of the object as they are written, which a key of another kind than a bool, an integer and a
	// string of valid UTF-8 doesn't tell apart, as a pointer, a float or an interface value, nor a key which a method
	// or a function decodes.
	var seen reflect.Value
	if rejectDuplicateNames && m.Len() > 0 {
		seen = reflect.MakeMap(reflect.MapOf(d.typ.Key(), emptyStructType))
	}
	// names are the names of the keys, if two keys may have the same name
	var names nameSet
	keyNames := rejectDuplicateNames && !uniqueKeyNames(d.keyDecoder, d.typ.Key().Kind(), ctx.Option.V2&V2AllowInvalidUTF8 != 0)
	k, v := reflect.New(d.typ.Key()), reflect.New(d.typ.Elem())
	for {
		start := cursor
		if buf[start] != '"' {
			return 0, syntaxErrorAt(buf, start, whereKey)
		}
		k.Elem().SetZero()
		next, err := d.keyDecoder.Decode(ctx, start, depth, k.UnsafePointer())
		if err != nil {
			return 0, err
		}
		if keyNames && !names.addCopy(ctx.V2Call(), nil, unquotedName(buf[start:next])) {
			return 0, &errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}
		}
		key := k.Elem()
		if old := m.MapIndex(key); old.IsValid() {
			if rejectDuplicateNames && (!seen.IsValid() || seen.MapIndex(key).IsValid()) {
				return 0, &errors.UnmarshalError{Start: start, Err: errors.ErrDuplicateName}
			}
			v.Elem().Set(old)
		} else {
			v.Elem().SetZero()
		}
		cursor = skipWhiteSpace(buf, next)
		if buf[cursor] != ':' {
			return 0, syntaxErrorAt(buf, cursor, whereAfterKey)
		}
		next, err = d.valueDecoder.Decode(ctx, cursor+1, depth, v.UnsafePointer())
		m.SetMapIndex(key, v.Elem())
		if seen.IsValid() {
			seen.SetMapIndex(key, reflect.Zero(emptyStructType))
		}
		if err != nil {
			return 0, err
		}
		cursor = skipWhiteSpace(buf, next)
		switch buf[cursor] {
		case '}':
			names.end(ctx.v2)
			return cursor + 1, nil
		case ',':
			cursor = skipWhiteSpace(buf, cursor+1)
		default:
			return 0, syntaxErrorAt(buf, cursor, whereAfterPair)
		}
	}
}

func (d *v2MapDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// v2AnyMapDecoder is the decoder of a map[string]interface{} of v2, which is decoded as the one of an interface{}: a
// nil map as a nil interface{}, by v1, and else merged ( see v2InterfaceDecoder.decodeMerging ).
type v2AnyMapDecoder struct {
	iface *v2InterfaceDecoder
}

func (d *v2AnyMapDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		*(*unsafe.Pointer)(p) = nil
		return cursor + 4, nil
	case '{':
	default:
		return ctx.skipTypeError(cursor, depth, interfaceMapType)
	}
	m := *(*map[string]any)(p)
	var v any
	var next int64
	var err error
	if m == nil {
		next, err = d.iface.Decode(ctx, cursor, depth, unsafe.Pointer(&v))
	} else {
		v = m
		next, err = d.iface.decodeMerging(ctx, cursor, depth, unsafe.Pointer(&v))
	}
	if m, ok := v.(map[string]any); ok {
		*(*map[string]any)(p) = m
	}
	return next, err
}

func (d *v2AnyMapDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}
