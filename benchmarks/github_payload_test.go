package benchmark

import (
	"fmt"
	"time"
)

// The payloads of the GitHub API ( see decode_github_test.go ): a page of 30 issues of a repository, by the REST
// API and by the GraphQL API. The keys and their order are the ones of the responses; the issues are mostly pull
// requests with markdown descriptions, some with labels, a milestone or reactions.

const (
	githubOwner = "example"
	githubRepo  = "json-codec"
	githubAPI   = "https://api.github.com"
	githubWeb   = "https://github.com"
)

var githubNow = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

type githubUserInfo struct {
	login string
	id    int64
}

var githubUsers = []githubUserInfo{{"maintainer", 209884}, {"contributor-one", 28623}, {"gopher42", 982358}, {"dependabot[bot]", 49699333}}

func githubRESTUser(u githubUserInfo) object {
	api := githubAPI + "/users/" + u.login
	typ := "User"
	if u.id == 49699333 {
		typ = "Bot"
	}
	return object{
		{"login", u.login},
		{"id", u.id},
		{"node_id", fmt.Sprintf("MDQ6VXNlcjE%d", u.id)},
		{"avatar_url", fmt.Sprintf("https://avatars.githubusercontent.com/u/%d?v=4", u.id)},
		{"gravatar_id", ""},
		{"url", api},
		{"html_url", githubWeb + "/" + u.login},
		{"followers_url", api + "/followers"},
		{"following_url", api + "/following{/other_user}"},
		{"gists_url", api + "/gists{/gist_id}"},
		{"starred_url", api + "/starred{/owner}{/repo}"},
		{"subscriptions_url", api + "/subscriptions"},
		{"organizations_url", api + "/orgs"},
		{"repos_url", api + "/repos"},
		{"events_url", api + "/events{/privacy}"},
		{"received_events_url", api + "/received_events"},
		{"type", typ},
		{"user_view_type", "public"},
		{"site_admin", false},
	}
}

var githubLabels = []struct{ name, color, description string }{
	{"bug", "d73a4a", "Something isn't working"},
	{"enhancement", "a2eeef", "New feature or request"},
	{"performance", "fbca04", "Faster encoding or decoding"},
	{"good first issue", "7057ff", "Good for newcomers"},
}

func githubRESTLabel(r *payloadRand, i int) object {
	l := githubLabels[i]
	return object{
		{"id", 1000000000 + int64(i)*7919},
		{"node_id", "LA_kwDOD" + r.token(base64URL, 14)},
		{"url", githubAPI + "/repos/" + githubOwner + "/" + githubRepo + "/labels/" + l.name},
		{"name", l.name},
		{"color", l.color},
		{"default", i < 2},
		{"description", l.description},
	}
}

// githubIssue is what the REST and the GraphQL payloads of an issue are made of.
type githubIssue struct {
	number                     int
	title, body                string
	author, closer             githubUserInfo
	pullRequest, closed        bool
	merged                     bool
	labels                     []int
	milestone                  bool
	comments                   int
	reactions                  [8]int
	created, updated, closedAt time.Time
}

func newGitHubIssues(r *payloadRand, bodyBytes, codeOneIn int) []githubIssue {
	issues := make([]githubIssue, 30)
	for i := range issues {
		is := &issues[i]
		is.number = 659 - i
		is.title = sentence(r)
		is.title = is.title[:len(is.title)-1]
		is.body = markdownWithCode(r, r.between(bodyBytes/10, bodyBytes*2), codeOneIn)
		is.author = githubUsers[r.intn(len(githubUsers))]
		is.closer = githubUsers[0]
		is.pullRequest = !r.chance(8)
		is.closed = !r.chance(6)
		is.merged = is.pullRequest && is.closed && !r.chance(5)
		for l := range githubLabels {
			if r.chance(4) {
				is.labels = append(is.labels, l)
			}
		}
		is.milestone = r.chance(5)
		is.comments = r.intn(4)
		for k := range is.reactions {
			if r.chance(6) {
				is.reactions[k] = r.between(1, 21)
			}
		}
		is.created = githubNow.Add(-time.Duration(r.between(3600, 90*86400)) * time.Second)
		is.updated = is.created.Add(time.Duration(r.between(60, 86400)) * time.Second)
		is.closedAt = is.updated
	}
	return issues
}

