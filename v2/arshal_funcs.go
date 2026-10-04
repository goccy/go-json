package json

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/jsontext"
)

var (
	errUnsupportedMutation = errors.New("unsupported calls must not read or write any tokens")
	errNonSingularValue    = errors.New("must read or write exactly one value")
)

// Marshalers is a list of functions that may override the marshal behavior of specific types. Populate
// WithMarshalers to use it with Marshal, MarshalWrite, or MarshalEncode. A nil *Marshalers is equivalent to an
// empty list. There are no exported fields or methods on Marshalers.
type Marshalers struct {
	_ [0]func() // not comparable

	// funcs are the functions, which the encoder calls and compiles the opcodes for.
	funcs *encoder.MarshalFuncs
}

// JoinMarshalers constructs a flattened list of marshal functions. If multiple functions in the list are
// applicable for a value of a given type, then those earlier in the list take precedence over those that come
// later. If a function returns errors.ErrUnsupported, then the next applicable function is called, otherwise the
// default marshaling behavior is used.
//
// For example, these two are equivalent:
//
//	JoinMarshalers(f0, JoinMarshalers(f1, f2), f3)
//	JoinMarshalers(f0, f1, f2, f3)
func JoinMarshalers(ms ...*Marshalers) *Marshalers {
	var funcs []encoder.MarshalFunc
	for _, m := range ms {
		if m != nil {
			funcs = append(funcs, m.funcs.Funcs...)
		}
	}
	if len(funcs) == 0 {
		return nil
	}
	return &Marshalers{funcs: &encoder.MarshalFuncs{Funcs: funcs}}
}

// MarshalFunc constructs a type-specific marshaler that specifies how to marshal values of type T. T can be any
// type except a named pointer. The function is always provided with a non-nil pointer value if T is an interface
// or pointer type.
//
// Implementations must follow the requirements of Marshaler.
//
// Implementations must not retain the value of T.
func MarshalFunc[T any](fn func(T) ([]byte, error)) *Marshalers {
	t := reflect.TypeFor[T]()
	assertCastableTo(t, true)
	return marshalersOf(encoder.MarshalFunc{
		Type: t,
		Append: func(ctx *encoder.RuntimeContext, b []byte, p unsafe.Pointer, from reflect.Type) ([]byte, error) {
			raw, err := fn(castTo[T](p, from))
			if err != nil {
				err = wrapErrUnsupported(err, "marshal function of type func(T) ([]byte, error)")
				return b, &ierrors.MethodError{GoType: t, Err: err, Kind: ierrors.MethodJSON}
			}
			out, err := appendRaw(ctx, b, raw)
			if err != nil {
				return b, &ierrors.MethodError{GoType: t, Err: err, Kind: ierrors.MethodJSON}
			}
			return out, nil
		},
	})
}

// MarshalToFunc constructs a type-specific marshaler that specifies how to marshal values of type T. T can be any
// type except a named pointer. The function is always provided with a non-nil pointer value if T is an interface
// or pointer type.
//
// Implementations must follow the requirements of MarshalerTo.
//
// Implementations must not retain the pointer to jsontext.Encoder or the value of T.
func MarshalToFunc[T any](fn func(*jsontext.Encoder, T) error) *Marshalers {
	t := reflect.TypeFor[T]()
	assertCastableTo(t, true)
	return marshalersOf(encoder.MarshalFunc{
		Type: t,
		Append: func(ctx *encoder.RuntimeContext, b []byte, p unsafe.Pointer, from reflect.Type) ([]byte, error) {
			return marshalTo(ctx, b, t, nil, func(enc *jsontext.Encoder) error {
				return fn(enc, castTo[T](p, from))
			})
		},
	})
}

func marshalersOf(f encoder.MarshalFunc) *Marshalers {
	return &Marshalers{funcs: &encoder.MarshalFuncs{Funcs: []encoder.MarshalFunc{f}}}
}

// wrapErrUnsupported returns the error of a function which must not report errors.ErrUnsupported.
func wrapErrUnsupported(err error, what string) error {
	if errors.Is(err, errors.ErrUnsupported) {
		return errors.New(what + " may not return errors.ErrUnsupported")
	}
	return err
}

// WithMarshalers specifies a list of type-specific marshalers to use, which can be used to override the default
// marshal behavior for values of particular types.
//
// This only affects marshaling and is ignored when unmarshaling.
func WithMarshalers(v *Marshalers) Options {
	return &options.Config{Set: options.MarshalersSet, Marshalers: v}
}

// assertCastableTo panics unless a value of some type is castable to the type to ( see castableTo ). Marshaling
// takes any type except a named pointer, which taking the address of a value never makes; unmarshaling takes an
// unnamed pointer or an interface only, as the value it is given must be changed by it.
func assertCastableTo(to reflect.Type, marshal bool) {
	switch to.Kind() {
	case reflect.Interface:
		return
	case reflect.Pointer:
		if to.Name() == "" {
			return
		}
	default:
		if marshal {
			return
		}
	}
	if marshal {
		panic(fmt.Sprintf("input type %v must be an interface type, an unnamed pointer type, or a non-pointer type", to))
	}
	panic(fmt.Sprintf("input type %v must be an interface type or an unnamed pointer type", to))
}

// castTo returns the value at p, of the type from, as the type T which it is castable to ( see castableTo ).
func castTo[T any](p unsafe.Pointer, from reflect.Type) T {
	var zero T
	switch reflect.TypeFor[T]().Kind() {
	case reflect.Interface:
		v, _ := reflect.NewAt(from, p).Interface().(T)
		return v
	case reflect.Pointer:
		return *(*T)(unsafe.Pointer(&p))
	default:
		if p == nil {
			return zero
		}
		return *(*T)(p)
	}
}
