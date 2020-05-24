package json

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"sync"
	"unsafe"
)

type Delim rune

func (d Delim) String() string {
	return string(d)
}

type decoder interface {
	decode([]byte, int64, uintptr) (int64, error)
}

type Decoder struct {
	s *stream
}

type decoderMap struct {
	sync.Map
}

func (m *decoderMap) get(k uintptr) decoder {
	if v, ok := m.Load(k); ok {
		return v.(decoder)
	}
	return nil
}

func (m *decoderMap) set(k uintptr, dec decoder) {
	m.Store(k, dec)
}

var (
	cachedDecoder     decoderMap
	unmarshalJSONType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
	unmarshalTextType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

func init() {
	cachedDecoder = decoderMap{}
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{s: &stream{r: r}}
}

// Buffered returns a reader of the data remaining in the Decoder's
// buffer. The reader is valid until the next call to Decode.
func (d *Decoder) Buffered() io.Reader {
	return d.s.buffered()
}

func (d *Decoder) validateType(typ *rtype, p uintptr) error {
	if typ.Kind() != reflect.Ptr || p == 0 {
		return &InvalidUnmarshalError{Type: rtype2type(typ)}
	}
	return nil
}

func (d *Decoder) decode(src []byte, header *interfaceHeader) error {
	typ := header.typ
	typeptr := uintptr(unsafe.Pointer(typ))

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := (*rtype)(unsafe.Pointer(typeptr))
	ptr := uintptr(header.ptr)

	if err := d.validateType(copiedType, ptr); err != nil {
		return err
	}
	dec := cachedDecoder.get(typeptr)
	if dec == nil {

		compiledDec, err := d.compileHead(copiedType)
		if err != nil {
			return err
		}
		cachedDecoder.set(typeptr, compiledDec)
		dec = compiledDec
	}
	if _, err := dec.decode(src, 0, ptr); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) decodeForUnmarshal(src []byte, v interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	header.typ.escape()
	return d.decode(src, header)
}

func (d *Decoder) decodeForUnmarshalNoEscape(src []byte, v interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	return d.decode(src, header)
}

func (d *Decoder) prepareForDecode() error {
	s := d.s
	for ; s.cursor < s.length || s.read(); s.cursor++ {
		switch s.char() {
		case ' ', '\t', '\r', '\n':
			continue
		case ',', ':':
			s.cursor++
			return nil
		}
		break
	}
	return nil
}

// Decode reads the next JSON-encoded value from its
// input and stores it in the value pointed to by v.
//
// See the documentation for Unmarshal for details about
// the conversion of JSON into a Go value.
func (d *Decoder) Decode(v interface{}) error {
	header := (*interfaceHeader)(unsafe.Pointer(&v))
	typ := header.typ
	ptr := uintptr(header.ptr)
	typeptr := uintptr(unsafe.Pointer(typ))
	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := (*rtype)(unsafe.Pointer(typeptr))

	if err := d.validateType(copiedType, ptr); err != nil {
		return err
	}

	dec := cachedDecoder.get(typeptr)
	if dec == nil {
		compiledDec, err := d.compileHead(typ)
		if err != nil {
			return err
		}
		cachedDecoder.set(typeptr, compiledDec)
		dec = compiledDec
	}
	if err := d.prepareForDecode(); err != nil {
		return err
	}
	s := d.s
	cursor, err := dec.decode(s.buf[s.cursor:], 0, ptr)
	s.cursor += cursor
	fmt.Println("cursor = ", cursor, "next buf = ", string(s.buf[s.cursor:]))
	if err != nil {
		return err
	}
	return nil
}

func (d *Decoder) More() bool {
	s := d.s
	for ; s.cursor < s.length || s.read(); s.cursor++ {
		switch s.char() {
		case ' ', '\n', '\r', '\t':
			continue
		case '}', ']':
			return false
		}
		break
	}
	return true
}

func (d *Decoder) Token() (Token, error) {
	s := d.s
	for ; s.cursor < s.length || s.read(); s.cursor++ {
		switch s.char() {
		case ' ', '\n', '\r', '\t':
			continue
		case '{':
			s.cursor++
			return Delim('{'), nil
		case '[':
			s.cursor++
			return Delim('['), nil
		case '}':
			s.cursor++
			return Delim('}'), nil
		case ']':
			s.cursor++
			return Delim(']'), nil
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		case '"':
		case 't':
		case 'f':
		case 'n':
		default:
			return nil, errInvalidCharacter(s.char(), "token", s.totalOffset())
		}
	}
	return nil, io.EOF
}

// DisallowUnknownFields causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
func (d *Decoder) DisallowUnknownFields() {

}

func (d *Decoder) InputOffset() int64 {
	return 0
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
func (d *Decoder) UseNumber() {

}
