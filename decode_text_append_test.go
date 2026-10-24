package json_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

// textAppender appends to its text, which encoding.TextUnmarshaler allows, before it keeps a copy of it.
type textAppender struct{ S string }

func (t *textAppender) UnmarshalText(b []byte) error {
	b = append(b, "XXXXXXXX"...)
	t.S = string(b[:len(b)-8])
	return nil
}

// textAppenderKey is textAppender for a key of a map.
type textAppenderKey string

func (k *textAppenderKey) UnmarshalText(b []byte) error {
	b = append(b, "XXXXXXXX"...)
	*k = textAppenderKey(b[:len(b)-8])
	return nil
}

type textAppenderFields struct {
	A textAppender
	B string
	N int
}

// An UnmarshalText method which appends to its text doesn't write into the input after the text: the values after it
// are decoded as encoding/json of Go 1.27 decodes them. The encoding/json of the Go versions before writes into its
// own input then, which it panics for.
func TestUnmarshalTextAppendLeavesInput(t *testing.T) {
	for _, c := range []struct {
		name string
		in   string
		out  func() any
		want any
	}{
		{"value", `{"A":"abcdefghijklmnop","B":"bbbbbbbb","N":1}`, func() any { return new(textAppenderFields) },
			&textAppenderFields{A: textAppender{S: "abcdefghijklmnop"}, B: "bbbbbbbb", N: 1}},
		{"value not ASCII", `{"A":"abcé","B":"b","N":1}`, func() any { return new(textAppenderFields) },
			&textAppenderFields{A: textAppender{S: "abcé"}, B: "b", N: 1}},
		{"value escaped", `{"A":"a\nb","B":"b","N":1}`, func() any { return new(textAppenderFields) },
			&textAppenderFields{A: textAppender{S: "a\nb"}, B: "b", N: 1}},
		{"key", `{"ké":1,"b":2,"c":3}`, func() any { return new(map[textAppenderKey]int) },
			&map[textAppenderKey]int{"ké": 1, "b": 2, "c": 3}},
	} {
		got := c.out()
		if err := json.Unmarshal([]byte(c.in), got); err != nil || !reflect.DeepEqual(got, c.want) {
			t.Fatalf("%s: got %+v, %v, want %+v", c.name, reflect.ValueOf(got).Elem(), err, reflect.ValueOf(c.want).Elem())
		}
		streamed := c.out()
		if err := json.NewDecoder(strings.NewReader(c.in)).Decode(streamed); err != nil || !reflect.DeepEqual(streamed, c.want) {
			t.Fatalf("%s: Decoder got %+v, %v, want %+v", c.name, reflect.ValueOf(streamed).Elem(), err, reflect.ValueOf(c.want).Elem())
		}
	}
}
