package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverInt(t *testing.T) {
	type structInt struct {
		A int `json:"a"`
	}
	type structIntOmitEmpty struct {
		A int `json:"a,omitempty"`
	}
	type structIntString struct {
		A int `json:"a,string"`
	}

	type structIntPtr struct {
		A *int `json:"a"`
	}
	type structIntPtrOmitEmpty struct {
		A *int `json:"a,omitempty"`
	}
	type structIntPtrString struct {
		A *int `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadIntZero
		{
			name:     "HeadIntZero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A int `json:"a"`
			}{},
		},
		{
			name:     "HeadIntZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadIntZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A int `json:"a,string"`
			}{},
		},

		// HeadInt
		{
			name:     "HeadInt",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadIntOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A int `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadIntString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A int `json:"a,string"`
			}{A: 1},
		},

		// HeadIntPtr
		{
			name:     "HeadIntPtr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int `json:"a"`
			}{A: intptr(1)},
		},
		{
			name:     "HeadIntPtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *int `json:"a,omitempty"`
			}{A: intptr(1)},
		},
		{
			name:     "HeadIntPtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *int `json:"a,string"`
			}{A: intptr(1)},
		},

		// HeadIntPtrNil
		{
			name:     "HeadIntPtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadIntPtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadIntPtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *int `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadIntZero
		{
			name:     "PtrHeadIntZero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A int `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadIntZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadIntZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A int `json:"a,string"`
			}{},
		},

		// PtrHeadInt
		{
			name:     "PtrHeadInt",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadIntOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A int `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadIntString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A int `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadIntPtr
		{
			name:     "PtrHeadIntPtr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int `json:"a"`
			}{A: intptr(1)},
		},
		{
			name:     "PtrHeadIntPtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *int `json:"a,omitempty"`
			}{A: intptr(1)},
		},
		{
			name:     "PtrHeadIntPtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *int `json:"a,string"`
			}{A: intptr(1)},
		},

		// PtrHeadIntPtrNil
		{
			name:     "PtrHeadIntPtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadIntPtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadIntPtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *int `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadIntNil
		{
			name:     "PtrHeadIntNil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadIntNilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadIntNilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int `json:"a,string"`
			})(nil),
		},

		// HeadIntZeroMultiFields
		{
			name:     "HeadIntZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A int `json:"a"`
				B int `json:"b"`
			}{},
		},
		{
			name:     "HeadIntZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A int `json:"a,omitempty"`
				B int `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadIntZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A int `json:"a,string"`
				B int `json:"b,string"`
			}{},
		},

		// HeadIntMultiFields
		{
			name:     "HeadIntMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadIntMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A int `json:"a,omitempty"`
				B int `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadIntMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A int `json:"a,string"`
				B int `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadIntPtrMultiFields
		{
			name:     "HeadIntPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(1), B: intptr(2)},
		},
		{
			name:     "HeadIntPtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *int `json:"a,omitempty"`
				B *int `json:"b,omitempty"`
			}{A: intptr(1), B: intptr(2)},
		},
		{
			name:     "HeadIntPtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *int `json:"a,string"`
				B *int `json:"b,string"`
			}{A: intptr(1), B: intptr(2)},
		},

		// HeadIntPtrNilMultiFields
		{
			name:     "HeadIntPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadIntPtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *int `json:"a,omitempty"`
				B *int `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadIntPtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *int `json:"a,string"`
				B *int `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadIntZeroMultiFields
		{
			name:     "PtrHeadIntZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A int `json:"a"`
				B int `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadIntZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A int `json:"a,omitempty"`
				B int `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadIntZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A int `json:"a,string"`
				B int `json:"b,string"`
			}{},
		},

		// PtrHeadIntMultiFields
		{
			name:     "PtrHeadIntMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadIntMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A int `json:"a,omitempty"`
				B int `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadIntMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A int `json:"a,string"`
				B int `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadIntPtrMultiFields
		{
			name:     "PtrHeadIntPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(1), B: intptr(2)},
		},
		{
			name:     "PtrHeadIntPtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *int `json:"a,omitempty"`
				B *int `json:"b,omitempty"`
			}{A: intptr(1), B: intptr(2)},
		},
		{
			name:     "PtrHeadIntPtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *int `json:"a,string"`
				B *int `json:"b,string"`
			}{A: intptr(1), B: intptr(2)},
		},

		// PtrHeadIntPtrNilMultiFields
		{
			name:     "PtrHeadIntPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntPtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *int `json:"a,omitempty"`
				B *int `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntPtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *int `json:"a,string"`
				B *int `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadIntNilMultiFields
		{
			name:     "PtrHeadIntNilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int `json:"a"`
				B *int `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadIntNilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int `json:"a,omitempty"`
				B *int `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadIntNilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *int `json:"a,string"`
				B *int `json:"b,string"`
			})(nil),
		},

		// HeadIntZeroNotRoot
		{
			name:     "HeadIntZeroNotRoot",
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
					A int `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadIntZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A int `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadIntZeroNotRootString",
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
					A int `json:"a,string"`
				}
			}{},
		},

		// HeadIntNotRoot
		{
			name:     "HeadIntNotRoot",
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
					A int `json:"a"`
				}
			}{A: struct {
				A int `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadIntNotRootOmitEmpty",
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
					A int `json:"a,omitempty"`
				}
			}{A: struct {
				A int `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadIntNotRootString",
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
					A int `json:"a,string"`
				}
			}{A: struct {
				A int `json:"a,string"`
			}{A: 1}},
		},

		// HeadIntPtrNotRoot
		{
			name:     "HeadIntPtrNotRoot",
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
					A *int `json:"a"`
				}
			}{A: struct {
				A *int `json:"a"`
			}{intptr(1)}},
		},
		{
			name:     "HeadIntPtrNotRootOmitEmpty",
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
					A *int `json:"a,omitempty"`
				}
			}{A: struct {
				A *int `json:"a,omitempty"`
			}{intptr(1)}},
		},
		{
			name:     "HeadIntPtrNotRootString",
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
					A *int `json:"a,string"`
				}
			}{A: struct {
				A *int `json:"a,string"`
			}{intptr(1)}},
		},

		// HeadIntPtrNilNotRoot
		{
			name:     "HeadIntPtrNilNotRoot",
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
					A *int `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadIntPtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *int `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadIntPtrNilNotRootString",
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
					A *int `json:"a,string"`
				}
			}{},
		},

		// PtrHeadIntZeroNotRoot
		{
			name:     "PtrHeadIntZeroNotRoot",
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
					A int `json:"a"`
				}
			}{A: new(struct {
				A int `json:"a"`
			})},
		},
		{
			name:     "PtrHeadIntZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A int `json:"a,omitempty"`
				}
			}{A: new(struct {
				A int `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadIntZeroNotRootString",
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
					A int `json:"a,string"`
				}
			}{A: new(struct {
				A int `json:"a,string"`
			})},
		},

		// PtrHeadIntNotRoot
		{
			name:     "PtrHeadIntNotRoot",
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
					A int `json:"a"`
				}
			}{A: &(struct {
				A int `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadIntNotRootOmitEmpty",
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
					A int `json:"a,omitempty"`
				}
			}{A: &(struct {
				A int `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadIntNotRootString",
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
					A int `json:"a,string"`
				}
			}{A: &(struct {
				A int `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadIntPtrNotRoot
		{
			name:     "PtrHeadIntPtrNotRoot",
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
					A *int `json:"a"`
				}
			}{A: &(struct {
				A *int `json:"a"`
			}{A: intptr(1)})},
		},
		{
			name:     "PtrHeadIntPtrNotRootOmitEmpty",
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
					A *int `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int `json:"a,omitempty"`
			}{A: intptr(1)})},
		},
		{
			name:     "PtrHeadIntPtrNotRootString",
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
					A *int `json:"a,string"`
				}
			}{A: &(struct {
				A *int `json:"a,string"`
			}{A: intptr(1)})},
		},

		// PtrHeadIntPtrNilNotRoot
		{
			name:     "PtrHeadIntPtrNilNotRoot",
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
					A *int `json:"a"`
				}
			}{A: &(struct {
				A *int `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadIntPtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *int `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *int `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadIntPtrNilNotRootString",
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
					A *int `json:"a,string"`
				}
			}{A: &(struct {
				A *int `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadIntNilNotRoot
		{
			name:     "PtrHeadIntNilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadIntNilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *int `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadIntNilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *int `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadIntZeroMultiFieldsNotRoot
		{
			name:     "HeadIntZeroMultiFieldsNotRoot",
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
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadIntZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A int `json:"a,omitempty"`
				}
				B struct {
					B int `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadIntZeroMultiFieldsNotRootString",
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
					A int `json:"a,string"`
				}
				B struct {
					B int `json:"b,string"`
				}
			}{},
		},

		// HeadIntMultiFieldsNotRoot
		{
			name:     "HeadIntMultiFieldsNotRoot",
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
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{A: struct {
				A int `json:"a"`
			}{A: 1}, B: struct {
				B int `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadIntMultiFieldsNotRootOmitEmpty",
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
					A int `json:"a,omitempty"`
				}
				B struct {
					B int `json:"b,omitempty"`
				}
			}{A: struct {
				A int `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadIntMultiFieldsNotRootString",
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
					A int `json:"a,string"`
				}
				B struct {
					B int `json:"b,string"`
				}
			}{A: struct {
				A int `json:"a,string"`
			}{A: 1}, B: struct {
				B int `json:"b,string"`
			}{B: 2}},
		},

		// HeadIntPtrMultiFieldsNotRoot
		{
			name:     "HeadIntPtrMultiFieldsNotRoot",
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
					A *int `json:"a"`
				}
				B struct {
					B *int `json:"b"`
				}
			}{A: struct {
				A *int `json:"a"`
			}{A: intptr(1)}, B: struct {
				B *int `json:"b"`
			}{B: intptr(2)}},
		},
		{
			name:     "HeadIntPtrMultiFieldsNotRootOmitEmpty",
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
					A *int `json:"a,omitempty"`
				}
				B struct {
					B *int `json:"b,omitempty"`
				}
			}{A: struct {
				A *int `json:"a,omitempty"`
			}{A: intptr(1)}, B: struct {
				B *int `json:"b,omitempty"`
			}{B: intptr(2)}},
		},
		{
			name:     "HeadIntPtrMultiFieldsNotRootString",
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
					A *int `json:"a,string"`
				}
				B struct {
					B *int `json:"b,string"`
				}
			}{A: struct {
				A *int `json:"a,string"`
			}{A: intptr(1)}, B: struct {
				B *int `json:"b,string"`
			}{B: intptr(2)}},
		},

		// HeadIntPtrNilMultiFieldsNotRoot
		{
			name:     "HeadIntPtrNilMultiFieldsNotRoot",
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
					A *int `json:"a"`
				}
				B struct {
					B *int `json:"b"`
				}
			}{A: struct {
				A *int `json:"a"`
			}{A: nil}, B: struct {
				B *int `json:"b"`
			}{B: nil}},
		},
		{
			name:     "HeadIntPtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *int `json:"a,omitempty"`
				}
				B struct {
					B *int `json:"b,omitempty"`
				}
			}{A: struct {
				A *int `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *int `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadIntPtrNilMultiFieldsNotRootString",
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
					A *int `json:"a,string"`
				}
				B struct {
					B *int `json:"b,string"`
				}
			}{A: struct {
				A *int `json:"a,string"`
			}{A: nil}, B: struct {
				B *int `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadIntZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadIntZeroMultiFieldsNotRoot",
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
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadIntZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A int `json:"a,omitempty"`
				}
				B struct {
					B int `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadIntZeroMultiFieldsNotRootString",
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
					A int `json:"a,string"`
				}
				B struct {
					B int `json:"b,string"`
				}
			}{},
		},

		// PtrHeadIntMultiFieldsNotRoot
		{
			name:     "PtrHeadIntMultiFieldsNotRoot",
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
					A int `json:"a"`
				}
				B struct {
					B int `json:"b"`
				}
			}{A: struct {
				A int `json:"a"`
			}{A: 1}, B: struct {
				B int `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadIntMultiFieldsNotRootOmitEmpty",
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
					A int `json:"a,omitempty"`
				}
				B struct {
					B int `json:"b,omitempty"`
				}
			}{A: struct {
				A int `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B int `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadIntMultiFieldsNotRootString",
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
					A int `json:"a,string"`
				}
				B struct {
					B int `json:"b,string"`
				}
			}{A: struct {
				A int `json:"a,string"`
			}{A: 1}, B: struct {
				B int `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadIntPtrMultiFieldsNotRoot
		{
			name:     "PtrHeadIntPtrMultiFieldsNotRoot",
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
					A *int `json:"a"`
				}
				B *struct {
					B *int `json:"b"`
				}
			}{A: &(struct {
				A *int `json:"a"`
			}{A: intptr(1)}), B: &(struct {
				B *int `json:"b"`
			}{B: intptr(2)})},
		},
		{
			name:     "PtrHeadIntPtrMultiFieldsNotRootOmitEmpty",
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
					A *int `json:"a,omitempty"`
				}
				B *struct {
					B *int `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int `json:"a,omitempty"`
			}{A: intptr(1)}), B: &(struct {
				B *int `json:"b,omitempty"`
			}{B: intptr(2)})},
		},
		{
			name:     "PtrHeadIntPtrMultiFieldsNotRootString",
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
					A *int `json:"a,string"`
				}
				B *struct {
					B *int `json:"b,string"`
				}
			}{A: &(struct {
				A *int `json:"a,string"`
			}{A: intptr(1)}), B: &(struct {
				B *int `json:"b,string"`
			}{B: intptr(2)})},
		},

		// PtrHeadIntPtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadIntPtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int `json:"a"`
				}
				B *struct {
					B *int `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntPtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *int `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntPtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *int `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadIntNilMultiFieldsNotRoot
		{
			name:     "PtrHeadIntNilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int `json:"a"`
				}
				B *struct {
					B *int `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntNilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int `json:"a,omitempty"`
				}
				B *struct {
					B *int `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntNilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int `json:"a,string"`
				}
				B *struct {
					B *int `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadIntDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadIntDoubleMultiFieldsNotRoot",
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
					A int `json:"a"`
					B int `json:"b"`
				}
				B *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
			}{A: &(struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadIntDoubleMultiFieldsNotRootOmitEmpty",
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
					A int `json:"a,omitempty"`
					B int `json:"b,omitempty"`
				}
				B *struct {
					A int `json:"a,omitempty"`
					B int `json:"b,omitempty"`
				}
			}{A: &(struct {
				A int `json:"a,omitempty"`
				B int `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A int `json:"a,omitempty"`
				B int `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadIntDoubleMultiFieldsNotRootString",
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
					A int `json:"a,string"`
					B int `json:"b,string"`
				}
				B *struct {
					A int `json:"a,string"`
					B int `json:"b,string"`
				}
			}{A: &(struct {
				A int `json:"a,string"`
				B int `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A int `json:"a,string"`
				B int `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadIntNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
				B *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A int `json:"a,omitempty"`
					B int `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A int `json:"a,omitempty"`
					B int `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A int `json:"a,string"`
					B int `json:"b,string"`
				}
				B *struct {
					A int `json:"a,string"`
					B int `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadIntNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
				B *struct {
					A int `json:"a"`
					B int `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int `json:"a,omitempty"`
					B int `json:"b,omitempty"`
				}
				B *struct {
					A int `json:"a,omitempty"`
					B int `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A int `json:"a,string"`
					B int `json:"b,string"`
				}
				B *struct {
					A int `json:"a,string"`
					B int `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadIntPtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadIntPtrDoubleMultiFieldsNotRoot",
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
					A *int `json:"a"`
					B *int `json:"b"`
				}
				B *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
			}{A: &(struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(1), B: intptr(2)}), B: &(struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(3), B: intptr(4)})},
		},
		{
			name:     "PtrHeadIntPtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *int `json:"a,omitempty"`
					B *int `json:"b,omitempty"`
				}
				B *struct {
					A *int `json:"a,omitempty"`
					B *int `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *int `json:"a,omitempty"`
				B *int `json:"b,omitempty"`
			}{A: intptr(1), B: intptr(2)}), B: &(struct {
				A *int `json:"a,omitempty"`
				B *int `json:"b,omitempty"`
			}{A: intptr(3), B: intptr(4)})},
		},
		{
			name:     "PtrHeadIntPtrDoubleMultiFieldsNotRootString",
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
					A *int `json:"a,string"`
					B *int `json:"b,string"`
				}
				B *struct {
					A *int `json:"a,string"`
					B *int `json:"b,string"`
				}
			}{A: &(struct {
				A *int `json:"a,string"`
				B *int `json:"b,string"`
			}{A: intptr(1), B: intptr(2)}), B: &(struct {
				A *int `json:"a,string"`
				B *int `json:"b,string"`
			}{A: intptr(3), B: intptr(4)})},
		},

		// PtrHeadIntPtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
				B *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *int `json:"a,omitempty"`
					B *int `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *int `json:"a,omitempty"`
					B *int `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *int `json:"a,string"`
					B *int `json:"b,string"`
				}
				B *struct {
					A *int `json:"a,string"`
					B *int `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadIntPtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
				B *struct {
					A *int `json:"a"`
					B *int `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int `json:"a,omitempty"`
					B *int `json:"b,omitempty"`
				}
				B *struct {
					A *int `json:"a,omitempty"`
					B *int `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadIntPtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *int `json:"a,string"`
					B *int `json:"b,string"`
				}
				B *struct {
					A *int `json:"a,string"`
					B *int `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadInt
		{
			name:     "AnonymousHeadInt",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structInt
				B int `json:"b"`
			}{
				structInt: structInt{A: 1},
				B:         2,
			},
		},
		{
			name:     "AnonymousHeadIntOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structIntOmitEmpty
				B int `json:"b,omitempty"`
			}{
				structIntOmitEmpty: structIntOmitEmpty{A: 1},
				B:                  2,
			},
		},
		{
			name:     "AnonymousHeadIntString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structIntString
				B int `json:"b,string"`
			}{
				structIntString: structIntString{A: 1},
				B:               2,
			},
		},

		// PtrAnonymousHeadInt
		{
			name:     "PtrAnonymousHeadInt",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structInt
				B int `json:"b"`
			}{
				structInt: &structInt{A: 1},
				B:         2,
			},
		},
		{
			name:     "PtrAnonymousHeadIntOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structIntOmitEmpty
				B int `json:"b,omitempty"`
			}{
				structIntOmitEmpty: &structIntOmitEmpty{A: 1},
				B:                  2,
			},
		},
		{
			name:     "PtrAnonymousHeadIntString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structIntString
				B int `json:"b,string"`
			}{
				structIntString: &structIntString{A: 1},
				B:               2,
			},
		},

		// NilPtrAnonymousHeadInt
		{
			name:     "NilPtrAnonymousHeadInt",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structInt
				B int `json:"b"`
			}{
				structInt: nil,
				B:         2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structIntOmitEmpty
				B int `json:"b,omitempty"`
			}{
				structIntOmitEmpty: nil,
				B:                  2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structIntString
				B int `json:"b,string"`
			}{
				structIntString: nil,
				B:               2,
			},
		},

		// AnonymousHeadIntPtr
		{
			name:     "AnonymousHeadIntPtr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: structIntPtr{A: intptr(1)},
				B:            intptr(2),
			},
		},
		{
			name:     "AnonymousHeadIntPtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structIntPtrOmitEmpty
				B *int `json:"b,omitempty"`
			}{
				structIntPtrOmitEmpty: structIntPtrOmitEmpty{A: intptr(1)},
				B:                     intptr(2),
			},
		},
		{
			name:     "AnonymousHeadIntPtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structIntPtrString
				B *int `json:"b,string"`
			}{
				structIntPtrString: structIntPtrString{A: intptr(1)},
				B:                  intptr(2),
			},
		},

		// AnonymousHeadIntPtrNil
		{
			name:     "AnonymousHeadIntPtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: structIntPtr{A: nil},
				B:            intptr(2),
			},
		},
		{
			name:     "AnonymousHeadIntPtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structIntPtrOmitEmpty
				B *int `json:"b,omitempty"`
			}{
				structIntPtrOmitEmpty: structIntPtrOmitEmpty{A: nil},
				B:                     intptr(2),
			},
		},
		{
			name:     "AnonymousHeadIntPtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structIntPtrString
				B *int `json:"b,string"`
			}{
				structIntPtrString: structIntPtrString{A: nil},
				B:                  intptr(2),
			},
		},

		// PtrAnonymousHeadIntPtr
		{
			name:     "PtrAnonymousHeadIntPtr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: &structIntPtr{A: intptr(1)},
				B:            intptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadIntPtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structIntPtrOmitEmpty
				B *int `json:"b,omitempty"`
			}{
				structIntPtrOmitEmpty: &structIntPtrOmitEmpty{A: intptr(1)},
				B:                     intptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadIntPtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structIntPtrString
				B *int `json:"b,string"`
			}{
				structIntPtrString: &structIntPtrString{A: intptr(1)},
				B:                  intptr(2),
			},
		},

		// NilPtrAnonymousHeadIntPtr
		{
			name:     "NilPtrAnonymousHeadIntPtr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structIntPtr
				B *int `json:"b"`
			}{
				structIntPtr: nil,
				B:            intptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntPtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structIntPtrOmitEmpty
				B *int `json:"b,omitempty"`
			}{
				structIntPtrOmitEmpty: nil,
				B:                     intptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntPtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structIntPtrString
				B *int `json:"b,string"`
			}{
				structIntPtrString: nil,
				B:                  intptr(2),
			},
		},

		// AnonymousHeadIntOnly
		{
			name:     "AnonymousHeadIntOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structInt
			}{
				structInt: structInt{A: 1},
			},
		},
		{
			name:     "AnonymousHeadIntOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structIntOmitEmpty
			}{
				structIntOmitEmpty: structIntOmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadIntOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structIntString
			}{
				structIntString: structIntString{A: 1},
			},
		},

		// PtrAnonymousHeadIntOnly
		{
			name:     "PtrAnonymousHeadIntOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structInt
			}{
				structInt: &structInt{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadIntOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structIntOmitEmpty
			}{
				structIntOmitEmpty: &structIntOmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadIntOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structIntString
			}{
				structIntString: &structIntString{A: 1},
			},
		},

		// NilPtrAnonymousHeadIntOnly
		{
			name:     "NilPtrAnonymousHeadIntOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structInt
			}{
				structInt: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structIntOmitEmpty
			}{
				structIntOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structIntString
			}{
				structIntString: nil,
			},
		},

		// AnonymousHeadIntPtrOnly
		{
			name:     "AnonymousHeadIntPtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structIntPtr
			}{
				structIntPtr: structIntPtr{A: intptr(1)},
			},
		},
		{
			name:     "AnonymousHeadIntPtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structIntPtrOmitEmpty
			}{
				structIntPtrOmitEmpty: structIntPtrOmitEmpty{A: intptr(1)},
			},
		},
		{
			name:     "AnonymousHeadIntPtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structIntPtrString
			}{
				structIntPtrString: structIntPtrString{A: intptr(1)},
			},
		},

		// AnonymousHeadIntPtrNilOnly
		{
			name:     "AnonymousHeadIntPtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structIntPtr
			}{
				structIntPtr: structIntPtr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadIntPtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structIntPtrOmitEmpty
			}{
				structIntPtrOmitEmpty: structIntPtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadIntPtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structIntPtrString
			}{
				structIntPtrString: structIntPtrString{A: nil},
			},
		},

		// PtrAnonymousHeadIntPtrOnly
		{
			name:     "PtrAnonymousHeadIntPtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structIntPtr
			}{
				structIntPtr: &structIntPtr{A: intptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadIntPtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structIntPtrOmitEmpty
			}{
				structIntPtrOmitEmpty: &structIntPtrOmitEmpty{A: intptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadIntPtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structIntPtrString
			}{
				structIntPtrString: &structIntPtrString{A: intptr(1)},
			},
		},

		// NilPtrAnonymousHeadIntPtrOnly
		{
			name:     "NilPtrAnonymousHeadIntPtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structIntPtr
			}{
				structIntPtr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntPtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structIntPtrOmitEmpty
			}{
				structIntPtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadIntPtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structIntPtrString
			}{
				structIntPtrString: nil,
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
