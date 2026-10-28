//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	"encoding"
	stdjsonv1 "encoding/json"
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json, of the root package and of encoding/json, given to the functions of both v2
// packages in the differential tests.

// v1OptionSet is a set of the options of encoding/json, a bit of each in the order of v1OptionPairs.
type v1OptionSet uint16

const (
	parseBytesWithLooseRFC4648 v1OptionSet = 1 << iota
	parseTimeWithLooseRFC3339
	unmarshalArrayFromAnyLength
	matchCaseSensitiveDelimiter
	mergeWithLegacySemantics
	callMethodsWithLegacySemantics
	formatByteArrayAsArray
	formatBytesWithLegacySemantics
	formatDurationAsNano
	reportErrorsWithLegacySemantics
	stringifyWithLegacySemantics
	omitEmptyWithLegacySemantics

	// v1OptionCount is the number of the options above.
	v1OptionCount = iota
	// allV1Options is every option, as DefaultOptionsV1 sets them.
	allV1Options v1OptionSet = 1<<v1OptionCount - 1
)

// every option has a pair: the length of the array type is 0 then, and it is another type, or a negative constant,
// otherwise.
var _ [0]struct{} = [len(v1OptionPairs) - v1OptionCount]struct{}{}

// v1OptionPairs are the options of the root package and of encoding/json, by the bits of v1OptionSet.
var v1OptionPairs = [...]struct {
	name string
	ours func(bool) json.Options
	std  func(bool) stdjson.Options
}{
	{"ParseBytesWithLooseRFC4648", gojson.ParseBytesWithLooseRFC4648, stdjsonv1.ParseBytesWithLooseRFC4648},
	{"ParseTimeWithLooseRFC3339", gojson.ParseTimeWithLooseRFC3339, stdjsonv1.ParseTimeWithLooseRFC3339},
	{"UnmarshalArrayFromAnyLength", gojson.UnmarshalArrayFromAnyLength, stdjsonv1.UnmarshalArrayFromAnyLength},
	{"MatchCaseSensitiveDelimiter", gojson.MatchCaseSensitiveDelimiter, stdjsonv1.MatchCaseSensitiveDelimiter},
	{"MergeWithLegacySemantics", gojson.MergeWithLegacySemantics, stdjsonv1.MergeWithLegacySemantics},
	{"CallMethodsWithLegacySemantics", gojson.CallMethodsWithLegacySemantics, stdjsonv1.CallMethodsWithLegacySemantics},
	{"FormatByteArrayAsArray", gojson.FormatByteArrayAsArray, stdjsonv1.FormatByteArrayAsArray},
	{"FormatBytesWithLegacySemantics", gojson.FormatBytesWithLegacySemantics, stdjsonv1.FormatBytesWithLegacySemantics},
	{"FormatDurationAsNano", gojson.FormatDurationAsNano, stdjsonv1.FormatDurationAsNano},
	{"ReportErrorsWithLegacySemantics", gojson.ReportErrorsWithLegacySemantics, stdjsonv1.ReportErrorsWithLegacySemantics},
	{"StringifyWithLegacySemantics", gojson.StringifyWithLegacySemantics, stdjsonv1.StringifyWithLegacySemantics},
	{"OmitEmptyWithLegacySemantics", gojson.OmitEmptyWithLegacySemantics, stdjsonv1.OmitEmptyWithLegacySemantics},
}

// String names the options of the set.
func (s v1OptionSet) String() string {
	var names []string
	for i, p := range v1OptionPairs {
		if s&(1<<i) != 0 {
			names = append(names, p.name)
		}
	}
	return "{" + strings.Join(names, ", ") + "}"
}

func (s v1OptionSet) ours() []json.Options {
	var opts []json.Options
	for i, p := range v1OptionPairs {
		if s&(1<<i) != 0 {
			opts = append(opts, p.ours(true))
		}
	}
	return opts
}

func (s v1OptionSet) std() []stdjson.Options {
	var opts []stdjson.Options
	for i, p := range v1OptionPairs {
		if s&(1<<i) != 0 {
			opts = append(opts, p.std(true))
		}
	}
	return opts
}

// v1DiffOptions are the rows of the options of encoding/json in the marshal differential tests ( see diffOptions ).
func v1DiffOptions() [][]json.Options {
	return [][]json.Options{
		{gojson.FormatDurationAsNano(true)},
		{gojson.FormatDurationAsNano(true), json.StringifyNumbers(true), json.Deterministic(true)},
		{gojson.FormatByteArrayAsArray(true)},
		{gojson.FormatBytesWithLegacySemantics(true)},
		{gojson.FormatByteArrayAsArray(true), gojson.FormatBytesWithLegacySemantics(true), json.StringifyNumbers(true),
			json.FormatNilSliceAsNull(true), json.Deterministic(true)},
		{gojson.OmitEmptyWithLegacySemantics(true)},
		{json.MatchCaseInsensitiveNames(true), json.Deterministic(true)},
		{json.MatchCaseInsensitiveNames(true), gojson.MatchCaseSensitiveDelimiter(true), json.Deterministic(true)},
		{gojson.MatchCaseSensitiveDelimiter(true), json.Deterministic(true)},
		{gojson.StringifyWithLegacySemantics(true)},
		{gojson.CallMethodsWithLegacySemantics(true)},
		{gojson.CallMethodsWithLegacySemantics(true), json.Deterministic(true)},
		{gojson.StringifyWithLegacySemantics(true), json.StringifyNumbers(true), gojson.OmitEmptyWithLegacySemantics(true)},
		{gojson.OmitEmptyWithLegacySemantics(true), json.StringifyNumbers(true), json.OmitZeroStructFields(true),
			json.Deterministic(true)},
	}
}

// v1DiffDelimiter is a struct of the names which MatchCaseSensitiveDelimiter matches or not.
type v1DiffDelimiter struct {
	FooBar int `json:"foo_bar"`
	Baz    int `json:"BAZ,case:ignore"`
	QuxX   int `json:"qux-x,case:ignore"`
	Strict int `json:"strict_s,case:strict"`
}

// v1DiffDelimiterTwo is a struct of two names which are the same but for an underscore.
type v1DiffDelimiterTwo struct {
	AB  int `json:"a_b"`
	AB2 int `json:"ab"`
}

// v1DiffDelimiterIgnore is a struct of two names of case:ignore which are the same but for an underscore.
type v1DiffDelimiterIgnore struct {
	AB  int `json:"a_b,case:ignore"`
	AB2 int `json:"ab,case:ignore"`
}

// v1DiffDelimiterFallback is a struct of an embedded fallback, whose names are matched with the names of the fields when
// marshaling.
type v1DiffDelimiterFallback struct {
	FooBar int            `json:"foo_bar"`
	Baz    int            `json:"BAZ,case:ignore"`
	Rest   map[string]int `json:",embed"`
}

// v1DiffStringify is a struct of the fields of the `string` option of StringifyWithLegacySemantics.
type v1DiffStringify struct {
	B  bool    `json:",string"`
	S  string  `json:",string"`
	I  int     `json:",string"`
	U  uint8   `json:",string"`
	F  float64 `json:",string"`
	PB *bool   `json:",string"`
	PS *string `json:",string"`
	PI *int    `json:",string"`
	E  string  `json:",omitempty,string"`
	EB bool    `json:",omitempty,string"`
}

// v1DiffStringifyPtr is a struct of a pointer to a pointer of the `string` option, which the option doesn't apply to
// with StringifyWithLegacySemantics.
type v1DiffStringifyPtr struct {
	PP **int `json:",string"`
}

// v1TextBool is a named bool whose UnmarshalText fails for "null" and sets it true for any other text.
type v1TextBool bool

func (b *v1TextBool) UnmarshalText(text []byte) error {
	if string(text) == "null" {
		return errV1Null
	}
	*b = true
	return nil
}

// errV1Null is the error of v1TextBool, an error of encoding/json, which both packages report as it is.
var errV1Null error = &stdjsonv1.UnsupportedValueError{Str: "null"}

// v1DiffStringifyPointers is a struct of pointers of the `string` option to values which the option doesn't apply to, and
// of a pointer of the option and another one to the same string.
type v1DiffStringifyPointers struct {
	PSl *[]int          `json:",string"`
	PSt *v1Pair         `json:",string"`
	PM  *map[string]int `json:",string"`
	PT  *v1TextBool     `json:",string"`
	P   *string         `json:",string"`
	A   *string
}

// v1DiffStringifyText is a struct of a string of the `string` option, of the texts which a JSON string escapes.
type v1DiffStringifyText struct {
	S string `json:",string"`
}

// v1Pair is a struct of two fields, which a JSON object of one of them merges into or not.
type v1Pair struct{ X, Y int }

// v1DiffMerge is a struct of the values which MergeWithLegacySemantics merges into or not.
type v1DiffMerge struct {
	B   bool
	I   int
	U   uint
	F   float64
	S   string
	A   [2]int
	AP  [2]v1Pair
	St  v1Pair
	Sl  []v1Pair
	M   map[string]v1Pair
	MA  map[string]any
	P   *int
	T   time.Time
	Tx  v1TextInByte
	BA  [2]byte
	Bs  []byte
	Any any
}

// newV1DiffMerge returns a value of v1DiffMerge whose fields are not zero.
func newV1DiffMerge() any {
	n := 9
	sl := make([]v1Pair, 2, 3)
	sl[0], sl[1] = v1Pair{1, 2}, v1Pair{3, 4}
	return &v1DiffMerge{B: true, I: 1, U: 2, F: 3, S: "s", A: [2]int{4, 5}, AP: [2]v1Pair{{1, 2}, {3, 4}},
		St: v1Pair{6, 7}, Sl: sl, M: map[string]v1Pair{"a": {1, 2}}, MA: map[string]any{"a": map[string]any{"x": 1.0}, "p": new(int)},
		P: &n, T: time.Unix(1, 0).UTC(), Tx: 3, BA: [2]byte{1, 2}, Bs: []byte{1}, Any: map[string]any{"k": 1.0}}
}

// v1CallHolder has values of v1PtrMarshaler in every kind of place.
type v1CallHolder struct {
	V v1PtrMarshaler
	P *v1PtrMarshaler
	A [1]v1PtrMarshaler
	S []v1PtrMarshaler
	M map[string]v1PtrMarshaler
	I any
	T v1PtrText
	N v1Node
}

// v1IntJSON is an int whose MarshalJSON writes a string, which is not called for a key with
// CallMethodsWithLegacySemantics.
type v1IntJSON int

func (v1IntJSON) MarshalJSON() ([]byte, error) { return []byte(`"json"`), nil }

// v1CallKeys has maps of the keys whose methods CallMethodsWithLegacySemantics calls or not, and interface values of
// the methods of marshaling holding nil pointers.
type v1CallKeys struct {
	V  map[v1ValueText]int
	S  map[v1StringText]int
	J  map[v1IntJSON]int
	P  map[*v1PtrText]int
	PV map[*v1ValueText]int
	M  stdjsonv1.Marshaler
	T  fmt.Stringer
}

// v1KeyJSON is a string of UnmarshalJSON only, which CallMethodsWithLegacySemantics doesn't call for a key.
type v1KeyJSON string

func (k *v1KeyJSON) UnmarshalJSON([]byte) error {
	*k = "json"
	return nil
}

// v1PtrAppend is a struct whose AppendText has a pointer receiver, which writes "nil" for a nil pointer.
type v1PtrAppend struct{ X int }

func (p *v1PtrAppend) AppendText(b []byte) ([]byte, error) {
	if p == nil {
		return append(b, "nil"...), nil
	}
	return append(b, "append"...), nil
}

// v1StringCall has fields of the `string` option of v1PtrMarshaler, whose method is called if the struct is
// addressable or through the pointer.
type v1StringCall struct {
	V v1PtrMarshaler  `json:",string"`
	P *v1PtrMarshaler `json:",string"`
}

