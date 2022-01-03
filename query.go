package json

import (
	"github.com/goccy/go-json/internal/encoder"
)

type (
	FieldQuery       = encoder.FieldQuery
	FieldQueryString = encoder.FieldQueryString
)

var (
	FieldQueryFromContext  = encoder.FieldQueryFromContext
	SetFieldQueryToContext = encoder.SetFieldQueryToContext
)

func BuildFieldQuery(fields ...FieldQueryString) (*FieldQuery, error) {
	query, err := Marshal(fields)
	if err != nil {
		return nil, err
	}
	return FieldQueryString(query).Build()
}

func BuildSubFieldQuery(name string) *SubFieldQuery {
	return &SubFieldQuery{name: name}
}

type SubFieldQuery struct {
	name string
}

func (q *SubFieldQuery) Fields(fields ...FieldQueryString) FieldQueryString {
	query, _ := Marshal(map[string][]FieldQueryString{q.name: fields})
	return FieldQueryString(query)
}
