package json_test

import (
	"context"
	stdjson "encoding/json"
	"strconv"
	"testing"
	"unsafe"

	"github.com/goccy/go-json"
)

// A marshaler whose method is on the pointer is called with the address of an addressable value, as
// encoding/json calls it: also for a type of the size of a pointer, which go-json used to copy first.

var receiverAddrs []unsafe.Pointer

type receiverOneWord struct{ p *int }

func (r *receiverOneWord) MarshalJSON() ([]byte, error) {
	receiverAddrs = append(receiverAddrs, unsafe.Pointer(r))
	return []byte(`1`), nil
}

type receiverOneWordText struct{ p *int }

func (r *receiverOneWordText) MarshalText() ([]byte, error) {
	receiverAddrs = append(receiverAddrs, unsafe.Pointer(r))
	return []byte(`t`), nil
}

type receiverOneWordContext struct{ p *int }

func (r *receiverOneWordContext) MarshalJSON(context.Context) ([]byte, error) {
	receiverAddrs = append(receiverAddrs, unsafe.Pointer(r))
	return []byte(`1`), nil
}

type receiverTwoWords struct{ p, q *int }

func (r *receiverTwoWords) MarshalJSON() ([]byte, error) {
	receiverAddrs = append(receiverAddrs, unsafe.Pointer(r))
	return []byte(`2`), nil
}

type receiverHolder struct {
	A receiverOneWord
	B receiverOneWordText
	C receiverTwoWords
	D []receiverOneWord
	E [1]receiverOneWordText
}

func (h *receiverHolder) addrs() []unsafe.Pointer {
	return []unsafe.Pointer{
		unsafe.Pointer(&h.A), unsafe.Pointer(&h.B), unsafe.Pointer(&h.C), unsafe.Pointer(&h.D[0]), unsafe.Pointer(&h.E[0]),
	}
}

func sameAddrs(a, b []unsafe.Pointer) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestEncodePointerReceiverGetsTheAddress(t *testing.T) {
	h := &receiverHolder{D: []receiverOneWord{{}}}
	want := h.addrs()

	receiverAddrs = nil
	if _, err := stdjson.Marshal(h); err != nil {
		t.Fatal(err)
	}
	if !sameAddrs(receiverAddrs, want) {
		t.Fatalf("encoding/json: got %v, want %v", receiverAddrs, want)
	}
	for _, marshal := range []struct {
		name string
		fn   func(any) ([]byte, error)
	}{
		{"Marshal", json.Marshal},
		{"MarshalIndent", func(v any) ([]byte, error) { return json.MarshalIndent(v, "", " ") }},
	} {
		receiverAddrs = nil
		if _, err := marshal.fn(h); err != nil {
			t.Fatal(err)
		}
		if !sameAddrs(receiverAddrs, want) {
			t.Errorf("%s: got %v, want %v", marshal.name, receiverAddrs, want)
		}
	}

	v := &struct{ C receiverOneWordContext }{}
	receiverAddrs = nil
	if _, err := json.MarshalContext(context.Background(), v); err != nil {
		t.Fatal(err)
	}
	if want := []unsafe.Pointer{unsafe.Pointer(&v.C)}; !sameAddrs(receiverAddrs, want) {
		t.Errorf("MarshalJSON(context.Context): got %v, want %v", receiverAddrs, want)
	}
}

// receiverState writes what the receiver holds, so that a marshaler called with a copy which differs is seen.
type receiverState struct{ p *int }

func (r *receiverState) MarshalJSON() ([]byte, error) {
	if r.p == nil {
		return []byte(`"nil"`), nil
	}
	return []byte(strconv.Itoa(*r.p)), nil
}

type receiverStateText struct{ p *int }

func (r *receiverStateText) MarshalText() ([]byte, error) {
	if r.p == nil {
		return []byte("nil"), nil
	}
	return []byte(strconv.Itoa(*r.p)), nil
}

type receiverStateFields struct {
	A receiverState
	B receiverState `json:",omitempty"`
	C receiverStateText
	D receiverStateText `json:",omitempty"`
	E []receiverState
	G map[string]receiverState
	H any
	I *receiverState
}

type receiverStateOnly struct {
	A receiverState
}

// receiverStateArray is encoded only by a pointer: go-json calls the marshaler of an element of an array which
// is not addressable, which encoding/json doesn't.
type receiverStateArray struct {
	F [1]receiverStateText
}

// The outputs of the marshalers with a pointer receiver of a type of the size of a pointer are the ones of
// encoding/json: called for an addressable value, whether the pointer it holds is nil or not, and not called for
// a value which is not addressable ( a map value, the value of an interface, a value passed to Marshal ).
func TestEncodePointerReceiverOfPointerSizedTypes(t *testing.T) {
	n := 7
	full := receiverStateFields{
		A: receiverState{&n}, B: receiverState{&n}, C: receiverStateText{&n}, D: receiverStateText{&n},
		E: []receiverState{{&n}, {}},
		G: map[string]receiverState{"k": {&n}}, H: receiverState{&n}, I: &receiverState{&n},
	}
	for _, v := range []any{
		&receiverStateFields{},
		&full,
		full,
		receiverStateFields{},
		&receiverStateOnly{},
		&receiverStateOnly{A: receiverState{&n}},
		receiverStateOnly{A: receiverState{&n}},
		[]receiverStateOnly{{A: receiverState{&n}}, {}},
		&receiverState{&n},
		receiverState{&n},
		&receiverStateText{},
		map[string]*receiverStateOnly{"k": {A: receiverState{&n}}},
		&receiverStateArray{},
		&receiverStateArray{F: [1]receiverStateText{{&n}}},
	} {
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
		wantIndent, err := stdjson.MarshalIndent(v, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		gotIndent, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if string(gotIndent) != string(wantIndent) {
			t.Errorf("%T indent: got %s, want %s", v, gotIndent, wantIndent)
		}
	}
}
