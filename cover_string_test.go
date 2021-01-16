package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverString(t *testing.T) {
	type structString struct {
		A string `json:"a"`
	}
	type structStringPtr struct {
		A *string `json:"a"`
	}

	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "HeadStringZero",
			expected: `{"a":""}`,
			data: struct {
				A string `json:"a"`
			}{},
		},
		{
			name:     "HeadString",
			expected: `{"a":"foo"}`,
			data: struct {
				A string `json:"a"`
			}{A: "foo"},
		},
		{
			name:     "HeadStringPtr",
			expected: `{"a":"foo"}`,
			data: struct {
				A *string `json:"a"`
			}{A: stringptr("foo")},
		},
		{
			name:     "HeadStringPtrNil",
			expected: `{"a":null}`,
			data: struct {
				A *string `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadStringZero",
			expected: `{"a":""}`,
			data: &struct {
				A string `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadString",
			expected: `{"a":"foo"}`,
			data: &struct {
				A string `json:"a"`
			}{A: "foo"},
		},
		{
			name:     "PtrHeadStringPtr",
			expected: `{"a":"foo"}`,
			data: &struct {
				A *string `json:"a"`
			}{A: stringptr("foo")},
		},
		{
			name:     "PtrHeadStringPtrNil",
			expected: `{"a":null}`,
			data: &struct {
				A *string `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadStringNil",
			expected: `null`,
			data: (*struct {
				A *string `json:"a"`
			})(nil),
		},
		{
			name:     "HeadStringZeroMultiFields",
			expected: `{"a":"","b":""}`,
			data: struct {
				A string `json:"a"`
				B string `json:"b"`
			}{},
		},
		{
			name:     "HeadStringMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			data: struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"},
		},
		{
			name:     "HeadStringPtrMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			data: struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},
		{
			name:     "HeadStringPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringZeroMultiFields",
			expected: `{"a":"","b":""}`,
			data: &struct {
				A string `json:"a"`
				B string `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadStringMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			data: &struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"},
		},
		{
			name:     "PtrHeadStringPtrMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			data: &struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},
		{
			name:     "PtrHeadStringPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			data: &struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringNilMultiFields",
			expected: `null`,
			data: (*struct {
				A *string `json:"a"`
				B *string `json:"b"`
			})(nil),
		},
		{
			name:     "HeadStringZeroNotRoot",
			expected: `{"A":{"a":""}}`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadStringNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
			}{A: struct {
				A string `json:"a"`
			}{A: "foo"}},
		},
		{
			name:     "HeadStringPtrNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			data: struct {
				A struct {
					A *string `json:"a"`
				}
			}{A: struct {
				A *string `json:"a"`
			}{stringptr("foo")}},
		},
		{
			name:     "HeadStringPtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A struct {
					A *string `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadStringZeroNotRoot",
			expected: `{"A":{"a":""}}`,
			data: struct {
				A *struct {
					A string `json:"a"`
				}
			}{A: new(struct {
				A string `json:"a"`
			})},
		},
		{
			name:     "PtrHeadStringNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			data: struct {
				A *struct {
					A string `json:"a"`
				}
			}{A: &(struct {
				A string `json:"a"`
			}{A: "foo"})},
		},
		{
			name:     "PtrHeadStringPtrNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			data: struct {
				A *struct {
					A *string `json:"a"`
				}
			}{A: &(struct {
				A *string `json:"a"`
			}{A: stringptr("foo")})},
		},
		{
			name:     "PtrHeadStringPtrNilNotRoot",
			expected: `{"A":{"a":null}}`,
			data: struct {
				A *struct {
					A *string `json:"a"`
				}
			}{A: &(struct {
				A *string `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadStringNilNotRoot",
			expected: `{"A":null}`,
			data: struct {
				A *struct {
					A *string `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadStringZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":""},"B":{"b":""}}`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadStringMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{A: struct {
				A string `json:"a"`
			}{A: "foo"}, B: struct {
				B string `json:"b"`
			}{B: "bar"}},
		},
		{
			name:     "HeadStringPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			data: struct {
				A struct {
					A *string `json:"a"`
				}
				B struct {
					B *string `json:"b"`
				}
			}{A: struct {
				A *string `json:"a"`
			}{A: stringptr("foo")}, B: struct {
				B *string `json:"b"`
			}{B: stringptr("bar")}},
		},
		{
			name:     "HeadStringPtrNilMultiFieldsNotRoot",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			data: struct {
				A struct {
					A *string `json:"a"`
				}
				B struct {
					B *string `json:"b"`
				}
			}{A: struct {
				A *string `json:"a"`
			}{A: nil}, B: struct {
				B *string `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadStringZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":""},"B":{"b":""}}`,
			data: &struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadStringMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			data: &struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{A: struct {
				A string `json:"a"`
			}{A: "foo"}, B: struct {
				B string `json:"b"`
			}{B: "bar"}},
		},
		{
			name:     "PtrHeadStringPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
				}
				B *struct {
					B *string `json:"b"`
				}
			}{A: &(struct {
				A *string `json:"a"`
			}{A: stringptr("foo")}), B: &(struct {
				B *string `json:"b"`
			}{B: stringptr("bar")})},
		},
		{
			name:     "PtrHeadStringPtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
				}
				B *struct {
					B *string `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringNilMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *string `json:"a"`
				}
				B *struct {
					B *string `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			data: &struct {
				A *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
				B *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
			}{A: &(struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"}), B: &(struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"})},
		},
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
				B *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
				B *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringPtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
				B *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
			}{A: &(struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")}), B: &(struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")})},
		},
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
				B *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			data: (*struct {
				A *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
				B *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadString",
			expected: `{"a":"foo","b":"bar"}`,
			data: struct {
				structString
				B string `json:"b"`
			}{
				structString: structString{A: "foo"},
				B:            "bar",
			},
		},
		{
			name:     "PtrAnonymousHeadString",
			expected: `{"a":"foo","b":"bar"}`,
			data: struct {
				*structString
				B string `json:"b"`
			}{
				structString: &structString{A: "foo"},
				B:            "bar",
			},
		},
		{
			name:     "NilPtrAnonymousHeadString",
			expected: `{"b":"baz"}`,
			data: struct {
				*structString
				B string `json:"b"`
			}{
				structString: nil,
				B:            "baz",
			},
		},
		{
			name:     "AnonymousHeadStringPtr",
			expected: `{"a":"foo","b":"bar"}`,
			data: struct {
				structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: structStringPtr{A: stringptr("foo")},
				B:               stringptr("bar"),
			},
		},
		{
			name:     "AnonymousHeadStringPtrNil",
			expected: `{"a":null,"b":"foo"}`,
			data: struct {
				structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: structStringPtr{A: nil},
				B:               stringptr("foo"),
			},
		},
		{
			name:     "PtrAnonymousHeadStringPtr",
			expected: `{"a":"foo","b":"bar"}`,
			data: struct {
				*structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: &structStringPtr{A: stringptr("foo")},
				B:               stringptr("bar"),
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringPtr",
			expected: `{"b":"foo"}`,
			data: struct {
				*structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: nil,
				B:               stringptr("foo"),
			},
		},
		{
			name:     "AnonymousHeadStringOnly",
			expected: `{"a":"foo"}`,
			data: struct {
				structString
			}{
				structString: structString{A: "foo"},
			},
		},
		{
			name:     "PtrAnonymousHeadStringOnly",
			expected: `{"a":"foo"}`,
			data: struct {
				*structString
			}{
				structString: &structString{A: "foo"},
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringOnly",
			expected: `{}`,
			data: struct {
				*structString
			}{
				structString: nil,
			},
		},
		{
			name:     "AnonymousHeadStringPtrOnly",
			expected: `{"a":"foo"}`,
			data: struct {
				structStringPtr
			}{
				structStringPtr: structStringPtr{A: stringptr("foo")},
			},
		},
		{
			name:     "AnonymousHeadStringPtrNilOnly",
			expected: `{"a":null}`,
			data: struct {
				structStringPtr
			}{
				structStringPtr: structStringPtr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadStringPtrOnly",
			expected: `{"a":"foo"}`,
			data: struct {
				*structStringPtr
			}{
				structStringPtr: &structStringPtr{A: stringptr("foo")},
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringPtrOnly",
			expected: `{}`,
			data: struct {
				*structStringPtr
			}{
				structStringPtr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
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
