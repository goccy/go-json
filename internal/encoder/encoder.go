package encoder

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"math"
	"math/bits"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/floatfmt"
	"github.com/goccy/go-json/internal/jsonnum"
	"github.com/goccy/go-json/internal/jsonstring"
	"github.com/goccy/go-json/internal/runtime"
)

func (t OpType) IsMultipleOpField() bool {
	switch t {
	case OpStructField:
		return true
	case OpStructFieldSlice:
		return true
	case OpStructFieldArray:
		return true
	case OpStructFieldMap:
		return true
	case OpStructFieldStruct:
		return true
	case OpStructFieldOmitEmpty:
		return true
	case OpStructFieldOmitEmptySlice:
		return true
	case OpStructFieldOmitEmptyArray:
		return true
	case OpStructFieldOmitEmptyMap:
		return true
	case OpStructFieldOmitEmptyStruct:
		return true
	case OpStructFieldSlicePtr:
		return true
	case OpStructFieldOmitEmptySlicePtr:
		return true
	case OpStructFieldArrayPtr:
		return true
	case OpStructFieldOmitEmptyArrayPtr:
		return true
	case OpStructFieldMapPtr:
		return true
	case OpStructFieldOmitEmptyMapPtr:
		return true
	}
	return false
}

type OpcodeSet struct {
	Type reflect.Type
	// IfaceIndir is whether a value of Type is stored indirectly in an interface value.
	// It is decided when the type is compiled, because deciding it is not cheap.
	IfaceIndir bool
	// DataWordIsAddr is whether the data word of an interface value of Type is the address of the value
	// which the opcodes take: the type is stored indirectly, or it is a pointer, which is the address of
	// the value it points to.
	DataWordIsAddr bool
	// IdentityIsFirstWord is whether a value of Type is a reference whose identity is its first word, which a
	// cycle passes again and again: a pointer, a map, a slice ( the address of its array ), or a type stored
	// directly in an interface value. The identity of another value is its address.
	IdentityIsFirstWord      bool
	NoescapeKeyCode          *Opcode
	EscapeKeyCode            *Opcode
	InterfaceNoescapeKeyCode *Opcode
	InterfaceEscapeKeyCode   *Opcode
	CodeLength               int
	EndCode                  *Opcode
	// Scalar is the opcode of the value if the type is encoded by a single opcode of a scalar ( a number,
	// a string, a bool, ... ), or nil. Such a value held by an interface value is encoded without a frame.
	Scalar     *Opcode
	Code       Code
	QueryCache map[string]*OpcodeSet
	// values is the pool of the values of Type in the heap, which MarshalOf copies its argument to.
	values  sync.Pool
	cacheMu sync.RWMutex
}

// ValueShape classifies a type by what a nil data word of an interface value of the type means.
type ValueShape uint

const (
	// ValueShapePointer is the type whose nil data word is encoded as null: a pointer, a map, ...
	ValueShapePointer ValueShape = iota
	// ValueShapeAggregate is the type whose nil data word is a value to encode, if the type is stored directly
	// in an interface value:
	//   - a struct or an array which consists of a single pointer. The pointer is nil, not the value.
	//   - a map which has a marshaler. encoding/json writes null instead of calling the marshaler
	//     only for a nil pointer, so the marshaler of a nil map is called.
	ValueShapeAggregate
)

// ShapeOf returns the shape of the type given by the pointer to its type descriptor.
//
// ShapeOf and IfaceIndir are functions of their own, not a part of the VM, and the VM calls them in the
// same form as it did before, because any change of the code of the VM changes the register allocation
// of the whole VM.
//
//go:noinline
func ShapeOf(typ unsafe.Pointer) ValueShape {
	switch runtime.TypeOfPtr(typ).Kind() {
	case reflect.Struct, reflect.Array:
		return ValueShapeAggregate
	case reflect.Map:
		// whether the type has a marshaler was decided when the type was compiled.
		codeSet, err := compileToGetUnfilteredCodeSet(uintptr(typ), 0)
		if err != nil {
			// the VM compiles the type right after this, and it reports the error.
			return ValueShapeAggregate
		}
		switch codeSet.Code.Kind() {
		case CodeKindMarshalJSON, CodeKindMarshalText:
			return ValueShapeAggregate
		}
	}
	return ValueShapePointer
}

// IfaceIndir reports whether a value of the type is stored indirectly in an interface value.
// It is decided only once per type, when the type is compiled.
//
//go:noinline
func IfaceIndir(typ unsafe.Pointer) bool {
	codeSet, err := compileToGetUnfilteredCodeSet(uintptr(typ), 0)
	if err != nil {
		// the VM compiles the type right after this, and it reports the error.
		return false
	}
	return codeSet.IfaceIndir
}

