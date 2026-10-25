package benchmark

import (
	"reflect"
	"strings"
	"testing"
	"time"

	gojson "github.com/goccy/go-json"
)

// The structs of the payloads of fields without tags, with the tags of the names of their objects, which v2 matches
// as they are without an option: v1 matches a name to a field without a tag regardless of their case, which v2 does
// only with MatchCaseInsensitiveNames, which also ignores the dashes and the underscores of the names.

type smallPayloadTagged struct {
	St   int    `json:"st"`
	Sid  int    `json:"sid"`
	Tt   string `json:"tt"`
	Gr   int    `json:"gr"`
	Uuid string `json:"uuid"`
	Ip   string `json:"ip"`
	Ua   string `json:"ua"`
	Tz   int    `json:"tz"`
	V    int    `json:"v"`
}

type mediumPayloadTagged struct {
	Person  *cbPersonTagged `json:"person"`
	Company string          `json:"company"`
}

type cbPersonTagged struct {
	Name     *CBName           `json:"name"`
	Github   *cbGithubTagged   `json:"github"`
	Gravatar *cbGravatarTagged `json:"gravatar"`
}

type cbGithubTagged struct {
	Followers int `json:"followers"`
}

type cbGravatarTagged struct {
	Avatars []*cbAvatarTagged `json:"avatars"`
}

type cbAvatarTagged struct {
	Url string `json:"url"`
}

type largePayloadTagged struct {
	Users  []*dsUserTagged     `json:"users"`
	Topics *dsTopicsListTagged `json:"topics"`
}

type dsUserTagged struct {
	Username string `json:"username"`
}

// dsTopicsListTagged is DSTopicsList with the tags, but for MoreTopicsUrl: v1 matches no field to the name
// more_topics_url of the payload, which v2 matches to it only by MatchCaseInsensitiveNames, and so v2 matches none
// without the tag, decoding as v1 does.
type dsTopicsListTagged struct {
	Topics        []*dsTopicTagged `json:"topics"`
	MoreTopicsUrl string
}

type dsTopicTagged struct {
	Id   int    `json:"id"`
	Slug string `json:"slug"`
}

type gitHubGraphQLIssuesTagged struct {
	Data struct {
		Repository struct {
			Name            string    `json:"name"`
			NameWithOwner   string    `json:"nameWithOwner"`
			Description     string    `json:"description"`
			URL             string    `json:"url"`
			StargazerCount  int       `json:"stargazerCount"`
			ForkCount       int       `json:"forkCount"`
			IsArchived      bool      `json:"isArchived"`
			CreatedAt       time.Time `json:"createdAt"`
			UpdatedAt       time.Time `json:"updatedAt"`
			PrimaryLanguage *struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"primaryLanguage"`
			Issues struct {
				TotalCount int `json:"totalCount"`
				PageInfo   struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []gitHubGraphQLIssueTagged `json:"nodes"`
			} `json:"issues"`
		} `json:"repository"`
	} `json:"data"`
}

type gitHubGraphQLIssueTagged struct {
	ID          string     `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	State       string     `json:"state"`
	StateReason *string    `json:"stateReason"`
	URL         string     `json:"url"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	ClosedAt    *time.Time `json:"closedAt"`
	Locked      bool       `json:"locked"`
	Author      struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatarUrl"`
		URL       string `json:"url"`
	} `json:"author"`
	AuthorAssociation string `json:"authorAssociation"`
	Labels            struct {
		Nodes []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Color       string `json:"color"`
			Description string `json:"description"`
		} `json:"nodes"`
	} `json:"labels"`
	Assignees struct {
		Nodes []struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatarUrl"`
		} `json:"nodes"`
	} `json:"assignees"`
	Milestone *struct {
		Title  string     `json:"title"`
		Number int        `json:"number"`
		State  string     `json:"state"`
		DueOn  *time.Time `json:"dueOn"`
	} `json:"milestone"`
	Comments struct {
		TotalCount int `json:"totalCount"`
	} `json:"comments"`
	ReactionGroups []struct {
		Content  string `json:"content"`
		Reactors struct {
			TotalCount int `json:"totalCount"`
		} `json:"reactors"`
	} `json:"reactionGroups"`
}

var (
	streamSmallTaggedPayload = &streamPayload{
		data:     func() []byte { return SmallFixture },
		newValue: func() any { return &smallPayloadTagged{} },
	}
	streamMediumTaggedPayload = &streamPayload{
		data:     func() []byte { return MediumFixture },
		newValue: func() any { return &mediumPayloadTagged{} },
	}
)

