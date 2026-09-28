package runtime

import (
	"math/big"
	"net/netip"
	"reflect"
	"testing"
	"time"
)

type promotedTime struct{ time.Time }

type promotedTimeAfterField struct {
	N int64
	time.Time
}

type promotedTimePointer struct{ *time.Time }

type promotedTimeNested struct {
	A, B int64
	promotedTimeAfterField
}

type promotedTimeThroughPointer struct {
	N int64
	*promotedTimeAfterField
}

// promotedThroughEmbeddedPointer embeds a pointer to a type which embeds time.Time.
type promotedThroughEmbeddedPointer struct{ *promotedTime }

type promotedBigInt struct{ big.Int }

type declaredOnValue struct{ time.Time }

func (declaredOnValue) MarshalJSON() ([]byte, error) { return nil, nil }

type declaredOnPointer struct{ time.Time }

func (*declaredOnPointer) MarshalJSON() ([]byte, error) { return nil, nil }

type declaredUnmarshal struct{ time.Time }

func (*declaredUnmarshal) UnmarshalJSON([]byte) error { return nil }

type userMarshaler struct{}

func (userMarshaler) MarshalJSON() ([]byte, error) { return nil, nil }

type promotedUser struct{ userMarshaler }

// promotedTwice has the MarshalJSON of time.Time, and no MarshalText: both embedded types have one.
type promotedTwice struct {
	time.Time
	netip.Addr
}

type promotedNestedDeclared struct{ declaredOnValue }

// A method which a type has only by embedding a type of the standard library is found, with the place of the
// embedded value; a method which the type, or a type between, declares is not, whatever its receiver.
func TestPromotedStdMethod(t *testing.T) {
	timeType := reflect.TypeOf(time.Time{})
	tests := []struct {
		typ    reflect.Type
		name   string
		ok     bool
		origin reflect.Type
		offset uintptr
		inline bool
	}{
		{reflect.TypeOf(promotedTime{}), "MarshalJSON", true, timeType, 0, true},
		{reflect.TypeOf(&promotedTime{}), "MarshalJSON", true, timeType, 0, true},
		{reflect.TypeOf(promotedTime{}), "MarshalText", true, timeType, 0, true},
		{reflect.TypeOf(&promotedTime{}), "UnmarshalJSON", true, timeType, 0, true},
		{reflect.TypeOf(struct{ time.Time }{}), "MarshalJSON", true, timeType, 0, true},
		{reflect.TypeOf(promotedTimeAfterField{}), "MarshalJSON", true, timeType, 8, true},
		{reflect.TypeOf(promotedTimeNested{}), "MarshalJSON", true, timeType, 24, true},
		{reflect.TypeOf(promotedTimePointer{}), "MarshalJSON", true, reflect.PointerTo(timeType), 0, false},
		{reflect.TypeOf(promotedTimeThroughPointer{}), "MarshalJSON", true, timeType, 16, false},
		{reflect.TypeOf(promotedThroughEmbeddedPointer{}), "MarshalJSON", true, timeType, 0, false},
		{reflect.TypeOf(&promotedThroughEmbeddedPointer{}), "MarshalJSON", true, timeType, 0, false},
		{reflect.TypeOf(&promotedBigInt{}), "MarshalJSON", true, reflect.TypeOf(big.Int{}), 0, true},
		{reflect.TypeOf(declaredOnValue{}), "MarshalJSON", false, nil, 0, false},
		{reflect.TypeOf(&declaredOnValue{}), "MarshalJSON", false, nil, 0, false},
		{reflect.TypeOf(&declaredOnPointer{}), "MarshalJSON", false, nil, 0, false},
		{reflect.TypeOf(&declaredUnmarshal{}), "UnmarshalJSON", false, nil, 0, false},
		{reflect.TypeOf(&declaredUnmarshal{}), "MarshalJSON", true, timeType, 0, true},
		{reflect.TypeOf(promotedUser{}), "MarshalJSON", false, nil, 0, false},
		{reflect.TypeOf(promotedTwice{}), "MarshalJSON", true, timeType, 0, true},
		{reflect.TypeOf(promotedTwice{}), "MarshalText", false, nil, 0, false},
		{reflect.TypeOf(promotedNestedDeclared{}), "MarshalJSON", false, nil, 0, false},
		{timeType, "MarshalJSON", false, nil, 0, false},
		{reflect.TypeOf(0), "MarshalJSON", false, nil, 0, false},
	}
	for _, test := range tests {
		origin, offset, inline, ok := PromotedStdMethod(test.typ, test.name)
		if ok != test.ok || origin != test.origin || offset != test.offset || inline != test.inline {
			t.Errorf("%v.%s: got (%v, %d, %v, %v), want (%v, %d, %v, %v)", test.typ, test.name,
				origin, offset, inline, ok, test.origin, test.offset, test.inline, test.ok)
		}
	}
}

// The wrapper of a promoted method is told from a declared method by the file of its code: it would be told no
// more if the compiler named the file otherwise, and the methods would be called and checked again.
func TestIsGeneratedCode(t *testing.T) {
	promoted, _ := reflect.TypeOf(promotedTime{}).MethodByName("MarshalJSON")
	declared, _ := reflect.TypeOf(declaredOnValue{}).MethodByName("MarshalJSON")
	if !isGeneratedCode(promoted.Func.Pointer()) {
		t.Fatal("the wrapper of a promoted method is not found as generated code")
	}
	if isGeneratedCode(declared.Func.Pointer()) {
		t.Fatal("a declared method is found as generated code")
	}
	if isGeneratedCode(0) {
		t.Fatal("no code is found as generated code")
	}
}
