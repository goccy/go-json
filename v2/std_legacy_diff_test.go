//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsonv1 "encoding/json"
	stdjson "encoding/json/v2"
	"errors"
	"strings"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The options of encoding/json, of the root package and of encoding/json, given to the functions of both v2
// packages in the differential tests.

// legacySet is a set of the options of encoding/json, a bit of each in the order of legacyOptionPairs.
type legacySet uint16

const (
	looseBase64 legacySet = 1 << iota
	looseRFC3339
	arrayAnyLength
	caseSensitiveDelimiter
	mergeLegacy
	callMethodsLegacy
	byteArrayAsArray
	bytesLegacy
	durationAsNano
	reportErrorsLegacy
	stringifyLegacy
	omitEmptyLegacy

	// allLegacy is every option, as DefaultOptionsV1 sets them.
	allLegacy = omitEmptyLegacy<<1 - 1
)

// legacyOptionPairs are the options of the root package and of encoding/json, by the bits of legacySet.
var legacyOptionPairs = [...]struct {
	ours func(bool) json.Options
	std  func(bool) stdjson.Options
}{
	{gojson.ParseBytesWithLooseRFC4648, stdjsonv1.ParseBytesWithLooseRFC4648},
	{gojson.ParseTimeWithLooseRFC3339, stdjsonv1.ParseTimeWithLooseRFC3339},
	{gojson.UnmarshalArrayFromAnyLength, stdjsonv1.UnmarshalArrayFromAnyLength},
	{gojson.MatchCaseSensitiveDelimiter, stdjsonv1.MatchCaseSensitiveDelimiter},
	{gojson.MergeWithLegacySemantics, stdjsonv1.MergeWithLegacySemantics},
	{gojson.CallMethodsWithLegacySemantics, stdjsonv1.CallMethodsWithLegacySemantics},
	{gojson.FormatByteArrayAsArray, stdjsonv1.FormatByteArrayAsArray},
	{gojson.FormatBytesWithLegacySemantics, stdjsonv1.FormatBytesWithLegacySemantics},
	{gojson.FormatDurationAsNano, stdjsonv1.FormatDurationAsNano},
	{gojson.ReportErrorsWithLegacySemantics, stdjsonv1.ReportErrorsWithLegacySemantics},
	{gojson.StringifyWithLegacySemantics, stdjsonv1.StringifyWithLegacySemantics},
	{gojson.OmitEmptyWithLegacySemantics, stdjsonv1.OmitEmptyWithLegacySemantics},
}

// the bits of legacySet are the pairs
const (
	_ = uint(int(allLegacy+1) - 1<<len(legacyOptionPairs))
	_ = uint(1<<len(legacyOptionPairs) - int(allLegacy+1))
)

func (l legacySet) ours() []json.Options {
	var opts []json.Options
	for i, p := range legacyOptionPairs {
		if l&(1<<i) != 0 {
			opts = append(opts, p.ours(true))
		}
	}
	return opts
}

func (l legacySet) std() []stdjson.Options {
	var opts []stdjson.Options
	for i, p := range legacyOptionPairs {
		if l&(1<<i) != 0 {
			opts = append(opts, p.std(true))
		}
	}
	return opts
}

// legacyDiffOptions are the rows of the options of encoding/json in the marshal differential tests ( see
// diffOptions ).
func legacyDiffOptions() [][]json.Options {
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
		{gojson.StringifyWithLegacySemantics(true), json.StringifyNumbers(true), gojson.OmitEmptyWithLegacySemantics(true)},
		{gojson.OmitEmptyWithLegacySemantics(true), json.StringifyNumbers(true), json.OmitZeroStructFields(true),
			json.Deterministic(true)},
	}
}

// legacyDiffDelimiter is a struct of the names which MatchCaseSensitiveDelimiter matches or not.
type legacyDiffDelimiter struct {
	FooBar int `json:"foo_bar"`
	Baz    int `json:"BAZ,case:ignore"`
	QuxX   int `json:"qux-x,case:ignore"`
	Strict int `json:"strict_s,case:strict"`
}

// legacyDiffDelimiterTwo is a struct of two names which are the same but for an underscore.
type legacyDiffDelimiterTwo struct {
	AB  int `json:"a_b"`
	AB2 int `json:"ab"`
}

