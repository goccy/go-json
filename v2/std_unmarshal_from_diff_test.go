//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The tests here compare the decoding by the methods and the functions which read a jsontext.Decoder with the one
// of encoding/json/v2, as the tests of std_unmarshal_diff_test.go do. The decoder of each package is a type of its
// own, so the types of the methods are made for each, and read their decoders by ufRead.

// ufDecoder is the decoder of either package, of its own types of the values, the pointers, the kinds and the
// tokens.
type ufDecoder[V ~[]byte, P ~string, K ~byte, T any] interface {
	ReadValue() (V, error)
	ReadToken() (T, error)
	PeekKind() K
	StackDepth() int
	StackIndex(int) (K, int64)
	StackPointer() P
	InputOffset() int64
}

// ufRead reads dec as the kind of its next value says, and keeps in got what it read: a string is read and kept with
// the pointer, the depth, the levels and the offset of the decoder after it, and opts, the options it has, but
// "unsupported", which is followed by errors.ErrUnsupported, and "two", which reads the next value too; an array
// is read whole and kept; null is declined; true is not read; false and a number are read and followed by an
// error; and an object is read up to its first name, or whole if it has none, and followed by an error.
func ufRead[V ~[]byte, P ~string, K ~byte, T any, D ufDecoder[V, P, K, T]](dec D, got *string, opts string) error {
	switch dec.PeekKind() {
	case 'n':
		return errors.ErrUnsupported
	case 't':
		return nil
	case 'f':
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		return errors.New("false")
	case '0':
		if _, err := dec.ReadValue(); err != nil {
			return err
		}
		return errors.New("a number")
	case '{':
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		if dec.PeekKind() == '}' {
			if _, err := dec.ReadToken(); err != nil {
				return err
			}
			return errors.New("an empty object")
		}
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		return errors.New("an object")
	}
	from := dec.InputOffset()
	v, err := dec.ReadValue()
	if err != nil {
		return err
	}
	levels := ""
	for i := range dec.StackDepth() + 1 {
		k, n := dec.StackIndex(i)
		levels += fmt.Sprintf("%c%d", max(k, '0'), n)
	}
	*got = fmt.Sprintf("%s at %s, %d, %s, from %d to %d, %s", v, dec.StackPointer(), dec.StackDepth(), levels, from, dec.InputOffset(), opts)
	switch string(v) {
	case `"unsupported"`:
		return errors.ErrUnsupported
	case `"two"`:
		_, err := dec.ReadValue()
		return err
	}
	return nil
}

// ufOurs and ufStd are decoded by UnmarshalJSONFrom of each package ( see ufRead ), and by UnmarshalJSON when it
// declines.
type ufOurs struct{ Got string }

func (u *ufOurs) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	stringify, _ := json.GetOption(dec.Options(), json.StringifyNumbers)
	duplicates, _ := json.GetOption(dec.Options(), jsontext.AllowDuplicateNames)
	return ufRead(dec, &u.Got, fmt.Sprint(stringify, duplicates))
}

func (u *ufOurs) UnmarshalJSON(b []byte) error {
	u.Got = "UnmarshalJSON " + string(b)
	return nil
}

type ufStd struct{ Got string }

func (u *ufStd) UnmarshalJSONFrom(dec *stdjsontext.Decoder) error {
	stringify, _ := stdjson.GetOption(dec.Options(), stdjson.StringifyNumbers)
	duplicates, _ := stdjson.GetOption(dec.Options(), stdjsontext.AllowDuplicateNames)
	return ufRead(dec, &u.Got, fmt.Sprint(stringify, duplicates))
}

func (u *ufStd) UnmarshalJSON(b []byte) error {
	u.Got = "UnmarshalJSON " + string(b)
	return nil
}

// ufOursTokens and ufStdTokens read their value token by token, to its end, and fail.
type ufOursTokens struct{}

func (*ufOursTokens) UnmarshalJSONFrom(dec *jsontext.Decoder) error { return ufReadTokens(dec) }

type ufStdTokens struct{}

func (*ufStdTokens) UnmarshalJSONFrom(dec *stdjsontext.Decoder) error { return ufReadTokens(dec) }

