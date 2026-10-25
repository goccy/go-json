package jsontext

import (
	"strconv"
	"strings"
	"testing"
)

// TestPutEncoderDropsLargeArrays checks that an encoder which formatted a large value goes back to the pool
// without its large arrays, which would be kept for the small values after it.
func TestPutEncoderDropsLargeArrays(t *testing.T) {
	v := []byte{'{'}
	for i := 100000; i > 0; i-- {
		if i < 100000 {
			v = append(v, ',')
		}
		v = strconv.AppendQuote(v, "name"+strconv.Itoa(i))
		v = append(v, ":1"...)
	}
	v = append(v, '}')
	e := getEncoder(nil, canonicalOptions...)
	if err := e.writeValue(v); err != nil {
		t.Fatal(err)
	}
	e.buf = e.buf[:0] // a buffer which is not used well is dropped already
	putEncoder(e)
	for name, n := range map[string]int{
		"buf":           cap(e.buf),
		"ro.buf":        cap(e.vs.ro.buf),
		"ro.names":      cap(e.vs.ro.names),
		"ro.members":    cap(e.vs.ro.members),
		"ro.sorted":     cap(e.vs.ro.sorted),
		"st.names.buf":  cap(e.st.names.buf),
		"st.names.ends": cap(e.st.names.ends),
	} {
		if n > 64<<10 {
			t.Errorf("cap(%s) = %d after putEncoder", name, n)
		}
	}
}

// A skipper reads the values which a call skips with nothing of one left for the next: the names of the objects of a
// value past 64 levels, and of a value of a name twice, are dropped, and the values after them are read alike.
func TestSkipperLeavesNoNames(t *testing.T) {
	skipper := takeSkipper(false, false)
	defer releaseSkipper(skipper)
	e := (*encoder)(skipper)
	deep := []byte(`{"a":1,"b":` + strings.Repeat("[", 70) + strings.Repeat("]", 70) + `}`)
	if !validValue(skipper, deep) {
		t.Fatal("the value past 64 levels is not valid")
	}
	if n := len(e.st.names.ends); n != 0 {
		t.Fatalf("%d names left by the value past 64 levels", n)
	}
	for _, v := range []string{`{"a":1,"a":2}`, `{"a":{"b":1,"b":2}}`, `[{"a":1,"a":1}]`} {
		if validValue(skipper, []byte(v)) {
			t.Fatalf("%s: valid", v)
		}
		if n := len(e.st.names.ends); n != 0 {
			t.Fatalf("%s: %d names left", v, n)
		}
	}
	for _, v := range []string{`{"a":1,"b":2}`, `[{"a":1},{"a":2}]`, `{"a":{"a":1},"b":{"a":1}}`} {
		if !validValue(skipper, []byte(v)) {
			t.Fatalf("%s: not valid", v)
		}
	}
	if n := len(e.st.names.ends); n != 0 {
		t.Fatalf("%d names left", n)
	}
}
