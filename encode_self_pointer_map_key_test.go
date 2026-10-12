package json_test

import (
	stdjson "encoding/json"
	"testing"

	"github.com/goccy/go-json"
)

// selfPointerTextKey names a key by MarshalText, and holds a pointer to itself at the address of the struct.
type selfPointerTextKey struct{ P any }

func (*selfPointerTextKey) MarshalText() ([]byte, error) { return []byte("S"), nil }

// A key whose pointers end at nil or lead back to themselves, as a key of type R *R, has no name: it fails when it
// is encoded, where a nil key was written as null, which is not JSON, and one pointing to itself was followed until
// the stack overflowed. A map without keys and the values of these types are written as encoding/json writes them.
func TestEncodeMapKeyPointingToItself(t *testing.T) {
	self := new(selfPointer)
	*self = selfPointer(self)
	viaInterface := new(any)
	*viaInterface = viaInterface
	first, second := new(cyclePointer), new(cyclePointer2)
	*first, *second = cyclePointer(second), cyclePointer2(first)
	for name, v := range map[string]any{
		"nil":                    map[selfPointer]int{nil: 1},
		"nil behind a pointer":   map[*selfPointer]int{new(selfPointer): 1},
		"nil in a field":         struct{ M map[selfPointer]int }{M: map[selfPointer]int{nil: 1}},
		"itself":                 map[selfPointer]int{selfPointer(self): 1},
		"two types":              map[cyclePointer]int{cyclePointer(second): 1},
		"itself through a key":   map[*any]int{viaInterface: 1},
		"itself in an interface": map[any]int{selfPointer(self): 1},
		"itself and nil":         map[selfPointer]int{selfPointer(self): 1, nil: 2},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := json.Marshal(v)
			if err == nil || got != nil {
				t.Errorf("got %q %v, want an error", got, err)
			}
			got, err = json.MarshalWithOption(v, json.UnorderedMap())
			if err == nil || got != nil {
				t.Errorf("unordered: got %q %v, want an error", got, err)
			}
			got, err = json.MarshalIndentWithOption(v, "", " ", json.Colorize(json.DefaultColorScheme))
			if err == nil || got != nil {
				t.Errorf("indented: got %q %v, want an error", got, err)
			}
		})
	}
	// as encoding/json of Go 1.27 writes them, where Go 1.25 refuses the type of the map.
	for name, tt := range map[string]struct {
		v    any
		want string
	}{
		"nil map":               {map[selfPointer]int(nil), `null`},
		"empty map":             {map[selfPointer]int{}, `{}`},
		"nil map in a field":    {struct{ M map[selfPointer]int }{}, `{"M":null}`},
		"empty map of pointers": {map[*selfPointer]int{}, `{}`},
	} {
		t.Run(name, func(t *testing.T) {
			if got, err := json.Marshal(tt.v); err != nil || string(got) != tt.want {
				t.Errorf("got %s %v, want %s", got, err, tt.want)
			}
		})
	}
	s := &selfPointerTextKey{}
	s.P = s
	var text any = &s.P
	for name, v := range map[string]any{
		"nil value":                  selfPointer(nil),
		"nil field":                  struct{ P selfPointer }{},
		"nil element":                []selfPointer{nil},
		"nil in an interface":        []any{selfPointer(nil)},
		"a struct at the same place": map[*any]int{&text: 1},
	} {
		t.Run(name, func(t *testing.T) {
			want, wantErr := stdjson.Marshal(v)
			got, err := json.Marshal(v)
			if string(got) != string(want) || (err == nil) != (wantErr == nil) {
				t.Errorf("got %s %v, want %s %v", got, err, want, wantErr)
			}
		})
	}
}
