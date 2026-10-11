package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// A Decoder decodes an object or an array which its buffer doesn't hold whole element by element, so that its memory
// is bounded whatever the size of the value. These tests decode values which white space makes long enough for it,
// and compare the results with the ones of the same input read at once, which is decoded whole.

type elementsInner struct {
	A int            `json:"a"`
	B string         `json:"b"`
	C []float64      `json:"c"`
	D map[string]int `json:"d"`
	E *elementsInner `json:"e"`
	F any            `json:"f"`
}

type elementsKey string

func (k *elementsKey) UnmarshalText(b []byte) error {
	if len(b) > 0 && b[0] == 'x' {
		return errors.New("elementsKey: x")
	}
	*k = elementsKey("k" + string(b))
	return nil
}

type elementsEmbedded struct {
	X int
	Y string
}

type elementsUnmarshaler struct {
	raw string
}

func (u *elementsUnmarshaler) UnmarshalJSON(b []byte) error {
	u.raw = string(b)
	return nil
}

// elementsHidden is an unexported struct type, a nil pointer to which is not allocated: its fields can't be set.
type elementsHidden struct {
	Z int
}

type elementsOuter struct {
	elementsEmbedded
	*elementsHidden
	Name   string                `json:"name"`
	Items  []elementsInner       `json:"items"`
	Arr    [2]elementsInner      `json:"arr"`
	ByID   map[int]elementsInner `json:"by_id"`
	ByKey  map[elementsKey]int   `json:"by_key"`
	Any    any                   `json:"any"`
	Ptr    *elementsInner        `json:"ptr"`
	Num    json.Number           `json:"num"`
	Str    int                   `json:"str,string"`
	Raw    json.RawMessage       `json:"raw"`
	Method elementsUnmarshaler   `json:"method"`
}

var elementsTargets = []struct {
	name string
	new  func() any
}{
	{"struct", func() any { return new(elementsOuter) }},
	{"pointer to struct", func() any { return new(*elementsOuter) }},
	{"any", func() any { return new(any) }},
	{"map", func() any { return new(map[string]any) }},
	{"map of structs", func() any { return new(map[string]elementsInner) }},
	{"map of ints", func() any { return new(map[int]any) }},
	{"slice", func() any { return new([]elementsInner) }},
	{"slice of any", func() any { return new([]any) }},
	{"array", func() any { return new([3]any) }},
	{"interface holding a pointer", func() any { var v any = &elementsInner{}; return &v }},
	// values which are read whole, however long
	{"string", func() any { return new(string) }},
	{"raw message", func() any { return new(json.RawMessage) }},
	{"unmarshaler", func() any { return new(elementsUnmarshaler) }},
	{"map of bool keys", func() any { return new(map[bool]int) }},
	{"interface with methods", func() any { return new(fmt.Stringer) }},
}

var elementsKeys = []string{
	"a", "b", "c", "d", "e", "f", "A", "name", "items", "arr", "by_id", "by_key", "any", "ptr", "num", "str", "raw",
	"method", "X", "Y", "Z", "unknown", `a`, "日本", "1", "-2", "x1",
}

// elementsValue writes a random value, nested up to depth.
func elementsValue(r *rand.Rand, b *strings.Builder, depth int) {
	n := r.Intn(10)
	if depth <= 0 {
		n = 4 + r.Intn(6)
	}
	switch n {
	case 0, 1:
		b.WriteByte('{')
		for i, m := 0, r.Intn(5); i < m; i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(`"` + elementsKeys[r.Intn(len(elementsKeys))] + `":`)
			elementsValue(r, b, depth-1)
		}
		b.WriteByte('}')
	case 2, 3:
		b.WriteByte('[')
		for i, m := 0, r.Intn(5); i < m; i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			elementsValue(r, b, depth-1)
		}
		b.WriteByte(']')
	case 4, 5:
		b.WriteString(strconv.Itoa(r.Intn(2000) - 1000))
	case 6:
		b.WriteString([]string{"1.5", "-0.25e3", "1e400", "123456789012345678901"}[r.Intn(4)])
	case 7:
		b.WriteString([]string{`"s"`, `"5"`, `"é\né"`, `""`, `"x1"`, `"` + strings.Repeat("l", 70000) + `"`}[r.Intn(6)])
	case 8:
		b.WriteString([]string{"true", "false"}[r.Intn(2)])
	default:
		b.WriteString("null")
	}
}

