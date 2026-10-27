//go:build go1.27 && goexperiment.jsonv2

package decoder

import (
	"bytes"
	stdjsontext "encoding/json/jsontext"
	stdjson "encoding/json/v2"
	"reflect"
	"sync"

	"github.com/goccy/go-json/internal/errors"
)

// stdUnmarshalerFromType is the interface of UnmarshalJSONFrom with the Decoder of encoding/json/jsontext, which
// encoding/json calls where it is built on encoding/json/v2.
var stdUnmarshalerFromType = reflect.TypeFor[stdjson.UnmarshalerFrom]()

// stdFromDecoder is a decoder of encoding/json/jsontext which an UnmarshalJSONFrom method reads from, and the reader
// of its input.
type stdFromDecoder struct {
	r   bytes.Reader
	dec *stdjsontext.Decoder
}

var stdFromDecoders = sync.Pool{New: func() any {
	d := new(stdFromDecoder)
	d.dec = stdjsontext.NewDecoder(&d.r)
	return d
}}

// stdUnmarshalFrom calls UnmarshalJSONFrom of v, a stdjson.UnmarshalerFrom, with a decoder of the value, as
// unmarshalFrom does.
func stdUnmarshalFrom(v any, value []byte) (bool, error) {
	d := stdFromDecoders.Get().(*stdFromDecoder)
	defer stdFromDecoders.Put(d)
	d.r.Reset(value)
	d.dec.Reset(&d.r, stdjsontext.AllowDuplicateNames(true), stdjsontext.AllowInvalidUTF8(true))
	err := v.(stdjson.UnmarshalerFrom).UnmarshalJSONFrom(d.dec)
	_, n := d.dec.StackIndex(0)
	declined, err := errors.MethodCallResult(err, d.dec.StackDepth(), n)
	d.r.Reset(nil)
	return declined, err
}
