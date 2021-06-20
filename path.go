package json

import (
	"reflect"

	"github.com/goccy/go-json/internal/decoder"
)

type PathString string

func (s PathString) Build() (*Path, error) {
	path, err := decoder.PathString(s).Build()
	if err != nil {
		return nil, err
	}
	return &Path{path: path}, nil
}

type Path struct {
	path decoder.Path
}

func (p *Path) Unmarshal(data []byte, v interface{}, optFuncs ...DecodeOptionFunc) error {
	return unmarshalPath(p, data, v, optFuncs...)
}

func (p *Path) Get(src, dst interface{}) error {
	return p.path.Get(reflect.ValueOf(src), reflect.ValueOf(dst))
}
