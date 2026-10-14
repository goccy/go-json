package json_test

import (
	"bufio"
	"bytes"
	stdjson "encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	json "github.com/goccy/go-json"
)

// go-json decodes as encoding/json of Go 1.27, which is made of encoding/json/v2, on every version of Go: the errors
// of the decoding are the ones of Go 1.27 but for the syntax errors, which are the ones of the running Go.
// TestDecodeAsGo127 checks the errors of the decodings of asGo127Cases, by Unmarshal and by a Decoder in every way it
// reads, against a golden file of the results of encoding/json of Go 1.27, which it checks against encoding/json where
// it is the one of Go 1.27, so that it can't go stale.

// updateDecodeAsGo127 writes the golden file of TestDecodeAsGo127 from encoding/json of Go 1.27.
var updateDecodeAsGo127 = flag.Bool("update", false, "write the golden file of TestDecodeAsGo127 from encoding/json of Go 1.27")

const asGo127Golden = "testdata/decode_as_go127.golden"

type asGo127Leaf struct{ N int }
type asGo127L2 struct{ B asGo127Leaf }
type asGo127L1 struct{ A asGo127L2 }
type asGo127Tagged struct {
	X struct {
		Y int `json:"yy"`
	} `json:"xx"`
}

// AsGo127Emb is an exported struct, which is embedded.
type AsGo127Emb struct{ Y int }
type asGo127EmbOuter struct {
	AsGo127Emb
	X int
}
type asGo127EmbPtrOuter struct {
	*AsGo127Emb
	X int
}
type asGo127Unexported struct{ Z int }
type asGo127EmbUnexp struct {
	asGo127Unexported
	X int
}
type asGo127EmbUnexpPtr struct {
	*asGo127Unexported
	X int
}
type asGo127Deep2 struct{ AsGo127Emb }
type asGo127EmbDeep struct{ asGo127Deep2 }
type asGo127EmbTagged struct {
	AsGo127Emb `json:"e"`
}

type asGo127MapSS struct{ M map[string]asGo127Leaf }

// asGo127Key is a key of a map whose UnmarshalText fails for "bad".
type asGo127Key string

func (k *asGo127Key) UnmarshalText(b []byte) error {
	if string(b) == "bad" {
		return errors.New("bad key")
	}
	*k = asGo127Key(b)
	return nil
}

type asGo127SliceS struct{ S []asGo127Leaf }
type asGo127ArrS struct{ A [2]int }

type asGo127IfaceHolder struct{ Env any }
type asGo127Inner struct{ Data any }

type asGo127Stringer interface{ String() string }
type asGo127SImpl struct{ N int }

func (s *asGo127SImpl) String() string { return "" }

type asGo127SImpl2 struct{ Data any }

func (s *asGo127SImpl2) String() string { return "" }

type asGo127MHolder struct{ V asGo127Stringer }

// asGo127Failing is a value whose UnmarshalJSON fails.
type asGo127Failing struct{}

func (u *asGo127Failing) UnmarshalJSON([]byte) error { return errors.New("u failed") }

type asGo127FailingHolder struct {
	A int
	U asGo127Failing
	B int
}

// asGo127StdTypeError is a value whose UnmarshalJSON returns a type error of encoding/json.
type asGo127StdTypeError struct{}

func (u *asGo127StdTypeError) UnmarshalJSON([]byte) error {
	return &stdjson.UnmarshalTypeError{Value: "x", Type: reflect.TypeOf(0), Offset: 1}
}

type asGo127StdTypeErrorHolder struct{ U asGo127StdTypeError }

// asGo127Text is a value whose UnmarshalText fails for "bad".
type asGo127Text struct{ s string }

func (t *asGo127Text) UnmarshalText(b []byte) error {
	if string(b) == "bad" {
		return errors.New("tu failed")
	}
	t.s = string(b)
	return nil
}
func (t *asGo127Text) String() string { return t.s }

