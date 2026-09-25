package json_test

// The cases of the reported issues of the encoder, by the number of the issue: each builds the value the issue
// describes, encodes it and compares the result byte for byte with encoding/json, unless the issue is about
// what encoding/json cannot express ( noted inline ). They are the regression tests of those reports.

import (
	"bytes"
	"context"
	stdjson "encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

// issueRecover turns a panic inside the encoder into a test failure
// instead of killing the whole test binary.
func issueRecover(t *testing.T) {
	t.Helper()
	if r := recover(); r != nil {
		t.Errorf("panic: %v", r)
	}
}

// issueCompare encodes v with both libraries and reports any difference.
func issueCompare(t *testing.T, name string, v any) {
	t.Helper()
	want, wantErr := stdjson.Marshal(v)
	got, gotErr := json.Marshal(v)
	if (wantErr == nil) != (gotErr == nil) {
		t.Errorf("%s: error mismatch: encoding/json=%v go-json=%v", name, wantErr, gotErr)
		return
	}
	if !bytes.Equal(want, got) {
		t.Errorf("%s:\n  encoding/json: %s\n  go-json:       %s", name, want, got)
	}
}

// issueCompareIndent is issueCompare for MarshalIndent.
func issueCompareIndent(t *testing.T, name string, v any, prefix, indent string) {
	t.Helper()
	want, wantErr := stdjson.MarshalIndent(v, prefix, indent)
	got, gotErr := json.MarshalIndent(v, prefix, indent)
	if (wantErr == nil) != (gotErr == nil) {
		t.Errorf("%s: error mismatch: encoding/json=%v go-json=%v", name, wantErr, gotErr)
		return
	}
	if !bytes.Equal(want, got) {
		t.Errorf("%s:\n  encoding/json: %s\n  go-json:       %s", name, want, got)
	}
}

// issueCompareEncoder is issueCompare for the streaming Encoder.
func issueCompareEncoder(t *testing.T, name string, v any) {
	t.Helper()
	var want, got bytes.Buffer
	wantErr := stdjson.NewEncoder(&want).Encode(v)
	gotErr := json.NewEncoder(&got).Encode(v)
	if (wantErr == nil) != (gotErr == nil) {
		t.Errorf("%s: error mismatch: encoding/json=%v go-json=%v", name, wantErr, gotErr)
		return
	}
	if !bytes.Equal(want.Bytes(), got.Bytes()) {
		t.Errorf("%s:\n  encoding/json: %q\n  go-json:       %q", name, want.String(), got.String())
	}
}

// ---------------------------------------------------------------- 401

type issue401Set map[string]struct{}

func (s issue401Set) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	if len(s) == 0 {
		return []byte("[]"), nil
	}
	// The issue iterates the map; a single key keeps the output deterministic.
	b := []byte{'['}
	for symbol := range s {
		b = append(b, '"')
		b = append(b, symbol...)
		b = append(b, '"', ',')
	}
	b[len(b)-1] = ']'
	return b, nil
}

type issue401PanicType map[uint64]struct{}

func (w issue401PanicType) MarshalJSON() ([]byte, error) {
	_ = w[1]
	return []byte{'1'}, nil
}

func TestIssue401(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "map of custom map with MarshalJSON", map[string]issue401Set{"foo": {"bar": {}}})
	issueCompare(t, "map[int] of map type reading itself in MarshalJSON", map[int]issue401PanicType{1: {15124: struct{}{}}})
}

// ---------------------------------------------------------------- 418

type issue418Embedding struct{}

func (n *issue418Embedding) MarshalJSON() ([]byte, error) {
	return []byte(`"this is MarshalJSON"`), nil
}

type issue418Parent struct {
	issue418Embedding
}

type issue418PtrParent struct{}

func (n *issue418PtrParent) MarshalJSON() ([]byte, error) {
	return []byte(`"this is MarshalJSON"`), nil
}

func TestIssue418(t *testing.T) {
	defer issueRecover(t)
	x := &issue418Parent{}
	issueCompareEncoder(t, "pointer-receiver MarshalJSON on embedded value, **Parent", &x)
	y := &issue418PtrParent{}
	issueCompareEncoder(t, "pointer-receiver MarshalJSON on the struct, **Parent", &y)
}

// ---------------------------------------------------------------- 449

