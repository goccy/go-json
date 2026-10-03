//go:build go1.27 && goexperiment.jsonv2

package benchmark

import (
	"bytes"
	stdjsontext "encoding/json/jsontext"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/iotest"

	"github.com/klauspost/compress/zstd"

	"github.com/goccy/go-json/jsontext"
)

// The benchmarks of jsontext do what the ones of encoding/json/v2 of Go do ( BenchmarkTestdata,
// BenchmarkSlowStreamingDecode, BenchmarkTextValue and BenchmarkAppendFormat ), on the same data, with
// github.com/goccy/go-json/jsontext ( GoJson ) and encoding/json/jsontext ( StdLib ) side by side.

// jsontextData is the JSON of the tests of encoding/json, by name, read from the source of the Go which runs the
// benchmarks, which has it compressed by zstd.
var jsontextData = sync.OnceValue(func() []jsontextEntry {
	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		panic(err)
	}
	files, err := filepath.Glob(filepath.Join(strings.TrimSpace(string(goroot)), "src", "encoding", "json", "internal", "jsontest", "_embed", "*.json.zst"))
	if err != nil {
		panic(err)
	}
	if len(files) == 0 {
		panic("no JSON data of the tests of encoding/json in the source of Go")
	}
	sort.Strings(files)
	entries := make([]jsontextEntry, 0, len(files))
	for _, file := range files {
		compressed, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		r, err := zstd.NewReader(bytes.NewReader(compressed))
		if err != nil {
			panic(err)
		}
		data, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			panic(err)
		}
		words := strings.Split(strings.TrimSuffix(filepath.Base(file), ".json.zst"), "_")
		for i, w := range words {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
		entries = append(entries, jsontextEntry{name: strings.Join(words, ""), data: data})
	}
	return entries
})

type jsontextEntry struct {
	name string
	data []byte
}

// jsontextCodec is what a benchmark does with one of the jsontext packages.
type jsontextCodec struct {
	name string
	// prepareTokens decodes data into the tokens which encodeTokens writes: strings and numbers are made by
	// String and Float, as a program which writes tokens makes them.
	prepareTokens func(data []byte) any
	encodeTokens  func(w io.Writer, tokens any) error
	encodeValue   func(w io.Writer, b []byte) error
	decodeTokens  func(r io.Reader) error
	decodeValue   func(r io.Reader) error
	isValid       func(b []byte) bool
	compact       func(b *[]byte) error
	indent        func(b *[]byte) error
	canonicalize  func(b *[]byte) error
	appendFormat  func(dst, src []byte) ([]byte, error)
}

func ignoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

