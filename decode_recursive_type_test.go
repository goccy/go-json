package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// Slice, map and array types which hold values of themselves, directly or through other slices, maps, arrays and
// pointers. SelfSlice is exported, so that the field which embeds it is exported.
type (
	selfArrayOfPtrs    [1]*selfArrayOfPtrs
	selfPtrToArray     *[1]selfPtrToArray
	selfNestedArray    [2][1]*selfNestedArray
	mutualArrayA       [1]*mutualArrayB
	mutualArrayB       [2]*mutualArrayA
	SelfSlice          []SelfSlice
	selfMap            map[string]selfMap
	selfPtrSlice       []*selfPtrSlice
	selfMapOfSlices    map[string][]selfMapOfSlices
	selfPtrMap         map[string]*selfPtrMap
	selfSliceOfPtrMaps []map[string]*selfSliceOfPtrMaps
	mutualSlice        []mutualMap
	mutualMap          map[string]mutualSlice
	selfArrayOfSlices  [1][]selfArrayOfSlices
	selfArrayOfMaps    [1]map[string]selfArrayOfMaps
	selfPtrToSlice     *[]selfPtrToSlice
)

// Map types whose keys hold the map, through pointers, arrays, slices and structs: no key is decoded into them.
type (
	selfKeyMap         map[*selfKeyMap]int
	selfPtrPtrKeyMap   map[**selfPtrPtrKeyMap]int
	selfArrayKeyMap    map[[1]*selfArrayKeyMap]int
	selfPtrArrayKeyMap map[*[1]*selfPtrArrayKeyMap]int
	selfSliceKeyMap    map[*[]selfSliceKeyMap]int
	selfStructKeyMap   map[selfStructKey]int
	selfStructKey      struct{ P *selfStructKeyMap }
	mutualKeyMap       map[*mutualValueMap]int
	mutualValueMap     map[string]mutualKeyMap
)

// textKeyMap is a map type which holds itself in the fields of its values, whose keys have an UnmarshalText which
// fails for the key "bad".
type (
	textKeyMap  map[textKey]textKeyElem
	textKeyElem struct{ F, G textKeyMap }
	textKey     string
)

func (k *textKey) UnmarshalText(b []byte) error {
	if string(b) == "bad" {
		return &stdjson.UnmarshalTypeError{Value: "string", Type: reflect.TypeOf(""), Offset: 7}
	}
	*k = textKey(b)
	return nil
}

// recursiveCheck is what checkRecursiveDecode compares with encoding/json.
type recursiveCheck int

const (
	// checkValuesAndErrors compares the values and the errors.
	checkValuesAndErrors recursiveCheck = iota
	// checkUndecodedKeys is for a map whose keys are not decoded: an encoding/json which does not decode as the one of
	// Go 1.27 ( see stdTypeErrorMayDiffer ) fails for the whole object and leaves the map nil, where Go 1.27 makes the
	// map and drops the entries, so the value of Unmarshal is not compared with the value it leaves then.
	checkUndecodedKeys
	// checkUndecodedPointerKeys is checkUndecodedKeys for keys which are pointers, whose errors are compared with
	// encoding/json by their kinds only, as TestDecodeMapKeyPointers does: go-json reports the type of the pointer, and
	// encoding/json of Go 1.27 the type it points to. The Decoders are compared in full with Unmarshal.
	checkUndecodedPointerKeys
	// checkMethodError is for a decoding which an unmarshal method fails: an encoding/json which does not decode as
	// the one of Go 1.27 stops at the error, where Go 1.27 drops the entry and goes on, so the value of Unmarshal is
	// not compared with the value it leaves.
	checkMethodError
)

