package json

import (
	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/errors"
)

var (
	NewSyntaxError    = errors.ErrSyntax
	NewMarshalerError = errors.ErrMarshaler
	// ContextWithSelectionForTest makes the context which MarshalJSON(context.Context) is given.
	ContextWithSelectionForTest = encoder.ContextWithSelection
)
