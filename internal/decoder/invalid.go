package decoder

import (
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type invalidDecoder struct {
	typ        reflect.Type
	kind       reflect.Kind
	structName string
	fieldName  string
}

func newInvalidDecoder(typ reflect.Type, structName, fieldName string) *invalidDecoder {
	return &invalidDecoder{
		typ:        typ,
		kind:       typ.Kind(),
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *invalidDecoder) DecodeStream(s *Stream, depth int64, p unsafe.Pointer) error {
	return &errors.UnmarshalTypeError{
		Value:  "object",
		Type:   d.typ,
		Offset: s.totalOffset(),
		Struct: d.structName,
		Field:  d.fieldName,
	}
}

func (d *invalidDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	return 0, &errors.UnmarshalTypeError{
		Value:  "object",
		Type:   d.typ,
		Offset: cursor,
		Struct: d.structName,
		Field:  d.fieldName,
	}
}

func (d *invalidDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, &errors.UnmarshalTypeError{
		Value:  "object",
		Type:   d.typ,
		Offset: cursor,
		Struct: d.structName,
		Field:  d.fieldName,
	}
}