// The decoder of a slice, a map or an array type which holds values of itself decodes them, as the decoder of a
// struct type which holds values of itself does, and the values and the errors are the ones of encoding/json of
// Go 1.27, at the top, in fields, in maps, in interface values and as the keys of maps, read whole or element by
// element.
func TestDecodeRecursiveType(t *testing.T) {
	deep := strings.Repeat("[", 400) + strings.Repeat("]", 400)
	deepError := strings.Repeat("[", 300) + "1" + strings.Repeat("]", 300)
	deepMapError := strings.Repeat(`{"a":`, 200) + "1" + strings.Repeat("}", 200)
	deepArray := strings.Repeat("[", 300) + "null" + strings.Repeat("]", 300)
	deepArrayError := strings.Repeat("[", 300) + `"x"` + strings.Repeat("]", 300)
	var long, longMap, longMutual strings.Builder
	long.WriteString("[")
	longMap.WriteString("{")
	longMutual.WriteString("[")
	for i := range 60 {
		if i > 0 {
			long.WriteString(",")
			longMap.WriteString(",")
			longMutual.WriteString(",")
		}
		long.WriteString("[[],[[]]]")
		fmt.Fprintf(&longMap, `"k%d":{"a":{},"b":null}`, i)
		fmt.Fprintf(&longMutual, `{"k%d":[{},{"b":[]}]}`, i)
	}
	long.WriteString("]")
	longMap.WriteString("}")
	longMutual.WriteString("]")
	// the type errors after the first read of 512 bytes, deep in the values
	longError := strings.TrimSuffix(long.String(), "]") + `,[[[[1]]]]]`
	longMapError := strings.TrimSuffix(longMap.String(), "}") + `,"z":{"a":{"b":{"c":1}}}}`
	longMutualError := strings.TrimSuffix(longMutual.String(), "]") + `,{"z":[{"a":[{"b":[1]}]}]}]`

	tests := []struct {
		in    string
		value func() any
		check recursiveCheck
	}{
		{`[[],[[]],null]`, func() any { return new(SelfSlice) }, checkValuesAndErrors},
		{`null`, func() any { return new(SelfSlice) }, checkValuesAndErrors},
		{deep, func() any { return new(SelfSlice) }, checkValuesAndErrors},
		{deepError, func() any { return new(SelfSlice) }, checkValuesAndErrors},
		{long.String(), func() any { return new(SelfSlice) }, checkValuesAndErrors},
		{longError, func() any { return new(SelfSlice) }, checkValuesAndErrors},
		{`[[],{}]`, func() any { return new(SelfSlice) }, checkValuesAndErrors},
		{`{"a":{"b":{}},"c":null}`, func() any { return new(selfMap) }, checkValuesAndErrors},
		{deepMapError, func() any { return new(selfMap) }, checkValuesAndErrors},
		{longMap.String(), func() any { return new(selfMap) }, checkValuesAndErrors},
		{longMapError, func() any { return new(selfMap) }, checkValuesAndErrors},
		{`[[],null,[null,[]]]`, func() any { return new(selfPtrSlice) }, checkValuesAndErrors},
		{`[[],null,[null,["x"]]]`, func() any { return new(selfPtrSlice) }, checkValuesAndErrors},
		{`{"a":[{"b":[]}],"c":null}`, func() any { return new(selfMapOfSlices) }, checkValuesAndErrors},
		{`{"a":[{"b":[1]}]}`, func() any { return new(selfMapOfSlices) }, checkValuesAndErrors},
		{`{"a":{"b":null},"c":{}}`, func() any { return new(selfPtrMap) }, checkValuesAndErrors},
		{`{"a":{"b":[]}}`, func() any { return new(selfPtrMap) }, checkValuesAndErrors},
		{`[{"a":[{"b":null}]},{}]`, func() any { return new(selfSliceOfPtrMaps) }, checkValuesAndErrors},
		{`[{"a":[{"b":true}]}]`, func() any { return new(selfSliceOfPtrMaps) }, checkValuesAndErrors},
		{`[{"a":[{}]},{"b":[]}]`, func() any { return new(mutualSlice) }, checkValuesAndErrors},
		{`{"a":[{"b":[{"c":1}]}]}`, func() any { return new(mutualMap) }, checkValuesAndErrors},
		{longMutual.String(), func() any { return new(mutualSlice) }, checkValuesAndErrors},
		{longMutualError, func() any { return new(mutualSlice) }, checkValuesAndErrors},
		{`[[[[]],[[[]]]]]`, func() any { return new(selfArrayOfSlices) }, checkValuesAndErrors},
		{`[[[["x"]]]]`, func() any { return new(selfArrayOfSlices) }, checkValuesAndErrors},
		{`[{"a":[{}],"b":[{"c":[null]}]}]`, func() any { return new(selfArrayOfMaps) }, checkValuesAndErrors},
		{`[{"a":[{"b":1}]}]`, func() any { return new(selfArrayOfMaps) }, checkValuesAndErrors},
		{`[[],null,[[]]]`, func() any { return new(selfPtrToSlice) }, checkValuesAndErrors},
		{`[[],[{}]]`, func() any { return new(selfPtrToSlice) }, checkValuesAndErrors},

		{`[[null]]`, func() any { return new(selfArrayOfPtrs) }, checkValuesAndErrors},
		{`[[[null]],1]`, func() any { return new(selfArrayOfPtrs) }, checkValuesAndErrors},
		{`[[["x"]]]`, func() any { return new(selfArrayOfPtrs) }, checkValuesAndErrors},
		{deepArray, func() any { return new(selfArrayOfPtrs) }, checkValuesAndErrors},
		{deepArrayError, func() any { return new(selfArrayOfPtrs) }, checkValuesAndErrors},
		{`[[null]]`, func() any { return new(selfPtrToArray) }, checkValuesAndErrors},
		{`[[[1]]]`, func() any { return new(selfPtrToArray) }, checkValuesAndErrors},
		{`[[[[null]]],[null]]`, func() any { return new(selfNestedArray) }, checkValuesAndErrors},
		{`[[true]]`, func() any { return new(selfNestedArray) }, checkValuesAndErrors},
		{`[[[null],[null]]]`, func() any { return new(mutualArrayA) }, checkValuesAndErrors},
		{`[[["x"]]]`, func() any { return new(mutualArrayA) }, checkValuesAndErrors},

		// fields
		{`{"X":[[null]],"A":[[[null],null]]}`, func() any {
			return new(struct {
				X selfArrayOfPtrs
				A mutualArrayA
			})
		}, checkValuesAndErrors},
		{`{"X":[[[true]]]}`, func() any { return new(struct{ X selfArrayOfPtrs }) }, checkValuesAndErrors},
		{`{"N":[[[[null]]],[null]],"P":[[null]]}`, func() any {
			return new(struct {
				N selfNestedArray
				P selfPtrToArray
			})
		}, checkValuesAndErrors},
		{`{"A":[[[null],[[[1]]]]]}`, func() any { return new(struct{ A mutualArrayA }) }, checkValuesAndErrors},
		{`{"X":[[],[[]]]}`, func() any { return new(struct{ X SelfSlice }) }, checkValuesAndErrors},
		{`{"X":[[],[[1]]]}`, func() any { return new(struct{ X SelfSlice }) }, checkValuesAndErrors},
		{`{"M":{"a":{"b":{}}}}`, func() any { return new(struct{ M selfMap }) }, checkValuesAndErrors},
		{`{"M":{"a":{"b":"x"}}}`, func() any { return new(struct{ M selfMap }) }, checkValuesAndErrors},
		{`{"X":[[[]],[]]}`, func() any { return new(struct{ X []SelfSlice }) }, checkValuesAndErrors},
		{`{"X":[[[]],[[true]]]}`, func() any { return new(struct{ X []SelfSlice }) }, checkValuesAndErrors},
		{`{"P":[[],[[]]]}`, func() any { return new(struct{ P *SelfSlice }) }, checkValuesAndErrors},
		{`{"P":[[],[[{}]]]}`, func() any { return new(struct{ P *SelfSlice }) }, checkValuesAndErrors},
		{`{"X":[[],[[]]]}`, func() any {
			return new(struct {
				X SelfSlice `json:",string"`
			})
		}, checkValuesAndErrors},
		{`{"X":"[[]]"}`, func() any {
			return new(struct {
				X SelfSlice `json:",string"`
			})
		}, checkValuesAndErrors},
		{`{"SelfSlice":[[],[[]]]}`, func() any { return new(struct{ SelfSlice }) }, checkValuesAndErrors},
		{`{"SelfSlice":[[],[[1]]]}`, func() any { return new(struct{ SelfSlice }) }, checkValuesAndErrors},
		{`{"X":[[],[[]]],"Y":{"a":{}},"Z":[{"a":[]}]}`, func() any {
			return new(struct {
				X SelfSlice
				Y selfMap
				Z mutualSlice
			})
		}, checkValuesAndErrors},

		// values of maps
		{`{"a":[[],[[]]],"b":null}`, func() any { return new(map[string]SelfSlice) }, checkValuesAndErrors},
		{`{"a":[[],[[1]]]}`, func() any { return new(map[string]SelfSlice) }, checkValuesAndErrors},
		{`{"a":[null,[]],"b":null}`, func() any { return new(map[string]selfPtrToSlice) }, checkValuesAndErrors},
		{`{"a":[null,[1]]}`, func() any { return new(map[string]selfPtrToSlice) }, checkValuesAndErrors},
		{`{"a":[[null]],"b":null}`, func() any { return new(map[string]selfPtrToArray) }, checkValuesAndErrors},
		{`{"a":[[[1]]]}`, func() any { return new(map[string]selfPtrToArray) }, checkValuesAndErrors},
		{`{"a":[[null]],"b":[[[["x"]]]]}`, func() any { return new(map[string]selfArrayOfPtrs) }, checkValuesAndErrors},
		{`{"a":[[[null],null]],"b":[[[[[]]]]]}`, func() any { return new(map[string]mutualArrayA) }, checkValuesAndErrors},

		// interface values
		{`[[],[[]]]`, func() any {
			var r SelfSlice
			var v any = &r
			return &v
		}, checkValuesAndErrors},
		{`[[],[[1]]]`, func() any {
			var r SelfSlice
			var v any = &r
			return &v
		}, checkValuesAndErrors},
		{`[[],[[]]]`, func() any {
			var v any = []SelfSlice{}
			return &v
		}, checkValuesAndErrors},
		{`[[],[[]]]`, func() any {
			var s []SelfSlice
			var v any = &s
			return &v
		}, checkValuesAndErrors},

		// keys of maps
		{`{"a":1}`, func() any { return new(map[selfPtrToSlice]int) }, checkUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(map[*selfPtrSlice]int) }, checkUndecodedPointerKeys},
		{`{}`, func() any { return new(map[*selfPtrSlice]int) }, checkUndecodedKeys},

		// keys which hold the map
		{`{"a":1}`, func() any { return new(selfKeyMap) }, checkUndecodedPointerKeys},
		{`{}`, func() any { return new(selfKeyMap) }, checkUndecodedKeys},
		{`null`, func() any { return new(selfKeyMap) }, checkUndecodedKeys},
		{`{"a":1}`, func() any { return new(selfPtrPtrKeyMap) }, checkUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(selfArrayKeyMap) }, checkUndecodedKeys},
		{`{}`, func() any { return new(selfArrayKeyMap) }, checkUndecodedKeys},
		{`{"a":1}`, func() any { return new(selfPtrArrayKeyMap) }, checkUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(selfSliceKeyMap) }, checkUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(selfStructKeyMap) }, checkUndecodedKeys},
		{`{"a":1}`, func() any { return new(mutualKeyMap) }, checkUndecodedPointerKeys},
		{`{"a":{"b":1}}`, func() any { return new(mutualValueMap) }, checkUndecodedPointerKeys},
		{`{"M":{"a":1},"N":2}`, func() any {
			return new(struct {
				M selfArrayKeyMap
				N int
			})
		}, checkUndecodedKeys},
		{`{"x":{"a":1},"y":{}}`, func() any { return new(map[string]selfStructKeyMap) }, checkUndecodedKeys},
		{`{"M":{"a":1},"N":2}`, func() any {
			return new(struct {
				M selfKeyMap
				N int
			})
		}, checkUndecodedPointerKeys},

		// the keys of a map which is reached again through the fields of its values, which have a method
		{`{"a":{"F":{"b":{"G":{"c":{}}}}}}`, func() any { return new(textKeyMap) }, checkValuesAndErrors},
		{`{"bad":{}}`, func() any { return new(textKeyMap) }, checkValuesAndErrors},
		{`{"a":{"F":{"b":{"G":{"bad":{}}}}}}`, func() any { return new(textKeyMap) }, checkMethodError},
		{`{"F":{"b":{"G":{"bad":{}}}}}`, func() any { return new(textKeyElem) }, checkMethodError},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d %T", i, test.value()), func(t *testing.T) {
			checkRecursiveDecode(t, test.in, test.value, test.check)
		})
	}
}

