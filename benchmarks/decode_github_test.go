package benchmark

import (
	stdjson "encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
)

// The decode benchmarks of the responses of the GitHub API, which is one of the most used JSON APIs: the list of
// the issues of a repository by the REST API ( GET /repos/{owner}/{repo}/issues ) and by the GraphQL API.
//
// The payloads are generated ( see github_payload_test.go ): the keys, their order, the types and the nulls are
// the ones of the responses, and the titles and the descriptions are markdown with lists, code and escapes.
//
// The REST types are the ones of the most used client, github.com/google/go-github: fields of pointers, times
// by a type with UnmarshalJSON, and keys of the response which no field has, as the API adds keys. The GraphQL
// types have no tags, as the ones of github.com/shurcooL/githubv4: their fields match the camel case keys by
// case folding.

// GitHubTimestamp is a time of the REST API, as go-github decodes it: an RFC 3339 string or Unix seconds.
type GitHubTimestamp struct {
	time.Time
}

func (t *GitHubTimestamp) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] != '"' {
		var sec int64
		if err := stdjson.Unmarshal(data, &sec); err != nil {
			return err
		}
		t.Time = time.Unix(sec, 0)
		return nil
	}
	var err error
	t.Time, err = time.Parse(`"`+time.RFC3339+`"`, string(data))
	return err
}

type GitHubUser struct {
	Login             *string `json:"login,omitempty"`
	ID                *int64  `json:"id,omitempty"`
	NodeID            *string `json:"node_id,omitempty"`
	AvatarURL         *string `json:"avatar_url,omitempty"`
	HTMLURL           *string `json:"html_url,omitempty"`
	GravatarID        *string `json:"gravatar_id,omitempty"`
	Type              *string `json:"type,omitempty"`
	SiteAdmin         *bool   `json:"site_admin,omitempty"`
	URL               *string `json:"url,omitempty"`
	EventsURL         *string `json:"events_url,omitempty"`
	FollowingURL      *string `json:"following_url,omitempty"`
	FollowersURL      *string `json:"followers_url,omitempty"`
	GistsURL          *string `json:"gists_url,omitempty"`
	OrganizationsURL  *string `json:"organizations_url,omitempty"`
	ReceivedEventsURL *string `json:"received_events_url,omitempty"`
	ReposURL          *string `json:"repos_url,omitempty"`
	StarredURL        *string `json:"starred_url,omitempty"`
	SubscriptionsURL  *string `json:"subscriptions_url,omitempty"`
}

type GitHubLabel struct {
	ID          *int64  `json:"id,omitempty"`
	URL         *string `json:"url,omitempty"`
	Name        *string `json:"name,omitempty"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
	Default     *bool   `json:"default,omitempty"`
	NodeID      *string `json:"node_id,omitempty"`
}

type GitHubMilestone struct {
	URL          *string          `json:"url,omitempty"`
	HTMLURL      *string          `json:"html_url,omitempty"`
	LabelsURL    *string          `json:"labels_url,omitempty"`
	ID           *int64           `json:"id,omitempty"`
	Number       *int             `json:"number,omitempty"`
	State        *string          `json:"state,omitempty"`
	Title        *string          `json:"title,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Creator      *GitHubUser      `json:"creator,omitempty"`
	OpenIssues   *int             `json:"open_issues,omitempty"`
	ClosedIssues *int             `json:"closed_issues,omitempty"`
	CreatedAt    *GitHubTimestamp `json:"created_at,omitempty"`
	UpdatedAt    *GitHubTimestamp `json:"updated_at,omitempty"`
	ClosedAt     *GitHubTimestamp `json:"closed_at,omitempty"`
	DueOn        *GitHubTimestamp `json:"due_on,omitempty"`
	NodeID       *string          `json:"node_id,omitempty"`
}

type GitHubPullRequestLinks struct {
	URL      *string          `json:"url,omitempty"`
	HTMLURL  *string          `json:"html_url,omitempty"`
	DiffURL  *string          `json:"diff_url,omitempty"`
	PatchURL *string          `json:"patch_url,omitempty"`
	MergedAt *GitHubTimestamp `json:"merged_at,omitempty"`
}

type GitHubReactions struct {
	TotalCount *int    `json:"total_count,omitempty"`
	PlusOne    *int    `json:"+1,omitempty"`
	MinusOne   *int    `json:"-1,omitempty"`
	Laugh      *int    `json:"laugh,omitempty"`
	Confused   *int    `json:"confused,omitempty"`
	Heart      *int    `json:"heart,omitempty"`
	Hooray     *int    `json:"hooray,omitempty"`
	Rocket     *int    `json:"rocket,omitempty"`
	Eyes       *int    `json:"eyes,omitempty"`
	URL        *string `json:"url,omitempty"`
}