type asGo127TextHolder struct{ T asGo127Text }

type asGo127StrOpt struct {
	N int     `json:",string"`
	S string  `json:",string"`
	B bool    `json:",string"`
	F float64 `json:",string"`
	P *int    `json:",string"`
	U uint8   `json:",string"`
}

type asGo127RawHolder struct {
	R stdjson.RawMessage
	N int
}
type asGo127NumHolder struct {
	I8  int8
	U   uint
	F32 float32
	I   int
}
type asGo127NumberHolder struct{ N stdjson.Number }
type asGo127TimeHolder struct{ T time.Time }
type asGo127BytesHolder struct{ B []byte }
type asGo127PtrHolder struct {
	P  *asGo127Leaf
	PP **int
}
type asGo127DeepMix struct {
	A []map[string]struct{ B [2]int }
}
type asGo127TwoInts struct{ A, B int }
type asGo127StrHolder struct{ S string }
type asGo127AnyHolder struct{ V any }
type asGo127MapHolder struct{ M map[string]int }
type asGo127MapIntHolder struct{ M map[int]asGo127Leaf }
type asGo127MapKeyHolder struct{ M map[asGo127Key]int }
type asGo127SliceAnyHolder struct{ L []any }
type asGo127ArrLeaf struct{ A [2]asGo127Leaf }
type asGo127NamedSlice []asGo127Leaf
type asGo127NamedMap map[string]asGo127Leaf

// asGo127Case is an input, and the value which it is decoded into.
type asGo127Case struct {
	name   string
	input  string
	target func() any
}

func asGo127New[T any]() func() any { return func() any { return new(T) } }

