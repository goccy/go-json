package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverInt32(t *testing.T) {
	type structInt32 struct {
		A int32 `json:"a"`
	}
	type structInt32OmitEmpty struct {
		A int32 `json:"a,omitempty"`
	}
	type structInt32String struct {
		A int32 `json:"a,string"`
	}

	type structInt32Ptr struct {
		A *int32 `json:"a"`
	}
	type structInt32PtrOmitEmpty struct {
		A *int32 `json:"a,omitempty"`
	}
	type structInt32PtrString struct {
		A *int32 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadInt32Zero
		{
			name:     "HeadInt32Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A int32 `json:"a"`
			}{},
		},
		{
			name:     "HeadInt32ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int32 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadInt32ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A int32 `json:"a,string"`
			}{},
		},

		// HeadInt32
		{
			name:     "HeadInt32",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadInt32OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int32 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadInt32String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A int32 `json:"a,string"`
			}{A: 1},
		},

		// HeadInt32Ptr
		{
			name:     "HeadInt32Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)},
		},
		{
			name:     "HeadInt32PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int32 `json:"a,omitempty"`
			}{A: int32ptr(1)},
		},
		{
			name:     "HeadInt32PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *int32 `json:"a,string"`
			}{A: int32ptr(1)},
		},

		// HeadInt32PtrNil
		{
			name:     "HeadInt32PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadInt32PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int32 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadInt32PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int32 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt32Zero
		{
			name:     "PtrHeadInt32Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A int32 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadInt32ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int32 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt32ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A int32 `json:"a,string"`
			}{},
		},

		// PtrHeadInt32
		{
			name:     "PtrHeadInt32",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int32 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt32OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int32 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt32String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A int32 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadInt32Ptr
		{
			name:     "PtrHeadInt32Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)},
		},
		{
			name:     "PtrHeadInt32PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int32 `json:"a,omitempty"`
			}{A: int32ptr(1)},
		},
		{
			name:     "PtrHeadInt32PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *int32 `json:"a,string"`
			}{A: int32ptr(1)},
		},

		// PtrHeadInt32PtrNil
		{
			name:     "PtrHeadInt32PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int32 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int32 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int32 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt32Nil
		{
			name:     "PtrHeadInt32Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int32 `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int32 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int32 `json:"a,string"`
			})(nil),
		},

		// HeadInt32ZeroMultiFields
		{
			name:     "HeadInt32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{},
		},
		{
			name:     "HeadInt32ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int32 `json:"a,omitempty"`
				B int32 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadInt32ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A int32 `json:"a,string"`
				B int32 `json:"b,string"`
			}{},
		},

		// HeadInt32MultiFields
		{
			name:     "HeadInt32MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt32MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int32 `json:"a,omitempty"`
				B int32 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt32MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A int32 `json:"a,string"`
				B int32 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadInt32PtrMultiFields
		{
			name:     "HeadInt32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},
		{
			name:     "HeadInt32PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int32 `json:"a,omitempty"`
				B *int32 `json:"b,omitempty"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},
		{
			name:     "HeadInt32PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *int32 `json:"a,string"`
				B *int32 `json:"b,string"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},

		// HeadInt32PtrNilMultiFields
		{
			name:     "HeadInt32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadInt32PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int32 `json:"a,omitempty"`
				B *int32 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadInt32PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int32 `json:"a,string"`
				B *int32 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt32ZeroMultiFields
		{
			name:     "PtrHeadInt32ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadInt32ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int32 `json:"a,omitempty"`
				B int32 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt32ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A int32 `json:"a,string"`
				B int32 `json:"b,string"`
			}{},
		},

		// PtrHeadInt32MultiFields
		{
			name:     "PtrHeadInt32MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt32MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int32 `json:"a,omitempty"`
				B int32 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt32MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A int32 `json:"a,string"`
				B int32 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadInt32PtrMultiFields
		{
			name:     "PtrHeadInt32PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},
		{
			name:     "PtrHeadInt32PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int32 `json:"a,omitempty"`
				B *int32 `json:"b,omitempty"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},
		{
			name:     "PtrHeadInt32PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *int32 `json:"a,string"`
				B *int32 `json:"b,string"`
			}{A: int32ptr(1), B: int32ptr(2)},
		},

		// PtrHeadInt32PtrNilMultiFields
		{
			name:     "PtrHeadInt32PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int32 `json:"a,omitempty"`
				B *int32 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int32 `json:"a,string"`
				B *int32 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt32NilMultiFields
		{
			name:     "PtrHeadInt32NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int32 `json:"a,omitempty"`
				B *int32 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int32 `json:"a,string"`
				B *int32 `json:"b,string"`
			})(nil),
		},

		// HeadInt32ZeroNotRoot
		{
			name:     "HeadInt32ZeroNotRoot",
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
					A int32 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt32ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A int32 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt32ZeroNotRootString",
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
					A int32 `json:"a,string"`
				}
			}{},
		},

		// HeadInt32NotRoot
		{
			name:     "HeadInt32NotRoot",
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
					A int32 `json:"a"`
				}
			}{A: struct {
				A int32 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadInt32NotRootOmitEmpty",
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
					A int32 `json:"a,omitempty"`
				}
			}{A: struct {
				A int32 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadInt32NotRootString",
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
					A int32 `json:"a,string"`
				}
			}{A: struct {
				A int32 `json:"a,string"`
			}{A: 1}},
		},

		// HeadInt32PtrNotRoot
		{
			name:     "HeadInt32PtrNotRoot",
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
					A *int32 `json:"a"`
				}
			}{A: struct {
				A *int32 `json:"a"`
			}{int32ptr(1)}},
		},
		{
			name:     "HeadInt32PtrNotRootOmitEmpty",
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
					A *int32 `json:"a,omitempty"`
				}
			}{A: struct {
				A *int32 `json:"a,omitempty"`
			}{int32ptr(1)}},
		},
		{
			name:     "HeadInt32PtrNotRootString",
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
					A *int32 `json:"a,string"`
				}
			}{A: struct {
				A *int32 `json:"a,string"`
			}{int32ptr(1)}},
		},

		// HeadInt32PtrNilNotRoot
		{
			name:     "HeadInt32PtrNilNotRoot",
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
					A *int32 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadInt32PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *int32 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt32PtrNilNotRootString",
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
					A *int32 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadInt32ZeroNotRoot
		{
			name:     "PtrHeadInt32ZeroNotRoot",
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
					A int32 `json:"a"`
				}
			}{A: new(struct {
				A int32 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadInt32ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A int32 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A int32 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadInt32ZeroNotRootString",
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
					A int32 `json:"a,string"`
				}
			}{A: new(struct {
				A int32 `json:"a,string"`
			})},
		},

		// PtrHeadInt32NotRoot
		{
			name:     "PtrHeadInt32NotRoot",
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
					A int32 `json:"a"`
				}
			}{A: &(struct {
				A int32 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt32NotRootOmitEmpty",
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
					A int32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A int32 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt32NotRootString",
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
					A int32 `json:"a,string"`
				}
			}{A: &(struct {
				A int32 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadInt32PtrNotRoot
		{
			name:     "PtrHeadInt32PtrNotRoot",
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
					A *int32 `json:"a"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)})},
		},
		{
			name:     "PtrHeadInt32PtrNotRootOmitEmpty",
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
					A *int32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int32 `json:"a,omitempty"`
			}{A: int32ptr(1)})},
		},
		{
			name:     "PtrHeadInt32PtrNotRootString",
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
					A *int32 `json:"a,string"`
				}
			}{A: &(struct {
				A *int32 `json:"a,string"`
			}{A: int32ptr(1)})},
		},

		// PtrHeadInt32PtrNilNotRoot
		{
			name:     "PtrHeadInt32PtrNilNotRoot",
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
					A *int32 `json:"a"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt32PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *int32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int32 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt32PtrNilNotRootString",
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
					A *int32 `json:"a,string"`
				}
			}{A: &(struct {
				A *int32 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadInt32NilNotRoot
		{
			name:     "PtrHeadInt32NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int32 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadInt32NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *int32 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt32NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int32 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadInt32ZeroMultiFieldsNotRoot
		{
			name:     "HeadInt32ZeroMultiFieldsNotRoot",
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
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadInt32ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A int32 `json:"a,omitempty"`
				}
				B struct {
					B int32 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt32ZeroMultiFieldsNotRootString",
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
					A int32 `json:"a,string"`
				}
				B struct {
					B int32 `json:"b,string"`
				}
			}{},
		},

		// HeadInt32MultiFieldsNotRoot
		{
			name:     "HeadInt32MultiFieldsNotRoot",
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
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{A: struct {
				A int32 `json:"a"`
			}{A: 1}, B: struct {
				B int32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadInt32MultiFieldsNotRootOmitEmpty",
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
					A int32 `json:"a,omitempty"`
				}
				B struct {
					B int32 `json:"b,omitempty"`
				}
			}{A: struct {
				A int32 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int32 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadInt32MultiFieldsNotRootString",
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
					A int32 `json:"a,string"`
				}
				B struct {
					B int32 `json:"b,string"`
				}
			}{A: struct {
				A int32 `json:"a,string"`
			}{A: 1}, B: struct {
				B int32 `json:"b,string"`
			}{B: 2}},
		},

		// HeadInt32PtrMultiFieldsNotRoot
		{
			name:     "HeadInt32PtrMultiFieldsNotRoot",
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
					A *int32 `json:"a"`
				}
				B struct {
					B *int32 `json:"b"`
				}
			}{A: struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)}, B: struct {
				B *int32 `json:"b"`
			}{B: int32ptr(2)}},
		},
		{
			name:     "HeadInt32PtrMultiFieldsNotRootOmitEmpty",
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
					A *int32 `json:"a,omitempty"`
				}
				B struct {
					B *int32 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int32 `json:"a,omitempty"`
			}{A: int32ptr(1)}, B: struct {
				B *int32 `json:"b,omitempty"`
			}{B: int32ptr(2)}},
		},
		{
			name:     "HeadInt32PtrMultiFieldsNotRootString",
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
					A *int32 `json:"a,string"`
				}
				B struct {
					B *int32 `json:"b,string"`
				}
			}{A: struct {
				A *int32 `json:"a,string"`
			}{A: int32ptr(1)}, B: struct {
				B *int32 `json:"b,string"`
			}{B: int32ptr(2)}},
		},

		// HeadInt32PtrNilMultiFieldsNotRoot
		{
			name:     "HeadInt32PtrNilMultiFieldsNotRoot",
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
					A *int32 `json:"a"`
				}
				B struct {
					B *int32 `json:"b"`
				}
			}{A: struct {
				A *int32 `json:"a"`
			}{A: nil}, B: struct {
				B *int32 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "HeadInt32PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *int32 `json:"a,omitempty"`
				}
				B struct {
					B *int32 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int32 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *int32 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadInt32PtrNilMultiFieldsNotRootString",
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
					A *int32 `json:"a,string"`
				}
				B struct {
					B *int32 `json:"b,string"`
				}
			}{A: struct {
				A *int32 `json:"a,string"`
			}{A: nil}, B: struct {
				B *int32 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadInt32ZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32ZeroMultiFieldsNotRoot",
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
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt32ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A int32 `json:"a,omitempty"`
				}
				B struct {
					B int32 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt32ZeroMultiFieldsNotRootString",
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
					A int32 `json:"a,string"`
				}
				B struct {
					B int32 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadInt32MultiFieldsNotRoot
		{
			name:     "PtrHeadInt32MultiFieldsNotRoot",
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
					A int32 `json:"a"`
				}
				B struct {
					B int32 `json:"b"`
				}
			}{A: struct {
				A int32 `json:"a"`
			}{A: 1}, B: struct {
				B int32 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt32MultiFieldsNotRootOmitEmpty",
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
					A int32 `json:"a,omitempty"`
				}
				B struct {
					B int32 `json:"b,omitempty"`
				}
			}{A: struct {
				A int32 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int32 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt32MultiFieldsNotRootString",
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
					A int32 `json:"a,string"`
				}
				B struct {
					B int32 `json:"b,string"`
				}
			}{A: struct {
				A int32 `json:"a,string"`
			}{A: 1}, B: struct {
				B int32 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadInt32PtrMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32PtrMultiFieldsNotRoot",
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
					A *int32 `json:"a"`
				}
				B *struct {
					B *int32 `json:"b"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
			}{A: int32ptr(1)}), B: &(struct {
				B *int32 `json:"b"`
			}{B: int32ptr(2)})},
		},
		{
			name:     "PtrHeadInt32PtrMultiFieldsNotRootOmitEmpty",
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
					A *int32 `json:"a,omitempty"`
				}
				B *struct {
					B *int32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int32 `json:"a,omitempty"`
			}{A: int32ptr(1)}), B: &(struct {
				B *int32 `json:"b,omitempty"`
			}{B: int32ptr(2)})},
		},
		{
			name:     "PtrHeadInt32PtrMultiFieldsNotRootString",
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
					A *int32 `json:"a,string"`
				}
				B *struct {
					B *int32 `json:"b,string"`
				}
			}{A: &(struct {
				A *int32 `json:"a,string"`
			}{A: int32ptr(1)}), B: &(struct {
				B *int32 `json:"b,string"`
			}{B: int32ptr(2)})},
		},

		// PtrHeadInt32PtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int32 `json:"a"`
				}
				B *struct {
					B *int32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int32 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *int32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int32 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *int32 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt32NilMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32NilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a"`
				}
				B *struct {
					B *int32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a,omitempty"`
				}
				B *struct {
					B *int32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a,string"`
				}
				B *struct {
					B *int32 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt32DoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32DoubleMultiFieldsNotRoot",
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
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
				B *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
			}{A: &(struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int32 `json:"a"`
				B int32 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt32DoubleMultiFieldsNotRootOmitEmpty",
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
					A int32 `json:"a,omitempty"`
					B int32 `json:"b,omitempty"`
				}
				B *struct {
					A int32 `json:"a,omitempty"`
					B int32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A int32 `json:"a,omitempty"`
				B int32 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A int32 `json:"a,omitempty"`
				B int32 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt32DoubleMultiFieldsNotRootString",
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
					A int32 `json:"a,string"`
					B int32 `json:"b,string"`
				}
				B *struct {
					A int32 `json:"a,string"`
					B int32 `json:"b,string"`
				}
			}{A: &(struct {
				A int32 `json:"a,string"`
				B int32 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A int32 `json:"a,string"`
				B int32 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadInt32NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
				B *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A int32 `json:"a,omitempty"`
					B int32 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A int32 `json:"a,omitempty"`
					B int32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int32 `json:"a,string"`
					B int32 `json:"b,string"`
				}
				B *struct {
					A int32 `json:"a,string"`
					B int32 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt32NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
				B *struct {
					A int32 `json:"a"`
					B int32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int32 `json:"a,omitempty"`
					B int32 `json:"b,omitempty"`
				}
				B *struct {
					A int32 `json:"a,omitempty"`
					B int32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int32 `json:"a,string"`
					B int32 `json:"b,string"`
				}
				B *struct {
					A int32 `json:"a,string"`
					B int32 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt32PtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32PtrDoubleMultiFieldsNotRoot",
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
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
				B *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
			}{A: &(struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(1), B: int32ptr(2)}), B: &(struct {
				A *int32 `json:"a"`
				B *int32 `json:"b"`
			}{A: int32ptr(3), B: int32ptr(4)})},
		},
		{
			name:     "PtrHeadInt32PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *int32 `json:"a,omitempty"`
					B *int32 `json:"b,omitempty"`
				}
				B *struct {
					A *int32 `json:"a,omitempty"`
					B *int32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int32 `json:"a,omitempty"`
				B *int32 `json:"b,omitempty"`
			}{A: int32ptr(1), B: int32ptr(2)}), B: &(struct {
				A *int32 `json:"a,omitempty"`
				B *int32 `json:"b,omitempty"`
			}{A: int32ptr(3), B: int32ptr(4)})},
		},
		{
			name:     "PtrHeadInt32PtrDoubleMultiFieldsNotRootString",
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
					A *int32 `json:"a,string"`
					B *int32 `json:"b,string"`
				}
				B *struct {
					A *int32 `json:"a,string"`
					B *int32 `json:"b,string"`
				}
			}{A: &(struct {
				A *int32 `json:"a,string"`
				B *int32 `json:"b,string"`
			}{A: int32ptr(1), B: int32ptr(2)}), B: &(struct {
				A *int32 `json:"a,string"`
				B *int32 `json:"b,string"`
			}{A: int32ptr(3), B: int32ptr(4)})},
		},

		// PtrHeadInt32PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
				B *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int32 `json:"a,omitempty"`
					B *int32 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *int32 `json:"a,omitempty"`
					B *int32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int32 `json:"a,string"`
					B *int32 `json:"b,string"`
				}
				B *struct {
					A *int32 `json:"a,string"`
					B *int32 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt32PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
				B *struct {
					A *int32 `json:"a"`
					B *int32 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a,omitempty"`
					B *int32 `json:"b,omitempty"`
				}
				B *struct {
					A *int32 `json:"a,omitempty"`
					B *int32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt32PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int32 `json:"a,string"`
					B *int32 `json:"b,string"`
				}
				B *struct {
					A *int32 `json:"a,string"`
					B *int32 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadInt32
		{
			name:     "AnonymousHeadInt32",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt32
				B int32 `json:"b"`
			}{
				structInt32: structInt32{A: 1},
				B:           2,
			},
		},
		{
			name:     "AnonymousHeadInt32OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt32OmitEmpty
				B int32 `json:"b,omitempty"`
			}{
				structInt32OmitEmpty: structInt32OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "AnonymousHeadInt32String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt32String
				B int32 `json:"b,string"`
			}{
				structInt32String: structInt32String{A: 1},
				B:                 2,
			},
		},

		// PtrAnonymousHeadInt32
		{
			name:     "PtrAnonymousHeadInt32",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt32
				B int32 `json:"b"`
			}{
				structInt32: &structInt32{A: 1},
				B:           2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt32OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt32OmitEmpty
				B int32 `json:"b,omitempty"`
			}{
				structInt32OmitEmpty: &structInt32OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt32String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt32String
				B int32 `json:"b,string"`
			}{
				structInt32String: &structInt32String{A: 1},
				B:                 2,
			},
		},

		// NilPtrAnonymousHeadInt32
		{
			name:     "NilPtrAnonymousHeadInt32",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt32
				B int32 `json:"b"`
			}{
				structInt32: nil,
				B:           2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt32OmitEmpty
				B int32 `json:"b,omitempty"`
			}{
				structInt32OmitEmpty: nil,
				B:                    2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt32String
				B int32 `json:"b,string"`
			}{
				structInt32String: nil,
				B:                 2,
			},
		},

		// AnonymousHeadInt32Ptr
		{
			name:     "AnonymousHeadInt32Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: structInt32Ptr{A: int32ptr(1)},
				B:              int32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt32PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt32PtrOmitEmpty
				B *int32 `json:"b,omitempty"`
			}{
				structInt32PtrOmitEmpty: structInt32PtrOmitEmpty{A: int32ptr(1)},
				B:                       int32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt32PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt32PtrString
				B *int32 `json:"b,string"`
			}{
				structInt32PtrString: structInt32PtrString{A: int32ptr(1)},
				B:                    int32ptr(2),
			},
		},

		// AnonymousHeadInt32PtrNil
		{
			name:     "AnonymousHeadInt32PtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: structInt32Ptr{A: nil},
				B:              int32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt32PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structInt32PtrOmitEmpty
				B *int32 `json:"b,omitempty"`
			}{
				structInt32PtrOmitEmpty: structInt32PtrOmitEmpty{A: nil},
				B:                       int32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt32PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structInt32PtrString
				B *int32 `json:"b,string"`
			}{
				structInt32PtrString: structInt32PtrString{A: nil},
				B:                    int32ptr(2),
			},
		},

		// PtrAnonymousHeadInt32Ptr
		{
			name:     "PtrAnonymousHeadInt32Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: &structInt32Ptr{A: int32ptr(1)},
				B:              int32ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt32PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt32PtrOmitEmpty
				B *int32 `json:"b,omitempty"`
			}{
				structInt32PtrOmitEmpty: &structInt32PtrOmitEmpty{A: int32ptr(1)},
				B:                       int32ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt32PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt32PtrString
				B *int32 `json:"b,string"`
			}{
				structInt32PtrString: &structInt32PtrString{A: int32ptr(1)},
				B:                    int32ptr(2),
			},
		},

		// NilPtrAnonymousHeadInt32Ptr
		{
			name:     "NilPtrAnonymousHeadInt32Ptr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt32Ptr
				B *int32 `json:"b"`
			}{
				structInt32Ptr: nil,
				B:              int32ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt32PtrOmitEmpty
				B *int32 `json:"b,omitempty"`
			}{
				structInt32PtrOmitEmpty: nil,
				B:                       int32ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt32PtrString
				B *int32 `json:"b,string"`
			}{
				structInt32PtrString: nil,
				B:                    int32ptr(2),
			},
		},

		// AnonymousHeadInt32Only
		{
			name:     "AnonymousHeadInt32Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt32
			}{
				structInt32: structInt32{A: 1},
			},
		},
		{
			name:     "AnonymousHeadInt32OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt32OmitEmpty
			}{
				structInt32OmitEmpty: structInt32OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadInt32OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt32String
			}{
				structInt32String: structInt32String{A: 1},
			},
		},

		// PtrAnonymousHeadInt32Only
		{
			name:     "PtrAnonymousHeadInt32Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt32
			}{
				structInt32: &structInt32{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt32OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt32OmitEmpty
			}{
				structInt32OmitEmpty: &structInt32OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt32OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt32String
			}{
				structInt32String: &structInt32String{A: 1},
			},
		},

		// NilPtrAnonymousHeadInt32Only
		{
			name:     "NilPtrAnonymousHeadInt32Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt32
			}{
				structInt32: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt32OmitEmpty
			}{
				structInt32OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt32String
			}{
				structInt32String: nil,
			},
		},

		// AnonymousHeadInt32PtrOnly
		{
			name:     "AnonymousHeadInt32PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt32Ptr
			}{
				structInt32Ptr: structInt32Ptr{A: int32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt32PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt32PtrOmitEmpty
			}{
				structInt32PtrOmitEmpty: structInt32PtrOmitEmpty{A: int32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt32PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt32PtrString
			}{
				structInt32PtrString: structInt32PtrString{A: int32ptr(1)},
			},
		},

		// AnonymousHeadInt32PtrNilOnly
		{
			name:     "AnonymousHeadInt32PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt32Ptr
			}{
				structInt32Ptr: structInt32Ptr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadInt32PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structInt32PtrOmitEmpty
			}{
				structInt32PtrOmitEmpty: structInt32PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadInt32PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt32PtrString
			}{
				structInt32PtrString: structInt32PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadInt32PtrOnly
		{
			name:     "PtrAnonymousHeadInt32PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt32Ptr
			}{
				structInt32Ptr: &structInt32Ptr{A: int32ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadInt32PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt32PtrOmitEmpty
			}{
				structInt32PtrOmitEmpty: &structInt32PtrOmitEmpty{A: int32ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadInt32PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt32PtrString
			}{
				structInt32PtrString: &structInt32PtrString{A: int32ptr(1)},
			},
		},

		// NilPtrAnonymousHeadInt32PtrOnly
		{
			name:     "NilPtrAnonymousHeadInt32PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt32Ptr
			}{
				structInt32Ptr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt32PtrOmitEmpty
			}{
				structInt32PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt32PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt32PtrString
			}{
				structInt32PtrString: nil,
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
