// Command server serves the GraphQL example: the server built on go-json, or the one which gqlgen generates.
//
//	go run ./cmd/server -server go-json
//	curl -s localhost:8080/query -H 'Content-Type: application/json' -d '{"query":"{ users(first: 2) { name } }"}'
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/goccy/go-json/examples/graphql"
	"github.com/goccy/go-json/examples/graphql/gojsonserver"
	"github.com/goccy/go-json/examples/graphql/gqlgenserver"
	"github.com/goccy/go-json/examples/graphql/store"
)

func main() {
	addr := flag.String("addr", ":8080", "the address to listen on")
	server := flag.String("server", "go-json", "the server to run: go-json or gqlgen")
	flag.Parse()

	s := store.New(1000, 5, 5)
	var handler http.Handler
	switch *server {
	case "go-json":
		h, err := gojsonserver.NewHandler(graphql.Schema, s)
		if err != nil {
			log.Fatal(err)
		}
		handler = h
	case "gqlgen":
		handler = gqlgenserver.NewHandler(s)
	default:
		log.Fatalf("unknown server %q", *server)
	}
	http.Handle("/query", handler)
	log.Printf("the %s server listens on %s", *server, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