// legacyDiffDelimiterIgnore is a struct of two names of case:ignore which are the same but for an underscore.
type legacyDiffDelimiterIgnore struct {
	AB  int `json:"a_b,case:ignore"`
	AB2 int `json:"ab,case:ignore"`
}

// legacyDiffDelimiterFallback is a struct of an embedded fallback, whose names are matched with the names of the
// fields when marshaling.
type legacyDiffDelimiterFallback struct {
	FooBar int            `json:"foo_bar"`
	Baz    int            `json:"BAZ,case:ignore"`
	Rest   map[string]int `json:",embed"`
}

// legacyDiffStringify is a struct of the fields of the `string` option of StringifyWithLegacySemantics.
type legacyDiffStringify struct {
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

// legacyDiffStringifyPtr is a struct of a pointer to a pointer of the `string` option, which the option doesn't
// apply to with StringifyWithLegacySemantics.
type legacyDiffStringifyPtr struct {
	PP **int `json:",string"`
}

// legacyTextBool is a named bool whose UnmarshalText fails for "null" and sets it true for any other text.
type legacyTextBool bool

func (b *legacyTextBool) UnmarshalText(text []byte) error {
	if string(text) == "null" {
		return errLegacyNull
	}
	*b = true
	return nil
}

// errLegacyNull is the error of legacyTextBool, an error of encoding/json, which both packages report as it is.
var errLegacyNull error = &stdjsonv1.UnsupportedValueError{Str: "null"}

// legacyDiffStringifyPointers is a struct of pointers of the `string` option to values which the option doesn't apply
// to, and of a pointer of the option and another one to the same string.
type legacyDiffStringifyPointers struct {
	PSl *[]int          `json:",string"`
	PSt *legacyPair     `json:",string"`
	PM  *map[string]int `json:",string"`
	PT  *legacyTextBool `json:",string"`
	P   *string         `json:",string"`
	A   *string
}

// legacyDiffStringifyText is a struct of a string of the `string` option, of the texts which a JSON string escapes.
type legacyDiffStringifyText struct {
	S string `json:",string"`
}

// legacyPair is a struct of two fields, which a JSON object of one of them merges into or not.
type legacyPair struct{ X, Y int }

// legacyDiffMerge is a struct of the values which MergeWithLegacySemantics merges into or not.
type legacyDiffMerge struct {
	B   bool
	I   int
	U   uint
	F   float64
	S   string
	A   [2]int
	AP  [2]legacyPair
	St  legacyPair
	Sl  []legacyPair
	M   map[string]legacyPair
	MA  map[string]any
	P   *int
	T   time.Time
	Tx  legacyTextInByte
	BA  [2]byte
	Bs  []byte
	Any any
}

// newLegacyDiffMerge returns a value of legacyDiffMerge whose fields are not zero.
func newLegacyDiffMerge() any {
	n := 9
	sl := make([]legacyPair, 2, 3)
	sl[0], sl[1] = legacyPair{1, 2}, legacyPair{3, 4}
	return &legacyDiffMerge{B: true, I: 1, U: 2, F: 3, S: "s", A: [2]int{4, 5}, AP: [2]legacyPair{{1, 2}, {3, 4}},
		St: legacyPair{6, 7}, Sl: sl, M: map[string]legacyPair{"a": {1, 2}}, MA: map[string]any{"a": map[string]any{"x": 1.0}, "p": new(int)},
		P: &n, T: time.Unix(1, 0).UTC(), Tx: 3, BA: [2]byte{1, 2}, Bs: []byte{1}, Any: map[string]any{"k": 1.0}}
}

// The error of a string of invalid UTF-8 of the `string` option with StringifyWithLegacySemantics is the one of
// encoding/json/v2: a SemanticError of the type of the string, whose error is a SyntacticError.
func TestStringifyInvalidUTF8MatchesStd(t *testing.T) {
	v := legacyDiffStringifyText{S: "x\xff"}
	_, err := json.Marshal(v, gojson.StringifyWithLegacySemantics(true))
	_, stdErr := stdjson.Marshal(v, stdjsonv1.StringifyWithLegacySemantics(true))
	if err == nil || stdErr == nil || err.Error() != stdErr.Error() {
		t.Fatalf("got %v, want %v", err, stdErr)
	}
	var se *jsontext.SyntacticError
	if !errors.As(err, &se) {
		t.Fatalf("got %v, of no SyntacticError", err)
	}
}

// legacyEmptyText is a struct whose method writes an empty JSON string: never empty by its Go value.
type legacyEmptyText struct{ X int }

func (legacyEmptyText) MarshalJSON() ([]byte, error) { return []byte(`""`), nil }

// legacyRaw is a slice whose method writes an empty JSON string for no bytes: empty by its Go value if it has none.
type legacyRaw []byte

func (r legacyRaw) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte(`""`), nil
	}
	return r, nil
}

