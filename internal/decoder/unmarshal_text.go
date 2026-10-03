package decoder

import (
	"encoding"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/jsonstring"
	"github.com/goccy/go-json/internal/runtime"
)

type unmarshalTextDecoder struct {
	typ        reflect.Type
	structName string
	fieldName  string
	// nullSetsZero is set for a value which null sets to its zero value ( see textUnmarshalerNullSetsZero ): null
	// leaves any other value as it is, and is not given to UnmarshalText.
	nullSetsZero bool
}

func newUnmarshalTextDecoder(typ reflect.Type, structName, fieldName string) *unmarshalTextDecoder {
	return &unmarshalTextDecoder{
		typ:          typ,
		structName:   structName,
		fieldName:    fieldName,
		nullSetsZero: textUnmarshalerNullSetsZero(typ.Elem().Kind()),
	}
}

var (
	nullbytes = []byte(`null`)
)

func (d *unmarshalTextDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	switch c := src[0]; {
	case c == 'n':
		if d.nullSetsZero {
			reflect.NewAt(d.typ.Elem(), p).Elem().SetZero()
		}
		return end, nil
	case c != '"':
		// a value of another kind than a string, which is a type error
		return ctx.textUnmarshalerKindError(start, depth, d.typ)
	}

	// a string, which skipValue checked
	src = jsonstring.UnquoteValid(src)
	v := *(*any)(unsafe.Pointer(&emptyInterface{
		typ: runtime.TypePtr(d.typ),
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&p)),
	}))
	if err := v.(encoding.TextUnmarshaler).UnmarshalText(src); err != nil {
		return ctx.methodError(cursor, end, err, d.structName, d.fieldName)
	}
	return end, nil
}

func (d *unmarshalTextDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: unmarshal text decoder does not support decode path")
}