// v1PtrTextMap is a map whose MarshalText has a pointer receiver.
type v1PtrTextMap map[string]int

func (*v1PtrTextMap) MarshalText() ([]byte, error) { return []byte("text"), nil }

// v1PtrOnePtr is a struct of one pointer, which is stored directly in an interface value, whose MarshalJSON has a
// pointer receiver.
type v1PtrOnePtr struct{ P *int }

func (*v1PtrOnePtr) MarshalJSON() ([]byte, error) { return []byte(`"one"`), nil }

// v1PtrOneArray is an array of one pointer, which is stored directly in an interface value, whose MarshalJSON has
// a pointer receiver.
type v1PtrOneArray [1]*int

func (*v1PtrOneArray) MarshalJSON() ([]byte, error) { return []byte(`"array"`), nil }

// v1PtrSlice is a slice whose MarshalJSON has a pointer receiver.
type v1PtrSlice []int

func (*v1PtrSlice) MarshalJSON() ([]byte, error) { return []byte(`"slice"`), nil }

// v1RecArray is an array of a pointer to itself, whose MarshalJSON has a pointer receiver.
type v1RecArray [1]*v1RecArray

func (*v1RecArray) MarshalJSON() ([]byte, error) { return []byte(`"a"`), nil }

// v1RecMap is a map of slices of itself, whose MarshalJSON has a pointer receiver.
type v1RecMap map[string][]v1RecMap

func (*v1RecMap) MarshalJSON() ([]byte, error) { return []byte(`"m"`), nil }

// v1RecText is a slice of itself, whose MarshalText has a pointer receiver.
type v1RecText []v1RecText

func (*v1RecText) MarshalText() ([]byte, error) { return []byte("t"), nil }

// v1RecPlain is a slice of itself of no method, whose cycle is the error of its type.
type v1RecPlain []v1RecPlain

// v1StringCallPtr has a field of the `string` option of a pointer to v1PtrMarshaler, whose method is called
// for the value it points to, which is addressable.
type v1StringCallPtr struct {
	P *v1PtrMarshaler `json:",string"`
}

// v1EmbeddedCall embeds v1PtrMarshaler, whose method of a pointer is the one of a pointer of the struct.
type v1EmbeddedCall struct{ v1PtrMarshaler }

// v1StdDeclined is v1Declined for encoding/json/v2.
type v1StdDeclined struct{ X int }

func (*v1StdDeclined) MarshalJSONTo(*stdjsontext.Encoder) error { return errors.ErrUnsupported }

func (*v1StdDeclined) MarshalJSON() ([]byte, error) { return []byte(`"json"`), nil }

// With CallMethodsWithLegacySemantics, as encoding/json/v2 writes them, through every function of marshaling: with
// functions of marshaling, the value which a pointer at the top level points to is addressable; the value after a
// method or a function which declines it is addressable as the value is; a key of an interface type of MarshalText or
// AppendText which holds a nil pointer is named by the method, called for the nil pointer.
func TestCallMethodsWithLegacySemanticsMatchesStd(t *testing.T) {
	// a function of another type, which makes the functions of marshaling be looked up
	other := []json.Options{json.WithMarshalers(json.MarshalFunc(func(time.Month) ([]byte, error) { return []byte(`"m"`), nil }))}
	stdOther := []stdjson.Options{stdjson.WithMarshalers(stdjson.MarshalFunc(func(time.Month) ([]byte, error) { return []byte(`"m"`), nil }))}
	// a function which declines every value of v1PtrMarshaler
	declines := []json.Options{json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, v1PtrMarshaler) error {
		return errors.ErrUnsupported
	}))}
	stdDeclines := []stdjson.Options{stdjson.WithMarshalers(stdjson.MarshalToFunc(func(*stdjsontext.Encoder, v1PtrMarshaler) error {
		return errors.ErrUnsupported
	}))}
	for _, c := range []struct {
		name         string
		ours, theirs any
		// the options given with CallMethodsWithLegacySemantics
		opts    []json.Options
		stdOpts []stdjson.Options
	}{
		{"pointer", &v1PtrMarshaler{}, &v1PtrMarshaler{}, other, stdOther},
		{"value", v1PtrMarshaler{}, v1PtrMarshaler{}, other, stdOther},
		{"pointer of an array", &[1]v1PtrMarshaler{}, &[1]v1PtrMarshaler{}, other, stdOther},
		{"pointer of a struct", &v1StringCall{}, &v1StringCall{}, other, stdOther},
		{"pointer of an embedding struct", &v1EmbeddedCall{}, &v1EmbeddedCall{}, other, stdOther},
		{"recursive", &v1Node{Next: &v1Node{}}, &v1Node{Next: &v1Node{}}, other, stdOther},
		{"declined pointer", &v1Declined{}, &v1StdDeclined{}, nil, nil},
		{"declined value", v1Declined{}, v1StdDeclined{}, nil, nil},
		{"declined elements", []v1Declined{{}}, []v1StdDeclined{{}}, nil, nil},
		{"declined values of a map", map[string]v1Declined{"a": {}}, map[string]v1StdDeclined{"a": {}}, nil, nil},
		{"declined by a function, pointer", &v1PtrMarshaler{}, &v1PtrMarshaler{}, declines, stdDeclines},
		{"declined by a function, value", v1PtrMarshaler{}, v1PtrMarshaler{}, declines, stdDeclines},
		{"declined by a function, elements", []v1PtrMarshaler{{}}, []v1PtrMarshaler{{}}, declines, stdDeclines},
		{"declined by a function, string option", &v1StringCall{}, &v1StringCall{}, declines, stdDeclines},
		{"declined by a function, string option of a pointer", v1StringCallPtr{P: &v1PtrMarshaler{}},
			v1StringCallPtr{P: &v1PtrMarshaler{}}, declines, stdDeclines},
		{"recursive with a function of pointers to pointers", v1RecSlice{v1RecSlice{}, nil}, v1RecSlice{v1RecSlice{}, nil},
			[]json.Options{json.WithMarshalers(json.MarshalFunc(func(**v1RecSlice) ([]byte, error) { return []byte(`"pp"`), nil }))},
			[]stdjson.Options{stdjson.WithMarshalers(stdjson.MarshalFunc(func(**v1RecSlice) ([]byte, error) { return []byte(`"pp"`), nil }))}},
		{"recursive with a declining function of any", v1RecSlice{v1RecSlice{}, nil}, v1RecSlice{v1RecSlice{}, nil},
			[]json.Options{json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, any) error { return errors.ErrUnsupported }))},
			[]stdjson.Options{stdjson.WithMarshalers(stdjson.MarshalToFunc(func(*stdjsontext.Encoder, any) error { return errors.ErrUnsupported }))}},
		{"key of a text interface of a nil pointer", map[encoding.TextMarshaler]int{(*v1PtrText)(nil): 1},
			map[encoding.TextMarshaler]int{(*v1PtrText)(nil): 1}, nil, nil},
		{"key of an appender interface of a nil pointer", map[encoding.TextAppender]int{(*v1PtrAppend)(nil): 1},
			map[encoding.TextAppender]int{(*v1PtrAppend)(nil): 1}, nil, nil},
		{"key of a JSON interface of a nil pointer", map[stdjsonv1.Marshaler]int{(*v1PtrMarshaler)(nil): 1},
			map[stdjsonv1.Marshaler]int{(*v1PtrMarshaler)(nil): 1}, nil, nil},
	} {
		want, wantErr := stdjson.Marshal(c.theirs, append(callMethodsWithLegacySemantics.std(), c.stdOpts...)...)
		for _, m := range marshalFuncs {
			// of an error, only that both fail is compared: the error of a key of an interface type which holds a nil
			// pointer is of another type ( see issues/encoder-small-deviations.md ), and the output of encoding/json/v2
			// is what it wrote before
			got, err := m.marshal(c.ours, append(callMethodsWithLegacySemantics.ours(), c.opts...)...)
			if (err == nil) != (wantErr == nil) || err == nil && got != string(want) {
				t.Errorf("%s: %s:\n got %s, %v\nwant %s, %v", c.name, m.name, got, err, want, wantErr)
			}
		}
	}
}

// v1LargeCall is a struct of many fields, the type of the fields of v1SharedCall.
type v1LargeCall struct {
	F00 int
	F01 int
	F02 int
	F03 int
	F04 int
	F05 int
	F06 int
	F07 int
	F08 int
	F09 int
	F10 int
	F11 int
	F12 int
	F13 int
	F14 int
	F15 int
	F16 int
	F17 int
	F18 int
	F19 int
	F20 int
	F21 int
	F22 int
	F23 int
	F24 int
	F25 int
	F26 int
	F27 int
	F28 int
	F29 int
	F30 int
	F31 int
	F32 int
	F33 int
	F34 int
	F35 int
	F36 int
	F37 int
	F38 int
	F39 int
	V   v1PtrMarshaler
}

// v1SharedCall has more fields than sharedStructFieldCount of the encoder, with the ones of the structs in it: a value
// which has it again is written by a jump to one code of it, addressable or not ( see v1SharedCalls ).
type v1SharedCall struct{ A, B, C, D v1LargeCall }

// v1SharedCalls has v1SharedCall addressable or not, as v1SharedCalls is or not, then addressable, through the
// pointer, and not, in the map.
type v1SharedCalls struct {
	X v1SharedCall
	P *v1SharedCall
	M map[string]v1SharedCall
}

// v1EmptyText is a struct whose method writes an empty JSON string: never empty by its Go value.
type v1EmptyText struct{ X int }

func (v1EmptyText) MarshalJSON() ([]byte, error) { return []byte(`""`), nil }

// v1Raw is a slice whose method writes an empty JSON string for no bytes: empty by its Go value if it has none.
type v1Raw []byte

func (r v1Raw) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte(`""`), nil
	}
	return r, nil
}

// v1DiffOmitEmpty is a struct of the fields of omitempty of OmitEmptyWithLegacySemantics.
type v1DiffOmitEmpty struct {
	B   bool            `json:",omitempty"`
	I   int             `json:",omitempty"`
	U   uint8           `json:",omitempty"`
	F   float64         `json:",omitempty"`
	S   string          `json:",omitempty"`
	P   *int            `json:",omitempty"`
	PP  **int           `json:",omitempty"`
	PS  *[]int          `json:",omitempty"`
	PM  *map[string]int `json:",omitempty"`
	PT  *struct{}       `json:",omitempty"`
	Sl  []int           `json:",omitempty"`
	Bs  []byte          `json:",omitempty"`
	M   map[string]int  `json:",omitempty"`
	A0  [0]int          `json:",omitempty"`
	A1  [1]int          `json:",omitempty"`
	St  struct{}        `json:",omitempty"`
	Any any             `json:",omitempty"`
	Raw v1Raw           `json:",omitempty"`
	MJ  v1EmptyText     `json:",omitempty"`
	PMJ *v1EmptyText    `json:",omitempty"`
	Z   int             `json:",omitempty,omitzero"`
	N   map[string]any  `json:",omitempty"`
	E   v1EmbeddedOmit  `json:",omitempty"`
}

// v1EmbeddedOmit is a struct of a field of omitempty, which is never empty itself.
type v1EmbeddedOmit struct {
	X []int `json:",omitempty"`
}

// v1DiffOmitEmptyString is a struct of the fields of omitempty with the string option, which a bool and a string fail
// by when they are written, without StringifyWithLegacySemantics.
type v1DiffOmitEmptyString struct {
	B bool    `json:",omitempty,string"`
	S string  `json:",omitempty,string"`
	I int     `json:",omitempty,string"`
	P *string `json:",omitempty,string"`
}

// v1DiffDuration is a struct of the time.Duration values of FormatDurationAsNano.
type v1DiffDuration struct {
	D    time.Duration
	P    *time.Duration
	S    time.Duration            `json:",string"`
	SP   *time.Duration           `json:",string"`
	E    time.Duration            `json:",omitempty"`
	Z    time.Duration            `json:",omitzero"`
	Keys map[time.Duration]string `json:",omitempty"`
	Any  any
}

