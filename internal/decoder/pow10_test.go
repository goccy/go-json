package decoder

import (
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"testing"
)

// mulPow10 gives the float64 of strconv.ParseFloat, bit for bit, whenever it gives one.
func TestMulPow10AsStrconv(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	decided := 0
	const n = 1000000
	for i := 0; i < n; i++ {
		var w uint64
		switch i % 3 {
		case 0:
			w = r.Uint64() % 10000000000000000000 // up to 19 digits
		case 1:
			w = r.Uint64() >> uint(r.Intn(64))
		default:
			w = uint64(r.Int63n(100000000000000000)) // 17 digits, as the shortest float64s
		}
		q := r.Intn(maxPow10Exp10-minPow10Exp10+1) + minPow10Exp10
		if i%2 == 0 {
			q = r.Intn(60) - 30 // the usual exponents
		}
		got, ok := mulPow10(w, q)
		if !ok {
			continue
		}
		decided++
		want, err := strconv.ParseFloat(strconv.FormatUint(w, 10)+"e"+strconv.Itoa(q), 64)
		if err != nil {
			t.Fatalf("%de%d: mulPow10 gave %v, strconv failed: %v", w, q, got, err)
		}
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("%de%d: got %v ( %x ), want %v ( %x )", w, q, got, math.Float64bits(got), want, math.Float64bits(want))
		}
	}
	// nearly every number is decided by the table; the rest are left to strconv.
	if decided < n*9/10 {
		t.Fatalf("only %d of %d numbers were decided", decided, n)
	}
	t.Logf("%d of %d decided", decided, n)
}

// A number exactly half-way between two float64 values is never decided wrongly: the table gives none, or the even one.
func TestMulPow10HalfWay(t *testing.T) {
	r := rand.New(rand.NewSource(5))
	checked, decided := 0, 0
	for i := 0; i < 10000; i++ {
		// the half-way point between m*2^e and (m+1)*2^e, written as a decimal: (2m+1) * 2^(e-1).
		// Its decimal has at most 19 digits for these exponents.
		m := uint64(1)<<52 | r.Uint64()&(1<<52-1)
		e := r.Intn(12) - 2
		x := new(big.Float).SetPrec(2000).SetMantExp(new(big.Float).SetUint64(2*m+1), e-1)
		s := x.Text('e', -1)
		want, err := strconv.ParseFloat(s, 64)
		if err != nil {
			t.Fatal(err)
		}
		mant, exp, ok := splitDecimal(s)
		if !ok {
			continue
		}
		checked++
		got, ok := mulPow10(mant, exp)
		if !ok {
			continue
		}
		decided++
		if math.Float64bits(got) != math.Float64bits(want) {
			t.Fatalf("%s: got %v, want %v", s, got, want)
		}
	}
	if checked < 5000 {
		t.Fatalf("only %d half-way points were checked", checked)
	}
	t.Logf("%d half-way points checked, %d decided by the table", checked, decided)
}

// splitDecimal returns the mantissa and the exponent of ten of a decimal of at most 19 digits.
func splitDecimal(s string) (uint64, int, bool) {
	var mant uint64
	digits, exp := 0, 0
	i := 0
	for ; i < len(s) && s[i] != 'e'; i++ {
		switch c := s[i]; {
		case c == '.':
			exp = -(len(s[i+1:]) - len(s[indexByte(s, 'e'):]))
		case c >= '0' && c <= '9':
			mant = mant*10 + uint64(c-'0')
			digits++
		}
	}
	if digits > 19 {
		return 0, 0, false
	}
	e, err := strconv.Atoi(s[i+1:])
	if err != nil {
		return 0, 0, false
	}
	return mant, exp + e, true
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return len(s)
}

// Every entry of the generated table is the truncated significand of its power of ten.
func TestPow10Table(t *testing.T) {
	for q := minPow10Exp10; q <= maxPow10Exp10; q++ {
		p := pow10Table[q-minPow10Exp10]
		t128 := new(big.Int).Lsh(new(big.Int).SetUint64(p.hi), 64)
		t128.Or(t128, new(big.Int).SetUint64(p.lo))
		if t128.BitLen() != 128 {
			t.Fatalf("1e%d: the significand has %d bits", q, t128.BitLen())
		}
		// T * 2^exp2 <= 10^q < (T+1) * 2^exp2, compared as rationals.
		pow := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(q))), nil))
		if q < 0 {
			pow.Inv(pow)
		}
		scale := new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(abs(p.exp2))))
		if p.exp2 < 0 {
			scale.Inv(scale)
		}
		low := new(big.Rat).Mul(new(big.Rat).SetInt(t128), scale)
		high := new(big.Rat).Mul(new(big.Rat).SetInt(new(big.Int).Add(t128, big.NewInt(1))), scale)
		if low.Cmp(pow) > 0 || high.Cmp(pow) <= 0 {
			t.Fatalf("1e%d is not in [T, T+1) * 2^%d", q, p.exp2)
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
