//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// cyclePointer is a pointer to itself, which holds null or itself, and cycleString a struct of a field of it of the
// `string` option.
type cyclePointer *cyclePointer

type cycleString struct {
	P cyclePointer `json:",string"`
	N int
}

// cycleA and cycleB are pointers to each other, and cycleTwo a struct of a field of them of the `string` option.
type (
	cycleA *cycleB
	cycleB *cycleA
)

type cycleTwo struct {
	A cycleA `json:",string"`
	N int
}

// The `string` option of a pointer to itself doesn't apply to it: it is written and read as without the option, as
// encoding/json/v2 writes and reads it, with the options of encoding/json too. The walk of its pointers ends.
func TestStringOptionOfPointerCycleMatchesStd(t *testing.T) {
	// pointers of the cycles to nil, whose end is null
	var a cyclePointer
	b := cyclePointer(&a)
	var y cycleB
	x := cycleA(&y)
	values := []any{cycleString{N: 1}, cycleString{P: b, N: 1}, cycleString{P: &b, N: 1}, cycleTwo{A: x, N: 1}}
	for _, set := range []legacySet{0, reportErrorsLegacy, stringifyLegacy, reportErrorsLegacy | stringifyLegacy, callMethodsLegacy} {
		for _, v := range values {
			want, wantErr := stdjson.Marshal(v, set.std()...)
			got, err := json.Marshal(v, set.ours()...)
			if g, w := describeLegacyError(err), describeLegacyError(wantErr); g != w || err == nil && string(got) != string(want) {
				t.Errorf("options %b: Marshal of %#v:\n got %s %s\nwant %s %s", set, v, got, g, want, w)
			}
		}
		var err error
		var cycle cyclePointer
		cycle = &cycle
		// encoding/json/v2 overflows its stack for the value, which is an error of go-json
		if _, err := json.Marshal(cycleString{P: cycle}, set.ours()...); err == nil {
			t.Errorf("options %b: Marshal of a cycle: got no error", set)
		}
		var ours, theirs cycleString
		err = json.Unmarshal([]byte(`{"P":null,"N":1}`), &ours, set.ours()...)
		stdErr := stdjson.Unmarshal([]byte(`{"P":null,"N":1}`), &theirs, set.std()...)
		if g, w := describeLegacyError(err), describeLegacyError(stdErr); g != w || ours.P != nil || ours.N != theirs.N {
			t.Errorf("options %b: Unmarshal of null:\n got %s %+v\nwant %s %+v", set, g, ours, w, theirs)
		}
		// encoding/json/v2 overflows its stack for a value which is not null, which is an error of go-json
		for _, in := range []string{`{"P":1,"N":1}`, `{"P":"1","N":1}`, `{"P":"null","N":1}`} {
			var ours cycleString
			if err := json.Unmarshal([]byte(in), &ours, set.ours()...); err == nil {
				t.Errorf("options %b: Unmarshal of %s: got no error", set, in)
			}
		}
	}
}

// cyclePointerFuncs has fields of the `string` option of pointers to a pointer of a cycle, which a function writes,
// and doublePointerFunc one of a pointer to a pointer which a function writes.
type cyclePointerFuncs struct {
	P  *cyclePointer  `json:",string"`
	PP **cyclePointer `json:",string"`
	N  int
}

type doublePointerFunc struct {
	D **int `json:",string"`
	N int
}

