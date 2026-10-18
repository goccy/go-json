package decoder

import (
	stderrors "errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

// The type errors are the ones of encoding/json of Go 1.27, which is made of encoding/json/v2, on every version of
// Go: a value which is not of the type of its destination is skipped, the decoding goes on, and the first such error
// is returned at the end, unless the input has a syntax error, which is returned instead. The error is recorded where
// it happens by its position only; the path to its value, which encoding/json reports as the Field of the error, is
// computed at the end by a walk of the input from the root ( see typeErrorPath ), so that the decoding of a value
// which has no error keeps no path.

// float64Type is the type which the type error of a number out of the range of a float64 decoded into an
// interface{} reports.
var float64Type = reflect.TypeOf(float64(0))

// pendingTypeError is the first type error of a decoding, as it is recorded.
type pendingTypeError struct {
	typ reflect.Type
	// start and end are the positions of the value in the buffer.
	start, end int64
	// value is the description of the value which the error reports, as "string" or "number 1.5".
	value string
	// kind is the kind of the value.
	kind jsonKind
	// atKey is set for the error of the key of a map entry, whose value starts at start too.
	atKey bool
	// literal is set for an error which reports the literal of the value in its value, as "number 1.5".
	literal bool
	// plain is an error which is not a type error, which encoding/json returns as it is: the error of an unmarshal
	// method.
	plain error
	// err is the cause of the error.
	err error
	// noContext is set for an error which is reported with no path and no offset, as encoding/json reports the ones
	// of a json.Number.
	noContext bool
	// atStart is set for an error whose offset is the start of the value, as encoding/json reports the ones of an
	// interface value which is kept.
	atStart bool
	// path is the path to the value of an error of a stream which decodes a value element by element, whose bytes
	// it drops after an element, from the value up to the root: the path is made while they are read ( see
	// Stream.keepTypeError ), and start and end are offsets in the whole input then.
	path []string
}

// jsonKind is the kind of a JSON value, which a decoder decodes and a type error reports.
type jsonKind uint8

const (
	noValue jsonKind = iota
	stringValue
	numberValue
	boolValue
	nullValue
	arrayValue
	objectValue
)

// kindOf returns the kind of the value which starts with c, or noValue if c starts no value.
func kindOf(c byte) jsonKind {
	switch c {
	case '"':
		return stringValue
	case 't', 'f':
		return boolValue
	case 'n':
		return nullValue
	case '[':
		return arrayValue
	case '{':
		return objectValue
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return numberValue
	}
	return noValue
}

// String returns the description of the kind which the type errors report.
func (k jsonKind) String() string {
	switch k {
	case stringValue:
		return "string"
	case boolValue:
		return "bool"
	case nullValue:
		return "null"
	case arrayValue:
		return "array"
	case objectValue:
		return "object"
	case numberValue:
		return "number"
	}
	return ""
}

// isOtherValue reports whether c starts a JSON value of another kind than kind, which is the kind of the values
// of a decoder: such a value is of another type, and is a type error, as a byte which starts no value is a
// syntax error.
func isOtherValue(c byte, kind jsonKind) bool {
	k := kindOf(c)
	return k != noValue && k != kind
}

// skipTypeError records the type error of the value at cursor, which is not of typ, if it is the first one, and
// returns the position after the value, which is skipped. A syntax error of the value is returned instead.
func (ctx *RuntimeContext) skipTypeError(cursor, depth int64, typ reflect.Type) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	if ctx.Option.V2 != 0 {
		// the semantics of encoding/json/v2 stop at the error
		return 0, &errors.UnmarshalError{GoType: typ, Start: cursor}
	}
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	if ctx.typeError == nil {
		kind := kindOf(buf[cursor])
		ctx.typeError = &pendingTypeError{typ: typ, start: cursor, end: end, value: kind.String(), kind: kind}
	}
	return end, nil
}

// numberTypeError records the type error of the number between start and end, which is not of typ because it
// is out of its range or not an integer: the error reports the literal of the number. It is not inlined, so that
// the decoders of the numbers keep the size they had.
//
//go:noinline
func (ctx *RuntimeContext) numberTypeError(start, end int64, typ reflect.Type) {
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{
			typ: typ, start: start, end: end, value: "number " + string(ctx.Buf[start:end]), kind: numberValue,
			literal: true,
		}
	}
}

// keyTypeError records the type error of the key of a map entry between start and end, the quotes of which are
// included, which is not a number of typ: the error reports the key.
func (ctx *RuntimeContext) keyTypeError(start, end int64, key []byte, typ reflect.Type) {
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{
			typ: typ, start: start, end: end, value: "number " + string(key), kind: stringValue, atKey: true,
			literal: true,
		}
	}
}

