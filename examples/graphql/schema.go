// Package graphql is an example of a GraphQL server built on the selection of go-json ( json.Selection ),
// compared with the one which gqlgen generates for the same schema.
package graphql

import _ "embed"

// Schema is the schema which both servers serve.
//
//go:embed schema.graphqls
var Schema string