// v1DiffBytes is a struct of the bytes of FormatByteArrayAsArray and FormatBytesWithLegacySemantics.
type v1DiffBytes struct {
	B   []byte
	NB  []v1Byte
	TB  []v1TextByte
	A   [2]byte
	NA  [2]v1Byte
	TA  [2]v1TextByte
	E   []v1Byte  `json:",omitempty"`
	EA  [2]v1Byte `json:",omitempty"`
	Z   [2]byte   `json:",omitzero"`
	P   *[2]byte
	Any any
}

// v1PtrJSONByte is a named byte type whose MarshalJSON has a pointer receiver.
type v1PtrJSONByte byte

func (*v1PtrJSONByte) MarshalJSON() ([]byte, error) { return []byte(`"j"`), nil }

// v1AppendByte is a named byte type of AppendText.
type v1AppendByte byte

func (v1AppendByte) AppendText(dst []byte) ([]byte, error) { return append(dst, 'a'), nil }

// v1TextInByte is a named byte type of UnmarshalText only: its slices and arrays are of bytes with
// FormatBytesWithLegacySemantics when marshaling too.
type v1TextInByte byte

func (b *v1TextInByte) UnmarshalText(text []byte) error {
	*b = v1TextInByte(len(text))
	return nil
}

// v1JSONInByte is a named byte type of UnmarshalJSON only.
type v1JSONInByte byte

func (b *v1JSONInByte) UnmarshalJSON(data []byte) error {
	*b = v1JSONInByte(len(data))
	return nil
}

// v1NamedBytes is a named slice type of a named byte type.
type v1NamedBytes []v1Byte

// v1DiffValues are the values of the options of encoding/json in the marshal differential tests ( see
// diffValues ).
func v1DiffValues() []any {
	d := time.Duration(-3)
	n := 3
	pm, ps := v1PtrMap{"k": 1}, v1PtrSlice{1}
	ppm := &pm
	cyclic := make(v1RecPlain, 1)
	cyclic[0] = cyclic
	return []any{
		time.Duration(1500), &d, []time.Duration{1, 2}, map[time.Duration]int{5: 1, -2: 2}, map[string]time.Duration{"a": 7},
		v1DiffDuration{}, v1DiffDuration{D: 1, P: &d, S: 2, SP: &d, E: 3, Z: 4, Keys: map[time.Duration]string{9: "x"}, Any: d},
		[]byte{1, 2, 3}, []byte(nil), [3]byte{1, 2, 3}, [0]byte{}, []v1Byte{1, 2, 3}, []v1Byte(nil), [3]v1Byte{1, 2, 3},
		[]v1TextByte{1, 2}, [2]v1TextByte{1, 2},
		map[[2]byte]int{{1, 2}: 1}, map[[2]v1Byte]int{{1, 2}: 1}, map[[2]v1TextByte]int{{1, 2}: 1},
		[]v1PtrJSONByte{1}, [1]v1PtrJSONByte{1}, []v1AppendByte{1}, [1]v1AppendByte{1},
		[]v1TextInByte{1, 2}, []v1JSONInByte{1, 2}, v1NamedBytes{1, 2}, v1NamedBytes(nil),
		[2][2]byte{{1, 2}, {3, 4}}, &[2]byte{1, 2}, &[]v1Byte{1},
		v1DiffOmitEmpty{}, v1DiffOmitEmpty{P: new(int), PP: new(*int), PS: &[]int{}, PM: &map[string]int{},
			PT: &struct{}{}, Sl: []int{}, Bs: []byte{}, M: map[string]int{}, Any: "", Raw: v1Raw{}, PMJ: &v1EmptyText{},
			N: map[string]any{}},
		v1DiffOmitEmpty{B: true, I: -1, U: 1, F: 0.5, S: "x", Sl: []int{1}, Any: 0, Raw: v1Raw(`1`), Z: 1,
			N: map[string]any{"a": nil}, E: v1EmbeddedOmit{X: []int{}}},
		v1DiffOmitEmptyString{}, v1DiffOmitEmptyString{I: 3},
		v1PtrMarshaler{}, &v1PtrMarshaler{}, [1]v1PtrMarshaler{}, &[1]v1PtrMarshaler{}, []v1PtrMarshaler{{}},
		v1Node{Next: &v1Node{Next: &v1Node{}}}, &v1Node{Next: &v1Node{}}, []v1Node{{Next: &v1Node{}}},
		map[string]v1Node{"a": {Next: &v1Node{}}},
		v1CallHolder{P: &v1PtrMarshaler{}, S: []v1PtrMarshaler{{}}, M: map[string]v1PtrMarshaler{"a": {}},
			I: v1PtrMarshaler{}, N: v1Node{Next: &v1Node{}}},
		&v1CallHolder{P: &v1PtrMarshaler{}, S: []v1PtrMarshaler{{}}, M: map[string]v1PtrMarshaler{"a": {}},
			I: &v1PtrMarshaler{}, N: v1Node{Next: &v1Node{}}},
		v1SharedCalls{P: &v1SharedCall{}, M: map[string]v1SharedCall{"a": {}}},
		&v1SharedCalls{P: &v1SharedCall{}, M: map[string]v1SharedCall{"a": {}}},
		v1CallKeys{V: map[v1ValueText]int{{}: 1}, S: map[v1StringText]int{"s": 1}, J: map[v1IntJSON]int{3: 1},
			P: map[*v1PtrText]int{{}: 1}, PV: map[*v1ValueText]int{{}: 1}, M: (*v1PtrMarshaler)(nil), T: (*net.IPNet)(nil)},
		v1CallKeys{P: map[*v1PtrText]int{nil: 1}, M: &v1PtrMarshaler{}},
		map[*v1ValueText]int{nil: 1}, []stdjsonv1.Marshaler{(*v1PtrMarshaler)(nil), nil, &v1PtrMarshaler{}, v1IntJSON(1)},
		map[encoding.TextMarshaler]int{&v1PtrText{}: 1}, v1StringCall{P: &v1PtrMarshaler{}}, &v1StringCall{},
		v1PtrMap{"k": 1}, &v1PtrMap{"k": 1}, map[string]v1PtrMap{"a": {"k": 1}}, []v1PtrMap{{"k": 1}},
		struct{ A [1]v1PtrMap }{[1]v1PtrMap{{"k": 1}}}, &struct{ A [1]v1PtrMap }{[1]v1PtrMap{{"k": 1}}},
		[]any{v1PtrMap{"k": 1}}, struct{ M v1PtrMap }{v1PtrMap{"k": 1}}, v1PtrTextMap{"k": 1},
		map[string]v1PtrTextMap{"a": {"k": 1}}, []any{v1PtrTextMap{"k": 1}}, v1PtrOnePtr{&n},
		map[string]v1PtrOnePtr{"a": {&n}}, struct{ A [1]v1PtrOnePtr }{[1]v1PtrOnePtr{{&n}}},
		[]any{v1PtrOnePtr{&n}}, v1PtrOneArray{&n}, map[string]v1PtrOneArray{"a": {&n}}, []any{v1PtrOneArray{&n}},
		struct{ F *v1PtrMap }{&pm}, []*v1PtrMap{&pm}, map[string]*v1PtrMap{"a": &pm}, &ppm,
		struct{ F *v1PtrSlice }{&ps}, []*v1PtrSlice{&ps}, &ps, struct{ F *v1PtrMap }{},
		v1RecSlice{v1RecSlice{}, nil}, &v1RecSlice{v1RecSlice{}}, map[string]v1RecSlice{"a": {nil}},
		v1RecArray{&v1RecArray{}}, v1RecMap{"a": {v1RecMap{}}}, &v1RecMap{"a": {nil}},
		v1RecText{v1RecText{}, nil}, cyclic, maxDepthAddressable{{Next: []maxDepthAddressable{{{}}}}},
		&maxDepthAddressable{{Next: []maxDepthAddressable{{{}}}}},
		[]v1StringCall{{}}, v1EmbeddedCall{}, &v1EmbeddedCall{},
		v1DiffStringify{}, v1DiffStringify{B: true, S: "x", I: -5, U: 7, F: 1.5, PB: new(bool), PS: new(string), PI: new(int),
			E: "e", EB: true},
		v1DiffStringifyPtr{}, v1DiffStringifyPtr{PP: new(*int)},
		v1DiffStringifyText{S: "q\"b\\<&>\u2028"},
		// invalid UTF-8, whose error is a SemanticError of the type of the string, of a SyntacticError, as encoding/json/v2
		// reports it with StringifyWithLegacySemantics
		v1DiffStringifyText{S: "\xff"}, v1DiffStringifyText{S: "x\xff"},
		v1DiffDelimiterFallback{FooBar: 1, Rest: map[string]int{"foobar": 2}},
		v1DiffDelimiterFallback{FooBar: 1, Rest: map[string]int{"FOO_BAR": 2}},
		v1DiffDelimiterFallback{Baz: 1, Rest: map[string]int{"baz": 2}},
		v1DiffDelimiterFallback{Rest: map[string]int{"b-a-z": 2, "x": 3}},
		v1DiffBytes{}, v1DiffBytes{B: []byte{1}, NB: []v1Byte{2}, TB: []v1TextByte{3}, A: [2]byte{4, 5},
			NA: [2]v1Byte{6, 7}, TA: [2]v1TextByte{1, 2}, E: []v1Byte{}, P: &[2]byte{8, 9}, Any: []v1Byte{1, 2}},
	}
}

