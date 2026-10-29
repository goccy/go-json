package json_test

import (
	"strconv"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// arrayOf returns a JSON array of n numbers.
func arrayOf(n int) []byte {
	elems := make([]string, n)
	for i := range elems {
		elems[i] = strconv.Itoa(i)
	}
	return []byte("[" + strings.Join(elems, ",") + "]")
}

// A slice is made for its own value, as with encoding/json: its capacity doesn't depend on what the previous calls
// decoded, which a value kept out of a call would keep alive. The memory of a map is compared with the one of
// encoding/json by TestDecodedFootprintAsEncodingJSON, which the pools don't count in.
func TestDecodedSliceCapacityIndependentOfPreviousCalls(t *testing.T) {
	small, large := arrayOf(1), arrayOf(100)
	capAfter := func(previous []byte) int {
		var p []int
		if err := json.Unmarshal(previous, &p); err != nil {
			t.Fatal(err)
		}
		var v []int
		if err := json.Unmarshal(small, &v); err != nil {
			t.Fatal(err)
		}
		return cap(v)
	}
	if afterSmall, afterLarge := capAfter(small), capAfter(large); afterSmall != afterLarge {
		t.Fatalf("a slice of 1 element has a capacity of %d after an array of 1 element, %d after one of 100", afterSmall, afterLarge)
	}
}
