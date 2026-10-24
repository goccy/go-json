//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

// The tests here compare UnmarshalRead with the one of encoding/json/v2, for the inputs of the tests of Unmarshal,
// read by readers of a few sizes of reads: a value is decoded whole or element by element as the reads cut it.

// readSizes are the sizes of the reads of the readers of the inputs: one byte, a few, the size of the reads of a
// stream, and 0 for the whole input at once.
var readSizes = []int{1, 7, 64, 512, 0}

// chunkReader reads its input n bytes at a time at most, or whole if n is 0.
type chunkReader struct {
	in []byte
	n  int
}

func newChunkReader(in string, n int) *chunkReader {
	return &chunkReader{in: []byte(in), n: n}
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.in) == 0 {
		return 0, io.EOF
	}
	n := len(p)
	if r.n > 0 {
		n = min(n, r.n)
	}
	n = copy(p, r.in[:min(n, len(r.in))])
	r.in = r.in[n:]
	return n, nil
}

// readTextByReads are the cases of an escape of a surrogate which is not a pair, whose error both packages describe
// by the bytes of the input after the escape which they have read when they find it: they depend on the sizes of the
// reads, and in go-json on the size of the buffer of a stream, which is reused. The error is compared without its
// text, which go-json doesn't imitate.
var readTextByReads = map[string]bool{
	"string of a lone surrogate":                     true,
	"string of a lone low surrogate":                 true,
	"string of a high surrogate before a letter":     true,
	"any name of a lone surrogate":                   true,
	"struct unknown value of a lone surrogate":       true,
	"string of an escape and a lone surrogate":       true,
	"string of a lone surrogate at the end":          true,
	"string of a high surrogate before a bad escape": true,
	"string long of a lone surrogate at the end":     true,
	"time of a lone surrogate":                       true,
	"text of a lone surrogate":                       true,
}

// describeReadError is describeUnmarshalError for the case of the name, without the text of the error if the text
// of the error of encoding/json/v2 depends on the size of its reads.
func describeReadError(name string, err error) string {
	if !readTextByReads[name] {
		return describeUnmarshalError(err)
	}
	var se *jsontext.SyntacticError
	var stdSe *stdjsontext.SyntacticError
	switch {
	case errors.As(err, &se):
		return fmt.Sprintf("SyntacticError{ByteOffset: %d, JSONPointer: %q}", se.ByteOffset, se.JSONPointer)
	case errors.As(err, &stdSe):
		return fmt.Sprintf("SyntacticError{ByteOffset: %d, JSONPointer: %q}", stdSe.ByteOffset, stdSe.JSONPointer)
	}
	return describeUnmarshalError(err)
}

// readDiff runs the test of each size of the reads of the input in. With the race detector, an input longer than
// 64 KiB, as a deep input spaced out, is not read a byte at a time, which takes the race detector seconds for each
// test: its reads of 7 bytes cut its tokens as well, and the tests without the race detector read it a byte at a time.
func readDiff(t *testing.T, name, in string, f func(t *testing.T, n int)) {
	for _, n := range readSizes {
		if n == 1 && raceEnabled && len(in) > 64<<10 {
			continue
		}
		t.Run(fmt.Sprintf("%s/reads-%d", name, n), func(t *testing.T) { f(t, n) })
	}
}

// spacedInputs are the inputs of a case which the tests read: the input, and the input spaced out ( see spaced ).
func spacedInputs(in string) []string {
	return []string{in, spaced(in, 300)}
}

