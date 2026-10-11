package json_test

import (
	stdjson "encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// A stream with a string which is not closed is not read past the line after it ( issue 701 ): a control character in a
// string fails within the read which brought it, as encoding/json, which reads a value while it checks it, does.

// readUntil is a reader of data, which hides its length, and tells the bytes read when the read which brought the byte
// at the position at returned.
type readUntil struct {
	data        string
	n           int
	at, reached int
}

func (r *readUntil) Read(p []byte) (int, error) {
	if r.n == len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.n:])
	if r.n <= r.at && r.at < r.n+n {
		r.reached = r.n + n
	}
	r.n += n
	return n, nil
}

type issue701Person struct {
	Name string `json:"name"`
}

type issue701Unmarshaler struct{}

func (*issue701Unmarshaler) UnmarshalJSON([]byte) error { return nil }

func TestDecodeStreamIssue701(t *testing.T) {
	rec := `{"name":"alice","age":30,"address":{"city":"Springfield"}}`
	lines := strings.Repeat(rec+"\n", 50000)
	targets := []struct {
		name string
		new  func() any
	}{
		{"struct", func() any { return new(issue701Person) }},
		{"any", func() any { return new(any) }},
		{"map", func() any { return new(map[string]any) }},
		{"raw message", func() any { return new(json.RawMessage) }},
		{"unmarshaler", func() any { return new(issue701Unmarshaler) }},
		{"another kind", func() any { return new(string) }},
		{"slice", func() any { return new([]issue701Person) }},
	}
	for _, in := range []struct {
		name   string
		stream string
		// bad is the position of the control character, or -1
		bad int
	}{
		{"issue", `{"name":"alice` + "\n" + lines, len(`{"name":"alice`)},
		{"array", `[{"name":"alice` + "\n" + lines, len(`[{"name":"alice`)},
		{"later record", rec + "\n" + `{"name":"bob` + "\n" + lines, len(rec + "\n" + `{"name":"bob`)},
		// a string which ends at the next quote, after which the grammar fails: it has no control character, and
		// the value is read until its brackets balance, here the end of the input, as master reads it
		{"minified", `{"name":"alice` + strings.Repeat(rec, 50000), -1},
	} {
		for _, target := range targets {
			r := &readUntil{data: in.stream, at: in.bad}
			dec := json.NewDecoder(r)
			// a type error of the first value goes on to the next, as encoding/json does
			var err error
			var typeErr *json.UnmarshalTypeError
			for i := 0; i < 2 && (err == nil || errors.As(err, &typeErr)); i++ {
				err = dec.Decode(target.new())
			}
			var syntax *json.SyntaxError
			if !errors.As(err, &syntax) {
				t.Errorf("%s into %s: got %v", in.name, target.name, err)
				continue
			}
			if in.bad >= 0 && r.n > r.reached {
				t.Errorf("%s into %s: %d bytes read, the control character came with the read up to %d", in.name, target.name, r.n, r.reached)
			}
		}
	}
	// encoding/json fails on these streams too
	if err := stdjson.NewDecoder(strings.NewReader(`{"name":"alice` + strings.Repeat(rec, 10))).Decode(new(issue701Person)); err == nil {
		t.Fatal("encoding/json decoded the minified stream")
	}
}
