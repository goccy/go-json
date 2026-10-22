//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	"encoding"
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The tests here compare the decoding by the methods and the functions of unmarshaling, and of the time types, with
// the one of encoding/json/v2, as the tests of std_unmarshal_diff_test.go do.

// umText is decoded by UnmarshalText, which fails for "bad".
type umText struct{ S string }

func (t *umText) UnmarshalText(b []byte) error {
	if string(b) == "bad" {
		return errors.New("bad text")
	}
	t.S = "text:" + string(b)
	return nil
}

// umJSON is decoded by UnmarshalJSON, which keeps the value it is given, and fails for "bad".
type umJSON struct{ Raw string }

func (j *umJSON) UnmarshalJSON(b []byte) error {
	if string(b) == `"bad"` {
		return errors.New("bad json")
	}
	j.Raw = string(b)
	return nil
}

// umBoth has both methods, of which UnmarshalJSON decodes it.
type umBoth struct{ S string }

func (b *umBoth) UnmarshalJSON(v []byte) error { b.S = "json:" + string(v); return nil }

func (b *umBoth) UnmarshalText(v []byte) error { b.S = "text:" + string(v); return nil }

// umUnsupported has an UnmarshalJSON which returns errors.ErrUnsupported, which is an error of a method.
type umUnsupported struct{}

func (*umUnsupported) UnmarshalJSON([]byte) error { return errors.ErrUnsupported }

// umTextKey is a key of a map decoded by UnmarshalText.
type umTextKey string

func (k *umTextKey) UnmarshalText(b []byte) error {
	*k = umTextKey("key:" + string(b))
	return nil
}

type umFields struct {
	T  umText
	P  *umText
	J  umJSON
	JP *umJSON
	Tm time.Time
	TP *time.Time
}

type umTimeString struct {
	T time.Time `json:",string"`
}

// umOptions are the options of a case, made for each package: the functions of unmarshaling are of the package.
type umOptions struct {
	ours []json.Options
	std  []stdjson.Options
}

func umFuncs[T any](fn func([]byte, T) error) umOptions {
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.UnmarshalFunc(fn))},
		std:  []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFunc(fn))},
	}
}

func umJoin(os ...umOptions) umOptions {
	var ours []*json.Unmarshalers
	var std []*stdjson.Unmarshalers
	for _, o := range os {
		ours = append(ours, umUnmarshalers(o.ours))
		std = append(std, umStdUnmarshalers(o.std))
	}
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.JoinUnmarshalers(ours...))},
		std:  []stdjson.Options{stdjson.WithUnmarshalers(stdjson.JoinUnmarshalers(std...))},
	}
}

func umUnmarshalers(opts []json.Options) *json.Unmarshalers {
	u, _ := json.GetOption(json.JoinOptions(opts...), json.WithUnmarshalers)
	return u
}

func umStdUnmarshalers(opts []stdjson.Options) *stdjson.Unmarshalers {
	u, _ := stdjson.GetOption(stdjson.JoinOptions(opts...), stdjson.WithUnmarshalers)
	return u
}

func (o umOptions) and(ours json.Options, std stdjson.Options) umOptions {
	return umOptions{ours: append(append([]json.Options{}, o.ours...), ours), std: append(append([]stdjson.Options{}, o.std...), std)}
}

// umNumbering returns the options of a function which decodes every string into the number of the strings it
// decoded before it: one name of an object is two keys.
func umNumbering() umOptions {
	var ours, theirs int
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.UnmarshalFunc(func(b []byte, v *string) error {
			ours++
			*v = strconv.Itoa(ours)
			return nil
		}))},
		std: []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFunc(func(b []byte, v *string) error {
			theirs++
			*v = strconv.Itoa(theirs)
			return nil
		}))},
	}
}

