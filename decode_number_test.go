package json_test

import (
	stdjson "encoding/json"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	json "github.com/goccy/go-json"
)

// The numbers are parsed in one pass, and a float64 is computed directly when that is exact. These tests
// compare the results with encoding/json.

type numberFields struct {
	I   int
	I8  int8
	I16 int16
	I32 int32
	I64 int64
	U   uint
	U8  uint8
	U16 uint16
	U32 uint32
	U64 uint64
	F64 float64
	A   any
}

func compareNumber(t *testing.T, field, literal string) {
	t.Helper()
	in := fmt.Sprintf(`{%q:%s}`, field, literal)
	var want, got numberFields
	errStd := stdjson.Unmarshal([]byte(in), &want)
	err := json.Unmarshal([]byte(in), &got)
	if (errStd == nil) != (err == nil) {
		t.Fatalf("%s: encoding/json: %v, go-json: %v", in, errStd, err)
	}
	if errStd != nil {
		var typeErr *json.UnmarshalTypeError
		var stdTypeErr *stdjson.UnmarshalTypeError
		if isStd, is := errorAs(errStd, &stdTypeErr), errorAs(err, &typeErr); isStd != is {
			t.Fatalf("%s: encoding/json: %v, go-json: %v", in, errStd, err)
		}
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s:\n got %+v\nwant %+v", in, got, want)
	}
}

func errorAs[T error](err error, target *T) bool {
	t, ok := err.(T)
	if ok {
		*target = t
	}
	return ok
}

func TestDecodeIntegerLimits(t *testing.T) {
	literals := []string{
		"0", "-0", "1", "-1", "7", "10", "99",
		"127", "128", "-128", "-129", "255", "256",
		"32767", "32768", "-32768", "-32769", "65535", "65536",
		"2147483647", "2147483648", "-2147483648", "-2147483649", "4294967295", "4294967296",
		"9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809",
		"18446744073709551615", "18446744073709551616", "99999999999999999999", "100000000000000000000",
		"1234567890123456789012345",
		// not integers
		"1.0", "100.0", "1e2", "1E2", "-1.5", "0.5", "1e-2",
	}
	for _, field := range []string{"I", "I8", "I16", "I32", "I64", "U", "U8", "U16", "U32", "U64"} {
		for _, literal := range literals {
			compareNumber(t, field, literal)
		}
	}
}

func TestDecodeFloatAsStrconv(t *testing.T) {
	// The float64 of a number is the one strconv.ParseFloat gives, bit for bit, whichever way it is computed.
	r := rand.New(rand.NewSource(3))
	var literals []string
	for i := 0; i < 20000; i++ {
		var f float64
		switch i % 4 {
		case 0:
			f = r.Float64() * math.Pow(10, float64(r.Intn(40)-20))
		case 1:
			f = float64(r.Int63n(1 << 53))
		case 2:
			f = math.Float64frombits(r.Uint64())
			if math.IsNaN(f) || math.IsInf(f, 0) {
				continue
			}
		case 3:
			f = float64(r.Intn(1000000)) / 1000
		}
		if r.Intn(2) == 0 {
			f = -f
		}
		for _, format := range []byte{'g', 'e', 'f'} {
			literals = append(literals, strconv.FormatFloat(f, format, -1, 64))
		}
		literals = append(literals, strconv.FormatFloat(f, 'e', r.Intn(20), 64))
	}
	literals = append(literals,
		"0", "-0", "0.0", "-0.0", "1e22", "1e23", "9007199254740992", "9007199254740993", "9007199254740993.0",
		"123456789012345678", "1234567890123456789", "12345678901234567890", "0.1", "0.30000000000000004",
		"1e-22", "1e-23", "4.9e-324", "1.7976931348623157e308", "2.2250738585072014e-308", "1E+2", "1e-0",
		"0.000000000000000000000000000001", "1000000000000000000000000.5",
	)
	for _, literal := range literals {
		want, err := strconv.ParseFloat(literal, 64)
		if err != nil {
			continue
		}
		var v struct {
			F float64
			A any
		}
		if err := json.Unmarshal([]byte(`{"F":`+literal+`,"A":`+literal+`}`), &v); err != nil {
			t.Fatalf("%s: %v", literal, err)
		}
		if math.Float64bits(v.F) != math.Float64bits(want) {
			t.Fatalf("%s: got %v ( %x ), want %v ( %x )", literal, v.F, math.Float64bits(v.F), want, math.Float64bits(want))
		}
		if a, ok := v.A.(float64); !ok || math.Float64bits(a) != math.Float64bits(want) {
			t.Fatalf("%s: into interface{}: got %v, want %v", literal, v.A, want)
		}
	}
}

func TestDecodeNumberMalformed(t *testing.T) {
	// What is not a number by the grammar is an error.
	// A fraction without a digit is still taken as a float, as before; it is left to the work on the grammar ( issue 394 ).
	lenient := map[string]bool{"1.": true, "1.e2": true}
	for _, literal := range []string{"-", "--1", "-a", "+1", "1-", "1e", "1e+", "1.e2", "1.", ".1", "0x10"} {
		for _, field := range []string{"I", "U", "F64", "A"} {
			var v numberFields
			err := json.Unmarshal([]byte(fmt.Sprintf(`{%q:%s}`, field, literal)), &v)
			if err == nil && !(lenient[literal] && (field == "F64" || field == "A")) {
				t.Fatalf("%s into %s: accepted as %+v", literal, field, v)
			}
		}
	}
}