// TakeValue returns the address of a zero value of Type in the heap.
//
// The runtime context keeps the value of the type it encoded last, because taking one from a sync.Pool costs
// as much as a small allocation, and a runtime context is already taken from a pool.
func (s *OpcodeSet) TakeValue(ctx *RuntimeContext) unsafe.Pointer {
	if ctx.valueCodeSet == s {
		return ctx.value
	}
	if ctx.valueCodeSet != nil {
		ctx.valueCodeSet.values.Put(ctx.value)
	}
	ctx.valueCodeSet = s
	if p := s.values.Get(); p != nil {
		ctx.value = p.(unsafe.Pointer)
	} else {
		ctx.value = reflect.New(s.Type).UnsafePointer()
	}
	return ctx.value
}

func (s *OpcodeSet) getQueryCache(hash string) *OpcodeSet {
	s.cacheMu.RLock()
	codeSet := s.QueryCache[hash]
	s.cacheMu.RUnlock()
	return codeSet
}

func (s *OpcodeSet) setQueryCache(hash string, codeSet *OpcodeSet) {
	s.cacheMu.Lock()
	s.QueryCache[hash] = codeSet
	s.cacheMu.Unlock()
}

type CompiledCode struct {
	Code    *Opcode
	Linked  bool // whether recursive code already have linked
	CurLen  uintptr
	NextLen uintptr
	// Embedded is whether the recursive struct is embedded in the struct which jumps to it.
	// The code to jump to is only the fields of the struct then: it has neither the braces nor the check of nil.
	Embedded bool
}

// StartDetectingCyclesAfter is the number of the frames ( see RuntimeContext.RecursiveLevel ) after which the
// values entered, and the maps, are recorded to detect a cycle.
//
// It is 998, not the 1000 of encoding/json, as it counts frames, not the depth of the JSON value: the top-level value
// is written in the frame of the caller, and the check is made before the frame of a value is entered, so the
// level is the depth minus 1 there. With 998, the first value recorded is at the depth 1000 and a cycle is found at
// the depth 1001, where encoding/json/v2 reports it: its byte offset and JSON pointer are the same.
const StartDetectingCyclesAfter = 998

// StartDetectingMapCyclesAfter is StartDetectingCyclesAfter for a map ( see RuntimeContext.RecordMap ), whose check
// is made in the frame of its value, one level past the check of a frame, which is made before it is entered.
const StartDetectingMapCyclesAfter = StartDetectingCyclesAfter + 1

// ErrUnsupportedValue returns the error of a cycle through the value at ptr, of the type of code: of the pointer
// to it for the opcode of a pointer.
func ErrUnsupportedValue(ctx *RuntimeContext, code *Opcode, ptr unsafe.Pointer) error {
	typ := runtime.TypeOfPtr(code.Type)
	if code.Op == OpRecursivePtr && ctx.Option.Flag&V2Option != 0 {
		return &errors.SemanticError{GoType: reflect.PointerTo(typ), Err: errors.ErrCycle}
	}
	return errCycle(ctx, typ, ptr)
}

// errCycle returns the error of a cycle through the value at p, of the type typ.
func errCycle(ctx *RuntimeContext, typ reflect.Type, p unsafe.Pointer) error {
	if ctx.Option.Flag&V2Option != 0 {
		return &errors.SemanticError{GoType: typ, Err: errors.ErrCycle}
	}
	return &errors.UnsupportedValueError{
		Value: reflect.NewAt(typ, p).Elem(),
		Str:   fmt.Sprintf("encountered a cycle via %s", typ),
	}
}

// ErrUnsupportedFloat returns the error of a float which is not finite, the value of the opcode.
func ErrUnsupportedFloat(ctx *RuntimeContext, code *Opcode, v float64) error {
	if ctx.Option.Flag&V2Option != 0 {
		return &errors.SemanticError{GoType: runtime.TypeOfPtr(code.Type), Err: fmt.Errorf("unsupported value: %v", v)}
	}
	return &errors.UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   strconv.FormatFloat(v, 'g', -1, 64),
	}
}

func ErrMarshalerWithCode(code *Opcode, err error) *errors.MarshalerError {
	return &errors.MarshalerError{
		Type: runtime.TypeOfPtr(code.Type),
		Err:  err,
	}
}

type emptyInterface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

type MapItem struct {
	Key   []byte
	Value []byte
}

type Mapslice struct {
	Items []MapItem
	// escaped is whether a text was escaped while the entries of a map whose keys are texts were encoded ( see
	// appendText ): a key, or a text of a value, which tells that a key may have an escape.
	escaped bool
}

