package json_test

import (
	stdjson "encoding/json"
	"strings"
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
	"0.", "0.e1", "1.", "1.e5",
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

func TestUnmarshalRejectsInvalidSkippedNumberLiterals(t *testing.T) {
	type target struct {
		Known bool `json:"known"`
	}

	check := func(input string) {
		t.Helper()

		var std target
		if err := stdjson.Unmarshal([]byte(input), &std); err == nil {
			t.Fatalf("test invariant broken: encoding/json accepted %s", input)
		}

		var got target
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("Unmarshal %s: expected error, got %+v", input, got)
		}

		var streamed target
		if err := json.NewDecoder(strings.NewReader(input)).Decode(&streamed); err == nil {
			t.Errorf("Decoder %s: expected error, got %+v", input, streamed)
		}
	}

	for _, lit := range invalidNumberLiterals {
		inputs := []string{
			`{"skip":` + lit + `,"known":true}`,
			`{"skip":{"nested":` + lit + `},"known":true}`,
			`{"skip":[` + lit + `],"known":true}`,
		}
		for _, input := range inputs {
			check(input)
		}
	}

	for _, input := range []string{
		`{"skip":1x,"known":true}`,
		`{"skip":[1x],"known":true}`,
		`{"skip":{"nested":1x},"known":true}`,
		`{"skip":[1 2],"known":true}`,
		`{"skip":{"nested":1 "other":2},"known":true}`,
	} {
		check(input)
	}
}

func TestDecodeFirstWinSkipsRemainingObjectTail(t *testing.T) {
	type target struct {
		Known int `json:"known"`
	}
	input := `{"known":1,"skip":[2],"known":9,"extra":{"nested":3},"after":4}`

	var got target
	if err := json.UnmarshalWithOption([]byte(input), &got, json.DecodeFieldPriorityFirstWin()); err != nil {
		t.Fatalf("UnmarshalWithOption FirstWin: unexpected error: %v", err)
	}
	if got.Known != 1 {
		t.Fatalf("UnmarshalWithOption FirstWin: got %+v, want Known=1", got)
	}

	var streamed target
	if err := json.NewDecoder(strings.NewReader(input)).DecodeWithOption(&streamed, json.DecodeFieldPriorityFirstWin()); err != nil {
		t.Fatalf("Decoder FirstWin: unexpected error: %v", err)
	}
	if streamed.Known != 1 {
		t.Fatalf("Decoder FirstWin: got %+v, want Known=1", streamed)
	}
}

func TestDecodeFirstWinRejectsInvalidSkippedObjectTail(t *testing.T) {
	type target struct {
		Known int `json:"known"`
	}
	input := `{"known":1,"skip":[1x]}`

	var std target
	if err := stdjson.Unmarshal([]byte(input), &std); err == nil {
		t.Fatal("test invariant broken: encoding/json accepted invalid skipped tail")
	}

	var got target
	if err := json.UnmarshalWithOption([]byte(input), &got, json.DecodeFieldPriorityFirstWin()); err == nil {
		t.Fatalf("UnmarshalWithOption FirstWin: expected error, got %+v", got)
	}

	var streamed target
	if err := json.NewDecoder(strings.NewReader(input)).DecodeWithOption(&streamed, json.DecodeFieldPriorityFirstWin()); err == nil {
		t.Fatalf("Decoder FirstWin: expected error, got %+v", streamed)
	}
}

func TestUnmarshalAcceptsValidSkippedNumberLiterals(t *testing.T) {
	type target struct {
		Known bool `json:"known"`
	}

	for _, lit := range validNumberLiterals {
		inputs := []string{
			`{"skip":` + lit + `,"known":true}`,
			`{"skip":{"nested":` + lit + `},"known":true}`,
			`{"skip":[` + lit + `],"known":true}`,
		}
		for _, input := range inputs {
			var std target
			if err := stdjson.Unmarshal([]byte(input), &std); err != nil {
				t.Fatalf("test invariant broken: encoding/json rejected skipped %q in %s: %v", lit, input, err)
			}

			var got target
			if err := json.Unmarshal([]byte(input), &got); err != nil {
				t.Errorf("Unmarshal skipped %q in %s: unexpected error: %v", lit, input, err)
			} else if got != std {
				t.Errorf("Unmarshal skipped %q in %s: got %+v, want %+v", lit, input, got, std)
			}

			var streamed target
			if err := json.NewDecoder(strings.NewReader(input)).Decode(&streamed); err != nil {
				t.Errorf("Decoder skipped %q in %s: unexpected error: %v", lit, input, err)
			} else if streamed != std {
				t.Errorf("Decoder skipped %q in %s: got %+v, want %+v", lit, input, streamed, std)
			}
		}
	}
}

func TestDecoderValidatesSkippedNumberAcrossBuffer(t *testing.T) {
	type target struct {
		Known bool `json:"known"`
	}

	valid := "1" + strings.Repeat("2", 700)
	validInput := `{"skip":` + valid + `,"known":true}`
	var validGot target
	if err := json.NewDecoder(strings.NewReader(validInput)).Decode(&validGot); err != nil {
		t.Fatalf("Decoder skipped long valid number: unexpected error: %v", err)
	}
	if !validGot.Known {
		t.Fatalf("Decoder skipped long valid number: got %+v", validGot)
	}

	invalid := "1" + strings.Repeat("2", 700) + "."
	invalidInput := `{"skip":` + invalid + `,"known":true}`
	var std target
	if err := stdjson.NewDecoder(strings.NewReader(invalidInput)).Decode(&std); err == nil {
		t.Fatal("test invariant broken: encoding/json accepted long invalid skipped number")
	}
	var invalidGot target
	if err := json.NewDecoder(strings.NewReader(invalidInput)).Decode(&invalidGot); err == nil {
		t.Fatalf("Decoder skipped long invalid number: expected error, got %+v", invalidGot)
	}
}
