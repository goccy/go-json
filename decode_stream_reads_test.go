package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"strings"
	"testing"
	"time"

	json "github.com/goccy/go-json"
)

// A Decoder reads no more than the value it decodes needs: a value which is not valid fails at the read of its first
// byte which the grammar of JSON doesn't have where it is, as encoding/json, which reads a value while it checks it,
// fails there, and a value is decoded once its last byte is read, without a read more.

// oneByteReader reads data a byte at a time, and counts the bytes read.
type oneByteReader struct {
	data string
	n    int
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if r.n == len(r.data) {
		return 0, io.EOF
	}
	p[0] = r.data[r.n]
	r.n++
	return 1, nil
}

type readsUnmarshaler struct{}

func (*readsUnmarshaler) UnmarshalJSON([]byte) error { return nil }

func TestDecodeStreamFailsAtInvalidByte(t *testing.T) {
	rec := `{"name":"alice","age":30,"tags":["a","b"],"address":{"city":"Springfield"}}`
	tail := strings.Repeat(rec+"\n", 2000)
	flat := strings.Repeat(rec, 2000)
	// the errors after more than a read of valid bytes
	long := `{"items":[` + strings.Repeat(rec+",", 100)
	inputs := []string{
		`{"name":"alice` + "\n" + tail,
		`{"name":"alice` + "\t" + flat,
		`{"name":"alice` + flat,
		`[{"name":"alice` + "\n" + tail,
		rec + "\n" + `{"name":"bob` + "\n" + tail,
		`{"name":"a\` + "\n" + tail,
		`"unterminated` + "\n" + tail,
		`"unterminated\x` + tail,
		`{"name":"a\x` + flat,
		`{"name":"a\u12x` + flat,
		`{"age":01` + flat,
		`{"age":1.x` + flat,
		`{"age":tru,` + flat,
		`{"age" 1,` + flat,
		`{"a":[1,2 3` + flat,
		`[` + flat,
		`{"tags":["a" "b"]` + flat,
		`{"a":{"b":[{"c":"x` + "\n" + tail,
		long + rec + rec + tail,
		long + `{"name" "x"}` + tail,
		long + `{"name":tru}` + tail,
		long + `{"name":nul` + tail,
		long + `{"age":1.}` + tail,
		long + `{"age":-x}` + tail,
		long + `{"age":01}` + tail,
		long + `{"age":1e}` + tail,
		long + rec + `]]` + tail,
		long + rec + `}` + tail,
		long + rec + `,]` + tail,
		long + `{"name":"a\x"}` + tail,
		long + `{"name":"a\u12g4"}` + tail,
		long + `{,}` + tail,
		long + `[}` + tail,
		long + `{"a":1,}` + tail,
	}
	targets := []struct {
		name string
		new  func(std bool) any
	}{
		{"struct", func(bool) any { return new(struct{ Name string }) }},
		{"any", func(bool) any { return new(any) }},
		{"map", func(bool) any { return new(map[string]any) }},
		{"slice", func(bool) any { return new([]struct{ Name string }) }},
		{"raw message", func(std bool) any {
			if std {
				return new(stdjson.RawMessage)
			}
			return new(json.RawMessage)
		}},
		{"unmarshaler", func(bool) any { return new(readsUnmarshaler) }},
		{"another kind", func(bool) any { return new(int) }},
	}
	for _, in := range inputs {
		for _, target := range targets {
			std := &oneByteReader{data: in}
			stdErr := decodeTwo(stdjson.NewDecoder(std).Decode, func() any { return target.new(true) })
			got := &oneByteReader{data: in}
			err := decodeTwo(json.NewDecoder(got).Decode, func() any { return target.new(false) })
			if (err == nil) != (stdErr == nil) {
				t.Errorf("%.30q into %s: got %v, encoding/json %v", in, target.name, err, stdErr)
				continue
			}
			// the bytes read byte by byte are the ones encoding/json reads
			if got.n > std.n {
				t.Errorf("%.30q into %s: %d bytes read, encoding/json %d", in, target.name, got.n, std.n)
			}
			// by reads of more than a byte, the read which brings the byte is the last one
			chunked := &chunkedReader{data: in, size: 4096}
			if err := decodeTwo(json.NewDecoder(chunked).Decode, func() any { return target.new(false) }); (err == nil) != (stdErr == nil) || chunked.n > std.n+4096 {
				t.Errorf("%.30q into %s by 4096 bytes: %v after %d bytes, encoding/json %v at %d", in, target.name, err, chunked.n, stdErr, std.n)
			}
			if stdErr != nil {
				decodePipeFails(t, in[:std.n], func() any { return target.new(false) })
			}
		}
	}
}

