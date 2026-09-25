//go:build !go1.27 || !goexperiment.jsonv2

package decoder

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

// newTypeError returns the type error as encoding/json before Go 1.27, and with GOEXPERIMENT=nojsonv2, reports
// it. Offset is after a literal and after the bracket of an array or an object, and at the first byte of a key.
// Struct is the struct type of the last field of the path, and Field the names of its fields, through their
// embedded fields since Go 1.24 ( see embeddedFieldNames ); the elements and the entries of the path have no name. An error in no field has neither.
func newTypeError(p *pendingTypeError, path []typeErrorStep, _ reflect.Type) *errors.UnmarshalTypeError {
	e := &errors.UnmarshalTypeError{Value: p.value, Type: p.typ, Offset: p.end, Err: p.err}
	if p.noContext {
		return e
	}
	switch {
	case p.atKey:
		e.Offset = p.start + 1
	case !p.literal && (p.kind == arrayValue || p.kind == objectValue):
		e.Offset = p.start + 1
	}
	var names []string
	for _, step := range path {
		if !step.isField {
			continue
		}
		e.Struct = step.structName
		if embeddedFieldNames {
			names = append(names, step.embedded...)
		}
		names = append(names, step.name)
	}
	e.Field = strings.Join(names, ".")
	return e
}

// base64Error records the error of the decoding of the string between start and end into typ, which is not
// base64: encoding/json before Go 1.27 returns the error of the decoding as it is.
func (ctx *RuntimeContext) base64Error(_, _ int64, _ reflect.Type, err error) {
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{plain: err}
	}
}

// numberKindError records the type error of the value at cursor, of another kind than a number or a string,
// which is decoded into a json.Number, and skips it.
func (ctx *RuntimeContext) numberKindError(cursor, depth int64, typ reflect.Type) (int64, error) {
	return ctx.skipTypeError(cursor, depth, typ)
}

// numberStringError returns the error of the string between start and end, which is not a number, decoded into
// a json.Number: encoding/json before Go 1.27 stops the decoding with it.
func (ctx *RuntimeContext) numberStringError(start, end int64, _ reflect.Type) (int64, error) {
	return 0, fmt.Errorf("json: invalid number literal, trying to unmarshal %q into Number", ctx.Buf[start:end])
}

// stringOptionUnquoted records the error of the value at cursor of a field of the string option, which is not in
// a string, and skips it: encoding/json before Go 1.27 reports it as an invalid use of the option.
func (ctx *RuntimeContext) stringOptionUnquoted(cursor, depth int64, typ reflect.Type) (int64, error) {
	end, err := skipValue(ctx.Buf, skipWhiteSpace(ctx.Buf, cursor), depth)
	if err != nil {
		return 0, err
	}
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{
			plain: fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal unquoted value into %v", typ),
		}
	}
	return end, nil
}

// stringOptionError records or returns the error of the string between start and end of a field of the string
// option, whose bytes are not a value of typ, as encoding/json before Go 1.27 tells it from the first byte of the
// value and the kind of typ: a number out of the range of typ is a type error, a value which starts as no value
// of its kind stops the decoding, and any other one is an invalid use of the option.
func (ctx *RuntimeContext) stringOptionError(start, end int64, value []byte, typ reflect.Type) (int64, error) {
	invalidUse := func() error {
		return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", value, typ)
	}
	save := func(err error) (int64, error) {
		if ctx.typeError == nil {
			ctx.typeError = &pendingTypeError{plain: err}
		}
		return end, nil
	}
	if len(value) == 0 {
		return save(invalidUse())
	}
	elem := typ
	for elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}
	invalidUse = func() error {
		return fmt.Errorf("json: invalid use of ,string struct tag, trying to unmarshal %q into %v", value, elem)
	}
	switch c := value[0]; {
	case c == 'n', c == 't', c == 'f':
		return save(invalidUse())
	case c == '"':
		if elem.Kind() == reflect.String {
			return 0, invalidUse()
		}
		if ctx.typeError == nil {
			ctx.typeError = &pendingTypeError{typ: elem, start: start, end: end, value: "string", kind: stringValue, literal: true}
		}
		return end, nil
	case c != '-' && c-'0' > 9:
		return 0, invalidUse()
	}
	switch elem.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		if ctx.typeError == nil {
			ctx.typeError = &pendingTypeError{
				typ: elem, start: start, end: end, value: "number " + string(value), kind: numberValue, literal: true,
			}
		}
		return end, nil
	}
	return 0, invalidUse()
}

// stringOptionUnquotedAllocates is whether a nil pointer of a field of the string option is set to a zero value
// when the value of the field is not in a string: encoding/json before Go 1.27 leaves it.
const stringOptionUnquotedAllocates = false

// StreamOffsetBase returns the position which the offsets of the type errors of a value of a stream are relative to,
// of the end of the previous value and of the start of the value: encoding/json before Go 1.27 reports them from
// the end of the previous value.
func StreamOffsetBase(prevEnd, _ int64) int64 {
	return prevEnd
}

// stringOptionAllocates reports whether a nil pointer of a field of the string option is set to a zero value when
// the value in the string of the field is not of its type: encoding/json before Go 1.27 leaves it for an empty
// string.
func stringOptionAllocates(value []byte) bool {
	return len(value) > 0
}

// textUnmarshalerKindError records the type error of the value at cursor, which is not a string, of a type whose
// pointer type typ implements encoding.TextUnmarshaler, and skips it: encoding/json before Go 1.27 reports the
// pointer type.
func (ctx *RuntimeContext) textUnmarshalerKindError(cursor, depth int64, typ reflect.Type) (int64, error) {
	return ctx.skipTypeError(cursor, depth, typ)
}

// mapKeySupported reports whether encoding/json before Go 1.27 decodes the keys of a map of keyType, which dec
// decodes: strings, integers and the types which implement encoding.TextUnmarshaler.
func mapKeySupported(keyType reflect.Type, dec Decoder) bool {
	if _, ok := dec.(*unmarshalTextDecoder); ok {
		return true
	}
	switch keyType.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

// unsupportedMapKeys records the type error of the object at cursor of a map whose keys are not decoded ( see
// mapKeySupported ), and skips it: encoding/json before Go 1.27 reports the map, which is left.
func (ctx *RuntimeContext) unsupportedMapKeys(d *mapDecoder, cursor, depth int64, _ unsafe.Pointer) (int64, error) {
	return ctx.skipTypeError(cursor, depth, d.mapType)
}
