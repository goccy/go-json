package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverUint64(t *testing.T) {
	type structUint64 struct {
		A uint64 `json:"a"`
	}
	type structUint64OmitEmpty struct {
		A uint64 `json:"a,omitempty"`
	}
	type structUint64String struct {
		A uint64 `json:"a,string"`
	}

	type structUint64Ptr struct {
		A *uint64 `json:"a"`
	}
	type structUint64PtrOmitEmpty struct {
		A *uint64 `json:"a,omitempty"`
	}
	type structUint64PtrString struct {
		A *uint64 `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadUint64Zero
		{
			name:     "HeadUint64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A uint64 `json:"a"`
			}{},
		},
		{
			name:     "HeadUint64ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint64 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadUint64ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A uint64 `json:"a,string"`
			}{},
		},

		// HeadUint64
		{
			name:     "HeadUint64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUint64OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint64 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadUint64String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A uint64 `json:"a,string"`
			}{A: 1},
		},

		// HeadUint64Ptr
		{
			name:     "HeadUint64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "HeadUint64PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint64 `json:"a,omitempty"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "HeadUint64PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *uint64 `json:"a,string"`
			}{A: uint64ptr(1)},
		},

		// HeadUint64PtrNil
		{
			name:     "HeadUint64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadUint64PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint64 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadUint64PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint64 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint64Zero
		{
			name:     "PtrHeadUint64Zero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A uint64 `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint64 `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A uint64 `json:"a,string"`
			}{},
		},

		// PtrHeadUint64
		{
			name:     "PtrHeadUint64",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint64 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint64OmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint64 `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUint64String",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A uint64 `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadUint64Ptr
		{
			name:     "PtrHeadUint64Ptr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "PtrHeadUint64PtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint64 `json:"a,omitempty"`
			}{A: uint64ptr(1)},
		},
		{
			name:     "PtrHeadUint64PtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *uint64 `json:"a,string"`
			}{A: uint64ptr(1)},
		},

		// PtrHeadUint64PtrNil
		{
			name:     "PtrHeadUint64PtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint64 `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint64 `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint64 `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUint64Nil
		{
			name:     "PtrHeadUint64Nil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a,string"`
			})(nil),
		},

		// HeadUint64ZeroMultiFields
		{
			name:     "HeadUint64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{},
		},
		{
			name:     "HeadUint64ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint64 `json:"a,omitempty"`
				B uint64 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadUint64ZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A uint64 `json:"a,string"`
				B uint64 `json:"b,string"`
			}{},
		},

		// HeadUint64MultiFields
		{
			name:     "HeadUint64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint64MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint64 `json:"a,omitempty"`
				B uint64 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUint64MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A uint64 `json:"a,string"`
				B uint64 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadUint64PtrMultiFields
		{
			name:     "HeadUint64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "HeadUint64PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint64 `json:"a,omitempty"`
				B *uint64 `json:"b,omitempty"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "HeadUint64PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *uint64 `json:"a,string"`
				B *uint64 `json:"b,string"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},

		// HeadUint64PtrNilMultiFields
		{
			name:     "HeadUint64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadUint64PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint64 `json:"a,omitempty"`
				B *uint64 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadUint64PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint64 `json:"a,string"`
				B *uint64 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint64ZeroMultiFields
		{
			name:     "PtrHeadUint64ZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint64 `json:"a,omitempty"`
				B uint64 `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A uint64 `json:"a,string"`
				B uint64 `json:"b,string"`
			}{},
		},

		// PtrHeadUint64MultiFields
		{
			name:     "PtrHeadUint64MultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint64MultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint64 `json:"a,omitempty"`
				B uint64 `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUint64MultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A uint64 `json:"a,string"`
				B uint64 `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadUint64PtrMultiFields
		{
			name:     "PtrHeadUint64PtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "PtrHeadUint64PtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint64 `json:"a,omitempty"`
				B *uint64 `json:"b,omitempty"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},
		{
			name:     "PtrHeadUint64PtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *uint64 `json:"a,string"`
				B *uint64 `json:"b,string"`
			}{A: uint64ptr(1), B: uint64ptr(2)},
		},

		// PtrHeadUint64PtrNilMultiFields
		{
			name:     "PtrHeadUint64PtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint64 `json:"a,omitempty"`
				B *uint64 `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint64 `json:"a,string"`
				B *uint64 `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint64NilMultiFields
		{
			name:     "PtrHeadUint64NilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a,omitempty"`
				B *uint64 `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint64 `json:"a,string"`
				B *uint64 `json:"b,string"`
			})(nil),
		},

		// HeadUint64ZeroNotRoot
		{
			name:     "HeadUint64ZeroNotRoot",
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
					A uint64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint64ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A uint64 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint64ZeroNotRootString",
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
					A uint64 `json:"a,string"`
				}
			}{},
		},

		// HeadUint64NotRoot
		{
			name:     "HeadUint64NotRoot",
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
					A uint64 `json:"a"`
				}
			}{A: struct {
				A uint64 `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUint64NotRootOmitEmpty",
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
					A uint64 `json:"a,omitempty"`
				}
			}{A: struct {
				A uint64 `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadUint64NotRootString",
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
					A uint64 `json:"a,string"`
				}
			}{A: struct {
				A uint64 `json:"a,string"`
			}{A: 1}},
		},

		// HeadUint64PtrNotRoot
		{
			name:     "HeadUint64PtrNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{A: struct {
				A *uint64 `json:"a"`
			}{uint64ptr(1)}},
		},
		{
			name:     "HeadUint64PtrNotRootOmitEmpty",
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
					A *uint64 `json:"a,omitempty"`
				}
			}{A: struct {
				A *uint64 `json:"a,omitempty"`
			}{uint64ptr(1)}},
		},
		{
			name:     "HeadUint64PtrNotRootString",
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
					A *uint64 `json:"a,string"`
				}
			}{A: struct {
				A *uint64 `json:"a,string"`
			}{uint64ptr(1)}},
		},

		// HeadUint64PtrNilNotRoot
		{
			name:     "HeadUint64PtrNilNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUint64PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *uint64 `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint64PtrNilNotRootString",
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
					A *uint64 `json:"a,string"`
				}
			}{},
		},

		// PtrHeadUint64ZeroNotRoot
		{
			name:     "PtrHeadUint64ZeroNotRoot",
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
					A uint64 `json:"a"`
				}
			}{A: new(struct {
				A uint64 `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUint64ZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A uint64 `json:"a,omitempty"`
				}
			}{A: new(struct {
				A uint64 `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadUint64ZeroNotRootString",
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
					A uint64 `json:"a,string"`
				}
			}{A: new(struct {
				A uint64 `json:"a,string"`
			})},
		},

		// PtrHeadUint64NotRoot
		{
			name:     "PtrHeadUint64NotRoot",
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
					A uint64 `json:"a"`
				}
			}{A: &(struct {
				A uint64 `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint64NotRootOmitEmpty",
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
					A uint64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A uint64 `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUint64NotRootString",
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
					A uint64 `json:"a,string"`
				}
			}{A: &(struct {
				A uint64 `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadUint64PtrNotRoot
		{
			name:     "PtrHeadUint64PtrNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)})},
		},
		{
			name:     "PtrHeadUint64PtrNotRootOmitEmpty",
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
					A *uint64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,omitempty"`
			}{A: uint64ptr(1)})},
		},
		{
			name:     "PtrHeadUint64PtrNotRootString",
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
					A *uint64 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,string"`
			}{A: uint64ptr(1)})},
		},

		// PtrHeadUint64PtrNilNotRoot
		{
			name:     "PtrHeadUint64PtrNilNotRoot",
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
					A *uint64 `json:"a"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint64PtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *uint64 `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUint64PtrNilNotRootString",
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
					A *uint64 `json:"a,string"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadUint64NilNotRoot
		{
			name:     "PtrHeadUint64NilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint64 `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64NilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *uint64 `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUint64NilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint64 `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadUint64ZeroMultiFieldsNotRoot
		{
			name:     "HeadUint64ZeroMultiFieldsNotRoot",
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
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadUint64ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A uint64 `json:"a,omitempty"`
				}
				B struct {
					B uint64 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUint64ZeroMultiFieldsNotRootString",
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
					A uint64 `json:"a,string"`
				}
				B struct {
					B uint64 `json:"b,string"`
				}
			}{},
		},

		// HeadUint64MultiFieldsNotRoot
		{
			name:     "HeadUint64MultiFieldsNotRoot",
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
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{A: struct {
				A uint64 `json:"a"`
			}{A: 1}, B: struct {
				B uint64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadUint64MultiFieldsNotRootOmitEmpty",
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
					A uint64 `json:"a,omitempty"`
				}
				B struct {
					B uint64 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint64 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint64 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadUint64MultiFieldsNotRootString",
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
					A uint64 `json:"a,string"`
				}
				B struct {
					B uint64 `json:"b,string"`
				}
			}{A: struct {
				A uint64 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint64 `json:"b,string"`
			}{B: 2}},
		},

		// HeadUint64PtrMultiFieldsNotRoot
		{
			name:     "HeadUint64PtrMultiFieldsNotRoot",
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
					A *uint64 `json:"a"`
				}
				B struct {
					B *uint64 `json:"b"`
				}
			}{A: struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)}, B: struct {
				B *uint64 `json:"b"`
			}{B: uint64ptr(2)}},
		},
		{
			name:     "HeadUint64PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint64 `json:"a,omitempty"`
				}
				B struct {
					B *uint64 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint64 `json:"a,omitempty"`
			}{A: uint64ptr(1)}, B: struct {
				B *uint64 `json:"b,omitempty"`
			}{B: uint64ptr(2)}},
		},
		{
			name:     "HeadUint64PtrMultiFieldsNotRootString",
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
					A *uint64 `json:"a,string"`
				}
				B struct {
					B *uint64 `json:"b,string"`
				}
			}{A: struct {
				A *uint64 `json:"a,string"`
			}{A: uint64ptr(1)}, B: struct {
				B *uint64 `json:"b,string"`
			}{B: uint64ptr(2)}},
		},

		// HeadUint64PtrNilMultiFieldsNotRoot
		{
			name:     "HeadUint64PtrNilMultiFieldsNotRoot",
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
					A *uint64 `json:"a"`
				}
				B struct {
					B *uint64 `json:"b"`
				}
			}{A: struct {
				A *uint64 `json:"a"`
			}{A: nil}, B: struct {
				B *uint64 `json:"b"`
			}{B: nil}},
		},
		{
			name:     "HeadUint64PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *uint64 `json:"a,omitempty"`
				}
				B struct {
					B *uint64 `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint64 `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *uint64 `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadUint64PtrNilMultiFieldsNotRootString",
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
					A *uint64 `json:"a,string"`
				}
				B struct {
					B *uint64 `json:"b,string"`
				}
			}{A: struct {
				A *uint64 `json:"a,string"`
			}{A: nil}, B: struct {
				B *uint64 `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadUint64ZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64ZeroMultiFieldsNotRoot",
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
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A uint64 `json:"a,omitempty"`
				}
				B struct {
					B uint64 `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadUint64ZeroMultiFieldsNotRootString",
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
					A uint64 `json:"a,string"`
				}
				B struct {
					B uint64 `json:"b,string"`
				}
			}{},
		},

		// PtrHeadUint64MultiFieldsNotRoot
		{
			name:     "PtrHeadUint64MultiFieldsNotRoot",
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
					A uint64 `json:"a"`
				}
				B struct {
					B uint64 `json:"b"`
				}
			}{A: struct {
				A uint64 `json:"a"`
			}{A: 1}, B: struct {
				B uint64 `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint64MultiFieldsNotRootOmitEmpty",
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
					A uint64 `json:"a,omitempty"`
				}
				B struct {
					B uint64 `json:"b,omitempty"`
				}
			}{A: struct {
				A uint64 `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint64 `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUint64MultiFieldsNotRootString",
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
					A uint64 `json:"a,string"`
				}
				B struct {
					B uint64 `json:"b,string"`
				}
			}{A: struct {
				A uint64 `json:"a,string"`
			}{A: 1}, B: struct {
				B uint64 `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadUint64PtrMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64PtrMultiFieldsNotRoot",
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
					A *uint64 `json:"a"`
				}
				B *struct {
					B *uint64 `json:"b"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
			}{A: uint64ptr(1)}), B: &(struct {
				B *uint64 `json:"b"`
			}{B: uint64ptr(2)})},
		},
		{
			name:     "PtrHeadUint64PtrMultiFieldsNotRootOmitEmpty",
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
					A *uint64 `json:"a,omitempty"`
				}
				B *struct {
					B *uint64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,omitempty"`
			}{A: uint64ptr(1)}), B: &(struct {
				B *uint64 `json:"b,omitempty"`
			}{B: uint64ptr(2)})},
		},
		{
			name:     "PtrHeadUint64PtrMultiFieldsNotRootString",
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
					A *uint64 `json:"a,string"`
				}
				B *struct {
					B *uint64 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,string"`
			}{A: uint64ptr(1)}), B: &(struct {
				B *uint64 `json:"b,string"`
			}{B: uint64ptr(2)})},
		},

		// PtrHeadUint64PtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64PtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a"`
				}
				B *struct {
					B *uint64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *uint64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *uint64 `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUint64NilMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64NilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a"`
				}
				B *struct {
					B *uint64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a,omitempty"`
				}
				B *struct {
					B *uint64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a,string"`
				}
				B *struct {
					B *uint64 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint64DoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64DoubleMultiFieldsNotRoot",
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
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
				B *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
			}{A: &(struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A uint64 `json:"a"`
				B uint64 `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint64DoubleMultiFieldsNotRootOmitEmpty",
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
					A uint64 `json:"a,omitempty"`
					B uint64 `json:"b,omitempty"`
				}
				B *struct {
					A uint64 `json:"a,omitempty"`
					B uint64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A uint64 `json:"a,omitempty"`
				B uint64 `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A uint64 `json:"a,omitempty"`
				B uint64 `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUint64DoubleMultiFieldsNotRootString",
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
					A uint64 `json:"a,string"`
					B uint64 `json:"b,string"`
				}
				B *struct {
					A uint64 `json:"a,string"`
					B uint64 `json:"b,string"`
				}
			}{A: &(struct {
				A uint64 `json:"a,string"`
				B uint64 `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A uint64 `json:"a,string"`
				B uint64 `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadUint64NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
				B *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A uint64 `json:"a,omitempty"`
					B uint64 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A uint64 `json:"a,omitempty"`
					B uint64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A uint64 `json:"a,string"`
					B uint64 `json:"b,string"`
				}
				B *struct {
					A uint64 `json:"a,string"`
					B uint64 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint64NilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
				B *struct {
					A uint64 `json:"a"`
					B uint64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint64 `json:"a,omitempty"`
					B uint64 `json:"b,omitempty"`
				}
				B *struct {
					A uint64 `json:"a,omitempty"`
					B uint64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64NilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint64 `json:"a,string"`
					B uint64 `json:"b,string"`
				}
				B *struct {
					A uint64 `json:"a,string"`
					B uint64 `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUint64PtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64PtrDoubleMultiFieldsNotRoot",
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
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
				B *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
			}{A: &(struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(1), B: uint64ptr(2)}), B: &(struct {
				A *uint64 `json:"a"`
				B *uint64 `json:"b"`
			}{A: uint64ptr(3), B: uint64ptr(4)})},
		},
		{
			name:     "PtrHeadUint64PtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *uint64 `json:"a,omitempty"`
					B *uint64 `json:"b,omitempty"`
				}
				B *struct {
					A *uint64 `json:"a,omitempty"`
					B *uint64 `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,omitempty"`
				B *uint64 `json:"b,omitempty"`
			}{A: uint64ptr(1), B: uint64ptr(2)}), B: &(struct {
				A *uint64 `json:"a,omitempty"`
				B *uint64 `json:"b,omitempty"`
			}{A: uint64ptr(3), B: uint64ptr(4)})},
		},
		{
			name:     "PtrHeadUint64PtrDoubleMultiFieldsNotRootString",
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
					A *uint64 `json:"a,string"`
					B *uint64 `json:"b,string"`
				}
				B *struct {
					A *uint64 `json:"a,string"`
					B *uint64 `json:"b,string"`
				}
			}{A: &(struct {
				A *uint64 `json:"a,string"`
				B *uint64 `json:"b,string"`
			}{A: uint64ptr(1), B: uint64ptr(2)}), B: &(struct {
				A *uint64 `json:"a,string"`
				B *uint64 `json:"b,string"`
			}{A: uint64ptr(3), B: uint64ptr(4)})},
		},

		// PtrHeadUint64PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
				B *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a,omitempty"`
					B *uint64 `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *uint64 `json:"a,omitempty"`
					B *uint64 `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint64 `json:"a,string"`
					B *uint64 `json:"b,string"`
				}
				B *struct {
					A *uint64 `json:"a,string"`
					B *uint64 `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUint64PtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
				B *struct {
					A *uint64 `json:"a"`
					B *uint64 `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a,omitempty"`
					B *uint64 `json:"b,omitempty"`
				}
				B *struct {
					A *uint64 `json:"a,omitempty"`
					B *uint64 `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUint64PtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint64 `json:"a,string"`
					B *uint64 `json:"b,string"`
				}
				B *struct {
					A *uint64 `json:"a,string"`
					B *uint64 `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadUint64
		{
			name:     "AnonymousHeadUint64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint64
				B uint64 `json:"b"`
			}{
				structUint64: structUint64{A: 1},
				B:            2,
			},
		},
		{
			name:     "AnonymousHeadUint64OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint64OmitEmpty
				B uint64 `json:"b,omitempty"`
			}{
				structUint64OmitEmpty: structUint64OmitEmpty{A: 1},
				B:                     2,
			},
		},
		{
			name:     "AnonymousHeadUint64String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint64String
				B uint64 `json:"b,string"`
			}{
				structUint64String: structUint64String{A: 1},
				B:                  2,
			},
		},

		// PtrAnonymousHeadUint64
		{
			name:     "PtrAnonymousHeadUint64",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint64
				B uint64 `json:"b"`
			}{
				structUint64: &structUint64{A: 1},
				B:            2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint64OmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint64OmitEmpty
				B uint64 `json:"b,omitempty"`
			}{
				structUint64OmitEmpty: &structUint64OmitEmpty{A: 1},
				B:                     2,
			},
		},
		{
			name:     "PtrAnonymousHeadUint64String",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint64String
				B uint64 `json:"b,string"`
			}{
				structUint64String: &structUint64String{A: 1},
				B:                  2,
			},
		},

		// NilPtrAnonymousHeadUint64
		{
			name:     "NilPtrAnonymousHeadUint64",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint64
				B uint64 `json:"b"`
			}{
				structUint64: nil,
				B:            2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64OmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint64OmitEmpty
				B uint64 `json:"b,omitempty"`
			}{
				structUint64OmitEmpty: nil,
				B:                     2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64String",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint64String
				B uint64 `json:"b,string"`
			}{
				structUint64String: nil,
				B:                  2,
			},
		},

		// AnonymousHeadUint64Ptr
		{
			name:     "AnonymousHeadUint64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: structUint64Ptr{A: uint64ptr(1)},
				B:               uint64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint64PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint64PtrOmitEmpty
				B *uint64 `json:"b,omitempty"`
			}{
				structUint64PtrOmitEmpty: structUint64PtrOmitEmpty{A: uint64ptr(1)},
				B:                        uint64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint64PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUint64PtrString
				B *uint64 `json:"b,string"`
			}{
				structUint64PtrString: structUint64PtrString{A: uint64ptr(1)},
				B:                     uint64ptr(2),
			},
		},

		// AnonymousHeadUint64PtrNil
		{
			name:     "AnonymousHeadUint64PtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: structUint64Ptr{A: nil},
				B:               uint64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint64PtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structUint64PtrOmitEmpty
				B *uint64 `json:"b,omitempty"`
			}{
				structUint64PtrOmitEmpty: structUint64PtrOmitEmpty{A: nil},
				B:                        uint64ptr(2),
			},
		},
		{
			name:     "AnonymousHeadUint64PtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structUint64PtrString
				B *uint64 `json:"b,string"`
			}{
				structUint64PtrString: structUint64PtrString{A: nil},
				B:                     uint64ptr(2),
			},
		},

		// PtrAnonymousHeadUint64Ptr
		{
			name:     "PtrAnonymousHeadUint64Ptr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: &structUint64Ptr{A: uint64ptr(1)},
				B:               uint64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint64PtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint64PtrOmitEmpty
				B *uint64 `json:"b,omitempty"`
			}{
				structUint64PtrOmitEmpty: &structUint64PtrOmitEmpty{A: uint64ptr(1)},
				B:                        uint64ptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUint64PtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUint64PtrString
				B *uint64 `json:"b,string"`
			}{
				structUint64PtrString: &structUint64PtrString{A: uint64ptr(1)},
				B:                     uint64ptr(2),
			},
		},

		// NilPtrAnonymousHeadUint64Ptr
		{
			name:     "NilPtrAnonymousHeadUint64Ptr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint64Ptr
				B *uint64 `json:"b"`
			}{
				structUint64Ptr: nil,
				B:               uint64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64PtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint64PtrOmitEmpty
				B *uint64 `json:"b,omitempty"`
			}{
				structUint64PtrOmitEmpty: nil,
				B:                        uint64ptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64PtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUint64PtrString
				B *uint64 `json:"b,string"`
			}{
				structUint64PtrString: nil,
				B:                     uint64ptr(2),
			},
		},

		// AnonymousHeadUint64Only
		{
			name:     "AnonymousHeadUint64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint64
			}{
				structUint64: structUint64{A: 1},
			},
		},
		{
			name:     "AnonymousHeadUint64OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint64OmitEmpty
			}{
				structUint64OmitEmpty: structUint64OmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadUint64OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint64String
			}{
				structUint64String: structUint64String{A: 1},
			},
		},

		// PtrAnonymousHeadUint64Only
		{
			name:     "PtrAnonymousHeadUint64Only",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint64
			}{
				structUint64: &structUint64{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64OnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint64OmitEmpty
			}{
				structUint64OmitEmpty: &structUint64OmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64OnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint64String
			}{
				structUint64String: &structUint64String{A: 1},
			},
		},

		// NilPtrAnonymousHeadUint64Only
		{
			name:     "NilPtrAnonymousHeadUint64Only",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64
			}{
				structUint64: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64OnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64OmitEmpty
			}{
				structUint64OmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64OnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64String
			}{
				structUint64String: nil,
			},
		},

		// AnonymousHeadUint64PtrOnly
		{
			name:     "AnonymousHeadUint64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint64Ptr
			}{
				structUint64Ptr: structUint64Ptr{A: uint64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint64PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint64PtrOmitEmpty
			}{
				structUint64PtrOmitEmpty: structUint64PtrOmitEmpty{A: uint64ptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUint64PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUint64PtrString
			}{
				structUint64PtrString: structUint64PtrString{A: uint64ptr(1)},
			},
		},

		// AnonymousHeadUint64PtrNilOnly
		{
			name:     "AnonymousHeadUint64PtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint64Ptr
			}{
				structUint64Ptr: structUint64Ptr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadUint64PtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structUint64PtrOmitEmpty
			}{
				structUint64PtrOmitEmpty: structUint64PtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadUint64PtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUint64PtrString
			}{
				structUint64PtrString: structUint64PtrString{A: nil},
			},
		},

		// PtrAnonymousHeadUint64PtrOnly
		{
			name:     "PtrAnonymousHeadUint64PtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint64Ptr
			}{
				structUint64Ptr: &structUint64Ptr{A: uint64ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64PtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint64PtrOmitEmpty
			}{
				structUint64PtrOmitEmpty: &structUint64PtrOmitEmpty{A: uint64ptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadUint64PtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUint64PtrString
			}{
				structUint64PtrString: &structUint64PtrString{A: uint64ptr(1)},
			},
		},

		// NilPtrAnonymousHeadUint64PtrOnly
		{
			name:     "NilPtrAnonymousHeadUint64PtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64Ptr
			}{
				structUint64Ptr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64PtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64PtrOmitEmpty
			}{
				structUint64PtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUint64PtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint64PtrString
			}{
				structUint64PtrString: nil,
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
