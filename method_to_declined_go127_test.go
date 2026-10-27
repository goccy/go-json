//go:build go1.27 && goexperiment.jsonv2

package json_test

import (
	"errors"
	"testing"

	"encoding/json/jsontext"
)

// The values whose MarshalJSONTo of encoding/json/jsontext declines, and its UnmarshalJSONFrom of a value of the
// string option, compared with encoding/json.

type stdToList struct {
	N    int
	Next *stdToList `json:",omitempty"`
}

func (v stdToList) MarshalJSONTo(enc *jsontext.Encoder) error {
	if v.N == 0 {
		return errors.ErrUnsupported
	}
	return enc.WriteToken(jsontext.Int(int64(v.N * 100)))
}

type stdToPtrJSON struct{ X int }

func (stdToPtrJSON) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }
func (*stdToPtrJSON) MarshalJSON() ([]byte, error)          { return []byte(`"pointer json"`), nil }

type stdToMixed struct{}

func (stdToMixed) MarshalJSON() ([]byte, error) { return []byte(`"json"`), nil }
func (*stdToMixed) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteToken(jsontext.String("to"))
}

type stdToChan chan int

func (stdToChan) MarshalJSONTo(*jsontext.Encoder) error { return errors.ErrUnsupported }

func TestEncodeStdMarshalJSONToDeclined(t *testing.T) {
	compareStdEncoding(t,
		stdToList{0, &stdToList{N: 1}}, &stdToList{0, &stdToList{0, &stdToList{N: 2}}}, []stdToList{{0, &stdToList{N: 3}}},
		stdToPtrJSON{}, &stdToPtrJSON{}, []stdToPtrJSON{{}},
		stdToMixed{}, &stdToMixed{}, []stdToMixed{{}}, &struct{ M stdToMixed }{},
		stdToChan(nil),
	)
}

type stdFromLength int

func (v *stdFromLength) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	raw, err := dec.ReadValue()
	*v = stdFromLength(len(raw))
	return err
}

func TestDecodeStdUnmarshalJSONFromStringOption(t *testing.T) {
	compareUnmarshal(t, `{"A":"12","P":"123"}`, func() any {
		return new(struct {
			A stdFromLength  `json:",string"`
			P *stdFromLength `json:",string"`
		})
	})
}
