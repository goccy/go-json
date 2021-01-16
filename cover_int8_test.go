package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverInt8(t *testing.T) {
	type structInt8 struct {
		A int8 `json:"a"`
	}
	type structInt8OmitEmpty struct {
		A int8 `json:"a,omitempty"`
	}
	type structInt8String struct {
		A int8 `json:"a,string"`
	}

	type structInt8Ptr struct {
		A *int8 `json:"a"`
	}
	type structInt8PtrOmitEmpty struct {
		A *int8 `json:"a,omitempty"`
	}
	type structInt8PtrString struct {
		A *int8 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadInt8Zero
		{
			name:     "HeadInt8Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A int8 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt8ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int8 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadInt8ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A int8 `json:"a,string"`
			}{},
		},

		// HeadInt8
		{
			name:     "HeadInt8",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt8OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int8 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadInt8String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A int8 `json:"a,string"`
			}{A: 1},
		},

		// HeadInt8Ptr
		{
			name:     "HeadInt8Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		{
			name:     "HeadInt8PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int8 `json:"a,omitempty"`
			}{A: int8ptr(1)},
		},
		{
			name:     "HeadInt8PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *int8 `json:"a,string"`
			}{A: int8ptr(1)},
		},

		// HeadInt8PtrNil
		{
			name:     "HeadInt8PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadInt8PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int8 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadInt8PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int8 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt8Zero
		{
			name:     "PtrHeadInt8Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A int8 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt8ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int8 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt8ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A int8 `json:"a,string"`
			}{},
		},

		// PtrHeadInt8
		{
			name:     "PtrHeadInt8",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt8OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int8 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt8String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A int8 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadInt8Ptr
		{
			name:     "PtrHeadInt8Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		{
			name:     "PtrHeadInt8PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int8 `json:"a,omitempty"`
			}{A: int8ptr(1)},
		},
		{
			name:     "PtrHeadInt8PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *int8 `json:"a,string"`
			}{A: int8ptr(1)},
		},

		// PtrHeadInt8PtrNil
		{
			name:     "PtrHeadInt8PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int8 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt8PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int8 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt8PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int8 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt8Nil
		{
			name:     "PtrHeadInt8Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int8 `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadInt8NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int8 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt8NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int8 `json:"a,string"`
			})(nil),
		},

		// HeadInt8ZeroMultiFields
		{
			name:     "HeadInt8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt8ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int8 `json:"a,omitempty"`
				B int8 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadInt8ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A int8 `json:"a,string"`
				B int8 `json:"b,string"`
			}{},
		},

		// HeadInt8MultiFields
		{
			name:     "HeadInt8MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt8MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int8 `json:"a,omitempty"`
				B int8 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt8MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A int8 `json:"a,string"`
				B int8 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadInt8PtrMultiFields
		{
			name:     "HeadInt8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
		{
			name:     "HeadInt8PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int8 `json:"a,omitempty"`
				B *int8 `json:"b,omitempty"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
		{
			name:     "HeadInt8PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *int8 `json:"a,string"`
				B *int8 `json:"b,string"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},

		// HeadInt8PtrNilMultiFields
		{
			name:     "HeadInt8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadInt8PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int8 `json:"a,omitempty"`
				B *int8 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadInt8PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int8 `json:"a,string"`
				B *int8 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt8ZeroMultiFields
		{
			name:     "PtrHeadInt8ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt8ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int8 `json:"a,omitempty"`
				B int8 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt8ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A int8 `json:"a,string"`
				B int8 `json:"b,string"`
			}{},
		},

		// PtrHeadInt8MultiFields
		{
			name:     "PtrHeadInt8MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt8MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int8 `json:"a,omitempty"`
				B int8 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt8MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A int8 `json:"a,string"`
				B int8 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadInt8PtrMultiFields
		{
			name:     "PtrHeadInt8PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
		{
			name:     "PtrHeadInt8PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int8 `json:"a,omitempty"`
				B *int8 `json:"b,omitempty"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
		{
			name:     "PtrHeadInt8PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *int8 `json:"a,string"`
				B *int8 `json:"b,string"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},

		// PtrHeadInt8PtrNilMultiFields
		{
			name:     "PtrHeadInt8PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int8 `json:"a,omitempty"`
				B *int8 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int8 `json:"a,string"`
				B *int8 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt8NilMultiFields
		{
			name:     "PtrHeadInt8NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadInt8NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int8 `json:"a,omitempty"`
				B *int8 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt8NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int8 `json:"a,string"`
				B *int8 `json:"b,string"`
			})(nil),
		},

		// HeadInt8ZeroNotRoot
		{
			name:     "HeadInt8ZeroNotRoot",
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
					A int8 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt8ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A int8 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt8ZeroNotRootString",
			expected: `{"A":{"a":"0"}}`,
			indentExpected: `
{
  "A": {
    "a": "0"
  }
}
`,
			data: struct {
				A struct {
					A int8 `json:"a,string"`
				}
			}{},
		},

		// HeadInt8NotRoot
		{
			name:     "HeadInt8NotRoot",
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
					A int8 `json:"a"`
				}
			}{A: struct {
				A int8 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt8NotRootOmitEmpty",
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
					A int8 `json:"a,omitempty"`
				}
			}{A: struct {
				A int8 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadInt8NotRootString",
			expected: `{"A":{"a":"1"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  }
}
`,
			data: struct {
				A struct {
					A int8 `json:"a,string"`
				}
			}{A: struct {
				A int8 `json:"a,string"`
			}{A: 1}},
		},

		// HeadInt8PtrNotRoot
		{
			name:     "HeadInt8PtrNotRoot",
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
					A *int8 `json:"a"`
				}
			}{A: struct {
				A *int8 `json:"a"`
			}{int8ptr(1)}},
		},
		{
			name:     "HeadInt8PtrNotRootOmitEmpty",
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
					A *int8 `json:"a,omitempty"`
				}
			}{A: struct {
				A *int8 `json:"a,omitempty"`
			}{int8ptr(1)}},
		},
		{
			name:     "HeadInt8PtrNotRootString",
			expected: `{"A":{"a":"1"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  }
}
`,
			data: struct {
				A struct {
					A *int8 `json:"a,string"`
				}
			}{A: struct {
				A *int8 `json:"a,string"`
			}{int8ptr(1)}},
		},

		// HeadInt8PtrNilNotRoot
		{
			name:     "HeadInt8PtrNilNotRoot",
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
					A *int8 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt8PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *int8 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt8PtrNilNotRootString",
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
					A *int8 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadInt8ZeroNotRoot
		{
			name:     "PtrHeadInt8ZeroNotRoot",
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
					A int8 `json:"a"`
				}
			}{A: new(struct {
				A int8 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt8ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A int8 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A int8 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadInt8ZeroNotRootString",
			expected: `{"A":{"a":"0"}}`,
			indentExpected: `
{
  "A": {
    "a": "0"
  }
}
`,
			data: struct {
				A *struct {
					A int8 `json:"a,string"`
				}
			}{A: new(struct {
				A int8 `json:"a,string"`
			})},
		},

		// PtrHeadInt8NotRoot
		{
			name:     "PtrHeadInt8NotRoot",
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
					A int8 `json:"a"`
				}
			}{A: &(struct {
				A int8 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt8NotRootOmitEmpty",
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
					A int8 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A int8 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt8NotRootString",
			expected: `{"A":{"a":"1"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  }
}
`,
			data: struct {
				A *struct {
					A int8 `json:"a,string"`
				}
			}{A: &(struct {
				A int8 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadInt8PtrNotRoot
		{
			name:     "PtrHeadInt8PtrNotRoot",
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
					A *int8 `json:"a"`
				}
			}{A: &(struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)})},
		},
		{
			name:     "PtrHeadInt8PtrNotRootOmitEmpty",
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
					A *int8 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int8 `json:"a,omitempty"`
			}{A: int8ptr(1)})},
		},
		{
			name:     "PtrHeadInt8PtrNotRootString",
			expected: `{"A":{"a":"1"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  }
}
`,
			data: struct {
				A *struct {
					A *int8 `json:"a,string"`
				}
			}{A: &(struct {
				A *int8 `json:"a,string"`
			}{A: int8ptr(1)})},
		},

		// PtrHeadInt8PtrNilNotRoot
		{
			name:     "PtrHeadInt8PtrNilNotRoot",
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
					A *int8 `json:"a"`
				}
			}{A: &(struct {
				A *int8 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt8PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *int8 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int8 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt8PtrNilNotRootString",
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
					A *int8 `json:"a,string"`
				}
			}{A: &(struct {
				A *int8 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadInt8NilNotRoot
		{
			name:     "PtrHeadInt8NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int8 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadInt8NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *int8 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt8NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int8 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadInt8ZeroMultiFieldsNotRoot
		{
			name:     "HeadInt8ZeroMultiFieldsNotRoot",
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
					A int8 `json:"a"`
				}
				B struct {
					B int8 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadInt8ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A int8 `json:"a,omitempty"`
				}
				B struct {
					B int8 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt8ZeroMultiFieldsNotRootString",
			expected: `{"A":{"a":"0"},"B":{"b":"0"}}`,
			indentExpected: `
{
  "A": {
    "a": "0"
  },
  "B": {
    "b": "0"
  }
}
`,
			data: struct {
				A struct {
					A int8 `json:"a,string"`
				}
				B struct {
					B int8 `json:"b,string"`
				}
			}{},
		},

		// HeadInt8MultiFieldsNotRoot
		{
			name:     "HeadInt8MultiFieldsNotRoot",
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
			name:     "HeadInt8MultiFieldsNotRootOmitEmpty",
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
					A int8 `json:"a,omitempty"`
				}
				B struct {
					B int8 `json:"b,omitempty"`
				}
			}{A: struct {
				A int8 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int8 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadInt8MultiFieldsNotRootString",
			expected: `{"A":{"a":"1"},"B":{"b":"2"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  },
  "B": {
    "b": "2"
  }
}
`,
			data: struct {
				A struct {
					A int8 `json:"a,string"`
				}
				B struct {
					B int8 `json:"b,string"`
				}
			}{A: struct {
				A int8 `json:"a,string"`
			}{A: 1}, B: struct {
				B int8 `json:"b,string"`
			}{B: 2}},
		},

		// HeadInt8PtrMultiFieldsNotRoot
		{
			name:     "HeadInt8PtrMultiFieldsNotRoot",
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
			name:     "HeadInt8PtrMultiFieldsNotRootOmitEmpty",
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
					A *int8 `json:"a,omitempty"`
				}
				B struct {
					B *int8 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int8 `json:"a,omitempty"`
			}{A: int8ptr(1)}, B: struct {
				B *int8 `json:"b,omitempty"`
			}{B: int8ptr(2)}},
		},
		{
			name:     "HeadInt8PtrMultiFieldsNotRootString",
			expected: `{"A":{"a":"1"},"B":{"b":"2"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  },
  "B": {
    "b": "2"
  }
}
`,
			data: struct {
				A struct {
					A *int8 `json:"a,string"`
				}
				B struct {
					B *int8 `json:"b,string"`
				}
			}{A: struct {
				A *int8 `json:"a,string"`
			}{A: int8ptr(1)}, B: struct {
				B *int8 `json:"b,string"`
			}{B: int8ptr(2)}},
		},

		// HeadInt8PtrNilMultiFieldsNotRoot
		{
			name:     "HeadInt8PtrNilMultiFieldsNotRoot",
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
			name:     "HeadInt8PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *int8 `json:"a,omitempty"`
				}
				B struct {
					B *int8 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int8 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *int8 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadInt8PtrNilMultiFieldsNotRootString",
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
					A *int8 `json:"a,string"`
				}
				B struct {
					B *int8 `json:"b,string"`
				}
			}{A: struct {
				A *int8 `json:"a,string"`
			}{A: nil}, B: struct {
				B *int8 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadInt8ZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8ZeroMultiFieldsNotRoot",
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
					A int8 `json:"a"`
				}
				B struct {
					B int8 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt8ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A int8 `json:"a,omitempty"`
				}
				B struct {
					B int8 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt8ZeroMultiFieldsNotRootString",
			expected: `{"A":{"a":"0"},"B":{"b":"0"}}`,
			indentExpected: `
{
  "A": {
    "a": "0"
  },
  "B": {
    "b": "0"
  }
}
`,
			data: &struct {
				A struct {
					A int8 `json:"a,string"`
				}
				B struct {
					B int8 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadInt8MultiFieldsNotRoot
		{
			name:     "PtrHeadInt8MultiFieldsNotRoot",
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
			name:     "PtrHeadInt8MultiFieldsNotRootOmitEmpty",
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
					A int8 `json:"a,omitempty"`
				}
				B struct {
					B int8 `json:"b,omitempty"`
				}
			}{A: struct {
				A int8 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int8 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt8MultiFieldsNotRootString",
			expected: `{"A":{"a":"1"},"B":{"b":"2"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  },
  "B": {
    "b": "2"
  }
}
`,
			data: &struct {
				A struct {
					A int8 `json:"a,string"`
				}
				B struct {
					B int8 `json:"b,string"`
				}
			}{A: struct {
				A int8 `json:"a,string"`
			}{A: 1}, B: struct {
				B int8 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadInt8PtrMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8PtrMultiFieldsNotRoot",
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
			name:     "PtrHeadInt8PtrMultiFieldsNotRootOmitEmpty",
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
					A *int8 `json:"a,omitempty"`
				}
				B *struct {
					B *int8 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int8 `json:"a,omitempty"`
			}{A: int8ptr(1)}), B: &(struct {
				B *int8 `json:"b,omitempty"`
			}{B: int8ptr(2)})},
		},
		{
			name:     "PtrHeadInt8PtrMultiFieldsNotRootString",
			expected: `{"A":{"a":"1"},"B":{"b":"2"}}`,
			indentExpected: `
{
  "A": {
    "a": "1"
  },
  "B": {
    "b": "2"
  }
}
`,
			data: &struct {
				A *struct {
					A *int8 `json:"a,string"`
				}
				B *struct {
					B *int8 `json:"b,string"`
				}
			}{A: &(struct {
				A *int8 `json:"a,string"`
			}{A: int8ptr(1)}), B: &(struct {
				B *int8 `json:"b,string"`
			}{B: int8ptr(2)})},
		},

		// PtrHeadInt8PtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			name:     "PtrHeadInt8PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int8 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *int8 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int8 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *int8 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt8NilMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8NilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
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
			name:     "PtrHeadInt8NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int8 `json:"a,omitempty"`
				}
				B *struct {
					B *int8 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt8NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int8 `json:"a,string"`
				}
				B *struct {
					B *int8 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt8DoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8DoubleMultiFieldsNotRoot",
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
			name:     "PtrHeadInt8DoubleMultiFieldsNotRootOmitEmpty",
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
					A int8 `json:"a,omitempty"`
					B int8 `json:"b,omitempty"`
				}
				B *struct {
					A int8 `json:"a,omitempty"`
					B int8 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A int8 `json:"a,omitempty"`
				B int8 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A int8 `json:"a,omitempty"`
				B int8 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt8DoubleMultiFieldsNotRootString",
			expected: `{"A":{"a":"1","b":"2"},"B":{"a":"3","b":"4"}}`,
			indentExpected: `
{
  "A": {
    "a": "1",
    "b": "2"
  },
  "B": {
    "a": "3",
    "b": "4"
  }
}
`,
			data: &struct {
				A *struct {
					A int8 `json:"a,string"`
					B int8 `json:"b,string"`
				}
				B *struct {
					A int8 `json:"a,string"`
					B int8 `json:"b,string"`
				}
			}{A: &(struct {
				A int8 `json:"a,string"`
				B int8 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A int8 `json:"a,string"`
				B int8 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadInt8NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A int8 `json:"a,omitempty"`
					B int8 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A int8 `json:"a,omitempty"`
					B int8 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int8 `json:"a,string"`
					B int8 `json:"b,string"`
				}
				B *struct {
					A int8 `json:"a,string"`
					B int8 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt8NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
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
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int8 `json:"a,omitempty"`
					B int8 `json:"b,omitempty"`
				}
				B *struct {
					A int8 `json:"a,omitempty"`
					B int8 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt8NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int8 `json:"a,string"`
					B int8 `json:"b,string"`
				}
				B *struct {
					A int8 `json:"a,string"`
					B int8 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt8PtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8PtrDoubleMultiFieldsNotRoot",
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
			name:     "PtrHeadInt8PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *int8 `json:"a,omitempty"`
					B *int8 `json:"b,omitempty"`
				}
				B *struct {
					A *int8 `json:"a,omitempty"`
					B *int8 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int8 `json:"a,omitempty"`
				B *int8 `json:"b,omitempty"`
			}{A: int8ptr(1), B: int8ptr(2)}), B: &(struct {
				A *int8 `json:"a,omitempty"`
				B *int8 `json:"b,omitempty"`
			}{A: int8ptr(3), B: int8ptr(4)})},
		},
		{
			name:     "PtrHeadInt8PtrDoubleMultiFieldsNotRootString",
			expected: `{"A":{"a":"1","b":"2"},"B":{"a":"3","b":"4"}}`,
			indentExpected: `
{
  "A": {
    "a": "1",
    "b": "2"
  },
  "B": {
    "a": "3",
    "b": "4"
  }
}
`,
			data: &struct {
				A *struct {
					A *int8 `json:"a,string"`
					B *int8 `json:"b,string"`
				}
				B *struct {
					A *int8 `json:"a,string"`
					B *int8 `json:"b,string"`
				}
			}{A: &(struct {
				A *int8 `json:"a,string"`
				B *int8 `json:"b,string"`
			}{A: int8ptr(1), B: int8ptr(2)}), B: &(struct {
				A *int8 `json:"a,string"`
				B *int8 `json:"b,string"`
			}{A: int8ptr(3), B: int8ptr(4)})},
		},

		// PtrHeadInt8PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
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
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int8 `json:"a,omitempty"`
					B *int8 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *int8 `json:"a,omitempty"`
					B *int8 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int8 `json:"a,string"`
					B *int8 `json:"b,string"`
				}
				B *struct {
					A *int8 `json:"a,string"`
					B *int8 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt8PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
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
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int8 `json:"a,omitempty"`
					B *int8 `json:"b,omitempty"`
				}
				B *struct {
					A *int8 `json:"a,omitempty"`
					B *int8 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt8PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int8 `json:"a,string"`
					B *int8 `json:"b,string"`
				}
				B *struct {
					A *int8 `json:"a,string"`
					B *int8 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadInt8
		{
			name:     "AnonymousHeadInt8",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt8
				B int8 `json:"b"`
			}{
				structInt8: structInt8{A: 1},
				B:          2,
			},
		},
		{
			name:     "AnonymousHeadInt8OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt8OmitEmpty
				B int8 `json:"b,omitempty"`
			}{
				structInt8OmitEmpty: structInt8OmitEmpty{A: 1},
				B:                   2,
			},
		},
		{
			name:     "AnonymousHeadInt8String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt8String
				B int8 `json:"b,string"`
			}{
				structInt8String: structInt8String{A: 1},
				B:                2,
			},
		},

		// PtrAnonymousHeadInt8
		{
			name:     "PtrAnonymousHeadInt8",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt8
				B int8 `json:"b"`
			}{
				structInt8: &structInt8{A: 1},
				B:          2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt8OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt8OmitEmpty
				B int8 `json:"b,omitempty"`
			}{
				structInt8OmitEmpty: &structInt8OmitEmpty{A: 1},
				B:                   2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt8String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt8String
				B int8 `json:"b,string"`
			}{
				structInt8String: &structInt8String{A: 1},
				B:                2,
			},
		},

		// NilPtrAnonymousHeadInt8
		{
			name:     "NilPtrAnonymousHeadInt8",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt8
				B int8 `json:"b"`
			}{
				structInt8: nil,
				B:          2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt8OmitEmpty
				B int8 `json:"b,omitempty"`
			}{
				structInt8OmitEmpty: nil,
				B:                   2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt8String
				B int8 `json:"b,string"`
			}{
				structInt8String: nil,
				B:                2,
			},
		},

		// AnonymousHeadInt8Ptr
		{
			name:     "AnonymousHeadInt8Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: structInt8Ptr{A: int8ptr(1)},
				B:             int8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt8PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt8PtrOmitEmpty
				B *int8 `json:"b,omitempty"`
			}{
				structInt8PtrOmitEmpty: structInt8PtrOmitEmpty{A: int8ptr(1)},
				B:                      int8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt8PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt8PtrString
				B *int8 `json:"b,string"`
			}{
				structInt8PtrString: structInt8PtrString{A: int8ptr(1)},
				B:                   int8ptr(2),
			},
		},

		// AnonymousHeadInt8PtrNil
		{
			name:     "AnonymousHeadInt8PtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: structInt8Ptr{A: nil},
				B:             int8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt8PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structInt8PtrOmitEmpty
				B *int8 `json:"b,omitempty"`
			}{
				structInt8PtrOmitEmpty: structInt8PtrOmitEmpty{A: nil},
				B:                      int8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt8PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structInt8PtrString
				B *int8 `json:"b,string"`
			}{
				structInt8PtrString: structInt8PtrString{A: nil},
				B:                   int8ptr(2),
			},
		},

		// PtrAnonymousHeadInt8Ptr
		{
			name:     "PtrAnonymousHeadInt8Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: &structInt8Ptr{A: int8ptr(1)},
				B:             int8ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt8PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt8PtrOmitEmpty
				B *int8 `json:"b,omitempty"`
			}{
				structInt8PtrOmitEmpty: &structInt8PtrOmitEmpty{A: int8ptr(1)},
				B:                      int8ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt8PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt8PtrString
				B *int8 `json:"b,string"`
			}{
				structInt8PtrString: &structInt8PtrString{A: int8ptr(1)},
				B:                   int8ptr(2),
			},
		},

		// NilPtrAnonymousHeadInt8Ptr
		{
			name:     "NilPtrAnonymousHeadInt8Ptr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt8Ptr
				B *int8 `json:"b"`
			}{
				structInt8Ptr: nil,
				B:             int8ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt8PtrOmitEmpty
				B *int8 `json:"b,omitempty"`
			}{
				structInt8PtrOmitEmpty: nil,
				B:                      int8ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt8PtrString
				B *int8 `json:"b,string"`
			}{
				structInt8PtrString: nil,
				B:                   int8ptr(2),
			},
		},

		// AnonymousHeadInt8Only
		{
			name:     "AnonymousHeadInt8Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt8
			}{
				structInt8: structInt8{A: 1},
			},
		},
		{
			name:     "AnonymousHeadInt8OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt8OmitEmpty
			}{
				structInt8OmitEmpty: structInt8OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadInt8OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt8String
			}{
				structInt8String: structInt8String{A: 1},
			},
		},

		// PtrAnonymousHeadInt8Only
		{
			name:     "PtrAnonymousHeadInt8Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt8
			}{
				structInt8: &structInt8{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt8OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt8OmitEmpty
			}{
				structInt8OmitEmpty: &structInt8OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt8OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt8String
			}{
				structInt8String: &structInt8String{A: 1},
			},
		},

		// NilPtrAnonymousHeadInt8Only
		{
			name:     "NilPtrAnonymousHeadInt8Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt8
			}{
				structInt8: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt8OmitEmpty
			}{
				structInt8OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt8String
			}{
				structInt8String: nil,
			},
		},

		// AnonymousHeadInt8PtrOnly
		{
			name:     "AnonymousHeadInt8PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt8Ptr
			}{
				structInt8Ptr: structInt8Ptr{A: int8ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt8PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt8PtrOmitEmpty
			}{
				structInt8PtrOmitEmpty: structInt8PtrOmitEmpty{A: int8ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt8PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt8PtrString
			}{
				structInt8PtrString: structInt8PtrString{A: int8ptr(1)},
			},
		},

		// AnonymousHeadInt8PtrNilOnly
		{
			name:     "AnonymousHeadInt8PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt8Ptr
			}{
				structInt8Ptr: structInt8Ptr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadInt8PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structInt8PtrOmitEmpty
			}{
				structInt8PtrOmitEmpty: structInt8PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadInt8PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt8PtrString
			}{
				structInt8PtrString: structInt8PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadInt8PtrOnly
		{
			name:     "PtrAnonymousHeadInt8PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt8Ptr
			}{
				structInt8Ptr: &structInt8Ptr{A: int8ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadInt8PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt8PtrOmitEmpty
			}{
				structInt8PtrOmitEmpty: &structInt8PtrOmitEmpty{A: int8ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadInt8PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt8PtrString
			}{
				structInt8PtrString: &structInt8PtrString{A: int8ptr(1)},
			},
		},

		// NilPtrAnonymousHeadInt8PtrOnly
		{
			name:     "NilPtrAnonymousHeadInt8PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt8Ptr
			}{
				structInt8Ptr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt8PtrOmitEmpty
			}{
				structInt8PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt8PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt8PtrString
			}{
				structInt8PtrString: nil,
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