var jsontextCodecs = []jsontextCodec{
	{
		name: "GoJson",
		prepareTokens: func(data []byte) any {
			var tokens []jsontext.Token
			d := jsontext.NewDecoder(bytes.NewReader(data))
			for {
				tok, err := d.ReadToken()
				if err == io.EOF {
					return tokens
				}
				if err != nil {
					panic(err)
				}
				switch tok.Kind() {
				case '"':
					tokens = append(tokens, jsontext.String(tok.String()))
				case '0':
					f, _ := tok.Float()
					tokens = append(tokens, jsontext.Float(f))
				default:
					tokens = append(tokens, tok.Clone())
				}
			}
		},
		encodeTokens: func(w io.Writer, tokens any) error {
			e := jsontext.NewEncoder(w)
			for _, tok := range tokens.([]jsontext.Token) {
				if err := e.WriteToken(tok); err != nil {
					return err
				}
			}
			return nil
		},
		encodeValue: func(w io.Writer, b []byte) error { return jsontext.NewEncoder(w).WriteValue(b) },
		decodeTokens: func(r io.Reader) error {
			d := jsontext.NewDecoder(r)
			for {
				if _, err := d.ReadToken(); err != nil {
					return ignoreEOF(err)
				}
			}
		},
		decodeValue: func(r io.Reader) error {
			_, err := jsontext.NewDecoder(r).ReadValue()
			return err
		},
		isValid:      func(b []byte) bool { return jsontext.Value(b).IsValid() },
		compact:      func(b *[]byte) error { return (*jsontext.Value)(b).Compact() },
		indent:       func(b *[]byte) error { return (*jsontext.Value)(b).Indent() },
		canonicalize: func(b *[]byte) error { return (*jsontext.Value)(b).Canonicalize() },
		appendFormat: func(dst, src []byte) ([]byte, error) { return jsontext.AppendFormat(dst, src) },
	},
	{
		name: "StdLib",
		prepareTokens: func(data []byte) any {
			var tokens []stdjsontext.Token
			d := stdjsontext.NewDecoder(bytes.NewReader(data))
			for {
				tok, err := d.ReadToken()
				if err == io.EOF {
					return tokens
				}
				if err != nil {
					panic(err)
				}
				switch tok.Kind() {
				case '"':
					tokens = append(tokens, stdjsontext.String(tok.String()))
				case '0':
					f, _ := tok.Float()
					tokens = append(tokens, stdjsontext.Float(f))
				default:
					tokens = append(tokens, tok.Clone())
				}
			}
		},
		encodeTokens: func(w io.Writer, tokens any) error {
			e := stdjsontext.NewEncoder(w)
			for _, tok := range tokens.([]stdjsontext.Token) {
				if err := e.WriteToken(tok); err != nil {
					return err
				}
			}
			return nil
		},
		encodeValue: func(w io.Writer, b []byte) error { return stdjsontext.NewEncoder(w).WriteValue(b) },
		decodeTokens: func(r io.Reader) error {
			d := stdjsontext.NewDecoder(r)
			for {
				if _, err := d.ReadToken(); err != nil {
					return ignoreEOF(err)
				}
			}
		},
		decodeValue: func(r io.Reader) error {
			_, err := stdjsontext.NewDecoder(r).ReadValue()
			return err
		},
		isValid:      func(b []byte) bool { return stdjsontext.Value(b).IsValid() },
		compact:      func(b *[]byte) error { return (*stdjsontext.Value)(b).Compact() },
		indent:       func(b *[]byte) error { return (*stdjsontext.Value)(b).Indent() },
		canonicalize: func(b *[]byte) error { return (*stdjsontext.Value)(b).Canonicalize() },
		appendFormat: func(dst, src []byte) ([]byte, error) { return stdjsontext.AppendFormat(dst, src) },
	},
}

// hiddenBuffer is a bytes.Buffer which the jsontext packages don't see as one: it is the streaming mode, where
// an Encoder writes to an io.Writer and a Decoder reads from an io.Reader through a buffer of its own.
type hiddenBuffer struct{ *bytes.Buffer }

// jsontextModes are the ways the data is given to a Decoder or taken from an Encoder: Buffered is a
// *bytes.Buffer, which the jsontext packages read and write in place.
var jsontextModes = []string{"Streaming", "Buffered"}

func jsontextWriter(mode string, buf []byte) io.Writer {
	if mode == "Streaming" {
		return hiddenBuffer{bytes.NewBuffer(buf[:0])}
	}
	return bytes.NewBuffer(buf[:0])
}

func jsontextReader(mode string, data []byte) io.Reader {
	if mode == "Streaming" {
		return hiddenBuffer{bytes.NewBuffer(data)}
	}
	return bytes.NewBuffer(data)
}

func runJsontext(b *testing.B, size int, run func(b *testing.B)) {
	b.ReportAllocs()
	b.SetBytes(int64(size))
	b.ResetTimer()
	for range b.N {
		run(b)
	}
}

var (
	goJSONJsontext = jsontextCodecs[0]
	stdJsontext    = jsontextCodecs[1]
)

// The benchmark functions are made for each library, Benchmark_Jsontext_<operation>_<library>: benchcheck
// compares only the ones of go-json with the base, and the ones of the standard library are to read next to them.

