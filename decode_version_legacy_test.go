//go:build !go1.27 || !goexperiment.jsonv2

package json_test

// methodErrorsAsIs is whether encoding/json returns the type error of an unmarshal method as it is: before Go 1.27,
// it sets the struct and the field of it.
const methodErrorsAsIs = false

// typeErrorsFromValueStart is whether the offset of a type error of a value of a stream is from the start of the value:
// Go 1.27 does, and encoding/json before it is from the end of the value before.
const typeErrorsFromValueStart = false

// tokenDepthCounted is whether encoding/json counts the objects and the arrays which Token opened in the depth of a
// value which Decode decodes in them: encoding/json before Go 1.27 doesn't.
const tokenDepthCounted = false
