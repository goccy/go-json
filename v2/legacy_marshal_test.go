package json_test

import (
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json when marshaling, on every version of Go, with the outputs which encoding/json/v2 of
// Go 1.27 writes ( see std_legacy_diff_test.go, which compares with it ).

type legacyMarshalCase struct {
	name string
	in   any
	opts []json.Options
	// want is the output, or "" for an error
	want string
}

func runLegacyMarshalCases(t *testing.T, cases []legacyMarshalCase) {
	t.Helper()
	for _, c := range cases {
		got, err := json.Marshal(c.in, append(c.opts, json.Deterministic(true))...)
		switch {
		case c.want == "" && err == nil:
			t.Errorf("%s: got %s, want an error", c.name, got)
		case c.want != "" && (err != nil || string(got) != c.want):
			t.Errorf("%s: got %s, %v, want %s", c.name, got, err, c.want)
		}
	}
}

func TestMarshalFormatDurationAsNano(t *testing.T) {
	nano := []json.Options{gojson.FormatDurationAsNano(true)}
	d := time.Duration(-3)
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"value", time.Duration(1500), nano, `1500`},
		{"value by default", time.Duration(1500), nil, ``},
		{"pointer", &d, nano, `-3`},
		{"stringified", time.Duration(1500), append(nano, json.StringifyNumbers(true)), `"1500"`},
		{"key", map[time.Duration]int{5: 1, -2: 2}, nano, `{"-2":2,"5":1}`},
		{"key by default", map[time.Duration]int{5: 1}, nil, ``},
		{"string option", struct {
			S time.Duration `json:",string"`
		}{2}, nano, `{"S":"2"}`},
		{"string option by default", struct {
			S time.Duration `json:",string"`
		}{2}, nil, ``},
		{"in an interface", []any{time.Duration(7)}, nano, `[7]`},
	})
}

func TestUnmarshalFormatDurationAsNano(t *testing.T) {
	nano := []json.Options{gojson.FormatDurationAsNano(true)}
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"value", `1500`, nano, newOf(time.Duration(0)), time.Duration(1500)},
		{"value by default", `1500`, nil, newOf(time.Duration(0)), nil},
		{"fraction", `1.5`, nano, newOf(time.Duration(0)), nil},
		{"string", `"1500"`, nano, newOf(time.Duration(0)), nil},
		{"stringified", `"1500"`, append(nano, json.StringifyNumbers(true)), newOf(time.Duration(0)), time.Duration(1500)},
		{"key", `{"5":1}`, nano, func() any { return new(map[time.Duration]int) }, map[time.Duration]int{5: 1}},
		{"string option", `{"S":"3"}`, nano, func() any {
			return new(struct {
				S time.Duration `json:",string"`
			})
		}, struct {
			S time.Duration `json:",string"`
		}{3}},
	})
}
