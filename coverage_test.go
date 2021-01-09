package json

import (
	"bytes"
	"strings"
	"testing"
)

func intptr(v int) *int          { return &v }
func int8ptr(v int8) *int8       { return &v }
func int16ptr(v int16) *int16    { return &v }
func int32ptr(v int32) *int32    { return &v }
func int64ptr(v int64) *int64    { return &v }
func uptr(v uint) *uint          { return &v }
func uint8ptr(v uint8) *uint8    { return &v }
func uint16ptr(v uint16) *uint16 { return &v }
func uint32ptr(v uint32) *uint32 { return &v }
func uint64ptr(v uint64) *uint64 { return &v }

func TestCoverStructHeadInt(t *testing.T) {
	type structInt struct {
		A int `json:"a"`
	}
	type structIntPtr struct {
		A *int `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadIntZero",
			expected: `{"a":0}`,
			data: struct {
				A int `json:"a"`
			}{},
		},
		{
			name:     "HeadInt",
			expected: `{"a":1}`,
			data: struct {
				A int `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadIntPtr",
			expected: `{"a":1}`,
			data: struct {
				A *int `json:"a"`
			}{A: intptr(1)},
		},
		{
			name:     "HeadIntPtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *int `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadIntZero",
			expected: `{"a":0}`,
			data: &struct {
				A int `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt",
			expected: `{"a":1}`,
			data: &struct {
				A int `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadIntPtr",
			expected: `{"a":1}`,
			data: &struct {
				A *int `json:"a"`
			}{A: intptr(1)},
		},
		{
			name:     "PtrHeadIntPtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *int `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadIntNil",
			expected: `null`,
			data: (*struct {
				A *int `json:"a"`
			})(nil),
		},
		{
			name:     "HeadIntZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A int `json:"a"`
				B int `json:"b"`
			}{},
		},
		{
			name:     "HeadIntMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadIntPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(1), B: intptr(2)},
		},
		{
			name:     "HeadIntPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A int `json:"a"`
				B int `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadIntMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadIntPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(1), B: intptr(2)},
		},
		{
			name:     "PtrHeadIntPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntNilMultiFields",
			expected: `null`,
			data: (*struct {
				A *int `json:"a"`
				B *int `json:"b"`
			})(nil),
		},

		{
			name:     "HeadIntZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A int `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadIntNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A int `json:"a"`
				}
			}{A: struct {
				A int `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadIntPtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *int `json:"a"`
				}
			}{A: struct {
				A *int `json:"a"`
			}{intptr(1)}},
		},
		{
			name:     "HeadIntPtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *int `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadIntZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A int `json:"a"`
				}
			}{A: new(struct {
				A int `json:"a"`
			})},
		},
		{
			name:     "PtrHeadIntNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A int `json:"a"`
				}
			}{A: &(struct {
				A int `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadIntPtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *int `json:"a"`
				}
			}{A: &(struct {
				A *int `json:"a"`
			}{A: intptr(1)})},
		},
		{
			name:     "PtrHeadIntPtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *int `json:"a"`
				}
			}{A: &(struct {
				A *int `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadIntNilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *int `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadIntZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadIntMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{A: struct {
				A int `json:"a"`
			}{A: 1}, B: struct {
				B int `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadIntPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *int `json:"a"`
				}
				B struct {
					B *int `json:"b"`
				}
			}{A: struct {
				A *int `json:"a"`
			}{A: intptr(1)}, B: struct {
				B *int `json:"b"`
			}{B: intptr(2)}},
		},
		{
			name:     "HeadIntPtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *int `json:"a"`
				}
				B struct {
					B *int `json:"b"`
				}
			}{A: struct {
				A *int `json:"a"`
			}{A: nil}, B: struct {
				B *int `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadIntZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadIntMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{A: struct {
				A int `json:"a"`
			}{A: 1}, B: struct {
				B int `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadIntPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *int `json:"a"`
				}
				B *struct {
					B *int `json:"b"`
				}
			}{A: &(struct {
				A *int `json:"a"`
			}{A: intptr(1)}), B: &(struct {
				B *int `json:"b"`
			}{B: intptr(2)})},
		},
		{
			name:     "PtrHeadIntPtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int `json:"a"`
				}
				B *struct {
					B *int `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntNilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int `json:"a"`
				}
				B *struct {
					B *int `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
				B *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
			}{A: &(struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
				B *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
				B *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntPtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
				B *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
			}{A: &(struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(1), B: intptr(2)}), B: &(struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(3), B: intptr(4)})},
		},
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
				B *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
				B *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadInt",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt
				B int `json:"b"`
			}{
				structInt: structInt{A: 1},
				B:         2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt
				B int `json:"b"`
			}{
				structInt: &structInt{A: 1},
				B:         2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt",
			expected: `{"b":2}`,
			data: struct {
				*structInt
				B int `json:"b"`
			}{
				structInt: nil,
				B:         2,
			},
		},
		{
			name:     "AnonymousHeadIntPtr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: structIntPtr{A: intptr(1)},
				B:            intptr(2),
			},
		},
		{
			name:     "AnonymousHeadIntPtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: structIntPtr{A: nil},
				B:            intptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadIntPtr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: &structIntPtr{A: intptr(1)},
				B:            intptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntPtr",
			expected: `{"b":2}`,
			data: struct {
				*structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: nil,
				B:            intptr(2),
			},
		},
		{
			name:     "AnonymousHeadIntOnly",
			expected: `{"a":1}`,
			data: struct {
				structInt
			}{
				structInt: structInt{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadIntOnly",
			expected: `{"a":1}`,
			data: struct {
				*structInt
			}{
				structInt: &structInt{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntOnly",
			expected: `{}`,
			data: struct {
				*structInt
			}{
				structInt: nil,
			},
		},
		{
			name:     "AnonymousHeadIntPtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structIntPtr
			}{
				structIntPtr: structIntPtr{A: intptr(1)},
			},
		},
		{
			name:     "AnonymousHeadIntPtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structIntPtr
			}{
				structIntPtr: structIntPtr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadIntPtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structIntPtr
			}{
				structIntPtr: &structIntPtr{A: intptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntPtrOnly",
			expected: `{}`,
			data: struct {
				*structIntPtr
			}{
				structIntPtr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadInt8(t *testing.T) {
	type structInt8 struct {
		A int8 `json:"a"`
	}
	type structInt8Ptr struct {
		A *int8 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadInt8Zero",
			expected: `{"a":0}`,
			data: struct {
				A int8 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt8",
			expected: `{"a":1}`,
			data: struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt8Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		{
			name:     "HeadInt8PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *int8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt8Zero",
			expected: `{"a":0}`,
			data: &struct {
				A int8 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt8",
			expected: `{"a":1}`,
			data: &struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt8Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		{
			name:     "PtrHeadInt8PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *int8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt8Nil",
			expected: `null`,
			data: (*struct {
				A *int8 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadInt8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt8MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
		{
			name:     "HeadInt8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt8MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
		{
			name:     "PtrHeadInt8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadInt8ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A int8 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt8NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A int8 `json:"a"`
				}
			}{A: struct {
				A int8 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt8PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *int8 `json:"a"`
				}
			}{A: struct {
				A *int8 `json:"a"`
			}{int8ptr(1)}},
		},
		{
			name:     "HeadInt8PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *int8 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt8ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A int8 `json:"a"`
				}
			}{A: new(struct {
				A int8 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt8NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A int8 `json:"a"`
				}
			}{A: &(struct {
				A int8 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt8PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *int8 `json:"a"`
				}
			}{A: &(struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)})},
		},
		{
			name:     "PtrHeadInt8PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *int8 `json:"a"`
				}
			}{A: &(struct {
				A *int8 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt8NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *int8 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadInt8ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A int8 `json:"a"`
				}
				B struct {
					B int8 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadInt8MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A int8 `json:"a"`
				}
				B struct {
					B int8 `json:"b"`
				}
			}{A: struct {
				A int8 `json:"a"`
			}{A: 1}, B: struct {
				B int8 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadInt8PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *int8 `json:"a"`
				}
				B struct {
					B *int8 `json:"b"`
				}
			}{A: struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)}, B: struct {
				B *int8 `json:"b"`
			}{B: int8ptr(2)}},
		},
		{
			name:     "HeadInt8PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *int8 `json:"a"`
				}
				B struct {
					B *int8 `json:"b"`
				}
			}{A: struct {
				A *int8 `json:"a"`
			}{A: nil}, B: struct {
				B *int8 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadInt8ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A int8 `json:"a"`
				}
				B struct {
					B int8 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt8MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A int8 `json:"a"`
				}
				B struct {
					B int8 `json:"b"`
				}
			}{A: struct {
				A int8 `json:"a"`
			}{A: 1}, B: struct {
				B int8 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt8PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *int8 `json:"a"`
				}
				B *struct {
					B *int8 `json:"b"`
				}
			}{A: &(struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)}), B: &(struct {
				B *int8 `json:"b"`
			}{B: int8ptr(2)})},
		},
		{
			name:     "PtrHeadInt8PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int8 `json:"a"`
				}
				B *struct {
					B *int8 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int8 `json:"a"`
				}
				B *struct {
					B *int8 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt8DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A int8 `json:"a"`
					B int8 `json:"b"`
				}
				B *struct {
					A int8 `json:"a"`
					B int8 `json:"b"`
				}
			}{A: &(struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A int8 `json:"a"`
					B int8 `json:"b"`
				}
				B *struct {
					A int8 `json:"a"`
					B int8 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A int8 `json:"a"`
					B int8 `json:"b"`
				}
				B *struct {
					A int8 `json:"a"`
					B int8 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt8PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *int8 `json:"a"`
					B *int8 `json:"b"`
				}
				B *struct {
					A *int8 `json:"a"`
					B *int8 `json:"b"`
				}
			}{A: &(struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(1), B: int8ptr(2)}), B: &(struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(3), B: int8ptr(4)})},
		},
		{
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int8 `json:"a"`
					B *int8 `json:"b"`
				}
				B *struct {
					A *int8 `json:"a"`
					B *int8 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int8 `json:"a"`
					B *int8 `json:"b"`
				}
				B *struct {
					A *int8 `json:"a"`
					B *int8 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadInt8",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt8
				B int8 `json:"b"`
			}{
				structInt8: structInt8{A: 1},
				B:          2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt8",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt8
				B int8 `json:"b"`
			}{
				structInt8: &structInt8{A: 1},
				B:          2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8",
			expected: `{"b":2}`,
			data: struct {
				*structInt8
				B int8 `json:"b"`
			}{
				structInt8: nil,
				B:          2,
			},
		},
		{
			name:     "AnonymousHeadInt8Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: structInt8Ptr{A: int8ptr(1)},
				B:             int8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt8PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: structInt8Ptr{A: nil},
				B:             int8ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt8Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: &structInt8Ptr{A: int8ptr(1)},
				B:             int8ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: nil,
				B:             int8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt8Only",
			expected: `{"a":1}`,
			data: struct {
				structInt8
			}{
				structInt8: structInt8{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt8Only",
			expected: `{"a":1}`,
			data: struct {
				*structInt8
			}{
				structInt8: &structInt8{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8Only",
			expected: `{}`,
			data: struct {
				*structInt8
			}{
				structInt8: nil,
			},
		},
		{
			name:     "AnonymousHeadInt8PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structInt8Ptr
			}{
				structInt8Ptr: structInt8Ptr{A: int8ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt8PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structInt8Ptr
			}{
				structInt8Ptr: structInt8Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadInt8PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structInt8Ptr
			}{
				structInt8Ptr: &structInt8Ptr{A: int8ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8PtrOnly",
			expected: `{}`,
			data: struct {
				*structInt8Ptr
			}{
				structInt8Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadInt16(t *testing.T) {
	type structInt16 struct {
		A int16 `json:"a"`
	}
	type structInt16Ptr struct {
		A *int16 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadInt16Zero",
			expected: `{"a":0}`,
			data: struct {
				A int16 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt16",
			expected: `{"a":1}`,
			data: struct {
				A int16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt16Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)},
		},
		{
			name:     "HeadInt16PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *int16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt16Zero",
			expected: `{"a":0}`,
			data: &struct {
				A int16 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt16",
			expected: `{"a":1}`,
			data: &struct {
				A int16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt16Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)},
		},
		{
			name:     "PtrHeadInt16PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *int16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt16Nil",
			expected: `null`,
			data: (*struct {
				A *int16 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadInt16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt16MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},
		{
			name:     "HeadInt16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt16MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},
		{
			name:     "PtrHeadInt16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadInt16ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A int16 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt16NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A int16 `json:"a"`
				}
			}{A: struct {
				A int16 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt16PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *int16 `json:"a"`
				}
			}{A: struct {
				A *int16 `json:"a"`
			}{int16ptr(1)}},
		},
		{
			name:     "HeadInt16PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *int16 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt16ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A int16 `json:"a"`
				}
			}{A: new(struct {
				A int16 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt16NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A int16 `json:"a"`
				}
			}{A: &(struct {
				A int16 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt16PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *int16 `json:"a"`
				}
			}{A: &(struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)})},
		},
		{
			name:     "PtrHeadInt16PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *int16 `json:"a"`
				}
			}{A: &(struct {
				A *int16 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt16NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *int16 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadInt16ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A int16 `json:"a"`
				}
				B struct {
					B int16 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadInt16MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A int16 `json:"a"`
				}
				B struct {
					B int16 `json:"b"`
				}
			}{A: struct {
				A int16 `json:"a"`
			}{A: 1}, B: struct {
				B int16 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadInt16PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *int16 `json:"a"`
				}
				B struct {
					B *int16 `json:"b"`
				}
			}{A: struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)}, B: struct {
				B *int16 `json:"b"`
			}{B: int16ptr(2)}},
		},
		{
			name:     "HeadInt16PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *int16 `json:"a"`
				}
				B struct {
					B *int16 `json:"b"`
				}
			}{A: struct {
				A *int16 `json:"a"`
			}{A: nil}, B: struct {
				B *int16 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadInt16ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A int16 `json:"a"`
				}
				B struct {
					B int16 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt16MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A int16 `json:"a"`
				}
				B struct {
					B int16 `json:"b"`
				}
			}{A: struct {
				A int16 `json:"a"`
			}{A: 1}, B: struct {
				B int16 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt16PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *int16 `json:"a"`
				}
				B *struct {
					B *int16 `json:"b"`
				}
			}{A: &(struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)}), B: &(struct {
				B *int16 `json:"b"`
			}{B: int16ptr(2)})},
		},
		{
			name:     "PtrHeadInt16PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int16 `json:"a"`
				}
				B *struct {
					B *int16 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int16 `json:"a"`
				}
				B *struct {
					B *int16 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt16DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A int16 `json:"a"`
					B int16 `json:"b"`
				}
				B *struct {
					A int16 `json:"a"`
					B int16 `json:"b"`
				}
			}{A: &(struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt16NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A int16 `json:"a"`
					B int16 `json:"b"`
				}
				B *struct {
					A int16 `json:"a"`
					B int16 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A int16 `json:"a"`
					B int16 `json:"b"`
				}
				B *struct {
					A int16 `json:"a"`
					B int16 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt16PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *int16 `json:"a"`
					B *int16 `json:"b"`
				}
				B *struct {
					A *int16 `json:"a"`
					B *int16 `json:"b"`
				}
			}{A: &(struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: int16ptr(1), B: int16ptr(2)}), B: &(struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: int16ptr(3), B: int16ptr(4)})},
		},
		{
			name:     "PtrHeadInt16PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int16 `json:"a"`
					B *int16 `json:"b"`
				}
				B *struct {
					A *int16 `json:"a"`
					B *int16 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int16 `json:"a"`
					B *int16 `json:"b"`
				}
				B *struct {
					A *int16 `json:"a"`
					B *int16 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadInt16",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt16
				B int16 `json:"b"`
			}{
				structInt16: structInt16{A: 1},
				B:           2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt16",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt16
				B int16 `json:"b"`
			}{
				structInt16: &structInt16{A: 1},
				B:           2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16",
			expected: `{"b":2}`,
			data: struct {
				*structInt16
				B int16 `json:"b"`
			}{
				structInt16: nil,
				B:           2,
			},
		},
		{
			name:     "AnonymousHeadInt16Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt16Ptr
				B *int16 `json:"b"`
			}{
				structInt16Ptr: structInt16Ptr{A: int16ptr(1)},
				B:              int16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt16PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structInt16Ptr
				B *int16 `json:"b"`
			}{
				structInt16Ptr: structInt16Ptr{A: nil},
				B:              int16ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt16Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt16Ptr
				B *int16 `json:"b"`
			}{
				structInt16Ptr: &structInt16Ptr{A: int16ptr(1)},
				B:              int16ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structInt16Ptr
				B *int16 `json:"b"`
			}{
				structInt16Ptr: nil,
				B:              int16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt16Only",
			expected: `{"a":1}`,
			data: struct {
				structInt16
			}{
				structInt16: structInt16{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt16Only",
			expected: `{"a":1}`,
			data: struct {
				*structInt16
			}{
				structInt16: &structInt16{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16Only",
			expected: `{}`,
			data: struct {
				*structInt16
			}{
				structInt16: nil,
			},
		},
		{
			name:     "AnonymousHeadInt16PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structInt16Ptr
			}{
				structInt16Ptr: structInt16Ptr{A: int16ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt16PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structInt16Ptr
			}{
				structInt16Ptr: structInt16Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadInt16PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structInt16Ptr
			}{
				structInt16Ptr: &structInt16Ptr{A: int16ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16PtrOnly",
			expected: `{}`,
			data: struct {
				*structInt16Ptr
			}{
				structInt16Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadInt32(t *testing.T) {
	type structInt32 struct {
		A int32 `json:"a"`
	}
	type structInt32Ptr struct {
		A *int32 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadInt32Zero",
			expected: `{"a":0}`,
			data: struct {
				A int32 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt32",
			expected: `{"a":1}`,
			data: struct {
				A int32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt32Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)},
		},
		{
			name:     "HeadInt32PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *int32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt32Zero",
			expected: `{"a":0}`,
			data: &struct {
				A int32 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt32",
			expected: `{"a":1}`,
			data: &struct {
				A int32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt32Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)},
		},
		{
			name:     "PtrHeadInt32PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *int32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt32Nil",
			expected: `null`,
			data: (*struct {
				A *int32 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadInt32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt32MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},
		{
			name:     "HeadInt32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt32MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},
		{
			name:     "PtrHeadInt32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadInt32ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A int32 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt32NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A int32 `json:"a"`
				}
			}{A: struct {
				A int32 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt32PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *int32 `json:"a"`
				}
			}{A: struct {
				A *int32 `json:"a"`
			}{int32ptr(1)}},
		},
		{
			name:     "HeadInt32PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *int32 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt32ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A int32 `json:"a"`
				}
			}{A: new(struct {
				A int32 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt32NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A int32 `json:"a"`
				}
			}{A: &(struct {
				A int32 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt32PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *int32 `json:"a"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)})},
		},
		{
			name:     "PtrHeadInt32PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *int32 `json:"a"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt32NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *int32 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadInt32ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadInt32MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{A: struct {
				A int32 `json:"a"`
			}{A: 1}, B: struct {
				B int32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadInt32PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *int32 `json:"a"`
				}
				B struct {
					B *int32 `json:"b"`
				}
			}{A: struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)}, B: struct {
				B *int32 `json:"b"`
			}{B: int32ptr(2)}},
		},
		{
			name:     "HeadInt32PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *int32 `json:"a"`
				}
				B struct {
					B *int32 `json:"b"`
				}
			}{A: struct {
				A *int32 `json:"a"`
			}{A: nil}, B: struct {
				B *int32 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadInt32ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt32MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{A: struct {
				A int32 `json:"a"`
			}{A: 1}, B: struct {
				B int32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt32PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *int32 `json:"a"`
				}
				B *struct {
					B *int32 `json:"b"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)}), B: &(struct {
				B *int32 `json:"b"`
			}{B: int32ptr(2)})},
		},
		{
			name:     "PtrHeadInt32PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int32 `json:"a"`
				}
				B *struct {
					B *int32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a"`
				}
				B *struct {
					B *int32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
				B *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
			}{A: &(struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
				B *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
				B *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
				B *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(1), B: int32ptr(2)}), B: &(struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(3), B: int32ptr(4)})},
		},
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
				B *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
				B *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadInt32",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt32
				B int32 `json:"b"`
			}{
				structInt32: structInt32{A: 1},
				B:           2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt32",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt32
				B int32 `json:"b"`
			}{
				structInt32: &structInt32{A: 1},
				B:           2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32",
			expected: `{"b":2}`,
			data: struct {
				*structInt32
				B int32 `json:"b"`
			}{
				structInt32: nil,
				B:           2,
			},
		},
		{
			name:     "AnonymousHeadInt32Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: structInt32Ptr{A: int32ptr(1)},
				B:              int32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt32PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: structInt32Ptr{A: nil},
				B:              int32ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt32Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: &structInt32Ptr{A: int32ptr(1)},
				B:              int32ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: nil,
				B:              int32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt32Only",
			expected: `{"a":1}`,
			data: struct {
				structInt32
			}{
				structInt32: structInt32{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt32Only",
			expected: `{"a":1}`,
			data: struct {
				*structInt32
			}{
				structInt32: &structInt32{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32Only",
			expected: `{}`,
			data: struct {
				*structInt32
			}{
				structInt32: nil,
			},
		},
		{
			name:     "AnonymousHeadInt32PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structInt32Ptr
			}{
				structInt32Ptr: structInt32Ptr{A: int32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt32PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structInt32Ptr
			}{
				structInt32Ptr: structInt32Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadInt32PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structInt32Ptr
			}{
				structInt32Ptr: &structInt32Ptr{A: int32ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32PtrOnly",
			expected: `{}`,
			data: struct {
				*structInt32Ptr
			}{
				structInt32Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadInt64(t *testing.T) {
	type structInt64 struct {
		A int64 `json:"a"`
	}
	type structInt64Ptr struct {
		A *int64 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadInt64Zero",
			expected: `{"a":0}`,
			data: struct {
				A int64 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt64",
			expected: `{"a":1}`,
			data: struct {
				A int64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt64Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)},
		},
		{
			name:     "HeadInt64PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *int64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt64Zero",
			expected: `{"a":0}`,
			data: &struct {
				A int64 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt64",
			expected: `{"a":1}`,
			data: &struct {
				A int64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt64Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)},
		},
		{
			name:     "PtrHeadInt64PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *int64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt64Nil",
			expected: `null`,
			data: (*struct {
				A *int64 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadInt64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt64MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},
		{
			name:     "HeadInt64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt64MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},
		{
			name:     "PtrHeadInt64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadInt64ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A int64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt64NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A int64 `json:"a"`
				}
			}{A: struct {
				A int64 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt64PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *int64 `json:"a"`
				}
			}{A: struct {
				A *int64 `json:"a"`
			}{int64ptr(1)}},
		},
		{
			name:     "HeadInt64PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *int64 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt64ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A int64 `json:"a"`
				}
			}{A: new(struct {
				A int64 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt64NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A int64 `json:"a"`
				}
			}{A: &(struct {
				A int64 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt64PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *int64 `json:"a"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)})},
		},
		{
			name:     "PtrHeadInt64PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *int64 `json:"a"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt64NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *int64 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadInt64ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadInt64MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{A: struct {
				A int64 `json:"a"`
			}{A: 1}, B: struct {
				B int64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadInt64PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *int64 `json:"a"`
				}
				B struct {
					B *int64 `json:"b"`
				}
			}{A: struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)}, B: struct {
				B *int64 `json:"b"`
			}{B: int64ptr(2)}},
		},
		{
			name:     "HeadInt64PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *int64 `json:"a"`
				}
				B struct {
					B *int64 `json:"b"`
				}
			}{A: struct {
				A *int64 `json:"a"`
			}{A: nil}, B: struct {
				B *int64 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadInt64ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt64MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{A: struct {
				A int64 `json:"a"`
			}{A: 1}, B: struct {
				B int64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt64PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *int64 `json:"a"`
				}
				B *struct {
					B *int64 `json:"b"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)}), B: &(struct {
				B *int64 `json:"b"`
			}{B: int64ptr(2)})},
		},
		{
			name:     "PtrHeadInt64PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int64 `json:"a"`
				}
				B *struct {
					B *int64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a"`
				}
				B *struct {
					B *int64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
				B *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
			}{A: &(struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
				B *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
				B *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
				B *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(1), B: int64ptr(2)}), B: &(struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(3), B: int64ptr(4)})},
		},
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
				B *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
				B *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadInt64",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt64
				B int64 `json:"b"`
			}{
				structInt64: structInt64{A: 1},
				B:           2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt64",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt64
				B int64 `json:"b"`
			}{
				structInt64: &structInt64{A: 1},
				B:           2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64",
			expected: `{"b":2}`,
			data: struct {
				*structInt64
				B int64 `json:"b"`
			}{
				structInt64: nil,
				B:           2,
			},
		},
		{
			name:     "AnonymousHeadInt64Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: structInt64Ptr{A: int64ptr(1)},
				B:              int64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt64PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: structInt64Ptr{A: nil},
				B:              int64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt64Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: &structInt64Ptr{A: int64ptr(1)},
				B:              int64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: nil,
				B:              int64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt64Only",
			expected: `{"a":1}`,
			data: struct {
				structInt64
			}{
				structInt64: structInt64{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt64Only",
			expected: `{"a":1}`,
			data: struct {
				*structInt64
			}{
				structInt64: &structInt64{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64Only",
			expected: `{}`,
			data: struct {
				*structInt64
			}{
				structInt64: nil,
			},
		},
		{
			name:     "AnonymousHeadInt64PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structInt64Ptr
			}{
				structInt64Ptr: structInt64Ptr{A: int64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt64PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structInt64Ptr
			}{
				structInt64Ptr: structInt64Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadInt64PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structInt64Ptr
			}{
				structInt64Ptr: &structInt64Ptr{A: int64ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64PtrOnly",
			expected: `{}`,
			data: struct {
				*structInt64Ptr
			}{
				structInt64Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadUint(t *testing.T) {
	type structUint struct {
		A uint `json:"a"`
	}
	type structUintPtr struct {
		A *uint `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadUintZero",
			expected: `{"a":0}`,
			data: struct {
				A uint `json:"a"`
			}{},
		},
		{
			name:     "HeadUint",
			expected: `{"a":1}`,
			data: struct {
				A uint `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUintPtr",
			expected: `{"a":1}`,
			data: struct {
				A *uint `json:"a"`
			}{A: uptr(1)},
		},
		{
			name:     "HeadUintPtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *uint `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUintZero",
			expected: `{"a":0}`,
			data: &struct {
				A uint `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint",
			expected: `{"a":1}`,
			data: &struct {
				A uint `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUintPtr",
			expected: `{"a":1}`,
			data: &struct {
				A *uint `json:"a"`
			}{A: uptr(1)},
		},
		{
			name:     "PtrHeadUintPtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *uint `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUintNil",
			expected: `null`,
			data: (*struct {
				A *uint `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUintZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{},
		},
		{
			name:     "HeadUintMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUintPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(1), B: uptr(2)},
		},
		{
			name:     "HeadUintPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUintMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUintPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(1), B: uptr(2)},
		},
		{
			name:     "PtrHeadUintPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintNilMultiFields",
			expected: `null`,
			data: (*struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			})(nil),
		},

		{
			name:     "HeadUintZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A uint `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUintNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A uint `json:"a"`
				}
			}{A: struct {
				A uint `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUintPtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *uint `json:"a"`
				}
			}{A: struct {
				A *uint `json:"a"`
			}{uptr(1)}},
		},
		{
			name:     "HeadUintPtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *uint `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUintZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A uint `json:"a"`
				}
			}{A: new(struct {
				A uint `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUintNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A uint `json:"a"`
				}
			}{A: &(struct {
				A uint `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUintPtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *uint `json:"a"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
			}{A: uptr(1)})},
		},
		{
			name:     "PtrHeadUintPtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *uint `json:"a"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUintNilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *uint `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUintZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadUintMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{A: struct {
				A uint `json:"a"`
			}{A: 1}, B: struct {
				B uint `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadUintPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *uint `json:"a"`
				}
				B struct {
					B *uint `json:"b"`
				}
			}{A: struct {
				A *uint `json:"a"`
			}{A: uptr(1)}, B: struct {
				B *uint `json:"b"`
			}{B: uptr(2)}},
		},
		{
			name:     "HeadUintPtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *uint `json:"a"`
				}
				B struct {
					B *uint `json:"b"`
				}
			}{A: struct {
				A *uint `json:"a"`
			}{A: nil}, B: struct {
				B *uint `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadUintZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadUintMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{A: struct {
				A uint `json:"a"`
			}{A: 1}, B: struct {
				B uint `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUintPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *uint `json:"a"`
				}
				B *struct {
					B *uint `json:"b"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
			}{A: uptr(1)}), B: &(struct {
				B *uint `json:"b"`
			}{B: uptr(2)})},
		},
		{
			name:     "PtrHeadUintPtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint `json:"a"`
				}
				B *struct {
					B *uint `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintNilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint `json:"a"`
				}
				B *struct {
					B *uint `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
				B *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
			}{A: &(struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
				B *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
				B *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintPtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
				B *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(1), B: uptr(2)}), B: &(struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(3), B: uptr(4)})},
		},
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
				B *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
				B *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadUint",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint
				B uint `json:"b"`
			}{
				structUint: structUint{A: 1},
				B:          2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint
				B uint `json:"b"`
			}{
				structUint: &structUint{A: 1},
				B:          2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint",
			expected: `{"b":2}`,
			data: struct {
				*structUint
				B uint `json:"b"`
			}{
				structUint: nil,
				B:          2,
			},
		},
		{
			name:     "AnonymousHeadUintPtr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: structUintPtr{A: uptr(1)},
				B:             uptr(2),
			},
		},
		{
			name:     "AnonymousHeadUintPtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: structUintPtr{A: nil},
				B:             uptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUintPtr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: &structUintPtr{A: uptr(1)},
				B:             uptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintPtr",
			expected: `{"b":2}`,
			data: struct {
				*structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: nil,
				B:             uptr(2),
			},
		},
		{
			name:     "AnonymousHeadUintOnly",
			expected: `{"a":1}`,
			data: struct {
				structUint
			}{
				structUint: structUint{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUintOnly",
			expected: `{"a":1}`,
			data: struct {
				*structUint
			}{
				structUint: &structUint{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintOnly",
			expected: `{}`,
			data: struct {
				*structUint
			}{
				structUint: nil,
			},
		},
		{
			name:     "AnonymousHeadUintPtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structUintPtr
			}{
				structUintPtr: structUintPtr{A: uptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUintPtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structUintPtr
			}{
				structUintPtr: structUintPtr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUintPtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structUintPtr
			}{
				structUintPtr: &structUintPtr{A: uptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintPtrOnly",
			expected: `{}`,
			data: struct {
				*structUintPtr
			}{
				structUintPtr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadUint8(t *testing.T) {
	type structUint8 struct {
		A uint8 `json:"a"`
	}
	type structUint8Ptr struct {
		A *uint8 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadUint8Zero",
			expected: `{"a":0}`,
			data: struct {
				A uint8 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint8",
			expected: `{"a":1}`,
			data: struct {
				A uint8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint8Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *uint8 `json:"a"`
			}{A: uint8ptr(1)},
		},
		{
			name:     "HeadUint8PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *uint8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint8Zero",
			expected: `{"a":0}`,
			data: &struct {
				A uint8 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint8",
			expected: `{"a":1}`,
			data: &struct {
				A uint8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint8Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *uint8 `json:"a"`
			}{A: uint8ptr(1)},
		},
		{
			name:     "PtrHeadUint8PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *uint8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint8Nil",
			expected: `null`,
			data: (*struct {
				A *uint8 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint8MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},
		{
			name:     "HeadUint8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint8MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},
		{
			name:     "PtrHeadUint8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint8ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A uint8 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint8NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A uint8 `json:"a"`
				}
			}{A: struct {
				A uint8 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint8PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *uint8 `json:"a"`
				}
			}{A: struct {
				A *uint8 `json:"a"`
			}{uint8ptr(1)}},
		},
		{
			name:     "HeadUint8PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *uint8 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint8ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A uint8 `json:"a"`
				}
			}{A: new(struct {
				A uint8 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint8NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A uint8 `json:"a"`
				}
			}{A: &(struct {
				A uint8 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint8PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *uint8 `json:"a"`
				}
			}{A: &(struct {
				A *uint8 `json:"a"`
			}{A: uint8ptr(1)})},
		},
		{
			name:     "PtrHeadUint8PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *uint8 `json:"a"`
				}
			}{A: &(struct {
				A *uint8 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint8NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *uint8 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint8ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A uint8 `json:"a"`
				}
				B struct {
					B uint8 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadUint8MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A uint8 `json:"a"`
				}
				B struct {
					B uint8 `json:"b"`
				}
			}{A: struct {
				A uint8 `json:"a"`
			}{A: 1}, B: struct {
				B uint8 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadUint8PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *uint8 `json:"a"`
				}
				B struct {
					B *uint8 `json:"b"`
				}
			}{A: struct {
				A *uint8 `json:"a"`
			}{A: uint8ptr(1)}, B: struct {
				B *uint8 `json:"b"`
			}{B: uint8ptr(2)}},
		},
		{
			name:     "HeadUint8PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *uint8 `json:"a"`
				}
				B struct {
					B *uint8 `json:"b"`
				}
			}{A: struct {
				A *uint8 `json:"a"`
			}{A: nil}, B: struct {
				B *uint8 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadUint8ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A uint8 `json:"a"`
				}
				B struct {
					B uint8 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint8MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A uint8 `json:"a"`
				}
				B struct {
					B uint8 `json:"b"`
				}
			}{A: struct {
				A uint8 `json:"a"`
			}{A: 1}, B: struct {
				B uint8 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint8PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a"`
				}
				B *struct {
					B *uint8 `json:"b"`
				}
			}{A: &(struct {
				A *uint8 `json:"a"`
			}{A: uint8ptr(1)}), B: &(struct {
				B *uint8 `json:"b"`
			}{B: uint8ptr(2)})},
		},
		{
			name:     "PtrHeadUint8PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a"`
				}
				B *struct {
					B *uint8 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint8 `json:"a"`
				}
				B *struct {
					B *uint8 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint8DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A uint8 `json:"a"`
					B uint8 `json:"b"`
				}
				B *struct {
					A uint8 `json:"a"`
					B uint8 `json:"b"`
				}
			}{A: &(struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint8NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A uint8 `json:"a"`
					B uint8 `json:"b"`
				}
				B *struct {
					A uint8 `json:"a"`
					B uint8 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A uint8 `json:"a"`
					B uint8 `json:"b"`
				}
				B *struct {
					A uint8 `json:"a"`
					B uint8 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint8PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a"`
					B *uint8 `json:"b"`
				}
				B *struct {
					A *uint8 `json:"a"`
					B *uint8 `json:"b"`
				}
			}{A: &(struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: uint8ptr(1), B: uint8ptr(2)}), B: &(struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: uint8ptr(3), B: uint8ptr(4)})},
		},
		{
			name:     "PtrHeadUint8PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a"`
					B *uint8 `json:"b"`
				}
				B *struct {
					A *uint8 `json:"a"`
					B *uint8 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint8 `json:"a"`
					B *uint8 `json:"b"`
				}
				B *struct {
					A *uint8 `json:"a"`
					B *uint8 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadUint8",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint8
				B uint8 `json:"b"`
			}{
				structUint8: structUint8{A: 1},
				B:           2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint8",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint8
				B uint8 `json:"b"`
			}{
				structUint8: &structUint8{A: 1},
				B:           2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8",
			expected: `{"b":2}`,
			data: struct {
				*structUint8
				B uint8 `json:"b"`
			}{
				structUint8: nil,
				B:           2,
			},
		},
		{
			name:     "AnonymousHeadUint8Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint8Ptr
				B *uint8 `json:"b"`
			}{
				structUint8Ptr: structUint8Ptr{A: uint8ptr(1)},
				B:              uint8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint8PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structUint8Ptr
				B *uint8 `json:"b"`
			}{
				structUint8Ptr: structUint8Ptr{A: nil},
				B:              uint8ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint8Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint8Ptr
				B *uint8 `json:"b"`
			}{
				structUint8Ptr: &structUint8Ptr{A: uint8ptr(1)},
				B:              uint8ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structUint8Ptr
				B *uint8 `json:"b"`
			}{
				structUint8Ptr: nil,
				B:              uint8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint8Only",
			expected: `{"a":1}`,
			data: struct {
				structUint8
			}{
				structUint8: structUint8{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint8Only",
			expected: `{"a":1}`,
			data: struct {
				*structUint8
			}{
				structUint8: &structUint8{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8Only",
			expected: `{}`,
			data: struct {
				*structUint8
			}{
				structUint8: nil,
			},
		},
		{
			name:     "AnonymousHeadUint8PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structUint8Ptr
			}{
				structUint8Ptr: structUint8Ptr{A: uint8ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint8PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structUint8Ptr
			}{
				structUint8Ptr: structUint8Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint8PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structUint8Ptr
			}{
				structUint8Ptr: &structUint8Ptr{A: uint8ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8PtrOnly",
			expected: `{}`,
			data: struct {
				*structUint8Ptr
			}{
				structUint8Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadUint16(t *testing.T) {
	type structUint16 struct {
		A uint16 `json:"a"`
	}
	type structUint16Ptr struct {
		A *uint16 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadUint16Zero",
			expected: `{"a":0}`,
			data: struct {
				A uint16 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint16",
			expected: `{"a":1}`,
			data: struct {
				A uint16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint16Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)},
		},
		{
			name:     "HeadUint16PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *uint16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint16Zero",
			expected: `{"a":0}`,
			data: &struct {
				A uint16 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint16",
			expected: `{"a":1}`,
			data: &struct {
				A uint16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint16Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)},
		},
		{
			name:     "PtrHeadUint16PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *uint16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint16Nil",
			expected: `null`,
			data: (*struct {
				A *uint16 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint16MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},
		{
			name:     "HeadUint16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint16MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},
		{
			name:     "PtrHeadUint16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint16ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A uint16 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint16NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A uint16 `json:"a"`
				}
			}{A: struct {
				A uint16 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint16PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *uint16 `json:"a"`
				}
			}{A: struct {
				A *uint16 `json:"a"`
			}{uint16ptr(1)}},
		},
		{
			name:     "HeadUint16PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *uint16 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint16ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A uint16 `json:"a"`
				}
			}{A: new(struct {
				A uint16 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint16NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A uint16 `json:"a"`
				}
			}{A: &(struct {
				A uint16 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint16PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *uint16 `json:"a"`
				}
			}{A: &(struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)})},
		},
		{
			name:     "PtrHeadUint16PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *uint16 `json:"a"`
				}
			}{A: &(struct {
				A *uint16 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint16NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *uint16 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint16ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A uint16 `json:"a"`
				}
				B struct {
					B uint16 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadUint16MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A uint16 `json:"a"`
				}
				B struct {
					B uint16 `json:"b"`
				}
			}{A: struct {
				A uint16 `json:"a"`
			}{A: 1}, B: struct {
				B uint16 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadUint16PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *uint16 `json:"a"`
				}
				B struct {
					B *uint16 `json:"b"`
				}
			}{A: struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)}, B: struct {
				B *uint16 `json:"b"`
			}{B: uint16ptr(2)}},
		},
		{
			name:     "HeadUint16PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *uint16 `json:"a"`
				}
				B struct {
					B *uint16 `json:"b"`
				}
			}{A: struct {
				A *uint16 `json:"a"`
			}{A: nil}, B: struct {
				B *uint16 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadUint16ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A uint16 `json:"a"`
				}
				B struct {
					B uint16 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint16MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A uint16 `json:"a"`
				}
				B struct {
					B uint16 `json:"b"`
				}
			}{A: struct {
				A uint16 `json:"a"`
			}{A: 1}, B: struct {
				B uint16 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint16PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a"`
				}
				B *struct {
					B *uint16 `json:"b"`
				}
			}{A: &(struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)}), B: &(struct {
				B *uint16 `json:"b"`
			}{B: uint16ptr(2)})},
		},
		{
			name:     "PtrHeadUint16PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a"`
				}
				B *struct {
					B *uint16 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint16 `json:"a"`
				}
				B *struct {
					B *uint16 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint16DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A uint16 `json:"a"`
					B uint16 `json:"b"`
				}
				B *struct {
					A uint16 `json:"a"`
					B uint16 `json:"b"`
				}
			}{A: &(struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint16NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A uint16 `json:"a"`
					B uint16 `json:"b"`
				}
				B *struct {
					A uint16 `json:"a"`
					B uint16 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A uint16 `json:"a"`
					B uint16 `json:"b"`
				}
				B *struct {
					A uint16 `json:"a"`
					B uint16 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint16PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a"`
					B *uint16 `json:"b"`
				}
				B *struct {
					A *uint16 `json:"a"`
					B *uint16 `json:"b"`
				}
			}{A: &(struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: uint16ptr(1), B: uint16ptr(2)}), B: &(struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: uint16ptr(3), B: uint16ptr(4)})},
		},
		{
			name:     "PtrHeadUint16PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a"`
					B *uint16 `json:"b"`
				}
				B *struct {
					A *uint16 `json:"a"`
					B *uint16 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint16 `json:"a"`
					B *uint16 `json:"b"`
				}
				B *struct {
					A *uint16 `json:"a"`
					B *uint16 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadUint16",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint16
				B uint16 `json:"b"`
			}{
				structUint16: structUint16{A: 1},
				B:            2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint16",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint16
				B uint16 `json:"b"`
			}{
				structUint16: &structUint16{A: 1},
				B:            2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16",
			expected: `{"b":2}`,
			data: struct {
				*structUint16
				B uint16 `json:"b"`
			}{
				structUint16: nil,
				B:            2,
			},
		},
		{
			name:     "AnonymousHeadUint16Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint16Ptr
				B *uint16 `json:"b"`
			}{
				structUint16Ptr: structUint16Ptr{A: uint16ptr(1)},
				B:               uint16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint16PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structUint16Ptr
				B *uint16 `json:"b"`
			}{
				structUint16Ptr: structUint16Ptr{A: nil},
				B:               uint16ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint16Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint16Ptr
				B *uint16 `json:"b"`
			}{
				structUint16Ptr: &structUint16Ptr{A: uint16ptr(1)},
				B:               uint16ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structUint16Ptr
				B *uint16 `json:"b"`
			}{
				structUint16Ptr: nil,
				B:               uint16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint16Only",
			expected: `{"a":1}`,
			data: struct {
				structUint16
			}{
				structUint16: structUint16{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint16Only",
			expected: `{"a":1}`,
			data: struct {
				*structUint16
			}{
				structUint16: &structUint16{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16Only",
			expected: `{}`,
			data: struct {
				*structUint16
			}{
				structUint16: nil,
			},
		},
		{
			name:     "AnonymousHeadUint16PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structUint16Ptr
			}{
				structUint16Ptr: structUint16Ptr{A: uint16ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint16PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structUint16Ptr
			}{
				structUint16Ptr: structUint16Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint16PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structUint16Ptr
			}{
				structUint16Ptr: &structUint16Ptr{A: uint16ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16PtrOnly",
			expected: `{}`,
			data: struct {
				*structUint16Ptr
			}{
				structUint16Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadUint32(t *testing.T) {
	type structUint32 struct {
		A uint32 `json:"a"`
	}
	type structUint32Ptr struct {
		A *uint32 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadUint32Zero",
			expected: `{"a":0}`,
			data: struct {
				A uint32 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint32",
			expected: `{"a":1}`,
			data: struct {
				A uint32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint32Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)},
		},
		{
			name:     "HeadUint32PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *uint32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint32Zero",
			expected: `{"a":0}`,
			data: &struct {
				A uint32 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint32",
			expected: `{"a":1}`,
			data: &struct {
				A uint32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint32Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)},
		},
		{
			name:     "PtrHeadUint32PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *uint32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint32Nil",
			expected: `null`,
			data: (*struct {
				A *uint32 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint32MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},
		{
			name:     "HeadUint32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint32MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},
		{
			name:     "PtrHeadUint32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint32ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A uint32 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint32NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A uint32 `json:"a"`
				}
			}{A: struct {
				A uint32 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint32PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *uint32 `json:"a"`
				}
			}{A: struct {
				A *uint32 `json:"a"`
			}{uint32ptr(1)}},
		},
		{
			name:     "HeadUint32PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *uint32 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint32ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A uint32 `json:"a"`
				}
			}{A: new(struct {
				A uint32 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint32NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A uint32 `json:"a"`
				}
			}{A: &(struct {
				A uint32 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint32PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *uint32 `json:"a"`
				}
			}{A: &(struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)})},
		},
		{
			name:     "PtrHeadUint32PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *uint32 `json:"a"`
				}
			}{A: &(struct {
				A *uint32 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint32NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *uint32 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint32ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A uint32 `json:"a"`
				}
				B struct {
					B uint32 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadUint32MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A uint32 `json:"a"`
				}
				B struct {
					B uint32 `json:"b"`
				}
			}{A: struct {
				A uint32 `json:"a"`
			}{A: 1}, B: struct {
				B uint32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadUint32PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *uint32 `json:"a"`
				}
				B struct {
					B *uint32 `json:"b"`
				}
			}{A: struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)}, B: struct {
				B *uint32 `json:"b"`
			}{B: uint32ptr(2)}},
		},
		{
			name:     "HeadUint32PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *uint32 `json:"a"`
				}
				B struct {
					B *uint32 `json:"b"`
				}
			}{A: struct {
				A *uint32 `json:"a"`
			}{A: nil}, B: struct {
				B *uint32 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadUint32ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A uint32 `json:"a"`
				}
				B struct {
					B uint32 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint32MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A uint32 `json:"a"`
				}
				B struct {
					B uint32 `json:"b"`
				}
			}{A: struct {
				A uint32 `json:"a"`
			}{A: 1}, B: struct {
				B uint32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint32PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a"`
				}
				B *struct {
					B *uint32 `json:"b"`
				}
			}{A: &(struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)}), B: &(struct {
				B *uint32 `json:"b"`
			}{B: uint32ptr(2)})},
		},
		{
			name:     "PtrHeadUint32PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a"`
				}
				B *struct {
					B *uint32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint32 `json:"a"`
				}
				B *struct {
					B *uint32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint32DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A uint32 `json:"a"`
					B uint32 `json:"b"`
				}
				B *struct {
					A uint32 `json:"a"`
					B uint32 `json:"b"`
				}
			}{A: &(struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint32NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A uint32 `json:"a"`
					B uint32 `json:"b"`
				}
				B *struct {
					A uint32 `json:"a"`
					B uint32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A uint32 `json:"a"`
					B uint32 `json:"b"`
				}
				B *struct {
					A uint32 `json:"a"`
					B uint32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint32PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a"`
					B *uint32 `json:"b"`
				}
				B *struct {
					A *uint32 `json:"a"`
					B *uint32 `json:"b"`
				}
			}{A: &(struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: uint32ptr(1), B: uint32ptr(2)}), B: &(struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: uint32ptr(3), B: uint32ptr(4)})},
		},
		{
			name:     "PtrHeadUint32PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a"`
					B *uint32 `json:"b"`
				}
				B *struct {
					A *uint32 `json:"a"`
					B *uint32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint32 `json:"a"`
					B *uint32 `json:"b"`
				}
				B *struct {
					A *uint32 `json:"a"`
					B *uint32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadUint32",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint32
				B uint32 `json:"b"`
			}{
				structUint32: structUint32{A: 1},
				B:            2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint32",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint32
				B uint32 `json:"b"`
			}{
				structUint32: &structUint32{A: 1},
				B:            2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32",
			expected: `{"b":2}`,
			data: struct {
				*structUint32
				B uint32 `json:"b"`
			}{
				structUint32: nil,
				B:            2,
			},
		},
		{
			name:     "AnonymousHeadUint32Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint32Ptr
				B *uint32 `json:"b"`
			}{
				structUint32Ptr: structUint32Ptr{A: uint32ptr(1)},
				B:               uint32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint32PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structUint32Ptr
				B *uint32 `json:"b"`
			}{
				structUint32Ptr: structUint32Ptr{A: nil},
				B:               uint32ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint32Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint32Ptr
				B *uint32 `json:"b"`
			}{
				structUint32Ptr: &structUint32Ptr{A: uint32ptr(1)},
				B:               uint32ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structUint32Ptr
				B *uint32 `json:"b"`
			}{
				structUint32Ptr: nil,
				B:               uint32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint32Only",
			expected: `{"a":1}`,
			data: struct {
				structUint32
			}{
				structUint32: structUint32{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint32Only",
			expected: `{"a":1}`,
			data: struct {
				*structUint32
			}{
				structUint32: &structUint32{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32Only",
			expected: `{}`,
			data: struct {
				*structUint32
			}{
				structUint32: nil,
			},
		},
		{
			name:     "AnonymousHeadUint32PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structUint32Ptr
			}{
				structUint32Ptr: structUint32Ptr{A: uint32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint32PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structUint32Ptr
			}{
				structUint32Ptr: structUint32Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint32PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structUint32Ptr
			}{
				structUint32Ptr: &structUint32Ptr{A: uint32ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32PtrOnly",
			expected: `{}`,
			data: struct {
				*structUint32Ptr
			}{
				structUint32Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}

func TestCoverStructHeadUint64(t *testing.T) {
	type structUint64 struct {
		A uint64 `json:"a"`
	}
	type structUint64Ptr struct {
		A *uint64 `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadUint64Zero",
			expected: `{"a":0}`,
			data: struct {
				A uint64 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint64",
			expected: `{"a":1}`,
			data: struct {
				A uint64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint64Ptr",
			expected: `{"a":1}`,
			data: struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "HeadUint64PtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *uint64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64Zero",
			expected: `{"a":0}`,
			data: &struct {
				A uint64 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint64",
			expected: `{"a":1}`,
			data: &struct {
				A uint64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint64Ptr",
			expected: `{"a":1}`,
			data: &struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "PtrHeadUint64PtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *uint64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64Nil",
			expected: `null`,
			data: (*struct {
				A *uint64 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint64MultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "HeadUint64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			data: &struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint64MultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			data: &struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "PtrHeadUint64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64NilMultiFields",
			expected: `null`,
			data: (*struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint64ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A struct {
					A uint64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint64NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A uint64 `json:"a"`
				}
			}{A: struct {
				A uint64 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint64PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A struct {
					A *uint64 `json:"a"`
				}
			}{A: struct {
				A *uint64 `json:"a"`
			}{uint64ptr(1)}},
		},
		{
			name:     "HeadUint64PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *uint64 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			data: struct {
				A *struct {
					A uint64 `json:"a"`
				}
			}{A: new(struct {
				A uint64 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint64NotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A uint64 `json:"a"`
				}
			}{A: &(struct {
				A uint64 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint64PtrNotRoot",
			expected: `{"A":{"a":1}}`,
			data: struct {
				A *struct {
					A *uint64 `json:"a"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)})},
		},
		{
			name:     "PtrHeadUint64PtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *uint64 `json:"a"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint64NilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *uint64 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint64ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: struct {
				A struct {
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadUint64MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{A: struct {
				A uint64 `json:"a"`
			}{A: 1}, B: struct {
				B uint64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadUint64PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: struct {
				A struct {
					A *uint64 `json:"a"`
				}
				B struct {
					B *uint64 `json:"b"`
				}
			}{A: struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)}, B: struct {
				B *uint64 `json:"b"`
			}{B: uint64ptr(2)}},
		},
		{
			name:     "HeadUint64PtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *uint64 `json:"a"`
				}
				B struct {
					B *uint64 `json:"b"`
				}
			}{A: struct {
				A *uint64 `json:"a"`
			}{A: nil}, B: struct {
				B *uint64 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadUint64ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			data: &struct {
				A struct {
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint64MultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A struct {
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{A: struct {
				A uint64 `json:"a"`
			}{A: 1}, B: struct {
				B uint64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint64PtrMultiFieldsNotRoot",
			expected: `{"A":{"a":1},"B":{"b":2}}`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a"`
				}
				B *struct {
					B *uint64 `json:"b"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)}), B: &(struct {
				B *uint64 `json:"b"`
			}{B: uint64ptr(2)})},
		},
		{
			name:     "PtrHeadUint64PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a"`
				}
				B *struct {
					B *uint64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64NilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a"`
				}
				B *struct {
					B *uint64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64DoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
				B *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
			}{A: &(struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
				B *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
				B *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64PtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":1,"b":2},"B":{"a":3,"b":4}}`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
				B *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)}), B: &(struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(3), B: uint64ptr(4)})},
		},
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
				B *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
				B *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadUint64",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint64
				B uint64 `json:"b"`
			}{
				structUint64: structUint64{A: 1},
				B:            2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint64",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint64
				B uint64 `json:"b"`
			}{
				structUint64: &structUint64{A: 1},
				B:            2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64",
			expected: `{"b":2}`,
			data: struct {
				*structUint64
				B uint64 `json:"b"`
			}{
				structUint64: nil,
				B:            2,
			},
		},
		{
			name:     "AnonymousHeadUint64Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: structUint64Ptr{A: uint64ptr(1)},
				B:               uint64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint64PtrNil",
			expected: `{"a":null,"b":2}`,
			data: struct {
				structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: structUint64Ptr{A: nil},
				B:               uint64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint64Ptr",
			expected: `{"a":1,"b":2}`,
			data: struct {
				*structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: &structUint64Ptr{A: uint64ptr(1)},
				B:               uint64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64Ptr",
			expected: `{"b":2}`,
			data: struct {
				*structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: nil,
				B:               uint64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint64Only",
			expected: `{"a":1}`,
			data: struct {
				structUint64
			}{
				structUint64: structUint64{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64Only",
			expected: `{"a":1}`,
			data: struct {
				*structUint64
			}{
				structUint64: &structUint64{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64Only",
			expected: `{}`,
			data: struct {
				*structUint64
			}{
				structUint64: nil,
			},
		},
		{
			name:     "AnonymousHeadUint64PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				structUint64Ptr
			}{
				structUint64Ptr: structUint64Ptr{A: uint64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint64PtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structUint64Ptr
			}{
				structUint64Ptr: structUint64Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64PtrOnly",
			expected: `{"a":1}`,
			data: struct {
				*structUint64Ptr
			}{
				structUint64Ptr: &structUint64Ptr{A: uint64ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64PtrOnly",
			expected: `{}`,
			data: struct {
				*structUint64Ptr
			}{
				structUint64Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}
