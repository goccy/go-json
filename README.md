# go-json

![Go](https://github.com/goccy/go-json/workflows/Go/badge.svg)
[![GoDoc](https://godoc.org/github.com/goccy/go-json?status.svg)](https://pkg.go.dev/github.com/goccy/go-json?tab=doc)
[![codecov](https://codecov.io/gh/goccy/go-json/branch/master/graph/badge.svg)](https://codecov.io/gh/goccy/go-json)

Fast JSON encoder/decoder compatible with encoding/json for Go

<img width="400px" src="https://user-images.githubusercontent.com/209884/92572337-42b42900-f2bf-11ea-973a-c74a359553a5.png"></img>

# Installation

```
go get github.com/goccy/go-json
```

# How to use

Replace import statement from `encoding/json` to `github.com/goccy/go-json`

```
-import "encoding/json"
+import "github.com/goccy/go-json"
```

# Benchmarks

```
$ cd benchmarks
$ go test -bench .
```

## Encode

<img width="755" alt="" src="https://user-images.githubusercontent.com/209884/107126463-76898e00-68f3-11eb-9a2b-685c022b07d3.png">
<img width="753" alt="" src="https://user-images.githubusercontent.com/209884/107126462-75f0f780-68f3-11eb-93ab-c3e9406c1fae.png">

## Decode

<img width="753" alt="" src="https://user-images.githubusercontent.com/209884/107126461-75f0f780-68f3-11eb-8ffa-9d2d4eb61fa3.png">
<img width="755" alt="" src="https://user-images.githubusercontent.com/209884/107126460-74bfca80-68f3-11eb-80e0-d7a6cbe5a5b0.png">
<img width="754" alt="" src="https://user-images.githubusercontent.com/209884/107126459-725d7080-68f3-11eb-9216-e810dd6f81b9.png">

# License

MIT
