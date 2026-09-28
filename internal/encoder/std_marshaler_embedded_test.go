package encoder

import (
	"reflect"
	"testing"
	"time"
)

type embeddedTime struct{ time.Time }

type embeddedTimePointer struct{ *time.Time }

type declaredTime struct{ time.Time }

func (declaredTime) MarshalJSON() ([]byte, error) { return []byte(`"t"`), nil }

// A type which has the MarshalJSON of time.Time only by embedding it is trusted, and written from the embedded
// value by AppendText when the value is in it; a type which declares the method is checked.
func TestEmbeddedStdMarshalerCall(t *testing.T) {
	tests := []struct {
		typ          reflect.Type
		trusted      bool
		appendOutput bool
	}{
		{reflect.TypeOf(embeddedTime{}), true, true},
		{reflect.TypeOf(&embeddedTime{}), true, true},
		{reflect.TypeOf(embeddedTimePointer{}), true, false},
		{reflect.TypeOf(declaredTime{}), false, false},
	}
	// AppendText is of Go 1.24: time.Time.MarshalJSON is called before it.
	hasAppendText := stdJSONAppender(reflect.TypeOf(time.Time{})) != nil
	for _, test := range tests {
		m := newMarshalerCall(test.typ, marshalJSONType)
		if m == nil || m.trusted != test.trusted || (m.appendOutput != nil) != (test.appendOutput && hasAppendText) {
			t.Errorf("%v: got %+v, want trusted %v and appendOutput %v", test.typ, m, test.trusted, test.appendOutput)
		}
	}
	text := newMarshalerCall(reflect.TypeOf(embeddedTime{}), marshalTextType)
	if text == nil || (text.appendOutput != nil) != hasAppendText {
		t.Fatal("the text of an embedded time.Time is not appended by AppendText")
	}
}
