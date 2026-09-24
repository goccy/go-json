package decoder

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type unmarshalJSONDecoder struct {
	typ        reflect.Type
	structName string
	fieldName  string
}

func newUnmarshalJSONDecoder(typ reflect.Type, structName, fieldName string) *unmarshalJSONDecoder {
	return &unmarshalJSONDecoder{
		typ:        typ,
		structName: structName,
		fieldName:  fieldName,
	}
}

func (d *unmarshalJSONDecoder) annotateError(cursor int64, err error) {
	switch e := err.(type) {
	case *errors.UnmarshalTypeError:
		e.Struct = d.structName
		e.Field = d.fieldName
	case *errors.SyntaxError:
		e.Offset = cursor
	}
}

func (d *unmarshalJSONDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	cursor = skipWhiteSpace(buf, cursor)
	start := cursor
	end, err := skipValue(buf, cursor, depth)
	if err != nil {
		return 0, err
	}
	src := buf[start:end]
	dst := make([]byte, len(src))
	copy(dst, src)

	v := *(*any)(unsafe.Pointer(&emptyInterface{
		typ: runtime.TypePtr(d.typ),
		ptr: p,
	}))
	// The method is chosen by what the type implements, not by the entry point:
	// json.Unmarshal may meet a type with the context method and UnmarshalContext one without it.
	switch v := v.(type) {
	case unmarshalerContext:
		c := ctx.Option.Context
		if (ctx.Option.Flags&ContextOption) == 0 || c == nil {
			c = context.Background()
		}
		if err := v.UnmarshalJSON(c, dst); err != nil {
			d.annotateError(cursor, err)
			return 0, err
		}
	case json.Unmarshaler:
		if err := v.UnmarshalJSON(dst); err != nil {
			d.annotateError(cursor, err)
			return 0, err
		}
	}
	return end, nil
}

func (d *unmarshalJSONDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	return nil, 0, fmt.Errorf("json: unmarshal json decoder does not support decode path")
}
