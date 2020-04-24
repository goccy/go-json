package json

import "bytes"

func Marshal(v interface{}) ([]byte, error) {
	var b *bytes.Buffer
	enc := NewEncoder(b)
	bytes, err := enc.encodeForMarshal(v)
	if err != nil {
		enc.release()
		return nil, err
	}
	enc.release()
	return bytes, nil
}

func Unmarshal(data []byte, v interface{}) error {
	src := make([]byte, len(data))
	copy(src, data)
	var dec Decoder
	return dec.decodeForUnmarshal(src, v)
}

func UnmarshalNoEscape(data []byte, v interface{}) error {
	src := make([]byte, len(data))
	copy(src, data)
	var dec Decoder
	return dec.decodeForUnmarshalNoEscape(src, v)
}
