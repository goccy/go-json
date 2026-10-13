package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// The errors of a value decoded element by element are the ones of the value read at once, whatever the size of the
// value, and a string which doesn't end fails at its control character wherever it is.

type elementsMethods interface{ M() }

// A string which is not closed fails at the end of its line, whatever the destination, without reading the rest of
// the input, as encoding/json does ( issue 701 ).
func TestDecodeStreamUnterminatedStringAnyDestination(t *testing.T) {
	rest := strings.Repeat(`{"name":"x","age":1}`+"\n", 1<<16)
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"name":"unterminated,"age":1}` + "\n" + rest,
		`[` + pad + `{"name":"x"},{"name":"unterminated,"age":1}` + "\n" + rest,
		`{"a":[` + pad + `1],"name":"unterminated,"age":1}` + "\n" + rest,
	} {
		for _, target := range []struct {
			name string
			new  func() any
		}{
			{"raw message", func() any { return new(json.RawMessage) }},
			{"map", func() any { return new(map[string]string) }},
			{"map of any", func() any { return new(map[string]any) }},
			{"interface with methods", func() any { return new(elementsMethods) }},
			{"slice of structs", func() any { return new([]struct{ Name string }) }},
			{"struct", func() any { return new(struct{ Name string }) }},
			{"string", func() any { return new(string) }},
		} {
			r := &countingReader{r: &chunkReader{data: []byte(in), chunk: 512}}
			err := json.NewDecoder(r).Decode(target.new())
			var serr *json.SyntaxError
			if !errors.As(err, &serr) {
				t.Fatalf("%.30q into %s: got %v", in, target.name, err)
			}
			if r.n > 256<<10 {
				t.Fatalf("%.30q into %s: %d bytes read for %v", in, target.name, r.n, err)
			}
		}
	}
}

// decodeAtOnceAndByChunks decodes in, one value, into a new value by newValue, at once ( see unmarshalReference ) and
// by a Decoder which reads it by chunks of chunk bytes, with DisallowUnknownFields if disallowUnknownFields is set, and
// returns the two errors, which sameError compares: the end of the input in the value is io.ErrUnexpectedEOF for both,
// as a Decoder reports it.
func decodeAtOnceAndByChunks(in string, chunk int, newValue func() any, disallowUnknownFields bool) (error, error) {
	whole := unmarshalReference([]byte(in), newValue(), disallowUnknownFields, false)
	if json.IsAtEnd(whole) {
		whole = io.ErrUnexpectedEOF
	}
	chunked := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk})
	if disallowUnknownFields {
		chunked.DisallowUnknownFields()
	}
	return whole, chunked.Decode(newValue())
}

// unmarshalReference decodes data, one value, by Unmarshal with the options of a Decoder, which decodes it whole, and
// returns the error which a Decoder which reads the value at once returns: the value is read whole first, as
// encoding/json reads it, so a syntax error of it is the error, where Unmarshal returns the error of a method or of
// an unknown field before it reads the rest of its input.
func unmarshalReference(data []byte, v any, disallowUnknownFields, useNumber bool) error {
	err := json.UnmarshalLikeDecoder(data, v, disallowUnknownFields, useNumber)
	var syntaxErr *json.SyntaxError
	if err != nil && !errors.As(err, &syntaxErr) {
		if serr := json.Unmarshal(data, new(json.RawMessage)); serr != nil {
			return serr
		}
	}
	return err
}

// sameError returns what differs between want, the error of a value read at once, and got, the one of a Decoder, or
// "": both have none, or are the end of the input, or are syntax errors of the same message at the same offset, from
// base for want, or are type errors of the same struct, field, offset and value, or are other errors of the same
// message.
func sameError(want, got error, base int64) string {
	var wantSyntax, gotSyntax *json.SyntaxError
	var wantType, gotType *json.UnmarshalTypeError
	switch {
	case want == nil || got == nil:
		if want != got {
			return fmt.Sprintf("got %v, want %v", got, want)
		}
	case want == io.ErrUnexpectedEOF || got == io.ErrUnexpectedEOF:
		if want != got {
			return fmt.Sprintf("got %v, want %v", got, want)
		}
	case errors.As(want, &wantSyntax):
		if !errors.As(got, &gotSyntax) || gotSyntax.Offset != base+wantSyntax.Offset || got.Error() != want.Error() {
			return fmt.Sprintf("got %v, want %v at %d", got, want, base+wantSyntax.Offset)
		}
	case errors.As(want, &wantType):
		if !errors.As(got, &gotType) || gotType.Struct != wantType.Struct || gotType.Field != wantType.Field ||
			gotType.Offset != wantType.Offset || gotType.Value != wantType.Value {
			return fmt.Sprintf("got %+v, want %+v", got, want)
		}
	default:
		if got.Error() != want.Error() {
			return fmt.Sprintf("got %v, want %v", got, want)
		}
	}
	return ""
}

