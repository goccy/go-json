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
	type structFloat32OmitEmpty struct {
		A float32 `json:"a,omitempty"`
	}
	type structFloat32String struct {
		A float32 `json:"a,string"`
	}

	type structFloat32Ptr struct {
		A *float32 `json:"a"`
	}
	type structFloat32PtrOmitEmpty struct {
		A *float32 `json:"a,omitempty"`
	}
	type structFloat32PtrString struct {
		A *float32 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadFloat32Zero
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
			name:     "HeadFloat32ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A float32 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadFloat32ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A float32 `json:"a,string"`
			}{},
		},

		// HeadFloat32
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
			name:     "HeadFloat32OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A float32 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadFloat32String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A float32 `json:"a,string"`
			}{A: 1},
		},

		// HeadFloat32Ptr
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
			name:     "HeadFloat32PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *float32 `json:"a,omitempty"`
			}{A: float32ptr(1)},
		},
		{
			name:     "HeadFloat32PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *float32 `json:"a,string"`
			}{A: float32ptr(1)},
		},

		// HeadFloat32PtrNil
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
			name:     "HeadFloat32PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *float32 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadFloat32PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *float32 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadFloat32Zero
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
			name:     "PtrHeadFloat32ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A float32 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadFloat32ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A float32 `json:"a,string"`
			}{},
		},

		// PtrHeadFloat32
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
			name:     "PtrHeadFloat32OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A float32 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadFloat32String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A float32 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadFloat32Ptr
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
			name:     "PtrHeadFloat32PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *float32 `json:"a,omitempty"`
			}{A: float32ptr(1)},
		},
		{
			name:     "PtrHeadFloat32PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *float32 `json:"a,string"`
			}{A: float32ptr(1)},
		},

		// PtrHeadFloat32PtrNil
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
			name:     "PtrHeadFloat32PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *float32 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat32PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *float32 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadFloat32Nil
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
			name:     "PtrHeadFloat32NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float32 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadFloat32NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float32 `json:"a,string"`
			})(nil),
		},

		// HeadFloat32ZeroMultiFields
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
			name:     "HeadFloat32ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A float32 `json:"a,omitempty"`
				B float32 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadFloat32ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A float32 `json:"a,string"`
				B float32 `json:"b,string"`
			}{},
		},

		// HeadFloat32MultiFields
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
			name:     "HeadFloat32MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A float32 `json:"a,omitempty"`
				B float32 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadFloat32MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A float32 `json:"a,string"`
				B float32 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadFloat32PtrMultiFields
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
			name:     "HeadFloat32PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *float32 `json:"a,omitempty"`
				B *float32 `json:"b,omitempty"`
			}{A: float32ptr(1), B: float32ptr(2)},
		},
		{
			name:     "HeadFloat32PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *float32 `json:"a,string"`
				B *float32 `json:"b,string"`
			}{A: float32ptr(1), B: float32ptr(2)},
		},

		// HeadFloat32PtrNilMultiFields
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
			name:     "HeadFloat32PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *float32 `json:"a,omitempty"`
				B *float32 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadFloat32PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *float32 `json:"a,string"`
				B *float32 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadFloat32ZeroMultiFields
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
			name:     "PtrHeadFloat32ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A float32 `json:"a,omitempty"`
				B float32 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadFloat32ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A float32 `json:"a,string"`
				B float32 `json:"b,string"`
			}{},
		},

		// PtrHeadFloat32MultiFields
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
			name:     "PtrHeadFloat32MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A float32 `json:"a,omitempty"`
				B float32 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadFloat32MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A float32 `json:"a,string"`
				B float32 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadFloat32PtrMultiFields
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
			name:     "PtrHeadFloat32PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *float32 `json:"a,omitempty"`
				B *float32 `json:"b,omitempty"`
			}{A: float32ptr(1), B: float32ptr(2)},
		},
		{
			name:     "PtrHeadFloat32PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *float32 `json:"a,string"`
				B *float32 `json:"b,string"`
			}{A: float32ptr(1), B: float32ptr(2)},
		},

		// PtrHeadFloat32PtrNilMultiFields
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
			name:     "PtrHeadFloat32PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *float32 `json:"a,omitempty"`
				B *float32 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *float32 `json:"a,string"`
				B *float32 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadFloat32NilMultiFields
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
			name:     "PtrHeadFloat32NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float32 `json:"a,omitempty"`
				B *float32 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadFloat32NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float32 `json:"a,string"`
				B *float32 `json:"b,string"`
			})(nil),
		},

		// HeadFloat32ZeroNotRoot
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
			name:     "HeadFloat32ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A float32 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadFloat32ZeroNotRootString",
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
					A float32 `json:"a,string"`
				}
			}{},
		},

		// HeadFloat32NotRoot
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
			name:     "HeadFloat32NotRootOmitEmpty",
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
					A float32 `json:"a,omitempty"`
				}
			}{A: struct {
				A float32 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadFloat32NotRootString",
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
					A float32 `json:"a,string"`
				}
			}{A: struct {
				A float32 `json:"a,string"`
			}{A: 1}},
		},

		// HeadFloat32PtrNotRoot
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
			name:     "HeadFloat32PtrNotRootOmitEmpty",
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
					A *float32 `json:"a,omitempty"`
				}
			}{A: struct {
				A *float32 `json:"a,omitempty"`
			}{float32ptr(1)}},
		},
		{
			name:     "HeadFloat32PtrNotRootString",
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
					A *float32 `json:"a,string"`
				}
			}{A: struct {
				A *float32 `json:"a,string"`
			}{float32ptr(1)}},
		},

		// HeadFloat32PtrNilNotRoot
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
			name:     "HeadFloat32PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *float32 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadFloat32PtrNilNotRootString",
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
					A *float32 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadFloat32ZeroNotRoot
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
			name:     "PtrHeadFloat32ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A float32 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A float32 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadFloat32ZeroNotRootString",
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
					A float32 `json:"a,string"`
				}
			}{A: new(struct {
				A float32 `json:"a,string"`
			})},
		},

		// PtrHeadFloat32NotRoot
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
			name:     "PtrHeadFloat32NotRootOmitEmpty",
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
					A float32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A float32 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadFloat32NotRootString",
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
					A float32 `json:"a,string"`
				}
			}{A: &(struct {
				A float32 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadFloat32PtrNotRoot
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
			name:     "PtrHeadFloat32PtrNotRootOmitEmpty",
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
					A *float32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *float32 `json:"a,omitempty"`
			}{A: float32ptr(1)})},
		},
		{
			name:     "PtrHeadFloat32PtrNotRootString",
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
					A *float32 `json:"a,string"`
				}
			}{A: &(struct {
				A *float32 `json:"a,string"`
			}{A: float32ptr(1)})},
		},

		// PtrHeadFloat32PtrNilNotRoot
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
			name:     "PtrHeadFloat32PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *float32 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *float32 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadFloat32PtrNilNotRootString",
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
					A *float32 `json:"a,string"`
				}
			}{A: &(struct {
				A *float32 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadFloat32NilNotRoot
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
			name:     "PtrHeadFloat32NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *float32 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat32NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *float32 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadFloat32ZeroMultiFieldsNotRoot
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
			name:     "HeadFloat32ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A float32 `json:"a,omitempty"`
				}
				B struct {
					B float32 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadFloat32ZeroMultiFieldsNotRootString",
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
					A float32 `json:"a,string"`
				}
				B struct {
					B float32 `json:"b,string"`
				}
			}{},
		},

		// HeadFloat32MultiFieldsNotRoot
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
			name:     "HeadFloat32MultiFieldsNotRootOmitEmpty",
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
					A float32 `json:"a,omitempty"`
				}
				B struct {
					B float32 `json:"b,omitempty"`
				}
			}{A: struct {
				A float32 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B float32 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadFloat32MultiFieldsNotRootString",
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
					A float32 `json:"a,string"`
				}
				B struct {
					B float32 `json:"b,string"`
				}
			}{A: struct {
				A float32 `json:"a,string"`
			}{A: 1}, B: struct {
				B float32 `json:"b,string"`
			}{B: 2}},
		},

		// HeadFloat32PtrMultiFieldsNotRoot
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
			name:     "HeadFloat32PtrMultiFieldsNotRootOmitEmpty",
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
					A *float32 `json:"a,omitempty"`
				}
				B struct {
					B *float32 `json:"b,omitempty"`
				}
			}{A: struct {
				A *float32 `json:"a,omitempty"`
			}{A: float32ptr(1)}, B: struct {
				B *float32 `json:"b,omitempty"`
			}{B: float32ptr(2)}},
		},
		{
			name:     "HeadFloat32PtrMultiFieldsNotRootString",
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
					A *float32 `json:"a,string"`
				}
				B struct {
					B *float32 `json:"b,string"`
				}
			}{A: struct {
				A *float32 `json:"a,string"`
			}{A: float32ptr(1)}, B: struct {
				B *float32 `json:"b,string"`
			}{B: float32ptr(2)}},
		},

		// HeadFloat32PtrNilMultiFieldsNotRoot
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
			name:     "HeadFloat32PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *float32 `json:"a,omitempty"`
				}
				B struct {
					B *float32 `json:"b,omitempty"`
				}
			}{A: struct {
				A *float32 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *float32 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadFloat32PtrNilMultiFieldsNotRootString",
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
					A *float32 `json:"a,string"`
				}
				B struct {
					B *float32 `json:"b,string"`
				}
			}{A: struct {
				A *float32 `json:"a,string"`
			}{A: nil}, B: struct {
				B *float32 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadFloat32ZeroMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A float32 `json:"a,omitempty"`
				}
				B struct {
					B float32 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadFloat32ZeroMultiFieldsNotRootString",
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
					A float32 `json:"a,string"`
				}
				B struct {
					B float32 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadFloat32MultiFieldsNotRoot
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
			name:     "PtrHeadFloat32MultiFieldsNotRootOmitEmpty",
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
					A float32 `json:"a,omitempty"`
				}
				B struct {
					B float32 `json:"b,omitempty"`
				}
			}{A: struct {
				A float32 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B float32 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadFloat32MultiFieldsNotRootString",
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
					A float32 `json:"a,string"`
				}
				B struct {
					B float32 `json:"b,string"`
				}
			}{A: struct {
				A float32 `json:"a,string"`
			}{A: 1}, B: struct {
				B float32 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadFloat32PtrMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32PtrMultiFieldsNotRootOmitEmpty",
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
					A *float32 `json:"a,omitempty"`
				}
				B *struct {
					B *float32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *float32 `json:"a,omitempty"`
			}{A: float32ptr(1)}), B: &(struct {
				B *float32 `json:"b,omitempty"`
			}{B: float32ptr(2)})},
		},
		{
			name:     "PtrHeadFloat32PtrMultiFieldsNotRootString",
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
					A *float32 `json:"a,string"`
				}
				B *struct {
					B *float32 `json:"b,string"`
				}
			}{A: &(struct {
				A *float32 `json:"a,string"`
			}{A: float32ptr(1)}), B: &(struct {
				B *float32 `json:"b,string"`
			}{B: float32ptr(2)})},
		},

		// PtrHeadFloat32PtrNilMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *float32 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *float32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float32 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *float32 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadFloat32NilMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float32 `json:"a,omitempty"`
				}
				B *struct {
					B *float32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat32NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float32 `json:"a,string"`
				}
				B *struct {
					B *float32 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadFloat32DoubleMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32DoubleMultiFieldsNotRootOmitEmpty",
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
					A float32 `json:"a,omitempty"`
					B float32 `json:"b,omitempty"`
				}
				B *struct {
					A float32 `json:"a,omitempty"`
					B float32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A float32 `json:"a,omitempty"`
				B float32 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A float32 `json:"a,omitempty"`
				B float32 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadFloat32DoubleMultiFieldsNotRootString",
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
					A float32 `json:"a,string"`
					B float32 `json:"b,string"`
				}
				B *struct {
					A float32 `json:"a,string"`
					B float32 `json:"b,string"`
				}
			}{A: &(struct {
				A float32 `json:"a,string"`
				B float32 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A float32 `json:"a,string"`
				B float32 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadFloat32NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A float32 `json:"a,omitempty"`
					B float32 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A float32 `json:"a,omitempty"`
					B float32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A float32 `json:"a,string"`
					B float32 `json:"b,string"`
				}
				B *struct {
					A float32 `json:"a,string"`
					B float32 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadFloat32NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A float32 `json:"a,omitempty"`
					B float32 `json:"b,omitempty"`
				}
				B *struct {
					A float32 `json:"a,omitempty"`
					B float32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat32NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A float32 `json:"a,string"`
					B float32 `json:"b,string"`
				}
				B *struct {
					A float32 `json:"a,string"`
					B float32 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadFloat32PtrDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *float32 `json:"a,omitempty"`
					B *float32 `json:"b,omitempty"`
				}
				B *struct {
					A *float32 `json:"a,omitempty"`
					B *float32 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *float32 `json:"a,omitempty"`
				B *float32 `json:"b,omitempty"`
			}{A: float32ptr(1), B: float32ptr(2)}), B: &(struct {
				A *float32 `json:"a,omitempty"`
				B *float32 `json:"b,omitempty"`
			}{A: float32ptr(3), B: float32ptr(4)})},
		},
		{
			name:     "PtrHeadFloat32PtrDoubleMultiFieldsNotRootString",
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
					A *float32 `json:"a,string"`
					B *float32 `json:"b,string"`
				}
				B *struct {
					A *float32 `json:"a,string"`
					B *float32 `json:"b,string"`
				}
			}{A: &(struct {
				A *float32 `json:"a,string"`
				B *float32 `json:"b,string"`
			}{A: float32ptr(1), B: float32ptr(2)}), B: &(struct {
				A *float32 `json:"a,string"`
				B *float32 `json:"b,string"`
			}{A: float32ptr(3), B: float32ptr(4)})},
		},

		// PtrHeadFloat32PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *float32 `json:"a,omitempty"`
					B *float32 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *float32 `json:"a,omitempty"`
					B *float32 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat32PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float32 `json:"a,string"`
					B *float32 `json:"b,string"`
				}
				B *struct {
					A *float32 `json:"a,string"`
					B *float32 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadFloat32PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadFloat32PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float32 `json:"a,omitempty"`
					B *float32 `json:"b,omitempty"`
				}
				B *struct {
					A *float32 `json:"a,omitempty"`
					B *float32 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat32PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float32 `json:"a,string"`
					B *float32 `json:"b,string"`
				}
				B *struct {
					A *float32 `json:"a,string"`
					B *float32 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadFloat32
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
			name:     "AnonymousHeadFloat32OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat32OmitEmpty
				B float32 `json:"b,omitempty"`
			}{
				structFloat32OmitEmpty: structFloat32OmitEmpty{A: 1},
				B:                      2,
			},
		},
		{
			name:     "AnonymousHeadFloat32String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structFloat32String
				B float32 `json:"b,string"`
			}{
				structFloat32String: structFloat32String{A: 1},
				B:                   2,
			},
		},

		// PtrAnonymousHeadFloat32
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
			name:     "PtrAnonymousHeadFloat32OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat32OmitEmpty
				B float32 `json:"b,omitempty"`
			}{
				structFloat32OmitEmpty: &structFloat32OmitEmpty{A: 1},
				B:                      2,
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structFloat32String
				B float32 `json:"b,string"`
			}{
				structFloat32String: &structFloat32String{A: 1},
				B:                   2,
			},
		},

		// NilPtrAnonymousHeadFloat32
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
			name:     "NilPtrAnonymousHeadFloat32OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat32OmitEmpty
				B float32 `json:"b,omitempty"`
			}{
				structFloat32OmitEmpty: nil,
				B:                      2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structFloat32String
				B float32 `json:"b,string"`
			}{
				structFloat32String: nil,
				B:                   2,
			},
		},

		// AnonymousHeadFloat32Ptr
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
			name:     "AnonymousHeadFloat32PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat32PtrOmitEmpty
				B *float32 `json:"b,omitempty"`
			}{
				structFloat32PtrOmitEmpty: structFloat32PtrOmitEmpty{A: float32ptr(1)},
				B:                         float32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat32PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structFloat32PtrString
				B *float32 `json:"b,string"`
			}{
				structFloat32PtrString: structFloat32PtrString{A: float32ptr(1)},
				B:                      float32ptr(2),
			},
		},

		// AnonymousHeadFloat32PtrNil
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
			name:     "AnonymousHeadFloat32PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structFloat32PtrOmitEmpty
				B *float32 `json:"b,omitempty"`
			}{
				structFloat32PtrOmitEmpty: structFloat32PtrOmitEmpty{A: nil},
				B:                         float32ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat32PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structFloat32PtrString
				B *float32 `json:"b,string"`
			}{
				structFloat32PtrString: structFloat32PtrString{A: nil},
				B:                      float32ptr(2),
			},
		},

		// PtrAnonymousHeadFloat32Ptr
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
			name:     "PtrAnonymousHeadFloat32PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat32PtrOmitEmpty
				B *float32 `json:"b,omitempty"`
			}{
				structFloat32PtrOmitEmpty: &structFloat32PtrOmitEmpty{A: float32ptr(1)},
				B:                         float32ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structFloat32PtrString
				B *float32 `json:"b,string"`
			}{
				structFloat32PtrString: &structFloat32PtrString{A: float32ptr(1)},
				B:                      float32ptr(2),
			},
		},

		// NilPtrAnonymousHeadFloat32Ptr
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
			name:     "NilPtrAnonymousHeadFloat32PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat32PtrOmitEmpty
				B *float32 `json:"b,omitempty"`
			}{
				structFloat32PtrOmitEmpty: nil,
				B:                         float32ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structFloat32PtrString
				B *float32 `json:"b,string"`
			}{
				structFloat32PtrString: nil,
				B:                      float32ptr(2),
			},
		},

		// AnonymousHeadFloat32Only
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
			name:     "AnonymousHeadFloat32OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat32OmitEmpty
			}{
				structFloat32OmitEmpty: structFloat32OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadFloat32OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structFloat32String
			}{
				structFloat32String: structFloat32String{A: 1},
			},
		},

		// PtrAnonymousHeadFloat32Only
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
			name:     "PtrAnonymousHeadFloat32OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat32OmitEmpty
			}{
				structFloat32OmitEmpty: &structFloat32OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structFloat32String
			}{
				structFloat32String: &structFloat32String{A: 1},
			},
		},

		// NilPtrAnonymousHeadFloat32Only
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
			name:     "NilPtrAnonymousHeadFloat32OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat32OmitEmpty
			}{
				structFloat32OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat32String
			}{
				structFloat32String: nil,
			},
		},

		// AnonymousHeadFloat32PtrOnly
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
			name:     "AnonymousHeadFloat32PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat32PtrOmitEmpty
			}{
				structFloat32PtrOmitEmpty: structFloat32PtrOmitEmpty{A: float32ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadFloat32PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structFloat32PtrString
			}{
				structFloat32PtrString: structFloat32PtrString{A: float32ptr(1)},
			},
		},

		// AnonymousHeadFloat32PtrNilOnly
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
			name:     "AnonymousHeadFloat32PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structFloat32PtrOmitEmpty
			}{
				structFloat32PtrOmitEmpty: structFloat32PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadFloat32PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structFloat32PtrString
			}{
				structFloat32PtrString: structFloat32PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadFloat32PtrOnly
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
			name:     "PtrAnonymousHeadFloat32PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat32PtrOmitEmpty
			}{
				structFloat32PtrOmitEmpty: &structFloat32PtrOmitEmpty{A: float32ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat32PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structFloat32PtrString
			}{
				structFloat32PtrString: &structFloat32PtrString{A: float32ptr(1)},
			},
		},

		// NilPtrAnonymousHeadFloat32PtrOnly
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
		{
			name:     "NilPtrAnonymousHeadFloat32PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat32PtrOmitEmpty
			}{
				structFloat32PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat32PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat32PtrString
			}{
				structFloat32PtrString: nil,
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
