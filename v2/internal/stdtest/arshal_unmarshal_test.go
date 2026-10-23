// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stdtest_test

import (
	"bytes"
	"encoding"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/netip"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json/jsontext"
	. "github.com/goccy/go-json/v2"
)

// unmarshalTest is a case of TestUnmarshal.
type unmarshalTest struct {
	name    CaseName
	opts    []Options
	inBuf   string
	inVal   any
	want    any
	wantErr error
	skip    bool
}

// unmarshalTests returns the cases of TestUnmarshal, made anew for each test which decodes them, as the decoding
// writes their values.
func unmarshalTests() []unmarshalTest {
	return []unmarshalTest{{
		name:    Name("Nil"),
		inBuf:   `null`,
		wantErr: EU(errNonNilReference),
	}, {
		name:    Name("NilPointer"),
		inBuf:   `null`,
		inVal:   (*string)(nil),
		want:    (*string)(nil),
		wantErr: EU(errNonNilReference).withType(0, T[*string]()),
	}, {
		name:    Name("NonPointer"),
		inBuf:   `null`,
		inVal:   "unchanged",
		want:    "unchanged",
		wantErr: EU(errNonNilReference).withType(0, T[string]()),
	}, {
		name:    Name("Bools/TrailingJunk"),
		inBuf:   `falsetrue`,
		inVal:   addr(true),
		want:    addr(false),
		wantErr: newInvalidCharacterError("t", "after top-level value", len64(`false`), ""),
	}, {
		name:  Name("Bools/Null"),
		inBuf: `null`,
		inVal: addr(true),
		want:  addr(false),
	}, {
		name:  Name("Bools"),
		inBuf: `[null,false,true]`,
		inVal: new([]bool),
		want:  addr([]bool{false, false, true}),
	}, {
		name:  Name("Bools/Named"),
		inBuf: `[null,false,true]`,
		inVal: new([]namedBool),
		want:  addr([]namedBool{false, false, true}),
	}, {
		name:    Name("Bools/Invalid/StringifiedFalse"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"false"`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(nil).withType('"', boolType),
	}, {
		name:    Name("Bools/Invalid/StringifiedTrue"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"true"`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(nil).withType('"', boolType),
	}, {
		name:  Name("Bools/StringifiedBool/True"),
		opts:  []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		inBuf: `"true"`,
		inVal: addr(false),
		want:  addr(true),
	}, {
		name:  Name("Bools/StringifiedBool/False"),
		opts:  []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		inBuf: `"false"`,
		inVal: addr(true),
		want:  addr(false),
	}, {
		name:    Name("Bools/StringifiedBool/InvalidWhitespace"),
		opts:    []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		inBuf:   `"false "`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(strconv.ErrSyntax).withVal(`"false "`).withType('"', boolType),
	}, {
		name:    Name("Bools/StringifiedBool/InvalidBool"),
		opts:    []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		inBuf:   `false`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(nil).withType('f', boolType),
	}, {
		name:    Name("Bools/Invalid/Number"),
		inBuf:   `0`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(nil).withType('0', boolType),
	}, {
		name:    Name("Bools/Invalid/String"),
		inBuf:   `""`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(nil).withType('"', boolType),
	}, {
		name:    Name("Bools/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(nil).withType('{', boolType),
	}, {
		name:    Name("Bools/Invalid/Array"),
		inBuf:   `[]`,
		inVal:   addr(true),
		want:    addr(true),
		wantErr: EU(nil).withType('[', boolType),
	}, {
		name:  Name("Bools/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `false`,
		inVal: addr(true),
		want:  addr(false),
	}, {
		name:  Name("Strings/Null"),
		inBuf: `null`,
		inVal: addr("something"),
		want:  addr(""),
	}, {
		name:  Name("Strings"),
		inBuf: `[null,"","hello","世界"]`,
		inVal: new([]string),
		want:  addr([]string{"", "", "hello", "世界"}),
	}, {
		name:  Name("Strings/Escaped"),
		inBuf: `[null,"","\u0068\u0065\u006c\u006c\u006f","\u4e16\u754c"]`,
		inVal: new([]string),
		want:  addr([]string{"", "", "hello", "世界"}),
	}, {
		name:  Name("Strings/Named"),
		inBuf: `[null,"","hello","世界"]`,
		inVal: new([]namedString),
		want:  addr([]namedString{"", "", "hello", "世界"}),
	}, {
		name:    Name("Strings/Invalid/False"),
		inBuf:   `false`,
		inVal:   addr("nochange"),
		want:    addr("nochange"),
		wantErr: EU(nil).withType('f', stringType),
	}, {
		name:    Name("Strings/Invalid/True"),
		inBuf:   `true`,
		inVal:   addr("nochange"),
		want:    addr("nochange"),
		wantErr: EU(nil).withType('t', stringType),
	}, {
		name:    Name("Strings/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr("nochange"),
		want:    addr("nochange"),
		wantErr: EU(nil).withType('{', stringType),
	}, {
		name:    Name("Strings/Invalid/Array"),
		inBuf:   `[]`,
		inVal:   addr("nochange"),
		want:    addr("nochange"),
		wantErr: EU(nil).withType('[', stringType),
	}, {
		name:  Name("Strings/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `"hello"`,
		inVal: addr("goodbye"),
		want:  addr("hello"),
	}, {
		name:  Name("Strings/StringifiedString"),
		opts:  []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		inBuf: `"\"foo\""`,
		inVal: new(string),
		want:  addr("foo"),
	}, {
		name:    Name("Strings/StringifiedString/InvalidWhitespace"),
		opts:    []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		inBuf:   `"\"foo\" "`,
		inVal:   new(string),
		want:    new(string),
		wantErr: EU(newInvalidCharacterError(" ", "after string value", 0, "")).withType('"', stringType),
	}, {
		name:    Name("Strings/StringifiedString/InvalidString"),
		opts:    []Options{unsupported("StringTag|StringifyWithLegacySemantics")},
		inBuf:   `""`,
		inVal:   new(string),
		want:    new(string),
		wantErr: EU(&jsontext.SyntacticError{Err: io.ErrUnexpectedEOF}).withType('"', stringType),
	}, {
		name:  Name("Bytes/Null"),
		inBuf: `null`,
		inVal: addr([]byte("something")),
		want:  addr([]byte(nil)),
	}, {
		name:  Name("Bytes"),
		inBuf: `[null,"","AQ==","AQI=","AQID"]`,
		inVal: new([][]byte),
		want:  addr([][]byte{nil, {}, {1}, {1, 2}, {1, 2, 3}}),
	}, {
		name:  Name("Bytes/Large"),
		inBuf: `"dGhlIHF1aWNrIGJyb3duIGZveCBqdW1wZWQgb3ZlciB0aGUgbGF6eSBkb2cgYW5kIGF0ZSB0aGUgaG9tZXdvcmsgdGhhdCBJIHNwZW50IHNvIG11Y2ggdGltZSBvbi4="`,
		inVal: new([]byte),
		want:  addr([]byte("the quick brown fox jumped over the lazy dog and ate the homework that I spent so much time on.")),
	}, {
		name:  Name("Bytes/Reuse"),
		inBuf: `"AQID"`,
		inVal: addr([]byte("changed")),
		want:  addr([]byte{1, 2, 3}),
	}, {
		name:  Name("Bytes/Escaped"),
		inBuf: `[null,"","\u0041\u0051\u003d\u003d","\u0041\u0051\u0049\u003d","\u0041\u0051\u0049\u0044"]`,
		inVal: new([][]byte),
		want:  addr([][]byte{nil, {}, {1}, {1, 2}, {1, 2, 3}}),
	}, {
		name:  Name("Bytes/Named"),
		inBuf: `[null,"","AQ==","AQI=","AQID"]`,
		inVal: new([]namedBytes),
		want:  addr([]namedBytes{nil, {}, {1}, {1, 2}, {1, 2, 3}}),
	}, {
		name:  Name("Bytes/NotStringified"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `[null,"","AQ==","AQI=","AQID"]`,
		inVal: new([][]byte),
		want:  addr([][]byte{nil, {}, {1}, {1, 2}, {1, 2, 3}}),
	}, {
		// NOTE: []namedByte is not assignable to []byte,
		// so the following should be treated as a slice of uints.
		name:  Name("Bytes/Invariant"),
		inBuf: `[null,[],[1],[1,2],[1,2,3]]`,
		inVal: new([][]namedByte),
		want:  addr([][]namedByte{nil, {}, {1}, {1, 2}, {1, 2, 3}}),
	}, {
		// NOTE: This differs in behavior from v1,
		// but keeps the representation of slices and arrays more consistent.
		name:  Name("Bytes/ByteArray"),
		inBuf: `"aGVsbG8="`,
		inVal: new([5]byte),
		want:  addr([5]byte{'h', 'e', 'l', 'l', 'o'}),
	}, {
		name:  Name("Bytes/ByteArray0/Valid"),
		inBuf: `""`,
		inVal: new([0]byte),
		want:  addr([0]byte{}),
	}, {
		name:  Name("Bytes/ByteArray0/Invalid"),
		inBuf: `"A"`,
		inVal: new([0]byte),
		want:  addr([0]byte{}),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 0), []byte("A"))
			return err
		}()).withType('"', T[[0]byte]()),
	}, {
		name:    Name("Bytes/ByteArray0/Overflow"),
		inBuf:   `"AA=="`,
		inVal:   new([0]byte),
		want:    addr([0]byte{}),
		wantErr: EU(errors.New("decoded length of 1 mismatches array length of 0")).withType('"', T[[0]byte]()),
	}, {
		name:  Name("Bytes/ByteArray1/Valid"),
		inBuf: `"AQ=="`,
		inVal: new([1]byte),
		want:  addr([1]byte{1}),
	}, {
		name:  Name("Bytes/ByteArray1/Invalid"),
		inBuf: `"$$=="`,
		inVal: new([1]byte),
		want:  addr([1]byte{}),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 1), []byte("$$=="))
			return err
		}()).withType('"', T[[1]byte]()),
	}, {
		name:    Name("Bytes/ByteArray1/Underflow"),
		inBuf:   `""`,
		inVal:   new([1]byte),
		want:    addr([1]byte{}),
		wantErr: EU(errors.New("decoded length of 0 mismatches array length of 1")).withType('"', T[[1]byte]()),
	}, {
		name:    Name("Bytes/ByteArray1/Overflow"),
		inBuf:   `"AQI="`,
		inVal:   new([1]byte),
		want:    addr([1]byte{1}),
		wantErr: EU(errors.New("decoded length of 2 mismatches array length of 1")).withType('"', T[[1]byte]()),
	}, {
		name:  Name("Bytes/ByteArray2/Valid"),
		inBuf: `"AQI="`,
		inVal: new([2]byte),
		want:  addr([2]byte{1, 2}),
	}, {
		name:  Name("Bytes/ByteArray2/Invalid"),
		inBuf: `"$$$="`,
		inVal: new([2]byte),
		want:  addr([2]byte{}),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 2), []byte("$$$="))
			return err
		}()).withType('"', T[[2]byte]()),
	}, {
		name:    Name("Bytes/ByteArray2/Underflow"),
		inBuf:   `"AQ=="`,
		inVal:   new([2]byte),
		want:    addr([2]byte{1, 0}),
		wantErr: EU(errors.New("decoded length of 1 mismatches array length of 2")).withType('"', T[[2]byte]()),
	}, {
		name:  Name("Bytes/ByteArray2/Underflow/Allowed"),
		opts:  []Options{unsupported("UnmarshalArrayFromAnyLength")},
		inBuf: `"AQ=="`,
		inVal: new([2]byte),
		want:  addr([2]byte{1, 0}),
	}, {
		name:    Name("Bytes/ByteArray2/Overflow"),
		inBuf:   `"AQID"`,
		inVal:   new([2]byte),
		want:    addr([2]byte{1, 2}),
		wantErr: EU(errors.New("decoded length of 3 mismatches array length of 2")).withType('"', T[[2]byte]()),
	}, {
		name:  Name("Bytes/ByteArray2/Overflow/Allowed"),
		opts:  []Options{unsupported("UnmarshalArrayFromAnyLength")},
		inBuf: `"AQID"`,
		inVal: new([2]byte),
		want:  addr([2]byte{1, 2}),
	}, {
		name:  Name("Bytes/ByteArray3/Valid"),
		inBuf: `"AQID"`,
		inVal: new([3]byte),
		want:  addr([3]byte{1, 2, 3}),
	}, {
		name:  Name("Bytes/ByteArray3/Invalid"),
		inBuf: `"$$$$"`,
		inVal: new([3]byte),
		want:  addr([3]byte{}),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 3), []byte("$$$$"))
			return err
		}()).withType('"', T[[3]byte]()),
	}, {
		name:    Name("Bytes/ByteArray3/Underflow"),
		inBuf:   `"AQI="`,
		inVal:   addr([3]byte{0xff, 0xff, 0xff}),
		want:    addr([3]byte{1, 2, 0}),
		wantErr: EU(errors.New("decoded length of 2 mismatches array length of 3")).withType('"', T[[3]byte]()),
	}, {
		name:    Name("Bytes/ByteArray3/Overflow"),
		inBuf:   `"AQIDAQ=="`,
		inVal:   new([3]byte),
		want:    addr([3]byte{1, 2, 3}),
		wantErr: EU(errors.New("decoded length of 4 mismatches array length of 3")).withType('"', T[[3]byte]()),
	}, {
		name:  Name("Bytes/ByteArray4/Valid"),
		inBuf: `"AQIDBA=="`,
		inVal: new([4]byte),
		want:  addr([4]byte{1, 2, 3, 4}),
	}, {
		name:  Name("Bytes/ByteArray4/Invalid"),
		inBuf: `"$$$$$$=="`,
		inVal: new([4]byte),
		want:  addr([4]byte{}),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 4), []byte("$$$$$$=="))
			return err
		}()).withType('"', T[[4]byte]()),
	}, {
		name:    Name("Bytes/ByteArray4/Underflow"),
		inBuf:   `"AQID"`,
		inVal:   new([4]byte),
		want:    addr([4]byte{1, 2, 3, 0}),
		wantErr: EU(errors.New("decoded length of 3 mismatches array length of 4")).withType('"', T[[4]byte]()),
	}, {
		name:    Name("Bytes/ByteArray4/Overflow"),
		inBuf:   `"AQIDBAU="`,
		inVal:   new([4]byte),
		want:    addr([4]byte{1, 2, 3, 4}),
		wantErr: EU(errors.New("decoded length of 5 mismatches array length of 4")).withType('"', T[[4]byte]()),
	}, {
		// NOTE: []namedByte is not assignable to []byte,
		// so the following should be treated as a array of uints.
		name:  Name("Bytes/NamedByteArray"),
		inBuf: `[104,101,108,108,111]`,
		inVal: new([5]namedByte),
		want:  addr([5]namedByte{'h', 'e', 'l', 'l', 'o'}),
	}, {
		name:  Name("Bytes/Valid/Denormalized"),
		inBuf: `"AR=="`,
		inVal: new([]byte),
		want:  addr([]byte{1}),
	}, {
		name:  Name("Bytes/Invalid/Unpadded1"),
		inBuf: `"AQ="`,
		inVal: addr([]byte("nochange")),
		want:  addr([]byte("nochange")),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 0), []byte("AQ="))
			return err
		}()).withType('"', bytesType),
	}, {
		name:  Name("Bytes/Invalid/Unpadded2"),
		inBuf: `"AQ"`,
		inVal: addr([]byte("nochange")),
		want:  addr([]byte("nochange")),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 0), []byte("AQ"))
			return err
		}()).withType('"', bytesType),
	}, {
		name:  Name("Bytes/Invalid/Character"),
		inBuf: `"@@@@"`,
		inVal: addr([]byte("nochange")),
		want:  addr([]byte("nochange")),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 3), []byte("@@@@"))
			return err
		}()).withType('"', bytesType),
	}, {
		name:    Name("Bytes/Invalid/Bool"),
		inBuf:   `true`,
		inVal:   addr([]byte("nochange")),
		want:    addr([]byte("nochange")),
		wantErr: EU(nil).withType('t', bytesType),
	}, {
		name:    Name("Bytes/Invalid/Number"),
		inBuf:   `0`,
		inVal:   addr([]byte("nochange")),
		want:    addr([]byte("nochange")),
		wantErr: EU(nil).withType('0', bytesType),
	}, {
		name:    Name("Bytes/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr([]byte("nochange")),
		want:    addr([]byte("nochange")),
		wantErr: EU(nil).withType('{', bytesType),
	}, {
		name:    Name("Bytes/Invalid/Array"),
		inBuf:   `[]`,
		inVal:   addr([]byte("nochange")),
		want:    addr([]byte("nochange")),
		wantErr: EU(nil).withType('[', bytesType),
	}, {
		name:  Name("Bytes/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `"aGVsbG8="`,
		inVal: new([]byte),
		want:  addr([]byte("hello")),
	}, {
		name:  Name("Ints/Null"),
		inBuf: `null`,
		inVal: addr(int(1)),
		want:  addr(int(0)),
	}, {
		name:  Name("Ints/Int"),
		inBuf: `1`,
		inVal: addr(int(0)),
		want:  addr(int(1)),
	}, {
		name:    Name("Ints/Int8/MinOverflow"),
		inBuf:   `-129`,
		inVal:   addr(int8(-1)),
		want:    addr(int8(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`-129`).withType('0', T[int8]()),
	}, {
		name:  Name("Ints/Int8/Min"),
		inBuf: `-128`,
		inVal: addr(int8(0)),
		want:  addr(int8(-128)),
	}, {
		name:  Name("Ints/Int8/Max"),
		inBuf: `127`,
		inVal: addr(int8(0)),
		want:  addr(int8(127)),
	}, {
		name:    Name("Ints/Int8/MaxOverflow"),
		inBuf:   `128`,
		inVal:   addr(int8(-1)),
		want:    addr(int8(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`128`).withType('0', T[int8]()),
	}, {
		name:    Name("Ints/Int16/MinOverflow"),
		inBuf:   `-32769`,
		inVal:   addr(int16(-1)),
		want:    addr(int16(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`-32769`).withType('0', T[int16]()),
	}, {
		name:  Name("Ints/Int16/Min"),
		inBuf: `-32768`,
		inVal: addr(int16(0)),
		want:  addr(int16(-32768)),
	}, {
		name:  Name("Ints/Int16/Max"),
		inBuf: `32767`,
		inVal: addr(int16(0)),
		want:  addr(int16(32767)),
	}, {
		name:    Name("Ints/Int16/MaxOverflow"),
		inBuf:   `32768`,
		inVal:   addr(int16(-1)),
		want:    addr(int16(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`32768`).withType('0', T[int16]()),
	}, {
		name:    Name("Ints/Int32/MinOverflow"),
		inBuf:   `-2147483649`,
		inVal:   addr(int32(-1)),
		want:    addr(int32(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`-2147483649`).withType('0', T[int32]()),
	}, {
		name:  Name("Ints/Int32/Min"),
		inBuf: `-2147483648`,
		inVal: addr(int32(0)),
		want:  addr(int32(-2147483648)),
	}, {
		name:  Name("Ints/Int32/Max"),
		inBuf: `2147483647`,
		inVal: addr(int32(0)),
		want:  addr(int32(2147483647)),
	}, {
		name:    Name("Ints/Int32/MaxOverflow"),
		inBuf:   `2147483648`,
		inVal:   addr(int32(-1)),
		want:    addr(int32(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`2147483648`).withType('0', T[int32]()),
	}, {
		name:    Name("Ints/Int64/MinOverflow"),
		inBuf:   `-9223372036854775809`,
		inVal:   addr(int64(-1)),
		want:    addr(int64(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`-9223372036854775809`).withType('0', T[int64]()),
	}, {
		name:  Name("Ints/Int64/Min"),
		inBuf: `-9223372036854775808`,
		inVal: addr(int64(0)),
		want:  addr(int64(-9223372036854775808)),
	}, {
		name:  Name("Ints/Int64/Max"),
		inBuf: `9223372036854775807`,
		inVal: addr(int64(0)),
		want:  addr(int64(9223372036854775807)),
	}, {
		name:    Name("Ints/Int64/MaxOverflow"),
		inBuf:   `9223372036854775808`,
		inVal:   addr(int64(-1)),
		want:    addr(int64(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`9223372036854775808`).withType('0', T[int64]()),
	}, {
		name:  Name("Ints/Named"),
		inBuf: `-6464`,
		inVal: addr(namedInt64(0)),
		want:  addr(namedInt64(-6464)),
	}, {
		name:  Name("Ints/Stringified"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `"-6464"`,
		inVal: new(int),
		want:  addr(int(-6464)),
	}, {
		name:    Name("Ints/Stringified/Invalid"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `-6464`,
		inVal:   new(int),
		want:    new(int),
		wantErr: EU(nil).withType('0', T[int]()),
	}, {
		name:    Name("Ints/Stringified/LeadingZero"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"00"`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"00"`).withType('"', T[int]()),
	}, {
		name:  Name("Ints/Escaped"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `"\u002d\u0036\u0034\u0036\u0034"`,
		inVal: new(int),
		want:  addr(int(-6464)),
	}, {
		name:  Name("Ints/Valid/NegativeZero"),
		inBuf: `-0`,
		inVal: addr(int(1)),
		want:  addr(int(0)),
	}, {
		name:    Name("Ints/Invalid/Fraction"),
		inBuf:   `1.0`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`1.0`).withType('0', T[int]()),
	}, {
		name:    Name("Ints/Invalid/Exponent"),
		inBuf:   `1e0`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`1e0`).withType('0', T[int]()),
	}, {
		name:    Name("Ints/Invalid/StringifiedFraction"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"1.0"`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"1.0"`).withType('"', T[int]()),
	}, {
		name:    Name("Ints/Invalid/StringifiedExponent"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"1e0"`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"1e0"`).withType('"', T[int]()),
	}, {
		name:    Name("Ints/Invalid/Overflow"),
		inBuf:   `100000000000000000000000000000`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrRange).withVal(`100000000000000000000000000000`).withType('0', T[int]()),
	}, {
		name:    Name("Ints/Invalid/OverflowSyntax"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"100000000000000000000000000000x"`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"100000000000000000000000000000x"`).withType('"', T[int]()),
	}, {
		name:    Name("Ints/Invalid/Whitespace"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"0 "`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"0 "`).withType('"', T[int]()),
	}, {
		name:    Name("Ints/Invalid/Bool"),
		inBuf:   `true`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(nil).withType('t', T[int]()),
	}, {
		name:    Name("Ints/Invalid/String"),
		inBuf:   `"0"`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(nil).withType('"', T[int]()),
	}, {
		name:    Name("Ints/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(nil).withType('{', T[int]()),
	}, {
		name:    Name("Ints/Invalid/Array"),
		inBuf:   `[]`,
		inVal:   addr(int(-1)),
		want:    addr(int(-1)),
		wantErr: EU(nil).withType('[', T[int]()),
	}, {
		name:  Name("Ints/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `1`,
		inVal: addr(int(0)),
		want:  addr(int(1)),
	}, {
		name:  Name("Uints/Null"),
		inBuf: `null`,
		inVal: addr(uint(1)),
		want:  addr(uint(0)),
	}, {
		name:  Name("Uints/Uint"),
		inBuf: `1`,
		inVal: addr(uint(0)),
		want:  addr(uint(1)),
	}, {
		name:  Name("Uints/Uint8/Min"),
		inBuf: `0`,
		inVal: addr(uint8(1)),
		want:  addr(uint8(0)),
	}, {
		name:  Name("Uints/Uint8/Max"),
		inBuf: `255`,
		inVal: addr(uint8(0)),
		want:  addr(uint8(255)),
	}, {
		name:    Name("Uints/Uint8/MaxOverflow"),
		inBuf:   `256`,
		inVal:   addr(uint8(1)),
		want:    addr(uint8(1)),
		wantErr: EU(strconv.ErrRange).withVal(`256`).withType('0', T[uint8]()),
	}, {
		name:  Name("Uints/Uint16/Min"),
		inBuf: `0`,
		inVal: addr(uint16(1)),
		want:  addr(uint16(0)),
	}, {
		name:  Name("Uints/Uint16/Max"),
		inBuf: `65535`,
		inVal: addr(uint16(0)),
		want:  addr(uint16(65535)),
	}, {
		name:    Name("Uints/Uint16/MaxOverflow"),
		inBuf:   `65536`,
		inVal:   addr(uint16(1)),
		want:    addr(uint16(1)),
		wantErr: EU(strconv.ErrRange).withVal(`65536`).withType('0', T[uint16]()),
	}, {
		name:  Name("Uints/Uint32/Min"),
		inBuf: `0`,
		inVal: addr(uint32(1)),
		want:  addr(uint32(0)),
	}, {
		name:  Name("Uints/Uint32/Max"),
		inBuf: `4294967295`,
		inVal: addr(uint32(0)),
		want:  addr(uint32(4294967295)),
	}, {
		name:    Name("Uints/Uint32/MaxOverflow"),
		inBuf:   `4294967296`,
		inVal:   addr(uint32(1)),
		want:    addr(uint32(1)),
		wantErr: EU(strconv.ErrRange).withVal(`4294967296`).withType('0', T[uint32]()),
	}, {
		name:  Name("Uints/Uint64/Min"),
		inBuf: `0`,
		inVal: addr(uint64(1)),
		want:  addr(uint64(0)),
	}, {
		name:  Name("Uints/Uint64/Max"),
		inBuf: `18446744073709551615`,
		inVal: addr(uint64(0)),
		want:  addr(uint64(18446744073709551615)),
	}, {
		name:    Name("Uints/Uint64/MaxOverflow"),
		inBuf:   `18446744073709551616`,
		inVal:   addr(uint64(1)),
		want:    addr(uint64(1)),
		wantErr: EU(strconv.ErrRange).withVal(`18446744073709551616`).withType('0', T[uint64]()),
	}, {
		name:  Name("Uints/Uintptr"),
		inBuf: `1`,
		inVal: addr(uintptr(0)),
		want:  addr(uintptr(1)),
	}, {
		name:  Name("Uints/Named"),
		inBuf: `6464`,
		inVal: addr(namedUint64(0)),
		want:  addr(namedUint64(6464)),
	}, {
		name:  Name("Uints/Stringified"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `"6464"`,
		inVal: new(uint),
		want:  addr(uint(6464)),
	}, {
		name:    Name("Uints/Stringified/Invalid"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `6464`,
		inVal:   new(uint),
		want:    new(uint),
		wantErr: EU(nil).withType('0', T[uint]()),
	}, {
		name:    Name("Uints/Stringified/LeadingZero"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"00"`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"00"`).withType('"', T[uint]()),
	}, {
		name:  Name("Uints/Escaped"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `"\u0036\u0034\u0036\u0034"`,
		inVal: new(uint),
		want:  addr(uint(6464)),
	}, {
		name:    Name("Uints/Invalid/NegativeOne"),
		inBuf:   `-1`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`-1`).withType('0', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/NegativeZero"),
		inBuf:   `-0`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`-0`).withType('0', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/Fraction"),
		inBuf:   `1.0`,
		inVal:   addr(uint(10)),
		want:    addr(uint(10)),
		wantErr: EU(strconv.ErrSyntax).withVal(`1.0`).withType('0', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/Exponent"),
		inBuf:   `1e0`,
		inVal:   addr(uint(10)),
		want:    addr(uint(10)),
		wantErr: EU(strconv.ErrSyntax).withVal(`1e0`).withType('0', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/StringifiedFraction"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"1.0"`,
		inVal:   addr(uint(10)),
		want:    addr(uint(10)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"1.0"`).withType('"', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/StringifiedExponent"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"1e0"`,
		inVal:   addr(uint(10)),
		want:    addr(uint(10)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"1e0"`).withType('"', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/Overflow"),
		inBuf:   `100000000000000000000000000000`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(strconv.ErrRange).withVal(`100000000000000000000000000000`).withType('0', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/OverflowSyntax"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"100000000000000000000000000000x"`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"100000000000000000000000000000x"`).withType('"', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/Whitespace"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"0 "`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"0 "`).withType('"', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/Bool"),
		inBuf:   `true`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(nil).withType('t', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/String"),
		inBuf:   `"0"`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(nil).withType('"', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(nil).withType('{', T[uint]()),
	}, {
		name:    Name("Uints/Invalid/Array"),
		inBuf:   `[]`,
		inVal:   addr(uint(1)),
		want:    addr(uint(1)),
		wantErr: EU(nil).withType('[', T[uint]()),
	}, {
		name:  Name("Uints/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `1`,
		inVal: addr(uint(0)),
		want:  addr(uint(1)),
	}, {
		name:  Name("Floats/Null"),
		inBuf: `null`,
		inVal: addr(float64(64.64)),
		want:  addr(float64(0)),
	}, {
		name:  Name("Floats/Float32/Pi"),
		inBuf: `3.14159265358979323846264338327950288419716939937510582097494459`,
		inVal: addr(float32(32.32)),
		want:  addr(float32(math.Pi)),
	}, {
		name:  Name("Floats/Float32/Underflow"),
		inBuf: `1e-1000`,
		inVal: addr(float32(32.32)),
		want:  addr(float32(0)),
	}, {
		name:    Name("Floats/Float32/Overflow"),
		inBuf:   `-1e1000`,
		inVal:   addr(float32(32.32)),
		want:    addr(float32(math.Inf(-1))),
		wantErr: EU(strconv.ErrRange).withVal(`-1e1000`).withType('0', T[float32]()),
	}, {
		name:  Name("Floats/Float64/Pi"),
		inBuf: `3.14159265358979323846264338327950288419716939937510582097494459`,
		inVal: addr(float64(64.64)),
		want:  addr(float64(math.Pi)),
	}, {
		name:  Name("Floats/Float64/Underflow"),
		inBuf: `1e-1000`,
		inVal: addr(float64(64.64)),
		want:  addr(float64(0)),
	}, {
		name:    Name("Floats/Float64/Overflow"),
		inBuf:   `-1e1000`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(math.Inf(-1))),
		wantErr: EU(strconv.ErrRange).withVal(`-1e1000`).withType('0', T[float64]()),
	}, {
		name:    Name("Floats/Any/Overflow"),
		inBuf:   `1e1000`,
		inVal:   new(any),
		want:    addr(any(float64(math.Inf(+1)))),
		wantErr: EU(strconv.ErrRange).withVal(`1e1000`).withType('0', T[float64]()),
	}, {
		name:  Name("Floats/Named"),
		inBuf: `64.64`,
		inVal: addr(namedFloat64(0)),
		want:  addr(namedFloat64(64.64)),
	}, {
		name:  Name("Floats/Stringified"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `"64.64"`,
		inVal: new(float64),
		want:  addr(float64(64.64)),
	}, {
		name:    Name("Floats/Stringified/Invalid"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `64.64`,
		inVal:   new(float64),
		want:    new(float64),
		wantErr: EU(nil).withType('0', T[float64]()),
	}, {
		name:  Name("Floats/Escaped"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `"\u0036\u0034\u002e\u0036\u0034"`,
		inVal: new(float64),
		want:  addr(float64(64.64)),
	}, {
		name:    Name("Floats/Invalid/NaN"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"NaN"`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"NaN"`).withType('"', float64Type),
	}, {
		name:    Name("Floats/Invalid/Infinity"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"Infinity"`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"Infinity"`).withType('"', float64Type),
	}, {
		name:    Name("Floats/Invalid/Whitespace"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"1 "`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"1 "`).withType('"', float64Type),
	}, {
		name:    Name("Floats/Invalid/GoSyntax"),
		opts:    []Options{StringifyNumbers(true)},
		inBuf:   `"1p-2"`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(strconv.ErrSyntax).withVal(`"1p-2"`).withType('"', float64Type),
	}, {
		name:    Name("Floats/Invalid/Bool"),
		inBuf:   `true`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(nil).withType('t', float64Type),
	}, {
		name:    Name("Floats/Invalid/String"),
		inBuf:   `"0"`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(nil).withType('"', float64Type),
	}, {
		name:    Name("Floats/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(nil).withType('{', float64Type),
	}, {
		name:    Name("Floats/Invalid/Array"),
		inBuf:   `[]`,
		inVal:   addr(float64(64.64)),
		want:    addr(float64(64.64)),
		wantErr: EU(nil).withType('[', float64Type),
	}, {
		name:  Name("Floats/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `1`,
		inVal: addr(float64(0)),
		want:  addr(float64(1)),
	}, {
		name:  Name("Maps/Null"),
		inBuf: `null`,
		inVal: addr(map[string]string{"key": "value"}),
		want:  new(map[string]string),
	}, {
		name:    Name("Maps/InvalidKey/Bool"),
		inBuf:   `{"true":"false"}`,
		inVal:   new(map[bool]bool),
		want:    addr(make(map[bool]bool)),
		wantErr: EU(nil).withPos(`{`, "/true").withType('"', boolType),
	}, {
		name:    Name("Maps/InvalidKey/NamedBool"),
		inBuf:   `{"true":"false"}`,
		inVal:   new(map[namedBool]bool),
		want:    addr(make(map[namedBool]bool)),
		wantErr: EU(nil).withPos(`{`, "/true").withType('"', T[namedBool]()),
	}, {
		name:    Name("Maps/InvalidKey/Array"),
		inBuf:   `{"key":"value"}`,
		inVal:   new(map[[1]string]string),
		want:    addr(make(map[[1]string]string)),
		wantErr: EU(nil).withPos(`{`, "/key").withType('"', T[[1]string]()),
	}, {
		name:    Name("Maps/InvalidKey/Channel"),
		inBuf:   `{"key":"value"}`,
		inVal:   new(map[chan string]string),
		want:    addr(make(map[chan string]string)),
		wantErr: EU(nil).withPos(`{`, "").withType(0, T[chan string]()),
	}, {
		name:  Name("Maps/ValidKey/Int"),
		inBuf: `{"0":0,"-1":1,"2":2,"-3":3}`,
		inVal: new(map[int]int),
		want:  addr(map[int]int{0: 0, -1: 1, 2: 2, -3: 3}),
	}, {
		name:  Name("Maps/ValidKey/NamedInt"),
		inBuf: `{"0":0,"-1":1,"2":2,"-3":3}`,
		inVal: new(map[namedInt64]int),
		want:  addr(map[namedInt64]int{0: 0, -1: 1, 2: 2, -3: 3}),
	}, {
		name:  Name("Maps/ValidKey/Uint"),
		inBuf: `{"0":0,"1":1,"2":2,"3":3}`,
		inVal: new(map[uint]uint),
		want:  addr(map[uint]uint{0: 0, 1: 1, 2: 2, 3: 3}),
	}, {
		name:  Name("Maps/ValidKey/NamedUint"),
		inBuf: `{"0":0,"1":1,"2":2,"3":3}`,
		inVal: new(map[namedUint64]uint),
		want:  addr(map[namedUint64]uint{0: 0, 1: 1, 2: 2, 3: 3}),
	}, {
		name:  Name("Maps/ValidKey/Float"),
		inBuf: `{"1.234":1.234,"12.34":12.34,"123.4":123.4}`,
		inVal: new(map[float64]float64),
		want:  addr(map[float64]float64{1.234: 1.234, 12.34: 12.34, 123.4: 123.4}),
	}, {
		name:    Name("Maps/DuplicateName/Int"),
		inBuf:   `{"0":1,"-0":-1}`,
		inVal:   new(map[int]int),
		want:    addr(map[int]int{0: 1}),
		wantErr: newDuplicateNameError("", []byte(`"-0"`), len64(`{"0":1,`)),
	}, {
		name:    Name("Maps/DuplicateName/Int/MergeWithLegacySemantics"),
		opts:    []Options{unsupported("MergeWithLegacySemantics")},
		inBuf:   `{"0":1,"-0":-1}`,
		inVal:   new(map[int]int),
		want:    addr(map[int]int{0: 1}),
		wantErr: newDuplicateNameError("", []byte(`"-0"`), len64(`{"0":1,`)),
	}, {
		name:  Name("Maps/DuplicateName/Int/AllowDuplicateNames"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"0":1,"-0":-1}`,
		inVal: new(map[int]int),
		want:  addr(map[int]int{0: -1}), // latter takes precedence
	}, {
		name:  Name("Maps/DuplicateName/Int/OverwriteExisting"),
		inBuf: `{"-0":-1}`,
		inVal: addr(map[int]int{0: 1}),
		want:  addr(map[int]int{0: -1}),
	}, {
		name:    Name("Maps/DuplicateName/Float"),
		inBuf:   `{"1.0":"1.0","1":"1","1e0":"1e0"}`,
		inVal:   new(map[float64]string),
		want:    addr(map[float64]string{1: "1.0"}),
		wantErr: newDuplicateNameError("", []byte(`"1"`), len64(`{"1.0":"1.0",`)),
	}, {
		name:  Name("Maps/DuplicateName/Float/AllowDuplicateNames"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"1.0":"1.0","1":"1","1e0":"1e0"}`,
		inVal: new(map[float64]string),
		want:  addr(map[float64]string{1: "1e0"}), // latter takes precedence
	}, {
		name:  Name("Maps/DuplicateName/Float/OverwriteExisting"),
		inBuf: `{"1.0":"1.0"}`,
		inVal: addr(map[float64]string{1: "1"}),
		want:  addr(map[float64]string{1: "1.0"}),
	}, {
		name:    Name("Maps/DuplicateName/NoCaseString"),
		inBuf:   `{"hello":"hello","HELLO":"HELLO"}`,
		inVal:   new(map[nocaseString]string),
		want:    addr(map[nocaseString]string{"hello": "hello"}),
		wantErr: newDuplicateNameError("", []byte(`"HELLO"`), len64(`{"hello":"hello",`)),
	}, {
		name:  Name("Maps/DuplicateName/NoCaseString/AllowDuplicateNames"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"hello":"hello","HELLO":"HELLO"}`,
		inVal: new(map[nocaseString]string),
		want:  addr(map[nocaseString]string{"hello": "HELLO"}), // latter takes precedence
	}, {
		name:  Name("Maps/DuplicateName/NoCaseString/OverwriteExisting"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"HELLO":"HELLO"}`,
		inVal: addr(map[nocaseString]string{"hello": "hello"}),
		want:  addr(map[nocaseString]string{"hello": "HELLO"}),
	}, {
		name:  Name("Maps/ValidKey/Interface"),
		inBuf: `{"false":"false","true":"true","string":"string","0":"0","[]":"[]","{}":"{}"}`,
		inVal: new(map[any]string),
		want: addr(map[any]string{
			"false":  "false",
			"true":   "true",
			"string": "string",
			"0":      "0",
			"[]":     "[]",
			"{}":     "{}",
		}),
	}, {
		name:  Name("Maps/InvalidValue/Channel"),
		inBuf: `{"key":"value"}`,
		inVal: new(map[string]chan string),
		want: addr(map[string]chan string{
			"key": nil,
		}),
		wantErr: EU(nil).withPos(`{"key":`, "/key").withType(0, T[chan string]()),
	}, {
		name:  Name("Maps/RecursiveMap"),
		inBuf: `{"buzz":{},"fizz":{"bar":{},"foo":{}}}`,
		inVal: new(recursiveMap),
		want: addr(recursiveMap{
			"fizz": {
				"foo": {},
				"bar": {},
			},
			"buzz": {},
		}),
	}, {
		// NOTE: The semantics differs from v1,
		// where existing map entries were not merged into.
		// See https://go.dev/issue/31924.
		name:  Name("Maps/Merge"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"k1":{"k2":"v2"},"k2":{"k1":"v1"},"k2":{"k2":"v2"}}`,
		inVal: addr(map[string]map[string]string{
			"k1": {"k1": "v1"},
		}),
		want: addr(map[string]map[string]string{
			"k1": {"k1": "v1", "k2": "v2"},
			"k2": {"k1": "v1", "k2": "v2"},
		}),
	}, {
		name:    Name("Maps/Invalid/Bool"),
		inBuf:   `true`,
		inVal:   addr(map[string]string{"key": "value"}),
		want:    addr(map[string]string{"key": "value"}),
		wantErr: EU(nil).withType('t', T[map[string]string]()),
	}, {
		name:    Name("Maps/Invalid/String"),
		inBuf:   `""`,
		inVal:   addr(map[string]string{"key": "value"}),
		want:    addr(map[string]string{"key": "value"}),
		wantErr: EU(nil).withType('"', T[map[string]string]()),
	}, {
		name:    Name("Maps/Invalid/Number"),
		inBuf:   `0`,
		inVal:   addr(map[string]string{"key": "value"}),
		want:    addr(map[string]string{"key": "value"}),
		wantErr: EU(nil).withType('0', T[map[string]string]()),
	}, {
		name:    Name("Maps/Invalid/Array"),
		inBuf:   `[]`,
		inVal:   addr(map[string]string{"key": "value"}),
		want:    addr(map[string]string{"key": "value"}),
		wantErr: EU(nil).withType('[', T[map[string]string]()),
	}, {
		name:  Name("Maps/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `{"hello":"goodbye"}`,
		inVal: addr(map[string]string{}),
		want:  addr(map[string]string{"hello": "goodbye"}),
	}, {
		name:  Name("Structs/Null"),
		inBuf: `null`,
		inVal: addr(structAll{String: "something"}),
		want:  addr(structAll{}),
	}, {
		name:  Name("Structs/Empty"),
		inBuf: `{}`,
		inVal: addr(structAll{
			String: "hello",
			Map:    map[string]string{},
			Slice:  []string{},
		}),
		want: addr(structAll{
			String: "hello",
			Map:    map[string]string{},
			Slice:  []string{},
		}),
	}, {
		name: Name("Structs/Normal"),
		inBuf: `{
	"Bool": true,
	"String": "hello",
	"Bytes": "AQID",
	"Int": -64,
	"Uint": 64,
	"Float": 3.14159,
	"Map": {"key": "value"},
	"StructScalars": {
		"Bool": true,
		"String": "hello",
		"Bytes": "AQID",
		"Int": -64,
		"Uint": 64,
		"Float": 3.14159
	},
	"StructMaps": {
		"MapBool": {"": true},
		"MapString": {"": "hello"},
		"MapBytes": {"": "AQID"},
		"MapInt": {"": -64},
		"MapUint": {"": 64},
		"MapFloat": {"": 3.14159}
	},
	"StructSlices": {
		"SliceBool": [true],
		"SliceString": ["hello"],
		"SliceBytes": ["AQID"],
		"SliceInt": [-64],
		"SliceUint": [64],
		"SliceFloat": [3.14159]
	},
	"Slice": ["fizz","buzz"],
	"Array": ["goodbye"],
	"Pointer": {},
	"Interface": null
}`,
		inVal: new(structAll),
		want: addr(structAll{
			Bool:   true,
			String: "hello",
			Bytes:  []byte{1, 2, 3},
			Int:    -64,
			Uint:   +64,
			Float:  3.14159,
			Map:    map[string]string{"key": "value"},
			StructScalars: structScalars{
				Bool:   true,
				String: "hello",
				Bytes:  []byte{1, 2, 3},
				Int:    -64,
				Uint:   +64,
				Float:  3.14159,
			},
			StructMaps: structMaps{
				MapBool:   map[string]bool{"": true},
				MapString: map[string]string{"": "hello"},
				MapBytes:  map[string][]byte{"": {1, 2, 3}},
				MapInt:    map[string]int64{"": -64},
				MapUint:   map[string]uint64{"": +64},
				MapFloat:  map[string]float64{"": 3.14159},
			},
			StructSlices: structSlices{
				SliceBool:   []bool{true},
				SliceString: []string{"hello"},
				SliceBytes:  [][]byte{{1, 2, 3}},
				SliceInt:    []int64{-64},
				SliceUint:   []uint64{+64},
				SliceFloat:  []float64{3.14159},
			},
			Slice:   []string{"fizz", "buzz"},
			Array:   [1]string{"goodbye"},
			Pointer: new(structAll),
		}),
	}, {
		name: Name("Structs/Merge"),
		inBuf: `{
	"Bool": false,
	"String": "goodbye",
	"Int": -64,
	"Float": 3.14159,
	"Map": {"k2": "v2"},
	"StructScalars": {
		"Bool": true,
		"String": "hello",
		"Bytes": "AQID",
		"Int": -64
	},
	"StructMaps": {
		"MapBool": {"": true},
		"MapString": {"": "hello"},
		"MapBytes": {"": "AQID"},
		"MapInt": {"": -64},
		"MapUint": {"": 64},
		"MapFloat": {"": 3.14159}
	},
	"StructSlices": {
		"SliceString": ["hello"],
		"SliceBytes": ["AQID"],
		"SliceInt": [-64],
		"SliceUint": [64]
	},
	"Slice": ["fizz","buzz"],
	"Array": ["goodbye"],
	"Pointer": {},
	"Interface": {"k2":"v2"}
}`,
		inVal: addr(structAll{
			Bool:   true,
			String: "hello",
			Bytes:  []byte{1, 2, 3},
			Uint:   +64,
			Float:  math.NaN(),
			Map:    map[string]string{"k1": "v1"},
			StructScalars: structScalars{
				String: "hello",
				Bytes:  make([]byte, 2, 4),
				Uint:   +64,
				Float:  3.14159,
			},
			StructMaps: structMaps{
				MapBool:  map[string]bool{"": false},
				MapBytes: map[string][]byte{"": {}},
				MapInt:   map[string]int64{"": 123},
				MapFloat: map[string]float64{"": math.Inf(+1)},
			},
			StructSlices: structSlices{
				SliceBool:  []bool{true},
				SliceBytes: [][]byte{nil, nil},
				SliceInt:   []int64{-123},
				SliceUint:  []uint64{+123},
				SliceFloat: []float64{3.14159},
			},
			Slice:     []string{"buzz", "fizz", "gizz"},
			Array:     [1]string{"hello"},
			Pointer:   new(structAll),
			Interface: map[string]string{"k1": "v1"},
		}),
		want: addr(structAll{
			Bool:   false,
			String: "goodbye",
			Bytes:  []byte{1, 2, 3},
			Int:    -64,
			Uint:   +64,
			Float:  3.14159,
			Map:    map[string]string{"k1": "v1", "k2": "v2"},
			StructScalars: structScalars{
				Bool:   true,
				String: "hello",
				Bytes:  []byte{1, 2, 3},
				Int:    -64,
				Uint:   +64,
				Float:  3.14159,
			},
			StructMaps: structMaps{
				MapBool:   map[string]bool{"": true},
				MapString: map[string]string{"": "hello"},
				MapBytes:  map[string][]byte{"": {1, 2, 3}},
				MapInt:    map[string]int64{"": -64},
				MapUint:   map[string]uint64{"": +64},
				MapFloat:  map[string]float64{"": 3.14159},
			},
			StructSlices: structSlices{
				SliceBool:   []bool{true},
				SliceString: []string{"hello"},
				SliceBytes:  [][]byte{{1, 2, 3}},
				SliceInt:    []int64{-64},
				SliceUint:   []uint64{+64},
				SliceFloat:  []float64{3.14159},
			},
			Slice:     []string{"fizz", "buzz"},
			Array:     [1]string{"goodbye"},
			Pointer:   new(structAll),
			Interface: map[string]string{"k1": "v1", "k2": "v2"},
		}),
	}, {
		name: Name("Structs/Stringified"),
		inBuf: `{
	"Int": "-64",
	"Uint": "64",
	"Float": "3.14159",
	"PointerInt": "-64",
	"PointerUint": "64",
	"PointerFloat": "3.14159"
}`,
		inVal: new(structStringified),
		want: addr(structStringified{
			Int:          -64,
			Uint:         +64,
			Float:        3.14159,
			PointerInt:   ptr(int64(-64)),
			PointerUint:  ptr(uint64(+64)),
			PointerFloat: ptr(float64(3.14159)),
		}),
	}, {
		name:    Name("Structs/Stringified/Invalid/EmptyIntString"),
		inBuf:   `{"Int":""}`,
		inVal:   new(structStringified),
		want:    new(structStringified),
		wantErr: EU(strconv.ErrSyntax).withVal(`""`).withPos(`{"Int":`, "/Int").withType('"', T[int64]()),
	}, {
		name:    Name("Structs/Stringified/Invalid/Bool"),
		inBuf:   `{"Bool": "true"}`,
		inVal:   new(structStringifiedBool),
		want:    new(structStringifiedBool),
		wantErr: EU(errInvalidStringTag).withPos(`{"Bool": `, "/Bool").withType(0, T[bool]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Bool"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Bool": true}`,
		inVal: new(structStringifiedBool),
		want:  ptr(structStringifiedBool{Bool: true}),
	}, {
		name:    Name("Structs/Stringified/Invalid/String"),
		inBuf:   `{"String": "\"hello\""}`,
		inVal:   new(structStringifiedString),
		want:    new(structStringifiedString),
		wantErr: EU(errInvalidStringTag).withPos(`{"String": `, "/String").withType(0, T[string]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/String"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"String": "\"hello\""}`,
		inVal: new(structStringifiedString),
		want:  ptr(structStringifiedString{String: `"hello"`}),
	}, {
		name:    Name("Structs/Stringified/Invalid/Bytes"),
		inBuf:   `{"Bytes": "AQID"}`,
		inVal:   new(structStringifiedBytes),
		want:    new(structStringifiedBytes),
		wantErr: EU(errInvalidStringTag).withPos(`{"Bytes": `, "/Bytes").withType(0, T[[]byte]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Bytes"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Bytes": "AQID"}`,
		inVal: new(structStringifiedBytes),
		want:  ptr(structStringifiedBytes{Bytes: []byte{1, 2, 3}}),
	}, {
		name:    Name("Structs/Stringified/Invalid/Map"),
		inBuf:   `{"Map": {"Key": "Value"}}`,
		inVal:   new(structStringifiedMap),
		want:    new(structStringifiedMap),
		wantErr: EU(errInvalidStringTag).withPos(`{"Map": `, "/Map").withType(0, T[map[string]string]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Map"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Map": {"Key": "Value"}}`,
		inVal: new(structStringifiedMap),
		want:  ptr(structStringifiedMap{Map: map[string]string{"Key": "Value"}}),
	}, {
		name:    Name("Structs/Stringified/Invalid/Slice"),
		inBuf:   `{"Slice": ["hello"]}`,
		inVal:   new(structStringifiedSlice),
		want:    new(structStringifiedSlice),
		wantErr: EU(errInvalidStringTag).withPos(`{"Slice": `, "/Slice").withType(0, T[[]string]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Slice"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Slice": ["hello"]}`,
		inVal: new(structStringifiedSlice),
		want:  ptr(structStringifiedSlice{Slice: []string{"hello"}}),
	}, {
		name:    Name("Structs/Stringified/Invalid/Array"),
		inBuf:   `{"Array": ["hello"]}`,
		inVal:   new(structStringifiedArray),
		want:    new(structStringifiedArray),
		wantErr: EU(errInvalidStringTag).withPos(`{"Array": `, "/Array").withType(0, T[[1]string]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Array"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Array": ["hello"]}`,
		inVal: new(structStringifiedArray),
		want:  ptr(structStringifiedArray{Array: [1]string{"hello"}}),
	}, {
		name:    Name("Structs/Stringified/Invalid/Struct"),
		inBuf:   `{"Struct": {"Bool": true}}`,
		inVal:   new(structStringifiedStruct),
		want:    new(structStringifiedStruct),
		wantErr: EU(errInvalidStringTag).withPos(`{"Struct": `, "/Struct").withType(0, T[structAll]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Struct"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Struct": {"Bool": true}}`,
		inVal: new(structStringifiedStruct),
		want:  ptr(structStringifiedStruct{Struct: structAll{Bool: true}}),
	}, {
		name:    Name("Structs/Stringified/Invalid/Pointer"),
		inBuf:   `{"Pointer": {"Bool": true}}`,
		inVal:   new(structStringifiedPointer),
		want:    ptr(structStringifiedPointer{Pointer: new(structAll)}),
		wantErr: EU(errInvalidStringTag).withPos(`{"Pointer": `, "/Pointer").withType(0, T[structAll]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Pointer"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Pointer": {"Bool": true}}`,
		inVal: new(structStringifiedPointer),
		want:  ptr(structStringifiedPointer{Pointer: ptr(structAll{Bool: true})}),
	}, {
		name:  Name("Structs/Stringified/PointerPointerInt"),
		inBuf: `{"Pointer": "5"}`,
		inVal: new(structStringifiedPointerPointerInt),
		want:  ptr(structStringifiedPointerPointerInt{Pointer: ptr(ptr(5))}),
	}, {
		name:    Name("Structs/Stringified/Invalid/Interface"),
		inBuf:   `{"Interface": null}`,
		inVal:   new(structStringifiedInterface),
		want:    new(structStringifiedInterface),
		wantErr: EU(errInvalidStringTag).withPos(`{"Interface": `, "/Interface").withType(0, T[any]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/Interface"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Interface": null}`,
		inVal: new(structStringifiedInterface),
		want:  new(structStringifiedInterface),
	}, {
		name: Name("Structs/LegacyStringified"),
		opts: []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf: `{
	"Bool": "true",
	"String": "\"hello\"",
	"Int": "-64",
	"Uint": "64",
	"Float": "3.14159",
	"PointerBool": "true",
	"PointerString": "\"hello\"",
	"PointerInt": "-64",
	"PointerUint": "64",
	"PointerFloat": "3.14159"
}`,
		inVal: new(structStringifiedLegacy),
		want: addr(structStringifiedLegacy{
			Bool:          true,
			String:        "hello",
			Int:           -64,
			Uint:          +64,
			Float:         3.14159,
			PointerBool:   ptr(true),
			PointerString: ptr("hello"),
			PointerInt:    ptr(int64(-64)),
			PointerUint:   ptr(uint64(+64)),
			PointerFloat:  ptr(float64(3.14159)),
		}),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/EmptyIntString"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Int":""}`,
		inVal:   new(structStringifiedLegacy),
		want:    new(structStringifiedLegacy),
		wantErr: EU(strconv.ErrSyntax).withVal(`""`).withPos(`{"Int":`, "/Int").withType('"', T[int64]()),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Bool"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Bool": true}`,
		inVal:   new(structStringifiedLegacy),
		wantErr: EU(nil).withPos(`{"Bool": `, "/Bool").withType('t', T[bool]()),
	}, {
		name:  Name("Structs/LegacyStringified/Invalid/String"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf: `{"String": "string"}`,
		inVal: new(structStringifiedLegacy),
		wantErr: EU(newInvalidCharacterError("s", "at start of string (expecting '\"')", 0, "")).
			withPos(`{"String": `, "/String").withType('"', T[string]()),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Bytes"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Bytes": "AQID"}`,
		inVal:   new(structStringifiedBytes),
		want:    new(structStringifiedBytes),
		wantErr: EU(errInvalidStringTag).withPos(`{"Bytes": `, "/Bytes").withType(0, T[[]byte]()),
	}, {
		name:  Name("Structs/LegacyStringified/Ignored/Bytes"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Bytes": "AQID"}`,
		inVal: new(structStringifiedBytes),
		want:  ptr(structStringifiedBytes{Bytes: []byte{1, 2, 3}}),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Map"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Map": {"Key": "Value"}}`,
		inVal:   new(structStringifiedMap),
		want:    new(structStringifiedMap),
		wantErr: EU(errInvalidStringTag).withPos(`{"Map": `, "/Map").withType(0, T[map[string]string]()),
	}, {
		name:  Name("Structs/LegacyStringified/Ignored/Map"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Map": {"Key": "Value"}}`,
		inVal: new(structStringifiedMap),
		want:  ptr(structStringifiedMap{Map: map[string]string{"Key": "Value"}}),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Slice"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Slice": ["hello"]}`,
		inVal:   new(structStringifiedSlice),
		want:    new(structStringifiedSlice),
		wantErr: EU(errInvalidStringTag).withPos(`{"Slice": `, "/Slice").withType(0, T[[]string]()),
	}, {
		name:  Name("Structs/LegacyStringified/Ignored/Slice"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Slice": ["hello"]}`,
		inVal: new(structStringifiedSlice),
		want:  ptr(structStringifiedSlice{Slice: []string{"hello"}}),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Array"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Array": ["hello"]}`,
		inVal:   new(structStringifiedArray),
		want:    new(structStringifiedArray),
		wantErr: EU(errInvalidStringTag).withPos(`{"Array": `, "/Array").withType(0, T[[1]string]()),
	}, {
		name:  Name("Structs/LegacyStringified/Ignored/Array"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Array": ["hello"]}`,
		inVal: new(structStringifiedArray),
		want:  ptr(structStringifiedArray{Array: [1]string{"hello"}}),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Struct"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Struct": {"Bool": true}}`,
		inVal:   new(structStringifiedStruct),
		want:    new(structStringifiedStruct),
		wantErr: EU(errInvalidStringTag).withPos(`{"Struct": `, "/Struct").withType(0, T[structAll]()),
	}, {
		name:  Name("Structs/LegacyStringified/Ignored/Struct"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Struct": {"Bool": true}}`,
		inVal: new(structStringifiedStruct),
		want:  ptr(structStringifiedStruct{Struct: structAll{Bool: true}}),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Pointer"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Pointer": {"Bool": true}}`,
		inVal:   new(structStringifiedPointer),
		want:    ptr(structStringifiedPointer{Pointer: new(structAll)}),
		wantErr: EU(errInvalidStringTag).withPos(`{"Pointer": `, "/Pointer").withType(0, T[structAll]()),
	}, {
		name:  Name("Structs/LegacyStringified/Ignored/Pointer"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Pointer": {"Bool": true}}`,
		inVal: new(structStringifiedPointer),
		want:  ptr(structStringifiedPointer{Pointer: ptr(structAll{Bool: true})}),
	}, {
		name:    Name("Structs/Stringified/Invalid/PointerPointerInt"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Pointer": 5}`,
		inVal:   new(structStringifiedPointerPointerInt),
		want:    ptr(structStringifiedPointerPointerInt{}),
		wantErr: EU(errInvalidStringTag).withPos(`{"Pointer": `, "/Pointer").withType(0, T[**int]()),
	}, {
		name:  Name("Structs/Stringified/Ignored/PointerPointerInt"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics|ReportErrorsWithLegacySemantics")},
		inBuf: `{"Pointer": 5}`,
		inVal: new(structStringifiedPointerPointerInt),
		want:  ptr(structStringifiedPointerPointerInt{Pointer: ptr(ptr(5))}),
	}, {
		name:    Name("Structs/LegacyStringified/Invalid/Interface"),
		opts:    []Options{unsupported("StringifyWithLegacySemantics")},
		inBuf:   `{"Interface": null}`,
		inVal:   new(structStringifiedInterface),
		want:    new(structStringifiedInterface),
		wantErr: EU(errInvalidStringTag).withPos(`{"Interface": `, "/Interface").withType(0, T[any]()),
	}, {
		name:  Name("Structs/LegacyStringified/Ignored/Interface"),
		opts:  []Options{unsupported("StringifyWithLegacySemantics"), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"Interface": null}`,
		inVal: new(structStringifiedInterface),
		want:  new(structStringifiedInterface),
	}, {
		name:    Name("Structs/Format/Bytes/Unsupported"),
		inBuf:   `{}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(errors.New("Go struct field Base16 has unsupported `format` tag option")).withType('{', T[structFormatBytes]()),
	}, {
		name: Name("Structs/Format/Bytes"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{
	"Base16": "0123456789abcdef",
	"Base32": "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567",
	"Base32Hex": "0123456789ABCDEFGHIJKLMNOPQRSTUV",
	"Base64": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
	"Base64URL": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_",
	"Array": [1, 2, 3, 4]
}`,
		inVal: new(structFormatBytes),
		want: addr(structFormatBytes{
			Base16:    []byte("\x01\x23\x45\x67\x89\xab\xcd\xef"),
			Base32:    []byte("\x00D2\x14\xc7BT\xb65τe:V\xd7\xc6u\xbew\xdf"),
			Base32Hex: []byte("\x00D2\x14\xc7BT\xb65τe:V\xd7\xc6u\xbew\xdf"),
			Base64:    []byte("\x00\x10\x83\x10Q\x87 \x92\x8b0ӏA\x14\x93QU\x97a\x96\x9bqן\x82\x18\xa3\x92Y\xa7\xa2\x9a\xab\xb2ۯ\xc3\x1c\xb3\xd3]\xb7㞻\xf3߿"),
			Base64URL: []byte("\x00\x10\x83\x10Q\x87 \x92\x8b0ӏA\x14\x93QU\x97a\x96\x9bqן\x82\x18\xa3\x92Y\xa7\xa2\x9a\xab\xb2ۯ\xc3\x1c\xb3\xd3]\xb7㞻\xf3߿"),
			Array:     []byte{1, 2, 3, 4},
		}),
	}, {
		name: Name("Structs/Format/ArrayBytes"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{
	"Base16": "01020304",
	"Base32": "AEBAGBA=",
	"Base32Hex": "0410610=",
	"Base64": "AQIDBA==",
	"Base64URL": "AQIDBA==",
	"Array": [1, 2, 3, 4],
	"Default": "AQIDBA=="
}`,
		inVal: new(structFormatArrayBytes),
		want: addr(structFormatArrayBytes{
			Base16:    [4]byte{1, 2, 3, 4},
			Base32:    [4]byte{1, 2, 3, 4},
			Base32Hex: [4]byte{1, 2, 3, 4},
			Base64:    [4]byte{1, 2, 3, 4},
			Base64URL: [4]byte{1, 2, 3, 4},
			Array:     [4]byte{1, 2, 3, 4},
			Default:   [4]byte{1, 2, 3, 4},
		}),
	}, {
		name: Name("Structs/Format/ArrayBytes/Legacy"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag"), unsupported("FormatBytesWithLegacySemantics")},
		inBuf: `{
	"Base16": "01020304",
	"Base32": "AEBAGBA=",
	"Base32Hex": "0410610=",
	"Base64": "AQIDBA==",
	"Base64URL": "AQIDBA==",
	"Array": [1, 2, 3, 4],
	"Default": [1, 2, 3, 4]
}`,
		inVal: new(structFormatArrayBytes),
		want: addr(structFormatArrayBytes{
			Base16:    [4]byte{1, 2, 3, 4},
			Base32:    [4]byte{1, 2, 3, 4},
			Base32Hex: [4]byte{1, 2, 3, 4},
			Base64:    [4]byte{1, 2, 3, 4},
			Base64URL: [4]byte{1, 2, 3, 4},
			Array:     [4]byte{1, 2, 3, 4},
			Default:   [4]byte{1, 2, 3, 4},
		}),
	}, {
		name:    Name("Structs/Format/Bytes/Invalid/Base16/WrongKind"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Base16": [1,2,3,4]}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(nil).withPos(`{"Base16": `, "/Base16").withType('[', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base16/AllPadding"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base16": "===="}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := hex.Decode(make([]byte, 2), []byte("====="))
			return err
		}()).withPos(`{"Base16": `, "/Base16").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base16/EvenPadding"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base16": "0123456789abcdef="}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := hex.Decode(make([]byte, 8), []byte("0123456789abcdef="))
			return err
		}()).withPos(`{"Base16": `, "/Base16").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base16/OddPadding"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base16": "0123456789abcdef0="}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := hex.Decode(make([]byte, 9), []byte("0123456789abcdef0="))
			return err
		}()).withPos(`{"Base16": `, "/Base16").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base16/NonAlphabet/LineFeed"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base16": "aa\naa"}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := hex.Decode(make([]byte, 9), []byte("aa\naa"))
			return err
		}()).withPos(`{"Base16": `, "/Base16").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base16/NonAlphabet/CarriageReturn"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base16": "aa\raa"}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := hex.Decode(make([]byte, 9), []byte("aa\raa"))
			return err
		}()).withPos(`{"Base16": `, "/Base16").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base16/NonAlphabet/Space"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base16": "aa aa"}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := hex.Decode(make([]byte, 9), []byte("aa aa"))
			return err
		}()).withPos(`{"Base16": `, "/Base16").withType('"', T[[]byte]()),
	}, {
		name: Name("Structs/Format/Bytes/Invalid/Base32/Padding"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `[
			{"Base32": "NA======"},
			{"Base32": "NBSQ===="},
			{"Base32": "NBSWY==="},
			{"Base32": "NBSWY3A="},
			{"Base32": "NBSWY3DP"}
		]`,
		inVal: new([]structFormatBytes),
		want: addr([]structFormatBytes{
			{Base32: []byte("h")},
			{Base32: []byte("he")},
			{Base32: []byte("hel")},
			{Base32: []byte("hell")},
			{Base32: []byte("hello")},
		}),
	}, {
		name: Name("Structs/Format/Bytes/Invalid/Base32/Invalid/NoPadding"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `[
				{"Base32": "NA"},
				{"Base32": "NBSQ"},
				{"Base32": "NBSWY"},
				{"Base32": "NBSWY3A"},
				{"Base32": "NBSWY3DP"}
			]`,
		inVal: new([]structFormatBytes),
		wantErr: EU(func() error {
			_, err := base32.StdEncoding.Decode(make([]byte, 1), []byte("NA"))
			return err
		}()).withPos(`[`+"\n\t\t\t\t"+`{"Base32": `, "/0/Base32").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base32/WrongAlphabet"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base32": "0123456789ABCDEFGHIJKLMNOPQRSTUV"}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := base32.StdEncoding.Decode(make([]byte, 20), []byte("0123456789ABCDEFGHIJKLMNOPQRSTUV"))
			return err
		}()).withPos(`{"Base32": `, "/Base32").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base32Hex/WrongAlphabet"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base32Hex": "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := base32.HexEncoding.Decode(make([]byte, 20), []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"))
			return err
		}()).withPos(`{"Base32Hex": `, "/Base32Hex").withType('"', T[[]byte]()),
	}, {
		name:    Name("Structs/Format/Bytes/Invalid/Base32/NonAlphabet/LineFeed"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Base32": "AAAA\nAAAA"}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(errors.New("illegal character '\\n' at offset 4")).withPos(`{"Base32": `, "/Base32").withType('"', T[[]byte]()),
	}, {
		name:    Name("Structs/Format/Bytes/Invalid/Base32/NonAlphabet/CarriageReturn"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Base32": "AAAA\rAAAA"}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(errors.New("illegal character '\\r' at offset 4")).withPos(`{"Base32": `, "/Base32").withType('"', T[[]byte]()),
	}, {
		name:    Name("Structs/Format/Bytes/Invalid/Base32/NonAlphabet/Space"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Base32": "AAAA AAAA"}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(base32.CorruptInputError(4)).withPos(`{"Base32": `, "/Base32").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base64/WrongAlphabet"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base64": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := base64.StdEncoding.Decode(make([]byte, 48), []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"))
			return err
		}()).withPos(`{"Base64": `, "/Base64").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Invalid/Base64URL/WrongAlphabet"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"Base64URL": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"}`,
		inVal: new(structFormatBytes),
		wantErr: EU(func() error {
			_, err := base64.URLEncoding.Decode(make([]byte, 48), []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"))
			return err
		}()).withPos(`{"Base64URL": `, "/Base64URL").withType('"', T[[]byte]()),
	}, {
		name:    Name("Structs/Format/Bytes/Invalid/Base64/NonAlphabet/LineFeed"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Base64": "aa=\n="}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(errors.New("illegal character '\\n' at offset 3")).withPos(`{"Base64": `, "/Base64").withType('"', T[[]byte]()),
	}, {
		name:    Name("Structs/Format/Bytes/Invalid/Base64/NonAlphabet/CarriageReturn"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Base64": "aa=\r="}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(errors.New("illegal character '\\r' at offset 3")).withPos(`{"Base64": `, "/Base64").withType('"', T[[]byte]()),
	}, {
		name:  Name("Structs/Format/Bytes/Base64/NonAlphabet/Ignored"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag"), unsupported("ParseBytesWithLooseRFC4648")},
		inBuf: `{"Base64": "aa=\r\n="}`,
		inVal: new(structFormatBytes),
		want:  &structFormatBytes{Base64: []byte{105}},
	}, {
		name:    Name("Structs/Format/Bytes/Invalid/Base64/NonAlphabet/Space"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Base64": "aa= ="}`,
		inVal:   new(structFormatBytes),
		wantErr: EU(base64.CorruptInputError(2)).withPos(`{"Base64": `, "/Base64").withType('"', T[[]byte]()),
	}, {
		name: Name("Structs/Format/Floats"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `[
	{"NonFinite": 3.141592653589793, "PointerNonFinite": 3.141592653589793},
	{"NonFinite": "-Infinity", "PointerNonFinite": "-Infinity"},
	{"NonFinite": "Infinity", "PointerNonFinite": "Infinity"}
]`,
		inVal: new([]structFormatFloats),
		want: addr([]structFormatFloats{
			{NonFinite: math.Pi, PointerNonFinite: addr(math.Pi)},
			{NonFinite: math.Inf(-1), PointerNonFinite: addr(math.Inf(-1))},
			{NonFinite: math.Inf(+1), PointerNonFinite: addr(math.Inf(+1))},
		}),
	}, {
		name:  Name("Structs/Format/Floats/NaN"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"NonFinite": "NaN"}`,
		inVal: new(structFormatFloats),
		// Avoid checking want since reflect.DeepEqual fails for NaNs.
	}, {
		name:    Name("Structs/Format/Floats/Invalid/NaN"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"NonFinite": "nan"}`,
		inVal:   new(structFormatFloats),
		wantErr: EU(nil).withPos(`{"NonFinite": `, "/NonFinite").withType('"', T[float64]()),
	}, {
		name:    Name("Structs/Format/Floats/Invalid/PositiveInfinity"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"NonFinite": "+Infinity"}`,
		inVal:   new(structFormatFloats),
		wantErr: EU(nil).withPos(`{"NonFinite": `, "/NonFinite").withType('"', T[float64]()),
	}, {
		name:    Name("Structs/Format/Floats/Invalid/NegativeInfinitySpace"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"NonFinite": "-Infinity "}`,
		inVal:   new(structFormatFloats),
		wantErr: EU(nil).withPos(`{"NonFinite": `, "/NonFinite").withType('"', T[float64]()),
	}, {
		name: Name("Structs/Format/Maps"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `[
	{"EmitNull": null, "PointerEmitNull": null, "EmitEmpty": null, "PointerEmitEmpty": null, "EmitDefault": null, "PointerEmitDefault": null},
	{"EmitNull": {}, "PointerEmitNull": {}, "EmitEmpty": {}, "PointerEmitEmpty": {}, "EmitDefault": {}, "PointerEmitDefault": {}},
	{"EmitNull": {"k": "v"}, "PointerEmitNull": {"k": "v"}, "EmitEmpty": {"k": "v"}, "PointerEmitEmpty": {"k": "v"}, "EmitDefault": {"k": "v"}, "PointerEmitDefault": {"k": "v"}}
]`,
		inVal: new([]structFormatMaps),
		want: addr([]structFormatMaps{{
			EmitNull: map[string]string(nil), PointerEmitNull: (*map[string]string)(nil),
			EmitEmpty: map[string]string(nil), PointerEmitEmpty: (*map[string]string)(nil),
			EmitDefault: map[string]string(nil), PointerEmitDefault: (*map[string]string)(nil),
		}, {
			EmitNull: map[string]string{}, PointerEmitNull: addr(map[string]string{}),
			EmitEmpty: map[string]string{}, PointerEmitEmpty: addr(map[string]string{}),
			EmitDefault: map[string]string{}, PointerEmitDefault: addr(map[string]string{}),
		}, {
			EmitNull: map[string]string{"k": "v"}, PointerEmitNull: addr(map[string]string{"k": "v"}),
			EmitEmpty: map[string]string{"k": "v"}, PointerEmitEmpty: addr(map[string]string{"k": "v"}),
			EmitDefault: map[string]string{"k": "v"}, PointerEmitDefault: addr(map[string]string{"k": "v"}),
		}}),
	}, {
		name: Name("Structs/Format/Slices"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `[
	{"EmitNull": null, "PointerEmitNull": null, "EmitEmpty": null, "PointerEmitEmpty": null, "EmitDefault": null, "PointerEmitDefault": null},
	{"EmitNull": [], "PointerEmitNull": [], "EmitEmpty": [], "PointerEmitEmpty": [], "EmitDefault": [], "PointerEmitDefault": []},
	{"EmitNull": ["v"], "PointerEmitNull": ["v"], "EmitEmpty": ["v"], "PointerEmitEmpty": ["v"], "EmitDefault": ["v"], "PointerEmitDefault": ["v"]}
]`,
		inVal: new([]structFormatSlices),
		want: addr([]structFormatSlices{{
			EmitNull: []string(nil), PointerEmitNull: (*[]string)(nil),
			EmitEmpty: []string(nil), PointerEmitEmpty: (*[]string)(nil),
			EmitDefault: []string(nil), PointerEmitDefault: (*[]string)(nil),
		}, {
			EmitNull: []string{}, PointerEmitNull: addr([]string{}),
			EmitEmpty: []string{}, PointerEmitEmpty: addr([]string{}),
			EmitDefault: []string{}, PointerEmitDefault: addr([]string{}),
		}, {
			EmitNull: []string{"v"}, PointerEmitNull: addr([]string{"v"}),
			EmitEmpty: []string{"v"}, PointerEmitEmpty: addr([]string{"v"}),
			EmitDefault: []string{"v"}, PointerEmitDefault: addr([]string{"v"}),
		}}),
	}, {
		name:    Name("Structs/Format/Invalid/Bool"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Bool":true}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Bool":`, "/Bool").withType(0, T[bool]()),
	}, {
		name:    Name("Structs/Format/Invalid/String"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"String": "string"}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"String": `, "/String").withType(0, T[string]()),
	}, {
		name:    Name("Structs/Format/Invalid/Bytes"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Bytes": "bytes"}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Bytes": `, "/Bytes").withType(0, T[[]byte]()),
	}, {
		name:    Name("Structs/Format/Invalid/Int"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Int":   1}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Int":   `, "/Int").withType(0, T[int64]()),
	}, {
		name:    Name("Structs/Format/Invalid/Uint"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Uint": 1}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Uint": `, "/Uint").withType(0, T[uint64]()),
	}, {
		name:    Name("Structs/Format/Invalid/Float"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Float" : 1}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Float" : `, "/Float").withType(0, T[float64]()),
	}, {
		name:    Name("Structs/Format/Invalid/Map"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Map":{}}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Map":`, "/Map").withType(0, T[map[string]string]()),
	}, {
		name:    Name("Structs/Format/Invalid/Struct"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Struct": {}}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Struct": `, "/Struct").withType(0, T[structAll]()),
	}, {
		name:    Name("Structs/Format/Invalid/Slice"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Slice": {}}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Slice": `, "/Slice").withType(0, T[[]string]()),
	}, {
		name:    Name("Structs/Format/Invalid/Array"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Array": []}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Array": `, "/Array").withType(0, T[[1]string]()),
	}, {
		name:    Name("Structs/Format/Invalid/Interface"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"Interface": "anything"}`,
		inVal:   new(structFormatInvalid),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"Interface": `, "/Interface").withType(0, T[any]()),
	}, {
		name:  Name("Structs/Embed/Zero"),
		inBuf: `{"D":""}`,
		inVal: new(structEmbedded),
		want:  new(structEmbedded),
	}, {
		name:  Name("Structs/Embed/Alloc"),
		inBuf: `{"E":"","F":"","G":"","A":"","B":"","D":""}`,
		inVal: new(structEmbedded),
		want: addr(structEmbedded{
			X: structEmbeddedL1{
				X:            &structEmbeddedL2{},
				StructEmbed1: StructEmbed1{},
			},
			StructEmbed2: &StructEmbed2{},
		}),
	}, {
		name:  Name("Structs/Embed/NonZero"),
		inBuf: `{"E":"E3","F":"F3","G":"G3","A":"A1","B":"B1","D":"D2"}`,
		inVal: new(structEmbedded),
		want: addr(structEmbedded{
			X: structEmbeddedL1{
				X:            &structEmbeddedL2{A: "A1", B: "B1" /* C: "C1" */},
				StructEmbed1: StructEmbed1{ /* C: "C2" */ D: "D2" /* E: "E2" */},
			},
			StructEmbed2: &StructEmbed2{E: "E3", F: "F3", G: "G3"},
		}),
	}, {
		name:  Name("Structs/Embed/Merge"),
		inBuf: `{"E":"E3","F":"F3","G":"G3","A":"A1","B":"B1","D":"D2"}`,
		inVal: addr(structEmbedded{
			X: structEmbeddedL1{
				X:            &structEmbeddedL2{B: "##", C: "C1"},
				StructEmbed1: StructEmbed1{C: "C2", E: "E2"},
			},
			StructEmbed2: &StructEmbed2{E: "##", G: "G3"},
		}),
		want: addr(structEmbedded{
			X: structEmbeddedL1{
				X:            &structEmbeddedL2{A: "A1", B: "B1", C: "C1"},
				StructEmbed1: StructEmbed1{C: "C2", D: "D2", E: "E2"},
			},
			StructEmbed2: &StructEmbed2{E: "E3", F: "F3", G: "G3"},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/Noop"),
		inBuf: `{"A":1,"B":2}`,
		inVal: new(structEmbedTextValue),
		want:  addr(structEmbedTextValue{A: 1, X: jsontext.Value(nil), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/MergeN1/Nil"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: new(structEmbedTextValue),
		want:  addr(structEmbedTextValue{A: 1, X: jsontext.Value(`{"fizz":"buzz"}`), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/MergeN1/Empty"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: addr(structEmbedTextValue{X: jsontext.Value{}}),
		want:  addr(structEmbedTextValue{A: 1, X: jsontext.Value(`{"fizz":"buzz"}`), B: 2}),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/MergeN1/Whitespace"),
		inBuf:   `{"A":1,"fizz":"buzz","B":2}`,
		inVal:   addr(structEmbedTextValue{X: jsontext.Value("\n\r\t ")}),
		want:    addr(structEmbedTextValue{A: 1, X: jsontext.Value("")}),
		wantErr: EU(errRawEmbedNotObject).withPos(`{"A":1,`, "/fizz").withType('"', T[jsontext.Value]()),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/MergeN1/Null"),
		inBuf:   `{"A":1,"fizz":"buzz","B":2}`,
		inVal:   addr(structEmbedTextValue{X: jsontext.Value("null")}),
		want:    addr(structEmbedTextValue{A: 1, X: jsontext.Value("null")}),
		wantErr: EU(errRawEmbedNotObject).withPos(`{"A":1,`, "/fizz").withType('"', T[jsontext.Value]()),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/MergeN1/ObjectN0"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: addr(structEmbedTextValue{X: jsontext.Value(` { } `)}),
		want:  addr(structEmbedTextValue{A: 1, X: jsontext.Value(` {"fizz":"buzz"}`), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/MergeN2/ObjectN1"),
		inBuf: `{"A":1,"fizz":"buzz","B":2,"foo": [ 1 , 2 , 3 ]}`,
		inVal: addr(structEmbedTextValue{X: jsontext.Value(` { "fizz" : "buzz" } `)}),
		want:  addr(structEmbedTextValue{A: 1, X: jsontext.Value(` { "fizz" : "buzz","fizz":"buzz","foo":[ 1 , 2 , 3 ]}`), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/Merge/EndObject"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: addr(structEmbedTextValue{X: jsontext.Value(` } `)}),
		// NOTE: This produces invalid output,
		// but the value being merged into is already invalid.
		want: addr(structEmbedTextValue{A: 1, X: jsontext.Value(`,"fizz":"buzz"}`), B: 2}),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/MergeInvalidValue"),
		inBuf:   `{"A":1,"fizz":nil,"B":2}`,
		inVal:   new(structEmbedTextValue),
		want:    addr(structEmbedTextValue{A: 1, X: jsontext.Value(`{"fizz":`)}),
		wantErr: newInvalidCharacterError("i", "in literal null (expecting 'u')", len64(`{"A":1,"fizz":n`), "/fizz"),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/CaseSensitive"),
		inBuf: `{"A":1,"fizz":"buzz","B":2,"a":3}`,
		inVal: new(structEmbedTextValue),
		want:  addr(structEmbedTextValue{A: 1, X: jsontext.Value(`{"fizz":"buzz","a":3}`), B: 2}),
	}, {
		name:    Name("Structs/EmbeddedFallback/TextValue/RejectDuplicateNames"),
		opts:    []Options{jsontext.AllowDuplicateNames(false)},
		inBuf:   `{"A":1,"fizz":"buzz","B":2,"fizz":"buzz"}`,
		inVal:   new(structEmbedTextValue),
		want:    addr(structEmbedTextValue{A: 1, X: jsontext.Value(`{"fizz":"buzz"}`), B: 2}),
		wantErr: newDuplicateNameError("", []byte(`"fizz"`), len64(`{"A":1,"fizz":"buzz","B":2,`)),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/AllowDuplicateNames"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"A":1,"fizz":"buzz","B":2,"fizz":"buzz"}`,
		inVal: new(structEmbedTextValue),
		want:  addr(structEmbedTextValue{A: 1, X: jsontext.Value(`{"fizz":"buzz","fizz":"buzz"}`), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/Nested/Noop"),
		inBuf: `{}`,
		inVal: new(structEmbedPointerEmbedTextValue),
		want:  new(structEmbedPointerEmbedTextValue),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/Nested/Alloc"),
		inBuf: `{"A":1,"fizz":"buzz"}`,
		inVal: new(structEmbedPointerEmbedTextValue),
		want: addr(structEmbedPointerEmbedTextValue{
			X: &struct {
				A int
				X jsontext.Value `json:",embed"`
			}{A: 1, X: jsontext.Value(`{"fizz":"buzz"}`)},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/TextValue/Nested/Merge"),
		inBuf: `{"fizz":"buzz"}`,
		inVal: addr(structEmbedPointerEmbedTextValue{
			X: &struct {
				A int
				X jsontext.Value `json:",embed"`
			}{A: 1},
		}),
		want: addr(structEmbedPointerEmbedTextValue{
			X: &struct {
				A int
				X jsontext.Value `json:",embed"`
			}{A: 1, X: jsontext.Value(`{"fizz":"buzz"}`)},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerTextValue/Noop"),
		inBuf: `{"A":1,"B":2}`,
		inVal: new(structEmbedPointerTextValue),
		want:  addr(structEmbedPointerTextValue{A: 1, X: nil, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerTextValue/Alloc"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: new(structEmbedPointerTextValue),
		want:  addr(structEmbedPointerTextValue{A: 1, X: addr(jsontext.Value(`{"fizz":"buzz"}`)), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerTextValue/Merge"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: addr(structEmbedPointerTextValue{X: addr(jsontext.Value(`{"fizz":"buzz"}`))}),
		want:  addr(structEmbedPointerTextValue{A: 1, X: addr(jsontext.Value(`{"fizz":"buzz","fizz":"buzz"}`)), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerTextValue/Nested/Nil"),
		inBuf: `{"fizz":"buzz"}`,
		inVal: new(structEmbedEmbedPointerTextValue),
		want: addr(structEmbedEmbedPointerTextValue{
			X: struct {
				X *jsontext.Value `json:",embed"`
			}{X: addr(jsontext.Value(`{"fizz":"buzz"}`))},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/Noop"),
		inBuf: `{"A":1,"B":2}`,
		inVal: new(structEmbedMapStringAny),
		want:  addr(structEmbedMapStringAny{A: 1, X: nil, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/MergeN1/Nil"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: new(structEmbedMapStringAny),
		want:  addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": "buzz"}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/MergeN1/Empty"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: addr(structEmbedMapStringAny{X: jsonObject{}}),
		want:  addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": "buzz"}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/MergeN1/ObjectN1"),
		inBuf: `{"A":1,"fizz":{"charlie":"DELTA","echo":"foxtrot"},"B":2}`,
		inVal: addr(structEmbedMapStringAny{X: jsonObject{"fizz": jsonObject{
			"alpha":   "bravo",
			"charlie": "delta",
		}}}),
		want: addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": jsonObject{
			"alpha":   "bravo",
			"charlie": "DELTA",
			"echo":    "foxtrot",
		}}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/MergeN2/ObjectN1"),
		inBuf: `{"A":1,"fizz":"buzz","B":2,"foo": [ 1 , 2 , 3 ]}`,
		inVal: addr(structEmbedMapStringAny{X: jsonObject{"fizz": "wuzz"}}),
		want:  addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": "buzz", "foo": jsonArray{1.0, 2.0, 3.0}}, B: 2}),
	}, {
		name:    Name("Structs/EmbeddedFallback/MapStringAny/MergeInvalidValue"),
		inBuf:   `{"A":1,"fizz":nil,"B":2}`,
		inVal:   new(structEmbedMapStringAny),
		want:    addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": nil}}),
		wantErr: newInvalidCharacterError("i", "in literal null (expecting 'u')", len64(`{"A":1,"fizz":n`), "/fizz"),
	}, {
		name:    Name("Structs/EmbeddedFallback/MapStringAny/MergeInvalidValue/Existing"),
		inBuf:   `{"A":1,"fizz":nil,"B":2}`,
		inVal:   addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": true}}),
		want:    addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": true}}),
		wantErr: newInvalidCharacterError("i", "in literal null (expecting 'u')", len64(`{"A":1,"fizz":n`), "/fizz"),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/CaseSensitive"),
		inBuf: `{"A":1,"fizz":"buzz","B":2,"a":3}`,
		inVal: new(structEmbedMapStringAny),
		want:  addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": "buzz", "a": 3.0}, B: 2}),
	}, {
		name:    Name("Structs/EmbeddedFallback/MapStringAny/RejectDuplicateNames"),
		opts:    []Options{jsontext.AllowDuplicateNames(false)},
		inBuf:   `{"A":1,"fizz":"buzz","B":2,"fizz":"buzz"}`,
		inVal:   new(structEmbedMapStringAny),
		want:    addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": "buzz"}, B: 2}),
		wantErr: newDuplicateNameError("", []byte(`"fizz"`), len64(`{"A":1,"fizz":"buzz","B":2,`)),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/AllowDuplicateNames"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"A":1,"fizz":{"one":1,"two":-2},"B":2,"fizz":{"two":2,"three":3}}`,
		inVal: new(structEmbedMapStringAny),
		want:  addr(structEmbedMapStringAny{A: 1, X: jsonObject{"fizz": jsonObject{"one": 1.0, "two": 2.0, "three": 3.0}}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/Nested/Noop"),
		inBuf: `{}`,
		inVal: new(structEmbedPointerEmbedMapStringAny),
		want:  new(structEmbedPointerEmbedMapStringAny),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/Nested/Alloc"),
		inBuf: `{"A":1,"fizz":"buzz"}`,
		inVal: new(structEmbedPointerEmbedMapStringAny),
		want: addr(structEmbedPointerEmbedMapStringAny{
			X: &struct {
				A int
				X jsonObject `json:",embed"`
			}{A: 1, X: jsonObject{"fizz": "buzz"}},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringAny/Nested/Merge"),
		inBuf: `{"fizz":"buzz"}`,
		inVal: addr(structEmbedPointerEmbedMapStringAny{
			X: &struct {
				A int
				X jsonObject `json:",embed"`
			}{A: 1},
		}),
		want: addr(structEmbedPointerEmbedMapStringAny{
			X: &struct {
				A int
				X jsonObject `json:",embed"`
			}{A: 1, X: jsonObject{"fizz": "buzz"}},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerMapStringAny/Noop"),
		inBuf: `{"A":1,"B":2}`,
		inVal: new(structEmbedPointerMapStringAny),
		want:  addr(structEmbedPointerMapStringAny{A: 1, X: nil, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerMapStringAny/Alloc"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: new(structEmbedPointerMapStringAny),
		want:  addr(structEmbedPointerMapStringAny{A: 1, X: addr(jsonObject{"fizz": "buzz"}), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerMapStringAny/Merge"),
		inBuf: `{"A":1,"fizz":"wuzz","B":2}`,
		inVal: addr(structEmbedPointerMapStringAny{X: addr(jsonObject{"fizz": "buzz"})}),
		want:  addr(structEmbedPointerMapStringAny{A: 1, X: addr(jsonObject{"fizz": "wuzz"}), B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/PointerMapStringAny/Nested/Nil"),
		inBuf: `{"fizz":"buzz"}`,
		inVal: new(structEmbedEmbedPointerMapStringAny),
		want: addr(structEmbedEmbedPointerMapStringAny{
			X: struct {
				X *jsonObject `json:",embed"`
			}{X: addr(jsonObject{"fizz": "buzz"})},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringInt"),
		inBuf: `{"zero": 0, "one": 1, "two": 2}`,
		inVal: new(structEmbedMapStringInt),
		want: addr(structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 1, "two": 2},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringInt/Null"),
		inBuf: `{"zero": 0, "one": null, "two": 2}`,
		inVal: new(structEmbedMapStringInt),
		want: addr(structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 0, "two": 2},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringInt/Invalid"),
		inBuf: `{"zero": 0, "one": {}, "two": 2}`,
		inVal: new(structEmbedMapStringInt),
		want: addr(structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 0},
		}),
		wantErr: EU(nil).withPos(`{"zero": 0, "one": `, "/one").withType('{', T[int]()),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapStringInt/StringifiedNumbers"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `{"zero": "0", "one": "1", "two": "2"}`,
		inVal: new(structEmbedMapStringInt),
		want: addr(structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 1, "two": 2},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringInt"),
		inBuf: `{"zero": 0, "one": 1, "two": 2}`,
		inVal: new(structEmbedMapNamedStringInt),
		want: addr(structEmbedMapNamedStringInt{
			X: map[namedString]int{"zero": 0, "one": 1, "two": 2},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringInt/Null"),
		inBuf: `{"zero": 0, "one": null, "two": 2}`,
		inVal: new(structEmbedMapNamedStringInt),
		want: addr(structEmbedMapNamedStringInt{
			X: map[namedString]int{"zero": 0, "one": 0, "two": 2},
		}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringInt/Invalid"),
		inBuf: `{"zero": 0, "one": {}, "two": 2}`,
		inVal: new(structEmbedMapNamedStringInt),
		want: addr(structEmbedMapNamedStringInt{
			X: map[namedString]int{"zero": 0, "one": 0},
		}),
		wantErr: EU(nil).withPos(`{"zero": 0, "one": `, "/one").withType('{', T[int]()),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringInt/StringifiedNumbers"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `{"zero": "0", "one": 1, "two": "2"}`,
		inVal: new(structEmbedMapNamedStringInt),
		want: addr(structEmbedMapNamedStringInt{
			X: map[namedString]int{"zero": 0, "one": 0},
		}),
		wantErr: EU(nil).withPos(`{"zero": "0", "one": `, "/one").withType('0', T[int]()),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringAny/Noop"),
		inBuf: `{"A":1,"B":2}`,
		inVal: new(structEmbedMapNamedStringAny),
		want:  addr(structEmbedMapNamedStringAny{A: 1, X: nil, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringAny/MergeN1/Nil"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: new(structEmbedMapNamedStringAny),
		want:  addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": "buzz"}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringAny/MergeN1/Empty"),
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: addr(structEmbedMapNamedStringAny{X: map[namedString]any{}}),
		want:  addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": "buzz"}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringAny/MergeN1/ObjectN1"),
		inBuf: `{"A":1,"fizz":{"charlie":"DELTA","echo":"foxtrot"},"B":2}`,
		inVal: addr(structEmbedMapNamedStringAny{X: map[namedString]any{"fizz": jsonObject{
			"alpha":   "bravo",
			"charlie": "delta",
		}}}),
		want: addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": jsonObject{
			"alpha":   "bravo",
			"charlie": "DELTA",
			"echo":    "foxtrot",
		}}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringAny/MergeN2/ObjectN1"),
		inBuf: `{"A":1,"fizz":"buzz","B":2,"foo": [ 1 , 2 , 3 ]}`,
		inVal: addr(structEmbedMapNamedStringAny{X: map[namedString]any{"fizz": "wuzz"}}),
		want:  addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": "buzz", "foo": jsonArray{1.0, 2.0, 3.0}}, B: 2}),
	}, {
		name:    Name("Structs/EmbeddedFallback/MapNamedStringAny/MergeInvalidValue"),
		inBuf:   `{"A":1,"fizz":nil,"B":2}`,
		inVal:   new(structEmbedMapNamedStringAny),
		want:    addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": nil}}),
		wantErr: newInvalidCharacterError("i", "in literal null (expecting 'u')", len64(`{"A":1,"fizz":n`), "/fizz"),
	}, {
		name:    Name("Structs/EmbeddedFallback/MapNamedStringAny/MergeInvalidValue/Existing"),
		inBuf:   `{"A":1,"fizz":nil,"B":2}`,
		inVal:   addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": true}}),
		want:    addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": true}}),
		wantErr: newInvalidCharacterError("i", "in literal null (expecting 'u')", len64(`{"A":1,"fizz":n`), "/fizz"),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringAny/CaseSensitive"),
		inBuf: `{"A":1,"fizz":"buzz","B":2,"a":3}`,
		inVal: new(structEmbedMapNamedStringAny),
		want:  addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": "buzz", "a": 3.0}, B: 2}),
	}, {
		name:    Name("Structs/EmbeddedFallback/MapNamedStringAny/RejectDuplicateNames"),
		opts:    []Options{jsontext.AllowDuplicateNames(false)},
		inBuf:   `{"A":1,"fizz":"buzz","B":2,"fizz":"buzz"}`,
		inVal:   new(structEmbedMapNamedStringAny),
		want:    addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": "buzz"}, B: 2}),
		wantErr: newDuplicateNameError("", []byte(`"fizz"`), len64(`{"A":1,"fizz":"buzz","B":2,`)),
	}, {
		name:  Name("Structs/EmbeddedFallback/MapNamedStringAny/AllowDuplicateNames"),
		opts:  []Options{jsontext.AllowDuplicateNames(true)},
		inBuf: `{"A":1,"fizz":{"one":1,"two":-2},"B":2,"fizz":{"two":2,"three":3}}`,
		inVal: new(structEmbedMapNamedStringAny),
		want:  addr(structEmbedMapNamedStringAny{A: 1, X: map[namedString]any{"fizz": map[string]any{"one": 1.0, "two": 2.0, "three": 3.0}}, B: 2}),
	}, {
		name:  Name("Structs/EmbeddedFallback/RejectUnknownMembers"),
		opts:  []Options{RejectUnknownMembers(true)},
		inBuf: `{"A":1,"fizz":"buzz","B":2}`,
		inVal: new(structEmbedTextValue),
		want: addr(structEmbedTextValue{
			A: 1,
			X: jsontext.Value(`{"fizz":"buzz"}`),
			B: 2,
		}),
	}, {
		name:  Name("Structs/UnknownIgnored"),
		opts:  []Options{RejectUnknownMembers(false)},
		inBuf: `{"unknown":"fizzbuzz"}`,
		inVal: new(structAll),
		want:  new(structAll),
	}, {
		name:    Name("Structs/RejectUnknownMembers"),
		opts:    []Options{RejectUnknownMembers(true)},
		inBuf:   `{"unknown":"fizzbuzz"}`,
		inVal:   new(structAll),
		want:    new(structAll),
		wantErr: EU(ErrUnknownName).withPos(`{`, "/unknown").withType('"', T[structAll]()),
	}, {
		name:  Name("Structs/UnexportedIgnored"),
		inBuf: `{"ignored":"unused"}`,
		inVal: new(structUnexportedIgnored),
		want:  new(structUnexportedIgnored),
	}, {
		name:  Name("Structs/IgnoredUnexportedEmbedded"),
		inBuf: `{"namedString":"unused"}`,
		inVal: new(structIgnoredUnexportedEmbedded),
		want:  new(structIgnoredUnexportedEmbedded),
	}, {
		name:  Name("Structs/NoCase/Exact"),
		inBuf: `{"Aaa":"Aaa","AA_A":"AA_A","AaA":"AaA","AAa":"AAa","AAA":"AAA"}`,
		inVal: new(structNoCase),
		want:  addr(structNoCase{AaA: "AaA", AAa: "AAa", Aaa: "Aaa", AAA: "AAA", AA_A: "AA_A"}),
	}, {
		name:  Name("Structs/NoCase/CaseInsensitiveDefault/Ambiguous"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"aa_a":"aa_a"}`,
		inVal: new(structNoCase),
		want:  addr(structNoCase{AaA: "aa_a"}),
	}, {
		name:    Name("Structs/NoCase/CaseInsensitiveDefault/Ambiguous"),
		inBuf:   `{"aa_a":"aa_a"}`,
		inVal:   new(structNoCase),
		want:    addr(structNoCase{}),
		wantErr: EU(errAmbiguousName).withPos(`{`, "/aa_a").withType('"', T[structNoCase]()),
	}, {
		name:  Name("Structs/NoCase/MatchCaseSensitiveDelimiter"),
		opts:  []Options{unsupported("MatchCaseSensitiveDelimiter")},
		inBuf: `{"aa_a":"aa_a"}`,
		inVal: new(structNoCase),
		want:  addr(structNoCase{}),
	}, {
		name:  Name("Structs/NoCase/MatchCaseInsensitiveNames+MatchCaseSensitiveDelimiter"),
		opts:  []Options{MatchCaseInsensitiveNames(true), unsupported("MatchCaseSensitiveDelimiter")},
		inBuf: `{"aa_a":"aa_a"}`,
		inVal: new(structNoCase),
		want:  addr(structNoCase{AA_A: "aa_a"}),
	}, {
		name:  Name("Structs/NoCase/Merge/AllowDuplicateNames"),
		opts:  []Options{jsontext.AllowDuplicateNames(true), unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"AaA":"AaA","aaa":"aaa","aAa":"aAa"}`,
		inVal: new(structNoCase),
		want:  addr(structNoCase{AaA: "aAa"}),
	}, {
		name:    Name("Structs/NoCase/Merge/AllowDuplicateNames/Ambiguous"),
		opts:    []Options{jsontext.AllowDuplicateNames(true)},
		inBuf:   `{"AaA":"AaA","aaa":"aaa","aAa":"aAa"}`,
		inVal:   new(structNoCase),
		want:    addr(structNoCase{AaA: "AaA"}),
		wantErr: EU(errAmbiguousName).withPos(`{"AaA":"AaA",`, "/aaa").withType('"', T[structNoCase]()),
	}, {
		name:    Name("Structs/NoCase/Merge/RejectDuplicateNames/Ambiguous"),
		opts:    []Options{jsontext.AllowDuplicateNames(false)},
		inBuf:   `{"AaA":"AaA","aaa":"aaa"}`,
		inVal:   new(structNoCase),
		want:    addr(structNoCase{AaA: "AaA"}),
		wantErr: EU(errAmbiguousName).withPos(`{"AaA":"AaA",`, "/aaa").withType('"', T[structNoCase]()),
	}, {
		name:  Name("Structs/CaseSensitive"),
		inBuf: `{"BOOL": true, "STRING": "hello", "BYTES": "AQID", "INT": -64, "UINT": 64, "FLOAT": 3.14159}`,
		inVal: new(structScalars),
		want:  addr(structScalars{}),
	}, {
		name:  Name("Structs/DuplicateName/NoCase/ExactDifferent"),
		inBuf: `{"AAA":"AAA","AaA":"AaA","AAa":"AAa","Aaa":"Aaa"}`,
		inVal: addr(structNoCaseEmbedTextValue{}),
		want:  addr(structNoCaseEmbedTextValue{AAA: "AAA", AaA: "AaA", AAa: "AAa", Aaa: "Aaa"}),
	}, {
		name:    Name("Structs/DuplicateName/NoCase/ExactConflict"),
		inBuf:   `{"AAA":"AAA","AAA":"AAA"}`,
		inVal:   addr(structNoCaseEmbedTextValue{}),
		want:    addr(structNoCaseEmbedTextValue{AAA: "AAA"}),
		wantErr: newDuplicateNameError("", []byte(`"AAA"`), len64(`{"AAA":"AAA",`)),
	}, {
		name:  Name("Structs/DuplicateName/NoCase/OverwriteExact"),
		inBuf: `{"AAA":"after"}`,
		inVal: addr(structNoCaseEmbedTextValue{AAA: "before"}),
		want:  addr(structNoCaseEmbedTextValue{AAA: "after"}),
	}, {
		name:    Name("Structs/AmbiguousName/NoCase"),
		inBuf:   `{"aaa":"aaa"}`,
		inVal:   addr(structNoCaseEmbedTextValue{}),
		want:    addr(structNoCaseEmbedTextValue{}),
		wantErr: EU(errAmbiguousName).withPos(`{`, "/aaa").withType('"', T[structNoCaseEmbedTextValue]()),
	}, {
		name:  Name("Structs/DuplicateName/Embed/Unknown"),
		inBuf: `{"unknown":""}`,
		inVal: addr(structNoCaseEmbedTextValue{}),
		want:  addr(structNoCaseEmbedTextValue{X: jsontext.Value(`{"unknown":""}`)}),
	}, {
		name:  Name("Structs/DuplicateName/Embed/UnknownMerge"),
		inBuf: `{"unknown":""}`,
		inVal: addr(structNoCaseEmbedTextValue{X: jsontext.Value(`{"unknown":""}`)}),
		want:  addr(structNoCaseEmbedTextValue{X: jsontext.Value(`{"unknown":"","unknown":""}`)}),
	}, {
		name:  Name("Structs/DuplicateName/Embed/NoCaseOkay"),
		inBuf: `{"b":"","B":""}`,
		inVal: addr(structNoCaseEmbedTextValue{}),
		want:  addr(structNoCaseEmbedTextValue{X: jsontext.Value(`{"b":"","B":""}`)}),
	}, {
		name:    Name("Structs/DuplicateName/Embed/ExactConflict"),
		inBuf:   `{"b":"","b":""}`,
		inVal:   addr(structNoCaseEmbedTextValue{}),
		want:    addr(structNoCaseEmbedTextValue{X: jsontext.Value(`{"b":""}`)}),
		wantErr: newDuplicateNameError("", []byte(`"b"`), len64(`{"b":"",`)),
	}, {
		name:    Name("Structs/Invalid/ErrUnexpectedEOF"),
		inBuf:   ``,
		inVal:   addr(structAll{}),
		want:    addr(structAll{}),
		wantErr: &jsontext.SyntacticError{Err: io.ErrUnexpectedEOF},
	}, {
		name:    Name("Structs/Invalid/ErrUnexpectedEOF"),
		inBuf:   " \n\r\t",
		inVal:   addr(structAll{}),
		want:    addr(structAll{}),
		wantErr: &jsontext.SyntacticError{Err: io.ErrUnexpectedEOF, ByteOffset: len64(" \n\r\t")},
	}, {
		name:    Name("Structs/Invalid/NestedErrUnexpectedEOF"),
		inBuf:   `{"Pointer":`,
		inVal:   addr(structAll{}),
		want:    addr(structAll{Pointer: new(structAll)}),
		wantErr: &jsontext.SyntacticError{ByteOffset: len64(`{"Pointer":`), JSONPointer: "/Pointer", Err: io.ErrUnexpectedEOF},
	}, {
		name:    Name("Structs/Invalid/Conflicting"),
		inBuf:   `{}`,
		inVal:   addr(structConflicting{}),
		want:    addr(structConflicting{}),
		wantErr: EU(errors.New(`Go struct fields A and B conflict over JSON object name "conflict"`)).withType('{', T[structConflicting]()),
	}, {
		name:    Name("Structs/Invalid/NoneExported"),
		inBuf:   ` {}`,
		inVal:   addr(structNoneExported{}),
		want:    addr(structNoneExported{}),
		wantErr: EU(errNoExportedFields).withPos(` `, "").withType('{', T[structNoneExported]()),
	}, {
		name:    Name("Structs/Invalid/MalformedTag"),
		inBuf:   `{}`,
		inVal:   addr(structMalformedTag{}),
		want:    addr(structMalformedTag{}),
		wantErr: EU(errors.New("Go struct field Malformed has malformed `json` tag: invalid character '\"' at start of option (expecting Unicode letter)")).withType('{', T[structMalformedTag]()),
	}, {
		name:    Name("Structs/Invalid/UnexportedTag"),
		inBuf:   `{}`,
		inVal:   addr(structUnexportedTag{}),
		want:    addr(structUnexportedTag{}),
		wantErr: EU(errors.New("unexported Go struct field unexported cannot have non-ignored `json:\"name\"` tag")).withType('{', T[structUnexportedTag]()),
	}, {
		name:    Name("Structs/Invalid/ExportedEmbedded"),
		inBuf:   `{"NamedString":"hello"}`,
		inVal:   addr(structExportedEmbedded{}),
		want:    addr(structExportedEmbedded{}),
		wantErr: EU(errors.New("embedded Go struct field NamedString of non-struct type must be explicitly given a JSON name")).withType('{', T[structExportedEmbedded]()),
	}, {
		name:  Name("Structs/Valid/ExportedEmbedded"),
		opts:  []Options{unsupported("ReportErrorsWithLegacySemantics")},
		inBuf: `{"NamedString":"hello"}`,
		inVal: addr(structExportedEmbedded{}),
		want:  addr(structExportedEmbedded{"hello"}),
	}, {
		name:  Name("Structs/Valid/ExportedEmbeddedTag"),
		inBuf: `{"name":"hello"}`,
		inVal: addr(structExportedEmbeddedTag{}),
		want:  addr(structExportedEmbeddedTag{"hello"}),
	}, {
		name:    Name("Structs/Invalid/UnexportedEmbedded"),
		inBuf:   `{}`,
		inVal:   addr(structUnexportedEmbedded{}),
		want:    addr(structUnexportedEmbedded{}),
		wantErr: EU(errors.New("embedded Go struct field namedString of non-struct type must be explicitly given a JSON name")).withType('{', T[structUnexportedEmbedded]()),
	}, {
		name:  Name("Structs/UnexportedEmbeddedStruct"),
		inBuf: `{"Bool":true,"FizzBuzz":5,"Addr":"192.168.0.1"}`,
		inVal: addr(structUnexportedEmbeddedStruct{}),
		want:  addr(structUnexportedEmbeddedStruct{structOmitZeroAll{Bool: true}, 5, structNestedAddr{netip.AddrFrom4([4]byte{192, 168, 0, 1})}}),
	}, {
		name:    Name("Structs/UnexportedEmbeddedStructPointer/Nil"),
		inBuf:   `{"Bool":true,"FizzBuzz":5}`,
		inVal:   addr(structUnexportedEmbeddedStructPointer{}),
		wantErr: EU(errNilField).withPos(`{"Bool":`, "/Bool").withType(0, T[structUnexportedEmbeddedStructPointer]()),
	}, {
		name:    Name("Structs/UnexportedEmbeddedStructPointer/Nil"),
		inBuf:   `{"FizzBuzz":5,"Addr":"192.168.0.1"}`,
		inVal:   addr(structUnexportedEmbeddedStructPointer{}),
		wantErr: EU(errNilField).withPos(`{"FizzBuzz":5,"Addr":`, "/Addr").withType(0, T[structUnexportedEmbeddedStructPointer]()),
	}, {
		name:  Name("Structs/UnexportedEmbeddedStructPointer/Nil"),
		inBuf: `{"Bool":true,"FizzBuzz":10,"Addr":"192.168.0.1"}`,
		inVal: addr(structUnexportedEmbeddedStructPointer{&structOmitZeroAll{Int: 5}, 5, &structNestedAddr{netip.AddrFrom4([4]byte{127, 0, 0, 1})}}),
		want:  addr(structUnexportedEmbeddedStructPointer{&structOmitZeroAll{Bool: true, Int: 5}, 10, &structNestedAddr{netip.AddrFrom4([4]byte{192, 168, 0, 1})}}),
	}, {
		name: Name("Structs/Unknown"),
		inBuf: `{
	"object0": {},
	"object1": {"key1": "value"},
	"object2": {"key1": "value", "key2": "value"},
	"objects": {"":{"":{"":{}}}},
	"array0": [],
	"array1": ["value1"],
	"array2": ["value1", "value2"],
	"array": [[[]]],
	"scalars": [null, false, true, "string", 12.345]
}`,
		inVal: addr(struct{}{}),
		want:  addr(struct{}{}),
	}, {
		name:  Name("Structs/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `{"Field":"Value"}`,
		inVal: addr(struct{ Field string }{}),
		want:  addr(struct{ Field string }{"Value"}),
	}, {
		name:  Name("Slices/Null"),
		inBuf: `null`,
		inVal: addr([]string{"something"}),
		want:  addr([]string(nil)),
	}, {
		name:  Name("Slices/Bool"),
		inBuf: `[true,false]`,
		inVal: new([]bool),
		want:  addr([]bool{true, false}),
	}, {
		name:  Name("Slices/String"),
		inBuf: `["hello","goodbye"]`,
		inVal: new([]string),
		want:  addr([]string{"hello", "goodbye"}),
	}, {
		name:  Name("Slices/Bytes"),
		inBuf: `["aGVsbG8=","Z29vZGJ5ZQ=="]`,
		inVal: new([][]byte),
		want:  addr([][]byte{[]byte("hello"), []byte("goodbye")}),
	}, {
		name:  Name("Slices/Int"),
		inBuf: `[-2,-1,0,1,2]`,
		inVal: new([]int),
		want:  addr([]int{-2, -1, 0, 1, 2}),
	}, {
		name:  Name("Slices/Uint"),
		inBuf: `[0,1,2,3,4]`,
		inVal: new([]uint),
		want:  addr([]uint{0, 1, 2, 3, 4}),
	}, {
		name:  Name("Slices/Float"),
		inBuf: `[3.14159,12.34]`,
		inVal: new([]float64),
		want:  addr([]float64{3.14159, 12.34}),
	}, {
		// NOTE: The semantics differs from v1, where the slice length is reset
		// and new elements are appended to the end.
		// See https://go.dev/issue/21092.
		name:  Name("Slices/Merge"),
		inBuf: `[{"k3":"v3"},{"k4":"v4"}]`,
		inVal: addr([]map[string]string{{"k1": "v1"}, {"k2": "v2"}}[:1]),
		want:  addr([]map[string]string{{"k3": "v3"}, {"k4": "v4"}}),
	}, {
		name:    Name("Slices/Invalid/Channel"),
		inBuf:   `["hello"]`,
		inVal:   new([]chan string),
		want:    addr([]chan string{nil}),
		wantErr: EU(nil).withPos(`[`, "/0").withType(0, T[chan string]()),
	}, {
		name:  Name("Slices/RecursiveSlice"),
		inBuf: `[[],[],[[]],[[],[]]]`,
		inVal: new(recursiveSlice),
		want: addr(recursiveSlice{
			{},
			{},
			{{}},
			{{}, {}},
		}),
	}, {
		name:    Name("Slices/Invalid/Bool"),
		inBuf:   `true`,
		inVal:   addr([]string{"nochange"}),
		want:    addr([]string{"nochange"}),
		wantErr: EU(nil).withType('t', T[[]string]()),
	}, {
		name:    Name("Slices/Invalid/String"),
		inBuf:   `""`,
		inVal:   addr([]string{"nochange"}),
		want:    addr([]string{"nochange"}),
		wantErr: EU(nil).withType('"', T[[]string]()),
	}, {
		name:    Name("Slices/Invalid/Number"),
		inBuf:   `0`,
		inVal:   addr([]string{"nochange"}),
		want:    addr([]string{"nochange"}),
		wantErr: EU(nil).withType('0', T[[]string]()),
	}, {
		name:    Name("Slices/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr([]string{"nochange"}),
		want:    addr([]string{"nochange"}),
		wantErr: EU(nil).withType('{', T[[]string]()),
	}, {
		name:  Name("Slices/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `[false,true]`,
		inVal: addr([]bool{true, false}),
		want:  addr([]bool{false, true}),
	}, {
		name:  Name("Arrays/Null"),
		inBuf: `null`,
		inVal: addr([1]string{"something"}),
		want:  addr([1]string{}),
	}, {
		name:  Name("Arrays/Bool"),
		inBuf: `[true,false]`,
		inVal: new([2]bool),
		want:  addr([2]bool{true, false}),
	}, {
		name:  Name("Arrays/String"),
		inBuf: `["hello","goodbye"]`,
		inVal: new([2]string),
		want:  addr([2]string{"hello", "goodbye"}),
	}, {
		name:  Name("Arrays/Bytes"),
		inBuf: `["aGVsbG8=","Z29vZGJ5ZQ=="]`,
		inVal: new([2][]byte),
		want:  addr([2][]byte{[]byte("hello"), []byte("goodbye")}),
	}, {
		name:  Name("Arrays/Int"),
		inBuf: `[-2,-1,0,1,2]`,
		inVal: new([5]int),
		want:  addr([5]int{-2, -1, 0, 1, 2}),
	}, {
		name:  Name("Arrays/Uint"),
		inBuf: `[0,1,2,3,4]`,
		inVal: new([5]uint),
		want:  addr([5]uint{0, 1, 2, 3, 4}),
	}, {
		name:  Name("Arrays/Float"),
		inBuf: `[3.14159,12.34]`,
		inVal: new([2]float64),
		want:  addr([2]float64{3.14159, 12.34}),
	}, {
		// NOTE: The semantics differs from v1, where elements are not merged.
		// This is to maintain consistent merge semantics with slices.
		name:  Name("Arrays/Merge"),
		inBuf: `[{"k3":"v3"},{"k4":"v4"}]`,
		inVal: addr([2]map[string]string{{"k1": "v1"}, {"k2": "v2"}}),
		want:  addr([2]map[string]string{{"k3": "v3"}, {"k4": "v4"}}),
	}, {
		name:    Name("Arrays/Invalid/Channel"),
		inBuf:   `["hello"]`,
		inVal:   new([1]chan string),
		want:    new([1]chan string),
		wantErr: EU(nil).withPos(`[`, "/0").withType(0, T[chan string]()),
	}, {
		name:    Name("Arrays/Invalid/Underflow"),
		inBuf:   `{"F":[   ]}`,
		inVal:   new(struct{ F [1]string }),
		want:    addr(struct{ F [1]string }{}),
		wantErr: EU(errArrayUnderflow).withPos(`{"F":[   `, "/F").withType(']', T[[1]string]()),
	}, {
		name:  Name("Arrays/Invalid/Underflow/UnmarshalArrayFromAnyLength"),
		opts:  []Options{unsupported("UnmarshalArrayFromAnyLength")},
		inBuf: `[-1,-2]`,
		inVal: addr([4]int{1, 2, 3, 4}),
		want:  addr([4]int{-1, -2, 0, 0}),
	}, {
		name:    Name("Arrays/Invalid/Overflow"),
		inBuf:   `["1","2"]`,
		inVal:   new([1]string),
		want:    addr([1]string{"1"}),
		wantErr: EU(errArrayOverflow).withPos(`["1","2"`, "").withType(']', T[[1]string]()),
	}, {
		name:  Name("Arrays/Invalid/Overflow/UnmarshalArrayFromAnyLength"),
		opts:  []Options{unsupported("UnmarshalArrayFromAnyLength")},
		inBuf: `[-1,-2,-3,-4,-5,-6]`,
		inVal: addr([4]int{1, 2, 3, 4}),
		want:  addr([4]int{-1, -2, -3, -4}),
	}, {
		name:    Name("Arrays/Invalid/Bool"),
		inBuf:   `true`,
		inVal:   addr([1]string{"nochange"}),
		want:    addr([1]string{"nochange"}),
		wantErr: EU(nil).withType('t', T[[1]string]()),
	}, {
		name:    Name("Arrays/Invalid/String"),
		inBuf:   `""`,
		inVal:   addr([1]string{"nochange"}),
		want:    addr([1]string{"nochange"}),
		wantErr: EU(nil).withType('"', T[[1]string]()),
	}, {
		name:    Name("Arrays/Invalid/Number"),
		inBuf:   `0`,
		inVal:   addr([1]string{"nochange"}),
		want:    addr([1]string{"nochange"}),
		wantErr: EU(nil).withType('0', T[[1]string]()),
	}, {
		name:    Name("Arrays/Invalid/Object"),
		inBuf:   `{}`,
		inVal:   addr([1]string{"nochange"}),
		want:    addr([1]string{"nochange"}),
		wantErr: EU(nil).withType('{', T[[1]string]()),
	}, {
		name:  Name("Arrays/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `[false,true]`,
		inVal: addr([2]bool{true, false}),
		want:  addr([2]bool{false, true}),
	}, {
		name:  Name("Pointers/NullL0"),
		inBuf: `null`,
		inVal: new(*string),
		want:  addr((*string)(nil)),
	}, {
		name:  Name("Pointers/NullL1"),
		inBuf: `null`,
		inVal: addr(new(*string)),
		want:  addr((**string)(nil)),
	}, {
		name:  Name("Pointers/Bool"),
		inBuf: `true`,
		inVal: addr(new(bool)),
		want:  addr(addr(true)),
	}, {
		name:  Name("Pointers/String"),
		inBuf: `"hello"`,
		inVal: addr(new(string)),
		want:  addr(addr("hello")),
	}, {
		name:  Name("Pointers/Bytes"),
		inBuf: `"aGVsbG8="`,
		inVal: addr(new([]byte)),
		want:  addr(addr([]byte("hello"))),
	}, {
		name:  Name("Pointers/Int"),
		inBuf: `-123`,
		inVal: addr(new(int)),
		want:  addr(addr(int(-123))),
	}, {
		name:  Name("Pointers/Uint"),
		inBuf: `123`,
		inVal: addr(new(int)),
		want:  addr(addr(int(123))),
	}, {
		name:  Name("Pointers/Float"),
		inBuf: `123.456`,
		inVal: addr(new(float64)),
		want:  addr(addr(float64(123.456))),
	}, {
		name:  Name("Pointers/Allocate"),
		inBuf: `"hello"`,
		inVal: addr((*string)(nil)),
		want:  addr(addr("hello")),
	}, {
		name:  Name("Points/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `true`,
		inVal: addr(new(bool)),
		want:  addr(addr(true)),
	}, {
		name:  Name("Interfaces/Empty/Null"),
		inBuf: `null`,
		inVal: new(any),
		want:  new(any),
	}, {
		name:  Name("Interfaces/NonEmpty/Null"),
		inBuf: `null`,
		inVal: new(io.Reader),
		want:  new(io.Reader),
	}, {
		name:    Name("Interfaces/NonEmpty/Invalid"),
		inBuf:   `"hello"`,
		inVal:   new(io.Reader),
		want:    new(io.Reader),
		wantErr: EU(errNilInterface).withType(0, T[io.Reader]()),
	}, {
		name:  Name("Interfaces/Empty/False"),
		inBuf: `false`,
		inVal: new(any),
		want: func() any {
			var vi any = false
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Empty/True"),
		inBuf: `true`,
		inVal: new(any),
		want: func() any {
			var vi any = true
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Empty/String"),
		inBuf: `"string"`,
		inVal: new(any),
		want: func() any {
			var vi any = "string"
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Empty/Number"),
		inBuf: `3.14159`,
		inVal: new(any),
		want: func() any {
			var vi any = 3.14159
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Empty/Object"),
		inBuf: `{"k":"v"}`,
		inVal: new(any),
		want: func() any {
			var vi any = map[string]any{"k": "v"}
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Empty/Array"),
		inBuf: `["v"]`,
		inVal: new(any),
		want: func() any {
			var vi any = []any{"v"}
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/NamedAny/String"),
		inBuf: `"string"`,
		inVal: new(namedAny),
		want: func() namedAny {
			var vi namedAny = "string"
			return &vi
		}(),
	}, {
		name:    Name("Interfaces/Invalid"),
		inBuf:   `]`,
		inVal:   new(any),
		want:    new(any),
		wantErr: newInvalidCharacterError("]", "at start of value", 0, ""),
	}, {
		// NOTE: The semantics differs from v1,
		// where existing map entries were not merged into.
		// See https://go.dev/issue/26946.
		// See https://go.dev/issue/33993.
		name:  Name("Interfaces/Merge/Map"),
		inBuf: `{"k2":"v2"}`,
		inVal: func() any {
			var vi any = map[string]string{"k1": "v1"}
			return &vi
		}(),
		want: func() any {
			var vi any = map[string]string{"k1": "v1", "k2": "v2"}
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Merge/Struct"),
		inBuf: `{"Array":["goodbye"]}`,
		inVal: func() any {
			var vi any = structAll{String: "hello"}
			return &vi
		}(),
		want: func() any {
			var vi any = structAll{String: "hello", Array: [1]string{"goodbye"}}
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Merge/NamedInt"),
		inBuf: `64`,
		inVal: func() any {
			var vi any = namedInt64(-64)
			return &vi
		}(),
		want: func() any {
			var vi any = namedInt64(+64)
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `true`,
		inVal: new(any),
		want: func() any {
			var vi any = true
			return &vi
		}(),
	}, {
		name:  Name("Interfaces/Any"),
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{[]any{nil, false, true, "", 0.0, map[string]any{}, []any{}}}),
	}, {
		name:  Name("Interfaces/Any/Named"),
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X namedAny }),
		want:  addr(struct{ X namedAny }{[]any{nil, false, true, "", 0.0, map[string]any{}, []any{}}}),
	}, {
		name:  Name("Interfaces/Any/Stringified"),
		opts:  []Options{StringifyNumbers(true)},
		inBuf: `{"X":"0"}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{"0"}),
	}, {
		name:  Name("Interfaces/Any/Maps/NonEmpty"),
		inBuf: `{"X":{"fizz":"buzz"}}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{map[string]any{"fizz": "buzz"}}),
	}, {
		name:    Name("Interfaces/Any/Maps/RejectDuplicateNames"),
		inBuf:   `{"X":{"fizz":"buzz","fizz":true}}`,
		inVal:   new(struct{ X any }),
		want:    addr(struct{ X any }{map[string]any{"fizz": "buzz"}}),
		wantErr: newDuplicateNameError("/X", []byte(`"fizz"`), len64(`{"X":{"fizz":"buzz",`)),
	}, {
		name:    Name("Interfaces/Any/Maps/AllowDuplicateNames"),
		opts:    []Options{jsontext.AllowDuplicateNames(true)},
		inBuf:   `{"X":{"fizz":"buzz","fizz":true}}`,
		inVal:   new(struct{ X any }),
		want:    addr(struct{ X any }{map[string]any{"fizz": "buzz"}}),
		wantErr: EU(nil).withPos(`{"X":{"fizz":"buzz","fizz":`, "/X/fizz").withType('t', T[string]()),
	}, {
		name:  Name("Interfaces/Any/Slices/NonEmpty"),
		inBuf: `{"X":["fizz","buzz"]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{[]any{"fizz", "buzz"}}),
	}, {
		name:  Name("Methods/NilPointer/Null"),
		inBuf: `{"X":null}`,
		inVal: addr(struct{ X *allMethods }{X: (*allMethods)(nil)}),
		want:  addr(struct{ X *allMethods }{X: (*allMethods)(nil)}), // method should not be called
	}, {
		name:  Name("Methods/NilPointer/Value"),
		inBuf: `{"X":"value"}`,
		inVal: addr(struct{ X *allMethods }{X: (*allMethods)(nil)}),
		want:  addr(struct{ X *allMethods }{X: &allMethods{method: "UnmarshalJSONFrom", value: []byte(`"value"`)}}),
	}, {
		name:  Name("Methods/NilInterface/Null"),
		inBuf: `{"X":null}`,
		inVal: addr(struct{ X MarshalerTo }{X: (*allMethods)(nil)}),
		want:  addr(struct{ X MarshalerTo }{X: nil}), // interface value itself is nil'd out
	}, {
		name:  Name("Methods/NilInterface/Value"),
		inBuf: `{"X":"value"}`,
		inVal: addr(struct{ X MarshalerTo }{X: (*allMethods)(nil)}),
		want:  addr(struct{ X MarshalerTo }{X: &allMethods{method: "UnmarshalJSONFrom", value: []byte(`"value"`)}}),
	}, {
		name:  Name("Methods/AllMethods"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X *allMethods }),
		want:  addr(struct{ X *allMethods }{X: &allMethods{method: "UnmarshalJSONFrom", value: []byte(`"hello"`)}}),
	}, {
		name:  Name("Methods/AllMethodsExceptJSONv2"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X *allMethodsExceptJSONv2 }),
		want:  addr(struct{ X *allMethodsExceptJSONv2 }{X: &allMethodsExceptJSONv2{allMethods: allMethods{method: "UnmarshalJSON", value: []byte(`"hello"`)}}}),
	}, {
		name:  Name("Methods/AllMethodsExceptJSONv1"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X *allMethodsExceptJSONv1 }),
		want:  addr(struct{ X *allMethodsExceptJSONv1 }{X: &allMethodsExceptJSONv1{allMethods: allMethods{method: "UnmarshalJSONFrom", value: []byte(`"hello"`)}}}),
	}, {
		name:  Name("Methods/AllMethodsExceptText"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X *allMethodsExceptText }),
		want:  addr(struct{ X *allMethodsExceptText }{X: &allMethodsExceptText{allMethods: allMethods{method: "UnmarshalJSONFrom", value: []byte(`"hello"`)}}}),
	}, {
		name:  Name("Methods/OnlyMethodJSONv2"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X *onlyMethodJSONv2 }),
		want:  addr(struct{ X *onlyMethodJSONv2 }{X: &onlyMethodJSONv2{allMethods: allMethods{method: "UnmarshalJSONFrom", value: []byte(`"hello"`)}}}),
	}, {
		name:  Name("Methods/OnlyMethodJSONv1"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X *onlyMethodJSONv1 }),
		want:  addr(struct{ X *onlyMethodJSONv1 }{X: &onlyMethodJSONv1{allMethods: allMethods{method: "UnmarshalJSON", value: []byte(`"hello"`)}}}),
	}, {
		name:  Name("Methods/OnlyMethodText"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X *onlyMethodText }),
		want:  addr(struct{ X *onlyMethodText }{X: &onlyMethodText{allMethods: allMethods{method: "UnmarshalText", value: []byte(`hello`)}}}),
	}, {
		name:  Name("Methods/Text/Null"),
		inBuf: `{"X":null}`,
		inVal: addr(struct{ X unmarshalTextFunc }{unmarshalTextFunc(func(b []byte) error {
			return errMustNotCall
		})}),
		want: addr(struct{ X unmarshalTextFunc }{nil}),
	}, {
		name:  Name("Methods/IP"),
		inBuf: `"192.168.0.100"`,
		inVal: new(net.IP),
		want:  addr(net.IPv4(192, 168, 0, 100)),
	}, {
		// NOTE: Fixes https://go.dev/issue/46516.
		name:  Name("Methods/Anonymous"),
		inBuf: `{"X":"hello"}`,
		inVal: new(struct{ X struct{ allMethods } }),
		want:  addr(struct{ X struct{ allMethods } }{X: struct{ allMethods }{allMethods{method: "UnmarshalJSONFrom", value: []byte(`"hello"`)}}}),
	}, {
		// NOTE: Fixes https://go.dev/issue/22967.
		name:  Name("Methods/Addressable"),
		inBuf: `{"V":"hello","M":{"K":"hello"},"I":"hello"}`,
		inVal: addr(struct {
			V allMethods
			M map[string]allMethods
			I any
		}{
			I: allMethods{}, // need to initialize with concrete value
		}),
		want: addr(struct {
			V allMethods
			M map[string]allMethods
			I any
		}{
			V: allMethods{method: "UnmarshalJSONFrom", value: []byte(`"hello"`)},
			M: map[string]allMethods{"K": {method: "UnmarshalJSONFrom", value: []byte(`"hello"`)}},
			I: allMethods{method: "UnmarshalJSONFrom", value: []byte(`"hello"`)},
		}),
	}, {
		// NOTE: Fixes https://go.dev/issue/29732.
		name:  Name("Methods/MapKey/JSONv2"),
		inBuf: `{"k1":"v1b","k2":"v2"}`,
		inVal: addr(map[structMethodJSONv2]string{{"k1"}: "v1a", {"k3"}: "v3"}),
		want:  addr(map[structMethodJSONv2]string{{"k1"}: "v1b", {"k2"}: "v2", {"k3"}: "v3"}),
	}, {
		// NOTE: Fixes https://go.dev/issue/29732.
		name:  Name("Methods/MapKey/JSONv1"),
		inBuf: `{"k1":"v1b","k2":"v2"}`,
		inVal: addr(map[structMethodJSONv1]string{{"k1"}: "v1a", {"k3"}: "v3"}),
		want:  addr(map[structMethodJSONv1]string{{"k1"}: "v1b", {"k2"}: "v2", {"k3"}: "v3"}),
	}, {
		name:  Name("Methods/MapKey/Text"),
		inBuf: `{"k1":"v1b","k2":"v2"}`,
		inVal: addr(map[structMethodText]string{{"k1"}: "v1a", {"k3"}: "v3"}),
		want:  addr(map[structMethodText]string{{"k1"}: "v1b", {"k2"}: "v2", {"k3"}: "v3"}),
	}, {
		name:  Name("Methods/JSONv2/ErrUnsupported"),
		inBuf: `{"fizz":123}`,
		inVal: addr(unsupportedMethodJSONv2{}),
		want:  addr(unsupportedMethodJSONv2{"called": 1, "fizz": 123}),
	}, {
		name:  Name("Methods/Invalid/JSONv2/Error"),
		inBuf: `{}`,
		inVal: addr(unmarshalJSONv2Func(func(*jsontext.Decoder) error {
			return errSomeError
		})),
		wantErr: EU(errSomeError).withType(0, T[unmarshalJSONv2Func]()),
	}, {
		name:  Name("Methods/Invalid/JSONv2/TooFew"),
		inBuf: `{}`,
		inVal: addr(unmarshalJSONv2Func(func(*jsontext.Decoder) error {
			return nil // do nothing
		})),
		wantErr: EU(errNonSingularValue).withType(0, T[unmarshalJSONv2Func]()),
	}, {
		name:  Name("Methods/Invalid/JSONv2/TooMany"),
		inBuf: `{}{}`,
		inVal: addr(unmarshalJSONv2Func(func(dec *jsontext.Decoder) error {
			dec.ReadValue()
			dec.ReadValue()
			return nil
		})),
		wantErr: EU(errNonSingularValue).withPos(`{}`, "").withType(0, T[unmarshalJSONv2Func]()),
	}, {
		name:  Name("Methods/Invalid/JSONv2/ErrUnsupported"),
		inBuf: `{}`,
		inVal: addr(unmarshalJSONv2Func(func(*jsontext.Decoder) error {
			return errors.ErrUnsupported
		})),
		wantErr: EU(nil).withType(0, T[unmarshalJSONv2Func]()),
	}, {
		name:  Name("Methods/Invalid/JSONv1/Error"),
		inBuf: `{}`,
		inVal: addr(unmarshalJSONv1Func(func([]byte) error {
			return errSomeError
		})),
		wantErr: EU(errSomeError).withType('{', T[unmarshalJSONv1Func]()),
	}, {
		name:  Name("Methods/Invalid/JSONv1/ErrUnsupported"),
		inBuf: `{}`,
		inVal: addr(unmarshalJSONv1Func(func([]byte) error {
			return errors.ErrUnsupported
		})),
		wantErr: EU(wrapErrUnsupported(errors.ErrUnsupported, "UnmarshalJSON method")).withType('{', T[unmarshalJSONv1Func]()),
	}, {
		name:  Name("Methods/Invalid/Text/Error"),
		inBuf: `"value"`,
		inVal: addr(unmarshalTextFunc(func([]byte) error {
			return errSomeError
		})),
		wantErr: EU(errSomeError).withType('"', T[unmarshalTextFunc]()),
	}, {
		name:  Name("Methods/Invalid/Text/Syntax"),
		inBuf: `{}`,
		inVal: addr(unmarshalTextFunc(func([]byte) error {
			panic("should not be called")
		})),
		wantErr: EU(errNonStringValue).withType('{', T[unmarshalTextFunc]()),
	}, {
		name:  Name("Methods/Invalid/Text/ErrUnsupported"),
		inBuf: `"value"`,
		inVal: addr(unmarshalTextFunc(func([]byte) error {
			return errors.ErrUnsupported
		})),
		wantErr: EU(wrapErrUnsupported(errors.ErrUnsupported, "UnmarshalText method")).withType('"', T[unmarshalTextFunc]()),
	}, {
		name:  Name("Duration/Null"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D1":null,"D2":null}`,
		inVal: addr(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{1, 1}),
		want: addr(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{0, 0}),
	}, {
		name:  Name("Duration/Zero"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D1":"0s","D2":0}`,
		inVal: addr(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{1, 1}),
		want: addr(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{0, 0}),
	}, {
		name:  Name("Duration/Positive"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D1":"34293h33m9.123456789s","D2":123456789123456789}`,
		inVal: new(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}),
		want: addr(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{
			123456789123456789,
			123456789123456789,
		}),
	}, {
		name:  Name("Duration/Negative"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D1":"-34293h33m9.123456789s","D2":-123456789123456789}`,
		inVal: new(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}),
		want: addr(struct {
			D1 time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
			D2 time.Duration `json:",format:nano"`
		}{
			-123456789123456789,
			-123456789123456789,
		}),
	}, {
		name:  Name("Duration/Nanos/String"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":"12345"}`,
		inVal: addr(struct {
			D time.Duration `json:",string,format:nano"`
		}{1}),
		want: addr(struct {
			D time.Duration `json:",string,format:nano"`
		}{12345}),
	}, {
		name:  Name("Duration/Nanos/String/Invalid"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":"+12345"}`,
		inVal: addr(struct {
			D time.Duration `json:",string,format:nano"`
		}{1}),
		want: addr(struct {
			D time.Duration `json:",string,format:nano"`
		}{1}),
		wantErr: EU(fmt.Errorf(`invalid duration "+12345": %w`, strconv.ErrSyntax)).withPos(`{"D":`, "/D").withType('"', timeDurationType),
	}, {
		name:  Name("Duration/Nanos/Mismatch"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":"34293h33m9.123456789s"}`,
		inVal: addr(struct {
			D time.Duration `json:",format:nano"`
		}{1}),
		want: addr(struct {
			D time.Duration `json:",format:nano"`
		}{1}),
		wantErr: EU(nil).withPos(`{"D":`, "/D").withType('"', timeDurationType),
	}, {
		name:  Name("Duration/Nanos"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":1.324}`,
		inVal: addr(struct {
			D time.Duration `json:",format:nano"`
		}{-1}),
		want: addr(struct {
			D time.Duration `json:",format:nano"`
		}{1}),
	}, {
		name:  Name("Duration/String/Mismatch"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":-123456789123456789}`,
		inVal: addr(struct {
			D time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
		}{1}),
		want: addr(struct {
			D time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
		}{1}),
		wantErr: EU(nil).withPos(`{"D":`, "/D").withType('0', timeDurationType),
	}, {
		name:  Name("Duration/String/Invalid"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":"5minkutes"}`,
		inVal: addr(struct {
			D time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
		}{1}),
		want: addr(struct {
			D time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
		}{1}),
		wantErr: EU(func() error {
			_, err := time.ParseDuration("5minkutes")
			return err
		}()).withPos(`{"D":`, "/D").withType('"', timeDurationType),
	}, {
		name:  Name("Duration/Syntax/Invalid"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":x}`,
		inVal: addr(struct {
			D time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
		}{1}),
		want: addr(struct {
			D time.Duration `json:",format:units"` // TODO(https://go.dev/issue/71631): Remove the format flag.
		}{1}),
		wantErr: newInvalidCharacterError("x", "at start of value", len64(`{"D":`), "/D"),
	}, {
		name: Name("Duration/Format"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{
			"D1": "12h34m56.078090012s",
			"D2": "12h34m56.078090012s",
			"D3": 45296.078090012,
			"D4": "45296.078090012",
			"D5": 45296078.090012,
			"D6": "45296078.090012",
			"D7": 45296078090.012,
			"D8": "45296078090.012",
			"D9": 45296078090012,
			"D10": "45296078090012",
			"D11": "PT12H34M56.078090012S"
        }`,
		inVal: new(structDurationFormat),
		want: addr(structDurationFormat{
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
		}),
	}, {
		name:  Name("Duration/Format/Invalid"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"D":"0s"}`,
		inVal: addr(struct {
			D time.Duration `json:",format:invalid"`
		}{1}),
		want: addr(struct {
			D time.Duration `json:",format:invalid"`
		}{1}),
		wantErr: EU(errInvalidFormatFlag).withPos(`{"D":`, "/D").withType(0, timeDurationType),
	}, {
		/* TODO(https://go.dev/issue/71631): Re-enable this test case.
		name:  Name("Duration/Format/Legacy"),
		opts:  []Options{unsupported("FormatDurationAsNano")},
		inBuf: `{"D1":45296078090012,"D2":"12h34m56.078090012s"}`,
		inVal: new(structDurationFormat),
		want: addr(structDurationFormat{
			D1: 12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
			D2: 12*time.Hour + 34*time.Minute + 56*time.Second + 78*time.Millisecond + 90*time.Microsecond + 12*time.Nanosecond,
		}),
		}, { */
		/* TODO(https://go.dev/issue/71631): Re-enable this test case.
		name:  Name("Duration/MapKey"),
		inBuf: `{"1s":""}`,
		inVal: new(map[time.Duration]string),
		want:  addr(map[time.Duration]string{time.Second: ""}),
		}, { */
		name:  Name("Duration/MapKey/Legacy"),
		opts:  []Options{unsupported("FormatDurationAsNano")},
		inBuf: `{"1000000000":""}`,
		inVal: new(map[time.Duration]string),
		want:  addr(map[time.Duration]string{time.Second: ""}),
	}, {
		/* TODO(https://go.dev/issue/71631): Re-enable this test case.
		name:  Name("Duration/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `"1s"`,
		inVal: addr(time.Duration(0)),
		want:  addr(time.Second),
		}, { */
		name:  Name("Time/Zero"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"T1":"0001-01-01T00:00:00Z","T2":"01 Jan 01 00:00 UTC","T3":"0001-01-01","T4":"0001-01-01T00:00:00Z","T5":"0001-01-01T00:00:00Z"}`,
		inVal: new(struct {
			T1 time.Time
			T2 time.Time `json:",format:RFC822"`
			T3 time.Time `json:",format:'2006-01-02'"`
			T4 time.Time `json:",omitzero"`
			T5 time.Time `json:",omitempty"`
		}),
		want: addr(struct {
			T1 time.Time
			T2 time.Time `json:",format:RFC822"`
			T3 time.Time `json:",format:'2006-01-02'"`
			T4 time.Time `json:",omitzero"`
			T5 time.Time `json:",omitempty"`
		}{
			mustParseTime(time.RFC3339Nano, "0001-01-01T00:00:00Z"),
			mustParseTime(time.RFC822, "01 Jan 01 00:00 UTC"),
			mustParseTime("2006-01-02", "0001-01-01"),
			mustParseTime(time.RFC3339Nano, "0001-01-01T00:00:00Z"),
			mustParseTime(time.RFC3339Nano, "0001-01-01T00:00:00Z"),
		}),
	}, {
		name: Name("Time/Format"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{
			"T1": "1234-01-02T03:04:05.000000006Z",
			"T2": "Mon Jan  2 03:04:05 1234",
			"T3": "Mon Jan  2 03:04:05 UTC 1234",
			"T4": "Mon Jan 02 03:04:05 +0000 1234",
			"T5": "02 Jan 34 03:04 UTC",
			"T6": "02 Jan 34 03:04 +0000",
			"T7": "Monday, 02-Jan-34 03:04:05 UTC",
			"T8": "Mon, 02 Jan 1234 03:04:05 UTC",
			"T9": "Mon, 02 Jan 1234 03:04:05 +0000",
			"T10": "1234-01-02T03:04:05Z",
			"T11": "1234-01-02T03:04:05.000000006Z",
			"T12": "3:04AM",
			"T13": "Jan  2 03:04:05",
			"T14": "Jan  2 03:04:05.000",
			"T15": "Jan  2 03:04:05.000000",
			"T16": "Jan  2 03:04:05.000000006",
			"T17": "1234-01-02 03:04:05",
			"T18": "1234-01-02",
			"T19": "03:04:05",
			"T20": "1234-01-02",
			"T21": "\"weird\"1234",
			"T22": -23225777754.999999994,
			"T23": "-23225777754.999999994",
			"T24": -23225777754999.999994,
			"T25": "-23225777754999.999994",
			"T26": -23225777754999999.994,
			"T27": "-23225777754999999.994",
			"T28": -23225777754999999994,
			"T29": "-23225777754999999994"
		}`,
		inVal: new(structTimeFormat),
		want: addr(structTimeFormat{
			mustParseTime(time.RFC3339Nano, "1234-01-02T03:04:05.000000006Z"),
			mustParseTime(time.ANSIC, "Mon Jan  2 03:04:05 1234"),
			mustParseTime(time.UnixDate, "Mon Jan  2 03:04:05 UTC 1234"),
			mustParseTime(time.RubyDate, "Mon Jan 02 03:04:05 +0000 1234"),
			mustParseTime(time.RFC822, "02 Jan 34 03:04 UTC"),
			mustParseTime(time.RFC822Z, "02 Jan 34 03:04 +0000"),
			mustParseTime(time.RFC850, "Monday, 02-Jan-34 03:04:05 UTC"),
			mustParseTime(time.RFC1123, "Mon, 02 Jan 1234 03:04:05 UTC"),
			mustParseTime(time.RFC1123Z, "Mon, 02 Jan 1234 03:04:05 +0000"),
			mustParseTime(time.RFC3339, "1234-01-02T03:04:05Z"),
			mustParseTime(time.RFC3339Nano, "1234-01-02T03:04:05.000000006Z"),
			mustParseTime(time.Kitchen, "3:04AM"),
			mustParseTime(time.Stamp, "Jan  2 03:04:05"),
			mustParseTime(time.StampMilli, "Jan  2 03:04:05.000"),
			mustParseTime(time.StampMicro, "Jan  2 03:04:05.000000"),
			mustParseTime(time.StampNano, "Jan  2 03:04:05.000000006"),
			mustParseTime(time.DateTime, "1234-01-02 03:04:05"),
			mustParseTime(time.DateOnly, "1234-01-02"),
			mustParseTime(time.TimeOnly, "03:04:05"),
			mustParseTime("2006-01-02", "1234-01-02"),
			mustParseTime(`\"weird\"2006`, `\"weird\"1234`),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
		}),
	}, {
		name:    Name("Time/Format/String/Invalid"),
		opts:    []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf:   `{"T":"2023-01-01T00:00:00Z"}`,
		inVal:   new(structTimeFormatStringInvalid),
		want:    new(structTimeFormatStringInvalid),
		wantErr: EU(errInvalidStringTag).withPos(`{"T":`, "/T").withType(0, timeTimeType),
	}, {
		name: Name("Time/Format/UnixString/InvalidNumber"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{
			"T23": -23225777754.999999994,
			"T25": -23225777754999.999994,
			"T27": -23225777754999999.994,
			"T29": -23225777754999999994
		}`,
		inVal:   new(structTimeFormat),
		want:    new(structTimeFormat),
		wantErr: EU(nil).withPos(`{`+"\n\t\t\t"+`"T23": `, "/T23").withType('0', timeTimeType),
	}, {
		name: Name("Time/Format/UnixString/InvalidString"),
		opts: []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{
			"T22": "-23225777754.999999994",
			"T24": "-23225777754999.999994",
			"T26": "-23225777754999999.994",
			"T28": "-23225777754999999994"
		}`,
		inVal:   new(structTimeFormat),
		want:    new(structTimeFormat),
		wantErr: EU(nil).withPos(`{`+"\n\t\t\t"+`"T22": `, "/T22").withType('"', timeTimeType),
	}, {
		name:  Name("Time/Format/Null"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"T1":null,"T2":null,"T3":null,"T4":null,"T5":null,"T6":null,"T7":null,"T8":null,"T9":null,"T10":null,"T11":null,"T12":null,"T13":null,"T14":null,"T15":null,"T16":null,"T17":null,"T18":null,"T19":null,"T20":null,"T21":null,"T22":null,"T23":null,"T24":null,"T25":null,"T26":null,"T27":null,"T28":null,"T29":null}`,
		inVal: addr(structTimeFormat{
			mustParseTime(time.RFC3339Nano, "1234-01-02T03:04:05.000000006Z"),
			mustParseTime(time.ANSIC, "Mon Jan  2 03:04:05 1234"),
			mustParseTime(time.UnixDate, "Mon Jan  2 03:04:05 UTC 1234"),
			mustParseTime(time.RubyDate, "Mon Jan 02 03:04:05 +0000 1234"),
			mustParseTime(time.RFC822, "02 Jan 34 03:04 UTC"),
			mustParseTime(time.RFC822Z, "02 Jan 34 03:04 +0000"),
			mustParseTime(time.RFC850, "Monday, 02-Jan-34 03:04:05 UTC"),
			mustParseTime(time.RFC1123, "Mon, 02 Jan 1234 03:04:05 UTC"),
			mustParseTime(time.RFC1123Z, "Mon, 02 Jan 1234 03:04:05 +0000"),
			mustParseTime(time.RFC3339, "1234-01-02T03:04:05Z"),
			mustParseTime(time.RFC3339Nano, "1234-01-02T03:04:05.000000006Z"),
			mustParseTime(time.Kitchen, "3:04AM"),
			mustParseTime(time.Stamp, "Jan  2 03:04:05"),
			mustParseTime(time.StampMilli, "Jan  2 03:04:05.000"),
			mustParseTime(time.StampMicro, "Jan  2 03:04:05.000000"),
			mustParseTime(time.StampNano, "Jan  2 03:04:05.000000006"),
			mustParseTime(time.DateTime, "1234-01-02 03:04:05"),
			mustParseTime(time.DateOnly, "1234-01-02"),
			mustParseTime(time.TimeOnly, "03:04:05"),
			mustParseTime("2006-01-02", "1234-01-02"),
			mustParseTime(`\"weird\"2006`, `\"weird\"1234`),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
			time.Unix(-23225777755, 6).UTC(),
		}),
		want: new(structTimeFormat),
	}, {
		name:  Name("Time/RFC3339/Mismatch"),
		inBuf: `{"T":1234}`,
		inVal: new(struct {
			T time.Time
		}),
		wantErr: EU(nil).withPos(`{"T":`, "/T").withType('0', timeTimeType),
	}, {
		name:  Name("Time/RFC3339/ParseError"),
		inBuf: `{"T":"2021-09-29T12:44:52"}`,
		inVal: new(struct {
			T time.Time
		}),
		wantErr: EU(func() error {
			_, err := time.Parse(time.RFC3339, "2021-09-29T12:44:52")
			return err
		}()).withPos(`{"T":`, "/T").withType('"', timeTimeType),
	}, {
		name:  Name("Time/Format/Invalid"),
		opts:  []Options{unsupported("ExperimentalSupportFormatTag")},
		inBuf: `{"T":""}`,
		inVal: new(struct {
			T time.Time `json:",format:UndefinedConstant"`
		}),
		wantErr: EU(errors.New(`invalid format flag "UndefinedConstant"`)).withPos(`{"T":`, "/T").withType(0, timeTimeType),
	}, {
		name:    Name("Time/Format/SingleDigitHour"),
		inBuf:   `{"T":"2000-01-01T1:12:34Z"}`,
		inVal:   new(struct{ T time.Time }),
		wantErr: EU(newParseTimeError(time.RFC3339, "2000-01-01T1:12:34Z", "15", "1", "")).withPos(`{"T":`, "/T").withType('"', timeTimeType),
	}, {
		name:    Name("Time/Format/SubsecondComma"),
		inBuf:   `{"T":"2000-01-01T00:00:00,000Z"}`,
		inVal:   new(struct{ T time.Time }),
		wantErr: EU(newParseTimeError(time.RFC3339, "2000-01-01T00:00:00,000Z", ".", ",", "")).withPos(`{"T":`, "/T").withType('"', timeTimeType),
	}, {
		name:    Name("Time/Format/TimezoneHourOverflow"),
		inBuf:   `{"T":"2000-01-01T00:00:00+24:00"}`,
		inVal:   new(struct{ T time.Time }),
		wantErr: EU(newParseTimeError(time.RFC3339, "2000-01-01T00:00:00+24:00", "Z07:00", "+24:00", ": timezone hour out of range")).withPos(`{"T":`, "/T").withType('"', timeTimeType),
	}, {
		name:    Name("Time/Format/TimezoneMinuteOverflow"),
		inBuf:   `{"T":"2000-01-01T00:00:00+00:60"}`,
		inVal:   new(struct{ T time.Time }),
		wantErr: EU(newParseTimeError(time.RFC3339, "2000-01-01T00:00:00+00:60", "Z07:00", "+00:60", ": timezone minute out of range")).withPos(`{"T":`, "/T").withType('"', timeTimeType),
	}, {
		name:  Name("Time/Syntax/Invalid"),
		inBuf: `{"T":x}`,
		inVal: new(struct {
			T time.Time
		}),
		wantErr: newInvalidCharacterError("x", "at start of value", len64(`{"T":`), "/T"),
	}, {
		name:  Name("Time/IgnoreInvalidFormat"),
		opts:  []Options{invalidFormatOption},
		inBuf: `"2000-01-01T00:00:00Z"`,
		inVal: addr(time.Time{}),
		want:  addr(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
	}, {
		name: Name("Structs/Format/Bytes/Array"),
		opts: []Options{
			unsupported("ExperimentalSupportFormatTag"),
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *byte) error {
				if string(b) == "true" {
					*v = 1
				} else {
					*v = 0
				}
				return nil
			})),
		},
		inBuf: `{"Array":[false,true,false,true,false,true]}`,
		inVal: new(struct {
			Array []byte `json:",format:array"`
		}),
		want: addr(struct {
			Array []byte `json:",format:array"`
		}{
			Array: []byte{0, 1, 0, 1, 0, 1},
		}),
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt/UnmarshalFunc"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *any) error {
				var err error
				*v, err = strconv.ParseFloat(string(bytes.Trim(b, `"`)), 64)
				return err
			})),
		},
		inBuf: `{"D":"1.1","E":"2.2","F":"3.3"}`,
		inVal: new(structEmbedMapStringAny),
		want:  addr(structEmbedMapStringAny{X: jsonObject{"D": 1.1, "E": 2.2, "F": 3.3}}),
	}, {
		name: Name("Structs/EmbeddedFallback/MapStringInt/UnmarshalFunc"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *int) error {
				i, err := strconv.ParseInt(string(bytes.Trim(b, `"`)), 10, 64)
				if err != nil {
					return err
				}
				*v = int(i)
				return nil
			})),
		},
		inBuf: `{"zero": "0", "one": "1", "two": "2"}`,
		inVal: new(structEmbedMapStringInt),
		want: addr(structEmbedMapStringInt{
			X: map[string]int{"zero": 0, "one": 1, "two": 2},
		}),
	}, {
		name: Name("Structs/EmbeddedFallback/MapNamedStringInt/UnmarshalFunc"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *int) error {
				i, err := strconv.ParseInt(string(bytes.Trim(b, `"`)), 10, 64)
				if err != nil {
					return err
				}
				*v = int(i)
				return nil
			})),
		},
		inBuf: `{"zero": "0", "one": "1", "two": "2"}`,
		inVal: new(structEmbedMapNamedStringInt),
		want: addr(structEmbedMapNamedStringInt{
			X: map[namedString]int{"zero": 0, "one": 1, "two": 2},
		}),
	}, {
		name: Name("Interfaces/Any/UnmarshalFunc/Any"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *any) error {
				*v = "called"
				return nil
			})),
		},
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{"called"}),
	}, {
		name: Name("Interfaces/Any/UnmarshalFunc/Bool"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *bool) error {
				*v = string(b) != "true"
				return nil
			})),
		},
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{[]any{nil, true, false, "", 0.0, map[string]any{}, []any{}}}),
	}, {
		name: Name("Interfaces/Any/UnmarshalFunc/String"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *string) error {
				*v = "called"
				return nil
			})),
		},
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{[]any{nil, false, true, "called", 0.0, map[string]any{}, []any{}}}),
	}, {
		name: Name("Interfaces/Any/UnmarshalFunc/Float64"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *float64) error {
				*v = 3.14159
				return nil
			})),
		},
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{[]any{nil, false, true, "", 3.14159, map[string]any{}, []any{}}}),
	}, {
		name: Name("Interfaces/Any/UnmarshalFunc/MapStringAny"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *map[string]any) error {
				*v = map[string]any{"called": nil}
				return nil
			})),
		},
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{[]any{nil, false, true, "", 0.0, map[string]any{"called": nil}, []any{}}}),
	}, {
		name: Name("Interfaces/Any/UnmarshalFunc/SliceAny"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *[]any) error {
				*v = []any{"called"}
				return nil
			})),
		},
		inBuf: `{"X":[null,false,true,"",0,{},[]]}`,
		inVal: new(struct{ X any }),
		want:  addr(struct{ X any }{[]any{"called"}}),
	}, {
		name: Name("Functions/String/V1"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *string) error {
				if string(b) != `""` {
					return fmt.Errorf("got %s, want %s", b, `""`)
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr("called"),
	}, {
		name:  Name("Functions/String/Empty"),
		opts:  []Options{WithUnmarshalers(nil)},
		inBuf: `"hello"`,
		inVal: addr(""),
		want:  addr("hello"),
	}, {
		name: Name("Functions/NamedString/V1/NoMatch"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *namedString) error {
				panic("should not be called")
			})),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr(""),
	}, {
		name: Name("Functions/NamedString/V1/Match"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *namedString) error {
				if string(b) != `""` {
					return fmt.Errorf("got %s, want %s", b, `""`)
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `""`,
		inVal: addr(namedString("")),
		want:  addr(namedString("called")),
	}, {
		name: Name("Functions/String/Empty1/NoMatch"),
		opts: []Options{
			WithUnmarshalers(new(Unmarshalers)),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr(""),
	}, {
		name: Name("Functions/String/Empty2/NoMatch"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers()),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr(""),
	}, {
		name: Name("Functions/String/V1/DirectError"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func([]byte, *string) error {
				return errSomeError
			})),
		},
		inBuf:   `""`,
		inVal:   addr(""),
		want:    addr(""),
		wantErr: EU(errSomeError).withType('"', reflect.PointerTo(stringType)),
	}, {
		name: Name("Functions/String/V1/SkipError"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func([]byte, *string) error {
				return errors.ErrUnsupported
			})),
		},
		inBuf:   `""`,
		inVal:   addr(""),
		want:    addr(""),
		wantErr: EU(wrapErrUnsupported(errors.ErrUnsupported, "unmarshal function of type func([]byte, T) error")).withType('"', reflect.PointerTo(stringType)),
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V1"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *nocaseString) error {
				if string(b) != `"hello"` {
					return fmt.Errorf("got %s, want %s", b, `"hello"`)
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[nocaseString]string{}),
		want:  addr(map[nocaseString]string{"called": "world"}),
	}, {
		name: Name("Functions/Map/Key/TextMarshaler/V1"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v encoding.TextMarshaler) error {
				if string(b) != `"hello"` {
					return fmt.Errorf("got %s, want %s", b, `"hello"`)
				}
				*v.(*nocaseString) = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[nocaseString]string{}),
		want:  addr(map[nocaseString]string{"called": "world"}),
	}, {
		name: Name("Functions/Map/Value/NoCaseString/V1"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *nocaseString) error {
				if string(b) != `"world"` {
					return fmt.Errorf("got %s, want %s", b, `"world"`)
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[string]nocaseString{}),
		want:  addr(map[string]nocaseString{"hello": "called"}),
	}, {
		name: Name("Functions/Map/Value/TextMarshaler/V1"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v encoding.TextMarshaler) error {
				if string(b) != `"world"` {
					return fmt.Errorf("got %s, want %s", b, `"world"`)
				}
				*v.(*nocaseString) = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[string]nocaseString{}),
		want:  addr(map[string]nocaseString{"hello": "called"}),
	}, {
		name: Name("Functions/Slice/Elem"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *string) error {
				*v = strings.Trim(strings.ToUpper(string(b)), `"`)
				return nil
			})),
		},
		inBuf: `["hello","World"]`,
		inVal: addr([]string{}),
		want:  addr([]string{"HELLO", "WORLD"}),
	}, {
		name: Name("Functions/Array/Elem"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFunc(func(b []byte, v *string) error {
				*v = strings.Trim(strings.ToUpper(string(b)), `"`)
				return nil
			})),
		},
		inBuf: `["hello","World"]`,
		inVal: addr([2]string{}),
		want:  addr([2]string{"HELLO", "WORLD"}),
	}, {
		name: Name("Functions/Precedence/NestedFirst"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				JoinUnmarshalers(
					UnmarshalFunc(func(b []byte, v *string) error {
						if string(b) != `"called"` {
							return fmt.Errorf("got %s, want %s", b, `"called"`)
						}
						*v = "called"
						return nil
					}),
				),
				UnmarshalFunc(func([]byte, *string) error {
					panic("should not be called")
				}),
			)),
		},
		inBuf: `"called"`,
		inVal: addr(""),
		want:  addr("called"),
	}, {
		name: Name("Functions/Precedence/NestedLast"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				UnmarshalFunc(func(b []byte, v *string) error {
					if string(b) != `"called"` {
						return fmt.Errorf("got %s, want %s", b, `"called"`)
					}
					*v = "called"
					return nil
				}),
				JoinUnmarshalers(
					UnmarshalFunc(func([]byte, *string) error {
						panic("should not be called")
					}),
				),
			)),
		},
		inBuf: `"called"`,
		inVal: addr(""),
		want:  addr("called"),
	}, {
		name: Name("Functions/String/V2"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				switch b, err := dec.ReadValue(); {
				case err != nil:
					return err
				case string(b) != `""`:
					return fmt.Errorf("got %s, want %s", b, `""`)
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr("called"),
	}, {
		name: Name("Functions/NamedString/V2/NoMatch"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *namedString) error {
				panic("should not be called")
			})),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr(""),
	}, {
		name: Name("Functions/NamedString/V2/Match"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *namedString) error {
				switch t, err := dec.ReadToken(); {
				case err != nil:
					return err
				case t.String() != ``:
					return fmt.Errorf("got %q, want %q", t, ``)
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `""`,
		inVal: addr(namedString("")),
		want:  addr(namedString("called")),
	}, {
		name: Name("Functions/String/V2/DirectError"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				return errSomeError
			})),
		},
		inBuf:   `""`,
		inVal:   addr(""),
		want:    addr(""),
		wantErr: EU(errSomeError).withType(0, reflect.PointerTo(stringType)),
	}, {
		name: Name("Functions/String/V2/TooFew"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				return nil
			})),
		},
		inBuf:   `""`,
		inVal:   addr(""),
		want:    addr(""),
		wantErr: EU(errNonSingularValue).withType(0, reflect.PointerTo(stringType)),
	}, {
		name: Name("Functions/String/V2/TooMany"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				if _, err := dec.ReadValue(); err != nil {
					return err
				}
				if _, err := dec.ReadValue(); err != nil {
					return err
				}
				return nil
			})),
		},
		inBuf:   `{"X":["",""]}`,
		inVal:   addr(struct{ X []string }{}),
		want:    addr(struct{ X []string }{[]string{""}}),
		wantErr: EU(errNonSingularValue).withPos(`{"X":["",`, "/X").withType(0, reflect.PointerTo(stringType)),
	}, {
		name: Name("Functions/String/V2/Skipped"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				return errors.ErrUnsupported
			})),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr(""),
	}, {
		name: Name("Functions/String/V2/ProcessBeforeSkip"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				if _, err := dec.ReadValue(); err != nil {
					return err
				}
				return errors.ErrUnsupported
			})),
		},
		inBuf:   `""`,
		inVal:   addr(""),
		want:    addr(""),
		wantErr: EU(errUnsupportedMutation).withType(0, reflect.PointerTo(stringType)),
	}, {
		name: Name("Functions/String/V2/WrappedUnsupported"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				return fmt.Errorf("wrap: %w", errors.ErrUnsupported)
			})),
		},
		inBuf: `""`,
		inVal: addr(""),
		want:  addr(""),
	}, {
		name: Name("Functions/Map/Key/NoCaseString/V2"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *nocaseString) error {
				switch t, err := dec.ReadToken(); {
				case err != nil:
					return err
				case t.String() != "hello":
					return fmt.Errorf("got %q, want %q", t, "hello")
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[nocaseString]string{}),
		want:  addr(map[nocaseString]string{"called": "world"}),
	}, {
		name: Name("Functions/Map/Key/TextMarshaler/V2"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v encoding.TextMarshaler) error {
				switch b, err := dec.ReadValue(); {
				case err != nil:
					return err
				case string(b) != `"hello"`:
					return fmt.Errorf("got %s, want %s", b, `"hello"`)
				}
				*v.(*nocaseString) = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[nocaseString]string{}),
		want:  addr(map[nocaseString]string{"called": "world"}),
	}, {
		name: Name("Functions/Map/Key/String/V1/DuplicateName"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				if _, err := dec.ReadValue(); err != nil {
					return err
				}
				xdDepth, xdLength := decoderPosition(dec)
				*v = fmt.Sprintf("%d-%d", xdDepth, xdLength)
				return nil
			})),
		},
		inBuf:   `{"name":"value","name":"value"}`,
		inVal:   addr(map[string]string{}),
		want:    addr(map[string]string{"1-1": "1-2"}),
		wantErr: newDuplicateNameError("", []byte(`"name"`), len64(`{"name":"value",`)),
	}, {
		name: Name("Functions/Map/Value/NoCaseString/V2"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *nocaseString) error {
				switch t, err := dec.ReadToken(); {
				case err != nil:
					return err
				case t.String() != "world":
					return fmt.Errorf("got %q, want %q", t, "world")
				}
				*v = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[string]nocaseString{}),
		want:  addr(map[string]nocaseString{"hello": "called"}),
	}, {
		name: Name("Functions/Map/Value/TextMarshaler/V2"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v encoding.TextMarshaler) error {
				switch b, err := dec.ReadValue(); {
				case err != nil:
					return err
				case string(b) != `"world"`:
					return fmt.Errorf("got %s, want %s", b, `"world"`)
				}
				*v.(*nocaseString) = "called"
				return nil
			})),
		},
		inBuf: `{"hello":"world"}`,
		inVal: addr(map[string]nocaseString{}),
		want:  addr(map[string]nocaseString{"hello": "called"}),
	}, {
		name: Name("Funtions/Struct/Fields"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				UnmarshalFunc(func(b []byte, v *bool) error {
					if string(b) != `"called1"` {
						return fmt.Errorf("got %s, want %s", b, `"called1"`)
					}
					*v = true
					return nil
				}),
				UnmarshalFunc(func(b []byte, v *string) error {
					if string(b) != `"called2"` {
						return fmt.Errorf("got %s, want %s", b, `"called2"`)
					}
					*v = "called2"
					return nil
				}),
				UnmarshalFromFunc(func(dec *jsontext.Decoder, v *[]byte) error {
					switch t, err := dec.ReadToken(); {
					case err != nil:
						return err
					case t.String() != "called3":
						return fmt.Errorf("got %q, want %q", t, "called3")
					}
					*v = []byte("called3")
					return nil
				}),
				UnmarshalFromFunc(func(dec *jsontext.Decoder, v *int64) error {
					switch b, err := dec.ReadValue(); {
					case err != nil:
						return err
					case string(b) != `"called4"`:
						return fmt.Errorf("got %s, want %s", b, `"called4"`)
					}
					*v = 123
					return nil
				}),
			)),
		},
		inBuf: `{"Bool":"called1","String":"called2","Bytes":"called3","Int":"called4","Uint":456,"Float":789}`,
		inVal: addr(structScalars{}),
		want:  addr(structScalars{Bool: true, String: "called2", Bytes: []byte("called3"), Int: 123, Uint: 456, Float: 789}),
	}, {
		name: Name("Functions/Struct/Embedded"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				UnmarshalFunc(func([]byte, *structEmbeddedL1) error {
					panic("should not be called")
				}),
				UnmarshalFromFunc(func(dec *jsontext.Decoder, v *StructEmbed2) error {
					panic("should not be called")
				}),
			)),
		},
		inBuf: `{"E":"E3","F":"F3","G":"G3","A":"A1","B":"B1","D":"D2"}`,
		inVal: new(structEmbedded),
		want: addr(structEmbedded{
			X: structEmbeddedL1{
				X:            &structEmbeddedL2{A: "A1", B: "B1" /* C: "C1" */},
				StructEmbed1: StructEmbed1{ /* C: "C2" */ D: "D2" /* E: "E2" */},
			},
			StructEmbed2: &StructEmbed2{E: "E3", F: "F3", G: "G3"},
		}),
	}, {
		name: Name("Functions/Pointer/Nil"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				t, err := dec.ReadToken()
				*v = strings.ToUpper(t.String())
				return err
			})),
		},
		inBuf: `{"X":"hello"}`,
		inVal: addr(struct{ X *string }{nil}),
		want:  addr(struct{ X *string }{addr("HELLO")}),
	}, {
		name: Name("Functions/Pointer/NonNil"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
				t, err := dec.ReadToken()
				*v = strings.ToUpper(t.String())
				return err
			})),
		},
		inBuf: `{"X":"hello"}`,
		inVal: addr(struct{ X *string }{addr("")}),
		want:  addr(struct{ X *string }{addr("HELLO")}),
	}, {
		name: Name("Functions/Interface/Nil"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v fmt.Stringer) error {
				panic("should not be called")
			})),
		},
		inBuf:   `{"X":"hello"}`,
		inVal:   addr(struct{ X fmt.Stringer }{nil}),
		want:    addr(struct{ X fmt.Stringer }{nil}),
		wantErr: EU(errNilInterface).withPos(`{"X":`, "/X").withType(0, T[fmt.Stringer]()),
	}, {
		name: Name("Functions/Interface/NetIP"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *fmt.Stringer) error {
				*v = net.IP{}
				return errors.ErrUnsupported
			})),
		},
		inBuf: `{"X":"1.1.1.1"}`,
		inVal: addr(struct{ X fmt.Stringer }{nil}),
		want:  addr(struct{ X fmt.Stringer }{net.IPv4(1, 1, 1, 1)}),
	}, {
		name: Name("Functions/Interface/NewPointerNetIP"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *fmt.Stringer) error {
				*v = new(net.IP)
				return errors.ErrUnsupported
			})),
		},
		inBuf: `{"X":"1.1.1.1"}`,
		inVal: addr(struct{ X fmt.Stringer }{nil}),
		want:  addr(struct{ X fmt.Stringer }{addr(net.IPv4(1, 1, 1, 1))}),
	}, {
		name: Name("Functions/Interface/NilPointerNetIP"),
		opts: []Options{
			WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, v *fmt.Stringer) error {
				*v = (*net.IP)(nil)
				return errors.ErrUnsupported
			})),
		},
		inBuf: `{"X":"1.1.1.1"}`,
		inVal: addr(struct{ X fmt.Stringer }{nil}),
		want:  addr(struct{ X fmt.Stringer }{addr(net.IPv4(1, 1, 1, 1))}),
	}, {
		name: Name("Functions/Interface/NilPointerNetIP/Override"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				UnmarshalFromFunc(func(dec *jsontext.Decoder, v *fmt.Stringer) error {
					*v = (*net.IP)(nil)
					return errors.ErrUnsupported
				}),
				UnmarshalFunc(func(b []byte, v *net.IP) error {
					b = bytes.ReplaceAll(b, []byte(`1`), []byte(`8`))
					return v.UnmarshalText(bytes.Trim(b, `"`))
				}),
			)),
		},
		inBuf: `{"X":"1.1.1.1"}`,
		inVal: addr(struct{ X fmt.Stringer }{nil}),
		want:  addr(struct{ X fmt.Stringer }{addr(net.IPv4(8, 8, 8, 8))}),
	}, {
		name: Name("Functions/Precedence/V1First"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				UnmarshalFunc(func(b []byte, v *string) error {
					if string(b) != `"called"` {
						return fmt.Errorf("got %s, want %s", b, `"called"`)
					}
					*v = "called"
					return nil
				}),
				UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
					panic("should not be called")
				}),
			)),
		},
		inBuf: `"called"`,
		inVal: addr(""),
		want:  addr("called"),
	}, {
		name: Name("Functions/Precedence/V2First"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
					switch t, err := dec.ReadToken(); {
					case err != nil:
						return err
					case t.String() != "called":
						return fmt.Errorf("got %q, want %q", t, "called")
					}
					*v = "called"
					return nil
				}),
				UnmarshalFunc(func([]byte, *string) error {
					panic("should not be called")
				}),
			)),
		},
		inBuf: `"called"`,
		inVal: addr(""),
		want:  addr("called"),
	}, {
		name: Name("Functions/Precedence/V2Skipped"),
		opts: []Options{
			WithUnmarshalers(JoinUnmarshalers(
				UnmarshalFromFunc(func(dec *jsontext.Decoder, v *string) error {
					return errors.ErrUnsupported
				}),
				UnmarshalFunc(func(b []byte, v *string) error {
					if string(b) != `"called"` {
						return fmt.Errorf("got %s, want %s", b, `"called"`)
					}
					*v = "called"
					return nil
				}),
			)),
		},
		inBuf: `"called"`,
		inVal: addr(""),
		want:  addr("called"),
	}}
}

