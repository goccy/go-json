package benchmark

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// The payloads of the benchmarks of the APIs ( decode_github_test.go, llm_api_test.go ) are generated when the
// benchmarks start, instead of being files of the repository: their keys, the order of the keys, their types
// and nulls are the ones of the APIs, and their values are made by a generator of a fixed seed, so that they are
// the same at every run and at every commit which a benchmark compares. The texts have the lengths, the escapes
// and the characters which are not ASCII of the ones of the APIs: markdown, Go source code and messages.

// payloadRand is a generator of pseudo-random numbers ( SplitMix64 ), whose sequence is fixed by its seed.
type payloadRand struct {
	state uint64
}

func newPayloadRand(seed uint64) *payloadRand {
	return &payloadRand{state: seed}
}

func (r *payloadRand) next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// intn returns a number in [0, n).
func (r *payloadRand) intn(n int) int {
	return int(r.next() % uint64(n))
}

// between returns a number in [min, max].
func (r *payloadRand) between(min, max int) int {
	return min + r.intn(max-min+1)
}

// chance is true once in n.
func (r *payloadRand) chance(n int) bool {
	return r.intn(n) == 0
}

func (r *payloadRand) pick(words []string) string {
	return words[r.intn(len(words))]
}

// token returns n characters of the alphabet, as the identifiers, the signatures and the cursors of the APIs.
func (r *payloadRand) token(alphabet string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[r.intn(len(alphabet))]
	}
	return string(b)
}

const (
	alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	base64URL    = alphanumeric + "-_"
	hexDigits    = "0123456789abcdef"
)

// object is a JSON object whose keys are in their order, as the APIs write them.
type object []field

type field struct {
	key   string
	value any
}

func (o object) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, f := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		b.Write(encodePayload(f.key))
		b.WriteByte(':')
		b.Write(encodePayload(f.value))
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// encodePayload encodes the value as the APIs do: '<', '>' and '&' are not escaped.
func encodePayload(v any) []byte {
	var b bytes.Buffer
	enc := stdjson.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return bytes.TrimSuffix(b.Bytes(), []byte("\n"))
}

// jsonLines encodes the events of a stream, one JSON document each.
func jsonLines(events []object) [][]byte {
	lines := make([][]byte, len(events))
	for i, e := range events {
		lines[i] = encodePayload(e)
	}
	return lines
}

// timestamp is the time of the APIs, in RFC 3339 and UTC.
func timestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// The words of the texts: the ones of the issues and the pull requests of a library, and of a coding agent.
var (
	nouns = strings.Fields(`decoder encoder struct field key value string number buffer cursor byte escape table
		hash map slice pointer type cache benchmark allocation loop word mask scan parser token stream reader
		writer error test case path option compiler function call register branch layout runtime memory copy
		length offset index array object literal rune quote backslash character input output result request
		response message tool content model session event delta chunk payload schema property interface method
		receiver generic release commit change regression fix issue report platform architecture version`)
	verbs = strings.Fields(`decodes encodes returns reads writes skips scans copies checks compares allocates
		reuses keeps handles escapes folds matches parses validates stores loads calls avoids measures reports
		breaks fixes adds removes moves splits merges replaces caches`)
	adjectives = strings.Fields(`short long plain escaped unknown nested empty large small first last next
		previous invalid valid exact loose fast slow hot cold new old same other whole`)
	smallWords = strings.Fields(`the a of to in is and for with by on at from as it that this when which not
		only each every more than then so if but or`)
	nonASCIIPunctuation = []string{"—", "→", "…", "✓", "×", "≤"}
)

// sentence returns a sentence of prose, which has code, numbers and references of issues now and then.
func sentence(r *payloadRand) string {
	n := r.between(6, 18)
	words := make([]string, 0, n)
	for i := 0; i < n; i++ {
		var w string
		switch k := r.intn(20); {
		case k < 7:
			w = r.pick(smallWords)
		case k < 12:
			w = r.pick(nouns)
		case k < 15:
			w = r.pick(verbs)
		case k < 17:
			w = r.pick(adjectives)
		case k == 17:
			w = "`" + identifier(r) + "`"
		case k == 18:
			w = fmt.Sprintf("%d", r.between(2, 4096))
		default:
			w = fmt.Sprintf("(#%d)", r.between(100, 659))
		}
		words = append(words, w)
	}
	if r.chance(4) {
		words[r.intn(n)] += ","
	}
	if r.chance(6) {
		i := r.intn(n)
		words[i] = `"` + words[i] + `"`
	}
	if r.chance(12) {
		words[r.intn(n)] += " " + r.pick(nonASCIIPunctuation)
	}
	s := strings.Join(words, " ")
	return strings.ToUpper(s[:1]) + s[1:] + "."
}

