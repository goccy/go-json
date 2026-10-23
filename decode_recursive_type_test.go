package json_test

import (
	stdjson "encoding/json"
	"errors"
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

// recursiveComparison is what compareRecursiveDecode compares with encoding/json.
type recursiveComparison int

const (
	// compareValuesAndErrors compares the values and the errors.
	compareValuesAndErrors recursiveComparison = iota
	// compareUndecodedKeys is for a map whose keys are not decoded: an encoding/json which does not decode as the one
	// of Go 1.27 ( see stdTypeErrorMayDiffer ) fails for the whole object and leaves the map nil, where Go 1.27 makes
	// the map and drops the entries, so the value of Unmarshal is not compared with the value it leaves then.
	compareUndecodedKeys
	// compareUndecodedPointerKeys is compareUndecodedKeys for keys which are pointers, whose errors are compared with
	// encoding/json by their kinds only, as TestDecodeMapKeyPointers does: go-json reports the type of the pointer, and
	// encoding/json of Go 1.27 the type it points to. The Decoders are compared in full with Unmarshal.
	compareUndecodedPointerKeys
	// compareMethodError is for a decoding which an unmarshal method fails: an encoding/json which does not decode as
	// the one of Go 1.27 stops at the error, where Go 1.27 drops the entry and goes on, so the value of Unmarshal is
	// not compared with the value it leaves.
	compareMethodError
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
		in      string
		value   func() any
		compare recursiveComparison
	}{
		{`[[],[[]],null]`, func() any { return new(SelfSlice) }, compareValuesAndErrors},
		{`null`, func() any { return new(SelfSlice) }, compareValuesAndErrors},
		{deep, func() any { return new(SelfSlice) }, compareValuesAndErrors},
		{deepError, func() any { return new(SelfSlice) }, compareValuesAndErrors},
		{long.String(), func() any { return new(SelfSlice) }, compareValuesAndErrors},
		{longError, func() any { return new(SelfSlice) }, compareValuesAndErrors},
		{`[[],{}]`, func() any { return new(SelfSlice) }, compareValuesAndErrors},
		{`{"a":{"b":{}},"c":null}`, func() any { return new(selfMap) }, compareValuesAndErrors},
		{deepMapError, func() any { return new(selfMap) }, compareValuesAndErrors},
		{longMap.String(), func() any { return new(selfMap) }, compareValuesAndErrors},
		{longMapError, func() any { return new(selfMap) }, compareValuesAndErrors},
		{`[[],null,[null,[]]]`, func() any { return new(selfPtrSlice) }, compareValuesAndErrors},
		{`[[],null,[null,["x"]]]`, func() any { return new(selfPtrSlice) }, compareValuesAndErrors},
		{`{"a":[{"b":[]}],"c":null}`, func() any { return new(selfMapOfSlices) }, compareValuesAndErrors},
		{`{"a":[{"b":[1]}]}`, func() any { return new(selfMapOfSlices) }, compareValuesAndErrors},
		{`{"a":{"b":null},"c":{}}`, func() any { return new(selfPtrMap) }, compareValuesAndErrors},
		{`{"a":{"b":[]}}`, func() any { return new(selfPtrMap) }, compareValuesAndErrors},
		{`[{"a":[{"b":null}]},{}]`, func() any { return new(selfSliceOfPtrMaps) }, compareValuesAndErrors},
		{`[{"a":[{"b":true}]}]`, func() any { return new(selfSliceOfPtrMaps) }, compareValuesAndErrors},
		{`[{"a":[{}]},{"b":[]}]`, func() any { return new(mutualSlice) }, compareValuesAndErrors},
		{`{"a":[{"b":[{"c":1}]}]}`, func() any { return new(mutualMap) }, compareValuesAndErrors},
		{longMutual.String(), func() any { return new(mutualSlice) }, compareValuesAndErrors},
		{longMutualError, func() any { return new(mutualSlice) }, compareValuesAndErrors},
		{`[[[[]],[[[]]]]]`, func() any { return new(selfArrayOfSlices) }, compareValuesAndErrors},
		{`[[[["x"]]]]`, func() any { return new(selfArrayOfSlices) }, compareValuesAndErrors},
		{`[{"a":[{}],"b":[{"c":[null]}]}]`, func() any { return new(selfArrayOfMaps) }, compareValuesAndErrors},
		{`[{"a":[{"b":1}]}]`, func() any { return new(selfArrayOfMaps) }, compareValuesAndErrors},
		{`[[],null,[[]]]`, func() any { return new(selfPtrToSlice) }, compareValuesAndErrors},
		{`[[],[{}]]`, func() any { return new(selfPtrToSlice) }, compareValuesAndErrors},

		{`[[null]]`, func() any { return new(selfArrayOfPtrs) }, compareValuesAndErrors},
		{`[[[null]],1]`, func() any { return new(selfArrayOfPtrs) }, compareValuesAndErrors},
		{`[[["x"]]]`, func() any { return new(selfArrayOfPtrs) }, compareValuesAndErrors},
		{deepArray, func() any { return new(selfArrayOfPtrs) }, compareValuesAndErrors},
		{deepArrayError, func() any { return new(selfArrayOfPtrs) }, compareValuesAndErrors},
		{`[[null]]`, func() any { return new(selfPtrToArray) }, compareValuesAndErrors},
		{`[[[1]]]`, func() any { return new(selfPtrToArray) }, compareValuesAndErrors},
		{`[[[[null]]],[null]]`, func() any { return new(selfNestedArray) }, compareValuesAndErrors},
		{`[[true]]`, func() any { return new(selfNestedArray) }, compareValuesAndErrors},
		{`[[[null],[null]]]`, func() any { return new(mutualArrayA) }, compareValuesAndErrors},
		{`[[["x"]]]`, func() any { return new(mutualArrayA) }, compareValuesAndErrors},

		// fields
		{`{"X":[[null]],"A":[[[null],null]]}`, func() any {
			return new(struct {
				X selfArrayOfPtrs
				A mutualArrayA
			})
		}, compareValuesAndErrors},
		{`{"X":[[[true]]]}`, func() any { return new(struct{ X selfArrayOfPtrs }) }, compareValuesAndErrors},
		{`{"N":[[[[null]]],[null]],"P":[[null]]}`, func() any {
			return new(struct {
				N selfNestedArray
				P selfPtrToArray
			})
		}, compareValuesAndErrors},
		{`{"A":[[[null],[[[1]]]]]}`, func() any { return new(struct{ A mutualArrayA }) }, compareValuesAndErrors},
		{`{"X":[[],[[]]]}`, func() any { return new(struct{ X SelfSlice }) }, compareValuesAndErrors},
		{`{"X":[[],[[1]]]}`, func() any { return new(struct{ X SelfSlice }) }, compareValuesAndErrors},
		{`{"M":{"a":{"b":{}}}}`, func() any { return new(struct{ M selfMap }) }, compareValuesAndErrors},
		{`{"M":{"a":{"b":"x"}}}`, func() any { return new(struct{ M selfMap }) }, compareValuesAndErrors},
		{`{"X":[[[]],[]]}`, func() any { return new(struct{ X []SelfSlice }) }, compareValuesAndErrors},
		{`{"X":[[[]],[[true]]]}`, func() any { return new(struct{ X []SelfSlice }) }, compareValuesAndErrors},
		{`{"P":[[],[[]]]}`, func() any { return new(struct{ P *SelfSlice }) }, compareValuesAndErrors},
		{`{"P":[[],[[{}]]]}`, func() any { return new(struct{ P *SelfSlice }) }, compareValuesAndErrors},
		{`{"X":[[],[[]]]}`, func() any {
			return new(struct {
				X SelfSlice `json:",string"`
			})
		}, compareValuesAndErrors},
		{`{"X":"[[]]"}`, func() any {
			return new(struct {
				X SelfSlice `json:",string"`
			})
		}, compareValuesAndErrors},
		{`{"SelfSlice":[[],[[]]]}`, func() any { return new(struct{ SelfSlice }) }, compareValuesAndErrors},
		{`{"SelfSlice":[[],[[1]]]}`, func() any { return new(struct{ SelfSlice }) }, compareValuesAndErrors},
		{`{"X":[[],[[]]],"Y":{"a":{}},"Z":[{"a":[]}]}`, func() any {
			return new(struct {
				X SelfSlice
				Y selfMap
				Z mutualSlice
			})
		}, compareValuesAndErrors},

		// values of maps
		{`{"a":[[],[[]]],"b":null}`, func() any { return new(map[string]SelfSlice) }, compareValuesAndErrors},
		{`{"a":[[],[[1]]]}`, func() any { return new(map[string]SelfSlice) }, compareValuesAndErrors},
		{`{"a":[null,[]],"b":null}`, func() any { return new(map[string]selfPtrToSlice) }, compareValuesAndErrors},
		{`{"a":[null,[1]]}`, func() any { return new(map[string]selfPtrToSlice) }, compareValuesAndErrors},
		{`{"a":[[null]],"b":null}`, func() any { return new(map[string]selfPtrToArray) }, compareValuesAndErrors},
		{`{"a":[[[1]]]}`, func() any { return new(map[string]selfPtrToArray) }, compareValuesAndErrors},
		{`{"a":[[null]],"b":[[[["x"]]]]}`, func() any { return new(map[string]selfArrayOfPtrs) }, compareValuesAndErrors},
		{`{"a":[[[null],null]],"b":[[[[[]]]]]}`, func() any { return new(map[string]mutualArrayA) }, compareValuesAndErrors},

		// interface values
		{`[[],[[]]]`, func() any {
			var r SelfSlice
			var v any = &r
			return &v
		}, compareValuesAndErrors},
		{`[[],[[1]]]`, func() any {
			var r SelfSlice
			var v any = &r
			return &v
		}, compareValuesAndErrors},
		{`[[],[[]]]`, func() any {
			var v any = []SelfSlice{}
			return &v
		}, compareValuesAndErrors},
		{`[[],[[]]]`, func() any {
			var s []SelfSlice
			var v any = &s
			return &v
		}, compareValuesAndErrors},

		// keys of maps
		{`{"a":1}`, func() any { return new(map[selfPtrToSlice]int) }, compareUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(map[*selfPtrSlice]int) }, compareUndecodedPointerKeys},
		{`{}`, func() any { return new(map[*selfPtrSlice]int) }, compareUndecodedKeys},

		// keys which hold the map
		{`{"a":1}`, func() any { return new(selfKeyMap) }, compareUndecodedPointerKeys},
		{`{}`, func() any { return new(selfKeyMap) }, compareUndecodedKeys},
		{`null`, func() any { return new(selfKeyMap) }, compareUndecodedKeys},
		{`{"a":1}`, func() any { return new(selfPtrPtrKeyMap) }, compareUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(selfArrayKeyMap) }, compareUndecodedKeys},
		{`{}`, func() any { return new(selfArrayKeyMap) }, compareUndecodedKeys},
		{`{"a":1}`, func() any { return new(selfPtrArrayKeyMap) }, compareUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(selfSliceKeyMap) }, compareUndecodedPointerKeys},
		{`{"a":1}`, func() any { return new(selfStructKeyMap) }, compareUndecodedKeys},
		{`{"a":1}`, func() any { return new(mutualKeyMap) }, compareUndecodedPointerKeys},
		{`{"a":{"b":1}}`, func() any { return new(mutualValueMap) }, compareUndecodedPointerKeys},
		{`{"M":{"a":1},"N":2}`, func() any {
			return new(struct {
				M selfArrayKeyMap
				N int
			})
		}, compareUndecodedKeys},
		{`{"x":{"a":1},"y":{}}`, func() any { return new(map[string]selfStructKeyMap) }, compareUndecodedKeys},
		{`{"M":{"a":1},"N":2}`, func() any {
			return new(struct {
				M selfKeyMap
				N int
			})
		}, compareUndecodedPointerKeys},

		// the keys of a map which is reached again through the fields of its values, which have a method
		{`{"a":{"F":{"b":{"G":{"c":{}}}}}}`, func() any { return new(textKeyMap) }, compareValuesAndErrors},
		{`{"bad":{}}`, func() any { return new(textKeyMap) }, compareValuesAndErrors},
		{`{"a":{"F":{"b":{"G":{"bad":{}}}}}}`, func() any { return new(textKeyMap) }, compareMethodError},
		{`{"F":{"b":{"G":{"bad":{}}}}}`, func() any { return new(textKeyElem) }, compareMethodError},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d %T", i, test.value()), func(t *testing.T) {
			compareRecursiveDecode(t, test.in, test.value, test.compare)
		})
	}
}

