package jsontext

import (
	"fmt"
	"strings"
)

// Options configures NewEncoder, Encoder.Reset, NewDecoder and Decoder.Reset, and the functions and methods
// which format values. Each takes a list of options, in which an option set later overrides a property set by
// an earlier one.
//
// A single Options type is used for encoding and decoding: an option which affects only one of them is ignored
// by the other.
//
//   - AllowDuplicateNames affects encoding and decoding
//   - AllowInvalidUTF8 affects encoding and decoding
//   - EscapeForHTML affects encoding only
//   - EscapeForJS affects encoding only
//   - PreserveRawStrings affects encoding only
//   - CanonicalizeRawInts affects encoding only
//   - CanonicalizeRawFloats affects encoding only
//   - ReorderRawObjects affects encoding only
//   - SpaceAfterColon affects encoding only
//   - SpaceAfterComma affects encoding only
//   - Multiline affects encoding only
//   - WithIndent affects encoding only
//   - WithIndentPrefix affects encoding only
type Options interface {
	applyTo(c *config)
}

// flag is a property of a config which is true or false, or, for indentSet and prefixSet, which is set.
type flag uint32

const (
	allowDuplicateNames flag = 1 << iota
	allowInvalidUTF8
	escapeForHTML
	escapeForJS
	preserveRawStrings
	canonicalizeRawInts
	canonicalizeRawFloats
	reorderRawObjects
	spaceAfterColon
	spaceAfterComma
	multiline
	indentSet
	prefixSet
	// omitTopLevelNewline is set by the functions and methods which format a single value, whose output has no
	// line feed after the value, as an Encoder writes.
	omitTopLevelNewline
)

// config is the state of the options: the properties which were set, and their values.
type config struct {
	set    flag // the properties which an option set
	value  flag // the values of the properties which were set, the others are false
	indent string
	prefix string
}

func (c *config) applyTo(dst *config) {
	dst.set |= c.set
	dst.value = dst.value&^c.set | c.value
	if c.set&indentSet != 0 {
		dst.indent = c.indent
	}
	if c.set&prefixSet != 0 {
		dst.prefix = c.prefix
	}
}

// apply sets the options in order.
func (c *config) apply(opts []Options) {
	for _, o := range opts {
		if o != nil {
			o.applyTo(c)
		}
	}
}

// has reports whether the property f is true.
func (c *config) has(f flag) bool { return c.value&f != 0 }

// whitespace is the white space which an encoder writes under c.
type whitespace struct {
	colon, comma bool // a space after a colon and after a comma
	multiline    bool
	indent       string
	prefix       string
}

func (c *config) whitespace() whitespace {
	w := whitespace{multiline: c.has(multiline), comma: c.has(spaceAfterComma)}
	// a multiline output has a space after a colon unless it was set otherwise.
	w.colon = c.has(spaceAfterColon) || w.multiline && c.set&spaceAfterColon == 0
	w.indent, w.prefix = "\t", c.prefix
	if c.set&indentSet != 0 {
		w.indent = c.indent
	}
	return w
}

// boolOption is an option of a flag: the flag shifted by one bit, with the value as the lowest bit. Its values
// are constants, so that an Options of it is made without an allocation.
type boolOption flag

func (o boolOption) applyTo(c *config) {
	f := flag(o) >> 1
	c.set |= f
	c.value &^= f
	if o&1 != 0 {
		c.value |= f
	}
}

func boolOf(f flag, v bool) Options {
	if v {
		return boolOption(f<<1 | 1)
	}
	return boolOption(f << 1)
}

// AllowDuplicateNames specifies that JSON objects may contain duplicate member names. Disabling the check of
// duplicate names may make encoding and decoding faster, but breaks compliance with RFC 7493, section 2.3.
// The input or output still complies with RFC 8259, which leaves the handling of duplicate names unspecified.
//
// This affects encoding and decoding.
func AllowDuplicateNames(v bool) Options { return boolOf(allowDuplicateNames, v) }

// AllowInvalidUTF8 specifies that JSON strings may contain invalid UTF-8, which is mangled as the Unicode
// replacement character, U+FFFD. This breaks compliance with RFC 7493, section 2.1, and RFC 8259, section 8.1.
//
// This affects encoding and decoding.
func AllowInvalidUTF8(v bool) Options { return boolOf(allowInvalidUTF8, v) }

// EscapeForHTML specifies that the characters '<', '>' and '&' in JSON strings are escaped as hexadecimal
// Unicode code points (e.g., <), so that the output is safe to embed within HTML.
//
// This affects encoding only.
func EscapeForHTML(v bool) Options { return boolOf(escapeForHTML, v) }

