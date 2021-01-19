package json_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverUint(t *testing.T) {
	type structUint struct {
		A uint `json:"a"`
	}
	type structUintOmitEmpty struct {
		A uint `json:"a,omitempty"`
	}
	type structUintString struct {
		A uint `json:"a,string"`
	}

	type structUintPtr struct {
		A *uint `json:"a"`
	}
	type structUintPtrOmitEmpty struct {
		A *uint `json:"a,omitempty"`
	}
	type structUintPtrString struct {
		A *uint `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadUintZero
		{
			name:     "HeadUintZero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: struct {
				A uint `json:"a"`
			}{},
		},
		{
			name:     "HeadUintZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadUintZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: struct {
				A uint `json:"a,string"`
			}{},
		},

		// HeadUint
		{
			name:     "HeadUint",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint `json:"a"`
			}{A: 1},
		},
		{
			name:     "HeadUintOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A uint `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "HeadUintString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A uint `json:"a,string"`
			}{A: 1},
		},

		// HeadUintPtr
		{
			name:     "HeadUintPtr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint `json:"a"`
			}{A: uptr(1)},
		},
		{
			name:     "HeadUintPtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				A *uint `json:"a,omitempty"`
			}{A: uptr(1)},
		},
		{
			name:     "HeadUintPtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				A *uint `json:"a,string"`
			}{A: uptr(1)},
		},

		// HeadUintPtrNil
		{
			name:     "HeadUintPtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadUintPtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadUintPtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *uint `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUintZero
		{
			name:     "PtrHeadUintZero",
			expected: `{"a":0}`,
			indentExpected: `
{
  "a": 0
}
`,
			data: &struct {
				A uint `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadUintZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUintZeroString",
			expected: `{"a":"0"}`,
			indentExpected: `
{
  "a": "0"
}
`,
			data: &struct {
				A uint `json:"a,string"`
			}{},
		},

		// PtrHeadUint
		{
			name:     "PtrHeadUint",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUintOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A uint `json:"a,omitempty"`
			}{A: 1},
		},
		{
			name:     "PtrHeadUintString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A uint `json:"a,string"`
			}{A: 1},
		},

		// PtrHeadUintPtr
		{
			name:     "PtrHeadUintPtr",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint `json:"a"`
			}{A: uptr(1)},
		},
		{
			name:     "PtrHeadUintPtrOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: &struct {
				A *uint `json:"a,omitempty"`
			}{A: uptr(1)},
		},
		{
			name:     "PtrHeadUintPtrString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: &struct {
				A *uint `json:"a,string"`
			}{A: uptr(1)},
		},

		// PtrHeadUintPtrNil
		{
			name:     "PtrHeadUintPtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUintPtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUintPtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *uint `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadUintNil
		{
			name:     "PtrHeadUintNil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadUintNilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUintNilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint `json:"a,string"`
			})(nil),
		},

		// HeadUintZeroMultiFields
		{
			name:     "HeadUintZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{},
		},
		{
			name:     "HeadUintZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A uint `json:"a,omitempty"`
				B uint `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadUintZeroMultiFields",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: struct {
				A uint `json:"a,string"`
				B uint `json:"b,string"`
			}{},
		},

		// HeadUintMultiFields
		{
			name:     "HeadUintMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUintMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A uint `json:"a,omitempty"`
				B uint `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "HeadUintMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A uint `json:"a,string"`
				B uint `json:"b,string"`
			}{A: 1, B: 2},
		},

		// HeadUintPtrMultiFields
		{
			name:     "HeadUintPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(1), B: uptr(2)},
		},
		{
			name:     "HeadUintPtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				A *uint `json:"a,omitempty"`
				B *uint `json:"b,omitempty"`
			}{A: uptr(1), B: uptr(2)},
		},
		{
			name:     "HeadUintPtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				A *uint `json:"a,string"`
				B *uint `json:"b,string"`
			}{A: uptr(1), B: uptr(2)},
		},

		// HeadUintPtrNilMultiFields
		{
			name:     "HeadUintPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadUintPtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *uint `json:"a,omitempty"`
				B *uint `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadUintPtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *uint `json:"a,string"`
				B *uint `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUintZeroMultiFields
		{
			name:     "PtrHeadUintZeroMultiFields",
			expected: `{"a":0,"b":0}`,
			indentExpected: `
{
  "a": 0,
  "b": 0
}
`,
			data: &struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadUintZeroMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A uint `json:"a,omitempty"`
				B uint `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadUintZeroMultiFieldsString",
			expected: `{"a":"0","b":"0"}`,
			indentExpected: `
{
  "a": "0",
  "b": "0"
}
`,
			data: &struct {
				A uint `json:"a,string"`
				B uint `json:"b,string"`
			}{},
		},

		// PtrHeadUintMultiFields
		{
			name:     "PtrHeadUintMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUintMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A uint `json:"a,omitempty"`
				B uint `json:"b,omitempty"`
			}{A: 1, B: 2},
		},
		{
			name:     "PtrHeadUintMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A uint `json:"a,string"`
				B uint `json:"b,string"`
			}{A: 1, B: 2},
		},

		// PtrHeadUintPtrMultiFields
		{
			name:     "PtrHeadUintPtrMultiFields",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(1), B: uptr(2)},
		},
		{
			name:     "PtrHeadUintPtrMultiFieldsOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: &struct {
				A *uint `json:"a,omitempty"`
				B *uint `json:"b,omitempty"`
			}{A: uptr(1), B: uptr(2)},
		},
		{
			name:     "PtrHeadUintPtrMultiFieldsString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: &struct {
				A *uint `json:"a,string"`
				B *uint `json:"b,string"`
			}{A: uptr(1), B: uptr(2)},
		},

		// PtrHeadUintPtrNilMultiFields
		{
			name:     "PtrHeadUintPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintPtrNilMultiFieldsOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *uint `json:"a,omitempty"`
				B *uint `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintPtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *uint `json:"a,string"`
				B *uint `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUintNilMultiFields
		{
			name:     "PtrHeadUintNilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadUintNilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint `json:"a,omitempty"`
				B *uint `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadUintNilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *uint `json:"a,string"`
				B *uint `json:"b,string"`
			})(nil),
		},

		// HeadUintZeroNotRoot
		{
			name:     "HeadUintZeroNotRoot",
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
					A uint `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUintZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A uint `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUintZeroNotRootString",
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
					A uint `json:"a,string"`
				}
			}{},
		},

		// HeadUintNotRoot
		{
			name:     "HeadUintNotRoot",
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
					A uint `json:"a"`
				}
			}{A: struct {
				A uint `json:"a"`
			}{A: 1}},
		},
		{
			name:     "HeadUintNotRootOmitEmpty",
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
					A uint `json:"a,omitempty"`
				}
			}{A: struct {
				A uint `json:"a,omitempty"`
			}{A: 1}},
		},
		{
			name:     "HeadUintNotRootString",
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
					A uint `json:"a,string"`
				}
			}{A: struct {
				A uint `json:"a,string"`
			}{A: 1}},
		},

		// HeadUintPtrNotRoot
		{
			name:     "HeadUintPtrNotRoot",
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
					A *uint `json:"a"`
				}
			}{A: struct {
				A *uint `json:"a"`
			}{uptr(1)}},
		},
		{
			name:     "HeadUintPtrNotRootOmitEmpty",
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
					A *uint `json:"a,omitempty"`
				}
			}{A: struct {
				A *uint `json:"a,omitempty"`
			}{uptr(1)}},
		},
		{
			name:     "HeadUintPtrNotRootString",
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
					A *uint `json:"a,string"`
				}
			}{A: struct {
				A *uint `json:"a,string"`
			}{uptr(1)}},
		},

		// HeadUintPtrNilNotRoot
		{
			name:     "HeadUintPtrNilNotRoot",
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
					A *uint `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadUintPtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *uint `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUintPtrNilNotRootString",
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
					A *uint `json:"a,string"`
				}
			}{},
		},

		// PtrHeadUintZeroNotRoot
		{
			name:     "PtrHeadUintZeroNotRoot",
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
					A uint `json:"a"`
				}
			}{A: new(struct {
				A uint `json:"a"`
			})},
		},
		{
			name:     "PtrHeadUintZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A uint `json:"a,omitempty"`
				}
			}{A: new(struct {
				A uint `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadUintZeroNotRootString",
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
					A uint `json:"a,string"`
				}
			}{A: new(struct {
				A uint `json:"a,string"`
			})},
		},

		// PtrHeadUintNotRoot
		{
			name:     "PtrHeadUintNotRoot",
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
					A uint `json:"a"`
				}
			}{A: &(struct {
				A uint `json:"a"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUintNotRootOmitEmpty",
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
					A uint `json:"a,omitempty"`
				}
			}{A: &(struct {
				A uint `json:"a,omitempty"`
			}{A: 1})},
		},
		{
			name:     "PtrHeadUintNotRootString",
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
					A uint `json:"a,string"`
				}
			}{A: &(struct {
				A uint `json:"a,string"`
			}{A: 1})},
		},

		// PtrHeadUintPtrNotRoot
		{
			name:     "PtrHeadUintPtrNotRoot",
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
					A *uint `json:"a"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
			}{A: uptr(1)})},
		},
		{
			name:     "PtrHeadUintPtrNotRootOmitEmpty",
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
					A *uint `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint `json:"a,omitempty"`
			}{A: uptr(1)})},
		},
		{
			name:     "PtrHeadUintPtrNotRootString",
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
					A *uint `json:"a,string"`
				}
			}{A: &(struct {
				A *uint `json:"a,string"`
			}{A: uptr(1)})},
		},

		// PtrHeadUintPtrNilNotRoot
		{
			name:     "PtrHeadUintPtrNilNotRoot",
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
					A *uint `json:"a"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUintPtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *uint `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *uint `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadUintPtrNilNotRootString",
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
					A *uint `json:"a,string"`
				}
			}{A: &(struct {
				A *uint `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadUintNilNotRoot
		{
			name:     "PtrHeadUintNilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadUintNilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *uint `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadUintNilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *uint `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadUintZeroMultiFieldsNotRoot
		{
			name:     "HeadUintZeroMultiFieldsNotRoot",
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
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadUintZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A uint `json:"a,omitempty"`
				}
				B struct {
					B uint `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadUintZeroMultiFieldsNotRootString",
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
					A uint `json:"a,string"`
				}
				B struct {
					B uint `json:"b,string"`
				}
			}{},
		},

		// HeadUintMultiFieldsNotRoot
		{
			name:     "HeadUintMultiFieldsNotRoot",
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
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{A: struct {
				A uint `json:"a"`
			}{A: 1}, B: struct {
				B uint `json:"b"`
			}{B: 2}},
		},
		{
			name:     "HeadUintMultiFieldsNotRootOmitEmpty",
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
					A uint `json:"a,omitempty"`
				}
				B struct {
					B uint `json:"b,omitempty"`
				}
			}{A: struct {
				A uint `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "HeadUintMultiFieldsNotRootString",
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
					A uint `json:"a,string"`
				}
				B struct {
					B uint `json:"b,string"`
				}
			}{A: struct {
				A uint `json:"a,string"`
			}{A: 1}, B: struct {
				B uint `json:"b,string"`
			}{B: 2}},
		},

		// HeadUintPtrMultiFieldsNotRoot
		{
			name:     "HeadUintPtrMultiFieldsNotRoot",
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
					A *uint `json:"a"`
				}
				B struct {
					B *uint `json:"b"`
				}
			}{A: struct {
				A *uint `json:"a"`
			}{A: uptr(1)}, B: struct {
				B *uint `json:"b"`
			}{B: uptr(2)}},
		},
		{
			name:     "HeadUintPtrMultiFieldsNotRootOmitEmpty",
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
					A *uint `json:"a,omitempty"`
				}
				B struct {
					B *uint `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint `json:"a,omitempty"`
			}{A: uptr(1)}, B: struct {
				B *uint `json:"b,omitempty"`
			}{B: uptr(2)}},
		},
		{
			name:     "HeadUintPtrMultiFieldsNotRootString",
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
					A *uint `json:"a,string"`
				}
				B struct {
					B *uint `json:"b,string"`
				}
			}{A: struct {
				A *uint `json:"a,string"`
			}{A: uptr(1)}, B: struct {
				B *uint `json:"b,string"`
			}{B: uptr(2)}},
		},

		// HeadUintPtrNilMultiFieldsNotRoot
		{
			name:     "HeadUintPtrNilMultiFieldsNotRoot",
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
					A *uint `json:"a"`
				}
				B struct {
					B *uint `json:"b"`
				}
			}{A: struct {
				A *uint `json:"a"`
			}{A: nil}, B: struct {
				B *uint `json:"b"`
			}{B: nil}},
		},
		{
			name:     "HeadUintPtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *uint `json:"a,omitempty"`
				}
				B struct {
					B *uint `json:"b,omitempty"`
				}
			}{A: struct {
				A *uint `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *uint `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadUintPtrNilMultiFieldsNotRootString",
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
					A *uint `json:"a,string"`
				}
				B struct {
					B *uint `json:"b,string"`
				}
			}{A: struct {
				A *uint `json:"a,string"`
			}{A: nil}, B: struct {
				B *uint `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadUintZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadUintZeroMultiFieldsNotRoot",
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
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadUintZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A uint `json:"a,omitempty"`
				}
				B struct {
					B uint `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadUintZeroMultiFieldsNotRootString",
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
					A uint `json:"a,string"`
				}
				B struct {
					B uint `json:"b,string"`
				}
			}{},
		},

		// PtrHeadUintMultiFieldsNotRoot
		{
			name:     "PtrHeadUintMultiFieldsNotRoot",
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
					A uint `json:"a"`
				}
				B struct {
					B uint `json:"b"`
				}
			}{A: struct {
				A uint `json:"a"`
			}{A: 1}, B: struct {
				B uint `json:"b"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUintMultiFieldsNotRootOmitEmpty",
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
					A uint `json:"a,omitempty"`
				}
				B struct {
					B uint `json:"b,omitempty"`
				}
			}{A: struct {
				A uint `json:"a,omitempty"`
			}{A: 1}, B: struct {
				B uint `json:"b,omitempty"`
			}{B: 2}},
		},
		{
			name:     "PtrHeadUintMultiFieldsNotRootString",
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
					A uint `json:"a,string"`
				}
				B struct {
					B uint `json:"b,string"`
				}
			}{A: struct {
				A uint `json:"a,string"`
			}{A: 1}, B: struct {
				B uint `json:"b,string"`
			}{B: 2}},
		},

		// PtrHeadUintPtrMultiFieldsNotRoot
		{
			name:     "PtrHeadUintPtrMultiFieldsNotRoot",
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
					A *uint `json:"a"`
				}
				B *struct {
					B *uint `json:"b"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
			}{A: uptr(1)}), B: &(struct {
				B *uint `json:"b"`
			}{B: uptr(2)})},
		},
		{
			name:     "PtrHeadUintPtrMultiFieldsNotRootOmitEmpty",
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
					A *uint `json:"a,omitempty"`
				}
				B *struct {
					B *uint `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint `json:"a,omitempty"`
			}{A: uptr(1)}), B: &(struct {
				B *uint `json:"b,omitempty"`
			}{B: uptr(2)})},
		},
		{
			name:     "PtrHeadUintPtrMultiFieldsNotRootString",
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
					A *uint `json:"a,string"`
				}
				B *struct {
					B *uint `json:"b,string"`
				}
			}{A: &(struct {
				A *uint `json:"a,string"`
			}{A: uptr(1)}), B: &(struct {
				B *uint `json:"b,string"`
			}{B: uptr(2)})},
		},

		// PtrHeadUintPtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadUintPtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint `json:"a"`
				}
				B *struct {
					B *uint `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintPtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint `json:"a,omitempty"`
				} `json:",omitempty"`
				B *struct {
					B *uint `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintPtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint `json:"a,string"`
				} `json:",string"`
				B *struct {
					B *uint `json:"b,string"`
				} `json:",string"`
			}{A: nil, B: nil},
		},

		// PtrHeadUintNilMultiFieldsNotRoot
		{
			name:     "PtrHeadUintNilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint `json:"a"`
				}
				B *struct {
					B *uint `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintNilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint `json:"a,omitempty"`
				}
				B *struct {
					B *uint `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintNilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint `json:"a,string"`
				}
				B *struct {
					B *uint `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUintDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUintDoubleMultiFieldsNotRoot",
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
					A uint `json:"a"`
					B uint `json:"b"`
				}
				B *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
			}{A: &(struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 1, B: 2}), B: &(struct {
				A uint `json:"a"`
				B uint `json:"b"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUintDoubleMultiFieldsNotRootOmitEmpty",
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
					A uint `json:"a,omitempty"`
					B uint `json:"b,omitempty"`
				}
				B *struct {
					A uint `json:"a,omitempty"`
					B uint `json:"b,omitempty"`
				}
			}{A: &(struct {
				A uint `json:"a,omitempty"`
				B uint `json:"b,omitempty"`
			}{A: 1, B: 2}), B: &(struct {
				A uint `json:"a,omitempty"`
				B uint `json:"b,omitempty"`
			}{A: 3, B: 4})},
		},
		{
			name:     "PtrHeadUintDoubleMultiFieldsNotRootString",
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
					A uint `json:"a,string"`
					B uint `json:"b,string"`
				}
				B *struct {
					A uint `json:"a,string"`
					B uint `json:"b,string"`
				}
			}{A: &(struct {
				A uint `json:"a,string"`
				B uint `json:"b,string"`
			}{A: 1, B: 2}), B: &(struct {
				A uint `json:"a,string"`
				B uint `json:"b,string"`
			}{A: 3, B: 4})},
		},

		// PtrHeadUintNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
				B *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A uint `json:"a,omitempty"`
					B uint `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A uint `json:"a,omitempty"`
					B uint `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A uint `json:"a,string"`
					B uint `json:"b,string"`
				}
				B *struct {
					A uint `json:"a,string"`
					B uint `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUintNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
				B *struct {
					A uint `json:"a"`
					B uint `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint `json:"a,omitempty"`
					B uint `json:"b,omitempty"`
				}
				B *struct {
					A uint `json:"a,omitempty"`
					B uint `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A uint `json:"a,string"`
					B uint `json:"b,string"`
				}
				B *struct {
					A uint `json:"a,string"`
					B uint `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadUintPtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUintPtrDoubleMultiFieldsNotRoot",
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
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
				B *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
			}{A: &(struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(1), B: uptr(2)}), B: &(struct {
				A *uint `json:"a"`
				B *uint `json:"b"`
			}{A: uptr(3), B: uptr(4)})},
		},
		{
			name:     "PtrHeadUintPtrDoubleMultiFieldsNotRootOmitEmpty",
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
					A *uint `json:"a,omitempty"`
					B *uint `json:"b,omitempty"`
				}
				B *struct {
					A *uint `json:"a,omitempty"`
					B *uint `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *uint `json:"a,omitempty"`
				B *uint `json:"b,omitempty"`
			}{A: uptr(1), B: uptr(2)}), B: &(struct {
				A *uint `json:"a,omitempty"`
				B *uint `json:"b,omitempty"`
			}{A: uptr(3), B: uptr(4)})},
		},
		{
			name:     "PtrHeadUintPtrDoubleMultiFieldsNotRootString",
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
					A *uint `json:"a,string"`
					B *uint `json:"b,string"`
				}
				B *struct {
					A *uint `json:"a,string"`
					B *uint `json:"b,string"`
				}
			}{A: &(struct {
				A *uint `json:"a,string"`
				B *uint `json:"b,string"`
			}{A: uptr(1), B: uptr(2)}), B: &(struct {
				A *uint `json:"a,string"`
				B *uint `json:"b,string"`
			}{A: uptr(3), B: uptr(4)})},
		},

		// PtrHeadUintPtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
				B *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *uint `json:"a,omitempty"`
					B *uint `json:"b,omitempty"`
				} `json:",omitempty"`
				B *struct {
					A *uint `json:"a,omitempty"`
					B *uint `json:"b,omitempty"`
				} `json:",omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *uint `json:"a,string"`
					B *uint `json:"b,string"`
				}
				B *struct {
					A *uint `json:"a,string"`
					B *uint `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadUintPtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
				B *struct {
					A *uint `json:"a"`
					B *uint `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint `json:"a,omitempty"`
					B *uint `json:"b,omitempty"`
				}
				B *struct {
					A *uint `json:"a,omitempty"`
					B *uint `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadUintPtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *uint `json:"a,string"`
					B *uint `json:"b,string"`
				}
				B *struct {
					A *uint `json:"a,string"`
					B *uint `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadUint
		{
			name:     "AnonymousHeadUint",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUint
				B uint `json:"b"`
			}{
				structUint: structUint{A: 1},
				B:          2,
			},
		},
		{
			name:     "AnonymousHeadUintOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUintOmitEmpty
				B uint `json:"b,omitempty"`
			}{
				structUintOmitEmpty: structUintOmitEmpty{A: 1},
				B:                   2,
			},
		},
		{
			name:     "AnonymousHeadUintString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUintString
				B uint `json:"b,string"`
			}{
				structUintString: structUintString{A: 1},
				B:                2,
			},
		},

		// PtrAnonymousHeadUint
		{
			name:     "PtrAnonymousHeadUint",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUint
				B uint `json:"b"`
			}{
				structUint: &structUint{A: 1},
				B:          2,
			},
		},
		{
			name:     "PtrAnonymousHeadUintOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUintOmitEmpty
				B uint `json:"b,omitempty"`
			}{
				structUintOmitEmpty: &structUintOmitEmpty{A: 1},
				B:                   2,
			},
		},
		{
			name:     "PtrAnonymousHeadUintString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUintString
				B uint `json:"b,string"`
			}{
				structUintString: &structUintString{A: 1},
				B:                2,
			},
		},

		// NilPtrAnonymousHeadUint
		{
			name:     "NilPtrAnonymousHeadUint",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUint
				B uint `json:"b"`
			}{
				structUint: nil,
				B:          2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUintOmitEmpty
				B uint `json:"b,omitempty"`
			}{
				structUintOmitEmpty: nil,
				B:                   2,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUintString
				B uint `json:"b,string"`
			}{
				structUintString: nil,
				B:                2,
			},
		},

		// AnonymousHeadUintPtr
		{
			name:     "AnonymousHeadUintPtr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: structUintPtr{A: uptr(1)},
				B:             uptr(2),
			},
		},
		{
			name:     "AnonymousHeadUintPtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				structUintPtrOmitEmpty
				B *uint `json:"b,omitempty"`
			}{
				structUintPtrOmitEmpty: structUintPtrOmitEmpty{A: uptr(1)},
				B:                      uptr(2),
			},
		},
		{
			name:     "AnonymousHeadUintPtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				structUintPtrString
				B *uint `json:"b,string"`
			}{
				structUintPtrString: structUintPtrString{A: uptr(1)},
				B:                   uptr(2),
			},
		},

		// AnonymousHeadUintPtrNil
		{
			name:     "AnonymousHeadUintPtrNil",
			expected: `{"a":null,"b":2}`,
			indentExpected: `
{
  "a": null,
  "b": 2
}
`,
			data: struct {
				structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: structUintPtr{A: nil},
				B:             uptr(2),
			},
		},
		{
			name:     "AnonymousHeadUintPtrNilOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				structUintPtrOmitEmpty
				B *uint `json:"b,omitempty"`
			}{
				structUintPtrOmitEmpty: structUintPtrOmitEmpty{A: nil},
				B:                      uptr(2),
			},
		},
		{
			name:     "AnonymousHeadUintPtrNilString",
			expected: `{"a":null,"b":"2"}`,
			indentExpected: `
{
  "a": null,
  "b": "2"
}
`,
			data: struct {
				structUintPtrString
				B *uint `json:"b,string"`
			}{
				structUintPtrString: structUintPtrString{A: nil},
				B:                   uptr(2),
			},
		},

		// PtrAnonymousHeadUintPtr
		{
			name:     "PtrAnonymousHeadUintPtr",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: &structUintPtr{A: uptr(1)},
				B:             uptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUintPtrOmitEmpty",
			expected: `{"a":1,"b":2}`,
			indentExpected: `
{
  "a": 1,
  "b": 2
}
`,
			data: struct {
				*structUintPtrOmitEmpty
				B *uint `json:"b,omitempty"`
			}{
				structUintPtrOmitEmpty: &structUintPtrOmitEmpty{A: uptr(1)},
				B:                      uptr(2),
			},
		},
		{
			name:     "PtrAnonymousHeadUintPtrString",
			expected: `{"a":"1","b":"2"}`,
			indentExpected: `
{
  "a": "1",
  "b": "2"
}
`,
			data: struct {
				*structUintPtrString
				B *uint `json:"b,string"`
			}{
				structUintPtrString: &structUintPtrString{A: uptr(1)},
				B:                   uptr(2),
			},
		},

		// NilPtrAnonymousHeadUintPtr
		{
			name:     "NilPtrAnonymousHeadUintPtr",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUintPtr
				B *uint `json:"b"`
			}{
				structUintPtr: nil,
				B:             uptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintPtrOmitEmpty",
			expected: `{"b":2}`,
			indentExpected: `
{
  "b": 2
}
`,
			data: struct {
				*structUintPtrOmitEmpty
				B *uint `json:"b,omitempty"`
			}{
				structUintPtrOmitEmpty: nil,
				B:                      uptr(2),
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintPtrString",
			expected: `{"b":"2"}`,
			indentExpected: `
{
  "b": "2"
}
`,
			data: struct {
				*structUintPtrString
				B *uint `json:"b,string"`
			}{
				structUintPtrString: nil,
				B:                   uptr(2),
			},
		},

		// AnonymousHeadUintOnly
		{
			name:     "AnonymousHeadUintOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUint
			}{
				structUint: structUint{A: 1},
			},
		},
		{
			name:     "AnonymousHeadUintOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUintOmitEmpty
			}{
				structUintOmitEmpty: structUintOmitEmpty{A: 1},
			},
		},
		{
			name:     "AnonymousHeadUintOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUintString
			}{
				structUintString: structUintString{A: 1},
			},
		},

		// PtrAnonymousHeadUintOnly
		{
			name:     "PtrAnonymousHeadUintOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUint
			}{
				structUint: &structUint{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUintOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUintOmitEmpty
			}{
				structUintOmitEmpty: &structUintOmitEmpty{A: 1},
			},
		},
		{
			name:     "PtrAnonymousHeadUintOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUintString
			}{
				structUintString: &structUintString{A: 1},
			},
		},

		// NilPtrAnonymousHeadUintOnly
		{
			name:     "NilPtrAnonymousHeadUintOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUint
			}{
				structUint: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUintOmitEmpty
			}{
				structUintOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUintString
			}{
				structUintString: nil,
			},
		},

		// AnonymousHeadUintPtrOnly
		{
			name:     "AnonymousHeadUintPtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUintPtr
			}{
				structUintPtr: structUintPtr{A: uptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUintPtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				structUintPtrOmitEmpty
			}{
				structUintPtrOmitEmpty: structUintPtrOmitEmpty{A: uptr(1)},
			},
		},
		{
			name:     "AnonymousHeadUintPtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				structUintPtrString
			}{
				structUintPtrString: structUintPtrString{A: uptr(1)},
			},
		},

		// AnonymousHeadUintPtrNilOnly
		{
			name:     "AnonymousHeadUintPtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUintPtr
			}{
				structUintPtr: structUintPtr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadUintPtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structUintPtrOmitEmpty
			}{
				structUintPtrOmitEmpty: structUintPtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadUintPtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structUintPtrString
			}{
				structUintPtrString: structUintPtrString{A: nil},
			},
		},

		// PtrAnonymousHeadUintPtrOnly
		{
			name:     "PtrAnonymousHeadUintPtrOnly",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUintPtr
			}{
				structUintPtr: &structUintPtr{A: uptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadUintPtrOnlyOmitEmpty",
			expected: `{"a":1}`,
			indentExpected: `
{
  "a": 1
}
`,
			data: struct {
				*structUintPtrOmitEmpty
			}{
				structUintPtrOmitEmpty: &structUintPtrOmitEmpty{A: uptr(1)},
			},
		},
		{
			name:     "PtrAnonymousHeadUintPtrOnlyString",
			expected: `{"a":"1"}`,
			indentExpected: `
{
  "a": "1"
}
`,
			data: struct {
				*structUintPtrString
			}{
				structUintPtrString: &structUintPtrString{A: uptr(1)},
			},
		},

		// NilPtrAnonymousHeadUintPtrOnly
		{
			name:     "NilPtrAnonymousHeadUintPtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUintPtr
			}{
				structUintPtr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintPtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUintPtrOmitEmpty
			}{
				structUintPtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadUintPtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structUintPtrString
			}{
				structUintPtrString: nil,
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
