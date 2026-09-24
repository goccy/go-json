package decoder

import (
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
)

type RuntimeContext struct {
	Buf    []byte
	Option *Option
}

var (
	runtimeContextPool = sync.Pool{
		New: func() any {
			return &RuntimeContext{
				Option: &Option{},
			}
		},
	}
)

func TakeRuntimeContext() *RuntimeContext {
	return runtimeContextPool.Get().(*RuntimeContext)
}

func ReleaseRuntimeContext(ctx *RuntimeContext) {
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
			return skipCompound(buf, cursor+1, 1, depth)
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