func TestIssue449(t *testing.T) {
	defer issueRecover(t)
	var key any = "x"
	obj := map[any]any{key: "xxxxx"}
	// encoding/json refuses map[interface{}]T as well (UnsupportedTypeError),
	// so the expected outcome is "both return an error".
	_, wantErr := stdjson.Marshal(obj)
	_, gotErr := json.Marshal(obj)
	if (wantErr == nil) != (gotErr == nil) {
		t.Errorf("error mismatch: encoding/json=%v go-json=%v", wantErr, gotErr)
	}
}

// ---------------------------------------------------------------- 450

type issue450Embedded struct {
	Embedded string `json:"Embedded"`
	Other    int    `json:"Other"`
}

type issue450Struct struct {
	issue450Embedded
	Name string `json:"Name"`
}

func TestIssue450(t *testing.T) {
	defer issueRecover(t)
	// The issue uses a type literally named Embedded embedding a field named
	// Embedded; the local type name has to differ, so the embedded field is
	// named after the local type and the inner field keeps the issue's name.
	type Embedded struct {
		Embedded string `json:"Embedded"`
		Other    int    `json:"Other"`
	}
	type Struct struct {
		Embedded
		Name string `json:"Name"`
	}
	issueCompare(t, "embedded struct whose field has the same name (local types)",
		Struct{Embedded: Embedded{Embedded: "inside", Other: 111}, Name: "outside"})
	issueCompare(t, "embedded struct whose field has the same name (file scope)",
		issue450Struct{issue450Embedded: issue450Embedded{Embedded: "inside", Other: 111}, Name: "outside"})
}

// ---------------------------------------------------------------- 458

type issue458Empty [0]int

func (issue458Empty) MarshalJSON() ([]byte, error) { return []byte(`"empty-array"`), nil }

type issue458Two [2]int

func (issue458Two) MarshalJSON() ([]byte, error) { return []byte(`"two-array"`), nil }

type issue458Holder struct {
	E issue458Empty  `json:"e,omitempty"`
	P *issue458Empty `json:"p,omitempty"`
	T issue458Two    `json:"t,omitempty"`
}

func TestIssue458(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "omitempty on array types with MarshalJSON", issue458Holder{})
	issueCompare(t, "omitempty on array types with MarshalJSON, pointer set", issue458Holder{P: &issue458Empty{}})
}

// ---------------------------------------------------------------- 460

type Issue460B struct {
	Bar string `json:"bar"`
}

type issue460A struct {
	Foo        string `json:"foo"`
	*Issue460B `json:",inline,omitempty"`
}

type issue460AOmitOnly struct {
	Foo        string `json:"foo"`
	*Issue460B `json:",omitempty"`
}

func TestIssue460(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "embedded pointer with inline,omitempty", issue460A{Foo: "fff", Issue460B: &Issue460B{Bar: "bbb"}})
	issueCompare(t, "embedded pointer with inline,omitempty, nil", issue460A{Foo: "fff"})
	issueCompare(t, "embedded pointer with omitempty", issue460AOmitOnly{Foo: "fff", Issue460B: &Issue460B{Bar: "bbb"}})
	issueCompare(t, "embedded pointer with omitempty, nil", issue460AOmitOnly{Foo: "fff"})
}

// ---------------------------------------------------------------- 468

type issue468Outer struct {
	*Issue468Inner
}

type Issue468Inner struct {
	Inner *Issue468Inner `json:"inner"`
}

func TestIssue468(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "embedded pointer to a recursive struct, nil", &issue468Outer{})
	issueCompare(t, "embedded pointer to a recursive struct, set",
		&issue468Outer{Issue468Inner: &Issue468Inner{Inner: &Issue468Inner{}}})
}

// ---------------------------------------------------------------- 486

type issue486Structure struct {
	Id int `json:"id"`
	*issue486Structure
}

type issue486AlternativeStructure struct {
	Id int `json:"id"`
	*Issue486Structure2
}

type Issue486Structure2 struct {
	Id          int                          `json:"id"`
	Alternative issue486AlternativeStructure `json:"alternative"`
}

func TestIssue486(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "struct embedding a nil pointer to itself", issue486Structure{Id: 99})
	issueCompare(t, "struct embedding a nil pointer to the enclosing struct",
		Issue486Structure2{Id: 99, Alternative: issue486AlternativeStructure{Id: 123}})
}

