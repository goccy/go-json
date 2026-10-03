package decoder

import (
	"fmt"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonnum"
)

// The numbers of the input are parsed by jsonnum, in the buffer which a nul byte ends. The functions here are
// inlined, so that they are direct calls, which jsonnum.ParseDigits is inlined into.

// maxUint64Digits is the number of the digits which a uint64 holds whatever they are.
const maxUint64Digits = jsonnum.MaxUint64Digits

func parseDigits(buf []byte, cursor int64) (uint64, int64, int) {
	return jsonnum.ParseDigits(buf, cursor)
}

func isFloatContinuation(c byte) bool {
	return jsonnum.IsFloatContinuation(c)
}

func parseFloatFast(buf []byte, cursor int64) (float64, int64, bool) {
	return jsonnum.ParseFloatTerminated(buf, cursor)
}

// numberEnd returns the position after the number which starts at start, validated by the grammar of the JSON
// numbers, which is followed by a byte which may end a value.
func numberEnd(buf []byte, start int64) (int64, error) {
	c := start
	if buf[c] == '-' {
		c++
	}
	switch {
	case buf[c] == '0':
		c++
	case '1' <= buf[c] && buf[c] <= '9':
		c = skipDigits(buf, c+1)
	default:
		return 0, errors.ErrSyntax(fmt.Sprintf("invalid character %s in numeric literal", quoteChar(buf[c])), c+1)
	}
	if buf[c] == '.' {
		c++
		if buf[c]-'0' > 9 {
			return 0, errors.ErrSyntax(fmt.Sprintf("invalid character %s after decimal point in numeric literal", quoteChar(buf[c])), c+1)
		}
		c = skipDigits(buf, c)
	}
	if buf[c] == 'e' || buf[c] == 'E' {
		c++
		if buf[c] == '+' || buf[c] == '-' {
			c++
		}
		if buf[c]-'0' > 9 {
			return 0, errors.ErrSyntax(fmt.Sprintf("invalid character %s in exponent of numeric literal", quoteChar(buf[c])), c+1)
		}
		c = skipDigits(buf, c)
	}
	if !validEndNumberChar[buf[c]] {
		return 0, errors.ErrSyntax(fmt.Sprintf("invalid character %s after top-level value", quoteChar(buf[c])), c+1)
	}
	return c, nil
}

// skipDigits returns the position of the first byte from c which is not a digit.
func skipDigits(buf []byte, c int64) int64 {
	for buf[c]-'0' <= 9 {
		c++
	}
	return c
}
