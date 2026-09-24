package json_test

import (
	stdjson "encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	json "github.com/goccy/go-json"
)

type unmarshalOfSmall struct {
	St   int
	Sid  int
	Tt   string
	Gr   int
	Uuid string
	Ip   string
	Ua   string
	Tz   int
	V    int
}

const unmarshalOfSmallInput = `{"st": 1,"sid": 486,"tt": "active","gr": 0,"uuid": "de305d54-75b4-431b-adb2-eb6b9e546014","ip": "127.0.0.1","ua": "user_agent","tz": -6,"v": 1}`

// unmarshalOfCase decodes the input by UnmarshalOf and by Unmarshal into two values which start the same.
func unmarshalOfCase[T any](t *testing.T, input string, start T) {
	t.Helper()
	got, want := start, start
	errOf := json.UnmarshalOf([]byte(input), &got)
	err := json.Unmarshal([]byte(input), &want)
	if (errOf == nil) != (err == nil) || (err != nil && errOf.Error() != err.Error()) {
		t.Fatalf("%s: UnmarshalOf: %v, Unmarshal: %v", input, errOf, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: UnmarshalOf: %#v, Unmarshal: %#v", input, got, want)
	}
}

type unmarshalOfLarge struct {
	A    [600]int
	Name string
}

func TestUnmarshalOfAsUnmarshal(t *testing.T) {
	unmarshalOfCase(t, unmarshalOfSmallInput, unmarshalOfSmall{})
	// into a value which exists: the fields absent from the input are kept.
	unmarshalOfCase(t, `{"st":5}`, unmarshalOfSmall{Tt: "kept", V: 9})
	unmarshalOfCase(t, `{"a":[1,2],"m":{"x":"y"},"p":{"n":3}}`, struct {
		A []int
		M map[string]string
		P *struct{ N int }
	}{})
	unmarshalOfCase(t, `[1,"a",{"b":null}]`, any(nil))
	unmarshalOfCase(t, `"2026-09-24T00:00:00Z"`, time.Time{})
	unmarshalOfCase(t, `123`, 0)
	unmarshalOfCase(t, `"s"`, "")
	unmarshalOfCase(t, `{"x":1}`, map[string]int{"y": 2})
	unmarshalOfCase(t, `[1,2,3]`, []int{})
	unmarshalOfCase(t, `{"Name":"large","A":[1,2,3]}`, unmarshalOfLarge{})
	// errors, with what was decoded before them
	unmarshalOfCase(t, `{"st":1,"tt":2}`, unmarshalOfSmall{})
	unmarshalOfCase(t, `{"st":1} x`, unmarshalOfSmall{})
	unmarshalOfCase(t, `{"st":`, unmarshalOfSmall{})

	var nilPtr *unmarshalOfSmall
	var invalid *json.InvalidUnmarshalError
	if err := json.UnmarshalOf([]byte(`{}`), nilPtr); !errors.As(err, &invalid) {
		t.Fatalf("expected InvalidUnmarshalError, got %v", err)
	}
}

func TestUnmarshalOfOtherTypesByTurns(t *testing.T) {
	// The values reused by a context are of the type decoded last: other types come by turns.
	for i := 0; i < 20; i++ {
		var a unmarshalOfSmall
		if err := json.UnmarshalOf([]byte(unmarshalOfSmallInput), &a); err != nil || a.Uuid != "de305d54-75b4-431b-adb2-eb6b9e546014" {
			t.Fatalf("got %+v, %v", a, err)
		}
		var b struct{ X []string }
		if err := json.UnmarshalOf([]byte(`{"X":["p","q"]}`), &b); err != nil || !reflect.DeepEqual(b.X, []string{"p", "q"}) {
			t.Fatalf("got %+v, %v", b, err)
		}
		// the reused value is zero: nothing of the previous call is decoded into
		var c unmarshalOfSmall
		if err := json.UnmarshalOf([]byte(`{"st":1}`), &c); err != nil || c != (unmarshalOfSmall{St: 1}) {
			t.Fatalf("got %+v, %v", c, err)
		}
	}
}

func TestUnmarshalOfZeroAllocation(t *testing.T) {
	if raceEnabled {
		t.Skip("sync.Pool drops items at random under the race detector")
	}
	data := []byte(unmarshalOfSmallInput)
	var v unmarshalOfSmall
	for _, opts := range [][]json.DecodeOptionFunc{nil, {json.DecodeNoCopyString()}} {
		allocs := testing.AllocsPerRun(1000, func() {
			var w unmarshalOfSmall
			if err := json.UnmarshalOf(data, &w, opts...); err != nil {
				t.Fatal(err)
			}
			v = w
		})
		// The strings share a buffer of the context, which is allocated once in a while.
		if allocs >= 0.1 {
			t.Fatalf("options %d: %v allocations per call", len(opts), allocs)
		}
	}
	if v.Ua != "user_agent" {
		t.Fatalf("got %+v", v)
	}
}

func TestDecodedStringsOwnership(t *testing.T) {
	type pair struct {
		A string
		B string
		M map[string]any
	}
	input := `{"A":"plain","B":"esc\"aped","M":{"key":"value"}}`
	decoders := map[string]func([]byte, *pair, ...json.DecodeOptionFunc) error{
		"Unmarshal": func(data []byte, v *pair, opts ...json.DecodeOptionFunc) error {
			return json.UnmarshalWithOption(data, v, opts...)
		},
		"UnmarshalOf": func(data []byte, v *pair, opts ...json.DecodeOptionFunc) error {
			return json.UnmarshalOf(data, v, opts...)
		},
	}
	for name, decode := range decoders {
		// By default the strings are copies: the input may be modified after the call,
		// and the buffer of the input is reused by the next call.
		data := []byte(input)
		var v pair
		if err := decode(data, &v); err != nil {
			t.Fatal(err)
		}
		for i := range data {
			data[i] = 'X'
		}
		var other pair
		if err := decode([]byte(strings.ReplaceAll(input, "plain", "PLAIN")), &other); err != nil {
			t.Fatal(err)
		}
		if v.A != "plain" || v.B != `esc"aped` || v.M["key"] != "value" {
			t.Fatalf("%s: the strings changed with the input: %+v", name, v)
		}
		// With DecodeNoCopyString a string without an escape refers to the input.
		data = []byte(input)
		v = pair{}
		if err := decode(data, &v, json.DecodeNoCopyString()); err != nil {
			t.Fatal(err)
		}
		if v.A != "plain" || v.B != `esc"aped` || v.M["key"] != "value" {
			t.Fatalf("%s: got %+v", name, v)
		}
		start := uintptr(unsafe.Pointer(&data[0]))
		within := func(s string) bool {
			p := uintptr(unsafe.Pointer(unsafe.StringData(s)))
			return p >= start && p < start+uintptr(len(data))
		}
		if !within(v.A) || !within(v.M["key"].(string)) {
			t.Fatalf("%s: a string without an escape doesn't refer to the input", name)
		}
		if within(v.B) {
			t.Fatalf("%s: an escaped string refers to the input, which has other bytes", name)
		}
	}
}

func TestDecodedEscapedStringsAsEncodingJSON(t *testing.T) {
	inputs := []string{
		`"plain"`, `"tab\there"`, `"quote\"and\\backslash\/slash"`, `"\u00e9\u3042"`, `"\ud83d\ude00 surrogate"`,
		`"\ud83d alone"`, "\"bad \xff byte\"", "\"bad \xff with \\n escape\"", `"` + strings.Repeat(`a\n`, 400) + `"`,
		`""`, `"\u0000"`,
	}
	type holder struct {
		S string
		M map[string]string
		A any
	}
	for _, in := range inputs {
		doc := `{"S":` + in + `,"M":{` + in + `:` + in + `},"A":[` + in + `]}`
		var want holder
		if err := stdjsonUnmarshal([]byte(doc), &want); err != nil {
			t.Fatalf("%s: %v", doc, err)
		}
		for name, decode := range map[string]func([]byte, *holder) error{
			"Unmarshal":   func(d []byte, v *holder) error { return json.Unmarshal(d, v) },
			"UnmarshalOf": func(d []byte, v *holder) error { return json.UnmarshalOf(d, v) },
			"NoCopyString": func(d []byte, v *holder) error {
				return json.UnmarshalWithOption(d, v, json.DecodeNoCopyString())
			},
			"Decoder": func(d []byte, v *holder) error { return json.NewDecoder(strings.NewReader(string(d))).Decode(v) },
		} {
			var got holder
			if err := decode([]byte(doc), &got); err != nil {
				t.Fatalf("%s %s: %v", name, doc, err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s %s:\n got %q\nwant %q", name, doc, got, want)
			}
		}
	}
}

func stdjsonUnmarshal(data []byte, v any) error {
	return stdjson.Unmarshal(data, v)
}