// HasTypeError reports whether the decoding has a type error, which TypeError returns: the type of the value
// is made for it only.
func (ctx *RuntimeContext) HasTypeError() bool {
	return ctx.typeError != nil
}

// DiscardTypeError drops the type error of a decoding which failed with another error.
func (ctx *RuntimeContext) DiscardTypeError() {
	ctx.typeError = nil
}

// TypeError returns the first type error of the decoding of the value at base into a value of typ, or nil if the
// decoding had none. Its offset is relative to offsetBase.
func (ctx *RuntimeContext) TypeError(typ reflect.Type, base, offsetBase int64) error {
	p := ctx.typeError
	if p == nil {
		return nil
	}
	ctx.typeError = nil
	if p.plain != nil {
		return p.plain
	}
	return p.toError(typeErrorPath(ctx.Buf, base, p), typ, offsetBase)
}

// jsonPointerEscaper escapes a token of a JSON pointer ( RFC 6901 ).
var jsonPointerEscaper = strings.NewReplacer("~", "~0", "/", "~1")

// toError returns the type error of p, whose value is at the end of path from the root value of typ, with its
// offset relative to offsetBase. Offset is after the value, or after the key of a map entry. Struct is the name of
// the root type, and Field the path from it to the value, whose steps are escaped as the tokens of a JSON pointer are.
// An error at the root has neither.
func (p *pendingTypeError) toError(path []string, typ reflect.Type, offsetBase int64) error {
	e := &errors.UnmarshalTypeError{Value: p.value, Type: p.typ, Err: p.err}
	if p.noContext {
		return e
	}
	switch {
	case p.atStart:
		e.Offset = p.start
	case !p.literal && (p.kind == arrayValue || p.kind == objectValue):
		e.Offset = p.start + 1
	default:
		e.Offset = p.end
	}
	e.Offset -= offsetBase
	if len(path) > 0 {
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		e.Struct = typ.Name()
		for i, name := range path {
			path[i] = jsonPointerEscaper.Replace(name)
		}
		e.Field = strings.Join(path, ".")
	}
	return e
}

// typeErrorPath returns the path from the value at cursor to the value of the error: the keys of the members and
// the indices of the elements of the input which the value is in, whatever the values they are decoded into. The input
// is read once up to the value of the error, keeping the objects and the arrays it is in, so that the walk is linear in
// the bytes before the value, whatever its depth.
func typeErrorPath(buf []byte, cursor int64, p *pendingTypeError) []string {
	var steps []typeErrorStep
	for {
		// a value, which has the error if it starts where the error does
		cursor = skipWhiteSpace(buf, cursor)
		if cursor >= p.start {
			return typeErrorNames(buf, steps)
		}
		switch buf[cursor] {
		case '{':
			cursor = skipWhiteSpace(buf, cursor+1)
			if buf[cursor] != '}' {
				steps = append(steps, typeErrorStep{index: -1})
				var ok bool
				if cursor, ok = typeErrorMember(buf, cursor, p, &steps[len(steps)-1]); !ok {
					return typeErrorNames(buf, steps)
				}
				continue
			}
			cursor++
		case '[':
			cursor = skipWhiteSpace(buf, cursor+1)
			if buf[cursor] != ']' {
				steps = append(steps, typeErrorStep{})
				continue
			}
			cursor++
		default:
			end, err := skipValue(buf, cursor, 0)
			if err != nil {
				return nil
			}
			cursor = end
		}
		// after the value: the next member or element of the innermost object or array which goes on
		for {
			if len(steps) == 0 {
				return nil
			}
			cursor = skipWhiteSpace(buf, cursor)
			step := &steps[len(steps)-1]
			if c := buf[cursor]; c == '}' || c == ']' {
				steps = steps[:len(steps)-1]
				cursor++
				continue
			} else if c != ',' {
				return nil
			}
			if step.index >= 0 {
				step.index++
				cursor++
				break
			}
			var ok bool
			if cursor, ok = typeErrorMember(buf, cursor+1, p, step); !ok {
				return typeErrorNames(buf, steps)
			}
			break
		}
	}
}

// typeErrorStep is a member of an object, by the position of its key, or an element of an array, by its index,
// which the value of a type error is in.
type typeErrorStep struct {
	keyAt int64
	// index is the index of an element, or -1 for a member.
	index int
}

// typeErrorMember reads the key of the member of step at cursor, and returns the position of its value, or false if
// the error is of the key.
func typeErrorMember(buf []byte, cursor int64, p *pendingTypeError, step *typeErrorStep) (int64, bool) {
	cursor = skipWhiteSpace(buf, cursor)
	step.keyAt = cursor
	if p.atKey && cursor == p.start {
		return 0, false
	}
	end, err := skipValue(buf, cursor, 0)
	if err != nil {
		return 0, false
	}
	// the colon after the key
	return skipWhiteSpace(buf, end) + 1, true
}

