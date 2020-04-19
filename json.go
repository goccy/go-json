package json

func Marshal(v interface{}) ([]byte, error) {
	enc := NewEncoder()
	defer enc.Release()
	return enc.Encode(v)
}