var asGo127Cases = []asGo127Case{
	// top level
	{"top/int/string", `"x"`, asGo127New[int]()},
	{"top/int/object", `{"a":1}`, asGo127New[int]()},
	{"top/slice/object", `{}`, asGo127New[[]int]()},
	{"top/struct/array", `[1]`, asGo127New[asGo127Leaf]()},
	{"top/slice/elem", `[1,"x"]`, asGo127New[[]int]()},
	{"top/array/elem", `[1,"x"]`, asGo127New[[2]int]()},
	{"top/array/overflow-elem", `[1,2,"x"]`, asGo127New[[2]int]()},
	{"top/namedslice/elem", `[{"N":1},{"N":"x"}]`, asGo127New[asGo127NamedSlice]()},
	{"top/namedmap/val", `{"k":{"N":"x"}}`, asGo127New[asGo127NamedMap]()},
	{"top/map/val", `{"a":"x"}`, asGo127New[map[string]int]()},
	{"top/map/key-slash", `{"a/b":"x"}`, asGo127New[map[string]int]()},
	{"top/map/key-tilde", `{"a~b":"x"}`, asGo127New[map[string]int]()},
	{"top/map/key-dot", `{"a.b":"x"}`, asGo127New[map[string]int]()},
	{"top/map/key-escaped", `{"a\u0062":"x"}`, asGo127New[map[string]int]()},
	{"top/mapint/key", `{"x":1}`, asGo127New[map[int]int]()},
	{"top/mapint/val", `{"1":"x"}`, asGo127New[map[int]int]()},
	{"top/mapint8/key-range", `{"300":1}`, asGo127New[map[int8]int]()},
	{"top/mapuint/key-neg", `{"-1":1}`, asGo127New[map[uint]int]()},
	{"top/mapkey/key-method", `{"bad":1,"ok":2}`, asGo127New[map[asGo127Key]int]()},
	{"top/mapkey/val", `{"ok":"x"}`, asGo127New[map[asGo127Key]int]()},
	{"top/mapbool/key", `{"true":1}`, asGo127New[map[bool]int]()},
	{"top/mapfloat/key", `{"1.5":1}`, asGo127New[map[float64]int]()},
	{"top/any/range", `1e400`, asGo127New[any]()},
	{"top/sliceany/range", `[1,1e400]`, asGo127New[[]any]()},
	{"top/mapany/range", `{"k":[1e400]}`, asGo127New[map[string]any]()},
	{"top/float32/range", `1e40`, asGo127New[float32]()},
	{"top/int8/range", `300`, asGo127New[int8]()},
	{"top/int/frac", `1.5`, asGo127New[int]()},
	{"top/uint/neg", `-1`, asGo127New[uint]()},
	{"top/string/number", `1`, asGo127New[string]()},
	{"top/bool/string", `"true"`, asGo127New[bool]()},
	{"top/ptrint/string", `"x"`, asGo127New[*int]()},
	// nested structs
	{"struct/nested", `{"A":{"B":{"N":"x"}}}`, asGo127New[asGo127L1]()},
	{"struct/nested-ci", `{"a":{"b":{"n":"x"}}}`, asGo127New[asGo127L1]()},
	{"struct/nested-escaped-key", `{"\u0041":{"B":{"N":"x"}}}`, asGo127New[asGo127L1]()},
	{"struct/nested-object-kind", `{"A":{"B":{"N":{"z":1}}}}`, asGo127New[asGo127L1]()},
	{"struct/nested-array-kind", `{"A":{"B":[1]}}`, asGo127New[asGo127L1]()},
	{"struct/tagged", `{"xx":{"yy":"x"}}`, asGo127New[asGo127Tagged]()},
	{"struct/tagged-ci", `{"XX":{"YY":"x"}}`, asGo127New[asGo127Tagged]()},
	{"struct/kind/bool", `{"N":true}`, asGo127New[asGo127Leaf]()},
	{"struct/kind/object", `{"N":{}}`, asGo127New[asGo127Leaf]()},
	{"struct/kind/array", `{"N":[1,2]}`, asGo127New[asGo127Leaf]()},
	{"struct/kind/obj-into-string", `{"S":{"a":1}}`, asGo127New[asGo127StrHolder]()},
	{"struct/first-wins", `{"A":"x","B":"y"}`, asGo127New[asGo127TwoInts]()},
	{"struct/then-syntax", `{"A":"x","B":}`, asGo127New[asGo127TwoInts]()},
	{"struct/unknown-then", `{"Z":{"q":[1,2]},"A":"x"}`, asGo127New[asGo127TwoInts]()},
	// embedded
	{"embedded/exported", `{"Y":"x"}`, asGo127New[asGo127EmbOuter]()},
	{"embedded/exported-ptr", `{"Y":"x"}`, asGo127New[asGo127EmbPtrOuter]()},
	{"embedded/unexported", `{"Z":"x"}`, asGo127New[asGo127EmbUnexp]()},
	{"embedded/unexported-ptr-settable", `{"Z":1}`, asGo127New[asGo127EmbUnexpPtr]()},
	{"embedded/unexported-ptr-bad", `{"Z":"x"}`, asGo127New[asGo127EmbUnexpPtr]()},
	{"embedded/deep", `{"Y":"x"}`, asGo127New[asGo127EmbDeep]()},
	{"embedded/tagged", `{"e":{"Y":"x"}}`, asGo127New[asGo127EmbTagged]()},
	// maps, slices and arrays in structs
	{"map/struct-val", `{"M":{"k":{"N":"x"}}}`, asGo127New[asGo127MapSS]()},
	{"map/val", `{"M":{"k":"x"}}`, asGo127New[asGo127MapHolder]()},
	{"map/key-slash", `{"M":{"a/b~c":"x"}}`, asGo127New[asGo127MapHolder]()},
	{"mapint/key", `{"M":{"x":{"N":1}}}`, asGo127New[asGo127MapIntHolder]()},
	{"mapint/val", `{"M":{"1":{"N":"x"}}}`, asGo127New[asGo127MapIntHolder]()},
	{"mapkey/key-method", `{"M":{"bad":1,"ok":2}}`, asGo127New[asGo127MapKeyHolder]()},
	{"mapkey/val", `{"M":{"ok":"x"}}`, asGo127New[asGo127MapKeyHolder]()},
	{"slice/struct", `{"S":[{"N":1},{"N":"x"}]}`, asGo127New[asGo127SliceS]()},
	{"array/elem", `{"A":[1,"x"]}`, asGo127New[asGo127ArrS]()},
	{"array/struct", `{"A":[{"N":1},{"N":"x"}]}`, asGo127New[asGo127ArrLeaf]()},
	{"deep/mix", `{"A":[{"k":{"B":[1,"x"]}}]}`, asGo127New[asGo127DeepMix]()},
	{"ptr/struct", `{"P":{"N":"x"}}`, asGo127New[asGo127PtrHolder]()},
	{"ptr/ptrptr", `{"PP":"x"}`, asGo127New[asGo127PtrHolder]()},
	// interfaces
	{"iface/ptr", `{"Env":{"N":"x"}}`, func() any { return &asGo127IfaceHolder{Env: &asGo127Leaf{}} }},
	{"iface/ptr-ci", `{"Env":{"n":"x"}}`, func() any { return &asGo127IfaceHolder{Env: &asGo127Leaf{}} }},
	{"iface/ptr-nested", `{"Env":{"Data":{"N":"x"}}}`, func() any {
		return &asGo127IfaceHolder{Env: &asGo127Inner{Data: &asGo127Leaf{}}}
	}},
	{"iface/ptr-nested-any", `{"Env":{"Data":{"n":[1,1e400]}}}`, func() any {
		return &asGo127IfaceHolder{Env: &asGo127Inner{}}
	}},
	{"iface/ptr-to-slice", `{"Env":[1,"x"]}`, func() any { return &asGo127IfaceHolder{Env: &[]int{}} }},
	{"iface/ptr-to-map", `{"Env":{"k":"x"}}`, func() any { return &asGo127IfaceHolder{Env: &map[string]int{}} }},
	{"iface/nonptr", `{"Env":{"N":"x"}}`, func() any { return &asGo127IfaceHolder{Env: asGo127Leaf{}} }},
	{"iface/any-map-range", `{"Env":{"a":[1,1e400]}}`, asGo127New[asGo127IfaceHolder]()},
	{"iface/any-range", `{"V":1e400}`, asGo127New[asGo127AnyHolder]()},
	{"iface/slice-of-ptr", `{"L":[{"N":1},{"N":"x"}]}`, func() any {
		return &asGo127SliceAnyHolder{L: []any{&asGo127Leaf{}, &asGo127Leaf{}}}
	}},
	{"iface/top-slice-of-ptr", `[{"N":"x"}]`, func() any { return &[]any{&asGo127Leaf{}} }},
	{"iface/map-of-ptr", `{"k":{"N":"x"}}`, func() any { return &map[string]any{"k": &asGo127Leaf{}} }},
	{"miface/ptr", `{"V":{"N":"x"}}`, func() any { return &asGo127MHolder{V: &asGo127SImpl{}} }},
	{"miface/nil", `{"V":{"N":1}}`, asGo127New[asGo127MHolder]()},
	{"miface/nil-scalar", `{"V":1}`, asGo127New[asGo127MHolder]()},
	{"miface/ptr-nested", `{"V":{"Data":{"N":"x"}}}`, func() any {
		return &asGo127MHolder{V: &asGo127SImpl2{Data: &asGo127Leaf{}}}
	}},
	{"miface/ptr-nested-any", `{"V":{"Data":{"n":[1e400]}}}`, func() any {
		return &asGo127MHolder{V: &asGo127SImpl2{}}
	}},
	{"miface/ptr-text", `{"V":1}`, func() any { return &asGo127MHolder{V: &asGo127Text{}} }},
	{"iface/ptr-text", `{"Env":1}`, func() any { return &asGo127IfaceHolder{Env: &asGo127Text{}} }},
	// methods
	{"method/json-error", `{"A":1,"U":1,"B":2}`, asGo127New[asGo127FailingHolder]()},
	{"method/json-after-type", `{"A":"x","U":1,"B":2}`, asGo127New[asGo127FailingHolder]()},
	{"method/json-before-type", `{"A":1,"U":1,"B":"x"}`, asGo127New[asGo127FailingHolder]()},
	{"method/json-std-type-error", `{"U":1}`, asGo127New[asGo127StdTypeErrorHolder]()},
	{"method/text-kind", `{"T":1}`, asGo127New[asGo127TextHolder]()},
	{"method/text-kind-object", `{"T":{"a":1}}`, asGo127New[asGo127TextHolder]()},
	{"method/text-error", `{"T":"bad"}`, asGo127New[asGo127TextHolder]()},
	{"method/text-top-kind", `1`, asGo127New[asGo127Text]()},
	{"time/kind", `{"T":1}`, asGo127New[asGo127TimeHolder]()},
	{"time/bad", `{"T":"bad"}`, asGo127New[asGo127TimeHolder]()},
	// the string option
	{"stropt/int-letters", `{"N":"x"}`, asGo127New[asGo127StrOpt]()},
	{"stropt/int-unquoted", `{"N":1}`, asGo127New[asGo127StrOpt]()},
	{"stropt/int-frac", `{"N":"1.5"}`, asGo127New[asGo127StrOpt]()},
	{"stropt/int-range", `{"N":"300000000000000000000"}`, asGo127New[asGo127StrOpt]()},
	{"stropt/uint8-range", `{"U":"300"}`, asGo127New[asGo127StrOpt]()},
	{"stropt/int-empty", `{"N":""}`, asGo127New[asGo127StrOpt]()},
	{"stropt/string-unquoted-inner", `{"S":"abc"}`, asGo127New[asGo127StrOpt]()},
	{"stropt/string-number", `{"S":1}`, asGo127New[asGo127StrOpt]()},
	{"stropt/bool-word", `{"B":"maybe"}`, asGo127New[asGo127StrOpt]()},
	{"stropt/float-letters", `{"F":"x"}`, asGo127New[asGo127StrOpt]()},
	{"stropt/ptr-letters", `{"P":"x"}`, asGo127New[asGo127StrOpt]()},
	// numbers
	{"num/int8-range", `{"I8":300}`, asGo127New[asGo127NumHolder]()},
	{"num/uint-neg", `{"U":-1}`, asGo127New[asGo127NumHolder]()},
	{"num/float32-range", `{"F32":1e40}`, asGo127New[asGo127NumHolder]()},
	{"num/int-frac", `{"I":1.5}`, asGo127New[asGo127NumHolder]()},
	{"num/int-exp", `{"I":1e3}`, asGo127New[asGo127NumHolder]()},
	{"number/bool", `{"N":true}`, asGo127New[asGo127NumberHolder]()},
	{"number/string-letters", `{"N":"abc"}`, asGo127New[asGo127NumberHolder]()},
	{"number/object", `{"N":{}}`, asGo127New[asGo127NumberHolder]()},
	// bytes
	{"bytes/bad-base64", `{"B":"!!"}`, asGo127New[asGo127BytesHolder]()},
	{"bytes/number", `{"B":1}`, asGo127New[asGo127BytesHolder]()},
	{"bytes/array-elem", `{"B":[1,"x"]}`, asGo127New[asGo127BytesHolder]()},
	{"bytes/array-range", `{"B":[300]}`, asGo127New[asGo127BytesHolder]()},
	// raw
	{"raw/then-error", `{"R":{"a":[1,2]},"N":"x"}`, asGo127New[asGo127RawHolder]()},
}

