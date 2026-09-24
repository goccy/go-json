package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// The keys of an object are matched with the fields of a struct as encoding/json does: an exact match first,
// then a match by case folding, in which the first field of a folded key wins. These tests decode the same
// inputs with encoding/json and compare.

func compareStructKeys[T any](t *testing.T, inputs []string) {
	t.Helper()
	for _, in := range inputs {
		var want, got, gotOf T
		errStd := stdjson.Unmarshal([]byte(in), &want)
		err := json.Unmarshal([]byte(in), &got)
		errOf := json.UnmarshalOf([]byte(in), &gotOf)
		if (errStd == nil) != (err == nil) || (errStd == nil) != (errOf == nil) {
			t.Fatalf("%s: encoding/json: %v, Unmarshal: %v, UnmarshalOf: %v", in, errStd, err, errOf)
		}
		if errStd != nil {
			continue
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s:\n got %+v\nwant %+v", in, got, want)
		}
		if !reflect.DeepEqual(gotOf, want) {
			t.Fatalf("%s: UnmarshalOf:\n got %+v\nwant %+v", in, gotOf, want)
		}
	}
}

// keyVariants returns the key in cases of its letters.
func keyVariants(key string) []string {
	return []string{key, strings.ToLower(key), strings.ToUpper(key), strings.ToUpper(key[:1]) + strings.ToLower(key[1:])}
}

type manyFields struct {
	F1  string `json:"f1"`
	F2  string `json:"f2"`
	F3  string `json:"f3"`
	F4  string `json:"f4"`
	F5  string `json:"f5"`
	F6  string `json:"f6"`
	F7  string `json:"f7"`
	F8  string `json:"f8"`
	F9  string `json:"f9"`
	F10 string `json:"f10"`
	F11 string `json:"f11"`
	F12 string `json:"f12"`
	F13 string `json:"f13"`
	F14 string `json:"f14"`
	F15 string `json:"f15"`
	F16 string `json:"f16"`
	F17 string `json:"f17"`
	F18 string `json:"someLongerFieldName"`
	F19 string `json:"aVeryLongFieldNameWhichIsLongerThanSixteenBytes"`
}

func TestStructKeyCaseInsensitiveManyFields(t *testing.T) {
	// A struct of more than 16 fields matches by case folding too ( issues 568 and 427 ).
	var inputs []string
	for _, key := range []string{"f1", "f17", "someLongerFieldName", "aVeryLongFieldNameWhichIsLongerThanSixteenBytes"} {
		for _, v := range keyVariants(key) {
			inputs = append(inputs, fmt.Sprintf(`{%q:"x"}`, v))
		}
	}
	compareStructKeys[manyFields](t, inputs)
}

type embeddedTagBase struct {
	B string `json:"bB"`
}

type embeddedTagOuter struct {
	embeddedTagBase
	A string `json:"aA"`
}

func TestStructKeyCaseInsensitiveEmbedded(t *testing.T) {
	// The fields of an embedded struct whose tags have upper case letters match by case folding ( issue 470 ).
	compareStructKeys[embeddedTagOuter](t, []string{
		`{"Aa":"111","Bb":"222"}`, `{"aa":"1","bb":"2"}`, `{"AA":"1","BB":"2"}`, `{"aA":"1","bB":"2"}`,
	})
}

type foldPrecedence struct {
	Upper  string `json:"NAME"`
	Lower  string `json:"name"`
	Kelvin string `json:"\u212a"` // KELVIN SIGN folds to K
	Sharp  string `json:"ſ"`      // LATIN SMALL LETTER LONG S folds to S
	Uni    string `json:"über"`
}

func TestStructKeyPrecedence(t *testing.T) {
	// An exact match wins over a match by case folding, and the first field wins among the folded matches.
	compareStructKeys[foldPrecedence](t, []string{
		`{"NAME":"a"}`, `{"name":"a"}`, `{"Name":"a"}`, `{"nAmE":"a","NAME":"b"}`,
		`{"k":"kelvin"}`, `{"K":"kelvin"}`, `{"\u212a":"kelvin"}`,
		`{"s":"sharp"}`, `{"S":"sharp"}`, `{"ſ":"sharp"}`,
		`{"über":"u"}`, `{"ÜBER":"u"}`, `{"Über":"u"}`,
	})
}

type escapedKeys struct {
	Slash string `json:"a/b"`
	Quote string `json:"q\"uote"`
	X     string `json:"x/"`
	Plain string `json:"plain"`
}

