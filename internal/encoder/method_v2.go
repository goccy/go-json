package encoder

import (
	"encoding"
	"encoding/json"
	stderrors "errors"
	"reflect"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

// The methods of marshaling of the semantics of encoding/json/v2. A value is always addressable for them: the
// method of the pointer to a type is called for a value of the type, at its address. The first of them which a
// type has writes its value: MarshalJSONTo, MarshalJSON, AppendText, MarshalText.

// V2Hooks are what the v2 json package gives the encoder: the functions which take jsontext, which the encoder
// doesn't import.
var V2Hooks struct {
	// MarshalerToType is the interface of MarshalJSONTo.
	MarshalerToType reflect.Type
	// FormatRaw appends the raw value which a method or a function returned, checked and formatted as the options
	// of the context want it, which AppendFormattedRaw doesn't append ( see AppendRaw ): a SyntacticError of
	// jsontext for a value which is not valid, at its place in the output.
	FormatRaw func(ctx *RuntimeContext, b, raw []byte) ([]byte, error)
	// AppendRawMembers appends the members of the JSON object raw, a jsontext.Value embedded as a fallback, after
	// b, each followed by a comma: check reports whether a name may be written, if it is not nil. Its errors are
	// the ones of the v2 json package, at their places in the output.
	AppendRawMembers func(ctx *RuntimeContext, b, raw []byte, check func(name []byte) bool) ([]byte, error)
	// MarshalTo calls MarshalJSONTo of recv, the pointer to a value of typ, with an encoder which writes after the
	// output b.
	MarshalTo func(ctx *RuntimeContext, b []byte, typ reflect.Type, recv any) ([]byte, error)
	// MarshalToOf returns the function which calls MarshalJSONTo of the value of typ at an address, as MarshalTo
	// does, with the method found once for the type, or appendDefault if it declines ( see ErrUseDefault ).
	MarshalToOf func(typ reflect.Type, appendDefault AppendFunc) AppendFunc
}

// ErrUseDefault is the error of a method or a function of marshaling which declined to write the value: the
// value is written as if it didn't have it.
var ErrUseDefault = stderrors.New("use the default representation")

// implementsAddr reports whether a value of the type t, which is addressable, has the methods of iface.
func implementsAddr(t, iface reflect.Type) bool {
	return t.Implements(iface) || reflect.PointerTo(t).Implements(iface)
}

// The methods of marshaling, in the order of their precedence.
const (
	noMethod = iota
	methodMarshalJSONTo
	methodMarshalJSON
	methodAppendText
	methodMarshalText
)

// v2MethodOf returns the method of marshaling which writes a value of the type, after the method after, or
// noMethod. A pointer or an interface value has none: the value it points to or holds has, which is addressable.
func v2MethodOf(typ reflect.Type, after int) int {
	if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Interface {
		return noMethod
	}
	switch {
	case after < methodMarshalJSONTo && V2Hooks.MarshalerToType != nil && implementsAddr(typ, V2Hooks.MarshalerToType):
		return methodMarshalJSONTo
	case after < methodMarshalJSON && implementsAddr(typ, marshalJSONType):
		return methodMarshalJSON
	case after < methodAppendText && implementsAddr(typ, appendTextType):
		return methodAppendText
	case after < methodMarshalText && implementsAddr(typ, marshalTextType):
		return methodMarshalText
	}
	return noMethod
}

// v2HasMethod reports whether a value of the type is written by a function or a method of marshaling: of the
// type of a default representation, after the ones which declined it ( see v2Code ).
func (c *Compiler) v2HasMethod(typ reflect.Type) bool {
	if typ == c.v2DefaultType {
		return v2MethodOf(typ, c.v2DefaultAfter) != noMethod
	}
	return len(c.funcs.funcsOf(typ)) > 0 || v2MethodOf(typ, noMethod) != noMethod
}

// v2MethodCode returns the code of a value of typ written by its method of marshaling after the method after,
// and true, if it has one.
func (c *Compiler) v2MethodCode(typ reflect.Type, after int) (Code, bool) {
	switch v2MethodOf(typ, after) {
	case methodMarshalJSONTo:
		return c.appendFuncCode(typ, V2Hooks.MarshalToOf(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
			return appendDefault(ctx, b, typ, p, methodMarshalJSONTo)
		})), true
	case methodMarshalJSON:
		if typ == rawMessageType {
			// the output of its method is the value itself, or null: the method, which allocates null, is not
			// called.
			return c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
				raw := *(*[]byte)(p)
				if raw == nil {
					raw = nullJSON
				}
				out, err := AppendRaw(ctx, b, raw)
				if err != nil {
					return b, &errors.MethodError{GoType: typ, Err: err, Kind: errors.MethodJSON}
				}
				return out, nil
			}), true
		}
		appendOutput := addrJSONAppender(typ)
		// the method is called by its code, as the encoder of v1 calls it ( see MarshalerCall ), or by reflect if
		// it is not found by its name.
		marshalJSON, ok := addrMethod[func(unsafe.Pointer) ([]byte, error)](typ, marshalJSONInterface)
		if !ok {
			marshalJSON = func(p unsafe.Pointer) ([]byte, error) {
				return reflect.NewAt(typ, p).Interface().(json.Marshaler).MarshalJSON()
			}
		}
		code := c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
			if appendOutput != nil {
				// the output is known, and valid: the method is called only for its error.
				if out, ok := appendOutput(b, p); ok {
					return out, nil
				}
			}
			raw, err := marshalJSON(p)
			return appendV2Raw(ctx, typ, b, raw, err)
		})
		if m := newMarshalerCall(reflect.PointerTo(typ), marshalJSONInterface); m != nil {
			// called by appendValue without the function ( see MarshalerCall.v2Raw ).
			m.nilIsNull, m.trusted, m.appendOutput, m.v2Raw, m.recv = false, false, appendOutput, true, typ
			m.appendValue = code.appendValue
			code.call = m
		}
		return code, true
	case methodAppendText:
		appendText, ok := addrMethod[func(unsafe.Pointer, []byte) ([]byte, error)](typ, appendTextInterface)
		if !ok {
			appendText = func(p unsafe.Pointer, b []byte) ([]byte, error) {
				return reflect.NewAt(typ, p).Interface().(textAppender).AppendText(b)
			}
		}
		return c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
			return appendMethodText(ctx, b, typ, "AppendText method", p, appendText)
		}), true
	case methodMarshalText:
		marshalText, ok := addrMethod[func(unsafe.Pointer) ([]byte, error)](typ, marshalTextInterface)
		if !ok {
			marshalText = func(p unsafe.Pointer) ([]byte, error) {
				return reflect.NewAt(typ, p).Interface().(encoding.TextMarshaler).MarshalText()
			}
		}
		return c.appendFuncCode(typ, func(ctx *RuntimeContext, b []byte, p unsafe.Pointer) ([]byte, error) {
			return appendMethodText(ctx, b, typ, "MarshalText method", p, func(p unsafe.Pointer, dst []byte) ([]byte, error) {
				text, err := marshalText(p)
				return append(dst, text...), err
			})
		}), true
	}
	return nil, false
}

