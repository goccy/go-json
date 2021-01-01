package json

import (
	"bytes"
	"strings"
	"testing"
)

func intptr(v int) *int {
	return &v
}

func int8ptr(v int8) *int8 {
	return &v
}

func headIntPtrNilNotRoot() interface{} {
	v := struct {
		A struct {
			A *int `json:"a"`
		}
	}{}
	return v
}

func ptrHeadIntNotRoot() interface{} {
	v := struct {
		A *struct {
			A int `json:"a"`
		}
	}{A: new(struct {
		A int `json:"a"`
	})}
	v.A.A = 1
	return v
}

func TestCoverage(t *testing.T) {
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
			name:     "HeadIntPtrNilMultiFieldsd",
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
		/*
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
		*/
		/*
			{
				name:     "PtrHeadIntNilNotRoot",
				expected: `{"A":null}`,
				data: struct {
					A *struct {
						A *int `json:"a"`
					}
				}{A: nil},
			},
		*/
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
			name:     "HeadIntPtrNilMultiFieldsd",
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
			name:     "Int8Head",
			expected: `{"a":1}`,
			data: struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "Int8PtrHead",
			expected: `{"a":1}`,
			data: struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		/*
			{
				name:     "Int8PtrNilHead",
				expected: `{"a":null}`,
				data: struct {
					A *int8 `json:"a"`
				}{A: nil},
			},
		*/
		{
			name:     "PtrInt8Head",
			expected: `{"a":1}`,
			data: &struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrInt8PtrHead",
			expected: `{"a":1}`,
			data: &struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		{
			name:     "Int8Field",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "Int8PtrField",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Errorf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Errorf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}