// unmarshalDiffV1Cases are the cases of the options of encoding/json.
func unmarshalDiffV1Cases() []unmarshalDiffCase {
	anyLength := unmarshalDiffOptions{v1Options: unmarshalArrayFromAnyLength}
	loose := unmarshalDiffOptions{v1Options: parseBytesWithLooseRFC4648}
	rfc := unmarshalDiffOptions{v1Options: parseTimeWithLooseRFC3339}
	nano := unmarshalDiffOptions{v1Options: formatDurationAsNano}
	nanoStringify := unmarshalDiffOptions{v1Options: formatDurationAsNano, stringifyNumbers: true}
	arrayOfBytes := unmarshalDiffOptions{v1Options: formatByteArrayAsArray}
	v1Bytes := unmarshalDiffOptions{v1Options: formatBytesWithLegacySemantics}
	bothBytes := unmarshalDiffOptions{v1Options: formatByteArrayAsArray | formatBytesWithLegacySemantics}
	bytesStruct := `{"B":[1],"NB":"AQ==","TB":["b"],"A":"AQI=","NA":[1,2],"TA":"AQI=","E":[],"P":[1,2],"Any":"AQ=="}`
	callMethods := unmarshalDiffOptions{v1Options: callMethodsWithLegacySemantics}
	merge := unmarshalDiffOptions{v1Options: mergeWithLegacySemantics}
	mergeNulls := `{"B":null,"I":null,"U":null,"F":null,"S":null,"A":null,"AP":null,"St":null,"Sl":null,"M":null,"MA":null,` +
		`"P":null,"T":null,"Tx":null,"BA":null,"Bs":null,"Any":null}`
	mergeValues := `{"A":[8,9],"AP":[{"X":9},{"Y":8}],"St":{"X":9},"Sl":[{"X":9}],"M":{"a":{"X":9}},"MA":{"a":{"y":2},"p":5},` +
		`"Any":{"j":2},"BA":"AwQ=","Tx":"bb"}`
	v1Stringify := unmarshalDiffOptions{v1Options: stringifyWithLegacySemantics}
	newStringify := func() any { return new(v1DiffStringify) }
	delimiter := unmarshalDiffOptions{v1Options: matchCaseSensitiveDelimiter}
	both := unmarshalDiffOptions{v1Options: matchCaseSensitiveDelimiter, matchCaseInsensitiveNames: true}
	newDelimiter := func() any { return new(v1DiffDelimiter) }
	newDelimiterTwo := func() any { return new(v1DiffDelimiterTwo) }
	return []unmarshalDiffCase{
		// CallMethodsWithLegacySemantics
		{name: "call methods of keys", in: `{"a":1}`, out: func() any { return new(map[v1KeyText]int) }, opts: callMethods},
		{name: "call methods of keys by default", in: `{"a":1}`, out: func() any { return new(map[v1KeyText]int) }},
		{name: "call methods of keys of JSON only", in: `{"a":1}`, out: func() any { return new(map[v1KeyJSON]int) }, opts: callMethods},
		{name: "call methods of keys of JSON only by default", in: `{"a":1}`, out: func() any { return new(map[v1KeyJSON]int) }},
		{name: "call methods of pointer keys", in: `{"a":1}`, out: func() any { return new(map[*v1KeyText]int) }, opts: callMethods,
			errorOnly: true},
		{name: "call methods of values", in: `{"a":"x"}`, out: func() any { return new(map[string]v1KeyText) }, opts: callMethods},
		// MergeWithLegacySemantics
		{name: "merge nulls", in: mergeNulls, out: newV1DiffMerge, opts: merge},
		{name: "merge nulls by default", in: mergeNulls, out: newV1DiffMerge},
		{name: "merge values", in: mergeValues, out: newV1DiffMerge, opts: merge},
		{name: "merge values by default", in: mergeValues, out: newV1DiffMerge},
		{name: "merge values of duplicate names", in: `{"MA":{"a":{"x":2},"a":{"y":3}},"M":{"a":{"Y":1},"a":{"X":5}}}`,
			out: newV1DiffMerge, opts: unmarshalDiffOptions{v1Options: mergeWithLegacySemantics, allowDuplicateNames: true}},
		{name: "merge a null into a pointer to a pointer held by an interface", in: `null`, out: func() any {
			n := 1
			pn := &n
			var v any = &pn
			return &v
		}, opts: merge},
		{name: "merge a number into a pointer held by an interface", in: `5`, out: func() any {
			var v any = new(int)
			return &v
		}, opts: merge},
		{name: "merge an object into a struct held by an interface", in: `{"X":5}`, out: func() any {
			var v any = v1Pair{1, 2}
			return &v
		}, opts: merge},
		{name: "merge a null of the string option", in: `{"I":null,"F":null}`, out: func() any {
			return &v1DiffStringify{I: 1, F: 2}
		}, opts: merge},
		{name: "merge a quoted null of the string option", in: `{"I":"null","S":"null"}`, out: func() any {
			return &v1DiffStringify{I: 1, S: "x"}
		}, opts: unmarshalDiffOptions{v1Options: mergeWithLegacySemantics | stringifyWithLegacySemantics}},
		{name: "merge a long object", in: `{"X":"` + longerThanStreamBuffer + `",` + mergeValues[1:], out: func() any {
			v := newV1DiffMerge()
			return &struct {
				X string
				*v1DiffMerge
			}{v1DiffMerge: v.(*v1DiffMerge)}
		}, opts: merge},
		// StringifyWithLegacySemantics
		{name: "stringify bool", in: `{"B":"true","EB":"false"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify bool by default", in: `{"B":"true"}`, out: newStringify},
		{name: "stringify bool unquoted", in: `{"B":true}`, out: newStringify, opts: v1Stringify},
		{name: "stringify bool of null", in: `{"B":"null"}`, out: func() any { return &v1DiffStringify{B: true} }, opts: v1Stringify},
		{name: "stringify bool of another text", in: `{"B":"yes"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify string", in: `{"S":"\"x\\\"y\""}`, out: newStringify, opts: v1Stringify},
		{name: "stringify string of null", in: `{"S":"null"}`, out: func() any { return &v1DiffStringify{S: "x"} }, opts: v1Stringify},
		{name: "stringify string not of a JSON string", in: `{"S":"x"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify string unquoted", in: `{"S":1}`, out: newStringify, opts: v1Stringify},
		{name: "stringify numbers of the syntax of Go", in: `{"I":"+5","U":"007","F":"1e1"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify number of a hexadecimal", in: `{"I":"0x10"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify float of an infinity", in: `{"F":"-Inf"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify float of an underscore", in: `{"F":"1_0"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify number too big", in: `{"U":"300"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify numbers of null", in: `{"I":"null","F":"null"}`, out: func() any { return &v1DiffStringify{I: 1, F: 2} }, opts: v1Stringify},
		{name: "stringify pointers", in: `{"PB":"true","PS":"\"s\"","PI":"3"}`, out: newStringify, opts: v1Stringify},
		{name: "stringify pointers of null", in: `{"PB":"null","PS":"null","PI":"null"}`, out: func() any {
			return &v1DiffStringify{PB: new(bool), PS: new(string), PI: new(int)}
		}, opts: v1Stringify},
		{name: "stringify pointers of a JSON null", in: `{"PB":null,"PS":null,"PI":null}`, out: func() any {
			return &v1DiffStringify{PB: new(bool), PS: new(string), PI: new(int)}
		}, opts: v1Stringify},
		{name: "stringify pointer to a pointer", in: `{"PP":"1"}`, out: func() any { return new(v1DiffStringifyPtr) }, opts: v1Stringify},
		{name: "stringify pointer to a pointer of null", in: `{"PP":null}`, out: func() any { return new(v1DiffStringifyPtr) }, opts: v1Stringify},
		{name: "stringify pointer to a pointer of an array", in: `{"PP":[1,{"a":2}]}`, out: func() any { return new(v1DiffStringifyPtr) },
			opts: v1Stringify},
		{name: "stringify pointer to a pointer cut", in: `{"PP":"1`, out: func() any { return new(v1DiffStringifyPtr) }, opts: v1Stringify},
		{name: "string option of a slice cut", in: `{"X":"1`, out: func() any {
			return new(struct {
				X []int `json:",string"`
			})
		}},
		{name: "stringify bool and string of a JSON null", in: `{"B":null,"S":null}`, out: func() any { return &v1DiffStringify{B: true, S: "x"} },
			opts: v1Stringify},
		{name: "stringify quoted null into a pointer to a slice", in: `{"PSl":"null"}`, out: func() any { return new(v1DiffStringifyPointers) },
			opts: v1Stringify},
		{name: "stringify quoted null into a pointer to a struct", in: `{"PSt":"null"}`, out: func() any { return new(v1DiffStringifyPointers) },
			opts: v1Stringify},
		{name: "stringify quoted null into a pointer to a map", in: `{"PM":"null"}`, out: func() any { return new(v1DiffStringifyPointers) },
			opts: v1Stringify},
		{name: "stringify quoted null into a pointer of a method", in: `{"PT":"null"}`, out: func() any { return new(v1DiffStringifyPointers) },
			opts: v1Stringify},
		{name: "stringify quoted null into a pointer which is set", in: `{"P":"null"}`, out: func() any {
			s := "keep"
			return &v1DiffStringifyPointers{P: &s, A: &s}
		}, opts: v1Stringify},
		{name: "stringify with StringifyNumbers", in: `["+5","null","1"]`, out: outOf([]int(nil)), opts: unmarshalDiffOptions{v1Options: stringifyWithLegacySemantics, stringifyNumbers: true}},
		{name: "stringify map keys", in: `{"null":1,"+2":2}`, out: func() any { return new(map[int]int) }, opts: v1Stringify},
		{name: "stringify map keys by default", in: `{"+2":2}`, out: func() any { return new(map[int]int) }},
		// MatchCaseSensitiveDelimiter
		{name: "delimiter name of another case", in: `{"FOO_BAR":1}`, out: newDelimiter, opts: both},
		{name: "delimiter name without the underscore", in: `{"foobar":1}`, out: newDelimiter, opts: both},
		{name: "delimiter name without the underscore insensitive", in: `{"foobar":1}`, out: newDelimiter, opts: insensitive},
		{name: "delimiter name of case:ignore", in: `{"baz":2}`, out: newDelimiter, opts: delimiter},
		{name: "delimiter name of case:ignore without the dash", in: `{"quxx":3,"Qux-X":4}`, out: newDelimiter, opts: delimiter},
		{name: "delimiter name of case:ignore without the dash by default", in: `{"quxx":3}`, out: newDelimiter},
		{name: "delimiter name of case:strict", in: `{"STRICT_S":4}`, out: newDelimiter, opts: both},
		{name: "delimiter names of the same fold", in: `{"A_B":1}`, out: newDelimiterTwo, opts: both},
		{name: "delimiter names of the same fold insensitive", in: `{"A_B":1}`, out: newDelimiterTwo, opts: insensitive},
		{name: "delimiter names of the same fold without the underscore", in: `{"AB":1}`, out: newDelimiterTwo, opts: both},
		{name: "delimiter names of case:ignore of the same fold", in: `{"A_B":1}`,
			out: func() any { return new(v1DiffDelimiterIgnore) }, opts: delimiter},
		{name: "delimiter names of case:ignore of the same fold by default", in: `{"A_B":1}`,
			out: func() any { return new(v1DiffDelimiterIgnore) }},
		{name: "delimiter names of a member twice", in: `{"foo_bar":1,"FOO_BAR":2}`, out: newDelimiter, opts: both},
		{name: "delimiter names in a long object", in: `{"x":"` + longerThanStreamBuffer + `","FOO_BAR":1,"foobar":2,"Baz":3}`,
			out: newDelimiter, opts: both},
		// FormatByteArrayAsArray
		{name: "byte array as array", in: `[1,2,3]`, out: outOf([3]byte{}), opts: arrayOfBytes},
		{name: "byte array as array of base64", in: `"AQID"`, out: outOf([3]byte{}), opts: arrayOfBytes},
		{name: "byte array as array by default", in: `[1,2,3]`, out: outOf([3]byte{})},
		{name: "byte array as array stringified", in: `["1","2"]`, out: outOf([2]byte{}),
			opts: unmarshalDiffOptions{v1Options: formatByteArrayAsArray, stringifyNumbers: true}},
		{name: "byte array as array of a number too big", in: `[1,256]`, out: outOf([2]byte{}), opts: arrayOfBytes},
		{name: "byte slice with FormatByteArrayAsArray", in: `"AQID"`, out: outOf([]byte(nil)), opts: arrayOfBytes},
		{name: "byte array key as array", in: `{"AQI=":1}`, out: func() any { return new(map[[2]byte]int) }, opts: arrayOfBytes},
		{name: "named byte array as array", in: `[1,2]`, out: outOf([2]v1Byte{}), opts: bothBytes},
		// FormatBytesWithLegacySemantics
		{name: "bytes of a JSON array", in: `[1,2,3]`, out: outOf([]byte(nil)), opts: v1Bytes},
		{name: "bytes of a JSON array by default", in: `[1,2,3]`, out: outOf([]byte(nil))},
		{name: "byte array of a JSON array", in: `[1,2]`, out: outOf([2]byte{}), opts: v1Bytes},
		{name: "bytes of a JSON array of an error", in: `[1,"x",3]`, out: outOf([]byte(nil)), opts: v1Bytes},
		{name: "bytes of a long JSON array", in: "[" + strings.Repeat("1,", streamBufferSize/len("1,")) + "1]", out: outOf([]byte(nil)), opts: v1Bytes},
		{name: "named bytes of base64", in: `"AQID"`, out: outOf([]v1Byte(nil)), opts: v1Bytes},
		{name: "named bytes of base64 by default", in: `"AQID"`, out: outOf([]v1Byte(nil))},
		{name: "named byte array of base64", in: `"AQI="`, out: outOf([2]v1Byte{}), opts: v1Bytes},
		{name: "named bytes of null", in: `null`, out: func() any { return &[]v1Byte{1} }, opts: v1Bytes},
		{name: "named bytes of an object", in: `{"a":1}`, out: outOf([]v1Byte(nil)), opts: v1Bytes},
		{name: "named bytes of an object of a syntax error", in: `{"a":1,}`, out: outOf([]v1Byte(nil)), opts: v1Bytes},
		{name: "named bytes of a bool", in: `true`, out: outOf([]v1Byte(nil)), opts: v1Bytes},
		{name: "text bytes of base64", in: `"AQI="`, out: outOf([]v1TextByte(nil)), opts: v1Bytes},
		{name: "text bytes of a JSON array", in: `["bb","b"]`, out: outOf([]v1TextByte(nil)), opts: v1Bytes},
		{name: "text bytes of a JSON array by default", in: `["bb","b"]`, out: outOf([]v1TextByte(nil))},
		{name: "byte array of a longer JSON array", in: `[1,2,3]`, out: outOf([2]byte{}), opts: v1Bytes},
		{name: "byte array of a shorter JSON array", in: `[1]`, out: outOf([2]byte{}), opts: v1Bytes},
		{name: "named byte array of a longer JSON array", in: `[1,2,3]`, out: outOf([2]v1Byte{}), opts: v1Bytes},
		{name: "byte array of a field of a longer JSON array", in: `{"A":[1,2,3]}`,
			out: func() any { return new(v1DiffBytes) }, opts: v1Bytes},
		{name: "byte arrays of JSON arrays of other lengths", in: `[[1,2,3],[4,5]]`, out: outOf([][2]byte(nil)),
			opts: v1Bytes},
		{name: "byte arrays as arrays of other lengths", in: `[[1,2,3],[4,5]]`, out: outOf([][2]byte(nil)), opts: arrayOfBytes},
		{name: "text-in bytes of base64", in: `"AQI="`, out: outOf([]v1TextInByte(nil)), opts: v1Bytes},
		{name: "text-in bytes of a JSON array", in: `["x","yy"]`, out: outOf([]v1TextInByte(nil)), opts: v1Bytes},
		{name: "JSON-in bytes of base64", in: `"AQI="`, out: outOf([]v1JSONInByte(nil)), opts: v1Bytes},
		{name: "JSON-in bytes of a JSON array", in: `[1,"x"]`, out: outOf([]v1JSONInByte(nil)), opts: v1Bytes},
		{name: "named slice of base64", in: `"AQI="`, out: outOf(v1NamedBytes(nil)), opts: v1Bytes},
		{name: "pointer to named bytes", in: `"AQI="`, out: outOf((*[]v1Byte)(nil)), opts: v1Bytes},
		{name: "pointer to a byte array as an array", in: `[1,2]`, out: outOf((*[2]byte)(nil)), opts: arrayOfBytes},
		{name: "named byte array key", in: `{"AQI=":1}`, out: func() any { return new(map[[2]v1Byte]int) }, opts: v1Bytes},
		{name: "bytes struct", in: bytesStruct, out: func() any { return new(v1DiffBytes) }, opts: v1Bytes},
		{name: "bytes struct with arrays", in: bytesStruct, out: func() any { return new(v1DiffBytes) }, opts: bothBytes},
		// FormatDurationAsNano
		{name: "duration nano", in: `1500`, out: outOf(time.Duration(0)), opts: nano},
		{name: "duration nano by default", in: `1500`, out: outOf(time.Duration(0))},
		{name: "duration nano negative", in: `-3`, out: outOf(time.Duration(0)), opts: nano},
		{name: "duration nano fraction", in: `1.5`, out: outOf(time.Duration(0)), opts: nano},
		{name: "duration nano string", in: `"1500"`, out: outOf(time.Duration(0)), opts: nano},
		{name: "duration nano stringified", in: `"1500"`, out: outOf(time.Duration(0)), opts: nanoStringify},
		{name: "duration nano null", in: `null`, out: outOf(time.Duration(7)), opts: nano},
		{name: "duration nano overflow", in: `1e30`, out: outOf(time.Duration(0)), opts: nano},
		{name: "duration nano key", in: `{"5":1,"-2":2}`, out: func() any { return new(map[time.Duration]int) }, opts: nano},
		{name: "duration nano key of a fraction", in: `{"1.5":1}`, out: func() any { return new(map[time.Duration]int) }, opts: nano},
		{name: "duration nano struct", in: `{"D":1,"P":2,"S":"3","SP":"4","E":5,"Z":6,"Keys":{"9":"x"},"Any":7}`,
			out: func() any { return new(v1DiffDuration) }, opts: nano},
		{name: "duration nano string option of a number", in: `{"S":3}`, out: func() any { return new(v1DiffDuration) }, opts: nano},
		{name: "duration nano slice", in: `[1,2]`, out: outOf([]time.Duration(nil)), opts: nano},
		// UnmarshalArrayFromAnyLength
		{name: "array longer than the Go array", in: `[1,2,3]`, out: outOf([2]int{}), opts: anyLength},
		{name: "array shorter than the Go array", in: `[1]`, out: outOf([3]int{7, 7, 7}), opts: anyLength},
		{name: "array empty", in: `[]`, out: outOf([2]int{7, 7}), opts: anyLength},
		{name: "array longer with invalid UTF-8 past the length", in: "[1,2,\"\xff\"]", out: outOf([2]int{}), opts: anyLength},
		{name: "array longer with a name twice past the length", in: `[1,2,{"a":1,"a":2}]`, out: outOf([2]int{}), opts: anyLength},
		{name: "array longer with a name twice past the length without the option", in: `[1,2,{"a":1,"a":2}]`,
			out: outOf([2]int{})},
		{name: "array longer with a syntax error past the length without the option", in: `[1,2,{"a"}]`, out: outOf([2]int{})},
		{name: "array longer with invalid UTF-8 in a long string past the length without the option",
			in: `[1,2,"` + longerThanStreamBuffer + "\xff\"]", out: outOf([2]int{})},
		{name: "array longer with a name twice past the length without the option of the text", in: `[1,2,{"a":1,"a":2}]`,
			out: outOf([2]int{}), opts: unmarshalDiffOptions{allowInvalidUTF8: true}},
		{name: "array longer with invalid UTF-8 past the length without the option of the text", in: "[1,2,\"\xff\"]",
			out: outOf([2]int{}), opts: unmarshalDiffOptions{allowDuplicateNames: true}},
		{name: "array longer with invalid UTF-8 past the length allowed", in: "[1,2,\"\xff\"]", out: outOf([2]int{}),
			opts: unmarshalDiffOptions{v1Options: unmarshalArrayFromAnyLength, allowInvalidUTF8: true}},
		{name: "array longer with a name twice past the length allowed", in: `[1,2,{"a":1,"a":2}]`, out: outOf([2]int{}),
			opts: unmarshalDiffOptions{v1Options: unmarshalArrayFromAnyLength, allowDuplicateNames: true}},
		{name: "array longer with a long string past the length", in: `[1,2,"` + longerThanStreamBuffer + `"]`, out: outOf([2]int{}), opts: anyLength},
		{name: "array longer with invalid UTF-8 in a long string past the length", in: `[1,2,"` + longerThanStreamBuffer + "\xff\"]",
			out: outOf([2]int{}), opts: anyLength},
		{name: "array longer with invalid UTF-8 in a long string past the length allowed", in: `[1,2,"` + longerThanStreamBuffer + "\xff\"]",
			out: outOf([2]int{}), opts: unmarshalDiffOptions{v1Options: unmarshalArrayFromAnyLength, allowInvalidUTF8: true}},
		{name: "array longer with a syntax error past the length", in: `[1,2,{"a"}]`, out: outOf([2]int{}), opts: anyLength},
		{name: "array longer of an error in the array", in: `[1,"x",3]`, out: outOf([2]int{}), opts: anyLength},
		{name: "array of arrays longer", in: `[[1,2,3],[4]]`, out: outOf([2][2]int{}), opts: anyLength},
		{name: "bytes array longer than the Go array", in: `"AQID"`, out: outOf([2]byte{}), opts: anyLength},
		{name: "bytes array shorter than the Go array", in: `"AQID"`, out: outOf([4]byte{7, 7, 7, 7}), opts: anyLength},
		{name: "array in a struct longer", in: `{"A":[1,2,3]}`, out: func() any { return new(struct{ A [2]int }) }, opts: anyLength},
		// ParseBytesWithLooseRFC4648
		{name: "bytes of base64 with a line feed", in: `"AQ\r\nID"`, out: outOf([]byte(nil)), opts: loose},
		{name: "bytes of base64 with a line feed strict", in: `"AQ\r\nID"`, out: outOf([]byte(nil))},
		{name: "bytes array of base64 with a line feed", in: `"AQ\r\nID"`, out: outOf([3]byte{}), opts: loose},
		{name: "bytes of base64 of a wrong character loose", in: `"AQ*ID"`, out: outOf([]byte(nil)), opts: loose},
		// ParseTimeWithLooseRFC3339
		{name: "time of an hour of one digit", in: `"2006-01-02T1:04:05Z"`, out: outOf(time.Time{}), opts: rfc},
		{name: "time of an hour of one digit strict", in: `"2006-01-02T1:04:05Z"`, out: outOf(time.Time{})},
		{name: "time of a comma before the fraction", in: `"2006-01-02T15:04:05,5Z"`, out: outOf(time.Time{}), opts: rfc},
		{name: "time of a zone out of range", in: `"2006-01-02T15:04:05+24:00"`, out: outOf(time.Time{}), opts: rfc},
		{name: "time key of an hour of one digit", in: `{"2006-01-02T1:04:05Z":1}`, out: func() any { return new(map[time.Time]int) }, opts: rfc},
		{name: "time of a wrong text loose", in: `"2006-01-02"`, out: outOf(time.Time{}), opts: rfc},
	}
}

// A value of the `string` option of a type which has no representation is the error of the type, as encoding/json/v2
// reports it, with StringifyWithLegacySemantics or not, through a pointer, and its null is null.
func TestStringOptionOfNoRepresentationMatchesStd(t *testing.T) {
	type C struct {
		F chan int `json:",string"`
	}
	type P struct {
		F *chan int `json:",string"`
	}
	type X struct {
		F complex64 `json:",string"`
	}
	ch := make(chan int)
	for _, set := range []v1OptionSet{0, stringifyWithLegacySemantics} {
		for _, v := range []any{C{}, P{}, P{F: &ch}, X{}} {
			got, err := json.Marshal(v, set.ours()...)
			want, stdErr := stdjson.Marshal(v, set.std()...)
			// the output of a call which succeeds: encoding/json/v2 returns what it wrote before an error
			if g, w := describeUnmarshalError(err), describeUnmarshalError(stdErr); g != w || err == nil && string(got) != string(want) {
				t.Errorf("Marshal %T, options %v:\n got %s %s\nwant %s %s", v, set, got, g, want, w)
			}
		}
		for _, in := range []string{`{"F":"1"}`, `{"F":1}`, `{"F":null}`} {
			for _, newOut := range []func() any{outOf(C{}), outOf(P{}), outOf(X{})} {
				ours, theirs := newOut(), newOut()
				err := json.Unmarshal([]byte(in), ours, set.ours()...)
				stdErr := stdjson.Unmarshal([]byte(in), theirs, set.std()...)
				if g, w := describeUnmarshalError(err), describeUnmarshalError(stdErr); g != w || !reflect.DeepEqual(ours, theirs) {
					t.Errorf("Unmarshal %s into %T, options %v:\n got %s %+v\nwant %s %+v", in, ours, set, g, valueOf(ours), w,
						valueOf(theirs))
				}
			}
		}
	}
}

// describeV1Error describes err, an error of encoding/json of the root package or of the standard library, by its
// type and its fields, to compare the errors of the two: the cause of a MarshalerError is described too, and the
// modal verb of a text of encoding/json/v2 is "cannot" ( see stableErrorText ).
func describeV1Error(err error) string {
	switch e := err.(type) {
	case nil:
		return "<nil>"
	case *gojson.UnsupportedTypeError:
		return "UnsupportedTypeError " + e.Type.String()
	case *stdjsonv1.UnsupportedTypeError:
		return "UnsupportedTypeError " + e.Type.String()
	case *gojson.UnsupportedValueError:
		return "UnsupportedValueError " + e.Str + " " + describeV1Value(e.Value)
	case *stdjsonv1.UnsupportedValueError:
		return "UnsupportedValueError " + e.Str + " " + describeV1Value(e.Value)
	case *gojson.MarshalerError:
		return "MarshalerError " + e.Type.String() + " " + stableErrorText(e.Error()) + " of " + describeV1Error(e.Err)
	case *stdjsonv1.MarshalerError:
		return "MarshalerError " + e.Type.String() + " " + stableErrorText(e.Error()) + " of " + describeV1Error(e.Err)
	case *gojson.UnmarshalTypeError:
		return fmt.Sprintf("UnmarshalTypeError %q %v %d %q %q %s of %s", e.Value, e.Type, e.Offset, e.Struct, e.Field,
			stableErrorText(e.Error()), describeV1Error(e.Err))
	case *stdjsonv1.UnmarshalTypeError:
		return fmt.Sprintf("UnmarshalTypeError %q %v %d %q %q %s of %s", e.Value, e.Type, e.Offset, e.Struct, e.Field,
			stableErrorText(e.Error()), describeV1Error(e.Err))
	case *gojson.InvalidUnmarshalError:
		return "InvalidUnmarshalError " + e.Error()
	case *stdjsonv1.InvalidUnmarshalError:
		return "InvalidUnmarshalError " + e.Error()
	case *gojson.SyntaxError:
		return fmt.Sprintf("SyntaxError %d %s", e.Offset, e.Error())
	case *stdjsonv1.SyntaxError:
		return fmt.Sprintf("SyntaxError %d %s", e.Offset, e.Error())
	case *json.SemanticError, *stdjson.SemanticError:
		// its fields, and its text, which tells the action of the call too
		return describeUnmarshalError(err) + " " + stableErrorText(err.Error())
	}
	return fmt.Sprintf("%T %v", err, err)
}

// describeV1Value describes the Value of an UnsupportedValueError.
func describeV1Value(v reflect.Value) string {
	if !v.IsValid() {
		return "no value"
	}
	return fmt.Sprintf("%v %v", v.Type(), v.Interface())
}

// namedAsOurs is a text of the types for encoding/json of the tests of ReportErrorsWithLegacySemantics, named as the
// types for the root package which they pair with: reportStdX is reportX.
func namedAsOurs(s string) string {
	return strings.ReplaceAll(s, "reportStd", "report")
}

// reportErrJSON and reportErrAppend are types whose methods of marshaling fail, of value receivers, as v1Text.
type reportErrJSON struct{ X int }

func (reportErrJSON) MarshalJSON() ([]byte, error) { return nil, errors.New("json failed") }

type reportErrAppend struct{ X int }

func (reportErrAppend) AppendText([]byte) ([]byte, error) { return nil, errors.New("append failed") }

// reportBadJSON is a type whose MarshalJSON writes no valid JSON.
type reportBadJSON struct{ X int }

func (reportBadJSON) MarshalJSON() ([]byte, error) { return []byte(`{x`), nil }

// reportWrapJSON is a type whose MarshalJSON fails by marshaling a reportPtrErrJSON: its error is a MarshalerError,
// which encoding/json wraps in another one.
type reportWrapJSON struct{ X int }

func (reportWrapJSON) MarshalJSON() ([]byte, error) { return gojson.Marshal(&reportPtrErrJSON{}) }

// reportStdWrapJSON is reportWrapJSON for encoding/json.
type reportStdWrapJSON struct{ X int }

func (reportStdWrapJSON) MarshalJSON() ([]byte, error) { return stdjsonv1.Marshal(&reportPtrErrJSON{}) }

// reportPtrErrJSON is a type whose MarshalJSON fails, of a pointer receiver.
type reportPtrErrJSON struct{ X int }

func (*reportPtrErrJSON) MarshalJSON() ([]byte, error) { return nil, errors.New("pointer json failed") }

// reportErrTo is a type whose MarshalJSONTo fails, and reportStdErrTo the one for encoding/json.
type reportErrTo struct{ X int }

func (reportErrTo) MarshalJSONTo(*jsontext.Encoder) error { return errors.New("to failed") }

type reportStdErrTo struct{ X int }

func (reportStdErrTo) MarshalJSONTo(*stdjsontext.Encoder) error { return errors.New("to failed") }

// reportNestedTo is a type whose MarshalJSONTo marshals a channel by MarshalEncode, and reportStdNestedTo the one for
// encoding/json: the error of the inner call, of encoding/json, is wrapped by the one of the method.
type reportNestedTo struct{ X int }

func (reportNestedTo) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, make(chan int))
}

type reportStdNestedTo struct{ X int }

func (reportStdNestedTo) MarshalJSONTo(enc *stdjsontext.Encoder) error {
	return stdjson.MarshalEncode(enc, make(chan int))
}

// reportNode is a struct of a pointer to itself, whose cycle is an error.
type reportNode struct{ Next *reportNode }

// reportFloat is a named float.
type reportFloat float64

// reportStringTags has fields of the `string` option of types which it doesn't apply to.
type reportStringTags struct {
	B  bool            `json:",string"`
	S  string          `json:",string"`
	Y  []byte          `json:",string"`
	M  map[string]int  `json:",string"`
	L  []int           `json:",string"`
	A  [1]int          `json:",string"`
	I  any             `json:",string"`
	J  any             `json:",string"`
	PP **int           `json:",string"`
	PB *bool           `json:",string"`
	T  time.Time       `json:",string"`
	St struct{ X int } `json:",string"`
}

// reportPtrNumJSON is a struct whose MarshalJSON is of its pointer, which a value held by an interface value doesn't
// have.
type reportPtrNumJSON struct{ X int }

func (*reportPtrNumJSON) MarshalJSON() ([]byte, error) { return []byte(`{"p":1}`), nil }

// reportToNested is a type whose MarshalJSONTo writes values in an array, and reportStdToNested the one for
// encoding/json; reportToNestedField and reportStdToNestedField have a field of the `string` option of them.
type reportToNested struct{}

func (reportToNested) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, []any{1, map[string]int{"a": 2}, struct{ Y int }{3}})
}

type reportStdToNested struct{}

func (reportStdToNested) MarshalJSONTo(enc *stdjsontext.Encoder) error {
	return stdjson.MarshalEncode(enc, []any{1, map[string]int{"a": 2}, struct{ Y int }{3}})
}

type reportToNestedField struct {
	F reportToNested `json:",string"`
	I any            `json:",string"`
}

type reportStdToNestedField struct {
	F reportStdToNested `json:",string"`
	I any               `json:",string"`
}

// reportStructErrors are structs whose definitions are errors of encoding/json/v2, which encoding/json ignores.
type (
	reportMalformed struct {
		A int `json:"a,omitempty,"`
		B int
	}
	reportUnexportedTag struct {
		a int `json:"a"` //nolint:unused // a field of a tag, which encoding/json/v2 rejects
		B int
	}
)

// v1ReportValues are the values of the errors of marshaling with ReportErrorsWithLegacySemantics.
func v1ReportValues() []any {
	n := &reportNode{}
	n.Next = n
	m := map[string]any{}
	m["a"] = m
	sl := []any{nil}
	sl[0] = sl
	pp := new(*int)
	var heldSlice, heldStruct any = []int{1}, struct{ X int }{4}
	// a year of five digits, which RFC 3339 has no text of
	outOfRFC3339 := time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)
	return []any{
		struct{ C chan int }{}, []any{make(chan int)}, complex(1, 2), map[string]func(){"a": nil},
		math.NaN(), math.Inf(1), struct{ F float32 }{float32(math.Inf(-1))}, reportFloat(math.NaN()),
		map[float64]int{math.Inf(1): 1}, map[any]int{math.Inf(-1): 1}, n, m, sl, time.Duration(5), struct{ D time.Duration }{3},
		reportConflict{}, reportMalformed{A: 1, B: 2}, reportUnexportedTag{B: 1},
		reportStringTags{I: 1, J: true, PP: pp, PB: new(bool)}, reportStringTags{I: []int{1}, J: map[string]int{"a": 1}},
		reportStringTags{I: struct{ X int }{1}, J: &[]int{2}}, reportStringTags{I: new(int), J: reportFloat(1.5)},
		reportStringTags{I: &heldSlice, J: &heldStruct}, reportStringTags{I: reportPtrNumJSON{1}, J: &reportPtrNumJSON{2}},
		reportStringTags{I: map[string]int{"a": 1}, J: &map[string]any{"b": []int{2}}},
		map[time.Time]int{outOfRFC3339: 1}, map[time.Time]int{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC): 1},
		reportErrJSON{}, &reportErrJSON{}, v1Text(0), reportErrAppend{}, reportBadJSON{}, reportPtrErrJSON{},
		&reportPtrErrJSON{}, []reportPtrErrJSON{{}}, map[v1Text]int{0: 1}, map[string]v1Text{"a": 0},
		outOfRFC3339,
		"\xff", map[string]int{"\xff": 1}, map[[2]int]int{{1, 2}: 1}, map[struct{ A int }]int{{1}: 1},
	}
}

