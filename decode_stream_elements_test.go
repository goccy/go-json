package json_test

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// A Decoder decodes an object or an array which its buffer doesn't hold whole element by element, so that its memory
// is bounded whatever the size of the value. These tests decode values which white space makes long enough for it,
// and compare the results with the ones of Unmarshal of each value, which decodes it whole.

type elementsInner struct {
	A int            `json:"a"`
	B string         `json:"b"`
	C []float64      `json:"c"`
	D map[string]int `json:"d"`
	E *elementsInner `json:"e"`
	F any            `json:"f"`
}

func (*elementsInner) String() string { return "elementsInner" }

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
	{"interface with methods holding a pointer", func() any { var v fmt.Stringer = &elementsInner{}; return &v }},
	{"bytes", func() any { return new([]byte) }},
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
	// the error of the decoding, which compareWithUnmarshal looks at
	typeErr   *json.UnmarshalTypeError
	syntaxErr *json.SyntaxError
	eof       bool
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
			res.typeErr = typeErr
		case errors.As(err, &syntaxErr):
			res.err = fmt.Sprintf("%v at %d", syntaxErr, syntaxErr.Offset)
			res.syntaxErr = syntaxErr
		case err != nil:
			res.err = fmt.Sprintf("%T %v", err, err)
			res.eof = err == io.ErrUnexpectedEOF || err == io.EOF
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
		// the options of Unmarshal which decodes as the set-up Decoder
		disallowUnknownFields, useNumber bool
	}{
		{"default", func(*json.Decoder) {}, false, false},
		{"DisallowUnknownFields", (*json.Decoder).DisallowUnknownFields, true, false},
		{"UseNumber", (*json.Decoder).UseNumber, false, true},
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
		var want []elementsResult
		for _, chunk := range []int{1, 64, 512, 4096} {
			what := fmt.Sprintf("case %d into %s, %s, by %d bytes: input %.300q", i, target.name, setUp.name, chunk, elementsShort(in))
			var got []elementsResult
			reader := &chunkReader{data: []byte(in), chunk: chunk}
			watchdog(t, what, func() { got = decodeElementsStream(reader, target.new, setUp.setUp) })
			read := int64(len(in) - len(reader.data))
			if msg := compareWithUnmarshal(in, read, got, target.new, setUp.disallowUnknownFields, setUp.useNumber); msg != "" {
				t.Fatalf("%s: %s", what, msg)
			}
			// the same results whatever the reads, but the syntax error, which is the one of the bytes read
			got = withSyntaxErrorKind(got)
			if want == nil {
				want = got
			} else if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s:\ngot  %.2000v\nwant %.2000v", what, got, want)
			}
		}
	}
}

// keptCases is the number of the streams TestDecodeStreamValuesKept decodes.
var keptCases = 100

// The values decoded from a stream keep their content after the values after them are decoded, whatever the reads,
// the option DecodeNoCopyString set or cleared on each call, and the Token calls between them: the buffers which their
// strings refer to are not read into again. Every value is kept until the end of the stream, and compared then with
// Unmarshal of its bytes. Token reads some strings at the top level, and the names of the members of some objects,
// whose values are decoded.
func TestDecodeStreamValuesKept(t *testing.T) {
	var noCopy json.DecodeOption
	json.DecodeNoCopyString()(&noCopy)
	setNoCopy := func(o *json.DecodeOption) { o.Flags |= noCopy.Flags }
	clearNoCopy := func(o *json.DecodeOption) { o.Flags &^= noCopy.Flags }
	r := rand.New(rand.NewSource(elementsSeed))
	for i := 0; i < keptCases; i++ {
		// the values, and whether each is read by Token, which reads a string at the top level as a token
		var texts []string
		var tokens []bool
		var b strings.Builder
		for n := 2 + r.Intn(6); len(texts) < n; {
			var v strings.Builder
			keptValue(r, &v, 3)
			text := v.String()
			texts = append(texts, text)
			tokens = append(tokens, (text[0] == '"' || text[0] == '{') && r.Intn(2) == 0)
			b.WriteString(text + " ")
		}
		in := b.String()
		want := make([]any, len(texts))
		for k, text := range texts {
			if err := json.Unmarshal([]byte(text), &want[k]); err != nil {
				t.Fatalf("%.80q: %v", text, err)
			}
		}
		for _, chunk := range []int{1, 7, 64, 512, 4096} {
			dec := json.NewDecoder(&chunkReader{data: []byte(in), chunk: chunk})
			got := make([]any, len(texts))
			for k := range texts {
				var err error
				switch {
				case tokens[k] && texts[k][0] == '{':
					got[k], err = keptMembersByToken(dec, r, setNoCopy, clearNoCopy)
				case tokens[k]:
					got[k], err = dec.Token()
				case r.Intn(2) == 0:
					err = dec.DecodeWithOption(&got[k], setNoCopy)
				default:
					err = dec.DecodeWithOption(&got[k], clearNoCopy)
				}
				if err != nil {
					t.Fatalf("case %d by %d: value %d: %v", i, chunk, k, err)
				}
			}
			for k := range texts {
				if !reflect.DeepEqual(got[k], want[k]) {
					t.Fatalf("case %d by %d: value %d is %#.200v at the end of the stream, want %#.200v", i, chunk, k, got[k], want[k])
				}
			}
		}
	}
}

