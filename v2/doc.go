// Package json converts Go values to JSON and back, with the API of encoding/json/v2 of Go 1.27: code which
// imports encoding/json/v2 can import github.com/goccy/go-json/v2 instead, and code which imports
// encoding/json/jsontext can import github.com/goccy/go-json/jsontext.
//
// It builds with every Go version this module supports, not only Go 1.27, and behaves as encoding/json/v2 of
// Go 1.27 does with any of them. Its types and functions are its own: they have the names, signatures and
// behavior of encoding/json/v2, but they are not aliases of it.
package json