// watchdog runs f, and fails the test with the stacks of the goroutines if f doesn't end within a while: a decoding
// which goes round without a read fails with where it does, instead of the test hanging.
func watchdog(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		buf := make([]byte, 1<<20)
		t.Fatalf("%s doesn't end:\n%s", what, buf[:runtime.Stack(buf, true)])
	}
}

// decodePipeFails checks that the input, from a reader which is not closed after it, fails without a wait for more.
func decodePipeFails(t *testing.T, in string, newValue func() any) {
	t.Helper()
	r, w := io.Pipe()
	defer w.Close()
	go func() {
		_, _ = w.Write([]byte(in))
		// the writer neither writes nor closes more: a read more waits for ever
	}()
	done := make(chan error, 1)
	go func() { done <- decodeTwo(json.NewDecoder(r).Decode, newValue) }()
	select {
	case err := <-done:
		if err == nil {
			t.Errorf("%.30q from a pipe: no error", in)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("%.30q from a pipe: Decode waits for input after the byte which is not valid", in)
	}
}

// decodeTwo decodes two values by decode, going on after a type error, as encoding/json does, and returns the error.
func decodeTwo(decode func(any) error, newValue func() any) error {
	var err error
	for i := 0; i < 2; i++ {
		var typeErr *json.UnmarshalTypeError
		var stdTypeErr *stdjson.UnmarshalTypeError
		if err = decode(newValue()); err != nil && !errors.As(err, &typeErr) && !errors.As(err, &stdTypeErr) {
			return err
		}
	}
	return err
}

// A value is decoded once its last byte is read: a reader which has no more yet is not waited for. A number at the top
// is ended by the byte after it, as encoding/json ends it.
func TestDecodeStreamNoWaitAfterValue(t *testing.T) {
	pad := strings.Repeat(" ", 70000)
	for _, in := range []string{
		`{"a":1}`,
		`{"a":[` + pad + `1,2],"b":{"c":"d"}}`,
		`[1,"x",{"y":null}]`,
		`[` + pad + `{"a":1},{"b":2}]`,
		`"string"`,
		`true`, `false`, `null`,
		`12 `,
	} {
		for _, newValue := range []func() any{
			func() any { return new(any) },
			func() any { return new(json.RawMessage) },
		} {
			r, w := io.Pipe()
			go func() {
				_, _ = w.Write([]byte(in))
				// the writer neither writes nor closes more: a read more waits for ever
			}()
			done := make(chan error, 1)
			go func() { done <- json.NewDecoder(r).Decode(newValue()) }()
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("%.20q: %v", in, err)
				}
			case <-time.After(10 * time.Second):
				t.Fatalf("%.20q: Decode waits for input after the value", in)
			}
			_ = w.Close()
		}
	}
	// the tokens of an array whose rest is not written yet
	r, w := io.Pipe()
	go func() { _, _ = w.Write([]byte(`[{"a":1}`)) }()
	done := make(chan error, 1)
	go func() {
		dec := json.NewDecoder(r)
		if _, err := dec.Token(); err != nil {
			done <- err
			return
		}
		done <- dec.Decode(new(any))
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Token or Decode waits for input after the token")
	}
	_ = w.Close()
}

// chunkedReader reads data by chunks of at most size bytes, and counts the bytes read, and where each read ended.
type chunkedReader struct {
	data string
	size int
	n    int
	ends []int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	if r.n == len(r.data) {
		return 0, io.EOF
	}
	n := copy(p[:min(len(p), r.size)], r.data[r.n:])
	r.n += n
	r.ends = append(r.ends, r.n)
	return n, nil
}

