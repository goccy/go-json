package decoder

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type uintDecoder struct {
	typ        reflect.Type
	kind       reflect.Kind
	op         func(unsafe.Pointer, uint64)
	structName string
	fieldName  string
}

func newUintDecoder(typ reflect.Type, structName, fieldName string, op func(unsafe.Pointer, uint64)) *uintDecoder {
	return &uintDecoder{
		typ:        typ,
		kind:       typ.Kind(),
		op:         op,
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *uintDecoder) typeError(buf []byte, offset int64) *errors.UnmarshalTypeError {
	return &errors.UnmarshalTypeError{
		Value:  fmt.Sprintf("number %s", string(buf)),
		Type:   d.typ,
		Offset: offset,
	}
}

func (d *uintDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == 'n' {
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		return cursor + 4, nil
	}
	if buf[cursor]-'0' > 9 {
		return 0, d.typeError([]byte{buf[cursor]}, cursor)
	}
	start := cursor
	u, next, digits := parseDigits(buf, cursor)
	if isFloatContinuation(buf[next]) || digits > maxUint64Digits+1 {
		// a number which is not an integer, or too large for any integer
		end := next
		for floatTable[buf[end]] {
			end++
		}
		return 0, d.typeError(buf[start:end], end)
	}
	if digits == maxUint64Digits+1 {
		// 20 digits: the value of the first 19 is exact, and the last one fits if the whole is less than 1<<64.
		var hi uint64
		for i := int64(0); i < maxUint64Digits; i++ {
			hi = hi*10 + uint64(buf[start+i]-'0')
		}
		lo := uint64(buf[start+maxUint64Digits] - '0')
		if hi > (1<<64-1)/10 || (hi == (1<<64-1)/10 && lo > (1<<64-1)%10) {
			return 0, d.typeError(buf[start:next], next)
		}
		u = hi*10 + lo
	}
	switch d.kind {
	case reflect.Uint8:
		if (1 << 8) <= u {
			return 0, d.typeError(buf[start:next], next)
		}
	case reflect.Uint16:
		if (1 << 16) <= u {
			return 0, d.typeError(buf[start:next], next)
		}
	case reflect.Uint32:
		if (1 << 32) <= u {
			return 0, d.typeError(buf[start:next], next)
		}
	}
	d.op(p, u)
	return next, nil
}

func (d *uintDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: uint decoder does not support decode path")
}
