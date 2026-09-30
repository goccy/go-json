package gojsonserver

import (
	"context"
	"errors"

	"github.com/goccy/go-json/examples/graphql/gql"
	"github.com/goccy/go-json/examples/graphql/store"
)

// The types of the schema are Go structs whose JSON keys are the names of the fields of GraphQL. A field
// without arguments is a plain Go field, which go-json writes when it is selected. A field with a resolver is a
// gql.Field, which the resolve phase resolves only when it is selected.

type Query struct {
	Typename string                 `json:"__typename"`
	User     gql.Field[userByID]    `json:"user"`
	Users    gql.Field[allUsers]    `json:"users"`
	Posts    gql.Field[allPosts]    `json:"posts"`
	Search   gql.Field[searchUnion] `json:"search"`
}

type User struct {
	Typename string                 `json:"__typename"`
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Bio      *string                `json:"bio"`
	Age      int                    `json:"age"`
	Score    float64                `json:"score"`
	Active   bool                   `json:"active"`
	Posts    gql.Field[userPosts]   `json:"posts"`
	Friends  gql.Field[userFriends] `json:"friends"`
	Secret   gql.Field[userSecret]  `json:"secret"`
}

type Post struct {
	Typename string                  `json:"__typename"`
	ID       string                  `json:"id"`
	Title    string                  `json:"title"`
	Body     string                  `json:"body"`
	Tags     []string                `json:"tags"`
	Likes    int                     `json:"likes"`
	Author   *User                   `json:"author"`
	Comments gql.Field[postComments] `json:"comments"`
}

type Comment struct {
	Typename string `json:"__typename"`
	ID       string `json:"id"`
	Body     string `json:"body"`
	Author   *User  `json:"author"`
}

// models are the values of the types of the schema for the values of the store.
type models struct {
	store    *store.Store
	users    map[*store.User]*User
	posts    map[*store.Post]*Post
	comments map[*store.Comment]*Comment
}

func newModels(s *store.Store) *models {
	m := &models{
		store:    s,
		users:    make(map[*store.User]*User, len(s.Users)),
		posts:    make(map[*store.Post]*Post, len(s.Posts)),
		comments: map[*store.Comment]*Comment{},
	}
	for _, u := range s.Users {
		m.users[u] = &User{
			Typename: "User", ID: u.ID, Name: u.Name, Email: u.Email, Bio: u.Bio,
			Age: u.Age, Score: u.Score, Active: u.Active,
			Posts:   gql.Field[userPosts]{R: userPosts{m: m, user: u}},
			Friends: gql.Field[userFriends]{R: userFriends{m: m, user: u}},
			Secret:  gql.Field[userSecret]{R: userSecret{user: u}},
		}
	}
	for _, p := range s.Posts {
		m.posts[p] = &Post{
			Typename: "Post", ID: p.ID, Title: p.Title, Body: p.Body, Tags: p.Tags, Likes: p.Likes,
			Author:   m.users[p.Author],
			Comments: gql.Field[postComments]{R: postComments{m: m, post: p}},
		}
		for _, c := range p.Comments {
			m.comments[c] = &Comment{Typename: "Comment", ID: c.ID, Body: c.Body, Author: m.users[c.Author]}
		}
	}
	return m
}

func (m *models) query() *Query {
	return &Query{
		Typename: "Query",
		User:     gql.Field[userByID]{R: userByID{m: m}},
		Users:    gql.Field[allUsers]{R: allUsers{m: m}},
		Posts:    gql.Field[allPosts]{R: allPosts{m: m}},
		Search:   gql.Field[searchUnion]{R: searchUnion{m: m}},
	}
}

func mapTo[S, M any](list []*S, index map[*S]*M) []*M {
	mapped := make([]*M, len(list))
	for i, v := range list {
		mapped[i] = index[v]
	}
	return mapped
}

type userByID struct{ m *models }

func (r userByID) Resolve(_ context.Context, field *gql.Selected) (any, error) {
	r.m.store.Fetch()
	u := r.m.store.User(field.String("id"))
	if u == nil {
		return nil, nil
	}
	return r.m.users[u], nil
}

type allUsers struct{ m *models }

func (r allUsers) Resolve(_ context.Context, field *gql.Selected) (any, error) {
	r.m.store.Fetch()
	return mapTo(store.First(r.m.store.Users, field.Int("first")), r.m.users), nil
}

