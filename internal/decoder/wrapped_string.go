package decoder

import (
	"fmt"
	"reflect"
	"unsafe"
)

type wrappedStringDecoder struct {
	typ           reflect.Type
	dec           Decoder
	stringDecoder *stringDecoder
	structName    string
	fieldName     string
	isPtrType     bool
}

func newWrappedStringDecoder(typ reflect.Type, dec Decoder, structName, fieldName string) *wrappedStringDecoder {
	return &wrappedStringDecoder{
		typ:           typ,
		dec:           dec,
		stringDecoder: newStringDecoder(structName, fieldName),
		structName:    structName,
		fieldName:     fieldName,
		isPtrType:     typ.Kind() == reflect.Ptr,
	}
}

func (d *wrappedStringDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	bytes, c, err := d.stringDecoder.decodeByte(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}
	if bytes == nil {
		if d.isPtrType {
			*(*unsafe.Pointer)(p) = nil
		}
		return c, nil
	}
	oldBuf := ctx.Buf
	ctx.Buf = NewInput(bytes)
	if _, err := d.dec.Decode(ctx, 0, depth, p); err != nil {
		return 0, err
	}
	ctx.Buf = oldBuf
	return c, nil
}

func (d *wrappedStringDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: wrapped string decoder does not support decode path")
}
