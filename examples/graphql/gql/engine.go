// Package gql is a GraphQL engine built on go-json, a proof of concept of a GraphQL server which is not
// generated: the types of the schema are Go structs whose JSON keys are the names of the GraphQL fields, and a
// field with a resolver is a Field.
//
// A request is served in two phases:
//
//  1. The resolve phase resolves the fields level by level: the fields at the same depth of the response are
//     found from the selection of the query, and resolved concurrently, a batch resolver for all of them at once.
//     A data loader gathers the fetches of the resolvers of a level, as with gqlgen.
//  2. The write phase writes the response by one call of json.MarshalContext with the selection of the query,
//     which writes the resolved values where their fields are.
//
// The query is parsed and validated by gqlparser, as gqlgen does. The selection of a query is made once and
// kept with the parsed query, and go-json keeps the code compiled for it in the selection.
package gql

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sync"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/validator"

	"github.com/goccy/go-json"
)

// Options are how an engine serves the requests.
type Options struct {
	// DisableBatch makes the batch resolvers resolve each field by Resolve, as the resolvers which are not.
	DisableBatch bool
	// RequestContext returns the context of a request, such as one with the data loaders of the request.
	RequestContext func(ctx context.Context) context.Context
}

// Engine serves the GraphQL requests on the root value of the queries.
type Engine struct {
	schema         *ast.Schema
	root           any
	batch          bool
	requestContext func(ctx context.Context) context.Context
	plans          planner
	// queries are the parsed queries with their selections, by the text of the query.
	queries sync.Map
}

var noFields, _ = json.Select()

// New returns the engine of the schema whose root value of the queries is root: a pointer to the struct of the
// type Query.
func New(schema string, root any, opts Options) (*Engine, error) {
	parsed, err := gqlparser.LoadSchema(&ast.Source{Name: "schema.graphqls", Input: schema})
	if err != nil {
		return nil, err
	}
	if reflect.TypeOf(root).Kind() != reflect.Pointer {
		return nil, fmt.Errorf("the root value is %T, not a pointer", root)
	}
	return &Engine{schema: parsed, root: root, batch: !opts.DisableBatch, requestContext: opts.RequestContext}, nil
}

type httpRequest struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

// preparedQuery is a parsed and validated query.
type preparedQuery struct {
	doc  *ast.QueryDocument
	errs gqlerror.List
	// selections are the selections of the operations, by the name of the operation and the values of the
	// variables which the directives that decide the fields depend on ( @skip and @include with a variable ).
	selections sync.Map
	// directiveVariables are the names of those variables, by the name of the operation.
	directiveVariables sync.Map
}

// request is the state of a request, which the resolvers and the write phase refer to by the context.
type request struct {
	engine    *Engine
	variables map[string]any
	// results are the resolved values of the fields.
	results map[resultKey]result
	errors  gqlerror.List
	// fields are what the resolvers of the selected fields are given.
	fields map[*json.SelectedField]*Selected
}

type requestKey struct{}