// checkRecursiveDecode compares the decoding of in into the value of newValue by Unmarshal, and by a Decoder which
// reads it a byte at a time and 512 bytes at a time, with the one of encoding/json, as check says. The type error of
// an encoding/json which does not decode as the one of Go 1.27 may differ from go-json's ( see stdTypeErrorMayDiffer ):
// the error of Unmarshal is not compared with it then, and the Decoders are compared in full with Unmarshal, as they
// are for keys which are pointers.
func checkRecursiveDecode(t *testing.T, in string, newValue func() any, check recursiveCheck) {
	t.Helper()
	stdDescribe := describeTypeError
	if check == checkUndecodedPointerKeys {
		stdDescribe = mapKeyKindError
	}
	compareErrors := func(name string, describe func(error) string, got error, from string, want error) {
		t.Helper()
		if describe(got) != describe(want) {
			t.Errorf("%s %.60s:\n got %s\n%s %s", name, in, describe(got), from, describe(want))
		}
	}
	compareValues := func(name string, got any, from string, want any) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s %.60s: decoded %#v, %s %#v", name, in, got, from, want)
		}
	}
	unmarshaled := newValue()
	unmarshalErr := json.Unmarshal([]byte(in), unmarshaled)
	std := newValue()
	stdErr := stdjson.Unmarshal([]byte(in), std)
	switch {
	case !stdTypeErrorMayDiffer(stdErr):
		compareErrors("Unmarshal", stdDescribe, unmarshalErr, "encoding/json", stdErr)
		compareValues("Unmarshal", unmarshaled, "encoding/json", std)
	case check == checkValuesAndErrors:
		compareValues("Unmarshal", unmarshaled, "encoding/json", std)
	}
	stdStream := newValue()
	stdStreamErr := stdjson.NewDecoder(strings.NewReader(in)).Decode(stdStream)
	want, wantErr, from, describe := stdStream, stdStreamErr, "encoding/json", stdDescribe
	if check == checkUndecodedPointerKeys || stdTypeErrorMayDiffer(stdStreamErr) {
		want, wantErr, from, describe = unmarshaled, unmarshalErr, "Unmarshal", describeTypeError
	}
	for _, d := range []struct {
		name string
		r    io.Reader
	}{
		{"Decoder by 1 byte", &oneByteReader{data: in}},
		{"Decoder by 512 bytes", &chunkedReader{data: in, size: 512}},
	} {
		got := newValue()
		compareErrors(d.name, describe, json.NewDecoder(d.r).Decode(got), from, wantErr)
		compareValues(d.name, got, from, want)
	}
}
