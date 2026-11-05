// Package generated keeps what gojsongen wrote for the struct types of a program, by type: a type is registered by the
// package which declares it, as the package is initialized, so that nothing imports the packages of the types.
package generated

import (
	"reflect"
	"sync"
)

// Field is a member of the JSON object of a struct, as codegen.FieldDescriptor describes it.
type Field struct {
	Name   string
	Offset uintptr
	Type   reflect.Type
	Tag    string
}

// Codec is what gojsongen wrote for a struct type.
type Codec struct {
	// Type is the struct type.
	Type reflect.Type
	// Fields are the members of the JSON object of the struct as gojsongen saw them, by the semantics of
	// encoding/json, in the order encoding/json writes them.
	Fields []Field
	// File is the file gojsongen wrote the codec in.
	File *File
}

// File is the codecs of a file which gojsongen wrote.
type File struct {
	Codecs []*Codec
}

var codecs sync.Map // map[reflect.Type]*Codec

// Register registers the codecs of f by their types, and reports whether none of the types had a codec: a type which
// has one keeps it.
func Register(f *File) bool {
	registered := true
	for _, c := range f.Codecs {
		if _, loaded := codecs.LoadOrStore(c.Type, c); loaded {
			registered = false
		}
	}
	return registered
}

// Lookup returns the codec of the type, or nil.
func Lookup(t reflect.Type) *Codec {
	if v, ok := codecs.Load(t); ok {
		return v.(*Codec)
	}
	return nil
}
