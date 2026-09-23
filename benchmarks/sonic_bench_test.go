/*
 * Copyright 2021 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// The benchmarks of the encoder of bytedance/sonic, with go-json beside it.
//
// The payload ( TwitterJson and the types of TwitterStruct ) is testdata/twitter.go of sonic v1.15.4, copied
// as it is: a directory named testdata can't be imported. The benchmarks are the ones of sonic's
// encoder/encoder_test.go, written for the libraries compared here. Both are the work of ByteDance Inc. under
// the Apache License 2.0, whose text is in licenses/sonic-LICENSE; the rest of this repository is under its
// own license ( see LICENSE at the root ).
//
// The benchmarks encode the payload as a value of interface{} ( Generic ) and as its struct ( Binding ),
// one goroutine and in parallel:
//
//   - Sonic is what sonic's BenchmarkEncoder_*_Sonic does: SortMapKeys, EscapeHTML and CompactMarshaler,
//     which is what go-json does by default, but for the normalization of UTF-8, which go-json also does by
//     default: GoJsonNoNormalize is go-json without it.
//   - SonicFast is what BenchmarkEncoder_*_Sonic_Fast does: no option. GoJsonLikeSonicFast is go-json with
//     the options which do the same: no sort of the keys, no escape of HTML, no normalization of UTF-8.
//   - StdLib is encoding/json, as in sonic.

package benchmark

import (
	stdjson "encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/option"
	gojson "github.com/goccy/go-json"
)

var (
	twitterGeneric     any
	twitterBinding     TwitterStruct
	sonicTwitter       = sonic.Config{SortMapKeys: true, EscapeHTML: true, CompactMarshaler: true}.Froze()
	sonicTwitterFast   = sonic.Config{}.Froze()
	noNormalizeOptions = []gojson.EncodeOptionFunc{gojson.DisableNormalizeUTF8()}
)

func init() {
	if err := stdjson.Unmarshal([]byte(TwitterJson), &twitterGeneric); err != nil {
		panic(err)
	}
	if err := stdjson.Unmarshal([]byte(TwitterJson), &twitterBinding); err != nil {
		panic(err)
	}
	opts := []option.CompileOption{option.WithCompileRecursiveDepth(10)}
	if depth, err := strconv.Atoi(os.Getenv("SONIC_MAX_INLINE_DEPTH")); err == nil {
		opts = append(opts, option.WithCompileMaxInlineDepth(depth))
	}
	if err := sonic.Pretouch(reflect.TypeOf(&twitterBinding), opts...); err != nil {
		panic(err)
	}
}

func benchTwitter(b *testing.B, marshal func() ([]byte, error)) {
	if _, err := marshal(); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(TwitterJson)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshal(); err != nil {
			b.Fatal(err)
		}
	}
}

func benchTwitterParallel(b *testing.B, marshal func() ([]byte, error)) {
	if _, err := marshal(); err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(TwitterJson)))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := marshal(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func Benchmark_TwitterGeneric_Sonic(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitter.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_SonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_StdLib(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return stdjson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_GoJson(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterGeneric_GoJsonNoNormalize(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(twitterGeneric, noNormalizeOptions...) })
}

func Benchmark_TwitterGeneric_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(twitterGeneric, likeSonicOptions...) })
}

func Benchmark_TwitterBinding_Sonic(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitter.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_SonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_StdLib(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return stdjson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_GoJson(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterBinding_GoJsonNoNormalize(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(&twitterBinding, noNormalizeOptions...) })
}

func Benchmark_TwitterBinding_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitter(b, func() ([]byte, error) { return gojson.MarshalWithOption(&twitterBinding, likeSonicOptions...) })
}

func Benchmark_TwitterParallelGeneric_Sonic(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitter.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_SonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_StdLib(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return stdjson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_GoJson(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.Marshal(twitterGeneric) })
}

func Benchmark_TwitterParallelGeneric_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.MarshalWithOption(twitterGeneric, likeSonicOptions...) })
}

func Benchmark_TwitterParallelBinding_Sonic(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitter.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_SonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return sonicTwitterFast.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_StdLib(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return stdjson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_GoJson(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.Marshal(&twitterBinding) })
}

func Benchmark_TwitterParallelBinding_GoJsonLikeSonicFast(b *testing.B) {
	benchTwitterParallel(b, func() ([]byte, error) { return gojson.MarshalWithOption(&twitterBinding, likeSonicOptions...) })
}

// what go-json writes for the payload is what encoding/json writes, and what sonic writes with the options
// of the benchmark.
func TestTwitterPayload(t *testing.T) {
	for _, v := range []any{twitterGeneric, &twitterBinding} {
		expected, err := stdjson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := gojson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(expected) {
			t.Fatalf("%T: go-json differs from encoding/json", v)
		}
		got, err = sonicTwitter.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(expected) {
			t.Logf("%T: sonic differs from encoding/json at byte %d", v, firstDifference(got, expected))
		}
	}
}

func firstDifference(a, b []byte) int {
	for i := range a {
		if i >= len(b) || a[i] != b[i] {
			return i
		}
	}
	return len(a)
}

const TwitterJson = `{
"statuses": [
	{
	"coordinates": null,
	"favorited": false,
	"truncated": false,
	"created_at": "Mon Sep 24 03:35:21 +0000 2012",
	"id_str": "250075927172759552",
	"entities": {
		"urls": [

		],
		"hashtags": [
		{
			"text": "freebandnames",
			"indices": [
			20,
			34
			]
		}
		],
		"user_mentions": [

		]
	},
	"in_reply_to_user_id_str": null,
	"contributors": null,
	"text": "Aggressive Ponytail #freebandnames",
	"metadata": {
		"iso_language_code": "en",
		"result_type": "recent"
	},
	"retweet_count": 0,
	"in_reply_to_status_id_str": null,
	"id": 250075927172759552,
	"geo": null,
	"retweeted": false,
	"in_reply_to_user_id": null,
	"place": null,
	"user": {
		"profile_sidebar_fill_color": "DDEEF6",
		"profile_sidebar_border_color": "C0DEED",
		"profile_background_tile": false,
		"name": "Sean Cummings",
		"profile_image_url": "https://a0.twimg.com/profile_images/2359746665/1v6zfgqo8g0d3mk7ii5s_normal.jpeg",
		"created_at": "Mon Apr 26 06:01:55 +0000 2010",
		"location": "LA, CA",
		"follow_request_sent": null,
		"profile_link_color": "0084B4",
		"is_translator": false,
		"id_str": "137238150",
		"entities": {
		"url": {
			"urls": [
			{
				"expanded_url": null,
				"url": "",
				"indices": [
				0,
				0
				]
			}
			]
		},
		"description": {
			"urls": [

			]
		}
		},
		"default_profile": true,
		"contributors_enabled": false,
		"favourites_count": 0,
		"url": null,
		"profile_image_url_https": "https://si0.twimg.com/profile_images/2359746665/1v6zfgqo8g0d3mk7ii5s_normal.jpeg",
		"utc_offset": -28800,
		"id": 137238150,
		"profile_use_background_image": true,
		"listed_count": 2,
		"profile_text_color": "333333",
		"lang": "en",
		"followers_count": 70,
		"protected": false,
		"notifications": null,
		"profile_background_image_url_https": "https://si0.twimg.com/images/themes/theme1/bg.png",
		"profile_background_color": "C0DEED",
		"verified": false,
		"geo_enabled": true,
		"time_zone": "Pacific Time (US & Canada)",
		"description": "Born 330 Live 310",
		"default_profile_image": false,
		"profile_background_image_url": "https://a0.twimg.com/images/themes/theme1/bg.png",
		"statuses_count": 579,
		"friends_count": 110,
		"following": null,
		"show_all_inline_media": false,
		"screen_name": "sean_cummings"
	},
	"in_reply_to_screen_name": null,
	"source": "<a href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for Mac</a>",
	"in_reply_to_status_id": null
	},
	{
	"coordinates": null,
	"favorited": false,
	"truncated": false,
	"created_at": "Fri Sep 21 23:40:54 +0000 2012",
	"id_str": "249292149810667520",
	"entities": {
		"urls": [

		],
		"hashtags": [
		{
			"text": "FreeBandNames",
			"indices": [
			20,
			34
			]
		}
		],
		"user_mentions": [

		]
	},
	"in_reply_to_user_id_str": null,
	"contributors": null,
	"text": "Thee Namaste Nerdz. #FreeBandNames",
	"metadata": {
		"iso_language_code": "pl",
		"result_type": "recent"
	},
	"retweet_count": 0,
	"in_reply_to_status_id_str": null,
	"id": 249292149810667520,
	"geo": null,
	"retweeted": false,
	"in_reply_to_user_id": null,
	"place": null,
	"user": {
		"profile_sidebar_fill_color": "DDFFCC",
		"profile_sidebar_border_color": "BDDCAD",
		"profile_background_tile": true,
		"name": "Chaz Martenstein",
		"profile_image_url": "https://a0.twimg.com/profile_images/447958234/Lichtenstein_normal.jpg",
		"created_at": "Tue Apr 07 19:05:07 +0000 2009",
		"location": "Durham, NC",
		"follow_request_sent": null,
		"profile_link_color": "0084B4",
		"is_translator": false,
		"id_str": "29516238",
		"entities": {
		"url": {
			"urls": [
			{
				"expanded_url": null,
				"url": "https://bullcityrecords.com/wnng/",
				"indices": [
				0,
				32
				]
			}
			]
		},
		"description": {
			"urls": [

			]
		}
		},
		"default_profile": false,
		"contributors_enabled": false,
		"favourites_count": 8,
		"url": "https://bullcityrecords.com/wnng/",
		"profile_image_url_https": "https://si0.twimg.com/profile_images/447958234/Lichtenstein_normal.jpg",
		"utc_offset": -18000,
		"id": 29516238,
		"profile_use_background_image": true,
		"listed_count": 118,
		"profile_text_color": "333333",
		"lang": "en",
		"followers_count": 2052,
		"protected": false,
		"notifications": null,
		"profile_background_image_url_https": "https://si0.twimg.com/profile_background_images/9423277/background_tile.bmp",
		"profile_background_color": "9AE4E8",
		"verified": false,
		"geo_enabled": false,
		"time_zone": "Eastern Time (US & Canada)",
		"description": "You will come to Durham, North Carolina. I will sell you some records then, here in Durham, North Carolina. Fun will happen.",
		"default_profile_image": false,
		"profile_background_image_url": "https://a0.twimg.com/profile_background_images/9423277/background_tile.bmp",
		"statuses_count": 7579,
		"friends_count": 348,
		"following": null,
		"show_all_inline_media": true,
		"screen_name": "bullcityrecords"
	},
	"in_reply_to_screen_name": null,
	"source": "web",
	"in_reply_to_status_id": null
	},
	{
	"coordinates": null,
	"favorited": false,
	"truncated": false,
	"created_at": "Fri Sep 21 23:30:20 +0000 2012",
	"id_str": "249289491129438208",
	"entities": {
		"urls": [

		],
		"hashtags": [
		{
			"text": "freebandnames",
			"indices": [
			29,
			43
			]
		}
		],
		"user_mentions": [

		]
	},
	"in_reply_to_user_id_str": null,
	"contributors": null,
	"text": "Mexican Heaven, Mexican Hell #freebandnames",
	"metadata": {
		"iso_language_code": "en",
		"result_type": "recent"
	},
	"retweet_count": 0,
	"in_reply_to_status_id_str": null,
	"id": 249289491129438208,
	"geo": null,
	"retweeted": false,
	"in_reply_to_user_id": null,
	"place": null,
	"user": {
		"profile_sidebar_fill_color": "99CC33",
		"profile_sidebar_border_color": "829D5E",
		"profile_background_tile": false,
		"name": "Thomas John Wakeman",
		"profile_image_url": "https://a0.twimg.com/profile_images/2219333930/Froggystyle_normal.png",
		"created_at": "Tue Sep 01 21:21:35 +0000 2009",
		"location": "Kingston New York",
		"follow_request_sent": null,
		"profile_link_color": "D02B55",
		"is_translator": false,
		"id_str": "70789458",
		"entities": {
		"url": {
			"urls": [
			{
				"expanded_url": null,
				"url": "",
				"indices": [
				0,
				0
				]
			}
			]
		},
		"description": {
			"urls": [

			]
		}
		},
		"default_profile": false,
		"contributors_enabled": false,
		"favourites_count": 19,
		"url": null,
		"profile_image_url_https": "https://si0.twimg.com/profile_images/2219333930/Froggystyle_normal.png",
		"utc_offset": -18000,
		"id": 70789458,
		"profile_use_background_image": true,
		"listed_count": 1,
		"profile_text_color": "3E4415",
		"lang": "en",
		"followers_count": 63,
		"protected": false,
		"notifications": null,
		"profile_background_image_url_https": "https://si0.twimg.com/images/themes/theme5/bg.gif",
		"profile_background_color": "352726",
		"verified": false,
		"geo_enabled": false,
		"time_zone": "Eastern Time (US & Canada)",
		"description": "Science Fiction Writer, sort of. Likes Superheroes, Mole People, Alt. Timelines.",
		"default_profile_image": false,
		"profile_background_image_url": "https://a0.twimg.com/images/themes/theme5/bg.gif",
		"statuses_count": 1048,
		"friends_count": 63,
		"following": null,
		"show_all_inline_media": false,
		"screen_name": "MonkiesFist"
	},
	"in_reply_to_screen_name": null,
	"source": "web",
	"in_reply_to_status_id": null
	},
	{
	"coordinates": null,
	"favorited": false,
	"truncated": false,
	"created_at": "Fri Sep 21 22:51:18 +0000 2012",
	"id_str": "249279667666817024",
	"entities": {
		"urls": [

		],
		"hashtags": [
		{
			"text": "freebandnames",
			"indices": [
			20,
			34
			]
		}
		],
		"user_mentions": [

		]
	},
	"in_reply_to_user_id_str": null,
	"contributors": null,
	"text": "The Foolish Mortals #freebandnames",
	"metadata": {
		"iso_language_code": "en",
		"result_type": "recent"
	},
	"retweet_count": 0,
	"in_reply_to_status_id_str": null,
	"id": 249279667666817024,
	"geo": null,
	"retweeted": false,
	"in_reply_to_user_id": null,
	"place": null,
	"user": {
		"profile_sidebar_fill_color": "BFAC83",
		"profile_sidebar_border_color": "615A44",
		"profile_background_tile": true,
		"name": "Marty Elmer",
		"profile_image_url": "https://a0.twimg.com/profile_images/1629790393/shrinker_2000_trans_normal.png",
		"created_at": "Mon May 04 00:05:00 +0000 2009",
		"location": "Wisconsin, USA",
		"follow_request_sent": null,
		"profile_link_color": "3B2A26",
		"is_translator": false,
		"id_str": "37539828",
		"entities": {
		"url": {
			"urls": [
			{
				"expanded_url": null,
				"url": "https://www.omnitarian.me",
				"indices": [
				0,
				24
				]
			}
			]
		},
		"description": {
			"urls": [

			]
		}
		},
		"default_profile": false,
		"contributors_enabled": false,
		"favourites_count": 647,
		"url": "https://www.omnitarian.me",
		"profile_image_url_https": "https://si0.twimg.com/profile_images/1629790393/shrinker_2000_trans_normal.png",
		"utc_offset": -21600,
		"id": 37539828,
		"profile_use_background_image": true,
		"listed_count": 52,
		"profile_text_color": "000000",
		"lang": "en",
		"followers_count": 608,
		"protected": false,
		"notifications": null,
		"profile_background_image_url_https": "https://si0.twimg.com/profile_background_images/106455659/rect6056-9.png",
		"profile_background_color": "EEE3C4",
		"verified": false,
		"geo_enabled": false,
		"time_zone": "Central Time (US & Canada)",
		"description": "Cartoonist, Illustrator, and T-Shirt connoisseur",
		"default_profile_image": false,
		"profile_background_image_url": "https://a0.twimg.com/profile_background_images/106455659/rect6056-9.png",
		"statuses_count": 3575,
		"friends_count": 249,
		"following": null,
		"show_all_inline_media": true,
		"screen_name": "Omnitarian"
	},
	"in_reply_to_screen_name": null,
	"source": "<a href=\"//twitter.com/download/iphone%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for iPhone</a>",
	"in_reply_to_status_id": null
	}
],
"search_metadata": {
	"max_id": 250126199840518145,
	"since_id": 24012619984051000,
	"refresh_url": "?since_id=250126199840518145&q=%23freebandnames&result_type=mixed&include_entities=1",
	"next_results": "?max_id=249279667666817023&q=%23freebandnames&count=4&include_entities=1&result_type=mixed",
	"count": 4,
	"completed_in": 0.035,
	"since_id_str": "24012619984051000",
	"query": "%23freebandnames",
	"max_id_str": "250126199840518145"
}
}`

type TwitterStruct struct {
	Statuses       []Statuses     `json:"statuses"`
	SearchMetadata SearchMetadata `json:"search_metadata"`
}

type Hashtags struct {
	Text    string `json:"text"`
	Indices []int  `json:"indices"`
}

type Entities struct {
	Urls         []any      `json:"urls"`
	Hashtags     []Hashtags `json:"hashtags"`
	UserMentions []any      `json:"user_mentions"`
}

type Metadata struct {
	IsoLanguageCode string `json:"iso_language_code"`
	ResultType      string `json:"result_type"`
}

type Urls struct {
	ExpandedURL any    `json:"expanded_url"`
	URL         string `json:"url"`
	Indices     []int  `json:"indices"`
}

type URL struct {
	Urls []Urls `json:"urls"`
}

type Description struct {
	Urls []any `json:"urls"`
}

type UserEntities struct {
	URL         URL         `json:"url"`
	Description Description `json:"description"`
}

type User struct {
	ProfileSidebarFillColor        string       `json:"profile_sidebar_fill_color"`
	ProfileSidebarBorderColor      string       `json:"profile_sidebar_border_color"`
	ProfileBackgroundTile          bool         `json:"profile_background_tile"`
	Name                           string       `json:"name"`
	ProfileImageURL                string       `json:"profile_image_url"`
	CreatedAt                      string       `json:"created_at"`
	Location                       string       `json:"location"`
	FollowRequestSent              any          `json:"follow_request_sent"`
	ProfileLinkColor               string       `json:"profile_link_color"`
	IsTranslator                   bool         `json:"is_translator"`
	IDStr                          string       `json:"id_str"`
	Entities                       UserEntities `json:"entities"`
	DefaultProfile                 bool         `json:"default_profile"`
	ContributorsEnabled            bool         `json:"contributors_enabled"`
	FavouritesCount                int          `json:"favourites_count"`
	URL                            any          `json:"url"`
	ProfileImageURLHTTPS           string       `json:"profile_image_url_https"`
	UtcOffset                      int          `json:"utc_offset"`
	ID                             int          `json:"id"`
	ProfileUseBackgroundImage      bool         `json:"profile_use_background_image"`
	ListedCount                    int          `json:"listed_count"`
	ProfileTextColor               string       `json:"profile_text_color"`
	Lang                           string       `json:"lang"`
	FollowersCount                 int          `json:"followers_count"`
	Protected                      bool         `json:"protected"`
	Notifications                  any          `json:"notifications"`
	ProfileBackgroundImageURLHTTPS string       `json:"profile_background_image_url_https"`
	ProfileBackgroundColor         string       `json:"profile_background_color"`
	Verified                       bool         `json:"verified"`
	GeoEnabled                     bool         `json:"geo_enabled"`
	TimeZone                       string       `json:"time_zone"`
	Description                    string       `json:"description"`
	DefaultProfileImage            bool         `json:"default_profile_image"`
	ProfileBackgroundImageURL      string       `json:"profile_background_image_url"`
	StatusesCount                  int          `json:"statuses_count"`
	FriendsCount                   int          `json:"friends_count"`
	Following                      any          `json:"following"`
	ShowAllInlineMedia             bool         `json:"show_all_inline_media"`
	ScreenName                     string       `json:"screen_name"`
}

type Statuses struct {
	Coordinates          any      `json:"coordinates"`
	Favorited            bool     `json:"favorited"`
	Truncated            bool     `json:"truncated"`
	CreatedAt            string   `json:"created_at"`
	IDStr                string   `json:"id_str"`
	Entities             Entities `json:"entities"`
	InReplyToUserIDStr   any      `json:"in_reply_to_user_id_str"`
	Contributors         any      `json:"contributors"`
	Text                 string   `json:"text"`
	Metadata             Metadata `json:"metadata"`
	RetweetCount         int      `json:"retweet_count"`
	InReplyToStatusIDStr any      `json:"in_reply_to_status_id_str"`
	ID                   int64    `json:"id"`
	Geo                  any      `json:"geo"`
	Retweeted            bool     `json:"retweeted"`
	InReplyToUserID      any      `json:"in_reply_to_user_id"`
	Place                any      `json:"place"`
	User                 User     `json:"user"`
	InReplyToScreenName  any      `json:"in_reply_to_screen_name"`
	Source               string   `json:"source"`
	InReplyToStatusID    any      `json:"in_reply_to_status_id"`
}

type SearchMetadata struct {
	MaxID       int64   `json:"max_id"`
	SinceID     int64   `json:"since_id"`
	RefreshURL  string  `json:"refresh_url"`
	NextResults string  `json:"next_results"`
	Count       int     `json:"count"`
	CompletedIn float64 `json:"completed_in"`
	SinceIDStr  string  `json:"since_id_str"`
	Query       string  `json:"query"`
	MaxIDStr    string  `json:"max_id_str"`
}
