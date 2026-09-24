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
	// The value is decoded from a copy of its bytes, which nothing else uses: its strings may refer to it.
	oldBuf, oldOrigin := ctx.Buf, ctx.origin
	ctx.Buf, ctx.origin = NewInput(bytes), nil
	_, err = d.dec.Decode(ctx, 0, depth, p)
	ctx.Buf, ctx.origin = oldBuf, oldOrigin
	if err != nil {
		return 0, err
	}
	return c, nil
}

func (d *wrappedStringDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: wrapped string decoder does not support decode path")
}
