package codegen_test

import (
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/goccy/go-json/codegen"
	"github.com/goccy/go-json/internal/generated"
)

type address struct {
	City string `json:"city"`
	Zip  string `json:"zip,omitempty"`
}

type user struct {
	Name string
	Home *address `json:"home"`
}

// the descriptors which a file of gojsongen has for address and user
var (
	addressDescriptor = codegen.StructDescriptor{
		Type: reflect.TypeFor[address](),
		Fields: []codegen.FieldDescriptor{
			{Name: "city", Offset: unsafe.Offsetof(address{}.City), Type: reflect.TypeFor[string](), Tag: "city"},
			{Name: "zip", Offset: unsafe.Offsetof(address{}.Zip), Type: reflect.TypeFor[string](), Tag: "zip,omitempty"},
		},
	}
	userDescriptor = codegen.StructDescriptor{
		Type: reflect.TypeFor[user](),
		Fields: []codegen.FieldDescriptor{
			{Name: "Name", Offset: unsafe.Offsetof(user{}.Name), Type: reflect.TypeFor[string]()},
			{Name: "home", Offset: unsafe.Offsetof(user{}.Home), Type: reflect.TypeFor[*address](), Tag: "home"},
		},
	}
)

// the types are registered as a file of gojsongen registers them, as the package is initialized
var registered = codegen.Register(codegen.Struct{Descriptor: &addressDescriptor}, codegen.Struct{Descriptor: &userDescriptor})

func TestRegister(t *testing.T) {
	if !registered {
		t.Fatal("Register() = false, want true")
	}
	a, u := generated.Lookup(reflect.TypeFor[address]()), generated.Lookup(reflect.TypeFor[user]())
	if a == nil || u == nil {
		t.Fatalf("Lookup() = %v, %v, want the codecs of address and user", a, u)
	}
	if a.File != u.File || !slices.Equal(a.File.Codecs, []*generated.Codec{a, u}) {
		t.Errorf("the codecs are not the ones of one file: %v, %v", a.File, u.File)
	}
	for _, tt := range []struct {
		codec      *generated.Codec
		descriptor *codegen.StructDescriptor
	}{{a, &addressDescriptor}, {u, &userDescriptor}} {
		want := make([]generated.Field, len(tt.descriptor.Fields))
		for i, f := range tt.descriptor.Fields {
			want[i] = generated.Field{Name: f.Name, Offset: f.Offset, Type: f.Type, Tag: f.Tag}
		}
		if tt.codec.Type != tt.descriptor.Type || !slices.Equal(tt.codec.Fields, want) {
			t.Errorf("codec of %v: %v %v, want %v %v", tt.descriptor.Type, tt.codec.Type, tt.codec.Fields, tt.descriptor.Type, want)
		}
	}

	// a type which has been registered keeps what it was registered with first
	again := codegen.StructDescriptor{Type: reflect.TypeFor[address]()}
	if codegen.Register(codegen.Struct{Descriptor: &again}) {
		t.Error("Register() of a registered type = true, want false")
	}
	if got := generated.Lookup(reflect.TypeFor[address]()); got != a {
		t.Errorf("Lookup() after a second Register() = %v, want the first codec %v", got, a)
	}
	if got := generated.Lookup(reflect.TypeFor[struct{ A int }]()); got != nil {
		t.Errorf("Lookup() of a type which has no descriptor = %v, want nil", got)
	}
}
