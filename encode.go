package json

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"io"
	"math"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	w                              io.Writer
	buf                            []byte
	enabledIndent                  bool
	enabledHTMLEscape              bool
	unorderedMap                   bool
	prefix                         []byte
	indentStr                      []byte
	structTypeToCompiledCode       map[uintptr]*compiledCode
	structTypeToCompiledIndentCode map[uintptr]*compiledCode
}

type compiledCode struct {
	code *opcode
}

const (
	bufSize = 1024
)

type opcodeMap struct {
	sync.Map
}

type opcodeSet struct {
	codeIndent *opcode
	code       *opcode
	ctx        sync.Pool
}

func (m *opcodeMap) get(k uintptr) *opcodeSet {
	if v, ok := m.Load(k); ok {
		return v.(*opcodeSet)
	}
	return nil
}

func (m *opcodeMap) set(k uintptr, op *opcodeSet) {
	m.Store(k, op)
}

var (
	encPool         sync.Pool
	codePool        sync.Pool
	cachedOpcode    opcodeMap
	marshalJSONType reflect.Type
	marshalTextType reflect.Type
)

func init() {
	encPool = sync.Pool{
		New: func() interface{} {
			return &Encoder{
				buf:                            make([]byte, 0, bufSize),
				structTypeToCompiledCode:       map[uintptr]*compiledCode{},
				structTypeToCompiledIndentCode: map[uintptr]*compiledCode{},
			}
		},
	}
	cachedOpcode = opcodeMap{}
	marshalJSONType = reflect.TypeOf((*Marshaler)(nil)).Elem()
	marshalTextType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
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
	return e.EncodeWithOption(v)
}

// EncodeWithOption call Encode with EncodeOption.
func (e *Encoder) EncodeWithOption(v interface{}, opts ...EncodeOption) error {
	for _, opt := range opts {
		if err := opt(e); err != nil {
			return err
		}
	}
	if err := e.encode(v); err != nil {
		return err
	}
	e.buf = append(e.buf, '\n')
	if _, err := e.w.Write(e.buf); err != nil {
		return err
	}
	e.buf = e.buf[:0]
	return nil
}

// SetEscapeHTML specifies whether problematic HTML characters should be escaped inside JSON quoted strings.
// The default behavior is to escape &, <, and > to \u0026, \u003c, and \u003e to avoid certain safety problems that can arise when embedding JSON in HTML.
//
// In non-HTML settings where the escaping interferes with the readability of the output, SetEscapeHTML(false) disables this behavior.
func (e *Encoder) SetEscapeHTML(on bool) {
	e.enabledHTMLEscape = on
}

// SetIndent instructs the encoder to format each subsequent encoded value as if indented by the package-level function Indent(dst, src, prefix, indent).
// Calling SetIndent("", "") disables indentation.
func (e *Encoder) SetIndent(prefix, indent string) {
	if prefix == "" && indent == "" {
		e.enabledIndent = false
		return
	}
	e.prefix = []byte(prefix)
	e.indentStr = []byte(indent)
	e.enabledIndent = true
}

func (e *Encoder) release() {
	e.w = nil
	encPool.Put(e)
}

func (e *Encoder) reset() {
	e.buf = e.buf[:0]
	e.enabledHTMLEscape = true
	e.enabledIndent = false
	e.unorderedMap = false
}

func (e *Encoder) encodeForMarshal(v interface{}) ([]byte, error) {
	if err := e.encode(v); err != nil {
		return nil, err
	}
	if e.enabledIndent {
		last := len(e.buf) - 1
		if e.buf[last] == '\n' {
			last--
		}
		length := last + 1
		copied := make([]byte, length)
		copy(copied, e.buf[0:length])
		return copied, nil
	}
	copied := make([]byte, len(e.buf))
	copy(copied, e.buf)
	return copied, nil
}