// legacyDiffOmitEmpty is a struct of the fields of omitempty of OmitEmptyWithLegacySemantics.
type legacyDiffOmitEmpty struct {
	B   bool               `json:",omitempty"`
	I   int                `json:",omitempty"`
	U   uint8              `json:",omitempty"`
	F   float64            `json:",omitempty"`
	S   string             `json:",omitempty"`
	P   *int               `json:",omitempty"`
	PP  **int              `json:",omitempty"`
	PS  *[]int             `json:",omitempty"`
	PM  *map[string]int    `json:",omitempty"`
	PT  *struct{}          `json:",omitempty"`
	Sl  []int              `json:",omitempty"`
	Bs  []byte             `json:",omitempty"`
	M   map[string]int     `json:",omitempty"`
	A0  [0]int             `json:",omitempty"`
	A1  [1]int             `json:",omitempty"`
	St  struct{}           `json:",omitempty"`
	Any any                `json:",omitempty"`
	Raw legacyRaw          `json:",omitempty"`
	MJ  legacyEmptyText    `json:",omitempty"`
	PMJ *legacyEmptyText   `json:",omitempty"`
	Z   int                `json:",omitempty,omitzero"`
	N   map[string]any     `json:",omitempty"`
	E   legacyEmbeddedOmit `json:",omitempty"`
}

// legacyEmbeddedOmit is a struct of a field of omitempty, which is never empty itself.
type legacyEmbeddedOmit struct {
	X []int `json:",omitempty"`
}

// legacyDiffOmitEmptyString is a struct of the fields of omitempty with the string option, which a bool and a string
// fail by when they are written, without StringifyWithLegacySemantics.
type legacyDiffOmitEmptyString struct {
	B bool    `json:",omitempty,string"`
	S string  `json:",omitempty,string"`
	I int     `json:",omitempty,string"`
	P *string `json:",omitempty,string"`
}

// legacyDiffDuration is a struct of the time.Duration values of FormatDurationAsNano.
type legacyDiffDuration struct {
	D    time.Duration
	P    *time.Duration
	S    time.Duration            `json:",string"`
	SP   *time.Duration           `json:",string"`
	E    time.Duration            `json:",omitempty"`
	Z    time.Duration            `json:",omitzero"`
	Keys map[time.Duration]string `json:",omitempty"`
	Any  any
}

// legacyDiffBytes is a struct of the bytes of FormatByteArrayAsArray and FormatBytesWithLegacySemantics.
type legacyDiffBytes struct {
	B   []byte
	NB  []legacyByte
	TB  []legacyTextByte
	A   [2]byte
	NA  [2]legacyByte
	TA  [2]legacyTextByte
	E   []legacyByte  `json:",omitempty"`
	EA  [2]legacyByte `json:",omitempty"`
	Z   [2]byte       `json:",omitzero"`
	P   *[2]byte
	Any any
}

// legacyPtrJSONByte is a named byte type whose MarshalJSON has a pointer receiver.
type legacyPtrJSONByte byte

func (*legacyPtrJSONByte) MarshalJSON() ([]byte, error) { return []byte(`"j"`), nil }

// legacyAppendByte is a named byte type of AppendText.
type legacyAppendByte byte

func (legacyAppendByte) AppendText(dst []byte) ([]byte, error) { return append(dst, 'a'), nil }

// legacyTextInByte is a named byte type of UnmarshalText only: its slices and arrays are of bytes with
// FormatBytesWithLegacySemantics when marshaling too.
type legacyTextInByte byte

