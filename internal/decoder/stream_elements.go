package decoder

import (
	stderrors "errors"
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonnum"
	"github.com/goccy/go-json/internal/runtime"
)

// A value of a stream is decoded from the buffer by the decoders of a byte slice when the buffer holds it whole. An
// object or an array which doesn't end within maxWholeValueSize bytes is decoded element by element instead: each
// element is decoded the same way, from the buffer if it holds the element whole, and the bytes of the elements
// decoded are dropped when the buffer is filled again. So the memory of a stream is bounded by a few times
// maxWholeValueSize, or by the longest string or number, whatever the size of a value. The strings decoded then are
// copied out of the buffer, which they would keep alive.
//
// The objects and the arrays of structs, maps, slices, arrays, pointers to them and interface{} values are decoded
// so. A value of any other type, as one which has an UnmarshalJSON method, which takes the bytes of the whole value,
// or a value of another kind than its type, whose type error the decoder of its type reports, is read whole.

// maxWholeValueSize is the size of the largest object or array which a stream reads whole: the buffer grows for one
// so long. A reader which tells the length of its data is read whole at once ( see grow ), which the buffer holds
// every value of then.
const maxWholeValueSize = 64 << 10

// errLongValue is what a scan of a value returns for an object or an array which doesn't end within the bytes it may
// read ( see scanValueUpTo ).
var errLongValue = stderrors.New("json: value too long to be read whole")

// scanFailed returns the error of Decode for the error err of the scan of the value at the cursor, which Decode decodes
// into p, of the pointer type typ, by dec: a value too long to be read whole is decoded element by element, as Decode
// does a value whole, or read whole and decoded if dec doesn't decode it so.
//
//go:noinline
func (s *Stream) scanFailed(err error, dec Decoder, typ, p unsafe.Pointer) error {
	if err != errLongValue {
		return s.fail(err)
	}
	ctx := s.ctx
	ctx.Option = s.Option
	start := s.offset + s.cursor
	s.index = &bracketIndex{}
	s.indexDeeper(0)
	decoded, err := s.decodeElements(dec, p, 0)
	s.index = nil
	if !decoded {
		var end int64
		if end, err = s.scanValue(); err == nil {
			err = s.decodeWhole(dec, p, 0, end)
		}
	}
	if err != nil {
		ctx.DiscardTypeError()
		return s.fail(err)
	}
	if countsScanned {
		s.scanned += s.offset + s.cursor - s.readStart
	}
	s.more.reset()
	if s.tokenState != tokenTopValue {
		s.tokenValueEnd()
	}
	if e := ctx.typeError; e != nil {
		ctx.typeError = nil
		if e.plain != nil {
			return e.plain
		}
		return e.toError(e.path, runtime.TypeOfPtr(typ), StreamOffsetBase(s.prevEnd, start))
	}
	return nil
}

// decodeElements decodes the object or the array at the cursor, nested in depth objects and arrays, by dec into p
// element by element. It reports false, having read nothing, if dec doesn't decode it so.
func (s *Stream) decodeElements(dec Decoder, p unsafe.Pointer, depth int64) (bool, error) {
	c := s.buf[s.cursor]
	switch d := dec.(type) {
	case *structDecoder:
		if c == '{' {
			return true, s.decodeStructElements(d, p, depth)
		}
	case *mapDecoder:
		if c == '{' && !d.keyUnsupported {
			return true, s.decodeMapElements(d, p, depth)
		}
	case *sliceDecoder:
		if c == '[' {
			return true, s.decodeSliceElements(d, p, depth)
		}
	case *arrayDecoder:
		if c == '[' {
			return true, s.decodeArrayElements(d, p, depth)
		}
	case *ptrDecoder:
		ptr := *(*unsafe.Pointer)(p)
		if ptr == nil {
			ptr = newValue(d.typ)
			*(*unsafe.Pointer)(p) = ptr
		}
		decoded, err := s.decodeElements(d.dec, ptr, depth)
		if err != nil {
			*(*unsafe.Pointer)(p) = nil
		}
		return decoded, err
	case *anonymousFieldDecoder:
		if *(*unsafe.Pointer)(p) == nil {
			*(*unsafe.Pointer)(p) = newValue(d.structType)
		}
		return s.decodeElements(d.dec, unsafe.Add(*(*unsafe.Pointer)(p), d.offset), depth)
	case *interfaceDecoder:
		return s.decodeInterfaceElements(d, p, depth)
	}
	return false, nil
}

