package json

import "fmt"

type Query struct {
	query *subQuery
	err   error
}

func (q *Query) String() string {
	if q.err != nil {
		return ""
	}
	if q.query == nil {
		return ""
	}
	b, err := Marshal(q.query.dump())
	if err != nil {
		return ""
	}
	return string(b)
}

func (q *Query) Error() error {
	return q.err
}

func (q *Query) Fields(fieldNameOrQueryList ...interface{}) *Query {
	for _, fieldNameOrQuery := range fieldNameOrQueryList {
		switch v := fieldNameOrQuery.(type) {
		case string:
			q.fields = append(q.fields, &subQuery{name: v})
		case *Query:
			q.fields = append(q.fields, v.query)
			q.err = v.err
		default:
			q.err = fmt.Errorf("children types must be string or *Query but found %T", fieldNameOrQuery)
		}
		if q.err != nil {
			break
		}
	}
	return q
}

func NewQuery(name ...string) *Query {
	if len(name) > 1 {
		return &Query{err: fmt.Errorf(
			"NewQuery's argument allow empty or single name only, but passed %v", name,
		)}
	}
	return &Query{query: &subQuery{name: name}}
}

type subQuery struct {
	name   string
	fields []*subQuery
}

func (q *subQuery) dump() interface{} {
	fields := []interface{}{}
	for _, field := range q.fields {
		fields = append(fields, field.dump())
	}
	if q.name != "" {
		return map[string][]interface{}{
			q.name: fields,
		}
	}
	return interface{}(fields)
}
