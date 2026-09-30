package graphql_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sort"
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
	// errors is whether the response has errors.
	errors bool
}

// queries are the queries which the servers are compared by: each server must write the same response.
var queries = []query{
	{name: "user", query: `{ user(id: "u2") { id name email bio age score active } }`},
	{name: "missing_user", query: `{ user(id: "none") { id name } }`},
	{name: "users_100", query: `{ users(first: 100) { id name age score active } }`},
	{name: "users_1000_wide", query: `{ users(first: 1000) { id name email bio age score active } }`},
	{name: "posts_nested", query: `{ posts(first: 50) { id title tags likes author { id name } comments(first: 5) { id body author { name } } } }`},
	{name: "friends_deep", query: `{ users(first: 20) { name friends(first: 5) { name friends(first: 5) { name posts(first: 2) { title } } } } }`},
	{name: "three_levels", query: `{ users(first: 10) { name posts(first: 2) { title comments(first: 2) { body } } } }`},
	{
		name: "search_union",
		query: `query Search($text: String!, $first: Int) {
			search(text: $text, first: $first) { __typename ... on User { id name posts(first: 1) { title } } ...PostFields }
		}
		fragment PostFields on Post { id title author { name } comments(first: 1) { body } }`,
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
	{name: "same_value_twice", query: `{ a: user(id: "u1") { friends(first: 2) { name } } b: user(id: "u1") { friends(first: 2) { id } } }`},
	{name: "operation_name", query: `query A { user(id: "u1") { id } } query B { user(id: "u1") { name } }`},
	{name: "errors", query: `{ users(first: 3) { id secret friends(first: 2) { secret } } }`, errors: true},
}

// server is a server under test.
type server struct {
	name    string
	handler http.Handler
}

// newServers returns the servers on the same store: gqlgen, the reference of the responses, first. The waits of
// the data loaders are short, so that the tests which compare the responses run fast.
func newServers(t testing.TB, s *store.Store, loaderWait time.Duration) []server {
	servers := []server{
		{"gqlgen", gqlgenserver.NewHandler(s, gqlgenserver.Options{})},
		{"gqlgen+dataloader", gqlgenserver.NewHandler(s, gqlgenserver.Options{LoaderWait: loaderWait})},
	}
	for _, v := range []struct {
		name string
		opts gojsonserver.Options
	}{
		{"go-json+batch", gojsonserver.Options{}},
		{"go-json", gojsonserver.Options{DisableBatch: true}},
		{"go-json+dataloader", gojsonserver.Options{LoaderWait: loaderWait}},
	} {
		h, err := gojsonserver.NewHandler(graphql.Schema, s, v.opts)
		if err != nil {
			t.Fatal(err)
		}
		servers = append(servers, server{v.name, h})
	}
	return servers
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

// response is a response whose errors are in the order of their paths: gqlgen adds the errors of the fields
// resolved concurrently in the order they happen.
type response struct {
	data   []byte
	errors []string
}

func parseResponse(t *testing.T, body []byte) response {
	t.Helper()
	var v struct {
		Errors []json.RawMessage `json:"errors"`
		Data   json.RawMessage   `json:"data"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("%s: %s", err, body)
	}
	r := response{data: v.Data}
	for _, e := range v.Errors {
		r.errors = append(r.errors, string(e))
	}
	sort.Strings(r.errors)
	return r
}

func TestSameResponses(t *testing.T) {
	servers := newServers(t, store.New(1000, 5, 5), 100*time.Microsecond)
	for _, q := range queries {
		operationName := ""
		if q.name == "operation_name" {
			operationName = "B"
		}
		t.Run(q.name, func(t *testing.T) {
			body := requestBody(t, q, operationName)
			// twice: the second is served by the selection kept for the query.
			for i := 0; i < 2; i++ {
				ref := serve(servers[0].handler, body)
				hasErrors := bytes.Contains(ref.Body.Bytes(), []byte(`"errors"`))
				if ref.Code != http.StatusOK || hasErrors != q.errors {
					t.Fatalf("gqlgen: %d %s", ref.Code, ref.Body)
				}
				want := parseResponse(t, ref.Body.Bytes())
				for _, s := range servers[1:] {
					got := serve(s.handler, body)
					if got.Code != ref.Code {
						t.Fatalf("%s: %d %s", s.name, got.Code, got.Body)
					}
					if !q.errors && !bytes.Equal(got.Body.Bytes(), ref.Body.Bytes()) {
						t.Fatalf("%s:\n%s\ngqlgen:\n%s", s.name, got.Body, ref.Body)
					}
					resp := parseResponse(t, got.Body.Bytes())
					if !bytes.Equal(resp.data, want.data) || !equalStrings(resp.errors, want.errors) {
						t.Fatalf("%s:\n%s\ngqlgen:\n%s", s.name, got.Body, ref.Body)
					}
				}
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
	if res.Code != http.StatusOK || !bytes.Contains(res.Body.Bytes(), []byte(`"data":{`)) {
		b.Fatalf("the request is not served: %d %s", res.Code, res.Body)
	}
}

// cpuServers are the servers compared by the CPU they take: without data loaders, whose waits are not CPU.
func cpuServers(b *testing.B) []server {
	var servers []server
	for _, s := range newServers(b, store.New(1000, 5, 5), time.Millisecond) {
		switch s.name {
		case "gqlgen", "go-json+batch", "go-json":
			servers = append(servers, s)
		}
	}
	return servers
}

func BenchmarkServers(b *testing.B) {
	servers := cpuServers(b)
	for _, q := range queries {
		if q.name == "operation_name" || q.errors {
			continue
		}
		body := requestBody(b, q, "")
		for _, s := range servers {
			b.Run(q.name+"/"+s.name, func(b *testing.B) {
				checkServes(b, s.handler, body)
				r := newRequester(body)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.serve(s.handler)
				}
				b.SetBytes(int64(r.w.n))
			})
		}
	}
}

// BenchmarkServersParallel serves the requests from all the CPUs, as a server does: the allocations of a
// request are paid by the GC which runs beside the requests.
func BenchmarkServersParallel(b *testing.B) {
	servers := cpuServers(b)
	for _, q := range queries {
		switch q.name {
		case "user", "users_100", "posts_nested", "friends_deep", "three_levels", "search_union":
		default:
			continue
		}
		body := requestBody(b, q, "")
		for _, s := range servers {
			b.Run(q.name+"/"+s.name, func(b *testing.B) {
				checkServes(b, s.handler, body)
				b.ReportAllocs()
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					r := newRequester(body)
					for pb.Next() {
						r.serve(s.handler)
					}
				})
			})
		}
	}
}

// BenchmarkServersWithLatency serves queries whose resolvers fetch from a store of a latency of 1 ms, as from a
// database, with the data loaders waiting 1 ms for a batch.
func BenchmarkServersWithLatency(b *testing.B) {
	s := store.New(1000, 5, 5)
	s.Latency = time.Millisecond
	servers := newServers(b, s, time.Millisecond)
	for _, q := range queries {
		switch q.name {
		case "three_levels", "friends_deep":
		default:
			continue
		}
		body := requestBody(b, q, "")
		for _, srv := range servers {
			b.Run(q.name+"/"+srv.name, func(b *testing.B) {
				checkServes(b, srv.handler, body)
				r := newRequester(body)
				fetches := s.Fetches.Load()
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.serve(srv.handler)
				}
				b.StopTimer()
				b.ReportMetric(float64(s.Fetches.Load()-fetches)/float64(b.N), "fetches/op")
			})
		}
	}
}
