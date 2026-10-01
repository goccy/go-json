package gql

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/goccy/go-json"
)

// Selected is the field which a resolver resolves.
type Selected struct {
	Name  string
	Alias string
	// Args are the arguments of the field, with the defaults of the schema and the variables of the request.
	Args map[string]any
	// Selection is what the query selects of the value of the field, or nil for a scalar or a union: a resolver
	// may look ahead at it.
	Selection *json.Selection
}

// Int returns the argument of an Int, or nil if it is null.
func (s *Selected) Int(name string) *int {
	var n int
	switch v := s.Args[name].(type) {
	case int:
		n = v
	case int64:
		n = int(v)
	case float64:
		n = int(v)
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return nil
		}
		n = int(i)
	default:
		return nil
	}
	return &n
}

// String returns the argument of a String or an ID, or "" if it is null.
func (s *Selected) String(name string) string {
	v, _ := s.Args[name].(string)
	return v
}

// Resolver resolves a field in the resolve phase, concurrently with the other fields of its level: the fields at
// the same depth of the response, which are resolved before the fields of their values. A data loader gathers
// the fetches of the resolvers of a level, as it does with the resolvers of gqlgen.
//
// The value is written with the selection of the field. A value which has fields to resolve is to be a pointer
// or a slice, whose values are the ones written.
type Resolver interface {
	Resolve(ctx context.Context, field *Selected) (any, error)
}

// BatchResolver resolves a field of all the values of a level at once, which the resolve phase knows before it
// resolves them: ResolveBatch of one of them is called with all of them, and returns their values in the same
// order, and nil errors or one error for each. It fetches them by one fetch without the wait of a data loader.
type BatchResolver[R any] interface {
	ResolveBatch(ctx context.Context, field *Selected, batch []R) ([]any, []error)
}

// Field is a field of a type which R resolves: R implements Resolver, and BatchResolver[R] to be resolved for
// all the values of a level at once. The JSON key of the Go field is the name of the GraphQL field.
type Field[R any] struct {
	R R
}

// MarshalJSON writes the value which the resolve phase resolved for the field.
func (f *Field[R]) MarshalJSON(ctx context.Context) ([]byte, error) {
	return writeResolved(ctx, unsafe.Pointer(f))
}

// resolvable is a Field of any resolver.
type resolvable interface {
	resolverType() reflect.Type
	canBatch() bool
	resolveOne(ctx context.Context, field *Selected) (any, error)
	resolveBatch(ctx context.Context, field *Selected, batch []resolvable) ([]any, []error)
}

var resolvableType = reflect.TypeFor[resolvable]()

func (f *Field[R]) resolverType() reflect.Type {
	return reflect.TypeFor[R]()
}

func (f *Field[R]) canBatch() bool {
	_, ok := any(f.R).(BatchResolver[R])
	return ok
}

func (f *Field[R]) resolveOne(ctx context.Context, field *Selected) (any, error) {
	r, ok := any(f.R).(Resolver)
	if !ok {
		return nil, fmt.Errorf("%T resolves the field %s by batches only", f.R, field.Name)
	}
	return r.Resolve(ctx, field)
}

func (f *Field[R]) resolveBatch(ctx context.Context, field *Selected, batch []resolvable) ([]any, []error) {
	rs := make([]R, len(batch))
	for i, r := range batch {
		rs[i] = r.(*Field[R]).R
	}
	return any(f.R).(BatchResolver[R]).ResolveBatch(ctx, field, rs)
}

// resultKey is what a resolved value is found by when it is written: the address of the field and the selected
// field it is written for.
type resultKey struct {
	ptr   unsafe.Pointer
	field *json.SelectedField
}

type result struct {
	value any
	err   error
}

// pathNode is an element of the path of a value, which makes the path of an error when one happens.
type pathNode struct {
	parent *pathNode
	elem   ast.PathElement
}

