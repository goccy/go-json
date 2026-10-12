package json

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	ierrors "github.com/goccy/go-json/internal/errors"
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

// selfPointerKey is a pointer to itself: its pointers never reach a value which names a key.
type selfPointerKey *selfPointerKey

// A key whose pointers lead back to themselves fails with a cycle, and a nil one has no name, as encoding/json/v2
// reports it: encoding them never ends or writes a name which is not a string.
func TestMarshalMapKeysPointingToThemselves(t *testing.T) {
	self := new(selfPointerKey)
	*self = selfPointerKey(self)
	viaInterface := new(any)
	*viaInterface = viaInterface
	for _, tt := range []struct {
		name  string
		in    any
		cycle bool
	}{
		{"nil", map[selfPointerKey]int{nil: 1}, false},
		{"nil behind a pointer", map[*selfPointerKey]int{new(selfPointerKey): 1}, false},
		{"itself", map[selfPointerKey]int{selfPointerKey(self): 1}, true},
		{"itself behind a pointer", map[*selfPointerKey]int{self: 1}, true},
		{"itself in an interface", map[any]int{selfPointerKey(self): 1}, true},
		{"itself through an interface", map[any]int{viaInterface: 1}, true},
	} {
		for _, deterministic := range []bool{false, true} {
			got, err := Marshal(tt.in, Deterministic(deterministic))
			if err == nil || got != nil || errors.Is(err, ierrors.ErrCycle) != tt.cycle {
				t.Errorf("%s (deterministic %v): got %q %v, want an error (a cycle %v)", tt.name, deterministic, got, err, tt.cycle)
			}
		}
	}

	// The pointers are followed one by one: a function of a type in the chain names the key before the pointers
	// lead back, and a pointer to a struct is not the one to its first field, which has the same address.
	byFunc := WithMarshalers(MarshalFunc(func(*any) ([]byte, error) { return []byte(`"R"`), nil }))
	s := &selfPointerTextKey{}
	s.P = s
	var text any = &s.P
	for _, tt := range []struct {
		name string
		in   any
		opts []Options
		want string
	}{
		{"by a function of the key", map[*any]int{viaInterface: 1}, []Options{byFunc}, `{"R":1}`},
		{"by a function inside", map[**any]int{&viaInterface: 1}, []Options{byFunc}, `{"R":1}`},
		{"a struct at the same place", map[*any]int{&text: 1}, nil, `{"S":1}`},
		{"a struct at the same place, in an interface", map[any]int{&text: 1}, nil, `{"S":1}`},
	} {
		got, err := Marshal(tt.in, tt.opts...)
		if err != nil || string(got) != tt.want {
			t.Errorf("%s: got %s %v, want %s", tt.name, got, err, tt.want)
		}
	}
}

// selfPointerTextKey names a key by MarshalText, and holds a pointer to itself at the address of the struct.
type selfPointerTextKey struct{ P any }

func (*selfPointerTextKey) MarshalText() ([]byte, error) { return []byte("S"), nil }

// The pointers and the interfaces of a key are followed one by one, not by a call each: a key of 200000 of them is
// named, where the stack would overflow.
func TestMarshalMapKeyOfManyPointers(t *testing.T) {
	deep := new(any)
	last := deep
	for range 200000 {
		next := new(any)
		*last = next
		last = next
	}
	*last = "end"
	for _, in := range []any{map[*any]int{deep: 1}, map[any]int{deep: 1}} {
		for _, deterministic := range []bool{false, true} {
			if got, err := Marshal(in, Deterministic(deterministic)); err != nil || string(got) != `{"end":1}` {
				t.Errorf("%T (deterministic %v): got %s %v, want {\"end\":1}", in, deterministic, got, err)
			}
		}
	}
}
