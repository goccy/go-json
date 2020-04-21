package json

import "errors"

var (
	ErrUnknownType     = errors.New("unknown type name")
	ErrCompileSlowPath = errors.New("detect dynamic type ( interface{} ) and compile with slow path")
)