func (b *legacyTextInByte) UnmarshalText(text []byte) error {
	*b = legacyTextInByte(len(text))
	return nil
}

// legacyJSONInByte is a named byte type of UnmarshalJSON only.
type legacyJSONInByte byte

func (b *legacyJSONInByte) UnmarshalJSON(data []byte) error {
	*b = legacyJSONInByte(len(data))
	return nil
}

// legacyNamedBytes is a named slice type of a named byte type.
type legacyNamedBytes []legacyByte

// legacyDiffValues are the values of the options of encoding/json in the marshal differential tests ( see
// diffValues ).
func legacyDiffValues() []any {
	d := time.Duration(-3)
	return []any{
		time.Duration(1500), &d, []time.Duration{1, 2}, map[time.Duration]int{5: 1, -2: 2}, map[string]time.Duration{"a": 7},
		legacyDiffDuration{}, legacyDiffDuration{D: 1, P: &d, S: 2, SP: &d, E: 3, Z: 4, Keys: map[time.Duration]string{9: "x"}, Any: d},
		[]byte{1, 2, 3}, []byte(nil), [3]byte{1, 2, 3}, [0]byte{}, []legacyByte{1, 2, 3}, []legacyByte(nil), [3]legacyByte{1, 2, 3},
		[]legacyTextByte{1, 2}, [2]legacyTextByte{1, 2},
		map[[2]byte]int{{1, 2}: 1}, map[[2]legacyByte]int{{1, 2}: 1}, map[[2]legacyTextByte]int{{1, 2}: 1},
		[]legacyPtrJSONByte{1}, [1]legacyPtrJSONByte{1}, []legacyAppendByte{1}, [1]legacyAppendByte{1},
		[]legacyTextInByte{1, 2}, []legacyJSONInByte{1, 2}, legacyNamedBytes{1, 2}, legacyNamedBytes(nil),
		[2][2]byte{{1, 2}, {3, 4}}, &[2]byte{1, 2}, &[]legacyByte{1},
		legacyDiffOmitEmpty{}, legacyDiffOmitEmpty{P: new(int), PP: new(*int), PS: &[]int{}, PM: &map[string]int{},
			PT: &struct{}{}, Sl: []int{}, Bs: []byte{}, M: map[string]int{}, Any: "", Raw: legacyRaw{}, PMJ: &legacyEmptyText{},
			N: map[string]any{}},
		legacyDiffOmitEmpty{B: true, I: -1, U: 1, F: 0.5, S: "x", Sl: []int{1}, Any: 0, Raw: legacyRaw(`1`), Z: 1,
			N: map[string]any{"a": nil}, E: legacyEmbeddedOmit{X: []int{}}},
		legacyDiffOmitEmptyString{}, legacyDiffOmitEmptyString{I: 3},
		legacyDiffStringify{}, legacyDiffStringify{B: true, S: "x", I: -5, U: 7, F: 1.5, PB: new(bool), PS: new(string), PI: new(int),
			E: "e", EB: true},
		legacyDiffStringifyPtr{}, legacyDiffStringifyPtr{PP: new(*int)},
		legacyDiffStringifyText{S: "q" + string(rune(34)) + "b" + string(rune(92)) + "<&>" + string(rune(0x2028))},
		legacyDiffStringifyText{S: "\xff"},
		legacyDiffDelimiterFallback{FooBar: 1, Rest: map[string]int{"foobar": 2}},
		legacyDiffDelimiterFallback{FooBar: 1, Rest: map[string]int{"FOO_BAR": 2}},
		legacyDiffDelimiterFallback{Baz: 1, Rest: map[string]int{"baz": 2}},
		legacyDiffDelimiterFallback{Rest: map[string]int{"b-a-z": 2, "x": 3}},
		legacyDiffBytes{}, legacyDiffBytes{B: []byte{1}, NB: []legacyByte{2}, TB: []legacyTextByte{3}, A: [2]byte{4, 5},
			NA: [2]legacyByte{6, 7}, TA: [2]legacyTextByte{1, 2}, E: []legacyByte{}, P: &[2]byte{8, 9}, Any: []legacyByte{1, 2}},
	}
}

