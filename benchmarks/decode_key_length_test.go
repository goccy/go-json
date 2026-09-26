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

// The decode benchmarks of the keys: structs of the same fields and values whose keys differ only, each decoded
// from an array of objects. The keys of ShortKeys are of less than 8 bytes, the ones of MediumKeys of 8 to 15
// bytes and the ones of LongKeys of 16 to 31 bytes. The keys of real payloads are of every one of these
// lengths: in the responses of the GitHub API, of Kubernetes, of Stripe and of Twitter, 40 to 60% of the keys
// are of 8 to 15 bytes and 5 to 30% longer. The keys of NonASCIIKeys are Japanese words.

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

type NonASCIIKeys struct {
	ID     int    `json:"番号"`
	Name   string `json:"名前"`
	Type   string `json:"種類"`
	State  string `json:"状態"`
	Count  int    `json:"件数"`
	Email  string `json:"メールアドレス"`
	URL    string `json:"ウェブサイト"`
	Size   int    `json:"大きさ"`
	Title  string `json:"表題"`
	Forks  int    `json:"複製数"`
	Owner  string `json:"所有者"`
	Issues int    `json:"課題数"`
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
	shortKeysFixture    = keyLengthFixture(reflect.TypeOf(ShortKeys{}))
	mediumKeysFixture   = keyLengthFixture(reflect.TypeOf(MediumKeys{}))
	longKeysFixture     = keyLengthFixture(reflect.TypeOf(LongKeys{}))
	nonASCIIKeysFixture = keyLengthFixture(reflect.TypeOf(NonASCIIKeys{}))
)

func init() {
	sonicTypes = append(sonicTypes, reflect.TypeOf([]ShortKeys{}), reflect.TypeOf([]MediumKeys{}), reflect.TypeOf([]LongKeys{}), reflect.TypeOf([]NonASCIIKeys{}))
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
	pretouchSonic()
	benchDecode[[]ShortKeys](b, shortKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_ShortKeys_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
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
	pretouchSonic()
	benchDecode[[]MediumKeys](b, mediumKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_MediumKeys_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
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
	pretouchSonic()
	benchDecode[[]LongKeys](b, longKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_LongKeys_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
	benchDecode[[]LongKeys](b, longKeysFixture, sonic.ConfigStd.Unmarshal)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[[]NonASCIIKeys](b, nonASCIIKeysFixture, stdjson.Unmarshal)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_GoJson(b *testing.B) {
	benchDecode[[]NonASCIIKeys](b, nonASCIIKeysFixture, gojson.Unmarshal)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_GoJsonUnmarshalOf(b *testing.B) {
	benchDecodeOf[[]NonASCIIKeys](b, nonASCIIKeysFixture)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_Sonic(b *testing.B) {
	pretouchSonic()
	benchDecode[[]NonASCIIKeys](b, nonASCIIKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_SonicStd(b *testing.B) {
	pretouchSonic()
	benchDecode[[]NonASCIIKeys](b, nonASCIIKeysFixture, sonic.ConfigStd.Unmarshal)
}

// UnknownNonASCIIKeys decodes the objects of NonASCIIKeys into ShortKeys, whose keys are ASCII: every key is of
// no field, and is skipped.
func Benchmark_Decode_UnknownNonASCIIKeys_Unmarshal_EncodingJson(b *testing.B) {
	benchDecode[[]ShortKeys](b, nonASCIIKeysFixture, stdjson.Unmarshal)
}

func Benchmark_Decode_UnknownNonASCIIKeys_Unmarshal_GoJson(b *testing.B) {
	benchDecode[[]ShortKeys](b, nonASCIIKeysFixture, gojson.Unmarshal)
}

func Benchmark_Decode_UnknownNonASCIIKeys_Unmarshal_Sonic(b *testing.B) {
	pretouchSonic()
	benchDecode[[]ShortKeys](b, nonASCIIKeysFixture, sonic.ConfigDefault.Unmarshal)
}

func Benchmark_Decode_ShortKeys_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[[]ShortKeys](b, shortKeysFixture)
}

func Benchmark_Decode_ShortKeys_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[[]ShortKeys](b, shortKeysFixture)
}

func Benchmark_Decode_ShortKeys_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[[]ShortKeys](b, shortKeysFixture)
}

func Benchmark_Decode_MediumKeys_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[[]MediumKeys](b, mediumKeysFixture)
}

func Benchmark_Decode_MediumKeys_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[[]MediumKeys](b, mediumKeysFixture)
}

func Benchmark_Decode_MediumKeys_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[[]MediumKeys](b, mediumKeysFixture)
}

func Benchmark_Decode_LongKeys_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[[]LongKeys](b, longKeysFixture)
}

func Benchmark_Decode_LongKeys_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[[]LongKeys](b, longKeysFixture)
}

func Benchmark_Decode_LongKeys_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[[]LongKeys](b, longKeysFixture)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[[]NonASCIIKeys](b, nonASCIIKeysFixture)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[[]NonASCIIKeys](b, nonASCIIKeysFixture)
}

func Benchmark_Decode_NonASCIIKeys_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[[]NonASCIIKeys](b, nonASCIIKeysFixture)
}

func Benchmark_Decode_UnknownNonASCIIKeys_Unmarshal_GoJsonUnmarshalOfNoCopyString(b *testing.B) {
	benchDecodeNoCopy[[]ShortKeys](b, nonASCIIKeysFixture)
}

func Benchmark_Decode_UnknownNonASCIIKeys_Unmarshal_SonicFastest(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastest[[]ShortKeys](b, nonASCIIKeysFixture)
}

func Benchmark_Decode_UnknownNonASCIIKeys_Unmarshal_SonicFastestValidating(b *testing.B) {
	pretouchSonic()
	benchDecodeSonicFastestValidating[[]ShortKeys](b, nonASCIIKeysFixture)
}
