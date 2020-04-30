package json

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/xerrors"
)

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	w    io.Writer
	buf  []byte
	pool sync.Pool
}

type EncodeOp func(*Encoder, *rtype, uintptr) error

const (
	bufSize = 1024
)

type EncodeOpMap struct {
	sync.Map
}

func (m *EncodeOpMap) Get(k string) *opcode { //EncodeOp {
	if v, ok := m.Load(k); ok {
		return v.(*opcode) //(EncodeOp)
	}
	return nil
}

func (m *EncodeOpMap) Set(k string, op *opcode) { // EncodeOp) {
	m.Store(k, op)
}

var (
	encPool            sync.Pool
	cachedEncodeOp     EncodeOpMap
	errCompileSlowPath = xerrors.New("json: detect dynamic type ( interface{} ) and compile with slow path")
)

func init() {
	encPool = sync.Pool{
		New: func() interface{} {
			return &Encoder{
				buf:  make([]byte, 0, bufSize),
				pool: encPool,
			}
		},
	}
	cachedEncodeOp = EncodeOpMap{}
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	enc := encPool.Get().(*Encoder)
	enc.w = w
	enc.reset()
	return enc
}

// Encode writes the JSON encoding of v to the stream, followed by a newline character.
//
// See the documentation for Marshal for details about the conversion of Go values to JSON.
func (e *Encoder) Encode(v interface{}) error {
	if err := e.encode(v); err != nil {
		return err
	}
	if _, err := e.w.Write(e.buf); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeForMarshal(v interface{}) ([]byte, error) {
	if err := e.encode(v); err != nil {
		return nil, err
	}
	copied := make([]byte, len(e.buf))
	copy(copied, e.buf)
	return copied, nil
}

// SetEscapeHTML specifies whether problematic HTML characters should be escaped inside JSON quoted strings.
// The default behavior is to escape &, <, and > to \u0026, \u003c, and \u003e to avoid certain safety problems that can arise when embedding JSON in HTML.
//
// In non-HTML settings where the escaping interferes with the readability of the output, SetEscapeHTML(false) disables this behavior.
func (e *Encoder) SetEscapeHTML(on bool) {

}

// SetIndent instructs the encoder to format each subsequent encoded value as if indented by the package-level function Indent(dst, src, prefix, indent).
// Calling SetIndent("", "") disables indentation.
func (e *Encoder) SetIndent(prefix, indent string) {

}

func (e *Encoder) release() {
	e.w = nil
	e.pool.Put(e)
}

func (e *Encoder) reset() {
	e.buf = e.buf[:0]
}

func (e *Encoder) encodeInt(v int) {
	e.encodeInt64(int64(v))
}

func (e *Encoder) encodeInt8(v int8) {
	e.encodeInt64(int64(v))
}

func (e *Encoder) encodeInt16(v int16) {
	e.encodeInt64(int64(v))
}

func (e *Encoder) encodeInt32(v int32) {
	e.encodeInt64(int64(v))
}

func (e *Encoder) encodeInt64(v int64) {
	e.buf = strconv.AppendInt(e.buf, v, 10)
}

func (e *Encoder) encodeUint(v uint) {
	e.encodeUint64(uint64(v))
}

func (e *Encoder) encodeUint8(v uint8) {
	e.encodeUint64(uint64(v))
}

func (e *Encoder) encodeUint16(v uint16) {
	e.encodeUint64(uint64(v))
}

func (e *Encoder) encodeUint32(v uint32) {
	e.encodeUint64(uint64(v))
}

func (e *Encoder) encodeUint64(v uint64) {
	e.buf = strconv.AppendUint(e.buf, v, 10)
}

func (e *Encoder) encodeFloat32(v float32) {
	e.buf = strconv.AppendFloat(e.buf, float64(v), 'f', -1, 32)
}

func (e *Encoder) encodeFloat64(v float64) {
	e.buf = strconv.AppendFloat(e.buf, v, 'f', -1, 64)
}

func (e *Encoder) encodeBool(v bool) {
	e.buf = strconv.AppendBool(e.buf, v)
}

func (e *Encoder) encodeBytes(b []byte) {
	e.buf = append(e.buf, b...)
}

func (e *Encoder) encodeString(s string) {
	b := *(*[]byte)(unsafe.Pointer(&s))
	e.buf = append(e.buf, b...)
}

func (e *Encoder) encodeByte(b byte) {
	e.buf = append(e.buf, b)
}

func (e *Encoder) encode(v interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	typ := header.typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	name := typ.String()
	if op := cachedEncodeOp.Get(name); op != nil {
		p := uintptr(header.ptr)
		op.ptr = p
		if err := e.run(op); err != nil {
			return err
		}
		//op(e, typ, p)
		return nil
	}
	op, err := e.compileOp(typ)
	if err != nil {
		if err == errCompileSlowPath {
			/*
				slowOp, err := e.compileSlowPath(typ)
				if err != nil {
					return err
				}
				op = slowOp
			*/
		} else {
			return err
		}
	}
	if name != "" {
		cachedEncodeOp.Set(name, op)
	}
	p := uintptr(header.ptr)
	op.ptr = p
	e.run(op)
	//op(e, typ, p)
	return nil
}

func (e *Encoder) compile(typ *rtype) (EncodeOp, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return e.compilePtr(typ)
	case reflect.Slice:
		return e.compileSlice(typ)
	case reflect.Struct:
		return e.compileStruct(typ)
	case reflect.Map:
		return e.compileMap(typ)
	case reflect.Array:
		return e.compileArray(typ)
	case reflect.Int:
		return e.compileInt()
	case reflect.Int8:
		return e.compileInt8()
	case reflect.Int16:
		return e.compileInt16()
	case reflect.Int32:
		return e.compileInt32()
	case reflect.Int64:
		return e.compileInt64()
	case reflect.Uint:
		return e.compileUint()
	case reflect.Uint8:
		return e.compileUint8()
	case reflect.Uint16:
		return e.compileUint16()
	case reflect.Uint32:
		return e.compileUint32()
	case reflect.Uint64:
		return e.compileUint64()
	case reflect.Uintptr:
		return e.compileUint()
	case reflect.Float32:
		return e.compileFloat32()
	case reflect.Float64:
		return e.compileFloat64()
	case reflect.String:
		return e.compileString()
	case reflect.Bool:
		return e.compileBool()
	case reflect.Interface:
		return nil, errCompileSlowPath
	}
	return nil, xerrors.Errorf("failed to encode type %s: %w", typ.String(), ErrUnsupportedType)
}

