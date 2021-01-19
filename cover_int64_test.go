package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverInt64(t *testing.T) {
	type structInt64 struct {
		A int64 `json:"a"`
	}
	type structInt64OmitEmpty struct {
		A int64 `json:"a,omitempty"`
	}
	type structInt64String struct {
		A int64 `json:"a,string"`
	}

	type structInt64Ptr struct {
		A *int64 `json:"a"`
	}
	type structInt64PtrOmitEmpty struct {
		A *int64 `json:"a,omitempty"`
	}
	type structInt64PtrString struct {
		A *int64 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadInt64Zero
		{
			name:     "HeadInt64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A int64 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt64ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int64 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadInt64ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A int64 `json:"a,string"`
			}{},
		},

		// HeadInt64
		{
			name:     "HeadInt64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt64OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int64 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadInt64String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A int64 `json:"a,string"`
			}{A: 1},
		},

		// HeadInt64Ptr
		{
			name:     "HeadInt64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)},
		},
		{
			name:     "HeadInt64PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int64 `json:"a,omitempty"`
			}{A: int64ptr(1)},
		},
		{
			name:     "HeadInt64PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *int64 `json:"a,string"`
			}{A: int64ptr(1)},
		},

		// HeadInt64PtrNil
		{
			name:     "HeadInt64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadInt64PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int64 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadInt64PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int64 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt64Zero
		{
			name:     "PtrHeadInt64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A int64 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt64ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int64 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt64ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A int64 `json:"a,string"`
			}{},
		},

		// PtrHeadInt64
		{
			name:     "PtrHeadInt64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt64OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int64 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt64String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A int64 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadInt64Ptr
		{
			name:     "PtrHeadInt64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)},
		},
		{
			name:     "PtrHeadInt64PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int64 `json:"a,omitempty"`
			}{A: int64ptr(1)},
		},
		{
			name:     "PtrHeadInt64PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *int64 `json:"a,string"`
			}{A: int64ptr(1)},
		},

		// PtrHeadInt64PtrNil
		{
			name:     "PtrHeadInt64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int64 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int64 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt64Nil
		{
			name:     "PtrHeadInt64Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int64 `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int64 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int64 `json:"a,string"`
			})(nil),
		},

		// HeadInt64ZeroMultiFields
		{
			name:     "HeadInt64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt64ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int64 `json:"a,omitempty"`
				B int64 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadInt64ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A int64 `json:"a,string"`
				B int64 `json:"b,string"`
			}{},
		},

		// HeadInt64MultiFields
		{
			name:     "HeadInt64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt64MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int64 `json:"a,omitempty"`
				B int64 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt64MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A int64 `json:"a,string"`
				B int64 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadInt64PtrMultiFields
		{
			name:     "HeadInt64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},
		{
			name:     "HeadInt64PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int64 `json:"a,omitempty"`
				B *int64 `json:"b,omitempty"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},
		{
			name:     "HeadInt64PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *int64 `json:"a,string"`
				B *int64 `json:"b,string"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},

		// HeadInt64PtrNilMultiFields
		{
			name:     "HeadInt64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadInt64PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int64 `json:"a,omitempty"`
				B *int64 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadInt64PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int64 `json:"a,string"`
				B *int64 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt64ZeroMultiFields
		{
			name:     "PtrHeadInt64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt64ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int64 `json:"a,omitempty"`
				B int64 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt64ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A int64 `json:"a,string"`
				B int64 `json:"b,string"`
			}{},
		},

		// PtrHeadInt64MultiFields
		{
			name:     "PtrHeadInt64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt64MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int64 `json:"a,omitempty"`
				B int64 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt64MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A int64 `json:"a,string"`
				B int64 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadInt64PtrMultiFields
		{
			name:     "PtrHeadInt64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},
		{
			name:     "PtrHeadInt64PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int64 `json:"a,omitempty"`
				B *int64 `json:"b,omitempty"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},
		{
			name:     "PtrHeadInt64PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *int64 `json:"a,string"`
				B *int64 `json:"b,string"`
			}{A: int64ptr(1), B: int64ptr(2)},
		},

		// PtrHeadInt64PtrNilMultiFields
		{
			name:     "PtrHeadInt64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int64 `json:"a,omitempty"`
				B *int64 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int64 `json:"a,string"`
				B *int64 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt64NilMultiFields
		{
			name:     "PtrHeadInt64NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int64 `json:"a,omitempty"`
				B *int64 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int64 `json:"a,string"`
				B *int64 `json:"b,string"`
			})(nil),
		},

		// HeadInt64ZeroNotRoot
		{
			name:     "HeadInt64ZeroNotRoot",
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
					A int64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt64ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A int64 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt64ZeroNotRootString",
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
					A int64 `json:"a,string"`
				}
			}{},
		},

		// HeadInt64NotRoot
		{
			name:     "HeadInt64NotRoot",
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
					A int64 `json:"a"`
				}
			}{A: struct {
				A int64 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt64NotRootOmitEmpty",
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
					A int64 `json:"a,omitempty"`
				}
			}{A: struct {
				A int64 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadInt64NotRootString",
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
					A int64 `json:"a,string"`
				}
			}{A: struct {
				A int64 `json:"a,string"`
			}{A: 1}},
		},

		// HeadInt64PtrNotRoot
		{
			name:     "HeadInt64PtrNotRoot",
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
					A *int64 `json:"a"`
				}
			}{A: struct {
				A *int64 `json:"a"`
			}{int64ptr(1)}},
		},
		{
			name:     "HeadInt64PtrNotRootOmitEmpty",
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
					A *int64 `json:"a,omitempty"`
				}
			}{A: struct {
				A *int64 `json:"a,omitempty"`
			}{int64ptr(1)}},
		},
		{
			name:     "HeadInt64PtrNotRootString",
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
					A *int64 `json:"a,string"`
				}
			}{A: struct {
				A *int64 `json:"a,string"`
			}{int64ptr(1)}},
		},

		// HeadInt64PtrNilNotRoot
		{
			name:     "HeadInt64PtrNilNotRoot",
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
					A *int64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt64PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *int64 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt64PtrNilNotRootString",
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
					A *int64 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadInt64ZeroNotRoot
		{
			name:     "PtrHeadInt64ZeroNotRoot",
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
					A int64 `json:"a"`
				}
			}{A: new(struct {
				A int64 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt64ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A int64 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A int64 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadInt64ZeroNotRootString",
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
					A int64 `json:"a,string"`
				}
			}{A: new(struct {
				A int64 `json:"a,string"`
			})},
		},

		// PtrHeadInt64NotRoot
		{
			name:     "PtrHeadInt64NotRoot",
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
					A int64 `json:"a"`
				}
			}{A: &(struct {
				A int64 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt64NotRootOmitEmpty",
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
					A int64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A int64 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt64NotRootString",
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
					A int64 `json:"a,string"`
				}
			}{A: &(struct {
				A int64 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadInt64PtrNotRoot
		{
			name:     "PtrHeadInt64PtrNotRoot",
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
					A *int64 `json:"a"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)})},
		},
		{
			name:     "PtrHeadInt64PtrNotRootOmitEmpty",
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
					A *int64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int64 `json:"a,omitempty"`
			}{A: int64ptr(1)})},
		},
		{
			name:     "PtrHeadInt64PtrNotRootString",
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
					A *int64 `json:"a,string"`
				}
			}{A: &(struct {
				A *int64 `json:"a,string"`
			}{A: int64ptr(1)})},
		},

		// PtrHeadInt64PtrNilNotRoot
		{
			name:     "PtrHeadInt64PtrNilNotRoot",
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
					A *int64 `json:"a"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt64PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *int64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int64 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt64PtrNilNotRootString",
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
					A *int64 `json:"a,string"`
				}
			}{A: &(struct {
				A *int64 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadInt64NilNotRoot
		{
			name:     "PtrHeadInt64NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int64 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadInt64NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *int64 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt64NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int64 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadInt64ZeroMultiFieldsNotRoot
		{
			name:     "HeadInt64ZeroMultiFieldsNotRoot",
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
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadInt64ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A int64 `json:"a,omitempty"`
				}
				B struct {
					B int64 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt64ZeroMultiFieldsNotRootString",
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
					A int64 `json:"a,string"`
				}
				B struct {
					B int64 `json:"b,string"`
				}
			}{},
		},

		// HeadInt64MultiFieldsNotRoot
		{
			name:     "HeadInt64MultiFieldsNotRoot",
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
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{A: struct {
				A int64 `json:"a"`
			}{A: 1}, B: struct {
				B int64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadInt64MultiFieldsNotRootOmitEmpty",
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
					A int64 `json:"a,omitempty"`
				}
				B struct {
					B int64 `json:"b,omitempty"`
				}
			}{A: struct {
				A int64 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int64 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadInt64MultiFieldsNotRootString",
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
					A int64 `json:"a,string"`
				}
				B struct {
					B int64 `json:"b,string"`
				}
			}{A: struct {
				A int64 `json:"a,string"`
			}{A: 1}, B: struct {
				B int64 `json:"b,string"`
			}{B: 2}},
		},

		// HeadInt64PtrMultiFieldsNotRoot
		{
			name:     "HeadInt64PtrMultiFieldsNotRoot",
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
					A *int64 `json:"a"`
				}
				B struct {
					B *int64 `json:"b"`
				}
			}{A: struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)}, B: struct {
				B *int64 `json:"b"`
			}{B: int64ptr(2)}},
		},
		{
			name:     "HeadInt64PtrMultiFieldsNotRootOmitEmpty",
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
					A *int64 `json:"a,omitempty"`
				}
				B struct {
					B *int64 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int64 `json:"a,omitempty"`
			}{A: int64ptr(1)}, B: struct {
				B *int64 `json:"b,omitempty"`
			}{B: int64ptr(2)}},
		},
		{
			name:     "HeadInt64PtrMultiFieldsNotRootString",
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
					A *int64 `json:"a,string"`
				}
				B struct {
					B *int64 `json:"b,string"`
				}
			}{A: struct {
				A *int64 `json:"a,string"`
			}{A: int64ptr(1)}, B: struct {
				B *int64 `json:"b,string"`
			}{B: int64ptr(2)}},
		},

		// HeadInt64PtrNilMultiFieldsNotRoot
		{
			name:     "HeadInt64PtrNilMultiFieldsNotRoot",
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
					A *int64 `json:"a"`
				}
				B struct {
					B *int64 `json:"b"`
				}
			}{A: struct {
				A *int64 `json:"a"`
			}{A: nil}, B: struct {
				B *int64 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "HeadInt64PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *int64 `json:"a,omitempty"`
				}
				B struct {
					B *int64 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int64 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *int64 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadInt64PtrNilMultiFieldsNotRootString",
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
					A *int64 `json:"a,string"`
				}
				B struct {
					B *int64 `json:"b,string"`
				}
			}{A: struct {
				A *int64 `json:"a,string"`
			}{A: nil}, B: struct {
				B *int64 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadInt64ZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64ZeroMultiFieldsNotRoot",
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
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt64ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A int64 `json:"a,omitempty"`
				}
				B struct {
					B int64 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt64ZeroMultiFieldsNotRootString",
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
					A int64 `json:"a,string"`
				}
				B struct {
					B int64 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadInt64MultiFieldsNotRoot
		{
			name:     "PtrHeadInt64MultiFieldsNotRoot",
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
					A int64 `json:"a"`
				}
				B struct {
					B int64 `json:"b"`
				}
			}{A: struct {
				A int64 `json:"a"`
			}{A: 1}, B: struct {
				B int64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt64MultiFieldsNotRootOmitEmpty",
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
					A int64 `json:"a,omitempty"`
				}
				B struct {
					B int64 `json:"b,omitempty"`
				}
			}{A: struct {
				A int64 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int64 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt64MultiFieldsNotRootString",
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
					A int64 `json:"a,string"`
				}
				B struct {
					B int64 `json:"b,string"`
				}
			}{A: struct {
				A int64 `json:"a,string"`
			}{A: 1}, B: struct {
				B int64 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadInt64PtrMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64PtrMultiFieldsNotRoot",
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
					A *int64 `json:"a"`
				}
				B *struct {
					B *int64 `json:"b"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
			}{A: int64ptr(1)}), B: &(struct {
				B *int64 `json:"b"`
			}{B: int64ptr(2)})},
		},
		{
			name:     "PtrHeadInt64PtrMultiFieldsNotRootOmitEmpty",
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
					A *int64 `json:"a,omitempty"`
				}
				B *struct {
					B *int64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int64 `json:"a,omitempty"`
			}{A: int64ptr(1)}), B: &(struct {
				B *int64 `json:"b,omitempty"`
			}{B: int64ptr(2)})},
		},
		{
			name:     "PtrHeadInt64PtrMultiFieldsNotRootString",
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
					A *int64 `json:"a,string"`
				}
				B *struct {
					B *int64 `json:"b,string"`
				}
			}{A: &(struct {
				A *int64 `json:"a,string"`
			}{A: int64ptr(1)}), B: &(struct {
				B *int64 `json:"b,string"`
			}{B: int64ptr(2)})},
		},

		// PtrHeadInt64PtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int64 `json:"a"`
				}
				B *struct {
					B *int64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int64 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *int64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int64 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *int64 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt64NilMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64NilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a"`
				}
				B *struct {
					B *int64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a,omitempty"`
				}
				B *struct {
					B *int64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a,string"`
				}
				B *struct {
					B *int64 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt64DoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64DoubleMultiFieldsNotRoot",
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
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
				B *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
			}{A: &(struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int64 `json:"a"`
				B int64 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt64DoubleMultiFieldsNotRootOmitEmpty",
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
					A int64 `json:"a,omitempty"`
					B int64 `json:"b,omitempty"`
				}
				B *struct {
					A int64 `json:"a,omitempty"`
					B int64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A int64 `json:"a,omitempty"`
				B int64 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A int64 `json:"a,omitempty"`
				B int64 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt64DoubleMultiFieldsNotRootString",
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
					A int64 `json:"a,string"`
					B int64 `json:"b,string"`
				}
				B *struct {
					A int64 `json:"a,string"`
					B int64 `json:"b,string"`
				}
			}{A: &(struct {
				A int64 `json:"a,string"`
				B int64 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A int64 `json:"a,string"`
				B int64 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadInt64NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
				B *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A int64 `json:"a,omitempty"`
					B int64 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A int64 `json:"a,omitempty"`
					B int64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int64 `json:"a,string"`
					B int64 `json:"b,string"`
				}
				B *struct {
					A int64 `json:"a,string"`
					B int64 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt64NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
				B *struct {
					A int64 `json:"a"`
					B int64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int64 `json:"a,omitempty"`
					B int64 `json:"b,omitempty"`
				}
				B *struct {
					A int64 `json:"a,omitempty"`
					B int64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int64 `json:"a,string"`
					B int64 `json:"b,string"`
				}
				B *struct {
					A int64 `json:"a,string"`
					B int64 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt64PtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64PtrDoubleMultiFieldsNotRoot",
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
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
				B *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
			}{A: &(struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(1), B: int64ptr(2)}), B: &(struct {
				A *int64 `json:"a"`
				B *int64 `json:"b"`
			}{A: int64ptr(3), B: int64ptr(4)})},
		},
		{
			name:     "PtrHeadInt64PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *int64 `json:"a,omitempty"`
					B *int64 `json:"b,omitempty"`
				}
				B *struct {
					A *int64 `json:"a,omitempty"`
					B *int64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int64 `json:"a,omitempty"`
				B *int64 `json:"b,omitempty"`
			}{A: int64ptr(1), B: int64ptr(2)}), B: &(struct {
				A *int64 `json:"a,omitempty"`
				B *int64 `json:"b,omitempty"`
			}{A: int64ptr(3), B: int64ptr(4)})},
		},
		{
			name:     "PtrHeadInt64PtrDoubleMultiFieldsNotRootString",
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
					A *int64 `json:"a,string"`
					B *int64 `json:"b,string"`
				}
				B *struct {
					A *int64 `json:"a,string"`
					B *int64 `json:"b,string"`
				}
			}{A: &(struct {
				A *int64 `json:"a,string"`
				B *int64 `json:"b,string"`
			}{A: int64ptr(1), B: int64ptr(2)}), B: &(struct {
				A *int64 `json:"a,string"`
				B *int64 `json:"b,string"`
			}{A: int64ptr(3), B: int64ptr(4)})},
		},

		// PtrHeadInt64PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
				B *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int64 `json:"a,omitempty"`
					B *int64 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *int64 `json:"a,omitempty"`
					B *int64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int64 `json:"a,string"`
					B *int64 `json:"b,string"`
				}
				B *struct {
					A *int64 `json:"a,string"`
					B *int64 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt64PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
				B *struct {
					A *int64 `json:"a"`
					B *int64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a,omitempty"`
					B *int64 `json:"b,omitempty"`
				}
				B *struct {
					A *int64 `json:"a,omitempty"`
					B *int64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt64PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int64 `json:"a,string"`
					B *int64 `json:"b,string"`
				}
				B *struct {
					A *int64 `json:"a,string"`
					B *int64 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadInt64
		{
			name:     "AnonymousHeadInt64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt64
				B int64 `json:"b"`
			}{
				structInt64: structInt64{A: 1},
				B:           2,
			},
		},
		{
			name:     "AnonymousHeadInt64OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt64OmitEmpty
				B int64 `json:"b,omitempty"`
			}{
				structInt64OmitEmpty: structInt64OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "AnonymousHeadInt64String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt64String
				B int64 `json:"b,string"`
			}{
				structInt64String: structInt64String{A: 1},
				B:                 2,
			},
		},

		// PtrAnonymousHeadInt64
		{
			name:     "PtrAnonymousHeadInt64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt64
				B int64 `json:"b"`
			}{
				structInt64: &structInt64{A: 1},
				B:           2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt64OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt64OmitEmpty
				B int64 `json:"b,omitempty"`
			}{
				structInt64OmitEmpty: &structInt64OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt64String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt64String
				B int64 `json:"b,string"`
			}{
				structInt64String: &structInt64String{A: 1},
				B:                 2,
			},
		},

		// NilPtrAnonymousHeadInt64
		{
			name:     "NilPtrAnonymousHeadInt64",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt64
				B int64 `json:"b"`
			}{
				structInt64: nil,
				B:           2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt64OmitEmpty
				B int64 `json:"b,omitempty"`
			}{
				structInt64OmitEmpty: nil,
				B:                    2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt64String
				B int64 `json:"b,string"`
			}{
				structInt64String: nil,
				B:                 2,
			},
		},

		// AnonymousHeadInt64Ptr
		{
			name:     "AnonymousHeadInt64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: structInt64Ptr{A: int64ptr(1)},
				B:              int64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt64PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt64PtrOmitEmpty
				B *int64 `json:"b,omitempty"`
			}{
				structInt64PtrOmitEmpty: structInt64PtrOmitEmpty{A: int64ptr(1)},
				B:                       int64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt64PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt64PtrString
				B *int64 `json:"b,string"`
			}{
				structInt64PtrString: structInt64PtrString{A: int64ptr(1)},
				B:                    int64ptr(2),
			},
		},

		// AnonymousHeadInt64PtrNil
		{
			name:     "AnonymousHeadInt64PtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: structInt64Ptr{A: nil},
				B:              int64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt64PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structInt64PtrOmitEmpty
				B *int64 `json:"b,omitempty"`
			}{
				structInt64PtrOmitEmpty: structInt64PtrOmitEmpty{A: nil},
				B:                       int64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt64PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structInt64PtrString
				B *int64 `json:"b,string"`
			}{
				structInt64PtrString: structInt64PtrString{A: nil},
				B:                    int64ptr(2),
			},
		},

		// PtrAnonymousHeadInt64Ptr
		{
			name:     "PtrAnonymousHeadInt64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: &structInt64Ptr{A: int64ptr(1)},
				B:              int64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt64PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt64PtrOmitEmpty
				B *int64 `json:"b,omitempty"`
			}{
				structInt64PtrOmitEmpty: &structInt64PtrOmitEmpty{A: int64ptr(1)},
				B:                       int64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt64PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt64PtrString
				B *int64 `json:"b,string"`
			}{
				structInt64PtrString: &structInt64PtrString{A: int64ptr(1)},
				B:                    int64ptr(2),
			},
		},

		// NilPtrAnonymousHeadInt64Ptr
		{
			name:     "NilPtrAnonymousHeadInt64Ptr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt64Ptr
				B *int64 `json:"b"`
			}{
				structInt64Ptr: nil,
				B:              int64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt64PtrOmitEmpty
				B *int64 `json:"b,omitempty"`
			}{
				structInt64PtrOmitEmpty: nil,
				B:                       int64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt64PtrString
				B *int64 `json:"b,string"`
			}{
				structInt64PtrString: nil,
				B:                    int64ptr(2),
			},
		},

		// AnonymousHeadInt64Only
		{
			name:     "AnonymousHeadInt64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt64
			}{
				structInt64: structInt64{A: 1},
			},
		},
		{
			name:     "AnonymousHeadInt64OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt64OmitEmpty
			}{
				structInt64OmitEmpty: structInt64OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadInt64OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt64String
			}{
				structInt64String: structInt64String{A: 1},
			},
		},

		// PtrAnonymousHeadInt64Only
		{
			name:     "PtrAnonymousHeadInt64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt64
			}{
				structInt64: &structInt64{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt64OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt64OmitEmpty
			}{
				structInt64OmitEmpty: &structInt64OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt64OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt64String
			}{
				structInt64String: &structInt64String{A: 1},
			},
		},

		// NilPtrAnonymousHeadInt64Only
		{
			name:     "NilPtrAnonymousHeadInt64Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt64
			}{
				structInt64: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt64OmitEmpty
			}{
				structInt64OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt64String
			}{
				structInt64String: nil,
			},
		},

		// AnonymousHeadInt64PtrOnly
		{
			name:     "AnonymousHeadInt64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt64Ptr
			}{
				structInt64Ptr: structInt64Ptr{A: int64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt64PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt64PtrOmitEmpty
			}{
				structInt64PtrOmitEmpty: structInt64PtrOmitEmpty{A: int64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt64PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt64PtrString
			}{
				structInt64PtrString: structInt64PtrString{A: int64ptr(1)},
			},
		},

		// AnonymousHeadInt64PtrNilOnly
		{
			name:     "AnonymousHeadInt64PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt64Ptr
			}{
				structInt64Ptr: structInt64Ptr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadInt64PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structInt64PtrOmitEmpty
			}{
				structInt64PtrOmitEmpty: structInt64PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadInt64PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt64PtrString
			}{
				structInt64PtrString: structInt64PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadInt64PtrOnly
		{
			name:     "PtrAnonymousHeadInt64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt64Ptr
			}{
				structInt64Ptr: &structInt64Ptr{A: int64ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadInt64PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt64PtrOmitEmpty
			}{
				structInt64PtrOmitEmpty: &structInt64PtrOmitEmpty{A: int64ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadInt64PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt64PtrString
			}{
				structInt64PtrString: &structInt64PtrString{A: int64ptr(1)},
			},
		},

		// NilPtrAnonymousHeadInt64PtrOnly
		{
			name:     "NilPtrAnonymousHeadInt64PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt64Ptr
			}{
				structInt64Ptr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt64PtrOmitEmpty
			}{
				structInt64PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt64PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt64PtrString
			}{
				structInt64PtrString: nil,
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