var (
	// rawMessageType is the type of json.RawMessage, which is jsontext.Value of the standard library where Go has
	// it.
	rawMessageType = reflect.TypeOf(json.RawMessage(nil))
	nullJSON       = []byte("null")
)

// appendV2Raw appends raw, which MarshalJSON of a value of typ returned with err, as a raw value of the v2
// semantics ( see AppendRaw ), or returns the error of the method.
func appendV2Raw(ctx *RuntimeContext, typ reflect.Type, b, raw []byte, err error) ([]byte, error) {
	if err == nil && ctx.Option.Flag&(RawRewriteOption|NormalizeUTF8Option|HTMLEscapeOption) == 0 && len(raw) > 0 &&
		!notPureRaw.Has(raw) && walkRaw[plainStrings](nil, raw, ctx.levelsOfRaw()) {
		// a value of ASCII without a control character and an escape, as most are, for the options which append
		// it as it is ( see AppendFormattedRaw ), without the calls of AppendRaw.
		return append(b, raw...), nil
	}
	if err != nil {
		return b, &errors.MethodError{GoType: typ, Err: unsupportedError(err, "MarshalJSON method"), Kind: errors.MethodJSON}
	}
	out, err := AppendRaw(ctx, b, raw)
	if err != nil {
		return b, &errors.MethodError{GoType: typ, Err: err, Kind: errors.MethodJSON}
	}
	return out, nil
}

// AppendRaw appends the raw value which a method or a function returned, checked and formatted as the options of
// the context want it: as it is, or as AppendFormattedRaw writes it, without a call of the v2 json package, which
// formats the other values.
func AppendRaw(ctx *RuntimeContext, b, raw []byte) ([]byte, error) {
	if ctx.Option.Flag&RawRewriteOption == 0 {
		if out, ok := AppendFormattedRaw(ctx, b, raw); ok {
			return out, nil
		}
	}
	return V2Hooks.FormatRaw(ctx, b, raw)
}

// addrMethod returns the method of iface of a value of typ at an address, which the value or the pointer to it
// has, as a function of the address, which is the data word of the pointer: it is looked up when the type is
// compiled, as the encoder of v1 does ( see methodCode ), and not converted to an interface value by every call.
// It reports false if the method is not found by its name, which the callers then call by reflect.
func addrMethod[F any](typ reflect.Type, iface *marshalerInterface) (F, bool) {
	var f F
	fn, ok := methodCode(reflect.PointerTo(typ), iface)
	if !ok {
		return f, false
	}
	code := &fn // a func value is a pointer to its code
	f = *(*F)(unsafe.Pointer(&code))
	return f, true
}

// addrJSONAppender returns the function which writes the output of MarshalJSON of the value of typ at an address
// without calling it, as the encoder of v1 does ( see newMarshalerCall ): for a type of the standard library whose
// output is known, or a type which has the method from such a type embedded in its value. It returns nil for the
// other types.
func addrJSONAppender(typ reflect.Type) func([]byte, unsafe.Pointer) ([]byte, bool) {
	if runtime.IsStdMarshalerType(typ) {
		return stdJSONAppender(typ)
	}
	if origin, offset, inline, ok := runtime.PromotedStdMethod(typ, runtime.MarshalJSONMethod); ok && inline {
		return embeddedAppender(stdJSONAppender(origin), offset)
	}
	return nil
}

