//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The cases of structs and maps of TestUnmarshalMatchesStd.

type udPoint struct {
	X int
	Y int `json:"y"`
}

type udCased struct {
	Name   string
	Ignore string `json:"IGNORE_me,case:ignore"`
	Strict string `json:"strict,case:strict"`
}

type udFolded struct {
	FooBar int
	Foobar int `json:"foo_bar"`
}

type udInner struct {
	A int
	B string
}

type udOuter struct {
	udInner
	C int
}

type udOuterPtr struct {
	*udInner
	C int
}

type udUnexportedEmbed struct{ X int }

type udOuterUnexportedPtr struct {
	*udUnexportedEmbed
	C int
}

type udMapFallback struct {
	A    int
	Rest map[string]any `json:",embed"`
}

type udNamedKey string

type udNamedMapFallback struct {
	A    int
	Rest map[udNamedKey]int `json:",embed"`
}

type udPtrFallback struct {
	A    int
	Rest *map[string]int `json:",embed"`
}

type udInlinedFallback struct {
	udMapFallback
	B int
}

type udStringOption struct {
	I  int            `json:",string"`
	P  *int           `json:",string"`
	F  float64        `json:",string"`
	U  uint8          `json:",string"`
	PP **int          `json:",string"`
	S  []int          `json:",string"`
	A  any            `json:",string"`
	B  bool           `json:",string"`
	T  string         `json:",string"`
	M  map[string]int `json:",string"`
}

type udFormat struct {
	B []byte `json:",format:hex"`
}

type udUnexportedOnly struct {
	a int
}

type udEmpty struct{}

type udNested struct {
	P  udPoint
	PP *udPoint
	L  []udPoint
	M  map[string]udPoint
}

type udRecursive struct {
	Name     string
	Children []udRecursive
	Next     *udRecursive
}

type udOmit struct {
	A int `json:"-"`
	B int `json:"-,"`
	C int `json:"c,omitempty,omitzero"`
}

// udWide is a struct of more fields than the 64 bits of a word.
type udWide struct {
	F0, F1, F2, F3, F4, F5, F6, F7, F8, F9, F10, F11, F12, F13, F14, F15, F16, F17, F18, F19 int
	F20, F21, F22, F23, F24, F25, F26, F27, F28, F29, F30, F31, F32, F33, F34, F35, F36, F37 int
	F38, F39, F40, F41, F42, F43, F44, F45, F46, F47, F48, F49, F50, F51, F52, F53, F54, F55 int
	F56, F57, F58, F59, F60, F61, F62, F63, F64, F65, F66, F67, F68, F69, F70, F71, F72, F73 int
}

// udNames has names which a name is told apart from by its words: of 7, 8 and 16 bytes, two of 19 bytes of the same
// first and last words, a name of a control character and one which ends with a nul, which an input has only
// escaped, a name which is not ASCII, and two names which differ only in case. A name of a backslash can't be: a
// tag can't have it. P nests it, for the names of no field of an object in another.
type udNames struct {
	Seven   int `json:"abcdefg"`
	Eight   int `json:"abcdefgh"`
	Sixteen int `json:"abcdefghijklmnop"`
	Long1   int `json:"abcdefgh_1_ijklmnop"`
	Long2   int `json:"abcdefgh_2_ijklmnop"`
	Control int `json:"a\bb"`
	Nul     int `json:"abc\x00"`
	Accent  int `json:"é"`
	Lower   int `json:"name"`
	Upper   int `json:"Name"`
	P       *udNames
}

// udDeep nests itself by R, for the values at the depth limit.
type udDeep struct {
	R *udDeep
}

// udKeyNames has a map whose keys are pointers, whose names may be the same: they are kept for the names an object
// has twice, as the names of no field are.
type udKeyNames struct {
	K map[*int]int
}

// udSnake has names of '_' and '-', as the structs of many APIs have: names of no field are looked up without their
// '_' and '-' among the names of the fields without theirs ( see foldKey of internal/decoder ). Two names of the same
// folded name make a name of it ambiguous; one of them is `case:strict`.
type udSnake struct {
	UserName  int `json:"user_name"`
	CreatedAt int `json:"created-at"`
	ID        int `json:"id"`
	AB        int `json:"a_b"`
	AB2       int `json:"ab"`
	Strict    int `json:"x_y,case:strict"`
	XY        int `json:"xy"`
	Ignore    int `json:"long_name_of_a_field,case:ignore"`
	Kelvin    int `json:"k_x"`
}

// udFoldRunes has names which other names fold to: by a rune which is not ASCII, by '_' and '-', which folding drops,
// and two of the same folded name, of which a name matches the one of `case:strict` only as it is.
type udFoldRunes struct {
	K      int `json:"k"`
	S      int `json:"s,case:ignore"`
	AB     int `json:"a_b"`
	Strict int `json:"Name,case:strict"`
	Loose  int `json:"NAME,case:ignore"`
}

// udFoldNoDashes has no name of a field with '_' or '-': a name with them matches a field by its folded name, which
// folding drops them from.
type udFoldNoDashes struct {
	K    int `json:"k"`
	S    int `json:"s,case:ignore"`
	Long int `json:"LongName"`
}