// A function of a pointer type on the way to a cycle of pointers writes its value of the `string` option with the
// option, as encoding/json/v2 calls it, but for a pointer to a pointer with StringifyWithLegacySemantics, which is the
// error of the option, or, with ReportErrorsWithLegacySemantics, a value which it doesn't apply to, whatever functions
// the pointers have.
func TestStringOptionOfFuncsOfPointersMatchesStd(t *testing.T) {
	ours := json.WithMarshalers(json.JoinMarshalers(
		json.MarshalToFunc(func(enc *jsontext.Encoder, _ *cyclePointer) error { return json.MarshalEncode(enc, 7) }),
		json.MarshalToFunc(func(enc *jsontext.Encoder, _ **int) error { return json.MarshalEncode(enc, 8) })))
	theirs := stdjson.WithMarshalers(stdjson.JoinMarshalers(
		stdjson.MarshalToFunc(func(enc *stdjsontext.Encoder, _ *cyclePointer) error { return stdjson.MarshalEncode(enc, 7) }),
		stdjson.MarshalToFunc(func(enc *stdjsontext.Encoder, _ **int) error { return stdjson.MarshalEncode(enc, 8) })))
	// functions of the types of the fields, which the values are written by, with the option
	oursOfFields := json.WithMarshalers(json.JoinMarshalers(
		json.MarshalFunc(func(***int) ([]byte, error) { return []byte("9"), nil }),
		json.MarshalToFunc(func(enc *jsontext.Encoder, _ *cyclePointer) error { return json.MarshalEncode(enc, 9) })))
	theirsOfFields := stdjson.WithMarshalers(stdjson.JoinMarshalers(
		stdjson.MarshalFunc(func(***int) ([]byte, error) { return []byte("9"), nil }),
		stdjson.MarshalToFunc(func(enc *stdjsontext.Encoder, _ *cyclePointer) error { return stdjson.MarshalEncode(enc, 9) })))
	var c cyclePointer
	q := &c
	n := 1
	pn := &n
	for _, set := range []legacySet{0, reportErrorsLegacy, stringifyLegacy, reportErrorsLegacy | stringifyLegacy, allLegacy} {
		for i, funcs := range [][2]any{{ours, theirs}, {oursOfFields, theirsOfFields}} {
			for _, v := range []any{cyclePointerFuncs{P: &c, PP: &q, N: 1}, doublePointerFunc{D: &pn, N: 1}, cycleString{N: 1}} {
				want, wantErr := stdjson.Marshal(v, append(set.std(), funcs[1].(stdjson.Options))...)
				got, err := json.Marshal(v, append(set.ours(), funcs[0].(json.Options))...)
				if g, w := describeLegacyError(err), describeLegacyError(wantErr); g != w || err == nil && string(got) != string(want) {
					t.Errorf("options %b, functions %d: Marshal of %T:\n got %s %s\nwant %s %s", set, i, v, got, g, want, w)
				}
			}
		}
	}
}

// doublePointerField has a field of the `string` option of a pointer to a pointer.
type doublePointerField struct {
	F **int `json:",string"`
	N int
}

// A pointer to a pointer of the `string` option, which a function of its type declines, is decoded without the option
// with ReportErrorsWithLegacySemantics and StringifyWithLegacySemantics, as encoding/json/v2 decodes it.
func TestStringOptionOfDecliningFuncOfPointersMatchesStd(t *testing.T) {
	ours := json.WithUnmarshalers(json.UnmarshalFromFunc(func(*jsontext.Decoder, ***int) error { return errors.ErrUnsupported }))
	theirs := stdjson.WithUnmarshalers(stdjson.UnmarshalFromFunc(func(*stdjsontext.Decoder, ***int) error { return errors.ErrUnsupported }))
	for _, set := range []legacySet{reportErrorsLegacy | stringifyLegacy, allLegacy} {
		for _, in := range []string{`{"F":1,"N":1}`, `{"F":"1","N":1}`, `{"F":null,"N":1}`} {
			var got, want doublePointerField
			err := json.Unmarshal([]byte(in), &got, append(set.ours(), ours)...)
			stdErr := stdjson.Unmarshal([]byte(in), &want, append(set.std(), theirs)...)
			if g, w := describeLegacyError(err), describeLegacyError(stdErr); g != w || got.N != want.N ||
				(got.F == nil) != (want.F == nil) || got.F != nil && (*got.F == nil) != (*want.F == nil) ||
				got.F != nil && *got.F != nil && **got.F != **want.F {
				t.Errorf("options %b: Unmarshal of %s:\n got %s %+v\nwant %s %+v", set, in, g, got, w, want)
			}
		}
	}
}
