package json

import (
	"errors"
	"io"
	"math"
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

type Decoder struct {
	r       io.Reader
	state   int
	literal []byte
}

type context struct {
	idx      int
	keys     [][]byte
	literals [][]byte
	start    int
	stack    int
}

func newContext() *context {
	return &context{
		keys:     make([][]byte, 64),
		literals: make([][]byte, 64),
	}
}

func (c *context) pushStack() {
	if len(c.keys) <= c.stack {
		c.keys = append(c.keys, nil)
		c.literals = append(c.literals, nil)
	}
	c.stack++
}

func (c *context) popStack() {
	c.stack--
}

func (c *context) setKey(key []byte) {
	c.keys[c.stack] = key
}

func (c *context) setLiteral(literal []byte) {
	c.literals[c.stack] = literal
}

func (c *context) key() ([]byte, error) {
	if len(c.keys) <= c.stack {
		return nil, errors.New("unexpected error")
	}
	key := c.keys[c.stack]
	if len(key) == 0 {
		return nil, errors.New("unexpected error")
	}
	return key, nil
}

func (c *context) literal() ([]byte, error) {
	if len(c.literals) <= c.stack {
		return nil, errors.New("unexpected error")
	}
	return c.literals[c.stack], nil
}

var (
	ctxPool        sync.Pool
	cachedDecodeOp map[string]DecodeOp
)

func init() {
	cachedDecodeOp = map[string]DecodeOp{}
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
	rv := reflect.ValueOf(v)
	typ := rv.Type()
	if typ.Kind() != reflect.Ptr {
		return ErrDecodePointer
	}
	name := typ.String()
	op, exists := cachedDecodeOp[name]
	if !exists {
		decodeOp, err := d.compile(rv.Elem())
		if err != nil {
			return err
		}
		if name != "" {
			cachedDecodeOp[name] = decodeOp
		}
		op = decodeOp
	}
	ptr := rv.Pointer()
	ctx := ctxPool.Get().(*context)
	if err := d.decode(ctx, src, ptr, op); err != nil {
		ctxPool.Put(ctx)
		return err
	}
	ctxPool.Put(ctx)
	return nil
}

func (d *Decoder) Decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	typ := rv.Type()
	if typ.Kind() != reflect.Ptr {
		return ErrDecodePointer
	}
	name := typ.String()
	op, exists := cachedDecodeOp[name]
	if !exists {
		decodeOp, err := d.compile(rv.Elem())
		if err != nil {
			return err
		}
		if name != "" {
			cachedDecodeOp[name] = decodeOp
		}
		op = decodeOp
	}
	ptr := rv.Pointer()
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
		if err := d.decode(ctx, buf[:n], ptr, op); err != nil {
			return err
		}
	}
	return nil
}

type DecodeOp func(uintptr, []byte, []byte) error

func (d *Decoder) compile(v reflect.Value) (DecodeOp, error) {
	switch v.Type().Kind() {
	case reflect.Struct:
		return d.compileStruct(v)
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
	}
	return nil, nil
}

func parseInt(b []byte) (int64, error) {
	isNegative := false
	if b[0] == '-' {
		b = b[1:]
		isNegative = true
	}
	maxDigit := len(b)
	sum := int64(0)
	for i := 0; i < maxDigit; i++ {
		c := int64(b[i]) - 48
		if 0 <= c && c <= 9 {
			digitValue := int64(math.Pow10(maxDigit - i - 1))
			sum += c * digitValue
		} else {
			return 0, errors.New("failed to parse int")
		}
	}
	if isNegative {
		return -1 * sum, nil
	}
	return sum, nil
}

func parseUint(b []byte) (uint64, error) {
	maxDigit := len(b)
	sum := uint64(0)
	for i := 0; i < maxDigit; i++ {
		c := uint64(b[i]) - 48
		if 0 <= c && c <= 9 {
			digitValue := uint64(math.Pow10(maxDigit - i - 1))
			sum += c * digitValue
		} else {
			return 0, errors.New("failed to parse uint")
		}
	}
	return sum, nil
}

func (d *Decoder) compileInt() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		i64, err := parseInt(src)
		if err != nil {
			return err
		}
		*(*int)(unsafe.Pointer(p)) = int(i64)
		return nil
	}, nil
}