type allPosts struct{ m *models }

func (r allPosts) Resolve(_ context.Context, field *gql.Selected) (any, error) {
	r.m.store.Fetch()
	return mapTo(store.First(r.m.store.Posts, field.Int("first")), r.m.posts), nil
}

type searchUnion struct{ m *models }

func (r searchUnion) Resolve(_ context.Context, field *gql.Selected) (any, error) {
	r.m.store.Fetch()
	results := r.m.store.Search(field.String("text"), field.Int("first"))
	values := make([]any, len(results))
	for i, result := range results {
		switch result := result.(type) {
		case *store.User:
			values[i] = r.m.users[result]
		case *store.Post:
			values[i] = r.m.posts[result]
		}
	}
	return values, nil
}

// userPosts, userFriends and postComments are resolved one by one ( Resolve, through the data loaders if the
// request has them ) or for all the values of a level at once ( ResolveBatch ).

type userPosts struct {
	m    *models
	user *store.User
}

func (r userPosts) Resolve(ctx context.Context, field *gql.Selected) (any, error) {
	posts, err := loadOne(ctx, r.user, r.m.store.PostsOf, func(l *store.Loaders) loader[*store.User, []*store.Post] { return l.PostsOf })
	if err != nil {
		return nil, err
	}
	return mapTo(store.First(posts, field.Int("first")), r.m.posts), nil
}

func (r userPosts) ResolveBatch(_ context.Context, field *gql.Selected, batch []userPosts) ([]any, []error) {
	users := make([]*store.User, len(batch))
	for i, b := range batch {
		users[i] = b.user
	}
	values := make([]any, len(batch))
	for i, posts := range r.m.store.PostsOf(users) {
		values[i] = mapTo(store.First(posts, field.Int("first")), r.m.posts)
	}
	return values, nil
}

type userFriends struct {
	m    *models
	user *store.User
}

func (r userFriends) Resolve(ctx context.Context, field *gql.Selected) (any, error) {
	friends, err := loadOne(ctx, r.user, r.m.store.FriendsOf, func(l *store.Loaders) loader[*store.User, []*store.User] { return l.FriendsOf })
	if err != nil {
		return nil, err
	}
	return mapTo(store.First(friends, field.Int("first")), r.m.users), nil
}

func (r userFriends) ResolveBatch(_ context.Context, field *gql.Selected, batch []userFriends) ([]any, []error) {
	users := make([]*store.User, len(batch))
	for i, b := range batch {
		users[i] = b.user
	}
	values := make([]any, len(batch))
	for i, friends := range r.m.store.FriendsOf(users) {
		values[i] = mapTo(store.First(friends, field.Int("first")), r.m.users)
	}
	return values, nil
}

type postComments struct {
	m    *models
	post *store.Post
}

func (r postComments) Resolve(ctx context.Context, field *gql.Selected) (any, error) {
	comments, err := loadOne(ctx, r.post, r.m.store.CommentsOf, func(l *store.Loaders) loader[*store.Post, []*store.Comment] { return l.CommentsOf })
	if err != nil {
		return nil, err
	}
	return mapTo(store.First(comments, field.Int("first")), r.m.comments), nil
}

func (r postComments) ResolveBatch(_ context.Context, field *gql.Selected, batch []postComments) ([]any, []error) {
	posts := make([]*store.Post, len(batch))
	for i, b := range batch {
		posts[i] = b.post
	}
	values := make([]any, len(batch))
	for i, comments := range r.m.store.CommentsOf(posts) {
		values[i] = mapTo(store.First(comments, field.Int("first")), r.m.comments)
	}
	return values, nil
}

type userSecret struct{ user *store.User }

func (r userSecret) Resolve(context.Context, *gql.Selected) (any, error) {
	return nil, errors.New("the secret of " + r.user.ID + " is not readable")
}

type loader[K comparable, V any] interface {
	Load(ctx context.Context, key K) (V, error)
}

// loadOne fetches the value of the key by the data loader of the request, or by a fetch of its own if the
// request has no data loaders.
func loadOne[K comparable, V any](ctx context.Context, key K, fetch func([]K) []V, loaderOf func(*store.Loaders) loader[K, V]) (V, error) {
	if loaders := store.LoadersFrom(ctx); loaders != nil {
		return loaderOf(loaders).Load(ctx, key)
	}
	return fetch([]K{key})[0], nil
}