func (e *Encoder) compileSlowPath(typ *rtype) (EncodeOp, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return e.compilePtrSlowPath(typ)
	case reflect.Slice:
		return e.compileSliceSlowPath(typ)
	case reflect.Struct:
		return e.compileStructSlowPath(typ)
	case reflect.Map:
		return e.compileMapSlowPath(typ)
	case reflect.Array:
		return e.compileArraySlowPath(typ)
	case reflect.Int:
		return e.compileInt()
	case reflect.Int8:
		return e.compileInt8()
	case reflect.Int16:
		return e.compileInt16()
	case reflect.Int32:
		return e.compileInt32()
	case reflect.Int64:
		return e.compileInt64()
	case reflect.Uint:
		return e.compileUint()
	case reflect.Uint8:
		return e.compileUint8()
	case reflect.Uint16:
		return e.compileUint16()
	case reflect.Uint32:
		return e.compileUint32()
	case reflect.Uint64:
		return e.compileUint64()
	case reflect.Uintptr:
		return e.compileUint()
	case reflect.Float32:
		return e.compileFloat32()
	case reflect.Float64:
		return e.compileFloat64()
	case reflect.String:
		return e.compileString()
	case reflect.Bool:
		return e.compileBool()
	case reflect.Interface:
		return e.compileInterface()
	}
	return nil, xerrors.Errorf("failed to encode type %s: %w", typ.String(), ErrUnsupportedType)
}

