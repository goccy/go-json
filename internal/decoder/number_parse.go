package decoder

// The numbers of the input are parsed in one pass: the digits are accumulated as they are read.

// maxUint64Digits is the number of the digits which a uint64 holds whatever they are: 19 nines are less than 1<<64.
const maxUint64Digits = 19

// parseDigits reads the digits at cursor, which starts a number after its sign, and returns their value, the
// position after them and their number. A number which starts with 0 has no other digit: the digit after it,
// if any, is not read. The value is exact for up to maxUint64Digits digits.
func parseDigits(buf []byte, cursor int64) (uint64, int64, int) {
	if buf[cursor] == '0' {
		return 0, cursor + 1, 1
	}
	start := cursor
	var u uint64
	for {
		d := buf[cursor] - '0'
		if d > 9 {
			break
		}
		u = u*10 + uint64(d)
		cursor++
	}
	return u, cursor, int(cursor - start)
}

// isFloatContinuation reports whether c continues an integer as a number which is not an integer.
func isFloatContinuation(c byte) bool {
	return c == '.' || c == 'e' || c == 'E'
}

// float64pow10 are the powers of ten which a float64 holds exactly.
var float64pow10 = [...]float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11,
	1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22,
}

// parseFloatFast parses the number at cursor, as the JSON grammar has it, when it is exactly the float64
// of its mantissa times or divided by a power of ten which a float64 holds exactly: then the result is
// correctly rounded, as by strconv.ParseFloat ( Clinger's fast path ). It returns false for any other number,
// or anything which is not a number by the grammar, which the caller parses as it did before.
func parseFloatFast(buf []byte, cursor int64) (float64, int64, bool) {
	neg := buf[cursor] == '-'
	if neg {
		cursor++
	}
	if buf[cursor]-'0' > 9 {
		return 0, 0, false
	}
	mantissa, cursor, digits := parseDigits(buf, cursor)
	exp := 0
	if buf[cursor] == '.' {
		cursor++
		start := cursor
		for {
			d := buf[cursor] - '0'
			if d > 9 {
				break
			}
			mantissa = mantissa*10 + uint64(d)
			cursor++
		}
		fraction := int(cursor - start)
		if fraction == 0 {
			return 0, 0, false
		}
		digits += fraction
		exp = -fraction
	}
	if c := buf[cursor]; c == 'e' || c == 'E' {
		cursor++
		expNeg := false
		switch buf[cursor] {
		case '-':
			expNeg = true
			cursor++
		case '+':
			cursor++
		}
		start := cursor
		e := 0
		for {
			d := buf[cursor] - '0'
			if d > 9 {
				break
			}
			if e < 10000 {
				e = e*10 + int(d)
			}
			cursor++
		}
		if cursor == start {
			return 0, 0, false
		}
		if expNeg {
			e = -e
		}
		exp += e
	}
	if digits > maxUint64Digits || mantissa > 1<<53 || exp < -22 || exp > 22 {
		return 0, 0, false
	}
	f := float64(mantissa)
	if exp < 0 {
		f /= float64pow10[-exp]
	} else {
		f *= float64pow10[exp]
	}
	if neg {
		f = -f
	}
	return f, cursor, true
}
