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

func (d *intDecoder) typeError(buf []byte, offset int64) *errors.UnmarshalTypeError {
	return &errors.UnmarshalTypeError{
		Value:  fmt.Sprintf("number %s", string(buf)),
		Type:   d.typ,
		Struct: d.structName,
		Field:  d.fieldName,
		Offset: offset,
	}
}

var (
	pow10i64 = [...]int64{
		1e00, 1e01, 1e02, 1e03, 1e04, 1e05, 1e06, 1e07, 1e08, 1e09,
		1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18,
	}
	pow10i64Len = len(pow10i64)
)

func (d *intDecoder) parseInt(b []byte) (int64, error) {
	isNegative := false
	if b[0] == '-' {
		b = b[1:]
		isNegative = true
	}
	maxDigit := len(b)
	if maxDigit > pow10i64Len {
		return 0, fmt.Errorf("invalid length of number")
	}
	sum := int64(0)
	for i := 0; i < maxDigit; i++ {
		c := int64(b[i]) - 48
		digitValue := pow10i64[maxDigit-i-1]
		sum += c * digitValue
	}
	if isNegative {
		return -1 * sum, nil
	}
	return sum, nil
}

var (
	numTable = [256]bool{
		'0': true,
		'1': true,
		'2': true,
		'3': true,
		'4': true,
		'5': true,
		'6': true,
		'7': true,
		'8': true,
		'9': true,
	}
)

var (
	numZeroBuf = []byte{'0'}
)

func (d *intDecoder) decodeByte(buf []byte, cursor int64) ([]byte, int64, error) {
	b := (*sliceHeader)(unsafe.Pointer(&buf)).data
	for {
		switch char(b, cursor) {
		case ' ', '\n', '\t', '\r':
			cursor++
			continue
		case '0':
			cursor++
			return numZeroBuf, cursor, nil
		case '-', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			start := cursor
			cursor++
			for numTable[char(b, cursor)] {
				cursor++
			}
			num := buf[start:cursor]
			return num, cursor, nil
		case 'n':
			if err := validateNull(buf, cursor); err != nil {
				return nil, 0, err
			}
			cursor += 4
			return nil, cursor, nil
		default:
			return nil, 0, d.typeError([]byte{char(b, cursor)}, cursor)
		}
	}
}

func (d *intDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	bytes, c, err := d.decodeByte(ctx.Buf, cursor)
	if err != nil {
		return 0, err
	}
	if bytes == nil {
		return c, nil
	}
	cursor = c

	i64, err := d.parseInt(bytes)
	if err != nil {
		return 0, d.typeError(bytes, cursor)
	}
	switch d.kind {
	case reflect.Int8:
		if i64 < -1*(1<<7) || (1<<7) <= i64 {
			return 0, d.typeError(bytes, cursor)
		}
	case reflect.Int16:
		if i64 < -1*(1<<15) || (1<<15) <= i64 {
			return 0, d.typeError(bytes, cursor)
		}
	case reflect.Int32:
		if i64 < -1*(1<<31) || (1<<31) <= i64 {
			return 0, d.typeError(bytes, cursor)
		}
	}
	d.op(p, i64)
	return cursor, nil
}

func (d *intDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: int decoder does not support decode path")
}