func githubRESTIssuesPayload() []byte {
	r := newPayloadRand(1)
	repo := githubAPI + "/repos/" + githubOwner + "/" + githubRepo
	var issues []any
	for _, is := range newGitHubIssues(r, 1600, 100) {
		url := fmt.Sprintf("%s/issues/%d", repo, is.number)
		kind := "issues"
		if is.pullRequest {
			kind = "pull"
		}
		labels := []any{}
		for _, l := range is.labels {
			labels = append(labels, githubRESTLabel(r, l))
		}
		var milestone any
		if is.milestone {
			milestone = object{
				{"url", repo + "/milestones/3"},
				{"html_url", githubWeb + "/" + githubOwner + "/" + githubRepo + "/milestone/3"},
				{"labels_url", repo + "/milestones/3/labels"},
				{"id", 11873412},
				{"node_id", "MI_kwDOD" + r.token(base64URL, 14)},
				{"number", 3},
				{"title", "v0.11.0"},
				{"description", sentence(r)},
				{"creator", githubRESTUser(githubUsers[0])},
				{"open_issues", 12},
				{"closed_issues", 41},
				{"state", "open"},
				{"created_at", timestamp(githubNow.AddDate(0, -4, 0))},
				{"updated_at", timestamp(githubNow.AddDate(0, 0, -2))},
				{"due_on", timestamp(githubNow.AddDate(0, 1, 0))},
				{"closed_at", nil},
			}
		}
		state, closedAt, closedBy, stateReason := "open", any(nil), any(nil), any(nil)
		if is.closed {
			state, closedAt, closedBy = "closed", timestamp(is.closedAt), githubRESTUser(is.closer)
			if !is.pullRequest {
				stateReason = "completed"
			}
		}
		o := object{
			{"url", url},
			{"repository_url", repo},
			{"labels_url", url + "/labels{/name}"},
			{"comments_url", url + "/comments"},
			{"events_url", url + "/events"},
			{"html_url", fmt.Sprintf("%s/%s/%s/%s/%d", githubWeb, githubOwner, githubRepo, kind, is.number)},
			{"id", 5523508452 + int64(is.number)*1537},
			{"node_id", "PR_kwDOD" + r.token(base64URL, 17)},
			{"number", is.number},
			{"title", is.title},
			{"user", githubRESTUser(is.author)},
			{"labels", labels},
			{"state", state},
			{"locked", false},
			{"assignees", []any{}},
			{"milestone", milestone},
			{"comments", is.comments},
			{"created_at", timestamp(is.created)},
			{"updated_at", timestamp(is.updated)},
			{"closed_at", closedAt},
			{"assignee", nil},
			{"author_association", "CONTRIBUTOR"},
			{"active_lock_reason", nil},
		}
		if is.pullRequest {
			var mergedAt any
			if is.merged {
				mergedAt = timestamp(is.closedAt)
			}
			pr := fmt.Sprintf("%s/%s/%s/pull/%d", githubWeb, githubOwner, githubRepo, is.number)
			o = append(o,
				field{"draft", false},
				field{"pull_request", object{
					{"url", fmt.Sprintf("%s/pulls/%d", repo, is.number)},
					{"html_url", pr},
					{"diff_url", pr + ".diff"},
					{"patch_url", pr + ".patch"},
					{"merged_at", mergedAt},
				}},
			)
		}
		total := 0
		for _, n := range is.reactions {
			total += n
		}
		o = append(o,
			field{"body", is.body},
			field{"closed_by", closedBy},
			field{"reactions", object{
				{"url", url + "/reactions"},
				{"total_count", total},
				{"+1", is.reactions[0]},
				{"-1", is.reactions[1]},
				{"laugh", is.reactions[2]},
				{"hooray", is.reactions[3]},
				{"confused", is.reactions[4]},
				{"heart", is.reactions[5]},
				{"rocket", is.reactions[6]},
				{"eyes", is.reactions[7]},
			}},
			field{"timeline_url", url + "/timeline"},
			field{"performed_via_github_app", nil},
			field{"state_reason", stateReason},
		)
		if !is.pullRequest {
			o = append(o,
				field{"sub_issues_summary", object{{"total", 0}, {"completed", 0}, {"percent_completed", 0}}},
				field{"issue_dependencies_summary", object{{"blocked_by", 0}, {"total_blocked_by", 0}, {"blocking", 0}, {"total_blocking", 0}}},
				field{"pinned_comment", nil},
			)
		}
		issues = append(issues, o)
	}
	return encodePayload(issues)
}

