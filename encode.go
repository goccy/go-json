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
	"sync/atomic"
	"unsafe"
)

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	w                              io.Writer
	ctx                            *encodeRuntimeContext
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

type opcodeSet struct {
	codeIndent *opcode
	code       *opcode
	codeLength int
}

func loadOpcodeMap() map[uintptr]*opcodeSet {
	p := atomic.LoadPointer(&cachedOpcode)
	return *(*map[uintptr]*opcodeSet)(unsafe.Pointer(&p))
}

func storeOpcodeSet(typ uintptr, set *opcodeSet, m map[uintptr]*opcodeSet) {
	newOpcodeMap := make(map[uintptr]*opcodeSet, len(m)+1)
	newOpcodeMap[typ] = set

	for k, v := range m {
		newOpcodeMap[k] = v
	}

	atomic.StorePointer(&cachedOpcode, *(*unsafe.Pointer)(unsafe.Pointer(&newOpcodeMap)))
}

var (
	encPool         sync.Pool
	codePool        sync.Pool
	cachedOpcode    unsafe.Pointer // map[uintptr]*opcodeSet
	marshalJSONType reflect.Type
	marshalTextType reflect.Type
)

func init() {
	encPool = sync.Pool{
		New: func() interface{} {
			return &Encoder{
				ctx: &encodeRuntimeContext{
					ptrs:     make([]uintptr, 128),
					keepRefs: make([]unsafe.Pointer, 0, 8),
				},
				buf:                            make([]byte, 0, bufSize),
				structTypeToCompiledCode:       map[uintptr]*compiledCode{},
				structTypeToCompiledIndentCode: map[uintptr]*compiledCode{},
			}
		},
	}
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
	var err error
	if e.buf, err = e.encode(v); err != nil {
		return err
	}
	if e.enabledIndent {
		e.buf = e.buf[:len(e.buf)-2]
	} else {
		e.buf = e.buf[:len(e.buf)-1]
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
	var err error
	if e.buf, err = e.encode(v); err != nil {
		return nil, err
	}
	if e.enabledIndent {
		copied := make([]byte, len(e.buf)-2)
		copy(copied, e.buf)
		return copied, nil
	}
	copied := make([]byte, len(e.buf)-1)
	copy(copied, e.buf)
	return copied, nil
}

func (e *Encoder) encode(v interface{}) ([]byte, error) {
	b := e.buf
	if v == nil {
		b = encodeNull(b)
		if e.enabledIndent {
			b = encodeIndentComma(b)
		} else {
			b = encodeComma(b)
		}
		return b, nil
	}
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	typ := header.typ

	typeptr := uintptr(unsafe.Pointer(typ))
	opcodeMap := loadOpcodeMap()
	if codeSet, exists := opcodeMap[typeptr]; exists {
		var code *opcode
		if e.enabledIndent {
			code = codeSet.codeIndent
		} else {
			code = codeSet.code
		}
		ctx := e.ctx
		p := uintptr(header.ptr)
		ctx.init(p, codeSet.codeLength)
		return e.run(ctx, b, code)
	}

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := *(**rtype)(unsafe.Pointer(&typeptr))

	codeIndent, err := e.compileHead(&encodeCompileContext{
		typ:        copiedType,
		root:       true,
		withIndent: true,
	})
	if err != nil {
		return nil, err
	}
	code, err := e.compileHead(&encodeCompileContext{
		typ:        copiedType,
		root:       true,
		withIndent: false,
	})
	if err != nil {
		return nil, err
	}
	codeIndent = copyOpcode(codeIndent)
	code = copyOpcode(code)
	codeLength := code.totalLength()
	codeSet := &opcodeSet{
		codeIndent: codeIndent,
		code:       code,
		codeLength: codeLength,
	}

	storeOpcodeSet(typeptr, codeSet, opcodeMap)
	p := uintptr(header.ptr)
	ctx := e.ctx
	ctx.init(p, codeLength)

	var c *opcode
	if e.enabledIndent {
		c = codeIndent
	} else {
		c = code
	}

	b, err = e.run(ctx, b, c)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func encodeFloat32(b []byte, v float32) []byte {
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

func encodeFloat64(b []byte, v float64) []byte {
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

func encodeBool(b []byte, v bool) []byte {
	if v {
		return append(b, "true"...)
	}
	return append(b, "false"...)
}

func encodeBytes(dst []byte, src []byte) []byte {
	return append(dst, src...)
}

func encodeNull(b []byte) []byte {
	return append(b, "null"...)
}

func encodeComma(b []byte) []byte {
	return append(b, ',')
}

func encodeIndentComma(b []byte) []byte {
	return append(b, ',', '\n')
}

func (e *Encoder) encodeKey(b []byte, code *opcode) []byte {
	if e.enabledHTMLEscape {
		return append(b, code.escapedKey...)
	}
	return append(b, code.key...)
}

func (e *Encoder) encodeString(b []byte, s string) []byte {
	if e.enabledHTMLEscape {
		return encodeEscapedString(b, s)
	}
	return encodeNoEscapedString(b, s)
}

func encodeByteSlice(b []byte, src []byte) []byte {
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

func (e *Encoder) encodeIndent(b []byte, indent int) []byte {
	b = append(b, e.prefix...)
	return append(b, bytes.Repeat(e.indentStr, indent)...)
}