// elementsPad inserts white space after the brackets, the commas and the colons out of the strings of the input, so
// that the objects and the arrays which have it are too long to be decoded whole.
func elementsPad(r *rand.Rand, in string) string {
	var b strings.Builder
	inString := false
	for i := 0; i < len(in); i++ {
		c := in[i]
		b.WriteByte(c)
		switch {
		case inString:
			if c == '\\' {
				i++
				b.WriteByte(in[i])
			} else if c == '"' {
				inString = false
			}
		case c == '"':
			inString = true
		case c == '{' || c == '[' || c == ',' || c == ':':
			if r.Intn(4) == 0 {
				b.WriteString(strings.Repeat(" ", 70000))
			}
		}
	}
	return b.String()
}

// elementsMutate breaks the input, so that it has a syntax error, or ends early.
func elementsMutate(r *rand.Rand, in string) string {
	switch r.Intn(6) {
	case 0:
		return in[:r.Intn(len(in))]
	case 1, 2:
		i := r.Intn(len(in))
		return in[:i] + string("{}[]:,\"x0\\\n"[r.Intn(11)]) + in[i+1:]
	}
	return in
}

type elementsResult struct {
	value  string
	err    string
	offset int64
}

// decodeElementsStream decodes the values of the input by a Decoder, until the input or a decoding fails.
func decodeElementsStream(r io.Reader, newValue func() any, setUp func(*json.Decoder)) []elementsResult {
	dec := json.NewDecoder(r)
	setUp(dec)
	var results []elementsResult
	for {
		v := newValue()
		err := dec.Decode(v)
		var res elementsResult
		var typeErr *json.UnmarshalTypeError
		if err == nil || errors.As(err, &typeErr) {
			b, merr := stdjson.Marshal(v)
			if merr != nil {
				b = []byte(merr.Error())
			}
			res.value = string(b)
		}
		var syntaxErr *json.SyntaxError
		switch {
		case typeErr != nil:
			res.err = fmt.Sprintf("%v at %d: %s %s %s", typeErr, typeErr.Offset, typeErr.Value, typeErr.Struct, typeErr.Field)
		case errors.As(err, &syntaxErr):
			res.err = fmt.Sprintf("%v at %d", syntaxErr, syntaxErr.Offset)
		case err != nil:
			res.err = fmt.Sprintf("%T %v", err, err)
		}
		res.offset = dec.InputOffset()
		results = append(results, res)
		if err != nil && typeErr == nil {
			return results
		}
	}
}

// elementsCases is the number of the inputs TestDecodeStreamElements decodes.
var (
	elementsCases       = 300
	elementsSeed  int64 = 1
)

func TestDecodeStreamElements(t *testing.T) {
	r := rand.New(rand.NewSource(elementsSeed))
	setUps := []struct {
		name  string
		setUp func(*json.Decoder)
	}{
		{"default", func(*json.Decoder) {}},
		{"DisallowUnknownFields", (*json.Decoder).DisallowUnknownFields},
		{"UseNumber", (*json.Decoder).UseNumber},
	}
	for i := 0; i < elementsCases; i++ {
		var b strings.Builder
		elementsValue(r, &b, 4)
		in := b.String()
		if r.Intn(3) == 0 {
			// one more value after it
			b.WriteString(" ")
			elementsValue(r, &b, 2)
			in = b.String()
		}
		in = elementsPad(r, in)
		if r.Intn(2) == 0 {
			in = elementsMutate(r, in)
		}
		target := elementsTargets[r.Intn(len(elementsTargets))]
		setUp := setUps[r.Intn(len(setUps))]
		// read at once, which is decoded whole
		want := decodeElementsStream(bytes.NewReader([]byte(in)), target.new, setUp.setUp)
		for _, chunk := range []int{1, 1000, 4096, 65536} {
			got := decodeElementsStream(&chunkReader{data: []byte(in), chunk: chunk}, target.new, setUp.setUp)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("case %d into %s, %s, chunk %d: input %q\ngot  %.2000v\nwant %.2000v",
					i, target.name, setUp.name, chunk, elementsShort(in), got, want)
			}
		}
	}
}

// elementsShort returns the input without its long runs of white space, to show it.
func elementsShort(in string) string {
	return strings.ReplaceAll(in, strings.Repeat(" ", 70000), "<pad>")
}