// Struct types which embed a struct type holding them, which is still being compiled when they are. The embedded
// types are exported, so that the fields which embed them are exported.
type (
	// embedded by value, reached through a slice
	AncestorInSlice struct {
		X int
		L []embedsAncestorInSlice
	}
	embedsAncestorInSlice struct {
		AncestorInSlice
		Y int
	}

	// embedded by pointer, reached through a slice
	PtrAncestorInSlice struct {
		X int
		L []embedsPtrAncestorInSlice
	}
	embedsPtrAncestorInSlice struct {
		*PtrAncestorInSlice
		Y int
	}

	// reached through a slice and a map
	AncestorInSliceAndMap struct {
		X int
		L []embedsAncestorInSliceAndMap
		M map[string]embedsAncestorInSliceAndMap
	}
	embedsAncestorInSliceAndMap struct {
		AncestorInSliceAndMap
		Y int
	}

	// reached through a pointer field
	AncestorInPtr struct {
		X int
		P *embedsAncestorInPtr
	}
	embedsAncestorInPtr struct {
		AncestorInPtr
		Y int
	}

	// embedding each other by pointer
	MutualEmbedA struct {
		*MutualEmbedB
		X int
	}
	MutualEmbedB struct {
		*MutualEmbedA
		Y int
	}

	// embedding each other by pointer, reached through a slice, with a key promoted from both at the same depth
	MutualEmbedInSliceA struct {
		*MutualEmbedInSliceB
		X int
		S []MutualEmbedInSliceB
	}
	MutualEmbedInSliceB struct {
		*MutualEmbedInSliceA
		Y int
		Z int `json:"X"`
	}

	// embedding itself by pointer, and embedded by value in the elements of its slice
	SelfEmbed struct {
		*SelfEmbed
		X int
		L []embedsSelfEmbed
	}
	embedsSelfEmbed struct {
		SelfEmbed
		W int
	}

	// embedded two levels down
	AncestorTwoLevels struct {
		X int
		L []embedsTwoLevelsOuter
	}
	EmbedsTwoLevelsInner struct {
		AncestorTwoLevels
		Y int
	}
	embedsTwoLevelsOuter struct {
		EmbedsTwoLevelsInner
		Z int
	}

	// embedded beside a struct with a field of the same key: neither field is promoted
	AncestorWithConflict struct {
		X int
		L []embedsWithConflict
	}
	ConflictingX       struct{ X int }
	embedsWithConflict struct {
		AncestorWithConflict
		ConflictingX
	}

	// embedded with a key of its own, so that its fields are not promoted
	TaggedAncestor struct {
		X int
		L []embedsTaggedAncestor
	}
	embedsTaggedAncestor struct {
		TaggedAncestor `json:"va"`
		Y              int
	}

	// with a field of the option ,string, promoted
	StringAncestor struct {
		X int `json:",string"`
		L []embedsStringAncestor
	}
	embedsStringAncestor struct {
		StringAncestor
		Y int
	}

	// embedded by pointer, reached through an array of pointers and a pointer to a slice
	AncestorInArrayAndSlicePtr struct {
		X int
		A [1]*embedsAncestorInArrayAndSlicePtr
		P *[]embedsAncestorInArrayAndSlicePtr
	}
	embedsAncestorInArrayAndSlicePtr struct {
		*AncestorInArrayAndSlicePtr
		Y int
	}

	// unexported, embedded by pointer, which can't be allocated
	unexportedAncestor struct {
		X int
		L []EmbedsUnexportedAncestor
	}
	EmbedsUnexportedAncestor struct {
		*unexportedAncestor
		Y int
	}

	// the same, in a struct embedded by value
	unexportedAncestorBelowValue struct {
		X int
		L []EmbedsUnexportedAncestorBelowValue
	}
	EmbedsUnexportedAncestorBelowValue struct {
		PointsToUnexportedAncestor
		Y int
	}
	PointsToUnexportedAncestor struct{ *unexportedAncestorBelowValue }

	// a struct embedded in a struct which it embeds by pointer, beside another embedded struct, whose fields are
	// promoted to both
	CycleWithSibling struct {
		CycleBack
		CycleSibling
	}
	CycleBack struct {
		*CycleWithSibling
		Y int
	}
	CycleSibling struct{ Z int }

	// the same, embedded by pointer
	PtrCycleWithSibling struct {
		*PtrCycleBack
		CycleSibling
		X int
	}
	PtrCycleBack struct {
		*PtrCycleWithSibling
		Y int
	}

	// the same, with the other embedded struct in the struct embedded by pointer
	CycleSiblingBelow struct {
		*CycleBackWithSibling
		X int
	}
	CycleBackWithSibling struct {
		*CycleSiblingBelow
		CycleSibling
		Y int
	}

	// the same, with a slice of the struct embedded
	SliceCycleWithSibling struct {
		SliceCycleBack
		CycleSibling
	}
	SliceCycleBack struct {
		*SliceCycleWithSibling
		Y int
		L []SliceCycleBack
	}
)

