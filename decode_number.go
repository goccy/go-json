package json

import (
	"unsafe"
)

type numberDecoder struct {
	*floatDecoder
	op func(unsafe.Pointer, Number)
}

func newNumberDecoder(op func(unsafe.Pointer, Number)) *numberDecoder {
	return &numberDecoder{
		floatDecoder: newFloatDecoder(nil),
		op:           op,
	}
}

func (d *numberDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	bytes, err := d.floatDecoder.decodeStreamByte(s)
	if err != nil {
		return err
	}
	str := *(*string)(unsafe.Pointer(&bytes))
	d.op(p, Number(str))
	return nil
}

func (d *numberDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	bytes, c, err := d.floatDecoder.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	cursor = c
	s := *(*string)(unsafe.Pointer(&bytes))
	d.op(p, Number(s))
	return cursor, nil
}
