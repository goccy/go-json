package json_test

import (
	stdjson "encoding/json"
	"reflect"
	"testing"

	"github.com/goccy/go-json"
)

// A struct which is a value in many places of a type is encoded, where it is again, by a jump to one code
// of it if it is large, as a recursive struct is: the code of a type whose structs refer to each other in
// many places would otherwise be exponential to their depth, which took gigabytes and minutes to compile.

// sharedLevel00 to sharedLevel16 are a chain of structs, each of which refers to the next in four places:
// the code of the first would have 4^16 copies of the last if every struct were copied.

type sharedLevel00 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel01           `json:"a"`
	B    *sharedLevel01           `json:"b,omitempty"`
	C    []sharedLevel01          `json:"c"`
	M    map[string]sharedLevel01 `json:"m,omitempty"`
}

type sharedLevel01 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel02           `json:"a"`
	B    *sharedLevel02           `json:"b,omitempty"`
	C    []sharedLevel02          `json:"c"`
	M    map[string]sharedLevel02 `json:"m,omitempty"`
}

type sharedLevel02 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel03           `json:"a"`
	B    *sharedLevel03           `json:"b,omitempty"`
	C    []sharedLevel03          `json:"c"`
	M    map[string]sharedLevel03 `json:"m,omitempty"`
}

type sharedLevel03 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel04           `json:"a"`
	B    *sharedLevel04           `json:"b,omitempty"`
	C    []sharedLevel04          `json:"c"`
	M    map[string]sharedLevel04 `json:"m,omitempty"`
}

type sharedLevel04 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel05           `json:"a"`
	B    *sharedLevel05           `json:"b,omitempty"`
	C    []sharedLevel05          `json:"c"`
	M    map[string]sharedLevel05 `json:"m,omitempty"`
}

type sharedLevel05 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel06           `json:"a"`
	B    *sharedLevel06           `json:"b,omitempty"`
	C    []sharedLevel06          `json:"c"`
	M    map[string]sharedLevel06 `json:"m,omitempty"`
}

type sharedLevel06 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel07           `json:"a"`
	B    *sharedLevel07           `json:"b,omitempty"`
	C    []sharedLevel07          `json:"c"`
	M    map[string]sharedLevel07 `json:"m,omitempty"`
}

type sharedLevel07 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel08           `json:"a"`
	B    *sharedLevel08           `json:"b,omitempty"`
	C    []sharedLevel08          `json:"c"`
	M    map[string]sharedLevel08 `json:"m,omitempty"`
}

type sharedLevel08 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel09           `json:"a"`
	B    *sharedLevel09           `json:"b,omitempty"`
	C    []sharedLevel09          `json:"c"`
	M    map[string]sharedLevel09 `json:"m,omitempty"`
}

type sharedLevel09 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel10           `json:"a"`
	B    *sharedLevel10           `json:"b,omitempty"`
	C    []sharedLevel10          `json:"c"`
	M    map[string]sharedLevel10 `json:"m,omitempty"`
}

type sharedLevel10 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel11           `json:"a"`
	B    *sharedLevel11           `json:"b,omitempty"`
	C    []sharedLevel11          `json:"c"`
	M    map[string]sharedLevel11 `json:"m,omitempty"`
}

type sharedLevel11 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel12           `json:"a"`
	B    *sharedLevel12           `json:"b,omitempty"`
	C    []sharedLevel12          `json:"c"`
	M    map[string]sharedLevel12 `json:"m,omitempty"`
}

type sharedLevel12 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel13           `json:"a"`
	B    *sharedLevel13           `json:"b,omitempty"`
	C    []sharedLevel13          `json:"c"`
	M    map[string]sharedLevel13 `json:"m,omitempty"`
}

type sharedLevel13 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel14           `json:"a"`
	B    *sharedLevel14           `json:"b,omitempty"`
	C    []sharedLevel14          `json:"c"`
	M    map[string]sharedLevel14 `json:"m,omitempty"`
}

type sharedLevel14 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel15           `json:"a"`
	B    *sharedLevel15           `json:"b,omitempty"`
	C    []sharedLevel15          `json:"c"`
	M    map[string]sharedLevel15 `json:"m,omitempty"`
}

type sharedLevel15 struct {
	Name string `json:"name"`
	S1   int    `json:"s1,omitempty"`
	S2   float64
	A    *sharedLevel16           `json:"a"`
	B    *sharedLevel16           `json:"b,omitempty"`
	C    []sharedLevel16          `json:"c"`
	M    map[string]sharedLevel16 `json:"m,omitempty"`
}

type sharedLevel16 struct{ Leaf int }

