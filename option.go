package json

type EncodeOption func(*Encoder) error

func UnorderedMap() EncodeOption {
	return func(e *Encoder) error {
		e.unorderedMap = true
		return nil
	}
}
