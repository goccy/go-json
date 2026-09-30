package store

import (
	"context"
	"time"

	"github.com/vikstrous/dataloadgen"
)

// PostsOf returns the posts of each of the users by one fetch.
func (s *Store) PostsOf(users []*User) [][]*Post {
	s.Fetch()
	posts := make([][]*Post, len(users))
	for i, u := range users {
		posts[i] = u.Posts
	}
	return posts
}

// FriendsOf returns the friends of each of the users by one fetch.
func (s *Store) FriendsOf(users []*User) [][]*User {
	s.Fetch()
	friends := make([][]*User, len(users))
	for i, u := range users {
		friends[i] = u.Friends
	}
	return friends
}

// CommentsOf returns the comments of each of the posts by one fetch.
func (s *Store) CommentsOf(posts []*Post) [][]*Comment {
	s.Fetch()
	comments := make([][]*Comment, len(posts))
	for i, p := range posts {
		comments[i] = p.Comments
	}
	return comments
}

// Loaders are the data loaders of a request, which gather the fetches of the resolvers called concurrently into
// one fetch, as a GraphQL server uses them against the N+1 problem.
type Loaders struct {
	PostsOf    *dataloadgen.Loader[*User, []*Post]
	FriendsOf  *dataloadgen.Loader[*User, []*User]
	CommentsOf *dataloadgen.Loader[*Post, []*Comment]
}

// NewLoaders returns the loaders of a request, which wait for the keys of a batch for wait.
func (s *Store) NewLoaders(wait time.Duration) *Loaders {
	opt := dataloadgen.WithWait(wait)
	return &Loaders{
		PostsOf: dataloadgen.NewLoader(func(_ context.Context, users []*User) ([][]*Post, []error) {
			return s.PostsOf(users), nil
		}, opt),
		FriendsOf: dataloadgen.NewLoader(func(_ context.Context, users []*User) ([][]*User, []error) {
			return s.FriendsOf(users), nil
		}, opt),
		CommentsOf: dataloadgen.NewLoader(func(_ context.Context, posts []*Post) ([][]*Comment, []error) {
			return s.CommentsOf(posts), nil
		}, opt),
	}
}

type loadersKey struct{}

// WithLoaders returns the context of a request which has the loaders.
func WithLoaders(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey{}, loaders)
}

// LoadersFrom returns the loaders of the request, or nil if the server doesn't use them.
func LoadersFrom(ctx context.Context) *Loaders {
	loaders, _ := ctx.Value(loadersKey{}).(*Loaders)
	return loaders
}
