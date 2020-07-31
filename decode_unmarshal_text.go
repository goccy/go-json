package json

import (
	"encoding"
	"unsafe"
)

type unmarshalTextDecoder struct {
	typ *rtype
}

func newUnmarshalTextDecoder(typ *rtype) *unmarshalTextDecoder {
	return &unmarshalTextDecoder{typ: typ}
}

func (d *unmarshalTextDecoder) decodeStream(s *stream, p uintptr) error {
	s.skipWhiteSpace()
	start := s.cursor
	if err := s.skipValue(); err != nil {
		return err
	}
	src := s.buf[start:s.cursor]
	v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: d.typ,
		ptr: unsafe.Pointer(p),
	}))
	if err := v.(encoding.TextUnmarshaler).UnmarshalText(src); err != nil {
		return err
	}
	return nil
}

func (d *unmarshalTextDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: d.typ,
		ptr: unsafe.Pointer(p),
	}))
	if err := v.(encoding.TextUnmarshaler).UnmarshalText(src); err != nil {
		return 0, err
	}
	return end, nil
}
