package json_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/goccy/go-json"
)

type selectionTestX struct {
	XA int
	XB string
	XC *selectionTestY
	XD bool
	XE float32
}

type selectionTestY struct {
	YA int
	YB string
	YC *selectionTestZ
	YD bool
	YE float32
}

type selectionTestZ struct {
	ZA string
	ZB bool
	ZC int
}

func (z *selectionTestZ) MarshalJSON(ctx context.Context) ([]byte, error) {
	type _selectionTestZ selectionTestZ
	return json.MarshalContext(ctx, (*_selectionTestZ)(z))
}

func newSelectionTestX() *selectionTestX {
	return &selectionTestX{
		XA: 1,
		XB: "xb",
		XC: &selectionTestY{
			YA: 2,
			YB: "yb",
			YC: &selectionTestZ{ZA: "za", ZB: true, ZC: 3},
			YD: true,
			YE: 4,
		},
		XD: true,
		XE: 5,
	}
}

func mustSelect(t *testing.T, fields ...json.SelectionField) *json.Selection {
	t.Helper()
	sel, err := json.Select(fields...)
	if err != nil {
		t.Fatal(err)
	}
	return sel
}

func mustParseSelection(t *testing.T, s string) *json.Selection {
	t.Helper()
	sel, err := json.ParseSelection(s)
	if err != nil {
		t.Fatal(err)
	}
	return sel
}

func marshalSelected(t *testing.T, v any, sel *json.Selection) string {
	t.Helper()
	b, err := json.MarshalWithOption(v, json.WithSelection(sel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSelectionNested(t *testing.T) {
	sel := mustSelect(t,
		json.Field("XA"),
		json.Field("XB"),
		json.Field("XC",
			json.Field("YA"),
			json.Field("YB"),
			json.Field("YC", json.Field("ZA"), json.Field("ZB")),
		),
	)
	if got, want := sel.String(), `XA XB XC { YA YB YC { ZA ZB } }`; got != want {
		t.Errorf("String: got %s, want %s", got, want)
	}
	want := `{"XA":1,"XB":"xb","XC":{"YA":2,"YB":"yb","YC":{"ZA":"za","ZB":true}}}`
	if got := marshalSelected(t, newSelectionTestX(), sel); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	// the selection in the context of MarshalContext is applied as the one of the option.
	ctx := json.ContextWithSelectionForTest(context.Background(), sel)
	b, err := json.MarshalContext(ctx, newSelectionTestX())
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != want {
		t.Errorf("MarshalContext: got %s, want %s", b, want)
	}
}

type selectionUser struct {
	ID      int                      `json:"id"`
	Name    string                   `json:"name"`
	Email   string                   `json:"email,omitempty"`
	Friends []*selectionUser         `json:"friends"`
	ByName  map[string]selectionUser `json:"byName"`
	Pair    [2]*selectionUser        `json:"pair"`
	Groups  [][]selectionUser        `json:"groups"`
	Any     any                      `json:"any"`
}

func TestSelectionOrderAndAlias(t *testing.T) {
	u := selectionUser{ID: 1, Name: "a", Email: "a@example.com"}
	for _, tt := range []struct {
		sel  string
		want string
	}{
		{`name id`, `{"name":"a","id":1}`},
		{`id displayName: name`, `{"id":1,"displayName":"a"}`},
		{`first: name second: name id`, `{"first":"a","second":"a","id":1}`},
		{`email name email`, `{"email":"a@example.com","name":"a"}`},
		{`unknown id`, `{"id":1}`},
		{`unknown`, `{}`},
		{`"id"`, `{"id":1}`},
		{`"user-id": id`, `{"user-id":1}`},
		{`{ name, id }`, `{"name":"a","id":1}`},
		{"# the name\nname\n# and the id\nid", `{"name":"a","id":1}`},
	} {
		if got := marshalSelected(t, u, mustParseSelection(t, tt.sel)); got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.sel, got, tt.want)
		}
	}
	// omitempty is kept.
	if got, want := marshalSelected(t, selectionUser{ID: 1}, mustParseSelection(t, `email id`)), `{"id":1}`; got != want {
		t.Errorf("omitempty: got %s, want %s", got, want)
	}
}

func TestSelectionOfCollections(t *testing.T) {
	u := &selectionUser{
		ID:      1,
		Name:    "a",
		Email:   "a@example.com",
		Friends: []*selectionUser{{ID: 2, Name: "b", Email: "b@example.com"}, nil, {ID: 3, Name: "c"}},
		ByName:  map[string]selectionUser{"d": {ID: 4, Name: "d", Email: "d@example.com"}},
		Pair:    [2]*selectionUser{{ID: 5, Name: "e"}, {ID: 6, Name: "f"}},
		Groups:  [][]selectionUser{{{ID: 7, Name: "g"}}, nil},
	}
	for _, tt := range []struct {
		sel  string
		want string
	}{
		{`friends { name }`, `{"friends":[{"name":"b"},null,{"name":"c"}]}`},
		{`byName { name id }`, `{"byName":{"d":{"name":"d","id":4}}}`},
		{`pair { id }`, `{"pair":[{"id":5},{"id":6}]}`},
		{`groups { name }`, `{"groups":[[{"name":"g"}],null]}`},
		{`friends { friends { id } id }`, `{"friends":[{"friends":null,"id":2},null,{"friends":null,"id":3}]}`},
		{`friends {} byName {} id {}`, `{"friends":[{},null,{}],"byName":{"d":{}},"id":1}`},
		{`{}`, `{}`},
	} {
		if got := marshalSelected(t, u, mustParseSelection(t, tt.sel)); got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.sel, got, tt.want)
		}
	}
	// the selection of a value of a list is the one of its elements.
	list := []selectionUser{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	sel := mustParseSelection(t, `name`)
	for _, v := range []any{list, &list, [2]selectionUser{list[0], list[1]}} {
		if got, want := marshalSelected(t, v, sel), `[{"name":"a"},{"name":"b"}]`; got != want {
			t.Errorf("%T: got %s, want %s", v, got, want)
		}
	}
	m := map[string]*selectionUser{"a": {ID: 1, Name: "a"}}
	if got, want := marshalSelected(t, m, sel), `{"a":{"name":"a"}}`; got != want {
		t.Errorf("map: got %s, want %s", got, want)
	}
}

