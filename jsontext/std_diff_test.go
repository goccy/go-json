//go:build go1.27 && goexperiment.jsonv2

package jsontext_test

import (
	"bytes"
	stdjsontext "encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/goccy/go-json/jsontext"
)

// The tests here compare github.com/goccy/go-json/jsontext with encoding/json/jsontext, which it replaces: the
// same input must give the same tokens, values, output and errors, byte for byte, at the same offsets.

// stdDiffInputs are inputs which reach the paths of the scanners: white space of every kind and length, strings
// with every kind of escape, UTF-8 of every length and invalid UTF-8, numbers of every form, duplicate names,
// and every place where the input can end.
var stdDiffInputs = []string{
	``, ` `, "\t\r\n ", `null`, `true`, `false`, `nul`, `tru`, `fals`, `nulll`, `0`, `-0`, `1`, `-1`, `01`, `1.`, `1.5`,
	`1e3`, `1E+3`, `1e-3`, `-`, `1e`, `1.5e+`, `123456789012345678901234567890`, `1.7976931348623157e309`,
	`""`, `"a"`, `"abc"`, `"\""`, `"\\"`, `"\/"`, `"\b\f\n\r\t"`, `"\u0000"`, `"\u001F"`, `"\u00e9"`, `"\u00E9"`,
	`"\ud83d\ude00"`, `"\ud83d"`, `"\ude00"`, `"\ud83dx"`, `"\u12"`, `"\x"`, `"é"`, `"日本語"`, `"😀"`,
	"\"\xff\"", "\"\xc3\"", "\"\xe6\x97\"", "\"\xed\xa0\x80\"", "\"\xf4\x90\x80\x80\"", "\"a\x01b\"", "\"<>&\"",
	"\"\u2028\u2029\"", `"abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"`,
	`"abcdefghijklmnopqrstuvwxyz\"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"`, `"unterminated`, `"abc\`,
	`[]`, `{}`, `[1,2,3]`, `[1,2,]`, `[,]`, `[1 2]`, `{"a":1}`, `{"a":1,"b":2}`, `{"a":1,"a":2}`, `{"a":1,"\u0061":2}`,
	`{"a":{"a":1},"b":[{"a":1,"a":2}]}`, `{"a"}`, `{"a":}`, `{1:2}`, `{"a":1,}`, `{"a" : 1 , "b" : [ 1 , 2 ] }`,
	"{\n\t\"a\": [\n\t\t1,\n\t\t2\n\t],\n\t\"b\": {\n\t\t\"c\": \"d\"\n\t}\n}", `[[[[[[[[]]]]]]]]`, `]`, `}`,
	`1 2 3`, `{} []`, `"a" "b"`, `[1] x`, `{"abcdefghijklmnopqrstuvwxyz":1,"abcdefghijklmnopqrstuvwxyA":2}`,
	// values which end after a streaming decoder read more input, before and after its buffer is filled
	`0{"0"`, `0{"0":`, `0{"a":1,"b"`, `[{"a":`, `{"abc":[1,2`,
	`0 {"` + strings.Repeat("a", 70) + `":[1,{"b":`, `0 {"a":"` + strings.Repeat("x", 70) + `","b"`,
	`[` + strings.Repeat(`{"k":1},`, 20) + `{"k"`,
}

func FuzzStdDiff(f *testing.F) {
	for _, in := range stdDiffInputs {
		f.Add([]byte(in))
	}
	f.Fuzz(func(t *testing.T, in []byte) {
		assertStdDiff(t, in)
	})
}

func assertStdDiff(t *testing.T, in []byte) {
	t.Helper()
	for _, streaming := range []bool{false, true} {
		for _, opts := range stdDiffDecoderOptions {
			got := readTokens(in, streaming, opts.goJSON...)
			want := readStdTokens(in, streaming, opts.std...)
			if got != want {
				t.Fatalf("ReadToken of %q ( streaming %v, %s ):\ngot:  %s\nwant: %s", in, streaming, opts.name, got, want)
			}
			got = readValues(in, streaming, opts.goJSON...)
			want = readStdValues(in, streaming, opts.std...)
			if got != want {
				t.Fatalf("ReadValue of %q ( streaming %v, %s ):\ngot:  %s\nwant: %s", in, streaming, opts.name, got, want)
			}
			got = readMixed(in, streaming, opts.goJSON...)
			want = readStdMixed(in, streaming, opts.std...)
			if got != want {
				t.Fatalf("ReadToken and ReadValue of %q ( streaming %v, %s ):\ngot:  %s\nwant: %s", in, streaming, opts.name, got, want)
			}
		}
	}
	for _, opts := range stdDiffFormatOptions {
		got, gotErr := jsontext.AppendFormat(nil, in, opts.goJSON...)
		want, wantErr := stdjsontext.AppendFormat(nil, in, opts.std...)
		if !bytes.Equal(got, want) || errorString(gotErr) != errorString(wantErr) {
			t.Fatalf("AppendFormat of %q ( %s ):\ngot:  %q, %v\nwant: %q, %v", in, opts.name, got, gotErr, want, wantErr)
		}
		if g, w := jsontext.Value(in).IsValid(opts.goJSON...), stdjsontext.Value(in).IsValid(opts.std...); g != w {
			t.Fatalf("IsValid of %q ( %s ): got %v, want %v", in, opts.name, g, w)
		}
		// the output formatted again by every option, whose white space may be the one to write already.
		if wantErr != nil {
			continue
		}
		for _, opts2 := range stdDiffFormatOptions {
			got, gotErr := jsontext.AppendFormat(nil, want, opts2.goJSON...)
			want2, wantErr := stdjsontext.AppendFormat(nil, want, opts2.std...)
			if !bytes.Equal(got, want2) || errorString(gotErr) != errorString(wantErr) {
				t.Fatalf("AppendFormat of %q ( %s ) formatted by %s:\ngot:  %q, %v\nwant: %q, %v", in, opts2.name, opts.name, got, gotErr, want2, wantErr)
			}
		}
	}
	gotUnquote, gotErr := jsontext.AppendUnquote(nil, in)
	wantUnquote, wantErr := stdjsontext.AppendUnquote(nil, in)
	if !bytes.Equal(gotUnquote, wantUnquote) || errorString(gotErr) != errorString(wantErr) {
		t.Fatalf("AppendUnquote of %q:\ngot:  %q, %v\nwant: %q, %v", in, gotUnquote, gotErr, wantUnquote, wantErr)
	}
	gotQuote, gotErr := jsontext.AppendQuote(nil, in)
	wantQuote, wantErr := stdjsontext.AppendQuote(nil, in)
	if !bytes.Equal(gotQuote, wantQuote) || errorString(gotErr) != errorString(wantErr) {
		t.Fatalf("AppendQuote of %q:\ngot:  %q, %v\nwant: %q, %v", in, gotQuote, gotErr, wantQuote, wantErr)
	}
	for _, opts := range stdDiffEncoderOptions {
		got := writeTokens(in, opts.goJSON...)
		want := writeStdTokens(in, opts.std...)
		if got != want {
			t.Fatalf("WriteToken of the tokens of %q ( %s ):\ngot:  %s\nwant: %s", in, opts.name, got, want)
		}
		got = writeMixed(in, opts.goJSON...)
		want = writeStdMixed(in, opts.std...)
		if got != want {
			t.Fatalf("WriteToken and WriteValue of %q ( %s ):\ngot:  %s\nwant: %s", in, opts.name, got, want)
		}
	}
}

// readMixed reads the input by ReadToken, and by ReadValue every third call, and logs the state of the decoder
// after each call: its offset, its stack, and, when it reads from a bytes.Buffer, the buffer which it didn't read.
func readMixed(in []byte, streaming bool, opts ...jsontext.Options) string {
	d := jsontext.NewDecoder(reader(in, streaming), opts...)
	var out bytes.Buffer
	for i := 0; ; i++ {
		var err error
		if i%3 == 2 {
			var v jsontext.Value
			v, err = d.ReadValue()
			fmt.Fprintf(&out, "[V %q", v)
		} else {
			var tok jsontext.Token
			tok, err = d.ReadToken()
			fmt.Fprintf(&out, "[T %q%s", tok.String(), numberAccessors(tok))
		}
		fmt.Fprintf(&out, " %d %q", d.InputOffset(), d.StackPointer())
		for l := 0; l <= d.StackDepth(); l++ {
			k, n := d.StackIndex(l)
			fmt.Fprintf(&out, " %v:%d", k, n)
		}
		if !streaming {
			fmt.Fprintf(&out, " %q", d.UnreadBuffer())
		}
		out.WriteString("] ")
		if err != nil {
			fmt.Fprintf(&out, "%s", valueErrorString(err, streaming))
			return out.String()
		}
	}
}

func readStdMixed(in []byte, streaming bool, opts ...stdjsontext.Options) string {
	d := stdjsontext.NewDecoder(reader(in, streaming), opts...)
	var out bytes.Buffer
	for i := 0; ; i++ {
		var err error
		if i%3 == 2 {
			var v stdjsontext.Value
			v, err = d.ReadValue()
			fmt.Fprintf(&out, "[V %q", v)
		} else {
			var tok stdjsontext.Token
			tok, err = d.ReadToken()
			fmt.Fprintf(&out, "[T %q%s", tok.String(), stdNumberAccessors(tok))
		}
		fmt.Fprintf(&out, " %d %q", d.InputOffset(), d.StackPointer())
		for l := 0; l <= d.StackDepth(); l++ {
			k, n := d.StackIndex(l)
			fmt.Fprintf(&out, " %v:%d", k, n)
		}
		if !streaming {
			fmt.Fprintf(&out, " %q", d.UnreadBuffer())
		}
		out.WriteString("] ")
		if err != nil {
			fmt.Fprintf(&out, "%s", valueErrorString(err, streaming))
			return out.String()
		}
	}
}

// numberAccessors are the results of the methods of a number token, or of Float of a string token which may
// hold NaN or an infinity.
func numberAccessors(tok jsontext.Token) string {
	switch tok.Kind() {
	case '0':
		i, ierr := tok.Int()
		u, uerr := tok.Uint()
		f, ferr := tok.Float()
		f32, f32err := tok.Float32()
		return fmt.Sprintf(" %d %v %d %v %v %v %v %v", i, ierr, u, uerr, f, ferr, f32, f32err)
	case '"':
		return " " + recovered(func() string { f, err := tok.Float(); return fmt.Sprint(f, err) })
	}
	return ""
}

func stdNumberAccessors(tok stdjsontext.Token) string {
	switch tok.Kind() {
	case '0':
		i, ierr := tok.Int()
		u, uerr := tok.Uint()
		f, ferr := tok.Float()
		f32, f32err := tok.Float32()
		return fmt.Sprintf(" %d %v %d %v %v %v %v %v", i, ierr, u, uerr, f, ferr, f32, f32err)
	case '"':
		return " " + recovered(func() string { f, err := tok.Float(); return fmt.Sprint(f, err) })
	}
	return ""
}

// recovered is the result of f, or the value of its panic.
func recovered(f func() string) (s string) {
	defer func() {
		if r := recover(); r != nil {
			s = fmt.Sprint("panic: ", r)
		}
	}()
	return f()
}

// writeMixed writes the tokens of in, which are read with invalid UTF-8 and duplicate names allowed, and writes
// every other object and array by WriteValue, with white space around it.
func writeMixed(in []byte, opts ...jsontext.Options) string {
	d := jsontext.NewDecoder(bytes.NewReader(in), jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true))
	var out bytes.Buffer
	e := jsontext.NewEncoder(&out, opts...)
	var log bytes.Buffer
	for i := 0; ; i++ {
		if k := d.PeekKind(); i%2 == 0 && (k == '{' || k == '[' || k == '"' || k == '0') {
			v, err := d.ReadValue()
			if err != nil {
				break
			}
			if err := e.WriteValue(append(append([]byte(" "), v...), " \n"...)); err != nil {
				fmt.Fprintf(&log, "%s at %d ", errorString(err), e.OutputOffset())
			}
			continue
		}
		tok, err := d.ReadToken()
		if err != nil {
			break
		}
		if err := e.WriteToken(tok); err != nil {
			fmt.Fprintf(&log, "%s at %d ", errorString(err), e.OutputOffset())
		}
	}
	return log.String() + completeOutput(&out, e.StackDepth())
}

func writeStdMixed(in []byte, opts ...stdjsontext.Options) string {
	d := stdjsontext.NewDecoder(bytes.NewReader(in), stdjsontext.AllowInvalidUTF8(true), stdjsontext.AllowDuplicateNames(true))
	var out bytes.Buffer
	e := stdjsontext.NewEncoder(&out, opts...)
	var log bytes.Buffer
	for i := 0; ; i++ {
		if k := d.PeekKind(); i%2 == 0 && (k == '{' || k == '[' || k == '"' || k == '0') {
			v, err := d.ReadValue()
			if err != nil {
				break
			}
			if err := e.WriteValue(append(append([]byte(" "), v...), " \n"...)); err != nil {
				fmt.Fprintf(&log, "%s at %d ", errorString(err), e.OutputOffset())
			}
			continue
		}
		tok, err := d.ReadToken()
		if err != nil {
			break
		}
		if err := e.WriteToken(tok); err != nil {
			fmt.Fprintf(&log, "%s at %d ", errorString(err), e.OutputOffset())
		}
	}
	return log.String() + completeOutput(&out, e.StackDepth())
}

type stdDiffOptions struct {
	name   string
	goJSON []jsontext.Options
	std    []stdjsontext.Options
}

var stdDiffDecoderOptions = []stdDiffOptions{
	{name: "default"},
	{
		name:   "AllowDuplicateNames",
		goJSON: []jsontext.Options{jsontext.AllowDuplicateNames(true)},
		std:    []stdjsontext.Options{stdjsontext.AllowDuplicateNames(true)},
	},
	{
		name:   "AllowInvalidUTF8",
		goJSON: []jsontext.Options{jsontext.AllowInvalidUTF8(true)},
		std:    []stdjsontext.Options{stdjsontext.AllowInvalidUTF8(true)},
	},
}

var stdDiffFormatOptions = append(stdDiffDecoderOptions,
	stdDiffOptions{
		name:   "Multiline",
		goJSON: []jsontext.Options{jsontext.Multiline(true)},
		std:    []stdjsontext.Options{stdjsontext.Multiline(true)},
	},
	stdDiffOptions{
		name:   "WithIndent",
		goJSON: []jsontext.Options{jsontext.WithIndentPrefix("\t"), jsontext.WithIndent("  "), jsontext.SpaceAfterComma(true)},
		std:    []stdjsontext.Options{stdjsontext.WithIndentPrefix("\t"), stdjsontext.WithIndent("  "), stdjsontext.SpaceAfterComma(true)},
	},
	stdDiffOptions{
		name:   "Escape",
		goJSON: []jsontext.Options{jsontext.EscapeForHTML(true), jsontext.EscapeForJS(true)},
		std:    []stdjsontext.Options{stdjsontext.EscapeForHTML(true), stdjsontext.EscapeForJS(true)},
	},
	stdDiffOptions{
		name:   "PreserveRawStrings",
		goJSON: []jsontext.Options{jsontext.PreserveRawStrings(true), jsontext.AllowInvalidUTF8(true)},
		std:    []stdjsontext.Options{stdjsontext.PreserveRawStrings(true), stdjsontext.AllowInvalidUTF8(true)},
	},
	stdDiffOptions{
		name: "Canonicalize",
		goJSON: []jsontext.Options{
			jsontext.CanonicalizeRawInts(true), jsontext.CanonicalizeRawFloats(true), jsontext.ReorderRawObjects(true),
		},
		std: []stdjsontext.Options{
			stdjsontext.CanonicalizeRawInts(true), stdjsontext.CanonicalizeRawFloats(true), stdjsontext.ReorderRawObjects(true),
		},
	},
	stdDiffOptions{
		name: "ReorderIndented",
		goJSON: []jsontext.Options{
			jsontext.ReorderRawObjects(true), jsontext.AllowDuplicateNames(true), jsontext.WithIndent("\t"),
			jsontext.SpaceAfterColon(true),
		},
		std: []stdjsontext.Options{
			stdjsontext.ReorderRawObjects(true), stdjsontext.AllowDuplicateNames(true), stdjsontext.WithIndent("\t"),
			stdjsontext.SpaceAfterColon(true),
		},
	},
)

var stdDiffEncoderOptions = []stdDiffOptions{
	{name: "default"},
	stdDiffFormatOptions[1],
	stdDiffFormatOptions[3],
	stdDiffFormatOptions[5],
	stdDiffFormatOptions[6],
}

func errorString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T: %v", unwrapIO(err), err)
}

// unwrapIO is the error, or io.ErrUnexpectedEOF which it wraps, whose type is the same in both packages.
func unwrapIO(err error) error {
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return io.ErrUnexpectedEOF
	}
	var serr *jsontext.SyntacticError
	var stdErr *stdjsontext.SyntacticError
	if errors.As(err, &serr) || errors.As(err, &stdErr) {
		return errors.New("SyntacticError")
	}
	return err
}

// reader is the input of a decoder. A streaming decoder reads it a byte at a time, through a buffer of its own:
// each decoder sees the input which it read so far and no more, whatever the size of its buffer.
func reader(in []byte, streaming bool) io.Reader {
	if streaming {
		return iotest.OneByteReader(bytes.NewReader(in))
	}
	return bytes.NewBuffer(in)
}

func readTokens(in []byte, streaming bool, opts ...jsontext.Options) string {
	d := jsontext.NewDecoder(reader(in, streaming), opts...)
	var out bytes.Buffer
	for {
		k := d.PeekKind()
		tok, err := d.ReadToken()
		fmt.Fprintf(&out, "[%v %v %q %d %d %q] ", k, tok.Kind(), tok.String(), d.InputOffset(), d.StackDepth(), d.StackPointer())
		if err != nil {
			fmt.Fprintf(&out, "%s", tokenErrorString(err, streaming))
			return out.String()
		}
	}
}

func readStdTokens(in []byte, streaming bool, opts ...stdjsontext.Options) string {
	d := stdjsontext.NewDecoder(reader(in, streaming), opts...)
	var out bytes.Buffer
	for {
		k := d.PeekKind()
		tok, err := d.ReadToken()
		fmt.Fprintf(&out, "[%v %v %q %d %d %q] ", k, tok.Kind(), tok.String(), d.InputOffset(), d.StackDepth(), d.StackPointer())
		if err != nil {
			fmt.Fprintf(&out, "%s", tokenErrorString(err, streaming))
			return out.String()
		}
	}
}

func readValues(in []byte, streaming bool, opts ...jsontext.Options) string {
	d := jsontext.NewDecoder(reader(in, streaming), opts...)
	var out bytes.Buffer
	for {
		v, err := d.ReadValue()
		fmt.Fprintf(&out, "[%q %d] ", v, d.InputOffset())
		if err != nil {
			fmt.Fprintf(&out, "%s", valueErrorString(err, streaming))
			return out.String()
		}
	}
}

func readStdValues(in []byte, streaming bool, opts ...stdjsontext.Options) string {
	d := stdjsontext.NewDecoder(reader(in, streaming), opts...)
	var out bytes.Buffer
	for {
		v, err := d.ReadValue()
		fmt.Fprintf(&out, "[%q %d] ", v, d.InputOffset())
		if err != nil {
			fmt.Fprintf(&out, "%s", valueErrorString(err, streaming))
			return out.String()
		}
	}
}

// writeTokens writes the tokens of in, which are read with invalid UTF-8 and duplicate names allowed, so that
// the encoder is the one which judges them.
func writeTokens(in []byte, opts ...jsontext.Options) string {
	d := jsontext.NewDecoder(bytes.NewReader(in), jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true))
	var out bytes.Buffer
	e := jsontext.NewEncoder(&out, opts...)
	var log bytes.Buffer
	for {
		tok, err := d.ReadToken()
		if err != nil {
			break
		}
		for _, t := range []jsontext.Token{tok, exactToken(tok)} {
			if err := e.WriteToken(t); err != nil {
				fmt.Fprintf(&log, "%s at %d ", errorString(err), e.OutputOffset())
			}
		}
	}
	return log.String() + completeOutput(&out, e.StackDepth())
}

func writeStdTokens(in []byte, opts ...stdjsontext.Options) string {
	d := stdjsontext.NewDecoder(bytes.NewReader(in), stdjsontext.AllowInvalidUTF8(true), stdjsontext.AllowDuplicateNames(true))
	var out bytes.Buffer
	e := stdjsontext.NewEncoder(&out, opts...)
	var log bytes.Buffer
	for {
		tok, err := d.ReadToken()
		if err != nil {
			break
		}
		for _, t := range []stdjsontext.Token{tok, exactStdToken(tok)} {
			if err := e.WriteToken(t); err != nil {
				fmt.Fprintf(&log, "%s at %d ", errorString(err), e.OutputOffset())
			}
		}
	}
	return log.String() + completeOutput(&out, e.StackDepth())
}

// valueErrorString is errorString of an error of ReadValue. From a streaming decoder, the pointer is left out:
// encoding/json/jsontext reads the name of an object member from its buffer after a read of more input moved
// it, so the name in its pointer is the one of other bytes (`0{"0"` reports "/" for the member "/0").
func valueErrorString(err error, streaming bool) string {
	if !streaming {
		return errorString(err)
	}
	var serr *jsontext.SyntacticError
	var stdErr *stdjsontext.SyntacticError
	switch {
	case errors.As(err, &serr):
		return fmt.Sprintf("SyntacticError at %d: %s", serr.ByteOffset, cutCharacter(serr.Err.Error()))
	case errors.As(err, &stdErr):
		return fmt.Sprintf("SyntacticError at %d: %s", stdErr.ByteOffset, cutCharacter(stdErr.Err.Error()))
	}
	return errorString(err)
}

// tokenErrorString is errorString of an error of ReadToken, with the invalid character left out from a
// streaming decoder (see cutCharacter).
func tokenErrorString(err error, streaming bool) string {
	if !streaming {
		return errorString(err)
	}
	return cutCharacter(errorString(err))
}

// quotedInput is the invalid character or escape sequence which an error text quotes.
var quotedInput = regexp.MustCompile("(invalid character) '(\\\\.|[^\\\\'])+'|(invalid (escape sequence|surrogate pair)) (`[^`]*`|\"(\\\\.|[^\\\\\"])*\")")

// cutCharacter is the text of an error without the character or escape sequence which it quotes as invalid:
// encoding/json/jsontext quotes the bytes of the input which its buffer holds, which the end of the buffer may
// cut, as a buffer of a byte at a time does, where github.com/goccy/go-json/jsontext reads the rest first, so
// that its errors don't depend on how the input is read. The text of the error of encoding/json/jsontext is not
// structured: it is matched as text.
func cutCharacter(s string) string {
	return quotedInput.ReplaceAllString(s, "$1$3")
}

// completeOutput is the output of an encoder whose top-level values are complete: the output of a value which
// is not complete may still be in the buffer of the encoder, which is written when it is full enough.
func completeOutput(out *bytes.Buffer, depth int) string {
	if depth > 0 {
		return "<incomplete>"
	}
	return out.String()
}

// exactToken is tok made by a constructor, as a program which writes tokens makes them.
func exactToken(tok jsontext.Token) jsontext.Token {
	switch tok.Kind() {
	case '"':
		return jsontext.String(tok.String())
	case '0':
		if n, err := tok.Int(); err == nil {
			return jsontext.Int(n)
		}
		f, _ := tok.Float()
		return jsontext.Float(f)
	}
	return tok
}

func exactStdToken(tok stdjsontext.Token) stdjsontext.Token {
	switch tok.Kind() {
	case '"':
		return stdjsontext.String(tok.String())
	case '0':
		if n, err := tok.Int(); err == nil {
			return stdjsontext.Int(n)
		}
		f, _ := tok.Float()
		return stdjsontext.Float(f)
	}
	return tok
}
