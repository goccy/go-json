package json_test

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	json "github.com/goccy/go-json"
)

// firstWinReference rewrites the JSON document so that every object keeps only the first of its keys which are
// the same regardless of case: what DecodeFieldPriorityFirstWin decodes is what encoding/json decodes then.
func firstWinReference(doc string) (string, error) {
	dec := stdjson.NewDecoder(strings.NewReader(doc))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return "", err
	}
	// decode again by tokens, which keep the order and the duplicates of the keys
	dec = stdjson.NewDecoder(strings.NewReader(doc))
	dec.UseNumber()
	var b bytes.Buffer
	var write func() error
	write = func() error {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case stdjson.Delim:
			if t == '[' {
				b.WriteByte('[')
				for i := 0; dec.More(); i++ {
					if i > 0 {
						b.WriteByte(',')
					}
					if err := write(); err != nil {
						return err
					}
				}
				_, err := dec.Token()
				b.WriteByte(']')
				return err
			}
			b.WriteByte('{')
			seen := map[string]bool{}
			first := true
			for dec.More() {
				k, err := dec.Token()
				if err != nil {
					return err
				}
				key := k.(string)
				folded := strings.ToLower(key)
				if seen[folded] {
					// the value of a later key is skipped
					var skip stdjson.RawMessage
					if err := dec.Decode(&skip); err != nil {
						return err
					}
					continue
				}
				seen[folded] = true
				if !first {
					b.WriteByte(',')
				}
				first = false
				kb, _ := stdjson.Marshal(key)
				b.Write(kb)
				b.WriteByte(':')
				if err := write(); err != nil {
					return err
				}
			}
			_, err := dec.Token()
			b.WriteByte('}')
			return err
		default:
			vb, err := stdjson.Marshal(t)
			b.Write(vb)
			return err
		}
	}
	if err := write(); err != nil {
		return "", err
	}
	return b.String(), nil
}

// firstWinStruct returns a struct type of n int fields, whose keys are f0, f1, ..., and a field "nested" of a
// struct of three fields, whose keys are a, b and c.
func firstWinStruct(n int) reflect.Type {
	nested := reflect.StructOf([]reflect.StructField{
		{Name: "A", Type: reflect.TypeOf(0), Tag: `json:"a"`},
		{Name: "B", Type: reflect.TypeOf(""), Tag: `json:"b"`},
		{Name: "C", Type: reflect.TypeOf([]int(nil)), Tag: `json:"c"`},
	})
	fields := []reflect.StructField{{Name: "Nested", Type: nested, Tag: `json:"nested"`}}
	for i := 0; i < n; i++ {
		fields = append(fields, reflect.StructField{Name: fmt.Sprintf("F%d", i), Type: reflect.TypeOf(0), Tag: reflect.StructTag(fmt.Sprintf(`json:"f%d"`, i))})
	}
	return reflect.StructOf(fields)
}

func TestDecodeFirstWin(t *testing.T) {
	// Under DecodeFieldPriorityFirstWin, the first of the keys of an object which match a field, regardless of
	// case, is decoded, and the value of a later one is skipped, and validated: for structs of up to 64 fields,
	// whose fields are remembered in a word, and of more.
	r := rand.New(rand.NewSource(23))
	randomCase := func(s string) string {
		b := []byte(s)
		for i := range b {
			if r.Intn(2) == 0 {
				b[i] = bytes.ToUpper(b[i : i+1])[0]
			}
		}
		return string(b)
	}
	for _, n := range []int{1, 3, 9, 63, 64, 65, 127, 128, 130} {
		typ := firstWinStruct(n)
		for i := 0; i < 300; i++ {
			var b strings.Builder
			b.WriteByte('{')
			keys := r.Intn(2*n + 4)
			for k := 0; k < keys; k++ {
				if k > 0 {
					b.WriteByte(',')
				}
				switch r.Intn(12) {
				case 0:
					b.WriteString(`"unknown":{"x":[1,2,"y"]}`)
				case 1:
					fmt.Fprintf(&b, `"%s":{"%s":%d,"%s":"s%d","A":%d,"c":[%d]}`, randomCase("nested"), randomCase("a"), r.Intn(100), randomCase("b"), k, r.Intn(100), k)
				case 2:
					if r.Intn(4) == 0 {
						// an invalid value, which is an error wherever it is
						fmt.Fprintf(&b, `"f%d":[1,,2]`, r.Intn(n))
						continue
					}
					fallthrough
				default:
					fmt.Fprintf(&b, `"%s":%d`, randomCase(fmt.Sprintf("f%d", r.Intn(n))), k)
				}
			}
			b.WriteByte('}')
			doc := b.String()
			ref, refErr := firstWinReference(doc)
			want := reflect.New(typ)
			if refErr == nil {
				if err := stdjson.Unmarshal([]byte(ref), want.Interface()); err != nil {
					t.Fatalf("encoding/json: %s: %v", ref, err)
				}
			}
			check := func(name string, got reflect.Value, err error) {
				t.Helper()
				if refErr != nil {
					if err == nil {
						t.Fatalf("%s, %d fields: %s: no error, encoding/json: %v", name, n, doc, refErr)
					}
					return
				}
				if err != nil {
					t.Fatalf("%s, %d fields: %s: %v", name, n, doc, err)
				}
				if !reflect.DeepEqual(got.Elem().Interface(), want.Elem().Interface()) {
					t.Fatalf("%s, %d fields: %s:\n got %+v\nwant %+v", name, n, doc, got.Elem().Interface(), want.Elem().Interface())
				}
			}
			got := reflect.New(typ)
			check("Unmarshal", got, json.UnmarshalWithOption([]byte(doc), got.Interface(), json.DecodeFieldPriorityFirstWin()))
			got = reflect.New(typ)
			check("Decoder", got, json.NewDecoder(strings.NewReader(doc)).DecodeWithOption(got.Interface(), json.DecodeFieldPriorityFirstWin()))
		}
	}
}
