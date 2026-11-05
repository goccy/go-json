// Package codegen is what the code which gojsongen writes calls: it is for that code only, and a program has no other
// use of it.
//
// A file which gojsongen writes registers the struct types it was written for, as the package of the types is
// initialized: a StructDescriptor of each type, the members of its JSON object as gojsongen saw them.
package codegen

import (
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/generated"
)

// Version1 is referred to by the code which gojsongen writes, as the version of this package which the code is
// written for: a version of this package which can't run that code has no Version1, so that the code doesn't compile
// until gojsongen writes it again.
const Version1 = 1

// StructDescriptor describes a struct type for the encoder and the decoder.
type StructDescriptor struct {
	// Type is the struct type.
	Type reflect.Type
	// Fields are the members of the JSON object of the struct by the semantics of encoding/json, in the order
	// encoding/json writes them.
	Fields []FieldDescriptor

	// codec is the codec Register registers for the descriptor.
	codec generated.Codec
}

// FieldDescriptor describes a member of the JSON object of a struct.
type FieldDescriptor struct {
	// Name is the name of the member.
	Name string
	// Offset is the offset of the field of the member from the start of the struct, through the structs embedded in
	// it. No member of a descriptor is reached through an embedded pointer.
	Offset uintptr
	// Type is the type of the field.
	Type reflect.Type
	// Tag is the `json` tag of the field.
	Tag string
}

// Struct is a struct type of a file which gojsongen wrote, as Register registers it.
type Struct struct {
	Descriptor *StructDescriptor
}

// Register registers the struct types of a file which gojsongen wrote, and reports whether none of the types had
// been registered: a type keeps what it was registered with first. The file calls it as its package is initialized.
func Register(structs ...Struct) bool {
	file := &generated.File{Codecs: make([]*generated.Codec, len(structs))}
	for i, s := range structs {
		d := s.Descriptor
		// the members are the same structs by their types, which the conversion of their pointers requires
		fields := unsafe.Slice((*generated.Field)(unsafe.SliceData(d.Fields)), len(d.Fields))
		d.codec = generated.Codec{Type: d.Type, Fields: fields, File: file}
		file.Codecs[i] = &d.codec
	}
	return generated.Register(file)
}
