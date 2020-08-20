package json

type wrappedStringDecoder struct {
	dec           decoder
	stringDecoder *stringDecoder
}

func newWrappedStringDecoder(dec decoder) *wrappedStringDecoder {
	return &wrappedStringDecoder{
		dec:           dec,
		stringDecoder: newStringDecoder(),
	}
}

func (d *wrappedStringDecoder) decodeStream(s *stream, p uintptr) error {
	bytes, err := d.stringDecoder.decodeStreamByte(s)
	if err != nil {
		return err
	}

	// save current state
	buf := s.buf
	length := s.length
	cursor := s.cursor

	// set content in string to stream
	bytes = append(bytes, nul)
	s.buf = bytes
	s.cursor = 0
	s.length = int64(len(bytes))
	if err := d.dec.decodeStream(s, p); err != nil {
		return nil
	}

	// restore state
	s.buf = buf
	s.length = length
	s.cursor = cursor
	return nil
}

func (d *wrappedStringDecoder) decode(buf []byte, cursor int64, p uintptr) (int64, error) {
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
