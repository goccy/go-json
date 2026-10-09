package json

import (
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
