package decoder

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type boolDecoder struct {
	structName string
	fieldName  string
	// typ is the type of the value, which the type errors report.
	typ reflect.Type
}

func newBoolDecoder(structName, fieldName string) *boolDecoder {
	return &boolDecoder{structName: structName, fieldName: fieldName, typ: reflect.TypeOf(false)}
}

func (d *boolDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	switch buf[cursor] {
	case 't':
		if err := validateTrue(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 4
		**(**bool)(unsafe.Pointer(&p)) = true
		return cursor, nil
	case 'f':
		if err := validateFalse(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 5
		**(**bool)(unsafe.Pointer(&p)) = false
		return cursor, nil
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 4
		return cursor, nil
	}
	if isOtherValue(buf[cursor], boolValue) {
		return ctx.skipTypeError(cursor, depth, d.typ)
	}
	return 0, errors.ErrUnexpectedEndOfJSON("bool", cursor)
}

func (d *boolDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: bool decoder does not support decode path")
}
