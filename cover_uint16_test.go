package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverUint16(t *testing.T) {
	type structUint16 struct {
		A uint16 `json:"a"`
	}
	type structUint16Ptr struct {
		A *uint16 `json:"a"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		{
			name:     "HeadUint16Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A uint16 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint16",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint16Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)},
		},
		{
			name:     "HeadUint16PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint16Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A uint16 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint16",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint16 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint16Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)},
		},
		{
			name:     "PtrHeadUint16PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint16 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint16Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint16 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint16MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},
		{
			name:     "HeadUint16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint16MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint16 `json:"a"`
				B uint16 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint16PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},
		{
			name:     "PtrHeadUint16PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint16 `json:"a"`
				B *uint16 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint16ZeroNotRoot",
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
					A uint16 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint16NotRoot",
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
					A uint16 `json:"a"`
				}
			}{A: struct {
				A uint16 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint16PtrNotRoot",
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
					A *uint16 `json:"a"`
				}
			}{A: struct {
				A *uint16 `json:"a"`
			}{uint16ptr(1)}},
		},
		{
			name:     "HeadUint16PtrNilNotRoot",
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
					A *uint16 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint16ZeroNotRoot",
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
					A uint16 `json:"a"`
				}
			}{A: new(struct {
				A uint16 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint16NotRoot",
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
					A uint16 `json:"a"`
				}
			}{A: &(struct {
				A uint16 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint16PtrNotRoot",
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
					A *uint16 `json:"a"`
				}
			}{A: &(struct {
				A *uint16 `json:"a"`
			}{A: uint16ptr(1)})},
		},
		{
			name:     "PtrHeadUint16PtrNilNotRoot",
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
					A *uint16 `json:"a"`
				}
			}{A: &(struct {
				A *uint16 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint16NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint16 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint16ZeroMultiFieldsNotRoot",
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint16
			}{
				structUint16: structUint16{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint16Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint16
			}{
				structUint16: &structUint16{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint16
			}{
				structUint16: nil,
			},
		},
		{
			name:     "AnonymousHeadUint16PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint16Ptr
			}{
				structUint16Ptr: structUint16Ptr{A: uint16ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint16PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint16Ptr
			}{
				structUint16Ptr: structUint16Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint16PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint16Ptr
			}{
				structUint16Ptr: &structUint16Ptr{A: uint16ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint16Ptr
			}{
				structUint16Ptr: nil,
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
