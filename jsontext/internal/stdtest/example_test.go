// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stdtest_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/goccy/go-json/jsontext"
)

// This example demonstrates the use of the [Encoder] and [Decoder] to
// parse and modify JSON without unmarshaling it into a concrete Go type.
func Example_stringReplace() {
	// Example input with non-idiomatic use of "Golang" instead of "Go".
	const input = `{
		"title": "Golang version 1 is released",
		"author": "Andrew Gerrand",
		"date": "2012-03-28",
		"text": "Today marks a major milestone in the development of the Golang programming language.",
		"otherArticles": [
			"Twelve Years of Golang",
			"The Laws of Reflection",
			"Learn Golang from your browser"
		]
	}`

	// Using a Decoder and Encoder, we can parse through every token,
	// check and modify the token if necessary, and
	// write the token to the output.
	var replacements []jsontext.Pointer
	in := strings.NewReader(input)
	dec := jsontext.NewDecoder(in)
	out := new(bytes.Buffer)
	enc := jsontext.NewEncoder(out, jsontext.Multiline(true)) // expand for readability
	for {
		// Read a token from the input.
		tok, err := dec.ReadToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		// Check whether the token contains the string "Golang" and
		// replace each occurrence with "Go" instead.
		if tok.Kind() == '"' && strings.Contains(tok.String(), "Golang") {
			replacements = append(replacements, dec.StackPointer())
			tok = jsontext.String(strings.ReplaceAll(tok.String(), "Golang", "Go"))
		}

		// Write the (possibly modified) token to the output.
		if err := enc.WriteToken(tok); err != nil {
			log.Fatal(err)
		}
	}

	// Print the list of replacements and the adjusted JSON output.
	if len(replacements) > 0 {
		fmt.Println(`Replaced "Golang" with "Go" in:`)
		for _, where := range replacements {
			fmt.Println("\t" + where)
		}
		fmt.Println()
	}
	fmt.Println("Result:", out.String())

	// Output:
	// Replaced "Golang" with "Go" in:
	// 	/title
	// 	/text
	// 	/otherArticles/0
	// 	/otherArticles/2
	//
	// Result: {
	// 	"title": "Go version 1 is released",
	// 	"author": "Andrew Gerrand",
	// 	"date": "2012-03-28",
	// 	"text": "Today marks a major milestone in the development of the Go programming language.",
	// 	"otherArticles": [
	// 		"Twelve Years of Go",
	// 		"The Laws of Reflection",
	// 		"Learn Go from your browser"
	// 	]
	// }
}
