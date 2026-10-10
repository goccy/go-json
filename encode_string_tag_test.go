package json_test

import (
	"regexp"
	"testing"

	"github.com/goccy/go-json"
)

type stringTagNamedPointer *int

// The option ,string applies to a field of a string, a number or a bool, through one pointer, and not to the value
// of a pointer to a pointer, as encoding/json of Go 1.27 does.
func TestEncodeStringTagPointers(t *testing.T) {
	i, s := 5, "x"
	pi, ps := &i, &s
	named := stringTagNamedPointer(&i)
	var nilPointer *int
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"pointer", struct {
			A *int `json:",string"`
		}{&i}, `{"A":"5"}`},
		{"named pointer", struct {
			A stringTagNamedPointer `json:",string"`
		}{named}, `{"A":"5"}`},
		{"pointer to a pointer", struct {
			A **int `json:",string"`
		}{&pi}, `{"A":5}`},
		{"pointer to a pointer of a string", struct {
			A **string `json:",string"`
		}{&ps}, `{"A":"x"}`},
		{"pointer to a nil pointer", struct {
			A **int `json:",string"`
		}{&nilPointer}, `{"A":null}`},
		{"pointer to a named pointer", struct {
			A *stringTagNamedPointer `json:",string"`
		}{&named}, `{"A":5}`},
		{"omitempty", struct {
			A **int `json:",omitempty,string"`
			B int   `json:",string"`
		}{&pi, 3}, `{"A":5,"B":"3"}`},
	}
	colors := regexp.MustCompile("\x1b\\[[0-9;]*m")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.in)
			if err != nil || string(got) != tt.want {
				t.Errorf("got %s, %v, want %s", got, err, tt.want)
			}
			got, err = json.MarshalWithOption(tt.in, json.Colorize(json.DefaultColorScheme))
			if got := colors.ReplaceAll(got, nil); err != nil || string(got) != tt.want {
				t.Errorf("colorized: got %s, %v, want %s", got, err, tt.want)
			}
		})
	}
}
