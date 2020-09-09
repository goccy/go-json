# go-json

![Go](https://github.com/goccy/go-json/workflows/Go/badge.svg)
[![GoDoc](https://godoc.org/github.com/goccy/go-json?status.svg)](https://pkg.go.dev/github.com/goccy/go-json?tab=doc)

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

### SmallStruct

<img src="https://user-images.githubusercontent.com/209884/89118973-5a8cd600-d4e5-11ea-8a07-775cf3e32a2f.png"></img>

### MediumStruct

<img src="https://user-images.githubusercontent.com/209884/89118974-5d87c680-d4e5-11ea-8f4e-dbb01c2dd861.png"></img>

### LargeStruct

<img src="https://user-images.githubusercontent.com/209884/89118977-5f518a00-d4e5-11ea-8bfe-1455fc71c963.png"></img>

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

# Status

## Type

Currently supported all types

## API

Implements All APIs

### Error

- [ ] `InvalidUTF8Error`
- [x] `InvalidUnmarshalError`
- [x] `MarshalerError`
- [x] `SyntaxError`
- [ ] `UnmarshalFieldError`
- [ ] `UnmarshalTypeError`
- [x] `UnsupportedTypeError`
- [x] `UnsupportedValueError`

# License

MIT
