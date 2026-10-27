//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsonv1 "encoding/json"
	stdjson "encoding/json/v2"
	"strings"
	"time"

	gojson "github.com/goccy/go-json"
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
	}
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
	return []unmarshalDiffCase{
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