func (n *pathNode) path() ast.Path {
	var path ast.Path
	for ; n != nil; n = n.parent {
		path = append(path, n.elem)
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// item is a value whose fields are to be resolved, with the selection of the value.
type item struct {
	v   reflect.Value
	sel *json.Selection
	// byType are the selections by the type of the value, for a value of a union or an interface.
	byType map[string]*json.Selection
	path   *pathNode
}

// task is a field to resolve.
type task struct {
	res   resolvable
	ptr   unsafe.Pointer
	field *json.SelectedField
	path  *pathNode
}

// resolveAll resolves the fields of the value level by level: the fields of a level are all found, resolved
// concurrently, and their values are the items of the next level.
func (e *Engine) resolveAll(ctx context.Context, req *request, root item) {
	frontier := []item{root}
	for len(frontier) > 0 {
		var tasks []task
		for _, it := range frontier {
			e.collect(it, &tasks)
		}
		tasks = req.unresolved(tasks)
		results := e.run(ctx, req, tasks)
		frontier = frontier[:0]
		for i, t := range tasks {
			r := results[i]
			req.results[resultKey{ptr: t.ptr, field: t.field}] = r
			if r.err != nil {
				req.errors = append(req.errors, fieldError(t, r.err))
				continue
			}
			byType := byTypeOf(t.field)
			if r.value == nil || (t.field.Sub == nil && byType == nil) {
				continue
			}
			frontier = append(frontier, item{v: reflect.ValueOf(r.value), sel: t.field.Sub, byType: byType, path: t.path})
		}
	}
}

// fieldError returns the error of the field of the task, with its path and its place in the query.
func fieldError(t task, err error) *gqlerror.Error {
	e := gqlerror.WrapPath(t.path.path(), err)
	if info, ok := t.field.Value.(*fieldInfo); ok && info.field.Position != nil {
		e.Locations = []gqlerror.Location{{Line: info.field.Position.Line, Column: info.field.Position.Column}}
	}
	return e
}

// unresolved returns the tasks whose fields are not resolved yet: a value reached twice by the same selected
// field is resolved once.
func (r *request) unresolved(tasks []task) []task {
	seen := make(map[resultKey]bool, len(tasks))
	kept := tasks[:0]
	for _, t := range tasks {
		key := resultKey{ptr: t.ptr, field: t.field}
		if _, done := r.results[key]; done || seen[key] {
			continue
		}
		seen[key] = true
		kept = append(kept, t)
	}
	return kept
}

// run resolves the tasks of a level concurrently: the ones of a batch resolver and of the same selected field
// by one call, and each of the others by a call of its own.
func (e *Engine) run(ctx context.Context, req *request, tasks []task) []result {
	results := make([]result, len(tasks))
	type group struct {
		field   *json.SelectedField
		indices []int
	}
	type groupKey struct {
		typ   reflect.Type
		field *json.SelectedField
	}
	var units []func()
	groups := map[groupKey]*group{}
	var ordered []*group
	for i, t := range tasks {
		if !e.batch || !t.res.canBatch() {
			selected := req.selected(t.field)
			units = append(units, func() {
				v, err := safeResolveOne(ctx, t.res, selected)
				results[i] = result{value: v, err: err}
			})
			continue
		}
		key := groupKey{typ: t.res.resolverType(), field: t.field}
		g := groups[key]
		if g == nil {
			g = &group{field: t.field}
			groups[key] = g
			ordered = append(ordered, g)
		}
		g.indices = append(g.indices, i)
	}
	for _, g := range ordered {
		selected := req.selected(g.field)
		batch := make([]resolvable, len(g.indices))
		for j, i := range g.indices {
			batch[j] = tasks[i].res
		}
		units = append(units, func() {
			values, errs := safeResolveBatch(ctx, batch, selected)
			for j, i := range g.indices {
				var r result
				if j < len(values) {
					r.value = values[j]
				}
				if errs != nil {
					r.err = errs[min(j, len(errs)-1)]
				}
				results[i] = r
			}
		})
	}
	switch len(units) {
	case 0:
	case 1:
		units[0]()
	default:
		var wg sync.WaitGroup
		for _, unit := range units[1:] {
			wg.Go(unit)
		}
		units[0]()
		wg.Wait()
	}
	return results
}

func safeResolveOne(ctx context.Context, res resolvable, field *Selected) (v any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("the resolver of %s panicked: %v", field.Name, r)
		}
	}()
	return res.resolveOne(ctx, field)
}

func safeResolveBatch(ctx context.Context, batch []resolvable, field *Selected) (values []any, errs []error) {
	defer func() {
		if r := recover(); r != nil {
			values, errs = nil, []error{fmt.Errorf("the resolver of %s panicked: %v", field.Name, r)}
		}
	}()
	values, errs = batch[0].resolveBatch(ctx, field, batch)
	if len(values) != len(batch) && len(errs) == 0 {
		errs = []error{fmt.Errorf("the batch resolver of %s returned %d values for %d fields", field.Name, len(values), len(batch))}
	}
	return values, errs
}

// collect appends the fields to resolve of the value, which the selection selects: the resolvers of the value
// and of the values in it which are not resolved by a resolver themselves.
func (e *Engine) collect(it item, tasks *[]task) {
	v := it.v
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if it.byType == nil && !e.plans.mayResolve(v.Type().Elem(), it.sel) {
			return
		}
		for i := 0; i < v.Len(); i++ {
			e.collect(item{v: v.Index(i), sel: it.sel, byType: it.byType, path: &pathNode{parent: it.path, elem: ast.PathIndex(i)}}, tasks)
		}
	case reflect.Struct:
		sel := it.sel
		if it.byType != nil {
			sel = it.byType[e.plans.typename(v)]
			if sel == nil {
				return
			}
		}
		for _, st := range e.plans.planOf(v.Type(), sel).steps {
			f := v.FieldByIndex(st.index)
			path := &pathNode{parent: it.path, elem: ast.PathName(st.field.Key)}
			if !st.resolver {
				e.collect(item{v: f, sel: st.field.Sub, byType: byTypeOf(st.field), path: path}, tasks)
				continue
			}
			if !f.CanAddr() {
				// the field is in a copy, which the encoding doesn't write: see Resolver.
				continue
			}
			addr := f.Addr()
			*tasks = append(*tasks, task{res: addr.Interface().(resolvable), ptr: addr.UnsafePointer(), field: st.field, path: path})
		}
	}
}