// asGo127TokenCase is an input which is read by turns by Token and Decode: steps has a 'T' for a Token, and a 'D'
// for a Decode into the next target.
type asGo127TokenCase struct {
	name    string
	input   string
	steps   string
	targets []func() any
}

var asGo127TokenCases = []asGo127TokenCase{
	{"token/float-range", `[1e400]`, "TT", nil},
	{"token/decode-elem", `[{"N":1}, {"N":"x"}]`, "TDD", []func() any{asGo127New[asGo127Leaf](), asGo127New[asGo127Leaf]()}},
	{"token/decode-member", `{"a": {"N":"x"}}`, "TTD", []func() any{asGo127New[asGo127Leaf]()}},
	{"token/decode-elem-nested", `[{"A":{"B":{"N":"x"}}}]`, "TD", []func() any{asGo127New[asGo127L1]()}},
}

// asGo127Record is what TestDecodeAsGo127 compares of a decoding: the kind of its error, and, unless it is a syntax
// error, whose message and offset are the ones of the running Go, the fields and the message of the error, and the
// value decoded.
type asGo127Record struct {
	Case    string `json:"case"`
	Layout  string `json:"layout"`
	Decoded string `json:"decoded,omitempty"`
	Kind    string `json:"kind"`
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	Offset  int64  `json:"offset,omitempty"`
	Struct  string `json:"struct,omitempty"`
	Field   string `json:"field,omitempty"`
	Err     string `json:"err,omitempty"`
	ErrKind string `json:"errKind,omitempty"`
	Message string `json:"message,omitempty"`
}

