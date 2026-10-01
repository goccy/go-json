//go:build go1.27 && goexperiment.jsonv2

package jsontext_test

import (
	"bytes"
	stdjsontext "encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
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
		if i%5 == 4 {
			err = d.SkipValue()
			out.WriteString("[S")
		} else if i%3 == 2 {
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
		if i%5 == 4 {
			err = d.SkipValue()
			out.WriteString("[S")
		} else if i%3 == 2 {
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

// TestStdDiffWhiteSpaceOptions compares the output of the options of white space, in every order of up to three of
// them, given to an Encoder and to the functions and methods which format a value, whose own options come first.
func TestStdDiffWhiteSpaceOptions(t *testing.T) {
	type option struct {
		goJSON jsontext.Options
		std    stdjsontext.Options
	}
	options := []option{
		{jsontext.Multiline(true), stdjsontext.Multiline(true)},
		{jsontext.Multiline(false), stdjsontext.Multiline(false)},
		{jsontext.SpaceAfterColon(true), stdjsontext.SpaceAfterColon(true)},
		{jsontext.SpaceAfterColon(false), stdjsontext.SpaceAfterColon(false)},
		{jsontext.SpaceAfterComma(true), stdjsontext.SpaceAfterComma(true)},
		{jsontext.WithIndent("  "), stdjsontext.WithIndent("  ")},
		{jsontext.WithIndentPrefix(" "), stdjsontext.WithIndentPrefix(" ")},
	}
	const in = `{"a":[1,2],"b":{}}`
	for n := range (len(options) + 1) * (len(options) + 1) * (len(options) + 1) {
		var goJSON []jsontext.Options
		var std []stdjsontext.Options
		for k := n; k > 0; k /= len(options) + 1 {
			if i := k%(len(options)+1) - 1; i >= 0 {
				goJSON, std = append(goJSON, options[i].goJSON), append(std, options[i].std)
			}
		}
		var got, want []string
		var gb, sb bytes.Buffer
		jsontext.NewEncoder(&gb, goJSON...).WriteValue(jsontext.Value(in))
		stdjsontext.NewEncoder(&sb, std...).WriteValue(stdjsontext.Value(in))
		got, want = append(got, gb.String()), append(want, sb.String())
		for m := range 4 {
			g, s := jsontext.Value(in), stdjsontext.Value(in)
			switch m {
			case 0:
				g.Format(goJSON...)
				s.Format(std...)
			case 1:
				g.Compact(goJSON...)
				s.Compact(std...)
			case 2:
				g.Indent(goJSON...)
				s.Indent(std...)
			case 3:
				g.Canonicalize(goJSON...)
				s.Canonicalize(std...)
			}
			got, want = append(got, string(g)), append(want, string(s))
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("options %d: output of Encoder, Format, Compact, Indent and Canonicalize:\ngot:  %q\nwant: %q", n, got, want)
		}
	}
}

// TestStdDiffWriteValueInContext compares the errors of Encoder.WriteValue in every place of the grammar, for
// values which are empty, cut, invalid, or followed by more.
func TestStdDiffWriteValueInContext(t *testing.T) {
	contexts := [][]string{{}, {"["}, {"[", "1"}, {"{"}, {"{", `"k"`}, {"{", `"k"`, "1"}, {"[", "{", `"k"`}}
	values := []string{"", " ", "1 2", "[", "{", "[1,", `{"a"`, `{"a":`, "x", `"a" 2`, "1 x", "]", "}", "[1]]", `"\u"`, `{"a":1,"a":2}`}
	for _, c := range contexts {
		for _, v := range values {
			g, s := jsontext.NewEncoder(io.Discard), stdjsontext.NewEncoder(io.Discard)
			for _, tok := range c {
				switch tok {
				case "[":
					g.WriteToken(jsontext.BeginArray)
					s.WriteToken(stdjsontext.BeginArray)
				case "{":
					g.WriteToken(jsontext.BeginObject)
					s.WriteToken(stdjsontext.BeginObject)
				default:
					g.WriteValue(jsontext.Value(tok))
					s.WriteValue(stdjsontext.Value(tok))
				}
			}
			got := fmt.Sprintf("%s | %d %q", errorString(g.WriteValue(jsontext.Value(v))), g.StackDepth(), g.StackPointer())
			want := fmt.Sprintf("%s | %d %q", errorString(s.WriteValue(stdjsontext.Value(v))), s.StackDepth(), s.StackPointer())
			if got != want {
				t.Errorf("WriteValue(%q) after %q:\ngot:  %s\nwant: %s", v, c, got, want)
			}
		}
	}
}

// errBoom is the error of the readers of TestStdDiffReaderErrors.
var errBoom = errors.New("boom")

// dataThenErrBoom returns all its input together with errBoom, and then io.EOF.
type dataThenErrBoom struct{ b []byte }

func (r *dataThenErrBoom) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, errBoom
}

// noInput returns n with no error 150 times before each read of r.
type noInput struct {
	r     io.Reader
	n     int
	empty int
}

func (r *noInput) Read(p []byte) (int, error) {
	if r.empty < 150 {
		r.empty++
		return r.n, nil
	}
	r.empty = 0
	return r.r.Read(p)
}

// TestStdDiffReaderErrors compares reads from readers which fail after the input, or with it, and from a
// bytes.Buffer which is written to after the decoder read it.
func TestStdDiffReaderErrors(t *testing.T) {
	readers := map[string]func(string) io.Reader{
		"error after the input": func(s string) io.Reader {
			return io.MultiReader(strings.NewReader(s), iotest.ErrReader(errBoom))
		},
		"error after the input read by bytes": func(s string) io.Reader {
			return io.MultiReader(iotest.OneByteReader(strings.NewReader(s)), iotest.ErrReader(errBoom))
		},
		"error with the input": func(s string) io.Reader { return &dataThenErrBoom{[]byte(s)} },
		"error after reads of no input": func(s string) io.Reader {
			return io.MultiReader(&noInput{r: strings.NewReader(s), n: 0}, iotest.ErrReader(errBoom))
		},
		"error after reads of a negative count": func(s string) io.Reader {
			return io.MultiReader(&noInput{r: strings.NewReader(s), n: -1}, iotest.ErrReader(errBoom))
		},
	}
	inputs := []string{``, ` `, `123`, `tru`, `"ab`, `[`, `[[[[`, `[1,2`, `[12`, `[1e`, `["\u12`, `[1,"x`, `[1,2]`,
		`{"a"`, `{"a":1`, `{"a":"b"`, `{"a":{"b":`, `{"a":[1,2]}`,
		// a delimiter which the next token can't need, a value in an object or array, and a cut literal
		`,`, `[,`, `{"a",`, `1 ,`, `[1:`, `{:`, `{"a":1:`, `{"a":{`, `[[1,`, `{"a":{"b":1,`, `[{`, `{"a":[1,`,
		`[tru`, `[f`, `[[fa`, `{"a":fa`, `[1,tr`}
	for name, newReader := range readers {
		for _, in := range inputs {
			for _, ops := range []string{"TTTT", "VVV", "TVVV", "SSS", "TSSS", "TTVV"} {
				d, s := jsontext.NewDecoder(newReader(in)), stdjsontext.NewDecoder(newReader(in))
				var got, want string
				for _, op := range ops {
					var err, stdErr error
					switch op {
					case 'T':
						_, err = d.ReadToken()
						_, stdErr = s.ReadToken()
					case 'V':
						_, err = d.ReadValue()
						_, stdErr = s.ReadValue()
					case 'S':
						err, stdErr = d.SkipValue(), s.SkipValue()
					}
					got += fmt.Sprintf("[%s | %d %d %q] ", errorString(err), d.InputOffset(), d.StackDepth(), d.StackPointer())
					want += fmt.Sprintf("[%s | %d %d %q] ", errorString(stdErr), s.InputOffset(), s.StackDepth(), s.StackPointer())
				}
				if got != want {
					t.Errorf("%s of %q from a reader with an %s:\ngot:  %s\nwant: %s", ops, in, name, got, want)
				}
			}
		}
	}
	for _, in := range []string{"1 ", "[1,", `{"a"`, "12"} {
		b, stdB := bytes.NewBufferString(in), bytes.NewBufferString(in)
		d, s := jsontext.NewDecoder(b), stdjsontext.NewDecoder(stdB)
		var got, want string
		for i := range 5 {
			if i == 2 {
				b.WriteString("2]")
				stdB.WriteString("2]")
			}
			_, err := d.ReadToken()
			_, stdErr := s.ReadToken()
			got, want = got+errorString(err)+" ", want+errorString(stdErr)+" "
		}
		got, want = got+b.String(), want+stdB.String()
		if got != want {
			t.Errorf("ReadToken of %q, written to after the second:\ngot:  %s\nwant: %s", in, got, want)
		}
	}
}

// countWriter keeps what it is given, and reports n(len(p)) written with no error, which it sums.
type countWriter struct {
	b   bytes.Buffer
	n   func(int) int
	sum int64
}

func (w *countWriter) Write(p []byte) (int, error) {
	w.b.Write(p)
	n := w.n(len(p))
	w.sum += int64(n)
	return n, nil
}

// TestStdDiffWriterCounts compares the output with writers which report a count other than the length of their
// input, with no error: the output is taken as written, whatever the count. The offset is the sum of the counts
// after each top-level value; it differs from the one of encoding/json/jsontext, which writes at other times.
func TestStdDiffWriterCounts(t *testing.T) {
	counts := map[string]func(int) int{
		"zero":      func(int) int { return 0 },
		"negative":  func(int) int { return -1 },
		"half":      func(n int) int { return n / 2 },
		"too large": func(n int) int { return n + 3 },
	}
	in := `[1,"a",{"b":[true,null]},` + strings.Repeat(`"`+strings.Repeat("x", 300)+`",`, 40) + `2] 3 `
	for name, count := range counts {
		w, stdW := &countWriter{n: count}, &countWriter{n: count}
		e, s := jsontext.NewEncoder(w), stdjsontext.NewEncoder(stdW)
		d, sd := jsontext.NewDecoder(strings.NewReader(in+in)), stdjsontext.NewDecoder(strings.NewReader(in+in))
		for {
			tok, err := d.ReadToken()
			stdTok, stdErr := sd.ReadToken()
			if err != nil || stdErr != nil {
				break
			}
			if err, stdErr := e.WriteToken(tok), s.WriteToken(stdTok); errorString(err) != errorString(stdErr) {
				t.Errorf("WriteToken(%v) to a writer which reports a count of %s = %v, want %v", tok, name, err, stdErr)
			}
			if e.StackDepth() == 0 && e.OutputOffset() != w.sum {
				t.Errorf("OutputOffset after a value to a writer which reports a count of %s = %d, want %d", name, e.OutputOffset(), w.sum)
			}
		}
		if w.b.String() != stdW.b.String() {
			t.Errorf("output to a writer which reports a count of %s:\ngot:  %s\nwant: %s", name, w.b.String(), stdW.b.String())
		}
	}
}

// TestStdDiffCoderOptions compares the options which an Encoder or Decoder returns, given to an Encoder and to
// Compact, before and after a Reset of the coder.
func TestStdDiffCoderOptions(t *testing.T) {
	const in = `{"b":1,"a":[2]}`
	sets := [][2][]any{
		{nil, nil},
		{{jsontext.Multiline(true)}, {stdjsontext.Multiline(true)}},
		{{jsontext.WithIndent("  ")}, {stdjsontext.WithIndent("  ")}},
		{{jsontext.Multiline(true), jsontext.SpaceAfterColon(false)}, {stdjsontext.Multiline(true), stdjsontext.SpaceAfterColon(false)}},
		{{jsontext.SpaceAfterComma(true)}, {stdjsontext.SpaceAfterComma(true)}},
		{{jsontext.Multiline(false)}, {stdjsontext.Multiline(false)}},
	}
	goOptions := func(s []any) (o []jsontext.Options) {
		for _, v := range s {
			o = append(o, v.(jsontext.Options))
		}
		return o
	}
	stdOptions := func(s []any) (o []stdjsontext.Options) {
		for _, v := range s {
			o = append(o, v.(stdjsontext.Options))
		}
		return o
	}
	format := func(o jsontext.Options, s stdjsontext.Options, extra [2][]any) (string, string) {
		var b, stdB bytes.Buffer
		jsontext.NewEncoder(&b, append([]jsontext.Options{o}, goOptions(extra[0])...)...).WriteValue(jsontext.Value(in))
		stdjsontext.NewEncoder(&stdB, append([]stdjsontext.Options{s}, stdOptions(extra[1])...)...).WriteValue(stdjsontext.Value(in))
		v, stdV := jsontext.Value(in), stdjsontext.Value(in)
		v.Compact(append([]jsontext.Options{o}, goOptions(extra[0])...)...)
		stdV.Compact(append([]stdjsontext.Options{s}, stdOptions(extra[1])...)...)
		return b.String() + string(v), stdB.String() + string(stdV)
	}
	for i, set := range sets {
		e, stdE := jsontext.NewEncoder(io.Discard, goOptions(set[0])...), stdjsontext.NewEncoder(io.Discard, stdOptions(set[1])...)
		d, stdD := jsontext.NewDecoder(strings.NewReader(""), goOptions(set[0])...), stdjsontext.NewDecoder(strings.NewReader(""), stdOptions(set[1])...)
		for j, extra := range sets {
			if got, want := format(e.Options(), stdE.Options(), extra); got != want {
				t.Errorf("options of an encoder of %d, then %d:\ngot:  %q\nwant: %q", i, j, got, want)
			}
			if got, want := format(d.Options(), stdD.Options(), extra); got != want {
				t.Errorf("options of a decoder of %d, then %d:\ngot:  %q\nwant: %q", i, j, got, want)
			}
		}
		o, stdO, dO, stdDO := e.Options(), stdE.Options(), d.Options(), stdD.Options()
		e.Reset(io.Discard, o) // the options of the encoder itself
		stdE.Reset(io.Discard, stdO)
		if got, want := format(e.Options(), stdE.Options(), sets[0]); got != want {
			t.Errorf("options of an encoder of %d reset with them:\ngot:  %q\nwant: %q", i, got, want)
		}
		e.Reset(io.Discard)
		stdE.Reset(io.Discard)
		d.Reset(strings.NewReader(""))
		stdD.Reset(strings.NewReader(""))
		if got, want := format(o, stdO, sets[0]); got != want {
			t.Errorf("options of an encoder of %d, after a Reset:\ngot:  %q\nwant: %q", i, got, want)
		}
		if got, want := format(dO, stdDO, sets[0]); got != want {
			t.Errorf("options of a decoder of %d, after a Reset:\ngot:  %q\nwant: %q", i, got, want)
		}
	}
}

// TestStdDiffNestedIndexedObjects reads nested objects of so many names that both have an index of their names,
// with an inner value which the fast scan gives up on, valid or not: the scan which reads it again finds the
// names of none of them.
func TestStdDiffNestedIndexedObjects(t *testing.T) {
	var b strings.Builder
	b.WriteString("{")
	for i := range 70 {
		fmt.Fprintf(&b, `"a%d":1,`, i)
	}
	b.WriteString(`"z":{`)
	for i := range 70 {
		fmt.Fprintf(&b, `"b%d":1,`, i)
	}
	head := b.String()
	for _, tail := range []string{
		`"x":tru}}`,
		`"b1":2}}`,
		`"x":"A"}}`,
		`"x":` + strings.Repeat("[", 64) + strings.Repeat("]", 64) + `}}`,
	} {
		assertStdDiff(t, []byte(head+tail))
	}
}

// TestStdDiffSameOptionsAgain uses the functions of values, and an encoder which is reset, again with the options
// of their last use, after uses which change how they scan.
func TestStdDiffSameOptionsAgain(t *testing.T) {
	const in = `{"b":1,"a":[2,{}]}`
	for _, multiline := range []bool{false, true} {
		o, stdO := jsontext.Multiline(multiline), stdjsontext.Multiline(multiline)
		for range 2 {
			if got, want := jsontext.Value(in).IsValid(o), stdjsontext.Value(in).IsValid(stdO); got != want {
				t.Errorf("IsValid with Multiline(%v): got %v, want %v", multiline, got, want)
			}
			got, err := jsontext.AppendFormat(nil, in, o)
			want, stdErr := stdjsontext.AppendFormat(nil, []byte(in), stdO)
			if string(got) != string(want) || (err == nil) != (stdErr == nil) {
				t.Errorf("AppendFormat with Multiline(%v):\ngot:  %q, %v\nwant: %q, %v", multiline, got, err, want, stdErr)
			}
		}
		var b, stdB bytes.Buffer
		e, stdE := jsontext.NewEncoder(&b, o), stdjsontext.NewEncoder(&stdB, stdO)
		for range 2 {
			e.Reset(&b, o)
			stdE.Reset(&stdB, stdO)
			e.WriteValue(jsontext.Value(in))
			stdE.WriteValue(stdjsontext.Value(in))
		}
		got, _ := jsontext.AppendFormat(nil, in, e.Options())
		want, _ := stdjsontext.AppendFormat(nil, []byte(in), stdE.Options())
		if b.String()+string(got) != stdB.String()+string(want) {
			t.Errorf("an encoder reset with Multiline(%v):\ngot:  %q\nwant: %q", multiline, b.String()+string(got), stdB.String()+string(want))
		}
	}
}

// TestStdDiffEdgeCases compares the results of the methods and functions at the edges of their input.
func TestStdDiffEdgeCases(t *testing.T) {
	var voided, stdVoided = func() func() string {
		d := jsontext.NewDecoder(strings.NewReader(`["abc", 12]`))
		d.ReadToken()
		tok, _ := d.ReadToken()
		d.ReadToken()
		return func() string { return tok.Kind().String() }
	}(), func() func() string {
		d := stdjsontext.NewDecoder(strings.NewReader(`["abc", 12]`))
		d.ReadToken()
		tok, _ := d.ReadToken()
		d.ReadToken()
		return func() string { return tok.Kind().String() }
	}()
	resetFromBuffer := func(newEncoder func(io.Writer) (func(io.Writer), func(string))) string {
		var b bytes.Buffer
		b.Grow(100)
		var w strings.Builder
		reset, write := newEncoder(&b)
		write("null")
		reset(&w)
		write("[")
		write(`"abc"`)
		b.WriteString("XXXXXXXXXX")
		write("]")
		return b.String() + " " + w.String()
	}
	for _, c := range []struct {
		name      string
		got, want func() string
	}{
		{"NewEncoder(nil)", func() string { jsontext.NewEncoder(nil); return "" }, func() string { stdjsontext.NewEncoder(nil); return "" }},
		{"NewDecoder(nil)", func() string { jsontext.NewDecoder(nil); return "" }, func() string { stdjsontext.NewDecoder(nil); return "" }},
		{"(*Value)(nil).UnmarshalJSON", func() string { return fmt.Sprint((*jsontext.Value)(nil).UnmarshalJSON([]byte("1"))) },
			func() string { return fmt.Sprint((*stdjsontext.Value)(nil).UnmarshalJSON([]byte("1"))) }},
		{"AppendFloat of 16 bits", func() string { return string(jsontext.AppendFloat(nil, 1.5, 16)) },
			func() string { return string(stdjsontext.AppendFloat(nil, 1.5, 16)) }},
		{"Float(2⁶³).Int", func() string { return fmt.Sprint(jsontext.Float(1 << 63).Int()) }, func() string { return fmt.Sprint(stdjsontext.Float(1 << 63).Int()) }},
		{"Float(2⁶⁴).Uint", func() string { return fmt.Sprint(jsontext.Float(1 << 64).Uint()) }, func() string { return fmt.Sprint(stdjsontext.Float(1 << 64).Uint()) }},
		{"Float32(2⁶³).Int", func() string { return fmt.Sprint(jsontext.Float32(1 << 63).Int()) }, func() string { return fmt.Sprint(stdjsontext.Float32(1 << 63).Int()) }},
		{"Kind of a voided token", voided, stdVoided},
		{"Options of Multiline after SpaceAfterComma", func() string {
			var b bytes.Buffer
			jsontext.NewEncoder(&b, jsontext.SpaceAfterComma(true), jsontext.NewEncoder(io.Discard, jsontext.Multiline(true)).Options()).WriteValue(jsontext.Value(`[1,2]`))
			return b.String()
		}, func() string {
			var b bytes.Buffer
			stdjsontext.NewEncoder(&b, stdjsontext.SpaceAfterComma(true), stdjsontext.NewEncoder(io.Discard, stdjsontext.Multiline(true)).Options()).WriteValue(stdjsontext.Value(`[1,2]`))
			return b.String()
		}},
		{"AppendUnquote to nil", func() string {
			var s string
			for _, in := range []string{``, `""`, `"\u"`, `x`} {
				b, err := jsontext.AppendUnquote(nil, in)
				s += fmt.Sprintf("%#v %v ", b, err)
			}
			return s
		}, func() string {
			var s string
			for _, in := range []string{``, `""`, `"\u"`, `x`} {
				b, err := stdjsontext.AppendUnquote(nil, in)
				s += fmt.Sprintf("%#v %v ", b, err)
			}
			return s
		}},
		{"Decoder.StackIndex beyond the depth", func() string {
			d := jsontext.NewDecoder(strings.NewReader(`[1]`))
			d.ReadToken()
			return recovered(func() string { d.StackIndex(2); return "" })
		}, func() string {
			d := stdjsontext.NewDecoder(strings.NewReader(`[1]`))
			d.ReadToken()
			return recovered(func() string { d.StackIndex(2); return "" })
		}},
		{"Encoder.StackIndex beyond the depth", func() string {
			e := jsontext.NewEncoder(io.Discard)
			e.WriteToken(jsontext.BeginArray)
			return recovered(func() string { e.StackIndex(3); return "" })
		}, func() string {
			e := stdjsontext.NewEncoder(io.Discard)
			e.WriteToken(stdjsontext.BeginArray)
			return recovered(func() string { e.StackIndex(3); return "" })
		}},
		{"tokens and values after the next call", func() string {
			var s string
			for _, r := range []func(string) io.Reader{
				func(in string) io.Reader { return strings.NewReader(in) },
				func(in string) io.Reader { return bytes.NewBufferString(in) },
			} {
				d := jsontext.NewDecoder(r(`"abc" 12`))
				tok, _ := d.ReadToken()
				d.PeekKind()
				s += recovered(tok.String) + " "
				d = jsontext.NewDecoder(r(`[1 2]`))
				d.ReadToken()
				tok, _ = d.ReadToken()
				d.ReadToken() // an error
				s += recovered(tok.String) + " "
				d = jsontext.NewDecoder(r(`"abc"`))
				v, _ := d.ReadValue()
				d.ReadToken() // io.EOF
				s += string(v) + " "
			}
			return s
		}, func() string {
			var s string
			for _, r := range []func(string) io.Reader{
				func(in string) io.Reader { return strings.NewReader(in) },
				func(in string) io.Reader { return bytes.NewBufferString(in) },
			} {
				d := stdjsontext.NewDecoder(r(`"abc" 12`))
				tok, _ := d.ReadToken()
				d.PeekKind()
				s += recovered(tok.String) + " "
				d = stdjsontext.NewDecoder(r(`[1 2]`))
				d.ReadToken()
				tok, _ = d.ReadToken()
				d.ReadToken() // an error
				s += recovered(tok.String) + " "
				d = stdjsontext.NewDecoder(r(`"abc"`))
				v, _ := d.ReadValue()
				d.ReadToken() // io.EOF
				s += string(v) + " "
			}
			return s
		}},
		{"a token after Reset", func() string {
			var s string
			for _, c := range [][2]string{{`"abc"`, "123"}, {`"abc"`, ""}, {"12", "true"}} {
				d := jsontext.NewDecoder(bytes.NewBufferString(c[0]))
				tok, _ := d.ReadToken()
				d.Reset(bytes.NewBufferString(c[1]))
				if c[1] != "" {
					d.ReadToken()
				}
				s += recovered(func() string { return tok.Kind().String() }) + " " + recovered(tok.String) + " " +
					recovered(func() string { return fmt.Sprint(tok.Int()) }) + " "
			}
			return s
		}, func() string {
			var s string
			for _, c := range [][2]string{{`"abc"`, "123"}, {`"abc"`, ""}, {"12", "true"}} {
				d := stdjsontext.NewDecoder(bytes.NewBufferString(c[0]))
				tok, _ := d.ReadToken()
				d.Reset(bytes.NewBufferString(c[1]))
				if c[1] != "" {
					d.ReadToken()
				}
				s += recovered(func() string { return tok.Kind().String() }) + " " + recovered(tok.String) + " " +
					recovered(func() string { return fmt.Sprint(tok.Int()) }) + " "
			}
			return s
		}},
		{"SyntacticError without Err", func() string { return (&jsontext.SyntacticError{ByteOffset: 5, JSONPointer: "/a"}).Error() },
			func() string { return (&stdjsontext.SyntacticError{ByteOffset: 5, JSONPointer: "/a"}).Error() }},
		{"Encoder.Reset from a bytes.Buffer", func() string {
			return resetFromBuffer(func(w io.Writer) (func(io.Writer), func(string)) {
				e := jsontext.NewEncoder(w)
				return func(w io.Writer) { e.Reset(w) }, func(v string) {
					if v == "[" || v == "]" {
						e.WriteToken(map[string]jsontext.Token{"[": jsontext.BeginArray, "]": jsontext.EndArray}[v])
					} else {
						e.WriteValue(jsontext.Value(v))
					}
				}
			})
		}, func() string {
			return resetFromBuffer(func(w io.Writer) (func(io.Writer), func(string)) {
				e := stdjsontext.NewEncoder(w)
				return func(w io.Writer) { e.Reset(w) }, func(v string) {
					if v == "[" || v == "]" {
						e.WriteToken(map[string]stdjsontext.Token{"[": stdjsontext.BeginArray, "]": stdjsontext.EndArray}[v])
					} else {
						e.WriteValue(stdjsontext.Value(v))
					}
				}
			})
		}},
	} {
		if got, want := recovered(c.got), recovered(c.want); got != want {
			t.Errorf("%s:\ngot:  %s\nwant: %s", c.name, got, want)
		}
	}
}

// TestStdDiffLongPointer compares the messages of the errors of long pointers, which are shown shortened, and of
// duplicate names within them.
func TestStdDiffLongPointer(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	tokens := []string{"a", "bc", "é", "日本", "~0", "~1"}
	for range 100000 {
		var b strings.Builder
		for b.Len() < 60+r.IntN(200) {
			b.WriteByte('/')
			n := r.IntN(4)
			if r.IntN(5) == 0 {
				n = r.IntN(60)
			}
			for range n {
				b.WriteString(tokens[r.IntN(len(tokens))])
			}
		}
		p := b.String()
		for _, errs := range [][2]error{{io.ErrUnexpectedEOF, io.ErrUnexpectedEOF}, {jsontext.ErrDuplicateName, stdjsontext.ErrDuplicateName}} {
			got := (&jsontext.SyntacticError{JSONPointer: jsontext.Pointer(p), Err: errs[0]}).Error()
			want := (&stdjsontext.SyntacticError{JSONPointer: stdjsontext.Pointer(p), Err: errs[1]}).Error()
			if got != want {
				t.Fatalf("SyntacticError.Error() of the pointer %q:\ngot:  %s\nwant: %s", p, got, want)
			}
		}
	}
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
