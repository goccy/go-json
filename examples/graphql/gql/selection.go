package gql

import (
	"fmt"
	"sort"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/goccy/go-json"
)

// selectionBuilder makes the json.Selection of a selection set of GraphQL, collecting its fields as the
// CollectFields algorithm of GraphQL does: the fields of the fragments whose type condition applies are merged
// in, the fields skipped by @skip and @include are left out, and the fields of the same response key are merged.
type selectionBuilder struct {
	schema *ast.Schema
	// variables are the variables of the request.
	variables map[string]any
	// usedVariables are the variables which the directives that decide the fields depend on: the selection
	// is the one of the requests whose values of them are the same.
	usedVariables map[string]bool
}

// fieldInfo is what a selected field is given for its resolver ( json.SelectionField.With ).
type fieldInfo struct {
	field *ast.Field
	// args are the arguments of the field if they don't depend on the variables.
	args map[string]any
	// byType are the selections of the value of a field of a union or an interface, by the name of the type
	// of the value: a fragment selects the fields of a type.
	byType map[string]*json.Selection
}

func (b *selectionBuilder) selection(sets []ast.SelectionSet, typeName string) (*json.Selection, error) {
	fields, err := b.fields(sets, typeName)
	if err != nil {
		return nil, err
	}
	return json.Select(fields...)
}

func (b *selectionBuilder) fields(sets []ast.SelectionSet, typeName string) ([]json.SelectionField, error) {
	var keys []string
	grouped := map[string][]*ast.Field{}
	visited := map[string]bool{}
	var collect func(set ast.SelectionSet)
	collect = func(set ast.SelectionSet) {
		for _, sel := range set {
			switch sel := sel.(type) {
			case *ast.Field:
				if !b.included(sel.Directives) {
					continue
				}
				key := sel.Alias
				if key == "" {
					key = sel.Name
				}
				if _, exists := grouped[key]; !exists {
					keys = append(keys, key)
				}
				grouped[key] = append(grouped[key], sel)
			case *ast.InlineFragment:
				if b.included(sel.Directives) && b.applies(sel.TypeCondition, typeName) {
					collect(sel.SelectionSet)
				}
			case *ast.FragmentSpread:
				if !b.included(sel.Directives) || visited[sel.Name] || !b.applies(sel.Definition.TypeCondition, typeName) {
					continue
				}
				visited[sel.Name] = true
				collect(sel.Definition.SelectionSet)
			}
		}
	}
	for _, set := range sets {
		collect(set)
	}
	fields := make([]json.SelectionField, 0, len(keys))
	for _, key := range keys {
		same := grouped[key]
		f := same[0]
		if f.Name == "__typename" {
			fields = append(fields, json.Field("__typename").As(key))
			continue
		}
		if f.Definition == nil {
			return nil, fmt.Errorf("unknown field %s on %s", f.Name, typeName)
		}
		var subSets []ast.SelectionSet
		for _, g := range same {
			subSets = append(subSets, g.SelectionSet)
		}
		info := b.info(f)
		named := b.schema.Types[f.Definition.Type.Name()]
		var field json.SelectionField
		switch named.Kind {
		case ast.Object:
			sub, err := b.selection(subSets, named.Name)
			if err != nil {
				return nil, err
			}
			field = json.FieldOf(f.Name, sub)
		case ast.Union, ast.Interface:
			// the value is written by its resolver, with the selection of its type.
			info.byType = map[string]*json.Selection{}
			for _, t := range b.schema.GetPossibleTypes(named) {
				sub, err := b.selection(subSets, t.Name)
				if err != nil {
					return nil, err
				}
				info.byType[t.Name] = sub
			}
			field = json.Field(f.Name)
		default:
			field = json.Field(f.Name)
		}
		field = field.As(key).With(info)
		fields = append(fields, field)
	}
	return fields, nil
}

// info returns what the resolver of the field is given: the field of the query, for the place of an error, and
// its arguments if they don't depend on the variables.
func (b *selectionBuilder) info(f *ast.Field) *fieldInfo {
	info := &fieldInfo{field: f}
	if !argumentsUseVariables(f.Arguments) {
		info.args = f.ArgumentMap(nil)
	}
	return info
}

// applies is whether the fragment of the type condition applies to a value of the type.
func (b *selectionBuilder) applies(condition, typeName string) bool {
	if condition == "" || condition == typeName {
		return true
	}
	for _, t := range b.schema.GetPossibleTypes(b.schema.Types[condition]) {
		if t.Name == typeName {
			return true
		}
	}
	return false
}

// included is whether @skip and @include of the directives let the selection be.
func (b *selectionBuilder) included(directives ast.DirectiveList) bool {
	if d := directives.ForName("skip"); d != nil && b.directiveIf(d) {
		return false
	}
	if d := directives.ForName("include"); d != nil && !b.directiveIf(d) {
		return false
	}
	return true
}

func (b *selectionBuilder) directiveIf(d *ast.Directive) bool {
	for _, arg := range d.Arguments {
		collectVariables(arg.Value, func(name string) {
			if b.usedVariables == nil {
				b.usedVariables = map[string]bool{}
			}
			b.usedVariables[name] = true
		})
	}
	v, _ := d.ArgumentMap(b.variables)["if"].(bool)
	return v
}

// directiveVariables returns the names of the variables which the directives that decide the fields depend on,
// sorted.
func (b *selectionBuilder) directiveVariables() []string {
	names := make([]string, 0, len(b.usedVariables))
	for name := range b.usedVariables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func argumentsUseVariables(args ast.ArgumentList) bool {
	uses := false
	for _, arg := range args {
		collectVariables(arg.Value, func(string) { uses = true })
	}
	return uses
}

// collectVariables calls f with the name of each variable in the value.
func collectVariables(v *ast.Value, f func(name string)) {
	if v == nil {
		return
	}
	if v.Kind == ast.Variable {
		f(v.Raw)
		return
	}
	for _, child := range v.Children {
		collectVariables(child.Value, f)
	}
}
