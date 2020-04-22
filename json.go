package json

import "bytes"

func Marshal(v interface{}) ([]byte, error) {
	var b *bytes.Buffer
	enc := NewEncoder(b)
	defer enc.release()
	return enc.encodeForMarshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	src := make([]byte, len(data))
	copy(src, data)
	var dec Decoder
	return dec.decodeForUnmarshal(src, v)
}
