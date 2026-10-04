package jsontext

import (
	"github.com/goccy/go-json/internal/options"
	"github.com/goccy/go-json/internal/textcoder"
)

func init() {
	textcoder.Attach = attachEncoder
	textcoder.Detach = detachEncoder
	textcoder.Position = encoderPosition
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
	for i := range st.levels {
		l := &st.levels[i]
		tl := textcoder.Level{Object: l.object, Count: l.count}
		if i < len(st.levels)-1 {
			tl.Count-- // the value of the next level, which is open, is not counted
		}
		if l.object && l.last >= l.first {
			for k := l.first; k <= l.last; k++ {
				tl.Names = append(tl.Names, st.names.get(k))
			}
		}
		levels = append(levels, tl)
	}
	k := KindNull
	if st.last().needName() {
		k = KindString
	}
	return levels, e.base + int64(len(e.buf)) + int64(e.delimLen(k))
}

// configureEncoder makes enc write by the options opts until the returned function is called.
func configureEncoder(enc any, opts options.Options) func() {
	e := &enc.(*Encoder).e
	saved := e.cfg
	var c config
	opts.ApplyTo(&c)
	e.cfg = c
	e.derive()
	return func() {
		e.cfg = saved
		e.derive()
	}
}

// attachEncoder makes the Encoder enc write after out, at the place of the output which levels describe.
func attachEncoder(enc any, out []byte, base int64, levels []textcoder.Level, opts options.Options) {
	e := &enc.(*Encoder).e
	var c config
	opts.ApplyTo(&c)
	e.cfg = c
	e.w = nil
	e.init(nil)
	e.buf = out
	e.base = base
	st := &e.st
	for i, l := range levels {
		if i > 0 {
			// the open value of the level which contains it is counted by the push.
			_ = st.push(l.Object)
		}
		cur := st.last()
		if !l.Object {
			cur.count = l.Count
			continue
		}
		for j, name := range l.Names {
			st.insertName(name, e.check)
			cur.count++ // the name
			if j < len(l.Names)-1 || l.Count == int64(2*len(l.Names)) {
				cur.count++ // its value
			}
		}
		// an object may be given without its names, which are not checked then.
		cur.count = l.Count
	}
}

// detachEncoder returns the output of enc and the place where it is.
func detachEncoder(enc any) ([]byte, int, int64) {
	e := &enc.(*Encoder).e
	return e.buf, e.st.depth(), e.st.last().count
}