// keptMembersByToken reads the object at the top level of dec by Token, the names of its members and the delimiters,
// and decodes their values, with DecodeNoCopyString set or cleared at random.
func keptMembersByToken(dec *json.Decoder, r *rand.Rand, opts ...json.DecodeOptionFunc) (map[string]any, error) {
	m := map[string]any{}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	for dec.More() {
		name, err := dec.Token()
		if err != nil {
			return nil, err
		}
		var v any
		if err := dec.DecodeWithOption(&v, opts[r.Intn(len(opts))]); err != nil {
			return nil, err
		}
		m[name.(string)] = v
	}
	_, err := dec.Token()
	return m, err
}

// keptValue writes a random value, nested up to depth, of strings of random letters, long and short, and numbers,
// with white space which makes some objects and arrays too long for a read.
func keptValue(r *rand.Rand, b *strings.Builder, depth int) {
	pad := func() {
		if r.Intn(4) == 0 {
			b.WriteString(strings.Repeat(" ", 600))
		}
	}
	switch n := r.Intn(6); {
	case depth > 0 && n < 2:
		open, end := byte('['), byte(']')
		if n == 0 {
			open, end = '{', '}'
		}
		b.WriteByte(open)
		for i, m := 0, r.Intn(5); i < m; i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			pad()
			if open == '{' {
				b.WriteString(`"k` + strconv.Itoa(r.Intn(100)) + `":`)
				pad()
			}
			keptValue(r, b, depth-1)
		}
		b.WriteByte(end)
	case n < 4:
		size := r.Intn(20)
		if n == 3 {
			size = r.Intn(700)
		}
		b.WriteByte('"')
		for i := 0; i < size; i++ {
			b.WriteByte(byte('a' + r.Intn(26)))
		}
		b.WriteByte('"')
	default:
		b.WriteString(strconv.Itoa(r.Intn(1000000)))
	}
}

// withSyntaxErrorKind returns the results with the kind of a syntax error only, whose message and offset are the ones
// of the bytes read when it is returned ( see compareWithUnmarshal ).
func withSyntaxErrorKind(results []elementsResult) []elementsResult {
	results = slices.Clone(results)
	for i := range results {
		if results[i].syntaxErr != nil {
			results[i].err, results[i].syntaxErr = "syntax error", nil
		}
	}
	return results
}

