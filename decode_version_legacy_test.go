//go:build !go1.27 || !goexperiment.jsonv2

package json_test

// methodErrorsAsIs is whether encoding/json returns the type error of an unmarshal method as it is: before Go 1.27,
// it sets the struct and the field of it.
const methodErrorsAsIs = false

// errorsShowWholeChar is whether the message of a syntax error at a character shows the character, which a stream
// reads whole first: Go 1.27 does, and encoding/json before it shows the byte.
const errorsShowWholeChar = false
