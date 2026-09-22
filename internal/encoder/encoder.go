package encoder

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
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
	DataWordIsAddr           bool
	NoescapeKeyCode          *Opcode
	EscapeKeyCode            *Opcode
	InterfaceNoescapeKeyCode *Opcode
	InterfaceEscapeKeyCode   *Opcode
	CodeLength               int
	EndCode                  *Opcode
	Code                     Code
	QueryCache               map[string]*OpcodeSet
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
		codeSet, err := compileToGetUnfilteredCodeSet(uintptr(typ))
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
	codeSet, err := compileToGetUnfilteredCodeSet(uintptr(typ))
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

const StartDetectingCyclesAfter = 1000

func ErrUnsupportedValue(code *Opcode, ptr unsafe.Pointer) *errors.UnsupportedValueError {
	v := *(*interface{})(unsafe.Pointer(&emptyInterface{
		typ: code.Type,
		ptr: ptr,
	}))
	return &errors.UnsupportedValueError{
		Value: reflect.ValueOf(v),
		Str:   fmt.Sprintf("encountered a cycle via %s", runtime.TypeOfPtr(code.Type)),
	}
}

func ErrUnsupportedFloat(v float64) *errors.UnsupportedValueError {
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
	// RawKey is the key as it is, if the key is a plain string. The items are sorted by it, as encoding/json
	// sorts the keys: the order of the encoded keys differs from it when a key has a character to escape.
	RawKey string
}

type Mapslice struct {
	Items []MapItem
}

