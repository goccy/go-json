package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverUint64(t *testing.T) {
	type structUint64 struct {
		A uint64 `json:"a"`
	}
	type structUint64Ptr struct {
		A *uint64 `json:"a"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		{
			name:     "HeadUint64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A uint64 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "HeadUint64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A uint64 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "PtrHeadUint64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a"`
			})(nil),
		},
		{
			name:     "HeadUint64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "HeadUint64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "PtrHeadUint64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			})(nil),
		},
		{
			name:     "HeadUint64ZeroNotRoot",
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
					A uint64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint64NotRoot",
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
					A uint64 `json:"a"`
				}
			}{A: struct {
				A uint64 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint64PtrNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{A: struct {
				A *uint64 `json:"a"`
			}{uint64ptr(1)}},
		},
		{
			name:     "HeadUint64PtrNilNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroNotRoot",
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
					A uint64 `json:"a"`
				}
			}{A: new(struct {
				A uint64 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint64NotRoot",
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
					A uint64 `json:"a"`
				}
			}{A: &(struct {
				A uint64 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint64PtrNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)})},
		},
		{
			name:     "PtrHeadUint64PtrNilNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint64NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint64 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "HeadUint64ZeroMultiFieldsNotRoot",
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			indentExpected: `
null
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
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
			indentExpected: `
{
  "b": 2
}
`,
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
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint64
			}{
				structUint64: structUint64{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint64
			}{
				structUint64: &structUint64{A: 1},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64
			}{
				structUint64: nil,
			},
		},
		{
			name:     "AnonymousHeadUint64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint64Ptr
			}{
				structUint64Ptr: structUint64Ptr{A: uint64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint64PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint64Ptr
			}{
				structUint64Ptr: structUint64Ptr{A: nil},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint64Ptr
			}{
				structUint64Ptr: &structUint64Ptr{A: uint64ptr(1)},
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64Ptr
			}{
				structUint64Ptr: nil,
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
