package json

import (
	"strconv"

	"github.com/goccy/go-json/internal/textcoder"
	"github.com/goccy/go-json/jsontext"
)

// The place of an error of marshaling is found from the output written before it, which the encoder returns with
// the error: the JSON pointer of the value which was to be written next. Nothing is kept for it while encoding.

// nextPointer returns the JSON pointer of the value which is written after out, the output of the encoder so far:
// valid JSON up to the place where the next value or object name starts, after a comma which ends the previous
// member or element, if any.
//
// It is the member whose name ends out, the next element of an array, or the object itself, whose next member
// name is not written yet.
func nextPointer(out []byte) jsontext.Pointer {
	type level struct {
		array bool
		index int    // the index of the next element of an array
		name  []byte // the quoted name of the member whose value is next, if the name has been written
	}
	var stack []level
	for i := 0; i < len(out); i++ {
		switch c := out[i]; c {
		case '{', '[':
			stack = append(stack, level{array: c == '['})
		case '}', ']':
			stack = stack[:len(stack)-1]
		case ',':
			if len(stack) > 0 {
				top := &stack[len(stack)-1]
				if top.array {
					top.index++
				} else {
					top.name = nil
				}
			}
		case '"':
			start := i
			for i++; i < len(out) && out[i] != '"'; i++ {
				if out[i] == '\\' {
					i++
				}
			}
			if len(stack) > 0 {
				top := &stack[len(stack)-1]
				if !top.array && top.name == nil {
					// a member name, which a colon follows: its value is next.
					top.name = out[start : i+1]
				}
			}
		}
	}
	// a member whose value is written is ended by a comma, which removes its name: a name left is of the member
	// whose value is next, or, at the end of out, of the member whose value was being written.
	var p jsontext.Pointer
	for _, l := range stack {
		switch {
		case l.array:
			p += "/" + jsontext.Pointer(strconv.Itoa(l.index))
		case l.name != nil:
			// the names are written by the encoder, as valid JSON strings.
			name, _ := jsontext.AppendUnquote(nil, l.name)
			p = p.AppendToken(string(name))
		}
	}
	return p
}

// levelsOf returns the levels of the output out, valid JSON up to the place of a value or of an object name,
// without the delimiter before it: the top level, and the objects and arrays which are open, appended to levels.
func levelsOf(out []byte, levels []textcoder.Level) []textcoder.Level {
	levels = append(levels, textcoder.Level{})
	for i := 0; i < len(out); i++ {
		switch c := out[i]; c {
		case '{', '[':
			// the value is counted when it ends.
			levels = append(levels, textcoder.Level{Object: c == '{'})
		case '}', ']':
			levels = levels[:len(levels)-1]
			levels[len(levels)-1].Count++
		case '"':
			start := i
			for i++; i < len(out) && out[i] != '"'; i++ {
				if out[i] == '\\' {
					i++
				}
			}
			top := &levels[len(levels)-1]
			if top.Object && top.Count%2 == 0 {
				name, _ := jsontext.AppendUnquote(nil, out[start:i+1])
				top.Names = append(top.Names, name)
			}
			top.Count++
		case ',', ':', ' ', '\t', '\n', '\r':
		default:
			// a number or a literal, up to the next delimiter.
			for i+1 < len(out) && !isDelim(out[i+1]) {
				i++
			}
			levels[len(levels)-1].Count++
		}
	}
	return levels
}

// pointerOf returns the JSON pointer of the place in levels, which levelsOf returned: of the last value of the
// innermost level if where is -1, of its next value if where is +1, and of the level itself if where is 0, or of
// the value of the last name of an object whose value is next.
func pointerOf(levels []textcoder.Level, where int8) jsontext.Pointer {
	var p jsontext.Pointer
	for i := 1; i < len(levels); i++ {
		l := &levels[i]
		innermost := i == len(levels)-1
		switch {
		case l.Object:
			// the last name, of the value which is open, next or last.
			if len(l.Names) > 0 && (l.Count%2 == 1 || !innermost || where == -1) {
				p = p.AppendToken(string(l.Names[len(l.Names)-1]))
			}
		case !innermost:
			p += "/" + jsontext.Pointer(strconv.FormatInt(l.Count, 10))
		case where == -1 && l.Count > 0:
			p += "/" + jsontext.Pointer(strconv.FormatInt(l.Count-1, 10))
		case where == +1:
			p += "/" + jsontext.Pointer(strconv.FormatInt(l.Count, 10))
		}
	}
	return p
}

func isDelim(c byte) bool {
	switch c {
	case ',', ':', '}', ']', ' ', '\t', '\n', '\r':
		return true
	}
	return false
}