func ufReadTokens[V ~[]byte, P ~string, K ~byte, T any, D ufDecoder[V, P, K, T]](dec D) error {
	depth := dec.StackDepth()
	for {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		if dec.StackDepth() == depth {
			return errors.New("read by tokens")
		}
	}
}

type ufFieldsOurs struct {
	A ufOurs
	B *ufOurs
	C []ufOurs
}

type ufFieldsStd struct {
	A ufStd
	B *ufStd
	C []ufStd
}

// ufValue is decoded by the functions of the cases, which are of either package for the same type.
type ufValue struct{ Got string }

// ufFrom returns the options of UnmarshalFromFunc of each package, of a function which reads by ufRead.
func ufFrom() umOptions {
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *ufValue) error {
			return ufRead(dec, &v.Got, "")
		}))},
		std: []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFromFunc(func(dec *stdjsontext.Decoder, v *ufValue) error {
			return ufRead(dec, &v.Got, "")
		}))},
	}
}

// ufFromString returns the options of a function of strings, which reads by ufRead into a string.
func ufFromString() umOptions {
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
			return ufRead(dec, v, "")
		}))},
		std: []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFromFunc(func(dec *stdjsontext.Decoder, v *string) error {
			return ufRead(dec, v, "")
		}))},
	}
}

// ufFromAny returns the options of a function of interface values, which reads by ufRead into a string it sets.
func ufFromAny() umOptions {
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *any) error {
			var got string
			err := ufRead(dec, &got, "")
			*v = got
			return err
		}))},
		std: []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFromFunc(func(dec *stdjsontext.Decoder, v *any) error {
			var got string
			err := ufRead(dec, &got, "")
			*v = got
			return err
		}))},
	}
}

// ufFromQuotedInt returns the options of a function of integers, which decodes an integer from a string and declines
// any other value.
func ufFromQuotedInt() umOptions {
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *int) error {
			if dec.PeekKind() != '"' {
				return errors.ErrUnsupported
			}
			b, err := dec.ReadValue()
			if err == nil {
				_, err = fmt.Sscanf(string(b), "%q", new(string))
				*v = len(b)
			}
			return err
		}))},
		std: []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFromFunc(func(dec *stdjsontext.Decoder, v *int) error {
			if dec.PeekKind() != '"' {
				return errors.ErrUnsupported
			}
			b, err := dec.ReadValue()
			if err == nil {
				_, err = fmt.Sscanf(string(b), "%q", new(string))
				*v = len(b)
			}
			return err
		}))},
	}
}

// ufFromThenFunc returns the options of the function of ufFrom joined with a function of UnmarshalFunc of the same
// type after it, which decodes what the first one declines.
func ufFromThenFunc() umOptions {
	then := umFuncs(func(b []byte, v *ufValue) error {
		v.Got = "UnmarshalFunc " + string(b)
		return nil
	})
	return umJoin(ufFrom(), then)
}

// ufTree and ufNode have values of themselves, which a function which declines some of their values decodes too.
type ufTree []ufTree

type ufNode struct {
	N *ufNode
	S []ufNode
	V int
}

// ufDecliningValues returns the options of a function of T, which decodes null into a value of T with a mark, and
// declines any other value: the values of T in it are decoded by the function again.
func ufDecliningValues[T any](mark func(*T)) umOptions {
	decode := func(read func() ([]byte, error), peek func() byte, v *T) error {
		if peek() != 'n' {
			return errors.ErrUnsupported
		}
		if _, err := read(); err != nil {
			return err
		}
		mark(v)
		return nil
	}
	return umOptions{
		ours: []json.Options{json.WithUnmarshalers(json.UnmarshalFromFunc(func(dec *jsontext.Decoder, v *T) error {
			return decode(func() ([]byte, error) { return dec.ReadValue() }, func() byte { return byte(dec.PeekKind()) }, v)
		}))},
		std: []stdjson.Options{stdjson.WithUnmarshalers(stdjson.UnmarshalFromFunc(func(dec *stdjsontext.Decoder, v *T) error {
			return decode(func() ([]byte, error) { return dec.ReadValue() }, func() byte { return byte(dec.PeekKind()) }, v)
		}))},
	}
}

