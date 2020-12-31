package json

import (
	"bytes"
	"strings"
	"testing"
)

func intptr(v int) *int {
	return &v
}

func int8ptr(v int8) *int8 {
	return &v
}

func TestCoverage(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		data     interface{}
	}{
		{
			name:     "IntHead",
			expected: `{"a":1}`,
			data: struct {
				A int `json:"a"`
			}{A: 1},
		},
		{
			name:     "IntPtrHead",
			expected: `{"a":1}`,
			data: struct {
				A *int `json:"a"`
			}{A: intptr(1)},
		},
		/*
			{
				name:     "IntPtrNilHead",
				expected: `{"a":null}`,
				data: struct {
					A *int `json:"a"`
				}{A: nil},
			},
		*/
		{
			name:     "PtrIntHead",
			expected: `{"a":1}`,
			data: &struct {
				A int `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrIntPtrHead",
			expected: `{"a":1}`,
			data: &struct {
				A *int `json:"a"`
			}{A: intptr(1)},
		},
		{
			name:     "IntField",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int `json:"a"`
				B int `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "IntPtrField",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int `json:"a"`
				B *int `json:"b"`
			}{A: intptr(1), B: intptr(2)},
		},
		{
			name:     "Int8Head",
			expected: `{"a":1}`,
			data: struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "Int8PtrHead",
			expected: `{"a":1}`,
			data: struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		/*
			{
				name:     "Int8PtrNilHead",
				expected: `{"a":null}`,
				data: struct {
					A *int8 `json:"a"`
				}{A: nil},
			},
		*/
		{
			name:     "PtrInt8Head",
			expected: `{"a":1}`,
			data: &struct {
				A int8 `json:"a"`
			}{A: 1},
		},
		{
			name:     "PtrInt8PtrHead",
			expected: `{"a":1}`,
			data: &struct {
				A *int8 `json:"a"`
			}{A: int8ptr(1)},
		},
		{
			name:     "Int8Field",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A int8 `json:"a"`
				B int8 `json:"b"`
			}{A: 1, B: 2},
		},
		{
			name:     "Int8PtrField",
			expected: `{"a":1,"b":2}`,
			data: struct {
				A *int8 `json:"a"`
				B *int8 `json:"b"`
			}{A: int8ptr(1), B: int8ptr(2)},
		},
	}
	for _, test := range tests {
		for _, htmlEscape := range []bool{true, false} {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.SetEscapeHTML(htmlEscape)
			if err := enc.Encode(test.data); err != nil {
				t.Fatalf("%s(htmlEscape:%T): %s: %s", test.name, htmlEscape, test.expected, err)
			}
			if strings.TrimRight(buf.String(), "\n") != test.expected {
				t.Fatalf("%s(htmlEscape:%T): expected %q but got %q", test.name, htmlEscape, test.expected, buf.String())
			}
		}
	}
}
