package json

import (
	stdjson "encoding/json"
	"errors"
	"testing"
	"time"
)

type numberFields struct {
	N  stdjson.Number  `json:"n,string"`
	P  *stdjson.Number `json:"p,string"`
	NP *stdjson.Number `json:"np,string"`
	O  stdjson.Number  `json:"o,omitempty"`
	OP *stdjson.Number `json:"op,omitempty"`
	V  stdjson.Number  `json:"v"`
	VP *stdjson.Number `json:"vp"`
}

// A json.Number is written as a number, as encoding/json/v2 writes it: in a string at the place of a name, for the
// `string` option and for StringifyNumbers, and 0 if it is "", which omitempty doesn't omit.
func TestMarshalNumber(t *testing.T) {
	five, empty := stdjson.Number("5"), stdjson.Number("")
	fields := numberFields{N: "12", P: &five, OP: &empty, V: "1.5", VP: &five}
	tests := []struct {
		name string
		v    any
		opts []Options
		want string
	}{
		{"keys", map[stdjson.Number]int{"1.5": 1, "10": 2, "": 3}, []Options{Deterministic(true)}, `{"0":3,"1.5":1,"10":2}`},
		{"keys in an interface", map[any]int{stdjson.Number("3"): 1, stdjson.Number(""): 2}, []Options{Deterministic(true)},
			`{"0":2,"3":1}`},
		{"pointer keys", map[*stdjson.Number]int{&empty: 1}, nil, `{"0":1}`},
		{"fields", fields, nil, `{"n":"12","p":"5","np":null,"o":0,"op":0,"v":1.5,"vp":5}`},
		{"fields with StringifyNumbers", fields, []Options{StringifyNumbers(true)},
			`{"n":"12","p":"5","np":null,"o":"0","op":"0","v":"1.5","vp":"5"}`},
		{"values with StringifyNumbers", []any{stdjson.Number("1e2"), []stdjson.Number{"2"}, map[string]stdjson.Number{"a": "3"}},
			[]Options{StringifyNumbers(true)}, `["1e2",["2"],{"a":"3"}]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.v, tt.opts...)
			if err != nil || string(got) != tt.want {
				t.Fatalf("got %s %v, want %s", got, err, tt.want)
			}
		})
	}
	invalid := stdjson.Number("x")
	for _, v := range []any{map[stdjson.Number]int{"x": 1}, map[any]int{invalid: 1}, map[*stdjson.Number]int{&invalid: 1},
		map[string]stdjson.Number{"a": "x"}} {
		if _, err := Marshal(v); err == nil {
			t.Errorf("%v: got no error for a json.Number which is not a number", v)
		}
	}
}

// A key of time.Duration fails as a value of it does: it has no representation of its own.
func TestMarshalDurationKey(t *testing.T) {
	for _, v := range []any{map[time.Duration]int{time.Second: 1}, map[any]int{time.Second: 1}} {
		var serr *SemanticError
		if _, err := Marshal(v); !errors.As(err, &serr) {
			t.Errorf("%T: got %v, want an error of no representation", v, err)
		}
	}
	if got, err := Marshal(map[time.Duration]int{}); err != nil || string(got) != "{}" {
		t.Errorf("an empty map: got %s %v, want {}", got, err)
	}
}
