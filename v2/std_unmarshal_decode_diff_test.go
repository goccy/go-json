//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The tests here compare UnmarshalDecode with the one of encoding/json/v2, for the inputs of the tests of Unmarshal,
// from decoders of readers of a few sizes of reads: at the top level of the decoder, in an array which the decoder
// read the bracket of, after the values before it, and called by a method which reads the elements of an array. The
// decoders are compared only up to the first error: where the decoder of encoding/json/v2 is after an error is not
// imitated.

// decodeDiffWhere are the places of the value which the tests decode: the top level, the second value of the
// top level, and an element of an array.
var decodeDiffWhere = []string{"top", "second", "element"}

// decodeDiffInput returns the input of the place where of in.
func decodeDiffInput(where, in string) string {
	switch where {
	case "second":
		return "1\n" + in
	case "element":
		return "[0," + in + "]"
	}
	return in
}

// decodeDiffRun decodes the value of the input at the place where by decode, after what is before it, which it reads
// by the other functions, and returns the error of the value, and what the decoder reads after it, if the value has
// no error.
func decodeDiffRun(where string, decode func() error, readValue func() error, readToken func() error) (error, error) {
	switch where {
	case "second":
		if err := readValue(); err != nil {
			return err, nil
		}
	case "element":
		if err := readToken(); err != nil {
			return err, nil
		}
		if err := readValue(); err != nil {
			return err, nil
		}
	}
	if err := decode(); err != nil {
		return err, nil
	}
	if where == "element" {
		return nil, readToken()
	}
	return nil, readValue()
}

func TestUnmarshalDecodeMatchesStd(t *testing.T) {
	for _, c := range append(unmarshalDiffCases(), unmarshalDiffStructCases()...) {
		for i, in := range spacedInputs(c.in) {
			for _, where := range decodeDiffWhere {
				readDiff(t, fmt.Sprintf("%s/%d/%s", c.name, i, where), func(t *testing.T, n int) {
					input := decodeDiffInput(where, in)
					ours, theirs := c.out(), c.out()
					dec := jsontext.NewDecoder(newChunkReader(input, n), c.opts.ours()...)
					err, after := decodeDiffRun(where,
						func() error { return json.UnmarshalDecode(dec, ours, c.opts.ours()...) },
						func() error { _, err := dec.ReadValue(); return err },
						func() error { _, err := dec.ReadToken(); return err })
					stdDec := stdjsontext.NewDecoder(newChunkReader(input, n), c.opts.std()...)
					stdErr, stdAfter := decodeDiffRun(where,
						func() error { return stdjson.UnmarshalDecode(stdDec, theirs, c.opts.std()...) },
						func() error { _, err := stdDec.ReadValue(); return err },
						func() error { _, err := stdDec.ReadToken(); return err })
					if got, want := describeReadError(c.name, err), describeReadError(c.name, stdErr); got != want {
						t.Fatalf("error:\n got %s\nwant %s", got, want)
					}
					if got, want := describeUnmarshalError(after), describeUnmarshalError(stdAfter); got != want {
						t.Fatalf("after the value:\n got %s\nwant %s", got, want)
					}
					if err == nil && !c.errorOnly && !reflect.DeepEqual(ours, theirs) {
						t.Fatalf("value:\n got %#v\nwant %#v", reflect.ValueOf(ours).Elem(), reflect.ValueOf(theirs).Elem())
					}
				})
			}
		}
	}
}

func TestUnmarshalDecodeMethodsMatchStd(t *testing.T) {
	for _, c := range umCases() {
		for i, in := range spacedInputs(c.in) {
			readDiff(t, fmt.Sprintf("%s/%d", c.name, i), func(t *testing.T, n int) {
				ours, theirs := c.out(), c.out()
				err := json.UnmarshalDecode(jsontext.NewDecoder(newChunkReader(in, n)), ours, c.opts.ours...)
				stdErr := stdjson.UnmarshalDecode(stdjsontext.NewDecoder(newChunkReader(in, n)), theirs, c.opts.std...)
				if got, want := describeUnmarshalError(err), describeUnmarshalError(stdErr); got != want {
					t.Fatalf("error:\n got %s\nwant %s", got, want)
				}
				if err == nil && !reflect.DeepEqual(ours, theirs) {
					t.Fatalf("value:\n got %#v\nwant %#v", reflect.ValueOf(ours).Elem(), reflect.ValueOf(theirs).Elem())
				}
			})
		}
	}
}

