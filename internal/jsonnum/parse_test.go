package jsonnum

import (
	"math"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

// TestParseFloatAsStrconv parses random numbers in their JSON forms, and checks that a number which ParseFloat
// parses has the float64 of strconv.ParseFloat, bit for bit.
func TestParseFloatAsStrconv(t *testing.T) {
	r := rand.New(rand.NewPCG(5, 6))
	inputs := []string{"0", "-0", "1", "3.14", "1e22", "1e23", "9007199254740993", "0.1", "-1.5e-300", "1e-400", "1e400",
		"123456789012345678901", "0.000000000000000000001", "1.7976931348623157e308", "4.9e-324", "1.", "", "-", "01"}
	for range 300000 {
		var s string
		switch r.IntN(4) {
		case 0:
			s = strconv.FormatFloat(math.Float64frombits(r.Uint64()), 'g', -1, 64)
		case 1:
			s = strconv.FormatFloat(r.Float64()*math.Pow(10, float64(r.IntN(40)-20)), 'f', r.IntN(20), 64)
		case 2:
			s = strconv.FormatUint(r.Uint64()>>r.IntN(64), 10)
		default:
			s = strconv.FormatFloat(r.NormFloat64(), 'e', r.IntN(20), 64)
		}
		inputs = append(inputs, s)
	}
	parsed := 0
	for _, s := range inputs {
		if !IsValid([]byte(s)) {
			if _, ok := ParseFloat([]byte(s)); ok {
				t.Fatalf("ParseFloat(%q) parses a text which is not a JSON number", s)
			}
			continue
		}
		f, ok := ParseFloat([]byte(s))
		if !ok {
			continue
		}
		parsed++
		want, err := strconv.ParseFloat(s, 64)
		if err != nil || math.Float64bits(f) != math.Float64bits(want) {
			t.Fatalf("ParseFloat(%q) = %v, want %v (%v)", s, f, want, err)
		}
	}
	if parsed < len(inputs)/2 {
		t.Errorf("ParseFloat parsed %d of %d numbers", parsed, len(inputs))
	}
}

// TestParseFloatTerminatedPastUint64 parses numbers whose digits make the mantissa wrap around to 0: the digits of a
// multiple of 1<<64, which 10^64 and its multiples are, with the point between any two of them or after them, and
// the digits around the point which wrap the mantissa on a digit of the fraction. A number which ParseFloatTerminated
// parses has the float64 of strconv.ParseFloat, bit for bit, whatever its length, and so has one which ParseFloat
// parses.
func TestParseFloatTerminatedPastUint64(t *testing.T) {
	r := rand.New(rand.NewPCG(9, 10))
	pow64 := new(big.Int).Lsh(big.NewInt(1), 64)
	var wrapping []string
	for k := range 40 {
		wrapping = append(wrapping, new(big.Int).Mul(pow64, big.NewInt(int64(k+1))).String())
	}
	for range 200 {
		k := new(big.Int).SetUint64(r.Uint64() >> r.IntN(64))
		wrapping = append(wrapping, k.Add(k, big.NewInt(1)).Mul(k, pow64).String())
	}
	for zeros := 64; zeros <= 70; zeros++ {
		wrapping = append(wrapping, "1"+strings.Repeat("0", zeros), "2"+strings.Repeat("0", zeros))
	}
	fractions := []string{"0", "5", "25", "123456", "0000000000000000001", "1234567890123456789"}
	var inputs []string
	for _, digits := range wrapping {
		for _, fraction := range fractions {
			inputs = append(inputs, digits+"."+fraction)
		}
		for point := 1; point < len(digits); point++ {
			inputs = append(inputs, digits[:point]+"."+digits[point:], digits[:point]+"."+digits[point:]+"00")
		}
		inputs = append(inputs, "0."+digits, "0.000"+digits, "0."+digits+"00")
	}
	for i, s := range inputs {
		if r.IntN(2) == 0 {
			s = "-" + s
		}
		if i%3 == 0 {
			s += "e" + strconv.Itoa(r.IntN(61)-30)
		}
		want, err := strconv.ParseFloat(s, 64)
		if f, end, ok := ParseFloatTerminated(append([]byte(s), 0), 0); ok {
			if err != nil || int(end) != len(s) || math.Float64bits(f) != math.Float64bits(want) {
				t.Fatalf("ParseFloatTerminated(%q) = %v, %d; want %v, %d (%v)", s, f, end, want, len(s), err)
			}
		}
		if f, ok := ParseFloat([]byte(s)); ok && (err != nil || math.Float64bits(f) != math.Float64bits(want)) {
			t.Fatalf("ParseFloat(%q) = %v, want %v (%v)", s, f, want, err)
		}
	}
	// The 0 of an integer part 0 and the zeros of the fraction before its first digit which is not zero are not
	// digits of the mantissa: a number of more digits than a mantissa holds, of at most MaxUint64Digits after them,
	// is parsed, and the digits of its exponent are not digits of the mantissa either. One of a mantissa of at most
	// 53 bits and a power of ten which a float64 holds is parsed whatever the digits are, as Clinger's fast path has
	// it.
	for zeros := range 30 {
		for _, digits := range []string{"1", "5", "123456789", "9007199254740993", "1234567890123456789"} {
			for _, exponent := range []int{0, 5, -5, 20} {
				s := "0." + strings.Repeat("0", zeros) + digits
				if exponent != 0 {
					s += "e" + strconv.Itoa(exponent)
				}
				if n := significantDigits(append([]byte(s), 0), 0); n != len(digits) {
					t.Fatalf("significantDigits(%q) = %d, want %d", s, n, len(digits))
				}
				want, _ := strconv.ParseFloat(s, 64)
				f, end, ok := ParseFloatTerminated(append([]byte(s), 0), 0)
				exp := exponent - zeros - len(digits)
				if clinger := len(digits) <= 15 && -22 <= exp && exp <= 22; !ok && clinger {
					t.Fatalf("ParseFloatTerminated(%q) is not parsed", s)
				}
				if ok && (int(end) != len(s) || math.Float64bits(f) != math.Float64bits(want)) {
					t.Fatalf("ParseFloatTerminated(%q) = %v, %d; want %v, %d", s, f, end, want, len(s))
				}
			}
		}
	}
}

// TestParseUintAsStrconv checks ParseUint against strconv.ParseUint for integers around the range of a uint64.
func TestParseUintAsStrconv(t *testing.T) {
	r := rand.New(rand.NewPCG(7, 8))
	inputs := []string{"", "0", "00", "01", "1", "18446744073709551615", "18446744073709551616", "99999999999999999999",
		"100000000000000000000", "1844674407370955161", "12a", "-1", "1.0"}
	for range 100000 {
		inputs = append(inputs, strconv.FormatUint(r.Uint64()>>r.IntN(64), 10))
		inputs = append(inputs, strconv.FormatUint(r.Uint64(), 10)+strconv.Itoa(r.IntN(10)))
	}
	for _, s := range inputs {
		got, ok := ParseUint([]byte(s))
		want, err := strconv.ParseUint(s, 10, 64)
		wantOK := err == nil && (s == "0" || s[0] != '0')
		if ok != wantOK || ok && got != want {
			t.Fatalf("ParseUint(%q) = %d, %v; want %d, %v", s, got, ok, want, wantOK)
		}
	}
}
