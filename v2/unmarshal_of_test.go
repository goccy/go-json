package json_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/goccy/go-json/jsontext"
	json "github.com/goccy/go-json/v2"
)

type unmarshalOfSmall struct {
	St   int    `json:"st"`
	Sid  int    `json:"sid"`
	Tt   string `json:"tt"`
	Gr   int    `json:"gr"`
	UUID string `json:"uuid"`
	IP   string `json:"ip"`
	Ua   string `json:"ua"`
	Tz   int    `json:"tz"`
	V    int    `json:"v"`
}

const unmarshalOfSmallInput = `{"st": 1,"sid": 486,"tt": "active","gr": 0,"uuid": "de305d54-75b4-431b-adb2-eb6b9e546014","ip": "127.0.0.1","ua": "user_agent","tz": -6,"v": 1}`

// unmarshalOfSmallStrings is the number of the strings of unmarshalOfSmallInput: tt, uuid, ip and ua.
const unmarshalOfSmallStrings = 4

type unmarshalOfLarge struct {
	A    [600]int
	Name string
}

// unmarshalOfFrom is decoded by UnmarshalJSONFrom, which is called on the copy of the value.
type unmarshalOfFrom struct{ N int }

func (u *unmarshalOfFrom) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	v, err := dec.ReadValue()
	u.N = len(v)
	return err
}

// unmarshalOfCase decodes the input by UnmarshalOf and by Unmarshal into two values which start the same.
func unmarshalOfCase[T any](t *testing.T, input string, start T, opts ...json.Options) {
	t.Helper()
	got, want := start, start
	errOf := json.UnmarshalOf([]byte(input), &got, opts...)
	err := json.Unmarshal([]byte(input), &want, opts...)
	if (errOf == nil) != (err == nil) || (err != nil && errOf.Error() != err.Error()) {
		t.Fatalf("%s: UnmarshalOf: %v, Unmarshal: %v", input, errOf, err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s: UnmarshalOf: %#v, Unmarshal: %#v", input, got, want)
	}
}

func TestUnmarshalOfAsUnmarshal(t *testing.T) {
	unmarshalOfCase(t, unmarshalOfSmallInput, unmarshalOfSmall{})
	// into a value which exists: the fields absent from the input are kept.
	unmarshalOfCase(t, `{"st":5}`, unmarshalOfSmall{Tt: "kept", V: 9})
	unmarshalOfCase(t, `{"A":[1,2],"M":{"x":"y"},"P":{"N":3}}`, struct {
		A []int
		M map[string]string
		P *struct{ N int }
	}{})
	unmarshalOfCase(t, `[1,"a",{"b":null}]`, any(nil))
	unmarshalOfCase(t, `"2026-09-24T00:00:00Z"`, time.Time{})
	unmarshalOfCase(t, `123`, 0)
	unmarshalOfCase(t, `"s"`, "")
	unmarshalOfCase(t, `{"x":1}`, map[string]int{"y": 2})
	unmarshalOfCase(t, `[1,2,3]`, []int{})
	unmarshalOfCase(t, `{"Name":"large","A":[1,2,3]}`, unmarshalOfLarge{})
	unmarshalOfCase(t, `[[1],{"a":2}]`, []unmarshalOfFrom{})
	unmarshalOfCase(t, `{"ST":1}`, unmarshalOfSmall{}, json.MatchCaseInsensitiveNames(true))
	// errors, with what was decoded before them
	unmarshalOfCase(t, `{"st":1,"tt":2}`, unmarshalOfSmall{})
	unmarshalOfCase(t, `{"st":1} x`, unmarshalOfSmall{})
	unmarshalOfCase(t, `{"st":`, unmarshalOfSmall{})
	unmarshalOfCase(t, `{"st":1,"st":2}`, unmarshalOfSmall{})

	var nilPtr *unmarshalOfSmall
	var serr *json.SemanticError
	if err := json.UnmarshalOf([]byte(`{}`), nilPtr); !errors.As(err, &serr) || serr.GoType != reflect.TypeFor[*unmarshalOfSmall]() {
		t.Fatalf("expected a SemanticError of the pointer type, got %v", err)
	}
	if err, want := json.UnmarshalOf([]byte(`{}`), nilPtr), json.Unmarshal([]byte(`{}`), nilPtr); err.Error() != want.Error() {
		t.Fatalf("UnmarshalOf: %v, Unmarshal: %v", err, want)
	}
}

func TestUnmarshalOfOtherTypesByTurns(t *testing.T) {
	// The values reused by a context are of the type decoded last: other types come by turns.
	for range 20 {
		var a unmarshalOfSmall
		if err := json.UnmarshalOf([]byte(unmarshalOfSmallInput), &a); err != nil || a.UUID != "de305d54-75b4-431b-adb2-eb6b9e546014" {
			t.Fatalf("got %+v, %v", a, err)
		}
		var b struct{ X []string }
		if err := json.UnmarshalOf([]byte(`{"X":["p","q"]}`), &b); err != nil || !reflect.DeepEqual(b.X, []string{"p", "q"}) {
			t.Fatalf("got %+v, %v", b, err)
		}
		// the reused value is zero: nothing of the previous call is decoded into
		var c unmarshalOfSmall
		if err := json.UnmarshalOf([]byte(`{"st":1}`), &c); err != nil || c != (unmarshalOfSmall{St: 1}) {
			t.Fatalf("got %+v, %v", c, err)
		}
	}
}

func TestUnmarshalOfAllocations(t *testing.T) {
	if raceEnabled {
		t.Skip("sync.Pool drops items at random under the race detector")
	}
	data := []byte(unmarshalOfSmallInput)
	var v unmarshalOfSmall
	allocs := testing.AllocsPerRun(1000, func() {
		var w unmarshalOfSmall
		if err := json.UnmarshalOf(data, &w); err != nil {
			t.Fatal(err)
		}
		v = w
	})
	// a string is a copy of its own, and nothing else is allocated
	if allocs != unmarshalOfSmallStrings {
		t.Fatalf("%v allocations per call, want %v", allocs, unmarshalOfSmallStrings)
	}
	if v.Ua != "user_agent" {
		t.Fatalf("got %+v", v)
	}
}
