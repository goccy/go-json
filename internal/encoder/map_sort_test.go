package encoder

import (
	"math/rand"
	"slices"
	"strconv"
	"testing"
)

// The keys of the objects of the Twitter payload of the benchmarks, which most of the keys of a JSON object are
// like: many share their first bytes.
var twitterObjectKeys = [][]string{
	{"max_id", "since_id", "refresh_url", "next_results", "count", "completed_in", "since_id_str", "query", "max_id_str"},
	{"coordinates", "favorited", "truncated", "created_at", "id_str", "entities", "in_reply_to_user_id_str", "contributors",
		"text", "metadata", "retweet_count", "in_reply_to_status_id_str", "id", "geo", "retweeted", "in_reply_to_user_id",
		"place", "user", "in_reply_to_screen_name", "source", "in_reply_to_status_id"},
	{"profile_sidebar_fill_color", "profile_sidebar_border_color", "profile_background_tile", "name", "profile_image_url",
		"created_at", "location", "follow_request_sent", "profile_link_color", "is_translator", "id_str", "entities",
		"default_profile", "contributors_enabled", "favourites_count", "url", "profile_image_url_https", "utc_offset", "id",
		"profile_use_background_image", "listed_count", "profile_text_color", "lang", "followers_count", "protected",
		"notifications", "profile_background_image_url_https", "profile_background_color", "verified", "geo_enabled",
		"time_zone", "description", "default_profile_image", "profile_background_image_url", "statuses_count",
		"friends_count", "following", "show_all_inline_media", "screen_name"},
}

// sortKeysInputs returns key sets of every size, in random orders as a map is ranged over: the keys of the Twitter
// payload, keys which share their first eight bytes or all of them, keys of zeros, and many keys.
func sortKeysInputs() [][]string {
	r := rand.New(rand.NewSource(1))
	var sets [][]string
	sets = append(sets, twitterObjectKeys...)
	for _, n := range []int{0, 1, 2, 3, 15, 16, 17, 40, 100, 1000} {
		var same, prefixed, zeros []string
		for i := 0; i < n; i++ {
			same = append(same, "key_of_the_same_prefix_"+strconv.Itoa(r.Intn(1<<20)))
			prefixed = append(prefixed, strconv.Itoa(r.Intn(1<<30)))
			zeros = append(zeros, string(make([]byte, r.Intn(12)))+strconv.Itoa(i))
		}
		sets = append(sets, same, prefixed, zeros)
	}
	var inputs [][]string
	for _, set := range sets {
		for i := 0; i < 4; i++ {
			keys := slices.Clone(set)
			r.Shuffle(len(keys), func(a, b int) { keys[a], keys[b] = keys[b], keys[a] })
			inputs = append(inputs, keys)
		}
	}
	return inputs
}

// SortKeys puts the entries in the order of their keys, as a sort of the strings does, for keys of any size and
// any first bytes, in any order.
func TestSortKeys(t *testing.T) {
	for _, keys := range sortKeysInputs() {
		c := &MapContext{Keys: keys}
		c.SortKeys()
		got := make([]string, len(keys))
		for i, e := range c.Order {
			got[i] = keys[e]
		}
		want := slices.Clone(keys)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("keys %q:\n got %q\nwant %q", keys, got, want)
		}
	}
}

func BenchmarkSortKeys(b *testing.B) {
	r := rand.New(rand.NewSource(2))
	for i, keys := range twitterObjectKeys {
		var orders [][]string
		for j := 0; j < 8; j++ {
			k := slices.Clone(keys)
			r.Shuffle(len(k), func(a, b int) { k[a], k[b] = k[b], k[a] })
			orders = append(orders, k)
		}
		b.Run("twitter"+strconv.Itoa(i)+"/"+strconv.Itoa(len(keys)), func(b *testing.B) {
			c := &MapContext{}
			for n := 0; n < b.N; n++ {
				c.Keys = orders[n%len(orders)]
				c.SortKeys()
			}
		})
	}
}
