// Package gojsonserver is a GraphQL server of the example built on go-json: a query is turned into a
// json.Selection, and the response is written by one call of json.MarshalContext with the selection.
//
// The query is parsed and validated by gqlparser, as gqlgen does. The selection of a query is made once and
// kept with the parsed query, and go-json keeps the code compiled for it in the selection: the fields which a
// query selects are found once, not for every value.
package gojsonserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/vektah/gqlparser/v2/validator"

	"github.com/goccy/go-json"
	"github.com/goccy/go-json/examples/graphql/store"
)

type Server struct {
	schema *ast.Schema
	store  *store.Store
	query  *Query
	// queries are the parsed queries with their selections, by the text of the query.
	queries sync.Map
}

// NewHandler returns the handler of the GraphQL requests on the store.
func NewHandler(schema string, s *store.Store) (*Server, error) {
	parsed, err := gqlparser.LoadSchema(&ast.Source{Name: "schema.graphqls", Input: schema})
	if err != nil {
		return nil, err
	}
	return &Server{schema: parsed, store: s, query: newQuery(s)}, nil
}

type request struct {
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

// requestState is what the resolvers of a request refer to by the context.
type requestState struct {
	store     *store.Store
	variables map[string]any
	// batched are the fields which have been fetched for all the values of a list ( prefetch ).
	batched map[*json.SelectedField]bool
}

type requestStateKey struct{}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrors(w, http.StatusBadRequest, gqlerror.List{gqlerror.Errorf("%s", err)})
		return
	}
	var req request
	if err := json.Unmarshal(body, &req); err != nil {
		writeErrors(w, http.StatusBadRequest, gqlerror.List{gqlerror.Errorf("json request body could not be decoded: %s", err)})
		return
	}
	prepared := s.prepare(req.Query)
	if len(prepared.errs) > 0 {
		writeErrors(w, http.StatusUnprocessableEntity, prepared.errs)
		return
	}
	op := prepared.doc.Operations.ForName(req.OperationName)
	if op == nil {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.Errorf("operation %s not found", req.OperationName)})
		return
	}
	if op.Operation != ast.Query {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.Errorf("only queries are served")})
		return
	}
	vars, verr := validator.VariableValues(s.schema, op, req.Variables)
	if verr != nil {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.WrapIfUnwrapped(verr)})
		return
	}
	sel, err := s.selectionOf(prepared, op, vars)
	if err != nil {
		writeErrors(w, http.StatusUnprocessableEntity, gqlerror.List{gqlerror.Errorf("%s", err)})
		return
	}
	ctx := context.WithValue(r.Context(), requestStateKey{}, &requestState{store: s.store, variables: vars})
	data, err := json.MarshalContext(ctx, s.query, json.WithSelection(sel))
	if err != nil {
		writeErrors(w, http.StatusOK, gqlerror.List{gqlerror.Errorf("%s", err)})
		return
	}
	buf := make([]byte, 0, len(data)+len(`{"data":}`))
	buf = append(buf, `{"data":`...)
	buf = append(buf, data...)
	buf = append(buf, '}')
	_, _ = w.Write(buf)
}

func (s *Server) prepare(query string) *preparedQuery {
	if prepared, ok := s.queries.Load(query); ok {
		return prepared.(*preparedQuery)
	}
	doc, errs := gqlparser.LoadQuery(s.schema, query)
	prepared, _ := s.queries.LoadOrStore(query, &preparedQuery{doc: doc, errs: errs})
	return prepared.(*preparedQuery)
}

// selectionOf returns the selection of the operation for the variables, which is kept with the query: the
// selection is used again, and so is the code which go-json compiles for it.
func (s *Server) selectionOf(prepared *preparedQuery, op *ast.OperationDefinition, vars map[string]any) (*json.Selection, error) {
	names, ok := prepared.directiveVariables.Load(op.Name)
	if !ok {
		b := &selectionBuilder{schema: s.schema, variables: vars}
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
	b := &selectionBuilder{schema: s.schema, variables: vars}
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