func TestUnmarshalDecodeFromMatchesStd(t *testing.T) {
	for _, c := range ufCases() {
		for i, in := range spacedInputs(c.in) {
			readDiff(t, fmt.Sprintf("%s/%d", c.name, i), func(t *testing.T, n int) {
				ours, theirs := c.ours(), c.theirs()
				err := json.UnmarshalDecode(jsontext.NewDecoder(newChunkReader(in, n)), ours, c.opts.ours...)
				stdErr := stdjson.UnmarshalDecode(stdjsontext.NewDecoder(newChunkReader(in, n)), theirs, c.opts.std...)
				want := strings.ReplaceAll(describeUnmarshalError(stdErr), "ufStd", "ufOurs")
				if got := describeUnmarshalError(err); got != want {
					t.Fatalf("error:\n got %s\nwant %s", got, want)
				}
				if got, want := ufText(ours), ufText(theirs); err == nil && got != want {
					t.Fatalf("value:\n got %s\nwant %s", got, want)
				}
			})
		}
	}
}

// The values of a stream of lines, decoded one after another, as a program reads a stream of records.
func TestUnmarshalDecodeLinesMatchStd(t *testing.T) {
	for _, c := range append(unmarshalDiffCases(), unmarshalDiffStructCases()...) {
		readDiff(t, c.name, func(t *testing.T, n int) {
			in := c.in + "\n" + c.in + "\n" + c.in
			dec := jsontext.NewDecoder(newChunkReader(in, n), c.opts.ours()...)
			stdDec := stdjsontext.NewDecoder(newChunkReader(in, n), c.opts.std()...)
			for i := 0; i < 4; i++ {
				ours, theirs := c.out(), c.out()
				err := json.UnmarshalDecode(dec, ours, c.opts.ours()...)
				stdErr := stdjson.UnmarshalDecode(stdDec, theirs, c.opts.std()...)
				if got, want := describeReadError(c.name, err), describeReadError(c.name, stdErr); got != want {
					t.Fatalf("value %d: error:\n got %s\nwant %s", i, got, want)
				}
				if err != nil {
					return
				}
				if !c.errorOnly && !reflect.DeepEqual(ours, theirs) {
					t.Fatalf("value %d:\n got %#v\nwant %#v", i, reflect.ValueOf(ours).Elem(), reflect.ValueOf(theirs).Elem())
				}
			}
		})
	}
}

// udElems and udStdElems decode an array by UnmarshalDecode of each element, from the decoder which their method is
// given, into a []any.
type udElems struct{ V []any }

func (u *udElems) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != ']' {
		var v any
		if err := json.UnmarshalDecode(dec, &v); err != nil {
			return err
		}
		u.V = append(u.V, v)
	}
	_, err := dec.ReadToken()
	return err
}

type udStdElems struct{ V []any }

func (u *udStdElems) UnmarshalJSONFrom(dec *stdjsontext.Decoder) error {
	if _, err := dec.ReadToken(); err != nil {
		return err
	}
	for dec.PeekKind() != ']' {
		var v any
		if err := stdjson.UnmarshalDecode(dec, &v); err != nil {
			return err
		}
		u.V = append(u.V, v)
	}
	_, err := dec.ReadToken()
	return err
}

