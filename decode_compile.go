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
			return d.compileBytes(elem, structName, fieldName)
		}
		return d.compileSlice(typ, structName, fieldName)
	case reflect.Array:
		return d.compileArray(typ, structName, fieldName)
	case reflect.Map:
		return d.compileMap(typ, structName, fieldName)
	case reflect.Interface:
		return d.compileInterface(typ, structName, fieldName)
	case reflect.Uintptr:
		return d.compileUint(typ, structName, fieldName)
	case reflect.Int:
		return d.compileInt(typ, structName, fieldName)
	case reflect.Int8:
		return d.compileInt8(typ, structName, fieldName)
	case reflect.Int16:
		return d.compileInt16(typ, structName, fieldName)
	case reflect.Int32:
		return d.compileInt32(typ, structName, fieldName)
	case reflect.Int64:
		return d.compileInt64(typ, structName, fieldName)
	case reflect.Uint:
		return d.compileUint(typ, structName, fieldName)
	case reflect.Uint8:
		return d.compileUint8(typ, structName, fieldName)
	case reflect.Uint16:
		return d.compileUint16(typ, structName, fieldName)
	case reflect.Uint32:
		return d.compileUint32(typ, structName, fieldName)
	case reflect.Uint64:
		return d.compileUint64(typ, structName, fieldName)
	case reflect.String:
		return d.compileString(structName, fieldName)
	case reflect.Bool:
		return d.compileBool(structName, fieldName)
	case reflect.Float32:
		return d.compileFloat32(structName, fieldName)
	case reflect.Float64:
		return d.compileFloat64(structName, fieldName)
	}
	return nil, &UnmarshalTypeError{
		Value:  "object",
		Type:   rtype2type(typ),
		Offset: 0,
	}
}

func (d *Decoder) compileMapKey(typ *rtype, structName, fieldName string) (decoder, error) {
	if rtype_ptrTo(typ).Implements(unmarshalTextType) {
		return newUnmarshalTextDecoder(rtype_ptrTo(typ), structName, fieldName), nil
	}
	dec, err := d.compile(typ, structName, fieldName)
	if err != nil {
		return nil, err
	}
	for {
		switch t := dec.(type) {
		case *stringDecoder, *interfaceDecoder:
			return dec, nil
		case *boolDecoder, *intDecoder, *uintDecoder, *numberDecoder:
			return newWrappedStringDecoder(dec, structName, fieldName), nil
		case *ptrDecoder:
			dec = t.dec
		default:
			goto ERROR
		}
	}
ERROR:
	return nil, &UnmarshalTypeError{
		Value:  "object",
		Type:   rtype2type(typ),
		Offset: 0,
	}
}

func (d *Decoder) compilePtr(typ *rtype, structName, fieldName string) (decoder, error) {
	dec, err := d.compile(typ.Elem(), structName, fieldName)
	if err != nil {
		return nil, err
	}
	return newPtrDecoder(dec, typ.Elem(), structName, fieldName), nil
}

func (d *Decoder) compileInt(typ *rtype, structName, fieldName string) (decoder, error) {
	return newIntDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int)(p) = int(v)
	}), nil
}

func (d *Decoder) compileInt8(typ *rtype, structName, fieldName string) (decoder, error) {
	return newIntDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int8)(p) = int8(v)
	}), nil
}

func (d *Decoder) compileInt16(typ *rtype, structName, fieldName string) (decoder, error) {
	return newIntDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int16)(p) = int16(v)
	}), nil
}

func (d *Decoder) compileInt32(typ *rtype, structName, fieldName string) (decoder, error) {
	return newIntDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int32)(p) = int32(v)
	}), nil
}

func (d *Decoder) compileInt64(typ *rtype, structName, fieldName string) (decoder, error) {
	return newIntDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v int64) {
		*(*int64)(p) = v
	}), nil
}

func (d *Decoder) compileUint(typ *rtype, structName, fieldName string) (decoder, error) {
	return newUintDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint)(p) = uint(v)
	}), nil
}

func (d *Decoder) compileUint8(typ *rtype, structName, fieldName string) (decoder, error) {
	return newUintDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint8)(p) = uint8(v)
	}), nil
}

func (d *Decoder) compileUint16(typ *rtype, structName, fieldName string) (decoder, error) {
	return newUintDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint16)(p) = uint16(v)
	}), nil
}

func (d *Decoder) compileUint32(typ *rtype, structName, fieldName string) (decoder, error) {
	return newUintDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v uint64) {
		*(*uint32)(p) = uint32(v)
	}), nil
}

