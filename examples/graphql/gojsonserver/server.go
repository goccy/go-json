// Package gojsonserver is the GraphQL server of the example built on go-json: the models and the resolvers of
// the schema on the engine of package gql, without code generation.
package gojsonserver

import (
	"context"
	"net/http"
	"time"

	"github.com/goccy/go-json/examples/graphql/gql"
	"github.com/goccy/go-json/examples/graphql/store"
)

// Options are how the server resolves the fields.
type Options struct {
	// LoaderWait is how long the data loaders of a request wait for the keys of a batch, or 0 if the server
	// uses no data loaders. The fields are resolved one by one through them, not by batch resolvers.
	LoaderWait time.Duration
	// DisableBatch makes every field resolved by a call of its own, even of a batch resolver.
	DisableBatch bool
}

// NewHandler returns the handler of the GraphQL requests on the store.
func NewHandler(schema string, s *store.Store, opts Options) (http.Handler, error) {
	gqlOpts := gql.Options{DisableBatch: opts.DisableBatch}
	if opts.LoaderWait > 0 {
		gqlOpts.DisableBatch = true
		gqlOpts.RequestContext = func(ctx context.Context) context.Context {
			return store.WithLoaders(ctx, s.NewLoaders(opts.LoaderWait))
		}
	}
	return gql.New(schema, newModels(s).query(), gqlOpts)
}
