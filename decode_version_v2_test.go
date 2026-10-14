//go:build go1.27 && goexperiment.jsonv2

package json_test

// stdDecodesAsGo127 is whether encoding/json of the running Go decodes as the one of Go 1.27, which go-json decodes
// as on every version of Go: it is made of encoding/json/v2 there only.
const stdDecodesAsGo127 = true

// tokenDepthCounted is whether encoding/json counts the objects and the arrays which Token opened in the depth of a
// value which Decode decodes in them: encoding/json made of encoding/json/v2 does.
const tokenDepthCounted = true
