package encoder

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonstring"
)

// The keys of a map for the semantics of encoding/json/v2: a key is encoded as a value is, in the place of a name,
// where a number is a string, and a key which is not a string is an error. A key of a string, an integer or a
// float, which the VM writes, is a different name for every key of the map. The other keys are named here, and
// two keys of a map may have the same name: the names of a map are recorded to report it.

// v2MapKeyCode returns the code of the key of a map of the type typ.
func (c *Compiler) v2MapKeyCode(typ reflect.Type) (Code, error) {
	funcs := c.funcs.funcsOf(typ)
	// a time.Duration and a json.Number are named by appendKeyNameAfter, wherever they are.
	if len(funcs) == 0 && v2MethodOf(typ, noMethod) == noMethod && typ != timeDurationType && typ != jsonNumberType {
		switch typ.Kind() {
		case reflect.String:
			return c.stringCode(typ, false)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return c.mapKeyCode(typ)
		case reflect.Float32:
			return &FloatCode{typ: typ, bitSize: 32, isString: true}, nil
		case reflect.Float64:
			return &FloatCode{typ: typ, bitSize: 64, isString: true}, nil
		}
	}
	if typ.Kind() == reflect.Interface && typ.NumMethod() > 0 {
		return c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
			return appendMapKeyName(ctx, b, typ, methodInterfaceOf(typ, p), funcs)
		}), nil
	}
	return c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
		return appendMapKeyName(ctx, b, typ, p, funcs)
	}), nil
}

// methodInterfaceOf returns the address of the key at p of a map whose keys are of typ, an interface type with
// methods, as a value of typ: p has the words of the interface{} which the key is read into ( see
// newInterfaceKeyCollector ), not the ones of typ, whose first word is an itab.
func methodInterfaceOf(typ reflect.Type, p unsafe.Pointer) unsafe.Pointer {
	v := reflect.New(typ)
	if key := *(*any)(p); key != nil {
		v.Elem().Set(reflect.ValueOf(key))
	}
	return v.UnsafePointer()
}

// appendMapKeyName appends the name of the key at p, of typ, by the functions of marshaling, its method or its
// kind, and records it in the context of the map, failing if the map has the name already. The names of a map
// whose entries are sorted are checked after they are ( see checkSortedNames ).
func appendMapKeyName(ctx *RuntimeContext, b []byte, typ reflect.Type, p unsafe.Pointer, funcs []*MarshalFunc) ([]byte, error) {
	start := len(b)
	out, err := b, ErrUseDefault
	ctx.KeyName = true // for a method or a function which writes to an encoder, at the place of a name
	if len(funcs) > 0 {
		out, err = appendByFuncs(ctx, b, p, typ, funcs)
	}
	written := err != ErrUseDefault || v2MethodOf(typ, noMethod) != noMethod // by a function or a method
	if err == ErrUseDefault {
		out, err = appendKeyName(ctx, b, reflect.NewAt(typ, p).Elem())
	}
	ctx.KeyName = false
	if err != nil {
		return b, err
	}
	if out[start] != '"' {
		return b, &errors.TextError{Err: errors.ErrNonStringName, GoType: typ}
	}
	if ctx.Option.Flag&(AllowDuplicateNamesOption|UnorderedMapOption) != UnorderedMapOption {
		return out, nil
	}
	m := ctx.mapContexts[ctx.mapDepth-1]
	if !m.namesKept {
		if m.names == nil {
			m.names = map[string]struct{}{}
		}
		clear(m.names)
		m.namesKept = true
	}
	name := string(out[start:])
	if _, ok := m.names[name]; ok {
		unquoted := jsonstring.AppendUnescaped(nil, []byte(name[1:len(name)-1]))
		err := &errors.TextError{Err: errors.ErrDuplicateName, Name: string(unquoted)}
		if written {
			// the name was written by a function or a method, which the error is reported for.
			err.GoType = typ
		}
		return b, err
	}
	m.names[name] = struct{}{}
	return out, nil
}

// appendKeyName appends the name of the value v, which is addressable, as a JSON string.
func appendKeyName(ctx *RuntimeContext, b []byte, v reflect.Value) ([]byte, error) {
	return appendKeyNameAfter(ctx, b, v, noMethod)
}

