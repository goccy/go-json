# go-json

Fast JSON encoder/decoder compatible with encoding/json for Go

# Status

WIP

# Benchmarks

```
$ go test -bench .
goos: darwin
goarch: amd64
pkg: github.com/goccy/go-json
Benchmark_Encode_jsoniter-12             5000000               375 ns/op              56 B/op          2 allocs/op
Benchmark_Encode_gojay-12                5000000               271 ns/op             512 B/op          1 allocs/op
Benchmark_Encode_gojson-12              10000000               214 ns/op              48 B/op          1 allocs/op
Benchmark_Decode_jsoniter-12             2000000               834 ns/op             208 B/op         13 allocs/op
Benchmark_Decode_gojay-12                3000000               546 ns/op             256 B/op          2 allocs/op
Benchmark_Decode_gojson-12               3000000               478 ns/op             144 B/op          1 allocs/op
PASS
ok      github.com/goccy/go-json        12.921s
```

# License

MIT
