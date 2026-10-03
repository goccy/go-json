package jsonnum

import (
	"math"
	"math/rand/v2"
	"strconv"
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