type GitHubIssue struct {
	ID                *int64                  `json:"id,omitempty"`
	Number            *int                    `json:"number,omitempty"`
	State             *string                 `json:"state,omitempty"`
	StateReason       *string                 `json:"state_reason,omitempty"`
	Locked            *bool                   `json:"locked,omitempty"`
	Title             *string                 `json:"title,omitempty"`
	Body              *string                 `json:"body,omitempty"`
	AuthorAssociation *string                 `json:"author_association,omitempty"`
	User              *GitHubUser             `json:"user,omitempty"`
	Labels            []*GitHubLabel          `json:"labels,omitempty"`
	Assignee          *GitHubUser             `json:"assignee,omitempty"`
	Comments          *int                    `json:"comments,omitempty"`
	ClosedAt          *GitHubTimestamp        `json:"closed_at,omitempty"`
	CreatedAt         *GitHubTimestamp        `json:"created_at,omitempty"`
	UpdatedAt         *GitHubTimestamp        `json:"updated_at,omitempty"`
	ClosedBy          *GitHubUser             `json:"closed_by,omitempty"`
	URL               *string                 `json:"url,omitempty"`
	HTMLURL           *string                 `json:"html_url,omitempty"`
	CommentsURL       *string                 `json:"comments_url,omitempty"`
	EventsURL         *string                 `json:"events_url,omitempty"`
	LabelsURL         *string                 `json:"labels_url,omitempty"`
	RepositoryURL     *string                 `json:"repository_url,omitempty"`
	Milestone         *GitHubMilestone        `json:"milestone,omitempty"`
	PullRequestLinks  *GitHubPullRequestLinks `json:"pull_request,omitempty"`
	Reactions         *GitHubReactions        `json:"reactions,omitempty"`
	Assignees         []*GitHubUser           `json:"assignees,omitempty"`
	NodeID            *string                 `json:"node_id,omitempty"`
	Draft             *bool                   `json:"draft,omitempty"`
	ActiveLockReason  *string                 `json:"active_lock_reason,omitempty"`
}

type GitHubGraphQLIssues struct {
	Data struct {
		Repository struct {
			Name            string
			NameWithOwner   string
			Description     string
			URL             string
			StargazerCount  int
			ForkCount       int
			IsArchived      bool
			CreatedAt       time.Time
			UpdatedAt       time.Time
			PrimaryLanguage *struct {
				Name  string
				Color string
			}
			Issues struct {
				TotalCount int
				PageInfo   struct {
					HasNextPage bool
					EndCursor   string
				}
				Nodes []GitHubGraphQLIssue
			}
		}
	}
}

type GitHubGraphQLIssue struct {
	ID          string
	Number      int
	Title       string
	Body        string
	State       string
	StateReason *string
	URL         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ClosedAt    *time.Time
	Locked      bool
	Author      struct {
		Login     string
		AvatarURL string
		URL       string
	}
	AuthorAssociation string
	Labels            struct {
		Nodes []struct {
			ID          string
			Name        string
			Color       string
			Description string
		}
	}
	Assignees struct {
		Nodes []struct {
			Login     string
			AvatarURL string
		}
	}
	Milestone *struct {
		Title  string
		Number int
		State  string
		DueOn  *time.Time
	}
	Comments struct {
		TotalCount int
	}
	ReactionGroups []struct {
		Content  string
		Reactors struct {
			TotalCount int
		}
	}
}

var (
	githubRESTIssues    = githubRESTIssuesPayload()
	githubGraphQLIssues = githubGraphQLIssuesPayload()
)

func init() {
	for _, typ := range []reflect.Type{reflect.TypeOf([]*GitHubIssue{}), reflect.TypeOf(GitHubGraphQLIssues{})} {
		if err := sonic.Pretouch(typ); err != nil {
			panic(err)
		}
	}
}

func TestGitHubPayloads(t *testing.T) {
	// go-json decodes the payloads as encoding/json does.
	var restWant, restGot []*GitHubIssue
	if err := stdjson.Unmarshal(githubRESTIssues, &restWant); err != nil {
		t.Fatal(err)
	}
	if err := gojson.Unmarshal(githubRESTIssues, &restGot); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restGot, restWant) {
		t.Fatal("the REST payload decodes differently")
	}
	var graphqlWant, graphqlGot GitHubGraphQLIssues
	if err := stdjson.Unmarshal(githubGraphQLIssues, &graphqlWant); err != nil {
		t.Fatal(err)
	}
	if err := gojson.Unmarshal(githubGraphQLIssues, &graphqlGot); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(graphqlGot, graphqlWant) {
		t.Fatal("the GraphQL payload decodes differently")
	}
	if len(graphqlGot.Data.Repository.Issues.Nodes) == 0 || graphqlGot.Data.Repository.NameWithOwner == "" {
		t.Fatal("the GraphQL fields are not matched by case folding")
	}
}

func Benchmark_Decode_GitHubREST_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[[]*GitHubIssue](b, githubRESTIssues, stdjson.Unmarshal)
}

func Benchmark_Decode_GitHubREST_Unmarshal_GoJson(b *testing.B) {
	benchDecode[[]*GitHubIssue](b, githubRESTIssues, gojson.Unmarshal)
}

func Benchmark_Decode_GitHubREST_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[[]*GitHubIssue](b, githubRESTIssues)
}

func Benchmark_Decode_GitHubREST_Unmarshal_Sonic(b *testing.B) {
	benchDecode[[]*GitHubIssue](b, githubRESTIssues, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_GitHubREST_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[[]*GitHubIssue](b, githubRESTIssues, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[GitHubGraphQLIssues](b, githubGraphQLIssues, stdjson.Unmarshal)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_GoJson(b *testing.B) {
	benchDecode[GitHubGraphQLIssues](b, githubGraphQLIssues, gojson.Unmarshal)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[GitHubGraphQLIssues](b, githubGraphQLIssues)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_Sonic(b *testing.B) {
	benchDecode[GitHubGraphQLIssues](b, githubGraphQLIssues, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[GitHubGraphQLIssues](b, githubGraphQLIssues, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_GitHubREST_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[[]*GitHubIssue](b, githubRESTIssues)
}

func Benchmark_Decode_GitHubREST_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeSonicFastest[[]*GitHubIssue](b, githubRESTIssues)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[GitHubGraphQLIssues](b, githubGraphQLIssues)
}

func Benchmark_Decode_GitHubGraphQL_Unmarshal_SonicFastest(b *testing.B) {
	benchDecodeSonicFastest[GitHubGraphQLIssues](b, githubGraphQLIssues)
}
