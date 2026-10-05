package json

import (
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"

	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/jsontext"
)

// ErrUnknownName reports that a JSON object member could not be unmarshaled because the name is not known to the
// target Go struct. It is wrapped directly by a SemanticError, whose JSONPointer points to the member:
//
//	var serr *json.SemanticError
//	if errors.As(err, &serr) && serr.Err == json.ErrUnknownName {
//		ptr := serr.JSONPointer // JSON pointer to the unknown member
//		name := ptr.LastToken() // the name itself
//		...
//	}
//
// It is only reported if RejectUnknownMembers is true.
var ErrUnknownName = errors.New("unknown object member name")

// errAmbiguousName reports that a JSON object member could be unmarshaled into more than one field of the Go
// struct.
var errAmbiguousName = errors.New("ambiguous object member name")

// SemanticError describes an error determining the meaning of JSON data as Go data, or vice versa.
//
// If a Marshaler, MarshalerTo, Unmarshaler or UnmarshalerFrom method returns a SemanticError when called by this
// package, its ByteOffset, JSONPointer and GoType are filled in by the caller if they are the zero value.
//
// The contents of this error, as this package makes it, may change over time.
type SemanticError struct {
	_ [0]func() // not comparable

	// action is "marshal" or "unmarshal", or empty for an error which a method or a function made.
	action string

	// ByteOffset is the offset in the input or output at or after which the error occurred.
	ByteOffset int64
	// JSONPointer points to the JSON value within which the error occurred (see RFC 6901).
	JSONPointer jsontext.Pointer

	// JSONKind is the kind of the JSON value which could not be handled. It may be zero if it is unknown.
	JSONKind jsontext.Kind
	// JSONValue is the JSON number or string which could not be unmarshaled. It is not set by marshaling.
	JSONValue jsontext.Value
	// GoType is the Go type which could not be handled. It may be nil if it is unknown.
	GoType reflect.Type

	// Err is the underlying error. It may be nil.
	Err error
}

// maxShownType is the length over which the Go type of an error is shown by its kind.
const maxShownType = 100

func (e *SemanticError) Error() string {
	var b strings.Builder
	b.WriteString("json: cannot")
	var preposition string
	switch e.action {
	case "marshal":
		b.WriteString(" marshal")
		preposition = " from"
	case "unmarshal":
		b.WriteString(" unmarshal")
		preposition = " into"
	default:
		b.WriteString(" handle")
		preposition = " with"
	}
	switch e.JSONKind {
	case 'n':
		b.WriteString(" JSON null")
	case 'f', 't':
		b.WriteString(" JSON boolean")
	case '"':
		b.WriteString(" JSON string")
	case '0':
		b.WriteString(" JSON number")
	case '{', '}':
		b.WriteString(" JSON object")
	case '[', ']':
		b.WriteString(" JSON array")
	default:
		if e.action == "" {
			preposition = ""
		}
	}
	if len(e.JSONValue) > 0 && len(e.JSONValue) < 100 {
		b.WriteByte(' ')
		b.Write(e.JSONValue)
	}
	if e.GoType != nil {
		b.WriteString(preposition)
		b.WriteString(" Go ")
		b.WriteString(shownType(e.GoType))
	}

	if e.Err == ErrUnknownName || e.Err == errAmbiguousName {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
		b.WriteByte(' ')
		b.WriteString(strconv.Quote(e.JSONPointer.LastToken()))
		if parent := e.JSONPointer.Parent(); parent != "" {
			b.WriteString(" within ")
			b.WriteString(strconv.Quote(ierrors.ShortPointer(string(parent))))
		}
		return b.String()
	}

	// the place is not shown again if the syntactic error which this error wraps shows it.
	serr, _ := e.Err.(*jsontext.SyntacticError)
	switch {
	case e.JSONPointer != "":
		if serr == nil || !e.JSONPointer.Contains(serr.JSONPointer) {
			b.WriteString(" within ")
			b.WriteString(strconv.Quote(ierrors.ShortPointer(string(e.JSONPointer))))
		}
	case e.ByteOffset > 0:
		if serr == nil || e.ByteOffset > serr.ByteOffset {
			b.WriteString(" after offset ")
			b.WriteString(strconv.FormatInt(e.ByteOffset, 10))
		}
	}

	if e.Err != nil {
		msg := e.Err.Error()
		if serr != nil {
			msg = strings.TrimPrefix(msg, "jsontext: ")
		}
		b.WriteString(": ")
		b.WriteString(msg)
	}
	return b.String()
}

