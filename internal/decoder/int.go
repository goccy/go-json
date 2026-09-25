package decoder

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type intDecoder struct {
	typ        reflect.Type
	kind       reflect.Kind
	op         func(unsafe.Pointer, int64)
	structName string
	fieldName  string
}

func newIntDecoder(typ reflect.Type, structName, fieldName string, op func(unsafe.Pointer, int64)) *intDecoder {
	return &intDecoder{
		typ:        typ,
		kind:       typ.Kind(),
		op:         op,
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *intDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		return cursor + 4, nil
	case '-':
		cursor++
	}
	if buf[cursor]-'0' > 9 {
		if cursor > start {
			return 0, errors.ErrSyntax(fmt.Sprintf("invalid character %s in numeric literal", quoteChar(buf[cursor])), cursor+1)
		}
		// a value of another kind
		return ctx.skipTypeError(cursor, depth, d.typ)
	}
	u, next, digits := parseDigits(buf, cursor)
	if isFloatContinuation(buf[next]) || digits > maxUint64Digits {
		// a number which is not an integer, or too large for any integer
		end, err := numberEnd(buf, start)
		if err != nil {
			return 0, err
		}
		ctx.numberTypeError(start, end, d.typ)
		return end, nil
	}
	neg := cursor > start
	if (neg && u > 1<<63) || (!neg && u > 1<<63-1) {
		ctx.numberTypeError(start, next, d.typ)
		return next, nil
	}
	i64 := int64(u)
	if neg {
		i64 = -i64
	}
	switch d.kind {
	case reflect.Int8:
		if i64 < -1*(1<<7) || (1<<7) <= i64 {
			ctx.numberTypeError(start, next, d.typ)
			return next, nil
		}
	case reflect.Int16:
		if i64 < -1*(1<<15) || (1<<15) <= i64 {
			ctx.numberTypeError(start, next, d.typ)
			return next, nil
		}
	case reflect.Int32:
		if i64 < -1*(1<<31) || (1<<31) <= i64 {
			ctx.numberTypeError(start, next, d.typ)
			return next, nil
		}
	}
	d.op(p, i64)
	return next, nil
}

func (d *intDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: int decoder does not support decode path")
}