// Sort sorts the items by their keys.
//
// It is not sort.Sort, which calls Less and Swap through an interface for every comparison:
// the maps to encode are small in most cases, and the sort was a tenth of the time to encode one.
func (m *Mapslice) Sort(byRawKey bool) {
	items := m.Items
	if byRawKey {
		if len(items) > maxItemsOfInsertionSort {
			slices.SortFunc(items, func(a, b MapItem) int {
				return strings.Compare(a.RawKey, b.RawKey)
			})
			return
		}
		insertionSortMapItemsByRawKey(items)
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

func insertionSortMapItemsByRawKey(items []MapItem) {
	for i := 1; i < len(items); i++ {
		if items[i-1].RawKey <= items[i].RawKey {
			continue
		}
		item := items[i]
		j := i
		for ; j > 0 && items[j-1].RawKey > item.RawKey; j-- {
			items[j] = items[j-1]
		}
		items[j] = item
	}
}

// maxItemsOfInsertionSort is the number of the items up to which the insertion sort is used.
// With keys which are random in their content and in their length, it is faster than slices.SortFunc up to
// 24 items and slower from 32 items ( BenchmarkVariant_MapSort ).
const maxItemsOfInsertionSort = 16

//nolint:unused
type mapIter struct {
	key         unsafe.Pointer
	elem        unsafe.Pointer
	t           unsafe.Pointer
	h           unsafe.Pointer
	buckets     unsafe.Pointer
	bptr        unsafe.Pointer
	overflow    unsafe.Pointer
	oldoverflow unsafe.Pointer
	startBucket uintptr
	offset      uint8
	wrapped     bool
	B           uint8
	i           uint8
	bucket      uintptr
	checkBucket uintptr
}

type MapContext struct {
	// parent is the context of the map which this map is in.
	parent *MapContext
	Start  int
	First  int
	Idx    int
	Slice  *Mapslice
	Buf    []byte
	Len    int
	Iter   mapIter
}

var mapContextPool = sync.Pool{
	New: func() interface{} {
		return &MapContext{
			Slice: &Mapslice{},
		}
	},
}

// NewMapContext returns the context to encode a map, which the runtime context refers to until it is released:
// the VM has it only in a slot, which the GC doesn't see.
func NewMapContext(rctx *RuntimeContext, mapLen int, unorderedMap bool) *MapContext {
	ctx := mapContextPool.Get().(*MapContext)
	ctx.parent = rctx.mapContext
	rctx.mapContext = ctx
	if !unorderedMap {
		if len(ctx.Slice.Items) < mapLen {
			ctx.Slice.Items = make([]MapItem, mapLen)
		} else {
			ctx.Slice.Items = ctx.Slice.Items[:mapLen]
		}
	}
	ctx.Buf = ctx.Buf[:0]
	ctx.Idx = 0
	ctx.Len = mapLen
	return ctx
}

func ReleaseMapContext(rctx *RuntimeContext, c *MapContext) {
	// a map is always released before the maps it is in.
	rctx.mapContext = c.parent
	c.parent = nil
	// the iterator refers to the map, which the pool must not keep alive.
	c.Iter = mapIter{}
	mapContextPool.Put(c)
}

//go:linkname MapIterInit runtime.mapiterinit
//go:noescape
func MapIterInit(mapType unsafe.Pointer, m unsafe.Pointer, it *mapIter)

//go:linkname MapIterKey reflect.mapiterkey
//go:noescape
func MapIterKey(it *mapIter) unsafe.Pointer

//go:linkname MapIterNext reflect.mapiternext
//go:noescape
func MapIterNext(it *mapIter)

//go:linkname MapLen reflect.maplen
//go:noescape
func MapLen(m unsafe.Pointer) int

func AppendByteSlice(_ *RuntimeContext, b []byte, src []byte) []byte {
	if src == nil {
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

func AppendFloat32(_ *RuntimeContext, b []byte, v float32) []byte {
	f64 := float64(v)
	abs := math.Abs(f64)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		f32 := float32(abs)
		if f32 < 1e-6 || f32 >= 1e21 {
			fmt = 'e'
		}
	}
	return strconv.AppendFloat(b, f64, fmt, -1, 32)
}

func AppendFloat64(_ *RuntimeContext, b []byte, v float64) []byte {
	abs := math.Abs(v)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if abs < 1e-6 || abs >= 1e21 {
			fmt = 'e'
		}
	}
	return strconv.AppendFloat(b, v, fmt, -1, 64)
}

func AppendBool(_ *RuntimeContext, b []byte, v bool) []byte {
	if v {
		return append(b, "true"...)
	}
	return append(b, "false"...)
}

var (
	floatTable = [256]bool{
		'0': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'9': true,
		'.': true,
		'e': true,
		'E': true,
		'+': true,
		'-': true,
	}
)

func AppendNumber(_ *RuntimeContext, b []byte, n json.Number) ([]byte, error) {
	if len(n) == 0 {
		return append(b, '0'), nil
	}
	for i := 0; i < len(n); i++ {
		if !floatTable[n[i]] {
			return nil, fmt.Errorf("json: invalid number literal %q", n)
		}
	}
	b = append(b, n...)
	return b, nil
}

// addrForMarshaler returns the pointer to the value held by v, to call a marshaler with a pointer receiver.
//
// The VM makes v from the address of the value, so the data word of v is that address unless the type is
// stored directly in an interface value. Only a pointer-sized type can be stored directly, so for the other
// sizes the pointer is made from the data word: the marshaler is called with the original value, as
// encoding/json does, and nothing is allocated. A pointer-sized value is copied, because its data word
// may be the value itself.
func addrForMarshaler(v interface{}, rv reflect.Value) reflect.Value {
	if rv.CanAddr() {
		return rv.Addr()
	}
	typ := rv.Type()
	if typ.Size() != unsafe.Sizeof(unsafe.Pointer(nil)) {
		return reflect.NewAt(typ, (*emptyInterface)(unsafe.Pointer(&v)).ptr)
	}
	newV := reflect.New(typ)
	newV.Elem().Set(rv)
	return newV
}

func AppendMarshalJSON(ctx *RuntimeContext, code *Opcode, b []byte, v interface{}) ([]byte, error) {
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
	marshalBuf := ctx.MarshalBuf[:0]
	marshalBuf = append(append(marshalBuf, bb...), nul)
	compactedBuf, err := compact(b, marshalBuf, (ctx.Option.Flag&HTMLEscapeOption) != 0)
	if err != nil {
		return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
	}
	ctx.MarshalBuf = marshalBuf
	return compactedBuf, nil
}

func AppendMarshalJSONIndent(ctx *RuntimeContext, code *Opcode, b []byte, v interface{}) ([]byte, error) {
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
	marshalBuf := ctx.MarshalBuf[:0]
	marshalBuf = append(append(marshalBuf, bb...), nul)
	indentedBuf, err := doIndent(
		b,
		marshalBuf,
		string(ctx.Prefix)+strings.Repeat(string(ctx.IndentStr), int(ctx.BaseIndent+code.Indent)),
		string(ctx.IndentStr),
		(ctx.Option.Flag&HTMLEscapeOption) != 0,
	)
	if err != nil {
		return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
	}
	ctx.MarshalBuf = marshalBuf
	return indentedBuf, nil
}

func AppendMarshalText(ctx *RuntimeContext, code *Opcode, b []byte, v interface{}) ([]byte, error) {
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
	marshaler, ok := v.(encoding.TextMarshaler)
	if !ok {
		return AppendNull(ctx, b), nil
	}
	bytes, err := marshaler.MarshalText()
	if err != nil {
		return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
	}
	return AppendString(ctx, b, *(*string)(unsafe.Pointer(&bytes))), nil
}

func AppendMarshalTextIndent(ctx *RuntimeContext, code *Opcode, b []byte, v interface{}) ([]byte, error) {
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
	marshaler, ok := v.(encoding.TextMarshaler)
	if !ok {
		return AppendNull(ctx, b), nil
	}
	bytes, err := marshaler.MarshalText()
	if err != nil {
		return nil, &errors.MarshalerError{Type: reflect.TypeOf(v), Err: err}
	}
	return AppendString(ctx, b, *(*string)(unsafe.Pointer(&bytes))), nil
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

func IsNilForMarshaler(v interface{}) bool {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(rv.Float()) == 0
	case reflect.Interface, reflect.Ptr, reflect.Func:
		return rv.IsNil()
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		// the same as encoding/json: they are empty if the length is zero, even if they are not nil.
		return rv.Len() == 0
	}
	return false
}