// typeErrorNames returns the names of the steps: the keys of the members, decoded, and the indices of the elements.
func typeErrorNames(buf []byte, steps []typeErrorStep) []string {
	if len(steps) == 0 {
		return nil
	}
	names := make([]string, len(steps))
	for i, step := range steps {
		if step.index >= 0 {
			names[i] = strconv.Itoa(step.index)
			continue
		}
		raw, next, info, err := skipStringDecoder.scanString(buf, step.keyAt)
		if next < 0 {
			raw, _, info, err = skipStringDecoder.scanStringRest(buf, raw, -next-1, info)
		}
		if err == nil {
			names[i] = string(decodeLiteral(raw, info))
		}
	}
	return names
}

// base64Error records the error of the decoding of the string between start and end into typ, which is not
// base64: it is a type error of the string, whose cause is the error.
func (ctx *RuntimeContext) base64Error(start, end int64, typ reflect.Type, err error) {
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{typ: typ, start: start, end: end, value: "string", kind: stringValue, err: err}
	}
}

// numberKindError records the type error of the value at cursor, of another kind than a number or a string,
// which is decoded into a json.Number, and skips it: encoding/json reports it with no path and no offset.
func (ctx *RuntimeContext) numberKindError(cursor, depth int64, typ reflect.Type) (int64, error) {
	first := ctx.typeError == nil
	next, err := ctx.skipTypeError(cursor, depth, typ)
	if err == nil && first {
		ctx.typeError.noContext = true
	}
	return next, err
}

// numberStringError records the type error of the string between start and end, which is not a number, decoded
// into a json.Number: encoding/json reports the string, with no path and no offset.
func (ctx *RuntimeContext) numberStringError(start, end int64, typ reflect.Type) (int64, error) {
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{
			typ: typ, start: start, end: end, value: "string " + string(ctx.Buf[start:end]), kind: stringValue,
			literal: true, err: strconv.ErrSyntax, noContext: true,
		}
	}
	return end, nil
}

// stringOptionUnquoted records the type error of the value at cursor of a field of the string option, which is not
// in a string, and skips it: encoding/json reports the value as a value of the type of the field.
func (ctx *RuntimeContext) stringOptionUnquoted(cursor, depth int64, typ reflect.Type) (int64, error) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == jsonNumberType {
		return ctx.numberKindError(cursor, depth, typ)
	}
	return ctx.skipTypeError(cursor, depth, typ)
}

// methodError records err, which an unmarshal method returned, if it is the first error: the decoding goes on after
// it, as after a type error, and it is returned as it is.
func (ctx *RuntimeContext) methodError(err error) {
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{plain: err}
	}
}

// stringOptionError records the type error of the string between start and end of a field of the string option,
// whose bytes are not a value of typ: the error reports the bytes as a number for a number type, and else the
// string, whose cause is a syntax error.
func (ctx *RuntimeContext) stringOptionError(start, end int64, value []byte, typ reflect.Type) (int64, error) {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == jsonNumberType {
		return ctx.numberStringError(start, end, typ)
	}
	if ctx.typeError != nil {
		return end, nil
	}
	p := &pendingTypeError{typ: typ, start: start, end: end, kind: stringValue, literal: true}
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		p.value = "number " + string(value)
	case reflect.String:
		// the bytes are decoded as a JSON string, whose syntax error is the cause
		p.value = "string"
		p.err = stringOptionStringError(value)
	default:
		p.value = "string " + string(ctx.Buf[start:end])
		p.err = strconv.ErrSyntax
	}
	ctx.typeError = p
	return end, nil
}

// errTextUnmarshalerKind is the cause of the type error of a value which is not a string, of a type which
// implements encoding.TextUnmarshaler.
var errTextUnmarshalerKind = stderrors.New("JSON value must be string type")

// textUnmarshalerKindError records the type error of the value at cursor, which is not a string, of a type whose
// pointer type typ implements encoding.TextUnmarshaler, and skips it: the error reports the type of the value, and
// the cause of the error.
func (ctx *RuntimeContext) textUnmarshalerKindError(cursor, depth int64, typ reflect.Type) (int64, error) {
	first := ctx.typeError == nil
	next, err := ctx.skipTypeError(cursor, depth, typ.Elem())
	if err == nil && first {
		ctx.typeError.err = errTextUnmarshalerKind
	}
	return next, err
}

