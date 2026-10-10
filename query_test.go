package json_test

import (
	"context"
	"sync"
	"testing"

	"github.com/goccy/go-json"
)

type queryTestX struct {
	XA int
	XB string
	XC *queryTestY
	XD bool
	XE float32
}

type queryTestY struct {
	YA int
	YB string
	YC *queryTestZ
	YD bool
	YE float32
}

type queryTestZ struct {
	ZA string
	ZB bool
	ZC int
}

func (z *queryTestZ) MarshalJSON(ctx context.Context) ([]byte, error) {
	type _queryTestZ queryTestZ
	return json.MarshalContext(ctx, (*_queryTestZ)(z))
}

func TestFieldQuery(t *testing.T) {
	query, err := json.BuildFieldQuery(
		"XA",
		"XB",
		json.BuildSubFieldQuery("XC").Fields(
			"YA",
			"YB",
			json.BuildSubFieldQuery("YC").Fields(
				"ZA",
				"ZB",
			),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	// by the query strings: a query keeps the key of its opcodes, which Build decides.
	want, err := (&json.FieldQuery{
		Fields: []*json.FieldQuery{
			{
				Name: "XA",
			},
			{
				Name: "XB",
			},
			{
				Name: "XC",
				Fields: []*json.FieldQuery{
					{
						Name: "YA",
					},
					{
						Name: "YB",
					},
					{
						Name: "YC",
						Fields: []*json.FieldQuery{
							{
								Name: "ZA",
							},
							{
								Name: "ZB",
							},
						},
					},
				},
			},
		},
	}).QueryString()
	if err != nil {
		t.Fatal(err)
	}
	if got, err := query.QueryString(); err != nil || got != want {
		t.Fatalf("cannot get query: got %s, %v, want %s", got, err, want)
	}
	queryStr, err := query.QueryString()
	if err != nil {
		t.Fatal(err)
	}
	if queryStr != `["XA","XB",{"XC":["YA","YB",{"YC":["ZA","ZB"]}]}]` {
		t.Fatalf("failed to create query string. %s", queryStr)
	}
	ctx := json.SetFieldQueryToContext(context.Background(), query)
	b, err := json.MarshalContext(ctx, &queryTestX{
		XA: 1,
		XB: "xb",
		XC: &queryTestY{
			YA: 2,
			YB: "yb",
			YC: &queryTestZ{
				ZA: "za",
				ZB: true,
				ZC: 3,
			},
			YD: true,
			YE: 4,
		},
		XD: true,
		XE: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"XA":1,"XB":"xb","XC":{"YA":2,"YB":"yb","YC":{"ZA":"za","ZB":true}}}`
	got := string(b)
	if expected != got {
		t.Fatalf("failed to encode with field query: expected %q but got %q", expected, got)
	}
}

// A query is built once and used by the encodings of many goroutines, which only read it.
func TestFieldQuerySharedByGoroutines(t *testing.T) {
	query, err := json.BuildFieldQuery("XA", "XB")
	if err != nil {
		t.Fatal(err)
	}
	ctx := json.SetFieldQueryToContext(context.Background(), query)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				b, err := json.MarshalContext(ctx, &queryTestX{XA: 1, XB: "xb"})
				if err != nil || string(b) != `{"XA":1,"XB":"xb"}` {
					t.Errorf("got %s, %v", b, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

type (
	queryTestRecursive struct {
		Y *queryTestEmbedsRecursive
		V int
	}
	queryTestEmbedsRecursive struct {
		*queryTestRecursive
		Q int
	}
)

// A query which selects none of the fields of a recursive struct embedded in another writes nothing of it.
func TestFieldQueryEmbeddedRecursiveStructWithNothingSelected(t *testing.T) {
	query, err := json.BuildFieldQuery(json.BuildSubFieldQuery("Y").Fields("Q", json.BuildSubFieldQuery("queryTestRecursive").Fields("Nope")))
	if err != nil {
		t.Fatal(err)
	}
	ctx := json.SetFieldQueryToContext(context.Background(), query)
	got, err := json.MarshalContext(ctx, &queryTestRecursive{Y: &queryTestEmbedsRecursive{&queryTestRecursive{V: 1}, 2}, V: 3})
	if err != nil || string(got) != `{"Y":{"Q":2}}` {
		t.Errorf("got %s, %v", got, err)
	}
}

func TestFieldQueryEmptyName(t *testing.T) {
	if _, err := json.BuildFieldQuery(""); err == nil {
		t.Error("got no error for an empty field name")
	}
}
