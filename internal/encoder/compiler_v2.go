package encoder

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"time"
	"unsafe"

	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonfields"
	"github.com/goccy/go-json/internal/runtime"
)

// The semantics of encoding/json/v2, which the compiler gives the opcodes of a type in the v2 mode: they are the
// ones of v1 but where v2 differs.

var (
	timeDurationType = reflect.TypeFor[time.Duration]()
	timeTimeType     = reflect.TypeFor[time.Time]()
)

var (
	errYearRange     = errors.New("year outside of range [0,9999]")
	errTimezoneRange = errors.New("timezone hour outside of range [0,23]")
	// errNoDuration is the error of a time.Duration, which has no representation of its own until
	// encoding/json/v2 decides one.
	errNoDuration = errors.New("no default representation")
)

// appendTime appends the time.Time at p in RFC 3339 with nanosecond precision, which can't represent a year
// before 0 or after 9999, nor a time zone of 24 hours or more.
func appendTime(_ *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
	t := (*time.Time)(p)
	if year := t.Year(); year < 0 || year > 9999 {
		return b, &ierrors.SemanticError{GoType: timeTimeType, Err: errYearRange}
	}
	if _, offset := t.Zone(); offset <= -24*60*60 || offset >= 24*60*60 {
		return b, &ierrors.SemanticError{GoType: timeTimeType, Err: errTimezoneRange}
	}
	b = append(b, '"')
	b = t.AppendFormat(b, time.RFC3339Nano)
	return append(b, '"'), nil
}

// v2Code returns the code of a value of typ by the semantics of encoding/json/v2, and true, if they differ from
// the ones of v1 for the type: false lets the type be compiled as v1 compiles it. A type with a method of a
// marshaler is not given to it.
func (c *Compiler) v2Code(typ reflect.Type) (Code, bool) {
	// the type of a default representation, which the functions and the method c.v2DefaultAfter declined, is
	// written by what is after them, once: a value of the type in it is written as any other ( see
	// appendDefault ).
	after, isDefault := noMethod, typ == c.v2DefaultType
	switch {
	case c.nextIsEmbedded:
		// the fields of an embedded struct are members of the struct which embeds it, which no function or
		// method writes.
		return nil, false
	case c.stringTag:
		// the value of the option, whose code is the one of a field of the option
		if code := c.stringTagCode(typ); code != nil {
			return code, true
		}
	}
	switch {
	case isDefault:
		after = c.v2DefaultAfter
		c.v2DefaultType = nil
	default:
		if code, ok := c.v2FuncsCode(typ); ok {
			return code, true
		}
	}
	switch typ {
	case timeDurationType:
		return c.appendFuncCode(typ, appendError(typ, errNoDuration)), true
	case timeTimeType:
		return c.appendFuncCode(typ, appendTime), true
	}
	if code, ok := c.v2MethodCode(typ, after); ok {
		return code, true
	}
	switch typ.Kind() {
	case reflect.Struct:
		if fields := v2FieldsOf(typ); fields.Err != nil {
			return c.appendFuncCode(typ, appendError(fields.Err.GoType, fields.Err.Err)), true
		} else if fields.FormatErr != nil {
			return c.appendFuncCode(typ, appendError(fields.FormatErr.GoType, fields.FormatErr.Err)), true
		}
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		// a value of a type which has no representation fails when it is marshaled, not when its type is
		// compiled: an empty map of it is {}.
		return c.appendFuncCode(typ, appendError(typ, nil)), true
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 && typ.Elem().PkgPath() != "" {
			// a slice of a named byte type is a list of its values, of their own representation: only []byte is
			// base64.
			code, err := c.sliceCode(typ)
			if err != nil {
				return nil, false
			}
			return code, true
		}
	case reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 && typ.Elem().PkgPath() == "" {
			n := typ.Len()
			return c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
				return AppendByteSlice(ctx, b, unsafe.Slice((*byte)(p), n)), nil
			}), true
		}
	case reflect.Ptr:
		if kind := typ.Elem().Kind(); kind == reflect.Map || kind == reflect.Slice {
			// A nil map or slice is {} or [], and a nil pointer is null: the opcodes of a map or a slice which a
			// pointer is followed to can't tell the one from the other, so the value is encoded by the opcodes
			// of its type, as a value of interface{} is, after the pointer.
			return &PtrCode{typ: typ, value: &InterfaceCode{typ: typ.Elem(), static: true}, ptrNum: 1}, true
		}
	}
	return nil, false
}

// appendFuncCode returns the code which writes a value of typ by fn, which is given the address of the value, as
// a method on the pointer is.
func (c *Compiler) appendFuncCode(typ reflect.Type, fn AppendFunc) *MarshalJSONCode {
	return &MarshalJSONCode{typ: typ, appendValue: fn, isAddrForMarshaler: true}
}

