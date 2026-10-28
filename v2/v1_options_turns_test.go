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

// marshalFuncs are the functions of marshaling, which a test calls each with the same value and options: MarshalEncode
// with the options of the call, to a writer and to a bytes.Buffer, which an encoder writes in place, and with the
// options of the encoder. The output is what the function wrote, without the line feed which an encoder writes after
// a value at the top level.
var marshalFuncs = []struct {
	name    string
	marshal func(v any, opts ...json.Options) (string, error)
}{
	{"Marshal", func(v any, opts ...json.Options) (string, error) {
		b, err := json.Marshal(v, opts...)
		return string(b), err
	}},
	{"MarshalWrite", func(v any, opts ...json.Options) (string, error) {
		var out strings.Builder
		err := json.MarshalWrite(&out, v, opts...)
		return out.String(), err
	}},
	{"MarshalEncode with the options of the call", func(v any, opts ...json.Options) (string, error) {
		var out strings.Builder
		err := json.MarshalEncode(jsontext.NewEncoder(&out), v, opts...)
		return strings.TrimSuffix(out.String(), "\n"), err
	}},
	{"MarshalEncode to a bytes.Buffer with the options of the call", func(v any, opts ...json.Options) (string, error) {
		var out bytes.Buffer
		err := json.MarshalEncode(jsontext.NewEncoder(&out), v, opts...)
		return strings.TrimSuffix(out.String(), "\n"), err
	}},
	{"MarshalEncode to a bytes.Buffer with the options of the encoder", func(v any, opts ...json.Options) (string, error) {
		var out bytes.Buffer
		err := json.MarshalEncode(jsontext.NewEncoder(&out, opts...), v)
		return strings.TrimSuffix(out.String(), "\n"), err
	}},
}

// unmarshalFuncs are the functions of unmarshaling, which a test calls each with the same input and options:
// UnmarshalDecode with the options of the call and with the options of the decoder.
var unmarshalFuncs = []struct {
	name      string
	unmarshal func(in string, out any, opts ...json.Options) error
}{
	{"Unmarshal", func(in string, out any, opts ...json.Options) error {
		return json.Unmarshal([]byte(in), out, opts...)
	}},
	{"UnmarshalRead", func(in string, out any, opts ...json.Options) error {
		return json.UnmarshalRead(strings.NewReader(in), out, opts...)
	}},
	{"UnmarshalDecode with the options of the call", func(in string, out any, opts ...json.Options) error {
		return json.UnmarshalDecode(jsontext.NewDecoder(strings.NewReader(in)), out, opts...)
	}},
	{"UnmarshalDecode with the options of the decoder", func(in string, out any, opts ...json.Options) error {
		return json.UnmarshalDecode(jsontext.NewDecoder(strings.NewReader(in), opts...), out)
	}},
}

// v1Declining is a struct whose method declines every value, which is then written by its default
// representation in a context of its own, which looks up the type of the value of its field there.
type v1Declining struct{ Any any }

func (v1Declining) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

// v1NestedArray decodes its value into a [2]byte by UnmarshalDecode with FormatByteArrayAsArray, whatever the
// options of the call around it are.
type v1NestedArray [2]byte

func (a *v1NestedArray) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return json.UnmarshalDecode(dec, (*[2]byte)(a), gojson.FormatByteArrayAsArray(true))
}

// v1TurnsBytes has bytes which the options of encoding/json change: in a field, in a value of an interface, and
// behind a method which declines them.
type v1TurnsBytes struct {
	A   [2]byte
	B   []v1Byte
	Any any
	M   v1Declining
}

// v1TurnsDuration has a time.Duration, which only FormatDurationAsNano marshals, in a field and in a value of an
// interface.
type v1TurnsDuration struct {
	D   time.Duration
	Any any
}

// v1TurnsSet are the options of a call by turns.
type v1TurnsSet struct {
	name                            string
	array, formatBytes, nano, funcs bool
}

var v1TurnsSets = []v1TurnsSet{
	{name: "default"},
	{name: "FormatByteArrayAsArray", array: true},
	{name: "FormatBytesWithLegacySemantics", formatBytes: true},
	{name: "FormatDurationAsNano", nano: true},
	{name: "all three", array: true, formatBytes: true, nano: true},
	{name: "functions", funcs: true},
	{name: "FormatByteArrayAsArray with functions", array: true, funcs: true},
	{name: "FormatDurationAsNano with functions", nano: true, funcs: true},
}

