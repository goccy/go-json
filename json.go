package json

import "bytes"

func Marshal(v interface{}) ([]byte, error) {
	var b *bytes.Buffer
	enc := NewEncoder(b)
	defer enc.release()
	return enc.encodeForMarshal(v)
}
