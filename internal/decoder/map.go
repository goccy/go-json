package decoder

import (
	"reflect"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type mapDecoder struct {
	mapType   reflect.Type
	keyType   reflect.Type
	valueType reflect.Type
	// mapTypePtr, keyPtrType and valuePtrType are the type descriptors of the map, of *K and of *V.
	mapTypePtr   unsafe.Pointer
	keyPtrType   unsafe.Pointer
	valuePtrType unsafe.Pointer
	// isStringAnyMap is whether the map is a map[string]interface{}, which is decoded without reflect.
	isStringAnyMap bool
	keyDecoder     Decoder
	valueDecoder   Decoder
	structName     string
	fieldName      string
}

func newMapDecoder(mapType reflect.Type, keyType reflect.Type, keyDec Decoder, valueType reflect.Type, valueDec Decoder, structName, fieldName string) *mapDecoder {
	_, isIfaceValue := valueDec.(*interfaceDecoder)
	return &mapDecoder{
		mapType:        mapType,
		keyDecoder:     keyDec,
		keyType:        keyType,
		valueType:      valueType,
		mapTypePtr:     runtime.TypePtr(mapType),
		keyPtrType:     ptrTypeOf(keyType),
		valuePtrType:   ptrTypeOf(valueType),
		isStringAnyMap: mapType == interfaceMapType && isIfaceValue,
		valueDecoder:   valueDec,
		structName:     structName,
		fieldName:      fieldName,
	}
}

// mapValue returns the reflect.Value of the map m.
func (d *mapDecoder) mapValue(m unsafe.Pointer) reflect.Value {
	// A map is a pointer in an interface value.
	return reflect.ValueOf(*(*any)(unsafe.Pointer(&emptyInterface{typ: d.mapTypePtr, ptr: m})))
}

func (d *mapDecoder) Decode(ctx *RuntimeContext, cursor, depth int64, p unsafe.Pointer) (int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	cursor = skipWhiteSpace(buf, cursor)
	buflen := int64(len(buf))
	if buflen < 2 {
		return 0, errors.ErrExpected("{} for map", cursor)
	}
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return 0, err
		}
		cursor += 4
		**(**unsafe.Pointer)(unsafe.Pointer(&p)) = nil
		return cursor, nil
	case '{':
	default:
		return 0, errors.ErrExpected("{ character for map value", cursor)
	}
	if d.isStringAnyMap {
		m := *(*map[string]any)(p)
		if m == nil {
			m, c, err := decodeNewStringAnyMap(ctx, d.valueDecoder.(*interfaceDecoder), cursor, depth)
			if err != nil {
				return 0, err
			}
			*(*map[string]any)(p) = m
			return c, nil
		}
		c, err := decodeStringAnyMap(ctx, d.valueDecoder.(*interfaceDecoder), m, cursor, depth)
		if err != nil {
			return 0, err
		}
		return c, nil
	}
	cursor++
	cursor = skipWhiteSpace(buf, cursor)
	mapValue := *(*unsafe.Pointer)(p)
	if mapValue == nil {
		mapValue = reflect.MakeMapWithSize(d.mapType, 0).UnsafePointer()
	}
	if buf[cursor] == '}' {
		**(**unsafe.Pointer)(unsafe.Pointer(&p)) = mapValue
		cursor++
		return cursor, nil
	}
	// The key and the value of every entry are decoded into the same zero values,
	// which reflect.Value.SetMapIndex copies into the map.
	k := newValue(d.keyType)
	v := newValue(d.valueType)
	mv := d.mapValue(mapValue)
	kv := valueAt(d.keyPtrType, k)
	vv := valueAt(d.valuePtrType, v)
	for {
		keyCursor, err := d.keyDecoder.Decode(ctx, cursor, depth, k)
		if err != nil {
			return 0, err
		}
		cursor = skipWhiteSpace(buf, keyCursor)
		if buf[cursor] != ':' {
			return 0, errors.ErrExpected("colon after object key", cursor)
		}
		cursor++
		valueCursor, err := d.valueDecoder.Decode(ctx, cursor, depth, v)
		if err != nil {
			return 0, err
		}
		mv.SetMapIndex(kv, vv)
		kv.SetZero()
		vv.SetZero()
		cursor = skipWhiteSpace(buf, valueCursor)
		if buf[cursor] == '}' {
			**(**unsafe.Pointer)(unsafe.Pointer(&p)) = mapValue
			cursor++
			return cursor, nil
		}
		if buf[cursor] != ',' {
			return 0, errors.ErrExpected("comma after object value", cursor)
		}
		cursor++
	}
}

func (d *mapDecoder) DecodePath(ctx *RuntimeContext, cursor, depth int64) ([][]byte, int64, error) {
	buf := ctx.Buf
	depth++
	if depth > maxDecodeNestingDepth {
		return nil, 0, errors.ErrExceededMaxDepth(buf[cursor], cursor)
	}

	cursor = skipWhiteSpace(buf, cursor)
	buflen := int64(len(buf))
	if buflen < 2 {
		return nil, 0, errors.ErrExpected("{} for map", cursor)
	}
	switch buf[cursor] {
	case 'n':
		if err := validateNull(buf, cursor); err != nil {
			return nil, 0, err
		}
		cursor += 4
		return [][]byte{nullbytes}, cursor, nil
	case '{':
	default:
		return nil, 0, errors.ErrExpected("{ character for map value", cursor)
	}
	cursor++
	cursor = skipWhiteSpace(buf, cursor)
	if buf[cursor] == '}' {
		cursor++
		return nil, cursor, nil
	}
	keyDecoder, ok := d.keyDecoder.(*stringDecoder)
	if !ok {
		return nil, 0, &errors.UnmarshalTypeError{
			Value:  "string",
			Type:   reflect.TypeOf(""),
			Offset: cursor,
			Struct: d.structName,
			Field:  d.fieldName,
		}
	}
	ret := [][]byte{}
	for {
		key, keyCursor, err := keyDecoder.decodeByte(buf, cursor)
		if err != nil {
			return nil, 0, err
		}
		cursor = skipWhiteSpace(buf, keyCursor)
		if buf[cursor] != ':' {
			return nil, 0, errors.ErrExpected("colon after object key", cursor)
		}
		cursor++
		child, found, err := ctx.Option.Path.Field(string(key))
		if err != nil {
			return nil, 0, err
		}
		if found {
			if child != nil {
				oldPath := ctx.Option.Path.node
				ctx.Option.Path.node = child
				paths, c, err := d.valueDecoder.DecodePath(ctx, cursor, depth)
				if err != nil {
					return nil, 0, err
				}
				ctx.Option.Path.node = oldPath
				ret = append(ret, paths...)
				cursor = c
			} else {
				start := cursor
				end, err := skipValue(buf, cursor, depth)
				if err != nil {
					return nil, 0, err
				}
				ret = append(ret, buf[start:end])
				cursor = end
			}
		} else {
			c, err := skipValue(buf, cursor, depth)
			if err != nil {
				return nil, 0, err
			}
			cursor = c
		}
		cursor = skipWhiteSpace(buf, cursor)
		if buf[cursor] == '}' {
			cursor++
			return ret, cursor, nil
		}
		if buf[cursor] != ',' {
			return nil, 0, errors.ErrExpected("comma after object value", cursor)
		}
		cursor++
	}
}
