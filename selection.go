package json

import (
	"context"
	"fmt"
	"strconv"

	"github.com/goccy/go-json/internal/encoder"
)

// Selection is the fields of a value which are written, in the order they are written, as a selection set of
// GraphQL selects the fields of a response. It is made by Select or ParseSelection, and given to an encoding by
// WithSelection.
//
//   - A field is found by its key in the JSON object of the value: the name in the json tag, or the name of the
//     Go field. A field is written with its alias if it has one.
//   - A field selected without fields is written whole. A field selected with fields is filtered by them.
//   - The selection of a slice, an array, a map or a pointer is the one of each of its elements, and the
//     selection of an interface value is the one of the value it holds. The keys of a map are not selected.
//   - A name which is not a field of the value is ignored, as the selection of a value which is not an object is.
//   - The fields of a struct embedded by a pointer are written together, where the first of them is selected:
//     they are written only if the pointer is not nil. The fields of a struct embedded as a value are written
//     where they are selected.
//   - A field whose type implements MarshalJSON(context.Context) is given the selection of its value by the
//     context ( SelectionFromContext ), and MarshalContext with that context applies it.
//
// A Selection is not changed after it is made, and is safe for concurrent use. It keeps the code which the
// encoder compiles for the types it selects the fields of, so a selection which is used again should be kept:
// making a new one for every encoding compiles the code every time.
type Selection = encoder.Selection

// SelectedField is a field of a Selection: the name of the field, the key it is written with, the selection of
// its value, and the value given by SelectionField.With.
type SelectedField = encoder.SelectedField

// SelectionField is a field of a selection, which Field or FieldOf makes.
type SelectionField struct {
	name   string
	alias  string
	value  any
	fields []SelectionField
	// sel is the selection of the value given by FieldOf.
	sel *Selection
}

// Field returns the field of the name, selected with the fields if they are given, or whole.
func Field(name string, fields ...SelectionField) SelectionField {
	return SelectionField{name: name, fields: fields}
}

// FieldOf returns the field of the name whose value is selected by sel, or whole if sel is nil. A selection of no
// field, Select(), writes an object without fields. A selection given to many fields is shared by them, with
// the code compiled for it.
func FieldOf(name string, sel *Selection) SelectionField {
	return SelectionField{name: name, sel: sel}
}

// As returns the field written with the key alias instead of its name.
func (f SelectionField) As(alias string) SelectionField {
	f.alias = alias
	return f
}

// With returns the field which has the value, such as the arguments of a field of GraphQL. MarshalJSON(context.Context)
// of the value of the field gets it by SelectedFieldFromContext: it is how a resolver knows its arguments.
func (f SelectionField) With(value any) SelectionField {
	f.value = value
	return f
}

func (f SelectionField) key() string {
	if f.alias != "" {
		return f.alias
	}
	return f.name
}

// isWhole is whether the value of the field is written whole.
func (f SelectionField) isWhole() bool {
	return len(f.fields) == 0 && f.sel == nil
}

// Select returns the selection of the fields. A selection of no field writes an object without fields.
//
// The fields of the same key are merged, as GraphQL merges them: the fields they select are selected, and the
// value is the first one given. It is an error that the fields of the same key are different fields, or that one
// is selected whole and one with fields.
func Select(fields ...SelectionField) (*Selection, error) {
	return buildSelection(fields)
}

func buildSelection(fields []SelectionField) (*Selection, error) {
	sel := &Selection{Fields: make([]*encoder.SelectedField, 0, len(fields))}
	byKey := map[string]int{}
	// the fields of each key, which are merged after all of them are known.
	var merged [][]SelectionField
	for _, field := range fields {
		if field.name == "" {
			return nil, fmt.Errorf("json: selection: a field has no name")
		}
		key := field.key()
		if i, exists := byKey[key]; exists {
			first := sel.Fields[i]
			if first.Name != field.name {
				return nil, fmt.Errorf("json: selection: the key %q is of the fields %q and %q", key, first.Name, field.name)
			}
			if merged[i][0].isWhole() != field.isWhole() {
				return nil, fmt.Errorf("json: selection: the field %q is selected both whole and with fields", key)
			}
			merged[i] = append(merged[i], field)
			if first.Value == nil {
				first.Value = field.value
			}
			continue
		}
		byKey[key] = len(sel.Fields)
		merged = append(merged, []SelectionField{field})
		sel.Fields = append(sel.Fields, &encoder.SelectedField{Name: field.name, Key: key, Value: field.value})
	}
	for i, field := range sel.Fields {
		sub, err := subSelection(merged[i])
		if err != nil {
			return nil, err
		}
		field.Sub = sub
	}
	return sel, nil
}

// subSelection returns the selection of the value of the fields of a key, or nil if it is written whole.
func subSelection(fields []SelectionField) (*Selection, error) {
	if fields[0].isWhole() {
		return nil, nil
	}
	if len(fields) == 1 && fields[0].sel != nil {
		return fields[0].sel, nil
	}
	var subFields []SelectionField
	for _, field := range fields {
		subFields = append(subFields, field.fields...)
		if field.sel != nil {
			subFields = append(subFields, fieldsOf(field.sel)...)
		}
	}
	return buildSelection(subFields)
}

// fieldsOf returns the fields of the selection, to be merged with others.
func fieldsOf(sel *Selection) []SelectionField {
	fields := make([]SelectionField, 0, len(sel.Fields))
	for _, f := range sel.Fields {
		field := SelectionField{name: f.Name, value: f.Value, sel: f.Sub}
		if f.Key != f.Name {
			field.alias = f.Key
		}
		fields = append(fields, field)
	}
	return fields
}

