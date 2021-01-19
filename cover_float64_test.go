package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverFloat64(t *testing.T) {
	type structFloat64 struct {
		A float64 `json:"a"`
	}
	type structFloat64OmitEmpty struct {
		A float64 `json:"a,omitempty"`
	}
	type structFloat64String struct {
		A float64 `json:"a,string"`
	}

	type structFloat64Ptr struct {
		A *float64 `json:"a"`
	}
	type structFloat64PtrOmitEmpty struct {
		A *float64 `json:"a,omitempty"`
	}
	type structFloat64PtrString struct {
		A *float64 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadFloat64Zero
		{
			name:     "HeadFloat64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A float64 `json:"a"`
			}{},
		},
		{
			name:     "HeadFloat64ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A float64 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadFloat64ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A float64 `json:"a,string"`
			}{},
		},

		// HeadFloat64
		{
			name:     "HeadFloat64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A float64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadFloat64OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A float64 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadFloat64String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A float64 `json:"a,string"`
			}{A: 1},
		},

		// HeadFloat64Ptr
		{
			name:     "HeadFloat64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)},
		},
		{
			name:     "HeadFloat64PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *float64 `json:"a,omitempty"`
			}{A: float64ptr(1)},
		},
		{
			name:     "HeadFloat64PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *float64 `json:"a,string"`
			}{A: float64ptr(1)},
		},

		// HeadFloat64PtrNil
		{
			name:     "HeadFloat64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *float64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadFloat64PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *float64 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadFloat64PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *float64 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadFloat64Zero
		{
			name:     "PtrHeadFloat64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A float64 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadFloat64ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A float64 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadFloat64ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A float64 `json:"a,string"`
			}{},
		},

		// PtrHeadFloat64
		{
			name:     "PtrHeadFloat64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A float64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadFloat64OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A float64 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadFloat64String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A float64 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadFloat64Ptr
		{
			name:     "PtrHeadFloat64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)},
		},
		{
			name:     "PtrHeadFloat64PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *float64 `json:"a,omitempty"`
			}{A: float64ptr(1)},
		},
		{
			name:     "PtrHeadFloat64PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *float64 `json:"a,string"`
			}{A: float64ptr(1)},
		},

		// PtrHeadFloat64PtrNil
		{
			name:     "PtrHeadFloat64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *float64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *float64 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *float64 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadFloat64Nil
		{
			name:     "PtrHeadFloat64Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a,string"`
			})(nil),
		},

		// HeadFloat64ZeroMultiFields
		{
			name:     "HeadFloat64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{},
		},
		{
			name:     "HeadFloat64ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A float64 `json:"a,omitempty"`
				B float64 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadFloat64ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A float64 `json:"a,string"`
				B float64 `json:"b,string"`
			}{},
		},

		// HeadFloat64MultiFields
		{
			name:     "HeadFloat64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadFloat64MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A float64 `json:"a,omitempty"`
				B float64 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadFloat64MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A float64 `json:"a,string"`
				B float64 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadFloat64PtrMultiFields
		{
			name:     "HeadFloat64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},
		{
			name:     "HeadFloat64PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *float64 `json:"a,omitempty"`
				B *float64 `json:"b,omitempty"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},
		{
			name:     "HeadFloat64PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *float64 `json:"a,string"`
				B *float64 `json:"b,string"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},

		// HeadFloat64PtrNilMultiFields
		{
			name:     "HeadFloat64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadFloat64PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *float64 `json:"a,omitempty"`
				B *float64 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadFloat64PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *float64 `json:"a,string"`
				B *float64 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadFloat64ZeroMultiFields
		{
			name:     "PtrHeadFloat64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadFloat64ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A float64 `json:"a,omitempty"`
				B float64 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadFloat64ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A float64 `json:"a,string"`
				B float64 `json:"b,string"`
			}{},
		},

		// PtrHeadFloat64MultiFields
		{
			name:     "PtrHeadFloat64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadFloat64MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A float64 `json:"a,omitempty"`
				B float64 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadFloat64MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A float64 `json:"a,string"`
				B float64 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadFloat64PtrMultiFields
		{
			name:     "PtrHeadFloat64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},
		{
			name:     "PtrHeadFloat64PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *float64 `json:"a,omitempty"`
				B *float64 `json:"b,omitempty"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},
		{
			name:     "PtrHeadFloat64PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *float64 `json:"a,string"`
				B *float64 `json:"b,string"`
			}{A: float64ptr(1), B: float64ptr(2)},
		},

		// PtrHeadFloat64PtrNilMultiFields
		{
			name:     "PtrHeadFloat64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *float64 `json:"a,omitempty"`
				B *float64 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *float64 `json:"a,string"`
				B *float64 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadFloat64NilMultiFields
		{
			name:     "PtrHeadFloat64NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a,omitempty"`
				B *float64 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *float64 `json:"a,string"`
				B *float64 `json:"b,string"`
			})(nil),
		},

		// HeadFloat64ZeroNotRoot
		{
			name:     "HeadFloat64ZeroNotRoot",
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
					A float64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadFloat64ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A float64 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadFloat64ZeroNotRootString",
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
					A float64 `json:"a,string"`
				}
			}{},
		},

		// HeadFloat64NotRoot
		{
			name:     "HeadFloat64NotRoot",
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
					A float64 `json:"a"`
				}
			}{A: struct {
				A float64 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadFloat64NotRootOmitEmpty",
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
					A float64 `json:"a,omitempty"`
				}
			}{A: struct {
				A float64 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadFloat64NotRootString",
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
					A float64 `json:"a,string"`
				}
			}{A: struct {
				A float64 `json:"a,string"`
			}{A: 1}},
		},

		// HeadFloat64PtrNotRoot
		{
			name:     "HeadFloat64PtrNotRoot",
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
					A *float64 `json:"a"`
				}
			}{A: struct {
				A *float64 `json:"a"`
			}{float64ptr(1)}},
		},
		{
			name:     "HeadFloat64PtrNotRootOmitEmpty",
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
					A *float64 `json:"a,omitempty"`
				}
			}{A: struct {
				A *float64 `json:"a,omitempty"`
			}{float64ptr(1)}},
		},
		{
			name:     "HeadFloat64PtrNotRootString",
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
					A *float64 `json:"a,string"`
				}
			}{A: struct {
				A *float64 `json:"a,string"`
			}{float64ptr(1)}},
		},

		// HeadFloat64PtrNilNotRoot
		{
			name:     "HeadFloat64PtrNilNotRoot",
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
					A *float64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadFloat64PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *float64 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadFloat64PtrNilNotRootString",
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
					A *float64 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadFloat64ZeroNotRoot
		{
			name:     "PtrHeadFloat64ZeroNotRoot",
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
					A float64 `json:"a"`
				}
			}{A: new(struct {
				A float64 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadFloat64ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A float64 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A float64 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadFloat64ZeroNotRootString",
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
					A float64 `json:"a,string"`
				}
			}{A: new(struct {
				A float64 `json:"a,string"`
			})},
		},

		// PtrHeadFloat64NotRoot
		{
			name:     "PtrHeadFloat64NotRoot",
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
					A float64 `json:"a"`
				}
			}{A: &(struct {
				A float64 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadFloat64NotRootOmitEmpty",
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
					A float64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A float64 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadFloat64NotRootString",
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
					A float64 `json:"a,string"`
				}
			}{A: &(struct {
				A float64 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadFloat64PtrNotRoot
		{
			name:     "PtrHeadFloat64PtrNotRoot",
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
					A *float64 `json:"a"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)})},
		},
		{
			name:     "PtrHeadFloat64PtrNotRootOmitEmpty",
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
					A *float64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *float64 `json:"a,omitempty"`
			}{A: float64ptr(1)})},
		},
		{
			name:     "PtrHeadFloat64PtrNotRootString",
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
					A *float64 `json:"a,string"`
				}
			}{A: &(struct {
				A *float64 `json:"a,string"`
			}{A: float64ptr(1)})},
		},

		// PtrHeadFloat64PtrNilNotRoot
		{
			name:     "PtrHeadFloat64PtrNilNotRoot",
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
					A *float64 `json:"a"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadFloat64PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *float64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *float64 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadFloat64PtrNilNotRootString",
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
					A *float64 `json:"a,string"`
				}
			}{A: &(struct {
				A *float64 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadFloat64NilNotRoot
		{
			name:     "PtrHeadFloat64NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *float64 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat64NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *float64 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadFloat64NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *float64 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadFloat64ZeroMultiFieldsNotRoot
		{
			name:     "HeadFloat64ZeroMultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadFloat64ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A float64 `json:"a,omitempty"`
				}
				B struct {
					B float64 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadFloat64ZeroMultiFieldsNotRootString",
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
					A float64 `json:"a,string"`
				}
				B struct {
					B float64 `json:"b,string"`
				}
			}{},
		},

		// HeadFloat64MultiFieldsNotRoot
		{
			name:     "HeadFloat64MultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{A: struct {
				A float64 `json:"a"`
			}{A: 1}, B: struct {
				B float64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadFloat64MultiFieldsNotRootOmitEmpty",
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
					A float64 `json:"a,omitempty"`
				}
				B struct {
					B float64 `json:"b,omitempty"`
				}
			}{A: struct {
				A float64 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B float64 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadFloat64MultiFieldsNotRootString",
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
					A float64 `json:"a,string"`
				}
				B struct {
					B float64 `json:"b,string"`
				}
			}{A: struct {
				A float64 `json:"a,string"`
			}{A: 1}, B: struct {
				B float64 `json:"b,string"`
			}{B: 2}},
		},

		// HeadFloat64PtrMultiFieldsNotRoot
		{
			name:     "HeadFloat64PtrMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
				}
				B struct {
					B *float64 `json:"b"`
				}
			}{A: struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)}, B: struct {
				B *float64 `json:"b"`
			}{B: float64ptr(2)}},
		},
		{
			name:     "HeadFloat64PtrMultiFieldsNotRootOmitEmpty",
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
					A *float64 `json:"a,omitempty"`
				}
				B struct {
					B *float64 `json:"b,omitempty"`
				}
			}{A: struct {
				A *float64 `json:"a,omitempty"`
			}{A: float64ptr(1)}, B: struct {
				B *float64 `json:"b,omitempty"`
			}{B: float64ptr(2)}},
		},
		{
			name:     "HeadFloat64PtrMultiFieldsNotRootString",
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
					A *float64 `json:"a,string"`
				}
				B struct {
					B *float64 `json:"b,string"`
				}
			}{A: struct {
				A *float64 `json:"a,string"`
			}{A: float64ptr(1)}, B: struct {
				B *float64 `json:"b,string"`
			}{B: float64ptr(2)}},
		},

		// HeadFloat64PtrNilMultiFieldsNotRoot
		{
			name:     "HeadFloat64PtrNilMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
				}
				B struct {
					B *float64 `json:"b"`
				}
			}{A: struct {
				A *float64 `json:"a"`
			}{A: nil}, B: struct {
				B *float64 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "HeadFloat64PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *float64 `json:"a,omitempty"`
				}
				B struct {
					B *float64 `json:"b,omitempty"`
				}
			}{A: struct {
				A *float64 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *float64 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadFloat64PtrNilMultiFieldsNotRootString",
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
					A *float64 `json:"a,string"`
				}
				B struct {
					B *float64 `json:"b,string"`
				}
			}{A: struct {
				A *float64 `json:"a,string"`
			}{A: nil}, B: struct {
				B *float64 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadFloat64ZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64ZeroMultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadFloat64ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A float64 `json:"a,omitempty"`
				}
				B struct {
					B float64 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadFloat64ZeroMultiFieldsNotRootString",
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
					A float64 `json:"a,string"`
				}
				B struct {
					B float64 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadFloat64MultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64MultiFieldsNotRoot",
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
					A float64 `json:"a"`
				}
				B struct {
					B float64 `json:"b"`
				}
			}{A: struct {
				A float64 `json:"a"`
			}{A: 1}, B: struct {
				B float64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadFloat64MultiFieldsNotRootOmitEmpty",
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
					A float64 `json:"a,omitempty"`
				}
				B struct {
					B float64 `json:"b,omitempty"`
				}
			}{A: struct {
				A float64 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B float64 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadFloat64MultiFieldsNotRootString",
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
					A float64 `json:"a,string"`
				}
				B struct {
					B float64 `json:"b,string"`
				}
			}{A: struct {
				A float64 `json:"a,string"`
			}{A: 1}, B: struct {
				B float64 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadFloat64PtrMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64PtrMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
				}
				B *struct {
					B *float64 `json:"b"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
			}{A: float64ptr(1)}), B: &(struct {
				B *float64 `json:"b"`
			}{B: float64ptr(2)})},
		},
		{
			name:     "PtrHeadFloat64PtrMultiFieldsNotRootOmitEmpty",
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
					A *float64 `json:"a,omitempty"`
				}
				B *struct {
					B *float64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *float64 `json:"a,omitempty"`
			}{A: float64ptr(1)}), B: &(struct {
				B *float64 `json:"b,omitempty"`
			}{B: float64ptr(2)})},
		},
		{
			name:     "PtrHeadFloat64PtrMultiFieldsNotRootString",
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
					A *float64 `json:"a,string"`
				}
				B *struct {
					B *float64 `json:"b,string"`
				}
			}{A: &(struct {
				A *float64 `json:"a,string"`
			}{A: float64ptr(1)}), B: &(struct {
				B *float64 `json:"b,string"`
			}{B: float64ptr(2)})},
		},

		// PtrHeadFloat64PtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a"`
				}
				B *struct {
					B *float64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *float64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *float64 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadFloat64NilMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64NilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a"`
				}
				B *struct {
					B *float64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a,omitempty"`
				}
				B *struct {
					B *float64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a,string"`
				}
				B *struct {
					B *float64 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadFloat64DoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64DoubleMultiFieldsNotRoot",
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
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
				B *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
			}{A: &(struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadFloat64DoubleMultiFieldsNotRootOmitEmpty",
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
					A float64 `json:"a,omitempty"`
					B float64 `json:"b,omitempty"`
				}
				B *struct {
					A float64 `json:"a,omitempty"`
					B float64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A float64 `json:"a,omitempty"`
				B float64 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A float64 `json:"a,omitempty"`
				B float64 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadFloat64DoubleMultiFieldsNotRootString",
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
					A float64 `json:"a,string"`
					B float64 `json:"b,string"`
				}
				B *struct {
					A float64 `json:"a,string"`
					B float64 `json:"b,string"`
				}
			}{A: &(struct {
				A float64 `json:"a,string"`
				B float64 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A float64 `json:"a,string"`
				B float64 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadFloat64NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
				B *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A float64 `json:"a,omitempty"`
					B float64 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A float64 `json:"a,omitempty"`
					B float64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A float64 `json:"a,string"`
					B float64 `json:"b,string"`
				}
				B *struct {
					A float64 `json:"a,string"`
					B float64 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadFloat64NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
				B *struct {
					A float64 `json:"a"`
					B float64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A float64 `json:"a,omitempty"`
					B float64 `json:"b,omitempty"`
				}
				B *struct {
					A float64 `json:"a,omitempty"`
					B float64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A float64 `json:"a,string"`
					B float64 `json:"b,string"`
				}
				B *struct {
					A float64 `json:"a,string"`
					B float64 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadFloat64PtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64PtrDoubleMultiFieldsNotRoot",
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
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
				B *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
			}{A: &(struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(1), B: float64ptr(2)}), B: &(struct {
				A *float64 `json:"a"`
				B *float64 `json:"b"`
			}{A: float64ptr(3), B: float64ptr(4)})},
		},
		{
			name:     "PtrHeadFloat64PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *float64 `json:"a,omitempty"`
					B *float64 `json:"b,omitempty"`
				}
				B *struct {
					A *float64 `json:"a,omitempty"`
					B *float64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *float64 `json:"a,omitempty"`
				B *float64 `json:"b,omitempty"`
			}{A: float64ptr(1), B: float64ptr(2)}), B: &(struct {
				A *float64 `json:"a,omitempty"`
				B *float64 `json:"b,omitempty"`
			}{A: float64ptr(3), B: float64ptr(4)})},
		},
		{
			name:     "PtrHeadFloat64PtrDoubleMultiFieldsNotRootString",
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
					A *float64 `json:"a,string"`
					B *float64 `json:"b,string"`
				}
				B *struct {
					A *float64 `json:"a,string"`
					B *float64 `json:"b,string"`
				}
			}{A: &(struct {
				A *float64 `json:"a,string"`
				B *float64 `json:"b,string"`
			}{A: float64ptr(1), B: float64ptr(2)}), B: &(struct {
				A *float64 `json:"a,string"`
				B *float64 `json:"b,string"`
			}{A: float64ptr(3), B: float64ptr(4)})},
		},

		// PtrHeadFloat64PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
				B *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a,omitempty"`
					B *float64 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *float64 `json:"a,omitempty"`
					B *float64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *float64 `json:"a,string"`
					B *float64 `json:"b,string"`
				}
				B *struct {
					A *float64 `json:"a,string"`
					B *float64 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadFloat64PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
				B *struct {
					A *float64 `json:"a"`
					B *float64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a,omitempty"`
					B *float64 `json:"b,omitempty"`
				}
				B *struct {
					A *float64 `json:"a,omitempty"`
					B *float64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadFloat64PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *float64 `json:"a,string"`
					B *float64 `json:"b,string"`
				}
				B *struct {
					A *float64 `json:"a,string"`
					B *float64 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadFloat64
		{
			name:     "AnonymousHeadFloat64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat64
				B float64 `json:"b"`
			}{
				structFloat64: structFloat64{A: 1},
				B:             2,
			},
		},
		{
			name:     "AnonymousHeadFloat64OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat64OmitEmpty
				B float64 `json:"b,omitempty"`
			}{
				structFloat64OmitEmpty: structFloat64OmitEmpty{A: 1},
				B:                      2,
			},
		},
		{
			name:     "AnonymousHeadFloat64String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structFloat64String
				B float64 `json:"b,string"`
			}{
				structFloat64String: structFloat64String{A: 1},
				B:                   2,
			},
		},

		// PtrAnonymousHeadFloat64
		{
			name:     "PtrAnonymousHeadFloat64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat64
				B float64 `json:"b"`
			}{
				structFloat64: &structFloat64{A: 1},
				B:             2,
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat64OmitEmpty
				B float64 `json:"b,omitempty"`
			}{
				structFloat64OmitEmpty: &structFloat64OmitEmpty{A: 1},
				B:                      2,
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structFloat64String
				B float64 `json:"b,string"`
			}{
				structFloat64String: &structFloat64String{A: 1},
				B:                   2,
			},
		},

		// NilPtrAnonymousHeadFloat64
		{
			name:     "NilPtrAnonymousHeadFloat64",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat64
				B float64 `json:"b"`
			}{
				structFloat64: nil,
				B:             2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat64OmitEmpty
				B float64 `json:"b,omitempty"`
			}{
				structFloat64OmitEmpty: nil,
				B:                      2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structFloat64String
				B float64 `json:"b,string"`
			}{
				structFloat64String: nil,
				B:                   2,
			},
		},

		// AnonymousHeadFloat64Ptr
		{
			name:     "AnonymousHeadFloat64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: structFloat64Ptr{A: float64ptr(1)},
				B:                float64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structFloat64PtrOmitEmpty
				B *float64 `json:"b,omitempty"`
			}{
				structFloat64PtrOmitEmpty: structFloat64PtrOmitEmpty{A: float64ptr(1)},
				B:                         float64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structFloat64PtrString
				B *float64 `json:"b,string"`
			}{
				structFloat64PtrString: structFloat64PtrString{A: float64ptr(1)},
				B:                      float64ptr(2),
			},
		},

		// AnonymousHeadFloat64PtrNil
		{
			name:     "AnonymousHeadFloat64PtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: structFloat64Ptr{A: nil},
				B:                float64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structFloat64PtrOmitEmpty
				B *float64 `json:"b,omitempty"`
			}{
				structFloat64PtrOmitEmpty: structFloat64PtrOmitEmpty{A: nil},
				B:                         float64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structFloat64PtrString
				B *float64 `json:"b,string"`
			}{
				structFloat64PtrString: structFloat64PtrString{A: nil},
				B:                      float64ptr(2),
			},
		},

		// PtrAnonymousHeadFloat64Ptr
		{
			name:     "PtrAnonymousHeadFloat64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: &structFloat64Ptr{A: float64ptr(1)},
				B:                float64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structFloat64PtrOmitEmpty
				B *float64 `json:"b,omitempty"`
			}{
				structFloat64PtrOmitEmpty: &structFloat64PtrOmitEmpty{A: float64ptr(1)},
				B:                         float64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structFloat64PtrString
				B *float64 `json:"b,string"`
			}{
				structFloat64PtrString: &structFloat64PtrString{A: float64ptr(1)},
				B:                      float64ptr(2),
			},
		},

		// NilPtrAnonymousHeadFloat64Ptr
		{
			name:     "NilPtrAnonymousHeadFloat64Ptr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat64Ptr
				B *float64 `json:"b"`
			}{
				structFloat64Ptr: nil,
				B:                float64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structFloat64PtrOmitEmpty
				B *float64 `json:"b,omitempty"`
			}{
				structFloat64PtrOmitEmpty: nil,
				B:                         float64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structFloat64PtrString
				B *float64 `json:"b,string"`
			}{
				structFloat64PtrString: nil,
				B:                      float64ptr(2),
			},
		},

		// AnonymousHeadFloat64Only
		{
			name:     "AnonymousHeadFloat64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat64
			}{
				structFloat64: structFloat64{A: 1},
			},
		},
		{
			name:     "AnonymousHeadFloat64OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat64OmitEmpty
			}{
				structFloat64OmitEmpty: structFloat64OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadFloat64OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structFloat64String
			}{
				structFloat64String: structFloat64String{A: 1},
			},
		},

		// PtrAnonymousHeadFloat64Only
		{
			name:     "PtrAnonymousHeadFloat64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat64
			}{
				structFloat64: &structFloat64{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat64OmitEmpty
			}{
				structFloat64OmitEmpty: &structFloat64OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structFloat64String
			}{
				structFloat64String: &structFloat64String{A: 1},
			},
		},

		// NilPtrAnonymousHeadFloat64Only
		{
			name:     "NilPtrAnonymousHeadFloat64Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64
			}{
				structFloat64: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64OmitEmpty
			}{
				structFloat64OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64String
			}{
				structFloat64String: nil,
			},
		},

		// AnonymousHeadFloat64PtrOnly
		{
			name:     "AnonymousHeadFloat64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat64Ptr
			}{
				structFloat64Ptr: structFloat64Ptr{A: float64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structFloat64PtrOmitEmpty
			}{
				structFloat64PtrOmitEmpty: structFloat64PtrOmitEmpty{A: float64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structFloat64PtrString
			}{
				structFloat64PtrString: structFloat64PtrString{A: float64ptr(1)},
			},
		},

		// AnonymousHeadFloat64PtrNilOnly
		{
			name:     "AnonymousHeadFloat64PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structFloat64Ptr
			}{
				structFloat64Ptr: structFloat64Ptr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structFloat64PtrOmitEmpty
			}{
				structFloat64PtrOmitEmpty: structFloat64PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadFloat64PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structFloat64PtrString
			}{
				structFloat64PtrString: structFloat64PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadFloat64PtrOnly
		{
			name:     "PtrAnonymousHeadFloat64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat64Ptr
			}{
				structFloat64Ptr: &structFloat64Ptr{A: float64ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structFloat64PtrOmitEmpty
			}{
				structFloat64PtrOmitEmpty: &structFloat64PtrOmitEmpty{A: float64ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadFloat64PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structFloat64PtrString
			}{
				structFloat64PtrString: &structFloat64PtrString{A: float64ptr(1)},
			},
		},

		// NilPtrAnonymousHeadFloat64PtrOnly
		{
			name:     "NilPtrAnonymousHeadFloat64PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64Ptr
			}{
				structFloat64Ptr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64PtrOmitEmpty
			}{
				structFloat64PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadFloat64PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structFloat64PtrString
			}{
				structFloat64PtrString: nil,
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
