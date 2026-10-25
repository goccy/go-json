//go:build go1.27 && goexperiment.jsonv2

package json_test

import "strings"

// The values which the decoding skips, as the ones of the members of no field of a struct, are compared with
// encoding/json/v2 under each mix of AllowDuplicateNames and AllowInvalidUTF8: their grammar, the text which v2
// allows in their strings and names, and the names twice of the objects in them, at every depth, in the value of a
// member and in an array.

type skippedValueStruct struct{ A, B int }

// skippedValue is a value which the decoding skips, by its name: surrogate is set for a value of an escape of a
// surrogate which is not a pair, whose error encoding/json/v2 describes by the bytes it has read ( see
// readTextByReads ).
type skippedValue struct {
	name, value string
	surrogate   bool
}

func skippedValues() []skippedValue {
	names := make([]string, 40)
	for i := range names {
		names[i] = `"n` + strings.Repeat("x", i) + `":1`
	}
	many := strings.Join(names, ",")
	long := strings.Repeat("abcdefgh", 10)
	deep := strings.Repeat("[", 70)
	deepEnd := strings.Repeat("]", 70)
	return []skippedValue{
		{name: "number", value: `-1.5e3`},
		{name: "literals", value: `[true,false,null]`},
		{name: "string", value: `"abc"`},
		{name: "long string", value: `"` + long + `"`},
		{name: "string of escapes", value: `"a\n\u0041\"\\"`},
		{name: "string of a pair of surrogates", value: `"\ud83d\ude00"`},
		{name: "string of a lone surrogate", value: `"\ud800"`, surrogate: true},
		{name: "string of a lone low surrogate", value: `"\udc00x"`, surrogate: true},
		{name: "long string of a lone surrogate", value: `"` + long + `\ud800"`, surrogate: true},
		{name: "string of UTF-8", value: "\"éあ\""},
		{name: "long string of UTF-8", value: "\"" + long + "é\""},
		{name: "string of invalid UTF-8", value: "\"a\xff\""},
		{name: "long string of invalid UTF-8", value: "\"" + long + "\xffx\""},
		{name: "string of a cut rune", value: "\"\xe3\x81\""},
		{name: "empty object", value: `{}`},
		{name: "empty array", value: `[]`},
		{name: "object", value: `{"a":1,"b":[1,2],"c":{"d":"e"}}`},
		{name: "object of a name twice", value: `{"a":1,"a":2}`},
		{name: "object of a name twice after a space", value: `{ "a" : 1 , "a" : 2 }`},
		{name: "object of a name twice by an escape", value: `{"a":1,"\u0061":2}`},
		{name: "object of a name twice by an escape first", value: `{"\u0061":1,"a":2}`},
		{name: "object of a name twice of escapes", value: `{"a\"":1,"a\u0022":2}`},
		{name: "object of names of other escapes", value: `{"a\n":1,"a\t":2,"a":3}`},
		{name: "object of a long name twice", value: `{"` + long + `":1,"` + long + `":2}`},
		{name: "object of long names of other middles", value: `{"` + long + `":1,"` + long[:30] + `X` + long[31:] + `":2}`},
		{name: "object of a name of UTF-8 twice", value: "{\"é\":1,\"é\":2}"},
		{name: "object of a name of UTF-8 twice by an escape", value: "{\"é\":1,\"\u00e9\":2}"},
		{name: "object of a name of invalid UTF-8", value: "{\"\xff\":1}"},
		{name: "object of a name of invalid UTF-8 twice", value: "{\"\xff\":1,\"\xff\":2}"},
		{name: "object of names of other invalid UTF-8", value: "{\"\xff\":1,\"\xfe\":2}"},
		{name: "object of a name of invalid UTF-8 and its replacement", value: "{\"\xff\":1,\"�\":2}"},
		{name: "object of names of a cut rune and its replacements", value: "{\"\xe3\x81\":1,\"\ufffd\ufffd\":2}"},
		{name: "object of names of a cut rune and one replacement", value: "{\"\xe3\x81\":1,\"\ufffd\":2}"},
		{name: "object of a long name of invalid UTF-8 twice", value: "{\"" + long + "\xff\":1,\"" + long + "\xfe\":2}"},
		{name: "object of a name of a lone surrogate", value: `{"\ud800":1}`, surrogate: true},
		{name: "object of the empty name twice", value: `{"":1,"":2}`},
		{name: "object of many names", value: `{` + many + `}`},
		{name: "object of many names and one twice", value: `{` + many + `,"nxxx":2}`},
		{name: "objects of the same names", value: `[{"a":1,"b":2},{"a":1,"b":2}]`},
		{name: "object in an object of the same names", value: `{"a":{"a":1,"b":{"a":1}},"b":2}`},
		{name: "object of a name twice in an object", value: `{"a":1,"b":{"c":1,"c":2}}`},
		{name: "object of a name twice after an object", value: `{"a":{"b":1},"a":2}`},
		{name: "object of a name twice after an object of names", value: `{"a":{"x":1,"y":2,"z":3},"b":1,"a":2}`},
		{name: "deep value", value: deep + deepEnd},
		{name: "deep value of an object", value: deep + `{"a":1,"b":2}` + deepEnd},
		{name: "deep value of a name twice", value: deep + `{"a":1,"a":2}` + deepEnd},
		{name: "object of a deep value and a name twice", value: `{"a":` + deep + deepEnd + `,"a":2}`},
		{name: "object of a deep value of invalid UTF-8", value: "{\"a\":" + deep + "\"\xff\"" + deepEnd + "}"},
		{name: "object of a missing colon", value: `{"a" 1}`},
		{name: "object of a trailing comma", value: `{"a":1,}`},
		{name: "array of a trailing comma", value: `[1,]`},
		{name: "object of a missing comma", value: `{"a":1 "b":2}`},
		{name: "object of a name twice before a syntax error", value: `{"a":1,"a":2,"b":}`},
		{name: "object of a syntax error before a name twice", value: `{"a":tru,"a":2}`},
		{name: "object of a bad number", value: `{"a":01}`},
		{name: "object of a cut fraction", value: `{"a":1.}`},
		{name: "string of a bad escape", value: `"a\x"`},
		{name: "string of a control character", value: "\"a\x01\""},
	}
}