// Sort sorts the items by their keys, which are the names of the keys as encoding/json sorts them: an encoded
// key is in the order of its name unless it has an escape, and then the names are decoded to be sorted by.
// A text which was escaped while the entries were encoded tells that a key may have one ( see appendText ).
//
// It is not sort.Sort, which calls Less and Swap through an interface for every comparison:
// the maps to encode are small in most cases, and the sort was a tenth of the time to encode one.
func (m *Mapslice) Sort() {
	items := m.Items
	if m.escaped {
		sortMapItemsByNames(items)
		return
	}
	if len(items) > maxItemsOfInsertionSort {
		slices.SortFunc(items, func(a, b MapItem) int {
			return bytes.Compare(a.Key, b.Key)
		})
		return
	}
	insertionSortMapItems(items)
}

func insertionSortMapItems(items []MapItem) {
	for i := 1; i < len(items); i++ {
		if bytes.Compare(items[i-1].Key, items[i].Key) <= 0 {
			continue
		}
		item := items[i]
		j := i
		for ; j > 0 && bytes.Compare(items[j-1].Key, item.Key) > 0; j-- {
			items[j] = items[j-1]
		}
		items[j] = item
	}
}

// sortMapItemsByNames sorts the items by the names their encoded keys are of, which one of them has an escape
// for: the escape of a character is not in the order of the character.
func sortMapItemsByNames(items []MapItem) {
	type named struct {
		name string
		item MapItem
	}
	byName := make([]named, len(items))
	for i, item := range items {
		byName[i] = named{name: encodedKeyName(item.Key), item: item}
	}
	slices.SortStableFunc(byName, func(a, b named) int {
		return strings.Compare(a.name, b.name)
	})
	for i := range byName {
		items[i] = byName[i].item
	}
}

// encodedKeyName returns the name of an encoded key: the string from its first quote, whose escapes are the
// ones AppendString writes, decoded. What follows the string, and what precedes its quote, as the codes of a
// color, is not of the name.
func encodedKeyName(key []byte) string {
	start := bytes.IndexByte(key, '"')
	if start < 0 {
		return string(key)
	}
	var name []byte
	for i := start + 1; i < len(key); i++ {
		c := key[i]
		switch {
		case c == '"':
			return string(name)
		case c != '\\' || i+1 == len(key):
			name = append(name, c)
		case key[i+1] == 'u' && i+5 < len(key):
			r, err := strconv.ParseUint(string(key[i+2:i+6]), 16, 16)
			if err != nil {
				name = append(name, c)
				continue
			}
			name = utf8.AppendRune(name, rune(r))
			i += 5
		default:
			i++
			switch e := key[i]; e {
			case 'n':
				name = append(name, '\n')
			case 'r':
				name = append(name, '\r')
			case 't':
				name = append(name, '\t')
			case 'b':
				name = append(name, '\b')
			case 'f':
				name = append(name, '\f')
			default:
				name = append(name, e)
			}
		}
	}
	return string(name)
}

// maxItemsOfInsertionSort is the number of the items up to which the insertion sort is used.
// With keys which are random in their content and in their length, it is faster than slices.SortFunc up to
// 24 items and slower from 32 items ( BenchmarkVariant_MapSort ).
const maxItemsOfInsertionSort = 16

// MapContext is what the VM encodes a map from: the entries read from the map ( MapLayout.Collect ), and
// the state of the entry being encoded.
type MapContext struct {
	layout *MapLayout
	// Keys are the keys of the entries, for keys of a string kind; RawKeys are the bytes of the keys of another
	// kind, keySize each; Values are the bytes of the values, valueSize each. The copies of the values may hold
	// pointers which the GC doesn't see, as the slots of the VM do: the map holds the values while they are
	// encoded.
	Keys    []string
	RawKeys []byte
	Values  []byte
	// Order is the entries sorted by their keys ( SortKeys ), for keys of a string kind; Sorted is whether
	// the entries are encoded in that order.
	Order    []int32
	prefixes []uint64 // the first bytes of the keys as numbers, while they are sorted
	Sorted   bool
	// DirectEntries is whether the entries of a sorted map are written directly by one call of the VM, up to one
	// whose value is left to its opcodes, as the entries of a map which is not sorted are written as it is read:
	// the values are written by one opcode of a scalar, or are of interface{} ( see MapLayout ).
	DirectEntries bool
	Len           int
	Idx           int
	// The entries of a sorted map whose keys are not of a string kind are encoded as they come and put in
	// the order of their encoded keys after: Start is where the key or the value being written starts,
	// First is where the entries start in the buffer, Slice has the entries and Buf is where they are copied.
	Start int
	First int
	Slice *Mapslice
	items []MapItem // what Slice.Items is made of, kept between the maps
	Buf   []byte
	// what the collector by reflect.MapIter works with: here so that nothing is allocated for a map.
	iter       reflect.MapIter
	keyIface   any
	valueIface any
	// names are the names of the keys written, of a map whose keys may have the same name ( see
	// appendMapKeyName ), if namesKept is set: they are cleared for each map which needs them.
	names     map[string]struct{}
	namesKept bool
	// recorded is whether the map is recorded for the detection of cycles ( see RuntimeContext.RecordMap ).
	recorded bool
}

