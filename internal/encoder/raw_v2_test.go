package encoder_test

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/jsontext"
)

// rawFlags are the options of the engine of the v2 semantics: its escaping of the strings.
const rawFlags = encoder.V2Option | encoder.TextEscapeOption | encoder.RejectInvalidUTF8Option

// appendFormattedRaw is encoder.AppendFormattedRaw with the escaping of flags.
func appendFormattedRaw(src string, flags encoder.OptionFlag) (string, bool) {
	ctx := encoder.TakeRuntimeContext()
	defer encoder.ReleaseRuntimeContext(ctx)
	ctx.Option.Flag = flags
	out, ok := encoder.AppendFormattedRaw(ctx, []byte("prefix"), []byte(src))
	if !ok {
		return "", false
	}
	return strings.TrimPrefix(string(out), "prefix"), true
}

// checkFormattedRaw fails if AppendFormattedRaw appends src but not as jsontext formats it with the options.
func checkFormattedRaw(t *testing.T, src string, flags encoder.OptionFlag) bool {
	t.Helper()
	got, ok := appendFormattedRaw(src, flags)
	if !ok {
		return false
	}
	opts := []jsontext.Options{
		jsontext.EscapeForHTML(flags&encoder.HTMLEscapeOption != 0),
		jsontext.EscapeForJS(flags&encoder.NormalizeUTF8Option != 0),
	}
	want, err := jsontext.AppendFormat(nil, src, opts...)
	if err != nil || got != string(want) {
		t.Fatalf("AppendFormattedRaw(%q) = %q; AppendFormat = %q, %v", src, got, want, err)
	}
	return true
}

// A raw value which AppendFormattedRaw appends is the one which jsontext formats.
func TestAppendFormattedRaw(t *testing.T) {
	deep := strings.Repeat("[", 65) + strings.Repeat("]", 65)
	var many strings.Builder
	many.WriteString("{")
	for i := range 70 {
		if i > 0 {
			many.WriteString(",")
		}
		many.WriteString(`"` + strings.Repeat("k", i+1) + `":0`)
	}
	many.WriteString("}")
	for _, tt := range []struct {
		in   string
		want bool
	}{
		{`null`, true},
		{`true`, true},
		{`-1.5e+10`, true},
		{`"text"`, true},
		{`"日本語"`, true},
		{`{"a":1,"b":[1,"x",{"a":2}],"c":{}}`, true},
		{`[{"a":1},{"a":1}]`, true},
		{`{"a":1,"aa":2,"a":3}`, false},
		{`{"a":{"b":1},"b":2,"a":3}`, false},
		{`{"ab":1,"a":2}`, true},
		{`{"a":1, "b":2}`, false},
		{` 1`, false},
		{`"a\"b"`, true},
		{`"\u0026 \u00e9 \/ \n \u001f \t"`, true},
		{`["cd x \u0026\u0026 go test \u003c"]`, true},
		{`"\ud83d\ude00"`, false},
		{`"\x"`, false},
		{`{"a\"":1}`, false},
		{"\"a\tb\"", false},
		{"\"\xff\"", false},
		{`"a<b"`, true},
		{"\" \"", true},
		{`"a long string of ascii only, longer than words"`, true},
		{`"a long string of ascii with a \" quote escaped"`, true},
		{"\"a long string of ascii with a \x01 control character\"", false},
		{"\"a long string of ascii with a \x7f delete character\"", true},
		{`"a long string with 日本語 and more after it"`, true},
		{"\"a long string with \xe6\x97 broken UTF-8 in it here\"", false},
		{"\"escaped \\n and broken \xe6\x97 UTF-8\"", false},
		{"{\"name with \xff\":1}", false},
		{`{"a long name of the object which is long":1,"a long name of the object which is long":2}`, false},
		{`01`, false},
		{`{"a"}`, false},
		{`[1,]`, false},
		{`tru`, false},
		{``, false},
		{deep, false},
		{many.String(), false},
	} {
		if got := checkFormattedRaw(t, tt.in, rawFlags); got != tt.want {
			t.Errorf("AppendFormattedRaw(%q) appended = %v; want %v", tt.in, got, tt.want)
		}
	}
	for _, tt := range []struct {
		in    string
		flags encoder.OptionFlag
		want  bool
	}{
		{`"a<b"`, encoder.HTMLEscapeOption, false},
		{`"a&b"`, encoder.HTMLEscapeOption | encoder.NormalizeUTF8Option, false},
		{"\" \"", encoder.NormalizeUTF8Option, false},
		{`"ab \u0026"`, encoder.HTMLEscapeOption | encoder.NormalizeUTF8Option, true},
	} {
		if got := checkFormattedRaw(t, tt.in, rawFlags|tt.flags); got != tt.want {
			t.Errorf("AppendFormattedRaw(%q, %b) appended = %v; want %v", tt.in, tt.flags, got, tt.want)
		}
	}

	// random values of JSON, which may have duplicate names, escapes, control characters and invalid UTF-8.
	r := rand.New(rand.NewPCG(1, 2))
	pieces := []string{"a", "b", "ab", "日本", `\n`, `\"`, `\u0026`, `\u00e9`, `\ud800`, "\x01", "\xff", "<", " ", strings.Repeat("x", 9)}
	str := func() string {
		var s strings.Builder
		s.WriteByte('"')
		for range r.IntN(4) {
			s.WriteString(pieces[r.IntN(len(pieces))])
		}
		s.WriteByte('"')
		return s.String()
	}
	var value func(depth int) string
	value = func(depth int) string {
		switch k := r.IntN(6); {
		case depth > 3 || k == 0:
			return []string{"1", "-2.5e3", "true", "null", "01"}[r.IntN(5)]
		case k == 1:
			return str()
		case k == 2:
			var s strings.Builder
			s.WriteByte('[')
			for i := range r.IntN(4) {
				if i > 0 {
					s.WriteByte(',')
				}
				s.WriteString(value(depth + 1))
			}
			s.WriteByte(']')
			return s.String()
		default:
			var s strings.Builder
			s.WriteByte('{')
			for i := range r.IntN(5) {
				if i > 0 {
					s.WriteByte(',')
				}
				s.WriteString(str() + ":" + value(depth+1))
			}
			s.WriteByte('}')
			return s.String()
		}
	}
	appended := 0
	flagSets := []encoder.OptionFlag{0, encoder.HTMLEscapeOption, encoder.NormalizeUTF8Option}
	for i := range 100000 {
		if checkFormattedRaw(t, value(0), rawFlags|flagSets[i%len(flagSets)]) {
			appended++
		}
	}
	if appended < 1000 {
		t.Fatalf("only %d random values are appended", appended)
	}
}