// skippedOptions are the mixes of the options of the text which v2 allows, by their names.
var skippedOptions = []struct {
	name string
	opts unmarshalDiffOptions
}{
	{"", unmarshalDiffOptions{}},
	{" allowing names twice", unmarshalDiffOptions{allowDuplicateNames: true}},
	{" allowing invalid UTF-8", unmarshalDiffOptions{allowInvalidUTF8: true}},
	{" allowing both", unmarshalDiffOptions{allowDuplicateNames: true, allowInvalidUTF8: true}},
}

// skippedValueCases are the cases of the skipped values, as the value of a member of no field and in an array.
func skippedValueCases() []unmarshalDiffCase {
	var cases []unmarshalDiffCase
	for _, v := range skippedValues() {
		for _, o := range skippedOptions {
			for _, at := range []struct{ name, in string }{
				{"", `{"A":1,"x":` + v.value + `,"B":2}`},
				{" in an array", `{"A":1,"x":[1,` + v.value + `],"B":2}`},
			} {
				name := "struct skipped " + v.name + at.name + o.name
				cases = append(cases, unmarshalDiffCase{name: name, in: at.in, out: outOf(skippedValueStruct{}), opts: o.opts})
			}
		}
	}
	return cases
}

func init() {
	for _, c := range skippedValueCasesOfSurrogates() {
		readTextByReads[c] = true
	}
}

// skippedValueCasesOfSurrogates are the names of the cases of the skipped values of a surrogate which is not a pair.
func skippedValueCasesOfSurrogates() []string {
	var names []string
	for _, v := range skippedValues() {
		if !v.surrogate {
			continue
		}
		for _, o := range skippedOptions {
			names = append(names, "struct skipped "+v.name+o.name, "struct skipped "+v.name+" in an array"+o.name)
		}
	}
	return names
}
