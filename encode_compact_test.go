package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// The inputs of Compact and Indent, and the outputs of MarshalJSON, valid and not.
var formatInputs = []string{
	`{}`, `[]`, `1`, `"a"`, `true`, `null`, ` { "a" : 1 , "b" : [ 1 , 2 ] } `, "{\"a\":1}\n", "[1,2]\n\t ",
	`01`, `1.`, `-`, `1e`, `1e+`, `.5`, `+1`, `0x1`, `1.5e10`, `-0`, `[01]`, `{"a":1.}`,
	`"\q"`, `"\u12zz"`, `"\u12"`, "\"a\x01b\"", "\"a\xffb\"", `"<&>"`, "\" \"", `"\/"`,
	`{"a":1,}`, `[1,]`, `{"a" 1}`, `{a:1}`, `[1 2]`, `{"a":1}}`, `{"a":1} x`, `tru`, `nul`, `falsey`, ``, ` `,
	`{"a":{"b":[1,{"c":null}]}}`, `[[[[]]]]`, `"unterminated`, `{"a":"b"`, `[`, `{`, `{"a":1,"a":2}`,
}

type rawOutput string

func (r rawOutput) MarshalJSON() ([]byte, error) { return []byte(r), nil }

// Compact, Indent and the outputs of MarshalJSON format what encoding/json formats, and reject what it rejects,
// whatever the version of Go is: the messages are the ones of the latest Go ( see encode_compact_go127_test.go ).
func TestFormatAsEncodingJSON(t *testing.T) {
	for _, in := range formatInputs {
		var got, want bytes.Buffer
		got.WriteString("x")
		want.WriteString("x")
		gerr := json.Compact(&got, []byte(in))
		werr := stdjson.Compact(&want, []byte(in))
		if (gerr == nil) != (werr == nil) || got.String() != want.String() {
			t.Errorf("Compact(%q): got %q, %v; want %q, %v", in, got.String(), gerr, want.String(), werr)
		}
		got.Reset()
		want.Reset()
		gerr = json.Indent(&got, []byte(in), ">", "\t")
		werr = stdjson.Indent(&want, []byte(in), ">", "\t")
		if (gerr == nil) != (werr == nil) || got.String() != want.String() {
			t.Errorf("Indent(%q): got %q, %v; want %q, %v", in, got.String(), gerr, want.String(), werr)
		}
		gout, gerr := json.Marshal(rawOutput(in))
		wout, werr := stdjson.Marshal(rawOutput(in))
		if (gerr == nil) != (werr == nil) || string(gout) != string(wout) {
			t.Errorf("Marshal of the output %q: got %q, %v; want %q, %v", in, gout, gerr, wout, werr)
		}
	}
}

// Compact and Indent append to what the buffer has.
func TestFormatAppendsToBuffer(t *testing.T) {
	var b bytes.Buffer
	b.WriteString("[1]")
	if err := json.Compact(&b, []byte(` { "a" : 1 } `)); err != nil {
		t.Fatal(err)
	}
	if err := json.Indent(&b, []byte(`[2]`), "", " "); err != nil {
		t.Fatal(err)
	}
	if want := "[1]{\"a\":1}[\n 2\n]"; b.String() != want {
		t.Fatalf("got %q, want %q", b.String(), want)
	}
}
