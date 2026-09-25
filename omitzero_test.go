package json

import (
	"bytes"
	"testing"
	"time"
)

// TestOmitZero tests the omitzero struct tag functionality
func TestOmitZero(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		// Basic primitive fields - omitted when zero
		{
			name: "int zero",
			value: struct {
				X int `json:"x,omitzero"`
			}{X: 0},
			expected: "{}",
		},
		{
			name: "int non-zero",
			value: struct {
				X int `json:"x,omitzero"`
			}{X: 42},
			expected: `{"x":42}`,
		},
		{
			name: "string zero",
			value: struct {
				X string `json:"x,omitzero"`
			}{X: ""},
			expected: "{}",
		},
		{
			name: "string non-zero",
			value: struct {
				X string `json:"x,omitzero"`
			}{X: "hello"},
			expected: `{"x":"hello"}`,
		},
		{
			name: "bool zero",
			value: struct {
				X bool `json:"x,omitzero"`
			}{X: false},
			expected: "{}",
		},
		{
			name: "bool non-zero",
			value: struct {
				X bool `json:"x,omitzero"`
			}{X: true},
			expected: `{"x":true}`,
		},
		{
			name: "float64 zero",
			value: struct {
				X float64 `json:"x,omitzero"`
			}{X: 0.0},
			expected: "{}",
		},
		{
			name: "float64 non-zero",
			value: struct {
				X float64 `json:"x,omitzero"`
			}{X: 3.14},
			expected: `{"x":3.14}`,
		},

		// Pointer fields - nil is zero, but non-nil pointing to zero is NOT zero
		{
			name: "pointer nil",
			value: struct {
				X *int `json:"x,omitzero"`
				Y int  `json:"y,omitzero"`
			}{X: nil, Y: 0},
			expected: "{}",
		},
		{
			name: "pointer to zero",
			value: struct {
				X *int `json:"x,omitzero"`
				Y int  `json:"y,omitzero"`
			}{X: intPtr(0), Y: 0},
			expected: `{"x":0}`,
		},
		{
			name: "pointer to non-zero",
			value: struct {
				X *int `json:"x,omitzero"`
				Y int  `json:"y,omitzero"`
			}{X: intPtr(42), Y: 0},
			expected: `{"x":42}`,
		},

		// Slice fields - nil is zero, but empty slice is NOT zero (key difference from omitempty)
		{
			name: "slice nil",
			value: struct {
				X []int `json:"x,omitzero"`
				Y int   `json:"y,omitzero"`
			}{X: nil, Y: 0},
			expected: "{}",
		},
		{
			name: "slice empty",
			value: struct {
				X []int `json:"x,omitzero"`
				Y int   `json:"y,omitzero"`
			}{X: []int{}, Y: 0},
			expected: `{"x":[]}`,
		},
		{
			name: "slice with values",
			value: struct {
				X []int `json:"x,omitzero"`
				Y int   `json:"y,omitzero"`
			}{X: []int{1, 2, 3}, Y: 0},
			expected: `{"x":[1,2,3]}`,
		},

		// Comparison: omitempty vs omitzero for empty slice
		{
			name: "omitempty with empty slice",
			value: struct {
				X []int `json:"x,omitempty"`
				Y int   `json:"y,omitempty"`
			}{X: []int{}, Y: 0},
			expected: "{}",
		},
		{
			name: "omitzero with empty slice",
			value: struct {
				X []int `json:"x,omitzero"`
				Y int   `json:"y,omitzero"`
			}{X: []int{}, Y: 0},
			expected: `{"x":[]}`,
		},

		// Mixed tags
		{
			name: "mixed omitempty and omitzero",
			value: struct {
				A int   `json:"a,omitempty"`
				B int   `json:"b,omitzero"`
				C []int `json:"c,omitempty"`
				D []int `json:"d,omitzero"`
			}{
				A: 0,
				B: 0,
				C: []int{},
				D: []int{},
			},
			expected: `{"d":[]}`,
		},

		// Multiple fields with omitzero
		{
			name: "multiple fields omitzero",
			value: struct {
				X int    `json:"x,omitzero"`
				Y string `json:"y,omitzero"`
				Z bool   `json:"z,omitzero"`
			}{
				X: 0,
				Y: "",
				Z: false,
			},
			expected: "{}",
		},
		{
			name: "multiple fields omitzero mixed",
			value: struct {
				X int    `json:"x,omitzero"`
				Y string `json:"y,omitzero"`
				Z bool   `json:"z,omitzero"`
			}{
				X: 42,
				Y: "hello",
				Z: true,
			},
			expected: `{"x":42,"y":"hello","z":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			if !bytes.Equal(b, []byte(tt.expected)) {
				t.Errorf("expected %s, got %s", tt.expected, string(b))
			}
		})
	}
}

// TestOmitZeroUnmarshal tests that unmarshal still works correctly with omitzero fields
func TestOmitZeroUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:  "unmarshal with omitzero field",
			input: `{"x":42}`,
			expected: struct {
				X int `json:"x,omitzero"`
			}{X: 42},
		},
		{
			name:  "unmarshal omitted field",
			input: `{}`,
			expected: struct {
				X int `json:"x,omitzero"`
			}{X: 0},
		},
		{
			name:  "unmarshal empty slice",
			input: `{"x":[]}`,
			expected: struct {
				X []int `json:"x,omitzero"`
			}{X: []int{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new instance of the expected type
			var result interface{}
			switch exp := tt.expected.(type) {
			case struct {
				X int `json:"x,omitzero"`
			}:
				var r struct {
					X int `json:"x,omitzero"`
				}
				if err := Unmarshal([]byte(tt.input), &r); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if r.X != exp.X {
					t.Errorf("expected X=%d, got X=%d", exp.X, r.X)
				}
				return
			case struct {
				X []int `json:"x,omitzero"`
			}:
				var r struct {
					X []int `json:"x,omitzero"`
				}
				if err := Unmarshal([]byte(tt.input), &r); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if len(r.X) != len(exp.X) {
					t.Errorf("expected X length=%d, got X length=%d", len(exp.X), len(r.X))
				}
				return
			}
			_ = result
		})
	}
}

// TestOmitZeroDirectInterface covers pointer-shaped structs (a single
// pointer-kind field) marshaled by value: the runtime stores such structs
// directly in the interface data word, so the head opcode must not treat the
// slot as a struct base address.
func TestOmitZeroDirectInterface(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name: "map nil",
			value: struct {
				X map[string]int `json:"x,omitzero"`
			}{X: nil},
			expected: "{}",
		},
		{
			name: "map empty non-nil",
			value: struct {
				X map[string]int `json:"x,omitzero"`
			}{X: map[string]int{}},
			expected: `{"x":{}}`,
		},
		{
			name: "map non-empty",
			value: struct {
				X map[string]int `json:"x,omitzero"`
			}{X: map[string]int{"a": 1}},
			expected: `{"x":{"a":1}}`,
		},
		{
			name: "slice empty non-nil",
			value: struct {
				X []int `json:"x,omitzero"`
			}{X: []int{}},
			expected: `{"x":[]}`,
		},
		{
			name: "slice non-empty",
			value: struct {
				X []int `json:"x,omitzero"`
			}{X: []int{1, 2}},
			expected: `{"x":[1,2]}`,
		},
		{
			name: "ptr to zero",
			value: struct {
				X *int `json:"x,omitzero"`
			}{X: intPtr(0)},
			expected: `{"x":0}`,
		},
		{
			name: "ptr to non-zero",
			value: struct {
				X *int `json:"x,omitzero"`
			}{X: intPtr(7)},
			expected: `{"x":7}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			if !bytes.Equal(b, []byte(tt.expected)) {
				t.Errorf("expected %s, got %s", tt.expected, string(b))
			}
		})
	}
}

