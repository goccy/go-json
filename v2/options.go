package json

import (
	"reflect"

	"github.com/goccy/go-json/internal/options"
)

// Options configures Marshal, MarshalWrite, MarshalEncode, Unmarshal, UnmarshalRead and UnmarshalDecode. It is
// the same type as jsontext.Options, so that the options of the jsontext package may be given to them as well,
// and the options of this package to jsontext.NewEncoder and jsontext.NewDecoder. In a list of options, an option
// given later overrides a property set by an earlier one.
//
// The options of this package:
//
//   - StringifyNumbers affects marshaling and unmarshaling
//   - Deterministic affects marshaling only
//   - FormatNilSliceAsNull affects marshaling only
//   - FormatNilMapAsNull affects marshaling only
//   - OmitZeroStructFields affects marshaling only
//   - MatchCaseInsensitiveNames affects marshaling and unmarshaling
//   - RejectUnknownMembers affects unmarshaling only
//   - WithMarshalers affects marshaling only
//   - WithUnmarshalers affects unmarshaling only
//
// An option which doesn't affect an operation is ignored by it.
type Options = options.Options

// JoinOptions returns the options of srcs as a single Options, in which a property set by a later option
// overrides the one of an earlier option.
func JoinOptions(srcs ...Options) Options {
	c := new(options.Config)
	c.Apply(srcs)
	return c
}

// GetOption returns the value which opts has for the property that setter sets, and whether opts sets it. If
// it doesn't, the value is the zero value of T.
//
//	v, ok := json.GetOption(opts, json.Deterministic)
//
// It is mostly for MarshalerTo and UnmarshalerFrom methods, and MarshalToFunc and UnmarshalFromFunc functions,
// which change the representation of a value by the options; those should usually ignore whether the property
// is set.
func GetOption[T any](opts Options, setter func(T) Options) (T, bool) {
	var zero T
	if opts == nil || setter == nil {
		return zero, false
	}
	// the property is the one which the setter sets: an option made by a setter of this module sets one property,
	// besides Multiline for the indentation options of jsontext.
	var probe options.Config
	setter(zero).ApplyTo(&probe)
	f := probe.Set
	if f&(options.IndentSet|options.PrefixSet) != 0 {
		f &= options.IndentSet | options.PrefixSet
	}
	if f == 0 || f&(f-1) != 0 {
		return zero, false
	}
	var c options.Config
	opts.ApplyTo(&c)
	if c.Set&f == 0 {
		if f != options.StringifyNumbers || c.Value&options.StringTag == 0 {
			return zero, false
		}
		// the options of a method of a value of the `string` option, as encoding/json/v2 reports them
		c.Value |= f
	}
	var v any
	switch f {
	case options.IndentSet:
		v = c.Indent
	case options.PrefixSet:
		v = c.Prefix
	case options.MarshalersSet:
		v = c.Marshalers
	case options.UnmarshalersSet:
		v = c.Unmarshalers
	default:
		v = c.Value&f != 0
	}
	if t, ok := v.(T); ok {
		return t, true
	}
	// a nil interface value of a pointer type, as the marshalers of a Config which were set to nil
	return zero, reflect.TypeFor[T]().Kind() == reflect.Pointer
}

// DefaultOptionsV2 sets every option of the v2 semantics: the options which the v1 json package sets by default,
// all false. Other options are not set.
func DefaultOptionsV2() Options {
	return defaultOptionsV2
}

var defaultOptionsV2 = &options.Config{
	Set: options.AllowDuplicateNames | options.AllowInvalidUTF8 | options.EscapeForHTML | options.EscapeForJS |
		options.PreserveRawStrings | options.Deterministic | options.FormatNilMapAsNull | options.FormatNilSliceAsNull |
		options.MatchCaseInsensitiveNames,
}

// StringifyNumbers specifies that a Go value which would be a JSON number is a JSON string of that number. When
// unmarshaling, the value is read from a JSON string which holds only the number, without white space.
//
// The string option of a struct field applies it to the field only.
//
// This affects marshaling and unmarshaling.
func StringifyNumbers(v bool) Options { return options.BoolOf(options.StringifyNumbers, v) }

// Deterministic specifies that marshaling the same value always writes the same bytes: for example, the
// entries of a Go map are written in the order of their keys. Of the values of the Go types, the output is
// the same for the same binary, but may differ between builds of a program, as a newer version may change the
// representation. It is not a canonical form: Deterministic and the canonical form of RFC 8785 are different
// things.
//
// This affects marshaling only.
func Deterministic(v bool) Options { return options.BoolOf(options.Deterministic, v) }

// FormatNilSliceAsNull specifies that a nil Go slice is the JSON null, rather than an empty JSON array (or an
// empty JSON string, of a ~[]byte).
//
// This affects marshaling only.
func FormatNilSliceAsNull(v bool) Options { return options.BoolOf(options.FormatNilSliceAsNull, v) }

// FormatNilMapAsNull specifies that a nil Go map is the JSON null, rather than an empty JSON object.
//
// This affects marshaling only.
func FormatNilMapAsNull(v bool) Options { return options.BoolOf(options.FormatNilMapAsNull, v) }

// OmitZeroStructFields specifies that the fields of a Go struct which are zero are left out: a value is zero if
// its type has an "IsZero() bool" method which returns true, or, without the method, if it is the zero value of
// its type. It is the omitzero option on every field of every struct.
//
// This affects marshaling only.
func OmitZeroStructFields(v bool) Options { return options.BoolOf(options.OmitZeroStructFields, v) }

// MatchCaseInsensitiveNames specifies that the names of the members of a JSON object are matched to the fields of
// a Go struct case-insensitively. If a name matches more than one field, the field whose name is the same is
// taken, and if none is, it is an error. A field with the case:strict or case:ignore option always matches its
// name case-sensitively or case-insensitively, whatever this option is.
//
// This affects marshaling and unmarshaling: when marshaling, the check of the duplicate names of the fields
// follows it.
func MatchCaseInsensitiveNames(v bool) Options {
	return options.BoolOf(options.MatchCaseInsensitiveNames, v)
}

// RejectUnknownMembers specifies that a member of a JSON object which matches no field of the Go struct is an
// error when unmarshaling.
//
// This affects unmarshaling only.
func RejectUnknownMembers(v bool) Options { return options.BoolOf(options.RejectUnknownMembers, v) }
