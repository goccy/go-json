// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stdtest_test

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	ierrors "github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/jsontext"
	. "github.com/goccy/go-json/v2"
)

var (
	errInvalidFormatFlag = errors.New(`invalid format flag "invalid"`)
	errSomeError         = errors.New("some error")
	errMustNotCall       = errors.New("must not call")
)

func T[T any]() reflect.Type { return reflect.TypeFor[T]() }

type (
	jsonObject = map[string]any
	jsonArray  = []any

	namedAny     any
	namedBool    bool
	namedString  string
	NamedString  string
	namedBytes   []byte
	namedInt64   int64
	namedUint64  uint64
	namedFloat64 float64
	namedByte    byte
	netipAddr    = netip.Addr

	recursiveMap     map[string]recursiveMap
	recursiveSlice   []recursiveSlice
	recursivePointer struct{ P *recursivePointer }

	structEmpty       struct{}
	structConflicting struct {
		A string `json:"conflict"`
		B string `json:"conflict"`
	}
	structNoneExported struct {
		unexported string
	}
	structUnexportedIgnored struct {
		ignored string `json:"-"`
	}
	structMalformedTag struct {
		Malformed string `json:"\""`
	}
	structUnexportedTag struct {
		unexported string `json:"name"`
	}
	structExportedEmbedded struct {
		NamedString
	}
	structExportedEmbeddedTag struct {
		NamedString `json:"name"`
	}
	structUnexportedEmbedded struct {
		namedString
	}
	structUnexportedEmbeddedTag struct {
		namedString `json:"name"`
	}
	structUnexportedEmbeddedMethodTag struct {
		// netipAddr cannot be marshaled since the MarshalText method
		// cannot be called on an unexported field.
		netipAddr `json:"name"`

		// Bogus MarshalText and AppendText methods are declared on
		// structUnexportedEmbeddedMethodTag to prevent it from
		// implementing those method interfaces.
	}
	structUnexportedEmbeddedStruct struct {
		structOmitZeroAll
		FizzBuzz int
		structNestedAddr
	}
	structUnexportedEmbeddedStructPointer struct {
		*structOmitZeroAll
		FizzBuzz int
		*structNestedAddr
	}
	structNestedAddr struct {
		Addr netip.Addr
	}
	structIgnoredUnexportedEmbedded struct {
		namedString `json:"-"`
	}
	structNoCase struct {
		Aaa  string `json:",case:strict"`
		AA_A string
		AaA  string `json:",case:ignore"`
		AAa  string `json:",case:ignore"`
		AAA  string
	}
	structScalars struct {
		unexported bool
		Ignored    bool `json:"-"`

		Bool   bool
		String string
		Bytes  []byte
		Int    int64
		Uint   uint64
		Float  float64
	}
	structSlices struct {
		unexported bool
		Ignored    bool `json:"-"`

		SliceBool   []bool
		SliceString []string
		SliceBytes  [][]byte
		SliceInt    []int64
		SliceUint   []uint64
		SliceFloat  []float64
	}
	structMaps struct {
		unexported bool
		Ignored    bool `json:"-"`

		MapBool   map[string]bool
		MapString map[string]string
		MapBytes  map[string][]byte
		MapInt    map[string]int64
		MapUint   map[string]uint64
		MapFloat  map[string]float64
	}
	structAll struct {
		Bool          bool
		String        string
		Bytes         []byte
		Int           int64
		Uint          uint64
		Float         float64
		Map           map[string]string
		StructScalars structScalars
		StructMaps    structMaps
		StructSlices  structSlices
		Slice         []string
		Array         [1]string
		Pointer       *structAll
		Interface     any
	}
	structOmitZeroAll struct {
		Bool          bool               `json:",omitzero"`
		String        string             `json:",omitzero"`
		Bytes         []byte             `json:",omitzero"`
		Int           int64              `json:",omitzero"`
		Uint          uint64             `json:",omitzero"`
		Float         float64            `json:",omitzero"`
		Map           map[string]string  `json:",omitzero"`
		StructScalars structScalars      `json:",omitzero"`
		StructMaps    structMaps         `json:",omitzero"`
		StructSlices  structSlices       `json:",omitzero"`
		Slice         []string           `json:",omitzero"`
		Array         [1]string          `json:",omitzero"`
		Pointer       *structOmitZeroAll `json:",omitzero"`
		Interface     any                `json:",omitzero"`
	}
	structOmitZeroMethodAll struct {
		ValueAlwaysZero                 valueAlwaysZero     `json:",omitzero"`
		ValueNeverZero                  valueNeverZero      `json:",omitzero"`
		PointerAlwaysZero               pointerAlwaysZero   `json:",omitzero"`
		PointerNeverZero                pointerNeverZero    `json:",omitzero"`
		PointerValueAlwaysZero          *valueAlwaysZero    `json:",omitzero"`
		PointerValueNeverZero           *valueNeverZero     `json:",omitzero"`
		PointerPointerAlwaysZero        *pointerAlwaysZero  `json:",omitzero"`
		PointerPointerNeverZero         *pointerNeverZero   `json:",omitzero"`
		PointerPointerValueAlwaysZero   **valueAlwaysZero   `json:",omitzero"`
		PointerPointerValueNeverZero    **valueNeverZero    `json:",omitzero"`
		PointerPointerPointerAlwaysZero **pointerAlwaysZero `json:",omitzero"`
		PointerPointerPointerNeverZero  **pointerNeverZero  `json:",omitzero"`
	}
	structOmitZeroMethodInterfaceAll struct {
		ValueAlwaysZero          isZeroer `json:",omitzero"`
		ValueNeverZero           isZeroer `json:",omitzero"`
		PointerValueAlwaysZero   isZeroer `json:",omitzero"`
		PointerValueNeverZero    isZeroer `json:",omitzero"`
		PointerPointerAlwaysZero isZeroer `json:",omitzero"`
		PointerPointerNeverZero  isZeroer `json:",omitzero"`
	}
	structOmitEmptyAll struct {
		Bool                  bool                    `json:",omitempty"`
		PointerBool           *bool                   `json:",omitempty"`
		String                string                  `json:",omitempty"`
		StringEmpty           stringMarshalEmpty      `json:",omitempty"`
		StringNonEmpty        stringMarshalNonEmpty   `json:",omitempty"`
		PointerString         *string                 `json:",omitempty"`
		PointerStringEmpty    *stringMarshalEmpty     `json:",omitempty"`
		PointerStringNonEmpty *stringMarshalNonEmpty  `json:",omitempty"`
		Bytes                 []byte                  `json:",omitempty"`
		BytesEmpty            bytesMarshalEmpty       `json:",omitempty"`
		BytesNonEmpty         bytesMarshalNonEmpty    `json:",omitempty"`
		PointerBytes          *[]byte                 `json:",omitempty"`
		PointerBytesEmpty     *bytesMarshalEmpty      `json:",omitempty"`
		PointerBytesNonEmpty  *bytesMarshalNonEmpty   `json:",omitempty"`
		Float                 float64                 `json:",omitempty"`
		PointerFloat          *float64                `json:",omitempty"`
		Map                   map[string]string       `json:",omitempty"`
		MapEmpty              mapMarshalEmpty         `json:",omitempty"`
		MapNonEmpty           mapMarshalNonEmpty      `json:",omitempty"`
		PointerMap            *map[string]string      `json:",omitempty"`
		PointerMapEmpty       *mapMarshalEmpty        `json:",omitempty"`
		PointerMapNonEmpty    *mapMarshalNonEmpty     `json:",omitempty"`
		Slice                 []string                `json:",omitempty"`
		SliceEmpty            sliceMarshalEmpty       `json:",omitempty"`
		SliceNonEmpty         sliceMarshalNonEmpty    `json:",omitempty"`
		PointerSlice          *[]string               `json:",omitempty"`
		PointerSliceEmpty     *sliceMarshalEmpty      `json:",omitempty"`
		PointerSliceNonEmpty  *sliceMarshalNonEmpty   `json:",omitempty"`
		Pointer               *structOmitZeroEmptyAll `json:",omitempty"`
		Interface             any                     `json:",omitempty"`
	}
	structOmitZeroEmptyAll struct {
		Bool      bool                    `json:",omitzero,omitempty"`
		String    string                  `json:",omitzero,omitempty"`
		Bytes     []byte                  `json:",omitzero,omitempty"`
		Int       int64                   `json:",omitzero,omitempty"`
		Uint      uint64                  `json:",omitzero,omitempty"`
		Float     float64                 `json:",omitzero,omitempty"`
		Map       map[string]string       `json:",omitzero,omitempty"`
		Slice     []string                `json:",omitzero,omitempty"`
		Array     [1]string               `json:",omitzero,omitempty"`
		Pointer   *structOmitZeroEmptyAll `json:",omitzero,omitempty"`
		Interface any                     `json:",omitzero,omitempty"`
	}
	structStringifiedLegacy struct {
		Bool          bool     `json:",string"`
		String        string   `json:",string"`
		Int           int64    `json:",string"`
		Uint          uint64   `json:",string"`
		Float         float64  `json:",string"`
		PointerBool   *bool    `json:",string"`
		PointerString *string  `json:",string"`
		PointerInt    *int64   `json:",string"`
		PointerUint   *uint64  `json:",string"`
		PointerFloat  *float64 `json:",string"`
	}
	structStringified struct {
		Int          int64    `json:",string"`
		Uint         uint64   `json:",string"`
		Float        float64  `json:",string"`
		PointerInt   *int64   `json:",string"`
		PointerUint  *uint64  `json:",string"`
		PointerFloat *float64 `json:",string"`
	}
	structStringifiedBool struct {
		Bool bool `json:",string"`
	}
	structStringifiedString struct {
		String string `json:",string"`
	}
	structStringifiedBytes struct {
		Bytes []byte `json:",string"`
	}
	structStringifiedMap struct {
		Map map[string]string `json:",string"`
	}
	structStringifiedSlice struct {
		Slice []string `json:",string"`
	}
	structStringifiedArray struct {
		Array [1]string `json:",string"`
	}
	structStringifiedStruct struct {
		Struct structAll `json:",string"`
	}
	structStringifiedPointer struct {
		Pointer *structAll `json:",string"`
	}
	structStringifiedPointerPointerInt struct {
		Pointer **int `json:",string"`
	}
	structStringifiedInterface struct {
		Interface any `json:",string"`
	}
	structFormatBytes struct {
		Base16    []byte `json:",format:base16"`
		Base32    []byte `json:",format:base32"`
		Base32Hex []byte `json:",format:base32hex"`
		Base64    []byte `json:",format:base64"`
		Base64URL []byte `json:",format:base64url"`
		Array     []byte `json:",format:array"`
	}
	structFormatArrayBytes struct {
		Base16    [4]byte `json:",format:base16"`
		Base32    [4]byte `json:",format:base32"`
		Base32Hex [4]byte `json:",format:base32hex"`
		Base64    [4]byte `json:",format:base64"`
		Base64URL [4]byte `json:",format:base64url"`
		Array     [4]byte `json:",format:array"`
		Default   [4]byte
	}
	structFormatFloats struct {
		NonFinite        float64  `json:",format:nonfinite"`
		PointerNonFinite *float64 `json:",format:nonfinite"`
	}
	structFormatMaps struct {
		EmitNull           map[string]string  `json:",format:emitnull"`
		PointerEmitNull    *map[string]string `json:",format:emitnull"`
		EmitEmpty          map[string]string  `json:",format:emitempty"`
		PointerEmitEmpty   *map[string]string `json:",format:emitempty"`
		EmitDefault        map[string]string
		PointerEmitDefault *map[string]string
	}
	structFormatSlices struct {
		EmitNull           []string  `json:",format:emitnull"`
		PointerEmitNull    *[]string `json:",format:emitnull"`
		EmitEmpty          []string  `json:",format:emitempty"`
		PointerEmitEmpty   *[]string `json:",format:emitempty"`
		EmitDefault        []string
		PointerEmitDefault *[]string
	}
	structFormatInvalid struct {
		Bool      bool              `json:",omitzero,format:invalid"`
		String    string            `json:",omitzero,format:invalid"`
		Bytes     []byte            `json:",omitzero,format:invalid"`
		Int       int64             `json:",omitzero,format:invalid"`
		Uint      uint64            `json:",omitzero,format:invalid"`
		Float     float64           `json:",omitzero,format:invalid"`
		Map       map[string]string `json:",omitzero,format:invalid"`
		Struct    structAll         `json:",omitzero,format:invalid"`
		Slice     []string          `json:",omitzero,format:invalid"`
		Array     [1]string         `json:",omitzero,format:invalid"`
		Interface any               `json:",omitzero,format:invalid"`
	}
	structDurationFormat struct {
		D1  time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
		D2  time.Duration `json:",format:units"`
		D3  time.Duration `json:",format:sec"`
		D4  time.Duration `json:",string,format:sec"`
		D5  time.Duration `json:",format:milli"`
		D6  time.Duration `json:",string,format:milli"`
		D7  time.Duration `json:",format:micro"`
		D8  time.Duration `json:",string,format:micro"`
		D9  time.Duration `json:",format:nano"`
		D10 time.Duration `json:",string,format:nano"`
		D11 time.Duration `json:",format:iso8601"`
	}
	structTimeFormat struct {
		T1  time.Time
		T2  time.Time `json:",format:ANSIC"`
		T3  time.Time `json:",format:UnixDate"`
		T4  time.Time `json:",format:RubyDate"`
		T5  time.Time `json:",format:RFC822"`
		T6  time.Time `json:",format:RFC822Z"`
		T7  time.Time `json:",format:RFC850"`
		T8  time.Time `json:",format:RFC1123"`
		T9  time.Time `json:",format:RFC1123Z"`
		T10 time.Time `json:",format:RFC3339"`
		T11 time.Time `json:",format:RFC3339Nano"`
		T12 time.Time `json:",format:Kitchen"`
		T13 time.Time `json:",format:Stamp"`
		T14 time.Time `json:",format:StampMilli"`
		T15 time.Time `json:",format:StampMicro"`
		T16 time.Time `json:",format:StampNano"`
		T17 time.Time `json:",format:DateTime"`
		T18 time.Time `json:",format:DateOnly"`
		T19 time.Time `json:",format:TimeOnly"`
		T20 time.Time `json:",format:'2006-01-02'"`
		T21 time.Time `json:",format:'\"weird\"2006'"`
		T22 time.Time `json:",format:unix"`
		T23 time.Time `json:",string,format:unix"`
		T24 time.Time `json:",format:unixmilli"`
		T25 time.Time `json:",string,format:unixmilli"`
		T26 time.Time `json:",format:unixmicro"`
		T27 time.Time `json:",string,format:unixmicro"`
		T28 time.Time `json:",format:unixnano"`
		T29 time.Time `json:",string,format:unixnano"`
	}
	structTimeFormatStringInvalid struct {
		T time.Time `json:",string,format:RFC3339"`
	}
	structEmbedded struct {
		X             structEmbeddedL1 `json:",embed"`
		*StructEmbed2                  // implicit embed
	}
	structEmbeddedL1 struct {
		X            *structEmbeddedL2 `json:",embed"`
		StructEmbed1 `json:",embed"`
	}
	structEmbeddedL2     struct{ A, B, C string }
	StructEmbed1         struct{ C, D, E string }
	StructEmbed2         struct{ E, F, G string }
	structEmbedTextValue struct {
		A int            `json:",omitzero"`
		X jsontext.Value `json:",embed"`
		B int            `json:",omitzero"`
	}
	structEmbedPointerTextValue struct {
		A int             `json:",omitzero"`
		X *jsontext.Value `json:",embed"`
		B int             `json:",omitzero"`
	}
	structEmbedPointerEmbedTextValue struct {
		X *struct {
			A int
			X jsontext.Value `json:",embed"`
		} `json:",embed"`
	}
	structEmbedEmbedPointerTextValue struct {
		X struct {
			X *jsontext.Value `json:",embed"`
		} `json:",embed"`
	}
	structEmbedMapStringAny struct {
		A int        `json:",omitzero"`
		X jsonObject `json:",embed"`
		B int        `json:",omitzero"`
	}
	structEmbedPointerMapStringAny struct {
		A int         `json:",omitzero"`
		X *jsonObject `json:",embed"`
		B int         `json:",omitzero"`
	}
	structEmbedPointerEmbedMapStringAny struct {
		X *struct {
			A int
			X jsonObject `json:",embed"`
		} `json:",embed"`
	}
	structEmbedEmbedPointerMapStringAny struct {
		X struct {
			X *jsonObject `json:",embed"`
		} `json:",embed"`
	}
	structEmbedMapStringInt struct {
		X map[string]int `json:",embed"`
	}
	structEmbedMapNamedStringInt struct {
		X map[namedString]int `json:",embed"`
	}
	structEmbedMapNamedStringAny struct {
		A int                 `json:",omitzero"`
		X map[namedString]any `json:",embed"`
		B int                 `json:",omitzero"`
	}
	structNoCaseEmbedTextValue struct {
		AAA  string         `json:",omitempty,case:strict"`
		AA_b string         `json:",omitempty"`
		AaA  string         `json:",omitempty,case:ignore"`
		AAa  string         `json:",omitempty,case:ignore"`
		Aaa  string         `json:",omitempty"`
		X    jsontext.Value `json:",embed"`
	}
	structNoCaseEmbedMapStringAny struct {
		AAA string     `json:",omitempty"`
		AaA string     `json:",omitempty,case:ignore"`
		AAa string     `json:",omitempty,case:ignore"`
		Aaa string     `json:",omitempty"`
		X   jsonObject `json:",embed"`
	}

	allMethods struct {
		method string // the method that was called
		value  []byte // the raw value to provide or store
	}
	allMethodsExceptJSONv2 struct {
		allMethods
		MarshalJSONTo     struct{} // cancel out MarshalJSONTo method with collision
		UnmarshalJSONFrom struct{} // cancel out UnmarshalJSONFrom method with collision
	}
	allMethodsExceptJSONv1 struct {
		allMethods
		MarshalJSON   struct{} // cancel out MarshalJSON method with collision
		UnmarshalJSON struct{} // cancel out UnmarshalJSON method with collision
	}
	allMethodsExceptText struct {
		allMethods
		MarshalText   struct{} // cancel out MarshalText method with collision
		UnmarshalText struct{} // cancel out UnmarshalText method with collision
	}
	onlyMethodJSONv2 struct {
		allMethods
		MarshalJSON   struct{} // cancel out MarshalJSON method with collision
		UnmarshalJSON struct{} // cancel out UnmarshalJSON method with collision
		MarshalText   struct{} // cancel out MarshalText method with collision
		UnmarshalText struct{} // cancel out UnmarshalText method with collision
	}
	onlyMethodJSONv1 struct {
		allMethods
		MarshalJSONTo     struct{} // cancel out MarshalJSONTo method with collision
		UnmarshalJSONFrom struct{} // cancel out UnmarshalJSONFrom method with collision
		MarshalText       struct{} // cancel out MarshalText method with collision
		UnmarshalText     struct{} // cancel out UnmarshalText method with collision
	}
	onlyMethodText struct {
		allMethods
		MarshalJSONTo     struct{} // cancel out MarshalJSONTo method with collision
		UnmarshalJSONFrom struct{} // cancel out UnmarshalJSONFrom method with collision
		MarshalJSON       struct{} // cancel out MarshalJSON method with collision
		UnmarshalJSON     struct{} // cancel out UnmarshalJSON method with collision
	}

	unsupportedMethodJSONv2 map[string]int

	structMethodJSONv2 struct{ value string }
	structMethodJSONv1 struct{ value string }
	structMethodText   struct{ value string }

	marshalJSONv2Func   func(*jsontext.Encoder) error
	marshalJSONv1Func   func() ([]byte, error)
	appendTextFunc      func([]byte) ([]byte, error)
	marshalTextFunc     func() ([]byte, error)
	unmarshalJSONv2Func func(*jsontext.Decoder) error
	unmarshalJSONv1Func func([]byte) error
	unmarshalTextFunc   func([]byte) error

	nocaseString string

	stringMarshalEmpty    string
	stringMarshalNonEmpty string
	bytesMarshalEmpty     []byte
	bytesMarshalNonEmpty  []byte
	mapMarshalEmpty       map[string]string
	mapMarshalNonEmpty    map[string]string
	sliceMarshalEmpty     []string
	sliceMarshalNonEmpty  []string

	valueAlwaysZero   string
	valueNeverZero    string
	pointerAlwaysZero string
	pointerNeverZero  string

	valueStringer   struct{}
	pointerStringer struct{}

	cyclicA struct {
		B1 cyclicB `json:",embed"`
		B2 cyclicB `json:",embed"`
	}
	cyclicB struct {
		F int
		A *cyclicA `json:",embed"`
	}
)

func (structUnexportedEmbeddedMethodTag) MarshalText() {}
func (structUnexportedEmbeddedMethodTag) AppendText()  {}

func (p *allMethods) MarshalJSONTo(enc *jsontext.Encoder) error {
	if got, want := "MarshalJSONTo", p.method; got != want {
		return fmt.Errorf("called wrong method: got %v, want %v", got, want)
	}
	return enc.WriteValue(p.value)
}
func (p *allMethods) MarshalJSON() ([]byte, error) {
	if got, want := "MarshalJSON", p.method; got != want {
		return nil, fmt.Errorf("called wrong method: got %v, want %v", got, want)
	}
	return p.value, nil
}
func (p *allMethods) MarshalText() ([]byte, error) {
	if got, want := "MarshalText", p.method; got != want {
		return nil, fmt.Errorf("called wrong method: got %v, want %v", got, want)
	}
	return p.value, nil
}

func (p *allMethods) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	p.method = "UnmarshalJSONFrom"
	val, err := dec.ReadValue()
	p.value = val
	return err
}
func (p *allMethods) UnmarshalJSON(val []byte) error {
	p.method = "UnmarshalJSON"
	p.value = val
	return nil
}
func (p *allMethods) UnmarshalText(val []byte) error {
	p.method = "UnmarshalText"
	p.value = val
	return nil
}

func (s *unsupportedMethodJSONv2) MarshalJSONTo(enc *jsontext.Encoder) error {
	(*s)["called"] += 1
	return errors.ErrUnsupported
}
func (s *unsupportedMethodJSONv2) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	(*s)["called"] += 1
	return errors.ErrUnsupported
}

