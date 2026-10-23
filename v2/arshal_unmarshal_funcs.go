package json

import (
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/decoder"
	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/jsontext"
)

// Unmarshalers is a list of functions that may override the unmarshal behavior of specific types. Populate
// WithUnmarshalers to use it with Unmarshal, UnmarshalRead, or UnmarshalDecode. A nil *Unmarshalers is equivalent
// to an empty list. There are no exported fields or methods on Unmarshalers.
type Unmarshalers struct {
	_ [0]func() // not comparable

	// funcs are the functions, which the decoder calls and compiles the decoders for.
	funcs *decoder.UnmarshalFuncs
}

// JoinUnmarshalers constructs a flattened list of unmarshal functions. If multiple functions in the list are
// applicable for a value of a given type, then those earlier in the list take precedence over those that come
// later. If a function returns errors.ErrUnsupported, then the next applicable function is called, otherwise the
// default unmarshaling behavior is used.
//
// For example, these two are equivalent:
//
//	JoinUnmarshalers(u0, JoinUnmarshalers(u1, u2), u3)
//	JoinUnmarshalers(u0, u1, u2, u3)
func JoinUnmarshalers(us ...*Unmarshalers) *Unmarshalers {
	var funcs []decoder.UnmarshalFunc
	for _, u := range us {
		if u != nil {
			funcs = append(funcs, u.funcs.Funcs...)
		}
	}
	if len(funcs) == 0 {
		return nil
	}
	return &Unmarshalers{funcs: &decoder.UnmarshalFuncs{Funcs: funcs}}
}

// UnmarshalFunc constructs a type-specific unmarshaler that specifies how to unmarshal values of type T. T must be
// an unnamed pointer or an interface type. The function is always provided with a non-nil pointer value.
//
// Implementations must follow the requirements of Unmarshaler.
//
// Implementations must not retain the input []byte or the value of T.
func UnmarshalFunc[T any](fn func([]byte, T) error) *Unmarshalers {
	t := reflect.TypeFor[T]()
	assertCastableTo(t, false)
	return &Unmarshalers{funcs: &decoder.UnmarshalFuncs{Funcs: []decoder.UnmarshalFunc{{
		Type: t,
		Decode: func(ctx *decoder.RuntimeContext, cursor, depth int64, p unsafe.Pointer, from reflect.Type) (int64, error) {
			start, end, err := ctx.ReadValue(cursor, depth, t, ierrors.UnmarshalJSONMethod)
			if err != nil {
				return 0, err
			}
			if err := fn(ctx.Buf[start:end:end], castTo[T](p, from)); err != nil {
				err = wrapErrUnsupported(err, "unmarshal function of type func([]byte, T) error")
				return 0, &ierrors.UnmarshalError{GoType: t, Start: start, Err: err, Method: ierrors.UnmarshalJSONMethod}
			}
			return end, nil
		},
	}}}}
}

// UnmarshalFromFunc constructs a type-specific unmarshaler that specifies how to unmarshal values of type T. T must
// be an unnamed pointer or an interface type. The function is always provided with a non-nil pointer value.
//
// Implementations must follow the requirements of UnmarshalerFrom.
//
// Implementations must not retain the pointer to jsontext.Decoder or the value of T.
func UnmarshalFromFunc[T any](fn func(*jsontext.Decoder, T) error) *Unmarshalers {
	t := reflect.TypeFor[T]()
	assertCastableTo(t, false)
	return &Unmarshalers{funcs: &decoder.UnmarshalFuncs{Funcs: []decoder.UnmarshalFunc{{
		Type: t,
		Decode: func(ctx *decoder.RuntimeContext, cursor, depth int64, p unsafe.Pointer, from reflect.Type) (int64, error) {
			return unmarshalFrom(ctx, cursor, depth, t, fromFunc[T]{fn, castTo[T](p, from)})
		},
		FromDecoder: true,
	}}}}
}

// WithUnmarshalers specifies a list of type-specific unmarshalers to use, which can be used to override the default
// unmarshal behavior for values of particular types.
//
// This only affects unmarshaling and is ignored when marshaling.
func WithUnmarshalers(v *Unmarshalers) Options {
	return &options.Config{Set: options.UnmarshalersSet, Unmarshalers: v}
}
