//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The tests here compare Unmarshal with the one of encoding/json/v2: the same input into the same value with the
// same options must give the same value, or an error of the same type at the same place. The values after an
// error are not compared: the decoding stops where encoding/json/v2 does, but what it leaves in the parts of the
// value it decoded may differ.

type unmarshalDiffByte byte

type unmarshalDiffEmpty interface{}

// unmarshalDiffCase is an input, decoded into a new value of the type which out makes, set to its initial value.
type unmarshalDiffCase struct {
	name string
	in   string
	out  func() any
	opts unmarshalDiffOptions
	// errorOnly is set for a value which is not compared, as a map of pointer keys, whose keys are new pointers.
	errorOnly bool
}

// unmarshalDiffOptions are the options of a case, which are given to both packages.
type unmarshalDiffOptions struct {
	stringifyNumbers, allowInvalidUTF8, allowDuplicateNames bool
	matchCaseInsensitiveNames, rejectUnknownMembers         bool
}

func (o unmarshalDiffOptions) ours() []json.Options {
	return []json.Options{
		json.StringifyNumbers(o.stringifyNumbers),
		jsontext.AllowInvalidUTF8(o.allowInvalidUTF8),
		jsontext.AllowDuplicateNames(o.allowDuplicateNames),
		json.MatchCaseInsensitiveNames(o.matchCaseInsensitiveNames),
		json.RejectUnknownMembers(o.rejectUnknownMembers),
	}
}

func (o unmarshalDiffOptions) std() []stdjson.Options {
	return []stdjson.Options{
		stdjson.StringifyNumbers(o.stringifyNumbers),
		stdjsontext.AllowInvalidUTF8(o.allowInvalidUTF8),
		stdjsontext.AllowDuplicateNames(o.allowDuplicateNames),
		stdjson.MatchCaseInsensitiveNames(o.matchCaseInsensitiveNames),
		stdjson.RejectUnknownMembers(o.rejectUnknownMembers),
	}
}

// outOf returns the function which makes a pointer to a new value of T set to v. v must refer to no memory, which
// the values of the two packages would share: a value which does, a map, a slice or a pointer which is not nil, is
// made by a function of its own.
func outOf[T any](v T) func() any {
	return func() any {
		p := new(T)
		*p = v
		return p
	}
}

var (
	stringify    = unmarshalDiffOptions{stringifyNumbers: true}
	invalidUTF8  = unmarshalDiffOptions{allowInvalidUTF8: true}
	duplicates   = unmarshalDiffOptions{allowDuplicateNames: true}
	deepArray    = strings.Repeat("[", 10001) + strings.Repeat("]", 10001)
	deepArrayMax = strings.Repeat("[", 10000) + strings.Repeat("]", 10000)
)