// FirstWin decodes the first key of a field, and skips the rest of an object once every field is decoded.
func TestDecodeStreamElementsFirstWin(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `{"name":` + pad + `"a","name":"b","X":1,"Y":"y","unknown":[` + pad + `1]}`
	want := decodeElementsFirstWin(t, bytes.NewReader([]byte(in)))
	got := decodeElementsFirstWin(t, &chunkReader{data: []byte(in), chunk: 512})
	if !reflect.DeepEqual(got, want) || got.Name != "a" || got.X != 1 {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func decodeElementsFirstWin(t *testing.T, r io.Reader) elementsOuter {
	t.Helper()
	var v elementsOuter
	if err := json.NewDecoder(r).DecodeWithOption(&v, json.DecodeFieldPriorityFirstWin()); err != nil {
		t.Fatal(err)
	}
	return v
}

// generatedReader reads an object of n records, which it makes as it is read: {"items":[{...},...],"count":n}.
type generatedReader struct {
	n, i    int
	pending []byte
	// read is the number of the bytes read, and peak the largest live heap seen in the reads.
	read, peak uint64
}

func (g *generatedReader) Read(p []byte) (int, error) {
	if g.read%(1<<20) < uint64(len(p)) {
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		g.peak = max(g.peak, m.HeapAlloc)
	}
	for len(g.pending) == 0 {
		switch {
		case g.i == 0:
			g.pending = []byte(`{"items":[`)
		case g.i <= g.n:
			sep := ","
			if g.i == 1 {
				sep = ""
			}
			g.pending = fmt.Appendf(nil, `%s{"a":%d,"b":"%s","c":[1.5,2.5],"d":{"k":1}}`, sep, g.i, strings.Repeat("v", 100))
		case g.i == g.n+1:
			g.pending = fmt.Appendf(nil, `],"count":%d}`, g.n)
		default:
			return 0, io.EOF
		}
		g.i++
	}
	n := copy(p, g.pending)
	g.pending = g.pending[n:]
	g.read += uint64(n)
	return n, nil
}

// A value much longer than the buffer is decoded with memory bounded by what is decoded, not by its size.
func TestDecodeStreamElementsMemory(t *testing.T) {
	const records = 150000 // about 22 MB
	var base runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&base)

	t.Run("skipped", func(t *testing.T) {
		r := &generatedReader{n: records}
		var v struct {
			Count int `json:"count"`
		}
		if err := json.NewDecoder(r).Decode(&v); err != nil {
			t.Fatal(err)
		}
		if v.Count != records {
			t.Fatalf("got %d", v.Count)
		}
		live := int64(r.peak) - int64(base.HeapAlloc)
		if live > 4<<20 {
			t.Fatalf("%d bytes read, %d bytes live at most", r.read, live)
		}
	})
	t.Run("decoded", func(t *testing.T) {
		r := &generatedReader{n: records / 3}
		var v struct {
			Items []struct {
				A int    `json:"a"`
				B string `json:"b"`
			} `json:"items"`
		}
		if err := json.NewDecoder(r).Decode(&v); err != nil {
			t.Fatal(err)
		}
		if len(v.Items) != records/3 || v.Items[9].A != 10 {
			t.Fatalf("got %d items", len(v.Items))
		}
		// the decoded items, 50000 of about 100 bytes of strings and 24 bytes of the struct, which a slice holds
		// with room to grow
		live := int64(r.peak) - int64(base.HeapAlloc)
		if live > 16<<20 {
			t.Fatalf("%d bytes read, %d bytes live at most", r.read, live)
		}
	})
}

// An object or an array decoded element by element is nested no deeper than a value decoded whole.
func TestDecodeStreamElementsDepth(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		"[" + pad + strings.Repeat("[", 9998) + strings.Repeat("]", 9999),
		"[" + pad + strings.Repeat("[", 9999) + strings.Repeat("]", 10000),
		"[" + pad + strings.Repeat("[", 10000) + strings.Repeat("]", 10001),
		// arrays which are each too long to be decoded whole
		strings.Repeat("[", 100000) + strings.Repeat("]", 100000),
	} {
		var v any
		got := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 1 << 20}).Decode(&v)
		want := json.NewDecoder(bytes.NewReader([]byte(in))).Decode(&v)
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("%d bytes: got %v, want %v", len(in), got, want)
		}
	}
}

// countingReader reads the input, and counts the bytes read.
type countingReader struct {
	r io.Reader
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

// A string which is not closed is a syntax error at the first control character, which a string can't have: the
// rest of the input is not read, as encoding/json doesn't ( issue 701 ).
func TestDecodeStreamUnterminatedString(t *testing.T) {
	rest := strings.Repeat(`{"name":"x","age":1}`+"\n", 1<<16)
	for _, in := range []string{
		`{"name":"unterminated,"age":1}` + "\n" + rest,
		`"unterminated` + "\n" + rest,
		`[{"name":"unterminated` + "\n" + rest,
	} {
		r := &countingReader{r: struct{ io.Reader }{strings.NewReader(in)}}
		var v any
		err := json.NewDecoder(r).Decode(&v)
		var serr *json.SyntaxError
		if !errors.As(err, &serr) {
			t.Fatalf("%.20q: got %v", in, err)
		}
		if r.n > 256<<10 {
			t.Fatalf("%.20q: %d bytes read for %v", in, r.n, err)
		}
	}
}