// decodeNested decodes the value at the cursor, nested in depth objects and arrays, by dec into p: whole if the
// buffer holds it, and else element by element.
func (s *Stream) decodeNested(dec Decoder, p unsafe.Pointer, depth int64) error {
	end, whole, err := s.elementEnd(depth, maxWholeValueSize)
	if err != nil {
		return err
	}
	if !whole {
		if decoded, err := s.decodeElements(dec, p, depth); decoded {
			return err
		}
		if end, _, err = s.elementEnd(depth, math.MaxInt64); err != nil {
			return err
		}
	}
	return s.decodeWhole(dec, p, depth, end)
}

// elementEnd returns the position after the value at the cursor, nested in depth objects and arrays, reading from the
// reader as needed. An object or an array is found by the index of the brackets if it ends within limit bytes, and
// else whole is false; any other value is read whole.
func (s *Stream) elementEnd(depth, limit int64) (int64, bool, error) {
	if c := s.buf[s.cursor]; c != '{' && c != '[' {
		end, err := s.scanValueUpTo(limit)
		return end, true, err
	}
	var rel int
	for {
		var end int64
		var ok bool
		if end, ok, rel = s.index.closeOf(s.offset+s.cursor, depth, rel); ok {
			return end + 1 - s.offset, true, nil
		}
		if s.length-s.cursor >= limit {
			return 0, false, nil
		}
		if !s.read() {
			return 0, false, s.truncatedValueError()
		}
	}
}

// bracketIndex is where the objects and the arrays of the value which a stream decodes element by element end, in
// what has been read: it is extended by every read, so that each byte is scanned once, and an element is found whole
// in the buffer by the index, without a scan of its own at every level it is nested in. Only the ones nested in at
// most bound objects and arrays have an entry: the elements looked up are the ones of the objects and the arrays
// decoded element by element, which are few levels, and the brackets of the blocks of bytes deeper in an element
// are not looked at one by one.
type bracketIndex struct {
	// closes are the brackets which close the objects and the arrays, in their order. An object or an array ends at
	// the first one after it which is as deep: the ones before are deeper, in it.
	closes []bracketClose
	// depth is the number of the objects and the arrays open after what has been read.
	depth int64
	// bound is the depth of the deepest brackets indexed: the depth of the closes ( see bracketClose ).
	bound int64
	// next is the first entry which closeOf looks at: the values are looked up in the order of the input.
	next int
	// inString and escaped are the state of the scan after what has been read ( see compoundScanner.scan ).
	inString, escaped uint64
	// buf is the buffer which the stream made while the value is decoded element by element: the strings decoded
	// then are copied, so nothing refers to it, and the stream reuses it in place ( see grow ).
	buf *byte
}

// bracketClose is a bracket which closes an object or an array: its total offset, and the number of the objects and
// the arrays which the one it closes is nested in, plus one for itself.
type bracketClose struct {
	at, depth int64
}

// indexRead extends the index by the n bytes which a read put after the data. It is not inlined, so that read keeps
// the size it had.
//
//go:noinline
func (s *Stream) indexRead(n int) {
	s.index.extend(s.buf, s.length, s.length+int64(n), s.offset)
}

// indexDeeper makes the index have the elements of the object or the array at the cursor, which is nested in depth
// objects and arrays: the bytes read from it are indexed again, to the depth of its elements or twice the depth of
// before, so that a value nested deep, which is decoded element by element at every level, has its bytes indexed
// again a few times only.
func (s *Stream) indexDeeper(depth int64) {
	x := s.index
	at := s.offset + s.cursor
	n := len(x.closes)
	for n > x.next && x.closes[n-1].at > at {
		n--
	}
	x.closes = x.closes[:n]
	x.depth, x.bound, x.inString, x.escaped = depth, max(depth+2, 2*x.bound), 0, 0
	x.extend(s.buf, s.cursor, s.length, s.offset)
}

// owns reports whether buf is the buffer which the stream made while the value is decoded element by element.
func (x *bracketIndex) owns(buf []byte) bool {
	return x.buf == unsafe.SliceData(buf)
}

