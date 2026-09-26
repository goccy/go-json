//go:build go1.27 && goexperiment.jsonv2

package json_test

// methodErrorsAsIs is whether encoding/json returns the type error of an unmarshal method as it is: Go 1.27 does.
const methodErrorsAsIs = true