func Benchmark_Jsontext_EncodeToken_GoJson(b *testing.B) { benchJsontextEncodeToken(b, goJSONJsontext) }
func Benchmark_Jsontext_EncodeToken_StdLib(b *testing.B) { benchJsontextEncodeToken(b, stdJsontext) }
func Benchmark_Jsontext_EncodeValue_GoJson(b *testing.B) { benchJsontextEncodeValue(b, goJSONJsontext) }
func Benchmark_Jsontext_EncodeValue_StdLib(b *testing.B) { benchJsontextEncodeValue(b, stdJsontext) }
func Benchmark_Jsontext_DecodeToken_GoJson(b *testing.B) { benchJsontextDecodeToken(b, goJSONJsontext) }
func Benchmark_Jsontext_DecodeToken_StdLib(b *testing.B) { benchJsontextDecodeToken(b, stdJsontext) }
func Benchmark_Jsontext_DecodeValue_GoJson(b *testing.B) { benchJsontextDecodeValue(b, goJSONJsontext) }
func Benchmark_Jsontext_DecodeValue_StdLib(b *testing.B) { benchJsontextDecodeValue(b, stdJsontext) }
func Benchmark_Jsontext_SlowStreamingDecode_GoJson(b *testing.B) {
	benchJsontextSlowStreamingDecode(b, goJSONJsontext)
}
func Benchmark_Jsontext_SlowStreamingDecode_StdLib(b *testing.B) {
	benchJsontextSlowStreamingDecode(b, stdJsontext)
}
func Benchmark_Jsontext_Value_GoJson(b *testing.B) { benchJsontextValue(b, goJSONJsontext) }
func Benchmark_Jsontext_Value_StdLib(b *testing.B) { benchJsontextValue(b, stdJsontext) }
func Benchmark_Jsontext_AppendFormat_GoJson(b *testing.B) {
	benchJsontextAppendFormat(b, goJSONJsontext)
}
func Benchmark_Jsontext_AppendFormat_StdLib(b *testing.B) { benchJsontextAppendFormat(b, stdJsontext) }

func benchJsontextEncodeToken(b *testing.B, c jsontextCodec) {
	for _, td := range jsontextData() {
		tokens := c.prepareTokens(td.data)
		buf := make([]byte, 0, 2*len(td.data))
		for _, mode := range jsontextModes {
			b.Run(td.name+"/"+mode, func(b *testing.B) {
				runJsontext(b, len(td.data), func(b *testing.B) {
					if err := c.encodeTokens(jsontextWriter(mode, buf), tokens); err != nil {
						b.Fatal(err)
					}
				})
			})
		}
	}
}

func benchJsontextEncodeValue(b *testing.B, c jsontextCodec) {
	for _, td := range jsontextData() {
		buf := make([]byte, 0, 2*len(td.data))
		for _, mode := range jsontextModes {
			b.Run(td.name+"/"+mode, func(b *testing.B) {
				runJsontext(b, len(td.data), func(b *testing.B) {
					if err := c.encodeValue(jsontextWriter(mode, buf), td.data); err != nil {
						b.Fatal(err)
					}
				})
			})
		}
	}
}

func benchJsontextDecodeToken(b *testing.B, c jsontextCodec) {
	for _, td := range jsontextData() {
		for _, mode := range jsontextModes {
			b.Run(td.name+"/"+mode, func(b *testing.B) {
				runJsontext(b, len(td.data), func(b *testing.B) {
					if err := c.decodeTokens(jsontextReader(mode, td.data)); err != nil {
						b.Fatal(err)
					}
				})
			})
		}
	}
}

func benchJsontextDecodeValue(b *testing.B, c jsontextCodec) {
	for _, td := range jsontextData() {
		for _, mode := range jsontextModes {
			b.Run(td.name+"/"+mode, func(b *testing.B) {
				runJsontext(b, len(td.data), func(b *testing.B) {
					if err := c.decodeValue(jsontextReader(mode, td.data)); err != nil {
						b.Fatal(err)
					}
				})
			})
		}
	}
}

