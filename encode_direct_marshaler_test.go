package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

// The types of this file are stored directly in an interface value: their data word is the value itself, not its
// address. Their methods are on the value, so the method set of the pointer to them has them too.

// directMarshalerTarget has the marshalers on the pointer, which directStruct promotes.
type directMarshalerTarget struct{ n int }

func (m *directMarshalerTarget) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprint(m.n)), nil }

// directStruct is a struct of one pointer.
type directStruct struct{ *directMarshalerTarget }

type directMap map[string]int

func (m directMap) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprint(len(m))), nil }

type directTextMap map[string]int

func (m directTextMap) MarshalText() ([]byte, error) { return []byte(fmt.Sprint("m", len(m))), nil }

type directFunc func() int

func (f directFunc) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprint(f())), nil }

// A pointer to such a type calls the marshaler with the value it points to, in a field of a struct as anywhere
// else: the method of the pointer is called with the pointer.
func TestEncodePointerToDirectMarshaler(t *testing.T) {
	s := &directStruct{&directMarshalerTarget{4}}
	m := &directMap{"a": 1, "b": 2}
	tm := &directTextMap{"a": 1}
	f := directFunc(func() int { return 9 })
	values := map[string]any{
		"field struct": struct {
			N int
			O *directStruct
		}{1, s},
		"field struct omitempty": struct {
			N int
			O *directStruct `json:",omitempty"`
		}{1, s},
		"field struct string": struct {
			O *directStruct `json:",string"`
		}{s},
		"first field struct": struct{ O *directStruct }{s},
		"field map": struct {
			N int
			O *directMap
		}{1, m},
		"field map omitempty": struct {
			N int
			O *directMap `json:",omitempty"`
		}{1, m},
		"field text map": struct {
			N int
			O *directTextMap
		}{1, tm},
		"field func": struct {
			N int
			O *directFunc
		}{1, &f},
		"nil fields": struct {
			S *directStruct
			M *directMap
			T *directTextMap
			F *directFunc
		}{},
		"nil fields omitempty": struct {
			S *directStruct  `json:",omitempty"`
			M *directMap     `json:",omitempty"`
			T *directTextMap `json:",omitempty"`
			F *directFunc    `json:",omitempty"`
		}{},
		"pointer to struct":         &struct{ O *directStruct }{s},
		"top level":                 s,
		"top level pointer":         &s,
		"top level map":             m,
		"slice":                     []*directStruct{s, nil},
		"slice of maps":             []*directMap{m},
		"map value":                 map[string]*directStruct{"k": s},
		"map value of maps":         map[string]*directMap{"k": m},
		"interface":                 struct{ A any }{s},
		"map key of text map value": map[string]*directTextMap{"k": tm},
	}
	for name, v := range values {
		want, wantErr := stdjson.Marshal(v)
		if wantErr != nil {
			t.Fatalf("%s: %v", name, wantErr)
		}
		for i, opts := range [][]json.EncodeOptionFunc{nil, {json.UnorderedMap()}, {json.Colorize(&json.ColorScheme{})}} {
			got, err := json.MarshalWithOption(v, opts...)
			if err != nil || string(got) != string(want) {
				t.Errorf("%s ( options %d ): got %s %v, want %s", name, i, got, err, want)
			}
		}
		wantIndent, _ := stdjson.MarshalIndent(v, "", " ")
		if got, err := json.MarshalIndent(v, "", " "); err != nil || string(got) != string(wantIndent) {
			t.Errorf("%s with indent: got %s %v, want %s", name, got, err, wantIndent)
		}
	}
}