type ufCase struct {
	name         string
	in           string
	ours, theirs func() any
	opts         umOptions
}

// ufOut returns the functions which make a pointer to a new zero value of O and of S.
func ufOut[O, S any]() (func() any, func() any) {
	return func() any { return new(O) }, func() any { return new(S) }
}

func ufCases() []ufCase {
	c := func(name, in string, out func() (func() any, func() any), opts umOptions) ufCase {
		ours, theirs := out()
		return ufCase{name: name, in: in, ours: ours, theirs: theirs, opts: opts}
	}
	invalidUTF8 := umOptions{ours: []json.Options{jsontext.AllowInvalidUTF8(true)}, std: []stdjson.Options{stdjsontext.AllowInvalidUTF8(true)}}
	duplicates := umOptions{ours: []json.Options{jsontext.AllowDuplicateNames(true)}, std: []stdjson.Options{stdjsontext.AllowDuplicateNames(true)}}
	stringify := umOptions{ours: []json.Options{json.StringifyNumbers(true)}, std: []stdjson.Options{stdjson.StringifyNumbers(true)}}
	return []ufCase{
		c("method at the top level", ` "a" `, ufOut[ufOurs, ufStd], umOptions{}),
		c("method of elements", `["a", "b",  "c"]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method of nested places", `{"x":[{"y":"a"},{"z":"b","w":["c"]}]}`, ufOut[map[string][]map[string]ufOurs, map[string][]map[string]ufStd], umOptions{}),
		c("method of an escaped name", `{"\u0078":"a"}`, ufOut[map[string]ufOurs, map[string]ufStd], umOptions{}),
		c("method of fields", `{"A":"a","B":"b","C":["c",["d"]]}`, ufOut[ufFieldsOurs, ufFieldsStd], umOptions{}),
		c("method of a null pointer", `{"B":null}`, ufOut[ufFieldsOurs, ufFieldsStd], umOptions{}),
		c("method of an array", `[[1, {"a" : 2}]]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method declining", `["a", null]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method reading nothing", `["a", true]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method error after its value", `{"x":1}`, ufOut[map[string]ufOurs, map[string]ufStd], umOptions{}),
		c("method error in an object", `[{"a":1}]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method error after the begin of an object", `[ {} ]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method error after a literal", `{"a": false}`, ufOut[map[string]ufOurs, map[string]ufStd], umOptions{}),
		c("method error after a literal at the top level", ` false `, ufOut[ufOurs, ufStd], umOptions{}),
		c("method of an array type", `["a", ["b"]]`, ufOut[[2]ufOurs, [2]ufStd], umOptions{}),
		c("method of white space around delimiters", `[ "a" , "b" ,"c"]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method of a member of white space", `{ "a" : "x" , "b":"y"}`, ufOut[map[string]ufOurs, map[string]ufStd], umOptions{}),
		c("method of keys of white space", `{ "a" : 1 , "b" :2}`, ufOut[map[ufOurs]int, map[ufStd]int], umOptions{}),
		c("method error after the end of an array", ` [true] `, ufOut[ufOursTokens, ufStdTokens], umOptions{}),
		c("method error after the end of nested arrays", `[[[{"a":null}]]]`, ufOut[[]ufOursTokens, []ufStdTokens], umOptions{}),
		c("method error after the end of a field", `{"X":[true]}`, ufOut[struct{ X ufOursTokens }, struct{ X ufStdTokens }], umOptions{}),
		c("method of no input", ``, ufOut[ufOurs, ufStd], umOptions{}),
		c("method of white space", ` `, ufOut[ufOurs, ufStd], umOptions{}),
		c("method of an element cut", `[`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method of an element cut after a comma", `["a",`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method of a member cut", `{"a":`, ufOut[map[string]ufOurs, map[string]ufStd], umOptions{}),
		c("method of a key cut", `{"a":1,`, ufOut[map[ufOurs]int, map[ufStd]int], umOptions{}),
		c("method of a first key cut", `{`, ufOut[map[ufOurs]int, map[ufStd]int], umOptions{}),
		c("method error in a nested object", `{"x":{"y":{"a":1}}}`, ufOut[map[string]map[string]ufOurs, map[string]map[string]ufStd], umOptions{}),
		c("method reading the next value", `["two", "x"]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method unsupported after reading", `["unsupported"]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method syntax error in its value", `[[1,}]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method syntax error after its value", `["a" "b"]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method duplicate name in its value", `[[{"a":1,"a":2}]]`, ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method duplicate name in its value allowed", `[[{"a":1,"a":2}]]`, ufOut[[]ufOurs, []ufStd], duplicates),
		c("method invalid UTF-8 in its value", "[\"a\xffb\"]", ufOut[[]ufOurs, []ufStd], umOptions{}),
		c("method invalid UTF-8 in its value allowed", "[\"a\xffb\"]", ufOut[[]ufOurs, []ufStd], invalidUTF8),
		c("method with options", `["a"]`, ufOut[[]ufOurs, []ufStd], stringify),
		c("method with other options", `["a"]`, ufOut[[]ufOurs, []ufStd], duplicates),
		c("method of map keys", `{"a":1, "b":2}`, ufOut[map[ufOurs]int, map[ufStd]int], umOptions{}),
		c("method of map keys of a name twice", `{"a":1,"a":2}`, ufOut[map[ufOurs]int, map[ufStd]int], umOptions{}),
		c("function", `["a",["b"]]`, ufOut[[]ufValue, []ufValue], ufFrom()),
		c("function error", `[1]`, ufOut[[]ufValue, []ufValue], ufFrom()),
		c("function declining to the next one", `["a", null]`, ufOut[[]ufValue, []ufValue], ufFromThenFunc()),
		c("function declining to the default", `[1, "22"]`, ufOut[[]int, []int], ufFromQuotedInt()),
		c("function declining to a default error", `[true]`, ufOut[[]int, []int], ufFromQuotedInt()),
		c("function of strings into any", `{"a":"x","b":["y"]}`, ufOut[any, any], ufFromString()),
		c("function of strings cut", `["a",`, ufOut[[]string, []string], ufFromString()),
		c("function of any", `{"a":"x","b":["y"]}`, ufOut[map[string]any, map[string]any], ufFromAny()),
		c("function of any cut", `{"a":`, ufOut[any, any], ufFromAny()),
		c("function of strings into a map of strings", `{"a":"x", "b":"y"}`, ufOut[map[string]string, map[string]string], ufFromString()),
		c("function of a type of itself", `[[null, []], null]`, ufOut[ufTree, ufTree],
			ufDecliningValues(func(v *ufTree) { *v = ufTree{nil} })),
		c("function of a struct of itself", `{"N":{"S":[null],"V":2},"S":[{"V":3},null],"V":1}`, ufOut[ufNode, ufNode],
			ufDecliningValues(func(v *ufNode) { v.V = -1 })),
		c("function of strings of a struct field", `{"A":"a"}`, ufOut[struct{ A string }, struct{ A string }], ufFromString()),
	}
}

func TestUnmarshalFromMatchesStd(t *testing.T) {
	for _, c := range ufCases() {
		t.Run(c.name, func(t *testing.T) {
			ours, theirs := c.ours(), c.theirs()
			err := json.Unmarshal([]byte(c.in), ours, c.opts.ours...)
			stdErr := stdjson.Unmarshal([]byte(c.in), theirs, c.opts.std...)
			want := strings.ReplaceAll(describeUnmarshalError(stdErr), "ufStd", "ufOurs")
			if got := describeUnmarshalError(err); got != want {
				t.Fatalf("error:\n got %s\nwant %s", got, want)
			}
			if got, want := ufText(ours), ufText(theirs); err == nil && got != want {
				t.Fatalf("value:\n got %s\nwant %s", got, want)
			}
		})
	}
}

// ufText returns the text of the value at p, whose type is the one of either package, of the same fields: its JSON
// text, which has no pointers, or its print of a map of struct keys, which has no pointers either.
func ufText(p any) string {
	if b, err := stdjson.Marshal(p, stdjson.Deterministic(true)); err == nil {
		return string(b)
	}
	return fmt.Sprintf("%+v", reflect.ValueOf(p).Elem())
}
