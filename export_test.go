package json

import (
	"github.com/goccy/go-json/internal/decoder"
	"github.com/goccy/go-json/internal/errors"
)

var (
	NewSyntaxError    = errors.ErrSyntax
	NewMarshalerError = errors.ErrMarshaler
	// IsAtEnd reports whether a syntax error of Unmarshal is the end of the input, which a Decoder reports as
	// io.ErrUnexpectedEOF.
	IsAtEnd = errors.IsAtEnd
)

// UnmarshalLikeDecoder is Unmarshal with the options which a Decoder has after DisallowUnknownFields and UseNumber,
// and the ones of optFuncs: it decodes a value as a Decoder which reads it at once does.
func UnmarshalLikeDecoder(data []byte, v any, disallowUnknownFields, useNumber bool, optFuncs ...DecodeOptionFunc) error {
	return unmarshal(data, v, append(optFuncs, func(o *DecodeOption) {
		if disallowUnknownFields {
			o.Flags |= decoder.DisallowUnknownFieldsOption
		}
		if useNumber {
			o.Flags |= decoder.UseNumberOption
		}
	})...)
}
