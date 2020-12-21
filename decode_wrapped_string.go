package json

import (
	"unsafe"
)

type wrappedStringDecoder struct {
	dec           decoder
	stringDecoder *stringDecoder
	structName    string
	fieldName     string
}

func newWrappedStringDecoder(dec decoder, structName, fieldName string) *wrappedStringDecoder {
	return &wrappedStringDecoder{
		dec:           dec,
		stringDecoder: newStringDecoder(structName, fieldName),
		structName:    structName,
		fieldName:     fieldName,
	}
}

func (d *wrappedStringDecoder) decodeStream(s *stream, p unsafe.Pointer) error {
	bytes, err := d.stringDecoder.decodeStreamByte(s)
	if err != nil {
		return err
	}
	b := make([]byte, len(bytes)+1)
	copy(b, bytes)
	if _, err := d.dec.decode(b, 0, p); err != nil {
		return err
	}
	return nil
}

func (d *wrappedStringDecoder) decode(buf []byte, cursor int64, p unsafe.Pointer) (int64, error) {
	bytes, c, err := d.stringDecoder.decodeByte(buf, cursor)
	if err != nil {
		return 0, err
	}
	bytes = append(bytes, nul)
	if _, err := d.dec.decode(bytes, 0, p); err != nil {
		return 0, err
	}
	return c, nil
}
