package json

func NewSyntaxError(msg string, offset int64) *SyntaxError {
	return &SyntaxError{
		msg:    msg,
		Offset: offset,
	}
}
