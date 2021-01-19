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
	type structUint8OmitEmpty struct {
		A uint8 `json:"a,omitempty"`
	}
	type structUint8String struct {
		A uint8 `json:"a,string"`
	}

	type structUint8Ptr struct {
		A *uint8 `json:"a"`
	}
	type structUint8PtrOmitEmpty struct {
		A *uint8 `json:"a,omitempty"`
	}
	type structUint8PtrString struct {
		A *uint8 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadUint8Zero
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
			name:     "HeadUint8ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint8 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadUint8ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A uint8 `json:"a,string"`
			}{},
		},

		// HeadUint8
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
			name:     "HeadUint8OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint8 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadUint8String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A uint8 `json:"a,string"`
			}{A: 1},
		},

		// HeadUint8Ptr
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
			name:     "HeadUint8PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint8 `json:"a,omitempty"`
			}{A: uint8ptr(1)},
		},
		{
			name:     "HeadUint8PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *uint8 `json:"a,string"`
			}{A: uint8ptr(1)},
		},

		// HeadUint8PtrNil
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
			name:     "HeadUint8PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint8 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadUint8PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint8 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint8Zero
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
			name:     "PtrHeadUint8ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint8 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint8ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A uint8 `json:"a,string"`
			}{},
		},

		// PtrHeadUint8
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
			name:     "PtrHeadUint8OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint8 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint8String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A uint8 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadUint8Ptr
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
			name:     "PtrHeadUint8PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint8 `json:"a,omitempty"`
			}{A: uint8ptr(1)},
		},
		{
			name:     "PtrHeadUint8PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *uint8 `json:"a,string"`
			}{A: uint8ptr(1)},
		},

		// PtrHeadUint8PtrNil
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
			name:     "PtrHeadUint8PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint8 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint8PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint8 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint8Nil
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
			name:     "PtrHeadUint8NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint8 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint8NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint8 `json:"a,string"`
			})(nil),
		},

		// HeadUint8ZeroMultiFields
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
			name:     "HeadUint8ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint8 `json:"a,omitempty"`
				B uint8 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadUint8ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A uint8 `json:"a,string"`
				B uint8 `json:"b,string"`
			}{},
		},

		// HeadUint8MultiFields
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
			name:     "HeadUint8MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint8 `json:"a,omitempty"`
				B uint8 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint8MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A uint8 `json:"a,string"`
				B uint8 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadUint8PtrMultiFields
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
			name:     "HeadUint8PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint8 `json:"a,omitempty"`
				B *uint8 `json:"b,omitempty"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},
		{
			name:     "HeadUint8PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *uint8 `json:"a,string"`
				B *uint8 `json:"b,string"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},

		// HeadUint8PtrNilMultiFields
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
			name:     "HeadUint8PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint8 `json:"a,omitempty"`
				B *uint8 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadUint8PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint8 `json:"a,string"`
				B *uint8 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint8ZeroMultiFields
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
			name:     "PtrHeadUint8ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint8 `json:"a,omitempty"`
				B uint8 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint8ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A uint8 `json:"a,string"`
				B uint8 `json:"b,string"`
			}{},
		},

		// PtrHeadUint8MultiFields
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
			name:     "PtrHeadUint8MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint8 `json:"a,omitempty"`
				B uint8 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint8MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A uint8 `json:"a,string"`
				B uint8 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadUint8PtrMultiFields
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
			name:     "PtrHeadUint8PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint8 `json:"a,omitempty"`
				B *uint8 `json:"b,omitempty"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},
		{
			name:     "PtrHeadUint8PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *uint8 `json:"a,string"`
				B *uint8 `json:"b,string"`
			}{A: uint8ptr(1), B: uint8ptr(2)},
		},

		// PtrHeadUint8PtrNilMultiFields
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
			name:     "PtrHeadUint8PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint8 `json:"a,omitempty"`
				B *uint8 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint8 `json:"a,string"`
				B *uint8 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint8NilMultiFields
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
			name:     "PtrHeadUint8NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint8 `json:"a,omitempty"`
				B *uint8 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint8NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint8 `json:"a,string"`
				B *uint8 `json:"b,string"`
			})(nil),
		},

		// HeadUint8ZeroNotRoot
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
			name:     "HeadUint8ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A uint8 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint8ZeroNotRootString",
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
					A uint8 `json:"a,string"`
				}
			}{},
		},

		// HeadUint8NotRoot
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
			name:     "HeadUint8NotRootOmitEmpty",
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
					A uint8 `json:"a,omitempty"`
				}
			}{A: struct {
				A uint8 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadUint8NotRootString",
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
					A uint8 `json:"a,string"`
				}
			}{A: struct {
				A uint8 `json:"a,string"`
			}{A: 1}},
		},

		// HeadUint8PtrNotRoot
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
			name:     "HeadUint8PtrNotRootOmitEmpty",
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
					A *uint8 `json:"a,omitempty"`
				}
			}{A: struct {
				A *uint8 `json:"a,omitempty"`
			}{uint8ptr(1)}},
		},
		{
			name:     "HeadUint8PtrNotRootString",
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
					A *uint8 `json:"a,string"`
				}
			}{A: struct {
				A *uint8 `json:"a,string"`
			}{uint8ptr(1)}},
		},

		// HeadUint8PtrNilNotRoot
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
			name:     "HeadUint8PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *uint8 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint8PtrNilNotRootString",
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
					A *uint8 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadUint8ZeroNotRoot
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
			name:     "PtrHeadUint8ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A uint8 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A uint8 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadUint8ZeroNotRootString",
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
					A uint8 `json:"a,string"`
				}
			}{A: new(struct {
				A uint8 `json:"a,string"`
			})},
		},

		// PtrHeadUint8NotRoot
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
			name:     "PtrHeadUint8NotRootOmitEmpty",
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
					A uint8 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A uint8 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint8NotRootString",
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
					A uint8 `json:"a,string"`
				}
			}{A: &(struct {
				A uint8 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadUint8PtrNotRoot
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
			name:     "PtrHeadUint8PtrNotRootOmitEmpty",
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
					A *uint8 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,omitempty"`
			}{A: uint8ptr(1)})},
		},
		{
			name:     "PtrHeadUint8PtrNotRootString",
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
					A *uint8 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,string"`
			}{A: uint8ptr(1)})},
		},

		// PtrHeadUint8PtrNilNotRoot
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
			name:     "PtrHeadUint8PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *uint8 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint8PtrNilNotRootString",
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
					A *uint8 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadUint8NilNotRoot
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
			name:     "PtrHeadUint8NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *uint8 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint8NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint8 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadUint8ZeroMultiFieldsNotRoot
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
			name:     "HeadUint8ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A uint8 `json:"a,omitempty"`
				}
				B struct {
					B uint8 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint8ZeroMultiFieldsNotRootString",
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
					A uint8 `json:"a,string"`
				}
				B struct {
					B uint8 `json:"b,string"`
				}
			}{},
		},

		// HeadUint8MultiFieldsNotRoot
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
			name:     "HeadUint8MultiFieldsNotRootOmitEmpty",
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
					A uint8 `json:"a,omitempty"`
				}
				B struct {
					B uint8 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint8 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint8 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadUint8MultiFieldsNotRootString",
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
					A uint8 `json:"a,string"`
				}
				B struct {
					B uint8 `json:"b,string"`
				}
			}{A: struct {
				A uint8 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint8 `json:"b,string"`
			}{B: 2}},
		},

		// HeadUint8PtrMultiFieldsNotRoot
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
			name:     "HeadUint8PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint8 `json:"a,omitempty"`
				}
				B struct {
					B *uint8 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint8 `json:"a,omitempty"`
			}{A: uint8ptr(1)}, B: struct {
				B *uint8 `json:"b,omitempty"`
			}{B: uint8ptr(2)}},
		},
		{
			name:     "HeadUint8PtrMultiFieldsNotRootString",
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
					A *uint8 `json:"a,string"`
				}
				B struct {
					B *uint8 `json:"b,string"`
				}
			}{A: struct {
				A *uint8 `json:"a,string"`
			}{A: uint8ptr(1)}, B: struct {
				B *uint8 `json:"b,string"`
			}{B: uint8ptr(2)}},
		},

		// HeadUint8PtrNilMultiFieldsNotRoot
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
			name:     "HeadUint8PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *uint8 `json:"a,omitempty"`
				}
				B struct {
					B *uint8 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint8 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *uint8 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadUint8PtrNilMultiFieldsNotRootString",
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
					A *uint8 `json:"a,string"`
				}
				B struct {
					B *uint8 `json:"b,string"`
				}
			}{A: struct {
				A *uint8 `json:"a,string"`
			}{A: nil}, B: struct {
				B *uint8 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadUint8ZeroMultiFieldsNotRoot
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
			name:     "PtrHeadUint8ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A uint8 `json:"a,omitempty"`
				}
				B struct {
					B uint8 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint8ZeroMultiFieldsNotRootString",
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
					A uint8 `json:"a,string"`
				}
				B struct {
					B uint8 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadUint8MultiFieldsNotRoot
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
			name:     "PtrHeadUint8MultiFieldsNotRootOmitEmpty",
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
					A uint8 `json:"a,omitempty"`
				}
				B struct {
					B uint8 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint8 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint8 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint8MultiFieldsNotRootString",
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
					A uint8 `json:"a,string"`
				}
				B struct {
					B uint8 `json:"b,string"`
				}
			}{A: struct {
				A uint8 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint8 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadUint8PtrMultiFieldsNotRoot
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
			name:     "PtrHeadUint8PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint8 `json:"a,omitempty"`
				}
				B *struct {
					B *uint8 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,omitempty"`
			}{A: uint8ptr(1)}), B: &(struct {
				B *uint8 `json:"b,omitempty"`
			}{B: uint8ptr(2)})},
		},
		{
			name:     "PtrHeadUint8PtrMultiFieldsNotRootString",
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
					A *uint8 `json:"a,string"`
				}
				B *struct {
					B *uint8 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,string"`
			}{A: uint8ptr(1)}), B: &(struct {
				B *uint8 `json:"b,string"`
			}{B: uint8ptr(2)})},
		},

		// PtrHeadUint8PtrNilMultiFieldsNotRoot
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
			name:     "PtrHeadUint8PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *uint8 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *uint8 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint8NilMultiFieldsNotRoot
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
			name:     "PtrHeadUint8NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint8 `json:"a,omitempty"`
				}
				B *struct {
					B *uint8 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint8NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint8 `json:"a,string"`
				}
				B *struct {
					B *uint8 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint8DoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint8DoubleMultiFieldsNotRootOmitEmpty",
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
					A uint8 `json:"a,omitempty"`
					B uint8 `json:"b,omitempty"`
				}
				B *struct {
					A uint8 `json:"a,omitempty"`
					B uint8 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A uint8 `json:"a,omitempty"`
				B uint8 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A uint8 `json:"a,omitempty"`
				B uint8 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint8DoubleMultiFieldsNotRootString",
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
					A uint8 `json:"a,string"`
					B uint8 `json:"b,string"`
				}
				B *struct {
					A uint8 `json:"a,string"`
					B uint8 `json:"b,string"`
				}
			}{A: &(struct {
				A uint8 `json:"a,string"`
				B uint8 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A uint8 `json:"a,string"`
				B uint8 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadUint8NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint8NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A uint8 `json:"a,omitempty"`
					B uint8 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A uint8 `json:"a,omitempty"`
					B uint8 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A uint8 `json:"a,string"`
					B uint8 `json:"b,string"`
				}
				B *struct {
					A uint8 `json:"a,string"`
					B uint8 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint8NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint8NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint8 `json:"a,omitempty"`
					B uint8 `json:"b,omitempty"`
				}
				B *struct {
					A uint8 `json:"a,omitempty"`
					B uint8 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint8NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint8 `json:"a,string"`
					B uint8 `json:"b,string"`
				}
				B *struct {
					A uint8 `json:"a,string"`
					B uint8 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint8PtrDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint8PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *uint8 `json:"a,omitempty"`
					B *uint8 `json:"b,omitempty"`
				}
				B *struct {
					A *uint8 `json:"a,omitempty"`
					B *uint8 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,omitempty"`
				B *uint8 `json:"b,omitempty"`
			}{A: uint8ptr(1), B: uint8ptr(2)}), B: &(struct {
				A *uint8 `json:"a,omitempty"`
				B *uint8 `json:"b,omitempty"`
			}{A: uint8ptr(3), B: uint8ptr(4)})},
		},
		{
			name:     "PtrHeadUint8PtrDoubleMultiFieldsNotRootString",
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
					A *uint8 `json:"a,string"`
					B *uint8 `json:"b,string"`
				}
				B *struct {
					A *uint8 `json:"a,string"`
					B *uint8 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint8 `json:"a,string"`
				B *uint8 `json:"b,string"`
			}{A: uint8ptr(1), B: uint8ptr(2)}), B: &(struct {
				A *uint8 `json:"a,string"`
				B *uint8 `json:"b,string"`
			}{A: uint8ptr(3), B: uint8ptr(4)})},
		},

		// PtrHeadUint8PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint8PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a,omitempty"`
					B *uint8 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *uint8 `json:"a,omitempty"`
					B *uint8 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint8PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint8 `json:"a,string"`
					B *uint8 `json:"b,string"`
				}
				B *struct {
					A *uint8 `json:"a,string"`
					B *uint8 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint8PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint8PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint8 `json:"a,omitempty"`
					B *uint8 `json:"b,omitempty"`
				}
				B *struct {
					A *uint8 `json:"a,omitempty"`
					B *uint8 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint8PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint8 `json:"a,string"`
					B *uint8 `json:"b,string"`
				}
				B *struct {
					A *uint8 `json:"a,string"`
					B *uint8 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadUint8
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
			name:     "AnonymousHeadUint8OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint8OmitEmpty
				B uint8 `json:"b,omitempty"`
			}{
				structUint8OmitEmpty: structUint8OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "AnonymousHeadUint8String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint8String
				B uint8 `json:"b,string"`
			}{
				structUint8String: structUint8String{A: 1},
				B:                 2,
			},
		},

		// PtrAnonymousHeadUint8
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
			name:     "PtrAnonymousHeadUint8OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint8OmitEmpty
				B uint8 `json:"b,omitempty"`
			}{
				structUint8OmitEmpty: &structUint8OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint8String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint8String
				B uint8 `json:"b,string"`
			}{
				structUint8String: &structUint8String{A: 1},
				B:                 2,
			},
		},

		// NilPtrAnonymousHeadUint8
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
			name:     "NilPtrAnonymousHeadUint8OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint8OmitEmpty
				B uint8 `json:"b,omitempty"`
			}{
				structUint8OmitEmpty: nil,
				B:                    2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint8String
				B uint8 `json:"b,string"`
			}{
				structUint8String: nil,
				B:                 2,
			},
		},

		// AnonymousHeadUint8Ptr
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
			name:     "AnonymousHeadUint8PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint8PtrOmitEmpty
				B *uint8 `json:"b,omitempty"`
			}{
				structUint8PtrOmitEmpty: structUint8PtrOmitEmpty{A: uint8ptr(1)},
				B:                       uint8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint8PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint8PtrString
				B *uint8 `json:"b,string"`
			}{
				structUint8PtrString: structUint8PtrString{A: uint8ptr(1)},
				B:                    uint8ptr(2),
			},
		},

		// AnonymousHeadUint8PtrNil
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
			name:     "AnonymousHeadUint8PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structUint8PtrOmitEmpty
				B *uint8 `json:"b,omitempty"`
			}{
				structUint8PtrOmitEmpty: structUint8PtrOmitEmpty{A: nil},
				B:                       uint8ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint8PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structUint8PtrString
				B *uint8 `json:"b,string"`
			}{
				structUint8PtrString: structUint8PtrString{A: nil},
				B:                    uint8ptr(2),
			},
		},

		// PtrAnonymousHeadUint8Ptr
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
			name:     "PtrAnonymousHeadUint8PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint8PtrOmitEmpty
				B *uint8 `json:"b,omitempty"`
			}{
				structUint8PtrOmitEmpty: &structUint8PtrOmitEmpty{A: uint8ptr(1)},
				B:                       uint8ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint8PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint8PtrString
				B *uint8 `json:"b,string"`
			}{
				structUint8PtrString: &structUint8PtrString{A: uint8ptr(1)},
				B:                    uint8ptr(2),
			},
		},

		// NilPtrAnonymousHeadUint8Ptr
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
			name:     "NilPtrAnonymousHeadUint8PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint8PtrOmitEmpty
				B *uint8 `json:"b,omitempty"`
			}{
				structUint8PtrOmitEmpty: nil,
				B:                       uint8ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint8PtrString
				B *uint8 `json:"b,string"`
			}{
				structUint8PtrString: nil,
				B:                    uint8ptr(2),
			},
		},

		// AnonymousHeadUint8Only
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
			name:     "AnonymousHeadUint8OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint8OmitEmpty
			}{
				structUint8OmitEmpty: structUint8OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadUint8OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint8String
			}{
				structUint8String: structUint8String{A: 1},
			},
		},

		// PtrAnonymousHeadUint8Only
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
			name:     "PtrAnonymousHeadUint8OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint8OmitEmpty
			}{
				structUint8OmitEmpty: &structUint8OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint8OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint8String
			}{
				structUint8String: &structUint8String{A: 1},
			},
		},

		// NilPtrAnonymousHeadUint8Only
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
			name:     "NilPtrAnonymousHeadUint8OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint8OmitEmpty
			}{
				structUint8OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint8String
			}{
				structUint8String: nil,
			},
		},

		// AnonymousHeadUint8PtrOnly
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
			name:     "AnonymousHeadUint8PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint8PtrOmitEmpty
			}{
				structUint8PtrOmitEmpty: structUint8PtrOmitEmpty{A: uint8ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint8PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint8PtrString
			}{
				structUint8PtrString: structUint8PtrString{A: uint8ptr(1)},
			},
		},

		// AnonymousHeadUint8PtrNilOnly
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
			name:     "AnonymousHeadUint8PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structUint8PtrOmitEmpty
			}{
				structUint8PtrOmitEmpty: structUint8PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadUint8PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint8PtrString
			}{
				structUint8PtrString: structUint8PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadUint8PtrOnly
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
			name:     "PtrAnonymousHeadUint8PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint8PtrOmitEmpty
			}{
				structUint8PtrOmitEmpty: &structUint8PtrOmitEmpty{A: uint8ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadUint8PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint8PtrString
			}{
				structUint8PtrString: &structUint8PtrString{A: uint8ptr(1)},
			},
		},

		// NilPtrAnonymousHeadUint8PtrOnly
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
		{
			name:     "NilPtrAnonymousHeadUint8PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint8PtrOmitEmpty
			}{
				structUint8PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint8PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint8PtrString
			}{
				structUint8PtrString: nil,
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
				stdresult := encodeByEncodingJSON(test.data, indent, htmlEscape)
				if buf.String() != stdresult {
					t.Errorf("%s(htmlEscape:%T): doesn't compatible with encoding/json. expected %q but got %q", test.name, htmlEscape, stdresult, buf.String())
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
