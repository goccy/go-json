package json_test

import (
	stdjson "encoding/json"
	"errors"
	"io"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// randomSkippedValue returns a random JSON value with white spaces, escapes and long strings anywhere, of up to
// depth levels.
func randomSkippedValue(r *rand.Rand, depth int) string {
	space := func() string {
		return []string{"", "", " ", "\n\t", "   ", "\r\n        "}[r.Intn(6)]
	}
	switch k := r.Intn(10); {
	case depth > 0 && k < 2:
		var b strings.Builder
		b.WriteString("{" + space())
		for i, n := 0, r.Intn(5); i < n; i++ {
			if i > 0 {
				b.WriteString("," + space())
			}
			b.WriteString(randomSkippedString(r) + space() + ":" + space() + randomSkippedValue(r, depth-1) + space())
		}
		return b.String() + "}"
	case depth > 0 && k < 4:
		var b strings.Builder
		b.WriteString("[" + space())
		for i, n := 0, r.Intn(5); i < n; i++ {
			if i > 0 {
				b.WriteString("," + space())
			}
			b.WriteString(randomSkippedValue(r, depth-1) + space())
		}
		return b.String() + "]"
	case k < 6:
		return randomSkippedString(r)
	case k < 8:
		return []string{"0", "-1", "12.5e-3", "1E+9", "99999999999999999999", "-0.0"}[r.Intn(6)]
	}
	return []string{"true", "false", "null"}[r.Intn(3)]
}

func randomSkippedString(r *rand.Rand) string {
	var b strings.Builder
	b.WriteByte('"')
	for i, n := 0, r.Intn(90); i < n; i++ {
		switch k := r.Intn(20); {
		case k == 0:
			b.WriteString([]string{`\"`, `\\`, `\/`, `\n`, `é`, `😀`, `\\\"`}[r.Intn(7)])
		case k == 1:
			b.WriteString([]string{"é", "あ", "{", "]", ":", ","}[r.Intn(6)])
		default:
			b.WriteByte(byte('a' + r.Intn(26)))
		}
	}
	b.WriteByte('"')
	return b.String()
}

// mutate changes a byte of s: it replaces, deletes or inserts one.
func mutate(r *rand.Rand, s string) string {
	if s == "" {
		return s
	}
	i := r.Intn(len(s))
	chars := []byte(`{}[]:,"\ 0-1e.tfnx` + "\x00\x01")
	c := chars[r.Intn(len(chars))]
	switch r.Intn(3) {
	case 0:
		return s[:i] + string(c) + s[i+1:]
	case 1:
		return s[:i] + s[i+1:]
	}
	return s[:i] + string(c) + s[i:]
}

func TestSkipGrammar(t *testing.T) {
	// A value which is skipped, as the one of a key which matches no field, is checked by the grammar of JSON as
	// encoding/json checks it: long values, which span blocks, escapes and white spaces anywhere, and values made
	// not valid by a change of a byte.
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		v := randomSkippedValue(r, 1+r.Intn(4))
		if r.Intn(2) == 0 {
			v = mutate(r, v)
		}
		doc := `{"skip":` + v + `,"X":1}`
		var got struct{ X int }
		err := json.Unmarshal([]byte(doc), &got)
		if want := stdjson.Valid([]byte(doc)); (err == nil) != want {
			t.Fatalf("%q: got error %v, valid by encoding/json %v", doc, err, want)
		}
		if err == nil && got.X != 1 {
			t.Fatalf("%q: decoded X %d", doc, got.X)
		}
	}
}

func TestSyntaxErrorMessages(t *testing.T) {
	// A syntax error is the first one of the input, with the message and the offset of encoding/json, wherever it
	// is: in a value which is decoded, in one which is skipped, or after the value.
	r := rand.New(rand.NewSource(5))
	for i := 0; i < 20000; i++ {
		doc := mutate(r, `{"skip":`+randomSkippedValue(r, 1+r.Intn(3))+`,"X":`+randomSkippedValue(r, 1+r.Intn(3))+`}`)
		for _, newValue := range []func() any{
			func() any { return new(any) },
			func() any { return new(struct{ X any }) },
		} {
			want := stdjson.Unmarshal([]byte(doc), newValue())
			var wantSyntax *stdjson.SyntaxError
			if !errors.As(want, &wantSyntax) {
				continue
			}
			got := json.Unmarshal([]byte(doc), newValue())
			var gotSyntax *json.SyntaxError
			if !errors.As(got, &gotSyntax) || gotSyntax.Error() != wantSyntax.Error() || gotSyntax.Offset != wantSyntax.Offset {
				t.Fatalf("%q:\n got %v ( %#v )\nwant %v at %d", doc, got, got, want, wantSyntax.Offset)
			}
		}
	}
}

func TestStreamSyntaxErrorMessages(t *testing.T) {
	// A syntax error of a value of a stream is the first one of the value, with the message and the offset in the
	// whole input of encoding/json, and is returned by every Decode after it. The values before it are decoded,
	// and the strings with escapes among them, which are decoded out of the buffer, keep their input.
	docs := []string{
		"1.,", "1. ", "1.", "-", "-a", "tru", "tru ", "truex", "nul,", `"abc`, `"a\q"`, "\"a\x00\"", "[1,2", `{"a":1`,
		`{"a":[1}`, `[1 2]`, `{"a" 1}`, `{"a":1,}`, "1 2 [", "1 2 x", `"a" "b\`, "\x01", `[1,2]]`, "[01]", "01",
	}
	r := rand.New(rand.NewSource(7))
	for i := 0; i < 10000; i++ {
		docs = append(docs, randomSkippedValue(r, 1+r.Intn(3))+"\n"+`{"a\"b":"c\nd","X":`+randomSkippedValue(r, 1+r.Intn(3))+`}`+" "+
			mutate(r, `{"skip":`+randomSkippedValue(r, 1+r.Intn(3))+`,"X":`+randomSkippedValue(r, 1+r.Intn(3))+`}`))
	}
	for _, doc := range docs {
		for _, newValue := range []func() any{
			func() any { return new(any) },
			func() any { return new(struct{ X any }) },
		} {
			gd := json.NewDecoder(strings.NewReader(doc))
			sd := stdjson.NewDecoder(strings.NewReader(doc))
			for n := 0; ; n++ {
				gotValue, wantValue := newValue(), newValue()
				got, want := gd.Decode(gotValue), sd.Decode(wantValue)
				if want == io.EOF || want == io.ErrUnexpectedEOF {
					if got != want {
						t.Fatalf("%q: value %d: got %v, want %v", doc, n, got, want)
					}
					break
				}
				var wantSyntax *stdjson.SyntaxError
				if !errors.As(want, &wantSyntax) {
					if want != nil {
						break
					}
					if got != nil {
						t.Fatalf("%q: value %d: got %v, want nil", doc, n, got)
					}
					if !reflect.DeepEqual(gotValue, wantValue) {
						t.Fatalf("%q: value %d: got %#v, want %#v", doc, n, gotValue, wantValue)
					}
					continue
				}
				var gotSyntax *json.SyntaxError
				if !errors.As(got, &gotSyntax) || gotSyntax.Error() != wantSyntax.Error() || gotSyntax.Offset != wantSyntax.Offset {
					t.Fatalf("%q: value %d:\n got %v ( %#v )\nwant %v at %d", doc, n, got, got, want, wantSyntax.Offset)
				}
				if again := gd.Decode(newValue()); again == nil || again.Error() != got.Error() {
					t.Fatalf("%q: Decode after %v = %v, want the same error", doc, got, again)
				}
				break
			}
		}
	}
}
