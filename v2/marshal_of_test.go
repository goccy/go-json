package json_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

type marshalOfInner struct {
	ID   int      `json:"id"`
	Tags []string `json:"tags"`
}

type marshalOfStruct struct {
	Name  string           `json:"name"`
	Inner marshalOfInner   `json:"inner"`
	Ptr   *marshalOfInner  `json:"ptr"`
	List  []marshalOfInner `json:"list"`
	To    marshalOfTo      `json:"to"`
}

// marshalOfTo writes itself to the encoder.
type marshalOfTo struct{ N int }

func (v marshalOfTo) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.Int(int64(v.N))); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.EndArray)
}

// MarshalOf and MarshalWriteOf write what Marshal and MarshalWrite write, for the values of every kind.
func TestMarshalOf(t *testing.T) {
	inner := marshalOfInner{ID: 1, Tags: []string{"a", "b"}}
	var nilPtr *marshalOfInner
	values := []any{
		marshalOfStruct{Name: "name", Inner: inner, Ptr: &inner, List: []marshalOfInner{inner}, To: marshalOfTo{3}},
		inner, &inner, nilPtr, 1, "s", []int{1, 2}, map[string]int{"a": 1}, marshalOfTo{4}, [2]bool{true},
	}
	check := func(name string, want []byte, wantErr error, got []byte, err error) {
		t.Helper()
		if (err == nil) != (wantErr == nil) || string(got) != string(want) {
			t.Errorf("%s = %s, %v; want %s, %v", name, got, err, want, wantErr)
		}
	}
	for _, v := range values {
		want, wantErr := json.Marshal(v)
		var buf bytes.Buffer
		switch v := v.(type) {
		case marshalOfStruct:
			got, err := json.MarshalOf(v)
			check("MarshalOf", want, wantErr, got, err)
			err = json.MarshalWriteOf(&buf, v)
			check("MarshalWriteOf", want, wantErr, buf.Bytes(), err)
		case marshalOfInner:
			got, err := json.MarshalOf(v)
			check("MarshalOf", want, wantErr, got, err)
			err = json.MarshalWriteOf(&buf, v)
			check("MarshalWriteOf", want, wantErr, buf.Bytes(), err)
		case *marshalOfInner:
			got, err := json.MarshalOf(v)
			check("MarshalOf", want, wantErr, got, err)
			err = json.MarshalWriteOf(&buf, v)
			check("MarshalWriteOf", want, wantErr, buf.Bytes(), err)
		default:
			got, err := json.MarshalOf(v)
			check("MarshalOf", want, wantErr, got, err)
			err = json.MarshalWriteOf(&buf, v)
			check("MarshalWriteOf", want, wantErr, buf.Bytes(), err)
		}
	}
	// the functions of marshaling are called as for Marshal: not for a nil pointer, which is null.
	called := false
	opts := json.WithMarshalers(json.MarshalFunc(func(*marshalOfInner) ([]byte, error) {
		called = true
		return []byte(`"f"`), nil
	}))
	if got, err := json.MarshalOf(nilPtr, opts); err != nil || string(got) != "null" || called {
		t.Errorf("MarshalOf(nil) = %s, %v, called %v; want null", got, err, called)
	}
	if got, err := json.MarshalOf(&inner, opts); err != nil || string(got) != `"f"` {
		t.Errorf("MarshalOf(&inner) = %s, %v; want \"f\"", got, err)
	}
	// an error is reported as Marshal reports it.
	_, wantErr := json.Marshal("\xff")
	if _, err := json.MarshalOf("\xff"); err == nil || err.Error() != wantErr.Error() {
		t.Errorf("MarshalOf error = %v; want %v", err, wantErr)
	}
	if err := json.MarshalWriteOf(io.Discard, "\xff"); err == nil || err.Error() != wantErr.Error() {
		t.Errorf("MarshalWriteOf error = %v; want %v", err, wantErr)
	}
	var serr *jsontext.SyntacticError
	if _, err := json.MarshalOf(marshalOfStruct{Name: "\xff"}); !errors.As(err, &serr) {
		t.Errorf("MarshalOf error = %v; want a SyntacticError", err)
	}
}

func TestMarshalWriteOfAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("the runtime context, which keeps the value to reuse, is dropped from its pool at random")
	}
	inner := marshalOfInner{ID: 1, Tags: []string{"a", "b"}}
	v := marshalOfStruct{Name: "name", Inner: inner, Ptr: &inner, List: []marshalOfInner{inner}, To: marshalOfTo{3}}
	if err := json.MarshalWriteOf(io.Discard, v); err != nil {
		t.Fatal(err)
	}
	// the output is written from the buffer of the encoder, without an allocation.
	if allocs := testing.AllocsPerRun(100, func() { _ = json.MarshalWriteOf(io.Discard, v) }); allocs > 0 {
		t.Fatalf("MarshalWriteOf: %v allocations", allocs)
	}
	list := []marshalOfTo{{1}, {2}}
	if allocs := testing.AllocsPerRun(100, func() { _ = json.MarshalWriteOf(io.Discard, list) }); allocs > 0 {
		t.Fatalf("MarshalWriteOf of a slice: %v allocations", allocs)
	}
	// the result is the only allocation of MarshalOf.
	if allocs := testing.AllocsPerRun(100, func() { _, _ = json.MarshalOf(v) }); allocs > 1 {
		t.Fatalf("MarshalOf: %v allocations", allocs)
	}
}