// compareWithUnmarshal compares the results of the decoding of the values of the input by a Decoder, which read
// the first read bytes of it, with the ones of Unmarshal of each value, which decodes it whole, with the options of
// the Decoder, and returns what differs, or "". A value decoded, or which has a type error, is the bytes up to the
// offset after it; the end of the input is in the rest of the input, and a syntax error in the rest of the bytes read:
// the message of an error at a character which the reads cut quotes the bytes of it which have been read. The offset of
// a type error of a value is from the end of the value before, or its start ( see typeErrorsFromValueStart ), as the
// offset of an error of Unmarshal of the bytes from there; the one of a syntax error is in the whole input. A value
// with an error other than a type error is not compared.
func compareWithUnmarshal(in string, read int64, got []elementsResult, newValue func() any, disallowUnknownFields, useNumber bool) string {
	prevEnd := int64(0)
	for k, res := range got {
		start := prevEnd
		for start < int64(len(in)) && strings.IndexByte(" \t\n\r", in[start]) >= 0 {
			start++
		}
		base := prevEnd
		if typeErrorsFromValueStart {
			base = start
		}
		if res.eof && start == int64(len(in)) {
			// the end of the input after the values
			return ""
		}
		end := int64(len(in))
		switch {
		case res.syntaxErr != nil:
			end = read
		case res.typeErr != nil || res.err == "":
			end = res.offset
		case !res.eof:
			// another error, after which the offset is the end of the value before: the value, which is valid, ends
			// where encoding/json ends it
			sd := stdjson.NewDecoder(strings.NewReader(in[start:]))
			var raw stdjson.RawMessage
			if sd.Decode(&raw) == nil {
				end = start + sd.InputOffset()
			}
		}
		v := newValue()
		err := unmarshalReference([]byte(in[base:end]), v, disallowUnknownFields, useNumber)
		var typeErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError
		switch {
		case res.typeErr != nil:
			if !errors.As(err, &typeErr) {
				return fmt.Sprintf("value %d: type error %v, Unmarshal %v", k, res.typeErr, err)
			}
			if typeErr.Struct != res.typeErr.Struct || typeErr.Field != res.typeErr.Field || typeErr.Offset != res.typeErr.Offset || typeErr.Value != res.typeErr.Value {
				return fmt.Sprintf("value %d: type error %+v, Unmarshal %+v", k, res.typeErr, typeErr)
			}
			// the value is decoded whole after a type error
			if b, merr := stdjson.Marshal(v); merr == nil && string(b) != res.value {
				return fmt.Sprintf("value %d with a type error: %.500s, Unmarshal %.500s", k, res.value, b)
			}
		case res.syntaxErr != nil:
			if !errors.As(err, &syntaxErr) || json.IsAtEnd(err) {
				return fmt.Sprintf("value %d: syntax error %v, Unmarshal %v", k, res.syntaxErr, err)
			}
			if base+syntaxErr.Offset != res.syntaxErr.Offset {
				return fmt.Sprintf("value %d: syntax error %v at %d, Unmarshal %v at %d from %d", k, res.syntaxErr, res.syntaxErr.Offset, syntaxErr, syntaxErr.Offset, base)
			}
		case res.eof:
			// the end of the input in the value, which is the end of the input of Unmarshal too
			if !json.IsAtEnd(err) {
				return fmt.Sprintf("value %d: the end of the input, Unmarshal %v", k, err)
			}
		case res.err != "":
			// another error, as one of a method, which Unmarshal returns as it is
			if err == nil || err.Error() != res.err[strings.IndexByte(res.err, ' ')+1:] {
				return fmt.Sprintf("value %d: %s, Unmarshal %v", k, res.err, err)
			}
		default:
			if err != nil {
				return fmt.Sprintf("value %d: decoded, Unmarshal %v", k, err)
			}
			b, merr := stdjson.Marshal(v)
			if merr != nil {
				b = []byte(merr.Error())
			}
			if string(b) != res.value {
				return fmt.Sprintf("value %d: %.500s, Unmarshal %.500s", k, res.value, b)
			}
		}
		prevEnd = res.offset
	}
	return ""
}

// elementsShort returns the input without its long runs of white space, to show it.
func elementsShort(in string) string {
	return strings.ReplaceAll(in, strings.Repeat(" ", 70000), "<pad>")
}

// FirstWin decodes the first key of a field, and skips the rest of an object once every field is decoded.
func TestDecodeStreamElementsFirstWin(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	in := `{"name":` + pad + `"a","name":"b","X":1,"Y":"y","unknown":[` + pad + `1]}`
	var want elementsOuter
	if err := json.UnmarshalWithOption([]byte(in), &want, json.DecodeFieldPriorityFirstWin()); err != nil {
		t.Fatal(err)
	}
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

// An object or an array decoded element by element is nested no deeper than a value decoded whole, and a nesting too
// deep is the error of Unmarshal.
func TestDecodeStreamElementsDepth(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		"[" + pad + strings.Repeat("[", 9998) + strings.Repeat("]", 9999),
		"[" + pad + strings.Repeat("[", 9999) + strings.Repeat("]", 10000),
		"[" + pad + strings.Repeat("[", 10000) + strings.Repeat("]", 10001),
		// the deepest array too long to be decoded whole, at the limit and past it
		strings.Repeat("[", 10000) + pad + strings.Repeat("]", 10000),
		strings.Repeat("[", 10001) + pad + strings.Repeat("]", 10001),
		// arrays which are each too long to be decoded whole
		strings.Repeat("[", 100000) + strings.Repeat("]", 100000),
	} {
		var v any
		got := json.NewDecoder(&chunkReader{data: []byte(in), chunk: 1 << 20}).Decode(&v)
		want := unmarshalReference([]byte(in), &v, false, false)
		if msg := sameError(want, got, 0); msg != "" {
			t.Fatalf("%d bytes: %s", len(in), msg)
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
