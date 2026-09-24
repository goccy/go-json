package decoder

import "context"

type OptionFlags uint8

const (
	FirstWinOption OptionFlags = 1 << iota
	ContextOption
	PathOption
	// UseNumberOption decodes a number into interface{} as json.Number instead of float64.
	UseNumberOption
	// DisallowUnknownFieldsOption makes an object key which matches no field of the struct an error.
	DisallowUnknownFieldsOption
)

type Option struct {
	Flags   OptionFlags
	Context context.Context
	Path    *Path
}