// ParseSelection returns the selection written in the syntax of a selection set of GraphQL: the fields,
// separated by white space or commas, each of which is a name, optionally preceded by an alias and a colon, and
// optionally followed by its fields in braces. The whole may be in braces. A name which is not a name of
// GraphQL is written as a JSON string. A comment starts with # and ends at the end of the line. Empty braces
// are the selection of no field, which GraphQL doesn't have: it is what String writes for it.
//
//	id displayName: name friends { name } "user-id"
//
// The arguments, the fragments and the directives of GraphQL are not a part of it: they are what a GraphQL
// server resolves before it knows the fields to write.
func ParseSelection(s string) (*Selection, error) {
	p := &selectionParser{src: s}
	p.skipIgnored()
	var fields []SelectionField
	var err error
	if p.peek() == '{' {
		p.pos++
		fields, err = p.parseFields('}')
	} else {
		fields, err = p.parseFields(0)
	}
	if err != nil {
		return nil, err
	}
	p.skipIgnored()
	if p.pos < len(p.src) {
		return nil, p.errorf("unexpected %q", p.src[p.pos])
	}
	return buildSelection(fields)
}

type selectionParser struct {
	src string
	pos int
}

func (p *selectionParser) errorf(format string, args ...any) error {
	return fmt.Errorf("json: selection: %s at offset %d", fmt.Sprintf(format, args...), p.pos)
}

func (p *selectionParser) peek() byte {
	if p.pos < len(p.src) {
		return p.src[p.pos]
	}
	return 0
}

// skipIgnored skips the white space, the commas and the comments, which GraphQL ignores between tokens.
func (p *selectionParser) skipIgnored() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r', ',':
			p.pos++
		case '#':
			for p.pos < len(p.src) && p.src[p.pos] != '\n' && p.src[p.pos] != '\r' {
				p.pos++
			}
		default:
			return
		}
	}
}

// parseFields parses the fields up to the closing brace, which is consumed, or up to the end if end is 0.
func (p *selectionParser) parseFields(end byte) ([]SelectionField, error) {
	var fields []SelectionField
	for {
		p.skipIgnored()
		c := p.peek()
		if end != 0 && c == end {
			p.pos++
			return fields, nil
		}
		if c == 0 {
			if end != 0 {
				return nil, p.errorf("missing '}'")
			}
			if len(fields) == 0 {
				return nil, p.errorf("no field")
			}
			return fields, nil
		}
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
}

func (p *selectionParser) parseField() (SelectionField, error) {
	name, err := p.parseName()
	if err != nil {
		return SelectionField{}, err
	}
	field := SelectionField{name: name}
	p.skipIgnored()
	if p.peek() == ':' {
		p.pos++
		p.skipIgnored()
		name, err := p.parseName()
		if err != nil {
			return SelectionField{}, err
		}
		field = SelectionField{name: name, alias: field.name}
		p.skipIgnored()
	}
	if p.peek() == '{' {
		p.pos++
		fields, err := p.parseFields('}')
		if err != nil {
			return SelectionField{}, err
		}
		field.fields = fields
		if len(fields) == 0 {
			field.sel = &Selection{}
		}
	}
	return field, nil
}

// parseName parses a name of GraphQL, or a JSON string for a name which is not one.
func (p *selectionParser) parseName() (string, error) {
	start := p.pos
	if p.peek() == '"' {
		p.pos++
		for p.pos < len(p.src) {
			switch p.src[p.pos] {
			case '\\':
				p.pos += 2
				continue
			case '"':
				p.pos++
				var name string
				if err := Unmarshal([]byte(p.src[start:p.pos]), &name); err != nil {
					p.pos = start
					return "", p.errorf("invalid string %s", p.src[start:p.pos])
				}
				if name == "" {
					p.pos = start
					return "", p.errorf("empty name")
				}
				return name, nil
			}
			p.pos++
		}
		p.pos = start
		return "", p.errorf("unterminated string")
	}
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' && p.pos > start {
			p.pos++
			continue
		}
		break
	}
	if p.pos == start {
		if p.pos < len(p.src) {
			return "", p.errorf("unexpected %s", strconv.QuoteRune(rune(p.src[p.pos])))
		}
		return "", p.errorf("missing name")
	}
	return p.src[start:p.pos], nil
}

// WithSelection makes the encoding write only the fields of the value which the selection selects.
// A nil selection writes the whole value.
func WithSelection(sel *Selection) EncodeOptionFunc {
	return func(opt *EncodeOption) {
		opt.Selection = sel
		if sel == nil {
			opt.Flag &^= encoder.SelectionOption
		} else {
			opt.Flag |= encoder.SelectionOption
		}
	}
}

// SelectionFromContext returns the selection which the context has, or nil if it has none or the value is
// written whole. MarshalJSON(context.Context) of a value encoded with a selection is given the selection of the
// value by its context.
func SelectionFromContext(ctx context.Context) *Selection {
	return encoder.SelectionFromContext(ctx)
}

// SelectedFieldFromContext returns the selected field whose value MarshalJSON(context.Context) is given the
// context for, or nil if the value is not the value of a selected field of a struct.
func SelectedFieldFromContext(ctx context.Context) *SelectedField {
	return encoder.SelectedFieldFromContext(ctx)
}

// setSelectionOfContext makes the encoding apply the selection of the context, which is given to
// MarshalJSON(context.Context), so that MarshalContext called with it writes what is selected.
func setSelectionOfContext(opt *EncodeOption, ctx context.Context) {
	if sel := encoder.SelectionFromContext(ctx); sel != nil {
		opt.Selection = sel
		opt.Flag |= encoder.SelectionOption
	}
}