func (e *Encoder) encode(v interface{}) error {
	if v == nil {
		e.encodeNull()
		return nil
	}
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	typ := header.typ

	typeptr := uintptr(unsafe.Pointer(typ))
	if codeSet := cachedOpcode.get(typeptr); codeSet != nil {
		var code *opcode
		if e.enabledIndent {
			code = codeSet.codeIndent
		} else {
			code = codeSet.code
		}
		ctx := codeSet.ctx.Get().(*encodeRuntimeContext)
		p := uintptr(header.ptr)
		ctx.init(p)
		err := e.run(ctx, code)
		codeSet.ctx.Put(ctx)
		return err
	}

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := (*rtype)(unsafe.Pointer(typeptr))

	codeIndent, err := e.compileHead(&encodeCompileContext{
		typ:        copiedType,
		root:       true,
		withIndent: true,
	})
	if err != nil {
		return err
	}
	code, err := e.compileHead(&encodeCompileContext{
		typ:        copiedType,
		root:       true,
		withIndent: false,
	})
	if err != nil {
		return err
	}
	codeIndent = copyOpcode(codeIndent)
	code = copyOpcode(code)
	codeLength := code.totalLength()
	codeSet := &opcodeSet{
		codeIndent: codeIndent,
		code:       code,
		ctx: sync.Pool{
			New: func() interface{} {
				return &encodeRuntimeContext{
					ptrs:     make([]uintptr, codeLength),
					keepRefs: make([]unsafe.Pointer, 8),
				}
			},
		},
	}
	cachedOpcode.set(typeptr, codeSet)
	p := uintptr(header.ptr)
	ctx := codeSet.ctx.Get().(*encodeRuntimeContext)
	ctx.init(p)

	var c *opcode
	if e.enabledIndent {
		c = codeIndent
	} else {
		c = code
	}

	if err := e.run(ctx, c); err != nil {
		codeSet.ctx.Put(ctx)
		return err
	}
	codeSet.ctx.Put(ctx)
	return nil
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
	e.buf = strconv.AppendFloat(e.buf, f64, fmt, -1, 32)
}

func (e *Encoder) encodeFloat64(v float64) {
	abs := math.Abs(v)
	fmt := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if abs < 1e-6 || abs >= 1e21 {
			fmt = 'e'
		}
	}
	e.buf = strconv.AppendFloat(e.buf, v, fmt, -1, 64)
}

func (e *Encoder) encodeBool(v bool) {
	e.buf = strconv.AppendBool(e.buf, v)
}

func (e *Encoder) encodeBytes(b []byte) {
	e.buf = append(e.buf, b...)
}

func (e *Encoder) encodeNull() {
	e.buf = append(e.buf, 'n', 'u', 'l', 'l')
}

func (e *Encoder) encodeKey(code *opcode) {
	if e.enabledHTMLEscape {
		e.encodeBytes(code.escapedKey)
	} else {
		e.encodeBytes(code.key)
	}
}

func (e *Encoder) encodeString(s string) {
	if e.enabledHTMLEscape {
		e.encodeEscapedString(s)
	} else {
		e.encodeNoEscapedString(s)
	}
}

func (e *Encoder) encodeByteSlice(b []byte) {
	encodedLen := base64.StdEncoding.EncodedLen(len(b))
	e.encodeByte('"')
	pos := len(e.buf)
	remainLen := cap(e.buf[pos:])
	var buf []byte
	if remainLen > encodedLen {
		buf = e.buf[pos : pos+encodedLen]
	} else {
		buf = make([]byte, encodedLen)
	}
	base64.StdEncoding.Encode(buf, b)
	e.encodeBytes(buf)
	e.encodeByte('"')
}

func (e *Encoder) encodeByte(b byte) {
	e.buf = append(e.buf, b)
}

func (e *Encoder) encodeIndent(indent int) {
	e.buf = append(e.buf, e.prefix...)
	e.buf = append(e.buf, bytes.Repeat(e.indentStr, indent)...)
}