func TestSelectionOfInterfaceValues(t *testing.T) {
	inner := &selectionUser{ID: 2, Name: "b", Any: &selectionUser{ID: 3, Name: "c"}}
	u := &selectionUser{ID: 1, Name: "a", Any: inner}
	// the value which an interface value holds is filtered by the selection of the place of the interface value,
	// not by the selection of the whole value.
	if got, want := marshalSelected(t, u, mustParseSelection(t, `id any { name any { id } }`)),
		`{"id":1,"any":{"name":"b","any":{"id":3}}}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	// selected whole.
	if got, want := marshalSelected(t, &selectionUser{ID: 1, Any: &selectionUser{ID: 2}}, mustParseSelection(t, `any`)),
		`{"any":{"id":2,"name":"","friends":null,"byName":null,"pair":[null,null],"groups":null,"any":null}}`; got != want {
		t.Errorf("got %s, want the whole value", got)
	}
	values := []any{&selectionUser{ID: 1, Name: "a"}, selectionUser{ID: 2, Name: "b"}, 3, "s", nil, map[string]any{"x": selectionUser{ID: 4}}}
	if got, want := marshalSelected(t, values, mustParseSelection(t, `name`)),
		`[{"name":"a"},{"name":"b"},3,"s",null,{"x":{"name":""}}]`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	anyMap := map[string]any{"a": &selectionUser{ID: 1, Name: "a"}, "b": 2}
	if got, want := marshalSelected(t, anyMap, mustParseSelection(t, `id`)), `{"a":{"id":1},"b":2}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

type selectionBase struct {
	ID      int    `json:"id"`
	Created string `json:"created"`
}

type selectionEmbedding struct {
	selectionBase
	Name string `json:"name"`
	*selectionExtra
}

type selectionExtra struct {
	Note  string `json:"note"`
	Owner string `json:"owner"`
}

func TestSelectionOfEmbeddedStructs(t *testing.T) {
	v := &selectionEmbedding{selectionBase: selectionBase{ID: 1, Created: "c"}, Name: "n", selectionExtra: &selectionExtra{Note: "x", Owner: "o"}}
	for _, tt := range []struct {
		sel  string
		want string
	}{
		{`name id`, `{"name":"n","id":1}`},
		{`created id`, `{"created":"c","id":1}`},
		{`owner key: id`, `{"owner":"o","key":1}`},
		{`name`, `{"name":"n"}`},
		{`note name created`, `{"note":"x","name":"n","created":"c"}`},
	} {
		if got := marshalSelected(t, v, mustParseSelection(t, tt.sel)); got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.sel, got, tt.want)
		}
	}
	// the fields of a struct embedded as a value are written where they are selected, and the fields of a
	// struct embedded by a pointer together.
	if got, want := marshalSelected(t, v, mustParseSelection(t, `id name created`)), `{"id":1,"name":"n","created":"c"}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if got, want := marshalSelected(t, v, mustParseSelection(t, `note name owner`)), `{"note":"x","owner":"o","name":"n"}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	v.selectionExtra = nil
	if got, want := marshalSelected(t, v, mustParseSelection(t, `note id`)), `{"id":1}`; got != want {
		t.Errorf("nil embedded: got %s, want %s", got, want)
	}
}