// readAfter reports whether a read followed the one which brought the byte at pos.
func (r *chunkedReader) readAfter(pos int) bool {
	for i, end := range r.ends {
		if end > pos {
			return i < len(r.ends)-1
		}
	}
	return false
}

// A control character of a string at the top fails within the read which brings it, wherever the escapes before it
// are in that read.
func TestDecodeStreamTopStringControlBeforeEscape(t *testing.T) {
	rest := strings.Repeat("z", 1<<20)
	for _, in := range []string{"\"ab\n\\\\cd" + rest + "\"", "\"ab\n\\\\cd" + rest, "\"\\\\\n\\\\" + rest} {
		for _, size := range []int{1, 16, 4096} {
			r := &chunkedReader{data: in, size: size}
			var v string
			if err := json.NewDecoder(r).Decode(&v); err == nil {
				t.Fatalf("%.10q by %d: no error", in, size)
			}
			if r.n > 4+size {
				t.Fatalf("%.10q by %d: %d bytes read", in, size, r.n)
			}
			// Token reads a string so too
			r = &chunkedReader{data: in, size: size}
			if _, err := json.NewDecoder(r).Token(); err == nil || r.n > 4+size {
				t.Fatalf("%.10q by %d: Token %v after %d bytes", in, size, err, r.n)
			}
		}
	}
}

