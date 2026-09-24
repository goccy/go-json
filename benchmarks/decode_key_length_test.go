package benchmark

import (
	stdjson "encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	gojson "github.com/goccy/go-json"
)

// The decode benchmarks of the key length: three structs of the same fields and values whose keys differ in
// length only, of less than 8 bytes, of 8 to 15 bytes and of 16 to 31 bytes, each decoded from an array of
// objects. The keys of real payloads are of every one of these lengths: in the responses of the GitHub API,
// of Kubernetes, of Stripe and of Twitter, 40 to 60% of the keys are of 8 to 15 bytes and 5 to 30% longer.

type ShortKeys struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	State  string `json:"state"`
	Count  int    `json:"count"`
	Email  string `json:"email"`
	URL    string `json:"url"`
	Size   int    `json:"size"`
	Title  string `json:"title"`
	Forks  int    `json:"forks"`
	Owner  string `json:"owner"`
	Issues int    `json:"issues"`
}

type MediumKeys struct {
	ID     int    `json:"repository_id"`
	Name   string `json:"full_name"`
	Type   string `json:"visibility"`
	State  string `json:"merge_state"`
	Count  int    `json:"forks_count"`
	Email  string `json:"author_email"`
	URL    string `json:"html_url"`
	Size   int    `json:"disk_usage"`
	Title  string `json:"description"`
	Forks  int    `json:"watchers_count"`
	Owner  string `json:"default_branch"`
	Issues int    `json:"open_issues"`
}

type LongKeys struct {
	ID     int    `json:"stargazers_count"`
	Name   string `json:"squash_merge_commit_title"`
	Type   string `json:"merge_commit_message"`
	State  string `json:"security_and_analysis"`
	Count  int    `json:"subscribers_count"`
	Email  string `json:"commit_author_email_address"`
	URL    string `json:"repository_html_url"`
	Size   int    `json:"open_issues_count"`
	Title  string `json:"repository_description"`
	Forks  int    `json:"delete_branch_on_merge"`
	Owner  string `json:"allow_update_branch"`
	Issues int    `json:"web_commit_signoff_required"`
}

// keyLengthFixture returns the JSON of an array of objects of the struct type, by its tags.
func keyLengthFixture(typ reflect.Type) []byte {
	values := []string{`12345`, `"go-json"`, `"public"`, `"clean"`, `42`, `"gopher@example.com"`,
		`"https://github.com/goccy/go-json"`, `2048`, `"Fast JSON encoder/decoder compatible with encoding/json"`,
		`3100`, `"master"`, `17`}
	var objects []string
	for i := 0; i < 16; i++ {
		var fields []string
		for j := 0; j < typ.NumField(); j++ {
			fields = append(fields, fmt.Sprintf("%q:%s", typ.Field(j).Tag.Get("json"), values[j]))
		}
		objects = append(objects, "{"+strings.Join(fields, ",")+"}")
	}
	return []byte("[" + strings.Join(objects, ",") + "]")
}

var (
	shortKeysFixture  = keyLengthFixture(reflect.TypeOf(ShortKeys{}))
	mediumKeysFixture = keyLengthFixture(reflect.TypeOf(MediumKeys{}))
	longKeysFixture   = keyLengthFixture(reflect.TypeOf(LongKeys{}))
)

func init() {
	for _, typ := range []reflect.Type{reflect.TypeOf([]ShortKeys{}), reflect.TypeOf([]MediumKeys{}), reflect.TypeOf([]LongKeys{})} {
		if err := sonic.Pretouch(typ); err != nil {
			panic(err)
		}
	}
}

func benchDecodeOf[T any](b *testing.B, data []byte) {
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v T
		if err := gojson.UnmarshalOf(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Decode_ShortKeys_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[[]ShortKeys](b, shortKeysFixture, stdjson.Unmarshal)
}

func Benchmark_Decode_ShortKeys_Unmarshal_GoJson(b *testing.B) {
	benchDecode[[]ShortKeys](b, shortKeysFixture, gojson.Unmarshal)
}

func Benchmark_Decode_ShortKeys_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[[]ShortKeys](b, shortKeysFixture)
}

func Benchmark_Decode_ShortKeys_Unmarshal_Sonic(b *testing.B) {
	benchDecode[[]ShortKeys](b, shortKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_ShortKeys_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[[]ShortKeys](b, shortKeysFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_MediumKeys_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[[]MediumKeys](b, mediumKeysFixture, stdjson.Unmarshal)
}

func Benchmark_Decode_MediumKeys_Unmarshal_GoJson(b *testing.B) {
	benchDecode[[]MediumKeys](b, mediumKeysFixture, gojson.Unmarshal)
}

func Benchmark_Decode_MediumKeys_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[[]MediumKeys](b, mediumKeysFixture)
}

func Benchmark_Decode_MediumKeys_Unmarshal_Sonic(b *testing.B) {
	benchDecode[[]MediumKeys](b, mediumKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_MediumKeys_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[[]MediumKeys](b, mediumKeysFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_LongKeys_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[[]LongKeys](b, longKeysFixture, stdjson.Unmarshal)
}

func Benchmark_Decode_LongKeys_Unmarshal_GoJson(b *testing.B) {
	benchDecode[[]LongKeys](b, longKeysFixture, gojson.Unmarshal)
}

func Benchmark_Decode_LongKeys_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[[]LongKeys](b, longKeysFixture)
}

func Benchmark_Decode_LongKeys_Unmarshal_Sonic(b *testing.B) {
	benchDecode[[]LongKeys](b, longKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_LongKeys_Unmarshal_SonicStd(b *testing.B) {
	benchDecode[[]LongKeys](b, longKeysFixture, sonic.ConfigStd.Unmarshal)
}
