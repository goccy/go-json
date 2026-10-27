package json_test

import (
	"strings"
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

// legacyByte is a named byte type, whose slices and arrays are of bytes only with FormatBytesWithLegacySemantics.
type legacyByte byte

// legacyTextByte is a named byte type with methods of marshaling, whose slices and arrays are JSON arrays of its
// values when marshaling, even with FormatBytesWithLegacySemantics: a value is a text of as many b as it is.
type legacyTextByte byte

func (b legacyTextByte) MarshalText() ([]byte, error) {
	return []byte(strings.Repeat("b", int(b))), nil
}

func (b *legacyTextByte) UnmarshalText(text []byte) error {
	*b = legacyTextByte(len(text))
	return nil
}

func TestMarshalFormatBytes(t *testing.T) {
	array := []json.Options{gojson.FormatByteArrayAsArray(true)}
	legacy := []json.Options{gojson.FormatBytesWithLegacySemantics(true)}
	both := append(array[:1:1], legacy...)
	runLegacyMarshalCases(t, []legacyMarshalCase{
		{"byte array as array", [3]byte{1, 2, 3}, array, `[1,2,3]`},
		{"byte array by default", [3]byte{1, 2, 3}, nil, `"AQID"`},
		{"byte array as array stringified", [2]byte{1, 2}, append(array, json.StringifyNumbers(true)), `["1","2"]`},
		{"byte slice with FormatByteArrayAsArray", []byte{1, 2, 3}, array, `"AQID"`},
		{"byte array key as array", map[[2]byte]int{{1, 2}: 1}, array, ``},
		{"byte array key", map[[2]byte]int{{1, 2}: 1}, nil, `{"AQI=":1}`},
		{"named bytes", []legacyByte{1, 2, 3}, legacy, `"AQID"`},
		{"named bytes by default", []legacyByte{1, 2, 3}, nil, `[1,2,3]`},
		{"nil named bytes", []legacyByte(nil), legacy, `""`},
		{"nil named bytes as null", []legacyByte(nil), append(legacy, json.FormatNilSliceAsNull(true)), `null`},
		{"named byte array", [2]legacyByte{1, 2}, legacy, `"AQI="`},
		{"named byte array as array", [2]legacyByte{1, 2}, both, `[1,2]`},
		{"named byte array key", map[[2]legacyByte]int{{1, 2}: 1}, legacy, `{"AQI=":1}`},
		{"text bytes", []legacyTextByte{1, 2}, legacy, `["b","bb"]`},
		{"text byte array", [2]legacyTextByte{1, 2}, legacy, `["b","bb"]`},
		{"named bytes in an interface", []any{[]legacyByte{1}}, legacy, `["AQ=="]`},
	})
}

func TestUnmarshalFormatBytes(t *testing.T) {
	array := []json.Options{gojson.FormatByteArrayAsArray(true)}
	legacy := []json.Options{gojson.FormatBytesWithLegacySemantics(true)}
	runLegacyUnmarshalCases(t, []legacyUnmarshalCase{
		{"byte array as array", `[1,2]`, array, newOf([2]byte{}), [2]byte{1, 2}},
		{"byte array as array of base64", `"AQI="`, array, newOf([2]byte{}), nil},
		{"byte array key as array", `{"AQI=":1}`, array, func() any { return new(map[[2]byte]int) }, nil},
		{"bytes of a JSON array", `[1,2,3]`, legacy, newOf([]byte(nil)), []byte{1, 2, 3}},
		{"bytes of a JSON array by default", `[1,2,3]`, nil, newOf([]byte(nil)), nil},
		{"byte array of a JSON array", `[1,2]`, legacy, newOf([2]byte{}), [2]byte{1, 2}},
		{"named bytes", `"AQID"`, legacy, newOf([]legacyByte(nil)), []legacyByte{1, 2, 3}},
		{"named bytes by default", `"AQID"`, nil, newOf([]legacyByte(nil)), nil},
		{"named byte array key", `{"AQI=":1}`, legacy, func() any { return new(map[[2]legacyByte]int) },
			map[[2]legacyByte]int{{1, 2}: 1}},
		{"text bytes of a JSON array", `["b","bb"]`, legacy, newOf([]legacyTextByte(nil)), []legacyTextByte{1, 2}},
		{"text bytes of base64", `"AQI="`, legacy, newOf([]legacyTextByte(nil)), []legacyTextByte{1, 2}},
	})
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