// selected returns what the resolver of the selected field is given.
func (r *request) selected(field *json.SelectedField) *Selected {
	if s, ok := r.fields[field]; ok {
		return s
	}
	s := &Selected{Name: field.Name, Alias: field.Key, Selection: field.Sub}
	if info, ok := field.Value.(*fieldInfo); ok {
		if info.args != nil {
			s.Args = info.args
		} else {
			s.Args = info.field.ArgumentMap(r.variables)
		}
	}
	r.fields[field] = s
	return s
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrors(w, http.StatusBadRequest, gqlerror.List{gqlerror.Errorf("%s", err)})
		return
	}
	var hr httpRequest
	if err := json.Unmarshal(body, &hr); err != nil {
		writeErrors(w, http.StatusBadRequest, gqlerror.List{gqlerror.Errorf("json request body could not be decoded: %s", err)})
		return
	}
	prepared := e.prepare(hr.Query)
	if len(prepared.errs) > 0 {
		writeErrors(w, http.StatusUnprocessableEntity, prepared.errs)
		return
	}
	op := prepared.doc.Operations.ForName(hr.OperationName)
	if op == nil {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.Errorf("operation %s not found", hr.OperationName)})
		return
	}
	if op.Operation != ast.Query {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.Errorf("only queries are served")})
		return
	}
	vars, verr := validator.VariableValues(e.schema, op, hr.Variables)
	if verr != nil {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.WrapIfUnwrapped(verr)})
		return
	}
	sel, err := e.selectionOf(prepared, op, vars)
	if err != nil {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.Errorf("%s", err)})
		return
	}

	req := &request{engine: e, variables: vars, results: map[resultKey]result{}, fields: map[*json.SelectedField]*Selected{}}
	ctx := r.Context()
	if e.requestContext != nil {
		ctx = e.requestContext(ctx)
	}
	ctx = context.WithValue(ctx, requestKey{}, req)
	e.resolveAll(ctx, req, item{v: reflect.ValueOf(e.root), sel: sel})

	data, err := json.MarshalContext(ctx, e.root, json.WithSelection(sel))
	if err != nil {
		writeErrors(w, http.StatusOK, gqlerror.List{gqlerror.Errorf("%s", err)})
		return
	}
	var buf []byte
	if len(req.errors) > 0 {
		errs, err := json.Marshal(req.errors)
		if err != nil {
			writeErrors(w, http.StatusOK, gqlerror.List{gqlerror.Errorf("%s", err)})
			return
		}
		buf = make([]byte, 0, len(errs)+len(data)+len(`{"errors":,"data":}`))
		buf = append(buf, `{"errors":`...)
		buf = append(buf, errs...)
		buf = append(buf, `,"data":`...)
	} else {
		buf = make([]byte, 0, len(data)+len(`{"data":}`))
		buf = append(buf, `{"data":`...)
	}
	buf = append(buf, data...)
	buf = append(buf, '}')
	_, _ = w.Write(buf)
}

func (e *Engine) prepare(query string) *preparedQuery {
	if prepared, ok := e.queries.Load(query); ok {
		return prepared.(*preparedQuery)
	}
	doc, errs := gqlparser.LoadQuery(e.schema, query)
	prepared, _ := e.queries.LoadOrStore(query, &preparedQuery{doc: doc, errs: errs})
	return prepared.(*preparedQuery)
}

// selectionOf returns the selection of the operation for the variables, which is kept with the query: the
// selection is used again, and so is the code which go-json compiles for it.
func (e *Engine) selectionOf(prepared *preparedQuery, op *ast.OperationDefinition, vars map[string]any) (*json.Selection, error) {
	names, ok := prepared.directiveVariables.Load(op.Name)
	if !ok {
		b := &selectionBuilder{schema: e.schema, variables: vars}
		if _, err := b.selection([]ast.SelectionSet{op.SelectionSet}, "Query"); err != nil {
			return nil, err
		}
		names, _ = prepared.directiveVariables.LoadOrStore(op.Name, b.directiveVariables())
	}
	key, err := selectionKey(op.Name, names.([]string), vars)
	if err != nil {
		return nil, err
	}
	if sel, ok := prepared.selections.Load(key); ok {
		return sel.(*json.Selection), nil
	}
	b := &selectionBuilder{schema: e.schema, variables: vars}
	sel, err := b.selection([]ast.SelectionSet{op.SelectionSet}, "Query")
	if err != nil {
		return nil, err
	}
	kept, _ := prepared.selections.LoadOrStore(key, sel)
	return kept.(*json.Selection), nil
}

// selectionKey returns the key of the selection of the operation for the values of the variables which decide
// its fields.
func selectionKey(operation string, names []string, vars map[string]any) (string, error) {
	if len(names) == 0 {
		return operation, nil
	}
	values := make([]any, len(names))
	for i, name := range names {
		values[i] = vars[name]
	}
	b, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return operation + string(b), nil
}

func writeErrors(w http.ResponseWriter, status int, errs gqlerror.List) {
	w.WriteHeader(status)
	b, err := json.Marshal(map[string]any{"errors": errs})
	if err != nil {
		b = []byte(fmt.Sprintf(`{"errors":[{"message":%q}]}`, err.Error()))
	}
	_, _ = w.Write(b)
}
