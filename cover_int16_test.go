package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverInt16(t *testing.T) {
	type structInt16 struct {
		A int16 `json:"a"`
	}
	type structInt16Ptr struct {
		A *int16 `json:"a"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		{
			name:     "HeadInt16Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A int16 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt16",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt16Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)},
		},
		{
			name:     "HeadInt16PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt16Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A int16 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt16",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt16Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)},
		},
		{
			name:     "PtrHeadInt16PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt16Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int16 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadInt16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt16MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},
		{
			name:     "HeadInt16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt16MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int16 `json:"a"`
				B int16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},
		{
			name:     "PtrHeadInt16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int16 `json:"a"`
				B *int16 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadInt16ZeroNotRoot",
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
					A int16 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt16NotRoot",
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
					A int16 `json:"a"`
				}
			}{A: struct {
				A int16 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt16PtrNotRoot",
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
					A *int16 `json:"a"`
				}
			}{A: struct {
				A *int16 `json:"a"`
			}{int16ptr(1)}},
		},
		{
			name:     "HeadInt16PtrNilNotRoot",
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
					A *int16 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt16ZeroNotRoot",
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
					A int16 `json:"a"`
				}
			}{A: new(struct {
				A int16 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt16NotRoot",
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
					A int16 `json:"a"`
				}
			}{A: &(struct {
				A int16 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt16PtrNotRoot",
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
					A *int16 `json:"a"`
				}
			}{A: &(struct {
				A *int16 `json:"a"`
			}{A: int16ptr(1)})},
		},
		{
			name:     "PtrHeadInt16PtrNilNotRoot",
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
					A *int16 `json:"a"`
				}
			}{A: &(struct {
				A *int16 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt16NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int16 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadInt16ZeroMultiFieldsNotRoot",
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt16
			}{
				structInt16: structInt16{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt16Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt16
			}{
				structInt16: &structInt16{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt16
			}{
				structInt16: nil,
			},
		},
		{
			name:     "AnonymousHeadInt16PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt16Ptr
			}{
				structInt16Ptr: structInt16Ptr{A: int16ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt16PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt16Ptr
			}{
				structInt16Ptr: structInt16Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadInt16PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt16Ptr
			}{
				structInt16Ptr: &structInt16Ptr{A: int16ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt16Ptr
			}{
				structInt16Ptr: nil,
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