// NewMapContext returns the context to encode a map: the runtime context has one for each level of the maps
// nested in each other, kept from a call to the next, and refers to it while the VM has it in a slot, which
// the GC doesn't see.
func NewMapContext(rctx *RuntimeContext) *MapContext {
	if rctx.mapDepth == len(rctx.mapContexts) {
		rctx.mapContexts = append(rctx.mapContexts, &MapContext{Slice: &Mapslice{}})
	}
	ctx := rctx.mapContexts[rctx.mapDepth]
	rctx.mapDepth++
	ctx.Buf = ctx.Buf[:0]
	ctx.Idx = 0
	ctx.Sorted = false
	ctx.DirectEntries = false
	// Items is set by SortByEncodedKeys, and tells the VM the entries are put in that order.
	ctx.Slice.Items = nil
	ctx.namesKept, ctx.recorded = false, false
	return ctx
}

// RecordMap records the map at p of code, which the VM encodes past StartDetectingMapCyclesAfter, in m, for the
// detection of cycles, as encoding/json records a map: a cycle through the values of a map, which are copies, has
// the map alone as the value which it reaches again. The maps are recorded apart from the values of the frames,
// which a map is not compared with.
func (c *RuntimeContext) RecordMap(code *Opcode, p unsafe.Pointer, m *MapContext) error {
	if slices.Contains(c.seenMaps, p) {
		m := p // the map, whose address is taken here only, which keeps p from escaping
		return errCycle(c, runtime.TypeOfPtr(code.Type), unsafe.Pointer(&m))
	}
	c.seenMaps = append(c.seenMaps, p)
	m.recorded = true
	return nil
}

// ScalarValue is whether the values of the map are written by one opcode of a scalar ( MapLayout.ScalarValue ).
func (c *MapContext) ScalarValue() bool {
	return c.layout.ScalarValue
}

// SortByEncodedKeys makes the context put the entries in the order of their encoded keys, for a sorted map
// whose keys are not of a string kind: the VM records the entries in Slice as it encodes them.
func (c *MapContext) SortByEncodedKeys() {
	if cap(c.items) < c.Len {
		c.items = make([]MapItem, c.Len)
	}
	c.Slice.Items = c.items[:c.Len]
	c.Slice.escaped = false
}

// appendText appends the text of a marshaler or of the key of a map as a string. A text which is escaped is
// written longer than itself and its quotes: that tells the map being encoded, if any, that its keys are to be
// sorted by their names ( see Mapslice.Sort ).
func appendText(ctx *RuntimeContext, b []byte, text string) []byte {
	n := len(b)
	b, _ = jsonstring.AppendQuoted(StringEscaper(ctx), b, text)
	if len(b)-n != len(text)+2 {
		ctx.textEscaped()
	}
	return b
}

// textEscaped tells the map being encoded, if any, that a text was escaped ( see appendText ).
func (c *RuntimeContext) textEscaped() {
	if c.mapDepth == 0 {
		return
	}
	if m := c.mapContexts[c.mapDepth-1]; m.layout != nil && m.layout.KeysMayEscape {
		m.Slice.escaped = true
	}
}

func ReleaseMapContext(rctx *RuntimeContext, c *MapContext) {
	// a map is always released before the maps it is in.
	rctx.mapDepth--
	if c.recorded {
		// the last record: the ones of the maps of its values were dropped before
		rctx.seenMaps = rctx.seenMaps[:len(rctx.seenMaps)-1]
	}
	// the keys refer to the map, which the context must not keep alive.
	clear(c.Keys)
	c.Keys = c.Keys[:0]
}

func AppendByteSlice(ctx *RuntimeContext, b []byte, src []byte) []byte {
	if src == nil {
		if NilSliceIsEmpty(ctx) {
			return append(b, `""`...)
		}
		return append(b, `null`...)
	}
	encodedLen := base64.StdEncoding.EncodedLen(len(src))
	b = append(b, '"')
	pos := len(b)
	remainLen := cap(b[pos:])
	var buf []byte
	if remainLen > encodedLen {
		buf = b[pos : pos+encodedLen]
	} else {
		buf = make([]byte, encodedLen)
	}
	base64.StdEncoding.Encode(buf, src)
	return append(append(b, buf...), '"')
}

// AppendFloat32 appends v as encoding/json writes a float32: by the shortest decimal which rounds to it, in the
// format 'f', or 'e' below 1e-6 and from 1e21, whose exponent has no leading zero.
func AppendFloat32(_ *RuntimeContext, b []byte, v float32) []byte {
	return appendFloatOfBits(b, float64(v), 32)
}

// AppendFloat64 appends v as encoding/json writes a float64, as AppendFloat32 does.
func AppendFloat64(_ *RuntimeContext, b []byte, v float64) []byte {
	return appendFloatOfBits(b, v, 64)
}

