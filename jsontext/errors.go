package jsontext

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrDuplicateName reports that a JSON token could not be encoded or decoded because its name is already the
// name of another member of the same JSON object. It is wrapped directly by a SyntacticError, whose
// JSONPointer points to the duplicate name:
//
//	var serr *jsontext.SyntacticError
//	if errors.As(err, &serr) && serr.Err == jsontext.ErrDuplicateName {
//		ptr := serr.JSONPointer // JSON pointer to the duplicate member
//		name := ptr.LastToken() // the name itself
//		...
//	}
//
// It is only reported if AllowDuplicateNames is false.
var ErrDuplicateName = errors.New("duplicate object member name")

// ErrNonStringName reports that a JSON token could not be encoded or decoded because it is not a string where
// a JSON object name is, which RFC 8259, section 4, requires to be a string. It is wrapped directly by a
// SyntacticError.
var ErrNonStringName = errors.New("object member name must be a string")

// errInvalidUTF8 reports invalid UTF-8 in a JSON string.
var errInvalidUTF8 = errors.New("invalid UTF-8")

// The errors of the grammar, which an encoder reports where a token doesn't fit the structure.
var (
	errMismatchDelim = errors.New("mismatching structural token for object or array")
	errMissingValue  = errors.New("missing value after object name")
	errMaxDepth      = errors.New("exceeded max depth")
	errInvalidToken  = errors.New("invalid jsontext.Token")
)

// SyntacticError describes an error of the JSON grammar found while encoding or decoding.
//
// The contents of this error, as this package makes it, may change over time.
type SyntacticError struct {
	_ [0]func() // not comparable

	// ByteOffset is the offset in the input or output at or after which the error occurred.
	ByteOffset int64
	// JSONPointer points to the JSON value within which the error occurred (see RFC 6901).
	JSONPointer Pointer

	// Err is the underlying error.
	Err error
}

func (e *SyntacticError) Error() string {
	if e.Err == nil {
		return "jsontext: syntactic error"
	}
	if e.Err == ErrDuplicateName {
		s := "jsontext: " + e.Err.Error() + " " + strconv.Quote(e.JSONPointer.LastToken())
		if parent := e.JSONPointer.Parent(); parent != "" {
			s += " within " + strconv.Quote(string(parent))
		}
		return s
	}
	b := append([]byte("jsontext: "), e.Err.Error()...)
	if e.JSONPointer != "" {
		b = append(b, " within "...)
		b = strconv.AppendQuote(b, shortPointer(string(e.JSONPointer)))
	}
	if e.ByteOffset > 0 {
		b = append(b, " after offset "...)
		b = strconv.AppendInt(b, e.ByteOffset, 10)
	}
	return string(b)
}

func (e *SyntacticError) Unwrap() error { return e.Err }

// maxShownPointer is the length over which the pointer of an error is shown shortened: its start and its end,
// of at most half of it each, cut where a token starts if they can be.
const maxShownPointer = 100

func shortPointer(p string) string {
	if len(p) <= maxShownPointer {
		return p
	}
	half := maxShownPointer / 2
	head := strings.LastIndexByte(p[1:half], '/') + 1
	sep := "/…"
	if head <= 0 {
		for head = half; !utf8.RuneStart(p[head]); head-- {
		}
		sep = "…"
	}
	tail := strings.IndexByte(p[len(p)-half:], '/')
	if tail >= 0 {
		tail += len(p) - half
	} else {
		for tail = len(p) - half; !utf8.RuneStart(p[tail]); tail++ {
		}
	}
	return p[:head] + sep + p[tail:]
}

// ioError is an error of the reader of a Decoder or the writer of an Encoder.
type ioError struct {
	write bool
	err   error
}

func (e *ioError) Error() string {
	if e.write {
		return "jsontext: write error: " + e.err.Error()
	}
	return "jsontext: read error: " + e.err.Error()
}

func (e *ioError) Unwrap() error { return e.err }

// errBufferWriteAfterNext reports a write to the bytes.Buffer which a Decoder reads in place.
var errBufferWriteAfterNext = errors.New("invalid bytes.Buffer.Write call after calling bytes.Buffer.Next")

// textError is an error of the text of a token: an invalid character or escape sequence.
type textError struct {
	msg string
}

func (e *textError) Error() string { return e.msg }

// invalidChar is the error of the character at the start of b, which is not expected where the text says.
func invalidChar(b []byte, where string) error {
	return &textError{"invalid character " + quoteChar(b) + " " + where}
}

// quoteChar quotes the character at the start of b as a Go rune literal, or its first byte if it isn't valid
// UTF-8.
func quoteChar(b []byte) string {
	r, n := utf8.DecodeRune(b)
	if r == utf8.RuneError && n <= 1 {
		if len(b) == 0 {
			return "''"
		}
		return `'\x` + strconv.FormatUint(uint64(b[0])>>4, 16) + strconv.FormatUint(uint64(b[0])&0xf, 16) + `'`
	}
	return strconv.QuoteRune(r)
}

// invalidEscape is the error of the escape sequence at the start of b, which is quoted by backquotes if it
// can be read so, and else as a Go string.
func invalidEscape(b []byte, what string) error {
	s := string(b)
	q := "`" + s + "`"
	if !utf8.ValidString(s) || strings.ContainsFunc(s, func(r rune) bool {
		return r == '`' || unicode.IsSpace(r) || !unicode.IsPrint(r)
	}) {
		q = strconv.Quote(s)
	}
	return &textError{"invalid " + what + " " + q + " in string"}
}