// With ReportErrorsWithLegacySemantics, every function of marshaling reports the errors of encoding/json, as
// encoding/json/v2 reports them with the option of encoding/json, and writes what it writes: encoding/json/v2 reports
// the same errors by each of its functions.
func TestMarshalReportErrorsMatchesStd(t *testing.T) {
	type pair struct {
		name         string
		ours, theirs any
	}
	var pairs []pair
	for i, v := range v1ReportValues() {
		pairs = append(pairs, pair{fmt.Sprintf("%d:%T", i, v), v, v})
	}
	pairs = append(pairs, pair{"wrapped", reportWrapJSON{}, reportStdWrapJSON{}}, pair{"to", reportErrTo{}, reportStdErrTo{}},
		pair{"nested", reportNestedTo{}, reportStdNestedTo{}}, pair{"nested in a slice", []reportNestedTo{{}}, []reportStdNestedTo{{}}},
		pair{"nested of the string option", reportToNestedField{I: reportToNested{}}, reportStdToNestedField{I: reportStdToNested{}}})
	for _, set := range []v1OptionSet{reportErrorsWithLegacySemantics, reportErrorsWithLegacySemantics | stringifyWithLegacySemantics, reportErrorsWithLegacySemantics | callMethodsWithLegacySemantics, allV1Options} {
		for _, p := range pairs {
			want, wantErr := stdjson.Marshal(p.theirs, set.std()...)
			w := namedAsOurs(describeV1Error(wantErr))
			for _, m := range marshalFuncs {
				got, err := m.marshal(p.ours, set.ours()...)
				if g := describeV1Error(err); g != w || err == nil && got != string(want) {
					t.Errorf("%s, options %v: %s:\n got %s %s\nwant %s %s", p.name, set, m.name, got, g, want, w)
				}
			}
		}
	}
}