// extend indexes buf[from:to], whose first byte is at the total offset offset.
func (x *bracketIndex) extend(buf []byte, from, to, offset int64) {
	if x.next >= 1024 && x.next*2 >= len(x.closes) {
		// the entries which closeOf looked past are dropped
		x.closes = x.closes[:copy(x.closes, x.closes[x.next:])]
		x.next = 0
	}
	const evenBits = 0x5555555555555555
	base := (*sliceHeader)(unsafe.Pointer(&buf)).data
	closes, depth, bound, inString, prevEscaped := x.closes, x.depth, x.bound, x.inString, x.escaped
	for pos := from; pos < to; pos += scanBlockSize {
		// the brackets out of the strings in the block, as compoundScanner.scan finds them
		var m scanMasks
		scanBlock(unsafe.Add(base, pos), &m)
		n := to - pos
		if n < scanBlockSize {
			valid := uint64(1)<<uint(n) - 1
			m.quote &= valid
			m.backslash &= valid
			m.open &= valid
			m.closing &= valid
		}
		backslash := m.backslash &^ prevEscaped
		followsEscape := backslash<<1 | prevEscaped
		oddStarts := backslash &^ evenBits &^ followsEscape
		evenSequences, carry := bits.Add64(oddStarts, backslash, 0)
		escaped := (evenBits ^ (evenSequences << 1)) & followsEscape
		if n < scanBlockSize {
			prevEscaped = (escaped >> uint(n)) & 1
		} else {
			prevEscaped = carry
		}
		inside := m.quote &^ escaped
		inside ^= inside << 1
		inside ^= inside << 2
		inside ^= inside << 4
		inside ^= inside << 8
		inside ^= inside << 16
		inside ^= inside << 32
		inside ^= inString
		inString = uint64(int64(inside) >> 63)
		outside := ^(inside | escaped)
		open, closing := m.open&outside, m.closing&outside
		if closing == 0 || depth-int64(bits.OnesCount64(closing)) >= bound {
			// no bracket of the block closes an object or an array as shallow as the bound
			depth += int64(bits.OnesCount64(open)) - int64(bits.OnesCount64(closing))
			continue
		}
		for brackets := open | closing; brackets != 0; brackets &= brackets - 1 {
			i := bits.TrailingZeros64(brackets)
			if open&(1<<uint(i)) != 0 {
				depth++
				continue
			}
			if depth <= bound {
				closes = append(closes, bracketClose{at: offset + pos + int64(i), depth: depth})
			}
			depth--
		}
	}
	x.closes, x.depth, x.inString, x.escaped = closes, depth, inString, prevEscaped
}

// closeOf returns the total offset of the bracket which closes the object or the array whose bracket is at the total
// offset at, nested in depth objects and arrays, if it has been read. The entries are looked at from the one rel
// after the first one which may close the value, which a lookup returns, so that the lookups of a value which has not
// been read whole go on after each other.
func (x *bracketIndex) closeOf(at, depth int64, rel int) (int64, bool, int) {
	for x.next < len(x.closes) && x.closes[x.next].at < at {
		x.next++
	}
	for i := x.next + rel; i < len(x.closes); i++ {
		if x.closes[i].depth == depth+1 {
			return x.closes[i].at, true, 0
		}
	}
	return 0, false, len(x.closes) - x.next
}

// decodeWhole decodes the value at the cursor, which ends at end, nested in depth objects and arrays, by dec into p,
// as Decode does a value.
func (s *Stream) decodeWhole(dec Decoder, p unsafe.Pointer, depth, end int64) error {
	ctx := s.ctx
	had := ctx.typeError
	cursor, err := s.callDecoder(dec, p, depth, end, true)
	if err != nil {
		if c := s.buf[s.cursor]; c == '-' || c-'0' <= 9 {
			// 01 is 0 and 1 ( see decodeError ): the first is decoded alone, and the second is an error of the
			// object or the array
			if n, _ := jsonnum.Scan(s.buf[s.cursor:end]); s.cursor+int64(n) < end {
				ctx.typeError = had
				cursor, err = s.callDecoder(dec, p, depth, s.cursor+int64(n), true)
			}
		}
		if err != nil {
			return s.elementError(err)
		}
	}
	if had == nil && ctx.typeError != nil {
		s.keepTypeError(dec)
	}
	s.cursor = cursor
	return nil
}

