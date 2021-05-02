package iimap

import "testing"

func TestTypeMapSimple(t *testing.T) {
	m := NewTypeMap()
	var i uintptr
	var v interface{}

	const n = 20000

	for i = 2; i < n; i += 2 {
		m.Set(i, i)
	}

	for i = 2; i < n; i += 2 {
		if v = m.Get(i); v != i {
			t.Errorf("didn't get expected value, %v, %v", i, v)
		}
		if v = m.Get(i + 1); v != nil {
			t.Errorf("didn't get expected 'not found' flag")
		}
	}

	if m.Size() != (n-2)/2 {
		t.Errorf("size (%d) is not right, should be %d", m.Size(), n-1)
	}
}