// With ReportErrorsWithLegacySemantics, the errors of the functions of marshaling are MarshalerErrors of the pointers to
// the values, of the names of the functions.
func TestMarshalReportErrorsOfFuncsMatchesStd(t *testing.T) {
	// the calls of the functions which decline, which are called once for a value
	var calls, stdCalls int
	for _, c := range []struct {
		name    string
		v       any
		opts    json.Options
		stdOpts stdjson.Options
	}{
		{"func", reportErrJSON{X: 1}, json.WithMarshalers(json.MarshalFunc(func(reportErrJSON) ([]byte, error) { return nil, errors.New("f") })),
			stdjson.WithMarshalers(stdjson.MarshalFunc(func(reportErrJSON) ([]byte, error) { return nil, errors.New("f") }))},
		{"func of bad JSON", 1, json.WithMarshalers(json.MarshalFunc(func(int) ([]byte, error) { return []byte(`[`), nil })),
			stdjson.WithMarshalers(stdjson.MarshalFunc(func(int) ([]byte, error) { return []byte(`[`), nil }))},
		{"func of an interface", []any{reportErrJSON{}}, json.WithMarshalers(json.MarshalFunc(func(stdjsonv1.Marshaler) ([]byte, error) { return nil, errors.New("i") })),
			stdjson.WithMarshalers(stdjson.MarshalFunc(func(stdjsonv1.Marshaler) ([]byte, error) { return nil, errors.New("i") }))},
		{"to func", 1, json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, int) error { return errors.New("t") })),
			stdjson.WithMarshalers(stdjson.MarshalToFunc(func(*stdjsontext.Encoder, int) error { return errors.New("t") }))},
		{"to func of an interface", []any{reportErrJSON{}}, json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, stdjsonv1.Marshaler) error { return errors.New("ti") })),
			stdjson.WithMarshalers(stdjson.MarshalToFunc(func(*stdjsontext.Encoder, stdjsonv1.Marshaler) error { return errors.New("ti") }))},
		{"declining to func of the string option", reportStringTags{I: []int{1}, J: struct{ X int }{2}},
			json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, []int) error { calls++; return errors.ErrUnsupported })),
			stdjson.WithMarshalers(stdjson.MarshalToFunc(func(*stdjsontext.Encoder, []int) error { stdCalls++; return errors.ErrUnsupported }))},
		{"declining to func of any of the string option", reportStringTags{I: []int{1}, J: struct{ X int }{2}},
			json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, any) error { return errors.ErrUnsupported })),
			stdjson.WithMarshalers(stdjson.MarshalToFunc(func(*stdjsontext.Encoder, any) error { return errors.ErrUnsupported }))},
		{"to func writing values of the string option", reportStringTags{I: []int{1}},
			json.WithMarshalers(json.MarshalToFunc(func(enc *jsontext.Encoder, _ []int) error { return json.MarshalEncode(enc, []any{[]uint{1}, 2}) })),
			stdjson.WithMarshalers(stdjson.MarshalToFunc(func(enc *stdjsontext.Encoder, _ []int) error { return stdjson.MarshalEncode(enc, []any{[]uint{1}, 2}) }))},
	} {
		got, err := json.Marshal(c.v, append(reportErrorsWithLegacySemantics.ours(), c.opts)...)
		want, stdErr := stdjson.Marshal(c.v, append(reportErrorsWithLegacySemantics.std(), c.stdOpts)...)
		if g, w := describeV1Error(err), describeV1Error(stdErr); g != w || err == nil && string(got) != string(want) {
			t.Errorf("%s:\n got %s %s\nwant %s %s", c.name, got, g, want, w)
		}
	}
	if calls != stdCalls {
		t.Errorf("calls of the declining function: got %d, want %d", calls, stdCalls)
	}
}

