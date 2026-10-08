package json

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/goccy/go-json/jsontext"
)

// misuseTokens and misuseValues are what a misusing method writes: tokens and raw values, most of which don't fit
// where they are written.
var (
	misuseTokens = []jsontext.Token{jsontext.BeginObject, jsontext.EndObject, jsontext.BeginArray, jsontext.EndArray,
		jsontext.String("k"), jsontext.Int(1), jsontext.Null, jsontext.True, jsontext.String("k2")}
	misuseValues = []string{"1", "}", "]", "{", "[", `"a"`, `{"a":1}`, `[1,`, "1 2", "1{}", `{"a":`, `[1]`, "true",
		"null", `"b":1`, ",", ":"}
)

// misuseOp is a write of a misusing method: a token, a raw value, or a nested MarshalEncode of a method which
// writes the ops of sub.
type misuseOp struct {
	token, value int
	sub          []misuseOp
}

// misuseOps decodes the writes of a method from the bytes of in, which it consumes.
func misuseOps(in *[]byte, depth int) []misuseOp {
	next := func() int {
		if len(*in) == 0 {
			return 0
		}
		b := (*in)[0]
		*in = (*in)[1:]
		return int(b)
	}
	ops := make([]misuseOp, 1+next()%7)
	for i := range ops {
		switch c := next() % 10; {
		case c < 6:
			ops[i] = misuseOp{token: next() % len(misuseTokens), value: -1}
		case c < 9 || depth > 1:
			ops[i] = misuseOp{token: -1, value: next() % len(misuseValues)}
		default:
			ops[i] = misuseOp{token: -1, value: -1, sub: misuseOps(in, depth+1)}
		}
	}
	return ops
}

// misusing writes its ops to the encoder, ignoring the errors but the last.
type misusing struct{ ops []misuseOp }

func (v misusing) MarshalJSONTo(e *jsontext.Encoder) error {
	var err error
	for _, op := range v.ops {
		switch {
		case op.token >= 0:
			err = e.WriteToken(misuseTokens[op.token])
		case op.value >= 0:
			err = e.WriteValue(jsontext.Value(misuseValues[op.value]))
		default:
			err = MarshalEncode(e, misusing{op.sub})
		}
	}
	return err
}

// marshalMisusing marshals a misusing method, decoded from in with the place of its value and the options, and
// reports a panic as a failure of t.
func marshalMisusing(t *testing.T, in []byte) {
	t.Helper()
	ops := misuseOps(&in, 0)
	var place, flags int
	if len(in) >= 2 {
		place, flags = int(in[0]), int(in[1])
	}
	v := any(misusing{ops})
	switch place % 5 {
	case 1:
		v = struct{ K any }{v}
	case 2:
		v = []any{v}
	case 3:
		v = map[string]any{"K": []any{v}}
	case 4:
		v = map[string]any{"a": map[string]any{"b": []any{1, v}}}
	}
	opts := []Options{jsontext.AllowDuplicateNames(flags&1 != 0), jsontext.Multiline(flags&2 != 0)}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic %v, place %d, flags %d, ops %+v", r, place%5, flags, ops)
		}
	}()
	if flags&4 != 0 {
		_ = MarshalWrite(new(strings.Builder), v, opts...)
	} else {
		_, _ = Marshal(v, opts...)
	}
}

// A method which misuses its encoder, by any writes, makes the call fail without a panic.
func FuzzMarshalMethodMisuse(f *testing.F) {
	f.Add([]byte("000000020000000000")) // values after the end of the array which the value is in
	f.Add([]byte{2, 9, 0, 9, 1, 3, 2, 1})
	f.Fuzz(marshalMisusing)
}

// The sequences of a fixed seed, which every run of the tests checks.
func TestMarshalMethodMisuseSequences(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	in := make([]byte, 64)
	for range 20000 {
		for i := range in {
			in[i] = byte(r.Uint32())
		}
		marshalMisusing(t, in)
	}
}