// benchJsontextSlowStreamingDecode reads from an io.Reader which gives one byte at a time.
func benchJsontextSlowStreamingDecode(b *testing.B, c jsontextCodec) {
	ws := strings.Repeat(" ", 4<<10)
	cases := []jsontextEntry{
		{"LargeString", []byte(`"` + strings.Repeat(" ", 4<<10) + `"`)},
		{"LargeNumber", []byte("0." + strings.Repeat("0", 4<<10))},
		{"LargeWhitespace/Null", []byte(ws + "null" + ws)},
		{"LargeWhitespace/Object", []byte(ws + "{" + ws + `"name1"` + ws + ":" + ws + `"value"` + ws + "," + ws + `"name2"` + ws + ":" + ws + `"value"` + ws + "}" + ws)},
		{"LargeWhitespace/Array", []byte(ws + "[" + ws + `"value"` + ws + "," + ws + `"value"` + ws + "]" + ws)},
	}
	for _, td := range cases {
		b.Run(td.name+"/Token", func(b *testing.B) {
			runJsontext(b, len(td.data), func(b *testing.B) {
				if err := c.decodeTokens(iotest.OneByteReader(bytes.NewReader(td.data))); err != nil {
					b.Fatal(err)
				}
			})
		})
		b.Run(td.name+"/Value", func(b *testing.B) {
			runJsontext(b, len(td.data), func(b *testing.B) {
				if err := c.decodeValue(iotest.OneByteReader(bytes.NewReader(td.data))); err != nil {
					b.Fatal(err)
				}
			})
		})
	}
}

// benchJsontextValue runs the methods of Value on CitmCatalog, as the benchmark of encoding/json/v2 does. Noop
// is a method on a value which it doesn't change: its output is its input.
func benchJsontextValue(b *testing.B, c jsontextCodec) {
	var data []byte
	for _, td := range jsontextData() {
		if td.name == "CitmCatalog" {
			data = td.data
		}
	}
	b.Run("IsValid", func(b *testing.B) {
		runJsontext(b, len(data), func(b *testing.B) {
			if !c.isValid(data) {
				b.Fatal("not valid")
			}
		})
	})
	methods := []struct {
		name   string
		format func(*[]byte) error
	}{
		{"Compact", c.compact},
		{"Indent", c.indent},
		{"Canonicalize", c.canonicalize},
	}
	var v []byte
	for _, m := range methods {
		b.Run(m.name, func(b *testing.B) {
			runJsontext(b, len(data), func(b *testing.B) {
				v = append(v[:0], data...)
				if err := m.format(&v); err != nil {
					b.Fatal(err)
				}
			})
		})
		v = append(v[:0], data...)
		if err := m.format(&v); err != nil {
			b.Fatal(err)
		}
		b.Run(m.name+"/Noop", func(b *testing.B) {
			runJsontext(b, len(data), func(b *testing.B) {
				if err := m.format(&v); err != nil {
					b.Fatal(err)
				}
			})
		})
	}
}

func benchJsontextAppendFormat(b *testing.B, c jsontextCodec) {
	input := []byte(`[ null , false , true , "fizzbuzz" , 3.14159 , { "fizz" : "buzz" } ]`)
	output := make([]byte, 0, len(input))
	runJsontext(b, len(input), func(b *testing.B) {
		var err error
		if output, err = c.appendFormat(output[:0], input); err != nil {
			b.Fatal(err)
		}
	})
}

// TestJsontextSameOutput checks that both packages write the same bytes for the data of the benchmarks, as it is
// and indented, which the benchmarks don't compare.
func TestJsontextSameOutput(t *testing.T) {
	for _, td := range jsontextData() {
		var outputs [2][]string
		for i, c := range jsontextCodecs {
			var w bytes.Buffer
			if err := c.encodeTokens(&w, c.prepareTokens(td.data)); err != nil {
				t.Fatal(err)
			}
			outputs[i] = append(outputs[i], w.String())
			indented := bytes.Clone(td.data)
			if err := c.indent(&indented); err != nil {
				t.Fatal(err)
			}
			for _, in := range [][]byte{td.data, indented} {
				for _, mode := range jsontextModes {
					w := jsontextWriter(mode, nil)
					if err := c.encodeValue(w, in); err != nil {
						t.Fatal(err)
					}
					outputs[i] = append(outputs[i], fmt.Sprint(w))
				}
				for _, format := range []func(*[]byte) error{c.compact, c.indent, c.canonicalize} {
					v := bytes.Clone(in)
					if err := format(&v); err != nil {
						t.Fatal(err)
					}
					outputs[i] = append(outputs[i], string(v))
				}
			}
		}
		for i := range outputs[0] {
			if outputs[0][i] != outputs[1][i] {
				t.Errorf("%s: output %d differs from the one of encoding/json/jsontext", td.name, i)
			}
		}
	}
}