// shownType is how an error shows a Go type: its name, or its kind if the name is too long, as the one of a struct
// type with many fields is. The kind of an unnamed struct type with an unexported field is prefixed by the name
// of the package of the field.
func shownType(t reflect.Type) string {
	s := t.String()
	if len(s) <= maxShownType {
		return s
	}
	s = t.Kind().String()
	if t.Kind() == reflect.Struct && t.Name() == "" {
		for i := range t.NumField() {
			if pkgPath := t.Field(i).PkgPath; pkgPath != "" {
				return pkgPath[strings.LastIndexByte(pkgPath, '/')+1:] + ".struct"
			}
		}
	}
	return s
}

func (e *SemanticError) Unwrap() error {
	return e.Err
}

// marshalError returns the error of marshaling which the encoder returned with out, the output written before it,
// at its place in the whole output of the call st.
func (st *callState) marshalError(out []byte, err error) error {
	switch e := err.(type) {
	case *ierrors.SemanticError:
		pos, ptr := st.at(out)
		return &SemanticError{action: "marshal", ByteOffset: pos, JSONPointer: ptr, GoType: e.GoType, Err: e.Err}
	case *ierrors.OutputError:
		return st.marshalError(e.Out, e.Err)
	case *ierrors.MethodError:
		return st.methodError(out, e)
	case *ierrors.TextError:
		if e.Out != nil {
			out = e.Out
		}
		pos, ptr := st.at(out)
		serr := &jsontext.SyntacticError{ByteOffset: pos, JSONPointer: ptr, Err: e.Err}
		if e.AtDelim && len(out) > 0 && out[len(out)-1] == ',' {
			serr.ByteOffset--
		}
		if e.Err == ierrors.ErrDuplicateName {
			serr.JSONPointer = ptr.AppendToken(e.Name)
		}
		if e.GoType != nil {
			return &SemanticError{action: "marshal", ByteOffset: pos, JSONPointer: ptr, GoType: e.GoType, Err: serr}
		}
		return serr
	}
	return err
}

// methodError returns the error which a method or a function of marshaling returned, which the encoder returned
// with out, the output written before the call.
func (st *callState) methodError(out []byte, e *ierrors.MethodError) error {
	pos, ptr := st.at(out)
	inner := e.Err
	if te, ok := inner.(*ierrors.TextError); ok {
		// invalid UTF-8 of a text, at the place of its string.
		inner = &jsontext.SyntacticError{ByteOffset: pos, JSONPointer: ptr, Err: te.Err}
	}
	switch e.Kind {
	case ierrors.MethodText:
		if serr, ok := inner.(*SemanticError); ok {
			return serr
		}
	case ierrors.MethodJSON:
		if serr, ok := inner.(*SemanticError); ok {
			// its place is in the value which the method returned, after the place of the value.
			at := *serr
			at.ByteOffset += pos
			at.JSONPointer = ptr + serr.JSONPointer
			return &at
		}
	case ierrors.MethodJSONTo:
		// the place is where the method stopped writing.
		levels := st.levelsOf(e.Out)
		top := levels[len(levels)-1]
		pos, _ = st.at(e.Out)
		if len(levels) > 1 && top.Count > 0 {
			pos++ // the comma or the colon before the next token
		}
		serr, ok := inner.(*SemanticError)
		if !ok {
			serr = &SemanticError{Err: inner}
		}
		at := *serr
		at.Err = toUnexpectedEOF(at.Err)
		if at.action == "" {
			at.action = "marshal"
		}
		if at.ByteOffset == 0 {
			at.ByteOffset = pos
		}
		if at.JSONPointer == "" {
			at.JSONPointer = pointerOf(levels, e.Where)
		}
		if at.GoType == nil {
			at.GoType = e.GoType
		}
		return &at
	}
	return &SemanticError{action: "marshal", ByteOffset: pos, JSONPointer: ptr, GoType: e.GoType, Err: inner}
}

// toUnexpectedEOF converts io.EOF to io.ErrUnexpectedEOF: the end of the input in the middle of a value.
func toUnexpectedEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}