func (e *Encoder) compilePtr(typ *rtype) (EncodeOp, error) {
	elem := typ.Elem()
	op, err := e.compile(elem)
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		return op(enc, elem, e.ptrToPtr(p))
	}, nil
}

func (e *Encoder) compilePtrSlowPath(typ *rtype) (EncodeOp, error) {
	elem := typ.Elem()
	op, err := e.compileSlowPath(elem)
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		return op(enc, typ.Elem(), e.ptrToPtr(p))
	}, nil
}

func (e *Encoder) compileInt() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeInt(e.ptrToInt(p))
		return nil
	}, nil
}

func (e *Encoder) compileInt8() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeInt8(e.ptrToInt8(p))
		return nil
	}, nil
}

func (e *Encoder) compileInt16() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeInt16(e.ptrToInt16(p))
		return nil
	}, nil
}

func (e *Encoder) compileInt32() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeInt32(e.ptrToInt32(p))
		return nil
	}, nil
}

func (e *Encoder) compileInt64() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeInt64(e.ptrToInt64(p))
		return nil
	}, nil
}

func (e *Encoder) compileUint() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeUint(e.ptrToUint(p))
		return nil
	}, nil
}

func (e *Encoder) compileUint8() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeUint8(e.ptrToUint8(p))
		return nil
	}, nil
}

func (e *Encoder) compileUint16() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeUint16(e.ptrToUint16(p))
		return nil
	}, nil
}

func (e *Encoder) compileUint32() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeUint32(e.ptrToUint32(p))
		return nil
	}, nil
}

func (e *Encoder) compileUint64() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeUint64(e.ptrToUint64(p))
		return nil
	}, nil
}

func (e *Encoder) compileFloat32() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeFloat32(e.ptrToFloat32(p))
		return nil
	}, nil
}

func (e *Encoder) compileFloat64() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeFloat64(e.ptrToFloat64(p))
		return nil
	}, nil
}

func (e *Encoder) compileString() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeEscapedString(e.ptrToString(p))
		return nil
	}, nil
}

func (e *Encoder) compileBool() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, p uintptr) error {
		enc.encodeBool(e.ptrToBool(p))
		return nil
	}, nil
}

func (e *Encoder) compileSlice(typ *rtype) (EncodeOp, error) {
	elem := typ.Elem()
	size := elem.Size()
	op, err := e.compile(elem)
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		enc.encodeByte('[')
		slice := (*reflect.SliceHeader)(unsafe.Pointer(base))
		num := slice.Len
		for i := 0; i < num; i++ {
			if err := op(enc, elem, slice.Data+uintptr(i)*size); err != nil {
				return err
			}
			if i != num-1 {
				enc.encodeByte(',')
			}
		}
		enc.encodeByte(']')
		return nil
	}, nil
}

func (e *Encoder) compileSliceSlowPath(typ *rtype) (EncodeOp, error) {
	op, err := e.compileSlowPath(typ.Elem())
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		size := typ.Elem().Size()
		enc.encodeByte('[')
		slice := (*reflect.SliceHeader)(unsafe.Pointer(base))
		num := slice.Len
		for i := 0; i < num; i++ {
			if err := op(enc, typ.Elem(), slice.Data+uintptr(i)*size); err != nil {
				return err
			}
			if i != num-1 {
				enc.encodeByte(',')
			}
		}
		enc.encodeByte(']')
		return nil
	}, nil
}

func (e *Encoder) compileArray(typ *rtype) (EncodeOp, error) {
	elem := typ.Elem()
	alen := typ.Len()
	size := elem.Size()
	op, err := e.compile(elem)
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		enc.encodeByte('[')
		for i := 0; i < alen; i++ {
			if i != 0 {
				enc.encodeByte(',')
			}
			if err := op(enc, elem, base+uintptr(i)*size); err != nil {
				return err
			}
		}
		enc.encodeByte(']')
		return nil
	}, nil
}

