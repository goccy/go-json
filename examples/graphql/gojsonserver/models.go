package gojsonserver

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/goccy/go-json/examples/graphql/store"
)

// The types of the schema are Go structs whose JSON keys are the names of the fields of GraphQL. A field
// without arguments is a plain Go field, which go-json writes when it is selected. A field with arguments is a
// resolver: a value whose MarshalJSON(context.Context) is called only when the field is selected.

type Query struct {
	Typename string        `json:"__typename"`
	User     userByID      `json:"user"`
	Users    listOf[*User] `json:"users"`
	Posts    listOf[*Post] `json:"posts"`
	Search   search        `json:"search"`
}

type User struct {
	Typename string        `json:"__typename"`
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Email    string        `json:"email"`
	Bio      *string       `json:"bio"`
	Age      int           `json:"age"`
	Score    float64       `json:"score"`
	Active   bool          `json:"active"`
	Posts    listOf[*Post] `json:"posts"`
	Friends  listOf[*User] `json:"friends"`
}

type Post struct {
	Typename string           `json:"__typename"`
	ID       string           `json:"id"`
	Title    string           `json:"title"`
	Body     string           `json:"body"`
	Tags     []string         `json:"tags"`
	Likes    int              `json:"likes"`
	Author   *User            `json:"author"`
	Comments listOf[*Comment] `json:"comments"`
}

type Comment struct {
	Typename string `json:"__typename"`
	ID       string `json:"id"`
	Body     string `json:"body"`
	Author   *User  `json:"author"`
}

// newQuery returns the root of the values of the store.
func newQuery(s *store.Store) *Query {
	users := make(map[*store.User]*User, len(s.Users))
	for _, u := range s.Users {
		users[u] = &User{
			Typename: "User", ID: u.ID, Name: u.Name, Email: u.Email, Bio: u.Bio,
			Age: u.Age, Score: u.Score, Active: u.Active,
		}
	}
	posts := make(map[*store.Post]*Post, len(s.Posts))
	for _, p := range s.Posts {
		post := &Post{Typename: "Post", ID: p.ID, Title: p.Title, Body: p.Body, Tags: p.Tags, Likes: p.Likes, Author: users[p.Author]}
		for _, c := range p.Comments {
			post.Comments = append(post.Comments, &Comment{Typename: "Comment", ID: c.ID, Body: c.Body, Author: users[c.Author]})
		}
		posts[p] = post
	}
	q := &Query{Typename: "Query"}
	for _, u := range s.Users {
		user := users[u]
		for _, p := range u.Posts {
			user.Posts = append(user.Posts, posts[p])
		}
		for _, f := range u.Friends {
			user.Friends = append(user.Friends, users[f])
		}
		q.Users = append(q.Users, user)
	}
	for _, p := range s.Posts {
		q.Posts = append(q.Posts, posts[p])
	}
	q.User = userByID{users: map[string]*User{}}
	for _, u := range q.Users {
		q.User.users[u.ID] = u
	}
	q.Search = search{store: s, users: users, posts: posts}
	return q
}

// listOf is a list field with the argument first, which writes the first elements of the list.
type listOf[T any] []T

func (l listOf[T]) MarshalJSON(ctx context.Context) ([]byte, error) {
	state, args, err := resolve(ctx)
	if err != nil {
		return nil, err
	}
	list := store.First(l, intArgument(args, "first"))
	if len(list) > 0 {
		state.prefetch(json.SelectionFromContext(ctx))
	}
	return json.MarshalContext(ctx, []T(list))
}

type userByID struct {
	users map[string]*User
}

func (r userByID) MarshalJSON(ctx context.Context) ([]byte, error) {
	_, args, err := resolve(ctx)
	if err != nil {
		return nil, err
	}
	id, _ := args["id"].(string)
	return json.MarshalContext(ctx, r.users[id])
}

type search struct {
	store *store.Store
	users map[*store.User]*User
	posts map[*store.Post]*Post
}

// MarshalJSON writes each result by the selection of its type: the fields of a union are selected by the
// fragments of each type.
func (r search) MarshalJSON(ctx context.Context) ([]byte, error) {
	_, args, err := resolve(ctx)
	if err != nil {
		return nil, err
	}
	text, _ := args["text"].(string)
	byType := json.SelectedFieldFromContext(ctx).Value.(*fieldInfo).byType
	b := []byte{'['}
	for i, result := range r.store.Search(text, intArgument(args, "first")) {
		if i > 0 {
			b = append(b, ',')
		}
		var v any
		var sel *json.Selection
		switch result := result.(type) {
		case *store.User:
			v, sel = r.users[result], byType["User"]
		case *store.Post:
			v, sel = r.posts[result], byType["Post"]
		}
		out, err := json.MarshalContext(ctx, v, json.WithSelection(sel))
		if err != nil {
			return nil, err
		}
		b = append(b, out...)
	}
	return append(b, ']'), nil
}