// elementError returns the error of the decoding of the value at the cursor, as decodeError does: a syntax error is
// the first one of the value by the grammar.
func (s *Stream) elementError(err error) error {
	if _, ok := err.(*errors.SyntaxError); ok {
		if serr := s.valueSyntaxError(); serr != nil {
			return serr
		}
	}
	return s.totalOffsetError(err)
}

// keepTypeError makes the type error which dec recorded in the value at the cursor one of the whole input: its path
// from the value is made while the buffer holds the value, and the objects and the arrays the value is in add theirs
// to it ( see addTypeErrorStep ).
func (s *Stream) keepTypeError(dec Decoder) {
	e := s.ctx.typeError
	if e.plain == nil {
		e.path = typeErrorPath(dec, s.buf[:s.length+1], s.cursor, e)
	}
	e.start += s.offset
	e.end += s.offset
}

// addTypeErrorStep adds the step to an element to the path of the type error recorded in it: the decoding of the
// element had none before, which it is called for.
func (s *Stream) addTypeErrorStep(step typeErrorStep) {
	e := s.ctx.typeError
	e.path = append([]typeErrorStep{step}, e.path...)
}

// syntaxError returns the syntax error of the byte at the cursor, which the grammar doesn't have where it is.
func (s *Stream) syntaxError(where syntaxWhere) error {
	s.fillLastChar()
	return s.grammarError(syntaxErrorAt(s.buf[:s.length+1], s.cursor, where))
}

// nextElement moves the cursor to the next element of the array whose elements are read: first is set before the
// first one. It reports false after the end of the array.
func (s *Stream) nextElement(first bool) (bool, error) {
	if !s.skipWhiteSpace() {
		return false, s.unexpectedEndError()
	}
	switch c := s.buf[s.cursor]; {
	case c == ']':
		s.cursor++
		return false, nil
	case first:
		return true, nil
	case c != ',':
		return false, s.syntaxError(whereAfterElement)
	}
	s.cursor++
	if !s.skipWhiteSpace() {
		return false, s.unexpectedEndError()
	}
	return true, nil
}

// nextMember moves the cursor after the key of the next member of the object whose members are read: first is set
// before the first one. It returns the key as the buffer has it, which the next read of the stream may drop, what it
// has, and the position it starts at. It reports false after the end of the object.
func (s *Stream) nextMember(first bool) ([]byte, stringInfo, int64, bool, error) {
	if !s.skipWhiteSpace() {
		return nil, stringInfo{}, 0, false, s.unexpectedEndError()
	}
	switch c := s.buf[s.cursor]; {
	case c == '}':
		s.cursor++
		return nil, stringInfo{}, 0, false, nil
	case first:
	case c != ',':
		return nil, stringInfo{}, 0, false, s.syntaxError(whereAfterPair)
	default:
		s.cursor++
		if !s.skipWhiteSpace() {
			return nil, stringInfo{}, 0, false, s.unexpectedEndError()
		}
	}
	if s.buf[s.cursor] != '"' {
		return nil, stringInfo{}, 0, false, s.syntaxError(whereKey)
	}
	if _, err := s.scanString(); err != nil {
		return nil, stringInfo{}, 0, false, err
	}
	start := s.cursor
	key, next, info, err := skipStringDecoder.scanString(s.buf[:s.length+1], start)
	if next < 0 {
		key, next, info, err = skipStringDecoder.scanStringRest(s.buf[:s.length+1], key, -next-1, info)
	}
	if err != nil {
		return nil, stringInfo{}, 0, false, s.elementError(err)
	}
	s.cursor = next
	return key, info, start, true, nil
}

// openElements moves the cursor after the bracket which opens the object or the array whose elements are read, nested
// in depth objects and arrays, and returns the depth of its elements. The depth is checked as the scan of a value
// checks it ( see compoundScanner ), which a value decoded element by element has no scan of.
func (s *Stream) openElements(depth int64) (int64, error) {
	if depth >= maxDecodeNestingDepth {
		return 0, s.totalOffsetError(errors.ErrExceededMaxDepth(s.buf[s.cursor], s.cursor))
	}
	if s.index.bound < depth+2 {
		// the elements are looked up in the index ( see elementEnd )
		s.indexDeeper(depth)
	}
	s.cursor++
	return depth + 1, nil
}

