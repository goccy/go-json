//go:build go1.24

package encoder

import (
	"reflect"
	"time"
	"unsafe"
)

// The marshalers of the types of the standard library are called often, time.Time the most, and what they return
// is known: their outputs are written by their appending methods instead, into the buffer, without the allocation
// of the output, and without the check of the output of MarshalJSON, which is valid and compact. Only the types
// of the packages which go-json imports anyway are here: importing a package for its type would link it into
// every program which uses go-json.

var (
	timeType    = reflect.TypeOf(time.Time{})
	timePtrType = reflect.PointerTo(timeType)
)

// stdMarshalerAppender returns the function which writes the output of the method of the interface iface of the
// receiver type recv, or nil for a type which is not one of them. The receiver is the data word of the interface
// value ( see MarshalerCall ): the address of a time.Time, for a time.Time and for a *time.Time alike.
func stdMarshalerAppender(recv, iface reflect.Type) func([]byte, unsafe.Pointer) ([]byte, bool) {
	if recv != timeType && recv != timePtrType {
		return nil
	}
	switch iface {
	case marshalJSONType:
		return appendTimeJSON
	case marshalTextType:
		return appendTimeText
	}
	return nil
}

// appendTimeJSON writes what time.Time.MarshalJSON returns: the text of AppendText quoted, which has nothing to
// escape. AppendText fails when MarshalJSON does, whose error is then the one of the call of MarshalJSON.
func appendTimeJSON(b []byte, p unsafe.Pointer) ([]byte, bool) {
	out, err := (*time.Time)(p).AppendText(append(b, '"'))
	if err != nil {
		return b, false
	}
	return append(out, '"'), true
}

// appendTimeText appends what time.Time.MarshalText returns.
func appendTimeText(b []byte, p unsafe.Pointer) ([]byte, bool) {
	out, err := (*time.Time)(p).AppendText(b)
	if err != nil {
		return b, false
	}
	return out, true
}