// The functions of the cases.
var (
	umLength = umFuncs(func(b []byte, v *int) error {
		*v = len(b)
		return nil
	})
	umDouble = umFuncs(func(b []byte, v *int) error {
		n, err := strconv.Atoi(string(b))
		*v = 2 * n
		return err
	})
	umFailing = umFuncs(func(b []byte, v *int) error {
		return errors.New("failing function")
	})
	umDeclining = umFuncs(func(b []byte, v *int) error {
		return errors.ErrUnsupported
	})
	umTextFunc = umFuncs(func(b []byte, v encoding.TextUnmarshaler) error {
		return v.UnmarshalText([]byte("func"))
	})
	umStringFunc = umFuncs(func(b []byte, v *string) error {
		*v = "func:" + string(b)
		return nil
	})
	umFloatFunc = umFuncs(func(b []byte, v *float64) error {
		*v = -1
		return nil
	})
	umMapFunc = umFuncs(func(b []byte, v *map[string]any) error {
		*v = map[string]any{"func": string(b)}
		return nil
	})
	umSliceFunc = umFuncs(func(b []byte, v *[]any) error {
		*v = []any{"func"}
		return nil
	})
	umStringFailing = umFuncs(func(b []byte, v *string) error {
		if string(b) == `"bad"` {
			return errors.New("failing function")
		}
		*v = string(b)
		return nil
	})
	umOneName = umFuncs(func(b []byte, v *string) error {
		*v = "name"
		return nil
	})
	umAnyFunc = umFuncs(func(b []byte, v *any) error {
		*v = "any:" + string(b)
		return nil
	})
)

type umCase struct {
	name string
	in   string
	out  func() any
	opts umOptions
}