// asGo127ErrorKind returns the kind of err: a syntax error, a type error, or another error.
func asGo127ErrorKind(err error) string {
	switch err.(type) {
	case *json.SyntaxError, *stdjson.SyntaxError:
		return "SyntaxError"
	case *json.UnmarshalTypeError, *stdjson.UnmarshalTypeError:
		return "UnmarshalTypeError"
	}
	return "other"
}

// asGo127Describe returns the record of a decoding into what v points to, if v is not nil, which returned err, an error
// of go-json or of encoding/json, as the two packages have types of their own.
func asGo127Describe(err error, v any) asGo127Record {
	r := asGo127Record{Kind: "nil"}
	if err != nil {
		r.Kind = asGo127ErrorKind(err)
	}
	if r.Kind == "SyntaxError" {
		// encoding/json checks the whole input before it decodes it, and go-json decodes the values before the error
		return r
	}
	if v != nil {
		r.Decoded = asGo127Value(reflect.ValueOf(v).Elem())
	}
	if err == nil {
		return r
	}
	var value, field, structName string
	var typ reflect.Type
	var offset int64
	var cause error
	switch e := err.(type) {
	case *json.UnmarshalTypeError:
		value, typ, offset, structName, field, cause = e.Value, e.Type, e.Offset, e.Struct, e.Field, e.Err
	case *stdjson.UnmarshalTypeError:
		value, typ, offset, structName, field, cause = e.Value, e.Type, e.Offset, e.Struct, e.Field, errors.Unwrap(e)
	}
	r.Value, r.Offset, r.Struct, r.Field, r.Message = value, offset, structName, field, err.Error()
	if typ != nil {
		r.Type = typ.String()
	}
	if cause != nil {
		r.Err, r.ErrKind = cause.Error(), asGo127ErrorKind(cause)
	}
	return r
}