// RunHook runs the VM of the options of ctx for the opcodes of a value, which ctx is set up for, after the output
// b: the run package sets it, which the encoder can't import.
var RunHook func(ctx *RuntimeContext, b []byte, codeSet *OpcodeSet) ([]byte, error)

// defaultKey is the key of the opcodes of the default representation of a type, in a compile mode, after the
// functions and a method.
type defaultKey struct {
	typ   reflect.Type
	mode  compileMode
	after int
}

// defaultCodeSets are the opcodes of the default representations ( see appendDefault ) without functions of
// marshaling: the ones with functions are kept by the functions.
var defaultCodeSets sync.Map // map[defaultKey]*OpcodeSet

// appendDefault appends the value at p, of typ, by its representation after the functions of marshaling and the
// method after, which declined to write it: the next method of it, or the representation of its type. The value
// is written by the opcodes of the representation, in a context of their own, after the output b: the VM runs
// once more within the method which called this.
func appendDefault(ctx *RuntimeContext, b []byte, typ reflect.Type, p unsafe.Pointer, after int) ([]byte, error) {
	key := defaultKey{typ: typ, mode: ctx.compileMode(), after: after}
	var funcs *MarshalFuncs
	cache := &defaultCodeSets
	if ctx.Option.Flag&MarshalFuncsOption != 0 {
		funcs = ctx.Option.Funcs
		cache = &funcs.defaults
	}
	v, ok := cache.Load(key)
	if !ok {
		c := newCompiler(key.mode)
		c.funcs = funcs
		c.v2DefaultType, c.v2DefaultAfter = typ, after
		code, err := c.valueCode(typ)
		if err != nil {
			return b, err
		}
		codeSet, err := c.codeToOpcodeSet(typ, code)
		if err != nil {
			return b, err
		}
		v, _ = cache.LoadOrStore(key, codeSet)
	}
	codeSet := v.(*OpcodeSet)
	return runInner(ctx, b, codeSet, p)
}

// runInner runs the opcodes of the value at p in a context of its own, after the output b, and returns the output
// without the comma after the value, which the caller writes. An error is at its place in that output.
func runInner(ctx *RuntimeContext, b []byte, codeSet *OpcodeSet, p unsafe.Pointer) ([]byte, error) {
	inner := TakeRuntimeContext()
	*inner.Option = *ctx.Option
	// the escaper of the options of ctx is the one of ctx, which records invalid UTF-8 for it: inner sets its own.
	inner.Option.EscaperFlags = noEscaperFlags
	inner.Init(p, codeSet.CodeLength)
	inner.CheckNames = false
	inner.RewriteFrom = ctx.RewriteFrom
	out, err := b, EnterNested(inner, ctx, codeSet, p)
	if err == nil {
		out, err = RunHook(inner, b, codeSet)
	}
	if invalidOut, invalid := inner.InvalidUTF8Output(); invalid {
		// the first error, before which the encoding went on.
		out, err = invalidOut, &errors.TextError{Err: errors.ErrInvalidUTF8}
	}
	ctx.CheckNames = ctx.CheckNames || inner.CheckNames
	ctx.RewriteFrom = inner.RewriteFrom
	inner.Option.V2, inner.Option.Funcs = nil, nil
	ReleaseRuntimeContext(inner)
	if err != nil {
		if _, ok := err.(*errors.OutputError); !ok {
			err = &errors.OutputError{Out: out, Err: err}
		}
		return b, err
	}
	ctx.Rewrote(len(out) - 1)
	return out[:len(out)-1], nil // without the comma after the value, which the caller writes
}

// appendMethodText appends the text which appendText appends for the value at p, as a JSON string.
func appendMethodText(ctx *RuntimeContext, b []byte, typ reflect.Type, what string, p unsafe.Pointer, appendText func(unsafe.Pointer, []byte) ([]byte, error)) ([]byte, error) {
	start := len(b)
	text, err := appendText(p, ctx.MarshalBuf[:0])
	if err != nil {
		return b, &errors.MethodError{GoType: typ, Err: unsupportedError(err, what), Kind: errors.MethodText}
	}
	ctx.MarshalBuf = text[:0]
	out, err := appendName(ctx, b, *(*string)(unsafe.Pointer(&text)))
	if err != nil {
		// invalid UTF-8, of the string at the place of the value.
		return b[:start], &errors.MethodError{GoType: typ, Err: err, Kind: errors.MethodText}
	}
	return out, nil
}

// unsupportedError returns the error of a method or a function, which must not report errors.ErrUnsupported:
// only MarshalJSONTo and the functions of MarshalToFunc may.
func unsupportedError(err error, what string) error {
	if stderrors.Is(err, stderrors.ErrUnsupported) {
		return stderrors.New(what + " may not return errors.ErrUnsupported")
	}
	return err
}
