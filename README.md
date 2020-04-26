# go-json

![Go](https://github.com/goccy/go-json/workflows/Go/badge.svg)

Fast JSON encoder/decoder compatible with encoding/json for Go

# Status

WIP

# Benchmarks

```
$ cd benchmarks
$ go test -bench .
```

## Environment

```
goos: darwin
goarch: amd64
```

## Decode

### SmallStruct

```
Benchmark_Decode_SmallStruct_EncodingJson-12             1000000              1660 ns/op             280 B/op          3 allocs/op
Benchmark_Decode_SmallStruct_JsonIter-12                 1000000              1284 ns/op             316 B/op         12 allocs/op
Benchmark_Decode_SmallStruct_EasyJson-12                 2000000               613 ns/op             240 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJay-12                    3000000               557 ns/op             256 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJayUnsafe-12              3000000               507 ns/op             112 B/op          1 allocs/op
Benchmark_Decode_SmallStruct_GoJson-12                   3000000               512 ns/op             256 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJsonNoEscape-12           3000000               466 ns/op             144 B/op          1 allocs/op
```

### MediumStruct

```
Benchmark_Decode_MediumStruct_EncodingJson-12             100000             20643 ns/op             384 B/op         10 allocs/op
Benchmark_Decode_MediumStruct_JsonIter-12                 200000             11367 ns/op            2985 B/op         81 allocs/op
Benchmark_Decode_MediumStruct_EasyJson-12                 200000              6635 ns/op             232 B/op          6 allocs/op
Benchmark_Decode_MediumStruct_GoJay-12                    500000              3398 ns/op            2449 B/op          8 allocs/op
Benchmark_Decode_MediumStruct_GoJayUnsafe-12              500000              3067 ns/op             144 B/op          7 allocs/op
```

### LargeStruct

```
Benchmark_Decode_LargeStruct_EncodingJson-12                5000            288411 ns/op             312 B/op          6 allocs/op
Benchmark_Decode_LargeStruct_JsonIter-12                   10000            180028 ns/op           41737 B/op       1137 allocs/op
Benchmark_Decode_LargeStruct_EasyJson-12                   10000            105801 ns/op             160 B/op          2 allocs/op
Benchmark_Decode_LargeStruct_GoJay-12                      50000             35966 ns/op           31244 B/op         77 allocs/op
Benchmark_Decode_LargeStruct_GoJayUnsafe-12                50000             32536 ns/op            2561 B/op         76 allocs/op
```

## Encode

### SmallStruct

```
Benchmark_Encode_SmallStruct_EncodingJson-12             1000000              1665 ns/op            1048 B/op          8 allocs/op
Benchmark_Encode_SmallStruct_JsonIter-12                 2000000               737 ns/op             984 B/op          7 allocs/op
Benchmark_Encode_SmallStruct_EasyJson-12                 3000000               550 ns/op             944 B/op          6 allocs/op
Benchmark_Encode_SmallStruct_GoJay-12                    3000000               421 ns/op             624 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_GoJson-12                   5000000               347 ns/op             256 B/op          2 allocs/op
```

### MediumStruct

```
Benchmark_Encode_MediumStruct_EncodingJson-12             500000              3800 ns/op            1712 B/op         24 allocs/op
Benchmark_Encode_MediumStruct_JsonIter-12                1000000              1438 ns/op            1536 B/op         20 allocs/op
Benchmark_Encode_MediumStruct_EasyJson-12                1000000              1154 ns/op            1320 B/op         19 allocs/op
Benchmark_Encode_MediumStruct_GoJay-12                   1000000              1040 ns/op             824 B/op         15 allocs/op
Benchmark_Encode_MediumStruct_GoJson-12                  2000000               898 ns/op             632 B/op         15 allocs/op
```

### LargeStruct

```
Benchmark_Encode_LargeStruct_EncodingJson-12               30000             53287 ns/op           20388 B/op        331 allocs/op
Benchmark_Encode_LargeStruct_JsonIter-12                  100000             21251 ns/op           20270 B/op        328 allocs/op
Benchmark_Encode_LargeStruct_EasyJson-12                  100000             21303 ns/op           15461 B/op        327 allocs/op
Benchmark_Encode_LargeStruct_GoJay-12                     100000             22500 ns/op           28049 B/op        323 allocs/op
Benchmark_Encode_LargeStruct_GoJson-12                    100000             17957 ns/op           14697 B/op        319 allocs/op
```

# License

MIT