// asGo127Value describes v with what its pointers and its interfaces lead to, and the entries of its maps in the
// order of their descriptions, so that the values decoded by both packages are compared by their descriptions.
func asGo127Value(v reflect.Value) string {
	var elems []string
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return "nil"
		}
		if v.Kind() == reflect.Interface {
			return v.Elem().Type().String() + "(" + asGo127Value(v.Elem()) + ")"
		}
		return "&" + asGo127Value(v.Elem())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			elems = append(elems, v.Type().Field(i).Name+":"+asGo127Value(v.Field(i)))
		}
		return "{" + strings.Join(elems, " ") + "}"
	case reflect.Map:
		if v.IsNil() {
			return "nil"
		}
		for iter := v.MapRange(); iter.Next(); {
			elems = append(elems, asGo127Value(iter.Key())+":"+asGo127Value(iter.Value()))
		}
		slices.Sort(elems)
		return "map[" + strings.Join(elems, " ") + "]"
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			return "nil"
		}
		for i := 0; i < v.Len(); i++ {
			elems = append(elems, asGo127Value(v.Index(i)))
		}
		return "[" + strings.Join(elems, " ") + "]"
	case reflect.String:
		return strconv.Quote(v.String())
	case reflect.Func:
		return fmt.Sprint(!v.IsNil())
	}
	return fmt.Sprint(v)
}

