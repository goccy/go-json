package json

import (
	"encoding"
	"reflect"

	"github.com/goccy/go-json/internal/jsonfields"
	"github.com/goccy/go-json/jsontext"
)

// Marshaler is implemented by types that can marshal themselves. It is recommended that types implement
// MarshalerTo unless the implementation is trying to avoid directly depending on the jsontext package.
//
// Implementations should return a buffer that is safe for the caller to retain and potentially mutate.
//
// Implementations must not return errors.ErrUnsupported.
//
// If the returned error is a SemanticError, then unpopulated fields of the error may be populated by this package
// with additional context. Errors of other types are wrapped within a SemanticError.
//
// Implementations should assume Deterministic is true and return deterministic output.
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

// MarshalerTo is implemented by types that can marshal themselves. It is recommended that types implement
// MarshalerTo instead of Marshaler since it is both more performant and more flexible. If a type implements both
// Marshaler and MarshalerTo, then MarshalerTo takes precedence. In such a case, both implementations should aim
// to have equivalent behavior for the default marshal options.
//
// The implementation must write only one JSON value to the Encoder. Alternatively, it may return
// errors.ErrUnsupported without mutating the Encoder. This package calling the method will use the next available
// JSON representation for the receiver type, as described in Marshal. Implementations must not retain the pointer
// to jsontext.Encoder.
//
// If the returned error is a SemanticError, then unpopulated fields of the error may be populated by this package
// with additional context. Errors of other types are wrapped within a SemanticError, except for IO errors.
//
// The MarshalJSONTo method should not be called directly as it may return sentinel errors that need special
// handling. Users should instead call MarshalEncode, which handles such cases.
//
// Implementations should inspect the marshal options from jsontext.Encoder.Options and adjust behavior to respect
// the options as necessary.
//
// The following options may be relevant to MarshalerTo implementations:
//
//   - Deterministic: if the implementation may produce non-deterministic output
//   - StringifyNumbers: if the type is represented as a JSON number
//
// Several options, such as FormatNilSliceAsNull, apply only to native Go types. Thus, these options are typically
// not directly relevant to MarshalerTo implementations. However, types representing a composite type should
// marshal contained types using MarshalEncode to ensure these options apply to the contained types. Similarly,
// WithMarshalers may influence marshaling of any contained type within a composite type.
//
// All other options are automatically handled outside of the MarshalerTo implementation, and thus are not
// relevant to implementations.
type MarshalerTo interface {
	MarshalJSONTo(*jsontext.Encoder) error
}

// The interfaces of the methods of marshaling and unmarshaling which a type may have. The ones of unmarshaling are
// for the rules of the fields of a struct: a type which has one of them can't be embedded.
type (
	textAppender    interface{ AppendText([]byte) ([]byte, error) }
	unmarshaler     interface{ UnmarshalJSON([]byte) error }
	unmarshalerFrom interface {
		UnmarshalJSONFrom(*jsontext.Decoder) error
	}
)

func init() {
	jsonfields.RawValueType = reflect.TypeFor[jsontext.Value]()
	jsonfields.MethodTypes = []reflect.Type{
		reflect.TypeFor[MarshalerTo](),
		reflect.TypeFor[Marshaler](),
		reflect.TypeFor[textAppender](),
		reflect.TypeFor[encoding.TextMarshaler](),
		reflect.TypeFor[unmarshalerFrom](),
		reflect.TypeFor[unmarshaler](),
		reflect.TypeFor[encoding.TextUnmarshaler](),
	}
}