type selectionOuter struct {
	selectionMiddle
	Y int `json:"y"`
}

type selectionMiddle struct {
	X int `json:"x"`
	*selectionExtra
	selectionBase
}

type selectionLoopA struct {
	Name string `json:"name"`
	*selectionLoopB
}

type selectionLoopB struct {
	Note string `json:"note"`
	*selectionLoopA
}

type selectionLoopHolder struct {
	selectionLoopB
	ID int `json:"id"`
}

func TestSelectionOfNestedEmbeddedStructs(t *testing.T) {
	v := &selectionOuter{
		selectionMiddle: selectionMiddle{X: 1, selectionExtra: &selectionExtra{Note: "n", Owner: "o"}, selectionBase: selectionBase{ID: 2, Created: "c"}},
		Y:               3,
	}
	for _, tt := range []struct {
		sel  string
		want string
	}{
		// a field of a struct embedded by a pointer in a struct embedded as a value is written once, with
		// the fields of the struct of the pointer.
		{`x owner y note id`, `{"x":1,"owner":"o","note":"n","y":3,"id":2}`},
		{`created y x`, `{"created":"c","y":3,"x":1}`},
		{`note note2: note`, `{"note":"n","note2":"n"}`},
	} {
		if got := marshalSelected(t, v, mustParseSelection(t, tt.sel)); got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.sel, got, tt.want)
		}
	}
	v.selectionExtra = nil
	if got, want := marshalSelected(t, v, mustParseSelection(t, `x owner y`)), `{"x":1,"y":3}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	loop := &selectionLoopHolder{
		selectionLoopB: selectionLoopB{Note: "b", selectionLoopA: &selectionLoopA{Name: "a", selectionLoopB: &selectionLoopB{Note: "hidden"}}},
		ID:             1,
	}
	whole, err := json.Marshal(loop)
	if err != nil {
		t.Fatal(err)
	}
	sel, want := reverseSelection(t, whole)
	if got := marshalSelected(t, loop, sel); !sameObject(t, got, want) {
		t.Errorf("got %s, want the fields of %s", got, want)
	}
}

// sameObject is whether the JSON objects have the same keys and values, in any order.
func sameObject(t *testing.T, a, b string) bool {
	t.Helper()
	var x, y map[string]any
	if err := json.Unmarshal([]byte(a), &x); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(b), &y); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprint(x) == fmt.Sprint(y)
}

type selectionPtrOnly struct{ P *int }
type selectionMapOnly struct{ M map[string]int }
type selectionMarshalerOnly struct{ T selectionText }
type selectionIfaceOnly struct{ I any }
type selectionOmit struct {
	A int    `json:"a,omitempty"`
	B string `json:"b,omitempty"`
}
type selectionDeep struct {
	selectionPtrOnly
	selectionOmit
}
type selectionText struct{ s string }

func (t selectionText) MarshalText() ([]byte, error) { return []byte(t.s), nil }

type selectionEmbeddings struct {
	X int
	selectionPtrOnly
	Y string
	selectionMapOnly
	selectionMarshalerOnly
	selectionIfaceOnly
	selectionDeep `json:"deep"`
	Z             *selectionOmit
	selectionOmit
}

type selectionSingleEmbedding struct {
	selectionPtrOnly
}

type selectionSingleEmbeddingOfTwo struct {
	selectionOmit
}

// A selection of every key of a value in the reverse order writes the values of the keys as the value is written
// without a selection.
func TestSelectionOfEveryKeyInReverse(t *testing.T) {
	n := 7
	values := []any{
		selectionEmbeddings{},
		&selectionEmbeddings{
			X: 1, Y: "y",
			selectionPtrOnly:       selectionPtrOnly{P: &n},
			selectionMapOnly:       selectionMapOnly{M: map[string]int{"k": 1}},
			selectionMarshalerOnly: selectionMarshalerOnly{T: selectionText{"t"}},
			selectionIfaceOnly:     selectionIfaceOnly{I: []int{1}},
			selectionDeep:          selectionDeep{selectionPtrOnly{&n}, selectionOmit{A: 2}},
			Z:                      &selectionOmit{B: "b"},
			selectionOmit:          selectionOmit{A: 3, B: "c"},
		},
		selectionSingleEmbedding{selectionPtrOnly{&n}},
		&selectionSingleEmbedding{},
		selectionSingleEmbeddingOfTwo{selectionOmit{A: 1, B: "b"}},
		&selectionSingleEmbeddingOfTwo{selectionOmit{B: "b"}},
		&selectionEmbedding{selectionBase: selectionBase{ID: 1, Created: "c"}, Name: "n"},
	}
	for _, v := range values {
		whole, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		sel, want := reverseSelection(t, whole)
		if got := marshalSelected(t, v, sel); got != want {
			t.Errorf("%T: got %s, want %s", v, got, want)
		}
		// the elements of a list and the values of a map are selected as the value is.
		list := []any{v, v}
		if got := marshalSelected(t, list, sel); got != "["+want+","+want+"]" {
			t.Errorf("list of %T: got %s, want [%s,%s]", v, got, want, want)
		}
		if got := marshalSelected(t, map[string]any{"k": v}, sel); got != `{"k":`+want+"}" {
			t.Errorf("map of %T: got %s, want {\"k\":%s}", v, got, want)
		}
	}
}

// reverseSelection returns the selection of every key of the JSON object in data in the reverse order, and the
// object written in that order.
func reverseSelection(t *testing.T, data []byte) (*json.Selection, string) {
	t.Helper()
	keys, raws := objectKeys(t, data)
	fields := make([]json.SelectionField, 0, len(keys))
	var want bytes.Buffer
	want.WriteByte('{')
	for i := len(keys) - 1; i >= 0; i-- {
		fields = append(fields, json.Field(keys[i]))
		if want.Len() > 1 {
			want.WriteByte(',')
		}
		key, err := json.Marshal(keys[i])
		if err != nil {
			t.Fatal(err)
		}
		want.Write(key)
		want.WriteByte(':')
		want.Write(raws[i])
	}
	want.WriteByte('}')
	if len(fields) == 0 {
		fields = append(fields, json.Field("none"))
	}
	return mustSelect(t, fields...), want.String()
}

// objectKeys returns the keys of the JSON object in data and their values, in the order they are written.
func objectKeys(t *testing.T, data []byte) ([]string, []json.RawMessage) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		t.Fatalf("not an object: %s", data)
	}
	var keys []string
	var raws []json.RawMessage
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatal(err)
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, tok.(string))
		raws = append(raws, raw)
	}
	return keys, raws
}

// selectionResolver is a field which is resolved only when it is written: a resolver of GraphQL.
type selectionResolver func(ctx context.Context) (any, error)

func (r selectionResolver) MarshalJSON(ctx context.Context) ([]byte, error) {
	v, err := r(ctx)
	if err != nil {
		return nil, err
	}
	return json.MarshalContext(ctx, v)
}

type selectionPost struct {
	Title  string            `json:"title"`
	Author selectionResolver `json:"author"`
}

func TestSelectionOfMarshalerContext(t *testing.T) {
	var resolved int
	var selections []string
	author := selectionResolver(func(ctx context.Context) (any, error) {
		resolved++
		if sel := json.SelectionFromContext(ctx); sel != nil {
			selections = append(selections, sel.String())
		} else {
			selections = append(selections, "whole")
		}
		return &selectionUser{ID: 1, Name: "a", Friends: []*selectionUser{{ID: 2}}}, nil
	})
	post := &selectionPost{Title: "t", Author: author}
	if got, want := marshalSelected(t, post, mustParseSelection(t, `title`)), `{"title":"t"}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if resolved != 0 {
		t.Errorf("a field which is not selected is resolved")
	}
	if got, want := marshalSelected(t, post, mustParseSelection(t, `author { name friends { id } } title`)),
		`{"author":{"name":"a","friends":[{"id":2}]},"title":"t"}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	// a field selected whole doesn't see the selection of its parent.
	got := marshalSelected(t, []*selectionPost{post}, mustParseSelection(t, `author`))
	if want := `[{"author":{"id":1,"name":"a","friends":[{"id":2,`; !strings.HasPrefix(got, want) {
		t.Errorf("got %s, want the whole author", got)
	}
	if want := []string{"name friends { id }", "whole"}; fmt.Sprint(selections) != fmt.Sprint(want) {
		t.Errorf("selections given to the marshaler: got %q, want %q", selections, want)
	}
}

func TestSelectionWithOptions(t *testing.T) {
	u := selectionUser{ID: 1, Name: "a<b", Friends: []*selectionUser{{ID: 2, Name: "c"}}}
	sel := mustParseSelection(t, `"a<b": name friends { id }`)

	b, err := json.MarshalIndentWithOption(u, "", " ", json.WithSelection(sel))
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\n \"a\\u003cb\": \"a\\u003cb\",\n \"friends\": [\n  {\n   \"id\": 2\n  }\n ]\n}"; string(b) != want {
		t.Errorf("indent: got %s, want %s", b, want)
	}

	b, err = json.MarshalWithOption(u, json.WithSelection(sel), json.DisableHTMLEscape())
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"a<b":"a<b","friends":[{"id":2}]}`; string(b) != want {
		t.Errorf("no escape: got %s, want %s", b, want)
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).EncodeWithOption(u, json.WithSelection(sel)); err != nil {
		t.Fatal(err)
	}
	if want := "{\"a\\u003cb\":\"a\\u003cb\",\"friends\":[{\"id\":2}]}\n"; buf.String() != want {
		t.Errorf("encoder: got %s, want %s", buf.String(), want)
	}

	b, err = json.MarshalOf(u, json.WithSelection(mustParseSelection(t, `name id`)))
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\"name\":\"a\\u003cb\",\"id\":1}"; string(b) != want {
		t.Errorf("MarshalOf: got %s, want %s", b, want)
	}

	b, err = json.MarshalWithOption(u, json.WithSelection(mustParseSelection(t, `name id`)), json.OptimizeFieldOrder())
	if err != nil {
		t.Fatal(err)
	}
	if want := "{\"name\":\"a\\u003cb\",\"id\":1}"; string(b) != want {
		t.Errorf("OptimizeFieldOrder: got %s, want %s", b, want)
	}

	// a nil selection writes the whole value, and the next call of the pooled context has no selection.
	whole, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	if got := marshalSelected(t, u, nil); got != string(whole) {
		t.Errorf("nil selection: got %s, want %s", got, whole)
	}
}

func TestSelectionConcurrently(t *testing.T) {
	sels := []*json.Selection{
		mustParseSelection(t, `id friends { name }`),
		mustParseSelection(t, `name any { id }`),
	}
	values := []any{
		&selectionUser{ID: 1, Name: "a", Friends: []*selectionUser{{Name: "b"}}, Any: &selectionUser{ID: 3}},
		[]selectionUser{{ID: 1, Name: "a"}},
	}
	want := map[[2]int]string{}
	for i, sel := range sels {
		for j, v := range values {
			want[[2]int{i, j}] = marshalSelected(t, v, sel)
		}
	}
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 200; n++ {
				i, j := (g+n)%len(sels), n%len(values)
				b, err := json.MarshalWithOption(values[j], json.WithSelection(sels[i]))
				if err != nil {
					t.Error(err)
					return
				}
				if string(b) != want[[2]int{i, j}] {
					t.Errorf("got %s, want %s", b, want[[2]int{i, j}])
					return
				}
			}
		}()
	}
	wg.Wait()
}

func TestSelectBuild(t *testing.T) {
	for _, tt := range []struct {
		fields []json.SelectionField
		want   string
	}{
		{[]json.SelectionField{json.Field("a"), json.Field("a")}, `a`},
		{[]json.SelectionField{json.Field("a", json.Field("x")), json.Field("b"), json.Field("a", json.Field("y"), json.Field("x"))}, `a { x y } b`},
		{[]json.SelectionField{json.Field("a").As("b"), json.Field("a")}, `b: a a`},
		{[]json.SelectionField{json.Field("user-id").As("user id")}, `"user id": "user-id"`},
		{nil, `{}`},
		{[]json.SelectionField{json.FieldOf("a", nil)}, `a`},
		{[]json.SelectionField{json.FieldOf("a", mustSelect(t))}, `a {}`},
		{[]json.SelectionField{json.FieldOf("a", mustSelect(t)), json.Field("a", json.Field("x"))}, `a { x }`},
		{[]json.SelectionField{json.FieldOf("a", mustParseSelection(t, `x y: z`)), json.Field("a", json.Field("x"), json.Field("w"))}, `a { x y: z w }`},
	} {
		sel, err := json.Select(tt.fields...)
		if err != nil {
			t.Fatal(err)
		}
		if got := sel.String(); got != tt.want {
			t.Errorf("got %s, want %s", got, tt.want)
		}
		parsed, err := json.ParseSelection(sel.String())
		if err != nil {
			t.Fatal(err)
		}
		if parsed.String() != sel.String() {
			t.Errorf("round trip: got %s, want %s", parsed, sel)
		}
	}
	for _, fields := range [][]json.SelectionField{
		{json.Field("")},
		{json.Field("a").As("b"), json.Field("c").As("b")},
		{json.Field("a"), json.Field("a", json.Field("x"))},
		{json.Field("a", json.Field("x")), json.Field("a")},
		{json.Field("a", json.Field(""))},
		{json.FieldOf("a", mustSelect(t)), json.Field("a")},
	} {
		if sel, err := json.Select(fields...); err == nil {
			t.Errorf("no error: %s", sel)
		}
	}
}

func TestParseSelectionErrors(t *testing.T) {
	for _, s := range []string{
		``,
		`   `,
		`a {`,
		`a }`,
		`{ a } b`,
		`a: `,
		`a: {`,
		`1a`,
		`a(id: 1)`,
		`...F`,
		`a @skip`,
		`""`,
		`"a`,
		`"\x"`,
		`a: b: c`,
	} {
		if sel, err := json.ParseSelection(s); err == nil {
			t.Errorf("%q: no error: %s", s, sel)
		}
	}
}

type selectionArgs struct {
	First int
}

type selectionFeed struct {
	Posts   selectionResolver  `json:"posts"`
	Pinned  *selectionResolver `json:"pinned"`
	Counter selectionResolver  `json:"counter"`
}

func TestSelectedFieldFromContext(t *testing.T) {
	var fields []string
	record := func(ctx context.Context) {
		field := json.SelectedFieldFromContext(ctx)
		if field == nil {
			fields = append(fields, "none")
			return
		}
		fields = append(fields, fmt.Sprintf("%s:%s:%v:%v", field.Key, field.Name, field.Value, field.Sub))
	}
	posts := selectionResolver(func(ctx context.Context) (any, error) {
		record(ctx)
		args, _ := json.SelectedFieldFromContext(ctx).Value.(selectionArgs)
		list := []selectionUser{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}, {ID: 3, Name: "c"}}
		return list[:args.First], nil
	})
	pinned := selectionResolver(func(ctx context.Context) (any, error) {
		record(ctx)
		return selectionUser{ID: 9, Name: "p"}, nil
	})
	counter := selectionResolver(func(ctx context.Context) (any, error) {
		record(ctx)
		return 42, nil
	})
	feed := &selectionFeed{Posts: posts, Pinned: &pinned, Counter: counter}
	sel := mustSelect(t,
		json.Field("posts", json.Field("name")).With(selectionArgs{First: 2}).As("top"),
		json.Field("pinned", json.Field("id")),
		json.Field("counter").With("x"),
	)
	if got, want := marshalSelected(t, feed, sel), `{"top":[{"name":"a"},{"name":"b"}],"pinned":{"id":9},"counter":42}`; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if want := []string{"top:posts:{2}:name", "pinned:pinned:<nil>:id", "counter:counter:x:<nil>"}; fmt.Sprint(fields) != fmt.Sprint(want) {
		t.Errorf("fields given to the marshalers: got %q, want %q", fields, want)
	}
	// a marshaler encoded without a selection is not given a field.
	fields = nil
	if _, err := json.Marshal(feed.Counter); err != nil {
		t.Fatal(err)
	}
	if _, err := json.MarshalWithOption(struct{ C selectionResolver }{counter}, json.WithSelection(mustParseSelection(t, `C`))); err != nil {
		t.Fatal(err)
	}
	if want := []string{"none", "C:C:<nil>:<nil>"}; fmt.Sprint(fields) != fmt.Sprint(want) {
		t.Errorf("got %q, want %q", fields, want)
	}
}