// asGo127Padded returns s with 300 spaces after every ',', ':', '{' and '[' out of its strings, so that a Decoder
// reads its objects and its arrays by several reads.
func asGo127Padded(s string) string {
	var b strings.Builder
	spaces := strings.Repeat(" ", 300)
	in, escaped := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		b.WriteByte(c)
		switch {
		case in && escaped:
			escaped = false
		case in && c == '\\':
			escaped = true
		case in && c == '"':
			in = false
		case in:
		case c == '"':
			in = true
		case c == ',', c == ':', c == '{', c == '[':
			b.WriteString(spaces)
		}
	}
	return b.String()
}

// asGo127Reader returns at most n bytes by a read.
type asGo127Reader struct {
	b []byte
	n int
}

func (r *asGo127Reader) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	k := min(r.n, len(p), len(r.b))
	copy(p, r.b[:k])
	r.b = r.b[k:]
	return k, nil
}

// asGo127Decoder is the Decoder of go-json or the one of encoding/json.
type asGo127Decoder interface {
	Decode(v any) error
	Token() (stdjson.Token, error)
}

type asGo127GoDecoder struct{ *json.Decoder }

func (d asGo127GoDecoder) Token() (stdjson.Token, error) { return d.Decoder.Token() }

// asGo127Reads are the ways a Decoder reads its input: from a bytes.Reader, or by reads of a few bytes.
var asGo127Reads = []struct {
	name   string
	reader func([]byte) io.Reader
}{
	{"bytes.Reader", func(b []byte) io.Reader { return bytes.NewReader(b) }},
	{"1-byte reads", func(b []byte) io.Reader { return &asGo127Reader{b: b, n: 1} }},
	{"64-byte reads", func(b []byte) io.Reader { return &asGo127Reader{b: b, n: 64} }},
	{"512-byte reads", func(b []byte) io.Reader { return &asGo127Reader{b: b, n: 512} }},
}

// asGo127Decoders returns the decoders of the input by go-json, or by encoding/json if std is set: by Unmarshal, and
// by a Decoder in every way it reads, after a value before it too, which must all return the error of Unmarshal.
func asGo127Decoders(std bool) []struct {
	name   string
	decode func(data []byte, v any) error
} {
	unmarshal := json.Unmarshal
	newDecoder := func(r io.Reader) asGo127Decoder { return asGo127GoDecoder{json.NewDecoder(r)} }
	if std {
		unmarshal = stdjson.Unmarshal
		newDecoder = func(r io.Reader) asGo127Decoder { return stdjson.NewDecoder(r) }
	}
	decoders := []struct {
		name   string
		decode func(data []byte, v any) error
	}{{"Unmarshal", unmarshal}}
	for _, read := range asGo127Reads {
		decoders = append(decoders, struct {
			name   string
			decode func(data []byte, v any) error
		}{"Decoder of " + read.name, func(data []byte, v any) error {
			return newDecoder(read.reader(data)).Decode(v)
		}}, struct {
			name   string
			decode func(data []byte, v any) error
		}{"Decoder after a value, of " + read.name, func(data []byte, v any) error {
			dec := newDecoder(read.reader(append([]byte("0\n"), data...)))
			var prior any
			if err := dec.Decode(&prior); err != nil {
				return fmt.Errorf("the value before: %w", err)
			}
			return dec.Decode(v)
		}})
	}
	return decoders
}

