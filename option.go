package json

import (
	"github.com/goccy/go-json/internal/encoder"
)

type EncodeOption = encoder.Option

type EncodeOptionFunc func(*EncodeOption)

func UnorderedMap() EncodeOptionFunc {
	return func(opt *EncodeOption) {
		opt.UnorderedMap = true
	}
}

func Debug() EncodeOptionFunc {
	return func(opt *EncodeOption) {
		opt.Debug = true
	}
}
