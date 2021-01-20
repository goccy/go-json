package json_test

import (
	"bytes"
	"testing"

	"github.com/goccy/go-json"
)

func TestCoverString(t *testing.T) {
	type structString struct {
		A string `json:"a"`
	}
	type structStringOmitEmpty struct {
		A string `json:"a,omitempty"`
	}
	type structStringString struct {
		A string `json:"a,string"`
	}

	type structStringPtr struct {
		A *string `json:"a"`
	}
	type structStringPtrOmitEmpty struct {
		A *string `json:"a,omitempty"`
	}
	type structStringPtrString struct {
		A *string `json:"a,string"`
	}

	tests := []struct {
		name           string
		expected       string
		indentExpected string
		data           interface{}
	}{
		// HeadStringZero
		{
			name:     "HeadStringZero",
			expected: `{"a":""}`,
			indentExpected: `
{
  "a": ""
}
`,
			data: struct {
				A string `json:"a"`
			}{},
		},
		{
			name:     "HeadStringZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A string `json:"a,omitempty"`
			}{},
		},
		{
			name:     "HeadStringZeroString",
			expected: `{"a":""}`,
			indentExpected: `
{
  "a": ""
}
`,
			data: struct {
				A string `json:"a,string"`
			}{},
		},

		// HeadString
		{
			name:     "HeadString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				A string `json:"a"`
			}{A: "foo"},
		},
		{
			name:     "HeadStringOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				A string `json:"a,omitempty"`
			}{A: "foo"},
		},
		{
			name:     "HeadStringString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				A string `json:"a,string"`
			}{A: "foo"},
		},

		// HeadStringPtr
		{
			name:     "HeadStringPtr",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				A *string `json:"a"`
			}{A: stringptr("foo")},
		},
		{
			name:     "HeadStringPtrOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				A *string `json:"a,omitempty"`
			}{A: stringptr("foo")},
		},
		{
			name:     "HeadStringPtrString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				A *string `json:"a,string"`
			}{A: stringptr("foo")},
		},

		// HeadStringPtrNil
		{
			name:     "HeadStringPtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *string `json:"a"`
			}{A: nil},
		},
		{
			name:     "HeadStringPtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *string `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "HeadStringPtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				A *string `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadStringZero
		{
			name:     "PtrHeadStringZero",
			expected: `{"a":""}`,
			indentExpected: `
{
  "a": ""
}
`,
			data: &struct {
				A string `json:"a"`
			}{},
		},
		{
			name:     "PtrHeadStringZeroOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A string `json:"a,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadStringZeroString",
			expected: `{"a":""}`,
			indentExpected: `
{
  "a": ""
}
`,
			data: &struct {
				A string `json:"a,string"`
			}{},
		},

		// PtrHeadString
		{
			name:     "PtrHeadString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: &struct {
				A string `json:"a"`
			}{A: "foo"},
		},
		{
			name:     "PtrHeadStringOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: &struct {
				A string `json:"a,omitempty"`
			}{A: "foo"},
		},
		{
			name:     "PtrHeadStringString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: &struct {
				A string `json:"a,string"`
			}{A: "foo"},
		},

		// PtrHeadStringPtr
		{
			name:     "PtrHeadStringPtr",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: &struct {
				A *string `json:"a"`
			}{A: stringptr("foo")},
		},
		{
			name:     "PtrHeadStringPtrOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: &struct {
				A *string `json:"a,omitempty"`
			}{A: stringptr("foo")},
		},
		{
			name:     "PtrHeadStringPtrString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: &struct {
				A *string `json:"a,string"`
			}{A: stringptr("foo")},
		},

		// PtrHeadStringPtrNil
		{
			name:     "PtrHeadStringPtrNil",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *string `json:"a"`
			}{A: nil},
		},
		{
			name:     "PtrHeadStringPtrNilOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *string `json:"a,omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadStringPtrNilString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: &struct {
				A *string `json:"a,string"`
			}{A: nil},
		},

		// PtrHeadStringNil
		{
			name:     "PtrHeadStringNil",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *string `json:"a"`
			})(nil),
		},
		{
			name:     "PtrHeadStringNilOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *string `json:"a,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadStringNilString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *string `json:"a,string"`
			})(nil),
		},

		// HeadStringZeroMultiFields
		{
			name:     "HeadStringZeroMultiFields",
			expected: `{"a":"","b":""}`,
			indentExpected: `
{
  "a": "",
  "b": ""
}
`,
			data: struct {
				A string `json:"a"`
				B string `json:"b"`
			}{},
		},
		{
			name:     "HeadStringZeroMultiFieldsOmitEmpty",
			expected: `{"a":"","b":""}`,
			indentExpected: `
{}
`,
			data: struct {
				A string `json:"a,omitempty"`
				B string `json:"b,omitempty"`
			}{},
		},
		{
			name:     "HeadStringZeroMultiFieldsString",
			expected: `{"a":"","b":""}`,
			indentExpected: `
{
  "a": "",
  "b": ""
}
`,
			data: struct {
				A string `json:"a,string"`
				B string `json:"b,string"`
			}{},
		},

		// HeadStringMultiFields
		{
			name:     "HeadStringMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"},
		},
		{
			name:     "HeadStringMultiFieldsOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				A string `json:"a,omitempty"`
				B string `json:"b,omitempty"`
			}{A: "foo", B: "bar"},
		},
		{
			name:     "HeadStringMultiFieldsString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				A string `json:"a,string"`
				B string `json:"b,string"`
			}{A: "foo", B: "bar"},
		},

		// HeadStringPtrMultiFields
		{
			name:     "HeadStringPtrMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},
		{
			name:     "HeadStringPtrMultiFieldsOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				A *string `json:"a,omitempty"`
				B *string `json:"b,omitempty"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},
		{
			name:     "HeadStringPtrMultiFieldsString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				A *string `json:"a,string"`
				B *string `json:"b,string"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},

		// HeadStringPtrNilMultiFields
		{
			name:     "HeadStringPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadStringPtrNilMultiFieldsOmitEmpty",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{}
`,
			data: struct {
				A *string `json:"a,omitempty"`
				B *string `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "HeadStringPtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: struct {
				A *string `json:"a,string"`
				B *string `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadStringZeroMultiFields
		{
			name:     "PtrHeadStringZeroMultiFields",
			expected: `{"a":"","b":""}`,
			indentExpected: `
{
  "a": "",
  "b": ""
}
`,
			data: &struct {
				A string `json:"a"`
				B string `json:"b"`
			}{},
		},
		{
			name:     "PtrHeadStringZeroMultiFieldsOmitEmpty",
			expected: `{"a":"","b":""}`,
			indentExpected: `
{}
`,
			data: &struct {
				A string `json:"a,omitempty"`
				B string `json:"b,omitempty"`
			}{},
		},
		{
			name:     "PtrHeadStringZeroMultiFieldsString",
			expected: `{"a":"","b":""}`,
			indentExpected: `
{
  "a": "",
  "b": ""
}
`,
			data: &struct {
				A string `json:"a,string"`
				B string `json:"b,string"`
			}{},
		},

		// PtrHeadStringMultiFields
		{
			name:     "PtrHeadStringMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: &struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"},
		},
		{
			name:     "PtrHeadStringMultiFieldsOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: &struct {
				A string `json:"a,omitempty"`
				B string `json:"b,omitempty"`
			}{A: "foo", B: "bar"},
		},
		{
			name:     "PtrHeadStringMultiFieldsString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: &struct {
				A string `json:"a,string"`
				B string `json:"b,string"`
			}{A: "foo", B: "bar"},
		},

		// PtrHeadStringPtrMultiFields
		{
			name:     "PtrHeadStringPtrMultiFields",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: &struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},
		{
			name:     "PtrHeadStringPtrMultiFieldsOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: &struct {
				A *string `json:"a,omitempty"`
				B *string `json:"b,omitempty"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},
		{
			name:     "PtrHeadStringPtrMultiFieldsString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: &struct {
				A *string `json:"a,string"`
				B *string `json:"b,string"`
			}{A: stringptr("foo"), B: stringptr("bar")},
		},

		// PtrHeadStringPtrNilMultiFields
		{
			name:     "PtrHeadStringPtrNilMultiFields",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringPtrNilMultiFieldsOmitEmpty",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *string `json:"a,omitempty"`
				B *string `json:"b,omitempty"`
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringPtrNilMultiFieldsString",
			expected: `{"a":null,"b":null}`,
			indentExpected: `
{
  "a": null,
  "b": null
}
`,
			data: &struct {
				A *string `json:"a,string"`
				B *string `json:"b,string"`
			}{A: nil, B: nil},
		},

		// PtrHeadStringNilMultiFields
		{
			name:     "PtrHeadStringNilMultiFields",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *string `json:"a"`
				B *string `json:"b"`
			})(nil),
		},
		{
			name:     "PtrHeadStringNilMultiFieldsOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *string `json:"a,omitempty"`
				B *string `json:"b,omitempty"`
			})(nil),
		},
		{
			name:     "PtrHeadStringNilMultiFieldsString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *string `json:"a,string"`
				B *string `json:"b,string"`
			})(nil),
		},

		// HeadStringZeroNotRoot
		{
			name:     "HeadStringZeroNotRoot",
			expected: `{"A":{"a":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadStringZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A string `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadStringZeroNotRootString",
			expected: `{"A":{"a":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a,string"`
				}
			}{},
		},

		// HeadStringNotRoot
		{
			name:     "HeadStringNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
			}{A: struct {
				A string `json:"a"`
			}{A: "foo"}},
		},
		{
			name:     "HeadStringNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a,omitempty"`
				}
			}{A: struct {
				A string `json:"a,omitempty"`
			}{A: "foo"}},
		},
		{
			name:     "HeadStringNotRootString",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a,string"`
				}
			}{A: struct {
				A string `json:"a,string"`
			}{A: "foo"}},
		},

		// HeadStringPtrNotRoot
		{
			name:     "HeadStringPtrNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A struct {
					A *string `json:"a"`
				}
			}{A: struct {
				A *string `json:"a"`
			}{stringptr("foo")}},
		},
		{
			name:     "HeadStringPtrNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A struct {
					A *string `json:"a,omitempty"`
				}
			}{A: struct {
				A *string `json:"a,omitempty"`
			}{stringptr("foo")}},
		},
		{
			name:     "HeadStringPtrNotRootString",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A struct {
					A *string `json:"a,string"`
				}
			}{A: struct {
				A *string `json:"a,string"`
			}{stringptr("foo")}},
		},

		// HeadStringPtrNilNotRoot
		{
			name:     "HeadStringPtrNilNotRoot",
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
					A *string `json:"a"`
				}
			}{},
		},
		{
			name:     "HeadStringPtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A struct {
					A *string `json:"a,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadStringPtrNilNotRootString",
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
					A *string `json:"a,string"`
				}
			}{},
		},

		// PtrHeadStringZeroNotRoot
		{
			name:     "PtrHeadStringZeroNotRoot",
			expected: `{"A":{"a":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  }
}
`,
			data: struct {
				A *struct {
					A string `json:"a"`
				}
			}{A: new(struct {
				A string `json:"a"`
			})},
		},
		{
			name:     "PtrHeadStringZeroNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A string `json:"a,omitempty"`
				}
			}{A: new(struct {
				A string `json:"a,omitempty"`
			})},
		},
		{
			name:     "PtrHeadStringZeroNotRootString",
			expected: `{"A":{"a":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  }
}
`,
			data: struct {
				A *struct {
					A string `json:"a,string"`
				}
			}{A: new(struct {
				A string `json:"a,string"`
			})},
		},

		// PtrHeadStringNotRoot
		{
			name:     "PtrHeadStringNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A *struct {
					A string `json:"a"`
				}
			}{A: &(struct {
				A string `json:"a"`
			}{A: "foo"})},
		},
		{
			name:     "PtrHeadStringNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A *struct {
					A string `json:"a,omitempty"`
				}
			}{A: &(struct {
				A string `json:"a,omitempty"`
			}{A: "foo"})},
		},
		{
			name:     "PtrHeadStringNotRootString",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A *struct {
					A string `json:"a,string"`
				}
			}{A: &(struct {
				A string `json:"a,string"`
			}{A: "foo"})},
		},

		// PtrHeadStringPtrNotRoot
		{
			name:     "PtrHeadStringPtrNotRoot",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A *struct {
					A *string `json:"a"`
				}
			}{A: &(struct {
				A *string `json:"a"`
			}{A: stringptr("foo")})},
		},
		{
			name:     "PtrHeadStringPtrNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A *struct {
					A *string `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *string `json:"a,omitempty"`
			}{A: stringptr("foo")})},
		},
		{
			name:     "PtrHeadStringPtrNotRootString",
			expected: `{"A":{"a":"foo"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  }
}
`,
			data: struct {
				A *struct {
					A *string `json:"a,string"`
				}
			}{A: &(struct {
				A *string `json:"a,string"`
			}{A: stringptr("foo")})},
		},

		// PtrHeadStringPtrNilNotRoot
		{
			name:     "PtrHeadStringPtrNilNotRoot",
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
					A *string `json:"a"`
				}
			}{A: &(struct {
				A *string `json:"a"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadStringPtrNilNotRootOmitEmpty",
			expected: `{"A":{}}`,
			indentExpected: `
{
  "A": {}
}
`,
			data: struct {
				A *struct {
					A *string `json:"a,omitempty"`
				}
			}{A: &(struct {
				A *string `json:"a,omitempty"`
			}{A: nil})},
		},
		{
			name:     "PtrHeadStringPtrNilNotRootString",
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
					A *string `json:"a,string"`
				}
			}{A: &(struct {
				A *string `json:"a,string"`
			}{A: nil})},
		},

		// PtrHeadStringNilNotRoot
		{
			name:     "PtrHeadStringNilNotRoot",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *string `json:"a"`
				}
			}{A: nil},
		},
		{
			name:     "PtrHeadStringNilNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				A *struct {
					A *string `json:"a,omitempty"`
				} `json:",omitempty"`
			}{A: nil},
		},
		{
			name:     "PtrHeadStringNilNotRootString",
			expected: `{"A":null}`,
			indentExpected: `
{
  "A": null
}
`,
			data: struct {
				A *struct {
					A *string `json:"a,string"`
				} `json:",string"`
			}{A: nil},
		},

		// HeadStringZeroMultiFieldsNotRoot
		{
			name:     "HeadStringZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":""},"B":{"b":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  },
  "B": {
    "b": ""
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{},
		},
		{
			name:     "HeadStringZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A string `json:"a,omitempty"`
				}
				B struct {
					B string `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "HeadStringZeroMultiFieldsNotRootString",
			expected: `{"A":{"a":""},"B":{"b":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  },
  "B": {
    "b": ""
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a,string"`
				}
				B struct {
					B string `json:"b,string"`
				}
			}{},
		},

		// HeadStringMultiFieldsNotRoot
		{
			name:     "HeadStringMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{A: struct {
				A string `json:"a"`
			}{A: "foo"}, B: struct {
				B string `json:"b"`
			}{B: "bar"}},
		},
		{
			name:     "HeadStringMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,

			data: struct {
				A struct {
					A string `json:"a,omitempty"`
				}
				B struct {
					B string `json:"b,omitempty"`
				}
			}{A: struct {
				A string `json:"a,omitempty"`
			}{A: "foo"}, B: struct {
				B string `json:"b,omitempty"`
			}{B: "bar"}},
		},
		{
			name:     "HeadStringMultiFieldsNotRootString",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: struct {
				A struct {
					A string `json:"a,string"`
				}
				B struct {
					B string `json:"b,string"`
				}
			}{A: struct {
				A string `json:"a,string"`
			}{A: "foo"}, B: struct {
				B string `json:"b,string"`
			}{B: "bar"}},
		},

		// HeadStringPtrMultiFieldsNotRoot
		{
			name:     "HeadStringPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: struct {
				A struct {
					A *string `json:"a"`
				}
				B struct {
					B *string `json:"b"`
				}
			}{A: struct {
				A *string `json:"a"`
			}{A: stringptr("foo")}, B: struct {
				B *string `json:"b"`
			}{B: stringptr("bar")}},
		},
		{
			name:     "HeadStringPtrMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: struct {
				A struct {
					A *string `json:"a,omitempty"`
				}
				B struct {
					B *string `json:"b,omitempty"`
				}
			}{A: struct {
				A *string `json:"a,omitempty"`
			}{A: stringptr("foo")}, B: struct {
				B *string `json:"b,omitempty"`
			}{B: stringptr("bar")}},
		},
		{
			name:     "HeadStringPtrMultiFieldsNotRootString",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: struct {
				A struct {
					A *string `json:"a,string"`
				}
				B struct {
					B *string `json:"b,string"`
				}
			}{A: struct {
				A *string `json:"a,string"`
			}{A: stringptr("foo")}, B: struct {
				B *string `json:"b,string"`
			}{B: stringptr("bar")}},
		},

		// HeadStringPtrNilMultiFieldsNotRoot
		{
			name:     "HeadStringPtrNilMultiFieldsNotRoot",
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
					A *string `json:"a"`
				}
				B struct {
					B *string `json:"b"`
				}
			}{A: struct {
				A *string `json:"a"`
			}{A: nil}, B: struct {
				B *string `json:"b"`
			}{B: nil}},
		},
		{
			name:     "HeadStringPtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{"a":null},"B":{"b":null}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: struct {
				A struct {
					A *string `json:"a,omitempty"`
				}
				B struct {
					B *string `json:"b,omitempty"`
				}
			}{A: struct {
				A *string `json:"a,omitempty"`
			}{A: nil}, B: struct {
				B *string `json:"b,omitempty"`
			}{B: nil}},
		},
		{
			name:     "HeadStringPtrNilMultiFieldsNotRootString",
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
					A *string `json:"a,string"`
				}
				B struct {
					B *string `json:"b,string"`
				}
			}{A: struct {
				A *string `json:"a,string"`
			}{A: nil}, B: struct {
				B *string `json:"b,string"`
			}{B: nil}},
		},

		// PtrHeadStringZeroMultiFieldsNotRoot
		{
			name:     "PtrHeadStringZeroMultiFieldsNotRoot",
			expected: `{"A":{"a":""},"B":{"b":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  },
  "B": {
    "b": ""
  }
}
`,
			data: &struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{},
		},
		{
			name:     "PtrHeadStringZeroMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{},"B":{}}`,
			indentExpected: `
{
  "A": {},
  "B": {}
}
`,
			data: &struct {
				A struct {
					A string `json:"a,omitempty"`
				}
				B struct {
					B string `json:"b,omitempty"`
				}
			}{},
		},
		{
			name:     "PtrHeadStringZeroMultiFieldsNotRootString",
			expected: `{"A":{"a":""},"B":{"b":""}}`,
			indentExpected: `
{
  "A": {
    "a": ""
  },
  "B": {
    "b": ""
  }
}
`,
			data: &struct {
				A struct {
					A string `json:"a,string"`
				}
				B struct {
					B string `json:"b,string"`
				}
			}{},
		},

		// PtrHeadStringMultiFieldsNotRoot
		{
			name:     "PtrHeadStringMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: &struct {
				A struct {
					A string `json:"a"`
				}
				B struct {
					B string `json:"b"`
				}
			}{A: struct {
				A string `json:"a"`
			}{A: "foo"}, B: struct {
				B string `json:"b"`
			}{B: "bar"}},
		},
		{
			name:     "PtrHeadStringMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: &struct {
				A struct {
					A string `json:"a,omitempty"`
				}
				B struct {
					B string `json:"b,omitempty"`
				}
			}{A: struct {
				A string `json:"a,omitempty"`
			}{A: "foo"}, B: struct {
				B string `json:"b,omitempty"`
			}{B: "bar"}},
		},
		{
			name:     "PtrHeadStringMultiFieldsNotRootString",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: &struct {
				A struct {
					A string `json:"a,string"`
				}
				B struct {
					B string `json:"b,string"`
				}
			}{A: struct {
				A string `json:"a,string"`
			}{A: "foo"}, B: struct {
				B string `json:"b,string"`
			}{B: "bar"}},
		},

		// PtrHeadStringPtrMultiFieldsNotRoot
		{
			name:     "PtrHeadStringPtrMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
				}
				B *struct {
					B *string `json:"b"`
				}
			}{A: &(struct {
				A *string `json:"a"`
			}{A: stringptr("foo")}), B: &(struct {
				B *string `json:"b"`
			}{B: stringptr("bar")})},
		},
		{
			name:     "PtrHeadStringPtrMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,omitempty"`
				}
				B *struct {
					B *string `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *string `json:"a,omitempty"`
			}{A: stringptr("foo")}), B: &(struct {
				B *string `json:"b,omitempty"`
			}{B: stringptr("bar")})},
		},
		{
			name:     "PtrHeadStringPtrMultiFieldsNotRootString",
			expected: `{"A":{"a":"foo"},"B":{"b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo"
  },
  "B": {
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,string"`
				}
				B *struct {
					B *string `json:"b,string"`
				}
			}{A: &(struct {
				A *string `json:"a,string"`
			}{A: stringptr("foo")}), B: &(struct {
				B *string `json:"b,string"`
			}{B: stringptr("bar")})},
		},

		// PtrHeadStringPtrNilMultiFieldsNotRoot
		{
			name:     "PtrHeadStringPtrNilMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
				}
				B *struct {
					B *string `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringPtrNilMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,omitempty"`
				}
				B *struct {
					B *string `json:"b,omitempty"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringPtrNilMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,string"`
				}
				B *struct {
					B *string `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadStringNilMultiFieldsNotRoot
		{
			name:     "PtrHeadStringNilMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *string `json:"a"`
				}
				B *struct {
					B *string `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringNilMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *string `json:"a,omitempty"`
				}
				B *struct {
					B *string `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringNilMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *string `json:"a,string"`
				}
				B *struct {
					B *string `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadStringDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadStringDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo",
    "b": "bar"
  },
  "B": {
    "a": "foo",
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
				B *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
			}{A: &(struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"}), B: &(struct {
				A string `json:"a"`
				B string `json:"b"`
			}{A: "foo", B: "bar"})},
		},
		{
			name:     "PtrHeadStringDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo",
    "b": "bar"
  },
  "B": {
    "a": "foo",
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A string `json:"a,omitempty"`
					B string `json:"b,omitempty"`
				}
				B *struct {
					A string `json:"a,omitempty"`
					B string `json:"b,omitempty"`
				}
			}{A: &(struct {
				A string `json:"a,omitempty"`
				B string `json:"b,omitempty"`
			}{A: "foo", B: "bar"}), B: &(struct {
				A string `json:"a,omitempty"`
				B string `json:"b,omitempty"`
			}{A: "foo", B: "bar"})},
		},
		{
			name:     "PtrHeadStringDoubleMultiFieldsNotRootString",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo",
    "b": "bar"
  },
  "B": {
    "a": "foo",
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A string `json:"a,string"`
					B string `json:"b,string"`
				}
				B *struct {
					A string `json:"a,string"`
					B string `json:"b,string"`
				}
			}{A: &(struct {
				A string `json:"a,string"`
				B string `json:"b,string"`
			}{A: "foo", B: "bar"}), B: &(struct {
				A string `json:"a,string"`
				B string `json:"b,string"`
			}{A: "foo", B: "bar"})},
		},

		// PtrHeadStringNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
				B *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A string `json:"a,omitempty"`
					B string `json:"b,omitempty"`
				}
				B *struct {
					A string `json:"a,omitempty"`
					B string `json:"b,omitempty"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A string `json:"a,string"`
					B string `json:"b,string"`
				}
				B *struct {
					A string `json:"a,string"`
					B string `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadStringNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
				B *struct {
					A string `json:"a"`
					B string `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A string `json:"a,omitempty"`
					B string `json:"b,omitempty"`
				}
				B *struct {
					A string `json:"a,omitempty"`
					B string `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A string `json:"a,string"`
					B string `json:"b,string"`
				}
				B *struct {
					A string `json:"a,string"`
					B string `json:"b,string"`
				}
			})(nil),
		},

		// PtrHeadStringPtrDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadStringPtrDoubleMultiFieldsNotRoot",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo",
    "b": "bar"
  },
  "B": {
    "a": "foo",
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
				B *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
			}{A: &(struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")}), B: &(struct {
				A *string `json:"a"`
				B *string `json:"b"`
			}{A: stringptr("foo"), B: stringptr("bar")})},
		},
		{
			name:     "PtrHeadStringPtrDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo",
    "b": "bar"
  },
  "B": {
    "a": "foo",
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,omitempty"`
					B *string `json:"b,omitempty"`
				}
				B *struct {
					A *string `json:"a,omitempty"`
					B *string `json:"b,omitempty"`
				}
			}{A: &(struct {
				A *string `json:"a,omitempty"`
				B *string `json:"b,omitempty"`
			}{A: stringptr("foo"), B: stringptr("bar")}), B: &(struct {
				A *string `json:"a,omitempty"`
				B *string `json:"b,omitempty"`
			}{A: stringptr("foo"), B: stringptr("bar")})},
		},
		{
			name:     "PtrHeadStringPtrDoubleMultiFieldsNotRootString",
			expected: `{"A":{"a":"foo","b":"bar"},"B":{"a":"foo","b":"bar"}}`,
			indentExpected: `
{
  "A": {
    "a": "foo",
    "b": "bar"
  },
  "B": {
    "a": "foo",
    "b": "bar"
  }
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,string"`
					B *string `json:"b,string"`
				}
				B *struct {
					A *string `json:"a,string"`
					B *string `json:"b,string"`
				}
			}{A: &(struct {
				A *string `json:"a,string"`
				B *string `json:"b,string"`
			}{A: stringptr("foo"), B: stringptr("bar")}), B: &(struct {
				A *string `json:"a,string"`
				B *string `json:"b,string"`
			}{A: stringptr("foo"), B: stringptr("bar")})},
		},

		// PtrHeadStringPtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRoot",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
				B *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,omitempty"`
					B *string `json:"b,omitempty"`
				}
				B *struct {
					A *string `json:"a,omitempty"`
					B *string `json:"b,omitempty"`
				}
			}{A: nil, B: nil},
		},
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRootString",
			expected: `{"A":null,"B":null}`,
			indentExpected: `
{
  "A": null,
  "B": null
}
`,
			data: &struct {
				A *struct {
					A *string `json:"a,string"`
					B *string `json:"b,string"`
				}
				B *struct {
					A *string `json:"a,string"`
					B *string `json:"b,string"`
				}
			}{A: nil, B: nil},
		},

		// PtrHeadStringPtrNilDoubleMultiFieldsNotRoot
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRoot",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
				B *struct {
					A *string `json:"a"`
					B *string `json:"b"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRootOmitEmpty",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *string `json:"a,omitempty"`
					B *string `json:"b,omitempty"`
				}
				B *struct {
					A *string `json:"a,omitempty"`
					B *string `json:"b,omitempty"`
				}
			})(nil),
		},
		{
			name:     "PtrHeadStringPtrNilDoubleMultiFieldsNotRootString",
			expected: `null`,
			indentExpected: `
null
`,
			data: (*struct {
				A *struct {
					A *string `json:"a,string"`
					B *string `json:"b,string"`
				}
				B *struct {
					A *string `json:"a,string"`
					B *string `json:"b,string"`
				}
			})(nil),
		},

		// AnonymousHeadString
		{
			name:     "AnonymousHeadString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				structString
				B string `json:"b"`
			}{
				structString: structString{A: "foo"},
				B:            "bar",
			},
		},
		{
			name:     "AnonymousHeadStringOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				structStringOmitEmpty
				B string `json:"b,omitempty"`
			}{
				structStringOmitEmpty: structStringOmitEmpty{A: "foo"},
				B:                     "bar",
			},
		},
		{
			name:     "AnonymousHeadStringString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				structStringString
				B string `json:"b,string"`
			}{
				structStringString: structStringString{A: "foo"},
				B:                  "bar",
			},
		},

		// PtrAnonymousHeadString
		{
			name:     "PtrAnonymousHeadString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				*structString
				B string `json:"b"`
			}{
				structString: &structString{A: "foo"},
				B:            "bar",
			},
		},
		{
			name:     "PtrAnonymousHeadStringOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				*structStringOmitEmpty
				B string `json:"b,omitempty"`
			}{
				structStringOmitEmpty: &structStringOmitEmpty{A: "foo"},
				B:                     "bar",
			},
		},
		{
			name:     "PtrAnonymousHeadStringString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				*structStringString
				B string `json:"b,string"`
			}{
				structStringString: &structStringString{A: "foo"},
				B:                  "bar",
			},
		},

		// NilPtrAnonymousHeadString
		{
			name:     "NilPtrAnonymousHeadString",
			expected: `{"b":"baz"}`,
			indentExpected: `
{
  "b": "baz"
}
`,
			data: struct {
				*structString
				B string `json:"b"`
			}{
				structString: nil,
				B:            "baz",
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringOmitEmpty",
			expected: `{"b":"baz"}`,
			indentExpected: `
{
  "b": "baz"
}
`,
			data: struct {
				*structStringOmitEmpty
				B string `json:"b,omitempty"`
			}{
				structStringOmitEmpty: nil,
				B:                     "baz",
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringString",
			expected: `{"b":"baz"}`,
			indentExpected: `
{
  "b": "baz"
}
`,
			data: struct {
				*structStringString
				B string `json:"b,string"`
			}{
				structStringString: nil,
				B:                  "baz",
			},
		},

		// AnonymousHeadStringPtr
		{
			name:     "AnonymousHeadStringPtr",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: structStringPtr{A: stringptr("foo")},
				B:               stringptr("bar"),
			},
		},
		{
			name:     "AnonymousHeadStringPtrOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				structStringPtrOmitEmpty
				B *string `json:"b,omitempty"`
			}{
				structStringPtrOmitEmpty: structStringPtrOmitEmpty{A: stringptr("foo")},
				B:                        stringptr("bar"),
			},
		},
		{
			name:     "AnonymousHeadStringPtrString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				structStringPtrString
				B *string `json:"b,string"`
			}{
				structStringPtrString: structStringPtrString{A: stringptr("foo")},
				B:                     stringptr("bar"),
			},
		},

		// AnonymousHeadStringPtrNil
		{
			name:     "AnonymousHeadStringPtrNil",
			expected: `{"a":null,"b":"foo"}`,
			indentExpected: `
{
  "a": null,
  "b": "foo"
}
`,
			data: struct {
				structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: structStringPtr{A: nil},
				B:               stringptr("foo"),
			},
		},
		{
			name:     "AnonymousHeadStringPtrNilOmitEmpty",
			expected: `{"b":"foo"}`,
			indentExpected: `
{
  "b": "foo"
}
`,
			data: struct {
				structStringPtrOmitEmpty
				B *string `json:"b,omitempty"`
			}{
				structStringPtrOmitEmpty: structStringPtrOmitEmpty{A: nil},
				B:                        stringptr("foo"),
			},
		},
		{
			name:     "AnonymousHeadStringPtrNilString",
			expected: `{"a":null,"b":"foo"}`,
			indentExpected: `
{
  "a": null,
  "b": "foo"
}
`,
			data: struct {
				structStringPtrString
				B *string `json:"b,string"`
			}{
				structStringPtrString: structStringPtrString{A: nil},
				B:                     stringptr("foo"),
			},
		},

		// PtrAnonymousHeadStringPtr
		{
			name:     "PtrAnonymousHeadStringPtr",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				*structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: &structStringPtr{A: stringptr("foo")},
				B:               stringptr("bar"),
			},
		},
		{
			name:     "PtrAnonymousHeadStringPtrOmitEmpty",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				*structStringPtrOmitEmpty
				B *string `json:"b,omitempty"`
			}{
				structStringPtrOmitEmpty: &structStringPtrOmitEmpty{A: stringptr("foo")},
				B:                        stringptr("bar"),
			},
		},
		{
			name:     "PtrAnonymousHeadStringPtrString",
			expected: `{"a":"foo","b":"bar"}`,
			indentExpected: `
{
  "a": "foo",
  "b": "bar"
}
`,
			data: struct {
				*structStringPtrString
				B *string `json:"b,string"`
			}{
				structStringPtrString: &structStringPtrString{A: stringptr("foo")},
				B:                     stringptr("bar"),
			},
		},

		// NilPtrAnonymousHeadStringPtr
		{
			name:     "NilPtrAnonymousHeadStringPtr",
			expected: `{"b":"foo"}`,
			indentExpected: `
{
  "b": "foo"
}
`,
			data: struct {
				*structStringPtr
				B *string `json:"b"`
			}{
				structStringPtr: nil,
				B:               stringptr("foo"),
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringPtrOmitEmpty",
			expected: `{"b":"foo"}`,
			indentExpected: `
{
  "b": "foo"
}
`,
			data: struct {
				*structStringPtrOmitEmpty
				B *string `json:"b,omitempty"`
			}{
				structStringPtrOmitEmpty: nil,
				B:                        stringptr("foo"),
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringPtrString",
			expected: `{"b":"foo"}`,
			indentExpected: `
{
  "b": "foo"
}
`,
			data: struct {
				*structStringPtrString
				B *string `json:"b,string"`
			}{
				structStringPtrString: nil,
				B:                     stringptr("foo"),
			},
		},

		// AnonymousHeadStringOnly
		{
			name:     "AnonymousHeadStringOnly",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				structString
			}{
				structString: structString{A: "foo"},
			},
		},
		{
			name:     "AnonymousHeadStringOnlyOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				structStringOmitEmpty
			}{
				structStringOmitEmpty: structStringOmitEmpty{A: "foo"},
			},
		},
		{
			name:     "AnonymousHeadStringOnlyString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				structStringString
			}{
				structStringString: structStringString{A: "foo"},
			},
		},

		// PtrAnonymousHeadStringOnly
		{
			name:     "PtrAnonymousHeadStringOnly",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				*structString
			}{
				structString: &structString{A: "foo"},
			},
		},
		{
			name:     "PtrAnonymousHeadStringOnlyOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				*structStringOmitEmpty
			}{
				structStringOmitEmpty: &structStringOmitEmpty{A: "foo"},
			},
		},
		{
			name:     "PtrAnonymousHeadStringOnlyString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				*structStringString
			}{
				structStringString: &structStringString{A: "foo"},
			},
		},

		// NilPtrAnonymousHeadStringOnly
		{
			name:     "NilPtrAnonymousHeadStringOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structString
			}{
				structString: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structStringOmitEmpty
			}{
				structStringOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structStringString
			}{
				structStringString: nil,
			},
		},

		// AnonymousHeadStringPtrOnly
		{
			name:     "AnonymousHeadStringPtrOnly",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				structStringPtr
			}{
				structStringPtr: structStringPtr{A: stringptr("foo")},
			},
		},
		{
			name:     "AnonymousHeadStringPtrOnlyOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				structStringPtrOmitEmpty
			}{
				structStringPtrOmitEmpty: structStringPtrOmitEmpty{A: stringptr("foo")},
			},
		},
		{
			name:     "AnonymousHeadStringPtrOnlyString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				structStringPtrString
			}{
				structStringPtrString: structStringPtrString{A: stringptr("foo")},
			},
		},

		// AnonymousHeadStringPtrNilOnly
		{
			name:     "AnonymousHeadStringPtrNilOnly",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structStringPtr
			}{
				structStringPtr: structStringPtr{A: nil},
			},
		},
		{
			name:     "AnonymousHeadStringPtrNilOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				structStringPtrOmitEmpty
			}{
				structStringPtrOmitEmpty: structStringPtrOmitEmpty{A: nil},
			},
		},
		{
			name:     "AnonymousHeadStringPtrNilOnlyString",
			expected: `{"a":null}`,
			indentExpected: `
{
  "a": null
}
`,
			data: struct {
				structStringPtrString
			}{
				structStringPtrString: structStringPtrString{A: nil},
			},
		},

		// PtrAnonymousHeadStringPtrOnly
		{
			name:     "PtrAnonymousHeadStringPtrOnly",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				*structStringPtr
			}{
				structStringPtr: &structStringPtr{A: stringptr("foo")},
			},
		},
		{
			name:     "PtrAnonymousHeadStringPtrOnlyOmitEmpty",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				*structStringPtrOmitEmpty
			}{
				structStringPtrOmitEmpty: &structStringPtrOmitEmpty{A: stringptr("foo")},
			},
		},
		{
			name:     "PtrAnonymousHeadStringPtrOnlyString",
			expected: `{"a":"foo"}`,
			indentExpected: `
{
  "a": "foo"
}
`,
			data: struct {
				*structStringPtrString
			}{
				structStringPtrString: &structStringPtrString{A: stringptr("foo")},
			},
		},

		// NilPtrAnonymousHeadStringPtrOnly
		{
			name:     "NilPtrAnonymousHeadStringPtrOnly",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structStringPtr
			}{
				structStringPtr: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringPtrOnlyOmitEmpty",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structStringPtrOmitEmpty
			}{
				structStringPtrOmitEmpty: nil,
			},
		},
		{
			name:     "NilPtrAnonymousHeadStringPtrOnlyString",
			expected: `{}`,
			indentExpected: `
{}
`,
			data: struct {
				*structStringPtrString
			}{
				structStringPtrString: nil,
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
			}
		}
	}
}
