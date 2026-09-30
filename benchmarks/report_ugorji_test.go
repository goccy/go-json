package benchmark

import (
	"reflect"
	"sync"

	"benchmark/report"

	"github.com/ugorji/go/codec"
)

// The configurations of ugorji/go/codec, a library of several formats whose JSON is configured by a JsonHandle.
//
// Its Encoder and Decoder are made for a handle and are meant to be reused, by ResetBytes, as its documentation
// says: they are kept in a pool, as the other libraries keep theirs. A value in an interface{} is decoded as
// encoding/json decodes it, an object as a map[string]interface{} and a number as a float64, by MapType and
// PreferFloat.

const ugorjiModule = "github.com/ugorji/go/codec"

// ugorjiJSON is a JsonHandle with its Encoders and Decoders.
type ugorjiJSON struct {
	h        *codec.JsonHandle
	enc, dec sync.Pool
}

func newUgorjiJSON(configure func(h *codec.JsonHandle)) *ugorjiJSON {
	h := &codec.JsonHandle{}
	h.MapType = reflect.TypeOf(map[string]any(nil))
	h.PreferFloat = true
	configure(h)
	u := &ugorjiJSON{h: h}
	u.enc.New = func() any { return codec.NewEncoderBytes(nil, h) }
	u.dec.New = func() any { return codec.NewDecoderBytes(nil, h) }
	return u
}

func (u *ugorjiJSON) marshal(v any) ([]byte, error) {
	e := u.enc.Get().(*codec.Encoder)
	var out []byte
	e.ResetBytes(&out)
	err := e.Encode(v)
	// the pooled Encoder must not keep the output alive
	var none []byte
	e.ResetBytes(&none)
	u.enc.Put(e)
	return out, err
}

func (u *ugorjiJSON) unmarshal(data []byte) func(any) error {
	return func(v any) error {
		d := u.dec.Get().(*codec.Decoder)
		d.ResetBytes(data)
		err := d.Decode(v)
		d.ResetBytes(nil)
		u.dec.Put(d)
		return err
	}
}

var (
	ugorjiStd = newUgorjiJSON(func(h *codec.JsonHandle) {
		h.Canonical = true
	})
	ugorjiFast = newUgorjiJSON(func(h *codec.JsonHandle) {
		h.HTMLCharsAsIs = true
		h.ZeroCopy = true
	})
)

func init() {
	insertAfter("std/segmentio", &reportConfig{
		Config: report.Config{ID: "std/ugorji", Library: "ugorji/go/codec", Category: "std", Title: "ugorji/go/codec",
			Setting: "codec.JsonHandle{ Canonical: true }; Encoder / Decoder reused by ResetBytes"},
		module: ugorjiModule, marshal: ugorjiStd.marshal, unmarshal: ugorjiStd.unmarshal,
	})
	insertAfter("fast/segmentio", &reportConfig{
		Config: report.Config{ID: "fast/ugorji", Library: "ugorji/go/codec", Category: "fast", Title: "ugorji/go/codec",
			Setting: "codec.JsonHandle{ HTMLCharsAsIs: true, ZeroCopy: true }; Encoder / Decoder reused by ResetBytes"},
		module: ugorjiModule, marshal: ugorjiFast.marshal, unmarshal: ugorjiFast.unmarshal,
	})
	insertAfter("fastest/segmentio", &reportConfig{
		Config: report.Config{ID: "fastest/ugorji", Library: "ugorji/go/codec", Category: "fastest", Title: "ugorji/go/codec",
			Setting: "codec.JsonHandle{ HTMLCharsAsIs: true, ZeroCopy: true }; Encoder / Decoder reused by ResetBytes"},
		module: ugorjiModule, marshal: ugorjiFast.marshal, unmarshal: ugorjiFast.unmarshal,
	})
}