// spaced returns in with n spaces before and after each delimiter and bracket out of its strings, which it may have
// only between its tokens: the input is longer than the buffer of a stream, which reads drop the bytes of and move,
// between the tokens and in them, and the places of its errors are the ones of encoding/json/v2 in it all the same.
func spaced(in string, n int) string {
	pad := strings.Repeat(" ", n)
	var b strings.Builder
	inString, escaped := false, false
	for i := 0; i < len(in); i++ {
		c := in[i]
		switch {
		case inString:
			b.WriteByte(c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		case c == '"':
			inString = true
		case c == '{' || c == '}' || c == '[' || c == ']' || c == ',' || c == ':':
			b.WriteString(pad)
			b.WriteByte(c)
			b.WriteString(pad)
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}

func TestUnmarshalReadMatchesStd(t *testing.T) {
	for _, c := range append(unmarshalDiffCases(), unmarshalDiffStructCases()...) {
		for i, in := range spacedInputs(c.in) {
			readDiff(t, fmt.Sprintf("%s/%d", c.name, i), in, func(t *testing.T, n int) {
				ours, theirs := c.out(), c.out()
				err := json.UnmarshalRead(newChunkReader(in, n), ours, c.opts.ours()...)
				stdErr := stdjson.UnmarshalRead(newChunkReader(in, n), theirs, c.opts.std()...)
				if got, want := describeReadError(c.name, err), describeReadError(c.name, stdErr); got != want {
					t.Fatalf("error:\n got %s\nwant %s", got, want)
				}
				if err == nil && !c.errorOnly && !reflect.DeepEqual(ours, theirs) {
					t.Fatalf("value:\n got %#v\nwant %#v", reflect.ValueOf(ours).Elem(), reflect.ValueOf(theirs).Elem())
				}
			})
		}
	}
}

func TestUnmarshalReadMethodsMatchStd(t *testing.T) {
	for _, c := range umCases() {
		for i, in := range spacedInputs(c.in) {
			readDiff(t, fmt.Sprintf("%s/%d", c.name, i), in, func(t *testing.T, n int) {
				ours, theirs := c.out(), c.out()
				err := json.UnmarshalRead(newChunkReader(in, n), ours, c.opts.ours...)
				stdErr := stdjson.UnmarshalRead(newChunkReader(in, n), theirs, c.opts.std...)
				if got, want := describeUnmarshalError(err), describeUnmarshalError(stdErr); got != want {
					t.Fatalf("error:\n got %s\nwant %s", got, want)
				}
				if err == nil && !reflect.DeepEqual(ours, theirs) {
					t.Fatalf("value:\n got %#v\nwant %#v", reflect.ValueOf(ours).Elem(), reflect.ValueOf(theirs).Elem())
				}
			})
		}
	}
}

func TestUnmarshalReadFromMatchesStd(t *testing.T) {
	for _, c := range ufCases() {
		for i, in := range spacedInputs(c.in) {
			readDiff(t, fmt.Sprintf("%s/%d", c.name, i), in, func(t *testing.T, n int) {
				ours, theirs := c.ours(), c.theirs()
				err := json.UnmarshalRead(newChunkReader(in, n), ours, c.opts.ours...)
				stdErr := stdjson.UnmarshalRead(newChunkReader(in, n), theirs, c.opts.std...)
				want := strings.ReplaceAll(describeUnmarshalError(stdErr), "ufStd", "ufOurs")
				if got := describeUnmarshalError(err); got != want {
					t.Fatalf("error:\n got %s\nwant %s", got, want)
				}
				if got, want := ufText(ours), ufText(theirs); err == nil && got != want {
					t.Fatalf("value:\n got %s\nwant %s", got, want)
				}
			})
		}
	}
}

// errReaderFailed is the error of the reader of TestUnmarshalReadReaderErrorMatchesStd.
var errReaderFailed = errors.New("the reader failed")

// The error of the reader is reported as encoding/json/v2 reports it, wherever in the input the reader fails.
func TestUnmarshalReadReaderErrorMatchesStd(t *testing.T) {
	for _, in := range []string{``, ` `, `{`, `{"a"`, `{"a":`, `{"a":[1,2`, `{"a":"bc`, `[{"a":1},`, `{"a":1}`, `123`} {
		for _, n := range readSizes {
			newReader := func() io.Reader {
				return io.MultiReader(newChunkReader(in, n), iotest.ErrReader(errReaderFailed))
			}
			var ours, theirs struct{ A any }
			err := json.UnmarshalRead(newReader(), &ours)
			stdErr := stdjson.UnmarshalRead(newReader(), &theirs)
			if got, want := describeUnmarshalError(err), describeUnmarshalError(stdErr); got != want {
				t.Fatalf("%q, reads of %d: error:\n got %s\nwant %s", in, n, got, want)
			}
			if got, want := errors.Is(err, errReaderFailed), errors.Is(stdErr, errReaderFailed); got != want {
				t.Fatalf("%q, reads of %d: the error of the reader is in the error: got %v, want %v", in, n, got, want)
			}
		}
	}
}