// newSharedLevels returns a value of sharedLevel00 whose structs are set in every place of every level down
// to depth, and along the first place below it.
func newSharedLevels(depth int) *sharedLevel00 {
	v := reflect.New(reflect.TypeOf(sharedLevel00{}))
	setSharedLevel(v.Elem(), 0, depth)
	return v.Interface().(*sharedLevel00)
}

func setSharedLevel(v reflect.Value, level, depth int) {
	if v.NumField() == 1 {
		v.Field(0).SetInt(int64(level))
		return
	}
	v.Field(0).SetString(string(rune('a' + level)))
	v.Field(1).SetInt(int64(level))
	v.Field(2).SetFloat(float64(level) + 0.5)
	next := v.Field(3).Type().Elem()
	a := reflect.New(next)
	setSharedLevel(a.Elem(), level+1, depth)
	v.Field(3).Set(a)
	if level >= depth {
		return
	}
	b := reflect.New(next)
	setSharedLevel(b.Elem(), level+1, depth)
	v.Field(4).Set(b)
	c := reflect.New(next).Elem()
	setSharedLevel(c, level+1, depth)
	v.Field(5).Set(reflect.Append(reflect.MakeSlice(v.Field(5).Type(), 0, 1), c))
	m := reflect.MakeMap(v.Field(6).Type())
	e := reflect.New(next).Elem()
	setSharedLevel(e, level+1, depth)
	m.SetMapIndex(reflect.ValueOf("k"), e)
	v.Field(6).Set(m)
}

func TestEncodeSharedStructs(t *testing.T) {
	for _, v := range []any{sharedLevel00{}, &sharedLevel00{}, newSharedLevels(0), *newSharedLevels(3), []any{newSharedLevels(2), sharedLevel05{}}} {
		assertEncodedAsStd(t, v)
	}
}

type sharedBig struct {
	F000 int
	F001 int
	F002 int
	F003 int
	F004 int
	F005 int
	F006 int
	F007 int
	F008 int
	F009 int
	F010 int
	F011 int
	F012 int
	F013 int
	F014 int
	F015 int
	F016 int
	F017 int
	F018 int
	F019 int
	F020 int
	F021 int
	F022 int
	F023 int
	F024 int
	F025 int
	F026 int
	F027 int
	F028 int
	F029 int
	F030 int
	F031 int
	F032 int
	F033 int
	F034 int
	F035 int
	F036 int
	F037 int
	F038 int
	F039 int
	F040 int
	F041 int
	F042 int
	F043 int
	F044 int
	F045 int
	F046 int
	F047 int
	F048 int
	F049 int
	F050 int
	F051 int
	F052 int
	F053 int
	F054 int
	F055 int
	F056 int
	F057 int
	F058 int
	F059 int
	F060 int
	F061 int
	F062 int
	F063 int
	F064 int
	F065 int
	F066 int
	F067 int
	F068 int
	F069 int
	F070 int
	F071 int
	F072 int
	F073 int
	F074 int
	F075 int
	F076 int
	F077 int
	F078 int
	F079 int
	F080 int
	F081 int
	F082 int
	F083 int
	F084 int
	F085 int
	F086 int
	F087 int
	F088 int
	F089 int
	F090 int
	F091 int
	F092 int
	F093 int
	F094 int
	F095 int
	F096 int
	F097 int
	F098 int
	F099 int
	F100 int
	F101 int
	F102 int
	F103 int
	F104 int
	F105 int
	F106 int
	F107 int
	F108 int
	F109 int
	F110 int
	F111 int
	F112 int
	F113 int
	F114 int
	F115 int
	F116 int
	F117 int
	F118 int
	F119 int
	F120 int
	F121 int
	F122 int
	F123 int
	F124 int
	F125 int
	F126 int
	F127 int
	F128 int
	F129 int
	F130 int
	F131 int
	F132 int
	F133 int
	F134 int
	F135 int
	F136 int
	F137 int
	F138 int
	F139 int
}

// sharedEmbedder embeds sharedBig, whose fields are its own, and hides one of them: an embedded struct is
// copied even if it is large.
type sharedEmbedder struct {
	sharedBig
	F005 string
	X    int
}

type sharedHolder struct {
	A  sharedBig
	B  *sharedBig
	C  []sharedBig
	D  map[string]sharedBig
	E  sharedEmbedder
	F  *sharedEmbedder
	G  any
	H  [2]sharedBig
	P  *sharedHolder
	PS []*sharedHolder
}

