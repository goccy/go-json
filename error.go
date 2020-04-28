package json

import "golang.org/x/xerrors"

var (
	ErrUnsupportedType = xerrors.New("json: unsupported type")
	ErrDecodePointer   = xerrors.New("json: required pointer type")
)