func umCases() []umCase {
	return []umCase{
		{name: "text", in: `"abc"`, out: outOf(umText{})},
		{name: "text escaped", in: `"a\nb\u00e9"`, out: outOf(umText{})},
		{name: "text null", in: `null`, out: outOf(umText{S: "x"})},
		{name: "text number", in: `1`, out: outOf(umText{})},
		{name: "text object", in: `{"S":"x"}`, out: outOf(umText{})},
		{name: "text error", in: `"bad"`, out: outOf(umText{})},
		{name: "text error in an array", in: `["a","bad"]`, out: outOf([]umText(nil))},
		{name: "text invalid UTF-8", in: "\"a\xffb\"", out: outOf(umText{})},
		{name: "text bad escape", in: `["a","b\x"]`, out: outOf([]umText(nil))},
		{name: "text invalid UTF-8 allowed", in: "\"a\xffb\"", out: outOf(umText{}),
			opts: umOptions{ours: []json.Options{jsontext.AllowInvalidUTF8(true)}, std: []stdjson.Options{stdjsontext.AllowInvalidUTF8(true)}}},
		{name: "json", in: ` {"a" : [1, 2]} `, out: outOf(umJSON{})},
		{name: "json null", in: `null`, out: outOf(umJSON{Raw: "x"})},
		{name: "json error", in: `"bad"`, out: outOf(umJSON{})},
		{name: "json error in an object", in: `{"a":{"b":"bad"}}`, out: outOf(map[string]map[string]umJSON(nil))},
		{name: "json syntax error", in: `{"a":}`, out: outOf(umJSON{})},
		{name: "json duplicate name", in: `{"a":1,"a":2}`, out: outOf(umJSON{})},
		{name: "json duplicate name allowed", in: `{"a":1,"a":2}`, out: outOf(umJSON{}),
			opts: umOptions{ours: []json.Options{jsontext.AllowDuplicateNames(true)}, std: []stdjson.Options{stdjsontext.AllowDuplicateNames(true)}}},
		{name: "json unsupported", in: `1`, out: outOf(umUnsupported{})},
		{name: "both methods", in: `"x"`, out: outOf(umBoth{})},
		{name: "fields", in: `{"T":"a","P":"b","J":[1],"JP":2,"Tm":"2006-01-02T15:04:05Z","TP":"2006-01-02T15:04:05Z"}`, out: outOf(umFields{})},
		{name: "fields null", in: `{"T":null,"P":null,"J":null,"JP":null,"Tm":null,"TP":null}`,
			out: func() any {
				v := umFields{T: umText{S: "x"}, P: &umText{}, J: umJSON{Raw: "x"}, JP: &umJSON{}, Tm: time.Unix(1, 0).UTC(), TP: new(time.Time)}
				return &v
			}},
		{name: "fields error", in: `{"T":"a","J":"bad"}`, out: outOf(umFields{})},
		{name: "map of text keys", in: `{"a":1,"b":2}`, out: outOf(map[umTextKey]int(nil))},
		{name: "map of text keys escaped", in: `{"\u0061":1}`, out: outOf(map[umTextKey]int(nil))},
		{name: "map of text values", in: `{"a":"x","b":"bad"}`, out: outOf(map[string]umText(nil))},
		{name: "time", in: `"2006-01-02T15:04:05Z"`, out: outOf(time.Time{})},
		{name: "time of nanoseconds and a zone", in: `"2006-01-02T15:04:05.123456789+09:00"`, out: outOf(time.Time{})},
		{name: "time escaped", in: `"\u0032006-01-02T15:04:05Z"`, out: outOf(time.Time{})},
		{name: "time null", in: `null`, out: outOf(time.Unix(1, 0).UTC())},
		{name: "time number", in: `1`, out: outOf(time.Time{})},
		{name: "time not a time", in: `"x"`, out: outOf(time.Time{})},
		{name: "time with a space", in: `"2006-01-02 15:04:05Z"`, out: outOf(time.Time{})},
		{name: "time of a short hour", in: `"2006-01-02T1:04:05Z"`, out: outOf(time.Time{})},
		{name: "time of a comma", in: `"2006-01-02T15:04:05,5Z"`, out: outOf(time.Time{})},
		{name: "time of a zone hour out of range", in: `"2006-01-02T15:04:05+24:00"`, out: outOf(time.Time{})},
		{name: "time of a zone minute out of range", in: `"2006-01-02T15:04:05+09:60"`, out: outOf(time.Time{})},
		{name: "time of hour 24", in: `"2006-01-02T24:00:00Z"`, out: outOf(time.Time{})},
		{name: "time bad escape", in: `["2006\x"]`, out: outOf([]time.Time(nil))},
		{name: "time in an array", in: `["2006-01-02T15:04:05Z","x"]`, out: outOf([]time.Time(nil))},
		{name: "time of the string option", in: `{"T":"2006-01-02T15:04:05Z"}`, out: outOf(umTimeString{})},
		{name: "function", in: `[1,22,333]`, out: outOf([]int(nil)), opts: umLength},
		{name: "function of a field", in: `{"X":12}`, out: outOf(struct{ X int }{}), opts: umDouble},
		{name: "function error", in: `[1,2]`, out: outOf([]int(nil)), opts: umFailing},
		{name: "function error of its own", in: `{"X":"a"}`, out: outOf(map[string]int(nil)), opts: umDouble},
		{name: "function syntax error", in: `[[1],{"a":}]`, out: outOf([]any(nil)), opts: umFuncs(func(b []byte, v *[]any) error { return nil })},
		{name: "function duplicate name", in: `[{"a":1,"a":2}]`, out: outOf([]map[string]int(nil)), opts: umFuncs(func(b []byte, v *map[string]int) error { return nil })},
		{name: "function declining", in: `[1]`, out: outOf([]int(nil)), opts: umDeclining},
		{name: "function of a map key", in: `{"1":1}`, out: outOf(map[int]int(nil)), opts: umLength},
		{name: "function of map keys of a name twice", in: `{"a":1,"a":2}`, out: outOf(map[string]int(nil)), opts: umNumbering()},
		{name: "function of map keys of a name twice allowed", in: `{"a":1,"a":2}`, out: outOf(map[string]int(nil)),
			opts: umNumbering().and(jsontext.AllowDuplicateNames(true), stdjsontext.AllowDuplicateNames(true))},
		{name: "function before a method", in: `"x"`, out: outOf(umText{}), opts: umTextFunc},
		{name: "function of an interface", in: `{"T":"x","P":"y"}`, out: outOf(umFields{}), opts: umTextFunc},
		{name: "functions joined", in: `[12]`, out: outOf([]int(nil)), opts: umJoin(umDouble, umLength)},
		{name: "functions joined the other way", in: `[12]`, out: outOf([]int(nil)), opts: umJoin(umLength, umDouble)},
		{name: "function with numbers stringified", in: `["12"]`, out: outOf([]int(nil)),
			opts: umLength.and(json.StringifyNumbers(true), stdjson.StringifyNumbers(true))},
		{name: "function of strings into any", in: `["a",1,{"b":"c"},[true,"d"]]`, out: outOf[any](nil), opts: umStringFunc},
		{name: "function of floats into any", in: `[1,"a",{"b":2}]`, out: outOf[any](nil), opts: umFloatFunc},
		{name: "function of maps into any", in: `[{"a":1},2]`, out: outOf[any](nil), opts: umMapFunc},
		{name: "function of any", in: `[1,"a"]`, out: outOf([]any(nil)), opts: umAnyFunc},
		{name: "function of strings into a map of any", in: `{"a":"x","b":[1,"y"]}`, out: outOf(map[string]any(nil)), opts: umStringFunc},
		{name: "function of strings into an existing map of any", in: `{"a":"x"}`,
			out: func() any { v := map[string]any{"z": 1.0}; return &v }, opts: umStringFunc},
		{name: "function of any into a map of any", in: `{"a":1,"b":[2]}`, out: outOf(map[string]any(nil)), opts: umAnyFunc},
		{name: "function of any into an existing map of any", in: `{"a":1}`,
			out: func() any { v := map[string]any{"z": 1.0}; return &v }, opts: umAnyFunc},
		{name: "function of any into a field of a map of any", in: `{"M":{"a":1}}`, out: outOf(struct{ M map[string]any }{}), opts: umAnyFunc},
		{name: "functions of any and strings into a map of any", in: `{"a":1,"b":"s"}`, out: outOf(map[string]any(nil)),
			opts: umJoin(umAnyFunc, umStringFunc)},
		{name: "function of strings into any of names of one key", in: `{"a":1,"b":2}`, out: outOf[any](nil), opts: umOneName},
		{name: "function of strings into a map of any of names of one key", in: `{"a":1,"b":2}`, out: outOf(map[string]any(nil)), opts: umOneName},
		{name: "function of strings into any of a syntax error", in: `["a",x]`, out: outOf[any](nil), opts: umStringFunc},
		{name: "function of strings into any of a name twice", in: `{"a":1,"a":2}`, out: outOf[any](nil), opts: umStringFunc},
		{name: "function of strings into any of a name twice allowed", in: `{"a":1,"a":2}`, out: outOf[any](nil),
			opts: umStringFunc.and(jsontext.AllowDuplicateNames(true), stdjsontext.AllowDuplicateNames(true))},
		{name: "function error of a name into any", in: `{"a":{"bad":1}}`, out: outOf[any](nil), opts: umStringFailing},
		{name: "function of slices into any", in: `{"a":[1],"b":{"c":[2]}}`, out: outOf[any](nil), opts: umSliceFunc},
		{name: "function of maps into any of arrays", in: `[[{"a":1}],{"b":2}]`, out: outOf[any](nil), opts: umMapFunc},
		{name: "function of maps into any of a map", in: `{"a":1}`, out: func() any { var v any = map[string]any{"z": 1.0}; return &v }, opts: umMapFunc},
		{name: "function of slices into any of a slice", in: `[1]`, out: func() any { var v any = []any{2.0}; return &v }, opts: umSliceFunc},
	}
}

func TestUnmarshalMethodsMatchStd(t *testing.T) {
	for _, c := range umCases() {
		t.Run(c.name, func(t *testing.T) {
			ours, theirs := c.out(), c.out()
			err := json.Unmarshal([]byte(c.in), ours, c.opts.ours...)
			stdErr := stdjson.Unmarshal([]byte(c.in), theirs, c.opts.std...)
			if got, want := describeUnmarshalError(err), describeUnmarshalError(stdErr); got != want {
				t.Fatalf("error:\n got %s\nwant %s", got, want)
			}
			if err == nil && !reflect.DeepEqual(ours, theirs) {
				t.Fatalf("value:\n got %#v\nwant %#v", reflect.ValueOf(ours).Elem(), reflect.ValueOf(theirs).Elem())
			}
		})
	}
}
