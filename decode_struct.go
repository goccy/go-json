package json

import (
	"errors"
	"unsafe"
)

type structFieldSet struct {
	dec    decoder
	offset uintptr
}

type structDecoder struct {
	fieldMap   map[string]*structFieldSet
	keyDecoder *stringDecoder
}

func newStructDecoder(fieldMap map[string]*structFieldSet) *structDecoder {
	return &structDecoder{
		fieldMap:   fieldMap,
		keyDecoder: newStringDecoder(),
	}
}

func (d *structDecoder) skipValue(ctx *context) error {
	ctx.skipWhiteSpace()
	braceCount := 0
	bracketCount := 0
	cursor := ctx.cursor
	buf := ctx.buf
	buflen := ctx.buflen
	for ; cursor < buflen; cursor++ {
		switch buf[cursor] {
		case '{':
			braceCount++
		case '[':
			bracketCount++
		case '}':
			braceCount--
			if braceCount == -1 && bracketCount == 0 {
				return nil
			}
		case ']':
			bracketCount--
		case ',':
			if bracketCount == 0 && braceCount == 0 {
				return nil
			}
		}
	}
	return errors.New("unexpected error value")
}

func (d *structDecoder) decode(ctx *context, p uintptr) error {
	ctx.skipWhiteSpace()
	buf := ctx.buf
	buflen := ctx.buflen
	cursor := ctx.cursor
	if buflen < 2 {
		return errors.New("unexpected error {}")
	}
	if buf[cursor] != '{' {
		return errors.New("unexpected error {")
	}
	cursor++
	for ; cursor < buflen; cursor++ {
		ctx.cursor = cursor
		key, err := d.keyDecoder.decodeByte(ctx)
		if err != nil {
			return err
		}
		cursor = ctx.skipWhiteSpace()
		if buf[cursor] != ':' {
			return errors.New("unexpected error invalid delimiter for object")
		}
		cursor++
		if cursor >= buflen {
			return errors.New("unexpected error missing value")
		}
		ctx.cursor = cursor
		k := *(*string)(unsafe.Pointer(&key))
		field, exists := d.fieldMap[k]
		if exists {
			//fmt.Printf("k = %s dec = %#v, p = %x\n", k, field.dec, p)
			if err := field.dec.decode(ctx, p+field.offset); err != nil {
				return err
			}
		} else {
			if err := d.skipValue(ctx); err != nil {
				return err
			}
		}
		cursor = ctx.skipWhiteSpace()
		if buf[cursor] == '}' {
			return nil
		}
		if buf[cursor] != ',' {
			return errors.New("unexpected error ,")
		}
	}
	return nil
}