// udReplacement has the name of the replacement character, which v1 decodes invalid UTF-8 into: such a name of no
// other field is invalid, of every case.
type udReplacement struct {
	R int `json:"�"`
	K int `json:"k"`
}

var (
	insensitive = unmarshalDiffOptions{matchCaseInsensitiveNames: true}
	reject      = unmarshalDiffOptions{rejectUnknownMembers: true}
)

func unmarshalDiffStructCases() []unmarshalDiffCase {
	wideNames := make([]string, 74)
	for i := range wideNames {
		wideNames[i] = `"F` + strconv.Itoa(i) + `":` + strconv.Itoa(i)
	}
	wide := "{" + strings.Join(wideNames, ",") + "}"
	unknownNames := make([]string, 40)
	for i := range unknownNames {
		unknownNames[i] = `"u` + strconv.Itoa(i) + `":` + strconv.Itoa(i)
	}
	manyUnknown := "{" + strings.Join(unknownNames, ",") + `,"u5":1}`
	// the names of no field of an object of many of them, which are in a map, around an object of few, and the other
	// way round
	manyAround := "{" + strings.Join(unknownNames, ",") + `,"P":{"u1":1,"u2":2},"u5":1}`
	manyInside := `{"Z":1,"P":{` + strings.Join(unknownNames, ",") + `},"Z":2}`
	manyInsideOnce := `{"Z":1,"P":{` + strings.Join(unknownNames, ",") + `},"u1":2}`
	// objects past 64 levels, which the fast scan of the skipper gives up after names which it took, after values
	// which it reads: v1 skips them, and jsontext reads them again
	deepObject := `{"a":1,"b":` + strings.Repeat("[", 70) + strings.Repeat("]", 70) + `}`
	deepDuplicate := `{"a":1,"b":` + strings.Repeat("[", 70) + `{"a":1,"a":2}` + strings.Repeat("]", 70) + `}`
	// a value of no field at the depth which v1 skips past 64 levels from: of 10000 levels with the objects around it,
	// as many as are allowed, and of one more
	deepUnknown := func(levels int) string {
		const around = 9970
		return strings.Repeat(`{"R":`, around-1) + `{"Z":` + strings.Repeat("[", levels) + strings.Repeat("]", levels) +
			strings.Repeat("}", around)
	}
	return []unmarshalDiffCase{
		{name: "struct", in: `{"X":1,"y":2}`, out: outOf(udPoint{})},
		{name: "struct merged", in: `{"y":2}`, out: outOf(udPoint{X: 5, Y: 6})},
		{name: "struct null", in: `null`, out: outOf(udPoint{X: 5})},
		{name: "struct empty", in: `{}`, out: outOf(udPoint{X: 5})},
		{name: "struct array", in: `[1]`, out: outOf(udPoint{})},
		{name: "struct string", in: `"x"`, out: outOf(udPoint{})},
		{name: "struct name of another case", in: `{"x":1,"Y":2}`, out: outOf(udPoint{})},
		{name: "struct name of another case insensitive", in: `{"x":1,"Y":2}`, out: outOf(udPoint{}), opts: insensitive},
		{name: "struct escaped name", in: `{"\u0058":1}`, out: outOf(udPoint{})},
		{name: "struct name of invalid UTF-8", in: "{\"X\xff\":1}", out: outOf(udPoint{})},
		{name: "struct name of invalid UTF-8 allowed", in: "{\"X\xff\":1}", out: outOf(udPoint{}), opts: invalidUTF8},
		{name: "struct field twice", in: `{"X":1,"X":2}`, out: outOf(udPoint{})},
		{name: "struct field twice allowed", in: `{"X":1,"X":2}`, out: outOf(udPoint{}), opts: duplicates},
		{name: "struct field twice by case", in: `{"X":1,"x":2}`, out: outOf(udPoint{}), opts: insensitive},
		{name: "struct field twice by escape", in: `{"X":1,"\u0058":2}`, out: outOf(udPoint{})},
		{name: "struct unknown name", in: `{"Z":1,"X":2}`, out: outOf(udPoint{})},
		{name: "struct unknown name twice", in: `{"Z":1,"Z":2}`, out: outOf(udPoint{})},
		{name: "struct unknown name twice allowed", in: `{"Z":1,"Z":2}`, out: outOf(udPoint{}), opts: duplicates},
		{name: "struct unknown name rejected", in: `{"X":1,"Z":2}`, out: outOf(udPoint{}), opts: reject},
		{name: "struct unknown value of invalid UTF-8", in: "{\"Z\":\"\xff\"}", out: outOf(udPoint{})},
		{name: "struct unknown value of a lone surrogate", in: `{"Z":["\ud800"]}`, out: outOf(udPoint{})},
		{name: "struct unknown long value of invalid UTF-8", in: "{\"Z\":\"abcdefgh\xffijklmnop\"}", out: outOf(udPoint{})},
		{name: "struct unknown long value of a lone surrogate", in: `{"Z":"abcdefgh\ud800ijklmnop"}`, out: outOf(udPoint{})},
		{name: "struct unknown value of a duplicate name", in: `{"Z":{"a":1,"a":2}}`, out: outOf(udPoint{})},
		{name: "struct unknown value of a duplicate name allowed", in: `{"Z":{"a":1,"a":2}}`, out: outOf(udPoint{}), opts: duplicates},
		{name: "struct unknown objects of the same names", in: `{"Z":{"a":1},"W":{"a":1},"V":[{"a":1},{"a":2}]}`, out: outOf(udPoint{})},
		{name: "struct unknown objects of the same names and one of a duplicate", in: `{"Z":{"a":1},"W":{"a":1},"V":{"b":1,"b":2}}`, out: outOf(udPoint{})},
		{name: "struct unknown object and a value of invalid UTF-8", in: "{\"Z\":{\"a\":1},\"W\":\"\xff\"}", out: outOf(udPoint{})},
		{name: "struct unknown object and a deep value", in: `{"Z":{"a":1},"W":` + deepObject + `,"V":{"a":1,"b":2}}`, out: outOf(udPoint{})},
		{name: "struct unknown object and a deep value of a duplicate", in: `{"Z":{"a":1},"W":` + deepDuplicate + `}`, out: outOf(udPoint{})},
		{name: "struct unknown array of objects of a duplicate name", in: `{"Z":[{"a":1},{"b":1,"b":2}]}`, out: outOf(udPoint{})},
		{name: "struct unknown array of objects", in: `{"Z":[ {"a":1},{"a":2}],"X":1}`, out: outOf(udPoint{})},
		{name: "struct unknown array of an object after a number of a duplicate name", in: `{"Z":[1,{"a":1,"a":2}]}`, out: outOf(udPoint{})},
		{name: "struct unknown array of an object after a number", in: `{"Z":[1,{"a":1,"b":2}],"X":1}`, out: outOf(udPoint{})},
		{name: "struct unknown value of a syntax error", in: `{"Z":[1,}`, out: outOf(udPoint{})},
		{name: "struct unknown deep value", in: `{"Z":` + strings.Repeat("[", 9999) + strings.Repeat("]", 9999) + `}`, out: outOf(udPoint{})},
		{name: "struct unknown too deep value", in: `{"Z":` + strings.Repeat("[", 10000) + strings.Repeat("]", 10000) + `}`, out: outOf(udPoint{})},
		{name: "struct value of another type", in: `{"X":"1"}`, out: outOf(udPoint{})},
		{name: "struct missing colon", in: `{"X" 1}`, out: outOf(udPoint{})},
		{name: "struct missing comma", in: `{"X":1 "y":2}`, out: outOf(udPoint{})},
		{name: "struct trailing comma", in: `{"X":1,}`, out: outOf(udPoint{})},
		{name: "struct name not a string", in: `{X:1}`, out: outOf(udPoint{})},
		{name: "struct cut", in: `{"X":1`, out: outOf(udPoint{})},
		{name: "struct case ignore", in: `{"ignore-me":"a","NAME":"b"}`, out: outOf(udCased{})},
		{name: "struct case ignore insensitive", in: `{"ignore-me":"a","NAME":"b","STRICT":"c"}`, out: outOf(udCased{}), opts: insensitive},
		{name: "struct case strict", in: `{"strict":"c"}`, out: outOf(udCased{}), opts: insensitive},
		{name: "struct folded names ambiguous", in: `{"FOOBAR":1}`, out: outOf(udFolded{}), opts: insensitive},
		{name: "struct folded name exact", in: `{"foo_bar":1,"FooBar":2}`, out: outOf(udFolded{}), opts: insensitive},
		{name: "struct folded name of no field", in: `{"FOOBAR":1}`, out: outOf(udFolded{})},
		{name: "struct embedded", in: `{"A":1,"B":"b","C":2}`, out: outOf(udOuter{})},
		{name: "struct embedded pointer", in: `{"A":1,"C":2}`, out: outOf(udOuterPtr{})},
		{name: "struct embedded pointer not set", in: `{"C":2}`, out: outOf(udOuterPtr{})},
		{name: "struct embedded unexported pointer", in: `{"X":1}`, out: outOf(udOuterUnexportedPtr{})},
		{name: "struct embedded unexported pointer set", in: `{"X":1}`, out: func() any { v := udOuterUnexportedPtr{udUnexportedEmbed: &udUnexportedEmbed{}}; return &v }},
		{name: "struct map fallback", in: `{"A":1,"B":[2],"C":{"d":3}}`, out: outOf(udMapFallback{})},
		{name: "struct map fallback of an empty name of a syntax error", in: `{"A":1,"":nul`, out: outOf(udMapFallback{})},
		{name: "struct empty name of no field of a syntax error", in: `{"":[ fals, true],"y":1}`, out: outOf(udPoint{})},
		{name: "struct map fallback merged", in: `{"B":{"y":2}}`, out: func() any { v := udMapFallback{Rest: map[string]any{"B": map[string]any{"x": 1.0}}}; return &v }},
		{name: "struct map fallback twice", in: `{"B":1,"B":2}`, out: outOf(udMapFallback{})},
		{name: "struct named map fallback", in: `{"B":2}`, out: outOf(udNamedMapFallback{})},
		{name: "struct named map fallback of another type", in: `{"B":"x"}`, out: outOf(udNamedMapFallback{})},
		{name: "struct pointer fallback", in: `{"B":2}`, out: outOf(udPtrFallback{})},
		{name: "struct inlined fallback", in: `{"A":1,"B":2,"C":3}`, out: outOf(udInlinedFallback{})},
		{name: "struct string option", in: `{"I":"1","P":"2","F":"1.5","U":"255","PP":"3"}`, out: outOf(udStringOption{})},
		{name: "struct string option plain", in: `{"I":1}`, out: outOf(udStringOption{})},
		{name: "struct string option null", in: `{"P":null,"I":null}`, out: outOf(udStringOption{})},
		{name: "struct string option overflow", in: `{"U":"256"}`, out: outOf(udStringOption{})},
		{name: "struct string option slice", in: `{"S":["1"]}`, out: outOf(udStringOption{})},
		{name: "struct string option any", in: `{"A":"1"}`, out: outOf(udStringOption{})},
		{name: "struct string option bool", in: `{"B":"true"}`, out: outOf(udStringOption{})},
		{name: "struct string option string", in: `{"T":"\"x\""}`, out: outOf(udStringOption{})},
		{name: "struct string option map", in: `{"M":{}}`, out: outOf(udStringOption{})},
		{name: "struct format", in: `{"B":"00"}`, out: outOf(udFormat{})},
		{name: "struct format null", in: `null`, out: outOf(udFormat{})},
		{name: "struct unexported only", in: `{"a":1}`, out: outOf(udUnexportedOnly{})},
		{name: "struct unexported only empty", in: `{}`, out: outOf(udUnexportedOnly{})},
		{name: "struct empty type", in: `{"a":1}`, out: outOf(udEmpty{})},
		{name: "struct nested", in: `{"P":{"X":1},"PP":{"y":2},"L":[{"X":3},null],"M":{"a":{"y":4}}}`, out: outOf(udNested{})},
		{name: "struct nested merged", in: `{"P":{"y":9},"PP":{"X":8},"M":{"a":{"X":7}}}`, out: func() any {
			return &udNested{P: udPoint{X: 1}, PP: &udPoint{Y: 2}, M: map[string]udPoint{"a": {Y: 3}}}
		}},
		{name: "struct nested error", in: `{"L":[{"X":1},{"y":"2"}]}`, out: outOf(udNested{})},
		{name: "struct nested duplicate", in: `{"M":{"a":{"X":1},"a":{"y":2}}}`, out: outOf(udNested{})},
		{name: "struct nested duplicate allowed", in: `{"M":{"a":{"X":1},"a":{"y":2}}}`, out: outOf(udNested{}), opts: duplicates},
		{name: "struct names of no field of nested objects", in: `{"Z":1,"P":{"Z":1,"W":2},"PP":{"Z":3},"W":4}`, out: outOf(udNested{})},
		{name: "struct names of no field of the same bits in a nested object", in: `{"aZb":1,"P":{"aXb":1,"aZb":2},"aYb":3}`, out: outOf(udNested{})},
		{name: "struct name of no field twice around a nested object", in: `{"Z":1,"P":{"Z":1},"Z":2}`, out: outOf(udNested{})},
		{name: "struct name of no field twice in a nested object", in: `{"Z":1,"P":{"W":1,"W":2},"Z":2}`, out: outOf(udNested{})},
		{name: "struct names of no field of the elements of an array", in: `{"L":[{"Z":1},{"Z":2},{"Z":3,"W":4}],"Z":5}`, out: outOf(udNested{})},
		{name: "struct name of no field twice around an array of objects", in: `{"Z":0,"L":[{"Z":1},{"Z":2}],"Z":3}`, out: outOf(udNested{})},
		{name: "struct names of no field of the values of a map", in: `{"Z":0,"M":{"a":{"Z":1},"b":{"Z":2,"W":3}},"W":4}`, out: outOf(udNested{})},
		{name: "struct name of no field twice after an error of a nested object", in: `{"Z":0,"L":[{"X":"x","Z":1}],"Z":2}`, out: outOf(udNested{})},
		{name: "struct many names of no field around a nested object", in: manyAround, out: outOf(udNames{P: &udNames{}})},
		{name: "struct many names of no field in a nested object", in: manyInside, out: outOf(udNames{P: &udNames{}})},
		{name: "struct many names of no field in a nested object once", in: manyInsideOnce, out: outOf(udNames{P: &udNames{}})},
		{name: "struct value of no field at the depth limit", in: deepUnknown(30), out: outOf(udDeep{})},
		{name: "struct value of no field past the depth limit", in: deepUnknown(31), out: outOf(udDeep{}), errorOnly: true},
		{name: "struct names of no field around a map of method keys", in: `{"Z":1,"K":{"1":1,"2":2},"W":2}`, out: outOf(udKeyNames{}), errorOnly: true},
		{name: "struct name of no field twice around a map of method keys", in: `{"Z":1,"K":{"Z":1,"2":2},"Z":2}`, out: outOf(udKeyNames{}), errorOnly: true},
		{name: "struct recursive", in: `{"Name":"a","Children":[{"Name":"b"}],"Next":{"Name":"c","Next":{"Name":"d"}}}`, out: outOf(udRecursive{})},
		{name: "struct ignored fields", in: `{"A":1,"-":2,"c":3}`, out: outOf(udOmit{})},
		{name: "struct of many fields", in: wide, out: outOf(udWide{})},
		{name: "struct of many fields twice", in: `{"F70":1,"F70":2}`, out: outOf(udWide{})},
		{name: "struct of many fields twice early", in: `{"F3":1,"F70":2,"F3":3}`, out: outOf(udWide{})},
		{name: "struct names by words", in: `{"abcdefg":1,"abcdefgh":2,"abcdefghijklmnop":3,"abcdefgh_1_ijklmnop":4,"abcdefgh_2_ijklmnop":5}`, out: outOf(udNames{})},
		{name: "struct name of the first and last words of fields", in: `{"abcdefgh_3_ijklmnop":1}`, out: outOf(udNames{})},
		{name: "struct name of the first and last words of fields twice", in: `{"abcdefgh_3_ijklmnop":1,"abcdefgh_3_ijklmnop":2}`, out: outOf(udNames{})},
		{name: "struct name of 7 bytes twice", in: `{"abcdefg":1,"abcdefg":2}`, out: outOf(udNames{})},
		{name: "struct names of the start of a field", in: `{"abcdef":1,"abcdefghi":2,"abcdefghijklmno":3}`, out: outOf(udNames{})},
		{name: "struct empty name", in: `{"":1}`, out: outOf(udNames{})},
		{name: "struct empty name twice", in: `{"":1,"":2}`, out: outOf(udNames{})},
		{name: "struct name of an escaped backslash", in: `{"a\\b":1}`, out: outOf(udNames{})},
		{name: "struct name of an escaped control character", in: `{"a\bb":1}`, out: outOf(udNames{})},
		{name: "struct name without the nul at the end of a field", in: `{"abc":1}`, out: outOf(udNames{})},
		{name: "struct name of a nul at the end", in: `{"abc\u0000":1}`, out: outOf(udNames{})},
		{name: "struct name of a control character", in: "{\"a\x08b\":1}", out: outOf(udNames{})},
		{name: "struct name not ASCII", in: `{"é":1}`, out: outOf(udNames{})},
		{name: "struct name not ASCII escaped", in: `{"\u00e9":1}`, out: outOf(udNames{})},
		{name: "struct name not ASCII of another case", in: `{"É":1}`, out: outOf(udNames{})},
		{name: "struct name not ASCII of another case insensitive", in: `{"É":1}`, out: outOf(udNames{}), opts: insensitive},
		{name: "struct names which differ in case", in: `{"name":1,"Name":2}`, out: outOf(udNames{})},
		{name: "struct names which differ in case insensitive", in: `{"name":1,"Name":2}`, out: outOf(udNames{}), opts: insensitive},
		{name: "struct name of fields which differ in case insensitive", in: `{"NAME":1}`, out: outOf(udNames{}), opts: insensitive},
		{name: "struct many names of no field twice", in: manyUnknown, out: outOf(udNames{})},
		{name: "struct many names of no field twice allowed", in: manyUnknown, out: outOf(udNames{}), opts: duplicates},
		{name: "struct name of the Kelvin sign", in: `{"\u212a":1}`, out: outOf(udFoldRunes{})},
		{name: "struct name of the Kelvin sign insensitive", in: `{"\u212a":1}`, out: outOf(udFoldRunes{}), opts: insensitive},
		{name: "struct name of the long s", in: `{"\u017f":1}`, out: outOf(udFoldRunes{})},
		{name: "struct name without the underscore of a field", in: `{"AB":1}`, out: outOf(udFoldRunes{})},
		{name: "struct name without the underscore of a field insensitive", in: `{"AB":1}`, out: outOf(udFoldRunes{}), opts: insensitive},
		{name: "struct name with a dash for the underscore of a field insensitive", in: `{"a-b":1}`, out: outOf(udFoldRunes{}), opts: insensitive},
		{name: "struct name of a strict field and an ignoring one", in: `{"Name":1,"name":2}`, out: outOf(udFoldRunes{})},
		{name: "struct name of a strict field and an ignoring one insensitive", in: `{"Name":1,"nAME":2}`, out: outOf(udFoldRunes{}), opts: insensitive},
		{name: "struct names of fields of dashes", in: `{"user_name":1,"created-at":2,"id":3}`, out: outOf(udSnake{})},
		{name: "struct names of fields of dashes insensitive", in: `{"UserName":1,"CREATED_AT":2,"Id":3}`, out: outOf(udSnake{}), opts: insensitive},
		{name: "struct names of fields of other dashes insensitive", in: `{"user-name":1,"created_at":2,"__i_d__":3}`, out: outOf(udSnake{}), opts: insensitive},
		{name: "struct names of no field of a struct of dashes insensitive", in: `{"other_name":1,"username_":2,"id":3,"user":4}`, out: outOf(udSnake{}), opts: insensitive},
		{name: "struct name of two fields of dashes insensitive", in: `{"A-B":1}`, out: outOf(udSnake{}), opts: insensitive},
		{name: "struct name of two fields of dashes", in: `{"A-B":1,"a_b":2,"ab":3}`, out: outOf(udSnake{})},
		{name: "struct name of a strict field of dashes insensitive", in: `{"X_Y":1,"x-y":2,"XY":3}`, out: outOf(udSnake{}), opts: insensitive},
		{name: "struct name of an ignoring field of dashes", in: `{"LongNameOfAField":1,"long-name-of-a-field":2}`, out: outOf(udSnake{})},
		{name: "struct name of an ignoring field of dashes once", in: `{"LONG_NAME_OF_A_FIELD":1}`, out: outOf(udSnake{})},
		{name: "struct name of the Kelvin sign of a field of dashes insensitive", in: `{"\u212a-X":1}`, out: outOf(udSnake{}), opts: insensitive},
		{name: "struct names twice of a field of dashes insensitive", in: `{"user_name":1,"USERNAME":2}`, out: outOf(udSnake{}), opts: insensitive},
		{name: "struct name with an underscore of no field with one", in: `{"_k":1,"long_name":2}`, out: outOf(udFoldNoDashes{})},
		{name: "struct long name of another case insensitive", in: `{"LONGNAME":1}`, out: outOf(udFoldNoDashes{}), opts: insensitive},
		{name: "struct long name of another case with dashes insensitive", in: `{"long-name":1,"LongName":2}`, out: outOf(udFoldNoDashes{}), opts: insensitive},
		{name: "struct name of invalid UTF-8 of the replacement character", in: "{\"\xff\":1}", out: outOf(udReplacement{})},
		{name: "struct name of invalid UTF-8 of the replacement character insensitive", in: "{\"\xff\":1}", out: outOf(udReplacement{}), opts: insensitive},
		{name: "struct name of invalid UTF-8 of the replacement character allowed insensitive", in: "{\"\xff\":1,\"K\":2}", out: outOf(udReplacement{}), opts: unmarshalDiffOptions{allowInvalidUTF8: true, matchCaseInsensitiveNames: true}},
		{name: "struct name with an underscore of no field with one insensitive", in: `{"_k":1,"long_name":2}`, out: outOf(udFoldNoDashes{}), opts: insensitive},
		{name: "struct name with a dash of an ignoring field with none", in: `{"s-":1,"-S":2}`, out: outOf(udFoldNoDashes{})},
		{name: "struct name with a dash of the Kelvin sign insensitive", in: `{"-\u212a":1}`, out: outOf(udFoldNoDashes{}), opts: insensitive},
		{name: "struct name with a dash of the Kelvin sign unescaped insensitive", in: `{"-K":1}`, out: outOf(udFoldNoDashes{}), opts: insensitive},
		{name: "struct name with a dash and the name of a field insensitive", in: `{"k":1,"k-":2}`, out: outOf(udFoldNoDashes{}), opts: insensitive},
		{name: "struct name with only dashes insensitive", in: `{"_-":1}`, out: outOf(udFoldNoDashes{}), opts: insensitive},
		{name: "map", in: `{"a":1,"b":2}`, out: outOf(map[string]int(nil))},
		{name: "map merged", in: `{"b":2}`, out: func() any { v := map[string]int{"a": 1}; return &v }},
		{name: "map null", in: `null`, out: func() any { v := map[string]int{"a": 1}; return &v }},
		{name: "map array", in: `[1]`, out: outOf(map[string]int(nil))},
		{name: "map duplicate", in: `{"a":1,"a":2}`, out: outOf(map[string]int(nil))},
		{name: "map duplicate allowed", in: `{"a":1,"a":2}`, out: outOf(map[string]int(nil)), opts: duplicates},
		{name: "map duplicate of an existing key", in: `{"a":2,"a":3}`, out: func() any { v := map[string]int{"a": 1}; return &v }},
		{name: "map existing key", in: `{"a":2}`, out: func() any { v := map[string]int{"a": 1}; return &v }},
		{name: "map of int keys", in: `{"1":1,"-2":2}`, out: outOf(map[int]int(nil))},
		{name: "map of int keys of a fraction", in: `{"1.5":1}`, out: outOf(map[int]int(nil))},
		{name: "map of int keys not numbers", in: `{"x":1}`, out: outOf(map[int]int(nil))},
		{name: "map of int keys of an equal number", in: `{"1":1,"01":2}`, out: outOf(map[int]int(nil))},
		{name: "map of uint8 keys overflow", in: `{"256":1}`, out: outOf(map[uint8]int(nil))},
		{name: "map of float keys", in: `{"1.5":1,"-0":2}`, out: outOf(map[float64]int(nil))},
		{name: "map of float keys equal", in: `{"0":1,"-0":2}`, out: outOf(map[float64]int(nil))},
		{name: "map of bool keys", in: `{"true":1}`, out: outOf(map[bool]int(nil))},
		{name: "map of interface keys", in: `{"a":1}`, out: outOf(map[any]int(nil))},
		{name: "map of named string keys", in: `{"a":1}`, out: outOf(map[udNamedKey]int(nil))},
		{name: "map of pointer keys", in: `{"a":1}`, out: outOf(map[*string]int(nil)), errorOnly: true},
		{name: "map of struct keys", in: `{"a":1}`, out: outOf(map[udPoint]int(nil))},
		{name: "map of struct keys empty", in: `{}`, out: outOf(map[udPoint]int(nil))},
		{name: "map of array keys", in: `{"a":1}`, out: outOf(map[[1]int]int(nil))},
		{name: "map of keys of invalid UTF-8", in: "{\"a\xff\":1}", out: outOf(map[string]int(nil))},
		{name: "map of keys of invalid UTF-8 allowed", in: "{\"a\xff\":1}", out: outOf(map[string]int(nil)), opts: invalidUTF8},
		{name: "map of struct values", in: `{"a":{"X":1}}`, out: outOf(map[string]udPoint(nil))},
		{name: "map of struct values merged", in: `{"a":{"X":1}}`, out: func() any { v := map[string]udPoint{"a": {Y: 2}}; return &v }},
		{name: "map of pointer values merged", in: `{"a":{"X":1}}`, out: func() any { v := map[string]*udPoint{"a": {Y: 2}}; return &v }},
		{name: "map value error", in: `{"a":1,"b":"x"}`, out: outOf(map[string]int(nil))},
		{name: "map of any", in: `{"a":[1,{"b":null}]}`, out: outOf(map[string]any(nil))},
		{name: "map of any merged", in: `{"a":{"y":2}}`, out: func() any { v := map[string]any{"a": map[string]any{"x": 1.0}}; return &v }},
		{name: "map of any duplicate", in: `{"a":1,"a":2}`, out: outOf(map[string]any(nil))},
		{name: "map of any names of nested objects", in: `{"a":{"b":1,"c":2},"b":3,"c":{"a":4}}`, out: outOf(map[string]any(nil))},
		{name: "map of any duplicate around a nested object", in: `{"a":{"b":1},"b":2,"a":3}`, out: outOf(map[string]any(nil))},
		{name: "any duplicate in a nested object", in: `[{"a":1},{"b":{"a":1,"a":2}}]`, out: outOf(any(nil))},
		{name: "map of any duplicate allowed", in: `{"a":{"x":1},"a":{"y":2}}`, out: outOf(map[string]any(nil)), opts: duplicates},
		{name: "map of any string", in: `"x"`, out: outOf(map[string]any(nil))},
		{name: "map of maps", in: `{"a":{"b":1}}`, out: outOf(map[string]map[string]int(nil))},
		{name: "map of maps merged", in: `{"a":{"c":2}}`, out: func() any { v := map[string]map[string]int{"a": {"b": 1}}; return &v }},
		{name: "map deep", in: strings.Repeat(`{"a":`, 10000) + `1` + strings.Repeat(`}`, 10000), out: outOf(map[string]any(nil))},
		{name: "map too deep", in: strings.Repeat(`{"a":`, 10001) + `1` + strings.Repeat(`}`, 10001), out: outOf(map[string]any(nil))},
		{name: "map of int keys stringified", in: `{"1":"2"}`, out: outOf(map[int]int(nil)), opts: stringify},
		{name: "map of pointer keys twice", in: `{"1":1,"1":2}`, out: outOf(map[*int]int(nil)), errorOnly: true},
		{name: "map of pointer keys twice allowed", in: `{"1":1,"1":2}`, out: outOf(map[*int]int(nil)), errorOnly: true, opts: duplicates},
		{name: "map of string pointer keys twice", in: `{"a":1,"a":2}`, out: outOf(map[*string]int(nil)), errorOnly: true},
		{name: "map of interface keys twice", in: `{"a":1,"a":2}`, out: outOf(map[any]int(nil))},
		{name: "map of float keys twice by name", in: `{"1":1,"1.0":2}`, out: outOf(map[float64]int(nil))},
		{name: "map of keys twice after replacement", in: "{\"\xff\":1,\"\xfe\":2}", out: outOf(map[string]int(nil)), opts: invalidUTF8},
		{name: "duration", in: `1`, out: outOf(time.Duration(0))},
		{name: "duration field", in: `{"D":1}`, out: outOf(udDuration{})},
		{name: "duration field null", in: `{"D":null}`, out: outOf(udDuration{})},
		{name: "duration map key", in: `{"1":1}`, out: outOf(map[time.Duration]int(nil))},
		{name: "duration map value", in: `{"a":1}`, out: outOf(map[string]time.Duration(nil))},
		{name: "duration string option", in: `{"D":"1"}`, out: outOf(udStringDuration{})},
		{name: "duration pointer field", in: `{"D":1}`, out: outOf(udDurationPointer{})},
		{name: "duration slice", in: `[1]`, out: outOf([]time.Duration(nil))},
		{name: "duration slice empty", in: `[]`, out: outOf([]time.Duration(nil))},
		{name: "map of pointer keys twice escaped", in: `{"1":1,"\u0031":2}`, out: outOf(map[*int]int(nil)), errorOnly: true},
		{name: "map of string pointer keys twice after replacement", in: "{\"a\xff\":1,\"a\xfe\":2}", out: outOf(map[*string]int(nil)), errorOnly: true, opts: invalidUTF8},
		{name: "struct embedded three levels", in: `{"A":1,"B":"b","C":2,"D":3}`, out: outOf(udThreeLevels{})},
		{name: "struct conflicting names at a depth", in: `{"X":1}`, out: outOf(udConflict{})},
		{name: "struct conflicting names rejected", in: `{"X":1}`, out: outOf(udConflict{}), opts: reject},
		{name: "struct tagged name over an untagged one", in: `{"X":1}`, out: outOf(udTaggedConflict{})},
		{name: "struct embedded pointer chain", in: `{"A":1,"E":2}`, out: outOf(udPtrChain{})},
		{name: "struct embedded unexported struct", in: `{"X":1}`, out: outOf(udUnexportedValue{})},
		{name: "struct case ignore of non-ASCII", in: `{"ÉTÉ":"x"}`, out: outOf(udNonASCIICased{})},
		{name: "struct case ignore ambiguous", in: `{"ab":1}`, out: outOf(udAmbiguousIgnore{})},
		{name: "map of int8 keys overflow", in: `{"128":1}`, out: outOf(map[int8]int(nil))},
		{name: "map of uintptr keys", in: `{"5":1}`, out: outOf(map[uintptr]int(nil))},
		{name: "map of float32 keys overflow", in: `{"1e40":1}`, out: outOf(map[float32]int(nil))},
		{name: "struct string option struct", in: `{"S":{"A":1}}`, out: outOf(udStringOptionMore{})},
		{name: "struct string option array", in: `{"A":["1"]}`, out: outOf(udStringOptionMore{})},
		{name: "struct string option pointer to a slice", in: `{"P":["1"]}`, out: outOf(udStringOptionMore{})},
		{name: "struct string option rejected value of a syntax error", in: `{"S":{"A":}}`, out: outOf(udStringOptionMore{})},
	}
}