// appendFloatOfBits appends the float of the precision bits, 64 or 32. NaN and the infinities, which the callers
// report as errors before, but a map key, are appended as strconv appends them.
func appendFloatOfBits(b []byte, f float64, bits int) []byte {
	const exponent = 0x7ff << 52
	if math.Float64bits(f)&exponent == exponent {
		return strconv.AppendFloat(b, f, 'g', -1, bits)
	}
	return floatfmt.AppendFloat(b, f, bits)
}

func AppendBool(_ *RuntimeContext, b []byte, v bool) []byte {
	if v {
		return append(b, "true"...)
	}
	return append(b, "false"...)
}

// AppendNumber appends n, which must be a JSON number by its grammar, as encoding/json writes a json.Number: an
// empty one as 0.
//
// A number of up to 24 bytes is copied by the words which overlap, without a call of memmove, and the bytes which
// are not digits are found in the words as they are loaded, which the grammar is then checked by, without a call:
// an integer at once, and another number by the places of its bytes which are not digits.
func AppendNumber(_ *RuntimeContext, b []byte, n json.Number) ([]byte, error) {
	l := len(n)
	if l == 0 {
		return append(b, '0'), nil
	}
	if l > 24 {
		if !jsonnum.IsValid(unsafe.Slice(unsafe.StringData(string(n)), l)) {
			return nil, invalidNumberError(n, false)
		}
		return append(b, n...), nil
	}
	if cap(b)-len(b) < l {
		b = growForNumber(b, l)
	}
	src := unsafe.Slice(unsafe.StringData(string(n)), l)
	dst := b[len(b) : len(b)+l]
	// nonDigits has the bit i of each byte i of n which is not a digit
	var nonDigits uint
	switch {
	case l > 16:
		first, mid, last := binary.LittleEndian.Uint64(src), binary.LittleEndian.Uint64(src[8:]), binary.LittleEndian.Uint64(src[l-8:])
		binary.LittleEndian.PutUint64(dst, first)
		binary.LittleEndian.PutUint64(dst[8:], mid)
		binary.LittleEndian.PutUint64(dst[l-8:], last)
		tf, tm, tl := jsonnum.NonDigitTops(first), jsonnum.NonDigitTops(mid), jsonnum.NonDigitTops(last)
		if tf|tm|tl == 0 && src[0] != '0' {
			return b[:len(b)+l], nil
		}
		nonDigits = uint(jsonnum.GatherTops(tf)) | uint(jsonnum.GatherTops(tm))<<8 | uint(jsonnum.GatherTops(tl))<<(l-8)
	case l >= 8:
		first, last := binary.LittleEndian.Uint64(src), binary.LittleEndian.Uint64(src[l-8:])
		binary.LittleEndian.PutUint64(dst, first)
		binary.LittleEndian.PutUint64(dst[l-8:], last)
		tf, tl := jsonnum.NonDigitTops(first), jsonnum.NonDigitTops(last)
		if tf|tl == 0 && src[0] != '0' {
			return b[:len(b)+l], nil
		}
		nonDigits = uint(jsonnum.GatherTops(tf)) | uint(jsonnum.GatherTops(tl))<<(l-8)
	case l >= 4:
		first, last := binary.LittleEndian.Uint32(src), binary.LittleEndian.Uint32(src[l-4:])
		binary.LittleEndian.PutUint32(dst, first)
		binary.LittleEndian.PutUint32(dst[l-4:], last)
		t := jsonnum.NonDigitTops(uint64(first) | uint64(last)<<32)
		if t == 0 && src[0] != '0' {
			return b[:len(b)+l], nil
		}
		m := uint(jsonnum.GatherTops(t))
		nonDigits = m&0xf | m>>4<<(l-4)
	default:
		for i := 0; i < l; i++ {
			c := src[i]
			dst[i] = c
			if c-'0' > 9 {
				nonDigits |= 1 << i
			}
		}
	}
	// a number without a sign or an exponent, the most common one, first
	if nonDigits&(nonDigits-1) == 0 {
		if nonDigits == 0 {
			if src[0] != '0' || l == 1 {
				return b[:len(b)+l], nil
			}
		} else if dot := bits.TrailingZeros(nonDigits); src[dot] == '.' && dot > 0 && dot < l-1 && (src[0] != '0' || dot == 1) {
			return b[:len(b)+l], nil
		}
	}
	// the others by the grammar in place: the bytes which are not digits are a sign, a decimal point, the exponent
	// and its sign, each at most once and in its place, whose bits are cleared one by one.
	start := 0
	if src[0] == '-' {
		nonDigits &^= 1
		start = 1
	}
	valid := false
	if start < l && nonDigits&(1<<start) == 0 && (src[start] != '0' || start+1 == l || nonDigits&(1<<(start+1)) != 0) {
		valid = nonDigits == 0
		p, fractionDigits := bits.TrailingZeros(nonDigits), true // digits after the decimal point, if there is one
		if !valid && src[p] == '.' {
			fractionDigits = p+1 < l && nonDigits&(1<<(p+1)) == 0
			nonDigits &^= 1 << p
			valid = fractionDigits && nonDigits == 0
			p = bits.TrailingZeros(nonDigits)
		}
		if !valid && fractionDigits && nonDigits != 0 && src[p]|0x20 == 'e' {
			nonDigits &^= 1 << p
			if p+1 < l && (src[p+1] == '+' || src[p+1] == '-') {
				nonDigits &^= 1 << (p + 1)
				p++
			}
			valid = p+1 < l && nonDigits == 0
		}
	}
	if !valid {
		return nil, invalidNumberError(n, false)
	}
	return b[:len(b)+l], nil
}

