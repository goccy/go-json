//go:build go1.27 && goexperiment.jsonv2

package encoder

import (
	"encoding"
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"math"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

// The keys of a map of an interface type, which encoding/json encodes when it is built on encoding/json/v2
// ( GOEXPERIMENT=jsonv2, the default since Go 1.27; this file uses the API of Go 1.27 ): the name of a key is
// made from its dynamic value as a value in the place of a name is, and a key whose value is not written as a
// string is an error.
//
//   - a string, an integer, a float32 and a float64 are their text; a float64 held by a key of interface{}
//     is not, as interface{} writes it as a number.
//   - a value with AppendText or MarshalText of its value, or of its pointer through a pointer, is its text;
//     MarshalJSON is not called for a name.
//   - a pointer is the name of its value.
//   - nil, a bool, a nil pointer, a struct, an array, a slice or a map is not a name; a complex number, a
//     channel or a function is an unsupported type.
//
// The keys are sorted by their names, as the keys of the maps whose keys are not strings are: OpMapKey sorts them
// after the last one is encoded ( see MapContext.SortEncodedKeys ).

// interfaceMapKeys is whether a map of an interface type as its key is encoded.
const interfaceMapKeys = true

var (
	emptyInterfaceType  = reflect.TypeOf((*any)(nil)).Elem()
	float64Type         = reflect.TypeOf(float64(0))
	textAppenderType    = reflect.TypeOf((*encoding.TextAppender)(nil)).Elem()
	jsonMarshalerType   = reflect.TypeOf((*jsonv1.Marshaler)(nil)).Elem()
	jsonMarshalerToType = reflect.TypeOf((*jsonv2.MarshalerTo)(nil)).Elem()
)

// appendInterfaceMapKey appends the name of the key of a map of an interface type, or of a type of pointers which
// is named by what they lead to ( see pointerMapKeyCode ), as a string. p is the address of the key as the map is
// read into the context: an interface{} for a key of an interface type, whatever the interface type is ( see
// newReflectCollector ), and the pointer for a key of pointers.
func appendInterfaceMapKey(ctx *RuntimeContext, code *Opcode, b []byte, p unsafe.Pointer) ([]byte, error) {
	keyType := runtime.TypeOfPtr(code.Type)
	var name string
	var err error
	if keyType.Kind() == reflect.Pointer {
		name, err = mapKeyName(reflect.NewAt(keyType, p).Elem(), nil)
	} else {
		name, err = mapKeyName(reflect.NewAt(emptyInterfaceType, p).Elem(), keyType)
	}
	if err != nil {
		return nil, err
	}
	return appendText(ctx, b, name), nil
}

// mapKeyName returns the name of the key v: its pointers and its interfaces are followed, one by one, to the value
// which names it. iface is the interface type of v when v is the interface{} which a key of an interface type is
// read into.
func mapKeyName(v reflect.Value, iface reflect.Type) (string, error) {
	var cycle keyCycle
	// byPointer is whether the value is the one of a pointer, whose methods on the pointer are called.
	byPointer := false
	for {
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				return "", notStringMapKey(v)
			}
			if iface == nil {
				// an interface which a pointer of the key points to.
				iface = v.Type()
			}
			v, byPointer = v.Elem(), false
			if v.Kind() == reflect.Pointer && v.IsNil() {
				// The method of the marshaler which the interface type is, if any, is called with the nil pointer. A
				// MarshalJSON is not called for a name.
				switch {
				case iface.Implements(jsonMarshalerToType), iface.Implements(jsonMarshalerType):
					return "", notStringMapKey(v)
				case iface.Implements(textAppenderType):
					return appendTextName(v.Interface().(encoding.TextAppender))
				case iface.Implements(marshalTextType):
					return marshalTextName(v.Interface().(encoding.TextMarshaler))
				}
				return "", notStringMapKey(v)
			}
			if iface == emptyInterfaceType && v.Type() == float64Type {
				// interface{} writes a float64 as a number, which is not a name.
				if f := v.Float(); math.IsNaN(f) || math.IsInf(f, 0) {
					return "", nonFiniteMapKey(v, f)
				}
				return "", notStringMapKey(v)
			}
			iface = nil
			continue
		case reflect.Pointer:
			if v.IsNil() {
				return "", notStringMapKey(v)
			}
			if cycle.reached(v) {
				return "", &errors.UnsupportedValueError{Value: v, Str: "encountered a cycle via " + v.Type().String()}
			}
			v, byPointer = v.Elem(), true
			continue
		}
		break
	}
	t := v.Type()
	switch {
	case t.Implements(textAppenderType):
		return appendTextName(v.Interface().(encoding.TextAppender))
	case byPointer && reflect.PointerTo(t).Implements(textAppenderType):
		return appendTextName(v.Addr().Interface().(encoding.TextAppender))
	case t.Implements(marshalTextType):
		return marshalTextName(v.Interface().(encoding.TextMarshaler))
	case byPointer && reflect.PointerTo(t).Implements(marshalTextType):
		return marshalTextName(v.Addr().Interface().(encoding.TextMarshaler))
	}
	switch t.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return "", nonFiniteMapKey(v, f)
		}
		bits := 64
		if t.Kind() == reflect.Float32 {
			bits = 32
		}
		return string(appendFloatOfBits(nil, f, bits)), nil
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return "", &errors.UnsupportedTypeError{Type: t}
	}
	return "", notStringMapKey(v)
}

func appendTextName(m encoding.TextAppender) (string, error) {
	b, err := m.AppendText(nil)
	if err != nil {
		return "", errors.ErrMarshaler(reflect.TypeOf(m), err, "AppendText")
	}
	return string(b), nil
}

func marshalTextName(m encoding.TextMarshaler) (string, error) {
	b, err := m.MarshalText()
	if err != nil {
		return "", errors.ErrMarshaler(reflect.TypeOf(m), err, "MarshalText")
	}
	return string(b), nil
}

// notStringMapKey is the error of a key whose value is not written as a string, as encoding/json reports it.
func notStringMapKey(v reflect.Value) error {
	return &errors.UnsupportedValueError{Value: v, Str: "object member name must be a string"}
}

// nonFiniteMapKey is the error of a key of a float which is not finite, as encoding/json reports it.
func nonFiniteMapKey(v reflect.Value, f float64) error {
	return &errors.UnsupportedValueError{Value: v, Str: strconv.FormatFloat(f, 'g', -1, 64)}
}
