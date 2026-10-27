package json_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The opcodes and the decoders which the options of encoding/json change are kept apart from the ones of other
// options: calls of different options by turns, through every function, through the values of interfaces and
// through the representation which a method declines to, each get the output of their own options, also with
// functions of marshaling and unmarshaling which every call shares.

// legacyDeclining is a struct whose method declines every value, which is then written by its default
// representation in a context of its own, which looks up the type of the value of its field there.
type legacyDeclining struct{ Any any }

func (legacyDeclining) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

// legacyNestedArray decodes its value into a [2]byte by UnmarshalDecode with FormatByteArrayAsArray, whatever the
// options of the call around it are.
type legacyNestedArray [2]byte

func (a *legacyNestedArray) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return json.UnmarshalDecode(dec, (*[2]byte)(a), gojson.FormatByteArrayAsArray(true))
}

// legacyTurnsBytes has bytes which the options of encoding/json change: in a field, in a value of an interface, and
// behind a method which declines them.
type legacyTurnsBytes struct {
	A   [2]byte
	B   []legacyByte
	Any any
	M   legacyDeclining
}

// legacyTurnsDuration has a time.Duration, which only FormatDurationAsNano marshals, in a field and in a value of an
// interface.
type legacyTurnsDuration struct {
	D   time.Duration
	Any any
}

// legacyTurnsSet is a set of the options of the calls by turns.
type legacyTurnsSet struct {
	name                       string
	array, legacy, nano, funcs bool
}

var legacyTurnsSets = []legacyTurnsSet{
	{name: "default"},
	{name: "FormatByteArrayAsArray", array: true},
	{name: "FormatBytesWithLegacySemantics", legacy: true},
	{name: "FormatDurationAsNano", nano: true},
	{name: "all three", array: true, legacy: true, nano: true},
	{name: "functions", funcs: true},
	{name: "FormatByteArrayAsArray with functions", array: true, funcs: true},
	{name: "FormatDurationAsNano with functions", nano: true, funcs: true},
}

// legacyTurns are the sets of the calls by turns: the default options after each other set, so that a call of the
// default options finds what a call of other options left in the context.
func legacyTurns() []legacyTurnsSet {
	var turns []legacyTurnsSet
	for _, s := range legacyTurnsSets[1:] {
		turns = append(turns, s, legacyTurnsSets[0])
	}
	return turns
}

func (s legacyTurnsSet) options(m *json.Marshalers, u *json.Unmarshalers) []json.Options {
	var opts []json.Options
	if s.array {
		opts = append(opts, gojson.FormatByteArrayAsArray(true))
	}
	if s.legacy {
		opts = append(opts, gojson.FormatBytesWithLegacySemantics(true))
	}
	if s.nano {
		opts = append(opts, gojson.FormatDurationAsNano(true))
	}
	if s.funcs {
		opts = append(opts, json.WithMarshalers(m), json.WithUnmarshalers(u))
	}
	return opts
}

// bytes is the output of legacyTurnsValue by the options.
func (s legacyTurnsSet) bytes() string {
	a, held, m := `"AQI="`, `"BAU="`, `"Bgc="`
	if s.array {
		a, held, m = `[1,2]`, `[4,5]`, `[6,7]`
	}
	b := `[3]`
	if s.legacy {
		b = `"Aw=="`
	}
	return `{"A":` + a + `,"B":` + b + `,"Any":` + held + `,"M":{"Any":` + m + `}}`
}

// duration is the output of legacyTurnsDuration{1, time.Duration(2)} by the options, or "" for an error.
func (s legacyTurnsSet) duration() string {
	if !s.nano {
		return ""
	}
	return `{"D":1,"Any":2}`
}

var legacyTurnsMarshals = []struct {
	name    string
	marshal func(v any, opts []json.Options) (string, error)
}{
	{"Marshal", func(v any, opts []json.Options) (string, error) {
		b, err := json.Marshal(v, opts...)
		return string(b), err
	}},
	{"MarshalWrite", func(v any, opts []json.Options) (string, error) {
		var buf bytes.Buffer
		err := json.MarshalWrite(&buf, v, opts...)
		return buf.String(), err
	}},
	{"MarshalEncode with the options of the call", func(v any, opts []json.Options) (string, error) {
		var buf bytes.Buffer
		err := json.MarshalEncode(jsontext.NewEncoder(&buf), v, opts...)
		return strings.TrimSuffix(buf.String(), "\n"), err
	}},
	{"MarshalEncode with the options of the encoder", func(v any, opts []json.Options) (string, error) {
		var buf bytes.Buffer
		err := json.MarshalEncode(jsontext.NewEncoder(&buf, opts...), v)
		return strings.TrimSuffix(buf.String(), "\n"), err
	}},
}