func (s structMethodJSONv2) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.String(s.value))
}
func (s *structMethodJSONv2) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	tok, err := dec.ReadToken()
	if err != nil {
		return err
	}
	if k := tok.Kind(); k != '"' {
		return EU(nil).withType(k, T[structMethodJSONv2]())
	}
	s.value = tok.String()
	return nil
}

func (s structMethodJSONv1) MarshalJSON() ([]byte, error) {
	return jsontext.AppendQuote(nil, s.value)
}
func (s *structMethodJSONv1) UnmarshalJSON(b []byte) error {
	if k := jsontext.Value(b).Kind(); k != '"' {
		return EU(nil).withType(k, T[structMethodJSONv1]())
	}
	b, _ = jsontext.AppendUnquote(nil, b)
	s.value = string(b)
	return nil
}

func (s structMethodText) MarshalText() ([]byte, error) {
	return []byte(s.value), nil
}
func (s *structMethodText) UnmarshalText(b []byte) error {
	s.value = string(b)
	return nil
}

func (f marshalJSONv2Func) MarshalJSONTo(enc *jsontext.Encoder) error {
	return f(enc)
}
func (f marshalJSONv1Func) MarshalJSON() ([]byte, error) {
	return f()
}
func (f appendTextFunc) AppendText(b []byte) ([]byte, error) {
	return f(b)
}
func (f marshalTextFunc) MarshalText() ([]byte, error) {
	return f()
}
func (f unmarshalJSONv2Func) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return f(dec)
}
func (f unmarshalJSONv1Func) UnmarshalJSON(b []byte) error {
	return f(b)
}
func (f unmarshalTextFunc) UnmarshalText(b []byte) error {
	return f(b)
}

func (k nocaseString) MarshalText() ([]byte, error) {
	return []byte(strings.ToLower(string(k))), nil
}
func (k *nocaseString) UnmarshalText(b []byte) error {
	*k = nocaseString(strings.ToLower(string(b)))
	return nil
}

func (stringMarshalEmpty) MarshalJSON() ([]byte, error)    { return []byte(`""`), nil }
func (stringMarshalNonEmpty) MarshalJSON() ([]byte, error) { return []byte(`"value"`), nil }
func (bytesMarshalEmpty) MarshalJSON() ([]byte, error)     { return []byte(`[]`), nil }
func (bytesMarshalNonEmpty) MarshalJSON() ([]byte, error)  { return []byte(`["value"]`), nil }
func (mapMarshalEmpty) MarshalJSON() ([]byte, error)       { return []byte(`{}`), nil }
func (mapMarshalNonEmpty) MarshalJSON() ([]byte, error)    { return []byte(`{"key":"value"}`), nil }
func (sliceMarshalEmpty) MarshalJSON() ([]byte, error)     { return []byte(`[]`), nil }
func (sliceMarshalNonEmpty) MarshalJSON() ([]byte, error)  { return []byte(`["value"]`), nil }

func (valueAlwaysZero) IsZero() bool    { return true }
func (valueNeverZero) IsZero() bool     { return false }
func (*pointerAlwaysZero) IsZero() bool { return true }
func (*pointerNeverZero) IsZero() bool  { return false }

func (valueStringer) String() string    { return "" }
func (*pointerStringer) String() string { return "" }

func addr[T any](v T) *T {
	return &v
}

func mustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

