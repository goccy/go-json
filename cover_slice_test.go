package json_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverSlice(t *testing.T) {
	type structSlice struct {
		A []int `json:"a"`
	}
	type structSliceOmitEmpty struct {
		A []int `json:"a,omitempty"`
	}
	type structSliceString struct {
		A []int `json:"a,string"`
	}
	type structSlicePtr struct {
		A *[]int `json:"a"`
	}
	type structSlicePtrOmitEmpty struct {
		A *[]int `json:"a,omitempty"`
	}
	type structSlicePtrString struct {
		A *[]int `json:"a,string"`
	}

	type structSlicePtrContent struct {
		A []*int `json:"a"`
	}
	type structSliceOmitEmptyPtrContent struct {
		A []*int `json:"a,omitempty"`
	}
	type structSliceStringPtrContent struct {
		A []*int `json:"a,string"`
	}
	type structSlicePtrPtrContent struct {
		A *[]*int `json:"a"`
	}
	type structSlicePtrOmitEmptyPtrContent struct {
		A *[]*int `json:"a,omitempty"`
	}
	type structSlicePtrStringPtrContent struct {
		A *[]*int `json:"a,string"`
	}

	tests := []struct {
		name string
		data interface{}
	}{
		// HeadSliceZero
		{
			name: "HeadSliceZero",
			data: struct {
				A []int `json:"a"`
			}{},
		},
		{
			name: "HeadSliceZeroOmitEmpty",
			data: struct {
				A []int `json:"a,omitempty"`
			}{},
		},
		{
			name: "HeadSliceZeroString",
			data: struct {
				A []int `json:"a,string"`
			}{},
		},

		// HeadSlice
		{
			name: "HeadSlice",
			data: struct {
				A []int `json:"a"`
			}{A: []int{-1}},
		},
		{
			name: "HeadSliceOmitEmpty",
			data: struct {
				A []int `json:"a,omitempty"`
			}{A: []int{-1}},
		},
		{
			name: "HeadSliceString",
			data: struct {
				A []int `json:"a,string"`
			}{A: []int{-1}},
		},

		// HeadSlicePtr
		{
			name: "HeadSlicePtr",
			data: struct {
				A *[]int `json:"a"`
			}{A: sliceptr([]int{-1})},
		},
		{
			name: "HeadSlicePtrOmitEmpty",
			data: struct {
				A *[]int `json:"a,omitempty"`
			}{A: sliceptr([]int{-1})},
		},
		{
			name: "HeadSlicePtrString",
			data: struct {
				A *[]int `json:"a,string"`
			}{A: sliceptr([]int{-1})},
		},

		// HeadSlicePtrNil
		{
			name: "HeadSlicePtrNil",
			data: struct {
				A *[]int `json:"a"`
			}{A: nil},
		},
		{
			name: "HeadSlicePtrNilOmitEmpty",
			data: struct {
				A *[]int `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name: "HeadSlicePtrNilString",
			data: struct {
				A *[]int `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadSliceZero
		{
			name: "PtrHeadSliceZero",
			data: &struct {
				A []int `json:"a"`
			}{},
		},
		{
			name: "PtrHeadSliceZeroOmitEmpty",
			data: &struct {
				A []int `json:"a,omitempty"`
			}{},
		},
		{
			name: "PtrHeadSliceZeroString",
			data: &struct {
				A []int `json:"a,string"`
			}{},
		},

		// PtrHeadSlice
		{
			name: "PtrHeadSlice",
			data: &struct {
				A []int `json:"a"`
			}{A: []int{-1}},
		},
		{
			name: "PtrHeadSliceOmitEmpty",
			data: &struct {
				A []int `json:"a,omitempty"`
			}{A: []int{-1}},
		},
		{
			name: "PtrHeadSliceString",
			data: &struct {
				A []int `json:"a,string"`
			}{A: []int{-1}},
		},

		// PtrHeadSlicePtr
		{
			name: "PtrHeadSlicePtr",
			data: &struct {
				A *[]int `json:"a"`
			}{A: sliceptr([]int{-1})},
		},
		{
			name: "PtrHeadSlicePtrOmitEmpty",
			data: &struct {
				A *[]int `json:"a,omitempty"`
			}{A: sliceptr([]int{-1})},
		},
		{
			name: "PtrHeadSlicePtrString",
			data: &struct {
				A *[]int `json:"a,string"`
			}{A: sliceptr([]int{-1})},
		},

		// PtrHeadSlicePtrNil
		{
			name: "PtrHeadSlicePtrNil",
			data: &struct {
				A *[]int `json:"a"`
			}{A: nil},
		},
		{
			name: "PtrHeadSlicePtrNilOmitEmpty",
			data: &struct {
				A *[]int `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name: "PtrHeadSlicePtrNilString",
			data: &struct {
				A *[]int `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadSliceNil
		{
			name: "PtrHeadSliceNil",
			data: (*struct {
				A *[]int `json:"a"`
			})(nil),
		},
		{
			name: "PtrHeadSliceNilOmitEmpty",
			data: (*struct {
				A *[]int `json:"a,omitempty"`
			})(nil),
		},
		{
			name: "PtrHeadSliceNilString",
			data: (*struct {
				A *[]int `json:"a,string"`
			})(nil),
		},

		// HeadSliceZeroMultiFields
		{
			name: "HeadSliceZeroMultiFields",
			data: struct {
				A []int `json:"a"`
				B []int `json:"b"`
				C []int `json:"c"`
			}{},
		},
		{
			name: "HeadSliceZeroMultiFieldsOmitEmpty",
			data: struct {
				A []int `json:"a,omitempty"`
				B []int `json:"b,omitempty"`
				C []int `json:"c,omitempty"`
			}{},
		},
		{
			name: "HeadSliceZeroMultiFields",
			data: struct {
				A []int `json:"a,string"`
				B []int `json:"b,string"`
				C []int `json:"c,string"`
			}{},
		},

		// HeadSliceMultiFields
		{
			name: "HeadSliceMultiFields",
			data: struct {
				A []int `json:"a"`
				B []int `json:"b"`
				C []int `json:"c"`
			}{A: []int{-1}, B: []int{-2}, C: []int{-3}},
		},
		{
			name: "HeadSliceMultiFieldsOmitEmpty",
			data: struct {
				A []int `json:"a,omitempty"`
				B []int `json:"b,omitempty"`
				C []int `json:"c,omitempty"`
			}{A: []int{-1}, B: []int{-2}, C: []int{-3}},
		},
		{
			name: "HeadSliceMultiFieldsString",
			data: struct {
				A []int `json:"a,string"`
				B []int `json:"b,string"`
				C []int `json:"c,string"`
			}{A: []int{-1}, B: []int{-2}, C: []int{-3}},
		},

		// HeadSlicePtrMultiFields
		{
			name: "HeadSlicePtrMultiFields",
			data: struct {
				A *[]int `json:"a"`
				B *[]int `json:"b"`
				C *[]int `json:"c"`
			}{A: sliceptr([]int{-1}), B: sliceptr([]int{-2}), C: sliceptr([]int{-3})},
		},
		{
			name: "HeadSlicePtrMultiFieldsOmitEmpty",
			data: struct {
				A *[]int `json:"a,omitempty"`
				B *[]int `json:"b,omitempty"`
				C *[]int `json:"c,omitempty"`
			}{A: sliceptr([]int{-1}), B: sliceptr([]int{-2}), C: sliceptr([]int{-3})},
		},
		{
			name: "HeadSlicePtrMultiFieldsString",
			data: struct {
				A *[]int `json:"a,string"`
				B *[]int `json:"b,string"`
				C *[]int `json:"c,string"`
			}{A: sliceptr([]int{-1}), B: sliceptr([]int{-2}), C: sliceptr([]int{-3})},
		},

		// HeadSlicePtrNilMultiFields
		{
			name: "HeadSlicePtrNilMultiFields",
			data: struct {
				A *[]int `json:"a"`
				B *[]int `json:"b"`
				C *[]int `json:"c"`
			}{A: nil, B: nil, C: nil},
		},
		{
			name: "HeadSlicePtrNilMultiFieldsOmitEmpty",
			data: struct {
				A *[]int `json:"a,omitempty"`
				B *[]int `json:"b,omitempty"`
				C *[]int `json:"c,omitempty"`
			}{A: nil, B: nil, C: nil},
		},
		{
			name: "HeadSlicePtrNilMultiFieldsString",
			data: struct {
				A *[]int `json:"a,string"`
				B *[]int `json:"b,string"`
				C *[]int `json:"c,string"`
			}{A: nil, B: nil, C: nil},
		},

		// PtrHeadSliceZeroMultiFields
		{
			name: "PtrHeadSliceZeroMultiFields",
			data: &struct {
				A []int `json:"a"`
				B []int `json:"b"`
			}{},
		},
		{
			name: "PtrHeadSliceZeroMultiFieldsOmitEmpty",
			data: &struct {
				A []int `json:"a,omitempty"`
				B []int `json:"b,omitempty"`
			}{},
		},
		{
			name: "PtrHeadSliceZeroMultiFieldsString",
			data: &struct {
				A []int `json:"a,string"`
				B []int `json:"b,string"`
			}{},
		},

		// PtrHeadSliceMultiFields
		{
			name: "PtrHeadSliceMultiFields",
			data: &struct {
				A []int `json:"a"`
				B []int `json:"b"`
			}{A: []int{-1}, B: nil},
		},
		{
			name: "PtrHeadSliceMultiFieldsOmitEmpty",
			data: &struct {
				A []int `json:"a,omitempty"`
				B []int `json:"b,omitempty"`
			}{A: []int{-1}, B: nil},
		},
		{
			name: "PtrHeadSliceMultiFieldsString",
			data: &struct {
				A []int `json:"a,string"`
				B []int `json:"b,string"`
			}{A: []int{-1}, B: nil},
		},

		// PtrHeadSlicePtrMultiFields
		{
			name: "PtrHeadSlicePtrMultiFields",
			data: &struct {
				A *[]int `json:"a"`
				B *[]int `json:"b"`
			}{A: sliceptr([]int{-1}), B: sliceptr([]int{-2})},
		},
		{
			name: "PtrHeadSlicePtrMultiFieldsOmitEmpty",
			data: &struct {
				A *[]int `json:"a,omitempty"`
				B *[]int `json:"b,omitempty"`
			}{A: sliceptr([]int{-1}), B: sliceptr([]int{-2})},
		},
		{
			name: "PtrHeadSlicePtrMultiFieldsString",
			data: &struct {
				A *[]int `json:"a,string"`
				B *[]int `json:"b,string"`
			}{A: sliceptr([]int{-1}), B: sliceptr([]int{-2})},
		},

		// PtrHeadSlicePtrNilMultiFields
		{
			name: "PtrHeadSlicePtrNilMultiFields",
			data: &struct {
				A *[]int `json:"a"`
				B *[]int `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadSlicePtrNilMultiFieldsOmitEmpty",
			data: &struct {
				A *[]int `json:"a,omitempty"`
				B *[]int `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadSlicePtrNilMultiFieldsString",
			data: &struct {
				A *[]int `json:"a,string"`
				B *[]int `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadSliceNilMultiFields
		{
			name: "PtrHeadSliceNilMultiFields",
			data: (*struct {
				A *[]int `json:"a"`
				B *[]int `json:"b"`
			})(nil),
		},
		{
			name: "PtrHeadSliceNilMultiFieldsOmitEmpty",
			data: (*struct {
				A *[]int `json:"a,omitempty"`
				B *[]int `json:"b,omitempty"`
			})(nil),
		},
		{
			name: "PtrHeadSliceNilMultiFieldsString",
			data: (*struct {
				A *[]int `json:"a,string"`
				B *[]int `json:"b,string"`
			})(nil),
		},

		// HeadSliceZeroNotRoot
		{
			name: "HeadSliceZeroNotRoot",
			data: struct {
				A struct {
					A []int `json:"a"`
				}
			}{},
		},
		{
			name: "HeadSliceZeroNotRootOmitEmpty",
			data: struct {
				A struct {
					A []int `json:"a,omitempty"`
				}
			}{},
		},
		{
			name: "HeadSliceZeroNotRootString",
			data: struct {
				A struct {
					A []int `json:"a,string"`
				}
			}{},
		},

		// HeadSliceNotRoot
		{
			name: "HeadSliceNotRoot",
			data: struct {
				A struct {
					A []int `json:"a"`
				}
			}{A: struct {
				A []int `json:"a"`
			}{A: true}},
		},
		{
			name: "HeadSliceNotRootOmitEmpty",
			data: struct {
				A struct {
					A []int `json:"a,omitempty"`
				}
			}{A: struct {
				A []int `json:"a,omitempty"`
			}{A: true}},
		},
		{
			name: "HeadSliceNotRootString",
			data: struct {
				A struct {
					A []int `json:"a,string"`
				}
			}{A: struct {
				A []int `json:"a,string"`
			}{A: true}},
		},

		// HeadSlicePtrNotRoot
		{
			name: "HeadSlicePtrNotRoot",
			data: struct {
				A struct {
					A *[]int `json:"a"`
				}
			}{A: struct {
				A *[]int `json:"a"`
			}{sliceptr(-1)}},
		},
		{
			name: "HeadSlicePtrNotRootOmitEmpty",
			data: struct {
				A struct {
					A *[]int `json:"a,omitempty"`
				}
			}{A: struct {
				A *[]int `json:"a,omitempty"`
			}{sliceptr(-1)}},
		},
		{
			name: "HeadSlicePtrNotRootString",
			data: struct {
				A struct {
					A *[]int `json:"a,string"`
				}
			}{A: struct {
				A *[]int `json:"a,string"`
			}{sliceptr(-1)}},
		},

		// HeadSlicePtrNilNotRoot
		{
			name: "HeadSlicePtrNilNotRoot",
			data: struct {
				A struct {
					A *[]int `json:"a"`
				}
			}{},
		},
		{
			name: "HeadSlicePtrNilNotRootOmitEmpty",
			data: struct {
				A struct {
					A *[]int `json:"a,omitempty"`
				}
			}{},
		},
		{
			name: "HeadSlicePtrNilNotRootString",
			data: struct {
				A struct {
					A *[]int `json:"a,string"`
				}
			}{},
		},

		// PtrHeadSliceZeroNotRoot
		{
			name: "PtrHeadBoolZeroNotRoot",
			data: struct {
				A *struct {
					A bool `json:"a"`
				}
			}{A: new(struct {
				A bool `json:"a"`
			})},
		},
		{
			name: "PtrHeadBoolZeroNotRootOmitEmpty",
			data: struct {
				A *struct {
					A bool `json:"a,omitempty"`
				}
			}{A: new(struct {
				A bool `json:"a,omitempty"`
			})},
		},
		{
			name: "PtrHeadBoolZeroNotRootString",
			data: struct {
				A *struct {
					A bool `json:"a,string"`
				}
			}{A: new(struct {
				A bool `json:"a,string"`
			})},
		},

		// PtrHeadBoolNotRoot
		{
			name: "PtrHeadBoolNotRoot",
			data: struct {
				A *struct {
					A bool `json:"a"`
				}
			}{A: &(struct {
				A bool `json:"a"`
			}{A: true})},
		},
		{
			name: "PtrHeadBoolNotRootOmitEmpty",
			data: struct {
				A *struct {
					A bool `json:"a,omitempty"`
				}
			}{A: &(struct {
				A bool `json:"a,omitempty"`
			}{A: true})},
		},
		{
			name: "PtrHeadBoolNotRootString",
			data: struct {
				A *struct {
					A bool `json:"a,string"`
				}
			}{A: &(struct {
				A bool `json:"a,string"`
			}{A: true})},
		},

		// PtrHeadBoolPtrNotRoot
		{
			name: "PtrHeadBoolPtrNotRoot",
			data: struct {
				A *struct {
					A *bool `json:"a"`
				}
			}{A: &(struct {
				A *bool `json:"a"`
			}{A: boolptr(true)})},
		},
		{
			name: "PtrHeadBoolPtrNotRootOmitEmpty",
			data: struct {
				A *struct {
					A *bool `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *bool `json:"a,omitempty"`
			}{A: boolptr(true)})},
		},
		{
			name: "PtrHeadBoolPtrNotRootString",
			data: struct {
				A *struct {
					A *bool `json:"a,string"`
				}
			}{A: &(struct {
				A *bool `json:"a,string"`
			}{A: boolptr(true)})},
		},

		// PtrHeadBoolPtrNilNotRoot
		{
			name: "PtrHeadBoolPtrNilNotRoot",
			data: struct {
				A *struct {
					A *bool `json:"a"`
				}
			}{A: &(struct {
				A *bool `json:"a"`
			}{A: nil})},
		},
		{
			name: "PtrHeadBoolPtrNilNotRootOmitEmpty",
			data: struct {
				A *struct {
					A *bool `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *bool `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name: "PtrHeadBoolPtrNilNotRootString",
			data: struct {
				A *struct {
					A *bool `json:"a,string"`
				}
			}{A: &(struct {
				A *bool `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadBoolNilNotRoot
		{
			name: "PtrHeadBoolNilNotRoot",
			data: struct {
				A *struct {
					A *bool `json:"a"`
				}
			}{A: nil},
		},
		{
			name: "PtrHeadBoolNilNotRootOmitEmpty",
			data: struct {
				A *struct {
					A *bool `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name: "PtrHeadBoolNilNotRootString",
			data: struct {
				A *struct {
					A *bool `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadBoolZeroMultiFieldsNotRoot
		{
			name: "HeadBoolZeroMultiFieldsNotRoot",
			data: struct {
				A struct {
					A bool `json:"a"`
				}
				B struct {
					B bool `json:"b"`
				}
			}{},
		},
		{
			name: "HeadBoolZeroMultiFieldsNotRootOmitEmpty",
			data: struct {
				A struct {
					A bool `json:"a,omitempty"`
				}
				B struct {
					B bool `json:"b,omitempty"`
				}
			}{},
		},
		{
			name: "HeadBoolZeroMultiFieldsNotRootString",
			data: struct {
				A struct {
					A bool `json:"a,string"`
				}
				B struct {
					B bool `json:"b,string"`
				}
			}{},
		},

		// HeadBoolMultiFieldsNotRoot
		{
			name: "HeadBoolMultiFieldsNotRoot",
			data: struct {
				A struct {
					A bool `json:"a"`
				}
				B struct {
					B bool `json:"b"`
				}
			}{A: struct {
				A bool `json:"a"`
			}{A: true}, B: struct {
				B bool `json:"b"`
			}{B: false}},
		},
		{
			name: "HeadBoolMultiFieldsNotRootOmitEmpty",
			data: struct {
				A struct {
					A bool `json:"a,omitempty"`
				}
				B struct {
					B bool `json:"b,omitempty"`
				}
			}{A: struct {
				A bool `json:"a,omitempty"`
			}{A: true}, B: struct {
				B bool `json:"b,omitempty"`
			}{B: false}},
		},
		{
			name: "HeadBoolMultiFieldsNotRootString",
			data: struct {
				A struct {
					A bool `json:"a,string"`
				}
				B struct {
					B bool `json:"b,string"`
				}
			}{A: struct {
				A bool `json:"a,string"`
			}{A: true}, B: struct {
				B bool `json:"b,string"`
			}{B: false}},
		},

		// HeadBoolPtrMultiFieldsNotRoot
		{
			name: "HeadBoolPtrMultiFieldsNotRoot",
			data: struct {
				A struct {
					A *bool `json:"a"`
				}
				B struct {
					B *bool `json:"b"`
				}
			}{A: struct {
				A *bool `json:"a"`
			}{A: boolptr(true)}, B: struct {
				B *bool `json:"b"`
			}{B: boolptr(false)}},
		},
		{
			name: "HeadBoolPtrMultiFieldsNotRootOmitEmpty",
			data: struct {
				A struct {
					A *bool `json:"a,omitempty"`
				}
				B struct {
					B *bool `json:"b,omitempty"`
				}
			}{A: struct {
				A *bool `json:"a,omitempty"`
			}{A: boolptr(true)}, B: struct {
				B *bool `json:"b,omitempty"`
			}{B: boolptr(false)}},
		},
		{
			name: "HeadBoolPtrMultiFieldsNotRootString",
			data: struct {
				A struct {
					A *bool `json:"a,string"`
				}
				B struct {
					B *bool `json:"b,string"`
				}
			}{A: struct {
				A *bool `json:"a,string"`
			}{A: boolptr(true)}, B: struct {
				B *bool `json:"b,string"`
			}{B: boolptr(false)}},
		},

		// HeadBoolPtrNilMultiFieldsNotRoot
		{
			name: "HeadBoolPtrNilMultiFieldsNotRoot",
			data: struct {
				A struct {
					A *bool `json:"a"`
				}
				B struct {
					B *bool `json:"b"`
				}
			}{A: struct {
				A *bool `json:"a"`
			}{A: nil}, B: struct {
				B *bool `json:"b"`
			}{B: nil}},
		},
		{
			name: "HeadBoolPtrNilMultiFieldsNotRootOmitEmpty",
			data: struct {
				A struct {
					A *bool `json:"a,omitempty"`
				}
				B struct {
					B *bool `json:"b,omitempty"`
				}
			}{A: struct {
				A *bool `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *bool `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name: "HeadBoolPtrNilMultiFieldsNotRootString",
			data: struct {
				A struct {
					A *bool `json:"a,string"`
				}
				B struct {
					B *bool `json:"b,string"`
				}
			}{A: struct {
				A *bool `json:"a,string"`
			}{A: nil}, B: struct {
				B *bool `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadBoolZeroMultiFieldsNotRoot
		{
			name: "PtrHeadBoolZeroMultiFieldsNotRoot",
			data: &struct {
				A struct {
					A bool `json:"a"`
				}
				B struct {
					B bool `json:"b"`
				}
			}{},
		},
		{
			name: "PtrHeadBoolZeroMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A struct {
					A bool `json:"a,omitempty"`
				}
				B struct {
					B bool `json:"b,omitempty"`
				}
			}{},
		},
		{
			name: "PtrHeadBoolZeroMultiFieldsNotRootString",
			data: &struct {
				A struct {
					A bool `json:"a,string"`
				}
				B struct {
					B bool `json:"b,string"`
				}
			}{},
		},

		// PtrHeadBoolMultiFieldsNotRoot
		{
			name: "PtrHeadBoolMultiFieldsNotRoot",
			data: &struct {
				A struct {
					A bool `json:"a"`
				}
				B struct {
					B bool `json:"b"`
				}
			}{A: struct {
				A bool `json:"a"`
			}{A: true}, B: struct {
				B bool `json:"b"`
			}{B: false}},
		},
		{
			name: "PtrHeadBoolMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A struct {
					A bool `json:"a,omitempty"`
				}
				B struct {
					B bool `json:"b,omitempty"`
				}
			}{A: struct {
				A bool `json:"a,omitempty"`
			}{A: true}, B: struct {
				B bool `json:"b,omitempty"`
			}{B: false}},
		},
		{
			name: "PtrHeadBoolMultiFieldsNotRootString",
			data: &struct {
				A struct {
					A bool `json:"a,string"`
				}
				B struct {
					B bool `json:"b,string"`
				}
			}{A: struct {
				A bool `json:"a,string"`
			}{A: true}, B: struct {
				B bool `json:"b,string"`
			}{B: false}},
		},

		// PtrHeadBoolPtrMultiFieldsNotRoot
		{
			name: "PtrHeadBoolPtrMultiFieldsNotRoot",
			data: &struct {
				A *struct {
					A *bool `json:"a"`
				}
				B *struct {
					B *bool `json:"b"`
				}
			}{A: &(struct {
				A *bool `json:"a"`
			}{A: boolptr(true)}), B: &(struct {
				B *bool `json:"b"`
			}{B: boolptr(false)})},
		},
		{
			name: "PtrHeadBoolPtrMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A *struct {
					A *bool `json:"a,omitempty"`
				}
				B *struct {
					B *bool `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *bool `json:"a,omitempty"`
			}{A: boolptr(true)}), B: &(struct {
				B *bool `json:"b,omitempty"`
			}{B: boolptr(false)})},
		},
		{
			name: "PtrHeadBoolPtrMultiFieldsNotRootString",
			data: &struct {
				A *struct {
					A *bool `json:"a,string"`
				}
				B *struct {
					B *bool `json:"b,string"`
				}
			}{A: &(struct {
				A *bool `json:"a,string"`
			}{A: boolptr(true)}), B: &(struct {
				B *bool `json:"b,string"`
			}{B: boolptr(false)})},
		},

		// PtrHeadBoolPtrNilMultiFieldsNotRoot
		{
			name: "PtrHeadBoolPtrNilMultiFieldsNotRoot",
			data: &struct {
				A *struct {
					A *bool `json:"a"`
				}
				B *struct {
					B *bool `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadBoolPtrNilMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A *struct {
					A *bool `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *bool `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadBoolPtrNilMultiFieldsNotRootString",
			data: &struct {
				A *struct {
					A *bool `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *bool `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadBoolNilMultiFieldsNotRoot
		{
			name: "PtrHeadBoolNilMultiFieldsNotRoot",
			data: (*struct {
				A *struct {
					A *bool `json:"a"`
				}
				B *struct {
					B *bool `json:"b"`
				}
			})(nil),
		},
		{
			name: "PtrHeadBoolNilMultiFieldsNotRootOmitEmpty",
			data: (*struct {
				A *struct {
					A *bool `json:"a,omitempty"`
				}
				B *struct {
					B *bool `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name: "PtrHeadBoolNilMultiFieldsNotRootString",
			data: (*struct {
				A *struct {
					A *bool `json:"a,string"`
				}
				B *struct {
					B *bool `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadBoolDoubleMultiFieldsNotRoot
		{
			name: "PtrHeadBoolDoubleMultiFieldsNotRoot",
			data: &struct {
				A *struct {
					A bool `json:"a"`
					B bool `json:"b"`
				}
				B *struct {
					A bool `json:"a"`
					B bool `json:"b"`
				}
			}{A: &(struct {
				A bool `json:"a"`
				B bool `json:"b"`
			}{A: true, B: false}), B: &(struct {
				A bool `json:"a"`
				B bool `json:"b"`
			}{A: true, B: false})},
		},
		{
			name: "PtrHeadBoolDoubleMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A *struct {
					A bool `json:"a,omitempty"`
					B bool `json:"b,omitempty"`
				}
				B *struct {
					A bool `json:"a,omitempty"`
					B bool `json:"b,omitempty"`
				}
			}{A: &(struct {
				A bool `json:"a,omitempty"`
				B bool `json:"b,omitempty"`
			}{A: true, B: false}), B: &(struct {
				A bool `json:"a,omitempty"`
				B bool `json:"b,omitempty"`
			}{A: true, B: false})},
		},
		{
			name: "PtrHeadBoolDoubleMultiFieldsNotRootString",
			data: &struct {
				A *struct {
					A bool `json:"a,string"`
					B bool `json:"b,string"`
				}
				B *struct {
					A bool `json:"a,string"`
					B bool `json:"b,string"`
				}
			}{A: &(struct {
				A bool `json:"a,string"`
				B bool `json:"b,string"`
			}{A: true, B: false}), B: &(struct {
				A bool `json:"a,string"`
				B bool `json:"b,string"`
			}{A: true, B: false})},
		},

		// PtrHeadBoolNilDoubleMultiFieldsNotRoot
		{
			name: "PtrHeadBoolNilDoubleMultiFieldsNotRoot",
			data: &struct {
				A *struct {
					A bool `json:"a"`
					B bool `json:"b"`
				}
				B *struct {
					A bool `json:"a"`
					B bool `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadBoolNilDoubleMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A *struct {
					A bool `json:"a,omitempty"`
					B bool `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A bool `json:"a,omitempty"`
					B bool `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadBoolNilDoubleMultiFieldsNotRootString",
			data: &struct {
				A *struct {
					A bool `json:"a,string"`
					B bool `json:"b,string"`
				}
				B *struct {
					A bool `json:"a,string"`
					B bool `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadBoolNilDoubleMultiFieldsNotRoot
		{
			name: "PtrHeadBoolNilDoubleMultiFieldsNotRoot",
			data: (*struct {
				A *struct {
					A bool `json:"a"`
					B bool `json:"b"`
				}
				B *struct {
					A bool `json:"a"`
					B bool `json:"b"`
				}
			})(nil),
		},
		{
			name: "PtrHeadBoolNilDoubleMultiFieldsNotRootOmitEmpty",
			data: (*struct {
				A *struct {
					A bool `json:"a,omitempty"`
					B bool `json:"b,omitempty"`
				}
				B *struct {
					A bool `json:"a,omitempty"`
					B bool `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name: "PtrHeadBoolNilDoubleMultiFieldsNotRootString",
			data: (*struct {
				A *struct {
					A bool `json:"a,string"`
					B bool `json:"b,string"`
				}
				B *struct {
					A bool `json:"a,string"`
					B bool `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadBoolPtrDoubleMultiFieldsNotRoot
		{
			name: "PtrHeadBoolPtrDoubleMultiFieldsNotRoot",
			data: &struct {
				A *struct {
					A *bool `json:"a"`
					B *bool `json:"b"`
				}
				B *struct {
					A *bool `json:"a"`
					B *bool `json:"b"`
				}
			}{A: &(struct {
				A *bool `json:"a"`
				B *bool `json:"b"`
			}{A: boolptr(true), B: boolptr(false)}), B: &(struct {
				A *bool `json:"a"`
				B *bool `json:"b"`
			}{A: boolptr(true), B: boolptr(false)})},
		},
		{
			name: "PtrHeadBoolPtrDoubleMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A *struct {
					A *bool `json:"a,omitempty"`
					B *bool `json:"b,omitempty"`
				}
				B *struct {
					A *bool `json:"a,omitempty"`
					B *bool `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *bool `json:"a,omitempty"`
				B *bool `json:"b,omitempty"`
			}{A: boolptr(true), B: boolptr(false)}), B: &(struct {
				A *bool `json:"a,omitempty"`
				B *bool `json:"b,omitempty"`
			}{A: boolptr(true), B: boolptr(false)})},
		},
		{
			name: "PtrHeadBoolPtrDoubleMultiFieldsNotRootString",
			data: &struct {
				A *struct {
					A *bool `json:"a,string"`
					B *bool `json:"b,string"`
				}
				B *struct {
					A *bool `json:"a,string"`
					B *bool `json:"b,string"`
				}
			}{A: &(struct {
				A *bool `json:"a,string"`
				B *bool `json:"b,string"`
			}{A: boolptr(true), B: boolptr(false)}), B: &(struct {
				A *bool `json:"a,string"`
				B *bool `json:"b,string"`
			}{A: boolptr(true), B: boolptr(false)})},
		},

		// PtrHeadBoolPtrNilDoubleMultiFieldsNotRoot
		{
			name: "PtrHeadBoolPtrNilDoubleMultiFieldsNotRoot",
			data: &struct {
				A *struct {
					A *bool `json:"a"`
					B *bool `json:"b"`
				}
				B *struct {
					A *bool `json:"a"`
					B *bool `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadBoolPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			data: &struct {
				A *struct {
					A *bool `json:"a,omitempty"`
					B *bool `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *bool `json:"a,omitempty"`
					B *bool `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name: "PtrHeadBoolPtrNilDoubleMultiFieldsNotRootString",
			data: &struct {
				A *struct {
					A *bool `json:"a,string"`
					B *bool `json:"b,string"`
				}
				B *struct {
					A *bool `json:"a,string"`
					B *bool `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadBoolPtrNilDoubleMultiFieldsNotRoot
		{
			name: "PtrHeadBoolPtrNilDoubleMultiFieldsNotRoot",
			data: (*struct {
				A *struct {
					A *bool `json:"a"`
					B *bool `json:"b"`
				}
				B *struct {
					A *bool `json:"a"`
					B *bool `json:"b"`
				}
			})(nil),
		},
		{
			name: "PtrHeadBoolPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			data: (*struct {
				A *struct {
					A *bool `json:"a,omitempty"`
					B *bool `json:"b,omitempty"`
				}
				B *struct {
					A *bool `json:"a,omitempty"`
					B *bool `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name: "PtrHeadBoolPtrNilDoubleMultiFieldsNotRootString",
			data: (*struct {
				A *struct {
					A *bool `json:"a,string"`
					B *bool `json:"b,string"`
				}
				B *struct {
					A *bool `json:"a,string"`
					B *bool `json:"b,string"`
				}
			})(nil),
		},
		/*
			// AnonymousHeadBool
			{
				name: "AnonymousHeadBool",
				data: struct {
					structBool
					B bool `json:"b"`
				}{
					structBool: structBool{A: true},
					B:          false,
				},
			},
			{
				name: "AnonymousHeadBoolOmitEmpty",
				data: struct {
					structBoolOmitEmpty
					B bool `json:"b,omitempty"`
				}{
					structBoolOmitEmpty: structBoolOmitEmpty{A: true},
					B:                   false,
				},
			},
			{
				name: "AnonymousHeadBoolString",
				data: struct {
					structBoolString
					B bool `json:"b,string"`
				}{
					structBoolString: structBoolString{A: true},
					B:                false,
				},
			},

			// PtrAnonymousHeadBool
			{
				name: "PtrAnonymousHeadBool",
				data: struct {
					*structBool
					B bool `json:"b"`
				}{
					structBool: &structBool{A: true},
					B:          false,
				},
			},
			{
				name: "PtrAnonymousHeadBoolOmitEmpty",
				data: struct {
					*structBoolOmitEmpty
					B bool `json:"b,omitempty"`
				}{
					structBoolOmitEmpty: &structBoolOmitEmpty{A: true},
					B:                   false,
				},
			},
			{
				name: "PtrAnonymousHeadBoolString",
				data: struct {
					*structBoolString
					B bool `json:"b,string"`
				}{
					structBoolString: &structBoolString{A: true},
					B:                false,
				},
			},

			// NilPtrAnonymousHeadBool
			{
				name: "NilPtrAnonymousHeadBool",
				data: struct {
					*structBool
					B bool `json:"b"`
				}{
					structBool: nil,
					B:          true,
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolOmitEmpty",
				data: struct {
					*structBoolOmitEmpty
					B bool `json:"b,omitempty"`
				}{
					structBoolOmitEmpty: nil,
					B:                   true,
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolString",
				data: struct {
					*structBoolString
					B bool `json:"b,string"`
				}{
					structBoolString: nil,
					B:                true,
				},
			},

			// AnonymousHeadBoolPtr
			{
				name: "AnonymousHeadBoolPtr",
				data: struct {
					structBoolPtr
					B *bool `json:"b"`
				}{
					structBoolPtr: structBoolPtr{A: boolptr(true)},
					B:             boolptr(false),
				},
			},
			{
				name: "AnonymousHeadBoolPtrOmitEmpty",
				data: struct {
					structBoolPtrOmitEmpty
					B *bool `json:"b,omitempty"`
				}{
					structBoolPtrOmitEmpty: structBoolPtrOmitEmpty{A: boolptr(true)},
					B:                      boolptr(false),
				},
			},
			{
				name: "AnonymousHeadBoolPtrString",
				data: struct {
					structBoolPtrString
					B *bool `json:"b,string"`
				}{
					structBoolPtrString: structBoolPtrString{A: boolptr(true)},
					B:                   boolptr(false),
				},
			},

			// AnonymousHeadBoolPtrNil
			{
				name: "AnonymousHeadBoolPtrNil",
				data: struct {
					structBoolPtr
					B *bool `json:"b"`
				}{
					structBoolPtr: structBoolPtr{A: nil},
					B:             boolptr(true),
				},
			},
			{
				name: "AnonymousHeadBoolPtrNilOmitEmpty",
				data: struct {
					structBoolPtrOmitEmpty
					B *bool `json:"b,omitempty"`
				}{
					structBoolPtrOmitEmpty: structBoolPtrOmitEmpty{A: nil},
					B:                      boolptr(true),
				},
			},
			{
				name: "AnonymousHeadBoolPtrNilString",
				data: struct {
					structBoolPtrString
					B *bool `json:"b,string"`
				}{
					structBoolPtrString: structBoolPtrString{A: nil},
					B:                   boolptr(true),
				},
			},

			// PtrAnonymousHeadBoolPtr
			{
				name: "PtrAnonymousHeadBoolPtr",
				data: struct {
					*structBoolPtr
					B *bool `json:"b"`
				}{
					structBoolPtr: &structBoolPtr{A: boolptr(true)},
					B:             boolptr(false),
				},
			},
			{
				name: "PtrAnonymousHeadBoolPtrOmitEmpty",
				data: struct {
					*structBoolPtrOmitEmpty
					B *bool `json:"b,omitempty"`
				}{
					structBoolPtrOmitEmpty: &structBoolPtrOmitEmpty{A: boolptr(true)},
					B:                      boolptr(false),
				},
			},
			{
				name: "PtrAnonymousHeadBoolPtrString",
				data: struct {
					*structBoolPtrString
					B *bool `json:"b,string"`
				}{
					structBoolPtrString: &structBoolPtrString{A: boolptr(true)},
					B:                   boolptr(false),
				},
			},

			// NilPtrAnonymousHeadBoolPtr
			{
				name: "NilPtrAnonymousHeadBoolPtr",
				data: struct {
					*structBoolPtr
					B *bool `json:"b"`
				}{
					structBoolPtr: nil,
					B:             boolptr(true),
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolPtrOmitEmpty",
				data: struct {
					*structBoolPtrOmitEmpty
					B *bool `json:"b,omitempty"`
				}{
					structBoolPtrOmitEmpty: nil,
					B:                      boolptr(true),
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolPtrString",
				data: struct {
					*structBoolPtrString
					B *bool `json:"b,string"`
				}{
					structBoolPtrString: nil,
					B:                   boolptr(true),
				},
			},

			// AnonymousHeadBoolOnly
			{
				name: "AnonymousHeadBoolOnly",
				data: struct {
					structBool
				}{
					structBool: structBool{A: true},
				},
			},
			{
				name: "AnonymousHeadBoolOnlyOmitEmpty",
				data: struct {
					structBoolOmitEmpty
				}{
					structBoolOmitEmpty: structBoolOmitEmpty{A: true},
				},
			},
			{
				name: "AnonymousHeadBoolOnlyString",
				data: struct {
					structBoolString
				}{
					structBoolString: structBoolString{A: true},
				},
			},

			// PtrAnonymousHeadBoolOnly
			{
				name: "PtrAnonymousHeadBoolOnly",
				data: struct {
					*structBool
				}{
					structBool: &structBool{A: true},
				},
			},
			{
				name: "PtrAnonymousHeadBoolOnlyOmitEmpty",
				data: struct {
					*structBoolOmitEmpty
				}{
					structBoolOmitEmpty: &structBoolOmitEmpty{A: true},
				},
			},
			{
				name: "PtrAnonymousHeadBoolOnlyString",
				data: struct {
					*structBoolString
				}{
					structBoolString: &structBoolString{A: true},
				},
			},

			// NilPtrAnonymousHeadBoolOnly
			{
				name: "NilPtrAnonymousHeadBoolOnly",
				data: struct {
					*structBool
				}{
					structBool: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolOnlyOmitEmpty",
				data: struct {
					*structBoolOmitEmpty
				}{
					structBoolOmitEmpty: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolOnlyString",
				data: struct {
					*structBoolString
				}{
					structBoolString: nil,
				},
			},

			// AnonymousHeadBoolPtrOnly
			{
				name: "AnonymousHeadBoolPtrOnly",
				data: struct {
					structBoolPtr
				}{
					structBoolPtr: structBoolPtr{A: boolptr(true)},
				},
			},
			{
				name: "AnonymousHeadBoolPtrOnlyOmitEmpty",
				data: struct {
					structBoolPtrOmitEmpty
				}{
					structBoolPtrOmitEmpty: structBoolPtrOmitEmpty{A: boolptr(true)},
				},
			},
			{
				name: "AnonymousHeadBoolPtrOnlyString",
				data: struct {
					structBoolPtrString
				}{
					structBoolPtrString: structBoolPtrString{A: boolptr(true)},
				},
			},

			// AnonymousHeadBoolPtrNilOnly
			{
				name: "AnonymousHeadBoolPtrNilOnly",
				data: struct {
					structBoolPtr
				}{
					structBoolPtr: structBoolPtr{A: nil},
				},
			},
			{
				name: "AnonymousHeadBoolPtrNilOnlyOmitEmpty",
				data: struct {
					structBoolPtrOmitEmpty
				}{
					structBoolPtrOmitEmpty: structBoolPtrOmitEmpty{A: nil},
				},
			},
			{
				name: "AnonymousHeadBoolPtrNilOnlyString",
				data: struct {
					structBoolPtrString
				}{
					structBoolPtrString: structBoolPtrString{A: nil},
				},
			},

			// PtrAnonymousHeadBoolPtrOnly
			{
				name: "PtrAnonymousHeadBoolPtrOnly",
				data: struct {
					*structBoolPtr
				}{
					structBoolPtr: &structBoolPtr{A: boolptr(true)},
				},
			},
			{
				name: "PtrAnonymousHeadBoolPtrOnlyOmitEmpty",
				data: struct {
					*structBoolPtrOmitEmpty
				}{
					structBoolPtrOmitEmpty: &structBoolPtrOmitEmpty{A: boolptr(true)},
				},
			},
			{
				name: "PtrAnonymousHeadBoolPtrOnlyString",
				data: struct {
					*structBoolPtrString
				}{
					structBoolPtrString: &structBoolPtrString{A: boolptr(true)},
				},
			},

			// NilPtrAnonymousHeadBoolPtrOnly
			{
				name: "NilPtrAnonymousHeadBoolPtrOnly",
				data: struct {
					*structBoolPtr
				}{
					structBoolPtr: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolPtrOnlyOmitEmpty",
				data: struct {
					*structBoolPtrOmitEmpty
				}{
					structBoolPtrOmitEmpty: nil,
				},
			},
			{
				name: "NilPtrAnonymousHeadBoolPtrOnlyString",
				data: struct {
					*structBoolPtrString
				}{
					structBoolPtrString: nil,
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
