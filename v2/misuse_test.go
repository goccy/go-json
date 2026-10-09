package json

import (
	"bytes"
	"testing"

	"github.com/goccy/go-json/jsontext"
)

// twoTopValues writes a number and an object, two values at the top level.
type twoTopValues struct{}

func (twoTopValues) MarshalJSONTo(e *jsontext.Encoder) error {
	if err := e.WriteToken(jsontext.Int(1)); err != nil {
		return err
	}
	return e.WriteValue(jsontext.Value(`{"a":[1]}`))
}

// pastOwnLevel writes an object with a raw value which fails in it, ignoring the errors, and closes the level its
// encoder is attached in after it.
type pastOwnLevel struct{}

func (pastOwnLevel) MarshalJSONTo(e *jsontext.Encoder) error {
	_ = e.WriteToken(jsontext.BeginObject)
	_ = e.WriteToken(jsontext.String("b"))
	_ = e.WriteValue(jsontext.Value(`[`))
	_ = e.WriteToken(jsontext.Int(7))
	_ = e.WriteToken(jsontext.EndObject)
	return e.WriteToken(jsontext.EndObject)
}

// closesItsPlace writes a value, then ends the array which its value is in, which is not its own, and writes on.
type closesItsPlace struct{}

func (closesItsPlace) MarshalJSONTo(e *jsontext.Encoder) error {
	_ = e.WriteValue(jsontext.Value(`1`))
	_ = e.WriteToken(jsontext.EndArray)
	return e.WriteValue(jsontext.Value(`}`))
}

// A method which misuses its encoder makes the call fail, as encoding/json/v2 does, without a panic.
func TestMarshalMethodMisuseFails(t *testing.T) {
	for name, marshal := range map[string]func() error{
		"two values":          func() error { _, err := Marshal(twoTopValues{}); return err },
		"two values, pointer": func() error { _, err := Marshal(&twoTopValues{}); return err },
		"two values, encoder": func() error { return MarshalEncode(jsontext.NewEncoder(new(bytes.Buffer)), twoTopValues{}) },
		"past its level":      func() error { _, err := Marshal(struct{ K pastOwnLevel }{}); return err },
		"its place":           func() error { _, err := Marshal(map[string]any{"K": []any{closesItsPlace{}}}); return err },
		"past its level, dups": func() error {
			_, err := Marshal(struct{ K pastOwnLevel }{}, jsontext.AllowDuplicateNames(true))
			return err
		},
	} {
		if err := marshal(); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

// hidesFailure writes an array with the value of a call of MarshalEncode which fails, whose error it doesn't return:
// it writes null in its place if fallback is set, or nothing.
type hidesFailure struct {
	value    any
	fallback bool
}

func (v hidesFailure) MarshalJSONTo(e *jsontext.Encoder) error {
	if err := e.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}
	if err := MarshalEncode(e, v.value); err != nil && v.fallback {
		if err := e.WriteToken(jsontext.Null); err != nil {
			return err
		}
	}
	return e.WriteToken(jsontext.EndArray)
}

// optionAtPlace calls MarshalEncode with an option which the encoder can't change at its place, which fails.
type optionAtPlace struct{}

func (optionAtPlace) MarshalJSONTo(e *jsontext.Encoder) error {
	return MarshalEncode(e, 1, jsontext.Multiline(true))
}

// A call of MarshalEncode which fails within a method fails the value of the method, whether the method returns the
// error or not: its output is valid JSON, without the value which failed.
func TestMarshalMethodHidingFailure(t *testing.T) {
	fails := make(chan int)
	for _, v := range []any{
		hidesFailure{fails, false},
		hidesFailure{fails, true},
		hidesFailure{[]any{1, fails}, true},
		hidesFailure{hidesFailure{fails, true}, true},
		hidesFailure{optionAtPlace{}, true},
	} {
		for _, place := range []any{v, []any{v}, struct{ A any }{v}, map[string]any{"a": v}} {
			if got, err := Marshal(place); err == nil {
				t.Errorf("%T %+v: no error, %s", place, place, got)
			}
		}
	}
	if _, err := Marshal(hidesFailure{1, false}); err != nil {
		t.Errorf("a value which doesn't fail: %v", err)
	}
}
