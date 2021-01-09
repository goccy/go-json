package json

import (
	"bytes"
	"strings"
	"testing"
)

func intptr(v int) *int       { return &v }
func int8ptr(v int8) *int8    { return &v }
func int16ptr(v int16) *int16 { return &v }

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
