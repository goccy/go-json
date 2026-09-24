package decoder

import (
	"reflect"
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/runtime"
)

type RuntimeContext struct {
	Buf    []byte
	Option *Option
	// slot is where a value of interface{} is decoded before it is put into a map or a slice.
	// The decoders of the values nested in it use it too: each of them takes the value out of it
	// before the next value is decoded, and puts its own value into it only at its end.
	slot any
	// anyStack holds the elements of the arrays being decoded into []interface{}: an array pushes
	// its elements above the ones of the arrays it is nested in, and pops them at its end.
	anyStack []any
	// floats and strings are the slabs in which the numbers and the strings decoded into interface{}
	// are kept, so that an interface value refers to them without an allocation of its own.
	// A slot of a slab is never written again once an interface value refers to it.
	floats  []float64
	strings []string
	// recentDecoders are the decoders of the types decoded last, in the sets indexed by the address of the type.
	// Only the contexts of the pool have them: a Decoder has a context of its own, which it would allocate
	// with them for every stream.
	recentDecoders *[recentDecoderSets]recentDecoderSet
}

const (
	// recentDecoderSets is the number of the sets of the recent decoders, of recentDecoderWays entries each,
	// as the recent opcodes of the encoder: the set of a type is the top bits of the product of its address
	// with an odd constant, and a set holds the two types hashed to it which were decoded last, so that the
	// type passed to Unmarshal and a type held by its values of interface{} never evict each other.
	recentDecoderSets      = 16
	recentDecoderHashShift = 64 - 4
	recentDecoderWays      = 2
)

type recentDecoder struct {
	typeptr uintptr
	dec     Decoder
}

// recentDecoderSet is the entries of a set, the one decoded last first.
type recentDecoderSet [recentDecoderWays]recentDecoder

// DecoderOf returns the decoder of the type, compiling it if the type is new.
//
// A runtime context remembers the decoders of the types it decoded last: the same types are decoded again
// and again in most of the programs, and this is cheaper than a lookup of the table shared by every goroutine.
// The first entry of the set is looked at here, which is inlined into the callers; the rest in lookupDecoder.
func (ctx *RuntimeContext) DecoderOf(typ unsafe.Pointer) (Decoder, error) {
	if ctx.recentDecoders == nil {
		return CompileToGetDecoder(typ)
	}
	set := &ctx.recentDecoders[(uint64(uintptr(typ))*runtime.TypeHashMultiplier)>>recentDecoderHashShift]
	if set[0].typeptr == uintptr(typ) {
		return set[0].dec, nil
	}
	return lookupDecoder(set, typ)
}

// lookupDecoder returns the decoder of the type from the second entry of its set, or from the shared table,
// compiling it if the type is new. A decoder found in the shared table takes the first entry of the set,
// and the one decoded before it is kept in the second.
func lookupDecoder(set *recentDecoderSet, typ unsafe.Pointer) (Decoder, error) {
	if set[1].typeptr == uintptr(typ) {
		return set[1].dec, nil
	}
	dec, err := CompileToGetDecoder(typ)
	if err != nil {
		return nil, err
	}
	set[1] = set[0]
	set[0] = recentDecoder{typeptr: uintptr(typ), dec: dec}
	return dec, nil
}

// boxSlabSize is the number of the values a slab of floats or strings holds.
const boxSlabSize = 32

var (
	float64TypePtr = runtime.TypePtr(reflect.TypeOf(float64(0)))
	stringTypePtr  = runtime.TypePtr(reflect.TypeOf(""))
)

// boxFloat returns f as an interface value, which refers to a slot of the slab of floats.
func (ctx *RuntimeContext) boxFloat(f float64) any {
	if len(ctx.floats) == cap(ctx.floats) {
		ctx.floats = make([]float64, 0, boxSlabSize)
	}
	ctx.floats = append(ctx.floats, f)
	return *(*any)(unsafe.Pointer(&emptyInterface{typ: float64TypePtr, ptr: unsafe.Pointer(&ctx.floats[len(ctx.floats)-1])}))
}

// boxString returns s as an interface value, which refers to a slot of the slab of strings.
func (ctx *RuntimeContext) boxString(s string) any {
	if len(ctx.strings) == cap(ctx.strings) {
		ctx.strings = make([]string, 0, boxSlabSize)
	}
	ctx.strings = append(ctx.strings, s)
	return *(*any)(unsafe.Pointer(&emptyInterface{typ: stringTypePtr, ptr: unsafe.Pointer(&ctx.strings[len(ctx.strings)-1])}))
}

// popAny removes the elements of anyStack from base, clearing them so that the stack keeps nothing alive.
func (ctx *RuntimeContext) popAny(base int) {
	clear(ctx.anyStack[base:])
	ctx.anyStack = ctx.anyStack[:base]
}

var (
	runtimeContextPool = sync.Pool{
		New: func() any {
			return &RuntimeContext{
				Option:         &Option{},
				recentDecoders: &[recentDecoderSets]recentDecoderSet{},
			}
		},
	}
)

