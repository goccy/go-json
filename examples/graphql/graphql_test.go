package graphql_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-json/examples/graphql"
	"github.com/goccy/go-json/examples/graphql/gojsonserver"
	"github.com/goccy/go-json/examples/graphql/gqlgenserver"
	"github.com/goccy/go-json/examples/graphql/store"
)

type query struct {
	name      string
	query     string
	variables map[string]any
}

// queries are the queries which the servers are compared by: each server must write the same response.
var queries = []query{
	{name: "user", query: `{ user(id: "u2") { id name email bio age score active } }`},
	{name: "missing_user", query: `{ user(id: "none") { id name } }`},
	{name: "users_100", query: `{ users(first: 100) { id name age score active } }`},
	{name: "users_1000_wide", query: `{ users(first: 1000) { id name email bio age score active } }`},
	{name: "posts_nested", query: `{ posts(first: 50) { id title tags likes author { id name } comments(first: 5) { id body author { name } } } }`},
	{name: "friends_deep", query: `{ users(first: 20) { name friends(first: 5) { name friends(first: 5) { name posts(first: 2) { title } } } } }`},
	{
		name: "search_union",
		query: `query Search($text: String!, $first: Int) {
			search(text: $text, first: $first) { __typename ... on User { id name } ...PostFields }
		}
		fragment PostFields on Post { id title author { name } }`,
		variables: map[string]any{"text": "1", "first": 30},
	},
	{name: "search_posts_only", query: `{ search(text: "User 1", first: 8) { ... on Post { title } } }`},
	{name: "aliases_and_order", query: `{ b: users(first: 2) { email } a: user(id: "u2") { active zzz: name id posts(first: 1) { likes t: title } } }`},
	{name: "skip", query: `query Q($s: Boolean!) { users(first: 3) { id name @skip(if: $s) email @include(if: $s) } }`, variables: map[string]any{"s": true}},
	{name: "include", query: `query Q($s: Boolean!) { users(first: 3) { id name @skip(if: $s) email @include(if: $s) } }`, variables: map[string]any{"s": false}},
	{name: "typename", query: `{ __typename users(first: 2) { __typename id posts(first: 1) { __typename } } }`},
	{name: "default_first", query: `{ users { id } }`},
	{name: "null_first", query: `{ user(id: "u3") { posts(first: null) { id } } }`},
	{name: "variable_first", query: `query Q($n: Int) { user(id: "u3") { friends(first: $n) { id } } }`, variables: map[string]any{"n": 3}},
	{name: "merged_fields", query: `{ user(id: "u1") { id posts(first: 2) { id } ... on User { posts(first: 2) { title } name } } }`},
	{name: "operation_name", query: `query A { user(id: "u1") { id } } query B { user(id: "u1") { name } }`},
}

func newHandlers(t testing.TB) (gojson, gqlgen http.Handler) {
	s := store.New(1000, 5, 5)
	server, err := gojsonserver.NewHandler(graphql.Schema, s)
	if err != nil {
		t.Fatal(err)
	}
	return server, gqlgenserver.NewHandler(s)
}