// prose returns paragraphs of about n bytes.
func prose(r *payloadRand, n int) string {
	var b strings.Builder
	for b.Len() < n {
		if b.Len() > 0 {
			if r.chance(3) {
				b.WriteString("\n\n")
			} else {
				b.WriteByte(' ')
			}
		}
		b.WriteString(sentence(r))
	}
	return b.String()
}

// markdown returns a text of about n bytes as the ones of issues, pull requests and answers of a model:
// headings, lists with bold text, paragraphs and blocks of code.
func markdown(r *payloadRand, n int) string {
	return markdownWithCode(r, n, 10)
}

// markdownWithCode is markdown whose parts are blocks of code once in codeOneIn.
func markdownWithCode(r *payloadRand, n, codeOneIn int) string {
	var b strings.Builder
	for b.Len() < n {
		if b.Len() > 0 {
			b.WriteString("\n\n")
		}
		if r.chance(codeOneIn) {
			b.WriteString("```go\n" + goSource(r, r.between(120, 600)) + "```")
			continue
		}
		switch k := r.intn(9); {
		case k < 2:
			b.WriteString("## " + strings.ToUpper(r.pick(nouns)[:1]) + r.pick(nouns)[1:] + " " + r.pick(nouns))
		case k < 6:
			for i, items := 0, r.between(2, 6); i < items; i++ {
				if i > 0 {
					b.WriteByte('\n')
				}
				if i > 0 && r.chance(3) {
					b.WriteString("  - " + sentence(r))
				} else {
					b.WriteString("- **" + sentence(r) + "** " + sentence(r))
				}
			}
		default:
			b.WriteString(prose(r, r.between(80, 400)))
		}
	}
	return b.String()
}

// identifier returns a name of Go code.
func identifier(r *payloadRand) string {
	s := r.pick(verbs)
	s = strings.TrimSuffix(s, "s") + strings.ToUpper(r.pick(nouns)[:1])
	return s + r.pick(nouns)[1:]
}

// goSource returns Go source code of about n bytes, whose indentation, strings and characters are the ones of
// the files which a coding agent reads: its escapes in JSON are its newlines, tabs, quotes and backslashes.
func goSource(r *payloadRand, n int) string {
	var b strings.Builder
	for b.Len() < n {
		name := identifier(r)
		fmt.Fprintf(&b, "// %s %s the %s %s of the %s.\n", name, r.pick(verbs), r.pick(adjectives), r.pick(nouns), r.pick(nouns))
		fmt.Fprintf(&b, "func (d *%sDecoder) %s(buf []byte, cursor int64) (int64, error) {\n", r.pick(nouns), name)
		depth := 1
		for i, lines := 0, r.between(4, 16); i < lines; i++ {
			indent := strings.Repeat("\t", depth)
			switch k := r.intn(9); k {
			case 0:
				fmt.Fprintf(&b, "%sfor %s < len(buf) {\n", indent, r.pick(nouns))
				depth++
			case 1:
				fmt.Fprintf(&b, "%sswitch buf[cursor] {\n%scase ' ', '\\n', '\\t', '\\r':\n", indent, indent)
				depth++
			case 2:
				fmt.Fprintf(&b, "%sreturn 0, errors.ErrSyntax(fmt.Sprintf(\"invalid character %%q in %s\", buf[cursor]), cursor)\n", indent, r.pick(nouns))
			case 3:
				fmt.Fprintf(&b, "%s%s := %s(buf[cursor:], %q)\n", indent, r.pick(nouns), identifier(r), r.pick(nouns))
			case 4:
				fmt.Fprintf(&b, "%sif %s == nil {\n", indent, r.pick(nouns))
				depth++
			default:
				fmt.Fprintf(&b, "%s%s = %s(%s, %d)\n", indent, r.pick(nouns), identifier(r), r.pick(nouns), r.between(0, 64))
			}
			if depth > 1 && r.chance(3) {
				depth--
				fmt.Fprintf(&b, "%s}\n", strings.Repeat("\t", depth))
			}
		}
		for depth > 1 {
			depth--
			fmt.Fprintf(&b, "%s}\n", strings.Repeat("\t", depth))
		}
		b.WriteString("\treturn cursor, nil\n}\n\n")
	}
	return b.String()
}

// chunks splits the text into the pieces of 1 to 7 bytes which a stream delivers, on the boundaries of runes.
func chunks(r *payloadRand, text string) []string {
	var pieces []string
	for len(text) > 0 {
		n := min(r.between(1, 7), len(text))
		for n < len(text) && !utf8.RuneStart(text[n]) {
			n++
		}
		pieces = append(pieces, text[:n])
		text = text[n:]
	}
	return pieces
}