// AppendNumberString is AppendNumber of a json.Number of a field with the option string, which the caller
// quotes: its error shows the number quoted, as encoding/json shows it.
func AppendNumberString(ctx *RuntimeContext, b []byte, n json.Number) ([]byte, error) {
	out, err := AppendNumber(ctx, b, n)
	if err != nil {
		return nil, invalidNumberError(n, true)
	}
	return out, nil
}

// growForNumber returns b with room for n bytes more, growing as append does.
//
//go:noinline
func growForNumber(b []byte, n int) []byte {
	grown := make([]byte, len(b), 2*cap(b)+n)
	copy(grown, b)
	return grown
}

// addrForMarshaler returns the pointer to the value held by v, to call a marshaler with a pointer receiver.
//
// The VM makes v from the address of the value ( interfaceOf ), so the data word of v is that address for every
// type, also for a type which is stored directly in an interface value, whose data word is then not the value:
// the marshaler is called with the original value, as encoding/json calls it for an addressable value, and
// nothing is allocated.
func addrForMarshaler(v any, rv reflect.Value) reflect.Value {
	return reflect.NewAt(rv.Type(), (*emptyInterface)(unsafe.Pointer(&v)).ptr)
}

// errStoppedByInvalidUTF8 stops an encoding which wrote a string of invalid UTF-8, whose error is the one of the
// string ( see RuntimeContext.InvalidUTF8Output ).
var errStoppedByInvalidUTF8 = stderrors.New("stopped by invalid UTF-8")

// appendValue appends the value at p by the function of m. An encoding which wrote a string of invalid UTF-8 goes
// on to its end, or to the next function, which is not called: a method or a function of marshaling is not called
// after the error, as it isn't by encoding/json/v2.
func appendValue(ctx *RuntimeContext, m *MarshalerCall, b []byte, p unsafe.Pointer) ([]byte, error) {
	if ctx.Option.Flag&RejectInvalidUTF8Option != 0 && ctx.HasInvalidUTF8() {
		return b, errStoppedByInvalidUTF8
	}
	return m.appendValue(ctx, b, p)
}

// appendV2MarshalJSON appends the output of MarshalJSON of the v2 semantics of the value at p, which it calls by
// its code, without the call of the function value of m ( see MarshalerCall.v2Raw ), as appendValue does after
// m.appendOutput, which AppendMarshalJSON looked at.
func appendV2MarshalJSON(ctx *RuntimeContext, m *MarshalerCall, b []byte, p unsafe.Pointer) ([]byte, error) {
	if ctx.Option.Flag&RejectInvalidUTF8Option != 0 && ctx.HasInvalidUTF8() {
		return b, errStoppedByInvalidUTF8
	}
	raw, err := m.call(p)
	return appendV2Raw(ctx, m.recv, b, raw, err)
}

// AppendMarshalJSON appends what MarshalJSON of the value returns, compacted. p is the data word of the
// interface value of the type of the opcode: the address of the value, or the pointer for a pointer type.
func AppendMarshalJSON(ctx *RuntimeContext, code *Opcode, b []byte, p unsafe.Pointer) ([]byte, error) {
	m := code.Marshaler
	if m == nil {
		return appendMarshalJSONByInterface(ctx, code, b, interfaceOf(code, p))
	}
	// a function of the v2 semantics has neither nilIsNull nor, but for MarshalJSON, appendOutput.
	if m.nilIsNull && p == nil {
		return AppendNull(ctx, b), nil
	}
	if m.appendOutput != nil {
		// the output of a type of the standard library is valid and compact: it is written as it is.
		if out, ok := m.appendOutput(b, p); ok {
			return out, nil
		}
	}
	var bb []byte
	var err error
	if code.Flags&(MarshalerContextFlags|MarshalerFuncFlags) != 0 {
		// the calls of a context and of the v2 semantics, which the others don't look at.
		if code.Flags&MarshalerFuncFlags != 0 {
			if m.v2Raw {
				return appendV2MarshalJSON(ctx, m, b, p)
			}
			return appendValue(ctx, m, b, p)
		}
		stdctx := ctx.marshalerContext()
		if ctx.Option.Flag&FieldQueryOption != 0 {
			stdctx = SetFieldQueryToContext(stdctx, code.FieldQuery)
		}
		bb, err = m.callContext(p, stdctx)
	} else {
		bb, err = m.call(p)
	}
	if err != nil {
		return nil, &errors.MarshalerError{Type: m.recv, Err: err}
	}
	escape := (ctx.Option.Flag & HTMLEscapeOption) != 0
	if out, ok := appendCompactOutput(b, bb, escape, m.trusted); ok {
		// the output is compact and valid: it is copied as it is.
		return out, nil
	}
	out, err := appendCompacted(ctx, b, bb, escape)
	if err != nil {
		return nil, &errors.MarshalerError{Type: m.recv, Err: err}
	}
	return out, nil
}

