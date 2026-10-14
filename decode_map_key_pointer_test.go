package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

type mapKeyPointerNamed *int

type mapKeyPointerBool bool

func (b *mapKeyPointerBool) UnmarshalText(text []byte) error {
	*b = len(text) == 2
	return nil
}

type mapKeyPointerBoolValue bool

func (b mapKeyPointerBoolValue) UnmarshalText([]byte) error { return nil }

type mapKeyPointerText string

func (k *mapKeyPointerText) UnmarshalText(b []byte) error {
	*k = mapKeyPointerText("t" + string(b))
	return nil
}

func mapKeyPointerDump(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return "nil"
		}
		return "&" + mapKeyPointerDump(v.Elem())
	case reflect.Interface:
		if v.IsNil() {
			return "nil"
		}
		return fmt.Sprintf("%T(%v)", v.Elem().Interface(), v.Elem().Interface())
	}
	return fmt.Sprintf("%#v", v.Interface())
}

func mapKeyPointerDumpMap(m reflect.Value) string {
	if m.IsNil() {
		return "nil"
	}
	var entries []string
	for _, k := range m.MapKeys() {
		entries = append(entries, mapKeyPointerDump(k)+":"+mapKeyPointerDump(m.MapIndex(k)))
	}
	sort.Strings(entries)
	return "{" + strings.Join(entries, ",") + "}"
}

// The keys of a map of a pointer type are decoded into the values the pointers point to, at any depth, as
// encoding/json of Go 1.27 decodes them, or fails for them ( see comparesWithStd ).
func TestDecodeMapKeyPointers(t *testing.T) {
	keyTypes := []reflect.Type{
		reflect.TypeFor[int](), reflect.TypeFor[*int](), reflect.TypeFor[**int](), reflect.TypeFor[***int](),
		reflect.TypeFor[mapKeyPointerNamed](), reflect.TypeFor[*mapKeyPointerNamed](),
		reflect.TypeFor[*uint8](), reflect.TypeFor[*float64](), reflect.TypeFor[*bool](),
		reflect.TypeFor[*string](), reflect.TypeFor[**string](), reflect.TypeFor[*json.Number](),
		reflect.TypeFor[*mapKeyPointerText](), reflect.TypeFor[**mapKeyPointerText](), reflect.TypeFor[*any](),
		reflect.TypeFor[bool](), reflect.TypeFor[mapKeyPointerBool](), reflect.TypeFor[*mapKeyPointerBool](),
		reflect.TypeFor[mapKeyPointerBoolValue](),
	}
	inputs := []string{`{"5":1}`, `{"-3":1,"4":2}`, `{"x":1}`, `{"x":1,"5":2,"y":3,"6":4}`, `{"5":1,"x":2}`, `{"":1}`, `{"xx":1}`}
	for _, keyType := range keyTypes {
		typ := reflect.MapOf(keyType, reflect.TypeFor[int]())
		for _, in := range inputs {
			t.Run(typ.String()+in, func(t *testing.T) {
				decode := func(unmarshal func([]byte, any) error) (string, error) {
					p := reflect.New(typ)
					err := unmarshal([]byte(in), p.Interface())
					return fmt.Sprintf("%s, error %t", mapKeyPointerDumpMap(p.Elem()), err != nil), err
				}
				got, _ := decode(json.Unmarshal)
				want, stdErr := decode(stdjson.Unmarshal)
				if comparesWithStd(stdErr) && got != want {
					t.Errorf("got %s, want %s", got, want)
				}
			})
		}
	}
}
