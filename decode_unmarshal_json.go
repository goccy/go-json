package json

import (
	"unsafe"
)

type unmarshalJSONDecoder struct {
	typ             *rtype
	isDoublePointer bool
	structName      string
	fieldName       string
}

func newUnmarshalJSONDecoder(typ *rtype, structName, fieldName string) *unmarshalJSONDecoder {
	return &unmarshalJSONDecoder{
		typ:        typ,
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *unmarshalJSONDecoder) annotateError(cursor int64, err error) {
	switch e := err.(type) {
	case *UnmarshalTypeError:
		e.Struct = d.structName
		e.Field = d.fieldName
	case *SyntaxError:
		e.Offset = cursor
	}
}

func (d *unmarshalJSONDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	s.skipWhiteSpace()
	start := s.cursor
	if err := s.skipValue(); err != nil {
		return err
	}
	src := s.buf[start:s.cursor]
	if d.isDoublePointer {
		newptr := unsafe_New(d.typ.Elem())
		v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
			typ: d.typ,
			ptr: newptr,
		}))
		if err := v.(Unmarshaler).UnmarshalJSON(src); err != nil {
			d.annotateError(s.cursor, err)
			return err
		}
		*(*unsafe.Pointer)(p) = newptr
	} else {
		v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
			typ: d.typ,
			ptr: p,
		}))
		if err := v.(Unmarshaler).UnmarshalJSON(src); err != nil {
			d.annotateError(s.cursor, err)
			return err
		}
	}
	return nil
}

func (d *unmarshalJSONDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	if d.isDoublePointer {
		newptr := unsafe_New(d.typ.Elem())
		v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
			typ: d.typ,
			ptr: newptr,
		}))
		if err := v.(Unmarshaler).UnmarshalJSON(src); err != nil {
			d.annotateError(cursor, err)
			return 0, err
		}
		*(*unsafe.Pointer)(p) = newptr
	} else {
		v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
			typ: d.typ,
			ptr: p,
		}))
		if err := v.(Unmarshaler).UnmarshalJSON(src); err != nil {
			d.annotateError(cursor, err)
			return 0, err
		}
	}
	return end, nil
}