func TestUnmarshal(t *testing.T) {
	runUnmarshalTests(t, "Unmarshal", func(in string, out any, opts ...Options) error {
		return Unmarshal([]byte(in), out, opts...)
	})
}

// TestUnmarshalRead runs the cases of TestUnmarshal by UnmarshalRead, from readers of one byte, of a few and of the
// whole input at a time, which make it decode a value element by element or whole.
func TestUnmarshalRead(t *testing.T) {
	for _, n := range []int{1, 7, 0} {
		t.Run(fmt.Sprintf("reads-%d", n), func(t *testing.T) {
			runUnmarshalTests(t, "UnmarshalRead", func(in string, out any, opts ...Options) error {
				r := io.Reader(strings.NewReader(in))
				if n > 0 {
					r = &chunkReader{r: r, n: n}
				}
				return UnmarshalRead(r, out, opts...)
			})
		})
	}
}

// chunkReader reads r n bytes at a time at most.
type chunkReader struct {
	r io.Reader
	n int
}

func (r *chunkReader) Read(p []byte) (int, error) { return r.r.Read(p[:min(len(p), r.n)]) }

// runUnmarshalTests runs the cases of TestUnmarshal by unmarshal.
func runUnmarshalTests(t *testing.T, what string, unmarshal func(in string, out any, opts ...Options) error) {
	for _, tt := range unmarshalTests() {
		if tt.skip {
			continue
		}
		t.Run(tt.name.Name, func(t *testing.T) {
			skipUnsupported(t, tt.opts)
			got := tt.inVal
			gotErr := unmarshal(tt.inBuf, got, tt.opts...)
			// the value is compared only without an error: what is decoded before an error is not specified.
			if tt.wantErr == nil && !reflect.DeepEqual(got, tt.want) && tt.want != nil {
				t.Errorf("%s: %s output mismatch:\ngot  %v\nwant %v", tt.name.Where, what, got, tt.want)
			}
			if !equalError(gotErr, tt.wantErr) {
				t.Errorf("%s: %s error mismatch:\ngot  %v\nwant %v", tt.name.Where, what, gotErr, tt.wantErr)
			}
		})
	}
}

