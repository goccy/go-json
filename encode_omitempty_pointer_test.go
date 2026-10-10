package json_test

import (
	"testing"

	"github.com/goccy/go-json"
)

type omitEmptyNamedPointer *int

// omitempty omits a field of a pointer type only if the pointer is nil, as encoding/json does: a pointer to a nil
// pointer or to a nil map is written, as null.
func TestEncodeOmitEmptyPointerToNilValue(t *testing.T) {
	var nilPointer *int
	var nilPointerPointer **int
	nilNamed := omitEmptyNamedPointer(nil)
	var nilMap map[string]int
	one := 1
	pointer := &one
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"pointer to a nil pointer", struct {
			A **int `json:",omitempty"`
		}{&nilPointer}, `{"A":null}`},
		{"pointer to a pointer to a nil pointer", struct {
			A ***int `json:",omitempty"`
		}{&nilPointerPointer}, `{"A":null}`},
		{"pointer to a nil named pointer", struct {
			A *omitEmptyNamedPointer `json:",omitempty"`
		}{&nilNamed}, `{"A":null}`},
		{"pointer to a nil map", struct {
			A *map[string]int `json:",omitempty"`
		}{&nilMap}, `{"A":null}`},
		{"between fields", struct {
			A int
			B **int `json:",omitempty"`
			C int
		}{1, &nilPointer, 2}, `{"A":1,"B":null,"C":2}`},
		{"pointer to a pointer", struct {
			A **int `json:",omitempty"`
		}{&pointer}, `{"A":1}`},
		{"nil pointer to a pointer", struct {
			A **int `json:",omitempty"`
		}{}, `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			if err != nil || string(got) != tt.want {
				t.Errorf("got %s, %v, want %s", got, err, tt.want)
			}
		})
	}
}