// reportUnmarshalJSON is a type whose UnmarshalJSON fails for a JSON number, and reportUnmarshalText one whose
// UnmarshalText fails for "bad".
type reportUnmarshalJSON int

func (v *reportUnmarshalJSON) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		*v = 5
		return errors.New("json method failed")
	}
	return nil
}

type reportUnmarshalText string

func (v *reportUnmarshalText) UnmarshalText(b []byte) error {
	if string(b) == "bad" {
		return errors.New("text method failed")
	}
	*v = reportUnmarshalText(b)
	return nil
}

// reportUnmarshalFrom is a type whose UnmarshalJSONFrom fails after it reads its value, and reportStdUnmarshalFrom the
// one for encoding/json.
type reportUnmarshalFrom int

func (v *reportUnmarshalFrom) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if _, err := dec.ReadValue(); err != nil {
		return err
	}
	return errors.New("from method failed")
}

type reportStdUnmarshalFrom int

func (v *reportStdUnmarshalFrom) UnmarshalJSONFrom(dec *stdjsontext.Decoder) error {
	if _, err := dec.ReadValue(); err != nil {
		return err
	}
	return errors.New("from method failed")
}

// reportUnexported is an unexported struct, which an embedded nil pointer to can't be set.
type reportUnexported struct{ X int }

// reportNilField embeds a pointer to an unexported struct.
type reportNilField struct {
	*reportUnexported
	Y int
}

// reportCase is an input and the value it is decoded into with ReportErrorsWithLegacySemantics, made by ours and
// theirs, which differ for the types of methods of jsontext only, with the options of encoding/json and others.
type reportCase struct {
	name         string
	in           string
	ours, theirs func() any
	opts         v1OptionSet
	extra        []json.Options
	stdExtra     []stdjson.Options
}

// reportCases are the cases of the unmarshaling with ReportErrorsWithLegacySemantics.
func reportCases() []reportCase {
	type R1 struct{ A, B int }
	type R2 struct {
		A, B int
		C    []int
	}
	type nested struct {
		A struct{ B int }
		C int
	}
	type rejected struct{ A int }
	type ambiguous struct{ Aa, AA int }
	type withReader struct{ R io.Reader }
	type withChan struct {
		C chan int
		N int
	}
	type withDuration struct{ D time.Duration }
	type withArray struct{ A [2]int }
	type withBytes struct{ B []byte }
	type withMethods struct {
		J reportUnmarshalJSON
		T reportUnmarshalText
		N int
	}
	type withArrayAfter struct {
		A [2]int
		B int
	}
	type withPointers struct {
		P *int
		Q **int
		R []*[2]int
		N int
	}
	type withEmbeddedMap struct {
		A int
		X map[string]int `json:",embed"`
		C int
	}
	type stringTags struct {
		B bool           `json:",string"`
		M map[string]int `json:",string"`
		T time.Time      `json:",string"`
		I int            `json:",string"`
	}
	// structs of no name, whose errors name no struct
	type (
		withInt8 = struct {
			A int8
			B int
		}
		withAny = struct {
			A int
			B any
		}
		withKeys = struct {
			M map[int]int
			N int
		}
		anyAfter = struct {
			X int
			A any
		}
	)
	setPointer := func() any {
		six := 6
		return &withPointers{P: &six}
	}
	inputs := []struct {
		name, in     string
		ours, theirs func() any
	}{
		{"kinds", `{"A":"x","B":"y","C":[1,"z",3]}`, outOf(R2{}), outOf(R2{})},
		{"int8 range", `{"A":300,"B":2}`, outOf(withInt8{}), outOf(withInt8{})},
		{"slice", `[1,"x",3]`, outOf([]int(nil)), outOf([]int(nil))},
		{"map values", `{"a":"x","b":2}`, outOf(map[string]int(nil)), outOf(map[string]int(nil))},
		{"map keys", `{"x":1,"2":2}`, outOf(map[int]int(nil)), outOf(map[int]int(nil))},
		{"fraction", `{"A":1.5,"B":2}`, outOf(R1{}), outOf(R1{})},
		{"negative uint", `-1`, outOf(uint(0)), outOf(uint(0))},
		{"float32 range", `1e100`, outOf(float32(0)), outOf(float32(0))},
		{"nested", `{"A":{"B":"x"},"C":2}`, outOf(nested{}), outOf(nested{})},
		{"nested object", `{"A":{"B":{"q":1}},"C":2}`, outOf(nested{}), outOf(nested{})},
		{"cut", `{"A":1,`, outOf(R1{}), outOf(R1{})},
		{"comma before the end", `{"A":1,}`, outOf(R1{}), outOf(R1{})},
		{"comma before the end of an array", `{"A":1,"C":[1,]}`, outOf(R2{}), outOf(R2{})},
		{"commas twice", `{"A":1,,"B":2}`, outOf(R1{}), outOf(R1{})},
		{"missing value", `{"A":1,"B":}`, outOf(R1{}), outOf(R1{})},
		{"after", `{"A":1} x`, outOf(R1{}), outOf(R1{})},
		{"empty", ``, outOf(R1{}), outOf(R1{})},
		{"literal", `{"A":tru}`, outOf(R1{}), outOf(R1{})},
		{"invalid character", `{"A":1,"B":x}`, outOf(R1{}), outOf(R1{})},
		{"duplicate", `{"A":1,"A":2}`, outOf(R1{}), outOf(R1{})},
		{"invalid UTF-8", "\"\xff\"", outOf(""), outOf("")},
		{"not a pointer", `1`, func() any { return 1 }, func() any { return 1 }},
		{"nil pointer", `1`, func() any { return (*int)(nil) }, func() any { return (*int)(nil) }},
		{"nil", `{`, func() any { return nil }, func() any { return nil }},
		{"ambiguous", `{"aA":1}`, outOf(ambiguous{}), outOf(ambiguous{})},
		{"nil field", `{"X":1,"Y":2}`, outOf(reportNilField{}), outOf(reportNilField{})},
		{"nil interface", `{"R":1}`, outOf(withReader{}), outOf(withReader{})},
		{"methods", `{"J":1,"T":"bad","N":3}`, outOf(withMethods{}), outOf(withMethods{})},
		{"text keys", `{"ok":1,"bad":2,"z":3}`, outOf(map[reportUnmarshalText]int(nil)), outOf(map[reportUnmarshalText]int(nil))},
		{"conflict", `{"x":1,"C":2}`, outOf(reportConflict{}), outOf(reportConflict{})},
		{"malformed", `{"a":1,"B":2}`, outOf(reportMalformed{}), outOf(reportMalformed{})},
		{"string tags", `{"B":true,"M":{"a":1},"T":"2020-01-02T03:04:05Z","I":"7"}`, outOf(stringTags{}), outOf(stringTags{})},
		{"chan", `{"C":1,"N":2}`, outOf(withChan{}), outOf(withChan{})},
		{"chan null", `{"C":null,"N":2}`, outOf(withChan{}), outOf(withChan{})},
		{"duration", `{"D":1}`, outOf(withDuration{}), outOf(withDuration{})},
		{"array length", `{"A":[1,2,3]}`, outOf(withArray{}), outOf(withArray{})},
		{"base64", `{"B":"!!"}`, outOf(withBytes{}), outOf(withBytes{})},
		{"any range", `{"a":1e400}`, outOf[any](nil), outOf[any](nil)},
		{"from", `[1,2]`, outOf([]reportUnmarshalFrom(nil)), outOf([]reportStdUnmarshalFrom(nil))},
		{"array elements", `[1,"x",3]`, outOf([3]int{}), outOf([3]int{})},
		{"any range in a slice", `[{"a":1e400,"b":1},2]`, outOf([]any(nil)), outOf([]any(nil))},
		{"any after an error", `{"A":"x","B":{"c":1}}`, outOf(withAny{}), outOf(withAny{})},
		{"nested keys", `{"M":{"x":1,"2":2},"N":3}`, outOf(withKeys{}), outOf(withKeys{})},
		{"values of no representation", `{"a":1,"b":null}`, outOf(map[string]chan int(nil)), outOf(map[string]chan int(nil))},
		{"embedded map", `{"A":1,"b":"x","c":2,"C":3}`, outOf(withEmbeddedMap{}), outOf(withEmbeddedMap{})},
		{"names twice in any", `{"a":{"x":1},"a":{"y":1e400},"b":[{"c":1,"c":2}]}`, outOf[any](nil), outOf[any](nil)},
		{"names twice in any after an error", `{"X":"x","A":{"q":1,"q":2}}`, outOf(anyAfter{}), outOf(anyAfter{})},
		{"names twice in any of a merged kind", `{"A":{"q":[1],"q":{"s":1e400}},"X":"x"}`, outOf(anyAfter{}), outOf(anyAfter{})},
		{"names twice in any of a merged array", `{"A":{"q":{"s":1},"q":[1e400]},"X":"x"}`, outOf(anyAfter{}), outOf(anyAfter{})},
		{"names twice in any of another kind", `{"A":{"q":[1],"q":{"s":1},"r":2},"X":"x"}`, outOf(anyAfter{}), outOf(anyAfter{})},
		{"names twice in any in an array", `{"A":[{"q":[1],"q":{"s":1}},3],"X":"x"}`, outOf(anyAfter{}), outOf(anyAfter{})},
		{"empty array length", `{"A":[],"B":1}`, outOf(withArrayAfter{}), outOf(withArrayAfter{})},
		{"empty array length last", `{"B":1,"A":[]}`, outOf(withArrayAfter{}), outOf(withArrayAfter{})},
		{"empty array length in an array", `[[],[1,2]]`, outOf([][2]int(nil)), outOf([][2]int(nil))},
		{"empty array length in a map", `{"k":[],"j":[1,2]}`, outOf(map[string][2]int(nil)), outOf(map[string][2]int(nil))},
		{"pointers of failed values", `{"P":"x","Q":"y","R":[[1,2,3]],"N":2}`, outOf(withPointers{}), outOf(withPointers{})},
		{"set pointer of a failed value", `{"P":"x","N":2}`, setPointer, setPointer},
	}
	var cases []reportCase
	// ReportErrorsWithLegacySemantics alone and with every option of encoding/json, and both with the options of
	// jsontext of DefaultOptionsV1, by which the structs are decoded as v1 decodes them
	for _, s := range []struct {
		opts v1OptionSet
		asV1 bool
	}{{reportErrorsWithLegacySemantics, false}, {allV1Options, false}, {allV1Options, true}, {reportErrorsWithLegacySemantics, true}} {
		for _, in := range inputs {
			c := reportCase{name: in.name, in: in.in, ours: in.ours, theirs: in.theirs, opts: s.opts}
			if s.asV1 {
				c.name += " as v1"
				c.extra = []json.Options{jsontext.AllowDuplicateNames(true), jsontext.AllowInvalidUTF8(true)}
				c.stdExtra = []stdjson.Options{stdjsontext.AllowDuplicateNames(true), stdjsontext.AllowInvalidUTF8(true)}
			}
			cases = append(cases, c)
		}
	}
	// values which fail in the values held by an interface value, which they are merged into
	merged := func() any {
		var v any = map[string]any{"a": nil, "c": 3}
		return &v
	}
	mergedOther := func() any {
		var v any = map[string]any{"q": []any{1.0}}
		return &v
	}
	// a function which fails for 1
	errFunc := errors.New("func failed")
	unmarshalFloat := func(b []byte, v *float64) error {
		if string(b) == "1" {
			return errFunc
		}
		*v = 9
		return nil
	}
	folded, stdFolded := []json.Options{json.MatchCaseInsensitiveNames(true)}, []stdjson.Options{stdjson.MatchCaseInsensitiveNames(true)}
	return append(cases,
		reportCase{name: "merged any of another kind", in: `{"q":{"s":1},"r":2}`, ours: mergedOther, theirs: mergedOther,
			opts: reportErrorsWithLegacySemantics},
		reportCase{name: "merged any", in: `{"a":[1e400,2],"b":{"x":1e400,"y":2},"d":1}`, ours: merged, theirs: merged,
			opts: reportErrorsWithLegacySemantics},
		reportCase{name: "merged any of a function", in: `{"a":[1,2],"b":{"x":1,"y":2},"d":1}`, ours: merged, theirs: merged,
			opts: reportErrorsWithLegacySemantics, extra: []json.Options{json.WithUnmarshalers(json.UnmarshalFunc(unmarshalFloat))},
			stdExtra: []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFunc(unmarshalFloat))}},
		reportCase{name: "unknown", in: `{"B":1,"A":2}`, ours: outOf(rejected{}), theirs: outOf(rejected{}), opts: reportErrorsWithLegacySemantics,
			extra: []json.Options{json.RejectUnknownMembers(true)}, stdExtra: []stdjson.Options{stdjson.RejectUnknownMembers(true)}},
		reportCase{name: "ambiguous folded", in: `{"aA":1}`, ours: outOf(ambiguous{}), theirs: outOf(ambiguous{}),
			opts: reportErrorsWithLegacySemantics, extra: folded, stdExtra: stdFolded},
		reportCase{name: "kinds folded", in: `{"A":"x","b":"y","C":[1,"z",3]}`, ours: outOf(R2{}), theirs: outOf(R2{}),
			opts: reportErrorsWithLegacySemantics, extra: folded, stdExtra: stdFolded},
		reportCase{name: "duplicate folded", in: `{"A":"x","a":2}`, ours: outOf(R1{}), theirs: outOf(R1{}),
			opts: reportErrorsWithLegacySemantics, extra: folded, stdExtra: stdFolded})
}