// A method which decodes the elements of its array by UnmarshalDecode, given a decoder by Unmarshal, UnmarshalRead and
// UnmarshalDecode.
func TestUnmarshalDecodeInMethodMatchesStd(t *testing.T) {
	long := strings.Repeat(`{"a":[1,"x",null]},`, 100)
	for _, in := range []string{
		`[1,"a",{"b":[true]}]`,
		`{"m":[1,2]}`,
		`[` + long + `2]`,
		`{"m":[` + long + `{"a":1 "b":2}]}`,
		`[1,"a\xff"]`,
		`[{"a":1,"a":2}]`,
		`[1,2`,
		`[[1,]]`,
	} {
		for _, n := range readSizes {
			for _, how := range []string{"Unmarshal", "UnmarshalRead", "UnmarshalDecode"} {
				t.Run(fmt.Sprintf("%.20s/reads-%d/%s", in, n, how), func(t *testing.T) {
					var ours struct{ M udElems }
					var theirs struct{ M udStdElems }
					input := in
					if strings.HasPrefix(in, "[") {
						input = `{"M":` + in + `}`
					} else {
						input = `{"M":[` + in + `]}`
					}
					var err, stdErr error
					switch how {
					case "Unmarshal":
						err = json.Unmarshal([]byte(input), &ours)
						stdErr = stdjson.Unmarshal([]byte(input), &theirs)
					case "UnmarshalRead":
						err = json.UnmarshalRead(newChunkReader(input, n), &ours)
						stdErr = stdjson.UnmarshalRead(newChunkReader(input, n), &theirs)
					default:
						err = json.UnmarshalDecode(jsontext.NewDecoder(newChunkReader(input, n)), &ours)
						stdErr = stdjson.UnmarshalDecode(stdjsontext.NewDecoder(newChunkReader(input, n)), &theirs)
					}
					want := strings.ReplaceAll(describeUnmarshalError(stdErr), "udStdElems", "udElems")
					if got := describeUnmarshalError(err); got != want {
						t.Fatalf("error:\n got %s\nwant %s", got, want)
					}
					if err == nil && !reflect.DeepEqual(ours.M.V, theirs.M.V) {
						t.Fatalf("value:\n got %#v\nwant %#v", ours.M.V, theirs.M.V)
					}
				})
			}
		}
	}
}

// failingAfterTokens reads three tokens of its value, and fails.
type failingAfterTokens struct{}

var errFailingAfterTokens = errors.New("failed after three tokens")

func (*failingAfterTokens) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	for range 3 {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
	}
	return errFailingAfterTokens
}

// The decoder after an error of its value: after a SemanticError, the decoder is after the value, or, if the rest of
// the value has an error, fails every read with it; after a syntax error, it is before the value, which it reads
// again; after the error of a method which read its value from a reader, it is before the value, or fails every read
// with the error if the value is no more in the buffer of the stream which read it.
func TestUnmarshalDecodeAfterError(t *testing.T) {
	invalidRest := `[1,2,"x",` + strings.Repeat("0,", 2000) + `}] 5`
	restErr := jsontext.NewDecoder(strings.NewReader(invalidRest)).SkipValue()
	for _, n := range readSizes {
		dec := jsontext.NewDecoder(newChunkReader(invalidRest, n))
		var serr *json.SemanticError
		if err := json.UnmarshalDecode(dec, new([]int)); !errors.As(err, &serr) {
			t.Fatalf("reads of %d: got %v, want a SemanticError", n, err)
		}
		for range 2 {
			var syntaxErr *jsontext.SyntacticError
			if _, err := dec.ReadToken(); !errors.As(err, &syntaxErr) || err.Error() != restErr.Error() {
				t.Fatalf("reads of %d: after the SemanticError of a value of an invalid rest: %v, want %v", n, err, restErr)
			}
		}
		dec = jsontext.NewDecoder(newChunkReader(`{"A":[1,2,3]} 5`, n))
		err := json.UnmarshalDecode(dec, new(struct{ A failingAfterTokens }))
		if !errors.Is(err, errFailingAfterTokens) {
			t.Fatalf("reads of %d: got %v", n, err)
		}
		next, rerr := dec.ReadValue()
		if n == 1 || n == 7 {
			// the method read its value from the reader: before the value, or failing every read
			if rerr != err && !(rerr == nil && string(next) == `{"A":[1,2,3]}`) {
				t.Fatalf("reads of %d: after the error of a method which read from the reader: %s, %v", n, next, rerr)
			}
		} else if rerr != nil || string(next) != "5" {
			// the method read its value from the input which the decoder read whole: after the value
			t.Fatalf("reads of %d: after the error of a method: %s, %v", n, next, rerr)
		}
		dec = jsontext.NewDecoder(newChunkReader(`{"A":"x","b":[1,2]} 3`, n))
		var v struct{ A int }
		if err := json.UnmarshalDecode(dec, &v); err == nil {
			t.Fatalf("reads of %d: no error", n)
		}
		var w int
		if err := json.UnmarshalDecode(dec, &w); err != nil || w != 3 {
			t.Fatalf("reads of %d: after the SemanticError: %v, %v", n, w, err)
		}
		if err := json.UnmarshalDecode(dec, &w); err != io.EOF {
			t.Fatalf("reads of %d: got %v, want io.EOF", n, err)
		}
		dec = jsontext.NewDecoder(newChunkReader("\"a\xff\" 3", n))
		var s string
		if err := json.UnmarshalDecode(dec, &s); err == nil {
			t.Fatalf("reads of %d: no error", n)
		}
		if err := json.UnmarshalDecode(dec, &s, jsontext.AllowInvalidUTF8(true)); err != nil || s != "a\ufffd" {
			t.Fatalf("reads of %d: the value read again: %q, %v", n, s, err)
		}
	}
}

