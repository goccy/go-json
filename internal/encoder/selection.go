package encoder

import (
	"context"
	"strconv"
	"strings"
	"sync"
)

// Selection is the fields of a value which are written, in the order they are written: what a selection set
// of GraphQL is. It is made once and is not changed after, so it is shared by the goroutines.
//
// A field is found by its key in the JSON object of the value. The selection of a list, an array, a map or a
// pointer is the one of each of its elements, and the selection of an interface value is the one of the value
// it holds. A name which is not a field of the value is ignored, as the selection of a value which is not an
// object is.
type Selection struct {
	Fields []*SelectedField

	// codeSets are the opcodes of the types filtered by the selection, by the opcodes of the type which are
	// not filtered. The opcodes live as long as the selection, which is what bounds them.
	mu       sync.RWMutex
	codeSets map[*OpcodeSet]*OpcodeSet
}

// SelectedField is a field of a selection.
type SelectedField struct {
	// Name is the key of the field in the JSON object of the value.
	Name string
	// Key is the key the field is written with: Name, or the alias of the field.
	Key string
	// Sub is the selection of the value of the field, or nil if the whole value is written.
	Sub *Selection
	// Value is what the user gave to the field, such as the arguments of a field of GraphQL, for
	// MarshalJSON(context.Context) of the value of the field ( SelectedFieldFromContext ).
	Value any
}

// selectedOf returns the selected field which only tells the selection, for a value which is not a field: the
// value of a list, of an interface value, or the whole value.
func selectedOf(sel *Selection) *SelectedField {
	if sel == nil {
		return nil
	}
	return &SelectedField{Sub: sel}
}

// Selection returns the selection of the value of the field, or nil if the whole value is written.
func (f *SelectedField) Selection() *Selection {
	if f == nil {
		return nil
	}
	return f.Sub
}

// String returns the selection in the syntax of a selection set of GraphQL, with the names which are not
// names of GraphQL quoted. Two selections of the same fields return the same string.
func (s *Selection) String() string {
	if len(s.Fields) == 0 {
		return "{}"
	}
	var b strings.Builder
	s.writeTo(&b)
	return b.String()
}

func (s *Selection) writeTo(b *strings.Builder) {
	for i, field := range s.Fields {
		if i > 0 {
			b.WriteByte(' ')
		}
		if field.Key != field.Name {
			writeSelectionName(b, field.Key)
			b.WriteString(": ")
		}
		writeSelectionName(b, field.Name)
		switch {
		case field.Sub == nil:
		case len(field.Sub.Fields) == 0:
			b.WriteString(" {}")
		default:
			b.WriteString(" { ")
			field.Sub.writeTo(b)
			b.WriteString(" }")
		}
	}
}

func writeSelectionName(b *strings.Builder, name string) {
	if IsSelectionName(name) {
		b.WriteString(name)
		return
	}
	b.WriteString(strconv.Quote(name))
}

// IsSelectionName is whether the name is a name of GraphQL, which is written without quotes.
func IsSelectionName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c == '_', 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
		case '0' <= c && c <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

// CodeSet returns the opcodes of the type of codeSet filtered by the selection, compiling them the first time.
// codeSet is the opcodes of the type which are not filtered.
func (s *Selection) CodeSet(codeSet *OpcodeSet, optimizeFieldOrder bool) (*OpcodeSet, error) {
	s.mu.RLock()
	selected := s.codeSets[codeSet]
	s.mu.RUnlock()
	if selected != nil {
		return selected, nil
	}
	compiler := newCompiler(optimizeFieldOrder)
	compiler.isFiltered = true
	selected, err := compiler.codeToOpcodeSet(codeSet.Type, codeSet.Code.Filter(s))
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if cached := s.codeSets[codeSet]; cached != nil {
		return cached, nil
	}
	if s.codeSets == nil {
		s.codeSets = map[*OpcodeSet]*OpcodeSet{}
	}
	s.codeSets[codeSet] = selected
	return selected, nil
}

// SelectCodeSet returns the opcodes of the type of codeSet filtered by the selection of the call: it is called
// only when the call has one ( SelectionOption ), which the callers check themselves, so that the encoding of
// a call without a selection costs no call.
func (c *RuntimeContext) SelectCodeSet(codeSet *OpcodeSet) (*OpcodeSet, error) {
	return c.Option.Selection.CodeSet(codeSet, c.Option.Flag&OptimizeFieldOrderOption != 0)
}

type selectedFieldKey struct{}

// SelectionFromContext returns the selection which the context has, or nil.
func SelectionFromContext(ctx context.Context) *Selection {
	field, _ := ctx.Value(selectedFieldKey{}).(*SelectedField)
	return field.Selection()
}

// SelectedFieldFromContext returns the selected field which the context has, or nil if it has none or the value
// is not the value of a field.
func SelectedFieldFromContext(ctx context.Context) *SelectedField {
	field, _ := ctx.Value(selectedFieldKey{}).(*SelectedField)
	if field == nil || field.Name == "" {
		return nil
	}
	return field
}

// ContextWithSelectedField returns the context which has the selected field. A nil field is kept as well: it
// tells that the whole value is written, and hides the selection of the parent context.
func ContextWithSelectedField(ctx context.Context, field *SelectedField) context.Context {
	return context.WithValue(ctx, selectedFieldKey{}, field)
}

// ContextWithSelection returns the context which has the selection of a value which is not a field.
func ContextWithSelection(ctx context.Context, sel *Selection) context.Context {
	return ContextWithSelectedField(ctx, selectedOf(sel))
}