// memberValue moves the cursor from after the key of a member to its value.
func (s *Stream) memberValue() error {
	if !s.skipWhiteSpace() {
		return s.unexpectedEndError()
	}
	if s.buf[s.cursor] != ':' {
		return s.syntaxError(whereAfterKey)
	}
	s.cursor++
	if !s.skipWhiteSpace() {
		return s.unexpectedEndError()
	}
	return nil
}

// skipNested skips the value at the cursor, nested in depth objects and arrays, which is checked by the grammar, as
// skipValue does: whole if the buffer holds it, and else element by element.
func (s *Stream) skipNested(depth int64) error {
	if _, whole, err := s.elementEnd(depth, maxWholeValueSize); err != nil || whole {
		if err != nil {
			return err
		}
		next, err := skipValue(s.buf[:s.length+1], s.cursor, depth)
		if err != nil {
			return s.grammarError(err)
		}
		s.cursor = next
		return nil
	}
	object := s.buf[s.cursor] == '{'
	depth, err := s.openElements(depth)
	if err != nil {
		return err
	}
	for first := true; ; first = false {
		var ok bool
		var err error
		if object {
			if _, _, _, ok, err = s.nextMember(first); ok {
				err = s.memberValue()
			}
		} else {
			ok, err = s.nextElement(first)
		}
		if !ok || err != nil {
			return err
		}
		if err := s.skipNested(depth); err != nil {
			return err
		}
	}
}

// decodeStructElements decodes the object at the cursor into the struct at p, as structDecoder.Decode does.
func (s *Stream) decodeStructElements(d *structDecoder, p unsafe.Pointer, depth int64) error {
	depth, err := s.openElements(depth)
	if err != nil {
		return err
	}
	flags := s.ctx.Option.Flags
	disallowUnknownFields := flags&DisallowUnknownFieldsOption != 0
	// seen are the fields decoded under FirstWinOption, by their indexes, and seenNum their number: once every one
	// is decoded, the rest of the object is skipped, as decodeObjectFirstWin does.
	var seen []uint64
	if flags&FirstWinOption != 0 {
		seen = make([]uint64, (d.fieldUniqueNameNum+63)/64)
	}
	seenNum := 0
	var arr [64]byte
	for first := true; ; first = false {
		key, info, _, ok, err := s.nextMember(first)
		if !ok {
			return err
		}
		var field *structFieldSet
		var inputName []byte
		if seenNum == 0 || seenNum < d.fieldUniqueNameNum {
			var decodedKey []byte
			field, decodedKey = d.keys.lookup(key, info)
			if field == nil && disallowUnknownFields {
				return unknownFieldError(decodedKey)
			}
			inputName = append(arr[:0], decodedKey...)
		}
		unsettable := field != nil && field.unsettable(p)
		if seen != nil && field != nil && !unsettable {
			w, bit := &seen[field.fieldIdx/64], uint64(1)<<uint(field.fieldIdx%64)
			if *w&bit != 0 {
				field = nil
			} else {
				*w |= bit
				seenNum++
			}
		}
		if err := s.memberValue(); err != nil {
			return err
		}
		had := s.ctx.typeError
		switch {
		case field == nil:
			err = s.skipNested(depth)
		case unsettable:
			err = s.decodeNested(unsettableFieldDecoder{d: d, err: field.err}, nil, depth)
		default:
			err = s.decodeNested(field.dec, unsafe.Add(p, field.offset), depth)
		}
		if err != nil {
			return err
		}
		if s.ctx.typeError != had {
			s.addTypeErrorStep(d.typeErrorStep(field, inputName))
		}
	}
}

// unsettableFieldDecoder is the decoder of the value of a field which can't be set ( see structFieldSet.unsettable ),
// which records its error and skips the value, as structDecoder.Decode does.
type unsettableFieldDecoder struct {
	d   *structDecoder
	err error
}

func (u unsettableFieldDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, _ unsafe.Pointer) (int64, error) {
	return u.d.unsettableField(ctx, cursor, depth, u.err)
}

