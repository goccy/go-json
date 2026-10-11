//go:build go1.27 && goexperiment.jsonv2

package json_test

// methodErrorsAsIs is whether encoding/json returns the type error of an unmarshal method as it is: Go 1.27 does.
const methodErrorsAsIs = true

// errorsShowWholeChar is whether the message of a syntax error at a character shows the character, which a stream
// reads whole first: Go 1.27 does, and encoding/json before it shows the byte.
const errorsShowWholeChar = true