// The depth of a value is counted from the root for a syntax error in an element, and the depth limit of encoding/json
// applies to the value as it does to a value read at once, whatever the reads: a nesting too deep is the error of
// Unmarshal, at the same offset.
func TestDecodeStreamElementsDepthErrors(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"a":` + pad + strings.Repeat(`{"a":`, 19999) + `null` + strings.Repeat("}", 20000),
		`{"a":` + pad + strings.Repeat(`{"a":`, 10001),
		`[` + pad + strings.Repeat(`[`, 9998) + `x` + strings.Repeat(`]`, 9999),
	} {
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in, chunk, func() any { return new(any) }, false)
			if msg := sameError(want, got, 0); msg != "" {
				t.Fatalf("%d bytes, chunk %d: %s", len(in), chunk, msg)
			}
		}
	}
}

// An unknown field is an error once the value is read, as the one of a value read at once is: the end of the input
// before the end of the value is the error, as encoding/json reports it.
func TestDecodeStreamElementsUnknownFieldError(t *testing.T) {
	type S struct{ A, B int }
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"a":1,` + pad + `"zz":1,"b":2`,
		`{"a":1,` + pad + `"zz":1,"b":2}`,
		`{"a":1,` + pad + `"zz":"unterminated` + "\n" + strings.Repeat(`{"b":2}`, 1<<16),
		// a syntax error after the unknown field is the error, as encoding/json reads the value first
		`{"a":1,` + pad + `"zz":1,"b":2 x`,
		`{"a":1,` + pad + `"zz":1,"b":[1 2`,
		`{"a":1,` + pad + `"zz":1,"b":"x` + "\x01",
		`{"a":1,` + pad + `"zz":{"c":[1,{"d":2}]},"b":2}`,
	} {
		var stdErr error
		{
			dec := stdjson.NewDecoder(strings.NewReader(in))
			dec.DisallowUnknownFields()
			stdErr = dec.Decode(new(S))
		}
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in, chunk, func() any { return new(S) }, true)
			if msg := sameError(want, got, 0); msg != "" {
				t.Fatalf("%.20q, chunk %d: %s (encoding/json: %v)", in, chunk, msg, stdErr)
			}
			var stdSyntax *stdjson.SyntaxError
			var syntax *json.SyntaxError
			if errors.Is(stdErr, io.ErrUnexpectedEOF) != errors.Is(got, io.ErrUnexpectedEOF) ||
				errors.As(stdErr, &stdSyntax) != errors.As(got, &syntax) {
				t.Fatalf("%.20q, chunk %d: got %v, encoding/json %v", in, chunk, got, stdErr)
			}
		}
	}
}

type elementsFailing struct{}

func (*elementsFailing) UnmarshalJSON([]byte) error { return errors.New("elementsFailing") }

// The error of a method is returned once the value is read, as the one of a value read at once is.
func TestDecodeStreamElementsMethodError(t *testing.T) {
	type S struct {
		A []elementsFailing
		B int
	}
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"A":[` + pad + `{},{}],"B":1}`,
		`{"A":[` + pad + `{},{}],"B":1 x`,
		`{"A":[` + pad + `{},{}],"B":[1 2`,
		`{"A":[` + pad + `{},{}],"B":1`,
		`[` + pad + `{"A":[{}]},{"B":2}]`,
	} {
		newValue := func() any { return new(S) }
		if in[0] == '[' {
			newValue = func() any { return new([]S) }
		}
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in, chunk, newValue, false)
			if msg := sameError(want, got, 0); msg != "" {
				t.Fatalf("%.20q, chunk %d: %s", in, chunk, msg)
			}
		}
	}
}

// elementsSeen keeps the input of every call of its method, which fails for a value which is not an object.
type elementsSeen struct {
	seen []string
}

func (e *elementsSeen) UnmarshalJSON(b []byte) error {
	e.seen = append(e.seen, string(b))
	if b[0] != '{' {
		return errors.New("elementsSeen: not an object")
	}
	return nil
}

// The method of an element which fails for a number is called once for it, as Unmarshal calls it.
func TestDecodeStreamElementsMethodErrorOfNumber(t *testing.T) {
	pad := strings.Repeat(" ", 1000)
	for _, in := range []string{`[` + pad + `{},5,{}]`, `[` + pad + `5,{}]`} {
		var want, got []elementsSeen
		wantErr := json.Unmarshal([]byte(in), &want)
		err := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 4096}).Decode(&got)
		if msg := sameError(wantErr, err, 0); wantErr == nil || msg != "" {
			t.Fatalf("%.20q: %s", in, msg)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%.20q: got %+v, want %+v", in, got, want)
		}
	}
}

type elementsBadKey string

func (k *elementsBadKey) UnmarshalText(b []byte) error {
	if string(b) == "bad" {
		return errors.New("elementsBadKey")
	}
	*k = elementsBadKey(b)
	return nil
}

// The error of the method of a key is returned once the value is read, as the one of a value read at once is: a syntax
// error or the end of the input after it is the error.
func TestDecodeStreamElementsKeyError(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	for _, in := range []struct {
		in  string
		new func() any
	}{
		{`{"C":{` + pad + `"bad":1,"ok":2}}`, func() any { return new(struct{ C map[elementsBadKey]int }) }},
		{`{"C":{` + pad + `"bad":1,"ok":2 x}}`, func() any { return new(struct{ C map[elementsBadKey]int }) }},
		{`{` + pad + `"bad":{},"ok":{} x}`, func() any { return new(map[elementsBadKey]any) }},
		{`{` + pad + `"bad":1,"x":2,`, func() any { return new(map[elementsBadKey]int) }},
		{`{` + pad + `"bad":1,"x":2}`, func() any { return new(map[elementsBadKey]int) }},
	} {
		var stdErr error
		{
			v := in.new()
			stdErr = stdjson.NewDecoder(strings.NewReader(in.in)).Decode(v)
		}
		for _, chunk := range []int{1000, 65536} {
			want, got := decodeAtOnceAndByChunks(in.in, chunk, in.new, false)
			if msg := sameError(want, got, 0); msg != "" {
				t.Fatalf("%.30q, chunk %d: %s (encoding/json: %v)", in.in, chunk, msg, stdErr)
			}
			var stdSyntax *stdjson.SyntaxError
			var syntax *json.SyntaxError
			if errors.Is(stdErr, io.ErrUnexpectedEOF) != errors.Is(got, io.ErrUnexpectedEOF) ||
				errors.As(stdErr, &stdSyntax) != errors.As(got, &syntax) {
				t.Fatalf("%.30q, chunk %d: got %v, encoding/json %v", in.in, chunk, got, stdErr)
			}
		}
	}
}

// The index of the brackets of a long array keeps only the part not decoded yet, whatever the length of the array.
func TestDecodeStreamElementsIndexBounded(t *testing.T) {
	element := "," + strings.Repeat(" ", 61) + "{}"
	measure := func(n int) uint64 {
		r := &chunkReader{data: []byte("[{}" + strings.Repeat(element, n) + "]"), chunk: 4096}
		var v []struct{}
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		if err := json.NewDecoder(r).Decode(&v); err != nil {
			t.Fatal(err)
		}
		runtime.ReadMemStats(&after)
		if len(v) != n+1 {
			t.Fatalf("got %d elements", len(v))
		}
		return after.TotalAlloc - before.TotalAlloc
	}
	small, large := measure(1<<12), measure(1<<17)
	if large > small+(1<<20) {
		t.Fatalf("%d bytes allocated for %d elements, %d for %d", large, 1<<17, small, 1<<12)
	}
}

// The length of a slice decoded element by element counts the elements decoded before an error, never the element
// which failed. A slice decoded whole from the buffer may count it ( see sliceDecoder.Decode ): the value after an error
// is not specified.
func TestDecodeStreamElementsSliceAfterError(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `[` + pad + `1,2,3,{]`
	for _, prior := range [][]int{nil, {9, 9, 9, 9, 9}} {
		v := append([]int(nil), prior...)
		if err := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 4096}).Decode(&v); err == nil {
			t.Fatal("no error")
		}
		if !reflect.DeepEqual(v, []int{1, 2, 3}) {
			t.Fatalf("prior %v: got %v", prior, v)
		}
	}
}

// A number which a read cuts is decoded once it is whole: its first digits are not written to the field, which the
// type error of the whole number leaves as it was.
func TestDecodeStreamElementsCutNumber(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `{"A":1,` + pad + `"X":1e400,"B":2}`
	type target struct{ A, X, B int }
	var whole, chunked target
	werr := json.NewDecoder(bytes.NewReader([]byte(in))).Decode(&whole)
	cerr := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 1}).Decode(&chunked)
	if werr == nil || cerr == nil || werr.Error() != cerr.Error() || whole != chunked {
		t.Fatalf("got %+v, %v, want %+v, %v", chunked, cerr, whole, werr)
	}
}

// An input which ends in an integer of a value which is skipped, or read whole, ends unexpectedly, whatever the reads:
// the integer which a read cuts is read on, and the end of the input after it is the error.
func TestDecodeStreamElementsEndInInteger(t *testing.T) {
	pad := strings.Repeat(" ", 1000)
	for _, in := range []string{
		`{"unknown":[` + pad + `1,12`,
		`{"unknown":{"a":` + pad + `12`,
		`[` + pad + `1,12`,
	} {
		for _, chunk := range []int{1, 7, 512} {
			for _, target := range []func() any{
				func() any { return new(struct{ Name string }) },
				func() any { return new(json.RawMessage) },
			} {
				whole, chunked := decodeAtOnceAndByChunks(in, chunk, target, false)
				if msg := sameError(whole, chunked, 0); whole == nil || msg != "" {
					t.Fatalf("%q by chunks of %d: %s", in[:12], chunk, msg)
				}
			}
		}
	}
}

type pooledElem struct {
	A string
	B int
	C []int
}

// An element which fails, whole or in part decoded, doesn't go back to the buffers which a later decoding of the type
// decodes into: the elements of a later array have nothing of it.
func TestDecodeStreamElementsFailedElementNotKept(t *testing.T) {
	var head strings.Builder
	head.WriteString("[")
	for i := 0; i < 64; i++ {
		fmt.Fprintf(&head, `{"B":%d},`, i)
	}
	var later strings.Builder
	later.WriteString("[")
	for i := 0; i < 100; i++ {
		if i > 0 {
			later.WriteString(",")
		}
		fmt.Fprintf(&later, `{"B":%d}`, i)
	}
	later.WriteString("]")
	for _, failing := range []string{`{"A":"kept","C":[1,2,3],"B":`, `{"A":"kept","C":[1,2,3] "B":1}]`} {
		var v []pooledElem
		if err := json.NewDecoder(strings.NewReader(head.String() + failing)).Decode(&v); err == nil {
			t.Fatalf("%q: no error", failing)
		}
		var w []pooledElem
		if err := json.Unmarshal([]byte(later.String()), &w); err != nil {
			t.Fatal(err)
		}
		for i, e := range w {
			if e.A != "" || e.C != nil {
				t.Fatalf("after %q: element %d of a later array is %+v", failing, i, e)
			}
		}
	}
}

type pointeeRecord struct {
	N     int `json:"n"`
	Inner struct {
		M int `json:"m"`
	} `json:"inner"`
}

func (*pointeeRecord) String() string { return "pointeeRecord" }

// A type error in what a pointer held by an interface value points to is the one of Unmarshal, whether a read cuts
// the value of the interface, which is then decoded element by element, or not, for an interface{} and an interface
// type with methods.
func TestDecodeStreamElementsInterfacePointeeTypeError(t *testing.T) {
	good := `{"Type":"x","Data":{"n":1}}` + "\n"
	for _, newValue := range []func() any{
		func() any {
			return &struct {
				Type string
				Data any
			}{Data: &pointeeRecord{}}
		},
		func() any {
			return &struct {
				Type string
				Data fmt.Stringer
			}{Data: &pointeeRecord{}}
		},
	} {
		for _, bad := range []string{`{"Type":"x","Data":{"n":"str"}}` + "\n", `{"Type":"x","Data":{"inner":{"m":"str"}}}` + "\n"} {
			// the record with the error starts at 476 with no space before it, which the first read of 512 bytes
			// holds whole, and at 486 with 10 spaces, where the read cuts the value of Data: every other place is
			// tried too
			for spaces := 0; spaces < len(bad); spaces++ {
				in := strings.Repeat(good, 17) + strings.Repeat(" ", spaces) + bad + strings.Repeat(good, 5)
				got := decodeElementsStream(strings.NewReader(in), newValue, func(*json.Decoder) {})
				what := fmt.Sprintf("%T, %d spaces before %q", newValue(), spaces, bad)
				if msg := compareWithUnmarshal(in, int64(len(in)), got, newValue, false, false); msg != "" {
					t.Fatalf("%s: %s", what, msg)
				}
				if len(got) != 24 || got[17].typeErr == nil {
					t.Fatalf("%s: got %+v, want the type error of the record 17 of 24", what, got)
				}
			}
		}
	}
}

type inputStepsBytes struct {
	I []byte `json:"i"`
}

type inputStepsMid struct {
	Y *struct {
		X int `json:"x"`
	} `json:"y"`
}

func (*inputStepsMid) String() string { return "inputStepsMid" }

// The path of a type error in what a pointer held by an interface value points to has every step of the input to the
// value, as the one of Unmarshal, whether the value is decoded element by element or not: the steps in an array decoded
// into []byte, and the ones in an interface type with methods in an interface{}, which the decoders of the types don't
// take.
func TestDecodeStreamElementsInterfacePointeeInputSteps(t *testing.T) {
	pad := strings.Repeat(" ", 600)
	for _, c := range []struct {
		in       string
		newValue func() any
	}{
		{`{"i":[[],1]}`, func() any { var v any = &inputStepsBytes{}; return &v }},
		{`{"f":{"i":[1,[],1]}}`, func() any { return &struct{ F any }{F: &inputStepsBytes{}} }},
		{`{"i":[1,"x"]}`, func() any { return &inputStepsBytes{} }},
		{`{"i":[1,"x",3]}`, func() any { var v any = &inputStepsBytes{}; return &v }},
		{`{"s":{"y":{"x":"str"}}}`, func() any {
			var v any = &struct{ S fmt.Stringer }{S: &inputStepsMid{}}
			return &v
		}},
		{`{"s":{"y":{"x":"str"}}}`, func() any { return &struct{ S fmt.Stringer }{S: &inputStepsMid{}} }},
	} {
		// as it is, which a read of 512 bytes holds whole, with white space which makes the outer object too long to
		// be decoded whole, and with white space which makes each object and array so
		for _, in := range []string{c.in, "{" + pad + c.in[1:], strings.NewReplacer(":", ":"+pad, "[", "["+pad).Replace(c.in)} {
			for _, chunk := range []int{1, 64, 512, 4096} {
				got := decodeElementsStream(&chunkReader{data: []byte(in), chunk: chunk}, c.newValue, func(*json.Decoder) {})
				what := fmt.Sprintf("%.60q into %T by %d", in, c.newValue(), chunk)
				if msg := compareWithUnmarshal(in, int64(len(in)), got, c.newValue, false, false); msg != "" {
					t.Fatalf("%s: %s", what, msg)
				}
				if len(got) == 0 || got[0].typeErr == nil {
					t.Fatalf("%s: got %+v, want a type error", what, got)
				}
			}
		}
	}
}

// methodSyntaxError fails for "bad" with the syntax error of Unmarshal of an object which is not closed: the error of
// a method which is of the type of the errors of the grammar.
type methodSyntaxError struct{}

func (*methodSyntaxError) UnmarshalJSON(b []byte) error {
	if string(b) == `"bad"` {
		var v any
		return json.Unmarshal([]byte("{"), &v)
	}
	return nil
}

// methodTypeError decodes its value by a type of its own, whose type error is the error of its method.
type methodTypeError struct {
	A int `json:"a"`
}

func (m *methodTypeError) UnmarshalJSON(b []byte) error {
	type plain methodTypeError
	return json.Unmarshal(b, (*plain)(m))
}

// The error of an unmarshal method, whatever its type, is the one of Unmarshal, wherever the reads cut the input: the
// end of the input or a syntax error after the value is the error then, and a type error keeps the offset which the
// method gave.
func TestDecodeStreamElementsMethodErrorTypes(t *testing.T) {
	pad := strings.Repeat(" ", 1000)
	syntaxErrors := func() any { return new([]methodSyntaxError) }
	typeErrors := func() any { return new([]methodTypeError) }
	for _, c := range []struct {
		in       string
		newValue func() any
	}{
		{`["ok", "bad", `, syntaxErrors},
		{`[` + pad + `"ok", "bad", `, syntaxErrors},
		{`["ok", "bad", "ok" "ok"]`, syntaxErrors},
		{`[` + pad + `"ok", "bad", ` + pad + `"ok" "ok"]`, syntaxErrors},
		{`[{"a":1}, {"a":"x"}, {"a":2}]`, typeErrors},
		{`[{"a":1}, ` + pad + `{"a":"x"}, {"a":2}]`, typeErrors},
	} {
		for _, chunk := range []int{1, 64, 512, 4096} {
			want, got := decodeAtOnceAndByChunks(c.in, chunk, c.newValue, false)
			if msg := sameError(want, got, 0); want == nil || msg != "" {
				t.Fatalf("%.30q by %d: %s, want %v", c.in, chunk, msg, want)
			}
		}
	}
}
