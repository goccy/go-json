package decoder

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

// The methods of unmarshaling of the semantics of encoding/json/v2, and the time types. A value is always
// addressable for them: the method of the pointer to a type is called for a value of the type, at its address.
// The first of them which a type has decodes its value: UnmarshalJSONFrom, UnmarshalJSON, UnmarshalText. A
// time.Time and a time.Duration are decoded as their own, before their methods, and a function of WithUnmarshalers
// which takes the type comes before all of them.

// The methods of unmarshaling, in the order of their precedence.
const (
	noMethod = iota
	methodUnmarshalJSONFrom
	methodUnmarshalJSON
	methodUnmarshalText
)

// v2MethodOf returns the method of unmarshaling which decodes a value of the type, or noMethod. A pointer and an
// interface value have none: the value they point to or hold has, which is addressable.
func v2MethodOf(typ reflect.Type) int {
	if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Interface {
		return noMethod
	}
	ptr := reflect.PointerTo(typ)
	if _, ok := ptr.MethodByName("UnmarshalJSONFrom"); ok {
		return methodUnmarshalJSONFrom
	}
	switch {
	case ptr.Implements(unmarshalJSONType):
		return methodUnmarshalJSON
	case ptr.Implements(unmarshalTextType):
		return methodUnmarshalText
	}
	return noMethod
}

// ReadValue reads the value at cursor whole, as the method of unmarshaling, or the function, of a value of typ is
// given it: it must have the text which v2 allows ( see skipValueV2 ). It returns the start and the end of the
// value. An error of the text is one of the value, at its start, which the v2 json package reads whole to find it
// again, as encoding/json/v2 reads it.
func (ctx *RuntimeContext) ReadValue(cursor, depth int64, typ reflect.Type, method errors.UnmarshalMethod) (int64, int64, error) {
	start := skipWhiteSpace(ctx.Buf, cursor)
	end, err := ctx.skipValueV2(start, depth)
	if err != nil {
		return 0, 0, &errors.UnmarshalError{GoType: typ, Start: start, Err: err, Method: method}
	}
	return start, end, nil
}

// v2UnmarshalJSONDecoder is the decoder of a value whose UnmarshalJSON decodes it, which is given the value as the
// input has it, null included. The bytes are a part of the input, which the method must copy to keep them, as the
// interface of encoding/json/v2 says.
type v2UnmarshalJSONDecoder struct {
	typ     reflect.Type
	ptrType unsafe.Pointer
}

func (d *v2UnmarshalJSONDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	start, end, err := ctx.ReadValue(cursor, depth, d.typ, errors.UnmarshalJSONMethod)
	if err != nil {
		return 0, err
	}
	v := *(*any)(unsafe.Pointer(&emptyInterface{typ: d.ptrType, ptr: p}))
	if err := v.(json.Unmarshaler).UnmarshalJSON(ctx.Buf[start:end:end]); err != nil {
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: start, Err: err, Method: errors.UnmarshalJSONMethod}
	}
	return end, nil
}

func (d *v2UnmarshalJSONDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// errNonStringValue is the error of a value of a type whose UnmarshalText decodes it, which is not a string.
var errNonStringValue = fmt.Errorf("JSON value must be string type")

// v2UnmarshalTextDecoder is the decoder of a value whose UnmarshalText decodes it, from a string: null is the zero
// value, which the method is not called for, and any other value is an error.
type v2UnmarshalTextDecoder struct {
	typ     reflect.Type
	ptrType unsafe.Pointer
	zero    zeroer
}

func (d *v2UnmarshalTextDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	start, end, err := ctx.ReadValue(cursor, depth, d.typ, errors.UnmarshalTextMethod)
	if err != nil {
		return 0, err
	}
	switch ctx.Buf[start] {
	case 'n':
		d.zero.zero(p)
		return end, nil
	case '"':
	default:
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: start, Err: errNonStringValue, Method: errors.UnmarshalTextMethod}
	}
	v := *(*any)(unsafe.Pointer(&emptyInterface{typ: d.ptrType, ptr: p}))
	if err := v.(encoding.TextUnmarshaler).UnmarshalText(unquotedName(ctx.Buf[start:end])); err != nil {
		return 0, &errors.UnmarshalError{GoType: d.typ, Start: start, Err: err, Method: errors.UnmarshalTextMethod}
	}
	return end, nil
}