// ---------------------------------------------------------------- 488

type issue488StringAlias string

func (d *issue488StringAlias) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", *d)), nil
}

type issue488WithOmit struct {
	A issue488StringAlias `json:"a,omitempty"`
}

func TestIssue488(t *testing.T) {
	defer issueRecover(t)
	var v issue488WithOmit
	issueCompare(t, "omitempty on a type with pointer-receiver MarshalJSON, empty", &v)
	v.A = "x"
	issueCompare(t, "omitempty on a type with pointer-receiver MarshalJSON, set", &v)
}

// ---------------------------------------------------------------- 499

type issue499Foo struct{}

type issue499CtxKey struct{}

func (issue499Foo) MarshalJSON(ctx context.Context) ([]byte, error) {
	if ctx != nil {
		if val := ctx.Value(issue499CtxKey{}); val != nil {
			return []byte(`"` + val.(string) + `"`), nil
		}
	}
	return []byte(`"foo"`), nil
}

func TestIssue499(t *testing.T) {
	defer issueRecover(t)
	// MarshalerContext is a go-json extension, so encoding/json cannot serve
	// as the oracle; the test asserts what the issue asks for: the context
	// of a MarshalContext call must not leak into a following Marshal call.
	foo := issue499Foo{}
	ctx := context.WithValue(context.Background(), issue499CtxKey{}, "bar")
	out, err := json.MarshalContext(ctx, foo)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `"bar"` {
		t.Errorf(`MarshalContext: expected "bar", got %s`, out)
	}
	for i := 0; i < 8; i++ {
		out, err = json.Marshal(foo)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != `"foo"` {
			t.Errorf(`Marshal after MarshalContext: expected "foo", got %s (leaked context)`, out)
			return
		}
	}
}

// ---------------------------------------------------------------- 503

type issue503RootInt struct {
	Child *issue503ChildInt `json:"c,omitempty"`
}

type issue503ChildInt struct {
	Value int `json:"v"`
}

type issue503RootBool struct {
	Child *issue503ChildBool `json:"c,omitempty"`
}

type issue503ChildBool struct {
	Value bool `json:"v"`
}

type issue503RootString struct {
	Child *issue503ChildString `json:"c,omitempty"`
}

type issue503ChildString struct {
	Value string `json:"v"`
}

func TestIssue503(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "omitempty pointer to struct whose first field is 0", issue503RootInt{Child: &issue503ChildInt{Value: 0}})
	issueCompare(t, "omitempty pointer to struct whose first field is 1", issue503RootInt{Child: &issue503ChildInt{Value: 1}})
	issueCompare(t, "omitempty pointer to struct whose first field is false", issue503RootBool{Child: &issue503ChildBool{Value: false}})
	issueCompare(t, "omitempty pointer to struct whose first field is true", issue503RootBool{Child: &issue503ChildBool{Value: true}})
	issueCompare(t, `omitempty pointer to struct whose first field is ""`, issue503RootString{Child: &issue503ChildString{Value: ""}})
	issueCompare(t, `omitempty pointer to struct whose first field is "a"`, issue503RootString{Child: &issue503ChildString{Value: "a"}})
}

// ---------------------------------------------------------------- 507

type issue507Config struct {
	Steps []struct {
		Config stdjson.RawMessage `json:"config"`
	}
}

func TestIssue507(t *testing.T) {
	defer issueRecover(t)
	// The issue's raw data holds a literal U+2028 before the "<".
	rawData := []byte("{\"steps\": [{\"config\":{\"message\": \"Fail \u2028<\"}}]}")
	var v issue507Config
	if err := stdjson.Unmarshal(rawData, &v); err != nil {
		t.Fatal(err)
	}
	issueCompare(t, "RawMessage with U+2028 and < (HTML escape + compaction)", v)

	// The same content through go-json's own RawMessage type.
	type config struct {
		Steps []struct {
			Config json.RawMessage `json:"config"`
		}
	}
	var v2 config
	if err := stdjson.Unmarshal(rawData, &v2); err != nil {
		t.Fatal(err)
	}
	issueCompare(t, "go-json RawMessage with U+2028 and <", v2)
}

// ---------------------------------------------------------------- 512

type issue512Common struct {
	Domains []string `json:"domains,omitempty" binding:"required,gte=1,lte=3,unique,dive,required"`
}