// unmarshalDiffLegacyCases are the cases of the options of encoding/json.
func unmarshalDiffLegacyCases() []unmarshalDiffCase {
	anyLength := unmarshalDiffOptions{legacy: arrayAnyLength}
	loose := unmarshalDiffOptions{legacy: looseBase64}
	rfc := unmarshalDiffOptions{legacy: looseRFC3339}
	nano := unmarshalDiffOptions{legacy: durationAsNano}
	nanoStringify := unmarshalDiffOptions{legacy: durationAsNano, stringifyNumbers: true}
	arrayOfBytes := unmarshalDiffOptions{legacy: byteArrayAsArray}
	legacyBytes := unmarshalDiffOptions{legacy: bytesLegacy}
	bothBytes := unmarshalDiffOptions{legacy: byteArrayAsArray | bytesLegacy}
	bytesStruct := `{"B":[1],"NB":"AQ==","TB":["b"],"A":"AQI=","NA":[1,2],"TA":"AQI=","E":[],"P":[1,2],"Any":"AQ=="}`
	merge := unmarshalDiffOptions{legacy: mergeLegacy}
	mergeNulls := `{"B":null,"I":null,"U":null,"F":null,"S":null,"A":null,"AP":null,"St":null,"Sl":null,"M":null,"MA":null,` +
		`"P":null,"T":null,"Tx":null,"BA":null,"Bs":null,"Any":null}`
	mergeValues := `{"A":[8,9],"AP":[{"X":9},{"Y":8}],"St":{"X":9},"Sl":[{"X":9}],"M":{"a":{"X":9}},"MA":{"a":{"y":2},"p":5},` +
		`"Any":{"j":2},"BA":"AwQ=","Tx":"bb"}`
	stringify := unmarshalDiffOptions{legacy: stringifyLegacy}
	stringifyNumbers := unmarshalDiffOptions{legacy: stringifyLegacy, stringifyNumbers: true}
	newStringify := func() any { return new(legacyDiffStringify) }
	delimiter := unmarshalDiffOptions{legacy: caseSensitiveDelimiter}
	insensitive := unmarshalDiffOptions{matchCaseInsensitiveNames: true}
	both := unmarshalDiffOptions{legacy: caseSensitiveDelimiter, matchCaseInsensitiveNames: true}
	newDelimiter := func() any { return new(legacyDiffDelimiter) }
	newDelimiterTwo := func() any { return new(legacyDiffDelimiterTwo) }
	return []unmarshalDiffCase{
		// MergeWithLegacySemantics
		{name: "merge nulls", in: mergeNulls, out: newLegacyDiffMerge, opts: merge},
		{name: "merge nulls by default", in: mergeNulls, out: newLegacyDiffMerge},
		{name: "merge values", in: mergeValues, out: newLegacyDiffMerge, opts: merge},
		{name: "merge values by default", in: mergeValues, out: newLegacyDiffMerge},
		{name: "merge values of duplicate names", in: `{"MA":{"a":{"x":2},"a":{"y":3}},"M":{"a":{"Y":1},"a":{"X":5}}}`,
			out: newLegacyDiffMerge, opts: unmarshalDiffOptions{legacy: mergeLegacy, allowDuplicateNames: true}},
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
			var v any = legacyPair{1, 2}
			return &v
		}, opts: merge},
		{name: "merge a null of the string option", in: `{"I":null,"F":null}`, out: func() any {
			return &legacyDiffStringify{I: 1, F: 2}
		}, opts: merge},
		{name: "merge a quoted null of the string option", in: `{"I":"null","S":"null"}`, out: func() any {
			return &legacyDiffStringify{I: 1, S: "x"}
		}, opts: unmarshalDiffOptions{legacy: mergeLegacy | stringifyLegacy}},
		{name: "merge a long object", in: `{"X":"` + strings.Repeat("x", 600) + `",` + mergeValues[1:], out: func() any {
			v := newLegacyDiffMerge()
			return &struct {
				X string
				*legacyDiffMerge
			}{legacyDiffMerge: v.(*legacyDiffMerge)}
		}, opts: merge},
		// StringifyWithLegacySemantics
		{name: "stringify bool", in: `{"B":"true","EB":"false"}`, out: newStringify, opts: stringify},
		{name: "stringify bool by default", in: `{"B":"true"}`, out: newStringify},
		{name: "stringify bool unquoted", in: `{"B":true}`, out: newStringify, opts: stringify},
		{name: "stringify bool of null", in: `{"B":"null"}`, out: func() any { return &legacyDiffStringify{B: true} }, opts: stringify},
		{name: "stringify bool of another text", in: `{"B":"yes"}`, out: newStringify, opts: stringify},
		{name: "stringify string", in: `{"S":"\"x\\\"y\""}`, out: newStringify, opts: stringify},
		{name: "stringify string of null", in: `{"S":"null"}`, out: func() any { return &legacyDiffStringify{S: "x"} }, opts: stringify},
		{name: "stringify string not of a JSON string", in: `{"S":"x"}`, out: newStringify, opts: stringify},
		{name: "stringify string unquoted", in: `{"S":1}`, out: newStringify, opts: stringify},
		{name: "stringify numbers of the syntax of Go", in: `{"I":"+5","U":"007","F":"1e1"}`, out: newStringify, opts: stringify},
		{name: "stringify number of a hexadecimal", in: `{"I":"0x10"}`, out: newStringify, opts: stringify},
		{name: "stringify float of an infinity", in: `{"F":"-Inf"}`, out: newStringify, opts: stringify},
		{name: "stringify float of an underscore", in: `{"F":"1_0"}`, out: newStringify, opts: stringify},
		{name: "stringify number too big", in: `{"U":"300"}`, out: newStringify, opts: stringify},
		{name: "stringify numbers of null", in: `{"I":"null","F":"null"}`, out: func() any { return &legacyDiffStringify{I: 1, F: 2} }, opts: stringify},
		{name: "stringify pointers", in: `{"PB":"true","PS":"\"s\"","PI":"3"}`, out: newStringify, opts: stringify},
		{name: "stringify pointers of null", in: `{"PB":"null","PS":"null","PI":"null"}`, out: func() any {
			return &legacyDiffStringify{PB: new(bool), PS: new(string), PI: new(int)}
		}, opts: stringify},
		{name: "stringify pointers of a JSON null", in: `{"PB":null,"PS":null,"PI":null}`, out: func() any {
			return &legacyDiffStringify{PB: new(bool), PS: new(string), PI: new(int)}
		}, opts: stringify},
		{name: "stringify pointer to a pointer", in: `{"PP":"1"}`, out: func() any { return new(legacyDiffStringifyPtr) }, opts: stringify},
		{name: "stringify pointer to a pointer of null", in: `{"PP":null}`, out: func() any { return new(legacyDiffStringifyPtr) }, opts: stringify},
		{name: "stringify pointer to a pointer of an array", in: `{"PP":[1,{"a":2}]}`, out: func() any { return new(legacyDiffStringifyPtr) },
			opts: stringify},
		{name: "stringify pointer to a pointer cut", in: `{"PP":"1`, out: func() any { return new(legacyDiffStringifyPtr) }, opts: stringify},
		{name: "string option of a slice cut", in: `{"X":"1`, out: func() any {
			return new(struct {
				X []int `json:",string"`
			})
		}},
		{name: "stringify bool and string of a JSON null", in: `{"B":null,"S":null}`, out: func() any { return &legacyDiffStringify{B: true, S: "x"} },
			opts: stringify},
		{name: "stringify quoted null into a pointer to a slice", in: `{"PSl":"null"}`, out: func() any { return new(legacyDiffStringifyPointers) },
			opts: stringify},
		{name: "stringify quoted null into a pointer to a struct", in: `{"PSt":"null"}`, out: func() any { return new(legacyDiffStringifyPointers) },
			opts: stringify},
		{name: "stringify quoted null into a pointer to a map", in: `{"PM":"null"}`, out: func() any { return new(legacyDiffStringifyPointers) },
			opts: stringify},
		{name: "stringify quoted null into a pointer of a method", in: `{"PT":"null"}`, out: func() any { return new(legacyDiffStringifyPointers) },
			opts: stringify},
		{name: "stringify quoted null into a pointer which is set", in: `{"P":"null"}`, out: func() any {
			s := "keep"
			return &legacyDiffStringifyPointers{P: &s, A: &s}
		}, opts: stringify},
		{name: "stringify with StringifyNumbers", in: `["+5","null","1"]`, out: outOf([]int(nil)), opts: stringifyNumbers},
		{name: "stringify map keys", in: `{"null":1,"+2":2}`, out: func() any { return new(map[int]int) }, opts: stringify},
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
			out: func() any { return new(legacyDiffDelimiterIgnore) }, opts: delimiter},
		{name: "delimiter names of case:ignore of the same fold by default", in: `{"A_B":1}`,
			out: func() any { return new(legacyDiffDelimiterIgnore) }},
		{name: "delimiter names of a member twice", in: `{"foo_bar":1,"FOO_BAR":2}`, out: newDelimiter, opts: both},
		{name: "delimiter names in a long object", in: `{"x":"` + strings.Repeat("x", 600) + `","FOO_BAR":1,"foobar":2,"Baz":3}`,
			out: newDelimiter, opts: both},
		// FormatByteArrayAsArray
		{name: "byte array as array", in: `[1,2,3]`, out: outOf([3]byte{}), opts: arrayOfBytes},
		{name: "byte array as array of base64", in: `"AQID"`, out: outOf([3]byte{}), opts: arrayOfBytes},
		{name: "byte array as array by default", in: `[1,2,3]`, out: outOf([3]byte{})},
		{name: "byte array as array stringified", in: `["1","2"]`, out: outOf([2]byte{}),
			opts: unmarshalDiffOptions{legacy: byteArrayAsArray, stringifyNumbers: true}},
		{name: "byte array as array of a number too big", in: `[1,256]`, out: outOf([2]byte{}), opts: arrayOfBytes},
		{name: "byte slice with FormatByteArrayAsArray", in: `"AQID"`, out: outOf([]byte(nil)), opts: arrayOfBytes},
		{name: "byte array key as array", in: `{"AQI=":1}`, out: func() any { return new(map[[2]byte]int) }, opts: arrayOfBytes},
		{name: "named byte array as array", in: `[1,2]`, out: outOf([2]legacyByte{}), opts: bothBytes},
		// FormatBytesWithLegacySemantics
		{name: "legacy bytes of a JSON array", in: `[1,2,3]`, out: outOf([]byte(nil)), opts: legacyBytes},
		{name: "legacy bytes of a JSON array by default", in: `[1,2,3]`, out: outOf([]byte(nil))},
		{name: "legacy byte array of a JSON array", in: `[1,2]`, out: outOf([2]byte{}), opts: legacyBytes},
		{name: "legacy bytes of a JSON array of an error", in: `[1,"x",3]`, out: outOf([]byte(nil)), opts: legacyBytes},
		{name: "legacy bytes of a long JSON array", in: "[" + strings.Repeat("1,", 600) + "1]", out: outOf([]byte(nil)), opts: legacyBytes},
		{name: "legacy named bytes of base64", in: `"AQID"`, out: outOf([]legacyByte(nil)), opts: legacyBytes},
		{name: "legacy named bytes of base64 by default", in: `"AQID"`, out: outOf([]legacyByte(nil))},
		{name: "legacy named byte array of base64", in: `"AQI="`, out: outOf([2]legacyByte{}), opts: legacyBytes},
		{name: "legacy named bytes of null", in: `null`, out: func() any { return &[]legacyByte{1} }, opts: legacyBytes},
		{name: "legacy named bytes of an object", in: `{"a":1}`, out: outOf([]legacyByte(nil)), opts: legacyBytes},
		{name: "legacy named bytes of an object of a syntax error", in: `{"a":1,}`, out: outOf([]legacyByte(nil)), opts: legacyBytes},
		{name: "legacy named bytes of a bool", in: `true`, out: outOf([]legacyByte(nil)), opts: legacyBytes},
		{name: "legacy text bytes of base64", in: `"AQI="`, out: outOf([]legacyTextByte(nil)), opts: legacyBytes},
		{name: "legacy text bytes of a JSON array", in: `["bb","b"]`, out: outOf([]legacyTextByte(nil)), opts: legacyBytes},
		{name: "legacy text bytes of a JSON array by default", in: `["bb","b"]`, out: outOf([]legacyTextByte(nil))},
		{name: "legacy byte array of a longer JSON array", in: `[1,2,3]`, out: outOf([2]byte{}), opts: legacyBytes},
		{name: "legacy byte array of a shorter JSON array", in: `[1]`, out: outOf([2]byte{}), opts: legacyBytes},
		{name: "legacy named byte array of a longer JSON array", in: `[1,2,3]`, out: outOf([2]legacyByte{}), opts: legacyBytes},
		{name: "legacy byte array of a field of a longer JSON array", in: `{"A":[1,2,3]}`,
			out: func() any { return new(legacyDiffBytes) }, opts: legacyBytes},
		{name: "legacy byte arrays of JSON arrays of other lengths", in: `[[1,2,3],[4,5]]`, out: outOf([][2]byte(nil)),
			opts: legacyBytes},
		{name: "byte arrays as arrays of other lengths", in: `[[1,2,3],[4,5]]`, out: outOf([][2]byte(nil)), opts: arrayOfBytes},
		{name: "legacy text-in bytes of base64", in: `"AQI="`, out: outOf([]legacyTextInByte(nil)), opts: legacyBytes},
		{name: "legacy text-in bytes of a JSON array", in: `["x","yy"]`, out: outOf([]legacyTextInByte(nil)), opts: legacyBytes},
		{name: "legacy JSON-in bytes of base64", in: `"AQI="`, out: outOf([]legacyJSONInByte(nil)), opts: legacyBytes},
		{name: "legacy JSON-in bytes of a JSON array", in: `[1,"x"]`, out: outOf([]legacyJSONInByte(nil)), opts: legacyBytes},
		{name: "legacy named slice of base64", in: `"AQI="`, out: outOf(legacyNamedBytes(nil)), opts: legacyBytes},
		{name: "legacy pointer to named bytes", in: `"AQI="`, out: outOf((*[]legacyByte)(nil)), opts: legacyBytes},
		{name: "pointer to a byte array as an array", in: `[1,2]`, out: outOf((*[2]byte)(nil)), opts: arrayOfBytes},
		{name: "legacy named byte array key", in: `{"AQI=":1}`, out: func() any { return new(map[[2]legacyByte]int) }, opts: legacyBytes},
		{name: "legacy bytes struct", in: bytesStruct, out: func() any { return new(legacyDiffBytes) }, opts: legacyBytes},
		{name: "legacy bytes struct with arrays", in: bytesStruct, out: func() any { return new(legacyDiffBytes) }, opts: bothBytes},
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
			out: func() any { return new(legacyDiffDuration) }, opts: nano},
		{name: "duration nano string option of a number", in: `{"S":3}`, out: func() any { return new(legacyDiffDuration) }, opts: nano},
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
			in: `[1,2,"` + strings.Repeat("x", 600) + "\xff\"]", out: outOf([2]int{})},
		{name: "array longer with a name twice past the length without the option of the text", in: `[1,2,{"a":1,"a":2}]`,
			out: outOf([2]int{}), opts: unmarshalDiffOptions{allowInvalidUTF8: true}},
		{name: "array longer with invalid UTF-8 past the length without the option of the text", in: "[1,2,\"\xff\"]",
			out: outOf([2]int{}), opts: unmarshalDiffOptions{allowDuplicateNames: true}},
		{name: "array longer with invalid UTF-8 past the length allowed", in: "[1,2,\"\xff\"]", out: outOf([2]int{}),
			opts: unmarshalDiffOptions{legacy: arrayAnyLength, allowInvalidUTF8: true}},
		{name: "array longer with a name twice past the length allowed", in: `[1,2,{"a":1,"a":2}]`, out: outOf([2]int{}),
			opts: unmarshalDiffOptions{legacy: arrayAnyLength, allowDuplicateNames: true}},
		{name: "array longer with a long string past the length", in: `[1,2,"` + strings.Repeat("x", 600) + `"]`, out: outOf([2]int{}), opts: anyLength},
		{name: "array longer with invalid UTF-8 in a long string past the length", in: `[1,2,"` + strings.Repeat("x", 600) + "\xff\"]",
			out: outOf([2]int{}), opts: anyLength},
		{name: "array longer with invalid UTF-8 in a long string past the length allowed", in: `[1,2,"` + strings.Repeat("x", 600) + "\xff\"]",
			out: outOf([2]int{}), opts: unmarshalDiffOptions{legacy: arrayAnyLength, allowInvalidUTF8: true}},
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