// appendError returns the function which fails for a value of typ with err.
func appendError(typ reflect.Type, err error) AppendFunc {
	return func(_ *RuntimeContext, b []byte, _ unsafe.Pointer) ([]byte, error) {
		return b, &ierrors.SemanticError{GoType: typ, Err: err}
	}
}

// v2Fields are the JSON object members of the struct types, which v2FieldsOf found.
var v2Fields sync.Map // map[reflect.Type]*jsonfields.Fields

// v2FieldsOf returns the JSON object members of the struct type t.
func v2FieldsOf(t reflect.Type) *jsonfields.Fields {
	if fields, ok := v2Fields.Load(t); ok {
		return fields.(*jsonfields.Fields)
	}
	fields, _ := v2Fields.LoadOrStore(t, jsonfields.Of(t))
	return fields.(*jsonfields.Fields)
}

// v2Object is the JSON object of a struct of the v2 semantics, while the struct and the structs embedded in it are
// compiled.
type v2Object struct {
	// members are the members of the object, by the key of their indexes ( see indexKey ).
	members map[string]*jsonfields.Field
	// embeds are the embedded structs which a member is in, by the key of their indexes.
	embeds map[string]bool
	// path is the index of the struct being compiled in the struct of the object.
	path []int
}

// indexKey is the key of the index of a field by reflect.Type.FieldByIndex.
func indexKey(index []int) string {
	return fmt.Sprint(index)
}

func newV2Object(fields *jsonfields.Fields) *v2Object {
	obj := &v2Object{members: map[string]*jsonfields.Field{}, embeds: map[string]bool{}}
	for i := range fields.List {
		f := &fields.List[i]
		obj.members[indexKey(f.Index)] = f
		for n := 1; n < len(f.Index); n++ {
			obj.embeds[indexKey(f.Index[:n])] = true
		}
	}
	return obj
}

// v2StructTags returns the tags of the fields of the struct type typ which are written to the JSON object of the
// v2 semantics, in their order: the members, and the embedded structs which members are in, whose indexes are
// in embeds. An embedded struct is a part of the object being compiled, at v2EmbedPath; another struct is the
// object of its own. The returned function ends the compile of the struct.
func (c *Compiler) v2StructTags(typ reflect.Type, embedded bool) (runtime.StructTags, map[*runtime.StructTag][]int, func()) {
	saved := c.v2Object
	var obj *v2Object
	if embedded && saved != nil {
		obj = &v2Object{members: saved.members, embeds: saved.embeds, path: c.v2EmbedPath}
	} else {
		obj = newV2Object(v2FieldsOf(typ))
	}
	c.v2Object = obj
	var tags runtime.StructTags
	embeds := map[*runtime.StructTag][]int{}
	for i := range typ.NumField() {
		path := append(slices.Clip(obj.path), i)
		key := indexKey(path)
		sf := typ.Field(i)
		if m, ok := obj.members[key]; ok {
			tags = append(tags, &runtime.StructTag{
				Key:         m.Name,
				QuotedKey:   m.QuotedName,
				IsTaggedKey: m.HasName,
				IsOmitEmpty: m.OmitEmpty,
				IsOmitZero:  m.OmitZero || c.omitZeroStructFields,
				IsString:    m.String,
				Field:       sf,
			})
		} else if obj.embeds[key] {
			tag := &runtime.StructTag{Key: sf.Name, Field: sf}
			tags = append(tags, tag)
			embeds[tag] = path
		}
	}
	return tags, embeds, func() { c.v2Object = saved }
}

// v2HiddenNames returns the names of the members of typ, a recursive struct embedded at path in the object being
// compiled, which the object doesn't write from it: its code is compiled when the recursive codes are linked, as
// the object of its own ( see StructCode.hiddenNames ).
func (c *Compiler) v2HiddenNames(typ reflect.Type, path []int) []string {
	written := map[string]bool{}
	for _, m := range c.v2Object.members {
		if len(m.Index) > len(path) && slices.Equal(m.Index[:len(path)], path) {
			written[m.Name] = true
		}
	}
	var hidden []string
	for _, m := range v2FieldsOf(typ).List {
		if !written[m.Name] {
			hidden = append(hidden, m.Name)
		}
	}
	return hidden
}

var errInvalidStringTag = errors.New("invalid use of `string` tag option")