func (e *Encoder) compileArraySlowPath(typ *rtype) (EncodeOp, error) {
	elem := typ.Elem()
	alen := typ.Len()
	op, err := e.compileSlowPath(elem)
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		elem := typ.Elem()
		size := elem.Size()
		enc.encodeByte('[')
		for i := 0; i < alen; i++ {
			if i != 0 {
				enc.encodeByte(',')
			}
			if err := op(enc, elem, base+uintptr(i)*size); err != nil {
				return err
			}
		}
		enc.encodeByte(']')
		return nil
	}, nil
}

func (e *Encoder) getTag(field reflect.StructField) string {
	return field.Tag.Get("json")
}

func (e *Encoder) isIgnoredStructField(field reflect.StructField) bool {
	if field.PkgPath != "" && !field.Anonymous {
		// private field
		return true
	}
	tag := e.getTag(field)
	if tag == "-" {
		return true
	}
	return false
}

type encodeStructField struct {
	op         EncodeOp
	fieldIndex int
}

func (e *Encoder) compileStruct(typ *rtype) (EncodeOp, error) {
	fieldNum := typ.NumField()
	opQueue := make([]*encodeStructField, 0, fieldNum)

	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if e.isIgnoredStructField(field) {
			continue
		}
		keyName := field.Name
		tag := e.getTag(field)
		opts := strings.Split(tag, ",")
		if len(opts) > 0 {
			if opts[0] != "" {
				keyName = opts[0]
			}
		}
		fieldType := type2rtype(field.Type)
		op, err := e.compile(fieldType)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf(`"%s":`, keyName)
		structField := &encodeStructField{fieldIndex: i}
		structField.op = func(enc *Encoder, typ *rtype, base uintptr) error {
			enc.encodeString(key)
			return op(enc, fieldType, base+field.Offset)
		}
		opQueue = append(opQueue, structField)
	}

	queueNum := len(opQueue)
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		enc.encodeByte('{')
		for i := 0; i < queueNum; i++ {
			if err := opQueue[i].op(enc, typ, base); err != nil {
				return err
			}
			if i != queueNum-1 {
				enc.encodeByte(',')
			}
		}
		enc.encodeByte('}')
		return nil
	}, nil
}

func (e *Encoder) compileStructSlowPath(typ *rtype) (EncodeOp, error) {
	fieldNum := typ.NumField()
	opQueue := make([]*encodeStructField, 0, fieldNum)

	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if e.isIgnoredStructField(field) {
			continue
		}
		keyName := field.Name
		tag := e.getTag(field)
		opts := strings.Split(tag, ",")
		if len(opts) > 0 {
			if opts[0] != "" {
				keyName = opts[0]
			}
		}
		fieldType := type2rtype(field.Type)
		op, err := e.compileSlowPath(fieldType)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf(`"%s":`, keyName)
		structField := &encodeStructField{fieldIndex: i}
		structField.op = func(enc *Encoder, typ *rtype, base uintptr) error {
			enc.encodeString(key)
			fieldType := type2rtype(typ.Field(structField.fieldIndex).Type)
			return op(enc, fieldType, base+field.Offset)
		}
		opQueue = append(opQueue, structField)
	}

	queueNum := len(opQueue)
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		enc.encodeByte('{')
		for i := 0; i < queueNum; i++ {
			if err := opQueue[i].op(enc, typ, base); err != nil {
				return err
			}
			if i != queueNum-1 {
				enc.encodeByte(',')
			}
		}
		enc.encodeByte('}')
		return nil
	}, nil
}

//go:linkname mapiterinit reflect.mapiterinit
//go:noescape
func mapiterinit(mapType *rtype, m unsafe.Pointer) unsafe.Pointer

//go:linkname mapiterkey reflect.mapiterkey
//go:noescape
func mapiterkey(it unsafe.Pointer) unsafe.Pointer