type issue512App struct {
	ID   uint   `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	issue512Common
}

type issue512Data struct {
	issue512App
}

type issue512AppGetResponse struct {
	TotalCount int            `json:"totalCount"`
	Data       []issue512Data `json:"data,omitempty"`
}

type issue512App2 struct {
	issue512Common
	ID   uint   `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type issue512Data2 struct {
	issue512App2
}

type issue512AppGetResponse2 struct {
	TotalCount int             `json:"totalCount"`
	Data       []issue512Data2 `json:"data,omitempty"`
}

func TestIssue512(t *testing.T) {
	defer issueRecover(t)
	resp := issue512AppGetResponse{TotalCount: 1, Data: make([]issue512Data, 1)}
	resp.Data[0] = issue512Data{issue512App: issue512App{ID: 1, Type: "custom"}}
	issueCompare(t, "doubly embedded struct, omitempty slice last", resp)

	resp2 := issue512AppGetResponse2{TotalCount: 1, Data: make([]issue512Data2, 1)}
	resp2.Data[0] = issue512Data2{issue512App2: issue512App2{ID: 1, Type: "custom"}}
	issueCompare(t, "doubly embedded struct, omitempty slice first", resp2)
}

// ---------------------------------------------------------------- 513

type Issue513Customer struct {
	TKID string               `gorm:"primarykey;not null;type:varchar(255);"`
	IDs  []*Issue513IDMapping `gorm:"foreignKey:TKID;"`
}

type Issue513IDMapping struct {
	TKID     string `gorm:"comment:TKID;index:idx_tk_id;not null;type:varchar(255);"`
	Customer *Issue513Customer
}

func TestIssue513(t *testing.T) {
	defer issueRecover(t)
	type MyCustomer struct {
		Issue513Customer
		MyField string
	}
	list := []MyCustomer{{Issue513Customer: Issue513Customer{TKID: "1"}, MyField: "111"}}
	issueCompare(t, "slice of struct embedding a mutually recursive struct", list)
	issueCompare(t, "the embedded struct alone", list[0].Issue513Customer)
}

// ---------------------------------------------------------------- 519

type issue519Body struct {
	Payload *issue519Detail `json:"p,omitempty"`
}

type issue519Detail struct {
	I issue519Item `json:"i"`
}

type issue519Item struct {
	A string `json:"a"`
	B string `json:"b,omitempty"`
}

func TestIssue519(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "omitempty pointer to struct holding a struct with omitempty string",
		issue519Body{Payload: &issue519Detail{I: issue519Item{A: "a", B: "b"}}})
}

// ---------------------------------------------------------------- 523

func TestIssue523(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "[]*time.Time{nil}", []*time.Time{nil})
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	issueCompare(t, "[]*time.Time{nil, &t}", []*time.Time{nil, &now})
}

// ---------------------------------------------------------------- 526

type issue526CondFn byte
type issue526Args map[string]any

type issue526Rule struct {
	Id        int               `json:"id"`
	Disabled  bool              `json:"disabled,omitempty"`
	Name      string            `json:"name,omitempty"`
	Condition issue526Condition `json:"condition,omitempty"`
}

type issue526Condition struct {
	Id     int                 `json:"-"`
	Fn     issue526CondFn      `json:"fn,omitempty"`
	Args   issue526Args        `json:"args,omitempty"`
	Nested []issue526Condition `json:"nested,omitempty"`
}

func TestIssue526(t *testing.T) {
	defer issueRecover(t)
	rule := issue526Rule{
		Condition: issue526Condition{
			Nested: []issue526Condition{{Args: issue526Args{"Value": 111}}},
		},
	}
	issueCompare(t, `recursive struct with a json:"-" field and omitempty map`, rule)
}

// ---------------------------------------------------------------- 537

// issue537List mirrors github.com/kaptinlin/jsonschema.List (the package is
// not available here): a recursive struct with two omitempty maps.
type issue537List struct {
	Valid            bool              `json:"valid"`
	EvaluationPath   string            `json:"evaluationPath"`
	SchemaLocation   string            `json:"schemaLocation"`
	InstanceLocation string            `json:"instanceLocation"`
	Annotations      map[string]any    `json:"annotations,omitempty"`
	Errors           map[string]string `json:"errors,omitempty"`
	Details          []issue537List    `json:"details,omitempty"`
}

