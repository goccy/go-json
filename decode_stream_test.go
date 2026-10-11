package json_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	json "github.com/goccy/go-json"
)

// The stream decoder reads a whole value into its buffer before decoding it, so a value
// which arrives in pieces is decoded exactly as by Unmarshal, wherever the reads end.

func TestDecodeStreamMultibyteRuneSplitByRead(t *testing.T) {
	type Struct struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	body := strings.Repeat("日本語更新テスト ", 320)
	input := `{"title": "utf8 repro", "body": "` + body + `"}`
	for _, chunk := range []int{1, 2, 3, 7, 512, 1000} {
		var out Struct
		r := &chunkReader{data: []byte(input), chunk: chunk}
		if err := json.NewDecoder(r).Decode(&out); err != nil {
			t.Fatalf("chunk %d: %v", chunk, err)
		}
		if out.Body != body {
			t.Fatalf("chunk %d: the body is corrupted: %q", chunk, out.Body)
		}
	}
	// one byte per read, into a plain string
	var s string
	if err := json.NewDecoder(iotest.OneByteReader(strings.NewReader(`"é"`))).Decode(&s); err != nil {
		t.Fatal(err)
	}
	if s != "é" {
		t.Fatalf("got %q", s)
	}
}

func TestDecodeStreamTruncatedKeyWithInvalidUTF8(t *testing.T) {
	// {" followed by 1024 bytes of 0xae, no closing quote or brace ( issue 609 ).
	data := append([]byte(`{"`), bytes.Repeat([]byte{0xae}, 1024)...)
	var v any
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&v)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
	// The same key, complete: every invalid byte becomes U+FFFD, as encoding/json does.
	data = append(append([]byte(`{"`), bytes.Repeat([]byte{0xae}, 1024)...), []byte(`":1}`)...)
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&v); err != nil {
		t.Fatal(err)
	}
	m := v.(map[string]any)
	if len(m) != 1 {
		t.Fatalf("got %d keys", len(m))
	}
	for k := range m {
		if k != strings.Repeat("�", 1024) {
			t.Fatalf("the key is not replaced: %q", k[:12])
		}
	}
}

func TestDecodeStreamTokenLongStringWithInvalidUTF8(t *testing.T) {
	// A string longer than the initial buffer, with invalid bytes ( issue 424 ).
	input := []byte(`"` + strings.Repeat("0", 1000) + strings.Repeat("\xe2", 100) + `"`)
	dec := json.NewDecoder(bytes.NewReader(input))
	tok, err := dec.Token()
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Repeat("0", 1000) + strings.Repeat("�", 100)
	if tok != want {
		t.Fatalf("got %q", tok)
	}
	if _, err := dec.Token(); err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

type failingReader struct {
	data []byte
	err  error
}

func (r *failingReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, r.err
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	return n, nil
}

func TestDecodeStreamReadError(t *testing.T) {
	// The error of the reader is reported, not io.EOF ( issues 485 and 403 ).
	errLimit := errors.New("request body too large")
	var v map[string]string
	err := json.NewDecoder(&failingReader{data: []byte(`{"a":"b",`), err: errLimit}).Decode(&v)
	if !errors.Is(err, errLimit) {
		t.Fatalf("expected the error of the reader, got %v", err)
	}
	// The data before the error is decoded.
	err = json.NewDecoder(&failingReader{data: []byte(`{"a":"b"}`), err: errLimit}).Decode(&v)
	if err != nil {
		t.Fatal(err)
	}
	if v["a"] != "b" {
		t.Fatalf("got %v", v)
	}
	// io.LimitReader cuts the input: the truncated value is an error, not a value.
	body := `{"region": "test", "title":"GreatSite.com","status": "suspended"}`
	r := io.LimitReader(strings.NewReader(body), 40)
	if err := json.NewDecoder(r).Decode(&v); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}

func TestDecodeStreamInputOffset(t *testing.T) {
	// The offset counts the bytes of the input, whatever the escapes ( issue 520 ).
	input := `{"test":"\""} {"a":"é\n"}`
	dec := json.NewDecoder(strings.NewReader(input))
	if got := dec.InputOffset(); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
	var v map[string]any
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v["test"] != "\"" {
		t.Fatalf("got %v", v)
	}
	if got := dec.InputOffset(); got != 13 {
		t.Fatalf("expected 13, got %d", got)
	}
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v["a"] != "é\n" {
		t.Fatalf("got %v", v)
	}
	if got := dec.InputOffset(); got != int64(len(input)) {
		t.Fatalf("expected %d, got %d", len(input), got)
	}
}

func TestDecodeStreamErrorOffset(t *testing.T) {
	// The offset of a syntax error counts from the beginning of the input, not of the buffer.
	input := strings.Repeat(`{"a":1} `, 100) + `{"a":x}`
	dec := json.NewDecoder(strings.NewReader(input))
	var v map[string]any
	var err error
	for err == nil {
		err = dec.Decode(&v)
	}
	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("expected a syntax error, got %v", err)
	}
	if syntaxErr.Offset < int64(len(input)-7) || syntaxErr.Offset > int64(len(input)) {
		t.Fatalf("the offset %d is not in the last value ( %d bytes )", syntaxErr.Offset, len(input))
	}
}

