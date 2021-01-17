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
	type structInt16OmitEmpty struct {
		A int16 `json:"a,omitempty"`
	}
	type structInt16String struct {
		A int16 `json:"a,string"`
	}

	type structInt16Ptr struct {
		A *int16 `json:"a"`
	}
	type structInt16PtrOmitEmpty struct {
		A *int16 `json:"a,omitempty"`
	}
	type structInt16PtrString struct {
		A *int16 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadInt16Zero
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
			name:     "HeadInt16ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int16 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadInt16ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A int16 `json:"a,string"`
			}{},
		},

		// HeadInt16
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
			name:     "HeadInt16OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int16 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadInt16String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A int16 `json:"a,string"`
			}{A: 1},
		},

		// HeadInt16Ptr
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
			name:     "HeadInt16PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int16 `json:"a,omitempty"`
			}{A: int16ptr(1)},
		},
		{
			name:     "HeadInt16PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *int16 `json:"a,string"`
			}{A: int16ptr(1)},
		},

		// HeadInt16PtrNil
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
			name:     "HeadInt16PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int16 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadInt16PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int16 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt16Zero
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
			name:     "PtrHeadInt16ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int16 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt16ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A int16 `json:"a,string"`
			}{},
		},

		// PtrHeadInt16
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
			name:     "PtrHeadInt16OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int16 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadInt16String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A int16 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadInt16Ptr
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
			name:     "PtrHeadInt16PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int16 `json:"a,omitempty"`
			}{A: int16ptr(1)},
		},
		{
			name:     "PtrHeadInt16PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *int16 `json:"a,string"`
			}{A: int16ptr(1)},
		},

		// PtrHeadInt16PtrNil
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
			name:     "PtrHeadInt16PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int16 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt16PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int16 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadInt16Nil
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
			name:     "PtrHeadInt16NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int16 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt16NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int16 `json:"a,string"`
			})(nil),
		},

		// HeadInt16ZeroMultiFields
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
			name:     "HeadInt16ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int16 `json:"a,omitempty"`
				B int16 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadInt16ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A int16 `json:"a,string"`
				B int16 `json:"b,string"`
			}{},
		},

		// HeadInt16MultiFields
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
			name:     "HeadInt16MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int16 `json:"a,omitempty"`
				B int16 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadInt16MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A int16 `json:"a,string"`
				B int16 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadInt16PtrMultiFields
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
			name:     "HeadInt16PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int16 `json:"a,omitempty"`
				B *int16 `json:"b,omitempty"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},
		{
			name:     "HeadInt16PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *int16 `json:"a,string"`
				B *int16 `json:"b,string"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},

		// HeadInt16PtrNilMultiFields
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
			name:     "HeadInt16PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int16 `json:"a,omitempty"`
				B *int16 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadInt16PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int16 `json:"a,string"`
				B *int16 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt16ZeroMultiFields
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
			name:     "PtrHeadInt16ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int16 `json:"a,omitempty"`
				B int16 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadInt16ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A int16 `json:"a,string"`
				B int16 `json:"b,string"`
			}{},
		},

		// PtrHeadInt16MultiFields
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
			name:     "PtrHeadInt16MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int16 `json:"a,omitempty"`
				B int16 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadInt16MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A int16 `json:"a,string"`
				B int16 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadInt16PtrMultiFields
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
			name:     "PtrHeadInt16PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int16 `json:"a,omitempty"`
				B *int16 `json:"b,omitempty"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},
		{
			name:     "PtrHeadInt16PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *int16 `json:"a,string"`
				B *int16 `json:"b,string"`
			}{A: int16ptr(1), B: int16ptr(2)},
		},

		// PtrHeadInt16PtrNilMultiFields
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
			name:     "PtrHeadInt16PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int16 `json:"a,omitempty"`
				B *int16 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int16 `json:"a,string"`
				B *int16 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt16NilMultiFields
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
			name:     "PtrHeadInt16NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int16 `json:"a,omitempty"`
				B *int16 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadInt16NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int16 `json:"a,string"`
				B *int16 `json:"b,string"`
			})(nil),
		},

		// HeadInt16ZeroNotRoot
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
			name:     "HeadInt16ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A int16 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt16ZeroNotRootString",
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
					A int16 `json:"a,string"`
				}
			}{},
		},

		// HeadInt16NotRoot
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
			name:     "HeadInt16NotRootOmitEmpty",
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
					A int16 `json:"a,omitempty"`
				}
			}{A: struct {
				A int16 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadInt16NotRootString",
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
					A int16 `json:"a,string"`
				}
			}{A: struct {
				A int16 `json:"a,string"`
			}{A: 1}},
		},

		// HeadInt16PtrNotRoot
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
			name:     "HeadInt16PtrNotRootOmitEmpty",
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
					A *int16 `json:"a,omitempty"`
				}
			}{A: struct {
				A *int16 `json:"a,omitempty"`
			}{int16ptr(1)}},
		},
		{
			name:     "HeadInt16PtrNotRootString",
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
					A *int16 `json:"a,string"`
				}
			}{A: struct {
				A *int16 `json:"a,string"`
			}{int16ptr(1)}},
		},

		// HeadInt16PtrNilNotRoot
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
			name:     "HeadInt16PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *int16 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt16PtrNilNotRootString",
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
					A *int16 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadInt16ZeroNotRoot
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
			name:     "PtrHeadInt16ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A int16 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A int16 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadInt16ZeroNotRootString",
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
					A int16 `json:"a,string"`
				}
			}{A: new(struct {
				A int16 `json:"a,string"`
			})},
		},

		// PtrHeadInt16NotRoot
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
			name:     "PtrHeadInt16NotRootOmitEmpty",
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
					A int16 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A int16 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadInt16NotRootString",
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
					A int16 `json:"a,string"`
				}
			}{A: &(struct {
				A int16 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadInt16PtrNotRoot
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
			name:     "PtrHeadInt16PtrNotRootOmitEmpty",
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
					A *int16 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int16 `json:"a,omitempty"`
			}{A: int16ptr(1)})},
		},
		{
			name:     "PtrHeadInt16PtrNotRootString",
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
					A *int16 `json:"a,string"`
				}
			}{A: &(struct {
				A *int16 `json:"a,string"`
			}{A: int16ptr(1)})},
		},

		// PtrHeadInt16PtrNilNotRoot
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
			name:     "PtrHeadInt16PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *int16 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int16 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadInt16PtrNilNotRootString",
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
					A *int16 `json:"a,string"`
				}
			}{A: &(struct {
				A *int16 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadInt16NilNotRoot
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
			name:     "PtrHeadInt16NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *int16 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadInt16NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int16 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadInt16ZeroMultiFieldsNotRoot
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
			name:     "HeadInt16ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A int16 `json:"a,omitempty"`
				}
				B struct {
					B int16 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadInt16ZeroMultiFieldsNotRootString",
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
					A int16 `json:"a,string"`
				}
				B struct {
					B int16 `json:"b,string"`
				}
			}{},
		},

		// HeadInt16MultiFieldsNotRoot
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
			name:     "HeadInt16MultiFieldsNotRootOmitEmpty",
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
					A int16 `json:"a,omitempty"`
				}
				B struct {
					B int16 `json:"b,omitempty"`
				}
			}{A: struct {
				A int16 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int16 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadInt16MultiFieldsNotRootString",
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
					A int16 `json:"a,string"`
				}
				B struct {
					B int16 `json:"b,string"`
				}
			}{A: struct {
				A int16 `json:"a,string"`
			}{A: 1}, B: struct {
				B int16 `json:"b,string"`
			}{B: 2}},
		},

		// HeadInt16PtrMultiFieldsNotRoot
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
			name:     "HeadInt16PtrMultiFieldsNotRootOmitEmpty",
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
					A *int16 `json:"a,omitempty"`
				}
				B struct {
					B *int16 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int16 `json:"a,omitempty"`
			}{A: int16ptr(1)}, B: struct {
				B *int16 `json:"b,omitempty"`
			}{B: int16ptr(2)}},
		},
		{
			name:     "HeadInt16PtrMultiFieldsNotRootString",
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
					A *int16 `json:"a,string"`
				}
				B struct {
					B *int16 `json:"b,string"`
				}
			}{A: struct {
				A *int16 `json:"a,string"`
			}{A: int16ptr(1)}, B: struct {
				B *int16 `json:"b,string"`
			}{B: int16ptr(2)}},
		},

		// HeadInt16PtrNilMultiFieldsNotRoot
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
			name:     "HeadInt16PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *int16 `json:"a,omitempty"`
				}
				B struct {
					B *int16 `json:"b,omitempty"`
				}
			}{A: struct {
				A *int16 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *int16 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadInt16PtrNilMultiFieldsNotRootString",
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
					A *int16 `json:"a,string"`
				}
				B struct {
					B *int16 `json:"b,string"`
				}
			}{A: struct {
				A *int16 `json:"a,string"`
			}{A: nil}, B: struct {
				B *int16 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadInt16ZeroMultiFieldsNotRoot
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
			name:     "PtrHeadInt16ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A int16 `json:"a,omitempty"`
				}
				B struct {
					B int16 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadInt16ZeroMultiFieldsNotRootString",
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
					A int16 `json:"a,string"`
				}
				B struct {
					B int16 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadInt16MultiFieldsNotRoot
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
			name:     "PtrHeadInt16MultiFieldsNotRootOmitEmpty",
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
					A int16 `json:"a,omitempty"`
				}
				B struct {
					B int16 `json:"b,omitempty"`
				}
			}{A: struct {
				A int16 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int16 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadInt16MultiFieldsNotRootString",
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
					A int16 `json:"a,string"`
				}
				B struct {
					B int16 `json:"b,string"`
				}
			}{A: struct {
				A int16 `json:"a,string"`
			}{A: 1}, B: struct {
				B int16 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadInt16PtrMultiFieldsNotRoot
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
			name:     "PtrHeadInt16PtrMultiFieldsNotRootOmitEmpty",
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
					A *int16 `json:"a,omitempty"`
				}
				B *struct {
					B *int16 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int16 `json:"a,omitempty"`
			}{A: int16ptr(1)}), B: &(struct {
				B *int16 `json:"b,omitempty"`
			}{B: int16ptr(2)})},
		},
		{
			name:     "PtrHeadInt16PtrMultiFieldsNotRootString",
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
					A *int16 `json:"a,string"`
				}
				B *struct {
					B *int16 `json:"b,string"`
				}
			}{A: &(struct {
				A *int16 `json:"a,string"`
			}{A: int16ptr(1)}), B: &(struct {
				B *int16 `json:"b,string"`
			}{B: int16ptr(2)})},
		},

		// PtrHeadInt16PtrNilMultiFieldsNotRoot
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
			name:     "PtrHeadInt16PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int16 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *int16 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int16 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *int16 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadInt16NilMultiFieldsNotRoot
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
			name:     "PtrHeadInt16NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int16 `json:"a,omitempty"`
				}
				B *struct {
					B *int16 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt16NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int16 `json:"a,string"`
				}
				B *struct {
					B *int16 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt16DoubleMultiFieldsNotRoot
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
			name:     "PtrHeadInt16DoubleMultiFieldsNotRootOmitEmpty",
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
					A int16 `json:"a,omitempty"`
					B int16 `json:"b,omitempty"`
				}
				B *struct {
					A int16 `json:"a,omitempty"`
					B int16 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A int16 `json:"a,omitempty"`
				B int16 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A int16 `json:"a,omitempty"`
				B int16 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadInt16DoubleMultiFieldsNotRootString",
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
					A int16 `json:"a,string"`
					B int16 `json:"b,string"`
				}
				B *struct {
					A int16 `json:"a,string"`
					B int16 `json:"b,string"`
				}
			}{A: &(struct {
				A int16 `json:"a,string"`
				B int16 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A int16 `json:"a,string"`
				B int16 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadInt16NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadInt16NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A int16 `json:"a,omitempty"`
					B int16 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A int16 `json:"a,omitempty"`
					B int16 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int16 `json:"a,string"`
					B int16 `json:"b,string"`
				}
				B *struct {
					A int16 `json:"a,string"`
					B int16 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt16NilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadInt16NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int16 `json:"a,omitempty"`
					B int16 `json:"b,omitempty"`
				}
				B *struct {
					A int16 `json:"a,omitempty"`
					B int16 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt16NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int16 `json:"a,string"`
					B int16 `json:"b,string"`
				}
				B *struct {
					A int16 `json:"a,string"`
					B int16 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadInt16PtrDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadInt16PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *int16 `json:"a,omitempty"`
					B *int16 `json:"b,omitempty"`
				}
				B *struct {
					A *int16 `json:"a,omitempty"`
					B *int16 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int16 `json:"a,omitempty"`
				B *int16 `json:"b,omitempty"`
			}{A: int16ptr(1), B: int16ptr(2)}), B: &(struct {
				A *int16 `json:"a,omitempty"`
				B *int16 `json:"b,omitempty"`
			}{A: int16ptr(3), B: int16ptr(4)})},
		},
		{
			name:     "PtrHeadInt16PtrDoubleMultiFieldsNotRootString",
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
					A *int16 `json:"a,string"`
					B *int16 `json:"b,string"`
				}
				B *struct {
					A *int16 `json:"a,string"`
					B *int16 `json:"b,string"`
				}
			}{A: &(struct {
				A *int16 `json:"a,string"`
				B *int16 `json:"b,string"`
			}{A: int16ptr(1), B: int16ptr(2)}), B: &(struct {
				A *int16 `json:"a,string"`
				B *int16 `json:"b,string"`
			}{A: int16ptr(3), B: int16ptr(4)})},
		},

		// PtrHeadInt16PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadInt16PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int16 `json:"a,omitempty"`
					B *int16 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *int16 `json:"a,omitempty"`
					B *int16 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadInt16PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int16 `json:"a,string"`
					B *int16 `json:"b,string"`
				}
				B *struct {
					A *int16 `json:"a,string"`
					B *int16 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadInt16PtrNilDoubleMultiFieldsNotRoot
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
			name:     "PtrHeadInt16PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int16 `json:"a,omitempty"`
					B *int16 `json:"b,omitempty"`
				}
				B *struct {
					A *int16 `json:"a,omitempty"`
					B *int16 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadInt16PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int16 `json:"a,string"`
					B *int16 `json:"b,string"`
				}
				B *struct {
					A *int16 `json:"a,string"`
					B *int16 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadInt16
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
			name:     "AnonymousHeadInt16OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt16OmitEmpty
				B int16 `json:"b,omitempty"`
			}{
				structInt16OmitEmpty: structInt16OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "AnonymousHeadInt16String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt16String
				B int16 `json:"b,string"`
			}{
				structInt16String: structInt16String{A: 1},
				B:                 2,
			},
		},

		// PtrAnonymousHeadInt16
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
			name:     "PtrAnonymousHeadInt16OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt16OmitEmpty
				B int16 `json:"b,omitempty"`
			}{
				structInt16OmitEmpty: &structInt16OmitEmpty{A: 1},
				B:                    2,
			},
		},
		{
			name:     "PtrAnonymousHeadInt16String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt16String
				B int16 `json:"b,string"`
			}{
				structInt16String: &structInt16String{A: 1},
				B:                 2,
			},
		},

		// NilPtrAnonymousHeadInt16
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
			name:     "NilPtrAnonymousHeadInt16OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt16OmitEmpty
				B int16 `json:"b,omitempty"`
			}{
				structInt16OmitEmpty: nil,
				B:                    2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt16String
				B int16 `json:"b,string"`
			}{
				structInt16String: nil,
				B:                 2,
			},
		},

		// AnonymousHeadInt16Ptr
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
			name:     "AnonymousHeadInt16PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt16PtrOmitEmpty
				B *int16 `json:"b,omitempty"`
			}{
				structInt16PtrOmitEmpty: structInt16PtrOmitEmpty{A: int16ptr(1)},
				B:                       int16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt16PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structInt16PtrString
				B *int16 `json:"b,string"`
			}{
				structInt16PtrString: structInt16PtrString{A: int16ptr(1)},
				B:                    int16ptr(2),
			},
		},

		// AnonymousHeadInt16PtrNil
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
			name:     "AnonymousHeadInt16PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structInt16PtrOmitEmpty
				B *int16 `json:"b,omitempty"`
			}{
				structInt16PtrOmitEmpty: structInt16PtrOmitEmpty{A: nil},
				B:                       int16ptr(2),
			},
		},
		{
			name:     "AnonymousHeadInt16PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structInt16PtrString
				B *int16 `json:"b,string"`
			}{
				structInt16PtrString: structInt16PtrString{A: nil},
				B:                    int16ptr(2),
			},
		},

		// PtrAnonymousHeadInt16Ptr
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
			name:     "PtrAnonymousHeadInt16PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt16PtrOmitEmpty
				B *int16 `json:"b,omitempty"`
			}{
				structInt16PtrOmitEmpty: &structInt16PtrOmitEmpty{A: int16ptr(1)},
				B:                       int16ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadInt16PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structInt16PtrString
				B *int16 `json:"b,string"`
			}{
				structInt16PtrString: &structInt16PtrString{A: int16ptr(1)},
				B:                    int16ptr(2),
			},
		},

		// NilPtrAnonymousHeadInt16Ptr
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
			name:     "NilPtrAnonymousHeadInt16PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt16PtrOmitEmpty
				B *int16 `json:"b,omitempty"`
			}{
				structInt16PtrOmitEmpty: nil,
				B:                       int16ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structInt16PtrString
				B *int16 `json:"b,string"`
			}{
				structInt16PtrString: nil,
				B:                    int16ptr(2),
			},
		},

		// AnonymousHeadInt16Only
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
			name:     "AnonymousHeadInt16OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt16OmitEmpty
			}{
				structInt16OmitEmpty: structInt16OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadInt16OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt16String
			}{
				structInt16String: structInt16String{A: 1},
			},
		},

		// PtrAnonymousHeadInt16Only
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
			name:     "PtrAnonymousHeadInt16OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt16OmitEmpty
			}{
				structInt16OmitEmpty: &structInt16OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadInt16OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt16String
			}{
				structInt16String: &structInt16String{A: 1},
			},
		},

		// NilPtrAnonymousHeadInt16Only
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
			name:     "NilPtrAnonymousHeadInt16OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt16OmitEmpty
			}{
				structInt16OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt16String
			}{
				structInt16String: nil,
			},
		},

		// AnonymousHeadInt16PtrOnly
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
			name:     "AnonymousHeadInt16PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt16PtrOmitEmpty
			}{
				structInt16PtrOmitEmpty: structInt16PtrOmitEmpty{A: int16ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadInt16PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structInt16PtrString
			}{
				structInt16PtrString: structInt16PtrString{A: int16ptr(1)},
			},
		},

		// AnonymousHeadInt16PtrNilOnly
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
			name:     "AnonymousHeadInt16PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structInt16PtrOmitEmpty
			}{
				structInt16PtrOmitEmpty: structInt16PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadInt16PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structInt16PtrString
			}{
				structInt16PtrString: structInt16PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadInt16PtrOnly
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
			name:     "PtrAnonymousHeadInt16PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt16PtrOmitEmpty
			}{
				structInt16PtrOmitEmpty: &structInt16PtrOmitEmpty{A: int16ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadInt16PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structInt16PtrString
			}{
				structInt16PtrString: &structInt16PtrString{A: int16ptr(1)},
			},
		},

		// NilPtrAnonymousHeadInt16PtrOnly
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
		{
			name:     "NilPtrAnonymousHeadInt16PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt16PtrOmitEmpty
			}{
				structInt16PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadInt16PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt16PtrString
			}{
				structInt16PtrString: nil,
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
