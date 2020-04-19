# go-json
Fast JSON encoder/decoder compatible with encoding/json for Go

# Benchmarks

```
$ go test -bench .
goos: darwin
goarch: amd64
pkg: github.com/goccy/go-json
Benchmark_jsoniter-12            5000000               377 ns/op              56 B/op          2 allocs/op
Benchmark_gojay-12               5000000               273 ns/op             512 B/op          1 allocs/op
Benchmark_gojson-12              5000000               242 ns/op              48 B/op          1 allocs/op
PASS
ok      github.com/goccy/go-json        5.392s
```
