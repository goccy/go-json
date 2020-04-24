package json

import (
	"io"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type Token interface{}

type Delim rune

const (
	stateNone int = iota
	stateLiteral
	stateObject
	stateArray
)

type decoder interface {
	decode(*context, uintptr) error
}

type Decoder struct {
	r     io.Reader
	state int
	value []byte
}

var (
	ctxPool       sync.Pool
	cachedDecoder map[string]decoder
)

func init() {
	cachedDecoder = map[string]decoder{}
	ctxPool = sync.Pool{
		New: func() interface{} {
			return newContext()
		},
	}
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (d *Decoder) Buffered() io.Reader {
	return d.r
}

func (d *Decoder) decodeForUnmarshal(src []byte, v interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	typ := header.typ
	if typ.Kind() != reflect.Ptr {
		return ErrDecodePointer
	}
	name := typ.String()
	dec, exists := cachedDecoder[name]
	if !exists {
		compiledDec, err := d.compile(typ.Elem())
		if err != nil {
			return err
		}
		if name != "" {
			cachedDecoder[name] = compiledDec
		}
		dec = compiledDec
	}
	ptr := uintptr(header.ptr)
	ctx := ctxPool.Get().(*context)
	ctx.setBuf(src)
	if err := dec.decode(ctx, ptr); err != nil {
		ctxPool.Put(ctx)
		return err
	}
	ctxPool.Put(ctx)
	return nil
}

func (d *Decoder) Decode(v interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	typ := header.typ
	if typ.Kind() != reflect.Ptr {
		return ErrDecodePointer
	}
	name := typ.String()
	dec, exists := cachedDecoder[name]
	if !exists {
		compiledDec, err := d.compile(typ.Elem())
		if err != nil {
			return err
		}
		if name != "" {
			cachedDecoder[name] = compiledDec
		}
		dec = compiledDec
	}
	ptr := uintptr(header.ptr)
	ctx := ctxPool.Get().(*context)
	defer ctxPool.Put(ctx)
	for {
		buf := make([]byte, 1024)
		n, err := d.r.Read(buf)
		if n == 0 || err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		ctx.setBuf(buf[:n])
		if err := dec.decode(ctx, ptr); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) compile(typ *rtype) (decoder, error) {
	switch typ.Kind() {
	case reflect.Ptr:
		return d.compilePtr(typ)
	case reflect.Struct:
		return d.compileStruct(typ)
	case reflect.Int:
		return d.compileInt()
	case reflect.Int8:
		return d.compileInt8()
	case reflect.Int16:
		return d.compileInt16()
	case reflect.Int32:
		return d.compileInt32()
	case reflect.Int64:
		return d.compileInt64()
	case reflect.Uint:
		return d.compileUint()
	case reflect.Uint8:
		return d.compileUint8()
	case reflect.Uint16:
		return d.compileUint16()
	case reflect.Uint32:
		return d.compileUint32()
	case reflect.Uint64:
		return d.compileUint64()
	case reflect.String:
		return d.compileString()
	case reflect.Bool:
		return d.compileBool()
	case reflect.Float32:
		return d.compileFloat32()
	case reflect.Float64:
		return d.compileFloat64()
	}
	return nil, nil
}

func (d *Decoder) compilePtr(typ *rtype) (decoder, error) {
	dec, err := d.compile(typ.Elem())
	if err != nil {
		return nil, err
	}
	return newPtrDecoder(dec, typ.Elem()), nil
}

func (d *Decoder) compileInt() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int)(unsafe.Pointer(p)) = int(v)
	}), nil
}

func (d *Decoder) compileInt8() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int8)(unsafe.Pointer(p)) = int8(v)
	}), nil
}

func (d *Decoder) compileInt16() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int16)(unsafe.Pointer(p)) = int16(v)
	}), nil
}

func (d *Decoder) compileInt32() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int32)(unsafe.Pointer(p)) = int32(v)
	}), nil
}

func (d *Decoder) compileInt64() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int64)(unsafe.Pointer(p)) = v
	}), nil
}

func (d *Decoder) compileUint() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint)(unsafe.Pointer(p)) = uint(v)
	}), nil
}

func (d *Decoder) compileUint8() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint8)(unsafe.Pointer(p)) = uint8(v)
	}), nil
}

func (d *Decoder) compileUint16() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint16)(unsafe.Pointer(p)) = uint16(v)
	}), nil
}

func (d *Decoder) compileUint32() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint32)(unsafe.Pointer(p)) = uint32(v)
	}), nil
}

func (d *Decoder) compileUint64() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint64)(unsafe.Pointer(p)) = v
	}), nil
}

func (d *Decoder) compileFloat32() (decoder, error) {
	return newFloatDecoder(func(p uintptr, v float64) {
		*(*float32)(unsafe.Pointer(p)) = float32(v)
	}), nil
}

func (d *Decoder) compileFloat64() (decoder, error) {
	return newFloatDecoder(func(p uintptr, v float64) {
		*(*float64)(unsafe.Pointer(p)) = v
	}), nil
}

func (d *Decoder) compileString() (decoder, error) {
	return newStringDecoder(), nil
}

func (d *Decoder) compileBool() (decoder, error) {
	return newBoolDecoder(), nil
}

func (d *Decoder) getTag(field reflect.StructField) string {
	return field.Tag.Get("json")
}

func (d *Decoder) isIgnoredStructField(field reflect.StructField) bool {
	if field.PkgPath != "" && !field.Anonymous {
		// private field
		return true
	}
	tag := d.getTag(field)
	if tag == "-" {
		return true
	}
	return false
}

func (d *Decoder) compileStruct(typ *rtype) (decoder, error) {
	fieldNum := typ.NumField()
	fieldMap := map[string]*structFieldSet{}
	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if d.isIgnoredStructField(field) {
			continue
		}
		keyName := field.Name
		tag := d.getTag(field)
		opts := strings.Split(tag, ",")
		if len(opts) > 0 {
			if opts[0] != "" {
				keyName = opts[0]
			}
		}
		dec, err := d.compile(type2rtype(field.Type))
		if err != nil {
			return nil, err
		}
		fieldSet := &structFieldSet{dec: dec, offset: field.Offset}
		fieldMap[field.Name] = fieldSet
		fieldMap[keyName] = fieldSet
		fieldMap[strings.ToLower(keyName)] = fieldSet
	}
	return newStructDecoder(fieldMap), nil
}

func (d *Decoder) DisallowUnknownFields() {

}

func (d *Decoder) InputOffset() int64 {
	return 0
}

func (d *Decoder) More() bool {
	return false
}

func (d *Decoder) Token() (Token, error) {
	return nil, nil
}

func (d *Decoder) UseNumber() {

}
