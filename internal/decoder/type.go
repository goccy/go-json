package decoder

import (
	"encoding"
	"encoding/json"
	"reflect"
	"unsafe"
)

type Decoder interface {
	Decode([]byte, int64, int64, unsafe.Pointer) (int64, error)
	DecodeStream(*Stream, int64, unsafe.Pointer) error
}

const (
	nul                   = '\000'
	maxDecodeNestingDepth = 10000
)

var (
	unmarshalJSONType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	unmarshalTextType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)
