//go:build !go1.27 || !goexperiment.jsonv2

package json_test

// methodErrorsAsIs is whether encoding/json returns the type error of an unmarshal method as it is: before Go 1.27,
// it sets the struct and the field of it.
const methodErrorsAsIs = false