func (u unsettableFieldDecoder) DecodePath(*RuntimeContext, int64, int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: unsettable field decoder does not support decode path")
}

// decodeMapElements decodes the object at the cursor into the map at p, as mapDecoder.Decode does.
func (s *Stream) decodeMapElements(d *mapDecoder, p unsafe.Pointer, depth int64) error {
	depth, err := s.openElements(depth)
	if err != nil {
		return err
	}
	mapValue := *(*unsafe.Pointer)(p)
	if mapValue == nil {
		mapValue = reflect.MakeMap(d.mapType).UnsafePointer()
	}
	t := d.temps.Get().(*mapTemps)
	mv := d.mapValue(mapValue)
	var arr [64]byte
	for first := true; ; first = false {
		key, info, start, ok, err := s.nextMember(first)
		if err != nil {
			return err
		}
		if !ok {
			*(*unsafe.Pointer)(p) = mapValue
			d.temps.Put(t)
			return nil
		}
		// the name of the entry, which a type error in it reports
		name := append(arr[:0], key...)
		had := s.ctx.typeError
		// the key is decoded by the decoder of the keys, as a value which ends after it
		end := s.cursor
		s.cursor = start
		keyEnd, err := s.callDecoder(d.keyDecoder, t.k, depth, end, true)
		drop := err == errMapKeyType
		if err != nil && !drop {
			return s.elementError(err)
		}
		if had == nil && s.ctx.typeError != nil {
			s.keepTypeError(d.keyDecoder)
		}
		s.cursor = keyEnd
		if err := s.memberValue(); err != nil {
			return err
		}
		if err := s.decodeNested(d.valueDecoder, t.v, depth); err != nil {
			return err
		}
		if !drop {
			mv.SetMapIndex(t.kv, t.vv)
		}
		t.kv.SetZero()
		t.vv.SetZero()
		if s.ctx.typeError != had {
			s.addTypeErrorStep(typeErrorStep{name: string(decodeLiteral(name, info))})
		}
	}
}

// decodeSliceElements decodes the array at the cursor into the slice at p, as sliceDecoder.Decode does.
func (s *Stream) decodeSliceElements(d *sliceDecoder, p unsafe.Pointer, depth int64) error {
	depth, err := s.openElements(depth)
	if err != nil {
		return err
	}
	dst := (*sliceHeader)(p)
	// the elements are decoded into the slice up to limit, and into a buffer after it, as Decode does
	limit := max(dst.cap, inPlaceSliceLength)
	var elems *sliceBuf
	for idx := 0; ; idx++ {
		ok, err := s.nextElement(idx == 0)
		if err != nil || !ok {
			switch {
			case elems != nil:
				d.storeBuf(dst, elems, idx)
			case err == nil && idx == 0 && dst.data == nil:
				dst.data = unsafe.Pointer(&zeroBase)
			case err == nil:
				dst.len = idx
			}
			return err
		}
		var ep unsafe.Pointer
		switch {
		case idx < limit:
			d.grow(dst, idx, idx+1)
			ep = unsafe.Add(dst.data, uintptr(idx)*d.size)
			dst.len = idx + 1
		case elems == nil:
			elems = d.takeBuf(dst, idx)
			fallthrough
		default:
			ep = d.bufElem(elems, idx)
		}
		had := s.ctx.typeError
		if err := s.decodeNested(d.valueDecoder, ep, depth); err != nil {
			if elems != nil {
				d.storeBuf(dst, elems, idx+1)
			}
			return err
		}
		if s.ctx.typeError != had {
			s.addTypeErrorStep(typeErrorStep{name: strconv.Itoa(idx)})
		}
	}
}

// decodeArrayElements decodes the array at the cursor into the array at p, as arrayDecoder.Decode does.
func (s *Stream) decodeArrayElements(d *arrayDecoder, p unsafe.Pointer, depth int64) error {
	depth, err := s.openElements(depth)
	if err != nil {
		return err
	}
	for idx := 0; ; idx++ {
		ok, err := s.nextElement(idx == 0)
		if err != nil {
			return err
		}
		if !ok {
			d.zeroFrom(p, idx)
			return nil
		}
		if idx >= d.alen {
			if err := s.skipNested(depth); err != nil {
				return err
			}
			continue
		}
		had := s.ctx.typeError
		if err := s.decodeNested(d.valueDecoder, unsafe.Add(p, uintptr(idx)*d.size), depth); err != nil {
			return err
		}
		if s.ctx.typeError != had {
			s.addTypeErrorStep(typeErrorStep{name: strconv.Itoa(idx)})
		}
	}
}