func byTypeOf(field *json.SelectedField) map[string]*json.Selection {
	if info, ok := field.Value.(*fieldInfo); ok {
		return info.byType
	}
	return nil
}

// planner keeps the plans of the types for the selections, which are made once: the resolve phase finds the
// fields to resolve without looking at the fields of a type which has none.
type planner struct {
	plans     sync.Map // planKey -> *plan
	resolves  sync.Map // planKey -> bool
	typenames sync.Map // reflect.Type -> []int
}

type planKey struct {
	typ reflect.Type
	sel *json.Selection
}

type plan struct {
	steps []step
}

// step is a field of a struct which the selection selects: a field to resolve, or a field whose value has fields
// to resolve.
type step struct {
	index    []int
	field    *json.SelectedField
	resolver bool
}

func (p *planner) planOf(t reflect.Type, sel *json.Selection) *plan {
	key := planKey{typ: t, sel: sel}
	if cached, ok := p.plans.Load(key); ok {
		return cached.(*plan)
	}
	pl := &plan{}
	if sel != nil {
		keys := jsonKeys(t)
		for _, field := range sel.Fields {
			index, ok := keys[field.Name]
			if !ok {
				continue
			}
			ft := t.FieldByIndex(index).Type
			switch {
			case reflect.PointerTo(ft).Implements(resolvableType):
				pl.steps = append(pl.steps, step{index: index, field: field, resolver: true})
			case byTypeOf(field) != nil || p.mayResolve(ft, field.Sub):
				pl.steps = append(pl.steps, step{index: index, field: field})
			}
		}
	}
	cached, _ := p.plans.LoadOrStore(key, pl)
	return cached.(*plan)
}

// mayResolve is whether a value of the type may have fields to resolve which the selection selects.
func (p *planner) mayResolve(t reflect.Type, sel *json.Selection) bool {
	if sel == nil {
		return false
	}
	key := planKey{typ: t, sel: sel}
	if cached, ok := p.resolves.Load(key); ok {
		return cached.(bool)
	}
	var may bool
	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		may = p.mayResolve(t.Elem(), sel)
	case reflect.Interface:
		may = true
	case reflect.Struct:
		may = len(p.planOf(t, sel).steps) > 0
	}
	p.resolves.Store(key, may)
	return may
}

// typename returns the name of the GraphQL type of the value: its field of the JSON key __typename.
func (p *planner) typename(v reflect.Value) string {
	t := v.Type()
	index, ok := p.typenames.Load(t)
	if !ok {
		index, _ = p.typenames.LoadOrStore(t, jsonKeys(t)["__typename"])
	}
	if index.([]int) == nil {
		return ""
	}
	return v.FieldByIndex(index.([]int)).String()
}

// jsonKeys returns the indices of the fields of the struct by their JSON keys, as go-json names them.
func jsonKeys(t reflect.Type) map[string][]int {
	keys := map[string][]int{}
	for _, f := range reflect.VisibleFields(t) {
		if !f.IsExported() || f.Anonymous && f.Tag.Get("json") == "" && f.Type.Kind() == reflect.Struct {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		if _, exists := keys[name]; !exists {
			keys[name] = f.Index
		}
	}
	return keys
}

// writeResolved writes the value resolved for the field at ptr, or null if it has not been resolved or failed.
func writeResolved(ctx context.Context, ptr unsafe.Pointer) ([]byte, error) {
	req, _ := ctx.Value(requestKey{}).(*request)
	field := json.SelectedFieldFromContext(ctx)
	if req == nil || field == nil {
		return []byte("null"), nil
	}
	r, ok := req.results[resultKey{ptr: ptr, field: field}]
	if !ok || r.err != nil || r.value == nil {
		return []byte("null"), nil
	}
	if byType := byTypeOf(field); byType != nil {
		return req.engine.writeByType(ctx, reflect.ValueOf(r.value), byType)
	}
	return json.MarshalContext(ctx, r.value)
}

// writeByType writes a value of a union or an interface, or a list of them, each with the selection of its type.
func (e *Engine) writeByType(ctx context.Context, v reflect.Value, byType map[string]*json.Selection) ([]byte, error) {
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return []byte("null"), nil
		}
		if v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Struct {
			break
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		b := []byte{'['}
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				b = append(b, ',')
			}
			out, err := e.writeByType(ctx, v.Index(i), byType)
			if err != nil {
				return nil, err
			}
			b = append(b, out...)
		}
		return append(b, ']'), nil
	}
	elem := v
	if elem.Kind() == reflect.Pointer {
		elem = elem.Elem()
	}
	sel := byType[e.plans.typename(elem)]
	if sel == nil {
		sel = noFields
	}
	return json.MarshalContext(ctx, v.Interface(), json.WithSelection(sel))
}
