package encoder

import (
	"unsafe"

	"github.com/goccy/go-json/internal/errors"
	"github.com/goccy/go-json/internal/jsonstring"
)

// The names of a map whose keys may have the same name, of the v2 semantics, which a function, a method or the
// kind of its key writes. The names of a map which is not sorted are checked as they are written ( see
// appendMapKeyName ); the names of a sorted map are checked after the entries are sorted, in their order.

// checkSortedNames is the function after a map whose keys may have the same name ( see OpAfterValue ): it
// reports the first name of the map, which ends the output b, which a name before it has, if the entries are
// sorted and the names may not be the same.
func checkSortedNames(ctx *RuntimeContext, b []byte, _ unsafe.Pointer) ([]byte, error) {
	if ctx.Option.Flag&(AllowDuplicateNamesOption|UnorderedMapOption) != 0 || b[len(b)-2] != '}' {
		return b, nil
	}
	start := lastValueStart(b)
	seen := map[string]struct{}{}
	for _, pos := range memberNames(b[start : len(b)-1]) {
		end := stringEnd(b, start+pos)
		name := string(b[start+pos : end+1])
		if _, ok := seen[name]; ok {
			unquoted := jsonstring.AppendUnescaped(nil, []byte(name[1:len(name)-1]))
			return b, &errors.TextError{Err: errors.ErrDuplicateName, Name: string(unquoted), Out: b[:start+pos]}
		}
		seen[name] = struct{}{}
	}
	return b, nil
}

// lastValueStart returns where the value which ends the output b, before its comma, starts. The output is
// compact and valid.
func lastValueStart(b []byte) int {
	i := len(b) - 2 // the last byte of the value
	switch b[i] {
	case '"':
		return stringStart(b, i)
	case '}', ']':
		depth := 0
		for ; i >= 0; i-- {
			switch b[i] {
			case '"':
				i = stringStart(b, i)
			case '}', ']':
				depth++
			case '{', '[':
				depth--
				if depth == 0 {
					return i
				}
			}
		}
		return 0
	}
	// a number or a literal, after a delimiter.
	for i > 0 {
		switch b[i-1] {
		case ',', ':', '[', '{':
			return i
		}
		i--
	}
	return i
}

// stringStart returns where the string whose closing quote is at end starts: the quote before it which no odd
// number of backslashes precedes, which would escape it.
func stringStart(b []byte, end int) int {
	for i := end - 1; i >= 0; i-- {
		if b[i] != '"' {
			continue
		}
		n := 0
		for j := i - 1; j >= 0 && b[j] == '\\'; j-- {
			n++
		}
		if n%2 == 0 {
			return i
		}
	}
	return 0
}

// stringEnd returns where the string which starts at start ends: its closing quote.
func stringEnd(b []byte, start int) int {
	for i := start + 1; i < len(b); i++ {
		switch b[i] {
		case '\\':
			i++
		case '"':
			return i
		}
	}
	return len(b) - 1
}

// memberNames returns where the names of the members of the object obj start: the object is compact and valid.
func memberNames(obj []byte) []int {
	var names []int
	depth := 0
	expectName := false
	for i := 0; i < len(obj); i++ {
		switch obj[i] {
		case '{', '[':
			depth++
			expectName = obj[i] == '{' && depth == 1
		case '}', ']':
			depth--
		case ',':
			expectName = depth == 1
		case '"':
			if expectName {
				names = append(names, i)
				expectName = false
			}
			i = stringEnd(obj, i)
		}
	}
	return names
}