type udDuration struct{ D time.Duration }

type udDurationPointer struct{ D *time.Duration }

type udStringDuration struct {
	D time.Duration `json:",string"`
}

type udLevel3 struct{ D int }

type udLevel2 struct {
	udLevel3
	C int
}

type udLevel1 struct {
	udLevel2
	B string
}

type udThreeLevels struct {
	udLevel1
	A int
}

type udConflictA struct{ X int }

type udConflictB struct{ X int }

type udConflict struct {
	udConflictA
	udConflictB
}

type udConflictTagged struct {
	Y int `json:"X"`
}

type udTaggedConflict struct {
	udConflictA
	udConflictTagged
}

type udChainInner struct{ E int }

type UdChainExported struct {
	*udChainInner
}

type udPtrChain struct {
	*UdChainExported
	A int
}

type udUnexportedValue struct {
	udConflictA
}

type udNonASCIICased struct {
	Ete string `json:"été,case:ignore"`
}

type udAmbiguousIgnore struct {
	A int `json:"a_b,case:ignore"`
	B int `json:"a-b,case:ignore"`
}

type udStringOptionMore struct {
	S udPoint `json:",string"`
	A [1]int  `json:",string"`
	P *[]int  `json:",string"`
}

// udRawFallback and udStdRawFallback are a struct of a raw embedded fallback for each package: the jsontext.Value of
// each is its own type, which the other doesn't take as one.
type udRawFallback struct {
	A    int
	Rest jsontext.Value `json:",embed"`
}