// A struct which embeds a struct type still being compiled, such as one which holds it, has the fields which that
// type promotes, as encoding/json decodes them.
func TestDecodeEmbeddedAncestor(t *testing.T) {
	tests := []struct {
		in    string
		value func() any
	}{
		{`{"X":1,"L":[{"X":2,"Y":3,"L":[{"X":4}]}]}`, func() any { return new(AncestorInSlice) }},
		{`{"X":1,"L":[{"Y":3,"X":"a"}]}`, func() any { return new(AncestorInSlice) }},
		{`{"X":1,"L":[{"X":2,"Y":3,"L":[{"X":4}]}]}`, func() any { return new(PtrAncestorInSlice) }},
		{`{"X":1,"L":[{"X":true}]}`, func() any { return new(PtrAncestorInSlice) }},
		{`{"X":1,"L":[{"X":2,"Y":3}],"M":{"a":{"X":4,"L":[{"X":5}],"Y":6}}}`, func() any { return new(AncestorInSliceAndMap) }},
		{`{"X":1,"P":{"X":2,"Y":3,"P":{"X":4}}}`, func() any { return new(AncestorInPtr) }},
		{`{"X":1,"Y":2}`, func() any { return new(MutualEmbedA) }},
		{`{"X":1,"Y":2}`, func() any { return new(MutualEmbedB) }},
		{`{"X":1,"Y":2,"S":[{"X":3,"Y":4,"S":[{"Y":5}]}]}`, func() any { return new(MutualEmbedInSliceA) }},
		{`{"X":1,"Y":2,"S":[{"X":3,"Y":4}]}`, func() any { return new(MutualEmbedInSliceB) }},
		{`{"X":1,"L":[{"X":2,"W":3,"L":[{"X":4}]}]}`, func() any { return new(SelfEmbed) }},
		{`{"X":1,"L":[{"X":2,"Y":3,"Z":4,"L":[{"Z":5}]}]}`, func() any { return new(AncestorTwoLevels) }},
		{`{"X":1,"L":[{"X":2,"L":[{}]}]}`, func() any { return new(AncestorWithConflict) }},
		{`{"X":1,"L":[{"va":{"X":2,"L":[{"Y":3}]},"X":9,"Y":4}]}`, func() any { return new(TaggedAncestor) }},
		{`{"X":"1","L":[{"X":"2","Y":3}]}`, func() any { return new(StringAncestor) }},
		{`{"X":"1","L":[{"X":2,"Y":3}]}`, func() any { return new(StringAncestor) }},
		{`{"X":1,"A":[{"X":2,"Y":3}],"P":[{"X":4,"Y":5}]}`, func() any { return new(AncestorInArrayAndSlicePtr) }},
		{`{"X":1,"L":[{"Y":3}]}`, func() any { return new(unexportedAncestor) }},
		{`{"X":1,"L":[{"X":2,"Y":3}]}`, func() any { return new(unexportedAncestor) }},
		{`{"X":1,"L":[{"Y":3}]}`, func() any { return new(unexportedAncestorBelowValue) }},
		{`{"X":1,"L":[{"X":2,"Y":3}]}`, func() any { return new(unexportedAncestorBelowValue) }},
		{`{"Y":2,"Z":1}`, func() any { return new(CycleWithSibling) }},
		{`{"Y":2,"Z":1}`, func() any { return new(CycleBack) }},
		{`{"X":3,"Y":2,"Z":1}`, func() any { return new(PtrCycleWithSibling) }},
		{`{"X":3,"Y":2,"Z":1}`, func() any { return new(PtrCycleBack) }},
		{`{"X":3,"Y":2,"Z":1}`, func() any { return new(CycleSiblingBelow) }},
		{`{"X":3,"Y":2,"Z":1}`, func() any { return new(CycleBackWithSibling) }},
		{`{"Y":1,"Z":2,"L":[{"Y":3,"Z":4}]}`, func() any { return new(SliceCycleWithSibling) }},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d %T", i, test.value()), func(t *testing.T) {
			compareRecursiveDecode(t, test.in, test.value, compareValuesAndErrors)
		})
	}
}

