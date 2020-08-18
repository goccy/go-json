package json

import "reflect"

func NewSyntaxError(msg string, offset int64) *SyntaxError {
	return &SyntaxError{
		msg:    msg,
		Offset: offset,
	}
}

func NewMarshalerError(typ reflect.Type, err error, msg string) *MarshalerError {
	return &MarshalerError{
		Type:       typ,
		Err:        err,
		sourceFunc: msg,
	}
}
