package json

import (
	"bytes"
	"errors"
	"reflect"
	"slices"

	"github.com/goccy/go-json/internal/encoder"
	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/jsontext"
)

var (
	errRawEmbedNotObject = errors.New("embedded raw value must be a JSON object")
	jsontextValueType    = reflect.TypeFor[jsontext.Value]()
	// lenient reads an embedded raw value, whose members are checked as they are written.
	lenient = JoinOptions(jsontext.AllowDuplicateNames(true), jsontext.AllowInvalidUTF8(true))
)

func init() {
	encoder.V2Hooks.AppendRawMembers = appendRawMembers
}

// appendRawMembers appends the members of the JSON object raw, a jsontext.Value embedded as a fallback, after b,
// which ends with the comma after the members before them or with the brace of the object, each member followed
// by a comma. check reports whether a name may be written, if it is not nil.
func appendRawMembers(ctx *encoder.RuntimeContext, b, raw []byte, check func(name []byte) bool) ([]byte, error) {
	if len(raw) == 0 {
		return b, nil
	}
	// an error of the raw value is reported at the place before the next member.
	st := stateOf(ctx)
	fail := func(b []byte, err error) error {
		pos, ptr := st.at(b)
		return &SemanticError{action: "marshal", ByteOffset: pos, JSONPointer: ptr, GoType: jsontextValueType, Err: toUnexpectedEOF(err)}
	}
	dec := jsontext.NewDecoder(bytes.NewReader(raw), lenient)
	tok, err := dec.ReadToken()
	if err != nil {
		return b, fail(b, err)
	}
	if tok.Kind() != '{' {
		return b, fail(b, errRawEmbedNotObject)
	}
	for dec.PeekKind() != '}' {
		name, err := dec.ReadValue()
		if err != nil {
			return b, fail(b, err)
		}
		if check != nil {
			unquoted, _ := jsontext.AppendUnquote(nil, name)
			if !check(unquoted) {
				return b, &ierrors.TextError{Err: ierrors.ErrDuplicateName, Name: string(unquoted), Out: b, AtDelim: true}
			}
		}
		if b, err = appendRawToken(ctx, b, name); err != nil {
			return b, err
		}
		b = append(b, ':')
		value, err := dec.ReadValue()
		if err != nil {
			return b, fail(b, err)
		}
		if b, err = appendRawToken(ctx, b, value); err != nil {
			return b, err
		}
		b = append(b, ',')
	}
	if _, err := dec.ReadToken(); err != nil {
		return b, fail(b, err)
	}
	if rest := bytes.TrimLeft(raw[dec.InputOffset():], " \t\r\n"); len(rest) > 0 {
		// a value after the object, which jsontext reports as it reports it after a top-level value.
		v := jsontext.Value(slices.Clone(raw))
		return b, fail(b, v.Compact(lenient))
	}
	return b, nil
}

// appendRawToken appends a name or a value of an embedded raw value, formatted as the options want it: its error
// is at its place in the output.
func appendRawToken(ctx *encoder.RuntimeContext, b, raw []byte) ([]byte, error) {
	out, err := appendRaw(ctx, b, raw)
	if serr, ok := err.(*jsontext.SyntacticError); ok && serr.JSONPointer != "" && len(raw) > 0 && raw[0] == '"' {
		// a name is in the object, not in the value of the member it names.
		serr.JSONPointer = serr.JSONPointer.Parent()
	}
	return out, err
}
