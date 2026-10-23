package json

import (
	"io"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/goccy/go-json/internal/decoder"
	"github.com/goccy/go-json/internal/options"
	iruntime "github.com/goccy/go-json/internal/runtime"
	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// UnmarshalRead deserializes a Go value from an io.Reader according to the provided unmarshal and decode options
// (while ignoring marshal or encode options). The input must be a single JSON value with optional whitespace
// interspersed. It consumes the entirety of io.Reader until io.EOF is encountered, without reporting an error for
// EOF. The output must be a non-nil pointer. See Unmarshal for details about the conversion of JSON into a Go value.
//
// The input is read 512 bytes at a time. A value which the bytes read hold whole is decoded as Unmarshal decodes it,
// and an object or an array which they don't is decoded member by member and element by element as it is read, so
// that the memory which the input takes doesn't grow with its size. Only a token longer than a read, and a value
// which a method or a function reads whole, as UnmarshalJSON, UnmarshalText, a function of UnmarshalFunc or a
// jsontext.Value, are held whole, as encoding/json/v2 holds them.
func UnmarshalRead(in io.Reader, out any, opts ...Options) error {
	e := (*emptyInterface)(unsafe.Pointer(&out))
	s := decoder.TakeStreamV2(in)
	err := unmarshalStream(s, e.typ, e.ptr, opts)
	decoder.ReleaseStreamV2(s)
	// the decoder refers to the value by its address only.
	runtime.KeepAlive(out)
	return err
}

// UnmarshalReadOf decodes the input of an io.Reader into the value out points to, as UnmarshalRead does, through a
// value in the heap which is reused, as UnmarshalOf does.
func UnmarshalReadOf[T any](in io.Reader, out *T, opts ...Options) error {
	ptrType := reflect.TypeFor[*T]()
	if out == nil {
		return &SemanticError{action: "unmarshal", GoType: ptrType, Err: errNonNilReference}
	}
	s := decoder.TakeStreamV2(in)
	p := reusedValue(s.Context(), out)
	err := unmarshalStream(s, iruntime.TypePtr(ptrType), p, opts)
	returnValue(p, out)
	decoder.ReleaseStreamV2(s)
	return err
}

// unmarshalStream decodes the input of the stream, a single value, into the value at p, of the pointer type typ, by
// the options.
func unmarshalStream(s *decoder.Stream, typ, p unsafe.Pointer, opts []Options) error {
	ctx := s.Context()
	cfg := beginCall(ctx, opts)
	err := decodeStream(s, typ, p, cfg)
	endCall(ctx, cfg)
	return err
}

// decodeStream decodes the input of the stream, a single value, into the value at p, of the pointer type typ, by the
// options of cfg, which the options of the context of the stream are made of.
func decodeStream(s *decoder.Stream, typ, p unsafe.Pointer, cfg *options.Config) error {
	dec, err := outputDecoder(s.Context(), typ, p)
	if err != nil {
		return err
	}
	if err = s.DecodeV2(dec, p, 0); err == nil {
		var ended bool
		if ended, err = s.AfterTopLevel(); err == nil && !ended {
			// more than white space after the value
			buf, base := s.Read()
			pl := s.Context().V2Call().Place
			return textcoder.AfterTopLevelError(buf[pl.From:pl.End], base+pl.From, cfg)
		}
	}
	switch err {
	case nil:
		return nil
	case s.ReadError():
		return textcoder.ReadError(err)
	case io.EOF:
		// no value, which is the unexpected end of the input
		buf, base := s.Read()
		return &jsontext.SyntacticError{ByteOffset: base + int64(len(buf)), Err: io.ErrUnexpectedEOF}
	}
	return streamError(s, err, cfg)
}

// streamError returns err, the error of the decoding of the stream, as encoding/json/v2 reports it: it is found again
// in the part of the buffer which its place has, read from its place in the innermost level of the path ( see
// decoder.StreamPlace ).
func streamError(s *decoder.Stream, err error, cfg *options.Config) error {
	switch err.(type) {
	case *SemanticError, *jsontext.SyntacticError:
		// the error of a decoder attached at a value, at its place
		return err
	}
	if textcoder.IsReadError(err) {
		return err
	}
	call := s.Context().V2Call()
	pl := call.Place
	if !pl.Set {
		return &SemanticError{action: "unmarshal", Err: err}
	}
	buf, base := s.Read()
	path := call.Path
	l := path[len(path)-1]
	att := &textcoder.DecoderAttachment{
		In: buf[pl.From:pl.End], Base: base + pl.From, Object: l.Object, Top: len(path) == 1, Count: l.Count,
		Outer: &streamPath{path: path}, OuterDepth: len(path) - 1, Opts: cfg,
	}
	return unmarshalError(&errorInput{buf: buf, base: base, att: att}, err, cfg)
}

// streamPath gives the levels of the path of a stream ( see textcoder.Outer ).
type streamPath struct {
	path   []textcoder.Level
	levels []textcoder.Level
}

func (p *streamPath) Levels() []textcoder.Level {
	p.levels = pathLevels(p.levels, p.path)
	return p.levels
}