func newSharedBig(base int) sharedBig {
	var v sharedBig
	rv := reflect.ValueOf(&v).Elem()
	for i := 0; i < rv.NumField(); i++ {
		rv.Field(i).SetInt(int64(base + i))
	}
	return v
}

func TestEncodeSharedNamedStructs(t *testing.T) {
	b := newSharedBig(1000)
	full := sharedHolder{
		A: newSharedBig(0),
		B: &b,
		C: []sharedBig{newSharedBig(2000), {}},
		D: map[string]sharedBig{"d": newSharedBig(3000)},
		E: sharedEmbedder{sharedBig: newSharedBig(4000), F005: "hidden", X: 1},
		F: &sharedEmbedder{sharedBig: newSharedBig(5000), X: 2},
		G: newSharedBig(6000),
		H: [2]sharedBig{newSharedBig(7000)},
	}
	full.P = &sharedHolder{B: &b, E: full.E}
	full.PS = []*sharedHolder{nil, {A: newSharedBig(8000)}}
	for _, v := range []any{sharedHolder{}, &sharedHolder{}, full, &full, []sharedHolder{full, {}}} {
		assertEncodedAsStd(t, v)
	}
}

func assertEncodedAsStd(t *testing.T, v any) {
	t.Helper()
	want, err := stdjson.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("%T: %v", v, err)
	}
	if string(got) != string(want) {
		t.Errorf("%T: got %s, want %s", v, got, want)
	}
	wantIndent, err := stdjson.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	gotIndent, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("%T: %v", v, err)
	}
	if string(gotIndent) != string(wantIndent) {
		t.Errorf("%T: indent: got %s, want %s", v, gotIndent, wantIndent)
	}
	// the order of the fields may change, but not what is written.
	ordered, err := json.MarshalWithOption(v, json.OptimizeFieldOrder())
	if err != nil {
		t.Fatalf("%T: %v", v, err)
	}
	var gotValue, wantValue any
	if err := stdjson.Unmarshal(ordered, &gotValue); err != nil {
		t.Fatalf("%T: %v: %s", v, err, ordered)
	}
	if err := stdjson.Unmarshal(want, &wantValue); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Errorf("%T: OptimizeFieldOrder: got %s, want %s", v, ordered, want)
	}
}

// A struct encoded by a jump, a recursive one or a large one in many places, is filtered by the selection of
// its own place.

type queryRecursive struct {
	Name  string
	Other string
	Child *queryRecursive
	List  []queryRecursive
}

func TestEncodeSelectionOfJumpedStructs(t *testing.T) {
	recursive := &queryRecursive{
		Name: "a", Other: "o",
		Child: &queryRecursive{Name: "b", Other: "p", Child: &queryRecursive{Name: "c", Other: "q"}},
		List:  []queryRecursive{{Name: "d", Other: "r", Child: &queryRecursive{Name: "e"}}},
	}
	b := newSharedBig(1000)
	c := []sharedBig{newSharedBig(2000)}
	holder := &sharedHolder{A: newSharedBig(0), B: &b, C: c, P: &sharedHolder{B: &b, C: c}}
	std := func(v any) string {
		b, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	for _, tt := range []struct {
		name   string
		v      any
		fields []json.SelectionField
		want   string
	}{
		{
			name:   "recursive",
			v:      recursive,
			fields: []json.SelectionField{json.Field("Name"), json.Field("Child", json.Field("Name"))},
			want:   `{"Name":"a","Child":{"Name":"b"}}`,
		},
		{
			name:   "recursive whole child",
			v:      recursive,
			fields: []json.SelectionField{json.Field("Other"), json.Field("Child"), json.Field("List")},
			want:   `{"Other":"o","Child":` + std(recursive.Child) + `,"List":` + std(recursive.List) + `}`,
		},
		{
			name: "recursive nested",
			v:    recursive,
			fields: []json.SelectionField{
				json.Field("Child", json.Field("Other"), json.Field("Child", json.Field("Name"))),
			},
			want: `{"Child":{"Other":"p","Child":{"Name":"c"}}}`,
		},
		{
			name: "shared",
			v:    holder,
			fields: []json.SelectionField{
				json.Field("A", json.Field("F001")),
				json.Field("B", json.Field("F002"), json.Field("F139")),
				json.Field("C"),
				json.Field("P", json.Field("B", json.Field("F004")), json.Field("C")),
			},
			want: `{"A":{"F001":1},"B":{"F002":1002,"F139":1139},"C":` + std(c) + `,"P":{"B":{"F004":1004},"C":` + std(c) + `}}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := json.Select(tt.fields...)
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.MarshalWithOption(tt.v, json.WithSelection(sel))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
