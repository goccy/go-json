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
	type structUint32OmitEmpty struct {
		A uint32 `json:"a,omitempty"`
	}
	type structUint32String struct {
		A uint32 `json:"a,string"`
	}

	type structUint32Ptr struct {
		A *uint32 `json:"a"`
	}
	type structUint32PtrOmitEmpty struct {
		A *uint32 `json:"a,omitempty"`
	}
	type structUint32PtrString struct {
		A *uint32 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadUint32Zero
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
			name:     "HeadUint32ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint32 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadUint32ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A uint32 `json:"a,string"`
			}{},
		},

		// HeadUint32
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
			name:     "HeadUint32OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint32 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadUint32String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A uint32 `json:"a,string"`
			}{A: 1},
		},

		// HeadUint32Ptr
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
			name:     "HeadUint32PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint32 `json:"a,omitempty"`
			}{A: uint32ptr(1)},
		},
		{
			name:     "HeadUint32PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *uint32 `json:"a,string"`
			}{A: uint32ptr(1)},
		},

		// HeadUint32PtrNil
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
			name:     "HeadUint32PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint32 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadUint32PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint32 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint32Zero
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
			name:     "PtrHeadUint32ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint32 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint32ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A uint32 `json:"a,string"`
			}{},
		},

		// PtrHeadUint32
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
			name:     "PtrHeadUint32OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint32 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint32String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A uint32 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadUint32Ptr
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
			name:     "PtrHeadUint32PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint32 `json:"a,omitempty"`
			}{A: uint32ptr(1)},
		},
		{
			name:     "PtrHeadUint32PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *uint32 `json:"a,string"`
			}{A: uint32ptr(1)},
		},

		// PtrHeadUint32PtrNil
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
			name:     "PtrHeadUint32PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint32 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint32PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint32 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint32Nil
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
			name:     "PtrHeadUint32NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint32 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint32NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint32 `json:"a,string"`
			})(nil),
		},

		// HeadUint32ZeroMultiFields
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
			name:     "HeadUint32ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint32 `json:"a,omitempty"`
				B uint32 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadUint32ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A uint32 `json:"a,string"`
				B uint32 `json:"b,string"`
			}{},
		},

		// HeadUint32MultiFields
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
			name:     "HeadUint32MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint32 `json:"a,omitempty"`
				B uint32 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint32MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A uint32 `json:"a,string"`
				B uint32 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadUint32PtrMultiFields
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
			name:     "HeadUint32PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint32 `json:"a,omitempty"`
				B *uint32 `json:"b,omitempty"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},
		{
			name:     "HeadUint32PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *uint32 `json:"a,string"`
				B *uint32 `json:"b,string"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},

		// HeadUint32PtrNilMultiFields
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
			name:     "HeadUint32PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint32 `json:"a,omitempty"`
				B *uint32 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadUint32PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint32 `json:"a,string"`
				B *uint32 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint32ZeroMultiFields
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
			name:     "PtrHeadUint32ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint32 `json:"a,omitempty"`
				B uint32 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint32ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A uint32 `json:"a,string"`
				B uint32 `json:"b,string"`
			}{},
		},

		// PtrHeadUint32MultiFields
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
			name:     "PtrHeadUint32MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint32 `json:"a,omitempty"`
				B uint32 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint32MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A uint32 `json:"a,string"`
				B uint32 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadUint32PtrMultiFields
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
			name:     "PtrHeadUint32PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint32 `json:"a,omitempty"`
				B *uint32 `json:"b,omitempty"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},
		{
			name:     "PtrHeadUint32PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *uint32 `json:"a,string"`
				B *uint32 `json:"b,string"`
			}{A: uint32ptr(1), B: uint32ptr(2)},
		},

		// PtrHeadUint32PtrNilMultiFields
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
			name:     "PtrHeadUint32PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint32 `json:"a,omitempty"`
				B *uint32 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint32 `json:"a,string"`
				B *uint32 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint32NilMultiFields
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
			name:     "PtrHeadUint32NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint32 `json:"a,omitempty"`
				B *uint32 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint32NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint32 `json:"a,string"`
				B *uint32 `json:"b,string"`
			})(nil),
		},

		// HeadUint32ZeroNotRoot
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
			name:     "HeadUint32ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A uint32 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint32ZeroNotRootString",
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
					A uint32 `json:"a,string"`
				}
			}{},
		},

		// HeadUint32NotRoot
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
			name:     "HeadUint32NotRootOmitEmpty",
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
					A uint32 `json:"a,omitempty"`
				}
			}{A: struct {
				A uint32 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadUint32NotRootString",
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
					A uint32 `json:"a,string"`
				}
			}{A: struct {
				A uint32 `json:"a,string"`
			}{A: 1}},
		},

		// HeadUint32PtrNotRoot
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
			name:     "HeadUint32PtrNotRootOmitEmpty",
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
					A *uint32 `json:"a,omitempty"`
				}
			}{A: struct {
				A *uint32 `json:"a,omitempty"`
			}{uint32ptr(1)}},
		},
		{
			name:     "HeadUint32PtrNotRootString",
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
					A *uint32 `json:"a,string"`
				}
			}{A: struct {
				A *uint32 `json:"a,string"`
			}{uint32ptr(1)}},
		},

		// HeadUint32PtrNilNotRoot
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
			name:     "HeadUint32PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *uint32 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint32PtrNilNotRootString",
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
					A *uint32 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadUint32ZeroNotRoot
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
			name:     "PtrHeadUint32ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A uint32 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A uint32 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadUint32ZeroNotRootString",
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
					A uint32 `json:"a,string"`
				}
			}{A: new(struct {
				A uint32 `json:"a,string"`
			})},
		},

		// PtrHeadUint32NotRoot
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
			name:     "PtrHeadUint32NotRootOmitEmpty",
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
					A uint32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A uint32 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint32NotRootString",
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
					A uint32 `json:"a,string"`
				}
			}{A: &(struct {
				A uint32 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadUint32PtrNotRoot
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
			name:     "PtrHeadUint32PtrNotRootOmitEmpty",
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
					A *uint32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,omitempty"`
			}{A: uint32ptr(1)})},
		},
		{
			name:     "PtrHeadUint32PtrNotRootString",
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
					A *uint32 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,string"`
			}{A: uint32ptr(1)})},
		},

		// PtrHeadUint32PtrNilNotRoot
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
			name:     "PtrHeadUint32PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *uint32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint32PtrNilNotRootString",
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
					A *uint32 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadUint32NilNotRoot
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
			name:     "PtrHeadUint32NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *uint32 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint32NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint32 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadUint32ZeroMultiFieldsNotRoot
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
			name:     "HeadUint32ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A uint32 `json:"a,omitempty"`
				}
				B struct {
					B uint32 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint32ZeroMultiFieldsNotRootString",
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
					A uint32 `json:"a,string"`
				}
				B struct {
					B uint32 `json:"b,string"`
				}
			}{},
		},

		// HeadUint32MultiFieldsNotRoot
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
			name:     "HeadUint32MultiFieldsNotRootOmitEmpty",
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
					A uint32 `json:"a,omitempty"`
				}
				B struct {
					B uint32 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint32 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint32 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadUint32MultiFieldsNotRootString",
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
					A uint32 `json:"a,string"`
				}
				B struct {
					B uint32 `json:"b,string"`
				}
			}{A: struct {
				A uint32 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint32 `json:"b,string"`
			}{B: 2}},
		},

		// HeadUint32PtrMultiFieldsNotRoot
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
			name:     "HeadUint32PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint32 `json:"a,omitempty"`
				}
				B struct {
					B *uint32 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint32 `json:"a,omitempty"`
			}{A: uint32ptr(1)}, B: struct {
				B *uint32 `json:"b,omitempty"`
			}{B: uint32ptr(2)}},
		},
		{
			name:     "HeadUint32PtrMultiFieldsNotRootString",
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
					A *uint32 `json:"a,string"`
				}
				B struct {
					B *uint32 `json:"b,string"`
				}
			}{A: struct {
				A *uint32 `json:"a,string"`
			}{A: uint32ptr(1)}, B: struct {
				B *uint32 `json:"b,string"`
			}{B: uint32ptr(2)}},
		},

		// HeadUint32PtrNilMultiFieldsNotRoot
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
			name:     "HeadUint32PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *uint32 `json:"a,omitempty"`
				}
				B struct {
					B *uint32 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint32 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *uint32 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadUint32PtrNilMultiFieldsNotRootString",
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
					A *uint32 `json:"a,string"`
				}
				B struct {
					B *uint32 `json:"b,string"`
				}
			}{A: struct {
				A *uint32 `json:"a,string"`
			}{A: nil}, B: struct {
				B *uint32 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadUint32ZeroMultiFieldsNotRoot
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
			name:     "PtrHeadUint32ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A uint32 `json:"a,omitempty"`
				}
				B struct {
					B uint32 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint32ZeroMultiFieldsNotRootString",
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
					A uint32 `json:"a,string"`
				}
				B struct {
					B uint32 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadUint32MultiFieldsNotRoot
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
			name:     "PtrHeadUint32MultiFieldsNotRootOmitEmpty",
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
					A uint32 `json:"a,omitempty"`
				}
				B struct {
					B uint32 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint32 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint32 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint32MultiFieldsNotRootString",
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
					A uint32 `json:"a,string"`
				}
				B struct {
					B uint32 `json:"b,string"`
				}
			}{A: struct {
				A uint32 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint32 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadUint32PtrMultiFieldsNotRoot
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
			name:     "PtrHeadUint32PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint32 `json:"a,omitempty"`
				}
				B *struct {
					B *uint32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,omitempty"`
			}{A: uint32ptr(1)}), B: &(struct {
				B *uint32 `json:"b,omitempty"`
			}{B: uint32ptr(2)})},
		},
		{
			name:     "PtrHeadUint32PtrMultiFieldsNotRootString",
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
					A *uint32 `json:"a,string"`
				}
				B *struct {
					B *uint32 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,string"`
			}{A: uint32ptr(1)}), B: &(struct {
				B *uint32 `json:"b,string"`
			}{B: uint32ptr(2)})},
		},

		// PtrHeadUint32PtrNilMultiFieldsNotRoot
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
			name:     "PtrHeadUint32PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *uint32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *uint32 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint32NilMultiFieldsNotRoot
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
			name:     "PtrHeadUint32NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint32 `json:"a,omitempty"`
				}
				B *struct {
					B *uint32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint32NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint32 `json:"a,string"`
				}
				B *struct {
					B *uint32 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint32DoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint32DoubleMultiFieldsNotRootOmitEmpty",
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
					A uint32 `json:"a,omitempty"`
					B uint32 `json:"b,omitempty"`
				}
				B *struct {
					A uint32 `json:"a,omitempty"`
					B uint32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A uint32 `json:"a,omitempty"`
				B uint32 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A uint32 `json:"a,omitempty"`
				B uint32 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint32DoubleMultiFieldsNotRootString",
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
					A uint32 `json:"a,string"`
					B uint32 `json:"b,string"`
				}
				B *struct {
					A uint32 `json:"a,string"`
					B uint32 `json:"b,string"`
				}
			}{A: &(struct {
				A uint32 `json:"a,string"`
				B uint32 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A uint32 `json:"a,string"`
				B uint32 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadUint32NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint32NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A uint32 `json:"a,omitempty"`
					B uint32 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A uint32 `json:"a,omitempty"`
					B uint32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A uint32 `json:"a,string"`
					B uint32 `json:"b,string"`
				}
				B *struct {
					A uint32 `json:"a,string"`
					B uint32 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint32NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint32NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint32 `json:"a,omitempty"`
					B uint32 `json:"b,omitempty"`
				}
				B *struct {
					A uint32 `json:"a,omitempty"`
					B uint32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint32NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint32 `json:"a,string"`
					B uint32 `json:"b,string"`
				}
				B *struct {
					A uint32 `json:"a,string"`
					B uint32 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint32PtrDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint32PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *uint32 `json:"a,omitempty"`
					B *uint32 `json:"b,omitempty"`
				}
				B *struct {
					A *uint32 `json:"a,omitempty"`
					B *uint32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,omitempty"`
				B *uint32 `json:"b,omitempty"`
			}{A: uint32ptr(1), B: uint32ptr(2)}), B: &(struct {
				A *uint32 `json:"a,omitempty"`
				B *uint32 `json:"b,omitempty"`
			}{A: uint32ptr(3), B: uint32ptr(4)})},
		},
		{
			name:     "PtrHeadUint32PtrDoubleMultiFieldsNotRootString",
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
					A *uint32 `json:"a,string"`
					B *uint32 `json:"b,string"`
				}
				B *struct {
					A *uint32 `json:"a,string"`
					B *uint32 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint32 `json:"a,string"`
				B *uint32 `json:"b,string"`
			}{A: uint32ptr(1), B: uint32ptr(2)}), B: &(struct {
				A *uint32 `json:"a,string"`
				B *uint32 `json:"b,string"`
			}{A: uint32ptr(3), B: uint32ptr(4)})},
		},

		// PtrHeadUint32PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint32PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a,omitempty"`
					B *uint32 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *uint32 `json:"a,omitempty"`
					B *uint32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint32PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint32 `json:"a,string"`
					B *uint32 `json:"b,string"`
				}
				B *struct {
					A *uint32 `json:"a,string"`
					B *uint32 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint32PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint32PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint32 `json:"a,omitempty"`
					B *uint32 `json:"b,omitempty"`
				}
				B *struct {
					A *uint32 `json:"a,omitempty"`
					B *uint32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint32PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint32 `json:"a,string"`
					B *uint32 `json:"b,string"`
				}
				B *struct {
					A *uint32 `json:"a,string"`
					B *uint32 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadUint32
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
			name:     "AnonymousHeadUint32OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint32OmitEmpty
				B uint32 `json:"b,omitempty"`
			}{
				structUint32OmitEmpty: structUint32OmitEmpty{A: 1},
				B:                     2,
			},
		},
		{
			name:     "AnonymousHeadUint32String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint32String
				B uint32 `json:"b,string"`
			}{
				structUint32String: structUint32String{A: 1},
				B:                  2,
			},
		},

		// PtrAnonymousHeadUint32
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
			name:     "PtrAnonymousHeadUint32OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint32OmitEmpty
				B uint32 `json:"b,omitempty"`
			}{
				structUint32OmitEmpty: &structUint32OmitEmpty{A: 1},
				B:                     2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint32String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint32String
				B uint32 `json:"b,string"`
			}{
				structUint32String: &structUint32String{A: 1},
				B:                  2,
			},
		},

		// NilPtrAnonymousHeadUint32
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
			name:     "NilPtrAnonymousHeadUint32OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint32OmitEmpty
				B uint32 `json:"b,omitempty"`
			}{
				structUint32OmitEmpty: nil,
				B:                     2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint32String
				B uint32 `json:"b,string"`
			}{
				structUint32String: nil,
				B:                  2,
			},
		},

		// AnonymousHeadUint32Ptr
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
			name:     "AnonymousHeadUint32PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint32PtrOmitEmpty
				B *uint32 `json:"b,omitempty"`
			}{
				structUint32PtrOmitEmpty: structUint32PtrOmitEmpty{A: uint32ptr(1)},
				B:                        uint32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint32PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint32PtrString
				B *uint32 `json:"b,string"`
			}{
				structUint32PtrString: structUint32PtrString{A: uint32ptr(1)},
				B:                     uint32ptr(2),
			},
		},

		// AnonymousHeadUint32PtrNil
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
			name:     "AnonymousHeadUint32PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structUint32PtrOmitEmpty
				B *uint32 `json:"b,omitempty"`
			}{
				structUint32PtrOmitEmpty: structUint32PtrOmitEmpty{A: nil},
				B:                        uint32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint32PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structUint32PtrString
				B *uint32 `json:"b,string"`
			}{
				structUint32PtrString: structUint32PtrString{A: nil},
				B:                     uint32ptr(2),
			},
		},

		// PtrAnonymousHeadUint32Ptr
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
			name:     "PtrAnonymousHeadUint32PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint32PtrOmitEmpty
				B *uint32 `json:"b,omitempty"`
			}{
				structUint32PtrOmitEmpty: &structUint32PtrOmitEmpty{A: uint32ptr(1)},
				B:                        uint32ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint32PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint32PtrString
				B *uint32 `json:"b,string"`
			}{
				structUint32PtrString: &structUint32PtrString{A: uint32ptr(1)},
				B:                     uint32ptr(2),
			},
		},

		// NilPtrAnonymousHeadUint32Ptr
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
			name:     "NilPtrAnonymousHeadUint32PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint32PtrOmitEmpty
				B *uint32 `json:"b,omitempty"`
			}{
				structUint32PtrOmitEmpty: nil,
				B:                        uint32ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint32PtrString
				B *uint32 `json:"b,string"`
			}{
				structUint32PtrString: nil,
				B:                     uint32ptr(2),
			},
		},

		// AnonymousHeadUint32Only
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
			name:     "AnonymousHeadUint32OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint32OmitEmpty
			}{
				structUint32OmitEmpty: structUint32OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadUint32OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint32String
			}{
				structUint32String: structUint32String{A: 1},
			},
		},

		// PtrAnonymousHeadUint32Only
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
			name:     "PtrAnonymousHeadUint32OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint32OmitEmpty
			}{
				structUint32OmitEmpty: &structUint32OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint32OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint32String
			}{
				structUint32String: &structUint32String{A: 1},
			},
		},

		// NilPtrAnonymousHeadUint32Only
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
			name:     "NilPtrAnonymousHeadUint32OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint32OmitEmpty
			}{
				structUint32OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint32String
			}{
				structUint32String: nil,
			},
		},

		// AnonymousHeadUint32PtrOnly
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
			name:     "AnonymousHeadUint32PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint32PtrOmitEmpty
			}{
				structUint32PtrOmitEmpty: structUint32PtrOmitEmpty{A: uint32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint32PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint32PtrString
			}{
				structUint32PtrString: structUint32PtrString{A: uint32ptr(1)},
			},
		},

		// AnonymousHeadUint32PtrNilOnly
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
			name:     "AnonymousHeadUint32PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structUint32PtrOmitEmpty
			}{
				structUint32PtrOmitEmpty: structUint32PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadUint32PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint32PtrString
			}{
				structUint32PtrString: structUint32PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadUint32PtrOnly
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
			name:     "PtrAnonymousHeadUint32PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint32PtrOmitEmpty
			}{
				structUint32PtrOmitEmpty: &structUint32PtrOmitEmpty{A: uint32ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadUint32PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint32PtrString
			}{
				structUint32PtrString: &structUint32PtrString{A: uint32ptr(1)},
			},
		},

		// NilPtrAnonymousHeadUint32PtrOnly
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
		{
			name:     "NilPtrAnonymousHeadUint32PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint32PtrOmitEmpty
			}{
				structUint32PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint32PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint32PtrString
			}{
				structUint32PtrString: nil,
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