// A control character in a string, in a \u escape too, is a syntax error once it is read: the bytes after it are not
// waited for.
func TestDecodeStreamControlInEscape(t *testing.T) {
	for _, in := range []string{"[\"\\u0\n", "{\"a\":\"\\u12\x01", "\"\\u\x01", "{\"\\u1\n", "\"\\\\u1\n", "\"\\u\\\n"} {
		r, w := io.Pipe()
		go func() {
			_, _ = w.Write([]byte(in))
			// the writer neither writes nor closes more: a read more waits for ever
		}()
		done := make(chan error, 1)
		go func() { done <- json.NewDecoder(r).Decode(new(any)) }()
		select {
		case err := <-done:
			var syntaxErr *json.SyntaxError
			if !errors.As(err, &syntaxErr) {
				t.Errorf("%q: got %v, want a syntax error", in, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("%q: Decode waits for input after the control character", in)
		}
		_ = w.Close()
	}
}

// A value where Token waits for the name of a member is read on only while it is valid: a control character in a
// string of it fails it, without a read more.
func TestDecodeStreamMemberNameControl(t *testing.T) {
	for _, in := range []string{"{[\"abc\n, 1, 2", "{{\"a\":\"b\x01"} {
		r, w := io.Pipe()
		go func() {
			_, _ = w.Write([]byte(in))
			// the writer neither writes nor closes more: a read more waits for ever
		}()
		done := make(chan error, 1)
		go func() {
			dec := json.NewDecoder(r)
			if _, err := dec.Token(); err != nil {
				done <- err
				return
			}
			done <- dec.Decode(new(any))
		}()
		select {
		case err := <-done:
			if !errors.As(err, new(*json.SyntaxError)) {
				t.Errorf("%q: got %v, want a syntax error", in, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("%q: Decode waits for input after the control character", in)
		}
		_ = w.Close()
	}
}

// The comma or the colon before a value which Decode reads after Token, followed by more white space than the buffer
// holds, is decoded as encoding/json decodes it, wherever the reads cut the input.
func TestDecodeStreamDelimiterBeforeLongSpace(t *testing.T) {
	space := strings.Repeat("\n", 600)
	for _, c := range []struct {
		in     string
		tokens int
	}{{"[:" + space + "]", 1}, {"[," + space, 1}, {"[1," + space, 2}} {
		in := c.in
		for _, size := range []int{1, 64, 512, 4096} {
			sd := stdjson.NewDecoder(&chunkedReader{data: in, size: size})
			gd := json.NewDecoder(&chunkedReader{data: in, size: size})
			for i := 0; i < c.tokens; i++ {
				if _, err := sd.Token(); err != nil {
					t.Fatal(err)
				}
				if _, err := gd.Token(); err != nil {
					t.Fatal(err)
				}
			}
			var want, got any
			stdErr := sd.Decode(&want)
			err := gd.Decode(&got)
			if errors.As(stdErr, new(*stdjson.SyntaxError)) != errors.As(err, new(*json.SyntaxError)) || (stdErr == nil) != (err == nil) {
				t.Fatalf("%.10q by %d: got %v, encoding/json %v", in, size, err, stdErr)
			}
		}
	}
}

type skippedMembers struct {
	Name string
	A    [1]any
}

// The values which nothing decodes, of every kind, are skipped as encoding/json skips them, whatever the reads cut, and
// the ones which are not valid fail, wherever they are.
func TestDecodeStreamSkipUnknownMembers(t *testing.T) {
	pad := `"pad":"` + strings.Repeat("p", 700) + `"`
	members := `"o":{"a":[1,{"b":"c"},null],"d":{}},"l":[[],{},[1,2,[3]],"x\"]",{"k":"v}"}],"s":"str]}","n":-1.5e3,` +
		`"t":true,"z":null,"e":{},"f":[],"A":[{"x":[1]},[2,{"y":3}],"s"]`
	valid := "{" + pad + "," + members + `,"Name":"x"}`
	inputs := []string{
		valid,
		strings.Replace(valid, `"b":"c"`, `"b" "c"`, 1),
		strings.Replace(valid, `[3]`, `[3}`, 1),
		strings.Replace(valid, `"k":"v}"`, `"k":v`, 1),
		strings.Replace(valid, `{"y":3}`, `{"y":3,}`, 1),
		valid[:len(valid)-1] + `,"late":{"x":[1,2,}]}}`,
		// deeper than the stream checks before a read, and too deep
		"{" + pad + "," + members + `,"deep":` + strings.Repeat("[", 100) + strings.Repeat("]", 100) + `,"Name":"x"}`,
		"{" + pad + `,"deep":` + strings.Repeat("[", 10001) + strings.Repeat("]", 10001) + `,"Name":"x"}`,
	}
	for _, in := range inputs {
		var want skippedMembers
		wantErr := stdjson.Unmarshal([]byte(in), &want)
		for _, size := range []int{1, 2, 3, 5, 7, 16, 64, 100, 300, 512, 1000, 4096} {
			for _, opts := range [][]json.DecodeOptionFunc{nil, {json.DecodeFieldPriorityFirstWin()}} {
				var got skippedMembers
				err := json.NewDecoder(&chunkedReader{data: in, size: size}).DecodeWithOption(&got, opts...)
				if (err == nil) != (wantErr == nil) {
					t.Fatalf("%.40q by %d: got %v, encoding/json %v", in, size, err, wantErr)
				}
				if err == nil && (got.Name != want.Name || fmt.Sprint(got.A) != fmt.Sprint(want.A)) {
					t.Fatalf("%.40q by %d: got %+v, want %+v", in, size, got, want)
				}
			}
		}
	}
}

// A random value, valid or made not valid by a change of a byte, after more than a read of valid bytes, read by small
// reads which cut it anywhere, into any kind of target, fails as encoding/json fails, at the read of the same byte,
// or is decoded as it.
func TestDecodeStreamGrammarAgainstStd(t *testing.T) {
	targets := []func(std bool) any{
		func(bool) any { return new(any) },
		func(std bool) any {
			if std {
				return new(stdjson.RawMessage)
			}
			return new(json.RawMessage)
		},
		func(bool) any { return new(struct{ V any }) },
		func(bool) any { return new(readsUnmarshaler) },
	}
	r := rand.New(rand.NewSource(3))
	pad := `"` + strings.Repeat("x", 600) + `"`
	for i := 0; i < 3000; i++ {
		v := randomSkippedValue(r, 1+r.Intn(4))
		if r.Intn(2) == 0 {
			v = mutate(r, v)
		}
		in := `{"pad":` + pad + `,"v":` + v + `}` + " " + `{"next":1}`
		if r.Intn(4) == 0 {
			in = v + " " + `{"next":1}`
		}
		size := []int{1, 64, 512, 4096}[r.Intn(4)]
		target := targets[r.Intn(len(targets))]
		std := &chunkedReader{data: in, size: size}
		sd := stdjson.NewDecoder(std)
		got := &chunkedReader{data: in, size: size}
		gd := json.NewDecoder(got)
		for n := 0; n < 2; n++ {
			want, value := target(true), target(false)
			stdErr := sd.Decode(want)
			var err error
			watchdog(t, fmt.Sprintf("%.200q by %d: value %d", in, size, n), func() { err = gd.Decode(value) })
			if errors.As(stdErr, new(*stdjson.SyntaxError)) != errors.As(err, new(*json.SyntaxError)) || (stdErr == nil) != (err == nil) {
				t.Fatalf("%q by %d: value %d: got %v, encoding/json %v", in, size, n, err, stdErr)
			}
			if stdErr != nil {
				// the read which brought the byte which encoding/json rejects is the last one: the reads are not the
				// ones of encoding/json, which has a buffer of its own
				var syntax *stdjson.SyntaxError
				if errors.As(stdErr, &syntax) && got.readAfter(int(syntax.Offset)-1) {
					t.Fatalf("%q by %d: value %d: %d bytes read, a read after the one of the byte at %d", in, size, n, got.n, syntax.Offset-1)
				}
				break
			}
			wantJSON, _ := stdjson.Marshal(want)
			gotJSON, _ := stdjson.Marshal(value)
			if !bytes.Equal(gotJSON, wantJSON) {
				t.Fatalf("%q by %d: value %d: got %s, want %s", in, size, n, gotJSON, wantJSON)
			}
		}
	}
}

// describeError returns what the depth tests compare of an error of go-json or of encoding/json: its kind, its offset,
// from base for a syntax error, and its message.
func describeError(err error, base int64) string {
	var syntax *json.SyntaxError
	var stdSyntax *stdjson.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var stdTypeErr *stdjson.UnmarshalTypeError
	switch {
	case err == nil:
		return "no error"
	case errors.As(err, &syntax):
		return fmt.Sprintf("syntax error at %d: %s", base+syntax.Offset, err)
	case errors.As(err, &stdSyntax):
		return fmt.Sprintf("syntax error at %d: %s", base+stdSyntax.Offset, err)
	case errors.As(err, &typeErr):
		return fmt.Sprintf("type error at %d: %s", typeErr.Offset, err)
	case errors.As(err, &stdTypeErr):
		return fmt.Sprintf("type error at %d: %s", stdTypeErr.Offset, err)
	}
	return err.Error()
}

// mixedNesting returns a number nested in n arrays and objects, which are arrays and objects in turn.
func mixedNesting(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			b.WriteString("[")
		} else {
			b.WriteString(`{"a":`)
		}
	}
	b.WriteString("1")
	for i := n - 1; i >= 0; i-- {
		if i%2 == 0 {
			b.WriteString("]")
		} else {
			b.WriteString("}")
		}
	}
	return b.String()
}

// A value nested deeper than the limit fails at the bracket past it, with the offset and the message of Unmarshal and
// of encoding/json, wherever the reads of the input end: a Decoder decodes the value whole when the reads have brought
// it, and else element by element, and skips a value which its target doesn't decode. A value nested as deep as the
// limit is decoded as Unmarshal and encoding/json decode it. The value may follow another value, or be in arrays which
// Token opened.
func TestDecodeStreamDepthErrorsByReads(t *testing.T) {
	type chain struct {
		A *chain `json:"a"`
	}
	type field struct {
		A any `json:"a"`
	}
	type otherField struct {
		B int `json:"b"`
	}
	targets := []struct {
		name string
		new  func() any
	}{
		{"any", func() any { return new(any) }},
		{"[]any", func() any { return new([]any) }},
		{"[][]any", func() any { return new([][]any) }},
		{"map[string]any", func() any { return new(map[string]any) }},
		{"struct{A any}", func() any { return new(field) }},
		{"recursive struct", func() any { return new(chain) }},
		{"struct of an unknown field", func() any { return new(otherField) }},
	}
	const (
		afterValue = "null "
		inTokens   = "[[["
	)
	type depthCase struct {
		name  string
		value string
		// before is what the input has before the value: a value decoded first, or arrays which Token opens
		before string
	}
	var cases []depthCase
	for _, n := range []int{9999, 10000, 10001, 10002} {
		arrays := strings.Repeat("[", n) + strings.Repeat("]", n)
		cases = append(cases,
			depthCase{fmt.Sprintf("arrays %d", n), arrays, ""},
			// null in the objects, which the recursive struct decodes too ( TestDecodeValueAtMaxDepth has the type
			// errors at the depth limit )
			depthCase{fmt.Sprintf("objects %d", n), strings.Repeat(`{"a":`, n) + "null" + strings.Repeat("}", n), ""},
			depthCase{fmt.Sprintf("arrays and objects %d", n), mixedNesting(n), ""},
			depthCase{fmt.Sprintf("arrays %d in an object", n), `{"a":` + arrays + `}`, ""},
			depthCase{fmt.Sprintf("arrays %d in an array before a string with a newline", n), "[" + arrays + `,"x` + "\n" + `"]`, ""},
			depthCase{fmt.Sprintf("arrays %d not closed", n), strings.Repeat("[", n), ""},
			depthCase{fmt.Sprintf("arrays %d after a value", n), arrays, afterValue},
			depthCase{fmt.Sprintf("arrays %d in arrays opened by Token", n), arrays, inTokens},
		)
	}
	// decode decodes the value of the input of dec into a new value by newValue, after what is before the value, and
	// returns the error, which is the one of what is before the value if it fails.
	decode := func(dec interface {
		Decode(any) error
		Token() (stdjson.Token, error)
		More() bool
	}, before string, newValue func() any) error {
		switch before {
		case afterValue:
			if err := dec.Decode(newValue()); err != nil {
				return fmt.Errorf("the value before: %w", err)
			}
		case inTokens:
			for range len(inTokens) {
				if _, err := dec.Token(); err != nil {
					return fmt.Errorf("Token: %w", err)
				}
			}
			if !dec.More() {
				return errors.New("More is false")
			}
		}
		return dec.Decode(newValue())
	}
	reads := []int{1, 2, 3, 7, 13, 64, 65, 511, 512, 513, 1 << 20}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			in := c.before + c.value
			if c.before == inTokens {
				in += "]]]"
			}
			base := int64(len(c.before))
			for _, target := range targets {
				want := unmarshalReference([]byte(c.value), target.new(), false, false)
				if json.IsAtEnd(want) {
					want = io.ErrUnexpectedEOF
				}
				// the error is the one of Unmarshal, but for the offset of a type error of a value which is not the
				// first of the input ( see typeErrorsFromValueStart ) and the one of an error after Token, which
				// encoding/json before Go 1.27 counts without the bytes which Token read
				var typeErr *json.UnmarshalTypeError
				sameAsUnmarshal := c.before != inTokens && (base == 0 || !errors.As(want, &typeErr))
				// encoding/json reads a value up to its end, its first syntax error or the end of the input before it
				// decodes it: its error is the same whatever its reads.
				// encoding/json of Go 1.27 counts the arrays which Token opened in the depth of the value.
				std := describeError(decode(stdjson.NewDecoder(strings.NewReader(in)), c.before, target.new), 0)
				sameAsStd := c.before != inTokens || !tokenDepthCounted
				var first string
				for _, size := range reads {
					got := describeError(decode(json.NewDecoder(&chunkReader{data: []byte(in), chunk: size}), c.before, target.new), 0)
					if first == "" {
						first = got
					} else if got != first {
						t.Fatalf("into %s by %d: got %.300s, by other reads %.300s", target.name, size, got, first)
					}
					if sameAsUnmarshal && got != describeError(want, base) {
						t.Fatalf("into %s by %d: got %.300s, Unmarshal %.300s", target.name, size, got, describeError(want, base))
					}
					if sameAsStd && got != std {
						t.Fatalf("into %s by %d: got %.300s, encoding/json %.300s", target.name, size, got, std)
					}
				}
			}
		})
	}
}