type udStdRawFallback struct {
	A    int
	Rest stdjsontext.Value `json:",embed"`
}

// TestUnmarshalRawFallbackMatchesStd compares the decoding into a struct of a raw embedded fallback, a
// jsontext.Value, with the one of encoding/json/v2, each of its own type: the errors, and the values without one.
func TestUnmarshalRawFallbackMatchesStd(t *testing.T) {
	for _, c := range []struct {
		name string
		in   string
		rest string
		opts unmarshalDiffOptions
	}{
		{name: "members", in: `{"A":1,"B":[2],"C":{"d":3}}`},
		{name: "appended", in: `{"B":2}`, rest: `{"a":1} `},
		{name: "after an empty object", in: `{"B":2}`, rest: `{}`},
		{name: "after a comma", in: `{"B":2}`, rest: `{"a":1,}`},
		{name: "not an object", in: `{"B":2}`, rest: `[1]`},
		{name: "white space", in: `{"B":2}`, rest: " \n"},
		{name: "invalid UTF-8", in: "{\"B\":\"\xff\"}"},
		{name: "invalid UTF-8 allowed", in: "{\"B\":\"\xff\"}", opts: invalidUTF8},
		{name: "rejecting unknown members", in: `{"B":2}`, opts: reject},
		{name: "escaped name", in: `{"\u0042":2}`},
		{name: "name twice", in: `{"B":1,"B":2}`},
		{name: "name twice allowed", in: `{"B":1,"B":2}`, opts: duplicates},
		{name: "syntax error in a member", in: `{"B":[1,}`},
	} {
		t.Run(c.name, func(t *testing.T) {
			ours := &udRawFallback{Rest: jsontext.Value(c.rest)}
			theirs := &udStdRawFallback{Rest: stdjsontext.Value(c.rest)}
			err := json.Unmarshal([]byte(c.in), ours, c.opts.ours()...)
			stdErr := stdjson.Unmarshal([]byte(c.in), theirs, c.opts.std()...)
			want := strings.ReplaceAll(describeUnmarshalError(stdErr), "udStdRawFallback", "udRawFallback")
			if got := describeUnmarshalError(err); got != want {
				t.Fatalf("error:\n got %s\nwant %s", got, want)
			}
			if err == nil && (ours.A != theirs.A || string(ours.Rest) != string(theirs.Rest)) {
				t.Fatalf("value:\n got %d %q\nwant %d %q", ours.A, ours.Rest, theirs.A, theirs.Rest)
			}
		})
	}
}
