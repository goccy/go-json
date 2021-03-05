package json_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

type coverMarshalJSON struct {
	A int
}

func (c coverMarshalJSON) MarshalJSON() ([]byte, error) {
	return []byte(`"hello"`), nil
}

type coverPtrMarshalJSON struct {
	B int
}

func (c *coverPtrMarshalJSON) MarshalJSON() ([]byte, error) {
	return []byte(`"hello"`), nil
}

func TestCoverMarshalJSON(t *testing.T) {
	type structMarshalJSON struct {
		A coverMarshalJSON `json:"a"`
	}
	type structMarshalJSONOmitEmpty struct {
		A coverMarshalJSON `json:"a,omitempty"`
	}
	type structMarshalJSONString struct {
		A coverMarshalJSON `json:"a,string"`
	}
	type structPtrMarshalJSON struct {
		A coverPtrMarshalJSON `json:"a"`
	}
	type structPtrMarshalJSONOmitEmpty struct {
		A coverPtrMarshalJSON `json:"a,omitempty"`
	}
	type structPtrMarshalJSONString struct {
		A coverPtrMarshalJSON `json:"a,string"`
	}

	type structMarshalJSONPtr struct {
		A *coverMarshalJSON `json:"a"`
	}
	type structMarshalJSONPtrOmitEmpty struct {
		A *coverMarshalJSON `json:"a,omitempty"`
	}
	type structMarshalJSONPtrString struct {
		A *coverMarshalJSON `json:"a,string"`
	}
	type structPtrMarshalJSONPtr struct {
		A *coverPtrMarshalJSON `json:"a"`
	}
	type structPtrMarshalJSONPtrOmitEmpty struct {
		A *coverPtrMarshalJSON `json:"a,omitempty"`
	}
	type structPtrMarshalJSONPtrString struct {
		A *coverPtrMarshalJSON `json:"a,string"`
	}

	tests := []struct {
		name string
		data interface{}
	}{
		// HeadMarshalJSONZero
		{
			name: "HeadMarshalJSONZero",
			data: struct {
				A coverMarshalJSON `json:"a"`
			}{},
		},
		{
			name: "HeadMarshalJSONZeroOmitEmpty",
			data: struct {
				A coverMarshalJSON `json:"a,omitempty"`
			}{},
		},
		{
			name: "HeadMarshalJSONZeroString",
			data: struct {
				A coverMarshalJSON `json:"a,string"`
			}{},
		},
		{
			name: "HeadPtrMarshalJSONZero",
			data: struct {
				A coverPtrMarshalJSON `json:"a"`
			}{},
		},
		{
			name: "HeadPtrMarshalJSONZeroOmitEmpty",
			data: struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
			}{},
		},
		{
			name: "HeadPtrMarshalJSONZeroString",
			data: struct {
				A coverPtrMarshalJSON `json:"a,string"`
			}{},
		},

		// HeadMarshalJSON
		{
			name: "HeadMarshalJSON",
			data: struct {
				A coverMarshalJSON `json:"a"`
			}{A: coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONOmitEmpty",
			data: struct {
				A coverMarshalJSON `json:"a,omitempty"`
			}{A: coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONString",
			data: struct {
				A coverMarshalJSON `json:"a,string"`
			}{A: coverMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSON",
			data: struct {
				A coverPtrMarshalJSON `json:"a"`
			}{A: coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONOmitEmpty",
			data: struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
			}{A: coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONString",
			data: struct {
				A coverPtrMarshalJSON `json:"a,string"`
			}{A: coverPtrMarshalJSON{}},
		},

		// HeadMarshalJSONPtr
		{
			name: "HeadMarshalJSONPtr",
			data: struct {
				A *coverMarshalJSON `json:"a"`
			}{A: &coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONPtrOmitEmpty",
			data: struct {
				A *coverMarshalJSON `json:"a,omitempty"`
			}{A: &coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONPtrString",
			data: struct {
				A *coverMarshalJSON `json:"a,string"`
			}{A: &coverMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONPtr",
			data: struct {
				A *coverPtrMarshalJSON `json:"a"`
			}{A: &coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONPtrOmitEmpty",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
			}{A: &coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONPtrString",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,string"`
			}{A: &coverPtrMarshalJSON{}},
		},

		// HeadMarshalJSONPtrNil
		{
			name: "HeadMarshalJSONPtrNil",
			data: struct {
				A *coverMarshalJSON `json:"a"`
			}{A: nil},
		},
		{
			name: "HeadMarshalJSONPtrNilOmitEmpty",
			data: struct {
				A *coverMarshalJSON `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name: "HeadMarshalJSONPtrNilString",
			data: struct {
				A *coverMarshalJSON `json:"a,string"`
			}{A: nil},
		},
		{
			name: "HeadPtrMarshalJSONPtrNil",
			data: struct {
				A *coverPtrMarshalJSON `json:"a"`
			}{A: nil},
		},
		{
			name: "HeadPtrMarshalJSONPtrNilOmitEmpty",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name: "HeadPtrMarshalJSONPtrNilString",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadMarshalJSONZero
		{
			name: "PtrHeadMarshalJSONZero",
			data: &struct {
				A coverMarshalJSON `json:"a"`
			}{},
		},
		{
			name: "PtrHeadMarshalJSONZeroOmitEmpty",
			data: &struct {
				A coverMarshalJSON `json:"a,omitempty"`
			}{},
		},
		{
			name: "PtrHeadMarshalJSONZeroString",
			data: &struct {
				A coverMarshalJSON `json:"a,string"`
			}{},
		},
		{
			name: "PtrHeadPtrMarshalJSONZero",
			data: &struct {
				A coverPtrMarshalJSON `json:"a"`
			}{},
		},
		{
			name: "PtrHeadPtrMarshalJSONZeroOmitEmpty",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
			}{},
		},
		{
			name: "PtrHeadPtrMarshalJSONZeroString",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,string"`
			}{},
		},

		// PtrHeadMarshalJSON
		{
			name: "PtrHeadMarshalJSON",
			data: &struct {
				A coverMarshalJSON `json:"a"`
			}{A: coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONOmitEmpty",
			data: &struct {
				A coverMarshalJSON `json:"a,omitempty"`
			}{A: coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONString",
			data: &struct {
				A coverMarshalJSON `json:"a,string"`
			}{A: coverMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSON",
			data: &struct {
				A coverPtrMarshalJSON `json:"a"`
			}{A: coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONOmitEmpty",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
			}{A: coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONString",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,string"`
			}{A: coverPtrMarshalJSON{}},
		},

		// PtrHeadMarshalJSONPtr
		{
			name: "PtrHeadMarshalJSONPtr",
			data: &struct {
				A *coverMarshalJSON `json:"a"`
			}{A: &coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONPtrOmitEmpty",
			data: &struct {
				A *coverMarshalJSON `json:"a,omitempty"`
			}{A: &coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONPtrString",
			data: &struct {
				A *coverMarshalJSON `json:"a,string"`
			}{A: &coverMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtr",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a"`
			}{A: &coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrOmitEmpty",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
			}{A: &coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrString",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,string"`
			}{A: &coverPtrMarshalJSON{}},
		},

		// PtrHeadMarshalJSONPtrNil
		{
			name: "PtrHeadMarshalJSONPtrNil",
			data: &struct {
				A *coverMarshalJSON `json:"a"`
			}{A: nil},
		},
		{
			name: "PtrHeadMarshalJSONPtrNilOmitEmpty",
			data: &struct {
				A *coverMarshalJSON `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name: "PtrHeadMarshalJSONPtrNilString",
			data: &struct {
				A *coverMarshalJSON `json:"a,string"`
			}{A: nil},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrNil",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a"`
			}{A: nil},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrNilOmitEmpty",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrNilString",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadMarshalJSONNil
		{
			name: "PtrHeadMarshalJSONNil",
			data: (*struct {
				A *coverMarshalJSON `json:"a"`
			})(nil),
		},
		{
			name: "PtrHeadMarshalJSONNilOmitEmpty",
			data: (*struct {
				A *coverMarshalJSON `json:"a,omitempty"`
			})(nil),
		},
		{
			name: "PtrHeadMarshalJSONNilString",
			data: (*struct {
				A *coverMarshalJSON `json:"a,string"`
			})(nil),
		},
		{
			name: "PtrHeadPtrMarshalJSONNil",
			data: (*struct {
				A *coverPtrMarshalJSON `json:"a"`
			})(nil),
		},
		{
			name: "PtrHeadPtrMarshalJSONNilOmitEmpty",
			data: (*struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
			})(nil),
		},
		{
			name: "PtrHeadPtrMarshalJSONNilString",
			data: (*struct {
				A *coverPtrMarshalJSON `json:"a,string"`
			})(nil),
		},

		// HeadMarshalJSONZeroMultiFields
		{
			name: "HeadMarshalJSONZeroMultiFields",
			data: struct {
				A coverMarshalJSON `json:"a"`
				B coverMarshalJSON `json:"b"`
				C coverMarshalJSON `json:"c"`
			}{},
		},
		{
			name: "HeadMarshalJSONZeroMultiFieldsOmitEmpty",
			data: struct {
				A coverMarshalJSON `json:"a,omitempty"`
				B coverMarshalJSON `json:"b,omitempty"`
				C coverMarshalJSON `json:"c,omitempty"`
			}{},
		},
		{
			name: "HeadMarshalJSONZeroMultiFields",
			data: struct {
				A coverMarshalJSON `json:"a,string"`
				B coverMarshalJSON `json:"b,string"`
				C coverMarshalJSON `json:"c,string"`
			}{},
		},
		{
			name: "HeadPtrMarshalJSONZeroMultiFields",
			data: struct {
				A coverPtrMarshalJSON `json:"a"`
				B coverPtrMarshalJSON `json:"b"`
				C coverPtrMarshalJSON `json:"c"`
			}{},
		},
		{
			name: "HeadPtrMarshalJSONZeroMultiFieldsOmitEmpty",
			data: struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
				B coverPtrMarshalJSON `json:"b,omitempty"`
				C coverPtrMarshalJSON `json:"c,omitempty"`
			}{},
		},
		{
			name: "HeadPtrMarshalJSONZeroMultiFields",
			data: struct {
				A coverPtrMarshalJSON `json:"a,string"`
				B coverPtrMarshalJSON `json:"b,string"`
				C coverPtrMarshalJSON `json:"c,string"`
			}{},
		},

		// HeadMarshalJSONMultiFields
		{
			name: "HeadMarshalJSONMultiFields",
			data: struct {
				A coverMarshalJSON `json:"a"`
				B coverMarshalJSON `json:"b"`
				C coverMarshalJSON `json:"c"`
			}{A: coverMarshalJSON{}, B: coverMarshalJSON{}, C: coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONMultiFieldsOmitEmpty",
			data: struct {
				A coverMarshalJSON `json:"a,omitempty"`
				B coverMarshalJSON `json:"b,omitempty"`
				C coverMarshalJSON `json:"c,omitempty"`
			}{A: coverMarshalJSON{}, B: coverMarshalJSON{}, C: coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONMultiFieldsString",
			data: struct {
				A coverMarshalJSON `json:"a,string"`
				B coverMarshalJSON `json:"b,string"`
				C coverMarshalJSON `json:"c,string"`
			}{A: coverMarshalJSON{}, B: coverMarshalJSON{}, C: coverMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONMultiFields",
			data: struct {
				A coverPtrMarshalJSON `json:"a"`
				B coverPtrMarshalJSON `json:"b"`
				C coverPtrMarshalJSON `json:"c"`
			}{A: coverPtrMarshalJSON{}, B: coverPtrMarshalJSON{}, C: coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONMultiFieldsOmitEmpty",
			data: struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
				B coverPtrMarshalJSON `json:"b,omitempty"`
				C coverPtrMarshalJSON `json:"c,omitempty"`
			}{A: coverPtrMarshalJSON{}, B: coverPtrMarshalJSON{}, C: coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONMultiFieldsString",
			data: struct {
				A coverPtrMarshalJSON `json:"a,string"`
				B coverPtrMarshalJSON `json:"b,string"`
				C coverPtrMarshalJSON `json:"c,string"`
			}{A: coverPtrMarshalJSON{}, B: coverPtrMarshalJSON{}, C: coverPtrMarshalJSON{}},
		},

		// HeadMarshalJSONPtrMultiFields
		{
			name: "HeadMarshalJSONPtrMultiFields",
			data: struct {
				A *coverMarshalJSON `json:"a"`
				B *coverMarshalJSON `json:"b"`
				C *coverMarshalJSON `json:"c"`
			}{A: &coverMarshalJSON{}, B: &coverMarshalJSON{}, C: &coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONPtrMultiFieldsOmitEmpty",
			data: struct {
				A *coverMarshalJSON `json:"a,omitempty"`
				B *coverMarshalJSON `json:"b,omitempty"`
				C *coverMarshalJSON `json:"c,omitempty"`
			}{A: &coverMarshalJSON{}, B: &coverMarshalJSON{}, C: &coverMarshalJSON{}},
		},
		{
			name: "HeadMarshalJSONPtrMultiFieldsString",
			data: struct {
				A *coverMarshalJSON `json:"a,string"`
				B *coverMarshalJSON `json:"b,string"`
				C *coverMarshalJSON `json:"c,string"`
			}{A: &coverMarshalJSON{}, B: &coverMarshalJSON{}, C: &coverMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONPtrMultiFields",
			data: struct {
				A *coverPtrMarshalJSON `json:"a"`
				B *coverPtrMarshalJSON `json:"b"`
				C *coverPtrMarshalJSON `json:"c"`
			}{A: &coverPtrMarshalJSON{}, B: &coverPtrMarshalJSON{}, C: &coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONPtrMultiFieldsOmitEmpty",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
				B *coverPtrMarshalJSON `json:"b,omitempty"`
				C *coverPtrMarshalJSON `json:"c,omitempty"`
			}{A: &coverPtrMarshalJSON{}, B: &coverPtrMarshalJSON{}, C: &coverPtrMarshalJSON{}},
		},
		{
			name: "HeadPtrMarshalJSONPtrMultiFieldsString",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,string"`
				B *coverPtrMarshalJSON `json:"b,string"`
				C *coverPtrMarshalJSON `json:"c,string"`
			}{A: &coverPtrMarshalJSON{}, B: &coverPtrMarshalJSON{}, C: &coverPtrMarshalJSON{}},
		},

		// HeadMarshalJSONPtrNilMultiFields
		{
			name: "HeadMarshalJSONPtrNilMultiFields",
			data: struct {
				A *coverMarshalJSON `json:"a"`
				B *coverMarshalJSON `json:"b"`
				C *coverMarshalJSON `json:"c"`
			}{A: nil, B: nil, C: nil},
		},
		{
			name: "HeadMarshalJSONPtrNilMultiFieldsOmitEmpty",
			data: struct {
				A *coverMarshalJSON `json:"a,omitempty"`
				B *coverMarshalJSON `json:"b,omitempty"`
				C *coverMarshalJSON `json:"c,omitempty"`
			}{A: nil, B: nil, C: nil},
		},
		{
			name: "HeadMarshalJSONPtrNilMultiFieldsString",
			data: struct {
				A *coverMarshalJSON `json:"a,string"`
				B *coverMarshalJSON `json:"b,string"`
				C *coverMarshalJSON `json:"c,string"`
			}{A: nil, B: nil, C: nil},
		},
		{
			name: "HeadPtrMarshalJSONPtrNilMultiFields",
			data: struct {
				A *coverPtrMarshalJSON `json:"a"`
				B *coverPtrMarshalJSON `json:"b"`
				C *coverPtrMarshalJSON `json:"c"`
			}{A: nil, B: nil, C: nil},
		},
		{
			name: "HeadPtrMarshalJSONPtrNilMultiFieldsOmitEmpty",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
				B *coverPtrMarshalJSON `json:"b,omitempty"`
				C *coverPtrMarshalJSON `json:"c,omitempty"`
			}{A: nil, B: nil, C: nil},
		},
		{
			name: "HeadPtrMarshalJSONPtrNilMultiFieldsString",
			data: struct {
				A *coverPtrMarshalJSON `json:"a,string"`
				B *coverPtrMarshalJSON `json:"b,string"`
				C *coverPtrMarshalJSON `json:"c,string"`
			}{A: nil, B: nil, C: nil},
		},

		// PtrHeadMarshalJSONZeroMultiFields
		{
			name: "PtrHeadMarshalJSONZeroMultiFields",
			data: &struct {
				A coverMarshalJSON `json:"a"`
				B coverMarshalJSON `json:"b"`
			}{},
		},
		{
			name: "PtrHeadMarshalJSONZeroMultiFieldsOmitEmpty",
			data: &struct {
				A coverMarshalJSON `json:"a,omitempty"`
				B coverMarshalJSON `json:"b,omitempty"`
			}{},
		},
		{
			name: "PtrHeadMarshalJSONZeroMultiFieldsString",
			data: &struct {
				A coverMarshalJSON `json:"a,string"`
				B coverMarshalJSON `json:"b,string"`
			}{},
		},
		{
			name: "PtrHeadPtrMarshalJSONZeroMultiFields",
			data: &struct {
				A coverPtrMarshalJSON `json:"a"`
				B coverPtrMarshalJSON `json:"b"`
			}{},
		},
		{
			name: "PtrHeadPtrMarshalJSONZeroMultiFieldsOmitEmpty",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
				B coverPtrMarshalJSON `json:"b,omitempty"`
			}{},
		},
		{
			name: "PtrHeadPtrMarshalJSONZeroMultiFieldsString",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,string"`
				B coverPtrMarshalJSON `json:"b,string"`
			}{},
		},

		// PtrHeadMarshalJSONMultiFields
		{
			name: "PtrHeadMarshalJSONMultiFields",
			data: &struct {
				A coverMarshalJSON `json:"a"`
				B coverMarshalJSON `json:"b"`
			}{A: coverMarshalJSON{}, B: coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONMultiFieldsOmitEmpty",
			data: &struct {
				A coverMarshalJSON `json:"a,omitempty"`
				B coverMarshalJSON `json:"b,omitempty"`
			}{A: coverMarshalJSON{}, B: coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONMultiFieldsString",
			data: &struct {
				A coverMarshalJSON `json:"a,string"`
				B coverMarshalJSON `json:"b,string"`
			}{A: coverMarshalJSON{}, B: coverMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONMultiFields",
			data: &struct {
				A coverPtrMarshalJSON `json:"a"`
				B coverPtrMarshalJSON `json:"b"`
			}{A: coverPtrMarshalJSON{}, B: coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONMultiFieldsOmitEmpty",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
				B coverPtrMarshalJSON `json:"b,omitempty"`
			}{A: coverPtrMarshalJSON{}, B: coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONMultiFieldsString",
			data: &struct {
				A coverPtrMarshalJSON `json:"a,string"`
				B coverPtrMarshalJSON `json:"b,string"`
			}{A: coverPtrMarshalJSON{}, B: coverPtrMarshalJSON{}},
		},

		// PtrHeadMarshalJSONPtrMultiFields
		{
			name: "PtrHeadMarshalJSONPtrMultiFields",
			data: &struct {
				A *coverMarshalJSON `json:"a"`
				B *coverMarshalJSON `json:"b"`
			}{A: &coverMarshalJSON{}, B: &coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONPtrMultiFieldsOmitEmpty",
			data: &struct {
				A *coverMarshalJSON `json:"a,omitempty"`
				B *coverMarshalJSON `json:"b,omitempty"`
			}{A: &coverMarshalJSON{}, B: &coverMarshalJSON{}},
		},
		{
			name: "PtrHeadMarshalJSONPtrMultiFieldsString",
			data: &struct {
				A *coverMarshalJSON `json:"a,string"`
				B *coverMarshalJSON `json:"b,string"`
			}{A: &coverMarshalJSON{}, B: &coverMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrMultiFields",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a"`
				B *coverPtrMarshalJSON `json:"b"`
			}{A: &coverPtrMarshalJSON{}, B: &coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrMultiFieldsOmitEmpty",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
				B *coverPtrMarshalJSON `json:"b,omitempty"`
			}{A: &coverPtrMarshalJSON{}, B: &coverPtrMarshalJSON{}},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrMultiFieldsString",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,string"`
				B *coverPtrMarshalJSON `json:"b,string"`
			}{A: &coverPtrMarshalJSON{}, B: &coverPtrMarshalJSON{}},
		},

		// PtrHeadMarshalJSONPtrNilMultiFields
		{
			name: "PtrHeadMarshalJSONPtrNilMultiFields",
			data: &struct {
				A *coverMarshalJSON `json:"a"`
				B *coverMarshalJSON `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadMarshalJSONPtrNilMultiFieldsOmitEmpty",
			data: &struct {
				A *coverMarshalJSON `json:"a,omitempty"`
				B *coverMarshalJSON `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadMarshalJSONPtrNilMultiFieldsString",
			data: &struct {
				A *coverMarshalJSON `json:"a,string"`
				B *coverMarshalJSON `json:"b,string"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrNilMultiFields",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a"`
				B *coverPtrMarshalJSON `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrNilMultiFieldsOmitEmpty",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,omitempty"`
				B *coverPtrMarshalJSON `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadPtrMarshalJSONPtrNilMultiFieldsString",
			data: &struct {
				A *coverPtrMarshalJSON `json:"a,string"`
				B *coverPtrMarshalJSON `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadMarshalJSONNilMultiFields
		{
			name: "PtrHeadMarshalJSONNilMultiFields",
			data: (*struct {
				A coverMarshalJSON `json:"a"`
				B coverMarshalJSON `json:"b"`
			})(nil),
		},
		{
			name: "PtrHeadMarshalJSONNilMultiFieldsOmitEmpty",
			data: (*struct {
				A coverMarshalJSON `json:"a,omitempty"`
				B coverMarshalJSON `json:"b,omitempty"`
			})(nil),
		},
		{
			name: "PtrHeadMarshalJSONNilMultiFieldsString",
			data: (*struct {
				A coverMarshalJSON `json:"a,string"`
				B coverMarshalJSON `json:"b,string"`
			})(nil),
		},
		{
			name: "PtrHeadPtrMarshalJSONNilMultiFields",
			data: (*struct {
				A coverPtrMarshalJSON `json:"a"`
				B coverPtrMarshalJSON `json:"b"`
			})(nil),
		},
		{
			name: "PtrHeadPtrMarshalJSONNilMultiFieldsOmitEmpty",
			data: (*struct {
				A coverPtrMarshalJSON `json:"a,omitempty"`
				B coverPtrMarshalJSON `json:"b,omitempty"`
			})(nil),
		},
		{
			name: "PtrHeadPtrMarshalJSONNilMultiFieldsString",
			data: (*struct {
				A coverPtrMarshalJSON `json:"a,string"`
				B coverPtrMarshalJSON `json:"b,string"`
			})(nil),
		},

		/*

			// PtrHeadMarshalJSONNilMultiFields
			{
				name: "PtrHeadMarshalJSONNilMultiFields",
				data: (*struct {
					A *[2]int `json:"a"`
					B *[2]int `json:"b"`
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONNilMultiFieldsOmitEmpty",
				data: (*struct {
					A *[2]int `json:"a,omitempty"`
					B *[2]int `json:"b,omitempty"`
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONNilMultiFieldsString",
				data: (*struct {
					A *[2]int `json:"a,string"`
					B *[2]int `json:"b,string"`
				})(nil),
			},

			// HeadMarshalJSONZeroNotRoot
			{
				name: "HeadMarshalJSONZeroNotRoot",
				data: struct {
					A struct {
						A [2]int `json:"a"`
					}
				}{},
			},
			{
				name: "HeadMarshalJSONZeroNotRootOmitEmpty",
				data: struct {
					A struct {
						A [2]int `json:"a,omitempty"`
					}
				}{},
			},
			{
				name: "HeadMarshalJSONZeroNotRootString",
				data: struct {
					A struct {
						A [2]int `json:"a,string"`
					}
				}{},
			},

			// HeadMarshalJSONNotRoot
			{
				name: "HeadMarshalJSONNotRoot",
				data: struct {
					A struct {
						A [2]int `json:"a"`
					}
				}{A: struct {
					A [2]int `json:"a"`
				}{A: [2]int{-1}}},
			},
			{
				name: "HeadMarshalJSONNotRootOmitEmpty",
				data: struct {
					A struct {
						A [2]int `json:"a,omitempty"`
					}
				}{A: struct {
					A [2]int `json:"a,omitempty"`
				}{A: [2]int{-1}}},
			},
			{
				name: "HeadMarshalJSONNotRootString",
				data: struct {
					A struct {
						A [2]int `json:"a,string"`
					}
				}{A: struct {
					A [2]int `json:"a,string"`
				}{A: [2]int{-1}}},
			},

			// HeadMarshalJSONPtrNotRoot
			{
				name: "HeadMarshalJSONPtrNotRoot",
				data: struct {
					A struct {
						A *[2]int `json:"a"`
					}
				}{A: struct {
					A *[2]int `json:"a"`
				}{arrayptr([2]int{-1})}},
			},
			{
				name: "HeadMarshalJSONPtrNotRootOmitEmpty",
				data: struct {
					A struct {
						A *[2]int `json:"a,omitempty"`
					}
				}{A: struct {
					A *[2]int `json:"a,omitempty"`
				}{arrayptr([2]int{-1})}},
			},
			{
				name: "HeadMarshalJSONPtrNotRootString",
				data: struct {
					A struct {
						A *[2]int `json:"a,string"`
					}
				}{A: struct {
					A *[2]int `json:"a,string"`
				}{arrayptr([2]int{-1})}},
			},

			// HeadMarshalJSONPtrNilNotRoot
			{
				name: "HeadMarshalJSONPtrNilNotRoot",
				data: struct {
					A struct {
						A *[2]int `json:"a"`
					}
				}{},
			},
			{
				name: "HeadMarshalJSONPtrNilNotRootOmitEmpty",
				data: struct {
					A struct {
						A *[2]int `json:"a,omitempty"`
					}
				}{},
			},
			{
				name: "HeadMarshalJSONPtrNilNotRootString",
				data: struct {
					A struct {
						A *[2]int `json:"a,string"`
					}
				}{},
			},

			// PtrHeadMarshalJSONZeroNotRoot
			{
				name: "PtrHeadMarshalJSONZeroNotRoot",
				data: struct {
					A *struct {
						A [2]int `json:"a"`
					}
				}{A: new(struct {
					A [2]int `json:"a"`
				})},
			},
			{
				name: "PtrHeadMarshalJSONZeroNotRootOmitEmpty",
				data: struct {
					A *struct {
						A [2]int `json:"a,omitempty"`
					}
				}{A: new(struct {
					A [2]int `json:"a,omitempty"`
				})},
			},
			{
				name: "PtrHeadMarshalJSONZeroNotRootString",
				data: struct {
					A *struct {
						A [2]int `json:"a,string"`
					}
				}{A: new(struct {
					A [2]int `json:"a,string"`
				})},
			},

			// PtrHeadMarshalJSONNotRoot
			{
				name: "PtrHeadMarshalJSONNotRoot",
				data: struct {
					A *struct {
						A [2]int `json:"a"`
					}
				}{A: &(struct {
					A [2]int `json:"a"`
				}{A: [2]int{-1}})},
			},
			{
				name: "PtrHeadMarshalJSONNotRootOmitEmpty",
				data: struct {
					A *struct {
						A [2]int `json:"a,omitempty"`
					}
				}{A: &(struct {
					A [2]int `json:"a,omitempty"`
				}{A: [2]int{-1}})},
			},
			{
				name: "PtrHeadMarshalJSONNotRootString",
				data: struct {
					A *struct {
						A [2]int `json:"a,string"`
					}
				}{A: &(struct {
					A [2]int `json:"a,string"`
				}{A: [2]int{-1}})},
			},

			// PtrHeadMarshalJSONPtrNotRoot
			{
				name: "PtrHeadMarshalJSONPtrNotRoot",
				data: struct {
					A *struct {
						A *[2]int `json:"a"`
					}
				}{A: &(struct {
					A *[2]int `json:"a"`
				}{A: arrayptr([2]int{-1})})},
			},
			{
				name: "PtrHeadMarshalJSONPtrNotRootOmitEmpty",
				data: struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,omitempty"`
				}{A: arrayptr([2]int{-1})})},
			},
			{
				name: "PtrHeadMarshalJSONPtrNotRootString",
				data: struct {
					A *struct {
						A *[2]int `json:"a,string"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,string"`
				}{A: arrayptr([2]int{-1})})},
			},

			// PtrHeadMarshalJSONPtrNilNotRoot
			{
				name: "PtrHeadMarshalJSONPtrNilNotRoot",
				data: struct {
					A *struct {
						A *[2]int `json:"a"`
					}
				}{A: &(struct {
					A *[2]int `json:"a"`
				}{A: nil})},
			},
			{
				name: "PtrHeadMarshalJSONPtrNilNotRootOmitEmpty",
				data: struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,omitempty"`
				}{A: nil})},
			},
			{
				name: "PtrHeadMarshalJSONPtrNilNotRootString",
				data: struct {
					A *struct {
						A *[2]int `json:"a,string"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,string"`
				}{A: nil})},
			},

			// PtrHeadMarshalJSONNilNotRoot
			{
				name: "PtrHeadMarshalJSONNilNotRoot",
				data: struct {
					A *struct {
						A *[2]int `json:"a"`
					}
				}{A: nil},
			},
			{
				name: "PtrHeadMarshalJSONNilNotRootOmitEmpty",
				data: struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
					} `json:",omitempty"`
				}{A: nil},
			},
			{
				name: "PtrHeadMarshalJSONNilNotRootString",
				data: struct {
					A *struct {
						A *[2]int `json:"a,string"`
					} `json:",string"`
				}{A: nil},
			},

			// HeadMarshalJSONZeroMultiFieldsNotRoot
			{
				name: "HeadMarshalJSONZeroMultiFieldsNotRoot",
				data: struct {
					A struct {
						A [2]int `json:"a"`
					}
					B struct {
						B [2]int `json:"b"`
					}
				}{},
			},
			{
				name: "HeadMarshalJSONZeroMultiFieldsNotRootOmitEmpty",
				data: struct {
					A struct {
						A [2]int `json:"a,omitempty"`
					}
					B struct {
						B [2]int `json:"b,omitempty"`
					}
				}{},
			},
			{
				name: "HeadMarshalJSONZeroMultiFieldsNotRootString",
				data: struct {
					A struct {
						A [2]int `json:"a,string"`
					}
					B struct {
						B [2]int `json:"b,string"`
					}
				}{},
			},

			// HeadMarshalJSONMultiFieldsNotRoot
			{
				name: "HeadMarshalJSONMultiFieldsNotRoot",
				data: struct {
					A struct {
						A [2]int `json:"a"`
					}
					B struct {
						B [2]int `json:"b"`
					}
				}{A: struct {
					A [2]int `json:"a"`
				}{A: [2]int{-1}}, B: struct {
					B [2]int `json:"b"`
				}{B: [2]int{0}}},
			},
			{
				name: "HeadMarshalJSONMultiFieldsNotRootOmitEmpty",
				data: struct {
					A struct {
						A [2]int `json:"a,omitempty"`
					}
					B struct {
						B [2]int `json:"b,omitempty"`
					}
				}{A: struct {
					A [2]int `json:"a,omitempty"`
				}{A: [2]int{-1}}, B: struct {
					B [2]int `json:"b,omitempty"`
				}{B: [2]int{1}}},
			},
			{
				name: "HeadMarshalJSONMultiFieldsNotRootString",
				data: struct {
					A struct {
						A [2]int `json:"a,string"`
					}
					B struct {
						B [2]int `json:"b,string"`
					}
				}{A: struct {
					A [2]int `json:"a,string"`
				}{A: [2]int{-1}}, B: struct {
					B [2]int `json:"b,string"`
				}{B: [2]int{1}}},
			},

			// HeadMarshalJSONPtrMultiFieldsNotRoot
			{
				name: "HeadMarshalJSONPtrMultiFieldsNotRoot",
				data: struct {
					A struct {
						A *[2]int `json:"a"`
					}
					B struct {
						B *[2]int `json:"b"`
					}
				}{A: struct {
					A *[2]int `json:"a"`
				}{A: arrayptr([2]int{-1})}, B: struct {
					B *[2]int `json:"b"`
				}{B: arrayptr([2]int{1})}},
			},
			{
				name: "HeadMarshalJSONPtrMultiFieldsNotRootOmitEmpty",
				data: struct {
					A struct {
						A *[2]int `json:"a,omitempty"`
					}
					B struct {
						B *[2]int `json:"b,omitempty"`
					}
				}{A: struct {
					A *[2]int `json:"a,omitempty"`
				}{A: arrayptr([2]int{-1})}, B: struct {
					B *[2]int `json:"b,omitempty"`
				}{B: arrayptr([2]int{1})}},
			},
			{
				name: "HeadMarshalJSONPtrMultiFieldsNotRootString",
				data: struct {
					A struct {
						A *[2]int `json:"a,string"`
					}
					B struct {
						B *[2]int `json:"b,string"`
					}
				}{A: struct {
					A *[2]int `json:"a,string"`
				}{A: arrayptr([2]int{-1})}, B: struct {
					B *[2]int `json:"b,string"`
				}{B: arrayptr([2]int{1})}},
			},

			// HeadMarshalJSONPtrNilMultiFieldsNotRoot
			{
				name: "HeadMarshalJSONPtrNilMultiFieldsNotRoot",
				data: struct {
					A struct {
						A *[2]int `json:"a"`
					}
					B struct {
						B *[2]int `json:"b"`
					}
				}{A: struct {
					A *[2]int `json:"a"`
				}{A: nil}, B: struct {
					B *[2]int `json:"b"`
				}{B: nil}},
			},
			{
				name: "HeadMarshalJSONPtrNilMultiFieldsNotRootOmitEmpty",
				data: struct {
					A struct {
						A *[2]int `json:"a,omitempty"`
					}
					B struct {
						B *[2]int `json:"b,omitempty"`
					}
				}{A: struct {
					A *[2]int `json:"a,omitempty"`
				}{A: nil}, B: struct {
					B *[2]int `json:"b,omitempty"`
				}{B: nil}},
			},
			{
				name: "HeadMarshalJSONPtrNilMultiFieldsNotRootString",
				data: struct {
					A struct {
						A *[2]int `json:"a,string"`
					}
					B struct {
						B *[2]int `json:"b,string"`
					}
				}{A: struct {
					A *[2]int `json:"a,string"`
				}{A: nil}, B: struct {
					B *[2]int `json:"b,string"`
				}{B: nil}},
			},

			// PtrHeadMarshalJSONZeroMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONZeroMultiFieldsNotRoot",
				data: &struct {
					A struct {
						A [2]int `json:"a"`
					}
					B struct {
						B [2]int `json:"b"`
					}
				}{},
			},
			{
				name: "PtrHeadMarshalJSONZeroMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A struct {
						A [2]int `json:"a,omitempty"`
					}
					B struct {
						B [2]int `json:"b,omitempty"`
					}
				}{},
			},
			{
				name: "PtrHeadMarshalJSONZeroMultiFieldsNotRootString",
				data: &struct {
					A struct {
						A [2]int `json:"a,string"`
					}
					B struct {
						B [2]int `json:"b,string"`
					}
				}{},
			},

			// PtrHeadMarshalJSONMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONMultiFieldsNotRoot",
				data: &struct {
					A struct {
						A [2]int `json:"a"`
					}
					B struct {
						B [2]int `json:"b"`
					}
				}{A: struct {
					A [2]int `json:"a"`
				}{A: [2]int{-1}}, B: struct {
					B [2]int `json:"b"`
				}{B: [2]int{1}}},
			},
			{
				name: "PtrHeadMarshalJSONMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A struct {
						A [2]int `json:"a,omitempty"`
					}
					B struct {
						B [2]int `json:"b,omitempty"`
					}
				}{A: struct {
					A [2]int `json:"a,omitempty"`
				}{A: [2]int{-1}}, B: struct {
					B [2]int `json:"b,omitempty"`
				}{B: [2]int{1}}},
			},
			{
				name: "PtrHeadMarshalJSONMultiFieldsNotRootString",
				data: &struct {
					A struct {
						A [2]int `json:"a,string"`
					}
					B struct {
						B [2]int `json:"b,string"`
					}
				}{A: struct {
					A [2]int `json:"a,string"`
				}{A: [2]int{-1}}, B: struct {
					B [2]int `json:"b,string"`
				}{B: [2]int{1}}},
			},

			// PtrHeadMarshalJSONPtrMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONPtrMultiFieldsNotRoot",
				data: &struct {
					A *struct {
						A *[2]int `json:"a"`
					}
					B *struct {
						B *[2]int `json:"b"`
					}
				}{A: &(struct {
					A *[2]int `json:"a"`
				}{A: arrayptr([2]int{-1})}), B: &(struct {
					B *[2]int `json:"b"`
				}{B: arrayptr([2]int{1})})},
			},
			{
				name: "PtrHeadMarshalJSONPtrMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
					}
					B *struct {
						B *[2]int `json:"b,omitempty"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,omitempty"`
				}{A: arrayptr([2]int{-1})}), B: &(struct {
					B *[2]int `json:"b,omitempty"`
				}{B: arrayptr([2]int{1})})},
			},
			{
				name: "PtrHeadMarshalJSONPtrMultiFieldsNotRootString",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,string"`
					}
					B *struct {
						B *[2]int `json:"b,string"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,string"`
				}{A: arrayptr([2]int{-1})}), B: &(struct {
					B *[2]int `json:"b,string"`
				}{B: arrayptr([2]int{1})})},
			},

			// PtrHeadMarshalJSONPtrNilMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONPtrNilMultiFieldsNotRoot",
				data: &struct {
					A *struct {
						A *[2]int `json:"a"`
					}
					B *struct {
						B *[2]int `json:"b"`
					}
				}{A: nil, B: nil},
			},
			{
				name: "PtrHeadMarshalJSONPtrNilMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
					} `json:",omitempty"`
					B *struct {
						B *[2]int `json:"b,omitempty"`
					} `json:",omitempty"`
				}{A: nil, B: nil},
			},
			{
				name: "PtrHeadMarshalJSONPtrNilMultiFieldsNotRootString",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,string"`
					} `json:",string"`
					B *struct {
						B *[2]int `json:"b,string"`
					} `json:",string"`
				}{A: nil, B: nil},
			},

			// PtrHeadMarshalJSONNilMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONNilMultiFieldsNotRoot",
				data: (*struct {
					A *struct {
						A *[2]int `json:"a"`
					}
					B *struct {
						B *[2]int `json:"b"`
					}
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONNilMultiFieldsNotRootOmitEmpty",
				data: (*struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
					}
					B *struct {
						B *[2]int `json:"b,omitempty"`
					}
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONNilMultiFieldsNotRootString",
				data: (*struct {
					A *struct {
						A *[2]int `json:"a,string"`
					}
					B *struct {
						B *[2]int `json:"b,string"`
					}
				})(nil),
			},

			// PtrHeadMarshalJSONDoubleMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONDoubleMultiFieldsNotRoot",
				data: &struct {
					A *struct {
						A [2]int `json:"a"`
						B [2]int `json:"b"`
					}
					B *struct {
						A [2]int `json:"a"`
						B [2]int `json:"b"`
					}
				}{A: &(struct {
					A [2]int `json:"a"`
					B [2]int `json:"b"`
				}{A: [2]int{-1}, B: [2]int{1}}), B: &(struct {
					A [2]int `json:"a"`
					B [2]int `json:"b"`
				}{A: [2]int{-1}, B: [2]int{1}})},
			},
			{
				name: "PtrHeadMarshalJSONDoubleMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A *struct {
						A [2]int `json:"a,omitempty"`
						B [2]int `json:"b,omitempty"`
					}
					B *struct {
						A [2]int `json:"a,omitempty"`
						B [2]int `json:"b,omitempty"`
					}
				}{A: &(struct {
					A [2]int `json:"a,omitempty"`
					B [2]int `json:"b,omitempty"`
				}{A: [2]int{-1}, B: [2]int{1}}), B: &(struct {
					A [2]int `json:"a,omitempty"`
					B [2]int `json:"b,omitempty"`
				}{A: [2]int{-1}, B: [2]int{1}})},
			},
			{
				name: "PtrHeadMarshalJSONDoubleMultiFieldsNotRootString",
				data: &struct {
					A *struct {
						A [2]int `json:"a,string"`
						B [2]int `json:"b,string"`
					}
					B *struct {
						A [2]int `json:"a,string"`
						B [2]int `json:"b,string"`
					}
				}{A: &(struct {
					A [2]int `json:"a,string"`
					B [2]int `json:"b,string"`
				}{A: [2]int{-1}, B: [2]int{1}}), B: &(struct {
					A [2]int `json:"a,string"`
					B [2]int `json:"b,string"`
				}{A: [2]int{-1}, B: [2]int{1}})},
			},

			// PtrHeadMarshalJSONNilDoubleMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONNilDoubleMultiFieldsNotRoot",
				data: &struct {
					A *struct {
						A [2]int `json:"a"`
						B [2]int `json:"b"`
					}
					B *struct {
						A [2]int `json:"a"`
						B [2]int `json:"b"`
					}
				}{A: nil, B: nil},
			},
			{
				name: "PtrHeadMarshalJSONNilDoubleMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A *struct {
						A [2]int `json:"a,omitempty"`
						B [2]int `json:"b,omitempty"`
					} `json:",omitempty"`
					B *struct {
						A [2]int `json:"a,omitempty"`
						B [2]int `json:"b,omitempty"`
					} `json:",omitempty"`
				}{A: nil, B: nil},
			},
			{
				name: "PtrHeadMarshalJSONNilDoubleMultiFieldsNotRootString",
				data: &struct {
					A *struct {
						A [2]int `json:"a,string"`
						B [2]int `json:"b,string"`
					}
					B *struct {
						A [2]int `json:"a,string"`
						B [2]int `json:"b,string"`
					}
				}{A: nil, B: nil},
			},

			// PtrHeadMarshalJSONNilDoubleMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONNilDoubleMultiFieldsNotRoot",
				data: (*struct {
					A *struct {
						A [2]int `json:"a"`
						B [2]int `json:"b"`
					}
					B *struct {
						A [2]int `json:"a"`
						B [2]int `json:"b"`
					}
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONNilDoubleMultiFieldsNotRootOmitEmpty",
				data: (*struct {
					A *struct {
						A [2]int `json:"a,omitempty"`
						B [2]int `json:"b,omitempty"`
					}
					B *struct {
						A [2]int `json:"a,omitempty"`
						B [2]int `json:"b,omitempty"`
					}
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONNilDoubleMultiFieldsNotRootString",
				data: (*struct {
					A *struct {
						A [2]int `json:"a,string"`
						B [2]int `json:"b,string"`
					}
					B *struct {
						A [2]int `json:"a,string"`
						B [2]int `json:"b,string"`
					}
				})(nil),
			},

			// PtrHeadMarshalJSONPtrDoubleMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONPtrDoubleMultiFieldsNotRoot",
				data: &struct {
					A *struct {
						A *[2]int `json:"a"`
						B *[2]int `json:"b"`
					}
					B *struct {
						A *[2]int `json:"a"`
						B *[2]int `json:"b"`
					}
				}{A: &(struct {
					A *[2]int `json:"a"`
					B *[2]int `json:"b"`
				}{A: arrayptr([2]int{-1}), B: arrayptr([2]int{1})}), B: &(struct {
					A *[2]int `json:"a"`
					B *[2]int `json:"b"`
				}{A: nil, B: nil})},
			},
			{
				name: "PtrHeadMarshalJSONPtrDoubleMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
						B *[2]int `json:"b,omitempty"`
					}
					B *struct {
						A *[2]int `json:"a,omitempty"`
						B *[2]int `json:"b,omitempty"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,omitempty"`
					B *[2]int `json:"b,omitempty"`
				}{A: arrayptr([2]int{-1}), B: arrayptr([2]int{1})}), B: &(struct {
					A *[2]int `json:"a,omitempty"`
					B *[2]int `json:"b,omitempty"`
				}{A: nil, B: nil})},
			},
			{
				name: "PtrHeadMarshalJSONPtrDoubleMultiFieldsNotRootString",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,string"`
						B *[2]int `json:"b,string"`
					}
					B *struct {
						A *[2]int `json:"a,string"`
						B *[2]int `json:"b,string"`
					}
				}{A: &(struct {
					A *[2]int `json:"a,string"`
					B *[2]int `json:"b,string"`
				}{A: arrayptr([2]int{-1}), B: arrayptr([2]int{1})}), B: &(struct {
					A *[2]int `json:"a,string"`
					B *[2]int `json:"b,string"`
				}{A: nil, B: nil})},
			},

			// PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRoot",
				data: &struct {
					A *struct {
						A *[2]int `json:"a"`
						B *[2]int `json:"b"`
					}
					B *struct {
						A *[2]int `json:"a"`
						B *[2]int `json:"b"`
					}
				}{A: nil, B: nil},
			},
			{
				name: "PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRootOmitEmpty",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
						B *[2]int `json:"b,omitempty"`
					} `json:",omitempty"`
					B *struct {
						A *[2]int `json:"a,omitempty"`
						B *[2]int `json:"b,omitempty"`
					} `json:",omitempty"`
				}{A: nil, B: nil},
			},
			{
				name: "PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRootString",
				data: &struct {
					A *struct {
						A *[2]int `json:"a,string"`
						B *[2]int `json:"b,string"`
					}
					B *struct {
						A *[2]int `json:"a,string"`
						B *[2]int `json:"b,string"`
					}
				}{A: nil, B: nil},
			},

			// PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRoot
			{
				name: "PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRoot",
				data: (*struct {
					A *struct {
						A *[2]int `json:"a"`
						B *[2]int `json:"b"`
					}
					B *struct {
						A *[2]int `json:"a"`
						B *[2]int `json:"b"`
					}
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRootOmitEmpty",
				data: (*struct {
					A *struct {
						A *[2]int `json:"a,omitempty"`
						B *[2]int `json:"b,omitempty"`
					}
					B *struct {
						A *[2]int `json:"a,omitempty"`
						B *[2]int `json:"b,omitempty"`
					}
				})(nil),
			},
			{
				name: "PtrHeadMarshalJSONPtrNilDoubleMultiFieldsNotRootString",
				data: (*struct {
					A *struct {
						A *[2]int `json:"a,string"`
						B *[2]int `json:"b,string"`
					}
					B *struct {
						A *[2]int `json:"a,string"`
						B *[2]int `json:"b,string"`
					}
				})(nil),
			},

			// AnonymousHeadMarshalJSON
			{
				name: "AnonymousHeadMarshalJSON",
				data: struct {
					structMarshalJSON
					B [2]int `json:"b"`
				}{
					structMarshalJSON: structMarshalJSON{A: [2]int{-1}},
					B:                 [2]int{1},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONOmitEmpty",
				data: struct {
					structMarshalJSONOmitEmpty
					B [2]int `json:"b,omitempty"`
				}{
					structMarshalJSONOmitEmpty: structMarshalJSONOmitEmpty{A: [2]int{-1}},
					B:                          [2]int{1},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONString",
				data: struct {
					structMarshalJSONString
					B [2]int `json:"b,string"`
				}{
					structMarshalJSONString: structMarshalJSONString{A: [2]int{-1}},
					B:                       [2]int{1},
				},
			},

			// PtrAnonymousHeadMarshalJSON
			{
				name: "PtrAnonymousHeadMarshalJSON",
				data: struct {
					*structMarshalJSON
					B [2]int `json:"b"`
				}{
					structMarshalJSON: &structMarshalJSON{A: [2]int{-1}},
					B:                 [2]int{1},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONOmitEmpty",
				data: struct {
					*structMarshalJSONOmitEmpty
					B [2]int `json:"b,omitempty"`
				}{
					structMarshalJSONOmitEmpty: &structMarshalJSONOmitEmpty{A: [2]int{-1}},
					B:                          [2]int{1},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONString",
				data: struct {
					*structMarshalJSONString
					B [2]int `json:"b,string"`
				}{
					structMarshalJSONString: &structMarshalJSONString{A: [2]int{-1}},
					B:                       [2]int{1},
				},
			},

			// PtrAnonymousHeadMarshalJSONNil
			{
				name: "PtrAnonymousHeadMarshalJSONNil",
				data: struct {
					*structMarshalJSON
					B [2]int `json:"b"`
				}{
					structMarshalJSON: &structMarshalJSON{A: [2]int{1}},
					B:                 [2]int{1},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONNilOmitEmpty",
				data: struct {
					*structMarshalJSONOmitEmpty
					B [2]int `json:"b,omitempty"`
				}{
					structMarshalJSONOmitEmpty: &structMarshalJSONOmitEmpty{A: [2]int{1}},
					B:                          [2]int{1},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONNilString",
				data: struct {
					*structMarshalJSONString
					B [2]int `json:"b,string"`
				}{
					structMarshalJSONString: &structMarshalJSONString{A: [2]int{1}},
					B:                       [2]int{1},
				},
			},

			// NilPtrAnonymousHeadMarshalJSON
			{
				name: "NilPtrAnonymousHeadMarshalJSON",
				data: struct {
					*structMarshalJSON
					B [2]int `json:"b"`
				}{
					structMarshalJSON: nil,
					B:                 [2]int{-1},
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONOmitEmpty",
				data: struct {
					*structMarshalJSONOmitEmpty
					B [2]int `json:"b,omitempty"`
				}{
					structMarshalJSONOmitEmpty: nil,
					B:                          [2]int{-1},
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONString",
				data: struct {
					*structMarshalJSONString
					B [2]int `json:"b,string"`
				}{
					structMarshalJSONString: nil,
					B:                       [2]int{-1},
				},
			},

			// AnonymousHeadMarshalJSONPtr
			{
				name: "AnonymousHeadMarshalJSONPtr",
				data: struct {
					structMarshalJSONPtr
					B *[2]int `json:"b"`
				}{
					structMarshalJSONPtr: structMarshalJSONPtr{A: arrayptr([2]int{-1})},
					B:                    nil,
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrOmitEmpty",
				data: struct {
					structMarshalJSONPtrOmitEmpty
					B *[2]int `json:"b,omitempty"`
				}{
					structMarshalJSONPtrOmitEmpty: structMarshalJSONPtrOmitEmpty{A: arrayptr([2]int{-1})},
					B:                             nil,
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrString",
				data: struct {
					structMarshalJSONPtrString
					B *[2]int `json:"b,string"`
				}{
					structMarshalJSONPtrString: structMarshalJSONPtrString{A: arrayptr([2]int{-1})},
					B:                          nil,
				},
			},

			// AnonymousHeadMarshalJSONPtrNil
			{
				name: "AnonymousHeadMarshalJSONPtrNil",
				data: struct {
					structMarshalJSONPtr
					B *[2]int `json:"b"`
				}{
					structMarshalJSONPtr: structMarshalJSONPtr{A: nil},
					B:                    arrayptr([2]int{-1}),
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrNilOmitEmpty",
				data: struct {
					structMarshalJSONPtrOmitEmpty
					B *[2]int `json:"b,omitempty"`
				}{
					structMarshalJSONPtrOmitEmpty: structMarshalJSONPtrOmitEmpty{A: nil},
					B:                             arrayptr([2]int{-1}),
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrNilString",
				data: struct {
					structMarshalJSONPtrString
					B *[2]int `json:"b,string"`
				}{
					structMarshalJSONPtrString: structMarshalJSONPtrString{A: nil},
					B:                          arrayptr([2]int{-1}),
				},
			},

			// PtrAnonymousHeadMarshalJSONPtr
			{
				name: "PtrAnonymousHeadMarshalJSONPtr",
				data: struct {
					*structMarshalJSONPtr
					B *[2]int `json:"b"`
				}{
					structMarshalJSONPtr: &structMarshalJSONPtr{A: arrayptr([2]int{-1})},
					B:                    nil,
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONPtrOmitEmpty",
				data: struct {
					*structMarshalJSONPtrOmitEmpty
					B *[2]int `json:"b,omitempty"`
				}{
					structMarshalJSONPtrOmitEmpty: &structMarshalJSONPtrOmitEmpty{A: arrayptr([2]int{-1})},
					B:                             nil,
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONPtrString",
				data: struct {
					*structMarshalJSONPtrString
					B *[2]int `json:"b,string"`
				}{
					structMarshalJSONPtrString: &structMarshalJSONPtrString{A: arrayptr([2]int{-1})},
					B:                          nil,
				},
			},

			// NilPtrAnonymousHeadMarshalJSONPtr
			{
				name: "NilPtrAnonymousHeadMarshalJSONPtr",
				data: struct {
					*structMarshalJSONPtr
					B *[2]int `json:"b"`
				}{
					structMarshalJSONPtr: nil,
					B:                    arrayptr([2]int{-1}),
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONPtrOmitEmpty",
				data: struct {
					*structMarshalJSONPtrOmitEmpty
					B *[2]int `json:"b,omitempty"`
				}{
					structMarshalJSONPtrOmitEmpty: nil,
					B:                             arrayptr([2]int{-1}),
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONPtrString",
				data: struct {
					*structMarshalJSONPtrString
					B *[2]int `json:"b,string"`
				}{
					structMarshalJSONPtrString: nil,
					B:                          arrayptr([2]int{-1}),
				},
			},

			// AnonymousHeadMarshalJSONOnly
			{
				name: "AnonymousHeadMarshalJSONOnly",
				data: struct {
					structMarshalJSON
				}{
					structMarshalJSON: structMarshalJSON{A: [2]int{-1}},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONOnlyOmitEmpty",
				data: struct {
					structMarshalJSONOmitEmpty
				}{
					structMarshalJSONOmitEmpty: structMarshalJSONOmitEmpty{A: [2]int{-1}},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONOnlyString",
				data: struct {
					structMarshalJSONString
				}{
					structMarshalJSONString: structMarshalJSONString{A: [2]int{-1}},
				},
			},

			// PtrAnonymousHeadMarshalJSONOnly
			{
				name: "PtrAnonymousHeadMarshalJSONOnly",
				data: struct {
					*structMarshalJSON
				}{
					structMarshalJSON: &structMarshalJSON{A: [2]int{-1}},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONOnlyOmitEmpty",
				data: struct {
					*structMarshalJSONOmitEmpty
				}{
					structMarshalJSONOmitEmpty: &structMarshalJSONOmitEmpty{A: [2]int{-1}},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONOnlyString",
				data: struct {
					*structMarshalJSONString
				}{
					structMarshalJSONString: &structMarshalJSONString{A: [2]int{-1}},
				},
			},

			// NilPtrAnonymousHeadMarshalJSONOnly
			{
				name: "NilPtrAnonymousHeadMarshalJSONOnly",
				data: struct {
					*structMarshalJSON
				}{
					structMarshalJSON: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONOnlyOmitEmpty",
				data: struct {
					*structMarshalJSONOmitEmpty
				}{
					structMarshalJSONOmitEmpty: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONOnlyString",
				data: struct {
					*structMarshalJSONString
				}{
					structMarshalJSONString: nil,
				},
			},

			// AnonymousHeadMarshalJSONPtrOnly
			{
				name: "AnonymousHeadMarshalJSONPtrOnly",
				data: struct {
					structMarshalJSONPtr
				}{
					structMarshalJSONPtr: structMarshalJSONPtr{A: arrayptr([2]int{-1})},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrOnlyOmitEmpty",
				data: struct {
					structMarshalJSONPtrOmitEmpty
				}{
					structMarshalJSONPtrOmitEmpty: structMarshalJSONPtrOmitEmpty{A: arrayptr([2]int{-1})},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrOnlyString",
				data: struct {
					structMarshalJSONPtrString
				}{
					structMarshalJSONPtrString: structMarshalJSONPtrString{A: arrayptr([2]int{-1})},
				},
			},

			// AnonymousHeadMarshalJSONPtrNilOnly
			{
				name: "AnonymousHeadMarshalJSONPtrNilOnly",
				data: struct {
					structMarshalJSONPtr
				}{
					structMarshalJSONPtr: structMarshalJSONPtr{A: nil},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrNilOnlyOmitEmpty",
				data: struct {
					structMarshalJSONPtrOmitEmpty
				}{
					structMarshalJSONPtrOmitEmpty: structMarshalJSONPtrOmitEmpty{A: nil},
				},
			},
			{
				name: "AnonymousHeadMarshalJSONPtrNilOnlyString",
				data: struct {
					structMarshalJSONPtrString
				}{
					structMarshalJSONPtrString: structMarshalJSONPtrString{A: nil},
				},
			},

			// PtrAnonymousHeadMarshalJSONPtrOnly
			{
				name: "PtrAnonymousHeadMarshalJSONPtrOnly",
				data: struct {
					*structMarshalJSONPtr
				}{
					structMarshalJSONPtr: &structMarshalJSONPtr{A: arrayptr([2]int{-1})},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONPtrOnlyOmitEmpty",
				data: struct {
					*structMarshalJSONPtrOmitEmpty
				}{
					structMarshalJSONPtrOmitEmpty: &structMarshalJSONPtrOmitEmpty{A: arrayptr([2]int{-1})},
				},
			},
			{
				name: "PtrAnonymousHeadMarshalJSONPtrOnlyString",
				data: struct {
					*structMarshalJSONPtrString
				}{
					structMarshalJSONPtrString: &structMarshalJSONPtrString{A: arrayptr([2]int{-1})},
				},
			},

			// NilPtrAnonymousHeadMarshalJSONPtrOnly
			{
				name: "NilPtrAnonymousHeadMarshalJSONPtrOnly",
				data: struct {
					*structMarshalJSONPtr
				}{
					structMarshalJSONPtr: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONPtrOnlyOmitEmpty",
				data: struct {
					*structMarshalJSONPtrOmitEmpty
				}{
					structMarshalJSONPtrOmitEmpty: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadMarshalJSONPtrOnlyString",
				data: struct {
					*structMarshalJSONPtrString
				}{
					structMarshalJSONPtrString: nil,
				},
			},
		*/
	}
	for _, test := range tests {
		for _, indent := range []bool{false} {
			for _, htmlEscape := range []bool{false} {
				fmt.Println(test.name)
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetEscapeHTML(htmlEscape)
				if indent {
					enc.SetIndent("", "  ")
				}
				if err := enc.Encode(test.data); err != nil {
					t.Fatalf("%s(htmlEscape:%v,indent:%v): %+v: %s", test.name, htmlEscape, indent, test.data, err)
				}
				stdresult := encodeByEncodingJSON(test.data, indent, htmlEscape)
				if buf.String() != stdresult {
					t.Errorf("%s(htmlEscape:%v,indent:%v): doesn't compatible with encoding/json. expected %q but got %q", test.name, htmlEscape, indent, stdresult, buf.String())
				}
			}
		}
	}
}
