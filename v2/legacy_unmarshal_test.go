package json_test

import (
	"reflect"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json when unmarshaling, on every version of Go, with the values which encoding/json/v2 of
// Go 1.27 decodes ( see std_legacy_diff_test.go, which compares with it ).

type legacyUnmarshalCase struct {
	name     string
	in       string
	opts     []json.Options
	newValue func() any
	// want is the value which the pointer made by newValue points to after the decoding, or nil for an error
	want any
}

func newOf[T any](v T) func() any {
	return func() any {
		p := new(T)
		*p = v
		return p
	}
}

func runLegacyUnmarshalCases(t *testing.T, cases []legacyUnmarshalCase) {
	t.Helper()
	for _, c := range cases {
		v := c.newValue()
		err := json.Unmarshal([]byte(c.in), v, c.opts...)
		switch {
		case c.want == nil && err == nil:
			t.Errorf("%s: got %#v, want an error", c.name, reflect.ValueOf(v).Elem().Interface())
		case c.want != nil && err != nil:
			t.Errorf("%s: %v", c.name, err)
		case c.want != nil && !reflect.DeepEqual(reflect.ValueOf(v).Elem().Interface(), c.want):
			t.Errorf("%s: got %#v, want %#v", c.name, reflect.ValueOf(v).Elem().Interface(), c.want)
		}
	}
}

func TestUnmarshalArrayFromAnyLength(t *testing.T) {
	anyLength := []json.Options{gojson.UnmarshalArrayFromAnyLength(true)}
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"longer", `[1,2,3]`, anyLength, newOf([2]int{}), [2]int{1, 2}},
		{"longer by default", `[1,2,3]`, nil, newOf([2]int{}), nil},
		{"shorter", `[1]`, anyLength, newOf([3]int{7, 7, 7}), [3]int{1, 0, 0}},
		{"shorter by default", `[1]`, nil, newOf([3]int{}), nil},
		{"invalid UTF-8 past the length", "[1,2,\"\xff\"]", anyLength, newOf([2]int{}), nil},
		{"a name twice past the length", `[1,2,{"a":1,"a":2}]`, anyLength, newOf([2]int{}), nil},
		{"bytes longer", `"AQID"`, anyLength, newOf([2]byte{}), [2]byte{1, 2}},
		{"bytes shorter", `"AQID"`, anyLength, newOf([4]byte{7, 7, 7, 7}), [4]byte{1, 2, 3, 0}},
		{"bytes shorter by default", `"AQID"`, nil, newOf([4]byte{}), nil},
	})
}

func TestParseBytesWithLooseRFC4648(t *testing.T) {
	loose := []json.Options{gojson.ParseBytesWithLooseRFC4648(true)}
	in := `"AQ` + "\\r\\n" + `ID"`
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"line feed", in, loose, newOf([]byte(nil)), []byte{1, 2, 3}},
		{"line feed by default", in, nil, newOf([]byte(nil)), nil},
		{"line feed into an array", in, loose, newOf([3]byte{}), [3]byte{1, 2, 3}},
		{"wrong character", `"AQ*ID"`, loose, newOf([]byte(nil)), nil},
	})
}

func TestParseTimeWithLooseRFC3339(t *testing.T) {
	loose := []json.Options{gojson.ParseTimeWithLooseRFC3339(true)}
	oneDigit := time.Date(2006, 1, 2, 1, 4, 5, 0, time.UTC)
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"hour of one digit", `"2006-01-02T1:04:05Z"`, loose, newOf(time.Time{}), oneDigit},
		{"hour of one digit by default", `"2006-01-02T1:04:05Z"`, nil, newOf(time.Time{}), nil},
		{"key of an hour of one digit", `{"2006-01-02T1:04:05Z":1}`, loose, func() any { return new(map[time.Time]int) },
			map[time.Time]int{oneDigit: 1}},
		{"wrong text", `"2006-01-02"`, loose, newOf(time.Time{}), nil},
	})
}
