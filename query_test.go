package json_test

import (
	"context"
	"reflect"
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
	ZD []queryTestW
	ZE map[string]queryTestV
}

type queryTestW struct {
	WA string
	WB bool
	WC int
}

type queryTestV struct {
	VA string
	VB bool
	VC int
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
				json.BuildSubFieldQuery("ZD").Fields(json.BuildSubFieldQuery("#").Fields(
					"WA", "WC")),
				json.BuildSubFieldQuery("ZE").Fields(json.BuildSubFieldQuery("#").Fields(
					"VA", "VC")),
			),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(query, &json.FieldQuery{
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
							{
								Name: "ZD",
								Fields: []*json.FieldQuery{
									{
										Name: "#",
										Fields: []*json.FieldQuery{
											{
												Name: "WA",
											},
											{
												Name: "WC",
											},
										},
									},
								},
							},
							{
								Name: "ZE",
								Fields: []*json.FieldQuery{
									{
										Name: "#",
										Fields: []*json.FieldQuery{
											{
												Name: "VA",
											},
											{
												Name: "VC",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}) {
		t.Fatal("cannot get query")
	}
	queryStr, err := query.QueryString()
	if err != nil {
		t.Fatal(err)
	}
	if queryStr != `["XA","XB",{"XC":["YA","YB",{"YC":["ZA","ZB",{"ZD":[{"#":["WA","WC"]}]},{"ZE":[{"#":["VA","VC"]}]}]}]}]` {
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
				ZD: []queryTestW{
					{WA: "wa1", WB: true, WC: 1},
					{WA: "wa2", WB: true, WC: 1},
				},
				ZE: map[string]queryTestV{
					"key1": {VA: "va1", VB: true, VC: 1},
					"key2": {VA: "va2", VB: true, VC: 1},
				},
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
	expected := `{"XA":1,"XB":"xb","XC":{"YA":2,"YB":"yb","YC":{"ZA":"za","ZB":true,"ZD":[{"WA":"wa1","WC":1},{"WA":"wa2","WC":1}],"ZE":{"key1":{"VA":"va1","VC":1},"key2":{"VA":"va2","VC":1}}}}}`
	got := string(b)
	if expected != got {
		t.Fatalf("failed to encode with field query: expected %q but got %q", expected, got)
	}
}
