package decoder

import (
	"reflect"
	"testing"
	"time"
)

func TestUnmarshalJSONDecoderRetainsNothing(t *testing.T) {
	// The bytes of the buffer are given to UnmarshalJSON of time.Time only.
	if !newUnmarshalJSONDecoder(reflect.PointerTo(reflect.TypeOf(time.Time{})), "", "").retainsNothing {
		t.Fatal("time.Time is not known to keep nothing")
	}
	type other struct{ time.Time }
	if newUnmarshalJSONDecoder(reflect.PointerTo(reflect.TypeOf(other{})), "", "").retainsNothing {
		t.Fatal("a type which embeds time.Time may keep the bytes")
	}
}