// interfaceOf returns the interface value of the type of the opcode whose data word is p.
func interfaceOf(code *Opcode, p unsafe.Pointer) any {
	return *(*any)(unsafe.Pointer(&emptyInterface{typ: code.Type, ptr: p}))
}

func appendMarshalJSONByInterface(ctx *RuntimeContext, code *Opcode, b []byte, v any) ([]byte, error) {
	rv := reflect.ValueOf(v) // convert by dynamic interface type
	if (code.Flags & AddrForMarshalerFlags) != 0 {
		rv = addrForMarshaler(v, rv)
	}

	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return AppendNull(ctx, b), nil
	}

	v = rv.Interface()
	var bb []byte
	if (code.Flags & MarshalerContextFlags) != 0 {
		marshaler, ok := v.(marshalerContext)
		if !ok {
			return AppendNull(ctx, b), nil
		}
		stdctx := ctx.marshalerContext()
		if ctx.Option.Flag&FieldQueryOption != 0 {
			stdctx = SetFieldQueryToContext(stdctx, code.FieldQuery)
		}
		b, err := marshaler.MarshalJSON(stdctx)
		if err != nil {
			return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
		}
		bb = b
	} else {
		marshaler, ok := v.(json.Marshaler)
		if !ok {
			return AppendNull(ctx, b), nil
		}
		b, err := marshaler.MarshalJSON()
		if err != nil {
			return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
		}
		bb = b
	}
	compactedBuf, err := appendCompact(ctx, b, bb, (ctx.Option.Flag&HTMLEscapeOption) != 0)
	if err != nil {
		return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
	}
	return compactedBuf, nil
}

func AppendMarshalJSONIndent(ctx *RuntimeContext, code *Opcode, b []byte, p unsafe.Pointer) ([]byte, error) {
	m := code.Marshaler
	if m == nil {
		return appendMarshalJSONIndentByInterface(ctx, code, b, interfaceOf(code, p))
	}
	if m.nilIsNull && p == nil {
		return AppendNull(ctx, b), nil
	}
	if m.appendOutput != nil {
		// the output of a type of the standard library is one token, which has nothing to indent.
		if out, ok := m.appendOutput(b, p); ok {
			return out, nil
		}
	}
	var bb []byte
	var err error
	if code.Flags&MarshalerFuncFlags != 0 {
		// a function of the v2 semantics, whose output is formatted after the encoding.
		return appendValue(ctx, m, b, p)
	}
	if (code.Flags & MarshalerContextFlags) != 0 {
		bb, err = m.callContext(p, ctx.marshalerContext())
	} else {
		bb, err = m.call(p)
	}
	if err != nil {
		return nil, &errors.MarshalerError{Type: m.recv, Err: err}
	}
	return appendIndentedMarshalJSON(ctx, code, b, bb)
}

func appendMarshalJSONIndentByInterface(ctx *RuntimeContext, code *Opcode, b []byte, v any) ([]byte, error) {
	rv := reflect.ValueOf(v) // convert by dynamic interface type
	if (code.Flags & AddrForMarshalerFlags) != 0 {
		rv = addrForMarshaler(v, rv)
	}
	// a nil pointer is null, as encoding/json does. The VM gives the address of the pointer,
	// so it is not known to be nil until here.
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return AppendNull(ctx, b), nil
	}
	v = rv.Interface()
	var bb []byte
	if (code.Flags & MarshalerContextFlags) != 0 {
		marshaler, ok := v.(marshalerContext)
		if !ok {
			return AppendNull(ctx, b), nil
		}
		b, err := marshaler.MarshalJSON(ctx.marshalerContext())
		if err != nil {
			return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
		}
		bb = b
	} else {
		marshaler, ok := v.(json.Marshaler)
		if !ok {
			return AppendNull(ctx, b), nil
		}
		b, err := marshaler.MarshalJSON()
		if err != nil {
			return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
		}
		bb = b
	}
	return appendIndentedMarshalJSON(ctx, code, b, bb)
}