var githubReactionContents = []string{"THUMBS_UP", "THUMBS_DOWN", "LAUGH", "HOORAY", "CONFUSED", "HEART", "ROCKET", "EYES"}

func githubGraphQLIssuesPayload() []byte {
	r := newPayloadRand(2)
	var nodes []any
	for _, is := range newGitHubIssues(r, 1800, 3) {
		state, stateReason, closedAt := "OPEN", any(nil), any(nil)
		if is.closed {
			state, stateReason, closedAt = "CLOSED", "COMPLETED", timestamp(is.closedAt)
		}
		labels := []any{}
		for _, l := range is.labels {
			labels = append(labels, object{
				{"id", "LA_kwDOD" + r.token(base64URL, 14)},
				{"name", githubLabels[l].name},
				{"color", githubLabels[l].color},
				{"description", githubLabels[l].description},
			})
		}
		var milestone any
		if is.milestone {
			milestone = object{{"title", "v0.11.0"}, {"number", 3}, {"state", "OPEN"}, {"dueOn", timestamp(githubNow.AddDate(0, 1, 0))}}
		}
		reactions := make([]any, len(githubReactionContents))
		for k, content := range githubReactionContents {
			reactions[k] = object{{"content", content}, {"reactors", object{{"totalCount", is.reactions[k]}}}}
		}
		nodes = append(nodes, object{
			{"id", "I_kwDOD" + r.token(base64URL, 11)},
			{"number", is.number},
			{"title", is.title},
			{"body", is.body},
			{"state", state},
			{"stateReason", stateReason},
			{"url", fmt.Sprintf("%s/%s/%s/issues/%d", githubWeb, githubOwner, githubRepo, is.number)},
			{"createdAt", timestamp(is.created)},
			{"updatedAt", timestamp(is.updated)},
			{"closedAt", closedAt},
			{"locked", false},
			{"author", object{
				{"login", is.author.login},
				{"avatarUrl", fmt.Sprintf("https://avatars.githubusercontent.com/u/%d?u=%s&v=4", is.author.id, r.token(hexDigits, 40))},
				{"url", githubWeb + "/" + is.author.login},
			}},
			{"authorAssociation", "CONTRIBUTOR"},
			{"labels", object{{"nodes", labels}}},
			{"assignees", object{{"nodes", []any{}}}},
			{"milestone", milestone},
			{"comments", object{{"totalCount", is.comments}}},
			{"reactionGroups", reactions},
		})
	}
	return encodePayload(object{{"data", object{{"repository", object{
		{"name", githubRepo},
		{"nameWithOwner", githubOwner + "/" + githubRepo},
		{"description", "Fast JSON encoder and decoder compatible with encoding/json"},
		{"url", githubWeb + "/" + githubOwner + "/" + githubRepo},
		{"stargazerCount", 3708},
		{"forkCount", 220},
		{"isArchived", false},
		{"createdAt", "2020-04-19T09:32:36Z"},
		{"updatedAt", timestamp(githubNow)},
		{"primaryLanguage", object{{"name", "Go"}, {"color", "#00ADD8"}}},
		{"issues", object{
			{"totalCount", 269},
			{"pageInfo", object{{"hasNextPage", true}, {"endCursor", "Y3Vyc29yOnYyOpK5MjAyNi0wOS0yNFQx" + r.token(base64URL, 20)}}},
			{"nodes", nodes},
		}},
	}}}}})
}
