package json_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

type pointerFuncsValue struct{ N int }

type pointerFuncsOmit struct {
	P *int `json:",omitempty"`
	A any  `json:",omitempty"`
}

// typeOf writes the type of the value a function is given.
func typeOf(v any) ([]byte, error) { return fmt.Appendf(nil, "%q", fmt.Sprintf("%T", v)), nil }

// A pointer at the top level is marshaled as the value it points to, addressable, as encoding/json/v2 does: the
// functions of a pointer to it are never given its address, and the functions of an interface type are given the
// pointer.
func TestMarshalPointerAtTopLevelForFuncs(t *testing.T) {
	x := pointerFuncsValue{N: 1}
	tests := []struct {
		name  string
		funcs *json.Marshalers
		in    any
		want  string
	}{
		{"function of the pointer to the pointer", json.MarshalFunc(func(**pointerFuncsValue) ([]byte, error) { return []byte(`"**"`), nil }), &x, `{"N":1}`},
		{"function of the pointer", json.MarshalFunc(func(*pointerFuncsValue) ([]byte, error) { return []byte(`"*"`), nil }), &x, `"*"`},
		{"function of any", json.MarshalFunc(typeOf), &x, `"*json_test.pointerFuncsValue"`},
		{"function of any, nil pointer", json.MarshalFunc(typeOf), (*pointerFuncsValue)(nil), `null`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := json.WithMarshalers(tt.funcs)
			check := func(how string, got []byte, err error) {
				t.Helper()
				if err != nil || string(got) != tt.want {
					t.Errorf("%s: got %s, %v, want %s", how, got, err, tt.want)
				}
			}
			got, err := json.Marshal(tt.in, opts)
			check("Marshal", got, err)
			got, err = json.MarshalOf(tt.in, opts)
			check("MarshalOf", got, err)
			var b bytes.Buffer
			err = json.MarshalEncode(jsontext.NewEncoder(&b), tt.in, opts)
			check("MarshalEncode", bytes.TrimSuffix(b.Bytes(), []byte("\n")), err)
		})
	}
}

// omitempty omits a field by what its function writes, for a nil pointer or interface value too, which a function
// of the pointer to it is called for.
func TestMarshalOmitEmptyNilWrittenByFunc(t *testing.T) {
	tests := []struct {
		name  string
		funcs *json.Marshalers
		want  string
	}{
		{"functions of the pointers to the fields", json.JoinMarshalers(
			json.MarshalFunc(func(**int) ([]byte, error) { return []byte(`"P"`), nil }),
			json.MarshalFunc(func(*any) ([]byte, error) { return []byte(`"A"`), nil }),
		), `{"P":"P","A":"A"}`},
		{"functions writing empty values", json.JoinMarshalers(
			json.MarshalFunc(func(**int) ([]byte, error) { return []byte(`null`), nil }),
			json.MarshalFunc(func(*any) ([]byte, error) { return []byte(`{}`), nil }),
		), `{}`},
		{"function of a nil pointer, which is not called", json.MarshalFunc(func(*int) ([]byte, error) { return []byte(`1`), nil }), `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(pointerFuncsOmit{}, json.WithMarshalers(tt.funcs))
			if err != nil || string(got) != tt.want {
				t.Errorf("got %s, %v, want %s", got, err, tt.want)
			}
		})
	}
}