func TakeRuntimeContext() *RuntimeContext {
	return runtimeContextPool.Get().(*RuntimeContext)
}

func ReleaseRuntimeContext(ctx *RuntimeContext) {
	// Nothing of the call is kept: the input is referred to by the decoded strings.
	ctx.Buf = nil
	ctx.slot = nil
	ctx.popAny(0)
	// The strings refer to the input: the slab is not kept, so that a context in the pool doesn't keep
	// the input of a previous call alive. The slab of floats refers to nothing and is kept.
	ctx.strings = nil
	runtimeContextPool.Put(ctx)
}

var (
	isWhiteSpace = [256]bool{}
)

func init() {
	isWhiteSpace[' '] = true
	isWhiteSpace['\n'] = true
	isWhiteSpace['\t'] = true
	isWhiteSpace['\r'] = true
}

func char(ptr unsafe.Pointer, offset int64) byte {
	return *(*byte)(unsafe.Add(ptr, offset))
}

func skipWhiteSpace(buf []byte, cursor int64) int64 {
	for isWhiteSpace[buf[cursor]] {
		cursor++
	}
	return cursor
}

func skipValue(buf []byte, cursor, depth int64) (int64, error) {
	for {
		switch buf[cursor] {
		case ' ', '\t', '\n', '\r':
			cursor++
			continue
		case '{', '[':
			// as skipCompound does, without its call
			sc := compoundScanner{depth: 1, maxDepth: maxDecodeNestingDepth - depth}
			lim := int64(len(buf))
			var (
				end   int64
				found bool
				err   error
			)
			if lim-cursor <= scanBlockSize {
				end, found, err = sc.scanBytes(buf, cursor+1, lim)
			} else {
				end, found, err = sc.scan(buf, cursor+1, lim)
			}
			if err != nil {
				return 0, err
			}
			if !found {
				return 0, errors.ErrUnexpectedEndOfJSON("object or array", end)
			}
			return end, nil
		case '"':
			for {
				cursor++
				switch buf[cursor] {
				case '\\':
					cursor++
					if buf[cursor] == nul {
						return 0, errors.ErrUnexpectedEndOfJSON("string of object", cursor)
					}
				case '"':
					return cursor + 1, nil
				case nul:
					return 0, errors.ErrUnexpectedEndOfJSON("string of object", cursor)
				}
			}
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			for {
				cursor++
				if floatTable[buf[cursor]] {
					continue
				}
				break
			}
			return cursor, nil
		case 't':
			if err := validateTrue(buf, cursor); err != nil {
				return 0, err
			}
			cursor += 4
			return cursor, nil
		case 'f':
			if err := validateFalse(buf, cursor); err != nil {
				return 0, err
			}
			cursor += 5
			return cursor, nil
		case 'n':
			if err := validateNull(buf, cursor); err != nil {
				return 0, err
			}
			cursor += 4
			return cursor, nil
		default:
			return cursor, errors.ErrUnexpectedEndOfJSON("null", cursor)
		}
	}
}

func validateTrue(buf []byte, cursor int64) error {
	if cursor+3 >= int64(len(buf)) {
		return errors.ErrUnexpectedEndOfJSON("true", cursor)
	}
	if buf[cursor+1] != 'r' {
		return errors.ErrInvalidCharacter(buf[cursor+1], "true", cursor+1)
	}
	if buf[cursor+2] != 'u' {
		return errors.ErrInvalidCharacter(buf[cursor+2], "true", cursor+2)
	}
	if buf[cursor+3] != 'e' {
		return errors.ErrInvalidCharacter(buf[cursor+3], "true", cursor+3)
	}
	return nil
}

func validateFalse(buf []byte, cursor int64) error {
	if cursor+4 >= int64(len(buf)) {
		return errors.ErrUnexpectedEndOfJSON("false", cursor)
	}
	if buf[cursor+1] != 'a' {
		return errors.ErrInvalidCharacter(buf[cursor+1], "false", cursor+1)
	}
	if buf[cursor+2] != 'l' {
		return errors.ErrInvalidCharacter(buf[cursor+2], "false", cursor+2)
	}
	if buf[cursor+3] != 's' {
		return errors.ErrInvalidCharacter(buf[cursor+3], "false", cursor+3)
	}
	if buf[cursor+4] != 'e' {
		return errors.ErrInvalidCharacter(buf[cursor+4], "false", cursor+4)
	}
	return nil
}

func validateNull(buf []byte, cursor int64) error {
	if cursor+3 >= int64(len(buf)) {
		return errors.ErrUnexpectedEndOfJSON("null", cursor)
	}
	if buf[cursor+1] != 'u' {
		return errors.ErrInvalidCharacter(buf[cursor+1], "null", cursor+1)
	}
	if buf[cursor+2] != 'l' {
		return errors.ErrInvalidCharacter(buf[cursor+2], "null", cursor+2)
	}
	if buf[cursor+3] != 'l' {
		return errors.ErrInvalidCharacter(buf[cursor+3], "null", cursor+3)
	}
	return nil
}