func (d *Decoder) compileUint64(typ *rtype, structName, fieldName string) (decoder, error) {
	return newUintDecoder(typ, structName, fieldName, func(p unsafe.Pointer, v uint64) {
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

func (d *Decoder) compileBytes(typ *rtype, structName, fieldName string) (decoder, error) {
	return newBytesDecoder(typ, structName, fieldName), nil
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
	keyDec, err := d.compileMapKey(typ.Key(), structName, fieldName)
	if err != nil {
		return nil, err
	}
	valueDec, err := d.compile(typ.Elem(), structName, fieldName)
	if err != nil {
		return nil, err
	}
	return newMapDecoder(typ, typ.Key(), keyDec, typ.Elem(), valueDec, structName, fieldName), nil
}

func (d *Decoder) compileInterface(typ *rtype, structName, fieldName string) (decoder, error) {
	return newInterfaceDecoder(typ, structName, fieldName), nil
}

func (d *Decoder) removeConflictFields(fieldMap map[string]*structFieldSet, conflictedMap map[string]struct{}, dec *structDecoder, baseOffset uintptr) {
	for k, v := range dec.fieldMap {
		if _, exists := conflictedMap[k]; exists {
			// already conflicted key
			continue
		}
		set, exists := fieldMap[k]
		if !exists {
			fieldSet := &structFieldSet{
				dec:         v.dec,
				offset:      baseOffset + v.offset,
				isTaggedKey: v.isTaggedKey,
			}
			fieldMap[k] = fieldSet
			lower := strings.ToLower(k)
			if _, exists := fieldMap[lower]; !exists {
				fieldMap[lower] = fieldSet
			}
			continue
		}
		if set.isTaggedKey {
			if v.isTaggedKey {
				// conflict tag key
				delete(fieldMap, k)
				conflictedMap[k] = struct{}{}
				conflictedMap[strings.ToLower(k)] = struct{}{}
			}
		} else {
			if v.isTaggedKey {
				fieldSet := &structFieldSet{
					dec:         v.dec,
					offset:      baseOffset + v.offset,
					isTaggedKey: v.isTaggedKey,
				}
				fieldMap[k] = fieldSet
				lower := strings.ToLower(k)
				if _, exists := fieldMap[lower]; !exists {
					fieldMap[lower] = fieldSet
				}
			} else {
				// conflict tag key
				delete(fieldMap, k)
				conflictedMap[k] = struct{}{}
				conflictedMap[strings.ToLower(k)] = struct{}{}
			}
		}
	}
}

func (d *Decoder) compileStruct(typ *rtype, structName, fieldName string) (decoder, error) {
	fieldNum := typ.NumField()
	conflictedMap := map[string]struct{}{}
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
		if field.Anonymous && !tag.isTaggedKey {
			if stDec, ok := dec.(*structDecoder); ok {
				if type2rtype(field.Type) == typ {
					// recursive definition
					continue
				}
				d.removeConflictFields(fieldMap, conflictedMap, stDec, uintptr(field.Offset))
			} else if pdec, ok := dec.(*ptrDecoder); ok {
				contentDec := pdec.contentDecoder()
				if pdec.typ == typ {
					// recursive definition
					continue
				}
				if dec, ok := contentDec.(*structDecoder); ok {
					for k, v := range dec.fieldMap {
						if _, exists := conflictedMap[k]; exists {
							// already conflicted key
							continue
						}
						set, exists := fieldMap[k]
						if !exists {
							fieldSet := &structFieldSet{
								dec:         newAnonymousFieldDecoder(pdec.typ, v.offset, v.dec),
								offset:      uintptr(field.Offset),
								isTaggedKey: v.isTaggedKey,
							}
							fieldMap[k] = fieldSet
							lower := strings.ToLower(k)
							if _, exists := fieldMap[lower]; !exists {
								fieldMap[lower] = fieldSet
							}
							continue
						}
						if set.isTaggedKey {
							if v.isTaggedKey {
								// conflict tag key
								delete(fieldMap, k)
								conflictedMap[k] = struct{}{}
								conflictedMap[strings.ToLower(k)] = struct{}{}
							}
						} else {
							if v.isTaggedKey {
								fieldSet := &structFieldSet{
									dec:         newAnonymousFieldDecoder(pdec.typ, v.offset, v.dec),
									offset:      uintptr(field.Offset),
									isTaggedKey: v.isTaggedKey,
								}
								fieldMap[k] = fieldSet
								lower := strings.ToLower(k)
								if _, exists := fieldMap[lower]; !exists {
									fieldMap[lower] = fieldSet
								}
							} else {
								// conflict tag key
								delete(fieldMap, k)
								conflictedMap[k] = struct{}{}
								conflictedMap[strings.ToLower(k)] = struct{}{}
							}
						}
					}
				}
			}
		} else {
			if tag.isString {
				dec = newWrappedStringDecoder(dec, structName, field.Name)
			}
			fieldSet := &structFieldSet{dec: dec, offset: field.Offset, isTaggedKey: tag.isTaggedKey}
			if tag.key != "" {
				fieldMap[tag.key] = fieldSet
				lower := strings.ToLower(tag.key)
				if _, exists := fieldMap[lower]; !exists {
					fieldMap[lower] = fieldSet
				}
			} else {
				fieldMap[field.Name] = fieldSet
				lower := strings.ToLower(field.Name)
				if _, exists := fieldMap[lower]; !exists {
					fieldMap[lower] = fieldSet
				}
			}
		}
	}
	delete(d.structTypeToDecoder, typeptr)
	return structDec, nil
}