// decodeInterfaceElements decodes the object or the array at the cursor into the interface{} at p, as
// interfaceDecoder.Decode does. It reports false for an interface type with methods, whose value is a type error,
// and for an interface value which holds a pointer, which the value is decoded into.
func (s *Stream) decodeInterfaceElements(d *interfaceDecoder, p unsafe.Pointer, depth int64) (bool, error) {
	if d.hasMethods {
		return false, nil
	}
	if h := (*emptyInterface)(p); h.ptr != nil {
		v := d.values.Load(uintptr(h.typ))
		if v == nil {
			var err error
			if v, err = d.valueDecoderOf(s.ctx, h.typ, *(*any)(p)); err != nil {
				return true, err
			}
		}
		if v.kind == ifaceValuePointee {
			// what the pointer points to is read whole, as the path of a type error in it is made by the interface
			// value ( see typeErrorPath )
			return false, nil
		}
	}
	var v any
	var err error
	if s.buf[s.cursor] == '{' {
		v, err = s.decodeAnyObject(d, depth)
	} else {
		v, err = s.decodeAnyArray(d, depth)
	}
	if err != nil {
		return true, err
	}
	*(*any)(p) = v
	return true, nil
}

// decodeAnyObject decodes the object at the cursor into a new map[string]interface{}, as decodeNewStringAnyMap does:
// the entries are pushed to the stack of entries of the context, and put into a map of their number at the end.
func (s *Stream) decodeAnyObject(d *interfaceDecoder, depth int64) (any, error) {
	depth, err := s.openElements(depth)
	if err != nil {
		return nil, err
	}
	ctx := s.ctx
	base := len(ctx.entryStack)
	for first := true; ; first = false {
		key, info, _, ok, err := s.nextMember(first)
		if !ok {
			if err != nil {
				ctx.popEntries(base)
				return nil, err
			}
			entries := ctx.entryStack[base:]
			m := make(map[string]any, len(entries))
			for i := range entries {
				m[entries[i].key] = entries[i].value
			}
			ctx.popEntries(base)
			return m, nil
		}
		name := ctx.copyString(decodeLiteral(key, info))
		if err := s.memberValue(); err != nil {
			ctx.popEntries(base)
			return nil, err
		}
		had := ctx.typeError
		if err := s.decodeNested(d, unsafe.Pointer(&ctx.slot), depth); err != nil {
			ctx.slot = nil
			ctx.popEntries(base)
			return nil, err
		}
		ctx.entryStack = append(ctx.entryStack, anyEntry{key: name, value: ctx.slot})
		ctx.slot = nil
		if ctx.typeError != had {
			s.addTypeErrorStep(typeErrorStep{name: name})
		}
	}
}

// decodeAnyArray decodes the array at the cursor into a new []interface{}, as decodeAnySlice does: the elements are
// pushed to the stack of the context, and copied into a slice of their number at the end.
func (s *Stream) decodeAnyArray(d *interfaceDecoder, depth int64) (any, error) {
	depth, err := s.openElements(depth)
	if err != nil {
		return nil, err
	}
	ctx := s.ctx
	base := len(ctx.anyStack)
	for idx := 0; ; idx++ {
		ok, err := s.nextElement(idx == 0)
		if err != nil {
			ctx.popAny(base)
			return nil, err
		}
		if !ok {
			if idx == 0 {
				return emptyAnySlice, nil
			}
			elems := make([]any, idx)
			copy(elems, ctx.anyStack[base:])
			ctx.popAny(base)
			return elems, nil
		}
		had := ctx.typeError
		if err := s.decodeNested(d, unsafe.Pointer(&ctx.slot), depth); err != nil {
			ctx.slot = nil
			ctx.popAny(base)
			return nil, err
		}
		ctx.anyStack = append(ctx.anyStack, ctx.slot)
		ctx.slot = nil
		if ctx.typeError != had {
			s.addTypeErrorStep(typeErrorStep{name: strconv.Itoa(idx)})
		}
	}
}