// TestOmitZeroPtrHead covers marshaling a struct pointer whose fields use omitzero.
func TestOmitZeroPtrHead(t *testing.T) {
	type T struct {
		X int    `json:"x,omitzero"`
		Y string `json:"y,omitzero"`
	}
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{name: "nil ptr", value: (*T)(nil), expected: "null"},
		{name: "ptr to zero struct", value: &T{}, expected: "{}"},
		{name: "ptr to non-zero", value: &T{X: 5, Y: "hi"}, expected: `{"x":5,"y":"hi"}`},
		{name: "ptr mixed", value: &T{X: 0, Y: "hi"}, expected: `{"y":"hi"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			if !bytes.Equal(b, []byte(tt.expected)) {
				t.Errorf("expected %s, got %s", tt.expected, string(b))
			}
		})
	}
}

// customZeroer implements IsZero() with semantics that deliberately disagree
// with reflect.Value.IsZero, proving the custom method is preferred.
type customZeroer struct {
	zero bool
}

func (c customZeroer) IsZero() bool { return c.zero }

// TestOmitZeroIsZeroer covers fields whose type implements IsZero() bool.
func TestOmitZeroIsZeroer(t *testing.T) {
	cases := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name: "custom zeroer reports zero",
			value: struct {
				X customZeroer `json:"x,omitzero"`
				Y int          `json:"y,omitzero"`
			}{X: customZeroer{zero: true}, Y: 0},
			expected: "{}",
		},
		{
			name: "custom zeroer reports non-zero",
			value: struct {
				X customZeroer `json:"x,omitzero"`
				Y int          `json:"y,omitzero"`
			}{X: customZeroer{zero: false}, Y: 0},
			expected: `{"x":{}}`,
		},
		{
			name: "time.Time zero value",
			value: struct {
				X time.Time `json:"x,omitzero"`
				Y int       `json:"y,omitzero"`
			}{X: time.Time{}, Y: 0},
			expected: "{}",
		},
		{
			name: "time.Time non-zero",
			value: struct {
				X time.Time `json:"x,omitzero"`
				Y int       `json:"y,omitzero"`
			}{X: time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC), Y: 0},
			expected: `{"x":"2026-08-29T12:00:00Z"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			b, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			if !bytes.Equal(b, []byte(tt.expected)) {
				t.Errorf("expected %s, got %s", tt.expected, string(b))
			}
		})
	}
}

