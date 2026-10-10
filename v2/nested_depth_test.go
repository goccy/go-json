package json_test

import (
	"errors"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// deepCall opens 100 arrays, in which it writes the next deepCall by MarshalEncode, until Calls calls, or forever
// if Calls is 0.
type deepCall struct{ N, Calls int }

func (v deepCall) MarshalJSONTo(e *jsontext.Encoder) error {
	if v.Calls > 0 && v.N == v.Calls {
		return e.WriteToken(jsontext.Int(1))
	}
	for range 100 {
		if err := e.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}
	}
	if err := json.MarshalEncode(e, deepCall{v.N + 1, v.Calls}); err != nil {
		return err
	}
	for range 100 {
		if err := e.WriteToken(jsontext.EndArray); err != nil {
			return err
		}
	}
	return nil
}

// deepTokens writes 100 arrays by tokens, deepRaw by a raw value.
type (
	deepTokens struct{}
	deepRaw    struct{}
)

func (deepTokens) MarshalJSONTo(e *jsontext.Encoder) error {
	for range 100 {
		if err := e.WriteToken(jsontext.BeginArray); err != nil {
			return err
		}
	}
	for range 100 {
		if err := e.WriteToken(jsontext.EndArray); err != nil {
			return err
		}
	}
	return nil
}

func (deepRaw) MarshalJSONTo(e *jsontext.Encoder) error {
	raw := make([]byte, 0, 200)
	for range 100 {
		raw = append(raw, '[')
	}
	for range 100 {
		raw = append(raw, ']')
	}
	return e.WriteValue(raw)
}

// deepDeclined is declined by a function, which writes its default representation, an array.
type deepDeclined []any

func nestedIn(n int, v any) any {
	for range n {
		v = []any{v}
	}
	return v
}

// The levels which a method writes, by tokens, a raw value, or a call of MarshalEncode, and the default
// representation of a value which a function declines, are as deep as the place of the value: they fail past 10000
// levels, as encoding/json/v2 fails, also for a method which calls MarshalEncode for a new value of its own forever.
func TestMarshalDepthOfNestedEncodings(t *testing.T) {
	decline := json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, deepDeclined) error { return errors.ErrUnsupported }))
	tests := []struct {
		name string
		of   func(levels int) any // a value of the levels
		opts []json.Options
	}{
		{"calls of MarshalEncode", func(n int) any { return deepCall{Calls: n / 100} }, nil},
		{"tokens", func(n int) any { return nestedIn(n-100, deepTokens{}) }, nil},
		{"a raw value", func(n int) any { return nestedIn(n-100, deepRaw{}) }, nil},
		{"a declined value", func(n int) any { return nestedIn(10, deepDeclined{nestedIn(n-12, 1)}) }, []json.Options{decline}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := json.Marshal(tt.of(10000), tt.opts...); err != nil {
				t.Errorf("10000 levels: %v", err)
			}
			var serr *jsontext.SyntacticError
			if _, err := json.Marshal(tt.of(10100), tt.opts...); !errors.As(err, &serr) {
				t.Errorf("10100 levels: got %v, want the error of the max depth", err)
			}
		})
	}
	var serr *jsontext.SyntacticError
	if _, err := json.Marshal(deepCall{}); !errors.As(err, &serr) {
		t.Errorf("calls without end: got %v, want the error of the max depth", err)
	}
}
