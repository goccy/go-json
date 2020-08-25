package json

import (
	"reflect"
	"strings"
	"unsafe"
)

func (d *Decoder) compileHead(typ *rtype) (decoder, error) {
	switch {
	case typ.Implements(unmarshalJSONType):
		return newUnmarshalJSONDecoder(typ), nil
	case rtype_ptrTo(typ).Implements(marshalJSONType):
		return newUnmarshalJSONDecoder(rtype_ptrTo(typ)), nil
	case typ.Implements(unmarshalTextType):
		return newUnmarshalTextDecoder(typ), nil
	case rtype_ptrTo(typ).Implements(unmarshalTextType):
		return newUnmarshalTextDecoder(rtype_ptrTo(typ)), nil
	}
	return d.compile(typ.Elem())
}

func (d *Decoder) compile(typ *rtype) (decoder, error) {
	switch {
	case typ.Implements(unmarshalJSONType):
		return newUnmarshalJSONDecoder(typ), nil
	case rtype_ptrTo(typ).Implements(marshalJSONType):
		return newUnmarshalJSONDecoder(rtype_ptrTo(typ)), nil
	case typ.Implements(unmarshalTextType):
		return newUnmarshalTextDecoder(typ), nil
	case rtype_ptrTo(typ).Implements(unmarshalTextType):
		return newUnmarshalTextDecoder(rtype_ptrTo(typ)), nil
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return d.compilePtr(typ)
	case reflect.Struct:
		return d.compileStruct(typ)
	case reflect.Slice:
		return d.compileSlice(typ)
	case reflect.Array:
		return d.compileArray(typ)
	case reflect.Map:
		return d.compileMap(typ)
	case reflect.Interface:
		return d.compileInterface(typ)
	case reflect.Uintptr:
		return d.compileUint()
	case reflect.Int:
		return d.compileInt()
	case reflect.Int8:
		return d.compileInt8()
	case reflect.Int16:
		return d.compileInt16()
	case reflect.Int32:
		return d.compileInt32()
	case reflect.Int64:
		return d.compileInt64()
	case reflect.Uint:
		return d.compileUint()
	case reflect.Uint8:
		return d.compileUint8()
	case reflect.Uint16:
		return d.compileUint16()
	case reflect.Uint32:
		return d.compileUint32()
	case reflect.Uint64:
		return d.compileUint64()
	case reflect.String:
		return d.compileString()
	case reflect.Bool:
		return d.compileBool()
	case reflect.Float32:
		return d.compileFloat32()
	case reflect.Float64:
		return d.compileFloat64()
	}
	return nil, &UnsupportedTypeError{Type: rtype2type(typ)}
}

func (d *Decoder) compilePtr(typ *rtype) (decoder, error) {
	dec, err := d.compile(typ.Elem())
	if err != nil {
		return nil, err
	}
	return newPtrDecoder(dec, typ.Elem()), nil
}

func (d *Decoder) compileInt() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int)(unsafe.Pointer(p)) = int(v)
	}), nil
}

func (d *Decoder) compileInt8() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int8)(unsafe.Pointer(p)) = int8(v)
	}), nil
}

func (d *Decoder) compileInt16() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int16)(unsafe.Pointer(p)) = int16(v)
	}), nil
}

func (d *Decoder) compileInt32() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int32)(unsafe.Pointer(p)) = int32(v)
	}), nil
}

func (d *Decoder) compileInt64() (decoder, error) {
	return newIntDecoder(func(p uintptr, v int64) {
		*(*int64)(unsafe.Pointer(p)) = v
	}), nil
}

func (d *Decoder) compileUint() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint)(unsafe.Pointer(p)) = uint(v)
	}), nil
}

func (d *Decoder) compileUint8() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint8)(unsafe.Pointer(p)) = uint8(v)
	}), nil
}

func (d *Decoder) compileUint16() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint16)(unsafe.Pointer(p)) = uint16(v)
	}), nil
}

func (d *Decoder) compileUint32() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint32)(unsafe.Pointer(p)) = uint32(v)
	}), nil
}

func (d *Decoder) compileUint64() (decoder, error) {
	return newUintDecoder(func(p uintptr, v uint64) {
		*(*uint64)(unsafe.Pointer(p)) = v
	}), nil
}

func (d *Decoder) compileFloat32() (decoder, error) {
	return newFloatDecoder(func(p uintptr, v float64) {
		*(*float32)(unsafe.Pointer(p)) = float32(v)
	}), nil
}

func (d *Decoder) compileFloat64() (decoder, error) {
	return newFloatDecoder(func(p uintptr, v float64) {
		*(*float64)(unsafe.Pointer(p)) = v
	}), nil
}

func (d *Decoder) compileString() (decoder, error) {
	return newStringDecoder(), nil
}

func (d *Decoder) compileBool() (decoder, error) {
	return newBoolDecoder(), nil
}

func (d *Decoder) compileSlice(typ *rtype) (decoder, error) {
	elem := typ.Elem()
	decoder, err := d.compile(elem)
	if err != nil {
		return nil, err
	}
	return newSliceDecoder(decoder, elem, elem.Size()), nil
}

func (d *Decoder) compileArray(typ *rtype) (decoder, error) {
	elem := typ.Elem()
	decoder, err := d.compile(elem)
	if err != nil {
		return nil, err
	}
	return newArrayDecoder(decoder, elem, typ.Len()), nil
}

func (d *Decoder) compileMap(typ *rtype) (decoder, error) {
	keyDec, err := d.compile(typ.Key())
	if err != nil {
		return nil, err
	}
	valueDec, err := d.compile(typ.Elem())
	if err != nil {
		return nil, err
	}
	return newMapDecoder(typ, keyDec, valueDec), nil
}

func (d *Decoder) compileInterface(typ *rtype) (decoder, error) {
	return newInterfaceDecoder(typ), nil
}

func (d *Decoder) compileStruct(typ *rtype) (decoder, error) {
	fieldNum := typ.NumField()
	fieldMap := map[string]*structFieldSet{}
	typeptr := uintptr(unsafe.Pointer(typ))
	if dec, exists := d.structTypeToDecoder[typeptr]; exists {
		return dec, nil
	}
	structDec := newStructDecoder(fieldMap)
	d.structTypeToDecoder[typeptr] = structDec
	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if isIgnoredStructField(field) {
			continue
		}
		tag := structTagFromField(field)
		dec, err := d.compile(type2rtype(field.Type))
		if err != nil {
			return nil, err
		}
		if tag.isString {
			dec = newWrappedStringDecoder(dec)
		}
		fieldSet := &structFieldSet{dec: dec, offset: field.Offset}
		fieldMap[field.Name] = fieldSet
		fieldMap[tag.key] = fieldSet
		fieldMap[strings.ToLower(tag.key)] = fieldSet
	}
	delete(d.structTypeToDecoder, typeptr)
	return structDec, nil
}
