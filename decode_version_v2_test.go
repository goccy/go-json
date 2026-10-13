//go:build go1.27 && goexperiment.jsonv2

package json_test

// stdDecodesAsGo127 is whether encoding/json of the running Go decodes as the one of Go 1.27, which go-json decodes
// as on every version of Go: it is made of encoding/json/v2 there only.
const stdDecodesAsGo127 = true