func appendIndentedMarshalJSON(ctx *RuntimeContext, code *Opcode, b []byte, bb []byte) ([]byte, error) {
	indentedBuf, err := appendIndent(
		ctx,
		b,
		bb,
		string(ctx.Prefix)+strings.Repeat(string(ctx.IndentStr), int(ctx.BaseIndent+code.Indent)),
		string(ctx.IndentStr),
		(ctx.Option.Flag&HTMLEscapeOption) != 0,
		false,
	)
	if err != nil {
		return nil, &errors.MarshalerError{Type: runtime.TypeOfPtr(code.Type), Err: err}
	}
	return indentedBuf, nil
}

// AppendMarshalText appends what MarshalText of the value returns, as a string. p is the data word of the
// interface value of the type of the opcode: the address of the value, or the pointer for a pointer type.
func AppendMarshalText(ctx *RuntimeContext, code *Opcode, b []byte, p unsafe.Pointer) ([]byte, error) {
	m := code.Marshaler
	if m == nil {
		if code.Flags&InterfaceMapKeyFlags != 0 {
			return appendInterfaceMapKey(ctx, code, b, p)
		}
		return appendMarshalTextByInterface(ctx, code, b, interfaceOf(code, p))
	}
	if m.nilIsNull && p == nil {
		return appendNilText(ctx, code, b), nil
	}
	var bytes []byte
	appended := false
	if m.appendOutput != nil {
		// the text of a type of the standard library is appended to a buffer of the context, not allocated.
		bytes, appended = m.appendOutput(ctx.MarshalBuf[:0], p)
		if appended {
			ctx.MarshalBuf = bytes
		}
	}
	if !appended {
		var err error
		if bytes, err = m.call(p); err != nil {
			return nil, &errors.MarshalerError{Type: m.recv, Err: err}
		}
	}
	// appendText, written here: it is not inlined, and this is the text of the key of most maps of texts.
	n := len(b)
	b, _ = jsonstring.AppendQuoted(StringEscaper(ctx), b, *(*string)(unsafe.Pointer(&bytes)))
	if len(b)-n != len(bytes)+2 {
		ctx.textEscaped()
	}
	return b, nil
}

func appendMarshalTextByInterface(ctx *RuntimeContext, code *Opcode, b []byte, v any) ([]byte, error) {
	rv := reflect.ValueOf(v) // convert by dynamic interface type
	if (code.Flags & AddrForMarshalerFlags) != 0 {
		rv = addrForMarshaler(v, rv)
	}
	// a nil pointer is null, or "" as the name of a key, as encoding/json does. The VM gives the address of
	// the pointer, so it is not known to be nil until here.
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return appendNilText(ctx, code, b), nil
	}
	v = rv.Interface()
	marshaler, ok := v.(encoding.TextMarshaler)
	if !ok {
		return AppendNull(ctx, b), nil
	}
	bytes, err := marshaler.MarshalText()
	if err != nil {
		return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
	}
	return appendText(ctx, b, *(*string)(unsafe.Pointer(&bytes))), nil
}

// appendNilText appends the text of a nil pointer: null, or "" as the name of a key of a map, as
// encoding/json writes them.
func appendNilText(ctx *RuntimeContext, code *Opcode, b []byte) []byte {
	if code.Flags&MapKeyFlags != 0 {
		return append(b, `""`...)
	}
	return AppendNull(ctx, b)
}

// AppendMarshalTextIndent is AppendMarshalText: the text has no indent.
func AppendMarshalTextIndent(ctx *RuntimeContext, code *Opcode, b []byte, p unsafe.Pointer) ([]byte, error) {
	return AppendMarshalText(ctx, code, b, p)
}

func AppendNull(_ *RuntimeContext, b []byte) []byte {
	return append(b, "null"...)
}

func AppendComma(_ *RuntimeContext, b []byte) []byte {
	return append(b, ',')
}

func AppendCommaIndent(_ *RuntimeContext, b []byte) []byte {
	return append(b, ',', '\n')
}

func AppendStructEnd(_ *RuntimeContext, b []byte) []byte {
	return append(b, '}', ',')
}

func AppendStructEndIndent(ctx *RuntimeContext, code *Opcode, b []byte) []byte {
	b = append(b, '\n')
	b = append(b, ctx.Prefix...)
	indentNum := ctx.BaseIndent + code.Indent - 1
	for i := uint32(0); i < indentNum; i++ {
		b = append(b, ctx.IndentStr...)
	}
	return append(b, '}', ',', '\n')
}

func AppendIndent(ctx *RuntimeContext, b []byte, indent uint32) []byte {
	b = append(b, ctx.Prefix...)
	indentNum := ctx.BaseIndent + indent
	for i := uint32(0); i < indentNum; i++ {
		b = append(b, ctx.IndentStr...)
	}
	return b
}