const issue537Payload = `{
  "valid": false,
  "evaluationPath": "",
  "schemaLocation": "",
  "instanceLocation": "",
  "annotations": {
    "description": "Blueprint type prototype\n\nThis is just a brief example of a common blueprint structure. Just few fields\nwere selected to demonstrate the JSON schema."
  },
  "errors": {
    "properties": "Properties 'registration', 'network' do not match their schemas",
    "required": "Required properties 'registration', 'network' are missing"
  },
  "details": [
    {
      "valid": false,
      "evaluationPath": "/properties/registration",
      "schemaLocation": "https://github.com/lzap/common-blueprint-example/blueprint#/properties/registration",
      "instanceLocation": "/registration",
      "annotations": {
        "description": "Registration details"
      },
      "errors": {
        "type": "Value is null but should be object"
      }
    },
    {
      "valid": false,
      "evaluationPath": "/properties/network",
      "schemaLocation": "https://github.com/lzap/common-blueprint-example/blueprint#/properties/network",
      "instanceLocation": "/network",
      "annotations": {
        "description": "Networking details"
      },
      "errors": {
        "type": "Value is null but should be object"
      }
    },
    {
      "valid": true,
      "evaluationPath": "/properties/name",
      "schemaLocation": "https://github.com/lzap/common-blueprint-example/blueprint#/properties/name",
      "instanceLocation": "/name",
      "annotations": {
        "description": "Name of the blueprint"
      }
    }
  ]
}`

func TestIssue537(t *testing.T) {
	defer issueRecover(t)
	var list issue537List
	if err := stdjson.Unmarshal([]byte(issue537Payload), &list); err != nil {
		t.Fatal(err)
	}
	issueCompareIndent(t, "MarshalIndent of a recursive struct with omitempty maps (jsonschema.List shape)", list, "", "  ")
	issueCompare(t, "Marshal of the same value", list)
}

// ---------------------------------------------------------------- 541

type issue541BotRetPtrID struct {
	Id   *string `json:"id"`
	Name string  `json:"name"`
}

type issue541RetPtr struct {
	Bot *issue541BotRetPtrID `json:"bot"`
}

type issue541BotRet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type issue541Ret struct {
	Bot issue541BotRet `json:"bot"`
}

func TestIssue541(t *testing.T) {
	defer issueRecover(t)
	// The issue only shows the types (panic in appendNormalizedHTMLString);
	// the values are reconstructed: nil / set pointer, empty / filled strings.
	id := "id-1"
	issueCompare(t, "pointer struct field, nil", issue541RetPtr{})
	issueCompare(t, "pointer struct field, set, nil id", issue541RetPtr{Bot: &issue541BotRetPtrID{Name: "n"}})
	issueCompare(t, "pointer struct field, set, id set", issue541RetPtr{Bot: &issue541BotRetPtrID{Id: &id, Name: "n"}})
	issueCompare(t, "value struct field, zero", issue541Ret{})
	issueCompare(t, "value struct field, set", issue541Ret{Bot: issue541BotRet{Id: "id", Name: "n"}})
	issueCompare(t, "pointer to the anonymous struct", &issue541Ret{Bot: issue541BotRet{Id: "id", Name: "n"}})
}

// ---------------------------------------------------------------- 543

type issue543Map map[string]struct{}

type issue543MapMap map[string]issue543Map

func (m issue543Map) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return stdjson.Marshal(keys)
}

func TestIssue543(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "map of map type with MarshalJSON", issue543MapMap{"test": issue543Map{"test": {}}})
}

// ---------------------------------------------------------------- 546

type issue546Device struct {
	Alerts [][]any
	Items  []issue546Device
	ID     int64
}

const issue546Data = `
    {
        "items": [
            {
                "alerts": [[2]],
                "id": 2
            }
        ],
        "id": 1
    }
`

func TestIssue546(t *testing.T) {
	defer issueRecover(t)
	var d issue546Device
	if err := stdjson.Unmarshal([]byte(issue546Data), &d); err != nil {
		t.Fatal(err)
	}
	issueCompare(t, "recursive struct with [][]any first", d)
}

// ---------------------------------------------------------------- 552

type issue552DBObject struct {
	ID int64 `json:"id,omitempty"`
}

