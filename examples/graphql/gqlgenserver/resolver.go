package gqlgenserver

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"

	"github.com/goccy/go-json/examples/graphql/store"
)

type Resolver struct {
	Store *store.Store
}

// Comments is the resolver for the comments field.
func (r *postResolver) Comments(ctx context.Context, obj *store.Post, first *int) ([]*store.Comment, error) {
	r.Store.Fetch()
	return store.First(obj.Comments, first), nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id string) (*store.User, error) {
	r.Store.Fetch()
	return r.Store.User(id), nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, first *int) ([]*store.User, error) {
	r.Store.Fetch()
	return store.First(r.Store.Users, first), nil
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context, first *int) ([]*store.Post, error) {
	r.Store.Fetch()
	return store.First(r.Store.Posts, first), nil
}

// Search is the resolver for the search field.
func (r *queryResolver) Search(ctx context.Context, text string, first *int) ([]store.SearchResult, error) {
	r.Store.Fetch()
	return r.Store.Search(text, first), nil
}

// Posts is the resolver for the posts field.
func (r *userResolver) Posts(ctx context.Context, obj *store.User, first *int) ([]*store.Post, error) {
	r.Store.Fetch()
	return store.First(obj.Posts, first), nil
}

// Friends is the resolver for the friends field.
func (r *userResolver) Friends(ctx context.Context, obj *store.User, first *int) ([]*store.User, error) {
	r.Store.Fetch()
	return store.First(obj.Friends, first), nil
}

// Post returns PostResolver implementation.
func (r *Resolver) Post() PostResolver { return &postResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// User returns UserResolver implementation.
func (r *Resolver) User() UserResolver { return &userResolver{r} }

type (
	postResolver  struct{ *Resolver }
	queryResolver struct{ *Resolver }
	userResolver  struct{ *Resolver }
)