type unmarshalerEOF struct{}

func (unmarshalerEOF) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return io.EOF // should be wrapped and converted by Unmarshal to io.ErrUnexpectedEOF
}

// TestUnmarshalEOF verifies that io.EOF is converted to io.ErrUnexpectedEOF by Unmarshal and UnmarshalRead.
func TestUnmarshalEOF(t *testing.T) {
	opts := WithUnmarshalers(UnmarshalFromFunc(func(dec *jsontext.Decoder, _ *struct{}) error {
		return io.EOF // should be wrapped and converted by Unmarshal to io.ErrUnexpectedEOF
	}))

	for _, in := range []string{"", "[", "[null", "[null]"} {
		for _, newOut := range []func() any{
			func() any { return new(unmarshalerEOF) },
			func() any { return new([]unmarshalerEOF) },
			func() any { return new(struct{}) },
			func() any { return new([]struct{}) },
		} {
			wantErr := io.ErrUnexpectedEOF
			if gotErr := Unmarshal([]byte(in), newOut(), opts); !errors.Is(gotErr, wantErr) {
				t.Errorf("Unmarshal = %v, want %v", gotErr, wantErr)
			}
			if gotErr := UnmarshalRead(strings.NewReader(in), newOut(), opts); !errors.Is(gotErr, wantErr) {
				t.Errorf("Unmarshal = %v, want %v", gotErr, wantErr)
			}
		}
	}
}
