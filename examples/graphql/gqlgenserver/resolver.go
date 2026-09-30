package gqlgenserver

import "github.com/goccy/go-json/examples/graphql/store"

// Resolver is the root of the resolvers, which gqlgen keeps when it generates the code again: the resolvers of
// the fields are in schema.resolvers.go.
type Resolver struct {
	Store *store.Store
}