// appendKeyNameAfter appends the name of the value v, which is addressable, by its method after the method
// after or by its kind.
func appendKeyNameAfter(ctx *RuntimeContext, b []byte, v reflect.Value, after int) ([]byte, error) {
	t := v.Type()
	switch v2MethodOf(t, after) {
	case methodMarshalJSONTo:
		// the encoder, at the place of a name, takes a string only.
		out, err := V2Hooks.MarshalTo(ctx, b, t, v.Addr().Interface())
		if err == ErrUseDefault {
			return appendKeyNameAfter(ctx, b, v, methodMarshalJSONTo)
		}
		return out, err
	case methodMarshalJSON:
		raw, err := v.Addr().Interface().(json.Marshaler).MarshalJSON()
		if err != nil {
			return b, &errors.MethodError{GoType: t, Err: unsupportedError(err, "MarshalJSON method"), Kind: errors.MethodJSON}
		}
		out, err := AppendRaw(ctx, b, raw)
		if err != nil {
			return b, &errors.MethodError{GoType: t, Err: err, Kind: errors.MethodJSON}
		}
		if out[len(b)] != '"' {
			return b, &errors.TextError{Err: errors.ErrNonStringName, GoType: t}
		}
		return out, nil
	case methodAppendText:
		m := v.Addr().Interface().(textAppender)
		return appendMethodText(ctx, b, t, "AppendText method", nil, func(_ unsafe.Pointer, dst []byte) ([]byte, error) {
			return m.AppendText(dst)
		})
	case methodMarshalText:
		m := v.Addr().Interface().(encoding.TextMarshaler)
		return appendMethodText(ctx, b, t, "MarshalText method", nil, func(_ unsafe.Pointer, dst []byte) ([]byte, error) {
			text, err := m.MarshalText()
			return append(dst, text...), err
		})
	}
	switch t {
	case timeDurationType:
		return b, &errors.SemanticError{GoType: t, Err: errNoDuration}
	case jsonNumberType:
		out, err := AppendNumberString(ctx, append(b, '"'), json.Number(v.String()))
		if err != nil {
			return b, err
		}
		return append(out, '"'), nil
	}
	switch t.Kind() {
	case reflect.String:
		return appendName(ctx, b, v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b = append(b, '"')
		return append(appendIntValue(b, v.Int()), '"'), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		b = append(b, '"')
		return append(appendUintValue(b, v.Uint()), '"'), nil
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return b, &errors.SemanticError{GoType: t, Err: errUnsupportedValue(f)}
		}
		bits := 64
		if t.Kind() == reflect.Float32 {
			bits = 32
		}
		b = append(b, '"')
		return append(appendFloatOfBits(b, f, bits), '"'), nil
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return b, &errors.TextError{Err: errors.ErrNonStringName, GoType: t}
		}
		elem := v.Elem()
		if t.Kind() == reflect.Interface {
			// the dynamic value, in a variable of its own, which is addressable.
			copied := reflect.New(elem.Type()).Elem()
			copied.Set(elem)
			elem = copied
		}
		// named by the functions of its own type first, as a key of that type is.
		if funcs := ctx.Option.Funcs.funcsOf(elem.Type()); len(funcs) > 0 {
			if out, err := appendByFuncs(ctx, b, elem.Addr().UnsafePointer(), elem.Type(), funcs); err != ErrUseDefault {
				return out, err
			}
		}
		return appendKeyName(ctx, b, elem)
	case reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 && t.Elem().PkgPath() == "" {
			src := unsafe.Slice((*byte)(v.Addr().UnsafePointer()), t.Len())
			b = append(b, '"')
			return append(base64.StdEncoding.AppendEncode(b, src), '"'), nil
		}
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return b, &errors.SemanticError{GoType: t}
	}
	return b, &errors.TextError{Err: errors.ErrNonStringName, GoType: t}
}

// appendName appends the name as a JSON string, which must be valid UTF-8 unless it is allowed.
func appendName(ctx *RuntimeContext, b []byte, name string) ([]byte, error) {
	if out, ok := jsonstring.AppendQuoted(textEscaper(ctx), b, name); ok {
		return out, nil
	} else if out, err := InvalidUTF8(ctx, b, out); err != nil {
		return b, err
	} else {
		return out, nil
	}
}

func appendIntValue(b []byte, v int64) []byte { return strconv.AppendInt(b, v, 10) }

func appendUintValue(b []byte, v uint64) []byte { return strconv.AppendUint(b, v, 10) }

func errUnsupportedValue(f float64) error { return fmt.Errorf("unsupported value: %v", f) }
