package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

type unsupportedFuncMarshaler func()

func (unsupportedFuncMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"f"`), nil }

// A value of a type which has no JSON representation, a channel, a function or a complex number, fails only when it
// is encoded, as encoding/json does: a type which has such a type in it encodes the values which have none.
func TestEncodeUnsupportedTypeWhenEncoded(t *testing.T) {
	ch := make(chan int)
	f := func() {}
	values := map[string]any{
		"nil pointer to a channel": struct{ P *chan int }{},
		"pointer to a channel":     struct{ P *chan int }{&ch},
		"empty slice of functions": []func(){},
		"nil slice of functions":   ([]func())(nil),
		"slice of a function":      []func(){f},
		"empty map of channels":    map[string]chan int{},
		"map of a channel":         map[string]chan int{"a": ch},
		"array of no functions":    [0]func(){},
		"array of a function":      [1]func(){},
		"function omitted by omitzero": struct {
			F func() `json:",omitzero"`
			A int
		}{},
		"function with omitempty": struct {
			F func() `json:",omitempty"`
		}{},
		"complex number": struct{ C complex128 }{1},
		"complex number omitted": struct {
			C complex128 `json:",omitzero"`
		}{},
		"channel":                                             ch,
		"pointer to a channel at first":                       &ch,
		"nil pointer at first":                                (*chan int)(nil),
		"function with a marshaler":                           struct{ F unsupportedFuncMarshaler }{},
		"nil function in an interface value":                  []any{(func())(nil)},
		"nil channel in an interface value":                   map[string]any{"a": (chan int)(nil)},
		"nil function with a marshaler in an interface value": []any{unsupportedFuncMarshaler(nil)},
	}
	for name, v := range values {
		t.Run(name, func(t *testing.T) {
			want, wantErr := stdjson.Marshal(v)
			got, err := json.Marshal(v)
			if string(got) != string(want) || (err == nil) != (wantErr == nil) {
				t.Errorf("got %s, %v, want %s, %v", got, err, want, wantErr)
			}
			want, wantErr = stdjson.MarshalIndent(v, "", " ")
			got, err = json.MarshalIndent(v, "", " ")
			if string(got) != string(want) || (err == nil) != (wantErr == nil) {
				t.Errorf("indented: got %s, %v, want %s, %v", got, err, want, wantErr)
			}
		})
	}
}
