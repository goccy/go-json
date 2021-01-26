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
	w                 io.Writer
	ctx               *encodeRuntimeContext
	ptr               unsafe.Pointer
	buf               []byte
	enabledIndent     bool
	enabledHTMLEscape bool
	unorderedMap      bool
	baseIndent        int
	prefix            []byte
	indentStr         []byte
}

type compiledCode struct {
	code    *opcode
	linked  bool // whether recursive code already have linked
	curLen  uintptr
	nextLen uintptr
}

const (
	bufSize = 1024
)

const (
	opCodeEscapedType = iota
	opCodeEscapedIndentType
	opCodeNoEscapeType
	opCodeNoEscapeIndentType
)

type opcodeSet struct {
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
				buf: make([]byte, 0, bufSize),
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

func newEncoder() *Encoder {
	enc := encPool.Get().(*Encoder)
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
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	e.ptr = header.ptr
	buf, err := e.encode(header, v == nil)
	if err != nil {
		return err
	}
	if e.enabledIndent {
		buf = buf[:len(buf)-2]
	} else {
		buf = buf[:len(buf)-1]
	}
	buf = append(buf, '\n')
	if _, err := e.w.Write(buf); err != nil {
		return err
	}
	e.buf = buf[:0]
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
	e.baseIndent = 0
	e.enabledHTMLEscape = true
	e.enabledIndent = false
	e.unorderedMap = false
}

func (e *Encoder) encodeForMarshal(header *interfaceHeader, isNil bool) ([]byte, error) {
	buf, err := e.encode(header, isNil)
	if err != nil {
		return nil, err
	}

	e.buf = buf

	if e.enabledIndent {
		// this line's description is the below.
		buf = buf[:len(buf)-2]

		copied := make([]byte, len(buf))
		copy(copied, buf)
		return copied, nil
	}

	// this line exists to escape call of `runtime.makeslicecopy` .
	// if use `make([]byte, len(buf)-1)` and `copy(copied, buf)`,
	// dst buffer size and src buffer size are differrent.
	// in this case, compiler uses `runtime.makeslicecopy`, but it is slow.
	buf = buf[:len(buf)-1]

	copied := make([]byte, len(buf))
	copy(copied, buf)
	return copied, nil
}

func (e *Encoder) encode(header *interfaceHeader, isNil bool) ([]byte, error) {
	b := e.buf[:0]
	if isNil {
		b = encodeNull(b)
		if e.enabledIndent {
			b = encodeIndentComma(b)
		} else {
			b = encodeComma(b)
		}
		return b, nil
	}
	typ := header.typ

	typeptr := uintptr(unsafe.Pointer(typ))
	codeSet, err := e.compileToGetCodeSet(typeptr)
	if err != nil {
		return nil, err
	}

	ctx := e.ctx
	p := uintptr(header.ptr)
	ctx.init(p, codeSet.codeLength)
	if e.enabledIndent {
		if e.enabledHTMLEscape {
			return e.runEscapedIndent(ctx, b, codeSet)
		} else {
			return e.runIndent(ctx, b, codeSet)
		}
	}
	if e.enabledHTMLEscape {
		return e.runEscaped(ctx, b, codeSet)
	}
	return e.run(ctx, b, codeSet)
}

func (e *Encoder) compileToGetCodeSet(typeptr uintptr) (*opcodeSet, error) {
	opcodeMap := loadOpcodeMap()
	if codeSet, exists := opcodeMap[typeptr]; exists {
		return codeSet, nil
	}

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := *(**rtype)(unsafe.Pointer(&typeptr))

	code, err := e.compileHead(&encodeCompileContext{
		typ:                      copiedType,
		root:                     true,
		structTypeToCompiledCode: map[uintptr]*compiledCode{},
	})
	if err != nil {
		return nil, err
	}
	code = copyOpcode(code)
	codeLength := code.totalLength()
	codeSet := &opcodeSet{
		code:       code,
		codeLength: codeLength,
	}

	storeOpcodeSet(typeptr, codeSet, opcodeMap)
	return codeSet, nil
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

func appendStructEnd(b []byte) []byte {
	return append(b, '}', ',')
}

func (e *Encoder) appendStructEndIndent(b []byte, indent int) []byte {
	b = append(b, '\n')
	b = append(b, e.prefix...)
	b = append(b, bytes.Repeat(e.indentStr, e.baseIndent+indent)...)
	return append(b, '}', ',', '\n')
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
	return append(b, bytes.Repeat(e.indentStr, e.baseIndent+indent)...)
}