// decodingNested decodes its value by UnmarshalDecode into X, and then, if it fails, does what mode says: returns
// the error, declines the value or ignores the error.
type decodingNested struct {
	mode string
	X    []int
}

func (m *decodingNested) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	err := json.UnmarshalDecode(dec, &m.X)
	switch {
	case err == nil:
		return nil
	case m.mode == "decline":
		return errors.ErrUnsupported
	case m.mode == "ignore":
		return nil
	}
	return err
}

// decodingNext reads its value, and, if it is 0, decodes the next one by UnmarshalDecode, whose error it ignores.
type decodingNext struct{}

func (*decodingNext) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	v, err := dec.ReadValue()
	if err != nil {
		return err
	}
	if string(v) == "0" {
		_ = json.UnmarshalDecode(dec, new([]int))
	}
	return nil
}

// A method of a stream whose UnmarshalDecode fails where the decoder of the method can be neither before the value
// nor after it: the call of the method fails, whatever the method returns, and no decoder reads an input which is
// not the one of the reader.
func TestUnmarshalDecodeInMethodAfterError(t *testing.T) {
	inputs := []string{
		// a SemanticError, whose rest has a syntax error
		`[1,"x",2}]`,
		`[1,"x",` + strings.Repeat("2,", 2000) + `2}]`,
		// a syntax error after the stream dropped the start of the value
		`[1,` + strings.Repeat("2,", 2000) + `2}]`,
	}
	for _, in := range inputs {
		for _, mode := range []string{"return", "decline", "ignore"} {
			for _, n := range readSizes {
				v := struct{ A decodingNested }{A: decodingNested{mode: mode}}
				if err := json.UnmarshalRead(newChunkReader(`{"A":`+in+`}`, n), &v); err == nil {
					t.Errorf("UnmarshalRead of %.20q, %s, reads of %d: no error", in, mode, n)
				}
				dec := jsontext.NewDecoder(newChunkReader(in+" 7", n))
				m := decodingNested{mode: mode}
				if err := json.UnmarshalDecode(dec, &m); err == nil {
					t.Errorf("UnmarshalDecode of %.20q, %s, reads of %d: no error", in, mode, n)
				}
				if next, err := dec.ReadValue(); err == nil {
					t.Errorf("UnmarshalDecode of %.20q, %s, reads of %d: the decoder read %.20q after it", in, mode, n, next)
				}
			}
		}
		for _, n := range readSizes {
			var v []decodingNext
			if err := json.UnmarshalRead(newChunkReader(`[0,`+in+`,7,8]`, n), &v); err == nil {
				t.Errorf("UnmarshalRead of %.20q after a method's value, reads of %d: no error", in, n)
			}
			dec := jsontext.NewDecoder(newChunkReader(`[0,`+in+`,7,8] 9`, n))
			if err := json.UnmarshalDecode(dec, &v); err == nil {
				t.Errorf("UnmarshalDecode of %.20q after a method's value, reads of %d: no error", in, n)
			}
		}
	}
}
