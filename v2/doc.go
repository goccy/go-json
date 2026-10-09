// Package json converts Go values to JSON and back, with the API of encoding/json/v2 of Go 1.27: code which
// imports encoding/json/v2 can import github.com/goccy/go-json/v2 instead, and code which imports
// encoding/json/jsontext can import github.com/goccy/go-json/jsontext.
//
// It builds with every Go version this module supports, not only Go 1.27, and behaves as encoding/json/v2 of
// Go 1.27 does with any of them. Its types and functions are its own: they have the names, signatures and
// behavior of encoding/json/v2, but they are not aliases of it.
//
// The encoder which a MarshalJSONTo method or a function of MarshalToFunc is given writes to an output of its own,
// which is checked when the method returns: the value fails if the method wrote other than one value at its place,
// or ended the object or the array its value is in. A call of MarshalEncode which fails writes nothing to the
// encoder, and fails the value of the method whether the method returns the error or not, also where
// encoding/json/v2, whose encoder is shared by the calls, succeeds by what the value wrote before it failed.
package json
