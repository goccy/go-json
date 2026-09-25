package json_test

import (
	stdjson "encoding/json"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	json "github.com/goccy/go-json"
)

type pointeeLevel int8

type pointeeName string

func (n pointeeName) Upper() string { return strings.ToUpper(string(n)) }

type pointeeFields struct {
	Bool       *bool
	Int        *int
	Int8       *int8
	Int16      *int16
	Int32      *int32
	Int64      *int64
	Uint       *uint
	Uint8      *uint8
	Uint16     *uint16
	Uint32     *uint32
	Uint64     *uint64
	Uintptr    *uintptr
	Float32    *float32
	Float64    *float64
	String     *string
	Time       *time.Time
	Level      *pointeeLevel
	Name       *pointeeName
	Names      []*pointeeName
	PtrPtr     **string
	Struct     *struct{ A *string }
	StringKind *json.Number
}

func TestDecodePointees(t *testing.T) {
	// The values which pointers are set to are allocated by the kinds of their types: the named types of the
	// basic kinds are decoded as encoding/json does, and are still intact after the collections of the garbage
	// which the decoding of other values causes.
	doc := `{"Bool":true,"Int":-1,"Int8":-8,"Int16":-16,"Int32":-32,"Int64":-64,"Uint":1,"Uint8":8,"Uint16":16,
		"Uint32":32,"Uint64":64,"Uintptr":7,"Float32":1.5,"Float64":2.5,"String":"string",
		"Time":"2026-09-25T12:00:00Z","Level":3,"Name":"name","Names":["a","b",null,"c"],"PtrPtr":"pp",
		"Struct":{"A":"a"},"StringKind":12.5}`
	var want pointeeFields
	if err := stdjson.Unmarshal([]byte(doc), &want); err != nil {
		t.Fatal(err)
	}
	var values []*pointeeFields
	for i := 0; i < 64; i++ {
		var got pointeeFields
		if err := json.Unmarshal([]byte(doc), &got); err != nil {
			t.Fatal(err)
		}
		values = append(values, &got)
		if i%8 == 0 {
			runtime.GC()
		}
	}
	runtime.GC()
	for _, got := range values {
		if !reflect.DeepEqual(*got, want) {
			t.Fatalf("got %+v, want %+v", *got, want)
		}
		if got.Name.Upper() != "NAME" {
			t.Fatalf("got the method of the name %q", got.Name.Upper())
		}
	}
}
