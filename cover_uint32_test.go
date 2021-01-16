package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverUint32(t *testing.T) {
	type structUint32 struct {
		A uint32 `json:"a"`
	}
	type structUint32Ptr struct {
		A *uint32 `json:"a"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		{
			name:     "HeadUint32Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A uint32 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint32",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint32Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)},
		},
		{
			name:     "HeadUint32PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint32Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A uint32 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint32",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint32Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)},
		},
		{
			name:     "PtrHeadUint32PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint32Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint32 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint32MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},
		{
			name:     "HeadUint32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint32MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint32 `json:"a"`
				B uint32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},
		{
			name:     "PtrHeadUint32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint32 `json:"a"`
				B *uint32 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint32ZeroNotRoot",
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
					A uint32 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint32NotRoot",
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
					A uint32 `json:"a"`
				}
			}{A: struct {
				A uint32 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint32PtrNotRoot",
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
					A *uint32 `json:"a"`
				}
			}{A: struct {
				A *uint32 `json:"a"`
			}{uint32ptr(1)}},
		},
		{
			name:     "HeadUint32PtrNilNotRoot",
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
					A *uint32 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint32ZeroNotRoot",
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
					A uint32 `json:"a"`
				}
			}{A: new(struct {
				A uint32 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint32NotRoot",
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
					A uint32 `json:"a"`
				}
			}{A: &(struct {
				A uint32 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint32PtrNotRoot",
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
					A *uint32 `json:"a"`
				}
			}{A: &(struct {
				A *uint32 `json:"a"`
			}{A: uint32ptr(1)})},
		},
		{
			name:     "PtrHeadUint32PtrNilNotRoot",
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
					A *uint32 `json:"a"`
				}
			}{A: &(struct {
				A *uint32 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint32NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint32 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint32ZeroMultiFieldsNotRoot",
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint32
			}{
				structUint32: structUint32{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint32Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint32
			}{
				structUint32: &structUint32{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint32
			}{
				structUint32: nil,
			},
		},
		{
			name:     "AnonymousHeadUint32PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint32Ptr
			}{
				structUint32Ptr: structUint32Ptr{A: uint32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint32PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint32Ptr
			}{
				structUint32Ptr: structUint32Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint32PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint32Ptr
			}{
				structUint32Ptr: &structUint32Ptr{A: uint32ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint32Ptr
			}{
				structUint32Ptr: nil,
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