// v2FieldValueCode returns the code of the value of a field of the v2 semantics, and true, if the options of its
// tag make it differ from the code of its type: a value which the `string` option doesn't apply to fails. The
// options of the field are changed for its opcodes.
func (c *Compiler) v2FieldValueCode(field *StructFieldCode) (Code, bool) {
	tag := *field.tag
	field.tag = &tag
	if tag.IsOmitEmpty {
		// the value is omitted if it would be null, "", {} or []: some kinds are known to be empty or not by the
		// value ( the check of omitempty of v1 ), the others by what they write ( unwriteEmpty ).
		switch kind := field.typ.Kind(); {
		case kind == reflect.Ptr && !c.v2HasMethod(field.typ) && c.v2NeverEmpty(field.typ.Elem()):
			// a nil pointer is null, and the value it points to is never empty.
		case kind == reflect.Ptr && !c.v2HasMethod(field.typ) && !c.v2HasMethod(field.typ.Elem()) &&
			field.typ.Elem().Kind() == reflect.String && !tag.IsString:
			// a nil pointer is null, and the string it points to is empty only if it is "".
			field.omitEmptyString = true
		case (kind == reflect.Ptr || kind == reflect.Interface) && len(c.funcs.funcsOf(field.typ)) == 0:
			// a nil value is null, unless a function writes it, as a function of the pointer to it does: with
			// functions, it is omitted by what is written, as a value of a method is.
			field.unwriteEmpty = true
		case field.typ == rawMessageType && v2MethodOf(field.typ, noMethod) == methodMarshalJSON &&
			len(c.funcs.funcsOf(field.typ)) == 0:
			// a nil json.RawMessage is null, which its method writes without a call ( see v2MethodCode ): it is
			// omitted by the check of nil, the others by what they write.
			field.emptyNil = true
			field.unwriteEmpty = true
		case c.v2HasMethod(field.typ):
			tag.IsOmitEmpty = false
			field.unwriteEmpty = true
		case c.v2NeverEmpty(field.typ):
			tag.IsOmitEmpty = false
		case kind == reflect.String || kind == reflect.Map || kind == reflect.Slice || kind == reflect.Array:
			// empty if its length is 0.
		default:
			tag.IsOmitEmpty = false
			field.unwriteEmpty = true
		}
	}
	// the check of nil writes null instead of calling a marshaler, which only a nil pointer is.
	field.isNilCheck = tag.IsOmitEmpty || field.typ.Kind() == reflect.Ptr
	if !tag.IsString {
		return nil, false
	}
	code := c.stringTagCode(field.typ)
	if code == nil {
		return nil, false // a number, which is written in a string
	}
	tag.IsString = false
	return code, true
}

// stringTagCode returns the code of a value of typ of the `string` option, as encoding/json/v2 writes it: nil for a
// number, after the pointers to it, whose nil is null, which is written in a string, and for a value which a method
// or a function writes in the compile mode of the option, whose options they are given ( see appendStringTag ); the
// error of the option for a value of another kind.
func (c *Compiler) stringTagCode(typ reflect.Type) Code {
	elem := typ
	ptrNum := 0
	for elem.Kind() == reflect.Ptr && !c.v2HasMethod(elem) {
		elem = elem.Elem()
		ptrNum++
	}
	var code Code
	switch {
	case c.v2HasMethod(elem) && elem != timeTimeType:
		if c.stringTag {
			return nil
		}
		code = c.appendFuncCode(elem, appendStringTag(elem))
	case isNumberKind(elem.Kind()) && elem != timeDurationType || elem == jsonNumberType:
		return nil
	case elem == timeDurationType:
		code = c.appendFuncCode(elem, appendError(elem, errNoDuration))
	default:
		code = c.appendFuncCode(elem, appendError(elem, errInvalidStringTag))
	}
	if ptrNum > 0 {
		code = &PtrCode{typ: typ, value: code, ptrNum: uint8(ptrNum)}
	}
	return code
}

// appendStringTag returns the function which appends a value of typ, which a method or a function writes, in the
// compile mode of the `string` option, whose options they are given, as the value written after them is.
func appendStringTag(typ reflect.Type) AppendFunc {
	return func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
		flags := ctx.Option.Flag
		ctx.Option.Flag |= StringTagOption | StringifyNumbersOption
		out, err := appendDefault(ctx, b, typ, p, allMethods)
		ctx.Option.Flag = flags
		return out, err
	}
}

// v2NeverEmpty reports whether a value of typ is known to be written as none of the empty values of omitempty:
// a boolean, a number, or an object with a member which is always written. The value of a type which a method or
// a function writes may be any.
func (c *Compiler) v2NeverEmpty(typ reflect.Type) bool {
	if c.v2HasMethod(typ) {
		// MarshalJSON of time.Time, or of a type which has it by embedding a time.Time, writes a time.
		return typ != c.v2DefaultType && len(c.funcs.funcsOf(typ)) == 0 &&
			v2MethodOf(typ, noMethod) == methodMarshalJSON && addrJSONAppender(typ) != nil
	}
	switch kind := typ.Kind(); {
	case kind == reflect.Bool || isNumberKind(kind), typ == jsonNumberType:
		// a json.Number is a number, 0 if it is "".
		return true
	case kind == reflect.Struct && !c.omitZeroStructFields:
		fields := v2FieldsOf(typ)
		if fields.Err != nil {
			return false
		}
		for _, f := range fields.List {
			// a member of an embedded struct is not written if a pointer to the struct is nil.
			if !f.OmitEmpty && !f.OmitZero && len(f.Index) == 1 {
				return true
			}
		}
	}
	return false
}

// isNumberKind reports whether a value of the kind is a JSON number.
func isNumberKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}