// TestTaggedPayloadsV2 checks that the v2 decoding of the payloads into the structs with the tags gives the values
// which go-json's own gives into the structs without them, and that each tag is the name which v1 matches to the
// field in the payload, which a value of the payload that a field has without its name doesn't show: the benchmarks
// of the two decode the same.
func TestTaggedPayloadsV2(t *testing.T) {
	for _, tc := range []struct {
		name             string
		data             []byte
		untagged, tagged any
	}{
		{"small", SmallFixture, &SmallPayload{}, &smallPayloadTagged{}},
		{"medium", MediumFixture, &MediumPayload{}, &mediumPayloadTagged{}},
		{"large", LargeFixture, &LargePayload{}, &largePayloadTagged{}},
		{"github graphql", githubGraphQLIssues, &GitHubGraphQLIssues{}, &gitHubGraphQLIssuesTagged{}},
	} {
		if err := gojson.Unmarshal(tc.data, tc.untagged); err != nil {
			t.Fatal(err)
		}
		if err := unmarshalV2(tc.data, tc.tagged); err != nil {
			t.Fatal(err)
		}
		if !sameFields(reflect.ValueOf(tc.untagged), reflect.ValueOf(tc.tagged)) {
			t.Errorf("%s: the payload decodes differently", tc.name)
		}
		var payload any
		if err := gojson.Unmarshal(tc.data, &payload); err != nil {
			t.Fatal(err)
		}
		names := map[string]bool{}
		collectNames(payload, names)
		sameNames(t, tc.name, reflect.TypeOf(tc.untagged), reflect.TypeOf(tc.tagged), names, map[reflect.Type]bool{})
	}
}

// collectNames puts the names of the objects of v, a value which an interface{} is decoded into, into names.
func collectNames(v any, names map[string]bool) {
	switch v := v.(type) {
	case map[string]any:
		for name, value := range v {
			names[name] = true
			collectNames(value, names)
		}
	case []any:
		for _, value := range v {
			collectNames(value, names)
		}
	}
}

// sameNames checks that the name of each field of the struct tagged, of the tag or of the field, is the name of the
// field of untagged, a struct of the same fields, by which v1 matches the names regardless of their case, and that
// the names of the payload which v1 matches so are the name of the field of tagged as they are, in the structs which
// the fields are or point to too, but not in a struct of the same type in both, as time.Time and CBName are, whose
// values sameFields compares.
func sameNames(
	t *testing.T, payload string, untagged, tagged reflect.Type, names map[string]bool, done map[reflect.Type]bool,
) {
	t.Helper()
	for untagged.Kind() == reflect.Pointer || untagged.Kind() == reflect.Slice {
		untagged, tagged = untagged.Elem(), tagged.Elem()
	}
	if untagged.Kind() != reflect.Struct || untagged == tagged || done[tagged] {
		return
	}
	done[tagged] = true
	for i := 0; i < untagged.NumField(); i++ {
		u, g := untagged.Field(i), tagged.Field(i)
		v1Name, name := fieldName(u), fieldName(g)
		if !strings.EqualFold(v1Name, name) {
			t.Errorf("%s: field %s of %s has the name %q, which v1 doesn't match to %q", payload, g.Name, tagged, name, v1Name)
		}
		for other := range names {
			if strings.EqualFold(other, v1Name) && other != name {
				t.Errorf("%s: field %s of %s has the name %q, which v2 doesn't match to %q of the payload",
					payload, g.Name, tagged, name, other)
			}
		}
		sameNames(t, payload, u.Type, g.Type, names, done)
	}
}

// fieldName returns the name of the field, of its tag or else its own.
func fieldName(f reflect.StructField) string {
	if name, _, _ := strings.Cut(f.Tag.Get("json"), ","); name != "" {
		return name
	}
	return f.Name
}

// sameFields reports whether a and b, of types which differ by their names and the tags of their fields only, hold
// the same values, field by field.
func sameFields(a, b reflect.Value) bool {
	if a.Kind() != b.Kind() {
		return false
	}
	switch a.Kind() {
	case reflect.Pointer:
		if a.IsNil() || b.IsNil() {
			return a.IsNil() == b.IsNil()
		}
		return sameFields(a.Elem(), b.Elem())
	case reflect.Slice:
		if a.IsNil() != b.IsNil() || a.Len() != b.Len() {
			return false
		}
		for i := 0; i < a.Len(); i++ {
			if !sameFields(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		if a.Type() == b.Type() {
			return reflect.DeepEqual(a.Interface(), b.Interface())
		}
		if a.NumField() != b.NumField() {
			return false
		}
		for i := 0; i < a.NumField(); i++ {
			if !sameFields(a.Field(i), b.Field(i)) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(a.Interface(), b.Interface())
}