func requestBody(t testing.TB, q query, operationName string) []byte {
	body, err := json.Marshal(map[string]any{"query": q.query, "variables": q.variables, "operationName": operationName})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func serve(h http.Handler, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSameResponses(t *testing.T) {
	gojson, gqlgen := newHandlers(t)
	for _, q := range queries {
		operationName := ""
		if q.name == "operation_name" {
			operationName = "B"
		}
		t.Run(q.name, func(t *testing.T) {
			body := requestBody(t, q, operationName)
			// twice: the second is served by the selection kept for the query.
			for i := 0; i < 2; i++ {
				want := serve(gqlgen, body)
				got := serve(gojson, body)
				if want.Code != http.StatusOK || bytes.Contains(want.Body.Bytes(), []byte(`"errors"`)) {
					t.Fatalf("gqlgen: %d %s", want.Code, want.Body)
				}
				if got.Code != want.Code || !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
					t.Fatalf("go-json:\n%d %s\ngqlgen:\n%d %s", got.Code, got.Body, want.Code, want.Body)
				}
			}
		})
	}
}

// discardWriter is the response writer of the benchmarks, which doesn't keep the response.
type discardWriter struct {
	header http.Header
	n      int
}

func (w *discardWriter) Header() http.Header         { return w.header }
func (w *discardWriter) WriteHeader(int)             {}
func (w *discardWriter) Write(b []byte) (int, error) { w.n += len(b); return len(b), nil }

type readCloser struct{ *bytes.Reader }

func (readCloser) Close() error { return nil }

// benchServer is a server which a benchmark serves the requests of the body by.
type benchServer struct {
	name    string
	handler http.Handler
}

// requester makes the requests of a benchmark, reusing the request and the response writer.
type requester struct {
	req    *http.Request
	reader *bytes.Reader
	body   []byte
	w      *discardWriter
}

func newRequester(body []byte) *requester {
	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Content-Type", "application/json")
	return &requester{req: req, reader: bytes.NewReader(body), body: body, w: &discardWriter{header: http.Header{}}}
}

func (r *requester) serve(h http.Handler) {
	r.reader.Reset(r.body)
	r.req.Body = readCloser{r.reader}
	r.w.n = 0
	h.ServeHTTP(r.w, r.req)
}

// checkServes fails the benchmark unless the server answers the body with data, so that a benchmark never
// measures a rejected request.
func checkServes(b *testing.B, h http.Handler, body []byte) {
	b.Helper()
	res := serve(h, body)
	if res.Code != http.StatusOK || bytes.Contains(res.Body.Bytes(), []byte(`"errors"`)) || !bytes.HasPrefix(res.Body.Bytes(), []byte(`{"data":{`)) {
		b.Fatalf("the request is not served: %d %s", res.Code, res.Body)
	}
}

func BenchmarkServers(b *testing.B) {
	gojson, gqlgen := newHandlers(b)
	for _, q := range queries {
		if q.name == "operation_name" {
			continue
		}
		body := requestBody(b, q, "")
		for _, server := range []benchServer{{"gqlgen", gqlgen}, {"go-json", gojson}} {
			b.Run(q.name+"/"+server.name, func(b *testing.B) {
				checkServes(b, server.handler, body)
				r := newRequester(body)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.serve(server.handler)
				}
				b.SetBytes(int64(r.w.n))
			})
		}
	}
}

// BenchmarkServersParallel serves the requests from all the CPUs, as a server does: the allocations of a
// request are paid by the GC which runs beside the requests.
func BenchmarkServersParallel(b *testing.B) {
	gojson, gqlgen := newHandlers(b)
	for _, q := range queries {
		switch q.name {
		case "user", "users_100", "posts_nested", "friends_deep", "search_union":
		default:
			continue
		}
		body := requestBody(b, q, "")
		for _, server := range []benchServer{{"gqlgen", gqlgen}, {"go-json", gojson}} {
			b.Run(q.name+"/"+server.name, func(b *testing.B) {
				checkServes(b, server.handler, body)
				b.ReportAllocs()
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					r := newRequester(body)
					for pb.Next() {
						r.serve(server.handler)
					}
				})
			})
		}
	}
}

// BenchmarkServersWithLatency serves a query whose resolvers fetch from a store of a latency, as a database:
// gqlgen runs the resolvers of the elements of a list concurrently, while the go-json server runs them one by
// one while it writes the response.
func BenchmarkServersWithLatency(b *testing.B) {
	s := store.New(1000, 5, 5)
	s.Latency = time.Millisecond
	gojson, err := gojsonserver.NewHandler(graphql.Schema, s)
	if err != nil {
		b.Fatal(err)
	}
	gqlgen := gqlgenserver.NewHandler(s)
	body := requestBody(b, query{query: `{ users(first: 10) { name posts(first: 2) { title } } }`}, "")
	for _, server := range []benchServer{{"gqlgen", gqlgen}, {"go-json", gojson}} {
		b.Run("users_10_posts/"+server.name, func(b *testing.B) {
			checkServes(b, server.handler, body)
			r := newRequester(body)
			fetches := s.Fetches.Load()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.serve(server.handler)
			}
			b.ReportMetric(float64(s.Fetches.Load()-fetches)/float64(b.N), "fetches/op")
		})
	}
}