func unmarshalDiffCases() []unmarshalDiffCase {
	return []unmarshalDiffCase{
		{name: "bool", in: `true`, out: outOf(false)},
		{name: "bool null", in: `null`, out: outOf(true)},
		{name: "bool string", in: `"true"`, out: outOf(true)},
		{name: "bool number after space", in: ` 1`, out: outOf(true)},
		{name: "bool array", in: `[true]`, out: outOf(true)},
		{name: "bool array of a syntax error", in: `[tx]`, out: outOf(true)},
		{name: "bool literal cut", in: `tru`, out: outOf(true)},
		{name: "bool string of invalid UTF-8", in: "\"a\xff\"", out: outOf(true)},
		{name: "int", in: `12`, out: outOf(0)},
		{name: "int negative zero", in: `-0`, out: outOf(7)},
		{name: "int null", in: `null`, out: outOf(7)},
		{name: "int fraction", in: `1.0`, out: outOf(7)},
		{name: "int exponent", in: `1e2`, out: outOf(7)},
		{name: "int fraction after space", in: "\n 1.5", out: outOf(7)},
		{name: "int overflow", in: `99999999999999999999`, out: outOf(7)},
		{name: "int8 overflow", in: `128`, out: outOf(int8(7))},
		{name: "int16 underflow", in: `-32769`, out: outOf(int16(7))},
		{name: "uint negative", in: `-1`, out: outOf(uint(7))},
		{name: "uint8 negative fraction", in: `-1.5`, out: outOf(uint8(7))},
		{name: "uintptr", in: `12`, out: outOf(uintptr(0))},
		{name: "int string", in: `"1"`, out: outOf(7)},
		{name: "int object", in: `{"a":1}`, out: outOf(7)},
		{name: "int array cut", in: `[1,2`, out: outOf(7)},
		{name: "int array of a syntax error", in: `[1,x]`, out: outOf(7)},
		{name: "int leading zero", in: `01`, out: outOf(7)},
		{name: "int stringified", in: `"12"`, out: outOf(7), opts: stringify},
		{name: "int stringified negative", in: `"-12"`, out: outOf(int8(7)), opts: stringify},
		{name: "int stringified escaped", in: `"\u0031"`, out: outOf(7), opts: stringify},
		{name: "int stringified plain", in: `12`, out: outOf(7), opts: stringify},
		{name: "int stringified space", in: `" 12"`, out: outOf(7), opts: stringify},
		{name: "int stringified null", in: `null`, out: outOf(7), opts: stringify},
		{name: "int stringified word", in: `"x"`, out: outOf(7), opts: stringify},
		{name: "int stringified fraction", in: `"1.5"`, out: outOf(7), opts: stringify},
		{name: "int8 stringified overflow", in: `"300"`, out: outOf(int8(7)), opts: stringify},
		{name: "uint stringified negative", in: `"-1"`, out: outOf(uint(7)), opts: stringify},
		{name: "float", in: `1.5`, out: outOf(0.0)},
		{name: "float negative zero", in: `-0`, out: outOf(1.0)},
		{name: "float overflow", in: `1e400`, out: outOf(0.0)},
		{name: "float32 overflow", in: `1e40`, out: outOf(float32(0))},
		{name: "float null", in: `null`, out: outOf(2.5)},
		{name: "float stringified", in: `"1.5e3"`, out: outOf(0.0), opts: stringify},
		{name: "float32 stringified overflow", in: `"1e40"`, out: outOf(float32(0)), opts: stringify},
		{name: "float stringified NaN", in: `"NaN"`, out: outOf(0.0), opts: stringify},
		{name: "string", in: `"abc"`, out: outOf("")},
		{name: "string of escapes", in: `"\u00e9\n\ud83d\ude00"`, out: outOf("")},
		{name: "string null", in: `null`, out: outOf("x")},
		{name: "string of invalid UTF-8", in: "\"a\xffb\"", out: outOf("x")},
		{name: "string of invalid UTF-8 allowed", in: "\"a\xffb\"", out: outOf("x"), opts: invalidUTF8},
		{name: "string of a lone surrogate", in: `"a\ud800b"`, out: outOf("x")},
		{name: "string of a lone low surrogate", in: `"a\udc00"`, out: outOf("x")},
		{name: "string of a high surrogate before a letter", in: `"\ud83dA"`, out: outOf("x")},
		{name: "string of a lone surrogate allowed", in: `"a\ud800b"`, out: outOf("x"), opts: invalidUTF8},
		{name: "string of an escaped replacement character", in: `"\ufffd"`, out: outOf("")},
		{name: "string of a replacement character", in: "\"\xef\xbf\xbd\"", out: outOf("")},
		{name: "string of an escaped backslash before u", in: `"\\ud800"`, out: outOf("")},
		{name: "string number", in: `1`, out: outOf("x")},
		{name: "string of a bad escape", in: `"\x"`, out: outOf("x")},
		{name: "string of a control byte", in: "\"a\tb\"", out: outOf("x")},
		{name: "string of simple escapes", in: `"a\"b\\c\/d\b\f\n\r\t"`, out: outOf("")},
		{name: "string of an escape and a byte which is not ASCII", in: "\"é\\n\"", out: outOf("")},
		{name: "string of an escape and invalid UTF-8", in: "\"\\n\xff\"", out: outOf("x")},
		{name: "string of an escape and invalid UTF-8 allowed", in: "\"\\n\xff\"", out: outOf("x"), opts: invalidUTF8},
		{name: "string of an escape and a replacement character", in: "\"\\n\xef\xbf\xbd\"", out: outOf("")},
		{name: "string of an escape and a control byte", in: "\"\\n\x01\"", out: outOf("x")},
		{name: "string of an escape and a lone surrogate", in: `"\n\ud800"`, out: outOf("x")},
		{name: "string of an escape and a surrogate pair", in: `"\n\ud83d\ude00"`, out: outOf("")},
		{name: "string of a lone surrogate at the end", in: `"a\ud800"`, out: outOf("x")},
		{name: "string of a high surrogate before a bad escape", in: `"\ud800\uZZZZ"`, out: outOf("x")},
		{name: "string of a surrogate pair in upper case", in: `"\uD83D\uDE00"`, out: outOf("")},
		{name: "string cut after a backslash", in: `"abc\`, out: outOf("x")},
		{name: "string cut after an escape", in: `"\n`, out: outOf("x")},
		{name: "string long of invalid UTF-8 at the end", in: "\"" + strings.Repeat("a", 100) + "\xff\"", out: outOf("x")},
		{name: "string long of an escape and invalid UTF-8 at the end", in: "\"" + strings.Repeat("a", 100) + "\\n\xff\"", out: outOf("x")},
		{name: "string long of a lone surrogate at the end", in: `"` + strings.Repeat("a", 100) + `\ud800"`, out: outOf("x")},
		{name: "string long of a byte which is not ASCII at the end", in: "\"" + strings.Repeat("a", 100) + "é\"", out: outOf("")},
		{name: "string longer than the arena of escapes", in: `"` + strings.Repeat(`a\n`, 300) + `"`, out: outOf("")},
		{name: "string longer than the arena of escapes and invalid UTF-8", in: "\"" + strings.Repeat("a\\n", 300) + "\xff\"", out: outOf("x")},
		{name: "strings of an escape and invalid UTF-8 allowed", in: "[\"\\n\xff\",\"ok\"]", out: outOf([]string(nil)), opts: invalidUTF8},
		{name: "named string", in: `"abc"`, out: outOf(unmarshalDiffText(""))},
		{name: "bytes", in: `"AQID"`, out: outOf([]byte(nil))},
		{name: "bytes null", in: `null`, out: func() any { v := []byte{1}; return &v }},
		{name: "bytes empty", in: `""`, out: outOf([]byte(nil))},
		{name: "bytes reused", in: `"AQ=="`, out: func() any { b := make([]byte, 1, 8); b[0] = 9; return &b }},
		{name: "bytes of a line feed", in: `"AQ\nID"`, out: outOf([]byte(nil))},
		{name: "bytes without padding", in: `"AQI"`, out: outOf([]byte(nil))},
		{name: "bytes of the URL alphabet", in: `"-_8="`, out: outOf([]byte(nil))},
		{name: "bytes array", in: `[1,2]`, out: outOf([]byte(nil))},
		{name: "bytes not base64", in: `"!!!!"`, out: func() any { v := []byte{5}; return &v }},
		{name: "bytes of invalid UTF-8", in: "\"AQ\xff\"", out: outOf([]byte(nil))},
		{name: "byte array", in: `"AQID"`, out: outOf([3]byte{})},
		{name: "byte array short", in: `"AQI="`, out: outOf([3]byte{9, 9, 9})},
		{name: "byte array null", in: `null`, out: outOf([3]byte{9, 9, 9})},
		{name: "named bytes", in: `[1,2]`, out: outOf([]unmarshalDiffByte(nil))},
		{name: "named bytes string", in: `"AQID"`, out: outOf([]unmarshalDiffByte(nil))},
		{name: "named byte array", in: `[1,2]`, out: outOf([2]unmarshalDiffByte{})},
		{name: "any null", in: `null`, out: outOf[any](5)},
		{name: "any values", in: `[true,1,"s",null,{"a":[]},-1.5e3]`, out: outOf[any](nil)},
		{name: "any duplicate name", in: `{"a":1,"a":2}`, out: outOf[any](nil)},
		{name: "any empty name of a syntax error", in: `{"":1e}`, out: outOf[any](nil)},
		{name: "any empty name of an object of a syntax error", in: `{"":{"x":1e}}`, out: outOf[any](nil)},
		{name: "map empty name of a syntax error", in: `{"":1e}`, out: outOf(map[string]any(nil))},
		{name: "any nested duplicate name", in: `[{"b":{"a":1,"a":2}}]`, out: outOf[any](nil)},
		{name: "any duplicate name allowed", in: `{"a":1,"a":2}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate objects allowed", in: `{"a":{"x":1},"a":{"y":2}}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate nested objects allowed", in: `{"a":{"x":{"p":1}},"a":{"x":{"q":2}}}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate arrays allowed", in: `{"a":[1],"a":[2,3]}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate after null allowed", in: `{"a":null,"a":1}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate null allowed", in: `{"a":1,"a":null}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate of another kind allowed", in: `{"a":1,"a":"x"}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate object after a string allowed", in: `{"a":"s","a":{"b":1}}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate in an array allowed", in: `[{"a":[{"x":1}],"a":[{"y":2}]}]`, out: outOf[any](nil), opts: duplicates},
		{name: "any int", in: `6`, out: outOf[any](5)},
		{name: "any int string", in: `"s"`, out: outOf[any](5)},
		{name: "any pointer", in: `6`, out: func() any { n := 5; var v any = &n; return &v }},
		{name: "any map merged", in: `{"b":2,"a":{"y":3}}`, out: func() any { var v any = map[string]any{"a": map[string]any{"x": 1.0}}; return &v }},
		{name: "any map of a duplicate name", in: `{"b":2,"b":3}`, out: func() any { var v any = map[string]any{"a": 1.0}; return &v }},
		{name: "any map of a duplicate name allowed", in: `{"b":2,"b":3}`, out: func() any { var v any = map[string]any{"a": 1.0}; return &v }, opts: duplicates},
		{name: "any map array", in: `[1]`, out: func() any { var v any = map[string]any{"a": 1.0}; return &v }},
		{name: "any nil map", in: `{"a":1}`, out: func() any { var v any = map[string]any(nil); return &v }},
		{name: "any slice reused", in: `[3]`, out: func() any { var v any = []any{1.0, 2.0}; return &v }},
		{name: "any slice object", in: `{}`, out: func() any { var v any = []any{1.0}; return &v }},
		{name: "any float object", in: `{}`, out: outOf[any](1.5)},
		{name: "any overflow", in: `1e400`, out: outOf[any](nil)},
		{name: "any nested overflow", in: `[1,{"x":[2,1e400]}]`, out: outOf[any](nil)},
		{name: "any name of invalid UTF-8", in: "{\"a\xff\":1}", out: outOf[any](nil)},
		{name: "any name of a lone surrogate", in: `[{"a\udc00":1}]`, out: outOf[any](nil)},
		{name: "any string of invalid UTF-8 allowed", in: "[\"a\xff\"]", out: outOf[any](nil), opts: invalidUTF8},
		{name: "any number in a string", in: `"1"`, out: outOf[any](nil), opts: stringify},
		{name: "named empty interface", in: `[1]`, out: outOf[unmarshalDiffEmpty](nil)},
		{name: "reader nil", in: `{}`, out: outOf[io.Reader](nil)},
		{name: "slice", in: `[1,2]`, out: func() any { v := []int{9, 9, 9}; return &v }},
		{name: "slice null", in: `null`, out: func() any { v := []int{9}; return &v }},
		{name: "slice empty", in: `[]`, out: outOf([]int(nil))},
		{name: "slice of a syntax error after a comma", in: `[1,]`, out: outOf([]int(nil))},
		{name: "slice of an element of another type", in: `[1,"x",3]`, out: func() any { v := []int{9, 9, 9}; return &v }},
		{name: "slice of pointers reused", in: `[2]`, out: func() any { n := 1; s := []*int{&n}; return &s }},
		{name: "slice of any reused", in: `["s"]`, out: func() any { v := []any{5}; return &v }},
		{name: "slice of a type error then a syntax error", in: `["x", }`, out: outOf([]int(nil))},
		{name: "slice of slices", in: `[[1],[],null,[2,3]]`, out: func() any { v := [][]int{{9, 9}}; return &v }},
		{name: "array", in: `[1,2]`, out: outOf([2]int{5, 5})},
		{name: "array short", in: `[1]`, out: outOf([2]int{5, 5})},
		{name: "array long", in: `[1,2,3]`, out: outOf([2]int{5, 5})},
		{name: "array long of invalid UTF-8", in: "[1,2,\"\xff\"]", out: outOf([2]int{5, 5})},
		{name: "array null", in: `null`, out: outOf([2]int{5, 5})},
		{name: "array of pointers reused", in: `[3]`, out: func() any { n := 1; a := [1]*int{&n}; return &a }},
		{name: "array nested short", in: `[[1]]`, out: outOf([][2]int(nil))},
		{name: "pointer null", in: `null`, out: func() any { n := 5; v := &n; return &v }},
		{name: "pointer allocated", in: `3`, out: outOf[*int](nil)},
		{name: "pointer to a pointer", in: `3`, out: outOf[**int](nil)},
		{name: "pointer reused", in: `4`, out: func() any { n := 1; p := &n; return &p }},
		{name: "chan", in: `1`, out: outOf[chan int](nil)},
		{name: "chan null", in: ` null`, out: outOf[chan int](nil)},
		{name: "func", in: `null`, out: outOf[func()](nil)},
		{name: "complex", in: `1`, out: outOf(complex128(0))},
		{name: "complex in a slice", in: `[1]`, out: outOf([]complex128(nil))},
		{name: "complex in a nested slice", in: `[[1,2],[3]]`, out: outOf([][]complex128(nil))},
		{name: "empty", in: ``, out: outOf(0)},
		{name: "white space", in: "  \n", out: outOf(0)},
		{name: "value after the value", in: `1 2`, out: outOf(0)},
		{name: "character after the value", in: `1 x`, out: outOf(0)},
		{name: "deep", in: deepArray, out: outOf[any](nil)},
		{name: "deepest", in: deepArrayMax, out: outOf[any](nil)},
		{name: "deep slices", in: deepArray, out: outOf(unmarshalDiffSlices(nil))},
		{name: "pointer of escaped names", in: `[[1,{"a/b~":[0,"x"]}]]`, out: outOf([][]any(nil))},
		{name: "pointer of escaped names of an error", in: `[{"a/b~":[0,1e400]}]`, out: outOf[any](nil)},
		{name: "any overflow before a syntax error", in: `[1e400, x]`, out: outOf[any](nil)},
		{name: "any overflow before a duplicate name", in: `[1e400, {"a":1,"a":2}]`, out: outOf[any](nil)},
		{name: "any overflow before a duplicate name allowed", in: `[1e400, {"a":1,"a":2}]`, out: outOf[any](nil), opts: duplicates},
		{name: "any overflow before a bad escape", in: `[1e400, "\x"]`, out: outOf[any](nil)},
		{name: "any overflow before the end", in: `[1e400, 1`, out: outOf[any](nil)},
		{name: "any overflow after a duplicate name allowed", in: `{"b":1e400,"a":1,"a":1}`, out: outOf[any](nil), opts: duplicates},
		{name: "empty byte array", in: `""`, out: outOf([0]byte{})},
		{name: "empty byte array array", in: `[]`, out: outOf([0]byte{})},
		{name: "empty byte array of a byte", in: `"AA=="`, out: outOf([0]byte{})},
		{name: "recursive slice", in: `[[],[[]],null]`, out: outOf(unmarshalDiffSlices(nil))},
		{name: "recursive slice of another type", in: `[[1]]`, out: outOf(unmarshalDiffSlices(nil))},
		{name: "recursive array", in: `[[null]]`, out: outOf(unmarshalDiffArray{})},
		{name: "recursive array short", in: `[[]]`, out: outOf(unmarshalDiffArray{})},
		{name: "deep recursive array", in: deepArray, out: outOf(unmarshalDiffArray{})},
		{name: "interface with methods holding a value", in: `"x"`, out: outOf[fmt.Stringer](unmarshalDiffStringer("s"))},
		{name: "interface with methods holding a pointer", in: `"x"`, out: func() any { s := unmarshalDiffStringer("s"); var v fmt.Stringer = &s; return &v }},
		{name: "interface with methods in a slice", in: `[null,{}]`, out: outOf([]io.Reader(nil))},
		{name: "any holding a nil pointer", in: `5`, out: outOf[any]((*int)(nil))},
		{name: "any holding a pointer to a struct", in: `{"y":2}`, out: func() any { var v any = &udPoint{X: 5}; return &v }},
		{name: "any holding a pointer to a struct of null", in: `null`, out: func() any { var v any = &udPoint{X: 5}; return &v }},
		{name: "any holding a pointer of another kind", in: `"x"`, out: func() any { var v any = new(int); return &v }},
		{name: "any holding a pointer to a pointer", in: `7`, out: func() any { n := new(int); var v any = &n; return &v }},
		{name: "any holding a pointer to a nil pointer", in: `7`, out: func() any { var n *int; var v any = &n; return &v }},
		{name: "any holding a pointer to a pointer of null", in: `{"a":null}`, out: func() any { n := new(int); m := map[string]any{"a": &n}; return &m }},
		{name: "any holding a named type", in: `"x"`, out: outOf[any](unmarshalDiffText("s"))},
		{name: "any holding an array", in: `[1,2]`, out: outOf[any]([2]int{5, 5})},
		{name: "any holding an array short", in: `[1]`, out: outOf[any]([2]int{5, 5})},
		{name: "any holding a slice", in: `[1,2]`, out: func() any { var v any = []int{5}; return &v }},
		{name: "any holding a typed nil map", in: `{"a":1}`, out: outOf[any](map[string]any(nil))},
		{name: "int64 stringified min", in: `"-9223372036854775808"`, out: outOf(int64(0)), opts: stringify},
		{name: "int64 stringified below min", in: `"-9223372036854775809"`, out: outOf(int64(0)), opts: stringify},
		{name: "uint64 stringified max", in: `"18446744073709551615"`, out: outOf(uint64(0)), opts: stringify},
		{name: "uint64 stringified above max", in: `"18446744073709551616"`, out: outOf(uint64(0)), opts: stringify},
		{name: "uint stringified negative zero", in: `"-0"`, out: outOf(uint(7)), opts: stringify},
		{name: "int stringified negative zero", in: `"-0"`, out: outOf(7), opts: stringify},
		{name: "float stringified negative zero", in: `"-0"`, out: outOf(7.0), opts: stringify},
		{name: "float stringified underflow", in: `"1e-400"`, out: outOf(7.0), opts: stringify},
		{name: "int stringified leading zero", in: `"01"`, out: outOf(7), opts: stringify},
		{name: "int stringified plus", in: `"+1"`, out: outOf(7), opts: stringify},
		{name: "int stringified empty", in: `""`, out: outOf(7), opts: stringify},
		{name: "int stringified space after", in: `"1 "`, out: outOf(7), opts: stringify},
		{name: "any stringified number", in: `1`, out: outOf[any](nil), opts: stringify},
		{name: "ints stringified", in: `["1",null,"-2"]`, out: outOf([]int(nil)), opts: stringify},
		{name: "ints stringified plain", in: `[1]`, out: outOf([]int(nil)), opts: stringify},
		{name: "string of an overlong encoding", in: "\"\xc0\xaf\"", out: outOf("")},
		{name: "string of an encoded surrogate", in: "\"\xed\xa0\x80\"", out: outOf("")},
		{name: "string of a cut sequence", in: "\"\xe3\x81\"", out: outOf("")},
		{name: "string of a byte F5", in: "\"\xf5\x80\x80\x80\"", out: outOf("")},
		{name: "string of an encoded surrogate allowed", in: "\"\xed\xa0\x80\"", out: outOf(""), opts: invalidUTF8},
		{name: "any names equal after replacement", in: "{\"\xff\":1,\"\xfe\":2}", out: outOf[any](nil), opts: invalidUTF8},
		{name: "any names equal after replacement allowed", in: "{\"\xff\":1,\"\xfe\":2}", out: outOf[any](nil), opts: unmarshalDiffOptions{allowInvalidUTF8: true, allowDuplicateNames: true}},
		{name: "deepest slice of any", in: deepArrayMax, out: outOf([]any(nil))},
		{name: "deep slice of any", in: deepArray, out: outOf([]any(nil))},
		{name: "deep slice of slices of any", in: deepArray, out: outOf([][]any(nil))},
		{name: "deep array of any", in: deepArray, out: outOf([1]any{})},
		{name: "any duplicate of another kind before an overflow allowed", in: `{"a":{"x":1},"a":5,"b":1e400}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate overflow into a string allowed", in: `{"a":"s","a":1e400}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate overflow into a bool allowed", in: `{"a":true,"a":1e400}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate overflow into an object allowed", in: `{"a":{},"a":1e400}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate object into an array before an overflow allowed", in: `{"a":[1],"a":{},"b":1e400}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate after a name before an overflow allowed", in: `{"q":1,"a":{"x":1},"a":5,"b":1e400}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate overflow in a slice allowed", in: `[{"a":"s","a":1e400}]`, out: outOf([]any(nil)), opts: duplicates},
		{name: "any duplicate nested overflow allowed", in: `{"a":1,"a":[[1e400]]}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate before a syntax error allowed", in: `{"a":[1],"a":{"x":}}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate before the end allowed", in: `{"a":{},"a":[1,`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate before a nested syntax error allowed", in: `{"a":1,"a":[[1,]]}`, out: outOf[any](nil), opts: duplicates},
		{name: "any duplicate overflow without the option", in: `{"a":"s","a":1e400}`, out: outOf[any](nil)},
		{name: "any duplicates before an overflow", in: `{"a":1,"a":2,"b":1e400}`, out: outOf[any](nil)},
	}
}

type unmarshalDiffStringer string

func (s unmarshalDiffStringer) String() string { return string(s) }

type unmarshalDiffArray [1]*unmarshalDiffArray

type unmarshalDiffText string

type unmarshalDiffSlices []unmarshalDiffSlices

func TestUnmarshalMatchesStd(t *testing.T) {
	for _, c := range append(unmarshalDiffCases(), unmarshalDiffStructCases()...) {
		t.Run(c.name, func(t *testing.T) {
			ours, theirs := c.out(), c.out()
			err := json.Unmarshal([]byte(c.in), ours, c.opts.ours()...)
			stdErr := stdjson.Unmarshal([]byte(c.in), theirs, c.opts.std()...)
			if got, want := describeUnmarshalError(err), describeUnmarshalError(stdErr); got != want {
				t.Fatalf("error:\n got %s\nwant %s", got, want)
			}
			if err == nil && !c.errorOnly && !reflect.DeepEqual(ours, theirs) {
				t.Fatalf("value:\n got %#v\nwant %#v", reflect.ValueOf(ours).Elem(), reflect.ValueOf(theirs).Elem())
			}
		})
	}
}

// udCalled counts the calls of its method in udCalls.
type udCalled struct{}

var udCalls int

func (*udCalled) UnmarshalJSON([]byte) error {
	udCalls++
	return nil
}

// TestUnmarshalStopsAtANumberError checks that a number which is not a value of its type stops the decoding, as
// encoding/json/v2 stops at its first error: the member after it is not decoded.
func TestUnmarshalStopsAtANumberError(t *testing.T) {
	type numbers struct {
		I8  int8
		U   uint
		F32 float32
		I   int
		M   udCalled
	}
	callsOf := func(unmarshal func(v any) error) (int, error) {
		udCalls = 0
		err := unmarshal(new(numbers))
		return udCalls, err
	}
	for _, in := range []string{`{"I8":300,"M":1}`, `{"U":-1,"M":1}`, `{"F32":1e40,"M":1}`, `{"I":1.5,"M":1}`} {
		want, stdErr := callsOf(func(v any) error { return stdjson.Unmarshal([]byte(in), v) })
		if stdErr == nil {
			t.Fatalf("%s: no error of encoding/json/v2", in)
		}
		got, err := callsOf(func(v any) error { return json.Unmarshal([]byte(in), v) })
		if err == nil || got != want {
			t.Fatalf("%s: %d calls, %v; want %d calls and an error", in, got, err, want)
		}
		for _, n := range readSizes {
			got, err := callsOf(func(v any) error { return json.UnmarshalRead(newChunkReader(in, n), v) })
			if err == nil || got != want {
				t.Fatalf("%s, reads of %d: %d calls, %v; want %d calls and an error", in, n, got, err, want)
			}
		}
	}
}

// TestUnmarshalLeavesWhatElementsPointTo checks that an element of a slice or an array which a pointer is decoded
// into is a new pointer, as encoding/json/v2 decodes into the zero value of an element: what the old one pointed to
// is left as it was.
func TestUnmarshalLeavesWhatElementsPointTo(t *testing.T) {
	for _, c := range []struct {
		name   string
		decode func(t *testing.T, unmarshal func([]byte, any) error) int
	}{
		{"slice", func(t *testing.T, unmarshal func([]byte, any) error) int {
			n := 1
			s := []*int{&n}
			if err := unmarshal([]byte(`[2]`), &s); err != nil {
				t.Fatal(err)
			}
			return n
		}},
		{"slice in an interface", func(t *testing.T, unmarshal func([]byte, any) error) int {
			n := 1
			var v any = []*int{&n}
			if err := unmarshal([]byte(`[2]`), &v); err != nil {
				t.Fatal(err)
			}
			return n
		}},
		{"array", func(t *testing.T, unmarshal func([]byte, any) error) int {
			n := 1
			a := [1]*int{&n}
			if err := unmarshal([]byte(`[2]`), &a); err != nil {
				t.Fatal(err)
			}
			return n
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := c.decode(t, func(in []byte, out any) error { return json.Unmarshal(in, out) })
			want := c.decode(t, func(in []byte, out any) error { return stdjson.Unmarshal(in, out) })
			if got != want {
				t.Fatalf("the value the element pointed to is %d, want %d", got, want)
			}
		})
	}
}

// describeUnmarshalError describes an error of either package by its type and its fields.
func describeUnmarshalError(err error) string {
	var (
		se     *json.SemanticError
		ye     *jsontext.SyntacticError
		stdSe  *stdjson.SemanticError
		stdYe  *stdjsontext.SyntacticError
		errMsg = func(err error) string {
			if err == nil {
				return "<nil>"
			}
			return err.Error()
		}
	)
	switch {
	case err == nil:
		return "nil"
	case errors.As(err, &se):
		return fmt.Sprintf("SemanticError{ByteOffset: %d, JSONPointer: %q, JSONKind: %v, JSONValue: %q, GoType: %v, Err: %s}",
			se.ByteOffset, se.JSONPointer, se.JSONKind, se.JSONValue, se.GoType, errMsg(se.Err))
	case errors.As(err, &stdSe):
		return fmt.Sprintf("SemanticError{ByteOffset: %d, JSONPointer: %q, JSONKind: %v, JSONValue: %q, GoType: %v, Err: %s}",
			stdSe.ByteOffset, stdSe.JSONPointer, stdSe.JSONKind, stdSe.JSONValue, stdSe.GoType, errMsg(stdSe.Err))
	case errors.As(err, &ye):
		return fmt.Sprintf("SyntacticError{ByteOffset: %d, JSONPointer: %q, Err: %s}", ye.ByteOffset, ye.JSONPointer, errMsg(ye.Err))
	case errors.As(err, &stdYe):
		return fmt.Sprintf("SyntacticError{ByteOffset: %d, JSONPointer: %q, Err: %s}", stdYe.ByteOffset, stdYe.JSONPointer, errMsg(stdYe.Err))
	}
	return fmt.Sprintf("%T: %v", err, err)
}
