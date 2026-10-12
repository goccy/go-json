package json_test

import (
	"encoding"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

type pointerTextKey struct{ S string }

func (p *pointerTextKey) MarshalText() ([]byte, error) { return []byte("p" + p.S), nil }

type valueTextKey struct{ S string }

func (v valueTextKey) MarshalText() ([]byte, error) { return []byte("v" + v.S), nil }

// directTextKey is stored in an interface value directly, as a map is: its MarshalText takes the map, not its
// address.
type directTextKey map[string]string

func (m directTextKey) MarshalText() ([]byte, error) { return []byte("m" + m["s"]), nil }

// directTextKeyPointer is a pointer of a type of its own, which has no methods, to a map of MarshalText.
type directTextKeyPointer *directTextKey

// shapedTextKey is a struct of one pointer, which an interface value holds directly, as a map.
type shapedTextKey struct{ p *string }

func (s shapedTextKey) MarshalText() ([]byte, error) { return []byte("s" + *s.p), nil }

type shapedTextKeyPointer *shapedTextKey

type methodKey interface{ M() }

type methodKeyValue struct{ A int }

func (methodKeyValue) M() {}

// A key of pointers to a value of MarshalText or to an interface is named by the value at the end of its pointers,
// as encoding/json of Go 1.27 names it, and a nil pointer on the way has no name: a key which was called with a
// pointer it doesn't take or named by the address of a map, and one written as null, 5 or an object, which is not
// JSON, fail or are named. A key in a field is named as one at the top.
func TestEncodeMapKeyPointersToNamedValues(t *testing.T) {
	p := &pointerTextKey{"d"}
	pp := &p
	var nilP *pointerTextKey
	v := &valueTextKey{"d"}
	var nilV *valueTextKey
	d := &directTextKey{"s": "d"}
	var nilD *directTextKey
	dp := directTextKeyPointer(d)
	str := "d"
	sp := shapedTextKeyPointer(&shapedTextKey{&str})
	var nothing any
	var nilText encoding.TextMarshaler
	var someText encoding.TextMarshaler = valueTextKey{"w"}
	var method methodKey = methodKeyValue{1}
	for _, tt := range []struct {
		name    string
		v       any
		want    string
		wantErr bool
	}{
		{name: "a pointer to a pointer of MarshalText", v: map[**pointerTextKey]int{&p: 1}, want: `{"pd":1}`},
		{name: "three pointers", v: map[***pointerTextKey]int{&pp: 1}, want: `{"pd":1}`},
		{name: "a nil pointer of MarshalText inside", v: map[**pointerTextKey]int{&nilP: 1}, wantErr: true},
		{name: "a nil key of pointers", v: map[**pointerTextKey]int{nil: 1}, wantErr: true},
		{name: "a nil pointer of MarshalText", v: map[*pointerTextKey]int{nil: 1}, want: `{"":1}`},
		{name: "a pointer to a pointer to a value of MarshalText", v: map[**valueTextKey]int{&v: 1}, want: `{"vd":1}`},
		{name: "a nil pointer to a value of MarshalText inside", v: map[**valueTextKey]int{&nilV: 1}, wantErr: true},
		{name: "a pointer to a pointer to a map of MarshalText", v: map[**directTextKey]int{&d: 1}, want: `{"md":1}`},
		{name: "a nil pointer to a map of MarshalText inside", v: map[**directTextKey]int{&nilD: 1}, wantErr: true},
		{name: "a pointer of its own type to a map of MarshalText", v: map[directTextKeyPointer]int{dp: 1}, want: `{"md":1}`},
		{name: "a pointer to one of its own type", v: map[*directTextKeyPointer]int{&dp: 1}, want: `{"md":1}`},
		{name: "a pointer of its own type to a struct of a pointer", v: map[shapedTextKeyPointer]int{sp: 1}, want: `{"sd":1}`},
		{name: "a pointer to a nil interface", v: map[*any]int{&nothing: 1}, wantErr: true},
		{name: "a nil pointer to an interface", v: map[*any]int{nil: 1}, wantErr: true},
		{name: "a pointer to a nil TextMarshaler", v: map[*encoding.TextMarshaler]int{&nilText: 1}, wantErr: true},
		{name: "a pointer to a TextMarshaler", v: map[*encoding.TextMarshaler]int{&someText: 1}, want: `{"vw":1}`},
		{name: "a pointer to an interface without text", v: map[*methodKey]int{&method: 1}, wantErr: true},
		{name: "in a field", v: struct{ M map[**pointerTextKey]int }{map[**pointerTextKey]int{&p: 1}}, want: `{"M":{"pd":1}}`},
		{name: "to a value, in a field", v: struct{ M map[**valueTextKey]int }{map[**valueTextKey]int{&v: 1}}, want: `{"M":{"vd":1}}`},
		{name: "to a map, in a field", v: struct{ M map[**directTextKey]int }{map[**directTextKey]int{&d: 1}}, want: `{"M":{"md":1}}`},
		{name: "a nil one in a field", v: struct{ M map[**pointerTextKey]int }{map[**pointerTextKey]int{&nilP: 1}}, wantErr: true},
		{name: "to a TextMarshaler, in a field", v: struct {
			M map[*encoding.TextMarshaler]int
		}{map[*encoding.TextMarshaler]int{&someText: 1}}, want: `{"M":{"vw":1}}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.v)
			if tt.wantErr {
				if err == nil || got != nil {
					t.Errorf("got %s %v, want an error", got, err)
				}
				return
			}
			if err != nil || string(got) != tt.want {
				t.Errorf("got %s %v, want %s", got, err, tt.want)
			}
		})
	}
}

// keyColors marks the strings, which the names of keys of MarshalText are colored as, and the numbers.
var keyColors = &json.ColorScheme{
	String: json.ColorFormat{Header: "<s>", Footer: "</s>"},
	Int:    json.ColorFormat{Header: "<i>", Footer: "</i>"},
}

// assertColoredKey checks that the colored output of v, compact and indented, is its plain output with the colors
// of keyColors, and that its one key, named name, is colored as a string, as the name of a key of MarshalText is.
func assertColoredKey(t *testing.T, v any, name string) {
	t.Helper()
	strip := strings.NewReplacer("<s>", "", "</s>", "", "<i>", "", "</i>", "")
	for _, indent := range []bool{false, true} {
		var plain, colored []byte
		var err error
		if indent {
			plain, err = json.MarshalIndent(v, "", " ")
			if err == nil {
				colored, err = json.MarshalIndentWithOption(v, "", " ", json.Colorize(keyColors))
			}
		} else {
			plain, err = json.Marshal(v)
			if err == nil {
				colored, err = json.MarshalWithOption(v, json.Colorize(keyColors))
			}
		}
		if err != nil {
			t.Fatal(err)
		}
		if got := strip.Replace(string(colored)); got != string(plain) {
			t.Errorf("indent %v: colored %s, without its colors %s, want %s", indent, colored, got, plain)
		}
		if key := `<s>"` + name + `"</s>:`; !strings.Contains(string(colored), key) {
			t.Errorf("indent %v: colored %s, want the key %s", indent, colored, key)
		}
	}
}

// A key of pointers which its value names is colored as a string, as a key of MarshalText is.
func TestEncodeMapKeyPointersColored(t *testing.T) {
	p := &pointerTextKey{"d"}
	pp := &p
	v := &valueTextKey{"d"}
	vv := &v
	var text encoding.TextMarshaler = valueTextKey{"w"}
	textp := &text
	d := &directTextKey{"s": "d"}
	for name, tt := range map[string]struct {
		v    any
		name string
	}{
		"a pointer of MarshalText":               {map[*pointerTextKey]int{p: 1}, "pd"},
		"two pointers to one of MarshalText":     {map[**pointerTextKey]int{&p: 1}, "pd"},
		"three pointers to one of MarshalText":   {map[***pointerTextKey]int{&pp: 1}, "pd"},
		"two pointers to a value of MarshalText": {map[**valueTextKey]int{&v: 1}, "vd"},
		"three pointers to a value":              {map[***valueTextKey]int{&vv: 1}, "vd"},
		"a pointer to a TextMarshaler":           {map[*encoding.TextMarshaler]int{&text: 1}, "vw"},
		"two pointers to a TextMarshaler":        {map[**encoding.TextMarshaler]int{&textp: 1}, "vw"},
		"two pointers to a map of MarshalText":   {map[**directTextKey]int{&d: 1}, "md"},
	} {
		t.Run(name, func(t *testing.T) { assertColoredKey(t, tt.v, tt.name) })
	}
}
