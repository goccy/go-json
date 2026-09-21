package json_test

import (
	"context"
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// The fields of an embedded struct are written as the fields of the struct which embeds it.

type embeddedTagTarget struct {
	Bar string `json:"bar"`
}

// omitempty of an embedded struct has nothing to omit.
type embeddedPtrWithOmitEmpty struct {
	Foo                string `json:"foo"`
	*embeddedTagTarget `json:",omitempty"`
}

type embeddedValueWithOmitEmpty struct {
	Foo               string `json:"foo"`
	embeddedTagTarget `json:",omitempty"`
}

// The name of an embedded struct is not a key, so it doesn't hide a field of the same name.
type EmbeddedSameName struct {
	EmbeddedSameName string `json:"EmbeddedSameName"`
	Other            int    `json:"Other"`
}

type embeddedSameNameHolder struct {
	EmbeddedSameName
	Name string `json:"Name"`
}

// All the fields of a struct embedded in itself are hidden by the same fields at the shallower depth.
type embeddedInItself struct {
	ID int `json:"id"`
	*embeddedInItself
}

type embeddedInItselfThroughA struct {
	A int `json:"a"`
	*embeddedInItselfThroughB
}

type embeddedInItselfThroughB struct {
	B int `json:"b"`
	*embeddedInItselfThroughA
}

// the field of the shallower depth hides the one of an embedded struct, which is not recursive.
type embeddedShadowedInner struct {
	ID int `json:"id"`
	X  int `json:"x"`
}

type embeddedShadowedOuter struct {
	ID int `json:"id"`
	*embeddedShadowedInner
}

func TestEncodeEmbeddedStruct(t *testing.T) {
	for _, test := range []struct {
		name string
		v    interface{}
	}{
		{"pointer with omitempty", embeddedPtrWithOmitEmpty{Foo: "f", embeddedTagTarget: &embeddedTagTarget{Bar: "b"}}},
		{"nil pointer with omitempty", embeddedPtrWithOmitEmpty{Foo: "f"}},
		{"value with omitempty", embeddedValueWithOmitEmpty{Foo: "f", embeddedTagTarget: embeddedTagTarget{Bar: "b"}}},
		{"zero value with omitempty", &embeddedValueWithOmitEmpty{}},
		{"field of the same name as the embedded struct", embeddedSameNameHolder{
			EmbeddedSameName: EmbeddedSameName{EmbeddedSameName: "inside", Other: 1}, Name: "outside",
		}},
		{"embedded in itself", embeddedInItself{ID: 99, embeddedInItself: &embeddedInItself{ID: 1}}},
		{"nil embedded in itself", &embeddedInItself{ID: 99}},
		{"embedded in itself through another struct", embeddedInItselfThroughA{
			A: 1, embeddedInItselfThroughB: &embeddedInItselfThroughB{B: 2, embeddedInItselfThroughA: &embeddedInItselfThroughA{A: 3}},
		}},
		{"hidden by the shallower field", embeddedShadowedOuter{ID: 99, embeddedShadowedInner: &embeddedShadowedInner{ID: 1, X: 2}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			expected, err := stdjson.Marshal(test.v)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				got, err := json.Marshal(test.v)
				if err != nil {
					t.Fatalf("call %d: %v", i, err)
				}
				if string(got) != string(expected) {
					t.Fatalf("call %d: expected %s but got %s", i, expected, got)
				}
			}
			expectedIndent, err := stdjson.MarshalIndent(test.v, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			gotIndent, err := json.MarshalIndent(test.v, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if string(gotIndent) != string(expectedIndent) {
				t.Fatalf("indent: expected %s but got %s", expectedIndent, gotIndent)
			}
		})
	}
}

// An array of no elements is always empty.
func TestEncodeOmitEmptyZeroLengthArray(t *testing.T) {
	v := struct {
		A [0]int     `json:"a,omitempty"`
		B [0]string  `json:"b,omitempty"`
		C [0]int     `json:"c"`
		D [1]int     `json:"d,omitempty"`
		E [0]f0Array `json:"e,omitempty"`
	}{}
	expected, err := stdjson.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(expected) {
		t.Fatalf("expected %s but got %s", expected, got)
	}
}

type f0Array [0]byte

func (f0Array) MarshalJSON() ([]byte, error) { return []byte(`"empty"`), nil }

// A number which a marshaler writes is valid JSON even if it is out of the range of float64.
type numberLiteralMarshaler struct{ literal string }

func (n numberLiteralMarshaler) MarshalJSON() ([]byte, error) { return []byte(n.literal), nil }

func TestEncodeNumberOutOfFloat64Range(t *testing.T) {
	for _, literal := range []string{"1e400", "-1e400", "1e-400", "123456789012345678901234567890123456789012345678901234567890", "1.5"} {
		for _, v := range []interface{}{
			numberLiteralMarshaler{literal},
			stdjson.RawMessage(literal),
			map[string]interface{}{"n": numberLiteralMarshaler{literal}},
		} {
			expected, err := stdjson.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(v)
			if err != nil {
				t.Fatalf("%s: %v", literal, err)
			}
			if string(got) != string(expected) {
				t.Fatalf("%s: expected %s but got %s", literal, expected, got)
			}
		}
	}
	// what is not a number is still an error.
	if _, err := json.Marshal(numberLiteralMarshaler{"1e"}); err == nil {
		t.Fatal("an invalid number must be an error")
	}
}

type contextValueMarshaler struct{}

type contextValueKey struct{}

func (contextValueMarshaler) MarshalJSON(ctx context.Context) ([]byte, error) {
	if ctx == nil {
		return []byte(`"nil context"`), nil
	}
	if v, ok := ctx.Value(contextValueKey{}).(string); ok {
		return []byte(`"` + v + `"`), nil
	}
	return []byte(`"no value"`), nil
}

// The context of a call must not be seen by the following calls, which reuse the same runtime context.
func TestEncodeContextIsNotKeptForNextCall(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextValueKey{}, "value")
	for i := 0; i < 3; i++ {
		got, err := json.MarshalContext(ctx, contextValueMarshaler{})
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != `"value"` {
			t.Fatalf("with context: got %s", got)
		}
		for _, marshal := range []func(interface{}) ([]byte, error){
			json.Marshal,
			func(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", " ") },
		} {
			got, err := marshal(contextValueMarshaler{})
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != `"no value"` {
				t.Fatalf("without context: got %s", got)
			}
		}
	}
}
