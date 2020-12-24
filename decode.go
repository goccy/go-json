package json

import (
	"encoding"
	"io"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

type Delim rune

func (d Delim) String() string {
	return string(d)
}

type decoder interface {
	decode([]byte, int64, unsafe.Pointer) (int64, error)
	decodeStream(*stream, unsafe.Pointer) error
}

type Decoder struct {
	s                     *stream
	disallowUnknownFields bool
	structTypeToDecoder   map[uintptr]decoder
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

const (
	nul = '\000'
)

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the JSON values requested.
func NewDecoder(r io.Reader) *Decoder {
	s := newStream(r)
	return &Decoder{
		s: s,
	}
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
	copiedType := *(**rtype)(unsafe.Pointer(&typeptr))
	ptr := uintptr(header.ptr)

	if err := d.validateType(copiedType, ptr); err != nil {
		return err
	}
	dec := cachedDecoder.get(typeptr)
	if dec == nil {
		d.structTypeToDecoder = map[uintptr]decoder{}
		compiledDec, err := d.compileHead(copiedType)
		if err != nil {
			return err
		}
		cachedDecoder.set(typeptr, compiledDec)
		dec = compiledDec
	}
	if _, err := dec.decode(src, 0, header.ptr); err != nil {
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
	for {
		switch s.char() {
		case ' ', '\t', '\r', '\n':
			s.cursor++
			continue
		case ',', ':':
			s.cursor++
			return nil
		case nul:
			if s.read() {
				continue
			}
			return io.EOF
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
	copiedType := *(**rtype)(unsafe.Pointer(&typeptr))

	if err := d.validateType(copiedType, ptr); err != nil {
		return err
	}

	dec := cachedDecoder.get(typeptr)
	if dec == nil {
		d.structTypeToDecoder = map[uintptr]decoder{}
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
	if err := dec.decodeStream(s, header.ptr); err != nil {
		return err
	}
	return nil
}

func (d *Decoder) More() bool {
	s := d.s
	for {
		switch s.char() {
		case ' ', '\n', '\r', '\t':
			s.cursor++
			continue
		case '}', ']':
			return false
		case nul:
			if s.read() {
				continue
			}
			return false
		}
		break
	}
	return true
}

func (d *Decoder) Token() (Token, error) {
	s := d.s
	for {
		c := s.char()
		switch c {
		case ' ', '\n', '\r', '\t':
			s.cursor++
		case '{', '[', ']', '}':
			s.cursor++
			return Delim(c), nil
		case ',', ':':
			s.cursor++
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			bytes := floatBytes(s)
			s := *(*string)(unsafe.Pointer(&bytes))
			f64, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
			return f64, nil
		case '"':
			bytes, err := stringBytes(s)
			if err != nil {
				return nil, err
			}
			return string(bytes), nil
		case 't':
			if err := trueBytes(s); err != nil {
				return nil, err
			}
			return true, nil
		case 'f':
			if err := falseBytes(s); err != nil {
				return nil, err
			}
			return false, nil
		case 'n':
			if err := nullBytes(s); err != nil {
				return nil, err
			}
			return nil, nil
		case nul:
			if s.read() {
				continue
			}
			return nil, io.EOF
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
	d.s.disallowUnknownFields = true
}

func (d *Decoder) InputOffset() int64 {
	return d.s.totalOffset()
}

// UseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
func (d *Decoder) UseNumber() {
	d.s.useNumber = true
}