// asGo127Token reads the input of c by its steps, by a Decoder of r, and returns the first error.
func asGo127Token(c asGo127TokenCase, dec asGo127Decoder) error {
	targets := c.targets
	for _, step := range c.steps {
		if step == 'T' {
			if _, err := dec.Token(); err != nil {
				return err
			}
			continue
		}
		if err := dec.Decode(targets[0]()); err != nil {
			return err
		}
		targets = targets[1:]
	}
	return nil
}

// asGo127Results returns the records of the decodings of the cases by go-json, or by encoding/json if std is set,
// and reports a decoding whose record is not the one of Unmarshal, or of the Decoder of a bytes.Reader for the
// inputs read by Token.
func asGo127Results(t *testing.T, std bool) []asGo127Record {
	t.Helper()
	var records []asGo127Record
	decoders := asGo127Decoders(std)
	for _, c := range asGo127Cases {
		for _, layout := range []string{"compact", "padded"} {
			data := []byte(c.input)
			if layout == "padded" {
				data = []byte(asGo127Padded(c.input))
			}
			var first asGo127Record
			for i, d := range decoders {
				target := c.target()
				r := asGo127Describe(d.decode(data, target), target)
				r.Case, r.Layout = c.name, layout
				if i == 0 {
					first = r
					records = append(records, r)
				} else if r != first {
					t.Errorf("%s, %s, std %t: %s returns\n%+v\nwhere Unmarshal returns\n%+v", c.name, layout, std, d.name, r, first)
				}
			}
		}
	}
	for _, c := range asGo127TokenCases {
		var first asGo127Record
		for i, read := range asGo127Reads[:2] {
			var dec asGo127Decoder = asGo127GoDecoder{json.NewDecoder(read.reader([]byte(c.input)))}
			if std {
				dec = stdjson.NewDecoder(read.reader([]byte(c.input)))
			}
			r := asGo127Describe(asGo127Token(c, dec), nil)
			r.Case, r.Layout = c.name, "tokens"
			if i == 0 {
				first = r
				records = append(records, r)
			} else if r != first {
				t.Errorf("%s, std %t: the Decoder of %s returns\n%+v\nwhere the one of a bytes.Reader returns\n%+v", c.name, std, read.name, r, first)
			}
		}
	}
	return records
}

func TestDecodeAsGo127(t *testing.T) {
	got := asGo127Results(t, false)
	if *updateDecodeAsGo127 {
		if !stdDecodesAsGo127 {
			t.Fatal("the golden file is written from encoding/json of Go 1.27 with GOEXPERIMENT=jsonv2")
		}
		var b bytes.Buffer
		enc := stdjson.NewEncoder(&b)
		for _, r := range asGo127Results(t, true) {
			if err := enc.Encode(r); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(asGo127Golden, b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	f, err := os.Open(asGo127Golden)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var want []asGo127Record
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r asGo127Record
		if err := stdjson.Unmarshal(scanner.Bytes(), &r); err != nil {
			t.Fatal(err)
		}
		want = append(want, r)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if stdDecodesAsGo127 {
		// the golden file is the one of encoding/json
		if std := asGo127Results(t, true); !reflect.DeepEqual(std, want) {
			t.Errorf("encoding/json differs from %s: run the test with -update", asGo127Golden)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("got %d records, %s has %d", len(got), asGo127Golden, len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s, %s:\n got %+v\nwant %+v", got[i].Case, got[i].Layout, got[i], want[i])
		}
	}
}
