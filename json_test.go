package json_test

import (
	"bytes"
	"testing"

	"github.com/goccy/go-json"
)

var validTests = []struct {
	data string
	ok   bool
}{
	{`foo`, false},
	{`}{`, false},
	{`{]`, false},
	{`{}`, true},
	{`{"foo":"bar"}`, true},
	{`{"foo":"bar","bar":{"baz":["qux"]}}`, true},
}

func TestValid(t *testing.T) {
	for _, tt := range validTests {
		if ok := json.Valid([]byte(tt.data)); ok != tt.ok {
			t.Errorf("Valid(%#q) = %v, want %v", tt.data, ok, tt.ok)
		}
	}
}

type example struct {
	compact string
	indent  string
}

var examples = []example{
	{`1`, `1`},
	{`{}`, `{}`},
	{`[]`, `[]`},
	{`{"":2}`, "{\n\t\"\": 2\n}"},
	{`[3]`, "[\n\t3\n]"},
	{`[1,2,3]`, "[\n\t1,\n\t2,\n\t3\n]"},
	{`{"x":1}`, "{\n\t\"x\": 1\n}"},
	{ex1, ex1i},
	{"{\"\":\"<>&\u2028\u2029\"}", "{\n\t\"\": \"<>&\u2028\u2029\"\n}"}, // See golang.org/issue/34070
}

var ex1 = `[true,false,null,"x",1,1.5,0,-5e+2]`

var ex1i = `[
	true,
	false,
	null,
	"x",
	1,
	1.5,
	0,
	-5e+2
]`

func TestCompact(t *testing.T) {
	var buf bytes.Buffer
	for _, tt := range examples {
		buf.Reset()
		t.Log("src = ", tt.compact)
		if err := json.Compact(&buf, []byte(tt.compact)); err != nil {
			t.Errorf("Compact(%#q): %v", tt.compact, err)
		} else if s := buf.String(); s != tt.compact {
			t.Errorf("Compact(%#q) = %#q, want original", tt.compact, s)
		}

		buf.Reset()
		if err := json.Compact(&buf, []byte(tt.indent)); err != nil {
			t.Errorf("Compact(%#q): %v", tt.indent, err)
			continue
		} else if s := buf.String(); s != tt.compact {
			t.Errorf("Compact(%#q) = %#q, want %#q", tt.indent, s, tt.compact)
		}
	}
}

func TestCompactSeparators(t *testing.T) {
	// U+2028 and U+2029 should be escaped inside strings.
	// They should not appear outside strings.
	tests := []struct {
		in, compact string
	}{
		{"{\"\u2028\": 1}", "{\"\u2028\":1}"},
		{"{\"\u2029\" :2}", "{\"\u2029\":2}"},
	}
	for _, tt := range tests {
		var buf bytes.Buffer
		if err := json.Compact(&buf, []byte(tt.in)); err != nil {
			t.Errorf("Compact(%q): %v", tt.in, err)
		} else if s := buf.String(); s != tt.compact {
			t.Errorf("Compact(%q) = %q, want %q", tt.in, s, tt.compact)
		}
	}
}

func TestIndent(t *testing.T) {
	var buf bytes.Buffer
	for _, tt := range examples {
		buf.Reset()
		if err := json.Indent(&buf, []byte(tt.indent), "", "\t"); err != nil {
			t.Errorf("Indent(%#q): %v", tt.indent, err)
		} else if s := buf.String(); s != tt.indent {
			t.Errorf("Indent(%#q) = %#q, want original", tt.indent, s)
		}

		buf.Reset()
		if err := json.Indent(&buf, []byte(tt.compact), "", "\t"); err != nil {
			t.Errorf("Indent(%#q): %v", tt.compact, err)
			continue
		} else if s := buf.String(); s != tt.indent {
			t.Errorf("Indent(%#q) = %#q, want %#q", tt.compact, s, tt.indent)
		}
	}
}
