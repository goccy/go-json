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
Benchmark_Decode_SmallStruct_EncodingJson-12             1000000              1697 ns/op             280 B/op          3 allocs/op
Benchmark_Decode_SmallStruct_JsonIter-12                 1000000              1292 ns/op             316 B/op         12 allocs/op
Benchmark_Decode_SmallStruct_EasyJson-12                 2000000               626 ns/op             240 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJay-12                    3000000               559 ns/op             256 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJayUnsafe-12              3000000               523 ns/op             112 B/op          1 allocs/op
Benchmark_Decode_SmallStruct_GoJson-12                   3000000               530 ns/op             256 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJsonNoEscape-12           3000000               478 ns/op             144 B/op          1 allocs/op
```

### MediumStruct

```
Benchmark_Decode_MediumStruct_EncodingJson-12             100000             20865 ns/op             384 B/op         10 allocs/op
Benchmark_Decode_MediumStruct_JsonIter-12                 200000             11459 ns/op            2985 B/op         81 allocs/op
Benchmark_Decode_MediumStruct_EasyJson-12                 200000              6521 ns/op             232 B/op          6 allocs/op
Benchmark_Decode_MediumStruct_GoJay-12                    500000              3503 ns/op            2449 B/op          8 allocs/op
Benchmark_Decode_MediumStruct_GoJayUnsafe-12              500000              3226 ns/op             144 B/op          7 allocs/op
Benchmark_Decode_MediumStruct_GoJson-12                   500000              3648 ns/op            2457 B/op          8 allocs/op
Benchmark_Decode_MediumStruct_GoJsonNoEscape-12           500000              3606 ns/op            2425 B/op          7 allocs/op
```

### LargeStruct

```
Benchmark_Decode_LargeStruct_EncodingJson-12                5000            293675 ns/op             312 B/op          6 allocs/op
Benchmark_Decode_LargeStruct_JsonIter-12                   10000            182299 ns/op           41737 B/op       1137 allocs/op
Benchmark_Decode_LargeStruct_EasyJson-12                   10000            107157 ns/op             160 B/op          2 allocs/op
Benchmark_Decode_LargeStruct_GoJay-12                      50000             36518 ns/op           31244 B/op         77 allocs/op
Benchmark_Decode_LargeStruct_GoJayUnsafe-12                50000             33196 ns/op            2561 B/op         76 allocs/op
Benchmark_Decode_LargeStruct_GoJson-12                     30000             56653 ns/op           31228 B/op         75 allocs/op
Benchmark_Decode_LargeStruct_GoJsonNoEscape-12             30000             56049 ns/op           31196 B/op         74 allocs/op
```

## Encode

### SmallStruct

```
Benchmark_Encode_SmallStruct_EncodingJson-12             1000000              1696 ns/op            1048 B/op          8 allocs/op
Benchmark_Encode_SmallStruct_JsonIter-12                 2000000               755 ns/op             984 B/op          7 allocs/op
Benchmark_Encode_SmallStruct_EasyJson-12                 3000000               536 ns/op             944 B/op          6 allocs/op
Benchmark_Encode_SmallStruct_GoJay-12                    3000000               417 ns/op             624 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_GoJson-12                   5000000               323 ns/op             144 B/op          1 allocs/op
```

### MediumStruct

```
Benchmark_Encode_MediumStruct_EncodingJson-12             300000              3885 ns/op            1712 B/op         24 allocs/op
Benchmark_Encode_MediumStruct_JsonIter-12                1000000              1420 ns/op            1536 B/op         20 allocs/op
Benchmark_Encode_MediumStruct_EasyJson-12                1000000              1148 ns/op            1320 B/op         19 allocs/op
Benchmark_Encode_MediumStruct_GoJay-12                   1000000              1044 ns/op             824 B/op         15 allocs/op
Benchmark_Encode_MediumStruct_GoJson-12                  3000000               585 ns/op             320 B/op          1 allocs/op
```

### LargeStruct

```
Benchmark_Encode_LargeStruct_EncodingJson-12               30000             53239 ns/op           20393 B/op        331 allocs/op
Benchmark_Encode_LargeStruct_JsonIter-12                  100000             21627 ns/op           20278 B/op        328 allocs/op
Benchmark_Encode_LargeStruct_EasyJson-12                  100000             21629 ns/op           15461 B/op        327 allocs/op
Benchmark_Encode_LargeStruct_GoJay-12                     100000             22256 ns/op           28048 B/op        323 allocs/op
Benchmark_Encode_LargeStruct_GoJson-12                    100000             17927 ns/op           14683 B/op        319 allocs/op
```

# License

MIT
