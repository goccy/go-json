package json

import (
	"io"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	w    io.Writer
	buf  []byte
	pool sync.Pool
}

const (
	bufSize = 1024
)

type opcodeMap struct {
	sync.Map
}

func (m *opcodeMap) Get(k *rtype) *opcode {
	if v, ok := m.Load(k); ok {
		return v.(*opcode)
	}
	return nil
}

func (m *opcodeMap) Set(k *rtype, op *opcode) {
	m.Store(k, op)
}

var (
	encPool      sync.Pool
	cachedOpcode opcodeMap
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
	cachedOpcode = opcodeMap{}
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
	if code := cachedOpcode.Get(typ); code != nil {
		p := uintptr(header.ptr)
		code.ptr = p
		if err := e.run(code); err != nil {
			return err
		}
		return nil
	}
	code, err := e.compile(typ)
	if err != nil {
		return err
	}
	cachedOpcode.Set(typ, code)
	p := uintptr(header.ptr)
	code.ptr = p
	return e.run(code)
}
