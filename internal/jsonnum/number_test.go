package jsonnum

import (
	"math/rand/v2"
	"regexp"
	"strings"
	"testing"
)

// NonDigits has the bits of the bytes which are not digits, whatever the other bytes are.
func TestNonDigits(t *testing.T) {
	for pos := 0; pos < 8; pos++ {
		for c := 0; c < 256; c++ {
			for _, fill := range []byte{'0', '5', '9', 'a', '/', ':', 0x80, 0xb0, 0xff, 0} {
				var w uint64
				var b [8]byte
				for i := 0; i < 8; i++ {
					b[i] = fill
					if i == pos {
						b[i] = byte(c)
					}
					w |= uint64(b[i]) << (8 * i)
				}
				got := NonDigits(w)
				for i := 0; i < 8; i++ {
					want := b[i] < '0' || b[i] > '9'
					set := got&(1<<i) != 0
					if set != want {
						t.Fatalf("% x: bit %d is %v, want %v", b, i, set, want)
					}
				}
			}
		}
	}
}

// numberPattern is the grammar of RFC 8259, section 6, which the scans are checked against.
var numberPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// referenceScan is Scan a byte at a time, without the runs of digits read 8 at a time.
func referenceScan(b []byte) (int, Place) {
	i := 0
	digits := func() int {
		start := i
		for i < len(b) && '0' <= b[i] && b[i] <= '9' {
			i++
		}
		return i - start
	}
	if i < len(b) && b[i] == '-' {
		i++
	}
	if i < len(b) && b[i] == '0' {
		i++
	} else if digits() == 0 {
		return i, Start
	}
	if i < len(b) && b[i] == '.' {
		i++
		if digits() == 0 {
			return i, Fraction
		}
	}
	if i < len(b) && (b[i] == 'e' || b[i] == 'E') {
		i++
		if i < len(b) && (b[i] == '+' || b[i] == '-') {
			i++
		}
		if digits() == 0 {
			return i, Exponent
		}
	}
	return i, Valid
}

// randomText is a string of the bytes of numbers and some others, long enough for the runs of 8 digits.
func randomText(r *rand.Rand) string {
	const alphabet = "-+.eE0123456789012345678901234567890123456789 x\x00\x80"
	var b strings.Builder
	for range r.IntN(30) {
		b.WriteByte(alphabet[r.IntN(len(alphabet))])
	}
	return b.String()
}

func TestScanByGrammar(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	inputs := []string{"", "-", "0", "-0", "01", "1.", "1.5", ".5", "+1", "1e", "1e+", "1e5", "1E-5", "1.5e10", "--1", "1-2",
		"0.0", "123456789012345678901234567890", "1.0000000000000000000001", "1e123456789012"}
	for range 200000 {
		inputs = append(inputs, randomText(r))
	}
	for _, in := range inputs {
		b := []byte(in)
		n, p := Scan(b)
		if wn, wp := referenceScan(b); n != wn || p != wp {
			t.Fatalf("Scan(%q) = %d, %d; want %d, %d", in, n, p, wn, wp)
		}
		if p == Valid && !numberPattern.Match(b[:n]) {
			t.Fatalf("Scan(%q) takes %q, which is not a number", in, b[:n])
		}
		if got, want := IsValid(b), numberPattern.Match(b); got != want {
			t.Fatalf("IsValid(%q) = %v, want %v", in, got, want)
		}
		// End takes a valid number which a byte follows: the byte, which can't continue it, is the next token's
		want := -1
		if p == Valid && n < len(b) {
			want = n
		}
		if got := End(b); got != want {
			t.Fatalf("End(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestScanPlaces(t *testing.T) {
	for _, c := range []struct {
		in    string
		n     int
		place Place
	}{
		{"", 0, Start}, {"-", 1, Start}, {"x", 0, Start}, {"-x", 1, Start}, {"1.", 2, Fraction}, {"1.x", 2, Fraction},
		{"1e", 2, Exponent}, {"1e+", 3, Exponent}, {"1E-x", 3, Exponent}, {"01", 1, Valid}, {"1.5e3x", 5, Valid},
	} {
		if n, p := Scan([]byte(c.in)); n != c.n || p != c.place {
			t.Errorf("Scan(%q) = %d, %d; want %d, %d", c.in, n, p, c.n, c.place)
		}
	}
}

func TestScanDigits(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for range 100000 {
		b := []byte(randomText(r))
		i := r.IntN(len(b) + 1)
		want := i
		for want < len(b) && '0' <= b[want] && b[want] <= '9' {
			want++
		}
		if got := ScanDigits(b, i); got != want {
			t.Fatalf("ScanDigits(%q, %d) = %d, want %d", b, i, got, want)
		}
	}
}
