package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// Numbers that strconv.ParseFloat accepts but the JSON grammar (RFC 8259) and
// encoding/json reject: leading zeros, a bare trailing dot, and a fraction with
// no integer part after a sign. go-json used to accept these, diverging from
// encoding/json.
var invalidNumberLiterals = []string{
	"00", "01", "0123", "007", "-00", "-01",
	"00.5", "01.5", "00e1",
	"1.", "1.e5",
	"-.5",
}

// Well-formed numbers that must keep decoding successfully.
var validNumberLiterals = []string{
	"0", "-0", "123", "-123",
	"0.5", "1.5", "-1.5", "0.0",
	"1e5", "1E5", "1e+5", "1e-5", "1.5e10", "-0.0e-1",
	"9223372036854775807",
}

func TestUnmarshalRejectsInvalidNumberLiterals(t *testing.T) {
	for _, lit := range invalidNumberLiterals {
		// Cross-check that encoding/json indeed rejects the literal, so this
		// test tracks stdlib rather than an arbitrary rule.
		var std interface{}
		if stdjson.Unmarshal([]byte(lit), &std) == nil {
			t.Fatalf("test invariant broken: encoding/json accepted %q", lit)
		}

		var iface interface{}
		if err := json.Unmarshal([]byte(lit), &iface); err == nil {
			t.Errorf("Unmarshal(%q) into interface{}: expected error, got value %v", lit, iface)
		}

		var f float64
		if err := json.Unmarshal([]byte(lit), &f); err == nil {
			t.Errorf("Unmarshal(%q) into float64: expected error, got value %v", lit, f)
		}

		var num json.Number
		if err := json.Unmarshal([]byte(lit), &num); err == nil {
			t.Errorf("Unmarshal(%q) into json.Number: expected error, got value %v", lit, num)
		}

		if json.Valid([]byte(lit)) {
			t.Errorf("Valid(%q) = true, want false", lit)
		}
	}
}

func TestUnmarshalAcceptsValidNumberLiterals(t *testing.T) {
	for _, lit := range validNumberLiterals {
		var iface interface{}
		if err := json.Unmarshal([]byte(lit), &iface); err != nil {
			t.Errorf("Unmarshal(%q) into interface{}: unexpected error: %v", lit, err)
		}

		var f float64
		if err := json.Unmarshal([]byte(lit), &f); err != nil {
			t.Errorf("Unmarshal(%q) into float64: unexpected error: %v", lit, err)
		}

		var num json.Number
		if err := json.Unmarshal([]byte(lit), &num); err != nil {
			t.Errorf("Unmarshal(%q) into json.Number: unexpected error: %v", lit, err)
		}

		if !json.Valid([]byte(lit)) {
			t.Errorf("Valid(%q) = false, want true", lit)
		}
	}
}