// errUnsettableEmbeddedPointer is the error of a field promoted from an embedded pointer to an unexported struct,
// which can't be set ( see structFieldSet.err ).
var errUnsettableEmbeddedPointer = stderrors.New("cannot set embedded pointer to unexported struct type")

// unsettableValueError records the type error of a field of the struct typ which can't be set, whose cause is err,
// for its value of the kind between start and end: it is a type error of the struct at the start of the value.
func (ctx *RuntimeContext) unsettableValueError(start, end int64, kind jsonKind, typ reflect.Type, err error) {
	if ctx.typeError == nil {
		ctx.typeError = &pendingTypeError{
			typ: typ, start: start, end: end, value: kind.String(), kind: kind, err: err, atStart: true,
		}
	}
}

// mapKeySupported reports whether the keys of a map of keyType, which dec decodes, are decoded: any key but a bool
// without a method of unmarshaling text, or a pointer to one, and a key of a type which no key is decoded into.
func mapKeySupported(keyType reflect.Type, dec Decoder) bool {
	if _, ok := dec.(*invalidDecoder); ok {
		return false
	}
	for keyType.Kind() == reflect.Pointer {
		keyType = keyType.Elem()
	}
	return keyType.Kind() != reflect.Bool || reflect.PointerTo(keyType).Implements(unmarshalTextType)
}

// unsupportedMapKeys records the type error of the first key of the object at cursor of a map whose keys are not
// decoded ( see mapKeySupported ), and skips the object: the error reports the key as a string of the type of the
// keys, and the map is made.
func (ctx *RuntimeContext) unsupportedMapKeys(d *mapDecoder, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	if *(*unsafe.Pointer)(p) == nil {
		*(*unsafe.Pointer)(p) = reflect.MakeMapWithSize(d.mapType, 0).UnsafePointer()
	}
	cursor = skipWhiteSpace(buf, cursor+1)
	if buf[cursor] == '}' {
		return cursor + 1, nil
	}
	// the keys and the values are in the object, a level deeper than it
	depth++
	for {
		if buf[cursor] != '"' {
			return 0, errors.ErrInvalidBeginningOfValue(buf[cursor], cursor)
		}
		keyStart := cursor
		keyEnd, err := skipValue(buf, cursor, depth)
		if err != nil {
			return 0, err
		}
		if ctx.typeError == nil {
			ctx.typeError = &pendingTypeError{
				typ: d.keyType, start: keyStart, end: keyEnd, value: "string", kind: stringValue, atKey: true,
				literal: true,
			}
		}
		cursor = skipWhiteSpace(buf, keyEnd)
		if buf[cursor] != ':' {
			return 0, errors.ErrExpected("colon after object key", cursor)
		}
		end, err := skipValue(buf, skipWhiteSpace(buf, cursor+1), depth)
		if err != nil {
			return 0, err
		}
		cursor = skipWhiteSpace(buf, end)
		switch buf[cursor] {
		case ',':
			cursor = skipWhiteSpace(buf, cursor+1)
		case '}':
			return cursor + 1, nil
		default:
			return 0, errors.ErrExpected("comma after object value", cursor)
		}
	}
}

// stringOptionStringError returns the syntax error of the bytes of a string of the string option of a string,
// which are not a JSON string and nothing else, as encoding/json reports it.
func stringOptionStringError(value []byte) error {
	if len(value) == 0 {
		return errors.ErrSyntax("unexpected end of JSON input", 0)
	}
	if value[0] != '"' {
		return errors.ErrSyntax(fmt.Sprintf("invalid character %s looking for beginning of object key string", quoteChar(value[0])), 0)
	}
	i := 1
	for {
		if i >= len(value) {
			return errors.ErrSyntax("unexpected end of JSON input", 0)
		}
		c := value[i]
		if c == '"' {
			i++
			break
		}
		if c != '\\' {
			i++
			continue
		}
		if i+1 >= len(value) {
			return errors.ErrSyntax("unexpected end of JSON input", 0)
		}
		switch value[i+1] {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			i += 2
			continue
		case 'u':
			if i+6 <= len(value) && isHex(value[i+2:i+6]) {
				i += 6
				continue
			}
			return errors.ErrSyntax(fmt.Sprintf("invalid escape sequence `%s` in string", value[i:min(i+6, len(value))]), 0)
		}
		return errors.ErrSyntax(fmt.Sprintf("invalid escape sequence `%s` in string", value[i:i+2]), 0)
	}
	if i < len(value) {
		return errors.ErrSyntax(fmt.Sprintf("invalid character %s after string value", quoteChar(value[i])), 0)
	}
	return nil
}

// isHex reports whether every byte of b is a hexadecimal digit.
func isHex(b []byte) bool {
	for _, c := range b {
		if !('0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F') {
			return false
		}
	}
	return true
}
