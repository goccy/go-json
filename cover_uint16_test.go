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
	type structUint16OmitEmpty struct {
		A uint16 `json:"a,omitempty"`
	}
	type structUint16String struct {
		A uint16 `json:"a,string"`
	}

	type structUint16Ptr struct {
		A *uint16 `json:"a"`
	}
	type structUint16PtrOmitEmpty struct {
		A *uint16 `json:"a,omitempty"`
	}
	type structUint16PtrString struct {
		A *uint16 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadUint16Zero
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
			name:     "HeadUint16ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint16 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadUint16ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A uint16 `json:"a,string"`
			}{},
		},

		// HeadUint16
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
			name:     "HeadUint16OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint16 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadUint16String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A uint16 `json:"a,string"`
			}{A: 1},
		},

		// HeadUint16Ptr
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
			name:     "HeadUint16PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint16 `json:"a,omitempty"`
			}{A: uint16ptr(1)},
		},
		{
			name:     "HeadUint16PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *uint16 `json:"a,string"`
			}{A: uint16ptr(1)},
		},

		// HeadUint16PtrNil
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
			name:     "HeadUint16PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint16 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadUint16PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint16 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint16Zero
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
			name:     "PtrHeadUint16ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint16 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint16ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A uint16 `json:"a,string"`
			}{},
		},

		// PtrHeadUint16
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
			name:     "PtrHeadUint16OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint16 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint16String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A uint16 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadUint16Ptr
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
			name:     "PtrHeadUint16PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint16 `json:"a,omitempty"`
			}{A: uint16ptr(1)},
		},
		{
			name:     "PtrHeadUint16PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *uint16 `json:"a,string"`
			}{A: uint16ptr(1)},
		},

		// PtrHeadUint16PtrNil
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
			name:     "PtrHeadUint16PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint16 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint16PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint16 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint16Nil
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
			name:     "PtrHeadUint16NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint16 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint16NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint16 `json:"a,string"`
			})(nil),
		},

		// HeadUint16ZeroMultiFields
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
			name:     "HeadUint16ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint16 `json:"a,omitempty"`
				B uint16 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadUint16ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A uint16 `json:"a,string"`
				B uint16 `json:"b,string"`
			}{},
		},

		// HeadUint16MultiFields
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
			name:     "HeadUint16MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint16 `json:"a,omitempty"`
				B uint16 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint16MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A uint16 `json:"a,string"`
				B uint16 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadUint16PtrMultiFields
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
			name:     "HeadUint16PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint16 `json:"a,omitempty"`
				B *uint16 `json:"b,omitempty"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},
		{
			name:     "HeadUint16PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *uint16 `json:"a,string"`
				B *uint16 `json:"b,string"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},

		// HeadUint16PtrNilMultiFields
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
			name:     "HeadUint16PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint16 `json:"a,omitempty"`
				B *uint16 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadUint16PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint16 `json:"a,string"`
				B *uint16 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint16ZeroMultiFields
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
			name:     "PtrHeadUint16ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint16 `json:"a,omitempty"`
				B uint16 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint16ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A uint16 `json:"a,string"`
				B uint16 `json:"b,string"`
			}{},
		},

		// PtrHeadUint16MultiFields
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
			name:     "PtrHeadUint16MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint16 `json:"a,omitempty"`
				B uint16 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint16MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A uint16 `json:"a,string"`
				B uint16 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadUint16PtrMultiFields
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
			name:     "PtrHeadUint16PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint16 `json:"a,omitempty"`
				B *uint16 `json:"b,omitempty"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},
		{
			name:     "PtrHeadUint16PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *uint16 `json:"a,string"`
				B *uint16 `json:"b,string"`
			}{A: uint16ptr(1), B: uint16ptr(2)},
		},

		// PtrHeadUint16PtrNilMultiFields
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
			name:     "PtrHeadUint16PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint16 `json:"a,omitempty"`
				B *uint16 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint16 `json:"a,string"`
				B *uint16 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint16NilMultiFields
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
			name:     "PtrHeadUint16NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint16 `json:"a,omitempty"`
				B *uint16 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint16NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint16 `json:"a,string"`
				B *uint16 `json:"b,string"`
			})(nil),
		},

		// HeadUint16ZeroNotRoot
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
			name:     "HeadUint16ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A uint16 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint16ZeroNotRootString",
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
					A uint16 `json:"a,string"`
				}
			}{},
		},

		// HeadUint16NotRoot
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
			name:     "HeadUint16NotRootOmitEmpty",
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
					A uint16 `json:"a,omitempty"`
				}
			}{A: struct {
				A uint16 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadUint16NotRootString",
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
					A uint16 `json:"a,string"`
				}
			}{A: struct {
				A uint16 `json:"a,string"`
			}{A: 1}},
		},

		// HeadUint16PtrNotRoot
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
			name:     "HeadUint16PtrNotRootOmitEmpty",
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
					A *uint16 `json:"a,omitempty"`
				}
			}{A: struct {
				A *uint16 `json:"a,omitempty"`
			}{uint16ptr(1)}},
		},
		{
			name:     "HeadUint16PtrNotRootString",
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
					A *uint16 `json:"a,string"`
				}
			}{A: struct {
				A *uint16 `json:"a,string"`
			}{uint16ptr(1)}},
		},

		// HeadUint16PtrNilNotRoot
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
			name:     "HeadUint16PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *uint16 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint16PtrNilNotRootString",
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
					A *uint16 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadUint16ZeroNotRoot
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
			name:     "PtrHeadUint16ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A uint16 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A uint16 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadUint16ZeroNotRootString",
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
					A uint16 `json:"a,string"`
				}
			}{A: new(struct {
				A uint16 `json:"a,string"`
			})},
		},

		// PtrHeadUint16NotRoot
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
			name:     "PtrHeadUint16NotRootOmitEmpty",
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
					A uint16 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A uint16 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint16NotRootString",
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
					A uint16 `json:"a,string"`
				}
			}{A: &(struct {
				A uint16 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadUint16PtrNotRoot
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
			name:     "PtrHeadUint16PtrNotRootOmitEmpty",
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
					A *uint16 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,omitempty"`
			}{A: uint16ptr(1)})},
		},
		{
			name:     "PtrHeadUint16PtrNotRootString",
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
					A *uint16 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,string"`
			}{A: uint16ptr(1)})},
		},

		// PtrHeadUint16PtrNilNotRoot
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
			name:     "PtrHeadUint16PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *uint16 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint16PtrNilNotRootString",
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
					A *uint16 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadUint16NilNotRoot
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
			name:     "PtrHeadUint16NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *uint16 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint16NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint16 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadUint16ZeroMultiFieldsNotRoot
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
			name:     "HeadUint16ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A uint16 `json:"a,omitempty"`
				}
				B struct {
					B uint16 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint16ZeroMultiFieldsNotRootString",
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
					A uint16 `json:"a,string"`
				}
				B struct {
					B uint16 `json:"b,string"`
				}
			}{},
		},

		// HeadUint16MultiFieldsNotRoot
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
			name:     "HeadUint16MultiFieldsNotRootOmitEmpty",
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
					A uint16 `json:"a,omitempty"`
				}
				B struct {
					B uint16 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint16 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint16 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadUint16MultiFieldsNotRootString",
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
					A uint16 `json:"a,string"`
				}
				B struct {
					B uint16 `json:"b,string"`
				}
			}{A: struct {
				A uint16 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint16 `json:"b,string"`
			}{B: 2}},
		},

		// HeadUint16PtrMultiFieldsNotRoot
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
			name:     "HeadUint16PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint16 `json:"a,omitempty"`
				}
				B struct {
					B *uint16 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint16 `json:"a,omitempty"`
			}{A: uint16ptr(1)}, B: struct {
				B *uint16 `json:"b,omitempty"`
			}{B: uint16ptr(2)}},
		},
		{
			name:     "HeadUint16PtrMultiFieldsNotRootString",
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
					A *uint16 `json:"a,string"`
				}
				B struct {
					B *uint16 `json:"b,string"`
				}
			}{A: struct {
				A *uint16 `json:"a,string"`
			}{A: uint16ptr(1)}, B: struct {
				B *uint16 `json:"b,string"`
			}{B: uint16ptr(2)}},
		},

		// HeadUint16PtrNilMultiFieldsNotRoot
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
			name:     "HeadUint16PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *uint16 `json:"a,omitempty"`
				}
				B struct {
					B *uint16 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint16 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *uint16 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadUint16PtrNilMultiFieldsNotRootString",
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
					A *uint16 `json:"a,string"`
				}
				B struct {
					B *uint16 `json:"b,string"`
				}
			}{A: struct {
				A *uint16 `json:"a,string"`
			}{A: nil}, B: struct {
				B *uint16 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadUint16ZeroMultiFieldsNotRoot
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
			name:     "PtrHeadUint16ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A uint16 `json:"a,omitempty"`
				}
				B struct {
					B uint16 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint16ZeroMultiFieldsNotRootString",
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
					A uint16 `json:"a,string"`
				}
				B struct {
					B uint16 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadUint16MultiFieldsNotRoot
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
			name:     "PtrHeadUint16MultiFieldsNotRootOmitEmpty",
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
					A uint16 `json:"a,omitempty"`
				}
				B struct {
					B uint16 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint16 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint16 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint16MultiFieldsNotRootString",
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
					A uint16 `json:"a,string"`
				}
				B struct {
					B uint16 `json:"b,string"`
				}
			}{A: struct {
				A uint16 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint16 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadUint16PtrMultiFieldsNotRoot
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
			name:     "PtrHeadUint16PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint16 `json:"a,omitempty"`
				}
				B *struct {
					B *uint16 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,omitempty"`
			}{A: uint16ptr(1)}), B: &(struct {
				B *uint16 `json:"b,omitempty"`
			}{B: uint16ptr(2)})},
		},
		{
			name:     "PtrHeadUint16PtrMultiFieldsNotRootString",
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
					A *uint16 `json:"a,string"`
				}
				B *struct {
					B *uint16 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,string"`
			}{A: uint16ptr(1)}), B: &(struct {
				B *uint16 `json:"b,string"`
			}{B: uint16ptr(2)})},
		},

		// PtrHeadUint16PtrNilMultiFieldsNotRoot
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
			name:     "PtrHeadUint16PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *uint16 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *uint16 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint16NilMultiFieldsNotRoot
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
			name:     "PtrHeadUint16NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint16 `json:"a,omitempty"`
				}
				B *struct {
					B *uint16 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint16NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint16 `json:"a,string"`
				}
				B *struct {
					B *uint16 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint16DoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint16DoubleMultiFieldsNotRootOmitEmpty",
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
					A uint16 `json:"a,omitempty"`
					B uint16 `json:"b,omitempty"`
				}
				B *struct {
					A uint16 `json:"a,omitempty"`
					B uint16 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A uint16 `json:"a,omitempty"`
				B uint16 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A uint16 `json:"a,omitempty"`
				B uint16 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint16DoubleMultiFieldsNotRootString",
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
					A uint16 `json:"a,string"`
					B uint16 `json:"b,string"`
				}
				B *struct {
					A uint16 `json:"a,string"`
					B uint16 `json:"b,string"`
				}
			}{A: &(struct {
				A uint16 `json:"a,string"`
				B uint16 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A uint16 `json:"a,string"`
				B uint16 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadUint16NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint16NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A uint16 `json:"a,omitempty"`
					B uint16 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A uint16 `json:"a,omitempty"`
					B uint16 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A uint16 `json:"a,string"`
					B uint16 `json:"b,string"`
				}
				B *struct {
					A uint16 `json:"a,string"`
					B uint16 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint16NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint16NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint16 `json:"a,omitempty"`
					B uint16 `json:"b,omitempty"`
				}
				B *struct {
					A uint16 `json:"a,omitempty"`
					B uint16 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint16NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint16 `json:"a,string"`
					B uint16 `json:"b,string"`
				}
				B *struct {
					A uint16 `json:"a,string"`
					B uint16 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint16PtrDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint16PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *uint16 `json:"a,omitempty"`
					B *uint16 `json:"b,omitempty"`
				}
				B *struct {
					A *uint16 `json:"a,omitempty"`
					B *uint16 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,omitempty"`
				B *uint16 `json:"b,omitempty"`
			}{A: uint16ptr(1), B: uint16ptr(2)}), B: &(struct {
				A *uint16 `json:"a,omitempty"`
				B *uint16 `json:"b,omitempty"`
			}{A: uint16ptr(3), B: uint16ptr(4)})},
		},
		{
			name:     "PtrHeadUint16PtrDoubleMultiFieldsNotRootString",
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
					A *uint16 `json:"a,string"`
					B *uint16 `json:"b,string"`
				}
				B *struct {
					A *uint16 `json:"a,string"`
					B *uint16 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint16 `json:"a,string"`
				B *uint16 `json:"b,string"`
			}{A: uint16ptr(1), B: uint16ptr(2)}), B: &(struct {
				A *uint16 `json:"a,string"`
				B *uint16 `json:"b,string"`
			}{A: uint16ptr(3), B: uint16ptr(4)})},
		},

		// PtrHeadUint16PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint16PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a,omitempty"`
					B *uint16 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *uint16 `json:"a,omitempty"`
					B *uint16 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint16PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint16 `json:"a,string"`
					B *uint16 `json:"b,string"`
				}
				B *struct {
					A *uint16 `json:"a,string"`
					B *uint16 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint16PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadUint16PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint16 `json:"a,omitempty"`
					B *uint16 `json:"b,omitempty"`
				}
				B *struct {
					A *uint16 `json:"a,omitempty"`
					B *uint16 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint16PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint16 `json:"a,string"`
					B *uint16 `json:"b,string"`
				}
				B *struct {
					A *uint16 `json:"a,string"`
					B *uint16 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadUint16
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
			name:     "AnonymousHeadUint16OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint16OmitEmpty
				B uint16 `json:"b,omitempty"`
			}{
				structUint16OmitEmpty: structUint16OmitEmpty{A: 1},
				B:                     2,
			},
		},
		{
			name:     "AnonymousHeadUint16String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint16String
				B uint16 `json:"b,string"`
			}{
				structUint16String: structUint16String{A: 1},
				B:                  2,
			},
		},

		// PtrAnonymousHeadUint16
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
			name:     "PtrAnonymousHeadUint16OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint16OmitEmpty
				B uint16 `json:"b,omitempty"`
			}{
				structUint16OmitEmpty: &structUint16OmitEmpty{A: 1},
				B:                     2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint16String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint16String
				B uint16 `json:"b,string"`
			}{
				structUint16String: &structUint16String{A: 1},
				B:                  2,
			},
		},

		// NilPtrAnonymousHeadUint16
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
			name:     "NilPtrAnonymousHeadUint16OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint16OmitEmpty
				B uint16 `json:"b,omitempty"`
			}{
				structUint16OmitEmpty: nil,
				B:                     2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint16String
				B uint16 `json:"b,string"`
			}{
				structUint16String: nil,
				B:                  2,
			},
		},

		// AnonymousHeadUint16Ptr
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
			name:     "AnonymousHeadUint16PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint16PtrOmitEmpty
				B *uint16 `json:"b,omitempty"`
			}{
				structUint16PtrOmitEmpty: structUint16PtrOmitEmpty{A: uint16ptr(1)},
				B:                        uint16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint16PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint16PtrString
				B *uint16 `json:"b,string"`
			}{
				structUint16PtrString: structUint16PtrString{A: uint16ptr(1)},
				B:                     uint16ptr(2),
			},
		},

		// AnonymousHeadUint16PtrNil
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
			name:     "AnonymousHeadUint16PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structUint16PtrOmitEmpty
				B *uint16 `json:"b,omitempty"`
			}{
				structUint16PtrOmitEmpty: structUint16PtrOmitEmpty{A: nil},
				B:                        uint16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint16PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structUint16PtrString
				B *uint16 `json:"b,string"`
			}{
				structUint16PtrString: structUint16PtrString{A: nil},
				B:                     uint16ptr(2),
			},
		},

		// PtrAnonymousHeadUint16Ptr
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
			name:     "PtrAnonymousHeadUint16PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint16PtrOmitEmpty
				B *uint16 `json:"b,omitempty"`
			}{
				structUint16PtrOmitEmpty: &structUint16PtrOmitEmpty{A: uint16ptr(1)},
				B:                        uint16ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint16PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint16PtrString
				B *uint16 `json:"b,string"`
			}{
				structUint16PtrString: &structUint16PtrString{A: uint16ptr(1)},
				B:                     uint16ptr(2),
			},
		},

		// NilPtrAnonymousHeadUint16Ptr
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
			name:     "NilPtrAnonymousHeadUint16PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint16PtrOmitEmpty
				B *uint16 `json:"b,omitempty"`
			}{
				structUint16PtrOmitEmpty: nil,
				B:                        uint16ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint16PtrString
				B *uint16 `json:"b,string"`
			}{
				structUint16PtrString: nil,
				B:                     uint16ptr(2),
			},
		},

		// AnonymousHeadUint16Only
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
			name:     "AnonymousHeadUint16OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint16OmitEmpty
			}{
				structUint16OmitEmpty: structUint16OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadUint16OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint16String
			}{
				structUint16String: structUint16String{A: 1},
			},
		},

		// PtrAnonymousHeadUint16Only
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
			name:     "PtrAnonymousHeadUint16OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint16OmitEmpty
			}{
				structUint16OmitEmpty: &structUint16OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint16OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint16String
			}{
				structUint16String: &structUint16String{A: 1},
			},
		},

		// NilPtrAnonymousHeadUint16Only
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
			name:     "NilPtrAnonymousHeadUint16OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint16OmitEmpty
			}{
				structUint16OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint16String
			}{
				structUint16String: nil,
			},
		},

		// AnonymousHeadUint16PtrOnly
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
			name:     "AnonymousHeadUint16PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint16PtrOmitEmpty
			}{
				structUint16PtrOmitEmpty: structUint16PtrOmitEmpty{A: uint16ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint16PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint16PtrString
			}{
				structUint16PtrString: structUint16PtrString{A: uint16ptr(1)},
			},
		},

		// AnonymousHeadUint16PtrNilOnly
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
			name:     "AnonymousHeadUint16PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structUint16PtrOmitEmpty
			}{
				structUint16PtrOmitEmpty: structUint16PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadUint16PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint16PtrString
			}{
				structUint16PtrString: structUint16PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadUint16PtrOnly
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
			name:     "PtrAnonymousHeadUint16PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint16PtrOmitEmpty
			}{
				structUint16PtrOmitEmpty: &structUint16PtrOmitEmpty{A: uint16ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadUint16PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint16PtrString
			}{
				structUint16PtrString: &structUint16PtrString{A: uint16ptr(1)},
			},
		},

		// NilPtrAnonymousHeadUint16PtrOnly
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
		{
			name:     "NilPtrAnonymousHeadUint16PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint16PtrOmitEmpty
			}{
				structUint16PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint16PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint16PtrString
			}{
				structUint16PtrString: nil,
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