func TestDecodeStreamValuesBackToBack(t *testing.T) {
	// Values with nothing between them, and a value cut by the end of the buffer.
	input := `0.1"hello"null true false {"a":[1,2,3]}[` + strings.Repeat("1,", 1000) + `1]"` + strings.Repeat("x", 2000) + `"`
	dec := json.NewDecoder(iotest.HalfReader(strings.NewReader(input)))
	var got []any
	for {
		var v any
		if err := dec.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if len(got) != 8 {
		t.Fatalf("got %d values", len(got))
	}
	if got[0] != 0.1 || got[1] != "hello" || got[2] != nil || got[3] != true || got[4] != false {
		t.Fatalf("got %v", got[:5])
	}
	if len(got[6].([]any)) != 1001 || got[7] != strings.Repeat("x", 2000) {
		t.Fatalf("the large values are wrong")
	}
}

func TestDecodeStreamNoReadPastValue(t *testing.T) {
	// Decode returns as soon as the value is complete: the reader is not read past it.
	pr, pw := io.Pipe()
	go func() {
		_, _ = pw.Write([]byte(`{"a": 1} `))
		// the pipe is left open: a read past the value would block
	}()
	dec := json.NewDecoder(pr)
	var v map[string]int
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v["a"] != 1 {
		t.Fatalf("got %v", v)
	}
	pw.Close()
}

func TestDecodeStreamMore(t *testing.T) {
	dec := json.NewDecoder(strings.NewReader(`{"a": [1, 2], "b": {}} 3`))
	var got []any
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, tok)
	}
	want := []any{json.Delim('{'), "a", json.Delim('['), 1.0, 2.0, json.Delim(']'), "b", json.Delim('{'), json.Delim('}'), json.Delim('}'), 3.0}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

type chunkReader struct {
	data  []byte
	chunk int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := min(len(p), r.chunk, len(r.data))
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

type countingReader struct {
	r io.Reader
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

func TestDecodeStreamUnterminatedStringStop(t *testing.T) {
	// Issue 701: An unescaped control character (< 0x20) inside a string literal
	// is invalid JSON per RFC 8259. The stream decoder must terminate scanning
	// at that offset rather than reading through the entire stream.
	tests := []struct {
		name    string
		prefix  string
		wantErr bool
	}{
		{
			name:    "unescaped newline in object string value",
			prefix:  `{"name":"alice` + "\n",
			wantErr: true,
		},
		{
			name:    "unescaped carriage return in object string value",
			prefix:  `{"name":"alice` + "\r",
			wantErr: true,
		},
		{
			name:    "unescaped raw tab in object string value",
			prefix:  `{"name":"alice` + "\t",
			wantErr: true,
		},
		{
			name:    "unescaped newline in array string element",
			prefix:  `["alice` + "\n",
			wantErr: true,
		},
		{
			name:    "valid escaped characters and formatted newlines",
			prefix:  "{\n  \"name\": \"alice\\n\\t\\r\"\n}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stream bytes.Buffer
			stream.WriteString(tt.prefix)
			for i := 0; i < 5000; i++ {
				stream.WriteString(`{"name":"alice","age":30,"address":{"city":"Springfield"}}` + "\n")
			}

			cr := &countingReader{r: bytes.NewReader(stream.Bytes())}
			var v any
			err := json.NewDecoder(cr).Decode(&v)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected decode error, got nil")
				}
				if cr.n > 1024 {
					t.Fatalf("expected decoder to read at most 1024 bytes, read %d bytes", cr.n)
				}
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Fatalf("expected SyntaxError, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Valid decode of the first value must not consume the subsequent 5000 lines.
				if cr.n > 2048 {
					t.Fatalf("expected bounded read, read %d bytes", cr.n)
				}
			}
		})
	}
}
