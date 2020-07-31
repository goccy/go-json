# go-json

![Go](https://github.com/goccy/go-json/workflows/Go/badge.svg)
[![GoDoc](https://godoc.org/github.com/goccy/go-json?status.svg)](https://pkg.go.dev/github.com/goccy/go-json?tab=doc)

Fast JSON encoder/decoder compatible with encoding/json for Go

# Status

WIP

## API

- [ ] `Compact`
- [ ] `HTMLEscape`
- [ ] `Indent`
- [x] `Marshal`
- [x] `MarshalIndent`
- [x] `Unmarshal`
- [ ] `Valid`
- [x] `NewDecoder`
- [x] `(*Decoder).Buffered`
- [x] `(*Decoder).Decode`
- [ ] `(*Decoder).DisallowUnknownFields`
- [x] `(*Decoder).InputOffset`
- [x] `(*Decoder).More`
- [x] `(*Decoder).Token`
- [ ] `(*Decoder).UseNumber`
- [x] `Delim`
- [x] `(Delim).String`
- [x] `NewEncoder`
- [x] `(*Encoder).Encode`
- [x] `(*Encoder).SetEscapeHTML`
- [x] `(*Encoder).SetIndent`

## Type

### Encoder

- [x] `int`, `int8`, `int16`, `int32`, `int64`
- [x] `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- [x] `float32`, `float64`
- [x] `string`
- [x] `struct`
- [x] `array`
- [x] `slice`
- [x] `map`
- [x] `interface{}`
- [x] `MarshalJSON`
- [x] `MarshalText`
- [x] `pointer`

### Decoder

- [x] `int`, `int8`, `int16`, `int32`, `int64`
- [x] `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- [x] `float32`, `float64`
- [x] `string`
- [x] `struct`
- [x] `array`
- [x] `slice`
- [x] `map`
- [x] `interface{}`
- [x] `UnmarshalJSON`
- [x] `UnmarshalText`
- [x] `pointer`

### Error

- [ ] `InvalidUTF8Error`
- [x] `InvalidUnmarshalError`
- [x] `MarshalerError`
- [x] `SyntaxError`
- [ ] `UnmarshalFieldError`
- [ ] `UnmarshalTypeError`
- [x] `UnsupportedTypeError`
- [x] `UnsupportedValueError`

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
Benchmark_Decode_SmallStruct_EncodingJson-12             1000000              1725 ns/op             280 B/op          3 allocs/op
Benchmark_Decode_SmallStruct_JsonIter-12                 1000000              1282 ns/op             316 B/op         12 allocs/op
Benchmark_Decode_SmallStruct_GoJay-12                    3000000               553 ns/op             256 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJayUnsafe-12              3000000               509 ns/op             112 B/op          1 allocs/op
Benchmark_Decode_SmallStruct_GoJson-12                   3000000               465 ns/op             256 B/op          2 allocs/op
Benchmark_Decode_SmallStruct_GoJsonNoEscape-12           3000000               407 ns/op             144 B/op          1 allocs/op
```

### MediumStruct

```
Benchmark_Decode_MediumStruct_EncodingJson-12             100000             20688 ns/op             384 B/op         10 allocs/op
Benchmark_Decode_MediumStruct_JsonIter-12                 200000             10513 ns/op            2985 B/op         81 allocs/op
Benchmark_Decode_MediumStruct_GoJay-12                    500000              3400 ns/op            2449 B/op          8 allocs/op
Benchmark_Decode_MediumStruct_GoJayUnsafe-12              500000              3095 ns/op             144 B/op          7 allocs/op
Benchmark_Decode_MediumStruct_GoJson-12                   500000              2662 ns/op            2457 B/op          9 allocs/op
Benchmark_Decode_MediumStruct_GoJsonNoEscape-12           500000              2614 ns/op            2425 B/op          8 allocs/op
```

### LargeStruct

```
Benchmark_Decode_LargeStruct_EncodingJson-12                5000            276637 ns/op             312 B/op          6 allocs/op
Benchmark_Decode_LargeStruct_JsonIter-12                   10000            158992 ns/op           41738 B/op       1137 allocs/op
Benchmark_Decode_LargeStruct_GoJay-12                      50000             36340 ns/op           31244 B/op         77 allocs/op
Benchmark_Decode_LargeStruct_GoJayUnsafe-12                50000             34337 ns/op            2561 B/op         76 allocs/op
Benchmark_Decode_LargeStruct_GoJson-12                     50000             39183 ns/op           30755 B/op         67 allocs/op
Benchmark_Decode_LargeStruct_GoJsonNoEscape-12             50000             38809 ns/op           30723 B/op         66 allocs/op
```

## Encode

### SmallStruct

```
Benchmark_Encode_SmallStruct_EncodingJson-12             1000000              1696 ns/op            1048 B/op          8 allocs/op
Benchmark_Encode_SmallStruct_JsonIter-12                 2000000               755 ns/op             984 B/op          7 allocs/op
Benchmark_Encode_SmallStruct_GoJay-12                    3000000               417 ns/op             624 B/op          2 allocs/op
Benchmark_Encode_SmallStruct_GoJson-12                   5000000               323 ns/op             144 B/op          1 allocs/op
```

### MediumStruct

```
Benchmark_Encode_MediumStruct_EncodingJson-12             300000              3885 ns/op            1712 B/op         24 allocs/op
Benchmark_Encode_MediumStruct_JsonIter-12                1000000              1420 ns/op            1536 B/op         20 allocs/op
Benchmark_Encode_MediumStruct_GoJay-12                   1000000              1044 ns/op             824 B/op         15 allocs/op
Benchmark_Encode_MediumStruct_GoJson-12                  3000000               585 ns/op             320 B/op          1 allocs/op
```

### LargeStruct

```
Benchmark_Encode_LargeStruct_EncodingJson-12               30000             53239 ns/op           20393 B/op        331 allocs/op
Benchmark_Encode_LargeStruct_JsonIter-12                  100000             21627 ns/op           20278 B/op        328 allocs/op
Benchmark_Encode_LargeStruct_GoJay-12                     100000             22256 ns/op           28048 B/op        323 allocs/op
Benchmark_Encode_LargeStruct_GoJson-12                    100000             17927 ns/op           14683 B/op        319 allocs/op
```

# License

MIT
