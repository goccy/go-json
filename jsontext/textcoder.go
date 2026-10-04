package jsontext

import (
	"unsafe"

	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/internal/textcoder"
)

func init() {
	textcoder.Attach = attachEncoder
	textcoder.Detach = detachEncoder
	textcoder.Position = encoderPosition
	textcoder.Place = encoderPlace
	textcoder.Configure = configureEncoder
	textcoder.Invalidate = func(enc any) {
		e := &enc.(*Encoder).e
		e.invalid = true
		e.vs.st = nil
	}
}

// encoderPosition returns the levels of enc and the offset of its next value.
func encoderPosition(enc any, levels []textcoder.Level) ([]textcoder.Level, int64) {
	e := &enc.(*Encoder).e
	if !e.ready() {
		e.setUp()
	}
	st := &e.st
	if st.outer != nil {
		st = st.full()
	}
	for i := range st.levels {
		l := &st.levels[i]
		tl := textcoder.Level{Object: l.object, Count: l.count}
		if i < len(st.levels)-1 {
			tl.Count-- // the value of the next level, which is open, is not counted
		}
		if l.object && l.last >= l.first {
			tl.Name = st.names.get(l.last)
		}
		levels = append(levels, tl)
	}
	_, _, offset := encoderPlace(enc)
	return levels, offset
}

// encoderPlace returns the innermost level of enc, whether it is the top level, and the offset of its next value,
// without the levels around it.
func encoderPlace(enc any) (textcoder.Level, bool, int64) {
	e := &enc.(*Encoder).e
	if !e.ready() {
		e.setUp()
	}
	l := e.st.last()
	k := KindNull
	if l.needName() {
		k = KindString
	}
	inner := textcoder.Level{Object: l.object, Count: l.count}
	return inner, e.st.depth() == 0, e.base + int64(len(e.buf)) + int64(e.delimLen(k))
}

// configureEncoder makes enc write by the options opts until the returned function is called.
func configureEncoder(enc any, opts options.Options) func() {
	e := &enc.(*Encoder).e
	saved := e.cfg
	e.cfg = config{}
	opts.ApplyTo(&e.cfg)
	e.derive()
	return func() {
		e.cfg = saved
		e.derive()
	}
}

// attachEncoder makes the Encoder enc write after out, at the place whose innermost level is inner ( see
// textcoder.Attach ).
func attachEncoder(enc unsafe.Pointer, out []byte, base int64, inner textcoder.Level, top bool, outer textcoder.Outer, opts *options.Config, same bool) (int, int, int64) {
	e := &(*Encoder)(enc).e
	if same && e.attached && e.ready() {
		// attached again with the same options, as most calls of the methods are: the state of the grammar and of
		// the scanner is reset, as init resets it, without the rest.
		e.invalid = false
		e.st.reset()
		vs := &e.vs
		vs.write, vs.keep = true, false
		vs.run = 0
		vs.wsFrom, vs.wsEnd = 0, 0
		if vs.in != nil || vs.out != nil {
			vs.in, vs.out = nil, nil
		}
	} else {
		e.cfg = *opts
		e.w = nil
		e.init(nil)
		e.attached = true
	}
	// the pointers are written only if they change, as a write of a pointer costs a write barrier while the GC
	// marks: the output is mostly the same array, whose length is set by a slice of itself.
	if unsafe.SliceData(e.buf) == unsafe.SliceData(out) && cap(e.buf) == cap(out) {
		e.buf = e.buf[:len(out)]
	} else {
		e.buf = out
	}
	e.base, e.maxValue = base, 0
	st := &e.st
	if !top {
		_ = st.push(inner.Object)
	}
	// the names of an object, which the value written at its place can't have, are not given.
	st.last().count = inner.Count
	if !same {
		st.outer = outer
	}
	st.outerDepth, st.outerCount = st.depth(), inner.Count
	skip := 0
	if len(out) == 0 && !top {
		k := KindNull
		if st.last().needName() {
			k = KindString
		}
		skip = e.delimLen(k)
		e.base -= int64(skip)
	}
	return skip, st.depth(), inner.Count
}

// detachEncoder returns the output of enc and the place where it is.
func detachEncoder(enc unsafe.Pointer) ([]byte, int, int64) {
	e := &(*Encoder)(enc).e
	return e.buf, e.st.depth(), e.st.last().count
}