// sameOutputs reports whether the outputs of ours and of encoding/json are the same: by reflect.DeepEqual for outputs
// of one type, which follows the pointers, and else by their Go syntax, in which the names of the types for
// encoding/json are the ones for ours.
func sameOutputs(ours, theirs any) bool {
	if reflect.TypeOf(ours) == reflect.TypeOf(theirs) {
		return reflect.DeepEqual(ours, theirs)
	}
	return fmt.Sprintf("%#v", valueOf(ours)) == namedAsOurs(fmt.Sprintf("%#v", valueOf(theirs)))
}

// valueOf is the value which out points to, or out if it is not a pointer.
func valueOf(out any) any {
	if v := reflect.ValueOf(out); v.Kind() == reflect.Pointer && !v.IsNil() {
		return v.Elem().Interface()
	}
	return out
}

// With ReportErrorsWithLegacySemantics, Unmarshal, UnmarshalRead and UnmarshalDecode report the errors of encoding/json,
// as encoding/json/v2 reports them with the option of encoding/json, and decode on after the errors of the values.
func TestUnmarshalReportErrorsMatchesStd(t *testing.T) {
	for _, c := range reportCases() {
		opts, stdOpts := append(c.opts.ours(), c.extra...), append(c.opts.std(), c.stdExtra...)
		ours, theirs := c.ours(), c.theirs()
		err := json.Unmarshal([]byte(c.in), ours, opts...)
		stdErr := stdjson.Unmarshal([]byte(c.in), theirs, stdOpts...)
		if g, w := describeV1Error(err), namedAsOurs(describeV1Error(stdErr)); g != w {
			t.Errorf("%s, options %v: Unmarshal:\n got %s\nwant %s", c.name, c.opts, g, w)
		} else if !sameOutputs(ours, theirs) {
			t.Errorf("%s, options %v: Unmarshal value:\n got %#v\nwant %#v", c.name, c.opts, valueOf(ours), valueOf(theirs))
		}
		for _, n := range readSizes {
			ours, theirs := c.ours(), c.theirs()
			err := json.UnmarshalRead(newChunkReader(c.in, n), ours, opts...)
			stdErr := stdjson.UnmarshalRead(newChunkReader(c.in, n), theirs, stdOpts...)
			if g, w := describeV1Error(err), namedAsOurs(describeV1Error(stdErr)); g != w {
				t.Errorf("%s, options %v, reads of %d: UnmarshalRead:\n got %s\nwant %s", c.name, c.opts, n, g, w)
			} else if !sameOutputs(ours, theirs) {
				t.Errorf("%s, options %v, reads of %d: UnmarshalRead value:\n got %#v\nwant %#v", c.name, c.opts, n,
					valueOf(ours), valueOf(theirs))
			}
			ours, theirs = c.ours(), c.theirs()
			err = json.UnmarshalDecode(jsontext.NewDecoder(newChunkReader(c.in, n)), ours, opts...)
			stdErr = stdjson.UnmarshalDecode(stdjsontext.NewDecoder(newChunkReader(c.in, n)), theirs, stdOpts...)
			if g, w := describeV1Error(err), namedAsOurs(describeV1Error(stdErr)); g != w {
				t.Errorf("%s, options %v, reads of %d: UnmarshalDecode:\n got %s\nwant %s", c.name, c.opts, n, g, w)
			} else if !sameOutputs(ours, theirs) {
				t.Errorf("%s, options %v, reads of %d: UnmarshalDecode value:\n got %#v\nwant %#v", c.name, c.opts, n,
					valueOf(ours), valueOf(theirs))
			}
		}
	}
}

// firstDecoderRead is the most bytes which the first read of a jsontext.Decoder over a reader reads: its first buffer of
// 4096 bytes ( see makeRoom of jsontext/decode.go ), less the room which it leaves after the input.
const firstDecoderRead = 4096 - textcoder.InputRoom

// With ReportErrorsWithLegacySemantics, the syntax error of an escape at the end of the first read of a decoder, which
// reads its input in parts, is the one of the whole input, as Unmarshal of encoding/json/v2 reports it, also for
// UnmarshalRead, whatever its reads are, which the error of encoding/json/v2 depends on: its text is cut at the end of
// a read.
func TestUnmarshalReportErrorsOfLongInputsMatchesStd(t *testing.T) {
	const badEscape, loneSurrogate = `\u12zz`, `"\ud800y"`
	for _, in := range []string{
		// the escape is cut by the end of the read, which its last byte is past
		`["` + strings.Repeat("a", firstDecoderRead-len(`["`)-len(`","`)-len(badEscape)+1) + `","` + badEscape + `"]`,
		// the string of the lone surrogate ends at the end of the read, which the text of the error is cut at
		`[` + strings.Repeat("1,", (firstDecoderRead-len(`[`)-len(loneSurrogate))/len("1,")) + loneSurrogate + `]`,
	} {
		for _, set := range []v1OptionSet{reportErrorsWithLegacySemantics, allV1Options} {
			var ours, theirs any
			err := json.Unmarshal([]byte(in), &ours, set.ours()...)
			want := describeV1Error(stdjson.Unmarshal([]byte(in), &theirs, set.std()...))
			if got := describeV1Error(err); got != want {
				t.Errorf("options %v: Unmarshal of %.20q...:\n got %s\nwant %s", set, in, got, want)
			}
			for _, n := range readSizes {
				err := json.UnmarshalRead(newChunkReader(in, n), &ours, set.ours()...)
				if got := describeV1Error(err); got != want {
					t.Errorf("options %v, reads of %d: UnmarshalRead of %.20q...:\n got %s\nwant %s", set, n, in, got, want)
				}
			}
		}
	}
}

// With ReportErrorsWithLegacySemantics, the errors of UnmarshalDecode of the values of an array or an object, which a
// decoder reads one after another, have the paths of the values in it, as encoding/json/v2 reports them.
func TestUnmarshalDecodeReportErrorsInValuesMatchesStd(t *testing.T) {
	type R struct{ A, B int }
	for _, in := range []string{`[{"A":"x"},{"B":1},"q",3]`, `{"a":{"A":"x"},"b/c":{"B":"y"},"d~":"q"}`} {
		for _, n := range readSizes {
			dec := jsontext.NewDecoder(newChunkReader(in, n))
			stdDec := stdjsontext.NewDecoder(newChunkReader(in, n))
			if _, err := dec.ReadToken(); err != nil {
				t.Fatal(err)
			}
			if _, err := stdDec.ReadToken(); err != nil {
				t.Fatal(err)
			}
			// a decoder which failed peeks no kind
			for i := 0; dec.PeekKind() != ']' && dec.PeekKind() != '}' && dec.PeekKind() != 0; i++ {
				if in[0] == '{' {
					// the name of the member
					if _, err := dec.ReadToken(); err != nil {
						t.Fatal(err)
					}
					if _, err := stdDec.ReadToken(); err != nil {
						t.Fatal(err)
					}
				}
				var ours, theirs R
				err := json.UnmarshalDecode(dec, &ours, reportErrorsWithLegacySemantics.ours()...)
				stdErr := stdjson.UnmarshalDecode(stdDec, &theirs, reportErrorsWithLegacySemantics.std()...)
				if g, w := describeV1Error(err), describeV1Error(stdErr); g != w || ours != theirs {
					t.Errorf("%s, reads of %d, value %d:\n got %s, %+v\nwant %s, %+v", in, n, i, g, ours, w, theirs)
				}
			}
			// both decoders end at the same kind: the end of the array or the object, or none after a failure
			if k, w := dec.PeekKind(), stdDec.PeekKind(); byte(k) != byte(w) {
				t.Errorf("%s, reads of %d: the values end at %v, want %v", in, n, k, w)
			}
		}
	}
}
