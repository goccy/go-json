package json_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/goccy/go-json"
)

// A Decoder reports the syntax error of a value as soon as it has read its byte, as encoding/json does, and
// doesn't wait for more input from a reader which has no more yet.
func TestDecoderDoesNotWaitAfterSyntaxError(t *testing.T) {
	for _, in := range []string{`{"a": x`, `[1, x`, `{"a" 1`, `[1 2`, `["a\q`, `{"a":[01`, "[\"\x01"} {
		r, w := io.Pipe()
		go func() { _, _ = w.Write([]byte(in)) }()
		done := make(chan error, 1)
		go func() {
			var v any
			done <- json.NewDecoder(r).Decode(&v)
		}()
		select {
		case err := <-done:
			var serr *json.SyntaxError
			if !errors.As(err, &serr) {
				t.Errorf("%q: got %v, want a syntax error", in, err)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("%q: the decoder waits for more input", in)
		}
		_ = w.Close()
	}
}

// Values whose bytes come one by one, which the decoder of jsontext reads to their ends, are decoded as they
// are from a buffer.
func TestDecoderReadsValuesByParts(t *testing.T) {
	inputs := []string{
		`{"a":[1,2,{"b":"c\"d\\eé"}],"f":null,"g":true} [1] "s" 2 {}`,
		strings.Repeat(`{"key":"`+strings.Repeat("v", 300)+`","n":[1.5,-2e3]}`, 20),
		`"` + strings.Repeat("\\n\\u00e9x", 200) + `"`,
	}
	for _, in := range inputs {
		decodeAll := func(r io.Reader) ([]any, error) {
			dec := json.NewDecoder(r)
			var out []any
			for {
				var v any
				if err := dec.Decode(&v); err == io.EOF {
					return out, nil
				} else if err != nil {
					return out, err
				}
				out = append(out, v)
			}
		}
		want, werr := decodeAll(bytes.NewReader([]byte(in)))
		got, gerr := decodeAll(iotest.OneByteReader(strings.NewReader(in)))
		if fmt.Sprint(gerr) != fmt.Sprint(werr) || !reflect.DeepEqual(got, want) {
			t.Fatalf("%.60q: got %v, %v; want %v, %v", in, got, gerr, want, werr)
		}
		// the tokens too
		dec := json.NewDecoder(iotest.OneByteReader(strings.NewReader(in)))
		ref := json.NewDecoder(strings.NewReader(in))
		for {
			gt, gerr := dec.Token()
			wt, werr := ref.Token()
			if fmt.Sprint(gt) != fmt.Sprint(wt) || fmt.Sprint(gerr) != fmt.Sprint(werr) {
				t.Fatalf("%.60q: token %v, %v; want %v, %v", in, gt, gerr, wt, werr)
			}
			if werr != nil {
				break
			}
		}
	}
}
