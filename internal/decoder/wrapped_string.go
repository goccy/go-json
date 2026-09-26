package decoder

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

var (
	// errStringOptionRest tells that the value of a string of the string option is followed by more bytes.
	errStringOptionRest = errors.New("json: value followed by more bytes in a string of the string option")
	// errMapKeyType is returned by the decoder of a key of a map which is not of the type of the keys, whose type
	// error is recorded: the map decodes the value of the entry, and drops the entry, as encoding/json does.
	errMapKeyType = errors.New("json: map key of another type")
)

type wrappedStringDecoder struct {
	typ           reflect.Type
	dec           Decoder
	stringDecoder *stringDecoder
	structName    string
	fieldName     string
	isPtrType     bool
	// isMapKey is set for the decoder of the keys of a map, whose errors are the ones of the keys.
	isMapKey bool
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
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	if c := buf[cursor]; c != 'n' && isOtherValue(c, stringValue) {
		// a value which is not in a string
		if d.isPtrType && stringOptionUnquotedAllocates {
			d.allocate(p)
		}
		return ctx.stringOptionUnquoted(cursor, depth, d.typ)
	}
	bytes, c, err := d.stringDecoder.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	if bytes == nil {
		if d.isPtrType {
			*(*unsafe.Pointer)(p) = nil
		}
		return c, nil
	}
	// The value is decoded from a copy of its bytes, which nothing else uses: its strings may refer to it. A type
	// error of it is the one of the string, which is made after it ( see stringOptionError ).
	saved := ctx.typeError
	ctx.typeError = nil
	oldBuf, oldOrigin := ctx.Buf, ctx.origin
	ctx.Buf, ctx.origin = NewInput(bytes), nil
	next, err := d.dec.Decode(ctx, 0, depth, p)
	if err == nil && skipWhiteSpace(ctx.Buf, next) != int64(len(bytes)) {
		// the value is followed by more bytes in the string
		err = errStringOptionRest
	}
	ctx.Buf, ctx.origin = oldBuf, oldOrigin
	failed := err != nil || ctx.typeError != nil
	ctx.typeError = saved
	if failed || len(bytes) == 0 {
		if d.isMapKey {
			ctx.keyTypeError(start, c, bytes, d.elemType())
			return c, errMapKeyType
		}
		if d.isPtrType && stringOptionAllocates(bytes) {
			d.allocate(p)
		}
		return ctx.stringOptionError(start, c, bytes, d.typ)
	}
	return c, nil
}

// elemType returns the type of the value, which the pointers of the type point to.
func (d *wrappedStringDecoder) elemType() reflect.Type {
	typ := d.typ
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

// allocate sets the pointer at p, if it is nil, to a zero value, as encoding/json does for a value it fails to
// decode.
func (d *wrappedStringDecoder) allocate(p unsafe.Pointer) {
	if *(*unsafe.Pointer)(p) == nil {
		*(*unsafe.Pointer)(p) = newValue(d.typ.Elem())
	}
}

func (d *wrappedStringDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: wrapped string decoder does not support decode path")
}
