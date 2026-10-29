package json_test

import (
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// arrayOf returns a JSON array of n numbers, and objectOf a JSON object of n members.
func arrayOf(n int) []byte {
	elems := make([]string, n)
	for i := range elems {
		elems[i] = strconv.Itoa(i)
	}
	return []byte("[" + strings.Join(elems, ",") + "]")
}

func objectOf(n int) []byte {
	members := make([]string, n)
	for i := range members {
		members[i] = `"k` + strconv.Itoa(i) + `":` + strconv.Itoa(i)
	}
	return []byte("{" + strings.Join(members, ",") + "}")
}

// allocatedBytes returns the bytes which f allocates, with the collector stopped, so that the pools keep what they
// have.
func allocatedBytes(f func()) uint64 {
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	f()
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc
}

// A slice or a map is made for its own value, as with encoding/json: its size doesn't depend on what the previous
// calls decoded, which a value kept out of a call would keep alive.
func TestDecodedSizeIndependentOfPreviousCalls(t *testing.T) {
	if raceEnabled {
		t.Skip("sync.Pool drops items at random under the race detector, which the bytes of a call count")
	}
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

	smallMap, largeMap := objectOf(1), objectOf(100)
	bytesAfter := func(previous []byte) uint64 {
		var p map[string]int
		if err := json.Unmarshal(previous, &p); err != nil {
			t.Fatal(err)
		}
		return allocatedBytes(func() {
			var v map[string]int
			if err := json.Unmarshal(smallMap, &v); err != nil {
				t.Fatal(err)
			}
		})
	}
	// the first call takes what the pools have no object for yet
	bytesAfter(smallMap)
	if afterSmall, afterLarge := bytesAfter(smallMap), bytesAfter(largeMap); afterSmall != afterLarge {
		t.Fatalf("a map of 1 member takes %d bytes after an object of 1 member, %d after one of 100", afterSmall, afterLarge)
	}
}