//go:linkname mapiternext reflect.mapiternext
//go:noescape
func mapiternext(it unsafe.Pointer)

//go:linkname maplen reflect.maplen
//go:noescape
func maplen(m unsafe.Pointer) int

type valueType struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

func (e *Encoder) compileMap(typ *rtype) (EncodeOp, error) {
	keyType := typ.Key()
	keyOp, err := e.compile(keyType)
	if err != nil {
		return nil, err
	}
	valueType := typ.Elem()
	valueOp, err := e.compile(valueType)
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		enc.encodeByte('{')
		mlen := maplen(unsafe.Pointer(base))
		iter := mapiterinit(typ, unsafe.Pointer(base))
		for i := 0; i < mlen; i++ {
			key := mapiterkey(iter)
			if i != 0 {
				enc.encodeByte(',')
			}
			value := mapitervalue(iter)
			keyptr := uintptr(key)
			if err := keyOp(enc, keyType, keyptr); err != nil {
				return err
			}
			enc.encodeByte(':')
			valueptr := uintptr(value)
			if err := valueOp(enc, valueType, valueptr); err != nil {
				return err
			}
			mapiternext(iter)
		}
		enc.encodeByte('}')
		return nil
	}, nil
}

func (e *Encoder) compileMapSlowPath(typ *rtype) (EncodeOp, error) {
	keyOp, err := e.compileSlowPath(typ.Key())
	if err != nil {
		return nil, err
	}
	valueOp, err := e.compileSlowPath(typ.Elem())
	if err != nil {
		return nil, err
	}
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		if base == 0 {
			enc.encodeString("null")
			return nil
		}
		enc.encodeByte('{')
		mlen := maplen(unsafe.Pointer(base))
		iter := mapiterinit(typ, unsafe.Pointer(base))
		for i := 0; i < mlen; i++ {
			key := mapiterkey(iter)
			if i != 0 {
				enc.encodeByte(',')
			}
			value := mapitervalue(iter)
			keyptr := uintptr(key)
			if err := keyOp(enc, typ.Key(), keyptr); err != nil {
				return err
			}
			enc.encodeByte(':')
			valueptr := uintptr(value)
			if err := valueOp(enc, typ.Elem(), valueptr); err != nil {
				return err
			}
			mapiternext(iter)
		}
		enc.encodeByte('}')
		return nil
	}, nil
}

func (e *Encoder) compileInterface() (EncodeOp, error) {
	return func(enc *Encoder, typ *rtype, base uintptr) error {
		v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
			typ: typ,
			ptr: unsafe.Pointer(base),
		}))
		vv := reflect.ValueOf(v).Interface()
		header := (*interfaceHeader)(unsafe.Pointer(&vv))
		t := header.typ
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		op, err := e.compileSlowPath(t)
		if err != nil {
			return err
		}
		return op(enc, t, uintptr(header.ptr))
	}, nil
}

func (e *Encoder) ptrToPtr(p uintptr) uintptr     { return *(*uintptr)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt(p uintptr) int         { return *(*int)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt8(p uintptr) int8       { return *(*int8)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt16(p uintptr) int16     { return *(*int16)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt32(p uintptr) int32     { return *(*int32)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToInt64(p uintptr) int64     { return *(*int64)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint(p uintptr) uint       { return *(*uint)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint8(p uintptr) uint8     { return *(*uint8)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint16(p uintptr) uint16   { return *(*uint16)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint32(p uintptr) uint32   { return *(*uint32)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToUint64(p uintptr) uint64   { return *(*uint64)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToFloat32(p uintptr) float32 { return *(*float32)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToFloat64(p uintptr) float64 { return *(*float64)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToBool(p uintptr) bool       { return *(*bool)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToByte(p uintptr) byte       { return *(*byte)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToBytes(p uintptr) []byte    { return *(*[]byte)(unsafe.Pointer(p)) }
func (e *Encoder) ptrToString(p uintptr) string   { return *(*string)(unsafe.Pointer(p)) }
