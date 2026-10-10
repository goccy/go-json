package json

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/goccy/go-json/jsontext"
)

type textStringerKey int

func (k textStringerKey) String() string { return fmt.Sprint(int(k)) }

func (k textStringerKey) MarshalText() ([]byte, error) { return []byte("k" + k.String()), nil }

type plainStringerKey int

func (k plainStringerKey) String() string { return "p" }

// The keys of a map of an interface type with methods are written by their dynamic values, as encoding/json/v2
// writes them.
func TestMarshalMapOfMethodInterfaceKeys(t *testing.T) {
	for _, tt := range []struct {
		in      any
		want    string
		wantErr bool
	}{
		{in: map[fmt.Stringer]int{textStringerKey(1): 1}, want: `{"k1":1}`},
		{in: map[fmt.Stringer]int{plainStringerKey(1): 1}, want: `{"1":1}`},
		{in: map[error]int{}, want: `{}`},
		{in: map[fmt.Stringer]int{nil: 2}, wantErr: true},
	} {
		got, err := Marshal(tt.in, Deterministic(true))
		if (err != nil) != tt.wantErr || !tt.wantErr && string(got) != tt.want {
			t.Errorf("%T: got %s %v, want %s (error %v)", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}

type funcKey string

// A key in an interface or behind a pointer is named by the functions of its own type, as a key of that type is.
func TestMarshalMapKeysByFunctionsOfTheirValues(t *testing.T) {
	a := funcKey("a")
	byValue := WithMarshalers(MarshalToFunc(func(e *jsontext.Encoder, k funcKey) error {
		return e.WriteToken(jsontext.String("k-" + string(k)))
	}))
	byPointer := WithMarshalers(MarshalToFunc(func(e *jsontext.Encoder, k *funcKey) error {
		return e.WriteToken(jsontext.String("p-" + string(*k)))
	}))
	declining := WithMarshalers(MarshalToFunc(func(*jsontext.Encoder, funcKey) error { return errors.ErrUnsupported }))
	tests := []struct {
		name string
		v    any
		opt  Options
		want string
	}{
		{"in an interface", map[any]int{funcKey("a"): 1, "b": 2}, byValue, `{"b":2,"k-a":1}`},
		{"behind a pointer", map[*funcKey]int{&a: 1}, byValue, `{"k-a":1}`},
		{"a pointer in an interface", map[any]int{&a: 1}, byValue, `{"k-a":1}`},
		{"by a function of the pointer", map[any]int{funcKey("a"): 1}, byPointer, `{"p-a":1}`},
		{"declined", map[any]int{funcKey("a"): 1}, declining, `{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := Marshal(tt.v, tt.opt, Deterministic(true)); err != nil || string(got) != tt.want {
				t.Errorf("got %s %v, want %s", got, err, tt.want)
			}
		})
	}
}

// numberKey writes its number by MarshalEncode, at the place of a name for a key of a map.
type numberKey int

func (k numberKey) MarshalJSONTo(e *jsontext.Encoder) error { return MarshalEncode(e, int(k)) }

// A number which MarshalEncode writes at the place of a name is written in a string, as encoding/json/v2 writes it:
// a key of a map which writes its number, and a call for an encoder whose name is next.
func TestMarshalEncodeNumberAsName(t *testing.T) {
	if got, err := Marshal(map[numberKey]int{3: 1}); err != nil || string(got) != `{"3":1}` {
		t.Errorf("a key: got %s %v, want {\"3\":1}", got, err)
	}
	if got, err := Marshal(struct{ A numberKey }{4}); err != nil || string(got) != `{"A":4}` {
		t.Errorf("a value: got %s %v, want {\"A\":4}", got, err)
	}
	for _, v := range []any{1, uint8(2), 1.5} {
		var b bytes.Buffer
		enc := jsontext.NewEncoder(&b)
		if err := enc.WriteToken(jsontext.BeginObject); err != nil {
			t.Fatal(err)
		}
		if err := MarshalEncode(enc, v); err != nil {
			t.Errorf("%T: %v", v, err)
			continue
		}
		if err := enc.WriteToken(jsontext.Null); err != nil {
			t.Fatal(err)
		}
		if err := enc.WriteToken(jsontext.EndObject); err != nil {
			t.Fatal(err)
		}
		if want := fmt.Sprintf("{%q:null}\n", fmt.Sprint(v)); b.String() != want {
			t.Errorf("%T: got %q, want %q", v, b.String(), want)
		}
	}
}