func (d *Decoder) compileInt8() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		i64, err := parseInt(src)
		if err != nil {
			return err
		}
		*(*int8)(unsafe.Pointer(p)) = int8(i64)
		return nil
	}, nil
}

func (d *Decoder) compileInt16() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		i64, err := parseInt(src)
		if err != nil {
			return err
		}
		*(*int16)(unsafe.Pointer(p)) = int16(i64)
		return nil
	}, nil
}

func (d *Decoder) compileInt32() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		i64, err := parseInt(src)
		if err != nil {
			return err
		}
		*(*int32)(unsafe.Pointer(p)) = int32(i64)
		return nil
	}, nil
}

func (d *Decoder) compileInt64() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		i64, err := parseInt(src)
		if err != nil {
			return err
		}
		*(*int64)(unsafe.Pointer(p)) = i64
		return nil
	}, nil
}

func (d *Decoder) compileUint() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		u64, err := parseUint(src)
		if err != nil {
			return err
		}
		*(*uint)(unsafe.Pointer(p)) = uint(u64)
		return nil
	}, nil
}

func (d *Decoder) compileUint8() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		u64, err := parseUint(src)
		if err != nil {
			return err
		}
		*(*uint8)(unsafe.Pointer(p)) = uint8(u64)
		return nil
	}, nil
}

func (d *Decoder) compileUint16() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		u64, err := parseUint(src)
		if err != nil {
			return err
		}
		*(*uint16)(unsafe.Pointer(p)) = uint16(u64)
		return nil
	}, nil
}

func (d *Decoder) compileUint32() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		u64, err := parseUint(src)
		if err != nil {
			return err
		}
		*(*uint32)(unsafe.Pointer(p)) = uint32(u64)
		return nil
	}, nil
}

func (d *Decoder) compileUint64() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		u64, err := parseUint(src)
		if err != nil {
			return err
		}
		*(*uint64)(unsafe.Pointer(p)) = u64
		return nil
	}, nil
}

func (d *Decoder) compileString() (DecodeOp, error) {
	return func(p uintptr, src []byte, _ []byte) error {
		*(*string)(unsafe.Pointer(p)) = *(*string)(unsafe.Pointer(&src))
		return nil
	}, nil
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

func (d *Decoder) compileStruct(v reflect.Value) (DecodeOp, error) {
	type opset struct {
		key []byte
		op  func(uintptr, []byte) error
	}
	typ := v.Type()
	fieldNum := typ.NumField()
	opMap := map[string]func(uintptr, []byte) error{}
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
		op, err := d.compile(v.Field(i))
		if err != nil {
			return nil, err
		}
		if op == nil {
			continue
		}
		fieldOp := func(base uintptr, value []byte) error {
			return op(base+field.Offset, value, nil)
		}
		opMap[field.Name] = fieldOp
		opMap[keyName] = fieldOp
		opMap[strings.ToLower(keyName)] = fieldOp
	}
	return func(p uintptr, key []byte, value []byte) error {
		k := *(*string)(unsafe.Pointer(&key))
		op, exists := opMap[k]
		if !exists {
			return nil
		}
		return op(p, value)
	}, nil
}

func (d *Decoder) decode(ctx *context, src []byte, ptr uintptr, op DecodeOp) error {
	slen := len(src)
	for i := ctx.idx; i < slen; i++ {
		c := src[i]
		switch c {
		case ' ':
			ctx.start++
		case '{':
			ctx.pushStack()
		case '}':
			key, err := ctx.key()
			if err != nil {
				return err
			}
			if err := op(ptr, key, src[ctx.start:i]); err != nil {
				return err
			}
			ctx.popStack()
		case '[':
			d.state = stateArray
		case ']':
			d.state = stateNone
		case ':':
			if len(d.literal) == 0 {
				return errors.New("unexpected error")
			}
			ctx.setKey(d.literal)
			ctx.start = i + 1
		case ',':
			literal := src[ctx.start:i]
			key, err := ctx.key()
			if err != nil {
				return err
			}
			if err := op(ptr, key, literal); err != nil {
				return err
			}
		case '"':
			start := i + 1
			for i = start; i < slen && src[i] != '"'; i++ {
				if src[i] == '\\' {
					i++
				}
			}
			end := i
			if end <= start {
				return errors.New("unexpected error")
			}
			d.literal = src[start:end]
		}
	}
	return nil
}

func (d *Decoder) parse(tokens []Token) {

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