// TestOmitZeroMapFields covers map-kind includes through every opcode path:
// pointer-marshal (indirect head) and non-head fields, where the map value
// must be stored rather than its address.
func TestOmitZeroMapFields(t *testing.T) {
	type singleMap struct {
		X map[string]int `json:"x,omitzero"`
	}
	type multiField struct {
		A int            `json:"a,omitzero"`
		X map[string]int `json:"x,omitzero"`
	}
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{name: "ptr marshal nil map", value: &singleMap{}, expected: "{}"},
		{name: "ptr marshal empty map", value: &singleMap{X: map[string]int{}}, expected: `{"x":{}}`},
		{name: "ptr marshal non-empty map", value: &singleMap{X: map[string]int{"a": 1}}, expected: `{"x":{"a":1}}`},
		{name: "field nil map", value: multiField{A: 1}, expected: `{"a":1}`},
		{name: "field empty map", value: multiField{A: 1, X: map[string]int{}}, expected: `{"a":1,"x":{}}`},
		{name: "field non-empty map", value: multiField{A: 1, X: map[string]int{"k": 2}}, expected: `{"a":1,"x":{"k":2}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := Marshal(tt.value)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}
			if !bytes.Equal(b, []byte(tt.expected)) {
				t.Errorf("expected %s, got %s", tt.expected, string(b))
			}
		})
	}
}

// Benchmark omitzero
func BenchmarkOmitZero(b *testing.B) {
	type testStruct struct {
		X int    `json:"x,omitzero"`
		Y string `json:"y,omitzero"`
		Z []int  `json:"z,omitzero"`
		W bool   `json:"w,omitzero"`
	}

	value := testStruct{
		X: 0,
		Y: "",
		Z: nil,
		W: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(value)
		if err != nil {
			b.Fatalf("failed to marshal: %v", err)
		}
	}
}

// Helper function to create an int pointer
func intPtr(v int) *int {
	return &v
}
