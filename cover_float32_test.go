package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverFloat32(t *testing.T) {
	type structFloat32 struct {
		A float32 `json:"a"`
	}
	type structFloat32Ptr struct {
		A *float32 `json:"a"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		{
			name:     "HeadFloat32Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A float32 `json:"a"`
			}{},
		},
		{
			name:     "HeadFloat32",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A float32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadFloat32Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *float32 `json:"a"`
			}{A: float32ptr(1)},
		},
		{
			name:     "HeadFloat32PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *float32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat32Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A float32 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadFloat32",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A float32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadFloat32Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *float32 `json:"a"`
			}{A: float32ptr(1)},
		},
		{
			name:     "PtrHeadFloat32PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *float32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat32Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float32 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadFloat32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A float32 `json:"a"`
				B float32 `json:"b"`
			}{},
		},
		{
			name:     "HeadFloat32MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A float32 `json:"a"`
				B float32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadFloat32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *float32 `json:"a"`
				B *float32 `json:"b"`
			}{A: float32ptr(1), B: float32ptr(2)},
		},
		{
			name:     "HeadFloat32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *float32 `json:"a"`
				B *float32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A float32 `json:"a"`
				B float32 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadFloat32MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A float32 `json:"a"`
				B float32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadFloat32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *float32 `json:"a"`
				B *float32 `json:"b"`
			}{A: float32ptr(1), B: float32ptr(2)},
		},
		{
			name:     "PtrHeadFloat32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *float32 `json:"a"`
				B *float32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float32 `json:"a"`
				B *float32 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadFloat32ZeroNotRoot",
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
					A float32 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadFloat32NotRoot",
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
					A float32 `json:"a"`
				}
			}{A: struct {
				A float32 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadFloat32PtrNotRoot",
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
					A *float32 `json:"a"`
				}
			}{A: struct {
				A *float32 `json:"a"`
			}{float32ptr(1)}},
		},
		{
			name:     "HeadFloat32PtrNilNotRoot",
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
					A *float32 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadFloat32ZeroNotRoot",
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
					A float32 `json:"a"`
				}
			}{A: new(struct {
				A float32 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadFloat32NotRoot",
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
					A float32 `json:"a"`
				}
			}{A: &(struct {
				A float32 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadFloat32PtrNotRoot",
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
					A *float32 `json:"a"`
				}
			}{A: &(struct {
				A *float32 `json:"a"`
			}{A: float32ptr(1)})},
		},
		{
			name:     "PtrHeadFloat32PtrNilNotRoot",
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
					A *float32 `json:"a"`
				}
			}{A: &(struct {
				A *float32 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadFloat32NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *float32 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadFloat32ZeroMultiFieldsNotRoot",
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
					A float32 `json:"a"`
				}
				B struct {
					B float32 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadFloat32MultiFieldsNotRoot",
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
					A float32 `json:"a"`
				}
				B struct {
					B float32 `json:"b"`
				}
			}{A: struct {
				A float32 `json:"a"`
			}{A: 1}, B: struct {
				B float32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadFloat32PtrMultiFieldsNotRoot",
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
					A *float32 `json:"a"`
				}
				B struct {
					B *float32 `json:"b"`
				}
			}{A: struct {
				A *float32 `json:"a"`
			}{A: float32ptr(1)}, B: struct {
				B *float32 `json:"b"`
			}{B: float32ptr(2)}},
		},
		{
			name:     "HeadFloat32PtrNilMultiFieldsNotRoot",
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
					A *float32 `json:"a"`
				}
				B struct {
					B *float32 `json:"b"`
				}
			}{A: struct {
				A *float32 `json:"a"`
			}{A: nil}, B: struct {
				B *float32 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadFloat32ZeroMultiFieldsNotRoot",
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
					A float32 `json:"a"`
				}
				B struct {
					B float32 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadFloat32MultiFieldsNotRoot",
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
					A float32 `json:"a"`
				}
				B struct {
					B float32 `json:"b"`
				}
			}{A: struct {
				A float32 `json:"a"`
			}{A: 1}, B: struct {
				B float32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadFloat32PtrMultiFieldsNotRoot",
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
					A *float32 `json:"a"`
				}
				B *struct {
					B *float32 `json:"b"`
				}
			}{A: &(struct {
				A *float32 `json:"a"`
			}{A: float32ptr(1)}), B: &(struct {
				B *float32 `json:"b"`
			}{B: float32ptr(2)})},
		},
		{
			name:     "PtrHeadFloat32PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float32 `json:"a"`
				}
				B *struct {
					B *float32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32NilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float32 `json:"a"`
				}
				B *struct {
					B *float32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat32DoubleMultiFieldsNotRoot",
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
					A float32 `json:"a"`
					B float32 `json:"b"`
				}
				B *struct {
					A float32 `json:"a"`
					B float32 `json:"b"`
				}
			}{A: &(struct {
				A float32 `json:"a"`
				B float32 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A float32 `json:"a"`
				B float32 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadFloat32NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A float32 `json:"a"`
					B float32 `json:"b"`
				}
				B *struct {
					A float32 `json:"a"`
					B float32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A float32 `json:"a"`
					B float32 `json:"b"`
				}
				B *struct {
					A float32 `json:"a"`
					B float32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat32PtrDoubleMultiFieldsNotRoot",
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
					A *float32 `json:"a"`
					B *float32 `json:"b"`
				}
				B *struct {
					A *float32 `json:"a"`
					B *float32 `json:"b"`
				}
			}{A: &(struct {
				A *float32 `json:"a"`
				B *float32 `json:"b"`
			}{A: float32ptr(1), B: float32ptr(2)}), B: &(struct {
				A *float32 `json:"a"`
				B *float32 `json:"b"`
			}{A: float32ptr(3), B: float32ptr(4)})},
		},
		{
			name:     "PtrHeadFloat32PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float32 `json:"a"`
					B *float32 `json:"b"`
				}
				B *struct {
					A *float32 `json:"a"`
					B *float32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float32 `json:"a"`
					B *float32 `json:"b"`
				}
				B *struct {
					A *float32 `json:"a"`
					B *float32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadFloat32",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat32
				B float32 `json:"b"`
			}{
				structFloat32: structFloat32{A: 1},
				B:             2,
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat32
				B float32 `json:"b"`
			}{
				structFloat32: &structFloat32{A: 1},
				B:             2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat32
				B float32 `json:"b"`
			}{
				structFloat32: nil,
				B:             2,
			},
		},
		{
			name:     "AnonymousHeadFloat32Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat32Ptr
				B *float32 `json:"b"`
			}{
				structFloat32Ptr: structFloat32Ptr{A: float32ptr(1)},
				B:                float32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat32PtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structFloat32Ptr
				B *float32 `json:"b"`
			}{
				structFloat32Ptr: structFloat32Ptr{A: nil},
				B:                float32ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat32Ptr
				B *float32 `json:"b"`
			}{
				structFloat32Ptr: &structFloat32Ptr{A: float32ptr(1)},
				B:                float32ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32Ptr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat32Ptr
				B *float32 `json:"b"`
			}{
				structFloat32Ptr: nil,
				B:                float32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat32Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat32
			}{
				structFloat32: structFloat32{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat32
			}{
				structFloat32: &structFloat32{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat32
			}{
				structFloat32: nil,
			},
		},
		{
			name:     "AnonymousHeadFloat32PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat32Ptr
			}{
				structFloat32Ptr: structFloat32Ptr{A: float32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadFloat32PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structFloat32Ptr
			}{
				structFloat32Ptr: structFloat32Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat32Ptr
			}{
				structFloat32Ptr: &structFloat32Ptr{A: float32ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat32Ptr
			}{
				structFloat32Ptr: nil,
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
