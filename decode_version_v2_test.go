//go:build go1.27 && goexperiment.jsonv2

package json_test

// methodErrorsAsIs is whether encoding/json returns the type error of an unmarshal method as it is: Go 1.27 does.
const methodErrorsAsIs = true

// typeErrorsFromValueStart is whether the offset of a type error of a value of a stream is from the start of the value:
// Go 1.27 does, and encoding/json before it is from the end of the value before.
const typeErrorsFromValueStart = true
