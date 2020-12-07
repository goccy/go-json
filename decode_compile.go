package json

import (
	"reflect"
	"strings"
	"unsafe"
)

func (d *Decoder) compileHead(typ *rtype) (decoder, error) {
	switch {
	case rtype_ptrTo(typ).Implements(unmarshalJSONType):
		return newUnmarshalJSONDecoder(rtype_ptrTo(typ), "", ""), nil
	case rtype_ptrTo(typ).Implements(unmarshalTextType):
		return newUnmarshalTextDecoder(rtype_ptrTo(typ), "", ""), nil
	}
	return d.compile(typ.Elem(), "", "")
}

func (d *Decoder) compile(typ *rtype, structName, fieldName string) (decoder, error) {
	switch {
	case rtype_ptrTo(typ).Implements(unmarshalJSONType):
		return newUnmarshalJSONDecoder(rtype_ptrTo(typ), structName, fieldName), nil
	case rtype_ptrTo(typ).Implements(unmarshalTextType):
		return newUnmarshalTextDecoder(rtype_ptrTo(typ), structName, fieldName), nil
	}

	switch typ.Kind() {
	case reflect.Ptr:
		return d.compilePtr(typ, structName, fieldName)
	case reflect.Struct:
		return d.compileStruct(typ, structName, fieldName)
	case reflect.Slice:
		elem := typ.Elem()
		if elem.Kind() == reflect.Uint8 {
			return d.compileBytes(structName, fieldName)
		}
		return d.compileSlice(typ, structName, fieldName)
	case reflect.Array:
		return d.compileArray(typ, structName, fieldName)
	case reflect.Map:
		return d.compileMap(typ, structName, fieldName)
	case reflect.Interface:
		return d.compileInterface(typ, structName, fieldName)
	case reflect.Uintptr:
		return d.compileUint(structName, fieldName)
	case reflect.Int:
		return d.compileInt(structName, fieldName)
	case reflect.Int8:
		return d.compileInt8(structName, fieldName)
	case reflect.Int16:
		return d.compileInt16(structName, fieldName)
	case reflect.Int32:
		return d.compileInt32(structName, fieldName)
	case reflect.Int64:
		return d.compileInt64(structName, fieldName)
	case reflect.Uint:
		return d.compileUint(structName, fieldName)
	case reflect.Uint8:
		return d.compileUint8(structName, fieldName)
	case reflect.Uint16:
		return d.compileUint16(structName, fieldName)
	case reflect.Uint32:
		return d.compileUint32(structName, fieldName)
	case reflect.Uint64:
		return d.compileUint64(structName, fieldName)
	case reflect.String:
		return d.compileString(structName, fieldName)
	case reflect.Bool:
		return d.compileBool(structName, fieldName)
	case reflect.Float32:
		return d.compileFloat32(structName, fieldName)
	case reflect.Float64:
		return d.compileFloat64(structName, fieldName)
	}
	return nil, &UnsupportedTypeError{Type: rtype2type(typ)}
}

func (d *Decoder) compilePtr(typ *rtype, structName, fieldName string) (decoder, error) {
	dec, err := d.compile(typ.Elem(), structName, fieldName)
	if err != nil {
		return nil, err
	}
	return newPtrDecoder(dec, typ.Elem(), structName, fieldName), nil
}

func (d *Decoder) compileInt(structName, fieldName string) (decoder, error) {
	return newIntDecoder(structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int)(p) = int(v)
	}), nil
}

func (d *Decoder) compileInt8(structName, fieldName string) (decoder, error) {
	return newIntDecoder(structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int8)(p) = int8(v)
	}), nil
}

func (d *Decoder) compileInt16(structName, fieldName string) (decoder, error) {
	return newIntDecoder(structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int16)(p) = int16(v)
	}), nil
}

func (d *Decoder) compileInt32(structName, fieldName string) (decoder, error) {
	return newIntDecoder(structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int32)(p) = int32(v)
	}), nil
}

func (d *Decoder) compileInt64(structName, fieldName string) (decoder, error) {
	return newIntDecoder(structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int64)(p) = v
	}), nil
}

