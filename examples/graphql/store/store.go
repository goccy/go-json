// Package store is the data which both GraphQL servers of the example serve: users, their posts and the
// comments on the posts, made the same every time, in memory.
package store

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type User struct {
	ID      string
	Name    string
	Email   string
	Bio     *string
	Age     int
	Score   float64
	Active  bool
	Posts   []*Post
	Friends []*User
}

type Post struct {
	ID       string
	Title    string
	Body     string
	Tags     []string
	Likes    int
	Author   *User
	Comments []*Comment
}

type Comment struct {
	ID     string
	Body   string
	Author *User
}

// SearchResult is a value of the union SearchResult: a *User or a *Post.
type SearchResult interface {
	isSearchResult()
}

func (*User) isSearchResult() {}
func (*Post) isSearchResult() {}

type Store struct {
	Users []*User
	Posts []*Post
	// Latency is how long a fetch of the store takes, as a query of a database would: every resolver of a
	// field with arguments fetches once.
	Latency time.Duration
	// Fetches is how many times the store has been fetched.
	Fetches   atomic.Int64
	usersByID map[string]*User
}

// Fetch waits for the latency of a fetch.
func (s *Store) Fetch() {
	s.Fetches.Add(1)
	if s.Latency > 0 {
		time.Sleep(s.Latency)
	}
}

// New returns the store of the users, each of which has postsPerUser posts with commentsPerPost comments.
func New(users, postsPerUser, commentsPerPost int) *Store {
	s := &Store{usersByID: map[string]*User{}}
	for i := 0; i < users; i++ {
		u := &User{
			ID:     fmt.Sprintf("u%d", i),
			Name:   fmt.Sprintf("User %d", i),
			Email:  fmt.Sprintf("user%d@example.com", i),
			Age:    20 + i%50,
			Score:  float64(i%1000) + 0.5,
			Active: i%3 != 0,
		}
		if i%2 == 0 {
			bio := fmt.Sprintf("I am user %d, and I write about \"Go\" & <JSON>.", i)
			u.Bio = &bio
		}
		s.Users = append(s.Users, u)
		s.usersByID[u.ID] = u
	}
	for i, u := range s.Users {
		for j := 1; j <= 10; j++ {
			u.Friends = append(u.Friends, s.Users[(i+j*7)%users])
		}
		for j := 0; j < postsPerUser; j++ {
			p := &Post{
				ID:     fmt.Sprintf("p%d-%d", i, j),
				Title:  fmt.Sprintf("Post %d of %s", j, u.Name),
				Body:   strings.Repeat(fmt.Sprintf("The body of the post %d. ", j), 8),
				Tags:   []string{"go", "json", fmt.Sprintf("tag%d", j%5)},
				Likes:  (i * j) % 100,
				Author: u,
			}
			for k := 0; k < commentsPerPost; k++ {
				p.Comments = append(p.Comments, &Comment{
					ID:     fmt.Sprintf("c%d-%d-%d", i, j, k),
					Body:   fmt.Sprintf("Comment %d", k),
					Author: s.Users[(i+k+1)%users],
				})
			}
			u.Posts = append(u.Posts, p)
			s.Posts = append(s.Posts, p)
		}
	}
	return s
}

func (s *Store) User(id string) *User {
	return s.usersByID[id]
}

// Search returns the users whose name and the posts whose title have the text, up to first of them, or all of
// them if first is nil.
func (s *Store) Search(text string, firstArg *int) []SearchResult {
	first := len(s.Users) + len(s.Posts)
	if firstArg != nil {
		first = *firstArg
	}
	var results []SearchResult
	for _, u := range s.Users {
		if len(results) >= first {
			return results
		}
		if strings.Contains(u.Name, text) {
			results = append(results, u)
		}
		for _, p := range u.Posts {
			if len(results) >= first {
				return results
			}
			if strings.Contains(p.Title, text) {
				results = append(results, p)
			}
		}
	}
	return results
}

// First returns up to the first n of the list, or the whole list if n is nil: the argument first of a list.
func First[T any](list []T, first *int) []T {
	if first == nil {
		return list
	}
	n := max(*first, 0)
	if n < len(list) {
		return list[:n]
	}
	return list
}