func (d *v2UnmarshalTextDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// compileMethodV2 returns the decoder of a value of typ by its method of unmarshaling, or nil if it has none.
func compileMethodV2(typ reflect.Type) Decoder {
	switch v2MethodOf(typ) {
	case methodUnmarshalJSONFrom:
		return &v2InvalidDecoder{typ: typ, err: errV2NotSupported}
	case methodUnmarshalJSON:
		return &v2UnmarshalJSONDecoder{typ: typ, ptrType: runtime.TypePtr(reflect.PointerTo(typ))}
	case methodUnmarshalText:
		return &v2UnmarshalTextDecoder{typ: typ, ptrType: runtime.TypePtr(reflect.PointerTo(typ)), zero: newZeroer(typ)}
	}
	return nil
}

// v2TimeDecoder is the decoder of a time.Time, from a string of RFC 3339, which time.Time.UnmarshalText reads, but
// what RFC 3339 doesn't allow and the time package does: an hour of one digit, a comma before the fraction of a
// second, and a time zone out of range. null is the zero time.
type v2TimeDecoder struct{}

func (d v2TimeDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	start, end, err := ctx.ReadValue(cursor, depth, timeType, errors.UnmarshalTextMethod)
	if err != nil {
		return 0, err
	}
	switch ctx.Buf[start] {
	case 'n':
		*(*time.Time)(p) = time.Time{}
		return end, nil
	case '"':
	default:
		return 0, &errors.UnmarshalError{GoType: timeType, Start: start, Method: errors.UnmarshalTextMethod}
	}
	text := unquotedName(ctx.Buf[start:end])
	var t time.Time
	if err := t.UnmarshalText(text); err != nil {
		return 0, &errors.UnmarshalError{GoType: timeType, Start: start, Err: err, Method: errors.UnmarshalTextMethod}
	}
	if err := strictRFC3339Error(text); err != nil {
		return 0, &errors.UnmarshalError{GoType: timeType, Start: start, Err: err, Method: errors.UnmarshalTextMethod}
	}
	*(*time.Time)(p) = t
	return end, nil
}

func (d v2TimeDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: v2 decoder does not support decode path")
}

// strictRFC3339Error returns the error of b, a time which time.Time.UnmarshalText read, which RFC 3339 doesn't allow,
// as encoding/json/v2 reports it, or nil.
func strictRFC3339Error(b []byte) error {
	parseError := func(valueElem, layoutElem, message string) error {
		return &time.ParseError{Layout: time.RFC3339, Value: string(b), LayoutElem: layoutElem, ValueElem: valueElem, Message: message}
	}
	switch {
	case b[len("2006-01-02T")+1] == ':':
		// the hour must have two digits
		return parseError(string(b[len("2006-01-02T"):][:1]), "15", "")
	case b[len("2006-01-02T15:04:05")] == ',':
		// the fraction of a second must follow a period
		return parseError(",", ".", "")
	case b[len(b)-1] != 'Z':
		// the zone is a string of the error only, as a time of a zone has no error but these
		zone := b[len(b)-len("Z07:00"):]
		switch {
		case dec2(zone[len("Z"):]) >= 24:
			return parseError(string(zone), "Z07:00", ": timezone hour out of range")
		case dec2(zone[len("Z07:"):]) >= 60:
			return parseError(string(zone), "Z07:00", ": timezone minute out of range")
		}
	}
	return nil
}

// dec2 returns the number of the two decimal digits at the start of b, or a number over 99 if they are not digits.
func dec2(b []byte) int {
	if len(b) < 2 || b[0]-'0' > 9 || b[1]-'0' > 9 {
		return 100
	}
	return int(b[0]-'0')*10 + int(b[1]-'0')
}
