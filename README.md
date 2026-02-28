# json5-go

High-performance JSON5 encoder/decoder for Go.

`json5-go` is a fast and compatible **JSON5** library for Go.
It is built on top of the architecture of [goccy/go-json](https://github.com/goccy/go-json) and inherits its performance-oriented design, while extending it to support the **JSON5 specification**.

## 🎯 Project Goal

This library does **not** aim to replace Go’s standard `encoding/json` package.

The Go standard library implements **strict JSON (RFC 8259)** — which is correct and desirable for production APIs and interoperable systems.

`json5-go` exists for a different purpose:

> To provide a high-performance solution when you need to parse or generate **JSON5**, not strict JSON.
> Implement json5 spec with high performance. [json5 spec](https://spec.json5.org)

JSON5 is commonly used for:

- Configuration files
- Developer-facing formats
- Human-friendly structured data
- Environments where comments and relaxed syntax are useful

If you need strict JSON for APIs, use `encoding/json`.
If you need **JSON5 support with performance in mind**, use `json5-go`.

## ✨ Features

- ✅ JSON5 specification support
- ✅ High-performance encoder and decoder
- ✅ Familiar `Marshal` / `Unmarshal` API
- ✅ Drop-in style usage (similar to `encoding/json`)
- ✅ Compatible with Go structs and tags
- ✅ Designed for low allocation and speed

## 📦 Installation

```bash
go get github.com/vayload/json5-go
```

## 🚀 Usage

Simply import `json5-go` instead of `encoding/json`:

```go
import json5 "github.com/vayload/json5-go"
```

### Example

```go
package main

import (
	"fmt"
	json5 "github.com/vayload/json5-go"
)

type Config struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func main() {
	input := []byte(`
	{
		// JSON5 comment
		name: "app",
		port: 8080,
	}
	`)

	var cfg Config
	if err := json5.Unmarshal(input, &cfg); err != nil {
		panic(err)
	}

	out, err := json5.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
```

## ⚙️ Design

This project is a fork of goccy/go-json and reuses its internal high-performance architecture.

The goal of this fork is to:

- Extend parsing and encoding to support JSON5
- Maintain compatibility with Go’s standard JSON APIs
- Keep performance competitive with modern Go JSON libraries
- Avoid unnecessary complexity

The focus is pragmatic: **JSON5 support without sacrificing speed.**

## 📌 Roadmap

- v0.9.0 — API stabilization and JSON5 feature completeness
- v1.0.0 — Stable release

Feature requests and contributions are welcome via GitHub Issues.

## 🧪 Testing

```bash
go test ./...
```

## 📄 License

MIT License.

This project includes code derived from [goccy/go-json](https://github.com/goccy/go-json), which is also licensed under the MIT License.
