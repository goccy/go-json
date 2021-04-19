package encoder

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-json/internal/errors"
)

func Indent(buf *bytes.Buffer, src []byte, prefix, indentStr string) error {
	if len(src) == 0 {
		return errors.ErrUnexpectedEndOfJSON("", 0)
	}
	buf.Grow(len(src))
	dst := buf.Bytes()
	newSrc := make([]byte, len(src)+1) // append nul byte to the end
	copy(newSrc, src)
	dst, err := doIndent(dst, newSrc, prefix, indentStr, false)
	if err != nil {
		return err
	}
	if _, err := buf.Write(dst); err != nil {
		return err
	}
	return nil
}

func doIndent(dst, src []byte, prefix, indentStr string, escape bool) ([]byte, error) {
	buf, cursor, err := indentValue(dst, src, 0, 0, []byte(prefix), []byte(indentStr), escape)
	if err != nil {
		return nil, err
	}
	if err := validateEndBuf(src, cursor); err != nil {
		return nil, err
	}
	return buf, nil
}

func indentValue(
	dst []byte,
	src []byte,
	indentNum int,
	cursor int64,
	prefix []byte,
	indentBytes []byte,
	escape bool) ([]byte, int64, error) {
	for {
		switch src[cursor] {
		case ' ', '\t', '\n', '\r':
			cursor++
			continue
		case '{':
			return indentObject(dst, src, indentNum, cursor, prefix, indentBytes, escape)
		case '}':
			return nil, 0, errors.ErrSyntax("unexpected character '}'", cursor)
		case '[':
			return indentArray(dst, src, indentNum, cursor, prefix, indentBytes, escape)
		case ']':
			return nil, 0, errors.ErrSyntax("unexpected character ']'", cursor)
		case '"':
			return compactString(dst, src, cursor, escape)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return compactNumber(dst, src, cursor)
		case 't':
			return compactTrue(dst, src, cursor)
		case 'f':
			return compactFalse(dst, src, cursor)
		case 'n':
			return compactNull(dst, src, cursor)
		default:
			return nil, 0, errors.ErrSyntax(fmt.Sprintf("unexpected character '%c'", src[cursor]), cursor)
		}
	}
}

func indentObject(
	dst []byte,
	src []byte,
	indentNum int,
	cursor int64,
	prefix []byte,
	indentBytes []byte,
	escape bool) ([]byte, int64, error) {
	if src[cursor] == '{' {
		dst = append(dst, '{')
	} else {
		return nil, 0, errors.ErrExpected("expected { character for object value", cursor)
	}
	cursor = skipWhiteSpace(src, cursor+1)
	if src[cursor] == '}' {
		dst = append(dst, '}')
		return dst, cursor + 1, nil
	}
	indentNum++
	var err error
	for {
		dst = append(append(append(dst, '\n'), prefix...), bytes.Repeat(indentBytes, indentNum)...)
		cursor = skipWhiteSpace(src, cursor)
		dst, cursor, err = compactString(dst, src, cursor, escape)
		if err != nil {
			return nil, 0, err
		}
		cursor = skipWhiteSpace(src, cursor)
		if src[cursor] != ':' {
			return nil, 0, errors.ErrSyntax(
				fmt.Sprintf("invalid character '%c' after object key", src[cursor]),
				cursor+1,
			)
		}
		dst = append(dst, ':', ' ')
		dst, cursor, err = indentValue(dst, src, indentNum, cursor+1, prefix, indentBytes, escape)
		if err != nil {
			return nil, 0, err
		}
		cursor = skipWhiteSpace(src, cursor)
		switch src[cursor] {
		case '}':
			dst = append(append(append(dst, '\n'), prefix...), bytes.Repeat(indentBytes, indentNum-1)...)
			dst = append(dst, '}')
			cursor++
			return dst, cursor, nil
		case ',':
			dst = append(dst, ',')
		default:
			return nil, 0, errors.ErrSyntax(
				fmt.Sprintf("invalid character '%c' after object key:value pair", src[cursor]),
				cursor+1,
			)
		}
		cursor++
	}
}

func indentArray(
	dst []byte,
	src []byte,
	indentNum int,
	cursor int64,
	prefix []byte,
	indentBytes []byte,
	escape bool) ([]byte, int64, error) {
	if src[cursor] == '[' {
		dst = append(dst, '[')
	} else {
		return nil, 0, errors.ErrExpected("expected [ character for array value", cursor)
	}
	cursor = skipWhiteSpace(src, cursor+1)
	if src[cursor] == ']' {
		dst = append(dst, ']')
		return dst, cursor + 1, nil
	}
	indentNum++
	var err error
	for {
		dst = append(append(append(dst, '\n'), prefix...), bytes.Repeat(indentBytes, indentNum)...)
		dst, cursor, err = indentValue(dst, src, indentNum, cursor, prefix, indentBytes, escape)
		if err != nil {
			return nil, 0, err
		}
		cursor = skipWhiteSpace(src, cursor)
		switch src[cursor] {
		case ']':
			dst = append(append(append(dst, '\n'), prefix...), bytes.Repeat(indentBytes, indentNum-1)...)
			dst = append(dst, ']')
			cursor++
			return dst, cursor, nil
		case ',':
			dst = append(dst, ',')
		default:
			return nil, 0, errors.ErrSyntax(
				fmt.Sprintf("invalid character '%c' after array value", src[cursor]),
				cursor+1,
			)
		}
		cursor++
	}
}