var legacyTurnsUnmarshals = []struct {
	name      string
	unmarshal func(in string, out any, opts []json.Options) error
}{
	{"Unmarshal", func(in string, out any, opts []json.Options) error {
		return json.Unmarshal([]byte(in), out, opts...)
	}},
	{"UnmarshalRead", func(in string, out any, opts []json.Options) error {
		return json.UnmarshalRead(strings.NewReader(in), out, opts...)
	}},
	{"UnmarshalDecode with the options of the call", func(in string, out any, opts []json.Options) error {
		return json.UnmarshalDecode(jsontext.NewDecoder(strings.NewReader(in)), out, opts...)
	}},
	{"UnmarshalDecode with the options of the decoder", func(in string, out any, opts []json.Options) error {
		return json.UnmarshalDecode(jsontext.NewDecoder(strings.NewReader(in), opts...), out)
	}},
}

var legacyTurnsInputs = []struct {
	name string
	in   string
	out  func() any
	// want is the value which out points to after the call by the options, or nil for an error
	want func(s legacyTurnsSet) any
}{
	{"byte array", `{"A":[1,2]}`, func() any { return new(struct{ A [2]byte }) }, func(s legacyTurnsSet) any {
		// a JSON array, which FormatBytesWithLegacySemantics reads element by element too
		if !s.array && !s.legacy {
			return nil
		}
		return struct{ A [2]byte }{[2]byte{1, 2}}
	}},
	{"named bytes", `{"B":"Aw=="}`, func() any { return new(struct{ B []legacyByte }) }, func(s legacyTurnsSet) any {
		if !s.legacy {
			return nil
		}
		return struct{ B []legacyByte }{[]legacyByte{3}}
	}},
	{"duration", `{"D":5}`, func() any { return new(struct{ D time.Duration }) }, func(s legacyTurnsSet) any {
		if !s.nano {
			return nil
		}
		return struct{ D time.Duration }{5}
	}},
	// a call of other options within the call, after which the bytes held by the interface are of the call around it:
	// the long string makes a stream decode the object member by member, and the call within it on the stream
	{"bytes after a nested call", `{"L":"` + strings.Repeat("x", 1000) + `","X":[3,4],"Any":"AQI="}`, func() any {
		return &struct {
			L   string
			X   legacyNestedArray
			Any any
		}{Any: new([2]byte)}
	}, func(s legacyTurnsSet) any {
		if s.array {
			return nil
		}
		return struct {
			L   string
			X   legacyNestedArray
			Any any
		}{L: strings.Repeat("x", 1000), X: legacyNestedArray{3, 4}, Any: &[2]byte{1, 2}}
	}},
	{"byte array held by an interface", `[1,2]`, func() any {
		var v any = new([2]byte)
		return &v
	}, func(s legacyTurnsSet) any {
		if !s.array && !s.legacy {
			return nil
		}
		return &[2]byte{1, 2}
	}},
}

func TestOptionsOfEncodingJSONByTurns(t *testing.T) {
	// functions of types which no value has, which every call of the functions shares
	marshalers := json.MarshalFunc(func(int32) ([]byte, error) { return []byte("0"), nil })
	unmarshalers := json.UnmarshalFunc(func([]byte, *int32) error { return nil })
	values := []struct {
		v    any
		want func(s legacyTurnsSet) string
	}{
		{legacyTurnsBytes{A: [2]byte{1, 2}, B: []legacyByte{3}, Any: [2]byte{4, 5}, M: legacyDeclining{[2]byte{6, 7}}},
			legacyTurnsSet.bytes},
		{legacyTurnsDuration{D: 1, Any: time.Duration(2)}, legacyTurnsSet.duration},
	}
	// each function by itself, so that a call finds what the call of the function before it left
	for _, m := range legacyTurnsMarshals {
		for range 2 {
			for _, s := range legacyTurns() {
				opts := s.options(marshalers, unmarshalers)
				for _, v := range values {
					got, err := m.marshal(v.v, opts)
					switch want := v.want(s); {
					case want == "" && err == nil:
						t.Errorf("%s, %s, %T: got %s, want an error", s.name, m.name, v.v, got)
					case want != "" && (err != nil || got != want):
						t.Errorf("%s, %s, %T: got %s, %v, want %s", s.name, m.name, v.v, got, err, want)
					}
				}
			}
		}
	}
	for _, u := range legacyTurnsUnmarshals {
		for range 2 {
			for _, s := range legacyTurns() {
				opts := s.options(marshalers, unmarshalers)
				for _, in := range legacyTurnsInputs {
					out := in.out()
					err := u.unmarshal(in.in, out, opts)
					got := reflect.ValueOf(out).Elem().Interface()
					switch want := in.want(s); {
					case want == nil && err == nil:
						t.Errorf("%s, %s, %s: got %#v, want an error", s.name, u.name, in.name, got)
					case want != nil && err != nil:
						t.Errorf("%s, %s, %s: %v", s.name, u.name, in.name, err)
					case want != nil && !reflect.DeepEqual(got, want):
						t.Errorf("%s, %s, %s: got %#v, want %#v", s.name, u.name, in.name, got, want)
					}
				}
			}
		}
	}
}
