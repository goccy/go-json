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

**Fastest**

<img width="700" alt="" src="https://user-images.githubusercontent.com/209884/102718073-82ac9280-4329-11eb-94f2-c5377a2feeed.png">
<img width="700" alt="" src="https://user-images.githubusercontent.com/209884/102718071-804a3880-4329-11eb-9e70-5de74e55a553.png">

## Decode

**So faster than json-iterator/go**

## json.Unmarshal

### SmallStruct

<img src="https://user-images.githubusercontent.com/209884/89118870-5b713800-d4e4-11ea-9c80-47008d998e70.png"></img>

### MediumStruct

<img src="https://user-images.githubusercontent.com/209884/89118884-86f42280-d4e4-11ea-965c-b72764870ed0.png"></img>

### LargeStruct

<img src="https://user-images.githubusercontent.com/209884/89118902-9c694c80-d4e4-11ea-94e6-8c888cdb6361.png"></img>

## Stream Decode

### SmallStruct

<img src="https://user-images.githubusercontent.com/209884/89118906-b0ad4980-d4e4-11ea-80fb-2a6e9e7a066e.png"></img>

### MediumStruct

<img src="https://user-images.githubusercontent.com/209884/89118917-c02c9280-d4e4-11ea-8ba8-776cdbf970df.png"></img>

### LargeStruct

<img src="https://user-images.githubusercontent.com/209884/89118920-c28eec80-d4e4-11ea-91cc-424cfe726539.png"></img>

# License

MIT