// v1Turns are the options of the calls by turns: the default options after each other options, so that a call of
// the default options finds what a call of other options left in the context.
func v1Turns() []v1TurnsSet {
	var turns []v1TurnsSet
	for _, s := range v1TurnsSets[1:] {
		turns = append(turns, s, v1TurnsSets[0])
	}
	return turns
}

func (s v1TurnsSet) options(m *json.Marshalers, u *json.Unmarshalers) []json.Options {
	var opts []json.Options
	if s.array {
		opts = append(opts, gojson.FormatByteArrayAsArray(true))
	}
	if s.formatBytes {
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

// bytes is the output of the value of v1TurnsBytes of TestOptionsOfEncodingJSONByTurns by the options.
func (s v1TurnsSet) bytes() string {
	a, held, m := `"AQI="`, `"BAU="`, `"Bgc="`
	if s.array {
		a, held, m = `[1,2]`, `[4,5]`, `[6,7]`
	}
	b := `[3]`
	if s.formatBytes {
		b = `"Aw=="`
	}
	return `{"A":` + a + `,"B":` + b + `,"Any":` + held + `,"M":{"Any":` + m + `}}`
}

// duration is the output of v1TurnsDuration{1, time.Duration(2)} by the options, or "" for an error.
func (s v1TurnsSet) duration() string {
	if !s.nano {
		return ""
	}
	return `{"D":1,"Any":2}`
}

// nestedCallOfTurns is the value of the input of a call of other options within the call, a struct of no name: the
// long string makes a stream decode the object member by member, and the call within it on the stream.
type nestedCallOfTurns = struct {
	L   string
	X   v1NestedArray
	Any any
}

var v1TurnsInputs = []struct {
	name string
	in   string
	out  func() any
	// want is the value which out points to after the call by the options, or nil for an error
	want func(s v1TurnsSet) any
}{
	{"byte array", `{"A":[1,2]}`, outOf(struct{ A [2]byte }{}), func(s v1TurnsSet) any {
		// a JSON array, which FormatBytesWithLegacySemantics reads element by element too
		if !s.array && !s.formatBytes {
			return nil
		}
		return struct{ A [2]byte }{[2]byte{1, 2}}
	}},
	{"named bytes", `{"B":"Aw=="}`, outOf(struct{ B []v1Byte }{}), func(s v1TurnsSet) any {
		if !s.formatBytes {
			return nil
		}
		return struct{ B []v1Byte }{[]v1Byte{3}}
	}},
	{"duration", `{"D":5}`, outOf(struct{ D time.Duration }{}), func(s v1TurnsSet) any {
		if !s.nano {
			return nil
		}
		return struct{ D time.Duration }{5}
	}},
	// the bytes held by the interface are decoded by the options of the call around the call within it
	{"bytes after a nested call", `{"L":"` + longerThanStreamBuffer + `","X":[3,4],"Any":"AQI="}`, func() any {
		return &nestedCallOfTurns{Any: new([2]byte)}
	}, func(s v1TurnsSet) any {
		if s.array {
			return nil
		}
		return nestedCallOfTurns{L: longerThanStreamBuffer, X: v1NestedArray{3, 4}, Any: &[2]byte{1, 2}}
	}},
	{"byte array held by an interface", `[1,2]`, func() any {
		var v any = new([2]byte)
		return &v
	}, func(s v1TurnsSet) any {
		if !s.array && !s.formatBytes {
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
		want func(s v1TurnsSet) string
	}{
		{v1TurnsBytes{A: [2]byte{1, 2}, B: []v1Byte{3}, Any: [2]byte{4, 5}, M: v1Declining{[2]byte{6, 7}}},
			v1TurnsSet.bytes},
		{v1TurnsDuration{D: 1, Any: time.Duration(2)}, v1TurnsSet.duration},
	}
	// each function by itself, so that a call finds what the call of the function before it left; the turns are taken
	// twice, so that the calls of the second round find the code of their types compiled by the first one, and its
	// first call comes after a call of the default options too.
	const rounds = 2
	for _, m := range marshalFuncs {
		for range rounds {
			for _, s := range v1Turns() {
				opts := s.options(marshalers, unmarshalers)
				for _, v := range values {
					got, err := m.marshal(v.v, opts...)
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
	for _, u := range unmarshalFuncs {
		for range rounds {
			for _, s := range v1Turns() {
				opts := s.options(marshalers, unmarshalers)
				for _, in := range v1TurnsInputs {
					out := in.out()
					err := u.unmarshal(in.in, out, opts...)
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
