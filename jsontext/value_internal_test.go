package jsontext

import (
	"strconv"
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
