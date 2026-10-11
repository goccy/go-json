package json_test

import (
	stdjson "encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	json "github.com/goccy/go-json"
)

// A Decoder reads no more than the value it decodes needs: a value which has a control character in a string fails at
// the read of that character, as encoding/json, which reads a value while it checks it, fails there, and a value is
// decoded once its last byte is read, without a read more.

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
	// a string which is not closed fails at the control character which follows it, as the line break after it ( issue
	// 701 ): the scan of a value stops there. Any other error of the grammar is found once the brackets of the value
	// balance, or at the end of the input.
	inputs := []string{
		`{"name":"alice` + "\n" + tail,
		`{"name":"alice` + "\t" + flat,
		`[{"name":"alice` + "\n" + tail,
		rec + "\n" + `{"name":"bob` + "\n" + tail,
		`{"name":"a\` + "\n" + tail,
		`"unterminated` + "\n" + tail,
		`{"a":{"b":[{"c":"x` + "\n" + tail,
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
			// a character cut by the read is read whole for the error
			if got.n > std.n+4 {
				t.Errorf("%.30q into %s: %d bytes read, encoding/json %d", in, target.name, got.n, std.n)
			}
		}
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

// chunkedReader reads data by chunks of at most size bytes, and counts the bytes read.
type chunkedReader struct {
	data string
	size int
	n    int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	if r.n == len(r.data) {
		return 0, io.EOF
	}
	n := copy(p[:min(len(p), r.size)], r.data[r.n:])
	r.n += n
	return n, nil
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

// A control character in a \u escape is the error which the input read at once has, which shows the bytes of the
// escape after it.
func TestDecodeStreamControlInEscape(t *testing.T) {
	for _, in := range []string{"[\"\\u0\n0e9\"]", "{\"a\":\"\\u12\x0134\"}", "\"\\u\x01000\"", "{\"\\u1\n23\":1}", "[\"\\u0\n"} {
		want := json.Unmarshal([]byte(in), new(any))
		for _, size := range []int{1, 2, 4096} {
			got := json.NewDecoder(&chunkedReader{data: in, size: size}).Decode(new(any))
			if want == nil || got == nil || got.Error() != want.Error() {
				t.Fatalf("%q by %d: got %v, want %v", in, size, got, want)
			}
		}
	}
}
