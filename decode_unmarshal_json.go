package json

import (
	"unsafe"
)

type unmarshalJSONDecoder struct {
	typ *rtype
}

func newUnmarshalJSONDecoder(typ *rtype) *unmarshalJSONDecoder {
	return &unmarshalJSONDecoder{typ: typ}
}

func (d *unmarshalJSONDecoder) decode(buf []byte, cursor int, p uintptr) (int, error) {
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	v := *(*interface{})(unsafe.Pointer(&interfaceHeader{
		typ: d.typ,
		ptr: unsafe.Pointer(p),
	}))
	if err := v.(Unmarshaler).UnmarshalJSON(src); err != nil {
		return 0, err
	}
	return end, nil
}
