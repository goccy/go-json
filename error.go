package json

import "golang.org/x/xerrors"

var (
	ErrUnsupportedType = xerrors.New("json: unsupported type")
	ErrCompileSlowPath = xerrors.New("json: detect dynamic type ( interface{} ) and compile with slow path")
	ErrDecodePointer   = xerrors.New("json: required pointer type")
)
