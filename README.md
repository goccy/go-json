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

# JSON library comparison

|  name  |  encoder | decoder | compatible with `encoding/json` |
| :----: | :------: | :-----: | :-----------------------------: |
| encoding/json |  ○ | ○ | N/A |
| [json-iterator/go](https://github.com/json-iterator/go) | ○ | ○ | △ |
| [easyjson](https://github.com/mailru/easyjson) | ○ | ○ |  ✗ |
| [gojay](https://github.com/francoispqt/gojay) | ○ | ○ |  ✗ |
| [segmentio/encoding/json](github.com/segmentio/encoding/json) | ○ | ○ | ○ |
| [jettison](github.com/wI2L/jettison) | ○ | ✗ | ✗ |
| [simdjson-go](https://github.com/minio/simdjson-go) | ✗ | ○ | ✗ |
| go-json | ○ | ○ | ○ |

- `json-iterator/go` isn't compatible with `encoding/json` in many ways, but it hasn't been supported for a long time.


# Benchmarks

```
$ cd benchmarks
$ go test -bench .
```

## Encode

<img width="700px" src="https://user-images.githubusercontent.com/209884/107126758-0845cb00-68f5-11eb-8db7-086fcf9bcfaa.png"></img>
<img width="700px" src="https://user-images.githubusercontent.com/209884/107126757-07ad3480-68f5-11eb-87aa-858cc5eacfcb.png"></img>

## Decode

<img width="700px" src="https://user-images.githubusercontent.com/209884/107126756-067c0780-68f5-11eb-938a-8bc61e3c5014.png"></img>
<img width="700px" src="https://user-images.githubusercontent.com/209884/107126754-054ada80-68f5-11eb-9f93-199f2d75bec7.png"></img>
<img width="700px" src="https://user-images.githubusercontent.com/209884/107126752-024fea00-68f5-11eb-8f63-f32844de2c99.png"></img>

# License

MIT
