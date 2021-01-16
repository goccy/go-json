package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverUint8(t *testing.T) {
	type structUint8 struct {
		A uint8 `json:"a"`
	}
	type structUint8Ptr struct {
		A *uint8 `json:"a"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		{
			name:     "HeadUint8Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A uint8 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint8",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint8Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint8 `json:"a"`
			}{A: uint8ptr(1)},
		},
		{
			name:     "HeadUint8PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint8Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A uint8 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint8",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint8Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint8 `json:"a"`
			}{A: uint8ptr(1)},
		},
		{
			name:     "PtrHeadUint8PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint8Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint8 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint8MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},
		{
			name:     "HeadUint8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint8MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint8 `json:"a"`
				B uint8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},
		{
			name:     "PtrHeadUint8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint8 `json:"a"`
				B *uint8 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint8ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			indentExpected: `
{
  "A": {
    "a": 0
  }
}
`,
			data: struct {
				A struct {
					A uint8 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint8NotRoot",
			expected: `{"A":{"a":1}}`,
			indentExpected: `
{
  "A": {
    "a": 1
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 1
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": null
  }
}
`,
			data: struct {
				A struct {
					A *uint8 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint8ZeroNotRoot",
			expected: `{"A":{"a":0}}`,
			indentExpected: `
{
  "A": {
    "a": 0
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 1
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 1
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": null
  }
}
`,
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
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint8 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint8ZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":0},"B":{"b":0}}`,
			indentExpected: `
{
  "A": {
    "a": 0
  },
  "B": {
    "b": 0
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 1
  },
  "B": {
    "b": 2
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 1
  },
  "B": {
    "b": 2
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": null
  },
  "B": {
    "b": null
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 0
  },
  "B": {
    "b": 0
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 1
  },
  "B": {
    "b": 2
  }
}
`,
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
			indentExpected: `
{
  "A": {
    "a": 1
  },
  "B": {
    "b": 2
  }
}
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": {
    "a": 1,
    "b": 2
  },
  "B": {
    "a": 3,
    "b": 4
  }
}
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": {
    "a": 1,
    "b": 2
  },
  "B": {
    "a": 3,
    "b": 4
  }
}
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint8
			}{
				structUint8: structUint8{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint8Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint8
			}{
				structUint8: &structUint8{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint8
			}{
				structUint8: nil,
			},
		},
		{
			name:     "AnonymousHeadUint8PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint8Ptr
			}{
				structUint8Ptr: structUint8Ptr{A: uint8ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint8PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint8Ptr
			}{
				structUint8Ptr: structUint8Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint8PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint8Ptr
			}{
				structUint8Ptr: &structUint8Ptr{A: uint8ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint8Ptr
			}{
				structUint8Ptr: nil,
			},
		},
	}
	for _, test := range tests {
		for _, indent := range []bool{true, false} {
			for _, htmlEscape := range []bool{true, false} {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetEscapeHTML(htmlEscape)
				if indent {
					enc.SetIndent("", "  ")
				}
				if err := enc.Encode(test.data); err != nil {
					t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
				}
				if indent {
					got := "\n" + buf.String()
					if got != test.indentExpected {
						t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.indentExpected, got)
					}
				} else {
					if strings.TrimRight(buf.String(), "\n") != test.expected {
						t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
					}
				}
			}
		}
	}
}