func (d *Decoder) compileUint(structName, fieldName string) (decoder, error) {
	return newUintDecoder(structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint)(p) = uint(v)
	}), nil
}

func (d *Decoder) compileUint8(structName, fieldName string) (decoder, error) {
	return newUintDecoder(structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint8)(p) = uint8(v)
	}), nil
}

func (d *Decoder) compileUint16(structName, fieldName string) (decoder, error) {
	return newUintDecoder(structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint16)(p) = uint16(v)
	}), nil
}

func (d *Decoder) compileUint32(structName, fieldName string) (decoder, error) {
	return newUintDecoder(structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint32)(p) = uint32(v)
	}), nil
}

func (d *Decoder) compileUint64(structName, fieldName string) (decoder, error) {
	return newUintDecoder(structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint64)(p) = v
	}), nil
}

func (d *Decoder) compileFloat32(structName, fieldName string) (decoder, error) {
	return newFloatDecoder(structName, fieldName, func(p unsafe.Pointer, v float64) {
		*(*float32)(p) = float32(v)
	}), nil
}

func (d *Decoder) compileFloat64(structName, fieldName string) (decoder, error) {
	return newFloatDecoder(structName, fieldName, func(p unsafe.Pointer, v float64) {
		*(*float64)(p) = v
	}), nil
}

func (d *Decoder) compileString(structName, fieldName string) (decoder, error) {
	return newStringDecoder(structName, fieldName), nil
}

func (d *Decoder) compileBool(structName, fieldName string) (decoder, error) {
	return newBoolDecoder(structName, fieldName), nil
}

func (d *Decoder) compileBytes(structName, fieldName string) (decoder, error) {
	return newBytesDecoder(structName, fieldName), nil
}

func (d *Decoder) compileSlice(typ *rtype, structName, fieldName string) (decoder, error) {
	elem := typ.Elem()
	decoder, err := d.compile(elem, structName, fieldName)
	if err != nil {
		return nil, err
	}
	return newSliceDecoder(decoder, elem, elem.Size(), structName, fieldName), nil
}

func (d *Decoder) compileArray(typ *rtype, structName, fieldName string) (decoder, error) {
	elem := typ.Elem()
	decoder, err := d.compile(elem, structName, fieldName)
	if err != nil {
		return nil, err
	}
	return newArrayDecoder(decoder, elem, typ.Len(), structName, fieldName), nil
}

func (d *Decoder) compileMap(typ *rtype, structName, fieldName string) (decoder, error) {
	keyDec, err := d.compile(typ.Key(), structName, fieldName)
	if err != nil {
		return nil, err
	}
	valueDec, err := d.compile(typ.Elem(), structName, fieldName)
	if err != nil {
		return nil, err
	}
	return newMapDecoder(typ, keyDec, valueDec, structName, fieldName), nil
}

func (d *Decoder) compileInterface(typ *rtype, structName, fieldName string) (decoder, error) {
	return newInterfaceDecoder(typ, structName, fieldName), nil
}

func (d *Decoder) compileStruct(typ *rtype, structName, fieldName string) (decoder, error) {
	fieldNum := typ.NumField()
	fieldMap := map[string]*structFieldSet{}
	typeptr := uintptr(unsafe.Pointer(typ))
	if dec, exists := d.structTypeToDecoder[typeptr]; exists {
		return dec, nil
	}
	structDec := newStructDecoder(structName, fieldName, fieldMap)
	d.structTypeToDecoder[typeptr] = structDec
	structName = typ.Name()
	for i := 0; i < fieldNum; i++ {
		field := typ.Field(i)
		if isIgnoredStructField(field) {
			continue
		}
		tag := structTagFromField(field)
		dec, err := d.compile(type2rtype(field.Type), structName, field.Name)
		if err != nil {
			return nil, err
		}
		if tag.isString {
			dec = newWrappedStringDecoder(dec, structName, field.Name)
		}
		fieldSet := &structFieldSet{dec: dec, offset: field.Offset}
		fieldMap[field.Name] = fieldSet
		fieldMap[tag.key] = fieldSet
		fieldMap[strings.ToLower(tag.key)] = fieldSet
	}
	delete(d.structTypeToDecoder, typeptr)
	return structDec, nil
}
