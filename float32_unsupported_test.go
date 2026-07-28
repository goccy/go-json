package json_test

import (
	stdjson "encoding/json"
	"math"
	"testing"

	json "github.com/goccy/go-json"
)

// nonFiniteFloat32 shapes exercises every float32 opcode family (bare, pointer,
// struct head/field/end, omitempty, string tag, slice, map) with a non-finite value.
func nonFiniteFloat32Shapes(f float32) map[string]interface{} {
	p := f
	return map[string]interface{}{
		"bare":        f,
		"ptr":         &p,
		"structHead":  struct{ A float32 }{f},
		"structPtr":   struct{ A *float32 }{&p},
		"structField": struct {
			X int
			A float32
		}{1, f},
		"omitempty": struct {
			A float32 `json:",omitempty"`
		}{f},
		"string": struct {
			A float32 `json:",string"`
		}{f},
		"slice": []float32{f},
		"map":   map[string]float32{"k": f},
	}
}

// float32 Inf/NaN must return *UnsupportedValueError and emit no output, matching
// encoding/json and the existing float64 behaviour, across all four VM variants.
func TestMarshalFloat32NonFinite(t *testing.T) {
	marshalers := []struct {
		name string
		fn   func(interface{}) ([]byte, error)
	}{
		{"plain", func(v interface{}) ([]byte, error) { return json.Marshal(v) }},
		{"indent", func(v interface{}) ([]byte, error) { return json.MarshalIndent(v, "", "\t") }},
		{"color", func(v interface{}) ([]byte, error) {
			return json.MarshalWithOption(v, json.Colorize(json.DefaultColorScheme))
		}},
		{"colorIndent", func(v interface{}) ([]byte, error) {
			return json.MarshalIndentWithOption(v, "", "\t", json.Colorize(json.DefaultColorScheme))
		}},
	}
	values := map[string]float32{
		"+Inf": float32(math.Inf(1)),
		"-Inf": float32(math.Inf(-1)),
		"NaN":  float32(math.NaN()),
	}
	for vname, f := range values {
		for sname, shape := range nonFiniteFloat32Shapes(f) {
			std, stdErr := stdjson.Marshal(shape)
			for _, m := range marshalers {
				got, err := m.fn(shape)
				t.Logf("%s/%s/%s: goccy=(%q,%v) stdlib=(%q,%v)", vname, sname, m.name, got, err, std, stdErr)
				if err == nil {
					t.Errorf("%s/%s/%s: expected error, got output %q", vname, sname, m.name, got)
					continue
				}
				if _, ok := err.(*json.UnsupportedValueError); !ok {
					t.Errorf("%s/%s/%s: got %T, want *UnsupportedValueError", vname, sname, m.name, err)
				}
				if len(got) != 0 {
					t.Errorf("%s/%s/%s: expected no output on error, got %q", vname, sname, m.name, got)
				}
			}
		}
	}
}

// Finite float32 must still marshal byte-identically to encoding/json (no over-fix).
func TestMarshalFloat32FiniteUnchanged(t *testing.T) {
	values := []float32{0, 1.5, -1.5, 1e30, math.MaxFloat32, math.SmallestNonzeroFloat32}
	for _, f := range values {
		for sname, shape := range nonFiniteFloat32Shapes(f) {
			std, stdErr := stdjson.Marshal(shape)
			if stdErr != nil {
				t.Fatalf("stdlib failed on finite %v/%s: %v", f, sname, stdErr)
			}
			got, err := json.Marshal(shape)
			t.Logf("finite %v/%s: goccy=%q stdlib=%q", f, sname, got, std)
			if err != nil {
				t.Errorf("finite %v/%s: unexpected error %v", f, sname, err)
				continue
			}
			if string(got) != string(std) {
				t.Errorf("finite %v/%s: goccy=%q stdlib=%q", f, sname, got, std)
			}
		}
	}
}