// invalidFormatOption is an option where the Format field is set,
// but jsonflags.FormatTag has been cleared. In such a case,
// the Format field should be ignored.
var invalidFormatOption = ignoredFormat{}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    CaseName
		opts    []Options
		in      any
		want    string
		wantErr error

		canonicalize bool // canonicalize the output before comparing?
		useWriter    bool // call MarshalWrite instead of Marshal
	}{{
		name: Name("Nil"),
		in:   nil,
		want: `null`,
	}, {
		name: Name("Bools"),
		in:   []bool{false, true},
		want: `[false,true]`,
	}, {
		name: Name("Bools/Named"),
		in:   []namedBool{false, true},
		want: `[false,true]`,
	}, {
		name: Name("Bools/NotStringified"),
		opts: []Options{StringifyNumbers(true)},
		in:   []bool{false, true},
		want: `[false,true]`,
	}, {
		name: Name("Bools/StringifiedBool/False"),
		opts: []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		in:   false,
		want: `"false"`,
	}, {
		name: Name("Bools/StringifiedBool/True"),
		opts: []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		in:   true,
		want: `"true"`,
	}, {
		name: Name("Bools/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   true,
		want: `true`,
	}, {
		name: Name("Strings"),
		in:   []string{"", "hello", "世界"},
		want: `["","hello","世界"]`,
	}, {
		name: Name("Strings/Named"),
		in:   []namedString{"", "hello", "世界"},
		want: `["","hello","世界"]`,
	}, {
		name: Name("Strings/StringifiedString"),
		opts: []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		in:   "hello",
		want: `"\"hello\""`,
	}, {
		name: Name("Strings/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   "string",
		want: `"string"`,
	}, {
		name: Name("Bytes"),
		in:   [][]byte{nil, {}, {1}, {1, 2}, {1, 2, 3}},
		want: `["","","AQ==","AQI=","AQID"]`,
	}, {
		name: Name("Bytes/FormatNilSliceAsNull"),
		opts: []Options{FormatNilSliceAsNull(true)},
		in:   [][]byte{nil, {}},
		want: `[null,""]`,
	}, {
		name: Name("Bytes/Large"),
		in:   []byte("the quick brown fox jumped over the lazy dog and ate the homework that I spent so much time on."),
		want: `"dGhlIHF1aWNrIGJyb3duIGZveCBqdW1wZWQgb3ZlciB0aGUgbGF6eSBkb2cgYW5kIGF0ZSB0aGUgaG9tZXdvcmsgdGhhdCBJIHNwZW50IHNvIG11Y2ggdGltZSBvbi4="`,
	}, {
		name: Name("Bytes/Named"),
		in:   []namedBytes{nil, {}, {1}, {1, 2}, {1, 2, 3}},
		want: `["","","AQ==","AQI=","AQID"]`,
	}, {
		name: Name("Bytes/NotStringified"),
		opts: []Options{StringifyNumbers(true)},
		in:   [][]byte{nil, {}, {1}, {1, 2}, {1, 2, 3}},
		want: `["","","AQ==","AQI=","AQID"]`,
	}, {
		// NOTE: []namedByte is not assignable to []byte,
		// so the following should be treated as a slice of uints.
		name: Name("Bytes/Invariant"),
		in:   [][]namedByte{nil, {}, {1}, {1, 2}, {1, 2, 3}},
		want: `[[],[],[1],[1,2],[1,2,3]]`,
	}, {
		// NOTE: This differs in behavior from v1,
		// but keeps the representation of slices and arrays more consistent.
		name: Name("Bytes/ByteArray"),
		in:   [5]byte{'h', 'e', 'l', 'l', 'o'},
		want: `"aGVsbG8="`,
	}, {
		// NOTE: []namedByte is not assignable to []byte,
		// so the following should be treated as an array of uints.
		name: Name("Bytes/NamedByteArray"),
		in:   [5]namedByte{'h', 'e', 'l', 'l', 'o'},
		want: `[104,101,108,108,111]`,
	}, {
		name: Name("Bytes/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   []byte("hello"),
		want: `"aGVsbG8="`,
	}, {
		name: Name("Ints"),
		in: []any{
			int(0), int8(math.MinInt8), int16(math.MinInt16), int32(math.MinInt32), int64(math.MinInt64), namedInt64(-6464),
		},
		want: `[0,-128,-32768,-2147483648,-9223372036854775808,-6464]`,
	}, {
		name: Name("Ints/Stringified"),
		opts: []Options{StringifyNumbers(true)},
		in: []any{
			int(0), int8(math.MinInt8), int16(math.MinInt16), int32(math.MinInt32), int64(math.MinInt64), namedInt64(-6464),
		},
		want: `["0","-128","-32768","-2147483648","-9223372036854775808","-6464"]`,
	}, {
		name: Name("Ints/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   int(0),
		want: `0`,
	}, {
		name: Name("Uints"),
		in: []any{
			uint(0), uint8(math.MaxUint8), uint16(math.MaxUint16), uint32(math.MaxUint32), uint64(math.MaxUint64), namedUint64(6464), uintptr(1234),
		},
		want: `[0,255,65535,4294967295,18446744073709551615,6464,1234]`,
	}, {
		name: Name("Uints/Stringified"),
		opts: []Options{StringifyNumbers(true)},
		in: []any{
			uint(0), uint8(math.MaxUint8), uint16(math.MaxUint16), uint32(math.MaxUint32), uint64(math.MaxUint64), namedUint64(6464),
		},
		want: `["0","255","65535","4294967295","18446744073709551615","6464"]`,
	}, {
		name: Name("Uints/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   uint(0),
		want: `0`,
	}, {
		name: Name("Floats"),
		in: []any{
			float32(math.MaxFloat32), float64(math.MaxFloat64), namedFloat64(64.64),
		},
		want: `[3.4028235e+38,1.7976931348623157e+308,64.64]`,
	}, {
		name: Name("Floats/Stringified"),
		opts: []Options{StringifyNumbers(true)},
		in: []any{
			float32(math.MaxFloat32), float64(math.MaxFloat64), namedFloat64(64.64),
		},
		want: `["3.4028235e+38","1.7976931348623157e+308","64.64"]`,
	}, {
		name:    Name("Floats/Invalid/NaN"),
		opts:    []Options{StringifyNumbers(true)},
		in:      math.NaN(),
		wantErr: EM(fmt.Errorf("unsupported value: %v", math.NaN())).withType(0, float64Type),
	}, {
		name:    Name("Floats/Invalid/PositiveInfinity"),
		in:      math.Inf(+1),
		wantErr: EM(fmt.Errorf("unsupported value: %v", math.Inf(+1))).withType(0, float64Type),
	}, {
		name:    Name("Floats/Invalid/NegativeInfinity"),
		in:      math.Inf(-1),
		wantErr: EM(fmt.Errorf("unsupported value: %v", math.Inf(-1))).withType(0, float64Type),
	}, {
		name: Name("Floats/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   float64(0),
		want: `0`,
	}, {
		name:    Name("Maps/InvalidKey/Bool"),
		in:      map[bool]string{false: "value"},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, boolType),
	}, {
		name:    Name("Maps/InvalidKey/NamedBool"),
		in:      map[namedBool]string{false: "value"},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[namedBool]()),
	}, {
		name:    Name("Maps/InvalidKey/Array"),
		in:      map[[1]string]string{{"key"}: "value"},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[[1]string]()),
	}, {
		name:    Name("Maps/InvalidKey/Channel"),
		in:      map[chan string]string{make(chan string): "value"},
		want:    `{`,
		wantErr: EM(nil).withPos(`{`, "").withType(0, T[chan string]()),
	}, {
		name:         Name("Maps/ValidKey/Int"),
		in:           map[int64]string{math.MinInt64: "MinInt64", 0: "Zero", math.MaxInt64: "MaxInt64"},
		canonicalize: true,
		want:         `{"-9223372036854775808":"MinInt64","0":"Zero","9223372036854775807":"MaxInt64"}`,
	}, {
		name:         Name("Maps/ValidKey/PointerInt"),
		in:           map[*int64]string{addr(int64(math.MinInt64)): "MinInt64", addr(int64(0)): "Zero", addr(int64(math.MaxInt64)): "MaxInt64"},
		canonicalize: true,
		want:         `{"-9223372036854775808":"MinInt64","0":"Zero","9223372036854775807":"MaxInt64"}`,
	}, {
		name:         Name("Maps/DuplicateName/PointerInt"),
		in:           map[*int64]string{addr(int64(0)): "0", addr(int64(0)): "0"},
		canonicalize: true,
		want:         `{"0":"0"`,
		wantErr:      newDuplicateNameError("", []byte(`"0"`), len64(`{"0":"0",`)),
	}, {
		name:         Name("Maps/ValidKey/NamedInt"),
		in:           map[namedInt64]string{math.MinInt64: "MinInt64", 0: "Zero", math.MaxInt64: "MaxInt64"},
		canonicalize: true,
		want:         `{"-9223372036854775808":"MinInt64","0":"Zero","9223372036854775807":"MaxInt64"}`,
	}, {
		name:         Name("Maps/ValidKey/Uint"),
		in:           map[uint64]string{0: "Zero", math.MaxUint64: "MaxUint64"},
		canonicalize: true,
		want:         `{"0":"Zero","18446744073709551615":"MaxUint64"}`,
	}, {
		name:         Name("Maps/ValidKey/NamedUint"),
		in:           map[namedUint64]string{0: "Zero", math.MaxUint64: "MaxUint64"},
		canonicalize: true,
		want:         `{"0":"Zero","18446744073709551615":"MaxUint64"}`,
	}, {
		name: Name("Maps/ValidKey/Float"),
		in:   map[float64]string{3.14159: "value"},
		want: `{"3.14159":"value"}`,
	}, {
		name:    Name("Maps/InvalidKey/Float/NaN"),
		in:      map[float64]string{math.NaN(): "NaN", math.NaN(): "NaN"},
		want:    `{`,
		wantErr: EM(errors.New("unsupported value: NaN")).withPos(`{`, "").withType(0, float64Type),
	}, {
		name: Name("Maps/ValidKey/Interface"),
		in: map[any]any{
			"key":               "key",
			namedInt64(-64):     int32(-32),
			namedUint64(+64):    uint32(+32),
			namedFloat64(64.64): float32(32.32),
		},
		canonicalize: true,
		want:         `{"-64":-32,"64":32,"64.64":32.32,"key":"key"}`,
	}, {
		name: Name("Maps/DuplicateName/String/AllowInvalidUTF8+AllowDuplicateNames"),
		opts: []Options{jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true)},
		in:   map[string]string{"\x80": "", "\x81": ""},
		want: `{"�":"","�":""}`,
	}, {
		name:    Name("Maps/DuplicateName/String/AllowInvalidUTF8"),
		opts:    []Options{jsontext.AllowInvalidUTF8(true)},
		in:      map[string]string{"\x80": "", "\x81": ""},
		want:    `{"�":""`,
		wantErr: newDuplicateNameError("", []byte(`"�"`), len64(`{"�":"",`)),
	}, {
		name: Name("Maps/DuplicateName/NoCaseString/AllowDuplicateNames"),
		opts: []Options{jsontext.AllowDuplicateNames(true)},
		in:   map[nocaseString]string{"hello": "", "HELLO": ""},
		want: `{"hello":"","hello":""}`,
	}, {
		name:    Name("Maps/DuplicateName/NoCaseString"),
		in:      map[nocaseString]string{"hello": "", "HELLO": ""},
		want:    `{"hello":""`,
		wantErr: EM(newDuplicateNameError("", []byte(`"hello"`), len64(`{"hello":"",`))).withPos(`{"hello":"",`, "").withType(0, T[nocaseString]()),
	}, {
		name: Name("Maps/DuplicateName/NaNs/Deterministic+AllowDuplicateNames"),
		opts: []Options{
			WithMarshalers(
				MarshalFunc(func(v float64) ([]byte, error) { return []byte(`"NaN"`), nil }),
			),
			Deterministic(true),
			jsontext.AllowDuplicateNames(true),
		},
		in:   map[float64]string{math.NaN(): "NaN", math.NaN(): "NaN"},
		want: `{"NaN":"NaN","NaN":"NaN"}`,
	}, {
		name: Name("Maps/InvalidValue/Channel"),
		in: map[string]chan string{
			"key": nil,
		},
		want:    `{"key"`,
		wantErr: EM(nil).withPos(`{"key":`, "/key").withType(0, T[chan string]()),
	}, {
		name: Name("Maps/String/Deterministic"),
		opts: []Options{Deterministic(true)},
		in:   map[string]int{"a": 0, "b": 1, "c": 2},
		want: `{"a":0,"b":1,"c":2}`,
	}, {
		name: Name("Maps/String/Deterministic+AllowInvalidUTF8+RejectDuplicateNames"),
		opts: []Options{
			Deterministic(true),
			jsontext.AllowInvalidUTF8(true),
			jsontext.AllowDuplicateNames(false),
		},
		in:      map[string]int{"\xff": 0, "\xfe": 1},
		want:    `{"�":1`,
		wantErr: newDuplicateNameError("", []byte(`"�"`), len64(`{"�":1,`)),
	}, {
		name: Name("Maps/String/Deterministic+AllowInvalidUTF8+AllowDuplicateNames"),
		opts: []Options{
			Deterministic(true),
			jsontext.AllowInvalidUTF8(true),
			jsontext.AllowDuplicateNames(true),
		},
		in:   map[string]int{"\xff": 0, "\xfe": 1},
		want: `{"�":1,"�":0}`,
	}, {
		name: Name("Maps/String/Deterministic+MarshalFuncs"),
		opts: []Options{
			Deterministic(true),
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v string) error {
				if p := enc.StackPointer(); p != "/X" {
					return fmt.Errorf("invalid stack pointer: got %s, want /X", p)
				}
				switch v {
				case "a":
					return enc.WriteToken(jsontext.String("b"))
				case "b":
					return enc.WriteToken(jsontext.String("a"))
				default:
					return fmt.Errorf("invalid value: %q", v)
				}
			})),
		},
		in:   map[namedString]map[string]int{"X": {"a": -1, "b": 1}},
		want: `{"X":{"a":1,"b":-1}}`,
	}, {
		name: Name("Maps/String/Deterministic+MarshalFuncs+RejectDuplicateNames"),
		opts: []Options{
			Deterministic(true),
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v string) error {
				if p := enc.StackPointer(); p != "/X" {
					return fmt.Errorf("invalid stack pointer: got %s, want /X", p)
				}
				switch v {
				case "a", "b":
					return enc.WriteToken(jsontext.String("x"))
				default:
					return fmt.Errorf("invalid value: %q", v)
				}
			})),
			jsontext.AllowDuplicateNames(false),
		},
		in:      map[namedString]map[string]int{"X": {"a": 1, "b": 1}},
		want:    `{"X":{"x":1`,
		wantErr: newDuplicateNameError("/X/x", nil, len64(`{"X":{"x":1,`)),
	}, {
		name: Name("Maps/String/Deterministic+MarshalFuncs+AllowDuplicateNames"),
		opts: []Options{
			Deterministic(true),
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v string) error {
				if p := enc.StackPointer(); p != "/X" {
					return fmt.Errorf("invalid stack pointer: got %s, want /0", p)
				}
				switch v {
				case "a", "b":
					return enc.WriteToken(jsontext.String("x"))
				default:
					return fmt.Errorf("invalid value: %q", v)
				}
			})),
			jsontext.AllowDuplicateNames(true),
		},
		in: map[namedString]map[string]int{"X": {"a": 1, "b": 1}},
		// NOTE: Since the names are identical, the exact values may be
		// non-deterministic since sort cannot distinguish between members.
		want: `{"X":{"x":1,"x":1}}`,
	}, {
		name: Name("Maps/RecursiveMap"),
		in: recursiveMap{
			"fizz": {
				"foo": {},
				"bar": nil,
			},
			"buzz": nil,
		},
		canonicalize: true,
		want:         `{"buzz":{},"fizz":{"bar":{},"foo":{}}}`,
	}, {
		name: Name("Maps/CyclicMap"),
		in: func() recursiveMap {
			m := recursiveMap{"k": nil}
			m["k"] = m
			return m
		}(),
		want:    strings.Repeat(`{"k":`, startDetectingCyclesAfter) + `{"k"`,
		wantErr: EM(errCycle).withPos(strings.Repeat(`{"k":`, startDetectingCyclesAfter+1), jsontext.Pointer(strings.Repeat("/k", startDetectingCyclesAfter+1))).withType(0, T[recursiveMap]()),
	}, {
		name: Name("Maps/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   map[string]string{},
		want: `{}`,
	}, {
		name: Name("Structs/Empty"),
		in:   structEmpty{},
		want: `{}`,
	}, {
		name: Name("Structs/UnexportedIgnored"),
		in:   structUnexportedIgnored{ignored: "ignored"},
		want: `{}`,
	}, {
		name: Name("Structs/IgnoredUnexportedEmbedded"),
		in:   structIgnoredUnexportedEmbedded{namedString: "ignored"},
		want: `{}`,
	}, {
		name: Name("Structs/NoCase"),
		in:   structNoCase{AaA: "AaA", AAa: "AAa", Aaa: "Aaa", AAA: "AAA", AA_A: "AA_A"},
		want: `{"Aaa":"Aaa","AA_A":"AA_A","AaA":"AaA","AAa":"AAa","AAA":"AAA"}`,
	}, {
		name: Name("Structs/NoCase/MatchCaseInsensitiveNames"),
		opts: []Options{MatchCaseInsensitiveNames(true)},
		in:   structNoCase{AaA: "AaA", AAa: "AAa", Aaa: "Aaa", AAA: "AAA", AA_A: "AA_A"},
		want: `{"Aaa":"Aaa","AA_A":"AA_A","AaA":"AaA","AAa":"AAa","AAA":"AAA"}`,
	}, {
		name: Name("Structs/NoCase/MatchCaseInsensitiveNames+MatchCaseSensitiveDelimiter"),
		opts: []Options{MatchCaseInsensitiveNames(true), unsupported("MatchCaseSensitiveDelimiter")},
		in:   structNoCase{AaA: "AaA", AAa: "AAa", Aaa: "Aaa", AAA: "AAA", AA_A: "AA_A"},
		want: `{"Aaa":"Aaa","AA_A":"AA_A","AaA":"AaA","AAa":"AAa","AAA":"AAA"}`,
	}, {
		name: Name("Structs/Normal"),
		opts: []Options{jsontext.Multiline(true)},
		in: structAll{
			Bool:   true,
			String: "hello",
			Bytes:  []byte{1, 2, 3},
			Int:    -64,
			Uint:   +64,
			Float:  3.14159,
			Map:    map[string]string{"key": "value"},
			StructScalars: structScalars{
				Bool:   true,
				String: "hello",
				Bytes:  []byte{1, 2, 3},
				Int:    -64,
				Uint:   +64,
				Float:  3.14159,
			},
			StructMaps: structMaps{
				MapBool:   map[string]bool{"": true},
				MapString: map[string]string{"": "hello"},
				MapBytes:  map[string][]byte{"": {1, 2, 3}},
				MapInt:    map[string]int64{"": -64},
				MapUint:   map[string]uint64{"": +64},
				MapFloat:  map[string]float64{"": 3.14159},
			},
			StructSlices: structSlices{
				SliceBool:   []bool{true},
				SliceString: []string{"hello"},
				SliceBytes:  [][]byte{{1, 2, 3}},
				SliceInt:    []int64{-64},
				SliceUint:   []uint64{+64},
				SliceFloat:  []float64{3.14159},
			},
			Slice:     []string{"fizz", "buzz"},
			Array:     [1]string{"goodbye"},
			Pointer:   new(structAll),
			Interface: (*structAll)(nil),
		},
		want: `{
	"Bool": true,
	"String": "hello",
	"Bytes": "AQID",
	"Int": -64,
	"Uint": 64,
	"Float": 3.14159,
	"Map": {
		"key": "value"
	},
	"StructScalars": {
		"Bool": true,
		"String": "hello",
		"Bytes": "AQID",
		"Int": -64,
		"Uint": 64,
		"Float": 3.14159
	},
	"StructMaps": {
		"MapBool": {
			"": true
		},
		"MapString": {
			"": "hello"
		},
		"MapBytes": {
			"": "AQID"
		},
		"MapInt": {
			"": -64
		},
		"MapUint": {
			"": 64
		},
		"MapFloat": {
			"": 3.14159
		}
	},
	"StructSlices": {
		"SliceBool": [
			true
		],
		"SliceString": [
			"hello"
		],
		"SliceBytes": [
			"AQID"
		],
		"SliceInt": [
			-64
		],
		"SliceUint": [
			64
		],
		"SliceFloat": [
			3.14159
		]
	},
	"Slice": [
		"fizz",
		"buzz"
	],
	"Array": [
		"goodbye"
	],
	"Pointer": {
		"Bool": false,
		"String": "",
		"Bytes": "",
		"Int": 0,
		"Uint": 0,
		"Float": 0,
		"Map": {},
		"StructScalars": {
			"Bool": false,
			"String": "",
			"Bytes": "",
			"Int": 0,
			"Uint": 0,
			"Float": 0
		},
		"StructMaps": {
			"MapBool": {},
			"MapString": {},
			"MapBytes": {},
			"MapInt": {},
			"MapUint": {},
			"MapFloat": {}
		},
		"StructSlices": {
			"SliceBool": [],
			"SliceString": [],
			"SliceBytes": [],
			"SliceInt": [],
			"SliceUint": [],
			"SliceFloat": []
		},
		"Slice": [],
		"Array": [
			""
		],
		"Pointer": null,
		"Interface": null
	},
	"Interface": null
}`,
	}, {
		name: Name("Structs/SpaceAfterColonAndComma"),
		opts: []Options{jsontext.SpaceAfterColon(true), jsontext.SpaceAfterComma(true)},
		in:   structOmitZeroAll{Int: 1, Uint: 1},
		want: `{"Int": 1, "Uint": 1}`,
	}, {
		name: Name("Structs/SpaceAfterColon"),
		opts: []Options{jsontext.SpaceAfterColon(true)},
		in:   structOmitZeroAll{Int: 1, Uint: 1},
		want: `{"Int": 1,"Uint": 1}`,
	}, {
		name: Name("Structs/SpaceAfterComma"),
		opts: []Options{jsontext.SpaceAfterComma(true)},
		in:   structOmitZeroAll{Int: 1, Uint: 1, Slice: []string{"a", "b"}},
		want: `{"Int":1, "Uint":1, "Slice":["a", "b"]}`,
	}, {
		name: Name("Structs/Stringified"),
		opts: []Options{jsontext.Multiline(true)},
		in: structStringified{
			Int:          -64,
			Uint:         +64,
			Float:        3.14159,
			PointerInt:   ptr(int64(-64)),
			PointerUint:  ptr(uint64(+64)),
			PointerFloat: ptr(float64(3.14159)),
		},
		want: `{
	"Int": "-64",
	"Uint": "64",
	"Float": "3.14159",
	"PointerInt": "-64",
	"PointerUint": "64",
	"PointerFloat": "3.14159"
}`,
	}, {
		name: Name("Structs/LegacyStringified"),
		opts: []Options{jsontext.Multiline(true), unsupported("StringifyWithLegacySemantics")},
		in: structStringifiedLegacy{
			Bool:          true,
			String:        "hello",
			Int:           -64,
			Uint:          +64,
			Float:         3.14159,
			PointerBool:   ptr(true),
			PointerString: ptr("hello"),
			PointerInt:    ptr(int64(-64)),
			PointerUint:   ptr(uint64(+64)),
			PointerFloat:  ptr(float64(3.14159)),
		},
		want: `{
	"Bool": "true",
	"String": "\"hello\"",
	"Int": "-64",
	"Uint": "64",
	"Float": "3.14159",
	"PointerBool": "true",
	"PointerString": "\"hello\"",
	"PointerInt": "-64",
	"PointerUint": "64",
	"PointerFloat": "3.14159"
}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/Bool"),
		in:      structStringifiedBool{},
		want:    `{"Bool"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Bool":`, "/Bool").withType(0, boolType),
	}, {
		name: Name("Structs/Stringified/Ignored/Bool"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedBool{},
		want: `{"Bool":false}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/String"),
		in:      structStringifiedString{},
		want:    `{"String"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"String":`, "/String").withType(0, stringType),
	}, {
		name: Name("Structs/Stringified/Ignored/String"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedString{},
		want: `{"String":""}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/Bytes"),
		in:      structStringifiedBytes{},
		want:    `{"Bytes"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Bytes":`, "/Bytes").withType(0, bytesType),
	}, {
		name: Name("Structs/Stringified/Ignored/Bytes"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedBytes{},
		want: `{"Bytes":""}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/Map"),
		in:      structStringifiedMap{},
		want:    `{"Map"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Map":`, "/Map").withType(0, reflect.TypeFor[map[string]string]()),
	}, {
		name: Name("Structs/Stringified/Ignored/Map"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedMap{},
		want: `{"Map":{}}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/Slice"),
		in:      structStringifiedSlice{},
		want:    `{"Slice"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Slice":`, "/Slice").withType(0, reflect.TypeFor[[]string]()),
	}, {
		name: Name("Structs/Stringified/Ignored/Slice"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedSlice{},
		want: `{"Slice":[]}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/Array"),
		in:      structStringifiedArray{},
		want:    `{"Array"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Array":`, "/Array").withType(0, reflect.TypeFor[[1]string]()),
	}, {
		name: Name("Structs/Stringified/Ignored/Array"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedArray{},
		want: `{"Array":[""]}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/Struct"),
		in:      structStringifiedStruct{},
		want:    `{"Struct"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Struct":`, "/Struct").withType(0, reflect.TypeFor[structAll]()),
	}, {
		name: Name("Structs/Stringified/Ignored/Struct"),
		opts: []Options{jsontext.Multiline(true), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedStruct{},
		want: `{
	"Struct": {
		"Bool": false,
		"String": "",
		"Bytes": "",
		"Int": 0,
		"Uint": 0,
		"Float": 0,
		"Map": {},
		"StructScalars": {
			"Bool": false,
			"String": "",
			"Bytes": "",
			"Int": 0,
			"Uint": 0,
			"Float": 0
		},
		"StructMaps": {
			"MapBool": {},
			"MapString": {},
			"MapBytes": {},
			"MapInt": {},
			"MapUint": {},
			"MapFloat": {}
		},
		"StructSlices": {
			"SliceBool": [],
			"SliceString": [],
			"SliceBytes": [],
			"SliceInt": [],
			"SliceUint": [],
			"SliceFloat": []
		},
		"Slice": [],
		"Array": [
			""
		],
		"Pointer": null,
		"Interface": null
	}
}`,
	}, {
		name: Name("Structs/Stringified/Invalid/Pointer"),
		in:   structStringifiedPointer{},
		want: `{"Pointer":null}`,
	}, {
		name: Name("Structs/Stringified/Ignored/Pointer"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedPointer{},
		want: `{"Pointer":null}`,
	}, {
		name: Name("Structs/Stringified/PointerPointerInt"),
		in:   structStringifiedPointerPointerInt{Pointer: ptr(ptr(5))},
		want: `{"Pointer":"5"}`,
	}, {
		name:    Name("Structs/Stringified/Invalid/Interface"),
		in:      structStringifiedInterface{Interface: 1000},
		want:    `{"Interface"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Interface":`, "/Interface").withType(0, anyType),
	}, {
		name: Name("Structs/Stringified/Ignored/Interface"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedInterface{Interface: 1000},
		want: `{"Interface":"1000"}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Bytes"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		want:    `{"Bytes"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Bytes":`, "/Bytes").withType(0, bytesType),
		in:      structStringifiedBytes{},
	}, {
		name: Name("Structs/LegacyStringified/Ignored/Bytes"),
		opts: []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedBytes{},
		want: `{"Bytes":""}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Map"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		in:      structStringifiedMap{},
		want:    `{"Map"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Map":`, "/Map").withType(0, reflect.TypeFor[map[string]string]()),
	}, {
		name: Name("Structs/LegacyStringified/Ignored/Map"),
		opts: []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedMap{},
		want: `{"Map":{}}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Slice"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		in:      structStringifiedSlice{},
		want:    `{"Slice"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Slice":`, "/Slice").withType(0, reflect.TypeFor[[]string]()),
	}, {
		name: Name("Structs/LegacyStringified/Ignored/Slice"),
		opts: []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedSlice{},
		want: `{"Slice":[]}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Array"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		in:      structStringifiedArray{},
		want:    `{"Array"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Array":`, "/Array").withType(0, reflect.TypeFor[[1]string]()),
	}, {
		name: Name("Structs/LegacyStringified/Ignored/Array"),
		opts: []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedArray{},
		want: `{"Array":[""]}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Struct"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		in:      structStringifiedStruct{},
		want:    `{"Struct"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Struct":`, "/Struct").withType(0, reflect.TypeFor[structAll]()),
	}, {
		name: Name("Structs/LegacyStringified/Ignored/Struct"),
		opts: []Options{jsontext.Multiline(true), unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedStruct{},
		want: `{
	"Struct": {
		"Bool": false,
		"String": "",
		"Bytes": "",
		"Int": 0,
		"Uint": 0,
		"Float": 0,
		"Map": {},
		"StructScalars": {
			"Bool": false,
			"String": "",
			"Bytes": "",
			"Int": 0,
			"Uint": 0,
			"Float": 0
		},
		"StructMaps": {
			"MapBool": {},
			"MapString": {},
			"MapBytes": {},
			"MapInt": {},
			"MapUint": {},
			"MapFloat": {}
		},
		"StructSlices": {
			"SliceBool": [],
			"SliceString": [],
			"SliceBytes": [],
			"SliceInt": [],
			"SliceUint": [],
			"SliceFloat": []
		},
		"Slice": [],
		"Array": [
			""
		],
		"Pointer": null,
		"Interface": null
	}
}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Pointer"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		in:      structStringifiedPointer{Pointer: new(structAll)},
		want:    `{"Pointer"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Pointer":`, "/Pointer").withType(0, reflect.TypeFor[structAll]()),
	}, {
		name: Name("Structs/LegacyStringified/Ignored/Pointer"),
		opts: []Options{jsontext.Multiline(true), unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedPointer{Pointer: new(structAll)},
		want: `{
	"Pointer": {
		"Bool": false,
		"String": "",
		"Bytes": "",
		"Int": 0,
		"Uint": 0,
		"Float": 0,
		"Map": {},
		"StructScalars": {
			"Bool": false,
			"String": "",
			"Bytes": "",
			"Int": 0,
			"Uint": 0,
			"Float": 0
		},
		"StructMaps": {
			"MapBool": {},
			"MapString": {},
			"MapBytes": {},
			"MapInt": {},
			"MapUint": {},
			"MapFloat": {}
		},
		"StructSlices": {
			"SliceBool": [],
			"SliceString": [],
			"SliceBytes": [],
			"SliceInt": [],
			"SliceUint": [],
			"SliceFloat": []
		},
		"Slice": [],
		"Array": [
			""
		],
		"Pointer": null,
		"Interface": null
	}
}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/PointerPointerInt"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		in:      structStringifiedPointerPointerInt{Pointer: ptr(ptr(5))},
		want:    `{"Pointer"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Pointer":`, "/Pointer").withType(0, reflect.TypeFor[**int]()),
	}, {
		name: Name("Structs/LegacyStringified/Ignored/PointerPointerInt"),
		opts: []Options{unsupported("StringifyWithLegacySemantics|ReportErrorsWithLegacySemantics")},
		in:   structStringifiedPointerPointerInt{Pointer: ptr(ptr(5))},
		want: `{"Pointer":5}`,
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Interface"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		in:      structStringifiedInterface{},
		want:    `{"Interface"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"Interface":`, "/Interface").withType(0, anyType),
	}, {
		name: Name("Structs/LegacyStringified/Ignored/Interface"),
		opts: []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		in:   structStringifiedInterface{Interface: 1000},
		want: `{"Interface":1000}`,
	}, {
		name: Name("Structs/OmitZero/Zero"),
		in:   structOmitZeroAll{},
		want: `{}`,
	}, {
		name: Name("Structs/OmitZeroOption/Zero"),
		opts: []Options{OmitZeroStructFields(true)},
		in:   structAll{},
		want: `{}`,
	}, {
		name: Name("Structs/OmitZero/NonZero"),
		opts: []Options{jsontext.Multiline(true)},
		in: structOmitZeroAll{
			Bool:          true,                                   // not omitted since true is non-zero
			String:        " ",                                    // not omitted since non-empty string is non-zero
			Bytes:         []byte{},                               // not omitted since allocated slice is non-zero
			Int:           1,                                      // not omitted since 1 is non-zero
			Uint:          1,                                      // not omitted since 1 is non-zero
			Float:         math.SmallestNonzeroFloat64,            // not omitted since still slightly above zero
			Map:           map[string]string{},                    // not omitted since allocated map is non-zero
			StructScalars: structScalars{unexported: true},        // not omitted since unexported is non-zero
			StructSlices:  structSlices{Ignored: true},            // not omitted since Ignored is non-zero
			StructMaps:    structMaps{MapBool: map[string]bool{}}, // not omitted since MapBool is non-zero
			Slice:         []string{},                             // not omitted since allocated slice is non-zero
			Array:         [1]string{" "},                         // not omitted since single array element is non-zero
			Pointer:       new(structOmitZeroAll),                 // not omitted since pointer is non-zero (even if all fields of the struct value are zero)
			Interface:     (*structOmitZeroAll)(nil),              // not omitted since interface value is non-zero (even if interface value is a nil pointer)
		},
		want: `{
	"Bool": true,
	"String": " ",
	"Bytes": "",
	"Int": 1,
	"Uint": 1,
	"Float": 5e-324,
	"Map": {},
	"StructScalars": {
		"Bool": false,
		"String": "",
		"Bytes": "",
		"Int": 0,
		"Uint": 0,
		"Float": 0
	},
	"StructMaps": {
		"MapBool": {},
		"MapString": {},
		"MapBytes": {},
		"MapInt": {},
		"MapUint": {},
		"MapFloat": {}
	},
	"StructSlices": {
		"SliceBool": [],
		"SliceString": [],
		"SliceBytes": [],
		"SliceInt": [],
		"SliceUint": [],
		"SliceFloat": []
	},
	"Slice": [],
	"Array": [
		" "
	],
	"Pointer": {},
	"Interface": null
}`,
	}, {
		name: Name("Structs/OmitZeroOption/NonZero"),
		opts: []Options{OmitZeroStructFields(true), jsontext.Multiline(true)},
		in: structAll{
			Bool:          true,
			String:        " ",
			Bytes:         []byte{},
			Int:           1,
			Uint:          1,
			Float:         math.SmallestNonzeroFloat64,
			Map:           map[string]string{},
			StructScalars: structScalars{unexported: true},
			StructSlices:  structSlices{Ignored: true},
			StructMaps:    structMaps{MapBool: map[string]bool{}},
			Slice:         []string{},
			Array:         [1]string{" "},
			Pointer:       new(structAll),
			Interface:     (*structAll)(nil),
		},
		want: `{
	"Bool": true,
	"String": " ",
	"Bytes": "",
	"Int": 1,
	"Uint": 1,
	"Float": 5e-324,
	"Map": {},
	"StructScalars": {},
	"StructMaps": {
		"MapBool": {}
	},
	"StructSlices": {},
	"Slice": [],
	"Array": [
		" "
	],
	"Pointer": {},
	"Interface": null
}`,
	}, {
		name: Name("Structs/OmitZeroMethod/Zero"),
		in:   structOmitZeroMethodAll{},
		want: `{"ValueNeverZero":"","PointerNeverZero":""}`,
	}, {
		name: Name("Structs/OmitZeroMethod/NonZero"),
		opts: []Options{jsontext.Multiline(true)},
		in: structOmitZeroMethodAll{
			ValueAlwaysZero:                 valueAlwaysZero("nonzero"),
			ValueNeverZero:                  valueNeverZero("nonzero"),
			PointerAlwaysZero:               pointerAlwaysZero("nonzero"),
			PointerNeverZero:                pointerNeverZero("nonzero"),
			PointerValueAlwaysZero:          addr(valueAlwaysZero("nonzero")),
			PointerValueNeverZero:           addr(valueNeverZero("nonzero")),
			PointerPointerAlwaysZero:        addr(pointerAlwaysZero("nonzero")),
			PointerPointerNeverZero:         addr(pointerNeverZero("nonzero")),
			PointerPointerValueAlwaysZero:   addr(addr(valueAlwaysZero("nonzero"))), // marshaled since **valueAlwaysZero does not implement IsZero
			PointerPointerValueNeverZero:    addr(addr(valueNeverZero("nonzero"))),
			PointerPointerPointerAlwaysZero: addr(addr(pointerAlwaysZero("nonzero"))), // marshaled since **pointerAlwaysZero does not implement IsZero
			PointerPointerPointerNeverZero:  addr(addr(pointerNeverZero("nonzero"))),
		},
		want: `{
	"ValueNeverZero": "nonzero",
	"PointerNeverZero": "nonzero",
	"PointerValueNeverZero": "nonzero",
	"PointerPointerNeverZero": "nonzero",
	"PointerPointerValueAlwaysZero": "nonzero",
	"PointerPointerValueNeverZero": "nonzero",
	"PointerPointerPointerAlwaysZero": "nonzero",
	"PointerPointerPointerNeverZero": "nonzero"
}`,
	}, {
		name: Name("Structs/OmitZeroMethod/Interface/Zero"),
		opts: []Options{jsontext.Multiline(true)},
		in:   structOmitZeroMethodInterfaceAll{},
		want: `{}`,
	}, {
		name: Name("Structs/OmitZeroMethod/Interface/PartialZero"),
		opts: []Options{jsontext.Multiline(true)},
		in: structOmitZeroMethodInterfaceAll{
			ValueAlwaysZero:          valueAlwaysZero(""),
			ValueNeverZero:           valueNeverZero(""),
			PointerValueAlwaysZero:   (*valueAlwaysZero)(nil),
			PointerValueNeverZero:    (*valueNeverZero)(nil), // nil pointer, so method not called
			PointerPointerAlwaysZero: (*pointerAlwaysZero)(nil),
			PointerPointerNeverZero:  (*pointerNeverZero)(nil), // nil pointer, so method not called
		},
		want: `{
	"ValueNeverZero": ""
}`,
	}, {
		name: Name("Structs/OmitZeroMethod/Interface/NonZero"),
		opts: []Options{jsontext.Multiline(true)},
		in: structOmitZeroMethodInterfaceAll{
			ValueAlwaysZero:          valueAlwaysZero("nonzero"),
			ValueNeverZero:           valueNeverZero("nonzero"),
			PointerValueAlwaysZero:   addr(valueAlwaysZero("nonzero")),
			PointerValueNeverZero:    addr(valueNeverZero("nonzero")),
			PointerPointerAlwaysZero: addr(pointerAlwaysZero("nonzero")),
			PointerPointerNeverZero:  addr(pointerNeverZero("nonzero")),
		},
		want: `{
	"ValueNeverZero": "nonzero",
	"PointerValueNeverZero": "nonzero",
	"PointerPointerNeverZero": "nonzero"
}`,
	}, {
		name: Name("Structs/OmitEmpty/Zero"),
		opts: []Options{jsontext.Multiline(true)},
		in:   structOmitEmptyAll{},
		want: `{
	"Bool": false,
	"StringNonEmpty": "value",
	"BytesNonEmpty": [
		"value"
	],
	"Float": 0,
	"MapNonEmpty": {
		"key": "value"
	},
	"SliceNonEmpty": [
		"value"
	]
}`,
	}, {
		name: Name("Structs/OmitEmpty/EmptyNonZero"),
		opts: []Options{jsontext.Multiline(true)},
		in: structOmitEmptyAll{
			String:                string(""),
			StringEmpty:           stringMarshalEmpty(""),
			StringNonEmpty:        stringMarshalNonEmpty(""),
			PointerString:         addr(string("")),
			PointerStringEmpty:    addr(stringMarshalEmpty("")),
			PointerStringNonEmpty: addr(stringMarshalNonEmpty("")),
			Bytes:                 []byte(""),
			BytesEmpty:            bytesMarshalEmpty([]byte("")),
			BytesNonEmpty:         bytesMarshalNonEmpty([]byte("")),
			PointerBytes:          addr([]byte("")),
			PointerBytesEmpty:     addr(bytesMarshalEmpty([]byte(""))),
			PointerBytesNonEmpty:  addr(bytesMarshalNonEmpty([]byte(""))),
			Map:                   map[string]string{},
			MapEmpty:              mapMarshalEmpty{},
			MapNonEmpty:           mapMarshalNonEmpty{},
			PointerMap:            addr(map[string]string{}),
			PointerMapEmpty:       addr(mapMarshalEmpty{}),
			PointerMapNonEmpty:    addr(mapMarshalNonEmpty{}),
			Slice:                 []string{},
			SliceEmpty:            sliceMarshalEmpty{},
			SliceNonEmpty:         sliceMarshalNonEmpty{},
			PointerSlice:          addr([]string{}),
			PointerSliceEmpty:     addr(sliceMarshalEmpty{}),
			PointerSliceNonEmpty:  addr(sliceMarshalNonEmpty{}),
			Pointer:               &structOmitZeroEmptyAll{},
			Interface:             []string{},
		},
		want: `{
	"Bool": false,
	"StringNonEmpty": "value",
	"PointerStringNonEmpty": "value",
	"BytesNonEmpty": [
		"value"
	],
	"PointerBytesNonEmpty": [
		"value"
	],
	"Float": 0,
	"MapNonEmpty": {
		"key": "value"
	},
	"PointerMapNonEmpty": {
		"key": "value"
	},
	"SliceNonEmpty": [
		"value"
	],
	"PointerSliceNonEmpty": [
		"value"
	]
}`,
	}, {
		name: Name("Structs/OmitEmpty/NonEmpty"),
		opts: []Options{jsontext.Multiline(true)},
		in: structOmitEmptyAll{
			Bool:                  true,
			PointerBool:           addr(true),
			String:                string("value"),
			StringEmpty:           stringMarshalEmpty("value"),
			StringNonEmpty:        stringMarshalNonEmpty("value"),
			PointerString:         addr(string("value")),
			PointerStringEmpty:    addr(stringMarshalEmpty("value")),
			PointerStringNonEmpty: addr(stringMarshalNonEmpty("value")),
			Bytes:                 []byte("value"),
			BytesEmpty:            bytesMarshalEmpty([]byte("value")),
			BytesNonEmpty:         bytesMarshalNonEmpty([]byte("value")),
			PointerBytes:          addr([]byte("value")),
			PointerBytesEmpty:     addr(bytesMarshalEmpty([]byte("value"))),
			PointerBytesNonEmpty:  addr(bytesMarshalNonEmpty([]byte("value"))),
			Float:                 math.Copysign(0, -1),
			PointerFloat:          addr(math.Copysign(0, -1)),
			Map:                   map[string]string{"": ""},
			MapEmpty:              mapMarshalEmpty{"key": "value"},
			MapNonEmpty:           mapMarshalNonEmpty{"key": "value"},
			PointerMap:            addr(map[string]string{"": ""}),
			PointerMapEmpty:       addr(mapMarshalEmpty{"key": "value"}),
			PointerMapNonEmpty:    addr(mapMarshalNonEmpty{"key": "value"}),
			Slice:                 []string{""},
			SliceEmpty:            sliceMarshalEmpty{"value"},
			SliceNonEmpty:         sliceMarshalNonEmpty{"value"},
			PointerSlice:          addr([]string{""}),
			PointerSliceEmpty:     addr(sliceMarshalEmpty{"value"}),
			PointerSliceNonEmpty:  addr(sliceMarshalNonEmpty{"value"}),
			Pointer:               &structOmitZeroEmptyAll{Float: math.SmallestNonzeroFloat64},
			Interface:             []string{""},
		},
		want: `{
	"Bool": true,
	"PointerBool": true,
	"String": "value",
	"StringNonEmpty": "value",
	"PointerString": "value",
	"PointerStringNonEmpty": "value",
	"Bytes": "dmFsdWU=",
	"BytesNonEmpty": [
		"value"
	],
	"PointerBytes": "dmFsdWU=",
	"PointerBytesNonEmpty": [
		"value"
	],
	"Float": -0,
	"PointerFloat": -0,
	"Map": {
		"": ""
	},
	"MapNonEmpty": {
		"key": "value"
	},
	"PointerMap": {
		"": ""
	},
	"PointerMapNonEmpty": {
		"key": "value"
	},
	"Slice": [
		""
	],
	"SliceNonEmpty": [
		"value"
	],
	"PointerSlice": [
		""
	],
	"PointerSliceNonEmpty": [
		"value"
	],
	"Pointer": {
		"Float": 5e-324
	},
	"Interface": [
		""
	]
}`,
	}, {
		name: Name("Structs/OmitEmpty/Legacy/Zero"),
		opts: []Options{unsupported("OmitEmptyWithLegacySemantics")},
		in:   structOmitEmptyAll{},
		want: `{}`,
	}, {
		name: Name("Structs/OmitEmpty/Legacy/NonEmpty"),
		opts: []Options{jsontext.Multiline(true), unsupported("OmitEmptyWithLegacySemantics")},
		in: structOmitEmptyAll{
			Bool:                  true,
			PointerBool:           addr(true),
			String:                string("value"),
			StringEmpty:           stringMarshalEmpty("value"),
			StringNonEmpty:        stringMarshalNonEmpty("value"),
			PointerString:         addr(string("value")),
			PointerStringEmpty:    addr(stringMarshalEmpty("value")),
			PointerStringNonEmpty: addr(stringMarshalNonEmpty("value")),
			Bytes:                 []byte("value"),
			BytesEmpty:            bytesMarshalEmpty([]byte("value")),
			BytesNonEmpty:         bytesMarshalNonEmpty([]byte("value")),
			PointerBytes:          addr([]byte("value")),
			PointerBytesEmpty:     addr(bytesMarshalEmpty([]byte("value"))),
			PointerBytesNonEmpty:  addr(bytesMarshalNonEmpty([]byte("value"))),
			Float:                 math.Copysign(0, -1),
			PointerFloat:          addr(math.Copysign(0, -1)),
			Map:                   map[string]string{"": ""},
			MapEmpty:              mapMarshalEmpty{"key": "value"},
			MapNonEmpty:           mapMarshalNonEmpty{"key": "value"},
			PointerMap:            addr(map[string]string{"": ""}),
			PointerMapEmpty:       addr(mapMarshalEmpty{"key": "value"}),
			PointerMapNonEmpty:    addr(mapMarshalNonEmpty{"key": "value"}),
			Slice:                 []string{""},
			SliceEmpty:            sliceMarshalEmpty{"value"},
			SliceNonEmpty:         sliceMarshalNonEmpty{"value"},
			PointerSlice:          addr([]string{""}),
			PointerSliceEmpty:     addr(sliceMarshalEmpty{"value"}),
			PointerSliceNonEmpty:  addr(sliceMarshalNonEmpty{"value"}),
			Pointer:               &structOmitZeroEmptyAll{Float: math.Copysign(0, -1)},
			Interface:             []string{""},
		},
		want: `{
	"Bool": true,
	"PointerBool": true,
	"String": "value",
	"StringEmpty": "",
	"StringNonEmpty": "value",
	"PointerString": "value",
	"PointerStringEmpty": "",
	"PointerStringNonEmpty": "value",
	"Bytes": "dmFsdWU=",
	"BytesEmpty": [],
	"BytesNonEmpty": [
		"value"
	],
	"PointerBytes": "dmFsdWU=",
	"PointerBytesEmpty": [],
	"PointerBytesNonEmpty": [
		"value"
	],
	"PointerFloat": -0,
	"Map": {
		"": ""
	},
	"MapEmpty": {},
	"MapNonEmpty": {
		"key": "value"
	},
	"PointerMap": {
		"": ""
	},
	"PointerMapEmpty": {},
	"PointerMapNonEmpty": {
		"key": "value"
	},
	"Slice": [
		""
	],
	"SliceEmpty": [],
	"SliceNonEmpty": [
		"value"
	],
	"PointerSlice": [
		""
	],
	"PointerSliceEmpty": [],
	"PointerSliceNonEmpty": [
		"value"
	],
	"Pointer": {},
	"Interface": [
		""
	]
}`,
	}, {
		name: Name("Structs/OmitEmpty/NonEmptyString"),
		in: struct {
			X string `json:",omitempty"`
		}{`"`},
		want: `{"X":"\""}`,
	}, {
		name: Name("Structs/OmitZeroEmpty/Zero"),
		in:   structOmitZeroEmptyAll{},
		want: `{}`,
	}, {
		name: Name("Structs/OmitZeroEmpty/Empty"),
		in: structOmitZeroEmptyAll{
			Bytes:     []byte{},
			Map:       map[string]string{},
			Slice:     []string{},
			Pointer:   &structOmitZeroEmptyAll{},
			Interface: []string{},
		},
		want: `{}`,
	}, {
		name: Name("Structs/OmitEmpty/PathologicalDepth"),
		in: func() any {
			type X struct {
				X *X `json:",omitempty"`
			}
			var make func(int) *X
			make = func(n int) *X {
				if n == 0 {
					return nil
				}
				return &X{make(n - 1)}
			}
			return make(100)
		}(),
		want:      `{}`,
		useWriter: true,
	}, {
		name: Name("Structs/OmitEmpty/PathologicalBreadth"),
		in: func() any {
			var fields []reflect.StructField
			for i := range 100 {
				fields = append(fields, reflect.StructField{
					Name: fmt.Sprintf("X%d", i),
					Type: T[stringMarshalEmpty](),
					Tag:  `json:",omitempty"`,
				})
			}
			return reflect.New(reflect.StructOf(fields)).Interface()
		}(),
		want:      `{}`,
		useWriter: true,
	}, {
		name: Name("Structs/OmitEmpty/PathologicalTree"),
		in: func() any {
			type X struct {
				XL, XR *X `json:",omitempty"`
			}
			var make func(int) *X
			make = func(n int) *X {
				if n == 0 {
					return nil
				}
				return &X{make(n - 1), make(n - 1)}
			}
			return make(8)
		}(),
		want:      `{}`,
		useWriter: true,
	}, {
		name: Name("Structs/OmitZeroEmpty/NonEmpty"),
		in: structOmitZeroEmptyAll{
			Bytes:     []byte("value"),
			Map:       map[string]string{"": ""},
			Slice:     []string{""},
			Pointer:   &structOmitZeroEmptyAll{Bool: true},
			Interface: []string{""},
		},
		want: `{"Bytes":"dmFsdWU=","Map":{"":""},"Slice":[""],"Pointer":{"Bool":true},"Interface":[""]}`,
	}, {
		name:    Name("Structs/Format/Bytes/Unsupported"),
		opts:    []Options{jsontext.Multiline(true)},
		in:      structFormatBytes{},
		wantErr: EM(errors.New("Go struct field Base16 has unsupported `format` tag option")).withType(0, T[structFormatBytes]()),
	}, {
		name: Name("Structs/Format/Bytes"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true)},
		in: structFormatBytes{
			Base16:    []byte("\x01\x23\x45\x67\x89\xab\xcd\xef"),
			Base32:    []byte("\x00D2\x14\xc7BT\xb65τe:V\xd7\xc6u\xbew\xdf"),
			Base32Hex: []byte("\x00D2\x14\xc7BT\xb65τe:V\xd7\xc6u\xbew\xdf"),
			Base64:    []byte("\x00\x10\x83\x10Q\x87 \x92\x8b0ӏA\x14\x93QU\x97a\x96\x9bqן\x82\x18\xa3\x92Y\xa7\xa2\x9a\xab\xb2ۯ\xc3\x1c\xb3\xd3]\xb7㞻\xf3߿"),
			Base64URL: []byte("\x00\x10\x83\x10Q\x87 \x92\x8b0ӏA\x14\x93QU\x97a\x96\x9bqן\x82\x18\xa3\x92Y\xa7\xa2\x9a\xab\xb2ۯ\xc3\x1c\xb3\xd3]\xb7㞻\xf3߿"),
			Array:     []byte{1, 2, 3, 4},
		},
		want: `{
	"Base16": "0123456789abcdef",
	"Base32": "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567",
	"Base32Hex": "0123456789ABCDEFGHIJKLMNOPQRSTUV",
	"Base64": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
	"Base64URL": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_",
	"Array": [
		1,
		2,
		3,
		4
	]
}`,
	}, {
		name: Name("Structs/Format/ArrayBytes"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true)},
		in: structFormatArrayBytes{
			Base16:    [4]byte{1, 2, 3, 4},
			Base32:    [4]byte{1, 2, 3, 4},
			Base32Hex: [4]byte{1, 2, 3, 4},
			Base64:    [4]byte{1, 2, 3, 4},
			Base64URL: [4]byte{1, 2, 3, 4},
			Array:     [4]byte{1, 2, 3, 4},
			Default:   [4]byte{1, 2, 3, 4},
		},
		want: `{
	"Base16": "01020304",
	"Base32": "AEBAGBA=",
	"Base32Hex": "0410610=",
	"Base64": "AQIDBA==",
	"Base64URL": "AQIDBA==",
	"Array": [
		1,
		2,
		3,
		4
	],
	"Default": "AQIDBA=="
}`,
	}, {
		name: Name("Structs/Format/ArrayBytes/Legacy"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true), unsupported("FormatByteArrayAsArray|FormatBytesWithLegacySemantics")},
		in: structFormatArrayBytes{
			Base16:    [4]byte{1, 2, 3, 4},
			Base32:    [4]byte{1, 2, 3, 4},
			Base32Hex: [4]byte{1, 2, 3, 4},
			Base64:    [4]byte{1, 2, 3, 4},
			Base64URL: [4]byte{1, 2, 3, 4},
			Array:     [4]byte{1, 2, 3, 4},
			Default:   [4]byte{1, 2, 3, 4},
		},
		want: `{
	"Base16": "01020304",
	"Base32": "AEBAGBA=",
	"Base32Hex": "0410610=",
	"Base64": "AQIDBA==",
	"Base64URL": "AQIDBA==",
	"Array": [
		1,
		2,
		3,
		4
	],
	"Default": [
		1,
		2,
		3,
		4
	]
}`,
	}, {
		name: Name("Structs/Format/Bytes/Array"),
		opts: []Options{
			unsupported("ExperimentalSupportFormatTag"),
			WithMarshalers(MarshalFunc(func(in byte) ([]byte, error) {
				if in > 3 {
					return []byte("true"), nil
				} else {
					return []byte("false"), nil
				}
			})),
		},
		in: struct {
			Array []byte `json:",format:array"`
		}{
			Array: []byte{1, 6, 2, 5, 3, 4},
		},
		want: `{"Array":[false,true,false,true,false,true]}`,
	}, {
		name: Name("Structs/Format/Floats"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true)},
		in: []structFormatFloats{
			{NonFinite: math.Pi, PointerNonFinite: addr(math.Pi)},
			{NonFinite: math.NaN(), PointerNonFinite: addr(math.NaN())},
			{NonFinite: math.Inf(-1), PointerNonFinite: addr(math.Inf(-1))},
			{NonFinite: math.Inf(+1), PointerNonFinite: addr(math.Inf(+1))},
		},
		want: `[
	{
		"NonFinite": 3.141592653589793,
		"PointerNonFinite": 3.141592653589793
	},
	{
		"NonFinite": "NaN",
		"PointerNonFinite": "NaN"
	},
	{
		"NonFinite": "-Infinity",
		"PointerNonFinite": "-Infinity"
	},
	{
		"NonFinite": "Infinity",
		"PointerNonFinite": "Infinity"
	}
]`,
	}, {
		name: Name("Structs/Format/Maps"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true)},
		in: []structFormatMaps{{
			EmitNull: map[string]string(nil), PointerEmitNull: addr(map[string]string(nil)),
			EmitEmpty: map[string]string(nil), PointerEmitEmpty: addr(map[string]string(nil)),
			EmitDefault: map[string]string(nil), PointerEmitDefault: addr(map[string]string(nil)),
		}, {
			EmitNull: map[string]string{}, PointerEmitNull: addr(map[string]string{}),
			EmitEmpty: map[string]string{}, PointerEmitEmpty: addr(map[string]string{}),
			EmitDefault: map[string]string{}, PointerEmitDefault: addr(map[string]string{}),
		}, {
			EmitNull: map[string]string{"k": "v"}, PointerEmitNull: addr(map[string]string{"k": "v"}),
			EmitEmpty: map[string]string{"k": "v"}, PointerEmitEmpty: addr(map[string]string{"k": "v"}),
			EmitDefault: map[string]string{"k": "v"}, PointerEmitDefault: addr(map[string]string{"k": "v"}),
		}},
		want: `[
	{
		"EmitNull": null,
		"PointerEmitNull": null,
		"EmitEmpty": {},
		"PointerEmitEmpty": {},
		"EmitDefault": {},
		"PointerEmitDefault": {}
	},
	{
		"EmitNull": {},
		"PointerEmitNull": {},
		"EmitEmpty": {},
		"PointerEmitEmpty": {},
		"EmitDefault": {},
		"PointerEmitDefault": {}
	},
	{
		"EmitNull": {
			"k": "v"
		},
		"PointerEmitNull": {
			"k": "v"
		},
		"EmitEmpty": {
			"k": "v"
		},
		"PointerEmitEmpty": {
			"k": "v"
		},
		"EmitDefault": {
			"k": "v"
		},
		"PointerEmitDefault": {
			"k": "v"
		}
	}
]`,
	}, {
		name: Name("Structs/Format/Maps/FormatNilMapAsNull"),
		opts: []Options{
			unsupported("ExperimentalSupportFormatTag"),
			FormatNilMapAsNull(true),
			jsontext.Multiline(true),
		},
		in: []structFormatMaps{{
			EmitNull: map[string]string(nil), PointerEmitNull: addr(map[string]string(nil)),
			EmitEmpty: map[string]string(nil), PointerEmitEmpty: addr(map[string]string(nil)),
			EmitDefault: map[string]string(nil), PointerEmitDefault: addr(map[string]string(nil)),
		}, {
			EmitNull: map[string]string{}, PointerEmitNull: addr(map[string]string{}),
			EmitEmpty: map[string]string{}, PointerEmitEmpty: addr(map[string]string{}),
			EmitDefault: map[string]string{}, PointerEmitDefault: addr(map[string]string{}),
		}, {
			EmitNull: map[string]string{"k": "v"}, PointerEmitNull: addr(map[string]string{"k": "v"}),
			EmitEmpty: map[string]string{"k": "v"}, PointerEmitEmpty: addr(map[string]string{"k": "v"}),
			EmitDefault: map[string]string{"k": "v"}, PointerEmitDefault: addr(map[string]string{"k": "v"}),
		}},
		want: `[
	{
		"EmitNull": null,
		"PointerEmitNull": null,
		"EmitEmpty": {},
		"PointerEmitEmpty": {},
		"EmitDefault": null,
		"PointerEmitDefault": null
	},
	{
		"EmitNull": {},
		"PointerEmitNull": {},
		"EmitEmpty": {},
		"PointerEmitEmpty": {},
		"EmitDefault": {},
		"PointerEmitDefault": {}
	},
	{
		"EmitNull": {
			"k": "v"
		},
		"PointerEmitNull": {
			"k": "v"
		},
		"EmitEmpty": {
			"k": "v"
		},
		"PointerEmitEmpty": {
			"k": "v"
		},
		"EmitDefault": {
			"k": "v"
		},
		"PointerEmitDefault": {
			"k": "v"
		}
	}
]`,
	}, {
		name: Name("Structs/Format/Slices"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true)},
		in: []structFormatSlices{{
			EmitNull: []string(nil), PointerEmitNull: addr([]string(nil)),
			EmitEmpty: []string(nil), PointerEmitEmpty: addr([]string(nil)),
			EmitDefault: []string(nil), PointerEmitDefault: addr([]string(nil)),
		}, {
			EmitNull: []string{}, PointerEmitNull: addr([]string{}),
			EmitEmpty: []string{}, PointerEmitEmpty: addr([]string{}),
			EmitDefault: []string{}, PointerEmitDefault: addr([]string{}),
		}, {
			EmitNull: []string{"v"}, PointerEmitNull: addr([]string{"v"}),
			EmitEmpty: []string{"v"}, PointerEmitEmpty: addr([]string{"v"}),
			EmitDefault: []string{"v"}, PointerEmitDefault: addr([]string{"v"}),
		}},
		want: `[
	{
		"EmitNull": null,
		"PointerEmitNull": null,
		"EmitEmpty": [],
		"PointerEmitEmpty": [],
		"EmitDefault": [],
		"PointerEmitDefault": []
	},
	{
		"EmitNull": [],
		"PointerEmitNull": [],
		"EmitEmpty": [],
		"PointerEmitEmpty": [],
		"EmitDefault": [],
		"PointerEmitDefault": []
	},
	{
		"EmitNull": [
			"v"
		],
		"PointerEmitNull": [
			"v"
		],
		"EmitEmpty": [
			"v"
		],
		"PointerEmitEmpty": [
			"v"
		],
		"EmitDefault": [
			"v"
		],
		"PointerEmitDefault": [
			"v"
		]
	}
]`,
	}, {
		name:    Name("Structs/Format/Invalid/Bool"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Bool: true},
		want:    `{"Bool"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Bool":`, "/Bool").withType(0, boolType),
	}, {
		name:    Name("Structs/Format/Invalid/String"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{String: "string"},
		want:    `{"String"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"String":`, "/String").withType(0, stringType),
	}, {
		name:    Name("Structs/Format/Invalid/Bytes"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Bytes: []byte("bytes")},
		want:    `{"Bytes"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Bytes":`, "/Bytes").withType(0, bytesType),
	}, {
		name:    Name("Structs/Format/Invalid/Int"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Int: 1},
		want:    `{"Int"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Int":`, "/Int").withType(0, T[int64]()),
	}, {
		name:    Name("Structs/Format/Invalid/Uint"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Uint: 1},
		want:    `{"Uint"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Uint":`, "/Uint").withType(0, T[uint64]()),
	}, {
		name:    Name("Structs/Format/Invalid/Float"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Float: 1},
		want:    `{"Float"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Float":`, "/Float").withType(0, T[float64]()),
	}, {
		name:    Name("Structs/Format/Invalid/Map"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Map: map[string]string{}},
		want:    `{"Map"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Map":`, "/Map").withType(0, T[map[string]string]()),
	}, {
		name:    Name("Structs/Format/Invalid/Struct"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Struct: structAll{Bool: true}},
		want:    `{"Struct"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Struct":`, "/Struct").withType(0, T[structAll]()),
	}, {
		name:    Name("Structs/Format/Invalid/Slice"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Slice: []string{}},
		want:    `{"Slice"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Slice":`, "/Slice").withType(0, T[[]string]()),
	}, {
		name:    Name("Structs/Format/Invalid/Array"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Array: [1]string{"string"}},
		want:    `{"Array"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Array":`, "/Array").withType(0, T[[1]string]()),
	}, {
		name:    Name("Structs/Format/Invalid/Interface"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structFormatInvalid{Interface: "anything"},
		want:    `{"Interface"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"Interface":`, "/Interface").withType(0, T[any]()),
	}, {
		name: Name("Structs/Embed/Zero"),
		in:   structEmbedded{},
		want: `{"D":""}`,
	}, {
		name: Name("Structs/Embed/Alloc"),
		in: structEmbedded{
			X: structEmbeddedL1{
				X:            &structEmbeddedL2{},
				StructEmbed1: StructEmbed1{},
			},
			StructEmbed2: &StructEmbed2{},
		},
		want: `{"A":"","B":"","D":"","E":"","F":"","G":""}`,
	}, {
		name: Name("Structs/Embed/NonZero"),
		in: structEmbedded{
			X: structEmbeddedL1{
				X:            &structEmbeddedL2{A: "A1", B: "B1", C: "C1"},
				StructEmbed1: StructEmbed1{C: "C2", D: "D2", E: "E2"},
			},
			StructEmbed2: &StructEmbed2{E: "E3", F: "F3", G: "G3"},
		},
		want: `{"A":"A1","B":"B1","D":"D2","E":"E3","F":"F3","G":"G3"}`,
	}, {
		name: Name("Structs/Embed/DualCycle"),
		in: cyclicA{
			B1: cyclicB{F: 1}, // B1.F ignored since it conflicts with B2.F
			B2: cyclicB{F: 2}, // B2.F ignored since it conflicts with B1.F
		},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/Nil"),
		in:   structEmbedTextValue{X: jsontext.Value(nil)},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/Empty"),
		in:   structEmbedTextValue{X: jsontext.Value("")},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/NonEmptyN1"),
		in:   structEmbedTextValue{X: jsontext.Value(` { "fizz" : "buzz" } `)},
		want: `{"fizz":"buzz"}`,
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/NonEmptyN2"),
		in:   structEmbedTextValue{X: jsontext.Value(` { "fizz" : "buzz" , "foo" : "bar" } `)},
		want: `{"fizz":"buzz","foo":"bar"}`,
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/NonEmptyWithOthers"),
		in: structEmbedTextValue{
			A: 1,
			X: jsontext.Value(` { "fizz" : "buzz" , "foo" : "bar" } `),
			B: 2,
		},
		// NOTE: Embedded fallback fields are always serialized last.
		want: `{"A":1,"B":2,"fizz":"buzz","foo":"bar"}`,
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/RejectDuplicateNames"),
		opts:    []Options{jsontext.AllowDuplicateNames(false)},
		in:      structEmbedTextValue{X: jsontext.Value(` { "fizz" : "buzz" , "fizz" : "buzz" } `)},
		want:    `{"fizz":"buzz"`,
		wantErr: newDuplicateNameError("/fizz", nil, len64(`{"fizz":"buzz"`)),
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/AllowDuplicateNames"),
		opts: []Options{jsontext.AllowDuplicateNames(true)},
		in:   structEmbedTextValue{X: jsontext.Value(` { "fizz" : "buzz" , "fizz" : "buzz" } `)},
		want: `{"fizz":"buzz","fizz":"buzz"}`,
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/RejectInvalidUTF8"),
		opts:    []Options{jsontext.AllowInvalidUTF8(false)},
		in:      structEmbedTextValue{X: jsontext.Value(`{"` + "\xde\xad\xbe\xef" + `":"value"}`)},
		want:    `{`,
		wantErr: newInvalidUTF8Error(len64(`{"`+"\xde\xad"), ""),
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/AllowInvalidUTF8"),
		opts: []Options{jsontext.AllowInvalidUTF8(true)},
		in:   structEmbedTextValue{X: jsontext.Value(`{"` + "\xde\xad\xbe\xef" + `":"value"}`)},
		want: `{"ޭ��":"value"}`,
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/InvalidWhitespace"),
		in:      structEmbedTextValue{X: jsontext.Value("\n\r\t ")},
		want:    `{`,
		wantErr: EM(io.ErrUnexpectedEOF).withPos(`{`, "").withType(0, T[jsontext.Value]()),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/InvalidObject"),
		in:      structEmbedTextValue{X: jsontext.Value(` true `)},
		want:    `{`,
		wantErr: EM(errRawEmbedNotObject).withPos(`{`, "").withType(0, T[jsontext.Value]()),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/InvalidObjectName"),
		in:      structEmbedTextValue{X: jsontext.Value(` { true : false } `)},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(" { "), "")).withPos(`{`, "").withType(0, T[jsontext.Value]()),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/InvalidEndObject"),
		in:      structEmbedTextValue{X: jsontext.Value(` { "name" : false , } `)},
		want:    `{"name":false`,
		wantErr: EM(newInvalidCharacterError(",", "at start of value", len64(` { "name" : false `), "")).withPos(`{"name":false,`, "").withType(0, T[jsontext.Value]()),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/InvalidDualObject"),
		in:      structEmbedTextValue{X: jsontext.Value(`{}{}`)},
		want:    `{`,
		wantErr: EM(newInvalidCharacterError("{", "after top-level value", len64(`{}`), "")).withPos(`{`, "").withType(0, T[jsontext.Value]()),
	}, {
		name: Name("Structs/EmbeddedFallback/TextValue/Nested/Nil"),
		in:   structEmbedPointerEmbedTextValue{},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/PointerTextValue/Nil"),
		in:   structEmbedPointerTextValue{},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/PointerTextValue/NonEmpty"),
		in:   structEmbedPointerTextValue{X: addr(jsontext.Value(` { "fizz" : "buzz" } `))},
		want: `{"fizz":"buzz"}`,
	}, {
		name: Name("Structs/EmbeddedFallback/PointerTextValue/Nested/Nil"),
		in:   structEmbedEmbedPointerTextValue{},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringAny/Nil"),
		in:   structEmbedMapStringAny{X: nil},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringAny/Empty"),
		in:   structEmbedMapStringAny{X: make(jsonObject)},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringAny/NonEmptyN1"),
		in:   structEmbedMapStringAny{X: jsonObject{"fizz": nil}},
		want: `{"fizz":null}`,
	}, {
		name:         Name("Structs/EmbeddedFallback/MapStringAny/NonEmptyN2"),
		in:           structEmbedMapStringAny{X: jsonObject{"fizz": time.Time{}, "buzz": math.Pi}},
		want:         `{"buzz":3.141592653589793,"fizz":"0001-01-01T00:00:00Z"}`,
		canonicalize: true,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringAny/NonEmptyWithOthers"),
		in: structEmbedMapStringAny{
			A: 1,
			X: jsonObject{"fizz": nil},
			B: 2,
		},
		// NOTE: Embedded fallback fields are always serialized last.
		want: `{"A":1,"B":2,"fizz":null}`,
	}, {
		name:    Name("Structs/EmbeddedFallback/MapStringAny/RejectInvalidUTF8"),
		opts:    []Options{jsontext.AllowInvalidUTF8(false)},
		in:      structEmbedMapStringAny{X: jsonObject{"\xde\xad\xbe\xef": nil}},
		want:    `{`,
		wantErr: EM(errInvalidUTF8).withPos(`{`, "").withType(0, stringType),
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringAny/AllowInvalidUTF8"),
		opts: []Options{jsontext.AllowInvalidUTF8(true)},
		in:   structEmbedMapStringAny{X: jsonObject{"\xde\xad\xbe\xef": nil}},
		want: `{"ޭ��":null}`,
	}, {
		name:    Name("Structs/EmbeddedFallback/MapStringAny/InvalidValue"),
		opts:    []Options{jsontext.AllowInvalidUTF8(true)},
		in:      structEmbedMapStringAny{X: jsonObject{"name": make(chan string)}},
		want:    `{"name"`,
		wantErr: EM(nil).withPos(`{"name":`, "/name").withType(0, T[chan string]()),
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringAny/Nested/Nil"),
		in:   structEmbedPointerEmbedMapStringAny{},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringAny/MarshalFunc"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v float64) ([]byte, error) {
				return []byte(fmt.Sprintf(`"%v"`, v)), nil
			})),
		},
		in:   structEmbedMapStringAny{X: jsonObject{"fizz": 3.14159}},
		want: `{"fizz":"3.14159"}`,
	}, {
		name: Name("Structs/EmbeddedFallback/PointerMapStringAny/Nil"),
		in:   structEmbedPointerMapStringAny{X: nil},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/PointerMapStringAny/NonEmpty"),
		in:   structEmbedPointerMapStringAny{X: addr(jsonObject{"name": "value"})},
		want: `{"name":"value"}`,
	}, {
		name: Name("Structs/EmbeddedFallback/PointerMapStringAny/Nested/Nil"),
		in:   structEmbedEmbedPointerMapStringAny{},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt"),
		in: structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 1, "two": 2},
		},
		want:         `{"one":1,"two":2,"zero":0}`,
		canonicalize: true,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt/Deterministic"),
		opts: []Options{Deterministic(true)},
		in: structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 1, "two": 2},
		},
		want: `{"one":1,"two":2,"zero":0}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt/Deterministic+AllowInvalidUTF8+RejectDuplicateNames"),
		opts: []Options{Deterministic(true), jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(false)},
		in: structEmbedMapStringInt{
			X: map[string]int{"\xff": 0, "\xfe": 1},
		},
		want:    `{"�":1`,
		wantErr: newDuplicateNameError("", []byte(`"�"`), len64(`{"�":1`)),
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt/Deterministic+AllowInvalidUTF8+AllowDuplicateNames"),
		opts: []Options{Deterministic(true), jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true)},
		in: structEmbedMapStringInt{
			X: map[string]int{"\xff": 0, "\xfe": 1},
		},
		want: `{"�":1,"�":0}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt/StringifiedNumbers"),
		opts: []Options{StringifyNumbers(true)},
		in: structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 1, "two": 2},
		},
		want:         `{"one":"1","two":"2","zero":"0"}`,
		canonicalize: true,
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt/MarshalFunc"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				// Marshalers do not affect the string key of embedded maps.
				MarshalFunc(func(v string) ([]byte, error) {
					return []byte(fmt.Sprintf(`"%q"`, strings.ToUpper(v))), nil
				}),
				MarshalFunc(func(v int) ([]byte, error) {
					return []byte(fmt.Sprintf(`"%v"`, v)), nil
				}),
			)),
		},
		in: structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 1, "two": 2},
		},
		want:         `{"one":"1","two":"2","zero":"0"}`,
		canonicalize: true,
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringInt"),
		in: structEmbedMapNamedStringInt{
			X: map[namedString]int{"zero": 0, "one": 1, "two": 2},
		},
		want:         `{"one":1,"two":2,"zero":0}`,
		canonicalize: true,
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringInt/Deterministic"),
		opts: []Options{Deterministic(true)},
		in: structEmbedMapNamedStringInt{
			X: map[namedString]int{"zero": 0, "one": 1, "two": 2},
		},
		want: `{"one":1,"two":2,"zero":0}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringAny/Nil"),
		in:   structEmbedMapNamedStringAny{X: nil},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringAny/Empty"),
		in:   structEmbedMapNamedStringAny{X: make(map[namedString]any)},
		want: `{}`,
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringAny/NonEmptyN1"),
		in:   structEmbedMapNamedStringAny{X: map[namedString]any{"fizz": nil}},
		want: `{"fizz":null}`,
	}, {
		name:         Name("Structs/EmbeddedFallback/MapNamedStringAny/NonEmptyN2"),
		in:           structEmbedMapNamedStringAny{X: map[namedString]any{"fizz": time.Time{}, "buzz": math.Pi}},
		want:         `{"buzz":3.141592653589793,"fizz":"0001-01-01T00:00:00Z"}`,
		canonicalize: true,
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringAny/NonEmptyWithOthers"),
		in: structEmbedMapNamedStringAny{
			A: 1,
			X: map[namedString]any{"fizz": nil},
			B: 2,
		},
		// NOTE: Embedded fallback fields are always serialized last.
		want: `{"A":1,"B":2,"fizz":null}`,
	}, {
		name:    Name("Structs/EmbeddedFallback/MapNamedStringAny/RejectInvalidUTF8"),
		opts:    []Options{jsontext.AllowInvalidUTF8(false)},
		in:      structEmbedMapNamedStringAny{X: map[namedString]any{"\xde\xad\xbe\xef": nil}},
		want:    `{`,
		wantErr: EM(errInvalidUTF8).withPos(`{`, "").withType(0, T[namedString]()),
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringAny/AllowInvalidUTF8"),
		opts: []Options{jsontext.AllowInvalidUTF8(true)},
		in:   structEmbedMapNamedStringAny{X: map[namedString]any{"\xde\xad\xbe\xef": nil}},
		want: `{"ޭ��":null}`,
	}, {
		name:    Name("Structs/EmbeddedFallback/MapNamedStringAny/InvalidValue"),
		opts:    []Options{jsontext.AllowInvalidUTF8(true)},
		in:      structEmbedMapNamedStringAny{X: map[namedString]any{"name": make(chan string)}},
		want:    `{"name"`,
		wantErr: EM(nil).withPos(`{"name":`, "/name").withType(0, T[chan string]()),
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringAny/MarshalFunc"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v float64) ([]byte, error) {
				return []byte(fmt.Sprintf(`"%v"`, v)), nil
			})),
		},
		in:   structEmbedMapNamedStringAny{X: map[namedString]any{"fizz": 3.14159}},
		want: `{"fizz":"3.14159"}`,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/Other"),
		in: structNoCaseEmbedTextValue{
			X: jsontext.Value(`{"dupe":"","dupe":""}`),
		},
		want:    `{"dupe":""`,
		wantErr: newDuplicateNameError("", []byte(`"dupe"`), len64(`{"dupe":""`)),
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/Other/AllowDuplicateNames"),
		opts: []Options{jsontext.AllowDuplicateNames(true)},
		in: structNoCaseEmbedTextValue{
			X: jsontext.Value(`{"dupe": "", "dupe": ""}`),
		},
		want: `{"dupe":"","dupe":""}`,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/ExactDifferent"),
		in: structNoCaseEmbedTextValue{
			X: jsontext.Value(`{"Aaa": "", "AaA": "", "AAa": "", "AAA": ""}`),
		},
		want: `{"Aaa":"","AaA":"","AAa":"","AAA":""}`,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/ExactConflict"),
		in: structNoCaseEmbedTextValue{
			X: jsontext.Value(`{"Aaa": "", "Aaa": ""}`),
		},
		want:    `{"Aaa":""`,
		wantErr: newDuplicateNameError("", []byte(`"Aaa"`), len64(`{"Aaa":""`)),
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/ExactConflict/AllowDuplicateNames"),
		opts: []Options{jsontext.AllowDuplicateNames(true)},
		in: structNoCaseEmbedTextValue{
			X: jsontext.Value(`{"Aaa": "", "Aaa": ""}`),
		},
		want: `{"Aaa":"","Aaa":""}`,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/NoCaseConflict"),
		in: structNoCaseEmbedTextValue{
			X: jsontext.Value(`{"Aaa": "", "AaA": "", "aaa": ""}`),
		},
		want:    `{"Aaa":"","AaA":""`,
		wantErr: newDuplicateNameError("", []byte(`"aaa"`), len64(`{"Aaa":"","AaA":""`)),
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/NoCaseConflict/AllowDuplicateNames"),
		opts: []Options{jsontext.AllowDuplicateNames(true)},
		in: structNoCaseEmbedTextValue{
			X: jsontext.Value(`{"Aaa": "", "AaA": "", "aaa": ""}`),
		},
		want: `{"Aaa":"","AaA":"","aaa":""}`,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/ExactDifferentWithField"),
		in: structNoCaseEmbedTextValue{
			AAA: "x",
			AaA: "x",
			X:   jsontext.Value(`{"Aaa": ""}`),
		},
		want: `{"AAA":"x","AaA":"x","Aaa":""}`,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/ExactConflictWithField"),
		in: structNoCaseEmbedTextValue{
			AAA: "x",
			AaA: "x",
			X:   jsontext.Value(`{"AAA": ""}`),
		},
		want:    `{"AAA":"x","AaA":"x"`,
		wantErr: newDuplicateNameError("", []byte(`"AAA"`), len64(`{"AAA":"x","AaA":"x"`)),
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedTextValue/NoCaseConflictWithField"),
		in: structNoCaseEmbedTextValue{
			AAA: "x",
			AaA: "x",
			X:   jsontext.Value(`{"aaa": ""}`),
		},
		want:    `{"AAA":"x","AaA":"x"`,
		wantErr: newDuplicateNameError("", []byte(`"aaa"`), len64(`{"AAA":"x","AaA":"x"`)),
	}, {
		name: Name("Structs/DuplicateName/MatchCaseInsensitiveDelimiter"),
		in: structNoCaseEmbedTextValue{
			AaA: "x",
			X:   jsontext.Value(`{"aa_a": ""}`),
		},
		want:    `{"AaA":"x"`,
		wantErr: newDuplicateNameError("", []byte(`"aa_a"`), len64(`{"AaA":"x"`)),
	}, {
		name: Name("Structs/DuplicateName/MatchCaseSensitiveDelimiter"),
		opts: []Options{unsupported("MatchCaseSensitiveDelimiter")},
		in: structNoCaseEmbedTextValue{
			AaA: "x",
			X:   jsontext.Value(`{"aa_a": ""}`),
		},
		want: `{"AaA":"x","aa_a":""}`,
	}, {
		name: Name("Structs/DuplicateName/MatchCaseInsensitiveNames+MatchCaseSensitiveDelimiter"),
		opts: []Options{MatchCaseInsensitiveNames(true), unsupported("MatchCaseSensitiveDelimiter")},
		in: structNoCaseEmbedTextValue{
			AaA: "x",
			X:   jsontext.Value(`{"aa_a": ""}`),
		},
		want: `{"AaA":"x","aa_a":""}`,
	}, {
		name: Name("Structs/DuplicateName/MatchCaseInsensitiveNames+MatchCaseSensitiveDelimiter"),
		opts: []Options{MatchCaseInsensitiveNames(true), unsupported("MatchCaseSensitiveDelimiter")},
		in: structNoCaseEmbedTextValue{
			AA_b: "x",
			X:    jsontext.Value(`{"aa_b": ""}`),
		},
		want:    `{"AA_b":"x"`,
		wantErr: newDuplicateNameError("", []byte(`"aa_b"`), len64(`{"AA_b":"x"`)),
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedMapStringAny/ExactDifferent"),
		in: structNoCaseEmbedMapStringAny{
			X: jsonObject{"Aaa": "", "AaA": "", "AAa": "", "AAA": ""},
		},
		want:         `{"AAA":"","AAa":"","AaA":"","Aaa":""}`,
		canonicalize: true,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedMapStringAny/ExactDifferentWithField"),
		in: structNoCaseEmbedMapStringAny{
			AAA: "x",
			AaA: "x",
			X:   jsonObject{"Aaa": ""},
		},
		want: `{"AAA":"x","AaA":"x","Aaa":""}`,
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedMapStringAny/ExactConflictWithField"),
		in: structNoCaseEmbedMapStringAny{
			AAA: "x",
			AaA: "x",
			X:   jsonObject{"AAA": ""},
		},
		want:    `{"AAA":"x","AaA":"x"`,
		wantErr: newDuplicateNameError("", []byte(`"AAA"`), len64(`{"AAA":"x","AaA":"x"`)),
	}, {
		name: Name("Structs/DuplicateName/NoCaseEmbedMapStringAny/NoCaseConflictWithField"),
		in: structNoCaseEmbedMapStringAny{
			AAA: "x",
			AaA: "x",
			X:   jsonObject{"aaa": ""},
		},
		want:    `{"AAA":"x","AaA":"x"`,
		wantErr: newDuplicateNameError("", []byte(`"aaa"`), len64(`{"AAA":"x","AaA":"x"`)),
	}, {
		name:    Name("Structs/Invalid/Conflicting"),
		in:      structConflicting{},
		want:    ``,
		wantErr: EM(errors.New("Go struct fields A and B conflict over JSON object name \"conflict\"")).withType(0, T[structConflicting]()),
	}, {
		name:    Name("Structs/Invalid/NoneExported"),
		in:      structNoneExported{},
		want:    ``,
		wantErr: EM(errNoExportedFields).withType(0, T[structNoneExported]()),
	}, {
		name:    Name("Structs/Invalid/MalformedTag"),
		in:      structMalformedTag{},
		want:    ``,
		wantErr: EM(errors.New("Go struct field Malformed has malformed `json` tag: invalid character '\"' at start of option (expecting Unicode letter)")).withType(0, T[structMalformedTag]()),
	}, {
		name:    Name("Structs/Invalid/UnexportedTag"),
		in:      structUnexportedTag{},
		want:    ``,
		wantErr: EM(errors.New("unexported Go struct field unexported cannot have non-ignored `json:\"name\"` tag")).withType(0, T[structUnexportedTag]()),
	}, {
		name:    Name("Structs/Invalid/ExportedEmbedded"),
		in:      structExportedEmbedded{"hello"},
		want:    ``,
		wantErr: EM(errors.New("embedded Go struct field NamedString of non-struct type must be explicitly given a JSON name")).withType(0, T[structExportedEmbedded]()),
	}, {
		name: Name("Structs/Valid/ExportedEmbedded"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structExportedEmbedded{"hello"},
		want: `{"NamedString":"hello"}`,
	}, {
		name: Name("Structs/Valid/ExportedEmbeddedTag"),
		in:   structExportedEmbeddedTag{"hello"},
		want: `{"name":"hello"}`,
	}, {
		name:    Name("Structs/Invalid/UnexportedEmbedded"),
		in:      structUnexportedEmbedded{},
		want:    ``,
		wantErr: EM(errors.New("embedded Go struct field namedString of non-struct type must be explicitly given a JSON name")).withType(0, T[structUnexportedEmbedded]()),
	}, {
		name: Name("Structs/Valid/UnexportedEmbedded"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structUnexportedEmbedded{},
		want: `{}`,
	}, {
		name:    Name("Structs/Invalid/UnexportedEmbeddedTag"),
		in:      structUnexportedEmbeddedTag{},
		wantErr: EM(errors.New("Go struct field namedString is not exported")).withType(0, T[structUnexportedEmbeddedTag]()),
	}, {
		name: Name("Structs/Valid/UnexportedEmbeddedTag"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structUnexportedEmbeddedTag{},
		want: `{}`,
	}, {
		name: Name("Structs/Invalid/UnexportedEmbeddedMethodTag"),
		opts: []Options{unsupported("ReportErrorsWithLegacySemantics")},
		in:   structUnexportedEmbeddedMethodTag{},
		want: `{}`,
	}, {
		name: Name("Structs/UnexportedEmbeddedStruct/Zero"),
		in:   structUnexportedEmbeddedStruct{},
		want: `{"FizzBuzz":0,"Addr":""}`,
	}, {
		name: Name("Structs/UnexportedEmbeddedStruct/NonZero"),
		in:   structUnexportedEmbeddedStruct{structOmitZeroAll{Bool: true}, 5, structNestedAddr{netip.AddrFrom4([4]byte{192, 168, 0, 1})}},
		want: `{"Bool":true,"FizzBuzz":5,"Addr":"192.168.0.1"}`,
	}, {
		name: Name("Structs/UnexportedEmbeddedStructPointer/Nil"),
		in:   structUnexportedEmbeddedStructPointer{},
		want: `{"FizzBuzz":0}`,
	}, {
		name: Name("Structs/UnexportedEmbeddedStructPointer/Zero"),
		in:   structUnexportedEmbeddedStructPointer{&structOmitZeroAll{}, 0, &structNestedAddr{}},
		want: `{"FizzBuzz":0,"Addr":""}`,
	}, {
		name: Name("Structs/UnexportedEmbeddedStructPointer/NonZero"),
		in:   structUnexportedEmbeddedStructPointer{&structOmitZeroAll{Bool: true}, 5, &structNestedAddr{netip.AddrFrom4([4]byte{192, 168, 0, 1})}},
		want: `{"Bool":true,"FizzBuzz":5,"Addr":"192.168.0.1"}`,
	}, {
		name: Name("Structs/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   struct{}{},
		want: `{}`,
	}, {
		name: Name("Slices/Interface"),
		in: []any{
			false, true,
			"hello", []byte("world"),
			int32(-32), namedInt64(-64),
			uint32(+32), namedUint64(+64),
			float32(32.32), namedFloat64(64.64),
		},
		want: `[false,true,"hello","d29ybGQ=",-32,-64,32,64,32.32,64.64]`,
	}, {
		name:    Name("Slices/Invalid/Channel"),
		in:      [](chan string){nil},
		want:    `[`,
		wantErr: EM(nil).withPos(`[`, "/0").withType(0, T[chan string]()),
	}, {
		name: Name("Slices/RecursiveSlice"),
		in: recursiveSlice{
			nil,
			{},
			{nil},
			{nil, {}},
		},
		want: `[[],[],[[]],[[],[]]]`,
	}, {
		name: Name("Slices/CyclicSlice"),
		in: func() recursiveSlice {
			s := recursiveSlice{{}}
			s[0] = s
			return s
		}(),
		want:    strings.Repeat(`[`, startDetectingCyclesAfter) + `[`,
		wantErr: EM(errCycle).withPos(strings.Repeat("[", startDetectingCyclesAfter+1), jsontext.Pointer(strings.Repeat("/0", startDetectingCyclesAfter+1))).withType(0, T[recursiveSlice]()),
	}, {
		name: Name("Slices/NonCyclicSlice"),
		in: func() []any {
			v := []any{nil, nil}
			v[1] = v[:1]
			for i := 1000; i > 0; i-- {
				v = []any{v}
			}
			return v
		}(),
		want: strings.Repeat(`[`, startDetectingCyclesAfter) + `[null,[null]]` + strings.Repeat(`]`, startDetectingCyclesAfter),
	}, {
		name: Name("Slices/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   []string{"hello", "goodbye"},
		want: `["hello","goodbye"]`,
	}, {
		name: Name("Arrays/Empty"),
		in:   [0]struct{}{},
		want: `[]`,
	}, {
		name: Name("Arrays/Bool"),
		in:   [2]bool{false, true},
		want: `[false,true]`,
	}, {
		name: Name("Arrays/String"),
		in:   [2]string{"hello", "goodbye"},
		want: `["hello","goodbye"]`,
	}, {
		name: Name("Arrays/Bytes"),
		in:   [2][]byte{[]byte("hello"), []byte("goodbye")},
		want: `["aGVsbG8=","Z29vZGJ5ZQ=="]`,
	}, {
		name: Name("Arrays/Int"),
		in:   [2]int64{math.MinInt64, math.MaxInt64},
		want: `[-9223372036854775808,9223372036854775807]`,
	}, {
		name: Name("Arrays/Uint"),
		in:   [2]uint64{0, math.MaxUint64},
		want: `[0,18446744073709551615]`,
	}, {
		name: Name("Arrays/Float"),
		in:   [2]float64{-math.MaxFloat64, +math.MaxFloat64},
		want: `[-1.7976931348623157e+308,1.7976931348623157e+308]`,
	}, {
		name:    Name("Arrays/Invalid/Channel"),
		in:      new([1]chan string),
		want:    `[`,
		wantErr: EM(nil).withPos(`[`, "/0").withType(0, T[chan string]()),
	}, {
		name: Name("Arrays/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   [2]string{"hello", "goodbye"},
		want: `["hello","goodbye"]`,
	}, {
		name: Name("Pointers/NilL0"),
		in:   (*int)(nil),
		want: `null`,
	}, {
		name: Name("Pointers/NilL1"),
		in:   new(*int),
		want: `null`,
	}, {
		name: Name("Pointers/Bool"),
		in:   addr(addr(bool(true))),
		want: `true`,
	}, {
		name: Name("Pointers/String"),
		in:   addr(addr(string("string"))),
		want: `"string"`,
	}, {
		name: Name("Pointers/Bytes"),
		in:   addr(addr([]byte("bytes"))),
		want: `"Ynl0ZXM="`,
	}, {
		name: Name("Pointers/Int"),
		in:   addr(addr(int(-100))),
		want: `-100`,
	}, {
		name: Name("Pointers/Uint"),
		in:   addr(addr(uint(100))),
		want: `100`,
	}, {
		name: Name("Pointers/Float"),
		in:   addr(addr(float64(3.14159))),
		want: `3.14159`,
	}, {
		name: Name("Pointers/CyclicPointer"),
		in: func() *recursivePointer {
			p := new(recursivePointer)
			p.P = p
			return p
		}(),
		want:    strings.Repeat(`{"P":`, startDetectingCyclesAfter) + `{"P"`,
		wantErr: EM(errCycle).withPos(strings.Repeat(`{"P":`, startDetectingCyclesAfter+1), jsontext.Pointer(strings.Repeat("/P", startDetectingCyclesAfter+1))).withType(0, T[*recursivePointer]()),
	}, {
		name: Name("Pointers/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   addr(addr(bool(true))),
		want: `true`,
	}, {
		name: Name("Interfaces/Nil/Empty"),
		in:   [1]any{nil},
		want: `[null]`,
	}, {
		name: Name("Interfaces/Nil/NonEmpty"),
		in:   [1]io.Reader{nil},
		want: `[null]`,
	}, {
		name: Name("Interfaces/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   [1]io.Reader{nil},
		want: `[null]`,
	}, {
		name: Name("Interfaces/Any"),
		in:   struct{ X any }{[]any{nil, false, "", 0.0, map[string]any{}, []any{}, [8]byte{}}},
		want: `{"X":[null,false,"",0,{},[],"AAAAAAAAAAA="]}`,
	}, {
		name: Name("Interfaces/Any/Named"),
		in:   struct{ X namedAny }{[]namedAny{nil, false, "", 0.0, map[string]namedAny{}, []namedAny{}, [8]byte{}}},
		want: `{"X":[null,false,"",0,{},[],"AAAAAAAAAAA="]}`,
	}, {
		name: Name("Interfaces/Any/Stringified"),
		opts: []Options{StringifyNumbers(true)},
		in:   struct{ X any }{0.0},
		want: `{"X":"0"}`,
	}, {
		name: Name("Interfaces/Any/MarshalFunc/Any"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v any) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   struct{ X any }{[]any{nil, false, "", 0.0, map[string]any{}, []any{}}},
		want: `"called"`,
	}, {
		name: Name("Interfaces/Any/MarshalFunc/Bool"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v bool) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   struct{ X any }{[]any{nil, false, "", 0.0, map[string]any{}, []any{}}},
		want: `{"X":[null,"called","",0,{},[]]}`,
	}, {
		name: Name("Interfaces/Any/MarshalFunc/String"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v string) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   struct{ X any }{[]any{nil, false, "", 0.0, map[string]any{}, []any{}}},
		want: `{"X":[null,false,"called",0,{},[]]}`,
	}, {
		name: Name("Interfaces/Any/MarshalFunc/Float64"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v float64) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   struct{ X any }{[]any{nil, false, "", 0.0, map[string]any{}, []any{}}},
		want: `{"X":[null,false,"","called",{},[]]}`,
	}, {
		name: Name("Interfaces/Any/MarshalFunc/MapStringAny"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v map[string]any) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   struct{ X any }{[]any{nil, false, "", 0.0, map[string]any{}, []any{}}},
		want: `{"X":[null,false,"",0,"called",[]]}`,
	}, {
		name: Name("Interfaces/Any/MarshalFunc/SliceAny"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v []any) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   struct{ X any }{[]any{nil, false, "", 0.0, map[string]any{}, []any{}}},
		want: `{"X":"called"}`,
	}, {
		name: Name("Interfaces/Any/MarshalFunc/Bytes"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v [8]byte) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   struct{ X any }{[8]byte{}},
		want: `{"X":"called"}`,
	}, {
		name:    Name("Interfaces/Any/Float/NaN"),
		in:      struct{ X any }{math.NaN()},
		want:    `{"X"`,
		wantErr: EM(fmt.Errorf("unsupported value: %v", math.NaN())).withType(0, reflect.TypeFor[float64]()).withPos(`{"X":`, "/X"),
	}, {
		name: Name("Interfaces/Any/Maps/Nil"),
		in:   struct{ X any }{map[string]any(nil)},
		want: `{"X":{}}`,
	}, {
		name: Name("Interfaces/Any/Maps/Nil/FormatNilMapAsNull"),
		opts: []Options{FormatNilMapAsNull(true)},
		in:   struct{ X any }{map[string]any(nil)},
		want: `{"X":null}`,
	}, {
		name: Name("Interfaces/Any/Maps/Empty"),
		in:   struct{ X any }{map[string]any{}},
		want: `{"X":{}}`,
	}, {
		name: Name("Interfaces/Any/Maps/Empty/Multiline"),
		opts: []Options{jsontext.Multiline(true), jsontext.WithIndent("")},
		in:   struct{ X any }{map[string]any{}},
		want: "{\n\"X\": {}\n}",
	}, {
		name: Name("Interfaces/Any/Maps/NonEmpty"),
		in:   struct{ X any }{map[string]any{"fizz": "buzz"}},
		want: `{"X":{"fizz":"buzz"}}`,
	}, {
		name: Name("Interfaces/Any/Maps/Deterministic"),
		opts: []Options{Deterministic(true)},
		in:   struct{ X any }{map[string]any{"alpha": "", "bravo": ""}},
		want: `{"X":{"alpha":"","bravo":""}}`,
	}, {
		name:    Name("Interfaces/Any/Maps/Deterministic+AllowInvalidUTF8+RejectDuplicateNames"),
		opts:    []Options{Deterministic(true), jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(false)},
		in:      struct{ X any }{map[string]any{"\xff": "", "\xfe": ""}},
		want:    `{"X":{"�":""`,
		wantErr: newDuplicateNameError("/X", []byte(`"�"`), len64(`{"X":{"�":"",`)),
	}, {
		name: Name("Interfaces/Any/Maps/Deterministic+AllowInvalidUTF8+AllowDuplicateNames"),
		opts: []Options{Deterministic(true), jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true)},
		in:   struct{ X any }{map[string]any{"\xff": "alpha", "\xfe": "bravo"}},
		want: `{"X":{"�":"bravo","�":"alpha"}}`,
	}, {
		name:    Name("Interfaces/Any/Maps/RejectInvalidUTF8"),
		in:      struct{ X any }{map[string]any{"\xff": "", "\xfe": ""}},
		want:    `{"X":{`,
		wantErr: newInvalidUTF8Error(len64(`{"X":{`), "/X"),
	}, {
		name:    Name("Interfaces/Any/Maps/AllowInvalidUTF8+RejectDuplicateNames"),
		opts:    []Options{jsontext.AllowInvalidUTF8(true)},
		in:      struct{ X any }{map[string]any{"\xff": "", "\xfe": ""}},
		want:    `{"X":{"�":""`,
		wantErr: newDuplicateNameError("/X", []byte(`"�"`), len64(`{"X":{"�":"",`)),
	}, {
		name: Name("Interfaces/Any/Maps/AllowInvalidUTF8+AllowDuplicateNames"),
		opts: []Options{jsontext.AllowInvalidUTF8(true), jsontext.AllowDuplicateNames(true)},
		in:   struct{ X any }{map[string]any{"\xff": "", "\xfe": ""}},
		want: `{"X":{"�":"","�":""}}`,
	}, {
		name: Name("Interfaces/Any/Maps/Cyclic"),
		in: func() any {
			m := map[string]any{}
			m[""] = m
			return struct{ X any }{m}
		}(),
		want:    `{"X"` + strings.Repeat(`:{""`, startDetectingCyclesAfter),
		wantErr: EM(errCycle).withPos(`{"X":`+strings.Repeat(`{"":`, startDetectingCyclesAfter), "/X"+jsontext.Pointer(strings.Repeat("/", startDetectingCyclesAfter))).withType(0, T[map[string]any]()),
	}, {
		name: Name("Interfaces/Any/Slices/Nil"),
		in:   struct{ X any }{[]any(nil)},
		want: `{"X":[]}`,
	}, {
		name: Name("Interfaces/Any/Slices/Nil/FormatNilSliceAsNull"),
		opts: []Options{FormatNilSliceAsNull(true)},
		in:   struct{ X any }{[]any(nil)},
		want: `{"X":null}`,
	}, {
		name: Name("Interfaces/Any/Slices/Empty"),
		in:   struct{ X any }{[]any{}},
		want: `{"X":[]}`,
	}, {
		name: Name("Interfaces/Any/Slices/Empty/Multiline"),
		opts: []Options{jsontext.Multiline(true), jsontext.WithIndent("")},
		in:   struct{ X any }{[]any{}},
		want: "{\n\"X\": []\n}",
	}, {
		name: Name("Interfaces/Any/Slices/NonEmpty"),
		in:   struct{ X any }{[]any{"fizz", "buzz"}},
		want: `{"X":["fizz","buzz"]}`,
	}, {
		name: Name("Interfaces/Any/Slices/Cyclic"),
		in: func() any {
			s := make([]any, 1)
			s[0] = s
			return struct{ X any }{s}
		}(),
		want:    `{"X":` + strings.Repeat(`[`, startDetectingCyclesAfter),
		wantErr: EM(errCycle).withPos(`{"X":`+strings.Repeat(`[`, startDetectingCyclesAfter), "/X"+jsontext.Pointer(strings.Repeat("/0", startDetectingCyclesAfter))).withType(0, T[[]any]()),
	}, {
		name: Name("Methods/NilPointer"),
		in:   struct{ X *allMethods }{X: (*allMethods)(nil)}, // method should not be called
		want: `{"X":null}`,
	}, {
		// NOTE: Fixes https://github.com/dominikh/go-tools/issues/975.
		name: Name("Methods/NilInterface"),
		in:   struct{ X MarshalerTo }{X: (*allMethods)(nil)}, // method should not be called
		want: `{"X":null}`,
	}, {
		name: Name("Methods/AllMethods"),
		in:   struct{ X *allMethods }{X: &allMethods{method: "MarshalJSONTo", value: []byte(`"hello"`)}},
		want: `{"X":"hello"}`,
	}, {
		name: Name("Methods/AllMethodsExceptJSONv2"),
		in:   struct{ X *allMethodsExceptJSONv2 }{X: &allMethodsExceptJSONv2{allMethods: allMethods{method: "MarshalJSON", value: []byte(`"hello"`)}}},
		want: `{"X":"hello"}`,
	}, {
		name: Name("Methods/AllMethodsExceptJSONv1"),
		in:   struct{ X *allMethodsExceptJSONv1 }{X: &allMethodsExceptJSONv1{allMethods: allMethods{method: "MarshalJSONTo", value: []byte(`"hello"`)}}},
		want: `{"X":"hello"}`,
	}, {
		name: Name("Methods/AllMethodsExceptText"),
		in:   struct{ X *allMethodsExceptText }{X: &allMethodsExceptText{allMethods: allMethods{method: "MarshalJSONTo", value: []byte(`"hello"`)}}},
		want: `{"X":"hello"}`,
	}, {
		name: Name("Methods/OnlyMethodJSONv2"),
		in:   struct{ X *onlyMethodJSONv2 }{X: &onlyMethodJSONv2{allMethods: allMethods{method: "MarshalJSONTo", value: []byte(`"hello"`)}}},
		want: `{"X":"hello"}`,
	}, {
		name: Name("Methods/OnlyMethodJSONv1"),
		in:   struct{ X *onlyMethodJSONv1 }{X: &onlyMethodJSONv1{allMethods: allMethods{method: "MarshalJSON", value: []byte(`"hello"`)}}},
		want: `{"X":"hello"}`,
	}, {
		name: Name("Methods/OnlyMethodText"),
		in:   struct{ X *onlyMethodText }{X: &onlyMethodText{allMethods: allMethods{method: "MarshalText", value: []byte(`hello`)}}},
		want: `{"X":"hello"}`,
	}, {
		name: Name("Methods/IP"),
		in:   net.IPv4(192, 168, 0, 100),
		want: `"192.168.0.100"`,
	}, {
		name: Name("Methods/NetIP"),
		in: struct {
			Addr     netip.Addr
			AddrPort netip.AddrPort
			Prefix   netip.Prefix
		}{
			Addr:     netip.AddrFrom4([4]byte{1, 2, 3, 4}),
			AddrPort: netip.AddrPortFrom(netip.AddrFrom4([4]byte{1, 2, 3, 4}), 1234),
			Prefix:   netip.PrefixFrom(netip.AddrFrom4([4]byte{1, 2, 3, 4}), 24),
		},
		want: `{"Addr":"1.2.3.4","AddrPort":"1.2.3.4:1234","Prefix":"1.2.3.4/24"}`,
	}, {
		// NOTE: Fixes https://go.dev/issue/46516.
		name: Name("Methods/Anonymous"),
		in:   struct{ X struct{ allMethods } }{X: struct{ allMethods }{allMethods{method: "MarshalJSONTo", value: []byte(`"hello"`)}}},
		want: `{"X":"hello"}`,
	}, {
		// NOTE: Fixes https://go.dev/issue/22967.
		name: Name("Methods/Addressable"),
		in: struct {
			V allMethods
			M map[string]allMethods
			I any
		}{
			V: allMethods{method: "MarshalJSONTo", value: []byte(`"hello"`)},
			M: map[string]allMethods{"K": {method: "MarshalJSONTo", value: []byte(`"hello"`)}},
			I: allMethods{method: "MarshalJSONTo", value: []byte(`"hello"`)},
		},
		want: `{"V":"hello","M":{"K":"hello"},"I":"hello"}`,
	}, {
		// NOTE: Fixes https://go.dev/issue/29732.
		name:         Name("Methods/MapKey/JSONv2"),
		in:           map[structMethodJSONv2]string{{"k1"}: "v1", {"k2"}: "v2"},
		want:         `{"k1":"v1","k2":"v2"}`,
		canonicalize: true,
	}, {
		// NOTE: Fixes https://go.dev/issue/29732.
		name:         Name("Methods/MapKey/JSONv1"),
		in:           map[structMethodJSONv1]string{{"k1"}: "v1", {"k2"}: "v2"},
		want:         `{"k1":"v1","k2":"v2"}`,
		canonicalize: true,
	}, {
		name:         Name("Methods/MapKey/Text"),
		in:           map[structMethodText]string{{"k1"}: "v1", {"k2"}: "v2"},
		want:         `{"k1":"v1","k2":"v2"}`,
		canonicalize: true,
	}, {
		name: Name("Methods/JSONv2/ErrUnsupported"),
		opts: []Options{Deterministic(true)},
		in:   unsupportedMethodJSONv2{"fizz": 123},
		want: `{"called":1,"fizz":123}`,
	}, {
		name: Name("Methods/Invalid/JSONv2/Error"),
		in: marshalJSONv2Func(func(*jsontext.Encoder) error {
			return errSomeError
		}),
		wantErr: EM(errSomeError).withType(0, T[marshalJSONv2Func]()),
	}, {
		name: Name("Methods/Invalid/JSONv2/TooFew"),
		in: marshalJSONv2Func(func(*jsontext.Encoder) error {
			return nil // do nothing
		}),
		wantErr: EM(errNonSingularValue).withType(0, T[marshalJSONv2Func]()),
	}, {
		name: Name("Methods/Invalid/JSONv2/TooMany"),
		in: marshalJSONv2Func(func(enc *jsontext.Encoder) error {
			enc.WriteToken(jsontext.Null)
			enc.WriteToken(jsontext.Null)
			return nil
		}),
		want:    `nullnull`,
		wantErr: EM(errNonSingularValue).withPos(`nullnull`, "").withType(0, T[marshalJSONv2Func]()),
	}, {
		name: Name("Methods/Invalid/JSONv2/ErrUnsupported"),
		in: marshalJSONv2Func(func(enc *jsontext.Encoder) error {
			return errors.ErrUnsupported
		}),
		wantErr: EM(nil).withType(0, T[marshalJSONv2Func]()),
	}, {
		name: Name("Methods/Invalid/JSONv1/Error"),
		in: marshalJSONv1Func(func() ([]byte, error) {
			return nil, errSomeError
		}),
		wantErr: EM(errSomeError).withType(0, T[marshalJSONv1Func]()),
	}, {
		name: Name("Methods/Invalid/JSONv1/Syntax"),
		in: marshalJSONv1Func(func() ([]byte, error) {
			return []byte("invalid"), nil
		}),
		wantErr: EM(newInvalidCharacterError("i", "at start of value", 0, "")).withType(0, T[marshalJSONv1Func]()),
	}, {
		name: Name("Methods/Invalid/JSONv1/ErrUnsupported"),
		in: marshalJSONv1Func(func() ([]byte, error) {
			return nil, errors.ErrUnsupported
		}),
		wantErr: EM(errors.New("MarshalJSON method may not return errors.ErrUnsupported")).withType(0, T[marshalJSONv1Func]()),
	}, {
		name: Name("Methods/AppendText"),
		in:   appendTextFunc(func(b []byte) ([]byte, error) { return append(b, "hello"...), nil }),
		want: `"hello"`,
	}, {
		name:    Name("Methods/AppendText/Error"),
		in:      appendTextFunc(func(b []byte) ([]byte, error) { return append(b, "hello"...), errSomeError }),
		wantErr: EM(errSomeError).withType(0, T[appendTextFunc]()),
	}, {
		name: Name("Methods/AppendText/NeedEscape"),
		in:   appendTextFunc(func(b []byte) ([]byte, error) { return append(b, `"`...), nil }),
		want: `"\""`,
	}, {
		name:    Name("Methods/AppendText/RejectInvalidUTF8"),
		in:      appendTextFunc(func(b []byte) ([]byte, error) { return append(b, "\xde\xad\xbe\xef"...), nil }),
		wantErr: EM(newInvalidUTF8Error(0, "")).withType(0, T[appendTextFunc]()),
	}, {
		name: Name("Methods/AppendText/AllowInvalidUTF8"),
		opts: []Options{jsontext.AllowInvalidUTF8(true)},
		in:   appendTextFunc(func(b []byte) ([]byte, error) { return append(b, "\xde\xad\xbe\xef"...), nil }),
		want: "\"\xde\xad\ufffd\ufffd\"",
	}, {
		name: Name("Methods/Invalid/Text/Error"),
		in: marshalTextFunc(func() ([]byte, error) {
			return nil, errSomeError
		}),
		wantErr: EM(errSomeError).withType(0, T[marshalTextFunc]()),
	}, {
		name: Name("Methods/Text/RejectInvalidUTF8"),
		in: marshalTextFunc(func() ([]byte, error) {
			return []byte("\xde\xad\xbe\xef"), nil
		}),
		wantErr: EM(newInvalidUTF8Error(0, "")).withType(0, T[marshalTextFunc]()),
	}, {
		name: Name("Methods/Text/AllowInvalidUTF8"),
		opts: []Options{jsontext.AllowInvalidUTF8(true)},
		in: marshalTextFunc(func() ([]byte, error) {
			return []byte("\xde\xad\xbe\xef"), nil
		}),
		want: "\"\xde\xad\ufffd\ufffd\"",
	}, {
		name: Name("Methods/Invalid/Text/ErrUnsupported"),
		in: marshalTextFunc(func() ([]byte, error) {
			return nil, errors.ErrUnsupported
		}),
		wantErr: EM(wrapErrUnsupported(errors.ErrUnsupported, "MarshalText method")).withType(0, T[marshalTextFunc]()),
	}, {
		name: Name("Methods/Invalid/MapKey/JSONv2/Syntax"),
		in: map[any]string{
			addr(marshalJSONv2Func(func(enc *jsontext.Encoder) error {
				return enc.WriteToken(jsontext.Null)
			})): "invalid",
		},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[marshalJSONv2Func]()),
	}, {
		name: Name("Methods/Invalid/MapKey/JSONv1/Syntax"),
		in: map[any]string{
			addr(marshalJSONv1Func(func() ([]byte, error) {
				return []byte(`null`), nil
			})): "invalid",
		},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[marshalJSONv1Func]()),
	}, {
		name: Name("Functions/Bool/V1"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(bool) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/Bool/Empty"),
		opts: []Options{WithMarshalers(nil)},
		in:   true,
		want: `true`,
	}, {
		name: Name("Functions/NamedBool/V1/NoMatch"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(namedBool) ([]byte, error) {
				return nil, errMustNotCall
			})),
		},
		in:   true,
		want: `true`,
	}, {
		name: Name("Functions/NamedBool/V1/Match"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(namedBool) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   namedBool(true),
		want: `"called"`,
	}, {
		name: Name("Functions/PointerBool/V1/Match"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v *bool) ([]byte, error) {
				_ = *v // must be a non-nil pointer
				return []byte(`"called"`), nil
			})),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/Bool/V2"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
				return enc.WriteToken(jsontext.String("called"))
			})),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/NamedBool/V2/NoMatch"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v namedBool) error {
				return errMustNotCall
			})),
		},
		in:   true,
		want: `true`,
	}, {
		name: Name("Functions/NamedBool/V2/Match"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v namedBool) error {
				return enc.WriteToken(jsontext.String("called"))
			})),
		},
		in:   namedBool(true),
		want: `"called"`,
	}, {
		name: Name("Functions/PointerBool/V2/Match"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v *bool) error {
				_ = *v // must be a non-nil pointer
				return enc.WriteToken(jsontext.String("called"))
			})),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/Bool/Empty1/NoMatch"),
		opts: []Options{
			WithMarshalers(new(Marshalers)),
		},
		in:   true,
		want: `true`,
	}, {
		name: Name("Functions/Bool/Empty2/NoMatch"),
		opts: []Options{
			WithMarshalers(JoinMarshalers()),
		},
		in:   true,
		want: `true`,
	}, {
		name: Name("Functions/Bool/V1/DirectError"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(bool) ([]byte, error) {
				return nil, errSomeError
			})),
		},
		in:      true,
		wantErr: EM(errSomeError).withType(0, T[bool]()),
	}, {
		name: Name("Functions/Bool/V1/SkipError"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(bool) ([]byte, error) {
				return nil, errors.ErrUnsupported
			})),
		},
		in:      true,
		wantErr: EM(wrapErrUnsupported(errors.ErrUnsupported, "marshal function of type func(T) ([]byte, error)")).withType(0, T[bool]()),
	}, {
		name: Name("Functions/Bool/V1/InvalidValue"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(bool) ([]byte, error) {
				return []byte("invalid"), nil
			})),
		},
		in:      true,
		wantErr: EM(newInvalidCharacterError("i", "at start of value", 0, "")).withType(0, T[bool]()),
	}, {
		name: Name("Functions/Bool/V2/DirectError"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
				return errSomeError
			})),
		},
		in:      true,
		wantErr: EM(errSomeError).withType(0, T[bool]()),
	}, {
		name: Name("Functions/Bool/V2/TooFew"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
				return nil
			})),
		},
		in:      true,
		wantErr: EM(errNonSingularValue).withType(0, T[bool]()),
	}, {
		name: Name("Functions/Bool/V2/TooMany"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
				enc.WriteValue([]byte(`"hello"`))
				enc.WriteValue([]byte(`"world"`))
				return nil
			})),
		},
		in:      true,
		want:    `"hello""world"`,
		wantErr: EM(errNonSingularValue).withPos(`"hello""world"`, "").withType(0, T[bool]()),
	}, {
		name: Name("Functions/Bool/V2/Skipped"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
				return errors.ErrUnsupported
			})),
		},
		in:   true,
		want: `true`,
	}, {
		name: Name("Functions/Bool/V2/ProcessBeforeSkip"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
				enc.WriteValue([]byte(`"hello"`))
				return errors.ErrUnsupported
			})),
		},
		in:      true,
		want:    `"hello"`,
		wantErr: EM(errUnsupportedMutation).withPos(`"hello"`, "").withType(0, T[bool]()),
	}, {
		name: Name("Functions/Bool/V2/WrappedUnsupportedError"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
				return fmt.Errorf("wrap: %w", errors.ErrUnsupported)
			})),
		},
		in:   true,
		want: `true`,
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V1"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v nocaseString) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   map[nocaseString]string{"hello": "world"},
		want: `{"called":"world"}`,
	}, {
		name: Name("Functions/Map/Key/PointerNoCaseString/V1"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v *nocaseString) ([]byte, error) {
				_ = *v // must be a non-nil pointer
				return []byte(`"called"`), nil
			})),
		},
		in:   map[nocaseString]string{"hello": "world"},
		want: `{"called":"world"}`,
	}, {
		name: Name("Functions/Map/Key/TextMarshaler/V1"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v encoding.TextMarshaler) ([]byte, error) {
				_ = *v.(*nocaseString) // must be a non-nil *nocaseString
				return []byte(`"called"`), nil
			})),
		},
		in:   map[nocaseString]string{"hello": "world"},
		want: `{"called":"world"}`,
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V1/InvalidValue"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v nocaseString) ([]byte, error) {
				return []byte(`null`), nil
			})),
		},
		in:      map[nocaseString]string{"hello": "world"},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[nocaseString]()),
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V2/InvalidKind"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v nocaseString) ([]byte, error) {
				return []byte(`null`), nil
			})),
		},
		in:      map[nocaseString]string{"hello": "world"},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[nocaseString]()),
	}, {
		name: Name("Functions/Map/Key/String/V1/DuplicateName"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v string) ([]byte, error) {
				return []byte(`"name"`), nil
			})),
		},
		in:   map[string]string{"name1": "value", "name2": "value"},
		want: `{"name":"name"`,
		wantErr: EM(newDuplicateNameError("", []byte(`"name"`), len64(`{"name":"name",`))).
			withPos(`{"name":"name",`, "").withType(0, T[string]()),
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V2"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v nocaseString) error {
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   map[nocaseString]string{"hello": "world"},
		want: `{"called":"world"}`,
	}, {
		name: Name("Functions/Map/Key/PointerNoCaseString/V2"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v *nocaseString) error {
				_ = *v // must be a non-nil pointer
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   map[nocaseString]string{"hello": "world"},
		want: `{"called":"world"}`,
	}, {
		name: Name("Functions/Map/Key/TextMarshaler/V2"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v encoding.TextMarshaler) error {
				_ = *v.(*nocaseString) // must be a non-nil *nocaseString
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   map[nocaseString]string{"hello": "world"},
		want: `{"called":"world"}`,
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V2/InvalidToken"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v nocaseString) error {
				return enc.WriteToken(jsontext.Null)
			})),
		},
		in:      map[nocaseString]string{"hello": "world"},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[nocaseString]()),
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V2/InvalidValue"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v nocaseString) error {
				return enc.WriteValue([]byte(`null`))
			})),
		},
		in:      map[nocaseString]string{"hello": "world"},
		want:    `{`,
		wantErr: EM(newNonStringNameError(len64(`{`), "")).withPos(`{`, "").withType(0, T[nocaseString]()),
	}, {
		name: Name("Functions/Map/Value/NoCaseString/V1"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v nocaseString) ([]byte, error) {
				return []byte(`"called"`), nil
			})),
		},
		in:   map[string]nocaseString{"hello": "world"},
		want: `{"hello":"called"}`,
	}, {
		name: Name("Functions/Map/Value/PointerNoCaseString/V1"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v *nocaseString) ([]byte, error) {
				_ = *v // must be a non-nil pointer
				return []byte(`"called"`), nil
			})),
		},
		in:   map[string]nocaseString{"hello": "world"},
		want: `{"hello":"called"}`,
	}, {
		name: Name("Functions/Map/Value/TextMarshaler/V1"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v encoding.TextMarshaler) ([]byte, error) {
				_ = *v.(*nocaseString) // must be a non-nil *nocaseString
				return []byte(`"called"`), nil
			})),
		},
		in:   map[string]nocaseString{"hello": "world"},
		want: `{"hello":"called"}`,
	}, {
		name: Name("Functions/Map/Value/NoCaseString/V2"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v nocaseString) error {
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   map[string]nocaseString{"hello": "world"},
		want: `{"hello":"called"}`,
	}, {
		name: Name("Functions/Map/Value/PointerNoCaseString/V2"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v *nocaseString) error {
				_ = *v // must be a non-nil pointer
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   map[string]nocaseString{"hello": "world"},
		want: `{"hello":"called"}`,
	}, {
		name: Name("Functions/Map/Value/TextMarshaler/V2"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v encoding.TextMarshaler) error {
				_ = *v.(*nocaseString) // must be a non-nil *nocaseString
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   map[string]nocaseString{"hello": "world"},
		want: `{"hello":"called"}`,
	}, {
		name: Name("Funtions/Struct/Fields"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalFunc(func(v bool) ([]byte, error) {
					return []byte(`"called1"`), nil
				}),
				MarshalFunc(func(v *string) ([]byte, error) {
					return []byte(`"called2"`), nil
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v []byte) error {
					return enc.WriteValue([]byte(`"called3"`))
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v *int64) error {
					return enc.WriteValue([]byte(`"called4"`))
				}),
			)),
		},
		in:   structScalars{},
		want: `{"Bool":"called1","String":"called2","Bytes":"called3","Int":"called4","Uint":0,"Float":0}`,
	}, {
		name: Name("Functions/Struct/OmitEmpty"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalFunc(func(v bool) ([]byte, error) {
					return []byte(`null`), nil
				}),
				MarshalFunc(func(v string) ([]byte, error) {
					return []byte(`"called1"`), nil
				}),
				MarshalFunc(func(v *stringMarshalNonEmpty) ([]byte, error) {
					return []byte(`""`), nil
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v bytesMarshalNonEmpty) error {
					return enc.WriteValue([]byte(`{}`))
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v *float64) error {
					return enc.WriteValue([]byte(`[]`))
				}),
				MarshalFunc(func(v mapMarshalNonEmpty) ([]byte, error) {
					return []byte(`"called2"`), nil
				}),
				MarshalFunc(func(v []string) ([]byte, error) {
					return []byte(`"called3"`), nil
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v *sliceMarshalNonEmpty) error {
					return enc.WriteValue([]byte(`"called4"`))
				}),
			)),
		},
		in:   structOmitEmptyAll{},
		want: `{"String":"called1","MapNonEmpty":"called2","Slice":"called3","SliceNonEmpty":"called4"}`,
	}, {
		name: Name("Functions/Struct/OmitZero"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalFunc(func(v bool) ([]byte, error) {
					panic("should not be called")
				}),
				MarshalFunc(func(v *string) ([]byte, error) {
					panic("should not be called")
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v []byte) error {
					panic("should not be called")
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v *int64) error {
					panic("should not be called")
				}),
			)),
		},
		in:   structOmitZeroAll{},
		want: `{}`,
	}, {
		name: Name("Functions/Struct/Embedded"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalFunc(func(v structEmbeddedL1) ([]byte, error) {
					panic("should not be called")
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v *StructEmbed2) error {
					panic("should not be called")
				}),
			)),
		},
		in:   structEmbedded{},
		want: `{"D":""}`,
	}, {
		name: Name("Functions/Slice/Elem"),
		opts: []Options{
			WithMarshalers(MarshalFunc(func(v bool) ([]byte, error) {
				return []byte(`"` + strconv.FormatBool(v) + `"`), nil
			})),
		},
		in:   []bool{true, false},
		want: `["true","false"]`,
	}, {
		name: Name("Functions/Array/Elem"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v *bool) error {
				return enc.WriteValue([]byte(`"` + strconv.FormatBool(*v) + `"`))
			})),
		},
		in:   [2]bool{true, false},
		want: `["true","false"]`,
	}, {
		name: Name("Functions/Pointer/Nil"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v *bool) error {
				panic("should not be called")
			})),
		},
		in:   struct{ X *bool }{nil},
		want: `{"X":null}`,
	}, {
		name: Name("Functions/Pointer/NonNil"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v *bool) error {
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   struct{ X *bool }{addr(false)},
		want: `{"X":"called"}`,
	}, {
		name: Name("Functions/Interface/Nil"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v fmt.Stringer) error {
				panic("should not be called")
			})),
		},
		in:   struct{ X fmt.Stringer }{nil},
		want: `{"X":null}`,
	}, {
		name: Name("Functions/Interface/NonNil/MatchInterface"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v fmt.Stringer) error {
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   struct{ X fmt.Stringer }{valueStringer{}},
		want: `{"X":"called"}`,
	}, {
		name: Name("Functions/Interface/NonNil/MatchConcrete"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v valueStringer) error {
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   struct{ X fmt.Stringer }{valueStringer{}},
		want: `{"X":"called"}`,
	}, {
		name: Name("Functions/Interface/NonNil/MatchPointer"),
		opts: []Options{
			WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, v *valueStringer) error {
				return enc.WriteValue([]byte(`"called"`))
			})),
		},
		in:   struct{ X fmt.Stringer }{valueStringer{}},
		want: `{"X":"called"}`,
	}, {
		name: Name("Functions/Interface/Any"),
		opts: []Options{
			WithMarshalers(func() *Marshalers {
				type P struct {
					D int
					N int64
				}
				type PV struct {
					P P
					V any
				}

				var lastChecks []func() error
				checkLast := func() error {
					for _, fn := range lastChecks {
						if err := fn(); err != nil {
							return err
						}
					}
					return errors.ErrUnsupported
				}
				makeValueChecker := func(name string, want []PV) func(e *jsontext.Encoder, v any) error {
					checkNext := func(e *jsontext.Encoder, v any) error {
						d, n := stackPosition(e)
						p := P{d, n}
						rv := reflect.ValueOf(v)
						pv := PV{p, v}
						switch {
						case len(want) == 0:
							return fmt.Errorf("%s: %v: got more values than expected", name, p)
						case !rv.IsValid() || rv.Kind() != reflect.Pointer || rv.IsNil():
							return fmt.Errorf("%s: %v: got %#v, want non-nil pointer type", name, p, v)
						case !reflect.DeepEqual(pv, want[0]):
							return fmt.Errorf("%s:\n\tgot  %#v\n\twant %#v", name, pv, want[0])
						default:
							want = want[1:]
							return errors.ErrUnsupported
						}
					}
					lastChecks = append(lastChecks, func() error {
						if len(want) > 0 {
							return fmt.Errorf("%s: did not get enough values, want %d more", name, len(want))
						}
						return nil
					})
					return checkNext
				}
				makePositionChecker := func(name string, want []P) func(e *jsontext.Encoder, v any) error {
					checkNext := func(e *jsontext.Encoder, v any) error {
						d, n := stackPosition(e)
						p := P{d, n}
						switch {
						case len(want) == 0:
							return fmt.Errorf("%s: %v: got more values than wanted", name, p)
						case p != want[0]:
							return fmt.Errorf("%s: got %v, want %v", name, p, want[0])
						default:
							want = want[1:]
							return errors.ErrUnsupported
						}
					}
					lastChecks = append(lastChecks, func() error {
						if len(want) > 0 {
							return fmt.Errorf("%s: did not get enough values, want %d more", name, len(want))
						}
						return nil
					})
					return checkNext
				}

				wantAny := []PV{
					{P{0, 0}, addr([]any{
						nil,
						valueStringer{},
						(*valueStringer)(nil),
						addr(valueStringer{}),
						(**valueStringer)(nil),
						addr((*valueStringer)(nil)),
						addr(addr(valueStringer{})),
						pointerStringer{},
						(*pointerStringer)(nil),
						addr(pointerStringer{}),
						(**pointerStringer)(nil),
						addr((*pointerStringer)(nil)),
						addr(addr(pointerStringer{})),
						"LAST",
					})},
					{P{1, 0}, addr(any(nil))},
					{P{1, 1}, addr(any(valueStringer{}))},
					{P{1, 1}, addr(valueStringer{})},
					{P{1, 2}, addr(any((*valueStringer)(nil)))},
					{P{1, 2}, addr((*valueStringer)(nil))},
					{P{1, 3}, addr(any(addr(valueStringer{})))},
					{P{1, 3}, addr(addr(valueStringer{}))},
					{P{1, 3}, addr(valueStringer{})},
					{P{1, 4}, addr(any((**valueStringer)(nil)))},
					{P{1, 4}, addr((**valueStringer)(nil))},
					{P{1, 5}, addr(any(addr((*valueStringer)(nil))))},
					{P{1, 5}, addr(addr((*valueStringer)(nil)))},
					{P{1, 5}, addr((*valueStringer)(nil))},
					{P{1, 6}, addr(any(addr(addr(valueStringer{}))))},
					{P{1, 6}, addr(addr(addr(valueStringer{})))},
					{P{1, 6}, addr(addr(valueStringer{}))},
					{P{1, 6}, addr(valueStringer{})},
					{P{1, 7}, addr(any(pointerStringer{}))},
					{P{1, 7}, addr(pointerStringer{})},
					{P{1, 8}, addr(any((*pointerStringer)(nil)))},
					{P{1, 8}, addr((*pointerStringer)(nil))},
					{P{1, 9}, addr(any(addr(pointerStringer{})))},
					{P{1, 9}, addr(addr(pointerStringer{}))},
					{P{1, 9}, addr(pointerStringer{})},
					{P{1, 10}, addr(any((**pointerStringer)(nil)))},
					{P{1, 10}, addr((**pointerStringer)(nil))},
					{P{1, 11}, addr(any(addr((*pointerStringer)(nil))))},
					{P{1, 11}, addr(addr((*pointerStringer)(nil)))},
					{P{1, 11}, addr((*pointerStringer)(nil))},
					{P{1, 12}, addr(any(addr(addr(pointerStringer{}))))},
					{P{1, 12}, addr(addr(addr(pointerStringer{})))},
					{P{1, 12}, addr(addr(pointerStringer{}))},
					{P{1, 12}, addr(pointerStringer{})},
					{P{1, 13}, addr(any("LAST"))},
					{P{1, 13}, addr("LAST")},
				}
				checkAny := makeValueChecker("any", wantAny)
				anyMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v any) error {
					return checkAny(enc, v)
				})

				var wantPointerAny []PV
				for _, v := range wantAny {
					if _, ok := v.V.(*any); ok {
						wantPointerAny = append(wantPointerAny, v)
					}
				}
				checkPointerAny := makeValueChecker("*any", wantPointerAny)
				pointerAnyMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v *any) error {
					return checkPointerAny(enc, v)
				})

				checkNamedAny := makeValueChecker("namedAny", wantAny)
				namedAnyMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v namedAny) error {
					return checkNamedAny(enc, v)
				})

				checkPointerNamedAny := makeValueChecker("*namedAny", nil)
				pointerNamedAnyMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v *namedAny) error {
					return checkPointerNamedAny(enc, v)
				})

				type stringer = fmt.Stringer
				var wantStringer []PV
				for _, v := range wantAny {
					if _, ok := v.V.(stringer); ok {
						wantStringer = append(wantStringer, v)
					}
				}
				checkStringer := makeValueChecker("stringer", wantStringer)
				stringerMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v stringer) error {
					return checkStringer(enc, v)
				})

				checkPointerStringer := makeValueChecker("*stringer", nil)
				pointerStringerMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v *stringer) error {
					return checkPointerStringer(enc, v)
				})

				wantValueStringer := []P{{1, 1}, {1, 3}, {1, 6}}
				checkValueValueStringer := makePositionChecker("valueStringer", wantValueStringer)
				valueValueStringerMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v valueStringer) error {
					return checkValueValueStringer(enc, v)
				})

				checkPointerValueStringer := makePositionChecker("*valueStringer", wantValueStringer)
				pointerValueStringerMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v *valueStringer) error {
					return checkPointerValueStringer(enc, v)
				})

				wantPointerStringer := []P{{1, 7}, {1, 9}, {1, 12}}
				checkValuePointerStringer := makePositionChecker("pointerStringer", wantPointerStringer)
				valuePointerStringerMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v pointerStringer) error {
					return checkValuePointerStringer(enc, v)
				})

				checkPointerPointerStringer := makePositionChecker("*pointerStringer", wantPointerStringer)
				pointerPointerStringerMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v *pointerStringer) error {
					return checkPointerPointerStringer(enc, v)
				})

				lastMarshaler := MarshalToFunc(func(enc *jsontext.Encoder, v string) error {
					return checkLast()
				})

				return JoinMarshalers(
					anyMarshaler,
					pointerAnyMarshaler,
					namedAnyMarshaler,
					pointerNamedAnyMarshaler, // never called
					stringerMarshaler,
					pointerStringerMarshaler, // never called
					valueValueStringerMarshaler,
					pointerValueStringerMarshaler,
					valuePointerStringerMarshaler,
					pointerPointerStringerMarshaler,
					lastMarshaler,
				)
			}()),
		},
		in: []any{
			nil,                           // nil
			valueStringer{},               // T
			(*valueStringer)(nil),         // *T
			addr(valueStringer{}),         // *T
			(**valueStringer)(nil),        // **T
			addr((*valueStringer)(nil)),   // **T
			addr(addr(valueStringer{})),   // **T
			pointerStringer{},             // T
			(*pointerStringer)(nil),       // *T
			addr(pointerStringer{}),       // *T
			(**pointerStringer)(nil),      // **T
			addr((*pointerStringer)(nil)), // **T
			addr(addr(pointerStringer{})), // **T
			"LAST",
		},
		want: `[null,{},null,{},null,null,{},{},null,{},null,null,{},"LAST"]`,
	}, {
		name: Name("Functions/Precedence/V1First"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalFunc(func(bool) ([]byte, error) {
					return []byte(`"called"`), nil
				}),
				MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
					panic("should not be called")
				}),
			)),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/Precedence/V2First"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
					return enc.WriteToken(jsontext.String("called"))
				}),
				MarshalFunc(func(bool) ([]byte, error) {
					panic("should not be called")
				}),
			)),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/Precedence/V2Skipped"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalToFunc(func(enc *jsontext.Encoder, v bool) error {
					return errors.ErrUnsupported
				}),
				MarshalFunc(func(bool) ([]byte, error) {
					return []byte(`"called"`), nil
				}),
			)),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/Precedence/NestedFirst"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				JoinMarshalers(
					MarshalFunc(func(bool) ([]byte, error) {
						return []byte(`"called"`), nil
					}),
				),
				MarshalFunc(func(bool) ([]byte, error) {
					panic("should not be called")
				}),
			)),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Functions/Precedence/NestedLast"),
		opts: []Options{
			WithMarshalers(JoinMarshalers(
				MarshalFunc(func(bool) ([]byte, error) {
					return []byte(`"called"`), nil
				}),
				JoinMarshalers(
					MarshalFunc(func(bool) ([]byte, error) {
						panic("should not be called")
					}),
				),
			)),
		},
		in:   true,
		want: `"called"`,
	}, {
		name: Name("Duration/Zero"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		in: struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{0, 0},
		want: `{"D1":"0s","D2":0}`,
	}, {
		name: Name("Duration/Positive"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		in: struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{
			123456789123456789,
			123456789123456789,
		},
		want: `{"D1":"34293h33m9.123456789s","D2":123456789123456789}`,
	}, {
		name: Name("Duration/Negative"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		in: struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{
			-123456789123456789,
			-123456789123456789,
		},
		want: `{"D1":"-34293h33m9.123456789s","D2":-123456789123456789}`,
	}, {
		name: Name("Duration/Nanos/String"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		in: struct {
			D1 time.Duration `json:",string,format:nano"`
			D2 time.Duration `json:",string,format:nano"`
			D3 time.Duration `json:",string,format:nano"`
		}{
			math.MinInt64,
			0,
			math.MaxInt64,
		},
		want: `{"D1":"-9223372036854775808","D2":"0","D3":"9223372036854775807"}`,
	}, {
		name: Name("Duration/Format/Invalid"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		in: struct {
			D time.Duration `json:",format:invalid"`
		}{},
		want:    `{"D"`,
		wantErr: EM(errInvalidFormatFlag).withPos(`{"D":`, "/D").withType(0, T[time.Duration]()),
	}, {
		/* TODO(https://go.dev/issue/71631): Re-enable this test case.
		name: Name("Duration/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   time.Duration(0),
		want: `"0s"`,
		}, { */
		name: Name("Duration/Format"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true)},
		in: structDurationFormat{
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
		},
		want: `{
	"D1": "12h34m56.078090012s",
	"D2": "12h34m56.078090012s",
	"D3": 45296.078090012,
	"D4": "45296.078090012",
	"D5": 45296078.090012,
	"D6": "45296078.090012",
	"D7": 45296078090.012,
	"D8": "45296078090.012",
	"D9": 45296078090012,
	"D10": "45296078090012",
	"D11": "PT12H34M56.078090012S"
}`,
	}, {
		/* TODO(https://go.dev/issue/71631): Re-enable this test case.
		name: Name("Duration/Format/Legacy"),
		opts: []Options{unsupported("FormatDurationAsNano")},
		in: structDurationFormat{
			D1: 12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			D2: 12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
		},
		want: `{"D1":45296078090012,"D2":"12h34m56.078090012s","D3":0,"D4":"0","D5":0,"D6":"0","D7":0,"D8":"0","D9":0,"D10":"0","D11":"PT0S"}`,
		}, { */
		/* TODO(https://go.dev/issue/71631): Re-enable this test case.
		name: Name("Duration/MapKey"),
		in:   map[time.Duration]string{time.Second: ""},
		want: `{"1s":""}`,
		}, { */
		name: Name("Duration/MapKey/Legacy"),
		opts: []Options{unsupported("FormatDurationAsNano")},
		in:   map[time.Duration]string{time.Second: ""},
		want: `{"1000000000":""}`,
	}, {
		name: Name("Time/Zero"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		in: struct {
			T1 time.Time
			T2 time.Time `json:",format:RFC822"`
			T3 time.Time `json:",format:'2006-01-02'"`
			T4 time.Time `json:",omitzero"`
			T5 time.Time `json:",omitempty"`
		}{
			time.Time{},
			time.Time{},
			time.Time{},
			// This is zero according to time.Time.IsZero,
			// but non-zero according to reflect.Value.IsZero.
			time.Date(1, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC", 0)),
			time.Time{},
		},
		want: `{"T1":"0001-01-01T00:00:00Z","T2":"01 Jan 01 00:00 UTC","T3":"0001-01-01","T5":"0001-01-01T00:00:00Z"}`,
	}, {
		name: Name("Time/Format"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), jsontext.Multiline(true)},
		in: structTimeFormat{
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
			time.Date(1234, 1, 2, 3, 4, 5, 6, time.UTC),
		},
		want: `{
	"T1": "1234-01-02T03:04:05.000000006Z",
	"T2": "Mon Jan  2 03:04:05 1234",
	"T3": "Mon Jan  2 03:04:05 UTC 1234",
	"T4": "Mon Jan 02 03:04:05 +0000 1234",
	"T5": "02 Jan 34 03:04 UTC",
	"T6": "02 Jan 34 03:04 +0000",
	"T7": "Monday, 02-Jan-34 03:04:05 UTC",
	"T8": "Mon, 02 Jan 1234 03:04:05 UTC",
	"T9": "Mon, 02 Jan 1234 03:04:05 +0000",
	"T10": "1234-01-02T03:04:05Z",
	"T11": "1234-01-02T03:04:05.000000006Z",
	"T12": "3:04AM",
	"T13": "Jan  2 03:04:05",
	"T14": "Jan  2 03:04:05.000",
	"T15": "Jan  2 03:04:05.000000",
	"T16": "Jan  2 03:04:05.000000006",
	"T17": "1234-01-02 03:04:05",
	"T18": "1234-01-02",
	"T19": "03:04:05",
	"T20": "1234-01-02",
	"T21": "\"weird\"1234",
	"T22": -23225777754.999999994,
	"T23": "-23225777754.999999994",
	"T24": -23225777754999.999994,
	"T25": "-23225777754999.999994",
	"T26": -23225777754999999.994,
	"T27": "-23225777754999999.994",
	"T28": -23225777754999999994,
	"T29": "-23225777754999999994"
}`,
	}, {
		name: Name("Time/Format/Invalid"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		in: struct {
			T time.Time `json:",format:UndefinedConstant"`
		}{},
		want:    `{"T"`,
		wantErr: EM(errors.New(`invalid format flag "UndefinedConstant"`)).withPos(`{"T":`, "/T").withType(0, timeTimeType),
	}, {
		name:    Name("Time/Format/String/Invalid"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		in:      structTimeFormatStringInvalid{},
		want:    `{"T"`,
		wantErr: EM(errInvalidStringTag).withPos(`{"T":`, "/T").withType(0, timeTimeType),
	}, {
		name: Name("Time/Format/YearOverflow"),
		in: struct {
			T1 time.Time
			T2 time.Time
		}{
			time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second),
			time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		want:    `{"T1":"9999-12-31T23:59:59Z","T2"`,
		wantErr: EM(errors.New(`year outside of range [0,9999]`)).withPos(`{"T1":"9999-12-31T23:59:59Z","T2":`, "/T2").withType(0, timeTimeType),
	}, {
		name: Name("Time/Format/YearUnderflow"),
		in: struct {
			T1 time.Time
			T2 time.Time
		}{
			time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second),
		},
		want:    `{"T1":"0000-01-01T00:00:00Z","T2"`,
		wantErr: EM(errors.New(`year outside of range [0,9999]`)).withPos(`{"T1":"0000-01-01T00:00:00Z","T2":`, "/T2").withType(0, timeTimeType),
	}, {
		name:    Name("Time/Format/YearUnderflow"),
		in:      struct{ T time.Time }{time.Date(-998, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)},
		want:    `{"T"`,
		wantErr: EM(errors.New(`year outside of range [0,9999]`)).withPos(`{"T":`, "/T").withType(0, timeTimeType),
	}, {
		name: Name("Time/Format/ZoneExact"),
		in:   struct{ T time.Time }{time.Date(2020, 1, 1, 0, 0, 0, 0, time.FixedZone("", 23*60*60+59*60))},
		want: `{"T":"2020-01-01T00:00:00+23:59"}`,
	}, {
		name:    Name("Time/Format/ZoneHourOverflow"),
		in:      struct{ T time.Time }{time.Date(2020, 1, 1, 0, 0, 0, 0, time.FixedZone("", 24*60*60))},
		want:    `{"T"`,
		wantErr: EM(errors.New(`timezone hour outside of range [0,23]`)).withPos(`{"T":`, "/T").withType(0, timeTimeType),
	}, {
		name:    Name("Time/Format/ZoneHourOverflow"),
		in:      struct{ T time.Time }{time.Date(2020, 1, 1, 0, 0, 0, 0, time.FixedZone("", 123*60*60))},
		want:    `{"T"`,
		wantErr: EM(errors.New(`timezone hour outside of range [0,23]`)).withPos(`{"T":`, "/T").withType(0, timeTimeType),
	}, {
		name: Name("Time/IgnoreInvalidFormat"),
		opts: []Options{invalidFormatOption},
		in:   time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		want: `"2000-01-01T00:00:00Z"`,
	}}

	for _, tt := range tests {
		t.Run(tt.name.Name, func(t *testing.T) {
			skipUnsupported(t, tt.opts)
			var got []byte
			var gotErr error
			if tt.useWriter {
				bb := new(struct{ bytes.Buffer }) // avoid optimizations with bytes.Buffer
				gotErr = MarshalWrite(bb, tt.in, tt.opts...)
				got = bb.Bytes()
			} else {
				got, gotErr = Marshal(tt.in, tt.opts...)
			}
			if tt.canonicalize {
				(*jsontext.Value)(&got).Canonicalize()
			}
			if tt.wantErr == nil && string(got) != tt.want {
				t.Errorf("%s: Marshal output mismatch:\ngot  %s\nwant %s", tt.name.Where, got, tt.want)
			}
			if !equalError(gotErr, tt.wantErr) {
				t.Errorf("%s: Marshal error mismatch:\ngot  %v\nwant %v", tt.name.Where, gotErr, tt.wantErr)
			}
		})
	}
}

func TestMarshalInvalidNamespace(t *testing.T) {
	tests := []struct {
		name CaseName
		val  any
	}{
		{Name("Map"), map[string]string{"X": "\xde\xad\xbe\xef"}},
		{Name("Struct"), struct{ X string }{"\xde\xad\xbe\xef"}},
		{Name("MapBadKey"), map[string]int{"\xde\xad\xbe\xef": 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name.Name, func(t *testing.T) {
			enc := jsontext.NewEncoder(new(bytes.Buffer))
			if err := MarshalEncode(enc, tt.val); err == nil {
				t.Fatalf("%s: MarshalEncode error is nil, want non-nil", tt.name.Where)
			}
			for _, tok := range []jsontext.Token{
				jsontext.Null, jsontext.String(""), jsontext.Int(0), jsontext.BeginObject, jsontext.EndObject, jsontext.BeginArray, jsontext.EndArray,
			} {
				if err := enc.WriteToken(tok); err == nil {
					t.Fatalf("%s: WriteToken error is nil, want non-nil", tt.name.Where)
				}
			}
			for _, val := range []string{`null`, `""`, `0`, `{}`, `[]`} {
				if err := enc.WriteValue([]byte(val)); err == nil {
					t.Fatalf("%s: WriteToken error is nil, want non-nil", tt.name.Where)
				}
			}
		})
	}
}

// TestMarshalEncodeInvalidNamespaceAtName verifies that MarshalEncode at an
// object-name position with an invalidated namespace surfaces the underlying
// namespace error rather than wrapping nil.

// TestMarshalEncodeInvalidNamespaceAtName verifies that MarshalEncode at an
// object-name position with an invalidated namespace surfaces the underlying
// namespace error rather than wrapping nil.
func TestMarshalEncodeInvalidNamespaceAtName(t *testing.T) {
	enc := jsontext.NewEncoder(new(bytes.Buffer))

	// The bad-UTF-8 key fails the map marshal at the key-write step,
	// leaving the encoder at an object-name position with the namespace
	// invalidated.
	if err := MarshalEncode(enc, map[string]int{"\xde\xad\xbe\xef": 0}); err == nil {
		t.Fatal("MarshalEncode error is nil, want non-nil")
	}

	var serr *jsontext.SyntacticError
	if err := MarshalEncode(enc, 0); !errors.As(err, &serr) || serr.Err == nil {
		t.Fatalf("MarshalEncode error = %v, want *jsontext.SyntacticError wrapping a non-nil error", err)
	}
}

func TestMarshalEncodeOptions(t *testing.T) {
	var calledFuncs int
	var calledOptions Options
	out := new(bytes.Buffer)
	enc := jsontext.NewEncoder(
		out,
		jsontext.AllowInvalidUTF8(true), // encoder-specific option
		WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, _ any) error {
			opts := enc.Options()
			if v, _ := GetOption(opts, jsontext.AllowInvalidUTF8); !v {
				t.Errorf("nested Options.AllowInvalidUTF8 = false, want true")
			}
			calledFuncs++
			calledOptions = opts
			return errors.ErrUnsupported
		})), // marshal-specific option; only relevant for MarshalEncode
	)

	if err := MarshalEncode(enc, "\xde\xad\xbe\xef"); err != nil {
		t.Fatalf("MarshalEncode: %v", err)
	}
	if calledFuncs != 1 {
		t.Fatalf("calledFuncs = %d, want 1", calledFuncs)
	}
	if err := MarshalEncode(enc, "\xde\xad\xbe\xef", calledOptions); err != nil {
		t.Fatalf("MarshalEncode: %v", err)
	}
	if calledFuncs != 2 {
		t.Fatalf("calledFuncs = %d, want 2", calledFuncs)
	}
	if err := MarshalEncode(enc, "\xde\xad\xbe\xef",
		jsontext.AllowInvalidUTF8(false), // expect to see invalid UTF-8 error
		WithMarshalers(nil),              // avoid calling previously registered marshalers
	); !errors.Is(err, ierrors.ErrInvalidUTF8) {
		t.Fatalf("MarshalEncode = %v, want %v", err, ierrors.ErrInvalidUTF8)
	}
	if err := MarshalEncode(enc, "\xde\xad\xbe\xef",
		WithMarshalers(nil), // should override
	); err != nil {
		t.Fatalf("MarshalEncode: %v", err)
	}
	if calledFuncs != 2 {
		t.Fatalf("calledFuncs = %d, want 2", calledFuncs)
	}
	if err := MarshalEncode(enc, "\xde\xad\xbe\xef"); err != nil {
		t.Fatalf("MarshalEncode: %v", err)
	}
	if calledFuncs != 3 {
		t.Fatalf("calledFuncs = %d, want 3", calledFuncs)
	}
	if err := MarshalEncode(enc, "\xde\xad\xbe\xef", JoinOptions(
		WithMarshalers(MarshalToFunc(func(enc *jsontext.Encoder, _ any) error {
			opts := enc.Options()
			if v, _ := GetOption(opts, jsontext.AllowInvalidUTF8); !v {
				t.Errorf("nested Options.AllowInvalidUTF8 = false, want true")
			}
			calledFuncs = math.MaxInt
			return errors.ErrUnsupported
		})), // should override
	)); err != nil {
		t.Fatalf("MarshalEncode: %v", err)
	}
	if calledFuncs != math.MaxInt {
		t.Fatalf("calledFuncs = %d, want %d", calledFuncs, math.MaxInt)
	}
	if out.String() != strings.Repeat("\"\xde\xad\ufffd\ufffd\"\n", 5) {
		t.Fatalf("output mismatch:\n\tgot:  %s\n\twant: %s", out.String(), strings.Repeat("\"\xde\xad\xbe\xef\"\n", 5))
	}

	// Reset with the encoder options as part of the arguments should not
	// observe mutations to the options until after Reset is done.
	opts := enc.Options()                                  // AllowInvalidUTF8 is currently true
	enc.Reset(out, jsontext.AllowInvalidUTF8(false), opts) // earlier AllowInvalidUTF8(false) should be overridden by latter AllowInvalidUTF8(true) in opts
	if v, _ := GetOption(enc.Options(), jsontext.AllowInvalidUTF8); v == false {
		t.Errorf("Options.AllowInvalidUTF8 = false, want true")
	}

	// Verify that AllowInvalidUTF8 and AllowDuplicateNames cannot be changed
	// when positioned at a JSON object name, but can be changed for the value.
	{
		var buf2 bytes.Buffer
		enc2 := jsontext.NewEncoder(&buf2)
		if err := enc2.WriteToken(jsontext.BeginObject); err != nil {
			t.Fatalf("WriteToken(BeginObject) = %v, want nil", err)
		}

		// Verify that you cannot change AllowDuplicateNames or AllowInvalidUTF8 settings for the JSON member name.
		if err := MarshalEncode(enc2, "name", jsontext.AllowDuplicateNames(true)); !isSemanticErr(err, errChangingDuplicateNames) {
			t.Errorf("MarshalEncode(name) = %v, want %v", err, errChangingDuplicateNames)
		}
		if err := MarshalEncode(enc2, "name", jsontext.AllowInvalidUTF8(true)); !isSemanticErr(err, errChangingInvalidUTF8) {
			t.Errorf("MarshalEncode(name) = %v, want %v", err, errChangingInvalidUTF8)
		}
		// Setting the same option value does not report an error.
		if err := MarshalEncode(enc2, "name", jsontext.AllowDuplicateNames(false)); err != nil {
			t.Errorf("MarshalEncode(name) = %v, want nil", err)
		}
		// At value position, changing AllowDuplicateNames is allowed.
		if err := MarshalEncode(enc2, jsontext.Value(`{"dupe":"value","dupe":"value"}`), jsontext.AllowDuplicateNames(true)); err != nil {
			t.Errorf("MarshalEncode(value) = %v, want nil", err)
		}

		// Verify that you can change AllowInvalidUTF8 for the JSON member value.
		if err := MarshalEncode(enc2, "name2"); err != nil {
			t.Errorf("MarshalEncode(name) = %v, want nil", err)
		}
		if err := MarshalEncode(enc2, "value\xde\xad\xbe\xef"); !errors.Is(err, ierrors.ErrInvalidUTF8) {
			t.Errorf("MarshalEncode(value) = %v, want %v", err, ierrors.ErrInvalidUTF8)
		}
		if err := MarshalEncode(enc2, "value\xde\xad\xbe\xef", jsontext.AllowInvalidUTF8(true)); err != nil {
			t.Errorf("MarshalEncode(value) = %v, want nil", err)
		}

		if err := enc2.WriteToken(jsontext.EndObject); err != nil {
			t.Errorf("WriteToken(EndObject) = %v, want nil", err)
		}
	}

	// Verify that whitespace options cannot change within a MarshalEncode call,
	// but setting them to identical values is allowed.
	{
		var buf3 bytes.Buffer
		enc3 := jsontext.NewEncoder(&buf3, jsontext.WithIndent("\t"))
		if err := MarshalEncode(enc3, "value", jsontext.Multiline(false)); !isSemanticErr(err, errChangingWhitespace) {
			t.Errorf("MarshalEncode = %v, want %v", err, errChangingWhitespace)
		}
		if err := MarshalEncode(enc3, "value", jsontext.WithIndent("  ")); !isSemanticErr(err, errChangingWhitespace) {
			t.Errorf("MarshalEncode = %v, want %v", err, errChangingWhitespace)
		}
		// Setting identical whitespace options does not report an error.
		if err := MarshalEncode(enc3, "value", enc3.Options()); err != nil {
			t.Errorf("MarshalEncode = %v, want nil", err)
		}
	}
}

// BenchmarkMarshalEncodeOptions is a minimal encode operation to measure
// the overhead of options setup before the marshal operation.