// EscapeForJS specifies that the characters U+2028 and U+2029 in JSON strings are escaped as hexadecimal
// Unicode code points (e.g.,  ), so that the output is valid to embed within JavaScript. See RFC 8259,
// section 12.
//
// This affects encoding only.
func EscapeForJS(v bool) Options { return boolOf(escapeForJS, v) }

// PreserveRawStrings specifies that the escape sequences of a raw JSON string in a Token or a Value are written
// as they are. The characters which EscapeForHTML and EscapeForJS escape are still escaped. If AllowInvalidUTF8
// is set, the bytes of invalid UTF-8 are written as they are too.
//
// This affects encoding only.
func PreserveRawStrings(v bool) Options { return boolOf(preserveRawStrings, v) }

// CanonicalizeRawInts specifies that a raw JSON integer (a number without a fraction and an exponent) in a
// Token or a Value is written in the canonical form of RFC 8785, section 3.2.2.3; -0 is written as 0.
//
// The numbers are IEEE 754 double precision numbers: an integer which needs more precision loses it. For
// example, integers beyond ±2⁵³ lose their precision: 1234567890123456789 is written as 1234567890123456800.
//
// This affects encoding only.
func CanonicalizeRawInts(v bool) Options { return boolOf(canonicalizeRawInts, v) }

// CanonicalizeRawFloats specifies that a raw JSON floating-point number (a number with a fraction or an
// exponent) in a Token or a Value is written in the canonical form of RFC 8785, section 3.2.2.3; -0 is written
// as 0.
//
// The numbers are IEEE 754 double precision numbers. A single precision number which is canonicalized is read
// back as the same single precision number. A number beyond ±1.7976931348623157e+308, the largest finite
// number, is saturated at it.
//
// This affects encoding only.
func CanonicalizeRawFloats(v bool) Options { return boolOf(canonicalizeRawFloats, v) }

// ReorderRawObjects specifies that the members of a raw JSON object in a Value are written in the order of RFC
// 8785, section 3.2.3.
//
// This affects encoding only.
func ReorderRawObjects(v bool) Options { return boolOf(reorderRawObjects, v) }

// SpaceAfterColon specifies that a space is written after the colon which follows a JSON object name.
//
// This affects encoding only.
func SpaceAfterColon(v bool) Options { return boolOf(spaceAfterColon, v) }

// SpaceAfterComma specifies that a space is written after the comma which follows a JSON object value or array
// element.
//
// This affects encoding only.
func SpaceAfterComma(v bool) Options { return boolOf(spaceAfterComma, v) }

// Multiline specifies that the output is written on multiple lines: every JSON object member and array element
// is on a new line, indented by its depth.
//
// Unless they are specified, SpaceAfterColon is true, SpaceAfterComma is false, and the indentation of
// WithIndent is "\t".
//
// If it is false, the output is a single line, whose only white space is the one of SpaceAfterColon and
// SpaceAfterComma.
//
// This affects encoding only.
func Multiline(v bool) Options { return boolOf(multiline, v) }

// indentOption is the option of WithIndent or WithIndentPrefix, which implies Multiline.
type indentOption struct {
	s      string
	prefix bool
}

func (o indentOption) applyTo(c *config) {
	c.set |= multiline
	c.value |= multiline
	if o.prefix {
		c.set |= prefixSet
		c.prefix = o.s
	} else {
		c.set |= indentSet
		c.indent = o.s
	}
}

// WithIndent specifies that the output is written on multiple lines, where each JSON object member and array
// element starts a line with the prefix of WithIndentPrefix and then indent once for each level of nesting.
// The indent must consist of spaces and tabs only.
//
// To write multiline output without a preference for the indentation, use Multiline.
//
// This affects encoding only. It implies Multiline(true).
func WithIndent(indent string) Options {
	checkIndent(indent, "indent")
	return indentOption{s: indent}
}

// WithIndentPrefix specifies that the output is written on multiple lines, where each JSON object member and
// array element starts a line with prefix and then the indentation of WithIndent once for each level of
// nesting. The prefix must consist of spaces and tabs only.
//
// This affects encoding only. It implies Multiline(true).
func WithIndentPrefix(prefix string) Options {
	checkIndent(prefix, "indent prefix")
	return indentOption{s: prefix, prefix: true}
}

func checkIndent(s, what string) {
	if i := strings.IndexFunc(s, func(r rune) bool { return r != ' ' && r != '\t' }); i >= 0 {
		panic(fmt.Sprintf("json: invalid character %s in %s", quoteChar([]byte(s[i:])), what))
	}
}
