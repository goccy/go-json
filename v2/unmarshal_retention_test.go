//go:build go1.24

package json_test

import (
	"io"
	"runtime"
	"strings"
	"testing"
	"unsafe"
	"weak"

	json "github.com/goccy/go-json/v2"
	"github.com/goccy/go-json/jsontext"
)

// A decoded value refers to memory of its own only, as with encoding/json/v2: a value kept out of a call keeps alive
// neither the other values of the call, nor the buffers of the decoding.

// aliveRef reports whether the object which it was made for is alive, without keeping it alive.
type aliveRef func() bool

// refOf returns the aliveRef of the object which p points into.
func refOf(p *byte) aliveRef {
	w := weak.Make(p)
	return func() bool { return w.Value() != nil }
}

// recordingReader is a reader which keeps an aliveRef of every buffer it reads into.
type recordingReader struct {
	r    io.Reader
	bufs []aliveRef
}

func (r *recordingReader) Read(p []byte) (int, error) {
	if len(p) != 0 {
		r.bufs = append(r.bufs, refOf(&p[0]))
	}
	return r.r.Read(p)
}

type retentionStrings struct {
	A, B, C string
	Pad     string
}

func TestDecodedValueKeepsOnlyItsOwnMemory(t *testing.T) {
	// strings longer than the objects of the tiny allocator of the runtime, which packs objects of less than 16 bytes
	// without pointers together: such a string keeps its neighbors alive, whatever allocated it
	a, b := strings.Repeat("a", 32), strings.Repeat("b", 32)
	escaped := strings.Repeat(`c\n`, 16)
	object := `{"A":"` + a + `","B":"` + b + `","C":"` + escaped + `"}`
	// an object which the read buffer doesn't hold whole, with a token longer than the buffer
	long := `{"A":"` + a + `","B":"` + b + `","C":"` + escaped + `","Pad":"` + strings.Repeat("p", 2000) + `"}`
	read := func(input string, decode func(io.Reader, *retentionStrings) error) func(*testing.T) (*retentionStrings, []aliveRef) {
		return func(t *testing.T) (*retentionStrings, []aliveRef) {
			v := new(retentionStrings)
			r := &recordingReader{r: strings.NewReader(input)}
			if err := decode(r, v); err != nil {
				t.Fatal(err)
			}
			// the buffers of the decoding, which are dropped
			return v, r.bufs
		}
	}
	unmarshalRead := func(r io.Reader, v *retentionStrings) error {
		return json.UnmarshalRead(r, v)
	}
	unmarshalDecode := func(r io.Reader, v *retentionStrings) error {
		return json.UnmarshalDecode(jsontext.NewDecoder(r), v)
	}
	for _, tc := range []struct {
		name   string
		decode func(*testing.T) (*retentionStrings, []aliveRef)
	}{
		{"Unmarshal", func(t *testing.T) (*retentionStrings, []aliveRef) {
			v := new(retentionStrings)
			if err := json.Unmarshal([]byte(object), v); err != nil {
				t.Fatal(err)
			}
			return v, nil
		}},
		{"UnmarshalRead", read(object, unmarshalRead)},
		{"UnmarshalRead element by element", read(long, unmarshalRead)},
		{"UnmarshalDecode", read(object, unmarshalDecode)},
		{"UnmarshalDecode element by element", read(long, unmarshalDecode)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, bufs := tc.decode(t)
			kept := v.B
			values := []aliveRef{refOf(unsafe.StringData(v.A)), refOf(unsafe.StringData(v.C))}
			if v.Pad != "" {
				values = append(values, refOf(unsafe.StringData(v.Pad)))
			}
			v = nil
			// The pools keep their objects over one collection ( their victim caches ), so that a pooled object which
			// refers to a decoded value keeps it alive; and over two the read buffer of a pooled stream, which is
			// reused by the next call, and which nothing decoded refers to.
			runtime.GC()
			for i, alive := range values {
				if alive() {
					t.Fatalf("a value which the kept value is not made of is alive ( %d of %d )", i, len(values))
				}
			}
			runtime.GC()
			for i, alive := range bufs {
				if alive() {
					t.Fatalf("a buffer of the decoding is alive ( %d of %d )", i, len(bufs))
				}
			}
			if kept != b {
				t.Fatalf("got %q", kept)
			}
		})
	}
}