// compareRecursiveDecode compares the decoding of in into the value of newValue by Unmarshal, and by a Decoder which
// reads it a byte at a time and 512 bytes at a time, with the one of encoding/json, as compare says. The type error of
// an encoding/json which does not decode as the one of Go 1.27 may differ from go-json's ( see stdTypeErrorMayDiffer ),
// and some type errors of Go 1.27 are other errors there: the error of Unmarshal is not compared with it then, and the
// Decoders are compared in full with Unmarshal, as they are for keys which are pointers.
func compareRecursiveDecode(t *testing.T, in string, newValue func() any, compare recursiveComparison) {
	t.Helper()
	stdDescribe := describeTypeError
	if compare == compareUndecodedPointerKeys {
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
	var typeErr *json.UnmarshalTypeError
	errMayDiffer := stdTypeErrorMayDiffer(stdErr) || !stdDecodesAsGo127 && stdErr != nil && errors.As(unmarshalErr, &typeErr)
	switch {
	case !errMayDiffer:
		compareErrors("Unmarshal", stdDescribe, unmarshalErr, "encoding/json", stdErr)
		compareValues("Unmarshal", unmarshaled, "encoding/json", std)
	case compare == compareValuesAndErrors:
		compareValues("Unmarshal", unmarshaled, "encoding/json", std)
	}
	stdStream := newValue()
	stdStreamErr := stdjson.NewDecoder(strings.NewReader(in)).Decode(stdStream)
	want, wantErr, from, describe := stdStream, stdStreamErr, "encoding/json", stdDescribe
	if compare == compareUndecodedPointerKeys || errMayDiffer || stdTypeErrorMayDiffer(stdStreamErr) {
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
