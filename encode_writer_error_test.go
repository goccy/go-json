package json

import (
	"context"
	"errors"
	"io"
	"testing"
)

// seqWriter returns the next scripted error for each Write. A nil error
// writes the full buffer successfully.
type seqWriter struct {
	errs   []error
	writes int
}

func (w *seqWriter) Write(p []byte) (int, error) {
	w.writes++
	var err error
	if len(w.errs) > 0 {
		err, w.errs = w.errs[0], w.errs[1:]
	}
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

type shortWriter struct {
	writes int
	err    error
}

func (w *shortWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.err != nil {
		return 0, w.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	w.err = io.ErrShortWrite
	return 1, w.err
}

func TestEncoderRetainsWriterError(t *testing.T) {
	want := errors.New("write failed")
	w := &seqWriter{errs: []error{want, nil, nil}}
	enc := NewEncoder(w)
	if err := enc.Encode(1); !errors.Is(err, want) {
		t.Fatalf("Encode: got %v, want %v", err, want)
	}
	if err := enc.Encode(2); !errors.Is(err, want) {
		t.Fatalf("second Encode: got %v, want the original writer error %v", err, want)
	}
	if w.writes != 1 {
		t.Fatalf("Encoder wrote again after a writer error: %d writes", w.writes)
	}

	w = &seqWriter{errs: []error{want, nil}}
	enc = NewEncoder(w)
	if err := enc.EncodeWithOption(1); !errors.Is(err, want) {
		t.Fatalf("EncodeWithOption: got %v, want %v", err, want)
	}
	if err := enc.EncodeWithOption(2); !errors.Is(err, want) {
		t.Fatalf("second EncodeWithOption: got %v, want %v", err, want)
	}
	if w.writes != 1 {
		t.Fatalf("EncodeWithOption wrote again after a writer error: %d writes", w.writes)
	}

	w = &seqWriter{errs: []error{want, nil}}
	enc = NewEncoder(w)
	if err := enc.EncodeContext(context.Background(), 1); !errors.Is(err, want) {
		t.Fatalf("EncodeContext: got %v, want %v", err, want)
	}
	if err := enc.EncodeContext(context.Background(), 2); !errors.Is(err, want) {
		t.Fatalf("second EncodeContext: got %v, want %v", err, want)
	}
	if w.writes != 1 {
		t.Fatalf("EncodeContext wrote again after a writer error: %d writes", w.writes)
	}
}

func TestEncoderRetainsShortWrite(t *testing.T) {
	w := &shortWriter{}
	enc := NewEncoder(w)
	if err := enc.Encode("hello"); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("Encode: got %v, want %v", err, io.ErrShortWrite)
	}
	if err := enc.Encode("world"); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("second Encode: got %v, want %v", err, io.ErrShortWrite)
	}
	if w.writes != 1 {
		t.Fatalf("Encoder wrote again after a short write: %d writes", w.writes)
	}
}