func TestStructKeyEscaped(t *testing.T) {
	// Escaped keys are decoded before they are matched ( issues 604, 577, 575, 445 ).
	compareStructKeys[escapedKeys](t, []string{
		`{"a\/b":"1"}`, `{"a\/Z":"1"}`, `{"x\/":"4"}`, `{"q\"uote":"q"}`, `{"\u0070lain":"p"}`, `{"PL\u0041IN":"p"}`,
		`{"\u0000":"z"}`, `{"a\u002fb":"1"}`,
		// malformed
		`{"\`, `{"\0d\`, `{"\q":1}`, `{"\0":1}`, `{"a\/`, `{"`, `{"a`, `{"a":`, `{"a":1,}`, `{"a" 1}`, `{1:2}`, `{null:1}`,
	})
}

func TestStructKeyMalformedSequence(t *testing.T) {
	// A sequence of malformed inputs never panics ( issue 604 ).
	type doc struct {
		ID        string              `json:"id"`
		Meta      map[string][]string `json:"meta"`
		ContentMD string              `json:"contentMd"`
	}
	inputs := []string{`[{"x":1},{"y":`, `{"\`, `{"\0d\`, `{"id`, `{"id":"\`, `{"contentMd":`}
	for i := 0; i < 100; i++ {
		var d doc
		_ = json.Unmarshal([]byte(inputs[i%len(inputs)]), &d)
		_ = json.UnmarshalOf([]byte(inputs[i%len(inputs)]), &d)
		_ = json.NewDecoder(strings.NewReader(inputs[i%len(inputs)])).Decode(&d)
	}
}

func TestStructKeyUnknownFields(t *testing.T) {
	type small struct {
		A int
		B string
	}
	compareStructKeys[small](t, []string{
		`{"unknown":1,"A":2}`, `{"unkn\u006fwn":{"x":[1,2]},"B":"b"}`, `{"a":1,"ab":2,"abc":3,"b":"x","bcd":"y"}`,
		`{"` + strings.Repeat("long", 20) + `":1,"A":3}`, `{"élan":1,"A":4}`,
	})
	// The key of an unknown field is reported decoded.
	dec := json.NewDecoder(strings.NewReader(`{"unkn\u006fwn":1}`))
	dec.DisallowUnknownFields()
	var v small
	if err := dec.Decode(&v); err == nil || !strings.Contains(err.Error(), `"unknown"`) {
		t.Fatalf("got %v", err)
	}
}

func TestStructKeyEveryLength(t *testing.T) {
	// A struct with a key of every length around the words the keys are read by, decoded from the middle and
	// from the end of the input, through Unmarshal and through a Decoder. A key differs from the others of its
	// length in its last byte, and "x" appended or its last byte removed makes a key of another field or of none.
	var fields []reflect.StructField
	var keys []string
	lengths := []int{}
	for n := 1; n <= 40; n++ {
		lengths = append(lengths, n)
	}
	lengths = append(lengths, 47, 48, 49, 62, 63, 64, 65, 100)
	for _, n := range lengths {
		key := strings.Repeat("k", n-1) + string(rune('A'+n%26))
		keys = append(keys, key)
		fields = append(fields, reflect.StructField{
			Name: fmt.Sprintf("F%d", n),
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(fmt.Sprintf(`json:%q`, key)),
		})
	}
	typ := reflect.StructOf(fields)
	for _, key := range keys {
		for _, v := range append(keyVariants(key), key+"x", key[:len(key)-1]) {
			for _, in := range []string{
				fmt.Sprintf(`{%q:"v","zzzzzzzzzzzzzzzzzzzz":1}`, v),
				fmt.Sprintf(`{"a":1,%q:"v"}`, v),
			} {
				want := reflect.New(typ)
				if err := stdjson.Unmarshal([]byte(in), want.Interface()); err != nil {
					t.Fatal(err)
				}
				got := reflect.New(typ)
				if err := json.Unmarshal([]byte(in), got.Interface()); err != nil {
					t.Fatal(err)
				}
				stream := reflect.New(typ)
				if err := json.NewDecoder(strings.NewReader(in)).Decode(stream.Interface()); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got.Interface(), want.Interface()) || !reflect.DeepEqual(stream.Interface(), want.Interface()) {
					t.Fatalf("%s:\n     got %+v\n  stream %+v\n    want %+v", in, got.Elem(), stream.Elem(), want.Elem())
				}
			}
		}
	}
}

type middleKeys struct {
	A string `json:"prefix01_first_middle_suffix01"`
	B string `json:"prefix01_other_middle_suffix01"`
	C string `json:"prefix01_first_middle_and_more_suffix01"`
}

func TestStructKeyLongDifferInTheMiddle(t *testing.T) {
	// Keys of more than 16 bytes whose first and last eight bytes are the same match by the bytes between them.
	compareStructKeys[middleKeys](t, []string{
		`{"prefix01_first_middle_suffix01":"a"}`, `{"prefix01_other_middle_suffix01":"b"}`,
		`{"PREFIX01_OTHER_MIDDLE_SUFFIX01":"b"}`, `{"prefix01_third_middle_suffix01":"none"}`,
		`{"prefix01_first_middle_and_more_suffix01":"c"}`, `{"prefix01_first_middle_and_less_suffix01":"none"}`,
		`{"Prefix01_First_Middle_And_More_Suffix01":"c"}`,
	})
}