type Issue552Item struct {
	issue552DBObject
}

type issue552Request struct {
	*Issue552Item
}

type issue552StudyLabel struct {
	HasPartList   bool                `json:"has_part_list"`
	PartItems     []*issue552ResLabel `json:"part_items"`
	ResourceItems []*issue552Resource `json:"resource_items"`
}

type issue552Resource struct {
	Name string `json:"name"`
}

type issue552ResLabel struct {
	ClassId     int    `json:"class_id"`
	ClassName   string `json:"class_name"`
	PercentRate int    `json:"percent_rate"`
	IsCompleted bool   `json:"is_completed"`

	issue552StudyLabel
}

func TestIssue552(t *testing.T) {
	defer issueRecover(t)
	var req issue552Request
	if err := stdjson.Unmarshal([]byte(`{"id": 1}`), &req); err != nil {
		t.Fatal(err)
	}
	issueCompareEncoder(t, "embedded pointer to struct embedding a struct with omitempty int", req)
	issueCompare(t, "same value through Marshal", req)

	label := &issue552ResLabel{
		ClassId:   150301,
		ClassName: "PreA1",
		issue552StudyLabel: issue552StudyLabel{
			PartItems:     []*issue552ResLabel{},
			ResourceItems: []*issue552Resource{{Name: "r"}},
		},
	}
	issueCompare(t, "embedded recursive struct after plain fields", label)
	issueCompare(t, "slice of the same", []*issue552ResLabel{label, {issue552StudyLabel: issue552StudyLabel{PartItems: []*issue552ResLabel{label}}}})
}

// ---------------------------------------------------------------- 554

type issue554Grandpa struct{ issue554Father }

type issue554Father struct {
	issue554Son
	A []string `json:"a,omitempty"`
}

type issue554Son struct {
	B string `json:"b"`
	C string `json:"c"`
}

func TestIssue554(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "doubly embedded struct, empty slice with omitempty last", &issue554Grandpa{})
	issueCompare(t, "same by value", issue554Grandpa{})
}

// ---------------------------------------------------------------- 559

type issue559Node struct {
	Label    string         `json:"label"`
	Value    string         `json:"value"`
	Meta     map[string]any `json:"meta,omitempty"`
	Children []issue559Node `json:"children,omitempty"`
}

func TestIssue559(t *testing.T) {
	defer issueRecover(t)
	tree := []issue559Node{
		{
			Label: "Parent",
			Value: "p1",
			Meta:  map[string]any{"code": "parent", "desc": "Parent node"},
			Children: []issue559Node{
				{
					Label: "Child1",
					Value: "c1",
					Meta:  map[string]any{"code": "child1", "desc": "First child"},
				},
			},
		},
	}
	issueCompare(t, "recursive struct with map[string]any in parent and child", tree)
}

// ---------------------------------------------------------------- 576

type issue576ErrWithOmitempty struct {
	Error string `json:"error,omitempty"`
}

type issue576MyResult struct {
	issue576ErrWithOmitempty
	Data string `json:"data"`
}

type issue576Result struct {
	issue576MyResult
}

type issue576ErrWithoutOmitempty struct {
	Error string `json:"error"`
}

type issue576MyResult2 struct {
	issue576ErrWithoutOmitempty
	Data string `json:"data"`
}

type issue576Result2 struct {
	issue576MyResult2
}

func TestIssue576(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "doubly embedded struct whose first field is omitempty and empty",
		issue576Result{issue576MyResult{Data: "hi"}})
	issueCompare(t, "doubly embedded struct whose first field is not omitempty",
		issue576Result2{issue576MyResult2{Data: "hi"}})
}

// ---------------------------------------------------------------- 581

type issue581Reason struct {
	Code string `json:"reasonCode,omitempty"`
}

type issue581Cov struct {
	Type string `json:"type"`
}

type issue581Inner struct {
	issue581Cov
	R *issue581Reason `json:"reason,omitempty"`
}

type issue581Outer struct {
	issue581Inner
}

func TestIssue581(t *testing.T) {
	defer issueRecover(t)
	issueCompare(t, "embedded struct with embedded value struct followed by omitempty nil pointer", issue581Outer{})
	issueCompare(t, "same with the pointer set", issue581Outer{issue581Inner{R: &issue581Reason{Code: "c"}}})
}
