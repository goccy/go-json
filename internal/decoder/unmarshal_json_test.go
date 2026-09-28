package decoder

import (
	"reflect"
	"testing"
	"time"
)

type embeddedTime struct{ time.Time }

type declaredTime struct{ time.Time }

func (*declaredTime) UnmarshalJSON([]byte) error { return nil }

func TestUnmarshalJSONDecoderRetainsNothing(t *testing.T) {
	// The bytes of the buffer are given to UnmarshalJSON of time.Time, and of a type which has it only by embedding
	// time.Time.
	if !newUnmarshalJSONDecoder(reflect.PointerTo(reflect.TypeOf(time.Time{})), "", "").retainsNothing {
		t.Fatal("time.Time is not known to keep nothing")
	}
	if !newUnmarshalJSONDecoder(reflect.PointerTo(reflect.TypeOf(embeddedTime{})), "", "").retainsNothing {
		t.Fatal("a type which only embeds time.Time is not known to keep nothing")
	}
	// A type which declares UnmarshalJSON may keep the bytes.
	if newUnmarshalJSONDecoder(reflect.PointerTo(reflect.TypeOf(declaredTime{})), "", "").retainsNothing {
		t.Fatal("a type which declares UnmarshalJSON may keep the bytes")
	}
}
