package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverFloat64(t *testing.T) {
	type structFloat64 struct {
		A float64 `json:"a"`
	}
	type structFloat64Ptr struct {
		A *float64 `json:"a"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		{
			name:     "HeadFloat64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A float64 `json:"a"`
			}{},
		},
		{
			name:     "HeadFloat64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A float64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadFloat64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)},
		},
		{
			name:     "HeadFloat64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *float64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A float64 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadFloat64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A float64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadFloat64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)},
		},
		{
			name:     "PtrHeadFloat64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *float64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat64Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadFloat64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{},
		},
		{
			name:     "HeadFloat64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadFloat64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},
		{
			name:     "HeadFloat64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadFloat64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadFloat64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},
		{
			name:     "PtrHeadFloat64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadFloat64ZeroNotRoot",
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
					A float64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadFloat64NotRoot",
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
					A float64 `json:"a"`
				}
			}{A: struct {
				A float64 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadFloat64PtrNotRoot",
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
					A *float64 `json:"a"`
				}
			}{A: struct {
				A *float64 `json:"a"`
			}{float64ptr(1)}},
		},
		{
			name:     "HeadFloat64PtrNilNotRoot",
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
					A *float64 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadFloat64ZeroNotRoot",
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
					A float64 `json:"a"`
				}
			}{A: new(struct {
				A float64 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadFloat64NotRoot",
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
					A float64 `json:"a"`
				}
			}{A: &(struct {
				A float64 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadFloat64PtrNotRoot",
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
					A *float64 `json:"a"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)})},
		},
		{
			name:     "PtrHeadFloat64PtrNilNotRoot",
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
					A *float64 `json:"a"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadFloat64NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *float64 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadFloat64ZeroMultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadFloat64MultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{A: struct {
				A float64 `json:"a"`
			}{A: 1}, B: struct {
				B float64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadFloat64PtrMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
				}
				B struct {
					B *float64 `json:"b"`
				}
			}{A: struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)}, B: struct {
				B *float64 `json:"b"`
			}{B: float64ptr(2)}},
		},
		{
			name:     "HeadFloat64PtrNilMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
				}
				B struct {
					B *float64 `json:"b"`
				}
			}{A: struct {
				A *float64 `json:"a"`
			}{A: nil}, B: struct {
				B *float64 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "PtrHeadFloat64ZeroMultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadFloat64MultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{A: struct {
				A float64 `json:"a"`
			}{A: 1}, B: struct {
				B float64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadFloat64PtrMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
				}
				B *struct {
					B *float64 `json:"b"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)}), B: &(struct {
				B *float64 `json:"b"`
			}{B: float64ptr(2)})},
		},
		{
			name:     "PtrHeadFloat64PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a"`
				}
				B *struct {
					B *float64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64NilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a"`
				}
				B *struct {
					B *float64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64DoubleMultiFieldsNotRoot",
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
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
				B *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
			}{A: &(struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
				B *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
				B *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64PtrDoubleMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
				B *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(1), B: float64ptr(2)}), B: &(struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(3), B: float64ptr(4)})},
		},
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
				B *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
				B *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "AnonymousHeadFloat64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat64
				B float64 `json:"b"`
			}{
				structFloat64: structFloat64{A: 1},
				B:             2,
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat64
				B float64 `json:"b"`
			}{
				structFloat64: &structFloat64{A: 1},
				B:             2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat64
				B float64 `json:"b"`
			}{
				structFloat64: nil,
				B:             2,
			},
		},
		{
			name:     "AnonymousHeadFloat64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: structFloat64Ptr{A: float64ptr(1)},
				B:                float64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: structFloat64Ptr{A: nil},
				B:                float64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: &structFloat64Ptr{A: float64ptr(1)},
				B:                float64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64Ptr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: nil,
				B:                float64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat64
			}{
				structFloat64: structFloat64{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat64
			}{
				structFloat64: &structFloat64{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64
			}{
				structFloat64: nil,
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat64Ptr
			}{
				structFloat64Ptr: structFloat64Ptr{A: float64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structFloat64Ptr
			}{
				structFloat64Ptr: structFloat64Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat64Ptr
			}{
				structFloat64Ptr: &structFloat64Ptr{A: float64ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64Ptr
			}{
				structFloat64Ptr: nil,
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
