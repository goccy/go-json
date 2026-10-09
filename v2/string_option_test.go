package json_test

import (
	"errors"
	"testing"
	"time"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// stringOptionNumber writes its number by MarshalEncode.
type stringOptionNumber int

func (v stringOptionNumber) MarshalJSONTo(e *jsontext.Encoder) error {
	return json.MarshalEncode(e, int(v))
}

// stringOptionArray writes its number in an array.
type stringOptionArray int

func (v stringOptionArray) MarshalJSONTo(e *jsontext.Encoder) error {
	if err := e.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}
	if err := json.MarshalEncode(e, int(v)); err != nil {
		return err
	}
	return e.WriteToken(jsontext.EndArray)
}

// stringOptionReport writes whether its encoder has StringifyNumbers.
type stringOptionReport struct{}

func (stringOptionReport) MarshalJSONTo(e *jsontext.Encoder) error {
	v, _ := json.GetOption(e.Options(), json.StringifyNumbers)
	return e.WriteToken(jsontext.Bool(v))
}

type stringOptionDecliningInt int

func (stringOptionDecliningInt) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

type stringOptionDecliningStruct struct{ N int }

func (stringOptionDecliningStruct) MarshalJSONTo(*jsontext.Encoder) error {
	return errors.ErrUnsupported
}

// A method or a function of a field of the `string` option is called with StringifyNumbers, as encoding/json/v2 calls
// it: for the value itself, not in an object or an array it begins, and the value which it declines to write is
// written with the option, which fails for a value which is not a number.
func TestMarshalStringOptionOfMethods(t *testing.T) {
	type number struct {
		X stringOptionNumber `json:",string"`
	}
	type array struct {
		X stringOptionArray `json:",string"`
	}
	type report struct {
		X stringOptionReport `json:",string"`
	}
	type untagged struct{ X stringOptionReport }
	type decliningInt struct {
		X stringOptionDecliningInt `json:",string"`
	}
	type pointer struct {
		X *stringOptionNumber `json:",string"`
	}
	type function struct {
		X int8 `json:",string"`
	}
	five := stringOptionNumber(5)
	encodeInt := json.WithMarshalers(json.MarshalToFunc(func(e *jsontext.Encoder, v int8) error { return json.MarshalEncode(e, int(v)) }))
	decline := json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, int8) error { return errors.ErrUnsupported }))
	tests := []struct {
		name string
		v    any
		opts []json.Options
		want string
	}{
		{"MarshalEncode", number{5}, nil, `{"X":"5"}`},
		{"MarshalEncode in an array", array{6}, nil, `{"X":[6]}`},
		{"MarshalEncode in an array with StringifyNumbers", array{6}, []json.Options{json.StringifyNumbers(true)}, `{"X":["6"]}`},
		{"GetOption", report{}, nil, `{"X":true}`},
		{"GetOption untagged", untagged{}, nil, `{"X":false}`},
		{"declined", decliningInt{3}, nil, `{"X":"3"}`},
		{"pointer", pointer{&five}, nil, `{"X":"5"}`},
		{"nil pointer", pointer{}, nil, `{"X":null}`},
		{"function", function{3}, []json.Options{encodeInt}, `{"X":"3"}`},
		{"declined by a function", function{3}, []json.Options{decline}, `{"X":"3"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := json.Marshal(tt.v, tt.opts...); err != nil || string(got) != tt.want {
				t.Errorf("got %s %v, want %s", got, err, tt.want)
			}
		})
	}
	type decliningStruct struct {
		X stringOptionDecliningStruct `json:",string"`
	}
	if got, err := json.Marshal(decliningStruct{}); err == nil {
		t.Errorf("a declined struct: got %s, want the error of the option", got)
	}
	// a time.Duration has no representation, with the option too.
	type duration struct {
		X stringOptionDuration `json:",string"`
	}
	if got, err := json.Marshal(duration{5}); err == nil {
		t.Errorf("MarshalEncode of a time.Duration: got %s, want an error", got)
	}
	type declinedDuration struct {
		X time.Duration `json:",string"`
	}
	declineDuration := json.WithMarshalers(json.MarshalToFunc(func(*jsontext.Encoder, time.Duration) error { return errors.ErrUnsupported }))
	if got, err := json.Marshal(declinedDuration{time.Second}, declineDuration); err == nil {
		t.Errorf("a declined time.Duration: got %s, want an error", got)
	}
}

// stringOptionDuration writes a time.Duration by MarshalEncode.
type stringOptionDuration int

func (v stringOptionDuration) MarshalJSONTo(e *jsontext.Encoder) error {
	return json.MarshalEncode(e, time.Duration(v))
}
