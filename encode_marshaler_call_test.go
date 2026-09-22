package json_test

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"errors"
	"testing"

	"github.com/goccy/go-json"
)

// The methods of the marshalers are called directly by their code, with the data word of the interface value
// as the receiver. Every kind of a receiver must be encoded as encoding/json does.

type valueMarshaler struct{ N int }

func (v valueMarshaler) MarshalJSON() ([]byte, error) { return []byte(`{"value":1}`), nil }

type ptrMarshaler struct{ N int }

func (v *ptrMarshaler) MarshalJSON() ([]byte, error) { return []byte(`{"ptr":1}`), nil }

// a value of the size of a pointer whose method is on the pointer: it is copied before the call.
type wordPtrMarshaler int64

func (v *wordPtrMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"word"`), nil }

type mapMarshaler map[string]int

func (m mapMarshaler) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte(`"nil map"`), nil
	}
	return []byte(`"map"`), nil
}

type textMarshaler struct{ S string }

func (t textMarshaler) MarshalText() ([]byte, error) { return []byte("text:" + t.S), nil }

type ptrTextMarshaler struct{ S string }

func (t *ptrTextMarshaler) MarshalText() ([]byte, error) { return []byte("ptrtext:" + t.S), nil }

type failingMarshaler struct{}

func (failingMarshaler) MarshalJSON() ([]byte, error) { return nil, errors.New("fails") }

type contextMarshaler struct{}

func (contextMarshaler) MarshalJSON(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		return nil, errors.New("no context")
	}
	return []byte(`"context"`), nil
}

func TestMarshalerCalledDirectly(t *testing.T) {
	var nilPtr *ptrMarshaler
	var nilMap mapMarshaler
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"value receiver", valueMarshaler{1}},
		{"pointer to value receiver", &valueMarshaler{1}},
		{"pointer receiver", &ptrMarshaler{1}},
		{"pointer receiver on a value", ptrMarshaler{1}},
		{"nil pointer", nilPtr},
		{"word with pointer receiver", wordPtrMarshaler(1)},
		{"pointer to word", func() *wordPtrMarshaler { w := wordPtrMarshaler(1); return &w }()},
		{"map with value receiver", mapMarshaler{"a": 1}},
		{"nil map with value receiver", nilMap},
		{"text", textMarshaler{"a"}},
		{"pointer to text", &textMarshaler{"a"}},
		{"pointer receiver text", &ptrTextMarshaler{"a"}},
		{"pointer receiver text on a value", ptrTextMarshaler{"a"}},
		{"fields", struct {
			V  valueMarshaler
			P  ptrMarshaler
			PP *ptrMarshaler
			NP *ptrMarshaler
			W  wordPtrMarshaler
			M  mapMarshaler
			NM mapMarshaler
			T  textMarshaler
			PT ptrTextMarshaler
		}{PP: &ptrMarshaler{1}, M: mapMarshaler{}}},
		{"omitempty", struct {
			NP *ptrMarshaler  `json:",omitempty"`
			PP *ptrMarshaler  `json:",omitempty"`
			V  valueMarshaler `json:",omitempty"`
		}{PP: &ptrMarshaler{1}}},
		{"slice", []valueMarshaler{{1}, {2}}},
		{"slice of pointer receiver", []ptrMarshaler{{1}, {2}}},
		{"map value", map[string]ptrMarshaler{"a": {1}}},
		{"interface", []interface{}{valueMarshaler{1}, &ptrMarshaler{1}, nilPtr, textMarshaler{"x"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			expected, err := stdjson.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expected, got) {
				t.Fatalf("expected %s but got %s", expected, got)
			}
			expected, err = stdjson.MarshalIndent(test.v, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			got, err = json.MarshalIndent(test.v, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(expected, got) {
				t.Fatalf("indent: expected %s but got %s", expected, got)
			}
		})
	}
}

func TestMarshalerCalledDirectlyError(t *testing.T) {
	_, err := json.Marshal(struct{ F failingMarshaler }{})
	var merr *json.MarshalerError
	if !errors.As(err, &merr) {
		t.Fatalf("expected a MarshalerError but got %v", err)
	}
	if merr.Type.Name() != "failingMarshaler" {
		t.Fatalf("unexpected type in the error: %v", merr.Type)
	}
	b, err := json.MarshalContext(context.Background(), []contextMarshaler{{}})
	if err != nil || string(b) != `["context"]` {
		t.Fatalf("context marshaler: %s, %v", b, err)
	}
	b, err = json.Marshal([]contextMarshaler{{}})
	if err != nil || string(b) != `["context"]` {
		t.Fatalf("context marshaler without a context: %s, %v", b, err)
	}
}
