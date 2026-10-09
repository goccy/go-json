package json

import (
	"errors"
	"testing"

	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/jsontext"
)

type declinedMap struct{ M map[string]*declinedMap }

type declinedList struct{ Next *declinedList }

func declineList(*jsontext.Encoder, declinedList) error { return errors.ErrUnsupported }

// encodedList writes the next value by MarshalEncode.
type encodedList struct{ Next *encodedList }

func (v encodedList) MarshalJSONTo(e *jsontext.Encoder) error { return MarshalEncode(e, v.Next) }

type encodedMap struct{ M map[string]encodedMap }

func (v encodedMap) MarshalJSONTo(e *jsontext.Encoder) error { return MarshalEncode(e, v.M) }

// decliningList declines to write itself by its method, after a function declined too.
type decliningList struct{ Next *decliningList }

func (decliningList) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

// declinedHead is a value whose first field is declined too, at the same address.
type declinedHead struct {
	Head declinedList
	Next *declinedHead
}

// sharedLeaves writes the same leaf twice, by MarshalEncode.
type sharedLeaves struct {
	Next *sharedLeaves
	Leaf *encodedList
}

func (v sharedLeaves) MarshalJSONTo(e *jsontext.Encoder) error {
	if err := e.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}
	for _, x := range []any{v.Leaf, v.Leaf, v.Next} {
		if err := MarshalEncode(e, x); err != nil {
			return err
		}
	}
	return e.WriteToken(jsontext.EndArray)
}

// A cycle which passes a value declined by a function, or MarshalEncode called by a method, fails as one which
// passes the values of the encoding does, also where encoding/json/v2 overflows the stack: a method which calls
// MarshalEncode with a pointer to itself.
func TestMarshalCycleOfNestedEncodings(t *testing.T) {
	declinedM := map[string]*declinedMap{}
	declinedM["a"] = &declinedMap{M: declinedM}
	declinedL := &declinedList{}
	declinedL.Next = declinedL
	encodedL := &encodedList{}
	encodedL.Next = encodedL
	encodedM := map[string]encodedMap{}
	encodedM["a"] = encodedMap{M: encodedM}
	tests := []struct {
		name string
		v    any
		opts []Options
	}{
		{"declined map", declinedM, []Options{WithMarshalers(MarshalToFunc(func(*jsontext.Encoder, declinedMap) error { return errors.ErrUnsupported }))}},
		{"declined pointer", declinedL, []Options{WithMarshalers(MarshalToFunc(declineList))}},
		{"method pointer", encodedL, nil},
		{"method map", encodedM, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Marshal(tt.v, tt.opts...); !errors.Is(err, ierrors.ErrCycle) {
				t.Fatalf("got %v, want a cycle", err)
			}
		})
	}
}

// Values nested deeper than the encoding starts to detect cycles at, by declined values and MarshalEncode, are not
// cycles: a value encoded again by its default representation, a field at the address of its value, and a value
// written twice, not in itself.
func TestMarshalDeepNestedEncodings(t *testing.T) {
	const depth = 1500
	var declinedL *declinedList
	var encodedL *encodedList
	var decliningL *decliningList
	var head *declinedHead
	var shared *sharedLeaves
	for range depth {
		declinedL = &declinedList{Next: declinedL}
		encodedL = &encodedList{Next: encodedL}
		decliningL = &decliningList{Next: decliningL}
		head = &declinedHead{Next: head}
		shared = &sharedLeaves{Next: shared, Leaf: &encodedList{}}
	}
	declineDecliningList := MarshalToFunc(func(*jsontext.Encoder, decliningList) error { return errors.ErrUnsupported })
	declineHead := MarshalToFunc(func(*jsontext.Encoder, declinedHead) error { return errors.ErrUnsupported })
	tests := []struct {
		name string
		v    any
		opts []Options
	}{
		{"declined pointer", declinedL, []Options{WithMarshalers(MarshalToFunc(declineList))}},
		{"method pointer", encodedL, nil},
		{"declined by a function and the method", decliningL, []Options{WithMarshalers(declineDecliningList)}},
		{"declined value and first field", head, []Options{WithMarshalers(JoinMarshalers(declineHead, MarshalToFunc(declineList)))}},
		{"value written twice", shared, nil},
		{"interface of a declined value", []any{declinedL}, []Options{WithMarshalers(MarshalToFunc(declineList))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Marshal(tt.v, tt.opts...); err != nil {
				t.Fatal(err)
			}
		})
	}
}
